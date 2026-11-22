package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/handoffkit"
	"github.com/heroiclabs/nakama/v3/internal/samplecleanup"
	"google.golang.org/grpc/codes"
)

const (
	sampleCleanupRPC   = "base_sample_cleanup"
	CodeAdminOnly      = "CONSOLE_ADMIN_ONLY"
	CodeUnknownSample  = "SAMPLE_UNKNOWN"
	CodeConfirmNeeded  = "SAMPLE_CONFIRM_REQUIRED"
	CodeServerOnly     = "SERVER_ONLY"
	sampleConfirmValue = samplecleanup.Sample
)

func (s *service) registerSampleCleanup(initializer runtime.Initializer) error {
	if err := initializer.RegisterConsoleHttp(consolePrefix+"/sample-data/stargame", s.sampleCleanupGet, http.MethodGet); err != nil {
		return err
	}
	if err := initializer.RegisterConsoleHttp(consolePrefix+"/sample-data/stargame/cleanup", s.sampleCleanupPost, http.MethodPost); err != nil {
		return err
	}
	return initializer.RegisterRpc(sampleCleanupRPC, s.rpcSampleCleanup)
}

func samplePluginLoaded() bool {
	d, ok := gamecatalog.Lookup("stargame")
	return ok && d.Plugin == samplecleanup.Sample
}

func (s *service) sampleCleanupGet(w http.ResponseWriter, r *http.Request) {
	preview, err := samplecleanup.Run(r.Context(), s.db, samplecleanup.Options{Env: s.env, Now: s.now()})
	if err != nil {
		s.writeError(w, err)
		return
	}
	last, err := samplecleanup.LastRun(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sample": samplecleanup.Sample, "preview": preview, "lastRun": last,
		"pluginLoaded": samplePluginLoaded(), "admin": r.Header.Get(handoffkit.AdminHeader) == "true",
	})
}

func (s *service) sampleCleanupPost(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(handoffkit.AdminHeader) != "true" {
		s.writeError(w, fail(http.StatusForbidden, CodeAdminOnly, "", "only a console admin can remove the sample data"))
		return
	}
	var in struct {
		Confirm string `json:"confirm"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if strings.TrimSpace(in.Confirm) != sampleConfirmValue {
		s.writeError(w, badRequest(CodeConfirmNeeded, "confirm", "send confirm: stargame"))
		return
	}
	res, err := s.sampleCleanup(r.Context(), "console:"+consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": res, "pluginLoaded": samplePluginLoaded()})
}

func (s *service) rpcSampleCleanup(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string); userID != "" {
		return "", runtime.NewError(CodeServerOnly, int(codes.PermissionDenied))
	}
	var in struct {
		Sample string `json:"sample"`
		Apply  bool   `json:"apply"`
		By     string `json:"by"`
	}
	if strings.TrimSpace(payload) != "" {
		if err := json.Unmarshal([]byte(payload), &in); err != nil {
			return "", runtime.NewError("INVALID_JSON", int(codes.InvalidArgument))
		}
	}
	if in.Sample != "" && in.Sample != samplecleanup.Sample {
		return "", runtime.NewError(CodeUnknownSample, int(codes.InvalidArgument))
	}
	by := strings.TrimSpace(in.By)
	if len(by) > 64 {
		by = by[:64]
	}
	if by == "" {
		by = "server"
	}
	var res *samplecleanup.Result
	var err error
	if in.Apply {
		res, err = s.sampleCleanup(ctx, "rpc:"+by)
	} else {
		res, err = samplecleanup.Run(ctx, s.db, samplecleanup.Options{Env: s.env, Now: s.now()})
	}
	if err != nil {
		return "", s.runtimeError(err)
	}
	body, err := json.Marshal(map[string]any{"result": res, "pluginLoaded": samplePluginLoaded()})
	if err != nil {
		return "", runtime.NewError("INTERNAL_ERROR", int(codes.Internal))
	}
	return string(body), nil
}

func (s *service) sampleCleanup(ctx context.Context, by string) (*samplecleanup.Result, error) {
	res, err := samplecleanup.Run(ctx, s.db, samplecleanup.Options{Apply: true, By: by, Env: s.env, Now: s.now()})
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Star Catcher sample cleanup by %s failed: %v", by, err)
		}
		return nil, err
	}
	for _, id := range res.Leaderboards {
		if s.nk == nil {
			continue
		}
		if err := s.nk.LeaderboardDelete(ctx, id); err != nil && s.logger != nil {
			s.logger.Warn("Star Catcher sample cleanup: leaderboard %s not deleted: %v", id, err)
		}
	}
	if s.logger != nil {
		s.logger.Info("Star Catcher sample cleanup by %s: deleted %d rows (%d related player rows), kept %d, changed %d, leaderboards %v", by, res.Deleted, res.Related, res.Kept, res.Changed, res.Leaderboards)
	}
	return res, nil
}
