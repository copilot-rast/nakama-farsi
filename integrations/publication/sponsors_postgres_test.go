package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func (h *harness) sponsorReset() {
	h.t.Helper()
	h.bannerReset()
	for _, q := range []string{
		`TRUNCATE publication_sponsor_seen, publication_sponsor_day, publication_sponsor CASCADE`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'spn:%'`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *harness) sponsor(key string, extra map[string]any) string {
	h.t.Helper()
	body := map[string]any{"key": key, "name": map[string]string{"fa": "اسپانسر " + key, "en": "Sponsor " + key}, "websiteUrl": "https://example.com/" + key}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(http.StatusCreated, http.MethodPost, "/sponsors", body)["id"].(string)
}

func (h *harness) raw(path string) (int, string, http.Header) {
	h.t.Helper()
	req := httptest.NewRequest(http.MethodGet, consolePrefix+path, nil)
	rec := httptest.NewRecorder()
	h.init.console.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String(), rec.Header()
}

func TestSponsorConsoleValidation(t *testing.T) {
	h := newHarness(t)
	h.sponsorReset()
	cases := []struct {
		body  map[string]any
		code  string
		field string
	}{
		{map[string]any{"key": "Bad", "name": map[string]string{"fa": "x"}}, "INVALID_KEY", "key"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"en": "x"}}, "DEFAULT_LANGUAGE_REQUIRED", "name"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"fa": "x"}, "color": "red"}, "INVALID_COLOR", "color"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"fa": "x"}, "websiteUrl": "/local"}, "INVALID_URL", "websiteUrl"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"fa": "x"}, "links": []map[string]any{{"type": "fax", "url": "https://a.b"}}}, "INVALID_LINK_TYPE", "links[0].type"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"fa": "x"}, "links": []map[string]any{{"type": "telegram", "url": "javascript:x"}}}, "INVALID_URL", "links[0].url"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"fa": "x"}, "contactEmail": "nope"}, "INVALID_EMAIL", "contactEmail"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"fa": "x"}, "contactPhone": "12"}, "INVALID_PHONE", "contactPhone"},
		{map[string]any{"key": "sp_a", "name": map[string]string{"fa": "x"}, "startsAt": "2030-01-02T00:00:00Z", "endsAt": "2030-01-01T00:00:00Z"}, "INVALID_WINDOW", "endsAt"},
	}
	for _, c := range cases {
		code, out := h.api(http.MethodPost, "/sponsors", c.body)
		if code != http.StatusBadRequest || out["error"] != c.code || out["field"] != c.field {
			t.Fatalf("%v = %d %v, want %s on %s", c.body, code, out, c.code, c.field)
		}
	}
	id := h.sponsor("sp_shop", map[string]any{"color": "#FF8800", "contactEmail": " Ads@Shop.IR ", "contactPhone": "۰۹۱۲۱۲۳۴۵۶۷",
		"links": []map[string]any{{"type": "instagram", "url": "https://instagram.com/shop", "label": map[string]string{"fa": "اینستاگرام"}}}})
	got := h.mustAPI(http.StatusOK, http.MethodGet, "/sponsors/"+id, nil)
	if got["color"] != "#ff8800" || got["contactEmail"] != "ads@shop.ir" || got["state"] != "live" || len(got["links"].([]any)) != 1 {
		t.Fatalf("sponsor = %v", got)
	}
	if code, out := h.api(http.MethodPut, "/sponsors/"+id, map[string]any{"key": "sp_other", "name": map[string]string{"fa": "x"}}); code != http.StatusBadRequest || out["error"] != "KEY_LOCKED" {
		t.Fatalf("key change = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/sponsors", map[string]any{"key": "sp_shop", "name": map[string]string{"fa": "x"}}); code != http.StatusConflict || out["error"] != "KEY_EXISTS" {
		t.Fatalf("dup = %d %v", code, out)
	}
	list := h.bannerList("sp_home")
	if code, out := h.api(http.MethodPost, "/banners", map[string]any{"key": "sp_b1", "listId": list, "title": map[string]string{"fa": "t"}, "sponsorId": "00000000-0000-0000-0000-000000000001"}); code != http.StatusBadRequest || out["error"] != "SPONSOR_NOT_FOUND" {
		t.Fatalf("missing sponsor = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/banners", map[string]any{"key": "sp_b1", "listId": list, "title": map[string]string{"fa": "t"}, "action": map[string]any{"type": "sponsor"}}); code != http.StatusBadRequest || out["error"] != "SPONSOR_REQUIRED" {
		t.Fatalf("sponsor action without sponsor = %d %v", code, out)
	}
	bare := h.sponsor("sp_bare", map[string]any{"websiteUrl": ""})
	if code, out := h.api(http.MethodPost, "/banners", map[string]any{"key": "sp_b1", "listId": list, "title": map[string]string{"fa": "t"}, "sponsorId": bare, "action": map[string]any{"type": "sponsor"}}); code != http.StatusBadRequest || out["error"] != "SPONSOR_NO_WEBSITE" {
		t.Fatalf("sponsor action without website = %d %v", code, out)
	}
	b := h.banner(list, "sp_b1", map[string]any{"sponsorId": id, "action": map[string]any{"type": "sponsor", "label": map[string]string{"fa": "برو"}}})
	bv := h.mustAPI(http.StatusOK, http.MethodGet, "/banners/"+b, nil)
	if bv["sponsorId"] != id || bv["sponsorKey"] != "sp_shop" {
		t.Fatalf("banner sponsor = %v", bv)
	}
	if got := items(h.mustAPI(http.StatusOK, http.MethodGet, "/banners?sponsorId="+id, nil)); len(got) != 1 {
		t.Fatalf("filter by sponsor = %v", got)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/banner-lists/"+list, map[string]any{"key": "sp_home", "name": map[string]string{"fa": "خانه"}, "sponsorId": bare})
	if code, out := h.api(http.MethodDelete, "/sponsors/"+id, nil); code != http.StatusConflict || out["error"] != "IN_USE" {
		t.Fatalf("delete used = %d %v", code, out)
	}
	if code, out := h.api(http.MethodDelete, "/sponsors/"+bare, nil); code != http.StatusConflict || out["error"] != "IN_USE" {
		t.Fatalf("delete placement sponsor = %d %v", code, out)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/banner-lists/"+list, map[string]any{"key": "sp_home", "name": map[string]string{"fa": "خانه"}})
	h.mustAPI(http.StatusOK, http.MethodDelete, "/sponsors/"+bare, nil)
}

func TestSponsorCountsAreIdempotentAndFollowTheWindow(t *testing.T) {
	h := newHarness(t)
	h.sponsorReset()
	shop := h.sponsor("sp_shop", map[string]any{"links": []map[string]any{{"type": "telegram", "url": "https://t.me/shop", "label": map[string]string{"fa": "تلگرام", "en": "Telegram"}}}})
	host := h.sponsor("sp_host", nil)
	later := h.sponsor("sp_later", map[string]any{"startsAt": time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)})
	list := h.mustAPI(http.StatusCreated, http.MethodPost, "/banner-lists", map[string]any{"key": "sp_home", "name": map[string]string{"fa": "خانه"}, "sponsorId": host})["id"].(string)
	b := h.banner(list, "sp_b1", map[string]any{"sponsorId": shop, "action": map[string]any{"type": "sponsor"}})
	plain := h.banner(list, "sp_b2", nil)
	hidden := h.banner(list, "sp_b3", map[string]any{"sponsorId": later})
	u, other := h.user(nil, "", "fa"), h.user(nil, "", "en")

	out := h.mustRPC(u, "publication_banners", map[string]any{"placement": "sp_home"})
	if keys := bannerKeys(out); !sameKeys(keys, "sp_b1", "sp_b2") {
		t.Fatalf("keys = %v", keys)
	}
	first := items(out)[0].(map[string]any)
	act := first["action"].(map[string]any)
	if first["sponsor"].(map[string]any)["key"] != "sp_shop" || act["type"] != "sponsor" || act["url"] != "https://example.com/sp_shop" {
		t.Fatalf("sponsored banner = %v", first)
	}
	if out["placement"].(map[string]any)["sponsor"].(map[string]any)["key"] != "sp_host" {
		t.Fatalf("placement sponsor = %v", out["placement"])
	}

	nodes := []*service{testService(), testService()}
	for _, n := range nodes {
		n.db, n.nk = h.s.db, h.nk
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			n := nodes[i%2]
			call := func(name string, payload map[string]any) {
				fn := n.clientRPCs()[name]
				if _, err := fn(h.ctx, u, mustJSON(payload)); err != nil {
					t.Error(err)
				}
			}
			call("publication_banner_view", map[string]any{"bannerIds": []string{b, plain, hidden}})
			call("publication_sponsor_view", map[string]any{"sponsorIds": []string{shop, host, later}})
			call("publication_banner_click", map[string]any{"bannerId": b})
			call("publication_sponsor_click", map[string]any{"sponsorId": shop})
		}(i)
	}
	wg.Wait()
	day := func(id string) (int, int) {
		var v, c int
		if err := h.s.db.QueryRowContext(h.ctx, `SELECT COALESCE(sum(views),0), COALESCE(sum(clicks),0) FROM publication_sponsor_day WHERE sponsor_id=$1`, id).Scan(&v, &c); err != nil {
			t.Fatal(err)
		}
		return v, c
	}
	if v, c := day(shop); v != 1 || c != 1 {
		t.Fatalf("shop = %d views %d clicks", v, c)
	}
	if v, c := day(host); v != 1 || c != 0 {
		t.Fatalf("host = %d views %d clicks", v, c)
	}
	if v, c := day(later); v != 0 || c != 0 {
		t.Fatalf("later = %d views %d clicks", v, c)
	}
	sv := h.mustRPC(other, "publication_sponsor_view", map[string]any{"sponsorIds": []string{shop, later}})
	if len(sv["counted"].([]any)) != 1 || len(sv["skipped"].([]any)) != 1 {
		t.Fatalf("other view = %v", sv)
	}
	click := h.mustRPC(other, "publication_sponsor_click", map[string]any{"sponsorId": shop, "link": "telegram", "lang": "en"})
	if click["counted"] != true || click["url"] != "https://t.me/shop" {
		t.Fatalf("link click = %v", click)
	}
	if code := h.rpcCode(other, "publication_sponsor_click", map[string]any{"sponsorId": shop, "link": "youtube"}); code != "SPONSOR_LINK_NOT_FOUND" {
		t.Fatalf("missing link = %s", code)
	}
	if code := h.rpcCode(other, "publication_sponsor_click", map[string]any{"sponsorId": later}); code != "SPONSOR_NOT_AVAILABLE" {
		t.Fatalf("later click = %s", code)
	}
	list2 := items(h.mustRPC(other, "publication_sponsors", map[string]any{"lang": "en"}))
	if len(list2) != 2 {
		t.Fatalf("live sponsors = %v", list2)
	}
	sp := list2[1].(map[string]any)
	if sp["key"] != "sp_shop" || sp["name"] != "Sponsor sp_shop" || sp["links"].([]any)[0].(map[string]any)["label"] != "Telegram" {
		t.Fatalf("sponsor view = %v", sp)
	}
	if _, has := sp["contactEmail"]; has {
		t.Fatalf("contact leaked: %v", sp)
	}

	h.mustAPI(http.StatusOK, http.MethodPut, "/sponsors/"+shop, map[string]any{"key": "sp_shop", "name": map[string]string{"fa": "x"}, "websiteUrl": "https://example.com/sp_shop", "endsAt": time.Now().Add(-time.Minute).UTC().Format(time.RFC3339), "startsAt": time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)})
	if keys := bannerKeys(h.mustRPC(u, "publication_banners", map[string]any{"placement": "sp_home"})); !sameKeys(keys, "sp_b2") {
		t.Fatalf("after sponsor ended = %v", keys)
	}
	if code := h.rpcCode(u, "publication_banner_click", map[string]any{"bannerId": b}); code != "BANNER_NOT_AVAILABLE" {
		t.Fatalf("ended sponsor click = %s", code)
	}
	listID := h.mustAPI(http.StatusOK, http.MethodGet, "/banner-lists/"+list, nil)
	prev := h.mustAPI(http.StatusOK, http.MethodGet, "/banner-lists/"+list+"/preview?userId="+u, nil)
	reasons := map[string]string{}
	for _, x := range prev["hidden"].([]any) {
		m := x.(map[string]any)
		reasons[m["key"].(string)] = m["reason"].(string)
	}
	if reasons["sp_b1"] != "sponsor_off" || reasons["sp_b3"] != "sponsor_off" || listID["sponsorId"] != host {
		t.Fatalf("preview reasons = %v", reasons)
	}

	rep := h.mustAPI(http.StatusOK, http.MethodGet, "/sponsors/"+shop+"/report?days=7", nil)
	if num(rep["views"]) != 2 || num(rep["clicks"]) != 2 || num(rep["players"]) != 2 || len(rep["days"].([]any)) != 7 {
		t.Fatalf("report = %v", rep)
	}
	banners := rep["banners"].([]any)
	if len(banners) != 1 || num(banners[0].(map[string]any)["views"]) != 1 || num(banners[0].(map[string]any)["clicks"]) != 1 {
		t.Fatalf("report banners = %v", banners)
	}
	if code, out := h.api(http.MethodGet, "/sponsors/"+shop+"/report?from=2030-01-02&to=2030-01-01", nil); code != http.StatusBadRequest || out["error"] != "INVALID_DATE_RANGE" {
		t.Fatalf("bad range = %d %v", code, out)
	}
	code, body, hdr := h.raw("/sponsors/" + shop + "/report.csv?days=3")
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if code != http.StatusOK || !strings.HasPrefix(hdr.Get("Content-Type"), "text/csv") || !strings.Contains(hdr.Get("Content-Disposition"), "sponsor-sp_shop-") ||
		!strings.HasPrefix(lines[0], "\xef\xbb\xbfday,views,clicks,players,ctr_percent") || !strings.HasPrefix(lines[4], "total,2,2,2,100.0") || !strings.Contains(body, "sp_b1,1,1,sp_home,100.0") {
		t.Fatalf("csv = %d %q", code, body)
	}

	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_sponsor_seen SET day = day - 60`); err != nil {
		t.Fatal(err)
	}
	if n, err := h.s.sweepSponsorSeen(h.ctx); err != nil || n == 0 {
		t.Fatalf("sweep = %d %v", n, err)
	}
}

func TestSponsorRateLimit(t *testing.T) {
	h := newHarness(t)
	h.sponsorReset()
	u := h.user(nil, "", "fa")
	for i := 0; i < sponsorListPerMinute; i++ {
		h.mustRPC(u, "publication_sponsors", map[string]any{})
	}
	if code := h.rpcCode(u, "publication_sponsors", map[string]any{}); code != "TOO_MANY_REQUESTS" {
		t.Fatalf("limit = %s", code)
	}
}

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}
