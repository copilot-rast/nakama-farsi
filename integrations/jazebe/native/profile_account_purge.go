package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// operatorAccountPurge is the legal erasure request (source
// admin-support.service.ts purge, PRD 12.2): super operators only, never for
// an active account. Personal data is deleted and the identity anonymized;
// ledgers, memberships, lottery records and audits stay under the anonymized
// user, as in the source. The Nakama user row is kept and banned, so the
// account can no longer sign in.
func operatorAccountPurge(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "super operator required")
	}
	var in struct {
		UserID string `json:"userId"`
		Reason string `json:"reason"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.UserID, err = requiredUUID(in.UserID); err != nil {
		return "", err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 3 || utf8.RuneCountInString(in.Reason) > 512 {
		return "", status.Error(codes.InvalidArgument, "reason must be 3 to 512 characters")
	}
	if in.UserID == actor {
		return "", status.Error(codes.PermissionDenied, "operator cannot purge own account")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "purge unavailable")
	}
	defer tx.Rollback()
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, in.UserID).Scan(&exists); err != nil {
		return "", status.Error(codes.Internal, "purge unavailable")
	}
	if !exists {
		return "", status.Error(codes.NotFound, "user not found")
	}
	var state string
	var purged sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT state,purged_at FROM jazebe_user_account WHERE user_id=$1 FOR UPDATE`, in.UserID).Scan(&state, &purged)
	if errors.Is(err, sql.ErrNoRows) {
		state = "active"
	} else if err != nil {
		return "", status.Error(codes.Internal, "purge unavailable")
	}
	if purged.Valid {
		return "", status.Error(codes.AlreadyExists, "account already purged")
	}
	if state == "active" {
		return "", status.Error(codes.FailedPrecondition, "purge allowed only for suspended or deactivated accounts")
	}
	var erased []string
	del := func(label, query string) error {
		res, err := tx.ExecContext(ctx, query, in.UserID)
		if err != nil {
			logger.Error("Jazebe purge %s failed: %v", label, err)
			return status.Error(codes.Internal, "purge failed")
		}
		if n, _ := res.RowsAffected(); n > 0 {
			erased = append(erased, label)
		}
		return nil
	}
	for _, step := range []struct{ label, query string }{
		{"kyc", `DELETE FROM jazebe_user_verification WHERE user_id=$1`},
		{"notifications", `DELETE FROM notification WHERE user_id=$1 AND code=1001`},
		{"notification_preferences", `DELETE FROM jazebe_notification_preference WHERE user_id=$1`},
		{"devices", `DELETE FROM user_device WHERE user_id=$1`},
	} {
		if err := del(step.label, step.query); err != nil {
			return "", err
		}
	}
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", status.Error(codes.Internal, "purge failed")
	}
	anon := "حذف‌شده-" + hex.EncodeToString(suffix[:])
	// Nakama identity: credentials, social IDs and profile fields are erased;
	// the username keeps the row unique; disable_time blocks sign-in.
	if _, err := tx.ExecContext(ctx, `UPDATE users SET username=$2,display_name=$3,avatar_url=NULL,email=NULL,password=NULL,
		facebook_id=NULL,google_id=NULL,gamecenter_id=NULL,steam_id=NULL,custom_id=NULL,facebook_instant_game_id=NULL,apple_id=NULL,
		location=NULL,timezone=NULL,metadata='{}',disable_time=now(),update_time=now() WHERE id=$1`,
		in.UserID, "purged-"+strings.ReplaceAll(in.UserID, "-", ""), anon); err != nil {
		logger.Error("Jazebe purge identity failed: %v", err)
		return "", status.Error(codes.Internal, "purge failed")
	}
	erased = append(erased, "identity")
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_user_profile SET display_name=$2,avatar_kind='preset',avatar_preset_index=0,
		avatar_media_id=NULL,updated_at=now() WHERE user_id=$1`, in.UserID, anon); err != nil {
		logger.Error("Jazebe purge profile failed: %v", err)
		return "", status.Error(codes.Internal, "purge failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_user_account(user_id,state) VALUES($1,$2)
		ON CONFLICT (user_id) DO UPDATE SET purged_at=now(),purged_by_operator_id=$3,session_revocation_pending=true,updated_at=now()`,
		in.UserID, state, actor); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.ForeignKeyViolation {
			return "", status.Error(codes.NotFound, "user not found")
		}
		return "", status.Error(codes.Internal, "purge failed")
	}
	params, _ := json.Marshal(map[string]any{"reason": in.Reason, "purged": erased})
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,params)
		VALUES('account.purge','operator',$1,'user',$2,$3::jsonb)`, actor, in.UserID, string(params)); err != nil {
		return "", status.Error(codes.Internal, "purge audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "purge failed")
	}
	// Sessions are revoked after commit; a failure stays pending for the
	// account revocation worker, as for suspensions.
	if err := revokeAccountSessions(ctx, nk, in.UserID); err == nil {
		if _, err := db.ExecContext(ctx, `UPDATE jazebe_user_account SET session_revocation_pending=false WHERE user_id=$1`, in.UserID); err == nil {
			erased = append(erased, "sessions")
		}
	} else {
		logger.Warn("Jazebe purge session revocation pending for %s: %v", in.UserID, err)
	}
	return encode(map[string]any{"userId": in.UserID, "purged": erased})
}

// refuseJazebeAccountDeletion runs before Nakama's client account deletion.
// Jazebe keeps ledgers, memberships and audits for a deleted person (source
// purge never removes them), so their foreign keys make a physical delete of
// such a user fail. The deletion is tried in a transaction that is always
// rolled back; when Jazebe rows would block it, the client gets a clear
// FailedPrecondition instead of an opaque internal error.
func refuseJazebeAccountDeletion(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule) error {
	user, err := userID(ctx)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return runtime.NewError("account deletion unavailable", int(codes.Internal))
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, user); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.ForeignKeyViolation {
			return runtime.NewError("account holds Jazebe planet data; deletion is handled by an operator purge request", int(codes.FailedPrecondition))
		}
		logger.Error("Jazebe account deletion probe failed: %v", err)
		return runtime.NewError("account deletion unavailable", int(codes.Internal))
	}
	return nil
}
