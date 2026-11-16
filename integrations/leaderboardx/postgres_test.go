package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type fakeCore struct {
	db  *sql.DB
	now func() time.Time
}

func (c *fakeCore) Create(ctx context.Context, id, sortOrder, operator, sched string, _ map[string]any) error {
	so := map[string]int{"asc": 0, "desc": 1}[sortOrder]
	op := map[string]int{"best": 0, "set": 1, "incr": 2, "decr": 3}[operator]
	var rs any
	if sched != "" {
		rs = sched
	}
	_, err := c.db.ExecContext(ctx, `INSERT INTO leaderboard (id, authoritative, sort_order, operator, reset_schedule, metadata, enable_ranks) VALUES ($1,true,$2,$3,$4,'{}',true) ON CONFLICT DO NOTHING`, id, so, op, rs)
	return err
}

func (c *fakeCore) Delete(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, `DELETE FROM leaderboard WHERE id=$1`, id)
	return err
}

func (c *fakeCore) Get(context.Context, string) (*api.Leaderboard, error) { return nil, nil }

func (c *fakeCore) board(ctx context.Context, id string) (string, string, time.Time, error) {
	var so, op int
	var sched string
	if err := c.db.QueryRowContext(ctx, `SELECT sort_order, operator, COALESCE(reset_schedule,'') FROM leaderboard WHERE id=$1`, id).Scan(&so, &op, &sched); err != nil {
		return "", "", time.Time{}, err
	}
	return []string{"asc", "desc"}[so], []string{"best", "set", "incr", "decr"}[op], expiry(sched, c.now()), nil
}

func (c *fakeCore) Write(ctx context.Context, id, ownerID, username string, score, sub int64, override *int) (*api.LeaderboardRecord, error) {
	so, op, exp, err := c.board(ctx, id)
	if err != nil {
		return nil, err
	}
	if override != nil {
		op = map[int]string{1: "best", 2: "set", 3: "incr", 4: "decr"}[*override]
	}
	upd := map[string]string{
		"incr": "score=r.score+EXCLUDED.score, subscore=r.subscore+EXCLUDED.subscore",
		"decr": "score=GREATEST(r.score-EXCLUDED.score,0), subscore=GREATEST(r.subscore-EXCLUDED.subscore,0)",
		"set":  "score=EXCLUDED.score, subscore=EXCLUDED.subscore",
		"best": "score=GREATEST(r.score,EXCLUDED.score), subscore=GREATEST(r.subscore,EXCLUDED.subscore)",
	}[op]
	if op == "best" && so == "asc" {
		upd = "score=LEAST(r.score,EXCLUDED.score), subscore=LEAST(r.subscore,EXCLUDED.subscore)"
	}
	if _, err := c.db.ExecContext(ctx, `INSERT INTO leaderboard_record AS r (leaderboard_id, owner_id, username, score, subscore, metadata, expiry_time) VALUES ($1,$2,$3,$4,$5,'{}',$6)
		ON CONFLICT (owner_id, leaderboard_id, expiry_time) DO UPDATE SET `+upd+`, num_score=r.num_score+1, update_time=now()`, id, ownerID, username, score, sub, exp); err != nil {
		return nil, err
	}
	rank, s, ss, _ := rankOf(ctx, c.db, id, so, exp, ownerID)
	return &api.LeaderboardRecord{LeaderboardId: id, OwnerId: ownerID, Username: wrapperspb.String(username), Score: s, Subscore: ss, Rank: int64(rank)}, nil
}

func (c *fakeCore) RecordDelete(ctx context.Context, id, owner string) error {
	_, _, exp, err := c.board(ctx, id)
	if err != nil {
		return err
	}
	_, err = c.db.ExecContext(ctx, `DELETE FROM leaderboard_record WHERE leaderboard_id=$1 AND owner_id=$2 AND expiry_time=$3`, id, owner, exp)
	return err
}

func (c *fakeCore) page(ctx context.Context, id string, offset, limit int) ([]*api.LeaderboardRecord, error) {
	so, _, exp, err := c.board(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := c.db.QueryContext(ctx, `SELECT owner_id::text, username, score, subscore FROM leaderboard_record WHERE leaderboard_id=$1 AND expiry_time=$2 ORDER BY `+coreOrder(so)+` OFFSET $3 LIMIT $4`, id, exp, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*api.LeaderboardRecord{}
	for rows.Next() {
		var r api.LeaderboardRecord
		var name string
		if err := rows.Scan(&r.OwnerId, &name, &r.Score, &r.Subscore); err != nil {
			return nil, err
		}
		r.Username, r.Rank, r.LeaderboardId = wrapperspb.String(name), int64(offset+len(out)+1), id
		out = append(out, &r)
	}
	return out, nil
}

func (c *fakeCore) List(ctx context.Context, id string, owners []string, limit int, cursor string) ([]*api.LeaderboardRecord, []*api.LeaderboardRecord, string, error) {
	off, _ := strconv.Atoi(cursor)
	recs, err := c.page(ctx, id, off, limit)
	if err != nil {
		return nil, nil, "", err
	}
	own := []*api.LeaderboardRecord{}
	so, _, exp, _ := c.board(ctx, id)
	for _, o := range owners {
		if rank, s, ss, ok := rankOf(ctx, c.db, id, so, exp, o); ok {
			own = append(own, &api.LeaderboardRecord{OwnerId: o, Score: s, Subscore: ss, Rank: int64(rank)})
		}
	}
	next := ""
	if len(recs) == limit {
		next = strconv.Itoa(off + limit)
	}
	return recs, own, next, nil
}

func (c *fakeCore) Haystack(ctx context.Context, id, owner string, limit int) ([]*api.LeaderboardRecord, error) {
	so, _, exp, _ := c.board(ctx, id)
	rank, _, _, ok := rankOf(ctx, c.db, id, so, exp, owner)
	if !ok {
		return nil, nil
	}
	return c.page(ctx, id, max(0, rank-1-limit/2), limit)
}

type fixture struct {
	t      *testing.T
	db     *sql.DB
	s      *service
	wallet *plugintest.Wallet
	mu     sync.Mutex
	clock  time.Time
	prizes []economyhooks.PrizeGrant
	notes  []string
	coin   string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, db: db, wallet: &plugintest.Wallet{DB: db}, clock: time.Now().UTC().Truncate(time.Second)}
	f.s = &service{db: db, nk: f.wallet, logger: plugintest.Logger{}, nudge: make(chan struct{}, 1)}
	f.s.now = func() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.clock }
	f.s.core = &fakeCore{db: db, now: f.s.now}
	f.s.prize = func(_ context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if g.SourceType != sourceType {
			return economyhooks.PrizeResult{}, errors.New("wrong source type")
		}
		for _, p := range f.prizes {
			if p.SourceID == g.SourceID && p.PrizeKey == g.PrizeKey && p.GroupKey == g.GroupKey {
				return economyhooks.PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey + g.GroupKey}}}, nil
			}
		}
		f.prizes = append(f.prizes, g)
		return economyhooks.PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey + g.GroupKey}}}, nil
	}
	f.s.notify = func(_ context.Context, userID, subject string, _ map[string]any, _ int) {
		f.mu.Lock()
		f.notes = append(f.notes, userID+":"+subject)
		f.mu.Unlock()
	}
	f.coin = plugintest.Key("coin")
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM leaderboard WHERE id LIKE 'pt_%'`)
		_, _ = db.Exec(`DELETE FROM leaderboardx_period WHERE board_key LIKE 'pt_%'`)
		_, _ = db.Exec(`DELETE FROM leaderboardx_score_log WHERE board_id IN (SELECT id FROM leaderboardx_board WHERE key LIKE 'pt_%')`)
		_, _ = db.Exec(`DELETE FROM leaderboardx_board WHERE key LIKE 'pt_%'`)
		_, _ = db.Exec(`UPDATE leaderboardx_settings SET value='{}' WHERE id=1`)
		_, _ = db.Exec(`DELETE FROM cluster_rate_limit WHERE key LIKE 'lbx:%'`)
		settings.Invalidate()
		invalidateBoards()
	})
	f.settings(func(*settings.All) {})
	return f
}

func (f *fixture) settings(change func(*settings.All)) {
	f.t.Helper()
	cfg := settings.Defaults()
	cfg.Events.EconomyEvent, cfg.Events.AnalyticsEvent = "", ""
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

func (f *fixture) passWeek() {
	f.mu.Lock()
	f.clock = expiry("0 0 * * 1", f.clock).Add(time.Minute)
	f.mu.Unlock()
}

func (f *fixture) user() string {
	id := plugintest.NewUser(f.t, f.db)
	f.t.Cleanup(func() {
		_, _ = f.db.Exec(`DELETE FROM leaderboardx_result WHERE user_id=$1`, id)
		_, _ = f.db.Exec(`DELETE FROM wallet_ledger WHERE user_id=$1`, id)
		_, _ = f.db.Exec(`DELETE FROM user_activity WHERE user_id=$1`, id)
	})
	return id
}

func (f *fixture) console(method string, h http.HandlerFunc, vars map[string]string, body any) (int, map[string]any) {
	f.t.Helper()
	raw, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "/x", bytes.NewReader(raw))
	for k, v := range vars {
		r.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	h(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func (f *fixture) board(body map[string]any) (string, string) {
	f.t.Helper()
	if body["key"] == nil {
		body["key"] = plugintest.Key("pt_lb")
	}
	if body["name"] == nil {
		body["name"] = map[string]string{"fa": "جدول", "en": "Board"}
	}
	code, out := f.console(http.MethodPost, f.s.boardSave, nil, body)
	if code != http.StatusCreated {
		f.t.Fatalf("board = %d %v", code, out)
	}
	b := out["board"].(map[string]any)
	return b["id"].(string), b["key"].(string)
}

func (f *fixture) rpc(fn apierr.Handler, user string, in any) (map[string]any, error) {
	f.t.Helper()
	raw, _ := json.Marshal(in)
	out, err := fn(context.Background(), user, string(raw))
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(out)
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m, nil
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

func (f *fixture) balance(user string) int64 { return plugintest.Balance(f.t, f.db, user)[f.coin] }

func TestBoardEditorValidatesAndCreatesCoreLeaderboard(t *testing.T) {
	f := newFixture(t)
	id, key := f.board(map[string]any{"period": "weekly", "operator": "incr", "source": map[string]any{"type": "economy_event", "event": "xp_gain", "calc": "value", "multiplier": 1},
		"tierBasis": "score", "tiers": []map[string]any{{"key": "gold", "threshold": 100}, {"key": "silver", "threshold": 10}},
		"rewards": []map[string]any{{"fromRank": 1, "toRank": 1, "currencies": map[string]int64{f.coin: 10}}}})
	var sched string
	if err := f.db.QueryRow(`SELECT reset_schedule FROM leaderboard WHERE id=$1`, key).Scan(&sched); err != nil || sched != "0 0 * * 1" {
		t.Fatalf("core leaderboard: %q %v", sched, err)
	}
	if c, out := f.console(http.MethodPost, f.s.boardSave, nil, map[string]any{"key": key, "name": map[string]string{"fa": "x"}}); c != 400 || out["error"] != "KEY_EXISTS" {
		t.Fatalf("duplicate key: %d %v", c, out)
	}
	for name, body := range map[string]map[string]any{
		"tier order":  {"tierBasis": "score", "tiers": []map[string]any{{"key": "a", "threshold": 10}, {"key": "b", "threshold": 100}}},
		"count+best":  {"operator": "best", "source": map[string]any{"type": "client", "calc": "count", "multiplier": 1}},
		"bad cron":    {"period": "custom", "resetSchedule": "every day"},
		"tier reward": {"rewards": []map[string]any{{"tier": "nope", "fromRank": 1, "toRank": 1, "currencies": map[string]int64{f.coin: 1}}}},
		"open ranks":  {"rewards": []map[string]any{{"fromRank": 1, "toRank": 0, "currencies": map[string]int64{f.coin: 1}}}},
	} {
		body["key"], body["name"] = plugintest.Key("pt_bad"), map[string]string{"fa": "x"}
		if c, out := f.console(http.MethodPost, f.s.boardSave, nil, body); c != 400 {
			t.Fatalf("%s accepted: %d %v", name, c, out)
		}
	}
	if c, out := f.console(http.MethodGet, f.s.boardGet, map[string]string{"id": id}, nil); c != 200 || out["coreFound"] != true {
		t.Fatalf("get: %d %v", c, out)
	}
}

func TestEconomyAndGameEventsCountOnce(t *testing.T) {
	f := newFixture(t)
	_, sum := f.board(map[string]any{"source": map[string]any{"type": "economy_event", "event": "xp_gain", "calc": "value", "multiplier": 1}})
	_, cnt := f.board(map[string]any{"source": map[string]any{"type": "economy_event", "event": "xp_gain", "calc": "count", "multiplier": 1}})
	_, stage := f.board(map[string]any{"operator": "best", "source": map[string]any{"type": "game_event", "event": "stage_completed", "context": "s1", "field": "stars", "calc": "value", "multiplier": 1}})
	u := f.user()
	ctx := context.Background()
	f.s.onEconomyEvent(ctx, playerevents.Recorded{ID: "e1", UserID: u, Key: "xp_gain", Value: 30})
	f.s.onEconomyEvent(ctx, playerevents.Recorded{ID: "e1", UserID: u, Key: "xp_gain", Value: 30})
	f.s.onEconomyEvent(ctx, playerevents.Recorded{ID: "e2", UserID: u, Key: "xp_gain", Value: 70})
	f.s.onEconomyEvent(ctx, playerevents.Recorded{ID: "e3", UserID: u, Key: "other", Value: 70})
	_, _ = f.s.onGameEvent(ctx, gameevents.Event{ID: "a1", UserID: u, Name: "stage_completed", Context: "s1", Fields: map[string]any{"stars": 2.0}})
	_, _ = f.s.onGameEvent(ctx, gameevents.Event{ID: "a2", UserID: u, Name: "stage_completed", Context: "s2", Fields: map[string]any{"stars": 3.0}})
	_, _ = f.s.onGameEvent(ctx, gameevents.Event{ID: "a3", UserID: u, Name: "stage_completed", Context: "s1", Fields: map[string]any{"stars": 1.0}})
	for key, want := range map[string]int64{sum: 100, cnt: 2, stage: 2} {
		var got int64
		if err := f.db.QueryRow(`SELECT score FROM leaderboard_record WHERE leaderboard_id=$1 AND owner_id=$2`, key, u).Scan(&got); err != nil || got != want {
			t.Fatalf("%s: score %d (%v), want %d", key, got, err, want)
		}
	}
}

func TestClientSubmitPolicies(t *testing.T) {
	f := newFixture(t)
	_, closed := f.board(map[string]any{"source": map[string]any{"type": "server", "calc": "value", "multiplier": 1}})
	_, open := f.board(map[string]any{"operator": "best", "source": map[string]any{"type": "client", "calc": "value", "multiplier": 1},
		"limits": map[string]any{"maxPerWrite": 500, "perMinute": 3, "onExceed": "reject"}})
	u := f.user()
	if _, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": closed, "score": 5, "requestId": "r0"}); code(err) != "LEADERBOARD_SUBMIT_NOT_ALLOWED" {
		t.Fatalf("server board: %v", err)
	}
	if _, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": open, "score": 5}); code(err) != "REQUEST_ID_REQUIRED" {
		t.Fatalf("request id: %v", err)
	}
	if _, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": open, "score": 900, "requestId": "r1"}); code(err) != "LEADERBOARD_SCORE_TOO_HIGH" {
		t.Fatalf("too high: %v", err)
	}
	out, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": open, "score": 300, "requestId": "r2"})
	if err != nil || out["accepted"] != true || out["score"] != 300.0 || out["rank"] != 1.0 {
		t.Fatalf("submit: %v %v", out, err)
	}
	out, err = f.rpc(f.s.rpcSubmit, u, map[string]any{"key": open, "score": 300, "requestId": "r2"})
	if err != nil || out["replayed"] != true || out["accepted"] != false {
		t.Fatalf("resend: %v %v", out, err)
	}
	if _, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": open, "score": 100, "requestId": "r3"}); err != nil {
		t.Fatal(err)
	}

	if _, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": open, "score": 400, "requestId": "r4"}); code(err) != "LEADERBOARD_RATE_LIMITED" {
		t.Fatalf("rate: %v", err)
	}
	var best int64
	_ = f.db.QueryRow(`SELECT score FROM leaderboard_record WHERE leaderboard_id=$1 AND owner_id=$2`, open, u).Scan(&best)
	if best != 300 {
		t.Fatalf("best = %d", best)
	}
	var refused int
	_ = f.db.QueryRow(`SELECT count(*) FROM leaderboardx_score_log l JOIN leaderboardx_board b ON b.id=l.board_id WHERE b.key=$1 AND l.outcome IN ('too_high','rate_limited')`, open).Scan(&refused)
	if refused != 2 {
		t.Fatalf("refused log = %d", refused)
	}
}

func (f *fixture) periodBoard(mode string) (string, string, []string) {
	id, key := f.board(map[string]any{"period": "weekly", "payMode": mode, "source": map[string]any{"type": "client", "calc": "value", "multiplier": 1},
		"tierBasis": "score", "tiers": []map[string]any{{"key": "gold", "threshold": 100}, {"key": "silver", "threshold": 0}},
		"rewards": []map[string]any{
			{"fromRank": 1, "toRank": 1, "currencies": map[string]int64{f.coin: 100}, "prizeKey": ""},
			{"tier": "silver", "fromRank": 1, "toRank": 1, "currencies": map[string]int64{f.coin: 7}},
		}})
	plugintest.Exec(f.t, f.db, `UPDATE leaderboardx_board SET created_at = now() - interval '30 days' WHERE id=$1`, id)
	invalidateBoards()
	users := []string{f.user(), f.user(), f.user()}
	for i, sc := range []int{250, 50, 20} {
		if _, err := f.rpc(f.s.rpcSubmit, users[i], map[string]any{"key": key, "score": sc, "requestId": "p" + strconv.Itoa(i)}); err != nil {
			f.t.Fatal(err)
		}
	}
	return id, key, users
}

func TestPeriodEndArchivesRanksTiersAndPaysOnce(t *testing.T) {
	f := newFixture(t)
	_, key, u := f.periodBoard("auto")
	view, err := f.rpc(f.s.rpcBoard, u[1], map[string]any{"key": key})
	if err != nil {
		t.Fatal(err)
	}
	me := view["me"].(map[string]any)
	if me["rank"] != 2.0 || me["tier"] != "silver" || me["nextTier"].(map[string]any)["need"] != 50.0 || len(view["top"].([]any)) != 3 {
		t.Fatalf("view: %v", view)
	}
	f.passWeek()
	if err := f.s.closeDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := f.s.closeDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	var periods, players, rewarded int
	_ = f.db.QueryRow(`SELECT count(*), max(players), max(rewarded) FROM leaderboardx_period WHERE board_key=$1`, key).Scan(&periods, &players, &rewarded)
	if periods != 1 || players != 3 || rewarded != 2 {
		t.Fatalf("periods=%d players=%d rewarded=%d", periods, players, rewarded)
	}
	if f.balance(u[0]) != 100 || f.balance(u[1]) != 7 || f.balance(u[2]) != 0 {
		t.Fatalf("balances %d %d %d", f.balance(u[0]), f.balance(u[1]), f.balance(u[2]))
	}
	f.s.retryPayments(context.Background(), 10)
	var pid string
	_ = f.db.QueryRow(`SELECT id::text FROM leaderboardx_period WHERE board_key=$1`, key).Scan(&pid)
	out, err := f.s.payResult(context.Background(), pid, u[0], false)
	if err != nil || out["replayed"] != true || f.balance(u[0]) != 100 {
		t.Fatalf("second pay: %v %v %d", out, err, f.balance(u[0]))
	}
	var tier string
	var tierRank int
	_ = f.db.QueryRow(`SELECT tier, tier_rank FROM leaderboardx_result WHERE period_id=$1 AND user_id=$2`, pid, u[2]).Scan(&tier, &tierRank)
	if tier != "silver" || tierRank != 2 {
		t.Fatalf("third player tier %s/%d", tier, tierRank)
	}
	hist, err := f.rpc(f.s.rpcHistory, u[0], map[string]any{"key": key})
	if err != nil || len(hist["items"].([]any)) != 1 || hist["items"].([]any)[0].(map[string]any)["status"] != "paid" {
		t.Fatalf("history: %v %v", hist, err)
	}
	win, err := f.rpc(f.s.rpcWinners, u[2], map[string]any{"key": key})
	if err != nil || len(win["winners"].([]any)) != 3 {
		t.Fatalf("winners: %v %v", win, err)
	}
	if len(f.notes) != 2 {
		t.Fatalf("notifications %v", f.notes)
	}

	view, _ = f.rpc(f.s.rpcBoard, u[0], map[string]any{"key": key})
	if len(view["top"].([]any)) != 0 || view["me"] != nil {
		t.Fatalf("new period: %v", view)
	}
}

func TestClaimModeAndExpiry(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) { a.Rewards.ClaimDays = 2 })
	_, key, u := f.periodBoard("claim")
	f.passWeek()
	if err := f.s.closeDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.balance(u[0]) != 0 {
		t.Fatal("claim mode paid at once")
	}
	list, _ := f.rpc(f.s.rpcList, u[0], map[string]any{})
	found := false
	for _, b := range list["boards"].([]any) {
		if m := b.(map[string]any); m["key"] == key {
			found = m["claimable"] == 1.0
		}
	}
	if !found {
		t.Fatalf("list claimable: %v", list)
	}
	out, err := f.rpc(f.s.rpcClaim, u[0], map[string]any{"all": true})
	if err != nil || len(out["claimed"].([]any)) != 1 || f.balance(u[0]) != 100 {
		t.Fatalf("claim: %v %v %d", out, err, f.balance(u[0]))
	}
	var pid string
	_ = f.db.QueryRow(`SELECT id::text FROM leaderboardx_period WHERE board_key=$1`, key).Scan(&pid)
	again, err := f.rpc(f.s.rpcClaim, u[0], map[string]any{"periodId": pid})
	if err != nil || again["claimed"].([]any)[0].(map[string]any)["replayed"] != true || f.balance(u[0]) != 100 {
		t.Fatalf("claim again: %v %v", again, err)
	}
	if _, err := f.rpc(f.s.rpcClaim, u[2], map[string]any{"periodId": pid}); code(err) != "LEADERBOARD_REWARD_NOT_CLAIMABLE" {
		t.Fatalf("no reward: %v", err)
	}
	f.advance(3 * 24 * time.Hour)
	if _, err := f.rpc(f.s.rpcClaim, u[1], map[string]any{"periodId": pid}); code(err) != "LEADERBOARD_REWARD_EXPIRED" {
		t.Fatalf("expired: %v", err)
	}
}

func TestFailedPaymentRetriesThenAdminRetry(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) { a.Rewards.MaxTries = 2 })
	_, key, u := f.periodBoard("auto")
	f.wallet.Fail = true
	f.passWeek()
	if err := f.s.closeDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	f.s.retryPayments(context.Background(), 10)
	var pid, status string
	_ = f.db.QueryRow(`SELECT id::text FROM leaderboardx_period WHERE board_key=$1`, key).Scan(&pid)
	_ = f.db.QueryRow(`SELECT status FROM leaderboardx_result WHERE period_id=$1 AND user_id=$2`, pid, u[0]).Scan(&status)
	if status != "failed" || f.balance(u[0]) != 0 {
		t.Fatalf("status %s balance %d", status, f.balance(u[0]))
	}
	f.wallet.Fail = false
	c, out := f.console(http.MethodPost, f.s.periodRetry, map[string]string{"id": pid}, nil)
	if c != 200 || out["paid"] != 2.0 || f.balance(u[0]) != 100 {
		t.Fatalf("retry: %d %v %d", c, out, f.balance(u[0]))
	}
	c, out = f.console(http.MethodGet, f.s.periodGet, map[string]string{"id": pid}, nil)
	if c != 200 || len(out["items"].([]any)) != 3 {
		t.Fatalf("period: %d %v", c, out)
	}
}

func TestScheduleChangeNeedsRestartAndManualClose(t *testing.T) {
	f := newFixture(t)
	id, key, u := f.periodBoard("auto")
	body := map[string]any{"name": map[string]string{"fa": "جدول"}, "period": "daily", "source": map[string]any{"type": "client", "calc": "value", "multiplier": 1},
		"rewards": []map[string]any{{"fromRank": 1, "toRank": 1, "currencies": map[string]int64{f.coin: 100}}}}
	if c, out := f.console(http.MethodPut, f.s.boardSave, map[string]string{"id": id}, body); c != 409 || out["error"] != "LEADERBOARD_RESTART_REQUIRED" {
		t.Fatalf("restart required: %d %v", c, out)
	}
	body["restart"], body["payOnRestart"] = true, false
	if c, out := f.console(http.MethodPut, f.s.boardSave, map[string]string{"id": id}, body); c != 200 {
		t.Fatalf("restart: %d %v", c, out)
	}
	var reason string
	var rewarded int
	_ = f.db.QueryRow(`SELECT reason, rewarded FROM leaderboardx_period WHERE board_key=$1`, key).Scan(&reason, &rewarded)
	var sched string
	_ = f.db.QueryRow(`SELECT reset_schedule FROM leaderboard WHERE id=$1`, key).Scan(&sched)
	if reason != "restart" || rewarded != 0 || sched != "0 0 * * *" || f.balance(u[0]) != 0 {
		t.Fatalf("after restart: %s %d %q", reason, rewarded, sched)
	}
	if _, err := f.rpc(f.s.rpcSubmit, u[1], map[string]any{"key": key, "score": 40, "requestId": "after"}); err != nil {
		t.Fatal(err)
	}
	f.advance(time.Minute)
	c, out := f.console(http.MethodPost, f.s.boardClose, map[string]string{"id": id}, map[string]any{"pay": true})
	if c != 200 {
		t.Fatalf("close: %d %v", c, out)
	}
	if f.balance(u[1]) != 100 {
		t.Fatalf("manual close paid %d", f.balance(u[1]))
	}
	var left int
	_ = f.db.QueryRow(`SELECT count(*) FROM leaderboard_record WHERE leaderboard_id=$1`, key).Scan(&left)
	if left != 0 {
		t.Fatalf("records left %d", left)
	}
	c, out = f.console(http.MethodGet, f.s.playerGet, map[string]string{"id": u[1]}, nil)
	if c != 200 || len(out["history"].([]any)) != 2 {
		t.Fatalf("player card: %d %v", c, out)
	}
	if c, out := f.console(http.MethodGet, f.s.report, nil, nil); c != 200 || out["totals"] == nil {
		t.Fatalf("report: %d %v", c, out)
	}
	if c, _ := f.console(http.MethodDelete, f.s.boardDelete, map[string]string{"id": id}, nil); c != 200 {
		t.Fatalf("delete: %d", c)
	}
	var core int
	_ = f.db.QueryRow(`SELECT count(*) FROM leaderboard WHERE id=$1`, key).Scan(&core)
	if core != 0 {
		t.Fatal("core leaderboard kept")
	}
}

func TestWindowEndClosesFinalPeriod(t *testing.T) {
	f := newFixture(t)
	end := f.clock.Add(2 * time.Hour)
	id, key := f.board(map[string]any{"endsAt": end.Format(time.RFC3339), "source": map[string]any{"type": "client", "calc": "value", "multiplier": 1},
		"tierBasis": "rank", "tiers": []map[string]any{{"key": "top", "threshold": 1}},
		"rewards": []map[string]any{{"tier": "top", "fromRank": 1, "toRank": 0, "currencies": map[string]int64{f.coin: 5}}}})
	u := f.user()
	if _, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": key, "score": 9, "requestId": "w1"}); err != nil {
		t.Fatal(err)
	}
	f.advance(3 * time.Hour)
	if _, err := f.rpc(f.s.rpcSubmit, u, map[string]any{"key": key, "score": 9, "requestId": "w2"}); code(err) != "LEADERBOARD_NOT_ACTIVE" {
		t.Fatalf("after the window: %v", err)
	}
	if err := f.s.closeDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	var reason string
	_ = f.db.QueryRow(`SELECT reason FROM leaderboardx_period WHERE board_key=$1`, key).Scan(&reason)
	b, _ := f.s.boardByID(context.Background(), id)
	if reason != "end" || b.FinishedAt == nil || f.balance(u) != 5 {
		t.Fatalf("end: %s %v %d", reason, b.FinishedAt, f.balance(u))
	}
	view, err := f.rpc(f.s.rpcBoard, u, map[string]any{"key": key})
	if err != nil || view["status"] != "ended" || view["me"].(map[string]any)["rank"] != 1.0 {
		t.Fatalf("ended view: %v %v", view, err)
	}
}
