package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/content/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/contenthooks"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/content"
	publicPrefix  = "/v2/content"
	maxPayload    = 8 * 1024
)

type service struct {
	db       *sql.DB
	logger   runtime.Logger
	store    objectstore.Store
	baseURL  string
	now      func() time.Time
	packFile *fileCache
	bg       context.Context
	holder   string
	asm      *assembler

	assembleHook func(sha string)
}

func newService(db *sql.DB, logger runtime.Logger, store objectstore.Store, baseURL string) *service {
	return &service{db: db, logger: logger, store: store, baseURL: strings.TrimRight(baseURL, "/"), now: time.Now, packFile: newFileCache(64),
		bg: context.Background(), holder: clusterkit.Holder(context.Background()), asm: newAssembler()}
}

func storageDir(env map[string]string, args []string) string {
	if dir := strings.TrimSpace(env["CONTENT_UPLOAD_DIR"]); dir != "" {
		return dir
	}
	if dir := strings.TrimSpace(env["ECONOMY_UPLOAD_DIR"]); dir != "" {
		return filepath.Join(dir, "content")
	}
	return filepath.Join(dataDirFromArgs(args), "economy-uploads", "content")
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	cfg, err := objectstore.ConfigFromEnv(env, objectstore.Namespace{LocalDir: storageDir(env, os.Args), Prefix: "content/"})
	if err != nil {
		logger.Error("Content module refused to start: %v", err)
		return err
	}
	if cfg.Driver == objectstore.DriverLocal {
		if err := os.MkdirAll(cfg.LocalDir, 0o750); err != nil {
			return err
		}
	}
	store, err := objectstore.New(cfg, nil)
	if err != nil {
		return err
	}
	base := strings.TrimSpace(env["CONTENT_PUBLIC_BASE_URL"])
	if base == "" {
		base = strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"])
	}
	svc := newService(db, logger, store, base)
	svc.bg = ctx
	svc.holder = clusterkit.Holder(ctx)
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, apierr.RPC(db, apierr.RPCOptions{Logger: logger, MaxPayload: maxPayload}, fn)); err != nil {
			return err
		}
	}
	if err := svc.registerPublicHTTP(initializer); err != nil {
		return err
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	contenthooks.SetCatalog(svc.catalogHook)
	go svc.loop(ctx)
	go svc.assembleLoop(ctx)
	logger.Info("Content module registered (%d client RPCs, packs in %s storage)", len(svc.clientRPCs()), store.Driver())
	return nil
}

func (s *service) clientRPCs() map[string]apierr.Handler {
	return map[string]apierr.Handler{
		"content_catalog": s.rpcCatalog,
	}
}

func (s *service) registerPublicHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	p := publicPrefix
	routes := []route{
		{p + "/catalog", http.MethodGet, s.httpContentIndex},
		{p + "/files/{pack}/{path:.+}", http.MethodGet, s.httpContentFile},
		{p + "/files/{pack}/{path:.+}", http.MethodHead, s.httpContentFile},
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
		s.logger.Warn("Content settings not read, using defaults: %v", err)
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
			if ok, err := clusterkit.AcquireLease(ctx, s.db, "content-sweep", holder, 9*time.Minute); err != nil || !ok {
				continue
			}
			if err := s.sweepPacks(ctx, s.settings(ctx)); err != nil && s.logger != nil {
				s.logger.Warn("Content pack sweep failed: %v", err)
			}
		}
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
		s.logger.Error("Content console request failed: %v", err)
	}
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}

func (s *service) httpFail(w http.ResponseWriter, r *http.Request, err error) {
	if e, ok := apierr.As(err); !ok || e.Status >= 500 {
		if s.logger != nil {
			s.logger.Error("Content request %s failed: %v", r.URL.Path, err)
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

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }

func (s *service) catalogHook(ctx context.Context, platform, appVersion string) (map[string]any, error) {
	raw, _ := json.Marshal(indexIn{Platform: platform, AppVersion: appVersion})
	out, err := s.rpcCatalog(ctx, "", string(raw))
	if err != nil {
		return nil, err
	}
	m, _ := out.(map[string]any)
	return m, nil
}
