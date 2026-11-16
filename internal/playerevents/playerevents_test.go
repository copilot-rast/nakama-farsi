package playerevents

import (
	"context"
	"errors"
	"testing"
)

func TestRecordGoesToTheRegisteredRecorder(t *testing.T) {
	SetRecorder(nil)
	if Available() {
		t.Fatal("no recorder yet")
	}
	if _, err := Record(context.Background(), Event{UserID: "u", Key: "login"}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err=%v", err)
	}
	var got Event
	SetRecorder(func(_ context.Context, ev Event) (Result, error) {
		got = ev
		return Result{"ok": true}, nil
	})
	defer SetRecorder(nil)
	if _, err := Record(context.Background(), Event{UserID: " ", Key: "login"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
	out, err := Record(context.Background(), Event{UserID: " u1 ", Key: " round_finished ", Value: 42})
	if err != nil || out["ok"] != true || got.UserID != "u1" || got.Key != "round_finished" || got.Value != 42 {
		t.Fatalf("out=%v err=%v got=%+v", out, err, got)
	}
}

func TestNotifyReachesObservers(t *testing.T) {
	var got []string
	Observe("a", func(_ context.Context, ev Recorded) { got = append(got, ev.ID+":"+ev.Key) })
	defer Observe("a", nil)
	Notify(context.Background(), Recorded{ID: "e1", UserID: "u", Key: "xp", Value: 3, Source: "client"})
	Observe("a", nil)
	Notify(context.Background(), Recorded{ID: "e2", UserID: "u", Key: "xp"})
	if len(got) != 1 || got[0] != "e1:xp" {
		t.Fatalf("got=%v", got)
	}
}
