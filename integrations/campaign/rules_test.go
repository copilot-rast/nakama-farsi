package main

import (
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
)

func f(v float64) *float64 { return &v }

func TestStarsFollowTheDefaultRules(t *testing.T) {
	cfg := settings.Defaults()
	stars := func(success bool, elapsed float64, health *float64, limit *float64) int {
		vals := resultValues(runResult{Success: success, Elapsed: elapsed, Health: health}, limit)
		return computeStars(success, vals, cfg.Stars.Rules, 3, 1, true)
	}
	cases := []struct {
		name    string
		success bool
		elapsed float64
		health  *float64
		limit   *float64
		want    int
	}{
		{"failed", false, 10, f(100), f(60), 0},
		{"fast and healthy", true, 42, f(75), f(60), 3},
		{"just over 30% saved", true, 42.1, f(90), f(60), 2},
		{"healthy but slow", true, 59, f(100), f(60), 1},
		{"fast but hurt", true, 10, f(60), f(60), 2},
		{"no health (missing passes)", true, 40, nil, f(60), 3},
		{"no time limit", true, 500, f(80), nil, 3},
		{"late finish clamps to the limit", true, 62, f(100), f(60), 1},
	}
	for _, c := range cases {
		if got := stars(c.success, c.elapsed, c.health, c.limit); got != c.want {
			t.Errorf("%s: stars = %d, want %d", c.name, got, c.want)
		}
	}
	vals := resultValues(runResult{Success: true, Elapsed: 40}, f(60))
	if computeStars(true, vals, cfg.Stars.Rules, 3, 1, false) != 1 {
		t.Error("a missing health field must fail when the setting says fail")
	}
}

func TestCustomFieldsDriveStarsAndCannotFakeServerFields(t *testing.T) {
	rules := []settings.StarRule{
		{Stars: 3, Conditions: []settings.Cond{{Field: "arrests", Op: ">=", Value: 5}, {Field: "crashes", Op: "==", Value: 0}}},
		{Stars: 2, Conditions: []settings.Cond{{Field: "arrests", Op: ">=", Value: 3}}},
	}
	vals := resultValues(runResult{Success: true, Elapsed: 10, Fields: map[string]float64{"arrests": 6, "crashes": 0, "timeSavedPercent": 99}}, f(20))
	if got := computeStars(true, vals, rules, 3, 1, false); got != 3 {
		t.Fatalf("stars = %d", got)
	}
	if vals["timeSavedPercent"] != 50 {
		t.Fatalf("a field overrode the server's timeSavedPercent: %v", vals["timeSavedPercent"])
	}
	vals = resultValues(runResult{Success: true, Elapsed: 10, Fields: map[string]float64{"arrests": 4, "crashes": 2}}, nil)
	if got := computeStars(true, vals, rules, 3, 1, false); got != 2 {
		t.Fatalf("stars = %d", got)
	}
	if got := computeStars(true, map[string]float64{}, rules, 3, 0, false); got != 0 {
		t.Fatalf("min on success 0 = %d", got)
	}
}

func TestDifferencePayout(t *testing.T) {
	m := []float64{0.2, 0.5, 1}
	full := int64(30000)
	claimed := int64(0)
	steps := []struct {
		stars int
		want  int64
	}{{1, 6000}, {1, 0}, {3, 24000}, {2, 0}, {3, 0}}
	for i, s := range steps {
		got := payout(full, multiplier(s.stars, m), claimed)
		if got != s.want {
			t.Fatalf("step %d (%d stars): pay %d, want %d", i, s.stars, got, s.want)
		}
		claimed += got
	}
	if claimed != full {
		t.Fatalf("total %d", claimed)
	}
	if scaled(15, 0.5) != 7 || multiplier(0, m) != 0 || multiplier(9, m) != 1 {
		t.Fatal("scaling")
	}
}

func TestCooldownAndDailyCap(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Tehran")
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	last := now.Add(-30 * time.Second)
	st := groupState{LastAt: &last, Day: dayKey(now, loc), Count: 2}
	if limitCheck(st, 60, 0, now, loc) != "cooldown" {
		t.Fatal("cooldown not applied")
	}
	if at := nextRewardAt(st, 60, now); at == nil || !at.Equal(last.Add(time.Minute)) {
		t.Fatalf("next = %v", at)
	}
	if limitCheck(st, 10, 2, now, loc) != "daily_limit" {
		t.Fatal("daily cap not applied")
	}
	if left := rewardsLeft(st, 3, now, loc); left == nil || *left != 1 {
		t.Fatalf("left = %v", left)
	}
	later := now.Add(time.Hour)
	if limitCheck(st, 10, 2, later, loc) != "" {
		t.Fatal("the cap did not restart on a new day in the zone")
	}
	if rewardsLeft(st, 0, now, loc) != nil || nextRewardAt(groupState{}, 60, now) != nil {
		t.Fatal("no cap / no state")
	}
}

func TestSettingsValidate(t *testing.T) {
	if f, c := settings.Defaults().Validate(); c != "" {
		t.Fatalf("defaults invalid: %s %s", f, c)
	}
	bad := settings.Defaults()
	bad.Stars.Multipliers = []float64{1, 1}
	if f, _ := bad.Validate(); f != "stars.multipliers" {
		t.Fatalf("field = %s", f)
	}
	bad = settings.Defaults()
	bad.Rewards.DayZone = "Mars/Base"
	if _, c := bad.Validate(); c != "INVALID_TIMEZONE" {
		t.Fatalf("code = %s", c)
	}
	bad = settings.Defaults()
	bad.Stars.Rules[0].Conditions[0].Op = "~"
	if f, c := bad.Validate(); c != "INVALID_OPERATOR" || f != "stars.rules[0].conditions[0].op" {
		t.Fatalf("%s %s", f, c)
	}
	if chapterPercent(5, 3, 3) != 500.0/9 || chapterPercent(20, 3, 3) != 100 || chapterPercent(1, 0, 3) != 0 {
		t.Fatal("chapter percent")
	}
}
