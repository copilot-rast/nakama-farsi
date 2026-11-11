package server

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"go.uber.org/zap"
)

type ClusterSessionCache struct {
	c     *Cluster
	local SessionCache
	log   *cluster.InvalidationLog
}

var _ SessionCache = (*ClusterSessionCache)(nil)

func newClusterSessionCache(c *Cluster, local SessionCache, console bool, keep time.Duration) *ClusterSessionCache {
	return &ClusterSessionCache{c: c, local: local, log: cluster.NewInvalidationLog(c.logger, c.rdb, c.keys.Invalidations(console), keep, c.name)}
}

func (s *ClusterSessionCache) start() {
	ctx, cancel := context.WithTimeout(s.c.ctx, 10*time.Second)
	last := s.log.Replay(ctx, s.apply)
	cancel()
	go s.log.Tail(s.c.ctx, last, s.apply)
}

func (s *ClusterSessionCache) apply(inv cluster.Invalidation) {
	if inv.Origin == s.c.name {
		return
	}
	ids := make([]uuid.UUID, 0, len(inv.UserIDs))
	for _, id := range inv.UserIDs {
		if u, err := uuid.FromString(id); err == nil {
			ids = append(ids, u)
		}
	}
	switch inv.Op {
	case "remove":
		for _, id := range ids {
			s.local.Remove(id, inv.SessionExp, inv.SessionID, inv.RefreshExp, inv.RefreshID)
		}
	case "removeall":
		if lc, ok := s.local.(*LocalSessionCache); ok && !inv.At.IsZero() {
			lc.invalidateAt(ids, inv.At.Unix(), true)
		} else {
			for _, id := range ids {
				s.local.RemoveAll(id)
			}
		}
	case "ban":
		if lc, ok := s.local.(*LocalSessionCache); ok && !inv.At.IsZero() {
			lc.invalidateAt(ids, inv.At.Unix(), false)
		} else {
			s.local.Ban(ids)
		}
	case "unban":
		s.local.Unban(ids)
	}
}

func (s *ClusterSessionCache) publish(inv cluster.Invalidation) {
	ctx, cancel := context.WithTimeout(s.c.ctx, 3*time.Second)
	defer cancel()
	if err := s.log.Append(ctx, inv); err != nil {
		s.c.logger.Error("Cluster session invalidation could not be published", zap.String("op", inv.Op), zap.Error(err))
	}
}

func idStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func (s *ClusterSessionCache) Stop() {
	s.local.Stop()
}

func (s *ClusterSessionCache) IsValidSession(userID uuid.UUID, exp int64, tokenId string) bool {
	return s.local.IsValidSession(userID, exp, tokenId)
}

func (s *ClusterSessionCache) IsValidRefresh(userID uuid.UUID, exp int64, tokenId string) bool {
	return s.local.IsValidRefresh(userID, exp, tokenId)
}

func (s *ClusterSessionCache) Add(userID uuid.UUID, sessionExp int64, sessionTokenId string, refreshExp int64, refreshTokenId string) {
	s.local.Add(userID, sessionExp, sessionTokenId, refreshExp, refreshTokenId)
}

func (s *ClusterSessionCache) Remove(userID uuid.UUID, sessionExp int64, sessionTokenId string, refreshExp int64, refreshTokenId string) {
	s.local.Remove(userID, sessionExp, sessionTokenId, refreshExp, refreshTokenId)
	s.publish(cluster.Invalidation{Op: "remove", UserIDs: []string{userID.String()}, SessionExp: sessionExp, SessionID: sessionTokenId, RefreshExp: refreshExp, RefreshID: refreshTokenId})
}

func (s *ClusterSessionCache) RemoveAll(userID uuid.UUID) {
	s.local.RemoveAll(userID)
	s.publish(cluster.Invalidation{Op: "removeall", UserIDs: []string{userID.String()}})
}

func (s *ClusterSessionCache) Ban(userIDs []uuid.UUID) {
	s.local.Ban(userIDs)
	if len(userIDs) > 0 {
		s.publish(cluster.Invalidation{Op: "ban", UserIDs: idStrings(userIDs)})
	}
}

func (s *ClusterSessionCache) Unban(userIDs []uuid.UUID) {
	s.local.Unban(userIDs)
	if len(userIDs) > 0 {
		s.publish(cluster.Invalidation{Op: "unban", UserIDs: idStrings(userIDs)})
	}
}

type ClusterSessionRegistry struct {
	c     *Cluster
	local SessionRegistry
}

var _ SessionRegistry = (*ClusterSessionRegistry)(nil)

func (r *ClusterSessionRegistry) Stop()                               { r.local.Stop() }
func (r *ClusterSessionRegistry) Count() int                          { return r.local.Count() }
func (r *ClusterSessionRegistry) Get(sessionID uuid.UUID) Session     { return r.local.Get(sessionID) }
func (r *ClusterSessionRegistry) Add(session Session)                 { r.local.Add(session) }
func (r *ClusterSessionRegistry) Remove(sessionID uuid.UUID)          { r.local.Remove(sessionID) }
func (r *ClusterSessionRegistry) Range(fn func(session Session) bool) { r.local.Range(fn) }

func (r *ClusterSessionRegistry) Disconnect(ctx context.Context, sessionID uuid.UUID, ban bool, reason ...runtime.PresenceReason) error {
	if r.local.Get(sessionID) != nil {
		return r.local.Disconnect(ctx, sessionID, ban, reason...)
	}
	rs := runtime.PresenceReasonDisconnect
	if len(reason) > 0 {
		rs = reason[0]
	}
	e := cluster.NewEnc(24).Raw16(sessionID).Bool(ban).Uvarint(uint64(rs))
	if node, ok := r.c.NodeForSession(sessionID); ok && node != r.c.name {
		r.c.send(node, frameDisconnect, e.Out())
		return nil
	}
	r.c.broadcast(frameDisconnect, e.Out())
	return nil
}

func (r *ClusterSessionRegistry) SingleSession(ctx context.Context, tracker Tracker, userID, sessionID uuid.UUID) {
	r.local.SingleSession(ctx, tracker, userID, sessionID)
	if r.c.tracker == nil {
		return
	}
	payload := cluster.NewEnc(32).Raw16(userID).Raw16(sessionID).Out()
	for _, node := range r.c.tracker.remoteNodes(PresenceStream{Mode: StreamModeNotifications, Subject: userID}) {
		r.c.send(node, frameSingleSession, payload)
	}
}

func (c *Cluster) handleSessionFrame(kind byte, payload []byte) {
	if c.sessions == nil {
		return
	}
	d := cluster.NewDec(payload)
	switch kind {
	case frameDisconnect:
		sid := uuid.UUID(d.Raw16())
		ban := d.Bool()
		reason := runtime.PresenceReason(d.Uvarint())
		if d.Err() == nil && c.sessions.local.Get(sid) != nil {
			_ = c.sessions.local.Disconnect(c.ctx, sid, ban, reason)
		}
	case frameSingleSession:
		uid := uuid.UUID(d.Raw16())
		keep := uuid.UUID(d.Raw16())
		if d.Err() == nil && c.localTracker != nil {
			c.sessions.local.SingleSession(c.ctx, c.localTracker, uid, keep)
		}
	}
}

func (s *LocalSessionCache) invalidateAt(userIDs []uuid.UUID, ts int64, clearTokens bool) {
	s.Lock()
	for _, userID := range userIDs {
		cache, found := s.cache[userID]
		if !found {
			cache = &sessionCacheUser{sessionTokens: make(map[string]int64), refreshTokens: make(map[string]int64)}
			s.cache[userID] = cache
		}
		if ts > cache.lastInvalidation {
			cache.lastInvalidation = ts
			if clearTokens {
				cache.sessionTokens = make(map[string]int64)
				cache.refreshTokens = make(map[string]int64)
			}
		}
	}
	s.Unlock()
}
