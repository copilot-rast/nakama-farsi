package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type missionReviewInput struct {
	PlanetID     string `json:"planetId"`
	CompletionID string `json:"completionId"`
	Decision     string `json:"decision"`
	Reason       string `json:"reason"`
}

func validateMissionReview(input missionReviewInput) (missionReviewInput, error) {
	var err error
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return input, err
	}
	input.CompletionID, err = requiredUUID(input.CompletionID)
	if err != nil {
		return input, err
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Decision != "approve" && input.Decision != "reject" {
		return input, status.Error(codes.InvalidArgument, "invalid review decision")
	}
	if !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) > 512 || (input.Decision == "reject" && input.Reason == "") {
		return input, status.Error(codes.InvalidArgument, "rejection requires a reason of at most 512 characters")
	}
	return input, nil
}

// A review decision and its payout are atomic. A second decision cannot move
// points because the pending completion row is locked and then leaves pending.
func mediaMissionReview(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input missionReviewInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input, err = validateMissionReview(input)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "review failed")
	}
	defer tx.Rollback()
	var allowed bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, input.PlanetID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "review authorization failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	var missionID, user, state string
	var groupID sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT mission_id,user_id,status,group_id FROM jazebe_mission_completion
		WHERE id=$1 AND planet_id=$2 AND user_id IS NOT NULL AND deleted_at IS NULL FOR UPDATE`, input.CompletionID, input.PlanetID).Scan(&missionID, &user, &state, &groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "completion not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "review failed")
	}
	if state != "pending_review" {
		return "", status.Error(codes.AlreadyExists, "completion already reviewed")
	}
	ref := missionRef{PlanetID: input.PlanetID, MissionID: missionID}
	m, err := loadMediaMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		m, err = loadCodeMission(ctx, tx, ref)
	}
	if errors.Is(err, sql.ErrNoRows) {
		m, err = loadContentMission(ctx, tx, ref)
	}
	if err != nil {
		logger.Error("Review mission lookup: %v", err)
		return "", status.Error(codes.Internal, "review failed")
	}
	now := time.Now().UTC()
	newState := "approved"
	if input.Decision == "reject" {
		newState = "rejected"
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_mission_completion SET status=$2,reviewer_admin_id=$3,
		review_note=$4,reviewed_at=$5 WHERE id=$1`, input.CompletionID, newState, actor, sql.NullString{String: input.Reason, Valid: input.Reason != ""}, now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "mission already completed for this period")
		}
		return "", status.Error(codes.Internal, "review failed")
	}
	if newState == "approved" {
		if m.TargetType == "group" {
			if !groupID.Valid {
				return "", status.Error(codes.FailedPrecondition, "group completion has no owning group")
			}
			err = advanceGroupMission(ctx, tx, ref, user, groupID.String, input.CompletionID, now)
		} else {
			err = paySoloMission(ctx, tx, ref, user, input.CompletionID, m, now)
		}
		if err != nil {
			logger.Error("Review payout: %v", err)
			return "", status.Error(codes.Internal, "review payout failed")
		}
	}
	if newState == "rejected" {
		if err := persistJazebeNotification(ctx, tx, user, "mission-review:"+input.CompletionID, "ثبت ماموریت رد شد", map[string]any{
			"category": "mission_game", "actionLink": "/missions", "missionId": missionID, "reason": input.Reason}, now); err != nil {
			return "", status.Error(codes.Internal, "review notice failed")
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.review','brand_admin',$1,'mission_completion',$2,$3,jsonb_build_object('decision',$4::text))`, actor, input.CompletionID, input.PlanetID, input.Decision); err != nil {
		return "", status.Error(codes.Internal, "review audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "review failed")
	}
	if newState == "approved" {
		missionMedalProgression(ctx, logger, db, input.PlanetID, user, input.CompletionID)
	}
	return encode(map[string]any{"completionId": input.CompletionID, "status": newState})
}

func missionReviewQueue(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID  string `json:"planetId"`
		MissionID string `json:"missionId"`
		Status    string `json:"status"`
		Page      int    `json:"page"`
		Limit     int    `json:"limit"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var missionID any
	if input.MissionID != "" {
		missionID, err = requiredUUID(input.MissionID)
		if err != nil {
			return "", err
		}
	}
	if input.Status == "" {
		input.Status = "pending_review"
	}
	if input.Status != "pending_review" && input.Status != "approved" && input.Status != "rejected" {
		return "", status.Error(codes.InvalidArgument, "invalid review status")
	}
	if input.Page == 0 {
		input.Page = 1
	}
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Page < 1 || input.Page > 1_000_000 || input.Limit < 1 || input.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid pagination")
	}
	var allowed bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, planetID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "review queue unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if missionID != nil {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission WHERE id=$1 AND planet_id=$2)`, missionID, planetID).Scan(&exists); err != nil {
			return "", status.Error(codes.Internal, "review queue unavailable")
		}
		if !exists {
			return "", status.Error(codes.NotFound, "mission not found")
		}
	}
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_mission_completion c WHERE c.planet_id=$1
		AND c.status=$2 AND c.deleted_at IS NULL AND ($3::uuid IS NULL OR c.mission_id=$3)`, planetID, input.Status, missionID).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "review queue unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT c.id,c.mission_id,m.title,m.functional_type,c.user_id,
		COALESCE(NULLIF(u.display_name,''),u.username),c.submitted_at,c.status,c.evidence,
		c.awarded_points,c.awarded_xp,c.review_note,c.reviewed_at
		FROM jazebe_mission_completion c JOIN jazebe_mission m ON m.id=c.mission_id
		JOIN users u ON u.id=c.user_id WHERE c.planet_id=$1 AND c.status=$2 AND c.deleted_at IS NULL
		AND ($3::uuid IS NULL OR c.mission_id=$3) ORDER BY c.submitted_at ASC LIMIT $4 OFFSET $5`, planetID, input.Status, missionID, input.Limit, (input.Page-1)*input.Limit)
	if err != nil {
		logger.Error("Mission review queue: %v", err)
		return "", status.Error(codes.Internal, "review queue unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, mission, title, kind, user, name, state string
		var submitted time.Time
		var evidence []byte
		var points, xp int
		var note sql.NullString
		var reviewed sql.NullTime
		if err := rows.Scan(&id, &mission, &title, &kind, &user, &name, &submitted, &state, &evidence, &points, &xp, &note, &reviewed); err != nil {
			return "", status.Error(codes.Internal, "review queue unavailable")
		}
		item := map[string]any{"id": id, "missionId": mission, "missionTitle": title, "functionalType": kind, "userId": user,
			"memberName": name, "submittedAt": submitted, "status": state, "reward": map[string]any{"points": points, "tickets": 0, "xp": xp}}
		if len(evidence) > 0 {
			var value any
			if json.Unmarshal(evidence, &value) == nil {
				item["evidence"] = value
			}
		}
		if note.Valid {
			item["reviewNote"] = note.String
		}
		if reviewed.Valid {
			item["reviewedAt"] = reviewed.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "review queue unavailable")
	}
	return encode(map[string]any{"items": items, "total": total})
}
