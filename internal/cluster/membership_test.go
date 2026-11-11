package cluster

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

type memberLog struct {
	mu    sync.Mutex
	ups   map[string]int
	downs map[string]int
	dead  map[string]int
	reinc int
}

func newMemberLog() *memberLog {
	return &memberLog{ups: map[string]int{}, downs: map[string]int{}, dead: map[string]int{}}
}

func (l *memberLog) events() MembershipEvents {
	return MembershipEvents{
		Up:   func(n NodeInfo) { l.mu.Lock(); l.ups[n.Name]++; l.mu.Unlock() },
		Down: func(n NodeInfo) { l.mu.Lock(); l.downs[n.Name]++; l.mu.Unlock() },
		Dead: func(n NodeInfo) { l.mu.Lock(); l.dead[n.Field()]++; l.mu.Unlock() },
		Reincarnated: func(string, string) {
			l.mu.Lock()
			l.reinc++
			l.mu.Unlock()
		},
	}
}

func (l *memberLog) get(m map[string]int, k string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return m[k]
}

func TestMembershipUpDownDeadAndReap(t *testing.T) {
	rdb, keys := testRedis(t)
	every, ttl := 50*time.Millisecond, 300*time.Millisecond
	ctx := context.Background()
	logs := map[string]*memberLog{}
	ms := map[string]*Membership{}
	for _, name := range []string{"n1", "n2", "n3"} {
		logs[name] = newMemberLog()
		ms[name] = NewMembership(zap.NewNop(), rdb, keys, NodeInfo{Name: name, Inc: NewIncarnation(), Addr: name + ":1"}, every, ttl, logs[name].events())
		if err := ms[name].Start(ctx); err != nil {
			t.Fatal(err)
		}
	}
	eventually(t, 3*time.Second, func() bool {
		for _, m := range ms {
			if len(m.AliveNames()) != 3 {
				return false
			}
		}
		return true
	}, "nodes did not see each other")
	if name, ok := ms["n1"].NodeByHash(NodeHash("n3")); !ok || name != "n3" {
		t.Fatalf("hash lookup failed: %q %v", name, ok)
	}

	ms["n2"].Stop(ctx)
	eventually(t, 2*time.Second, func() bool {
		return logs["n1"].get(logs["n1"].downs, "n2") == 1 && logs["n3"].get(logs["n3"].downs, "n2") == 1
	}, "graceful leave not seen")
	time.Sleep(2 * ttl)
	if logs["n1"].get(logs["n1"].dead, "n2/"+ms["n2"].Self().Inc) != 0 {
		t.Fatal("graceful leave must not need a reap")
	}

	n3 := ms["n3"].Self()
	ms["n3"].KillForTest()
	start := time.Now()
	eventually(t, 3*time.Second, func() bool { return logs["n1"].get(logs["n1"].downs, "n3") == 1 }, "crash not detected")
	if d := time.Since(start); d > ttl+4*every+200*time.Millisecond {
		t.Fatalf("death detection too slow: %v", d)
	}
	eventually(t, time.Second, func() bool { return logs["n1"].get(logs["n1"].dead, n3.Field()) > 0 }, "dead entry not reported")

	holders := 0
	for i := 0; i < 5; i++ {
		ok, err := ms["n1"].TryReap(ctx, n3, "h")
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			holders++
		}
	}
	if holders != 1 {
		t.Fatalf("reap lease must have exactly one holder, got %d", holders)
	}
	if err := ms["n1"].FinishReap(ctx, n3, "h"); err != nil {
		t.Fatal(err)
	}
	if exists, _ := rdb.HExists(ctx, keys.Nodes(), n3.Field()).Result(); exists {
		t.Fatal("reaped node entry must be gone")
	}

	again := NewMembership(zap.NewNop(), rdb, keys, NodeInfo{Name: "n3", Inc: n3.Inc, Addr: "n3:1"}, every, ttl, newMemberLog().events())
	if err := again.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer again.Stop(ctx)
	if again.Self().Inc == n3.Inc {
		t.Fatal("a reaped incarnation must register under a new one")
	}
	eventually(t, 2*time.Second, func() bool { return logs["n1"].get(logs["n1"].ups, "n3") == 2 }, "restarted node not seen")
	ms["n1"].Stop(ctx)
}

func TestMembershipRestartSameNameReplacesIncarnation(t *testing.T) {
	rdb, keys := testRedis(t)
	every, ttl := 50*time.Millisecond, 300*time.Millisecond
	ctx := context.Background()
	watch := newMemberLog()
	w := NewMembership(zap.NewNop(), rdb, keys, NodeInfo{Name: "w", Inc: NewIncarnation()}, every, ttl, watch.events())
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer w.Stop(ctx)
	old := NewMembership(zap.NewNop(), rdb, keys, NodeInfo{Name: "x", Inc: NewIncarnation()}, every, ttl, nil2())
	if err := old.Start(ctx); err != nil {
		t.Fatal(err)
	}
	oldInfo := old.Self()
	time.Sleep(ttl + 2*every)
	eventually(t, 2*time.Second, func() bool { return w.IsAlive("x") }, "x not alive")
	old.KillForTest()
	fresh := NewMembership(zap.NewNop(), rdb, keys, NodeInfo{Name: "x", Inc: NewIncarnation()}, every, ttl, nil2())
	if err := fresh.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer fresh.Stop(ctx)
	eventually(t, 2*time.Second, func() bool { return watch.get(watch.dead, oldInfo.Field()) > 0 }, "old incarnation not reported dead")
	eventually(t, 2*time.Second, func() bool {
		info, ok := w.Info("x")
		return ok && info.Inc == fresh.Self().Inc
	}, "new incarnation not alive")
	if watch.get(watch.downs, "x") != 1 || watch.get(watch.ups, "x") != 2 {
		t.Fatalf("expected one down and two ups, got %d/%d", watch.get(watch.downs, "x"), watch.get(watch.ups, "x"))
	}
}

func nil2() MembershipEvents { return MembershipEvents{} }

func TestMembershipDoesNotJudgeRightAfterStart(t *testing.T) {
	rdb, keys := testRedis(t)
	ctx := context.Background()
	stale := NodeInfo{Name: "ghost", Inc: "dead", Addr: "x"}
	if err := rdb.HSet(ctx, keys.Nodes(), stale.Field(), `{"name":"ghost","inc":"dead","addr":"x"}`).Err(); err != nil {
		t.Fatal(err)
	}
	l := newMemberLog()
	m := NewMembership(zap.NewNop(), rdb, keys, NodeInfo{Name: "m", Inc: NewIncarnation()}, 50*time.Millisecond, 400*time.Millisecond, l.events())
	if err := m.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer m.Stop(ctx)
	time.Sleep(200 * time.Millisecond)
	if l.get(l.dead, stale.Field()) != 0 {
		t.Fatal("must not judge before a full ttl of good heartbeats")
	}
	eventually(t, 2*time.Second, func() bool { return l.get(l.dead, stale.Field()) > 0 }, "stale entry never reported")
}
