package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/assets/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
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
		{p + "/entries", http.MethodGet, s.entryList},
		{p + "/entries", http.MethodPost, s.entryCreate},
		{p + "/entries/{id}", http.MethodGet, s.entryGet},
		{p + "/entries/{id}", http.MethodPut, s.entryUpdate},
		{p + "/entries/{id}", http.MethodDelete, s.entryDelete},
		{p + "/entries/{id}/history", http.MethodGet, s.entryHistory},
		{p + "/entries/{id}/rollback", http.MethodPost, s.entryRollback},
		{p + "/history/{hid}", http.MethodGet, s.historyGet},
		{p + "/deleted", http.MethodGet, s.deletedList},
		{p + "/push", http.MethodPost, s.consolePush},
		{p + "/export", http.MethodGet, s.export},
		{p + "/packs", http.MethodGet, s.packList},
		{p + "/packs/events", http.MethodGet, s.packEvents},
		{p + "/packs/preview", http.MethodGet, s.packPreview},
		{p + "/packs/{id}", http.MethodGet, s.packGet},
		{p + "/packs/{id}", http.MethodDelete, s.packDelete},
		{p + "/packs/{id}/activate", http.MethodPost, s.packActivate},
		{p + "/packs/{id}/deactivate", http.MethodPost, s.packDeactivate},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func consoleActor(r *http.Request) string {
	if u := pluginkit.ConsoleUser(r); u != "" {
		return "console:" + u
	}
	return "console"
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt,
		"storage": s.store.Driver(), "publicStorage": s.store.PublicURL(blobKey("x")) != "", "serverBaseUrl": s.baseURL,
		"scopes": []string{ScopeRead, ScopeWrite, ScopeDelete, ScopeContent, ScopePublish}})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
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

func (s *service) overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	out := map[string]any{}
	var total, bytes int64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*), coalesce(sum(size_bytes),0) FROM assets_entry`).Scan(&total, &bytes); err != nil {
		s.fail(w, err)
		return
	}
	out["total"], out["totalBytes"] = total, bytes
	group := func(q string) []map[string]any {
		list := []map[string]any{}
		rows, err := s.db.QueryContext(ctx, q)
		if err != nil {
			return list
		}
		defer rows.Close()
		for rows.Next() {
			var k string
			var n int64
			if rows.Scan(&k, &n) == nil {
				list = append(list, map[string]any{"key": k, "count": n})
			}
		}
		return list
	}
	out["folders"] = group(`SELECT folder, count(*) FROM assets_entry GROUP BY folder ORDER BY folder LIMIT 2000`)
	out["types"] = group(`SELECT type, count(*) FROM assets_entry GROUP BY type ORDER BY count(*) DESC, type LIMIT 500`)
	var packs, active int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE active) FROM assets_pack WHERE status='ready'`).Scan(&packs, &active)
	out["packs"], out["activePacks"] = packs, active
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) entryList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 50, 500)
	sc := scopeFromQuery(r)
	if t := strings.TrimSpace(r.URL.Query().Get("type")); t != "" {
		sc.Types = []string{t}
	}
	args := []any{}
	where, err := sc.where(&args)
	if err != nil {
		s.fail(w, err)
		return
	}
	a := &pluginkit.Args{List: args}
	if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
		ph := a.Add(pluginkit.LikePattern(q))
		where += " AND (asset_id ILIKE " + ph + " OR name ILIKE " + ph + " OR type ILIKE " + ph + ")"
	}
	order := "folder, asset_id"
	switch r.URL.Query().Get("sort") {
	case "updated":
		order = "updated_at DESC, asset_id"
	case "size":
		order = "size_bytes DESC, asset_id"
	case "id":
		order = "asset_id"
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM assets_entry WHERE `+where, a.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := a.Add(limit), a.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+entryCols+` FROM assets_entry WHERE `+where+` ORDER BY `+order+` LIMIT `+lim+` OFFSET `+off, a.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []entry{}
	for rows.Next() {
		e, err := scanEntry(rows, false)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, e)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) entryGet(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	found, err := loadEntries(r.Context(), s.db, []string{id}, false, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	e, ok := found[id]
	if !ok {
		s.fail(w, apierr.NotFound("ASSET_NOT_FOUND"))
		return
	}
	hist, err := s.history(r.Context(), id, 50)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": e, "history": hist})
}

type consoleEntry struct {
	AssetID     string          `json:"assetId"`
	SourceGUID  string          `json:"sourceGuid"`
	Type        string          `json:"type"`
	Folder      string          `json:"folder"`
	Name        string          `json:"name"`
	Data        json.RawMessage `json:"data"`
	BaseVersion *int            `json:"baseVersion"`
}

func (s *service) saveOne(ctx context.Context, r *http.Request, in consoleEntry) (entry, error) {
	out, err := s.push(ctx, pushIn{Items: []entryIn{{AssetID: in.AssetID, SourceGUID: in.SourceGUID, Type: in.Type, Folder: in.Folder, Name: in.Name, Data: in.Data, BaseVersion: in.BaseVersion}}}, consoleActor(r), true)
	if err != nil {
		return entry{}, err
	}
	if len(out.Conflicts) > 0 {
		return entry{}, apierr.Conflict("ASSET_VERSION_CONFLICT").With("conflicts", out.Conflicts)
	}
	found, err := loadEntries(ctx, s.db, []string{strings.TrimSpace(in.AssetID)}, false, true)
	if err != nil {
		return entry{}, err
	}
	return found[strings.TrimSpace(in.AssetID)], nil
}

func (s *service) entryCreate(w http.ResponseWriter, r *http.Request) {
	var in consoleEntry
	if err := readJSON(r, &in, 17<<20); err != nil {
		s.fail(w, err)
		return
	}
	found, err := loadEntries(r.Context(), s.db, []string{strings.TrimSpace(in.AssetID)}, false, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	if _, ok := found[strings.TrimSpace(in.AssetID)]; ok {
		s.fail(w, apierr.Conflict("ASSET_EXISTS").WithField("assetId"))
		return
	}
	in.BaseVersion = nil
	e, err := s.saveOne(r.Context(), r, in)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusCreated, map[string]any{"item": e})
}

func (s *service) entryUpdate(w http.ResponseWriter, r *http.Request) {
	var in consoleEntry
	if err := readJSON(r, &in, 17<<20); err != nil {
		s.fail(w, err)
		return
	}
	in.AssetID = pluginkit.PathVar(r, "id")
	found, err := loadEntries(r.Context(), s.db, []string{in.AssetID}, false, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	cur, ok := found[in.AssetID]
	if !ok {
		s.fail(w, apierr.NotFound("ASSET_NOT_FOUND"))
		return
	}
	if in.BaseVersion == nil {
		s.fail(w, apierr.BadRequest("ASSET_INVALID").WithField("baseVersion"))
		return
	}
	if in.SourceGUID == "" {
		in.SourceGUID = cur.SourceGUID
	}
	if strings.TrimSpace(in.Type) == "" {
		in.Type = cur.Type
	}
	e, err := s.saveOne(r.Context(), r, in)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": e})
}

func (s *service) entryDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	var base *int
	if v, err := strconv.Atoi(r.URL.Query().Get("baseVersion")); err == nil {
		base = &v
	}
	out, err := s.push(r.Context(), pushIn{Deletes: []deleteIn{{AssetID: id, BaseVersion: base}}}, consoleActor(r), true)
	if err != nil {
		s.fail(w, err)
		return
	}
	if len(out.Rows) == 1 && out.Rows[0].Action == actMissing {
		s.fail(w, apierr.NotFound("ASSET_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) entryHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 500 {
		limit = 100
	}
	list, err := s.history(r.Context(), pluginkit.PathVar(r, "id"), limit)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (s *service) historyGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(pluginkit.PathVar(r, "hid"), 10, 64)
	if err != nil {
		s.fail(w, apierr.NotFound("ASSET_HISTORY_NOT_FOUND"))
		return
	}
	h, err := s.historyState(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": h})
}

func (s *service) entryRollback(w http.ResponseWriter, r *http.Request) {
	var in struct {
		HistoryID   int64 `json:"historyId"`
		BaseVersion *int  `json:"baseVersion"`
	}
	if err := readJSON(r, &in, 4096); err != nil {
		s.fail(w, err)
		return
	}
	e, err := s.rollback(r.Context(), pluginkit.PathVar(r, "id"), in.HistoryID, in.BaseVersion, consoleActor(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": e})
}

func (s *service) deletedList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT DISTINCT ON (h.asset_id) h.id, h.asset_id, h.version, h.action, h.source_guid, h.type, h.folder, h.name, h.hash, h.size_bytes, h.restored_from, h.actor, h.created_at
		FROM assets_entry_history h WHERE h.action='deleted' AND NOT EXISTS (SELECT 1 FROM assets_entry e WHERE e.asset_id=h.asset_id)
		ORDER BY h.asset_id, h.id DESC LIMIT 500`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	out := []historyRow{}
	for rows.Next() {
		var h historyRow
		var from *int64
		if err := rows.Scan(&h.ID, &h.AssetID, &h.Version, &h.Action, &h.SourceGUID, &h.Type, &h.Folder, &h.Name, &h.Hash, &h.Size, &from, &h.Actor, &h.CreatedAt); err != nil {
			s.fail(w, err)
			return
		}
		h.RestoredFrom = from
		out = append(out, h)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *service) consolePush(w http.ResponseWriter, r *http.Request) {
	var in pushIn
	if err := readJSON(r, &in, 64<<20); err != nil {
		s.fail(w, err)
		return
	}
	out, err := s.push(r.Context(), in, consoleActor(r), true)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

// export answers the entries of a folder in the push body shape (to move data between servers).
func (s *service) export(w http.ResponseWriter, r *http.Request) {
	sc := scopeFromQuery(r)
	args := []any{}
	where, err := sc.where(&args)
	if err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT asset_id, source_guid, type, folder, name, data::text FROM assets_entry WHERE `+where+` ORDER BY folder, asset_id LIMIT 20000`, args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, guid, typ, folder, name, data string
		if err := rows.Scan(&id, &guid, &typ, &folder, &name, &data); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, map[string]any{"assetId": id, "sourceGuid": guid, "type": typ, "folder": folder, "name": name, "data": json.RawMessage(data)})
	}
	w.Header().Set("Content-Disposition", `attachment; filename="assets-export.json"`)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// ---- packs ----

func (s *service) packList(w http.ResponseWriter, r *http.Request) {
	list, err := s.listPacks(r.Context(), r.URL.Query().Get("platform"), r.URL.Query().Get("status"), 500)
	if err != nil {
		s.fail(w, err)
		return
	}
	current := map[string]string{}
	// The pack each platform's index answers when the app version is not given.
	platforms := []string{}
	seen := map[string]bool{}
	for _, p := range list {
		if !seen[p.Platform] {
			seen[p.Platform] = true
			platforms = append(platforms, p.Platform)
		}
	}
	for _, pl := range platforms {
		if p, err := s.latestPack(r.Context(), pl, ""); err == nil && p != nil {
			current[pl] = p.ID
		}
	}
	cfg := s.settings(r.Context()).Packs
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": list, "current": current, "platforms": cfg.Platforms})
}

func (s *service) packGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, err := s.loadPack(ctx, s.db, pluginkit.PathVar(r, "id"), false)
	if err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT f.path, f.sha256, f.size, coalesce(b.stored, false) FROM assets_pack_file f LEFT JOIN assets_blob b ON b.sha256=f.sha256 WHERE f.pack_id=$1 ORDER BY f.path LIMIT 5000`, p.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	files := []map[string]any{}
	for rows.Next() {
		var path, sha string
		var size int64
		var stored bool
		if rows.Scan(&path, &sha, &size, &stored) == nil {
			files = append(files, map[string]any{"path": path, "sha256": sha, "size": size, "stored": stored})
		}
	}
	rows.Close()
	events, _ := s.events(ctx, p.ID, 50)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": p, "files": files, "events": events, "index": indexView(&p, s.publicBase(r))})
}

func (s *service) events(ctx context.Context, packID string, limit int) ([]map[string]any, error) {
	q := `SELECT id, pack_id::text, platform, version, action, actor, created_at FROM assets_pack_event`
	args := []any{}
	if packID != "" {
		q += ` WHERE pack_id=$1`
		args = append(args, packID)
	}
	args = append(args, limit)
	q += ` ORDER BY id DESC LIMIT $` + strconv.Itoa(len(args))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var pid, platform, version, action, actor string
		var at any
		if err := rows.Scan(&id, &pid, &platform, &version, &action, &actor, &at); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "packId": pid, "platform": platform, "version": version, "action": action, "actor": actor, "createdAt": at})
	}
	return out, rows.Err()
}

func (s *service) packEvents(w http.ResponseWriter, r *http.Request) {
	list, err := s.events(r.Context(), "", 200)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (s *service) packPreview(w http.ResponseWriter, r *http.Request) {
	in := indexIn{Platform: r.URL.Query().Get("platform"), AppVersion: r.URL.Query().Get("appVersion")}
	if err := in.check(); err != nil {
		s.fail(w, err)
		return
	}
	p, err := s.latestPack(r.Context(), in.Platform, in.AppVersion)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"platform": in.Platform, "appVersion": in.AppVersion, "pack": indexView(p, s.publicBase(r))})
}

func (s *service) packActivate(w http.ResponseWriter, r *http.Request) {
	p, err := s.activate(r.Context(), pluginkit.PathVar(r, "id"), true, consoleActor(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": p})
}

func (s *service) packDeactivate(w http.ResponseWriter, r *http.Request) {
	p, err := s.activate(r.Context(), pluginkit.PathVar(r, "id"), false, consoleActor(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": p})
}

func (s *service) packDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.deletePack(r.Context(), pluginkit.PathVar(r, "id"), consoleActor(r)); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
