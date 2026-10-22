package main

import (
	"testing"
	"time"
)

func TestLotteryDateMatchesSourceIntlFixtures(t *testing.T) {
	for _, tc := range []struct{ at, want string }{
		{"2026-10-22T12:00:00Z", "۳۰ مهر ۱۴۰۵"},
		{"2025-03-20T12:00:00Z", "۳۰ اسفند ۱۴۰۳"},
		{"2024-03-19T21:00:00Z", "۱ فروردین ۱۴۰۳"},
		{"2024-03-20T00:00:00Z", "۱ فروردین ۱۴۰۳"},
	} {
		at, err := time.Parse(time.RFC3339, tc.at)
		if err != nil {
			t.Fatal(err)
		}
		if got := lotteryDateFa(at); got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.at, got, tc.want)
		}
	}
}

func TestLotteryMonthsAgoSourceBuckets(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	if lotteryMonthsAgo(now.Add(-29*24*time.Hour), now) != 0 || lotteryMonthsAgo(now.Add(-30*24*time.Hour), now) != 1 {
		t.Fatal("wrong 30-day month bucket")
	}
}

func TestLotteryRaffleFormatting(t *testing.T) {
	if got := lotteryRialFa("1000000"); got != "۱٬۰۰۰٬۰۰۰ ریال" {
		t.Fatalf("rial format: %q", got)
	}
	at, _ := time.Parse(time.RFC3339, "2026-10-22T12:00:03Z")
	if got := lotteryDrawInstantFa(at); got != "۱۴۰۵/۰۷/۳۰ · ۱۵:۳۰:۰۳" {
		t.Fatalf("draw instant: %q", got)
	}
	if got := lotteryTrackingCode("12345678-aaaa-bbbb"); got != "RF-123456" {
		t.Fatalf("tracking code: %q", got)
	}
}
