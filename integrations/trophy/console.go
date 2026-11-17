package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/trophy/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/jackc/pgx/v5/pgconn"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func bad(code, field string) error { return pluginkit.Bad(code, field, "") }

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	p := consolePrefix
	routes := []struct {
		path, method string
		handler      http.HandlerFunc
	}{
		{p + "/settings", http.MethodGet, s.settingsGet},
		{p + "/settings", http.MethodPut, s.settingsPut},
		{p + "/refs", http.MethodGet, s.refs},
		{p + "/leagues", http.MethodGet, s.leagueList},
		{p + "/leagues", http.MethodPost, s.leagueSave},
		{p + "/leagues/{id}", http.MethodPut, s.leagueSave},
		{p + "/leagues/{id}", http.MethodDelete, s.leagueDelete},
		{p + "/players", http.MethodGet, s.playerList},
		{p + "/players/{id}", http.MethodGet, s.playerGet},
		{p + "/players/{id}", http.MethodPut, s.playerSet},
		{p + "/prizes/retry", http.MethodPost, s.prizeRetry},
		{p + "/seasons", http.MethodGet, s.seasonList},
		{p + "/seasons/reset", http.MethodPost, s.seasonReset},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt,
		"prizes": economyhooks.PrizesAvailable()})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Normalize()
	if field, code := in.Validate(); code != "" {
		s.fail(w, bad(code, field))
		return
	}
	if err := settings.Save(r.Context(), s.db, in, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.wake()
	s.settingsGet(w, r)
}

func (s *service) refs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	out := map[string]any{}
	for name, q := range map[string]string{
		"prizes":        `SELECT key, name::text FROM economy_prize ORDER BY key LIMIT 2000`,
		"economyEvents": `SELECT key, name::text FROM economy_event_type ORDER BY key LIMIT 2000`,
		"games":         `SELECT key, name::text FROM economy_game ORDER BY key LIMIT 500`,
	} {
		list := []map[string]string{}
		rows, err := s.db.QueryContext(ctx, q)
		if err == nil {
			for rows.Next() {
				var key, raw string
				if rows.Scan(&key, &raw) == nil {
					name := key
					if langs != nil {
						if v := langs.Resolve(localized.Parse(raw), lang); v != "" {
							name = v
						}
					}
					list = append(list, map[string]string{"key": key, "name": name})
				}
			}
			rows.Close()
		}
		out[name] = list
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) cleanText(ctx context.Context, field string, in map[string]string, required bool, max int) (map[string]string, error) {
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out, err := langs.Clean(field, localized.Text(in), required, max)
	if err != nil {
		var le *localized.Error
		if errors.As(err, &le) {
			return nil, bad(le.Code, field)
		}
		return nil, err
	}
	return out, nil
}

func (s *service) leagueList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.leagues(ctx)
	if err != nil {
		s.fail(w, err)
		return
	}
	counts := map[string]int{}
	rows, err := s.db.QueryContext(ctx, `SELECT league_key, count(*) FROM trophy_player GROUP BY 1`)
	if err == nil {
		for rows.Next() {
			var k string
			var n int
			if rows.Scan(&k, &n) == nil {
				counts[k] = n
			}
		}
		rows.Close()
	}
	items := make([]map[string]any, 0, len(list))
	for _, l := range list {
		raw, _ := json.Marshal(l)
		m := map[string]any{}
		_ = json.Unmarshal(raw, &m)
		m["players"] = counts[l.Key]
		items = append(items, m)
	}
	var total int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM trophy_player`).Scan(&total)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "players": total, "unranked": counts[""]})
}

func (s *service) leagueSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in league
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Key = strings.TrimSpace(in.Key)
	if !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	var err error
	if in.Name, err = s.cleanText(ctx, "name", in.Name, true, 80); err != nil {
		s.fail(w, err)
		return
	}
	if in.Description, err = s.cleanText(ctx, "description", in.Description, false, 500); err != nil {
		s.fail(w, err)
		return
	}
	if in.MinTrophies < 0 || in.MinTrophies > 10000000 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "minTrophies"))
		return
	}
	if in.ResetTo != nil && (*in.ResetTo < 0 || *in.ResetTo > 10000000) {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "resetTo"))
		return
	}
	if len(in.IconURL) > 500 || len(in.Color) > 32 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "iconUrl"))
		return
	}
	in.PrizeKey = strings.TrimSpace(in.PrizeKey)
	if in.PrizeKey != "" {
		var ok bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE key=$1)`, in.PrizeKey).Scan(&ok)
		if !ok {
			s.fail(w, bad("PRIZE_NOT_FOUND", "prizeKey"))
			return
		}
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	name, _ := json.Marshal(in.Name)
	desc, _ := json.Marshal(in.Description)
	meta, _ := json.Marshal(in.Metadata)
	by := pluginkit.ConsoleUser(r)
	var res sql.Result
	if id == "" {
		res, err = s.db.ExecContext(ctx, `INSERT INTO trophy_league (id, key, name, description, icon_url, color, min_trophies, floor, reset_to, prize_key, is_active, metadata, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, uuid.Must(uuid.NewV4()).String(), in.Key, name, desc, in.IconURL, in.Color, in.MinTrophies, in.Floor, in.ResetTo, in.PrizeKey, in.Active, meta, by)
	} else {
		res, err = s.db.ExecContext(ctx, `UPDATE trophy_league SET key=$2, name=$3, description=$4, icon_url=$5, color=$6, min_trophies=$7, floor=$8, reset_to=$9, prize_key=$10, is_active=$11,
			metadata=$12, updated_by=$13, updated_at=now() WHERE id::text=$1`, id, in.Key, name, desc, in.IconURL, in.Color, in.MinTrophies, in.Floor, in.ResetTo, in.PrizeKey, in.Active, meta, by)
	}
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			field := "key"
			if strings.Contains(pg.ConstraintName, "min") {
				field = "minTrophies"
			}
			s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "TROPHY_LEAGUE_EXISTS", Field: field})
			return
		}
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("TROPHY_LEAGUE_NOT_FOUND"))
		return
	}
	s.invalidate()
	s.leagueList(w, r)
}

func (s *service) leagueDelete(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM trophy_league WHERE id::text=$1`, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("TROPHY_LEAGUE_NOT_FOUND"))
		return
	}
	s.invalidate()
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) playerList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 25, 100)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	lk := strings.TrimSpace(r.URL.Query().Get("league"))
	args := &pluginkit.Args{}
	where := []string{"true"}
	if q != "" {
		if uuidPattern.MatchString(q) {
			where = append(where, "p.user_id::text = "+args.Add(strings.ToLower(q)))
		} else {
			where = append(where, "u.username ILIKE "+args.Add(pluginkit.LikePattern(q)))
		}
	}
	if r.URL.Query().Has("league") {
		where = append(where, "p.league_key = "+args.Add(lk))
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM trophy_player p LEFT JOIN users u ON u.id=p.user_id WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.user_id::text, COALESCE(u.username,''), COALESCE(u.display_name,''), p.trophies, p.best, p.season_best, p.league_key, p.wins, p.losses, p.draws, p.updated_at
		FROM trophy_player p LEFT JOIN users u ON u.id=p.user_id WHERE `+cond+` ORDER BY p.trophies DESC, p.user_id LIMIT `+args.Add(limit)+` OFFSET `+args.Add((page-1)*limit), args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, uname, dname, lk string
		var t, best, sb, wins, losses, draws int
		var at time.Time
		if err := rows.Scan(&id, &uname, &dname, &t, &best, &sb, &lk, &wins, &losses, &draws, &at); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, map[string]any{"userId": id, "username": uname, "displayName": dname, "trophies": t, "best": best, "seasonBest": sb, "league": lk,
			"wins": wins, "losses": losses, "draws": draws, "updatedAt": at.UTC().Format(time.RFC3339)})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) playerGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if !uuidPattern.MatchString(id) {
		s.fail(w, bad("INVALID_USER_ID", "id"))
		return
	}
	cfg := s.settings(ctx)
	st := s.standing(ctx, id, cfg.Amounts.Start)
	l := s.ladder(ctx)
	lk := ""
	if cur := l.at(st.Trophies); cur != nil {
		lk = cur.Key
	}
	logs, err := s.logRows(ctx, id, 50, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	prizes := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT claim_key, league_key, season, prize_key, status, tries, last_error, created_at, paid_at FROM trophy_prize WHERE user_id::text=$1 ORDER BY created_at DESC`, id)
	if err == nil {
		for rows.Next() {
			var ck, lk2, pk, status, lastErr string
			var season, tries int
			var at time.Time
			var paid *time.Time
			if rows.Scan(&ck, &lk2, &season, &pk, &status, &tries, &lastErr, &at, &paid) == nil {
				var pa any
				if paid != nil {
					pa = paid.UTC().Format(time.RFC3339)
				}
				prizes = append(prizes, map[string]any{"claimKey": ck, "league": lk2, "season": season, "prizeKey": pk, "status": status, "tries": tries, "error": lastErr,
					"createdAt": at.UTC().Format(time.RFC3339), "paidAt": pa})
			}
		}
		rows.Close()
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"userId": id, "trophies": st.Trophies, "best": st.Best, "seasonBest": st.Season, "wins": st.Wins, "losses": st.Losses,
		"draws": st.Draws, "league": lk, "log": logs, "prizes": prizes})
}

func (s *service) playerSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if !uuidPattern.MatchString(id) {
		s.fail(w, bad("INVALID_USER_ID", "id"))
		return
	}
	var in struct {
		Trophies *int   `json:"trophies"`
		Mode     string `json:"mode"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if in.Mode == "" {
		in.Mode = "set"
	}
	if in.Trophies == nil || (in.Mode != "set" && in.Mode != "add") || (in.Mode == "set" && (*in.Trophies < 0 || *in.Trophies > 10000000)) || *in.Trophies > 10000000 || *in.Trophies < -10000000 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "trophies"))
		return
	}
	var exists bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id::text=$1)`, id).Scan(&exists)
	if !exists {
		s.fail(w, pluginkit.Missing("USER_NOT_FOUND"))
		return
	}
	by := pluginkit.ConsoleUser(r)
	ch, err := s.adminWrite(ctx, id, *in.Trophies, in.Mode, by)
	if err != nil {
		s.fail(w, err)
		return
	}
	_ = activity.Record(ctx, s.db, activity.Entry{UserID: id, Kind: activity.KindEvent, Action: "trophy_set", Actor: by, Data: map[string]any{"delta": ch.Delta, "trophies": ch.Trophies}})
	s.playerGet(w, r)
}

func (s *service) prizeRetry(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `UPDATE trophy_prize SET status='pending', tries=0 WHERE status='failed'`)
	if err != nil {
		s.fail(w, err)
		return
	}
	n, _ := res.RowsAffected()
	s.payPending(r.Context(), "", 200)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"retried": n})
}

func (s *service) seasonList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.db.QueryContext(ctx, `SELECT number, started_at, ends_at, ended_at, players_reset, mode, actor FROM trophy_season ORDER BY number DESC LIMIT 100`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []seasonRow{}
	for rows.Next() {
		var x seasonRow
		if err := rows.Scan(&x.Number, &x.StartedAt, &x.EndsAt, &x.EndedAt, &x.Reset, &x.Mode, &x.Actor); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, x)
	}
	cfg := s.settings(ctx)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "enabled": cfg.Season.Enabled, "schedule": cfg.Season.Schedule, "mode": cfg.Season.Mode,
		"nextRun": nextRun(cfg.Season.Schedule, s.now())})
}

func (s *service) seasonReset(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Confirm bool `json:"confirm"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if !in.Confirm {
		s.fail(w, bad("CONFIRM_REQUIRED", "confirm"))
		return
	}
	if err := s.seasonTick(r.Context(), pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.seasonList(w, r)
}
