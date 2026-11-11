package server

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type ClusterStatusRegistry struct {
	c     *Cluster
	local *LocalStatusRegistry
	store *cluster.PresenceStore
	ops   chan followChange
	dirty *atomic.Bool
}

type followChange struct {
	added   []uuid.UUID
	removed []uuid.UUID
	done    chan struct{}
}

var _ StatusRegistry = (*ClusterStatusRegistry)(nil)

func newClusterStatusRegistry(c *Cluster, local *LocalStatusRegistry) *ClusterStatusRegistry {
	s := &ClusterStatusRegistry{c: c, local: local, store: cluster.NewPresenceStore(c.rdb, c.keys), ops: make(chan followChange, 65536), dirty: atomic.NewBool(false)}
	local.followHook = func(added, removed []uuid.UUID) {
		s.ops <- followChange{added: added, removed: removed}
	}
	return s
}

func (s *ClusterStatusRegistry) start() {
	go s.work()
}

func (s *ClusterStatusRegistry) indexKey() string {
	return s.c.keys.NodeFollows(s.c.members.Self().Field())
}

func (s *ClusterStatusRegistry) work() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		var first followChange
		select {
		case <-s.c.ctx.Done():
			return
		case <-ticker.C:
			if s.dirty.CompareAndSwap(true, false) {
				s.republish(s.c.ctx)
			}
			continue
		case first = <-s.ops:
		}
		batch := []followChange{first}
	fill:
		for len(batch) < 512 {
			select {
			case ch := <-s.ops:
				batch = append(batch, ch)
			default:
				break fill
			}
		}
		var added, removed []string
		final := make(map[string]bool, len(added)+len(removed))
		order := make([]string, 0, len(added)+len(removed))
		for _, ch := range batch {
			for _, u := range ch.added {
				k := u.String()
				if _, ok := final[k]; !ok {
					order = append(order, k)
				}
				final[k] = true
			}
			for _, u := range ch.removed {
				k := u.String()
				if _, ok := final[k]; !ok {
					order = append(order, k)
				}
				final[k] = false
			}
		}
		added, removed = added[:0], removed[:0]
		for _, k := range order {
			if final[k] {
				added = append(added, k)
			} else {
				removed = append(removed, k)
			}
		}
		ctx, cancel := context.WithTimeout(s.c.ctx, 5*time.Second)
		if err := s.store.Follow(ctx, s.c.name, s.indexKey(), added, removed); err != nil {
			s.c.logger.Error("Cluster follow write failed", zap.Error(err))
			s.dirty.Store(true)
		}
		cancel()
		for _, ch := range batch {
			if ch.done != nil {
				close(ch.done)
			}
		}
	}
}

func (s *ClusterStatusRegistry) followedUsers() []string {
	s.local.RLock()
	out := make([]string, 0, len(s.local.byUser))
	for u := range s.local.byUser {
		out = append(out, u.String())
	}
	s.local.RUnlock()
	return out
}

func (s *ClusterStatusRegistry) republish(ctx context.Context) {
	users := s.followedUsers()
	for start := 0; start < len(users); start += 1000 {
		if err := s.store.Follow(ctx, s.c.name, s.indexKey(), users[start:min(start+1000, len(users))], nil); err != nil {
			s.dirty.Store(true)
			return
		}
	}
}

func (s *ClusterStatusRegistry) leaveAll(ctx context.Context) {
	users := s.followedUsers()
	for start := 0; start < len(users); start += 1000 {
		_ = s.store.Follow(ctx, s.c.name, "", nil, users[start:min(start+1000, len(users))])
	}
	_ = s.store.Del(ctx, s.indexKey())
}

func (s *ClusterStatusRegistry) reapFollows(ctx context.Context, info cluster.NodeInfo) {
	key := s.c.keys.NodeFollows(info.Field())
	if !s.c.members.IsAlive(info.Name) {
		_ = s.store.ScanSet(ctx, key, 1000, func(users []string) error {
			return s.store.Follow(ctx, info.Name, "", nil, users)
		})
	}
	_ = s.store.Del(ctx, key)
}

func (s *ClusterStatusRegistry) Stop() {
	s.local.Stop()
}

func (s *ClusterStatusRegistry) Follow(sessionID uuid.UUID, userIDs map[uuid.UUID]struct{}) {
	s.local.Follow(sessionID, userIDs)
	done := make(chan struct{})
	select {
	case s.ops <- followChange{done: done}:
	default:
		return
	}
	select {
	case <-done:
	case <-s.c.ctx.Done():
	case <-time.After(2 * time.Second):
	}
}

func (s *ClusterStatusRegistry) Unfollow(sessionID uuid.UUID, userIDs []uuid.UUID) {
	s.local.Unfollow(sessionID, userIDs)
}

func (s *ClusterStatusRegistry) UnfollowAll(sessionID uuid.UUID) {
	s.local.UnfollowAll(sessionID)
}

func (s *ClusterStatusRegistry) online(ids []uuid.UUID) []bool {
	refs := make([]cluster.StreamRef, len(ids))
	for i, id := range ids {
		refs[i] = streamRef(PresenceStream{Mode: StreamModeStatus, Subject: id})
	}
	ctx, cancel := context.WithTimeout(s.c.ctx, 3*time.Second)
	defer cancel()
	out, err := s.store.Exists(ctx, refs)
	if err != nil || len(out) != len(ids) {
		out = make([]bool, len(ids))
		for i, id := range ids {
			out[i] = s.local.IsOnline(id)
		}
		return out
	}
	for i, id := range ids {
		if !out[i] && s.local.IsOnline(id) {
			out[i] = true
		}
	}
	return out
}

func (s *ClusterStatusRegistry) IsOnline(userID uuid.UUID) bool {
	return s.online([]uuid.UUID{userID})[0]
}

func (s *ClusterStatusRegistry) FillOnlineUsers(users []*api.User) {
	if len(users) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(users))
	for i, u := range users {
		ids[i] = uuid.FromStringOrNil(u.Id)
	}
	for i, on := range s.online(ids) {
		users[i].Online = on
	}
}

func (s *ClusterStatusRegistry) FillOnlineAccounts(accounts []*api.Account) {
	if len(accounts) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(accounts))
	for i, a := range accounts {
		ids[i] = uuid.FromStringOrNil(a.User.Id)
	}
	for i, on := range s.online(ids) {
		accounts[i].User.Online = on
	}
}

func (s *ClusterStatusRegistry) FillOnlineFriends(friends []*api.Friend) {
	if len(friends) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(friends))
	for i, f := range friends {
		ids[i] = uuid.FromStringOrNil(f.User.Id)
	}
	for i, on := range s.online(ids) {
		friends[i].User.Online = on
	}
}

func (s *ClusterStatusRegistry) FillOnlineGroupUsers(groupUsers []*api.GroupUserList_GroupUser) {
	if len(groupUsers) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(groupUsers))
	for i, g := range groupUsers {
		ids[i] = uuid.FromStringOrNil(g.User.Id)
	}
	for i, on := range s.online(ids) {
		groupUsers[i].User.Online = on
	}
}

func (s *ClusterStatusRegistry) Queue(userID uuid.UUID, joins, leaves []*rtapi.UserPresence) {
	s.local.Queue(userID, joins, leaves)
}
