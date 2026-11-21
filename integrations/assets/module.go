// Package main is the assets plugin: a versioned store of typed JSON entries
// (ScriptableObject data and the like) that editor tools push and game clients
// read as a manifest and a delta fetch, and remote content packs (Addressables
// catalogs and bundles) uploaded per platform and served to the game through a
// content index. Nothing in it names a game.
package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/assets/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/assets"
	publicPrefix  = "/v2/assets"
	maxPayload    = 64 * 1024

	ScopeRead    = "assets:read"
	ScopeWrite   = "assets:write"
	ScopeDelete  = "assets:delete"
	ScopeContent = "content:write"
	ScopePublish = "content:publish"
)

type clientRPC = apierr.Handler

type service struct {
	db       *sql.DB
	logger   runtime.Logger
	store    objectstore.Store
	baseURL  string
	now      func() time.Time
	packFile *fileCache
}

func newService(db *sql.DB, logger runtime.Logger, store objectstore.Store, baseURL string) *service {
	return &service{db: db, logger: logger, store: store, baseURL: strings.TrimRight(baseURL, "/"), now: time.Now, packFile: newFileCache(64)}
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	store, err := objectstore.FromEnv(env, objectstore.Namespace{LocalDir: filepath.Join(dataDirFromArgs(os.Args), "assets-content"), Prefix: "assets/"})
	if err != nil {
		return err
	}
	base := strings.TrimSpace(env["ASSETS_PUBLIC_BASE_URL"])
	if base == "" {
		base = strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"])
	}
	svc := newService(db, logger, store, base)
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, apierr.RPC(db, apierr.RPCOptions{Logger: logger, MaxPayload: maxPayload}, fn)); err != nil {
			return err
		}
	}
	if err := apikey.RegisterScopes(ctx, db, "assets", map[string]string{
		ScopeRead:    "Read data entries (manifest and fetch) from an editor tool",
		ScopeWrite:   "Push data entries from an editor tool",
		ScopeDelete:  "Delete data entries in a push",
		ScopeContent: "Upload content packs (remote Addressables)",
		ScopePublish: "Activate an uploaded content pack from the editor",
	}); err != nil {
		return err
	}
	if err := svc.registerPublicHTTP(initializer); err != nil {
		return err
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	go svc.loop(ctx)
	logger.Info("Assets module registered (%d client RPCs, content in %s storage)", len(svc.clientRPCs()), store.Driver())
	return nil
}

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"assets_manifest":      s.rpcManifest,
		"assets_fetch":         s.rpcFetch,
		"assets_content_index": s.rpcContentIndex,
	}
}

func (s *service) registerPublicHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	p := publicPrefix
	routes := []route{
		{p + "/manifest", http.MethodGet, s.httpManifest},
		{p + "/fetch", http.MethodPost, s.httpFetch},
		{p + "/content-index", http.MethodGet, s.httpContentIndex},
		{p + "/content/{pack}/{path:.+}", http.MethodGet, s.httpContentFile},
		{p + "/content/{pack}/{path:.+}", http.MethodHead, s.httpContentFile},
		{p + "/editor/limits", http.MethodGet, apikey.Require(s.db, ScopeRead, s.editorLimits)},
		{p + "/editor/manifest", http.MethodGet, apikey.Require(s.db, ScopeRead, s.editorManifest)},
		{p + "/editor/fetch", http.MethodPost, apikey.Require(s.db, ScopeRead, s.editorFetch)},
		{p + "/editor/push", http.MethodPost, apikey.Require(s.db, ScopeWrite, s.editorPush)},
		{p + "/editor/packs", http.MethodGet, apikey.Require(s.db, ScopeContent, s.editorPackList)},
		{p + "/editor/packs", http.MethodPost, apikey.Require(s.db, ScopeContent, s.editorPackCreate)},
		{p + "/editor/packs/{id}/blobs/{sha}/{part}", http.MethodPut, apikey.Require(s.db, ScopeContent, s.editorPackPart)},
		{p + "/editor/packs/{id}/complete", http.MethodPost, apikey.Require(s.db, ScopeContent, s.editorPackComplete)},
	}
	for _, r := range routes {
		if err := initializer.RegisterHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Assets settings not read, using defaults: %v", err)
	}
	return cfg
}

func (s *service) loop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ok, err := clusterkit.AcquireLease(ctx, s.db, "assets-sweep", holder, 9*time.Minute); err != nil || !ok {
				continue
			}
			s.sweep(ctx)
		}
	}
}

// sweep trims entry history, removes abandoned uploads, old inactive packs and files no pack uses.
func (s *service) sweep(ctx context.Context) {
	cfg := s.settings(ctx)
	s.trimHistory(ctx, cfg)
	if err := s.sweepPacks(ctx, cfg); err != nil && s.logger != nil {
		s.logger.Warn("Assets pack sweep failed: %v", err)
	}
}

func (s *service) fail(w http.ResponseWriter, err error) {
	if pluginkit.WriteError(w, err) {
		return
	}
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		pluginkit.WriteJSON(w, e.Status, map[string]any{"error": e.Code, "message": e.Code, "field": e.Field, "meta": e.Meta})
		return
	}
	if s.logger != nil {
		s.logger.Error("Assets console request failed: %v", err)
	}
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}

func (s *service) httpFail(w http.ResponseWriter, r *http.Request, err error) {
	if e, ok := apierr.As(err); !ok || e.Status >= 500 {
		if s.logger != nil {
			s.logger.Error("Assets request %s failed: %v", r.URL.Path, err)
		}
	}
	apierr.WriteHTTP(w, r, s.db, err)
}

func dataDirFromArgs(args []string) string {
	for i, arg := range args {
		for _, flag := range []string{"--data_dir", "-data_dir"} {
			if arg == flag && i+1 < len(args) {
				return args[i+1]
			}
			if strings.HasPrefix(arg, flag+"=") {
				return strings.TrimPrefix(arg, flag+"=")
			}
		}
	}
	return "./data"
}

var errNoRows = sql.ErrNoRows

func isNoRows(err error) bool { return errors.Is(err, errNoRows) }
