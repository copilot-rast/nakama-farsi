package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"hash/fnv"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/heroiclabs/nakama/v3/internal/stealthbots"
)

const botLeaseStale = 30 * time.Minute

func (g *gameRecord) stealth() stealthbots.Settings {
	if g == nil {
		return stealthbots.Defaults()
	}
	return stealthbots.Parse(g.Stealth)
}

func (s *service) stealthOf(ctx context.Context, gameKey string) stealthbots.Settings {
	g, err := s.gameByKey(ctx, gameKey)
	if err != nil {
		return stealthbots.Defaults()
	}
	return g.stealth()
}

func (s *service) stealthConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/game-stealth", http.MethodGet, s.consoleStealthList},
		{consolePrefix + "/game-stealth/{key}", http.MethodGet, s.consoleStealthGet},
		{consolePrefix + "/game-stealth/{key}", http.MethodPut, s.consoleStealthPut},
	}
}

func (s *service) stealthView(ctx context.Context, key string) (map[string]any, error) {
	var raw, name string
	err := s.db.QueryRowContext(ctx, `SELECT stealth::text, name::text FROM economy_game WHERE key=$1`, key).Scan(&raw, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("GAME_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	st := stealthbots.Parse([]byte(raw))
	var busy, online, pool int
	_ = s.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM economy_bot WHERE enabled), (SELECT count(*) FROM economy_bot_lease WHERE busy AND game_key=$1 AND leased_at > now() - interval '30 minutes'),
		(SELECT count(*) FROM economy_bot_lease WHERE NOT busy AND online_until > now())`, key).Scan(&pool, &busy, &online)
	var guarded int
	if st.Guard {
		_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM economy_bot_guard WHERE game_key=$1 AND day=current_date AND wins >= $2`, key, st.GuardDailyWins).Scan(&guarded)
	}
	return map[string]any{"key": key, "name": json.RawMessage(name), "settings": st, "defaults": stealthbots.Defaults(), "saved": strings.TrimSpace(raw) != "{}",
		"pool": map[string]any{"enabled": pool, "inMatch": busy, "onlineAfterMatch": online, "guardedPlayersToday": guarded}}, nil
}

func (s *service) consoleStealthList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT key, name::text, stealth::text FROM economy_game ORDER BY sort, key`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var key, name, raw string
		if err := rows.Scan(&key, &name, &raw); err != nil {
			s.writeError(w, err)
			return
		}
		st := stealthbots.Parse([]byte(raw))
		items = append(items, map[string]any{"key": key, "name": json.RawMessage(name), "enabled": st.Enabled, "guard": st.Guard})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *service) consoleStealthGet(w http.ResponseWriter, r *http.Request) {
	v, err := s.stealthView(r.Context(), pathVar(r, "key"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *service) consoleStealthPut(w http.ResponseWriter, r *http.Request) {
	key := pathVar(r, "key")
	st := stealthbots.Defaults()
	if err := pluginkit.ReadBody(r, &st, 8192); err != nil {
		s.writeError(w, err)
		return
	}
	st = st.Clean()
	raw, _ := json.Marshal(st)
	res, err := s.db.ExecContext(r.Context(), `UPDATE economy_game SET stealth=$2::jsonb, updated_at=now() WHERE key=$1`, key, string(raw))
	s.logger.Info("Economy game %s stealth bots set by %s: enabled=%v", key, pluginkit.ConsoleUser(r), st.Enabled)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("GAME_NOT_FOUND"))
		return
	}
	s.games.mu.Lock()
	s.games.checked = time.Time{}
	s.games.mu.Unlock()
	stealthbots.Reset()
	v, err := s.stealthView(r.Context(), key)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *service) scrubGameView(g *gameRecord, view map[string]any) {
	if settings, ok := view["settings"].(map[string]any); ok {
		view["settings"] = stealthbots.WithoutStealth(settings)
	}
	if !g.stealth().Enabled {
		return
	}
	raw, err := json.Marshal(view)
	if err != nil {
		return
	}
	clean := map[string]any{}
	if json.Unmarshal(raw, &clean) != nil {
		return
	}
	stealthbots.Scrub(clean)
	for k := range view {
		delete(view, k)
	}
	for k, v := range clean {
		view[k] = v
	}
}

func botLevelWindows(q gamecatalog.BotQuery) [][2]int {
	if q.Level <= 0 || q.LevelRange <= 0 {
		return [][2]int{{q.MinLevel, 1000}}
	}
	out := [][2]int{}
	for _, r := range []int{q.LevelRange, q.LevelRange * 2, q.LevelRange * 4} {
		lo := q.Level - r
		if lo < q.MinLevel {
			lo = q.MinLevel
		}
		out = append(out, [2]int{lo, q.Level + r})
	}
	return append(out, [2]int{q.MinLevel, 1000})
}

func (s *service) leaseBots(ctx context.Context, q gamecatalog.BotQuery) ([]gamecatalog.Bot, error) {
	exclude := []string{}
	for _, id := range q.Exclude {
		if _, err := uuid.FromString(id); err == nil {
			exclude = append(exclude, id)
		}
	}
	cur, amount := "", int64(0)
	if q.Fee != nil && q.Fee.Amount > 0 {
		cur, amount = q.Fee.CurrencyKey, q.Fee.Amount
	}
	out := []gamecatalog.Bot{}
	taken := map[string]bool{}
	st := s.stealthOf(ctx, q.Game)
	wins := botLevelWindows(q)
	nearFirst := st.GrowNearLevel && len(wins) > 1
	grow := func() {
		for _, b := range s.growBots(ctx, q, st, q.Count-len(out)) {
			taken[b.UserID] = true
			out = append(out, b)
		}
	}
	for i, win := range wins {
		if len(out) >= q.Count {
			break
		}
		if nearFirst && i == len(wins)-1 {
			grow()
			if len(out) >= q.Count {
				break
			}
		}
		rows, err := s.db.QueryContext(ctx, `SELECT b.user_id::text, u.username, COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), b.level, b.template, b.look::text
			FROM economy_bot b JOIN users u ON u.id=b.user_id LEFT JOIN economy_bot_lease l ON l.user_id=b.user_id
			WHERE b.enabled AND u.disable_time <= '1970-01-01 00:00:01+00' AND b.level BETWEEN $1 AND $8 AND NOT (b.user_id::text = ANY($2::text[]))
			AND (b.rooms = '[]'::jsonb OR b.rooms @> jsonb_build_array(jsonb_build_object('game',$3::text,'room',$4::text)) OR b.rooms @> jsonb_build_array(jsonb_build_object('game',$3::text,'room','')))
			AND ($5 = '' OR COALESCE((u.wallet->>$5)::bigint, 0) >= $6)
			AND (l.user_id IS NULL OR NOT l.busy OR l.leased_at < now() - interval '30 minutes')
			AND NOT EXISTS (SELECT 1 FROM economy_game_entry e JOIN economy_game_match m ON m.id=e.match_id WHERE e.user_id=b.user_id AND e.bot AND e.status IN ('paid','free')
				AND m.status IN ('open','started') AND m.id <> $9 AND m.created_at > now() - interval '2 hours')
			ORDER BY random() LIMIT $7`, win[0], exclude, q.Game, q.Room, cur, amount, q.Count+10, win[1], q.MatchID)
		if err != nil {
			return nil, err
		}
		cands := []gamecatalog.Bot{}
		for rows.Next() {
			var b gamecatalog.Bot
			var look string
			if err := rows.Scan(&b.UserID, &b.Username, &b.Name, &b.Avatar, &b.Level, &b.Template, &look); err != nil {
				rows.Close()
				return nil, err
			}
			b.Look = map[string]string{}
			_ = json.Unmarshal([]byte(look), &b.Look)
			if b.Name == "" {
				b.Name = b.Username
			}
			cands = append(cands, b)
		}
		rows.Close()
		for _, b := range cands {
			if len(out) >= q.Count {
				break
			}
			if taken[b.UserID] {
				continue
			}
			var got string
			err := s.db.QueryRowContext(ctx, `INSERT INTO economy_bot_lease AS l (user_id, match_id, game_key, busy, leased_at) VALUES ($1::uuid,$2,$3,true,now())
				ON CONFLICT (user_id) DO UPDATE SET match_id=EXCLUDED.match_id, game_key=EXCLUDED.game_key, busy=true, leased_at=now()
				WHERE NOT l.busy OR l.leased_at < now() - interval '30 minutes' OR l.match_id = EXCLUDED.match_id RETURNING user_id::text`, b.UserID, q.MatchID, q.Game).Scan(&got)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return nil, err
			}
			taken[b.UserID] = true
			out = append(out, b)
		}
	}
	if !nearFirst && len(out) < q.Count {
		grow()
	}
	return out, nil
}

func (s *service) releaseBotLeases(ctx context.Context, matchID, gameKey string) {
	if matchID == "" {
		return
	}
	st := s.stealthOf(ctx, gameKey)
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text FROM economy_bot_lease WHERE match_id=$1 AND busy`, matchID)
	if err != nil {
		return
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		after := st.OnlineAfter()
		if !st.Enabled {
			after = 0
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE economy_bot_lease SET busy=false, online_until=now() + $3 * interval '1 second' WHERE user_id=$1::uuid AND match_id=$2`, id, matchID, int(after.Seconds())); err != nil {
			s.logger.Warn("Economy bot lease %s release failed: %v", id, err)
		}
	}
}

func (s *service) matchHasBot(ctx context.Context, matchID string) bool {
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_game_entry WHERE match_id=$1 AND bot AND status IN ('paid','free'))
		OR EXISTS (SELECT 1 FROM economy_bot_lease WHERE match_id=$1 AND match_id <> '')`, matchID).Scan(&ok)
	return ok
}

func (s *service) stealthAfterMatch(ctx context.Context, matchID, gameKey string, humanWinners []string) {
	st := s.stealthOf(ctx, gameKey)
	if st.Enabled && st.Guard && len(humanWinners) > 0 && s.matchHasBot(ctx, matchID) {
		for _, w := range humanWinners {
			if err := stealthbots.GuardAdd(ctx, s.db, w, gameKey); err != nil {
				s.logger.Warn("Economy bot guard for %s failed: %v", w, err)
			}
		}
	}
	s.releaseBotLeases(ctx, matchID, gameKey)
}

func botXPFraction(userID string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(userID))
	return 0.08 + float64(h.Sum32()%840)/1000
}

func (p gameProvider) ReleaseBots(ctx context.Context, matchID string) error {
	var game string
	_ = p.s.db.QueryRowContext(ctx, `SELECT game_key FROM economy_bot_lease WHERE match_id=$1 LIMIT 1`, matchID).Scan(&game)
	p.s.releaseBotLeases(ctx, matchID, game)
	return nil
}

func (s *service) stealthPool(ctx context.Context) bool {
	return stealthbots.Active(ctx, s.db)
}
