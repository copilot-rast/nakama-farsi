package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/friendroom"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

var (
	queueLocks   sync.Map
	queueLobbies = struct {
		sync.Mutex
		ids map[string][]string
	}{ids: map[string][]string{}}
)

func queueLock(key string) *sync.Mutex {
	m, _ := queueLocks.LoadOrStore(key, &sync.Mutex{})
	return m.(*sync.Mutex)
}

func recentLobbies(key string) []string {
	queueLobbies.Lock()
	defer queueLobbies.Unlock()
	return append([]string{}, queueLobbies.ids[key]...)
}

func rememberLobby(key, id string, keep bool) {
	queueLobbies.Lock()
	defer queueLobbies.Unlock()
	list := []string{}
	for _, x := range queueLobbies.ids[key] {
		if x != id {
			list = append(list, x)
		}
	}
	if keep {
		list = append(list, id)
	}
	if len(list) > 50 {
		list = list[len(list)-50:]
	}
	queueLobbies.ids[key] = list
}

func (s *service) gameLiveRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_game_queue":  s.rpcGameQueue,
		"economy_game_active": s.rpcGameActive,
		"economy_player_card": s.rpcPlayerCard,
		"economy_game_stats":  s.rpcGameStats,
	}
}

func (s *service) gameLiveConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/game-matches/{id}/end", http.MethodPost, s.consoleGameMatchEnd},
		{consolePrefix + "/game-matches/{id}/live", http.MethodGet, s.consoleGameMatchLive},
		{consolePrefix + "/game-matches/{id}/actions", http.MethodGet, s.consoleGameMatchActions},
		{consolePrefix + "/game-stats", http.MethodGet, s.consoleGameStats},
		{consolePrefix + "/game-stats/{userId}/{game}", http.MethodPut, s.consoleGameStatsSet},
	}
}

func (s *service) rpcGameQueue(ctx context.Context, userID, payload string) (any, error) {
	var in gameRoomRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	g, t, err := s.realtimeGame(ctx, userID, in, gamecatalog.PathMatchmaking)
	if err != nil {
		return nil, err
	}
	room, err := g.liveRoom(t.Room)
	if err != nil {
		return nil, err
	}
	pub := g.public()
	out := map[string]any{"key": g.Key, "room": room.Key, "players": room.Seats(pub), "humanSeats": room.Humans(pub), "botFillSeconds": room.BotFillSeconds}
	if g.stealth().Enabled {
		delete(out, "humanSeats")
		delete(out, "botFillSeconds")
	}
	reserve := friendroom.Signal{Action: friendroom.ActionReserve, UserID: userID}.Data()
	try := func(id string) bool {
		ans, err := s.nk.MatchSignal(ctx, id, reserve)
		if err != nil {
			return false
		}
		ok, _ := friendroom.ParseAnswer(ans)
		return ok
	}
	lockKey := g.Key + "/" + room.Key
	lock := queueLock(lockKey)
	lock.Lock()
	defer lock.Unlock()
	for _, id := range recentLobbies(lockKey) {
		if try(id) {
			out["matchId"], out["created"] = id, false
			return out, nil
		}
		rememberLobby(lockKey, id, false)
	}
	query := "+label.game:" + g.Key + " +label.mode:" + friendroom.ModeMatchmaking + " +label.status:" + friendroom.StatusWaiting + " +label.free:>=1"
	if room.Key != "" {
		query += " +label.room:" + room.Key
	}
	min := 0
	if found, err := s.nk.MatchList(ctx, 20, true, "", &min, nil, query); err == nil {
		for _, m := range found {
			if try(m.GetMatchId()) {
				out["matchId"], out["created"] = m.GetMatchId(), false
				return out, nil
			}
		}
	}
	p := friendroom.Params{Game: g.Key, Room: room.Key, Mode: friendroom.ModeMatchmaking}
	id, err := s.nk.MatchCreate(ctx, g.Key, p.Map())
	if err != nil {
		s.logger.Error("Economy queue match for %s failed: %v", g.Key, err)
		return nil, fail(http.StatusServiceUnavailable, "MATCH_CREATE_FAILED", "", "the match could not be made, try again")
	}
	if !try(id) {
		return nil, fail(http.StatusServiceUnavailable, "MATCH_CREATE_FAILED", "", "the match could not be made, try again")
	}
	rememberLobby(lockKey, id, true)
	out["matchId"], out["created"] = id, true
	return out, nil
}

func (s *service) rpcGameActive(ctx context.Context, userID, payload string) (any, error) {
	var in struct{}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	a, err := s.gameActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"match": a}, nil
}

type matchRow struct {
	status, game string
}

func (s *service) matchStatus(ctx context.Context, id string) (matchRow, error) {
	var m matchRow
	err := s.db.QueryRowContext(ctx, `SELECT status, game_key FROM economy_game_match WHERE id=$1`, id).Scan(&m.status, &m.game)
	if errors.Is(err, sql.ErrNoRows) {
		return m, notFound("MATCH_NOT_FOUND")
	}
	return m, err
}

func (s *service) liveMatch(ctx context.Context, id string) bool {
	if s.nk == nil || !strings.Contains(id, ".") {
		return false
	}
	m, err := s.nk.MatchGet(ctx, id)
	return err == nil && m != nil
}

func (s *service) signalMatch(ctx context.Context, id string, sig friendroom.Signal, refused string) error {
	ans, err := s.nk.MatchSignal(ctx, id, sig.Data())
	if err != nil {
		return notFound("MATCH_NOT_LIVE")
	}
	ok, reason := friendroom.ParseAnswer(ans)
	if !ok {
		e := conflict(refused, "the match refused the action")
		if reason != "" {
			e.Meta = map[string]any{"reason": reason}
		}
		return e
	}
	return nil
}

func (s *service) consoleGameMatchCancel(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Reason string `json:"reason"`
	}
	if r.ContentLength > 0 {
		if err := pluginkit.ReadBody(r, &in, 4096); err != nil {
			s.writeError(w, err)
			return
		}
	}
	m, err := s.matchStatus(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	staff := pluginkit.ConsoleUser(r)
	reason := strings.TrimSpace(in.Reason)
	if m.status == "finished" {
		s.writeError(w, conflict("MATCH_FINISHED", "this match has already finished"))
		return
	}
	if m.status != "cancelled" && s.liveMatch(r.Context(), id) {
		if err := s.signalMatch(r.Context(), id, friendroom.Signal{Action: friendroom.ActionAdminCancel, UserID: staff, Reason: reason}, "NOT_CANCELLABLE"); err != nil {
			s.writeError(w, err)
			return
		}
		var refunded int
		_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM economy_game_entry WHERE match_id=$1 AND status='refunded'`, id).Scan(&refunded)
		writeJSON(w, http.StatusOK, map[string]any{"refunded": refunded, "via": "match"})
		return
	}
	n, err := s.cancelMatch(r.Context(), id, true, "admin-cancel", staff)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"refunded": n, "via": "server"})
}

func (s *service) consoleGameMatchEnd(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		WinnerUserID string `json:"winnerUserId"`
		Reason       string `json:"reason"`
	}
	if err := pluginkit.ReadBody(r, &in, 4096); err != nil {
		s.writeError(w, err)
		return
	}
	winner := strings.TrimSpace(in.WinnerUserID)
	if _, err := uuid.FromString(winner); err != nil {
		s.writeError(w, badRequest("WINNER_REQUIRED", "winnerUserId", "pick the winner"))
		return
	}
	m, err := s.matchStatus(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	switch m.status {
	case "finished":
		s.writeError(w, conflict("MATCH_FINISHED", "this match has already finished"))
		return
	case "cancelled":
		s.writeError(w, conflict("MATCH_CANCELLED", "this match was cancelled"))
		return
	case "open":
		s.writeError(w, conflict("MATCH_NOT_STARTED", "a waiting match has nothing to settle; cancel it instead"))
		return
	}
	var seated bool
	_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM economy_game_entry WHERE match_id=$1 AND user_id::text=$2 AND status IN ('paid','free'))`, id, winner).Scan(&seated)
	if !seated {
		s.writeError(w, badRequest("WINNER_NOT_SEATED", "winnerUserId", "this player is not in the match"))
		return
	}
	staff := pluginkit.ConsoleUser(r)
	reason := strings.TrimSpace(in.Reason)
	if s.liveMatch(r.Context(), id) {
		if err := s.signalMatch(r.Context(), id, friendroom.Signal{Action: friendroom.ActionAdminEnd, UserID: staff, Winner: winner, Reason: reason}, "NOT_ENDABLE"); err != nil {
			s.writeError(w, err)
			return
		}
		payouts, _ := s.settledPayouts(r.Context(), id)
		writeJSON(w, http.StatusOK, map[string]any{"via": "match", "payouts": payouts})
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT user_id::text FROM economy_game_entry WHERE match_id=$1 AND status IN ('paid','free') ORDER BY created_at, user_id`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	results := []gamecatalog.Result{{UserID: winner, Rank: 1, Status: gamecatalog.ResultWon}}
	rank := 1
	for rows.Next() {
		var u string
		if rows.Scan(&u) == nil && u != winner {
			rank++
			results = append(results, gamecatalog.Result{UserID: u, Rank: rank, Status: gamecatalog.ResultLost})
		}
	}
	rows.Close()
	payouts, err := s.gameSettle(r.Context(), gamecatalog.Settlement{MatchID: id, Results: results, Reason: "admin-force-end", ForcedBy: staff, Note: reason})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"via": "server", "payouts": payouts})
}

func (s *service) consoleGameMatchLive(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if !s.liveMatch(r.Context(), id) {
		s.writeError(w, notFound("MATCH_NOT_LIVE"))
		return
	}
	ans, err := s.nk.MatchSignal(r.Context(), id, friendroom.Signal{Action: friendroom.ActionAdminView, UserID: pluginkit.ConsoleUser(r)}.Data())
	if err != nil {
		s.writeError(w, notFound("MATCH_NOT_LIVE"))
		return
	}
	if !json.Valid([]byte(ans)) {
		s.writeError(w, conflict("NO_LIVE_VIEW", "this game does not share its live state"))
		return
	}
	writeJSON(w, http.StatusOK, json.RawMessage(ans))
}

func (s *service) consoleGameMatchActions(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	q := r.URL.Query()
	after, _ := strconv.Atoi(q.Get("after"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('seq',a.seq,'userId',a.user_id,'username',u.username,'kind',a.kind,'data',a.data,'at',a.at)::text
		FROM economy_game_action a LEFT JOIN users u ON u.id=a.user_id WHERE a.match_id=$1 AND a.seq > $2 ORDER BY a.seq LIMIT $3`, id, after, limit)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			s.writeError(w, err)
			return
		}
		items = append(items, json.RawMessage(raw))
	}
	var total int
	_ = s.db.QueryRowContext(r.Context(), `SELECT actions FROM economy_game_match WHERE id=$1`, id).Scan(&total)
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}
