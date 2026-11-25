package main

import (
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/friendroom"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/stealthbots"
)

func TestStealthPoolGrowsRealAccountsWithLevelLook(t *testing.T) {
	h := newStealthHarness(t)
	h.ludoGame([]map[string]any{{"key": "two", "players": 2}}, nil)
	skin := h.taxonomy("base_category", "item_type", "skin_"+randomSuffix(), map[string]string{"fa": "پوسته", "en": "Skin"})["id"].(string)
	h.item(map[string]any{"key": "sk_basic", "typeId": skin, "rarity": "common", "isStarter": true})
	h.item(map[string]any{"key": "sk_cheap", "typeId": skin, "rarity": "common", "price": map[string]any{"currencyKey": "coin", "amount": 50}})
	h.item(map[string]any{"key": "sk_rare", "typeId": skin, "rarity": "rare", "price": map[string]any{"currencyKey": "coin", "amount": 900}})
	h.item(map[string]any{"key": "sk_legend", "typeId": skin, "rarity": "legendary", "price": map[string]any{"currencyKey": "coin", "amount": 400}})
	tmpl := h.bot(2, 0)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_bot SET look='{"skin":"sk_cheap"}', rooms='[{"game":"other_game","room":""}]' WHERE user_id=$1`, tmpl); err != nil {
		t.Fatal(err)
	}
	h.stealthOn("tst_board", map[string]any{"poolMax": 3, "lookSpread": 0, "levelRange": 0})
	p := gameProvider{s: h.s}
	lease := func(level int) []gamecatalog.Bot {
		t.Helper()
		got, err := p.Bots(h.ctx, gamecatalog.BotQuery{Game: "tst_board", Room: "two", Count: 1, Level: level, LevelRange: 0, MatchID: "gr-" + randomSuffix()})
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	low := lease(1)
	high := lease(30)
	if len(low) != 1 || len(high) != 1 {
		t.Fatalf("an empty pool grows: %v %v", low, high)
	}
	for _, b := range append(low, high...) {
		var source, md string
		var old bool
		if err := h.s.db.QueryRowContext(h.ctx, `SELECT b.source, u.metadata::text, u.create_time < now() - interval '19 days' FROM economy_bot b JOIN users u ON u.id=b.user_id WHERE b.user_id=$1`, b.UserID).Scan(&source, &md, &old); err != nil {
			t.Fatalf("grown bot %s has no account: %v", b.UserID, err)
		}
		if source != botGrowSource || md != "{}" || !old {
			t.Fatal(source, md, old)
		}
		if n := h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id=$1 AND game_key='tst_board' AND games > 0`, b.UserID); n != 1 {
			t.Fatal("grown bot has starting stats")
		}
		if n := h.count(`SELECT count(*) FROM economy_bot_lease WHERE user_id=$1 AND busy`, b.UserID); n != 1 {
			t.Fatal("grown bot is leased")
		}
		card, err := h.s.playerCard(h.ctx, h.user(), b.UserID, "tst_board", "", "en")
		if err != nil || card["username"] != b.Username || card["gameStats"].(gameStatsView).Games == 0 {
			t.Fatalf("card: %v %v", card, err)
		}
		if m := clientMarkers(t, card); len(m) != 0 {
			t.Fatal(m)
		}
	}
	if low[0].Level != 1 || high[0].Level != 30 {
		t.Fatal(low[0].Level, high[0].Level)
	}
	if low[0].Look["skin"] != "sk_basic" || high[0].Look["skin"] != "sk_legend" {
		t.Fatalf("look follows the level: low %v high %v", low[0].Look, high[0].Look)
	}
	if none := lease(5); len(none) != 0 {
		t.Fatalf("pool max stops growth: %v", none)
	}
	h.stealthOn("tst_board", map[string]any{"autoGrow": false, "poolMax": 50})
	if none := lease(5); len(none) != 0 {
		t.Fatalf("auto grow off: %v", none)
	}
}

func TestStealthGhostHasAProfile(t *testing.T) {
	h := newStealthHarness(t)
	h.ludoGame([]map[string]any{{"key": "two", "players": 2}}, nil)
	h.stealthOn("tst_board", nil)
	id := uuid.Must(uuid.NewV4()).String()
	if err := (gameProvider{s: h.s}).KeepGhost(h.ctx, "tst_board", gamecatalog.Bot{UserID: id, Username: "sina_1392", Name: "سینا", Level: 7}); err != nil {
		t.Fatal(err)
	}
	card, err := h.s.playerCard(h.ctx, h.user(), id, "tst_board", "", "fa")
	if err != nil {
		t.Fatal(err)
	}
	gs := card["gameStats"].(gameStatsView)
	if card["username"] != "sina_1392" || card["displayName"] != "سینا" || gs.Games < 10 || gs.Wins == 0 || card["level"].(map[string]any)["level"] != 7 {
		t.Fatalf("ghost card: %v", card)
	}
	again, _ := h.s.playerCard(h.ctx, h.user(), id, "tst_board", "", "fa")
	if again["gameStats"].(gameStatsView).Games != gs.Games {
		t.Fatal("ghost stats stay the same")
	}
	users := stealthbots.GhostUsers(h.ctx, h.s.db, []string{id}, nil, nil)
	if len(users) != 1 || users[0].Username != "sina_1392" || users[0].CreateTime.AsTime().After(users[0].UpdateTime.AsTime()) {
		t.Fatal(users)
	}
	if _, err := h.s.playerCard(h.ctx, h.user(), uuid.Must(uuid.NewV4()).String(), "", "", "fa"); err == nil {
		t.Fatal("unknown id still not found")
	}
}

func TestRecordStatsOnceForPlayersAndPoolBots(t *testing.T) {
	h := newStealthHarness(t)
	h.ludoGame([]map[string]any{{"key": "two", "players": 2}}, nil)
	p := gameProvider{s: h.s}
	human, bot := h.user(), h.bot(4, 0)
	play := func(statuses ...string) string {
		t.Helper()
		m := "rs-" + randomSuffix()
		for _, e := range []gamecatalog.Entry{{MatchID: m, Game: "tst_board", Room: "two", UserID: human}, {MatchID: m, Game: "tst_board", Room: "two", UserID: bot, Bot: true}} {
			if _, err := p.Enter(h.ctx, e); err != nil {
				t.Fatal(err)
			}
		}
		if err := p.Start(h.ctx, m); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Finish(h.ctx, m, nil); err != nil {
			t.Fatal(err)
		}
		res := []gamecatalog.Result{{UserID: human, Status: statuses[0]}, {UserID: bot, Status: statuses[1]}}
		for i := 0; i < 2; i++ {
			if err := p.RecordStats(h.ctx, m, res); err != nil {
				t.Fatal(err)
			}
		}
		return m
	}
	play(gamecatalog.ResultWon, gamecatalog.ResultLost)
	if n := h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id=$1 AND games=1 AND wins=1`, human); n != 1 {
		t.Fatal("human stats written once")
	}
	if n := h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id=$1`, bot); n != 0 {
		t.Fatal("stealth off: bot stats stay")
	}
	h.stealthOn("tst_board", nil)
	play(gamecatalog.ResultDraw, gamecatalog.ResultDraw)
	play(gamecatalog.ResultLost, gamecatalog.ResultWon)
	if n := h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id=$1 AND games=3 AND wins=1 AND losses=1`, human); n != 1 {
		t.Fatal("draw counts a game")
	}
	if n := h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id=$1 AND games=2 AND wins=1 AND losses=0`, bot); n != 1 {
		t.Fatal("pool bot stats grow with stealth progress on")
	}
}

func TestBotMatchIsOnePerPlayer(t *testing.T) {
	h := newStealthHarness(t)
	for name, fn := range h.s.gameRoomRPCs() {
		h.rpcs[name] = fn
	}
	nk := &roomNakama{fakeNakama: h.nk, matches: map[string]*friendroom.Params{}, status: map[string]string{}, friends: map[string][]string{}}
	h.s.nk = nk
	h.boardGame(map[string]any{"botsAllowed": true})
	user := h.user()
	h.give(user, 100)
	first := h.mustRPC(user, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"})
	second := h.mustRPC(user, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"})
	if first["matchId"] == second["matchId"] {
		t.Fatal("stealth off keeps today's behaviour")
	}
	h.stealthOn("tst_board", nil)
	a := h.mustRPC(user, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"})
	b := h.mustRPC(user, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"})
	if a["matchId"] != b["matchId"] {
		t.Fatalf("a live match is returned again: %v %v", a, b)
	}
	id, made, err := h.s.oneMatch(h.ctx, user, "tst_board", func() (string, error) { t.Fatal("no second match"); return "", nil })
	if err != nil || made || id != a["matchId"] {
		t.Fatal(id, made, err)
	}
	delete(nk.matches, a["matchId"].(string))
	c := h.mustRPC(user, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"})
	if c["matchId"] == a["matchId"] {
		t.Fatal("a gone match is not reused")
	}
	h.stealthOn("tst_board", map[string]any{"oneLiveMatch": false})
	d := h.mustRPC(user, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"})
	if d["matchId"] == c["matchId"] {
		t.Fatal("setting off")
	}
}
