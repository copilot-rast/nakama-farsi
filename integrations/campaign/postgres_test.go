package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type fixture struct {
	t      *testing.T
	db     *sql.DB
	s      *service
	wallet *plugintest.Wallet
	mu     sync.Mutex
	clock  time.Time
	levels map[string]int
	prizes []economyhooks.PrizeGrant
	coin   string
	meta   string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, db: db, wallet: &plugintest.Wallet{DB: db}, clock: time.Now().UTC().Truncate(time.Second), levels: map[string]int{}}
	f.s = &service{db: db, nk: f.wallet, logger: plugintest.Logger{}, spend: economyhooks.SpendRegen, refund: economyhooks.RefundRegen}
	f.s.now = func() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.clock }
	f.s.level = func(_ context.Context, userID, _ string) (int, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		return f.levels[userID], nil
	}
	f.s.prize = func(_ context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, p := range f.prizes {
			if p.SourceID == g.SourceID && p.PrizeKey == g.PrizeKey && p.GroupKey == g.GroupKey {
				return economyhooks.PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey + g.GroupKey}}}, nil
			}
		}
		f.prizes = append(f.prizes, g)
		return economyhooks.PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey + g.GroupKey}}}, nil
	}
	f.coin, f.meta = plugintest.Key("coin"), plugintest.Key("xp")
	plugintest.Exec(t, db, `INSERT INTO economy_meta (id, key, name) VALUES (gen_random_uuid(), $1, '{"fa":"تجربه"}')`, f.meta)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM campaign_stage WHERE chapter_id IN (SELECT id FROM campaign_chapter WHERE key LIKE 'pt_%')`)
		_, _ = db.Exec(`DELETE FROM campaign_chapter WHERE key LIKE 'pt_%'`)
		_, _ = db.Exec(`DELETE FROM campaign_freeride WHERE key LIKE 'pt_%'`)
		_, _ = db.Exec(`DELETE FROM economy_meta WHERE key=$1`, f.meta)
		_, _ = db.Exec(`UPDATE campaign_settings SET value='{}' WHERE id=1`)
		settings.Invalidate()
	})
	f.settings(func(a *settings.All) {})
	return f
}

func (f *fixture) settings(change func(*settings.All)) {
	f.t.Helper()
	cfg := settings.Defaults()
	cfg.XP.MetaKey = f.meta
	change(&cfg)
	if err := settings.Save(context.Background(), f.db, cfg, "test"); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) advance(d time.Duration) {
	f.mu.Lock()
	f.clock = f.clock.Add(d)
	f.mu.Unlock()
}

func (f *fixture) user() string {
	id := plugintest.NewUser(f.t, f.db)
	f.t.Cleanup(func() {
		for _, t := range []string{"campaign_attempt", "campaign_user_stage", "campaign_user", "campaign_user_chapter", "campaign_reward_group", "campaign_claim", "economy_user_meta", "wallet_ledger"} {
			_, _ = f.db.Exec(`DELETE FROM `+t+` WHERE user_id=$1`, id)
		}
	})
	return id
}

func (f *fixture) console(method string, h http.HandlerFunc, id string, body any) (int, map[string]any) {
	f.t.Helper()
	raw, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "/x", bytes.NewReader(raw))
	if id != "" {
		r.SetPathValue("id", id)
	}
	w := httptest.NewRecorder()
	h(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func (f *fixture) chapter(body map[string]any) string {
	f.t.Helper()
	if body["key"] == nil {
		body["key"] = plugintest.Key("pt_ch")
	}
	if body["name"] == nil {
		body["name"] = map[string]string{"fa": "فصل"}
	}
	body["isActive"] = true
	code, out := f.console(http.MethodPost, f.s.chapterSave, "", body)
	if code != 200 {
		f.t.Fatalf("chapter = %d %v", code, out)
	}
	return out["id"].(string)
}

func (f *fixture) stage(chapterID string, body map[string]any) string {
	f.t.Helper()
	key := plugintest.Key("pt_st")
	body["key"], body["chapterId"], body["isActive"] = key, chapterID, true
	if body["name"] == nil {
		body["name"] = map[string]string{"fa": "مرحله"}
	}
	code, out := f.console(http.MethodPost, f.s.stageSave, "", body)
	if code != 200 {
		f.t.Fatalf("stage = %d %v", code, out)
	}
	return key
}

func (f *fixture) start(user, stage string) string {
	f.t.Helper()
	out, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+stage+`"}`)
	if err != nil {
		f.t.Fatalf("start %s: %v", stage, err)
	}
	return out.(map[string]any)["attemptId"].(string)
}

func (f *fixture) complete(user, attempt string, success bool, elapsed float64, health any) (map[string]any, error) {
	raw, _ := json.Marshal(map[string]any{"attemptId": attempt, "success": success, "elapsedSeconds": elapsed, "healthPercent": health})
	out, err := f.s.rpcComplete(context.Background(), user, string(raw))
	if err != nil {
		return nil, err
	}
	return roundTrip(out), nil
}

func roundTrip(v any) map[string]any {
	raw, _ := json.Marshal(v)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func (f *fixture) play(user, stage string, elapsed float64, health any) map[string]any {
	f.t.Helper()
	a := f.start(user, stage)
	f.advance(time.Duration(elapsed*float64(time.Second)) + time.Second)
	out, err := f.complete(user, a, true, elapsed, health)
	if err != nil {
		f.t.Fatalf("complete %s: %v", stage, err)
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

func (f *fixture) metaValue(user string) float64 {
	var v float64
	_ = f.db.QueryRow(`SELECT COALESCE((SELECT value FROM economy_user_meta um JOIN economy_meta m ON m.id=um.meta_id WHERE um.user_id=$1 AND m.key=$2),0)::float8`, user, f.meta).Scan(&v)
	return v
}

func TestPostgresStarsDifferencePayoutResendAndUnlocks(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) { a.Stars.MissingField = settings.MissingPass })
	var published, starts []gameevents.Event
	gameevents.Subscribe("analytics", func(_ context.Context, ev gameevents.Event) (any, error) {
		if ev.Name == "" {
			starts = append(starts, ev)
			return nil, nil
		}
		published = append(published, ev)
		return []string{"trophy"}, nil
	})
	t.Cleanup(func() { gameevents.Subscribe("analytics", nil) })
	ch := f.chapter(map[string]any{})
	s1 := f.stage(ch, map[string]any{"mode": "ordered", "timeLimitSeconds": 60, "minSeconds": 20, "xp": 100, "rewards": map[string]any{"currencies": map[string]int64{f.coin: 1000}}})
	s2 := f.stage(ch, map[string]any{"mode": "ordered", "sort": 10, "rewards": map[string]any{"currencies": map[string]int64{f.coin: 10}}})
	user := f.user()

	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+s2+`"}`); code(err) != "STAGE_LOCKED" {
		t.Fatalf("second stage before the first = %v", err)
	}
	a := f.start(user, s1)
	f.advance(10 * time.Second)
	if _, err := f.complete(user, a, true, 10, 90); code(err) != "STAGE_FINISHED_TOO_FAST" {
		t.Fatalf("too fast = %v", err)
	}
	f.advance(35 * time.Second)
	if _, err := f.complete(user, a, true, 80, 90); code(err) != "INVALID_RESULT" {
		t.Fatalf("longer than the server clock = %v", err)
	}
	var rejected int
	_ = f.db.QueryRow(`SELECT rejected FROM campaign_attempt WHERE id=$1`, a).Scan(&rejected)
	if rejected != 2 {
		t.Fatalf("rejected = %d", rejected)
	}
	out, err := f.complete(user, a, true, 45, 80)
	if err != nil {
		t.Fatal(err)
	}
	if out["stars"] != 2.0 || out["xpGained"] != 50.0 || plugintest.Balance(t, f.db, user)[f.coin] != 500 || f.metaValue(user) != 50 {
		t.Fatalf("first win = %v wallet=%v meta=%v", out, plugintest.Balance(t, f.db, user), f.metaValue(user))
	}
	if len(published) != 1 || published[0].Context != s1 || published[0].Fields["stars"] != 2.0 || out["trophies"] == nil {
		t.Fatalf("events = %+v, trophies=%v", published, out["trophies"])
	}
	if ty := published[0].Typed; published[0].Source != "campaign" || len(ty) != 2 || ty[0].Status != "Complete" || ty[0].P2 != s1 || *ty[0].Score != 2 ||
		ty[1].Flow != "Source" || ty[1].Currency != f.coin || *ty[1].Amount != 500 {
		t.Fatalf("typed = %+v", published[0].Typed)
	}
	if len(starts) == 0 || starts[0].Typed[0].Status != "Start" || starts[0].Typed[0].P2 != s1 {
		t.Fatalf("start events = %+v", starts)
	}
	again, err := f.complete(user, a, true, 45, 80)
	if err != nil || again["replayed"] != true || again["stars"] != 2.0 || plugintest.Balance(t, f.db, user)[f.coin] != 500 || len(published) != 1 {
		t.Fatalf("resend = %v %v wallet=%v", again, err, plugintest.Balance(t, f.db, user))
	}
	f.advance(25 * time.Hour)
	if again, err := f.complete(user, a, true, 45, 80); err != nil || again["replayed"] != true {
		t.Fatalf("a finished attempt answers after its window too: %v %v", again, err)
	}

	f.play(user, s2, 5, nil)
	if out := f.play(user, s1, 30, 100); out["stars"] != 3.0 || out["rewards"].([]any)[0].(map[string]any)["amount"] != 500.0 || out["xpGained"] != 50.0 {
		t.Fatalf("3-star replay = %v", out)
	}
	if out := f.play(user, s1, 59, 100); out["stars"] != 1.0 || len(out["rewards"].([]any)) != 0 || out["xpGained"] != 0.0 || out["bestStars"] != 3.0 {
		t.Fatalf("worse replay = %v", out)
	}
	if w := plugintest.Balance(t, f.db, user)[f.coin]; w != 1010 || f.metaValue(user) != 100 {
		t.Fatalf("wallet=%d meta=%v", w, f.metaValue(user))
	}
	var kinds int
	_ = f.db.QueryRow(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'sourceType'='stage' AND metadata->>'reason'='stage_complete'`, user).Scan(&kinds)
	if kinds != 3 {
		t.Fatalf("ledger rows = %d", kinds)
	}

	a1 := f.start(user, s1)
	a2 := f.start(user, s2)
	if _, err := f.complete(user, a1, true, 1, nil); code(err) != "ATTEMPT_CLOSED" {
		t.Fatalf("closed = %v", err)
	}
	f.advance(25 * time.Hour)
	if _, err := f.complete(user, a2, true, 1, nil); code(err) != "ATTEMPT_EXPIRED" {
		t.Fatalf("expired = %v", err)
	}
	other := f.user()
	if _, err := f.complete(other, a1, true, 1, nil); code(err) != "ATTEMPT_NOT_FOUND" {
		t.Fatalf("another player's attempt = %v", err)
	}

	view := roundTrip(must(f.s.rpcChapters(context.Background(), user, `{}`)))
	var mine map[string]any
	for _, c := range view["chapters"].([]any) {
		if st := c.(map[string]any)["stages"].([]any); len(st) > 0 && st[0].(map[string]any)["key"] == s1 {
			mine = c.(map[string]any)
		}
	}
	if mine == nil || mine["stars"] != 6.0 || mine["unlocked"] != true || view["xp"] != 100.0 {
		t.Fatalf("chapters = %v", mine)
	}
}

func must(v any, err error) any {
	if err != nil {
		panic(err)
	}
	return v
}

func TestPostgresConcurrentCompletionsPayOnce(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) { a.Attempts.CloseOthers = false })
	ch := f.chapter(map[string]any{})
	st := f.stage(ch, map[string]any{"mode": "ordered", "xp": 10, "rewards": map[string]any{"currencies": map[string]int64{f.coin: 300}}})
	user := f.user()
	same := f.start(user, st)
	var others []string
	for i := 0; i < 4; i++ {
		others = append(others, f.start(user, st))
	}
	f.advance(20 * time.Second)
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := same
			if i < len(others) {
				id = others[i]
			}
			if _, err := f.complete(user, id, true, 10, 100); err != nil && code(err) != "STAGE_BUSY" {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if w := plugintest.Balance(t, f.db, user)[f.coin]; w != 300 || f.metaValue(user) != 10 {
		t.Fatalf("wallet=%d meta=%v (paid more than the full reward)", w, f.metaValue(user))
	}
}

func TestPostgresRepeatableStageAndFreerideCaps(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) { a.Rewards.DayZone = "UTC"; a.Freeride.TableEnabled = true })
	ch := f.chapter(map[string]any{})
	ride := f.stage(ch, map[string]any{"mode": "repeatable", "xp": 20, "cooldownSeconds": 600, "dailyLimit": 2, "rewardGroup": "pt_roam",
		"rewards": map[string]any{"currencies": map[string]int64{f.coin: 100}}})
	user := f.user()
	if out := f.play(user, ride, 5, 10); out["rewardLimited"] != nil || out["stars"] != 1.0 || plugintest.Balance(t, f.db, user)[f.coin] != 100 {
		t.Fatalf("first ride = %v", out)
	}
	if out := f.play(user, ride, 5, 10); out["rewardLimited"] != "cooldown" || out["xpGained"] != 0.0 || out["nextRewardAt"] == nil {
		t.Fatalf("second ride = %v", out)
	}
	f.advance(11 * time.Minute)
	f.play(user, ride, 5, 10)

	code_, fr := f.console(http.MethodPost, f.s.freerideSave, "", map[string]any{"key": plugintest.Key("pt_fr"), "name": map[string]string{"fa": "گردش"}, "isActive": true,
		"rewards": map[string]any{"currencies": map[string]int64{f.coin: 7}, "prizeKey": "pt_box"}, "xp": 3, "rewardGroup": "pt_roam", "cooldownSeconds": 600, "dailyLimit": 2})
	if code_ != 200 {
		t.Fatalf("freeride = %d %v", code_, fr)
	}
	key := fr["key"].(string)
	f.advance(11 * time.Minute)
	if _, err := f.s.claimFreeride(context.Background(), user, key, "r1"); code(err) != "REWARD_LIMITED" {
		t.Fatalf("daily cap shared with the stage = %v", err)
	}
	var e *apierr.Error
	_, err := f.s.claimFreeride(context.Background(), user, key, "r1")
	if !errors.As(err, &e) || e.Meta["reason"] != "daily_limit" || e.Meta["rewardsLeftToday"] != 0 {
		t.Fatalf("meta = %v", err)
	}
	f.advance(24 * time.Hour)
	other := f.user()
	if _, err := f.s.claimFreeride(context.Background(), other, key, ""); code(err) != "REQUEST_ID_REQUIRED" {
		t.Fatalf("no request id = %v", err)
	}
	out := roundTrip(must(f.s.claimFreeride(context.Background(), user, key, "r2")))
	again := roundTrip(must(f.s.claimFreeride(context.Background(), user, key, "r2")))
	if out["xpGained"] != 3.0 || again["replayed"] != true || plugintest.Balance(t, f.db, user)[f.coin] != 207 {
		t.Fatalf("claim = %v again=%v wallet=%v", out, again, plugintest.Balance(t, f.db, user))
	}
	if len(f.prizes) != 1 || f.prizes[0].SourceType != "stage" || f.prizes[0].PrizeKey != "pt_box" {
		t.Fatalf("prizes = %+v", f.prizes)
	}
	if _, err := f.s.claimFreeride(context.Background(), user, key, "r3"); code(err) != "REWARD_LIMITED" {
		t.Fatalf("cooldown after the claim = %v", err)
	}
	f.levels[other] = 1
	plugintest.Exec(t, f.db, `UPDATE campaign_freeride SET min_level=3 WHERE key=$1`, key)
	if _, err := f.s.claimFreeride(context.Background(), other, key, "x"); code(err) != "LEVEL_TOO_LOW" {
		t.Fatalf("min level = %v", err)
	}
}

func TestPostgresOnceStageTierPrizesMilestonesAndRules(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) {
		a.Unlock.ChapterDefault = settings.UnlockPrevious
		a.Stars.MissingField = settings.MissingPass
	})
	first := f.chapter(map[string]any{"sort": -2000000001, "milestones": []map[string]any{{"percent": 50, "currencies": map[string]int64{f.coin: 40}}, {"percent": 100, "prizeGroupKey": "pt_chest"}}})
	quiz := f.stage(first, map[string]any{"mode": "once", "rewards": map[string]any{"currencies": map[string]int64{f.coin: 5000}}})
	a := f.stage(first, map[string]any{"mode": "ordered", "sort": 1, "tierPrizes": []map[string]any{{"stars": 2, "prizeKey": "pt_silver"}, {"stars": 3, "prizeKey": "pt_gold"}}})
	second := f.chapter(map[string]any{"sort": -2000000000})
	b := f.stage(second, map[string]any{"mode": "ordered", "unlock": []map[string]any{{"type": "level", "value": 2}, {"type": "meta", "key": f.meta, "value": 1}}})
	user := f.user()

	f.play(user, quiz, 5, nil)
	if out := f.play(user, quiz, 5, nil); len(out["rewards"].([]any)) != 0 {
		t.Fatalf("a once stage paid again: %v", out)
	}
	if w := plugintest.Balance(t, f.db, user)[f.coin]; w != 5000 {
		t.Fatalf("quiz wallet = %d", w)
	}
	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+b+`"}`); code(err) != "STAGE_LOCKED" {
		t.Fatalf("next chapter before finishing = %v", err)
	}
	out := f.play(user, a, 50, 60)
	if out["stars"] != 2.0 || len(out["milestoneRewards"].([]any)) != 1 || len(f.prizes) != 1 || f.prizes[0].PrizeKey != "pt_silver" {
		t.Fatalf("tier = %v prizes=%+v", out, f.prizes)
	}
	out = f.play(user, a, 10, 100)
	if out["stars"] != 3.0 || len(f.prizes) != 3 || f.prizes[1].PrizeKey != "pt_gold" || f.prizes[2].GroupKey != "pt_chest" || len(out["milestoneRewards"].([]any)) != 0 {
		t.Fatalf("gold + 100%% milestone = %v prizes=%+v", out, f.prizes)
	}
	f.play(user, a, 10, 100)
	if len(f.prizes) != 3 {
		t.Fatalf("prizes paid again: %+v", f.prizes)
	}
	_, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+b+`"}`)
	var e *apierr.Error
	if !errors.As(err, &e) || e.Code != "STAGE_LOCKED" {
		t.Fatalf("level rule = %v", err)
	}
	own := 0
	for _, r := range e.Meta["locks"].([]Rule) {
		if r.Type == "level" || r.Type == "meta" && r.Key == f.meta {
			own++
		}
	}
	if own != 2 {
		t.Fatalf("locks = %v", e.Meta["locks"])
	}
	f.levels[user] = 2
	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+b+`"}`); code(err) != "STAGE_LOCKED" {
		t.Fatalf("meta rule = %v", err)
	}
	plugintest.Exec(t, f.db, `INSERT INTO economy_user_meta (user_id, meta_id, value) SELECT $1, id, 5 FROM economy_meta WHERE key=$2`, user, f.meta)
	f.start(user, b)
}

func TestPostgresStarDefaultsMissingHealthAndMinTime(t *testing.T) {
	f := newFixture(t)
	ch := f.chapter(map[string]any{"sort": 1})
	st := f.stage(ch, map[string]any{"mode": "repeatable", "xp": 10})
	user := f.user()
	a := f.start(user, st)
	if _, err := f.complete(user, a, true, 1, 100); code(err) != "STAGE_FINISHED_TOO_FAST" {
		t.Fatalf("zero-time run = %v", err)
	}
	f.advance(40 * time.Second)
	out, err := f.complete(user, a, true, 35, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out["stars"] != 1.0 {
		t.Fatalf("no health sent still got %v stars", out["stars"])
	}
	if _, ok := out["nextLevelXp"]; !ok {
		t.Fatalf("nextLevelXp missing: %v", out)
	}
}

func TestPostgresServerStartKeepsOtherAttemptsOpen(t *testing.T) {
	f := newFixture(t)
	ch := f.chapter(map[string]any{})
	mission := f.stage(ch, map[string]any{"mode": "repeatable", "rewards": map[string]any{"currencies": map[string]int64{f.coin: 10}}})
	quiz := f.stage(ch, map[string]any{"mode": "repeatable"})
	user := f.user()
	saved := f.start(user, mission)
	if _, err := f.s.hookStart(context.Background(), user, quiz); err != nil {
		t.Fatal(err)
	}
	f.advance(10 * time.Second)
	if _, err := f.complete(user, saved, true, 10, 100); err != nil {
		t.Fatalf("saved result after a server start = %v", err)
	}

	f.settings(func(a *settings.All) { a.Attempts.ServerClosesOthers = true })
	saved = f.start(user, mission)
	if _, err := f.s.hookStart(context.Background(), user, quiz); err != nil {
		t.Fatal(err)
	}
	f.advance(10 * time.Second)
	if _, err := f.complete(user, saved, true, 10, 100); code(err) != "ATTEMPT_CLOSED" {
		t.Fatalf("serverClosesOthers on = %v", err)
	}

	f.settings(func(a *settings.All) {})
	saved = f.start(user, mission)
	f.start(user, quiz)
	if _, err := f.complete(user, saved, true, 10, 100); code(err) != "ATTEMPT_CLOSED" {
		t.Fatalf("a client start still closes others = %v", err)
	}
}
