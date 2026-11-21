package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/itemkinds"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	consolePrefix = "/v2/console/base/itemkinds"
	owner         = "itemkinds"
)

type service struct {
	db     *sql.DB
	logger runtime.Logger
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	svc := &service{db: db, logger: logger}
	economyhooks.SetItemStates(owner, svc.itemStates)
	economyhooks.SetItemBuyCheck(owner, svc.checkBuy)
	economyhooks.SetSlotResolver(owner, svc.slot)
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	logger.Info("Item kinds module registered (%d kinds so far)", len(itemkinds.List()))
	return nil
}

func (s *service) settings(ctx context.Context) itemkinds.Settings {
	return itemkinds.Load(ctx, s.db)
}

func (s *service) itemStates(ctx context.Context, userID, lang string, itemIDs []string) (map[string]economyhooks.ItemState, error) {
	cfg := s.settings(ctx)
	want := cfg.Inventory.Profile == "always" || (cfg.Inventory.Profile == "ask" && economyhooks.HasOption(ctx, "profile"))
	if !want || len(itemIDs) == 0 {
		return nil, nil
	}
	profiles, err := itemkinds.Profiles(ctx, s.db, itemIDs)
	if err != nil {
		return nil, err
	}
	byKind := map[string][]itemkinds.Profile{}
	out := map[string]economyhooks.ItemState{}
	for id, p := range profiles {
		if p.Status == "deleted" {
			continue
		}
		byKind[p.Kind] = append(byKind[p.Kind], p)
		out[id] = economyhooks.ItemState{Fields: map[string]any{"kind": p.Kind, "profile": p.Data}}
	}
	for key, list := range byKind {
		k := itemkinds.Get(key)
		if k == nil || k.View == nil {
			continue
		}
		extra, err := k.View(ctx, itemkinds.ViewInput{Q: s.db, UserID: userID, Profiles: list, Config: cfg.KindConfig(k)})
		if err != nil {
			s.logger.Warn("Item kind %s view failed: %v", key, err)
			continue
		}
		for id, fields := range extra {
			st := out[id]
			if st.Fields == nil {
				st.Fields = map[string]any{}
			}
			for fk, fv := range fields {
				st.Fields[fk] = fv
			}
			out[id] = st
		}
	}
	return out, nil
}

func (s *service) checkBuy(ctx context.Context, tx *sql.Tx, b economyhooks.ItemBuy) error {
	ps, err := itemkinds.Profiles(ctx, tx, []string{b.ItemID})
	if err != nil {
		return err
	}
	p, ok := ps[b.ItemID]
	if !ok {
		return nil
	}
	k := itemkinds.Get(p.Kind)
	if k == nil || k.CheckBuy == nil {
		return nil
	}
	cfg := s.settings(ctx)
	if !cfg.KindEnabled(k) {
		return nil
	}
	err = k.CheckBuy(ctx, itemkinds.BuyInput{Tx: tx, UserID: b.UserID, ItemKey: b.ItemKey, Profile: p, Config: cfg.KindConfig(k)})
	var r *itemkinds.Refusal
	if errors.As(err, &r) {
		return &economyhooks.Refusal{Status: http.StatusConflict, Code: r.Code, Message: r.Message, Meta: r.Meta}
	}
	return err
}

func (s *service) slot(ctx context.Context, itemID, typeKey string) string {
	ps, err := itemkinds.Profiles(ctx, s.db, []string{itemID})
	if err != nil {
		return ""
	}
	p, ok := ps[itemID]
	if !ok {
		return ""
	}
	k := itemkinds.Get(p.Kind)
	if k == nil || k.Slot == nil || !s.settings(ctx).KindEnabled(k) {
		return ""
	}
	return k.Slot(p)
}

func (s *service) fail(w http.ResponseWriter, err error) {
	if pluginkit.WriteError(w, err) {
		return
	}
	var rej *itemkinds.RejectError
	if errors.As(err, &rej) {
		pluginkit.WriteJSON(w, http.StatusConflict, map[string]any{"error": rej.Code, "message": rej.Detail, "rows": rej.Rows})
		return
	}
	var fe *itemkinds.FieldError
	if errors.As(err, &fe) {
		pluginkit.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "INVALID_PROFILE", "message": fe.Message, "field": fe.Path})
		return
	}
	if r, ok := economyhooks.AsRefusal(err); ok {
		field, _ := r.Meta["field"].(string)
		pluginkit.WriteJSON(w, r.Status, map[string]any{"error": r.Code, "message": r.Message, "field": field})
		return
	}
	if s.logger != nil {
		s.logger.Error("Item kinds console request failed: %v", err)
	}
	pluginkit.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}
