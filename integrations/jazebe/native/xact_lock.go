package main

import (
	"context"
	"database/sql"
	"hash/fnv"
)

// lockXact serializes transactions that use the same key until the caller's
// transaction ends. It replaces pg_advisory_xact_lock, which CockroachDB does
// not implement: the upsert takes a row write lock on PostgreSQL and
// CockroachDB alike. As with hashtext, a hash collision only adds waiting.
func lockXact(ctx context.Context, tx *sql.Tx, key string) error {
	h := fnv.New64a()
	h.Write([]byte(key))
	_, err := tx.ExecContext(ctx, `INSERT INTO jazebe_xact_lock (lock_key) VALUES ($1)
		ON CONFLICT (lock_key) DO UPDATE SET locked_at = now()`, int64(h.Sum64()))
	return err
}
