package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func ticketPrice(h *harness, amount int) map[string]any {
	return map[string]any{"currencyId": h.currencyID("coin"), "amount": amount}
}

func TestChanceSlotsChest(t *testing.T) {
	h := newChanceHarness(t)
	coin := h.currencyID("coin")
	gem := h.coinPrize("chest_gem", 7, true)
	bad := func(body map[string]any, code string) {
		t.Helper()
		base := map[string]any{"key": "bad_chest", "name": map[string]string{"fa": "بد"}, "prizeType": "slots", "price": ticketPrice(h, 5)}
		for k, v := range body {
			base[k] = v
		}
		if c, out := h.api(http.MethodPost, "/chance-games", base); c != 400 || out["error"] != code {
			t.Fatalf("%v = %d %v, want %s", body, c, out, code)
		}
	}
	bad(map[string]any{"prize": []any{map[string]any{"chance": 0, "kind": "currency", "currencyId": coin, "min": 1, "max": 2}}}, "SLOTS_ALL_ZERO")
	bad(map[string]any{"prize": []any{map[string]any{"chance": 50, "kind": "currency", "currencyId": coin, "min": 5, "max": 2}}}, "INVALID_SLOT_RANGE")
	bad(map[string]any{"prize": []any{map[string]any{"chance": 50, "prizeId": gem, "jackpot": true}}}, "JACKPOT_NOT_FOR_SLOTS")
	bad(map[string]any{"prize": []any{map[string]any{"chance": 50, "prizeId": gem}}, "pityThreshold": 5}, "PITY_NOT_FOR_SLOTS")

	g := h.chanceGame(map[string]any{"key": "chest", "kind": "chest", "prizeType": "slots", "price": ticketPrice(h, 5), "prize": []any{
		map[string]any{"chance": 100, "kind": "currency", "currencyId": coin, "min": 120, "max": 180},
		map[string]any{"chance": 0.0001, "prizeId": gem},
	}})
	if g["prizeType"] != "slots" || g["rerollEmpty"] != true {
		t.Fatalf("chest = %v", g)
	}
	user := h.user()
	h.give(user, 5)
	out := h.mustRPC(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]})
	oc := out["outcome"].([]any)
	if out["status"] != "done" || len(oc) < 1 {
		t.Fatalf("chest play = %v", out)
	}
	first := oc[0].(map[string]any)
	amount := int64(first["amount"].(float64))
	if first["kind"] != "currency" || first["currencyKey"] != "coin" || amount < 120 || amount > 180 {
		t.Fatalf("coin slot = %v", first)
	}
	extra := int64(0)
	if len(oc) == 2 {
		extra = 7
	}
	if w := h.wallet(user); w["coin"] != amount+extra {
		t.Fatalf("wallet %v, want %d", w, amount+extra)
	}
	hist := h.mustRPC(user, "economy_chance_history", map[string]any{"id": g["id"]})["items"].([]any)[0].(map[string]any)
	if len(hist["outcome"].([]any)) != len(oc) || len(hist["rewards"].([]any)) < 1 {
		t.Fatalf("history = %v", hist)
	}

	rare := h.chanceGame(map[string]any{"key": "rare", "prizeType": "slots", "intervalTime": 1, "prize": []any{
		map[string]any{"chance": 1, "kind": "currency", "currencyId": coin, "min": 3, "max": 3},
	}})
	base := time.Now().UTC()
	for i := 0; i < 15; i++ {
		h.at(base.Add(time.Duration(i*2) * time.Minute))
		r := h.mustRPC(user, "economy_chance_play", map[string]any{"id": rare["id"]})
		if len(r["outcome"].([]any)) != 1 {
			t.Fatalf("re-roll on: play %d = %v", i, r)
		}
	}
	off := h.chanceGame(map[string]any{"key": "off", "prizeType": "slots", "intervalTime": 1, "rerollEmpty": false, "prize": []any{
		map[string]any{"chance": 0.0001, "kind": "currency", "currencyId": coin, "min": 3, "max": 3},
	}})
	h.at(base.Add(time.Hour))
	if r := h.mustRPC(user, "economy_chance_play", map[string]any{"id": off["id"]}); len(r["outcome"].([]any)) != 0 || r["status"] != "done" {
		t.Fatalf("re-roll off keeps an all-miss: %v", r)
	}
	poor := h.user()
	if code := h.rpcCode(poor, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]}); code != "CURRENCY_IS_NOT_ENOUGH" {
		t.Fatalf("poor = %s", code)
	}
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE status='pending'`); n != 0 {
		t.Fatalf("pending = %d", n)
	}
}

func TestChanceLadderPityAndReveal(t *testing.T) {
	h := newChanceHarness(t)
	jack := h.coinPrize("jackpot", 1000, true)
	small := h.coinPrize("small", 1, true)
	if c, out := h.api(http.MethodPost, "/chance-games", map[string]any{"key": "w0", "name": map[string]string{"fa": "x"}, "intervalTime": 60, "pityThreshold": 3,
		"prize": []any{row(100, small)}}); c != 400 || out["error"] != "PITY_NEEDS_JACKPOT" {
		t.Fatalf("pity without jackpot = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPost, "/chance-games", map[string]any{"key": "w0", "name": map[string]string{"fa": "x"}, "intervalTime": 60, "priceLadder": []int{5},
		"prize": []any{row(100, small)}}); c != 400 || out["error"] != "LADDER_NEEDS_PRICE" {
		t.Fatalf("ladder without price = %d %v", c, out)
	}
	g := h.chanceGame(map[string]any{"key": "wheel", "price": ticketPrice(h, 10), "priceLadder": []int{10, 15, 20}, "pityThreshold": 3, "revealAck": true,
		"prize": []any{row(99.9999, small), map[string]any{"chance": 0.0001, "prizeId": jack, "jackpot": true}}})
	user := h.user()
	h.give(user, 1000)
	view := h.mustRPC(user, "economy_chance_game", map[string]any{"id": g["id"]})["game"].(map[string]any)
	if view["price"].(map[string]any)["amount"].(float64) != 10 || view["pendingReveal"] != nil || view["revealAck"] != true {
		t.Fatalf("view = %v", view)
	}
	if code := h.rpcCode(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"], "expectedCost": 15}); code != "CHANCE_PRICE_CHANGED" {
		t.Fatalf("expected cost = %s", code)
	}
	spent := []float64{}
	jackpots := []int{}
	var last string
	for i := 1; i <= 6; i++ {
		out := h.mustRPC(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]})
		spent = append(spent, out["spent"].(map[string]any)["amount"].(float64))
		if out["slot"].(float64) == 1 {
			jackpots = append(jackpots, i)
		}
		last = out["playId"].(string)
		if out["needsAck"] != true {
			t.Fatalf("needsAck = %v", out)
		}
	}
	if spent[0] != 10 || spent[1] != 15 || spent[2] != 20 || spent[5] != 20 {
		t.Fatalf("ladder = %v", spent)
	}
	if len(jackpots) != 2 || jackpots[0] != 3 || jackpots[1] != 6 {
		t.Fatalf("pity forces the jackpot on plays 3 and 6: %v", jackpots)
	}
	if n := h.count(`SELECT count(*) FROM economy_chance_play WHERE user_id=$1 AND forced`, user); n != 2 {
		t.Fatalf("forced = %d", n)
	}
	view = h.mustRPC(user, "economy_chance_game", map[string]any{"id": g["id"]})["game"].(map[string]any)
	rev, _ := view["pendingReveal"].(map[string]any)
	if rev == nil || rev["playId"] != last || view["price"].(map[string]any)["amount"].(float64) != 20 {
		t.Fatalf("pending reveal = %v", view)
	}
	other := h.user()
	if code := h.rpcCode(other, "economy_chance_ack", map[string]any{"playId": last}); code != "PLAY_NOT_FOUND" {
		t.Fatalf("other ack = %s", code)
	}
	if a := h.mustRPC(user, "economy_chance_ack", map[string]any{"playId": last}); a["acknowledged"] != true {
		t.Fatalf("ack = %v", a)
	}
	rev, _ = h.mustRPC(user, "economy_chance_game", map[string]any{"id": g["id"]})["game"].(map[string]any)["pendingReveal"].(map[string]any)
	if rev == nil || rev["playId"] == last {
		t.Fatalf("the previous unacked play shows next: %v", rev)
	}
}

func TestBoosterMultipliesMetaRules(t *testing.T) {
	h := newEventHarness(t)
	typeID := h.eventType("match_end", false, nil)
	xp := h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "xp", "name": map[string]string{"fa": "تجربه", "en": "XP"}})
	h.mustAPI(201, http.MethodPost, "/meta-rules", map[string]any{"key": "xp_match", "name": map[string]string{"fa": "x"}, "eventTypeId": typeID, "metaId": xp["id"],
		"action": "add", "valueSource": "fixed", "amount": 10, "isActive": true})
	if c, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "bad_boost", "name": map[string]string{"fa": "x"},
		"items": []map[string]any{{"type": "booster", "key": "xp", "action": "give", "value": map[string]any{"multiplier": 1, "minutes": 60}}}}); c != 400 || out["error"] != "INVALID_VALUE" {
		t.Fatalf("bad booster = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "bad_boost", "name": map[string]string{"fa": "x"},
		"items": []map[string]any{{"type": "booster", "key": "nope", "action": "give", "value": map[string]any{"multiplier": 2, "minutes": 60}}}}); c != 400 || out["error"] != "META_NOT_FOUND" {
		t.Fatalf("booster meta = %d %v", c, out)
	}
	give := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "xp2", "name": map[string]string{"fa": "دوبرابر"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "booster", "key": "xp", "action": "give", "value": map[string]any{"multiplier": 2, "minutes": 60}}}})["id"].(string)
	set := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "xp3", "name": map[string]string{"fa": "سه‌برابر"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "booster", "key": "xp", "action": "set", "value": map[string]any{"multiplier": 3, "minutes": 30}}}})["id"].(string)
	user := h.user()
	grant := func(prize string) {
		t.Helper()
		if _, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: prize, SourceType: "admin", SourceID: uuid.Must(uuid.NewV4()).String(), Reason: "test"}); err != nil {
			t.Fatal(err)
		}
	}
	meta := func() int {
		return h.count(`SELECT value::int FROM economy_user_meta WHERE user_id=$1 AND meta_id=$2`, user, xp["id"])
	}
	base := time.Now().UTC()
	h.at(base)
	h.s.recordServerEvent(h.ctx, user, "match_end", 1, nil)
	if meta() != 10 {
		t.Fatalf("no booster = %d", meta())
	}
	grant(give)
	h.s.recordServerEvent(h.ctx, user, "match_end", 1, nil)
	if meta() != 30 {
		t.Fatalf("2x booster = %d", meta())
	}
	grant(give)
	left := h.count(`SELECT EXTRACT(EPOCH FROM expires_at - now())::int FROM economy_booster WHERE user_id=$1`, user)
	if left < 119*60 || left > 120*60 {
		t.Fatalf("give extends the time: %d s", left)
	}
	items := h.mustRPC(user, "economy_boosters", map[string]any{"lang": "en"})["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["metaKey"] != "xp" || items[0].(map[string]any)["multiplier"].(float64) != 2 {
		t.Fatalf("boosters = %v", items)
	}
	grant(set)
	left = h.count(`SELECT EXTRACT(EPOCH FROM expires_at - now())::int FROM economy_booster WHERE user_id=$1`, user)
	if left > 30*60 || left < 29*60 {
		t.Fatalf("set restarts the time: %d s", left)
	}
	h.s.recordServerEvent(h.ctx, user, "match_end", 1, nil)
	if meta() != 60 {
		t.Fatalf("3x booster, no stacking = %d", meta())
	}
	h.at(base.Add(2 * time.Hour))
	h.s.recordServerEvent(h.ctx, user, "match_end", 1, nil)
	if meta() != 70 {
		t.Fatalf("after expiry = %d", meta())
	}
	h.at(time.Now().UTC().Add(3 * time.Hour))
	if items := h.mustRPC(user, "economy_boosters", map[string]any{})["items"].([]any); len(items) != 0 {
		t.Fatalf("expired boosters listed: %v", items)
	}
}

func TestStreakWeekStartExtrasAndInfinity(t *testing.T) {
	h := newEventHarness(t)
	login := h.typeIDByKey("login")
	day := h.coinPrize("wk_day", 1, true)
	extra := h.coinPrize("wk_extra", 100, true)
	if c, out := h.api(http.MethodPost, "/streaks", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "eventTypeId": login, "infinity": true, "infinityCurrency": "coin", "infinityMin": 1, "infinityMax": 2}); c != 400 || out["error"] != "INFINITY_NEEDS_CYCLE" {
		t.Fatalf("infinity without cycle = %d %v", c, out)
	}
	h.mustAPI(201, http.MethodPost, "/streaks", map[string]any{"key": "track", "name": map[string]string{"fa": "جایزهٔ روزانه"}, "eventTypeId": login,
		"periodDays": 1, "missAction": "week_start", "weekLength": 3, "cycleLength": 6, "isActive": true,
		"infinity": true, "infinityCurrency": "coin", "infinityMin": 150, "infinityMax": 180,
		"rewards":  []map[string]any{{"day": 1, "prizeId": day, "grantMode": "auto"}, {"day": 2, "prizeId": day, "grantMode": "auto"}, {"day": 3, "prizeId": day, "grantMode": "auto"}, {"day": 4, "prizeId": day, "grantMode": "auto"}, {"day": 5, "prizeId": day, "grantMode": "auto"}, {"day": 6, "prizeId": day, "grantMode": "auto"}},
		"extras":   []map[string]any{{"day": 2, "prizeId": extra, "grantMode": "auto"}, {"day": 4, "prizeId": extra, "grantMode": "auto"}}})
	user := h.user()
	base := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)
	on := func(d int) {
		h.at(base.Add(time.Duration(d) * 24 * time.Hour))
		h.s.recordServerEvent(h.ctx, user, "login", 1, nil)
	}
	view := func() map[string]any {
		return h.mustRPC(user, "economy_streaks", map[string]any{})["items"].([]any)[0].(map[string]any)
	}
	for d := 0; d < 5; d++ {
		on(d)
	}
	if v := view(); v["count"].(float64) != 5 || h.wallet(user)["coin"] != 5+200 {
		t.Fatalf("five days with extras on day 2 and 4: %v %v", v, h.wallet(user))
	}
	on(7)
	v := view()
	if v["count"].(float64) != 4 || h.wallet(user)["coin"] != 205 {
		t.Fatalf("a miss in week two goes back to day 4 without the extra: %v %v", v, h.wallet(user))
	}
	on(8)
	on(9)
	if v := view(); v["count"].(float64) != 6 || v["phase"] != "track" {
		t.Fatalf("track end = %v", v)
	}
	before := h.wallet(user)["coin"]
	on(10)
	v = view()
	got := h.wallet(user)["coin"] - before
	if v["phase"] != "infinity" || v["count"].(float64) != 7 || got < 150 || got > 180 {
		t.Fatalf("infinity day = %v +%d", v, got)
	}
	before = h.wallet(user)["coin"]
	on(14)
	got = h.wallet(user)["coin"] - before
	if v := view(); v["count"].(float64) != 8 || v["broken"] == true || got < 150 || got > 180 {
		t.Fatalf("infinity never falls back: %v +%d", v, got)
	}
}
