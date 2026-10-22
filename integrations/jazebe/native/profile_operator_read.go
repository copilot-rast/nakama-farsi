package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type operatorUserInput struct {
	UserID       string `json:"userId"`
	Search       string `json:"search"`
	State        string `json:"state"`
	Verification string `json:"verification"`
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
}

func validateOperatorUserFilter(in operatorUserInput) (operatorUserInput, error) {
	in.Search = strings.TrimSpace(in.Search)
	if !utf8.ValidString(in.Search) || utf8.RuneCountInString(in.Search) > 64 {
		return in, status.Error(codes.InvalidArgument, "invalid user search")
	}
	switch in.State {
	case "", "active", "suspended", "deactivated":
	default:
		return in, status.Error(codes.InvalidArgument, "invalid account state")
	}
	switch in.Verification {
	case "", "none", "review", "verified", "rejected":
	default:
		return in, status.Error(codes.InvalidArgument, "invalid verification status")
	}
	if in.Page < 0 || in.Page > 100000 || in.Limit < 0 || in.Limit > 100 {
		return in, status.Error(codes.InvalidArgument, "invalid user page")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	return in, nil
}

const operatorUserSelect = `SELECT u.id,u.username,u.email,u.verify_time,u.disable_time,u.create_time,
	p.display_name,p.avatar_kind,p.avatar_preset_index,a.state,a.last_login_at,a.deactivated_at,
	a.suspended_at,a.suspended_reason,v.status,ri.invite_code,
	(SELECT count(*) FROM jazebe_planet_membership pm WHERE pm.user_id=u.id)
	FROM users u
	LEFT JOIN jazebe_user_profile p ON p.user_id=u.id
	LEFT JOIN jazebe_user_account a ON a.user_id=u.id
	LEFT JOIN jazebe_user_verification v ON v.user_id=u.id
	LEFT JOIN jazebe_referral_identity ri ON ri.user_id=u.id`

func scanOperatorUser(row interface{ Scan(...any) error }) (map[string]any, error) {
	var id, username string
	var email, displayName, avatarKind, accountState, suspendedReason, verification, inviteCode sql.NullString
	var avatarIndex sql.NullInt64
	var verifiedAt, disabledAt, lastLoginAt, deactivatedAt, suspendedAt sql.NullTime
	var createdAt time.Time
	var joined int
	err := row.Scan(&id, &username, &email, &verifiedAt, &disabledAt, &createdAt,
		&displayName, &avatarKind, &avatarIndex, &accountState, &lastLoginAt, &deactivatedAt,
		&suspendedAt, &suspendedReason, &verification, &inviteCode, &joined)
	if err != nil {
		return nil, err
	}
	var emailValue, displayValue, avatarValue, lastLoginValue, deactivatedValue, suspendedValue, reasonValue, inviteValue any
	if email.Valid {
		emailValue = email.String
	}
	if displayName.Valid {
		displayValue = displayName.String
	}
	if avatarKind.Valid {
		if avatarKind.String == "upload" {
			avatarValue = -1
		} else if avatarIndex.Valid {
			avatarValue = avatarIndex.Int64
		}
	}
	if lastLoginAt.Valid {
		lastLoginValue = lastLoginAt.Time
	}
	if deactivatedAt.Valid {
		deactivatedValue = deactivatedAt.Time
	}
	if suspendedAt.Valid {
		suspendedValue = suspendedAt.Time
	}
	if suspendedReason.Valid {
		reasonValue = suspendedReason.String
	}
	if inviteCode.Valid {
		inviteValue = inviteCode.String
	}
	state := "active"
	if accountState.Valid {
		state = accountState.String
	}
	coreStatus := "active"
	if disabledAt.Valid && disabledAt.Time.Unix() != 0 {
		coreStatus = "suspended"
	}
	verifyState := "none"
	if verification.Valid {
		verifyState = verification.String
	}
	return map[string]any{
		"id": id, "name": nil, "lastName": nil, "username": username, "email": emailValue,
		"emailVerified": verifiedAt.Valid && verifiedAt.Time.Unix() != 0,
		"phoneNumber":   nil, "phoneNumberVerified": false, "status": coreStatus,
		"createdAt": createdAt, "displayName": displayValue, "avatarIndex": avatarValue,
		"accountState": state, "verificationStatus": verifyState, "planetsJoined": joined,
		"lastLoginAt": lastLoginValue, "deactivatedAt": deactivatedValue,
		"suspendedAt": suspendedValue, "suspendedReason": reasonValue, "inviteCode": inviteValue,
	}, nil
}

func operatorUserList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := operatorReportActor(ctx); err != nil {
		return "", err
	}
	var in operatorUserInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	in, err := validateOperatorUserFilter(in)
	if err != nil {
		return "", err
	}
	filter := ` WHERE u.id<>'00000000-0000-0000-0000-000000000000'::uuid
		AND ($1='' OR p.display_name ILIKE $1 OR u.username ILIKE $1 OR u.email ILIKE $1)
		AND ($2='' OR COALESCE(a.state,'active')=$2)
		AND ($3='' OR COALESCE(v.status,'none')=$3)`
	search := ""
	if in.Search != "" {
		search = "%" + in.Search + "%"
	}
	var total int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM users u
		LEFT JOIN jazebe_user_profile p ON p.user_id=u.id
		LEFT JOIN jazebe_user_account a ON a.user_id=u.id
		LEFT JOIN jazebe_user_verification v ON v.user_id=u.id`+filter,
		search, in.State, in.Verification).Scan(&total); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, operatorUserSelect+filter+` ORDER BY u.create_time DESC,u.id DESC LIMIT $4 OFFSET $5`,
		search, in.State, in.Verification, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		item, e := scanOperatorUser(rows)
		if e != nil {
			return "", e
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

func operatorUserGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := operatorReportActor(ctx); err != nil {
		return "", err
	}
	var in operatorUserInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.UserID)
	if err != nil {
		return "", err
	}
	if id == "00000000-0000-0000-0000-000000000000" {
		return "", status.Error(codes.NotFound, "user not found")
	}
	user, err := scanOperatorUser(db.QueryRowContext(ctx, operatorUserSelect+` WHERE u.id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT m.planet_id,p.name,m.join_path,m.joined_at,m.is_active,m.last_activity_at,
		COALESCE(s.total_xp,0),COALESCE(s.level,1),COALESCE(s.medals_count,0)
		FROM jazebe_planet_membership m JOIN jazebe_planet p ON p.id=m.planet_id
		LEFT JOIN jazebe_user_planet_stats s ON s.planet_id=m.planet_id AND s.user_id=m.user_id
		WHERE m.user_id=$1 ORDER BY m.joined_at DESC,m.planet_id DESC LIMIT 501`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	planets := make([]map[string]any, 0)
	for rows.Next() {
		var planetID, name, joinPath string
		var joinedAt time.Time
		var active bool
		var last sql.NullTime
		var xp int64
		var level, medals int
		if err = rows.Scan(&planetID, &name, &joinPath, &joinedAt, &active, &last, &xp, &level, &medals); err != nil {
			return "", err
		}
		var activity any
		if last.Valid {
			activity = last.Time
		}
		planets = append(planets, map[string]any{"planetId": planetID, "planetName": name, "joinPath": joinPath,
			"joinedAt": joinedAt, "isActive": active, "lastActivityAt": activity, "totalXp": xp, "level": level, "medalsCount": medals})
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	if len(planets) > 500 {
		return "", status.Error(codes.ResourceExhausted, "too many planet memberships for detail response")
	}
	user["planets"] = planets
	// Nakama has no trusted phone or sealed KYC identity fields. A status-only
	// verification record must not be presented as complete identity evidence.
	user["verification"] = map[string]any{"status": user["verificationStatus"], "identityAvailable": false}
	return encode(user)
}

func operatorUserAudit(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := operatorReportActor(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "super operator required")
	}
	var in operatorUserInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.UserID, err = requiredUUID(in.UserID)
	if err != nil {
		return "", err
	}
	in, err = validateOperatorUserFilter(in)
	if err != nil {
		return "", err
	}
	var total int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_audit_log WHERE target_type='user' AND target_id=$1`, in.UserID).Scan(&total); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT action,actor_type,actor_id,params,result,created_at
		FROM jazebe_audit_log WHERE target_type='user' AND target_id=$1
		ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, in.UserID, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var action, actorType, result string
		var actorID sql.NullString
		var params []byte
		var createdAt time.Time
		if err = rows.Scan(&action, &actorType, &actorID, &params, &result, &createdAt); err != nil {
			return "", err
		}
		var actorValue any
		if actorID.Valid {
			actorValue = actorID.String
		}
		var paramValue any
		if len(params) > 0 {
			paramValue = json.RawMessage(params)
		}
		items = append(items, map[string]any{"action": action, "actorType": actorType, "actorId": actorValue, "params": paramValue, "result": result, "createdAt": createdAt})
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}
