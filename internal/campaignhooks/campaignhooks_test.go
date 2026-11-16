package campaignhooks

import (
	"context"
	"errors"
	"testing"
)

func TestSlotRoutesCallsAndOwners(t *testing.T) {
	ctx := context.Background()
	if _, err := Start(ctx, "u", "s"); !errors.Is(err, ErrUnavailable) || Available() {
		t.Fatalf("no campaign: %v", err)
	}
	SetCampaign(func(_ context.Context, u, s string) (string, error) { return u + ":" + s, nil },
		func(_ context.Context, u string, c Completion) (map[string]any, error) {
			return map[string]any{"user": u, "attempt": c.AttemptID, "success": c.Success}, nil
		})
	defer SetCampaign(nil, nil)
	if id, err := Start(ctx, "u", "s"); err != nil || id != "u:s" {
		t.Fatalf("start: %v %v", id, err)
	}
	if out, err := Complete(ctx, "u", Completion{AttemptID: "a", Success: true}); err != nil || out["attempt"] != "a" || out["success"] != true {
		t.Fatalf("complete: %v %v", out, err)
	}
	if o, _ := OwnerOf(ctx, "quiz-1"); o != "" {
		t.Fatalf("owner without registration: %q", o)
	}
	SetOwner("quiz", func(_ context.Context, k string) (string, bool) { return "pack-" + k, k == "quiz-1" })
	defer SetOwner("quiz", nil)
	if o, r := OwnerOf(ctx, "quiz-1"); o != "quiz" || r != "pack-quiz-1" {
		t.Fatalf("owner: %q %q", o, r)
	}
	if o, _ := OwnerOf(ctx, "story-1"); o != "" {
		t.Fatalf("not owned stage: %q", o)
	}
}
