package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/console"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
)

type consoleStatus = console.StatusList_Status

var protoUnmarshal = proto.Unmarshal

type clusterTestSession struct {
	id       uuid.UUID
	userID   uuid.UUID
	username string
	format   SessionFormat
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	closeMu  sync.Mutex
	got      []*rtapi.Envelope
	closed   bool
	onClose  func()
}

func (s *clusterTestSession) Logger() *zap.Logger          { return zap.NewNop() }
func (s *clusterTestSession) ID() uuid.UUID                { return s.id }
func (s *clusterTestSession) UserID() uuid.UUID            { return s.userID }
func (s *clusterTestSession) Vars() map[string]string      { return nil }
func (s *clusterTestSession) Headers() map[string][]string { return nil }
func (s *clusterTestSession) ClientIP() string             { return "" }
func (s *clusterTestSession) ClientPort() string           { return "" }
func (s *clusterTestSession) Lang() string                 { return "en" }
func (s *clusterTestSession) Context() context.Context     { return s.ctx }
func (s *clusterTestSession) Username() string             { return s.username }
func (s *clusterTestSession) SetUsername(u string)         { s.username = u }
func (s *clusterTestSession) Expiry() int64                { return 0 }
func (s *clusterTestSession) Consume()                     {}
func (s *clusterTestSession) Format() SessionFormat        { return s.format }
func (s *clusterTestSession) CloseLock()                   { s.closeMu.Lock() }
func (s *clusterTestSession) CloseUnlock()                 { s.closeMu.Unlock() }
func (s *clusterTestSession) Send(env *rtapi.Envelope, _ bool) error {
	s.mu.Lock()
	s.got = append(s.got, env)
	s.mu.Unlock()
	return nil
}
func (s *clusterTestSession) SendBytes(payload []byte, _ bool) error {
	env := &rtapi.Envelope{}
	if s.format == SessionFormatProtobuf {
		if err := protoUnmarshal(payload, env); err != nil {
			return err
		}
	} else if err := protojsonUnmarshaler.Unmarshal(payload, env); err != nil {
		return err
	}
	return s.Send(env, true)
}
func (s *clusterTestSession) Close(_ string, _ runtime.PresenceReason, envs ...*rtapi.Envelope) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.got = append(s.got, envs...)
	s.mu.Unlock()
	if s.onClose != nil {
		s.onClose()
	}
	s.cancel()
}

func (s *clusterTestSession) envelopes() []*rtapi.Envelope {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*rtapi.Envelope{}, s.got...)
}

func (s *clusterTestSession) count(pred func(*rtapi.Envelope) bool) int {
	n := 0
	for _, e := range s.envelopes() {
		if pred(e) {
			n++
		}
	}
	return n
}

type clusterTestNode struct {
	t        *testing.T
	name     string
	config   *config
	cluster  *Cluster
	sessions SessionRegistry
	cache    SessionCache
	streams  StreamManager
	status   StatusRegistry
	tracker  Tracker
	router   MessageRouter
	handler  StatusHandler
}

func clusterTestRedis(t *testing.T) string {
	addr := os.Getenv("CLUSTER_TEST_REDIS")
	if addr == "" {
		t.Skip("CLUSTER_TEST_REDIS not set")
	}
	return addr
}

func newClusterTestNode(t *testing.T, name, prefix string) *clusterTestNode {
	t.Helper()
	log := NewJSONLogger(os.Stdout, zapcore.ErrorLevel, JSONFormat)
	c := NewConfig(log)
	c.Name = name
	c.Cluster.Enabled = true
	c.Cluster.RedisAddress = clusterTestRedis(t)
	c.Cluster.RedisDB = 14
	c.Cluster.Prefix = prefix
	c.Cluster.Port = 0
	c.Cluster.AdvertiseAddress = "127.0.0.1"
	c.Cluster.HeartbeatMs = 50
	c.Cluster.NodeTTLMs = 400
	c.Cluster.Secret = "test-secret"
	c.Matchmaker.IntervalSec = 1
	n := &clusterTestNode{t: t, name: name, config: c}
	cl := NewCluster(log, log, c, protojsonMarshaler, protojsonUnmarshaler)
	n.cluster = cl
	n.sessions = cl.SessionRegistry(NewLocalSessionRegistry(metrics))
	n.status = cl.StatusRegistry(NewLocalStatusRegistry(log, c, n.sessions, protojsonMarshaler))
	n.tracker = cl.Tracker(StartLocalTracker(log, c, n.sessions, n.status, metrics, protojsonMarshaler))
	n.router = cl.Router(NewLocalMessageRouter(n.sessions, n.tracker, protojsonMarshaler))
	n.cache = cl.SessionCache(NewLocalSessionCache(3600, 7200), false)
	n.streams = cl.StreamManager(NewLocalStreamManager(c, n.sessions, n.tracker))
	n.handler = cl.StatusHandler(&fixedStatusHandler{name: name, sessions: n.sessions})
	n.tracker.SetMatchJoinListener(func(uuid.UUID, []*MatchPresence) {})
	n.tracker.SetMatchLeaveListener(func(uuid.UUID, []*MatchPresence) {})
	n.tracker.SetPartyJoinListener(func(uuid.UUID, []*Presence) {})
	n.tracker.SetPartyLeaveListener(func(uuid.UUID, []*Presence) {})
	clusterTestExtra(n, log)
	cl.Start()
	t.Cleanup(func() { n.stop() })
	return n
}

var clusterTestExtra = func(n *clusterTestNode, log *zap.Logger) {}

func (n *clusterTestNode) stop() {
	n.cluster.Stop()
	n.tracker.Stop()
	n.status.Stop()
}

func (n *clusterTestNode) kill() {
	n.cluster.members.KillForTest()
	n.cluster.transport.Stop()
	n.cluster.cancel()
}

type fixedStatusHandler struct {
	name     string
	sessions SessionRegistry
}

func (f *fixedStatusHandler) GetStatus(context.Context) ([]*consoleStatus, error) {
	return []*consoleStatus{{Name: f.name, SessionCount: int32(f.sessions.Count())}}, nil
}

func (n *clusterTestNode) connect(user uuid.UUID, username string, format SessionFormat) *clusterTestSession {
	gen := uuid.NewGenWithHWAF(func() (net.HardwareAddr, error) {
		h := NodeToHash(n.name)
		return h[:], nil
	})
	id := uuid.Must(gen.NewV1())
	ctx, cancel := context.WithCancel(context.Background())
	s := &clusterTestSession{id: id, userID: user, username: username, format: format, ctx: ctx, cancel: cancel}
	s.onClose = func() {
		n.tracker.UntrackAll(id, runtime.PresenceReasonDisconnect)
		n.status.UnfollowAll(id)
		n.sessions.Remove(id)
	}
	n.sessions.Add(s)
	return s
}

func testPrefix() string {
	return fmt.Sprintf("ct%d", time.Now().UnixNano())
}

func waitCluster(t *testing.T, nodes ...*clusterTestNode) {
	t.Helper()
	clusterEventually(t, 5*time.Second, func() bool {
		for _, n := range nodes {
			if len(n.cluster.transport.Peers()) != len(nodes)-1 {
				return false
			}
		}
		return true
	}, "nodes did not connect")
}

func clusterEventually(t *testing.T, d time.Duration, fn func() bool, msg string, args ...any) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf(msg, args...)
}

func isStreamData(e *rtapi.Envelope) bool { return e.GetStreamData() != nil }

func TestClusterRoutesToRemoteSessionsStreamsAndAll(t *testing.T) {
	prefix := testPrefix()
	a := newClusterTestNode(t, "ra", prefix)
	b := newClusterTestNode(t, "rb", prefix)
	c := newClusterTestNode(t, "rc", prefix)
	waitCluster(t, a, b, c)

	ua, ub := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	sa := a.connect(ua, "a", SessionFormatJson)
	sb := b.connect(ub, "b", SessionFormatProtobuf)
	stream := PresenceStream{Mode: StreamModeChannel, Label: "room"}
	if ok, _ := b.tracker.Track(context.Background(), sb.id, stream, ub, PresenceMeta{Username: "b"}); !ok {
		t.Fatal("track failed")
	}

	msg := func(s string) *rtapi.Envelope {
		return &rtapi.Envelope{Message: &rtapi.Envelope_StreamData{StreamData: &rtapi.StreamData{Data: s}}}
	}
	a.router.SendToPresenceIDs(zap.NewNop(), []*PresenceID{{Node: "rb", SessionID: sb.id}, {Node: "ra", SessionID: sa.id}}, msg("direct"), true)
	clusterEventually(t, 2*time.Second, func() bool { return sb.count(isStreamData) == 1 && sa.count(isStreamData) == 1 }, "direct message not delivered on both nodes")

	c.router.SendToStream(zap.NewNop(), stream, msg("stream"), true)
	clusterEventually(t, 2*time.Second, func() bool { return sb.count(isStreamData) == 2 }, "stream message not delivered to the remote member")
	time.Sleep(100 * time.Millisecond)
	if sa.count(isStreamData) != 1 {
		t.Fatal("non-member got the stream message")
	}

	c.router.SendToAll(zap.NewNop(), msg("all"), true)
	clusterEventually(t, 2*time.Second, func() bool { return sb.count(isStreamData) == 3 && sa.count(isStreamData) == 2 }, "broadcast not delivered")

	st, err := a.handler.GetStatus(context.Background())
	if err != nil || len(st) != 3 || st[0].Name != "ra" || st[1].Name != "rb" || st[1].SessionCount != 1 {
		t.Fatalf("cluster status wrong: %v %v", st, err)
	}
}
