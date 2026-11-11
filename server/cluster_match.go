package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func init() {
	registerCallErrors(runtime.ErrMatchIdInvalid, runtime.ErrMatchNotFound, runtime.ErrMatchBusy, runtime.ErrMatchStateFailed, runtime.ErrMatchLabelTooLong)
}

type ClusterMatchRegistry struct {
	c     *Cluster
	local *LocalMatchRegistry
}

var _ MatchRegistry = (*ClusterMatchRegistry)(nil)

type matchListCall struct {
	Limit         int     `json:"limit"`
	Authoritative *bool   `json:"auth,omitempty"`
	Label         *string `json:"label,omitempty"`
	MinSize       *int32  `json:"min,omitempty"`
	MaxSize       *int32  `json:"max,omitempty"`
	Query         *string `json:"query,omitempty"`
	Node          *string `json:"node,omitempty"`
}

type matchListReply struct {
	Matches []json.RawMessage `json:"matches"`
	Nodes   []string          `json:"nodes"`
}

type matchCall struct {
	ID         string            `json:"id,omitempty"`
	MatchID    uuid.UUID         `json:"mid,omitempty"`
	Node       string            `json:"node,omitempty"`
	UserID     uuid.UUID         `json:"uid,omitempty"`
	SessionID  uuid.UUID         `json:"sid,omitempty"`
	Username   string            `json:"un,omitempty"`
	Expiry     int64             `json:"exp,omitempty"`
	Vars       map[string]string `json:"vars,omitempty"`
	ClientIP   string            `json:"ip,omitempty"`
	ClientPort string            `json:"port,omitempty"`
	FromNode   string            `json:"from,omitempty"`
	Metadata   map[string]string `json:"md,omitempty"`
	Data       string            `json:"data,omitempty"`
}

type matchJoinReply struct {
	Found     bool             `json:"found"`
	Allow     bool             `json:"allow"`
	IsNew     bool             `json:"new"`
	Reason    string           `json:"reason"`
	Label     string           `json:"label"`
	Presences []*MatchPresence `json:"presences"`
}

type matchGetReply struct {
	Match json.RawMessage `json:"match,omitempty"`
	Node  string          `json:"node"`
}

type matchStateReply struct {
	Presences []*wirePresence `json:"presences"`
	Tick      int64           `json:"tick"`
	State     string          `json:"state"`
}

func matchNode(id string) (string, bool) {
	parts := strings.SplitN(id, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	return parts[1], true
}

func (r *ClusterMatchRegistry) remoteNode(node string) bool {
	return node != "" && node != r.c.name
}

func (r *ClusterMatchRegistry) callJSON(ctx context.Context, node string, kind byte, req any, reply any) error {
	payload, _ := json.Marshal(req)
	out, err := r.c.call(ctx, node, kind, payload)
	if err != nil {
		return err
	}
	if reply == nil || len(out) == 0 {
		return nil
	}
	return json.Unmarshal(out, reply)
}

func (r *ClusterMatchRegistry) CreateMatch(ctx context.Context, createFn RuntimeMatchCreateFunction, module string, params map[string]any) (string, error) {
	return r.local.CreateMatch(ctx, createFn, module, params)
}

func (r *ClusterMatchRegistry) NewMatch(logger *zap.Logger, id uuid.UUID, core RuntimeMatchCore, stopped *atomic.Bool, params map[string]any) (*MatchHandler, error) {
	return r.local.NewMatch(logger, id, core, stopped, params)
}

func (r *ClusterMatchRegistry) GetMatch(ctx context.Context, id string) (*api.Match, string, error) {
	node, ok := matchNode(id)
	if !ok || !r.remoteNode(node) {
		return r.local.GetMatch(ctx, id)
	}
	if !r.c.members.IsAlive(node) {
		return nil, "", nil
	}
	reply := &matchGetReply{}
	if err := r.callJSON(ctx, node, callMatchGet, &matchCall{ID: id}, reply); err != nil {
		return nil, "", err
	}
	if len(reply.Match) == 0 {
		return nil, "", nil
	}
	m := &api.Match{}
	if err := protojson.Unmarshal(reply.Match, m); err != nil {
		return nil, "", err
	}
	return m, reply.Node, nil
}

func (r *ClusterMatchRegistry) RemoveMatch(id uuid.UUID, stream PresenceStream) {
	r.local.RemoveMatch(id, stream)
}

func (r *ClusterMatchRegistry) UpdateMatchLabel(id uuid.UUID, tickRate int, handlerName, label string, createTime int64) error {
	return r.local.UpdateMatchLabel(id, tickRate, handlerName, label, createTime)
}

func (r *ClusterMatchRegistry) ClearMatchLabel(id uuid.UUID, node string) error {
	return r.local.ClearMatchLabel(id, node)
}

func toListCall(limit int, authoritative *wrapperspb.BoolValue, label *wrapperspb.StringValue, minSize *wrapperspb.Int32Value, maxSize *wrapperspb.Int32Value, query *wrapperspb.StringValue, node *wrapperspb.StringValue) *matchListCall {
	c := &matchListCall{Limit: limit}
	if authoritative != nil {
		v := authoritative.Value
		c.Authoritative = &v
	}
	if label != nil {
		v := label.Value
		c.Label = &v
	}
	if minSize != nil {
		v := minSize.Value
		c.MinSize = &v
	}
	if maxSize != nil {
		v := maxSize.Value
		c.MaxSize = &v
	}
	if query != nil {
		v := query.Value
		c.Query = &v
	}
	if node != nil {
		v := node.Value
		c.Node = &v
	}
	return c
}

func (c *matchListCall) args() (*wrapperspb.BoolValue, *wrapperspb.StringValue, *wrapperspb.Int32Value, *wrapperspb.Int32Value, *wrapperspb.StringValue, *wrapperspb.StringValue) {
	var a *wrapperspb.BoolValue
	var l, q, n *wrapperspb.StringValue
	var mn, mx *wrapperspb.Int32Value
	if c.Authoritative != nil {
		a = wrapperspb.Bool(*c.Authoritative)
	}
	if c.Label != nil {
		l = wrapperspb.String(*c.Label)
	}
	if c.MinSize != nil {
		mn = wrapperspb.Int32(*c.MinSize)
	}
	if c.MaxSize != nil {
		mx = wrapperspb.Int32(*c.MaxSize)
	}
	if c.Query != nil {
		q = wrapperspb.String(*c.Query)
	}
	if c.Node != nil {
		n = wrapperspb.String(*c.Node)
	}
	return a, l, mn, mx, q, n
}

func (r *ClusterMatchRegistry) ListMatches(ctx context.Context, limit int, authoritative *wrapperspb.BoolValue, label *wrapperspb.StringValue, minSize *wrapperspb.Int32Value, maxSize *wrapperspb.Int32Value, query *wrapperspb.StringValue, node *wrapperspb.StringValue) ([]*api.Match, []string, error) {
	if node != nil && r.remoteNode(node.Value) {
		return r.listRemote(ctx, node.Value, toListCall(limit, authoritative, label, minSize, maxSize, query, node))
	}
	matches, nodes, err := r.local.ListMatches(ctx, limit, authoritative, label, minSize, maxSize, query, node)
	if err != nil || node != nil || (authoritative != nil && !authoritative.Value) || limit == 0 {
		return matches, nodes, err
	}
	auth := true
	req := toListCall(limit, wrapperspb.Bool(true), label, minSize, maxSize, query, nil)
	req.Authoritative = &auth
	payload, _ := json.Marshal(req)
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	seen := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		seen[m.MatchId] = struct{}{}
	}
	for _, res := range r.c.fanout(callCtx, callMatchList, payload, false) {
		if res.err != nil {
			continue
		}
		reply := &matchListReply{}
		if json.Unmarshal(res.out, reply) != nil {
			continue
		}
		for i, raw := range reply.Matches {
			m := &api.Match{}
			if protojson.Unmarshal(raw, m) != nil {
				continue
			}
			if _, dup := seen[m.MatchId]; dup {
				continue
			}
			seen[m.MatchId] = struct{}{}
			matches = append(matches, m)
			if i < len(reply.Nodes) {
				nodes = append(nodes, reply.Nodes[i])
			} else {
				nodes = append(nodes, res.node)
			}
		}
	}
	if limit > 0 && len(matches) > limit {
		matches, nodes = matches[:limit], nodes[:limit]
	}
	return matches, nodes, nil
}

func (r *ClusterMatchRegistry) listRemote(ctx context.Context, node string, req *matchListCall) ([]*api.Match, []string, error) {
	reply := &matchListReply{}
	if err := r.callJSON(ctx, node, callMatchList, req, reply); err != nil {
		return nil, nil, err
	}
	out := make([]*api.Match, 0, len(reply.Matches))
	for _, raw := range reply.Matches {
		m := &api.Match{}
		if protojson.Unmarshal(raw, m) == nil {
			out = append(out, m)
		}
	}
	return out, reply.Nodes[:min(len(reply.Nodes), len(out))], nil
}

func (r *ClusterMatchRegistry) Stop(graceSeconds int) chan struct{} {
	return r.local.Stop(graceSeconds)
}

func (r *ClusterMatchRegistry) Count() int {
	return r.local.Count()
}

func (r *ClusterMatchRegistry) JoinAttempt(ctx context.Context, id uuid.UUID, node string, userID, sessionID uuid.UUID, username string, sessionExpiry int64, vars map[string]string, clientIP, clientPort, fromNode string, metadata map[string]string) (bool, bool, bool, string, string, []*MatchPresence) {
	if !r.remoteNode(node) {
		return r.local.JoinAttempt(ctx, id, node, userID, sessionID, username, sessionExpiry, vars, clientIP, clientPort, fromNode, metadata)
	}
	if !r.c.members.IsAlive(node) {
		return false, false, false, "", "", nil
	}
	reply := &matchJoinReply{}
	callCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	if err := r.callJSON(callCtx, node, callMatchJoinAttempt, &matchCall{MatchID: id, Node: node, UserID: userID, SessionID: sessionID, Username: username, Expiry: sessionExpiry, Vars: vars, ClientIP: clientIP, ClientPort: clientPort, FromNode: fromNode, Metadata: metadata}, reply); err != nil {
		r.c.logger.Warn("Remote match join attempt failed", zap.String("node", node), zap.Error(err))
		return false, false, false, "", "", nil
	}
	return reply.Found, reply.Allow, reply.IsNew, reply.Reason, reply.Label, reply.Presences
}

func (r *ClusterMatchRegistry) Join(id uuid.UUID, presences []*MatchPresence) {
	r.local.Join(id, presences)
}

func (r *ClusterMatchRegistry) Leave(id uuid.UUID, presences []*MatchPresence) {
	r.local.Leave(id, presences)
}

func (r *ClusterMatchRegistry) Kick(stream PresenceStream, presences []*MatchPresence) {
	var local []*MatchPresence
	for _, p := range presences {
		if p.Node == r.c.name || p.Node == "" {
			local = append(local, p)
			continue
		}
		if r.c.tracker != nil {
			r.c.tracker.Untrack(p.SessionID, stream, p.UserID)
		}
	}
	if len(local) > 0 {
		r.local.Kick(stream, local)
	}
}

func (r *ClusterMatchRegistry) SendData(id uuid.UUID, node string, userID, sessionID uuid.UUID, username, fromNode string, opCode int64, data []byte, reliable bool, receiveTime int64) {
	if !r.remoteNode(node) {
		r.local.SendData(id, node, userID, sessionID, username, fromNode, opCode, data, reliable, receiveTime)
		return
	}
	e := cluster.NewEnc(len(data) + 96).Raw16(id).Str(node).Raw16(userID).Raw16(sessionID).Str(username).Str(fromNode).Varint(opCode).Bytes(data).Bool(reliable).Varint(receiveTime)
	r.c.send(node, frameMatchData, e.Out())
}

func (r *ClusterMatchRegistry) handleData(payload []byte) {
	d := cluster.NewDec(payload)
	id := uuid.UUID(d.Raw16())
	node := d.Str()
	userID := uuid.UUID(d.Raw16())
	sessionID := uuid.UUID(d.Raw16())
	username := d.Str()
	fromNode := d.Str()
	opCode := d.Varint()
	data := d.Bytes()
	reliable := d.Bool()
	receiveTime := d.Varint()
	if d.Err() != nil {
		return
	}
	r.local.SendData(id, node, userID, sessionID, username, fromNode, opCode, data, reliable, receiveTime)
}

func (r *ClusterMatchRegistry) Signal(ctx context.Context, id, data string) (string, error) {
	node, ok := matchNode(id)
	if !ok || !r.remoteNode(node) {
		return r.local.Signal(ctx, id, data)
	}
	if !r.c.members.IsAlive(node) {
		return "", runtime.ErrMatchNotFound
	}
	var out string
	if err := r.callJSON(ctx, node, callMatchSignal, &matchCall{ID: id, Data: data}, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (r *ClusterMatchRegistry) GetState(ctx context.Context, id uuid.UUID, node string) ([]*rtapi.UserPresence, int64, string, error) {
	if !r.remoteNode(node) {
		return r.local.GetState(ctx, id, node)
	}
	if !r.c.members.IsAlive(node) {
		return nil, 0, "", runtime.ErrMatchNotFound
	}
	reply := &matchStateReply{}
	if err := r.callJSON(ctx, node, callMatchState, &matchCall{MatchID: id, Node: node}, reply); err != nil {
		return nil, 0, "", err
	}
	out := make([]*rtapi.UserPresence, 0, len(reply.Presences))
	for _, p := range reply.Presences {
		out = append(out, fromWire(p))
	}
	return out, reply.Tick, reply.State, nil
}

func (r *ClusterMatchRegistry) handleCall(ctx context.Context, kind byte, payload []byte) ([]byte, error) {
	if kind == callMatchList {
		req := &matchListCall{}
		if err := json.Unmarshal(payload, req); err != nil {
			return nil, err
		}
		a, l, mn, mx, q, n := req.args()
		matches, nodes, err := r.local.ListMatches(ctx, req.Limit, a, l, mn, mx, q, n)
		if err != nil {
			return nil, err
		}
		reply := &matchListReply{Nodes: nodes}
		for _, m := range matches {
			b, err := protojson.Marshal(m)
			if err == nil {
				reply.Matches = append(reply.Matches, b)
			}
		}
		return json.Marshal(reply)
	}
	req := &matchCall{}
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	switch kind {
	case callMatchGet:
		m, node, err := r.local.GetMatch(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		reply := &matchGetReply{Node: node}
		if m != nil {
			reply.Match, _ = protojson.Marshal(m)
		}
		return json.Marshal(reply)
	case callMatchJoinAttempt:
		found, allow, isNew, reason, label, presences := r.local.JoinAttempt(ctx, req.MatchID, req.Node, req.UserID, req.SessionID, req.Username, req.Expiry, req.Vars, req.ClientIP, req.ClientPort, req.FromNode, req.Metadata)
		return json.Marshal(&matchJoinReply{Found: found, Allow: allow, IsNew: isNew, Reason: reason, Label: label, Presences: presences})
	case callMatchSignal:
		out, err := r.local.Signal(ctx, req.ID, req.Data)
		if err != nil {
			return nil, err
		}
		return json.Marshal(out)
	case callMatchState:
		presences, tick, state, err := r.local.GetState(ctx, req.MatchID, req.Node)
		if err != nil {
			return nil, err
		}
		reply := &matchStateReply{Tick: tick, State: state}
		for _, p := range presences {
			reply.Presences = append(reply.Presences, toWire(p))
		}
		return json.Marshal(reply)
	}
	return nil, errors.New("unknown match call")
}
