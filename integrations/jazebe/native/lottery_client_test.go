package main

import (
	"testing"
	"time"
)

func TestLotteryClientPresentation(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	if got := lotteryHoursUntil(now.Add(time.Minute), now); got != 1 {
		t.Fatalf("one minute should show one hour, got %d", got)
	}
	if got := lotteryHoursUntil(now.Add(-time.Second), now); got != 0 {
		t.Fatalf("closed draw should show zero, got %d", got)
	}
	if got := lotteryPrizeImage("discount_coupon", ""); got != "prize-percent" {
		t.Fatalf("coupon image: %s", got)
	}
	if got := lotteryPrizeImage("cash", " custom.png "); got != "custom.png" {
		t.Fatalf("custom image: %s", got)
	}
	if got := lotteryPrizeTypeFa("planet_points"); got != "امتیاز سیاره" {
		t.Fatalf("localized prize type: %s", got)
	}
}

func TestLotteryFinishedRangeGuards(t *testing.T) {
	for _, tc := range []struct {
		key  string
		want int
	}{{"", 999}, {"all", 999}, {"m1", 0}, {"m3", 2}, {"y", 11}} {
		got, err := lotteryRangeMax(tc.key)
		if err != nil || got != tc.want {
			t.Fatalf("%q: %d %v", tc.key, got, err)
		}
	}
	if _, err := lotteryRangeMax("unbounded"); err == nil {
		t.Fatal("accepted invalid range")
	}
}

func TestLotteryDetailLifecycle(t *testing.T) {
	if got := lotteryLifecycle("active", false); got != "nogroup" {
		t.Fatal(got)
	}
	if got := lotteryLifecycle("active", true); got != "active" {
		t.Fatal(got)
	}
	for _, status := range []string{"drawing", "finished", "cancelled"} {
		if got := lotteryLifecycle(status, false); got != status {
			t.Fatalf("%s: %s", status, got)
		}
	}
}
