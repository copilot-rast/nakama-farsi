package matchmakerkit

import (
	"errors"
	"testing"
)

func TestRemove(t *testing.T) {
	Set(nil)
	if err := Remove("s", "t"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("no remover: %v", err)
	}
	var got string
	Set(func(sessionID, ticket string) error { got = sessionID + "/" + ticket; return nil })
	defer Set(nil)
	if err := Remove("s1", "t1"); err != nil || got != "s1/t1" {
		t.Fatalf("%v %q", err, got)
	}
}
