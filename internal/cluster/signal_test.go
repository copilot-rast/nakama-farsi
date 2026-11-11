package cluster

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSignalsOffSleepAndOnWake(t *testing.T) {
	start := time.Now()
	if WaitSignal(context.Background(), "x", 60*time.Millisecond) || time.Since(start) < 55*time.Millisecond {
		t.Fatal("without a cluster a wait must sleep the whole time")
	}
	rdb, keys := testRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := NewRedisSignals(ctx, zap.NewNop(), rdb, keys)
	b := NewRedisSignals(ctx, zap.NewNop(), rdb, keys)
	time.Sleep(200 * time.Millisecond)
	got := make(chan bool, 1)
	go func() { got <- a.Wait(context.Background(), "chat:1", 3*time.Second) }()
	time.Sleep(50 * time.Millisecond)
	start = time.Now()
	b.Signal("chat:1")
	if !<-got || time.Since(start) > time.Second {
		t.Fatal("a signal from another node must wake the waiter at once")
	}
	if a.Wait(context.Background(), "chat:2", 100*time.Millisecond) {
		t.Fatal("other topics must not wake")
	}
}

func TestInvalidationLogReplaysAndTails(t *testing.T) {
	rdb, keys := testRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l := NewInvalidationLog(zap.NewNop(), rdb, keys.Invalidations(false), time.Hour, "n1")
	if err := l.Append(ctx, Invalidation{Op: "removeall", UserIDs: []string{"u1"}}); err != nil {
		t.Fatal(err)
	}
	got := make(chan Invalidation, 10)
	go NewInvalidationLog(zap.NewNop(), rdb, keys.Invalidations(false), time.Hour, "n2").Run(ctx, func(inv Invalidation) { got <- inv })
	first := <-got
	if first.Op != "removeall" || len(first.UserIDs) != 1 || first.Origin != "n1" {
		t.Fatalf("replay wrong: %+v", first)
	}
	_ = l.Append(ctx, Invalidation{Op: "ban", UserIDs: []string{"u2", "u3"}})
	select {
	case inv := <-got:
		if inv.Op != "ban" || len(inv.UserIDs) != 2 {
			t.Fatalf("tail wrong: %+v", inv)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("new record not tailed")
	}
}
