package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
)

type appError struct {
	Code    string
	Message string
	Field   string
	Status  int
}

func (e *appError) Error() string {
	if e.Message != "" {
		return e.Code + ": " + e.Message
	}
	return e.Code
}

func fail(status int, code, field, message string) *appError {
	return &appError{Code: code, Message: message, Field: field, Status: status}
}

func badRequest(code, field, message string) *appError {
	return fail(http.StatusBadRequest, code, field, message)
}

func notFound(code string) *appError {
	return fail(http.StatusNotFound, code, "", "")
}

func conflict(code, message string) *appError {
	return fail(http.StatusConflict, code, "", message)
}

func asApp(err error) (*appError, bool) {
	var app *appError
	if errors.As(err, &app) {
		return app, true
	}
	return nil, false
}

func grpcCode(status int) codes.Code {
	switch status {
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict, http.StatusPaymentRequired:
		return codes.FailedPrecondition
	case http.StatusServiceUnavailable, http.StatusBadGateway:
		return codes.Unavailable
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	default:
		return codes.Internal
	}
}

func isUniqueViolation(err error) (string, bool) {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
		return pg.ConstraintName, true
	}
	return "", false
}

func isForeignKeyViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == pgerrcode.ForeignKeyViolation
}

type clientRPC func(ctx context.Context, userID string, payload string) (any, error)

func (s *service) rpc(fn clientRPC) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if _, err := uuid.FromString(userID); err != nil {
			return "", runtime.NewError("UNAUTHENTICATED", int(codes.Unauthenticated))
		}
		if len(payload) > 16384 {
			return "", runtime.NewError("REQUEST_TOO_LARGE", int(codes.InvalidArgument))
		}
		out, err := fn(ctx, userID, payload)
		if err != nil {
			return "", s.runtimeError(err)
		}
		body, err := json.Marshal(out)
		if err != nil {
			return "", runtime.NewError("INTERNAL_ERROR", int(codes.Internal))
		}
		return string(body), nil
	}
}

func (s *service) runtimeError(err error) error {
	if app, ok := asApp(err); ok {
		return runtime.NewError(app.Code, int(grpcCode(app.Status)))
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return runtime.NewError("TIMEOUT", int(codes.DeadlineExceeded))
	case errors.Is(err, context.Canceled):
		return runtime.NewError("CANCELED", int(codes.Canceled))
	}
	s.logger.Error("Publication RPC failed: %v", err)
	return runtime.NewError("INTERNAL_ERROR", int(codes.Internal))
}

func decodeStrict(payload string, out any) error {
	if payload == "" {
		payload = "{}"
	}
	if err := json.Unmarshal([]byte(payload), out); err != nil {
		return badRequest("INVALID_JSON", "", "request body is not valid JSON for this call")
	}
	return nil
}
