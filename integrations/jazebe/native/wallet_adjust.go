package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var adjustmentKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{8,128}$`)

type walletAdjustmentInput struct {
	PlanetID       string `json:"planetId"`
	UserID         string `json:"userId"`
	Direction      string `json:"direction"`
	Amount         int64  `json:"amount"`
	Reason         string `json:"reason"`
	Reference      string `json:"reference"`
	IdempotencyKey string `json:"idempotencyKey"`
}

func validateWalletAdjustment(input walletAdjustmentInput) (walletAdjustmentInput, error) {
	var err error
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return input, err
	}
	input.UserID, err = requiredUUID(input.UserID)
	if err != nil {
		return input, err
	}
	input.Reason = strings.TrimSpace(input.Reason)
	input.Reference = strings.TrimSpace(input.Reference)
	if input.Direction != "credit" && input.Direction != "debit" {
		return input, status.Error(codes.InvalidArgument, "invalid direction")
	}
	if input.Amount < 1 || input.Amount > 1_000_000_000 {
		return input, status.Error(codes.InvalidArgument, "amount must be 1 to 1000000000")
	}
	if !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) < 3 || utf8.RuneCountInString(input.Reason) > 512 ||
		!utf8.ValidString(input.Reference) || utf8.RuneCountInString(input.Reference) > 128 {
		return input, status.Error(codes.InvalidArgument, "invalid adjustment reason or reference")
	}
	if !adjustmentKeyPattern.MatchString(input.IdempotencyKey) {
		return input, status.Error(codes.InvalidArgument, "invalid idempotency key")
	}
	return input, nil
}

// Source-compatible deterministic UUID for a scoped idempotency key.
func adjustmentSourceID(scope string) string {
	sum := sha256.Sum256([]byte(scope))
	h := hex.EncodeToString(sum[:])
	v := (int(sum[8]>>4) & 3) | 8
	return fmt.Sprintf("%s-%s-5%s-%x%s-%s", h[:8], h[8:12], h[13:16], v, h[17:20], h[20:32])
}

func walletAdjust(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "Jazebe operator role required")
	}
	var input walletAdjustmentInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input, err = validateWalletAdjustment(input)
	if err != nil {
		return "", err
	}
	large := input.Amount > 10_000
	if large && !isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "super operator required for large adjustment")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "wallet adjustment failed")
	}
	defer tx.Rollback()
	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE id=$1)`, input.PlanetID).Scan(&exists)
	if err != nil {
		return "", status.Error(codes.Internal, "wallet adjustment failed")
	}
	if !exists {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, input.UserID).Scan(&exists)
	if err != nil {
		return "", status.Error(codes.Internal, "wallet adjustment failed")
	}
	if !exists {
		return "", status.Error(codes.NotFound, "user not found")
	}
	sourceID := adjustmentSourceID("adjust:" + actor + ":" + input.PlanetID + ":" + input.UserID + ":" + input.IdempotencyKey)
	description := input.Reason
	if input.Reference != "" {
		description += " (" + input.Reference + ")"
	}
	if utf8.RuneCountInString(description) > 256 {
		description = string([]rune(description)[:256])
	}
	movement, err := applyPointMovement(ctx, tx, input.PlanetID, input.UserID, input.Direction, input.Amount, "operator_adjustment", sourceID, description, time.Now().UTC())
	if errors.Is(err, errInsufficientPoints) {
		return "", status.Error(codes.FailedPrecondition, "insufficient points")
	}
	if err != nil {
		logger.Error("Wallet adjustment: %v", err)
		return "", status.Error(codes.Internal, "wallet adjustment failed")
	}
	if movement.Applied {
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
			VALUES ('wallet.adjust','operator',$1,'wallet',$2,$3,jsonb_build_object('adjustmentId',$4::text,'transactionId',$5::text,
			'userId',$6::text,'direction',$7::text,'amount',$8::bigint,'balanceAfter',$9::bigint,'reason',$10::text,'reference',$11::text,'large',$12::boolean))`, actor, movement.WalletID, input.PlanetID,
			sourceID, movement.TransactionID, input.UserID, input.Direction, input.Amount, movement.Balance, input.Reason, input.Reference, large)
		if err != nil {
			return "", status.Error(codes.Internal, "wallet adjustment audit failed")
		}
		if err := persistJazebeNotification(ctx, tx, input.UserID, "wallet-adjust:"+sourceID, "موجودی کیف پول تغییر کرد", map[string]any{
			"category": "system", "actionLink": "/wallet", "planetId": input.PlanetID,
			"transactionId": movement.TransactionID, "direction": input.Direction, "amount": input.Amount, "reason": input.Reason}, time.Now().UTC()); err != nil {
			return "", status.Error(codes.Internal, "wallet adjustment notice failed")
		}
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "wallet adjustment failed")
	}
	return encode(map[string]any{"transactionId": movement.TransactionID, "balance": movement.Balance, "large": large, "applied": movement.Applied})
}
