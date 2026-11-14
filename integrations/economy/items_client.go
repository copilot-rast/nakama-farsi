package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) itemRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_inventory":      s.rpcInventory,
		"economy_item_buy":       s.rpcItemBuy,
		"economy_item_unlock":    s.rpcItemUnlock,
		"economy_item_upgrade":   s.rpcItemUpgrade,
		"economy_item_equip":     s.rpcItemEquip(true),
		"economy_item_unequip":   s.rpcItemEquip(false),
		"economy_item_use":       s.rpcItemUse,
		"economy_purchases_mine": s.rpcPurchasesMine,
	}
}

func (s *service) rpcInventory(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Type string `json:"type"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	return s.inventoryView(ctx, userID, langs, lang, strings.TrimSpace(in.Type), false)
}

func (s *service) itemAnswer(ctx context.Context, userID, lang string, res itemResult) (any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	view, err := s.oneItemView(ctx, userID, res.ItemKey, langs, lang)
	if err != nil {
		return nil, err
	}
	wallet, err := s.userWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "item": view, "paid": res.Paid, "wallet": wallet, "replayed": res.Replayed}
	if res.Level > 0 {
		out["level"] = res.Level
	}
	if res.Service != "" {
		out["service"] = res.Service
		out["units"] = res.Units
	}
	return out, nil
}

func (s *service) rpcItemBuy(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key       string `json:"key"`
		RequestID string `json:"requestId"`
		Lang      string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	res, err := s.buyItem(ctx, userID, in.Key, in.RequestID)
	if err != nil {
		return nil, err
	}
	return s.itemAnswer(ctx, userID, in.Lang, res)
}

func (s *service) rpcItemUnlock(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key  string `json:"key"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	res, err := s.unlockItem(ctx, userID, in.Key)
	if err != nil {
		return nil, err
	}
	return s.itemAnswer(ctx, userID, in.Lang, res)
}

func (s *service) rpcItemUpgrade(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key       string `json:"key"`
		FromLevel *int   `json:"fromLevel"`
		RequestID string `json:"requestId"`
		Lang      string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	res, err := s.upgradeItem(ctx, userID, in.Key, in.FromLevel, in.RequestID)
	if err != nil {
		return nil, err
	}
	return s.itemAnswer(ctx, userID, in.Lang, res)
}

func (s *service) rpcItemEquip(on bool) clientRPC {
	return func(ctx context.Context, userID, payload string) (any, error) {
		var in struct {
			Key  string `json:"key"`
			Lang string `json:"lang"`
		}
		if err := decodeStrict(payload, &in); err != nil {
			return nil, err
		}
		slot, off, err := s.equipItem(ctx, userID, in.Key, on)
		if err != nil {
			return nil, err
		}
		langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
		if err != nil {
			return nil, err
		}
		view, err := s.oneItemView(ctx, userID, in.Key, langs, lang)
		if err != nil {
			return nil, err
		}
		return map[string]any{"lang": lang, "slot": slot, "item": view, "unequipped": off}, nil
	}
}

func (s *service) rpcItemUse(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key       string `json:"key"`
		Service   string `json:"service"`
		Units     int    `json:"units"`
		RequestID string `json:"requestId"`
		Lang      string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	res, err := s.useItem(ctx, userID, in.Key, in.Service, in.Units, in.RequestID)
	if err != nil {
		return nil, err
	}
	return s.itemAnswer(ctx, userID, in.Lang, res)
}

func (s *service) purchaseCounts(ctx context.Context, userID string) (map[string]purchaseCount, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.product_id::text, pr.key, count(*), max(COALESCE(p.verified_at, p.updated_at)) FROM economy_payment p JOIN economy_product pr ON pr.id=p.product_id
		WHERE p.user_id=$1 AND p.status='success' AND COALESCE(p.outcome,'') <> 'duplicate' AND p.friend_gift_id IS NULL GROUP BY p.product_id, pr.key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]purchaseCount{}
	for rows.Next() {
		var c purchaseCount
		if err := rows.Scan(&c.ProductID, &c.Key, &c.Count, &c.LastPurchasedAt); err != nil {
			return nil, err
		}
		out[c.ProductID] = c
	}
	return out, rows.Err()
}

func (s *service) rpcPurchasesMine(ctx context.Context, userID, payload string) (any, error) {
	var in struct{}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	counts, err := s.purchaseCounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := []purchaseCount{}
	for _, c := range counts {
		items = append(items, c)
	}
	sortPurchases(items)
	return map[string]any{"items": items}, nil
}

func (s *service) itemConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/users/{userId}/items", http.MethodGet, s.consoleUserItems},
		{consolePrefix + "/users/{userId}/items", http.MethodPost, s.consoleUserItemGrant},
		{consolePrefix + "/users/{userId}/items/{itemId}", http.MethodDelete, s.consoleUserItemRevoke},
	}
}

func (s *service) consoleUserItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := s.ensureStarters(ctx, userID); err != nil {
		s.writeError(w, err)
		return
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, r.URL.Query().Get("lang"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	view, err := s.inventoryView(ctx, userID, langs, lang, strings.TrimSpace(r.URL.Query().Get("type")), true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *service) consoleUserItemGrant(w http.ResponseWriter, r *http.Request) {
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var in struct {
		ItemID string `json:"itemId"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	added, err := s.adminGrantItem(r.Context(), userID, strings.TrimSpace(in.ItemID), consoleName(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !added {
		s.writeError(w, conflict("ALREADY_OWNED", "the player already owns this item"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"granted": true})
}

func (s *service) consoleUserItemRevoke(w http.ResponseWriter, r *http.Request) {
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	removed, err := s.adminRevokeItem(r.Context(), userID, pathVar(r, "itemId"), consoleName(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !removed {
		s.writeError(w, conflict("NOT_OWNED", "the player does not own this item"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}
