package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/content/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/jackc/pgx/v5/pgconn"
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
		{p + "/packs", http.MethodGet, s.packList},
		{p + "/packs", http.MethodPost, s.packCreate},
		{p + "/packs/events", http.MethodGet, s.packEvents},
		{p + "/packs/preview", http.MethodGet, s.packPreview},
		{p + "/packs/{id}", http.MethodGet, s.packGet},
		{p + "/packs/{id}", http.MethodDelete, s.packDelete},
		{p + "/packs/{id}/blobs/{sha}/{part}", http.MethodPut, s.packPart},
		{p + "/packs/{id}/complete", http.MethodPost, s.packComplete},
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

func uniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt,
		"storage": s.store.Driver(), "publicBaseUrl": s.indexBase(r.Context()), "maxPartBytes": settings.MaxPartBytes})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	if in.Packs.Platforms == nil {
		in.Packs.Platforms = []string{}
	}
	in.Packs.BaseURL = strings.TrimSpace(in.Packs.BaseURL)
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

func (s *service) packList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.listPacks(ctx, r.URL.Query().Get("platform"), r.URL.Query().Get("status"), 500)
	if err != nil {
		s.fail(w, err)
		return
	}
	cfg := s.settings(ctx).Packs
	platforms := append([]string{}, cfg.Platforms...)
	seen := map[string]bool{}
	for _, p := range platforms {
		seen[p] = true
	}
	for _, p := range list {
		if !seen[p.Platform] {
			seen[p.Platform] = true
			platforms = append(platforms, p.Platform)
		}
	}
	current := map[string]string{}
	for _, pl := range platforms {
		if p, err := s.latestPack(ctx, pl, ""); err == nil && p != nil {
			current[pl] = p.ID
		}
	}
	var stored, bytes int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*), coalesce(sum(size),0) FROM content_blob WHERE stored`).Scan(&stored, &bytes)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": list, "current": current, "platforms": platforms, "partBytes": cfg.PartBytes,
		"storage": map[string]any{"driver": s.store.Driver(), "files": stored, "bytes": bytes}})
}

func (s *service) packGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, err := s.loadPack(ctx, s.db, pluginkit.PathVar(r, "id"), false)
	if err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT f.path, f.sha256, f.size, coalesce(b.stored, false) FROM content_pack_file f LEFT JOIN content_blob b ON b.sha256=f.sha256 WHERE f.pack_id=$1 ORDER BY f.path LIMIT 5000`, p.ID)
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
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": p, "files": files, "events": events, "index": indexView(&p, s.indexBaseFor(r))})
}

func (s *service) indexBaseFor(r *http.Request) string {
	if b := s.indexBase(r.Context()); b != "" {
		return b
	}
	return s.publicBase(r)
}

func (s *service) events(ctx context.Context, packID string, limit int) ([]map[string]any, error) {
	q := `SELECT id, pack_id::text, platform, version, action, actor, created_at FROM content_pack_event`
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
	if err := in.check(false); err != nil {
		s.fail(w, err)
		return
	}
	p, err := s.latestPack(r.Context(), in.Platform, in.AppVersion)
	if err != nil {
		s.fail(w, err)
		return
	}
	view, err := s.indexOf(r.Context(), p, s.indexBaseFor(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"platform": in.Platform, "appVersion": in.AppVersion, "pack": view})
}

func (s *service) packCreate(w http.ResponseWriter, r *http.Request) {
	var in packCreateIn
	if err := pluginkit.ReadBody(r, &in, 3<<20); err != nil {
		s.fail(w, err)
		return
	}
	out, err := s.createPack(r.Context(), in, consoleActor(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) packPart(w http.ResponseWriter, r *http.Request) {
	sha := strings.ToLower(pluginkit.PathVar(r, "sha"))
	part, err := strconv.Atoi(pluginkit.PathVar(r, "part"))
	if !sha256Pattern.MatchString(sha) || err != nil {
		s.fail(w, apierr.BadRequest("PACK_PART_INVALID"))
		return
	}
	out, err := s.putPart(r.Context(), pluginkit.PathVar(r, "id"), sha, part, r.Body)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) packComplete(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Activate bool `json:"activate"`
	}
	if err := pluginkit.ReadBody(r, &in, 4096); err != nil {
		s.fail(w, err)
		return
	}
	p, err := s.completePack(r.Context(), pluginkit.PathVar(r, "id"), in.Activate, consoleActor(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"item": p})
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
