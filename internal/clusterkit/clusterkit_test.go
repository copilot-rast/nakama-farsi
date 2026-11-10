package clusterkit

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const disposableTestDBEnv = "CLUSTERKIT_DISPOSABLE_TEST_DB_URL"

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv(disposableTestDBEnv)
	if dsn == "" {
		t.Skip(disposableTestDBEnv + " not set; PostgreSQL NOT_RUN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	var name string
	if err := db.QueryRow(`SELECT current_database()`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, "_disposable") {
		t.Fatalf("database %q is not disposable", name)
	}
	return db
}

func TestHolderNamesNodeAndProcess(t *testing.T) {
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_NODE, "nakama-a")
	h := Holder(ctx)
	if !strings.HasPrefix(h, "nakama-a/") || len(h) != len("nakama-a/")+8 || Holder(ctx) != h {
		t.Fatalf("holder = %q", h)
	}
}

func TestLeaseOneHolderAtATime(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	name := fmt.Sprintf("test-lease-%d", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM cluster_lease WHERE name=$1`, name) })
	var winners atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ok, err := AcquireLease(ctx, db, name, fmt.Sprintf("node-%d", i), time.Minute)
			if err != nil {
				t.Error(err)
			}
			if ok {
				winners.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("%d nodes hold the lease", winners.Load())
	}
	var holder string
	if err := db.QueryRow(`SELECT holder FROM cluster_lease WHERE name=$1`, name).Scan(&holder); err != nil {
		t.Fatal(err)
	}
	if ok, _ := AcquireLease(ctx, db, name, holder, time.Minute); !ok {
		t.Fatal("holder cannot renew its lease")
	}
	other := "node-x"
	if ok, _ := AcquireLease(ctx, db, name, other, time.Minute); ok {
		t.Fatal("live lease taken by another node")
	}
	if err := ReleaseLease(ctx, db, name, other); err != nil {
		t.Fatal(err)
	}
	if ok, _ := AcquireLease(ctx, db, name, other, time.Minute); ok {
		t.Fatal("release by a non-holder freed the lease")
	}
	if err := ReleaseLease(ctx, db, name, holder); err != nil {
		t.Fatal(err)
	}
	if ok, _ := AcquireLease(ctx, db, name, other, 50*time.Millisecond); !ok {
		t.Fatal("released lease not free")
	}
	time.Sleep(120 * time.Millisecond)
	if ok, _ := AcquireLease(ctx, db, name, holder, time.Minute); !ok {
		t.Fatal("expired lease not taken over")
	}
}

func TestRateLimitIsSharedAndAtomic(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	key := fmt.Sprintf("test-limit-%d", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM cluster_rate_limit WHERE key=$1`, key) })
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d, err := Allow(ctx, db, key, 10, 300*time.Millisecond)
			if err != nil {
				t.Error(err)
				return
			}
			if d.Allowed {
				allowed.Add(1)
			} else if d.RetryAfter <= 0 || d.RetryAfter > 300*time.Millisecond {
				t.Errorf("retry after = %v", d.RetryAfter)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 10 {
		t.Fatalf("%d of 40 allowed, want 10", allowed.Load())
	}
	time.Sleep(350 * time.Millisecond)
	d, err := Allow(ctx, db, key, 10, 300*time.Millisecond)
	if err != nil || !d.Allowed || d.Count != 1 {
		t.Fatalf("new window = %+v %v", d, err)
	}
	if _, err := db.Exec(`UPDATE cluster_rate_limit SET expires_at = now() - interval '2 hours' WHERE key=$1`, key); err != nil {
		t.Fatal(err)
	}
	if n, err := SweepRateLimits(ctx, db, time.Hour); err != nil || n < 1 {
		t.Fatalf("sweep = %d %v", n, err)
	}
	if _, err := Allow(ctx, db, "", 1, time.Second); err == nil {
		t.Fatal("empty key accepted")
	}
}
