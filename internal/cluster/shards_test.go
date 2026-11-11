package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestShardOwnershipAndFencedCommit(t *testing.T) {
	rdb, keys := testRedis(t)
	ctx := context.Background()
	s := NewShards(rdb, keys, 4, 200*time.Millisecond)
	if ok, _ := s.Acquire(ctx, 1, "a/1/x"); !ok {
		t.Fatal("first acquire must win")
	}
	if ok, _ := s.Acquire(ctx, 1, "b/1/y"); ok {
		t.Fatal("second acquire must lose")
	}
	owners, _ := s.Owners(ctx)
	if owners[1] != "a" || owners[0] != "" {
		t.Fatalf("owners %v", owners)
	}
	if ok, _ := s.PutTicket(ctx, 1, "b/1/y", "t1", "{}"); ok {
		t.Fatal("non-owner must not write tickets")
	}
	for _, tk := range []string{"t1", "t2", "t3"} {
		if ok, err := s.PutTicket(ctx, 1, "a/1/x", tk, "{}"); !ok || err != nil {
			t.Fatal("owner write failed", err)
		}
	}
	if res, _, _ := s.Commit(ctx, 1, "a/1/x", []string{"t1", "t2"}); res != CommitOK {
		t.Fatalf("commit %v", res)
	}
	if res, missing, _ := s.Commit(ctx, 1, "a/1/x", []string{"t2", "t3"}); res != CommitMissing || len(missing) != 1 || missing[0] != "t2" {
		t.Fatalf("expected t2 missing, got %v %v", res, missing)
	}
	left, _ := s.Tickets(ctx, 1)
	if len(left) != 1 || left["t3"] != "{}" {
		t.Fatalf("t3 must stay after a failed commit: %v", left)
	}
	time.Sleep(300 * time.Millisecond)
	if ok, _ := s.Acquire(ctx, 1, "b/1/y"); !ok {
		t.Fatal("lease must be free after its ttl")
	}
	if res, _, _ := s.Commit(ctx, 1, "a/1/x", []string{"t3"}); res != CommitLost {
		t.Fatal("old owner must not commit after losing the lease")
	}
	if r, _ := s.Renew(ctx, 1, "a/1/x"); r != RenewLost {
		t.Fatal("old owner renew must report lost")
	}
	_ = rdb.Del(ctx, s.OwnerKey(1))
	if r, _ := s.Renew(ctx, 1, "b/1/y"); r != RenewReclaimed {
		t.Fatal("a vanished lease must be reclaimed by its holder")
	}
	if p := Preferred(3, []string{"n1", "n2", "n3"}); p != Preferred(3, []string{"n3", "n1", "n2"}) {
		t.Fatal("preference must not depend on order")
	}
}

func TestShardExactlyOnceWithFailovers(t *testing.T) {
	rdb, keys := testRedis(t)
	ctx := context.Background()
	s := NewShards(rdb, keys, 1, 60*time.Millisecond)
	const tickets = 300
	values := map[string]string{}
	for i := 0; i < tickets; i++ {
		values[fmt.Sprintf("t%03d", i)] = "{}"
	}
	if err := s.PutTickets(ctx, 0, values); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	committed := map[string]int{}
	removed := map[string]bool{}
	stop := time.Now().Add(1500 * time.Millisecond)
	wg := sync.WaitGroup{}
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(n)))
			for time.Now().Before(stop) {
				token := fmt.Sprintf("n%d/%d/%d", n, n, r.Int63())
				if ok, _ := s.Acquire(ctx, 0, token); !ok {
					time.Sleep(5 * time.Millisecond)
					continue
				}
				owned, _ := s.Tickets(ctx, 0)
				pool := make([]string, 0, len(owned))
				for k := range owned {
					pool = append(pool, k)
				}
				r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
				for len(pool) >= 2 && time.Now().Before(stop) {
					group := pool[:2]
					pool = pool[2:]
					if r.Intn(10) == 0 {
						time.Sleep(80 * time.Millisecond)
					}
					if r.Intn(8) == 0 {
						if n, _ := rdb.HDel(ctx, s.TicketsKey(0), group[0]).Result(); n == 1 {
							mu.Lock()
							removed[group[0]] = true
							mu.Unlock()
						}
					}
					res, _, err := s.Commit(ctx, 0, token, group)
					if err != nil || res == CommitLost {
						break
					}
					if res == CommitOK {
						mu.Lock()
						for _, g := range group {
							committed[g]++
						}
						mu.Unlock()
					}
				}
				_ = s.Release(ctx, 0, token)
			}
		}(n)
	}
	wg.Wait()
	if len(committed) == 0 {
		t.Fatal("nothing was committed")
	}
	for tk, n := range committed {
		if n != 1 {
			t.Fatalf("ticket %s committed %d times", tk, n)
		}
		if removed[tk] {
			t.Fatalf("removed ticket %s was committed", tk)
		}
	}
	t.Logf("committed %d tickets, removed %d", len(committed), len(removed))
}

func TestPreferredSpreadsShards(t *testing.T) {
	for _, n := range []int{3, 4, 30} {
		names := make([]string, n)
		for i := range names {
			names[i] = fmt.Sprintf("node%d", i+1)
		}
		counts := map[string]int{}
		shards := 16 * n
		for i := 0; i < shards; i++ {
			counts[Preferred(i, names)]++
		}
		for _, name := range names {
			if c := counts[name]; c < 16/3 || c > 16*3 {
				t.Fatalf("%d nodes: %s prefers %d of %d shards", n, name, c, shards)
			}
		}
	}
}
