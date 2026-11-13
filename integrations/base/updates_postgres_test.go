package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/buildinfo"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

const testBase = "37618ad4b927a9dd0671599ffc7387a7386f70fa"

type fakeParent struct {
	mu       sync.Mutex
	compares atomic.Int32
	auth     []string
	inm      []string
	status   int
	delay    time.Duration
	ahead    int
}

func (f *fakeParent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.auth = append(f.auth, r.Header.Get("Authorization"))
	f.inm = append(f.inm, r.Header.Get("If-None-Match"))
	status, delay, ahead := f.status, f.delay, f.ahead
	f.mu.Unlock()
	if strings.Contains(r.URL.Path, "/compare/") {
		f.compares.Add(1)
		time.Sleep(delay)
	}
	if status == http.StatusTooManyRequests {
		w.Header().Set("Retry-After", "600")
		w.WriteHeader(status)
		return
	}
	if status != 0 {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"message":"x"}`))
		return
	}
	etag := fmt.Sprintf(`"%s-%d"`, r.URL.Path, ahead)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("ETag", etag)
	switch {
	case strings.Contains(r.URL.Path, "/compare/"):
		commits := []map[string]any{}
		for i := 0; i < ahead; i++ {
			commits = append(commits, map[string]any{"sha": fmt.Sprintf("%040d", i+1), "html_url": "https://github.com/x", "commit": map[string]any{"message": fmt.Sprintf("Change %d", i+1), "author": map[string]any{"name": "a", "date": "2026-11-13T00:00:00Z"}, "committer": map[string]any{"date": "2026-11-13T00:00:00Z"}}})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ahead", "ahead_by": ahead, "total_commits": ahead, "commits": commits,
			"files": []map[string]any{{"filename": "migrate/sql/20261201000000-x.sql", "status": "added", "additions": 3}, {"filename": "server/a.go", "status": "modified", "additions": 1}}})
	case strings.HasSuffix(r.URL.Path, "/releases"):
		_, _ = w.Write([]byte(`[{"tag_name":"v3.43.0","name":"3.43.0","published_at":"2026-11-13T09:00:00Z","html_url":"https://github.com/heroiclabs/nakama/releases/tag/v3.43.0","body":"notes"}]`))
	case strings.HasSuffix(r.URL.Path, "/tags"):
		_, _ = w.Write([]byte(`[{"name":"v3.43.0","commit":{"sha":"` + fmt.Sprintf("%040d", 1) + `"}}]`))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *fakeParent) set(fn func(*fakeParent)) {
	f.mu.Lock()
	fn(f)
	f.mu.Unlock()
}

func testInfo() buildinfo.Info {
	return buildinfo.Info{Commit: "7947112", Branch: "dev", Known: true, Parent: &buildinfo.Parent{Provider: "github", Repo: "heroiclabs/nakama", Branch: "master", Commit: testBase, Date: "2026-11-12T15:32:42Z"}}
}

func newUpdatesHarness(t *testing.T) (*harness, *fakeParent) {
	h := newHarness(t)
	for _, q := range []string{
		`DELETE FROM base_updates_settings`,
		`DELETE FROM base_updates_state`,
		`INSERT INTO base_updates_state (id) VALUES (1)`,
		`DELETE FROM cluster_lease WHERE name='` + updatesLease + `'`,
		`DELETE FROM cluster_rate_limit WHERE key='base:updates:check'`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	fake := &fakeParent{ahead: 3}
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	key, err := secretbox.Base.DeriveKey([]byte(strings.Repeat("k", 40)), secretbox.BaseUpdatesInfo)
	if err != nil {
		t.Fatal(err)
	}
	h.s.updates = updatesConfig{api: srv.URL, key: key, info: testInfo}
	return h, fake
}

func (h *harness) tick(holder string) bool {
	h.t.Helper()
	ran, err := h.s.updatesTick(h.ctx, holder)
	if err != nil {
		h.t.Fatal(err)
	}
	return ran
}

func TestUpdatesCheckStoresResult(t *testing.T) {
	h, fake := newUpdatesHarness(t)
	out := h.mustAPI(200, http.MethodGet, "/updates", nil)
	settings := out["settings"].(map[string]any)
	if settings["repo"] != "heroiclabs/nakama" || settings["branch"] != "master" || settings["intervalHours"] != float64(6) || settings["saved"] != false {
		t.Fatalf("defaults %v", settings)
	}
	if out["state"].(map[string]any)["status"] != "never" || out["result"] != nil {
		t.Fatalf("state %v", out["state"])
	}
	if !h.tick("node-a") {
		t.Fatal("first tick did not check")
	}
	if h.tick("node-a") {
		t.Fatal("checked again before the interval")
	}
	out = h.mustAPI(200, http.MethodGet, "/updates", nil)
	st := out["state"].(map[string]any)
	res := out["result"].(map[string]any)
	if st["status"] != "ok" || st["checks"] != float64(1) || res["aheadBy"] != float64(3) || len(res["commits"].([]any)) != 3 || out["current"] != true {
		t.Fatalf("after check %v %v", st, res)
	}
	if _, ok := res["allReleases"]; ok {
		t.Fatal("internal parts leaked into the view")
	}
	if _, ok := res["sources"]; ok {
		t.Fatal("etags leaked into the view")
	}
	next, _ := time.Parse(time.RFC3339Nano, st["nextAt"].(string))
	if d := time.Until(next); d < 5*time.Hour || d > 7*time.Hour {
		t.Fatalf("next check in %v", d)
	}
	warn := res["warnings"].([]any)
	if len(warn) == 0 || warn[0].(map[string]any)["kind"] != "migrations" {
		t.Fatalf("warnings %v", warn)
	}
	sum := h.mustAPI(200, http.MethodGet, "/updates/summary", nil)
	if sum["newCommits"] != float64(3) || sum["latestRelease"] != "v3.43.0" || sum["status"] != "ok" {
		t.Fatalf("summary %v", sum)
	}
	if fake.compares.Load() != 1 {
		t.Fatalf("compare calls %d", fake.compares.Load())
	}
	h.mustAPI(202, http.MethodPost, "/updates/check", nil)
	if !h.tick("node-a") {
		t.Fatal("queued check did not run")
	}
	fake.mu.Lock()
	sent := fake.inm[len(fake.inm)-3]
	fake.mu.Unlock()
	if sent == "" {
		t.Fatal("second check did not send the ETag")
	}
	out = h.mustAPI(200, http.MethodGet, "/updates", nil)
	if out["result"].(map[string]any)["aheadBy"] != float64(3) || out["state"].(map[string]any)["queued"] != false {
		t.Fatalf("304 lost the result: %v", out)
	}
}

func TestUpdatesWorkerRunsOnceAcrossNodes(t *testing.T) {
	h, fake := newUpdatesHarness(t)
	fake.set(func(f *fakeParent) { f.delay = 300 * time.Millisecond })
	var wg sync.WaitGroup
	var ran atomic.Int32
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			node := &service{db: h.s.db, nk: h.s.nk, logger: testLogger{}, now: time.Now, updates: h.s.updates}
			ok, err := node.updatesTick(context.Background(), fmt.Sprintf("node-%d/p", i))
			if err != nil {
				t.Error(err)
			}
			if ok {
				ran.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if ran.Load() != 1 || fake.compares.Load() != 1 {
		t.Fatalf("ran on %d nodes, %d compare calls", ran.Load(), fake.compares.Load())
	}
	var checks int
	if err := h.s.db.QueryRow(`SELECT checks FROM base_updates_state WHERE id=1`).Scan(&checks); err != nil || checks != 1 {
		t.Fatalf("checks %d %v", checks, err)
	}
	h.mustAPI(202, http.MethodPost, "/updates/check", nil)
	if h.tick("node-other/p") {
		t.Fatal("another node took a lease that was still held")
	}
	if _, err := h.s.db.Exec(`UPDATE cluster_lease SET expires_at = now() - INTERVAL '1 second' WHERE name=$1`, updatesLease); err != nil {
		t.Fatal(err)
	}
	if !h.tick("node-other/p") {
		t.Fatal("expired lease was not taken over")
	}
}

func TestUpdatesSettings(t *testing.T) {
	h, fake := newUpdatesHarness(t)
	put := func(body map[string]any) map[string]any {
		return h.mustAPI(200, http.MethodPut, "/updates/settings", body)
	}
	base := func() map[string]any {
		return map[string]any{"provider": "github", "repo": "heroiclabs/nakama", "branch": "master", "intervalHours": 6, "enabled": true}
	}
	bad := func(code string, mod func(map[string]any)) {
		b := base()
		mod(b)
		h.expectError(400, code, http.MethodPut, "/updates/settings", b)
	}
	bad("INVALID_REPO", func(b map[string]any) { b["repo"] = "https://example.com/a/b" })
	bad("INVALID_BRANCH", func(b map[string]any) { b["branch"] = "a..b" })
	bad("INVALID_INTERVAL", func(b map[string]any) { b["intervalHours"] = 0 })
	bad("INVALID_INTERVAL", func(b map[string]any) { b["intervalHours"] = 169 })
	bad("UNSUPPORTED_PROVIDER", func(b map[string]any) { b["provider"] = "bitbucket" })
	bad("INVALID_GITHUB_TOKEN", func(b map[string]any) { b["token"] = "short" })
	b := base()
	b["repo"] = "https://github.com/copilot-rast/nakama-farsi.git"
	b["branch"] = "stage"
	b["token"] = "github_pat_" + strings.Repeat("A", 30)
	out := put(b)
	if out["repo"] != "copilot-rast/nakama-farsi" || out["branch"] != "stage" || out["token"] != secretMask || out["hasToken"] != true || out["saved"] != true {
		t.Fatalf("saved %v", out)
	}
	var sealed string
	if err := h.s.db.QueryRow(`SELECT token_sealed FROM base_updates_settings`).Scan(&sealed); err != nil || sealed == "" || strings.Contains(sealed, "AAAA") {
		t.Fatalf("token not sealed: %v", err)
	}
	if !h.tick("n/p") {
		t.Fatal("changed settings did not queue a check")
	}
	st := h.mustAPI(200, http.MethodGet, "/updates", nil)["state"].(map[string]any)
	if st["status"] != "base_mismatch" || fake.compares.Load() != 0 {
		t.Fatalf("a parent that does not match the recorded base must not be compared: %v", st)
	}
	b = base()
	b["token"] = secretMask
	out = put(b)
	if out["hasToken"] != true {
		t.Fatal("mask did not keep the token")
	}
	if !h.tick("n/p") {
		t.Fatal("repo change did not queue a check")
	}
	fake.mu.Lock()
	auth := fake.auth[len(fake.auth)-1]
	fake.mu.Unlock()
	if auth != "Bearer github_pat_"+strings.Repeat("A", 30) {
		t.Fatal("stored token was not sent")
	}
	b["token"] = ""
	if out = put(b); out["hasToken"] != false || out["token"] != "" {
		t.Fatalf("token not cleared %v", out)
	}
	b["enabled"] = false
	put(b)
	h.expectError(409, "UPDATES_OFF", http.MethodPost, "/updates/check", nil)
	if sum := h.mustAPI(200, http.MethodGet, "/updates/summary", nil); sum["newCommits"] != float64(0) {
		t.Fatalf("switched off still shows a badge: %v", sum)
	}
	h.s.updates.key = nil
	b["enabled"] = true
	b["token"] = "github_pat_" + strings.Repeat("B", 30)
	h.expectError(400, "SECRET_KEY_MISSING", http.MethodPut, "/updates/settings", b)
}

func TestUpdatesErrorsAndLimits(t *testing.T) {
	h, fake := newUpdatesHarness(t)
	if !h.tick("n/p") {
		t.Fatal("no first check")
	}
	fake.set(func(f *fakeParent) { f.status = http.StatusTooManyRequests })
	h.mustAPI(202, http.MethodPost, "/updates/check", nil)
	h.tick("n/p")
	out := h.mustAPI(200, http.MethodGet, "/updates", nil)
	st := out["state"].(map[string]any)
	next, _ := time.Parse(time.RFC3339Nano, st["nextAt"].(string))
	if st["status"] != "rate_limited" || time.Until(next) < 9*time.Minute || out["result"] == nil || out["result"].(map[string]any)["aheadBy"] != float64(3) {
		t.Fatalf("rate limit state %v", out)
	}
	h.mustAPI(202, http.MethodPost, "/updates/check", nil)
	if h.tick("n/p") {
		t.Fatal("checked while rate limited")
	}
	if _, err := h.s.db.Exec(`UPDATE base_updates_state SET next_at=now()`); err != nil {
		t.Fatal(err)
	}
	fake.set(func(f *fakeParent) { f.status = http.StatusInternalServerError })
	h.tick("n/p")
	st = h.mustAPI(200, http.MethodGet, "/updates", nil)["state"].(map[string]any)
	next, _ = time.Parse(time.RFC3339Nano, st["nextAt"].(string))
	if st["status"] != "error" || !strings.Contains(st["error"].(string), "500") || time.Until(next) > 31*time.Minute {
		t.Fatalf("error state %v", st)
	}
	if _, err := h.s.db.Exec(`UPDATE base_updates_state SET next_at=now()`); err != nil {
		t.Fatal(err)
	}
	fake.set(func(f *fakeParent) { f.status = http.StatusNotFound })
	h.tick("n/p")
	if st = h.mustAPI(200, http.MethodGet, "/updates", nil)["state"].(map[string]any); st["status"] != "base_missing" {
		t.Fatalf("404 state %v", st)
	}
	fake.set(func(f *fakeParent) { f.status = 0 })
	h.s.updates.info = func() buildinfo.Info { return buildinfo.Info{} }
	if _, err := h.s.db.Exec(`UPDATE base_updates_state SET next_at=now()`); err != nil {
		t.Fatal(err)
	}
	h.tick("n/p")
	if st = h.mustAPI(200, http.MethodGet, "/updates", nil)["state"].(map[string]any); st["status"] != "no_base" {
		t.Fatalf("no build info state %v", st)
	}
	for i := 0; i < updatesCheckLimit-2; i++ {
		h.mustAPI(202, http.MethodPost, "/updates/check", nil)
	}
	code, body := h.api(http.MethodPost, "/updates/check", nil)
	if code != 429 || body["error"] != "TOO_MANY_CHECKS" || body["retryAfter"] == nil {
		t.Fatalf("check limit %d %v", code, body)
	}
}
