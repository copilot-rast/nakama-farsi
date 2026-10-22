package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type jazebeRPC func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error)

// Nakama's token cache is local to a node. Every migrated RPC checks the
// authoritative database state, including when a token was issued before a
// restriction or arrives on another node. This does not alter Nakama core APIs.
func guardJazebeRPC(next jazebeRPC) jazebeRPC {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		caller, err := userID(ctx)
		if err != nil {
			return "", err
		}
		var active bool
		if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users u
			LEFT JOIN jazebe_user_account a ON a.user_id=u.id
			WHERE u.id=$1 AND u.disable_time='1970-01-01 00:00:00 UTC'
			AND COALESCE(a.state,'active')='active')`, caller).Scan(&active); err != nil {
			return "", err
		}
		if !active {
			return "", status.Error(codes.PermissionDenied, "account is not active")
		}
		return next(ctx, logger, db, nk, payload)
	}
}
