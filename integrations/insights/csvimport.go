package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	systemUserID   = "00000000-0000-0000-0000-000000000000"
	importDeadline = 30 * time.Minute
)

type importOpts struct {
	Race       string
	Label      string
	Player     string
	PlayerName string
	Replace    bool
	Name       string
	CellSize   float64
	Create     bool
}

type importFile struct {
	Name      string         `json:"name"`
	Status    string         `json:"status"`
	Rows      int            `json:"rows"`
	Dropped   int            `json:"dropped"`
	Reasons   map[string]int `json:"reasons"`
	Points    int            `json:"points"`
	Cells     int            `json:"cells"`
	Players   int            `json:"players"`
	Label     string         `json:"label"`
	UploadID  string         `json:"uploadId,omitempty"`
	Duplicate bool           `json:"duplicate"`
	Error     string         `json:"error,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

type importer struct {
	s       *service
	ctx     context.Context
	cfg     settings.All
	opts    importOpts
	mapKey  string
	size    float64
	by      string
	files   []importFile
	maxRows int
	maxB    int64
}

func (o *importOpts) set(name, v string) error {
	v = strings.TrimSpace(v)
	switch name {
	case "race":
		if v != "" && v != "file" && v != "label" {
			return pluginkit.Bad("INVALID_POLICY", "race", "")
		}
		o.Race = v
	case "label":
		o.Label = cleanLabel(v, 100)
	case "player":
		if v != "" && v != "file" && v != "fixed" {
			return pluginkit.Bad("INVALID_POLICY", "player", "")
		}
		o.Player = v
	case "playerName":
		o.PlayerName = cleanLabel(v, 32)
	case "mode":
		if v != "" && v != "add" && v != "replace" {
			return pluginkit.Bad("INVALID_POLICY", "mode", "")
		}
		o.Replace = v == "replace"
	case "name":
		o.Name = cleanLabel(v, 100)
	case "create":
		o.Create = v == "true" || v == "1"
	case "cellSize":
		if v == "" {
			return nil
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0.1 || f > 1000 {
			return pluginkit.Bad("VALUE_OUT_OF_RANGE", "cellSize", "")
		}
		o.CellSize = f
	}
	return nil
}

func (o importOpts) check() error {
	if o.Race == "label" && o.Label == "" {
		return pluginkit.Bad("FIELD_REQUIRED", "label", "")
	}
	if o.Player == "fixed" && o.PlayerName == "" {
		return pluginkit.Bad("FIELD_REQUIRED", "playerName", "")
	}
	return nil
}

func (s *service) importCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rc := http.NewResponseController(w)
	_ = rc.SetReadDeadline(time.Now().Add(importDeadline))
	_ = rc.SetWriteDeadline(time.Now().Add(importDeadline))
	key := pluginkit.PathVar(r, "key")
	if !mapKeyPattern.MatchString(key) {
		s.fail(w, pluginkit.Bad("HEATMAP_INVALID", "key", ""))
		return
	}
	mr, err := r.MultipartReader()
	if err != nil {
		s.fail(w, pluginkit.Bad("INVALID_UPLOAD", "file", "send multipart/form-data with option fields first and one or more file parts"))
		return
	}
	cfg := s.settings(ctx)
	im := &importer{s: s, ctx: ctx, cfg: cfg, mapKey: key, by: pluginkit.ConsoleUser(r), files: []importFile{},
		maxRows: cfg.Heatmap.ImportMaxRows, maxB: int64(cfg.Heatmap.ImportMaxMB) << 20}
	ready := false
	replaced := false
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			s.fail(w, pluginkit.Bad("INVALID_UPLOAD", "file", "the multipart body is broken"))
			return
		}
		if part.FileName() == "" {
			raw, _ := io.ReadAll(io.LimitReader(part, 4096))
			_ = part.Close()
			if ready {
				continue
			}
			if err := im.opts.set(part.FormName(), string(raw)); err != nil {
				s.fail(w, err)
				return
			}
			continue
		}
		if !ready {
			if err := im.opts.check(); err != nil {
				_ = part.Close()
				s.fail(w, err)
				return
			}
			if err := im.prepare(); err != nil {
				_ = part.Close()
				s.fail(w, err)
				return
			}
			if im.opts.Replace {
				if err := s.clearMap(ctx, key); err != nil {
					_ = part.Close()
					s.fail(w, err)
					return
				}
				replaced = true
			}
			ready = true
		}
		im.part(part)
		_ = part.Close()
	}
	if !ready {
		s.fail(w, pluginkit.Bad("FILE_REQUIRED", "file", "add at least one CSV file"))
		return
	}
	m, err := s.loadMap(ctx, key, cfg.Heatmap.CellSize)
	if err != nil {
		s.fail(w, err)
		return
	}
	tot := map[string]int{"files": len(im.files)}
	for _, f := range im.files {
		tot[f.Status]++
		tot["rows"] += f.Rows
		tot["dropped"] += f.Dropped
		tot["points"] += f.Points
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"map": m, "replaced": replaced, "files": im.files, "totals": tot,
		"limits": map[string]any{"maxMB": cfg.Heatmap.ImportMaxMB, "maxRows": cfg.Heatmap.ImportMaxRows, "maxFiles": cfg.Heatmap.ImportMaxFiles}})
}

func (im *importer) prepare() error {
	ctx, db := im.ctx, im.s.db
	var size float64
	err := db.QueryRowContext(ctx, `SELECT cell_size FROM insights_heatmap_map WHERE key=$1`, im.mapKey).Scan(&size)
	if errors.Is(err, sql.ErrNoRows) {
		if !im.opts.Create {
			return pluginkit.Missing("HEATMAP_MAP_NOT_FOUND")
		}
		name := im.opts.Name
		if name == "" {
			name = im.mapKey
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO insights_heatmap_map (key, name, cell_size) VALUES ($1,$2,$3) ON CONFLICT (key) DO NOTHING`, im.mapKey, name, im.opts.CellSize); err != nil {
			return err
		}
		size = im.opts.CellSize
	} else if err != nil {
		return err
	}
	if size <= 0 {
		size = im.cfg.Heatmap.CellSize
	}
	im.size = size
	return nil
}

func (s *service) clearMap(ctx context.Context, key string) error {
	s.dropRaw(ctx, `SELECT id::text, raw_key FROM insights_heatmap_upload WHERE map_key=$1 AND raw_key <> ''`, key)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, q := range []string{
		`DELETE FROM insights_heatmap_upload WHERE map_key=$1`,
		`DELETE FROM insights_heatmap_cell WHERE map_key=$1`,
		`UPDATE insights_heatmap_map SET uploads=0, points=0, last_upload_at=NULL, min_x=NULL, max_x=NULL, min_z=NULL, max_z=NULL, updated_at=now() WHERE key=$1`,
	} {
		if _, err := tx.ExecContext(ctx, q, key); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func baseName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	return cleanLabel(path.Base(name), 200)
}

func labelFor(name string) string {
	b := path.Base(strings.ReplaceAll(name, "\\", "/"))
	for _, ext := range []string{".gz", ".csv", ".txt"} {
		if strings.HasSuffix(strings.ToLower(b), ext) {
			b = b[:len(b)-len(ext)]
		}
	}
	return cleanLabel(b, 100)
}

func (im *importer) fail(name string, err error) {
	f := importFile{Name: name, Status: "failed", Reasons: map[string]int{}, Error: "HEATMAP_INVALID"}
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		f.Error, f.Meta = e.Code, e.Meta
	} else if im.s.logger != nil {
		im.s.logger.Error("Heatmap CSV import of %s failed: %v", name, err)
		f.Error = "INTERNAL_ERROR"
	}
	im.files = append(im.files, f)
}

func (im *importer) full(name string) bool {
	if len(im.files) < im.cfg.Heatmap.ImportMaxFiles {
		return false
	}
	im.files = append(im.files, importFile{Name: name, Status: "skipped", Reasons: map[string]int{}, Error: "TOO_MANY_FILES", Meta: map[string]any{"maxFiles": im.cfg.Heatmap.ImportMaxFiles}})
	return true
}

func (im *importer) part(p *multipart.Part) {
	name := baseName(p.FileName())
	if im.full(name) {
		return
	}
	br := bufio.NewReaderSize(p, 4096)
	head, _ := br.Peek(4)
	if bytes.HasPrefix(head, []byte("PK\x03\x04")) || strings.HasSuffix(strings.ToLower(name), ".zip") {
		im.zip(name, br)
		return
	}
	im.file(name, br)
}

func (im *importer) zip(name string, r io.Reader) {
	tmp, err := os.CreateTemp("", "heatmap-*.zip")
	if err != nil {
		im.fail(name, err)
		return
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()
	n, err := io.Copy(tmp, io.LimitReader(r, im.maxB+1))
	if err != nil {
		im.fail(name, apierr.BadRequest("INVALID_UPLOAD").WithField("file"))
		return
	}
	if n > im.maxB {
		im.fail(name, apierr.New(413, "HEATMAP_TOO_LARGE").With("maxMB", im.cfg.Heatmap.ImportMaxMB))
		return
	}
	zr, err := zip.NewReader(tmp, n)
	if err != nil {
		im.fail(name, apierr.BadRequest("HEATMAP_INVALID").WithField("file").WithDetail("not a zip file"))
		return
	}
	found := 0
	for _, e := range zr.File {
		lower := strings.ToLower(e.Name)
		base := path.Base(strings.ReplaceAll(e.Name, "\\", "/"))
		if e.FileInfo().IsDir() || strings.HasPrefix(lower, "__macosx/") || strings.HasPrefix(base, ".") {
			continue
		}
		if !strings.HasSuffix(lower, ".csv") && !strings.HasSuffix(lower, ".csv.gz") && !strings.HasSuffix(lower, ".txt") {
			continue
		}
		found++
		entry := name + "/" + cleanLabel(e.Name, 200)
		if im.full(entry) {
			return
		}
		rc, err := e.Open()
		if err != nil {
			im.fail(entry, apierr.BadRequest("HEATMAP_INVALID").WithField("file"))
			continue
		}
		im.file(entry, bufio.NewReaderSize(rc, 4096))
		_ = rc.Close()
	}
	if found == 0 {
		im.fail(name, apierr.BadRequest("HEATMAP_INVALID").WithField("file").WithDetail("the zip has no .csv files"))
	}
}

type limitErr struct{}

func (limitErr) Error() string { return "too large" }

type capReader struct {
	r    io.Reader
	left int64
}

func (c *capReader) Read(p []byte) (int, error) {
	if c.left <= 0 {
		return 0, limitErr{}
	}
	if int64(len(p)) > c.left {
		p = p[:c.left]
	}
	n, err := c.r.Read(p)
	c.left -= int64(n)
	return n, err
}

func (im *importer) file(name string, br *bufio.Reader) {
	if im.full(name) {
		return
	}
	var src io.Reader = br
	if head, _ := br.Peek(2); len(head) == 2 && head[0] == 0x1f && head[1] == 0x8b {
		zr, err := gzip.NewReader(br)
		if err != nil {
			im.fail(name, apierr.BadRequest("HEATMAP_INVALID").WithField("file").WithDetail("data is not gzip"))
			return
		}
		defer zr.Close()
		src = zr
	}
	h := sha256.New()
	hc := im.cfg.Heatmap
	agg := newHeatAgg(im.size, hc.KeepRawDays > 0)
	tooMany := apierr.New(413, "HEATMAP_TOO_LARGE").With("maxRows", im.maxRows)
	fixed := ""
	if im.opts.Player == "fixed" {
		fixed = im.opts.PlayerName
	}
	rng, seen := 0, 0
	st, err := scanCSV(io.TeeReader(&capReader{r: src, left: im.maxB}, h), func(p point) error {
		if seen++; seen > im.maxRows {
			return tooMany
		}
		if fixed != "" {
			p.Player = fixed
		}
		if !validPoint(p, hc.MaxKmh) {
			rng++
			return nil
		}
		agg.add(p)
		return nil
	})
	if errors.As(err, new(limitErr)) {
		err = apierr.New(413, "HEATMAP_TOO_LARGE").With("maxMB", hc.ImportMaxMB)
	}
	if err == nil && st.Rows > im.maxRows {
		err = tooMany
	}
	if err != nil {
		im.fail(name, err)
		return
	}
	if st.Reasons == nil {
		st.Reasons = map[string]int{}
	}
	if rng > 0 {
		st.Dropped += rng
		st.Reasons[dropRange] = rng
	}
	label := labelFor(name)
	if im.opts.Race == "label" {
		label = im.opts.Label
	}
	out := importFile{Name: name, Rows: st.Rows, Dropped: st.Dropped, Reasons: st.Reasons, Label: label}
	if agg.kept == 0 {
		out.Status, out.Error = "failed", "HEATMAP_INVALID"
		im.files = append(im.files, out)
		return
	}
	sum := hex.EncodeToString(h.Sum(nil))
	id, dup, err := im.store(name, label, sum, agg, st.Dropped)
	if err != nil {
		im.fail(name, err)
		return
	}
	out.Points, out.Cells, out.Players = agg.kept, len(agg.cells), len(agg.players)
	if dup {
		out.Status, out.Duplicate, out.Points, out.Cells = "duplicate", true, 0, 0
	} else {
		out.Status, out.UploadID = "stored", id
	}
	im.files = append(im.files, out)
}

func (im *importer) store(name, label, sum string, agg *heatAgg, dropped int) (string, bool, error) {
	ctx, s := im.ctx, im.s
	var prior string
	err := s.db.QueryRowContext(ctx, `SELECT id::text FROM insights_heatmap_upload WHERE map_key=$1 AND content_hash=$2`, im.mapKey, sum).Scan(&prior)
	if err == nil {
		return prior, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	now := s.now().UTC()
	id := uuid.Must(uuid.NewV4()).String()
	rawKey := ""
	if im.cfg.Heatmap.KeepRawDays > 0 {
		rawKey = rawKeyFor(im.mapKey, id)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO insights_heatmap_upload (id, user_id, map_key, upload_key, part, points, dropped, players, duration_ms, raw_key, created_at, source, label, file_name, content_hash, uploaded_by)
		VALUES ($1,$2,$3,$4,0,$5,$6,$7,$8,$9,$10,'console',$11,$12,$13,$14)`,
		id, systemUserID, im.mapKey, "csv-"+strings.ReplaceAll(id, "-", ""), agg.kept, dropped, len(agg.players), agg.durationMs(), rawKey, now, label, name, sum, im.by); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return "", true, nil
		}
		return "", false, err
	}
	day := now.In(im.cfg.Location()).Format(time.DateOnly)
	if err := s.writeHeat(ctx, tx, im.mapKey, agg, day, now); err != nil {
		return "", false, err
	}
	if rawKey != "" {
		if err := s.store.Put(ctx, rawKey, agg.rawBytes(), "application/gzip"); err != nil {
			return "", false, err
		}
	}
	if err := tx.Commit(); err != nil {
		if rawKey != "" {
			_ = s.store.Delete(ctx, rawKey)
		}
		return "", false, err
	}
	return id, false, nil
}
