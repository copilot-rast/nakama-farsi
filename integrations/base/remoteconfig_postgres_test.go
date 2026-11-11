package main

import (
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/activity"
)

func TestRemoteConfigConsoleAndClient(t *testing.T) {
	h := newHarness(t)
	for _, q := range []string{`DELETE FROM base_config_key WHERE key LIKE 'zt.%'`, `DELETE FROM base_config_history WHERE key LIKE 'zt.%'`} {
		if _, err := h.s.db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		h.s.db.Exec(`DELETE FROM base_config_key WHERE key LIKE 'zt.%'`)
		h.s.db.Exec(`DELETE FROM base_config_history WHERE key LIKE 'zt.%'`)
	})
	vip := h.mustAPI(201, "POST", "/tags", map[string]any{"key": "zt-vip", "keyType": "user", "name": map[string]string{"fa": "ویژه"}})
	product := h.mustAPI(201, "POST", "/tags", map[string]any{"key": "zt-prod", "keyType": "product", "name": map[string]string{"fa": "محصول"}})
	h.expectError(400, "INVALID_KEY", "POST", "/remote-config", map[string]any{"key": "Bad Key", "type": "bool", "value": true})
	h.expectError(400, "INVALID_TYPE", "POST", "/remote-config", map[string]any{"key": "zt.x", "type": "date", "value": true})
	h.expectError(400, "INVALID_VALUE", "POST", "/remote-config", map[string]any{"key": "zt.x", "type": "bool", "value": "yes"})
	h.expectError(400, "DEFAULT_LANGUAGE_REQUIRED", "POST", "/remote-config", map[string]any{"key": "zt.x", "type": "text", "value": map[string]string{"en": "Hi"}})
	h.expectError(400, "TAG_NOT_FOUND", "POST", "/remote-config", map[string]any{"key": "zt.x", "type": "bool", "value": false, "rules": []map[string]any{{"value": true, "tagIds": []string{product["id"].(string)}}}})
	h.expectError(400, "MIN_ABOVE_MAX", "POST", "/remote-config", map[string]any{"key": "zt.x", "type": "bool", "value": false, "rules": []map[string]any{{"value": true, "minVersion": "2.0", "maxVersion": "1.0"}}})
	h.expectError(400, "INVALID_PERCENT", "POST", "/remote-config", map[string]any{"key": "zt.x", "type": "bool", "value": false, "rules": []map[string]any{{"value": true, "percent": 120}}})

	banner := h.mustAPI(201, "POST", "/remote-config", map[string]any{"key": "zt.banner", "type": "text", "isActive": true, "description": "home banner",
		"value": map[string]string{"fa": "سلام", "en": "Hello"},
		"rules": []map[string]any{{"name": "vip", "value": map[string]string{"fa": "سلام ویژه", "en": "Hello VIP"}, "tagIds": []string{vip["id"].(string)}}}})
	sw := h.mustAPI(201, "POST", "/remote-config", map[string]any{"key": "zt.rush", "type": "bool", "isActive": true, "value": false,
		"rules": []map[string]any{{"name": "web-new", "value": true, "platforms": []string{"web"}, "minVersion": "1.4"}}})
	if banner["version"] != float64(1) || sw["rules"].([]any)[0].(map[string]any)["name"] != "web-new" {
		t.Fatalf("created %v %v", banner, sw)
	}
	h.expectError(409, "ALREADY_EXISTS", "POST", "/remote-config", map[string]any{"key": "zt.rush", "type": "bool", "value": false, "isActive": true})

	u := h.user("fa")
	out := h.mustRPC(u, "base_remote_config", map[string]any{"platform": "web", "version": "1.3", "keys": []string{"zt.banner", "zt.rush"}})
	vals := out["values"].(map[string]any)
	if vals["zt.banner"] != "سلام" || vals["zt.rush"] != false {
		t.Fatalf("default values %v", out)
	}
	out = h.mustRPC(u, "base_remote_config", map[string]any{"platform": "web", "version": "1.4.2", "lang": "en", "keys": []string{"zt.banner", "zt.rush"}})
	if vals := out["values"].(map[string]any); vals["zt.banner"] != "Hello" || vals["zt.rush"] != true || out["variants"].(map[string]any)["zt.rush"] != "web-new" {
		t.Fatalf("web 1.4.2 %v", out)
	}
	h.mustAPI(200, "PUT", "/players/"+u+"/tags", map[string]any{"tagIds": []string{vip["id"].(string)}})
	out = h.mustRPC(u, "base_remote_config", map[string]any{"keys": []string{"zt.banner"}})
	if out["values"].(map[string]any)["zt.banner"] != "سلام ویژه" {
		t.Fatalf("vip banner %v", out)
	}
	if code := h.rpcCode(u, "base_remote_config", map[string]any{"platform": "tv"}); code != "INVALID_PLATFORM" {
		t.Fatal(code)
	}

	h.mustRPC(u, "base_remote_config", map[string]any{"platform": "web", "version": "1.0"})
	h.mustRPC(u, "base_remote_config", map[string]any{"platform": "web", "version": "1.0"})
	var changes int
	h.s.db.QueryRow(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND kind=$2`, u, activity.KindConfig).Scan(&changes)
	if changes != 0 {
		t.Fatalf("first fetch counted as change: %d", changes)
	}
	upd := h.mustAPI(200, "PUT", "/remote-config/"+sw["id"].(string), map[string]any{"key": "zt.rush", "type": "bool", "isActive": true, "value": true, "rules": []any{}})
	if upd["version"] != float64(2) {
		t.Fatalf("update %v", upd)
	}
	h.expectError(409, "TYPE_LOCKED", "PUT", "/remote-config/"+sw["id"].(string), map[string]any{"key": "zt.rush", "type": "number", "isActive": true, "value": 1})
	h.mustRPC(u, "base_remote_config", map[string]any{"platform": "web", "version": "1.0"})
	h.s.db.QueryRow(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND kind=$2 AND data::text LIKE '%zt.rush:default%'`, u, activity.KindConfig).Scan(&changes)
	if changes != 1 {
		t.Fatalf("config change activity %d", changes)
	}

	hist := items(h.mustAPI(200, "GET", "/remote-config/"+sw["id"].(string)+"/history", nil))
	if len(hist) != 2 || hist[0]["version"] != float64(2) || hist[1]["action"] != "create" {
		t.Fatalf("history %v", hist)
	}
	back := h.mustAPI(200, "POST", "/remote-config/"+sw["id"].(string)+"/rollback", map[string]any{"version": 1})
	if back["version"] != float64(3) || back["value"] != false || len(back["rules"].([]any)) != 1 {
		t.Fatalf("rollback %v", back)
	}
	h.expectError(404, "VERSION_NOT_FOUND", "POST", "/remote-config/"+sw["id"].(string)+"/rollback", map[string]any{"version": 99})
	h.mustAPI(200, "DELETE", "/remote-config/"+sw["id"].(string), nil)
	restored := h.mustAPI(201, "POST", "/remote-config/"+sw["id"].(string)+"/rollback", map[string]any{"version": 3})
	if restored["key"] != "zt.rush" || restored["version"] != float64(5) {
		t.Fatalf("restore deleted %v", restored)
	}
	prev := h.mustAPI(200, "POST", "/remote-config/preview", map[string]any{"userId": u, "platform": "web", "version": "2.0"})
	if prev["values"].(map[string]any)["zt.rush"] != true || prev["buckets"].(map[string]any)["zt.rush"] == nil {
		t.Fatalf("preview %v", prev)
	}
}

func TestActivityLogAndRetention(t *testing.T) {
	h := newHarness(t)
	u := h.user("fa")
	for i, at := range []time.Duration{0, 10 * 24 * time.Hour, 400 * 24 * time.Hour} {
		if _, err := h.s.db.Exec(`INSERT INTO base_activity (user_id, kind, action, ip, country, created_at) VALUES ($1,'signin','device','10.0.0.1','IR', now() - $2::INT8 * INTERVAL '1 second')`, u, int64(at.Seconds())); err != nil {
			t.Fatal(i, err)
		}
	}
	h.s.record(h.ctx, activity.Entry{UserID: u, Kind: activity.KindTicket, Action: "created", RefID: "t1", Data: map[string]any{"number": "T-AAAAAA"}})
	h.s.record(h.ctx, activity.Entry{UserID: u, Kind: "nope", Action: "x"})
	list := items(h.mustAPI(200, "GET", "/activity?userId="+u, nil))
	if len(list) != 4 || list[0]["kind"] != "ticket" {
		t.Fatalf("activity %v", list)
	}
	if got := items(h.mustAPI(200, "GET", "/activity?userId="+u+"&kind=signin&country=ir", nil)); len(got) != 3 {
		t.Fatalf("filtered %d", len(got))
	}
	if got := items(h.mustAPI(200, "GET", "/activity?userId="+u+"&q=T-AAAA", nil)); len(got) != 1 {
		t.Fatalf("search %d", len(got))
	}
	h.expectError(400, "INVALID_DATE", "GET", "/activity?from=yesterday", nil)
	mine := h.mustRPC(u, "base_my_activity", map[string]any{"limit": 2})
	if len(mine["items"].([]any)) != 2 || mine["before"] == nil {
		t.Fatalf("mine %v", mine)
	}
	if _, err := h.s.db.Exec(`UPDATE base_support_settings SET activity_retention_days=30`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.s.db.Exec(`UPDATE base_support_settings SET activity_retention_days=180`) })
	n, err := h.s.sweepActivity(h.ctx)
	if err != nil || n < 1 {
		t.Fatalf("sweep %d %v", n, err)
	}
	if got := items(h.mustAPI(200, "GET", "/activity?userId="+u, nil)); len(got) != 3 {
		t.Fatalf("after sweep %d", len(got))
	}
	sum := h.mustAPI(200, "GET", "/activity/summary?userId="+u, nil)
	if sum["counts"].(map[string]any)["signin"] != float64(2) || sum["retentionDays"] != float64(30) {
		t.Fatalf("summary %v", sum)
	}
}
