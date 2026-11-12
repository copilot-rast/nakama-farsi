package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

func newChanceHarness(t *testing.T) *harness {
	h := newEventHarness(t)
	for _, q := range []string{
		`TRUNCATE economy_chance_play, economy_chance_state, economy_chance_game, economy_prize_group_member, economy_prize_group CASCADE`,
		`INSERT INTO economy_event_type (id, key, name, is_system, value_min, rate_limit) VALUES (gen_random_uuid(), 'ChanceGame:Played', '{"fa":"بازی شانس"}', true, 0, 1000)`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for name, fn := range h.s.chanceRPCs() {
		h.rpcs[name] = fn
	}
	return h
}

func (h *harness) chanceGame(body map[string]any) map[string]any {
	h.t.Helper()
	base := map[string]any{"name": map[string]string{"fa": "شانس " + body["key"].(string), "en": body["key"].(string)}, "isActive": true, "prizeType": "prize"}
	for k, v := range body {
		base[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/chance-games", base)
}

func (h *harness) currencyID(key string) string {
	h.t.Helper()
	var id string
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT id::text FROM economy_currency WHERE key=$1`, key).Scan(&id); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func row(chance float64, prize string) map[string]any {
	r := map[string]any{"chance": chance}
	if prize != "" {
		r["prizeId"] = prize
	}
	return r
}

func TestChancePickIsCumulative(t *testing.T) {
	rows := []chanceRow{{Chance: 30}, {Chance: 40}, {Chance: 30}}
	for roll, want := range map[int]int{0: 0, 25 * chanceScale: 0, 30*chanceScale - 1: 0, 30 * chanceScale: 1, 50 * chanceScale: 1, 85 * chanceScale: 2, chanceTotal - 1: 2} {
		if got := pickSlot(rows, roll); got != want {
			t.Fatalf("roll %d = slot %d, want %d", roll, got, want)
		}
	}
	zero := []chanceRow{{Chance: 0}, {Chance: 100}, {Chance: 0}}
	for _, roll := range []int{0, 1, chanceTotal - 1} {
		if got := pickSlot(zero, roll); got != 1 {
			t.Fatalf("a zero chance row was picked: roll %d slot %d", roll, got)
		}
	}
	seen := map[int]int{}
	for i := 0; i < 2000; i++ {
		v, err := secureInt(chanceTotal)
		if err != nil || v < 0 || v >= chanceTotal {
			t.Fatalf("secure roll %d %v", v, err)
		}
		seen[pickSlot(rows, v)]++
	}
	if seen[0] < 400 || seen[1] < 600 || seen[2] < 400 {
		t.Fatalf("roll spread looks wrong: %v", seen)
	}
	if _, err := drawGroupMember(nil); err == nil {
		t.Fatal("an empty group must not pay")
	}
}

func TestChanceGameValidation(t *testing.T) {
	h := newChanceHarness(t)
	p10 := h.coinPrize("c10", 10, true)
	bad := func(body map[string]any, code string, field string) map[string]any {
		t.Helper()
		base := map[string]any{"key": "bad", "name": map[string]string{"fa": "بد"}, "intervalTime": 60, "prize": []any{row(100, p10)}}
		for k, v := range body {
			base[k] = v
		}
		c, out := h.api(http.MethodPost, "/chance-games", base)
		if c != 400 || out["error"] != code || (field != "" && out["field"] != field) {
			t.Fatalf("%v = %d %v, want %s %s", body, c, out, code, field)
		}
		return out
	}
	out := bad(map[string]any{"prize": []any{row(30, p10), row(40, ""), row(20, "")}}, "TOTAL_CHANCE_MUST_BE_100", "prize")
	if meta, _ := out["meta"].(map[string]any); meta == nil || meta["totalChance"].(float64) != 90 {
		t.Fatalf("total meta = %v", out)
	}
	bad(map[string]any{"prize": []any{row(33.3333, p10), row(33.3333, ""), row(33.3333, "")}}, "TOTAL_CHANCE_MUST_BE_100", "prize")
	bad(map[string]any{"prize": []any{}}, "PRIZE_LIST_REQUIRED", "prize")
	bad(map[string]any{"prize": []any{row(101, p10)}}, "INVALID_PRIZE_CHANCE", "prize[0]")
	bad(map[string]any{"prize": []any{row(50, p10), map[string]any{"chance": -1}}}, "INVALID_PRIZE_CHANCE", "prize[1]")
	bad(map[string]any{"prize": []any{row(50, ""), row(50, "00000000-0000-0000-0000-000000000001")}}, "PRIZE_NOT_FOUND", "prize[1]")
	bad(map[string]any{"prize": []any{map[string]any{"chance": 100, "prizeGroupId": "00000000-0000-0000-0000-000000000001"}}}, "PRIZE_TYPE_MISMATCH", "prize[0]")
	bad(map[string]any{"prizeType": "box"}, "INVALID_PRIZE_TYPE", "prizeType")
	bad(map[string]any{"intervalTime": nil}, "INTERVAL_OR_PRICE_REQUIRED", "")
	bad(map[string]any{"intervalTime": 1.5}, "INVALID_INTERVAL_TIME", "intervalTime")
	bad(map[string]any{"intervalTime": 0}, "INVALID_INTERVAL_TIME", "intervalTime")
	bad(map[string]any{"price": map[string]any{"currencyId": h.currencyID("coin")}}, "INVALID_PRICE_PAYLOAD", "price")
	bad(map[string]any{"price": map[string]any{"currencyId": "00000000-0000-0000-0000-000000000001", "amount": 5}}, "CURRENCY_NOT_FOUND", "price.currencyId")
	bad(map[string]any{"price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 0}}, "INVALID_PRICE_AMOUNT", "price.amount")
	okRows := []any{row(33.3334, p10), row(33.3333, ""), row(33.3333, "")}
	g := h.chanceGame(map[string]any{"key": "wheel", "intervalTime": 60, "prize": okRows})
	if g["prizeType"] != "prize" || len(g["prize"].([]any)) != 3 {
		t.Fatalf("created = %v", g)
	}
	if c, out := h.api(http.MethodPut, "/chance-games/"+g["id"].(string), map[string]any{"key": "wheel2", "name": map[string]string{"fa": "x"}, "intervalTime": 5, "prize": okRows}); c != 400 || out["error"] != "KEY_IMMUTABLE" {
		t.Fatalf("key change = %d %v", c, out)
	}
	if c, out := h.api(http.MethodDelete, "/prizes/"+p10, nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete prize used by a chance game = %d %v", c, out)
	}

	p20 := h.coinPrize("c20", 20, true)
	if c, out := h.api(http.MethodPost, "/prize-groups", map[string]any{"key": "g0", "name": map[string]string{"fa": "گروه"}, "members": []any{}}); c != 400 || out["error"] != "PRIZE_GROUP_EMPTY" {
		t.Fatalf("empty group = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPost, "/prize-groups", map[string]any{"key": "g0", "name": map[string]string{"fa": "گروه"}, "members": []any{map[string]any{"prizeId": p20}, map[string]any{"prizeId": p20}}}); c != 400 || out["error"] != "DUPLICATE_PRIZE" || out["field"] != "members[1]" {
		t.Fatalf("duplicate member = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPost, "/prize-groups", map[string]any{"key": "g0", "name": map[string]string{"fa": "گروه"}, "members": []any{map[string]any{"prizeId": p20, "weight": 0}}}); c != 400 || out["error"] != "INVALID_WEIGHT" {
		t.Fatalf("zero weight = %d %v", c, out)
	}
	grp := h.mustAPI(201, http.MethodPost, "/prize-groups", map[string]any{"key": "g1", "name": map[string]string{"fa": "گروه"}, "members": []any{map[string]any{"prizeId": p20, "weight": 3}, map[string]any{"prizeId": p10}}})
	if m := grp["members"].([]any); len(m) != 2 || m[0].(map[string]any)["weight"].(float64) != 3 || m[1].(map[string]any)["weight"].(float64) != 1 {
		t.Fatalf("members = %v", grp)
	}
	bad(map[string]any{"prizeType": "prizeGroup", "prize": []any{row(100, p10)}}, "PRIZE_TYPE_MISMATCH", "prize[0]")
	bad(map[string]any{"prizeType": "prizeGroup", "prize": []any{map[string]any{"chance": 100, "prizeGroupId": "00000000-0000-0000-0000-000000000001"}}}, "PRIZE_GROUP_NOT_FOUND", "prize[0]")
	h.chanceGame(map[string]any{"key": "box", "prizeType": "prizeGroup", "price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 5},
		"prize": []any{map[string]any{"chance": 80, "prizeGroupId": grp["id"]}, map[string]any{"chance": 20}}})
	if c, out := h.api(http.MethodDelete, "/prize-groups/"+grp["id"].(string), nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete a used group = %d %v", c, out)
	}
	if c, out := h.api(http.MethodDelete, "/prizes/"+p20, nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete a group member prize = %d %v", c, out)
	}
	gem := h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "gem", "name": map[string]string{"fa": "گوهر"}, "type": "hard"})
	h.chanceGame(map[string]any{"key": "gems", "price": map[string]any{"currencyId": gem["id"], "amount": 1}, "prize": []any{row(100, "")}})
	if c, out := h.api(http.MethodDelete, "/currencies/"+gem["id"].(string), nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete a ticket currency = %d %v", c, out)
	}
	list := h.mustAPI(200, http.MethodGet, "/chance-games?mode=ticket", nil)
	if list["total"].(float64) != 2 {
		t.Fatalf("mode filter = %v", list)
	}
	if got := h.mustAPI(200, http.MethodGet, "/prize-groups/"+grp["id"].(string), nil); got["games"].(float64) != 1 {
		t.Fatalf("group games = %v", got)
	}
}

func TestChanceFreePlayWaitsForTurn(t *testing.T) {
	h := newChanceHarness(t)
	start := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	h.at(start)
	p50 := h.coinPrize("c50", 50, true)
	free := h.chanceGame(map[string]any{"key": "daily", "intervalTime": 60, "prize": []any{row(100, p50)}})
	ticketOnly := h.chanceGame(map[string]any{"key": "paid", "price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 5}, "prize": []any{row(100, "")}})
	user := h.user()
	id := free["id"].(string)
	for _, name := range []string{"economy_chance_can_play", "economy_chance_next_play_time", "economy_chance_play"} {
		if code := h.rpcCode(user, name, map[string]any{"id": ticketOnly["id"]}); code != "INTERVAL_TIME_IS_NULL" {
			t.Fatalf("%s on a ticket-only game = %s", name, code)
		}
	}
	if code := h.rpcCode(user, "economy_chance_play_with_ticket", map[string]any{"id": id}); code != "TICKET_PLAY_IS_NOT_SUPPORT" {
		t.Fatalf("ticket on a free game = %s", code)
	}
	if code := h.rpcCode(user, "economy_chance_play", map[string]any{"id": "00000000-0000-0000-0000-000000000001"}); code != "CHANCE_GAME_NOT_FOUND" {
		t.Fatalf("unknown game = %s", code)
	}
	if out := h.mustRPC(user, "economy_chance_can_play", map[string]any{"id": id}); out["canPlay"] != true || out["code"] != "SUCCESS" {
		t.Fatalf("can play = %v", out)
	}
	res := h.mustRPC(user, "economy_chance_play", map[string]any{"id": id, "lang": "en"})
	if res["code"] != "SUCCESS" || res["prizeId"] != p50 || res["status"] != "done" || res["nextPlayTime"].(float64) != 3600 {
		t.Fatalf("play = %v", res)
	}
	if w := h.wallet(user); w["coin"] != 50 {
		t.Fatalf("wallet after a free win = %v", w)
	}
	h.at(start.Add(20 * time.Minute))
	_, err := h.rpc(user, "economy_chance_play", map[string]any{"id": id})
	app, ok := asApp(err)
	if !ok || app.Code != "WAITING_FOR_TURN" || app.Status != http.StatusForbidden || app.Meta["nextPlayTime"].(int64) != 2400 {
		t.Fatalf("second play = %v", err)
	}
	if out := h.mustRPC(user, "economy_chance_can_play", map[string]any{"id": id}); out["canPlay"] != false {
		t.Fatalf("can play while waiting = %v", out)
	}
	if out := h.mustRPC(user, "economy_chance_next_play_time", map[string]any{"id": id}); out["nextPlayTime"].(float64) != 2400 {
		t.Fatalf("next play time = %v", out)
	}
	h.at(start.Add(61 * time.Minute))
	if out := h.mustRPC(user, "economy_chance_next_play_time", map[string]any{"id": id}); out["nextPlayTime"].(float64) != 0 {
		t.Fatalf("next play time after the wait = %v", out)
	}
	h.mustRPC(user, "economy_chance_play", map[string]any{"key": "daily"})
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE kind='chance' AND user_id=$1 AND status='granted'`, user); n != 2 {
		t.Fatalf("granted chance claims = %d", n)
	}
	if n := h.count(`SELECT count(DISTINCT id) FROM economy_reward_claim WHERE kind='chance' AND user_id=$1`, user); n != 2 {
		t.Fatalf("claim ids = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='ChanceGame:Played' AND e.user_id=$1 AND e.metadata->>'entityId'=$2`, user, id); n != 2 {
		t.Fatalf("ChanceGame:Played events = %d", n)
	}
	hist := h.mustRPC(user, "economy_chance_history", map[string]any{"lang": "en"})
	if items := hist["items"].([]any); len(items) != 2 || items[0].(map[string]any)["prizeId"] != p50 || len(items[0].(map[string]any)["rewards"].([]any)) != 1 {
		t.Fatalf("history = %v", hist)
	}
	games := h.mustRPC(user, "economy_chance_games", map[string]any{"lang": "en"})
	if items := games["items"].([]any); len(items) != 2 {
		t.Fatalf("games = %v", games)
	}
	plays := h.mustAPI(200, http.MethodGet, "/users/"+user+"/chance-plays", nil)
	if plays["total"].(float64) != 2 {
		t.Fatalf("console plays = %v", plays)
	}
	if c, out := h.api(http.MethodDelete, "/chance-games/"+id, nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete a played game = %d %v", c, out)
	}
}

func TestChanceParallelFreePlayOnce(t *testing.T) {
	h := newChanceHarness(t)
	p := h.coinPrize("c7", 7, true)
	g := h.chanceGame(map[string]any{"key": "spin", "intervalTime": 1440, "prize": []any{row(60, p), row(40, "")}})
	user := h.user()
	var wg sync.WaitGroup
	var mu sync.Mutex
	codes := map[string]int{}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.s.chancePlayRPC(h.ctx, user, `{"id":"`+g["id"].(string)+`"}`, "free")
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				codes["SUCCESS"]++
			} else if app, ok := asApp(err); ok {
				codes[app.Code]++
			} else {
				codes[err.Error()]++
			}
		}()
	}
	wg.Wait()
	if codes["SUCCESS"] != 1 || codes["WAITING_FOR_TURN"] != 15 {
		t.Fatalf("parallel free plays = %v", codes)
	}
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1`, user); n != 1 {
		t.Fatalf("play rows = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE status='pending'`); n != 0 {
		t.Fatalf("pending plays = %d", n)
	}
	won := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1 AND prize_id IS NOT NULL`, user)
	if w := h.wallet(user); w["coin"] != int64(7*won) {
		t.Fatalf("wallet %v for %d wins", w, won)
	}
}

func TestChanceParallelTicketPlaysSpendOnce(t *testing.T) {
	h := newChanceHarness(t)
	p := h.coinPrize("c50", 50, true)
	g := h.chanceGame(map[string]any{"key": "ticket", "price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 10}, "prize": []any{row(100, p)}})
	user := h.user()
	h.give(user, 5)
	_, err := h.rpc(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]})
	if app, ok := asApp(err); !ok || app.Code != "CURRENCY_IS_NOT_ENOUGH" || app.Status != http.StatusConflict {
		t.Fatalf("short balance = %v", err)
	}
	h.give(user, 25)
	var wg sync.WaitGroup
	var mu sync.Mutex
	codes := map[string]int{}
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.s.chancePlayRPC(h.ctx, user, `{"id":"`+g["id"].(string)+`"}`, "ticket")
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				codes["SUCCESS"]++
			} else if app, ok := asApp(err); ok {
				codes[app.Code]++
			} else {
				codes[err.Error()]++
			}
		}()
	}
	wg.Wait()
	if codes["SUCCESS"] < 3 || codes["SUCCESS"]+codes["CURRENCY_IS_NOT_ENOUGH"] != 12 {
		t.Fatalf("parallel ticket plays = %v", codes)
	}
	plays := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1`, user)
	if plays != codes["SUCCESS"] {
		t.Fatalf("%d play rows for %d wins", plays, codes["SUCCESS"])
	}
	debits := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'sourceType'='chance' AND metadata->>'part'='debit'`, user)
	if debits != plays {
		t.Fatalf("%d debits for %d plays", debits, plays)
	}
	if w := h.wallet(user); w["coin"] != 30+int64(plays)*40 {
		t.Fatalf("wallet %v after %d plays", w, plays)
	}
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE status='pending'`); n != 0 {
		t.Fatalf("pending plays = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE kind='chance' AND status<>'granted'`); n != 0 {
		t.Fatalf("open claims = %d", n)
	}
	small := h.coinPrize("c5", 5, true)
	cheap := h.chanceGame(map[string]any{"key": "cheap", "price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 10}, "prize": []any{row(100, small)}})
	poor := h.user()
	h.give(poor, 10)
	codes = map[string]int{}
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.s.chancePlayRPC(h.ctx, poor, `{"id":"`+cheap["id"].(string)+`"}`, "ticket")
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				codes["SUCCESS"]++
			} else if app, ok := asApp(err); ok {
				codes[app.Code]++
			}
		}()
	}
	wg.Wait()
	if codes["SUCCESS"] != 1 || codes["CURRENCY_IS_NOT_ENOUGH"] != 5 {
		t.Fatalf("one ticket of balance, parallel = %v", codes)
	}
	if w := h.wallet(poor); w["coin"] != 5 {
		t.Fatalf("a won prize must not pay for a second ticket: %v", w)
	}
}

func TestChancePrizeGroupAndFailures(t *testing.T) {
	h := newChanceHarness(t)
	a := h.coinPrize("ga", 1, true)
	b := h.coinPrize("gb", 2, true)
	grp := h.mustAPI(201, http.MethodPost, "/prize-groups", map[string]any{"key": "pool", "name": map[string]string{"fa": "سبد"}, "members": []any{map[string]any{"prizeId": a}, map[string]any{"prizeId": b}}})
	g := h.chanceGame(map[string]any{"key": "loot", "prizeType": "prizeGroup", "price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 1},
		"prize": []any{map[string]any{"chance": 100, "prizeGroupId": grp["id"]}}})
	user := h.user()
	h.give(user, 1)
	got := map[string]int{}
	for i := 0; i < 40; i++ {
		res := h.mustRPC(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"], "lang": "en"})
		if res["prizeGroupId"] != grp["id"] {
			t.Fatalf("group id = %v", res)
		}
		got[res["prizeId"].(string)]++
	}
	if got[a] == 0 || got[b] == 0 || got[a]+got[b] != 40 {
		t.Fatalf("group draws = %v", got)
	}

	h.nk.failWallet = true
	before := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1`, user)
	if _, err := h.rpc(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]}); err == nil {
		t.Fatal("play with the wallet down must fail")
	}
	h.nk.failWallet = false
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1`, user); n != before {
		t.Fatalf("a failed play left a row: %d -> %d", before, n)
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE kind='chance' AND status<>'granted'`); n != 0 {
		t.Fatalf("a failed play left a claim: %d", n)
	}

	once := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "once", "name": map[string]string{"fa": "یک بار"}, "isRepeatable": false,
		"items": []map[string]any{{"type": "currency", "key": "coin", "action": "add", "value": 100}}})
	single := h.chanceGame(map[string]any{"key": "single", "intervalTime": 1, "prize": []any{row(100, once["id"].(string))}})
	lucky := h.user()
	if res := h.mustRPC(lucky, "economy_chance_play", map[string]any{"id": single["id"]}); res["prizeId"] != once["id"] {
		t.Fatalf("first win of a one-time prize = %v", res)
	}
	h.at(time.Now().Add(2 * time.Minute))
	if res := h.mustRPC(lucky, "economy_chance_play", map[string]any{"id": single["id"]}); res["prizeId"] != nil || res["slot"].(float64) != 0 {
		t.Fatalf("a one-time prize already owned must count as nothing: %v", res)
	}
	h.s.now = time.Now
	if w := h.wallet(lucky); w["coin"] != 100 {
		t.Fatalf("one-time prize paid twice: %v", w)
	}

	nothing := h.chanceGame(map[string]any{"key": "blank", "price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 2}, "prize": []any{row(100, "")}})
	poor := h.user()
	h.give(poor, 2)
	res := h.mustRPC(poor, "economy_chance_play_with_ticket", map[string]any{"id": nothing["id"]})
	if res["prizeId"] != nil || res["status"] != "done" {
		t.Fatalf("empty slot = %v", res)
	}
	if w := h.wallet(poor); w["coin"] != 0 {
		t.Fatalf("empty slot still costs the ticket: %v", w)
	}

	stuck := h.user()
	h.give(stuck, 5)
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_chance_play (id, game_id, user_id, mode, status, roll, slot, price, created_at) VALUES (gen_random_uuid(), $1, $2, 'ticket', 'pending', 5, 0, '{"currencyKey":"coin","amount":2}', now() - interval '10 minutes')`, nothing["id"], stuck); err != nil {
		t.Fatal(err)
	}
	if n := h.s.recoverChancePlays(h.ctx, 5); n != 1 {
		t.Fatalf("recovered = %d", n)
	}
	if w := h.wallet(stuck); w["coin"] != 3 {
		t.Fatalf("recovered play wallet = %v", w)
	}
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1 AND status='done' AND event_at IS NOT NULL`, stuck); n != 1 {
		t.Fatalf("recovered play not done = %d", n)
	}
	if n := h.s.recoverChancePlays(h.ctx, 5); n != 0 {
		t.Fatalf("second recovery = %d", n)
	}
	broke := h.user()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_chance_play (id, game_id, user_id, mode, status, roll, slot, price, created_at) VALUES (gen_random_uuid(), $1, $2, 'ticket', 'pending', 5, 0, '{"currencyKey":"coin","amount":2}', now() - interval '10 minutes')`, nothing["id"], broke); err != nil {
		t.Fatal(err)
	}
	h.s.recoverChancePlays(h.ctx, 5)
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1`, broke); n != 0 {
		t.Fatalf("an unpaid stuck play must be cancelled: %d", n)
	}
}

func TestChanceRPCErrorShape(t *testing.T) {
	msg := chanceErrorMessage(waitingForTurn(90 * time.Second))
	var body map[string]any
	if err := json.Unmarshal([]byte(msg), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "WAITING_FOR_TURN" || body["status"].(float64) != 403 || body["meta"].(map[string]any)["nextPlayTime"].(float64) != 90 {
		t.Fatalf("error message = %s", msg)
	}
	if grpcCode(http.StatusForbidden) != 7 || grpcCode(http.StatusConflict) != 9 || grpcCode(http.StatusBadRequest) != 3 || grpcCode(http.StatusNotFound) != 5 {
		t.Fatal("status to gRPC code mapping changed")
	}
	s := testService()
	rpc := s.chanceRPC(func(context.Context, string, string) (any, error) { return nil, notEnough("coin", 10, 3) })
	_, err := rpc(context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "00000000-0000-0000-0000-000000000009"), nil, nil, nil, "{}")
	re, ok := err.(*runtime.Error)
	if !ok || re.Code != 9 || !strings.Contains(re.Message, `"CURRENCY_IS_NOT_ENOUGH"`) || !strings.Contains(re.Message, `"status":409`) {
		t.Fatalf("wrapped error = %#v", err)
	}
}
