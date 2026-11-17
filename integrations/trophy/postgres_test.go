package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/trophy/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type fixture struct {
	t      *testing.T
	s      *service
	event  string
	mu     sync.Mutex
	grants []economyhooks.PrizeGrant
	notes  []string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, event: plugintest.Key("tr_played")}
	s := newService(db, nil, plugintest.Logger{})
	s.prize = func(_ context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.grants = append(f.grants, g)
		return economyhooks.PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey}}}, nil
	}
	s.notify = func(_ context.Context, userID, subject string, _ map[string]any, _ int, _ bool) {
		f.mu.Lock()
		f.notes = append(f.notes, subject)
		f.mu.Unlock()
	}
	f.s = s
	ctx := context.Background()
	plugintest.Exec(t, db, `DELETE FROM trophy_league`)
	plugintest.Exec(t, db, `DELETE FROM trophy_season`)
	for _, l := range []struct {
		key   string
		min   int
		floor bool
		prize string
	}{{"t_bronze", 0, false, ""}, {"t_silver", 100, true, "t_prize_silver"}, {"t_gold", 300, true, "t_prize_gold"}} {
		plugintest.Exec(t, db, `INSERT INTO trophy_league (id, key, name, min_trophies, floor, prize_key) VALUES ($1,$2,'{"en":"x"}',$3,$4,$5)`, uuid.Must(uuid.NewV4()).String(), l.key, l.min, l.floor, l.prize)
	}
	cfg := settings.Defaults()
	cfg.Match.Event = f.event
	cfg.Match.Modes = []string{"online"}
	if err := settings.Save(ctx, db, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = settings.Save(context.Background(), db, settings.Defaults(), "")
		plugintest.Exec(t, db, `DELETE FROM trophy_league`)
		plugintest.Exec(t, db, `DELETE FROM trophy_season`)
	})
	s.invalidate()
	return f
}

func (f *fixture) user() string {
	id := plugintest.NewUser(f.t, f.s.db)
	f.t.Cleanup(func() {
		for _, q := range []string{`DELETE FROM trophy_log WHERE user_id=$1`, `DELETE FROM trophy_prize WHERE user_id=$1`, `DELETE FROM trophy_player WHERE user_id=$1`} {
			_, _ = f.s.db.Exec(q, id)
		}
	})
	return id
}

func (f *fixture) play(user, opponent, match string, won bool, mode string) {
	f.s.onEconomyEvent(context.Background(), playerevents.Recorded{ID: uuid.Must(uuid.NewV4()).String(), UserID: user, Key: f.event, Value: 1,
		Metadata: map[string]any{"matchId": match, "opponentId": opponent, "won": won, "mode": mode}})
}

func (f *fixture) trophies(user string) (int, string) {
	var t int
	var lk string
	_ = f.s.db.QueryRow(`SELECT trophies, league_key FROM trophy_player WHERE user_id=$1`, user).Scan(&t, &lk)
	return t, lk
}

func TestMatchResultsMoveTrophiesOncePerMatch(t *testing.T) {
	f := newFixture(t)
	a, b := f.user(), f.user()
	m := uuid.Must(uuid.NewV4()).String()
	f.play(a, b, m, true, "online")
	f.play(b, a, m, false, "online")
	f.play(a, b, m, true, "online")
	if got, lk := f.trophies(a); got != 30 || lk != "t_bronze" {
		t.Fatalf("winner %d %s", got, lk)
	}
	if got, _ := f.trophies(b); got != 0 {
		t.Fatalf("loser never goes below zero: %d", got)
	}
	var n int
	_ = f.s.db.QueryRow(`SELECT count(*) FROM trophy_log WHERE user_id=$1`, a).Scan(&n)
	if n != 1 {
		t.Fatalf("a resent match result must not count twice: %d log rows", n)
	}
	f.play(a, b, uuid.Must(uuid.NewV4()).String(), true, "friend")
	if got, _ := f.trophies(a); got != 30 {
		t.Fatalf("friend matches are not counted: %d", got)
	}
}

func TestLeaguePrizeIsPaidOnceAndFloorsHold(t *testing.T) {
	f := newFixture(t)
	a, b := f.user(), f.user()
	for i := 0; i < 4; i++ {
		f.play(a, b, uuid.Must(uuid.NewV4()).String(), true, "online")
	}
	if got, lk := f.trophies(a); got != 120 || lk != "t_silver" {
		t.Fatalf("after 4 wins %d %s", got, lk)
	}
	f.play(a, b, uuid.Must(uuid.NewV4()).String(), false, "online")
	if got, lk := f.trophies(a); got != 100 || lk != "t_silver" {
		t.Fatalf("silver floor holds at 100: %d %s", got, lk)
	}
	f.play(a, b, uuid.Must(uuid.NewV4()).String(), false, "online")
	if got, _ := f.trophies(a); got != 100 {
		t.Fatalf("floor again: %d", got)
	}
	ctx := context.Background()
	if _, err := f.s.adminWrite(ctx, a, 90, "set", "admin"); err != nil {
		t.Fatal(err)
	}
	if got, lk := f.trophies(a); got != 90 || lk != "t_bronze" {
		t.Fatalf("admin can move a player below the floor: %d %s", got, lk)
	}
	f.play(a, b, uuid.Must(uuid.NewV4()).String(), true, "online")
	if got, lk := f.trophies(a); got != 120 || lk != "t_silver" {
		t.Fatalf("back to silver %d %s", got, lk)
	}
	silver := 0
	for _, g := range f.grants {
		if g.PrizeKey == "t_prize_silver" && g.UserID == a {
			silver++
			if g.SourceType != "trophy" {
				t.Fatalf("source type %s", g.SourceType)
			}
		}
	}
	if silver != 1 {
		t.Fatalf("silver prize paid %d times, want once", silver)
	}
	var status string
	_ = f.s.db.QueryRow(`SELECT status FROM trophy_prize WHERE user_id=$1 AND league_key='t_silver'`, a).Scan(&status)
	if status != "paid" {
		t.Fatalf("prize row %s", status)
	}
}

func TestScaledAmountsReadTheOpponentsTrophiesBeforeTheMatch(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	cfg := f.s.settings(ctx)
	cfg.Amounts.Scale = true
	if err := settings.Save(ctx, f.s.db, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	a, b := f.user(), f.user()
	if _, err := f.s.adminWrite(ctx, b, 500, "set", "admin"); err != nil {
		t.Fatal(err)
	}
	m := uuid.Must(uuid.NewV4()).String()
	f.play(b, a, m, false, "online")
	f.play(a, b, m, true, "online")
	if got, _ := f.trophies(a); got != 40 {
		t.Fatalf("a beat a 500-trophy player from 0: want 40, got %d", got)
	}
	if got, _ := f.trophies(b); got != 470 {
		t.Fatalf("b lost to a 0-trophy player and pays more: want 470, got %d", got)
	}
}

func TestSeasonResetCutsTrophiesAboveTheLine(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	cfg := f.s.settings(ctx)
	cfg.Season.Enabled, cfg.Season.Above, cfg.Season.KeepPercent = true, 200, 50
	if err := settings.Save(ctx, f.s.db, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	a, b := f.user(), f.user()
	_, _ = f.s.adminWrite(ctx, a, 400, "set", "admin")
	_, _ = f.s.adminWrite(ctx, b, 150, "set", "admin")
	if err := f.s.seasonTick(ctx, ""); err != nil {
		t.Fatal(err)
	}
	cur, _ := f.s.currentSeason(ctx, f.s.db)
	if cur == nil || cur.Number != 1 || cur.EndsAt == nil {
		t.Fatalf("season 1 opens with an end time: %+v", cur)
	}
	later := cur.EndsAt.Add(time.Minute)
	f.s.now = func() time.Time { return later }
	if err := f.s.seasonTick(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if got, lk := f.trophies(a); got != 300 || lk != "t_gold" {
		t.Fatalf("400 → 200 + 100 = 300: %d %s", got, lk)
	}
	if got, _ := f.trophies(b); got != 150 {
		t.Fatalf("below the line: %d", got)
	}
	cur, _ = f.s.currentSeason(ctx, f.s.db)
	if cur == nil || cur.Number != 2 {
		t.Fatalf("season 2 opened: %+v", cur)
	}
	if err := f.s.seasonTick(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if got, _ := f.trophies(a); got != 300 {
		t.Fatalf("one reset per season: %d", got)
	}
}
