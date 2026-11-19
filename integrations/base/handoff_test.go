package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
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

type handoffSocket struct{ key string }

func (s handoffSocket) GetServerKey() string { return s.key }
func (s handoffSocket) GetPort() int         { return 7350 }
func (s handoffSocket) GetAddress() string   { return "" }
func (s handoffSocket) GetProtocol() string  { return "tcp" }

type handoffConfig struct {
	runtime.Config
	key string
}

func (c handoffConfig) GetSocket() runtime.SocketConfig { return handoffSocket{c.key} }

func TestHandoffUsesLiveServerKey(t *testing.T) {
	key := "live-" + strings.Repeat("k", 20)
	s := &service{config: handoffConfig{key: key}, env: map[string]string{"HANDOFF_PUBLIC_HOST": "api.game.test"}}
	rec := httptest.NewRecorder()
	s.handoffDownload(rec, httptest.NewRequest(http.MethodGet, "/v2/console/base/handoff/download", nil))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, key) || !strings.Contains(body, "https://api.game.test") || strings.Contains(body, "{{SERVER_") {
		t.Fatalf("download %d without the live key", rec.Code)
	}
	rec = httptest.NewRecorder()
	s.handoffGet(rec, httptest.NewRequest(http.MethodGet, "/v2/console/base/handoff", nil))
	var meta struct {
		Connection map[string]any `json:"connection"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil || meta.Connection["serverKey"] != key || meta.Connection["host"] != "api.game.test" {
		t.Fatalf("meta %s", rec.Body.String())
	}
	rec = httptest.NewRecorder()
	s.connectionGet(rec, httptest.NewRequest(http.MethodGet, "/v2/console/base/handoff/connection", nil))
	if !strings.Contains(rec.Body.String(), key) {
		t.Fatalf("connection %s", rec.Body.String())
	}
}
