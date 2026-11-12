package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const disposableTestDBEnv = "STARGAME_DISPOSABLE_TEST_DB_URL"

const testSecret = "stargame-test-secret-0123456789abcdef0123"

type fakeInit struct {
	runtime.Initializer
	router *mux.Router
}

func (f *fakeInit) RegisterHttp(path string, h func(http.ResponseWriter, *http.Request), methods ...string) error {
	f.router.HandleFunc(path, h).Methods(methods...)
	return nil
}

func testSite(t *testing.T) *mux.Router {
	t.Helper()
	site, err := newWebSite("the-key")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeInit{router: mux.NewRouter()}
	if err := site.register(f); err != nil {
		t.Fatal(err)
	}
	return f.router
}

func get(r http.Handler, method, target string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestWebServesEmbeddedGame(t *testing.T) {
	r := testSite(t)
	for target, ct := range map[string]string{
		"/stargame/":           "text/html; charset=utf-8",
		"/stargame/index.html": "text/html; charset=utf-8",
		"/stargame/app.js":     "text/javascript; charset=utf-8",
		"/stargame/style.css":  "text/css; charset=utf-8",
		"/stargame/fonts/Vazirmatn-FD-NL-Bold.woff2": "font/woff2",
	} {
		w := get(r, http.MethodGet, target, nil)
		if w.Code != 200 || w.Header().Get("Content-Type") != ct || w.Body.Len() == 0 {
			t.Fatalf("%s: %d %q %d bytes", target, w.Code, w.Header().Get("Content-Type"), w.Body.Len())
		}
		if w.Header().Get("ETag") == "" || w.Header().Get("Cache-Control") == "" {
			t.Fatalf("%s: missing cache headers %v", target, w.Header())
		}
		again := get(r, http.MethodGet, target, map[string]string{"If-None-Match": w.Header().Get("ETag")})
		if again.Code != http.StatusNotModified {
			t.Fatalf("%s: conditional get = %d", target, again.Code)
		}
	}
	if w := get(r, http.MethodGet, "/stargame/app.js", nil); w.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("app.js cache = %q", w.Header().Get("Cache-Control"))
	}
	if w := get(r, http.MethodHead, "/stargame/", nil); w.Code != 200 || w.Body.Len() != 0 {
		t.Fatalf("head = %d %d", w.Code, w.Body.Len())
	}
	if !strings.Contains(get(r, http.MethodGet, "/stargame/", nil).Body.String(), "app.js") {
		t.Fatal("index does not load app.js")
	}
}

func TestWebRedirectsAndRefuses(t *testing.T) {
	r := testSite(t)
	w := get(r, http.MethodGet, "/stargame?payment=p1&success=true", nil)
	if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != "/stargame/?payment=p1&success=true" {
		t.Fatalf("redirect = %d %q", w.Code, w.Header().Get("Location"))
	}
	if w := get(r, http.MethodGet, "/stargame", nil); w.Header().Get("Location") != "/stargame/" {
		t.Fatalf("bare redirect = %q", w.Header().Get("Location"))
	}
	for _, target := range []string{"/stargame/fonts", "/stargame/fonts/", "/stargame/missing.js", "/stargame/../main.go", "/stargame/./app.js", "/stargame//app.js", "/stargame/web/app.js"} {
		if w := get(r, http.MethodGet, target, nil); w.Code != http.StatusNotFound && w.Code != http.StatusMovedPermanently {
			t.Fatalf("%s = %d, want 404", target, w.Code)
		} else if strings.Contains(w.Body.String(), "<a href") {
			t.Fatalf("%s lists a directory", target)
		}
	}
	if w := get(r, http.MethodPost, "/stargame/", nil); w.Code == 200 {
		t.Fatal("POST must not be served")
	}
}

func TestWebConfig(t *testing.T) {
	w := get(testSite(t), http.MethodGet, "/stargame/config.json", nil)
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || w.Code != 200 || body["serverKey"] != "the-key" {
		t.Fatalf("config = %d %s %v", w.Code, w.Body.String(), err)
	}
	if w.Header().Get("Cache-Control") != "no-store" || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("config headers = %v", w.Header())
	}
}

func TestSeedEnabled(t *testing.T) {
	for v, want := range map[string]bool{"": true, "true": true, "false": false, " FALSE ": false, "0": true} {
		if got := seedEnabled(map[string]string{"STARGAME_SEED": v}); got != want {
			t.Errorf("STARGAME_SEED=%q -> %v", v, got)
		}
	}
}

func openDisposableTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv(disposableTestDBEnv)
	if dsn == "" {
		t.Skip(disposableTestDBEnv + " not set; PostgreSQL NOT_RUN")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil || !strings.HasSuffix(cfg.Database, "_disposable") || (cfg.Host != "127.0.0.1" && cfg.Host != "localhost" && !strings.HasPrefix(cfg.Host, "/")) {
		t.Fatalf("%s must name a local database ending in _disposable", disposableTestDBEnv)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	var name string
	if err := db.QueryRow(`SELECT current_database()`).Scan(&name); err != nil || !strings.HasSuffix(name, "_disposable") {
		db.Close()
		t.Fatalf("connected database %q is not disposable: %v", name, err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func resetSeedData(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, q := range []string{
		`DELETE FROM economy_gift_use`,
		`DELETE FROM economy_gift WHERE code IN ('STAR100','HALF50','STARFAN')`,
		`DELETE FROM publication_banner WHERE key LIKE 'star\_%'`,
		`DELETE FROM publication_banner_list WHERE key='menu_carousel'`,
		`DELETE FROM economy_user_achievement`,
		`DELETE FROM economy_achievement WHERE key LIKE 'star\_%'`,
		`DELETE FROM base_category WHERE key_type='achievement' AND key LIKE 'star\_%'`,
		`DELETE FROM economy_chance_play`,
		`DELETE FROM economy_chance_game WHERE key LIKE 'star\_%'`,
		`DELETE FROM economy_prize_group WHERE key LIKE 'star\_%'`,
		`DELETE FROM economy_payment_log`,
		`DELETE FROM economy_payment`,
		`DELETE FROM economy_product WHERE key LIKE 'star\_%'`,
		`DELETE FROM economy_user_item`,
		`DELETE FROM economy_item WHERE key LIKE 'star\_%'`,
		`DELETE FROM economy_reward_claim`,
		`DELETE FROM economy_user_task_group`,
		`DELETE FROM economy_task WHERE key LIKE 'star\_%'`,
		`DELETE FROM economy_task_group WHERE key LIKE 'star\_%'`,
		`DELETE FROM base_category WHERE key_type='task' AND key LIKE 'star\_%'`,
		`DELETE FROM economy_streak WHERE key='daily_login'`,
		`DELETE FROM economy_level_table WHERE key='star_levels'`,
		`DELETE FROM economy_meta_rule WHERE key LIKE 'star\_%'`,
		`DELETE FROM economy_event_rule WHERE key LIKE 'star\_%'`,
		`DELETE FROM economy_event_type WHERE key IN ('round_finished','golden_star')`,
		`DELETE FROM economy_meta WHERE key IN ('xp','level')`,
		`DELETE FROM economy_user_prize`,
		`DELETE FROM economy_prize WHERE key LIKE 'star\_%'`,
		`UPDATE economy_referral_settings SET referrer_prize_id=NULL, referee_prize_id=NULL, goal_event_type_id=NULL`,
		`DELETE FROM economy_prize WHERE key IN ('referral_inviter','referral_invited')`,
		`DELETE FROM base_config_key WHERE key LIKE 'stargame.%'`,
		`DELETE FROM base_config_history WHERE key LIKE 'stargame.%'`,
		`DELETE FROM base_category WHERE key_type='product' AND key IN ('star_coin_packs','star_skins')`,
		`DELETE FROM economy_currency WHERE key='coin'`,
		`DELETE FROM base_tag WHERE key_type='user' AND key='star-vip'`,
		`DELETE FROM base_language WHERE code='ar'`,
		`DELETE FROM base_app_version WHERE platform IN ('web','android')`,
		`UPDATE economy_payment_method SET is_active=false, config='', updated_at=now()`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
}

func count(t *testing.T, db *sql.DB, q string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

func TestSeedPostgres(t *testing.T) {
	db := openDisposableTestDB(t)
	resetSeedData(t, db)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	env := map[string]string{secretbox.SecretEnv: testSecret}

	key, _ := secretbox.Economy.DeriveKey([]byte(testSecret), secretbox.EconomyConfigInfo)
	adminConfig, _ := secretbox.Economy.Seal(key, []byte(`{"merchantId":"real-merchant","sandbox":false}`))
	if _, err := db.Exec(`UPDATE economy_payment_method SET is_active=false, config=$1 WHERE method='zarinpal'`, adminConfig); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	reports := make([]*seedReport, 4)
	errs := make([]error, 4)
	for i := range reports {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			reports[i], errs[i] = seed(ctx, db, env)
		}(i)
	}
	wg.Wait()
	created := 0
	for i, err := range errs {
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
		created += len(reports[i].created)
	}
	if want := 1 + 1 + 2 + 14 + 1 + 7 + 2 + 2 + 2 + 4 + 2 + 1 + 1 + 4 + 2 + 2 + 1 + 2 + 2 + 1 + 11 + 4 + 2 + 2 + 2 + 5 + 1 + 1 + 1 + 5 + 2 + 2; created != want {
		t.Fatalf("created %d rows across nodes, want %d: %v", created, want, reports)
	}

	for q, want := range map[string]int{
		`SELECT count(*) FROM base_language WHERE code IN ('fa','en','ar')`:                3,
		`SELECT count(*) FROM base_language WHERE is_default`:                              1,
		`SELECT count(*) FROM economy_currency WHERE key='coin' AND default_value=50`:      1,
		`SELECT count(*) FROM base_category WHERE key IN ('star_coin_packs','star_skins')`: 2,
		`SELECT count(*) FROM economy_prize WHERE key LIKE 'star\_%'`:                      17,
		`SELECT count(*) FROM economy_achievement_tier r JOIN economy_achievement a ON a.id=r.achievement_id WHERE a.key='star_ach_rounds' AND r.prize_id IS NOT NULL AND r.grant_mode='claim'`:                                                            3,
		`SELECT count(*) FROM publication_banner b JOIN publication_banner_list l ON l.id=b.list_id WHERE l.key='menu_carousel' AND b.key LIKE 'star\_bn\_%' AND b.is_active`:                                                                              5,
		`SELECT count(*) FROM publication_banner WHERE key='star_bn_vip' AND cardinality(audience_tag_ids)=1 AND action_type='product' AND action_value='star_pack_vip'`:                                                                                   1,
		`SELECT count(*) FROM publication_banner WHERE key='star_bn_welcome' AND player_age='new' AND image_url ? 'en'`:                                                                                                                                    1,
		`SELECT count(*) FROM economy_prize_group_member m JOIN economy_prize_group g ON g.id=m.group_id WHERE g.key IN ('star_box_coins','star_box_big')`:                                                                                                 5,
		`SELECT count(*) FROM economy_chance_game WHERE key='star_wheel_daily' AND interval_time=1440 AND price IS NULL AND prize_type='prize' AND jsonb_array_length(prize)=4`:                                                                            1,
		`SELECT count(*) FROM economy_chance_game g JOIN economy_currency c ON c.id::text=g.price->>'currencyId' WHERE g.key='star_lucky_box' AND c.key='coin' AND (g.price->>'amount')::int=15 AND g.interval_time IS NULL AND g.prize_type='prizeGroup'`: 1,
		`SELECT count(*) FROM economy_achievement WHERE key='star_ach_legend' AND is_hidden AND metric='high_value'`:                                                                                                                                       1,
		`SELECT count(*) FROM economy_achievement a JOIN economy_task_group g ON g.id=a.task_group_id WHERE a.key='star_ach_hunter' AND g.key='star_hunter_chain'`:                                                                                         1,
		`SELECT count(*) FROM economy_gift g JOIN economy_prize z ON z.id=g.prize_id WHERE g.code='STARFAN' AND z.items @> '[{"type":"achievement","key":"star_ach_supporter"}]'`:                                                                          1,
		`SELECT count(*) FROM economy_event_type WHERE key IN ('round_finished','golden_star','login')`:                                                                                                                                                    3,
		`SELECT count(*) FROM economy_event_rule WHERE key LIKE 'star\_%' AND is_active`:                                                                                                                                                                   4,
		`SELECT count(*) FROM economy_meta_rule r JOIN economy_meta m ON m.id=r.meta_id WHERE r.key LIKE 'star\_%' AND m.key='xp'`:                                                                                                                         2,
		`SELECT count(*) FROM economy_level_step s JOIN economy_level_table t ON t.id=s.table_id WHERE t.key='star_levels'`:                                                                                                                                6,
		`SELECT count(*) FROM economy_streak_reward r JOIN economy_streak s ON s.id=r.streak_id WHERE s.key='daily_login' AND s.cycle_length=7`:                                                                                                            7,
		`SELECT count(*) FROM economy_product WHERE key LIKE 'star\_%' AND is_active`:                                                                                                                                                                      9,
		`SELECT count(*) FROM economy_item i JOIN base_category c ON c.id=i.type_id WHERE i.key LIKE 'star\_%' AND c.key_type='item_type'`:                                                                                                                 4,
		`SELECT count(*) FROM economy_item WHERE key='star_basket_classic' AND is_starter`:                                                                                                                                                                 1,
		`SELECT count(*) FROM economy_item WHERE key='star_frame_gold' AND unlock_level=2 AND level_table_id IS NOT NULL`:                                                                                                                                  1,
		`SELECT count(*) FROM economy_product WHERE key='star_starter_pack' AND is_one_time AND max_level=2 AND purchase_message->>'fa' <> ''`:                                                                                                             1,
		`SELECT count(*) FROM economy_product p JOIN economy_prize z ON z.id=p.prize_id WHERE p.key='star_item_ruby' AND p.check_contents AND z.items @> '[{"type":"item","key":"star_basket_ruby"}]'`:                                                     1,
		`SELECT count(*) FROM economy_product p JOIN base_tag t ON t.id::text = ANY(p.audience_tag_ids) WHERE p.key='star_pack_vip' AND t.key='star-vip' AND t.is_public`:                                                                                  1,
		`SELECT count(*) FROM economy_gift WHERE code IN ('STAR100','HALF50') AND is_active`:                                                                                                                                                               2,
		`SELECT count(*) FROM base_app_version WHERE platform IN ('web','android') AND latest_version='1.4.0' AND min_version='1.0.0'`:                                                                                                                     2,
		`SELECT count(*) FROM economy_referral_settings WHERE referrer_prize_id=(SELECT id FROM economy_prize WHERE key='referral_inviter') AND goal_event_type_id=(SELECT id FROM economy_event_type WHERE key='round_finished')`:                         1,
		`SELECT count(*) FROM base_config_key k JOIN base_config_history h ON h.key_id=k.id WHERE k.key IN ('stargame.banner','stargame.golden_rush') AND h.version=1`:                                                                                     2,
		`SELECT count(*) FROM economy_task WHERE key LIKE 'star\_%' AND grant_mode='claim'`:                                                                                                                                                                11,
		`SELECT count(*) FROM economy_task c JOIN economy_task p ON p.id=c.parent_id WHERE p.key='star_mission_starter' AND p.condition='children'`:                                                                                                        3,
		`SELECT count(*) FROM economy_task t JOIN economy_task_group g ON g.id=t.group_id WHERE g.key='star_hunter_chain'`:                                                                                                                                 3,
		`SELECT count(*) FROM economy_task WHERE key='star_mission_veteran' AND cardinality(depends_on)=1 AND min_level=2 AND level_table_id IS NOT NULL`:                                                                                                  1,
		`SELECT count(*) FROM economy_payment_method WHERE is_active`:                                                                                                                                                                                      4,
		`SELECT count(*) FROM economy_product WHERE key='star_skin_gold' AND coin_price @> '{"currencyKey":"coin","amount":80}' AND payment_methods = ARRAY['in_app']`:                                                                                     1,
		`SELECT count(*) FROM economy_product WHERE key='star_pack_small' AND price_toman=10000 AND sku @> '{"bazaar":"star_pack_small","myket":"star_pack_small"}'`:                                                                                       1,
		`SELECT count(*) FROM economy_gift WHERE code='STAR100' AND prize_id=(SELECT id FROM economy_prize WHERE key='star_coins_100') AND name->>'en'='Star Catcher: 100 free coins'`:                                                                     1,
		`SELECT count(*) FROM economy_gift WHERE code='HALF50' AND type='discount' AND discount_percent=50 AND prize_id IS NULL`:                                                                                                                           1,
		`SELECT count(*) FROM economy_prize WHERE key='star_coins_500' AND items @> '[{"type":"currency","key":"coin","action":"add","value":500}]' AND is_repeatable`:                                                                                     1,
		`SELECT count(*) FROM economy_prize WHERE key='star_skin_neon' AND items @> '[{"type":"storage","key":"skin_neon","action":"give","value":{"skin":"skin_neon"}}]' AND NOT is_repeatable`:                                                           1,
	} {
		if got := count(t, db, q); got != want {
			t.Errorf("%s = %d, want %d", q, got, want)
		}
	}

	var stillAdmin string
	_ = db.QueryRow(`SELECT config FROM economy_payment_method WHERE method='zarinpal'`).Scan(&stillAdmin)
	if stillAdmin != adminConfig {
		t.Fatal("an admin-configured payment method was overwritten")
	}
	for method, want := range map[string]string{"in_app": `{}`, "zibal": `{"sandbox":true}`, "bazaar": `{"packageName":"ir.nakama.stargame","sandbox":true}`, "myket": `{"packageName":"ir.nakama.stargame","sandbox":true}`} {
		var sealed string
		_ = db.QueryRow(`SELECT config FROM economy_payment_method WHERE method=$1`, method).Scan(&sealed)
		plain, err := secretbox.Economy.Open(key, sealed)
		if err != nil || string(plain) != want {
			t.Fatalf("%s config = %q %v", method, plain, err)
		}
	}

	if _, err := db.Exec(`UPDATE economy_product SET name='{"fa":"نام ادمین","en":"Admin name"}', is_active=false WHERE key='star_skin_gold'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE economy_payment_method SET is_active=false WHERE method='zibal'`); err != nil {
		t.Fatal(err)
	}
	again, err := seed(ctx, db, env)
	if err != nil || len(again.created) != 0 {
		t.Fatalf("second seed = %v %v", again, err)
	}
	if got := count(t, db, `SELECT count(*) FROM economy_product WHERE key='star_skin_gold' AND name->>'en'='Admin name' AND NOT is_active`); got != 1 {
		t.Fatal("admin product edit was overwritten")
	}
	if got := count(t, db, `SELECT count(*) FROM economy_payment_method WHERE method='zibal' AND NOT is_active`); got != 1 {
		t.Fatal("admin deactivation of a seeded method was overwritten")
	}

	resetSeedData(t, db)
	noSecret, err := seed(ctx, db, map[string]string{})
	if err != nil || len(noSecret.skipped) != 1 || count(t, db, `SELECT count(*) FROM economy_payment_method WHERE is_active`) != 0 {
		t.Fatalf("seed without secret = %v %v", noSecret, err)
	}
	if count(t, db, `SELECT count(*) FROM economy_product WHERE key LIKE 'star\_%'`) != 9 {
		t.Fatal("catalog must still be seeded without a secret")
	}
}
