package main

import (
	"testing"
	"time"
)

func mustLoc(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestEventClockDayBoundaryInTehranAndUTC(t *testing.T) {
	tehran := eventClock{loc: mustLoc(t, "Asia/Tehran"), weekStart: time.Saturday}
	utc := eventClock{loc: time.UTC, weekStart: time.Monday}
	before := time.Date(2026, 11, 11, 20, 29, 59, 0, time.UTC)
	after := time.Date(2026, 11, 11, 20, 30, 0, 0, time.UTC)
	if got := dateKey(tehran.periodStart("day", before)); got != "2026-11-11" {
		t.Fatalf("23:59:59 in Tehran is still the 11th, got %s", got)
	}
	if got := dateKey(tehran.periodStart("day", after)); got != "2026-11-12" {
		t.Fatalf("00:00 in Tehran is the 12th, got %s", got)
	}
	if dateKey(utc.periodStart("day", before)) != "2026-11-11" || dateKey(utc.periodStart("day", after)) != "2026-11-11" {
		t.Fatal("both instants are the 11th in UTC")
	}
	end := tehran.periodEnd("day", before)
	if end == nil || !end.Equal(after) {
		t.Fatalf("the Tehran day ends at 20:30 UTC, got %v", end)
	}
	if utcEnd := utc.periodEnd("day", before); !utcEnd.Equal(time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("the UTC day ends at midnight, got %v", utcEnd)
	}
	if tehran.periodEnd("total", before) != nil || dateKey(tehran.periodStart("total", before)) != "1970-01-01" {
		t.Fatal("the total window never ends")
	}
	if tehran.dayIndex(after)-tehran.dayIndex(before) != 1 || utc.dayIndex(after) != utc.dayIndex(before) {
		t.Fatal("day index must follow the local date")
	}
}

func TestEventClockWeekStartsOnTheConfiguredDay(t *testing.T) {
	tehran := eventClock{loc: mustLoc(t, "Asia/Tehran"), weekStart: time.Saturday}
	friday := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	if friday.Weekday() != time.Friday {
		t.Fatal("fixture must be a Friday")
	}
	if got := dateKey(tehran.periodStart("week", friday)); got != "2026-11-07" {
		t.Fatalf("a Saturday week holding Friday the 13th starts on the 7th, got %s", got)
	}
	saturdayLocal := time.Date(2026, 11, 13, 20, 30, 0, 0, time.UTC)
	if got := dateKey(tehran.periodStart("week", saturdayLocal)); got != "2026-11-14" {
		t.Fatalf("Saturday 00:00 Tehran opens a new week, got %s", got)
	}
	end := tehran.periodEnd("week", friday)
	if !end.Equal(saturdayLocal) {
		t.Fatalf("the week ends at Saturday 00:00 Tehran, got %v", end)
	}
	monday := eventClock{loc: time.UTC, weekStart: time.Monday}
	if got := dateKey(monday.periodStart("week", friday)); got != "2026-11-09" {
		t.Fatalf("a Monday week starts on the 9th, got %s", got)
	}
}

func TestStreakIndexUsesLocalDaysAndPeriodLength(t *testing.T) {
	clk := eventClock{loc: mustLoc(t, "Asia/Tehran")}
	a := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	b := time.Date(2026, 11, 11, 21, 0, 0, 0, time.UTC)
	if clk.streakIndex(b, 1)-clk.streakIndex(a, 1) != 1 {
		t.Fatal("crossing Tehran midnight is a new streak day")
	}
	start := clk.streakPeriodStart(clk.streakIndex(b, 1), 1)
	if !start.Equal(time.Date(2026, 11, 11, 20, 30, 0, 0, time.UTC)) {
		t.Fatalf("period start is Tehran midnight, got %v", start)
	}
	if clk.streakIndex(a, 3) != floorDiv(clk.dayIndex(a), 3) {
		t.Fatal("three-day periods divide the day index")
	}
	if floorDiv(-1, 3) != -1 || floorDiv(-3, 3) != -1 || floorDiv(5, 3) != 1 {
		t.Fatal("floorDiv rounds down")
	}
}

func TestRuleDue(t *testing.T) {
	once := ruleDef{Repeat: "once", Threshold: 3}
	if ruleDue(once, 2, 0, false, 0) != 0 || ruleDue(once, 3, 0, false, 0) != 1 || ruleDue(once, 9, 1, true, 1) != 0 || ruleDue(once, 5, 0, true, 1) != 0 {
		t.Fatal("once grants a single claim ever")
	}
	period := ruleDef{Repeat: "period", Threshold: 3}
	if ruleDue(period, 3, 0, true, 4) != 1 || ruleDue(period, 7, 1, true, 5) != 0 {
		t.Fatal("period grants one claim per window")
	}
	every := ruleDef{Repeat: "every", Threshold: 5}
	if ruleDue(every, 4, 0, false, 0) != 0 || ruleDue(every, 5, 0, false, 0) != 1 || ruleDue(every, 17, 1, true, 1) != 2 || ruleDue(every, 15, 3, true, 3) != 0 {
		t.Fatal("every grants one claim per multiple")
	}
	capped := ruleDef{Repeat: "every", Threshold: 1, MaxClaims: 3}
	if ruleDue(capped, 10, 0, false, 0) != 3 || ruleDue(capped, 10, 2, true, 2) != 1 || ruleDue(capped, 10, 3, true, 3) != 0 {
		t.Fatal("max claims caps every rule")
	}
	decimal := ruleDef{Repeat: "every", Threshold: 0.1}
	if ruleDue(decimal, 0.3, 0, false, 0) != 3 {
		t.Fatal("decimal thresholds must not lose a step to rounding")
	}
	if ruleMeasure("count", 4, 10, 7) != 4 || ruleMeasure("sum", 4, 10, 7) != 10 || ruleMeasure("max", 4, 10, 7) != 7 {
		t.Fatal("measure picks the aggregate")
	}
}

func last(v int64) *int64 { return &v }

func TestAdvanceStreakGraceResetAndStepBack(t *testing.T) {
	rule := streakRule{Grace: 1, MissAction: "reset"}
	st, adv, reset := advanceStreak(streakState{}, 100, rule)
	if !adv || reset || st.Count != 1 || st.Run != 1 || *st.Last != 100 {
		t.Fatalf("first day: %+v", st)
	}
	same, adv, _ := advanceStreak(st, 100, rule)
	if adv || same.Count != 1 {
		t.Fatal("a second event on the same day changes nothing")
	}
	st, _, _ = advanceStreak(st, 101, rule)
	if st.Count != 2 || st.GraceUsed != 0 {
		t.Fatalf("next day: %+v", st)
	}
	st, adv, reset = advanceStreak(st, 103, rule)
	if !adv || reset || st.Count != 3 || st.GraceUsed != 1 {
		t.Fatalf("one missed day is covered by the grace day: %+v", st)
	}
	st, adv, reset = advanceStreak(st, 105, rule)
	if !adv || !reset || st.Count != 1 || st.Run != 2 || st.GraceUsed != 0 || st.Best != 3 {
		t.Fatalf("a second miss resets the streak: %+v", st)
	}
	if !streakBroken(streakState{Count: 4, Last: last(10)}, 12, streakRule{}) || streakBroken(streakState{Count: 4, Last: last(10)}, 11, streakRule{}) {
		t.Fatal("broken means a period was missed beyond the grace")
	}
	back := streakRule{Grace: 0, MissAction: "step_back"}
	st, _, reset = advanceStreak(streakState{Count: 5, Best: 5, Run: 1, Last: last(10)}, 13, back)
	if reset || st.Count != 4 || st.Run != 1 {
		t.Fatalf("two missed days step back two and count today: %+v", st)
	}
	st, _, _ = advanceStreak(streakState{Count: 1, Run: 1, Last: last(10)}, 20, back)
	if st.Count != 1 || st.Run != 1 {
		t.Fatalf("stepping back never goes below day one: %+v", st)
	}
	old, adv, _ := advanceStreak(streakState{Count: 3, Run: 1, Last: last(50)}, 49, rule)
	if adv || old.Count != 3 {
		t.Fatal("an earlier period (clock skew) is ignored")
	}
}

func TestStreakRewardDayAndLevels(t *testing.T) {
	if streakRewardDay(1, 7) != 1 || streakRewardDay(7, 7) != 7 || streakRewardDay(8, 7) != 1 || streakRewardDay(15, 7) != 1 || streakRewardDay(9, 0) != 9 || streakRewardDay(0, 7) != 0 {
		t.Fatal("reward days repeat every cycle")
	}
	steps := []levelStep{{Level: 1, Threshold: 0}, {Level: 2, Threshold: 100}, {Level: 3, Threshold: 300}}
	if levelFor(steps, -5) != 0 || levelFor(steps, 0) != 1 || levelFor(steps, 99.99) != 1 || levelFor(steps, 100) != 2 || levelFor(steps, 5000) != 3 {
		t.Fatal("level is the highest step reached")
	}
	if metaRuleAmount("event", 2, 35) != 70 || metaRuleAmount("fixed", 10, 35) != 10 {
		t.Fatal("meta amount uses the event value only when asked")
	}
	if formatAmount(0.30000000000000004) != "0.3000" || formatAmount(12) != "12.0000" {
		t.Fatal("amounts keep four decimals")
	}
}
