package main

import (
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/trophy/internal/settings"
)

func testLadder() ladder {
	five := 300
	return newLadder([]league{
		{Key: "bronze", MinTrophies: 0, Active: true},
		{Key: "gold", MinTrophies: 600, Floor: true, Active: true},
		{Key: "silver", MinTrophies: 200, Floor: true, ResetTo: &five, Active: true},
		{Key: "hidden", MinTrophies: 100, Active: false},
	})
}

func TestLadderOrdersActiveLeaguesByMinimum(t *testing.T) {
	l := testLadder()
	if len(l) != 3 || l[0].Key != "bronze" || l[1].Key != "silver" || l[2].Key != "gold" {
		t.Fatalf("ladder %+v", l)
	}
	for trophies, want := range map[int]string{0: "bronze", 199: "bronze", 200: "silver", 599: "silver", 600: "gold", 9000: "gold"} {
		if got := l.at(trophies); got == nil || got.Key != want {
			t.Fatalf("%d trophies: %v, want %s", trophies, got, want)
		}
	}
	if n := l.next(250); n == nil || n.Key != "gold" {
		t.Fatalf("next of 250: %v", n)
	}
	if n := l.next(700); n != nil {
		t.Fatalf("no league after the top one: %v", n)
	}
	if newLadder(nil).at(10) != nil {
		t.Fatal("no leagues, no league")
	}
}

func TestMatchDeltaFixedAndScaled(t *testing.T) {
	a := settings.Defaults().Amounts
	if d := matchDelta(a, outcomeWin, 100, nil); d != 30 {
		t.Fatalf("fixed win %d", d)
	}
	if d := matchDelta(a, outcomeLoss, 100, nil); d != -20 {
		t.Fatalf("fixed loss %d", d)
	}
	if d := matchDelta(a, outcomeDraw, 100, nil); d != 0 {
		t.Fatalf("draw %d", d)
	}
	a.Scale = true
	strong, weak := 600, 100
	if d := matchDelta(a, outcomeWin, 100, &strong); d != 40 {
		t.Fatalf("beating a stronger player gives more: %d", d)
	}
	if d := matchDelta(a, outcomeWin, 600, &weak); d != 20 {
		t.Fatalf("beating a weaker player gives less: %d", d)
	}
	if d := matchDelta(a, outcomeLoss, 100, &strong); d != -10 {
		t.Fatalf("losing to a stronger player costs less: %d", d)
	}
	huge := 100000
	if d := matchDelta(a, outcomeWin, 0, &huge); d != a.MaxWin {
		t.Fatalf("win is capped: %d", d)
	}
	if d := matchDelta(a, outcomeLoss, 0, &huge); d != -a.MinLoss {
		t.Fatalf("loss has a floor: %d", d)
	}
}

func TestApplyDeltaFloorsAtLeagueMinimum(t *testing.T) {
	l := testLadder()
	if got := applyDelta(l, true, 210, -30); got != 200 {
		t.Fatalf("silver floors at 200, got %d", got)
	}
	if got := applyDelta(l, false, 210, -30); got != 180 {
		t.Fatalf("floors off, got %d", got)
	}
	if got := applyDelta(l, true, 150, -200); got != 0 {
		t.Fatalf("bronze has no floor but never goes negative, got %d", got)
	}
	if got := applyDelta(l, true, 590, 30); got != 620 {
		t.Fatalf("a win crosses into gold, got %d", got)
	}
}

func TestSeasonResetModes(t *testing.T) {
	l := testLadder()
	cut := settings.Season{Mode: settings.ResetCut, Above: 1000, KeepPercent: 50}
	if got := seasonReset(cut, l, 1600); got != 1300 {
		t.Fatalf("cut keeps half above 1000: %d", got)
	}
	if got := seasonReset(cut, l, 900); got != 900 {
		t.Fatalf("below the line nothing changes: %d", got)
	}
	if got := seasonReset(settings.Season{Mode: settings.ResetLeague}, l, 450); got != 300 {
		t.Fatalf("silver resets to its reset value: %d", got)
	}
	if got := seasonReset(settings.Season{Mode: settings.ResetLeague}, l, 800); got != 800 {
		t.Fatalf("gold has no reset value: %d", got)
	}
	if got := seasonReset(settings.Season{Mode: settings.ResetZero}, l, 800); got != 0 {
		t.Fatalf("zero: %d", got)
	}
}

func TestReadMatchFiltersModesAndReadsFields(t *testing.T) {
	m := settings.Defaults().Match
	m.Modes = []string{"online"}
	m.SkipField = "vsBot"
	if _, ok := readMatch(m, map[string]any{"mode": "friend", "won": true}); ok {
		t.Fatal("friend matches do not count")
	}
	if _, ok := readMatch(m, map[string]any{"mode": "online", "won": true, "vsBot": true}); ok {
		t.Fatal("bot matches are skipped")
	}
	f, ok := readMatch(m, map[string]any{"mode": "online", "won": true, "matchId": "m1", "opponentId": "u2"})
	if !ok || f.Outcome != outcomeWin || f.MatchID != "m1" || f.Opponent != "u2" {
		t.Fatalf("facts %+v %v", f, ok)
	}
	f, _ = readMatch(m, map[string]any{"mode": "online", "won": false, "draw": true})
	if f.Outcome != outcomeDraw {
		t.Fatalf("draw %+v", f)
	}
	f, _ = readMatch(m, map[string]any{"mode": "online"})
	if f.Outcome != outcomeLoss {
		t.Fatalf("no win flag is a loss %+v", f)
	}
}

func TestSettingsValidate(t *testing.T) {
	d := settings.Defaults()
	if f, c := d.Validate(); c != "" {
		t.Fatalf("defaults invalid: %s %s", f, c)
	}
	d.Season.Enabled, d.Season.Schedule = true, "bad cron"
	if f, _ := d.Validate(); f != "season.schedule" {
		t.Fatalf("bad cron: %s", f)
	}
	d = settings.Defaults()
	d.Amounts.MinWin, d.Amounts.MaxWin = 60, 50
	if f, _ := d.Validate(); f != "amounts.minWin" {
		t.Fatalf("min above max: %s", f)
	}
	if k := claimKey(settings.Prizes{OncePer: settings.OnceSeason}, "gold", 3); k != "gold:s3" {
		t.Fatal(k)
	}
	if k := claimKey(settings.Prizes{OncePer: settings.OnceEver}, "gold", 3); k != "gold" {
		t.Fatal(k)
	}
}
