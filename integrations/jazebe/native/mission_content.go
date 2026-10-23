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
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type contentMissionDraft struct {
	PlanetID          string                    `json:"planetId"`
	Title             string                    `json:"title"`
	ShortDescription  string                    `json:"shortDescription"`
	RewardPoints      int                       `json:"rewardPoints"`
	RewardTickets     int                       `json:"rewardTickets"`
	RewardXp          int                       `json:"rewardXp"`
	TargetType        string                    `json:"targetType"`
	GroupTargetValue  int                       `json:"groupTargetValue"`
	TargetLotteryID   string                    `json:"targetLotteryId"`
	RepeatPattern     string                    `json:"repeatPattern"`
	RepeatPeriod      string                    `json:"repeatPeriod"`
	DisplayConditions *missionDisplayConditions `json:"displayConditions"`
	StartsAt          string                    `json:"startsAt"`
	EndsAt            string                    `json:"endsAt"`
	Content           struct {
		EvidenceTypes      []string `json:"evidenceTypes"`
		Instructions       string   `json:"instructions"`
		AcceptanceCriteria string   `json:"acceptanceCriteria"`
		MinWords           int      `json:"minWords"`
		MaxFileSizeMb      int      `json:"maxFileSizeMb"`
	} `json:"content"`
}

func validateContentDraft(input contentMissionDraft) (time.Time, time.Time, error) {
	if _, err := supportedReviewedRepeat(input.RepeatPattern, input.RepeatPeriod); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !utf8.ValidString(input.Title) || utf8.RuneCountInString(strings.TrimSpace(input.Title)) < 1 || utf8.RuneCountInString(input.Title) > 128 ||
		!utf8.ValidString(input.ShortDescription) || utf8.RuneCountInString(input.ShortDescription) > 512 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid mission title or description")
	}
	if input.RewardPoints < 0 || input.RewardPoints > 1_000_000 || input.RewardXp < 0 || input.RewardXp > 1_000_000 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid reward")
	}
	if err := validateMissionTargetRewards(input.TargetType, input.GroupTargetValue, input.RewardPoints, input.RewardTickets, input.TargetLotteryID); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if len(input.Content.EvidenceTypes) < 1 || len(input.Content.EvidenceTypes) > 4 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "evidence types required")
	}
	seen := map[string]bool{}
	for _, v := range input.Content.EvidenceTypes {
		if v != "text" && v != "code" && v != "img" && v != "video" {
			return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid evidence type")
		}
		if seen[v] {
			return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "duplicate evidence type")
		}
		seen[v] = true
	}
	if !utf8.ValidString(input.Content.Instructions) || utf8.RuneCountInString(input.Content.Instructions) > 1024 ||
		!utf8.ValidString(input.Content.AcceptanceCriteria) || utf8.RuneCountInString(input.Content.AcceptanceCriteria) > 1024 ||
		input.Content.MinWords < 0 || input.Content.MinWords > 10_000 || input.Content.MaxFileSizeMb < 0 || input.Content.MaxFileSizeMb > 512 {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid content configuration")
	}
	start, err := time.Parse(time.RFC3339, input.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid start")
	}
	end, err := time.Parse(time.RFC3339, input.EndsAt)
	if err != nil || !end.After(start) {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, "invalid end")
	}
	return start, end, nil
}

func contentMissionDraftCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input contentMissionDraft
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	start, end, err := validateContentDraft(input)
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
		return "", status.Error(codes.Internal, "content mission creation failed")
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
		return "", status.Error(codes.Internal, "content mission creation failed")
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
	config, err := json.Marshal(input.Content)
	if err != nil {
		return "", status.Error(codes.Internal, "content config failed")
	}
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission (id,planet_id,title,short_description,functional_type,target_type,repeat_pattern,repeat_period,
		detection_method,reward_points,reward_xp,type_config,starts_at,ends_at,status,created_by,display_conditions,reward_tickets,target_lottery_id,group_target_value)
		VALUES ($1,$2,$3,$4,'content_creation',$5,$6,$7,'manual_review',$8,$9,$10,$11,$12,'draft',$13,$14::jsonb,$15,$16,$17)`,
		id, planetID, strings.TrimSpace(input.Title), strings.TrimSpace(input.ShortDescription), input.TargetType, input.RepeatPattern,
		sql.NullString{String: input.RepeatPeriod, Valid: input.RepeatPeriod != ""}, input.RewardPoints,
		input.RewardXp, string(config), start, end, actor, display, input.RewardTickets, sql.NullString{String: input.TargetLotteryID, Valid: input.TargetLotteryID != ""}, sql.NullInt64{Int64: int64(input.GroupTargetValue), Valid: input.TargetType == "group"})
	if err != nil {
		logger.Error("Content mission insert: %v", err)
		return "", status.Error(codes.Internal, "content mission creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.create','brand_admin',$1,'mission',$2,$3,jsonb_build_object('functionalType','content_creation','status','draft'))`, actor, id, planetID); err != nil {
		return "", status.Error(codes.Internal, "content mission audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "content mission creation failed")
	}
	return encode(map[string]any{"id": id, "planetId": planetID, "status": "draft", "functionalType": "content_creation"})
}

func loadContentMission(ctx context.Context, tx *sql.Tx, ref missionRef) (mediaMission, error) {
	var m mediaMission
	err := tx.QueryRowContext(ctx, `SELECT m.title,m.status,p.status,m.detection_method,m.repeat_pattern,COALESCE(m.repeat_period,''),
		m.starts_at,m.ends_at,m.reward_points,m.reward_xp,m.type_config,m.display_conditions,m.target_type,m.reward_tickets,m.target_lottery_id
		FROM jazebe_mission m JOIN jazebe_planet p ON p.id=m.planet_id
		WHERE m.id=$1 AND m.planet_id=$2 AND m.functional_type='content_creation' FOR UPDATE OF m`, ref.MissionID, ref.PlanetID).Scan(
		&m.Title, &m.Status, &m.PlanetStatus, &m.Detection, &m.Repeat, &m.Period, &m.StartsAt, &m.EndsAt, &m.RewardPoints, &m.RewardXp, &m.Config, &m.Display, &m.TargetType, &m.RewardTickets, &m.TargetLotteryID)
	return m, err
}

func contentEvidenceTypes(config []byte) (map[string]bool, error) {
	var c struct {
		EvidenceTypes []string `json:"evidenceTypes"`
	}
	if err := json.Unmarshal(config, &c); err != nil {
		return nil, err
	}
	allowed := map[string]bool{}
	for _, v := range c.EvidenceTypes {
		if v != "text" && v != "code" && v != "img" && v != "video" {
			return nil, errors.New("invalid evidence type")
		}
		allowed[v] = true
	}
	if len(allowed) == 0 {
		return nil, errors.New("no supported evidence type")
	}
	return allowed, nil
}

func contentMissionActivate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
		return "", status.Error(codes.Internal, "content activation failed")
	}
	defer tx.Rollback()
	m, err := loadContentMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "content mission not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "content activation failed")
	}
	var allowed bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, ref.PlanetID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "content authorization failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if m.Status != "draft" && m.Status != "paused" {
		return "", status.Error(codes.FailedPrecondition, "mission cannot activate")
	}
	if m.EndsAt.Before(time.Now()) || m.Detection != "manual_review" {
		return "", status.Error(codes.FailedPrecondition, "mission action unavailable")
	}
	if _, err := missionPeriodKey(m.Repeat, m.Period, time.Now().UTC()); err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid mission repeat configuration")
	}
	if _, err := contentEvidenceTypes(m.Config); err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid content configuration")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission SET status='active' WHERE id=$1`, ref.MissionID); err != nil {
		return "", status.Error(codes.Internal, "content activation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.status','brand_admin',$1,'mission',$2,$3,jsonb_build_object('from',$4::text,'to','active'))`, actor, ref.MissionID, ref.PlanetID, m.Status); err != nil {
		return "", status.Error(codes.Internal, "content activation audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "content activation failed")
	}
	return encode(map[string]any{"missionId": ref.MissionID, "status": "active"})
}

func contentMissionSubmit(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID  string   `json:"planetId"`
		MissionID string   `json:"missionId"`
		Types     []string `json:"types"`
		Note      string   `json:"note"`
		MediaIDs  []string `json:"mediaIds"`
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
	note := strings.TrimSpace(input.Note)
	if !utf8.ValidString(note) || utf8.RuneCountInString(note) > 512 || len(input.MediaIDs) > 6 {
		return "", status.Error(codes.InvalidArgument, "invalid evidence note or media count")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "evidence submission failed")
	}
	defer tx.Rollback()
	m, err := loadContentMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "content mission not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "evidence submission failed")
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
	if m.Detection != "manual_review" {
		return "", status.Error(codes.FailedPrecondition, "content action unavailable")
	}
	periodKey, err := missionPeriodKey(m.Repeat, m.Period, now)
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, "content repeat mode unavailable")
	}
	allowed, err := contentEvidenceTypes(m.Config)
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid content configuration")
	}
	types := input.Types
	if len(types) == 0 {
		for _, kind := range []string{"img", "video", "text", "code"} {
			if allowed[kind] {
				types = append(types, kind)
			}
		}
	}
	if len(types) == 0 || len(types) > 4 {
		return "", status.Error(codes.InvalidArgument, "invalid evidence types")
	}
	seen := map[string]bool{}
	for _, kind := range types {
		if !allowed[kind] || seen[kind] {
			return "", status.Error(codes.InvalidArgument, "evidence type not allowed")
		}
		seen[kind] = true
	}
	var cfg struct {
		MaxFileSizeMb int `json:"maxFileSizeMb"`
	}
	if err := json.Unmarshal(m.Config, &cfg); err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid content configuration")
	}
	media := make([]map[string]any, 0, len(input.MediaIDs))
	files := make([]string, 0, len(input.MediaIDs))
	seenID := make(map[string]bool)
	for _, rawID := range input.MediaIDs {
		id, err := requiredUUID(rawID)
		if err != nil || seenID[id] {
			return "", status.Error(codes.InvalidArgument, "invalid evidence media IDs")
		}
		seenID[id] = true
		var mime string
		var size int
		err = tx.QueryRowContext(ctx, `SELECT mime_type,expected_bytes FROM jazebe_media_upload
			WHERE id=$1 AND planet_id=$2 AND owner_user_id=$3 AND purpose='evidence'
			AND status='complete' AND deleted_at IS NULL`, id, ref.PlanetID, user).Scan(&mime, &size)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.InvalidArgument, "evidence media not owned")
		}
		if err != nil {
			return "", status.Error(codes.Internal, "evidence media unavailable")
		}
		kind := "img"
		if strings.HasPrefix(mime, "video/") {
			kind = "video"
		}
		if !seen[kind] || (cfg.MaxFileSizeMb > 0 && size > cfg.MaxFileSizeMb*1024*1024) {
			return "", status.Error(codes.InvalidArgument, "evidence media does not match mission")
		}
		url := "/v2/jazebe/media/" + id
		files = append(files, url)
		media = append(media, map[string]any{"id": id, "url": url, "mimeType": mime, "sizeBytes": size})
	}
	if !(len(media) > 0 && (seen["img"] || seen["video"])) && !(note != "" && (seen["text"] || seen["code"])) {
		return "", status.Error(codes.InvalidArgument, "mission evidence required")
	}
	var existing bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission_completion WHERE mission_id=$1 AND user_id=$2
		AND deleted_at IS NULL AND (status='pending_review' OR (status='approved' AND COALESCE(period_key,'')=$3)))`,
		ref.MissionID, user, periodKey).Scan(&existing); err != nil {
		return "", status.Error(codes.Internal, "evidence submission failed")
	}
	if existing {
		return "", status.Error(codes.AlreadyExists, "mission already completed or pending review")
	}
	evidence, _ := json.Marshal(map[string]any{"types": types, "note": note, "files": files, "media": media})
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission_completion (id,planet_id,mission_id,user_id,group_id,status,period_key,evidence,submitted_at)
		VALUES ($1,$2,$3,$4,$5,'pending_review',$6,$7,$8)`, id, ref.PlanetID, ref.MissionID, user, groupID,
		sql.NullString{String: periodKey, Valid: periodKey != ""}, string(evidence), now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "mission already completed or pending review")
		}
		logger.Error("Content completion insert: %v", err)
		return "", status.Error(codes.Internal, "evidence submission failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "evidence submission failed")
	}
	return encode(map[string]any{"missionId": ref.MissionID, "completionId": id, "state": "pending"})
}
