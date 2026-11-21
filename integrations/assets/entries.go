package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/heroiclabs/nakama/v3/integrations/assets/canon"
	"github.com/heroiclabs/nakama/v3/integrations/assets/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/itemkinds"
	"github.com/jackc/pgx/v5/pgconn"
)

var assetIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)

const (
	maxGUID   = 64
	maxType   = 255
	maxName   = 255
	maxFolder = 512
)

type entryIn struct {
	AssetID     string          `json:"assetId"`
	SourceGUID  string          `json:"sourceGuid"`
	Type        string          `json:"type"`
	Folder      string          `json:"folder"`
	Name        string          `json:"name"`
	Data        json.RawMessage `json:"data"`
	BaseVersion *int            `json:"baseVersion"`

	doc canon.Doc
}

type deleteIn struct {
	AssetID     string `json:"assetId"`
	BaseVersion *int   `json:"baseVersion"`
}

type pushIn struct {
	Items   []entryIn  `json:"items"`
	Deletes []deleteIn `json:"deletes"`
	DryRun  bool       `json:"dryRun"`
	Force   bool       `json:"force"`
}

type entry struct {
	AssetID    string          `json:"assetId"`
	SourceGUID string          `json:"sourceGuid"`
	Type       string          `json:"type"`
	Folder     string          `json:"folder"`
	Name       string          `json:"name"`
	Data       json.RawMessage `json:"data,omitempty"`
	Hash       string          `json:"hash"`
	Size       int             `json:"size"`
	Version    int             `json:"version"`
	UpdatedBy  string          `json:"updatedBy"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// manifestItem is what a client compares with its cache.
type manifestItem struct {
	AssetID    string `json:"assetId"`
	SourceGUID string `json:"sourceGuid,omitempty"`
	Type       string `json:"type"`
	Folder     string `json:"folder"`
	Name       string `json:"name"`
	Version    int    `json:"version"`
	Hash       string `json:"hash"`
}

type pushRow struct {
	AssetID         string `json:"assetId"`
	Action          string `json:"action"`
	Version         *int   `json:"version"`
	PreviousVersion *int   `json:"previousVersion,omitempty"`
	BaseVersion     *int   `json:"baseVersion,omitempty"`
	Hash            string `json:"hash,omitempty"`
}

type conflict struct {
	AssetID     string `json:"assetId"`
	Version     *int   `json:"version"`
	BaseVersion *int   `json:"baseVersion"`
}

const (
	actCreated   = "created"
	actUpdated   = "updated"
	actUnchanged = "unchanged"
	actDeleted   = "deleted"
	actConflict  = "conflict"
	actMissing   = "missing"
)

func intp(v int) *int { return &v }

func badText(s string, max int) bool {
	if len(s) > max {
		return true
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// cleanFolder turns "\Cars\BMW/" into "Cars/BMW"; ok is false for empty, "." or ".." segments.
func cleanFolder(raw string) (string, bool) {
	f := strings.Trim(strings.TrimSpace(strings.ReplaceAll(raw, `\`, "/")), "/")
	if f == "" {
		return "", true
	}
	if len(f) > maxFolder {
		return "", false
	}
	for _, seg := range strings.Split(f, "/") {
		if seg == "" || seg == "." || seg == ".." || badText(seg, maxFolder) || strings.TrimSpace(seg) != seg {
			return "", false
		}
	}
	return f, true
}

func invalid(i int, assetID, field, why string) error {
	return apierr.BadRequest("ASSET_INVALID").WithField(fmt.Sprintf("items[%d].%s", i, field)).With("assetId", assetID).WithDetail(why)
}

// cleanItems checks and normalizes pushed items; the data document is parsed once here.
func cleanItems(items []entryIn, cfg settings.Entries) error {
	for i := range items {
		it := &items[i]
		it.AssetID = strings.TrimSpace(it.AssetID)
		it.SourceGUID = strings.TrimSpace(it.SourceGUID)
		it.Type = strings.TrimSpace(it.Type)
		it.Name = strings.TrimSpace(it.Name)
		if !assetIDPattern.MatchString(it.AssetID) {
			return invalid(i, it.AssetID, "assetId", "1-128 of A-Z a-z 0-9 _ . : -")
		}
		if badText(it.SourceGUID, maxGUID) {
			return invalid(i, it.AssetID, "sourceGuid", "at most 64 characters")
		}
		if it.Type == "" || badText(it.Type, maxType) {
			return invalid(i, it.AssetID, "type", "required, at most 255 characters")
		}
		if it.Name == "" {
			it.Name = it.AssetID
		}
		if badText(it.Name, maxName) {
			return invalid(i, it.AssetID, "name", "at most 255 characters")
		}
		f, ok := cleanFolder(it.Folder)
		if !ok {
			return invalid(i, it.AssetID, "folder", "segments may not be empty, . or ..")
		}
		it.Folder = f
		if len(it.Data) == 0 {
			return invalid(i, it.AssetID, "data", "a JSON object is required")
		}
		doc, err := canon.Parse(it.Data)
		if err != nil {
			return invalid(i, it.AssetID, "data", err.Error())
		}
		if len(doc.Compact) > cfg.MaxEntryBytes {
			return apierr.BadRequest("ASSET_TOO_LARGE").WithField(fmt.Sprintf("items[%d].data", i)).With("assetId", it.AssetID).With("max", cfg.MaxEntryBytes).With("size", len(doc.Compact))
		}
		it.doc = doc
	}
	return nil
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const entryCols = `asset_id, source_guid, type, folder, name, hash, size_bytes, version, updated_by, created_at, updated_at`

func scanEntry(sc interface{ Scan(...any) error }, withData bool) (entry, error) {
	var e entry
	dest := []any{&e.AssetID, &e.SourceGUID, &e.Type, &e.Folder, &e.Name, &e.Hash, &e.Size, &e.Version, &e.UpdatedBy, &e.CreatedAt, &e.UpdatedAt}
	var data []byte
	if withData {
		dest = append(dest, &data)
	}
	if err := sc.Scan(dest...); err != nil {
		return e, err
	}
	if withData {
		e.Data = json.RawMessage(data)
	}
	return e, nil
}

func loadEntries(ctx context.Context, q queryer, ids []string, lock bool, withData bool) (map[string]entry, error) {
	out := map[string]entry{}
	if len(ids) == 0 {
		return out, nil
	}
	query := `SELECT ` + entryCols
	if withData {
		query += `, data::text`
	}
	query += ` FROM assets_entry WHERE asset_id = ANY($1) ORDER BY asset_id`
	if lock {
		query += ` FOR UPDATE`
	}
	rows, err := q.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e, err := scanEntry(rows, withData)
		if err != nil {
			return nil, err
		}
		out[e.AssetID] = e
	}
	return out, rows.Err()
}

func same(e entry, it entryIn) bool {
	return e.Hash == it.doc.Hash && e.Folder == it.Folder && e.Name == it.Name && e.Type == it.Type && e.SourceGUID == it.SourceGUID
}

func versionsEqual(a *int, b int) bool { return a != nil && *a == b }

type pushOut struct {
	DryRun    bool              `json:"dryRun"`
	Created   int               `json:"created"`
	Updated   int               `json:"updated"`
	Unchanged int               `json:"unchanged"`
	Deleted   int               `json:"deleted"`
	Conflicts []conflict        `json:"conflicts"`
	Rows      []pushRow         `json:"rows"`
	Items     []manifestItem    `json:"items"`
	Routing   *itemkinds.RunOut `json:"routing,omitempty"`
}

func uniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

// push writes a batch of entries and deletions. A stale baseVersion (the entry changed since the
// tool read it) is a conflict; sending exactly what the server holds never is.
func (s *service) push(ctx context.Context, in pushIn, actor string, mayDelete bool) (pushOut, error) {
	cfg := s.settings(ctx).Entries
	n := len(in.Items) + len(in.Deletes)
	if n == 0 {
		return pushOut{}, apierr.BadRequest("ASSET_PUSH_EMPTY")
	}
	if n > cfg.MaxPushItems {
		return pushOut{}, apierr.BadRequest("ASSET_PUSH_TOO_MANY").With("max", cfg.MaxPushItems)
	}
	if len(in.Deletes) > 0 && !mayDelete {
		return pushOut{}, apierr.Forbidden("ASSET_DELETE_NOT_ALLOWED").With("scope", ScopeDelete)
	}
	if in.Force && !cfg.AllowForce {
		return pushOut{}, apierr.Forbidden("ASSET_FORCE_NOT_ALLOWED")
	}
	if err := cleanItems(in.Items, cfg); err != nil {
		return pushOut{}, err
	}
	for i := range in.Deletes {
		in.Deletes[i].AssetID = strings.TrimSpace(in.Deletes[i].AssetID)
		if !assetIDPattern.MatchString(in.Deletes[i].AssetID) {
			return pushOut{}, apierr.BadRequest("ASSET_INVALID").WithField(fmt.Sprintf("deletes[%d].assetId", i))
		}
	}
	// Duplicates inside the push.
	byID := map[string][]int{}
	byGUID := map[string][]string{}
	ids := []string{}
	for i, it := range in.Items {
		if len(byID[it.AssetID]) == 0 {
			ids = append(ids, it.AssetID)
		}
		byID[it.AssetID] = append(byID[it.AssetID], i)
		if it.SourceGUID != "" && (len(byGUID[it.SourceGUID]) == 0 || byGUID[it.SourceGUID][len(byGUID[it.SourceGUID])-1] != it.AssetID) {
			byGUID[it.SourceGUID] = append(byGUID[it.SourceGUID], it.AssetID)
		}
	}
	dups := []map[string]any{}
	for _, id := range ids {
		if idx := byID[id]; len(idx) > 1 {
			files := []map[string]any{}
			for _, i := range idx {
				files = append(files, map[string]any{"folder": in.Items[i].Folder, "name": in.Items[i].Name, "sourceGuid": in.Items[i].SourceGUID, "onServer": false})
			}
			dups = append(dups, map[string]any{"assetId": id, "assets": files})
		}
	}
	if len(dups) > 0 {
		return pushOut{}, apierr.BadRequest("ASSET_ID_DUPLICATED").With("duplicates", dups)
	}
	guids := []string{}
	for g, list := range byGUID {
		if len(list) > 1 {
			return pushOut{}, apierr.BadRequest("ASSET_SOURCE_DUPLICATED").With("sourceGuid", g).With("assetIds", list)
		}
		guids = append(guids, g)
	}
	deleted := map[string]bool{}
	delIDs := []string{}
	for _, d := range in.Deletes {
		if len(byID[d.AssetID]) > 0 || deleted[d.AssetID] {
			return pushOut{}, apierr.BadRequest("ASSET_ID_DUPLICATED").With("assetId", d.AssetID)
		}
		deleted[d.AssetID] = true
		delIDs = append(delIDs, d.AssetID)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return pushOut{}, err
	}
	defer func() { _ = tx.Rollback() }()
	all := append(append([]string{}, ids...), delIDs...)
	sort.Strings(all)
	current, err := loadEntries(ctx, tx, all, true, false)
	if err != nil {
		return pushOut{}, err
	}
	// Source file checks against the server.
	dups = []map[string]any{}
	for _, it := range in.Items {
		if e, ok := current[it.AssetID]; ok && e.SourceGUID != "" && it.SourceGUID != "" && e.SourceGUID != it.SourceGUID {
			dups = append(dups, map[string]any{"assetId": it.AssetID, "assets": []map[string]any{
				{"folder": e.Folder, "name": e.Name, "sourceGuid": e.SourceGUID, "onServer": true},
				{"folder": it.Folder, "name": it.Name, "sourceGuid": it.SourceGUID, "onServer": false}}})
		}
	}
	if len(dups) > 0 {
		return pushOut{}, apierr.BadRequest("ASSET_ID_DUPLICATED").With("duplicates", dups)
	}
	if len(guids) > 0 {
		rows, err := tx.QueryContext(ctx, `SELECT asset_id, source_guid FROM assets_entry WHERE source_guid = ANY($1)`, guids)
		if err != nil {
			return pushOut{}, err
		}
		owner := map[string]string{}
		for rows.Next() {
			var id, g string
			if err := rows.Scan(&id, &g); err != nil {
				rows.Close()
				return pushOut{}, err
			}
			owner[g] = id
		}
		rows.Close()
		changes := []map[string]any{}
		for _, it := range in.Items {
			if it.SourceGUID == "" {
				continue
			}
			if prev, ok := owner[it.SourceGUID]; ok && prev != it.AssetID && !deleted[prev] && len(byID[prev]) == 0 {
				changes = append(changes, map[string]any{"assetId": it.AssetID, "previousAssetId": prev, "folder": it.Folder, "name": it.Name})
			}
		}
		if len(changes) > 0 {
			return pushOut{}, apierr.BadRequest("ASSET_SOURCE_ID_CHANGED").With("changes", changes)
		}
	}
	// Version conflicts.
	out := pushOut{DryRun: in.DryRun, Conflicts: []conflict{}, Rows: []pushRow{}, Items: []manifestItem{}}
	stale := map[string]bool{}
	for _, it := range in.Items {
		e, ok := current[it.AssetID]
		switch {
		case ok && !same(e, it) && !versionsEqual(it.BaseVersion, e.Version) && !in.Force:
			out.Conflicts = append(out.Conflicts, conflict{AssetID: it.AssetID, Version: intp(e.Version), BaseVersion: it.BaseVersion})
			stale[it.AssetID] = true
		case !ok && it.BaseVersion != nil && !in.Force:
			out.Conflicts = append(out.Conflicts, conflict{AssetID: it.AssetID, Version: nil, BaseVersion: it.BaseVersion})
			stale[it.AssetID] = true
		}
	}
	for _, d := range in.Deletes {
		if e, ok := current[d.AssetID]; ok && !versionsEqual(d.BaseVersion, e.Version) && !in.Force {
			out.Conflicts = append(out.Conflicts, conflict{AssetID: d.AssetID, Version: intp(e.Version), BaseVersion: d.BaseVersion})
			stale[d.AssetID] = true
		}
	}
	if len(out.Conflicts) > 0 && cfg.ConflictMode == settings.ConflictReject {
		return pushOut{}, apierr.Conflict("ASSET_VERSION_CONFLICT").With("conflicts", out.Conflicts)
	}
	if cfg.MaxEntries > 0 {
		creates := 0
		for _, it := range in.Items {
			if _, ok := current[it.AssetID]; !ok && !stale[it.AssetID] {
				creates++
			}
		}
		if creates > 0 {
			var total int
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM assets_entry`).Scan(&total); err != nil {
				return pushOut{}, err
			}
			if total-len(in.Deletes)+creates > cfg.MaxEntries {
				return pushOut{}, apierr.Conflict("ASSET_LIMIT_REACHED").With("max", cfg.MaxEntries)
			}
		}
	}
	for _, it := range in.Items {
		if stale[it.AssetID] {
			c := conflictFor(out.Conflicts, it.AssetID)
			out.Rows = append(out.Rows, pushRow{AssetID: it.AssetID, Action: actConflict, Version: c.Version, BaseVersion: it.BaseVersion})
			continue
		}
		e, ok := current[it.AssetID]
		if ok && same(e, it) {
			out.Unchanged++
			out.Rows = append(out.Rows, pushRow{AssetID: it.AssetID, Action: actUnchanged, Version: intp(e.Version), Hash: e.Hash})
			out.Items = append(out.Items, manifestItem{AssetID: e.AssetID, SourceGUID: e.SourceGUID, Type: e.Type, Folder: e.Folder, Name: e.Name, Version: e.Version, Hash: e.Hash})
			continue
		}
		var version int
		action := actUpdated
		if ok {
			version = e.Version + 1
			if _, err := tx.ExecContext(ctx, `UPDATE assets_entry SET source_guid=$2, type=$3, folder=$4, name=$5, data=$6, hash=$7, size_bytes=$8, version=$9, updated_by=$10, updated_at=now() WHERE asset_id=$1`,
				it.AssetID, it.SourceGUID, it.Type, it.Folder, it.Name, string(it.doc.Compact), it.doc.Hash, len(it.doc.Compact), version, actor); err != nil {
				return pushOut{}, err
			}
			out.Updated++
		} else {
			action = actCreated
			if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(version), 0) + 1 FROM assets_entry_history WHERE asset_id=$1`, it.AssetID).Scan(&version); err != nil {
				return pushOut{}, err
			}
			res, err := tx.ExecContext(ctx, `INSERT INTO assets_entry (asset_id, source_guid, type, folder, name, data, hash, size_bytes, version, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT (asset_id) DO NOTHING`,
				it.AssetID, it.SourceGUID, it.Type, it.Folder, it.Name, string(it.doc.Compact), it.doc.Hash, len(it.doc.Compact), version, actor)
			if err != nil {
				return pushOut{}, err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				// Another push created it after this one read the table.
				if cfg.ConflictMode == settings.ConflictReject {
					return pushOut{}, apierr.Conflict("ASSET_VERSION_CONFLICT").With("conflicts", []conflict{{AssetID: it.AssetID, BaseVersion: it.BaseVersion}})
				}
				out.Conflicts = append(out.Conflicts, conflict{AssetID: it.AssetID, BaseVersion: it.BaseVersion})
				out.Rows = append(out.Rows, pushRow{AssetID: it.AssetID, Action: actConflict})
				continue
			}
			out.Created++
		}
		if err := addHistory(ctx, tx, it.AssetID, version, action, it.SourceGUID, it.Type, it.Folder, it.Name, string(it.doc.Compact), it.doc.Hash, len(it.doc.Compact), nil, actor); err != nil {
			return pushOut{}, err
		}
		row := pushRow{AssetID: it.AssetID, Action: action, Version: intp(version), Hash: it.doc.Hash}
		if ok {
			row.PreviousVersion = intp(e.Version)
		}
		out.Rows = append(out.Rows, row)
		out.Items = append(out.Items, manifestItem{AssetID: it.AssetID, SourceGUID: it.SourceGUID, Type: it.Type, Folder: it.Folder, Name: it.Name, Version: version, Hash: it.doc.Hash})
	}
	for _, d := range in.Deletes {
		if stale[d.AssetID] {
			c := conflictFor(out.Conflicts, d.AssetID)
			out.Rows = append(out.Rows, pushRow{AssetID: d.AssetID, Action: actConflict, Version: c.Version, BaseVersion: d.BaseVersion})
			continue
		}
		if _, ok := current[d.AssetID]; !ok {
			out.Rows = append(out.Rows, pushRow{AssetID: d.AssetID, Action: actMissing})
			continue
		}
		if err := deleteEntry(ctx, tx, d.AssetID, actor); err != nil {
			return pushOut{}, err
		}
		out.Deleted++
		out.Rows = append(out.Rows, pushRow{AssetID: d.AssetID, Action: actDeleted, PreviousVersion: intp(current[d.AssetID].Version)})
	}
	routing, err := s.route(ctx, tx, out, in.DryRun, actor)
	if err != nil {
		return pushOut{}, err
	}
	out.Routing = routing
	if in.DryRun {
		return out, nil
	}
	if err := tx.Commit(); err != nil {
		if uniqueViolation(err) {
			return pushOut{}, apierr.Conflict("ASSET_VERSION_CONFLICT").With("conflicts", []conflict{})
		}
		return pushOut{}, err
	}
	if routing != nil {
		itemkinds.SaveRun(ctx, s.db, itemkinds.Load(ctx, s.db), "push", actor, routing)
	}
	return out, nil
}

func (s *service) route(ctx context.Context, tx *sql.Tx, out pushOut, dryRun bool, actor string) (*itemkinds.RunOut, error) {
	if len(itemkinds.List()) == 0 {
		return nil, nil
	}
	in := itemkinds.RunInput{DryRun: dryRun, Actor: actor, Source: "push"}
	for _, r := range out.Rows {
		switch r.Action {
		case actCreated, actUpdated:
			in.Changed = append(in.Changed, r.AssetID)
		case actDeleted:
			in.Deleted = append(in.Deleted, r.AssetID)
		}
	}
	if len(in.Changed) == 0 && len(in.Deleted) == 0 {
		return nil, nil
	}
	res, err := itemkinds.Run(ctx, tx, itemkinds.Load(ctx, s.db), in)
	var rej *itemkinds.RejectError
	if errors.As(err, &rej) {
		return nil, apierr.Conflict(rej.Code).WithDetail(rej.Detail).With("rows", rej.Rows)
	}
	if err != nil {
		return nil, err
	}
	if !res.Enabled {
		return nil, nil
	}
	return &res, nil
}

func conflictFor(list []conflict, id string) conflict {
	for _, c := range list {
		if c.AssetID == id {
			return c
		}
	}
	return conflict{AssetID: id}
}

func addHistory(ctx context.Context, tx queryer, id string, version int, action, guid, typ, folder, name, data, hash string, size int, from *int64, actor string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO assets_entry_history (asset_id, version, action, source_guid, type, folder, name, data, hash, size_bytes, restored_from, actor) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, version, action, guid, typ, folder, name, data, hash, size, from, actor)
	return err
}

// deleteEntry removes an entry and keeps its last state in the history (so it can be restored).
func deleteEntry(ctx context.Context, tx queryer, id, actor string) error {
	_, err := tx.ExecContext(ctx, `WITH d AS (DELETE FROM assets_entry WHERE asset_id=$1 RETURNING *)
		INSERT INTO assets_entry_history (asset_id, version, action, source_guid, type, folder, name, data, hash, size_bytes, actor)
		SELECT asset_id, version, 'deleted', source_guid, type, folder, name, data, hash, size_bytes, $2 FROM d`, id, actor)
	return err
}

// scope selects entries by folder (with or without sub-folders) and types.
type scope struct {
	Folder    string   `json:"folder"`
	Recursive *bool    `json:"recursive"`
	Types     []string `json:"types"`
}

func (sc scope) where(args *[]any) (string, error) {
	f, ok := cleanFolder(sc.Folder)
	if !ok {
		return "", apierr.BadRequest("ASSET_INVALID").WithField("folder")
	}
	conds := []string{"true"}
	recursive := sc.Recursive == nil || *sc.Recursive
	if f != "" || !recursive {
		*args = append(*args, f)
		n := strconv.Itoa(len(*args))
		if recursive {
			*args = append(*args, likePrefix(f)+"/%")
			conds = append(conds, "(folder = $"+n+" OR folder LIKE $"+strconv.Itoa(len(*args))+")")
		} else {
			conds = append(conds, "folder = $"+n)
		}
	}
	if len(sc.Types) > 0 {
		if len(sc.Types) > 100 {
			return "", apierr.BadRequest("ASSET_INVALID").WithField("types")
		}
		*args = append(*args, sc.Types)
		conds = append(conds, "type = ANY($"+strconv.Itoa(len(*args))+")")
	}
	return strings.Join(conds, " AND "), nil
}

func likePrefix(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (s *service) manifest(ctx context.Context, sc scope, withGUID bool) ([]manifestItem, error) {
	args := []any{}
	where, err := sc.where(&args)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT asset_id, source_guid, type, folder, name, version, hash FROM assets_entry WHERE `+where+` ORDER BY folder, asset_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []manifestItem{}
	for rows.Next() {
		var m manifestItem
		if err := rows.Scan(&m.AssetID, &m.SourceGUID, &m.Type, &m.Folder, &m.Name, &m.Version, &m.Hash); err != nil {
			return nil, err
		}
		if !withGUID {
			m.SourceGUID = ""
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type fetchIn struct {
	scope
	Have  map[string]int `json:"have"`
	IDs   []string       `json:"ids"`
	Limit int            `json:"limit"`
}

type fetchOut struct {
	Items     []entry  `json:"items"`
	Removed   []string `json:"removed"`
	Missing   []string `json:"missing,omitempty"`
	Unchanged int      `json:"unchanged"`
	Total     int      `json:"total"`
	More      bool     `json:"more"`
}

// fetch answers only what the client does not have: entries whose version differs from the
// client's (have), and ids the client has that are gone from the scope. With ids it answers
// those entries (unknown ones in missing).
func (s *service) fetch(ctx context.Context, in fetchIn) (fetchOut, error) {
	cfg := s.settings(ctx).Entries
	limit := cfg.MaxFetchItems
	if in.Limit > 0 && in.Limit < limit {
		limit = in.Limit
	}
	out := fetchOut{Items: []entry{}, Removed: []string{}}
	if len(in.IDs) > 0 {
		if len(in.IDs) > cfg.MaxFetchItems {
			return out, apierr.BadRequest("ASSET_FETCH_TOO_MANY").With("max", cfg.MaxFetchItems)
		}
		seen := map[string]bool{}
		ids := []string{}
		for _, id := range in.IDs {
			id = strings.TrimSpace(id)
			if id != "" && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		found, err := loadEntries(ctx, s.db, ids, false, true)
		if err != nil {
			return out, err
		}
		sort.Strings(ids)
		for _, id := range ids {
			e, ok := found[id]
			switch {
			case !ok:
				out.Missing = append(out.Missing, id)
			case in.Have != nil && in.Have[id] == e.Version:
				out.Unchanged++
			default:
				out.Items = append(out.Items, e)
			}
		}
		out.Total = len(found)
		return out, nil
	}
	if len(in.Have) > 50000 {
		return out, apierr.BadRequest("ASSET_FETCH_TOO_MANY").With("max", 50000)
	}
	list, err := s.manifest(ctx, in.scope, false)
	if err != nil {
		return out, err
	}
	out.Total = len(list)
	inScope := map[string]bool{}
	changed := []string{}
	for _, m := range list {
		inScope[m.AssetID] = true
		if v, ok := in.Have[m.AssetID]; ok && v == m.Version {
			out.Unchanged++
			continue
		}
		changed = append(changed, m.AssetID)
	}
	for id := range in.Have {
		if !inScope[id] {
			out.Removed = append(out.Removed, id)
		}
	}
	sort.Strings(out.Removed)
	sort.Strings(changed)
	if len(changed) > limit {
		changed, out.More = changed[:limit], true
	}
	found, err := loadEntries(ctx, s.db, changed, false, true)
	if err != nil {
		return out, err
	}
	for _, id := range changed {
		if e, ok := found[id]; ok {
			out.Items = append(out.Items, e)
		} else {
			out.More = true // removed between the two reads; the next call settles it
		}
	}
	return out, nil
}

type historyRow struct {
	ID           int64           `json:"id"`
	AssetID      string          `json:"assetId"`
	Version      int             `json:"version"`
	Action       string          `json:"action"`
	SourceGUID   string          `json:"sourceGuid"`
	Type         string          `json:"type"`
	Folder       string          `json:"folder"`
	Name         string          `json:"name"`
	Hash         string          `json:"hash"`
	Size         int             `json:"size"`
	RestoredFrom *int64          `json:"restoredFrom"`
	Actor        string          `json:"actor"`
	CreatedAt    time.Time       `json:"createdAt"`
	Data         json.RawMessage `json:"data,omitempty"`
}

func (s *service) history(ctx context.Context, assetID string, limit int) ([]historyRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, asset_id, version, action, source_guid, type, folder, name, hash, size_bytes, restored_from, actor, created_at
		FROM assets_entry_history WHERE asset_id=$1 ORDER BY id DESC LIMIT $2`, assetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []historyRow{}
	for rows.Next() {
		var h historyRow
		var from sql.NullInt64
		if err := rows.Scan(&h.ID, &h.AssetID, &h.Version, &h.Action, &h.SourceGUID, &h.Type, &h.Folder, &h.Name, &h.Hash, &h.Size, &from, &h.Actor, &h.CreatedAt); err != nil {
			return nil, err
		}
		if from.Valid {
			v := from.Int64
			h.RestoredFrom = &v
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *service) historyState(ctx context.Context, id int64) (historyRow, error) {
	var h historyRow
	var from sql.NullInt64
	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT id, asset_id, version, action, source_guid, type, folder, name, hash, size_bytes, restored_from, actor, created_at, data::text FROM assets_entry_history WHERE id=$1`, id).
		Scan(&h.ID, &h.AssetID, &h.Version, &h.Action, &h.SourceGUID, &h.Type, &h.Folder, &h.Name, &h.Hash, &h.Size, &from, &h.Actor, &h.CreatedAt, &data)
	if isNoRows(err) {
		return h, apierr.NotFound("ASSET_HISTORY_NOT_FOUND")
	}
	if from.Valid {
		v := from.Int64
		h.RestoredFrom = &v
	}
	h.Data = json.RawMessage(data)
	return h, err
}

// rollback writes an older state (a history row) as the entry's next version; it also brings a
// deleted entry back. baseVersion guards against a change made since the admin opened the entry.
func (s *service) rollback(ctx context.Context, assetID string, historyID int64, baseVersion *int, actor string) (entry, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return entry{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var h historyRow
	var data string
	err = tx.QueryRowContext(ctx, `SELECT asset_id, version, source_guid, type, folder, name, data::text, hash, size_bytes FROM assets_entry_history WHERE id=$1`, historyID).
		Scan(&h.AssetID, &h.Version, &h.SourceGUID, &h.Type, &h.Folder, &h.Name, &data, &h.Hash, &h.Size)
	if isNoRows(err) || (err == nil && h.AssetID != assetID) {
		return entry{}, apierr.NotFound("ASSET_HISTORY_NOT_FOUND")
	}
	if err != nil {
		return entry{}, err
	}
	cur, err := loadEntries(ctx, tx, []string{assetID}, true, false)
	if err != nil {
		return entry{}, err
	}
	e, exists := cur[assetID]
	if exists && !versionsEqual(baseVersion, e.Version) {
		return entry{}, apierr.Conflict("ASSET_VERSION_CONFLICT").With("conflicts", []conflict{{AssetID: assetID, Version: intp(e.Version), BaseVersion: baseVersion}})
	}
	if !exists && baseVersion != nil {
		return entry{}, apierr.Conflict("ASSET_VERSION_CONFLICT").With("conflicts", []conflict{{AssetID: assetID, BaseVersion: baseVersion}})
	}
	if h.SourceGUID != "" {
		var other string
		err := tx.QueryRowContext(ctx, `SELECT asset_id FROM assets_entry WHERE source_guid=$1 AND asset_id<>$2 LIMIT 1`, h.SourceGUID, assetID).Scan(&other)
		if err == nil {
			return entry{}, apierr.BadRequest("ASSET_SOURCE_ID_CHANGED").With("changes", []map[string]any{{"assetId": assetID, "previousAssetId": other}})
		}
	}
	var version int
	if exists {
		version = e.Version + 1
		_, err = tx.ExecContext(ctx, `UPDATE assets_entry SET source_guid=$2, type=$3, folder=$4, name=$5, data=$6, hash=$7, size_bytes=$8, version=$9, updated_by=$10, updated_at=now() WHERE asset_id=$1`,
			assetID, h.SourceGUID, h.Type, h.Folder, h.Name, data, h.Hash, h.Size, version, actor)
	} else {
		if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(version), 0) + 1 FROM assets_entry_history WHERE asset_id=$1`, assetID).Scan(&version); err != nil {
			return entry{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO assets_entry (asset_id, source_guid, type, folder, name, data, hash, size_bytes, version, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			assetID, h.SourceGUID, h.Type, h.Folder, h.Name, data, h.Hash, h.Size, version, actor)
	}
	if err != nil {
		if uniqueViolation(err) {
			return entry{}, apierr.Conflict("ASSET_VERSION_CONFLICT")
		}
		return entry{}, err
	}
	if err := addHistory(ctx, tx, assetID, version, "restored", h.SourceGUID, h.Type, h.Folder, h.Name, data, h.Hash, h.Size, &historyID, actor); err != nil {
		return entry{}, err
	}
	if err := tx.Commit(); err != nil {
		return entry{}, err
	}
	out, err := loadEntries(ctx, s.db, []string{assetID}, false, true)
	if err != nil {
		return entry{}, err
	}
	return out[assetID], nil
}

// trimHistory keeps the newest keepVersions states of every entry (and drops older ones past
// keepDays), and forgets deleted entries after keepDeletedDays.
func (s *service) trimHistory(ctx context.Context, cfg settings.All) {
	h := cfg.History
	if h.KeepVersions > 0 {
		cond := ""
		args := []any{h.KeepVersions}
		if h.KeepDays > 0 {
			cond = ` AND created_at < now() - make_interval(days => $2)`
			args = append(args, h.KeepDays)
		}
		_, _ = s.db.ExecContext(ctx, `DELETE FROM assets_entry_history WHERE id IN (
			SELECT id FROM (SELECT id, created_at, row_number() OVER (PARTITION BY asset_id ORDER BY id DESC) AS rn FROM assets_entry_history) x
			WHERE rn > $1`+cond+` LIMIT 20000)`, args...)
	} else if h.KeepDays > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM assets_entry_history WHERE id IN (
			SELECT id FROM (SELECT id, created_at, row_number() OVER (PARTITION BY asset_id ORDER BY id DESC) AS rn FROM assets_entry_history) x
			WHERE rn > 1 AND created_at < now() - make_interval(days => $1) LIMIT 20000)`, h.KeepDays)
	}
	if h.KeepDeletedDays > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM assets_entry_history h WHERE NOT EXISTS (SELECT 1 FROM assets_entry e WHERE e.asset_id=h.asset_id)
			AND h.asset_id IN (SELECT asset_id FROM assets_entry_history WHERE action='deleted' GROUP BY asset_id HAVING max(created_at) < now() - make_interval(days => $1))`, h.KeepDeletedDays)
	}
}

// ---- game reads ----

type manifestReq struct {
	scope
	ETag string `json:"etag"`
}

func manifestETag(items []manifestItem) string {
	b, _ := json.Marshal(items)
	d, _ := canon.Parse(append(append([]byte(`{"i":`), b...), '}'))
	return `"` + d.Hash[:32] + `"`
}

func (s *service) rpcManifest(ctx context.Context, _ string, payload string) (any, error) {
	var in manifestReq
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	items, err := s.manifest(ctx, in.scope, false)
	if err != nil {
		return nil, err
	}
	etag := manifestETag(items)
	if in.ETag != "" && in.ETag == etag {
		return map[string]any{"etag": etag, "notModified": true, "count": len(items)}, nil
	}
	return map[string]any{"etag": etag, "notModified": false, "count": len(items), "items": items}, nil
}

func (s *service) rpcFetch(ctx context.Context, _ string, payload string) (any, error) {
	var in fetchIn
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	return s.fetch(ctx, in)
}

func boolQuery(r *http.Request, name string) *bool {
	v := r.URL.Query().Get(name)
	if v == "" {
		return nil
	}
	b := v == "true" || v == "1"
	return &b
}

func scopeFromQuery(r *http.Request) scope {
	q := r.URL.Query()
	sc := scope{Folder: q.Get("folder"), Recursive: boolQuery(r, "recursive")}
	if t := strings.TrimSpace(q.Get("types")); t != "" {
		sc.Types = strings.Split(t, ",")
	}
	return sc
}

func (s *service) httpManifest(w http.ResponseWriter, r *http.Request) {
	cfg := s.settings(r.Context()).Entries
	if cfg.Access != settings.AccessPublic {
		s.httpFail(w, r, apierr.Forbidden("ASSETS_PRIVATE"))
		return
	}
	items, err := s.manifest(r.Context(), scopeFromQuery(r), false)
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	etag := manifestETag(items)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(cfg.CacheSeconds))
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"etag": etag, "count": len(items), "items": items}, false)
}

func (s *service) httpFetch(w http.ResponseWriter, r *http.Request) {
	if s.settings(r.Context()).Entries.Access != settings.AccessPublic {
		s.httpFail(w, r, apierr.Forbidden("ASSETS_PRIVATE"))
		return
	}
	var in fetchIn
	if err := readJSON(r, &in, 4<<20); err != nil {
		s.httpFail(w, r, err)
		return
	}
	out, err := s.fetch(r.Context(), in)
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out, true)
}
