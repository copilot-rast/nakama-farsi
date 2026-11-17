package apidocs

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"embed"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

//go:embed ui/swagger-ui-bundle.js.gz ui/swagger-ui.css.gz
var assets embed.FS

type Options struct {
	DB         *sql.DB
	SchemaFile string
	RPCIDs     func() []string
	Authorized func(r *http.Request) bool
}

type Handler struct {
	opts  Options
	mu    sync.Mutex
	cache map[string][]byte
	mtime time.Time
	count int
}

func NewHandler(opts Options) *Handler {
	return &Handler{opts: opts, cache: map[string][]byte{}}
}

const page = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>API docs</title>
<link rel="stylesheet" href="/docs/ui/swagger-ui.css">
<style>body{margin:0}.swagger-ui .topbar{display:none}</style>
</head>
<body>
<div id="docs"></div>
<script src="/docs/ui/swagger-ui-bundle.js"></script>
<script src="/docs/ui/init.js"></script>
</body>
</html>
`

const initJS = `window.ui = SwaggerUIBundle({
  url: "/docs/openapi.json",
  dom_id: "#docs",
  deepLinking: true,
  validatorUrl: null,
  persistAuthorization: true,
  displayOperationId: true,
  docExpansion: "none",
  filter: true,
  tryItOutEnabled: false
});
`

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	set := Load(r.Context(), h.opts.DB)
	if !set.Enabled {
		http.NotFound(w, r)
		return
	}
	if set.ConsoleLogin && (h.opts.Authorized == nil || !h.opts.Authorized(r)) {
		w.Header().Set("WWW-Authenticate", `Basic realm="console"`)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Console login required.","message":"Console login required.","code":16}`))
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	switch strings.TrimSuffix(r.URL.Path, "/") {
	case "/docs":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'")
		_, _ = io.WriteString(w, page)
	case "/docs/ui/init.js":
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		_, _ = io.WriteString(w, initJS)
	case "/docs/ui/swagger-ui-bundle.js":
		h.asset(w, r, "ui/swagger-ui-bundle.js.gz", "text/javascript; charset=utf-8")
	case "/docs/ui/swagger-ui.css":
		h.asset(w, r, "ui/swagger-ui.css.gz", "text/css; charset=utf-8")
	case "/docs/openapi.json":
		body, err := h.Document(serverURL(r))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"API document unavailable.","message":"API document unavailable.","code":13}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(body)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) asset(w http.ResponseWriter, r *http.Request, name, kind string) {
	raw, err := assets.ReadFile(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", kind)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Vary", "Accept-Encoding")
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(raw)
		return
	}
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = io.Copy(w, zr)
}

func (h *Handler) Document(server string) ([]byte, error) {
	var raw []byte
	var mtime time.Time
	if h.opts.SchemaFile != "" {
		if st, err := os.Stat(h.opts.SchemaFile); err == nil {
			mtime = st.ModTime()
		}
	}
	var ids []string
	if h.opts.RPCIDs != nil {
		ids = h.opts.RPCIDs()
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !mtime.Equal(h.mtime) || len(ids) != h.count {
		h.cache = map[string][]byte{}
		h.mtime, h.count = mtime, len(ids)
	}
	if got, ok := h.cache[server]; ok {
		return got, nil
	}
	if !mtime.IsZero() {
		b, err := os.ReadFile(h.opts.SchemaFile)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	out, err := Document(raw, ids, server)
	if err != nil {
		return nil, err
	}
	if len(h.cache) > 16 {
		h.cache = map[string][]byte{}
	}
	h.cache[server] = out
	return out, nil
}

func serverURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p == "https" || p == "http" {
		scheme = p
	}
	host := r.Host
	if fh := r.Header.Get("X-Forwarded-Host"); fh != "" {
		host = strings.TrimSpace(strings.Split(fh, ",")[0])
	}
	if host == "" || strings.ContainsAny(host, " /\\\"'<>") {
		return "/"
	}
	return scheme + "://" + host
}
