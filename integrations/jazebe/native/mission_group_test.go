package main

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGroupMissionRewardShape(t *testing.T) {
	if err := validateMissionTargetRewards("group", 2, 0, 1, "lottery"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		target                string
		goal, points, tickets int
		lottery               string
	}{
		{"group", 0, 0, 1, "lottery"},
		{"group", 2, 1, 0, ""},
		{"group", 2, 0, 1, ""},
		{"unknown", 0, 0, 0, ""},
	} {
		if err := validateMissionTargetRewards(tc.target, tc.goal, tc.points, tc.tickets, tc.lottery); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("accepted invalid reward %#v: %v", tc, err)
		}
	}
}

func TestCodeAndContentGroupDraftRules(t *testing.T) {
	code := codeMissionDraft{Title: "Code", DetectionMethod: "platform_auto", TargetType: "group", GroupTargetValue: 2, RewardTickets: 1, TargetLotteryID: "lottery", StartsAt: "2026-10-22T00:00:00Z", EndsAt: "2026-10-23T00:00:00Z"}
	code.Code.CodeMode = "shared"
	code.Code.MaxUses = 2
	if _, _, err := validateCodeDraft(code); err != nil {
		t.Fatal(err)
	}
	code.RewardPoints = 1
	if _, _, err := validateCodeDraft(code); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("group points accepted: %v", err)
	}
	content := contentMissionDraft{Title: "Evidence", TargetType: "group", GroupTargetValue: 2, RewardTickets: 1, TargetLotteryID: "lottery", StartsAt: "2026-10-22T00:00:00Z", EndsAt: "2026-10-23T00:00:00Z"}
	content.Content.EvidenceTypes = []string{"text"}
	if _, _, err := validateContentDraft(content); err != nil {
		t.Fatal(err)
	}
}
