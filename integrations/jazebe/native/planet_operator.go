package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type operatorTransition struct {
	from, to, timestamp string
	reasonRequired      bool
}

var planetOperatorTransitions = map[string]operatorTransition{
	"approve":   {from: "pending_review", to: "active", timestamp: "activated_at = COALESCE(activated_at, now())"},
	"reject":    {from: "pending_review", to: "setup", reasonRequired: true},
	"suspend":   {from: "active", to: "suspended", timestamp: "suspended_at = now()", reasonRequired: true},
	"unsuspend": {from: "suspended", to: "active", timestamp: "suspended_at = NULL"},
}

func planetOperatorAction(ctx context.Context, logger runtime.Logger, db *sql.DB, payload, action string) (string, error) {
	operatorID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(operatorID, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "Jazebe operator role required")
	}
	var input struct {
		PlanetID string `json:"planetId"`
		Reason   string `json:"reason"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	id, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	reason := strings.TrimSpace(input.Reason)
	if len(reason) > 512 {
		return "", status.Error(codes.InvalidArgument, "reason is too long")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "planet transition failed")
	}
	defer tx.Rollback()
	var from, to, setClause string
	if action == "archive" {
		to, setClause = "archived", "archived_at = now()"
		if reason == "" {
			return "", status.Error(codes.InvalidArgument, "reason is required")
		}
		var current string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM jazebe_planet WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", status.Error(codes.NotFound, "planet not found")
			}
			return "", status.Error(codes.Internal, "planet transition failed")
		}
		if current == "archived" {
			return "", status.Error(codes.FailedPrecondition, "planet already archived")
		}
		from = current
	} else {
		rule := planetOperatorTransitions[action]
		from, to, setClause = rule.from, rule.to, rule.timestamp
		if rule.reasonRequired && reason == "" {
			return "", status.Error(codes.InvalidArgument, "reason is required")
		}
	}
	query := `UPDATE jazebe_planet SET status=$2`
	if setClause != "" {
		query += ", " + setClause
	}
	query += ` WHERE id=$1 AND status=$3 RETURNING status`
	var updated string
	if err := tx.QueryRowContext(ctx, query, id, to, from).Scan(&updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.FailedPrecondition, "invalid planet transition or planet not found")
		}
		logger.Error("Jazebe operator transition failed: %v", err)
		return "", status.Error(codes.Internal, "planet transition failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log
		(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ($1,'operator',$2,'planet',$3,$3,jsonb_build_object('from',$4::text,'to',$5::text,'reason',$6::text))`,
		"planet."+action, operatorID, id, from, to, reason); err != nil {
		logger.Error("Jazebe planet audit write failed: %v", err)
		return "", status.Error(codes.Internal, "planet transition failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "planet transition failed")
	}
	return encode(map[string]any{"planetId": id, "status": updated, "from": from})
}

func planetReject(ctx context.Context, l runtime.Logger, db *sql.DB, _ runtime.NakamaModule, p string) (string, error) {
	return planetOperatorAction(ctx, l, db, p, "reject")
}
func planetApprove(ctx context.Context, l runtime.Logger, db *sql.DB, _ runtime.NakamaModule, p string) (string, error) {
	return planetOperatorAction(ctx, l, db, p, "approve")
}
func planetSuspend(ctx context.Context, l runtime.Logger, db *sql.DB, _ runtime.NakamaModule, p string) (string, error) {
	return planetOperatorAction(ctx, l, db, p, "suspend")
}
func planetUnsuspend(ctx context.Context, l runtime.Logger, db *sql.DB, _ runtime.NakamaModule, p string) (string, error) {
	return planetOperatorAction(ctx, l, db, p, "unsuspend")
}
func planetArchive(ctx context.Context, l runtime.Logger, db *sql.DB, _ runtime.NakamaModule, p string) (string, error) {
	return planetOperatorAction(ctx, l, db, p, "archive")
}
