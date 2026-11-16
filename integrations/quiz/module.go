package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/contentimport"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	consolePrefix = "/v2/console/base/quiz"
	maxPayload    = 64 * 1024
	sourceType    = "quiz"
)

type clientRPC = apierr.Handler

var webhookBox = secretbox.Box{Salt: "nakama-quiz", AAD: "quiz-result-webhook"}

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	now    func() time.Time

	prize func(ctx context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error)

	campaignStart    func(ctx context.Context, userID, stageKey string) (string, error)
	campaignComplete func(ctx context.Context, userID string, c campaignhooks.Completion) (map[string]any, error)
	campaignOn       func() bool

	secretKey []byte

	send func(ctx context.Context, url string, body []byte, headers map[string]string) (int, error)
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger) *service {
	s := &service{db: db, nk: nk, logger: logger, now: time.Now, prize: economyhooks.GrantPrize,
		campaignStart: campaignhooks.Start, campaignComplete: campaignhooks.Complete, campaignOn: campaignhooks.Available}
	s.send = s.postWebhook
	return s
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	svc := newService(db, nk, logger)
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	if secret, err := secretbox.Secret(env); err == nil {
		if key, err := webhookBox.DeriveKey(secret, "quiz-webhook-secret"); err == nil {
			svc.secretKey = key
		}
	}
	for name, fn := range svc.allRPCs() {
		if err := initializer.RegisterRpc(name, apierr.RPC(db, apierr.RPCOptions{Logger: logger, MaxPayload: maxPayload}, fn)); err != nil {
			return err
		}
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	campaignhooks.SetOwner("quiz", svc.ownsStage)
	contentimport.Register("quiz", svc.importJSON)
	go svc.loop(ctx)
	logger.Info("Quiz module registered (%d client RPCs)", len(svc.allRPCs()))
	return nil
}

func (s *service) allRPCs() map[string]clientRPC {
	out := s.clientRPCs()
	for k, v := range s.duelRPCs() {
		out[k] = v
	}
	return out
}

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"quiz_packs":           s.rpcPacks,
		"quiz_pack_random":     s.rpcRandom,
		"quiz_play":            s.rpcPlay,
		"quiz_session":         s.rpcSession,
		"quiz_answer":          s.rpcAnswer,
		"quiz_answer_all":      s.rpcAnswerAll,
		"quiz_finish":          s.rpcFinish,
		"quiz_hint_buy":        s.rpcHintBuy,
		"quiz_missions":        s.rpcMissions,
		"quiz_mission_start":   s.rpcMissionStart,
		"quiz_stats":           s.rpcStats,
		"quiz_report_reasons":  s.rpcReportReasons,
		"quiz_report_question": s.rpcReportQuestion,
		"quiz_result_report":   s.rpcResultReport,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Quiz settings not read, using defaults: %v", err)
	}
	return cfg
}

func (s *service) zone(cfg settings.All) *time.Location {
	loc, err := time.LoadLocation(cfg.Rewards.DayZone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (s *service) ownsStage(ctx context.Context, stageKey string) (string, bool) {
	if stageKey == "" || !s.settings(ctx).Campaign.Enabled {
		return "", false
	}
	var key string
	if err := s.db.QueryRowContext(ctx, `SELECT key FROM quiz_pack WHERE campaign_stage_key=$1 AND is_active`, stageKey).Scan(&key); err != nil {
		return "", false
	}
	return key, true
}

func (s *service) loop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	n := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ok, err := clusterkit.AcquireLease(ctx, s.db, "quiz-worker", holder, 25*time.Second); err != nil || !ok {
				continue
			}
			s.closeExpired(ctx, 50)
			s.retryPayments(ctx, 50)
			s.deliverWebhooks(ctx, 20)
			s.duelWorker(ctx, 50)
			if n++; n%60 == 1 {
				s.sweep(ctx)
			}
		}
	}
}

func (s *service) sweep(ctx context.Context) {
	cfg := s.settings(ctx)
	s.sweepDuels(ctx, cfg)
	if cfg.Retention.SessionDays > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM quiz_session WHERE id IN (SELECT id FROM quiz_session WHERE status IN ('closed','done') AND started_at < now() - make_interval(days => $1) LIMIT 5000)`, cfg.Retention.SessionDays)
	}
	if cfg.Retention.DeliveryDays > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM quiz_result_delivery WHERE status <> 'pending' AND created_at < now() - make_interval(days => $1)`, cfg.Retention.DeliveryDays)
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM quiz_user_day WHERE day < to_char(now() - interval '7 days', 'YYYY-MM-DD')`)
	for _, t := range []string{"quiz_user_pack", "quiz_user_question", "quiz_user_mission", "quiz_user_hint", "quiz_session"} {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM `+t+` WHERE user_id IN (SELECT DISTINCT x.user_id FROM `+t+` x LEFT JOIN users u ON u.id=x.user_id WHERE u.id IS NULL LIMIT 200)`)
	}
}

func retryable(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && (pg.Code == "40001" || pg.Code == "40P01")
}

func isUnique(err error) bool {
	var pg interface{ SQLState() string }
	return errors.As(err, &pg) && pg.SQLState() == "23505"
}

func (s *service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	var err error
	for try := 1; try <= 3; try++ {
		var tx *sql.Tx
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err = fn(tx); err != nil {
			_ = tx.Rollback()
		} else if err = tx.Commit(); err == nil {
			return nil
		}
		if !retryable(err) {
			return err
		}
	}
	return apierr.Conflict("QUIZ_BUSY").Retry(1)
}

func (s *service) fail(w http.ResponseWriter, err error) {
	if pluginkit.WriteError(w, err) {
		return
	}
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		pluginkit.WriteJSON(w, e.Status, map[string]any{"error": e.Code, "message": e.Code, "field": e.Field, "meta": e.Meta})
		return
	}
	s.logger.Error("Quiz console request failed: %v", err)
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}
