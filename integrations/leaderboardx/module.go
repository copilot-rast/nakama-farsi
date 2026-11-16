package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/leaderboardx"
	maxPayload    = 16 * 1024
	sourceType    = "leaderboard"
	owner         = "leaderboardx"
)

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	now    func() time.Time
	core   coreBoards

	prize func(ctx context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error)

	notify func(ctx context.Context, userID, subject string, content map[string]any, code int)
	nudge  chan struct{}
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger) *service {
	s := &service{db: db, nk: nk, logger: logger, now: time.Now, core: nkBoards{nk: nk}, prize: economyhooks.GrantPrize, nudge: make(chan struct{}, 1)}
	s.notify = func(ctx context.Context, userID, subject string, content map[string]any, code int) {
		if err := nk.NotificationSend(ctx, userID, subject, content, code, "", true); err != nil {
			logger.Debug("Leaderboard notification to %s not sent: %v", userID, err)
		}
	}
	return s
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

	if err := initializer.RegisterLeaderboardReset(func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, lb *api.Leaderboard, reset int64) error {
		svc.wake()
		return nil
	}); err != nil {
		return err
	}
	playerevents.Observe(owner, svc.onEconomyEvent)
	gameevents.Subscribe(owner, svc.onGameEvent)
	go svc.syncCore(ctx)
	go svc.loop(ctx)
	logger.Info("Leaderboardx module registered (%d client RPCs)", len(svc.clientRPCs()))
	return nil
}

func (s *service) clientRPCs() map[string]apierr.Handler {
	return map[string]apierr.Handler{
		"leaderboardx_list":    s.rpcList,
		"leaderboardx_board":   s.rpcBoard,
		"leaderboardx_submit":  s.rpcSubmit,
		"leaderboardx_history": s.rpcHistory,
		"leaderboardx_claim":   s.rpcClaim,
		"leaderboardx_winners": s.rpcWinners,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Leaderboard settings not read, using defaults: %v", err)
	}
	return cfg
}

func (s *service) wake() {
	select {
	case s.nudge <- struct{}{}:
	default:
	}
}

func (s *service) syncCore(ctx context.Context) {
	time.Sleep(2 * time.Second)
	boards, err := s.boards(ctx, "")
	if err != nil {
		return
	}
	for _, b := range boards {
		if err := s.core.Create(ctx, b.Key, b.SortOrder, b.Operator, b.ResetSchedule, map[string]any{"module": owner}); err != nil {
			s.logger.Warn("Leaderboard %s: core leaderboard not created: %v", b.Key, err)
		}
	}
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
		case <-s.nudge:

			time.Sleep(2 * time.Second)
		}
		if ok, err := clusterkit.AcquireLease(ctx, s.db, "leaderboardx-worker", holder, 50*time.Second); err != nil || !ok {
			continue
		}
		s.work(ctx)
		if n++; n%30 == 1 {
			s.sweep(ctx)
		}
	}
}

func (s *service) work(ctx context.Context) {
	if err := s.closeDue(ctx); err != nil {
		s.logger.Warn("Leaderboard period close failed: %v", err)
	}
	s.retryPayments(ctx, 200)
}

func (s *service) sweep(ctx context.Context) {
	cfg := s.settings(ctx)
	_, _ = s.db.ExecContext(ctx, `UPDATE leaderboardx_result SET status='expired' WHERE status='claimable' AND claim_expires_at < now()`)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM leaderboardx_score_log WHERE ctid IN (SELECT ctid FROM leaderboardx_score_log WHERE created_at < now() - make_interval(days => $1) LIMIT 20000)`, cfg.Scores.LogDays)
	if cfg.History.KeepDays > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM leaderboardx_period p WHERE p.period_end < now() - make_interval(days => $1)
			AND NOT EXISTS (SELECT 1 FROM leaderboardx_result r WHERE r.period_id=p.id AND r.status IN ('pending','claimable'))`, cfg.History.KeepDays)
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
	s.logger.Error("Leaderboard console request failed: %v", err)
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}

func (s *service) boards(ctx context.Context, filter string) ([]*board, error) {
	q := `SELECT ` + boardCols + ` FROM leaderboardx_board`
	if filter == "active" {
		q += ` WHERE is_active`
	}
	rows, err := s.db.QueryContext(ctx, q+` ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*board{}
	for rows.Next() {
		b, err := scanBoard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

var errNoBoard = errors.New("board not found")

func (s *service) boardByKey(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, key string) (*board, error) {
	b, err := scanBoard(q.QueryRowContext(ctx, `SELECT `+boardCols+` FROM leaderboardx_board WHERE key=$1`, key))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNoBoard
	}
	return b, err
}

func (s *service) boardByID(ctx context.Context, id string) (*board, error) {
	b, err := scanBoard(s.db.QueryRowContext(ctx, `SELECT `+boardCols+` FROM leaderboardx_board WHERE id::text=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNoBoard
	}
	return b, err
}

type person struct {
	Username    string
	DisplayName string
	AvatarURL   string
}

func (s *service) people(ctx context.Context, ids []string) map[string]person {
	out := map[string]person{}
	if len(ids) == 0 {
		return out
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, COALESCE(username,''), COALESCE(display_name,''), COALESCE(avatar_url,'') FROM users WHERE id::text = ANY($1)`, ids)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var p person
		if rows.Scan(&id, &p.Username, &p.DisplayName, &p.AvatarURL) == nil {
			out[id] = p
		}
	}
	return out
}
