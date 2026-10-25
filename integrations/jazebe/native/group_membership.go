package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type groupRef struct {
	PlanetID string `json:"planetId"`
	GroupID  string `json:"groupId"`
}

func readGroupRef(payload string) (groupRef, error) {
	var ref groupRef
	if err := decode(payload, &ref); err != nil {
		return ref, err
	}
	var err error
	ref.PlanetID, err = requiredUUID(ref.PlanetID)
	if err != nil {
		return ref, err
	}
	ref.GroupID, err = requiredUUID(ref.GroupID)
	return ref, err
}

// The user lock serializes simultaneous joins and request replacements across
// different groups. The row lock protects a group's capacity.
func groupJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group join unavailable")
	}
	defer tx.Rollback()
	if err := lockXact(ctx, tx, "774103:"+ref.PlanetID+user); err != nil {
		return "", status.Error(codes.Internal, "group join unavailable")
	}
	var mode, groupStatus, planetStatus, name string
	var brandActive bool
	var count, cap int
	err = tx.QueryRowContext(ctx, `SELECT g.mode,g.status,p.status,b.is_active,g.name,g.member_count,g.member_cap
		FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE g.id=$1 AND g.planet_id=$2 FOR UPDATE OF g`, ref.GroupID, ref.PlanetID).
		Scan(&mode, &groupStatus, &planetStatus, &brandActive, &name, &count, &cap)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group join unavailable")
	}
	if groupStatus != "active" || planetStatus != "active" || !brandActive {
		return "", status.Error(codes.FailedPrecondition, "group unavailable")
	}
	if err := requireUsablePlanetMembership(ctx, tx, ref.PlanetID, user); err != nil {
		return "", err
	}
	var displayName sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT nullif(trim(display_name),'') FROM users WHERE id=$1`, user).Scan(&displayName); err != nil {
		return "", status.Error(codes.Internal, "profile unavailable")
	}
	if !displayName.Valid {
		return "", status.Error(codes.FailedPrecondition, "display name required")
	}
	var current string
	err = tx.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`,
		ref.PlanetID, user).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "group membership unavailable")
	}
	if current == ref.GroupID {
		return encode(map[string]any{"joined": true, "requested": false})
	}
	if current != "" {
		return "", status.Error(codes.AlreadyExists, "already in another group")
	}
	if count >= cap {
		return "", status.Error(codes.ResourceExhausted, "group full")
	}
	var pending string
	err = tx.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_join_request WHERE planet_id=$1 AND user_id=$2 AND status='pending'`,
		ref.PlanetID, user).Scan(&pending)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "group requests unavailable")
	}
	if mode == "approval" && pending == ref.GroupID {
		return encode(map[string]any{"joined": false, "requested": true})
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_join_request SET status='auto_cancelled',decided_at=now()
		WHERE planet_id=$1 AND user_id=$2 AND status='pending'`, ref.PlanetID, user); err != nil {
		return "", status.Error(codes.Internal, "group requests unavailable")
	}
	if mode == "approval" {
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_join_request(id,planet_id,group_id,user_id)
			VALUES($1,$2,$3,$4)`, uuid.Must(uuid.NewV4()).String(), ref.PlanetID, ref.GroupID, user)
		if err != nil {
			return "", status.Error(codes.Internal, "group request failed")
		}
		if err := tx.Commit(); err != nil {
			return "", status.Error(codes.Internal, "group request failed")
		}
		return encode(map[string]any{"joined": false, "requested": true})
	}
	membershipID := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_membership(id,group_id,planet_id,user_id,role,source)
		VALUES($1,$2,$3,$4,'member','list_join')`, membershipID, ref.GroupID, ref.PlanetID, user)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "already in group")
		}
		return "", status.Error(codes.Internal, "group join failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group SET member_count=member_count+1,updated_at=now() WHERE id=$1`, ref.GroupID); err != nil {
		return "", status.Error(codes.Internal, "group count update failed")
	}
	var first bool
	if err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership(planet_id,user_id,join_path,last_activity_at)
		VALUES($1,$2,'first_activity',now()) ON CONFLICT(planet_id,user_id) DO NOTHING RETURNING true`,
		ref.PlanetID, user).Scan(&first); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "planet membership creation failed")
	}
	if first {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=fans_count+1 WHERE id=$1`, ref.PlanetID); err != nil {
			return "", status.Error(codes.Internal, "fan count update failed")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=now() WHERE planet_id=$1 AND user_id=$2`, ref.PlanetID, user); err != nil {
		return "", status.Error(codes.Internal, "planet activity update failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
		VALUES($1,$2,$3,'join',$4,'member.joined',jsonb_build_object('userId',$4::uuid,'name',$5::text,'source','list_join','text',$6::text))`,
		uuid.Must(uuid.NewV4()).String(), ref.PlanetID, ref.GroupID, user, displayName.String, displayName.String+" به گروه پیوست"); err != nil {
		return "", status.Error(codes.Internal, "group activity failed")
	}
	if err := enqueueActivityXP(ctx, tx, ref.PlanetID, user, membershipID, "group_join"); err != nil {
		return "", status.Error(codes.Internal, "group progression event failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group join failed")
	}
	groupJoinProgression(ctx, logger, db, ref.PlanetID, user, membershipID)
	return encode(map[string]any{"joined": true, "requested": false})
}

func groupMembershipState(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	var joined, requested bool
	err = db.QueryRowContext(ctx, `SELECT
		EXISTS(SELECT 1 FROM jazebe_group_membership WHERE group_id=$1 AND planet_id=$2 AND user_id=$3 AND status='active'),
		EXISTS(SELECT 1 FROM jazebe_group_join_request WHERE group_id=$1 AND planet_id=$2 AND user_id=$3 AND status='pending')`,
		ref.GroupID, ref.PlanetID, user).Scan(&joined, &requested)
	if err != nil {
		return "", status.Error(codes.Internal, "group membership unavailable")
	}
	return encode(map[string]any{"joined": joined, "requested": requested})
}

func groupCancelRequest(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	result, err := db.ExecContext(ctx, `UPDATE jazebe_group_join_request SET status='cancelled',decided_at=now()
		WHERE planet_id=$1 AND group_id=$2 AND user_id=$3 AND status='pending'`, ref.PlanetID, ref.GroupID, user)
	if err != nil {
		return "", status.Error(codes.Internal, "group request cancellation failed")
	}
	count, err := result.RowsAffected()
	if err != nil {
		return "", status.Error(codes.Internal, "group request cancellation failed")
	}
	if count == 0 {
		var member bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership
			WHERE planet_id=$1 AND group_id=$2 AND user_id=$3 AND status='active')`, ref.PlanetID, ref.GroupID, user).Scan(&member); err != nil {
			return "", status.Error(codes.Internal, "group membership unavailable")
		}
		if member {
			return "", status.Error(codes.FailedPrecondition, "join request already decided")
		}
		return "", status.Error(codes.NotFound, "join request not found")
	}
	return encode(map[string]any{"requested": false})
}

func groupLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group leave unavailable")
	}
	defer tx.Rollback()
	var groupStatus string
	err = tx.QueryRowContext(ctx, `SELECT g.status FROM jazebe_group g
		JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE g.id=$1 AND g.planet_id=$2 AND p.status='active' AND b.is_active FOR UPDATE OF g`,
		ref.GroupID, ref.PlanetID).Scan(&groupStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group leave unavailable")
	}
	if groupStatus == "disbanded" {
		return "", status.Error(codes.FailedPrecondition, "group unavailable")
	}
	var role string
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership
		WHERE group_id=$1 AND planet_id=$2 AND user_id=$3 AND status='active' FOR UPDATE`,
		ref.GroupID, ref.PlanetID, user).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "active group membership required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group leave unavailable")
	}
	if role == "creator" {
		return "", status.Error(codes.PermissionDenied, "creator must transfer leadership or disband")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_membership SET status='left',left_at=now()
		WHERE group_id=$1 AND user_id=$2 AND status='active'`, ref.GroupID, user); err != nil {
		return "", status.Error(codes.Internal, "group leave failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_invite SET revoked_at=now()
		WHERE group_id=$1 AND created_by_user_id=$2 AND revoked_at IS NULL`, ref.GroupID, user); err != nil {
		return "", status.Error(codes.Internal, "group invite revocation failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='cancelled',decided_at=now()
		WHERE group_id=$1 AND status='pending' AND (from_user_id=$2 OR to_user_id=$2)`, ref.GroupID, user); err != nil {
		return "", status.Error(codes.Internal, "leadership offer cancellation failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group SET member_count=member_count-1,updated_at=now() WHERE id=$1`, ref.GroupID); err != nil {
		return "", status.Error(codes.Internal, "group count update failed")
	}
	var displayName sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT nullif(trim(display_name),'') FROM users WHERE id=$1`, user).Scan(&displayName); err != nil {
		return "", status.Error(codes.Internal, "profile unavailable")
	}
	name := displayName.String
	if name == "" {
		name = "کاربر ناشناس"
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
		VALUES($1,$2,$3,'leave',$4,'member.left',jsonb_build_object('userId',$4::uuid,'name',$5::text,'text',$6::text))`,
		uuid.Must(uuid.NewV4()).String(), ref.PlanetID, ref.GroupID, user, name, name+" از گروه خارج شد"); err != nil {
		return "", status.Error(codes.Internal, "group activity failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group leave failed")
	}
	if err := groupChatKickUser(nk, ref, user); err != nil {
		logger.Warn("Jazebe group stream eviction failed: %v", err)
	}
	return encode(map[string]any{"newCreatorUserId": nil})
}
