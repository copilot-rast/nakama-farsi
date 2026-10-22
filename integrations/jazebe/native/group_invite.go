package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var groupInviteTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{20,64}$`)

func groupInviteToken(payload string) (string, error) {
	var input struct {
		Token string `json:"token"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	value := strings.TrimSpace(input.Token)
	if !groupInviteTokenPattern.MatchString(value) {
		return "", status.Error(codes.InvalidArgument, "invalid invite token")
	}
	return value, nil
}

func groupInviteMint(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	var active bool
	err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group g JOIN jazebe_group_membership m ON m.group_id=g.id
		JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE g.id=$1 AND g.planet_id=$2 AND g.status='active' AND p.status='active' AND b.is_active
		AND m.user_id=$3 AND m.status='active')`,
		ref.GroupID, ref.PlanetID, user).Scan(&active)
	if err != nil {
		return "", status.Error(codes.Internal, "group invite unavailable")
	}
	if !active {
		return "", status.Error(codes.PermissionDenied, "active group membership required")
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return "", status.Error(codes.Internal, "invite token unavailable")
	}
	token := base64.RawURLEncoding.EncodeToString(secret[:])
	digest := sha256.Sum256([]byte(token))
	expires := time.Now().UTC().Add(7 * 24 * time.Hour)
	_, err = db.ExecContext(ctx, `INSERT INTO jazebe_group_invite(id,planet_id,group_id,created_by_user_id,token_hash,expires_at)
		VALUES($1,$2,$3,$4,$5,$6)`, uuid.Must(uuid.NewV4()).String(), ref.PlanetID, ref.GroupID, user, digest[:], expires)
	if err != nil {
		return "", status.Error(codes.Internal, "group invite creation failed")
	}
	base := strings.TrimRight(os.Getenv("JAZEBE_MAIN_APP_URL"), "/")
	return encode(map[string]any{"token": token, "url": base + "/g/" + token, "expiresAt": expires})
}

func groupInviteRevoke(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		Token    string `json:"token"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	input.Token = strings.TrimSpace(input.Token)
	if !groupInviteTokenPattern.MatchString(input.Token) {
		return "", status.Error(codes.InvalidArgument, "invalid invite token")
	}
	var role string
	err = db.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership
		WHERE group_id=$1 AND planet_id=$2 AND user_id=$3 AND status='active'`, ref.GroupID, ref.PlanetID, user).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "active group membership required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group invite unavailable")
	}
	digest := sha256.Sum256([]byte(input.Token))
	var owner string
	err = db.QueryRowContext(ctx, `SELECT created_by_user_id FROM jazebe_group_invite
		WHERE group_id=$1 AND planet_id=$2 AND token_hash=$3`, ref.GroupID, ref.PlanetID, digest[:]).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group invite not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group invite unavailable")
	}
	if owner != user && role != "creator" && role != "admin" {
		return "", status.Error(codes.PermissionDenied, "invite owner or manager required")
	}
	if _, err := db.ExecContext(ctx, `UPDATE jazebe_group_invite SET revoked_at=COALESCE(revoked_at,now())
		WHERE group_id=$1 AND planet_id=$2 AND token_hash=$3`, ref.GroupID, ref.PlanetID, digest[:]); err != nil {
		return "", status.Error(codes.Internal, "group invite revocation failed")
	}
	return encode(map[string]any{"success": true})
}

func groupInvitePreview(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		Token string `json:"token"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	token := strings.TrimSpace(input.Token)
	empty := map[string]any{"kind": "group", "status": "invalid", "planet": nil, "group": nil, "inviter": nil,
		"expiresAt": nil, "viewer": map[string]any{"planetMember": false, "groupMember": false, "otherGroupId": nil}}
	if !groupInviteTokenPattern.MatchString(token) {
		return encode(empty)
	}
	digest := sha256.Sum256([]byte(token))
	var planetID, groupID, planetName, planetStatus, groupName, groupStatus string
	var planetSlug, icon, color, inviter sql.NullString
	var expires time.Time
	var revoked sql.NullTime
	var uses, members, cap int
	var maxUses sql.NullInt64
	var brandActive bool
	err = db.QueryRowContext(ctx, `SELECT i.planet_id,i.group_id,p.name,p.slug,p.status,b.is_active,
		g.name,g.status,g.icon,g.color,g.member_count,g.member_cap,
		coalesce(nullif(trim(u.display_name),''),''),i.expires_at,i.revoked_at,i.use_count,i.max_uses
		FROM jazebe_group_invite i JOIN jazebe_planet p ON p.id=i.planet_id
		JOIN jazebe_brand b ON b.id=p.brand_id JOIN jazebe_group g ON g.id=i.group_id
		JOIN users u ON u.id=i.created_by_user_id WHERE i.token_hash=$1`, digest[:]).Scan(
		&planetID, &groupID, &planetName, &planetSlug, &planetStatus, &brandActive,
		&groupName, &groupStatus, &icon, &color, &members, &cap, &inviter, &expires, &revoked, &uses, &maxUses)
	if errors.Is(err, sql.ErrNoRows) {
		return missionInvitePreview(ctx, db, user, token, empty)
	}
	if err != nil {
		return "", status.Error(codes.Internal, "invite preview unavailable")
	}
	var planetMember bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet_membership
		WHERE planet_id=$1 AND user_id=$2)`, planetID, user).Scan(&planetMember); err != nil {
		return "", status.Error(codes.Internal, "invite preview unavailable")
	}
	var current sql.NullString
	err = db.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`,
		planetID, user).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "invite preview unavailable")
	}
	groupMember := current.Valid && current.String == groupID
	var other any
	if current.Valid && !groupMember {
		other = current.String
	}
	viewStatus := "valid"
	if revoked.Valid {
		viewStatus = "revoked"
	} else if time.Now().UTC().After(expires) || (maxUses.Valid && int64(uses) >= maxUses.Int64) {
		viewStatus = "expired"
	} else if planetStatus != "active" || !brandActive {
		viewStatus = "planet_unavailable"
	} else if groupStatus != "active" {
		viewStatus = "group_unavailable"
	} else if !groupMember && members >= cap {
		viewStatus = "group_full"
	}
	var iconValue, colorValue, inviterView any
	if icon.Valid {
		iconValue = icon.String
	}
	if color.Valid {
		colorValue = color.String
	}
	if inviter.Valid && inviter.String != "" {
		inviterView = map[string]any{"displayName": inviter.String}
	}
	return encode(map[string]any{"kind": "group", "status": viewStatus, "planet": map[string]any{
		"id": planetID, "name": planetName, "slug": planetSlug.String},
		"group":   map[string]any{"id": groupID, "name": groupName, "icon": iconValue, "color": colorValue, "members": members, "cap": cap},
		"inviter": inviterView, "expiresAt": expires, "viewer": map[string]any{
			"planetMember": planetMember, "groupMember": groupMember, "otherGroupId": other}})
}

func groupInviteJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	token, err := groupInviteToken(payload)
	if err != nil {
		return "", status.Error(codes.NotFound, "group invite not found")
	}
	digest := sha256.Sum256([]byte(token))
	var planetID, groupID string
	err = db.QueryRowContext(ctx, `SELECT planet_id,group_id FROM jazebe_group_invite WHERE token_hash=$1`, digest[:]).Scan(&planetID, &groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return missionInviteJoin(ctx, logger, db, user, token)
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group invite unavailable")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group invite join unavailable")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(774103,hashtext($1))`, planetID+user); err != nil {
		return "", status.Error(codes.Internal, "group invite join unavailable")
	}
	var groupStatus, planetStatus string
	var brandActive bool
	var members, cap int
	err = tx.QueryRowContext(ctx, `SELECT g.status,p.status,b.is_active,g.member_count,g.member_cap
		FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE g.id=$1 AND g.planet_id=$2 FOR UPDATE OF g`, groupID, planetID).
		Scan(&groupStatus, &planetStatus, &brandActive, &members, &cap)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group invite join unavailable")
	}
	var inviteID string
	var revoked sql.NullTime
	var expires time.Time
	var uses int
	var maxUses sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT id,revoked_at,expires_at,use_count,max_uses FROM jazebe_group_invite
		WHERE token_hash=$1 AND planet_id=$2 AND group_id=$3 FOR UPDATE`, digest[:], planetID, groupID).
		Scan(&inviteID, &revoked, &expires, &uses, &maxUses)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group invite not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group invite join unavailable")
	}
	if revoked.Valid {
		return "", status.Error(codes.PermissionDenied, "group invite revoked")
	}
	if time.Now().UTC().After(expires) || (maxUses.Valid && int64(uses) >= maxUses.Int64) {
		return "", status.Error(codes.PermissionDenied, "group invite expired")
	}
	if planetStatus != "active" || !brandActive {
		return "", status.Error(codes.FailedPrecondition, "planet unavailable")
	}
	if groupStatus != "active" {
		return "", status.Error(codes.FailedPrecondition, "group unavailable")
	}
	if err := requireUsablePlanetMembership(ctx, tx, planetID, user); err != nil {
		return "", err
	}
	var displayName sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT nullif(trim(display_name),'') FROM users WHERE id=$1`, user).Scan(&displayName); err != nil {
		return "", status.Error(codes.Internal, "profile unavailable")
	}
	if !displayName.Valid {
		return "", status.Error(codes.FailedPrecondition, "display name required")
	}
	var current sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`,
		planetID, user).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "group membership unavailable")
	}
	if current.Valid && current.String == groupID {
		return encode(map[string]any{"planetId": planetID, "groupId": groupID, "joined": false, "planetJoined": false})
	}
	if current.Valid {
		return "", status.Error(codes.AlreadyExists, "already in another group")
	}
	if members >= cap {
		return "", status.Error(codes.ResourceExhausted, "group full")
	}
	membershipID := uuid.Must(uuid.NewV4()).String()
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_membership(id,group_id,planet_id,user_id,role,source)
		VALUES($1,$2,$3,$4,'member','invite_link')`, membershipID, groupID, planetID, user); err != nil {
		return "", status.Error(codes.Internal, "group invite membership failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group SET member_count=member_count+1,updated_at=now() WHERE id=$1`, groupID); err != nil {
		return "", status.Error(codes.Internal, "group count update failed")
	}
	var planetJoined bool
	if err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership(planet_id,user_id,join_path,last_activity_at)
		VALUES($1,$2,'group_invite_link',now()) ON CONFLICT(planet_id,user_id) DO NOTHING RETURNING true`, planetID, user).
		Scan(&planetJoined); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "planet membership failed")
	}
	if planetJoined {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=fans_count+1 WHERE id=$1`, planetID); err != nil {
			return "", status.Error(codes.Internal, "fan count update failed")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=now() WHERE planet_id=$1 AND user_id=$2`, planetID, user); err != nil {
		return "", status.Error(codes.Internal, "planet activity failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_join_request SET status='auto_cancelled',decided_at=now()
		WHERE planet_id=$1 AND user_id=$2 AND status='pending'`, planetID, user); err != nil {
		return "", status.Error(codes.Internal, "group request cleanup failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
		VALUES($1,$2,$3,'join',$4,'member.joined',jsonb_build_object('userId',$4,'name',$5,'source','invite_link','text',$6))`,
		uuid.Must(uuid.NewV4()).String(), planetID, groupID, user, displayName.String, displayName.String+" به گروه پیوست"); err != nil {
		return "", status.Error(codes.Internal, "group activity failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_invite SET use_count=use_count+1 WHERE id=$1`, inviteID); err != nil {
		return "", status.Error(codes.Internal, "group invite usage failed")
	}
	if err := enqueueActivityXP(ctx, tx, planetID, user, membershipID, "group_join"); err != nil {
		return "", status.Error(codes.Internal, "group progression event failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group invite join failed")
	}
	groupJoinProgression(ctx, logger, db, planetID, user, membershipID)
	return encode(map[string]any{"planetId": planetID, "groupId": groupID, "joined": true, "planetJoined": planetJoined})
}
