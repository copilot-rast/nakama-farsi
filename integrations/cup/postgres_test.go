package main

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/cuphooks"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type fixture struct {
	t        *testing.T
	s        *service
	db       *sql.DB
	clock    time.Time
	wallet   *plugintest.Wallet
	mu       sync.Mutex
	notes    map[string][]string
	grants   []economyhooks.PrizeGrant
	launches []cuphooks.Launch
	cups     []string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, db: db, clock: time.Now().UTC().Truncate(time.Second), notes: map[string][]string{}}
	f.wallet = &plugintest.Wallet{DB: db}
	s := newService(db, f.wallet, plugintest.Logger{})
	s.now = func() time.Time { return f.clock }
	s.rnd = func(int) int { return 0 }
	s.prize = func(_ context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.grants = append(f.grants, g)
		return economyhooks.PrizeResult{}, nil
	}
	s.notify = func(_ context.Context, n note) {
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, u := range n.users {
			f.notes[u] = append(f.notes[u], n.subject)
		}
	}
	s.launch = func(_ context.Context, l cuphooks.Launch) (cuphooks.Launched, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.launches = append(f.launches, l)
		return cuphooks.Launched{GameMatchID: "g-" + l.MatchID + "-" + strconv.Itoa(l.GameNo)}, nil
	}
	f.s = s
	if err := settings.Save(context.Background(), db, settings.Defaults(), "test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx := context.Background()
		for _, id := range f.cups {
			_, _ = db.ExecContext(ctx, `DELETE FROM cup WHERE id::text=$1`, id)
		}
		_, _ = db.ExecContext(ctx, `DELETE FROM cup WHERE template_key LIKE 'cpt_%'`)
		_, _ = db.ExecContext(ctx, `DELETE FROM cup_template WHERE key LIKE 'cpt_%'`)
	})
	return f
}

func (f *fixture) users(n int, coins int64) []string {
	out := []string{}
	for i := 0; i < n; i++ {
		u := plugintest.NewUser(f.t, f.db)
		if coins > 0 {
			f.wallet.Give(f.t, u, map[string]int64{"coin": coins}, nil)
		}
		out = append(out, u)
	}
	return out
}

func (f *fixture) rules() settings.Rules {
	r := settings.DefaultRules()
	r.Game = "tgame"
	r.Capacity, r.MinPlayers = 8, 2
	return r
}

func (f *fixture) cup(r settings.Rules, players []string) string {
	f.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	f.cups = append(f.cups, id)
	plugintest.Exec(f.t, f.db, `INSERT INTO cup (id, name, status, signup_at, starts_at, rules) VALUES ($1,'{"en":"Test cup"}','signup',$2,$3,$4)`,
		id, f.clock.Add(-time.Hour), f.clock.Add(time.Hour), jsonOf(r))
	for i, u := range players {
		f.clock = f.clock.Add(time.Millisecond)
		if _, err := f.s.rpcSignup(context.Background(), u, `{"cupId":"`+id+`"}`); err != nil {
			f.t.Fatalf("signup %d: %v", i, err)
		}
	}
	return id
}

func (f *fixture) start(id string) {
	f.t.Helper()
	var starts time.Time
	_ = f.db.QueryRow(`SELECT starts_at FROM cup WHERE id=$1`, id).Scan(&starts)
	if f.clock.Before(starts) {
		f.clock = starts
	}
	f.s.advanceCups(context.Background())
}

func (f *fixture) tick(d time.Duration) {
	f.clock = f.clock.Add(d)
	f.s.advanceCups(context.Background())
}

func (f *fixture) bracket(id string) *bracket {
	f.t.Helper()
	b, err := f.s.readBracket(context.Background(), id)
	if err != nil {
		f.t.Fatal(err)
	}
	return b
}

func (f *fixture) matchOf(id, user string) *matchRow {
	f.t.Helper()
	m := f.bracket(id).activeMatchOf(user)
	if m == nil {
		f.t.Fatalf("no active match for %s", user)
	}
	return m
}

func (f *fixture) ready(id, user string) map[string]any {
	f.t.Helper()
	out, err := f.s.ready(context.Background(), user, id)
	if err != nil {
		f.t.Fatalf("ready %s: %v", user, err)
	}
	return out
}

func (f *fixture) play(id, a, b, winner string) {
	f.t.Helper()
	f.ready(id, a)
	out := f.ready(id, b)
	if out["launched"] != true {
		f.t.Fatalf("not launched: %v", out)
	}
	f.report(f.matchOf(id, a), winner, false)
}

func (f *fixture) report(m *matchRow, winner string, draw bool) {
	f.t.Helper()
	if m.GameMatchID == "" {
		f.t.Fatalf("match %s has no game: %+v", m.ID, m)
	}
	if err := f.s.report(context.Background(), cuphooks.Result{GameMatchID: m.GameMatchID, Winner: winner, Draw: draw, Present: []string{m.A, m.B}}); err != nil {
		f.t.Fatalf("report: %v", err)
	}
}

func (f *fixture) player(id, user string) *playerRow {
	return f.bracket(id).players[user]
}

func (f *fixture) has(user, subject string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.notes[user] {
		if s == subject {
			return true
		}
	}
	return false
}

func (f *fixture) granted(user, prize string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, g := range f.grants {
		if g.UserID == user && g.PrizeKey == prize {
			return true
		}
	}
	return false
}

func TestCupWithByesBestOfThreeFinalAndPrizes(t *testing.T) {
	f := newFixture(t)
	p := f.users(6, 1000)
	for i, u := range p {
		plugintest.Exec(t, f.db, `INSERT INTO trophy_player (user_id, trophies) VALUES ($1,$2) ON CONFLICT (user_id) DO UPDATE SET trophies=EXCLUDED.trophies`, u, 600-i*100)
	}
	t.Cleanup(func() {
		for _, u := range p {
			_, _ = f.db.Exec(`DELETE FROM trophy_player WHERE user_id=$1`, u)
		}
	})
	r := f.rules()
	r.Seeding = settings.SeedTrophies
	r.Entry.Fee = settings.Fee{CurrencyKey: "coin", Amount: 100}
	r.Prizes = settings.Prizes{Champion: "cp_champ", RunnerUp: "cp_ru", Semi: "cp_semi", Reached: []settings.Reached{{Round: 2, PrizeKey: "cp_r2"}},
		Pool: settings.Pool{Percent: 50, Shares: settings.Shares{Champion: 50, RunnerUp: 25, Semi: 12}}}
	r.RoundBestOf = []settings.RoundBestOf{{Round: "final", BestOf: 3}}
	id := f.cup(r, p)
	if bal := plugintest.Balance(t, f.db, p[0])["coin"]; bal != 900 {
		t.Fatalf("entry fee not taken: %d", bal)
	}
	f.start(id)
	b := f.bracket(id)
	if b.cup.Status != "running" || b.cup.Size != 8 || b.cup.Rounds != 3 || b.cup.Pool.Amount != 300 {
		t.Fatalf("cup %+v", b.cup)
	}
	for i, u := range p {
		if b.players[u].Seed != i+1 {
			t.Fatalf("player %d seed %d", i, b.players[u].Seed)
		}
	}
	if m := b.match(kindMain, 1, 0); m.Status != stBye || m.Winner != p[0] {
		t.Fatalf("seed 1 bye: %+v", m)
	}
	if m := b.match(kindMain, 2, 0); m.A != p[0] || b.players[p[0]].Reached != 2 {
		t.Fatalf("bye winner not placed: %+v", m)
	}
	f.play(id, p[3], p[4], p[3])
	m := f.bracket(id).match(kindMain, 1, 1)
	if m.Status != stFinished || m.Winner != p[3] {
		t.Fatalf("4 v 5: %+v", m)
	}
	if err := f.s.report(context.Background(), cuphooks.Result{GameMatchID: m.GameMatchID, Winner: p[4]}); err != nil {
		t.Fatalf("duplicate report: %v", err)
	}
	if again := f.bracket(id).match(kindMain, 1, 1); again.Winner != p[3] || again.WinsA+again.WinsB != 1 {
		t.Fatalf("duplicate result counted: %+v", again)
	}
	f.ready(id, p[2])
	f.tick(121 * time.Minute)
	if m := f.bracket(id).match(kindMain, 1, 3); m.Winner != p[2] || m.Reason != "no_show" {
		t.Fatalf("no-show: %+v", m)
	}
	f.play(id, p[0], p[3], p[0])
	f.play(id, p[1], p[2], p[1])
	final := f.bracket(id).match(kindMain, 3, 0)
	if final.BestOf != 3 || final.Status != stOpen || final.A != p[0] || final.B != p[1] {
		t.Fatalf("final: %+v", final)
	}
	f.play(id, p[0], p[1], p[1])
	final = f.matchOf(id, p[0])
	if final.Status != stPlaying || final.Games != 2 || final.WinsB != 1 {
		t.Fatalf("next game not launched: %+v", final)
	}
	f.report(final, p[0], false)
	final = f.matchOf(id, p[0])
	f.report(final, "", true)
	final = f.matchOf(id, p[0])
	if final.Draws != 1 || final.Games != 4 || final.Status != stPlaying {
		t.Fatalf("draw not replayed: %+v", final)
	}
	f.report(final, p[0], false)
	b = f.bracket(id)
	if b.cup.Status != "finished" || b.cup.Champion != p[0] {
		t.Fatalf("cup not finished: %+v", b.cup)
	}
	places := map[int]int{}
	for _, u := range p {
		places[b.players[u].Place]++
	}
	if b.players[p[0]].Place != 1 || b.players[p[1]].Place != 2 || b.players[p[2]].Place != 3 || b.players[p[3]].Place != 3 || b.players[p[4]].Place != 5 {
		t.Fatalf("places %v", places)
	}
	if got := plugintest.Balance(t, f.db, p[0])["coin"]; got != 1050 {
		t.Fatalf("champion pool share: %d", got)
	}
	if got := plugintest.Balance(t, f.db, p[2])["coin"]; got != 936 {
		t.Fatalf("semi pool share: %d", got)
	}
	for _, c := range []struct{ u, k string }{{p[0], "cp_champ"}, {p[1], "cp_ru"}, {p[2], "cp_semi"}, {p[3], "cp_semi"}, {p[0], "cp_r2"}, {p[3], "cp_r2"}} {
		if !f.granted(c.u, c.k) {
			t.Fatalf("%s not granted", c.k)
		}
	}
	if f.granted(p[4], "cp_r2") {
		t.Fatal("round 2 prize for a first round loser")
	}
	f.s.payPending(context.Background(), id, 50)
	if got := plugintest.Balance(t, f.db, p[0])["coin"]; got != 1050 {
		t.Fatalf("prizes paid twice: %d", got)
	}
	for _, s := range []string{"cup_round_started", "cup_advanced", "cup_finished", "cup_prize"} {
		if !f.has(p[0], s) {
			t.Fatalf("champion missing %s", s)
		}
	}
	if !f.has(p[4], "cup_eliminated") || !f.has(p[4], "cup_opponent_ready") {
		t.Fatal("loser notifications")
	}
}

func TestLiveModeCheckInAndForfeit(t *testing.T) {
	f := newFixture(t)
	p := f.users(4, 0)
	r := f.rules()
	r.Timing.Mode = settings.ModeLive
	r.Timing.CheckInMinutes, r.Timing.RoundMinutes = 5, 20
	r.Timing.NoShow = settings.NoShowHigherSeed
	id := f.cup(r, p)
	f.start(id)
	b := f.bracket(id)
	m0, m1 := b.match(kindMain, 1, 0), b.match(kindMain, 1, 1)
	if m0.Status != stOpen || m0.StartsAt == nil || !m0.StartsAt.Equal(f.clock.Add(5*time.Minute)) {
		t.Fatalf("check-in window: %+v", m0)
	}
	f.ready(id, m0.A)
	if out := f.ready(id, m0.B); out["launched"] == true {
		t.Fatal("live match launched before its start time")
	}
	f.ready(id, m1.A)
	f.tick(5 * time.Minute)
	b = f.bracket(id)
	if got := b.match(kindMain, 1, 0); got.Status != stPlaying {
		t.Fatalf("checked-in match not launched: %+v", got)
	}
	if got := b.match(kindMain, 1, 1); got.Winner != m1.A || got.Reason != "no_show" {
		t.Fatalf("no check-in did not forfeit: %+v", got)
	}
	f.report(b.match(kindMain, 1, 0), m0.A, false)
	final := f.bracket(id).match(kindMain, 2, 0)
	if final.Status != stWaiting || !final.StartsAt.Equal(b.cup.StartsAt.Add(25*time.Minute)) {
		t.Fatalf("final schedule: %+v", final)
	}
	if _, err := f.s.ready(context.Background(), m0.A, id); err == nil {
		t.Fatal("check-in before the window opened")
	}
	f.tick(15 * time.Minute)
	if got := f.bracket(id).match(kindMain, 2, 0); got.Status != stOpen || !f.has(m0.A, "cup_check_in") {
		t.Fatalf("check-in not opened: %+v", got)
	}
	f.tick(5 * time.Minute)
	b = f.bracket(id)
	want := m0.A
	if b.players[m1.A].Seed < b.players[m0.A].Seed {
		want = m1.A
	}
	if got := b.match(kindMain, 2, 0); got.Winner != want || got.Reason != "double_no_show:higher_seed" || b.cup.Status != "finished" {
		t.Fatalf("double no-show rule: %+v %s", got, b.cup.Status)
	}
}

func TestThirdPlaceBothOutAndWalkovers(t *testing.T) {
	f := newFixture(t)
	p := f.users(4, 0)
	r := f.rules()
	r.ThirdPlace = true
	r.Prizes.Third = "cp_third"
	id := f.cup(r, p)
	f.start(id)
	b := f.bracket(id)
	if b.third() == nil {
		t.Fatal("no third place match")
	}
	m0 := b.match(kindMain, 1, 0)
	f.play(id, m0.A, m0.B, m0.B)
	f.tick(121 * time.Minute)
	b = f.bracket(id)
	if got := b.match(kindMain, 1, 1); got.Status != stVoid {
		t.Fatalf("both out: %+v", got)
	}
	if final := b.match(kindMain, 2, 0); final.Winner != m0.B || final.Reason != "walkover" {
		t.Fatalf("final walkover: %+v", final)
	}
	if th := b.third(); th.Winner != m0.A || th.Reason != "walkover" {
		t.Fatalf("third walkover: %+v", th)
	}
	if b.cup.Status != "finished" || b.players[m0.A].Place != 3 || b.players[b.match(kindMain, 1, 1).A].Place != 0 {
		t.Fatalf("places after both out: %+v", b.cup)
	}
	if !f.granted(m0.A, "cp_third") {
		t.Fatal("third prize")
	}
}

func TestRefundsAndUnderfill(t *testing.T) {
	f := newFixture(t)
	p := f.users(3, 500)
	r := f.rules()
	r.MinPlayers = 4
	r.Entry.Fee = settings.Fee{CurrencyKey: "coin", Amount: 100}
	r.Entry.WithdrawRefundPercent = 50
	id := f.cup(r, p)
	if _, err := f.s.rpcSignup(context.Background(), p[0], `{"cupId":"`+id+`"}`); err == nil {
		t.Fatal("signed up twice")
	}
	if _, err := f.s.rpcWithdraw(context.Background(), p[0], `{"cupId":"`+id+`"}`); err != nil {
		t.Fatal(err)
	}
	if got := plugintest.Balance(t, f.db, p[0])["coin"]; got != 450 {
		t.Fatalf("half refund: %d", got)
	}
	f.clock = f.clock.Add(time.Second)
	if _, err := f.s.rpcSignup(context.Background(), p[0], `{"cupId":"`+id+`"}`); err != nil {
		t.Fatalf("signup again: %v", err)
	}
	f.start(id)
	b := f.bracket(id)
	if b.cup.Status != "cancelled" || b.cup.CancelReason != "underfill" {
		t.Fatalf("underfill: %+v", b.cup)
	}
	for i, want := range []int64{450, 500, 500} {
		if got := plugintest.Balance(t, f.db, p[i])["coin"]; got != want {
			t.Fatalf("player %d balance %d want %d", i, got, want)
		}
	}
	f.s.retryRefunds(context.Background(), 100)
	if got := plugintest.Balance(t, f.db, p[1])["coin"]; got != 500 {
		t.Fatalf("refund paid twice: %d", got)
	}
	if !f.has(p[1], "cup_cancelled") {
		t.Fatal("cancel notification")
	}
	poor := f.users(1, 10)
	id2 := f.cup(r, nil)
	if _, err := f.s.rpcSignup(context.Background(), poor[0], `{"cupId":"`+id2+`"}`); err == nil {
		t.Fatal("signed up without the fee")
	}
}

func TestRequirementsAndCapacity(t *testing.T) {
	f := newFixture(t)
	p := f.users(3, 0)
	r := f.rules()
	r.Capacity = 2
	r.Entry.Requirements.MinTrophies = 10
	plugintest.Exec(t, f.db, `INSERT INTO trophy_player (user_id, trophies) VALUES ($1,20),($2,20) ON CONFLICT (user_id) DO UPDATE SET trophies=20`, p[0], p[1])
	t.Cleanup(func() { _, _ = f.db.Exec(`DELETE FROM trophy_player WHERE user_id = ANY($1)`, p) })
	id := f.cup(r, p[:2])
	if _, err := f.s.rpcSignup(context.Background(), p[2], `{"cupId":"`+id+`"}`); err == nil {
		t.Fatal("full cup accepted a player")
	}
	r.Capacity = 8
	id2 := f.cup(r, nil)
	if _, err := f.s.rpcSignup(context.Background(), p[2], `{"cupId":"`+id2+`"}`); err == nil {
		t.Fatal("trophy requirement not checked")
	}
}

func TestRecurringTemplateCreatesOneCup(t *testing.T) {
	f := newFixture(t)
	key := plugintest.Key("cpt")
	r := f.rules()
	plugintest.Exec(t, f.db, `INSERT INTO cup_template (id, key, name, active, schedule, signup_minutes, rules) VALUES ($1,$2,'{"en":"Friday"}',true,'0 18 * * 5',60,$3)`,
		uuid.Must(uuid.NewV4()).String(), key, jsonOf(r))
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if err := f.s.scheduleTemplates(ctx); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	var starts time.Time
	_ = f.db.QueryRow(`SELECT count(*), max(starts_at) FROM cup WHERE template_key=$1`, key).Scan(&n, &starts)
	if n != 1 || starts.UTC().Weekday() != time.Friday || starts.UTC().Hour() != 18 {
		t.Fatalf("recurring cups %d at %v", n, starts)
	}
	f.clock = starts.Add(time.Minute)
	f.s.work(ctx)
	_ = f.db.QueryRow(`SELECT count(*) FROM cup WHERE template_key=$1`, key).Scan(&n)
	if n != 2 {
		t.Fatalf("next week's cup not created: %d", n)
	}
	plugintest.Exec(t, f.db, `UPDATE cup_template SET active=false WHERE key=$1`, key)
}

func TestGameVoidsTimeoutsAndAdminActions(t *testing.T) {
	f := newFixture(t)
	p := f.users(8, 0)
	r := f.rules()
	id := f.cup(r, p)
	f.start(id)
	b := f.bracket(id)
	m := b.match(kindMain, 1, 0)
	f.ready(id, m.A)
	f.ready(id, m.B)
	m = f.matchOf(id, m.A)
	if err := f.s.report(context.Background(), cuphooks.Result{GameMatchID: m.GameMatchID, Void: true}); err != nil {
		t.Fatal(err)
	}
	if got := f.matchOf(id, m.A); got.Status != stOpen || !f.has(m.A, "cup_game_void") {
		t.Fatalf("void game: %+v", got)
	}
	f.ready(id, m.A)
	f.ready(id, m.B)
	m = f.matchOf(id, m.A)
	if err := f.s.report(context.Background(), cuphooks.Result{MatchID: m.ID, GameNo: m.Games, GameMatchID: "other", Void: true, Present: []string{m.B}}); err != nil {
		t.Fatal(err)
	}
	if got := f.bracket(id).match(kindMain, 1, 0); got.Winner != m.B || got.Reason != "forfeit" {
		t.Fatalf("forfeit by absence: %+v", got)
	}
	m1 := f.bracket(id).match(kindMain, 1, 1)
	f.ready(id, m1.A)
	f.ready(id, m1.B)
	f.tick(31 * time.Minute)
	if got := f.bracket(id).match(kindMain, 1, 1); got.Status != stOpen {
		t.Fatalf("game timeout: %+v", got)
	}
	m2 := f.bracket(id).match(kindMain, 1, 2)
	_, err := f.s.withCup(context.Background(), id, false, func(tx *sql.Tx, b *bracket) error {
		b.players[m2.A].Status = "disqualified"
		b.touchPlayer(m2.A)
		x := b.matchByID(m2.ID)
		b.finish(x, x.B, "disqualified")
		y := b.matchByID(m1.ID)
		b.finish(y, y.A, "admin")
		b.advance()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	b = f.bracket(id)
	if b.match(kindMain, 1, 2).Winner != m2.B || b.players[m2.A].Status != "disqualified" || b.match(kindMain, 2, 0).A != m.B || b.match(kindMain, 2, 0).B != m1.A || b.match(kindMain, 2, 1).A != m2.B {
		t.Fatal("disqualify and forced result")
	}
}
