package upstream

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeGitHub struct {
	mu      sync.Mutex
	calls   []string
	headers []http.Header
	status  int
	retry   string
	reset   string
	body    map[string]string
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.calls = append(f.calls, r.URL.RequestURI())
	f.headers = append(f.headers, r.Header.Clone())
	status, retry, reset := f.status, f.retry, f.reset
	f.mu.Unlock()
	w.Header().Set("X-RateLimit-Limit", "60")
	w.Header().Set("X-RateLimit-Remaining", "57")
	w.Header().Set("X-RateLimit-Reset", "1900000000")
	if status != 0 {
		if reset != "" {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", reset)
		}
		if retry != "" {
			w.Header().Set("Retry-After", retry)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"message":"nope"}`))
		return
	}
	var key string
	switch {
	case strings.Contains(r.URL.Path, "/compare/"):
		key = "compare"
	case strings.HasSuffix(r.URL.Path, "/releases"):
		key = "releases"
	case strings.HasSuffix(r.URL.Path, "/tags"):
		key = "tags"
	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
		return
	}
	etag := `"` + key + `-v1"`
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("ETag", etag)
	_, _ = w.Write([]byte(f.body[key]))
}

func (f *fakeGitHub) last() http.Header {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.headers[len(f.headers)-1]
}

const (
	baseSHA = "1111111111111111111111111111111111111111"
	headSHA = "3333333333333333333333333333333333333333"
	midSHA  = "2222222222222222222222222222222222222222"
)

func fixture() *fakeGitHub {
	compare := map[string]any{
		"status": "ahead", "ahead_by": 2, "behind_by": 0, "total_commits": 2,
		"commits": []any{
			map[string]any{"sha": midSHA, "html_url": "https://github.com/o/r/commit/" + midSHA, "parents": []any{map[string]any{"sha": baseSHA}},
				"commit": map[string]any{"message": "feat(server)!: drop the old matchmaker flag\n\nBREAKING CHANGE: removed", "author": map[string]any{"name": "Ann", "date": "2026-11-12T10:00:00Z"}, "committer": map[string]any{"date": "2026-11-12T10:00:00Z"}},
				"author": map[string]any{"login": "ann"}},
			map[string]any{"sha": headSHA, "html_url": "https://github.com/o/r/commit/" + headSHA, "parents": []any{map[string]any{"sha": midSHA}, map[string]any{"sha": "x"}},
				"commit": map[string]any{"message": "Merge pull request #1\n\nbody", "author": map[string]any{"name": "Bob", "date": "2026-11-13T10:00:00Z"}, "committer": map[string]any{"date": "2026-11-13T10:00:00Z"}}},
		},
		"files": []any{
			map[string]any{"filename": "migrate/sql/20261201000000-new.sql", "status": "added", "additions": 10},
			map[string]any{"filename": "server/matchmaker.go", "status": "modified", "additions": 3, "deletions": 1, "patch": "@@ -1 +1 @@\n-x := os.Getenv(\"OLD_FLAG\")\n+x := os.Getenv(\"NAKAMA_NEW_FLAG\")"},
			map[string]any{"filename": "build/nakama-start.sh", "status": "modified", "additions": 1, "patch": "@@ -1 +1,2 @@\n+    --runtime.env \"FOO_BAR=${FOO_BAR:-}\""},
			map[string]any{"filename": "go.mod", "status": "modified", "additions": 2, "deletions": 1, "patch": "@@ -3 +3 @@\n-go 1.27.3\n+go 1.28.0\n+\tgithub.com/example/lib v1.2.3"},
			map[string]any{"filename": "vendor/github.com/example/lib/a.go", "status": "added", "additions": 50},
			map[string]any{"filename": "console/ui/src/app.ts", "status": "modified", "additions": 4},
			map[string]any{"filename": "README.md", "status": "modified", "additions": 1},
			map[string]any{"filename": "server/match_test.go", "status": "modified", "additions": 1},
		},
	}
	rel := []any{
		map[string]any{"tag_name": "v3.43.0", "name": "3.43.0", "published_at": "2026-11-13T12:00:00Z", "html_url": "https://github.com/o/r/releases/tag/v3.43.0", "body": "### Removed\n- old flag"},
		map[string]any{"tag_name": "v3.43.0-rc1", "name": "draft", "published_at": "2026-11-13T11:00:00Z", "draft": true},
		map[string]any{"tag_name": "v3.42.0", "name": "3.42.0", "published_at": "2026-11-03T12:00:00Z", "html_url": "https://github.com/o/r/releases/tag/v3.42.0", "body": "older"},
	}
	tags := []any{
		map[string]any{"name": "v3.43.0", "commit": map[string]any{"sha": headSHA}},
		map[string]any{"name": "v3.42.0", "commit": map[string]any{"sha": baseSHA}},
	}
	enc := func(v any) string { b, _ := json.Marshal(v); return string(b) }
	return &fakeGitHub{body: map[string]string{"compare": enc(compare), "releases": enc(rel), "tags": enc(tags)}}
}

func request() Request {
	return Request{Repo: "o/r", Branch: "master", Base: baseSHA, BaseDate: time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)}
}

func warning(res *Result, kind string) *Warning {
	for i := range res.Warnings {
		if res.Warnings[i].Kind == kind {
			return &res.Warnings[i]
		}
	}
	return nil
}

func TestCheckReadsCompareReleasesAndTags(t *testing.T) {
	fake := fixture()
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := &Client{BaseURL: srv.URL}
	res, err := c.Check(context.Background(), request(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.AheadBy != 2 || res.Status != "ahead" || res.HeadSHA != headSHA || len(res.Commits) != 2 || res.Commits[0].SHA != headSHA || !res.Commits[0].Merge {
		t.Fatalf("compare not read newest first: %+v", res)
	}
	if !res.Commits[1].Breaks || res.Commits[1].Author != "ann" || res.Commits[1].Title != "feat(server)!: drop the old matchmaker flag" {
		t.Fatalf("commit fields: %+v", res.Commits[1])
	}
	if !strings.Contains(fake.calls[0], "/repos/o/r/compare/"+baseSHA+"...master") {
		t.Fatalf("compare path %s", fake.calls[0])
	}
	h := fake.last()
	if h.Get("Authorization") != "" || h.Get("User-Agent") == "" || h.Get("Accept") != "application/vnd.github+json" {
		t.Fatalf("headers %v", h)
	}
	keys := []string{}
	for _, a := range res.Areas {
		keys = append(keys, a.Key)
	}
	if strings.Join(keys, ",") != "migrations,core,console,deps,build,e2e,docs" {
		t.Fatalf("areas %v", keys)
	}
	if res.FileCount != 8 {
		t.Fatalf("file count %d", res.FileCount)
	}
	if len(res.Releases) != 1 || res.Releases[0].Tag != "v3.43.0" || !res.Releases[0].Breaks || res.LatestRelease == nil || res.LatestRelease.Tag != "v3.43.0" {
		t.Fatalf("releases %+v latest %+v", res.Releases, res.LatestRelease)
	}
	if len(res.Tags) != 1 || res.Tags[0].Name != "v3.43.0" || res.LatestTag.Name != "v3.43.0" {
		t.Fatalf("tags %+v", res.Tags)
	}
	if w := warning(res, "migrations"); w == nil || w.Items[0] != "20261201000000-new.sql" {
		t.Fatalf("migration warning %+v", res.Warnings)
	}
	if w := warning(res, "env"); w == nil || strings.Join(w.Items, ",") != "FOO_BAR,NAKAMA_NEW_FLAG" {
		t.Fatalf("env warning %+v", w)
	}
	if w := warning(res, "goVersion"); w == nil || strings.Join(w.Items, ",") != "go 1.27.3 → 1.28.0" {
		t.Fatalf("go warning %+v", w)
	}
	if w := warning(res, "deps"); w == nil || len(w.Items) != 1 || w.Items[0] != "github.com/example/lib" {
		t.Fatalf("deps warning %+v", w)
	}
	if w := warning(res, "breaking"); w == nil || len(w.Items) != 2 {
		t.Fatalf("breaking warning %+v", w)
	}
	if res.Rate.Limit != 60 || res.Rate.Remaining != 57 {
		t.Fatalf("rate %+v", res.Rate)
	}
}

func TestCheckSendsETagsAndReusesNotModified(t *testing.T) {
	fake := fixture()
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, Token: "ghp_secret"}
	first, err := c.Check(context.Background(), request(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if fake.last().Get("Authorization") != "Bearer ghp_secret" {
		t.Fatal("token not sent")
	}
	raw, _ := json.Marshal(first)
	var stored Result
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatal(err)
	}
	second, err := c.Check(context.Background(), request(), &stored)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(second.NotModified, ",") != "compare,releases,tags" {
		t.Fatalf("not modified %v", second.NotModified)
	}
	if fake.last().Get("If-None-Match") != `"tags-v1"` {
		t.Fatalf("etag not sent: %v", fake.last())
	}
	if second.AheadBy != 2 || len(second.Commits) != 2 || len(second.Releases) != 1 || len(second.Tags) != 1 || warning(second, "env") == nil || warning(second, "breaking") == nil {
		t.Fatalf("data lost on 304: %+v", second)
	}
	if second.Sources["compare"].ETag != `"compare-v1"` {
		t.Fatalf("etag lost: %+v", second.Sources)
	}
	other := request()
	other.Branch = "dev"
	if _, err := c.Check(context.Background(), other, &stored); err != nil {
		t.Fatal(err)
	}
	if fake.headers[len(fake.headers)-3].Get("If-None-Match") != "" {
		t.Fatal("etag of another compare path was sent")
	}
}

func TestCheckRateLimits(t *testing.T) {
	fake := fixture()
	srv := httptest.NewServer(fake)
	defer srv.Close()
	now := time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC)
	c := &Client{BaseURL: srv.URL, Now: func() time.Time { return now }}
	fake.status, fake.reset = http.StatusForbidden, "1900000000"
	_, err := c.Check(context.Background(), request(), nil)
	var rl *RateLimitError
	if !errors.As(err, &rl) || rl.Status != 403 || !rl.Until.Equal(time.Unix(1900000000, 0).UTC()) {
		t.Fatalf("403 rate limit: %v", err)
	}
	fake.status, fake.reset, fake.retry = http.StatusTooManyRequests, "", "120"
	_, err = c.Check(context.Background(), request(), nil)
	if !errors.As(err, &rl) || rl.Status != 429 || !rl.Until.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("429 rate limit: %v", err)
	}
	fake.status, fake.retry = http.StatusForbidden, ""
	_, err = c.Check(context.Background(), request(), nil)
	var he *HTTPError
	if errors.As(err, &rl) || !errors.As(err, &he) || he.Status != 403 {
		t.Fatalf("plain 403 is not a rate limit: %v", err)
	}
}

func TestCheckErrors(t *testing.T) {
	fake := fixture()
	srv := httptest.NewServer(fake)
	c := &Client{BaseURL: srv.URL, Token: "ghp_topsecret"}
	fake.status = http.StatusNotFound
	if _, err := c.Check(context.Background(), request(), nil); !IsNotFound(err) {
		t.Fatalf("404: %v", err)
	}
	fake.status = http.StatusUnauthorized
	_, err := c.Check(context.Background(), request(), nil)
	if err == nil || strings.Contains(err.Error(), "topsecret") || !strings.Contains(err.Error(), "401") {
		t.Fatalf("401: %v", err)
	}
	fake.status = http.StatusBadGateway
	if _, err := c.Check(context.Background(), request(), nil); err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("502: %v", err)
	}
	fake.status = 0
	fake.body["releases"] = "{not json"
	if _, err := c.Check(context.Background(), request(), nil); err == nil {
		t.Fatal("bad JSON accepted")
	}
	srv.Close()
	_, err = c.Check(context.Background(), request(), nil)
	if err == nil || strings.Contains(err.Error(), "topsecret") {
		t.Fatalf("network error: %v", err)
	}
}

func TestNormalizeAndAreas(t *testing.T) {
	for in, want := range map[string]string{
		"heroiclabs/nakama":                              "heroiclabs/nakama",
		"https://github.com/copilot-rast/nakama-farsi":   "copilot-rast/nakama-farsi",
		"https://github.com/copilot-rast/nakama-farsi/":  "copilot-rast/nakama-farsi",
		"git@github.com:copilot-rast/nakama-farsi.git":   "copilot-rast/nakama-farsi",
		"https://github.com/a/b/c":                       "",
		"../etc":                                         "",
		"a/..":                                           "",
		"https://evil.example.com/heroiclabs/nakama.git": "",
	} {
		got, ok := NormalizeRepo(in)
		if got != want || ok != (want != "") {
			t.Errorf("NormalizeRepo(%q) = %q %v", in, got, ok)
		}
	}
	for b, ok := range map[string]bool{"main": true, "release/3.42": true, "a..b": false, "": false, "x y": false, "/x": false} {
		if ValidBranch(b) != ok {
			t.Errorf("ValidBranch(%q)", b)
		}
	}
	for p, want := range map[string]string{
		"migrate/sql/x.sql": "migrations", "migrate/migrate.go": "core", "integrations/base/a.go": "plugins", "samples/stargame/x.go": "plugins",
		"console/ui/plane/src/a.js": "console", "go.sum": "deps", "vendor/x/y.go": "deps", "build/Dockerfile": "build", "docker-compose.yml": "build",
		".github/workflows/x.yml": "build", "docs/x.md": "docs", "CHANGELOG.md": "docs", "e2e/specs/a.ts": "e2e", "server/a_test.go": "e2e",
		"server/api.go": "core", "main.go": "core", "something.txt": "other",
	} {
		if got := AreaOf(p); got != want {
			t.Errorf("AreaOf(%q) = %q want %q", p, got, want)
		}
	}
	if Breaking("Fix the leaderboard reset") || !Breaking("chore!: x") || !Breaking("This is a breaking change") || !Breaking("Deprecated the old API") {
		t.Fatal("breaking keywords")
	}
}
