package inapp

import (
	"context"
	"errors"
	"testing"
)

func TestSenderRegistry(t *testing.T) {
	SetSender(nil)
	if Available() {
		t.Fatal("no sender yet")
	}
	if _, err := Send(context.Background(), Message{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err %v", err)
	}
	SetSender(func(_ context.Context, m Message) (string, error) { return m.UserIDs[0], nil })
	defer SetSender(nil)
	if id, err := Send(context.Background(), Message{UserIDs: []string{"u1"}}); err != nil || id != "u1" {
		t.Fatalf("send %q %v", id, err)
	}
}
