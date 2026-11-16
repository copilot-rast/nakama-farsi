package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type fixture struct {
	t         *testing.T
	db        *sql.DB
	s         *service
	wallet    *plugintest.Wallet
	mu        sync.Mutex
	clock     time.Time
	prizes    []economyhooks.PrizeGrant
	completes []campaignhooks.Completion
	coin      string
	prefix    string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, db: db, wallet: &plugintest.Wallet{DB: db}, clock: time.Now().UTC().Truncate(time.Second), prefix: plugintest.Key("qt")}
	f.s = &service{db: db, nk: f.wallet, logger: plugintest.Logger{}}
	f.s.now = func() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.clock }
	f.s.prize = func(_ context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, p := range f.prizes {
			if p.SourceID == g.SourceID {
				return economyhooks.PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey}}}, nil
			}
		}
		f.prizes = append(f.prizes, g)
		return economyhooks.PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey}}}, nil
	}
	f.s.campaignOn = func() bool { return true }
	f.s.campaignStart = func(_ context.Context, userID, stageKey string) (string, error) { return "attempt-" + stageKey, nil }
	f.s.campaignComplete = func(_ context.Context, userID string, c campaignhooks.Completion) (map[string]any, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.completes = append(f.completes, c)
		return map[string]any{"attemptId": c.AttemptID, "rewards": []any{}}, nil
	}
	f.s.send = func(context.Context, string, []byte, map[string]string) (int, error) { return 200, nil }
	f.coin = plugintest.Key("coin")
	t.Cleanup(func() {
		p := f.prefix + "%"
		_, _ = db.Exec(`DELETE FROM quiz_result_template WHERE key LIKE $1`, p)
		_, _ = db.Exec(`DELETE FROM quiz_question WHERE key LIKE $1 OR pack_id IN (SELECT id FROM quiz_pack WHERE key LIKE $1) OR category LIKE $1`, p)
		_, _ = db.Exec(`DELETE FROM quiz_pack WHERE key LIKE $1`, p)
		_, _ = db.Exec(`DELETE FROM quiz_mission WHERE key LIKE $1`, p)
		_, _ = db.Exec(`UPDATE quiz_settings SET value='{}' WHERE id=1`)
		settings.Invalidate()
	})
	f.settings(func(*settings.All) {})
	return f
}

func (f *fixture) settings(change func(*settings.All)) {
	f.t.Helper()
	cfg := settings.Defaults()
	cfg.Play.ShuffleQuestions, cfg.Play.ShuffleOptions = false, false
	change(&cfg)
	if err := settings.Save(context.Background(), f.db, cfg, "test"); err != nil {
		f.t.Fatal(err)
	}
	settings.Invalidate()
}

func (f *fixture) advance(d time.Duration) {
	f.mu.Lock()
	f.clock = f.clock.Add(d)
	f.mu.Unlock()
}

func (f *fixture) user() string {
	u := plugintest.NewUser(f.t, f.db)
	f.t.Cleanup(func() {
		for _, q := range []string{`DELETE FROM quiz_session WHERE user_id=$1`, `DELETE FROM quiz_user_pack WHERE user_id=$1`, `DELETE FROM quiz_user_question WHERE user_id=$1`,
			`DELETE FROM quiz_user_mission WHERE user_id=$1`, `DELETE FROM quiz_user_hint WHERE user_id=$1`, `DELETE FROM quiz_user_day WHERE user_id=$1`,
			`DELETE FROM quiz_question_report WHERE user_id=$1`, `DELETE FROM quiz_result_delivery WHERE user_id=$1`, `DELETE FROM wallet_ledger WHERE user_id=$1`} {
			_, _ = f.db.Exec(q, u)
		}
	})
	return u
}

func (f *fixture) console(h http.HandlerFunc, method, target string, body any, vars map[string]string) (int, map[string]any) {
	f.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, target, &buf)
	for k, v := range vars {
		r.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	h(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func (f *fixture) mission(key string, body map[string]any) string {
	f.t.Helper()
	body["key"] = f.prefix + key
	if body["name"] == nil {
		body["name"] = map[string]string{"fa": key}
	}
	code, out := f.console(f.s.missionSave, http.MethodPost, "/", body, nil)
	if code != 200 {
		f.t.Fatalf("mission: %d %v", code, out)
	}
	return out["id"].(string)
}

func (f *fixture) pack(key string, body map[string]any) string {
	f.t.Helper()
	body["key"] = f.prefix + key
	if body["name"] == nil {
		body["name"] = map[string]string{"fa": key, "en": key}
	}
	code, out := f.console(f.s.packSave, http.MethodPost, "/", body, nil)
	if code != 200 {
		f.t.Fatalf("pack: %d %v", code, out)
	}
	return out["pack"].(map[string]any)["id"].(string)
}

func (f *fixture) question(packID string, extra map[string]any) string {
	f.t.Helper()
	body := map[string]any{"packId": packID, "type": "options", "question": map[string]string{"fa": "سوال", "en": "Question"},
		"options":     []map[string]any{{"id": 0, "text": map[string]string{"fa": "الف"}}, {"id": 1, "text": map[string]string{"fa": "ب"}}, {"id": 2, "text": map[string]string{"fa": "ج"}, "correct": true}},
		"explanation": map[string]string{"fa": "چون ج"}}
	for k, v := range extra {
		body[k] = v
	}
	code, out := f.console(f.s.questionSave, http.MethodPost, "/", body, nil)
	if code != 200 {
		f.t.Fatalf("question: %d %v", code, out)
	}
	return out["question"].(map[string]any)["id"].(string)
}

func (f *fixture) rpc(fn clientRPC, userID string, in any) (map[string]any, error) {
	f.t.Helper()
	raw, _ := json.Marshal(in)
	out, err := fn(context.Background(), userID, string(raw))
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(out)
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m, nil
}

func (f *fixture) must(fn clientRPC, userID string, in any) map[string]any {
	f.t.Helper()
	out, err := f.rpc(fn, userID, in)
	if err != nil {
		f.t.Fatalf("rpc: %v", err)
	}
	return out
}

func code(err error) string {
	if e, ok := apierr.As(err); ok {
		return e.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func questionsOf(view map[string]any) []map[string]any {
	out := []map[string]any{}
	list, _ := view["questions"].([]any)
	for _, q := range list {
		out = append(out, q.(map[string]any))
	}
	return out
}

func TestPostgresPlayAnswerFinishPaysOnce(t *testing.T) {
	f := newFixture(t)
	k := f.pack("p1", map[string]any{"passPercent": 50, "reward": map[string]any{"currencies": map[string]int64{f.coin: 100}, "prizeKey": "box"}})
	q1 := f.question(k, map[string]any{"reward": map[string]any{"currencies": map[string]int64{f.coin: 7}}})
	q2 := f.question(k, nil)
	u := f.user()
	view := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "p1"})
	qs := questionsOf(view)
	if len(qs) != 2 || view["resumed"] != false {
		t.Fatalf("view: %v", view)
	}
	for _, q := range qs {
		for _, o := range q["options"].([]any) {
			if _, leaked := o.(map[string]any)["correct"]; leaked {
				t.Fatal("the right answer was sent before answering")
			}
		}
	}
	sid := view["sessionId"].(string)
	again := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "p1"})
	if again["sessionId"] != sid || again["resumed"] != true {
		t.Fatalf("a second play resumes the open session: %v", again)
	}
	a1 := f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q1, "optionId": 2})
	if a1["outcome"] != "correct" || a1["replayed"] != false || a1["explanation"] != "چون ج" {
		t.Fatalf("answer: %v", a1)
	}
	resend := f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q1, "optionId": 0})
	if resend["outcome"] != "correct" || resend["replayed"] != true || resend["correctCount"].(float64) != 1 {
		t.Fatalf("a resend answers the stored outcome: %v", resend)
	}
	if _, err := f.rpc(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q2, "optionId": 9}); code(err) != "QUIZ_OPTION_NOT_FOUND" {
		t.Fatalf("unknown option: %v", err)
	}
	a2 := f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q2, "optionId": 1})
	res, _ := a2["result"].(map[string]any)
	if a2["outcome"] != "wrong" || res == nil || res["passed"] != true || res["score"].(float64) != 50 {
		t.Fatalf("auto finish: %v", a2)
	}
	if bal := plugintest.Balance(t, f.db, u)[f.coin]; bal != 107 {
		t.Fatalf("pack reward + question reward = 107, got %d", bal)
	}
	fin := f.must(f.s.rpcFinish, u, map[string]any{"sessionId": sid})
	if fin["replayed"] != true {
		t.Fatalf("finish again replays: %v", fin)
	}
	if bal := plugintest.Balance(t, f.db, u)[f.coin]; bal != 107 || len(f.prizes) != 1 || f.prizes[0].SourceType != "quiz" {
		t.Fatalf("paid once: %d %v", bal, f.prizes)
	}

	view = f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "p1"})
	sid2 := view["sessionId"].(string)
	all := f.must(f.s.rpcAnswerAll, u, map[string]any{"sessionId": sid2, "answers": []map[string]any{{"questionId": q1, "optionId": 2}, {"questionId": q2, "optionId": 2}}})
	res = all["result"].(map[string]any)
	if res["passed"] != true || len(res["rewards"].([]any)) != 0 || len(res["questionRewards"].([]any)) != 0 {
		t.Fatalf("second pass pays nothing: %v", res)
	}
	if bal := plugintest.Balance(t, f.db, u)[f.coin]; bal != 107 {
		t.Fatalf("balance after second pass %d", bal)
	}

	f.settings(func(a *settings.All) { a.Rewards.PackMode = settings.RewardEvery })
	view = f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "p1"})
	f.must(f.s.rpcAnswerAll, u, map[string]any{"sessionId": view["sessionId"], "answers": []map[string]any{{"questionId": q1, "optionId": 2}, {"questionId": q2, "optionId": 2}}})
	if bal := plugintest.Balance(t, f.db, u)[f.coin]; bal != 207 {
		t.Fatalf("every pass: %d", bal)
	}
	var plays, wins int
	_ = f.db.QueryRow(`SELECT plays, wins FROM quiz_user_pack WHERE user_id=$1 AND pack_id=$2`, u, k).Scan(&plays, &wins)
	if plays != 3 || wins != 3 {
		t.Fatalf("stats %d %d", plays, wins)
	}
}

func TestPostgresServerClock(t *testing.T) {
	f := newFixture(t)
	k := f.pack("clock", map[string]any{"questionSeconds": 10, "timeSeconds": 60, "passPercent": 0})
	q1 := f.question(k, nil)
	q2 := f.question(k, nil)
	f.question(k, nil)
	u := f.user()
	view := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "clock"})
	sid := view["sessionId"].(string)
	if questionsOf(view)[0]["deadlineAt"] == nil {
		t.Fatal("the current question has a deadline")
	}
	f.advance(15 * time.Second)
	late := f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q1, "optionId": 2})
	if late["outcome"] != "late" || late["correct"] != false {
		t.Fatalf("late answer: %v", late)
	}
	f.advance(5 * time.Second)
	ok := f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q2, "optionId": 2})
	if ok["outcome"] != "correct" {
		t.Fatalf("the clock restarts at the previous answer: %v", ok)
	}
	f.advance(2 * time.Minute)
	if _, err := f.rpc(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q2, "optionId": 2}); err != nil {
		t.Fatalf("a resend after the time still answers the stored outcome: %v", err)
	}
	_, err := f.rpc(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": questionsOf(view)[2]["id"], "optionId": 2})
	if code(err) != "QUIZ_TIME_OVER" {
		t.Fatalf("pack time over: %v", err)
	}
	fin := f.must(f.s.rpcFinish, u, map[string]any{"sessionId": sid})
	if fin["expired"] != true || fin["correct"].(float64) != 1 || fin["total"].(float64) != 3 {
		t.Fatalf("scored on time over: %v", fin)
	}

	f.settings(func(a *settings.All) { a.Timing.LateAnswer = settings.LateReject })
	view = f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "clock"})
	f.advance(20 * time.Second)
	if _, err := f.rpc(f.s.rpcAnswer, u, map[string]any{"sessionId": view["sessionId"], "questionId": q1, "optionId": 2}); code(err) != "QUIZ_QUESTION_TIME_OVER" {
		t.Fatalf("reject: %v", err)
	}
}

func TestPostgresOneByOneAndReveal(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) {
		a.Play.Delivery = settings.DeliveryOneByOne
		a.Reveal.Mode = settings.RevealFinish
	})
	k := f.pack("obo", map[string]any{})
	q1 := f.question(k, nil)
	q2 := f.question(k, nil)
	u := f.user()
	view := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "obo"})
	if qs := questionsOf(view); len(qs) != 1 || qs[0]["id"] != q1 {
		t.Fatalf("only the current question is sent: %v", qs)
	}
	sid := view["sessionId"]
	if _, err := f.rpc(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q2, "optionId": 2}); code(err) != "QUIZ_QUESTION_NOT_CURRENT" {
		t.Fatalf("not current: %v", err)
	}
	a := f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q1, "optionId": 2})
	if _, revealed := a["correctOptionIds"]; revealed {
		t.Fatal("after_finish: no reveal on answer")
	}
	if next, _ := a["next"].(map[string]any); next == nil || next["id"] != q2 {
		t.Fatalf("next question: %v", a)
	}
	a = f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": sid, "questionId": q2, "optionId": 0})
	res := a["result"].(map[string]any)
	review := res["review"].([]any)
	if res["passed"] != false || len(review) != 2 || review[1].(map[string]any)["outcome"] != "wrong" {
		t.Fatalf("review on finish: %v", res)
	}

	f.settings(func(a *settings.All) { a.Reveal.AllowBeforeAnswer = true })
	f.console(f.s.packSave, http.MethodPut, "/", map[string]any{"name": map[string]string{"fa": "x"}, "sendAnswers": true}, map[string]string{"id": k})
	view = f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "obo"})
	opts := questionsOf(view)[0]["options"].([]any)
	if opts[2].(map[string]any)["correct"] != true {
		t.Fatalf("send answers: %v", opts)
	}
}

func TestPostgresHintsPaidOnce(t *testing.T) {
	f := newFixture(t)
	k := f.pack("hint", map[string]any{})
	q := f.question(k, nil)
	code0, h := f.console(f.s.hintSave, http.MethodPost, "/", map[string]any{"name": map[string]string{"fa": "راهنما"}, "text": map[string]string{"fa": "گزینه سوم"},
		"price": map[string]any{"currencyKey": f.coin, "amount": 30}}, map[string]string{"id": q})
	if code0 != 200 {
		t.Fatalf("hint: %v", h)
	}
	hid := h["id"].(string)
	u := f.user()
	if _, err := f.rpc(f.s.rpcHintBuy, u, map[string]any{"hintId": hid}); code(err) != "QUIZ_INSUFFICIENT_CURRENCY_FOR_HINT" {
		t.Fatalf("no money: %v", err)
	}
	f.wallet.Give(t, u, map[string]int64{f.coin: 100}, map[string]any{"reason": "test"})
	var wg sync.WaitGroup
	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.rpc(f.s.rpcHintBuy, u, map[string]any{"hintId": hid})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("parallel buy: %v", err)
		}
	}
	if bal := plugintest.Balance(t, f.db, u)[f.coin]; bal != 70 {
		t.Fatalf("paid once: %d", bal)
	}
	out := f.must(f.s.rpcHintBuy, u, map[string]any{"hintId": hid})
	if out["replayed"] != true || out["text"] != "گزینه سوم" {
		t.Fatalf("owned hint: %v", out)
	}
	view := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "hint"})
	hints := questionsOf(view)[0]["hints"].([]any)
	if len(hints) != 1 || hints[0].(map[string]any)["owned"] != true || hints[0].(map[string]any)["text"] != "گزینه سوم" {
		t.Fatalf("hint in session: %v", hints)
	}
	var ledger int
	_ = f.db.QueryRow(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'sourceType'='quiz'`, u).Scan(&ledger)
	if ledger != 1 {
		t.Fatalf("one ledger row with sourceType quiz, got %d", ledger)
	}
}

func TestPostgresMissionsLimitsAndCampaign(t *testing.T) {
	f := newFixture(t)
	m := f.mission("m1", map[string]any{"orderMode": "sequential", "reward": map[string]any{"currencies": map[string]int64{f.coin: 500}}})
	a := f.pack("ma", map[string]any{"missionId": m, "sort": 1})
	b := f.pack("mb", map[string]any{"missionId": m, "sort": 2})
	qa := f.question(a, nil)
	qb := f.question(b, nil)
	u := f.user()
	if _, err := f.rpc(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "mb"}); code(err) != "QUIZ_MISSION_PACK_LOCKED" {
		t.Fatalf("sequential lock: %v", err)
	}
	v := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "ma"})
	f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": v["sessionId"], "questionId": qa, "optionId": 2})
	v = f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "mb"})
	out := f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": v["sessionId"], "questionId": qb, "optionId": 2})
	res := out["result"].(map[string]any)
	if done := res["missionsCompleted"].([]any); len(done) != 1 || done[0] != f.prefix+"m1" {
		t.Fatalf("mission completed: %v", res)
	}
	if bal := plugintest.Balance(t, f.db, u)[f.coin]; bal != 500 {
		t.Fatalf("mission reward: %d", bal)
	}
	ms := f.must(f.s.rpcMissions, u, map[string]any{"key": f.prefix + "m1"})
	if ms["rewarded"] != true || len(ms["items"].([]any)) != 2 {
		t.Fatalf("mission view: %v", ms)
	}

	f.settings(func(s *settings.All) { s.Limits.DailyPlays = 2 })
	u2 := f.user()
	f.must(f.s.rpcPlay, u2, map[string]any{"packKey": f.prefix + "ma"})
	if _, err := f.rpc(f.s.rpcPlay, u2, map[string]any{"packKey": f.prefix + "mb"}); code(err) != "QUIZ_MISSION_PACK_LOCKED" {
		t.Fatalf("lock first: %v", err)
	}
	f.settings(func(s *settings.All) { s.Limits.DailyPlays = 1 })
	solo := f.pack("solo", map[string]any{"campaignStageKey": ""})
	f.question(solo, nil)
	if _, err := f.rpc(f.s.rpcPlay, u2, map[string]any{"packKey": f.prefix + "solo"}); code(err) != "QUIZ_DAILY_PLAY_LIMIT" {
		t.Fatalf("daily limit: %v", err)
	}

	f.settings(func(*settings.All) {})
	stage := f.prefix + "-stage"
	f.db.Exec(`UPDATE quiz_pack SET campaign_stage_key=$2 WHERE id=$1`, solo, stage)
	v = f.must(f.s.rpcPlay, u2, map[string]any{"packKey": f.prefix + "solo"})
	if v["campaignAttemptId"] != "attempt-"+stage {
		t.Fatalf("campaign attempt: %v", v)
	}
	var qid string
	_ = f.db.QueryRow(`SELECT id::text FROM quiz_question WHERE pack_id=$1`, solo).Scan(&qid)
	out = f.must(f.s.rpcAnswer, u2, map[string]any{"sessionId": v["sessionId"], "questionId": qid, "optionId": 2})
	f.must(f.s.rpcFinish, u2, map[string]any{"sessionId": v["sessionId"]})
	if len(f.completes) != 1 || !f.completes[0].Success || f.completes[0].AttemptID != "attempt-"+stage || out["result"].(map[string]any)["campaign"] == nil {
		t.Fatalf("campaign completion: %v %v", f.completes, out)
	}
	if ref, ok := f.s.ownsStage(context.Background(), stage); !ok || ref != f.prefix+"solo" {
		t.Fatalf("owner hook: %v %v", ref, ok)
	}
}

func TestPostgresPoolCategoryAndRetries(t *testing.T) {
	f := newFixture(t)
	cat := f.prefix + "-bank"
	for i := 0; i < 6; i++ {
		f.question("", map[string]any{"category": cat})
	}
	f.pack("pool", map[string]any{"poolCategory": cat, "questionCount": 4, "passCorrect": 3})
	f.settings(func(a *settings.All) { a.Limits.MaxPlaysPerPack = 1; a.Play.ShuffleQuestions = true })
	u := f.user()
	v := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "pool"})
	if len(questionsOf(v)) != 4 || v["passCorrect"].(float64) != 3 {
		t.Fatalf("drawn from the bank: %v", v)
	}
	f.must(f.s.rpcFinish, u, map[string]any{"sessionId": v["sessionId"]})
	if _, err := f.rpc(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "pool"}); code(err) != "QUIZ_PACK_PLAY_LIMIT" {
		t.Fatalf("play limit: %v", err)
	}
	packs := f.must(f.s.rpcPacks, u, map[string]any{})
	found := false
	for _, p := range packs["packs"].([]any) {
		if p.(map[string]any)["key"] == f.prefix+"pool" {
			found = true
			if p.(map[string]any)["questionCount"].(float64) != 4 || p.(map[string]any)["plays"].(float64) != 1 {
				t.Fatalf("pack view: %v", p)
			}
		}
	}
	if !found {
		t.Fatal("pack listed")
	}
}

func TestPostgresReportsAndReview(t *testing.T) {
	f := newFixture(t)
	k := f.pack("rep", map[string]any{})
	q := f.question(k, nil)
	u := f.user()
	if _, err := f.rpc(f.s.rpcReportQuestion, u, map[string]any{"questionId": q, "reasonKey": "typo"}); code(err) != "QUIZ_REPORT_NOT_SEEN" {
		t.Fatalf("not seen: %v", err)
	}
	f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "rep"})
	if _, err := f.rpc(f.s.rpcReportQuestion, u, map[string]any{"questionId": q, "reasonKey": "nope"}); code(err) != "QUIZ_REPORT_REASON_NOT_FOUND" {
		t.Fatalf("reason: %v", err)
	}
	r1 := f.must(f.s.rpcReportQuestion, u, map[string]any{"questionId": q, "reasonKey": "wrong_answer", "note": "پاسخ ب است"})
	r2 := f.must(f.s.rpcReportQuestion, u, map[string]any{"questionId": q, "reasonKey": "typo"})
	if r1["duplicate"] != false || r2["duplicate"] != true || r1["reportId"] != r2["reportId"] {
		t.Fatalf("one open report per player and question: %v %v", r1, r2)
	}
	st, list := f.console(f.s.reportList, http.MethodGet, "/?questionId="+q, nil, nil)
	if st != 200 || list["total"].(float64) != 1 {
		t.Fatalf("console list: %v", list)
	}
	st, _ = f.console(f.s.reportReview, http.MethodPut, "/", map[string]any{"status": "resolved", "resolution": "fixed", "disableQuestion": true}, map[string]string{"id": r1["reportId"].(string)})
	if st != 200 {
		t.Fatal("review")
	}
	var status string
	var active bool
	_ = f.db.QueryRow(`SELECT r.status, q.is_active FROM quiz_question_report r JOIN quiz_question q ON q.id=r.question_id WHERE r.id=$1`, r1["reportId"]).Scan(&status, &active)
	if status != "resolved" || active {
		t.Fatalf("resolved and disabled: %s %v", status, active)
	}
	f.settings(func(a *settings.All) { a.Reports.DailyLimit = 1 })
	if _, err := f.rpc(f.s.rpcReportQuestion, u, map[string]any{"questionId": q, "reasonKey": "typo"}); code(err) != "QUIZ_REPORT_LIMIT" {
		t.Fatalf("daily report limit: %v", err)
	}
}

func TestPostgresImportExport(t *testing.T) {
	f := newFixture(t)
	doc := map[string]any{
		"missions": []map[string]any{{"key": f.prefix + "im", "name": map[string]string{"fa": "ماموریت"}}},
		"packs":    []map[string]any{{"key": f.prefix + "ip", "missionKey": f.prefix + "im", "name": map[string]string{"fa": "پک"}, "questionCount": 1}},
		"questions": []map[string]any{{"key": f.prefix + "q1", "packKey": f.prefix + "ip", "question": map[string]string{"fa": "؟"},
			"options": []map[string]any{{"id": 0, "text": map[string]string{"fa": "a"}, "correct": true}, {"id": 1, "text": map[string]string{"fa": "b"}}},
			"hints":   []map[string]any{{"name": map[string]string{"fa": "h"}, "text": map[string]string{"fa": "t"}}}}},
		"rows": []map[string]string{{"Pack Key": f.prefix + "ip", "Type": "text", "Question": "Largest ocean?", "Accepted Answers": "Pacific", "Answer Language": "en"}},
	}
	bad := map[string]any{"questions": []map[string]any{{"packKey": f.prefix + "ip", "question": map[string]string{"fa": "x"}, "options": []map[string]any{{"id": 0, "text": map[string]string{"fa": "a"}}}}}}
	st, out := f.console(f.s.importPost, http.MethodPost, "/", bad, nil)
	if st != 400 || out["error"] != "QUIZ_IMPORT_HAS_ERRORS" {
		t.Fatalf("errors refuse everything: %d %v", st, out)
	}
	st, out = f.console(f.s.importPost, http.MethodPost, "/", doc, nil)
	if st != 200 {
		t.Fatalf("import: %d %v", st, out)
	}
	c := out["counts"].(map[string]any)
	if c["questions"].(map[string]any)["created"].(float64) != 2 || c["packs"].(map[string]any)["created"].(float64) != 1 {
		t.Fatalf("counts: %v", c)
	}
	st, out = f.console(f.s.importPost, http.MethodPost, "/", map[string]any{"questions": doc["questions"]}, nil)
	if st != 200 || out["counts"].(map[string]any)["questions"].(map[string]any)["updated"].(float64) != 1 {
		t.Fatalf("a question with a known key is updated: %v", out)
	}
	var pid string
	_ = f.db.QueryRow(`SELECT id::text FROM quiz_pack WHERE key=$1`, f.prefix+"ip").Scan(&pid)
	st, exp := f.console(f.s.exportGet, http.MethodGet, "/?packId="+pid, nil, nil)
	qs := exp["questions"].([]any)
	if st != 200 || len(qs) != 2 || len(exp["missions"].([]any)) != 1 {
		t.Fatalf("export: %v", exp)
	}
	var hints int
	_ = f.db.QueryRow(`SELECT count(*) FROM quiz_hint h JOIN quiz_question q ON q.id=h.question_id WHERE q.key=$1`, f.prefix+"q1").Scan(&hints)
	if hints != 1 {
		t.Fatalf("hints imported once: %d", hints)
	}

	u := f.user()
	f.settings(func(a *settings.All) { a.Scoring.PassPercent = 0 })
	v := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "ip"})
	if len(questionsOf(v)) != 1 {
		t.Fatalf("question count 1: %v", v)
	}
	q := questionsOf(v)[0]
	in := map[string]any{"sessionId": v["sessionId"], "questionId": q["id"], "optionId": 0}
	if q["type"] == "text" {
		in = map[string]any{"sessionId": v["sessionId"], "questionId": q["id"], "text": " pacific "}
	}
	if a := f.must(f.s.rpcAnswer, u, in); a["outcome"] != "correct" {
		t.Fatalf("answer: %v", a)
	}
}

func TestPostgresResultReport(t *testing.T) {
	f := newFixture(t)
	m := f.mission("rm", map[string]any{})
	k := f.pack("rp", map[string]any{"missionId": m})
	q := f.question(k, nil)
	st, tpl := f.console(f.s.templateSave, http.MethodPost, "/", map[string]any{"key": f.prefix + "tpl", "name": "result", "missionId": m, "body": "{user.username}: {status} {score} {quiz.correct}/{quiz.total}",
		"wonText": "قبول", "lostText": "مردود", "webhookUrl": "https://example.com/hook"}, nil)
	if st != 200 {
		t.Fatalf("template: %v", tpl)
	}
	if st, bad := f.console(f.s.templateSave, http.MethodPost, "/", map[string]any{"key": f.prefix + "tpl2", "name": "x", "body": "y", "webhookUrl": "http://127.0.0.1/x"}, nil); st != 400 || bad["error"] != "QUIZ_REPORT_WEBHOOK_URL_IS_INVALID" {
		t.Fatalf("private webhook refused: %v", bad)
	}
	u := f.user()
	v := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "rp"})
	f.must(f.s.rpcAnswer, u, map[string]any{"sessionId": v["sessionId"], "questionId": q, "optionId": 2})
	out := f.must(f.s.rpcResultReport, u, map[string]any{"templateKey": f.prefix + "tpl"})
	body, _ := out["body"].(string)
	if !bytes.Contains([]byte(body), []byte(": قبول 100 1/1")) {
		t.Fatalf("rendered: %q", body)
	}
	f.s.deliverWebhooks(context.Background(), 10)
	var status string
	_ = f.db.QueryRow(`SELECT status FROM quiz_result_delivery WHERE user_id=$1`, u).Scan(&status)
	if status != "sent" {
		t.Fatalf("delivery: %s", status)
	}
}

func TestPostgresExpiredSessionsAndPaymentRetry(t *testing.T) {
	f := newFixture(t)
	k := f.pack("exp", map[string]any{"timeSeconds": 30, "passPercent": 0, "reward": map[string]any{"currencies": map[string]int64{f.coin: 10}}})
	f.question(k, nil)
	u := f.user()
	v := f.must(f.s.rpcPlay, u, map[string]any{"packKey": f.prefix + "exp"})
	f.wallet.Fail = true
	_, _ = f.db.Exec(`UPDATE quiz_session SET deadline_at=now() - interval '1 hour', expires_at=now() - interval '1 minute' WHERE id=$1`, v["sessionId"])
	f.s.closeExpired(context.Background(), 10)
	var status string
	_ = f.db.QueryRow(`SELECT status FROM quiz_session WHERE id=$1`, v["sessionId"]).Scan(&status)
	if status != "paying" {
		t.Fatalf("wallet down leaves paying: %s", status)
	}
	f.wallet.Fail = false
	_, _ = f.db.Exec(`UPDATE quiz_session SET completed_at=now() - interval '1 minute' WHERE id=$1`, v["sessionId"])
	f.s.retryPayments(context.Background(), 10)
	_ = f.db.QueryRow(`SELECT status FROM quiz_session WHERE id=$1`, v["sessionId"]).Scan(&status)
	if status != "done" || plugintest.Balance(t, f.db, u)[f.coin] != 10 {
		t.Fatalf("retried: %s %v", status, plugintest.Balance(t, f.db, u))
	}
	f.s.retryPayments(context.Background(), 10)
	if plugintest.Balance(t, f.db, u)[f.coin] != 10 {
		t.Fatal("paid once")
	}
}
