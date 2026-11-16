package main

import (
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
)

func tiered(basis string, th ...float64) *board {
	b := &board{SortOrder: "desc", TierBasis: basis}
	for i, t := range th {
		b.Tiers = append(b.Tiers, Tier{Key: []string{"gold", "silver", "bronze", "iron"}[i], Threshold: t})
	}
	return b
}

func TestTierIndexByBasis(t *testing.T) {
	score := tiered("score", 1000, 500, 0)
	for _, c := range []struct {
		score int64
		want  int
	}{{1500, 0}, {1000, 0}, {999, 1}, {500, 1}, {10, 2}, {0, 2}} {
		if got := tierIndex(score, 9, c.score, 100); got != c.want {
			t.Fatalf("score %d: tier %d, want %d", c.score, got, c.want)
		}
	}
	asc := tiered("score", 30, 60)
	asc.SortOrder = "asc"
	if tierIndex(asc, 1, 25, 0) != 0 || tierIndex(asc, 1, 45, 0) != 1 || tierIndex(asc, 1, 90, 0) != -1 {
		t.Fatal("ascending score tiers")
	}
	rank := tiered("rank", 1, 3, 10)
	if tierIndex(rank, 1, 0, 0) != 0 || tierIndex(rank, 2, 0, 0) != 1 || tierIndex(rank, 10, 0, 0) != 2 || tierIndex(rank, 11, 0, 0) != -1 {
		t.Fatal("rank tiers")
	}
	pct := tiered("percent", 1, 10, 50)

	if tierIndex(pct, 2, 0, 200) != 0 || tierIndex(pct, 3, 0, 200) != 1 || tierIndex(pct, 100, 0, 200) != 2 || tierIndex(pct, 101, 0, 200) != -1 {
		t.Fatal("percent tiers")
	}

	if tierIndex(pct, 1, 0, 3) != 0 || percentRank(1, 3) != 1 {
		t.Fatal("percent rounds up to one player")
	}
	if tierIndex(&board{TierBasis: "none"}, 1, 10, 1) != -1 {
		t.Fatal("no tiers")
	}
}

func TestNextTier(t *testing.T) {
	b := tiered("score", 1000, 500, 0)
	if n := nextTier(b, 1, 5, 700, 10); n["key"] != "gold" || n["need"] != 300.0 {
		t.Fatalf("silver -> gold: %v", n)
	}
	if nextTier(b, 0, 1, 1200, 10) != nil {
		t.Fatal("best tier has no next")
	}
	r := tiered("rank", 3, 10)
	if n := nextTier(r, -1, 25, 0, 50); n["key"] != "silver" || n["need"] != 15 {
		t.Fatalf("outside -> silver: %v", n)
	}
}

func TestPlanForRowsAndStack(t *testing.T) {
	b := tiered("score", 1000, 0)
	b.Rewards = []RewardRow{
		{FromRank: 1, ToRank: 1, Currencies: map[string]int64{"gem": 50}, PrizeKey: "crown"},
		{FromRank: 1, ToRank: 3, Currencies: map[string]int64{"coin": 100}},
		{Tier: "gold", FromRank: 1, ToRank: 0, Currencies: map[string]int64{"coin": 10}},
		{Tier: "silver", FromRank: 1, ToRank: 2, PrizeGroupKey: "box"},
	}
	first := planFor(b, placed{Rank: 1, Score: 2000, Tier: "gold", TierRank: 1}, settings.StackFirst, "p:u1")
	if first == nil || first.Currencies["gem"] != 50 || first.Currencies["coin"] != 0 || len(first.Prizes) != 1 || first.Prizes[0].SourceID != "p:u1:r0" {
		t.Fatalf("first: %+v", first)
	}
	all := planFor(b, placed{Rank: 1, Score: 2000, Tier: "gold", TierRank: 1}, settings.StackAll, "p:u1")
	if all.Currencies["gem"] != 50 || all.Currencies["coin"] != 110 || len(all.Rows) != 3 {
		t.Fatalf("all: %+v", all)
	}
	silver := planFor(b, placed{Rank: 7, Score: 300, Tier: "silver", TierRank: 2}, settings.StackFirst, "p:u7")
	if silver == nil || len(silver.Prizes) != 1 || silver.Prizes[0].PrizeGroupKey != "box" || silver.Currencies != nil {
		t.Fatalf("silver: %+v", silver)
	}
	if planFor(b, placed{Rank: 9, Score: 100, Tier: "silver", TierRank: 3}, settings.StackAll, "x") != nil {
		t.Fatal("silver rank 3 gets nothing")
	}
	b.RewardMinScore = 2500
	if planFor(b, placed{Rank: 1, Score: 2000, Tier: "gold", TierRank: 1}, settings.StackAll, "x") != nil {
		t.Fatal("below the reward minimum")
	}
}

func TestAssignTiersCountsRankInsideTier(t *testing.T) {
	b := tiered("score", 100, 0)
	list := []placed{{Rank: 1, Score: 300}, {Rank: 2, Score: 150}, {Rank: 3, Score: 90}, {Rank: 4, Score: 5}}
	assignTiers(b, list)
	if list[1].Tier != "gold" || list[1].TierRank != 2 || list[2].Tier != "silver" || list[2].TierRank != 1 || list[3].TierRank != 2 {
		t.Fatalf("%+v", list)
	}
}

func TestScheduleAndPeriodBounds(t *testing.T) {
	if c, ok := schedule("weekly", ""); !ok || c != "0 0 * * 1" {
		t.Fatal("weekly preset")
	}
	if _, ok := schedule("custom", "not a cron"); ok {
		t.Fatal("bad cron accepted")
	}
	if c, ok := schedule("custom", "*/15 * * * *"); !ok || c != "*/15 * * * *" {
		t.Fatal("custom cron")
	}
	if _, ok := schedule("yearly", ""); ok {
		t.Fatal("unknown period")
	}
	now := time.Date(2026, 11, 18, 15, 4, 0, 0, time.UTC)
	start, end := periodBounds("0 0 * * 1", now)
	if start == nil || end == nil || !start.Equal(time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)) || !end.Equal(time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("weekly bounds %v %v", start, end)
	}
	if s, e := periodBounds("", now); s != nil || e != nil {
		t.Fatal("no reset = no bounds")
	}
	if !expiry("", now).Equal(epoch) {
		t.Fatal("expiry without reset is the epoch")
	}
}

func TestScoreDeltaAndClamp(t *testing.T) {
	if scoreDelta(Source{Calc: "count"}, 77) != 1 || scoreDelta(Source{Calc: "value", Multiplier: 100}, 1.235) != 124 || scoreDelta(Source{Calc: "value"}, 3) != 3 {
		t.Fatal("delta")
	}
	if v, o := clampWrite(Limits{MaxPerWrite: 50, OnExceed: "clamp"}, 80); v != 50 || o != "clamped" {
		t.Fatal("clamp")
	}
	if _, o := clampWrite(Limits{MaxPerWrite: 50, OnExceed: "reject"}, 80); o != "too_high" {
		t.Fatal("reject")
	}
	if v, o := clampWrite(Limits{}, 80); v != 80 || o != "" {
		t.Fatal("no cap")
	}
}

func TestBoardStatus(t *testing.T) {
	now := time.Now()
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	if (&board{Active: true}).status(now) != "active" || (&board{}).status(now) != "inactive" ||
		(&board{Active: true, StartsAt: &future}).status(now) != "upcoming" || (&board{Active: true, EndsAt: &past}).status(now) != "ended" ||
		(&board{Active: true, FinishedAt: &past}).status(now) != "ended" {
		t.Fatal("status")
	}
	cfg := settings.Defaults()
	if !visible(&board{Active: true, EndsAt: &past}, cfg, now) || visible(&board{Active: true, EndsAt: &past}, settings.All{}, now) {
		t.Fatal("ended boards stay visible for showEndedDays")
	}
}

func TestSettingsValidate(t *testing.T) {
	d := settings.Defaults()
	if f, c := d.Validate(); c != "" {
		t.Fatalf("defaults invalid: %s %s", f, c)
	}
	d.Rewards.PayMode = "later"
	if f, _ := d.Validate(); f != "rewards.payMode" {
		t.Fatal("pay mode")
	}
	d = settings.Defaults()
	d.Display.MaxLimit = 10
	if f, _ := d.Validate(); f != "display.maxLimit" {
		t.Fatal("max limit below top limit")
	}
}
