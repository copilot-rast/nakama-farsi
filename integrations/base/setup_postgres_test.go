package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func resetBrandSetup(h *harness) {
	h.t.Helper()
	for _, q := range []string{
		`INSERT INTO base_brand (id) VALUES (1) ON CONFLICT DO NOTHING`,
		`UPDATE base_brand SET name='{}', short_name='', logo_url='', logo_dark_url='', mark_url='', accent='', welcome='{}', updated_by=''`,
		`DELETE FROM base_setup`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *harness) apiAs(user, method, path string, body any) (int, map[string]any, http.Header) {
	h.t.Helper()
	raw, _ := json.Marshal(body)
	if body == nil {
		raw = nil
	}
	req := httptest.NewRequest(method, "/v2/console/base"+path, bytes.NewReader(raw))
	if user != "" {
		req.Header.Set(consoleUserHeader, user)
	}
	rec := httptest.NewRecorder()
	h.init.console.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out, rec.Header()
}

func TestSetupClaimOnceUnderConcurrency(t *testing.T) {
	h := newHarness(t)
	resetBrandSetup(h)
	st := h.mustAPI(http.StatusOK, http.MethodGet, "/setup", nil)
	if st["status"] != "pending" || st["nextStep"] != "brand" {
		t.Fatalf("fresh setup = %v", st)
	}
	const admins = 16
	var wg sync.WaitGroup
	var mu sync.Mutex
	winners := []string{}
	for i := 0; i < admins; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "admin" + string(rune('a'+i))
			claimed, _, err := h.s.claimSetup(h.ctx, name)
			if err != nil {
				t.Error(err)
				return
			}
			if claimed {
				mu.Lock()
				winners = append(winners, name)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if len(winners) != 1 {
		t.Fatalf("claim winners = %v, want exactly one", winners)
	}
	st = h.mustAPI(http.StatusOK, http.MethodGet, "/setup", nil)
	if st["status"] != "in_progress" || st["firstOpenedBy"] != winners[0] || st["firstOpenedAt"] == nil {
		t.Fatalf("after claim = %v", st)
	}
	code, out, _ := h.apiAs("late", http.MethodPost, "/setup/claim", nil)
	if code != http.StatusOK || out["claimed"] != false {
		t.Fatalf("second claim = %d %v", code, out)
	}
}

func TestSetupStateMachine(t *testing.T) {
	h := newHarness(t)
	resetBrandSetup(h)
	_, out, _ := h.apiAs("first", http.MethodPost, "/setup/claim", nil)
	if out["claimed"] != true {
		t.Fatalf("first claim = %v", out)
	}
	_, st, _ := h.apiAs("first", http.MethodPut, "/setup", map[string]any{"steps": map[string]string{"brand": "done", "languages": "skipped"}})
	if st["nextStep"] != "currency" || st["status"] != "in_progress" {
		t.Fatalf("after two steps = %v", st)
	}
	_, st, _ = h.apiAs("first", http.MethodPut, "/setup", map[string]any{"status": "skipped"})
	if st["status"] != "skipped" || st["nextStep"] != "currency" {
		t.Fatalf("later keeps progress = %v", st)
	}
	_, out, _ = h.apiAs("second", http.MethodPost, "/setup/claim", nil)
	if out["claimed"] != false {
		t.Fatalf("skipped setup must not be claimed again: %v", out)
	}
	h.expectError(http.StatusBadRequest, "INVALID_STEP", http.MethodPut, "/setup", map[string]any{"steps": map[string]string{"nope": "done"}})
	h.expectError(http.StatusBadRequest, "INVALID_STEP_STATUS", http.MethodPut, "/setup", map[string]any{"steps": map[string]string{"sms": "maybe"}})
	h.expectError(http.StatusBadRequest, "INVALID_STATUS", http.MethodPut, "/setup", map[string]any{"status": "done"})
	steps := map[string]string{"currency": "done", "payment": "done", "sms": "skipped", "email": "done"}
	_, st, _ = h.apiAs("second", http.MethodPut, "/setup", map[string]any{"steps": steps, "status": "completed"})
	if st["status"] != "completed" || st["completedBy"] != "second" || st["nextStep"] != "review" || st["firstOpenedBy"] != "first" {
		t.Fatalf("completed = %v", st)
	}
	_, st, _ = h.apiAs("third", http.MethodPut, "/setup", map[string]any{"status": "skipped"})
	if st["status"] != "completed" || st["completedBy"] != "second" {
		t.Fatalf("later on a completed setup must keep it completed: %v", st)
	}
	_, st, _ = h.apiAs("third", http.MethodPut, "/setup", map[string]any{"steps": map[string]string{"sms": "pending"}})
	if st["nextStep"] != "sms" || st["status"] != "completed" {
		t.Fatalf("reopening a step = %v", st)
	}
	_, st, _ = h.apiAs("third", http.MethodPut, "/setup", map[string]any{"status": "pending"})
	if st["status"] != "pending" || st["firstOpenedAt"] != nil || st["completedAt"] != nil || len(st["steps"].(map[string]any)) != 0 {
		t.Fatalf("reset = %v", st)
	}
	_, out, _ = h.apiAs("fourth", http.MethodPost, "/setup/claim", nil)
	if out["claimed"] != true {
		t.Fatalf("claim after reset = %v", out)
	}
}

func TestSetupManualOpenStopsAutoOpen(t *testing.T) {
	h := newHarness(t)
	resetBrandSetup(h)
	_, st, _ := h.apiAs("curious", http.MethodPut, "/setup", map[string]any{"status": "in_progress"})
	if st["firstOpenedBy"] != "curious" {
		t.Fatalf("manual open = %v", st)
	}
	_, out, _ := h.apiAs("admin", http.MethodPost, "/setup/claim", nil)
	if out["claimed"] != false {
		t.Fatalf("claim after a manual open = %v", out)
	}
}

func TestBrandSaveValidateAndPublic(t *testing.T) {
	h := newHarness(t)
	resetBrandSetup(h)
	pub := h.mustAPI(http.StatusOK, http.MethodGet, "/brand/public", nil)
	if pub["custom"] != false || pub["accent"] != "" {
		t.Fatalf("default brand = %v", pub)
	}
	h.expectError(http.StatusBadRequest, "INVALID_COLOR", http.MethodPut, "/brand", map[string]any{"accent": "purple"})
	h.expectError(http.StatusBadRequest, "INVALID_URL", http.MethodPut, "/brand", map[string]any{"logoUrl": "javascript:alert(1)"})
	h.expectError(http.StatusBadRequest, "DEFAULT_LANGUAGE_REQUIRED", http.MethodPut, "/brand", map[string]any{"name": map[string]string{"en": "Star"}})
	h.expectError(http.StatusBadRequest, "INVALID_LANGUAGE", http.MethodPut, "/brand", map[string]any{"name": map[string]string{"fa": "ستاره", "xx": "?"}})
	h.expectError(http.StatusBadRequest, "TEXT_TOO_LONG", http.MethodPut, "/brand", map[string]any{"shortName": "abcdefghijklmnopqrstuvwxyz"})
	code, saved, _ := h.apiAs("boss", http.MethodPut, "/brand", map[string]any{
		"name": map[string]string{"fa": " ستاره‌چین ", "en": "Starpicker"}, "shortName": "ستاره", "accent": "#0EA",
		"logoUrl": "https://cdn.example.com/l.png", "logoDarkUrl": "/v2/economy/files/x.png", "markUrl": "", "welcome": map[string]string{"fa": "خوش آمدید"},
	})
	if code != http.StatusOK || saved["accent"] != "#00eeaa" || saved["updatedBy"] != "boss" || saved["custom"] != true {
		t.Fatalf("save = %d %v", code, saved)
	}
	code, pub, hdr := h.apiAs("", http.MethodGet, "/brand/public", nil)
	if code != http.StatusOK || pub["updatedBy"] != nil || pub["name"].(map[string]any)["fa"] != "ستاره‌چین" || hdr.Get("ETag") == "" || hdr.Get("Cache-Control") == "no-store" {
		t.Fatalf("public = %d %v %v", code, pub, hdr)
	}
	req := httptest.NewRequest(http.MethodGet, "/v2/console/base/brand/public", nil)
	req.Header.Set("If-None-Match", hdr.Get("ETag"))
	rec := httptest.NewRecorder()
	h.init.console.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Fatalf("revalidation = %d", rec.Code)
	}
	user := h.user("en")
	out := h.mustRPC(user, "base_brand", map[string]any{})
	if out["name"] != "Starpicker" || out["accent"] != "#00eeaa" || out["lang"] != "en" || out["custom"] != true {
		t.Fatalf("base_brand en = %v", out)
	}
	out = h.mustRPC(user, "base_brand", map[string]any{"lang": "fa"})
	if out["name"] != "ستاره‌چین" || out["welcome"] != "خوش آمدید" {
		t.Fatalf("base_brand fa = %v", out)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/brand", map[string]any{})
	out = h.mustRPC(user, "base_brand", map[string]any{})
	if out["custom"] != false || out["name"] != "" {
		t.Fatalf("reset brand = %v", out)
	}
}
