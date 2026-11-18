package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandoffDefaultsServeEveryConsoleUser(t *testing.T) {
	s := &service{}
	rec := httptest.NewRecorder()
	s.handoffGet(rec, httptest.NewRequest(http.MethodGet, "/v2/console/base/handoff", nil))
	var meta struct {
		Visible   bool             `json:"visible"`
		Changelog []map[string]any `json:"changelog"`
		Sections  []map[string]any `json:"sections"`
		FileName  string           `json:"fileName"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil || !meta.Visible || len(meta.Changelog) == 0 || len(meta.Sections) == 0 || !strings.HasSuffix(meta.FileName, ".html") {
		t.Fatalf("meta %d %s", rec.Code, rec.Body.String())
	}
	req := httptest.NewRequest(http.MethodGet, "/v2/console/base/handoff/download", nil)
	req.Host = "console.example:7351"
	req.Header.Set("X-Forwarded-Proto", "https")
	rec = httptest.NewRecorder()
	s.handoffDownload(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Content-Disposition"), "attachment") || !strings.Contains(body, `id="part-general"`) || !strings.Contains(body, "https://console.example:7351/#/system/sdk") {
		t.Fatalf("download %d %s", rec.Code, rec.Header())
	}
}

func TestRequestOriginRejectsOddHosts(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "a\"b"
	if requestOrigin(r) != "" {
		t.Fatal("odd host accepted")
	}
}
