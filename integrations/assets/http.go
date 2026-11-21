package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
)

func writeJSON(w http.ResponseWriter, status int, body any, noStore bool) {
	w.Header().Set("Content-Type", "application/json")
	if noStore {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}

func readJSON(r *http.Request, out any, max int64) error {
	raw, err := io.ReadAll(io.LimitReader(r.Body, max+1))
	if err != nil {
		return apierr.BadRequest("INVALID_JSON").WithDetail("cannot read the body (it may be over the server's request size limit)")
	}
	if int64(len(raw)) > max {
		return apierr.New(http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE").With("max", max)
	}
	if strings.TrimSpace(string(raw)) == "" {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return apierr.BadRequest("INVALID_JSON").WithDetail(err.Error())
	}
	return nil
}

func actorOf(key *apikey.Key) string {
	if key == nil {
		return "editor"
	}
	return "editor:" + key.Name
}

// ---- editor endpoints (API key) ----

func (s *service) editorManifest(w http.ResponseWriter, r *http.Request, _ *apikey.Key) {
	items, err := s.manifest(r.Context(), scopeFromQuery(r), true)
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "items": items, "etag": manifestETag(items)}, true)
}

func (s *service) editorFetch(w http.ResponseWriter, r *http.Request, _ *apikey.Key) {
	var in struct {
		AssetIDs []string `json:"assetIds"`
	}
	if err := readJSON(r, &in, 1<<20); err != nil {
		s.httpFail(w, r, err)
		return
	}
	if len(in.AssetIDs) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"items": []entry{}, "missing": []string{}}, true)
		return
	}
	out, err := s.fetch(r.Context(), fetchIn{IDs: in.AssetIDs})
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	if out.Missing == nil {
		out.Missing = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out.Items, "missing": out.Missing}, true)
}

func (s *service) editorPush(w http.ResponseWriter, r *http.Request, key *apikey.Key) {
	cfg := s.settings(r.Context()).Entries
	var in pushIn
	if err := readJSON(r, &in, int64(cfg.MaxPushItems)*int64(cfg.MaxEntryBytes+1024)+4096); err != nil {
		s.httpFail(w, r, err)
		return
	}
	out, err := s.push(r.Context(), in, actorOf(key), apikey.Allows(key.Scopes, ScopeDelete, true))
	if err != nil {
		s.httpFail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out, true)
}

// editorLimits tells a tool how to split its work (push size, fetch size, upload part size).
func (s *service) editorLimits(w http.ResponseWriter, r *http.Request, _ *apikey.Key) {
	cfg := s.settings(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{"maxEntryBytes": cfg.Entries.MaxEntryBytes, "maxPushItems": cfg.Entries.MaxPushItems, "maxFetchItems": cfg.Entries.MaxFetchItems,
		"conflictMode": cfg.Entries.ConflictMode, "allowForce": cfg.Entries.AllowForce, "partBytes": cfg.Packs.PartBytes, "maxFileBytes": cfg.Packs.MaxFileBytes,
		"maxPackBytes": cfg.Packs.MaxPackBytes, "maxFiles": cfg.Packs.MaxFiles, "platforms": cfg.Packs.Platforms, "catalogPattern": cfg.Packs.CatalogPattern}, true)
}
