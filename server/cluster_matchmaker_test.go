package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"go.uber.org/zap"
)

func (n *clusterTestNode) matchmaker() Matchmaker {
	if n.cluster.matchmaker != nil {
		return n.cluster.matchmaker
	}
	return NewClusterAwareMatchmaker(n.cluster, zap.NewNop(), zap.NewNop(), n.config, n.router, metrics, &Runtime{})
}

func isMatched(e *rtapi.Envelope) bool { return e.GetMatchmakerMatched() != nil }

func waitShardsOwned(t *testing.T, nodes ...*clusterTestNode) {
	clusterEventually(t, 10*time.Second, func() bool {
		total := 0
		for _, n := range nodes {
			if n.cluster.matchmaker != nil {
				total += len(n.cluster.matchmaker.ownedShards())
			}
		}
		if total != nodes[0].config.Cluster.MatchmakerShards {
			return false
		}
		for _, n := range nodes {
			for i := 0; i < n.config.Cluster.MatchmakerShards; i++ {
				if n.cluster.matchmaker.ownerOf(i) == "" {
					return false
				}
			}
		}
		return true
	}, "matchmaker shards not owned")
}

func TestClusterMatchmakerMatchesEachTicketOnceAcrossNodesAndFailover(t *testing.T) {
	prefix := testPrefix()
	var nodes []*clusterTestNode
	for _, name := range []string{"ma", "mb", "mc"} {
		n := newClusterTestNode(t, name, prefix)
		n.config.Cluster.MatchmakerShards = 4
		n.config.Cluster.MatchmakerShardProperty = "mode"
		nodes = append(nodes, n)
	}
	waitCluster(t, nodes...)
	for _, n := range nodes {
		n.matchmaker()
	}
	waitShardsOwned(t, nodes...)

	type player struct {
		n *clusterTestNode
		s *clusterTestSession
	}
	var players []player
	for i := 0; i < 24; i++ {
		n := nodes[i%3]
		players = append(players, player{n: n, s: n.connect(uuid.Must(uuid.NewV4()), fmt.Sprintf("p%d", i), SessionFormatJson)})
	}
	add := func(p player, mode string) error {
		_, _, err := p.n.cluster.matchmaker.Add(context.Background(), []*MatchmakerPresence{{UserId: p.s.userID.String(), SessionId: p.s.id.String(), Username: p.s.username, Node: p.n.name, SessionID: p.s.id}}, p.s.id.String(), "", "*", 2, 2, 1, map[string]string{"mode": mode}, map[string]float64{})
		return err
	}
	for round := 1; round <= 3; round++ {
		for i, p := range players {
			if err := add(p, fmt.Sprintf("m%d", i%2)); err != nil {
				t.Fatalf("round %d add: %v", round, err)
			}
		}
		clusterEventually(t, 10*time.Second, func() bool {
			for _, p := range players {
				if p.s.count(isMatched) < round {
					return false
				}
			}
			return true
		}, "round %d: not every player matched", round)
		time.Sleep(1500 * time.Millisecond)
		for _, p := range players {
			if got := p.s.count(isMatched); got != round {
				t.Fatalf("round %d: player matched %d times", round, got)
			}
		}
	}

	victim := nodes[1]
	if len(victim.cluster.matchmaker.ownedShards()) == 0 {
		victim = nodes[0]
		if len(victim.cluster.matchmaker.ownedShards()) == 0 {
			victim = nodes[2]
		}
	}
	var survivors []*clusterTestNode
	for _, n := range nodes {
		if n != victim {
			survivors = append(survivors, n)
		}
	}
	var alive []player
	for _, p := range players {
		if p.n != victim {
			alive = append(alive, p)
		}
	}
	for i, p := range alive {
		if err := add(p, fmt.Sprintf("m%d", i%2)); err != nil {
			t.Fatalf("add before kill: %v", err)
		}
	}
	victim.kill()
	clusterEventually(t, 20*time.Second, func() bool {
		for _, p := range alive {
			if p.s.count(isMatched) < 4 {
				return false
			}
		}
		return true
	}, "players not matched after the shard owner died")
	time.Sleep(2 * time.Second)
	for _, p := range alive {
		if got := p.s.count(isMatched); got != 4 {
			t.Fatalf("after failover a player was matched %d times", got)
		}
	}
	waitShardsOwned(t, survivors...)

	p := alive[0]
	ticket, _, err := p.n.cluster.matchmaker.Add(context.Background(), []*MatchmakerPresence{{UserId: p.s.userID.String(), SessionId: p.s.id.String(), Username: "x", Node: p.n.name, SessionID: p.s.id}}, p.s.id.String(), "", "*", 2, 2, 1, map[string]string{"mode": "solo"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.n.cluster.matchmaker.RemoveSession(p.s.id.String(), ticket); err != nil {
		t.Fatal(err)
	}
	q := alive[1]
	if err := add(q, "solo"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2500 * time.Millisecond)
	if p.s.count(isMatched) != 4 || q.s.count(isMatched) != 4 {
		t.Fatal("a removed ticket was matched")
	}
	q.s.Close("bye", runtime.PresenceReasonDisconnect)
	_ = q.n.cluster.matchmaker.RemoveSessionAll(q.s.id.String())
}
