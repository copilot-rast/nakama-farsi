package cuphooks

import (
	"context"
	"errors"
	"testing"
)

func TestRegisterLookupAndReport(t *testing.T) {
	if _, ok := Lookup("x"); ok {
		t.Fatal("nothing registered yet")
	}
	Register(Game{Key: "b", Launch: func(context.Context, Launch) (Launched, error) { return Launched{GameMatchID: "m1"}, nil }})
	Register(Game{Key: "a", Launch: func(context.Context, Launch) (Launched, error) { return Launched{}, nil }})
	Register(Game{Key: "", Launch: func(context.Context, Launch) (Launched, error) { return Launched{}, nil }})
	defer Unregister("a")
	defer Unregister("b")
	list := Games()
	if len(list) != 2 || list[0].Key != "a" || list[1].Key != "b" {
		t.Fatalf("games %v", list)
	}
	g, _ := Lookup("b")
	if out, _ := g.Launch(context.Background(), Launch{}); out.GameMatchID != "m1" {
		t.Fatal("launch not called")
	}
	if err := Report(context.Background(), Result{GameMatchID: "m"}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("report without the plugin: %v", err)
	}
	var got Result
	SetReporter(func(_ context.Context, r Result) error { got = r; return nil })
	defer SetReporter(nil)
	if err := Report(context.Background(), Result{}); err == nil {
		t.Fatal("a result without a match id is refused")
	}
	if err := Report(context.Background(), Result{GameMatchID: "m", Winner: "u"}); err != nil || got.Winner != "u" || !Available() {
		t.Fatalf("report %v %v", err, got)
	}
}
