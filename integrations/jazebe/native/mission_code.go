package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const missionCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type codeMissionDraft struct {
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
	Code              struct {
		CodeMode  string `json:"codeMode"`
		MaxUses   int    `json:"maxUses"`
		ExpiresAt string `json:"expiresAt"`
	} `json:"code"`
}

func validateCodeDraft(m codeMissionDraft) (time.Time, time.Time, error) {
	if _, err := supportedReviewedRepeat(m.RepeatPattern, m.RepeatPeriod); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !utf8.ValidString(m.Title) || utf8.RuneCountInString(strings.TrimSpace(m.Title)) < 1 || utf8.RuneCountInString(m.Title) > 128 ||
		!utf8.ValidString(m.ShortDescription) || utf8.RuneCountInString(m.ShortDescription) > 512 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission title or description")
	}
	if m.DetectionMethod != "platform_auto" && m.DetectionMethod != "manual_review" {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid detection method")
	}
	if m.RewardPoints < 0 || m.RewardPoints > 1_000_000 || m.RewardXp < 0 || m.RewardXp > 1_000_000 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid reward")
	}
	if err := validateMissionTargetRewards(m.TargetType, m.GroupTargetValue, m.RewardPoints, m.RewardTickets, m.TargetLotteryID); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if m.Code.CodeMode != "unique" && m.Code.CodeMode != "shared" {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid code mode")
	}
	if m.Code.CodeMode == "unique" && m.Code.MaxUses > 1 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "unique code can be used once")
	}
	if m.Code.CodeMode == "shared" && (m.Code.MaxUses < 1 || m.Code.MaxUses > 1_000_000) {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid shared code uses")
	}
	start, err := time.Parse(time.RFC3339, m.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid start")
	}
	end, err := time.Parse(time.RFC3339, m.EndsAt)
	if err != nil || !end.After(start) {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid end")
	}
	if m.Code.ExpiresAt != "" {
		if _, err := time.Parse(time.RFC3339, m.Code.ExpiresAt); err != nil {
			return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid code expiration")
		}
	}
	return start, end, nil
}

func codeMissionDraftCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input codeMissionDraft
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	start, end, err := validateCodeDraft(input)
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "code mission creation failed")
	}
	defer tx.Rollback()
	var min, max, ticketMax int
	err = tx.QueryRowContext(ctx, `SELECT s.mission_reward_point_min,s.mission_reward_point_max,s.mission_reward_ticket_max FROM jazebe_planet_settings s
		JOIN jazebe_planet p ON p.id=s.planet_id JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active`, planetID, actor).Scan(&min, &max, &ticketMax)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "code mission creation failed")
	}
	if input.RewardPoints > 0 && (input.RewardPoints < min || input.RewardPoints > max) {
		return "", status.Error(codes.FailedPrecondition, "points outside planet policy")
	}
	if input.RewardTickets > ticketMax {
		return "", status.Error(codes.FailedPrecondition, "tickets outside planet policy")
	}
	if err := verifyMissionPrerequisites(ctx, tx, planetID, input.DisplayConditions); err != nil {
		return "", err
	}
	config, err := json.Marshal(input.Code)
	if err != nil {
		return "", status.Error(codes.Internal, "code config failed")
	}
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission (id,planet_id,title,short_description,functional_type,target_type,repeat_pattern,repeat_period,
		detection_method,reward_points,reward_xp,type_config,starts_at,ends_at,status,created_by,display_conditions,reward_tickets,target_lottery_id,group_target_value)
		VALUES ($1,$2,$3,$4,'code_entry',$5,$6,$7,$8,$9,$10,$11,$12,$13,'draft',$14,$15::jsonb,$16,$17,$18)`, id, planetID,
		strings.TrimSpace(input.Title), strings.TrimSpace(input.ShortDescription), input.TargetType, input.RepeatPattern,
		sql.NullString{String: input.RepeatPeriod, Valid: input.RepeatPeriod != ""}, input.DetectionMethod,
		input.RewardPoints, input.RewardXp, string(config), start, end, actor, display, input.RewardTickets, sql.NullString{String: input.TargetLotteryID, Valid: input.TargetLotteryID != ""}, sql.NullInt64{Int64: int64(input.GroupTargetValue), Valid: input.TargetType == "group"})
	if err != nil {
		logger.Error("Code mission insert: %v", err)
		return "", status.Error(codes.Internal, "code mission creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.create','brand_admin',$1,'mission',$2,$3,jsonb_build_object('functionalType','code_entry','status','draft'))`, actor, id, planetID); err != nil {
		return "", status.Error(codes.Internal, "code mission audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "code mission creation failed")
	}
	return encode(map[string]any{"id": id, "planetId": planetID, "status": "draft", "functionalType": "code_entry"})
}

func loadCodeMission(ctx context.Context, tx *sql.Tx, ref missionRef) (mediaMission, error) {
	var m mediaMission
	err := tx.QueryRowContext(ctx, `SELECT m.title,m.status,p.status,m.detection_method,m.repeat_pattern,COALESCE(m.repeat_period,''),
		m.starts_at,m.ends_at,m.reward_points,m.reward_xp,m.type_config,m.display_conditions,m.target_type,m.reward_tickets,m.target_lottery_id
		FROM jazebe_mission m JOIN jazebe_planet p ON p.id=m.planet_id
		WHERE m.id=$1 AND m.planet_id=$2 AND m.functional_type='code_entry' FOR UPDATE OF m`, ref.MissionID, ref.PlanetID).Scan(
		&m.Title, &m.Status, &m.PlanetStatus, &m.Detection, &m.Repeat, &m.Period, &m.StartsAt, &m.EndsAt, &m.RewardPoints, &m.RewardXp, &m.Config, &m.Display, &m.TargetType, &m.RewardTickets, &m.TargetLotteryID)
	return m, err
}

func codeMissionActivate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
		return "", status.Error(codes.Internal, "code activation failed")
	}
	defer tx.Rollback()
	m, err := loadCodeMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "code mission not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "code activation failed")
	}
	var allowed bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, ref.PlanetID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "code authorization failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if m.Status != "draft" && m.Status != "paused" {
		return "", status.Error(codes.FailedPrecondition, "mission cannot activate")
	}
	if m.EndsAt.Before(time.Now()) || (m.Detection != "platform_auto" && m.Detection != "manual_review") {
		return "", status.Error(codes.FailedPrecondition, "mission action unavailable")
	}
	if _, err := missionPeriodKey(m.Repeat, m.Period, time.Now().UTC()); err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid mission repeat configuration")
	}
	var config struct {
		CodeMode string `json:"codeMode"`
		MaxUses  int    `json:"maxUses"`
	}
	if json.Unmarshal(m.Config, &config) != nil || (config.CodeMode != "unique" && config.CodeMode != "shared") {
		return "", status.Error(codes.FailedPrecondition, "invalid code config")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission SET status='active' WHERE id=$1`, ref.MissionID); err != nil {
		return "", status.Error(codes.Internal, "code activation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.status','brand_admin',$1,'mission',$2,$3,jsonb_build_object('from',$4,'to','active'))`, actor, ref.MissionID, ref.PlanetID, m.Status); err != nil {
		return "", status.Error(codes.Internal, "code activation audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "code activation failed")
	}
	return encode(map[string]any{"missionId": ref.MissionID, "status": "active"})
}

func renderMissionCode(format string) (string, error) {
	var out strings.Builder
	for _, r := range format {
		if r != '#' {
			out.WriteRune(r)
			continue
		}
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(missionCodeAlphabet))))
		if err != nil {
			return "", err
		}
		out.WriteByte(missionCodeAlphabet[n.Int64()])
	}
	return out.String(), nil
}

type codeGenerationInput struct {
	PlanetID   string `json:"planetId"`
	MissionID  string `json:"missionId"`
	Count      int    `json:"count"`
	CodeFormat string `json:"codeFormat"`
	Mode       string `json:"mode"`
	MaxUses    int    `json:"maxUses"`
	ExpiresAt  string `json:"expiresAt"`
}

func codeMissionGenerate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input codeGenerationInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref := missionRef{PlanetID: input.PlanetID, MissionID: input.MissionID}
	ref.PlanetID, err = requiredUUID(ref.PlanetID)
	if err != nil {
		return "", err
	}
	ref.MissionID, err = requiredUUID(ref.MissionID)
	if err != nil {
		return "", err
	}
	if input.Count < 1 || input.Count > 1000 {
		return "", status.Error(codes.InvalidArgument, "count must be 1 to 1000")
	}
	format := strings.TrimSpace(input.CodeFormat)
	if format == "" {
		format = "######"
	}
	if !utf8.ValidString(format) || utf8.RuneCountInString(format) < 3 || utf8.RuneCountInString(format) > 48 {
		return "", status.Error(codes.InvalidArgument, "invalid code format")
	}
	var expires any
	if input.ExpiresAt != "" {
		v, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid expiration")
		}
		expires = v
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "code generation failed")
	}
	defer tx.Rollback()
	m, err := loadCodeMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "code mission not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "code generation failed")
	}
	var allowed bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, ref.PlanetID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "code authorization failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	var config struct {
		CodeMode string `json:"codeMode"`
		MaxUses  int    `json:"maxUses"`
	}
	if json.Unmarshal(m.Config, &config) != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid code config")
	}
	mode := input.Mode
	if mode == "" {
		mode = config.CodeMode
	}
	if mode != "unique" && mode != "shared" {
		return "", status.Error(codes.InvalidArgument, "invalid code mode")
	}
	uses := input.MaxUses
	if uses == 0 {
		if mode == "unique" {
			uses = 1
		} else {
			uses = config.MaxUses
		}
	}
	if uses < 1 || uses > 1_000_000 || (mode == "unique" && uses != 1) {
		return "", status.Error(codes.InvalidArgument, "invalid max uses")
	}
	codesOut := make([]string, 0, input.Count)
	for i := 0; i < input.Count; i++ {
		inserted := false
		for attempt := 0; attempt < 5; attempt++ {
			code, err := renderMissionCode(format)
			if err != nil {
				return "", status.Error(codes.Internal, "code generation failed")
			}
			var id string
			err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_mission_code (id,planet_id,mission_id,code,mode,max_uses,expires_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (mission_id,code) DO NOTHING RETURNING id`, uuid.Must(uuid.NewV4()).String(), ref.PlanetID, ref.MissionID, code, mode, uses, expires).Scan(&id)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				logger.Error("Code stock insert: %v", err)
				return "", status.Error(codes.Internal, "code generation failed")
			}
			codesOut = append(codesOut, code)
			inserted = true
			break
		}
		if !inserted {
			break
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.codes.generate','brand_admin',$1,'mission',$2,$3,jsonb_build_object('created',$4))`, actor, ref.MissionID, ref.PlanetID, len(codesOut)); err != nil {
		return "", status.Error(codes.Internal, "code generation audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "code generation failed")
	}
	return encode(map[string]any{"created": len(codesOut), "codes": codesOut})
}

func codeMissionRedeem(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID  string `json:"planetId"`
		MissionID string `json:"missionId"`
		Code      string `json:"code"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref := missionRef{PlanetID: input.PlanetID, MissionID: input.MissionID}
	ref.PlanetID, err = requiredUUID(ref.PlanetID)
	if err != nil {
		return "", err
	}
	ref.MissionID, err = requiredUUID(ref.MissionID)
	if err != nil {
		return "", err
	}
	code := strings.TrimSpace(input.Code)
	if !utf8.ValidString(code) || utf8.RuneCountInString(code) < 1 || utf8.RuneCountInString(code) > 64 {
		return "", status.Error(codes.InvalidArgument, "invalid code")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "code redemption failed")
	}
	defer tx.Rollback()
	m, err := loadCodeMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "code mission not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "code redemption failed")
	}
	now := time.Now().UTC()
	if err := mediaOpen(m, now); err != nil {
		return "", err
	}
	if err := requireMissionVisible(ctx, tx, ref, user, m.Display, now); err != nil {
		return "", err
	}
	groupID, err := activeGroupForMission(ctx, tx, ref.PlanetID, user, m.TargetType == "group" || m.RewardTickets > 0)
	if err != nil {
		return "", err
	}
	periodKey, err := missionPeriodKey(m.Repeat, m.Period, now)
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, "mission repeat mode unavailable")
	}
	var existing bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission_completion WHERE mission_id=$1 AND user_id=$2
		AND deleted_at IS NULL AND (status='pending_review' OR (status='approved' AND COALESCE(period_key,'')=$3)))`,
		ref.MissionID, user, periodKey).Scan(&existing); err != nil {
		return "", status.Error(codes.Internal, "code redemption failed")
	}
	if existing {
		return "", status.Error(codes.AlreadyExists, "mission already completed or pending review")
	}
	var codeID, mode string
	var uses, maxUses int
	var assigned sql.NullString
	var expires sql.NullTime
	var active bool
	err = tx.QueryRowContext(ctx, `SELECT id,mode,used_count,max_uses,assigned_user_id,expires_at,is_active
		FROM jazebe_mission_code WHERE mission_id=$1 AND code=$2 FOR UPDATE`, ref.MissionID, code).Scan(&codeID, &mode, &uses, &maxUses, &assigned, &expires, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.InvalidArgument, "mission code invalid")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "code redemption failed")
	}
	if !active || uses >= maxUses || (expires.Valid && !now.Before(expires.Time)) {
		return "", status.Error(codes.FailedPrecondition, "mission code used or expired")
	}
	if mode == "unique" && assigned.Valid && assigned.String != user {
		return "", status.Error(codes.InvalidArgument, "mission code invalid")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission_code SET used_count=used_count+1,
		assigned_user_id=CASE WHEN mode='unique' THEN $2::uuid ELSE assigned_user_id END WHERE id=$1`, codeID, user); err != nil {
		return "", status.Error(codes.Internal, "code redemption failed")
	}
	completionID := uuid.Must(uuid.NewV4()).String()
	state := "approved"
	if m.Detection == "manual_review" {
		state = "pending_review"
	}
	evidence, _ := json.Marshal(map[string]string{"code": code})
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission_completion (id,planet_id,mission_id,user_id,group_id,status,period_key,evidence,submitted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, completionID, ref.PlanetID, ref.MissionID, user, groupID, state,
		sql.NullString{String: periodKey, Valid: periodKey != ""}, string(evidence), now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "mission already completed")
		}
		return "", status.Error(codes.Internal, "code redemption failed")
	}
	if state == "approved" {
		if m.TargetType == "group" {
			err = advanceGroupMission(ctx, tx, ref, user, groupID.String, completionID, now)
		} else {
			err = paySoloMission(ctx, tx, ref, user, completionID, m, now)
		}
		if err != nil {
			logger.Error("Code payout: %v", err)
			return "", status.Error(codes.Internal, "code payout failed")
		}
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "code redemption failed")
	}
	if state == "approved" {
		missionMedalProgression(ctx, logger, db, ref.PlanetID, user, completionID)
	}
	result := "done"
	if state == "pending_review" {
		result = "pending"
	}
	return encode(map[string]any{"missionId": ref.MissionID, "completionId": completionID, "state": result})
}

func codeMissionCodes(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	ref := missionRef{PlanetID: input.PlanetID, MissionID: input.MissionID}
	ref.PlanetID, err = requiredUUID(ref.PlanetID)
	if err != nil {
		return "", err
	}
	ref.MissionID, err = requiredUUID(ref.MissionID)
	if err != nil {
		return "", err
	}
	if input.Status != "" && input.Status != "all" && input.Status != "available" && input.Status != "used" && input.Status != "expired" {
		return "", status.Error(codes.InvalidArgument, "invalid code filter")
	}
	if input.Page == 0 {
		input.Page = 1
	}
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Page < 1 || input.Limit < 1 || input.Limit > 100 || input.Page > 1_000_000 {
		return "", status.Error(codes.InvalidArgument, "invalid pagination")
	}
	var allowed bool
	err = db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission m JOIN jazebe_planet p ON p.id=m.planet_id
		JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id WHERE m.id=$1 AND m.planet_id=$2
		AND m.functional_type='code_entry' AND a.user_id=$3 AND a.is_active)`, ref.MissionID, ref.PlanetID, actor).Scan(&allowed)
	if err != nil {
		return "", status.Error(codes.Internal, "code stock unavailable")
	}
	if !allowed {
		return "", status.Error(codes.NotFound, "code mission not found")
	}
	filter := "true"
	switch input.Status {
	case "available":
		filter = "is_active AND used_count=0 AND (expires_at IS NULL OR expires_at>now())"
	case "used":
		filter = "used_count>0"
	case "expired":
		filter = "expires_at<=now()"
	}
	query := `SELECT id,code,mode,max_uses,used_count,assigned_user_id,expires_at,is_active,created_at
		FROM jazebe_mission_code WHERE mission_id=$1 AND ` + filter + ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := db.QueryContext(ctx, query, ref.MissionID, input.Limit, (input.Page-1)*input.Limit)
	if err != nil {
		logger.Error("Code stock list: %v", err)
		return "", status.Error(codes.Internal, "code stock unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, code, mode string
		var uses, maxUses int
		var assigned sql.NullString
		var expires sql.NullTime
		var active bool
		var created time.Time
		if err := rows.Scan(&id, &code, &mode, &maxUses, &uses, &assigned, &expires, &active, &created); err != nil {
			return "", status.Error(codes.Internal, "code stock unavailable")
		}
		item := map[string]any{"id": id, "code": code, "mode": mode, "maxUses": maxUses, "usedCount": uses, "isActive": active, "createdAt": created}
		if assigned.Valid {
			item["assignedUserId"] = assigned.String
		}
		if expires.Valid {
			item["expiresAt"] = expires.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "code stock unavailable")
	}
	var total, available, used, expired int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_mission_code WHERE mission_id=$1 AND `+filter, ref.MissionID).Scan(&total); err != nil {
		return "", status.Error(codes.Internal, "code stock unavailable")
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE is_active AND used_count<max_uses AND (expires_at IS NULL OR expires_at>now())),
		count(*) FILTER (WHERE used_count>0),count(*) FILTER (WHERE expires_at<=now())
		FROM jazebe_mission_code WHERE mission_id=$1`, ref.MissionID).Scan(&available, &used, &expired); err != nil {
		return "", status.Error(codes.Internal, "code stock unavailable")
	}
	return encode(map[string]any{"codes": items, "total": total, "inventory": map[string]int{"available": available, "used": used, "expired": expired}})
}
