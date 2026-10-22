package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var groupHexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type groupCreateInput struct {
	PlanetID    string `json:"planetId"`
	Name        string `json:"name"`
	Description string `json:"desc"`
	Mode        string `json:"mode"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	Cap         int    `json:"cap"`
}

func validateGroupCreate(input groupCreateInput) (groupCreateInput, error) {
	var err error
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return input, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if !utf8.ValidString(input.Name) || utf8.RuneCountInString(input.Name) < 1 || utf8.RuneCountInString(input.Name) > 64 ||
		!utf8.ValidString(input.Description) || utf8.RuneCountInString(input.Description) > 280 ||
		(input.Mode != "open" && input.Mode != "approval") || input.Cap < 1 || input.Cap > 10000 {
		return input, status.Error(codes.InvalidArgument, "invalid group fields")
	}
	if input.Color != "" && !groupHexColor.MatchString(input.Color) {
		return input, status.Error(codes.InvalidArgument, "invalid group color")
	}
	if input.Icon != "" {
		switch input.Icon {
		case "star", "bolt", "trophy", "ticket", "crown", "user", "magnifier", "drink":
		default:
			return input, status.Error(codes.InvalidArgument, "invalid group icon")
		}
	}
	return input, nil
}

// A Jazebe group is planet-scoped. It shares Nakama user IDs but keeps the
// source membership rules, which Nakama's general-purpose groups do not have.
func groupCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input groupCreateInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input, err = validateGroupCreate(input)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group creation unavailable")
	}
	defer tx.Rollback()
	var enabled bool
	var maximum sql.NullInt64
	var allowedModes []byte
	var capMin, capMax int
	var planetStatus string
	var brandActive bool
	err = tx.QueryRowContext(ctx, `SELECT s.group_creation_enabled,s.group_max_count,s.allowed_group_modes,
		s.group_member_cap_min,s.group_member_cap_max,p.status,b.is_active FROM jazebe_planet_settings s
		JOIN jazebe_planet p ON p.id=s.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE p.id=$1 FOR UPDATE OF s`, input.PlanetID).
		Scan(&enabled, &maximum, &allowedModes, &capMin, &capMax, &planetStatus, &brandActive)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group settings unavailable")
	}
	if planetStatus != "active" || !brandActive || !enabled {
		return "", status.Error(codes.FailedPrecondition, "group creation disabled")
	}
	if err := requireUsablePlanetMembership(ctx, tx, input.PlanetID, user); err != nil {
		return "", err
	}
	if input.Cap < capMin || input.Cap > capMax {
		return "", status.Error(codes.InvalidArgument, "group cap outside planet policy")
	}
	var modes []string
	if json.Unmarshal(allowedModes, &modes) != nil {
		return "", status.Error(codes.Internal, "group settings invalid")
	}
	modeAllowed := false
	for _, mode := range modes {
		if mode == input.Mode {
			modeAllowed = true
			break
		}
	}
	if !modeAllowed {
		return "", status.Error(codes.FailedPrecondition, "group mode not allowed")
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_group WHERE planet_id=$1 AND status='active'`, input.PlanetID).Scan(&count); err != nil {
		return "", status.Error(codes.Internal, "group creation unavailable")
	}
	if maximum.Valid && int64(count) >= maximum.Int64 {
		return "", status.Error(codes.ResourceExhausted, "group limit reached")
	}
	var displayName sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT nullif(trim(display_name),'') FROM users WHERE id=$1`, user).Scan(&displayName); err != nil {
		return "", status.Error(codes.Internal, "profile unavailable")
	}
	if !displayName.Valid {
		return "", status.Error(codes.FailedPrecondition, "display name required")
	}
	var already bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership
		WHERE planet_id=$1 AND user_id=$2 AND status='active')`, input.PlanetID, user).Scan(&already); err != nil {
		return "", status.Error(codes.Internal, "group membership unavailable")
	}
	if already {
		return "", status.Error(codes.AlreadyExists, "already in group")
	}
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group(id,planet_id,name,description,icon,color,mode,member_cap,creator_user_id)
		VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7,$8,$9)`,
		id, input.PlanetID, input.Name, input.Description, input.Icon, input.Color, input.Mode, input.Cap, user)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "group name already reserved")
		}
		logger.Error("Jazebe group insert: %v", err)
		return "", status.Error(codes.Internal, "group creation failed")
	}
	membershipID := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_membership(id,group_id,planet_id,user_id,role,source)
		VALUES($1,$2,$3,$4,'creator','creation')`, membershipID, id, input.PlanetID, user)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "already in group")
		}
		return "", status.Error(codes.Internal, "group membership creation failed")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_join_request SET status='auto_cancelled',decided_at=now()
		WHERE planet_id=$1 AND user_id=$2 AND status='pending'`, input.PlanetID, user); err != nil {
		return "", status.Error(codes.Internal, "group request cleanup failed")
	}
	var first bool
	if err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership(planet_id,user_id,join_path,last_activity_at)
		VALUES($1,$2,'first_activity',now()) ON CONFLICT(planet_id,user_id) DO NOTHING RETURNING true`, input.PlanetID, user).Scan(&first); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "planet membership creation failed")
	}
	if first {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=fans_count+1 WHERE id=$1`, input.PlanetID); err != nil {
			return "", status.Error(codes.Internal, "fan count update failed")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=now() WHERE planet_id=$1 AND user_id=$2`, input.PlanetID, user); err != nil {
		return "", status.Error(codes.Internal, "planet activity update failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('group.create','user',$1,'group',$2,$3,jsonb_build_object('name',$4))`, user, id, input.PlanetID, input.Name); err != nil {
		return "", status.Error(codes.Internal, "group audit failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
		VALUES($1,$2,$3,'join',$4,'group.created',jsonb_build_object('name',$5,'text',$6))`,
		uuid.Must(uuid.NewV4()).String(), input.PlanetID, id, user, input.Name, displayName.String+" گروه را ساخت"); err != nil {
		return "", status.Error(codes.Internal, "group activity failed")
	}
	if err := enqueueActivityXP(ctx, tx, input.PlanetID, user, membershipID, "group_join"); err != nil {
		return "", status.Error(codes.Internal, "group progression event failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group creation failed")
	}
	groupJoinProgression(ctx, logger, db, input.PlanetID, user, membershipID)
	color := input.Color
	if color == "" {
		color = "#2F6EBC"
	}
	view := map[string]any{"id": id, "name": input.Name, "members": 1, "cap": input.Cap, "mode": input.Mode, "rank": 1, "color": color}
	if input.Description != "" {
		view["sub"] = input.Description
		view["desc"] = input.Description
	}
	if input.Icon != "" {
		view["icon"] = input.Icon
	}
	return encode(view)
}

func groupMeta(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := userID(ctx); err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var enabled bool
	var maximum sql.NullInt64
	var modes []byte
	var capMin, capMax, adminCap, active int
	err = db.QueryRowContext(ctx, `SELECT s.group_creation_enabled,s.group_max_count,s.allowed_group_modes,
		s.group_member_cap_min,s.group_member_cap_max,s.group_admin_cap,
		(SELECT count(*) FROM jazebe_group g WHERE g.planet_id=s.planet_id AND g.status='active')
		FROM jazebe_planet_settings s JOIN jazebe_planet p ON p.id=s.planet_id
		JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE p.id=$1 AND p.status='active' AND b.is_active`, planetID).Scan(&enabled, &maximum, &modes, &capMin, &capMax, &adminCap, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group settings unavailable")
	}
	var allowed []string
	if json.Unmarshal(modes, &allowed) != nil {
		return "", status.Error(codes.Internal, "group settings invalid")
	}
	limitReached := maximum.Valid && int64(active) >= maximum.Int64
	var max any
	if maximum.Valid {
		max = maximum.Int64
	}
	return encode(map[string]any{"enabled": enabled, "limitReached": limitReached, "canCreate": enabled && !limitReached,
		"allowedModes": allowed, "capMin": capMin, "capMax": capMax, "adminCap": adminCap, "max": max})
}

func groupNameAvailability(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := userID(ctx); err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		Name     string `json:"name"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	input.Name = strings.TrimSpace(input.Name)
	if !utf8.ValidString(input.Name) || utf8.RuneCountInString(input.Name) < 1 || utf8.RuneCountInString(input.Name) > 64 {
		return "", status.Error(codes.InvalidArgument, "invalid group name")
	}
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group WHERE planet_id=$1 AND lower(name)=lower($2))`,
		planetID, input.Name).Scan(&exists); err != nil {
		return "", status.Error(codes.Internal, "group name lookup unavailable")
	}
	return encode(map[string]any{"name": input.Name, "available": !exists})
}
