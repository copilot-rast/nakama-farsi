package cluster

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

const (
	serviceName   = "nakama.cluster.Peer"
	methodCall    = "/" + serviceName + "/Call"
	methodPipe    = "/" + serviceName + "/Pipe"
	mdSecret      = "nk-secret"
	mdFrom        = "nk-from"
	maxBatchCount = 256
	maxBatchBytes = 256 << 10
)

var (
	ErrUnknownPeer  = errors.New("cluster: unknown peer")
	ErrUnauthorized = errors.New("cluster: unauthorized peer")
)

type Handler interface {
	HandleFrame(from string, kind byte, payload []byte)
	HandleCall(ctx context.Context, from string, kind byte, payload []byte) ([]byte, error)
	PeerReset(node string)
}

type rawCodec struct{}

func (rawCodec) Marshal(v any) ([]byte, error) {
	switch m := v.(type) {
	case *[]byte:
		return *m, nil
	case []byte:
		return m, nil
	}
	return nil, fmt.Errorf("cluster: cannot marshal %T", v)
}

func (rawCodec) Unmarshal(data []byte, v any) error {
	p, ok := v.(*[]byte)
	if !ok {
		return fmt.Errorf("cluster: cannot unmarshal into %T", v)
	}
	*p = data
	return nil
}

func (rawCodec) Name() string {
	return "nkraw"
}

type peerService interface{}

type TransportStats struct {
	FramesSent    int64
	FramesDropped int64
	FramesRecv    int64
	BatchesSent   int64
	Calls         int64
	CallErrors    int64
	Reconnects    int64
}

type Transport struct {
	logger      *zap.Logger
	name        string
	secret      string
	queueSize   int
	callTimeout time.Duration
	handler     Handler

	server *grpc.Server
	lis    net.Listener

	mu    sync.RWMutex
	peers map[string]*peer

	framesSent    *atomic.Int64
	framesDropped *atomic.Int64
	framesRecv    *atomic.Int64
	batchesSent   *atomic.Int64
	calls         *atomic.Int64
	callErrors    *atomic.Int64
	reconnects    *atomic.Int64
	lastDropLog   *atomic.Int64
}

type peer struct {
	name   string
	addr   string
	conn   *grpc.ClientConn
	calls  *grpc.ClientConn
	q      chan []byte
	ctx    context.Context
	cancel context.CancelFunc
	resync *atomic.Bool
	done   chan struct{}
}

func NewTransport(logger *zap.Logger, name, secret string, queueSize int, callTimeout time.Duration, handler Handler) *Transport {
	if queueSize <= 0 {
		queueSize = 65536
	}
	if callTimeout <= 0 {
		callTimeout = 5 * time.Second
	}
	return &Transport{
		logger:        logger,
		name:          name,
		secret:        secret,
		queueSize:     queueSize,
		callTimeout:   callTimeout,
		handler:       handler,
		peers:         make(map[string]*peer),
		framesSent:    atomic.NewInt64(0),
		framesDropped: atomic.NewInt64(0),
		framesRecv:    atomic.NewInt64(0),
		batchesSent:   atomic.NewInt64(0),
		calls:         atomic.NewInt64(0),
		callErrors:    atomic.NewInt64(0),
		reconnects:    atomic.NewInt64(0),
		lastDropLog:   atomic.NewInt64(0),
	}
}

func (t *Transport) Stats() TransportStats {
	return TransportStats{
		FramesSent:    t.framesSent.Load(),
		FramesDropped: t.framesDropped.Load(),
		FramesRecv:    t.framesRecv.Load(),
		BatchesSent:   t.batchesSent.Load(),
		Calls:         t.calls.Load(),
		CallErrors:    t.callErrors.Load(),
		Reconnects:    t.reconnects.Load(),
	}
}

func (t *Transport) Listen(address string, port int) error {
	lis, err := net.Listen("tcp", net.JoinHostPort(address, fmt.Sprint(port)))
	if err != nil {
		return err
	}
	t.lis = lis
	t.server = grpc.NewServer(
		grpc.ForceServerCodec(rawCodec{}),
		grpc.MaxRecvMsgSize(64<<20),
		grpc.MaxSendMsgSize(64<<20),
		grpc.KeepaliveParams(keepalive.ServerParameters{Time: 10 * time.Second, Timeout: 5 * time.Second}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{MinTime: 5 * time.Second, PermitWithoutStream: true}),
	)
	desc := grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*peerService)(nil),
		Methods:     []grpc.MethodDesc{{MethodName: "Call", Handler: t.serveCall}},
		Streams:     []grpc.StreamDesc{{StreamName: "Pipe", Handler: t.servePipe, ClientStreams: true}},
	}
	t.server.RegisterService(&desc, struct{}{})
	go func() {
		if err := t.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.logger.Error("Cluster transport stopped", zap.Error(err))
		}
	}()
	return nil
}

func (t *Transport) Port() int {
	if t.lis == nil {
		return 0
	}
	return t.lis.Addr().(*net.TCPAddr).Port
}

func (t *Transport) auth(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrUnauthorized
	}
	secrets := md.Get(mdSecret)
	froms := md.Get(mdFrom)
	if len(secrets) != 1 || len(froms) != 1 || froms[0] == "" {
		return "", ErrUnauthorized
	}
	if subtle.ConstantTimeCompare([]byte(secrets[0]), []byte(t.secret)) != 1 {
		return "", ErrUnauthorized
	}
	return froms[0], nil
}

func (t *Transport) serveCall(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	from, err := t.auth(ctx)
	if err != nil {
		return nil, err
	}
	var req []byte
	if err := dec(&req); err != nil {
		return nil, err
	}
	if len(req) < 1 {
		return nil, ErrShortFrame
	}
	out, herr := t.handler.HandleCall(ctx, from, req[0], req[1:])
	resp := make([]byte, 0, len(out)+1)
	if herr != nil {
		resp = append(resp, 1)
		resp = append(resp, herr.Error()...)
	} else {
		resp = append(resp, 0)
		resp = append(resp, out...)
	}
	return &resp, nil
}

func (t *Transport) servePipe(_ any, stream grpc.ServerStream) error {
	from, err := t.auth(stream.Context())
	if err != nil {
		return err
	}
	for {
		var batch []byte
		if err := stream.RecvMsg(&batch); err != nil {
			if stream.Context().Err() == nil {
				t.handler.PeerReset(from)
			}
			return nil
		}
		if err := SplitFrames(batch, func(kind byte, payload []byte) {
			t.framesRecv.Inc()
			t.handler.HandleFrame(from, kind, payload)
		}); err != nil {
			t.logger.Warn("Cluster frame batch could not be parsed", zap.String("from", from), zap.Error(err))
		}
	}
}

func (t *Transport) Stop() {
	t.mu.Lock()
	peers := t.peers
	t.peers = make(map[string]*peer)
	t.mu.Unlock()
	for _, p := range peers {
		p.cancel()
		_ = p.conn.Close()
		_ = p.calls.Close()
	}
	if t.server != nil {
		t.server.Stop()
	}
}

func (t *Transport) HasPeer(name string) bool {
	t.mu.RLock()
	_, ok := t.peers[name]
	t.mu.RUnlock()
	return ok
}

func (t *Transport) Peers() []string {
	t.mu.RLock()
	out := make([]string, 0, len(t.peers))
	for name := range t.peers {
		out = append(out, name)
	}
	t.mu.RUnlock()
	return out
}

func (t *Transport) AddPeer(name, addr string) error {
	if name == t.name {
		return nil
	}
	t.mu.Lock()
	if existing, ok := t.peers[name]; ok {
		if existing.addr == addr {
			t.mu.Unlock()
			return nil
		}
		delete(t.peers, name)
		existing.cancel()
		_ = existing.conn.Close()
		_ = existing.calls.Close()
	}
	dial := func() (*grpc.ClientConn, error) {
		return grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.ForceCodec(rawCodec{}), grpc.MaxCallRecvMsgSize(64<<20), grpc.MaxCallSendMsgSize(64<<20)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 10 * time.Second, Timeout: 5 * time.Second, PermitWithoutStream: true}),
		)
	}
	conn, err := dial()
	if err != nil {
		t.mu.Unlock()
		return err
	}
	calls, err := dial()
	if err != nil {
		_ = conn.Close()
		t.mu.Unlock()
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, mdSecret, t.secret, mdFrom, t.name)
	p := &peer{name: name, addr: addr, conn: conn, calls: calls, q: make(chan []byte, t.queueSize), ctx: ctx, cancel: cancel, resync: atomic.NewBool(false), done: make(chan struct{})}
	t.peers[name] = p
	t.mu.Unlock()
	go t.writeLoop(p)
	return nil
}

func (t *Transport) RemovePeer(name string) {
	t.mu.Lock()
	p, ok := t.peers[name]
	if ok {
		delete(t.peers, name)
	}
	t.mu.Unlock()
	if ok {
		p.cancel()
		_ = p.conn.Close()
		_ = p.calls.Close()
	}
}

func (t *Transport) writeLoop(p *peer) {
	defer close(p.done)
	backoff := 50 * time.Millisecond
	first := true
	for p.ctx.Err() == nil {
		stream, err := p.conn.NewStream(p.ctx, &grpc.StreamDesc{StreamName: "Pipe", ClientStreams: true}, methodPipe)
		if err != nil {
			if !sleepCtx(p.ctx, backoff) {
				return
			}
			backoff = min(backoff*2, 2*time.Second)
			continue
		}
		if !first {
			t.reconnects.Inc()
			t.handler.PeerReset(p.name)
		}
		first = false
		backoff = 50 * time.Millisecond
		if err := t.pump(p, stream); err != nil && p.ctx.Err() == nil {
			t.logger.Warn("Cluster peer stream broken, reconnecting", zap.String("peer", p.name), zap.Error(err))
			first = false
			if !sleepCtx(p.ctx, backoff) {
				return
			}
		}
	}
}

func (t *Transport) pump(p *peer, stream grpc.ClientStream) error {
	defer func() { _ = stream.CloseSend() }()
	for {
		var f []byte
		select {
		case <-p.ctx.Done():
			return nil
		case f = <-p.q:
		}
		batch := make([]byte, 0, max(len(f), 4096))
		batch = append(batch, f...)
		n := 1
	fill:
		for n < maxBatchCount && len(batch) < maxBatchBytes {
			select {
			case f2 := <-p.q:
				batch = append(batch, f2...)
				n++
			default:
				break fill
			}
		}
		if err := stream.SendMsg(&batch); err != nil {
			return err
		}
		t.framesSent.Add(int64(n))
		t.batchesSent.Inc()
		if p.resync.Load() && len(p.q) == 0 {
			p.resync.Store(false)
			resync := AppendFrame(nil, FrameResync, nil)
			if err := stream.SendMsg(&resync); err != nil {
				p.resync.Store(true)
				return err
			}
		}
	}
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

const FrameResync byte = 0

func (t *Transport) Send(node string, kind byte, payload []byte) bool {
	t.mu.RLock()
	p, ok := t.peers[node]
	t.mu.RUnlock()
	if !ok {
		return false
	}
	return t.enqueue(p, AppendFrame(make([]byte, 0, len(payload)+6), kind, payload))
}

func (t *Transport) enqueue(p *peer, frame []byte) bool {
	select {
	case p.q <- frame:
		return true
	default:
		t.framesDropped.Inc()
		p.resync.Store(true)
		now := time.Now().Unix()
		if last := t.lastDropLog.Load(); now-last >= 5 && t.lastDropLog.CompareAndSwap(last, now) {
			t.logger.Warn("Cluster peer queue full, dropping frames", zap.String("peer", p.name), zap.Int64("dropped_total", t.framesDropped.Load()))
		}
		return false
	}
}

func (t *Transport) Broadcast(kind byte, payload []byte) {
	frame := AppendFrame(make([]byte, 0, len(payload)+6), kind, payload)
	t.mu.RLock()
	peers := make([]*peer, 0, len(t.peers))
	for _, p := range t.peers {
		peers = append(peers, p)
	}
	t.mu.RUnlock()
	for _, p := range peers {
		t.enqueue(p, frame)
	}
}

type CallError struct {
	Message string
}

func (e *CallError) Error() string {
	return e.Message
}

func (t *Transport) Call(ctx context.Context, node string, kind byte, payload []byte) ([]byte, error) {
	t.mu.RLock()
	p, ok := t.peers[node]
	t.mu.RUnlock()
	if !ok {
		return nil, ErrUnknownPeer
	}
	t.calls.Inc()
	if _, has := ctx.Deadline(); !has {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.callTimeout)
		defer cancel()
	}
	ctx = metadata.AppendToOutgoingContext(ctx, mdSecret, t.secret, mdFrom, t.name)
	req := make([]byte, 0, len(payload)+1)
	req = append(req, kind)
	req = append(req, payload...)
	var resp []byte
	if err := p.calls.Invoke(ctx, methodCall, &req, &resp); err != nil {
		t.callErrors.Inc()
		return nil, err
	}
	if len(resp) < 1 {
		t.callErrors.Inc()
		return nil, ErrShortFrame
	}
	if resp[0] != 0 {
		return nil, &CallError{Message: string(resp[1:])}
	}
	return resp[1:], nil
}
