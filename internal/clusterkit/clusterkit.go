package clusterkit

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

type Queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

var processID = func() string {
	raw := make([]byte, 4)
	_, _ = rand.Read(raw)
	return hex.EncodeToString(raw)
}()

func Holder(ctx context.Context) string {
	node, _ := ctx.Value(runtime.RUNTIME_CTX_NODE).(string)
	if node = strings.TrimSpace(node); node == "" {
		node, _ = os.Hostname()
	}
	return node + "/" + processID
}

func millis(d time.Duration) int64 {
	if d < time.Millisecond {
		return 1
	}
	return d.Milliseconds()
}

func AcquireLease(ctx context.Context, q Queryer, name, holder string, ttl time.Duration) (bool, error) {
	if name == "" || holder == "" {
		return false, errors.New("lease name and holder are required")
	}
	var got string
	err := q.QueryRowContext(ctx, `INSERT INTO cluster_lease (name, holder, expires_at) VALUES ($1, $2, now() + $3::INT8 * INTERVAL '1 millisecond')
		ON CONFLICT (name) DO UPDATE SET holder = EXCLUDED.holder, expires_at = EXCLUDED.expires_at, updated_at = now()
		WHERE cluster_lease.holder = EXCLUDED.holder OR cluster_lease.expires_at <= now()
		RETURNING holder`, name, holder, millis(ttl)).Scan(&got)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return got == holder, nil
}

func ReleaseLease(ctx context.Context, q Queryer, name, holder string) error {
	_, err := q.ExecContext(ctx, `DELETE FROM cluster_lease WHERE name = $1 AND holder = $2`, name, holder)
	return err
}

type Decision struct {
	Allowed    bool
	Count      int64
	Limit      int64
	RetryAfter time.Duration
}

func Allow(ctx context.Context, q Queryer, key string, limit int64, window time.Duration) (Decision, error) {
	d := Decision{Limit: limit}
	if key == "" || limit < 1 || window <= 0 {
		return d, errors.New("rate limit key, limit and window are required")
	}
	var left int64
	err := q.QueryRowContext(ctx, `INSERT INTO cluster_rate_limit (key, window_start, count, expires_at) VALUES ($1, now(), 1, now() + $2::INT8 * INTERVAL '1 millisecond')
		ON CONFLICT (key) DO UPDATE SET
			window_start = CASE WHEN cluster_rate_limit.expires_at <= now() THEN now() ELSE cluster_rate_limit.window_start END,
			count = CASE WHEN cluster_rate_limit.expires_at <= now() THEN 1 ELSE cluster_rate_limit.count + 1 END,
			expires_at = CASE WHEN cluster_rate_limit.expires_at <= now() THEN EXCLUDED.expires_at ELSE cluster_rate_limit.expires_at END
		RETURNING count, GREATEST(0, CEIL(EXTRACT(EPOCH FROM (expires_at - now())) * 1000))::INT8`, key, millis(window)).Scan(&d.Count, &left)
	if err != nil {
		return d, err
	}
	d.Allowed = d.Count <= limit
	if !d.Allowed {
		d.RetryAfter = time.Duration(left) * time.Millisecond
	}
	return d, nil
}

func SweepRateLimits(ctx context.Context, q Queryer, olderThan time.Duration) (int64, error) {
	res, err := q.ExecContext(ctx, `DELETE FROM cluster_rate_limit WHERE expires_at < now() - $1::INT8 * INTERVAL '1 millisecond'`, millis(olderThan))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
