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

func groupLeadershipOffer(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	var id, from, to string
	var expires time.Time
	err = db.QueryRowContext(ctx, `SELECT o.id,o.from_user_id,o.to_user_id,o.expires_at
		FROM jazebe_group_leadership_offer o JOIN jazebe_group g ON g.id=o.group_id
		JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		LEFT JOIN jazebe_planet_membership pm ON pm.planet_id=p.id AND pm.user_id=$3
		WHERE o.group_id=$1 AND o.planet_id=$2 AND o.status='pending' AND o.expires_at>now()
		AND (o.from_user_id=$3 OR o.to_user_id=$3) AND p.status='active' AND b.is_active
		AND COALESCE(pm.is_active,true)`, ref.GroupID, ref.PlanetID, actor).Scan(&id, &from, &to, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return encode(map[string]any{"offer": nil})
	}
	if err != nil {
		return "", status.Error(codes.Internal, "leadership offer unavailable")
	}
	direction, other := "outgoing", to
	if to == actor {
		direction, other = "incoming", from
	}
	return encode(map[string]any{"offer": map[string]any{"id": id, "direction": direction, "otherUserId": other, "expiresAt": expires}})
}

func groupRespondLeadership(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		Accept   bool   `json:"accept"`
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "leadership response unavailable")
	}
	defer tx.Rollback()
	var groupStatus, creator, name string
	err = tx.QueryRowContext(ctx, `SELECT g.status,g.creator_user_id,g.name FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id
		JOIN jazebe_brand b ON b.id=p.brand_id WHERE g.id=$1 AND g.planet_id=$2 AND p.status='active' AND b.is_active FOR UPDATE OF g`, input.GroupID, input.PlanetID).Scan(&groupStatus, &creator, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	if err := requireUsablePlanetMembership(ctx, tx, input.PlanetID, actor); err != nil {
		return "", err
	}
	var offerID, from string
	var expires time.Time
	err = tx.QueryRowContext(ctx, `SELECT id,from_user_id,expires_at FROM jazebe_group_leadership_offer
		WHERE group_id=$1 AND to_user_id=$2 AND status='pending' FOR UPDATE`, input.GroupID, actor).Scan(&offerID, &from, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "leadership offer not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "leadership offer unavailable")
	}
	if !expires.After(time.Now().UTC()) {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='expired',decided_at=now() WHERE id=$1`, offerID); err != nil {
			return "", status.Error(codes.Internal, "offer expiry failed")
		}
		if err = tx.Commit(); err != nil {
			return "", status.Error(codes.Internal, "offer expiry failed")
		}
		return "", status.Error(codes.NotFound, "leadership offer expired")
	}
	if !input.Accept {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='declined',decided_at=now() WHERE id=$1`, offerID); err != nil {
			return "", status.Error(codes.Internal, "offer decline failed")
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
			VALUES('group.transfer.decline','user',$1,'group',$2,$3,jsonb_build_object('offerId',$4))`, actor, input.GroupID, input.PlanetID, offerID); err != nil {
			return "", status.Error(codes.Internal, "offer audit failed")
		}
		if err = persistJazebeNotification(ctx, tx, from, "group-offer-declined:"+offerID, "پیشنهاد رهبری پذیرفته نشد",
			map[string]any{"category": "group", "actionLink": "/group", "groupId": input.GroupID, "offerId": offerID}, time.Now().UTC()); err != nil {
			return "", status.Error(codes.Internal, "offer notice failed")
		}
		if err = tx.Commit(); err != nil {
			return "", status.Error(codes.Internal, "offer decline failed")
		}
		return encode(map[string]any{"status": "declined"})
	}
	if groupStatus != "active" {
		return "", status.Error(codes.FailedPrecondition, "group inactive")
	}
	var fromRole, toRole string
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active' FOR UPDATE`, input.GroupID, from).Scan(&fromRole)
	if errors.Is(err, sql.ErrNoRows) {
		return groupCancelStaleOffer(ctx, tx, offerID)
	}
	if err != nil {
		return "", status.Error(codes.Internal, "creator unavailable")
	}
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active' FOR UPDATE`, input.GroupID, actor).Scan(&toRole)
	if errors.Is(err, sql.ErrNoRows) {
		return groupCancelStaleOffer(ctx, tx, offerID)
	}
	if err != nil {
		return "", status.Error(codes.Internal, "administrator unavailable")
	}
	if creator != from || fromRole != "creator" || toRole != "admin" {
		return groupCancelStaleOffer(ctx, tx, offerID)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_membership SET role='member' WHERE group_id=$1 AND user_id=$2 AND status='active'`, input.GroupID, from); err != nil {
		return "", status.Error(codes.Internal, "creator transfer failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_membership SET role='creator' WHERE group_id=$1 AND user_id=$2 AND status='active'`, input.GroupID, actor); err != nil {
		return "", status.Error(codes.Internal, "creator transfer failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group SET creator_user_id=$2,creator_transferred_at=now(),updated_at=now() WHERE id=$1`, input.GroupID, actor); err != nil {
		return "", status.Error(codes.Internal, "creator transfer failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='accepted',decided_at=now() WHERE id=$1`, offerID); err != nil {
		return "", status.Error(codes.Internal, "offer acceptance failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('group.transfer','user',$1,'group',$2,$3,jsonb_build_object('from',$4,'to',$1,'offerId',$5))`, actor, input.GroupID, input.PlanetID, from, offerID); err != nil {
		return "", status.Error(codes.Internal, "transfer audit failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
		VALUES($1,$2,$3,'role',$4,'member.transfer',jsonb_build_object('userId',$4,'fromUserId',$5,'groupName',$6))`, uuid.Must(uuid.NewV4()).String(), input.PlanetID, input.GroupID, actor, from, name); err != nil {
		return "", status.Error(codes.Internal, "transfer activity failed")
	}
	if err = persistJazebeNotification(ctx, tx, from, "group-offer-accepted:"+offerID, "رهبری گروه منتقل شد",
		map[string]any{"category": "group", "actionLink": "/group", "groupId": input.GroupID, "offerId": offerID}, time.Now().UTC()); err != nil {
		return "", status.Error(codes.Internal, "transfer notice failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "creator transfer failed")
	}
	return encode(map[string]any{"status": "accepted", "creatorUserId": actor})
}

func groupCancelStaleOffer(ctx context.Context, tx *sql.Tx, id string) (string, error) {
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='cancelled',decided_at=now() WHERE id=$1`, id); err != nil {
		return "", status.Error(codes.Internal, "offer cancellation failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "offer cancellation failed")
	}
	return "", status.Error(codes.NotFound, "leadership offer no longer valid")
}

func groupWithdrawLeadership(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	result, err := db.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='cancelled',decided_at=now()
		WHERE group_id=$1 AND planet_id=$2 AND from_user_id=$3 AND status='pending'
		AND EXISTS(SELECT 1 FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id
			JOIN jazebe_brand b ON b.id=p.brand_id LEFT JOIN jazebe_planet_membership pm
			ON pm.planet_id=p.id AND pm.user_id=$3 WHERE g.id=$1 AND p.status='active' AND b.is_active
			AND COALESCE(pm.is_active,true))`, ref.GroupID, ref.PlanetID, actor)
	if err != nil {
		return "", status.Error(codes.Internal, "offer withdrawal failed")
	}
	count, err := result.RowsAffected()
	if err != nil {
		return "", status.Error(codes.Internal, "offer withdrawal failed")
	}
	if count == 0 {
		return "", status.Error(codes.NotFound, "leadership offer not found")
	}
	return encode(map[string]any{"status": "cancelled"})
}
