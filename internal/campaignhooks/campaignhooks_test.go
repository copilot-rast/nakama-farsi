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

func TestGatesAndObserversRunInNameOrder(t *testing.T) {
	ctx := context.Background()
	order := []string{}
	SetStartGate("b", func(_ context.Context, c StartCheck) error { order = append(order, "b"); return nil })
	SetStartGate("a", func(_ context.Context, c StartCheck) error {
		order = append(order, "a")
		if c.ItemKey == "x" {
			return errors.New("blocked")
		}
		return nil
	})
	defer SetStartGate("a", nil)
	defer SetStartGate("b", nil)
	if err := CheckStart(ctx, StartCheck{ItemKey: "x"}); err == nil || len(order) != 1 || order[0] != "a" {
		t.Fatalf("first gate stops: %v %v", err, order)
	}
	if err := CheckStart(ctx, StartCheck{ItemKey: "y"}); err != nil || len(order) != 3 {
		t.Fatalf("all gates: %v %v", err, order)
	}
	if HasResultObservers() {
		t.Fatal("no observers yet")
	}
	SetResultObserver("ok", func(_ context.Context, r Result) (any, error) { return r.ItemKey, nil })
	SetResultObserver("bad", func(_ context.Context, r Result) (any, error) { return nil, errors.New("down") })
	SetResultObserver("none", func(_ context.Context, r Result) (any, error) { return nil, nil })
	defer SetResultObserver("ok", nil)
	defer SetResultObserver("bad", nil)
	defer SetResultObserver("none", nil)
	out, errs := Observe(ctx, Result{ItemKey: "car"})
	if out["ok"] != "car" || len(out) != 1 || errs["bad"] == nil || len(errs) != 1 {
		t.Fatalf("observe: %v %v", out, errs)
	}
}
