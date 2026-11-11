package server

func (c *Cluster) SessionRegistry(local SessionRegistry) SessionRegistry {
	if c == nil {
		return local
	}
	c.sessionRegistry = local
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
