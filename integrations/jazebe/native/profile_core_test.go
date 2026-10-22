package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProfileNameAndAvatarBounds(t *testing.T) {
	if got, err := validProfileName("  مهرسا  "); err != nil || got != "مهرسا" {
		t.Fatalf("Persian display name: %q %v", got, err)
	}
	for _, name := range []string{" ", "a", "a\nname", "123456789012345678901"} {
		if _, err := validProfileName(name); status.Code(err) != codes.InvalidArgument {
			t.Errorf("invalid name %q accepted: %v", name, err)
		}
	}
	for _, index := range []int{-1, 0, 5} {
		if err := validAvatar(index); err != nil {
			t.Errorf("index %d: %v", index, err)
		}
	}
	for _, index := range []int{-2, 6} {
		if status.Code(validAvatar(index)) != codes.InvalidArgument {
			t.Errorf("index %d accepted", index)
		}
	}
}

func TestPersianMonthKeyUsesTehranCalendar(t *testing.T) {
	before := time.Date(2026, 3, 20, 18, 0, 0, 0, time.UTC)
	after := time.Date(2026, 3, 20, 21, 0, 0, 0, time.UTC)
	if got := persianMonthKey(before); got != "1404-12" {
		t.Fatalf("before Nowruz: %s", got)
	}
	if got := persianMonthKey(after); got != "1405-01" {
		t.Fatalf("after Nowruz: %s", got)
	}
}

func TestProfileLevelDerivedFromXP(t *testing.T) {
	thresholds, _ := json.Marshal([]profileLevel{{Level: 1, Key: "newcomer", TitleFa: "تازه‌وارد", MinXp: 0},
		{Level: 2, Key: "active", TitleFa: "عضو فعال", MinXp: 100},
		{Level: 3, Key: "veteran", TitleFa: "کهنه‌کار", MinXp: 300}})
	for _, tc := range []struct {
		xp    int64
		level int
		max   int64
	}{{0, 1, 100}, {99, 1, 100}, {100, 2, 300}, {300, 3, 300}} {
		level, _, max := profileLevelFor(tc.xp, thresholds)
		if level != tc.level || max != tc.max {
			t.Errorf("xp %d: level %d max %d", tc.xp, level, max)
		}
	}
}

func TestMissionReferralAttributionRequiresInternalSignupContext(t *testing.T) {
	if referralMissionAcceptAllowed(context.Background()) {
		t.Fatal("ordinary referral RPC context can claim mission token")
	}
	if !referralMissionAcceptAllowed(context.WithValue(context.Background(), referralSignupContextKey{}, true)) {
		t.Fatal("first profile setup cannot emit signup attribution")
	}
}
