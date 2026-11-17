package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const avatarTypeJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'imageUrl',t.image_url,'iconUrl',t.icon_url,'sort',t.sort,'isActive',t.is_active,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'avatars',(SELECT count(*) FROM economy_avatar a WHERE a.type_id=t.id),
	'items',(SELECT count(*) FROM economy_avatar_item i WHERE t.id::text = ANY(i.type_ids)))`

const avatarCategoryJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'iconUrl',t.icon_url,'layer',t.layer,'required',t.required,'slots',t.slots,'sort',t.sort,'isActive',t.is_active,
	'createdAt',t.created_at,'updatedAt',t.updated_at,
	'count',(SELECT count(*) FROM economy_avatar_item i WHERE i.category_id=t.id),
	'avatars',(SELECT count(DISTINCT e.avatar_id) FROM economy_avatar_equip e WHERE e.category_id=t.id),
	'preview',(SELECT COALESCE(jsonb_agg(jsonb_build_object('key',p.key,'imageUrl',p.image_url,'iconUrl',p.icon_url,'value',p.value)),'[]'::jsonb)
		FROM (SELECT i.key, i.image_url, i.icon_url, i.value FROM economy_avatar_item i WHERE i.category_id=t.id ORDER BY i.sort, i.key LIMIT 6) p))`

const avatarItemJSON = `jsonb_build_object('id',t.id,'categoryId',t.category_id,'key',t.key,'name',t.name,'imageUrl',t.image_url,'iconUrl',t.icon_url,'value',t.value,'access',t.access,
	'price',t.price,'productId',t.product_id,'itemKey',t.item_key,'requirements',t.requirements,'typeIds',to_jsonb(t.type_ids),'isDefault',t.is_default,'sort',t.sort,'isActive',t.is_active,
	'createdAt',t.created_at,'updatedAt',t.updated_at,
	'categoryKey',(SELECT c.key FROM economy_avatar_category c WHERE c.id=t.category_id),
	'categoryName',(SELECT c.name FROM economy_avatar_category c WHERE c.id=t.category_id),
	'productKey',(SELECT p.key FROM economy_product p WHERE p.id=t.product_id),
	'types',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',y.id,'key',y.key,'name',y.name) ORDER BY y.sort, y.key),'[]'::jsonb) FROM economy_avatar_type y WHERE y.id::text = ANY(t.type_ids)),
	'owners',(SELECT count(*) FROM economy_avatar_unlock u WHERE u.item_id=t.id),
	'equipped',(SELECT count(*) FROM economy_avatar_equip e WHERE e.item_id=t.id))`

func avatarResources() []resource {
	return []resource{
		{
			name:    "avatar-types",
			table:   "economy_avatar_type",
			json:    avatarTypeJSON,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean"},
			save:    saveAvatarType,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_avatar WHERE type_id=$1) OR EXISTS (SELECT 1 FROM economy_avatar_item WHERE $1::uuid::text = ANY(type_ids))`)
			},
		},
		{
			name:    "avatar-categories",
			table:   "economy_avatar_category",
			json:    avatarCategoryJSON,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean"},
			save:    saveAvatarCategory,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_avatar_unlock u JOIN economy_avatar_item i ON i.id=u.item_id WHERE i.category_id=$1)`)
			},
		},
		{
			name:    "avatar-items",
			table:   "economy_avatar_item",
			json:    avatarItemJSON,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"categoryId": "t.category_id::text = ?", "access": "t.access = ?", "isActive": "t.is_active = ?::boolean", "typeId": "? = ANY(t.type_ids)"},
			save:    saveAvatarItem,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_avatar_unlock WHERE item_id=$1)`)
			},
		},
	}
}

func boolOr(v *bool, d bool) bool {
	if v == nil {
		return d
	}
	return *v
}

func saveAvatarType(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key      string `json:"key"`
		Name     i18n   `json:"name"`
		ImageURL string `json:"imageUrl"`
		IconURL  string `json:"iconUrl"`
		Sort     int    `json:"sort"`
		IsActive *bool  `json:"isActive"`
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
	if err := keepKey(ctx, tx, "economy_avatar_type", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 100)
	if err != nil {
		return err
	}
	img, err := cleanURL("imageUrl", in.ImageURL)
	if err != nil {
		return err
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	if in.Sort < -100000 || in.Sort > 100000 {
		return badRequest("INVALID_SORT", "sort", "between -100000 and 100000")
	}
	return execSave(ctx, tx, "economy_avatar_type", id, create, []string{"key", "name", "image_url", "icon_url", "sort", "is_active"},
		[]any{key, i18nJSON(name), img, icon, in.Sort, boolOr(in.IsActive, true)})
}

func saveAvatarCategory(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key      string `json:"key"`
		Name     i18n   `json:"name"`
		IconURL  string `json:"iconUrl"`
		Layer    int    `json:"layer"`
		Required bool   `json:"required"`
		Slots    *int   `json:"slots"`
		Sort     int    `json:"sort"`
		IsActive *bool  `json:"isActive"`
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
	if err := keepKey(ctx, tx, "economy_avatar_category", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 100)
	if err != nil {
		return err
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	if in.Layer < -1000 || in.Layer > 1000 {
		return badRequest("INVALID_LAYER", "layer", "between -1000 and 1000")
	}
	if in.Sort < -100000 || in.Sort > 100000 {
		return badRequest("INVALID_SORT", "sort", "between -100000 and 100000")
	}
	slots := 1
	if in.Slots != nil {
		slots = *in.Slots
	}
	if slots < 1 || slots > 20 {
		return badRequest("INVALID_SLOTS", "slots", "between 1 and 20")
	}
	if in.Required && !create {
		var items, plain int
		if err := tx.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE access='free' AND is_active AND (requirements IS NULL OR requirements='null'::jsonb) AND type_ids='{}') FROM economy_avatar_item WHERE category_id=$1`, id).Scan(&items, &plain); err != nil {
			return err
		}
		if items > 0 && plain == 0 {
			return badRequest("REQUIRED_NEEDS_FREE", "required", "a required category needs one free active item with no condition that fits every avatar type")
		}
	}
	return execSave(ctx, tx, "economy_avatar_category", id, create, []string{"key", "name", "icon_url", "layer", "required", "slots", "sort", "is_active"},
		[]any{key, i18nJSON(name), icon, in.Layer, in.Required, slots, in.Sort, boolOr(in.IsActive, true)})
}

func saveAvatarItem(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		CategoryID   string        `json:"categoryId"`
		Key          string        `json:"key"`
		Name         i18n          `json:"name"`
		ImageURL     string        `json:"imageUrl"`
		IconURL      string        `json:"iconUrl"`
		Value        string        `json:"value"`
		Access       string        `json:"access"`
		Price        *coinPrice    `json:"price"`
		ProductID    string        `json:"productId"`
		ItemKey      string        `json:"itemKey"`
		Requirements *gameReqInput `json:"requirements"`
		TypeIDs      []string      `json:"typeIds"`
		IsDefault    bool          `json:"isDefault"`
		Sort         int           `json:"sort"`
		IsActive     *bool         `json:"isActive"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	cat, err := refExists(ctx, tx, "economy_avatar_category", "categoryId", "AVATAR_CATEGORY_NOT_FOUND", in.CategoryID, true)
	if err != nil {
		return err
	}
	if !create {
		var old string
		if err := tx.QueryRowContext(ctx, `SELECT category_id::text FROM economy_avatar_item WHERE id=$1`, id).Scan(&old); err == nil && old != *cat {
			return badRequest("CATEGORY_IMMUTABLE", "categoryId", "an item cannot move to another category")
		}
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if err := keepKey(ctx, tx, "economy_avatar_item", id, create, key); err != nil {
		return err
	}
	var taken bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_avatar_item WHERE category_id=$1 AND key=$2 AND id<>$3)`, *cat, key, id).Scan(&taken); err != nil {
		return err
	}
	if taken {
		return fail(http.StatusConflict, "ALREADY_EXISTS", "key", "another item in this category has this key")
	}
	name, err := cleanI18n(langs, "name", in.Name, false, 100)
	if err != nil {
		return err
	}
	img, err := cleanURL("imageUrl", in.ImageURL)
	if err != nil {
		return err
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	value := strings.TrimSpace(in.Value)
	if len([]rune(value)) > 64 {
		return badRequest("INVALID_VALUE", "value", "at most 64 letters")
	}
	if img == "" && value == "" {
		return badRequest("OPTION_LOOK_REQUIRED", "imageUrl", "give the item a picture or a value")
	}
	access := strings.TrimSpace(in.Access)
	if access == "" {
		access = "free"
	}
	if err := oneOf("access", "INVALID_ACCESS", access, "free", "currency", "product", "item"); err != nil {
		return err
	}
	var price any
	var product any
	itemKey := ""
	switch access {
	case "currency":
		if in.Price == nil || strings.TrimSpace(in.Price.CurrencyKey) == "" {
			return badRequest("AVATAR_PRICE_REQUIRED", "price.currencyKey", "pick the currency")
		}
		if in.Price.Amount < 1 || in.Price.Amount > 1000000000000 {
			return badRequest("INVALID_PRICE", "price.amount", "between 1 and 1000000000000")
		}
		p := coinPrice{CurrencyKey: strings.TrimSpace(in.Price.CurrencyKey), Amount: in.Price.Amount}
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, p.CurrencyKey).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("CURRENCY_NOT_FOUND", "price.currencyKey", "currency does not exist")
		}
		rawPrice, _ := json.Marshal(p)
		price = string(rawPrice)
	case "product":
		ref, err := refExists(ctx, tx, "economy_product", "productId", "PRODUCT_NOT_FOUND", in.ProductID, true)
		if err != nil {
			return err
		}
		product = *ref
	case "item":
		itemKey = strings.TrimSpace(in.ItemKey)
		if itemKey == "" {
			return badRequest("ITEM_REQUIRED", "itemKey", "pick the item that opens this avatar item")
		}
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item WHERE key=$1)`, itemKey).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("ITEM_NOT_FOUND", "itemKey", "item does not exist")
		}
	}
	req, err := cleanGameRequirements(ctx, tx, langs, "requirements", in.Requirements, false)
	if err != nil {
		return err
	}
	var reqJSON any
	if req != nil {
		req.ConsumeItems, req.Refund, req.Paths, req.Inherit = false, "", nil, nil
		for i := range req.Metas {
			req.Metas[i].Consume = 0
		}
		if !req.Empty() {
			b, _ := json.Marshal(req)
			reqJSON = string(b)
		}
	}
	types := []string{}
	seen := map[string]bool{}
	for _, raw := range in.TypeIDs {
		ref, err := refExists(ctx, tx, "economy_avatar_type", "typeIds", "AVATAR_TYPE_NOT_FOUND", raw, true)
		if err != nil {
			return err
		}
		if !seen[*ref] {
			seen[*ref] = true
			types = append(types, *ref)
		}
	}
	if len(types) > 50 {
		return badRequest("TOO_MANY_TYPES", "typeIds", "at most 50 avatar types")
	}
	active := boolOr(in.IsActive, true)
	if in.IsDefault && (access != "free" || reqJSON != nil || !active) {
		return badRequest("DEFAULT_MUST_BE_FREE", "isDefault", "the default item must be free, active and without a condition")
	}
	if in.Sort < -100000 || in.Sort > 100000 {
		return badRequest("INVALID_SORT", "sort", "between -100000 and 100000")
	}
	if err := execSave(ctx, tx, "economy_avatar_item", id, create,
		[]string{"category_id", "key", "name", "image_url", "icon_url", "value", "access", "price", "product_id", "item_key", "requirements", "type_ids", "is_default", "sort", "is_active"},
		[]any{*cat, key, i18nJSON(name), img, icon, value, access, price, product, itemKey, reqJSON, types, in.IsDefault, in.Sort, active}); err != nil {
		return err
	}
	if in.IsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE economy_avatar_item SET is_default=false, updated_at=now() WHERE category_id=$1 AND id<>$2 AND is_default`, *cat, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) avatarConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/avatar-settings", http.MethodGet, s.consoleAvatarSettingsGet},
		{consolePrefix + "/avatar-settings", http.MethodPut, s.consoleAvatarSettingsPut},
		{consolePrefix + "/users/{userId}/avatars", http.MethodGet, s.consoleUserAvatars},
	}
}

func (s *service) consoleAvatarSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadAvatarSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *service) consoleAvatarSettingsPut(w http.ResponseWriter, r *http.Request) {
	st, err := loadAvatarSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	st.TagLimits = nil
	if err := readBody(r, &st); err != nil {
		s.writeError(w, err)
		return
	}
	if st.MaxAvatars < 1 || st.MaxAvatars > avatarMaxLimit {
		s.writeError(w, badRequest("INVALID_SETTING", "maxAvatars", "between 1 and 50"))
		return
	}
	if st.NameMaxLength < 1 || st.NameMaxLength > avatarMaxNameLength {
		s.writeError(w, badRequest("INVALID_SETTING", "nameMaxLength", "between 1 and 64"))
		return
	}
	if err := oneOf("slotFull", "INVALID_SETTING", st.SlotFull, avatarSlotReplace, avatarSlotRefuse); err != nil {
		s.writeError(w, err)
		return
	}
	if len(st.TagLimits) > avatarMaxTagLimits {
		s.writeError(w, badRequest("INVALID_SETTING", "tagLimits", "at most 20 rows"))
		return
	}
	limits := []avatarTagLimit{}
	seen := map[string]bool{}
	err = s.inTx(r.Context(), func(tx *sql.Tx) error {
		for i, l := range st.TagLimits {
			field := "tagLimits[" + itoa(i) + "]"
			ids, err := checkUserTags(r.Context(), tx, []string{l.TagID})
			if err != nil {
				if app, ok := asApp(err); ok {
					app.Field = field + ".tagId"
				}
				return err
			}
			if l.MaxAvatars < 1 || l.MaxAvatars > avatarMaxLimit {
				return badRequest("INVALID_SETTING", field+".maxAvatars", "between 1 and 50")
			}
			if seen[ids[0]] {
				return badRequest("DUPLICATE_TAG", field+".tagId", "each tag once")
			}
			seen[ids[0]] = true
			limits = append(limits, avatarTagLimit{TagID: ids[0], MaxAvatars: l.MaxAvatars})
		}
		st.TagLimits, st.UpdatedAt = limits, nil
		raw, _ := json.Marshal(st)
		_, err := tx.ExecContext(r.Context(), `INSERT INTO economy_avatar_settings (id, settings, updated_at) VALUES (1,$1::jsonb,now()) ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_at=now()`, string(raw))
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleAvatarSettingsGet(w, r)
}

func (s *service) consoleUserAvatars(w http.ResponseWriter, r *http.Request) {
	userID := pathVar(r, "userId")
	if _, err := uuid.FromString(userID); err != nil {
		s.writeError(w, badRequest("INVALID_ID", "userId", "must be a UUID"))
		return
	}
	ctx := r.Context()
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	st, err := loadAvatarSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	c, err := loadAvatarCatalog(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	x := &avCtx{s: s, st: st, cat: c, user: userID, langs: langs, lang: langs.Default}
	if x.own, err = s.avatarOwnership(ctx, s.db, c, []string{userID}); err != nil {
		s.writeError(w, err)
		return
	}
	list, err := loadAvatars(ctx, s.db, `a.user_id=$1::uuid`, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	views := []map[string]any{}
	for _, a := range list {
		views = append(views, x.avatarView(a))
	}
	bought := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT c.key, i.key, u.source, COALESCE(u.paid::text,'null'), u.created_at FROM economy_avatar_unlock u JOIN economy_avatar_item i ON i.id=u.item_id
		JOIN economy_avatar_category c ON c.id=i.category_id WHERE u.user_id=$1 ORDER BY u.created_at DESC LIMIT 200`, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cat, key, source, paid string
		var at sql.NullTime
		if err := rows.Scan(&cat, &key, &source, &paid, &at); err != nil {
			s.writeError(w, err)
			return
		}
		bought = append(bought, map[string]any{"category": cat, "key": key, "source": source, "paid": json.RawMessage(paid), "createdAt": at.Time})
	}
	if err := rows.Err(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, err)
		return
	}
	tags, err := s.userTagIDs(ctx, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"avatars": views, "unlocks": bought, "limit": map[string]any{"max": st.limitFor(tags), "count": len(list)}})
}
