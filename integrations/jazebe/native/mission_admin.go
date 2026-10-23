package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type adminMissionInput struct {
	MissionID string `json:"missionId"`
	Reason    string `json:"reason"`
}

func validateAdminMissionInput(input adminMissionInput, requireReason bool) (adminMissionInput, error) {
	var err error
	input.MissionID, err = requiredUUID(input.MissionID)
	if err != nil {
		return input, err
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) > 512 || (requireReason && input.Reason == "") {
		return input, status.Error(codes.InvalidArgument, "reason must contain at most 512 characters and is required for suspension")
	}
	return input, nil
}

func missionOperator(ctx context.Context) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "Jazebe operator role required")
	}
	return actor, nil
}

// Operator suspension hides a mission from members. Unsuspension leaves it
// paused so brand administrators can review it before activating it again.
func missionAdminTransition(ctx context.Context, logger runtime.Logger, db *sql.DB, payload string, suspend bool) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	var input adminMissionInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input, err = validateAdminMissionInput(input, suspend)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "mission moderation failed")
	}
	defer tx.Rollback()
	var planetID, title, current string
	err = tx.QueryRowContext(ctx, `SELECT planet_id,title,status FROM jazebe_mission WHERE id=$1 FOR UPDATE`, input.MissionID).Scan(&planetID, &title, &current)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "mission not found")
	}
	if err != nil {
		logger.Error("Mission moderation lookup: %v", err)
		return "", status.Error(codes.Internal, "mission moderation failed")
	}
	to, action, subject := "suspended", "mission.suspend", "Jazebe mission suspended"
	if suspend {
		if current == "suspended" {
			return "", status.Error(codes.FailedPrecondition, "mission already suspended")
		}
	} else {
		if current != "suspended" {
			return "", status.Error(codes.FailedPrecondition, "mission is not suspended")
		}
		to, action, subject = "paused", "mission.unsuspend", "Jazebe mission suspension lifted"
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission SET status=$2 WHERE id=$1`, input.MissionID, to); err != nil {
		return "", status.Error(codes.Internal, "mission moderation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ($1,'operator',$2,'mission',$3,$4,jsonb_build_object('reason',$5::text,'previousStatus',$6::text,'title',$7::text))`,
		action, actor, input.MissionID, planetID, input.Reason, current, title); err != nil {
		return "", status.Error(codes.Internal, "mission moderation audit failed")
	}
	rows, err := tx.QueryContext(ctx, `SELECT a.user_id FROM jazebe_brand_admin a JOIN jazebe_planet p ON p.brand_id=a.brand_id
		WHERE p.id=$1 AND a.is_active`, planetID)
	if err != nil {
		return "", status.Error(codes.Internal, "mission moderation notification failed")
	}
	admins := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "mission moderation notification failed")
		}
		admins = append(admins, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "mission moderation notification failed")
	}
	for _, id := range admins {
		if err := persistJazebeNotification(ctx, tx, id, action+":"+input.MissionID, subject,
			map[string]any{"missionId": input.MissionID, "planetId": planetID, "title": title, "reason": input.Reason, "status": to}, time.Now().UTC()); err != nil {
			return "", status.Error(codes.Internal, "mission moderation notification failed")
		}
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "mission moderation failed")
	}
	return encode(map[string]any{"missionId": input.MissionID, "planetId": planetID, "status": to, "previousStatus": current})
}

func missionAdminSuspend(ctx context.Context, l runtime.Logger, db *sql.DB, _ runtime.NakamaModule, p string) (string, error) {
	return missionAdminTransition(ctx, l, db, p, true)
}

func missionAdminUnsuspend(ctx context.Context, l runtime.Logger, db *sql.DB, _ runtime.NakamaModule, p string) (string, error) {
	return missionAdminTransition(ctx, l, db, p, false)
}

func missionAdminDetail(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var input adminMissionInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	id, err := requiredUUID(input.MissionID)
	if err != nil {
		return "", err
	}
	var planetID, title, kind, target, state string
	var points, tickets, xp int
	var start, end, created time.Time
	err = db.QueryRowContext(ctx, `SELECT planet_id,title,functional_type,target_type,status,reward_points,reward_tickets,reward_xp,
		starts_at,ends_at,created_at FROM jazebe_mission WHERE id=$1`, id).Scan(&planetID, &title, &kind, &target, &state,
		&points, &tickets, &xp, &start, &end, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "mission not found")
	}
	if err != nil {
		logger.Error("Operator mission detail: %v", err)
		return "", status.Error(codes.Internal, "mission unavailable")
	}
	return encode(map[string]any{"id": id, "planetId": planetID, "title": title, "functionalType": kind,
		"targetType": target, "status": state, "rewardPoints": points, "rewardTickets": tickets, "rewardXp": xp,
		"startsAt": start, "endsAt": end, "createdAt": created})
}

func missionAdminList(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		Status   string `json:"status"`
		Search   string `json:"search"`
		Page     int    `json:"page"`
		Limit    int    `json:"limit"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	if input.PlanetID != "" {
		var err error
		input.PlanetID, err = requiredUUID(input.PlanetID)
		if err != nil {
			return "", err
		}
	}
	if input.Status != "" && input.Status != "draft" && input.Status != "active" && input.Status != "paused" && input.Status != "expired" && input.Status != "suspended" {
		return "", status.Error(codes.InvalidArgument, "invalid mission status")
	}
	input.Search = strings.TrimSpace(input.Search)
	if !utf8.ValidString(input.Search) || utf8.RuneCountInString(input.Search) > 128 {
		return "", status.Error(codes.InvalidArgument, "invalid search")
	}
	if input.Page == 0 {
		input.Page = 1
	}
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Page < 1 || input.Page > 10000 || input.Limit < 1 || input.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid pagination")
	}
	search := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(input.Search)
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_mission
		WHERE ($1='' OR planet_id::text=$1) AND ($2='' OR status=$2)
		AND ($3='' OR title ILIKE '%' || $3 || '%' ESCAPE '\')`, input.PlanetID, input.Status, search).Scan(&total); err != nil {
		logger.Error("Operator mission count: %v", err)
		return "", status.Error(codes.Internal, "mission list unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT id,planet_id,title,functional_type,target_type,status,reward_points,reward_tickets,
		starts_at,ends_at,created_at FROM jazebe_mission
		WHERE ($1='' OR planet_id::text=$1) AND ($2='' OR status=$2)
		AND ($3='' OR title ILIKE '%' || $3 || '%' ESCAPE '\')
		ORDER BY created_at DESC,id DESC LIMIT $4 OFFSET $5`, input.PlanetID, input.Status, search, input.Limit, (input.Page-1)*input.Limit)
	if err != nil {
		logger.Error("Operator mission list: %v", err)
		return "", status.Error(codes.Internal, "mission list unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, planetID, title, kind, target, state string
		var points, tickets int
		var start, end, created time.Time
		if err := rows.Scan(&id, &planetID, &title, &kind, &target, &state, &points, &tickets, &start, &end, &created); err != nil {
			return "", status.Error(codes.Internal, "mission list unavailable")
		}
		items = append(items, map[string]any{"id": id, "planetId": planetID, "title": title,
			"functionalType": kind, "targetType": target, "status": state, "rewardPoints": points,
			"rewardTickets": tickets, "startsAt": start, "endsAt": end, "createdAt": created})
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "mission list unavailable")
	}
	return encode(map[string]any{"missions": items, "total": total, "page": input.Page, "limit": input.Limit})
}
