package server

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"go.uber.org/zap"
)

type ClusterLeaderboardScheduler struct {
	c     *Cluster
	local *LocalLeaderboardScheduler
	cache LeaderboardCache
	db    *sql.DB
}

var _ LeaderboardScheduler = (*ClusterLeaderboardScheduler)(nil)

func newClusterLeaderboardScheduler(c *Cluster, local *LocalLeaderboardScheduler, cache LeaderboardCache, db *sql.DB) *ClusterLeaderboardScheduler {
	s := &ClusterLeaderboardScheduler{c: c, local: local, cache: cache, db: db}
	local.claim = s.claim
	return s
}

func schedulerLeaseName(kind, id string, ts int64) string {
	return "lbs:" + kind + ":" + id + ":" + strconv.FormatInt(ts, 10)
}

func (s *ClusterLeaderboardScheduler) claim(kind, id string, ts int64) bool {
	ctx, cancel := context.WithTimeout(s.c.ctx, 5*time.Second)
	defer cancel()
	ok, err := clusterkit.AcquireLease(ctx, s.db, schedulerLeaseName(kind, id, ts), s.c.holder(), 6*time.Hour)
	if err != nil {
		s.c.logger.Error("Leaderboard scheduler claim failed, skipping on this node", zap.String("kind", kind), zap.String("id", id), zap.Error(err))
		return false
	}
	if !ok {
		s.c.logger.Debug("Leaderboard scheduler callback already run by another node", zap.String("kind", kind), zap.String("id", id), zap.Int64("ts", ts))
	}
	return ok
}

func (s *ClusterLeaderboardScheduler) sweepLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-s.c.ctx.Done():
			return
		case <-ticker.C:
		}
		ctx, cancel := context.WithTimeout(s.c.ctx, 30*time.Second)
		if ok, err := clusterkit.AcquireLease(ctx, s.db, "lbs-sweep", s.c.holder(), 50*time.Minute); err == nil && ok {
			if _, err := s.db.ExecContext(ctx, `DELETE FROM cluster_lease WHERE name LIKE 'lbs:%' AND expires_at < now()`); err != nil {
				s.c.logger.Warn("Leaderboard scheduler claim sweep failed", zap.Error(err))
			}
		}
		cancel()
	}
}

func (s *ClusterLeaderboardScheduler) Start(runtime *Runtime) {
	s.local.Start(runtime)
	go s.sweepLoop()
}

func (s *ClusterLeaderboardScheduler) Pause()  { s.local.Pause() }
func (s *ClusterLeaderboardScheduler) Resume() { s.local.Resume() }
func (s *ClusterLeaderboardScheduler) Stop()   { s.local.Stop() }
func (s *ClusterLeaderboardScheduler) Update() { s.local.Update() }

func (s *ClusterLeaderboardScheduler) handleChange(payload []byte) {
	ctx, cancel := context.WithTimeout(s.c.ctx, 30*time.Second)
	defer cancel()
	if err := s.cache.RefreshAllLeaderboards(ctx); err != nil {
		s.c.logger.Error("Leaderboard cache refresh after a change on another node failed", zap.Error(err))
		return
	}
	s.local.Update()
}

type ClusterLeaderboardCache struct {
	LeaderboardCache
	c *Cluster
}

func (l *ClusterLeaderboardCache) changed(id string) {
	l.c.broadcast(frameLeaderboard, []byte(id))
}

func (l *ClusterLeaderboardCache) Create(ctx context.Context, id string, authoritative bool, sortOrder, operator int, resetSchedule, metadata string, enableRanks bool) (*Leaderboard, bool, error) {
	lb, created, err := l.LeaderboardCache.Create(ctx, id, authoritative, sortOrder, operator, resetSchedule, metadata, enableRanks)
	if err == nil && created {
		l.changed(id)
	}
	return lb, created, err
}

func (l *ClusterLeaderboardCache) CreateTournament(ctx context.Context, id string, authoritative bool, sortOrder, operator int, resetSchedule, metadata, title, description string, category, startTime, endTime, duration, maxSize, maxNumScore int, joinRequired, enableRanks bool) (*Leaderboard, bool, error) {
	lb, created, err := l.LeaderboardCache.CreateTournament(ctx, id, authoritative, sortOrder, operator, resetSchedule, metadata, title, description, category, startTime, endTime, duration, maxSize, maxNumScore, joinRequired, enableRanks)
	if err == nil && created {
		l.changed(id)
	}
	return lb, created, err
}

func (l *ClusterLeaderboardCache) Delete(ctx context.Context, rankCache LeaderboardRankCache, scheduler LeaderboardScheduler, id string) (bool, error) {
	ok, err := l.LeaderboardCache.Delete(ctx, rankCache, scheduler, id)
	if err == nil && ok {
		l.changed(id)
	}
	return ok, err
}
