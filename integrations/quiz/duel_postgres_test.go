package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type noteNK struct {
	*plugintest.Wallet
	mu    sync.Mutex
	notes []map[string]any
}

func (n *noteNK) NotificationSend(_ context.Context, userID, subject string, content map[string]interface{}, code int, sender string, persistent bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.notes = append(n.notes, map[string]any{"to": userID, "subject": subject, "event": content["event"], "code": code, "persistent": persistent})
	return nil
}

func (n *noteNK) count(to, event string) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	c := 0
	for _, x := range n.notes {
		if x["to"] == to && x["event"] == event {
			c++
		}
	}
	return c
}

type duelFix struct {
	*fixture
	nk   *noteNK
	cats []string
}

func newDuelFix(t *testing.T, change func(*settings.Duel)) *duelFix {
	f := newFixture(t)
	nk := &noteNK{Wallet: f.wallet}
	f.s.nk = nk
	df := &duelFix{fixture: f, nk: nk}
	t.Cleanup(func() {
		p := f.prefix + "%"
		_, _ = f.db.Exec(`DELETE FROM quiz_duel d WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id=d.player_a)`)
		_, _ = f.db.Exec(`DELETE FROM quiz_duel_answer a WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id=a.user_id)`)
		_, _ = f.db.Exec(`DELETE FROM quiz_category WHERE key LIKE $1`, p)
	})
	f.settings(func(c *settings.All) {
		c.Duel.WinPrize, c.Duel.LosePrize, c.Duel.DrawPrize = "win_prize", "lose_prize", "draw_prize"
		c.Duel.PerCorrect = settings.Price{CurrencyKey: f.coin, Amount: 2}
		if change != nil {
			change(&c.Duel)
		}
	})
	for i := 0; i < 4; i++ {
		key := f.prefix + "_cat" + string(rune('a'+i))
		plugintest.Exec(t, f.db, `INSERT INTO quiz_category (id, key, name) VALUES ($1,$2,$3)`, uuid.Must(uuid.NewV4()).String(), key, `{"fa":"دسته","en":"Category"}`)
		for j := 0; j < 6; j++ {
			plugintest.Exec(t, f.db, `INSERT INTO quiz_question (id, type, question, options, category) VALUES ($1,'options',$2,$3,$4)`, uuid.Must(uuid.NewV4()).String(),
				`{"fa":"سوال","en":"Q"}`, `[{"id":0,"text":{"fa":"الف"}},{"id":1,"text":{"fa":"ب"}},{"id":2,"text":{"fa":"ج"},"correct":true}]`, key)
		}
		df.cats = append(df.cats, key)
	}
	return df
}

func (f *duelFix) view(user, id string) map[string]any {
	return f.must(f.s.rpcDuelGet, user, map[string]any{"duelId": id})
}

func stepOf(v map[string]any) string { return v["step"].(map[string]any)["kind"].(string) }

func (f *duelFix) playTurn(user, id string, right bool) int {
	f.t.Helper()
	answered := 0
	for i := 0; i < 50; i++ {
		v := f.view(user, id)
		switch stepOf(v) {
		case "pick":
			offers := v["step"].(map[string]any)["offers"].([]any)
			if len(offers) == 0 {
				f.t.Fatalf("no offers: %v", v)
			}
			f.must(f.s.rpcDuelPick, user, map[string]any{"duelId": id, "categoryKey": offers[0].(map[string]any)["key"]})
		case "answer":
			q := f.must(f.s.rpcDuelQuestion, user, map[string]any{"duelId": id})["question"].(map[string]any)
			opt := 2
			if !right {
				opt = 0
			}
			f.must(f.s.rpcDuelAnswer, user, map[string]any{"duelId": id, "questionId": q["questionId"], "optionId": opt})
			answered++
		default:
			return answered
		}
	}
	f.t.Fatal("turn did not end")
	return answered
}

func TestDuelFullRandomGame(t *testing.T) {
	f := newDuelFix(t, nil)
	a, b := f.user(), f.user()
	start := f.must(f.s.rpcDuelStart, a, map[string]any{"mode": "random"})
	id := start["duelId"].(string)
	if start["joined"] != false || stepOf(start) != "pick" {
		t.Fatalf("start: %v", start)
	}
	offers := start["step"].(map[string]any)["offers"].([]any)
	if len(offers) != 3 {
		t.Fatalf("offers: %v", offers)
	}
	again := f.view(a, id)["step"].(map[string]any)["offers"].([]any)
	if again[0].(map[string]any)["key"] != offers[0].(map[string]any)["key"] {
		t.Fatal("offers changed on re-read")
	}
	if n := f.playTurn(a, id, true); n != 3 {
		t.Fatalf("turn 1 answered %d", n)
	}
	if v := f.view(a, id); v["status"] != "waiting" || v["turn"] != "none" {
		t.Fatalf("after turn 1: %v %v", v["status"], v["turn"])
	}
	joined := f.must(f.s.rpcDuelStart, b, map[string]any{"mode": "random"})
	if joined["duelId"] != id || joined["joined"] != true || stepOf(joined) != "answer" {
		t.Fatalf("join: %v", joined)
	}
	if rl := joined["roundList"].([]any)[0].(map[string]any); rl["opponent"] != nil {
		t.Fatalf("opponent answers shown before playing the round: %v", rl)
	}
	if _, err := f.rpc(f.s.rpcDuelQuestion, a, map[string]any{"duelId": id}); code(err) != "QUIZ_DUEL_NOT_YOUR_TURN" {
		t.Fatalf("a out of turn: %v", err)
	}
	turns := 1
	for cur, right := b, false; ; {
		v := f.view(cur, id)
		if v["status"] == "finished" {
			break
		}
		n := f.playTurn(cur, id, right)
		turns++
		if n == 0 || turns > 10 {
			t.Fatalf("turn %d by %s answered %d", turns, cur, n)
		}
		if cur == a {
			cur, right = b, false
		} else {
			cur, right = a, true
		}
	}
	if turns != 7 {
		t.Fatalf("turns %d", turns)
	}
	va := f.view(a, id)
	if va["result"] != "win" || va["reason"] != "score" || va["you"].(map[string]any)["score"].(float64) != 18 || va["opponent"].(map[string]any)["score"].(float64) != 0 {
		t.Fatalf("result: %v", va)
	}
	if vb := f.view(b, id); vb["result"] != "lose" {
		t.Fatalf("b result %v", vb["result"])
	}
	f.s.payDuel(context.Background(), id)
	f.s.payDuel(context.Background(), id)
	if len(f.prizes) != 2 || f.prizes[0].PrizeKey == f.prizes[1].PrizeKey {
		t.Fatalf("prizes: %+v", f.prizes)
	}
	if w := plugintest.Balance(t, f.db, a); w[f.coin] != 36 {
		t.Fatalf("a wallet %v", w)
	}
	if w := plugintest.Balance(t, f.db, b); w[f.coin] != 0 {
		t.Fatalf("b wallet %v", w)
	}
	if f.nk.count(a, "finished") != 1 || f.nk.count(b, "finished") != 1 || f.nk.count(a, "turn") == 0 {
		t.Fatalf("notes: %v", f.nk.notes)
	}
	st := f.must(f.s.rpcDuelStats, a, map[string]any{})
	if st["played"].(float64) != 1 || st["wins"].(float64) != 1 || st["correct"].(float64) != 18 {
		t.Fatalf("stats %v", st)
	}
}

func TestDuelAnswerReplayLateAndTimeout(t *testing.T) {
	f := newDuelFix(t, nil)
	a, b := f.user(), f.user()
	id := f.must(f.s.rpcDuelStart, a, map[string]any{"mode": "random"})["duelId"].(string)
	offers := f.view(a, id)["step"].(map[string]any)["offers"].([]any)
	f.must(f.s.rpcDuelPick, a, map[string]any{"duelId": id, "categoryKey": offers[1].(map[string]any)["key"]})
	if _, err := f.rpc(f.s.rpcDuelPick, a, map[string]any{"duelId": id, "categoryKey": offers[0].(map[string]any)["key"]}); code(err) != "QUIZ_DUEL_NOT_PICKING" {
		t.Fatalf("second pick: %v", err)
	}
	q1 := f.must(f.s.rpcDuelQuestion, a, map[string]any{"duelId": id})["question"].(map[string]any)
	if _, ok := q1["options"].([]any)[0].(map[string]any)["correct"]; ok {
		t.Fatal("correct flag sent")
	}
	again := f.must(f.s.rpcDuelQuestion, a, map[string]any{"duelId": id})["question"].(map[string]any)
	if again["questionId"] != q1["questionId"] || again["servedAt"] != q1["servedAt"] {
		t.Fatal("re-serving restarted the clock")
	}
	r1 := f.must(f.s.rpcDuelAnswer, a, map[string]any{"duelId": id, "questionId": q1["questionId"], "optionId": 2})
	r2 := f.must(f.s.rpcDuelAnswer, a, map[string]any{"duelId": id, "questionId": q1["questionId"], "optionId": 0})
	if r1["outcome"] != "correct" || r1["replayed"] != false || r2["outcome"] != "correct" || r2["replayed"] != true {
		t.Fatalf("replay: %v / %v", r1, r2)
	}
	q2 := f.must(f.s.rpcDuelQuestion, a, map[string]any{"duelId": id})["question"].(map[string]any)
	f.advance(23 * time.Second)
	if r := f.must(f.s.rpcDuelAnswer, a, map[string]any{"duelId": id, "questionId": q2["questionId"], "optionId": 2}); r["outcome"] != "late" {
		t.Fatalf("late: %v", r)
	}
	q3 := f.must(f.s.rpcDuelQuestion, a, map[string]any{"duelId": id})["question"].(map[string]any)
	f.advance(30 * time.Second)
	out := f.must(f.s.rpcDuelQuestion, a, map[string]any{"duelId": id})
	if out["question"] != nil || out["duel"].(map[string]any)["status"] != "waiting" {
		t.Fatalf("expired question not closed: %v", out)
	}
	var outcome string
	_ = f.db.QueryRow(`SELECT outcome FROM quiz_duel_answer WHERE duel_id=$1 AND question_id=$2`, id, q3["questionId"]).Scan(&outcome)
	if outcome != "late" {
		t.Fatalf("q3 %s", outcome)
	}
	f.must(f.s.rpcDuelStart, b, map[string]any{"mode": "random"})
	f.advance(25 * time.Hour)
	f.s.duelWorker(context.Background(), 50)
	v := f.view(a, id)
	if v["status"] != "finished" || v["reason"] != "timeout" || v["result"] != "win" {
		t.Fatalf("timeout: %v %v %v", v["status"], v["reason"], v["result"])
	}
	f.s.duelWorker(context.Background(), 50)
	if len(f.prizes) != 2 {
		t.Fatalf("prizes %d", len(f.prizes))
	}
}

func TestDuelTimeoutByScore(t *testing.T) {
	f := newDuelFix(t, func(d *settings.Duel) { d.TimeoutResult = settings.TimeoutScore })
	a, b := f.user(), f.user()
	id := f.must(f.s.rpcDuelStart, a, map[string]any{"mode": "random"})["duelId"].(string)
	f.playTurn(a, id, false)
	f.must(f.s.rpcDuelStart, b, map[string]any{"mode": "random"})
	f.advance(25 * time.Hour)
	f.s.duelWorker(context.Background(), 50)
	if v := f.view(b, id); v["result"] != "draw" || v["reason"] != "timeout" {
		t.Fatalf("score timeout: %v %v", v["result"], v["reason"])
	}
}

func TestDuelEntryFeeRefundsAndFriends(t *testing.T) {
	f := newDuelFix(t, func(d *settings.Duel) { d.EntryFee = settings.Price{CurrencyKey: "gem_q", Amount: 10} })
	fee := "gem_q"
	a, b, c := f.user(), f.user(), f.user()
	if _, err := f.rpc(f.s.rpcDuelStart, a, map[string]any{"mode": "random"}); code(err) != "QUIZ_DUEL_NOT_ENOUGH_CURRENCY" {
		t.Fatalf("no money: %v", err)
	}
	var n int
	_ = f.db.QueryRow(`SELECT count(*) FROM quiz_duel WHERE player_a=$1`, a).Scan(&n)
	if n != 0 {
		t.Fatal("duel kept after a failed fee")
	}
	f.wallet.Give(t, a, map[string]int64{fee: 25}, map[string]any{})
	id := f.must(f.s.rpcDuelStart, a, map[string]any{"mode": "random"})["duelId"].(string)
	if w := plugintest.Balance(t, f.db, a); w[fee] != 15 {
		t.Fatalf("fee not taken: %v", w)
	}
	f.playTurn(a, id, true)
	f.advance(49 * time.Hour)
	f.s.duelWorker(context.Background(), 50)
	f.s.duelWorker(context.Background(), 50)
	if v := f.view(a, id); v["status"] != "expired" {
		t.Fatalf("not expired: %v", v["status"])
	}
	if w := plugintest.Balance(t, f.db, a); w[fee] != 25 {
		t.Fatalf("refund: %v", w)
	}
	if _, err := f.rpc(f.s.rpcDuelStart, a, map[string]any{"mode": "friend", "userId": b}); code(err) != "QUIZ_DUEL_NOT_FRIENDS" {
		t.Fatalf("friends only: %v", err)
	}
	plugintest.Exec(t, f.db, `INSERT INTO user_edge (source_id, position, destination_id, state) VALUES ($1, 1, $2, 0), ($2, 1, $1, 0)`, a, b)
	inv := f.must(f.s.rpcDuelStart, a, map[string]any{"mode": "friend", "userId": b})
	fid := inv["duelId"].(string)
	if inv["status"] != "invited" || f.nk.count(b, "invite") != 1 {
		t.Fatalf("invite: %v", inv["status"])
	}
	if v := f.view(b, fid); stepOf(v) != "accept" {
		t.Fatalf("b step %v", stepOf(v))
	}
	if _, err := f.rpc(f.s.rpcDuelAccept, c, map[string]any{"duelId": fid}); code(err) != "QUIZ_DUEL_NOT_FOUND" {
		t.Fatalf("stranger accept: %v", err)
	}
	f.must(f.s.rpcDuelDecline, b, map[string]any{"duelId": fid})
	f.s.duelWorker(context.Background(), 50)
	if w := plugintest.Balance(t, f.db, a); w[fee] != 25 {
		t.Fatalf("decline refund: %v", w)
	}
	inv2 := f.must(f.s.rpcDuelStart, a, map[string]any{"mode": "friend", "userId": b})["duelId"].(string)
	f.playTurn(a, inv2, true)
	if _, err := f.rpc(f.s.rpcDuelAccept, b, map[string]any{"duelId": inv2}); code(err) != "QUIZ_DUEL_NOT_ENOUGH_CURRENCY" {
		t.Fatalf("b without money: %v", err)
	}
	f.wallet.Give(t, b, map[string]int64{fee: 10}, map[string]any{})
	acc := f.must(f.s.rpcDuelAccept, b, map[string]any{"duelId": inv2})
	if acc["status"] != "active" || stepOf(acc) != "answer" || acc["turnDeadline"] == nil {
		t.Fatalf("accept: %v %v", acc["status"], stepOf(acc))
	}
	f.must(f.s.rpcDuelForfeit, b, map[string]any{"duelId": inv2})
	if v := f.view(a, inv2); v["result"] != "win" || v["reason"] != "forfeit" {
		t.Fatalf("forfeit %v %v", v["result"], v["reason"])
	}
	code2, _ := f.console(f.s.duelCancel, http.MethodPost, "/", nil, map[string]string{"id": inv2})
	if code2 != http.StatusConflict {
		t.Fatalf("cancel finished duel: %d", code2)
	}
}

func TestDuelLimitsAndConsole(t *testing.T) {
	f := newDuelFix(t, func(d *settings.Duel) { d.MaxActive = 1 })
	a := f.user()
	id := f.must(f.s.rpcDuelStart, a, map[string]any{"mode": "random"})["duelId"].(string)
	if _, err := f.rpc(f.s.rpcDuelStart, a, map[string]any{"mode": "random"}); code(err) != "QUIZ_DUEL_ACTIVE_LIMIT" {
		t.Fatalf("limit: %v", err)
	}
	c, out := f.console(f.s.duelList, http.MethodGet, "/?status=open&player="+a, nil, nil)
	if c != 200 || out["total"].(float64) != 1 {
		t.Fatalf("list: %d %v", c, out)
	}
	c, out = f.console(f.s.duelGet, http.MethodGet, "/", nil, map[string]string{"id": id})
	if c != 200 || len(out["rounds"].([]any)) != 6 {
		t.Fatalf("detail: %d %v", c, out)
	}
	c, _ = f.console(f.s.duelCancel, http.MethodPost, "/", nil, map[string]string{"id": id})
	if c != 200 {
		t.Fatalf("cancel %d", c)
	}
	if v := f.view(a, id); v["status"] != "cancelled" {
		t.Fatalf("cancelled: %v", v["status"])
	}
	c, out = f.console(f.s.duelCategorySave, http.MethodPost, "/", map[string]any{"key": f.prefix + "_new", "name": map[string]string{"fa": "تاریخ"}, "color": "#12AB34"}, nil)
	if c != 201 {
		t.Fatalf("category: %d %v", c, out)
	}
	c, _ = f.console(f.s.duelCategorySave, http.MethodPost, "/", map[string]any{"key": f.prefix + "_bad", "name": map[string]string{"fa": "x"}, "color": "red"}, nil)
	if c != 400 {
		t.Fatalf("bad color: %d", c)
	}
	c, _ = f.console(f.s.duelCategoryDelete, http.MethodDelete, "/", nil, map[string]string{"id": f.idOfCategory(f.cats[0])})
	if c != 409 {
		t.Fatalf("in use: %d", c)
	}
	c, out = f.console(f.s.duelStats, http.MethodGet, "/?days=7", nil, nil)
	if c != 200 || out["started"].(float64) < 1 {
		t.Fatalf("stats %d %v", c, out)
	}
	cfg := settings.Defaults()
	cfg.Duel.Rounds = 0
	c, out = f.console(f.s.settingsPut, http.MethodPut, "/", cfg, nil)
	if c != 400 || out["field"] != "duel.rounds" {
		t.Fatalf("settings: %d %v", c, out)
	}
}

func (f *duelFix) idOfCategory(key string) string {
	var id string
	_ = f.db.QueryRow(`SELECT id::text FROM quiz_category WHERE key=$1`, key).Scan(&id)
	return id
}
