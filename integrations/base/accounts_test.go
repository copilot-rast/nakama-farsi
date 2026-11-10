package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"1.2.10", "1.2.9", 1},
		{"1.2", "1.2.0", 0},
		{"v2.0.0", "2.0.0", 0},
		{"2.0.0-beta.1", "2.0.0", -1},
		{"2.0.0+45", "2.0.0", 0},
		{"2.0.0-beta.2", "2.0.0-beta.1", 1},
		{"10", "9.9.9", 1},
		{"1.0.0.1", "1.0.0", 1},
	} {
		got, ok := compareVersions(tc.a, tc.b)
		if !ok || got != tc.want {
			t.Errorf("compare(%s, %s) = %d, %v; want %d", tc.a, tc.b, got, ok, tc.want)
		}
	}
	for _, bad := range []string{"", "a.b", "1..2", "1.2.3.4.5", "-1", "1.2 beta"} {
		if _, ok := parseVersion(bad); ok {
			t.Errorf("parseVersion(%q) should fail", bad)
		}
	}
}

func TestVersionStatus(t *testing.T) {
	for _, tc := range []struct{ cur, latest, min, want string }{
		{"1.0.0", "1.2.0", "1.1.0", versionRequired},
		{"1.1.0", "1.2.0", "1.1.0", versionAvailable},
		{"1.2.0", "1.2.0", "1.1.0", versionOK},
		{"1.3.0", "1.2.0", "", versionOK},
		{"1.0.0", "1.2.0", "", versionAvailable},
	} {
		if got := versionStatus(tc.cur, tc.latest, tc.min); got != tc.want {
			t.Errorf("status(%s, %s, %s) = %s, want %s", tc.cur, tc.latest, tc.min, got, tc.want)
		}
	}
}

func TestAccountRules(t *testing.T) {
	r := defaultRules()
	r.UsernameMin, r.UsernameMax, r.UsernameCharset = 4, 10, "latin"
	r.PasswordMin, r.PasswordLetter, r.PasswordDigit = 10, true, true
	r.DisplayNameMax = 5
	for name, want := range map[string]string{"abc": "USERNAME_TOO_SHORT", "abcdefghijk": "USERNAME_TOO_LONG", "علی‌رضا": "USERNAME_INVALID_CHARACTERS", "ali_99": ""} {
		if got := r.usernameError(name); got != want {
			t.Errorf("username %q = %q, want %q", name, got, want)
		}
	}
	r.UsernameCharset = "any"
	if got := r.usernameError("علی‌رضا"); got != "" {
		t.Errorf("Persian names are allowed with any: %q", got)
	}
	for pw, want := range map[string]string{"short1": "PASSWORD_TOO_SHORT", "0123456789": "PASSWORD_NEEDS_LETTER", "abcdefghij": "PASSWORD_NEEDS_DIGIT", "abcdefghi1": "", "رمزعبورمن۱۲": ""} {
		if got := r.passwordError(pw); got != want {
			t.Errorf("password %q = %q, want %q", pw, got, want)
		}
	}
	if r.displayNameError("سلام دنیا") != "DISPLAY_NAME_TOO_LONG" || r.displayNameError("سلام") != "" {
		t.Error("display name length counts letters")
	}
	bad := defaultRules()
	bad.UsernameMin, bad.UsernameMax = 10, 5
	if _, err := bad.clean(); err == nil {
		t.Error("min above max must fail")
	}
	bad = defaultRules()
	bad.PasswordMin = 6
	if _, err := bad.clean(); err == nil {
		t.Error("password below Nakama's 8 must fail")
	}
}

func TestFlagEmoji(t *testing.T) {
	if flagEmoji("IR") != "🇮🇷" || flagEmoji("de") != "" || flagEmoji("USA") != "" {
		t.Fatal("flag emoji")
	}
}

func TestAccountRulesPostgres(t *testing.T) {
	h := newHarness(t)
	t.Cleanup(func() {
		_, _ = h.s.db.Exec(`UPDATE base_account_rules SET username_min=1, username_max=128, username_charset='any', password_min=8, password_letter=false, password_digit=false, display_name_max=255`)
	})
	got := h.mustAPI(http.StatusOK, http.MethodGet, "/account-rules", nil)
	if got["usernameMin"] != float64(1) || got["usernameMax"] != float64(128) || got["passwordMin"] != float64(8) {
		t.Fatalf("defaults = %v", got)
	}
	h.expectError(http.StatusBadRequest, "INVALID_RULE", http.MethodPut, "/account-rules", map[string]any{"usernameMin": 9, "usernameMax": 4, "passwordMin": 8, "displayNameMax": 20})
	saved := h.mustAPI(http.StatusOK, http.MethodPut, "/account-rules", map[string]any{"usernameMin": 5, "usernameMax": 12, "usernameCharset": "latin", "passwordMin": 10, "passwordDigit": true, "displayNameMax": 20})
	if saved["usernameMax"] != float64(12) || saved["passwordDigit"] != true {
		t.Fatalf("saved = %v", saved)
	}
	pub := h.mustRPC(h.user("fa"), "base_account_rules", map[string]any{})
	if pub["usernameMin"] != float64(5) || pub["usernameCharset"] != "latin" || pub["updatedBy"] != nil {
		t.Fatalf("client rules = %v", pub)
	}
	ctx := context.Background()
	if _, err := h.s.beforeAuthEmail(ctx, nil, nil, nil, &api.AuthenticateEmailRequest{Account: &api.AccountEmail{Email: "new@example.com", Password: "abcdefghij"}, Username: "player1"}); err == nil || err.(*runtime.Error).Message != "PASSWORD_NEEDS_DIGIT" {
		t.Fatalf("weak password = %v", err)
	}
	if _, err := h.s.beforeAuthEmail(ctx, nil, nil, nil, &api.AuthenticateEmailRequest{Account: &api.AccountEmail{Email: "new@example.com", Password: "abcdefghi1"}, Username: "abc"}); err == nil || err.(*runtime.Error).Message != "USERNAME_TOO_SHORT" {
		t.Fatalf("short username = %v", err)
	}
	if _, err := h.s.beforeAuthEmail(ctx, nil, nil, nil, &api.AuthenticateEmailRequest{Account: &api.AccountEmail{Email: "new@example.com", Password: "abcdefghi1"}, Username: "player1"}); err != nil {
		t.Fatalf("good sign-up = %v", err)
	}
	old := h.user("fa")
	if _, err := h.s.db.Exec(`UPDATE users SET email='old@example.com', username='x' WHERE id=$1`, old); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.beforeAuthEmail(ctx, nil, nil, nil, &api.AuthenticateEmailRequest{Account: &api.AccountEmail{Email: "OLD@example.com", Password: "short123"}}); err != nil {
		t.Fatalf("existing accounts sign in under old rules: %v", err)
	}
	uctx := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, old)
	if _, err := h.s.beforeUpdateAccount(uctx, nil, nil, nil, &api.UpdateAccountRequest{Username: wrapperspb.String("x")}); err != nil {
		t.Fatalf("keeping the same username = %v", err)
	}
	if _, err := h.s.beforeUpdateAccount(uctx, nil, nil, nil, &api.UpdateAccountRequest{Username: wrapperspb.String("سارا_جان")}); err == nil || err.(*runtime.Error).Message != "USERNAME_INVALID_CHARACTERS" {
		t.Fatalf("latin only = %v", err)
	}
	if _, err := h.s.beforeUpdateAccount(uctx, nil, nil, nil, &api.UpdateAccountRequest{DisplayName: wrapperspb.String("یک نام خیلی خیلی بلند برای نمایش")}); err == nil || err.(*runtime.Error).Message != "DISPLAY_NAME_TOO_LONG" {
		t.Fatalf("display name = %v", err)
	}
	if _, err := h.s.beforeLinkEmail(uctx, nil, nil, nil, &api.AccountEmail{Email: "z@example.com", Password: "abc"}); err == nil {
		t.Fatal("linking an email checks the password")
	}
}

func TestAppVersionsPostgres(t *testing.T) {
	h := newHarness(t)
	if _, err := h.s.db.Exec(`DELETE FROM base_app_version WHERE platform IN ('ios','linux')`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = h.s.db.Exec(`DELETE FROM base_app_version WHERE platform IN ('ios','linux')`) })
	h.expectError(http.StatusBadRequest, "INVALID_PLATFORM", http.MethodPost, "/app-versions", map[string]any{"platform": "symbian", "latestVersion": "1.0"})
	h.expectError(http.StatusBadRequest, "INVALID_VERSION", http.MethodPost, "/app-versions", map[string]any{"platform": "ios", "latestVersion": "one"})
	h.expectError(http.StatusBadRequest, "MIN_ABOVE_LATEST", http.MethodPost, "/app-versions", map[string]any{"platform": "ios", "latestVersion": "1.0.0", "minVersion": "1.1"})
	h.expectError(http.StatusBadRequest, "INVALID_URL", http.MethodPost, "/app-versions", map[string]any{"platform": "ios", "latestVersion": "1.0.0", "storeLinks": []any{map[string]any{"store": "app_store", "url": "javascript:alert(1)"}}})
	h.expectError(http.StatusBadRequest, "DEFAULT_LANGUAGE_REQUIRED", http.MethodPost, "/app-versions", map[string]any{"platform": "ios", "latestVersion": "1.0.0", "updateMessage": map[string]any{"en": "Update"}})
	row := h.mustAPI(http.StatusCreated, http.MethodPost, "/app-versions", map[string]any{"platform": "ios", "latestVersion": "2.1.0", "minVersion": "2.0.0",
		"storeLinks":    []any{map[string]any{"store": "app_store", "url": "https://apps.apple.com/app/id1"}, map[string]any{"store": "direct", "url": "https://example.com/app.ipa", "label": map[string]any{"fa": "دانلود"}}},
		"updateMessage": map[string]any{"fa": "نسخه‌ی تازه آمد", "en": "A new version is out"}, "forceMessage": map[string]any{"fa": "باید به‌روز کنید", "en": "Please update"}})
	h.expectError(http.StatusConflict, "ALREADY_EXISTS", http.MethodPost, "/app-versions", map[string]any{"platform": "ios", "latestVersion": "3.0.0"})
	user := h.user("en")
	ok := h.mustRPC(user, "base_app_version", map[string]any{"platform": "ios", "version": "2.1.0"})
	if ok["status"] != versionOK || ok["message"] != "" {
		t.Fatalf("ok = %v", ok)
	}
	avail := h.mustRPC(user, "base_app_version", map[string]any{"platform": "ios", "version": "2.0.5", "store": "direct"})
	stores := avail["stores"].([]any)
	if avail["status"] != versionAvailable || avail["message"] != "A new version is out" || avail["storeUrl"] != "https://example.com/app.ipa" || len(stores) != 2 ||
		stores[0].(map[string]any)["name"] != "App Store" || stores[1].(map[string]any)["name"] != "دانلود" {
		t.Fatalf("available = %v", avail)
	}
	req := h.mustRPC(user, "base_app_version", map[string]any{"platform": "ios", "version": "1.9", "lang": "fa"})
	if req["status"] != versionRequired || req["message"] != "باید به‌روز کنید" || req["storeUrl"] != "https://apps.apple.com/app/id1" || req["minimum"] != "2.0.0" {
		t.Fatalf("required = %v", req)
	}
	none := h.mustRPC(user, "base_app_version", map[string]any{"platform": "linux", "version": "0.1"})
	if none["status"] != versionOK {
		t.Fatalf("no row = %v", none)
	}
	if code := h.rpcCode(user, "base_app_version", map[string]any{"platform": "ios", "version": "x"}); code != "INVALID_VERSION" {
		t.Fatalf("bad version = %s", code)
	}
	id := row["id"].(string)
	h.mustAPI(http.StatusOK, http.MethodPut, "/app-versions/"+id, map[string]any{"platform": "ios", "latestVersion": "2.1.0", "minVersion": "2.0.0", "isActive": false})
	if off := h.mustRPC(user, "base_app_version", map[string]any{"platform": "ios", "version": "1.0"}); off["status"] != versionOK {
		t.Fatalf("inactive row = %v", off)
	}
	list := h.mustAPI(http.StatusOK, http.MethodGet, "/app-versions?q=ios", nil)
	if len(items(list)) != 1 {
		t.Fatalf("list = %v", list)
	}
	h.mustAPI(http.StatusOK, http.MethodDelete, "/app-versions/"+id, nil)
}

func TestCountriesPostgres(t *testing.T) {
	h := newHarness(t)
	t.Cleanup(func() {
		_, _ = h.s.db.Exec(`UPDATE base_country SET is_active=true, sort=100 WHERE code IN ('DE','TR')`)
		_, _ = h.s.db.Exec(`DELETE FROM base_country WHERE code='XA'`)
	})
	list := h.mustAPI(http.StatusOK, http.MethodGet, "/countries?q=ایران", nil)
	got := items(list)
	if len(got) != 1 || got[0]["code"] != "IR" || got[0]["dialCode"] != "+98" || got[0]["flag"] != "🇮🇷" {
		t.Fatalf("search = %v", list)
	}
	if total := h.mustAPI(http.StatusOK, http.MethodGet, "/countries?limit=1", nil)["total"]; total.(float64) < 200 {
		t.Fatalf("seeded countries = %v", total)
	}
	de := h.mustAPI(http.StatusOK, http.MethodGet, "/countries/de", nil)
	h.mustAPI(http.StatusOK, http.MethodPut, "/countries/DE", map[string]any{"name": de["name"], "dialCode": "+49", "isActive": false})
	h.expectError(http.StatusBadRequest, "INVALID_DIAL_CODE", http.MethodPut, "/countries/TR", map[string]any{"name": map[string]any{"fa": "ترکیه"}, "dialCode": "+0"})
	h.mustAPI(http.StatusOK, http.MethodPut, "/countries/TR", map[string]any{"name": map[string]any{"fa": "ترکیه", "en": "Türkiye"}, "dialCode": "90", "sort": 0})
	h.expectError(http.StatusBadRequest, "INVALID_COUNTRY_CODE", http.MethodPost, "/countries", map[string]any{"code": "XYZ", "name": map[string]any{"fa": "x"}})
	h.mustAPI(http.StatusCreated, http.MethodPost, "/countries", map[string]any{"code": "xa", "name": map[string]any{"fa": "آزمایشی"}, "dialCode": "+999"})
	h.expectError(http.StatusConflict, "ALREADY_EXISTS", http.MethodPost, "/countries", map[string]any{"code": "XA", "name": map[string]any{"fa": "آزمایشی"}})
	user := h.user("en")
	out := h.mustRPC(user, "base_countries", map[string]any{})
	all := items(out)
	if all[0]["code"] != "TR" || all[0]["name"] != "Türkiye" {
		t.Fatalf("sort first = %v", all[0])
	}
	for _, c := range all {
		if c["code"] == "DE" {
			t.Fatal("inactive countries stay out of the client list")
		}
	}
	if code := h.rpcCode(user, "base_set_country", map[string]any{"code": "DE"}); code != "COUNTRY_NOT_AVAILABLE" {
		t.Fatalf("inactive pick = %s", code)
	}
	h.mustRPC(user, "base_set_country", map[string]any{"code": "ir"})
	if sel := h.mustRPC(user, "base_countries", map[string]any{"lang": "fa"}); sel["selected"] != "IR" {
		t.Fatalf("selected = %v", sel["selected"])
	}
	h.mustAPI(http.StatusOK, http.MethodDelete, "/countries/XA", nil)
}

func TestGiftAndCurrencyTaxonomyInUse(t *testing.T) {
	h := newHarness(t)
	cat := h.mustAPI(http.StatusCreated, http.MethodPost, "/categories", map[string]any{"key": "premium", "keyType": "currency", "name": map[string]any{"fa": "ویژه"}})
	tg := h.mustAPI(http.StatusCreated, http.MethodPost, "/tags", map[string]any{"key": "spring", "keyType": "gift", "name": map[string]any{"fa": "بهار"}})
	h.mustAPI(http.StatusCreated, http.MethodPost, "/tags", map[string]any{"key": "shiny", "keyType": "currency", "name": map[string]any{"fa": "براق"}})
	types := h.mustAPI(http.StatusOK, http.MethodGet, "/tags/key-types", nil)
	keys := map[string]bool{}
	for _, it := range types["items"].([]any) {
		keys[it.(map[string]any)["key"].(string)] = true
	}
	if !keys["gift"] || !keys["currency"] {
		t.Fatalf("tag key types = %v", types)
	}
	if _, err := h.s.db.Exec(`DELETE FROM economy_currency WHERE key='e2e_tax'`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = h.s.db.Exec(`DELETE FROM economy_currency WHERE key='e2e_tax'`) })
	if _, err := h.s.db.Exec(`INSERT INTO economy_currency (id, key, name, type, category_id, tag_ids) VALUES (gen_random_uuid(), 'e2e_tax', '{"fa":"x"}', 'soft', $1, ARRAY[$2::text])`, cat["id"], tg["id"]); err != nil {
		t.Fatal(err)
	}
	h.expectError(http.StatusConflict, "IN_USE", http.MethodDelete, "/categories/"+cat["id"].(string), nil)
	h.mustAPI(http.StatusOK, http.MethodDelete, "/tags/"+tg["id"].(string), nil)
	var left int
	if err := h.s.db.QueryRow(`SELECT cardinality(tag_ids) FROM economy_currency WHERE key='e2e_tax'`).Scan(&left); err != nil || left != 0 {
		t.Fatalf("deleted tag must leave tag_ids: %d %v", left, err)
	}
	user := h.user("fa")
	out := h.mustRPC(user, "base_tags", map[string]any{"keyType": "currency"})
	if len(items(out)) != 1 || items(out)[0]["key"] != "shiny" {
		t.Fatalf("currency tags for clients = %v", out)
	}
}
