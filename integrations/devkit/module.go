package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

//go:generate go run -mod=vendor ./cmd/devkitgen -root ../..

//go:embed schema/rpc-schema.json
var rpcSchemaJSON []byte

const (
	consolePrefix = "/v2/console/base/devkit"
	publicPrefix  = "/v2/devkit"
	ScopeSchema   = "schema:read"
)

type clientRPC = apierr.Handler

type service struct {
	db     *sql.DB
	logger runtime.Logger
	now    func() time.Time
	schema schemaDoc
	// store keeps the Unity package archives (nil: the registry answers unavailable).
	store objectstore.Store
	// publicBase is the API address from the environment (DEVKIT_PUBLIC_BASE_URL, else the assets or
	// economy one), used for the registry address the console shows.
	publicBase string
}

type schemaDoc struct {
	raw  map[string]any
	rpcs []map[string]any
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	svc := &service{db: db, logger: logger, now: time.Now}
	if err := svc.loadSchema(); err != nil {
		return err
	}
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, apierr.RPC(db, apierr.RPCOptions{Logger: logger}, fn)); err != nil {
			return err
		}
	}
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	for _, k := range []string{"DEVKIT_PUBLIC_BASE_URL", "ASSETS_PUBLIC_BASE_URL", "ECONOMY_PUBLIC_BASE_URL"} {
		if v := strings.TrimRight(strings.TrimSpace(env[k]), "/"); v != "" {
			svc.publicBase = v
			break
		}
	}
	if store, err := packageStore(env); err != nil {
		logger.Warn("Devkit package registry has no storage: %v", err)
	} else {
		svc.store = store
	}
	if err := apikey.RegisterScopes(ctx, db, "devkit", map[string]string{ScopeSchema: "Read the client RPC schema",
		ScopeUpmRead: "Install Unity packages from the registry (/upm)", ScopeUpmPublish: "Publish Unity packages to the registry"}); err != nil {
		logger.Warn("Devkit could not register its API scopes: %v", err)
	}
	if err := svc.registerPublicHTTP(initializer); err != nil {
		return err
	}
	if err := svc.registerPackageHTTP(initializer); err != nil {
		return err
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	logger.Info("Devkit module registered (%d error codes, %d client RPCs in the schema)", len(apierr.Defaults()), len(svc.schema.rpcs))
	return nil
}

func (s *service) loadSchema() error {
	var doc map[string]any
	if err := json.Unmarshal(rpcSchemaJSON, &doc); err != nil {
		return err
	}
	s.schema.raw = doc
	list, _ := doc["rpcs"].([]any)
	for _, r := range list {
		if m, ok := r.(map[string]any); ok {
			s.schema.rpcs = append(s.schema.rpcs, m)
		}
	}
	return nil
}

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"devkit_error_catalog": s.rpcErrorCatalog,
		"devkit_rpc_schema":    s.rpcSchema,
	}
}

type catalogItem struct {
	Status            int    `json:"status"`
	Message           string `json:"message"`
	RetryAfterSeconds *int   `json:"retryAfterSeconds,omitempty"`
}

type catalogView struct {
	Lang         string                 `json:"lang"`
	FallbackLang string                 `json:"fallbackLang"`
	Version      string                 `json:"version"`
	CacheSeconds int                    `json:"cacheSeconds"`
	Mode         string                 `json:"mode"`
	ModuleMode   string                 `json:"moduleMode"`
	Errors       map[string]catalogItem `json:"errors"`
}

func (s *service) catalog(ctx context.Context, lang string, only []string) (catalogView, error) {
	cfg, _ := settings.Load(ctx, s.db)
	entries, version, err := apierr.Catalog(ctx, s.db)
	if err != nil {
		return catalogView{}, err
	}
	fallback := apierr.FallbackLang(ctx, s.db)
	view := catalogView{Lang: lang, FallbackLang: fallback, Version: version + "-" + lang, CacheSeconds: cfg.Errors.CacheSeconds, Mode: cfg.Errors.Mode, ModuleMode: cfg.Errors.ModuleMode, Errors: map[string]catalogItem{}}
	want := map[string]bool{}
	for _, c := range only {
		if c = strings.TrimSpace(c); c != "" {
			want[c] = true
		}
	}
	for code, e := range entries {
		if len(want) > 0 && !want[code] {
			continue
		}
		view.Errors[code] = catalogItem{Status: e.Status, Message: e.MessageFor(lang, fallback), RetryAfterSeconds: e.RetryAfter}
	}
	return view, nil
}

func (s *service) rpcErrorCatalog(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang  string   `json:"lang"`
		Codes []string `json:"codes"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if len(in.Codes) > 500 {
		return nil, apierr.BadRequest("TOO_MANY_CODES").With("max", 500)
	}
	lang := apierr.ResolveLang(ctx, s.db, userID, in.Lang, "")
	return s.catalog(ctx, lang, in.Codes)
}

func (s *service) schemaView(module string) (map[string]any, error) {
	if module == "" {
		return s.schema.raw, nil
	}
	out := map[string]any{}
	for k, v := range s.schema.raw {
		out[k] = v
	}
	list := []map[string]any{}
	for _, r := range s.schema.rpcs {
		if r["module"] == module {
			list = append(list, r)
		}
	}
	if len(list) == 0 {
		return nil, apierr.NotFound("DEVKIT_UNKNOWN_MODULE").With("module", module)
	}
	out["rpcs"] = list
	return out, nil
}

func (s *service) rpcSchema(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Module string `json:"module"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg, _ := settings.Load(ctx, s.db)
	if cfg.Schema.Access != settings.AccessPublic && cfg.Schema.Access != settings.AccessSession {
		return nil, apierr.Forbidden("DEVKIT_SCHEMA_DISABLED")
	}
	return s.schemaView(strings.TrimSpace(in.Module))
}

func (s *service) registerPublicHTTP(initializer runtime.Initializer) error {
	if err := initializer.RegisterHttp(publicPrefix+"/errors", s.httpErrors, http.MethodGet, http.MethodHead); err != nil {
		return err
	}
	if err := initializer.RegisterHttp(publicPrefix+"/rpc-schema", s.httpSchema, http.MethodGet); err != nil {
		return err
	}
	return initializer.RegisterHttp(publicPrefix+"/key", s.httpKey, http.MethodGet)
}

func (s *service) httpErrors(w http.ResponseWriter, r *http.Request) {
	cfg, _ := settings.Load(r.Context(), s.db)
	if !cfg.Errors.PublicCatalog {
		apierr.WriteHTTP(w, r, s.db, apierr.Forbidden("DEVKIT_CATALOG_PRIVATE"))
		return
	}
	lang := apierr.ResolveLang(r.Context(), s.db, "", r.URL.Query().Get("lang"), r.Header.Get("Accept-Language"))
	var only []string
	if c := r.URL.Query().Get("codes"); c != "" {
		only = strings.Split(c, ",")
	}
	view, err := s.catalog(r.Context(), lang, only)
	if err != nil {
		s.logger.Error("Devkit error catalog failed: %v", err)
		apierr.WriteHTTP(w, r, s.db, err)
		return
	}
	etag := `"` + view.Version + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(cfg.Errors.CacheSeconds))
	w.Header().Set("Vary", "Accept-Language")
	if len(only) == 0 && r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *service) httpSchema(w http.ResponseWriter, r *http.Request) {
	cfg, _ := settings.Load(r.Context(), s.db)
	if cfg.Schema.Access == settings.AccessConsole {
		apierr.WriteHTTP(w, r, s.db, apierr.Forbidden("DEVKIT_SCHEMA_DISABLED"))
		return
	}
	serve := func(w http.ResponseWriter, r *http.Request) {
		view, err := s.schemaView(strings.TrimSpace(r.URL.Query().Get("module")))
		if err != nil {
			apierr.WriteHTTP(w, r, s.db, err)
			return
		}
		writeJSON(w, http.StatusOK, view)
	}
	if cfg.Schema.Access == settings.AccessPublic && apikey.FromRequest(r) == "" {
		serve(w, r)
		return
	}
	apikey.Require(s.db, ScopeSchema, func(w http.ResponseWriter, r *http.Request, _ *apikey.Key) { serve(w, r) })(w, r)
}

func (s *service) httpKey(w http.ResponseWriter, r *http.Request) {
	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	apikey.Require(s.db, scope, func(w http.ResponseWriter, r *http.Request, key *apikey.Key) {
		writeJSON(w, http.StatusOK, key)
	})(w, r)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	if w.Header().Get("Cache-Control") == "" {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}
