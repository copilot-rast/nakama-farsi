package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
)

type resource struct {
	name    string
	table   string
	json    string
	search  []string
	filters map[string]string
	save    func(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error
	remove  func(s *service, ctx context.Context, tx *sql.Tx, id string) error
}

func resources() []resource {
	return []resource{
		{
			name:    "categories",
			table:   "economy_category",
			json:    `jsonb_build_object('id',t.id,'kind',t.kind,'key',t.key,'name',t.name,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at)`,
			search:  []string{"t.key", "t.name->>'fa'", "t.name->>'en'"},
			filters: map[string]string{"kind": "t.kind = ?"},
			save:    saveCategory,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE category_id=$1) OR EXISTS (SELECT 1 FROM economy_product WHERE category_id=$1)
					OR EXISTS (SELECT 1 FROM publication_faq WHERE category_id=$1) OR EXISTS (SELECT 1 FROM publication_news WHERE category_id=$1)`)
			},
		},
		{
			name:    "tags",
			table:   "economy_tag",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'createdAt',t.created_at,'updatedAt',t.updated_at)`,
			search:  []string{"t.key", "t.name->>'fa'", "t.name->>'en'"},
			filters: map[string]string{},
			save:    saveTag,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				if _, err := tx.ExecContext(ctx, `UPDATE economy_prize SET tag_ids = array_remove(tag_ids, $1::text), updated_at=now() WHERE $1::text = ANY(tag_ids)`, id); err != nil {
					return err
				}
				_, err := tx.ExecContext(ctx, `UPDATE economy_product SET tag_ids = array_remove(tag_ids, $1::text), updated_at=now() WHERE $1::text = ANY(tag_ids)`, id)
				return err
			},
		},
		{
			name:    "currencies",
			table:   "economy_currency",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'type',t.type,'defaultValue',t.default_value,'iconUrl',t.icon_url,'metadata',t.metadata,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at)`,
			search:  []string{"t.key", "t.name->>'fa'", "t.name->>'en'"},
			filters: map[string]string{"type": "t.type = ?"},
			save:    saveCurrency,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_currency c JOIN economy_product p ON p.coin_price->>'currencyKey' = c.key WHERE c.id=$1)
					OR EXISTS (SELECT 1 FROM economy_currency c JOIN economy_prize z ON z.items @> jsonb_build_array(jsonb_build_object('type','currency','key',c.key)) WHERE c.id=$1)`)
			},
		},
		{
			name:    "metas",
			table:   "economy_meta",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'defaultValue',t.default_value,'serverOnly',t.server_only,'settings',t.settings,'createdAt',t.created_at,'updatedAt',t.updated_at)`,
			search:  []string{"t.key", "t.name->>'fa'", "t.name->>'en'"},
			filters: map[string]string{},
			save:    saveMeta,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_meta m JOIN economy_prize z ON z.items @> jsonb_build_array(jsonb_build_object('type','meta','key',m.key)) WHERE m.id=$1)`)
			},
		},
		{
			name:    "prizes",
			table:   "economy_prize",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'message',t.message,'categoryId',t.category_id,'tagIds',to_jsonb(t.tag_ids),'isRepeatable',t.is_repeatable,'items',t.items,'metadata',t.metadata,'createdAt',t.created_at,'updatedAt',t.updated_at)`,
			search:  []string{"t.key", "t.name->>'fa'", "t.name->>'en'"},
			filters: map[string]string{"categoryId": "t.category_id::text = ?"},
			save:    savePrize,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_product WHERE prize_id=$1) OR EXISTS (SELECT 1 FROM economy_gift WHERE prize_id=$1) OR EXISTS (SELECT 1 FROM economy_user_prize WHERE prize_id=$1)`)
			},
		},
		{
			name:    "products",
			table:   "economy_product",
			json:    productJSON,
			search:  []string{"t.key", "t.name->>'fa'", "t.name->>'en'"},
			filters: map[string]string{"categoryId": "t.category_id::text = ?", "isActive": "t.is_active = ?::boolean", "method": "? = ANY(t.payment_methods)"},
			save:    saveProduct,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_payment WHERE product_id=$1)`)
			},
		},
		{
			name:    "gifts",
			table:   "economy_gift",
			json:    `jsonb_build_object('id',t.id,'name',t.name,'type',t.type,'code',t.code,'prizeId',t.prize_id,'discountPercent',t.discount_percent,'isPromo',t.is_promo,'perUserLimit',t.per_user_limit,'maxTotalUses',t.max_total_uses,'totalUsed',t.total_used,'validFrom',t.valid_from,'expireAt',t.expire_at,'isActive',t.is_active,'createdAt',t.created_at,'updatedAt',t.updated_at)`,
			search:  []string{"t.name", "t.code"},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "type": "t.type = ?"},
			save:    saveGift,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_gift_use WHERE gift_id=$1) OR EXISTS (SELECT 1 FROM economy_payment WHERE gift_id=$1)`)
			},
		},
	}
}

const productJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'isActive',t.is_active,'categoryId',t.category_id,'tagIds',to_jsonb(t.tag_ids),'priceToman',t.price_toman,'coinPrice',t.coin_price,'discountPercent',t.discount_percent,'paymentMethods',to_jsonb(t.payment_methods),'sku',t.sku,'isSubscription',t.is_subscription,'prizeId',t.prize_id,'iconUrl',t.icon_url,'coverUrl',t.cover_url,'minLevel',t.min_level,'metadata',t.metadata,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at)`

func refuseIfUsed(ctx context.Context, tx *sql.Tx, id, query string) error {
	var used bool
	if err := tx.QueryRowContext(ctx, query, id).Scan(&used); err != nil {
		return err
	}
	if used {
		return conflict("IN_USE", "this record is referenced and cannot be deleted")
	}
	return nil
}

func (s *service) resourceList(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := readPage(r)
		where := &sqlWhere{}
		if p.Q != "" {
			parts := make([]string, 0, len(res.search))
			pattern := likePattern(p.Q)
			n := where.next()
			for _, col := range res.search {
				parts = append(parts, "COALESCE("+col+",'') ILIKE "+n)
			}
			where.args = append(where.args, pattern)
			where.parts = append(where.parts, "("+strings.Join(parts, " OR ")+")")
		}
		for param, expr := range res.filters {
			if v := strings.TrimSpace(r.URL.Query().Get(param)); v != "" {
				if strings.Contains(expr, "boolean") && v != "true" && v != "false" {
					s.writeError(w, badRequest("INVALID_FILTER", param, "must be true or false"))
					return
				}
				where.add(expr, v)
			}
		}
		var total int64
		if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM `+res.table+` t`+where.sql(), where.args...).Scan(&total); err != nil {
			s.writeError(w, err)
			return
		}
		args := append(append([]any{}, where.args...), p.Limit, p.Offset)
		n := len(where.args)
		rows, err := s.db.QueryContext(r.Context(), `SELECT `+res.json+`::text FROM `+res.table+` t`+where.sql()+
			` ORDER BY t.created_at DESC, t.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2), args...)
		if err != nil {
			s.writeError(w, err)
			return
		}
		defer rows.Close()
		items := []json.RawMessage{}
		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				s.writeError(w, err)
				return
			}
			items = append(items, json.RawMessage(raw))
		}
		if err := rows.Err(); err != nil {
			s.writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit})
	}
}

func (s *service) fetchResource(ctx context.Context, q queryer, res resource, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+res.json+`::text FROM `+res.table+` t WHERE t.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *service) resourceGet(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := s.fetchResource(r.Context(), s.db, res, pathVar(r, "id"))
		if err != nil {
			s.writeError(w, err)
			return
		}
		writeRaw(w, http.StatusOK, raw)
	}
}

func (s *service) resourceWrite(res resource, create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body json.RawMessage
		if err := readBody(r, &body); err != nil {
			s.writeError(w, err)
			return
		}
		id := pathVar(r, "id")
		if create {
			id = uuid.Must(uuid.NewV4()).String()
		} else if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			if err := res.save(s, r.Context(), tx, id, create, body); err != nil {
				return err
			}
			var err error
			out, err = s.fetchResource(r.Context(), tx, res, id)
			return err
		})
		if err != nil {
			s.writeError(w, mapWriteError(err))
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeRaw(w, status, out)
	}
}

func (s *service) resourceDelete(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathVar(r, "id")
		if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			var found string
			if err := tx.QueryRowContext(r.Context(), `SELECT id::text FROM `+res.table+` WHERE id=$1 FOR UPDATE`, id).Scan(&found); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return notFound("NOT_FOUND")
				}
				return err
			}
			if err := res.remove(s, r.Context(), tx, id); err != nil {
				return err
			}
			_, err := tx.ExecContext(r.Context(), `DELETE FROM `+res.table+` WHERE id=$1`, id)
			return err
		})
		if err != nil {
			if isForeignKeyViolation(err) {
				err = conflict("IN_USE", "this record is referenced and cannot be deleted")
			}
			s.writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{})
	}
}

func mapWriteError(err error) error {
	if constraint, ok := isUniqueViolation(err); ok {
		field := "key"
		if strings.Contains(constraint, "code") {
			field = "code"
		}
		return fail(http.StatusConflict, "ALREADY_EXISTS", field, "a record with this "+field+" already exists")
	}
	if isForeignKeyViolation(err) {
		return badRequest("REFERENCE_NOT_FOUND", "", "a referenced record does not exist")
	}
	return err
}

func (s *service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func execSave(ctx context.Context, tx *sql.Tx, table, id string, create bool, cols []string, vals []any) error {
	if create {
		ph := make([]string, len(cols))
		for i := range cols {
			ph[i] = "$" + itoa(i+2)
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO `+table+` (id,`+strings.Join(cols, ",")+`) VALUES ($1,`+strings.Join(ph, ",")+`)`, append([]any{id}, vals...)...)
		return err
	}
	sets := make([]string, len(cols))
	for i, c := range cols {
		sets[i] = c + "=$" + itoa(i+2)
	}
	res, err := tx.ExecContext(ctx, `UPDATE `+table+` SET `+strings.Join(sets, ",")+`, updated_at=now() WHERE id=$1`, append([]any{id}, vals...)...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return notFound("NOT_FOUND")
	}
	return nil
}

func i18nJSON(t i18n) string {
	b, _ := json.Marshal(t)
	return string(b)
}

func stringsJSON(v []string) string {
	if v == nil {
		v = []string{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func itoa(n int) string {
	return strings.TrimSpace(jsonInt(int64(n)))
}

func decodeInto(raw []byte, out any) error {
	if err := json.Unmarshal(raw, out); err != nil {
		return badRequest("INVALID_JSON", "", "invalid body: "+err.Error())
	}
	return nil
}

func checkCategory(ctx context.Context, tx *sql.Tx, id *string, kind string) error {
	if id == nil {
		return nil
	}
	var k string
	err := tx.QueryRowContext(ctx, `SELECT kind FROM economy_category WHERE id=$1`, *id).Scan(&k)
	if errors.Is(err, sql.ErrNoRows) {
		return badRequest("CATEGORY_NOT_FOUND", "categoryId", "category does not exist")
	}
	if err != nil {
		return err
	}
	if k != kind {
		return badRequest("CATEGORY_KIND_MISMATCH", "categoryId", "category kind must be "+kind)
	}
	return nil
}

func checkTags(ctx context.Context, tx *sql.Tx, ids []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, raw := range ids {
		id, err := cleanUUID("tagIds", raw, true)
		if err != nil {
			return nil, err
		}
		if seen[*id] {
			continue
		}
		seen[*id] = true
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_tag WHERE id=$1)`, *id).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, badRequest("TAG_NOT_FOUND", "tagIds", "tag does not exist")
		}
		out = append(out, *id)
	}
	return out, nil
}

func saveCategory(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Kind string `json:"kind"`
		Key  string `json:"key"`
		Name i18n   `json:"name"`
		Sort int    `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	switch in.Kind {
	case "product", "prize", "gift", "currency", "faq", "news":
	default:
		return badRequest("INVALID_KIND", "kind", "kind must be product, prize, gift, currency, faq or news")
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n("name", in.Name, true, 128)
	if err != nil {
		return err
	}
	if !create {
		var oldKind string
		if err := tx.QueryRowContext(ctx, `SELECT kind FROM economy_category WHERE id=$1`, id).Scan(&oldKind); err == nil && oldKind != in.Kind {
			var used bool
			_ = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE category_id=$1) OR EXISTS (SELECT 1 FROM economy_product WHERE category_id=$1)
				OR EXISTS (SELECT 1 FROM publication_faq WHERE category_id=$1) OR EXISTS (SELECT 1 FROM publication_news WHERE category_id=$1)`, id).Scan(&used)
			if used {
				return badRequest("IN_USE", "kind", "a used category cannot change kind")
			}
		}
	}
	return execSave(ctx, tx, "economy_category", id, create, []string{"kind", "key", "name", "sort"}, []any{in.Kind, key, i18nJSON(name), in.Sort})
}

func saveTag(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key  string `json:"key"`
		Name i18n   `json:"name"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n("name", in.Name, true, 128)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "economy_tag", id, create, []string{"key", "name"}, []any{key, i18nJSON(name)})
}

func saveCurrency(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key          string          `json:"key"`
		Name         i18n            `json:"name"`
		Type         string          `json:"type"`
		DefaultValue int64           `json:"defaultValue"`
		IconURL      string          `json:"iconUrl"`
		Metadata     json.RawMessage `json:"metadata"`
		Sort         int             `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	key := strings.TrimSpace(in.Key)
	if !currencyPattern.MatchString(key) {
		return badRequest("INVALID_KEY", "key", "use 2 to 32 lower-case letters, digits or _")
	}
	if !create {
		var old string
		if err := tx.QueryRowContext(ctx, `SELECT key FROM economy_currency WHERE id=$1`, id).Scan(&old); err == nil && old != key {
			return badRequest("KEY_IMMUTABLE", "key", "a currency key is the wallet entry name and cannot change")
		}
	}
	name, err := cleanI18n("name", in.Name, true, 128)
	if err != nil {
		return err
	}
	if in.Type != "soft" && in.Type != "hard" {
		return badRequest("INVALID_TYPE", "type", "type must be soft or hard")
	}
	if in.DefaultValue < 0 {
		return badRequest("INVALID_VALUE", "defaultValue", "default value cannot be negative")
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	meta, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "economy_currency", id, create, []string{"key", "name", "type", "default_value", "icon_url", "metadata", "sort"},
		[]any{key, i18nJSON(name), in.Type, in.DefaultValue, icon, meta, in.Sort})
}

func saveMeta(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key          string          `json:"key"`
		Name         i18n            `json:"name"`
		DefaultValue json.RawMessage `json:"defaultValue"`
		ServerOnly   bool            `json:"serverOnly"`
		Settings     json.RawMessage `json:"settings"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if !create {
		var old string
		if err := tx.QueryRowContext(ctx, `SELECT key FROM economy_meta WHERE id=$1`, id).Scan(&old); err == nil && old != key {
			return badRequest("KEY_IMMUTABLE", "key", "a meta key is used by prize items and cannot change")
		}
	}
	name, err := cleanI18n("name", in.Name, true, 128)
	if err != nil {
		return err
	}
	def := "0"
	if len(in.DefaultValue) > 0 && string(in.DefaultValue) != "null" {
		v, ok := numericFromJSON(in.DefaultValue)
		if !ok {
			return badRequest("INVALID_VALUE", "defaultValue", "must be a number with at most 4 decimals")
		}
		def = v
	}
	settingsJSON, err := cleanObject("settings", in.Settings)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "economy_meta", id, create, []string{"key", "name", "default_value", "server_only", "settings"},
		[]any{key, i18nJSON(name), def, in.ServerOnly, settingsJSON})
}

func checkPrizeItems(ctx context.Context, tx *sql.Tx, items []prizeItem) error {
	for _, item := range items {
		var exists bool
		switch item.Type {
		case "currency":
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, item.Key).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return badRequest("CURRENCY_NOT_FOUND", "items", "currency "+item.Key+" does not exist")
			}
		case "meta":
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_meta WHERE key=$1)`, item.Key).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return badRequest("META_NOT_FOUND", "items", "meta "+item.Key+" does not exist")
			}
		}
	}
	return nil
}

func savePrize(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key          string          `json:"key"`
		Name         i18n            `json:"name"`
		Message      i18n            `json:"message"`
		CategoryID   string          `json:"categoryId"`
		TagIDs       []string        `json:"tagIds"`
		IsRepeatable *bool           `json:"isRepeatable"`
		Items        []prizeItem     `json:"items"`
		Metadata     json.RawMessage `json:"metadata"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n("name", in.Name, true, 128)
	if err != nil {
		return err
	}
	message, err := cleanI18n("message", in.Message, false, 1000)
	if err != nil {
		return err
	}
	category, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, category, "prize"); err != nil {
		return err
	}
	tags, err := checkTags(ctx, tx, in.TagIDs)
	if err != nil {
		return err
	}
	items, err := normalizePrizeItems(in.Items)
	if err != nil {
		return err
	}
	if err := checkPrizeItems(ctx, tx, items); err != nil {
		return err
	}
	repeatable := true
	if in.IsRepeatable != nil {
		repeatable = *in.IsRepeatable
	}
	meta, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	itemsJSON, _ := json.Marshal(items)
	return execSave(ctx, tx, "economy_prize", id, create,
		[]string{"key", "name", "message", "category_id", "tag_ids", "is_repeatable", "items", "metadata"},
		[]any{key, i18nJSON(name), i18nJSON(message), category, tags, repeatable, string(itemsJSON), meta})
}

type coinPrice struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

type productInput struct {
	Key             string            `json:"key"`
	Name            i18n              `json:"name"`
	Description     i18n              `json:"description"`
	IsActive        bool              `json:"isActive"`
	CategoryID      string            `json:"categoryId"`
	TagIDs          []string          `json:"tagIds"`
	PriceToman      *int64            `json:"priceToman"`
	CoinPrice       *coinPrice        `json:"coinPrice"`
	DiscountPercent int               `json:"discountPercent"`
	PaymentMethods  []string          `json:"paymentMethods"`
	SKU             map[string]string `json:"sku"`
	IsSubscription  bool              `json:"isSubscription"`
	PrizeID         string            `json:"prizeId"`
	IconURL         string            `json:"iconUrl"`
	CoverURL        string            `json:"coverUrl"`
	MinLevel        int               `json:"minLevel"`
	Metadata        json.RawMessage   `json:"metadata"`
	Sort            int               `json:"sort"`
}

func validateProductShape(in *productInput) error {
	methods := []string{}
	seen := map[string]bool{}
	for _, m := range in.PaymentMethods {
		if !validMethod(m) {
			return badRequest("INVALID_PAYMENT_METHOD", "paymentMethods", "unknown payment method "+m)
		}
		if !seen[m] {
			seen[m] = true
			methods = append(methods, m)
		}
	}
	in.PaymentMethods = methods
	if in.DiscountPercent < 0 || in.DiscountPercent > 100 {
		return badRequest("INVALID_DISCOUNT", "discountPercent", "discount must be 0 to 100")
	}
	if in.MinLevel < 0 {
		return badRequest("INVALID_VALUE", "minLevel", "min level cannot be negative")
	}
	if in.PriceToman != nil && *in.PriceToman < 0 {
		return badRequest("INVALID_PRICE", "priceToman", "price cannot be negative")
	}
	if in.CoinPrice != nil {
		in.CoinPrice.CurrencyKey = strings.TrimSpace(in.CoinPrice.CurrencyKey)
		if !currencyPattern.MatchString(in.CoinPrice.CurrencyKey) {
			return badRequest("INVALID_KEY", "coinPrice", "invalid currency key")
		}
		if in.CoinPrice.Amount <= 0 {
			return badRequest("INVALID_PRICE", "coinPrice", "coin price must be positive")
		}
	}
	sku := map[string]string{}
	for k, v := range in.SKU {
		v = strings.TrimSpace(v)
		if k != "bazaar" && k != "myket" {
			return badRequest("INVALID_SKU", "sku", "sku keys are bazaar and myket")
		}
		if v == "" {
			continue
		}
		if !skuPattern.MatchString(v) {
			return badRequest("INVALID_SKU", "sku", "invalid sku")
		}
		sku[k] = v
	}
	in.SKU = sku
	for _, m := range in.PaymentMethods {
		switch {
		case isGateway(m):
			if in.PriceToman == nil || *in.PriceToman < 100 {
				return badRequest("PRICE_REQUIRED", "priceToman", "gateway methods need a price of at least 100 Toman")
			}
			if discounted(*in.PriceToman, in.DiscountPercent) != 0 && discounted(*in.PriceToman, in.DiscountPercent) < 100 {
				return badRequest("PRICE_TOO_LOW", "discountPercent", "the discounted price must be 0 or at least 100 Toman")
			}
		case m == "in_app":
			if in.CoinPrice == nil {
				return badRequest("COIN_PRICE_REQUIRED", "coinPrice", "in_app needs a coin price")
			}
		case isStore(m):
			if sku[m] == "" {
				return badRequest("SKU_REQUIRED", "sku", m+" needs a sku")
			}
			if m == "myket" && in.IsSubscription {
				return badRequest("SUBSCRIPTION_NOT_SUPPORTED", "isSubscription", "Myket subscriptions are not supported")
			}
		}
	}
	return nil
}

func saveProduct(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in productInput
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n("name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n("description", in.Description, false, 4000)
	if err != nil {
		return err
	}
	if err := validateProductShape(&in); err != nil {
		return err
	}
	category, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, category, "product"); err != nil {
		return err
	}
	tags, err := checkTags(ctx, tx, in.TagIDs)
	if err != nil {
		return err
	}
	prize, err := cleanUUID("prizeId", in.PrizeID, true)
	if err != nil {
		return err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE id=$1)`, *prize).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return badRequest("PRIZE_NOT_FOUND", "prizeId", "prize does not exist")
	}
	var coin any
	if in.CoinPrice != nil {
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, in.CoinPrice.CurrencyKey).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return badRequest("CURRENCY_NOT_FOUND", "coinPrice", "currency does not exist")
		}
		b, _ := json.Marshal(in.CoinPrice)
		coin = string(b)
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	cover, err := cleanURL("coverUrl", in.CoverURL)
	if err != nil {
		return err
	}
	meta, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	skuJSON, _ := json.Marshal(in.SKU)
	return execSave(ctx, tx, "economy_product", id, create,
		[]string{"key", "name", "description", "is_active", "category_id", "tag_ids", "price_toman", "coin_price", "discount_percent", "payment_methods", "sku", "is_subscription", "prize_id", "icon_url", "cover_url", "min_level", "metadata", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), in.IsActive, category, tags, in.PriceToman, coin, in.DiscountPercent, in.PaymentMethods, string(skuJSON), in.IsSubscription, *prize, icon, cover, in.MinLevel, meta, in.Sort})
}

func saveGift(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Name            string  `json:"name"`
		Type            string  `json:"type"`
		Code            *string `json:"code"`
		PrizeID         string  `json:"prizeId"`
		DiscountPercent *int    `json:"discountPercent"`
		IsPromo         bool    `json:"isPromo"`
		PerUserLimit    *int    `json:"perUserLimit"`
		MaxTotalUses    *int    `json:"maxTotalUses"`
		ValidFrom       string  `json:"validFrom"`
		ExpireAt        string  `json:"expireAt"`
		IsActive        *bool   `json:"isActive"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" || len([]rune(name)) > 128 {
		return badRequest("INVALID_NAME", "name", "name is required (max 128)")
	}
	var code *string
	if in.Code != nil && strings.TrimSpace(*in.Code) != "" {
		c := strings.ToUpper(strings.TrimSpace(*in.Code))
		if !giftCodePattern.MatchString(c) {
			return badRequest("INVALID_CODE", "code", "code must be 3 to 32 of A-Z 0-9 _ -")
		}
		code = &c
	}
	var prize *string
	var percent *int
	var err error
	switch in.Type {
	case "reward":
		if prize, err = cleanUUID("prizeId", in.PrizeID, true); err != nil {
			return err
		}
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE id=$1)`, *prize).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return badRequest("PRIZE_NOT_FOUND", "prizeId", "prize does not exist")
		}
	case "discount":
		if in.DiscountPercent == nil || *in.DiscountPercent < 1 || *in.DiscountPercent > 100 {
			return badRequest("INVALID_DISCOUNT", "discountPercent", "discount must be 1 to 100")
		}
		percent = in.DiscountPercent
		if code == nil {
			return badRequest("CODE_REQUIRED", "code", "a discount needs a code")
		}
	default:
		return badRequest("INVALID_TYPE", "type", "type must be reward or discount")
	}
	perUser := 1
	if in.PerUserLimit != nil {
		perUser = *in.PerUserLimit
	}
	if perUser < 1 || perUser > 1000000 {
		return badRequest("INVALID_LIMIT", "perUserLimit", "per-user limit must be at least 1")
	}
	if in.MaxTotalUses != nil && *in.MaxTotalUses < 1 {
		return badRequest("INVALID_LIMIT", "maxTotalUses", "max total uses must be at least 1 or empty")
	}
	from, err := cleanTime("validFrom", in.ValidFrom)
	if err != nil {
		return err
	}
	expire, err := cleanTime("expireAt", in.ExpireAt)
	if err != nil {
		return err
	}
	if from != nil && expire != nil && !expire.After(*from) {
		return badRequest("INVALID_WINDOW", "expireAt", "expireAt must be after validFrom")
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	return execSave(ctx, tx, "economy_gift", id, create,
		[]string{"name", "type", "code", "prize_id", "discount_percent", "is_promo", "per_user_limit", "max_total_uses", "valid_from", "expire_at", "is_active"},
		[]any{name, in.Type, code, prize, percent, in.IsPromo, perUser, in.MaxTotalUses, from, expire, active})
}
