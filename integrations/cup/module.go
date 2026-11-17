package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"net/http"
	"regexp"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/cuphooks"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/cups"
	maxPayload    = 16 * 1024
	sourceType    = "cup"
	owner         = "cup"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	now    func() time.Time
	rnd    func(int) int

	prize  func(ctx context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error)
	debit  func(ctx context.Context, userID, grantID string, amounts map[string]int64, meta map[string]any) error
	credit func(ctx context.Context, userID, grantID string, amounts map[string]int64, meta map[string]any) error
	notify func(ctx context.Context, n note)
	launch func(ctx context.Context, l cuphooks.Launch) (cuphooks.Launched, error)
	nudge  chan struct{}
}

func cryptoInt(n int) int {
	if n <= 1 {
		return 0
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n))
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger) *service {
	s := &service{db: db, nk: nk, logger: logger, now: time.Now, rnd: cryptoInt, prize: economyhooks.GrantPrize, nudge: make(chan struct{}, 1)}
	s.debit = func(ctx context.Context, userID, grantID string, amounts map[string]int64, meta map[string]any) error {
		return pluginkit.Debit(ctx, s.nk, s.db, userID, grantID, amounts, meta)
	}
	s.credit = func(ctx context.Context, userID, grantID string, amounts map[string]int64, meta map[string]any) error {
		applied, err := pluginkit.WalletApplied(ctx, s.db, userID, grantID)
		if err != nil || applied {
			return err
		}
		md := map[string]any{"grantId": grantID}
		for k, v := range meta {
			md[k] = v
		}
		_, _, err = s.nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: amounts, Metadata: md}}, true)
		return err
	}
	s.notify = s.send
	s.launch = func(ctx context.Context, l cuphooks.Launch) (cuphooks.Launched, error) {
		g, ok := cuphooks.Lookup(l.Game)
		if !ok {
			return cuphooks.Launched{}, cuphooks.ErrNoGame
		}
		return g.Launch(ctx, l)
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
	if err := initializer.RegisterRpc("cup_report", apierr.RPC(db, apierr.RPCOptions{Logger: logger, MaxPayload: maxPayload, AllowAnonymous: true}, svc.rpcReport)); err != nil {
		return err
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	cuphooks.SetReporter(svc.report)
	go svc.loop(ctx)
	logger.Info("Cup module registered (%d client RPCs)", len(svc.clientRPCs()))
	return nil
}

func (s *service) clientRPCs() map[string]apierr.Handler {
	return map[string]apierr.Handler{
		"cup_list":     s.rpcList,
		"cup_get":      s.rpcGet,
		"cup_signup":   s.rpcSignup,
		"cup_withdraw": s.rpcWithdraw,
		"cup_my_match": s.rpcMyMatch,
		"cup_ready":    s.rpcReady,
		"cup_start":    s.rpcStart,
		"cup_history":  s.rpcHistory,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Cup settings not read, using defaults: %v", err)
	}
	return cfg
}

func (s *service) wake() {
	select {
	case s.nudge <- struct{}{}:
	default:
	}
}

func (s *service) loop(ctx context.Context) {
	holder := clusterkit.Holder(ctx)
	tick := time.Duration(s.settings(ctx).Limits.TickSeconds) * time.Second
	timer := time.NewTimer(tick)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-s.nudge:
			time.Sleep(300 * time.Millisecond)
		}
		cfg := s.settings(ctx)
		tick = time.Duration(cfg.Limits.TickSeconds) * time.Second
		timer.Reset(tick)
		if ok, err := clusterkit.AcquireLease(ctx, s.db, "cup-worker", holder, tick+20*time.Second); err != nil || !ok {
			continue
		}
		s.work(ctx)
	}
}

func (s *service) work(ctx context.Context) {
	s.advanceCups(ctx)
	if err := s.scheduleTemplates(ctx); err != nil {
		s.logger.Warn("Cup templates not scheduled: %v", err)
	}
	s.retryRefunds(ctx, 200)
	s.payPending(ctx, "", 500)
}

func (s *service) send(ctx context.Context, n note) {
	cfg := s.settings(ctx)
	if s.nk != nil && cfg.Events.Notify {
		list := make([]*runtime.NotificationSend, 0, len(n.users))
		for _, id := range n.users {
			list = append(list, &runtime.NotificationSend{UserID: id, Subject: n.subject, Content: n.content, Code: cfg.Events.Code, Persistent: n.persistent})
		}
		if err := s.nk.NotificationsSend(ctx, list); err != nil {
			s.logger.Debug("Cup notification %s not sent: %v", n.subject, err)
		}
	}
	if n.inapp && cfg.Events.InApp && inapp.Available() {
		title, body := s.inappText(ctx, n)
		if len(title) == 0 {
			return
		}
		cupID, _ := n.content["cupId"].(string)
		if _, err := inapp.Send(ctx, inapp.Message{UserIDs: n.users, Title: title, Body: body, Action: &inapp.Action{Type: "cup", Value: cupID}, CreatedBy: "cup"}); err != nil {
			s.logger.Debug("Cup in-app message %s not sent: %v", n.subject, err)
		}
	}
}

func (s *service) cupName(ctx context.Context, id string) map[string]string {
	var raw string
	if err := s.db.QueryRowContext(ctx, `SELECT name::text FROM cup WHERE id::text=$1`, id).Scan(&raw); err != nil {
		return map[string]string{}
	}
	return textMap(raw)
}

func (s *service) inappText(ctx context.Context, n note) (map[string]string, map[string]string) {
	cupID, _ := n.content["cupId"].(string)
	name := s.cupName(ctx, cupID)
	fa, en := name["fa"], name["en"]
	if fa == "" {
		fa = en
	}
	if en == "" {
		en = fa
	}
	title := map[string]string{"fa": fa, "en": en}
	switch n.subject {
	case "cup_round_started", "cup_check_in":
		return title, map[string]string{"fa": "نوبت بازی بعدی‌ات رسیده. اعلام حضور کن.", "en": "Check-in for your next match is open. Check in now."}
	case "cup_finished":
		return title, map[string]string{"fa": "تورنمنت تمام شد. نتیجه و جایزه‌ات را ببین.", "en": "The cup is over. See your prize."}
	case "cup_cancelled":
		return title, map[string]string{"fa": "تورنمنت لغو شد و ورودی‌ات برگشت داده شد.", "en": "The cup was cancelled and your entry was refunded."}
	case "cup_prize":
		return title, map[string]string{"fa": "جایزه تورنمنت به حسابت واریز شد.", "en": "Your cup prize has been paid."}
	}
	return nil, nil
}

func (s *service) publish(ctx context.Context, events []cupEvent) {
	if !s.settings(ctx).Events.Publish || !gameevents.Has() {
		return
	}
	for _, e := range events {
		id := e.name + ":" + jsonOf(e.fields)
		if _, errs := gameevents.Publish(ctx, gameevents.Event{ID: truncate(id, 200), UserID: e.userID, Name: e.name, Fields: e.fields, Source: owner}); len(errs) > 0 {
			s.logger.Debug("Cup event %s not handled by every subscriber: %v", e.name, errs)
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
	s.logger.Error("Cup console request failed: %v", err)
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}

type texts struct {
	langs *localized.Set
	lang  string
}

func (s *service) texts(ctx context.Context, userID, lang string) texts {
	langs, l, _ := localized.ForUser(ctx, s.db, userID, lang)
	return texts{langs: langs, lang: l}
}

func (t texts) of(m map[string]string, fallback string) string {
	if t.langs != nil {
		if v := t.langs.Resolve(localized.Text(m), t.lang); v != "" {
			return v
		}
	}
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return fallback
}
