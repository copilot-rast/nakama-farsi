package main

import (
	"context"
	"database/sql"
	"errors"

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

// nakamaRPCError converts the module's gRPC status errors into Nakama runtime
// errors. Nakama maps only *runtime.Error to a client status code; any other
// error reaches clients as 13 Internal (HTTP 500) carrying its raw text. Errors
// without a status, such as driver errors, are logged and replaced with a
// generic message so SQL and schema details are not disclosed.
func nakamaRPCError(next jazebeRPC) jazebeRPC {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		result, err := next(ctx, logger, db, nk, payload)
		if err == nil {
			return result, nil
		}
		if runtimeErr, ok := errors.AsType[*runtime.Error](err); ok {
			return result, runtimeErr
		}
		if st, ok := status.FromError(err); ok && st.Code() != codes.OK && st.Code() != codes.Unknown {
			return result, runtime.NewError(st.Message(), int(st.Code()))
		}
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return result, runtime.NewError("request timed out", int(codes.DeadlineExceeded))
		case errors.Is(err, context.Canceled):
			return result, runtime.NewError("request canceled", int(codes.Canceled))
		}
		if logger != nil {
			logger.Error("Jazebe RPC failed: %v", err)
		}
		return result, runtime.NewError("internal error", int(codes.Internal))
	}
}
