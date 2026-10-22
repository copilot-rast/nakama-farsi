package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type missionRef struct {
	PlanetID  string `json:"planetId"`
	MissionID string `json:"missionId"`
}

func readMissionRef(payload string) (missionRef, error) {
	var ref missionRef
	if err := decode(payload, &ref); err != nil {
		return ref, err
	}
	var err error
	ref.PlanetID, err = requiredUUID(ref.PlanetID)
	if err != nil {
		return ref, err
	}
	ref.MissionID, err = requiredUUID(ref.MissionID)
	return ref, err
}

type mediaMission struct {
	Title, Status, PlanetStatus, Detection, Repeat, Period string
	TargetType                                             string
	StartsAt, EndsAt                                       time.Time
	RewardPoints, RewardXp, RewardTickets                  int
	TargetLotteryID                                        sql.NullString
	Config, Display                                        []byte
}

type xpLevel struct {
	Level int    `json:"level"`
	Key   string `json:"key"`
	MinXp int64  `json:"minXp"`
}

func levelAt(total int64, raw []byte, current int) (int, string) {
	level, key := 1, "newcomer"
	var thresholds []xpLevel
	if json.Unmarshal(raw, &thresholds) != nil || len(thresholds) != 3 {
		return current, key
	}
	sort.Slice(thresholds, func(i, j int) bool { return thresholds[i].Level < thresholds[j].Level })
	previous := thresholds[0].MinXp
	for i := 1; i < 3; i++ {
		step := thresholds[i]
		if step.Level != i+1 || step.MinXp <= previous {
			break
		}
		if total >= step.MinXp {
			level, key = step.Level, step.Key
		}
		previous = step.MinXp
	}
	if current > level {
		return current, ""
	}
	return level, key
}

func loadMediaMission(ctx context.Context, tx *sql.Tx, ref missionRef) (mediaMission, error) {
	var m mediaMission
	err := tx.QueryRowContext(ctx, `SELECT m.title,m.status,p.status,m.detection_method,m.repeat_pattern,COALESCE(m.repeat_period,''),
		m.starts_at,m.ends_at,m.reward_points,m.reward_xp,m.type_config,m.display_conditions,m.target_type,m.reward_tickets,m.target_lottery_id
		FROM jazebe_mission m JOIN jazebe_planet p ON p.id=m.planet_id
		WHERE m.id=$1 AND m.planet_id=$2 AND m.functional_type='media_view' FOR UPDATE OF m`,
		ref.MissionID, ref.PlanetID).Scan(&m.Title, &m.Status, &m.PlanetStatus, &m.Detection, &m.Repeat, &m.Period,
		&m.StartsAt, &m.EndsAt, &m.RewardPoints, &m.RewardXp, &m.Config, &m.Display, &m.TargetType, &m.RewardTickets, &m.TargetLotteryID)
	return m, err
}

func mediaWatchSeconds(config []byte) (int, error) {
	var c struct {
		MinWatchSec int    `json:"minWatchSec"`
		TotalSec    int    `json:"totalSec"`
		AssetURL    string `json:"assetUrl"`
	}
	if err := json.Unmarshal(config, &c); err != nil {
		return 0, err
	}
	if c.MinWatchSec < 1 || c.TotalSec < c.MinWatchSec || c.AssetURL == "" {
		return 0, errors.New("incomplete media mission")
	}
	return c.MinWatchSec, nil
}

func mediaOpen(m mediaMission, now time.Time) error {
	if m.Status != "active" || m.PlanetStatus != "active" || now.Before(m.StartsAt) || !now.Before(m.EndsAt) {
		return status.Error(codes.FailedPrecondition, "mission is not open")
	}
	return nil
}

// Draft activation is restricted to media missions whose action and payout
// paths are implemented. It does not silently activate other mission types.
func mediaMissionActivate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
		return "", status.Error(codes.Internal, "mission activation failed")
	}
	defer tx.Rollback()
	m, err := loadMediaMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "media mission not found")
	}
	if err != nil {
		logger.Error("Mission activation lookup: %v", err)
		return "", status.Error(codes.Internal, "mission activation failed")
	}
	var allowed bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, ref.PlanetID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "mission authorization failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if m.Status != "draft" && m.Status != "paused" {
		return "", status.Error(codes.FailedPrecondition, "mission cannot be activated from current status")
	}
	if (m.Repeat != "once" && m.Repeat != "periodic") || (m.Detection != "platform_auto" && m.Detection != "manual_review") || m.EndsAt.Before(time.Now()) {
		return "", status.Error(codes.FailedPrecondition, "mission action path is not available for this configuration")
	}
	if _, err := missionPeriodKey(m.Repeat, m.Period, time.Now().UTC()); err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid mission repeat configuration")
	}
	if _, err := mediaWatchSeconds(m.Config); err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid media configuration")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission SET status='active' WHERE id=$1`, ref.MissionID); err != nil {
		return "", status.Error(codes.Internal, "mission activation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.status','brand_admin',$1,'mission',$2,$3,jsonb_build_object('from',$4,'to','active'))`, actor, ref.MissionID, ref.PlanetID, m.Status); err != nil {
		return "", status.Error(codes.Internal, "mission activation audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "mission activation failed")
	}
	return encode(map[string]any{"missionId": ref.MissionID, "status": "active"})
}

func mediaMissionStart(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readMissionRef(payload)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "media start failed")
	}
	defer tx.Rollback()
	m, err := loadMediaMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "media mission not found")
	}
	if err != nil {
		logger.Error("Media mission lookup: %v", err)
		return "", status.Error(codes.Internal, "media start failed")
	}
	now := time.Now().UTC()
	if err := mediaOpen(m, now); err != nil {
		return "", err
	}
	if err := requireMissionVisible(ctx, tx, ref, user, m.Display, now); err != nil {
		return "", err
	}
	if _, err := activeGroupForMission(ctx, tx, ref.PlanetID, user, m.TargetType == "group" || m.RewardTickets > 0); err != nil {
		return "", err
	}
	periodKey, err := missionPeriodKey(m.Repeat, m.Period, now)
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, "mission repeat mode unavailable")
	}
	startKey := periodKey
	if startKey == "" {
		startKey = "once"
	}
	need, err := mediaWatchSeconds(m.Config)
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid media configuration")
	}
	var done bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission_completion WHERE mission_id=$1 AND user_id=$2
		AND deleted_at IS NULL AND (status='pending_review' OR (status='approved' AND COALESCE(period_key,'')=$3)))`, ref.MissionID, user, periodKey).Scan(&done); err != nil {
		return "", status.Error(codes.Internal, "media start failed")
	}
	if done {
		return "", status.Error(codes.AlreadyExists, "mission already completed or pending review")
	}
	var started time.Time
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_mission_media_start (mission_id,planet_id,user_id,period_key,started_at)
		VALUES ($1,$2,$3,$4,$5) ON CONFLICT (mission_id,user_id,period_key) DO UPDATE
		SET started_at=jazebe_mission_media_start.started_at RETURNING started_at`, ref.MissionID, ref.PlanetID, user, startKey, now).Scan(&started)
	if err != nil {
		logger.Error("Media start insert: %v", err)
		return "", status.Error(codes.Internal, "media start failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "media start failed")
	}
	return encode(map[string]any{"missionId": ref.MissionID, "startedAt": started.UTC().Format(time.RFC3339Nano), "completableAt": started.Add(time.Duration(need) * time.Second).UTC().Format(time.RFC3339Nano)})
}

func mediaMissionComplete(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readMissionRef(payload)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "mission completion failed")
	}
	defer tx.Rollback()
	m, err := loadMediaMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "media mission not found")
	}
	if err != nil {
		logger.Error("Mission completion lookup: %v", err)
		return "", status.Error(codes.Internal, "mission completion failed")
	}
	now := time.Now().UTC()
	if err := mediaOpen(m, now); err != nil {
		return "", err
	}
	if err := requireMissionVisible(ctx, tx, ref, user, m.Display, now); err != nil {
		return "", err
	}
	periodKey, err := missionPeriodKey(m.Repeat, m.Period, now)
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, "mission repeat mode unavailable")
	}
	startKey := periodKey
	if startKey == "" {
		startKey = "once"
	}
	need, err := mediaWatchSeconds(m.Config)
	if err != nil {
		return "", status.Error(codes.FailedPrecondition, "invalid media configuration")
	}
	var started time.Time
	err = tx.QueryRowContext(ctx, `SELECT started_at FROM jazebe_mission_media_start
		WHERE mission_id=$1 AND user_id=$2 AND period_key=$3 FOR UPDATE`, ref.MissionID, user, startKey).Scan(&started)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.FailedPrecondition, "start media before completing")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "mission completion failed")
	}
	if now.Sub(started) < time.Duration(need)*time.Second {
		return "", status.Error(codes.FailedPrecondition, "media watch time too short")
	}
	var existing bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission_completion WHERE mission_id=$1 AND user_id=$2
		AND deleted_at IS NULL AND (status='pending_review' OR (status='approved' AND COALESCE(period_key,'')=$3)))`, ref.MissionID, user, periodKey).Scan(&existing); err != nil {
		return "", status.Error(codes.Internal, "mission completion failed")
	}
	if existing {
		return "", status.Error(codes.AlreadyExists, "mission already completed or pending review")
	}
	completionID := uuid.Must(uuid.NewV4()).String()
	groupID, err := activeGroupForMission(ctx, tx, ref.PlanetID, user, m.TargetType == "group" || m.RewardTickets > 0)
	if err != nil {
		return "", err
	}
	state := "approved"
	if m.Detection == "manual_review" {
		state = "pending_review"
	} else if m.Detection != "platform_auto" {
		return "", status.Error(codes.FailedPrecondition, "unsupported detection method")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_mission_completion (id,planet_id,mission_id,user_id,group_id,status,period_key,submitted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, completionID, ref.PlanetID, ref.MissionID, user, groupID, state,
		sql.NullString{String: periodKey, Valid: periodKey != ""}, now); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "mission already completed or pending review")
		}
		logger.Error("Completion insert: %v", err)
		return "", status.Error(codes.Internal, "mission completion failed")
	}
	if state == "approved" {
		if m.TargetType == "group" {
			err = advanceGroupMission(ctx, tx, ref, user, groupID.String, completionID, now)
		} else {
			err = paySoloMission(ctx, tx, ref, user, completionID, m, now)
		}
		if err != nil {
			logger.Error("Mission payout: %v", err)
			return "", status.Error(codes.Internal, "mission payout failed")
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM jazebe_mission_media_start WHERE mission_id=$1 AND user_id=$2 AND period_key=$3`, ref.MissionID, user, startKey); err != nil {
		return "", status.Error(codes.Internal, "mission completion failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "mission completion failed")
	}
	if state == "approved" {
		missionMedalProgression(ctx, logger, db, ref.PlanetID, user, completionID)
	}
	return encode(map[string]any{"missionId": ref.MissionID, "completionId": completionID, "state": map[bool]string{true: "done", false: "pending"}[state == "approved"]})
}

// All custom point and XP movements use the same SQL transaction as the
// completion. Nakama's own wallet has a separate transaction boundary.
func paySoloMission(ctx context.Context, tx *sql.Tx, ref missionRef, user, completionID string, m mediaMission, now time.Time) error {
	if m.TargetType == "group" {
		return status.Error(codes.FailedPrecondition, "group missions use group payout")
	}
	var pointID any
	if m.RewardPoints > 0 {
		movement, err := applyPointMovement(ctx, tx, ref.PlanetID, user, "credit", int64(m.RewardPoints), "mission", completionID, m.Title, now)
		if err != nil {
			return err
		}
		if !movement.Applied {
			return nil
		}
		pointID = movement.TransactionID
	}
	if m.RewardTickets > 0 {
		if !m.TargetLotteryID.Valid {
			return status.Error(codes.FailedPrecondition, "mission lottery required")
		}
		groupID, err := activeGroupForMission(ctx, tx, ref.PlanetID, user, true)
		if err != nil {
			return err
		}
		if _, err = issueLotteryTickets(ctx, tx, ticketIssue{LotteryID: m.TargetLotteryID.String, PlanetID: ref.PlanetID, GroupID: groupID.String, UserID: user, CreatorType: "member", Source: "solo_mission", SourceID: completionID, Count: m.RewardTickets}, now); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload) VALUES($1,$2,$3,'mission',$4,'mission.completed',jsonb_build_object('missionId',$5::uuid,'tickets',$6))`, uuid.Must(uuid.NewV4()).String(), ref.PlanetID, groupID.String, user, ref.MissionID, m.RewardTickets); err != nil {
			return status.Error(codes.Internal, "mission group activity failed")
		}
	}
	if m.RewardXp > 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_xp_record (id,planet_id,user_id,amount,source,source_id)
			VALUES ($1,$2,$3,$4,'mission',$5)`, uuid.Must(uuid.NewV4()).String(), ref.PlanetID, user, m.RewardXp, completionID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats (planet_id,user_id)
		VALUES ($1,$2) ON CONFLICT (planet_id,user_id) DO NOTHING`, ref.PlanetID, user); err != nil {
		return err
	}
	var total int64
	var current int
	if err := tx.QueryRowContext(ctx, `UPDATE jazebe_user_planet_stats
		SET total_xp=total_xp+$3,missions_completed=missions_completed+1,last_activity_at=$4,tickets_created=tickets_created+$5
		WHERE planet_id=$1 AND user_id=$2 RETURNING total_xp,level`, ref.PlanetID, user, m.RewardXp, now, m.RewardTickets).Scan(&total, &current); err != nil {
		return err
	}
	var thresholds []byte
	if err := tx.QueryRowContext(ctx, `SELECT xp_level_thresholds FROM jazebe_planet_settings WHERE planet_id=$1`, ref.PlanetID).Scan(&thresholds); err != nil {
		return err
	}
	level, key := levelAt(total, thresholds, current)
	if level > current {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET level=$3,level_key=$4
		WHERE planet_id=$1 AND user_id=$2`, ref.PlanetID, user, level, key); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission_completion SET awarded_points=$1,awarded_xp=$2,point_transaction_id=$3,awarded_tickets=$5
		WHERE id=$4`, m.RewardPoints, m.RewardXp, pointID, completionID, m.RewardTickets); err != nil {
		return err
	}
	var inserted bool
	if err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership (planet_id,user_id,join_path,last_activity_at)
		VALUES ($1,$2,'first_activity',$3) ON CONFLICT (planet_id,user_id) DO NOTHING RETURNING true`, ref.PlanetID, user, now).Scan(&inserted); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if inserted {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=fans_count+1 WHERE id=$1`, ref.PlanetID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=$3 WHERE planet_id=$1 AND user_id=$2`, ref.PlanetID, user, now); err != nil {
		return err
	}
	if err := enqueueActivityXP(ctx, tx, ref.PlanetID, user, completionID, "mission"); err != nil {
		return err
	}
	return persistJazebeNotification(ctx, tx, user, "mission:"+completionID, "ماموریت تکمیل شد", map[string]any{
		"category": "mission_game", "actionLink": "/missions", "missionId": ref.MissionID,
		"completionId": completionID, "points": m.RewardPoints, "xp": m.RewardXp, "title": m.Title}, now)
}

func pointBalance(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	id, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var balance int64
	err = db.QueryRowContext(ctx, `SELECT balance FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, id, user).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		balance = 0
	} else if err != nil {
		logger.Error("Point balance lookup: %v", err)
		return "", status.Error(codes.Internal, "point balance unavailable")
	}
	return encode(map[string]any{"planetId": id, "balance": balance})
}
