package main

import (
	"context"
	"testing"
	"time"
)

func TestMedalSpecialEventEligibility(t *testing.T) {
	c := resolvedMedalCondition{Type: "special_event", EventKey: "campaign-launch"}
	for _, test := range []struct {
		key  string
		want bool
	}{{"campaign-launch", true}, {"other", false}, {"", false}} {
		got, err := medalEligible(context.Background(), nil, "", "", test.key, c, time.Now())
		if err != nil || got != test.want {
			t.Fatalf("key %q: got %v, %v", test.key, got, err)
		}
	}
}

func TestMedalUnsupportedConditionRejects(t *testing.T) {
	for _, c := range []resolvedMedalCondition{
		{Type: "activity_count", Activity: "untracked", Threshold: 1},
	} {
		if _, err := medalEligible(context.Background(), nil, "", "", "", c, time.Now()); err == nil {
			t.Fatalf("%s should require source ledger", c.Activity)
		}
	}
}

func TestMedalBackfillWorkerFlag(t *testing.T) {
	t.Setenv("JAZEBE_MEDAL_BACKFILL_ENABLED", "false")
	if err := startMedalBackfillWorker(context.Background(), nil, nil); err != nil {
		t.Fatalf("disabled worker: %v", err)
	}
	t.Setenv("JAZEBE_MEDAL_BACKFILL_ENABLED", "sometimes")
	if err := startMedalBackfillWorker(context.Background(), nil, nil); err == nil {
		t.Fatal("invalid worker flag accepted")
	}
}

func TestMedalDisplayIcon(t *testing.T) {
	for _, test := range []struct{ icon, want string }{
		{"heart", "heart"}, {"🏅", "🏅"}, {"not-a-renderable-icon", "star"}, {"", "star"},
	} {
		if got := medalDisplayIcon(test.icon); got != test.want {
			t.Fatalf("%q: got %q, want %q", test.icon, got, test.want)
		}
	}
}
