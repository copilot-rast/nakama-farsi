package main

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/samplecleanup"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

var sharedSeedKinds = map[string]bool{"language": true, "currency": true, "meta": true, "referral rule": true, "method": true, "streak milestone": true}

var seedKindToCleanup = map[string]string{
	"category": "product_category", "prize": "prize", "player tag": "player_tag", "product": "product", "gift": "gift",
	"event type": "event_type", "event rule": "event_rule", "meta rule": "meta_rule", "level table": "level_table", "streak": "streak",
	"config key": "config_key", "task category": "task_category", "task group": "task_group", "task": "task", "item": "item",
	"achievement category": "achievement_category", "achievement": "achievement", "prize group": "prize_group", "chance game": "chance_game",
	"game category": "game_category", "game": "game", "banner placement": "banner_list", "banner": "banner", "sponsor": "sponsor",
	"automatic notification": "auto_rule", "app version": "app_version",
}

func methodConfig(t *testing.T, db *sql.DB, key []byte, method string) (string, bool) {
	t.Helper()
	var sealed string
	var active bool
	if err := db.QueryRow(`SELECT config, is_active FROM economy_payment_method WHERE method=$1`, method).Scan(&sealed, &active); err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	if sealed == "" {
		return "", active
	}
	plain, err := secretbox.Economy.Open(key, sealed)
	if err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	return string(plain), active
}

func kept(res *samplecleanup.Result, kind, key string) string {
	for _, tb := range res.Tables {
		if tb.Kind != kind {
			continue
		}
		for _, k := range tb.Kept {
			if k.Key == key {
				return k.Reason
			}
		}
	}
	return ""
}

func change(res *samplecleanup.Result, kind, key string) string {
	for _, c := range res.Changes {
		if c.Kind == kind && c.Key == key {
			return c.Action + ":" + c.Reason
		}
	}
	return ""
}

func TestSampleCleanupPostgres(t *testing.T) {
	db := openDisposableTestDB(t)
	resetSeedData(t, db)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM economy_product WHERE key='real_pack'`)
		db.Exec(`DELETE FROM economy_user_item WHERE user_id='11111111-1111-4111-8111-111111111111'`)
		resetSeedData(t, db)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	env := map[string]string{secretbox.SecretEnv: testSecret}
	key, _ := secretbox.Economy.DeriveKey([]byte(testSecret), secretbox.EconomyConfigInfo)
	seal := func(s string) string {
		v, err := secretbox.Economy.Seal(key, []byte(s))
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	adminZarinpal := seal(`{"merchantId":"real-merchant","sandbox":false}`)
	if _, err := db.Exec(`UPDATE economy_payment_method SET is_active=true, config=$1 WHERE method='zarinpal'`, adminZarinpal); err != nil {
		t.Fatal(err)
	}

	report, err := seed(ctx, db, env)
	if err != nil {
		t.Fatal(err)
	}

	known := map[string]bool{}
	for kind, keys := range samplecleanup.Keys() {
		for _, k := range keys {
			known[kind+":"+k] = true
		}
	}
	for _, c := range report.created {
		kind, k, _ := strings.Cut(c, ":")
		if sharedSeedKinds[kind] {
			continue
		}
		name, ok := seedKindToCleanup[kind]
		if !ok {
			t.Fatalf("seed kind %q has no cleanup mapping (%s)", kind, c)
		}
		if !known[name+":"+k] {
			t.Errorf("seeded %s is not in the cleanup list", c)
		}
	}

	adminMyket := seal(`{"packageName":"ir.real.app","sandbox":false}`)
	editedBazaar := seal(`{"packageName":"ir.nakama.stargame","sandbox":true,"rsaKey":"real"}`)
	for q, args := range map[string][]any{
		`UPDATE economy_payment_method SET config=$1 WHERE method='myket'`:  {adminMyket},
		`UPDATE economy_payment_method SET config=$1 WHERE method='bazaar'`: {editedBazaar},
		`INSERT INTO economy_product (id, key, name, description, is_active, payment_methods, sku, prize_id) SELECT gen_random_uuid(), 'real_pack', '{"en":"Real"}', '{}', true, '{in_app}', '{}', id FROM economy_prize WHERE key='star_coins_100'`: nil,
		`INSERT INTO economy_payment (id, user_id, product_id, method, status) SELECT gen_random_uuid(), '11111111-1111-4111-8111-111111111111', id, 'zarinpal', 'success' FROM economy_product WHERE key='star_pack_small'`:                            nil,
		`INSERT INTO economy_user_item (user_id, item_id, source) SELECT '11111111-1111-4111-8111-111111111111', id, 'starter' FROM economy_item WHERE key='star_basket_classic'`:                                                                    nil,
		`UPDATE publication_banner_list SET name='{"fa":"بنرهای ما","en":"Our banners"}' WHERE key='menu_carousel'`:                                                                                                                                  nil,
	} {
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	before := count(t, db, `SELECT count(*) FROM economy_prize WHERE key LIKE 'star\_%'`)
	preview, err := samplecleanup.Run(ctx, db, samplecleanup.Options{Env: env})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Applied || preview.Deleted < 80 || count(t, db, `SELECT count(*) FROM economy_prize WHERE key LIKE 'star\_%'`) != before {
		t.Fatalf("preview deleted=%d applied=%v, prizes %d -> %d", preview.Deleted, preview.Applied, before, count(t, db, `SELECT count(*) FROM economy_prize WHERE key LIKE 'star\_%'`))
	}
	if n, _ := samplecleanup.LastRun(ctx, db); n != nil && n.By == "test" {
		t.Fatal("preview must not record a run")
	}

	res, err := samplecleanup.Run(ctx, db, samplecleanup.Options{Apply: true, By: "test", Env: env})
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != preview.Deleted || res.Kept != preview.Kept || res.Changed != preview.Changed {
		t.Fatalf("apply %d/%d/%d differs from preview %d/%d/%d", res.Deleted, res.Kept, res.Changed, preview.Deleted, preview.Kept, preview.Changed)
	}
	for kind, k := range map[string]string{"product": "star_pack_small", "prize": "star_coins_100", "banner_list": "menu_carousel"} {
		if kept(res, kind, k) == "" {
			t.Errorf("%s %s should be kept: %+v", kind, k, res.Tables)
		}
	}
	if r := kept(res, "product", "star_pack_small"); r != "in_use_switched_off" {
		t.Errorf("star_pack_small reason %q", r)
	}
	if r := kept(res, "banner_list", "menu_carousel"); r != "edited" {
		t.Errorf("menu_carousel reason %q", r)
	}
	if r := kept(res, "currency", "coin"); r != "edited" {
		t.Errorf("coin must stay while a kept prize pays it, reason %q", r)
	}
	if res.Kept != 5 {
		t.Errorf("kept %d rows, want 5: %+v", res.Kept, res.Tables)
	}
	if res.Related < 1 {
		t.Errorf("the owned starter basket must count as a related row, got %d", res.Related)
	}

	for q, want := range map[string]int{
		`SELECT count(*) FROM economy_prize WHERE key LIKE 'star\_%'`:                                                   1,
		`SELECT count(*) FROM economy_product WHERE key LIKE 'star\_%'`:                                                 1,
		`SELECT count(*) FROM economy_product WHERE key='star_pack_small' AND NOT is_active`:                            1,
		`SELECT count(*) FROM economy_product WHERE key='real_pack' AND is_active`:                                      1,
		`SELECT count(*) FROM economy_payment`:                                                                          1,
		`SELECT count(*) FROM economy_item WHERE key LIKE 'star\_%'`:                                                    0,
		`SELECT count(*) FROM economy_user_item WHERE user_id='11111111-1111-4111-8111-111111111111'`:                   0,
		`SELECT count(*) FROM economy_task WHERE key LIKE 'star\_%'`:                                                    0,
		`SELECT count(*) FROM economy_task_group WHERE key='star_hunter_chain'`:                                         0,
		`SELECT count(*) FROM economy_achievement WHERE key LIKE 'star\_%'`:                                             0,
		`SELECT count(*) FROM economy_chance_game WHERE key LIKE 'star\_%'`:                                             0,
		`SELECT count(*) FROM economy_prize_group WHERE key LIKE 'star\_%'`:                                             0,
		`SELECT count(*) FROM economy_game WHERE key IN ('stargame','stargame_tictactoe')`:                              0,
		`SELECT count(*) FROM economy_event_rule WHERE key LIKE 'star\_%'`:                                              0,
		`SELECT count(*) FROM economy_meta_rule WHERE key LIKE 'star\_%'`:                                               0,
		`SELECT count(*) FROM economy_level_table WHERE key='star_levels'`:                                              0,
		`SELECT count(*) FROM economy_streak WHERE key='daily_login'`:                                                   0,
		`SELECT count(*) FROM economy_event_type WHERE key IN ('round_finished','golden_star')`:                         0,
		`SELECT count(*) FROM economy_event_type WHERE key='login'`:                                                     1,
		`SELECT count(*) FROM economy_gift WHERE code IN ('STAR100','HALF50','STARFAN')`:                                0,
		`SELECT count(*) FROM publication_banner WHERE key LIKE 'star\_bn\_%'`:                                          0,
		`SELECT count(*) FROM publication_banner_list WHERE key='menu_carousel'`:                                        1,
		`SELECT count(*) FROM publication_banner_list WHERE key='luck_sponsor'`:                                         0,
		`SELECT count(*) FROM publication_sponsor WHERE key='star_sp_nova'`:                                             0,
		`SELECT count(*) FROM publication_auto_rule WHERE key LIKE 'star\_auto\_%'`:                                     0,
		`SELECT count(*) FROM base_config_key WHERE key LIKE 'stargame.%'`:                                              0,
		`SELECT count(*) FROM base_config_history WHERE key LIKE 'stargame.%'`:                                          0,
		`SELECT count(*) FROM base_category WHERE key LIKE 'star\_%'`:                                                   1,
		`SELECT count(*) FROM base_tag WHERE key='star-vip'`:                                                            0,
		`SELECT count(*) FROM base_app_version WHERE platform IN ('web','android')`:                                     0,
		`SELECT count(*) FROM economy_prize WHERE key IN ('referral_inviter','referral_invited')`:                       0,
		`SELECT count(*) FROM economy_referral_settings WHERE referrer_prize_id IS NULL AND goal_event_type_id IS NULL`: 1,
		`SELECT count(*) FROM base_language WHERE code IN ('fa','en','ar')`:                                             3,
		`SELECT count(*) FROM economy_currency WHERE key='coin'`:                                                        1,
		`SELECT count(*) FROM economy_meta WHERE key IN ('xp','level')`:                                                 2,
		`SELECT count(*) FROM storage WHERE collection='base_maintenance' AND key='stargame_cleanup'`:                   1,
	} {
		if got := count(t, db, q); got != want {
			t.Errorf("%s = %d, want %d", q, got, want)
		}
	}

	for method, want := range map[string]string{"zarinpal": `{"merchantId":"real-merchant","sandbox":false}`, "myket": `{"packageName":"ir.real.app","sandbox":false}`, "bazaar": `{"packageName":"ir.nakama.stargame","sandbox":true,"rsaKey":"real"}`, "zibal": ""} {
		got, active := methodConfig(t, db, key, method)
		if got != want || active != (want != "") {
			t.Errorf("%s config = %q active=%v, want %q", method, got, active, want)
		}
	}
	if _, active := methodConfig(t, db, key, "in_app"); !active {
		t.Error("in_app must stay on: coin purchases use it")
	}
	if change(res, "payment_method", "zibal") != "reset:" || change(res, "payment_method", "bazaar") != "kept:sample_package_edited" || change(res, "payment_method", "zarinpal") != "" || change(res, "payment_method", "myket") != "" {
		t.Errorf("method changes = %+v", res.Changes)
	}
	if change(res, "referral_rule", "default") != "reset:" {
		t.Errorf("referral change = %+v", res.Changes)
	}

	last, err := samplecleanup.LastRun(ctx, db)
	if err != nil || last == nil || last.By != "test" || last.Deleted != res.Deleted {
		t.Fatalf("last run = %+v %v", last, err)
	}

	again, err := samplecleanup.Run(ctx, db, samplecleanup.Options{Apply: true, By: "test-2", Env: env})
	if err != nil {
		t.Fatal(err)
	}
	if again.Deleted != 0 || again.Changed != 0 || again.Related != 0 || len(again.Leaderboards) != 0 {
		t.Fatalf("second run is not a no-op: %+v", again)
	}
	keptKeys := []string{}
	for _, tb := range again.Tables {
		for _, k := range tb.Kept {
			keptKeys = append(keptKeys, tb.Kind+":"+k.Key)
		}
	}
	sort.Strings(keptKeys)
	if strings.Join(keptKeys, ",") != "banner_list:menu_carousel,currency:coin,prize:star_coins_100,product:star_pack_small,product_category:star_coin_packs" {
		t.Fatalf("second run kept = %v", keptKeys)
	}
}

func TestSampleCleanupForkPostgres(t *testing.T) {
	db := openDisposableTestDB(t)
	resetSeedData(t, db)
	const referee = "22222222-2222-4222-8222-222222222222"
	const referrer = "33333333-3333-4333-8333-333333333333"
	cleanup := func() {
		db.Exec(`DELETE FROM economy_referral`)
		db.Exec(`DELETE FROM economy_event_type WHERE key='rs_match_finished'`)
		db.Exec(`DELETE FROM economy_prize WHERE key IN ('rs_oil_100','rs_oil_50')`)
		db.Exec(`DELETE FROM users WHERE id IN ($1,$2)`, referee, referrer)
	}
	cleanup()
	t.Cleanup(func() {
		cleanup()
		resetSeedData(t, db)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	env := map[string]string{secretbox.SecretEnv: testSecret}
	if _, err := seed(ctx, db, env); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO users (id, username) VALUES ('` + referee + `','fork_referee'), ('` + referrer + `','fork_referrer')`,
		`INSERT INTO economy_event_type (id, key, name, client_allowed, value_min, value_max) VALUES (gen_random_uuid(), 'rs_match_finished', '{"en":"Match finished"}', true, 1, 1)`,
		`INSERT INTO economy_prize (id, key, name, items) VALUES (gen_random_uuid(), 'rs_oil_100', '{"en":"100 Oil"}', '[{"type":"currency","key":"oil","action":"add","value":100}]'), (gen_random_uuid(), 'rs_oil_50', '{"en":"50 Oil"}', '[{"type":"currency","key":"oil","action":"add","value":50}]')`,
		`INSERT INTO economy_referral (id, referrer_id, referee_id, code, status, referrer_prize_id, referee_prize_id, goal_event_type_id, goal_measure, goal_value)
			SELECT gen_random_uuid(), '` + referrer + `', '` + referee + `', 'FORKCODE', 'pending', (SELECT id FROM economy_prize WHERE key='referral_inviter'), (SELECT id FROM economy_prize WHERE key='referral_invited'), (SELECT id FROM economy_event_type WHERE key='round_finished'), 'count', 1`,
		`UPDATE economy_referral_settings SET goal_event_type_id=(SELECT id FROM economy_event_type WHERE key='rs_match_finished'), referrer_prize_id=(SELECT id FROM economy_prize WHERE key='rs_oil_100'), referee_prize_id=(SELECT id FROM economy_prize WHERE key='rs_oil_50')`,
		`UPDATE base_app_version SET update_message='{"en":"Road Storm has an update"}', store_links='[{"store":"bazaar","url":"https://cafebazaar.ir/app/ir.roadstorm.game","label":{}},{"store":"myket","url":"https://myket.ir/app/ir.nakama.stargame","label":{}}]' WHERE platform='android'`,
		`UPDATE base_app_version SET update_message='{"en":"Play on the web"}' WHERE platform='web'`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	res, err := samplecleanup.Run(ctx, db, samplecleanup.Options{Apply: true, By: "test", Env: env})
	if err != nil {
		t.Fatal(err)
	}
	if c := change(res, "app_version_link", "android"); c != "reset:removed_1" {
		t.Errorf("android link change %q: %+v", c, res.Changes)
	}
	if c := change(res, "app_version", "web"); c != "deleted:only_sample_links" {
		t.Errorf("web change %q: %+v", c, res.Changes)
	}
	if r := kept(res, "event_type", "round_finished"); r != "in_use_switched_off" {
		t.Errorf("round_finished reason %q", r)
	}
	for q, want := range map[string]int{
		`SELECT count(*) FROM base_app_version WHERE platform='web'`:                                                       0,
		`SELECT count(*) FROM base_app_version WHERE platform='android' AND store_links::text LIKE '%ir.roadstorm.game%'`:  1,
		`SELECT count(*) FROM base_app_version WHERE platform='android' AND store_links::text LIKE '%ir.nakama.stargame%'`: 0,
		`SELECT count(*) FROM economy_currency WHERE key='coin'`:                                                           0,
		`SELECT count(*) FROM economy_event_type WHERE key='rs_match_finished' AND is_active`:                              1,
		`SELECT count(*) FROM economy_event_type WHERE key='round_finished' AND NOT is_active`:                             1,
		`SELECT count(*) FROM economy_referral_settings WHERE goal_event_type_id=(SELECT id FROM economy_event_type WHERE key='rs_match_finished')`: 1,
	} {
		if got := count(t, db, q); got != want {
			t.Errorf("%s = %d, want %d", q, got, want)
		}
	}
	again, err := samplecleanup.Run(ctx, db, samplecleanup.Options{Apply: true, By: "test-2", Env: env})
	if err != nil {
		t.Fatal(err)
	}
	if again.Changed != 0 || again.Deleted != 0 {
		t.Fatalf("second run is not a no-op: %+v", again)
	}
}
