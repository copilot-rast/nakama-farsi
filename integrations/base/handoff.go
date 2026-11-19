package main

import (
	"net/http"
	"strings"
	"time"

	content "github.com/heroiclabs/nakama/v3/docs/handoff"
	"github.com/heroiclabs/nakama/v3/internal/buildinfo"
	"github.com/heroiclabs/nakama/v3/internal/handoffkit"
)

const (
	CodeHandoffOff        = "HANDOFF_OFF"
	CodeHandoffAdminsOnly = "HANDOFF_ADMINS_ONLY"
)

func handoffAdmin(r *http.Request) bool {
	return r.Header.Get(handoffkit.AdminHeader) == "true"
}

func (s *service) handoffAccess(r *http.Request) (handoffkit.Settings, string) {
	set := handoffkit.LoadSettings(r.Context(), s.db)
	if !set.Enabled {
		return set, CodeHandoffOff
	}
	if !set.Allows(handoffAdmin(r)) {
		return set, CodeHandoffAdminsOnly
	}
	return set, ""
}

func (s *service) handoffServer() handoffkit.Server {
	out := handoffkit.Server{Env: s.env}
	if s.config != nil && s.config.GetSocket() != nil {
		out.ServerKey = s.config.GetSocket().GetServerKey()
	}
	return out
}

func (s *service) connectionGet(w http.ResponseWriter, r *http.Request) {
	set := handoffkit.LoadSettings(r.Context(), s.db)
	if !handoffAdmin(r) && set.Audience == handoffkit.AudienceAdmins {
		set.ShowServerKey = false
	}
	writeJSON(w, http.StatusOK, map[string]any{"connection": handoffkit.Resolve(s.handoffServer(), set), "showServerKey": set.ShowServerKey})
}

func (s *service) handoffGet(w http.ResponseWriter, r *http.Request) {
	set, refused := s.handoffAccess(r)
	out := map[string]any{"visible": refused == "", "settings": set, "admin": handoffAdmin(r)}
	if refused != "" {
		out["reason"] = refused
		writeJSON(w, http.StatusOK, out)
		return
	}
	b, err := handoffkit.Load(content.FS)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var project any
	if b.Project != nil {
		m := b.Project.Meta
		project = map[string]any{"name": m.Name, "title": m.Title, "titleEn": m.TitleEn, "summary": m.Summary, "summaryEn": m.SummaryEn}
	}
	sections := []map[string]string{}
	for _, sec := range b.Sections() {
		sections = append(sections, map[string]string{"scope": sec.Scope, "id": sec.ID, "path": sec.Path, "title": sec.Title, "titleEn": sec.TitleEn})
	}
	general := b.General.Meta
	out["general"] = map[string]any{"title": general.Title, "titleEn": general.TitleEn, "summary": general.Summary, "summaryEn": general.SummaryEn}
	out["project"] = project
	out["sections"] = sections
	out["changelog"] = b.Changelog()
	out["fileName"] = handoffkit.FileName(b, time.Now())
	out["commit"] = buildinfo.Get().Commit
	out["connection"] = handoffkit.Resolve(s.handoffServer(), set)
	writeJSON(w, http.StatusOK, out)
}

func (s *service) handoffDownload(w http.ResponseWriter, r *http.Request) {
	set, refused := s.handoffAccess(r)
	if refused != "" {
		s.writeError(w, fail(http.StatusForbidden, refused, "", "the handoff download is not available for this console user"))
		return
	}
	b, err := handoffkit.Load(content.FS)
	if err != nil {
		s.writeError(w, err)
		return
	}
	now := time.Now()
	raw, err := handoffkit.Render(content.FS, b, handoffkit.Options{ConsoleURL: requestOrigin(r), Commit: buildinfo.Get().Commit, GeneratedAt: now, Connection: handoffkit.Resolve(s.handoffServer(), set)})
	if err != nil {
		s.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+handoffkit.FileName(b, now)+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]); p == "http" || p == "https" {
		scheme = p
	}
	host := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0])
	if host == "" {
		host = r.Host
	}
	if host == "" || strings.ContainsAny(host, " \"'<>/\\") {
		return ""
	}
	return scheme + "://" + host
}
