package dayclock

import (
	"testing"
	"time"
)

func TestLegacyResetAt0130UTC(t *testing.T) {
	c := New("UTC", 90, nil)
	cases := []struct {
		at   string
		want string
	}{
		{"2026-11-21T01:29:59Z", "2026-11-20"},
		{"2026-11-21T01:30:00Z", "2026-11-21"},
		{"2026-11-21T23:59:59Z", "2026-11-21"},
		{"2026-11-22T01:00:00Z", "2026-11-21"},
	}
	for _, tc := range cases {
		at, _ := time.Parse(time.RFC3339, tc.at)
		key, start, end := c.Bounds(at)
		if key != tc.want {
			t.Fatalf("%s: day %s, want %s", tc.at, key, tc.want)
		}
		if at.Before(start) || !at.Before(end) || end.Sub(start) != 24*time.Hour {
			t.Fatalf("%s: bounds %s..%s", tc.at, start, end)
		}
	}
}

func TestMidnightDefaultMatchesCalendarDay(t *testing.T) {
	tehran, _ := time.LoadLocation("Asia/Tehran")
	c := New("", 0, tehran)
	at, _ := time.Parse(time.RFC3339, "2026-11-21T20:29:00Z")
	key, start, _ := c.Bounds(at)
	if key != "2026-11-21" || !start.Equal(time.Date(2026, 11, 21, 0, 0, 0, 0, tehran)) {
		t.Fatalf("got %s %s", key, start)
	}
	at, _ = time.Parse(time.RFC3339, "2026-11-21T20:31:00Z")
	if key, _, _ := c.Bounds(at); key != "2026-11-22" {
		t.Fatalf("got %s", key)
	}
}

func TestBadZoneFallsBack(t *testing.T) {
	c := New("Nope/Zone", 2000, time.UTC)
	if c.Loc != time.UTC || c.StartMinute != 0 {
		t.Fatalf("got %+v", c)
	}
	if ValidZone("Nope/Zone") || !ValidZone("") || !ValidZone("UTC") {
		t.Fatal("zone check")
	}
}
