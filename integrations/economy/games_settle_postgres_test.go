package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

func (h *harness) ludoGame(rooms []map[string]any, extra map[string]any) map[string]any {
	h.t.Helper()
	body := map[string]any{"key": "tst_board", "isActive": true, "name": map[string]string{"fa": "منچ", "en": "Ludo"}, "minPlayers": 2, "maxPlayers": 4, "botsAllowed": true, "rooms": rooms}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/games", body)
}

func (h *harness) bot(level int, coins int64) string {
	h.t.Helper()
	out := h.mustAPI(200, http.MethodPost, "/bots/bulk", map[string]any{"names": []string{"bot_" + randomSuffix() + strings.Repeat("x", level%3)}, "levelMin": level, "levelMax": level, "template": "normal", "wallet": map[string]int64{"coin": coins}})
	if out["created"] != float64(1) {
		h.t.Fatalf("bot not created: %v", out)
	}
	var id string
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT user_id::text FROM economy_bot ORDER BY created_at DESC, user_id LIMIT 1`).Scan(&id); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func newSettleHarness(t *testing.T) *harness {
	h := newGameHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_game_action, economy_game_stats, economy_bot`); err != nil {
		t.Fatal(err)
	}
	for name, fn := range h.s.gameLiveRPCs() {
		h.rpcs[name] = fn
	}
	return h
}

func TestGameRoomPlayFieldsValidation(t *testing.T) {
	h := newSettleHarness(t)
	base := func(room map[string]any) map[string]any {
		r := map[string]any{"key": "r2", "entryFee": map[string]any{"currencyKey": "coin", "amount": 10}}
		for k, v := range room {
			r[k] = v
		}
		return map[string]any{"key": "tst_board", "name": map[string]string{"fa": "منچ"}, "minPlayers": 2, "maxPlayers": 4, "rooms": []map[string]any{r}}
	}
	for _, tc := range []struct {
		room map[string]any
		code string
	}{
		{map[string]any{"players": 6}, "INVALID_ROOM_PLAYERS"},
		{map[string]any{"players": 2, "humanSeats": 3}, "INVALID_HUMAN_SEATS"},
		{map[string]any{"payout": "pot"}, "INVALID_PAYOUT_PERCENT"},
		{map[string]any{"payout": "pot", "payoutPercent": 120}, "INVALID_PAYOUT_PERCENT"},
		{map[string]any{"payout": "jackpot"}, "INVALID_PAYOUT"},
		{map[string]any{"botFillSeconds": 900}, "INVALID_BOT_FILL"},
		{map[string]any{"newPlayersOnly": true}, "REPLACES_REQUIRED"},
		{map[string]any{"winnerXp": -1}, "INVALID_XP"},
	} {
		code, out := h.api(http.MethodPost, "/games", base(tc.room))
		if code != 400 || out["error"] != tc.code {
			t.Fatalf("%v: %d %v", tc.room, code, out)
		}
	}
	code, out := h.api(http.MethodPost, "/games", map[string]any{"key": "tst_board", "name": map[string]string{"fa": "منچ"}, "minPlayers": 2, "maxPlayers": 4,
		"rooms": []map[string]any{{"key": "aa"}, {"key": "bb", "newPlayersOnly": true, "replaces": "aa"}, {"key": "cc", "newPlayersOnly": true, "replaces": "aa"}}})
	if code != 400 || out["error"] != "REPLACES_TAKEN" {
		t.Fatalf("two new player rooms on one slot: %d %v", code, out)
	}
}

func TestGameSettlePotLoserPrizeRefundAndStats(t *testing.T) {
	h := newSettleHarness(t)
	loser := h.coinPrize("tst_lose_"+randomSuffix(), 7, true)
	h.ludoGame([]map[string]any{{"key": "four", "players": 4, "humanSeats": 2, "botFillSeconds": 15, "botsPay": true, "entryFee": map[string]any{"currencyKey": "coin", "amount": 100},
		"payout": "pot", "payoutPercent": 75, "loserPrizeId": loser, "winnerXp": 170, "loserXp": 130}}, nil)
	a, b := h.user(), h.user()
	h.give(a, 500)
	h.give(b, 500)
	botA, botB := h.bot(5, 1000), h.bot(9, 50)
	p := gameProvider{s: h.s}
	match := "tst-settle-" + randomSuffix()
	for _, u := range []string{a, b} {
		if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "four", UserID: u}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "four", UserID: a, Bot: true}); gamecatalog.Code(err) != "NOT_A_BOT" {
		t.Fatalf("a human cannot enter as a bot: %v", err)
	}
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "four", UserID: botB, Bot: true}); gamecatalog.Code(err) != "INSUFFICIENT_BALANCE" {
		t.Fatalf("a poor bot cannot pay: %v", err)
	}
	picked, err := p.Bots(h.ctx, gamecatalog.BotQuery{Game: "tst_board", Room: "four", Count: 2, Fee: &gamecatalog.Fee{CurrencyKey: "coin", Amount: 100}})
	if err != nil || len(picked) != 1 || picked[0].UserID != botA || picked[0].Level != 5 || picked[0].Template != "normal" {
		t.Fatalf("only the funded bot is picked: %v %v", picked, err)
	}
	if r, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "four", UserID: botA, Bot: true, Path: gamecatalog.PathBot}); err != nil || !r.Charged {
		t.Fatalf("bot pays from its wallet: %v %v", r, err)
	}
	h.give(botB, 100)
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "four", UserID: botB, Bot: true, Path: gamecatalog.PathBot}); err != nil {
		t.Fatal(err)
	}
	if err := p.Start(h.ctx, match); err != nil {
		t.Fatal(err)
	}
	if err := p.Log(h.ctx, match, []gamecatalog.Action{{UserID: a, Kind: "roll", Data: map[string]any{"face": 6}}, {UserID: a, Kind: "move", Data: map[string]any{"token": 1}}}); err != nil {
		t.Fatal(err)
	}
	if err := p.Log(h.ctx, match, []gamecatalog.Action{{Kind: "turn"}}); err != nil {
		t.Fatal(err)
	}
	if n := h.count(`SELECT count(*) FROM economy_game_action WHERE match_id=$1 AND seq IN (1,2,3)`, match); n != 3 {
		t.Fatalf("actions %d", n)
	}
	payouts, err := p.Settle(h.ctx, gamecatalog.Settlement{MatchID: match, Reason: "finished", Results: []gamecatalog.Result{
		{UserID: a, Rank: 1, Status: "won"}, {UserID: botA, Rank: 2, Status: "lost"}, {UserID: b, Rank: 3, Status: "forfeit"}, {UserID: botB, Rank: 4, Status: "refunded"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(payouts) != 4 {
		t.Fatalf("payouts %v", payouts)
	}
	if w := h.wallet(a)["coin"]; w != 400+225 {
		t.Fatalf("winner gets 75%% of the three paid fees: %d", w)
	}
	if w := h.wallet(botA)["coin"]; w != 900+7 {
		t.Fatalf("losing bot keeps the loser prize: %d", w)
	}
	if w := h.wallet(b)["coin"]; w != 400 {
		t.Fatalf("forfeit gets nothing back: %d", w)
	}
	if w := h.wallet(botB)["coin"]; w != 150 {
		t.Fatalf("refunded bot gets its fee back: %d", w)
	}
	again, err := p.Settle(h.ctx, gamecatalog.Settlement{MatchID: match, Results: []gamecatalog.Result{{UserID: b, Rank: 1, Status: "won"}}})
	if err != nil || len(again) != 4 || h.wallet(b)["coin"] != 400 {
		t.Fatalf("settle is idempotent: %v %v", again, err)
	}
	stats := h.mustRPC(a, "economy_game_stats", map[string]any{"game": "tst_board"})["items"].([]any)[0].(map[string]any)
	if stats["wins"] != float64(1) || stats["totalGames"] != float64(1) || stats["winnings"].(map[string]any)["coin"] != float64(225) {
		t.Fatalf("winner stats %v", stats)
	}
	fs := h.mustRPC(b, "economy_game_stats", map[string]any{"game": "tst_board"})["items"].([]any)[0].(map[string]any)
	if fs["forfeits"] != float64(1) || fs["totalGames"] != float64(0) {
		t.Fatalf("forfeit stats keep games unchanged: %v", fs)
	}
	if h.count(`SELECT count(*) FROM economy_game_stats WHERE user_id IN ($1,$2)`, botA, botB) != 0 {
		t.Fatal("bots have no career stats")
	}
	m := h.mustAPI(200, http.MethodGet, "/game-matches/"+match, nil)
	if m["endReason"] != "finished" || m["bots"] != float64(2) || m["actions"] != float64(3) {
		t.Fatalf("match row %v", m)
	}
	acts := h.mustAPI(200, http.MethodGet, "/game-matches/"+match+"/actions?after=1", nil)
	if len(acts["items"].([]any)) != 2 || acts["total"] != float64(3) {
		t.Fatalf("actions %v", acts)
	}
	card := h.mustRPC(b, "economy_player_card", map[string]any{"userId": botA, "game": "tst_board"})
	if card["level"].(map[string]any)["level"] != float64(5) || card["gameStats"].(map[string]any)["totalGames"] != float64(0) {
		t.Fatalf("bot card %v", card)
	}
}

func TestGameOneMatchOverdraftNewPlayersAndVoid(t *testing.T) {
	h := newSettleHarness(t)
	h.ludoGame([]map[string]any{
		{"key": "board1", "players": 2, "entryFee": map[string]any{"currencyKey": "coin", "amount": 250}, "overdraft": true},
		{"key": "starter", "players": 2, "newPlayersOnly": true, "replaces": "board1", "entryFee": map[string]any{"currencyKey": "coin", "amount": 250}},
		{"key": "board2", "players": 2, "disabled": true},
	}, map[string]any{"oneMatchPerPlayer": true})
	vet, fresh := h.user(), h.user()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_game_stats (user_id, game_key, games, wins) VALUES ($1,'tst_board',1,1)`, vet); err != nil {
		t.Fatal(err)
	}
	h.give(vet, 100)
	view := h.mustRPC(fresh, "economy_game", map[string]any{"key": "tst_board"})
	rooms := view["rooms"].([]any)
	if len(rooms) != 1 || rooms[0].(map[string]any)["key"] != "starter" || view["newPlayer"] != true {
		t.Fatalf("a new player sees the starter room in the board1 slot: %v", rooms)
	}
	vview := h.mustRPC(vet, "economy_game", map[string]any{"key": "tst_board"})
	if r := vview["rooms"].([]any); len(r) != 1 || r[0].(map[string]any)["key"] != "board1" {
		t.Fatalf("a veteran sees board1 only: %v", r)
	}
	if c := h.rpcCode(fresh, "economy_game_join", map[string]any{"key": "tst_board", "room": "board1"}); c != "NEW_PLAYER_ROOM_REQUIRED" {
		t.Fatal(c)
	}
	if c := h.rpcCode(vet, "economy_game_join", map[string]any{"key": "tst_board", "room": "starter"}); c != "VETERAN_ROOM_FORBIDDEN" {
		t.Fatal(c)
	}
	if c := h.rpcCode(vet, "economy_game_join", map[string]any{"key": "tst_board", "room": "board2"}); c != "ROOM_DISABLED" {
		t.Fatal(c)
	}
	h.mustRPC(vet, "economy_game_join", map[string]any{"key": "tst_board", "room": "board1"})
	p := gameProvider{s: h.s}
	m1, m2 := "tst-one-"+randomSuffix(), "tst-two-"+randomSuffix()
	if r, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: m1, Game: "tst_board", Room: "board1", UserID: vet}); err != nil || !r.Charged {
		t.Fatalf("overdraft entry: %v %v", r, err)
	}
	if w := h.wallet(vet)["coin"]; w != -150 {
		t.Fatalf("overdraft wallet %d", w)
	}
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: m2, Game: "tst_board", Room: "board1", UserID: vet}); gamecatalog.Code(err) != "ACTIVE_MATCH_EXISTS" {
		t.Fatalf("one active match: %v", err)
	}
	act := h.mustRPC(vet, "economy_game_active", map[string]any{})["match"].(map[string]any)
	if act["matchId"] != m1 || act["room"] != "board1" {
		t.Fatalf("active %v", act)
	}
	if err := p.Start(h.ctx, m1); err != nil {
		t.Fatal(err)
	}
	if err := p.Release(h.ctx, m1, vet); err != nil {
		t.Fatal(err)
	}
	if out := h.mustRPC(vet, "economy_game_active", map[string]any{}); out["match"] != nil {
		t.Fatalf("released player is free: %v", out)
	}
	if n, err := p.Void(h.ctx, m1, "admin-cancel", "staff"); err != nil || n != 1 || h.wallet(vet)["coin"] != 100 {
		t.Fatalf("void refunds a started match: %d %v %d", n, err, h.wallet(vet)["coin"])
	}
	if code, out := h.api(http.MethodPost, "/game-matches/"+m1+"/end", map[string]any{"winnerUserId": vet}); code != 409 || out["error"] != "MATCH_CANCELLED" {
		t.Fatalf("end after void: %d %v", code, out)
	}
	m3 := "tst-three-" + randomSuffix()
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: m3, Game: "tst_board", Room: "board1", UserID: vet}); err != nil {
		t.Fatal(err)
	}
	other := h.user()
	h.give(other, 1000)
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_game_stats (user_id, game_key, games, losses) VALUES ($1,'tst_board',1,1)`, other); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: m3, Game: "tst_board", Room: "board1", UserID: other}); err != nil {
		t.Fatal(err)
	}
	if code, out := h.api(http.MethodPost, "/game-matches/"+m3+"/end", map[string]any{"winnerUserId": other}); code != 409 || out["error"] != "MATCH_NOT_STARTED" {
		t.Fatalf("end before start: %d %v", code, out)
	}
	if err := p.Start(h.ctx, m3); err != nil {
		t.Fatal(err)
	}
	out := h.mustAPI(200, http.MethodPost, "/game-matches/"+m3+"/end", map[string]any{"winnerUserId": other, "reason": "stuck"})
	if out["via"] != "server" {
		t.Fatalf("end %v", out)
	}
	row := h.mustAPI(200, http.MethodGet, "/game-matches/"+m3, nil)
	if row["status"] != "finished" || row["endReason"] != "admin-force-end" || row["note"] != "stuck" {
		t.Fatalf("forced row %v", row)
	}
}

func TestBotConsoleBulkImportStatusTopup(t *testing.T) {
	h := newSettleHarness(t)
	h.ludoGame([]map[string]any{{"key": "r1"}, {"key": "r2"}}, nil)
	out := h.mustAPI(200, http.MethodPost, "/bots/bulk", map[string]any{"count": 5, "prefix": "tbot", "levelMin": 2, "levelMax": 25,
		"bands": []map[string]any{{"min": 2, "max": 5, "template": "easy"}, {"min": 6, "max": 10, "template": "normal"}, {"min": 11, "max": 25, "template": "hard"}},
		"wallet": map[string]int64{"coin": 100000}, "rooms": []map[string]any{{"game": "tst_board", "room": "r1"}}})
	if out["created"] != float64(5) {
		t.Fatalf("bulk %v", out)
	}
	if h.count(`SELECT count(*) FROM economy_bot WHERE (level BETWEEN 2 AND 5 AND template='easy') OR (level BETWEEN 6 AND 10 AND template='normal') OR (level BETWEEN 11 AND 25 AND template='hard')`) != 5 {
		t.Fatal("level bands pick the template")
	}
	if code, o := h.api(http.MethodPost, "/bots/bulk", map[string]any{"count": 1, "prefix": "x", "rooms": []map[string]any{{"game": "tst_board", "room": "nope"}}}); code != 400 || o["error"] != "ROOM_NOT_FOUND" {
		t.Fatalf("bad room %d %v", code, o)
	}
	sfx := randomSuffix()
	csvText := "username,level,template,coin,rooms,look.avatar\nکامران_موسوی" + sfx + ",7,hard,500,tst_board:r2,avatar-boy\nStormKing" + sfx + ",30,,0,,\nbad name,3,,,,\n"
	dry := h.mustAPI(200, http.MethodPost, "/bots/import", map[string]any{"csv": csvText, "dryRun": true})
	if dry["valid"] != float64(2) || len(dry["problems"].([]any)) != 1 {
		t.Fatalf("dry run %v", dry)
	}
	imp := h.mustAPI(200, http.MethodPost, "/bots/import", map[string]any{"csv": csvText, "template": "normal"})
	if imp["created"] != float64(2) {
		t.Fatalf("import %v", imp)
	}
	again := h.mustAPI(200, http.MethodPost, "/bots/import", map[string]any{"csv": "username\nStormKing" + sfx + "\n"})
	if again["created"] != float64(0) || again["problems"].([]any)[0].(map[string]any)["code"] != "USERNAME_TAKEN" {
		t.Fatalf("duplicate %v", again)
	}
	list := h.mustAPI(200, http.MethodGet, "/bots?q=کامران_موسوی"+sfx, nil)
	items := list["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("search %v", list)
	}
	k := items[0].(map[string]any)
	if k["template"] != "hard" || k["level"] != float64(7) || k["look"].(map[string]any)["avatar"] != "avatar-boy" || k["wallet"].(map[string]any)["coin"] != float64(500) {
		t.Fatalf("imported bot %v", k)
	}
	id := k["userId"].(string)
	if l := h.mustAPI(200, http.MethodGet, "/bots?game=tst_board&room=r1", nil); l["total"] != float64(6) {
		t.Fatalf("bots for r1 = bound plus unbound: %v", l["total"])
	}
	h.mustAPI(200, http.MethodPut, "/bots/"+id, map[string]any{"level": 12, "note": "veteran", "wallet": map[string]int64{"coin": 900}})
	got := h.mustAPI(200, http.MethodGet, "/bots/"+id, nil)
	if got["level"] != float64(12) || got["note"] != "veteran" || got["wallet"].(map[string]any)["coin"] != float64(900) {
		t.Fatalf("put %v", got)
	}
	st := h.mustAPI(200, http.MethodPost, "/bots/status", map[string]any{"ids": []string{id}, "enabled": false})
	if st["updated"] != float64(1) {
		t.Fatalf("status %v", st)
	}
	picked, _ := gameProvider{s: h.s}.Bots(h.ctx, gamecatalog.BotQuery{Game: "tst_board", Room: "r2", Count: 10})
	for _, b := range picked {
		if b.UserID == id {
			t.Fatal("a disabled bot is not picked")
		}
	}
	tu := h.mustAPI(200, http.MethodPost, "/bots/topup", map[string]any{"all": true, "currency": "coin", "below": 1000, "amount": 5000})
	if tu["toppedUp"] != float64(2) {
		t.Fatalf("topup %v", tu)
	}
}
