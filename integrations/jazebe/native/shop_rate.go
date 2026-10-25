package main

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type shopRateContextKey struct{}

func shopRateAlreadyChecked(ctx context.Context) bool {
	checked, _ := ctx.Value(shopRateContextKey{}).(bool)
	return checked
}

func shopRateConfig() (int, int, error) {
	limit, window := 20, 60
	for _, item := range []struct {
		name string
		out  *int
	}{{"RATE_LIMIT_PURCHASE", &limit}, {"RATE_LIMIT_WINDOW_SECONDS", &window}} {
		raw := strings.TrimSpace(os.Getenv(item.name))
		if raw == "" {
			continue
		}
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100000 {
			return 0, 0, status.Error(codes.InvalidArgument, "invalid shop rate configuration")
		}
		*item.out = value
	}
	return limit, window, nil
}

// A row-level UPSERT enforces the shared purchase budget before any debit.
// Replays and failed attempts count, matching the source route guard.
func shopPurchaseRate(ctx context.Context, db *sql.DB, user string) error {
	limit, window, err := shopRateConfig()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	var count int
	var started time.Time
	err = db.QueryRowContext(ctx, `INSERT INTO jazebe_shop_purchase_rate(user_id,window_started_at,count) VALUES($1,$2,1) ON CONFLICT(user_id) DO UPDATE SET window_started_at=CASE WHEN jazebe_shop_purchase_rate.window_started_at<=$2::timestamptz-$3::INT8*INTERVAL '1 second' THEN $2 ELSE jazebe_shop_purchase_rate.window_started_at END,count=CASE WHEN jazebe_shop_purchase_rate.window_started_at<=$2::timestamptz-$3::INT8*INTERVAL '1 second' THEN 1 ELSE jazebe_shop_purchase_rate.count+1 END RETURNING count,window_started_at`, user, now, window).Scan(&count, &started)
	if err != nil {
		return status.Error(codes.Internal, "purchase rate limit unavailable")
	}
	if count > limit {
		return status.Error(codes.ResourceExhausted, "purchase rate limit reached")
	}
	return nil
}
