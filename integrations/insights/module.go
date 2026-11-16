package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/insights"
	maxPayload    = 256 * 1024
)

type service struct {
	db     *sql.DB
	logger runtime.Logger
	store  objectstore.Store
	now    func() time.Time
}

func storageDir(env map[string]string, args []string) string {
	if dir := strings.TrimSpace(env["INSIGHTS_UPLOAD_DIR"]); dir != "" {
		return dir
	}
	if dir := strings.TrimSpace(env["ECONOMY_UPLOAD_DIR"]); dir != "" {
		return filepath.Join(dir, "insights")
	}
	return filepath.Join(dataDirFromArgs(args), "economy-uploads", "insights")
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	cfg, err := objectstore.ConfigFromEnv(env, objectstore.Namespace{LocalDir: storageDir(env, os.Args), Prefix: "insights/"})
	if err != nil {
		logger.Error("Insights module refused to start: %v", err)
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
	svc := &service{db: db, logger: logger, store: store, now: time.Now}
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, apierr.RPC(db, apierr.RPCOptions{Logger: logger, MaxPayload: maxPayload}, fn)); err != nil {
			return err
		}
	}
	if err := initializer.RegisterEventSessionStart(svc.onSocketStart); err != nil {
		return err
	}
	if err := initializer.RegisterEventSessionEnd(svc.onSocketEnd); err != nil {
		return err
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	go svc.loop(ctx)
	logger.Info("Insights module registered (%d client RPCs, heatmap traces in %s storage)", len(svc.clientRPCs()), store.Driver())
	return nil
}

func (s *service) clientRPCs() map[string]apierr.Handler {
	return map[string]apierr.Handler{
		"insights_config":  s.rpcConfig,
		"insights_session": s.rpcSession,
		"insights_heatmap": s.rpcHeatmap,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Insights settings not read, using defaults: %v", err)
	}
	return cfg
}

func (s *service) rpcConfig(ctx context.Context, _ string, _ string) (any, error) {
	cfg := s.settings(ctx)
	maps := []string{}
	rows, err := s.db.QueryContext(ctx, `SELECT key FROM insights_heatmap_map WHERE enabled ORDER BY key LIMIT 500`)
	if err == nil {
		for rows.Next() {
			var k string
			if rows.Scan(&k) == nil {
				maps = append(maps, k)
			}
		}
		rows.Close()
	}
	return map[string]any{
		"sessions": map[string]any{"source": cfg.Sessions.Source, "pingSeconds": cfg.Sessions.PingSeconds, "timeoutSeconds": cfg.Sessions.TimeoutSeconds},
		"heatmap": map[string]any{"enabled": cfg.Heatmap.Enabled, "maxPoints": cfg.Heatmap.MaxPoints, "maxBytes": cfg.Heatmap.MaxBytes,
			"unknownMaps": cfg.Heatmap.UnknownMaps, "maps": maps, "format": "t,x,z,player,kmh"},
	}, nil
}

func (s *service) loop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	n := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if ok, err := clusterkit.AcquireLease(ctx, s.db, "insights-worker", holder, 50*time.Second); err != nil || !ok {
			continue
		}
		cfg := s.settings(ctx)
		if _, err := s.closeIdle(ctx, cfg); err != nil && s.logger != nil {
			s.logger.Warn("Insights idle session close failed: %v", err)
		}
		if n++; n%60 == 1 {
			s.sweep(ctx, cfg)
		}
	}
}

func (s *service) sweep(ctx context.Context, cfg settings.All) {
	if d := cfg.Sessions.KeepDays; d > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM insights_session WHERE ctid IN (SELECT ctid FROM insights_session WHERE day < current_date - $1::int LIMIT 50000)`, d)
		_, _ = s.db.ExecContext(ctx, `DELETE FROM insights_player_day WHERE ctid IN (SELECT ctid FROM insights_player_day WHERE day < current_date - $1::int LIMIT 50000)`, d)
	}
	if d := cfg.Heatmap.KeepCellsDays; d > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM insights_heatmap_cell WHERE ctid IN (SELECT ctid FROM insights_heatmap_cell WHERE day < current_date - $1::int LIMIT 50000)`, d)
	}
	if d := cfg.Heatmap.KeepRawDays; d > 0 {
		s.dropRaw(ctx, `SELECT id::text, raw_key FROM insights_heatmap_upload WHERE raw_key <> '' AND created_at < now() - make_interval(days => $1) LIMIT 2000`, d)
	}
}

func (s *service) dropRaw(ctx context.Context, q string, args ...any) {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return
	}
	type item struct{ id, key string }
	list := []item{}
	for rows.Next() {
		var it item
		if rows.Scan(&it.id, &it.key) == nil {
			list = append(list, it)
		}
	}
	rows.Close()
	for _, it := range list {
		_ = s.store.Delete(ctx, it.key)
		_, _ = s.db.ExecContext(ctx, `UPDATE insights_heatmap_upload SET raw_key='' WHERE id=$1`, it.id)
	}
}

func (s *service) onSocketStart(ctx context.Context, _ runtime.Logger, _ *api.Event) {
	if s.settings(ctx).Sessions.Source != settings.SourceSocket {
		return
	}
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	sid, _ := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	vars, _ := ctx.Value(runtime.RUNTIME_CTX_VARS).(map[string]string)
	if userID == "" || sid == "" {
		return
	}
	in := sessionIn{Event: "start", SessionID: "socket:" + sid, AppVersion: vars["appVersion"], Platform: vars["platform"]}
	if _, err := s.session(ctx, userID, in, settings.SourceSocket); err != nil && s.logger != nil {
		s.logger.Debug("Insights socket session start for %s not stored: %v", userID, err)
	}
}

func (s *service) onSocketEnd(ctx context.Context, _ runtime.Logger, _ *api.Event) {
	if s.settings(ctx).Sessions.Source != settings.SourceSocket {
		return
	}
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	sid, _ := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	if userID == "" || sid == "" {
		return
	}
	if _, err := s.session(ctx, userID, sessionIn{Event: "end", SessionID: "socket:" + sid}, settings.SourceSocket); err != nil && s.logger != nil {
		s.logger.Debug("Insights socket session end for %s not stored: %v", userID, err)
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
		s.logger.Error("Insights console request failed: %v", err)
	}
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
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
