package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/inventoryx"
	maxPayload    = 8 * 1024
	owner         = "inventoryx"
)

type clientRPC = apierr.Handler

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	// level reads a player's level through the economy plugin; replaced in tests.
	level func(ctx context.Context, userID, tableID string) (int, error)
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger) *service {
	return &service{db: db, nk: nk, logger: logger, level: economyhooks.Level}
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
	if err := apikey.RegisterScopes(ctx, db, owner, map[string]string{ScopeItems: "Import economy items (bulk JSON) from a tool"}); err != nil {
		return err
	}
	if err := initializer.RegisterHttp("/v2/inventoryx/editor/items", apikey.Require(db, ScopeItems, svc.httpImport), http.MethodPost); err != nil {
		return err
	}
	economyhooks.SetItemStates(owner, svc.itemStates)
	go svc.loop(ctx)
	logger.Info("Inventoryx module registered (%d client RPCs)", len(svc.clientRPCs()))
	return nil
}

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"inventoryx_items":   s.rpcItems,
		"inventoryx_unlock":  s.rpcUnlock,
		"inventoryx_upgrade": s.rpcUpgrade,
	}
}

func (s *service) settings(ctx context.Context) settings.All {
	cfg, err := settings.Load(ctx, s.db)
	if err != nil && s.logger != nil {
		s.logger.Warn("Inventoryx settings not read, using defaults: %v", err)
	}
	return cfg
}

func (s *service) loop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ok, err := clusterkit.AcquireLease(ctx, s.db, "inventoryx-sweep", holder, 4*time.Minute); err != nil || !ok {
				continue
			}
			s.sweep(ctx)
		}
	}
}

// sweep removes expired request answers and rows of deleted players.
func (s *service) sweep(ctx context.Context) {
	_, _ = s.db.ExecContext(ctx, `DELETE FROM inventoryx_request WHERE expires_at < now()`)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM inventoryx_user_track WHERE user_id IN (SELECT DISTINCT t.user_id FROM inventoryx_user_track t LEFT JOIN users u ON u.id=t.user_id WHERE u.id IS NULL LIMIT 200)`)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM inventoryx_user_unlock WHERE user_id IN (SELECT DISTINCT t.user_id FROM inventoryx_user_unlock t LEFT JOIN users u ON u.id=t.user_id WHERE u.id IS NULL LIMIT 200)`)
}

func (s *service) rpcItems(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Keys []string `json:"keys"`
		Lang string   `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	keys := cleanKeys(in.Keys)
	if len(keys) > 500 {
		return nil, apierr.BadRequest("TOO_MANY_KEYS").WithField("keys")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	where := ` WHERE i.is_active AND EXISTS (SELECT 1 FROM inventoryx_item x WHERE x.item_id=i.id)`
	args := []any{}
	if len(keys) > 0 {
		where = ` WHERE i.is_active AND i.key = ANY($1)`
		args = append(args, keys)
	}
	rows, err := s.db.QueryContext(ctx, itemSelect+where+` ORDER BY i.sort, i.key`, args...)
	if err != nil {
		return nil, err
	}
	items, err := scanItems(rows)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	configs, err := loadConfigs(ctx, s.db, ids)
	if err != nil {
		return nil, err
	}
	st, err := s.loadPlayer(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	level := s.levelReader(ctx, userID)
	list := []map[string]any{}
	for _, it := range items {
		v := itemView(it, configs[it.ID], st, level, langs, lang)
		v["key"] = it.Key
		v["owned"] = st.owned[it.ID] || it.Starter
		v["unlockLevel"] = it.UnlockLevel
		list = append(list, v)
	}
	return map[string]any{"lang": lang, "level": level(""), "items": list}, nil
}

// answer is the reply of unlock and upgrade: the result, the item's new state and the wallet.
func (s *service) answer(ctx context.Context, userID, lang string, r result) (any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, itemSelect+` WHERE i.key=$1`, r.ItemKey)
	if err != nil {
		return nil, err
	}
	items, err := scanItems(rows)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "itemKey": r.ItemKey, "action": r.Action, "paid": r.Paid, "replayed": r.Replayed}
	if r.Action == "unlock" {
		out["unlocked"] = true
		out["owned"] = r.Owned
	} else {
		out["track"], out["level"], out["value"] = r.Track, r.Level, r.Value
	}
	if len(items) == 1 {
		configs, err := loadConfigs(ctx, s.db, []string{items[0].ID})
		if err != nil {
			return nil, err
		}
		st, err := s.loadPlayer(ctx, userID, []string{items[0].ID})
		if err != nil {
			return nil, err
		}
		v := itemView(items[0], configs[items[0].ID], st, s.levelReader(ctx, userID), langs, lang)
		v["key"] = items[0].Key
		v["owned"] = st.owned[items[0].ID] || items[0].Starter
		out["item"] = v
	}
	var raw string
	if err := s.db.QueryRowContext(ctx, `SELECT wallet::text FROM users WHERE id=$1`, userID).Scan(&raw); err == nil {
		wallet := map[string]int64{}
		_ = json.Unmarshal([]byte(raw), &wallet)
		out["wallet"] = wallet
	}
	return out, nil
}

func (s *service) rpcUnlock(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key         string `json:"key"`
		RequestID   string `json:"requestId"`
		CurrencyKey string `json:"currencyKey"`
		Lang        string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	r, err := s.unlock(ctx, userID, in.Key, in.RequestID, in.CurrencyKey)
	if err != nil {
		return nil, err
	}
	return s.answer(ctx, userID, in.Lang, r)
}

func (s *service) rpcUpgrade(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key       string `json:"key"`
		Track     string `json:"track"`
		FromLevel *int   `json:"fromLevel"`
		RequestID string `json:"requestId"`
		Lang      string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	r, err := s.upgrade(ctx, userID, in.Key, in.Track, in.FromLevel, in.RequestID)
	if err != nil {
		return nil, err
	}
	return s.answer(ctx, userID, in.Lang, r)
}

func (s *service) fail(w http.ResponseWriter, err error) {
	if pluginkit.WriteError(w, err) {
		return
	}
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		pluginkit.WriteJSON(w, e.Status, map[string]any{"error": e.Code, "message": e.Code, "field": e.Field, "meta": e.Meta})
		return
	}
	s.logger.Error("Inventoryx console request failed: %v", err)
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}
