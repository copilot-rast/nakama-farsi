package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/clanhooks"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/clans"
	maxPayload    = 16 * 1024
	sourceType    = "clan"
	owner         = "clan"
)

type groupOps interface {
	Create(ctx context.Context, ownerID, name, lang, description, avatar string, open bool, maxCount int) (string, error)
	Update(ctx context.Context, id, name, lang, description, avatar string, open bool, maxCount int) error
	Delete(ctx context.Context, id string) error
	Join(ctx context.Context, id, userID, username string) error
	Leave(ctx context.Context, id, userID, username string) error
	Add(ctx context.Context, id string, userIDs []string) error
	Kick(ctx context.Context, id string, userIDs []string) error
	Promote(ctx context.Context, id string, userIDs []string) error
	Demote(ctx context.Context, id string, userIDs []string) error
	Say(ctx context.Context, id string, content map[string]any) error
}

type nkGroups struct{ nk runtime.NakamaModule }

func (g nkGroups) Create(ctx context.Context, ownerID, name, lang, description, avatar string, open bool, maxCount int) (string, error) {
	grp, err := g.nk.GroupCreate(ctx, ownerID, name, ownerID, lang, description, avatar, open, map[string]any{"clan": true}, maxCount)
	if err != nil {
		return "", err
	}
	return grp.Id, nil
}

func (g nkGroups) Update(ctx context.Context, id, name, lang, description, avatar string, open bool, maxCount int) error {
	return g.nk.GroupUpdate(ctx, id, "", name, "", lang, description, avatar, open, nil, maxCount)
}

func (g nkGroups) Delete(ctx context.Context, id string) error { return g.nk.GroupDelete(ctx, id) }

func (g nkGroups) Join(ctx context.Context, id, userID, username string) error {
	return g.nk.GroupUserJoin(ctx, id, userID, username)
}

func (g nkGroups) Leave(ctx context.Context, id, userID, username string) error {
	return g.nk.GroupUserLeave(ctx, id, userID, username)
}

func (g nkGroups) Add(ctx context.Context, id string, userIDs []string) error {
	return g.nk.GroupUsersAdd(ctx, "", id, userIDs)
}

func (g nkGroups) Kick(ctx context.Context, id string, userIDs []string) error {
	return g.nk.GroupUsersKick(ctx, "", id, userIDs)
}

func (g nkGroups) Promote(ctx context.Context, id string, userIDs []string) error {
	return g.nk.GroupUsersPromote(ctx, "", id, userIDs)
}

func (g nkGroups) Demote(ctx context.Context, id string, userIDs []string) error {
	return g.nk.GroupUsersDemote(ctx, "", id, userIDs)
}

func (g nkGroups) Say(ctx context.Context, id string, content map[string]any) error {
	ch, err := g.nk.ChannelIdBuild(ctx, "", id, runtime.Group)
	if err != nil {
		return err
	}
	_, err = g.nk.ChannelMessageSend(ctx, ch, content, "00000000-0000-0000-0000-000000000000", "", true)
	return err
}

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	now    func() time.Time
	groups groupOps

	prize  func(ctx context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error)
	notify func(ctx context.Context, userIDs []string, subject string, content map[string]any, persistent bool)
	debit  func(ctx context.Context, userID, grantID string, amounts map[string]int64) error
	nudge  chan struct{}

	mu      sync.Mutex
	cached  road
	fetched time.Time
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger) *service {
	s := &service{db: db, nk: nk, logger: logger, now: time.Now, groups: nkGroups{nk: nk}, prize: economyhooks.GrantPrize, nudge: make(chan struct{}, 1)}
	s.notify = func(ctx context.Context, userIDs []string, subject string, content map[string]any, persistent bool) {
		cfg := s.settings(ctx)
		if nk == nil || !cfg.Events.Notify || len(userIDs) == 0 {
			return
		}
		list := make([]*runtime.NotificationSend, 0, len(userIDs))
		for _, id := range userIDs {
			list = append(list, &runtime.NotificationSend{UserID: id, Subject: subject, Content: content, Code: cfg.Events.Code, Persistent: persistent})
		}
		if err := nk.NotificationsSend(ctx, list); err != nil {
			logger.Debug("Clan notification %s not sent: %v", subject, err)
		}
	}
	s.debit = func(ctx context.Context, userID, grantID string, amounts map[string]int64) error {
		return pluginkit.Debit(ctx, nk, db, userID, grantID, amounts, map[string]any{"sourceType": sourceType, "reason": "clan create"})
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
	if err := svc.registerGroupGuards(initializer); err != nil {
		return err
	}
	clanhooks.Set(svc.membership, svc.memberIDs, svc.rivalOf)
	playerevents.Observe(owner, svc.onEconomyEvent)
	gameevents.Subscribe(owner, svc.onGameEvent)
	go svc.loop(ctx)
	logger.Info("Clan module registered (%d client RPCs)", len(svc.clientRPCs()))
	return nil
}

func (s *service) clientRPCs() map[string]apierr.Handler {
	return map[string]apierr.Handler{
		"clan_create":         s.rpcCreate,
		"clan_get":            s.rpcGet,
		"clan_mine":           s.rpcMine,
		"clan_search":         s.rpcSearch,
		"clan_update":         s.rpcUpdate,
		"clan_join":           s.rpcJoin,
		"clan_leave":          s.rpcLeave,
		"clan_requests":       s.rpcRequests,
		"clan_request_accept": s.rpcAccept,
		"clan_request_reject": s.rpcReject,
		"clan_invite":         s.rpcInvite,
		"clan_invites":        s.rpcInvites,
		"clan_invite_accept":  s.rpcInviteAccept,
		"clan_invite_decline": s.rpcInviteDecline,
		"clan_kick":           s.rpcKick,
		"clan_promote":        s.rpcPromote,
		"clan_demote":         s.rpcDemote,
		"clan_transfer":       s.rpcTransfer,
		"clan_levels":         s.rpcLevels,
		"clan_xp_history":     s.rpcXPHistory,
		"clan_war_current":    s.rpcWarCurrent,
		"clan_war_history":    s.rpcWarHistory,
		"clan_war_signup":     s.rpcWarSignup,
		"clan_war_opponents":  s.rpcWarOpponents,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Clan settings not read, using defaults: %v", err)
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

func (s *service) road(ctx context.Context) road {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.fetched.IsZero() && time.Since(s.fetched) < 10*time.Second {
		return s.cached
	}
	list, err := s.levels(ctx)
	if err != nil {
		return s.cached
	}
	s.cached, s.fetched = newRoad(list), time.Now()
	return s.cached
}

func (s *service) levels(ctx context.Context) ([]level, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT level, xp, name::text, icon_url, member_cap, prize_key, perks::text, updated_by FROM clan_level ORDER BY level`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []level{}
	for rows.Next() {
		var l level
		var name, perks string
		if err := rows.Scan(&l.Level, &l.XP, &name, &l.IconURL, &l.MemberCap, &l.PrizeKey, &perks, &l.UpdatedBy); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(name), &l.Name)
		_ = json.Unmarshal([]byte(perks), &l.Perks)
		if l.Name == nil {
			l.Name = map[string]string{}
		}
		if l.Perks == nil {
			l.Perks = map[string]any{}
		}
		out = append(out, l)
	}
	return out, rows.Err()
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
		case <-s.nudge:
			time.Sleep(500 * time.Millisecond)
		}
		if ok, err := clusterkit.AcquireLease(ctx, s.db, "clan-worker", holder, 25*time.Second); err != nil || !ok {
			continue
		}
		s.work(ctx)
		if n++; n%120 == 1 {
			s.sweep(ctx)
		}
	}
}

func (s *service) work(ctx context.Context) {
	if err := s.scheduleRounds(ctx); err != nil {
		s.logger.Warn("Clan war scheduling failed: %v", err)
	}
	s.advanceRounds(ctx)
	s.payPending(ctx, "", 500)
}

func (s *service) sweep(ctx context.Context) {
	_, _ = s.db.ExecContext(ctx, `DELETE FROM clan_request WHERE kind='invite' AND expires_at IS NOT NULL AND expires_at < now()`)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM clan_cooldown WHERE created_at < now() - INTERVAL '30 days'`)
	s.reconcileCaps(ctx)
}

func (s *service) fail(w http.ResponseWriter, err error) {
	if pluginkit.WriteError(w, err) {
		return
	}
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		pluginkit.WriteJSON(w, e.Status, map[string]any{"error": e.Code, "message": e.Code, "field": e.Field, "meta": e.Meta})
		return
	}
	s.logger.Error("Clan console request failed: %v", err)
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}
