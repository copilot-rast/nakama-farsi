package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type missionScanner interface{ Scan(...any) error }

func cheaperFallbackMission(current, candidate map[string]any) map[string]any {
	reward := candidate["reward"].(map[string]any)
	points := reward["points"].(int)
	if candidate["type"] != "solo" || points <= 0 || reward["tickets"].(int) != 0 {
		return current
	}
	if current == nil || points < current["reward"].(map[string]any)["points"].(int) {
		return candidate
	}
	return current
}

func scanMissionView(row missionScanner, planetID string, viewer missionViewer) (map[string]any, bool, error) {
	var id, title, description, kind, detection, state, repeat, period, target string
	var points, xp, tickets int
	var start, end time.Time
	var config, display []byte
	if err := row.Scan(&id, &title, &description, &kind, &points, &xp, &tickets, &target, &start, &end, &detection, &config, &repeat, &period, &state, &display); err != nil {
		return nil, false, err
	}
	visible, err := missionDisplaySatisfied(display, viewer)
	if err != nil {
		return nil, false, err
	}
	if state == "approved" {
		state = "done"
	} else if state == "pending_review" {
		state = "pending"
	} else if state == "rejected" {
		state = "rejected"
	} else {
		state = "doable"
	}
	var cfg map[string]any
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, false, err
	}
	viewKind := "code"
	if kind == "media_view" {
		viewKind = "media"
	} else if kind == "content_creation" {
		viewKind = "content"
	}
	return map[string]any{"id": id, "planetId": planetID, "kind": viewKind, "type": target, "repeat": repeat, "repeatPeriod": period,
		"title": title, "desc": description, "reward": map[string]any{"points": points, "tickets": tickets, "xp": xp},
		"startsAt": start, "endsAt": end, "detectionMethod": detection, "config": cfg, "state": state}, visible, nil
}

// The native catalog includes only action types that this module can execute.
// Unsupported source mission types are never advertised as actionable.
func missionList(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input missionListInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	if input.Repeat != "" && input.Repeat != "all" && input.Repeat != "daily" && input.Repeat != "weekly" &&
		input.Repeat != "monthly" && input.Repeat != "once" && input.Repeat != "unlimited" {
		return "", status.Error(codes.InvalidArgument, "invalid mission repeat filter")
	}
	if input.Type != "" && input.Type != "all" && input.Type != "solo" && input.Type != "group" {
		return "", status.Error(codes.InvalidArgument, "invalid mission type filter")
	}
	order := "m.reward_points DESC,m.sort_order ASC,m.created_at DESC"
	switch input.Sort {
	case "", "reward":
	case "deadline":
		order = "m.ends_at ASC,m.id ASC"
	case "fresh":
		order = "m.created_at DESC,m.id ASC"
	default:
		return "", status.Error(codes.InvalidArgument, "invalid mission sort")
	}
	var active bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE id=$1 AND status='active')`, planetID).Scan(&active); err != nil {
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	if !active {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	viewer, err := missionViewerFor(ctx, db, planetID, user, time.Now().UTC())
	if err != nil {
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT m.id,m.title,COALESCE(m.short_description,''),m.functional_type,m.reward_points,
		m.reward_xp,m.reward_tickets,m.target_type,m.starts_at,m.ends_at,m.detection_method,m.type_config,m.repeat_pattern,COALESCE(m.repeat_period,''),
		COALESCE((SELECT c.status FROM jazebe_mission_completion c WHERE c.mission_id=m.id AND c.user_id=$2
			AND c.deleted_at IS NULL AND c.status IN ('approved','pending_review','rejected')
			AND (m.repeat_pattern='unlimited' AND c.status IN ('pending_review','rejected')
				OR m.repeat_pattern='once' AND c.period_key IS NULL OR m.repeat_pattern='periodic' AND c.period_key =
				CASE m.repeat_period WHEN 'weekly' THEN to_char(now() AT TIME ZONE 'Asia/Tehran','IYYY-"W"IW')
				WHEN 'monthly' THEN to_char(now() AT TIME ZONE 'Asia/Tehran','YYYY-MM')
				ELSE to_char(now() AT TIME ZONE 'Asia/Tehran','YYYY-MM-DD') END)
			ORDER BY CASE c.status WHEN 'approved' THEN 0 WHEN 'pending_review' THEN 1 ELSE 2 END,c.submitted_at DESC LIMIT 1),''),m.display_conditions
		FROM jazebe_mission m WHERE m.planet_id=$1 AND m.status='active'
		AND m.functional_type IN ('media_view','code_entry','content_creation') AND m.starts_at<=now() AND m.ends_at>now()
		ORDER BY `+order, planetID, user)
	if err != nil {
		logger.Error("Mission catalog: %v", err)
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	var fallback map[string]any
	for rows.Next() {
		view, visible, err := scanMissionView(rows, planetID, viewer)
		if err != nil {
			return "", status.Error(codes.Internal, "mission catalog unavailable")
		}
		if visible {
			items = append(items, view)
		} else {
			// Jazebe lists the cheapest solo point mission if all display
			// conditions hide the catalog. Direct action still checks visibility.
			fallback = cheaperFallbackMission(fallback, view)
		}
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	if len(items) == 0 && fallback != nil {
		items = append(items, fallback)
	}
	if input.Repeat != "" && input.Repeat != "all" || input.Type != "" && input.Type != "all" {
		filtered := make([]map[string]any, 0, len(items))
		for _, item := range items {
			repeat := item["repeat"].(string)
			if repeat == "periodic" {
				repeat = item["repeatPeriod"].(string)
			}
			if input.Repeat != "" && input.Repeat != "all" && repeat != input.Repeat {
				continue
			}
			if input.Type != "" && input.Type != "all" && item["type"] != input.Type {
				continue
			}
			filtered = append(filtered, item)
		}
		items = filtered
	}
	return encode(map[string]any{"missions": items})
}

func missionDetail(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readMissionRef(payload)
	if err != nil {
		return "", err
	}
	var active bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE id=$1 AND status='active')`, ref.PlanetID).Scan(&active); err != nil {
		return "", status.Error(codes.Internal, "mission unavailable")
	}
	if !active {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	viewer, err := missionViewerFor(ctx, db, ref.PlanetID, user, time.Now().UTC())
	if err != nil {
		return "", status.Error(codes.Internal, "mission unavailable")
	}
	view, visible, err := scanMissionView(db.QueryRowContext(ctx, `SELECT m.id,m.title,COALESCE(m.short_description,''),m.functional_type,m.reward_points,
		m.reward_xp,m.reward_tickets,m.target_type,m.starts_at,m.ends_at,m.detection_method,m.type_config,m.repeat_pattern,COALESCE(m.repeat_period,''),
		COALESCE((SELECT c.status FROM jazebe_mission_completion c WHERE c.mission_id=m.id AND c.user_id=$3
			AND c.deleted_at IS NULL AND c.status IN ('approved','pending_review','rejected')
			AND (m.repeat_pattern='unlimited' AND c.status IN ('pending_review','rejected')
				OR m.repeat_pattern='once' AND c.period_key IS NULL OR m.repeat_pattern='periodic' AND c.period_key =
				CASE m.repeat_period WHEN 'weekly' THEN to_char(now() AT TIME ZONE 'Asia/Tehran','IYYY-"W"IW')
				WHEN 'monthly' THEN to_char(now() AT TIME ZONE 'Asia/Tehran','YYYY-MM')
				ELSE to_char(now() AT TIME ZONE 'Asia/Tehran','YYYY-MM-DD') END)
			ORDER BY CASE c.status WHEN 'approved' THEN 0 WHEN 'pending_review' THEN 1 ELSE 2 END,c.submitted_at DESC LIMIT 1),''),m.display_conditions
		FROM jazebe_mission m WHERE m.planet_id=$1 AND m.id=$2 AND m.status NOT IN ('draft','suspended')
		AND m.functional_type IN ('media_view','code_entry','content_creation')`, ref.PlanetID, ref.MissionID, user), ref.PlanetID, viewer)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "mission not found")
	}
	if err != nil {
		logger.Error("Mission detail: %v", err)
		return "", status.Error(codes.Internal, "mission unavailable")
	}
	if !visible {
		return "", status.Error(codes.NotFound, "mission not found")
	}
	return encode(view)
}

type missionListInput struct {
	PlanetID string `json:"planetId"`
	Repeat   string `json:"repeat,omitempty"`
	Type     string `json:"type,omitempty"`
	Sort     string `json:"sort,omitempty"`
}

func mediaMissionList(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input missionListInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var active bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE id=$1 AND status='active')`, planetID).Scan(&active); err != nil {
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	if !active {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	viewer, err := missionViewerFor(ctx, db, planetID, user, time.Now().UTC())
	if err != nil {
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT m.id,m.title,COALESCE(m.short_description,''),m.reward_points,m.reward_xp,
		m.starts_at,m.ends_at,m.detection_method,m.type_config,m.repeat_pattern,COALESCE(m.repeat_period,''),
		COALESCE((SELECT c.status FROM jazebe_mission_completion c
			WHERE c.mission_id=m.id AND c.user_id=$2 AND c.deleted_at IS NULL AND c.status IN ('approved','pending_review','rejected')
			AND (m.repeat_pattern='unlimited' AND c.status IN ('pending_review','rejected')
				OR m.repeat_pattern='once' AND c.period_key IS NULL OR m.repeat_pattern='periodic' AND c.period_key =
				CASE m.repeat_period WHEN 'weekly' THEN to_char(now() AT TIME ZONE 'Asia/Tehran','IYYY-"W"IW')
				WHEN 'monthly' THEN to_char(now() AT TIME ZONE 'Asia/Tehran','YYYY-MM')
				ELSE to_char(now() AT TIME ZONE 'Asia/Tehran','YYYY-MM-DD') END)
			ORDER BY CASE c.status WHEN 'approved' THEN 0 WHEN 'pending_review' THEN 1 ELSE 2 END,c.submitted_at DESC LIMIT 1),''),m.display_conditions
		FROM jazebe_mission m WHERE m.planet_id=$1 AND m.status='active' AND m.functional_type='media_view'
			AND m.starts_at<=now() AND m.ends_at>now()
		ORDER BY m.reward_points DESC,m.sort_order ASC,m.created_at DESC LIMIT 100`, planetID, user)
	if err != nil {
		logger.Error("Mission list: %v", err)
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, title, description, detection, state, repeat, period string
		var points, xp int
		var start, end time.Time
		var config, display []byte
		if err := rows.Scan(&id, &title, &description, &points, &xp, &start, &end, &detection, &config, &repeat, &period, &state, &display); err != nil {
			return "", status.Error(codes.Internal, "mission catalog unavailable")
		}
		visible, err := missionDisplaySatisfied(display, viewer)
		if err != nil {
			return "", status.Error(codes.Internal, "mission catalog unavailable")
		}
		if !visible {
			continue
		}
		need, err := mediaWatchSeconds(config)
		if err != nil {
			continue
		}
		if state == "approved" {
			state = "done"
		} else if state == "pending_review" {
			state = "pending"
		} else if state == "rejected" {
			state = "rejected"
		} else {
			state = "doable"
		}
		items = append(items, map[string]any{"id": id, "planetId": planetID, "kind": "media", "type": "solo", "repeat": repeat, "repeatPeriod": period,
			"title": title, "desc": description, "reward": map[string]any{"points": points, "xp": xp},
			"startsAt": start, "endsAt": end, "detectionMethod": detection, "minWatchSec": need, "state": state})
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "mission catalog unavailable")
	}
	return encode(map[string]any{"missions": items})
}

func mediaMissionHistory(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input missionListInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE id=$1 AND status='active')`, planetID).Scan(&exists); err != nil {
		return "", status.Error(codes.Internal, "mission history unavailable")
	}
	if !exists {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	rows, err := db.QueryContext(ctx, `SELECT c.id,c.mission_id,m.title,c.status,c.awarded_points,c.awarded_xp,
		COALESCE(c.review_note,''),c.submitted_at FROM jazebe_mission_completion c
		JOIN jazebe_mission m ON m.id=c.mission_id WHERE c.planet_id=$1 AND c.user_id=$2
		AND c.deleted_at IS NULL AND c.status IN ('approved','pending_review','rejected')
		ORDER BY c.submitted_at DESC LIMIT 100`, planetID, user)
	if err != nil {
		logger.Error("Mission history: %v", err)
		return "", status.Error(codes.Internal, "mission history unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, missionID, title, state, reason string
		var points, xp int
		var at time.Time
		if err := rows.Scan(&id, &missionID, &title, &state, &points, &xp, &reason, &at); err != nil {
			return "", status.Error(codes.Internal, "mission history unavailable")
		}
		item := map[string]any{"id": id, "missionId": missionID, "title": title, "status": state, "awardedPoints": points, "awardedXp": xp, "submittedAt": at}
		if reason != "" && state == "rejected" {
			item["reason"] = reason
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "mission history unavailable")
	}
	return encode(map[string]any{"history": items})
}

// The completion identifier is always resolved through the caller's Nakama
// identity and active planet; knowing another member's UUID grants no access.
func missionHistoryItem(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID     string `json:"planetId"`
		CompletionID string `json:"completionId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	completionID, err := requiredUUID(input.CompletionID)
	if err != nil {
		return "", err
	}
	var active bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE id=$1 AND status='active')`, planetID).Scan(&active); err != nil {
		return "", status.Error(codes.Internal, "mission history unavailable")
	}
	if !active {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	var id, missionID, title, state, reason string
	var points, xp int
	var at time.Time
	err = db.QueryRowContext(ctx, `SELECT c.id,c.mission_id,m.title,c.status,c.awarded_points,c.awarded_xp,
		COALESCE(c.review_note,''),c.submitted_at FROM jazebe_mission_completion c
		JOIN jazebe_mission m ON m.id=c.mission_id WHERE c.id=$1 AND c.planet_id=$2 AND c.user_id=$3
		AND c.deleted_at IS NULL AND c.status IN ('approved','pending_review','rejected')`, completionID, planetID, user).
		Scan(&id, &missionID, &title, &state, &points, &xp, &reason, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "completion not found")
	}
	if err != nil {
		logger.Error("Mission history item: %v", err)
		return "", status.Error(codes.Internal, "mission history unavailable")
	}
	item := map[string]any{"id": id, "missionId": missionID, "title": title, "status": state,
		"awardedPoints": points, "awardedXp": xp, "submittedAt": at}
	if reason != "" && state == "rejected" {
		item["reason"] = reason
	}
	return encode(item)
}

func mediaMissionWithdraw(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID     string `json:"planetId"`
		CompletionID string `json:"completionId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	completionID, err := requiredUUID(input.CompletionID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "withdrawal failed")
	}
	defer tx.Rollback()
	var state string
	err = tx.QueryRowContext(ctx, `SELECT status FROM jazebe_mission_completion WHERE id=$1 AND planet_id=$2 AND user_id=$3 AND deleted_at IS NULL FOR UPDATE`, completionID, planetID, user).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "completion not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "withdrawal failed")
	}
	if state != "pending_review" {
		return "", status.Error(codes.FailedPrecondition, "submission is not pending")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission_completion SET deleted_at=now() WHERE id=$1`, completionID); err != nil {
		return "", status.Error(codes.Internal, "withdrawal failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "withdrawal failed")
	}
	return encode(map[string]any{"completionId": completionID, "state": "doable"})
}
