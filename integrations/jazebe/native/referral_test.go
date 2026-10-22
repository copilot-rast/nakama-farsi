package main

import (
	"context"
	"testing"
	"time"
)

func TestReferralQualificationRules(t *testing.T) {
	for _, tc := range []struct {
		raw    string
		want   []string
		target int
	}{
		{`{}`, []string{"registered", "qualified"}, 1},
		{`{"scope":"group","successCondition":"signup","targetCount":3}`, []string{"qualified"}, 3},
		{`{"scope":"planet","successCondition":"first_activity"}`, []string{"qualified"}, 1},
	} {
		c, err := parseReferralMissionConfig([]byte(tc.raw))
		if err != nil {
			t.Fatal(err)
		}
		got := referralWantedStatuses(c)
		if len(got) != len(tc.want) {
			t.Fatalf("%s: %v", tc.raw, got)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("%s: %v", tc.raw, got)
			}
		}
		if referralTarget(c) != tc.target {
			t.Fatalf("%s: target %d", tc.raw, referralTarget(c))
		}
	}
	for _, bad := range []string{`{"scope":"other"}`, `{"successCondition":"click"}`, `{"windowDays":-1}`, `{broken`} {
		if _, err := parseReferralMissionConfig([]byte(bad)); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
}

func TestReferralWorkerConfiguration(t *testing.T) {
	t.Setenv("JAZEBE_REFERRAL_WORKER_ENABLED", "false")
	if err := startReferralCompletionWorker(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_REFERRAL_WORKER_ENABLED", "sometimes")
	if err := startReferralCompletionWorker(context.Background(), nil, nil); err == nil {
		t.Fatal("invalid worker flag accepted")
	}
}

func TestInviteMissionDraftValidation(t *testing.T) {
	valid := inviteMissionDraft{Title: "Invite", RepeatPattern: "once", TargetType: "solo", StartsAt: "2026-10-22T10:00:00Z", EndsAt: "2026-10-23T10:00:00Z",
		Invite: referralMissionConfig{Scope: "planet", SuccessCondition: "first_activity", TargetCount: 2}}
	if _, _, err := validateInviteMissionDraft(valid); err != nil {
		t.Fatal(err)
	}
	bad := valid
	bad.Invite.TargetCount = 0
	if _, _, err := validateInviteMissionDraft(bad); err == nil {
		t.Fatal("zero invite target accepted")
	}
	bad = valid
	bad.RepeatPattern = "periodic"
	bad.RepeatPeriod = ""
	if _, _, err := validateInviteMissionDraft(bad); err == nil {
		t.Fatal("periodic mission without cycle accepted")
	}
}

func TestReferralWindowAndTokenShape(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	if referralWindow(referralMissionConfig{}, now) != nil {
		t.Fatal("window unexpectedly limited")
	}
	start := referralWindow(referralMissionConfig{WindowDays: 7}, now).(time.Time)
	if !start.Equal(now.Add(-7 * 24 * time.Hour)) {
		t.Fatal("wrong referral window")
	}
	token, err := referralToken(32)
	if err != nil || !referralTokenPattern.MatchString(token) {
		t.Fatalf("bad invite token %q: %v", token, err)
	}
	for _, bad := range []string{"", "x", "has space", "abc/def"} {
		if referralTokenPattern.MatchString(bad) {
			t.Errorf("accepted %q", bad)
		}
	}
}
