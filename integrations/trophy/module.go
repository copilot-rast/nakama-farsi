package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/trophy/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/trophies"
	maxPayload    = 16 * 1024
	sourceType    = "trophy"
	owner         = "trophy"
)

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	now    func() time.Time

	prize  func(ctx context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error)
	notify func(ctx context.Context, userID, subject string, content map[string]any, code int, persistent bool)
	nudge  chan struct{}

	mu      sync.Mutex
	cached  ladder
	fetched time.Time
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger) *service {
	s := &service{db: db, nk: nk, logger: logger, now: time.Now, prize: economyhooks.GrantPrize, nudge: make(chan struct{}, 1)}
	s.notify = func(ctx context.Context, userID, subject string, content map[string]any, code int, persistent bool) {
		if nk == nil {
			return
		}
		if err := nk.NotificationSend(ctx, userID, subject, content, code, "", persistent); err != nil {
			logger.Debug("Trophy notification to %s not sent: %v", userID, err)
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
	if err := initializer.RegisterBeforeRt("MatchmakerAdd", hookchain.Before("rt:MatchmakerAdd", svc.beforeMatchmakerAdd)); err != nil {
		return err
	}
	playerevents.Observe(owner, svc.onEconomyEvent)
	gameevents.Subscribe(owner, svc.onGameEvent)
	go svc.loop(ctx)
	logger.Info("Trophy module registered (%d client RPCs)", len(svc.clientRPCs()))
	return nil
}

func (s *service) clientRPCs() map[string]apierr.Handler {
	return map[string]apierr.Handler{
		"trophy_me":      s.rpcMe,
		"trophy_leagues": s.rpcLeagues,
		"trophy_history": s.rpcHistory,
		"trophy_players": s.rpcPlayers,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Trophy settings not read, using defaults: %v", err)
	}
	return cfg
}

func (s *service) wake() {
	select {
	case s.nudge <- struct{}{}:
	default:
	}
}

func (s *service) invalidate() {
	s.mu.Lock()
	s.fetched = time.Time{}
	s.mu.Unlock()
}

func (s *service) ladder(ctx context.Context) ladder {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.fetched.IsZero() && time.Since(s.fetched) < 10*time.Second {
		return s.cached
	}
	list, err := s.leagues(ctx)
	if err != nil {
		return s.cached
	}
	s.cached, s.fetched = newLadder(list), time.Now()
	return s.cached
}

const leagueCols = `id::text, key, name::text, description::text, icon_url, color, min_trophies, floor, reset_to, prize_key, is_active, metadata::text, updated_by`

type scanner interface{ Scan(dest ...any) error }

func scanLeague(row scanner) (league, error) {
	var l league
	var name, desc, meta string
	var reset sql.NullInt64
	if err := row.Scan(&l.ID, &l.Key, &name, &desc, &l.IconURL, &l.Color, &l.MinTrophies, &l.Floor, &reset, &l.PrizeKey, &l.Active, &meta, &l.UpdatedBy); err != nil {
		return l, err
	}
	_ = json.Unmarshal([]byte(name), &l.Name)
	_ = json.Unmarshal([]byte(desc), &l.Description)
	_ = json.Unmarshal([]byte(meta), &l.Metadata)
	if l.Name == nil {
		l.Name = map[string]string{}
	}
	if l.Description == nil {
		l.Description = map[string]string{}
	}
	if l.Metadata == nil {
		l.Metadata = map[string]any{}
	}
	if reset.Valid {
		v := int(reset.Int64)
		l.ResetTo = &v
	}
	return l, nil
}

func (s *service) leagues(ctx context.Context) ([]league, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+leagueCols+` FROM trophy_league ORDER BY min_trophies, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []league{}
	for rows.Next() {
		l, err := scanLeague(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
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
			time.Sleep(time.Second)
		}
		if ok, err := clusterkit.AcquireLease(ctx, s.db, "trophy-worker", holder, 50*time.Second); err != nil || !ok {
			continue
		}
		s.work(ctx)
		if n++; n%60 == 1 {
			s.sweep(ctx)
		}
	}
}

func (s *service) work(ctx context.Context) {
	if err := s.seasonTick(ctx, ""); err != nil {
		s.logger.Warn("Trophy season check failed: %v", err)
	}
	s.payPending(ctx, "", 200)
}

func (s *service) sweep(ctx context.Context) {
	cfg := s.settings(ctx)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM trophy_log WHERE id IN (SELECT id FROM trophy_log WHERE created_at < now() - make_interval(days => $1) LIMIT 20000)`, cfg.Display.LogDays)
}

func (s *service) fail(w http.ResponseWriter, err error) {
	if pluginkit.WriteError(w, err) {
		return
	}
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		pluginkit.WriteJSON(w, e.Status, map[string]any{"error": e.Code, "message": e.Code, "field": e.Field, "meta": e.Meta})
		return
	}
	s.logger.Error("Trophy console request failed: %v", err)
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}
