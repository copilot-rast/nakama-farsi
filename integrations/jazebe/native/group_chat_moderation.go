package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func chatDeleteAllowed(role string, own, asAdmin bool, created, now time.Time) (bool, error) {
	moderation := !own || (asAdmin && (role == "creator" || role == "admin"))
	if moderation && role != "creator" && role != "admin" {
		return false, status.Error(codes.PermissionDenied, "group manager required to delete another member's message")
	}
	if !moderation && now.Sub(created) > 15*time.Minute {
		return false, status.Error(codes.PermissionDenied, "self-delete window expired")
	}
	return moderation, nil
}

func groupChatDelete(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		Seq      int64  `json:"seq"`
		AsAdmin  bool   `json:"asAdmin"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref, err := readGroupRef(jsonString(groupRef{PlanetID: input.PlanetID, GroupID: input.GroupID}))
	if err != nil {
		return "", err
	}
	if input.Seq < 1 {
		return "", status.Error(codes.InvalidArgument, "invalid message sequence")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "chat unavailable")
	}
	defer tx.Rollback()
	role, err := chatMember(ctx, tx, ref, actor, true)
	if err != nil {
		return "", err
	}
	var id, kind string
	var sender sql.NullString
	var created time.Time
	var deleted sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT id,kind,sender_user_id,created_at,deleted_at FROM jazebe_group_message
		WHERE group_id=$1 AND seq=$2 FOR UPDATE`, ref.GroupID, input.Seq).Scan(&id, &kind, &sender, &created, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "message not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "message unavailable")
	}
	if kind != "text" {
		return "", status.Error(codes.PermissionDenied, "only text messages can be deleted")
	}
	if deleted.Valid {
		return "", status.Error(codes.AlreadyExists, "message already deleted")
	}
	moderation, err := chatDeleteAllowed(role, sender.Valid && sender.String == actor, input.AsAdmin, created, time.Now().UTC())
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_message SET deleted_at=now(),deleted_by=$2,deleted_as_admin=$3 WHERE id=$1`, id, actor, moderation); err != nil {
		return "", status.Error(codes.Internal, "message delete failed")
	}
	if moderation {
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
			VALUES('group.message.delete','user',$1,'group_message',$2,$3,jsonb_build_object('groupId',$4,'seq',$5,'senderUserId',$6))`, actor, id, ref.PlanetID, ref.GroupID, input.Seq, sender); err != nil {
			return "", status.Error(codes.Internal, "message moderation audit failed")
		}
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "message delete failed")
	}
	if err := groupChatBroadcast(ctx, db, nk, ref, "message.deleted", map[string]any{"seq": input.Seq, "asAdmin": moderation, "byUserId": actor}); err != nil {
		logger.Warn("Jazebe group chat realtime deletion failed: %v", err)
	}
	return encode(map[string]any{"seq": input.Seq, "deleted": true})
}

func groupChatReport(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		Seq      int64  `json:"seq"`
		Reason   string `json:"reason"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref, err := readGroupRef(jsonString(groupRef{PlanetID: input.PlanetID, GroupID: input.GroupID}))
	if err != nil {
		return "", err
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Seq < 1 || !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) < 1 || utf8.RuneCountInString(input.Reason) > 512 {
		return "", status.Error(codes.InvalidArgument, "invalid message report")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "chat unavailable")
	}
	defer tx.Rollback()
	if _, err := chatMember(ctx, tx, ref, actor, false); err != nil {
		return "", err
	}
	var id, kind string
	var sender sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id,kind,sender_user_id FROM jazebe_group_message WHERE group_id=$1 AND seq=$2`, ref.GroupID, input.Seq).Scan(&id, &kind, &sender)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "message not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "message unavailable")
	}
	if sender.Valid && sender.String == actor {
		return "", status.Error(codes.InvalidArgument, "cannot report own message")
	}
	if kind != "text" {
		return "", status.Error(codes.InvalidArgument, "only text messages can be reported")
	}
	runes := []rune(input.Reason)
	reason := string(runes)
	var note any
	if len(runes) > 64 {
		reason = string(runes[:64])
		note = input.Reason
	}
	reportID := uuid.Must(uuid.NewV4()).String()
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_user_report(id,planet_id,reporter_user_id,subject_type,subject_id,reason,note)
		VALUES($1,$2,$3,'message',$4,$5,$6)
		ON CONFLICT (planet_id,reporter_user_id,subject_type,subject_id)
			WHERE status IN ('open','reviewing') AND subject_type IN ('message','group') DO NOTHING
		RETURNING id`, reportID, ref.PlanetID, actor, id, reason, note).Scan(&reportID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT id FROM jazebe_user_report WHERE reporter_user_id=$1 AND planet_id=$2 AND subject_type='message'
			AND subject_id=$3 AND status IN ('open','reviewing')`, actor, ref.PlanetID, id).Scan(&reportID)
	}
	if err != nil {
		return "", status.Error(codes.Internal, "message report failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "message report failed")
	}
	return encode(map[string]any{"id": reportID})
}

func groupChatModerate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		Seq      int64  `json:"seq"`
		Reason   string `json:"reason"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref, err := readGroupRef(jsonString(groupRef{PlanetID: input.PlanetID, GroupID: input.GroupID}))
	if err != nil {
		return "", err
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Seq < 1 || !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) < 1 || utf8.RuneCountInString(input.Reason) > 512 {
		return "", status.Error(codes.InvalidArgument, "moderation reason required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "chat moderation unavailable")
	}
	defer tx.Rollback()
	var brandID, groupStatus string
	err = tx.QueryRowContext(ctx, `SELECT p.brand_id,g.status FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id
		WHERE g.id=$1 AND g.planet_id=$2 FOR SHARE OF g`, ref.GroupID, ref.PlanetID).Scan(&brandID, &groupStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	_ = groupStatus // Staff can moderate retained chat after a group is disbanded.
	actorType := "brand_admin"
	if isOperator(actor, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		actorType = "operator"
	} else {
		var allowed bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_brand_admin a JOIN jazebe_brand b ON b.id=a.brand_id
			WHERE a.brand_id=$1 AND a.user_id=$2 AND a.is_active AND b.is_active AND a.role IN ('main','marketing'))`, brandID, actor).Scan(&allowed)
		if err != nil {
			return "", status.Error(codes.Internal, "moderation authorization unavailable")
		}
		if !allowed {
			return "", status.Error(codes.PermissionDenied, "brand moderation role required")
		}
	}
	var id, kind string
	var sender sql.NullString
	var deleted sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT id,kind,sender_user_id,deleted_at FROM jazebe_group_message
		WHERE group_id=$1 AND seq=$2 FOR UPDATE`, ref.GroupID, input.Seq).Scan(&id, &kind, &sender, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "message not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "message unavailable")
	}
	if kind != "text" {
		return "", status.Error(codes.PermissionDenied, "only text messages can be moderated")
	}
	if deleted.Valid {
		return "", status.Error(codes.AlreadyExists, "message already deleted")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_message SET deleted_at=now(),deleted_by=$2,deleted_as_admin=true WHERE id=$1`, id, actor); err != nil {
		return "", status.Error(codes.Internal, "message moderation failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_report SET status='actioned',handled_by_admin_id=$2,handled_at=now()
		WHERE subject_type='message' AND subject_id=$1 AND status IN ('open','reviewing')`, id, actor); err != nil {
		return "", status.Error(codes.Internal, "report moderation failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('group.message.delete',$1,$2,'group_message',$3,$4,jsonb_build_object('groupId',$5,'seq',$6,'reason',$7,'senderUserId',$8))`, actorType, actor, id, ref.PlanetID, ref.GroupID, input.Seq, input.Reason, sender); err != nil {
		return "", status.Error(codes.Internal, "moderation audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "message moderation failed")
	}
	if err := groupChatBroadcast(ctx, db, nk, ref, "message.deleted", map[string]any{"seq": input.Seq, "asAdmin": true, "byUserId": actor}); err != nil {
		logger.Warn("Jazebe group chat moderation delivery failed: %v", err)
	}
	return encode(map[string]any{"seq": input.Seq, "deleted": true})
}
