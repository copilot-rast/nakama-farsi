package server

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	registerCallErrors(
		ErrPartyNotFound,
		ErrPartyHiddenNonEmptyLabel,
		runtime.ErrPartyClosed,
		runtime.ErrPartyFull,
		runtime.ErrPartyJoinRequestDuplicate,
		runtime.ErrPartyJoinRequestAlreadyMember,
		runtime.ErrPartyJoinRequestsFull,
		runtime.ErrPartyNotLeader,
		runtime.ErrPartyNotMember,
		runtime.ErrPartyNotRequest,
		runtime.ErrPartyAcceptRequest,
		runtime.ErrPartyRemove,
		runtime.ErrPartyRemoveSelf,
		runtime.ErrPartyLabelTooLong,
		ErrNodeNotFound,
		ErrSessionNotFound,
	)
}

type ClusterPartyRegistry struct {
	c     *Cluster
	local *LocalPartyRegistry
}

var _ PartyRegistry = (*ClusterPartyRegistry)(nil)

type wirePresence struct {
	UserID    string `json:"u"`
	SessionID string `json:"s"`
	Username  string `json:"n"`
}

func toWire(p *rtapi.UserPresence) *wirePresence {
	if p == nil {
		return nil
	}
	return &wirePresence{UserID: p.UserId, SessionID: p.SessionId, Username: p.Username}
}

func fromWire(p *wirePresence) *rtapi.UserPresence {
	if p == nil {
		return nil
	}
	return &rtapi.UserPresence{UserId: p.UserID, SessionId: p.SessionID, Username: p.Username}
}

type partyCall struct {
	Op        string             `json:"op"`
	ID        uuid.UUID          `json:"id"`
	Node      string             `json:"node"`
	SessionID string             `json:"sid"`
	FromNode  string             `json:"from"`
	Presence  *wirePresence      `json:"p,omitempty"`
	Joiner    *Presence          `json:"j,omitempty"`
	Query     string             `json:"q,omitempty"`
	Min       int                `json:"min,omitempty"`
	Max       int                `json:"max,omitempty"`
	Multiple  int                `json:"mul,omitempty"`
	SP        map[string]string  `json:"sp,omitempty"`
	NP        map[string]float64 `json:"np,omitempty"`
	Ticket    string             `json:"t,omitempty"`
	OpCode    int64              `json:"oc,omitempty"`
	Data      []byte             `json:"d,omitempty"`
	Label     string             `json:"l,omitempty"`
	Open      bool               `json:"o,omitempty"`
	Hidden    bool               `json:"h,omitempty"`
}

type partyReply struct {
	Bool      bool            `json:"b,omitempty"`
	Ticket    string          `json:"t,omitempty"`
	IDs       []*PresenceID   `json:"ids,omitempty"`
	Presences []*wirePresence `json:"ps,omitempty"`
}

type partyListCall struct {
	Limit      int    `json:"limit"`
	Open       *bool  `json:"open,omitempty"`
	ShowHidden bool   `json:"hidden"`
	Query      string `json:"query"`
}

func (r *ClusterPartyRegistry) remote(ctx context.Context, req *partyCall) (*partyReply, error) {
	payload, _ := json.Marshal(req)
	out, err := r.c.call(ctx, req.Node, callParty, payload)
	if err != nil {
		if !r.c.members.IsAlive(req.Node) {
			return nil, ErrPartyNotFound
		}
		return nil, err
	}
	reply := &partyReply{}
	if len(out) > 0 {
		if err := json.Unmarshal(out, reply); err != nil {
			return nil, err
		}
	}
	return reply, nil
}

func (r *ClusterPartyRegistry) isLocal(node string) bool {
	return node == r.c.name
}

func (r *ClusterPartyRegistry) Init(matchmaker Matchmaker, tracker Tracker, streamManager StreamManager, router MessageRouter) {
	r.local.Init(matchmaker, tracker, streamManager, router)
}

func (r *ClusterPartyRegistry) Create(open, hidden bool, maxSize int, leader *rtapi.UserPresence, label string) (*PartyHandler, error) {
	return r.local.Create(open, hidden, maxSize, leader, label)
}

func (r *ClusterPartyRegistry) Delete(id uuid.UUID) {
	r.local.Delete(id)
}

func (r *ClusterPartyRegistry) Count() int {
	return r.local.Count()
}

func (r *ClusterPartyRegistry) Join(id uuid.UUID, presences []*Presence) {
	r.local.Join(id, presences)
}

func (r *ClusterPartyRegistry) Leave(id uuid.UUID, presences []*Presence) {
	r.local.Leave(id, presences)
}

func (r *ClusterPartyRegistry) PartyJoinRequest(ctx context.Context, id uuid.UUID, node string, presence *Presence) (bool, error) {
	if r.isLocal(node) {
		return r.local.PartyJoinRequest(ctx, id, node, presence)
	}
	reply, err := r.remote(ctx, &partyCall{Op: "join", ID: id, Node: node, Joiner: presence})
	if err != nil {
		return false, err
	}
	return reply.Bool, nil
}

func (r *ClusterPartyRegistry) PartyPromote(ctx context.Context, id uuid.UUID, node, sessionID, fromNode string, presence *rtapi.UserPresence) error {
	if r.isLocal(node) {
		return r.local.PartyPromote(ctx, id, node, sessionID, fromNode, presence)
	}
	_, err := r.remote(ctx, &partyCall{Op: "promote", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode, Presence: toWire(presence)})
	return err
}

func (r *ClusterPartyRegistry) PartyAccept(ctx context.Context, id uuid.UUID, node, sessionID, fromNode string, presence *rtapi.UserPresence) error {
	if r.isLocal(node) {
		return r.local.PartyAccept(ctx, id, node, sessionID, fromNode, presence)
	}
	_, err := r.remote(ctx, &partyCall{Op: "accept", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode, Presence: toWire(presence)})
	return err
}

func (r *ClusterPartyRegistry) PartyRemove(ctx context.Context, id uuid.UUID, node, sessionID, fromNode string, presence *rtapi.UserPresence) error {
	if r.isLocal(node) {
		return r.local.PartyRemove(ctx, id, node, sessionID, fromNode, presence)
	}
	_, err := r.remote(ctx, &partyCall{Op: "remove", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode, Presence: toWire(presence)})
	return err
}

func (r *ClusterPartyRegistry) PartyClose(ctx context.Context, id uuid.UUID, node, sessionID, fromNode string) error {
	if r.isLocal(node) {
		return r.local.PartyClose(ctx, id, node, sessionID, fromNode)
	}
	_, err := r.remote(ctx, &partyCall{Op: "close", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode})
	return err
}

func (r *ClusterPartyRegistry) PartyJoinRequestList(ctx context.Context, id uuid.UUID, node, sessionID, fromNode string) ([]*rtapi.UserPresence, error) {
	if r.isLocal(node) {
		return r.local.PartyJoinRequestList(ctx, id, node, sessionID, fromNode)
	}
	reply, err := r.remote(ctx, &partyCall{Op: "requests", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode})
	if err != nil {
		return nil, err
	}
	out := make([]*rtapi.UserPresence, 0, len(reply.Presences))
	for _, p := range reply.Presences {
		out = append(out, fromWire(p))
	}
	return out, nil
}

func (r *ClusterPartyRegistry) PartyMatchmakerAdd(ctx context.Context, id uuid.UUID, node, sessionID, fromNode, query string, minCount, maxCount, countMultiple int, stringProperties map[string]string, numericProperties map[string]float64) (string, []*PresenceID, error) {
	if r.isLocal(node) {
		return r.local.PartyMatchmakerAdd(ctx, id, node, sessionID, fromNode, query, minCount, maxCount, countMultiple, stringProperties, numericProperties)
	}
	reply, err := r.remote(ctx, &partyCall{Op: "mmadd", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode, Query: query, Min: minCount, Max: maxCount, Multiple: countMultiple, SP: stringProperties, NP: numericProperties})
	if err != nil {
		return "", nil, err
	}
	return reply.Ticket, reply.IDs, nil
}

func (r *ClusterPartyRegistry) PartyMatchmakerRemove(ctx context.Context, id uuid.UUID, node, sessionID, fromNode, ticket string) error {
	if r.isLocal(node) {
		return r.local.PartyMatchmakerRemove(ctx, id, node, sessionID, fromNode, ticket)
	}
	_, err := r.remote(ctx, &partyCall{Op: "mmremove", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode, Ticket: ticket})
	return err
}

func (r *ClusterPartyRegistry) PartyDataSend(ctx context.Context, id uuid.UUID, node, sessionID, fromNode string, opCode int64, data []byte) error {
	if r.isLocal(node) {
		return r.local.PartyDataSend(ctx, id, node, sessionID, fromNode, opCode, data)
	}
	_, err := r.remote(ctx, &partyCall{Op: "data", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode, OpCode: opCode, Data: data})
	return err
}

func (r *ClusterPartyRegistry) PartyUpdate(ctx context.Context, id uuid.UUID, node, sessionID, fromNode, label string, open, hidden bool) error {
	if r.isLocal(node) {
		return r.local.PartyUpdate(ctx, id, node, sessionID, fromNode, label, open, hidden)
	}
	if !(label == "" || label == "{}") && hidden {
		return ErrPartyHiddenNonEmptyLabel
	}
	_, err := r.remote(ctx, &partyCall{Op: "update", ID: id, Node: node, SessionID: sessionID, FromNode: fromNode, Label: label, Open: open, Hidden: hidden})
	return err
}

func (r *ClusterPartyRegistry) PartyList(ctx context.Context, limit int, open *bool, showHidden bool, query, cursor string) ([]*api.Party, string, error) {
	if cursor != "" {
		return r.local.PartyList(ctx, limit, open, showHidden, query, cursor)
	}
	payload, _ := json.Marshal(&partyListCall{Limit: limit, Open: open, ShowHidden: showHidden, Query: query})
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var out []*api.Party
	var firstErr error
	for _, res := range r.c.fanout(callCtx, callPartyList, payload, true) {
		if res.err != nil {
			if firstErr == nil && res.node == r.c.name {
				firstErr = res.err
			}
			continue
		}
		list := &api.PartyList{}
		if err := protojson.Unmarshal(res.out, list); err == nil {
			out = append(out, list.Parties...)
		}
	}
	if firstErr != nil && len(out) == 0 {
		return nil, "", firstErr
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PartyId < out[j].PartyId })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, "", nil
}

func (r *ClusterPartyRegistry) LabelUpdate(id uuid.UUID, node, label string, open, hidden bool, maxSize int, createTime time.Time) error {
	return r.local.LabelUpdate(id, node, label, open, hidden, maxSize, createTime)
}

func (r *ClusterPartyRegistry) handleCall(ctx context.Context, kind byte, payload []byte) ([]byte, error) {
	if kind == callPartyList {
		req := &partyListCall{}
		if err := json.Unmarshal(payload, req); err != nil {
			return nil, err
		}
		parties, _, err := r.local.PartyList(ctx, req.Limit, req.Open, req.ShowHidden, req.Query, "")
		if err != nil {
			return nil, err
		}
		return protojson.Marshal(&api.PartyList{Parties: parties})
	}
	req := &partyCall{}
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	l := r.local
	reply := &partyReply{}
	var err error
	switch req.Op {
	case "join":
		reply.Bool, err = l.PartyJoinRequest(ctx, req.ID, req.Node, req.Joiner)
	case "promote":
		err = l.PartyPromote(ctx, req.ID, req.Node, req.SessionID, req.FromNode, fromWire(req.Presence))
	case "accept":
		err = l.PartyAccept(ctx, req.ID, req.Node, req.SessionID, req.FromNode, fromWire(req.Presence))
	case "remove":
		err = l.PartyRemove(ctx, req.ID, req.Node, req.SessionID, req.FromNode, fromWire(req.Presence))
	case "close":
		err = l.PartyClose(ctx, req.ID, req.Node, req.SessionID, req.FromNode)
	case "requests":
		var ps []*rtapi.UserPresence
		ps, err = l.PartyJoinRequestList(ctx, req.ID, req.Node, req.SessionID, req.FromNode)
		for _, p := range ps {
			reply.Presences = append(reply.Presences, toWire(p))
		}
	case "mmadd":
		reply.Ticket, reply.IDs, err = l.PartyMatchmakerAdd(ctx, req.ID, req.Node, req.SessionID, req.FromNode, req.Query, req.Min, req.Max, req.Multiple, req.SP, req.NP)
	case "mmremove":
		err = l.PartyMatchmakerRemove(ctx, req.ID, req.Node, req.SessionID, req.FromNode, req.Ticket)
	case "data":
		err = l.PartyDataSend(ctx, req.ID, req.Node, req.SessionID, req.FromNode, req.OpCode, req.Data)
	case "update":
		err = l.PartyUpdate(ctx, req.ID, req.Node, req.SessionID, req.FromNode, req.Label, req.Open, req.Hidden)
	default:
		err = errors.New("unknown party call")
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(reply)
}

type ClusterStreamManager struct {
	c     *Cluster
	local StreamManager
}

var _ StreamManager = (*ClusterStreamManager)(nil)

type streamUserCall struct {
	Op          string         `json:"op"`
	Stream      PresenceStream `json:"stream"`
	UserID      uuid.UUID      `json:"uid"`
	SessionID   uuid.UUID      `json:"sid"`
	Hidden      bool           `json:"hidden"`
	Persistence bool           `json:"persistence"`
	Status      string         `json:"status"`
}

type streamUserReply struct {
	Success bool `json:"ok"`
	New     bool `json:"new"`
}

func (m *ClusterStreamManager) route(sessionID uuid.UUID) (string, bool) {
	node, ok := m.c.NodeForSession(sessionID)
	if !ok {
		return "", false
	}
	return node, node != m.c.name
}

func (m *ClusterStreamManager) forward(node string, req *streamUserCall) (*streamUserReply, error) {
	payload, _ := json.Marshal(req)
	ctx, cancel := context.WithTimeout(m.c.ctx, 5*time.Second)
	defer cancel()
	out, err := m.c.call(ctx, node, callStreamUser, payload)
	if err != nil {
		return nil, err
	}
	reply := &streamUserReply{}
	if err := json.Unmarshal(out, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (m *ClusterStreamManager) UserJoin(stream PresenceStream, userID, sessionID uuid.UUID, hidden, persistence bool, status string) (bool, bool, error) {
	if node, remote := m.route(sessionID); remote {
		reply, err := m.forward(node, &streamUserCall{Op: "join", Stream: stream, UserID: userID, SessionID: sessionID, Hidden: hidden, Persistence: persistence, Status: status})
		if err != nil {
			return false, false, err
		}
		return reply.Success, reply.New, nil
	}
	return m.local.UserJoin(stream, userID, sessionID, hidden, persistence, status)
}

func (m *ClusterStreamManager) UserUpdate(stream PresenceStream, userID, sessionID uuid.UUID, hidden, persistence bool, status string) (bool, error) {
	if node, remote := m.route(sessionID); remote {
		reply, err := m.forward(node, &streamUserCall{Op: "update", Stream: stream, UserID: userID, SessionID: sessionID, Hidden: hidden, Persistence: persistence, Status: status})
		if err != nil {
			return false, err
		}
		return reply.Success, nil
	}
	return m.local.UserUpdate(stream, userID, sessionID, hidden, persistence, status)
}

func (m *ClusterStreamManager) UserLeave(stream PresenceStream, userID, sessionID uuid.UUID) error {
	if node, remote := m.route(sessionID); remote {
		_, err := m.forward(node, &streamUserCall{Op: "leave", Stream: stream, UserID: userID, SessionID: sessionID})
		return err
	}
	return m.local.UserLeave(stream, userID, sessionID)
}

func (m *ClusterStreamManager) handleCall(payload []byte) ([]byte, error) {
	req := &streamUserCall{}
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	reply := &streamUserReply{}
	var err error
	switch req.Op {
	case "join":
		reply.Success, reply.New, err = m.local.UserJoin(req.Stream, req.UserID, req.SessionID, req.Hidden, req.Persistence, req.Status)
		if err == nil && reply.Success && req.Stream.Mode == StreamModeParty && m.c.tracker != nil {
			m.c.tracker.local.UntrackLocalByModes(req.SessionID, partyStreamMode, req.Stream)
		}
	case "update":
		reply.Success, err = m.local.UserUpdate(req.Stream, req.UserID, req.SessionID, req.Hidden, req.Persistence, req.Status)
	case "leave":
		err = m.local.UserLeave(req.Stream, req.UserID, req.SessionID)
		reply.Success = err == nil
	default:
		err = errors.New("unknown stream call")
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(reply)
}
