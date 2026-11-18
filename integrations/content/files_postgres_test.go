package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/content/internal/settings"
)

func TestFilesPackForWebClients(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) {
		c.Packs.Platforms = append(c.Packs.Platforms, "Web")
		c.Packs.RedirectToStorage = false
	})
	files := map[string][]byte{
		"Web/anim/hero.riv":      []byte("RIVE-binary"),
		"Web/img/board bg.webp":  []byte("RIFF....WEBPVP8 "),
		"Web/sfx/win.ogg":        []byte("OggS-data"),
		"Web/levels/level1.json": []byte(`{"words":["سلام"]}`),
	}
	if _, err := f.s.createPack(ctx, packCreateIn{Platform: "Web", Version: f.prefix + "-a", Files: upload{files: files}.list()}, "t"); errCode(err) != "PACK_CATALOG_MISSING" {
		t.Fatalf("addressables without a catalog = %v", err)
	}
	if _, err := f.s.createPack(ctx, packCreateIn{Kind: "zip", Platform: "Web", Version: f.prefix + "-z", Files: upload{files: files}.list()}, "t"); errCode(err) != "PACK_INVALID" {
		t.Fatalf("unknown kind = %v", err)
	}
	p := f.uploadKind(KindFiles, "Web", f.prefix+"-1", "", "", upload{files: files}, true)
	if p.Kind != KindFiles || p.CatalogPath != "" || p.FileCount != 4 || !p.Active {
		t.Fatalf("pack = %+v", p)
	}
	out, err := f.s.rpcCatalog(ctx, "", `{"platform":"Web"}`)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(out)
	var cat struct {
		Pack struct {
			Kind       string `json:"kind"`
			Version    string `json:"version"`
			CatalogURL string `json:"catalogUrl"`
			Files      []struct {
				Path, URL, SHA256, ContentType, Version string
				Size                                    int64
			} `json:"files"`
		} `json:"pack"`
	}
	if err := json.Unmarshal(raw, &cat); err != nil {
		t.Fatal(err)
	}
	if cat.Pack.Kind != KindFiles || cat.Pack.CatalogURL != "" || len(cat.Pack.Files) != 4 {
		t.Fatalf("catalog = %s", raw)
	}
	byPath := map[string]int{}
	for i, x := range cat.Pack.Files {
		byPath[x.Path] = i
	}
	img := cat.Pack.Files[byPath["img/board bg.webp"]]
	sum := sha256.Sum256(files["Web/img/board bg.webp"])
	if img.SHA256 != hex.EncodeToString(sum[:]) || img.Size != int64(len(files["Web/img/board bg.webp"])) || img.ContentType != "image/webp" || img.Version != f.prefix+"-1" ||
		img.URL != "http://cdn.test/v2/content/files/"+p.ID+"/img/board%20bg.webp" {
		t.Fatalf("webp entry = %+v", img)
	}
	for path, want := range map[string]string{"anim/hero.riv": "application/octet-stream", "img/board bg.webp": "image/webp", "sfx/win.ogg": "audio/ogg", "levels/level1.json": "application/json"} {
		r := httptest.NewRequest(http.MethodGet, "/v2/content/files/"+p.ID+"/x", nil)
		r = withVars(r, map[string]string{"pack": p.ID, "path": path})
		w := httptest.NewRecorder()
		f.s.httpContentFile(w, r)
		entry := cat.Pack.Files[byPath[path]]
		if w.Code != 200 || w.Header().Get("Content-Type") != want || w.Header().Get("ETag") != `"`+entry.SHA256+`"` || !strings.HasPrefix(w.Header().Get("Cache-Control"), "public, max-age=") {
			t.Fatalf("%s = %d %v", path, w.Code, w.Header())
		}
		if w.Body.String() != string(files["Web/"+path]) {
			t.Fatalf("%s body = %q", path, w.Body.String())
		}
		r2 := withVars(httptest.NewRequest(http.MethodGet, "/x", nil), map[string]string{"pack": p.ID, "path": path})
		r2.Header.Set("If-None-Match", `"`+entry.SHA256+`"`)
		w2 := httptest.NewRecorder()
		f.s.httpContentFile(w2, r2)
		if w2.Code != http.StatusNotModified {
			t.Fatalf("%s revalidate = %d", path, w2.Code)
		}
	}
}

func withVars(r *http.Request, vars map[string]string) *http.Request {
	for k, v := range vars {
		r.SetPathValue(k, v)
	}
	return r
}
