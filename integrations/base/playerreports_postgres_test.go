package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

func (h *harness) reportTitle(key string) string {
	h.t.Helper()
	out := h.mustAPI(201, "POST", "/report-titles", map[string]any{"key": key, "name": map[string]string{"fa": "عنوان " + key, "en": "Title " + key}})
	id := out["id"].(string)
	h.t.Cleanup(func() { h.s.db.Exec(`DELETE FROM base_report_title WHERE id=$1`, id) })
	return id
}

func resetPlayerReports(h *harness) {
	h.s.db.Exec(`DELETE FROM base_player_report`)
	h.s.db.Exec(`DELETE FROM cluster_rate_limit WHERE key LIKE 'player-report:%'`)
	h.s.db.Exec(`DELETE FROM base_restriction`)
}

func TestPlayerReportClientRules(t *testing.T) {
	h := newHarness(t)
	resetPlayerReports(h)
	key := "pr_cheat_" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	titleID := h.reportTitle(key)
	other := h.reportTitle(key + "_b")
	a, b := h.user("fa"), h.user("en")

	if code := h.rpcCode(a, "base_report_player", map[string]any{"userId": a, "titleKey": key}); code != "CANNOT_REPORT_SELF" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "base_report_player", map[string]any{"userId": uuid.Must(uuid.NewV4()).String(), "titleKey": key}); code != "USER_NOT_FOUND" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "base_report_player", map[string]any{"userId": b, "titleKey": "nope"}); code != "REPORT_TITLE_NOT_FOUND" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "base_report_player", map[string]any{"userId": b, "titleKey": key, "note": strings.Repeat("ب", 1001)}); code != "TEXT_TOO_LONG" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "base_report_player", map[string]any{"userId": b, "titleKey": key, "matchId": strings.Repeat("m", 129)}); code != "INVALID_CONTEXT" {
		t.Fatal(code)
	}

	out := h.mustRPC(a, "base_report_player", map[string]any{"userId": b, "titleKey": key, "note": strings.Repeat("ب", 1000), "matchId": "m1.node", "contentId": "c-9", "gameKey": "stargame"})
	if out["duplicate"] != false || out["reportId"] == "" {
		t.Fatalf("first %v", out)
	}
	first := out["reportId"].(string)
	again := h.mustRPC(a, "base_report_player", map[string]any{"userId": b, "titleKey": key})
	if again["duplicate"] != true || again["reportId"] != first {
		t.Fatalf("duplicate %v", again)
	}
	diff := h.mustRPC(a, "base_report_player", map[string]any{"userId": b, "titleKey": key + "_b"})
	if diff["duplicate"] != false {
		t.Fatalf("another reason is a new report: %v", diff)
	}
	h.s.db.Exec(`UPDATE base_player_report SET status='dismissed' WHERE id=$1`, first)
	if h.mustRPC(a, "base_report_player", map[string]any{"userId": b, "titleKey": key})["duplicate"] != false {
		t.Fatal("a closed report must not block a new one")
	}

	mine := h.mustRPC(a, "base_my_reports", map[string]any{})
	list := items(mine)
	if len(list) != 3 || mine["total"].(float64) != 3 || mine["lang"] != "fa" {
		t.Fatalf("mine %v", mine)
	}
	found := false
	for _, it := range list {
		if it["id"] == first {
			found = true
			if it["status"] != "dismissed" || it["title"] != "عنوان "+key || it["gameKey"] != "stargame" || it["matchId"] != "m1.node" || it["userId"] != b {
				t.Fatalf("item %v", it)
			}
			if _, leak := it["action"]; leak {
				t.Fatal("the reporter must not see the decision details")
			}
		}
	}
	if !found {
		t.Fatal("first report missing")
	}
	if len(items(h.mustRPC(b, "base_my_reports", map[string]any{}))) != 0 {
		t.Fatal("the reported player sees reports made against them")
	}

	used := h.mustAPI(200, "GET", "/report-titles/"+titleID, nil)
	if used["used"].(float64) != 2 {
		t.Fatalf("title used %v", used["used"])
	}
	_ = other

	h.s.db.Exec(`DELETE FROM cluster_rate_limit WHERE key LIKE 'player-report:%'`)
	c := h.user("fa")
	var wg sync.WaitGroup
	var mu sync.Mutex
	codes := map[string]int{}
	for i := 0; i < playerReportLimit+5; i++ {
		target := h.user("fa")
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.rpc(c, "base_report_player", map[string]any{"userId": target, "titleKey": key})
			code := "ok"
			if app, ok := asApp(err); ok {
				code = app.Code
			} else if err != nil {
				code = err.Error()
			}
			mu.Lock()
			codes[code]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	if codes["ok"] != playerReportLimit || codes["TOO_MANY_REPORTS"] != 5 {
		t.Fatalf("rate limit %v", codes)
	}

	d := h.user("fa")
	codes = map[string]int{}
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := h.rpc(d, "base_report_player", map[string]any{"userId": b, "titleKey": key})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				codes[err.Error()]++
			} else {
				codes[fmt.Sprint(out["duplicate"])]++
			}
		}()
	}
	wg.Wait()
	if codes["false"] != 1 || codes["true"] != 5 {
		t.Fatalf("parallel duplicates %v", codes)
	}
}

func TestPlayerReportReview(t *testing.T) {
	h := newHarness(t)
	resetPlayerReports(h)
	key := "pr_abuse_" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	h.reportTitle(key)
	a, b, c := h.user("fa"), h.user("fa"), h.user("fa")
	r1 := h.mustRPC(a, "base_report_player", map[string]any{"userId": c, "titleKey": key, "note": "فحاشی در بازی"})["reportId"].(string)
	r2 := h.mustRPC(b, "base_report_player", map[string]any{"userId": c, "titleKey": key})["reportId"].(string)
	r3 := h.mustRPC(a, "base_report_player", map[string]any{"userId": b, "titleKey": key})["reportId"].(string)

	list := h.mustAPI(200, "GET", "/player-reports?status=open", nil)
	if list["total"].(float64) != 3 || list["openTotal"].(float64) != 3 {
		t.Fatalf("list %v", list)
	}
	if l := h.mustAPI(200, "GET", "/player-reports?targetId="+c, nil); l["total"].(float64) != 2 {
		t.Fatalf("by target %v", l)
	}
	h.expectError(400, "INVALID_FILTER", "GET", "/player-reports?status=nope", nil)
	h.expectError(404, "REPORT_NOT_FOUND", "GET", "/player-reports/"+uuid.Must(uuid.NewV4()).String(), nil)

	got := h.mustAPI(200, "GET", "/player-reports/"+r1, nil)
	if got["note"] != "فحاشی در بازی" || got["status"] != "open" || len(got["others"].([]any)) != 1 || got["target"].(map[string]any)["counts"].(map[string]any)["reports"].(float64) != 2 {
		t.Fatalf("detail %v", got)
	}

	rev := h.mustAPI(200, "POST", "/player-reports/"+r1+"/status", map[string]any{"status": "reviewing"})
	if rev["status"] != "reviewing" || rev["reviewStartedAt"] == nil {
		t.Fatalf("reviewing %v", rev)
	}
	h.expectError(400, "INVALID_REPORT_STATUS", "POST", "/player-reports/"+r1+"/status", map[string]any{"status": "resolved"})
	sum := h.mustAPI(200, "GET", "/player-reports/summary", nil)
	if sum["open"].(float64) != 2 || sum["reviewing"].(float64) != 1 || sum["active"].(float64) != 3 {
		t.Fatalf("summary %v", sum)
	}
	h.expectError(400, "TEXT_REQUIRED", "POST", "/player-reports/"+r1+"/notes", map[string]any{"text": "  "})
	noted := h.mustAPI(200, "POST", "/player-reports/"+r1+"/notes", map[string]any{"text": "ویدیو را دیدم"})
	if notes := noted["notes"].([]any); len(notes) != 1 || notes[0].(map[string]any)["text"] != "ویدیو را دیدم" {
		t.Fatalf("notes %v", noted["notes"])
	}

	h.expectError(400, "END_REQUIRED", "POST", "/player-reports/"+r1+"/resolve", map[string]any{"action": "mute", "reason": map[string]string{"fa": "سکوت"}})
	h.expectError(400, "INVALID_MOD_ACTION", "POST", "/player-reports/"+r1+"/resolve", map[string]any{"action": "delete"})
	done := h.mustAPI(200, "POST", "/player-reports/"+r1+"/resolve", map[string]any{"action": "mute", "minutes": 60, "reason": map[string]string{"fa": "به خاطر فحاشی"}, "note": "قطعی", "closeOthers": true})
	if done["status"] != "resolved" || done["action"] != "mute" || done["restrictionId"] == nil || done["resolutionNote"] != "قطعی" {
		t.Fatalf("resolved %v", done)
	}
	mute, err := restrict.Active(h.ctx, h.s.db, c, restrict.KindMute)
	if err != nil || mute == nil {
		t.Fatalf("mute not active: %v", err)
	}
	var reportRef string
	h.s.db.QueryRow(`SELECT report_id::text FROM base_restriction WHERE id=$1`, done["restrictionId"]).Scan(&reportRef)
	if reportRef != r1 {
		t.Fatalf("restriction report id %q", reportRef)
	}
	if o := h.mustAPI(200, "GET", "/player-reports/"+r2, nil); o["status"] != "resolved" || o["action"] != "mute" {
		t.Fatalf("closeOthers left %v", o["status"])
	}
	if o := h.mustAPI(200, "GET", "/player-reports/"+r3, nil); o["status"] != "open" {
		t.Fatalf("a report about another player was closed: %v", o["status"])
	}
	h.expectError(409, "REPORT_CLOSED", "POST", "/player-reports/"+r1+"/resolve", map[string]any{"action": "dismiss"})
	h.expectError(409, "REPORT_CLOSED", "POST", "/player-reports/"+r1+"/status", map[string]any{"status": "open"})
	f := h.fake()
	if len(f.notes) == 0 || f.notes[len(f.notes)-1].user != c || f.notes[len(f.notes)-1].code != restrict.NotifyMuted {
		t.Fatalf("live notice %+v", f.notes)
	}

	dis := h.mustAPI(200, "POST", "/player-reports/"+r3+"/resolve", map[string]any{"action": "dismiss", "note": "مدرکی نبود"})
	if dis["status"] != "dismissed" || dis["restrictionId"] != nil {
		t.Fatalf("dismiss %v", dis)
	}
	player := h.mustAPI(200, "GET", "/players/"+c, nil)
	if pr := player["playerReports"].(map[string]any); pr["against"].(float64) != 2 || pr["open"].(float64) != 0 {
		t.Fatalf("player card counts %v", pr)
	}
	var acts int
	h.s.db.QueryRow(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND kind='report' AND action='received'`, c).Scan(&acts)
	if acts != 2 {
		t.Fatalf("activity rows %d", acts)
	}
}

func TestPlayerReportRules(t *testing.T) {
	h := newHarness(t)
	resetPlayerReports(h)
	t.Cleanup(func() {
		h.s.db.Exec(`UPDATE base_report_rules SET per_hour=10, duplicate_hours=24, min_account_hours=0, note_required=false, trusted_tag_ids='{}', trusted_per_hour=30 WHERE id=1`)
	})
	key := "pr_rule_" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	h.reportTitle(key)
	got := h.mustAPI(200, "GET", "/report-rules", nil)
	if got["perHour"].(float64) != 10 || got["duplicateHours"].(float64) != 24 || got["noteRequired"] != false {
		t.Fatalf("defaults %v", got)
	}
	h.expectError(400, "INVALID_RULE", "PUT", "/report-rules", map[string]any{"perHour": -1})
	a, b, c := h.user("fa"), h.user("fa"), h.user("fa")
	trustedTag := h.userTag("trusted"+key, c)
	h.mustAPI(200, "PUT", "/report-rules", map[string]any{"perHour": 1, "duplicateHours": 0, "minAccountHours": 0, "noteRequired": true, "trustedTagIds": []string{trustedTag}, "trustedPerHour": 0})
	if code := h.rpcCode(a, "base_report_player", map[string]any{"userId": b, "titleKey": key}); code != "NOTE_REQUIRED" {
		t.Fatal(code)
	}
	first := h.mustRPC(a, "base_report_player", map[string]any{"userId": b, "titleKey": key, "note": "x"})
	if code := h.rpcCode(a, "base_report_player", map[string]any{"userId": b, "titleKey": key, "note": "x"}); code != "TOO_MANY_REPORTS" {
		t.Fatalf("no duplicate window: the repeat is a new report and hits the limit: %s %v", code, first)
	}
	mine := h.mustRPC(a, "base_my_reports", map[string]any{})
	if l := mine["limits"].(map[string]any); l["perHour"].(float64) != 1 || l["remaining"].(float64) != 0 || l["noteRequired"] != true || l["trusted"] != false {
		t.Fatalf("limits %v", l)
	}
	for i := 0; i < 5; i++ {
		h.mustRPC(c, "base_report_player", map[string]any{"userId": b, "titleKey": key, "note": "x"})
	}
	if l := h.mustRPC(c, "base_my_reports", map[string]any{})["limits"].(map[string]any); l["trusted"] != true || l["remaining"] != nil {
		t.Fatalf("trusted unlimited %v", l)
	}
	h.mustAPI(200, "PUT", "/report-rules", map[string]any{"perHour": 10, "duplicateHours": 24, "minAccountHours": 48, "trustedPerHour": 30})
	if code := h.rpcCode(h.user("fa"), "base_report_player", map[string]any{"userId": b, "titleKey": key}); code != "ACCOUNT_TOO_NEW" {
		t.Fatal(code)
	}
}
