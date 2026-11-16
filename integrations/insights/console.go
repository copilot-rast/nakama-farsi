package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	p := consolePrefix
	routes := []route{
		{p + "/settings", http.MethodGet, s.settingsGet},
		{p + "/settings", http.MethodPut, s.settingsPut},
		{p + "/overview", http.MethodGet, s.overview},
		{p + "/sessions", http.MethodGet, s.sessionList},
		{p + "/heatmaps", http.MethodGet, s.mapList},
		{p + "/heatmaps/uploads/{id}/trace", http.MethodGet, s.uploadTrace},
		{p + "/heatmaps/{key}", http.MethodPut, s.mapSave},
		{p + "/heatmaps/{key}", http.MethodDelete, s.mapDelete},
		{p + "/heatmaps/{key}/cells", http.MethodGet, s.mapCells},
		{p + "/heatmaps/{key}/uploads", http.MethodGet, s.mapUploads},
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
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt, "storage": s.store.Driver()})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	in.Timezone = strings.TrimSpace(in.Timezone)
	if field, code := in.Validate(); code != "" {
		s.fail(w, pluginkit.Bad(code, field, ""))
		return
	}
	if err := settings.Save(r.Context(), s.db, in, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.settingsGet(w, r)
}

type span struct {
	from, to time.Time
	fromS    string
	toS      string
	today    string
	days     int
}

func (s *service) span(r *http.Request, cfg settings.All, def int) (span, error) {
	loc := cfg.Location()
	today := s.now().In(loc)
	q := r.URL.Query()
	parse := func(name string, fallback time.Time) (time.Time, error) {
		v := strings.TrimSpace(q.Get(name))
		if v == "" {
			return time.Date(fallback.Year(), fallback.Month(), fallback.Day(), 0, 0, 0, 0, time.UTC), nil
		}
		t, err := time.Parse(time.DateOnly, v)
		if err != nil {
			return t, pluginkit.Bad("INVALID_DATE", name, "")
		}
		return t, nil
	}
	to, err := parse("to", today)
	if err != nil {
		return span{}, err
	}
	from, err := parse("from", to.AddDate(0, 0, -(def-1)))
	if err != nil {
		return span{}, err
	}
	if from.After(to) {
		return span{}, pluginkit.Bad("INVALID_DATE", "from", "")
	}
	days := int(to.Sub(from).Hours()/24) + 1
	if days > 366 {
		return span{}, pluginkit.Bad("VALUE_OUT_OF_RANGE", "from", "")
	}
	return span{from: from, to: to, fromS: from.Format(time.DateOnly), toS: to.Format(time.DateOnly), today: today.Format(time.DateOnly), days: days}, nil
}

type filter struct {
	sql  []string
	args []any
}

func (f *filter) add(col, v string) {
	if v == "" {
		return
	}
	f.args = append(f.args, v)
	f.sql = append(f.sql, col+" = $"+strconv.Itoa(len(f.args)))
}

func (f *filter) arg(v any) string {
	f.args = append(f.args, v)
	return "$" + strconv.Itoa(len(f.args))
}

func (f *filter) where(extra ...string) string {
	all := append(append([]string{}, f.sql...), extra...)
	if len(all) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(all, " AND ")
}

func (f filter) clone() *filter {
	return &filter{sql: append([]string{}, f.sql...), args: append([]any{}, f.args...)}
}

func dims(r *http.Request, prefix string) *filter {
	f := &filter{}
	f.add(prefix+"platform", strings.TrimSpace(r.URL.Query().Get("platform")))
	f.add(prefix+"app_version", strings.TrimSpace(r.URL.Query().Get("appVersion")))
	return f
}

func rate(a, b int64) *float64 {
	if b <= 0 {
		return nil
	}
	v := math.Round(float64(a)/float64(b)*10000) / 10000
	return &v
}

func round1(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	x := math.Round(v.Float64*10) / 10
	return &x
}

func (s *service) overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := s.settings(ctx)
	sp, err := s.span(r, cfg, 30)
	if err != nil {
		s.fail(w, err)
		return
	}
	base := dims(r, "")
	type dayRow struct {
		Day        string  `json:"day"`
		DAU        int64   `json:"dau"`
		NewPlayers int64   `json:"newPlayers"`
		Sessions   int64   `json:"sessions"`
		AvgSeconds float64 `json:"avgSeconds"`
	}
	series := make([]dayRow, 0, sp.days)
	idx := map[string]int{}
	for d := sp.from; !d.After(sp.to); d = d.AddDate(0, 0, 1) {
		idx[d.Format(time.DateOnly)] = len(series)
		series = append(series, dayRow{Day: d.Format(time.DateOnly)})
	}
	f := base.clone()
	w1 := f.where("day >= "+f.arg(sp.fromS)+"::date", "day <= "+f.arg(sp.toS)+"::date")
	rows, err := s.db.QueryContext(ctx, `SELECT day::text, count(*)::int8 FROM insights_player_day`+w1+` GROUP BY 1`, f.args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	for rows.Next() {
		var d string
		var n int64
		if rows.Scan(&d, &n) == nil {
			if i, ok := idx[d]; ok {
				series[i].DAU = n
			}
		}
	}
	rows.Close()
	f = base.clone()
	w2 := f.where("first_day >= "+f.arg(sp.fromS)+"::date", "first_day <= "+f.arg(sp.toS)+"::date")
	var newPlayers int64
	rows, err = s.db.QueryContext(ctx, `SELECT first_day::text, count(*)::int8 FROM insights_player`+w2+` GROUP BY 1`, f.args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	for rows.Next() {
		var d string
		var n int64
		if rows.Scan(&d, &n) == nil {
			newPlayers += n
			if i, ok := idx[d]; ok {
				series[i].NewPlayers = n
			}
		}
	}
	rows.Close()
	f = base.clone()
	w3 := f.where("day >= "+f.arg(sp.fromS)+"::date", "day <= "+f.arg(sp.toS)+"::date")
	length := `coalesce(seconds, LEAST(EXTRACT(EPOCH FROM (last_seen_at - started_at)), ` + strconv.Itoa(cfg.Sessions.MaxSeconds) + `))`
	rows, err = s.db.QueryContext(ctx, `SELECT day::text, count(*)::int8, coalesce(avg(`+length+`),0)::float8 FROM insights_session`+w3+` GROUP BY 1`, f.args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	for rows.Next() {
		var d dayRow
		if rows.Scan(&d.Day, &d.Sessions, &d.AvgSeconds) == nil {
			if i, ok := idx[d.Day]; ok {
				series[i].Sessions, series[i].AvgSeconds = d.Sessions, math.Round(d.AvgSeconds*10)/10
			}
		}
	}
	rows.Close()
	var sessions, players int64
	var avg, median, total sql.NullFloat64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*)::int8, count(DISTINCT user_id)::int8, avg(`+length+`)::float8, percentile_cont(0.5) WITHIN GROUP (ORDER BY `+length+`)::float8, sum(`+length+`)::float8 FROM insights_session`+w3, f.args...).Scan(&sessions, &players, &avg, &median, &total); err != nil {
		s.fail(w, err)
		return
	}
	var dauSum int64
	for _, d := range series {
		dauSum += d.DAU
	}
	active := map[string]any{"day": sp.toS}
	for name, back := range map[string]int{"dau": 0, "wau": 6, "mau": 29} {
		f = base.clone()
		var n int64
		wq := f.where("day >= "+f.arg(sp.to.AddDate(0, 0, -back).Format(time.DateOnly))+"::date", "day <= "+f.arg(sp.toS)+"::date")
		_ = s.db.QueryRowContext(ctx, `SELECT count(DISTINCT user_id)::int8 FROM insights_player_day`+wq, f.args...).Scan(&n)
		active[name] = n
	}
	var open int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*)::int8 FROM insights_session WHERE ended_at IS NULL`).Scan(&open)
	retention, cohorts, err := s.retention(ctx, base, sp, cfg)
	if err != nil {
		s.fail(w, err)
		return
	}
	breakdown := func(col string) []map[string]any {
		f := base.clone()
		out := []map[string]any{}
		wq := f.where("day >= "+f.arg(sp.fromS)+"::date", "day <= "+f.arg(sp.toS)+"::date")
		rows, err := s.db.QueryContext(ctx, `SELECT `+col+`, count(DISTINCT user_id)::int8, count(*)::int8 FROM insights_session`+wq+` GROUP BY 1 ORDER BY 2 DESC LIMIT 20`, f.args...)
		if err != nil {
			return out
		}
		defer rows.Close()
		for rows.Next() {
			var k string
			var p, n int64
			if rows.Scan(&k, &p, &n) == nil {
				out = append(out, map[string]any{"key": k, "players": p, "sessions": n})
			}
		}
		return out
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{
		"range": map[string]any{"from": sp.fromS, "to": sp.toS, "days": sp.days, "today": sp.today},
		"totals": map[string]any{"sessions": sessions, "players": players, "newPlayers": newPlayers, "avgSeconds": round1(avg), "medianSeconds": round1(median),
			"totalSeconds": round1(total), "sessionsPerPlayer": rate(sessions, players), "avgDau": math.Round(float64(dauSum)/float64(sp.days)*10) / 10, "open": open},
		"active": active, "series": series, "retention": retention, "cohorts": cohorts,
		"retentionMode": cfg.Retention.Mode, "retentionDays": cfg.Retention.Days, "sessionSource": cfg.Sessions.Source,
		"platforms": breakdown("platform"), "versions": breakdown("app_version"),
	})
}

type retentionRow struct {
	Day      int      `json:"day"`
	Cohort   int64    `json:"cohort"`
	Returned int64    `json:"returned"`
	Rate     *float64 `json:"rate"`
}

func (s *service) retention(ctx context.Context, base *filter, sp span, cfg settings.All) ([]retentionRow, []map[string]any, error) {
	cmp := "="
	if cfg.Retention.Mode == settings.ModeRolling {
		cmp = ">="
	}
	out := []retentionRow{}
	for _, n := range cfg.Retention.Days {
		f := base.clone()
		nArg := f.arg(n)
		wq := f.where("first_day >= "+f.arg(sp.fromS)+"::date", "first_day <= "+f.arg(sp.toS)+"::date", "first_day + "+nArg+"::int <= "+f.arg(sp.today)+"::date")
		var cohort, back int64
		if err := s.db.QueryRowContext(ctx, `SELECT count(*)::int8, count(*) FILTER (WHERE EXISTS (SELECT 1 FROM insights_player_day d WHERE d.user_id=p.user_id AND d.day `+cmp+` p.first_day + `+nArg+`::int))::int8
			FROM insights_player p`+wq, f.args...).Scan(&cohort, &back); err != nil {
			return nil, nil, err
		}
		out = append(out, retentionRow{Day: n, Cohort: cohort, Returned: back, Rate: rate(back, cohort)})
	}
	f := base.clone()
	cols := []string{}
	for _, n := range cfg.Retention.Days {
		a := f.arg(n)
		cols = append(cols, `CASE WHEN p.first_day + `+a+`::int <= TODAY THEN count(*) FILTER (WHERE EXISTS (SELECT 1 FROM insights_player_day d WHERE d.user_id=p.user_id AND d.day `+cmp+` p.first_day + `+a+`::int)) END`)
	}
	today := f.arg(sp.today) + "::date"
	for i := range cols {
		cols[i] = strings.ReplaceAll(cols[i], "TODAY", today)
	}
	wq := f.where("first_day >= "+f.arg(sp.fromS)+"::date", "first_day <= "+f.arg(sp.toS)+"::date")
	lim := f.arg(cfg.Retention.CohortRows)
	rows, err := s.db.QueryContext(ctx, `SELECT p.first_day::text, count(*)::int8, `+strings.Join(cols, ", ")+` FROM insights_player p`+wq+` GROUP BY p.first_day ORDER BY p.first_day DESC LIMIT `+lim, f.args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	cohorts := []map[string]any{}
	for rows.Next() {
		var day string
		var size int64
		vals := make([]sql.NullInt64, len(cfg.Retention.Days))
		dest := []any{&day, &size}
		for i := range vals {
			dest = append(dest, &vals[i])
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, nil, err
		}
		cells := []map[string]any{}
		for i, n := range cfg.Retention.Days {
			c := map[string]any{"day": n, "returned": nil, "rate": nil}
			if vals[i].Valid {
				c["returned"], c["rate"] = vals[i].Int64, rate(vals[i].Int64, size)
			}
			cells = append(cells, c)
		}
		cohorts = append(cohorts, map[string]any{"day": day, "size": size, "cells": cells})
	}
	return out, cohorts, rows.Err()
}

func (s *service) sessionList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 20, 100)
	f := dims(r, "s.")
	if u := strings.TrimSpace(r.URL.Query().Get("userId")); u != "" {
		f.sql = append(f.sql, "s.user_id::text = "+f.arg(u))
	}
	switch r.URL.Query().Get("status") {
	case "open":
		f.sql = append(f.sql, "s.ended_at IS NULL")
	case "ended":
		f.sql = append(f.sql, "s.ended_at IS NOT NULL")
	}
	wq := f.where()
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM insights_session s`+wq, f.args...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := f.arg(limit), f.arg((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT s.id::text, s.user_id::text, coalesce(u.username,''), coalesce(u.display_name,''), s.client_id, s.source, s.day::text, s.started_at, s.last_seen_at, s.ended_at, s.seconds, s.pings, s.app_version, s.platform, s.device, s.os
		FROM insights_session s LEFT JOIN users u ON u.id=s.user_id`+wq+` ORDER BY s.started_at DESC LIMIT `+lim+` OFFSET `+off, f.args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, uid, username, display, client, source, day, version, platform, device, os string
		var started, last time.Time
		var ended sql.NullTime
		var secs sql.NullInt64
		var pings int
		if err := rows.Scan(&id, &uid, &username, &display, &client, &source, &day, &started, &last, &ended, &secs, &pings, &version, &platform, &device, &os); err != nil {
			s.fail(w, err)
			return
		}
		it := map[string]any{"id": id, "userId": uid, "username": username, "displayName": display, "sessionId": client, "source": source, "day": day, "startedAt": started, "lastSeenAt": last,
			"endedAt": nil, "seconds": nil, "pings": pings, "appVersion": version, "platform": platform, "device": device, "os": os}
		if ended.Valid {
			it["endedAt"] = ended.Time
		}
		if secs.Valid {
			it["seconds"] = secs.Int64
		}
		items = append(items, it)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

type bounds struct {
	MinX *float64 `json:"minX"`
	MaxX *float64 `json:"maxX"`
	MinZ *float64 `json:"minZ"`
	MaxZ *float64 `json:"maxZ"`
}

func (b bounds) complete() bool {
	return b.MinX != nil && b.MaxX != nil && b.MinZ != nil && b.MaxZ != nil && *b.MinX < *b.MaxX && *b.MinZ < *b.MaxZ
}

func nullF(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	x := v.Float64
	return &x
}

const mapCols = `key, name, cell_size, background_url, image_bounds::text, min_x, max_x, min_z, max_z, enabled, uploads, points, last_upload_at, created_at`

func (s *service) scanMap(sc interface{ Scan(...any) error }, def float64) (map[string]any, error) {
	var key, name, bg, img string
	var size float64
	var minX, maxX, minZ, maxZ sql.NullFloat64
	var enabled bool
	var uploads, points int64
	var last sql.NullTime
	var created time.Time
	if err := sc.Scan(&key, &name, &size, &bg, &img, &minX, &maxX, &minZ, &maxZ, &enabled, &uploads, &points, &last, &created); err != nil {
		return nil, err
	}
	var ib bounds
	_ = json.Unmarshal([]byte(img), &ib)
	eff := size
	if eff <= 0 {
		eff = def
	}
	m := map[string]any{"key": key, "name": name, "cellSize": size, "effectiveCellSize": eff, "backgroundUrl": bg, "imageBounds": ib,
		"bounds": bounds{nullF(minX), nullF(maxX), nullF(minZ), nullF(maxZ)}, "enabled": enabled, "uploads": uploads, "points": points, "lastUploadAt": nil, "createdAt": created}
	if last.Valid {
		m["lastUploadAt"] = last.Time
	}
	return m, nil
}

func (s *service) mapList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := s.settings(ctx)
	rows, err := s.db.QueryContext(ctx, `SELECT `+mapCols+` FROM insights_heatmap_map ORDER BY coalesce(last_upload_at, created_at) DESC, key LIMIT 500`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		m, err := s.scanMap(rows, cfg.Heatmap.CellSize)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, m)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "enabled": cfg.Heatmap.Enabled, "cellSize": cfg.Heatmap.CellSize, "unknownMaps": cfg.Heatmap.UnknownMaps})
}

func (s *service) loadMap(ctx context.Context, key string, def float64) (map[string]any, error) {
	m, err := s.scanMap(s.db.QueryRowContext(ctx, `SELECT `+mapCols+` FROM insights_heatmap_map WHERE key=$1`, key), def)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pluginkit.Missing("HEATMAP_MAP_NOT_FOUND")
	}
	return m, err
}

func (s *service) mapSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := pluginkit.PathVar(r, "key")
	if !mapKeyPattern.MatchString(key) {
		s.fail(w, pluginkit.Bad("HEATMAP_INVALID", "key", ""))
		return
	}
	var in struct {
		Name          string  `json:"name"`
		CellSize      float64 `json:"cellSize"`
		BackgroundURL string  `json:"backgroundUrl"`
		ImageBounds   bounds  `json:"imageBounds"`
		Enabled       *bool   `json:"enabled"`
	}
	if err := pluginkit.ReadBody(r, &in, 16*1024); err != nil {
		s.fail(w, err)
		return
	}
	in.Name = cleanLabel(in.Name, 100)
	in.BackgroundURL = strings.TrimSpace(in.BackgroundURL)
	if in.CellSize < 0 || in.CellSize > 1000 || (in.CellSize > 0 && in.CellSize < 0.1) {
		s.fail(w, pluginkit.Bad("VALUE_OUT_OF_RANGE", "cellSize", ""))
		return
	}
	if in.BackgroundURL != "" {
		u, err := url.Parse(in.BackgroundURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(in.BackgroundURL) > 1000 {
			s.fail(w, pluginkit.Bad("INVALID_URL", "backgroundUrl", ""))
			return
		}
	}
	img := "{}"
	if in.ImageBounds.MinX != nil || in.ImageBounds.MaxX != nil || in.ImageBounds.MinZ != nil || in.ImageBounds.MaxZ != nil {
		if !in.ImageBounds.complete() {
			s.fail(w, pluginkit.Bad("VALUE_OUT_OF_RANGE", "imageBounds", ""))
			return
		}
		raw, _ := json.Marshal(in.ImageBounds)
		img = string(raw)
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	if in.Name == "" {
		in.Name = key
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO insights_heatmap_map (key, name, cell_size, background_url, image_bounds, enabled) VALUES ($1,$2,$3,$4,$5::jsonb,$6)
		ON CONFLICT (key) DO UPDATE SET name=EXCLUDED.name, cell_size=EXCLUDED.cell_size, background_url=EXCLUDED.background_url, image_bounds=EXCLUDED.image_bounds, enabled=EXCLUDED.enabled, updated_at=now()`,
		key, in.Name, in.CellSize, in.BackgroundURL, img, enabled); err != nil {
		s.fail(w, err)
		return
	}
	m, err := s.loadMap(ctx, key, s.settings(ctx).Heatmap.CellSize)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": m})
}

func (s *service) mapDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := pluginkit.PathVar(r, "key")
	s.dropRaw(ctx, `SELECT id::text, raw_key FROM insights_heatmap_upload WHERE map_key=$1 AND raw_key <> ''`, key)
	res, err := s.db.ExecContext(ctx, `DELETE FROM insights_heatmap_map WHERE key=$1`, key)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("HEATMAP_MAP_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) mapCells(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := s.settings(ctx)
	m, err := s.loadMap(ctx, pluginkit.PathVar(r, "key"), cfg.Heatmap.CellSize)
	if err != nil {
		s.fail(w, err)
		return
	}
	sp, err := s.span(r, cfg, 30)
	if err != nil {
		s.fail(w, err)
		return
	}
	size := m["effectiveCellSize"].(float64)
	if v := r.URL.Query().Get("size"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			size = f
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT cx, cz, sum(points)::int8, (sum(speed_sum)/GREATEST(sum(points),1))::float8, max(speed_max)::float8 FROM insights_heatmap_cell
		WHERE map_key=$1 AND size=$2 AND day >= $3::date AND day <= $4::date GROUP BY cx, cz ORDER BY 3 DESC LIMIT 60000`, m["key"], size, sp.fromS, sp.toS)
	if err != nil {
		s.fail(w, err)
		return
	}
	cells := [][]any{}
	var total int64
	var maxN int64
	for rows.Next() {
		var cx, cz int
		var n int64
		var avg, mx float64
		if rows.Scan(&cx, &cz, &n, &avg, &mx) == nil {
			total += n
			if n > maxN {
				maxN = n
			}
			cells = append(cells, []any{cx, cz, n, math.Round(avg*10) / 10, math.Round(mx*10) / 10})
		}
	}
	rows.Close()
	sizes := []float64{}
	srows, err := s.db.QueryContext(ctx, `SELECT DISTINCT size FROM insights_heatmap_cell WHERE map_key=$1 ORDER BY 1`, m["key"])
	if err == nil {
		for srows.Next() {
			var v float64
			if srows.Scan(&v) == nil {
				sizes = append(sizes, v)
			}
		}
		srows.Close()
	}
	var uploads, players int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*)::int8, count(DISTINCT user_id)::int8 FROM insights_heatmap_upload WHERE map_key=$1 AND created_at >= $2 AND created_at < $3`,
		m["key"], sp.from.In(cfg.Location()), sp.to.AddDate(0, 0, 1).In(cfg.Location())).Scan(&uploads, &players)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"map": m, "size": size, "sizes": sizes, "cells": cells,
		"range": map[string]any{"from": sp.fromS, "to": sp.toS}, "totals": map[string]any{"points": total, "maxPoints": maxN, "cells": len(cells), "uploads": uploads, "players": players}})
}

func (s *service) mapUploads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 20, 100)
	key := pluginkit.PathVar(r, "key")
	var total int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM insights_heatmap_upload WHERE map_key=$1`, key).Scan(&total)
	rows, err := s.db.QueryContext(ctx, `SELECT h.id::text, h.user_id::text, coalesce(u.username,''), coalesce(u.display_name,''), h.upload_key, h.part, h.points, h.dropped, h.players, h.duration_ms, h.raw_key <> '', h.created_at
		FROM insights_heatmap_upload h LEFT JOIN users u ON u.id=h.user_id WHERE h.map_key=$1 ORDER BY h.created_at DESC LIMIT $2 OFFSET $3`, key, limit, (page-1)*limit)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, uid, username, display, upload string
		var part, points, dropped, players int
		var dur int64
		var raw bool
		var at time.Time
		if err := rows.Scan(&id, &uid, &username, &display, &upload, &part, &points, &dropped, &players, &dur, &raw, &at); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, map[string]any{"id": id, "userId": uid, "username": username, "displayName": display, "uploadId": upload, "part": part, "points": points, "dropped": dropped,
			"players": players, "durationMs": dur, "hasTrace": raw, "createdAt": at})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) uploadTrace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var key, mapKey string
	if err := s.db.QueryRowContext(ctx, `SELECT raw_key, map_key FROM insights_heatmap_upload WHERE id::text=$1`, id).Scan(&key, &mapKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.fail(w, pluginkit.Missing("HEATMAP_UPLOAD_NOT_FOUND"))
			return
		}
		s.fail(w, err)
		return
	}
	if key == "" {
		s.fail(w, pluginkit.Missing("HEATMAP_TRACE_GONE"))
		return
	}
	pts, err := s.trace(ctx, key, s.settings(ctx).Heatmap.MaxTracePoints)
	if err != nil {
		if errors.Is(err, objectstore.ErrNotFound) {
			s.fail(w, pluginkit.Missing("HEATMAP_TRACE_GONE"))
			return
		}
		s.fail(w, err)
		return
	}
	order := []string{}
	byPlayer := map[string][][]float64{}
	for _, p := range pts {
		if _, ok := byPlayer[p.Player]; !ok {
			order = append(order, p.Player)
		}
		byPlayer[p.Player] = append(byPlayer[p.Player], []float64{p.T, p.X, p.Z, p.Kmh})
	}
	players := []map[string]any{}
	for _, name := range order {
		players = append(players, map[string]any{"player": name, "points": byPlayer[name]})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "map": mapKey, "players": players, "points": len(pts)})
}
