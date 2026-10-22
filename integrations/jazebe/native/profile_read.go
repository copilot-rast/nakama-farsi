package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type profileReadInput struct {
	PlanetID string `json:"planetId"`
	UserID   string `json:"userId"`
}

type profileLevel struct {
	Level   int    `json:"level"`
	Key     string `json:"key"`
	TitleFa string `json:"titleFa"`
	MinXp   int64  `json:"minXp"`
}

func profileLevelFor(xp int64, raw []byte) (int, string, int64) {
	levels := []profileLevel{{Level: 1, Key: "newcomer", TitleFa: "تازه‌وارد", MinXp: 0}, {Level: 2, Key: "active", TitleFa: "عضو فعال", MinXp: 0}, {Level: 3, Key: "veteran", TitleFa: "کهنه‌کار", MinXp: 0}}
	var configured []profileLevel
	if json.Unmarshal(raw, &configured) == nil && len(configured) == 3 && configured[0].Level == 1 && configured[1].Level == 2 && configured[2].Level == 3 {
		levels = configured
	}
	reached := levels[0]
	for i := 1; i < 3; i++ {
		if levels[i].MinXp <= levels[i-1].MinXp || xp < levels[i].MinXp {
			break
		}
		reached = levels[i]
	}
	next := xp
	if reached.Level < 3 && levels[reached.Level].MinXp > reached.MinXp {
		next = levels[reached.Level].MinXp
	}
	if reached.TitleFa == "" {
		reached.TitleFa = []string{"تازه‌وارد", "عضو فعال", "کهنه‌کار"}[reached.Level-1]
	}
	return reached.Level, reached.TitleFa, next
}

func profileProgress(ctx context.Context, db *sql.DB, planetID, user string) (map[string]any, []map[string]any, error) {
	var name string
	var xp int64
	var thresholds []byte
	err := db.QueryRowContext(ctx, `SELECT p.name,COALESCE(s.total_xp,0),ps.xp_level_thresholds
		FROM jazebe_planet p JOIN jazebe_planet_settings ps ON ps.planet_id=p.id
		LEFT JOIN jazebe_user_planet_stats s ON s.planet_id=p.id AND s.user_id=$2 WHERE p.id=$1`, planetID, user).Scan(&name, &xp, &thresholds)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return nil, nil, err
	}
	rows, err := db.QueryContext(ctx, `SELECT l.name,l.icon,a.acquired_at FROM jazebe_planet_medal pm
		JOIN jazebe_medal_library l ON l.id=pm.medal_library_id
		LEFT JOIN jazebe_medal_acquisition a ON a.planet_medal_id=pm.id AND a.user_id=$2
		WHERE pm.planet_id=$1 AND (pm.status='active' OR a.id IS NOT NULL)
		ORDER BY pm.created_at,pm.id`, planetID, user)
	if err != nil {
		return nil, nil, err
	}
	medals := []map[string]any{}
	earned := 0
	for rows.Next() {
		var label string
		var icon sql.NullString
		var acquired sql.NullTime
		if err = rows.Scan(&label, &icon, &acquired); err != nil {
			break
		}
		glyph := "star"
		if icon.Valid {
			glyph = medalDisplayIcon(icon.String)
		}
		item := map[string]any{"fa": label, "icon": glyph, "earned": acquired.Valid}
		if acquired.Valid {
			earned++
			item["date"] = lotteryDateFa(acquired.Time)
		}
		medals = append(medals, item)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	level, title, xpMax := profileLevelFor(xp, thresholds)
	return map[string]any{"fa": name, "level": level, "levelName": title, "xp": xp, "xpMax": xpMax, "medals": earned}, medals, nil
}

func profileOverview(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileReadInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	var exists bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users u LEFT JOIN jazebe_user_account a ON a.user_id=u.id
		WHERE u.id=$1 AND u.disable_time='1970-01-01 00:00:00 UTC' AND COALESCE(a.state,'active')='active')`, user).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		return "", status.Error(codes.NotFound, "account not found")
	}
	base, err := profileBase(ctx, db, user)
	if err != nil {
		return "", err
	}
	planetID := in.PlanetID
	if planetID != "" {
		planetID, err = requiredUUID(planetID)
		if err != nil {
			return "", err
		}
	} else {
		err = db.QueryRowContext(ctx, `SELECT planet_id FROM jazebe_planet_membership WHERE user_id=$1 AND is_active
			ORDER BY last_activity_at DESC NULLS LAST,joined_at DESC LIMIT 1`, user).Scan(&planetID)
		if errors.Is(err, sql.ErrNoRows) {
			planetID = ""
		} else if err != nil {
			return "", err
		}
	}
	var planet any
	medals := []map[string]any{}
	if planetID != "" {
		p, m, e := profileProgress(ctx, db, planetID, user)
		if e != nil {
			return "", e
		}
		planet = p
		medals = m
	}
	name, kind := "", "preset"
	index := 0
	var avatar any
	if base != nil {
		name = base.DisplayName
		kind = base.AvatarKind
		index = base.AvatarIndex
		avatar = base.AvatarAsset
	}
	var verify sql.NullString
	err = db.QueryRowContext(ctx, `SELECT status FROM jazebe_user_verification WHERE user_id=$1`, user).Scan(&verify)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	state := "none"
	if verify.String == "verified" || verify.String == "review" {
		state = verify.String
	}
	var pending bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_lottery_winner_share WHERE user_id=$1 AND blocked_by_kyc)`, user).Scan(&pending); err != nil {
		return "", err
	}
	// Nakama's native user has no trusted phone-number column. Never present a
	// mutable metadata value as a verified phone or expose private identifiers.
	return encode(map[string]any{"user": map[string]any{"name": name, "phone": "", "avatarIndex": index, "avatarKind": kind, "avatarUrl": avatar},
		"planet": planet, "medals": medals, "verifyStatus": state, "verifyRewardPending": pending})
}

func profilePublic(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	viewer, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileReadInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	target, err := requiredUUID(in.UserID)
	if err != nil {
		return "", err
	}
	var allowed bool
	err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_brand b ON b.id=p.brand_id
		LEFT JOIN jazebe_planet_membership vm ON vm.planet_id=p.id AND vm.user_id=$2
		JOIN users vu ON vu.id=$2 AND vu.disable_time='1970-01-01 00:00:00 UTC'
		LEFT JOIN jazebe_user_account va ON va.user_id=vu.id
		JOIN users u ON u.id=$3 AND u.disable_time='1970-01-01 00:00:00 UTC'
		LEFT JOIN jazebe_user_account a ON a.user_id=u.id
		WHERE p.id=$1 AND p.status='active' AND b.is_active
		AND (vm.user_id IS NULL OR vm.is_active) AND COALESCE(va.state,'active')='active'
		AND COALESCE(a.state,'active')='active')`, planetID, viewer, target).Scan(&allowed)
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", status.Error(codes.NotFound, "public profile not found")
	}
	base, err := profileBase(ctx, db, target)
	if err != nil {
		return "", err
	}
	if base == nil {
		return "", status.Error(codes.NotFound, "public profile not found")
	}
	progress, wall, err := profileProgress(ctx, db, planetID, target)
	if err != nil {
		return "", err
	}
	earned := []map[string]any{}
	for _, item := range wall {
		if item["earned"] == true {
			earned = append(earned, map[string]any{"fa": item["fa"], "icon": item["icon"]})
		}
	}
	return encode(map[string]any{"name": base.DisplayName, "avatarIndex": base.AvatarIndex, "avatarKind": base.AvatarKind,
		"avatarUrl": base.AvatarAsset, "planetFa": progress["fa"], "level": progress["level"],
		"levelName": progress["levelName"], "medals": earned})
}
