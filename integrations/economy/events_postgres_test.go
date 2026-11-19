package main

import (
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

func newEventHarness(t *testing.T) *harness {
	h := newHarness(t)
	for _, q := range []string{
		`TRUNCATE economy_reward_claim, economy_user_streak, economy_streak_reward, economy_streak, economy_user_level, economy_level_step, economy_level_table,
			economy_meta_rule, economy_event_rule_progress, economy_event_rule, economy_player_event_day, economy_player_event_total, economy_player_event, economy_event_type CASCADE`,
		`UPDATE economy_event_settings SET timezone='Asia/Tehran', week_start=6, retention_days=30, player_rate_limit=120`,
		`INSERT INTO economy_event_type (id, key, name, is_system) VALUES (gen_random_uuid(), 'level_up', '{"fa":"سطح"}', true), (gen_random_uuid(), 'login', '{"fa":"ورود"}', true)`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'ev:%'`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for name, fn := range h.s.eventRPCs() {
		h.rpcs[name] = fn
	}
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "coin", "name": map[string]string{"fa": "سکه", "en": "Coin"}, "type": "soft"})
	return h
}

func (h *harness) coinPrize(key string, amount int, repeatable bool) string {
	h.t.Helper()
	p := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": key, "name": map[string]string{"fa": key, "en": key}, "isRepeatable": repeatable,
		"items": []map[string]any{{"type": "currency", "key": "coin", "action": "add", "value": amount}}})
	return p["id"].(string)
}

func (h *harness) eventType(key string, client bool, extra map[string]any) string {
	h.t.Helper()
	body := map[string]any{"key": key, "name": map[string]string{"fa": "رویداد " + key, "en": key}, "clientAllowed": client, "isActive": true, "rateLimit": 1000}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/event-types", body)["id"].(string)
}

func (h *harness) at(t time.Time) {
	h.s.now = func() time.Time { return t }
}

func (h *harness) record(user, key string, value float64) map[string]any {
	h.t.Helper()
	return h.mustRPC(user, "economy_event_record", map[string]any{"key": key, "value": value})
}

func (h *harness) count(q string, args ...any) int {
	h.t.Helper()
	var n int
	if err := h.s.db.QueryRowContext(h.ctx, q, args...).Scan(&n); err != nil {
		h.t.Fatal(err)
	}
	return n
}

func TestEventTypesValidationAndRecording(t *testing.T) {
	h := newEventHarness(t)
	if code, out := h.api(http.MethodPost, "/event-types", map[string]any{"key": "Bad Key", "name": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "INVALID_EVENT_KEY" {
		t.Fatalf("bad key = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/event-types", map[string]any{"key": "score", "name": map[string]string{"fa": "x"}, "valueMin": 10, "valueMax": 5}); code != 400 || out["error"] != "INVALID_RANGE" {
		t.Fatalf("range = %d %v", code, out)
	}
	h.eventType("round_finished", false, nil)
	star := h.eventType("golden_star", true, map[string]any{"valueMin": 0, "valueMax": 1, "rateLimit": 3})
	user := h.user()
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "nope"}); code != "EVENT_TYPE_NOT_FOUND" {
		t.Fatalf("unknown event = %s", code)
	}
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "round_finished", "value": 5}); code != "EVENT_NOT_ALLOWED" {
		t.Fatalf("server-only event from a client = %s", code)
	}
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "golden_star", "value": 2}); code != "VALUE_OUT_OF_RANGE" {
		t.Fatalf("value range = %s", code)
	}
	big := map[string]any{}
	for i := 0; i < 300; i++ {
		big[uuid.Must(uuid.NewV4()).String()] = i
	}
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "golden_star", "metadata": big}); code != "METADATA_TOO_LARGE" {
		t.Fatalf("metadata = %s", code)
	}
	for i := 0; i < 3; i++ {
		h.mustRPC(user, "economy_event_record", map[string]any{"key": "golden_star", "metadata": map[string]any{"round": i}})
	}
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "golden_star"}); code != "TOO_MANY_EVENTS" {
		t.Fatalf("rate limit per event = %s", code)
	}
	if n := h.count(`SELECT count FROM economy_player_event_total WHERE user_id=$1 AND type_id=$2`, user, star); n != 3 {
		t.Fatalf("total = %d", n)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_event_type SET is_active=false WHERE id=$1`, star); err != nil {
		t.Fatal(err)
	}
	if code := h.rpcCode(h.user(), "economy_event_record", map[string]any{"key": "golden_star"}); code != "EVENT_TYPE_NOT_FOUND" {
		t.Fatalf("inactive event = %s", code)
	}
	if code, out := h.api(http.MethodDelete, "/event-types/"+star, nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete with history = %d %v", code, out)
	}
	var system string
	_ = h.s.db.QueryRow(`SELECT id::text FROM economy_event_type WHERE key='login'`).Scan(&system)
	if code, out := h.api(http.MethodDelete, "/event-types/"+system, nil); code != 409 || out["error"] != "SYSTEM_EVENT" {
		t.Fatalf("delete system = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPut, "/event-types/"+system, map[string]any{"key": "login", "name": map[string]string{"fa": "ورود"}, "clientAllowed": true}); code != 400 || out["error"] != "SYSTEM_EVENT_SERVER_ONLY" {
		t.Fatalf("client login = %d %v", code, out)
	}
	admin := h.mustAPI(200, http.MethodPost, "/users/"+user+"/events", map[string]any{"key": "round_finished", "value": 40})
	if admin["key"] != "round_finished" {
		t.Fatalf("admin event = %v", admin)
	}
	log := h.mustAPI(200, http.MethodGet, "/event-log?userId="+user+"&source=admin", nil)
	if log["total"].(float64) != 1 {
		t.Fatalf("log = %v", log)
	}
}

func TestEventRulePeriodsAcrossTehranAndUTCMidnight(t *testing.T) {
	h := newEventHarness(t)
	typeID := h.eventType("round_finished", true, nil)
	prize := h.coinPrize("coins_20", 20, true)
	rule := h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "three_a_day", "name": map[string]string{"fa": "سه دور در روز"}, "eventTypeId": typeID,
		"measure": "count", "period": "day", "threshold": 3, "repeat": "period", "prizeId": prize, "grantMode": "auto", "isActive": true})
	if code, out := h.api(http.MethodPost, "/event-rules", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "eventTypeId": typeID, "measure": "count", "period": "total", "threshold": 3, "repeat": "period", "prizeId": prize, "grantMode": "auto"}); code != 400 || out["error"] != "REPEAT_NEEDS_PERIOD" {
		t.Fatalf("period on total = %d %v", code, out)
	}
	user := h.user()
	h.at(time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC))
	h.record(user, "round_finished", 1)
	h.record(user, "round_finished", 1)
	h.at(time.Date(2026, 11, 11, 20, 31, 0, 0, time.UTC))
	out := h.record(user, "round_finished", 1)
	if len(out["claims"].([]any)) != 0 || h.wallet(user)["coin"] != 0 {
		t.Fatalf("the third round fell on the next Tehran day: %v", out)
	}
	h.record(user, "round_finished", 1)
	out = h.record(user, "round_finished", 1)
	claims := out["claims"].([]any)
	if len(claims) != 1 || claims[0].(map[string]any)["status"] != "granted" || h.wallet(user)["coin"] != 20 {
		t.Fatalf("three rounds on the 12th earn the prize once: %v wallet %v", out, h.wallet(user))
	}
	h.record(user, "round_finished", 1)
	if h.wallet(user)["coin"] != 20 {
		t.Fatal("a fourth round the same day earns nothing")
	}
	if n := h.count(`SELECT count(*) FROM economy_event_rule_progress WHERE rule_id=$1 AND user_id=$2`, rule["id"], user); n != 2 {
		t.Fatalf("two Tehran days = %d rows", n)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_event_settings SET timezone='UTC'`); err != nil {
		t.Fatal(err)
	}
	other := h.user()
	h.at(time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC))
	h.record(other, "round_finished", 1)
	h.record(other, "round_finished", 1)
	h.at(time.Date(2026, 11, 11, 20, 31, 0, 0, time.UTC))
	h.record(other, "round_finished", 1)
	if h.wallet(other)["coin"] != 20 {
		t.Fatal("in UTC the three rounds are on the same day")
	}
	progress := h.mustRPC(other, "economy_events_progress", map[string]any{"lang": "en"})
	item := progress["items"].([]any)[0].(map[string]any)
	if item["current"].(float64) != 3 || item["done"] != true || item["claims"].(float64) != 1 || item["periodEndsAt"] == nil {
		t.Fatalf("progress = %v", item)
	}
	h.at(time.Date(2026, 11, 12, 0, 5, 0, 0, time.UTC))
	item = h.mustRPC(other, "economy_events_progress", map[string]any{})["items"].([]any)[0].(map[string]any)
	if item["current"].(float64) != 0 || item["done"] != false {
		t.Fatalf("a new UTC day starts from zero: %v", item)
	}
	if code, out := h.api(http.MethodDelete, "/event-rules/"+rule["id"].(string), nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete rule with claims = %d %v", code, out)
	}
}

func TestEventRuleEveryOnceClaimModeAndAudience(t *testing.T) {
	h := newEventHarness(t)
	typeID := h.eventType("stars", true, nil)
	prize := h.coinPrize("coins_5", 5, true)
	h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "every_100", "name": map[string]string{"fa": "هر ۱۰۰"}, "eventTypeId": typeID,
		"measure": "sum", "period": "total", "threshold": 100, "repeat": "every", "maxClaims": 3, "prizeId": prize, "grantMode": "auto", "isActive": true})
	record := h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "best_80", "name": map[string]string{"fa": "رکورد"}, "eventTypeId": typeID,
		"measure": "max", "period": "week", "threshold": 80, "repeat": "once", "prizeId": h.coinPrize("coins_50", 50, true), "grantMode": "claim", "isActive": true})
	user := h.user()
	h.at(time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	h.record(user, "stars", 90)
	h.record(user, "stars", 250)
	if w := h.wallet(user)["coin"]; w != 15 {
		t.Fatalf("340 stars cross 100, 200 and 300: wallet %d", w)
	}
	h.record(user, "stars", 500)
	if w := h.wallet(user)["coin"]; w != 15 {
		t.Fatalf("max claims stops at three: wallet %d", w)
	}
	ready := h.mustRPC(user, "economy_rewards_mine", map[string]any{"status": "ready"})["items"].([]any)
	if len(ready) != 1 || ready[0].(map[string]any)["source"] != "best_80" {
		t.Fatalf("one claimable record reward: %v", ready)
	}
	claimID := ready[0].(map[string]any)["claimId"].(string)
	if code := h.rpcCode(h.user(), "economy_reward_claim", map[string]any{"claimId": claimID}); code != "CLAIM_NOT_FOUND" {
		t.Fatalf("someone else's claim = %s", code)
	}
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = h.rpc(user, "economy_reward_claim", map[string]any{"claimId": claimID})
		}()
	}
	wg.Wait()
	if w := h.wallet(user)["coin"]; w != 65 {
		t.Fatalf("six parallel claims pay once: wallet %d", w)
	}
	if code := h.rpcCode(user, "economy_reward_claim", map[string]any{"claimId": claimID}); code != "ALREADY_CLAIMED" {
		t.Fatalf("second claim = %s", code)
	}
	h.at(time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC))
	h.record(user, "stars", 95)
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE ref_id=$1`, record["id"]); n != 1 {
		t.Fatalf("once means once across weeks: %d", n)
	}

	tag := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})
	h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "vip_only", "name": map[string]string{"fa": "ویژه"}, "eventTypeId": typeID,
		"measure": "count", "period": "total", "threshold": 1, "repeat": "once", "prizeId": prize, "grantMode": "claim", "audienceTagIds": []string{tag["id"].(string)}, "isActive": true})
	plain, vip := h.user(), h.user()
	if _, err := h.s.db.Exec(`INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, vip, tag["id"]); err != nil {
		t.Fatal(err)
	}
	h.record(plain, "stars", 1)
	h.record(vip, "stars", 1)
	if n := h.count(`SELECT count(*) FROM economy_reward_claim c JOIN economy_event_rule r ON r.id=c.owner_id WHERE r.key='vip_only'`); n != 1 {
		t.Fatalf("only the tagged player qualifies: %d", n)
	}
	items := h.mustRPC(plain, "economy_events_progress", map[string]any{})["items"].([]any)
	for _, it := range items {
		if it.(map[string]any)["key"] == "vip_only" {
			t.Fatal("hidden rule shown to an untagged player")
		}
	}
	h.at(time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC))
	if code, out := h.api(http.MethodPost, "/event-rules", map[string]any{"key": "window", "name": map[string]string{"fa": "x"}, "eventTypeId": typeID, "measure": "count", "period": "total", "threshold": 1, "repeat": "once", "prizeId": prize, "grantMode": "auto",
		"startsAt": "2026-12-01T00:00:00Z", "endsAt": "2026-11-01T00:00:00Z"}); code != 400 || out["error"] != "INVALID_WINDOW" {
		t.Fatalf("window = %d %v", code, out)
	}
	h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "december", "name": map[string]string{"fa": "دی"}, "eventTypeId": typeID, "measure": "count", "period": "total", "threshold": 1, "repeat": "once", "prizeId": prize, "grantMode": "claim",
		"startsAt": "2026-12-01T00:00:00Z", "isActive": true})
	h.record(plain, "stars", 1)
	if n := h.count(`SELECT count(*) FROM economy_reward_claim c JOIN economy_event_rule r ON r.id=c.owner_id WHERE r.key='december'`); n != 0 {
		t.Fatal("a rule before its start date does not count")
	}
}

func TestEventGrantsAreIdempotentOnManyNodes(t *testing.T) {
	h := newEventHarness(t)
	typeID := h.eventType("tick", true, nil)
	prize := h.coinPrize("coins_3", 3, true)
	h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "every_5", "name": map[string]string{"fa": "هر پنج"}, "eventTypeId": typeID,
		"measure": "count", "period": "total", "threshold": 5, "repeat": "every", "prizeId": prize, "grantMode": "auto", "isActive": true})
	user := h.user()
	nodes := make([]*service, 4)
	for i := range nodes {
		n := testService()
		n.db, n.nk, n.store = h.s.db, h.nk, h.s.store
		nodes[i] = n
	}
	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := 1.0
			_, err := nodes[i%len(nodes)].recordEvent(h.ctx, eventInput{UserID: user, Key: "tick", Value: &v, Source: "server"})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if n := h.count(`SELECT count FROM economy_player_event_total WHERE user_id=$1`, user); n != 40 {
		t.Fatalf("total = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND status='granted'`, user); n != 8 {
		t.Fatalf("40 ticks are 8 claims, got %d", n)
	}
	if w := h.wallet(user)["coin"]; w != 24 {
		t.Fatalf("8 claims pay 24 coins exactly once each, got %d", w)
	}
	var claimID string
	_ = h.s.db.QueryRow(`SELECT id::text FROM economy_reward_claim WHERE user_id=$1 LIMIT 1`, user).Scan(&claimID)
	for _, n := range nodes {
		_, _ = n.grantClaim(h.ctx, claimID, 1)
	}
	if w := h.wallet(user)["coin"]; w != 24 {
		t.Fatalf("granting a granted claim again changed the wallet: %d", w)
	}

	stuck := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.Exec(`INSERT INTO economy_reward_claim (id, user_id, kind, owner_id, ref_id, scope, prize_id, grant_mode, created_at) VALUES ($1,$2,'event',$3,$3,'crash',$4,'auto',now() - interval '5 minutes')`,
		stuck, user, uuid.Must(uuid.NewV4()).String(), prize); err != nil {
		t.Fatal(err)
	}
	var done int
	var mu sync.Mutex
	for _, n := range nodes {
		wg.Add(1)
		go func(n *service) {
			defer wg.Done()
			c := n.retryAutoClaims(h.ctx, 5)
			mu.Lock()
			done += c
			mu.Unlock()
		}(n)
	}
	wg.Wait()
	if done != 1 || h.wallet(user)["coin"] != 27 {
		t.Fatalf("a claim left by a crashed node is granted once by one node: done=%d wallet=%d", done, h.wallet(user)["coin"])
	}
	if _, err := h.s.db.Exec(`UPDATE economy_player_event SET created_at = now() - interval '40 days' WHERE user_id=$1`, user); err != nil {
		t.Fatal(err)
	}
	removed, err := h.s.sweepEvents(h.ctx)
	if err != nil || removed != 40 {
		t.Fatalf("sweep removed %d: %v", removed, err)
	}
	if n := h.count(`SELECT count FROM economy_player_event_total WHERE user_id=$1`, user); n != 40 {
		t.Fatal("the sweep keeps the totals")
	}
}

func TestMetaRulesAndLevelTables(t *testing.T) {
	h := newEventHarness(t)
	typeID := h.eventType("round_finished", false, nil)
	xp := h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "xp", "name": map[string]string{"fa": "تجربه"}})
	level := h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "level", "name": map[string]string{"fa": "سطح"}, "defaultValue": 1})
	h.mustAPI(201, http.MethodPost, "/meta-rules", map[string]any{"key": "xp_per_score", "name": map[string]string{"fa": "امتیاز تجربه"}, "eventTypeId": typeID, "metaId": xp["id"],
		"action": "add", "valueSource": "event", "amount": 2, "isActive": true})
	h.mustAPI(201, http.MethodPost, "/meta-rules", map[string]any{"key": "best", "name": map[string]string{"fa": "رکورد"}, "eventTypeId": typeID,
		"metaId": h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "best", "name": map[string]string{"fa": "رکورد"}})["id"], "action": "max", "valueSource": "event", "amount": 1, "isActive": true})
	auto := h.coinPrize("level_2", 20, true)
	claim := h.coinPrize("level_3", 50, true)
	bonusXP := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "xp_boost", "name": map[string]string{"fa": "تجربهٔ اضافه"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "meta", "key": "xp", "action": "add", "value": 300}}})["id"].(string)
	if code, out := h.api(http.MethodPost, "/level-tables", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "metaId": xp["id"], "steps": []map[string]any{{"level": 2, "threshold": 100}, {"level": 3, "threshold": 50}}}); code != 400 || out["error"] != "THRESHOLDS_NOT_ORDERED" {
		t.Fatalf("order = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/level-tables", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "metaId": xp["id"], "levelMetaId": xp["id"], "steps": []map[string]any{{"level": 1, "threshold": 0}}}); code != 400 || out["error"] != "SAME_META" {
		t.Fatalf("same meta = %d %v", code, out)
	}
	table := h.mustAPI(201, http.MethodPost, "/level-tables", map[string]any{"key": "levels", "name": map[string]string{"fa": "سطح‌ها"}, "metaId": xp["id"], "levelMetaId": level["id"], "isActive": true,
		"steps": []map[string]any{
			{"level": 1, "threshold": 0, "name": map[string]string{"fa": "تازه‌کار"}},
			{"level": 2, "threshold": 100, "prizeId": auto, "grantMode": "auto", "name": map[string]string{"fa": "ستاره‌چین"}},
			{"level": 3, "threshold": 300, "prizeId": claim, "grantMode": "claim"},
			{"level": 4, "threshold": 600, "prizeId": bonusXP, "grantMode": "auto"},
			{"level": 5, "threshold": 900},
		}})
	if len(table["steps"].([]any)) != 5 {
		t.Fatalf("steps = %v", table["steps"])
	}
	user := h.user()
	ups := 0
	h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "level_watch", "name": map[string]string{"fa": "سطح"}, "eventTypeId": h.typeIDByKey("level_up"),
		"measure": "max", "period": "total", "threshold": 3, "repeat": "once", "prizeId": h.coinPrize("level_up_3", 7, true), "grantMode": "auto", "isActive": true})
	out, err := h.s.recordEvent(h.ctx, eventInput{UserID: user, Key: "round_finished", Value: ptr(60), Source: "server"})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range out.Levels {
		if l.Key == "levels" && l.To == 2 {
			ups++
		}
	}
	if h.wallet(user)["coin"] != 20 || out.Metas["xp"] != 120 || ups != 1 {
		t.Fatalf("120 xp reaches level 2 with its auto prize: %+v wallet %v", out, h.wallet(user))
	}
	if n := h.count(`SELECT value::int FROM economy_user_meta m JOIN economy_meta x ON x.id=m.meta_id WHERE x.key='level' AND user_id=$1`, user); n != 2 {
		t.Fatalf("level meta = %d", n)
	}
	if p, err := h.s.levelProgressHook(h.ctx, user, "levels"); err != nil || p.Level != 2 || p.Value != 120 || p.LevelStart != 100 || p.Next == nil || *p.Next != 300 || p.MaxLevel != 5 {
		t.Fatalf("progress at level 2 = %+v %v", p, err)
	}
	if p, err := h.s.levelProgressHook(h.ctx, h.user(), "levels"); err != nil || p.Level != 1 || p.Value != 0 || p.Next == nil || *p.Next != 100 {
		t.Fatalf("a new player is on the zero-threshold level = %+v %v", p, err)
	}
	if _, err := h.s.levelProgressHook(h.ctx, user, "missing"); !errors.Is(err, economyhooks.ErrNoLevelTable) {
		t.Fatalf("unknown table = %v", err)
	}
	if _, err := h.s.recordEvent(h.ctx, eventInput{UserID: user, Key: "round_finished", Value: ptr(40), Source: "server"}); err != nil {
		t.Fatal(err)
	}
	if n := h.count(`SELECT value::int FROM economy_user_meta m JOIN economy_meta x ON x.id=m.meta_id WHERE x.key='best' AND user_id=$1`, user); n != 60 {
		t.Fatalf("max meta keeps the best = %d", n)
	}
	out, _ = h.s.recordEvent(h.ctx, eventInput{UserID: user, Key: "round_finished", Value: ptr(100), Source: "server"})
	if len(out.Levels) == 0 || out.Levels[0].To != 3 {
		t.Fatalf("400 xp is level 3: %+v", out.Levels)
	}
	if h.wallet(user)["coin"] != 27 {
		t.Fatalf("level 3 prize waits for a claim, level_up rule paid 7: %v", h.wallet(user))
	}
	levels := h.mustRPC(user, "economy_levels", map[string]any{"lang": "fa"})["items"].([]any)[0].(map[string]any)
	steps := levels["steps"].([]any)
	if levels["level"].(float64) != 3 || steps[1].(map[string]any)["status"] != "claimed" || steps[2].(map[string]any)["status"] != "claimable" || steps[3].(map[string]any)["status"] != "locked" || levels["next"].(map[string]any)["level"].(float64) != 4 {
		t.Fatalf("level view = %v", levels)
	}
	res := h.mustRPC(user, "economy_reward_claim", map[string]any{"claimId": steps[2].(map[string]any)["claimId"]})
	if h.wallet(user)["coin"] != 77 || len(res["rewards"].([]any)) != 1 {
		t.Fatalf("claimed level 3: %v %v", res, h.wallet(user))
	}
	out, _ = h.s.recordEvent(h.ctx, eventInput{UserID: user, Key: "round_finished", Value: ptr(150), Source: "server"})
	if n := h.count(`SELECT level FROM economy_user_level WHERE user_id=$1`, user); n != 5 {
		t.Fatalf("700 xp is level 4 whose prize adds 300 xp, which cascades to level 5: got %d (%+v)", n, out.Levels)
	}
	if p, err := h.s.levelProgressHook(h.ctx, user, "levels"); err != nil || p.Level != 5 || p.Next != nil || p.LevelStart != 900 {
		t.Fatalf("progress at the last level = %+v %v", p, err)
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE kind='level' AND user_id=$1`, user); n != 3 {
		t.Fatalf("level claims = %d", n)
	}
	if code, out := h.api(http.MethodDelete, "/level-tables/"+table["id"].(string), nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete used table = %d %v", code, out)
	}
	progress := h.mustAPI(200, http.MethodGet, "/users/"+user+"/progress?lang=fa", nil)
	if len(progress["levels"].([]any)) != 1 || len(progress["claims"].([]any)) == 0 || len(progress["totals"].([]any)) < 2 {
		t.Fatalf("player progress = %v", progress)
	}
}

func ptr(v float64) *float64 { return &v }

func (h *harness) typeIDByKey(key string) string {
	var id string
	if err := h.s.db.QueryRow(`SELECT id::text FROM economy_event_type WHERE key=$1`, key).Scan(&id); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func TestStreakGraceResetCycleAndClaims(t *testing.T) {
	h := newEventHarness(t)
	login := h.typeIDByKey("login")
	day := map[int]string{}
	for d := 1; d <= 3; d++ {
		day[d] = h.coinPrize("streak_"+itoa(d), d*10, true)
	}
	if code, out := h.api(http.MethodPost, "/streaks", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "eventTypeId": login, "cycleLength": 3, "rewards": []map[string]any{{"day": 4, "prizeId": day[1]}}}); code != 400 || out["error"] != "DAY_OUTSIDE_CYCLE" {
		t.Fatalf("day outside cycle = %d %v", code, out)
	}
	streak := h.mustAPI(201, http.MethodPost, "/streaks", map[string]any{"key": "daily_login", "name": map[string]string{"fa": "ورود روزانه"}, "eventTypeId": login,
		"periodDays": 1, "gracePeriods": 1, "missAction": "reset", "cycleLength": 3, "isActive": true,
		"rewards": []map[string]any{{"day": 1, "prizeId": day[1], "grantMode": "auto"}, {"day": 2, "prizeId": day[2]}, {"day": 3, "prizeId": day[3]}}})
	user := h.user()
	login1 := func(t0 time.Time) {
		h.at(t0)
		h.s.recordServerEvent(h.ctx, user, "login", 1, nil)
	}
	base := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)
	login1(base)
	login1(base.Add(3 * time.Hour))
	view := func() map[string]any {
		return h.mustRPC(user, "economy_streaks", map[string]any{})["items"].([]any)[0].(map[string]any)
	}
	v := view()
	if v["count"].(float64) != 1 || v["countedThisPeriod"] != true || h.wallet(user)["coin"] != 10 {
		t.Fatalf("day one counts once and pays its auto reward: %v %v", v, h.wallet(user))
	}
	login1(base.Add(24 * time.Hour))
	v = view()
	days := v["days"].([]any)
	if v["count"].(float64) != 2 || days[1].(map[string]any)["status"] != "claimable" || days[2].(map[string]any)["status"] != "locked" {
		t.Fatalf("day two = %v", v)
	}
	login1(base.Add(72 * time.Hour))
	v = view()
	if v["count"].(float64) != 3 || v["graceLeft"].(float64) != 0 {
		t.Fatalf("one missed day uses the grace: %v", v)
	}
	ready := v["claimable"].([]any)
	if len(ready) != 2 {
		t.Fatalf("days two and three wait for a claim: %v", ready)
	}
	for _, id := range ready {
		h.mustRPC(user, "economy_reward_claim", map[string]any{"claimId": id})
	}
	if h.wallet(user)["coin"] != 60 {
		t.Fatalf("10+20+30 = %d", h.wallet(user)["coin"])
	}
	login1(base.Add(96 * time.Hour))
	if v := view(); v["count"].(float64) != 4 || v["days"].([]any)[0].(map[string]any)["status"] != "claimed" || h.wallet(user)["coin"] != 70 {
		t.Fatalf("day four starts the cycle again at reward day one: %v %v", v, h.wallet(user))
	}
	h.at(base.Add(144 * time.Hour))
	if v := view(); v["broken"] != true || v["count"].(float64) != 0 {
		t.Fatalf("a missed day with no grace left breaks the streak: %v", v)
	}
	login1(base.Add(144 * time.Hour))
	if v := view(); v["count"].(float64) != 1 || v["best"].(float64) != 4 || v["run"].(float64) != 2 || h.wallet(user)["coin"] != 80 {
		t.Fatalf("after the reset day one pays again in the new run: %v %v", v, h.wallet(user))
	}
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/streaks/"+streak["id"].(string)+"/reset", nil)
	if v := view(); v["count"].(float64) != 0 || v["run"].(float64) != 3 {
		t.Fatalf("admin reset = %v", v)
	}
	if code, out := h.api(http.MethodDelete, "/streaks/"+streak["id"].(string), nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete used streak = %d %v", code, out)
	}
	var wg sync.WaitGroup
	fresh := h.user()
	h.at(base.Add(200 * time.Hour))
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.s.recordServerEvent(h.ctx, fresh, "login", 1, nil)
		}()
	}
	wg.Wait()
	if n := h.count(`SELECT count FROM economy_user_streak WHERE user_id=$1`, fresh); n != 1 || h.wallet(fresh)["coin"] != 10 {
		t.Fatalf("ten parallel logins on one day count once: count %d wallet %d", n, h.wallet(fresh)["coin"])
	}
}

func (h *harness) milestoneStreak(key, eventTypeID, miss string, grace int, repay bool, rewardDays int, milestones []map[string]any) map[string]any {
	h.t.Helper()
	rewards := []map[string]any{}
	for d := 1; d <= rewardDays; d++ {
		rewards = append(rewards, map[string]any{"day": d, "prizeId": h.coinPrize(key+"_d"+itoa(d), d, true), "grantMode": "auto"})
	}
	return h.mustAPI(201, http.MethodPost, "/streaks", map[string]any{"key": key, "name": map[string]string{"fa": key}, "eventTypeId": eventTypeID,
		"periodDays": 1, "gracePeriods": grace, "missAction": miss, "isActive": true, "repayAfterFall": repay, "rewards": rewards, "milestones": milestones})
}

func streakItem(out map[string]any, key string) map[string]any {
	for _, raw := range out["streaks"].([]any) {
		if it := raw.(map[string]any); it["key"] == key {
			return it
		}
	}
	return nil
}

func (h *harness) streakView(user, key string) map[string]any {
	h.t.Helper()
	for _, raw := range h.mustRPC(user, "economy_streaks", map[string]any{})["items"].([]any) {
		if it := raw.(map[string]any); it["key"] == key {
			return it
		}
	}
	h.t.Fatalf("streak %s not listed", key)
	return nil
}

func TestStreakMilestonesFloorRepayAndPrize(t *testing.T) {
	h := newEventHarness(t)
	h.eventType("visit", false, nil)
	visit := h.typeIDByKey("visit")
	bonus3 := h.coinPrize("ms_bonus_3", 1000, true)
	bonus5 := h.coinPrize("ms_bonus_5", 5000, true)
	if code, out := h.api(http.MethodPost, "/streaks", map[string]any{"key": "dup", "name": map[string]string{"fa": "x"}, "eventTypeId": visit,
		"milestones": []map[string]any{{"day": 3}, {"day": 3}}}); code != 400 || out["error"] != "DUPLICATE_MILESTONE" {
		t.Fatalf("duplicate milestone = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/streaks", map[string]any{"key": "far", "name": map[string]string{"fa": "x"}, "eventTypeId": visit,
		"milestones": []map[string]any{{"day": 4000}}}); code != 400 || out["error"] != "INVALID_MILESTONE_DAY" {
		t.Fatalf("milestone day range = %d %v", code, out)
	}
	sk := h.milestoneStreak("safe_reset", visit, "reset", 0, false, 8, []map[string]any{{"day": 3, "prizeId": bonus3, "grantMode": "auto"}, {"day": 5, "prizeId": bonus5}, {"day": 10}})
	if ms := sk["milestones"].([]any); len(ms) != 3 || sk["repayAfterFall"] != false || ms[2].(map[string]any)["prizeId"] != nil {
		t.Fatalf("saved milestones = %v", sk)
	}
	user := h.user()
	base := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)
	visitOn := func(user string, day int) map[string]any {
		h.at(base.Add(time.Duration(day) * 24 * time.Hour))
		return h.mustAPI(200, http.MethodPost, "/users/"+user+"/events", map[string]any{"key": "visit", "value": 1})
	}
	for d := 0; d < 6; d++ {
		visitOn(user, d)
	}
	v := h.streakView(user, "safe_reset")
	if v["count"].(float64) != 6 || v["floor"].(float64) != 5 || v["nextMilestone"].(float64) != 10 || h.wallet(user)["coin"] != 21+1000 {
		t.Fatalf("six days: floor 5, next 10, days 1..6 paid and the auto milestone prize: %v %v", v, h.wallet(user))
	}
	ms := v["milestones"].([]any)
	if ms[0].(map[string]any)["status"] != "claimed" || ms[1].(map[string]any)["status"] != "claimable" || ms[2].(map[string]any)["status"] != "locked" || ms[1].(map[string]any)["safe"] != true {
		t.Fatalf("milestone states = %v", ms)
	}
	h.mustRPC(user, "economy_reward_claim", map[string]any{"claimId": ms[1].(map[string]any)["claimId"]})
	h.at(base.Add(9 * 24 * time.Hour))
	v = h.streakView(user, "safe_reset")
	if v["broken"] != true || v["floored"] != true || v["count"].(float64) != 5 || v["run"].(float64) != 1 {
		t.Fatalf("a broken streak shows the milestone, not zero: %v", v)
	}
	if days := v["days"].([]any); days[5].(map[string]any)["status"] != "claimed" {
		t.Fatalf("day six stays claimed after the fall: %v", days[5])
	}
	out := visitOn(user, 9)
	st := streakItem(out, "safe_reset")
	if st["count"].(float64) != 6 || st["floored"] != true || st["reset"] != false || st["floor"].(float64) != 5 {
		t.Fatalf("the next counted day after the fall is the milestone plus one: %v", st)
	}
	if h.wallet(user)["coin"] != 21+1000+5000 {
		t.Fatalf("day six is not paid twice with re-pay off: %d", h.wallet(user)["coin"])
	}
	visitOn(user, 10)
	if h.wallet(user)["coin"] != 21+1000+5000+7 {
		t.Fatalf("day seven was never paid, so it pays: %d", h.wallet(user)["coin"])
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND scope LIKE 'm%'`, user); n != 2 {
		t.Fatalf("milestone prizes are claimed once: %d", n)
	}

	h.mustAPI(200, http.MethodPost, "/users/"+user+"/streaks/"+sk["id"].(string)+"/reset", nil)
	if v := h.streakView(user, "safe_reset"); v["count"].(float64) != 0 || v["floor"].(float64) != 0 || v["peak"].(float64) != 0 || v["best"].(float64) != 7 {
		t.Fatalf("console reset clears the floor and keeps best: %v", v)
	}
	before := h.wallet(user)["coin"]
	for d := 20; d < 26; d++ {
		visitOn(user, d)
	}
	if got := h.wallet(user)["coin"] - before; got != 21 {
		t.Fatalf("a new run pays days again but never the milestone prizes: %d", got)
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND scope LIKE 'm%'`, user); n != 2 {
		t.Fatalf("milestone prizes stay once ever: %d", n)
	}

	h.eventType("visit_repay", false, nil)
	repaySk := h.milestoneStreak("safe_repay", h.typeIDByKey("visit_repay"), "reset", 0, true, 8, []map[string]any{{"day": 3}})
	other := h.user()
	visitRepay := func(day int) map[string]any {
		h.at(base.Add(time.Duration(day) * 24 * time.Hour))
		return h.mustAPI(200, http.MethodPost, "/users/"+other+"/events", map[string]any{"key": "visit_repay", "value": 1})
	}
	for d := 0; d < 5; d++ {
		visitRepay(d)
	}
	visitRepay(8)
	if c := h.wallet(other)["coin"]; c != 15+4 {
		t.Fatalf("with re-pay on, day four pays again after the fall to three: %d", c)
	}
	visitRepay(10)
	visitRepay(11)
	if c := h.wallet(other)["coin"]; c != 15+4+4+5 {
		t.Fatalf("each fall starts a new chain of day rewards: %d", c)
	}
	if v := h.streakView(other, "safe_repay"); v["repayAfterFall"] != true || v["days"].([]any)[3].(map[string]any)["status"] != "claimed" {
		t.Fatalf("re-pay view = %v", v)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_streak SET repay_after_fall=false WHERE id=$1`, repaySk["id"]); err != nil {
		t.Fatal(err)
	}
	visitRepay(14)
	if c := h.wallet(other)["coin"]; c != 15+4+4+5 {
		t.Fatalf("turning re-pay off stops paying days already paid: %d", c)
	}

	h.eventType("visit_back", false, nil)
	h.milestoneStreak("safe_back", h.typeIDByKey("visit_back"), "step_back", 1, false, 0, []map[string]any{{"day": 4}})
	third := h.user()
	visitBack := func(day int) map[string]any {
		h.at(base.Add(time.Duration(day) * 24 * time.Hour))
		return streakItem(h.mustAPI(200, http.MethodPost, "/users/"+third+"/events", map[string]any{"key": "visit_back", "value": 1}), "safe_back")
	}
	for d := 0; d < 6; d++ {
		visitBack(d)
	}
	if st := visitBack(8); st["count"].(float64) != 6 || st["floored"] != false {
		t.Fatalf("two missed days with one grace day: one step back, then today: %v", st)
	}
	if st := visitBack(15); st["count"].(float64) != 5 || st["floored"] != true {
		t.Fatalf("stepping back stops at milestone four: %v", st)
	}

	fresh := h.user()
	for d := 0; d < 5; d++ {
		visitOn(fresh, d)
	}
	h.at(base.Add(8 * 24 * time.Hour))
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.s.recordServerEvent(h.ctx, fresh, "visit", 1, nil)
		}()
	}
	wg.Wait()
	if n := h.count(`SELECT count FROM economy_user_streak u JOIN economy_streak s ON s.id=u.streak_id WHERE u.user_id=$1 AND s.key='safe_reset'`, fresh); n != 6 {
		t.Fatalf("ten parallel events after a fall count once from the floor: %d", n)
	}
	if n := h.count(`SELECT falls FROM economy_user_streak u JOIN economy_streak s ON s.id=u.streak_id WHERE u.user_id=$1 AND s.key='safe_reset'`, fresh); n != 1 {
		t.Fatalf("one fall recorded: %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1`, fresh); n != 8 {
		t.Fatalf("six days and two milestone prizes, nothing doubled: %d", n)
	}
}

func TestStreakWithoutMilestonesKeepsOldScopes(t *testing.T) {
	h := newEventHarness(t)
	h.eventType("plain", false, nil)
	sk := h.milestoneStreak("plain_sk", h.typeIDByKey("plain"), "step_back", 0, false, 4, nil)
	user := h.user()
	base := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)
	for _, d := range []int{0, 1, 2, 3, 6} {
		h.at(base.Add(time.Duration(d) * 24 * time.Hour))
		h.s.recordServerEvent(h.ctx, user, "plain", 1, nil)
	}
	v := h.streakView(user, "plain_sk")
	if v["count"].(float64) != 3 || v["floor"].(float64) != 0 || v["nextMilestone"] != nil || len(v["milestones"].([]any)) != 0 || v["repayAfterFall"] != false {
		t.Fatalf("step back without milestones = %v", v)
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND scope !~ '^r1:[0-9]+$'`, user); n != 0 {
		t.Fatal("claims keep the r<run>:<day> scope")
	}
	if c := h.wallet(user)["coin"]; c != 10 {
		t.Fatalf("days stepped back to are not paid twice: %d", c)
	}
	if got := h.mustAPI(200, http.MethodGet, "/streaks/"+sk["id"].(string), nil); len(got["milestones"].([]any)) != 0 {
		t.Fatalf("admin view = %v", got)
	}
}
