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

func missionInvitePreview(ctx context.Context, db *sql.DB, user, token string, empty map[string]any) (string, error) {
	var planetID, planetName, planetStatus, missionStatus, inviterID string
	var slug, groupID, groupName, groupStatus, icon, color, inviterName sql.NullString
	var members, cap sql.NullInt64
	var active bool
	var start, end time.Time
	var configRaw []byte
	err := db.QueryRowContext(ctx, `SELECT p.id,p.name,p.slug,p.status,b.is_active,m.status,m.starts_at,m.ends_at,
		i.inviter_user_id,nullif(trim(u.display_name),''),i.group_id,g.name,g.status,g.icon,g.color,
		g.member_count,g.member_cap,m.type_config
		FROM jazebe_mission_invite i JOIN jazebe_mission m ON m.id=i.mission_id
		JOIN jazebe_planet p ON p.id=i.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		JOIN users u ON u.id=i.inviter_user_id LEFT JOIN jazebe_group g ON g.id=i.group_id
		WHERE i.token=$1`, token).Scan(&planetID, &planetName, &slug, &planetStatus, &active, &missionStatus, &start, &end,
		&inviterID, &inviterName, &groupID, &groupName, &groupStatus, &icon, &color, &members, &cap, &configRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return encode(empty)
	}
	if err != nil {
		return "", err
	}
	config, err := parseReferralMissionConfig(configRaw)
	if err != nil {
		return "", err
	}
	if config.Scope != "group" {
		groupID = sql.NullString{}
	}
	var planetMember bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2 AND is_active)`, planetID, user).Scan(&planetMember); err != nil {
		return "", err
	}
	var current sql.NullString
	err = db.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`, planetID, user).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	groupMember := groupID.Valid && current.Valid && groupID.String == current.String
	var other any
	if current.Valid && !groupMember {
		other = current.String
	}
	state := "valid"
	if missionStatus != "active" || time.Now().UTC().Before(start) || !time.Now().UTC().Before(end) {
		state = "expired"
	} else if planetStatus != "active" || !active {
		state = "planet_unavailable"
	}
	var group any
	if groupID.Valid {
		if groupStatus.String != "active" {
			state = "group_unavailable"
		} else if !groupMember && members.Int64 >= cap.Int64 {
			state = "group_full"
		}
		var inviterMember bool
		if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active')`, groupID.String, inviterID).Scan(&inviterMember); err != nil {
			return "", err
		}
		if !inviterMember && state != "expired" {
			state = "revoked"
		}
		group = map[string]any{"id": groupID.String, "name": groupName.String, "icon": nullableString(icon), "color": nullableString(color), "members": members.Int64, "cap": cap.Int64}
	} else if config.Scope == "group" && state != "expired" {
		state = "revoked"
	}
	var inviter any
	if inviterName.Valid {
		inviter = map[string]any{"displayName": inviterName.String}
	}
	kind := "planet"
	if groupID.Valid {
		kind = "group"
	}
	return encode(map[string]any{"kind": kind, "status": state, "planet": map[string]any{"id": planetID, "name": planetName, "slug": slug.String},
		"group": group, "inviter": inviter, "expiresAt": end, "viewer": map[string]any{"planetMember": planetMember, "groupMember": groupMember, "otherGroupId": other}})
}

// A mission invitation is attributed only when this transaction creates the
// invited membership. A token preview or a second join never creates credit.
func missionInviteJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, user, token string) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var invite referralInvite
	var missionStatus, planetStatus string
	var brandActive bool
	var start, end time.Time
	err = tx.QueryRowContext(ctx, `SELECT i.id,i.planet_id,i.inviter_user_id,i.mission_id,i.invitee_user_id,i.group_id,i.status,
		m.type_config,m.status,p.status,b.is_active,m.starts_at,m.ends_at
		FROM jazebe_mission_invite i JOIN jazebe_mission m ON m.id=i.mission_id
		JOIN jazebe_planet p ON p.id=i.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE i.token=$1 FOR UPDATE OF i`, token).Scan(&invite.ID, &invite.PlanetID, &invite.Inviter,
		&invite.MissionID, &invite.Invitee, &invite.GroupID, &invite.Status, &invite.Config, &missionStatus, &planetStatus, &brandActive, &start, &end)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "invitation not found")
	}
	if err != nil {
		return "", err
	}
	if invite.Inviter == user {
		return "", status.Error(codes.PermissionDenied, "self invitation is not allowed")
	}
	now := time.Now().UTC()
	if missionStatus != "active" || planetStatus != "active" || !brandActive || now.Before(start) || !now.Before(end) {
		return "", status.Error(codes.FailedPrecondition, "invitation unavailable")
	}
	config, err := parseReferralMissionConfig(invite.Config)
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(774103,hashtext($1))`, invite.PlanetID+user); err != nil {
		return "", err
	}
	if err = requireUsablePlanetMembership(ctx, tx, invite.PlanetID, user); err != nil {
		return "", err
	}
	var display sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT nullif(trim(display_name),'') FROM users WHERE id=$1`, user).Scan(&display); err != nil {
		return "", err
	}
	if !display.Valid {
		return "", status.Error(codes.FailedPrecondition, "display name required")
	}
	var alreadyPlanet bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2)`, invite.PlanetID, user).Scan(&alreadyPlanet); err != nil {
		return "", err
	}
	groupID := ""
	joined := false
	planetJoined := false
	membershipID := ""
	if config.Scope == "group" {
		if !invite.GroupID.Valid {
			return "", status.Error(codes.FailedPrecondition, "invite group unavailable")
		}
		groupID = invite.GroupID.String
		var groupStatus string
		var members, cap int
		err = tx.QueryRowContext(ctx, `SELECT status,member_count,member_cap FROM jazebe_group WHERE id=$1 AND planet_id=$2 FOR UPDATE`, groupID, invite.PlanetID).Scan(&groupStatus, &members, &cap)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "invite group not found")
		}
		if err != nil {
			return "", err
		}
		if groupStatus != "active" {
			return "", status.Error(codes.FailedPrecondition, "invite group unavailable")
		}
		var inviterMember bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active')`, groupID, invite.Inviter).Scan(&inviterMember); err != nil {
			return "", err
		}
		if !inviterMember {
			return "", status.Error(codes.FailedPrecondition, "inviter left group")
		}
		var current sql.NullString
		err = tx.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`, invite.PlanetID, user).Scan(&current)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		if current.Valid {
			if current.String != groupID {
				return "", status.Error(codes.AlreadyExists, "already in another group")
			}
			return encode(map[string]any{"planetId": invite.PlanetID, "groupId": groupID, "joined": false, "planetJoined": false})
		}
		if members >= cap {
			return "", status.Error(codes.ResourceExhausted, "group full")
		}
		membershipID = uuid.Must(uuid.NewV4()).String()
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_membership(id,group_id,planet_id,user_id,role,source)
			VALUES($1,$2,$3,$4,'member','invite_link')`, membershipID, groupID, invite.PlanetID, user); err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group SET member_count=member_count+1,updated_at=now() WHERE id=$1`, groupID); err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_join_request SET status='auto_cancelled',decided_at=now()
			WHERE planet_id=$1 AND user_id=$2 AND status='pending'`, invite.PlanetID, user); err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
			VALUES($1,$2,$3,'join',$4,'member.joined',jsonb_build_object('userId',$4::uuid,'name',$5::text,'source','invite_link','text',$6::text))`,
			uuid.Must(uuid.NewV4()).String(), invite.PlanetID, groupID, user, display.String, display.String+" به گروه پیوست"); err != nil {
			return "", err
		}
		if err = enqueueActivityXP(ctx, tx, invite.PlanetID, user, membershipID, "group_join"); err != nil {
			return "", err
		}
		joined = true
	} else if alreadyPlanet {
		return encode(map[string]any{"planetId": invite.PlanetID, "groupId": nil, "joined": false, "planetJoined": false})
	}
	joinPath := "join_button"
	if joined {
		joinPath = "group_invite_link"
	}
	if err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership(planet_id,user_id,join_path,last_activity_at)
		VALUES($1,$2,$3,now()) ON CONFLICT(planet_id,user_id) DO NOTHING RETURNING true`, invite.PlanetID, user, joinPath).Scan(&planetJoined); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if planetJoined {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=fans_count+1 WHERE id=$1`, invite.PlanetID); err != nil {
			return "", err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=now() WHERE planet_id=$1 AND user_id=$2`, invite.PlanetID, user); err != nil {
		return "", err
	}
	if _, err = referralIdentity(ctx, tx, user); err != nil {
		return "", err
	}
	var duplicate string
	err = tx.QueryRowContext(ctx, `SELECT id FROM jazebe_mission_invite WHERE inviter_user_id=$1 AND invitee_user_id=$2 AND mission_id=$3`, invite.Inviter, user, invite.MissionID.String).Scan(&duplicate)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if errors.Is(err, sql.ErrNoRows) {
		inviteID := invite.ID
		if invite.Invitee.Valid {
			inviteID = uuid.Must(uuid.NewV4()).String()
			newToken, e := referralToken(32)
			if e != nil {
				return "", e
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission_invite(id,planet_id,mission_id,inviter_user_id,invitee_user_id,token,group_id,status,registered_at)
				VALUES($1,$2,$3,$4,$5,$6,$7,'registered',$8)`, inviteID, invite.PlanetID, invite.MissionID.String, invite.Inviter, user, newToken, nullableInviteID(invite.GroupID), now)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE jazebe_mission_invite SET invitee_user_id=$2,status='registered',registered_at=$3,updated_at=$3 WHERE id=$1 AND status='pending'`, inviteID, user, now)
		}
		if err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_referral_identity SET referred_by_user_id=$2,updated_at=$3 WHERE user_id=$1 AND referred_by_user_id IS NULL`, user, invite.Inviter, now); err != nil {
			return "", err
		}
		if config.Scope == "group" || config.SuccessCondition == "signup" {
			if err = referralQualify(ctx, tx, inviteID, invite.PlanetID, invite.Inviter, now); err != nil {
				return "", err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	if joined {
		groupJoinProgression(ctx, logger, db, invite.PlanetID, user, membershipID)
	}
	var group any
	if groupID != "" {
		group = groupID
	}
	return encode(map[string]any{"planetId": invite.PlanetID, "groupId": group, "joined": joined, "planetJoined": planetJoined})
}
