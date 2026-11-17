package clanhooks

import (
	"context"
	"testing"
)

func TestHooksAnswerNothingUntilSet(t *testing.T) {
	ctx := context.Background()
	if _, ok := ClanOf(ctx, "u"); ok || Available() || Members(ctx, "c") != nil {
		t.Fatal("hooks must be empty before the clan plugin sets them")
	}
	if _, ok := WarRival(ctx, "a", "b"); ok {
		t.Fatal("no rival without the clan plugin")
	}
	Set(func(context.Context, string) (Membership, bool) { return Membership{ClanID: "c", Role: "leader"}, true },
		func(context.Context, string) []string { return []string{"a", "b"} },
		func(_ context.Context, a, b string) (Rival, bool) { return Rival{WarID: "w"}, a != b })
	defer Set(nil, nil, nil)
	if m, ok := ClanOf(ctx, "u"); !ok || m.ClanID != "c" {
		t.Fatalf("member %v %v", m, ok)
	}
	if len(Members(ctx, "c")) != 2 {
		t.Fatal("members")
	}
	if _, ok := WarRival(ctx, "a", "a"); ok {
		t.Fatal("a player is never their own rival")
	}
	if r, ok := WarRival(ctx, "a", "b"); !ok || r.WarID != "w" {
		t.Fatalf("rival %v %v", r, ok)
	}
}
