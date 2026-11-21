package insightsapi

import (
	"context"
	"errors"
	"testing"
)

func TestSendNeedsASinkAndPassesEvents(t *testing.T) {
	Register(nil)
	if _, err := Send(context.Background(), Event{Name: "x"}); !errors.Is(err, ErrNoSink) {
		t.Fatalf("no sink: %v", err)
	}
	if r, err := Send(context.Background()); err != nil || r.Accepted != 0 {
		t.Fatalf("empty send: %v %v", r, err)
	}
	var got []Event
	Register(func(_ context.Context, ev []Event) (Result, error) { got = ev; return Result{Accepted: len(ev)}, nil })
	defer Register(nil)
	if !Enabled() {
		t.Fatal("enabled")
	}
	r, err := Send(context.Background(), Event{Name: "DistanceOverall", Tags: Tags("car", "1", "mission", "2"), Value: Num(150)}, Path("Distance:SUV", 3))
	if err != nil || r.Accepted != 2 || len(got) != 2 || got[0].Tags[1].Group != "mission" || *got[0].Value != 150 || got[1].Path != "Distance:SUV" || *got[1].Value != 3 {
		t.Fatalf("send: %+v %v %+v", r, err, got)
	}
}
