package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
)

type legacyLayer struct {
	Part     string `json:"part"`
	Option   string `json:"option"`
	ImageURL string `json:"imageUrl"`
	Value    string `json:"value"`
	Layer    int    `json:"layer"`
}

type legacyLook struct {
	Parts  map[string]string `json:"parts"`
	Layers []legacyLayer     `json:"layers"`
}

func (c *avCatalog) legacyLook(a *avRow, owned func(*avItem) bool) legacyLook {
	out := legacyLook{Parts: map[string]string{}, Layers: []legacyLayer{}}
	for _, r := range c.resolve(a, owned) {
		for i, it := range r.items {
			if i == 0 {
				out.Parts[r.cat.Key] = it.Key
			}
			out.Layers = append(out.Layers, legacyLayer{Part: r.cat.Key, Option: it.Key, ImageURL: it.ImageURL, Value: it.Value, Layer: r.cat.Layer})
		}
	}
	sort.SliceStable(out.Layers, func(i, j int) bool { return out.Layers[i].Layer < out.Layers[j].Layer })
	return out
}

func legacyFail(code string, grpc codes.Code, meta map[string]any) error {
	raw, _ := json.Marshal(map[string]any{"code": code, "meta": meta})
	return runtime.NewError(string(raw), int(grpc))
}

func (s *service) rpcLegacyAvatar(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarContext(ctx, userID, in.Lang, true)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": x.lang, "enabled": x.st.Enabled, "parts": []any{}, "look": legacyLook{Parts: map[string]string{}, Layers: []legacyLayer{}}}
	if !x.st.Enabled {
		return out, nil
	}
	products, shop, err := s.avatarItemExtras(ctx, x)
	if err != nil {
		return nil, err
	}
	a, err := s.pickAvatar(ctx, x, "")
	if err != nil {
		if app, ok := asApp(err); !ok || app.Code != "AVATAR_NOT_FOUND" {
			return nil, err
		}
		a = nil
	}
	parts := make([]map[string]any, 0, len(x.cat.cats))
	for _, c := range x.cat.cats {
		opts := make([]map[string]any, 0, len(c.Items))
		for _, it := range c.Items {
			if a != nil && !it.fits(a.TypeID) {
				continue
			}
			v := map[string]any{"key": it.Key, "name": x.langs.Resolve(it.Name, x.lang), "imageUrl": it.ImageURL, "iconUrl": it.IconURL, "value": it.Value, "access": it.Access, "isDefault": it.IsDefault, "owned": x.usable(it)}
			switch it.Access {
			case "item":
				v["itemKey"] = it.ItemKey
				if p, ok := shop[it.ItemKey]; ok {
					v["item"] = p
				}
			case "currency":
				v["price"] = it.Price
			case "product":
				if p, ok := products[it.ProductID]; ok {
					v["product"] = p
				}
			}
			opts = append(opts, v)
		}
		parts = append(parts, map[string]any{"key": c.Key, "name": x.langs.Resolve(c.Name, x.lang), "layer": c.Layer, "required": c.Required, "options": opts})
	}
	out["parts"] = parts
	out["look"] = x.cat.legacyLook(a, x.owned)
	return out, nil
}

func (s *service) rpcLegacyAvatarSave(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Look map[string]string `json:"look"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarContext(ctx, userID, "", true)
	if err != nil {
		return nil, err
	}
	if !x.st.Enabled {
		return nil, legacyFail("AVATAR_DISABLED", codes.FailedPrecondition, nil)
	}
	a, err := s.pickAvatar(ctx, x, "")
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(in.Look))
	for k := range in.Look {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	chosen := []*avItem{}
	cleared := []string{}
	for _, partKey := range keys {
		optKey := strings.TrimSpace(in.Look[partKey])
		cat := x.cat.catKey[partKey]
		if cat == nil {
			return nil, legacyFail("AVATAR_PART_NOT_FOUND", codes.NotFound, map[string]any{"part": partKey})
		}
		if optKey == "" {
			if cat.Required {
				return nil, legacyFail("AVATAR_PART_REQUIRED", codes.InvalidArgument, map[string]any{"part": partKey})
			}
			cleared = append(cleared, cat.ID)
			continue
		}
		it := x.cat.find(partKey, optKey)
		if it == nil {
			return nil, legacyFail("AVATAR_OPTION_NOT_FOUND", codes.NotFound, map[string]any{"part": partKey, "option": optKey})
		}
		if !it.fits(a.TypeID) {
			return nil, legacyFail("AVATAR_TYPE_NOT_ALLOWED", codes.PermissionDenied, map[string]any{"part": partKey, "option": optKey})
		}
		if !x.usable(it) {
			return nil, legacyFail("AVATAR_OPTION_LOCKED", codes.PermissionDenied, map[string]any{"part": partKey, "option": optKey, "itemKey": it.ItemKey})
		}
		chosen = append(chosen, it)
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if err := avatarLock(ctx, tx, userID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM economy_avatar_equip WHERE avatar_id=$1`, a.ID); err != nil {
			return err
		}
		for _, it := range chosen {
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_avatar_equip (avatar_id, item_id, category_id) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, a.ID, it.ID, it.Cat.ID); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `UPDATE economy_avatar SET cleared=$2::text[], updated_at=now() WHERE id=$1`, a.ID, cleared)
		return err
	})
	if err != nil {
		return nil, err
	}
	fresh, err := avatarForUser(ctx, s.db, userID, a.ID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"look": x.cat.legacyLook(fresh, x.owned)}, nil
}

func (s *service) rpcLegacyAvatarGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserIDs []string `json:"userIds"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	ids, err := cleanUserIDs(in.UserIDs)
	if err != nil {
		return nil, err
	}
	st, err := loadAvatarSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	looks := map[string]legacyLook{}
	if !st.Enabled {
		for _, id := range ids {
			looks[id] = legacyLook{Parts: map[string]string{}, Layers: []legacyLayer{}}
		}
		return map[string]any{"enabled": false, "looks": looks}, nil
	}
	c, err := loadAvatarCatalog(ctx, s.db)
	if err != nil {
		return nil, err
	}
	rows, own, err := s.activeLooks(ctx, c, ids)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		uid := id
		looks[id] = c.legacyLook(rows[id], func(it *avItem) bool { return own.owns(uid, it) })
	}
	return map[string]any{"enabled": true, "looks": looks}, nil
}

func (s *service) legacyAvatarRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"base_avatar":      s.rpcLegacyAvatar,
		"base_avatar_save": s.rpcLegacyAvatarSave,
		"base_avatar_get":  s.rpcLegacyAvatarGet,
	}
}
