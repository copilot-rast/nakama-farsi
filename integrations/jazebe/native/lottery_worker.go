package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Opt-in while existing Nakama deployments may not have applied Jazebe SQL.
// Every node may run this worker: the conditional flip and lottery row lock
// serialize a draw, and a committed drawing seed survives process crashes.
func startLotteryDrawWorker(ctx context.Context, logger runtime.Logger, db *sql.DB) error {
	enabled := os.Getenv("JAZEBE_LOTTERY_DRAW_WORKER_ENABLED")
	if enabled == "" || enabled == "false" {
		return nil
	}
	if enabled != "true" {
		return errors.New("JAZEBE_LOTTERY_DRAW_WORKER_ENABLED must be true or false")
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			passCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			if err := drawDueLotteries(passCtx, logger, db, time.Now().UTC()); err != nil {
				logger.Error("Jazebe lottery draw pass: %v", err)
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

func drawDueLotteries(ctx context.Context, logger runtime.Logger, db *sql.DB, now time.Time) error {
	rows, err := db.QueryContext(ctx, `SELECT id FROM jazebe_lottery WHERE (status='active' AND draw_at<=$1) OR status='drawing' ORDER BY draw_at,id LIMIT 20`, now)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = drawScheduledLottery(ctx, db, id, now); err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				continue
			}
			logger.Error("Jazebe lottery %s draw failed: %v", id, err)
		}
	}
	return nil
}

func drawScheduledLottery(ctx context.Context, db *sql.DB, id string, now time.Time) error {
	seed, err := newLotteryDrawSeed()
	if err != nil {
		return err
	}
	meta, _ := json.Marshal(map[string]any{"algorithm": lotteryDrawAlgorithm, "snapshotAt": now.Format(time.RFC3339Nano), "trigger": "schedule"})
	res, err := db.ExecContext(ctx, `UPDATE jazebe_lottery SET status='drawing',draw_seed=COALESCE(draw_seed,$2),draw_meta=$3::jsonb,updated_at=$4 WHERE id=$1 AND status='active' AND draw_at<=$4`, id, seed, string(meta), now)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var state string
		err = db.QueryRowContext(ctx, `SELECT status FROM jazebe_lottery WHERE id=$1`, id).Scan(&state)
		if err != nil {
			return err
		}
		if state != "drawing" {
			return status.Error(codes.FailedPrecondition, "lottery not drawable")
		}
	}
	_, err = runLotteryDraw(ctx, db, id, "", "system", "schedule", now)
	return err
}
