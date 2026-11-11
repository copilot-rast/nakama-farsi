package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	frameEnvelopeSessions byte = iota + 1
	frameEnvelopeStream
	frameEnvelopeAll
	framePresence
	frameUntrack
	frameUntrackStream
	frameDisconnect
	frameSingleSession
	frameMatchData
	frameLeaderboard
	frameHostLost
)

const (
	callNodeStatus byte = iota + 1
	callStreamCounts
	callStreamUser
	callMatchGet
	callMatchJoinAttempt
	callMatchSignal
	callMatchState
	callMatchList
	callParty
	callPartyList
	callMatchmakerAdd
	callMatchmakerRemove
)

type Cluster struct {
	logger      *zap.Logger
	config      Config
	cc          *ClusterConfig
	name        string
	hash        [6]byte
	marshaler   *protojson.MarshalOptions
	unmarshaler *protojson.UnmarshalOptions

	rdb       redis.UniversalClient
	keys      cluster.Keys
	members   *cluster.Membership
	transport *cluster.Transport

	sessionRegistry SessionRegistry
	localTracker    Tracker
	tracker         *ClusterTracker
	statusRegistry  *ClusterStatusRegistry
	partyRegistry   *ClusterPartyRegistry
	matchRegistry   *ClusterMatchRegistry
	sessions        *ClusterSessionRegistry
	sessionCaches   []*ClusterSessionCache
	scheduler       *ClusterLeaderboardScheduler
	matchmaker      *ClusterMatchmaker
	streamManager   *ClusterStreamManager
	localRouter     MessageRouter
	router          *ClusterRouter
	statusHandler   *ClusterStatusHandler

	reapMu  sync.Mutex
	reaping map[string]struct{}

	ctx      context.Context
	cancel   context.CancelFunc
	started  bool
	stopOnce sync.Once
}

func NewCluster(logger, startupLogger *zap.Logger, config Config, marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) *Cluster {
	cc := config.GetCluster()
	if cc == nil || !cc.Enabled {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := &Cluster{
		logger:      logger.With(zap.String("subsystem", "cluster")),
		config:      config,
		cc:          cc,
		name:        config.GetName(),
		hash:        NodeToHash(config.GetName()),
		marshaler:   marshaler,
		unmarshaler: unmarshaler,
		keys:        cluster.Keys{Prefix: cc.Prefix},
		reaping:     make(map[string]struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}
	c.rdb = cluster.NewRedis(cluster.RedisOptions{
		Addresses:  cc.RedisAddress,
		Username:   cc.RedisUsername,
		Password:   cc.RedisPassword,
		DB:         cc.RedisDB,
		MasterName: cc.RedisMasterName,
		TLS:        cc.RedisTLS,
	})
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := c.rdb.Ping(pingCtx).Err(); err != nil {
		startupLogger.Fatal("Cluster Redis is not reachable", zap.String("redis_address", cc.RedisAddress), zap.Error(err))
	}
	secret := cc.Secret
	if secret == "" {
		secret = config.GetSession().EncryptionKey
	}
	c.transport = cluster.NewTransport(c.logger, c.name, secret, cc.SendQueueSize, time.Duration(cc.CallTimeoutMs)*time.Millisecond, c)
	if err := c.transport.Listen(cc.Address, cc.Port); err != nil {
		startupLogger.Fatal("Cluster port could not be opened", zap.Int("cluster.port", cc.Port), zap.Error(err))
	}
	advertise := cc.AdvertiseAddress
	if advertise == "" {
		advertise = firstIPv4()
	}
	self := cluster.NodeInfo{Name: c.name, Inc: cluster.NewIncarnation(), Addr: net.JoinHostPort(advertise, strconv.Itoa(c.transport.Port()))}
	c.members = cluster.NewMembership(c.logger, c.rdb, c.keys, self, time.Duration(cc.HeartbeatMs)*time.Millisecond, time.Duration(cc.NodeTTLMs)*time.Millisecond, cluster.MembershipEvents{
		Up:           c.onNodeUp,
		Down:         c.onNodeDown,
		Dead:         c.onNodeDead,
		Epoch:        c.onEpoch,
		Reincarnated: c.onReincarnated,
		RedisUp:      c.onRedisUp,
	})
	startupLogger.Info("Cluster enabled", zap.String("node", c.name), zap.String("advertise", self.Addr), zap.String("redis", cc.RedisAddress), zap.String("prefix", cc.Prefix))
	return c
}

func firstIPv4() string {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func (c *Cluster) Enabled() bool {
	return c != nil
}

func (c *Cluster) Name() string {
	return c.name
}

func (c *Cluster) Redis() redis.UniversalClient {
	return c.rdb
}

func (c *Cluster) Self() cluster.NodeInfo {
	return c.members.Self()
}

func (c *Cluster) NodeForSession(sessionID uuid.UUID) (string, bool) {
	h := HashFromId(sessionID)
	if h == c.hash {
		return c.name, true
	}
	return c.members.NodeByHash(h)
}

func (c *Cluster) Start() {
	if c == nil {
		return
	}
	if err := c.members.Start(c.ctx); err != nil {
		c.logger.Fatal("Cluster registration failed", zap.Error(err))
	}
	if c.tracker != nil {
		c.tracker.start()
	}
	if c.statusRegistry != nil {
		c.statusRegistry.start()
	}
	if c.matchmaker != nil {
		c.matchmaker.start()
	}
	for _, sc := range c.sessionCaches {
		sc.start()
	}
	cluster.SetSignals(cluster.NewRedisSignals(c.ctx, c.logger, c.rdb, c.keys))
	c.started = true
	c.logger.Info("Cluster started", zap.Strings("nodes", c.members.AliveNames()))
}

func (c *Cluster) Stop() {
	if c == nil {
		return
	}
	c.stopOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if c.matchmaker != nil {
			c.matchmaker.stop(ctx)
		}
		if c.tracker != nil {
			c.tracker.leaveAll(ctx)
		}
		if c.statusRegistry != nil {
			c.statusRegistry.leaveAll(ctx)
		}
		if c.started {
			c.members.Stop(ctx)
		}
		c.cancel()
		c.transport.Stop()
		_ = c.rdb.Close()
	})
}

func (c *Cluster) onNodeUp(info cluster.NodeInfo) {
	if err := c.transport.AddPeer(info.Name, info.Addr); err != nil {
		c.logger.Error("Cluster peer could not be added", zap.String("peer", info.Name), zap.Error(err))
		return
	}
	c.logger.Info("Cluster node up", zap.String("peer", info.Name), zap.String("addr", info.Addr))
}

func (c *Cluster) onNodeDown(info cluster.NodeInfo) {
	c.transport.RemovePeer(info.Name)
	c.logger.Warn("Cluster node down", zap.String("peer", info.Name), zap.String("inc", info.Inc))
	go c.hostLost(info.Name)
	if c.matchmaker != nil {
		go c.matchmaker.nodeDown(info.Name)
	}
}

func (c *Cluster) onEpoch() {
	c.logger.Warn("Cluster Redis state was reset, republishing this node's state")
	go c.republish()
}

func (c *Cluster) onReincarnated(old, current string) {
	c.logger.Warn("Cluster node registered under a new incarnation, republishing this node's state", zap.String("old", old), zap.String("inc", current))
	go c.republish()
}

func (c *Cluster) onRedisUp() {
	go c.republish()
}

func (c *Cluster) republish() {
	if c.tracker != nil {
		c.tracker.reconcile(c.ctx)
	}
	if c.statusRegistry != nil {
		c.statusRegistry.republish(c.ctx)
	}
	if c.matchmaker != nil {
		c.matchmaker.republish(c.ctx)
	}
}

func (c *Cluster) peers() []string {
	return c.transport.Peers()
}

func (c *Cluster) send(node string, kind byte, payload []byte) bool {
	if node == c.name {
		c.HandleFrame(c.name, kind, payload)
		return true
	}
	return c.transport.Send(node, kind, payload)
}

func (c *Cluster) broadcast(kind byte, payload []byte) {
	c.transport.Broadcast(kind, payload)
}

func (c *Cluster) call(ctx context.Context, node string, kind byte, payload []byte) ([]byte, error) {
	if node == c.name {
		return c.HandleCall(ctx, c.name, kind, payload)
	}
	out, err := c.transport.Call(ctx, node, kind, payload)
	if err != nil {
		return nil, mapCallError(err)
	}
	return out, nil
}

type fanoutResult struct {
	node string
	out  []byte
	err  error
}

func (c *Cluster) fanout(ctx context.Context, kind byte, payload []byte, includeSelf bool) []fanoutResult {
	nodes := c.members.AliveNames()
	results := make([]fanoutResult, len(nodes))
	wg := &sync.WaitGroup{}
	for i, node := range nodes {
		if node == c.name && !includeSelf {
			results[i] = fanoutResult{node: node, err: errSkipSelf}
			continue
		}
		wg.Add(1)
		go func(i int, node string) {
			defer wg.Done()
			out, err := c.call(ctx, node, kind, payload)
			results[i] = fanoutResult{node: node, out: out, err: err}
		}(i, node)
	}
	wg.Wait()
	out := results[:0]
	for _, r := range results {
		if !errors.Is(r.err, errSkipSelf) {
			out = append(out, r)
		}
	}
	return out
}

var errSkipSelf = errors.New("skip self")

var clusterCallErrors = map[string]error{}

func registerCallErrors(errs ...error) {
	for _, err := range errs {
		clusterCallErrors[err.Error()] = err
	}
}

func mapCallError(err error) error {
	var ce *cluster.CallError
	if errors.As(err, &ce) {
		if known, ok := clusterCallErrors[ce.Message]; ok {
			return known
		}
		return errors.New(ce.Message)
	}
	return err
}

func (c *Cluster) HandleFrame(from string, kind byte, payload []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("Cluster frame handler panic", zap.String("from", from), zap.Uint8("kind", kind), zap.Any("panic", r))
		}
	}()
	switch kind {
	case cluster.FrameResync:
		if c.tracker != nil {
			c.tracker.resyncNode(from)
		}
	case framePresence:
		if c.tracker != nil {
			c.tracker.handlePresence(from, payload)
		}
	case frameUntrack:
		if c.tracker != nil {
			c.tracker.handleUntrack(payload)
		}
	case frameUntrackStream:
		if c.tracker != nil {
			c.tracker.handleUntrackStream(payload)
		}
	case frameHostLost:
		c.hostLost(string(payload))
	case frameMatchData:
		if c.matchRegistry != nil {
			c.matchRegistry.handleData(payload)
		}
	case frameDisconnect, frameSingleSession:
		c.handleSessionFrame(kind, payload)
	case frameLeaderboard:
		if c.scheduler != nil {
			c.scheduler.handleChange(payload)
		}
	case frameEnvelopeSessions:
		c.router.handleSessions(payload)
	case frameEnvelopeStream:
		c.router.handleStream(payload)
	case frameEnvelopeAll:
		c.router.handleAll(payload)
	default:
		c.logger.Warn("Unknown cluster frame", zap.String("from", from), zap.Uint8("kind", kind))
	}
}

func (c *Cluster) HandleCall(ctx context.Context, from string, kind byte, payload []byte) ([]byte, error) {
	switch kind {
	case callNodeStatus:
		return c.statusHandler.handleCall(ctx)
	case callStreamCounts:
		return c.tracker.handleCounts(payload)
	case callStreamUser:
		if c.streamManager == nil {
			return nil, ErrSessionNotFound
		}
		return c.streamManager.handleCall(payload)
	case callMatchGet, callMatchJoinAttempt, callMatchSignal, callMatchState, callMatchList:
		if c.matchRegistry == nil {
			return nil, runtime.ErrMatchNotFound
		}
		return c.matchRegistry.handleCall(ctx, kind, payload)
	case callParty, callPartyList:
		if c.partyRegistry == nil {
			return nil, ErrPartyNotFound
		}
		return c.partyRegistry.handleCall(ctx, kind, payload)
	case callMatchmakerAdd, callMatchmakerRemove:
		if c.matchmaker == nil {
			return nil, runtime.ErrMatchmakerNotAvailable
		}
		return c.matchmaker.handleCall(ctx, kind, payload)
	}
	return nil, fmt.Errorf("unknown cluster call %d", kind)
}

func (c *Cluster) PeerReset(node string) {
	if c.tracker != nil {
		c.tracker.resyncNode(node)
	}
}

func (c *Cluster) holder() string {
	return c.name + "/" + strconv.Itoa(os.Getpid())
}

func (c *Cluster) onNodeDead(info cluster.NodeInfo) {
	c.reapMu.Lock()
	if _, busy := c.reaping[info.Field()]; busy {
		c.reapMu.Unlock()
		return
	}
	c.reaping[info.Field()] = struct{}{}
	c.reapMu.Unlock()
	go func() {
		defer func() {
			c.reapMu.Lock()
			delete(c.reaping, info.Field())
			c.reapMu.Unlock()
		}()
		c.reap(info)
	}()
}

func (c *Cluster) reap(info cluster.NodeInfo) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Minute)
	defer cancel()
	ok, err := c.members.TryReap(ctx, info, c.holder())
	if err != nil || !ok {
		return
	}
	start := time.Now()
	c.reapState(ctx, info)
	if err := c.members.FinishReap(ctx, info, c.holder()); err != nil {
		c.logger.Error("Cluster reap could not finish", zap.String("node", info.Field()), zap.Error(err))
		return
	}
	c.logger.Warn("Cluster node reaped", zap.String("node", info.Field()), zap.Duration("took", time.Since(start)))
}

func (c *Cluster) reapState(ctx context.Context, info cluster.NodeInfo) {
	n := 0
	if c.tracker != nil {
		n = c.tracker.reapPresences(ctx, info)
	}
	if c.statusRegistry != nil {
		c.statusRegistry.reapFollows(ctx, info)
	}
	c.logger.Warn("Cluster dead node presences removed", zap.String("node", info.Field()), zap.Int("presences", n))
}
