// Command reconcile repairs Jazebe planet fan counters without changing Nakama core.
// Run it once from a scheduler at 04:00 in the deployment's intended timezone.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const reconcileSQL = `WITH counts AS (
    SELECT p.id, count(m.user_id)::integer AS fans
      FROM jazebe_planet p
      LEFT JOIN jazebe_planet_membership m ON m.planet_id = p.id AND m.is_active
     GROUP BY p.id
)
UPDATE jazebe_planet p SET fans_count = counts.fans
  FROM counts WHERE p.id = counts.id AND p.fans_count <> counts.fans`

// reconcileLockKey is the fan reconciliation's row in jazebe_xact_lock.
const reconcileLockKey int64 = 774101

// Wallet balances are caches of the append-only ledger. Report drift for
// operator investigation; never silently change money in reconciliation.
const walletDriftSQL = `SELECT w.id,w.planet_id,w.user_id,w.balance,
    COALESCE(SUM(CASE WHEN t.direction='credit' THEN t.amount ELSE -t.amount END),0) AS ledger
    FROM jazebe_point_wallet w LEFT JOIN jazebe_point_transaction t ON t.wallet_id=w.id
    GROUP BY w.id
    HAVING w.balance<>COALESCE(SUM(CASE WHEN t.direction='credit' THEN t.amount ELSE -t.amount END),0)`

// Incomplete ticketed uploads can be discarded after their upload window.
// Completed assets are never removed here because mission evidence may refer to them.
const expiredUploadSQL = `DELETE FROM jazebe_media_upload
    WHERE id IN (SELECT id FROM jazebe_media_upload
        WHERE status='open' AND expires_at < now()
        ORDER BY expires_at LIMIT 1000 FOR UPDATE SKIP LOCKED)`

// A completed evidence upload can be abandoned without ever being submitted.
// Preserve even soft-deleted completions as references, matching Jazebe's sweep.
const withdrawAbandonedEvidenceSQL = `UPDATE jazebe_media_upload SET status='deleted',deleted_at=now()
    WHERE id IN (SELECT m.id FROM jazebe_media_upload m
        WHERE m.purpose='evidence' AND m.status='complete'
          AND m.created_at < now() - interval '7 days'
          AND NOT EXISTS (SELECT 1 FROM jazebe_mission_completion c
            WHERE c.evidence IS NOT NULL AND strpos(c.evidence::text,m.id::text)>0)
        ORDER BY m.created_at LIMIT 200 FOR UPDATE OF m SKIP LOCKED)`

const purgeWithdrawnUploadsSQL = `DELETE FROM jazebe_media_upload
    WHERE id IN (SELECT m.id FROM jazebe_media_upload m
        WHERE m.status='deleted'
          AND m.deleted_at < now() - interval '1 day'
          AND NOT EXISTS (SELECT 1 FROM jazebe_mission_completion c
            WHERE c.evidence IS NOT NULL AND strpos(c.evidence::text,m.id::text)>0)
          AND NOT EXISTS (SELECT 1 FROM jazebe_mission mission
            WHERE strpos(mission.type_config::text,m.id::text)>0)
          AND NOT EXISTS (SELECT 1 FROM users u
            WHERE strpos(coalesce(u.avatar_url,''),m.id::text)>0)
        ORDER BY m.deleted_at LIMIT 200 FOR UPDATE OF m SKIP LOCKED)`

func removeExpiredUploads(ctx context.Context, db *sql.DB) (int64, error) {
	result, err := db.ExecContext(ctx, expiredUploadSQL)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func sweepMediaLifecycle(ctx context.Context, db *sql.DB) (withdrawn, purged int64, err error) {
	result, err := db.ExecContext(ctx, withdrawAbandonedEvidenceSQL)
	if err != nil {
		return 0, 0, err
	}
	withdrawn, err = result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	result, err = db.ExecContext(ctx, purgeWithdrawnUploadsSQL)
	if err != nil {
		return withdrawn, 0, err
	}
	purged, err = result.RowsAffected()
	return withdrawn, purged, err
}

func reportWalletDrift(ctx context.Context, db *sql.DB) (int, error) {
	rows, err := db.QueryContext(ctx, walletDriftSQL)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var walletID, planetID, userID string
		var balance, ledger int64
		if err := rows.Scan(&walletID, &planetID, &userID, &balance, &ledger); err != nil {
			return count, err
		}
		log.Printf("wallet drift wallet=%s planet=%s user=%s balance=%d ledger=%d", walletID, planetID, userID, balance, ledger)
		count++
	}
	return count, rows.Err()
}

func reconcile(ctx context.Context, db *sql.DB) (int64, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	// A locked jazebe_xact_lock row prevents duplicate runs across Nakama
	// deployments; CockroachDB has no advisory locks. NOWAIT reports a running
	// reconciliation instead of queueing behind it.
	var key int64
	err = tx.QueryRowContext(ctx, `SELECT lock_key FROM jazebe_xact_lock WHERE lock_key = $1 FOR UPDATE NOWAIT`, reconcileLockKey).Scan(&key)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.LockNotAvailable:
		return 0, errors.New("planet fan reconciliation is already running")
	case errors.Is(err, sql.ErrNoRows):
		// First run: the inserted row stays locked until this transaction ends.
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_xact_lock (lock_key) VALUES ($1)`, reconcileLockKey); err != nil {
			return 0, err
		}
	case err != nil:
		return 0, err
	}
	result, err := tx.ExecContext(ctx, reconcileSQL)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func main() {
	dsn := os.Getenv("JAZEBE_DATABASE_URL")
	if dsn == "" {
		log.Fatal("JAZEBE_DATABASE_URL is required")
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	count, err := reconcile(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("reconciled fan counters on %d planets", count)
	drift, err := reportWalletDrift(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("detected wallet balance drift on %d wallets", drift)
	expired, err := removeExpiredUploads(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("removed %d expired incomplete uploads", expired)
	withdrawn, purged, err := sweepMediaLifecycle(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("withdrawn %d abandoned evidence uploads and purged %d unreferenced withdrawn uploads", withdrawn, purged)
}
