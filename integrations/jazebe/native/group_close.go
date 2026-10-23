package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func groupCanRemove(actorRole, targetRole string, self bool) error {
	if self || targetRole == "creator" {
		return status.Error(codes.PermissionDenied, "cannot remove group creator or self")
	}
	if actorRole != "creator" && (actorRole != "admin" || targetRole != "member") {
		return status.Error(codes.PermissionDenied, "group manager cannot remove this member")
	}
	return nil
}

func groupRemoveMember(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		MemberID string `json:"memberId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	input.GroupID, err = requiredUUID(input.GroupID)
	if err != nil {
		return "", err
	}
	input.MemberID, err = requiredUUID(input.MemberID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group removal unavailable")
	}
	defer tx.Rollback()
	var groupStatus, name string
	err = tx.QueryRowContext(ctx, `SELECT g.status,g.name FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id
		JOIN jazebe_brand b ON b.id=p.brand_id WHERE g.id=$1 AND g.planet_id=$2 AND p.status='active' AND b.is_active FOR UPDATE OF g`, input.GroupID, input.PlanetID).Scan(&groupStatus, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	if groupStatus != "active" {
		return "", status.Error(codes.FailedPrecondition, "group inactive")
	}
	if err := requireUsablePlanetMembership(ctx, tx, input.PlanetID, actor); err != nil {
		return "", err
	}
	var actorRole, targetRole string
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active'`, input.GroupID, actor).Scan(&actorRole)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "group manager required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group manager unavailable")
	}
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active' FOR UPDATE`, input.GroupID, input.MemberID).Scan(&targetRole)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "member not in group")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group member unavailable")
	}
	if err := groupCanRemove(actorRole, targetRole, actor == input.MemberID); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_membership SET status='removed',left_at=now()
		WHERE group_id=$1 AND user_id=$2 AND status='active'`, input.GroupID, input.MemberID); err != nil {
		return "", status.Error(codes.Internal, "member removal failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group SET member_count=member_count-1,updated_at=now() WHERE id=$1`, input.GroupID); err != nil {
		return "", status.Error(codes.Internal, "group count failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='cancelled',decided_at=now()
		WHERE group_id=$1 AND status='pending' AND (from_user_id=$2 OR to_user_id=$2)`, input.GroupID, input.MemberID); err != nil {
		return "", status.Error(codes.Internal, "offer cleanup failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_invite SET revoked_at=now() WHERE group_id=$1 AND created_by_user_id=$2 AND revoked_at IS NULL`, input.GroupID, input.MemberID); err != nil {
		return "", status.Error(codes.Internal, "invite cleanup failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
		VALUES($1,$2,$3,'leave',$4,'member.removed',jsonb_build_object('userId',$5::text))`, uuid.Must(uuid.NewV4()).String(), input.PlanetID, input.GroupID, actor, input.MemberID); err != nil {
		return "", status.Error(codes.Internal, "group activity failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('group.member.remove','user',$1,'group',$2,$3,jsonb_build_object('userId',$4::text))`, actor, input.GroupID, input.PlanetID, input.MemberID); err != nil {
		return "", status.Error(codes.Internal, "removal audit failed")
	}
	if err = persistJazebeNotification(ctx, tx, input.MemberID, "group-remove:"+uuid.Must(uuid.NewV4()).String(), "از گروه حذف شدی",
		map[string]any{"category": "group", "actionLink": "/groups", "groupId": input.GroupID, "groupName": name}, time.Now().UTC()); err != nil {
		return "", status.Error(codes.Internal, "removal notice failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "member removal failed")
	}
	if err := groupChatKickUser(nk, groupRef{PlanetID: input.PlanetID, GroupID: input.GroupID}, input.MemberID); err != nil {
		logger.Warn("Jazebe group stream eviction failed: %v", err)
	}
	return encode(map[string]any{"removed": true, "userId": input.MemberID})
}

func groupDisband(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group disband unavailable")
	}
	defer tx.Rollback()
	var creator, groupStatus, name string
	err = tx.QueryRowContext(ctx, `SELECT g.creator_user_id,g.status,g.name FROM jazebe_group g
		JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE g.id=$1 AND g.planet_id=$2 AND p.status='active' AND b.is_active FOR UPDATE OF g`, ref.GroupID, ref.PlanetID).Scan(&creator, &groupStatus, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	if creator != actor {
		return "", status.Error(codes.PermissionDenied, "group creator required")
	}
	if err := requireUsablePlanetMembership(ctx, tx, ref.PlanetID, actor); err != nil {
		return "", err
	}
	if groupStatus == "disbanded" {
		return "", status.Error(codes.FailedPrecondition, "group already disbanded")
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_group_membership WHERE group_id=$1 AND status='active'`, ref.GroupID)
	if err != nil {
		return "", status.Error(codes.Internal, "group members unavailable")
	}
	members := make([]string, 0)
	for rows.Next() {
		var user string
		if rows.Scan(&user) != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "group members unavailable")
		}
		members = append(members, user)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return "", status.Error(codes.Internal, "group members unavailable")
	}
	rows.Close()
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_membership SET status='left',left_at=now() WHERE group_id=$1 AND status='active'`, ref.GroupID); err != nil {
		return "", status.Error(codes.Internal, "group disband failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_join_request SET status='auto_cancelled',decided_at=now() WHERE group_id=$1 AND status='pending'`, ref.GroupID); err != nil {
		return "", status.Error(codes.Internal, "group request cleanup failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='cancelled',decided_at=now() WHERE group_id=$1 AND status='pending'`, ref.GroupID); err != nil {
		return "", status.Error(codes.Internal, "offer cleanup failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_invite SET revoked_at=now() WHERE group_id=$1 AND revoked_at IS NULL`, ref.GroupID); err != nil {
		return "", status.Error(codes.Internal, "invite cleanup failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group SET status='disbanded',member_count=0,disbanded_at=now(),disbanded_by_user_id=$2,updated_at=now() WHERE id=$1`, ref.GroupID, actor); err != nil {
		return "", status.Error(codes.Internal, "group disband failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('group.disband','user',$1,'group',$2,$3,jsonb_build_object('members',$4::bigint))`, actor, ref.GroupID, ref.PlanetID, len(members)); err != nil {
		return "", status.Error(codes.Internal, "disband audit failed")
	}
	eventID := uuid.Must(uuid.NewV4()).String()
	for _, user := range members {
		if user != actor {
			if err = persistJazebeNotification(ctx, tx, user, "group-disband:"+eventID+":"+user, "گروه منحل شد",
				map[string]any{"category": "group", "actionLink": "/groups", "groupId": ref.GroupID, "groupName": name}, time.Now().UTC()); err != nil {
				return "", status.Error(codes.Internal, "disband notice failed")
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group disband failed")
	}
	if err := nk.StreamClose(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel); err != nil {
		logger.Warn("Jazebe group stream closure failed: %v", err)
	}
	return encode(map[string]any{"id": ref.GroupID, "status": "disbanded", "members": 0})
}
