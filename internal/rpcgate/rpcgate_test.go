package rpcgate

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestCheck(t *testing.T) {
	Set(nil)
	if Active() {
		t.Fatal("no gate")
	}
	if c, err := Check(context.Background(), "x", "u"); err != nil || c != codes.OK {
		t.Fatal("open without a gate")
	}
	Set(func(_ context.Context, id, user string) (codes.Code, error) {
		if id == "blocked" {
			return codes.Unavailable, errors.New("MAINTENANCE")
		}
		return codes.OK, nil
	})
	defer Set(nil)
	if c, err := Check(context.Background(), "blocked", "u"); err == nil || c != codes.Unavailable {
		t.Fatal("blocked rpc passed")
	}
	if _, err := Check(context.Background(), "blocked", ""); err != nil {
		t.Fatal("server calls without a user are never gated")
	}
	if _, err := Check(context.Background(), "open", "u"); err != nil {
		t.Fatal("other rpc refused")
	}
}
