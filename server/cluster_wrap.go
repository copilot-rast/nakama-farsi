package server

import (
	"database/sql"
	"time"
)

func (c *Cluster) SessionRegistry(local SessionRegistry) SessionRegistry {
	if c == nil {
		return local
	}
	c.sessionRegistry = local
	c.sessions = &ClusterSessionRegistry{c: c, local: local}
	return c.sessions
}

func (c *Cluster) SessionCache(local SessionCache, console bool) SessionCache {
	if c == nil {
		return local
	}
	keep := time.Duration(max(c.config.GetSession().TokenExpirySec, c.config.GetSession().RefreshTokenExpirySec)+60) * time.Second
	if console {
		keep = time.Duration(c.config.GetConsole().TokenExpirySec+60) * time.Second
	}
	sc := newClusterSessionCache(c, local, console, keep)
	c.sessionCaches = append(c.sessionCaches, sc)
	return sc
}

func (c *Cluster) MatchRegistry(local MatchRegistry) MatchRegistry {
	if c == nil {
		return local
	}
	if lm, ok := local.(*LocalMatchRegistry); ok {
		c.matchRegistry = &ClusterMatchRegistry{c: c, local: lm}
		return c.matchRegistry
	}
	return local
}

func (c *Cluster) LeaderboardCache(local LeaderboardCache) LeaderboardCache {
	if c == nil {
		return local
	}
	return &ClusterLeaderboardCache{LeaderboardCache: local, c: c}
}

func (c *Cluster) LeaderboardScheduler(local LeaderboardScheduler, cache LeaderboardCache, db *sql.DB) LeaderboardScheduler {
	if c == nil {
		return local
	}
	if ls, ok := local.(*LocalLeaderboardScheduler); ok {
		c.scheduler = newClusterLeaderboardScheduler(c, ls, cache, db)
		return c.scheduler
	}
	return local
}

func (c *Cluster) Tracker(local Tracker) Tracker {
	if c == nil {
		return local
	}
	c.localTracker = local
	if lt, ok := local.(*LocalTracker); ok {
		c.tracker = newClusterTracker(c, lt)
		return c.tracker
	}
	return local
}

func (c *Cluster) Router(local MessageRouter) MessageRouter {
	if c == nil {
		return local
	}
	c.localRouter = NewLocalMessageRouter(c.sessionRegistry, c.localTracker, c.marshaler)
	c.router = &ClusterRouter{c: c, local: c.localRouter}
	return c.router
}

func (c *Cluster) StatusHandler(local StatusHandler) StatusHandler {
	if c == nil {
		return local
	}
	c.statusHandler = &ClusterStatusHandler{c: c, local: local}
	return c.statusHandler
}

func (c *Cluster) StatusRegistry(local StatusRegistry) StatusRegistry {
	if c == nil {
		return local
	}
	if ls, ok := local.(*LocalStatusRegistry); ok {
		c.statusRegistry = newClusterStatusRegistry(c, ls)
		return c.statusRegistry
	}
	return local
}

func (c *Cluster) StreamManager(local StreamManager) StreamManager {
	if c == nil {
		return local
	}
	c.streamManager = &ClusterStreamManager{c: c, local: local}
	return c.streamManager
}

func (c *Cluster) PartyRegistry(local PartyRegistry) PartyRegistry {
	if c == nil {
		return local
	}
	if lp, ok := local.(*LocalPartyRegistry); ok {
		c.partyRegistry = &ClusterPartyRegistry{c: c, local: lp}
		return c.partyRegistry
	}
	return local
}
