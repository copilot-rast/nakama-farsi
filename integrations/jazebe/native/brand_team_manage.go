package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Brand team management (source brand-team.service.ts): the active main
// administrator adds an existing user, changes roles and deactivates or
// reactivates administrators. Changes are serialized per brand, the brand
// always keeps an active main administrator, and every change is audited.
//
// Adaptation: the source identifies the invitee by a registered phone
// number. Nakama-Farsi has no trusted phone identity yet, so lookup and
// invite take an exact Nakama user ID or username instead (no search).

type brandTeamInput struct {
	BrandID  string `json:"brandId"`
	MemberID string `json:"memberId"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Role     string `json:"role"`
}

func validBrandRole(role string) bool {
	return role == "main" || role == "technical" || role == "marketing"
}

func requireMainBrandAdmin(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, brandID, user string) error {
	var allowed bool
	if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_brand_admin a JOIN jazebe_brand b ON b.id=a.brand_id
		WHERE a.brand_id=$1 AND a.user_id=$2 AND a.role='main' AND a.is_active AND b.is_active)`, brandID, user).Scan(&allowed); err != nil {
		return status.Error(codes.Internal, "brand team unavailable")
	}
	if !allowed {
		return status.Error(codes.PermissionDenied, "active main brand administrator required")
	}
	return nil
}

func decodeBrandTeam(payload string) (brandTeamInput, error) {
	var in brandTeamInput
	if err := decode(payload, &in); err != nil {
		return in, err
	}
	var err error
	if in.BrandID, err = requiredUUID(in.BrandID); err != nil {
		return in, err
	}
	return in, nil
}

// brandTeamUser resolves an exact Nakama user ID or username.
func brandTeamUser(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, in brandTeamInput) (string, string, error) {
	var id, name string
	var err error
	switch {
	case in.UserID != "":
		var uid string
		if uid, err = requiredUUID(in.UserID); err != nil {
			return "", "", err
		}
		err = q.QueryRowContext(ctx, `SELECT u.id,COALESCE(NULLIF(trim(u.display_name),''),'') FROM users u WHERE u.id=$1 AND u.disable_time='1970-01-01 00:00:00 UTC'`, uid).Scan(&id, &name)
	case strings.TrimSpace(in.Username) != "":
		username := strings.TrimSpace(in.Username)
		if utf8.RuneCountInString(username) > 128 {
			return "", "", status.Error(codes.InvalidArgument, "invalid username")
		}
		err = q.QueryRowContext(ctx, `SELECT u.id,COALESCE(NULLIF(trim(u.display_name),''),'') FROM users u WHERE u.username=$1 AND u.disable_time='1970-01-01 00:00:00 UTC'`, username).Scan(&id, &name)
	default:
		return "", "", status.Error(codes.InvalidArgument, "userId or username required")
	}
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", status.Error(codes.Internal, "user lookup unavailable")
	}
	return id, name, nil
}

func brandTeamLookup(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	in, err := decodeBrandTeam(payload)
	if err != nil {
		return "", err
	}
	if err := requireMainBrandAdmin(ctx, db, in.BrandID, user); err != nil {
		return "", err
	}
	id, name, err := brandTeamUser(ctx, db, in)
	if err != nil {
		return "", err
	}
	if id == "" {
		return encode(map[string]any{"found": false, "displayName": nil})
	}
	var display any
	if name != "" {
		display = name
	}
	return encode(map[string]any{"found": true, "userId": id, "displayName": display})
}

func brandTeamAudit(ctx context.Context, tx *sql.Tx, actorType, action, actor, brandID, memberID string, params map[string]any) error {
	raw, _ := json.Marshal(params)
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,brand_id,params)
		VALUES($1,$2,$3,'brand_admin',$4,$5,$6::jsonb)`, action, actorType, actor, memberID, brandID, string(raw)); err != nil {
		return status.Error(codes.Internal, "brand team audit failed")
	}
	return nil
}

// lockBrandTeam serializes team changes per brand and returns the target.
func lockBrandTeam(ctx context.Context, tx *sql.Tx, brandID, memberID string) (role string, active bool, err error) {
	if err = lockXact(ctx, tx, "jz-brand-team:"+brandID); err != nil {
		return "", false, status.Error(codes.Internal, "brand team unavailable")
	}
	err = tx.QueryRowContext(ctx, `SELECT role,is_active FROM jazebe_brand_admin WHERE brand_id=$1 AND user_id=$2`, brandID, memberID).Scan(&role, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, status.Error(codes.NotFound, "brand administrator not found")
	}
	if err != nil {
		return "", false, status.Error(codes.Internal, "brand team unavailable")
	}
	return role, active, nil
}

func assertAnotherMain(ctx context.Context, tx *sql.Tx, brandID string) error {
	var mains int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_brand_admin WHERE brand_id=$1 AND role='main' AND is_active`, brandID).Scan(&mains); err != nil {
		return status.Error(codes.Internal, "brand team unavailable")
	}
	if mains <= 1 {
		return status.Error(codes.FailedPrecondition, "brand must keep an active main administrator")
	}
	return nil
}

func brandTeamInvite(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	in, err := decodeBrandTeam(payload)
	if err != nil {
		return "", err
	}
	in.FullName = strings.TrimSpace(in.FullName)
	if !utf8.ValidString(in.FullName) || utf8.RuneCountInString(in.FullName) < 2 || utf8.RuneCountInString(in.FullName) > 128 {
		return "", status.Error(codes.InvalidArgument, "full name must be 2 to 128 characters")
	}
	if !validBrandRole(in.Role) {
		return "", status.Error(codes.InvalidArgument, "invalid brand role")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	defer tx.Rollback()
	if err := requireMainBrandAdmin(ctx, tx, in.BrandID, actor); err != nil {
		return "", err
	}
	if err = lockXact(ctx, tx, "jz-brand-team:"+in.BrandID); err != nil {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	target, _, err := brandTeamUser(ctx, tx, in)
	if err != nil {
		return "", err
	}
	if target == "" {
		return "", status.Error(codes.NotFound, "user not found")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_brand_member(brand_id,user_id,role) VALUES($1,$2,'admin') ON CONFLICT (brand_id,user_id) DO NOTHING`, in.BrandID, target); err != nil {
		return "", status.Error(codes.Internal, "brand membership failed")
	}
	var active sql.NullBool
	err = tx.QueryRowContext(ctx, `SELECT is_active FROM jazebe_brand_admin WHERE brand_id=$1 AND user_id=$2`, in.BrandID, target).Scan(&active)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	switch {
	case active.Valid && active.Bool:
		return "", status.Error(codes.AlreadyExists, "user is already a brand administrator")
	case active.Valid:
		// A deactivated administrator returns with the role asked for now.
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_brand_admin SET is_active=true,role=$3,full_name=$4 WHERE brand_id=$1 AND user_id=$2`, in.BrandID, target, in.Role, in.FullName)
	default:
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_brand_admin(brand_id,user_id,role,full_name,is_active) VALUES($1,$2,$3,$4,true)`, in.BrandID, target, in.Role, in.FullName)
	}
	if err != nil {
		logger.Error("Jazebe brand admin add failed: %v", err)
		return "", status.Error(codes.Internal, "brand administrator add failed")
	}
	if err := brandTeamAudit(ctx, tx, "brand_admin", "brand.admin_add", actor, in.BrandID, target, map[string]any{"userId": target, "role": in.Role}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "brand administrator add failed")
	}
	return brandTeamList(ctx, logger, db, nk, jsonString(map[string]string{"brandId": in.BrandID}))
}

func brandTeamRole(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	in, err := decodeBrandTeam(payload)
	if err != nil {
		return "", err
	}
	if in.MemberID, err = requiredUUID(in.MemberID); err != nil {
		return "", err
	}
	if !validBrandRole(in.Role) {
		return "", status.Error(codes.InvalidArgument, "invalid brand role")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	defer tx.Rollback()
	if err := requireMainBrandAdmin(ctx, tx, in.BrandID, actor); err != nil {
		return "", err
	}
	if err := brandTeamChangeRole(ctx, tx, "brand_admin", actor, in.BrandID, in.MemberID, in.Role); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "brand role update failed")
	}
	return brandTeamList(ctx, logger, db, nk, jsonString(map[string]string{"brandId": in.BrandID}))
}

func brandTeamSetActive(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string, activate bool) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	in, err := decodeBrandTeam(payload)
	if err != nil {
		return "", err
	}
	if in.MemberID, err = requiredUUID(in.MemberID); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	defer tx.Rollback()
	if err := requireMainBrandAdmin(ctx, tx, in.BrandID, actor); err != nil {
		return "", err
	}
	if err := brandTeamSetActiveTx(ctx, tx, "brand_admin", actor, in.BrandID, in.MemberID, activate); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "brand administrator update failed")
	}
	return brandTeamList(ctx, logger, db, nk, jsonString(map[string]string{"brandId": in.BrandID}))
}

func brandTeamDeactivate(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, p string) (string, error) {
	return brandTeamSetActive(ctx, l, db, nk, p, false)
}

func brandTeamReactivate(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, p string) (string, error) {
	return brandTeamSetActive(ctx, l, db, nk, p, true)
}

// brandTeamChangeRole changes an administrator's role; the brand keeps an
// active main administrator. Shared by the brand team page and operators.
func brandTeamChangeRole(ctx context.Context, tx *sql.Tx, actorType, actor, brandID, memberID, newRole string) error {
	role, active, err := lockBrandTeam(ctx, tx, brandID, memberID)
	if err != nil || role == newRole {
		return err
	}
	if role == "main" && active {
		if err := assertAnotherMain(ctx, tx, brandID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_brand_admin SET role=$3 WHERE brand_id=$1 AND user_id=$2`, brandID, memberID, newRole); err != nil {
		return status.Error(codes.Internal, "brand role update failed")
	}
	return brandTeamAudit(ctx, tx, actorType, "brand.admin_role", actor, brandID, memberID, map[string]any{"from": role, "to": newRole})
}

// brandTeamSetActiveTx deactivates or reactivates an administrator; the last
// active main administrator cannot be deactivated.
func brandTeamSetActiveTx(ctx context.Context, tx *sql.Tx, actorType, actor, brandID, memberID string, activate bool) error {
	role, active, err := lockBrandTeam(ctx, tx, brandID, memberID)
	if err != nil || active == activate {
		return err
	}
	if !activate && role == "main" {
		if err := assertAnotherMain(ctx, tx, brandID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_brand_admin SET is_active=$3 WHERE brand_id=$1 AND user_id=$2`, brandID, memberID, activate); err != nil {
		return status.Error(codes.Internal, "brand administrator update failed")
	}
	action := "brand.admin_deactivate"
	if activate {
		action = "brand.admin_reactivate"
	}
	return brandTeamAudit(ctx, tx, actorType, action, actor, brandID, memberID, map[string]any{"role": role})
}
