package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type texts struct {
	langs *localized.Set
	lang  string
}

func (s *service) texts(ctx context.Context, userID, lang string) texts {
	langs, l, _ := localized.ForUser(ctx, s.db, userID, lang)
	return texts{langs: langs, lang: l}
}

func (t texts) of(m map[string]string, fallback string) string {
	if t.langs != nil {
		if v := t.langs.Resolve(localized.Text(m), t.lang); v != "" {
			return v
		}
	}
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return fallback
}

type standing struct {
	Trophies int
	Best     int
	Season   int
	Wins     int
	Losses   int
	Draws    int
}

func (s *service) standing(ctx context.Context, userID string, start int) standing {
	st := standing{Trophies: start, Best: start, Season: start}
	_ = s.db.QueryRowContext(ctx, `SELECT trophies, best, season_best, wins, losses, draws FROM trophy_player WHERE user_id::text=$1`, userID).
		Scan(&st.Trophies, &st.Best, &st.Season, &st.Wins, &st.Losses, &st.Draws)
	return st
}

func leagueView(l *league, tx texts) map[string]any {
	if l == nil {
		return nil
	}
	return map[string]any{"key": l.Key, "name": tx.of(l.Name, l.Key), "iconUrl": l.IconURL, "color": l.Color, "minTrophies": l.MinTrophies}
}

func (s *service) seasonView(ctx context.Context) map[string]any {
	cur, err := s.currentSeason(ctx, s.db)
	if err != nil || cur == nil {
		return nil
	}
	var ends any
	if cur.EndsAt != nil {
		ends = cur.EndsAt.UTC().Format(time.RFC3339)
	}
	return map[string]any{"number": cur.Number, "startedAt": cur.StartedAt.UTC().Format(time.RFC3339), "endsAt": ends}
}

func (s *service) rpcMe(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	tx := s.texts(ctx, userID, in.Lang)
	l := s.ladder(ctx)
	st := s.standing(ctx, userID, cfg.Amounts.Start)
	next := l.next(st.Trophies)
	var nextView map[string]any
	if next != nil {
		nextView = map[string]any{"key": next.Key, "name": tx.of(next.Name, next.Key), "iconUrl": next.IconURL, "color": next.Color, "minTrophies": next.MinTrophies,
			"need": next.MinTrophies - st.Trophies}
	}
	var pending int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM trophy_prize WHERE user_id::text=$1 AND status='pending'`, userID).Scan(&pending)
	return map[string]any{"enabled": cfg.Enabled, "trophies": st.Trophies, "best": st.Best, "seasonBest": st.Season, "wins": st.Wins, "losses": st.Losses, "draws": st.Draws,
		"league": leagueView(l.at(st.Trophies), tx), "nextLeague": nextView, "season": s.seasonView(ctx), "pendingPrizes": pending,
		"amounts": map[string]any{"win": cfg.Amounts.Win, "loss": cfg.Amounts.Loss, "draw": cfg.Amounts.Draw, "scaled": cfg.Amounts.Scale}}, nil
}

func (s *service) rpcLeagues(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	tx := s.texts(ctx, userID, in.Lang)
	l := s.ladder(ctx)
	st := s.standing(ctx, userID, cfg.Amounts.Start)
	claims := map[string]string{}
	rows, err := s.db.QueryContext(ctx, `SELECT league_key, status FROM trophy_prize WHERE user_id::text=$1`, userID)
	if err == nil {
		for rows.Next() {
			var k, st string
			if rows.Scan(&k, &st) == nil {
				claims[k] = st
			}
		}
		rows.Close()
	}
	cur := l.index(st.Trophies)
	out := make([]map[string]any, 0, len(l))
	for i, x := range l {
		var maxT any
		if i+1 < len(l) {
			maxT = l[i+1].MinTrophies - 1
		}
		var prize any
		if x.PrizeKey != "" {
			status := claims[x.Key]
			if status == "" {
				status = "locked"
			}
			prize = map[string]any{"prizeKey": x.PrizeKey, "status": status}
		}
		out = append(out, map[string]any{"key": x.Key, "name": tx.of(x.Name, x.Key), "description": tx.of(x.Description, ""), "iconUrl": x.IconURL, "color": x.Color,
			"minTrophies": x.MinTrophies, "maxTrophies": maxT, "floor": x.Floor, "prize": prize, "reached": st.Best >= x.MinTrophies, "current": i == cur, "metadata": x.Metadata})
	}
	return map[string]any{"leagues": out, "trophies": st.Trophies, "best": st.Best, "season": s.seasonView(ctx)}, nil
}

func (s *service) rpcHistory(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit int `json:"limit"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	limit := in.Limit
	if limit <= 0 || limit > cfg.Display.HistoryLimit {
		limit = cfg.Display.HistoryLimit
	}
	items, err := s.logRows(ctx, userID, limit, cfg.Display.ShowOpponent)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items}, nil
}

func (s *service) logRows(ctx context.Context, userID string, limit int, opponent bool) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.id, t.reason, t.delta, t.before, t.after, t.match_id, t.opponent_id, t.opponent_trophies, t.league_before, t.league_after, t.actor, t.created_at,
		COALESCE(u.username,''), COALESCE(u.display_name,'')
		FROM trophy_log t LEFT JOIN users u ON t.opponent_id <> '' AND u.id::text = t.opponent_id WHERE t.user_id::text=$1 ORDER BY t.id DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var reason, matchID, oppID, lb, la, actor, uname, dname string
		var delta, before, after int
		var oppT sql.NullInt64
		var at time.Time
		if err := rows.Scan(&id, &reason, &delta, &before, &after, &matchID, &oppID, &oppT, &lb, &la, &actor, &at, &uname, &dname); err != nil {
			return nil, err
		}
		row := map[string]any{"id": id, "reason": reason, "delta": delta, "before": before, "after": after, "matchId": matchID, "leagueBefore": lb, "leagueAfter": la,
			"createdAt": at.UTC().Format(time.RFC3339), "opponent": nil}
		if opponent && oppID != "" {
			var t any
			if oppT.Valid {
				t = oppT.Int64
			}
			row["opponent"] = map[string]any{"userId": oppID, "username": uname, "displayName": dname, "trophies": t}
		}
		if actor != "" {
			row["actor"] = actor
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *service) rpcPlayers(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserIDs []string `json:"userIds"`
		Lang    string   `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	ids := []string{}
	seen := map[string]bool{}
	for _, id := range in.UserIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		if !uuidPattern.MatchString(id) {
			return nil, apierr.BadRequest("INVALID_USER_ID").WithField("userIds")
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, apierr.BadRequest("TROPHY_USERS_REQUIRED").WithField("userIds")
	}
	if len(ids) > 100 {
		return nil, apierr.BadRequest("TROPHY_TOO_MANY_USERS").WithField("userIds").With("max", 100)
	}
	cfg := s.settings(ctx)
	tx := s.texts(ctx, userID, in.Lang)
	l := s.ladder(ctx)
	found := map[string]standing{}
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text, trophies, best FROM trophy_player WHERE user_id::text = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var st standing
		if rows.Scan(&id, &st.Trophies, &st.Best) == nil {
			found[id] = st
		}
	}
	out := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		st, ok := found[id]
		if !ok {
			st = standing{Trophies: cfg.Amounts.Start, Best: cfg.Amounts.Start}
		}
		out = append(out, map[string]any{"userId": id, "trophies": st.Trophies, "best": st.Best, "league": leagueView(l.at(st.Trophies), tx)})
	}
	return map[string]any{"players": out}, nil
}

