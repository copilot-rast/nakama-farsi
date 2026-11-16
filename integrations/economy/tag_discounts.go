package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/tagdiscount"
)

func tagDiscountResources() []resource {
	return []resource{{
		name:    "tag-discounts",
		table:   "economy_tag_discount",
		json:    tagDiscountJSON,
		search:  []string{"t.key", translatedSearch("t.name")},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean", "tagId": "? = ANY(t.tag_ids)"},
		save:    saveTagDiscount,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			return nil
		},
	}}
}

const tagDiscountJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'tagIds',to_jsonb(t.tag_ids),'percent',t.percent,'allItems',t.all_items,'itemKeys',to_jsonb(t.item_keys),'itemIds',(SELECT COALESCE(jsonb_agg(i.id ORDER BY i.key),'[]'::jsonb) FROM economy_item i WHERE i.key = ANY(t.item_keys)),
	'itemTypeIds',to_jsonb(t.item_type_ids),'allProducts',t.all_products,'productIds',to_jsonb(t.product_ids),'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,'sort',t.sort,
	'createdAt',t.created_at,'updatedAt',t.updated_at,
	'tags',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',g.id,'key',g.key,'name',g.name) ORDER BY g.key),'[]'::jsonb) FROM base_tag g WHERE g.id::text = ANY(t.tag_ids)),
	'players',(SELECT count(DISTINCT ut.user_id) FROM base_user_tag ut WHERE ut.tag_id::text = ANY(t.tag_ids) AND (ut.expires_at IS NULL OR ut.expires_at > now())))`

func saveTagDiscount(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string   `json:"key"`
		Name        i18n     `json:"name"`
		TagIDs      []string `json:"tagIds"`
		Percent     int      `json:"percent"`
		AllItems    bool     `json:"allItems"`
		ItemKeys    []string `json:"itemKeys"`
		ItemIDs     []string `json:"itemIds"`
		ItemTypeIDs []string `json:"itemTypeIds"`
		AllProducts bool     `json:"allProducts"`
		ProductIDs  []string `json:"productIds"`
		StartsAt    string   `json:"startsAt"`
		EndsAt      string   `json:"endsAt"`
		IsActive    *bool    `json:"isActive"`
		Sort        int      `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if err := keepKey(ctx, tx, "economy_tag_discount", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	if in.Percent < 1 || in.Percent > 100 {
		return badRequest("INVALID_PERCENT", "percent", "a whole percent from 1 to 100")
	}
	tags, err := checkUserTags(ctx, tx, in.TagIDs)
	if err != nil {
		var app *appError
		if errors.As(err, &app) {
			app.Field = "tagIds"
		}
		return err
	}
	if len(tags) == 0 {
		return badRequest("TAGS_REQUIRED", "tagIds", "pick at least one player tag")
	}
	items := []string{}
	seen := map[string]bool{}
	for _, raw := range in.ItemIDs {
		ref, err := refExists(ctx, tx, "economy_item", "itemIds", "ITEM_NOT_FOUND", raw, true)
		if err != nil {
			return err
		}
		var k string
		if err := tx.QueryRowContext(ctx, `SELECT key FROM economy_item WHERE id=$1`, *ref).Scan(&k); err != nil {
			return err
		}
		in.ItemKeys = append(in.ItemKeys, k)
	}
	for _, k := range in.ItemKeys {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item WHERE key=$1)`, k).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("ITEM_NOT_FOUND", "itemKeys", "item "+k+" does not exist")
		}
		items = append(items, k)
	}
	types := []string{}
	for _, raw := range in.ItemTypeIDs {
		ref, err := refExists(ctx, tx, "base_category", "itemTypeIds", "ITEM_TYPE_NOT_FOUND", raw, true)
		if err != nil {
			return err
		}
		types = append(types, *ref)
	}
	products := []string{}
	for _, raw := range in.ProductIDs {
		ref, err := refExists(ctx, tx, "economy_product", "productIds", "PRODUCT_NOT_FOUND", raw, true)
		if err != nil {
			return err
		}
		products = append(products, *ref)
	}
	if len(items) > 200 || len(types) > 50 || len(products) > 200 {
		return badRequest("TOO_MANY_TARGETS", "itemKeys", "at most 200 items, 50 item types and 200 products")
	}
	if !in.AllItems && !in.AllProducts && len(items)+len(types)+len(products) == 0 {
		return badRequest("TARGETS_REQUIRED", "itemKeys", "pick the items, item types or products this discount applies to")
	}
	starts, ends, err := cleanWindow(in.StartsAt, in.EndsAt)
	if err != nil {
		return err
	}
	active := in.IsActive == nil || *in.IsActive
	return execSave(ctx, tx, "economy_tag_discount", id, create,
		[]string{"key", "name", "tag_ids", "percent", "all_items", "item_keys", "item_type_ids", "all_products", "product_ids", "starts_at", "ends_at", "is_active", "sort"},
		[]any{key, i18nJSON(name), tags, in.Percent, in.AllItems, items, types, in.AllProducts, products, starts, ends, active, in.Sort})
}

func (s *service) tagDiscountConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/tag-discount-settings", http.MethodGet, s.consoleTagDiscountSettingsGet},
		{consolePrefix + "/tag-discount-settings", http.MethodPut, s.consoleTagDiscountSettingsPut},
		{consolePrefix + "/tag-discount-preview", http.MethodGet, s.consoleTagDiscountPreview},
	}
}

func (s *service) consoleTagDiscountSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := tagdiscount.LoadSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"settings": st})
}

func (s *service) consoleTagDiscountSettingsPut(w http.ResponseWriter, r *http.Request) {
	st, err := tagdiscount.LoadSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := readBody(r, &st); err != nil {
		s.writeError(w, err)
		return
	}
	if err := oneOf("combine", "INVALID_SETTING", st.Combine, tagdiscount.CombineBest, tagdiscount.CombineSum); err != nil {
		s.writeError(w, err)
		return
	}
	if err := oneOf("withOwnDiscount", "INVALID_SETTING", st.WithOwnDiscount, tagdiscount.OwnStack, tagdiscount.OwnBest); err != nil {
		s.writeError(w, err)
		return
	}
	if st.MaxPercent < 1 || st.MaxPercent > 100 {
		s.writeError(w, badRequest("INVALID_SETTING", "maxPercent", "between 1 and 100"))
		return
	}
	st.UpdatedAt = ""
	raw, _ := json.Marshal(st)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_tag_discount_settings (id, settings, updated_at) VALUES (1,$1,now()) ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_at=now()`, string(raw)); err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleTagDiscountSettingsGet(w, r)
}

func (s *service) consoleTagDiscountPreview(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("userId"))
	if _, err := cleanUUID("userId", userID, true); err != nil {
		s.writeError(w, err)
		return
	}
	set, err := tagdiscount.ForUser(r.Context(), s.db, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	items := []map[string]any{}
	defs, err := loadItems(r.Context(), s.db, `WHERE t.is_active AND t.price IS NOT NULL`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	for _, d := range defs {
		m := set.Item(d.Key, d.TypeID)
		if m.Percent == 0 || d.Price == nil {
			continue
		}
		items = append(items, map[string]any{"key": d.Key, "name": d.Name, "price": d.Price, "finalPrice": coinPrice{CurrencyKey: d.Price.CurrencyKey, Amount: tagdiscount.Apply(d.Price.Amount, m.Percent)}, "percent": m.Percent, "rules": m.Rules})
	}
	products := []map[string]any{}
	rows, err := s.db.QueryContext(r.Context(), `SELECT id::text, key, name::text, price_toman, coin_price::text, discount_percent FROM economy_product WHERE is_active ORDER BY sort, created_at`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, key, name string
		var toman sql.NullInt64
		var coin sql.NullString
		var own int
		if err := rows.Scan(&id, &key, &name, &toman, &coin, &own); err != nil {
			s.writeError(w, err)
			return
		}
		m := set.Product(id)
		if m.Percent == 0 {
			continue
		}
		row := map[string]any{"id": id, "key": key, "name": json.RawMessage(name), "percent": m.Percent, "rules": m.Rules, "totalPercent": set.WithOwn(own, m.Percent)}
		if toman.Valid {
			row["priceToman"] = toman.Int64
			pv := productView{ID: id, PriceToman: &toman.Int64, DiscountPercent: own}.withTagDiscount(set)
			row["finalToman"] = *pv.TagFinalToman
		}
		if coin.Valid {
			var cp coinPrice
			if json.Unmarshal([]byte(coin.String), &cp) == nil && cp.CurrencyKey != "" {
				row["coinPrice"] = cp
				row["finalCoin"] = coinPrice{CurrencyKey: cp.CurrencyKey, Amount: tagdiscount.Apply(cp.Amount, m.Percent)}
			}
		}
		products = append(products, row)
	}
	rules := []map[string]any{}
	for _, rl := range set.Rules {
		rules = append(rules, map[string]any{"id": rl.ID, "key": rl.Key, "name": rl.Name, "percent": rl.Percent})
	}
	writeJSON(w, http.StatusOK, map[string]any{"settings": set.Settings, "rules": rules, "items": items, "products": products})
}

func (s *service) userDiscounts(ctx context.Context, q tagdiscount.Queryer, userID string) *tagdiscount.Set {
	set, err := tagdiscount.ForUser(ctx, q, userID)
	if err != nil {
		s.logger.Warn("Economy tag discounts for %s not loaded: %v", userID, err)
		return &tagdiscount.Set{Settings: tagdiscount.Defaults()}
	}
	return set
}

func (p productView) withTagDiscount(set *tagdiscount.Set) productView {
	m := set.Product(p.ID)
	if m.Percent <= 0 {
		return p
	}
	p.TagDiscountPercent = m.Percent
	if p.PriceToman != nil {
		own := discounted(*p.PriceToman, p.DiscountPercent)
		final := own
		if set.Settings.WithOwnDiscount == tagdiscount.OwnBest {
			if m.Percent > p.DiscountPercent {
				final = discounted(*p.PriceToman, m.Percent)
			}
		} else {
			final = tagdiscount.Apply(own, m.Percent)
		}
		p.TagFinalToman = &final
	}
	p.DiscountPercent = set.WithOwn(p.DiscountPercent, m.Percent)
	if p.CoinPrice != nil {
		cp := *p.CoinPrice
		p.CoinOriginal = &coinPrice{CurrencyKey: cp.CurrencyKey, Amount: cp.Amount}
		cp.Amount = tagdiscount.Apply(cp.Amount, m.Percent)
		p.CoinPrice = &cp
	}
	return p
}
