package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/assets/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
)

func TestEditorPackActivateCheckedBeforeUpload(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) { a.Packs.ActivateFromEditor = true })
	call := func(scopes []string, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		f.s.editorPackCreate(w, httptest.NewRequest(http.MethodPost, "/packs", strings.NewReader(body)), &apikey.Key{ID: "k", Name: "k", Scopes: scopes})
		return w
	}
	w := call([]string{ScopeContent}, `{"platform":"Android","version":"`+f.prefix+`-act","files":[],"activate":true}`)
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "PACK_ACTIVATE_DENIED") {
		t.Fatalf("content:write key asking to activate = %d %s", w.Code, w.Body.String())
	}
	var n int
	if err := f.s.db.QueryRow(`SELECT count(*) FROM assets_pack WHERE version=$1`, f.prefix+"-act").Scan(&n); err != nil || n != 0 {
		t.Fatalf("pack created before the activate check: %d %v", n, err)
	}
}
