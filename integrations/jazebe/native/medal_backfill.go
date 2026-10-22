package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const medalBackfillPage = 50
const nilMedalCursor = "00000000-0000-0000-0000-000000000000"

func queueMedalBackfill(ctx context.Context, tx *sql.Tx, planetID, medalID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO jazebe_medal_backfill(planet_medal_id,planet_id) VALUES($1,$2)
		ON CONFLICT(planet_medal_id) DO UPDATE SET last_user_id=$3,evaluated=0,awarded=0,attempts=0,
		last_error=NULL,next_attempt_at=now(),completed_at=NULL,updated_at=now()`, medalID, planetID, nilMedalCursor)
	return err
}

// One short transaction owns the job row and advances its cursor only after
// all acquisitions in this page commit. SKIP LOCKED prevents duplicate work
// across Nakama nodes; unique acquisition IDs protect retries after crashes.
func runMedalBackfillBatch(ctx context.Context, db *sql.DB) (worked bool, resultErr error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	var medalID, planetID, cursor string
	defer func() {
		_ = tx.Rollback()
		if resultErr != nil && medalID != "" {
			recordCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = db.ExecContext(recordCtx, `UPDATE jazebe_medal_backfill SET attempts=attempts+1,last_error=$2,
				next_attempt_at=now()+interval '1 minute'*LEAST(60,attempts+1),updated_at=now()
				WHERE planet_medal_id=$1`, medalID, resultErr.Error())
		}
	}()
	err = tx.QueryRowContext(ctx, `SELECT planet_medal_id,planet_id,last_user_id FROM jazebe_medal_backfill
		WHERE completed_at IS NULL AND next_attempt_at<=now()
		ORDER BY next_attempt_at,planet_medal_id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&medalID, &planetID, &cursor)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT status='active' FROM jazebe_planet_medal WHERE id=$1 AND planet_id=$2`, medalID, planetID).Scan(&active); err != nil {
		return true, err
	}
	if !active {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_medal_backfill SET completed_at=now(),updated_at=now() WHERE planet_medal_id=$1`, medalID)
		if err != nil {
			return true, err
		}
		return true, tx.Commit()
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_planet_membership
		WHERE planet_id=$1 AND is_active AND user_id>$2 ORDER BY user_id LIMIT $3`, planetID, cursor, medalBackfillPage)
	if err != nil {
		return true, err
	}
	users := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return true, err
		}
		users = append(users, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return true, err
	}
	awarded := 0
	now := time.Now().UTC()
	for _, user := range users {
		n, evalErr := evaluateMedalsInTx(ctx, tx, planetID, user, "", "", medalID, now)
		if evalErr != nil {
			return true, evalErr
		}
		awarded += n
		cursor = user
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_medal_backfill SET last_user_id=$2,evaluated=evaluated+$3,
		awarded=awarded+$4,completed_at=CASE WHEN $5::boolean THEN now() ELSE NULL END,
		last_error=NULL,updated_at=now() WHERE planet_medal_id=$1`, medalID, cursor, len(users), awarded, len(users) < medalBackfillPage)
	if err != nil {
		return true, err
	}
	return true, tx.Commit()
}

func startMedalBackfillWorker(ctx context.Context, logger runtime.Logger, db *sql.DB) error {
	enabled := os.Getenv("JAZEBE_MEDAL_BACKFILL_ENABLED")
	if enabled == "" || enabled == "false" {
		return nil
	}
	if enabled != "true" {
		return errors.New("JAZEBE_MEDAL_BACKFILL_ENABLED must be true or false")
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			for i := 0; i < 10 && ctx.Err() == nil; i++ {
				passCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				worked, err := runMedalBackfillBatch(passCtx, db)
				cancel()
				if err != nil {
					logger.Error("Jazebe medal backfill: %v", err)
					break
				}
				if !worked {
					break
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

func medalBrandBackfillStatus(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	var in medalBrandInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if _, _, err := medalBrandAccess(ctx, db, in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	var evaluated, awarded int64
	var attempts int
	var lastError sql.NullString
	var completed sql.NullTime
	var cursor string
	err = db.QueryRowContext(ctx, `SELECT last_user_id,evaluated,awarded,attempts,last_error,completed_at
		FROM jazebe_medal_backfill WHERE planet_id=$1 AND planet_medal_id=$2`, in.PlanetID, id).Scan(&cursor, &evaluated, &awarded, &attempts, &lastError, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "medal backfill not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "medal backfill unavailable")
	}
	return encode(map[string]any{"planetMedalId": id, "lastUserId": cursor, "evaluated": evaluated, "awarded": awarded,
		"attempts": attempts, "lastError": nullableString(lastError), "completed": completed.Valid})
}
