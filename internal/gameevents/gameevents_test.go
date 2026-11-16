package gameevents

import (
	"context"
	"errors"
	"testing"
)

func TestPublishReachesEverySubscriberAndKeepsErrorsApart(t *testing.T) {
	if a, e := Publish(context.Background(), Event{ID: "1", UserID: "u", Name: "x"}); len(a) != 0 || len(e) != 0 {
		t.Fatalf("no subscribers: %v %v", a, e)
	}
	var seen []string
	Subscribe("b", func(_ context.Context, ev Event) (any, error) {
		seen = append(seen, "b")
		return nil, errors.New("down")
	})
	Subscribe("a", func(_ context.Context, ev Event) (any, error) { seen = append(seen, "a:"+ev.Context); return 7, nil })
	defer func() { Subscribe("a", nil); Subscribe("b", nil) }()
	a, e := Publish(context.Background(), Event{ID: "1", UserID: "u", Name: "stage_completed", Context: "s1"})
	if len(seen) != 2 || seen[0] != "a:s1" || a["a"] != 7 || e["b"] == nil {
		t.Fatalf("seen=%v answers=%v errs=%v", seen, a, e)
	}
	if _, e := Publish(context.Background(), Event{UserID: "u", Name: "x"}); !errors.Is(e[""], ErrInvalid) {
		t.Fatalf("missing id accepted: %v", e)
	}
}

func TestPublishTypedOnlyAndHelpers(t *testing.T) {
	var got Event
	Subscribe("t", func(_ context.Context, ev Event) (any, error) { got = ev; return nil, nil })
	defer Subscribe("t", nil)
	if !Has() {
		t.Fatal("Has after subscribe")
	}
	score := int64(3)
	_, e := Publish(context.Background(), Event{ID: "1", UserID: "u", Source: " campaign ", Typed: []Typed{Progression("Complete", "c1", "s1", "", &score), Resource("Source", "coin", 50, "Stage", "s1")}})
	if len(e) != 0 || got.Source != "campaign" || len(got.Typed) != 2 || *got.Typed[1].Amount != 50 || *got.Typed[0].Score != 3 {
		t.Fatalf("typed only: %v %+v", e, got)
	}
	b := Business("IRR", 5000, "product", "p1", "shop", "bazaar", "r1")
	if b.Type != "business" || *b.Amount != 5000 {
		t.Fatalf("business %+v", b)
	}
}
