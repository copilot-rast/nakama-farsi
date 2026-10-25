package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type referralJob struct{ mission, inviter, planet string }

func inviteCompleteIfReachedTx(ctx context.Context, tx *sql.Tx, job referralJob, now time.Time) (string, error) {
	ref := missionRef{MissionID: job.mission, PlanetID: job.planet}
	if err := lockXact(ctx, tx, "jazebe_referral_completion:"+job.mission+":"+job.inviter); err != nil {
		return "", err
	}
	m, err := loadInviteMission(ctx, tx, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if mediaOpen(m, now) != nil {
		return "", nil
	}
	if m.Detection != "platform_auto" {
		return "", status.Error(codes.FailedPrecondition, "invite mission detection unavailable")
	}
	c, err := parseReferralMissionConfig(m.Config)
	if err != nil {
		return "", err
	}
	if c.TargetCount < 1 {
		return "", status.Error(codes.FailedPrecondition, "invite target required")
	}
	if err = requireMissionVisible(ctx, tx, ref, job.inviter, m.Display, now); err != nil {
		if status.Code(err) != codes.Internal {
			return "", nil
		}
		return "", err
	}
	period, err := missionPeriodKey(m.Repeat, m.Period, now)
	if err != nil {
		return "", err
	}
	if m.Repeat != "unlimited" {
		var existing bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_mission_completion WHERE mission_id=$1 AND user_id=$2
			AND status IN ('approved','pending_review') AND deleted_at IS NULL
			AND (status='pending_review' OR COALESCE(period_key,'')=$3))`, job.mission, job.inviter, period).Scan(&existing); err != nil {
			return "", err
		}
		if existing {
			return "", nil
		}
	}
	allowRegistered := c.Scope == "planet" && c.SuccessCondition == "signup"
	since := referralWindow(c, now)
	var available int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_mission_invite WHERE mission_id=$1 AND inviter_user_id=$2
		AND (status='qualified' OR ($3::boolean AND status='registered')) AND completion_id IS NULL
		AND ($4::timestamptz IS NULL OR COALESCE(qualified_at,registered_at)>=$4)`, job.mission, job.inviter, allowRegistered, since).Scan(&available); err != nil {
		return "", err
	}
	if available < c.TargetCount {
		return "", nil
	}
	group, err := activeGroupForMission(ctx, tx, job.planet, job.inviter, m.TargetType == "group" || m.RewardTickets > 0)
	if err != nil {
		return "", err
	}
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission_completion(id,planet_id,mission_id,user_id,group_id,status,period_key,submitted_at)
		VALUES($1,$2,$3,$4,$5,'approved',$6,$7)`, id, job.planet, job.mission, job.inviter, group,
		sql.NullString{String: period, Valid: period != ""}, now)
	if err != nil {
		return "", err
	}
	var claimed int
	if err = tx.QueryRowContext(ctx, `WITH chosen AS (
		SELECT id FROM jazebe_mission_invite WHERE mission_id=$2 AND inviter_user_id=$3
		AND (status='qualified' OR ($4::boolean AND status='registered')) AND completion_id IS NULL
		AND ($5::timestamptz IS NULL OR COALESCE(qualified_at,registered_at)>=$5)
		ORDER BY COALESCE(qualified_at,registered_at),id LIMIT $6 FOR UPDATE
		), changed AS (UPDATE jazebe_mission_invite i SET completion_id=$1,updated_at=$7
		FROM chosen WHERE i.id=chosen.id RETURNING i.id) SELECT count(*) FROM changed`,
		id, job.mission, job.inviter, allowRegistered, since, c.TargetCount, now).Scan(&claimed); err != nil {
		return "", err
	}
	if claimed != c.TargetCount {
		return "", status.Error(codes.Aborted, "invite qualification changed during payout")
	}
	if m.TargetType == "group" {
		err = advanceGroupMission(ctx, tx, ref, job.inviter, group.String, id, now)
	} else {
		err = paySoloMission(ctx, tx, ref, job.inviter, id, m, now)
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// Jobs are durable. Each transaction locks a job row; source invites and
// financial side effects commit together. Another Nakama node may run this
// worker without double-spending invitation rows.
func startReferralCompletionWorker(ctx context.Context, logger runtime.Logger, db *sql.DB) error {
	enabled := os.Getenv("JAZEBE_REFERRAL_WORKER_ENABLED")
	if enabled == "false" {
		return nil
	}
	if enabled != "" && enabled != "true" {
		return errors.New("JAZEBE_REFERRAL_WORKER_ENABLED must be true or false")
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			pass, cancel := context.WithTimeout(ctx, 45*time.Second)
			if err := runReferralCompletionJobs(pass, logger, db); err != nil && ctx.Err() == nil {
				logger.Error("Jazebe referral completion scan: %v", err)
			}
			cancel()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

func runReferralCompletionJobs(ctx context.Context, logger runtime.Logger, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `SELECT mission_id,inviter_user_id,planet_id FROM jazebe_referral_completion_job
		WHERE pending AND next_attempt_at<=now() ORDER BY next_attempt_at,mission_id LIMIT 25`)
	if err != nil {
		return err
	}
	jobs := []referralJob{}
	for rows.Next() {
		var j referralJob
		if err = rows.Scan(&j.mission, &j.inviter, &j.planet); err != nil {
			rows.Close()
			return err
		}
		jobs = append(jobs, j)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, j := range jobs {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = processReferralCompletionJob(ctx, logger, db, j); err != nil {
			logger.Error("Jazebe referral completion %s/%s: %v", j.mission, j.inviter, err)
		}
	}
	return nil
}

func processReferralCompletionJob(ctx context.Context, logger runtime.Logger, db *sql.DB, j referralJob) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var pending bool
	err = tx.QueryRowContext(ctx, `SELECT pending FROM jazebe_referral_completion_job WHERE mission_id=$1 AND inviter_user_id=$2 FOR UPDATE SKIP LOCKED`, j.mission, j.inviter).Scan(&pending)
	if errors.Is(err, sql.ErrNoRows) || !pending {
		return nil
	}
	if err != nil {
		return err
	}
	completionID, err := inviteCompleteIfReachedTx(ctx, tx, j, time.Now().UTC())
	if err != nil {
		_ = tx.Rollback()
		message := err.Error()
		if len(message) > 512 {
			message = message[:512]
		}
		_, recordErr := db.ExecContext(ctx, `UPDATE jazebe_referral_completion_job SET attempts=attempts+1,
			next_attempt_at=now()+LEAST(interval '1 hour',interval '1 minute'*(attempts+1)),
			last_error=$3,updated_at=now() WHERE mission_id=$1 AND inviter_user_id=$2`, j.mission, j.inviter, message)
		if recordErr != nil {
			return recordErr
		}
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_referral_completion_job SET pending=$3,last_error=NULL,
		next_attempt_at=now()+interval '1 second',updated_at=now()
		WHERE mission_id=$1 AND inviter_user_id=$2`, j.mission, j.inviter, completionID != ""); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if completionID != "" {
		missionMedalProgression(ctx, logger, db, j.planet, j.inviter, completionID)
	}
	return nil
}
