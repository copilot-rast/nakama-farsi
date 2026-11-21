package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/assets/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

var (
	packVersionPattern = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._+-]{0,39}$`)
	appVersionPattern  = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._+-]{0,31}$`)
	sha256Pattern      = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

const maxPathLen = 512

type packFileIn struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type pack struct {
	ID            string     `json:"id"`
	Platform      string     `json:"platform"`
	Version       string     `json:"version"`
	MinAppVersion string     `json:"minAppVersion"`
	MaxAppVersion string     `json:"maxAppVersion"`
	Status        string     `json:"status"`
	Active        bool       `json:"active"`
	CatalogPath   string     `json:"catalogPath"`
	HashPath      string     `json:"hashPath"`
	CatalogHash   string     `json:"catalogHash"`
	FileCount     int        `json:"fileCount"`
	TotalBytes    int64      `json:"totalBytes"`
	Note          string     `json:"note"`
	CreatedBy     string     `json:"createdBy"`
	CreatedAt     time.Time  `json:"createdAt"`
	CompletedAt   *time.Time `json:"completedAt"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	ActivatedAt   *time.Time `json:"activatedAt"`
	ActivatedBy   string     `json:"activatedBy"`
}

const packCols = `id::text, platform, version, min_app_version, max_app_version, status, active, catalog_path, hash_path, catalog_hash, file_count, total_bytes, note, created_by, created_at, completed_at, expires_at, activated_at, activated_by`

func scanPack(sc interface{ Scan(...any) error }) (pack, error) {
	var p pack
	var completed, expires, activated sql.NullTime
	err := sc.Scan(&p.ID, &p.Platform, &p.Version, &p.MinAppVersion, &p.MaxAppVersion, &p.Status, &p.Active, &p.CatalogPath, &p.HashPath, &p.CatalogHash, &p.FileCount, &p.TotalBytes, &p.Note, &p.CreatedBy, &p.CreatedAt, &completed, &expires, &activated, &p.ActivatedBy)
	if completed.Valid {
		t := completed.Time
		p.CompletedAt = &t
	}
	if expires.Valid {
		t := expires.Time
		p.ExpiresAt = &t
	}
	if activated.Valid {
		t := activated.Time
		p.ActivatedAt = &t
	}
	return p, err
}

func (s *service) loadPack(ctx context.Context, q queryer, id string, lock bool) (pack, error) {
	if _, err := uuid.FromString(id); err != nil {
		return pack{}, apierr.NotFound("PACK_NOT_FOUND")
	}
	query := `SELECT ` + packCols + ` FROM assets_pack WHERE id=$1`
	if lock {
		query += ` FOR UPDATE`
	}
	p, err := scanPack(q.QueryRowContext(ctx, query, id))
	if isNoRows(err) {
		return p, apierr.NotFound("PACK_NOT_FOUND")
	}
	return p, err
}

// cleanPackPath normalizes a file path inside a pack ("a\\b/./c" → "a/b/c"); ok is false for
// "..", empty paths and control characters. Paths that start with the platform folder lose it.
func cleanPackPath(raw, platform string) (string, bool) {
	p := strings.ReplaceAll(strings.TrimSpace(raw), `\`, "/")
	parts := []string{}
	for _, seg := range strings.Split(p, "/") {
		switch seg {
		case "", ".":
			continue
		case "..":
			return "", false
		}
		if badText(seg, 255) {
			return "", false
		}
		parts = append(parts, seg)
	}
	if len(parts) > 1 && parts[0] == platform {
		parts = parts[1:]
	}
	out := strings.Join(parts, "/")
	if out == "" || len(out) > maxPathLen {
		return "", false
	}
	return out, true
}

func junkFile(p string) bool {
	base := path.Base(p)
	return strings.HasPrefix(p, "__MACOSX/") || base == ".DS_Store" || base == "Thumbs.db"
}

// compareVersions orders app versions by their dot-separated parts: numbers by value, other
// parts as text ("1.10" > "1.9", "2.0" == "2").
func compareVersions(a, b string) int {
	split := func(v string) []string {
		return strings.FieldsFunc(v, func(r rune) bool { return r == '.' || r == '-' || r == '+' || r == '_' })
	}
	pa, pb := split(a), split(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		x, y := "0", "0"
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		nx, ex := strconv.ParseInt(x, 10, 64)
		ny, ey := strconv.ParseInt(y, 10, 64)
		switch {
		case ex == nil && ey == nil:
			if nx != ny {
				if nx < ny {
					return -1
				}
				return 1
			}
		case ex == nil:
			return 1 // a number is newer than a label ("1.0" > "1.beta")
		case ey == nil:
			return -1
		default:
			if c := strings.Compare(x, y); c != 0 {
				return c
			}
		}
	}
	return 0
}

// inRange reports whether app version v is inside [min, max] (an empty bound is open; an empty v matches all).
func inRange(v, min, max string) bool {
	if v == "" {
		return true
	}
	if min != "" && compareVersions(v, min) < 0 {
		return false
	}
	if max != "" && compareVersions(v, max) > 0 {
		return false
	}
	return true
}

func contentType(p string) string {
	switch strings.ToLower(path.Ext(p)) {
	case ".json":
		return "application/json"
	case ".hash", ".txt":
		return "text/plain; charset=utf-8"
	case ".xml":
		return "application/xml"
	}
	return "application/octet-stream"
}

func blobKey(sha string) string        { return "packs/blobs/" + sha }
func partKey(sha string, n int) string { return "packs/parts/" + sha + "/" + strconv.Itoa(n) }
func partsFor(size int64, partBytes int) int {
	if size <= 0 {
		return 1
	}
	return int((size + int64(partBytes) - 1) / int64(partBytes))
}

type packCreateIn struct {
	Platform      string       `json:"platform"`
	Version       string       `json:"version"`
	MinAppVersion string       `json:"minAppVersion"`
	MaxAppVersion string       `json:"maxAppVersion"`
	Note          string       `json:"note"`
	Files         []packFileIn `json:"files"`
	Activate      bool         `json:"activate"`
}

type missingFile struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	Size      int64  `json:"size"`
	PartBytes int    `json:"partBytes"`
	Parts     int    `json:"parts"`
	Received  []int  `json:"received"`
}

// createPack starts (or resumes) an upload: it records the file list and answers the files the
// server does not hold yet; files already stored for an earlier pack are reused.
func (s *service) createPack(ctx context.Context, in packCreateIn, actor string) (map[string]any, error) {
	cfg := s.settings(ctx).Packs
	in.Platform = strings.TrimSpace(in.Platform)
	in.Version = strings.TrimSpace(in.Version)
	in.MinAppVersion = strings.TrimSpace(in.MinAppVersion)
	in.MaxAppVersion = strings.TrimSpace(in.MaxAppVersion)
	in.Note = strings.TrimSpace(in.Note)
	switch {
	case in.Platform == "":
		return nil, apierr.BadRequest("PLATFORM_REQUIRED").WithField("platform")
	case !cfg.PlatformAllowed(in.Platform):
		return nil, apierr.BadRequest("PACK_PLATFORM_NOT_ALLOWED").WithField("platform").With("platforms", cfg.Platforms)
	case !packVersionPattern.MatchString(in.Version):
		return nil, apierr.BadRequest("PACK_INVALID").WithField("version")
	case in.MinAppVersion != "" && !appVersionPattern.MatchString(in.MinAppVersion):
		return nil, apierr.BadRequest("PACK_INVALID").WithField("minAppVersion")
	case in.MaxAppVersion != "" && !appVersionPattern.MatchString(in.MaxAppVersion):
		return nil, apierr.BadRequest("PACK_INVALID").WithField("maxAppVersion")
	case in.MinAppVersion != "" && in.MaxAppVersion != "" && compareVersions(in.MinAppVersion, in.MaxAppVersion) > 0:
		return nil, apierr.BadRequest("PACK_INVALID").WithField("maxAppVersion")
	case len([]rune(in.Note)) > 500:
		return nil, apierr.BadRequest("PACK_INVALID").WithField("note")
	case len(in.Files) == 0:
		return nil, apierr.BadRequest("PACK_INVALID").WithField("files")
	}
	files := []packFileIn{}
	seen := map[string]bool{}
	var total int64
	for i, f := range in.Files {
		p, ok := cleanPackPath(f.Path, in.Platform)
		if !ok {
			return nil, apierr.BadRequest("PACK_INVALID").WithField(fmt.Sprintf("files[%d].path", i))
		}
		if junkFile(p) {
			continue
		}
		sha := strings.ToLower(strings.TrimSpace(f.SHA256))
		if !sha256Pattern.MatchString(sha) {
			return nil, apierr.BadRequest("PACK_INVALID").WithField(fmt.Sprintf("files[%d].sha256", i))
		}
		if f.Size < 0 {
			return nil, apierr.BadRequest("PACK_INVALID").WithField(fmt.Sprintf("files[%d].size", i))
		}
		if f.Size > cfg.MaxFileBytes {
			return nil, apierr.BadRequest("PACK_FILE_TOO_LARGE").WithField(fmt.Sprintf("files[%d].size", i)).With("path", p).With("max", cfg.MaxFileBytes)
		}
		if seen[p] {
			return nil, apierr.BadRequest("PACK_INVALID").WithField(fmt.Sprintf("files[%d].path", i)).With("path", p)
		}
		seen[p] = true
		total += f.Size
		files = append(files, packFileIn{Path: p, SHA256: sha, Size: f.Size})
	}
	if len(files) > cfg.MaxFiles {
		return nil, apierr.BadRequest("PACK_TOO_MANY_FILES").With("max", cfg.MaxFiles)
	}
	if total > cfg.MaxPackBytes {
		return nil, apierr.BadRequest("PACK_TOO_LARGE").With("max", cfg.MaxPackBytes).With("size", total)
	}
	catalog, hashPath, err := findCatalog(files, cfg)
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	shaSize := map[string]int64{}
	for _, f := range files {
		if prev, ok := shaSize[f.SHA256]; ok && prev != f.Size {
			return nil, apierr.BadRequest("PACK_INVALID").WithField("files").With("sha256", f.SHA256)
		}
		shaSize[f.SHA256] = f.Size
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var id, status string
	err = tx.QueryRowContext(ctx, `SELECT id::text, status FROM assets_pack WHERE platform=$1 AND version=$2 FOR UPDATE`, in.Platform, in.Version).Scan(&id, &status)
	switch {
	case err == nil && status == "ready":
		return nil, apierr.Conflict("PACK_VERSION_EXISTS").With("packId", id).With("platform", in.Platform).With("version", in.Version)
	case err == nil:
		// Resume: the earlier upload of the same version is replaced by this file list.
		if _, err := tx.ExecContext(ctx, `DELETE FROM assets_pack_file WHERE pack_id=$1`, id); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE assets_pack SET min_app_version=$2, max_app_version=$3, note=$4, catalog_path=$5, hash_path=$6, file_count=$7, total_bytes=$8, created_by=$9, expires_at=now() + make_interval(hours => $10) WHERE id=$1`,
			id, in.MinAppVersion, in.MaxAppVersion, in.Note, catalog, hashPath, len(files), total, actor, cfg.UploadHours); err != nil {
			return nil, err
		}
	case isNoRows(err):
		id = uuid.Must(uuid.NewV4()).String()
		if _, err := tx.ExecContext(ctx, `INSERT INTO assets_pack (id, platform, version, min_app_version, max_app_version, status, catalog_path, hash_path, file_count, total_bytes, note, created_by, expires_at)
			VALUES ($1,$2,$3,$4,$5,'uploading',$6,$7,$8,$9,$10,$11, now() + make_interval(hours => $12))`,
			id, in.Platform, in.Version, in.MinAppVersion, in.MaxAppVersion, catalog, hashPath, len(files), total, in.Note, actor, cfg.UploadHours); err != nil {
			if uniqueViolation(err) {
				return nil, apierr.Conflict("PACK_VERSION_EXISTS").With("platform", in.Platform).With("version", in.Version)
			}
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO assets_pack_event (pack_id, platform, version, action, actor) VALUES ($1,$2,$3,'uploading',$4)`, id, in.Platform, in.Version, actor); err != nil {
			return nil, err
		}
	default:
		return nil, err
	}
	paths, shas, sizes := []string{}, []string{}, []int64{}
	for _, f := range files {
		paths, shas, sizes = append(paths, f.Path), append(shas, f.SHA256), append(sizes, f.Size)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO assets_pack_file (pack_id, path, sha256, size) SELECT $1, * FROM unnest($2::text[], $3::text[], $4::bigint[])`, id, paths, shas, sizes); err != nil {
		return nil, err
	}
	uniq := []string{}
	for sha := range shaSize {
		uniq = append(uniq, sha)
	}
	sort.Strings(uniq)
	for _, sha := range uniq {
		size := shaSize[sha]
		if _, err := tx.ExecContext(ctx, `INSERT INTO assets_blob (sha256, size, part_bytes, parts, content_type) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (sha256) DO NOTHING`,
			sha, size, cfg.PartBytes, partsFor(size, cfg.PartBytes), contentType(firstPath(files, sha))); err != nil {
			return nil, err
		}
	}
	missing, err := missingBlobs(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	p, err := s.loadPack(ctx, s.db, id, false)
	if err != nil {
		return nil, err
	}
	return map[string]any{"pack": p, "partBytes": cfg.PartBytes, "missing": missing}, nil
}

func firstPath(files []packFileIn, sha string) string {
	for _, f := range files {
		if f.SHA256 == sha {
			return f.Path
		}
	}
	return ""
}

// findCatalog picks the Addressables catalog (a root file matching the pattern) and its .hash.
func findCatalog(files []packFileIn, cfg settings.Packs) (string, string, error) {
	re, err := regexp.Compile(cfg.CatalogPattern)
	if err != nil {
		return "", "", err
	}
	catalogs := []string{}
	names := map[string]bool{}
	for _, f := range files {
		names[f.Path] = true
		if !strings.Contains(f.Path, "/") && re.MatchString(f.Path) {
			catalogs = append(catalogs, f.Path)
		}
	}
	if len(catalogs) == 0 {
		return "", "", apierr.BadRequest("PACK_CATALOG_MISSING").With("pattern", cfg.CatalogPattern)
	}
	if len(catalogs) > 1 {
		return "", "", apierr.BadRequest("PACK_CATALOG_AMBIGUOUS").With("catalogs", catalogs)
	}
	c := catalogs[0]
	h := strings.TrimSuffix(c, path.Ext(c)) + ".hash"
	if !names[h] {
		if cfg.RequireHash {
			return "", "", apierr.BadRequest("PACK_HASH_MISSING").With("path", h)
		}
		h = ""
	}
	return c, h, nil
}

func missingBlobs(ctx context.Context, q queryer, packID string) ([]missingFile, error) {
	rows, err := q.QueryContext(ctx, `SELECT f.path, f.sha256, b.size, b.part_bytes, b.parts, array_to_json(b.received)::text FROM assets_pack_file f JOIN assets_blob b ON b.sha256=f.sha256
		WHERE f.pack_id=$1 AND NOT b.stored ORDER BY f.path`, packID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []missingFile{}
	seen := map[string]bool{}
	for rows.Next() {
		var m missingFile
		var raw string
		if err := rows.Scan(&m.Path, &m.SHA256, &m.Size, &m.PartBytes, &m.Parts, &raw); err != nil {
			return nil, err
		}
		received := []int64{}
		_ = json.Unmarshal([]byte(raw), &received)
		if seen[m.SHA256] {
			continue
		}
		seen[m.SHA256] = true
		m.Received = []int{}
		for _, r := range received {
			m.Received = append(m.Received, int(r))
		}
		sort.Ints(m.Received)
		out = append(out, m)
	}
	return out, rows.Err()
}

// putPart stores one part of a file; the part that completes the file joins the parts, checks the
// hash and stores the file once.
func (s *service) putPart(ctx context.Context, packID, sha string, part int, body io.Reader) (map[string]any, error) {
	p, err := s.loadPack(ctx, s.db, packID, false)
	if err != nil {
		return nil, err
	}
	if p.Status != "uploading" {
		return nil, apierr.Conflict("PACK_NOT_UPLOADING")
	}
	var inPack bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM assets_pack_file WHERE pack_id=$1 AND sha256=$2)`, packID, sha).Scan(&inPack); err != nil {
		return nil, err
	}
	if !inPack {
		return nil, apierr.NotFound("PACK_FILE_UNKNOWN").With("sha256", sha)
	}
	var size int64
	var partBytes, parts int
	var stored bool
	if err := s.db.QueryRowContext(ctx, `SELECT size, part_bytes, parts, stored FROM assets_blob WHERE sha256=$1`, sha).Scan(&size, &partBytes, &parts, &stored); err != nil {
		if isNoRows(err) {
			return nil, apierr.NotFound("PACK_FILE_UNKNOWN").With("sha256", sha)
		}
		return nil, err
	}
	if stored {
		return map[string]any{"sha256": sha, "part": part, "parts": parts, "stored": true}, nil
	}
	if part < 0 || part >= parts {
		return nil, apierr.BadRequest("PACK_PART_INVALID").With("parts", parts)
	}
	want := int64(partBytes)
	if part == parts-1 {
		want = size - int64(parts-1)*int64(partBytes)
	}
	data, err := io.ReadAll(io.LimitReader(body, want+1))
	if err != nil {
		return nil, apierr.BadRequest("PACK_PART_INVALID").WithDetail("cannot read the part (it may be over the server's request size limit)").With("partBytes", partBytes)
	}
	if int64(len(data)) != want {
		return nil, apierr.BadRequest("PACK_PART_INVALID").With("expected", want).With("received", len(data)).With("partBytes", partBytes)
	}
	if parts == 1 {
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != sha {
			return nil, apierr.BadRequest("PACK_FILE_HASH_MISMATCH").With("sha256", sha)
		}
	}
	if err := s.store.Put(ctx, partKey(sha, part), data, "application/octet-stream"); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var received int
	if err := tx.QueryRowContext(ctx, `UPDATE assets_blob SET received = CASE WHEN $2::int = ANY(received) THEN received ELSE array_append(received, $2::int) END
		WHERE sha256=$1 AND NOT stored RETURNING cardinality(received)`, sha, part).Scan(&received); err != nil {
		if isNoRows(err) {
			return map[string]any{"sha256": sha, "part": part, "parts": parts, "stored": true}, nil
		}
		return nil, err
	}
	if received < parts {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return map[string]any{"sha256": sha, "part": part, "parts": parts, "received": received, "stored": false}, nil
	}
	// All parts are in: join them (this transaction holds the blob row, so only one request does it).
	var buf bytes.Buffer
	for i := 0; i < parts; i++ {
		chunk := data
		if i != part {
			obj, err := s.store.Get(ctx, partKey(sha, i))
			if err != nil {
				return nil, err
			}
			chunk, err = io.ReadAll(obj.Body)
			obj.Body.Close()
			if err != nil {
				return nil, err
			}
		}
		buf.Write(chunk)
	}
	sum := sha256.Sum256(buf.Bytes())
	if hex.EncodeToString(sum[:]) != sha || int64(buf.Len()) != size {
		_, _ = tx.ExecContext(ctx, `UPDATE assets_blob SET received='{}' WHERE sha256=$1`, sha)
		_ = tx.Commit()
		s.deleteParts(ctx, sha, parts)
		return nil, apierr.BadRequest("PACK_FILE_HASH_MISMATCH").With("sha256", sha)
	}
	var ctype string
	_ = tx.QueryRowContext(ctx, `SELECT content_type FROM assets_blob WHERE sha256=$1`, sha).Scan(&ctype)
	if err := s.store.Put(ctx, blobKey(sha), buf.Bytes(), ctype); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE assets_blob SET stored=true, stored_at=now() WHERE sha256=$1`, sha); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.deleteParts(ctx, sha, parts)
	return map[string]any{"sha256": sha, "part": part, "parts": parts, "received": parts, "stored": true}, nil
}

func (s *service) deleteParts(ctx context.Context, sha string, parts int) {
	for i := 0; i < parts; i++ {
		_ = s.store.Delete(ctx, partKey(sha, i))
	}
}

// completePack checks that every file is stored, reads the catalog hash and marks the pack ready.
func (s *service) completePack(ctx context.Context, packID string, activate bool, actor string) (pack, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return pack{}, err
	}
	defer func() { _ = tx.Rollback() }()
	p, err := s.loadPack(ctx, tx, packID, true)
	if err != nil {
		return pack{}, err
	}
	if p.Status != "uploading" {
		return pack{}, apierr.Conflict("PACK_NOT_UPLOADING")
	}
	missing, err := missingBlobs(ctx, tx, packID)
	if err != nil {
		return pack{}, err
	}
	if len(missing) > 0 {
		paths := []string{}
		for _, m := range missing {
			paths = append(paths, m.Path)
		}
		return pack{}, apierr.Conflict("PACK_FILES_MISSING").With("paths", paths).With("missing", missing)
	}
	catalogHash := ""
	if p.HashPath != "" {
		var sha string
		if err := tx.QueryRowContext(ctx, `SELECT sha256 FROM assets_pack_file WHERE pack_id=$1 AND path=$2`, packID, p.HashPath).Scan(&sha); err != nil {
			return pack{}, err
		}
		obj, err := s.store.Get(ctx, blobKey(sha))
		if err != nil {
			return pack{}, err
		}
		raw, _ := io.ReadAll(io.LimitReader(obj.Body, 256))
		obj.Body.Close()
		catalogHash = strings.TrimSpace(string(raw))
	}
	if _, err := tx.ExecContext(ctx, `UPDATE assets_pack SET status='ready', completed_at=now(), expires_at=NULL, catalog_hash=$2 WHERE id=$1`, packID, catalogHash); err != nil {
		return pack{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO assets_pack_event (pack_id, platform, version, action, actor) VALUES ($1,$2,$3,'ready',$4)`, packID, p.Platform, p.Version, actor); err != nil {
		return pack{}, err
	}
	if activate {
		if err := s.setActive(ctx, tx, p, true, actor); err != nil {
			return pack{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return pack{}, err
	}
	return s.loadPack(ctx, s.db, packID, false)
}

// setActive switches a ready pack on (activated now, so it wins the index for its range) or off.
func (s *service) setActive(ctx context.Context, tx queryer, p pack, on bool, actor string) error {
	cfg := s.settings(ctx).Packs
	if on {
		if cfg.SingleActive {
			rows, err := tx.QueryContext(ctx, `SELECT `+packCols+` FROM assets_pack WHERE platform=$1 AND active AND id<>$2 FOR UPDATE`, p.Platform, p.ID)
			if err != nil {
				return err
			}
			others := []pack{}
			for rows.Next() {
				o, err := scanPack(rows)
				if err != nil {
					rows.Close()
					return err
				}
				others = append(others, o)
			}
			rows.Close()
			for _, o := range others {
				if rangesOverlap(o, p) {
					if _, err := tx.ExecContext(ctx, `UPDATE assets_pack SET active=false WHERE id=$1`, o.ID); err != nil {
						return err
					}
					if _, err := tx.ExecContext(ctx, `INSERT INTO assets_pack_event (pack_id, platform, version, action, actor) VALUES ($1,$2,$3,'deactivated',$4)`, o.ID, o.Platform, o.Version, actor); err != nil {
						return err
					}
				}
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE assets_pack SET active=true, activated_at=now(), activated_by=$2 WHERE id=$1`, p.ID, actor); err != nil {
			return err
		}
	} else if _, err := tx.ExecContext(ctx, `UPDATE assets_pack SET active=false WHERE id=$1`, p.ID); err != nil {
		return err
	}
	action := "activated"
	if !on {
		action = "deactivated"
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO assets_pack_event (pack_id, platform, version, action, actor) VALUES ($1,$2,$3,$4,$5)`, p.ID, p.Platform, p.Version, action, actor)
	return err
}

func rangesOverlap(a, b pack) bool {
	lowA, highA, lowB, highB := a.MinAppVersion, a.MaxAppVersion, b.MinAppVersion, b.MaxAppVersion
	if highA != "" && lowB != "" && compareVersions(highA, lowB) < 0 {
		return false
	}
	if highB != "" && lowA != "" && compareVersions(highB, lowA) < 0 {
		return false
	}
	return true
}

func (s *service) activate(ctx context.Context, id string, on bool, actor string) (pack, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return pack{}, err
	}
	defer func() { _ = tx.Rollback() }()
	p, err := s.loadPack(ctx, tx, id, true)
	if err != nil {
		return pack{}, err
	}
	if p.Status != "ready" {
		return pack{}, apierr.Conflict("PACK_NOT_READY")
	}
	if err := s.setActive(ctx, tx, p, on, actor); err != nil {
		return pack{}, err
	}
	if err := tx.Commit(); err != nil {
		return pack{}, err
	}
	return s.loadPack(ctx, s.db, id, false)
}

func (s *service) deletePack(ctx context.Context, id, actor string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	p, err := s.loadPack(ctx, tx, id, true)
	if err != nil {
		return err
	}
	if p.Active {
		return apierr.Conflict("PACK_ACTIVE")
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM assets_pack WHERE id=$1`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO assets_pack_event (pack_id, platform, version, action, actor) VALUES ($1,$2,$3,'deleted',$4)`, p.ID, p.Platform, p.Version, actor); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.packFile.drop(id)
	return nil
}

// ---- content index ----

func (s *service) publicBase(r *http.Request) string {
	if b := strings.TrimRight(s.settings(r.Context()).Packs.BaseURL, "/"); b != "" {
		return b
	}
	if s.baseURL != "" {
		return s.baseURL
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func (s *service) indexBase(ctx context.Context) string {
	if b := strings.TrimRight(s.settings(ctx).Packs.BaseURL, "/"); b != "" {
		return b
	}
	return s.baseURL
}

// latestPack is the active pack of the platform activated last whose app version range holds appVersion.
func (s *service) latestPack(ctx context.Context, platform, appVersion string) (*pack, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+packCols+` FROM assets_pack WHERE platform=$1 AND active AND status='ready' ORDER BY activated_at DESC, created_at DESC`, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanPack(rows)
		if err != nil {
			return nil, err
		}
		if inRange(appVersion, p.MinAppVersion, p.MaxAppVersion) {
			return &p, nil
		}
	}
	return nil, rows.Err()
}

func indexView(p *pack, base string) any {
	if p == nil {
		return nil
	}
	root := base + publicPrefix + "/content/" + p.ID
	out := map[string]any{"id": p.ID, "platform": p.Platform, "version": p.Version, "minAppVersion": p.MinAppVersion, "maxAppVersion": p.MaxAppVersion,
		"baseUrl": root, "catalogUrl": root + "/" + p.CatalogPath, "catalogPath": p.CatalogPath, "hashUrl": "", "catalogHash": p.CatalogHash,
		"fileCount": p.FileCount, "totalBytes": p.TotalBytes, "activatedAt": p.ActivatedAt}
	if p.HashPath != "" {
		out["hashUrl"] = root + "/" + p.HashPath
	}
	return out
}

type indexIn struct {
	Platform   string `json:"platform"`
	AppVersion string `json:"appVersion"`
}

func (in *indexIn) check() error {
	in.Platform, in.AppVersion = strings.TrimSpace(in.Platform), strings.TrimSpace(in.AppVersion)
	if in.Platform == "" {
		return apierr.BadRequest("PLATFORM_REQUIRED").WithField("platform")
	}
	if !settings.PlatformPattern.MatchString(in.Platform) {
		return apierr.BadRequest("PACK_PLATFORM_NOT_ALLOWED").WithField("platform")
	}
	if in.AppVersion != "" && !appVersionPattern.MatchString(in.AppVersion) {
		return apierr.BadRequest("PACK_INVALID").WithField("appVersion")
	}
	return nil
}

func (s *service) rpcContentIndex(ctx context.Context, _ string, payload string) (any, error) {
	var in indexIn
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if err := in.check(); err != nil {
		return nil, err
	}
	p, err := s.latestPack(ctx, in.Platform, in.AppVersion)
	if err != nil {
		return nil, err
	}
	return map[string]any{"platform": in.Platform, "appVersion": in.AppVersion, "pack": indexView(p, s.indexBase(ctx))}, nil
}

func (s *service) httpContentIndex(w http.ResponseWriter, r *http.Request) {
	cfg := s.settings(r.Context()).Packs
	if !cfg.PublicIndex {
		s.httpFail(w, r, apierr.Forbidden("CONTENT_INDEX_PRIVATE"))
		return
	}
	in := indexIn{Platform: r.URL.Query().Get("platform"), AppVersion: r.URL.Query().Get("appVersion")}
	if err := in.check(); err != nil {
		s.httpFail(w, r, err)
		return
	}
	p, err := s.latestPack(r.Context(), in.Platform, in.AppVersion)
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	body, _ := json.Marshal(map[string]any{"platform": in.Platform, "appVersion": in.AppVersion, "pack": indexView(p, s.publicBase(r))})
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(cfg.IndexCacheSeconds))
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(body, '\n'))
}

// ---- serving pack files ----

type cachedFile struct {
	sha  string
	size int64
}

// fileCache keeps the file lists of ready packs for a minute (a pack deleted on another node stops
// being served here within that time).
type fileCache struct {
	mu    sync.Mutex
	max   int
	ttl   time.Duration
	order []string
	packs map[string]map[string]cachedFile
	at    map[string]time.Time
}

func newFileCache(max int) *fileCache {
	return &fileCache{max: max, ttl: time.Minute, packs: map[string]map[string]cachedFile{}, at: map[string]time.Time{}}
}

func (c *fileCache) get(id string) (map[string]cachedFile, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.packs[id]
	if ok && time.Since(c.at[id]) > c.ttl {
		return nil, false
	}
	return m, ok
}

func (c *fileCache) put(id string, m map[string]cachedFile) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.packs[id]; !ok {
		c.order = append(c.order, id)
		if len(c.order) > c.max {
			delete(c.packs, c.order[0])
			delete(c.at, c.order[0])
			c.order = c.order[1:]
		}
	}
	c.packs[id] = m
	c.at[id] = time.Now()
}

func (c *fileCache) drop(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.packs, id)
}

// packFiles lists the files of a ready pack (cached: a ready pack never changes).
func (s *service) packFiles(ctx context.Context, id string) (map[string]cachedFile, error) {
	if m, ok := s.packFile.get(id); ok {
		return m, nil
	}
	if _, err := uuid.FromString(id); err != nil {
		return nil, apierr.NotFound("PACK_NOT_FOUND")
	}
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM assets_pack WHERE id=$1`, id).Scan(&status); err != nil || status != "ready" {
		return nil, apierr.NotFound("PACK_NOT_FOUND")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT path, sha256, size FROM assets_pack_file WHERE pack_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]cachedFile{}
	for rows.Next() {
		var p string
		var f cachedFile
		if err := rows.Scan(&p, &f.sha, &f.size); err != nil {
			return nil, err
		}
		m[p] = f
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.packFile.put(id, m)
	return m, nil
}

func (s *service) httpContentFile(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "pack")
	p, ok := cleanPackPath(pluginkit.PathVar(r, "path"), "")
	if !ok {
		http.NotFound(w, r)
		return
	}
	files, err := s.packFiles(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	f, ok := files[p]
	if !ok {
		http.NotFound(w, r)
		return
	}
	cache := "public, max-age=31536000, immutable"
	if ext := strings.ToLower(path.Ext(p)); ext == ".json" || ext == ".hash" || ext == ".bin" {
		cache = "public, max-age=300"
	}
	etag := `"` + f.sha + `"`
	cfg := s.settings(r.Context()).Packs
	if cfg.RedirectToStorage {
		if target := s.store.PublicURL(blobKey(f.sha)); target != "" {
			w.Header().Set("Cache-Control", cache)
			http.Redirect(w, r, target, http.StatusFound)
			return
		}
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", cache)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	obj, err := s.store.Get(r.Context(), blobKey(f.sha))
	if err != nil {
		if !errors.Is(err, objectstore.ErrNotFound) && s.logger != nil {
			s.logger.Warn("Assets content file %s/%s from %s storage failed: %v", id, p, s.store.Driver(), err)
		}
		http.NotFound(w, r)
		return
	}
	defer obj.Body.Close()
	w.Header().Set("Content-Type", contentType(p))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if rs, ok := obj.Body.(io.ReadSeeker); ok {
		http.ServeContent(w, r, path.Base(p), obj.ModTime, rs)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(f.size, 10))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = io.Copy(w, obj.Body)
	}
}

// ---- retention ----

func (s *service) sweepPacks(ctx context.Context, cfg settings.All) error {
	if _, err := s.db.ExecContext(ctx, `WITH d AS (DELETE FROM assets_pack WHERE status='uploading' AND expires_at < now() RETURNING id, platform, version)
		INSERT INTO assets_pack_event (pack_id, platform, version, action, actor) SELECT id, platform, version, 'expired', 'system' FROM d`); err != nil {
		return err
	}
	if k := cfg.Packs.KeepInactive; k > 0 {
		if _, err := s.db.ExecContext(ctx, `WITH old AS (
				SELECT id FROM (SELECT id, row_number() OVER (PARTITION BY platform ORDER BY coalesce(activated_at, completed_at, created_at) DESC) AS rn
				FROM assets_pack WHERE status='ready' AND NOT active) x WHERE rn > $1),
			d AS (DELETE FROM assets_pack WHERE id IN (SELECT id FROM old) RETURNING id, platform, version)
			INSERT INTO assets_pack_event (pack_id, platform, version, action, actor) SELECT id, platform, version, 'removed', 'system' FROM d`, k); err != nil {
			return err
		}
	}
	return s.collectBlobs(ctx, time.Duration(cfg.Packs.UploadHours)*time.Hour)
}

// collectBlobs removes stored files and unfinished parts that no pack lists any more.
func (s *service) collectBlobs(ctx context.Context, grace time.Duration) error {
	rows, err := s.db.QueryContext(ctx, `SELECT sha256, parts, stored FROM assets_blob b WHERE created_at < $1
		AND NOT EXISTS (SELECT 1 FROM assets_pack_file f WHERE f.sha256=b.sha256) LIMIT 500`, s.now().Add(-grace))
	if err != nil {
		return err
	}
	type blob struct {
		sha    string
		parts  int
		stored bool
	}
	list := []blob{}
	for rows.Next() {
		var b blob
		if err := rows.Scan(&b.sha, &b.parts, &b.stored); err != nil {
			rows.Close()
			return err
		}
		list = append(list, b)
	}
	rows.Close()
	for _, b := range list {
		res, err := s.db.ExecContext(ctx, `DELETE FROM assets_blob b WHERE sha256=$1 AND NOT EXISTS (SELECT 1 FROM assets_pack_file f WHERE f.sha256=b.sha256)`, b.sha)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			continue
		}
		if b.stored {
			_ = s.store.Delete(ctx, blobKey(b.sha))
		}
		s.deleteParts(ctx, b.sha, b.parts)
	}
	return nil
}

// ---- editor endpoints ----

func (s *service) editorPackList(w http.ResponseWriter, r *http.Request, _ *apikey.Key) {
	list, err := s.listPacks(r.Context(), r.URL.Query().Get("platform"), r.URL.Query().Get("status"), 50)
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list}, true)
}

func (s *service) editorPackCreate(w http.ResponseWriter, r *http.Request, key *apikey.Key) {
	var in packCreateIn
	if err := readJSON(r, &in, 8<<20); err != nil {
		s.httpFail(w, r, err)
		return
	}
	if in.Activate {
		if err := s.activateDenied(r.Context(), key); err != nil {
			s.httpFail(w, r, err)
			return
		}
	}
	out, err := s.createPack(r.Context(), in, actorOf(key))
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out, true)
}

func (s *service) editorPackPart(w http.ResponseWriter, r *http.Request, _ *apikey.Key) {
	sha := strings.ToLower(pluginkit.PathVar(r, "sha"))
	part, err := strconv.Atoi(pluginkit.PathVar(r, "part"))
	if !sha256Pattern.MatchString(sha) || err != nil {
		s.httpFail(w, r, apierr.BadRequest("PACK_PART_INVALID"))
		return
	}
	out, err := s.putPart(r.Context(), pluginkit.PathVar(r, "id"), sha, part, r.Body)
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out, true)
}

func (s *service) activateDenied(ctx context.Context, key *apikey.Key) error {
	if !s.settings(ctx).Packs.ActivateFromEditor {
		return apierr.Forbidden("PACK_ACTIVATE_DENIED").With("reason", "settings")
	}
	if !apikey.Allows(key.Scopes, ScopePublish, true) {
		return apierr.Forbidden("PACK_ACTIVATE_DENIED").With("scope", ScopePublish)
	}
	return nil
}

func (s *service) editorPackComplete(w http.ResponseWriter, r *http.Request, key *apikey.Key) {
	var in struct {
		Activate bool `json:"activate"`
	}
	if err := readJSON(r, &in, 4096); err != nil {
		s.httpFail(w, r, err)
		return
	}
	if in.Activate {
		if err := s.activateDenied(r.Context(), key); err != nil {
			s.httpFail(w, r, err)
			return
		}
	}
	p, err := s.completePack(r.Context(), pluginkit.PathVar(r, "id"), in.Activate, actorOf(key))
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"pack": p}, true)
}

func (s *service) listPacks(ctx context.Context, platform, status string, limit int) ([]pack, error) {
	args := []any{}
	where := "true"
	if platform != "" {
		args = append(args, platform)
		where += " AND platform=$" + strconv.Itoa(len(args))
	}
	switch status {
	case "uploading", "ready":
		args = append(args, status)
		where += " AND status=$" + strconv.Itoa(len(args))
	case "active":
		where += " AND active"
	}
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+packCols+` FROM assets_pack WHERE `+where+` ORDER BY platform, created_at DESC LIMIT $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []pack{}
	for rows.Next() {
		p, err := scanPack(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
