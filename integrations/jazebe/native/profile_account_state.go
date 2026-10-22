package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type operatorAccountAction struct {
	UserID string `json:"userId"`
	Reason string `json:"reason"`
}

func validateOperatorAccountAction(in operatorAccountAction) (operatorAccountAction, error) {
	var err error
	in.UserID, err = requiredUUID(in.UserID)
	if err != nil {
		return in, err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 3 || utf8.RuneCountInString(in.Reason) > 512 {
		return in, status.Error(codes.InvalidArgument, "reason must contain 3 to 512 characters")
	}
	return in, nil
}

// The database transition and audit record commit together. Revocation is
// deliberately recorded as pending before commit so a failed local cache or
// socket operation cannot masquerade as a fully completed restriction.
func operatorAccountTransition(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload, action string) (string, error) {
	actor, err := operatorReportActor(ctx)
	if err != nil {
		return "", err
	}
	var in operatorAccountAction
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in, err = validateOperatorAccountAction(in)
	if err != nil {
		return "", err
	}
	if actor == in.UserID {
		return "", status.Error(codes.PermissionDenied, "operator cannot change own account state")
	}
	if isOperator(in.UserID, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) &&
		!isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "super operator target requires super operator")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var disabledAt time.Time
	err = tx.QueryRowContext(ctx, `SELECT disable_time FROM users WHERE id=$1 FOR UPDATE`, in.UserID).Scan(&disabledAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "account not found")
	}
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_account(user_id) VALUES($1) ON CONFLICT(user_id) DO NOTHING`, in.UserID); err != nil {
		return "", err
	}
	var state string
	var ownedAt sql.NullTime
	var revocationPending bool
	if err = tx.QueryRowContext(ctx, `SELECT state,core_disable_time,session_revocation_pending FROM jazebe_user_account WHERE user_id=$1 FOR UPDATE`, in.UserID).Scan(&state, &ownedAt, &revocationPending); err != nil {
		return "", err
	}
	previous := state
	switch action {
	case "suspend":
		if state == "suspended" {
			return "", status.Error(codes.FailedPrecondition, "account already suspended")
		}
		// An existing core ban has unknown provenance. Never overwrite it and
		// later claim authority to remove it.
		if disabledAt.Unix() != 0 {
			return "", status.Error(codes.FailedPrecondition, "account already restricted by Nakama")
		}
		if err = tx.QueryRowContext(ctx, `UPDATE users SET disable_time=now(),update_time=now() WHERE id=$1 RETURNING disable_time`, in.UserID).Scan(&disabledAt); err != nil {
			return "", err
		}
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_account SET state='suspended',suspended_by_operator_id=$2,
			suspended_reason=$3,suspended_at=now(),core_disable_time=$4,session_revocation_pending=true,updated_at=now()
			WHERE user_id=$1`, in.UserID, actor, in.Reason, disabledAt)
	case "unsuspend", "reactivate":
		required := "suspended"
		if action == "reactivate" {
			required = "deactivated"
		}
		if state != required {
			return "", status.Error(codes.FailedPrecondition, "account is not "+required)
		}
		if revocationPending {
			return "", status.Error(codes.FailedPrecondition, "session revocation is pending")
		}
		// A ban placed by Nakama after Jazebe's restriction changes the
		// timestamp. Refuse restoration rather than clearing that later ban.
		if !ownedAt.Valid || !disabledAt.Equal(ownedAt.Time) {
			return "", status.Error(codes.FailedPrecondition, "core ban ownership cannot be established")
		}
		result, e := tx.ExecContext(ctx, `UPDATE users SET disable_time='1970-01-01 00:00:00 UTC',update_time=now()
			WHERE id=$1 AND disable_time=$2`, in.UserID, ownedAt.Time)
		if e != nil {
			return "", e
		}
		if n, e := result.RowsAffected(); e != nil || n != 1 {
			return "", status.Error(codes.Aborted, "core ban changed concurrently")
		}
		if action == "unsuspend" {
			_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_account SET state='active',suspended_by_operator_id=NULL,
				suspended_reason=NULL,suspended_at=NULL,reactivated_at=now(),core_disable_time=NULL,updated_at=now()
				WHERE user_id=$1`, in.UserID)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_account SET state='active',deactivated_at=NULL,
				reactivated_at=now(),core_disable_time=NULL,updated_at=now() WHERE user_id=$1`, in.UserID)
		}
	default:
		return "", status.Error(codes.InvalidArgument, "invalid account action")
	}
	if err != nil {
		return "", err
	}
	var auditID int64
	if err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,params)
		VALUES($1,'operator',$2,'user',$3,jsonb_build_object('reason',$4,'previousState',$5,'newState',$6)) RETURNING id`,
		"account."+action, actor, in.UserID, in.Reason, previous, map[string]string{"suspend": "suspended", "unsuspend": "active", "reactivate": "active"}[action]).Scan(&auditID); err != nil {
		return "", err
	}
	if action == "suspend" || action == "unsuspend" {
		title, body := "تعلیق حساب برداشته شد", "حساب کاربری تو دوباره فعال شد و می‌توانی وارد شوی."
		if action == "suspend" {
			title, body = "حساب تو تعلیق شد", "حساب کاربری تو توسط پشتیبانی جاذبه تعلیق شد. دلیل: "+in.Reason
		}
		if err = persistJazebeNotification(ctx, tx, in.UserID, "account-state:"+strconv.FormatInt(auditID, 10),
			title, map[string]any{"title": title, "body": body, "category": "critical", "priority": "urgent",
				"subjectType": "account", "subjectId": in.UserID}, time.Now().UTC()); err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	if action == "suspend" {
		if err = retryAccountSessionRevocation(ctx, db, nk, in.UserID); err != nil {
			logger.Error("Jazebe account session revocation pending for %s: %v", in.UserID, err)
			return "", status.Error(codes.Unavailable, "account restricted; session revocation pending")
		}
	}
	return encode(map[string]any{"userId": in.UserID, "state": map[string]string{"suspend": "suspended", "unsuspend": "active", "reactivate": "active"}[action]})
}

func retryAccountSessionRevocation(ctx context.Context, db *sql.DB, nk runtime.NakamaModule, user string) error {
	var pending bool
	err := db.QueryRowContext(ctx, `SELECT session_revocation_pending FROM jazebe_user_account
		WHERE user_id=$1 AND state='suspended'`, user).Scan(&pending)
	if errors.Is(err, sql.ErrNoRows) {
		return status.Error(codes.FailedPrecondition, "account is not suspended")
	}
	if err != nil {
		return err
	}
	if !pending {
		return nil
	}
	if err = revokeAccountSessions(ctx, nk, user); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE jazebe_user_account SET session_revocation_pending=false,updated_at=now()
		WHERE user_id=$1 AND state='suspended' AND session_revocation_pending`, user)
	return err
}

func operatorAccountRevocationRetry(ctx context.Context, _ runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := operatorReportActor(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		UserID string `json:"userId"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.UserID)
	if err != nil {
		return "", err
	}
	if isOperator(id, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) &&
		!isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "super operator target requires super operator")
	}
	if err = retryAccountSessionRevocation(ctx, db, nk, id); err != nil {
		return "", status.Error(codes.Unavailable, "session revocation failed")
	}
	return encode(map[string]any{"userId": id, "revocationPending": false})
}

func startAccountRevocationWorker(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) error {
	enabled := os.Getenv("JAZEBE_ACCOUNT_REVOCATION_WORKER_ENABLED")
	if enabled == "" || enabled == "false" {
		return nil
	}
	if enabled != "true" {
		return errors.New("JAZEBE_ACCOUNT_REVOCATION_WORKER_ENABLED must be true or false")
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			workerCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			rows, err := db.QueryContext(workerCtx, `SELECT user_id FROM jazebe_user_account
				WHERE state='suspended' AND session_revocation_pending ORDER BY suspended_at LIMIT 100`)
			if err != nil {
				logger.Error("Jazebe account revocation scan failed: %v", err)
			} else {
				users := make([]string, 0, 100)
				for rows.Next() {
					var user string
					if err = rows.Scan(&user); err != nil {
						break
					}
					users = append(users, user)
				}
				if err == nil {
					err = rows.Err()
				}
				rows.Close()
				if err != nil {
					logger.Error("Jazebe account revocation scan failed: %v", err)
				}
				for _, user := range users {
					if err = retryAccountSessionRevocation(workerCtx, db, nk, user); err != nil {
						logger.Error("Jazebe account revocation retry failed for %s: %v", user, err)
					}
				}
			}
			cancel()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

func revokeAccountSessions(ctx context.Context, nk runtime.NakamaModule, user string) error {
	if err := nk.SessionLogout(user, "", ""); err != nil {
		return err
	}
	// Notifications are present for every Nakama realtime connection. Keep
	// the durable pending marker if any disconnect fails.
	presences, err := nk.StreamUserList(0, user, "", "", true, true)
	if err != nil {
		return err
	}
	for _, presence := range presences {
		if err = nk.SessionDisconnect(ctx, presence.GetSessionId()); err != nil {
			return err
		}
	}
	return nil
}

func operatorAccountSuspend(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return operatorAccountTransition(ctx, logger, db, nk, payload, "suspend")
}

func operatorAccountUnsuspend(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return operatorAccountTransition(ctx, logger, db, nk, payload, "unsuspend")
}

func operatorAccountReactivate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return operatorAccountTransition(ctx, logger, db, nk, payload, "reactivate")
}
