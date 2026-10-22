package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

func enqueueActivityXP(ctx context.Context, tx *sql.Tx, planetID, user, sourceID, source string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO jazebe_progression_event(id,planet_id,user_id,kind,source_id)
		VALUES($1,$2,$3,$4,$5) ON CONFLICT(kind,source_id,user_id) DO NOTHING`,
		uuid.Must(uuid.NewV4()).String(), planetID, user, source, sourceID)
	return err
}

// Jazebe awards configured purchase XP after the order transaction commits.
// The order ID is the XP ledger source ID, so replay cannot award it twice.
// Errors here are logged without changing the already committed purchase.
func shopPurchaseProgression(ctx context.Context, logger runtime.Logger, db *sql.DB, planetID, user, orderID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := awardActivityXP(ctx, db, planetID, user, orderID, "purchase"); err != nil {
		logger.Error("Jazebe purchase XP for order %s: %v", orderID, err)
	}
}

func groupJoinProgression(ctx context.Context, logger runtime.Logger, db *sql.DB, planetID, user, membershipID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := awardActivityXP(ctx, db, planetID, user, membershipID, "group_join"); err != nil {
		logger.Error("Jazebe group join XP for membership %s: %v", membershipID, err)
	}
}

func missionMedalProgression(ctx context.Context, logger runtime.Logger, db *sql.DB, planetID, user, completionID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := awardActivityXP(ctx, db, planetID, user, completionID, "mission"); err != nil {
		logger.Error("Jazebe mission medals for completion %s: %v", completionID, err)
	}
}

func gameMedalProgression(ctx context.Context, logger runtime.Logger, db *sql.DB, planetID, user, roundID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := awardActivityXP(ctx, db, planetID, user, roundID, "game"); err != nil {
		logger.Error("Jazebe game medals for round %s: %v", roundID, err)
	}
}

func awardActivityXP(ctx context.Context, db *sql.DB, planetID, user, sourceID, source string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var processed sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT processed_at FROM jazebe_progression_event
		WHERE planet_id=$1 AND user_id=$2 AND kind=$3 AND source_id=$4 FOR UPDATE`,
		planetID, user, source, sourceID).Scan(&processed)
	if err != nil {
		return err
	}
	if processed.Valid {
		return nil
	}
	if err = settleReferralActivityTx(ctx, tx, planetID, user, time.Now().UTC()); err != nil {
		return err
	}
	if source == "mission" || source == "game" {
		if _, err := evaluateMedalsInTx(ctx, tx, planetID, user, source, "", "", time.Now().UTC()); err != nil {
			return err
		}
		return finishActivityXP(ctx, tx, planetID, user, sourceID, source)
	}
	var amount int
	var thresholds []byte
	// Only the two known activities can select a settings column.
	column := "xp_per_purchase"
	if source == "group_join" {
		column = "xp_per_group_join"
	} else if source != "purchase" {
		return errors.New("unsupported XP activity")
	}
	err = tx.QueryRowContext(ctx, `SELECT `+column+`,xp_level_thresholds FROM jazebe_planet_settings WHERE planet_id=$1`, planetID).Scan(&amount, &thresholds)
	if err != nil {
		return err
	}
	if amount <= 0 {
		return finishActivityWithMedals(ctx, tx, planetID, user, sourceID, source)
	}
	var inserted bool
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_xp_record(id,planet_id,user_id,amount,source,source_id) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(planet_id,user_id,source,source_id) DO NOTHING RETURNING true`, uuid.Must(uuid.NewV4()).String(), planetID, user, amount, source, sourceID).Scan(&inserted)
	if errors.Is(err, sql.ErrNoRows) {
		return finishActivityWithMedals(ctx, tx, planetID, user, sourceID, source)
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id) VALUES($1,$2) ON CONFLICT(planet_id,user_id) DO NOTHING`, planetID, user); err != nil {
		return err
	}
	var total int64
	var current int
	err = tx.QueryRowContext(ctx, `UPDATE jazebe_user_planet_stats SET total_xp=total_xp+$3,last_activity_at=now() WHERE planet_id=$1 AND user_id=$2 RETURNING total_xp,level`, planetID, user, amount).Scan(&total, &current)
	if err != nil {
		return err
	}
	level, key := levelAt(total, thresholds, current)
	if level > current {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET level=$3,level_key=$4 WHERE planet_id=$1 AND user_id=$2`, planetID, user, level, key); err != nil {
			return err
		}
		title := "عضو فعال"
		if level == 3 {
			title = "کهنه‌کار"
		}
		if err = persistJazebeNotification(ctx, tx, user, source+"-level:"+sourceID, "سطح تو بالا رفت", map[string]any{"category": "mission_game", "actionLink": "/profile", "title": "سطح تو بالا رفت", "body": "تبریک! به سطح «" + title + "» رسیدی", "level": level}, time.Now().UTC()); err != nil {
			return err
		}
	}
	return finishActivityWithMedals(ctx, tx, planetID, user, sourceID, source)
}

func finishActivityWithMedals(ctx context.Context, tx *sql.Tx, planetID, user, sourceID, source string) error {
	activity := source
	if source == "group_join" {
		activity = "group_joined"
	}
	if _, err := evaluateMedalsInTx(ctx, tx, planetID, user, activity, "", "", time.Now().UTC()); err != nil {
		return err
	}
	if source == "group_join" {
		var creator string
		err := tx.QueryRowContext(ctx, `SELECT g.creator_user_id FROM jazebe_group_membership m
			JOIN jazebe_group g ON g.id=m.group_id WHERE m.id=$1 AND m.planet_id=$2`, sourceID, planetID).Scan(&creator)
		if err != nil {
			return err
		}
		if creator != user {
			if _, err = evaluateMedalsInTx(ctx, tx, planetID, creator, "group_joined", "", "", time.Now().UTC()); err != nil {
				return err
			}
		}
	}
	return finishActivityXP(ctx, tx, planetID, user, sourceID, source)
}

func finishActivityXP(ctx context.Context, tx *sql.Tx, planetID, user, sourceID, source string) error {
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_progression_event
		SET processed_at=now(),attempts=attempts+1
		WHERE planet_id=$1 AND user_id=$2 AND kind=$3 AND source_id=$4 AND processed_at IS NULL`,
		planetID, user, source, sourceID); err != nil {
		return err
	}
	return tx.Commit()
}

// An opt-in retry worker drains committed events whose immediate processing
// was interrupted. Row locks in awardActivityXP serialize all Nakama nodes.
func startProgressionWorker(ctx context.Context, logger runtime.Logger, db *sql.DB) error {
	enabled := os.Getenv("JAZEBE_PROGRESSION_WORKER_ENABLED")
	if enabled == "" || enabled == "false" {
		return nil
	}
	if enabled != "true" {
		return errors.New("JAZEBE_PROGRESSION_WORKER_ENABLED must be true or false")
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			passCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			if err := retryProgressionEvents(passCtx, logger, db); err != nil && ctx.Err() == nil {
				logger.Error("Jazebe progression retry: %v", err)
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

func retryProgressionEvents(ctx context.Context, logger runtime.Logger, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `SELECT planet_id,user_id,kind,source_id FROM jazebe_progression_event
		WHERE processed_at IS NULL ORDER BY created_at,id LIMIT 100`)
	if err != nil {
		return err
	}
	type event struct{ planet, user, kind, source string }
	events := make([]event, 0, 100)
	for rows.Next() {
		var e event
		if err = rows.Scan(&e.planet, &e.user, &e.kind, &e.source); err != nil {
			rows.Close()
			return err
		}
		events = append(events, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, e := range events {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = awardActivityXP(ctx, db, e.planet, e.user, e.source, e.kind); err != nil {
			logger.Error("Jazebe progression event %s/%s: %v", e.kind, e.source, err)
		}
	}
	return nil
}
