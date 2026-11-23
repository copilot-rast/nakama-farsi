package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/configtransfer"
)

const (
	ctDBAEnv = "CONFIGTRANSFER_TEST_DB_A_URL"
	ctDBBEnv = "CONFIGTRANSFER_TEST_DB_B_URL"
)

var ctSecrets = []string{"SECRET-MERCHANT-ZP-999", "SECRET-BAZAAR-777", "SECRET-GAME-KEY-123", "SECRET-WEBHOOK-555", "SECRET-RULES-TOKEN-321", "SECRET-BOARD-PASS-42"}

func openCTDB(t *testing.T, env string) *sql.DB {
	t.Helper()
	dsn := os.Getenv(env)
	if dsn == "" {
		t.Skip(env + " not set; PostgreSQL NOT_RUN")
	}
	t.Setenv(disposableTestDBEnv, dsn)
	return openDisposableTestDB(t)
}

func ctExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("%v\n%s", err, q)
	}
}

func ctClean(t *testing.T, db *sql.DB) {
	t.Helper()
	ctExec(t, db, `UPDATE economy_referral_settings SET referrer_prize_id=NULL, referee_prize_id=NULL, goal_event_type_id=NULL`)
	ctExec(t, db, `UPDATE economy_task SET parent_id=NULL, depends_on='{}' WHERE key LIKE 'ct\_%'`)
	for _, q := range []string{
		`DELETE FROM economy_achievement WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_task WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_task_group WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_streak WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_meta_relation WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_meta_rule WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_event_rule WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_product WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_item WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_level_table WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_prize_group WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_prize WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_event_type WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_meta WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_currency WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_game WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_chance_game WHERE key LIKE 'ct\_%'`,
		`DELETE FROM leaderboardx_board WHERE key LIKE 'ct\_%'`,
		`DELETE FROM base_tag WHERE key LIKE 'ct\_%'`,
		`DELETE FROM base_category WHERE key LIKE 'ct\_%'`,
		`DELETE FROM economy_payment_method WHERE method IN ('zarinpal','bazaar')`,
		`DELETE FROM base_app_version WHERE platform='web'`,
	} {
		ctExec(t, db, q)
	}
}

func ctSeedA(t *testing.T, db *sql.DB, s *service) {
	t.Helper()
	ctExec(t, db, `INSERT INTO base_category (id, key_type, key, name) VALUES (gen_random_uuid(),'item_type','ct_car','{"fa":"ماشین","en":"Car"}'), (gen_random_uuid(),'item','ct_cars','{"en":"Cars"}')`)
	ctExec(t, db, `INSERT INTO base_tag (id, key_type, key, name, is_public) VALUES (gen_random_uuid(),'item','ct_vip','{"en":"VIP"}',false), (gen_random_uuid(),'user','ct_vip_user','{"en":"VIP player"}',true)`)
	ctExec(t, db, `INSERT INTO economy_currency (id, key, name, type, default_value, tag_ids) VALUES (gen_random_uuid(),'ct_gem','{"en":"Gem"}','hard',5, ARRAY[(SELECT id::text FROM base_tag WHERE key='ct_vip')])`)
	ctExec(t, db, `INSERT INTO economy_meta (id, key, name, default_value) VALUES (gen_random_uuid(),'ct_xp','{"en":"XP"}',0)`)
	ctExec(t, db, `INSERT INTO economy_event_type (id, key, name, client_allowed) VALUES (gen_random_uuid(),'ct_win','{"en":"Win"}',true)`)
	ctExec(t, db, `INSERT INTO economy_prize (id, key, name, items, tag_ids, category_id) VALUES
		(gen_random_uuid(),'ct_prize1','{"en":"Ten gems"}','[{"type":"currency","key":"ct_gem","action":"add","value":10}]', ARRAY[(SELECT id::text FROM base_tag WHERE key='ct_vip')], (SELECT id FROM base_category WHERE key='ct_cars')),
		(gen_random_uuid(),'ct_prize2','{"en":"Pack"}','[{"type":"currency","key":"ct_gem","action":"add","value":99.5}]','{}',NULL)`)
	ctExec(t, db, `INSERT INTO economy_prize_group (id, key, name) VALUES (gen_random_uuid(),'ct_group','{"en":"Group"}')`)
	ctExec(t, db, `INSERT INTO economy_prize_group_member (group_id, prize_id, weight) SELECT g.id, p.id, CASE p.key WHEN 'ct_prize1' THEN 3 ELSE 1 END FROM economy_prize_group g, economy_prize p WHERE g.key='ct_group' AND p.key IN ('ct_prize1','ct_prize2')`)
	ctExec(t, db, `INSERT INTO economy_level_table (id, key, name, meta_id) VALUES (gen_random_uuid(),'ct_levels','{"en":"Levels"}',(SELECT id FROM economy_meta WHERE key='ct_xp'))`)
	ctExec(t, db, `INSERT INTO economy_level_step (id, table_id, level, threshold, prize_id) SELECT gen_random_uuid(), t.id, l, l*100, CASE WHEN l=2 THEN (SELECT id FROM economy_prize WHERE key='ct_prize1') END FROM economy_level_table t, generate_series(1,3) l WHERE t.key='ct_levels'`)
	ctExec(t, db, `INSERT INTO economy_item (id, key, type_id, name, category_id, tag_ids, level_table_id, unlock_level, metadata) VALUES
		(gen_random_uuid(),'ct_car1',(SELECT id FROM base_category WHERE key='ct_car'),'{"en":"Car 1"}',(SELECT id FROM base_category WHERE key='ct_cars'), ARRAY[(SELECT id::text FROM base_tag WHERE key='ct_vip')],(SELECT id FROM economy_level_table WHERE key='ct_levels'),2,'{"speed":3,"apiKey":"SECRET-GAME-KEY-123"}')`)
	ctExec(t, db, `INSERT INTO economy_product (id, key, name, is_active, prize_id, level_table_id, audience_tag_ids, price_toman, payment_methods) VALUES
		(gen_random_uuid(),'ct_pack','{"en":"Pack"}',true,(SELECT id FROM economy_prize WHERE key='ct_prize2'),(SELECT id FROM economy_level_table WHERE key='ct_levels'), ARRAY[(SELECT id::text FROM base_tag WHERE key='ct_vip_user')], 50000, ARRAY['zarinpal'])`)
	ctExec(t, db, `INSERT INTO economy_event_rule (id, key, name, event_type_id, measure, period, threshold, repeat, prize_id, grant_mode) VALUES
		(gen_random_uuid(),'ct_rule','{"en":"Rule"}',(SELECT id FROM economy_event_type WHERE key='ct_win'),'count','day',3,'period',(SELECT id FROM economy_prize WHERE key='ct_prize1'),'auto')`)
	ctExec(t, db, `INSERT INTO economy_meta_rule (id, key, name, event_type_id, meta_id, action, value_source, amount) VALUES
		(gen_random_uuid(),'ct_mrule','{"en":"XP per win"}',(SELECT id FROM economy_event_type WHERE key='ct_win'),(SELECT id FROM economy_meta WHERE key='ct_xp'),'add','fixed',25)`)
	ctExec(t, db, `INSERT INTO economy_meta_relation (id, key, name, meta_id, operator, value, targets, prize_type, prize_id) VALUES
		(gen_random_uuid(),'ct_rel','{"en":"Rel"}',(SELECT id FROM economy_meta WHERE key='ct_xp'),'gte',100, jsonb_build_array(jsonb_build_object('metaId',(SELECT id::text FROM economy_meta WHERE key='ct_xp'),'action','add','amount',1)),'prize',(SELECT id FROM economy_prize WHERE key='ct_prize1'))`)
	ctExec(t, db, `INSERT INTO economy_task_group (id, key, name) VALUES (gen_random_uuid(),'ct_tg','{"en":"Daily"}')`)
	ctExec(t, db, `INSERT INTO economy_task (id, key, name, condition, event_type_id, target, group_id, prize_id) VALUES
		(gen_random_uuid(),'ct_t1','{"en":"Win once"}','event',(SELECT id FROM economy_event_type WHERE key='ct_win'),1,(SELECT id FROM economy_task_group WHERE key='ct_tg'),(SELECT id FROM economy_prize WHERE key='ct_prize1'))`)
	ctExec(t, db, `INSERT INTO economy_task (id, key, name, condition, event_type_id, target, parent_id, depends_on) VALUES
		(gen_random_uuid(),'ct_t0','{"en":"Win twice"}','event',(SELECT id FROM economy_event_type WHERE key='ct_win'),2,(SELECT id FROM economy_task WHERE key='ct_t1'),ARRAY[(SELECT id::text FROM economy_task WHERE key='ct_t1')])`)
	ctExec(t, db, `INSERT INTO economy_achievement (id, key, name, description, hint, source, task_id) VALUES
		(gen_random_uuid(),'ct_ach','{"en":"Ach"}','{}','{}','task',(SELECT id FROM economy_task WHERE key='ct_t1'))`)
	ctExec(t, db, `INSERT INTO economy_achievement_tier (id, achievement_id, tier, threshold, name, prize_id) SELECT gen_random_uuid(), a.id, 1, 1, '{"en":"Bronze"}', (SELECT id FROM economy_prize WHERE key='ct_prize1') FROM economy_achievement a WHERE a.key='ct_ach'`)
	ctExec(t, db, `INSERT INTO economy_streak (id, key, name, event_type_id) VALUES (gen_random_uuid(),'ct_streak','{"en":"Streak"}',(SELECT id FROM economy_event_type WHERE key='ct_win'))`)
	ctExec(t, db, `INSERT INTO economy_streak_reward (id, streak_id, day, prize_id) SELECT gen_random_uuid(), s.id, d, (SELECT id FROM economy_prize WHERE key='ct_prize1') FROM economy_streak s, generate_series(1,2) d WHERE s.key='ct_streak'`)
	ctExec(t, db, `INSERT INTO economy_referral_settings (id, is_enabled) VALUES (1, true) ON CONFLICT (id) DO NOTHING`)
	ctExec(t, db, `UPDATE economy_referral_settings SET is_enabled=true, referrer_prize_id=(SELECT id FROM economy_prize WHERE key='ct_prize1'), goal_event_type_id=(SELECT id FROM economy_event_type WHERE key='ct_win'), goal_value=3`)
	ctExec(t, db, `INSERT INTO economy_game (id, key, name, settings, rooms, requirements, min_players, max_players) VALUES
		(gen_random_uuid(),'ct_game','{"en":"Race"}','{"mode":"x","apiKey":"SECRET-GAME-KEY-123","nested":{"webhookSecret":"SECRET-WEBHOOK-555","keep":1}}',
		 jsonb_build_array(jsonb_build_object('key','r1','prizeId',(SELECT id::text FROM economy_prize WHERE key='ct_prize1'),'requirements',jsonb_build_object('itemIds',jsonb_build_array((SELECT id::text FROM economy_item WHERE key='ct_car1')),'metas',jsonb_build_array(jsonb_build_object('metaId',(SELECT id::text FROM economy_meta WHERE key='ct_xp'),'operator','gte','value',1))))),
		 jsonb_build_object('tagIds',jsonb_build_array((SELECT id::text FROM base_tag WHERE key='ct_vip_user'))),2,4)`)
	ctExec(t, db, `INSERT INTO economy_chance_game (id, key, name, prize_type, prize, price) VALUES
		(gen_random_uuid(),'ct_chest','{"en":"Chest"}','prize', jsonb_build_array(jsonb_build_object('chance',100,'prizeId',(SELECT id::text FROM economy_prize WHERE key='ct_prize1'))), jsonb_build_object('currencyId',gen_random_uuid()::text,'amount',5))`)
	ctExec(t, db, `INSERT INTO leaderboardx_board (id, key, name, metadata, rewards) VALUES (gen_random_uuid(),'ct_board','{"en":"Weekly"}','{"password":"SECRET-BOARD-PASS-42","color":"red"}','[{"fromRank":1,"toRank":1,"prizeKey":"ct_prize1"}]')`)
	ctExec(t, db, `INSERT INTO base_app_version (id, platform, latest_version, min_version, store_links, update_message, force_message, is_active) VALUES (gen_random_uuid(),'web','9.9.9','1.0.0','{}','{}','{}',true)`)
	ctExec(t, db, `INSERT INTO base_server_rules (id, value, updated_by) VALUES (1, jsonb_build_object('maintenance',jsonb_build_object('enabled',false,'allowTagIds',jsonb_build_array((SELECT id::text FROM base_tag WHERE key='ct_vip_user'))),'docs',jsonb_build_object('enabled',true),'integration',jsonb_build_object('apiToken','SECRET-RULES-TOKEN-321','url','https://x.test')),'test')
		ON CONFLICT (id) DO UPDATE SET value=EXCLUDED.value`)
	ctSetMethod(t, db, s, "zarinpal", true, `{"merchantId":"SECRET-MERCHANT-ZP-999","sandbox":false}`)
	ctSetMethod(t, db, s, "bazaar", true, `{"packageName":"com.ct.game","apiSecret":"SECRET-BAZAAR-777","sandbox":false}`)
}

func ctSetMethod(t *testing.T, db *sql.DB, s *service, method string, active bool, plain string) {
	t.Helper()
	sealed, err := seal(s.cfg.sealKey, []byte(plain))
	if err != nil {
		t.Fatal(err)
	}
	ctExec(t, db, `INSERT INTO economy_payment_method (method, is_active, config) VALUES ($1,$2,$3) ON CONFLICT (method) DO UPDATE SET is_active=EXCLUDED.is_active, config=EXCLUDED.config`, method, active, sealed)
}

func ctRegister(t *testing.T, s *service) {
	t.Helper()
	all := append(append(configtransfer.BaseAreas(), s.transferAreas()...), configtransfer.LeaderboardAreas()...)
	keys := make([]string, 0, len(all))
	for _, a := range all {
		keys = append(keys, a.Key)
	}
	configtransfer.Unregister(keys...)
	if err := configtransfer.Register(all...); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { configtransfer.Unregister(keys...) })
}

func ctExportJSON(t *testing.T, db *sql.DB, name string) (*configtransfer.File, []byte) {
	t.Helper()
	f, err := configtransfer.Export(context.Background(), db, configtransfer.ExportOptions{Source: configtransfer.Source{Name: name, By: "admin"}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	back, err := configtransfer.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return back, raw
}

func ctRow(f *configtransfer.File, area, table, key string) *configtransfer.Row {
	a := f.Area(area)
	if a == nil {
		return nil
	}
	tb := a.Table(table)
	if tb == nil {
		return nil
	}
	for i := range tb.Rows {
		if tb.Rows[i].Key == key {
			return &tb.Rows[i]
		}
	}
	return nil
}

func ctStatus(res *configtransfer.Result, table, key string) *configtransfer.RowResult {
	for _, a := range res.Areas {
		for i := range a.Rows {
			if a.Rows[i].Table == table && a.Rows[i].Key == key {
				return &a.Rows[i]
			}
		}
	}
	return nil
}

func ctID(t *testing.T, db *sql.DB, q string, args ...any) string {
	t.Helper()
	var id string
	if err := db.QueryRow(q, args...).Scan(&id); err != nil {
		t.Fatalf("%v: %s", err, q)
	}
	return id
}

func TestConfigTransferRoundTripPostgres(t *testing.T) {
	dbA := openCTDB(t, ctDBAEnv)
	defer dbA.Close()
	dbB := openCTDB(t, ctDBBEnv)
	defer dbB.Close()
	s := testService()
	ctRegister(t, s)
	ctClean(t, dbA)
	ctClean(t, dbB)
	ctSeedA(t, dbA, s)

	ctExec(t, dbB, `INSERT INTO economy_game (id, key, name, settings) VALUES (gen_random_uuid(),'ct_game','{"en":"Old"}','{"mode":"old","apiKey":"B-OWN-GAME-KEY","nested":{"webhookSecret":"B-OWN-HOOK","keep":0}}')`)
	ctSetMethod(t, dbB, s, "bazaar", false, `{"packageName":"com.old","apiSecret":"B-OWN-BAZAAR"}`)
	ctExec(t, dbB, `INSERT INTO base_category (id, key_type, key, name) VALUES (gen_random_uuid(),'item_type','ct_car','{"en":"x"}')`)
	ctExec(t, dbB, `INSERT INTO economy_item (id, key, type_id, name, is_active) VALUES (gen_random_uuid(),'ct_only_b',(SELECT id FROM base_category WHERE key='ct_car'),'{"en":"B only"}',true)`)

	f, raw := ctExportJSON(t, dbA, "dev-server")
	text := string(raw)
	for _, sec := range append(ctSecrets, strings.Repeat("ab", 24)) {
		if strings.Contains(text, sec) {
			t.Fatalf("secret %q leaked into the export", sec)
		}
	}
	var sealed string
	if err := dbA.QueryRow(`SELECT config FROM economy_payment_method WHERE method='zarinpal'`).Scan(&sealed); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, sealed) {
		t.Fatal("sealed payment config leaked into the export")
	}
	if strings.Contains(strings.ToLower(text), "merchantid") || strings.Contains(text, "apiSecret") {
		t.Fatal("secret field names must not be in the export at all")
	}
	if f.Format != configtransfer.Format || f.Version != configtransfer.FormatVersion || f.Source.Name != "dev-server" || f.Source.ExportedAt == "" {
		t.Fatalf("bad header %+v", f.Source)
	}
	if f.Counts["items"] < 1 || f.Counts["payment_methods"] < 2 {
		t.Fatalf("bad counts %+v", f.Counts)
	}
	item := ctRow(f, "items", "economy_item", "ct_car1")
	if item == nil || item.Values["type_id"] != "item_type/ct_car" || item.Values["level_table_id"] != "ct_levels" {
		t.Fatalf("refs must be keys: %+v", item)
	}
	if tags, _ := item.Values["tag_ids"].([]any); len(tags) != 1 || tags[0] != "item/ct_vip" {
		t.Fatalf("tag ids must be keys: %+v", item.Values["tag_ids"])
	}
	if _, ok := item.Values["id"]; ok {
		t.Fatal("uuid ids must not be exported")
	}
	task := ctRow(f, "tasks", "economy_task", "ct_t0")
	if task == nil || task.Values["parent_id"] != "ct_t1" {
		t.Fatalf("self ref must be a key: %+v", task)
	}
	game := ctRow(f, "games", "economy_game", "ct_game")
	gameText, _ := json.Marshal(game.Values)
	if !strings.Contains(string(gameText), `"prizeId":"ct_prize1"`) || !strings.Contains(string(gameText), `"itemIds":["ct_car1"]`) || !strings.Contains(string(gameText), `"metaId":"ct_xp"`) || !strings.Contains(string(gameText), `"tagIds":["user/ct_vip_user"]`) {
		t.Fatalf("json refs must be keys: %s", gameText)
	}

	ctx := context.Background()
	prev, err := configtransfer.Apply(ctx, dbB, f, configtransfer.Options{By: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if prev.Applied {
		t.Fatal("preview must not apply")
	}
	if st := ctStatus(prev, "economy_item", "ct_car1"); st == nil || st.Status != configtransfer.StatusNew {
		t.Fatalf("preview item: %+v", st)
	}
	var n int
	_ = dbB.QueryRow(`SELECT count(*) FROM economy_item WHERE key='ct_car1'`).Scan(&n)
	if n != 0 {
		t.Fatal("preview wrote rows")
	}

	res, err := configtransfer.Apply(ctx, dbB, f, configtransfer.Options{Commit: true, By: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Applied || res.Totals.Errors != 0 {
		for _, a := range res.Areas {
			for _, r := range a.Rows {
				if r.Status == configtransfer.StatusError {
					t.Logf("error row %s %s: %s", r.Table, r.Key, r.Reason)
				}
			}
		}
		t.Fatalf("apply failed: %+v", res.Totals)
	}
	for _, c := range []struct{ table, key, status string }{
		{"economy_item", "ct_car1", configtransfer.StatusNew},
		{"economy_task", "ct_t0", configtransfer.StatusNew},
		{"base_category", "item_type/ct_car", configtransfer.StatusChanged},
		{"economy_game", "ct_game", configtransfer.StatusChanged},
		{"economy_payment_method", "zarinpal", configtransfer.StatusSecretNeeded},
		{"economy_payment_method", "bazaar", configtransfer.StatusChanged},
		{"economy_level_step", "ct_levels/2", configtransfer.StatusNew},
		{"economy_prize_group_member", "ct_group/ct_prize1", configtransfer.StatusNew},
	} {
		st := ctStatus(res, c.table, c.key)
		if st == nil || st.Status != c.status {
			t.Fatalf("%s %s: want %s, got %+v", c.table, c.key, c.status, st)
		}
	}
	broken := ctStatus(res, "economy_chance_game", "ct_chest")
	if broken == nil || broken.Status != configtransfer.StatusMissingRefs || broken.Reason != configtransfer.ReasonSourceBroken || len(broken.Missing) != 1 || broken.Missing[0].Field != "price" || !broken.Missing[0].SourceBroken {
		t.Fatalf("a ref that is already broken in the source must be named and marked: %+v", broken)
	}
	if st := ctStatus(res, "economy_item", "ct_only_b"); st != nil {
		t.Fatal("rows not in the file must not be touched without switch off")
	}

	if got, want := ctID(t, dbB, `SELECT type_id::text FROM economy_item WHERE key='ct_car1'`), ctID(t, dbB, `SELECT id::text FROM base_category WHERE key='ct_car'`); got != want {
		t.Fatalf("type ref not resolved to B id: %s != %s", got, want)
	}
	if ctID(t, dbB, `SELECT id::text FROM economy_prize WHERE key='ct_prize1'`) == ctID(t, dbA, `SELECT id::text FROM economy_prize WHERE key='ct_prize1'`) {
		t.Fatal("B must get its own ids")
	}
	if got, want := ctID(t, dbB, `SELECT rooms->0->>'prizeId' FROM economy_game WHERE key='ct_game'`), ctID(t, dbB, `SELECT id::text FROM economy_prize WHERE key='ct_prize1'`); got != want {
		t.Fatalf("json prize ref not resolved: %s != %s", got, want)
	}
	if got, want := ctID(t, dbB, `SELECT rooms->0->'requirements'->'itemIds'->>0 FROM economy_game WHERE key='ct_game'`), ctID(t, dbB, `SELECT id::text FROM economy_item WHERE key='ct_car1'`); got != want {
		t.Fatalf("json item ref not resolved: %s != %s", got, want)
	}
	if got, want := ctID(t, dbB, `SELECT depends_on[1] FROM economy_task WHERE key='ct_t0'`), ctID(t, dbB, `SELECT id::text FROM economy_task WHERE key='ct_t1'`); got != want {
		t.Fatalf("depends_on not resolved: %s != %s", got, want)
	}
	if got, want := ctID(t, dbB, `SELECT value->'maintenance'->'allowTagIds'->>0 FROM base_server_rules WHERE id=1`), ctID(t, dbB, `SELECT id::text FROM base_tag WHERE key='ct_vip_user'`); got != want {
		t.Fatalf("server rules tag not resolved: %s != %s", got, want)
	}
	if got := ctID(t, dbB, `SELECT settings::text FROM economy_game WHERE key='ct_game'`); !strings.Contains(got, `"B-OWN-GAME-KEY"`) || !strings.Contains(got, `"B-OWN-HOOK"`) || !strings.Contains(got, `"mode": "x"`) {
		t.Fatalf("target json secrets must stay and plain values change: %s", got)
	}
	if ctID(t, dbB, `SELECT referrer_prize_id::text FROM economy_referral_settings WHERE id=1`) != ctID(t, dbB, `SELECT id::text FROM economy_prize WHERE key='ct_prize1'`) {
		t.Fatal("referral prize not resolved")
	}
	bz, err := s.loadMethod(ctx, dbB, "bazaar")
	if err != nil {
		t.Fatal(err)
	}
	if !bz.Active || bz.Config.str("apiSecret") != "B-OWN-BAZAAR" || bz.Config.str("packageName") != "com.ct.game" {
		t.Fatalf("bazaar must keep B's secret and take the plain fields: %+v", bz)
	}
	var zp int
	_ = dbB.QueryRow(`SELECT count(*) FROM economy_payment_method WHERE method='zarinpal'`).Scan(&zp)
	if zp != 0 {
		t.Fatal("a method without keys must not be written")
	}

	fb, _ := ctExportJSON(t, dbB, "prod")
	for _, fa := range f.Areas {
		if fa.Key == "payment_methods" {
			continue
		}
		for _, ft := range fa.Tables {
			for _, r := range ft.Rows {
				if !strings.Contains(r.Key, "ct_") {
					continue
				}
				if st := ctStatus(res, ft.Table, r.Key); st != nil && st.Status == configtransfer.StatusMissingRefs {
					continue
				}
				other := ctRow(fb, fa.Key, ft.Table, r.Key)
				if other == nil {
					t.Fatalf("%s %s missing in B", ft.Table, r.Key)
				}
				a, _ := json.Marshal(r.Values)
				b, _ := json.Marshal(other.Values)
				if string(a) != string(b) {
					t.Fatalf("%s %s differs:\nA %s\nB %s", ft.Table, r.Key, a, b)
				}
			}
		}
	}

	again, err := configtransfer.Apply(ctx, dbB, f, configtransfer.Options{Commit: true, By: "test"})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range again.Areas {
		for _, r := range a.Rows {
			if r.Status == configtransfer.StatusSame || (r.Table == "economy_payment_method" && r.Key == "zarinpal" && r.Status == configtransfer.StatusSecretNeeded) {
				continue
			}
			if r.Status == configtransfer.StatusMissingRefs && r.Reason == configtransfer.ReasonSourceBroken {
				first := ctStatus(res, r.Table, r.Key)
				if first == nil || first.Status != r.Status || first.Reason != r.Reason || len(r.Missing) == 0 {
					t.Fatalf("a source-broken row must be stable and named: %+v first %+v", r, first)
				}
				continue
			}
			t.Fatalf("second import must be all same: %s %s %s %+v %+v %s", r.Table, r.Key, r.Status, r.Changes, r.Missing, r.Reason)
		}
	}

	off, err := configtransfer.Apply(ctx, dbB, f, configtransfer.Options{Areas: []string{"items"}, SwitchOff: map[string]bool{"items": true}, By: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if st := ctStatus(off, "economy_item", "ct_only_b"); st == nil || st.Status != configtransfer.StatusSwitchedOff {
		t.Fatalf("switch off: %+v", st)
	}
	for _, a := range off.Areas {
		if a.Key != "items" && a.Selected {
			t.Fatalf("area %s must not be selected", a.Key)
		}
	}

	ctExec(t, dbB, `DELETE FROM economy_item WHERE key='ct_car1'`)
	bad, _ := ctExportJSON(t, dbA, "dev")
	ctRow(bad, "items", "economy_item", "ct_car1").Values["type_id"] = "item_type/ct_nope"
	miss, err := configtransfer.Apply(ctx, dbB, bad, configtransfer.Options{Areas: []string{"items", "games"}, By: "test"})
	if err != nil {
		t.Fatal(err)
	}
	st := ctStatus(miss, "economy_item", "ct_car1")
	if st == nil || st.Status != configtransfer.StatusMissingRefs || len(st.Missing) != 1 || st.Missing[0].Key != "item_type/ct_nope" {
		t.Fatalf("missing ref: %+v", st)
	}
	if st := ctStatus(miss, "economy_game", "ct_game"); st == nil || st.Status != configtransfer.StatusMissingRefs {
		t.Fatalf("dependent row must miss too: %+v", st)
	}
	picked, err := configtransfer.Apply(ctx, dbB, f, configtransfer.Options{Areas: []string{"items"}, Rows: map[string][]string{"items": {}}, By: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if st := ctStatus(picked, "economy_item", "ct_car1"); st == nil || st.Status != configtransfer.StatusSkipped {
		t.Fatalf("unpicked row must be skipped: %+v", st)
	}
}
