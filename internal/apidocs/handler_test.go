package apidocs

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func get(h http.Handler, path string, hdr map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "http://api.test"+path, nil)
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestDocsPageUsesOnlyOwnAssets(t *testing.T) {
	h := NewHandler(Options{RPCIDs: func() []string { return []string{"x_ping"} }})
	page := get(h, "/docs", nil)
	if page.Code != 200 || !strings.Contains(page.Body.String(), "/docs/ui/swagger-ui-bundle.js") {
		t.Fatalf("page %d", page.Code)
	}
	for _, m := range regexp.MustCompile(`(?:src|href)="([^"]+)"`).FindAllStringSubmatch(page.Body.String(), -1) {
		if !strings.HasPrefix(m[1], "/docs/") {
			t.Errorf("external asset %s", m[1])
		}
		if got := get(h, m[1], nil); got.Code != 200 || got.Body.Len() < 100 {
			t.Errorf("%s: %d %d bytes", m[1], got.Code, got.Body.Len())
		}
	}
	js := get(h, "/docs/ui/init.js", nil).Body.String()
	if !strings.Contains(js, "validatorUrl: null") || strings.Contains(js, "http") {
		t.Errorf("init script may reach outside: %s", js)
	}
	gz := get(h, "/docs/ui/swagger-ui.css", map[string]string{"Accept-Encoding": "gzip"})
	if gz.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("gzip not served")
	}
	zr, err := gzip.NewReader(gz.Body)
	if err != nil {
		t.Fatal(err)
	}
	css, _ := io.ReadAll(zr)
	if regexp.MustCompile(`url\(["']?https?:`).Match(css) || strings.Contains(string(css), "@import") {
		t.Error("css loads remote files")
	}
	doc := get(h, "/docs/openapi.json", nil)
	if doc.Code != 200 || !strings.Contains(doc.Body.String(), `"/v2/rpc/x_ping"`) {
		t.Fatalf("doc without a schema file: %d %s", doc.Code, doc.Body.String())
	}
}

func TestDocsLoginSwitch(t *testing.T) {
	store.Defaults = func() rulesRow { return rulesRow{Docs: Settings{Enabled: true, ConsoleLogin: true}} }
	defer func() { store.Defaults = func() rulesRow { return rulesRow{Docs: DefaultSettings()} } }()
	h := NewHandler(Options{Authorized: func(r *http.Request) bool { return r.Header.Get("Authorization") == "Bearer console" }})
	if got := get(h, "/docs/openapi.json", nil); got.Code != http.StatusUnauthorized || got.Header().Get("WWW-Authenticate") == "" {
		t.Fatalf("no login asked: %d", got.Code)
	}
	if got := get(h, "/docs/openapi.json", map[string]string{"Authorization": "Bearer console"}); got.Code != 200 {
		t.Fatalf("console login refused: %d", got.Code)
	}
	store.Defaults = func() rulesRow { return rulesRow{Docs: Settings{Enabled: false}} }
	if got := get(h, "/docs", map[string]string{"Authorization": "Bearer console"}); got.Code != http.StatusNotFound {
		t.Fatalf("docs off still served: %d", got.Code)
	}
}

func TestDefaultsAreOnWithoutLogin(t *testing.T) {
	d := DefaultSettings()
	if !d.Enabled || d.ConsoleLogin {
		t.Fatal(d)
	}
}
