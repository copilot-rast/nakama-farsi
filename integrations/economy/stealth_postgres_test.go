package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/stealthbots"
)

func newStealthHarness(t *testing.T) *harness {
	h := newSettleHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_bot_lease, economy_bot_guard`); err != nil {
		t.Fatal(err)
	}
	stealthbots.Reset()
	return h
}

func (h *harness) stealthOn(key string, change map[string]any) map[string]any {
	h.t.Helper()
	body := map[string]any{"enabled": true}
	for k, v := range change {
		body[k] = v
	}
	out := h.mustAPI(200, http.MethodPut, "/game-stealth/"+key, body)
	stealthbots.Reset()
	return out
}

func clientMarkers(t *testing.T, v any) []string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return stealthbots.Markers(raw)
}

func TestStealthClientViewsHideBotSettings(t *testing.T) {
	h := newStealthHarness(t)
	h.ludoGame([]map[string]any{{"key": "two", "players": 2, "humanSeats": 1, "botFillSeconds": 12}}, map[string]any{"settings": map[string]any{"roundSeconds": 30}})
	user := h.user()
	off := h.mustRPC(user, "economy_game", map[string]any{"key": "tst_board", "lang": "en"})
	if off["botsAllowed"] != true || len(clientMarkers(t, off)) == 0 {
		t.Fatalf("stealth off keeps today's answer: %v", off)
	}
	if code, _ := h.api(http.MethodPut, "/game-stealth/nope", map[string]any{"enabled": true}); code != 404 {
		t.Fatal(code)
	}
	got := h.stealthOn("tst_board", map[string]any{"arriveMinSeconds": 30, "arriveMaxSeconds": 2})
	st := got["settings"].(map[string]any)
	if st["enabled"] != true || st["arriveMinSeconds"] != float64(2) || st["arriveMaxSeconds"] != float64(30) || st["trophies"] != true {
		t.Fatalf("saved settings are cleaned: %v", st)
	}
	view := h.mustRPC(user, "economy_game", map[string]any{"key": "tst_board", "lang": "en"})
	if m := clientMarkers(t, view); len(m) != 0 {
		t.Fatalf("economy_game leaks %v: %v", m, view)
	}
	if view["settings"].(map[string]any)["roundSeconds"] != float64(30) {
		t.Fatalf("other settings stay: %v", view["settings"])
	}
	list := h.mustRPC(user, "economy_games", map[string]any{"lang": "en"})
	if m := clientMarkers(t, list); len(m) != 0 {
		t.Fatalf("economy_games leaks %v", m)
	}
	q := h.mustAPI(200, http.MethodGet, "/game-stealth", nil)
	if len(q["items"].([]any)) != 1 {
		t.Fatal(q)
	}
}

func TestStealthBotLeaseOutcomesAndPresence(t *testing.T) {
	h := newStealthHarness(t)
	h.ludoGame([]map[string]any{{"key": "two", "players": 2, "entryFee": map[string]any{"currencyKey": "coin", "amount": 10}}}, nil)
	h.stealthOn("tst_board", map[string]any{"guard": true, "guardDailyWins": 1, "guardBoost": 2, "onlineAfterMinMinutes": 5, "onlineAfterMaxMinutes": 6, "onlineIdlePercent": 0})
	b3, b5, b12 := h.bot(3, 1000), h.bot(5, 1000), h.bot(12, 1000)
	for _, id := range []string{b3, b5, b12} {
		var md string
		var old bool
		_ = h.s.db.QueryRowContext(h.ctx, `SELECT metadata::text, create_time < now() - interval '19 days' FROM users WHERE id=$1`, id).Scan(&md, &old)
		if md != "{}" || !old {
			t.Fatalf("bot account looks fresh or flagged: %s old=%v", md, old)
		}
	}
	p := gameProvider{s: h.s}
	lease := func(match string) []gamecatalog.Bot {
		t.Helper()
		got, err := p.Bots(h.ctx, gamecatalog.BotQuery{Game: "tst_board", Room: "two", Count: 1, Level: 4, LevelRange: 1, MatchID: match})
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	m1, m2, m3, m4 := "st-1-"+randomSuffix(), "st-2-"+randomSuffix(), "st-3-"+randomSuffix(), "st-4-"+randomSuffix()
	first, second, third := lease(m1), lease(m2), lease(m3)
	if len(first) != 1 || len(second) != 1 || len(third) != 1 {
		t.Fatal(first, second, third)
	}
	near := map[string]bool{b3: true, b5: true}
	if !near[first[0].UserID] || !near[second[0].UserID] || first[0].UserID == second[0].UserID || third[0].UserID != b12 {
		t.Fatalf("level window then widen, one match per bot: %v %v %v", first, second, third)
	}
	if none := lease(m4); len(none) != 0 {
		t.Fatalf("every bot is busy: %v", none)
	}
	if on := stealthbots.Online(h.ctx, h.s.db, []string{first[0].UserID}); !on[first[0].UserID] {
		t.Fatal("a bot in a match is online")
	}
	human := h.user()
	h.give(human, 100)
	bot := first[0].UserID
	for _, e := range []gamecatalog.Entry{{MatchID: m1, Game: "tst_board", Room: "two", UserID: human}, {MatchID: m1, Game: "tst_board", Room: "two", UserID: bot, Bot: true}} {
		if _, err := p.Enter(h.ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.Start(h.ctx, m1); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Settle(h.ctx, gamecatalog.Settlement{MatchID: m1, Results: []gamecatalog.Result{{UserID: human, Rank: 1, Status: gamecatalog.ResultWon}, {UserID: bot, Rank: 2, Status: gamecatalog.ResultLost}}}); err != nil {
		t.Fatal(err)
	}
	if n := h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id=$1 AND losses=1`, bot); n != 1 {
		t.Fatal("bot stats grow with its games")
	}
	if n := h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id=$1 AND wins=1`, human); n != 1 {
		t.Fatal("human stats")
	}
	if n := h.count(`SELECT count(*) FROM economy_bot_lease WHERE user_id=$1 AND NOT busy AND online_until > now() + interval '4 minutes' AND online_until < now() + interval '7 minutes'`, bot); n != 1 {
		t.Fatal("lease released with an online tail")
	}
	if on := stealthbots.Online(h.ctx, h.s.db, []string{bot, b12}); !on[bot] || !on[b12] {
		t.Fatal(on)
	}
	st := h.s.stealthOf(h.ctx, "tst_board")
	if st.BoostFor(h.ctx, h.s.db, human, "tst_board") != 2 {
		t.Fatal("guard counts wins against bots")
	}
	if again := lease("st-5-" + randomSuffix()); len(again) != 1 || again[0].UserID != bot {
		t.Fatalf("released bot can play again: %v", again)
	}
	if err := p.ReleaseBots(h.ctx, m2); err != nil {
		t.Fatal(err)
	}
	h.mustAPI(200, http.MethodPut, "/game-stealth/tst_board", map[string]any{"enabled": false})
	stealthbots.Reset()
	if on := stealthbots.Online(h.ctx, h.s.db, []string{bot}); on[bot] {
		t.Fatal("stealth off: bots are never online")
	}
	list := h.mustAPI(200, http.MethodGet, "/bots?q="+bot, nil)
	items, _ := list["items"].([]any)
	if len(items) != 1 {
		t.Fatal(list)
	}
}
