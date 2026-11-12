package main

import (
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (h *harness) bannerReset() {
	h.t.Helper()
	for _, q := range []string{
		`TRUNCATE publication_banner_seen, publication_banner_day, publication_banner, publication_banner_list CASCADE`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'bnr:%'`,
		`DELETE FROM base_tag WHERE key LIKE 'bt_%'`,
		`DELETE FROM economy_level_table`,
		`DELETE FROM economy_meta WHERE key='level'`,
		`UPDATE economy_event_settings SET timezone='Asia/Tehran' WHERE id=1`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *harness) tag(kind, key string) string {
	h.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_tag (id, key_type, key, name) VALUES ($1,$2,$3,'{"fa":"ت"}')`, id, kind, key); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *harness) bannerList(key string) string {
	h.t.Helper()
	out := h.mustAPI(201, "POST", "/banner-lists", map[string]any{"key": key, "name": map[string]string{"fa": "خانه", "en": "Home"}, "layout": "carousel", "maxItems": 3, "rotateSeconds": 4})
	return out["id"].(string)
}

func (h *harness) banner(list, key string, extra map[string]any) string {
	h.t.Helper()
	body := map[string]any{"key": key, "listId": list, "title": map[string]string{"fa": "عنوان " + key, "en": "Title " + key}}
	for k, v := range extra {
		body[k] = v
	}
	out := h.mustAPI(201, "POST", "/banners", body)
	return out["id"].(string)
}

func bannerKeys(m map[string]any) []string {
	out := []string{}
	for _, it := range items(m) {
		out = append(out, it.(map[string]any)["key"].(string))
	}
	return out
}

func sameKeys(a []string, b ...string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestBannerConsoleValidation(t *testing.T) {
	h := newHarness(t)
	h.bannerReset()
	if code, out := h.api("POST", "/banner-lists", map[string]any{"key": "Home Top", "name": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "INVALID_KEY" {
		t.Fatalf("bad key = %d %v", code, out)
	}
	if code, out := h.api("POST", "/banner-lists", map[string]any{"key": "home_top", "name": map[string]string{"fa": "x"}, "layout": "wall"}); code != 400 || out["error"] != "INVALID_LAYOUT" {
		t.Fatalf("bad layout = %d %v", code, out)
	}
	list := h.bannerList("home_top")
	if code, out := h.api("POST", "/banner-lists", map[string]any{"key": "home_top", "name": map[string]string{"fa": "x"}}); code != 409 || out["error"] != "KEY_EXISTS" {
		t.Fatalf("dup key = %d %v", code, out)
	}
	userTag := h.tag("user", "bt_vip")
	newsTag := h.tag("news", "bt_news")
	bannerTag := h.tag("banner", "bt_promo")
	cases := []struct {
		body map[string]any
		code string
	}{
		{map[string]any{"key": "b1", "listId": list}, "DEFAULT_LANGUAGE_REQUIRED"},
		{map[string]any{"key": "b1", "listId": uuid.Must(uuid.NewV4()).String(), "title": map[string]string{"fa": "x"}}, "BANNER_LIST_NOT_FOUND"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "action": map[string]any{"type": "url", "url": map[string]string{"fa": "javascript:alert(1)"}}}, "INVALID_URL"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "action": map[string]any{"type": "url", "url": map[string]string{"fa": "/relative"}}}, "INVALID_URL"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "action": map[string]any{"type": "product", "value": "nope"}}, "ACTION_TARGET_NOT_FOUND"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "action": map[string]any{"type": "screen", "value": "<script>"}}, "INVALID_ACTION"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "action": map[string]any{"type": "teleport"}}, "INVALID_ACTION"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "startsAt": "2026-12-02T00:00:00Z", "endsAt": "2026-12-01T00:00:00Z"}, "INVALID_WINDOW"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "dayStart": "10:00"}, "INVALID_HOURS"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "dayStart": "18:00", "dayEnd": "09:00"}, "INVALID_HOURS"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "dayStart": "25:00", "dayEnd": "26:00"}, "INVALID_TIME_OF_DAY"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "days": []int{7}}, "INVALID_WEEKDAY"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "audienceTagIds": []string{newsTag}}, "TAG_KIND_MISMATCH"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "tagIds": []string{userTag}}, "TAG_KIND_MISMATCH"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "audienceTagIds": []string{userTag}, "excludeTagIds": []string{userTag}}, "TAG_BOTH_WAYS"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "minLevel": 5, "maxLevel": 2}, "INVALID_RANGE"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "minVersion": "2.0", "maxVersion": "1.0"}, "INVALID_RANGE"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "platforms": []string{"psp"}}, "INVALID_PLATFORM"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "countries": []string{"IRN"}}, "INVALID_COUNTRY_CODE"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "languages": []string{"zz"}}, "INVALID_LANGUAGE"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "imageUrl": map[string]string{"en": "https://x/y.png"}}, "DEFAULT_LANGUAGE_REQUIRED"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "bgColor": "red"}, "INVALID_COLOR"},
		{map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "x"}, "playerAge": "old"}, "INVALID_PLAYER_AGE"},
	}
	for i, c := range cases {
		if code, out := h.api("POST", "/banners", c.body); code < 400 || out["error"] != c.code {
			t.Fatalf("case %d = %d %v, want %s", i, code, out, c.code)
		}
	}
	id := h.banner(list, "b1", map[string]any{"tagIds": []string{bannerTag}, "dayStart": "۰۹:۰۰", "dayEnd": "24:00", "days": []int{6, 0, 6}, "platforms": []string{"WEB"}, "countries": []string{"ir"},
		"imageUrl": map[string]string{"fa": "https://cdn.example/fa.png", "en": "https://cdn.example/en.png"},
		"action": map[string]any{"type": "url", "url": map[string]string{"fa": "https://example.com/fa", "en": "https://example.com/en"}, "label": map[string]string{"fa": "برو"}}})
	got := h.mustAPI(200, "GET", "/banners/"+id, nil)
	if got["dayStart"] != "09:00" || got["dayEnd"] != "24:00" || got["state"] != "live" || got["listKey"] != "home_top" {
		t.Fatalf("banner = %v", got)
	}
	if d := got["days"].([]any); len(d) != 2 || num(d[0]) != 0 || num(d[1]) != 6 {
		t.Fatalf("days = %v", got["days"])
	}
	if p := got["platforms"].([]any); len(p) != 1 || p[0] != "web" {
		t.Fatalf("platforms = %v", got["platforms"])
	}
	if c := got["countries"].([]any); len(c) != 1 || c[0] != "IR" {
		t.Fatalf("countries = %v", got["countries"])
	}
	body := map[string]any{"key": "b2", "listId": list, "title": map[string]string{"fa": "x"}}
	h.mustAPI(200, "PUT", "/banners/"+id, map[string]any{"key": "b1", "listId": list, "title": map[string]string{"fa": "تازه"}})
	if code, out := h.api("PUT", "/banners/"+id, body); code != 400 || out["error"] != "KEY_LOCKED" {
		t.Fatalf("key change = %d %v", code, out)
	}
	if code, out := h.api("DELETE", "/banner-lists/"+list, nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete used list = %d %v", code, out)
	}
	listed := h.mustAPI(200, "GET", "/banners?state=live&listId="+list, nil)
	if num(listed["total"]) != 1 {
		t.Fatalf("live list = %v", listed)
	}
	if num(h.mustAPI(200, "GET", "/banners?tagId="+bannerTag, nil)["total"]) != 0 {
		t.Fatal("tag filter should be empty after the update dropped tags")
	}
	h.mustAPI(200, "DELETE", "/banners/"+id, nil)
	h.mustAPI(200, "DELETE", "/banner-lists/"+list, nil)
}

func TestBannerTargetingAndOrder(t *testing.T) {
	h := newHarness(t)
	h.bannerReset()
	list := h.bannerList("menu_carousel")
	vip := h.tag("user", "bt_vip")
	banned := h.tag("user", "bt_banned")
	old := h.user(map[string]any{"country": "ir"}, "", "fa")
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE users SET create_time = now() - interval '30 days' WHERE id=$1`, old); err != nil {
		t.Fatal(err)
	}
	fresh := h.user(map[string]any{"country": "DE"}, "", "en")
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, old, vip); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	h.banner(list, "low", map[string]any{"priority": 1})
	h.banner(list, "high", map[string]any{"priority": 50})
	h.banner(list, "vip_only", map[string]any{"priority": 10, "audienceTagIds": []string{vip}})
	h.banner(list, "not_banned", map[string]any{"priority": 9, "excludeTagIds": []string{banned}})
	h.banner(list, "android", map[string]any{"priority": 8, "platforms": []string{"android"}})
	h.banner(list, "new_version", map[string]any{"priority": 7, "minVersion": "1.4.0"})
	h.banner(list, "iran", map[string]any{"priority": 6, "countries": []string{"IR"}})
	h.banner(list, "english", map[string]any{"priority": 5, "languages": []string{"en"}})
	h.banner(list, "newcomers", map[string]any{"priority": 4, "playerAge": "new", "newDays": 3})
	h.banner(list, "returning", map[string]any{"priority": 3, "playerAge": "returning", "newDays": 3})
	h.banner(list, "future", map[string]any{"priority": 100, "startsAt": now.Add(time.Hour).Format(time.RFC3339)})
	h.banner(list, "past", map[string]any{"priority": 100, "endsAt": now.Add(-time.Minute).Format(time.RFC3339)})
	h.banner(list, "off", map[string]any{"priority": 100, "isActive": false})
	h.banner(list, "level3", map[string]any{"priority": 2, "minLevel": 3})

	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_banner_list SET max_items=20 WHERE id=$1`, list); err != nil {
		t.Fatal(err)
	}
	a := bannerKeys(h.mustRPC(old, "publication_banners", map[string]any{"placement": "menu_carousel", "platform": "web", "version": "1.4.1"}))
	if !sameKeys(a, "high", "vip_only", "not_banned", "new_version", "iran", "returning", "low") {
		t.Fatalf("old player sees %v", a)
	}
	b := bannerKeys(h.mustRPC(fresh, "publication_banners", map[string]any{"placement": "menu_carousel", "platform": "android", "version": "1.3"}))
	if !sameKeys(b, "high", "not_banned", "android", "english", "newcomers", "low") {
		t.Fatalf("fresh player sees %v", b)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, fresh, banned); err != nil {
		t.Fatal(err)
	}
	lv := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_meta (id, key, name) VALUES ($1,'level','{"fa":"سطح"}')`, lv); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ($1,$2,3)`, fresh, lv); err != nil {
		t.Fatal(err)
	}
	b = bannerKeys(h.mustRPC(fresh, "publication_banners", map[string]any{"placement": "menu_carousel", "platform": "android", "version": "1.3", "lang": "fa"}))
	if !sameKeys(b, "high", "android", "newcomers", "level3", "low") {
		t.Fatalf("fresh player after tag/level/lang sees %v", b)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_banner_list SET max_items=2 WHERE id=$1`, list); err != nil {
		t.Fatal(err)
	}
	b = bannerKeys(h.mustRPC(fresh, "publication_banners", map[string]any{"placement": "menu_carousel", "platform": "android"}))
	if !sameKeys(b, "high", "android") {
		t.Fatalf("max items = %v", b)
	}
	preview := h.mustAPI(200, "GET", "/banner-lists/"+list+"/preview?userId="+fresh+"&platform=android&version=1.3", nil)
	reasons := map[string]string{}
	for _, x := range preview["hidden"].([]any) {
		m := x.(map[string]any)
		reasons[m["key"].(string)] = m["reason"].(string)
	}
	for k, want := range map[string]string{"vip_only": "tag", "not_banned": "excluded", "new_version": "version", "iran": "country", "returning": "returning", "future": "not_started", "past": "ended", "off": "inactive", "level3": "max_items"} {
		if reasons[k] != want {
			t.Fatalf("preview reason %s = %q, want %q (%v)", k, reasons[k], want, reasons)
		}
	}
	if code := h.rpcCode(old, "publication_banners", map[string]any{"placement": "nope_list"}); code != "BANNER_LIST_NOT_FOUND" {
		t.Fatalf("unknown placement = %s", code)
	}
	if code := h.rpcCode(old, "publication_banners", map[string]any{"placement": "menu_carousel", "platform": "psp"}); code != "INVALID_PLATFORM" {
		t.Fatalf("bad platform = %s", code)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_banner_list SET is_active=false WHERE id=$1`, list); err != nil {
		t.Fatal(err)
	}
	if code := h.rpcCode(old, "publication_banners", map[string]any{"placement": "menu_carousel"}); code != "BANNER_LIST_NOT_FOUND" {
		t.Fatalf("inactive placement = %s", code)
	}
}

func TestBannerScheduleInProjectZone(t *testing.T) {
	h := newHarness(t)
	h.bannerReset()
	list := h.bannerList("zone_list")
	u := h.user(nil, "", "fa")
	id := h.banner(list, "evening", map[string]any{"dayStart": "18:00", "dayEnd": "22:00", "days": []int{5}})
	zone, _ := time.LoadLocation("Asia/Tehran")
	v, err := h.s.bannerViewer(h.ctx, u, "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := loadBanners(h.ctx, h.s.db, `WHERE b.id=$2`, nil, id)
	if err != nil || len(rows) != 1 {
		t.Fatal(err, rows)
	}
	for _, c := range []struct {
		at   time.Time
		want string
	}{
		{time.Date(2026, 11, 13, 19, 0, 0, 0, zone), ""},
		{time.Date(2026, 11, 13, 22, 0, 0, 0, zone), "hours"},
		{time.Date(2026, 11, 13, 17, 59, 0, 0, zone), "hours"},
		{time.Date(2026, 11, 14, 19, 0, 0, 0, zone), "day"},
		{time.Date(2026, 11, 13, 15, 30, 0, 0, time.UTC), ""},
	} {
		v.Now = c.at
		got, err := rows[0].eligible(h.ctx, v)
		if err != nil || got != c.want {
			t.Fatalf("at %v = %q %v, want %q", c.at, got, err, c.want)
		}
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_event_settings SET timezone='UTC' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	v, _ = h.s.bannerViewer(h.ctx, u, "", "", "", "")
	v.Now = time.Date(2026, 11, 13, 15, 30, 0, 0, time.UTC)
	if got, _ := rows[0].eligible(h.ctx, v); got != "hours" {
		t.Fatalf("UTC zone = %q, want hours", got)
	}
}

func TestBannerCountersAreIdempotentAndEligible(t *testing.T) {
	h := newHarness(t)
	h.bannerReset()
	list := h.bannerList("count_list")
	vip := h.tag("user", "bt_vip")
	u := h.user(nil, "", "fa")
	other := h.user(nil, "", "fa")
	open := h.banner(list, "open", map[string]any{"action": map[string]any{"type": "screen", "value": "shop", "label": map[string]string{"fa": "بزن"}}})
	secret := h.banner(list, "secret", map[string]any{"audienceTagIds": []string{vip}})
	ended := h.banner(list, "ended", nil)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_banner SET ends_at = now() - interval '1 second' WHERE id=$1`, ended); err != nil {
		t.Fatal(err)
	}
	out := h.mustRPC(u, "publication_banner_view", map[string]any{"bannerIds": []string{open, secret, ended, open}})
	if len(out["counted"].([]any)) != 1 || len(out["skipped"].([]any)) != 2 {
		t.Fatalf("first view = %v", out)
	}
	out = h.mustRPC(u, "publication_banner_view", map[string]any{"bannerIds": []string{open}})
	if len(out["counted"].([]any)) != 0 {
		t.Fatalf("second view counted = %v", out)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = h.rpc(other, "publication_banner_view", map[string]any{"bannerIds": []string{open}})
			_, _ = h.rpc(other, "publication_banner_click", map[string]any{"bannerId": open})
		}()
	}
	wg.Wait()
	click := h.mustRPC(u, "publication_banner_click", map[string]any{"bannerId": open})
	if click["counted"] != true || click["action"].(map[string]any)["value"] != "shop" || click["action"].(map[string]any)["label"] != "بزن" {
		t.Fatalf("click = %v", click)
	}
	if again := h.mustRPC(u, "publication_banner_click", map[string]any{"bannerId": open}); again["counted"] != false {
		t.Fatalf("second click = %v", again)
	}
	if code := h.rpcCode(u, "publication_banner_click", map[string]any{"bannerId": secret}); code != "BANNER_NOT_AVAILABLE" {
		t.Fatalf("click secret = %s", code)
	}
	if code := h.rpcCode(u, "publication_banner_click", map[string]any{"bannerId": ended}); code != "BANNER_NOT_AVAILABLE" {
		t.Fatalf("click ended = %s", code)
	}
	if code := h.rpcCode(u, "publication_banner_view", map[string]any{"bannerIds": []string{"x"}}); code != "INVALID_ID" {
		t.Fatalf("bad id = %s", code)
	}
	var views, clicks int
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT COALESCE(sum(views),0), COALESCE(sum(clicks),0) FROM publication_banner_day WHERE banner_id=$1`, open).Scan(&views, &clicks); err != nil {
		t.Fatal(err)
	}
	if views != 2 || clicks != 2 {
		t.Fatalf("counters = %d views %d clicks, want 2 and 2", views, clicks)
	}
	var secretRows int
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT count(*) FROM publication_banner_day WHERE banner_id IN ($1,$2)`, secret, ended).Scan(&secretRows); err != nil || secretRows != 0 {
		t.Fatalf("ineligible banners counted: %d %v", secretRows, err)
	}
	stats := h.mustAPI(200, "GET", "/banners/"+open+"/stats?days=7", nil)
	if num(stats["views"]) != 2 || num(stats["clicks"]) != 2 || len(stats["days"].([]any)) != 7 || stats["timezone"] != "Asia/Tehran" {
		t.Fatalf("stats = %v", stats)
	}
	got := h.mustAPI(200, "GET", "/banners/"+open, nil)
	if num(got["views7"]) != 2 || num(got["clicks"]) != 2 {
		t.Fatalf("row counters = %v", got)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_banner_seen SET day = day - 1 WHERE user_id=$1`, u); err != nil {
		t.Fatal(err)
	}
	if out := h.mustRPC(u, "publication_banner_view", map[string]any{"bannerIds": []string{open}}); len(out["counted"].([]any)) != 1 {
		t.Fatalf("next day view = %v", out)
	}
	for i := 0; i < bannerCountPerMinute; i++ {
		_, _ = h.rpc(u, "publication_banner_view", map[string]any{"bannerIds": []string{open}})
	}
	if code := h.rpcCode(u, "publication_banner_view", map[string]any{"bannerIds": []string{open}}); code != "TOO_MANY_REQUESTS" {
		t.Fatalf("rate limit = %s", code)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_banner_seen SET day = day - 100 WHERE user_id=$1`, u); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM users WHERE id=$1`, other); err != nil {
		t.Fatal(err)
	}
	n, err := h.s.sweepBannerSeen(h.ctx)
	if err != nil || n != 3 {
		t.Fatalf("sweep = %d %v", n, err)
	}
}

func TestBannerActionTargetMustStayVisible(t *testing.T) {
	h := newHarness(t)
	h.bannerReset()
	list := h.bannerList("target_list")
	vip := h.tag("user", "bt_vip")
	u := h.user(nil, "", "fa")
	prize := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_prize (id, key, name, items) VALUES ($1,'bt_prize','{"fa":"ج"}','[]')`, prize); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = h.s.db.ExecContext(h.ctx, `TRUNCATE publication_banner CASCADE`)
		_, _ = h.s.db.ExecContext(h.ctx, `DELETE FROM economy_product WHERE key LIKE 'bt_%'`)
		_, _ = h.s.db.ExecContext(h.ctx, `DELETE FROM economy_prize WHERE key LIKE 'bt_%'`)
	})
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_product (id, key, name, is_active, prize_id) VALUES ($1,'bt_pack','{"fa":"بسته"}',true,$2)`, uuid.Must(uuid.NewV4()).String(), prize); err != nil {
		t.Fatal(err)
	}
	h.banner(list, "pack", map[string]any{"action": map[string]any{"type": "product", "value": "bt_pack"}})
	if keys := bannerKeys(h.mustRPC(u, "publication_banners", map[string]any{"placement": "target_list"})); !sameKeys(keys, "pack") {
		t.Fatalf("visible product = %v", keys)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_product SET audience_tag_ids = ARRAY[$1::text] WHERE key='bt_pack'`, vip); err != nil {
		t.Fatal(err)
	}
	if keys := bannerKeys(h.mustRPC(u, "publication_banners", map[string]any{"placement": "target_list"})); len(keys) != 0 {
		t.Fatalf("product only for VIP still shown: %v", keys)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, u, vip); err != nil {
		t.Fatal(err)
	}
	if keys := bannerKeys(h.mustRPC(u, "publication_banners", map[string]any{"placement": "target_list"})); !sameKeys(keys, "pack") {
		t.Fatalf("VIP player = %v", keys)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_product SET is_active=false WHERE key='bt_pack'`); err != nil {
		t.Fatal(err)
	}
	if keys := bannerKeys(h.mustRPC(u, "publication_banners", map[string]any{"placement": "target_list"})); len(keys) != 0 {
		t.Fatalf("inactive product still shown: %v", keys)
	}
}
