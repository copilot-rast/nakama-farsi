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

type inviteMissionDraft struct {
	PlanetID          string                    `json:"planetId"`
	Title             string                    `json:"title"`
	ShortDescription  string                    `json:"shortDescription"`
	TargetType        string                    `json:"targetType"`
	GroupTargetValue  int                       `json:"groupTargetValue"`
	RepeatPattern     string                    `json:"repeatPattern"`
	RepeatPeriod      string                    `json:"repeatPeriod"`
	RewardPoints      int                       `json:"rewardPoints"`
	RewardTickets     int                       `json:"rewardTickets"`
	RewardXp          int                       `json:"rewardXp"`
	TargetLotteryID   string                    `json:"targetLotteryId"`
	StartsAt          string                    `json:"startsAt"`
	EndsAt            string                    `json:"endsAt"`
	DisplayConditions *missionDisplayConditions `json:"displayConditions"`
	Invite            referralMissionConfig     `json:"invite"`
}

func validateInviteMissionDraft(in inviteMissionDraft) (time.Time, time.Time, error) {
	if !utf8.ValidString(in.Title) || utf8.RuneCountInString(strings.TrimSpace(in.Title)) < 1 || utf8.RuneCountInString(in.Title) > 128 || utf8.RuneCountInString(in.ShortDescription) > 512 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid invite mission title")
	}
	if in.RepeatPattern == "" {
		in.RepeatPattern = "once"
	}
	if _, err := supportedReviewedRepeat(in.RepeatPattern, in.RepeatPeriod); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if in.TargetType == "" {
		in.TargetType = "solo"
	}
	if in.RewardPoints < 0 || in.RewardPoints > 1000000 || in.RewardXp < 0 || in.RewardXp > 1000000 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid invite mission rewards")
	}
	if err := validateMissionTargetRewards(in.TargetType, in.GroupTargetValue, in.RewardPoints, in.RewardTickets, in.TargetLotteryID); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if in.Invite.Scope != "planet" && in.Invite.Scope != "group" || in.Invite.SuccessCondition != "signup" && in.Invite.SuccessCondition != "first_activity" || in.Invite.TargetCount < 1 || in.Invite.TargetCount > 10000 || in.Invite.WindowDays < 0 || in.Invite.WindowDays > 3650 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid invite mission condition")
	}
	start, err := time.Parse(time.RFC3339, in.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission start")
	}
	end, err := time.Parse(time.RFC3339, in.EndsAt)
	if err != nil || !end.After(start) {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission end")
	}
	return start, end, nil
}

func inviteMissionDraftCreate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in inviteMissionDraft
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if in.RepeatPattern == "" {
		in.RepeatPattern = "once"
	}
	if in.TargetType == "" {
		in.TargetType = "solo"
	}
	start, end, err := validateInviteMissionDraft(in)
	if err != nil {
		return "", err
	}
	if in.TargetLotteryID != "" {
		in.TargetLotteryID, err = requiredUUID(in.TargetLotteryID)
		if err != nil {
			return "", err
		}
	}
	display, err := validateMissionDisplayConditions(in.DisplayConditions)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var min, max, ticketMax int
	err = tx.QueryRowContext(ctx, `SELECT s.mission_reward_point_min,s.mission_reward_point_max,s.mission_reward_ticket_max
		FROM jazebe_planet_settings s JOIN jazebe_planet p ON p.id=s.planet_id
		JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id WHERE p.id=$1 AND a.user_id=$2 AND a.is_active`, planetID, actor).Scan(&min, &max, &ticketMax)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if err != nil {
		return "", err
	}
	if in.RewardPoints > 0 && (in.RewardPoints < min || in.RewardPoints > max) || in.RewardTickets > ticketMax {
		return "", status.Error(codes.FailedPrecondition, "rewards outside planet policy")
	}
	if err = verifyMissionPrerequisites(ctx, tx, planetID, in.DisplayConditions); err != nil {
		return "", err
	}
	cfg, _ := json.Marshal(in.Invite)
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission(id,planet_id,title,short_description,functional_type,target_type,
		repeat_pattern,repeat_period,detection_method,reward_points,reward_tickets,reward_xp,target_lottery_id,
		group_target_value,display_conditions,type_config,starts_at,ends_at,status,created_by)
		VALUES($1,$2,$3,$4,'invite',$5,$6,$7,'platform_auto',$8,$9,$10,$11,$12,$13::jsonb,$14::jsonb,$15,$16,'draft',$17)`,
		id, planetID, strings.TrimSpace(in.Title), strings.TrimSpace(in.ShortDescription), in.TargetType, in.RepeatPattern,
		sql.NullString{String: in.RepeatPeriod, Valid: in.RepeatPeriod != ""}, in.RewardPoints, in.RewardTickets, in.RewardXp,
		sql.NullString{String: in.TargetLotteryID, Valid: in.TargetLotteryID != ""}, sql.NullInt64{Int64: int64(in.GroupTargetValue), Valid: in.TargetType == "group"},
		display, string(cfg), start, end, actor)
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('mission.create','brand_admin',$1,'mission',$2,$3,jsonb_build_object('functionalType','invite','status','draft'))`, actor, id, planetID); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return encode(map[string]any{"id": id, "planetId": planetID, "status": "draft", "functionalType": "invite"})
}

func loadInviteMission(ctx context.Context, tx *sql.Tx, ref missionRef) (mediaMission, error) {
	var m mediaMission
	err := tx.QueryRowContext(ctx, `SELECT m.title,m.status,p.status,m.detection_method,m.repeat_pattern,COALESCE(m.repeat_period,''),
		m.starts_at,m.ends_at,m.reward_points,m.reward_xp,m.type_config,m.display_conditions,m.target_type,m.reward_tickets,m.target_lottery_id
		FROM jazebe_mission m JOIN jazebe_planet p ON p.id=m.planet_id
		WHERE m.id=$1 AND m.planet_id=$2 AND m.functional_type='invite' FOR UPDATE OF m`, ref.MissionID, ref.PlanetID).
		Scan(&m.Title, &m.Status, &m.PlanetStatus, &m.Detection, &m.Repeat, &m.Period, &m.StartsAt, &m.EndsAt, &m.RewardPoints, &m.RewardXp, &m.Config, &m.Display, &m.TargetType, &m.RewardTickets, &m.TargetLotteryID)
	return m, err
}

func inviteMissionActivate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readMissionRef(payload)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	m, err := loadInviteMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "invite mission not found")
	}
	if err != nil {
		return "", err
	}
	var allowed bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, ref.PlanetID, actor).Scan(&allowed); err != nil {
		return "", err
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "brand administrator required")
	}
	if m.Status != "draft" && m.Status != "paused" || !m.EndsAt.After(time.Now().UTC()) || m.Detection != "platform_auto" {
		return "", status.Error(codes.FailedPrecondition, "invite mission cannot activate")
	}
	if _, err = missionPeriodKey(m.Repeat, m.Period, time.Now().UTC()); err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid repeat period")
	}
	cfg, err := parseReferralMissionConfig(m.Config)
	if err != nil || cfg.TargetCount < 1 {
		return "", status.Error(codes.FailedPrecondition, "invite target required")
	}
	if m.RewardTickets > 0 && !m.TargetLotteryID.Valid {
		return "", status.Error(codes.FailedPrecondition, "mission lottery required")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_mission SET status='active' WHERE id=$1`, ref.MissionID); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('mission.status','brand_admin',$1,'mission',$2,$3,jsonb_build_object('from',$4::text,'to','active'))`, actor, ref.MissionID, ref.PlanetID, m.Status); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return encode(map[string]any{"missionId": ref.MissionID, "status": "active"})
}
