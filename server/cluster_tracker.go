package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	syncAtomic "sync/atomic"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	NotificationCodeMatchEnded int32 = -100
	clusterTrackerWorkers            = 16
	clusterTrackerBatch              = 512
	clusterViewPendingMax            = 20000
	indexSep                         = "\x1e"
)

type ClusterTracker struct {
	c     *Cluster
	local *LocalTracker
	store *cluster.PresenceStore

	mu    sync.Mutex
	views map[PresenceStream]*streamView

	workers []chan *trackerJob
	dirty   *atomic.Bool
	wg      sync.WaitGroup
	stopped *atomic.Bool
}

var _ Tracker = (*ClusterTracker)(nil)

type viewOp struct {
	leave   bool
	version int64
	deliver bool
	p       *Presence
}

type streamView struct {
	ready       bool
	snapVersion int64
	remote      map[presenceCompact]*Presence
	pending     [][]viewOp
}

type trackerItem struct {
	p        *Presence
	prev     *Presence
	leave    bool
	deliver  bool
	snapshot bool
}

type trackerJob struct {
	items []trackerItem
	done  chan struct{}
}

func newClusterTracker(c *Cluster, local *LocalTracker) *ClusterTracker {
	t := &ClusterTracker{
		c:       c,
		local:   local,
		store:   cluster.NewPresenceStore(c.rdb, c.keys),
		views:   make(map[PresenceStream]*streamView),
		dirty:   atomic.NewBool(false),
		stopped: atomic.NewBool(false),
	}
	t.workers = make([]chan *trackerJob, clusterTrackerWorkers)
	for i := range t.workers {
		t.workers[i] = make(chan *trackerJob, 8192)
	}
	local.clusterHook = t.hook
	return t
}

func (t *ClusterTracker) start() {
	_ = t.store.LoadScripts(t.c.ctx)
	for i := range t.workers {
		t.wg.Add(1)
		go t.work(t.workers[i])
	}
	go t.repairLoop()
}

func (t *ClusterTracker) repairLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-t.c.ctx.Done():
			return
		case <-ticker.C:
		}
		if t.dirty.CompareAndSwap(true, false) {
			t.reconcile(t.c.ctx)
		}
	}
}

func streamRef(s PresenceStream) cluster.StreamRef {
	if s.Mode == StreamModeNotifications || s.Mode == StreamModeStatus {
		return cluster.StreamRef{Tag: "u:" + s.Subject.String(), Suffix: strconv.Itoa(int(s.Mode))}
	}
	return cluster.StreamRef{Tag: "s:" + streamKey(s)}
}

func uuidOrEmpty(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

func streamKey(s PresenceStream) string {
	return strconv.Itoa(int(s.Mode)) + cluster.Sep + uuidOrEmpty(s.Subject) + cluster.Sep + uuidOrEmpty(s.Subcontext) + cluster.Sep + s.Label
}

func parseStreamKey(k string) (PresenceStream, bool) {
	parts := strings.SplitN(k, cluster.Sep, 4)
	if len(parts) != 4 {
		return PresenceStream{}, false
	}
	mode, err := strconv.Atoi(parts[0])
	if err != nil {
		return PresenceStream{}, false
	}
	s := PresenceStream{Mode: uint8(mode), Label: parts[3]}
	if parts[1] != "" {
		s.Subject = uuid.FromStringOrNil(parts[1])
	}
	if parts[2] != "" {
		s.Subcontext = uuid.FromStringOrNil(parts[2])
	}
	return s, true
}

func presenceField(node string, sessionID, userID uuid.UUID) string {
	return node + cluster.Sep + sessionID.String() + cluster.Sep + userID.String()
}

func parsePresenceField(f string) (string, uuid.UUID, uuid.UUID, bool) {
	parts := strings.SplitN(f, cluster.Sep, 3)
	if len(parts) != 3 {
		return "", uuid.Nil, uuid.Nil, false
	}
	sid, err1 := uuid.FromString(parts[1])
	uid, err2 := uuid.FromString(parts[2])
	if err1 != nil || err2 != nil {
		return "", uuid.Nil, uuid.Nil, false
	}
	return parts[0], sid, uid, true
}

func metaValue(m PresenceMeta) string {
	b := func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}
	return strconv.Itoa(int(m.Format)) + cluster.Sep + b(m.Hidden) + cluster.Sep + b(m.Persistence) + cluster.Sep + m.Username + cluster.Sep + m.Status
}

func parseMetaValue(v string) PresenceMeta {
	parts := strings.SplitN(v, cluster.Sep, 5)
	var m PresenceMeta
	if len(parts) != 5 {
		return m
	}
	f, _ := strconv.Atoi(parts[0])
	m.Format = SessionFormat(f)
	m.Hidden = parts[1] == "1"
	m.Persistence = parts[2] == "1"
	m.Username = parts[3]
	m.Status = parts[4]
	return m
}

func parseEntry(stream PresenceStream, field, value string) *Presence {
	node, sid, uid, ok := parsePresenceField(field)
	if !ok {
		return nil
	}
	return &Presence{ID: PresenceID{Node: node, SessionID: sid}, Stream: stream, UserID: uid, Meta: parseMetaValue(value)}
}

func copyPresence(p *Presence, reason runtime.PresenceReason) *Presence {
	return &Presence{ID: p.ID, Stream: p.Stream, UserID: p.UserID, Meta: PresenceMeta{
		Format:      p.Meta.Format,
		Hidden:      p.Meta.Hidden,
		Persistence: p.Meta.Persistence,
		Username:    p.Meta.Username,
		Status:      p.Meta.Status,
		Reason:      uint32(reason),
	}}
}

func (t *ClusterTracker) indexKey() string {
	return t.c.keys.NodePresences(t.c.members.Self().Field())
}

func indexMember(p *Presence) string {
	return streamKey(p.Stream) + indexSep + presenceField(p.ID.Node, p.ID.SessionID, p.UserID)
}

func (t *ClusterTracker) hook(joins, leaves []*Presence, updates [][2]*Presence, reason runtime.PresenceReason, silent bool) {
	items := make([]trackerItem, 0, len(joins)+len(leaves)+len(updates))
	for _, p := range joins {
		cp := copyPresence(p, runtime.PresenceReasonJoin)
		items = append(items, trackerItem{p: cp, deliver: !cp.Meta.Hidden && !silent, snapshot: t.prepareView(cp.Stream)})
	}
	for _, u := range updates {
		prev := copyPresence(u[0], runtime.PresenceReasonUpdate)
		next := copyPresence(u[1], runtime.PresenceReasonUpdate)
		items = append(items, trackerItem{p: next, prev: prev, deliver: !silent, snapshot: t.prepareView(next.Stream)})
	}
	if len(leaves) > 0 {
		released := make(map[PresenceStream]struct{}, len(leaves))
		for _, p := range leaves {
			cp := copyPresence(p, reason)
			items = append(items, trackerItem{p: cp, leave: true, deliver: !cp.Meta.Hidden && !silent})
			released[cp.Stream] = struct{}{}
		}
		for s := range released {
			t.releaseView(s)
		}
	}
	if len(items) == 0 || t.stopped.Load() {
		return
	}
	w := t.workers[int(items[0].p.ID.SessionID[15])%len(t.workers)]
	if len(joins) == 0 && len(updates) == 0 {
		w <- &trackerJob{items: items}
		return
	}
	done := make(chan struct{})
	w <- &trackerJob{items: items, done: done}
	select {
	case <-done:
	case <-t.c.ctx.Done():
	case <-time.After(3 * time.Second):
	}
}

func (t *ClusterTracker) prepareView(s PresenceStream) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.views[s]; ok {
		return false
	}
	t.views[s] = &streamView{remote: make(map[presenceCompact]*Presence)}
	return true
}

func (t *ClusterTracker) releaseView(s PresenceStream) {
	t.mu.Lock()
	if _, ok := t.views[s]; ok && t.local.CountByStream(s) == 0 {
		delete(t.views, s)
	}
	t.mu.Unlock()
}

func (t *ClusterTracker) work(ch chan *trackerJob) {
	defer t.wg.Done()
	for {
		var job *trackerJob
		select {
		case <-t.c.ctx.Done():
			return
		case job = <-ch:
		}
		jobs := []*trackerJob{job}
		n := len(job.items)
	fill:
		for n < clusterTrackerBatch {
			select {
			case j := <-ch:
				jobs = append(jobs, j)
				n += len(j.items)
			default:
				break fill
			}
		}
		items := make([]trackerItem, 0, n)
		for _, j := range jobs {
			items = append(items, j.items...)
		}
		t.process(t.c.ctx, items, t.c.name, t.indexKey())
		for _, j := range jobs {
			if j.done != nil {
				close(j.done)
			}
		}
	}
}

func (t *ClusterTracker) flush() {
	done := make([]chan struct{}, 0, len(t.workers))
	for _, w := range t.workers {
		ch := make(chan struct{})
		select {
		case w <- &trackerJob{done: ch}:
			done = append(done, ch)
		default:
		}
	}
	timeout := time.After(5 * time.Second)
	for _, ch := range done {
		select {
		case <-ch:
		case <-t.c.ctx.Done():
			return
		case <-timeout:
			return
		}
	}
}

type outFrame struct {
	stream PresenceStream
	ops    []viewOp
}

func (t *ClusterTracker) process(ctx context.Context, items []trackerItem, node, index string) {
	if len(items) == 0 {
		return
	}
	ops := make([]cluster.PresenceOp, len(items))
	for i, it := range items {
		ops[i] = cluster.PresenceOp{
			Ref:          streamRef(it.p.Stream),
			Leave:        it.leave,
			Field:        presenceField(it.p.ID.Node, it.p.ID.SessionID, it.p.UserID),
			Meta:         metaValue(it.p.Meta),
			Node:         node,
			Followers:    it.p.Stream.Mode == StreamModeStatus,
			WantSnapshot: it.snapshot,
		}
		if index != "" {
			ops[i].Index = index
			ops[i].IndexMember = indexMember(it.p)
		}
	}
	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	results, err := t.store.Apply(callCtx, ops)
	cancel()
	if err != nil {
		t.c.logger.Error("Cluster presence write failed", zap.Int("ops", len(ops)), zap.Error(err))
		t.dirty.Store(true)
	}
	if len(results) != len(items) {
		return
	}
	frames := make(map[string]map[PresenceStream]*outFrame)
	add := func(target string, s PresenceStream, op viewOp) {
		byStream, ok := frames[target]
		if !ok {
			byStream = make(map[PresenceStream]*outFrame)
			frames[target] = byStream
		}
		f, ok := byStream[s]
		if !ok {
			f = &outFrame{stream: s}
			byStream[s] = f
		}
		f.ops = append(f.ops, op)
	}
	for i, it := range items {
		r := results[i]
		if r.Err != nil {
			continue
		}
		if it.snapshot {
			t.installSnapshot(it.p.Stream, r.Version, r.Snapshot)
		}
		if it.leave && !r.Changed {
			continue
		}
		var vops []viewOp
		if it.prev != nil {
			vops = []viewOp{{leave: true, version: r.Version, deliver: it.deliver && !it.prev.Meta.Hidden, p: it.prev}, {version: r.Version, deliver: it.deliver && !it.p.Meta.Hidden, p: it.p}}
		} else {
			vops = []viewOp{{leave: it.leave, version: r.Version, deliver: it.deliver, p: it.p}}
		}
		for _, target := range t.targets(it.p.Stream, r.Nodes, r.Followers, node) {
			for _, op := range vops {
				add(target, it.p.Stream, op)
			}
		}
	}
	for target, byStream := range frames {
		for _, f := range byStream {
			e := cluster.NewEnc(64 + 96*len(f.ops))
			encStream(e, f.stream)
			e.Uvarint(uint64(len(f.ops)))
			for _, op := range f.ops {
				e.Bool(op.leave).Varint(op.version).Bool(op.deliver)
				encPresence(e, op.p)
			}
			t.c.send(target, framePresence, e.Out())
		}
	}
}

func (t *ClusterTracker) targets(s PresenceStream, nodes, followers []string, exclude string) []string {
	set := make(map[string]struct{}, len(nodes)+len(followers)+1)
	for _, n := range nodes {
		set[n] = struct{}{}
	}
	for _, n := range followers {
		set[n] = struct{}{}
	}
	if (s.Mode == StreamModeMatchAuthoritative || s.Mode == StreamModeParty) && s.Label != "" {
		set[s.Label] = struct{}{}
	}
	delete(set, exclude)
	if exclude != t.c.name {
		set[t.c.name] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for n := range set {
		if n == t.c.name || t.c.members.IsAlive(n) {
			out = append(out, n)
		}
	}
	return out
}

func (t *ClusterTracker) installSnapshot(s PresenceStream, version int64, snap map[string]string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	v, ok := t.views[s]
	if !ok {
		return
	}
	v.remote = make(map[presenceCompact]*Presence, len(snap))
	for field, value := range snap {
		p := parseEntry(s, field, value)
		if p == nil || p.ID.Node == t.c.name {
			continue
		}
		v.remote[presenceCompact{ID: p.ID, Stream: s, UserID: p.UserID}] = p
	}
	v.snapVersion = version
	v.ready = true
	for _, batch := range v.pending {
		applyViewOps(v, batch)
	}
	v.pending = nil
}

func applyViewOps(v *streamView, ops []viewOp) {
	for _, op := range ops {
		if op.version <= v.snapVersion {
			continue
		}
		pc := presenceCompact{ID: op.p.ID, Stream: op.p.Stream, UserID: op.p.UserID}
		if op.leave {
			delete(v.remote, pc)
		} else {
			v.remote[pc] = op.p
		}
	}
}

func (t *ClusterTracker) handlePresence(from string, payload []byte) {
	d := cluster.NewDec(payload)
	s := decStream(d)
	n := int(d.Uvarint())
	if d.Err() != nil || n > d.Len() {
		return
	}
	ops := make([]viewOp, 0, n)
	for i := 0; i < n; i++ {
		op := viewOp{leave: d.Bool(), version: d.Varint(), deliver: d.Bool()}
		op.p = decPresence(d, s)
		if d.Err() != nil {
			return
		}
		if op.p.ID.Node == t.c.name {
			continue
		}
		ops = append(ops, op)
	}
	t.mu.Lock()
	if v, ok := t.views[s]; ok {
		if !v.ready {
			if len(v.pending) < clusterViewPendingMax {
				v.pending = append(v.pending, ops)
			}
		} else {
			applyViewOps(v, ops)
		}
	}
	t.mu.Unlock()
	var joins, leaves []*Presence
	for _, op := range ops {
		if !op.deliver {
			continue
		}
		if op.leave {
			leaves = append(leaves, op.p)
		} else {
			joins = append(joins, op.p)
		}
	}
	if len(joins) > 0 || len(leaves) > 0 {
		t.local.queueEvent(joins, leaves)
	}
}

func (t *ClusterTracker) remote(s PresenceStream) []*Presence {
	t.mu.Lock()
	if v, ok := t.views[s]; ok && v.ready {
		out := make([]*Presence, 0, len(v.remote))
		for _, p := range v.remote {
			out = append(out, p)
		}
		t.mu.Unlock()
		return out
	}
	t.mu.Unlock()
	ctx, cancel := context.WithTimeout(t.c.ctx, 3*time.Second)
	defer cancel()
	lists, err := t.store.List(ctx, []cluster.StreamRef{streamRef(s)})
	if err != nil || len(lists) != 1 {
		return nil
	}
	out := make([]*Presence, 0, len(lists[0]))
	for field, value := range lists[0] {
		if p := parseEntry(s, field, value); p != nil && p.ID.Node != t.c.name && t.c.members.IsAlive(p.ID.Node) {
			out = append(out, p)
		}
	}
	return out
}

func (t *ClusterTracker) remoteIDs(fill map[PresenceStream][]*PresenceID) {
	var missing []PresenceStream
	t.mu.Lock()
	for s := range fill {
		if v, ok := t.views[s]; ok && v.ready {
			for pc := range v.remote {
				id := pc.ID
				fill[s] = append(fill[s], &id)
			}
		} else {
			missing = append(missing, s)
		}
	}
	t.mu.Unlock()
	if len(missing) == 0 {
		return
	}
	refs := make([]cluster.StreamRef, len(missing))
	for i, s := range missing {
		refs[i] = streamRef(s)
	}
	ctx, cancel := context.WithTimeout(t.c.ctx, 3*time.Second)
	defer cancel()
	fields, err := t.store.Fields(ctx, refs)
	if err != nil {
		return
	}
	for i, s := range missing {
		for _, f := range fields[i] {
			node, sid, _, ok := parsePresenceField(f)
			if !ok || node == t.c.name || !t.c.members.IsAlive(node) {
				continue
			}
			fill[s] = append(fill[s], &PresenceID{Node: node, SessionID: sid})
		}
	}
}

func (t *ClusterTracker) remoteNodes(s PresenceStream) []string {
	t.mu.Lock()
	if v, ok := t.views[s]; ok && v.ready {
		set := make(map[string]struct{}, 4)
		for pc := range v.remote {
			set[pc.ID.Node] = struct{}{}
		}
		t.mu.Unlock()
		out := make([]string, 0, len(set))
		for n := range set {
			out = append(out, n)
		}
		return out
	}
	t.mu.Unlock()
	ctx, cancel := context.WithTimeout(t.c.ctx, 3*time.Second)
	defer cancel()
	nodes, err := t.store.Nodes(ctx, streamRef(s))
	if err != nil {
		return nil
	}
	out := nodes[:0]
	for _, n := range nodes {
		if n != t.c.name && t.c.members.IsAlive(n) {
			out = append(out, n)
		}
	}
	return out
}

func (t *ClusterTracker) SetMatchJoinListener(f func(id uuid.UUID, joins []*MatchPresence)) {
	t.local.SetMatchJoinListener(f)
}

func (t *ClusterTracker) SetMatchLeaveListener(f func(id uuid.UUID, leaves []*MatchPresence)) {
	t.local.SetMatchLeaveListener(f)
}

func (t *ClusterTracker) SetPartyJoinListener(f func(id uuid.UUID, joins []*Presence)) {
	t.local.SetPartyJoinListener(f)
}

func (t *ClusterTracker) SetPartyLeaveListener(f func(id uuid.UUID, leaves []*Presence)) {
	t.local.SetPartyLeaveListener(f)
}

func (t *ClusterTracker) Stop() {
	t.local.Stop()
}

func (t *ClusterTracker) Track(ctx context.Context, sessionID uuid.UUID, stream PresenceStream, userID uuid.UUID, meta PresenceMeta) (bool, bool) {
	return t.local.Track(ctx, sessionID, stream, userID, meta)
}

func (t *ClusterTracker) TrackMulti(ctx context.Context, sessionID uuid.UUID, ops []*TrackerOp, userID uuid.UUID) bool {
	return t.local.TrackMulti(ctx, sessionID, ops, userID)
}

func (t *ClusterTracker) remoteSession(sessionID uuid.UUID) (string, bool) {
	node, ok := t.c.NodeForSession(sessionID)
	if !ok || node == t.c.name {
		return "", false
	}
	return node, true
}

func (t *ClusterTracker) Untrack(sessionID uuid.UUID, stream PresenceStream, userID uuid.UUID) {
	if node, ok := t.remoteSession(sessionID); ok {
		e := cluster.NewEnc(80).Raw16(sessionID).Raw16(userID)
		encStream(e, stream)
		t.c.send(node, frameUntrack, e.Out())
		return
	}
	t.local.Untrack(sessionID, stream, userID)
}

func (t *ClusterTracker) UntrackMulti(sessionID uuid.UUID, streams []*PresenceStream, userID uuid.UUID) {
	if _, ok := t.remoteSession(sessionID); ok {
		for _, s := range streams {
			t.Untrack(sessionID, *s, userID)
		}
		return
	}
	t.local.UntrackMulti(sessionID, streams, userID)
}

func (t *ClusterTracker) UntrackAll(sessionID uuid.UUID, reason runtime.PresenceReason) {
	t.local.UntrackAll(sessionID, reason)
}

func (t *ClusterTracker) Update(ctx context.Context, sessionID uuid.UUID, stream PresenceStream, userID uuid.UUID, meta PresenceMeta) bool {
	return t.local.Update(ctx, sessionID, stream, userID, meta)
}

func (t *ClusterTracker) UntrackByStream(stream PresenceStream) {
	nodes := t.remoteNodes(stream)
	t.local.UntrackByStream(stream)
	if len(nodes) == 0 {
		return
	}
	e := cluster.NewEnc(48)
	encStream(e, stream)
	payload := e.Out()
	for _, n := range nodes {
		t.c.send(n, frameUntrackStream, payload)
	}
}

func (t *ClusterTracker) UntrackLocalByStream(stream PresenceStream) {
	t.local.UntrackLocalByStream(stream)
}

func (t *ClusterTracker) UntrackLocalByModes(sessionID uuid.UUID, modes map[uint8]struct{}, skipStream PresenceStream) {
	t.local.UntrackLocalByModes(sessionID, modes, skipStream)
}

func (t *ClusterTracker) ListNodesForStream(stream PresenceStream) map[string]struct{} {
	out := t.local.ListNodesForStream(stream)
	for _, n := range t.remoteNodes(stream) {
		out[n] = struct{}{}
	}
	return out
}

func (t *ClusterTracker) StreamExists(stream PresenceStream) bool {
	if t.local.StreamExists(stream) {
		return true
	}
	return len(t.remote(stream)) > 0
}

func (t *ClusterTracker) Count() int {
	return t.local.Count()
}

func (t *ClusterTracker) CountByStream(stream PresenceStream) int {
	return t.local.CountByStream(stream) + len(t.remote(stream))
}

func (t *ClusterTracker) CountByStreamModeFilter(modes map[uint8]*uint8) map[*PresenceStream]int32 {
	e := cluster.NewEnc(16).Uvarint(uint64(len(modes)))
	for m := range modes {
		e.U8(m)
	}
	ctx, cancel := context.WithTimeout(t.c.ctx, 3*time.Second)
	defer cancel()
	totals := make(map[PresenceStream]int32)
	for _, r := range t.c.fanout(ctx, callStreamCounts, e.Out(), true) {
		if r.err != nil {
			continue
		}
		d := cluster.NewDec(r.out)
		n := int(d.Uvarint())
		for i := 0; i < n && d.Err() == nil; i++ {
			s := decStream(d)
			c := int32(d.Varint())
			if d.Err() == nil {
				totals[s] += c
			}
		}
	}
	out := make(map[*PresenceStream]int32, len(totals))
	for s, c := range totals {
		cs := s
		out[&cs] = c
	}
	return out
}

func (t *ClusterTracker) handleCounts(payload []byte) ([]byte, error) {
	d := cluster.NewDec(payload)
	n := int(d.Uvarint())
	modes := make(map[uint8]*uint8, n)
	for i := 0; i < n; i++ {
		m := d.U8()
		modes[m] = &m
	}
	if d.Err() != nil {
		return nil, d.Err()
	}
	counts := t.local.CountByStreamModeFilter(modes)
	e := cluster.NewEnc(16 + 48*len(counts)).Uvarint(uint64(len(counts)))
	for s, c := range counts {
		encStream(e, *s)
		e.Varint(int64(c))
	}
	return e.Out(), nil
}

func (t *ClusterTracker) GetLocalBySessionIDStreamUserID(sessionID uuid.UUID, stream PresenceStream, userID uuid.UUID) *PresenceMeta {
	return t.local.GetLocalBySessionIDStreamUserID(sessionID, stream, userID)
}

func (t *ClusterTracker) ListByStream(stream PresenceStream, includeHidden bool, includeNotHidden bool) []*Presence {
	out := t.local.ListByStream(stream, includeHidden, includeNotHidden)
	if !includeHidden && !includeNotHidden {
		return out
	}
	for _, p := range t.remote(stream) {
		if (p.Meta.Hidden && includeHidden) || (!p.Meta.Hidden && includeNotHidden) {
			out = append(out, p)
		}
	}
	return out
}

func (t *ClusterTracker) ListLocalSessionIDByStream(stream PresenceStream) []uuid.UUID {
	return t.local.ListLocalSessionIDByStream(stream)
}

func (t *ClusterTracker) ListPresenceIDByStream(stream PresenceStream) []*PresenceID {
	fill := map[PresenceStream][]*PresenceID{stream: t.local.ListPresenceIDByStream(stream)}
	t.remoteIDs(fill)
	return fill[stream]
}

func (t *ClusterTracker) ListPresenceIDByStreams(fill map[PresenceStream][]*PresenceID) {
	if len(fill) == 0 {
		return
	}
	t.local.ListPresenceIDByStreams(fill)
	t.remoteIDs(fill)
}

func (t *ClusterTracker) handleUntrack(payload []byte) {
	d := cluster.NewDec(payload)
	sid := uuid.UUID(d.Raw16())
	uid := uuid.UUID(d.Raw16())
	s := decStream(d)
	if d.Err() == nil {
		t.local.Untrack(sid, s, uid)
	}
}

func (t *ClusterTracker) handleUntrackStream(payload []byte) {
	d := cluster.NewDec(payload)
	s := decStream(d)
	if d.Err() == nil {
		t.local.UntrackLocalByStream(s)
	}
}

func (t *ClusterTracker) localPresences() []*Presence {
	t.local.RLock()
	out := make([]*Presence, 0, t.local.count.Load())
	for _, bySession := range t.local.presencesBySession {
		for _, p := range bySession {
			out = append(out, copyPresence(p, runtime.PresenceReason(syncAtomic.LoadUint32(&p.Meta.Reason))))
		}
	}
	t.local.RUnlock()
	return out
}

func (t *ClusterTracker) resyncNode(node string) {
	go t.resyncAll()
}

func (t *ClusterTracker) resyncAll() {
	t.mu.Lock()
	streams := make([]PresenceStream, 0, len(t.views))
	for s, v := range t.views {
		if v.ready {
			v.ready = false
			v.pending = nil
			streams = append(streams, s)
		}
	}
	t.mu.Unlock()
	for start := 0; start < len(streams); start += 500 {
		end := min(start+500, len(streams))
		batch := streams[start:end]
		refs := make([]cluster.StreamRef, len(batch))
		for i, s := range batch {
			refs[i] = streamRef(s)
		}
		ctx, cancel := context.WithTimeout(t.c.ctx, 10*time.Second)
		snaps, err := t.store.Snapshots(ctx, refs)
		cancel()
		if err != nil {
			t.dirty.Store(true)
			t.c.logger.Error("Cluster view resync failed", zap.Error(err))
			continue
		}
		for i, s := range batch {
			t.installSnapshot(s, snaps[i].Version, snaps[i].Entries)
		}
	}
}

func (t *ClusterTracker) reconcile(ctx context.Context) {
	t.flush()
	local := t.localPresences()
	have := make(map[string]struct{}, len(local))
	items := make([]trackerItem, 0, len(local))
	for _, p := range local {
		have[indexMember(p)] = struct{}{}
		items = append(items, trackerItem{p: p, deliver: !p.Meta.Hidden})
	}
	for start := 0; start < len(items); start += clusterTrackerBatch {
		t.processRepublish(ctx, items[start:min(start+clusterTrackerBatch, len(items))])
	}
	var stale []trackerItem
	_ = t.store.ScanSet(ctx, t.indexKey(), 1000, func(members []string) error {
		for _, m := range members {
			if _, ok := have[m]; ok {
				continue
			}
			if p := parseIndexMember(m); p != nil && p.ID.Node == t.c.name {
				stale = append(stale, trackerItem{p: p, leave: true, deliver: !p.Meta.Hidden})
			}
		}
		return nil
	})
	for start := 0; start < len(stale); start += clusterTrackerBatch {
		t.process(ctx, stale[start:min(start+clusterTrackerBatch, len(stale))], t.c.name, t.indexKey())
	}
	t.resyncAll()
	t.c.logger.Info("Cluster presences republished", zap.Int("presences", len(items)), zap.Int("stale", len(stale)))
}

func (t *ClusterTracker) processRepublish(ctx context.Context, items []trackerItem) {
	ops := make([]cluster.PresenceOp, len(items))
	for i, it := range items {
		ops[i] = cluster.PresenceOp{Ref: streamRef(it.p.Stream), Field: presenceField(it.p.ID.Node, it.p.ID.SessionID, it.p.UserID), Meta: metaValue(it.p.Meta), Node: t.c.name, Followers: it.p.Stream.Mode == StreamModeStatus, Index: t.indexKey(), IndexMember: indexMember(it.p)}
	}
	results, err := t.store.Apply(ctx, ops)
	if err != nil || len(results) != len(items) {
		t.dirty.Store(true)
		return
	}
	for i, r := range results {
		if !r.Changed {
			continue
		}
		it := items[i]
		e := cluster.NewEnc(160)
		encStream(e, it.p.Stream)
		e.Uvarint(1).Bool(false).Varint(r.Version).Bool(it.deliver)
		encPresence(e, it.p)
		for _, target := range t.targets(it.p.Stream, r.Nodes, r.Followers, t.c.name) {
			t.c.send(target, framePresence, e.Out())
		}
	}
}

func parseIndexMember(m string) *Presence {
	i := strings.Index(m, indexSep)
	if i < 0 {
		return nil
	}
	s, ok := parseStreamKey(m[:i])
	if !ok {
		return nil
	}
	return parseEntry(s, m[i+len(indexSep):], "")
}

func (t *ClusterTracker) leaveAll(ctx context.Context) {
	t.stopped.Store(true)
	t.flush()
	local := t.localPresences()
	items := make([]trackerItem, 0, len(local))
	for _, p := range local {
		p.Meta.Reason = uint32(runtime.PresenceReasonDisconnect)
		items = append(items, trackerItem{p: p, leave: true, deliver: !p.Meta.Hidden})
	}
	for start := 0; start < len(items); start += clusterTrackerBatch {
		t.process(ctx, items[start:min(start+clusterTrackerBatch, len(items))], t.c.name, t.indexKey())
	}
	_ = t.store.Del(ctx, t.indexKey())
}

func (t *ClusterTracker) reapPresences(ctx context.Context, info cluster.NodeInfo) int {
	key := t.c.keys.NodePresences(info.Field())
	total := 0
	_ = t.store.ScanSet(ctx, key, 500, func(members []string) error {
		items := make([]trackerItem, 0, len(members))
		for _, m := range members {
			p := parseIndexMember(m)
			if p == nil {
				continue
			}
			items = append(items, trackerItem{p: p, leave: true})
		}
		if len(items) == 0 {
			return nil
		}
		total += t.processReap(ctx, info.Name, items)
		t.c.members.RenewReap(ctx, info)
		return nil
	})
	_ = t.store.Del(ctx, key)
	return total
}

func (t *ClusterTracker) processReap(ctx context.Context, node string, items []trackerItem) int {
	ops := make([]cluster.PresenceOp, len(items))
	for i, it := range items {
		ops[i] = cluster.PresenceOp{Ref: streamRef(it.p.Stream), Leave: true, Field: presenceField(it.p.ID.Node, it.p.ID.SessionID, it.p.UserID), Node: node, Followers: it.p.Stream.Mode == StreamModeStatus}
	}
	results, err := t.store.Apply(ctx, ops)
	if err != nil || len(results) != len(items) {
		return 0
	}
	n := 0
	for i, r := range results {
		if !r.Changed {
			continue
		}
		n++
		p := items[i].p
		p.Meta = parseMetaValue(r.OldMeta)
		p.Meta.Reason = uint32(runtime.PresenceReasonDisconnect)
		e := cluster.NewEnc(160)
		encStream(e, p.Stream)
		e.Uvarint(1).Bool(true).Varint(r.Version).Bool(!p.Meta.Hidden)
		encPresence(e, p)
		for _, target := range t.targets(p.Stream, r.Nodes, r.Followers, node) {
			t.c.send(target, framePresence, e.Out())
		}
	}
	return n
}

func (c *Cluster) hostLost(node string) {
	if c.tracker == nil {
		return
	}
	lt := c.tracker.local
	type lost struct {
		stream   PresenceStream
		sessions []uuid.UUID
	}
	var found []lost
	lt.RLock()
	for _, mode := range []uint8{StreamModeMatchAuthoritative, StreamModeParty} {
		for s, byStream := range lt.presencesByStream[mode] {
			if s.Label != node {
				continue
			}
			l := lost{stream: s}
			for pc := range byStream {
				l.sessions = append(l.sessions, pc.ID.SessionID)
			}
			found = append(found, l)
		}
	}
	lt.RUnlock()
	for _, l := range found {
		id := fmt.Sprintf("%v.%v", l.stream.Subject.String(), l.stream.Label)
		var env *rtapi.Envelope
		if l.stream.Mode == StreamModeParty {
			env = &rtapi.Envelope{Message: &rtapi.Envelope_PartyClose{PartyClose: &rtapi.PartyClose{PartyId: id}}}
		} else {
			content, _ := json.Marshal(map[string]string{"match_id": id, "reason": "node_down", "node": node})
			env = &rtapi.Envelope{Message: &rtapi.Envelope_Notifications{Notifications: &rtapi.Notifications{Notifications: []*api.Notification{{
				Id:         uuid.Must(uuid.NewV4()).String(),
				Subject:    "match_ended",
				Content:    string(content),
				Code:       NotificationCodeMatchEnded,
				CreateTime: &timestamppb.Timestamp{Seconds: time.Now().Unix()},
				Persistent: false,
			}}}}}
		}
		ids := make([]*PresenceID, len(l.sessions))
		for i, sid := range l.sessions {
			ids[i] = &PresenceID{Node: c.name, SessionID: sid}
		}
		c.localRouter.SendToPresenceIDs(c.logger, ids, env, true)
		lt.UntrackLocalByStream(l.stream)
	}
	if len(found) > 0 {
		c.logger.Warn("Cluster host node lost, ended its matches and parties for local players", zap.String("node", node), zap.Int("streams", len(found)))
	}
}
