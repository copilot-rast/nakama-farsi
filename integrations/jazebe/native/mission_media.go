package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mediaMissionDraft struct {
	PlanetID          string                    `json:"planetId"`
	Title             string                    `json:"title"`
	ShortDescription  string                    `json:"shortDescription"`
	DetectionMethod   string                    `json:"detectionMethod"`
	RepeatPattern     string                    `json:"repeatPattern"`
	RepeatPeriod      string                    `json:"repeatPeriod"`
	RewardPoints      int                       `json:"rewardPoints"`
	RewardTickets     int                       `json:"rewardTickets"`
	RewardXp          int                       `json:"rewardXp"`
	TargetType        string                    `json:"targetType"`
	GroupTargetValue  int                       `json:"groupTargetValue"`
	TargetLotteryID   string                    `json:"targetLotteryId"`
	DisplayConditions *missionDisplayConditions `json:"displayConditions"`
	StartsAt          string                    `json:"startsAt"`
	EndsAt            string                    `json:"endsAt"`
	Media             struct {
		AssetURL    string `json:"assetUrl"`
		TotalSec    int    `json:"totalSec"`
		MinWatchSec int    `json:"minWatchSec"`
		Skippable   bool   `json:"skippable"`
	} `json:"media"`
}

func validateMediaMissionDraft(m mediaMissionDraft) (time.Time, time.Time, error) {
	if _, err := supportedMissionRepeat(m.RepeatPattern, m.RepeatPeriod); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if strings.TrimSpace(m.Title) == "" || !utf8.ValidString(m.Title) || utf8.RuneCountInString(m.Title) > 128 ||
		!utf8.ValidString(m.ShortDescription) || utf8.RuneCountInString(m.ShortDescription) > 512 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission title or description")
	}
	if m.DetectionMethod != "platform_auto" && m.DetectionMethod != "manual_review" {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid media detection method")
	}
	if m.RewardPoints < 0 || m.RewardPoints > 1_000_000 || m.RewardXp < 0 || m.RewardXp > 1_000_000 || m.RewardTickets < 0 || m.RewardTickets > 1_000_000 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission reward")
	}
	if m.TargetType != "" && m.TargetType != "solo" && m.TargetType != "group" || m.TargetType == "group" && (m.RewardPoints > 0 || m.GroupTargetValue < 1 || m.GroupTargetValue > 1_000_000) {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid group mission configuration")
	}
	if m.RewardTickets > 0 && m.TargetLotteryID == "" {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "ticket reward requires lottery")
	}
	if strings.TrimSpace(m.Media.AssetURL) == "" || len(m.Media.AssetURL) > 2048 ||
		m.Media.TotalSec < 1 || m.Media.TotalSec > 86400 || m.Media.MinWatchSec < 1 || m.Media.MinWatchSec > m.Media.TotalSec {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid media watch configuration")
	}
	start, err := time.Parse(time.RFC3339, m.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission start")
	}
	end, err := time.Parse(time.RFC3339, m.EndsAt)
	if err != nil || !end.After(start) {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission end")
	}
	return start, end, nil
}

// This is the media_view definition slice of BrandMission_create. It creates a
// draft only; activation requires a working action and reward service.
func mediaMissionDraftCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input mediaMissionDraft
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	start, end, err := validateMediaMissionDraft(input)
	if err != nil {
		return "", err
	}
	if input.RepeatPattern == "" {
		input.RepeatPattern = "once"
	}
	if input.TargetType == "" {
		input.TargetType = "solo"
	}
	if input.TargetLotteryID != "" {
		input.TargetLotteryID, err = requiredUUID(input.TargetLotteryID)
		if err != nil {
			return "", err
		}
	}
	display, err := validateMissionDisplayConditions(input.DisplayConditions)
	if err != nil {
		return "", err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return "", status.Error(codes.Internal, "mission ID generation failed")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "mission draft creation failed")
	}
	defer tx.Rollback()
	var pointMin, pointMax, ticketMax int
	err = tx.QueryRowContext(ctx, `SELECT s.mission_reward_point_min,s.mission_reward_point_max,s.mission_reward_ticket_max
		FROM jazebe_planet_settings s JOIN jazebe_planet p ON p.id=s.planet_id
		JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id AND a.user_id=$2 AND a.is_active
		WHERE p.id=$1`,
		planetID, ownerID).Scan(&pointMin, &pointMax, &ticketMax)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		logger.Error("Jazebe mission settings lookup failed: %v", err)
		return "", status.Error(codes.Internal, "mission draft creation failed")
	}
	if input.RewardPoints > 0 && (input.RewardPoints < pointMin || input.RewardPoints > pointMax) {
		return "", status.Error(codes.FailedPrecondition, "mission points outside planet policy")
	}
	if input.RewardTickets > ticketMax {
		return "", status.Error(codes.FailedPrecondition, "mission tickets outside planet policy")
	}
	if err := verifyMissionPrerequisites(ctx, tx, planetID, input.DisplayConditions); err != nil {
		return "", err
	}
	input.Media.AssetURL = strings.TrimSpace(input.Media.AssetURL)
	config, err := json.Marshal(input.Media)
	if err != nil {
		return "", status.Error(codes.Internal, "mission config encoding failed")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission
		(id,planet_id,title,short_description,functional_type,target_type,repeat_pattern,repeat_period,
		detection_method,reward_points,reward_xp,type_config,starts_at,ends_at,status,created_by,display_conditions,reward_tickets,target_lottery_id,group_target_value)
		VALUES ($1,$2,$3,$4,'media_view',$5,$6,$7,$8,$9,$10,$11,$12,$13,'draft',$14,$15::jsonb,$16,$17,$18)`,
		id.String(), planetID, strings.TrimSpace(input.Title), strings.TrimSpace(input.ShortDescription),
		input.TargetType, input.RepeatPattern, sql.NullString{String: input.RepeatPeriod, Valid: input.RepeatPeriod != ""},
		input.DetectionMethod, input.RewardPoints, input.RewardXp, string(config), start, end, ownerID, display, input.RewardTickets, sql.NullString{String: input.TargetLotteryID, Valid: input.TargetLotteryID != ""}, sql.NullInt64{Int64: int64(input.GroupTargetValue), Valid: input.TargetType == "group"})
	if err != nil {
		logger.Error("Jazebe mission draft insert failed: %v", err)
		return "", status.Error(codes.Internal, "mission draft creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log
		(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.create','brand_admin',$1,'mission',$2,$3,jsonb_build_object('functionalType','media_view','status','draft'))`,
		ownerID, id.String(), planetID); err != nil {
		logger.Error("Jazebe mission audit insert failed: %v", err)
		return "", status.Error(codes.Internal, "mission draft creation failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "mission draft creation failed")
	}
	return encode(map[string]any{"id": id.String(), "planetId": planetID, "status": "draft", "functionalType": "media_view"})
}
