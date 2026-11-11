package cluster

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func testRedis(t *testing.T) (redis.UniversalClient, Keys) {
	t.Helper()
	addr := os.Getenv("CLUSTER_TEST_REDIS")
	if addr == "" {
		t.Skip("CLUSTER_TEST_REDIS not set")
	}
	rdb := NewRedis(RedisOptions{Addresses: addr, DB: 15})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not reachable: %v", err)
	}
	keys := Keys{Prefix: fmt.Sprintf("t%d", time.Now().UnixNano())}
	t.Cleanup(func() {
		ctx := context.Background()
		iter := rdb.Scan(ctx, 0, keys.Prefix+":*", 1000).Iterator()
		for iter.Next(ctx) {
			rdb.Del(ctx, iter.Val())
		}
		_ = rdb.Close()
	})
	return rdb, keys
}

func eventually(t *testing.T, d time.Duration, fn func() bool, msg string, args ...any) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf(msg, args...)
}
