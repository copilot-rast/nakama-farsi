package main

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

type reqWorld struct {
	h                       *harness
	pass, vipPass, starter  string
	vip, cheat              string
	ticket, xp              string
	passProduct, tktProduct string
}

func newReqWorld(t *testing.T) *reqWorld {
	h := newGameHarness(t)
	for name, fn := range h.s.itemRPCs() {
		h.rpcs[name] = fn
	}
	w := &reqWorld{h: h}
	skin := h.taxonomy("base_category", "item_type", "pass", map[string]string{"fa": "مجوز", "en": "Pass"})["id"].(string)
	w.pass = h.item(map[string]any{"key": "arena_pass", "typeId": skin, "name": map[string]string{"fa": "مجوز آرنا", "en": "Arena pass"}})["id"].(string)
	w.vipPass = h.item(map[string]any{"key": "vip_pass", "typeId": skin})["id"].(string)
	w.starter = h.item(map[string]any{"key": "starter_hat", "typeId": skin, "isStarter": true})["id"].(string)
	w.vip = h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه", "en": "VIP"})["id"].(string)
	w.cheat = h.taxonomy("base_tag", "user", "cheater", map[string]string{"fa": "متقلب", "en": "Cheater"})["id"].(string)
	w.ticket = h.meta("ticket", 0)
	w.xp = h.meta("xp", 0)
	passPrize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "pz_pass", "name": map[string]string{"fa": "مجوز"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "item", "key": "arena_pass", "action": "give", "value": 1}}})["id"].(string)
	tktPrize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "pz_tkt", "name": map[string]string{"fa": "بلیت"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "meta", "key": "ticket", "action": "add", "value": 3}}})["id"].(string)
	w.passProduct = h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "buy_pass", "isActive": true, "name": map[string]string{"fa": "خرید مجوز", "en": "Buy pass"}, "prizeId": passPrize, "paymentMethods": []string{"zibal"}, "priceToman": 50000})["id"].(string)
	w.tktProduct = h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "buy_tickets", "isActive": true, "name": map[string]string{"fa": "سه بلیت", "en": "Three tickets"}, "prizeId": tktPrize, "paymentMethods": []string{"zibal"}, "priceToman": 20000})["id"].(string)
	return w
}

func (w *reqWorld) tagUser(user, tag string) {
	w.h.t.Helper()
	if _, err := w.h.s.db.ExecContext(w.h.ctx, `INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, user, tag); err != nil {
		w.h.t.Fatal(err)
	}
}

func (w *reqWorld) setMeta(user, id string, v float64) {
	w.h.t.Helper()
	if _, err := w.h.s.db.ExecContext(w.h.ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ($1,$2,$3) ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value`, user, id, v); err != nil {
		w.h.t.Fatal(err)
	}
}

func (w *reqWorld) metaOf(user, id string) float64 {
	var v float64
	_ = w.h.s.db.QueryRowContext(w.h.ctx, `SELECT COALESCE((SELECT value::float8 FROM economy_user_meta WHERE user_id=$1 AND meta_id=$2),0)`, user, id).Scan(&v)
	return v
}

func (w *reqWorld) owns(user, item string) bool {
	return w.h.count(`SELECT count(*) FROM economy_user_item WHERE user_id=$1 AND item_id=$2`, user, item) == 1
}

func (w *reqWorld) giveItem(user, item string) {
	w.h.t.Helper()
	if _, err := w.h.s.db.ExecContext(w.h.ctx, `INSERT INTO economy_user_item (user_id, item_id, source) VALUES ($1,$2,'buy') ON CONFLICT DO NOTHING`, user, item); err != nil {
		w.h.t.Fatal(err)
	}
}

func (w *reqWorld) game(gameReq map[string]any, rooms []map[string]any) map[string]any {
	w.h.t.Helper()
	body := map[string]any{"key": "tst_board", "isActive": true, "name": map[string]string{"fa": "بازی تخته", "en": "Board game"}, "minPlayers": 2, "maxPlayers": 2, "rooms": rooms}
	if gameReq != nil {
		body["requirements"] = gameReq
	}
	return w.h.mustAPI(201, http.MethodPost, "/games", body)
}

func roomByKey(view map[string]any, key string) map[string]any {
	for _, r := range view["rooms"].([]any) {
		if r.(map[string]any)["key"] == key {
			return r.(map[string]any)
		}
	}
	return nil
}

func TestGameRequirementsValidation(t *testing.T) {
	w := newReqWorld(t)
	h := w.h
	base := func(req map[string]any) map[string]any {
		return map[string]any{"key": "tst_board", "name": map[string]string{"fa": "تخته"}, "minPlayers": 2, "maxPlayers": 2,
			"rooms": []map[string]any{{"key": "arena", "requirements": req}}}
	}
	for _, tc := range []struct {
		req   map[string]any
		code  string
		field string
	}{
		{map[string]any{"itemIds": []string{"00000000-0000-0000-0000-000000000001"}}, "ITEM_NOT_FOUND", "rooms[0].requirements.itemIds"},
		{map[string]any{"itemIds": []string{w.pass}, "itemsMode": "most"}, "INVALID_REQUIREMENT", "rooms[0].requirements.itemsMode"},
		{map[string]any{"tagIds": []string{w.vip}, "excludeTagIds": []string{w.vip}}, "TAG_BOTH_WAYS", "rooms[0].requirements.excludeTagIds"},
		{map[string]any{"metas": []map[string]any{{"metaId": w.ticket, "operator": "about", "value": 1}}}, "INVALID_OPERATOR", "rooms[0].requirements.metas[0].operator"},
		{map[string]any{"metas": []map[string]any{{"metaId": w.ticket, "value": 1, "consume": -1}}}, "INVALID_CONSUME", "rooms[0].requirements.metas[0].consume"},
		{map[string]any{"itemIds": []string{w.starter}, "consumeItems": true}, "STARTER_TICKET", "rooms[0].requirements.consumeItems"},
		{map[string]any{"itemIds": []string{w.pass}, "paths": []string{"arena"}}, "INVALID_PATH", "rooms[0].requirements.paths"},
		{map[string]any{"itemIds": []string{w.pass}, "refund": "maybe"}, "INVALID_REQUIREMENT", "rooms[0].requirements.refund"},
	} {
		code, out := h.api(http.MethodPost, "/games", base(tc.req))
		if code != 400 || out["error"] != tc.code || out["field"] != tc.field {
			t.Fatalf("%s: %d %v", tc.code, code, out)
		}
	}
	code, out := h.api(http.MethodPost, "/games", map[string]any{"key": "tst_board", "name": map[string]string{"fa": "تخته"}, "requirements": map[string]any{"tagIds": []string{w.pass}}})
	if code != 400 || out["error"] != "TAG_NOT_FOUND" || out["field"] != "requirements.tagIds" {
		t.Fatalf("game-level tag %d %v", code, out)
	}
	g := w.game(map[string]any{"minLevel": 2, "paths": []string{"matchmaking", "bot", "friend"}}, []map[string]any{
		{"key": "arena", "requirements": map[string]any{"itemIds": []string{w.pass, w.pass}, "consumeItems": true, "paths": []string{"bot"}, "inherit": false, "message": map[string]string{"fa": "مجوز لازم است"}}},
		{"key": "open", "requirements": map[string]any{}},
	})
	req := g["requirements"].(map[string]any)
	if req["minLevel"] != float64(2) || req["paths"] != nil || req["itemsMode"] != "any" || req["lock"] != "show" {
		t.Fatalf("all paths are stored as none and defaults are filled: %v", req)
	}
	rooms := g["rooms"].([]any)
	arena := rooms[0].(map[string]any)["requirements"].(map[string]any)
	if len(arena["itemIds"].([]any)) != 1 || arena["consumeItems"] != true || arena["inherit"] != false || arena["paths"].([]any)[0] != "bot" {
		t.Fatalf("room set %v", arena)
	}
	if _, ok := rooms[1].(map[string]any)["requirements"]; ok {
		t.Fatalf("an empty room set is dropped: %v", rooms[1])
	}
	if c, out := h.api(http.MethodDelete, "/items/"+w.pass, nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("an item used by a room cannot be deleted: %d %v", c, out)
	}
}

func TestGameRequirementsListJoinAndRefusal(t *testing.T) {
	w := newReqWorld(t)
	h := w.h
	w.game(map[string]any{"excludeTagIds": []string{w.cheat}}, []map[string]any{
		{"key": "open"},
		{"key": "arena", "requirements": map[string]any{"itemIds": []string{w.pass, w.vipPass}, "message": map[string]string{"fa": "اول مجوز آرنا بگیر", "en": "Get an arena pass first"}}},
		{"key": "vip", "requirements": map[string]any{"tagIds": []string{w.vip}, "lock": "hide"}},
		{"key": "pro", "requirements": map[string]any{"metas": []map[string]any{{"metaId": w.xp, "operator": "gte", "value": 50}}, "minLevel": 3}},
		{"key": "bots", "requirements": map[string]any{"itemIds": []string{w.pass}, "paths": []string{"bot"}}},
	})
	user := h.user()
	list := h.mustRPC(user, "economy_games", map[string]any{"lang": "en"})
	view := list["items"].([]any)[0].(map[string]any)
	if roomByKey(view, "vip") != nil || len(view["rooms"].([]any)) != 4 {
		t.Fatalf("a hidden room the player cannot enter is left out: %v", view["rooms"])
	}
	arena := roomByKey(view, "arena")
	if arena["eligible"] != false || arena["locked"] != true || arena["levelLocked"] != false {
		t.Fatalf("arena %v", arena)
	}
	miss := arena["missing"].([]any)[0].(map[string]any)
	if miss["type"] != "item" || miss["mode"] != "any" || len(miss["items"].([]any)) != 2 || miss["items"].([]any)[0].(map[string]any)["name"] != "Arena pass" {
		t.Fatalf("missing %v", miss)
	}
	prods := arena["products"].([]any)
	if len(prods) != 1 || prods[0].(map[string]any)["key"] != "buy_pass" || prods[0].(map[string]any)["satisfies"].([]any)[0] != "item:arena_pass" {
		t.Fatalf("products found from the prize contents: %v", prods)
	}
	reqView := arena["requirements"].(map[string]any)
	if reqView["items"].(map[string]any)["list"].([]any)[0].(map[string]any)["owned"] != false || reqView["consumes"] != false {
		t.Fatalf("requirements view %v", reqView)
	}
	bots := roomByKey(view, "bots")
	if bots["eligible"] != true || bots["eligibleBy"].(map[string]any)["bot"] != false || bots["eligibleBy"].(map[string]any)["matchmaking"] != true {
		t.Fatalf("a bot-only set does not lock matchmaking: %v", bots)
	}
	pro := roomByKey(view, "pro")
	if len(pro["missing"].([]any)) != 2 {
		t.Fatalf("level and meta: %v", pro["missing"])
	}
	_, err := h.rpc(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "arena", "lang": "en"})
	app, ok := asApp(err)
	if !ok || app.Code != codeRequirementsNotMet || app.Status != 403 || app.Meta["message"] != "Get an arena pass first" || app.Meta["room"] != "arena" || app.Meta["path"] != "matchmaking" {
		t.Fatalf("join refusal %v", err)
	}
	if len(app.Meta["products"].([]map[string]any)) != 1 {
		t.Fatalf("refusal products %v", app.Meta["products"])
	}
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "bots", "path": "bot"}); code != codeRequirementsNotMet {
		t.Fatal(code)
	}
	if code := h.rpcCode(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "bots", "path": "arena"}); code != "INVALID_PATH" {
		t.Fatal(code)
	}
	h.mustRPC(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "bots"})
	ctx := context.WithValue(h.ctx, runtime.RUNTIME_CTX_USER_ID, user)
	env := &rtapi.Envelope{Message: &rtapi.Envelope_MatchmakerAdd{MatchmakerAdd: &rtapi.MatchmakerAdd{Query: "*", MinCount: 2, MaxCount: 2,
		StringProperties: map[string]string{"game": "tst_board", "room": "arena"}}}}
	if _, err := h.s.beforeMatchmakerAdd(ctx, nil, nil, nil, env); err == nil || !strings.Contains(err.Error(), codeRequirementsNotMet) || !strings.Contains(err.Error(), "buy_pass") {
		t.Fatalf("the matchmaker hook refuses with the products: %v", err)
	}
	if err := (gameProvider{s: h.s}).Check(h.ctx, gamecatalog.Entry{Game: "tst_board", Room: "bots", UserID: user, Path: gamecatalog.PathBot}); gamecatalog.Code(err) != codeRequirementsNotMet {
		t.Fatalf("check for a bot room %v", err)
	}

	w.giveItem(user, w.vipPass)
	w.tagUser(user, w.vip)
	w.setMeta(user, w.xp, 50)
	h.setLevel(user, 3)
	view = h.mustRPC(user, "economy_games", map[string]any{"lang": "en"})["items"].([]any)[0].(map[string]any)
	for _, k := range []string{"open", "arena", "vip", "pro"} {
		if r := roomByKey(view, k); r == nil || r["eligible"] != true || r["locked"] != false || len(r["products"].([]any)) != 0 {
			t.Fatalf("room %s after meeting everything: %v", k, r)
		}
	}
	h.mustRPC(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "arena"})

	w.tagUser(user, w.cheat)
	_, err = h.rpc(user, "economy_game_join", map[string]any{"key": "tst_board", "room": "open"})
	if app, ok := asApp(err); !ok || app.Code != codeRequirementsNotMet || app.Meta["missing"].([]map[string]any)[0]["type"] != "excludedTag" ||
		len(app.Meta["missing"].([]map[string]any)[0]["tags"].([]map[string]any)) != 0 {
		t.Fatalf("an excluded tag refuses without naming a private tag: %v", err)
	}
}

func TestGameRequirementTicketsConsumedOnceAndReturned(t *testing.T) {
	w := newReqWorld(t)
	h := w.h
	w.game(nil, []map[string]any{
		{"key": "pass", "requirements": map[string]any{"itemIds": []string{w.pass}, "consumeItems": true}},
		{"key": "tickets", "requirements": map[string]any{"metas": []map[string]any{{"metaId": w.ticket, "operator": "gte", "value": 1, "consume": 1}}}},
		{"key": "kept", "requirements": map[string]any{"metas": []map[string]any{{"metaId": w.ticket, "operator": "gte", "value": 2, "consume": 2}}, "refund": "keep"}},
	})
	p := gameProvider{s: h.s}
	a := h.user()
	w.setMeta(a, w.ticket, 3)
	match := "t-" + randomSuffix()
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	charged := make(chan gamecatalog.EntryResult, 10)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node := testService()
			node.db, node.nk = h.s.db, h.nk
			r, err := gameProvider{s: node}.Enter(h.ctx, gamecatalog.Entry{MatchID: match, Game: "tst_board", Room: "tickets", UserID: a})
			if err != nil {
				errs <- err
				return
			}
			charged <- r
		}()
	}
	wg.Wait()
	close(errs)
	close(charged)
	for err := range errs {
		t.Fatal(err)
	}
	took := 0
	for r := range charged {
		if r.Tickets != nil {
			took++
		}
	}
	if took != 1 || w.metaOf(a, w.ticket) != 2 {
		t.Fatalf("one ticket per entry under parallel joins: took=%d left=%v", took, w.metaOf(a, w.ticket))
	}
	if _, err := p.Leave(h.ctx, match, a); err != nil {
		t.Fatal(err)
	}
	if w.metaOf(a, w.ticket) != 3 {
		t.Fatalf("a leave before the start returns the ticket: %v", w.metaOf(a, w.ticket))
	}
	if _, err := p.Leave(h.ctx, match, a); err != nil || w.metaOf(a, w.ticket) != 3 {
		t.Fatalf("a second leave returns nothing more: %v %v", err, w.metaOf(a, w.ticket))
	}

	b := h.user()
	pm := "p-" + randomSuffix()
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: pm, Game: "tst_board", Room: "pass", UserID: b}); gamecatalog.Code(err) != codeRequirementsNotMet {
		t.Fatalf("no pass %v", err)
	} else if ge, ok := err.(*gamecatalog.Error); !ok || len(ge.Meta["products"].([]map[string]any)) != 1 {
		t.Fatalf("enter refusal carries the products: %v", err)
	}
	w.giveItem(b, w.pass)
	r, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: pm, Game: "tst_board", Room: "pass", UserID: b, Path: gamecatalog.PathFriend})
	if err != nil || r.Tickets["items"].([]string)[0] != "arena_pass" || w.owns(b, w.pass) {
		t.Fatalf("the pass is used up on entry: %+v %v", r, err)
	}
	if r2, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: pm, Game: "tst_board", Room: "pass", UserID: b}); err != nil || !r2.Rejoined || r2.Tickets != nil {
		t.Fatalf("a rejoin needs no new pass: %+v %v", r2, err)
	}
	if h.count(`SELECT count(*) FROM economy_game_entry WHERE match_id=$1 AND path='friend' AND tickets IS NOT NULL`, pm) != 1 {
		t.Fatal("entry keeps the path and the tickets")
	}
	if n, err := p.Cancel(h.ctx, pm); err != nil || n != 0 || !w.owns(b, w.pass) {
		t.Fatalf("a cancelled match gives the pass back: %d %v owns=%v", n, err, w.owns(b, w.pass))
	}
	if h.count(`SELECT count(*) FROM economy_user_item WHERE user_id=$1 AND item_id=$2 AND source='buy'`, b, w.pass) != 1 {
		t.Fatal("the returned pass keeps its source")
	}
	if h.count(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND action IN ('game_ticket','game_ticket_returned')`, b) != 2 {
		t.Fatal("activity rows for the ticket and its return")
	}

	c := h.user()
	w.setMeta(c, w.ticket, 2)
	km := "k-" + randomSuffix()
	if _, err := p.Enter(h.ctx, gamecatalog.Entry{MatchID: km, Game: "tst_board", Room: "kept", UserID: c}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Leave(h.ctx, km, c); err != nil || w.metaOf(c, w.ticket) != 0 {
		t.Fatalf("refund keep does not give tickets back: %v %v", err, w.metaOf(c, w.ticket))
	}
	view := h.mustRPC(c, "economy_games", map[string]any{"lang": "en"})["items"].([]any)[0].(map[string]any)
	kept := roomByKey(view, "kept")
	if kept["eligible"] != false || kept["consumes"] != true || kept["products"].([]any)[0].(map[string]any)["key"] != "buy_tickets" {
		t.Fatalf("a player without tickets sees the ticket pack: %v", kept)
	}
	detail := h.mustAPI(200, http.MethodGet, "/game-matches/"+km, nil)
	e := detail["entries"].([]any)[0].(map[string]any)
	if e["path"] != "matchmaking" || e["tickets"].(map[string]any)["refund"] != "keep" {
		t.Fatalf("match detail shows the tickets %v", e)
	}
}
