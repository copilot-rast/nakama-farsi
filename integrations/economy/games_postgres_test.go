package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

const gameTestSchema = `{"type":"object","additionalProperties":false,"properties":{
	"roundSeconds":{"type":"integer","minimum":10,"maximum":120,"default":30,"title":{"fa":"زمان دور","en":"Round time"}},
	"board":{"type":"string","enum":["classic","mini"],"default":"classic"}}}`

func (n *fakeNakama) MatchCreate(_ context.Context, module string, params map[string]any) (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls++
	raw, _ := json.Marshal(params)
	return module + "." + string(raw), nil
}

func newGameHarness(t *testing.T) *harness {
	h := newEventHarness(t)
	for _, q := range []string{
		`TRUNCATE economy_game_entry, economy_game_match, economy_game, economy_chance_play, economy_chance_state, economy_chance_game, economy_prize_group_member, economy_prize_group CASCADE`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for name, fn := range h.s.gameRPCs() {
		h.rpcs[name] = fn
	}
	h.s.games.ttl = time.Nanosecond
	h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "level", "name": map[string]string{"fa": "سطح"}, "defaultValue": 1})
	if err := gamecatalog.Register(gamecatalog.Definition{Key: "tst_board", Realtime: true, Plugin: "test", Schema: json.RawMessage(gameTestSchema)}); err != nil {
		t.Fatal(err)
	}
	if err := gamecatalog.Register(gamecatalog.Definition{Key: "tst_solo", Plugin: "test"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		gamecatalog.Unregister("tst_board")
		gamecatalog.Unregister("tst_solo")
	})
	return h
}

func (h *harness) setLevel(user string, level int) {
	h.t.Helper()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) SELECT $1, id, $2 FROM economy_meta WHERE key='level'
		ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value`, user, level); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) boardGame(extra map[string]any) map[string]any {
	h.t.Helper()
	prize := h.coinPrize("tst_win_"+randomSuffix(), 30, true)
	body := map[string]any{"key": "tst_board", "isActive": true, "name": map[string]string{"fa": "بازی تخته", "en": "Board game"},
		"maintenance": map[string]string{"fa": "در حال تعمیر", "en": "Under repair"}, "minPlayers": 2, "maxPlayers": 2, "turnSeconds": 20,
		"rooms": []map[string]any{
			{"key": "bronze", "name": map[string]string{"fa": "برنزی", "en": "Bronze"}, "entryFee": map[string]any{"currencyKey": "coin", "amount": 10}, "prizeId": prize},
			{"key": "gold", "name": map[string]string{"fa": "طلایی", "en": "Gold"}, "entryFee": map[string]any{"currencyKey": "coin", "amount": 50}, "prizeId": prize, "minLevel": 3},
		}}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/games", body)
}

func randomSuffix() string {
	return strings.ReplaceAll(time.Now().Format("150405.000000"), ".", "")
}

func (h *harness) ledgerRows(user, reason string) int {
	return h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'reason'=$2`, user, reason)
}

func TestGameSettingsValidationAndRegistry(t *testing.T) {
	h := newGameHarness(t)
	once := h.coinPrize("tst_once", 5, false)
	reg := h.mustAPI(200, http.MethodGet, "/games-registry", nil)
	if len(reg["unconfigured"].([]any)) < 2 {
		t.Fatalf("registered games without a row must be listed: %v", reg["unconfigured"])
	}
	base := func(extra map[string]any) map[string]any {
		b := map[string]any{"key": "tst_board", "name": map[string]string{"fa": "تخته"}, "minPlayers": 2, "maxPlayers": 4}
		for k, v := range extra {
			b[k] = v
		}
		return b
	}
	for _, tc := range []struct {
		body  map[string]any
		code  string
		field string
	}{
		{base(map[string]any{"key": "Bad-Key"}), "INVALID_GAME_KEY", "key"},
		{base(map[string]any{"maxPlayers": 1}), "MAX_BELOW_MIN", "maxPlayers"},
		{base(map[string]any{"settings": map[string]any{"roundSeconds": 5}}), "INVALID_SETTING", "settings.roundSeconds"},
		{base(map[string]any{"settings": map[string]any{"board": "huge"}}), "INVALID_SETTING", "settings.board"},
		{base(map[string]any{"settings": map[string]any{"colour": "red"}}), "UNKNOWN_SETTING", "settings.colour"},
		{base(map[string]any{"rooms": []map[string]any{{"key": "a1"}, {"key": "a1"}}}), "DUPLICATE_ROOM", "rooms[1]"},
		{base(map[string]any{"rooms": []map[string]any{{"key": "a1", "entryFee": map[string]any{"currencyKey": "coin"}}}}), "INVALID_ENTRY_FEE", "rooms[0]"},
		{base(map[string]any{"rooms": []map[string]any{{"key": "a1", "entryFee": map[string]any{"currencyKey": "gem", "amount": 5}}}}), "CURRENCY_NOT_FOUND", "rooms[0]"},
		{base(map[string]any{"rooms": []map[string]any{{"key": "a1", "prizeId": once}}}), "PRIZE_NOT_REPEATABLE", "rooms[0]"},
		{base(map[string]any{"leaderboardId": "nope_board"}), "LEADERBOARD_NOT_FOUND", "leaderboardId"},
		{base(map[string]any{"platforms": []string{"psp"}}), "INVALID_PLATFORM", "platforms"},
		{base(map[string]any{"minVersion": "2.0.0", "maxVersion": "1.0.0"}), "VERSION_RANGE", "maxVersion"},
	} {
		code, out := h.api(http.MethodPost, "/games", tc.body)
		if code != 400 || out["error"] != tc.code || out["field"] != tc.field {
			t.Fatalf("%s: %d %v", tc.code, code, out)
		}
	}
	g := h.mustAPI(201, http.MethodPost, "/games", base(map[string]any{"settings": map[string]any{"roundSeconds": 45}, "platforms": []string{"web", "android"}}))
	if g["settings"].(map[string]any)["board"] != "classic" || g["settings"].(map[string]any)["roundSeconds"] != float64(45) {
		t.Fatalf("defaults from the schema must fill missing settings: %v", g["settings"])
	}
	if code, out := h.api(http.MethodPut, "/games/"+g["id"].(string), base(map[string]any{"key": "tst_other"})); code != 400 || out["error"] != "KEY_IMMUTABLE" {
		t.Fatalf("key change: %d %v", code, out)
	}
	h.mustAPI(201, http.MethodPost, "/games", map[string]any{"key": "tst_gone", "name": map[string]string{"fa": "رفته"}})
	reg = h.mustAPI(200, http.MethodGet, "/games-registry", nil)
	missing := map[string]bool{}
	for _, m := range reg["missingHandler"].([]any) {
		missing[m.(map[string]any)["key"].(string)] = true
	}
	if !missing["tst_gone"] || missing["tst_board"] {
		t.Fatalf("missing handler list %v", reg["missingHandler"])
	}
	for _, it := range reg["items"].([]any) {
		m := it.(map[string]any)
		if m["key"] == "tst_board" && (m["configured"] != true || m["schema"] == nil || m["realtime"] != true) {
			t.Fatalf("registry row %v", m)
		}
	}
}

func TestGameListAndJoinEligibility(t *testing.T) {
	h := newGameHarness(t)
	g := h.boardGame(map[string]any{"platforms": []string{"web"}, "minVersion": "1.2.0", "settings": map[string]any{"roundSeconds": 40}})
	h.mustAPI(201, http.MethodPost, "/games", map[string]any{"key": "tst_solo", "isActive": false, "name": map[string]string{"fa": "تنها"}})
	user := h.user()
	h.give(user, 30)

	list := h.mustRPC(user, "economy_games", map[string]any{"lang": "en", "platform": "web", "version": "1.4.0"})
	items := list["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("a disabled game without a maintenance message stays hidden: %v", items)
	}
	view := items[0].(map[string]any)
	rooms := view["rooms"].([]any)
	if view["name"] != "Board game" || view["realtime"] != true || len(rooms) != 2 || rooms[0].(map[string]any)["name"] != "Bronze" ||
		rooms[1].(map[string]any)["locked"] != true || rooms[0].(map[string]any)["entryFee"].(map[string]any)["currencyName"] != "Coin" {
		t.Fatalf("view %v", view)
	}
	if view["settings"].(map[string]any)["roundSeconds"] != float64(40) {
		t.Fatalf("settings %v", view["settings"])
	}
	if n := len(h.mustRPC(user, "economy_games", map[string]any{"platform": "android"})["items"].([]any)); n != 0 {
		t.Fatalf("android must not see a web-only game: %d", n)
	}
	if n := len(h.mustRPC(user, "economy_games", map[string]any{"platform": "web", "version": "1.0.0"})["items"].([]any)); n != 0 {
		t.Fatalf("an old version must not see the game: %d", n)
	}
	if code := h.rpcCode(user, "economy_game", map[string]any{"key": "tst_board", "platform": "web", "version": "1.0.0"}); code != "UPDATE_REQUIRED" {
		t.Fatal(code)
	}
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_board", "platform": "web"}); code != "ROOM_REQUIRED" {
		t.Fatal(code)
	}
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "diamond"}); code != "ROOM_NOT_FOUND" {
		t.Fatal(code)
	}
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "gold"}); code != "LEVEL_TOO_LOW" {
		t.Fatal(code)
	}
	h.setLevel(user, 3)
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "gold"}); code != "INSUFFICIENT_BALANCE" {
		t.Fatal(code)
	}
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "bronze", "platform": "ios"}); code != "PLATFORM_NOT_SUPPORTED" {
		t.Fatal(code)
	}
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_solo"}); code != "GAME_DISABLED" {
		t.Fatal(code)
	}
	ticket := h.mustRPC(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "bronze", "platform": "web", "version": "1.2.0"})
	if ticket["query"] != "+properties.game:tst_board +properties.room:bronze" || ticket["minCount"] != float64(2) || ticket["maxCount"] != float64(2) ||
		ticket["stringProperties"].(map[string]any)["room"] != "bronze" || ticket["numericProperties"].(map[string]any)["level"] != float64(3) {
		t.Fatalf("ticket %v", ticket)
	}

	ctx := context.WithValue(h.ctx, runtime.RUNTIME_CTX_USER_ID, user)
	env := &rtapi.Envelope{Message: &rtapi.Envelope_MatchmakerAdd{MatchmakerAdd: &rtapi.MatchmakerAdd{Query: "*", MinCount: 2, MaxCount: 8,
		StringProperties: map[string]string{"game": "tst_board", "room": "bronze", "platform": "web"}}}}
	out, err := h.s.beforeMatchmakerAdd(ctx, nil, nil, nil, env)
	if err != nil {
		t.Fatal(err)
	}
	m := out.GetMatchmakerAdd()
	if m.Query != "+properties.game:tst_board +properties.room:bronze" || m.MaxCount != 2 || m.MinCount != 2 || m.NumericProperties["level"] != 3 {
		t.Fatalf("the hook must force the game's ticket: %+v", m)
	}
	env.GetMatchmakerAdd().StringProperties["room"] = "gold"
	if _, err := h.s.beforeMatchmakerAdd(ctx, nil, nil, nil, env); err == nil || !strings.Contains(err.Error(), "INSUFFICIENT_BALANCE") {
		t.Fatalf("hook err %v", err)
	}
	plain := &rtapi.Envelope{Message: &rtapi.Envelope_MatchmakerAdd{MatchmakerAdd: &rtapi.MatchmakerAdd{Query: "*", MinCount: 2, MaxCount: 4}}}
	if out, err := h.s.beforeMatchmakerAdd(ctx, nil, nil, nil, plain); err != nil || out.GetMatchmakerAdd().Query != "*" {
		t.Fatalf("a ticket without a game is left alone: %v %v", out, err)
	}

	h.mustAPI(200, http.MethodPut, "/games/"+g["id"].(string), map[string]any{"key": "tst_board", "isActive": false, "name": map[string]string{"fa": "بازی تخته", "en": "Board game"},
		"maintenance": map[string]string{"fa": "در حال تعمیر", "en": "Under repair"}, "minPlayers": 2, "maxPlayers": 2})
	off := h.mustRPC(user, "economy_games", map[string]any{"lang": "fa"})["items"].([]any)
	if len(off) != 1 || off[0].(map[string]any)["isActive"] != false || off[0].(map[string]any)["maintenance"] != "در حال تعمیر" {
		t.Fatalf("a switched-off game shows its maintenance message: %v", off)
	}
	_, err = h.rpc(user, "economy_game_join", map[string]any{"key": "tst_board", "lang": "en"})
	if app, ok := asApp(err); !ok || app.Code != "GAME_DISABLED" || app.Meta["message"] != "Under repair" {
		t.Fatalf("disabled join %v", err)
	}
}

func TestGameEntryFeeOncePerMatchUnderParallelJoins(t *testing.T) {
	h := newGameHarness(t)
	h.boardGame(nil)
	a, b, poor := h.user(), h.user(), h.user()
	h.give(a, 100)
	h.give(b, 100)
	h.give(poor, 5)
	p := gameProvider{s: h.s}
	match := "m-" + randomSuffix()
	var wg sync.WaitGroup
	results := make(chan gamecatalog.EntryResult, 20)
	errs := make(chan error, 20)
	for i := 0; i < 10; i++ {
		for _, u := range []string{a, b} {
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				node := testService()
				node.db, node.nk = h.s.db, h.nk
				r, err := gameProvider{s: node}.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "bronze", UserID: u})
				if err != nil {
					errs <- err
					return
				}
				results <- r
			}(u)
		}
	}
	wg.Wait()
	close(errs)
	close(results)
	for err := range errs {
		t.Fatal(err)
	}
	charged := 0
	for r := range results {
		if r.Charged {
			charged++
		}
	}
	if charged != 2 || h.wallet(a)["coin"] != 90 || h.wallet(b)["coin"] != 90 {
		t.Fatalf("charged=%d a=%d b=%d", charged, h.wallet(a)["coin"], h.wallet(b)["coin"])
	}
	if n := h.ledgerRows(a, gameEntryReason); n != 1 {
		t.Fatalf("one debit row per player, got %d", n)
	}
	if n := h.count(`SELECT players FROM economy_game_match WHERE id=$1`, match); n != 2 {
		t.Fatalf("players %d", n)
	}
	_, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "bronze", UserID: poor})
	if gamecatalog.Code(err) != "INSUFFICIENT_BALANCE" || h.count(`SELECT count(*) FROM economy_game_entry WHERE match_id=$1 AND user_id=$2`, match, poor) != 0 {
		t.Fatalf("a player who cannot pay leaves no entry: %v", err)
	}
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "gold", UserID: a}); gamecatalog.Code(err) != "LEVEL_TOO_LOW" {
		t.Fatalf("err %v", err)
	}
	h.setLevel(a, 5)
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "gold", UserID: a}); gamecatalog.Code(err) != "MATCH_MISMATCH" {
		t.Fatalf("err %v", err)
	}
	if err := p.Start(h.ctx, match); err != nil {
		t.Fatal(err)
	}
	if r, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "bronze", UserID: a}); err != nil || !r.Rejoined || r.Charged {
		t.Fatalf("a rejoin after the start is free: %+v %v", r, err)
	}
	c := h.user()
	h.give(c, 100)
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "bronze", UserID: c}); gamecatalog.Code(err) != "MATCH_STARTED" {
		t.Fatalf("err %v", err)
	}

	var pw sync.WaitGroup
	for i := 0; i < 6; i++ {
		pw.Add(1)
		go func() {
			defer pw.Done()
			if _, err := p.Finish(h.ctx, match, []string{a}); err != nil {
				t.Error(err)
			}
		}()
	}
	pw.Wait()
	if h.wallet(a)["coin"] != 120 || h.wallet(b)["coin"] != 90 {
		t.Fatalf("the prize is paid once: a=%d b=%d", h.wallet(a)["coin"], h.wallet(b)["coin"])
	}
	if n := h.count(`SELECT count(*) FROM economy_user_prize WHERE source_type='game' AND user_id=$1 AND claimed`, a); n != 1 {
		t.Fatalf("prize rows %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_game_entry WHERE match_id=$1 AND prize_status='paid'`, match); n != 1 {
		t.Fatalf("paid entries %d", n)
	}
	if _, err := p.Cancel(h.ctx, match); gamecatalog.Code(err) != "MATCH_STARTED" {
		t.Fatalf("a finished match cannot be cancelled: %v", err)
	}
}

func TestGameRefundBeforeStart(t *testing.T) {
	h := newGameHarness(t)
	h.boardGame(nil)
	a, b := h.user(), h.user()
	h.give(a, 100)
	h.give(b, 100)
	p := gameProvider{s: h.s}
	match := "r-" + randomSuffix()
	for _, u := range []string{a, b} {
		if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "bronze", UserID: u}); err != nil {
			t.Fatal(err)
		}
	}
	if ok, err := p.Leave(h.ctx, match, a); err != nil || !ok || h.wallet(a)["coin"] != 100 {
		t.Fatalf("leaving before the start refunds: %v %v %d", ok, err, h.wallet(a)["coin"])
	}
	if ok, _ := p.Leave(h.ctx, match, a); ok || h.wallet(a)["coin"] != 100 {
		t.Fatal("a second leave refunds nothing")
	}
	if r, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "bronze", UserID: a}); err != nil || !r.Charged || h.wallet(a)["coin"] != 90 {
		t.Fatalf("joining again charges again: %+v %v %d", r, err, h.wallet(a)["coin"])
	}
	var wg sync.WaitGroup
	total := 0
	var mu sync.Mutex
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := p.Cancel(h.ctx, match)
			if err != nil {
				t.Error(err)
			}
			mu.Lock()
			total += n
			mu.Unlock()
		}()
	}
	wg.Wait()
	if total != 2 || h.wallet(a)["coin"] != 100 || h.wallet(b)["coin"] != 100 {
		t.Fatalf("cancel refunds each paid entry once: total=%d a=%d b=%d", total, h.wallet(a)["coin"], h.wallet(b)["coin"])
	}
	if n := h.ledgerRows(a, gameRefundReason); n != 2 {
		t.Fatalf("refund rows for a: %d", n)
	}
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "bronze", UserID: b}); gamecatalog.Code(err) != "MATCH_CANCELLED" {
		t.Fatalf("err %v", err)
	}
	if err := p.Start(h.ctx, match); gamecatalog.Code(err) != "MATCH_CANCELLED" {
		t.Fatalf("err %v", err)
	}

	started := "s-" + randomSuffix()
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: started, Game: "tst_board", Room: "bronze", UserID: a}); err != nil {
		t.Fatal(err)
	}
	if err := p.Start(h.ctx, started); err != nil {
		t.Fatal(err)
	}
	if ok, err := p.Leave(h.ctx, started, a); err != nil || ok || h.wallet(a)["coin"] != 90 {
		t.Fatalf("leaving after the start forfeits the fee: %v %v", ok, err)
	}

	old := "o-" + randomSuffix()
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: old, Game: "tst_board", Room: "bronze", UserID: b}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_game_match SET created_at=now() - interval '1 hour' WHERE id=$1`, old); err != nil {
		t.Fatal(err)
	}
	if n := h.s.expireGameMatches(h.ctx, 10); n != 1 || h.wallet(b)["coin"] != 100 {
		t.Fatalf("an open match left behind is cancelled and refunded: %d %d", n, h.wallet(b)["coin"])
	}
	detail := h.mustAPI(200, http.MethodGet, "/game-matches/"+old, nil)
	if detail["status"] != "cancelled" || len(detail["entries"].([]any)) != 1 {
		t.Fatalf("detail %v", detail)
	}
	list := h.mustAPI(200, http.MethodGet, "/game-matches?gameKey=tst_board&status=cancelled", nil)
	if list["total"] != float64(2) {
		t.Fatalf("cancelled list %v", list["total"])
	}
}

func TestGamePrizeGroupAndMatchmakerMatched(t *testing.T) {
	h := newGameHarness(t)
	p1 := h.coinPrize("tst_g1", 7, true)
	group := h.mustAPI(201, http.MethodPost, "/prize-groups", map[string]any{"key": "tst_pg", "name": map[string]string{"fa": "گروه"}, "members": []map[string]any{{"prizeId": p1, "weight": 1}}})
	h.mustAPI(201, http.MethodPost, "/games", map[string]any{"key": "tst_board", "isActive": true, "name": map[string]string{"fa": "تخته"}, "minPlayers": 2, "maxPlayers": 2,
		"rooms": []map[string]any{{"key": "free", "prizeGroupId": group["id"]}}})
	a, b := h.user(), h.user()
	p := gameProvider{s: h.s}
	match := "g-" + randomSuffix()
	for _, u := range []string{a, b} {
		r, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "free", UserID: u})
		if err != nil || r.Status != "free" || r.Charged {
			t.Fatalf("free room %+v %v", r, err)
		}
	}
	out, err := p.Finish(h.ctx, match, []string{b, "not-a-uuid"})
	if err != nil || len(out) != 1 || out[0].PrizeID != p1 || h.wallet(b)["coin"] != 7 {
		t.Fatalf("group prize %+v %v %d", out, err, h.wallet(b)["coin"])
	}
	again, err := p.Finish(h.ctx, match, []string{a})
	if err != nil || len(again) != 1 || again[0].UserID != b || h.wallet(a)["coin"] != 0 || h.wallet(b)["coin"] != 7 {
		t.Fatalf("a second finish keeps the first winners: %+v %v", again, err)
	}

	id, err := h.s.matchmakerMatched(h.ctx, nil, nil, h.nk, []runtime.MatchmakerEntry{
		&fakeEntry{user: a, props: map[string]any{"game": "tst_board", "room": "free"}}, &fakeEntry{user: b, props: map[string]any{"game": "tst_board", "room": "free"}}})
	if err != nil || !strings.HasPrefix(id, "tst_board.") || !strings.Contains(id, a+","+b) {
		t.Fatalf("matched %q %v", id, err)
	}
	if id, _ := h.s.matchmakerMatched(h.ctx, nil, nil, h.nk, []runtime.MatchmakerEntry{&fakeEntry{user: a, props: map[string]any{"game": "tst_solo"}}}); id != "" {
		t.Fatalf("a game without a match handler stays a relayed match: %q", id)
	}
	if id, _ := h.s.matchmakerMatched(h.ctx, nil, nil, h.nk, []runtime.MatchmakerEntry{&fakeEntry{user: a, props: map[string]any{}}}); id != "" {
		t.Fatalf("a ticket without a game is left alone: %q", id)
	}
}

func TestGameSettingsCacheAcrossNodes(t *testing.T) {
	h := newGameHarness(t)
	g := h.boardGame(map[string]any{"settings": map[string]any{"roundSeconds": 30}})
	clock := time.Now()
	nodeB := testService()
	nodeB.db, nodeB.nk = h.s.db, h.nk
	nodeB.now = func() time.Time { return clock }
	read := func() int {
		got, err := gameProvider{s: nodeB}.Game(h.ctx, "tst_board")
		if err != nil {
			t.Fatal(err)
		}
		return got.Int("roundSeconds", 0)
	}
	if v := read(); v != 30 {
		t.Fatalf("first read %d", v)
	}
	body := map[string]any{"key": "tst_board", "isActive": true, "name": map[string]string{"fa": "بازی تخته"}, "minPlayers": 2, "maxPlayers": 2, "settings": map[string]any{"roundSeconds": 50}}
	h.mustAPI(200, http.MethodPut, "/games/"+g["id"].(string), body)
	if v := read(); v != 30 {
		t.Fatalf("inside the cache window node B keeps its copy: %d", v)
	}
	clock = clock.Add(3 * time.Second)
	if v := read(); v != 50 {
		t.Fatalf("after the window node B sees the change from node A: %d", v)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_game SET settings='{"roundSeconds":70}' WHERE key='tst_board'`); err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(3 * time.Second)
	if v := read(); v != 70 {
		t.Fatalf("a direct database change bumps the version too: %d", v)
	}
	var before int64
	_ = h.s.db.QueryRow(`SELECT version FROM economy_game_version WHERE id=1`).Scan(&before)
	clock = clock.Add(3 * time.Second)
	read()
	if nodeB.games.version != before {
		t.Fatalf("node B tracks the version %d vs %d", nodeB.games.version, before)
	}
	if _, err := (gameProvider{s: nodeB}).Game(h.ctx, "tst_missing"); err != gamecatalog.ErrNotFound {
		t.Fatalf("err %v", err)
	}
	gamecatalog.SetProvider(gameProvider{s: nodeB})
	defer gamecatalog.SetProvider(nil)
	if got, err := gamecatalog.Settings(h.ctx, "tst_board"); err != nil || got.Int("roundSeconds", 0) != 70 || len(got.Rooms) != 0 {
		t.Fatalf("helper %v %v", got, err)
	}
}

type fakeEntry struct {
	runtime.MatchmakerEntry
	user  string
	props map[string]any
}

func (e *fakeEntry) GetProperties() map[string]any { return e.props }
func (e *fakeEntry) GetPresence() runtime.Presence { return &fakePresence{user: e.user} }

type fakePresence struct {
	runtime.Presence
	user string
}

func (p *fakePresence) GetUserId() string { return p.user }
