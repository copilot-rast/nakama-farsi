package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama-common/runtime"
)

type fakeNakama struct {
	runtime.NakamaModule
	db *sql.DB
}

func (n *fakeNakama) AccountUpdateId(ctx context.Context, userID, _ string, _ map[string]any, _, _, _, langTag, _ string) error {
	_, err := n.db.ExecContext(ctx, `UPDATE users SET lang_tag=$2 WHERE id=$1`, userID, langTag)
	return err
}

type testLogger struct{}

func (testLogger) Debug(string, ...any)                       {}
func (testLogger) Info(string, ...any)                        {}
func (testLogger) Warn(string, ...any)                        {}
func (testLogger) Error(string, ...any)                       {}
func (l testLogger) WithField(string, any) runtime.Logger     { return l }
func (l testLogger) WithFields(map[string]any) runtime.Logger { return l }
func (testLogger) Fields() map[string]any                     { return nil }

type fakeInit struct {
	runtime.Initializer
	console *mux.Router
}

func (f *fakeInit) RegisterConsoleHttp(path string, h func(http.ResponseWriter, *http.Request), methods ...string) error {
	f.console.HandleFunc(path, h).Methods(methods...)
	return nil
}

type harness struct {
	t    *testing.T
	s    *service
	init *fakeInit
	rpcs map[string]clientRPC
	ctx  context.Context
}

func newHarness(t *testing.T) *harness {
	db := openDisposableTestDB(t)
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	for _, q := range []string{
		`TRUNCATE base_category, base_tag CASCADE`,
		`DELETE FROM base_language WHERE code NOT IN ('fa','en')`,
		`UPDATE base_language SET is_default = false WHERE code = 'en'`,
		`UPDATE base_language SET is_default = (code = 'fa'), is_active = true, name = CASE code WHEN 'fa' THEN '{"fa":"فارسی","en":"Persian"}'::jsonb ELSE '{"fa":"انگلیسی","en":"English"}'::jsonb END`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	s := &service{db: db, nk: &fakeNakama{db: db}, logger: testLogger{}}
	init := &fakeInit{console: mux.NewRouter()}
	if err := s.registerConsoleHTTP(init); err != nil {
		t.Fatal(err)
	}
	return &harness{t: t, s: s, init: init, rpcs: s.clientRPCs(), ctx: ctx}
}

func (h *harness) user(lang string) string {
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO users (id, username, lang_tag) VALUES ($1,$2,$3)`, id, "b"+strings.ReplaceAll(id[:13], "-", ""), lang); err != nil {
		h.t.Fatal(err)
	}
	h.t.Cleanup(func() { _, _ = h.s.db.Exec(`DELETE FROM users WHERE id=$1`, id) })
	return id
}

func (h *harness) api(method, path string, body any) (int, map[string]any) {
	h.t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "/v2/console/base"+path, reader)
	rec := httptest.NewRecorder()
	h.init.console.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func (h *harness) mustAPI(want int, method, path string, body any) map[string]any {
	h.t.Helper()
	code, out := h.api(method, path, body)
	if code != want {
		h.t.Fatalf("%s %s = %d %v, want %d", method, path, code, out, want)
	}
	return out
}

func (h *harness) expectError(want int, code, method, path string, body any) {
	h.t.Helper()
	got, out := h.api(method, path, body)
	if got != want || out["error"] != code {
		h.t.Fatalf("%s %s = %d %v, want %d %s", method, path, got, out, want, code)
	}
}

func (h *harness) rpc(user, name string, payload any) (map[string]any, error) {
	h.t.Helper()
	raw, _ := json.Marshal(payload)
	out, err := h.rpcs[name](h.ctx, user, string(raw))
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(out)
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m, nil
}

func (h *harness) mustRPC(user, name string, payload any) map[string]any {
	h.t.Helper()
	out, err := h.rpc(user, name, payload)
	if err != nil {
		h.t.Fatalf("%s: %v", name, err)
	}
	return out
}

func (h *harness) rpcCode(user, name string, payload any) string {
	h.t.Helper()
	_, err := h.rpc(user, name, payload)
	if err == nil {
		h.t.Fatalf("%s should fail", name)
	}
	if app, ok := asApp(err); ok {
		return app.Code
	}
	return err.Error()
}

func items(out map[string]any) []map[string]any {
	list, _ := out["items"].([]any)
	res := make([]map[string]any, 0, len(list))
	for _, it := range list {
		res = append(res, it.(map[string]any))
	}
	return res
}

func TestLanguagesConsole(t *testing.T) {
	h := newHarness(t)
	list := items(h.mustAPI(200, http.MethodGet, "/languages", nil))
	if len(list) != 2 || list[0]["code"] != "fa" || list[0]["isDefault"] != true || list[1]["code"] != "en" || list[0]["id"] != "fa" {
		t.Fatalf("seed = %v", list)
	}
	h.expectError(400, "INVALID_LANGUAGE_CODE", http.MethodPost, "/languages", map[string]any{"code": "arabic", "nativeName": "x", "name": map[string]string{"fa": "x"}})
	h.expectError(400, "DEFAULT_LANGUAGE_REQUIRED", http.MethodPost, "/languages", map[string]any{"code": "ar", "nativeName": "العربية", "name": map[string]string{"en": "Arabic"}})
	h.expectError(400, "INVALID_DIRECTION", http.MethodPost, "/languages", map[string]any{"code": "ar", "nativeName": "العربية", "direction": "up", "name": map[string]string{"fa": "عربی"}})
	ar := h.mustAPI(201, http.MethodPost, "/languages", map[string]any{"code": "AR", "nativeName": "العربية", "direction": "rtl", "sort": 3,
		"name": map[string]string{"fa": "عربی", "en": "Arabic", "ar": "العربية"}})
	if ar["code"] != "ar" || ar["isActive"] != true || ar["isDefault"] != false || ar["direction"] != "rtl" || ar["name"].(map[string]any)["ar"] != "العربية" {
		t.Fatalf("ar = %v", ar)
	}
	h.expectError(409, "ALREADY_EXISTS", http.MethodPost, "/languages", map[string]any{"code": "ar", "nativeName": "x", "name": map[string]string{"fa": "x"}})
	h.expectError(409, "DEFAULT_LANGUAGE_LOCKED", http.MethodPut, "/languages/fa", map[string]any{"nativeName": "فارسی", "direction": "rtl", "isActive": false, "isDefault": true, "name": map[string]string{"fa": "فارسی"}})
	h.expectError(400, "DEFAULT_LANGUAGE_REQUIRED", http.MethodPut, "/languages/fa", map[string]any{"nativeName": "فارسی", "direction": "rtl", "isDefault": false, "name": map[string]string{"fa": "فارسی"}})
	h.expectError(409, "DEFAULT_LANGUAGE_LOCKED", http.MethodDelete, "/languages/fa", nil)
	h.expectError(409, "LANGUAGE_ACTIVE", http.MethodDelete, "/languages/ar", nil)
	h.expectError(404, "LANGUAGE_NOT_FOUND", http.MethodGet, "/languages/zz", nil)
	h.expectError(400, "DEFAULT_MUST_BE_ACTIVE", http.MethodPut, "/languages/ar", map[string]any{"nativeName": "العربية", "direction": "rtl", "isActive": false, "isDefault": true, "name": map[string]string{"fa": "عربی"}})

	en := h.mustAPI(200, http.MethodPut, "/languages/en", map[string]any{"nativeName": "English", "direction": "ltr", "isDefault": true, "name": map[string]string{"fa": "انگلیسی", "en": "English"}})
	if en["isDefault"] != true {
		t.Fatalf("en default = %v", en)
	}
	if fa := h.mustAPI(200, http.MethodGet, "/languages/fa", nil); fa["isDefault"] != false || fa["isActive"] != true {
		t.Fatalf("fa after switch = %v", fa)
	}
	h.expectError(400, "DEFAULT_LANGUAGE_REQUIRED", http.MethodPost, "/categories", map[string]any{"keyType": "product", "key": "c", "name": map[string]string{"fa": "دسته"}})
	h.mustAPI(200, http.MethodPut, "/languages/fa", map[string]any{"nativeName": "فارسی", "direction": "rtl", "isDefault": true, "sort": 1, "name": map[string]string{"fa": "فارسی", "en": "Persian"}})

	h.mustAPI(200, http.MethodPut, "/languages/ar", map[string]any{"nativeName": "العربية", "direction": "rtl", "isActive": false, "name": map[string]string{"fa": "عربی", "en": "Arabic"}})
	if l := h.mustAPI(200, http.MethodGet, "/languages?isActive=true", nil); l["total"].(float64) != 2 {
		t.Fatalf("active filter = %v", l)
	}
	if l := h.mustAPI(200, http.MethodGet, "/languages?q=Arabic", nil); l["total"].(float64) != 1 {
		t.Fatalf("search = %v", l)
	}
	h.mustAPI(200, http.MethodDelete, "/languages/ar", nil)
	h.expectError(404, "LANGUAGE_NOT_FOUND", http.MethodDelete, "/languages/ar", nil)
}

func TestCategoriesAndTagsConsole(t *testing.T) {
	h := newHarness(t)
	types := items(h.mustAPI(200, http.MethodGet, "/categories/key-types", nil))
	if len(types) != 6 || types[0]["key"] != "product" || types[0]["name"].(map[string]any)["fa"] == "" {
		t.Fatalf("key types = %v", types)
	}
	if tt := items(h.mustAPI(200, http.MethodGet, "/tags/key-types", nil)); len(tt) != 5 || tt[4]["key"] != "general" {
		t.Fatalf("tag key types = %v", tt)
	}
	h.expectError(400, "INVALID_KEY_TYPE", http.MethodPost, "/categories", map[string]any{"keyType": "general", "key": "c", "name": map[string]string{"fa": "x"}})
	h.expectError(400, "INVALID_LANGUAGE", http.MethodPost, "/categories", map[string]any{"keyType": "product", "key": "c", "name": map[string]string{"fa": "x", "de": "y"}})
	h.expectError(400, "DEFAULT_LANGUAGE_REQUIRED", http.MethodPost, "/categories", map[string]any{"keyType": "product", "key": "c", "name": map[string]string{"en": "x"}})
	h.expectError(400, "INVALID_KEY", http.MethodPost, "/categories", map[string]any{"keyType": "product", "key": "Bad Key", "name": map[string]string{"fa": "x"}})
	h.expectError(400, "INVALID_URL", http.MethodPost, "/categories", map[string]any{"keyType": "product", "key": "c", "iconUrl": "javascript:alert(1)", "name": map[string]string{"fa": "x"}})
	cat := h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"keyType": "product", "key": "packs", "name": map[string]string{"fa": "بسته‌ها", "en": "Packs"},
		"description": map[string]string{"fa": "همه بسته‌ها"}, "iconUrl": "/v2/economy/files/a.png", "metadata": map[string]any{"color": "red"}, "sort": 2})
	if cat["keyType"] != "product" || cat["description"].(map[string]any)["fa"] != "همه بسته‌ها" || cat["metadata"].(map[string]any)["color"] != "red" {
		t.Fatalf("category = %v", cat)
	}
	h.expectError(409, "ALREADY_EXISTS", http.MethodPost, "/categories", map[string]any{"keyType": "product", "key": "packs", "name": map[string]string{"fa": "x"}})
	newsCat := h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"keyType": "news", "key": "packs", "name": map[string]string{"fa": "رویداد", "en": "Events"}})
	if l := h.mustAPI(200, http.MethodGet, "/categories?keyType=news", nil); l["total"].(float64) != 1 {
		t.Fatalf("keyType filter = %v", l)
	}
	if l := h.mustAPI(200, http.MethodGet, "/categories?q=Events", nil); l["total"].(float64) != 1 || items(l)[0]["id"] != newsCat["id"] {
		t.Fatalf("translated search = %v", l)
	}
	h.expectError(400, "INVALID_KEY_TYPE", http.MethodGet, "/categories?keyType=zzz", nil)
	updated := h.mustAPI(200, http.MethodPut, "/categories/"+cat["id"].(string), map[string]any{"keyType": "product", "key": "packs", "name": map[string]string{"fa": "بسته", "en": "Packs"}, "sort": 5})
	if updated["sort"].(float64) != 5 || len(updated["description"].(map[string]any)) != 0 {
		t.Fatalf("update = %v", updated)
	}
	h.mustAPI(404, http.MethodGet, "/categories/"+uuid.Must(uuid.NewV4()).String(), nil)

	tag := h.mustAPI(201, http.MethodPost, "/tags", map[string]any{"keyType": "general", "key": "hot", "name": map[string]string{"fa": "داغ", "en": "Hot"}})
	h.mustAPI(201, http.MethodPost, "/tags", map[string]any{"keyType": "news", "key": "hot", "name": map[string]string{"fa": "داغ"}})
	h.expectError(400, "INVALID_KEY_TYPE", http.MethodPost, "/tags", map[string]any{"keyType": "gift", "key": "g", "name": map[string]string{"fa": "x"}})

	prizeID := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.Exec(`INSERT INTO economy_prize (id, key, name, items, tag_ids) VALUES ($1,$2,'{"fa":"x"}','[]',ARRAY[$3])`, prizeID, "base-test-"+prizeID[:8], tag["id"]); err != nil {
		t.Fatal(err)
	}
	productID := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.Exec(`INSERT INTO economy_product (id, key, name, prize_id, category_id, tag_ids) VALUES ($1,$2,'{"fa":"x"}',$3,$4,ARRAY[$5])`, productID, "base-test-"+productID[:8], prizeID, cat["id"], tag["id"]); err != nil {
		t.Fatal(err)
	}
	h.expectError(409, "IN_USE", http.MethodDelete, "/categories/"+cat["id"].(string), nil)
	h.expectError(400, "IN_USE", http.MethodPut, "/categories/"+cat["id"].(string), map[string]any{"keyType": "prize", "key": "packs", "name": map[string]string{"fa": "x"}})
	h.mustAPI(200, http.MethodDelete, "/tags/"+tag["id"].(string), nil)
	var left int
	_ = h.s.db.QueryRow(`SELECT cardinality(tag_ids) FROM economy_product WHERE id=$1`, productID).Scan(&left)
	if left != 0 {
		t.Fatalf("tag not stripped from product: %d", left)
	}
	_, _ = h.s.db.Exec(`DELETE FROM economy_product WHERE id=$1`, productID)
	_, _ = h.s.db.Exec(`DELETE FROM economy_prize WHERE id=$1`, prizeID)
	h.mustAPI(200, http.MethodDelete, "/categories/"+cat["id"].(string), nil)
	h.mustAPI(404, http.MethodDelete, "/categories/"+cat["id"].(string), nil)
}

func TestClientRPCsResolveLanguage(t *testing.T) {
	h := newHarness(t)
	h.mustAPI(201, http.MethodPost, "/languages", map[string]any{"code": "ar", "nativeName": "العربية", "direction": "rtl", "sort": 3, "name": map[string]string{"fa": "عربی", "en": "Arabic", "ar": "العربية"}})
	h.mustAPI(201, http.MethodPost, "/languages", map[string]any{"code": "tr", "nativeName": "Türkçe", "isActive": false, "name": map[string]string{"fa": "ترکی"}})
	h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"keyType": "faq", "key": "account", "sort": 1, "name": map[string]string{"fa": "حساب", "en": "Account", "ar": "الحساب"}, "description": map[string]string{"fa": "درباره حساب"}})
	h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"keyType": "faq", "key": "pay", "sort": 2, "name": map[string]string{"fa": "پرداخت"}})
	h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"keyType": "news", "key": "events", "name": map[string]string{"fa": "رویداد"}})
	h.mustAPI(201, http.MethodPost, "/tags", map[string]any{"keyType": "general", "key": "hot", "name": map[string]string{"fa": "داغ", "en": "Hot"}})
	h.mustAPI(201, http.MethodPost, "/tags", map[string]any{"keyType": "news", "key": "live", "name": map[string]string{"fa": "زنده", "ar": "مباشر"}})
	h.mustAPI(201, http.MethodPost, "/tags", map[string]any{"keyType": "product", "key": "sale", "name": map[string]string{"fa": "حراج"}})

	enUser, faUser, odd := h.user("en"), h.user("fa"), h.user("zz")
	langs := h.mustRPC(enUser, "base_languages", map[string]any{})
	if langs["default"] != "fa" || langs["lang"] != "en" || len(items(langs)) != 3 {
		t.Fatalf("languages = %v", langs)
	}
	first := items(langs)[0]
	if first["code"] != "fa" || first["name"] != "Persian" || first["nativeName"] != "فارسی" || first["direction"] != "rtl" || first["isDefault"] != true {
		t.Fatalf("fa item = %v", first)
	}
	if items(h.mustRPC(enUser, "base_languages", map[string]any{"lang": "ar"}))[2]["name"] != "العربية" {
		t.Fatal("ar names")
	}
	if l := h.mustRPC(odd, "base_languages", map[string]any{}); l["lang"] != "fa" {
		t.Fatalf("unknown lang_tag falls back to default: %v", l)
	}

	cases := []struct {
		user, lang, want, wantLang string
	}{
		{faUser, "ar", "الحساب", "ar"},
		{faUser, "en", "Account", "en"},
		{enUser, "", "Account", "en"},
		{faUser, "", "حساب", "fa"},
		{faUser, "tr", "حساب", "fa"},
		{enUser, "xx", "Account", "en"},
		{odd, "", "حساب", "fa"},
		{faUser, "AR-sa", "الحساب", "ar"},
	}
	for _, c := range cases {
		out := h.mustRPC(c.user, "base_categories", map[string]any{"keyType": "faq", "lang": c.lang})
		list := items(out)
		if out["lang"] != c.wantLang || len(list) != 2 || list[0]["name"] != c.want || list[0]["key"] != "account" {
			t.Errorf("categories lang=%q = %v, want %s/%s", c.lang, out, c.want, c.wantLang)
		}
	}
	ar := items(h.mustRPC(faUser, "base_categories", map[string]any{"keyType": "faq", "lang": "ar"}))
	if ar[1]["name"] != "پرداخت" || ar[0]["description"] != "درباره حساب" {
		t.Fatalf("fallback to default = %v", ar)
	}
	if code := h.rpcCode(faUser, "base_categories", map[string]any{}); code != "INVALID_KEY_TYPE" {
		t.Fatal(code)
	}
	if tags := items(h.mustRPC(faUser, "base_tags", map[string]any{"keyType": "news", "lang": "ar"})); len(tags) != 2 || tags[0]["name"] != "داغ" || tags[1]["name"] != "مباشر" {
		t.Fatalf("news tags = %v", tags)
	}
	if tags := items(h.mustRPC(faUser, "base_tags", map[string]any{})); len(tags) != 3 {
		t.Fatalf("all tags = %v", tags)
	}
	if code := h.rpcCode(faUser, "base_tags", map[string]any{"keyType": "gift"}); code != "INVALID_KEY_TYPE" {
		t.Fatal(code)
	}

	if out := h.mustRPC(faUser, "base_set_language", map[string]any{"lang": "AR"}); out["lang"] != "ar" {
		t.Fatalf("set = %v", out)
	}
	var tag string
	_ = h.s.db.QueryRow(`SELECT lang_tag FROM users WHERE id=$1`, faUser).Scan(&tag)
	if tag != "ar" {
		t.Fatalf("lang_tag = %q", tag)
	}
	if items(h.mustRPC(faUser, "base_categories", map[string]any{"keyType": "faq"}))[0]["name"] != "الحساب" {
		t.Fatal("lang_tag not used")
	}
	for _, bad := range []string{"tr", "zz", ""} {
		if code := h.rpcCode(faUser, "base_set_language", map[string]any{"lang": bad}); code != "LANGUAGE_NOT_AVAILABLE" {
			t.Fatalf("set %q = %s", bad, code)
		}
	}
}

func TestLanguageCoverage(t *testing.T) {
	h := newHarness(t)
	h.mustAPI(201, http.MethodPost, "/languages", map[string]any{"code": "ar", "nativeName": "العربية", "direction": "rtl", "name": map[string]string{"fa": "عربی"}})
	h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"keyType": "faq", "key": "a", "name": map[string]string{"fa": "الف", "en": "A"}})
	h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"keyType": "faq", "key": "b", "name": map[string]string{"fa": "ب", "en": "B", "ar": "ب"}, "description": map[string]string{"fa": "توضیح"}})
	out := h.mustAPI(200, http.MethodGet, "/languages/coverage", nil)
	if out["default"] != "fa" || len(out["languages"].([]any)) != 3 {
		t.Fatalf("coverage head = %v", out)
	}
	var cats map[string]any
	for _, it := range items(out) {
		if it["module"] == "base" && it["resource"] == "categories" {
			cats = it
		}
	}
	missing := cats["missing"].(map[string]any)
	if cats["total"].(float64) != 2 || missing["fa"].(float64) != 0 || missing["en"].(float64) != 1 || missing["ar"].(float64) != 2 {
		t.Fatalf("category coverage = %v", cats)
	}
	if len(items(out)) < 10 {
		t.Fatalf("coverage resources = %d", len(items(out)))
	}
}
