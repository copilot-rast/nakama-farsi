package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"github.com/redis/go-redis/v9"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	registerCallErrors(
		runtime.ErrMatchmakerQueryInvalid,
		runtime.ErrMatchmakerDuplicateSession,
		runtime.ErrMatchmakerIndex,
		runtime.ErrMatchmakerDelete,
		runtime.ErrMatchmakerNotAvailable,
		runtime.ErrMatchmakerTooManyTickets,
		runtime.ErrMatchmakerTicketNotFound,
	)
}

type ClusterMatchmaker struct {
	c       *Cluster
	logger  *zap.Logger
	config  Config
	router  MessageRouter
	metrics Metrics
	runtime *Runtime
	shards  *cluster.Shards
	prop    string

	mu         sync.Mutex
	owned      map[int]*ownedShard
	owners     []string
	freeRounds []int
	firstSeen  map[string]time.Time
	lastSweep  time.Time
	origin     map[string]int
	bySession  map[string]map[string]struct{}
	byParty    map[string]map[string]struct{}

	matchedFn func([][]*MatchmakerEntry)
	statsFn   func(*api.MatchmakerStats)
	active    *atomic.Bool
	stopped   *atomic.Bool
	stats     *atomic.Pointer[api.MatchmakerStats]
	done      chan struct{}
}

type ownedShard struct {
	idx   int
	token string
	mm    *LocalMatchmaker
}

var _ Matchmaker = (*ClusterMatchmaker)(nil)

func NewClusterAwareMatchmaker(c *Cluster, logger, startupLogger *zap.Logger, config Config, router MessageRouter, metrics Metrics, runtime *Runtime) Matchmaker {
	if c == nil {
		return NewLocalMatchmaker(logger, startupLogger, config, router, metrics, runtime)
	}
	cc := config.GetCluster()
	m := &ClusterMatchmaker{
		c:         c,
		logger:    logger,
		config:    config,
		router:    router,
		metrics:   metrics,
		runtime:   runtime,
		shards:    cluster.NewShards(c.rdb, c.keys, cc.MatchmakerShards, time.Duration(cc.NodeTTLMs+cc.HeartbeatMs)*time.Millisecond),
		prop:      cc.MatchmakerShardProperty,
		owned:     make(map[int]*ownedShard),
		firstSeen: make(map[string]time.Time),
		origin:    make(map[string]int),
		bySession: make(map[string]map[string]struct{}),
		byParty:   make(map[string]map[string]struct{}),
		active:    atomic.NewBool(true),
		stopped:   atomic.NewBool(false),
		stats:     atomic.NewPointer[api.MatchmakerStats](&api.MatchmakerStats{}),
		done:      make(chan struct{}),
	}
	m.freeRounds = make([]int, m.shards.Count())
	c.matchmaker = m
	if c.started {
		m.start()
	}
	return m
}

func (m *ClusterMatchmaker) start() {
	go m.loop()
}

func (m *ClusterMatchmaker) loop() {
	defer close(m.done)
	every := time.Duration(m.config.GetCluster().HeartbeatMs) * time.Millisecond
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	m.tick()
	for {
		select {
		case <-m.c.ctx.Done():
			return
		case <-ticker.C:
		}
		if m.stopped.Load() {
			return
		}
		m.tick()
	}
}

func (m *ClusterMatchmaker) tick() {
	ctx, cancel := context.WithTimeout(m.c.ctx, 3*time.Second)
	defer cancel()
	alive := m.c.members.AliveNames()
	now := time.Now()
	m.mu.Lock()
	for _, n := range alive {
		if _, ok := m.firstSeen[n]; !ok {
			m.firstSeen[n] = now
		}
	}
	for n := range m.firstSeen {
		if !m.c.members.IsAlive(n) {
			delete(m.firstSeen, n)
		}
	}
	owned := make([]*ownedShard, 0, len(m.owned))
	for _, o := range m.owned {
		owned = append(owned, o)
	}
	m.mu.Unlock()

	for _, o := range owned {
		res, err := m.shards.Renew(ctx, o.idx, o.token)
		if err != nil {
			continue
		}
		switch res {
		case cluster.RenewLost:
			m.logger.Warn("Matchmaker shard lease lost", zap.Int("shard", o.idx))
			m.drop(o, false)
		case cluster.RenewReclaimed:
			m.persistAll(ctx, o)
		}
	}

	owners, err := m.shards.Owners(ctx)
	if err != nil {
		return
	}
	m.mu.Lock()
	m.owners = owners
	m.mu.Unlock()
	for i, owner := range owners {
		preferred := cluster.Preferred(i, alive)
		m.mu.Lock()
		o := m.owned[i]
		if owner == "" {
			m.freeRounds[i]++
		} else {
			m.freeRounds[i] = 0
		}
		free := m.freeRounds[i]
		since, seen := m.firstSeen[preferred]
		m.mu.Unlock()
		switch {
		case owner == "" && o == nil && (preferred == m.c.name || free >= 3) && !m.stopped.Load():
			m.acquire(ctx, i)
		case o != nil && owner == m.c.name && preferred != m.c.name && seen && now.Sub(since) >= 10*time.Second:
			m.logger.Info("Matchmaker shard handed to its preferred owner", zap.Int("shard", i), zap.String("to", preferred))
			m.drop(o, true)
		}
	}
	m.refreshStats()
	if now.Sub(m.lastSweep) >= 30*time.Second {
		m.lastSweep = now
		go m.sweepOrphans()
	}
}

func (m *ClusterMatchmaker) sweepOrphans() {
	m.mu.Lock()
	owned := make([]*ownedShard, 0, len(m.owned))
	for _, o := range m.owned {
		owned = append(owned, o)
	}
	m.mu.Unlock()
	cutoff := time.Now().Add(-20 * time.Second).UnixNano()
	for _, o := range owned {
		type check struct {
			ticket string
			ref    cluster.StreamRef
			field  string
		}
		var checks []check
		o.mm.Lock()
		for ticket, index := range o.mm.indexes {
			if index.PartyId != "" || index.CreatedAt > cutoff {
				continue
			}
			for _, e := range index.Entries {
				uid := uuid.FromStringOrNil(e.Presence.UserId)
				sid := uuid.FromStringOrNil(e.Presence.SessionId)
				checks = append(checks, check{ticket: ticket, ref: streamRef(PresenceStream{Mode: StreamModeNotifications, Subject: uid}), field: presenceField(e.Presence.Node, sid, uid)})
			}
		}
		o.mm.Unlock()
		if len(checks) == 0 {
			continue
		}
		ctx, cancel := context.WithTimeout(m.c.ctx, 10*time.Second)
		pipe := m.c.rdb.Pipeline()
		cmds := make([]*redis.BoolCmd, len(checks))
		store := cluster.NewPresenceStore(m.c.rdb, m.c.keys)
		for i, ch := range checks {
			cmds[i] = pipe.HExists(ctx, store.P(ch.ref), ch.field)
		}
		_, err := pipe.Exec(ctx)
		cancel()
		if err != nil && err != redis.Nil {
			continue
		}
		gone := make(map[string]struct{})
		for i, cmd := range cmds {
			if ok, err := cmd.Result(); err == nil && !ok {
				gone[checks[i].ticket] = struct{}{}
			}
		}
		if len(gone) == 0 {
			continue
		}
		list := make([]string, 0, len(gone))
		for t := range gone {
			list = append(list, t)
		}
		o.mm.Remove(list)
		dctx, dcancel := context.WithTimeout(m.c.ctx, 5*time.Second)
		_ = m.shards.DeleteTickets(dctx, o.idx, list...)
		dcancel()
		m.logger.Warn("Matchmaker tickets of closed sessions removed", zap.Int("shard", o.idx), zap.Int("tickets", len(list)))
	}
}

func (m *ClusterMatchmaker) acquire(ctx context.Context, i int) {
	token := m.c.members.Self().Field() + "/" + uuid.Must(uuid.NewV4()).String()
	ok, err := m.shards.Acquire(ctx, i, token)
	if err != nil || !ok {
		return
	}
	inst, ok := NewLocalMatchmaker(m.logger, m.logger, m.config, m.router, m.metrics, m.runtime).(*LocalMatchmaker)
	if !ok {
		_ = m.shards.Release(ctx, i, token)
		return
	}
	o := &ownedShard{idx: i, token: token, mm: inst}
	inst.commitHook = func(entries [][]*MatchmakerEntry) [][]*MatchmakerEntry {
		return m.commit(o, entries)
	}
	if !m.active.Load() {
		inst.Pause()
	}
	m.mu.Lock()
	if m.matchedFn != nil {
		inst.OnMatchedEntries(m.matchedFn)
	}
	m.owned[i] = o
	m.owners[i] = m.c.name
	m.mu.Unlock()
	loaded := m.load(ctx, o)
	m.logger.Info("Matchmaker shard acquired", zap.Int("shard", i), zap.Int("tickets", loaded))
}

func (m *ClusterMatchmaker) load(ctx context.Context, o *ownedShard) int {
	raw, err := m.shards.Tickets(ctx, o.idx)
	if err != nil {
		return 0
	}
	extracts := make([]*MatchmakerExtract, 0, len(raw))
	var stale []string
	for ticket, value := range raw {
		ex := &MatchmakerExtract{}
		if err := json.Unmarshal([]byte(value), ex); err != nil {
			stale = append(stale, ticket)
			continue
		}
		if !m.extractAlive(ex) {
			stale = append(stale, ticket)
			continue
		}
		for _, p := range ex.Presences {
			p.SessionID = uuid.FromStringOrNil(p.SessionId)
		}
		ex.Node = m.c.name
		extracts = append(extracts, ex)
	}
	_ = m.shards.DeleteTickets(ctx, o.idx, stale...)
	if err := o.mm.Insert(extracts); err != nil {
		m.logger.Error("Matchmaker shard tickets could not be loaded", zap.Int("shard", o.idx), zap.Error(err))
	}
	return len(extracts)
}

func (m *ClusterMatchmaker) extractAlive(ex *MatchmakerExtract) bool {
	for _, p := range ex.Presences {
		if p.Node != "" && !m.c.members.IsAlive(p.Node) {
			return false
		}
	}
	if host := partyHost(ex.PartyId); host != "" && !m.c.members.IsAlive(host) {
		return false
	}
	return true
}

func partyHost(partyID string) string {
	if i := strings.IndexByte(partyID, '.'); i >= 0 {
		return partyID[i+1:]
	}
	return ""
}

func (m *ClusterMatchmaker) persistAll(ctx context.Context, o *ownedShard) {
	values := make(map[string]string)
	for _, ex := range o.mm.Extract() {
		if b, err := json.Marshal(ex); err == nil {
			values[ex.Ticket] = string(b)
		}
	}
	if err := m.shards.PutTickets(ctx, o.idx, values); err != nil {
		m.logger.Error("Matchmaker shard tickets could not be republished", zap.Int("shard", o.idx), zap.Error(err))
	}
}

func (m *ClusterMatchmaker) drop(o *ownedShard, release bool) {
	m.mu.Lock()
	if m.owned[o.idx] != o {
		m.mu.Unlock()
		return
	}
	delete(m.owned, o.idx)
	if m.owners != nil {
		m.owners[o.idx] = ""
	}
	m.mu.Unlock()
	o.mm.Stop()
	if release {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = m.shards.Release(ctx, o.idx, o.token)
		cancel()
	}
}

func (m *ClusterMatchmaker) commit(o *ownedShard, matched [][]*MatchmakerEntry) [][]*MatchmakerEntry {
	ctx, cancel := context.WithTimeout(m.c.ctx, 5*time.Second)
	defer cancel()
	out := matched[:0]
	for i, entries := range matched {
		seen := make(map[string]struct{}, len(entries))
		tickets := make([]string, 0, len(entries))
		for _, e := range entries {
			if _, ok := seen[e.Ticket]; !ok {
				seen[e.Ticket] = struct{}{}
				tickets = append(tickets, e.Ticket)
			}
		}
		res, missing, err := m.shards.Commit(ctx, o.idx, o.token, tickets)
		if err != nil || res == cluster.CommitLost {
			m.logger.Warn("Matchmaker shard lost while committing, matches dropped for the next owner", zap.Int("shard", o.idx), zap.Int("matches", len(matched)-i), zap.Error(err))
			go m.drop(o, false)
			return out
		}
		if res == cluster.CommitMissing {
			gone := make(map[string]struct{}, len(missing))
			for _, t := range missing {
				gone[t] = struct{}{}
			}
			keep := make([]string, 0, len(tickets))
			for _, t := range tickets {
				if _, ok := gone[t]; !ok {
					keep = append(keep, t)
				}
			}
			m.reinsert(ctx, o, keep)
			continue
		}
		out = append(out, entries)
		m.forget(tickets)
	}
	return out
}

func (m *ClusterMatchmaker) reinsert(ctx context.Context, o *ownedShard, tickets []string) {
	raw, err := m.shards.Tickets(ctx, o.idx, tickets...)
	if err != nil || len(raw) == 0 {
		return
	}
	extracts := make([]*MatchmakerExtract, 0, len(raw))
	for _, value := range raw {
		ex := &MatchmakerExtract{}
		if json.Unmarshal([]byte(value), ex) == nil {
			for _, p := range ex.Presences {
				p.SessionID = uuid.FromStringOrNil(p.SessionId)
			}
			ex.Node = m.c.name
			extracts = append(extracts, ex)
		}
	}
	_ = o.mm.Insert(extracts)
}

func (m *ClusterMatchmaker) forget(tickets []string) {
	m.mu.Lock()
	for _, t := range tickets {
		delete(m.origin, t)
	}
	m.mu.Unlock()
}

func (m *ClusterMatchmaker) shardFor(stringProperties map[string]string) int {
	key := ""
	if m.prop != "" {
		key = stringProperties[m.prop]
	}
	return m.shards.ShardOf(key)
}

func (m *ClusterMatchmaker) lookupOwner(ctx context.Context, i int) string {
	cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	owners, err := m.shards.Owners(cctx)
	if err != nil {
		return m.ownerOf(i)
	}
	m.mu.Lock()
	for j := range owners {
		if _, mine := m.owned[j]; mine {
			owners[j] = m.c.name
		}
	}
	m.owners = owners
	m.mu.Unlock()
	return owners[i]
}

func (m *ClusterMatchmaker) ownerOf(i int) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.owned[i]; ok {
		return m.c.name
	}
	if m.owners == nil || i >= len(m.owners) {
		return ""
	}
	return m.owners[i]
}

type mmAddRequest struct {
	Shard             int                   `json:"shard"`
	Presences         []*MatchmakerPresence `json:"presences"`
	SessionID         string                `json:"session_id"`
	PartyID           string                `json:"party_id"`
	Query             string                `json:"query"`
	MinCount          int                   `json:"min"`
	MaxCount          int                   `json:"max"`
	CountMultiple     int                   `json:"multiple"`
	StringProperties  map[string]string     `json:"sp"`
	NumericProperties map[string]float64    `json:"np"`
}

type mmAddResponse struct {
	Ticket    string `json:"ticket"`
	CreatedAt int64  `json:"created_at"`
}

type mmRemoveRequest struct {
	Shard   int      `json:"shard"`
	Tickets []string `json:"tickets"`
}

func (m *ClusterMatchmaker) Add(ctx context.Context, presences []*MatchmakerPresence, sessionID, partyId, query string, minCount, maxCount, countMultiple int, stringProperties map[string]string, numericProperties map[string]float64) (string, int64, error) {
	if m.stopped.Load() {
		return "", 0, runtime.ErrMatchmakerNotAvailable
	}
	shard := m.shardFor(stringProperties)
	req := &mmAddRequest{Shard: shard, Presences: presences, SessionID: sessionID, PartyID: partyId, Query: query, MinCount: minCount, MaxCount: maxCount, CountMultiple: countMultiple, StringProperties: stringProperties, NumericProperties: numericProperties}
	payload, _ := json.Marshal(req)
	var resp *mmAddResponse
	err := runtime.ErrMatchmakerNotAvailable
	callCtx, cancelCall := context.WithTimeout(context.WithoutCancel(ctx), m.shards.TTL()+5*time.Second)
	defer cancelCall()
	ctx = callCtx
	deadline := time.Now().Add(m.shards.TTL() + 2*time.Second)
	for attempt := 0; ; attempt++ {
		owner := m.ownerOf(shard)
		if owner == "" || attempt > 0 {
			owner = m.lookupOwner(ctx, shard)
		}
		if owner == m.c.name {
			resp, err = m.addLocal(ctx, req)
		} else if owner != "" {
			var out []byte
			out, err = m.c.call(ctx, owner, callMatchmakerAdd, payload)
			if err == nil {
				resp = &mmAddResponse{}
				err = json.Unmarshal(out, resp)
			} else if !isKnownMatchmakerError(err) {
				err = runtime.ErrMatchmakerNotAvailable
			}
		}
		if err == nil || !errors.Is(err, runtime.ErrMatchmakerNotAvailable) || time.Now().After(deadline) || ctx.Err() != nil || m.stopped.Load() {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		return "", 0, err
	}
	m.mu.Lock()
	m.origin[resp.Ticket] = shard
	if sessionID != "" {
		addTicketRef(m.bySession, sessionID, resp.Ticket)
	}
	if partyId != "" {
		addTicketRef(m.byParty, partyId, resp.Ticket)
	}
	m.mu.Unlock()
	if sessionID != "" && m.c.sessionRegistry != nil {
		if sid, perr := uuid.FromString(sessionID); perr == nil && m.c.sessionRegistry.Get(sid) == nil {
			_ = m.RemoveSessionAll(sessionID)
			return "", 0, runtime.ErrMatchmakerNotAvailable
		}
	}
	return resp.Ticket, resp.CreatedAt, nil
}

func isKnownMatchmakerError(err error) bool {
	for _, known := range []error{runtime.ErrMatchmakerQueryInvalid, runtime.ErrMatchmakerDuplicateSession, runtime.ErrMatchmakerIndex, runtime.ErrMatchmakerTooManyTickets, runtime.ErrMatchmakerNotAvailable} {
		if errors.Is(err, known) {
			return true
		}
	}
	return false
}

func addTicketRef(index map[string]map[string]struct{}, key, ticket string) {
	set, ok := index[key]
	if !ok {
		set = make(map[string]struct{}, 1)
		index[key] = set
	}
	set[ticket] = struct{}{}
}

func (m *ClusterMatchmaker) addLocal(ctx context.Context, req *mmAddRequest) (*mmAddResponse, error) {
	m.mu.Lock()
	o := m.owned[req.Shard]
	m.mu.Unlock()
	if o == nil {
		return nil, runtime.ErrMatchmakerNotAvailable
	}
	for _, p := range req.Presences {
		if p.SessionID == uuid.Nil {
			p.SessionID = uuid.FromStringOrNil(p.SessionId)
		}
	}
	ticket, createdAt, err := o.mm.Add(ctx, req.Presences, req.SessionID, req.PartyID, req.Query, req.MinCount, req.MaxCount, req.CountMultiple, req.StringProperties, req.NumericProperties)
	if err != nil {
		return nil, err
	}
	if ticket == "" {
		return nil, runtime.ErrMatchmakerNotAvailable
	}
	ex := &MatchmakerExtract{Presences: req.Presences, SessionID: req.SessionID, PartyId: req.PartyID, Query: req.Query, MinCount: req.MinCount, MaxCount: req.MaxCount, CountMultiple: req.CountMultiple, StringProperties: req.StringProperties, NumericProperties: req.NumericProperties, Ticket: ticket, Count: len(req.Presences), CreatedAt: createdAt, Node: m.c.name}
	value, _ := json.Marshal(ex)
	pctx, cancel := context.WithTimeout(m.c.ctx, 3*time.Second)
	ok, perr := m.shards.PutTicket(pctx, req.Shard, o.token, ticket, string(value))
	cancel()
	if perr != nil || !ok {
		o.mm.Remove([]string{ticket})
		return nil, runtime.ErrMatchmakerNotAvailable
	}
	return &mmAddResponse{Ticket: ticket, CreatedAt: createdAt}, nil
}

func (m *ClusterMatchmaker) handleCall(ctx context.Context, kind byte, payload []byte) ([]byte, error) {
	switch kind {
	case callMatchmakerAdd:
		req := &mmAddRequest{}
		if err := json.Unmarshal(payload, req); err != nil {
			return nil, err
		}
		resp, err := m.addLocal(ctx, req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	case callMatchmakerRemove:
		req := &mmRemoveRequest{}
		if err := json.Unmarshal(payload, req); err != nil {
			return nil, err
		}
		m.removeLocal(req.Shard, req.Tickets)
		return nil, nil
	}
	return nil, errors.New("unknown matchmaker call")
}

func (m *ClusterMatchmaker) removeLocal(shard int, tickets []string) {
	m.mu.Lock()
	var targets []*ownedShard
	if shard >= 0 {
		if o := m.owned[shard]; o != nil {
			targets = append(targets, o)
		}
	} else {
		for _, o := range m.owned {
			targets = append(targets, o)
		}
	}
	m.mu.Unlock()
	for _, o := range targets {
		o.mm.Remove(tickets)
	}
}

func (m *ClusterMatchmaker) removeTickets(tickets []string) {
	if len(tickets) == 0 {
		return
	}
	byShard := make(map[int][]string)
	m.mu.Lock()
	for _, t := range tickets {
		shard, ok := m.origin[t]
		if !ok {
			shard = -1
		}
		byShard[shard] = append(byShard[shard], t)
		delete(m.origin, t)
	}
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(m.c.ctx, 3*time.Second)
	defer cancel()
	for shard, list := range byShard {
		if shard >= 0 {
			_ = m.shards.DeleteTickets(ctx, shard, list...)
			if owner := m.ownerOf(shard); owner == m.c.name {
				m.removeLocal(shard, list)
			} else if owner != "" {
				payload, _ := json.Marshal(&mmRemoveRequest{Shard: shard, Tickets: list})
				_, _ = m.c.call(ctx, owner, callMatchmakerRemove, payload)
			}
			continue
		}
		_ = m.shards.DeleteTicketsEverywhere(ctx, list...)
		payload, _ := json.Marshal(&mmRemoveRequest{Shard: -1, Tickets: list})
		for _, r := range m.c.fanout(ctx, callMatchmakerRemove, payload, true) {
			_ = r
		}
	}
}

func (m *ClusterMatchmaker) takeRefs(index map[string]map[string]struct{}, key, ticket string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	set, ok := index[key]
	if !ok {
		return nil
	}
	if ticket != "" {
		if _, has := set[ticket]; !has {
			return nil
		}
		delete(set, ticket)
		if len(set) == 0 {
			delete(index, key)
		}
		return []string{ticket}
	}
	delete(index, key)
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	return out
}

func (m *ClusterMatchmaker) RemoveSession(sessionID, ticket string) error {
	tickets := m.takeRefs(m.bySession, sessionID, ticket)
	if len(tickets) == 0 {
		return runtime.ErrMatchmakerTicketNotFound
	}
	m.removeTickets(tickets)
	return nil
}

func (m *ClusterMatchmaker) RemoveSessionAll(sessionID string) error {
	m.removeTickets(m.takeRefs(m.bySession, sessionID, ""))
	return nil
}

func (m *ClusterMatchmaker) RemoveParty(partyID, ticket string) error {
	tickets := m.takeRefs(m.byParty, partyID, ticket)
	if len(tickets) == 0 {
		return runtime.ErrMatchmakerTicketNotFound
	}
	m.removeTickets(tickets)
	return nil
}

func (m *ClusterMatchmaker) RemovePartyAll(partyID string) error {
	m.removeTickets(m.takeRefs(m.byParty, partyID, ""))
	return nil
}

func (m *ClusterMatchmaker) RemoveAll(node string) {
	m.nodeDown(node)
}

func (m *ClusterMatchmaker) Remove(tickets []string) {
	m.removeTickets(tickets)
}

func (m *ClusterMatchmaker) nodeDown(node string) {
	m.mu.Lock()
	owned := make([]*ownedShard, 0, len(m.owned))
	for _, o := range m.owned {
		owned = append(owned, o)
	}
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(m.c.ctx, 5*time.Second)
	defer cancel()
	for _, o := range owned {
		var gone []string
		o.mm.Lock()
		for ticket, index := range o.mm.indexes {
			if partyHost(index.PartyId) == node {
				gone = append(gone, ticket)
				continue
			}
			for _, e := range index.Entries {
				if e.Presence.Node == node {
					gone = append(gone, ticket)
					break
				}
			}
		}
		o.mm.Unlock()
		if len(gone) > 0 {
			o.mm.Remove(gone)
			_ = m.shards.DeleteTickets(ctx, o.idx, gone...)
			m.logger.Warn("Matchmaker tickets of a dead node removed", zap.String("node", node), zap.Int("shard", o.idx), zap.Int("tickets", len(gone)))
		}
	}
}

func (m *ClusterMatchmaker) republish(ctx context.Context) {
	m.mu.Lock()
	owned := make([]*ownedShard, 0, len(m.owned))
	for _, o := range m.owned {
		owned = append(owned, o)
	}
	m.mu.Unlock()
	for _, o := range owned {
		m.persistAll(ctx, o)
	}
}

func (m *ClusterMatchmaker) Insert(extracts []*MatchmakerExtract) error {
	byShard := make(map[int][]*MatchmakerExtract)
	for _, ex := range extracts {
		i := m.shardFor(ex.StringProperties)
		byShard[i] = append(byShard[i], ex)
	}
	for i, list := range byShard {
		m.mu.Lock()
		o := m.owned[i]
		m.mu.Unlock()
		if o == nil {
			continue
		}
		if err := o.mm.Insert(list); err != nil {
			return err
		}
		m.persistAll(m.c.ctx, o)
	}
	return nil
}

func (m *ClusterMatchmaker) Extract() []*MatchmakerExtract {
	m.mu.Lock()
	owned := make([]*ownedShard, 0, len(m.owned))
	for _, o := range m.owned {
		owned = append(owned, o)
	}
	m.mu.Unlock()
	var out []*MatchmakerExtract
	for _, o := range owned {
		out = append(out, o.mm.Extract()...)
	}
	return out
}

func (m *ClusterMatchmaker) Pause() {
	m.active.Store(false)
	m.mu.Lock()
	for _, o := range m.owned {
		o.mm.Pause()
	}
	m.mu.Unlock()
}

func (m *ClusterMatchmaker) Resume() {
	m.active.Store(true)
	m.mu.Lock()
	for _, o := range m.owned {
		o.mm.Resume()
	}
	m.mu.Unlock()
}

func (m *ClusterMatchmaker) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	m.stop(ctx)
}

func (m *ClusterMatchmaker) stop(ctx context.Context) {
	if !m.stopped.CompareAndSwap(false, true) {
		return
	}
	m.mu.Lock()
	owned := make([]*ownedShard, 0, len(m.owned))
	for _, o := range m.owned {
		owned = append(owned, o)
	}
	m.owned = make(map[int]*ownedShard)
	m.mu.Unlock()
	for _, o := range owned {
		o.mm.Stop()
		_ = m.shards.Release(ctx, o.idx, o.token)
	}
}

func (m *ClusterMatchmaker) OnMatchedEntries(fn func(entries [][]*MatchmakerEntry)) {
	m.mu.Lock()
	m.matchedFn = fn
	for _, o := range m.owned {
		o.mm.OnMatchedEntries(fn)
	}
	m.mu.Unlock()
}

func (m *ClusterMatchmaker) OnStatsUpdate(fn func(stats *api.MatchmakerStats)) {
	m.mu.Lock()
	m.statsFn = fn
	m.mu.Unlock()
}

func (m *ClusterMatchmaker) refreshStats() {
	m.mu.Lock()
	owned := make([]*ownedShard, 0, len(m.owned))
	for _, o := range m.owned {
		owned = append(owned, o)
	}
	fn := m.statsFn
	m.mu.Unlock()
	total := &api.MatchmakerStats{}
	for _, o := range owned {
		s := o.mm.GetStats()
		if s == nil {
			continue
		}
		total.TicketCount += s.TicketCount
		if s.OldestTicketCreateTime != nil && (total.OldestTicketCreateTime == nil || s.OldestTicketCreateTime.AsTime().Before(total.OldestTicketCreateTime.AsTime())) {
			total.OldestTicketCreateTime = timestamppb.New(s.OldestTicketCreateTime.AsTime())
		}
		total.Completions = append(total.Completions, s.Completions...)
	}
	m.stats.Store(total)
	if fn != nil {
		fn(total)
	}
}

func (m *ClusterMatchmaker) GetStats() *api.MatchmakerStats {
	return m.stats.Load()
}

func (m *ClusterMatchmaker) SetStats(stats *api.MatchmakerStats) {
	m.stats.Store(stats)
}

func (m *ClusterMatchmaker) ownedShards() []int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]int, 0, len(m.owned))
	for i := range m.owned {
		out = append(out, i)
	}
	return out
}
