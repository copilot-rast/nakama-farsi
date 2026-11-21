package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/events"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/store"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) queryFilter(r *http.Request, cfg settings.All, def int) (store.Filter, span, error) {
	sp, err := s.span(r, cfg, def)
	if err != nil {
		return store.Filter{}, sp, err
	}
	q := r.URL.Query()
	f := store.Filter{From: sp.fromS, To: sp.toS, TZ: cfg.Location().String(), Platform: strings.TrimSpace(q.Get("platform")), AppVersion: strings.TrimSpace(q.Get("appVersion")),
		UserID: strings.TrimSpace(q.Get("userId")), SessionID: strings.TrimSpace(q.Get("sessionId")), Source: strings.TrimSpace(q.Get("source")),
		IncludeFlagged: q.Get("includeFlagged") == "true" || q.Get("includeFlagged") == "1"}
	if q.Has("game") {
		g := strings.TrimSpace(q.Get("game"))
		if g != "*" {
			if !events.GamePattern.MatchString(g) {
				return f, sp, pluginkit.Bad("INVALID_GAME", "game", "")
			}
			f.Game = &g
		}
	}
	for _, n := range q["name"] {
		for _, part := range strings.Split(n, ",") {
			if part = strings.TrimSpace(part); part != "" {
				f.Names = append(f.Names, part)
			}
		}
	}
	for k, vals := range q {
		if !strings.HasPrefix(k, "tag.") {
			continue
		}
		g := strings.TrimPrefix(k, "tag.")
		if !events.GroupPattern.MatchString(g) {
			return f, sp, pluginkit.Bad("INVALID_TAG", k, "")
		}
		for _, v := range vals {
			for _, id := range strings.Split(v, ",") {
				if id = strings.TrimSpace(id); id != "" {
					if f.Tags == nil {
						f.Tags = map[string][]string{}
					}
					f.Tags[g] = append(f.Tags[g], id)
				}
			}
		}
	}
	if err := f.Validate(); err != nil {
		return f, sp, pluginkit.Bad("INVALID_QUERY", "", "")
	}
	return f, sp, nil
}

func (s *service) queryFail(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrBadQuery) {
		s.fail(w, pluginkit.Bad("INVALID_QUERY", "", ""))
		return
	}
	var ce *pluginkit.ConsoleError
	if errors.As(err, &ce) {
		s.fail(w, err)
		return
	}
	if s.logger != nil {
		s.logger.Warn("Insights store query failed: %v", err)
	}
	pluginkit.WriteJSON(w, http.StatusBadGateway, map[string]any{"error": "STORE_UNAVAILABLE", "message": err.Error()})
}

func (s *service) withStore(w http.ResponseWriter, r *http.Request, def int, fn func(ctx context.Context, st store.Store, f store.Filter, sp span, cfg settings.All) (any, error)) {
	cfg := s.settings(r.Context())
	f, sp, err := s.queryFilter(r, cfg, def)
	if err != nil {
		s.fail(w, err)
		return
	}
	st, err := s.eventStore(r.Context())
	if err != nil {
		s.queryFail(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	out, err := fn(ctx, st, f, sp, cfg)
	if err != nil {
		s.queryFail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func rangeView(sp span) map[string]any {
	return map[string]any{"from": sp.fromS, "to": sp.toS, "days": sp.days, "today": sp.today}
}

func (s *service) explore(w http.ResponseWriter, r *http.Request) {
	s.withStore(w, r, 30, func(ctx context.Context, st store.Store, f store.Filter, sp span, _ settings.All) (any, error) {
		q := r.URL.Query()
		aq := store.AggQuery{Filter: f, Metric: q.Get("metric"), Bucket: q.Get("bucket")}
		for _, g := range q["groupBy"] {
			for _, part := range strings.Split(g, ",") {
				if part = strings.TrimSpace(part); part != "" {
					aq.GroupBy = append(aq.GroupBy, part)
				}
			}
		}
		aq.Limit, _ = strconv.Atoi(q.Get("limit"))
		if err := aq.Normalize(); err != nil {
			return nil, err
		}
		rows, err := st.Aggregate(ctx, aq)
		if err != nil {
			return nil, err
		}
		return map[string]any{"range": rangeView(sp), "metric": aq.Metric, "bucket": aq.Bucket, "groupBy": nonNil(aq.GroupBy), "rows": rows, "limit": aq.Limit, "store": st.Kind()}, nil
	})
}

func (s *service) activity(w http.ResponseWriter, r *http.Request) {
	s.withStore(w, r, 30, func(ctx context.Context, st store.Store, f store.Filter, sp span, _ settings.All) (any, error) {
		a, err := st.Activity(ctx, f)
		if err != nil {
			return nil, err
		}
		var stick *float64
		if a.MAU > 0 {
			stick = store.Rate(a.DAU, a.MAU)
		}
		top, err := st.Aggregate(ctx, store.AggQuery{Filter: f, Metric: store.MetricCount, GroupBy: []string{"name"}, Limit: 10})
		if err != nil {
			return nil, err
		}
		platforms, err := st.Aggregate(ctx, store.AggQuery{Filter: f, Metric: store.MetricUsers, GroupBy: []string{"platform"}, Limit: 10})
		if err != nil {
			return nil, err
		}
		return map[string]any{"range": rangeView(sp), "activity": a, "stickiness": stick, "topEvents": top, "platforms": platforms, "store": st.Kind()}, nil
	})
}

func (s *service) retentionQuery(w http.ResponseWriter, r *http.Request) {
	s.withStore(w, r, 30, func(ctx context.Context, st store.Store, f store.Filter, sp span, cfg settings.All) (any, error) {
		q := store.RetentionQuery{Filter: f, Days: cfg.Retention.Days, Mode: cfg.Retention.Mode, Rows: cfg.Retention.CohortRows, Today: sp.today}
		if v := r.URL.Query().Get("days"); v != "" {
			q.Days = nil
			for _, p := range strings.Split(v, ",") {
				n, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil {
					return nil, store.ErrBadQuery
				}
				q.Days = append(q.Days, n)
			}
		}
		if m := r.URL.Query().Get("mode"); m != "" {
			q.Mode = m
		}
		res, err := st.Retention(ctx, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"range": rangeView(sp), "days": q.Days, "mode": q.Mode, "summary": res.Summary, "cohorts": res.Cohorts, "store": st.Kind()}, nil
	})
}

func (s *service) eventList(w http.ResponseWriter, r *http.Request) {
	s.withStore(w, r, 30, func(ctx context.Context, st store.Store, f store.Filter, sp span, _ settings.All) (any, error) {
		q := store.EventQuery{Filter: f, Asc: r.URL.Query().Get("order") == "asc"}
		q.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		if c := r.URL.Query().Get("cursor"); c != "" {
			at, id, ok := strings.Cut(c, "|")
			t, err := time.Parse(time.RFC3339Nano, at)
			if !ok || err != nil {
				return nil, store.ErrBadQuery
			}
			if _, err := uuid.FromString(id); err != nil {
				return nil, store.ErrBadQuery
			}
			q.BeforeAt, q.BeforeID = &t, id
		}
		list, err := st.Events(ctx, q)
		if err != nil {
			return nil, err
		}
		next := ""
		lim := q.Limit
		if lim <= 0 || lim > 1000 {
			lim = 100
		}
		if len(list) == lim {
			last := list[len(list)-1]
			next = last.At.UTC().Format(time.RFC3339Nano) + "|" + last.ID
		}
		return map[string]any{"range": rangeView(sp), "items": list, "next": next, "store": st.Kind()}, nil
	})
}

func (s *service) valueList(w http.ResponseWriter, r *http.Request) {
	s.withStore(w, r, 30, func(ctx context.Context, st store.Store, f store.Filter, sp span, _ settings.All) (any, error) {
		q := store.ValuesQuery{Filter: f, Kind: r.URL.Query().Get("kind")}
		q.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		list, err := st.Values(ctx, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"kind": q.Kind, "items": list}, nil
	})
}

func (s *service) sessionTimeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	var uid, client, day, platform, version string
	var started, last time.Time
	var ended sql.NullTime
	var secs sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT user_id::text, client_id, day::text, started_at, last_seen_at, ended_at, seconds, platform, app_version FROM insights_session WHERE id=$1`, id).
		Scan(&uid, &client, &day, &started, &last, &ended, &secs, &platform, &version)
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	st, err := s.eventStore(ctx)
	if err != nil {
		s.queryFail(w, err)
		return
	}
	cfg := s.settings(ctx)
	end := last
	if ended.Valid {
		end = ended.Time
	}
	loc := cfg.Location()
	f := store.Filter{From: started.In(loc).AddDate(0, 0, -1).Format(time.DateOnly), To: end.In(loc).AddDate(0, 0, 1).Format(time.DateOnly), UserID: uid, SessionID: client, IncludeFlagged: true}
	list, err := st.Events(ctx, store.EventQuery{Filter: f, Asc: true, Limit: 1000})
	if err != nil {
		s.queryFail(w, err)
		return
	}
	matched := "session"
	if len(list) == 0 {
		f.SessionID = ""
		all, err := st.Events(ctx, store.EventQuery{Filter: f, Asc: true, Limit: 1000})
		if err != nil {
			s.queryFail(w, err)
			return
		}
		for _, e := range all {
			if !e.At.Before(started.Add(-time.Minute)) && !e.At.After(end.Add(time.Minute)) {
				list = append(list, e)
			}
		}
		matched = "time"
	}
	rank := func(n string) int {
		switch n {
		case "session_start":
			return 0
		case "session_end":
			return 2
		}
		return 1
	}
	sort.SliceStable(list, func(i, j int) bool {
		if !list[i].At.Equal(list[j].At) {
			return list[i].At.Before(list[j].At)
		}
		return rank(list[i].Name) < rank(list[j].Name)
	})
	out := map[string]any{"session": map[string]any{"id": id, "userId": uid, "sessionId": client, "day": day, "startedAt": started, "lastSeenAt": last, "endedAt": nil, "seconds": nil, "platform": platform, "appVersion": version},
		"events": list, "matchedBy": matched}
	if ended.Valid {
		out["session"].(map[string]any)["endedAt"] = ended.Time
	}
	if secs.Valid {
		out["session"].(map[string]any)["seconds"] = secs.Int64
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

type funnelView struct {
	ID            string       `json:"id"`
	Game          string       `json:"game"`
	Name          string       `json:"name"`
	Steps         []store.Step `json:"steps"`
	WindowSeconds int          `json:"windowSeconds"`
	CreatedBy     string       `json:"createdBy"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

func (s *service) loadFunnel(ctx context.Context, id string) (*funnelView, error) {
	var f funnelView
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT id::text, game, name, steps, window_seconds, created_by, updated_at FROM insights_funnel WHERE id=$1`, id).Scan(&f.ID, &f.Game, &f.Name, &raw, &f.WindowSeconds, &f.CreatedBy, &f.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pluginkit.Missing("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(raw, &f.Steps)
	return &f, nil
}

func (s *service) funnelList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT id::text, game, name, steps, window_seconds, created_by, updated_at FROM insights_funnel ORDER BY name LIMIT 500`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []funnelView{}
	for rows.Next() {
		var f funnelView
		var raw []byte
		if err := rows.Scan(&f.ID, &f.Game, &f.Name, &raw, &f.WindowSeconds, &f.CreatedBy, &f.UpdatedAt); err != nil {
			s.fail(w, err)
			return
		}
		_ = json.Unmarshal(raw, &f.Steps)
		items = append(items, f)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func cleanFunnel(in *funnelView) error {
	in.Name = cleanLabel(in.Name, 120)
	if in.Name == "" {
		return pluginkit.Bad("REQUIRED", "name", "")
	}
	if !events.GamePattern.MatchString(in.Game) {
		return pluginkit.Bad("INVALID_GAME", "game", "")
	}
	if in.WindowSeconds == 0 {
		in.WindowSeconds = 86400
	}
	q := store.FunnelQuery{Steps: in.Steps, Window: time.Duration(in.WindowSeconds) * time.Second}
	if len(in.Steps) < 2 || len(in.Steps) > 12 {
		return pluginkit.Bad("VALUE_OUT_OF_RANGE", "steps", "")
	}
	for i, st := range in.Steps {
		if !events.ValidName(st.Name, 128) {
			return pluginkit.Bad("INVALID_EVENT_NAME", "steps."+strconv.Itoa(i)+".name", "")
		}
	}
	if err := store.ValidateFunnel(&q); err != nil {
		return pluginkit.Bad("VALUE_OUT_OF_RANGE", "windowSeconds", "")
	}
	return nil
}

func (s *service) funnelSave(w http.ResponseWriter, r *http.Request) {
	var in funnelView
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	if err := cleanFunnel(&in); err != nil {
		s.fail(w, err)
		return
	}
	id := pluginkit.PathVar(r, "id")
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
	} else if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	steps, _ := json.Marshal(in.Steps)
	res, err := s.db.ExecContext(r.Context(), `INSERT INTO insights_funnel (id, game, name, steps, window_seconds, created_by) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET game=EXCLUDED.game, name=EXCLUDED.name, steps=EXCLUDED.steps, window_seconds=EXCLUDED.window_seconds, updated_at=now()`,
		id, in.Game, in.Name, steps, in.WindowSeconds, pluginkit.ConsoleUser(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	f, err := s.loadFunnel(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, f)
}

func (s *service) funnelDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM insights_funnel WHERE id=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) funnelRun(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	fn, err := s.loadFunnel(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.withStore(w, r, 30, func(ctx context.Context, st store.Store, f store.Filter, sp span, _ settings.All) (any, error) {
		if !r.URL.Query().Has("game") {
			g := fn.Game
			f.Game = &g
		}
		win := fn.WindowSeconds
		if v, err := strconv.Atoi(r.URL.Query().Get("windowSeconds")); err == nil && v > 0 {
			win = v
		}
		steps, err := st.Funnel(ctx, store.FunnelQuery{Filter: f, Steps: fn.Steps, Window: time.Duration(win) * time.Second})
		if err != nil {
			return nil, err
		}
		return map[string]any{"range": rangeView(sp), "funnel": fn, "windowSeconds": win, "steps": funnelResult(steps), "store": st.Kind()}, nil
	})
}

func funnelResult(steps []store.FunnelStep) []map[string]any {
	out := make([]map[string]any, len(steps))
	var first int64
	for i, st := range steps {
		if i == 0 {
			first = st.Users
		}
		m := map[string]any{"index": st.Index, "name": st.Name, "users": st.Users, "medianSeconds": st.MedianSeconds, "fromFirst": store.Rate(st.Users, first), "fromPrevious": nil, "dropOff": int64(0)}
		if i > 0 {
			prev := steps[i-1].Users
			m["fromPrevious"] = store.Rate(st.Users, prev)
			m["dropOff"] = prev - st.Users
		}
		out[i] = m
	}
	return out
}

func (s *service) eventSettingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	d := settings.Defaults()
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"events": cfg.Events, "storage": cfg.Storage, "defaults": map[string]any{"events": d.Events, "storage": d.Storage},
		"timezone": cfg.Timezone, "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt})
}

func (s *service) eventSettingsPut(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Events  *settings.Events  `json:"events"`
		Storage *settings.Storage `json:"storage"`
	}
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	cfg, _, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	if in.Events != nil {
		cfg.Events = *in.Events
	}
	if in.Storage != nil {
		cfg.Storage = *in.Storage
	}
	if field, code := cfg.Validate(); code != "" {
		s.fail(w, pluginkit.Bad(code, field, ""))
		return
	}
	if err := settings.Save(r.Context(), s.db, cfg, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.eventSettingsGet(w, r)
}
