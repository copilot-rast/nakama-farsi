package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/eventcontext"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	consolePrefix = "/v2/console/base/campaign"
	maxPayload    = 16 * 1024
	sourceType    = "stage"
)

type clientRPC = apierr.Handler

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	now    func() time.Time
	level  func(ctx context.Context, userID, tableID string) (int, error)
	prize  func(ctx context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error)
	spend  func(ctx context.Context, userID, currency string, amount int64, sourceID string) (economyhooks.RegenSpend, error)
	refund func(ctx context.Context, userID, currency, sourceID string) (economyhooks.RegenRefund, error)
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger) *service {
	return &service{db: db, nk: nk, logger: logger, now: time.Now, level: economyhooks.Level, prize: economyhooks.GrantPrize,
		spend: economyhooks.SpendRegen, refund: economyhooks.RefundRegen}
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	svc := newService(db, nk, logger)
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, apierr.RPC(db, apierr.RPCOptions{Logger: logger, MaxPayload: maxPayload}, fn)); err != nil {
			return err
		}
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	campaignhooks.SetCampaign(svc.hookStart, svc.hookComplete)
	eventcontext.Register(eventcontext.Provider{Name: "campaign", Sources: contextSources, Spans: svc.contextSpans})
	go svc.loop(ctx)
	logger.Info("Campaign module registered (%d client RPCs)", len(svc.clientRPCs()))
	return nil
}

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"campaign_chapters":         s.rpcChapters,
		"campaign_attempt_start":    s.rpcStart,
		"campaign_attempt_complete": s.rpcComplete,
		"campaign_freeride_claim":   s.rpcFreerideClaim,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Campaign settings not read, using defaults: %v", err)
	}
	return cfg
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
			if ok, err := clusterkit.AcquireLease(ctx, s.db, "campaign-worker", holder, 50*time.Second); err != nil || !ok {
				continue
			}
			s.retryPayments(ctx, 50)
			if n++; n%30 == 1 {
				s.sweep(ctx)
			}
		}
	}
}

func (s *service) sweep(ctx context.Context) {
	cfg := s.settings(ctx)
	_, _ = s.db.ExecContext(ctx, `UPDATE campaign_attempt SET status='closed', closed_at=now() WHERE status='open' AND expires_at < now()`)
	if cfg.Attempts.KeepDays > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM campaign_attempt WHERE id IN (SELECT id FROM campaign_attempt WHERE status IN ('closed','done') AND started_at < now() - make_interval(days => $1) LIMIT 5000)`, cfg.Attempts.KeepDays)
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM campaign_claim WHERE status='done' AND expires_at < now()`)
	for _, t := range []string{"campaign_user_stage", "campaign_user", "campaign_user_chapter", "campaign_reward_group"} {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM `+t+` WHERE user_id IN (SELECT DISTINCT x.user_id FROM `+t+` x LEFT JOIN users u ON u.id=x.user_id WHERE u.id IS NULL LIMIT 200)`)
	}
}

func retryable(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && (pg.Code == "40001" || pg.Code == "40P01")
}

func (s *service) fail(w http.ResponseWriter, err error) {
	if pluginkit.WriteError(w, err) {
		return
	}
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		pluginkit.WriteJSON(w, e.Status, map[string]any{"error": e.Code, "message": e.Code, "field": e.Field, "meta": e.Meta})
		return
	}
	s.logger.Error("Campaign console request failed: %v", err)
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}
