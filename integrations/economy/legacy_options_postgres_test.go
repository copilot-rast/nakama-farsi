package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

func (h *harness) chestGame(key string, cap int) map[string]any {
	h.t.Helper()
	return h.chanceGame(map[string]any{"key": key, "kind": "chest", "prizeType": "slots", "isActive": false, "freeReset": "daily", "rerollCap": cap, "prize": []any{
		map[string]any{"chance": 70, "kind": "currency", "currencyId": h.currencyID("coin"), "min": 120, "max": 180},
		map[string]any{"chance": 30, "kind": "currency", "currencyId": h.currencyID("gem"), "min": 3, "max": 6},
	}})
}

func chestTotals(t *testing.T, items []prizeItem) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	for _, it := range items {
		if it.Type != "chest" {
			t.Fatalf("not a chest entry: %+v", it)
		}
		if len(it.Rewards) == 0 {
			t.Fatalf("chest opened empty: %+v", it)
		}
		for _, r := range it.Rewards {
			v, _ := int64FromJSON(r.Value)
			out[r.Key] += v
		}
	}
	return out
}

func TestPrizeOpensChestAtClaimAndRandomAmount(t *testing.T) {
	h := newChanceHarness(t)
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس", "en": "Gem"}, "type": "hard"})
	g := h.chestGame("chest-1", 10000)
	if g["rerollCap"] != float64(10000) || g["isActive"] != false {
		t.Fatalf("chest game %v", g)
	}
	if c, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "chest", "key": "nope", "action": "open", "value": 1}}}); c != 400 || out["error"] != "CHEST_NOT_FOUND" {
		t.Fatalf("missing chest = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "currency", "key": "coin", "action": "random", "value": map[string]any{"min": 9, "max": 3}}}}); c != 400 || out["error"] != "INVALID_VALUE" {
		t.Fatalf("bad random = %d %v", c, out)
	}
	road := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "road", "name": map[string]string{"fa": "جاده"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "chest", "key": "chest-1", "action": "open", "value": 1}}})["id"].(string)
	user := h.user()
	for i := 0; i < 20; i++ {
		before := h.wallet(user)
		got, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: road, SourceType: "admin", SourceID: "chest-" + itoa(i), Reason: "test"})
		if err != nil {
			t.Fatal(err)
		}
		sum := chestTotals(t, got)
		after := h.wallet(user)
		if after["coin"]-before["coin"] != sum["coin"] || after["gem"]-before["gem"] != sum["gem"] {
			t.Fatalf("paid %v, wallet %v -> %v", sum, before, after)
		}
		if sum["coin"] != 0 && (sum["coin"] < 120 || sum["coin"] > 180) || sum["gem"] != 0 && (sum["gem"] < 3 || sum["gem"] > 6) {
			t.Fatalf("out of range %v", sum)
		}
		again, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: road, SourceType: "admin", SourceID: "chest-" + itoa(i), Reason: "test"})
		if err != nil || h.wallet(user)["coin"] != after["coin"] || chestTotals(t, again)["coin"] != sum["coin"] {
			t.Fatalf("a replay pays nothing and shows the same chest: %v %v", again, err)
		}
	}
	if n := h.count(`SELECT count(*) FROM economy_chance_play`); n != 0 {
		t.Fatalf("opening a chest from a prize is not a play: %d", n)
	}

	inf := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "inf", "name": map[string]string{"fa": "بی‌پایان"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "currency", "key": "coin", "action": "random", "value": map[string]any{"min": 150, "max": 180}}}})["id"].(string)
	for i := 0; i < 10; i++ {
		before := h.wallet(user)["coin"]
		got, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: inf, SourceType: "admin", SourceID: "inf-" + itoa(i), Reason: "test"})
		if err != nil {
			t.Fatal(err)
		}
		v, _ := int64FromJSON(got[0].Value)
		if got[0].Action != "add" || v < 150 || v > 180 || h.wallet(user)["coin"]-before != v {
			t.Fatalf("random amount %v", got)
		}
	}

	login := h.typeIDByKey("login")
	h.mustAPI(201, http.MethodPost, "/streaks", map[string]any{"key": "daily", "name": map[string]string{"fa": "روزانه"}, "eventTypeId": login, "periodDays": 1, "isActive": true,
		"rewards": []map[string]any{{"day": 1, "prizeId": road, "grantMode": "claim"}}})
	p2 := h.user()
	h.s.recordServerEvent(h.ctx, p2, "login", 1, nil)
	sk := h.mustRPC(p2, "economy_streaks", map[string]any{})["items"].([]any)[0].(map[string]any)
	claim := sk["claimable"].([]any)[0].(string)
	res := h.mustRPC(p2, "economy_reward_claim", map[string]any{"claimId": claim})
	rw := res["rewards"].([]any)[0].(map[string]any)
	inner := rw["rewards"].([]any)
	if rw["type"] != "chest" || rw["key"] != "chest-1" || len(inner) == 0 {
		t.Fatalf("claim response shows the opened chest: %v", res)
	}
	total := int64(0)
	for _, x := range inner {
		if x.(map[string]any)["key"] == "coin" {
			total += int64(x.(map[string]any)["value"].(float64))
		}
	}
	if h.wallet(p2)["coin"] != total {
		t.Fatalf("chest paid in the same claim: %v %v", h.wallet(p2), inner)
	}
}

func TestChanceDailyFreeTurnFreeFirstAndLevelGate(t *testing.T) {
	h := newChanceHarness(t)
	coin := h.currencyID("coin")
	win := h.coinPrize("wheel_win", 5, true)
	g := h.chanceGame(map[string]any{"key": "lucky-wheel", "freeReset": "daily", "freeFirst": true, "dayZone": "UTC", "dayStartMinute": 90, "minLevel": 5,
		"dailyLimit": 3, "price": map[string]any{"currencyId": coin, "amount": 10}, "priceLadder": []int{10, 15}, "prize": []any{row(100, win)}})
	if g["freeReset"] != "daily" || g["intervalTime"] != nil || g["minLevel"] != float64(5) || g["dayStartMinute"] != float64(90) {
		t.Fatalf("wheel %v", g)
	}
	if c, out := h.api(http.MethodPut, "/chance-games/"+g["id"].(string), map[string]any{"key": "lucky-wheel", "name": map[string]string{"fa": "x"}, "prize": []any{row(100, win)}, "freeReset": "daily", "dayZone": "Mars/Base"}); c != 400 || out["error"] != "INVALID_TIMEZONE" {
		t.Fatalf("bad zone = %d %v", c, out)
	}
	user := h.user()
	h.give(user, 100)
	day := time.Date(2026, 11, 21, 2, 0, 0, 0, time.UTC)
	h.at(day)
	id := map[string]any{"id": g["id"]}
	if code := h.rpcCode(user, "economy_chance_play", id); code != "CHANCE_LOCKED" {
		t.Fatalf("level 1 plays: %s", code)
	}
	view := h.mustRPC(user, "economy_chance_game", id)["game"].(map[string]any)
	if view["locked"] != true || view["canPlay"] != false || view["minLevel"] != float64(5) {
		t.Fatalf("locked view %v", view)
	}
	h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "level", "name": map[string]string{"fa": "سطح"}, "defaultValue": 1})
	if code := h.rpcCode(user, "economy_chance_play", id); code != "CHANCE_LOCKED" {
		t.Fatalf("level 1 on the level meta: %s", code)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) SELECT $1, id, 5 FROM economy_meta WHERE key='level'
		ON CONFLICT (user_id, meta_id) DO UPDATE SET value=5`, user); err != nil {
		t.Fatal(err)
	}
	_, err := h.rpc(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"], "expectedCost": 10})
	app, _ := asApp(err)
	if app == nil || app.Code != "CHANCE_PRICE_CHANGED" || app.Meta["amount"] != int64(0) || app.Meta["freeFirst"] != true {
		t.Fatalf("free turn first: %v", err)
	}
	if h.wallet(user)["coin"] != 100 {
		t.Fatal("refused ticket took coins")
	}
	h.mustRPC(user, "economy_chance_play", id)
	h.at(day.Add(time.Hour))
	if code := h.rpcCode(user, "economy_chance_play", id); code != "WAITING_FOR_TURN" {
		t.Fatalf("second free turn same day: %s", code)
	}
	left := int64(h.mustRPC(user, "economy_chance_next_play_time", id)["nextPlayTime"].(float64))
	if want := int64(time.Date(2026, 11, 22, 1, 30, 0, 0, time.UTC).Sub(day.Add(time.Hour)).Seconds()); left != want {
		t.Fatalf("free turn back at 01:30 UTC: %d want %d", left, want)
	}
	if out := h.mustRPC(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"], "expectedCost": 10}); out["nextTicketCost"] != float64(15) {
		t.Fatalf("ladder %v", out)
	}
	h.mustRPC(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"], "expectedCost": 15})
	if code := h.rpcCode(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]}); code != "CHANCE_DAILY_LIMIT" {
		t.Fatalf("1 free + 2 paid: %s", code)
	}
	h.at(time.Date(2026, 11, 22, 1, 29, 0, 0, time.UTC))
	if code := h.rpcCode(user, "economy_chance_play", id); code != "CHANCE_DAILY_LIMIT" {
		t.Fatalf("before 01:30 still the same day: %s", code)
	}
	h.at(time.Date(2026, 11, 22, 1, 31, 0, 0, time.UTC))
	view = h.mustRPC(user, "economy_chance_game", id)["game"].(map[string]any)
	if view["canPlay"] != true || view["playsToday"] != float64(0) || view["price"].(map[string]any)["amount"] != float64(10) {
		t.Fatalf("new day after 01:30 UTC: %v", view)
	}
	if code := h.rpcCode(user, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]}); code != "CHANCE_PRICE_CHANGED" {
		t.Fatalf("new day: free turn first again: %s", code)
	}
}

func TestStreakLegacyDayClaimWindowSignupDelayInfinity(t *testing.T) {
	h := newEventHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_event_type (id, key, name, is_system, rate_limit) VALUES (gen_random_uuid(), 'online', '{"fa":"آنلاین"}', true, 1000)`); err != nil {
		t.Fatal(err)
	}
	online := h.typeIDByKey("online")
	day := h.coinPrize("sk_day", 10, true)
	inf := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "sk_inf", "name": map[string]string{"fa": "بی‌پایان"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "currency", "key": "coin", "action": "random", "value": map[string]any{"min": 150, "max": 180}}}})["id"].(string)
	if c, out := h.api(http.MethodPost, "/streaks", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "eventTypeId": online, "claimWindow": "week"}); c != 400 || out["error"] != "INVALID_CLAIM_WINDOW" {
		t.Fatalf("bad window = %d %v", c, out)
	}
	sk := h.mustAPI(201, http.MethodPost, "/streaks", map[string]any{"key": "mench_daily", "name": map[string]string{"fa": "روزانه"}, "eventTypeId": online, "periodDays": 1,
		"missAction": "week_start", "weekLength": 7, "cycleLength": 3, "isActive": true, "dayZone": "UTC", "dayStartMinute": 90, "signupDelayDays": 1, "claimWindow": "period",
		"infinity": true, "infinityMode": "claim", "infinityPrizeId": inf,
		"rewards": []map[string]any{{"day": 1, "prizeId": day, "grantMode": "claim"}, {"day": 2, "prizeId": day, "grantMode": "claim"}, {"day": 3, "prizeId": day, "grantMode": "claim"}}})
	if sk["claimWindow"] != "period" || sk["infinityMode"] != "claim" || sk["signupDelayDays"] != float64(1) {
		t.Fatalf("streak %v", sk)
	}
	user := h.user()
	signup := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE users SET create_time=$2 WHERE id=$1`, user, signup); err != nil {
		t.Fatal(err)
	}
	view := func() map[string]any {
		return h.mustRPC(user, "economy_streaks", map[string]any{})["items"].([]any)[0].(map[string]any)
	}
	connect := func(at time.Time, session string) {
		h.at(at)
		ctx := context.WithValue(context.WithValue(h.ctx, runtime.RUNTIME_CTX_USER_ID, user), runtime.RUNTIME_CTX_SESSION_ID, session)
		h.s.onSessionStart(ctx, nil, nil)
	}
	connect(signup, "s1")
	if v := view(); v["count"] != float64(1) || len(v["claimable"].([]any)) != 0 {
		t.Fatalf("signup day counts but has no reward: %v", v)
	}
	h.at(time.Date(2026, 11, 21, 1, 0, 0, 0, time.UTC))
	h.s.onlineSweep(h.ctx)
	if v := view(); v["count"] != float64(1) {
		t.Fatalf("before 01:30 UTC it is still the signup day: %v", v)
	}
	h.at(time.Date(2026, 11, 21, 1, 31, 0, 0, time.UTC))
	if n := h.s.onlineSweep(h.ctx); n != 1 {
		t.Fatalf("still connected at the new day: %d", n)
	}
	v := view()
	if v["count"] != float64(2) || len(v["claimable"].([]any)) != 1 {
		t.Fatalf("day 2 by being online: %v", v)
	}
	claim := v["claimable"].([]any)[0].(string)
	h.mustRPC(user, "economy_reward_claim", map[string]any{"claimId": claim})
	h.s.onSessionEnd(context.WithValue(h.ctx, runtime.RUNTIME_CTX_SESSION_ID, "s1"), nil, nil)
	connect(time.Date(2026, 11, 22, 3, 0, 0, 0, time.UTC), "s2")
	v = view()
	if v["count"] != float64(3) || len(v["claimable"].([]any)) != 1 {
		t.Fatalf("day 3: %v", v)
	}
	unclaimed := v["claimable"].([]any)[0].(string)
	h.at(time.Date(2026, 11, 23, 1, 31, 0, 0, time.UTC))
	_, err := h.rpc(user, "economy_reward_claim", map[string]any{"claimId": unclaimed})
	if app, ok := asApp(err); !ok || app.Code != "CLAIM_EXPIRED" || app.Status != http.StatusGone {
		t.Fatalf("yesterday's reward expired: %v", err)
	}
	if v := view(); len(v["claimable"].([]any)) != 0 {
		t.Fatalf("expired reward not claimable: %v", v)
	}
	before := h.wallet(user)["coin"]
	h.s.onlineSweep(h.ctx)
	v = view()
	if v["phase"] != "infinity" {
		t.Fatalf("after the track ends: %v", v)
	}
	infView := v["infinity"].(map[string]any)
	if infView["mode"] != "claim" || infView["status"] != "claimable" || h.wallet(user)["coin"] != before {
		t.Fatalf("infinity day is a claimable reward, not paid straight: %v %v", infView, h.wallet(user))
	}
	res := h.mustRPC(user, "economy_reward_claim", map[string]any{"claimId": infView["claimId"]})
	got := h.wallet(user)["coin"] - before
	if got < 150 || got > 180 {
		t.Fatalf("infinity claim paid %d %v", got, res)
	}
}

func TestEntryOfferStartsAtFirstFinishedMatch(t *testing.T) {
	h := newSettleHarness(t)
	for _, q := range []string{
		`TRUNCATE economy_user_offer_period, economy_user_offer_state, economy_user_offer, economy_offer CASCADE`,
		`DELETE FROM economy_offer_settings`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for name, fn := range h.s.offerRPCs() {
		h.rpcs[name] = fn
	}
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس", "en": "Gem"}, "type": "hard"})
	h.mustAPI(200, http.MethodPut, "/payment-methods/in_app", map[string]any{"isActive": true})
	prize := h.coinPrize("np_pack", 1500, false)
	prod := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "new-player-offer-1", "isActive": false, "name": map[string]string{"fa": "تازه‌وارد"}, "prizeId": prize,
		"paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "gem", "amount": 100}})["id"].(string)
	h.offerSettings(map[string]any{"entryTimerStart": "trigger", "dailyEnabled": false})
	if c, out := h.api(http.MethodPost, "/entry-offers", map[string]any{"key": "np", "name": map[string]string{"fa": "x"}, "productId": prod, "triggerType": "game", "triggerValue": 1, "triggerRef": "nope"}); c != 400 || out["error"] != "GAME_NOT_FOUND" {
		t.Fatalf("unknown game = %d %v", c, out)
	}
	h.ludoGame([]map[string]any{{"key": "two", "players": 2, "entryFee": map[string]any{"currencyKey": "coin", "amount": 10}}}, nil)
	h.offer("entry-offers", "new-player-offer-1", prod, map[string]any{"triggerType": "game", "triggerValue": 1, "triggerRef": "tst_board", "durationMinutes": 4320, "discountPercent": 85})
	a, b := h.user(), h.user()
	h.give(a, 100)
	h.give(b, 100)
	if n := len(entryOf(h.mustRPC(a, "economy_offers", map[string]any{}))); n != 0 {
		t.Fatalf("no offer before any match: %d", n)
	}
	play := func(results ...gamecatalog.Result) {
		p := gameProvider{s: h.s}
		match := "tst-np-" + randomSuffix()
		for _, r := range results {
			if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "two", UserID: r.UserID}); err != nil {
				t.Fatal(err)
			}
		}
		if err := p.Start(h.ctx, match); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Settle(h.ctx, gamecatalog.Settlement{MatchID: match, Results: results}); err != nil {
			t.Fatal(err)
		}
	}
	end := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	h.at(end)
	play(gamecatalog.Result{UserID: a, Rank: 1, Status: "won"}, gamecatalog.Result{UserID: b, Rank: 2, Status: "forfeit"})
	if n := h.count(`SELECT count(*) FROM economy_user_offer WHERE user_id=$1 AND kind='entry'`, b); n != 0 {
		t.Fatalf("a forfeit is not a finished match: %d", n)
	}
	var expires time.Time
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT expires_at FROM economy_user_offer WHERE user_id=$1 AND kind='entry'`, a).Scan(&expires); err != nil {
		t.Fatalf("offer opened at settle: %v", err)
	}
	if !expires.Equal(end.Add(72 * time.Hour)) {
		t.Fatalf("72 h from the match end: %s", expires)
	}
	h.at(end.Add(5 * time.Hour))
	items := entryOf(h.mustRPC(a, "economy_offers", map[string]any{}))
	if len(items) != 1 || items[0].(map[string]any)["secondsLeft"] != float64(67*3600) {
		t.Fatalf("timer counts from the match end, not from the first list: %v", items)
	}
}

func TestScoredDailyOffersAndCoinFloor(t *testing.T) {
	h := newOfferHarness(t)
	keys := []string{"pawn-star", "avatar-nurse", "banner-dawn", "dice-yummy", "pawn-viking"}
	ids := map[string]string{}
	for i, k := range keys {
		prod := h.coinProduct("p-"+k, int64(1+i*100), 0, map[string]any{"paymentMethods": []string{"in_app"}, "priceToman": nil})
		ids[h.offer("offers", k, prod, map[string]any{"discountPercent": 10, "weight": 0, "metadata": map[string]any{"rarity": []string{"common", "common", "rare", "epic", ""}[i], "createdAt": "2026-08-24T14:49:00Z"}})] = k
	}
	h.offerSettings(map[string]any{"weighting": "scored", "timezone": "UTC", "resetHour": 1, "resetMinute": 30, "count": 3, "noRepeatDays": 0, "boughtCooldownDays": 0, "coinPriceFloor": 1})
	user := h.user()
	h.gems(user, 150)
	now := time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)
	h.at(now)
	got := offerKeys(dailyOf(h.mustRPC(user, "economy_offers", map[string]any{})))
	cands := []scoredCandidate{}
	for i, k := range keys {
		cands = append(cands, scoredCandidate{ID: k, Key: k, Currency: "gem", Price: int64(1 + i*100), HasPrice: true, Rarity: []string{"common", "common", "rare", "epic", ""}[i], CreatedAt: time.Date(2026, 8, 24, 14, 49, 0, 0, time.UTC)})
	}
	want := pickScored(cands, 3, user+":2026-11-21", map[string]int64{"gem": 150}, time.Date(2026, 11, 21, 1, 30, 0, 0, time.UTC), now, defaultOfferScore())
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("scored picks %v, want %v", got, want)
	}
	for _, it := range dailyOf(h.mustRPC(user, "economy_offers", map[string]any{})) {
		o := it.(map[string]any)
		if o["key"] == "pawn-star" && o["coinPrice"].(map[string]any)["amount"] != float64(1) {
			t.Fatalf("a 1 gem item at 10%% off stays 1: %v", o)
		}
	}
}

func TestPaymentPendingMinutesSetting(t *testing.T) {
	h := newHarness(t)
	h.mustAPI(200, http.MethodPut, "/payment-settings", map[string]any{"pendingMinutes": 0})
	t.Cleanup(func() { h.s.db.Exec(`UPDATE economy_payment_settings SET settings = settings - 'pendingMinutes'`) })
	if c, out := h.api(http.MethodPut, "/payment-settings", map[string]any{"pendingMinutes": 20000}); c != 400 || out["field"] != "pendingMinutes" {
		t.Fatalf("range = %d %v", c, out)
	}
	if d := h.s.pendingTTL(h.ctx); d != h.s.cfg.pendingTTL {
		t.Fatalf("0 keeps the server default: %s", d)
	}
	h.mustAPI(200, http.MethodPut, "/payment-settings", map[string]any{"pendingMinutes": 1440})
	if d := h.s.pendingTTL(h.ctx); d != 24*time.Hour {
		t.Fatalf("24 h: %s", d)
	}
}
