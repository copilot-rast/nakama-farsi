package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/commpacks"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

const (
	tttTickRate  = 5
	opStart      = 1
	opMove       = 2
	opUpdate     = 3
	opResult     = 4
	opError      = 5
	opComm       = 6
	opCommMsg    = 7
	tttEndTicks  = 10 * tttTickRate
	defaultTurn  = 15
	defaultWait  = 60
	defaultLimit = 300
)

var tttLines = [8][3]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, {0, 4, 8}, {2, 4, 6}}

type tttState struct {
	matchID    string
	game       string
	room       string
	allowed    map[string]bool
	presences  map[string]runtime.Presence
	order      []string
	marks      map[string]string
	board      [9]string
	turn       string
	started    bool
	done       bool
	endTick    int64
	turnTicks  int64
	turnEnd    int64
	waitTicks  int64
	limitTicks int64
	startTick  int64
	firstMove  string
	forfeit    bool
	winner     string
	reason     string
	payouts    []gamecatalog.Payout
	db         *sql.DB
}

type tttMatch struct{}

func (m *tttMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	s := &tttState{allowed: map[string]bool{}, presences: map[string]runtime.Presence{}, marks: map[string]string{}, firstMove: "random", forfeit: true}
	s.matchID, _ = ctx.Value(runtime.RUNTIME_CTX_MATCH_ID).(string)
	s.db = db
	s.game, _ = params["game"].(string)
	if s.game == "" {
		s.game = tttGameKey
	}
	s.room, _ = params["room"].(string)
	if users, _ := params["users"].(string); users != "" {
		for _, u := range strings.Split(users, ",") {
			s.allowed[u] = true
		}
	}
	turn, wait, limit := defaultTurn, defaultWait, defaultLimit
	if g, err := gamecatalog.Settings(ctx, s.game); err == nil {
		if g.TurnSeconds > 0 {
			turn = g.TurnSeconds
		}
		if g.MatchSeconds > 0 {
			limit = g.MatchSeconds
		}
		wait = g.Int("waitSeconds", defaultWait)
		s.firstMove = g.String("firstMove", "random")
		s.forfeit = g.Bool("forfeitOnLeave", true)
	} else {
		logger.Warn("Tic-tac-toe settings not read, defaults used: %v", err)
	}
	s.turnTicks, s.waitTicks, s.limitTicks = int64(turn*tttTickRate), int64(wait*tttTickRate), int64(limit*tttTickRate)
	label, _ := json.Marshal(map[string]any{"game": s.game, "room": s.room, "open": true})
	return s, tttTickRate, string(label)
}

func (m *tttMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	s := state.(*tttState)
	uid := presence.GetUserId()
	if s.done {
		return s, false, "MATCH_OVER"
	}
	if len(s.allowed) > 0 && !s.allowed[uid] {
		return s, false, "NOT_INVITED"
	}
	if _, ok := s.marks[uid]; !ok && len(s.marks) >= 2 {
		return s, false, "MATCH_FULL"
	}
	if _, err := gamecatalog.Enter(ctx, gamecatalog.Entry{MatchID: s.matchID, Game: s.game, Room: s.room, UserID: uid}); err != nil {
		code := gamecatalog.Code(err)
		if code == "" {
			logger.Warn("Tic-tac-toe entry for %s failed: %v", uid, err)
			code = "ENTRY_FAILED"
		}
		return s, false, code
	}
	return s, true, ""
}

func (m *tttMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	s := state.(*tttState)
	for _, p := range presences {
		uid := p.GetUserId()
		s.presences[uid] = p
		if _, ok := s.marks[uid]; !ok {
			s.order = append(s.order, uid)
			s.marks[uid] = ""
		}
	}
	if !s.started && len(s.presences) == 2 && len(s.order) == 2 {
		if err := gamecatalog.Start(ctx, s.matchID); err != nil {
			logger.Warn("Tic-tac-toe start failed: %v", err)
		}
		first := 0
		if s.firstMove == "random" {
			if n, err := rand.Int(rand.Reader, big.NewInt(2)); err == nil {
				first = int(n.Int64())
			}
		}
		s.marks[s.order[first]], s.marks[s.order[1-first]] = "X", "O"
		s.turn = s.order[first]
		s.started, s.startTick, s.turnEnd = true, tick, tick+s.turnTicks
		if err := commpacks.Open(ctx, db, s.matchID, s.order); err != nil {
			logger.Warn("Tic-tac-toe comm channel not opened: %v", err)
		}
		label, _ := json.Marshal(map[string]any{"game": s.game, "room": s.room, "open": false})
		_ = dispatcher.MatchLabelUpdate(string(label))
		m.broadcast(dispatcher, opStart, s, tick)
		return s
	}
	if s.started {
		m.broadcast(dispatcher, opUpdate, s, tick)
	}
	return s
}

func (m *tttMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	s := state.(*tttState)
	for _, p := range presences {
		uid := p.GetUserId()
		delete(s.presences, uid)
		if !s.started {
			if _, err := gamecatalog.Leave(ctx, s.matchID, uid); err != nil {
				logger.Warn("Tic-tac-toe refund for %s failed: %v", uid, err)
			}
			delete(s.marks, uid)
			for i, u := range s.order {
				if u == uid {
					s.order = append(s.order[:i], s.order[i+1:]...)
					break
				}
			}
			continue
		}
		if !s.done && s.forfeit {
			m.finish(ctx, logger, dispatcher, s, tick, m.other(s, uid), "left")
		}
	}
	return s
}

func (m *tttMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	s := state.(*tttState)
	if s.done {
		if tick >= s.endTick {
			return nil
		}
		return s
	}
	if !s.started {
		if tick >= s.waitTicks {
			if _, err := gamecatalog.Cancel(ctx, s.matchID); err != nil {
				logger.Warn("Tic-tac-toe cancel failed: %v", err)
			}
			m.send(dispatcher, opResult, map[string]any{"winner": "", "reason": "cancelled", "refunded": true}, nil)
			return nil
		}
		return s
	}
	for _, msg := range messages {
		if msg.GetOpCode() == opComm {
			m.comm(ctx, logger, db, dispatcher, s, msg)
			continue
		}
		if msg.GetOpCode() != opMove {
			continue
		}
		uid := msg.GetUserId()
		var in struct {
			Cell int `json:"cell"`
		}
		if err := json.Unmarshal(msg.GetData(), &in); err != nil || in.Cell < 0 || in.Cell > 8 {
			m.send(dispatcher, opError, map[string]any{"code": "BAD_MOVE"}, []runtime.Presence{msg})
			continue
		}
		if uid != s.turn {
			m.send(dispatcher, opError, map[string]any{"code": "NOT_YOUR_TURN"}, []runtime.Presence{msg})
			continue
		}
		if s.board[in.Cell] != "" {
			m.send(dispatcher, opError, map[string]any{"code": "CELL_TAKEN"}, []runtime.Presence{msg})
			continue
		}
		s.board[in.Cell] = s.marks[uid]
		if winLine(s.board, s.marks[uid]) {
			m.finish(ctx, logger, dispatcher, s, tick, uid, "line")
			return s
		}
		if full(s.board) {
			m.finish(ctx, logger, dispatcher, s, tick, "", "draw")
			return s
		}
		s.turn = m.other(s, uid)
		s.turnEnd = tick + s.turnTicks
		m.broadcast(dispatcher, opUpdate, s, tick)
	}
	if tick >= s.turnEnd {
		m.finish(ctx, logger, dispatcher, s, tick, m.other(s, s.turn), "timeout")
		return s
	}
	if tick-s.startTick >= s.limitTicks {
		m.finish(ctx, logger, dispatcher, s, tick, "", "time_limit")
	}
	return s
}

func (m *tttMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	s := state.(*tttState)
	if !s.started {
		if _, err := gamecatalog.Cancel(ctx, s.matchID); err != nil {
			logger.Warn("Tic-tac-toe cancel on shutdown failed: %v", err)
		}
	} else if !s.done {
		m.finish(ctx, logger, dispatcher, s, tick, "", "server_stop")
	}
	return s
}

func (m *tttMatch) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

func (m *tttMatch) other(s *tttState, uid string) string {
	for _, u := range s.order {
		if u != uid {
			return u
		}
	}
	return ""
}

func (m *tttMatch) finish(ctx context.Context, logger runtime.Logger, dispatcher runtime.MatchDispatcher, s *tttState, tick int64, winner, reason string) {
	s.done, s.winner, s.reason, s.endTick = true, winner, reason, tick+tttEndTicks
	var winners []string
	if winner != "" {
		winners = []string{winner}
	}
	if err := commpacks.Close(ctx, s.db, s.matchID, "ended"); err != nil {
		logger.Warn("Tic-tac-toe comm channel not closed: %v", err)
	}
	payouts, err := gamecatalog.Finish(ctx, s.matchID, winners)
	if err != nil {
		logger.Warn("Tic-tac-toe finish failed: %v", err)
	}
	s.payouts = payouts
	m.broadcast(dispatcher, opResult, s, tick)
}

func (m *tttMatch) comm(ctx context.Context, logger runtime.Logger, db *sql.DB, dispatcher runtime.MatchDispatcher, s *tttState, msg runtime.MatchData) {
	me := []runtime.Presence{msg}
	if _, ok := s.marks[msg.GetUserId()]; !ok {
		m.send(dispatcher, opError, map[string]any{"code": "NOT_A_MEMBER"}, me)
		return
	}
	var in commpacks.Input
	if err := json.Unmarshal(msg.GetData(), &in); err != nil {
		m.send(dispatcher, opError, map[string]any{"code": commpacks.CodeInvalid}, me)
		return
	}
	out, err := commpacks.Send(ctx, db, s.matchID, msg.GetUserId(), in)
	if r, ok := commpacks.AsRefusal(err); ok {
		m.send(dispatcher, opError, r, me)
		return
	}
	if err != nil {
		logger.Warn("Tic-tac-toe comm message failed: %v", err)
		m.send(dispatcher, opError, map[string]any{"code": "SEND_FAILED"}, me)
		return
	}
	m.send(dispatcher, opCommMsg, out, nil)
}

func (m *tttMatch) view(s *tttState, tick int64) map[string]any {
	left := (s.turnEnd - tick) / tttTickRate
	if left < 0 {
		left = 0
	}
	v := map[string]any{"board": s.board, "marks": s.marks, "turn": s.turn, "turnLeft": left, "turnSeconds": s.turnTicks / tttTickRate, "room": s.room, "players": s.order}
	if s.done {
		v["winner"], v["reason"], v["payouts"] = s.winner, s.reason, s.payouts
	}
	return v
}

func (m *tttMatch) broadcast(dispatcher runtime.MatchDispatcher, op int64, s *tttState, tick int64) {
	m.send(dispatcher, op, m.view(s, tick), nil)
}

func (m *tttMatch) send(dispatcher runtime.MatchDispatcher, op int64, body any, to []runtime.Presence) {
	raw, _ := json.Marshal(body)
	_ = dispatcher.BroadcastMessage(op, raw, to, nil, true)
}

func winLine(b [9]string, mark string) bool {
	for _, l := range tttLines {
		if b[l[0]] == mark && b[l[1]] == mark && b[l[2]] == mark {
			return true
		}
	}
	return false
}

func full(b [9]string) bool {
	for _, c := range b {
		if c == "" {
			return false
		}
	}
	return true
}
