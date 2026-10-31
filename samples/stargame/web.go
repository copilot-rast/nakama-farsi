package main

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

//go:embed web
var webFiles embed.FS

const webPrefix = "/stargame/"

var contentTypes = map[string]string{
	".html":  "text/html; charset=utf-8",
	".css":   "text/css; charset=utf-8",
	".js":    "text/javascript; charset=utf-8",
	".json":  "application/json; charset=utf-8",
	".woff2": "font/woff2",
	".txt":   "text/plain; charset=utf-8",
	".svg":   "image/svg+xml",
	".png":   "image/png",
	".ico":   "image/x-icon",
}

type webFile struct {
	body        []byte
	contentType string
	etag        string
	cache       string
}

type webSite struct {
	files     map[string]webFile
	serverKey string
	started   time.Time
}

func newWebSite(serverKey string) (*webSite, error) {
	root, err := fs.Sub(webFiles, "web")
	if err != nil {
		return nil, err
	}
	site := &webSite{files: map[string]webFile{}, serverKey: serverKey, started: time.Now()}
	err = fs.WalkDir(root, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ct, ok := contentTypes[path.Ext(name)]
		if !ok {
			return nil
		}
		body, err := fs.ReadFile(root, name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		cache := "no-cache"
		if strings.HasPrefix(name, "fonts/") {
			cache = "public, max-age=604800"
		}
		site.files[name] = webFile{body: body, contentType: ct, etag: `"` + hex.EncodeToString(sum[:8]) + `"`, cache: cache}
		return nil
	})
	return site, err
}

func (s *webSite) register(initializer runtime.Initializer) error {
	if err := initializer.RegisterHttp("/stargame", s.redirect, http.MethodGet, http.MethodHead); err != nil {
		return err
	}
	return initializer.RegisterHttp(webPrefix+"{file:.*}", s.serve, http.MethodGet, http.MethodHead)
}

func (s *webSite) redirect(w http.ResponseWriter, r *http.Request) {
	target := webPrefix
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

func (s *webSite) serve(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, webPrefix)
	if name == "" {
		name = "index.html"
	}
	if name == "config.json" {
		s.serveConfig(w, r)
		return
	}
	f, ok := s.files[name]
	if !ok || path.Clean(name) != name {
		w.Header().Set("Cache-Control", "no-store")
		http.NotFound(w, r)
		return
	}
	h := w.Header()
	h.Set("Content-Type", f.contentType)
	h.Set("Cache-Control", f.cache)
	h.Set("ETag", f.etag)
	h.Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, name, s.started, bytes.NewReader(f.body))
}

func (s *webSite) serveConfig(w http.ResponseWriter, r *http.Request) {
	body, _ := json.Marshal(map[string]string{"serverKey": s.serverKey})
	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(body)
}
