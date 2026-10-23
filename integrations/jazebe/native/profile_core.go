package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type profileInput struct {
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	AvatarIndex   *int   `json:"avatarIndex"`
	AvatarMediaID string `json:"avatarMediaId"`
	InviteToken   string `json:"inviteToken"`
	Version       string `json:"version"`
}

type baseProfile struct {
	DisplayName string `json:"displayName"`
	AvatarIndex int    `json:"avatarIndex"`
	AvatarKind  string `json:"avatarKind"`
	AvatarAsset any    `json:"avatarAsset"`
}

func validProfileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 20 || strings.ContainsAny(name, "\x00\r\n") {
		return "", status.Error(codes.InvalidArgument, "display name must contain 2 to 20 characters")
	}
	return name, nil
}

func validAvatar(index int) error {
	if index < -1 || index > 5 {
		return status.Error(codes.InvalidArgument, "invalid avatar index")
	}
	return nil
}

func persianMonthKey(at time.Time) string {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		loc = time.FixedZone("IR", 3*3600+1800)
	}
	date := at.In(loc)
	y, m, _ := gregorianToJalali(date.Year(), int(date.Month()), date.Day())
	return fmt.Sprintf("%04d-%02d", y, m)
}

func profileBase(ctx context.Context, db *sql.DB, user string) (*baseProfile, error) {
	var name, kind string
	var preset sql.NullInt64
	var media sql.NullString
	err := db.QueryRowContext(ctx, `SELECT display_name,avatar_kind,avatar_preset_index,avatar_media_id
		FROM jazebe_user_profile WHERE user_id=$1`, user).Scan(&name, &kind, &preset, &media)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	index := 0
	if kind == "upload" {
		index = -1
	} else if preset.Valid {
		index = int(preset.Int64)
	}
	var asset any
	if media.Valid {
		asset = "/v2/jazebe/media/" + media.String
	}
	return &baseProfile{DisplayName: name, AvatarKind: kind, AvatarIndex: index, AvatarAsset: asset}, nil
}

func profileBootstrap(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND disable_time='1970-01-01 00:00:00 UTC')`, user).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		return "", status.Error(codes.NotFound, "active user not found")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_account(user_id,last_login_at) VALUES($1,now()) ON CONFLICT(user_id) DO UPDATE SET last_login_at=now(),updated_at=now()`, user); err != nil {
		return "", err
	}
	var state, terms string
	if err = tx.QueryRowContext(ctx, `SELECT a.state,COALESCE(a.terms_version,'') FROM jazebe_user_account a WHERE user_id=$1`, user).Scan(&state, &terms); err != nil {
		return "", err
	}
	if state != "active" {
		return "", status.Error(codes.PermissionDenied, "account unavailable")
	}
	var current string
	if err = tx.QueryRowContext(ctx, `SELECT terms_version FROM jazebe_platform_policy WHERE key='default'`).Scan(&current); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	profile, err := profileBase(ctx, db, user)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"profile": profile, "accountState": state, "terms": map[string]any{"current": current, "accepted": terms == current}})
}

func profileTermsAccept(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	if !policyTermsPattern.MatchString(in.Version) {
		return "", status.Error(codes.InvalidArgument, "invalid terms version")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users u LEFT JOIN jazebe_user_account a ON a.user_id=u.id
		WHERE u.id=$1 AND u.disable_time='1970-01-01 00:00:00 UTC' AND COALESCE(a.state,'active')='active')`, user).Scan(&active); err != nil {
		return "", err
	}
	if !active {
		return "", status.Error(codes.PermissionDenied, "account unavailable")
	}
	var current string
	if err = tx.QueryRowContext(ctx, `SELECT terms_version FROM jazebe_platform_policy WHERE key='default'`).Scan(&current); err != nil {
		return "", err
	}
	if current != in.Version {
		return "", status.Error(codes.FailedPrecondition, "terms version changed")
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO jazebe_user_account(user_id,terms_version,terms_accepted_at)
		VALUES($1,$2,now()) ON CONFLICT(user_id) DO UPDATE SET terms_version=$2,terms_accepted_at=now(),updated_at=now()
		WHERE jazebe_user_account.state='active'`, user, current)
	if err != nil {
		return "", err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return "", status.Error(codes.PermissionDenied, "account unavailable")
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return encode(map[string]any{"current": current, "accepted": true})
}

func profileNameAvailable(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	name, err := validProfileName(in.Name)
	if err != nil {
		return "", err
	}
	var taken bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_user_profile WHERE lower(display_name)=lower($1) AND user_id<>$2)
		OR EXISTS(SELECT 1 FROM users WHERE lower(display_name)=lower($1) AND id<>$2)`, name, user).Scan(&taken); err != nil {
		return "", err
	}
	suggestions := []string{}
	if taken {
		base := []rune(name)
		if len(base) > 17 {
			base = base[:17]
		}
		rows, e := db.QueryContext(ctx, `SELECT candidate FROM (
			SELECT $1::text || '_' || lpad(n::text,2,'0') AS candidate FROM generate_series(10,25) n
		) s WHERE NOT EXISTS(SELECT 1 FROM jazebe_user_profile WHERE lower(display_name)=lower(candidate))
			AND NOT EXISTS(SELECT 1 FROM users WHERE lower(display_name)=lower(candidate)) LIMIT 3`, string(base))
		if e != nil {
			return "", e
		}
		for rows.Next() {
			var candidate string
			if e = rows.Scan(&candidate); e != nil {
				break
			}
			suggestions = append(suggestions, candidate)
		}
		if e == nil {
			e = rows.Err()
		}
		rows.Close()
		if e != nil {
			return "", e
		}
	}
	return encode(map[string]any{"name": name, "available": !taken, "suggestions": suggestions})
}

func profileSetup(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return profileSetName(ctx, logger, db, payload, true)
}
func profileRename(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return profileSetName(ctx, nil, db, payload, false)
}

func profileSetName(ctx context.Context, logger runtime.Logger, db *sql.DB, payload string, setup bool) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	requested := in.Name
	if setup {
		requested = in.DisplayName
	}
	name, err := validProfileName(requested)
	if err != nil {
		return "", err
	}
	if setup && (in.AvatarIndex == nil || validAvatar(*in.AvatarIndex) != nil) {
		return "", status.Error(codes.InvalidArgument, "invalid avatar index")
	}
	if in.InviteToken != "" && !referralTokenPattern.MatchString(in.InviteToken) {
		return "", status.Error(codes.InvalidArgument, "invalid invite token")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	// Serialize both same-user updates and competing names across Nakama nodes.
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users u LEFT JOIN jazebe_user_account a ON a.user_id=u.id
		WHERE u.id=$1 AND u.disable_time='1970-01-01 00:00:00 UTC' AND COALESCE(a.state,'active')='active')`, user).Scan(&active); err != nil {
		return "", err
	}
	if !active {
		return "", status.Error(codes.PermissionDenied, "account unavailable")
	}
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "jazebe_profile_user:"+user); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "jazebe_profile_name:"+strings.ToLower(name)); err != nil {
		return "", err
	}
	var old string
	var changed sql.NullTime
	var oldAvatar sql.NullString
	var count int
	err = tx.QueryRowContext(ctx, `SELECT display_name,display_name_changed_at,display_name_change_count_month,avatar_media_id FROM jazebe_user_profile WHERE user_id=$1 FOR UPDATE`, user).Scan(&old, &changed, &count, &oldAvatar)
	first := errors.Is(err, sql.ErrNoRows)
	if err != nil && !first {
		return "", err
	}
	if !setup && first {
		return "", status.Error(codes.NotFound, "profile not found")
	}
	if first || old != name {
		var taken bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_user_profile WHERE lower(display_name)=lower($1) AND user_id<>$2)
			OR EXISTS(SELECT 1 FROM users WHERE lower(display_name)=lower($1) AND id<>$2)`, name, user).Scan(&taken); err != nil {
			return "", err
		}
		if taken {
			return "", status.Error(codes.AlreadyExists, "display name taken")
		}
	}
	now := time.Now().UTC()
	if first {
		index := *in.AvatarIndex
		kind := "preset"
		var preset any = index
		if index == -1 {
			kind = "upload"
			preset = nil
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_profile(user_id,display_name,avatar_kind,avatar_preset_index)
			VALUES($1,$2,$3,$4)`, user, name, kind, preset)
	} else if old != name {
		var limit int
		if err = tx.QueryRowContext(ctx, `SELECT display_name_change_limit_per_month FROM jazebe_platform_policy WHERE key='default'`).Scan(&limit); err != nil {
			return "", err
		}
		if !changed.Valid || persianMonthKey(changed.Time) != persianMonthKey(now) {
			count = 0
		}
		if count >= limit {
			return "", status.Error(codes.ResourceExhausted, "monthly display name change limit reached")
		}
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_profile SET display_name=$2,display_name_changed_at=$3,
			display_name_change_count_month=$4,updated_at=$3 WHERE user_id=$1`, user, name, now, count+1)
	}
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return "", status.Error(codes.AlreadyExists, "display name taken")
		}
		return "", err
	}
	if setup && !first {
		index := *in.AvatarIndex
		kind := "preset"
		var preset any = index
		if index == -1 {
			kind = "upload"
			preset = nil
		}
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_profile SET avatar_kind=$2::text,avatar_preset_index=$3,
			avatar_media_id=CASE WHEN $2::text='preset' THEN NULL ELSE avatar_media_id END,updated_at=now() WHERE user_id=$1`, user, kind, preset)
		if err != nil {
			return "", err
		}
	}
	res, err := tx.ExecContext(ctx, `UPDATE users SET display_name=$2,update_time=now() WHERE id=$1 AND disable_time='1970-01-01 00:00:00 UTC'`, user, name)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return "", status.Error(codes.NotFound, "active user not found")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_account(user_id) VALUES($1) ON CONFLICT(user_id) DO NOTHING`, user); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	if setup && !first && *in.AvatarIndex != -1 && oldAvatar.Valid {
		if _, e := db.ExecContext(ctx, `UPDATE jazebe_media_upload SET status='deleted',deleted_at=now()
			WHERE id=$1 AND owner_user_id=$2 AND purpose='avatar' AND status='complete'
			AND NOT EXISTS(SELECT 1 FROM jazebe_user_profile WHERE avatar_media_id=$1)`, oldAvatar.String, user); e != nil && logger != nil {
			logger.Error("Jazebe prior avatar retirement failed for user %s: %v", user, e)
		}
	}
	// Jazebe emits this only after the first profile creation; listener failure
	// is logged and does not undo the completed sign-up.
	if setup && first && in.InviteToken != "" {
		invitationPayload, _ := json.Marshal(map[string]string{"token": in.InviteToken})
		_, e := referralAccept(context.WithValue(ctx, referralSignupContextKey{}, true), logger, db, nil, string(invitationPayload))
		if e != nil && logger != nil {
			logger.Error("Jazebe signup attribution failed for user %s: %v", user, e)
		}
	}
	returnProfile, err := profileBase(ctx, db, user)
	if err != nil {
		return "", err
	}
	return encode(returnProfile)
}

func profileSetAvatar(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	if in.AvatarIndex == nil || validAvatar(*in.AvatarIndex) != nil {
		return "", status.Error(codes.InvalidArgument, "invalid avatar index")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users u LEFT JOIN jazebe_user_account a ON a.user_id=u.id
		WHERE u.id=$1 AND u.disable_time='1970-01-01 00:00:00 UTC' AND COALESCE(a.state,'active')='active')`, user).Scan(&active); err != nil {
		return "", err
	}
	if !active {
		return "", status.Error(codes.PermissionDenied, "account unavailable")
	}
	var oldMedia sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT avatar_media_id FROM jazebe_user_profile WHERE user_id=$1 FOR UPDATE`, user).Scan(&oldMedia); errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "profile not found")
	}
	if err != nil {
		return "", err
	}
	var media any
	kind := "preset"
	var preset any = *in.AvatarIndex
	if *in.AvatarIndex == -1 {
		kind = "upload"
		preset = nil
		id, e := requiredUUID(in.AvatarMediaID)
		if e != nil {
			return "", e
		}
		var lockedID string
		err = tx.QueryRowContext(ctx, `SELECT id FROM jazebe_media_upload WHERE id=$1 AND owner_user_id=$2
			AND purpose='avatar' AND status='complete' AND deleted_at IS NULL FOR SHARE`, id, user).Scan(&lockedID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.PermissionDenied, "completed owned avatar required")
		}
		if err != nil {
			return "", err
		}
		media = id
	}
	res, err := tx.ExecContext(ctx, `UPDATE jazebe_user_profile SET avatar_kind=$2,avatar_preset_index=$3,avatar_media_id=$4,updated_at=now()
		WHERE user_id=$1`, user, kind, preset, media)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", status.Error(codes.NotFound, "profile not found")
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	if oldMedia.Valid && oldMedia.String != media {
		if _, e := db.ExecContext(ctx, `UPDATE jazebe_media_upload SET status='deleted',deleted_at=now()
			WHERE id=$1 AND owner_user_id=$2 AND purpose='avatar' AND status='complete'
			AND NOT EXISTS(SELECT 1 FROM jazebe_user_profile WHERE avatar_media_id=$1)`, oldMedia.String, user); e != nil && logger != nil {
			logger.Error("Jazebe prior avatar retirement failed for user %s: %v", user, e)
		}
	}
	profile, err := profileBase(ctx, db, user)
	if err != nil {
		return "", err
	}
	return encode(profile)
}
