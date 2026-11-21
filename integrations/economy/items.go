package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/heroiclabs/nakama/v3/internal/tagdiscount"
)

var itemRequestNS = uuid.Must(uuid.FromString("4f1c2a52-8d0e-4b8e-9a51-6f1f2d6b7c11"))

var itemRarities = []string{"common", "uncommon", "rare", "epic", "legendary"}

var serviceKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,31}$`)

const itemRequestTTL = 24 * time.Hour

type itemService struct {
	Key         string `json:"key"`
	Name        i18n   `json:"name"`
	CurrencyKey string `json:"currencyKey"`
	UnitPrice   int64  `json:"unitPrice"`
	MaxUnits    int    `json:"maxUnits"`
}

type itemDef struct {
	ID           string          `json:"id"`
	Key          string          `json:"key"`
	TypeID       string          `json:"typeId"`
	TypeKey      string          `json:"typeKey"`
	TypeName     i18n            `json:"typeName"`
	Name         i18n            `json:"name"`
	Description  i18n            `json:"description"`
	IconURL      string          `json:"iconUrl"`
	ImageURL     string          `json:"imageUrl"`
	SoundURL     string          `json:"soundUrl"`
	CategoryID   *string         `json:"categoryId"`
	CategoryKey  string          `json:"categoryKey"`
	TagIDs       []string        `json:"tagIds"`
	Tags         []string        `json:"tags"`
	Rarity       string          `json:"rarity"`
	ParentKey    string          `json:"parentKey"`
	RequiresKey  string          `json:"requiresKey"`
	IsStarter    bool            `json:"isStarter"`
	LevelTableID *string         `json:"levelTableId"`
	UnlockLevel  *int            `json:"unlockLevel"`
	Price        *coinPrice      `json:"price"`
	Upgrades     []coinPrice     `json:"upgrades"`
	Services     []itemService   `json:"services"`
	Metadata     json.RawMessage `json:"metadata"`
	Sort         int             `json:"sort"`
	IsActive     bool            `json:"isActive"`
	Giftable     *bool           `json:"giftable"`
}

func (d itemDef) needs() string {
	if d.RequiresKey != "" {
		return d.RequiresKey
	}
	return d.ParentKey
}

func (d itemDef) unlock() int {
	if d.UnlockLevel == nil {
		return 0
	}
	return *d.UnlockLevel
}

func (d itemDef) table() string {
	if d.LevelTableID == nil {
		return ""
	}
	return *d.LevelTableID
}

const itemJSON = `jsonb_build_object('id',t.id,'key',t.key,'typeId',t.type_id,'typeKey',(SELECT c.key FROM base_category c WHERE c.id=t.type_id),'typeName',(SELECT c.name FROM base_category c WHERE c.id=t.type_id),
	'name',t.name,'description',t.description,'iconUrl',t.icon_url,'imageUrl',t.image_url,'soundUrl',t.sound_url,'categoryId',t.category_id,'categoryKey',COALESCE((SELECT c.key FROM base_category c WHERE c.id=t.category_id),''),
	'tagIds',to_jsonb(t.tag_ids),'tags',COALESCE((SELECT jsonb_agg(g.key ORDER BY g.sort, g.key) FROM base_tag g WHERE g.id::text = ANY(t.tag_ids)),'[]'::jsonb),
	'rarity',t.rarity,'parentKey',COALESCE(t.parent_key,''),'requiresKey',COALESCE(t.requires_key,''),'isStarter',t.is_starter,'levelTableId',t.level_table_id,'unlockLevel',t.unlock_level,
	'price',t.price,'upgrades',t.upgrades,'services',t.services,'metadata',t.metadata,'sort',t.sort,'isActive',t.is_active,'giftable',t.giftable,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'owners',(SELECT count(*) FROM economy_user_item u WHERE u.item_id=t.id),'equippedBy',(SELECT count(*) FROM economy_user_item u WHERE u.item_id=t.id AND u.equipped))`

func itemResources() []resource {
	return []resource{{
		name:   "items",
		table:  "economy_item",
		json:   itemJSON,
		search: []string{"t.key", translatedSearch("t.name")},
		filters: map[string]string{
			"typeId": "t.type_id::text = ?", "categoryId": "t.category_id::text = ?", "isActive": "t.is_active = ?::boolean",
			"isStarter": "t.is_starter = ?::boolean", "rarity": "t.rarity = ?",
			"forSale": "(t.price IS NOT NULL) = ?::boolean",
		},
		save: saveItem,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_user_item WHERE item_id=$1)
				OR EXISTS (SELECT 1 FROM economy_item i JOIN economy_prize z ON z.items @> jsonb_build_array(jsonb_build_object('type','item','key',i.key)) WHERE i.id=$1)
				OR EXISTS (SELECT 1 FROM economy_item i JOIN economy_item o ON o.parent_key=i.key OR o.requires_key=i.key WHERE i.id=$1 AND o.id<>i.id)
				OR `+gameReqUsed("itemIds"))
		},
	}}
}

func cleanPrice(ctx context.Context, tx *sql.Tx, field string, p *coinPrice) (*coinPrice, error) {
	if p == nil {
		return nil, nil
	}
	p.CurrencyKey = strings.TrimSpace(p.CurrencyKey)
	if !currencyPattern.MatchString(p.CurrencyKey) {
		return nil, badRequest("INVALID_KEY", field, "invalid currency key")
	}
	if p.Amount <= 0 || p.Amount > 1_000_000_000_000 {
		return nil, badRequest("INVALID_PRICE", field, "price must be a positive whole number")
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, p.CurrencyKey).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, badRequest("CURRENCY_NOT_FOUND", field, "currency does not exist")
	}
	return p, nil
}

func saveItem(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key          string          `json:"key"`
		TypeID       string          `json:"typeId"`
		Name         i18n            `json:"name"`
		Description  i18n            `json:"description"`
		IconURL      string          `json:"iconUrl"`
		ImageURL     string          `json:"imageUrl"`
		SoundURL     string          `json:"soundUrl"`
		CategoryID   string          `json:"categoryId"`
		TagIDs       []string        `json:"tagIds"`
		Rarity       string          `json:"rarity"`
		ParentKey    string          `json:"parentKey"`
		RequiresKey  string          `json:"requiresKey"`
		IsStarter    bool            `json:"isStarter"`
		LevelTableID string          `json:"levelTableId"`
		UnlockLevel  *int            `json:"unlockLevel"`
		Price        *coinPrice      `json:"price"`
		Upgrades     []coinPrice     `json:"upgrades"`
		Services     []itemService   `json:"services"`
		Metadata     json.RawMessage `json:"metadata"`
		Sort         int             `json:"sort"`
		IsActive     *bool           `json:"isActive"`
		Giftable     *bool           `json:"giftable"`
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
	if !create {
		var old string
		if err := tx.QueryRowContext(ctx, `SELECT key FROM economy_item WHERE id=$1`, id).Scan(&old); err == nil && old != key {
			return badRequest("KEY_IMMUTABLE", "key", "the key of an item cannot change")
		}
	}
	typeID, err := cleanUUID("typeId", in.TypeID, true)
	if err != nil {
		return err
	}
	var kt string
	err = tx.QueryRowContext(ctx, `SELECT key_type FROM base_category WHERE id=$1`, *typeID).Scan(&kt)
	if errors.Is(err, sql.ErrNoRows) {
		return badRequest("ITEM_TYPE_NOT_FOUND", "typeId", "item type does not exist")
	}
	if err != nil {
		return err
	}
	if kt != "item_type" {
		return badRequest("CATEGORY_KIND_MISMATCH", "typeId", "the type must be an item type")
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	iconURL, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	imageURL, err := cleanURL("imageUrl", in.ImageURL)
	if err != nil {
		return err
	}
	soundURL, err := cleanURL("soundUrl", in.SoundURL)
	if err != nil {
		return err
	}
	category, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, category, "item"); err != nil {
		return err
	}
	tags, err := checkTags(ctx, tx, in.TagIDs, "item")
	if err != nil {
		return err
	}
	rarity := strings.TrimSpace(in.Rarity)
	if rarity == "" {
		rarity = "common"
	}
	okRarity := false
	for _, r := range itemRarities {
		if r == rarity {
			okRarity = true
		}
	}
	if !okRarity {
		return badRequest("INVALID_RARITY", "rarity", "unknown rarity")
	}
	refKey := func(field, v string) (any, error) {
		v = strings.TrimSpace(v)
		if v == "" {
			return nil, nil
		}
		if v == key {
			return nil, badRequest("SELF_REFERENCE", field, "an item cannot need itself")
		}
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item WHERE key=$1)`, v).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, badRequest("ITEM_NOT_FOUND", field, "item "+v+" does not exist")
		}
		return v, nil
	}
	parent, err := refKey("parentKey", in.ParentKey)
	if err != nil {
		return err
	}
	requires, err := refKey("requiresKey", in.RequiresKey)
	if err != nil {
		return err
	}
	if in.UnlockLevel != nil && (*in.UnlockLevel < 1 || *in.UnlockLevel > 10000) {
		return badRequest("INVALID_LEVEL", "unlockLevel", "unlock level must be 1 or more")
	}
	if in.UnlockLevel != nil && *in.UnlockLevel > 0 && in.IsStarter {
		return badRequest("STARTER_HAS_LEVEL", "unlockLevel", "a starter item is owned from the start")
	}
	table, err := cleanUUID("levelTableId", in.LevelTableID, false)
	if err != nil {
		return err
	}
	if table != nil {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_level_table WHERE id=$1)`, *table).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return badRequest("LEVEL_TABLE_NOT_FOUND", "levelTableId", "level table does not exist")
		}
	}
	price, err := cleanPrice(ctx, tx, "price", in.Price)
	if err != nil {
		return err
	}
	var priceJSON any
	if price != nil {
		b, _ := json.Marshal(price)
		priceJSON = string(b)
	}
	if len(in.Upgrades) > 50 {
		return badRequest("TOO_MANY_UPGRADES", "upgrades", "at most 50 upgrade levels")
	}
	upgrades := []coinPrice{}
	for i := range in.Upgrades {
		p, err := cleanPrice(ctx, tx, "upgrades["+itoa(i)+"]", &in.Upgrades[i])
		if err != nil {
			return err
		}
		upgrades = append(upgrades, *p)
	}
	if len(in.Services) > 20 {
		return badRequest("TOO_MANY_SERVICES", "services", "at most 20 services")
	}
	services := []itemService{}
	seen := map[string]bool{}
	for i, sv := range in.Services {
		field := "services[" + itoa(i) + "]"
		sv.Key = strings.TrimSpace(sv.Key)
		if !serviceKeyPattern.MatchString(sv.Key) {
			return badRequest("INVALID_KEY", field, "use lower-case letters, digits, _ . - (max 32)")
		}
		if seen[sv.Key] {
			return badRequest("DUPLICATE_SERVICE", field, "service keys must be unique")
		}
		seen[sv.Key] = true
		n, err := cleanI18n(langs, field, sv.Name, false, 64)
		if err != nil {
			return err
		}
		sv.Name = n
		p, err := cleanPrice(ctx, tx, field, &coinPrice{CurrencyKey: sv.CurrencyKey, Amount: sv.UnitPrice})
		if err != nil {
			return err
		}
		sv.CurrencyKey = p.CurrencyKey
		if sv.MaxUnits == 0 {
			sv.MaxUnits = 1
		}
		if sv.MaxUnits < 1 || sv.MaxUnits > 1000 {
			return badRequest("INVALID_UNITS", field, "max units must be 1 to 1000")
		}
		services = append(services, sv)
	}
	meta, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	upJSON, _ := json.Marshal(upgrades)
	svJSON, _ := json.Marshal(services)
	return execSave(ctx, tx, "economy_item", id, create,
		[]string{"key", "type_id", "name", "description", "icon_url", "image_url", "sound_url", "category_id", "tag_ids", "rarity", "parent_key", "requires_key", "is_starter", "level_table_id", "unlock_level", "price", "upgrades", "services", "metadata", "sort", "is_active", "giftable"},
		[]any{key, *typeID, i18nJSON(name), i18nJSON(desc), iconURL, imageURL, soundURL, category, tags, rarity, parent, requires, in.IsStarter, table, in.UnlockLevel, priceJSON, string(upJSON), string(svJSON), meta, in.Sort, active, in.Giftable})
}

type rowsQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func loadItems(ctx context.Context, q rowsQueryer, where string, args ...any) ([]itemDef, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+itemJSON+`::text FROM economy_item t `+where+` ORDER BY t.sort, t.key`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []itemDef{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var d itemDef
		if err := json.Unmarshal([]byte(raw), &d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func loadItemByKey(ctx context.Context, q queryer, key string) (itemDef, error) {
	var d itemDef
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+itemJSON+`::text FROM economy_item t WHERE t.key=$1`, strings.TrimSpace(key)).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return d, notFound("ITEM_NOT_FOUND")
	}
	if err != nil {
		return d, err
	}
	return d, json.Unmarshal([]byte(raw), &d)
}

const levelReachedSQL = `SELECT GREATEST(COALESCE((SELECT level FROM economy_user_level WHERE table_id=$1 AND user_id=$2),0),
	COALESCE((SELECT max(s.level) FROM economy_level_step s JOIN economy_level_table lt ON lt.id=s.table_id JOIN economy_meta m ON m.id=lt.meta_id
		LEFT JOIN economy_user_meta um ON um.meta_id=lt.meta_id AND um.user_id=$2 WHERE s.table_id=$1 AND s.threshold <= COALESCE(um.value, m.default_value)),0))`

type levelCache struct {
	userID string
	q      queryer
	byID   map[string]int
	main   *string
}

func newLevelCache(q queryer, userID string) *levelCache {
	return &levelCache{userID: userID, q: q, byID: map[string]int{}}
}

func (c *levelCache) level(ctx context.Context, tableID string) (int, error) {
	if tableID == "" {
		if c.main == nil {
			var id string
			err := c.q.QueryRowContext(ctx, `SELECT id::text FROM economy_level_table WHERE is_active ORDER BY sort, created_at LIMIT 1`).Scan(&id)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return 0, err
			}
			c.main = &id
		}
		tableID = *c.main
	}
	if v, ok := c.byID[tableID]; ok {
		return v, nil
	}
	var lv int
	if tableID == "" {
		var f sql.NullFloat64
		err := c.q.QueryRowContext(ctx, `SELECT COALESCE(um.value, m.default_value)::float8 FROM economy_meta m
			LEFT JOIN economy_user_meta um ON um.meta_id=m.id AND um.user_id=$1 WHERE m.key='level'`, c.userID).Scan(&f)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
		lv = int(f.Float64)
	} else if err := c.q.QueryRowContext(ctx, levelReachedSQL, tableID, c.userID).Scan(&lv); err != nil {
		return 0, err
	}
	c.byID[tableID] = lv
	return lv, nil
}

type ownedItem struct {
	Source       string
	SourceRef    string
	OwnedAt      time.Time
	UpgradeLevel int
	Equipped     bool
	Slot         string
	Version      int
}

func ownedItems(ctx context.Context, q rowsQueryer, userID string) (map[string]ownedItem, error) {
	rows, err := q.QueryContext(ctx, `SELECT item_id::text, source, source_ref, owned_at, upgrade_level, equipped, slot, version FROM economy_user_item WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]ownedItem{}
	for rows.Next() {
		var id string
		var o ownedItem
		if err := rows.Scan(&id, &o.Source, &o.SourceRef, &o.OwnedAt, &o.UpgradeLevel, &o.Equipped, &o.Slot, &o.Version); err != nil {
			return nil, err
		}
		out[id] = o
	}
	return out, rows.Err()
}

type execQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *service) ensureStarters(ctx context.Context, userID string) error {
	if err := s.ensureStarterItems(ctx, userID); err != nil {
		return err
	}
	return s.ensureLevelUnlocks(ctx, userID)
}

func (s *service) ensureLevelUnlocks(ctx context.Context, userID string) error {
	var missing bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item i WHERE i.is_active AND (i.price IS NULL OR i.price = 'null'::jsonb) AND COALESCE(i.unlock_level,0) > 0
		AND NOT EXISTS (SELECT 1 FROM economy_user_item u WHERE u.user_id=$1 AND u.item_id=i.id))`, userID).Scan(&missing); err != nil || !missing {
		return err
	}
	var added []itemDef
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		added = nil
		rows, err := tx.QueryContext(ctx, `SELECT i.key FROM economy_item i WHERE i.is_active AND (i.price IS NULL OR i.price = 'null'::jsonb) AND COALESCE(i.unlock_level,0) > 0
			AND NOT EXISTS (SELECT 1 FROM economy_user_item u WHERE u.user_id=$1 AND u.item_id=i.id) ORDER BY i.unlock_level, i.sort, i.key LIMIT 200`, userID)
		if err != nil {
			return err
		}
		var keys []string
		for rows.Next() {
			var k string
			if err := rows.Scan(&k); err != nil {
				rows.Close()
				return err
			}
			keys = append(keys, k)
		}
		rows.Close()
		for _, k := range keys {
			d, err := s.activeItem(ctx, tx, k)
			if err != nil {
				return err
			}
			if d.Price != nil || d.unlock() == 0 {
				continue
			}
			if err := s.checkItemAccess(ctx, tx, userID, d); err != nil {
				if _, ok := asApp(err); ok {
					continue
				}
				return err
			}
			res, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref) VALUES ($1,$2,'unlock',$3) ON CONFLICT (user_id, item_id) DO NOTHING`, userID, d.ID, "L"+itoa(d.unlock()))
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n > 0 {
				added = append(added, d)
			}
		}
		return nil
	})
	for _, d := range added {
		s.recordServerEvent(ctx, userID, "item_unlocked", float64(d.unlock()), map[string]any{"itemKey": d.Key, "auto": true})
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "unlocked", RefID: d.Key, Actor: "system", Data: map[string]any{"item": d.Key, "level": d.unlock()}})
	}
	return err
}

func (s *service) ensureStarterItems(ctx context.Context, userID string) error {
	var missing bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item i WHERE i.is_starter AND i.is_active
		AND NOT EXISTS (SELECT 1 FROM economy_user_item u WHERE u.user_id=$1 AND u.item_id=i.id))`, userID).Scan(&missing); err != nil || !missing {
		return err
	}
	var added []string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		added = nil
		rows, err := tx.QueryContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref)
			SELECT $1, i.id, 'starter', '' FROM economy_item i WHERE i.is_starter AND i.is_active ORDER BY i.sort, i.key
			ON CONFLICT (user_id, item_id) DO NOTHING RETURNING item_id::text`, userID)
		if err != nil {
			return err
		}
		var ids []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			ids = append(ids, id)
		}
		rows.Close()
		for _, id := range ids {
			var key string
			if err := tx.QueryRowContext(ctx, `UPDATE economy_user_item u SET equipped=true, slot=c.key, updated_at=now()
				FROM economy_item i JOIN base_category c ON c.id=i.type_id
				WHERE u.user_id=$1 AND u.item_id=$2 AND i.id=u.item_id
				AND NOT EXISTS (SELECT 1 FROM economy_user_item e WHERE e.user_id=$1 AND e.slot=c.key AND e.equipped)
				RETURNING i.key`, userID, id).Scan(&key); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			added = append(added, id)
		}
		return nil
	})
	if err == nil && len(added) > 0 {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "starter", Actor: "system", Data: map[string]any{"count": len(added)}})
	}
	return err
}

func (s *service) claimItemRequest(ctx context.Context, tx *sql.Tx, userID, requestID, action, itemKey, detail string) (json.RawMessage, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM economy_item_request WHERE user_id=$1 AND request_id=$2 AND expires_at < now()`, userID, requestID); err != nil {
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO economy_item_request (user_id, request_id, action, item_key, detail, expires_at) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (user_id, request_id) DO NOTHING`, userID, requestID, action, itemKey, detail, time.Now().Add(itemRequestTTL))
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return nil, nil
	}
	var a, k, d, result string
	if err := tx.QueryRowContext(ctx, `SELECT action, item_key, detail, result::text FROM economy_item_request WHERE user_id=$1 AND request_id=$2`, userID, requestID).Scan(&a, &k, &d, &result); err != nil {
		return nil, err
	}
	if a != action || k != itemKey || d != detail {
		return nil, conflict("REQUEST_ID_REUSED", "this requestId was used for another action")
	}
	return json.RawMessage(result), nil
}

func saveItemRequest(ctx context.Context, tx *sql.Tx, userID, requestID string, result any) error {
	b, _ := json.Marshal(result)
	_, err := tx.ExecContext(ctx, `UPDATE economy_item_request SET result=$3 WHERE user_id=$1 AND request_id=$2`, userID, requestID, string(b))
	return err
}

func cleanRequestID(raw string, required bool) (string, bool, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		if required {
			return "", false, badRequest("REQUEST_ID_REQUIRED", "requestId", "send a requestId of 1-64 characters")
		}
		return uuid.Must(uuid.NewV4()).String(), true, nil
	}
	if len(id) > 64 {
		return "", false, badRequest("REQUEST_ID_REQUIRED", "requestId", "send a requestId of 1-64 characters")
	}
	return id, false, nil
}

func requestGrantID(userID, requestID string) string {
	return uuid.NewV5(itemRequestNS, userID+":"+requestID).String()
}

func (s *service) walletChange(ctx context.Context, tx *sql.Tx, userID, grantID string, changes map[string]int64, meta map[string]any) error {
	for k, v := range changes {
		if v == 0 {
			delete(changes, k)
		}
	}
	if len(changes) == 0 {
		return nil
	}
	md := map[string]any{"module": "economy"}
	for k, v := range meta {
		md[k] = v
	}
	if err := pluginkit.WalletTx(ctx, tx, userID, grantID, changes, md); err != nil {
		var neg *pluginkit.ErrInsufficient
		if errors.As(err, &neg) {
			return fail(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "", "not enough "+neg.Path)
		}
		return err
	}
	return nil
}

func owns(ctx context.Context, q queryer, userID, key string) (bool, error) {
	var ok bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1 AND i.key=$2)`, userID, key).Scan(&ok)
	return ok, err
}

type itemView = map[string]any

func itemPublic(d itemDef, langs *localized.Set, lang string) itemView {
	services := []map[string]any{}
	for _, sv := range d.Services {
		services = append(services, map[string]any{"key": sv.Key, "name": langs.Resolve(sv.Name, lang), "currencyKey": sv.CurrencyKey, "unitPrice": sv.UnitPrice, "maxUnits": sv.MaxUnits})
	}
	var meta any = map[string]any{}
	if len(d.Metadata) > 0 {
		_ = json.Unmarshal(d.Metadata, &meta)
	}
	tags := d.Tags
	if tags == nil {
		tags = []string{}
	}
	return itemView{
		"id": d.ID, "key": d.Key, "type": d.TypeKey, "typeName": langs.Resolve(d.TypeName, lang), "name": langs.Resolve(d.Name, lang),
		"description": langs.Resolve(d.Description, lang), "iconUrl": d.IconURL, "imageUrl": d.ImageURL, "soundUrl": d.SoundURL, "rarity": d.Rarity,
		"categoryKey": d.CategoryKey, "tags": tags, "parentKey": d.ParentKey, "requiresKey": d.RequiresKey, "isStarter": d.IsStarter,
		"unlockLevel": d.unlock(), "price": d.Price, "maxUpgradeLevel": len(d.Upgrades), "services": services, "metadata": meta, "sort": d.Sort, "isActive": d.IsActive, "saleStopped": !d.IsActive,
	}
}

func (s *service) inventoryView(ctx context.Context, userID string, langs *localized.Set, lang, typeKey string, includeInactive bool) (map[string]any, error) {
	where := `WHERE t.is_active`
	if includeInactive {
		where = `WHERE (t.is_active OR EXISTS (SELECT 1 FROM economy_user_item u WHERE u.item_id=t.id AND u.user_id=$1))`
	} else {
		where += ` AND $1::uuid IS NOT NULL`
	}
	defs, err := loadItems(ctx, s.db, where, userID)
	if err != nil {
		return nil, err
	}
	owned, err := ownedItems(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	lc := newLevelCache(s.db, userID)
	disc := s.userDiscounts(ctx, s.db, userID)
	mainLevel, err := lc.level(ctx, "")
	if err != nil {
		return nil, err
	}
	ownedKeys := map[string]bool{}
	for _, d := range defs {
		if _, ok := owned[d.ID]; ok {
			ownedKeys[d.Key] = true
		}
	}
	states := economyhooks.Items(ctx, userID, lang, itemIDs(defs))
	items := []itemView{}
	equipped := map[string]string{}
	counts := map[string]int{"owned": 0, "total": 0}
	for _, d := range defs {
		o, has := owned[d.ID]
		if has && o.Equipped {
			equipped[o.Slot] = d.Key
		}
		if typeKey != "" && d.TypeKey != typeKey {
			continue
		}
		v, err := s.itemState(ctx, d, o, has, lc, ownedKeys, langs, lang, states[d.ID])
		if err != nil {
			return nil, err
		}
		itemDiscountView(v, d, disc)
		counts["total"]++
		if has {
			counts["owned"]++
		}
		items = append(items, v)
	}
	return map[string]any{"lang": lang, "level": mainLevel, "items": items, "equipped": equipped, "counts": counts}, nil
}

func (s *service) itemState(ctx context.Context, d itemDef, o ownedItem, has bool, lc *levelCache, ownedKeys map[string]bool, langs *localized.Set, lang string, ext economyhooks.ItemState) (itemView, error) {
	v := itemPublic(d, langs, lang)
	level, err := lc.level(ctx, d.table())
	if err != nil {
		return nil, err
	}
	need := d.needs()
	needMet := need == "" || ownedKeys[need]
	locked := !has && d.unlock() > 0 && level < d.unlock() && !ext.Unlocked
	v["owned"] = has
	v["equipped"] = has && o.Equipped
	v["locked"] = locked
	v["playerLevel"] = level
	v["requirementMet"] = needMet
	v["canBuy"] = !has && d.IsActive && d.Price != nil && !locked && needMet
	v["canUnlock"] = !has && d.IsActive && d.Price == nil && d.unlock() > 0 && !locked && needMet
	v["upgradeLevel"] = 0
	v["nextUpgradePrice"] = nil
	if has {
		v["source"] = o.Source
		v["ownedAt"] = o.OwnedAt
		v["upgradeLevel"] = o.UpgradeLevel
		v["version"] = o.Version
		if o.UpgradeLevel < len(d.Upgrades) {
			v["nextUpgradePrice"] = d.Upgrades[o.UpgradeLevel]
		}
	}
	economyhooks.Fill(v, ext.Fields)
	return v, nil
}

func (s *service) oneItemView(ctx context.Context, userID, key string, langs *localized.Set, lang string) (itemView, error) {
	d, err := loadItemByKey(ctx, s.db, key)
	if err != nil {
		return nil, err
	}
	owned, err := ownedItems(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	ownedKeys := map[string]bool{}
	if len(owned) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT i.key FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1`, userID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var k string
			if err := rows.Scan(&k); err == nil {
				ownedKeys[k] = true
			}
		}
		rows.Close()
	}
	o, has := owned[d.ID]
	v, err := s.itemState(ctx, d, o, has, newLevelCache(s.db, userID), ownedKeys, langs, lang, economyhooks.Items(ctx, userID, lang, []string{d.ID})[d.ID])
	if err != nil {
		return nil, err
	}
	itemDiscountView(v, d, s.userDiscounts(ctx, s.db, userID))
	return v, nil
}

func itemDiscountView(v itemView, d itemDef, set *tagdiscount.Set) {
	v["finalPrice"] = d.Price
	v["discountPercent"] = 0
	v["discounts"] = []string{}
	if d.Price == nil {
		return
	}
	m := set.Item(d.Key, d.TypeID)
	if m.Percent <= 0 {
		return
	}
	v["finalPrice"] = coinPrice{CurrencyKey: d.Price.CurrencyKey, Amount: tagdiscount.Apply(d.Price.Amount, m.Percent)}
	v["discountPercent"] = m.Percent
	v["discounts"] = m.Rules
}

type itemResult struct {
	ItemKey  string      `json:"itemKey"`
	Action   string      `json:"action"`
	Paid     []coinPrice `json:"paid"`
	Level    int         `json:"level,omitempty"`
	Service  string      `json:"service,omitempty"`
	Units    int         `json:"units,omitempty"`
	Replayed bool        `json:"replayed,omitempty"`
	fresh    bool
	value    float64
}

func decodeItemResult(raw json.RawMessage) itemResult {
	var r itemResult
	_ = json.Unmarshal(raw, &r)
	if r.Paid == nil {
		r.Paid = []coinPrice{}
	}
	r.Replayed = true
	return r
}

func (s *service) activeItem(ctx context.Context, tx *sql.Tx, key string) (itemDef, error) {
	d, err := loadItemByKey(ctx, tx, key)
	if err != nil {
		return d, err
	}
	if !d.IsActive {
		return d, notFound("ITEM_NOT_FOUND")
	}
	return d, nil
}

func (s *service) checkItemAccess(ctx context.Context, tx *sql.Tx, userID string, d itemDef) error {
	if d.unlock() > 0 {
		lv, err := newLevelCache(tx, userID).level(ctx, d.table())
		if err != nil {
			return err
		}
		if lv < d.unlock() {
			st := economyhooks.Items(ctx, userID, "", []string{d.ID})[d.ID]
			if !st.Unlocked {
				e := conflict("LEVEL_TOO_LOW", "the player's level is too low for this item")
				e.Meta = map[string]any{"level": lv, "unlockLevel": d.unlock()}
				economyhooks.Fill(e.Meta, st.LockMeta)
				return e
			}
		}
	}
	if need := d.needs(); need != "" {
		ok, err := owns(ctx, tx, userID, need)
		if err != nil {
			return err
		}
		if !ok {
			return conflict("REQUIRES_ITEM", "the player must own "+need+" first")
		}
	}
	return nil
}

func (s *service) buyItem(ctx context.Context, userID, key, rawRequest string) (itemResult, error) {
	requestID, _, err := cleanRequestID(rawRequest, false)
	if err != nil {
		return itemResult{}, err
	}
	if err := s.ensureStarters(ctx, userID); err != nil {
		return itemResult{}, err
	}
	key = strings.TrimSpace(key)
	grantID := requestGrantID(userID, requestID)
	disc := s.userDiscounts(ctx, s.db, userID)
	var out itemResult
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		prev, err := s.claimItemRequest(ctx, tx, userID, requestID, "buy", key, "")
		if err != nil {
			return err
		}
		if prev != nil {
			out = decodeItemResult(prev)
			return nil
		}
		d, err := s.activeItem(ctx, tx, key)
		if err != nil {
			return err
		}
		if d.Price == nil {
			return conflict("NOT_FOR_SALE", "this item is not for sale")
		}
		if err := s.checkItemAccess(ctx, tx, userID, d); err != nil {
			return err
		}
		if err := economyhooks.CheckItemBuy(ctx, tx, economyhooks.ItemBuy{UserID: userID, ItemID: d.ID, ItemKey: d.Key, TypeKey: d.TypeKey, ParentKey: d.ParentKey}); err != nil {
			return hookError(err)
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref) VALUES ($1,$2,'buy',$3) ON CONFLICT (user_id, item_id) DO NOTHING`, userID, d.ID, grantID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return conflict("ALREADY_OWNED", "the player already owns this item")
		}
		price := *d.Price
		meta := map[string]any{"reason": "item_buy", "sourceType": "item", "sourceId": d.Key, "itemKey": d.Key}
		if m := disc.Item(d.Key, d.TypeID); m.Percent > 0 {
			price.Amount = tagdiscount.Apply(price.Amount, m.Percent)
			meta["discountPercent"] = m.Percent
			meta["discounts"] = m.Rules
			meta["originalAmount"] = d.Price.Amount
		}
		if price.Amount > 0 {
			if err := s.walletChange(ctx, tx, userID, grantID, map[string]int64{price.CurrencyKey: -price.Amount}, meta); err != nil {
				return err
			}
		}
		out = itemResult{ItemKey: d.Key, Action: "buy", Paid: []coinPrice{price}, fresh: true, value: float64(price.Amount)}
		return saveItemRequest(ctx, tx, userID, requestID, out)
	})
	if err != nil {
		return out, err
	}
	if out.fresh {
		s.recordServerEvent(ctx, userID, "item_bought", out.value, map[string]any{"itemKey": out.ItemKey, "currency": out.Paid[0].CurrencyKey})
		activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "bought", RefID: out.ItemKey, Actor: "player", Data: map[string]any{"item": out.ItemKey, "price": out.Paid[0].Amount, "currency": out.Paid[0].CurrencyKey}}))
	}
	return out, nil
}

func (s *service) unlockItem(ctx context.Context, userID, key string) (itemResult, error) {
	if err := s.ensureStarterItems(ctx, userID); err != nil {
		return itemResult{}, err
	}
	var out itemResult
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		d, err := s.activeItem(ctx, tx, key)
		if err != nil {
			return err
		}
		if d.Price != nil || d.unlock() == 0 {
			return conflict("NOT_UNLOCKABLE", "this item is not unlocked by level")
		}
		if err := s.checkItemAccess(ctx, tx, userID, d); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref) VALUES ($1,$2,'unlock',$3) ON CONFLICT (user_id, item_id) DO NOTHING`, userID, d.ID, "L"+itoa(d.unlock()))
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return conflict("ALREADY_OWNED", "the player already owns this item")
		}
		out = itemResult{ItemKey: d.Key, Action: "unlock", Paid: []coinPrice{}, Level: d.unlock(), fresh: true, value: float64(d.unlock())}
		return nil
	})
	if err == nil && out.fresh {
		s.recordServerEvent(ctx, userID, "item_unlocked", out.value, map[string]any{"itemKey": out.ItemKey})
		activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "unlocked", RefID: out.ItemKey, Actor: "player", Data: map[string]any{"item": out.ItemKey, "level": out.Level}}))
	}
	return out, err
}

func (s *service) upgradeItem(ctx context.Context, userID, key string, fromLevel *int, rawRequest string) (itemResult, error) {
	requestID, _, err := cleanRequestID(rawRequest, false)
	if err != nil {
		return itemResult{}, err
	}
	if err := s.ensureStarters(ctx, userID); err != nil {
		return itemResult{}, err
	}
	detail := ""
	if fromLevel != nil {
		detail = itoa(*fromLevel)
	}
	key = strings.TrimSpace(key)
	grantID := requestGrantID(userID, requestID)
	var out itemResult
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		prev, err := s.claimItemRequest(ctx, tx, userID, requestID, "upgrade", key, detail)
		if err != nil {
			return err
		}
		if prev != nil {
			out = decodeItemResult(prev)
			return nil
		}
		d, err := s.activeItem(ctx, tx, key)
		if err != nil {
			return err
		}
		var level int
		err = tx.QueryRowContext(ctx, `SELECT upgrade_level FROM economy_user_item WHERE user_id=$1 AND item_id=$2 FOR UPDATE`, userID, d.ID).Scan(&level)
		if errors.Is(err, sql.ErrNoRows) {
			return conflict("NOT_OWNED", "the player does not own this item")
		}
		if err != nil {
			return err
		}
		if fromLevel != nil && *fromLevel != level {
			return conflict("UPGRADE_CHANGED", "the upgrade level changed; read the item again")
		}
		if level >= len(d.Upgrades) {
			return conflict("UPGRADE_MAXED", "this item is at its last upgrade level")
		}
		price := d.Upgrades[level]
		if err := s.walletChange(ctx, tx, userID, grantID, map[string]int64{price.CurrencyKey: -price.Amount}, map[string]any{"reason": "item_upgrade", "sourceType": "item", "sourceId": d.Key, "itemKey": d.Key, "level": level + 1}); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_item SET upgrade_level=$3, version=version+1, updated_at=now() WHERE user_id=$1 AND item_id=$2`, userID, d.ID, level+1); err != nil {
			return err
		}
		out = itemResult{ItemKey: d.Key, Action: "upgrade", Paid: []coinPrice{price}, Level: level + 1, fresh: true, value: float64(level + 1)}
		return saveItemRequest(ctx, tx, userID, requestID, out)
	})
	if err == nil && out.fresh {
		s.recordServerEvent(ctx, userID, "item_upgraded", out.value, map[string]any{"itemKey": out.ItemKey, "level": out.Level})
		activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "upgraded", RefID: out.ItemKey, Actor: "player", Data: map[string]any{"item": out.ItemKey, "level": out.Level, "price": out.Paid[0].Amount, "currency": out.Paid[0].CurrencyKey}}))
	}
	return out, err
}

func (s *service) equipItem(ctx context.Context, userID, key string, on bool) (string, []string, error) {
	if err := s.ensureStarters(ctx, userID); err != nil {
		return "", nil, err
	}
	var slot string
	unequipped := []string{}
	changed := false
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		unequipped = []string{}
		d, err := loadItemByKey(ctx, tx, key)
		if err != nil {
			return err
		}
		if on && !d.IsActive {
			return notFound("ITEM_NOT_FOUND")
		}
		slot = d.TypeKey
		typeID := d.TypeID
		if custom := economyhooks.Slot(ctx, d.ID, d.TypeKey); custom != "" {
			slot, typeID = custom, ""
		}
		rows, err := tx.QueryContext(ctx, `SELECT u.item_id::text, u.equipped, i.key FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id
			WHERE u.user_id=$1 AND (i.type_id::text=$2 OR u.slot=$3 OR u.item_id::text=$4) ORDER BY u.item_id FOR UPDATE OF u`, userID, typeID, slot, d.ID)
		if err != nil {
			return err
		}
		found := false
		was := false
		others := map[string]string{}
		for rows.Next() {
			var id, k string
			var eq bool
			if err := rows.Scan(&id, &eq, &k); err != nil {
				rows.Close()
				return err
			}
			if id == d.ID {
				found = true
				was = eq
			} else if eq {
				others[id] = k
			}
		}
		rows.Close()
		if !found {
			return conflict("NOT_OWNED", "the player does not own this item")
		}
		if !on {
			if was {
				changed = true
				_, err := tx.ExecContext(ctx, `UPDATE economy_user_item SET equipped=false, version=version+1, updated_at=now() WHERE user_id=$1 AND item_id=$2`, userID, d.ID)
				return err
			}
			return nil
		}
		for id, k := range others {
			if _, err := tx.ExecContext(ctx, `UPDATE economy_user_item SET equipped=false, version=version+1, updated_at=now() WHERE user_id=$1 AND item_id=$2`, userID, id); err != nil {
				return err
			}
			unequipped = append(unequipped, k)
		}
		sort.Strings(unequipped)
		if was {
			return nil
		}
		changed = true
		_, err = tx.ExecContext(ctx, `UPDATE economy_user_item SET equipped=true, slot=$3, version=version+1, updated_at=now() WHERE user_id=$1 AND item_id=$2`, userID, d.ID, slot)
		return err
	})
	if err == nil && on && changed {
		s.recordServerEvent(ctx, userID, "item_equipped", 1, map[string]any{"itemKey": key, "slot": slot})
	}
	return slot, unequipped, err
}

func (s *service) useItem(ctx context.Context, userID, key, service string, units int, rawRequest string) (itemResult, error) {
	requestID, _, err := cleanRequestID(rawRequest, true)
	if err != nil {
		return itemResult{}, err
	}
	if err := s.ensureStarters(ctx, userID); err != nil {
		return itemResult{}, err
	}
	if units == 0 {
		units = 1
	}
	key = strings.TrimSpace(key)
	service = strings.TrimSpace(service)
	grantID := requestGrantID(userID, requestID)
	var out itemResult
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		prev, err := s.claimItemRequest(ctx, tx, userID, requestID, "use", key, service+"#"+itoa(units))
		if err != nil {
			return err
		}
		if prev != nil {
			out = decodeItemResult(prev)
			return nil
		}
		d, err := s.activeItem(ctx, tx, key)
		if err != nil {
			return err
		}
		ok, err := owns(ctx, tx, userID, d.Key)
		if err != nil {
			return err
		}
		if !ok {
			return conflict("NOT_OWNED", "the player does not own this item")
		}
		var sv *itemService
		for i := range d.Services {
			if d.Services[i].Key == service {
				sv = &d.Services[i]
			}
		}
		if sv == nil {
			return notFound("SERVICE_NOT_FOUND")
		}
		if units < 1 || units > sv.MaxUnits {
			return badRequest("INVALID_UNITS", "units", "units must be 1 to "+itoa(sv.MaxUnits))
		}
		paid := coinPrice{CurrencyKey: sv.CurrencyKey, Amount: sv.UnitPrice * int64(units)}
		if err := s.walletChange(ctx, tx, userID, grantID, map[string]int64{paid.CurrencyKey: -paid.Amount}, map[string]any{"reason": "item_service", "sourceType": "item", "sourceId": d.Key, "itemKey": d.Key, "service": sv.Key, "requestId": requestID}); err != nil {
			return err
		}
		out = itemResult{ItemKey: d.Key, Action: "use", Paid: []coinPrice{paid}, Service: sv.Key, Units: units, fresh: true, value: float64(units)}
		return saveItemRequest(ctx, tx, userID, requestID, out)
	})
	if err == nil && out.fresh {
		s.recordServerEvent(ctx, userID, "item_used", out.value, map[string]any{"itemKey": out.ItemKey, "service": out.Service})
	}
	return out, err
}

func (s *service) adminGrantItem(ctx context.Context, userID, itemID, adminID string) (bool, error) {
	if _, err := uuid.FromString(itemID); err != nil {
		return false, notFound("ITEM_NOT_FOUND")
	}
	var key string
	if err := s.db.QueryRowContext(ctx, `SELECT key FROM economy_item WHERE id=$1`, itemID).Scan(&key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, notFound("ITEM_NOT_FOUND")
		}
		return false, err
	}
	res, err := s.db.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref) VALUES ($1,$2,'admin',$3) ON CONFLICT (user_id, item_id) DO NOTHING`, userID, itemID, adminID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "granted", RefID: key, Actor: "admin", Data: map[string]any{"item": key, "admin": adminID}})
	}
	return n > 0, nil
}

func (s *service) adminRevokeItem(ctx context.Context, userID, itemID, adminID string) (bool, error) {
	if _, err := uuid.FromString(itemID); err != nil {
		return false, notFound("ITEM_NOT_FOUND")
	}
	var key string
	var starter bool
	if err := s.db.QueryRowContext(ctx, `SELECT key, is_starter FROM economy_item WHERE id=$1`, itemID).Scan(&key, &starter); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, notFound("ITEM_NOT_FOUND")
		}
		return false, err
	}
	if starter {
		return false, conflict("STARTER_ITEM", "starter items belong to every player and cannot be taken back")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM economy_user_item WHERE user_id=$1 AND item_id=$2`, userID, itemID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "revoked", RefID: key, Actor: "admin", Data: map[string]any{"item": key, "admin": adminID}})
	}
	return n > 0, nil
}

func (s *service) sweepItemRequests(ctx context.Context) {
	_, _ = s.db.ExecContext(ctx, `DELETE FROM economy_item_request WHERE expires_at < now()`)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM economy_user_item WHERE user_id IN (SELECT DISTINCT u.user_id FROM economy_user_item u LEFT JOIN users x ON x.id=u.user_id WHERE x.id IS NULL LIMIT 200)`)
}
