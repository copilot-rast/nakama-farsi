package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
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
