package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGuardJazebeRPCRequiresIdentity(t *testing.T) {
	called := false
	guarded := guardJazebeRPC(func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
		called = true
		return "", nil
	})
	if _, err := guarded(context.Background(), nil, nil, nil, "{}"); err == nil || called {
		t.Fatalf("anonymous caller reached Jazebe RPC: called=%v err=%v", called, err)
	}
}

func TestNakamaRPCErrorMapsStatusCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   error
		code int
		msg  string
	}{
		{"permission", status.Error(codes.PermissionDenied, "Jazebe operator role required"), 7, "Jazebe operator role required"},
		{"invalid", status.Error(codes.InvalidArgument, "name required"), 3, "name required"},
		{"notfound", status.Error(codes.NotFound, "planet not found"), 5, "planet not found"},
		{"internal", status.Error(codes.Internal, "brand creation failed"), 13, "brand creation failed"},
		{"runtime", runtime.NewError("already typed", 9), 9, "already typed"},
		{"driver", errors.New(`ERROR: relation "jazebe_secret" does not exist (SQLSTATE 42P01)`), 13, "internal error"},
		{"wrapped", fmt.Errorf("tx: %w", sql.ErrConnDone), 13, "internal error"},
		{"deadline", context.DeadlineExceeded, 4, "request timed out"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := nakamaRPCError(func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
				return "", tc.in
			})
			_, err := wrapped(context.Background(), nil, nil, nil, "{}")
			var re *runtime.Error
			if !errors.As(err, &re) || re.Code != tc.code || re.Message != tc.msg {
				t.Fatalf("got %#v, want code=%d msg=%q", err, tc.code, tc.msg)
			}
		})
	}
	ok := nakamaRPCError(func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
		return `{"ok":true}`, nil
	})
	if out, err := ok(context.Background(), nil, nil, nil, "{}"); err != nil || out != `{"ok":true}` {
		t.Fatalf("success path changed: %q %v", out, err)
	}
}
