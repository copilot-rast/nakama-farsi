package server

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
)

func channelJoins(s *clusterTestSession, user uuid.UUID) int {
	return s.count(func(e *rtapi.Envelope) bool {
		ev := e.GetChannelPresenceEvent()
		if ev == nil {
			return false
		}
		for _, j := range ev.Joins {
			if j.UserId == user.String() {
				return true
			}
		}
		return false
	})
}

func channelLeaves(s *clusterTestSession, user uuid.UUID) int {
	return s.count(func(e *rtapi.Envelope) bool {
		ev := e.GetChannelPresenceEvent()
		if ev == nil {
			return false
		}
		for _, j := range ev.Leaves {
			if j.UserId == user.String() {
				return true
			}
		}
		return false
	})
}

func statusEvents(s *clusterTestSession, user uuid.UUID, leave bool, status string) int {
	return s.count(func(e *rtapi.Envelope) bool {
		ev := e.GetStatusPresenceEvent()
		if ev == nil {
			return false
		}
		list := ev.Joins
		if leave {
			list = ev.Leaves
		}
		for _, j := range list {
			if j.UserId == user.String() && (status == "" || j.GetStatus().GetValue() == status) {
				return true
			}
		}
		return false
	})
}

func presenceUsers(ps []*Presence) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.UserID.String())
	}
	sort.Strings(out)
	return out
}

func TestClusterPresenceJoinLeaveAndListing(t *testing.T) {
	prefix := testPrefix()
	a := newClusterTestNode(t, "pa", prefix)
	b := newClusterTestNode(t, "pb", prefix)
	waitCluster(t, a, b)
	ua, ub := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	sa := a.connect(ua, "a", SessionFormatJson)
	sb := b.connect(ub, "b", SessionFormatJson)
	room := PresenceStream{Mode: StreamModeChannel, Label: "lobby"}
	ctx := context.Background()
	a.tracker.Track(ctx, sa.id, room, ua, PresenceMeta{Username: "a"})
	clusterEventually(t, 2*time.Second, func() bool { return len(b.tracker.ListByStream(room, false, true)) == 1 }, "b cannot list a's presence")
	b.tracker.Track(ctx, sb.id, room, ub, PresenceMeta{Username: "b"})
	clusterEventually(t, 2*time.Second, func() bool { return channelJoins(sa, ub) == 1 }, "a did not see b join")
	clusterEventually(t, 2*time.Second, func() bool {
		return a.tracker.CountByStream(room) == 2 && b.tracker.CountByStream(room) == 2
	}, "counts differ: %d %d", a.tracker.CountByStream(room), b.tracker.CountByStream(room))
	if got := presenceUsers(a.tracker.ListByStream(room, false, true)); len(got) != 2 {
		t.Fatalf("a lists %v", got)
	}
	if !a.tracker.StreamExists(PresenceStream{Mode: StreamModeChannel, Label: "lobby"}) || a.tracker.StreamExists(PresenceStream{Mode: StreamModeChannel, Label: "nobody"}) {
		t.Fatal("stream exists check wrong")
	}
	ids := a.tracker.ListPresenceIDByStream(room)
	if len(ids) != 2 {
		t.Fatalf("presence ids %v", ids)
	}

	b.tracker.Untrack(sb.id, room, ub)
	clusterEventually(t, 2*time.Second, func() bool { return channelLeaves(sa, ub) == 1 }, "a did not see b leave")
	clusterEventually(t, 2*time.Second, func() bool { return a.tracker.CountByStream(room) == 1 }, "a still counts b")

	b.tracker.Track(ctx, sb.id, room, ub, PresenceMeta{Username: "b"})
	clusterEventually(t, 2*time.Second, func() bool { return channelJoins(sa, ub) == 2 }, "rejoin not seen")
	sb.Close("bye", runtime.PresenceReasonDisconnect)
	clusterEventually(t, 2*time.Second, func() bool { return channelLeaves(sa, ub) == 2 }, "disconnect leave not seen")

	sa.Close("bye", runtime.PresenceReasonDisconnect)
	clusterEventually(t, 2*time.Second, func() bool {
		n, _ := a.cluster.rdb.Exists(ctx, a.cluster.tracker.store.P(streamRef(room))).Result()
		return n == 0
	}, "room key not removed when empty")
}

func TestClusterStatusFollowOnlineAndUpdate(t *testing.T) {
	prefix := testPrefix()
	a := newClusterTestNode(t, "sa", prefix)
	b := newClusterTestNode(t, "sb", prefix)
	waitCluster(t, a, b)
	ua, ub := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	sa := a.connect(ua, "a", SessionFormatJson)
	a.status.Follow(sa.id, map[uuid.UUID]struct{}{ub: {}})
	clusterEventually(t, 2*time.Second, func() bool {
		v, _ := a.cluster.rdb.HGet(context.Background(), a.cluster.tracker.store.FollowersKey("u:"+ub.String()), "sa").Result()
		return v == "1"
	}, "follow not stored")
	if a.status.IsOnline(ub) {
		t.Fatal("b must be offline")
	}
	sb := b.connect(ub, "b", SessionFormatJson)
	statusStream := PresenceStream{Mode: StreamModeStatus, Subject: ub}
	b.tracker.TrackMulti(context.Background(), sb.id, []*TrackerOp{
		{Stream: PresenceStream{Mode: StreamModeNotifications, Subject: ub}, Meta: PresenceMeta{Username: "b", Hidden: true}},
		{Stream: statusStream, Meta: PresenceMeta{Username: "b"}},
	}, ub)
	clusterEventually(t, 2*time.Second, func() bool { return statusEvents(sa, ub, false, "") == 1 }, "a did not get b online")
	clusterEventually(t, 2*time.Second, func() bool { return a.status.IsOnline(ub) }, "b not online from a")
	users := []*api.User{{Id: ub.String()}, {Id: ua.String()}}
	a.status.FillOnlineUsers(users)
	if !users[0].Online || users[1].Online {
		t.Fatalf("online flags wrong: %v %v", users[0].Online, users[1].Online)
	}
	b.tracker.Update(context.Background(), sb.id, statusStream, ub, PresenceMeta{Username: "b", Status: "playing"})
	clusterEventually(t, 2*time.Second, func() bool { return statusEvents(sa, ub, false, "playing") == 1 }, "status text not delivered")
	ps := a.tracker.ListByStream(statusStream, false, true)
	if len(ps) != 1 || ps[0].GetStatus() != "playing" {
		t.Fatalf("status listing wrong: %v", ps)
	}
	sb.Close("bye", runtime.PresenceReasonDisconnect)
	clusterEventually(t, 2*time.Second, func() bool { return statusEvents(sa, ub, true, "playing") >= 1 }, "offline not delivered")
	clusterEventually(t, 2*time.Second, func() bool { return !a.status.IsOnline(ub) }, "b still online")
	a.status.UnfollowAll(sa.id)
	clusterEventually(t, 2*time.Second, func() bool {
		n, _ := a.cluster.rdb.HLen(context.Background(), a.cluster.tracker.store.FollowersKey("u:"+ub.String())).Result()
		return n == 0
	}, "follow not removed")
}

func TestClusterNodeDeathSendsLeavesAndCleansRedis(t *testing.T) {
	prefix := testPrefix()
	a := newClusterTestNode(t, "da", prefix)
	b := newClusterTestNode(t, "db", prefix)
	waitCluster(t, a, b)
	ua, ub := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	sa := a.connect(ua, "a", SessionFormatJson)
	sb := b.connect(ub, "b", SessionFormatJson)
	ctx := context.Background()
	room := PresenceStream{Mode: StreamModeChannel, Label: "deathroom"}
	match := PresenceStream{Mode: StreamModeMatchAuthoritative, Subject: uuid.Must(uuid.NewV4()), Label: "db"}
	a.tracker.Track(ctx, sa.id, room, ua, PresenceMeta{Username: "a"})
	a.tracker.Track(ctx, sa.id, match, ua, PresenceMeta{Username: "a"})
	b.tracker.Track(ctx, sb.id, room, ub, PresenceMeta{Username: "b"})
	a.status.Follow(sa.id, map[uuid.UUID]struct{}{ub: {}})
	b.tracker.Track(ctx, sb.id, PresenceStream{Mode: StreamModeStatus, Subject: ub}, ub, PresenceMeta{Username: "b"})
	clusterEventually(t, 2*time.Second, func() bool { return channelJoins(sa, ub) == 1 && statusEvents(sa, ub, false, "") == 1 }, "setup events missing")

	start := time.Now()
	b.kill()
	clusterEventually(t, 5*time.Second, func() bool { return channelLeaves(sa, ub) == 1 }, "no leave event after node death")
	took := time.Since(start)
	clusterEventually(t, 2*time.Second, func() bool { return statusEvents(sa, ub, true, "") == 1 }, "no offline event after node death")
	clusterEventually(t, 2*time.Second, func() bool {
		return sa.count(func(e *rtapi.Envelope) bool {
			n := e.GetNotifications()
			return n != nil && len(n.Notifications) == 1 && n.Notifications[0].Code == NotificationCodeMatchEnded
		}) == 1
	}, "no match_ended notice for the match hosted on the dead node")
	if a.tracker.GetLocalBySessionIDStreamUserID(sa.id, match, ua) != nil {
		t.Fatal("local presence in the lost match must be removed")
	}
	if a.tracker.CountByStream(room) != 1 || a.status.IsOnline(ub) {
		t.Fatal("dead node's presence still visible")
	}
	clusterEventually(t, 3*time.Second, func() bool {
		n, _ := a.cluster.rdb.HExists(ctx, a.cluster.keys.Nodes(), b.cluster.members.Self().Field()).Result()
		return !n
	}, "dead node entry not reaped")
	t.Logf("leave event %v after kill", took)
}

func TestClusterViewsStayConsistentUnderChurn(t *testing.T) {
	prefix := testPrefix()
	nodes := []*clusterTestNode{newClusterTestNode(t, "va", prefix), newClusterTestNode(t, "vb", prefix), newClusterTestNode(t, "vc", prefix)}
	waitCluster(t, nodes...)
	streams := make([]PresenceStream, 4)
	for i := range streams {
		streams[i] = PresenceStream{Mode: StreamModeChannel, Label: fmt.Sprintf("churn%d", i)}
	}
	type member struct {
		n *clusterTestNode
		s *clusterTestSession
	}
	var members []member
	for _, n := range nodes {
		for i := 0; i < 20; i++ {
			u := uuid.Must(uuid.NewV4())
			members = append(members, member{n: n, s: n.connect(u, "u", SessionFormatJson)})
		}
	}
	wg := sync.WaitGroup{}
	for _, m := range members {
		wg.Add(1)
		go func(m member) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(m.s.id[15])))
			for i := 0; i < 30; i++ {
				s := streams[r.Intn(len(streams))]
				if r.Intn(2) == 0 {
					m.n.tracker.Track(context.Background(), m.s.id, s, m.s.userID, PresenceMeta{Username: "u"})
				} else {
					m.n.tracker.Untrack(m.s.id, s, m.s.userID)
				}
				time.Sleep(time.Duration(r.Intn(3)) * time.Millisecond)
			}
		}(m)
	}
	wg.Wait()
	store := cluster.NewPresenceStore(nodes[0].cluster.rdb, nodes[0].cluster.keys)
	clusterEventually(t, 5*time.Second, func() bool {
		for _, s := range streams {
			lists, err := store.List(context.Background(), []cluster.StreamRef{streamRef(s)})
			if err != nil {
				return false
			}
			want := len(lists[0])
			for _, n := range nodes {
				if n.tracker.CountByStream(s) != want {
					return false
				}
				localCount := n.cluster.tracker.local.CountByStream(s)
				if localCount > 0 {
					n.cluster.tracker.mu.Lock()
					v := n.cluster.tracker.views[s]
					ok := v != nil && v.ready && len(v.remote)+localCount == want
					n.cluster.tracker.mu.Unlock()
					if !ok {
						return false
					}
				}
			}
		}
		return true
	}, "views diverged from Redis")
	for _, m := range members {
		m.s.Close("bye", runtime.PresenceReasonDisconnect)
	}
	clusterEventually(t, 5*time.Second, func() bool {
		for _, s := range streams {
			if n, _ := store.Count(context.Background(), streamRef(s)); n != 0 {
				return false
			}
		}
		return true
	}, "presences left in Redis after everyone left")
}
