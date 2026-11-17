package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	mapKeyPattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,63}$`)
	uploadKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,64}$`)
)

type point struct {
	T, X, Z float64
	Player  string
	Kmh     float64
}

type heatmapIn struct {
	Map      string  `json:"map"`
	UploadID string  `json:"uploadId"`
	Part     int     `json:"part"`
	Format   string  `json:"format"`
	Data     string  `json:"data"`
	Points   [][]any `json:"points"`
}

const (
	dropColumns = "columns"
	dropNumber  = "number"
	dropRange   = "range"
)

type csvLayout struct {
	sep                  string
	t, x, z, player, kmh int
}

var defaultLayout = csvLayout{sep: ",", t: 0, x: 1, z: 2, player: 3, kmh: 4}

var headerAliases = map[string]string{
	"t": "t", "time": "t", "timestamp": "t", "sec": "t", "seconds": "t", "elapsed": "t",
	"x": "x", "posx": "x", "positionx": "x",
	"z": "z", "posz": "z", "positionz": "z",
	"player": "player", "name": "player", "playername": "player", "driver": "player", "racer": "player", "playerid": "player",
	"kmh": "kmh", "speed": "kmh", "speedkmh": "kmh",
}

type csvStats struct {
	Rows    int            `json:"rows"`
	Dropped int            `json:"dropped"`
	Reasons map[string]int `json:"reasons"`
}

func (c *csvStats) drop(reason string) {
	c.Dropped++
	if c.Reasons == nil {
		c.Reasons = map[string]int{}
	}
	c.Reasons[reason]++
}

func headerKey(v string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(v) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func detectSep(line string) string {
	switch {
	case strings.Contains(line, ","):
		return ","
	case strings.Contains(line, ";"):
		return ";"
	case strings.Contains(line, "\t"):
		return "\t"
	}
	return ","
}

func headerLayout(fields []string, sep string) (csvLayout, bool) {
	l := csvLayout{sep: sep, t: -1, x: -1, z: -1, player: -1, kmh: -1}
	for i, f := range fields {
		switch headerAliases[headerKey(f)] {
		case "t":
			if l.t < 0 {
				l.t = i
			}
		case "x":
			if l.x < 0 {
				l.x = i
			}
		case "z":
			if l.z < 0 {
				l.z = i
			}
		case "player":
			if l.player < 0 {
				l.player = i
			}
		case "kmh":
			if l.kmh < 0 {
				l.kmh = i
			}
		}
	}
	if l.x < 0 || l.z < 0 {
		return csvLayout{}, false
	}
	return l, true
}

func col(f []string, i int) string {
	if i < 0 || i >= len(f) {
		return ""
	}
	return strings.Trim(strings.TrimSpace(f[i]), `"'`)
}

func scanCSV(r io.Reader, fn func(point) error) (csvStats, error) {
	st := csvStats{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 4096), 64*1024)
	var layout *csvLayout
	for sc.Scan() {
		raw := strings.TrimSpace(strings.TrimPrefix(sc.Text(), "\ufeff"))
		if raw == "" {
			continue
		}
		first := layout == nil
		if first {
			l := defaultLayout
			l.sep = detectSep(raw)
			layout = &l
		}
		f := strings.Split(raw, layout.sep)
		need := max(layout.x, layout.z, layout.t) + 1
		if len(f) < need || len(f) < 2 {
			if first {
				continue
			}
			st.Rows++
			st.drop(dropColumns)
			continue
		}
		x, e2 := strconv.ParseFloat(col(f, layout.x), 64)
		z, e3 := strconv.ParseFloat(col(f, layout.z), 64)
		var t float64
		var e1 error
		if layout.t >= 0 {
			t, e1 = strconv.ParseFloat(col(f, layout.t), 64)
		}
		if e1 != nil || e2 != nil || e3 != nil {
			if first {
				if l, ok := headerLayout(f, layout.sep); ok {
					layout = &l
				}
				continue
			}
			st.Rows++
			st.drop(dropNumber)
			continue
		}
		st.Rows++
		p := point{T: t, X: x, Z: z, Player: cleanLabel(col(f, layout.player), 32)}
		if v := col(f, layout.kmh); v != "" {
			if k, err := strconv.ParseFloat(v, 64); err == nil {
				p.Kmh = k
			}
		}
		if err := fn(p); err != nil {
			return st, err
		}
	}
	if err := sc.Err(); err != nil {
		if errors.As(err, new(limitErr)) {
			return st, err
		}
		return st, apierr.BadRequest("HEATMAP_INVALID").WithField("data").WithDetail(err.Error())
	}
	return st, nil
}

func readCSV(r io.Reader, max int) ([]point, int, error) {
	out := []point{}
	st, err := scanCSV(r, func(p point) error {
		out = append(out, p)
		if len(out) > max {
			return apierr.New(413, "HEATMAP_TOO_LARGE").With("maxPoints", max)
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return out, st.Dropped, nil
}

func validPoint(p point, maxKmh float64) bool {
	return !(math.IsNaN(p.X) || math.IsNaN(p.Z) || math.IsInf(p.X, 0) || math.IsInf(p.Z, 0) || math.Abs(p.X) > 1e7 || math.Abs(p.Z) > 1e7 || math.IsNaN(p.T) || math.IsInf(p.T, 0) || math.IsNaN(p.Kmh) || p.Kmh < 0 || p.Kmh > maxKmh)
}

type heatAgg struct {
	size                               float64
	cells                              map[cellKey]*cellAgg
	players                            map[string]bool
	minT, maxT, minX, maxX, minZ, maxZ float64
	kept                               int
	raw                                *bytes.Buffer
	zw                                 *gzip.Writer
}

func newHeatAgg(size float64, keepRaw bool) *heatAgg {
	a := &heatAgg{size: size, cells: map[cellKey]*cellAgg{}, players: map[string]bool{},
		minT: math.Inf(1), maxT: math.Inf(-1), minX: math.Inf(1), maxX: math.Inf(-1), minZ: math.Inf(1), maxZ: math.Inf(-1)}
	if keepRaw {
		a.raw = &bytes.Buffer{}
		a.zw, _ = gzip.NewWriterLevel(a.raw, gzip.BestCompression)
		_, _ = a.zw.Write([]byte("t,x,z,player,kmh\n"))
	}
	return a
}

func (a *heatAgg) add(p point) {
	a.kept++
	a.players[p.Player] = true
	a.minT, a.maxT = math.Min(a.minT, p.T), math.Max(a.maxT, p.T)
	a.minX, a.maxX, a.minZ, a.maxZ = math.Min(a.minX, p.X), math.Max(a.maxX, p.X), math.Min(a.minZ, p.Z), math.Max(a.maxZ, p.Z)
	k := cellKey{int(math.Floor(p.X / a.size)), int(math.Floor(p.Z / a.size))}
	c := a.cells[k]
	if c == nil {
		c = &cellAgg{}
		a.cells[k] = c
	}
	c.points++
	c.speedSum += p.Kmh
	c.speedMax = math.Max(c.speedMax, p.Kmh)
	if a.zw != nil {
		writePoint(a.zw, p)
	}
}

func (a *heatAgg) rawBytes() []byte {
	if a.zw == nil {
		return nil
	}
	_ = a.zw.Close()
	return a.raw.Bytes()
}

func (a *heatAgg) durationMs() int64 {
	if a.kept == 0 {
		return 0
	}
	return int64(math.Round((a.maxT - a.minT) * 1000))
}

func (s *service) writeHeat(ctx context.Context, tx *sql.Tx, mapKey string, a *heatAgg, day string, now time.Time) error {
	keys := make([]cellKey, 0, len(a.cells))
	for k := range a.cells {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].cx != keys[j].cx {
			return keys[i].cx < keys[j].cx
		}
		return keys[i].cz < keys[j].cz
	})
	cxs, czs, counts, sums, maxes := make([]int64, 0, len(keys)), make([]int64, 0, len(keys)), make([]int64, 0, len(keys)), make([]float64, 0, len(keys)), make([]float64, 0, len(keys))
	for _, k := range keys {
		c := a.cells[k]
		cxs, czs, counts, sums, maxes = append(cxs, int64(k.cx)), append(czs, int64(k.cz)), append(counts, c.points), append(sums, c.speedSum), append(maxes, c.speedMax)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO insights_heatmap_cell (map_key, size, day, cx, cz, points, speed_sum, speed_max)
		SELECT $1, $2, $3::date, u.cx, u.cz, u.n, u.s, u.m FROM unnest($4::int[], $5::int[], $6::bigint[], $7::float8[], $8::float8[]) AS u(cx, cz, n, s, m)
		ON CONFLICT (map_key, size, day, cx, cz) DO UPDATE SET points=insights_heatmap_cell.points+EXCLUDED.points,
		speed_sum=insights_heatmap_cell.speed_sum+EXCLUDED.speed_sum, speed_max=GREATEST(insights_heatmap_cell.speed_max, EXCLUDED.speed_max)`,
		mapKey, a.size, day, cxs, czs, counts, sums, maxes); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE insights_heatmap_map SET uploads=uploads+1, points=points+$2, last_upload_at=$3,
		min_x=LEAST(coalesce(min_x,$4),$4), max_x=GREATEST(coalesce(max_x,$5),$5), min_z=LEAST(coalesce(min_z,$6),$6), max_z=GREATEST(coalesce(max_z,$7),$7) WHERE key=$1`,
		mapKey, a.kept, now, a.minX, a.maxX, a.minZ, a.maxZ)
	return err
}

func num(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f, err == nil
	}
	return 0, false
}

func (in heatmapIn) points(cfg settings.Heatmap) ([]point, int, error) {
	switch strings.ToLower(strings.TrimSpace(in.Format)) {
	case "", "csv":
		if in.Data == "" && len(in.Points) > 0 {
			break
		}
		if len(in.Data) > cfg.MaxBytes {
			return nil, 0, apierr.New(413, "HEATMAP_TOO_LARGE").With("maxBytes", cfg.MaxBytes)
		}
		return readCSV(strings.NewReader(in.Data), cfg.MaxPoints)
	case "csv.gz", "gzip":
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in.Data))
		if err != nil {
			return nil, 0, apierr.BadRequest("HEATMAP_INVALID").WithField("data").WithDetail("data is not base64")
		}
		zr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, 0, apierr.BadRequest("HEATMAP_INVALID").WithField("data").WithDetail("data is not gzip")
		}
		plain, err := io.ReadAll(io.LimitReader(zr, int64(cfg.MaxBytes)+1))
		if err != nil {
			return nil, 0, apierr.BadRequest("HEATMAP_INVALID").WithField("data").WithDetail("gzip data is broken")
		}
		if len(plain) > cfg.MaxBytes {
			return nil, 0, apierr.New(413, "HEATMAP_TOO_LARGE").With("maxBytes", cfg.MaxBytes)
		}
		return readCSV(bytes.NewReader(plain), cfg.MaxPoints)
	default:
		return nil, 0, apierr.BadRequest("HEATMAP_INVALID").WithField("format")
	}
	if len(in.Points) > cfg.MaxPoints {
		return nil, 0, apierr.New(413, "HEATMAP_TOO_LARGE").With("maxPoints", cfg.MaxPoints)
	}
	out := make([]point, 0, len(in.Points))
	dropped := 0
	for _, row := range in.Points {
		if len(row) < 3 {
			dropped++
			continue
		}
		t, ok1 := num(row[0])
		x, ok2 := num(row[1])
		z, ok3 := num(row[2])
		if !ok1 || !ok2 || !ok3 {
			dropped++
			continue
		}
		p := point{T: t, X: x, Z: z}
		if len(row) > 3 && row[3] != nil {
			p.Player = cleanLabel(fmt.Sprint(row[3]), 32)
		}
		if len(row) > 4 {
			p.Kmh, _ = num(row[4])
		}
		out = append(out, p)
	}
	return out, dropped, nil
}

type mapRow struct {
	Key      string
	CellSize float64
	Enabled  bool
}

func (s *service) heatmapMap(ctx context.Context, q queryer, key string, cfg settings.Heatmap) (mapRow, error) {
	m := mapRow{Key: key}
	err := q.QueryRowContext(ctx, `SELECT cell_size, enabled FROM insights_heatmap_map WHERE key=$1`, key).Scan(&m.CellSize, &m.Enabled)
	if errors.Is(err, sql.ErrNoRows) {
		if cfg.UnknownMaps != settings.MapsAccept {
			return m, apierr.NotFound("HEATMAP_MAP_UNKNOWN").WithField("map").With("map", key)
		}
		if _, err := q.ExecContext(ctx, `INSERT INTO insights_heatmap_map (key, name) VALUES ($1,$1) ON CONFLICT (key) DO NOTHING`, key); err != nil {
			return m, err
		}
		next := cfg
		next.UnknownMaps = settings.MapsReject
		return s.heatmapMap(ctx, q, key, next)
	}
	if err != nil {
		return m, err
	}
	if !m.Enabled {
		return m, apierr.Forbidden("HEATMAP_MAP_DISABLED").WithField("map").With("map", key)
	}
	if m.CellSize <= 0 {
		m.CellSize = cfg.CellSize
	}
	return m, nil
}

func writePoint(w io.Writer, p point) {
	_, _ = fmt.Fprintf(w, "%s,%s,%s,%s,%s\n", strconv.FormatFloat(p.T, 'f', -1, 64), strconv.FormatFloat(p.X, 'f', -1, 64), strconv.FormatFloat(p.Z, 'f', -1, 64), strings.ReplaceAll(p.Player, ",", " "), strconv.FormatFloat(p.Kmh, 'f', -1, 64))
}

type cellKey struct{ cx, cz int }

type cellAgg struct {
	points   int64
	speedSum float64
	speedMax float64
}

func (s *service) rpcHeatmap(ctx context.Context, userID, payload string) (any, error) {
	var in heatmapIn
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	return s.heatmap(ctx, userID, in)
}

func (s *service) heatmap(ctx context.Context, userID string, in heatmapIn) (map[string]any, error) {
	cfg := s.settings(ctx)
	hc := cfg.Heatmap
	in.Map, in.UploadID = strings.TrimSpace(in.Map), strings.TrimSpace(in.UploadID)
	switch {
	case !hc.Enabled:
		return nil, apierr.Forbidden("HEATMAP_DISABLED")
	case !mapKeyPattern.MatchString(in.Map):
		return nil, apierr.BadRequest("HEATMAP_INVALID").WithField("map")
	case !uploadKeyPattern.MatchString(in.UploadID):
		return nil, apierr.BadRequest("HEATMAP_INVALID").WithField("uploadId")
	case in.Part < 0 || in.Part > 999:
		return nil, apierr.BadRequest("HEATMAP_INVALID").WithField("part")
	}
	var prior int
	err := s.db.QueryRowContext(ctx, `SELECT points FROM insights_heatmap_upload WHERE user_id=$1 AND upload_key=$2 AND part=$3`, userID, in.UploadID, in.Part).Scan(&prior)
	if err == nil {
		return map[string]any{"uploadId": in.UploadID, "part": in.Part, "duplicate": true, "accepted": prior}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	pts, dropped, err := in.points(hc)
	if err != nil {
		return nil, err
	}
	kept := pts[:0]
	for _, p := range pts {
		if !validPoint(p, hc.MaxKmh) {
			dropped++
			continue
		}
		kept = append(kept, p)
	}
	if len(kept) == 0 {
		return nil, apierr.BadRequest("HEATMAP_INVALID").WithField("data").With("dropped", dropped)
	}
	now := s.now().UTC()
	var recent int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM insights_heatmap_upload WHERE user_id=$1 AND created_at > $2`, userID, now.Add(-time.Hour)).Scan(&recent); err != nil {
		return nil, err
	}
	if recent >= hc.UploadsPerHour {
		return nil, apierr.TooManyRequests("ANALYTICS_RATE_LIMITED", 60).With("limit", hc.UploadsPerHour)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	m, err := s.heatmapMap(ctx, tx, in.Map, hc)
	if err != nil {
		return nil, err
	}
	agg := newHeatAgg(m.CellSize, hc.KeepRawDays > 0)
	for _, p := range kept {
		agg.add(p)
	}
	id := uuid.Must(uuid.NewV4()).String()
	rawKey := ""
	if hc.KeepRawDays > 0 {
		rawKey = rawKeyFor(in.Map, id)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO insights_heatmap_upload (id, user_id, map_key, upload_key, part, points, dropped, players, duration_ms, raw_key, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, userID, in.Map, in.UploadID, in.Part, agg.kept, dropped, len(agg.players), agg.durationMs(), rawKey, now); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return map[string]any{"uploadId": in.UploadID, "part": in.Part, "duplicate": true, "accepted": agg.kept}, nil
		}
		return nil, err
	}
	day := now.In(cfg.Location()).Format(time.DateOnly)
	if err := s.writeHeat(ctx, tx, in.Map, agg, day, now); err != nil {
		return nil, err
	}
	if rawKey != "" {
		if err := s.store.Put(ctx, rawKey, agg.rawBytes(), "application/gzip"); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		if rawKey != "" {
			_ = s.store.Delete(ctx, rawKey)
		}
		return nil, err
	}
	return map[string]any{"uploadId": in.UploadID, "part": in.Part, "duplicate": false, "accepted": agg.kept, "dropped": dropped, "players": len(agg.players), "cells": len(agg.cells)}, nil
}

func rawKeyFor(mapKey, id string) string {
	return "heatmaps/" + strings.ReplaceAll(mapKey, ":", "_") + "/" + id + ".csv.gz"
}

func (s *service) trace(ctx context.Context, key string, max int) ([]point, error) {
	obj, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()
	zr, err := gzip.NewReader(obj.Body)
	if err != nil {
		return nil, err
	}
	pts, _, err := readCSV(zr, 1<<30)
	if err != nil {
		return nil, err
	}
	if len(pts) > max {
		pts = pts[:max]
	}
	return pts, nil
}
