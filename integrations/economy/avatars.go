package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	avatarMaxIDs        = 100
	avatarMaxLimit      = 50
	avatarMaxTagLimits  = 20
	avatarMaxNameLength = 64
	avatarSlotReplace   = "replace"
	avatarSlotRefuse    = "refuse"
)

type avatarTagLimit struct {
	TagID      string `json:"tagId"`
	MaxAvatars int    `json:"maxAvatars"`
}

type avatarSettings struct {
	Enabled       bool             `json:"enabled"`
	MaxAvatars    int              `json:"maxAvatars"`
	TagLimits     []avatarTagLimit `json:"tagLimits"`
	NameMaxLength int              `json:"nameMaxLength"`
	AutoCreate    bool             `json:"autoCreate"`
	KeepLast      bool             `json:"keepLast"`
	SlotFull      string           `json:"slotFull"`
	KeepBought    bool             `json:"keepBought"`
	UpdatedAt     *time.Time       `json:"updatedAt,omitempty"`
}

func defaultAvatarSettings() avatarSettings {
	return avatarSettings{Enabled: true, MaxAvatars: 1, TagLimits: []avatarTagLimit{}, NameMaxLength: 24, AutoCreate: true, KeepLast: true, SlotFull: avatarSlotReplace, KeepBought: true}
}

func (st *avatarSettings) normalize() {
	d := defaultAvatarSettings()
	if st.MaxAvatars < 1 || st.MaxAvatars > avatarMaxLimit {
		st.MaxAvatars = d.MaxAvatars
	}
	if st.NameMaxLength < 1 || st.NameMaxLength > avatarMaxNameLength {
		st.NameMaxLength = d.NameMaxLength
	}
	if st.SlotFull != avatarSlotReplace && st.SlotFull != avatarSlotRefuse {
		st.SlotFull = d.SlotFull
	}
	if st.TagLimits == nil {
		st.TagLimits = []avatarTagLimit{}
	}
}

func loadAvatarSettings(ctx context.Context, q queryer) (avatarSettings, error) {
	st := defaultAvatarSettings()
	var raw string
	var at time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_at FROM economy_avatar_settings WHERE id=1`).Scan(&raw, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	_ = json.Unmarshal([]byte(raw), &st)
	st.normalize()
	st.UpdatedAt = &at
	return st, nil
}

func (st avatarSettings) limitFor(tags map[string]bool) int {
	max := st.MaxAvatars
	for _, l := range st.TagLimits {
		if tags[l.TagID] && l.MaxAvatars > max {
			max = l.MaxAvatars
		}
	}
	return max
}

type avType struct {
	ID       string
	Key      string
	Name     i18n
	ImageURL string
	IconURL  string
}

type avCategory struct {
	ID       string
	Key      string
	Name     i18n
	IconURL  string
	Layer    int
	Slots    int
	Required bool
	Items    []*avItem
}

type avItem struct {
	ID        string
	Key       string
	Cat       *avCategory
	Name      i18n
	ImageURL  string
	IconURL   string
	Value     string
	Access    string
	Price     *coinPrice
	ProductID string
	ItemKey   string
	Req       *gamecatalog.Requirements
	TypeIDs   []string
	IsDefault bool
}

func (it *avItem) fits(typeID string) bool {
	if len(it.TypeIDs) == 0 {
		return true
	}
	for _, id := range it.TypeIDs {
		if id == typeID {
			return true
		}
	}
	return false
}

func (it *avItem) plain() bool {
	return it.Access == "free" && it.Req.Empty()
}

type avCatalog struct {
	types    []*avType
	typeByID map[string]*avType
	typeKey  map[string]*avType
	cats     []*avCategory
	catKey   map[string]*avCategory
	items    map[string]*avItem
}

func (c *avCatalog) find(catKey, key string) *avItem {
	cat := c.catKey[strings.TrimSpace(catKey)]
	if cat == nil {
		return nil
	}
	key = strings.TrimSpace(key)
	for _, it := range cat.Items {
		if it.Key == key {
			return it
		}
	}
	return nil
}

func (c *avCatalog) fallback(cat *avCategory, typeID string) *avItem {
	for _, it := range cat.Items {
		if it.IsDefault && it.plain() && it.fits(typeID) {
			return it
		}
	}
	if cat.Required {
		for _, it := range cat.Items {
			if it.plain() && it.fits(typeID) {
				return it
			}
		}
	}
	return nil
}

func (c *avCatalog) defaultType() string {
	if len(c.types) == 0 {
		return ""
	}
	return c.types[0].ID
}

func loadAvatarCatalog(ctx context.Context, q rowsQueryer) (*avCatalog, error) {
	c := &avCatalog{typeByID: map[string]*avType{}, typeKey: map[string]*avType{}, catKey: map[string]*avCategory{}, items: map[string]*avItem{}}
	rows, err := q.QueryContext(ctx, `SELECT id::text, key, name::text, image_url, icon_url FROM economy_avatar_type WHERE is_active ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		t := &avType{}
		var name string
		if err := rows.Scan(&t.ID, &t.Key, &name, &t.ImageURL, &t.IconURL); err != nil {
			rows.Close()
			return nil, err
		}
		t.Name = parseI18n(name)
		c.types = append(c.types, t)
		c.typeByID[t.ID], c.typeKey[t.Key] = t, t
	}
	rows.Close()
	byID := map[string]*avCategory{}
	rows, err = q.QueryContext(ctx, `SELECT id::text, key, name::text, icon_url, layer, slots, required FROM economy_avatar_category WHERE is_active ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		cat := &avCategory{}
		var name string
		if err := rows.Scan(&cat.ID, &cat.Key, &name, &cat.IconURL, &cat.Layer, &cat.Slots, &cat.Required); err != nil {
			rows.Close()
			return nil, err
		}
		cat.Name = parseI18n(name)
		c.cats = append(c.cats, cat)
		c.catKey[cat.Key], byID[cat.ID] = cat, cat
	}
	rows.Close()
	if len(c.cats) == 0 {
		return c, nil
	}
	rows, err = q.QueryContext(ctx, `SELECT i.id::text, i.category_id::text, i.key, i.name::text, i.image_url, i.icon_url, i.value, i.access, COALESCE(i.price::text,''),
			COALESCE(i.product_id::text,''), i.item_key, COALESCE(i.requirements::text,''), i.type_ids::text, i.is_default
		FROM economy_avatar_item i JOIN economy_avatar_category c ON c.id=i.category_id WHERE i.is_active AND c.is_active ORDER BY i.sort, i.key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		it := &avItem{}
		var catID, name, price, req, types string
		if err := rows.Scan(&it.ID, &catID, &it.Key, &name, &it.ImageURL, &it.IconURL, &it.Value, &it.Access, &price, &it.ProductID, &it.ItemKey, &req, &types, &it.IsDefault); err != nil {
			return nil, err
		}
		it.Name = parseI18n(name)
		if price != "" && price != "null" {
			p := coinPrice{}
			if json.Unmarshal([]byte(price), &p) == nil && p.CurrencyKey != "" {
				it.Price = &p
			}
		}
		if req != "" && req != "null" {
			r := &gamecatalog.Requirements{}
			if json.Unmarshal([]byte(req), r) == nil && !r.Empty() {
				it.Req = r
			}
		}
		it.TypeIDs = parseTextArray(types)
		cat := byID[catID]
		if cat == nil {
			continue
		}
		it.Cat = cat
		cat.Items = append(cat.Items, it)
		c.items[it.ID] = it
	}
	return c, rows.Err()
}

type avOwnership struct {
	unlocked map[string]map[string]bool
	shop     map[string]map[string]bool
	products map[string]map[string]bool
}

func (o avOwnership) owns(userID string, it *avItem) bool {
	switch it.Access {
	case "free":
		return true
	case "currency":
		return o.unlocked[userID][it.ID]
	case "product":
		return o.unlocked[userID][it.ID] || o.products[userID][it.ProductID]
	case "item":
		return o.shop[userID][it.ItemKey]
	}
	return false
}

func addPair(m map[string]map[string]bool, a, b string) {
	if m[a] == nil {
		m[a] = map[string]bool{}
	}
	m[a][b] = true
}

func scanPairs(rows *sql.Rows, err error, into map[string]map[string]bool) error {
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var a, b string
		if err := rows.Scan(&a, &b); err != nil {
			return err
		}
		addPair(into, a, b)
	}
	return rows.Err()
}

func (s *service) avatarOwnership(ctx context.Context, q rowsQueryer, c *avCatalog, users []string) (avOwnership, error) {
	o := avOwnership{unlocked: map[string]map[string]bool{}, shop: map[string]map[string]bool{}, products: map[string]map[string]bool{}}
	if len(users) == 0 {
		return o, nil
	}
	shopKeys, productIDs, paid := []string{}, []string{}, false
	seen := map[string]bool{}
	for _, it := range c.items {
		switch it.Access {
		case "currency":
			paid = true
		case "product":
			paid = true
			if !seen["p"+it.ProductID] {
				seen["p"+it.ProductID] = true
				productIDs = append(productIDs, it.ProductID)
			}
		case "item":
			if !seen["i"+it.ItemKey] {
				seen["i"+it.ItemKey] = true
				shopKeys = append(shopKeys, it.ItemKey)
			}
		}
	}
	if paid {
		rows, err := q.QueryContext(ctx, `SELECT user_id::text, item_id::text FROM economy_avatar_unlock WHERE user_id::text = ANY($1::text[])`, users)
		if err := scanPairs(rows, err, o.unlocked); err != nil {
			return o, err
		}
	}
	if len(shopKeys) > 0 {
		rows, err := q.QueryContext(ctx, `SELECT DISTINCT u.user_id::text, i.key FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id::text = ANY($1::text[]) AND i.key = ANY($2::text[])`, users, shopKeys)
		if err := scanPairs(rows, err, o.shop); err != nil {
			return o, err
		}
	}
	if len(productIDs) > 0 {
		rows, err := q.QueryContext(ctx, `SELECT DISTINCT user_id::text, product_id::text FROM economy_payment WHERE user_id::text = ANY($1::text[]) AND product_id::text = ANY($2::text[])
			AND status='success' AND friend_gift_id IS NULL AND COALESCE(outcome,'') <> 'duplicate'`, users, productIDs)
		if err := scanPairs(rows, err, o.products); err != nil {
			return o, err
		}
	}
	return o, nil
}

func (s *service) avatarFacts(ctx context.Context, c *avCatalog, userID string) (reqFacts, error) {
	items := false
	metas := []string{}
	seen := map[string]bool{}
	level := false
	for _, it := range c.items {
		if it.Req.Empty() {
			continue
		}
		items = items || len(it.Req.ItemIDs) > 0
		level = level || it.Req.MinLevel > 0
		for _, m := range it.Req.Metas {
			if !seen[m.MetaID] {
				seen[m.MetaID] = true
				metas = append(metas, m.MetaID)
			}
		}
	}
	lv := 0
	if level {
		var err error
		if lv, err = s.playerLevel(ctx, userID); err != nil {
			return reqFacts{}, err
		}
	}
	return s.loadReqFacts(ctx, s.db, userID, lv, items, metas)
}

func avReqOK(it *avItem, f reqFacts) (bool, []reqMiss) {
	if it.Req.Empty() {
		return true, nil
	}
	miss := reqMissing(it.Req, f)
	return len(miss) == 0, miss
}

func (st avatarSettings) usable(it *avItem, owned, reqOK bool) bool {
	if !owned {
		return false
	}
	return reqOK || (it.Access != "free" && st.KeepBought)
}

type avEquip struct {
	ItemID     string
	CategoryID string
}

type avRow struct {
	ID        string
	UserID    string
	TypeID    string
	Name      string
	Active    bool
	Cleared   map[string]bool
	CreatedAt time.Time
	Equips    []avEquip
}

func loadAvatars(ctx context.Context, q rowsQueryer, where string, args ...any) ([]*avRow, error) {
	rows, err := q.QueryContext(ctx, `SELECT a.id::text, a.user_id::text, COALESCE(a.type_id::text,''), a.name, a.is_active, a.cleared::text, a.created_at FROM economy_avatar a WHERE `+where+` ORDER BY a.created_at, a.id`, args...)
	if err != nil {
		return nil, err
	}
	out := []*avRow{}
	at := map[string]*avRow{}
	ids := []string{}
	for rows.Next() {
		a := &avRow{Cleared: map[string]bool{}}
		var cleared string
		if err := rows.Scan(&a.ID, &a.UserID, &a.TypeID, &a.Name, &a.Active, &cleared, &a.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		for _, id := range parseTextArray(cleared) {
			a.Cleared[id] = true
		}
		out = append(out, a)
		at[a.ID] = a
		ids = append(ids, a.ID)
	}
	rows.Close()
	if err := rows.Err(); err != nil || len(ids) == 0 {
		return out, err
	}
	erows, err := q.QueryContext(ctx, `SELECT avatar_id::text, item_id::text, category_id::text FROM economy_avatar_equip WHERE avatar_id::text = ANY($1::text[]) ORDER BY equipped_at, item_id`, ids)
	if err != nil {
		return nil, err
	}
	defer erows.Close()
	for erows.Next() {
		var aid string
		var e avEquip
		if err := erows.Scan(&aid, &e.ItemID, &e.CategoryID); err != nil {
			return nil, err
		}
		if a := at[aid]; a != nil {
			a.Equips = append(a.Equips, e)
		}
	}
	return out, erows.Err()
}

type avLayer struct {
	Category string `json:"category"`
	Item     string `json:"item"`
	ImageURL string `json:"imageUrl"`
	Value    string `json:"value"`
	Layer    int    `json:"layer"`
}

type avLook struct {
	Items  map[string][]string `json:"items"`
	Layers []avLayer           `json:"layers"`
}

type avResolved struct {
	cat   *avCategory
	items []*avItem
	fill  bool
}

func (c *avCatalog) resolve(a *avRow, owned func(*avItem) bool) []avResolved {
	typeID := ""
	cleared := map[string]bool{}
	by := map[string][]*avItem{}
	if a != nil {
		typeID, cleared = a.TypeID, a.Cleared
		for _, e := range a.Equips {
			it := c.items[e.ItemID]
			if it == nil || it.Cat.ID != e.CategoryID || !it.fits(typeID) || !owned(it) {
				continue
			}
			by[it.Cat.ID] = append(by[it.Cat.ID], it)
		}
	}
	out := []avResolved{}
	for _, cat := range c.cats {
		list := by[cat.ID]
		if len(list) > cat.Slots {
			list = list[len(list)-cat.Slots:]
		}
		fill := false
		if len(list) == 0 {
			if cleared[cat.ID] && !cat.Required {
				continue
			}
			it := c.fallback(cat, typeID)
			if it == nil {
				continue
			}
			list, fill = []*avItem{it}, true
		}
		out = append(out, avResolved{cat: cat, items: list, fill: fill})
	}
	return out
}

func (c *avCatalog) look(a *avRow, owned func(*avItem) bool) avLook {
	out := avLook{Items: map[string][]string{}, Layers: []avLayer{}}
	for _, r := range c.resolve(a, owned) {
		for _, it := range r.items {
			out.Items[r.cat.Key] = append(out.Items[r.cat.Key], it.Key)
			out.Layers = append(out.Layers, avLayer{Category: r.cat.Key, Item: it.Key, ImageURL: it.ImageURL, Value: it.Value, Layer: r.cat.Layer})
		}
	}
	sort.SliceStable(out.Layers, func(i, j int) bool { return out.Layers[i].Layer < out.Layers[j].Layer })
	return out
}

func avatarLock(ctx context.Context, tx *sql.Tx, userID string) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('economy_avatar:' || $1))`, userID)
	return err
}

func cleanAvatarName(raw string, max int) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	if len([]rune(name)) > max {
		return "", badRequest("INVALID_NAME", "name", "at most "+itoa(max)+" letters")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", badRequest("INVALID_NAME", "name", "the name has a control character")
		}
	}
	return name, nil
}

type avCtx struct {
	s     *service
	st    avatarSettings
	cat   *avCatalog
	user  string
	langs *localized.Set
	lang  string
	own   avOwnership
	facts reqFacts
	names *reqNames
}

func (s *service) avatarContext(ctx context.Context, userID, lang string, facts bool) (*avCtx, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	st, err := loadAvatarSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	c, err := loadAvatarCatalog(ctx, s.db)
	if err != nil {
		return nil, err
	}
	x := &avCtx{s: s, st: st, cat: c, user: userID, langs: langs, lang: lang}
	if x.own, err = s.avatarOwnership(ctx, s.db, c, []string{userID}); err != nil {
		return nil, err
	}
	if facts {
		if x.facts, err = s.avatarFacts(ctx, c, userID); err != nil {
			return nil, err
		}
		x.names = newReqNames(s, langs, lang)
		if err := x.names.load(ctx); err != nil {
			return nil, err
		}
	}
	return x, nil
}

func (x *avCtx) owned(it *avItem) bool {
	return x.own.owns(x.user, it)
}

func (x *avCtx) usable(it *avItem) bool {
	ok, _ := avReqOK(it, x.facts)
	return x.st.usable(it, x.owned(it), ok)
}

func (x *avCtx) typeView(id string) any {
	t := x.cat.typeByID[id]
	if t == nil {
		return nil
	}
	return map[string]any{"key": t.Key, "name": x.langs.Resolve(t.Name, x.lang), "imageUrl": t.ImageURL, "iconUrl": t.IconURL}
}

func (x *avCtx) avatarView(a *avRow) map[string]any {
	return map[string]any{"id": a.ID, "name": a.Name, "type": x.typeView(a.TypeID), "isActive": a.Active, "createdAt": a.CreatedAt.UTC().Format(time.RFC3339), "look": x.cat.look(a, x.owned)}
}

func (s *service) myAvatars(ctx context.Context, x *avCtx) ([]*avRow, error) {
	list, err := loadAvatars(ctx, s.db, `a.user_id=$1::uuid`, x.user)
	if err != nil || len(list) > 0 || !x.st.AutoCreate {
		return list, err
	}
	if _, err := s.createAvatar(ctx, x, "", "", true); err != nil {
		if app, ok := asApp(err); ok && app.Code == "AVATAR_LIMIT" {
			return list, nil
		}
		return nil, err
	}
	return loadAvatars(ctx, s.db, `a.user_id=$1::uuid`, x.user)
}

func (s *service) avatarLimit(ctx context.Context, x *avCtx, count int) (map[string]any, error) {
	tags, err := s.userTagIDs(ctx, x.user)
	if err != nil {
		return nil, err
	}
	max := x.st.limitFor(tags)
	return map[string]any{"max": max, "count": count, "canCreate": count < max}, nil
}

func (s *service) avatarListAnswer(ctx context.Context, x *avCtx) (map[string]any, error) {
	list, err := s.myAvatars(ctx, x)
	if err != nil {
		return nil, err
	}
	views := make([]map[string]any, 0, len(list))
	active := ""
	for _, a := range list {
		if a.Active {
			active = a.ID
		}
		views = append(views, x.avatarView(a))
	}
	limit, err := s.avatarLimit(ctx, x, len(list))
	if err != nil {
		return nil, err
	}
	return map[string]any{"lang": x.lang, "enabled": x.st.Enabled, "activeId": active, "avatars": views, "limit": limit, "nameMaxLength": x.st.NameMaxLength}, nil
}

func (s *service) createAvatar(ctx context.Context, x *avCtx, name, typeKey string, activate bool) (string, error) {
	typeID := ""
	if typeKey = strings.TrimSpace(typeKey); typeKey != "" {
		t := x.cat.typeKey[typeKey]
		if t == nil {
			return "", notFound("AVATAR_TYPE_NOT_FOUND")
		}
		typeID = t.ID
	} else {
		typeID = x.cat.defaultType()
	}
	tags, err := s.userTagIDs(ctx, x.user)
	if err != nil {
		return "", err
	}
	max := x.st.limitFor(tags)
	id := uuid.Must(uuid.NewV4()).String()
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if err := avatarLock(ctx, tx, x.user); err != nil {
			return err
		}
		var count int
		var hasActive bool
		if err := tx.QueryRowContext(ctx, `SELECT count(*), COALESCE(bool_or(is_active), false) FROM economy_avatar WHERE user_id=$1`, x.user).Scan(&count, &hasActive); err != nil {
			return err
		}
		if count >= max {
			return conflict("AVATAR_LIMIT", "the player has as many avatars as allowed")
		}
		active := activate || !hasActive
		if active && hasActive {
			if _, err := tx.ExecContext(ctx, `UPDATE economy_avatar SET is_active=false, updated_at=now() WHERE user_id=$1 AND is_active`, x.user); err != nil {
				return err
			}
		}
		var tid any
		if typeID != "" {
			tid = typeID
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO economy_avatar (id, user_id, type_id, name, is_active) VALUES ($1,$2,$3,$4,$5)`, id, x.user, tid, name, active)
		return err
	})
	return id, err
}

func avatarForUser(ctx context.Context, q rowsQueryer, userID, id string) (*avRow, error) {
	if _, err := uuid.FromString(strings.TrimSpace(id)); err != nil {
		return nil, notFound("AVATAR_NOT_FOUND")
	}
	list, err := loadAvatars(ctx, q, `a.id=$1::uuid AND a.user_id=$2::uuid`, strings.TrimSpace(id), userID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, notFound("AVATAR_NOT_FOUND")
	}
	return list[0], nil
}

func (s *service) pickAvatar(ctx context.Context, x *avCtx, id string) (*avRow, error) {
	if strings.TrimSpace(id) != "" {
		return avatarForUser(ctx, s.db, x.user, id)
	}
	list, err := s.myAvatars(ctx, x)
	if err != nil {
		return nil, err
	}
	for _, a := range list {
		if a.Active {
			return a, nil
		}
	}
	if len(list) > 0 {
		return list[0], nil
	}
	return nil, notFound("AVATAR_NOT_FOUND")
}

func (s *service) avatarItemExtras(ctx context.Context, x *avCtx) (map[string]map[string]any, map[string]map[string]any, error) {
	products := map[string]map[string]any{}
	shop := map[string]map[string]any{}
	pids, keys := []string{}, []string{}
	for _, it := range x.cat.items {
		switch it.Access {
		case "product":
			pids = append(pids, it.ProductID)
		case "item":
			keys = append(keys, it.ItemKey)
		}
	}
	if len(pids) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text, icon_url, price_toman, COALESCE(coin_price::text,'null'), is_active FROM economy_product WHERE id::text = ANY($1::text[])`, pids)
		if err != nil {
			return nil, nil, err
		}
		for rows.Next() {
			var id, key, name, icon, coin string
			var toman sql.NullInt64
			var active bool
			if err := rows.Scan(&id, &key, &name, &icon, &toman, &coin, &active); err != nil {
				rows.Close()
				return nil, nil, err
			}
			v := map[string]any{"key": key, "name": x.langs.Resolve(parseI18n(name), x.lang), "iconUrl": icon, "coinPrice": json.RawMessage(coin), "isActive": active, "priceToman": nil}
			if toman.Valid {
				v["priceToman"] = toman.Int64
			}
			products[id] = v
		}
		rows.Close()
	}
	if len(keys) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT key, name::text, icon_url, COALESCE(price::text,'null') FROM economy_item WHERE key = ANY($1::text[])`, keys)
		if err != nil {
			return nil, nil, err
		}
		for rows.Next() {
			var key, name, icon, price string
			if err := rows.Scan(&key, &name, &icon, &price); err != nil {
				rows.Close()
				return nil, nil, err
			}
			shop[key] = map[string]any{"key": key, "name": x.langs.Resolve(parseI18n(name), x.lang), "iconUrl": icon, "price": json.RawMessage(price)}
		}
		rows.Close()
	}
	return products, shop, nil
}

func (x *avCtx) itemView(it *avItem, a *avRow, equipped map[string]bool, products, shop map[string]map[string]any) (map[string]any, bool) {
	owned := x.owned(it)
	ok, miss := avReqOK(it, x.facts)
	if !ok && (!owned || it.Access == "free") && it.Req != nil && it.Req.Lock == "hide" {
		return nil, false
	}
	types := []string{}
	for _, id := range it.TypeIDs {
		if t := x.cat.typeByID[id]; t != nil {
			types = append(types, t.Key)
		}
	}
	missing := []map[string]any{}
	for _, m := range miss {
		missing = append(missing, x.names.missView(m))
	}
	v := map[string]any{"key": it.Key, "category": it.Cat.Key, "name": x.langs.Resolve(it.Name, x.lang), "imageUrl": it.ImageURL, "iconUrl": it.IconURL, "value": it.Value,
		"access": it.Access, "isDefault": it.IsDefault, "types": types, "owned": owned, "locked": !ok, "missing": missing, "usable": x.st.usable(it, owned, ok),
		"price": nil, "product": nil, "shopItem": nil, "message": "", "fits": true, "equipped": false}
	if it.Req != nil && len(it.Req.Message) > 0 && !ok {
		v["message"] = x.langs.Resolve(i18n(it.Req.Message), x.lang)
	}
	switch it.Access {
	case "currency":
		v["price"] = it.Price
	case "product":
		if p, found := products[it.ProductID]; found {
			v["product"] = p
		}
	case "item":
		if p, found := shop[it.ItemKey]; found {
			v["shopItem"] = p
		} else {
			v["shopItem"] = map[string]any{"key": it.ItemKey}
		}
	}
	if a != nil {
		v["fits"] = it.fits(a.TypeID)
		v["equipped"] = equipped[it.ID]
	}
	return v, true
}

func (s *service) rpcAvatarCatalog(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang     string `json:"lang"`
		AvatarID string `json:"avatarId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	x, err := s.avatarContext(ctx, userID, in.Lang, true)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": x.lang, "enabled": x.st.Enabled, "avatarId": "", "types": []any{}, "categories": []any{}}
	if !x.st.Enabled {
		return out, nil
	}
	var a *avRow
	if a, err = s.pickAvatar(ctx, x, in.AvatarID); err != nil {
		if app, ok := asApp(err); !ok || app.Code != "AVATAR_NOT_FOUND" || in.AvatarID != "" {
			return nil, err
		}
		a = nil
	}
	equipped := map[string]bool{}
	if a != nil {
		out["avatarId"] = a.ID
		for _, r := range x.cat.resolve(a, x.owned) {
			for _, it := range r.items {
				equipped[it.ID] = true
			}
		}
	}
	products, shop, err := s.avatarItemExtras(ctx, x)
	if err != nil {
		return nil, err
	}
	types := make([]map[string]any, 0, len(x.cat.types))
	for _, t := range x.cat.types {
		types = append(types, map[string]any{"key": t.Key, "name": x.langs.Resolve(t.Name, x.lang), "imageUrl": t.ImageURL, "iconUrl": t.IconURL})
	}
	cats := make([]map[string]any, 0, len(x.cat.cats))
	for _, c := range x.cat.cats {
		items := make([]map[string]any, 0, len(c.Items))
		for _, it := range c.Items {
			if v, ok := x.itemView(it, a, equipped, products, shop); ok {
				items = append(items, v)
			}
		}
		cats = append(cats, map[string]any{"key": c.Key, "name": x.langs.Resolve(c.Name, x.lang), "iconUrl": c.IconURL, "layer": c.Layer, "required": c.Required, "slots": c.Slots, "items": items})
	}
	out["types"], out["categories"] = types, cats
	return out, nil
}

func (s *service) rpcAvatarList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarContext(ctx, userID, in.Lang, false)
	if err != nil {
		return nil, err
	}
	if !x.st.Enabled {
		return map[string]any{"lang": x.lang, "enabled": false, "activeId": "", "avatars": []any{}, "limit": map[string]any{"max": 0, "count": 0, "canCreate": false}, "nameMaxLength": x.st.NameMaxLength}, nil
	}
	return s.avatarListAnswer(ctx, x)
}

func (s *service) avatarEnabled(ctx context.Context, userID, lang string, facts bool) (*avCtx, error) {
	x, err := s.avatarContext(ctx, userID, lang, facts)
	if err != nil {
		return nil, err
	}
	if !x.st.Enabled {
		return nil, conflict("AVATAR_DISABLED", "the avatar designer is off")
	}
	return x, nil
}

func (s *service) oneAvatarAnswer(ctx context.Context, x *avCtx, id string) (any, error) {
	out, err := s.avatarListAnswer(ctx, x)
	if err != nil {
		return nil, err
	}
	for _, v := range out["avatars"].([]map[string]any) {
		if v["id"] == id {
			out["avatar"] = v
		}
	}
	return out, nil
}

func (s *service) rpcAvatarCreate(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Activate bool   `json:"activate"`
		Lang     string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarEnabled(ctx, userID, in.Lang, false)
	if err != nil {
		return nil, err
	}
	name, err := cleanAvatarName(in.Name, x.st.NameMaxLength)
	if err != nil {
		return nil, err
	}
	id, err := s.createAvatar(ctx, x, name, in.Type, in.Activate)
	if err != nil {
		return nil, err
	}
	activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "avatar_created", RefID: id, Actor: "player", Data: map[string]any{"name": name, "type": in.Type}}))
	return s.oneAvatarAnswer(ctx, x, id)
}

func (s *service) rpcAvatarRename(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarEnabled(ctx, userID, in.Lang, false)
	if err != nil {
		return nil, err
	}
	name, err := cleanAvatarName(in.Name, x.st.NameMaxLength)
	if err != nil {
		return nil, err
	}
	a, err := avatarForUser(ctx, s.db, userID, in.ID)
	if err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_avatar SET name=$3, updated_at=now() WHERE id=$1 AND user_id=$2`, a.ID, userID, name); err != nil {
		return nil, err
	}
	return s.oneAvatarAnswer(ctx, x, a.ID)
}

func (s *service) rpcAvatarDelete(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID   string `json:"id"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarEnabled(ctx, userID, in.Lang, false)
	if err != nil {
		return nil, err
	}
	a, err := avatarForUser(ctx, s.db, userID, in.ID)
	if err != nil {
		return nil, err
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if err := avatarLock(ctx, tx, userID); err != nil {
			return err
		}
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM economy_avatar WHERE user_id=$1`, userID).Scan(&count); err != nil {
			return err
		}
		if count <= 1 && x.st.KeepLast {
			return conflict("AVATAR_LAST", "the player must keep one avatar")
		}
		res, err := tx.ExecContext(ctx, `DELETE FROM economy_avatar WHERE id=$1 AND user_id=$2`, a.ID, userID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return notFound("AVATAR_NOT_FOUND")
		}
		if a.Active {
			_, err = tx.ExecContext(ctx, `UPDATE economy_avatar SET is_active=true, updated_at=now() WHERE id=(SELECT id FROM economy_avatar WHERE user_id=$1 ORDER BY created_at, id LIMIT 1)`, userID)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "avatar_deleted", RefID: a.ID, Actor: "player", Data: map[string]any{"name": a.Name}}))
	return s.avatarListAnswer(ctx, x)
}

func (s *service) rpcAvatarActivate(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID   string `json:"id"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarEnabled(ctx, userID, in.Lang, false)
	if err != nil {
		return nil, err
	}
	a, err := avatarForUser(ctx, s.db, userID, in.ID)
	if err != nil {
		return nil, err
	}
	if !a.Active {
		err = s.inTx(ctx, func(tx *sql.Tx) error {
			if err := avatarLock(ctx, tx, userID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE economy_avatar SET is_active=false, updated_at=now() WHERE user_id=$1 AND is_active`, userID); err != nil {
				return err
			}
			res, err := tx.ExecContext(ctx, `UPDATE economy_avatar SET is_active=true, updated_at=now() WHERE id=$1 AND user_id=$2`, a.ID, userID)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return notFound("AVATAR_NOT_FOUND")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return s.oneAvatarAnswer(ctx, x, a.ID)
}

func (s *service) checkEquip(x *avCtx, a *avRow, it *avItem) error {
	if !it.fits(a.TypeID) {
		return fail(http.StatusForbidden, "AVATAR_TYPE_NOT_ALLOWED", "", "this item does not fit this avatar type")
	}
	owned := x.owned(it)
	ok, _ := avReqOK(it, x.facts)
	if x.st.usable(it, owned, ok) {
		return nil
	}
	if !ok && (it.Access == "free" || !owned) {
		return fail(http.StatusForbidden, "AVATAR_ITEM_LOCKED", "", "the player does not meet this item's condition")
	}
	return fail(http.StatusForbidden, "AVATAR_ITEM_NOT_OWNED", "", "buy this item first")
}

func (s *service) rpcAvatarEquip(ctx context.Context, userID, payload string) (any, error) {
	return s.avatarWear(ctx, userID, payload, true)
}

func (s *service) rpcAvatarUnequip(ctx context.Context, userID, payload string) (any, error) {
	return s.avatarWear(ctx, userID, payload, false)
}

func (s *service) avatarWear(ctx context.Context, userID, payload string, on bool) (any, error) {
	var in struct {
		ID       string `json:"id"`
		Category string `json:"category"`
		Key      string `json:"key"`
		Lang     string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	x, err := s.avatarEnabled(ctx, userID, in.Lang, on)
	if err != nil {
		return nil, err
	}
	a, err := s.pickAvatar(ctx, x, in.ID)
	if err != nil {
		return nil, err
	}
	if x.cat.catKey[strings.TrimSpace(in.Category)] == nil {
		return nil, notFound("AVATAR_CATEGORY_NOT_FOUND")
	}
	it := x.cat.find(in.Category, in.Key)
	if it == nil {
		return nil, notFound("AVATAR_ITEM_NOT_FOUND")
	}
	if on {
		if err := s.checkEquip(x, a, it); err != nil {
			return nil, err
		}
		err = s.equipAvatarItem(ctx, x, a, it)
	} else {
		err = s.unequipAvatarItem(ctx, x, a, it)
	}
	if err != nil {
		return nil, err
	}
	return s.oneAvatarAnswer(ctx, x, a.ID)
}

func (s *service) equipAvatarItem(ctx context.Context, x *avCtx, a *avRow, it *avItem) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := avatarLock(ctx, tx, x.user); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT item_id::text FROM economy_avatar_equip WHERE avatar_id=$1 AND category_id=$2 ORDER BY equipped_at, item_id`, a.ID, it.Cat.ID)
		if err != nil {
			return err
		}
		current := []string{}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			if id == it.ID {
				rows.Close()
				return nil
			}
			current = append(current, id)
		}
		rows.Close()
		for len(current) >= it.Cat.Slots {
			if it.Cat.Slots > 1 && x.st.SlotFull == avatarSlotRefuse {
				return conflict("AVATAR_SLOT_FULL", "this category is full; take an item off first")
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM economy_avatar_equip WHERE avatar_id=$1 AND item_id=$2`, a.ID, current[0]); err != nil {
				return err
			}
			current = current[1:]
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_avatar_equip (avatar_id, item_id, category_id) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, a.ID, it.ID, it.Cat.ID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_avatar SET cleared=array_remove(cleared, $2), updated_at=now() WHERE id=$1`, a.ID, it.Cat.ID)
		return err
	})
}

func (s *service) unequipAvatarItem(ctx context.Context, x *avCtx, a *avRow, it *avItem) error {
	shown := false
	for _, r := range x.cat.resolve(a, x.owned) {
		if r.cat.ID != it.Cat.ID {
			continue
		}
		for _, v := range r.items {
			if v.ID == it.ID {
				shown = true
			}
		}
		if shown && r.cat.Required && len(r.items) <= 1 {
			return conflict("AVATAR_CATEGORY_REQUIRED", "this category always needs an item")
		}
	}
	stored := false
	for _, e := range a.Equips {
		if e.ItemID == it.ID {
			stored = true
		}
	}
	if !shown && !stored {
		return conflict("AVATAR_NOT_EQUIPPED", "this item is not on the avatar")
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := avatarLock(ctx, tx, x.user); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM economy_avatar_equip WHERE avatar_id=$1 AND item_id=$2`, a.ID, it.ID); err != nil {
			return err
		}
		var left int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM economy_avatar_equip WHERE avatar_id=$1 AND category_id=$2`, a.ID, it.Cat.ID).Scan(&left); err != nil {
			return err
		}
		if left == 0 && !it.Cat.Required {
			_, err := tx.ExecContext(ctx, `UPDATE economy_avatar SET cleared=array_append(array_remove(cleared, $2), $2), updated_at=now() WHERE id=$1`, a.ID, it.Cat.ID)
			return err
		}
		_, err := tx.ExecContext(ctx, `UPDATE economy_avatar SET updated_at=now() WHERE id=$1`, a.ID)
		return err
	})
}

func (s *service) rpcAvatarBuy(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Category  string `json:"category"`
		Key       string `json:"key"`
		RequestID string `json:"requestId"`
		Lang      string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	requestID, generated, err := cleanRequestID(in.RequestID, false)
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	x, err := s.avatarEnabled(ctx, userID, in.Lang, true)
	if err != nil {
		return nil, err
	}
	it := x.cat.find(in.Category, in.Key)
	if it == nil {
		return nil, notFound("AVATAR_ITEM_NOT_FOUND")
	}
	switch it.Access {
	case "free":
		return nil, conflict("AVATAR_ITEM_FREE", "this item is free")
	case "product":
		if !x.owned(it) {
			return nil, conflict("AVATAR_PRODUCT_REQUIRED", "this item comes with a shop product; buy the product")
		}
		return nil, conflict("AVATAR_ALREADY_OWNED", "the player already owns this item")
	}
	if ok, _ := avReqOK(it, x.facts); !ok {
		return nil, fail(http.StatusForbidden, "AVATAR_ITEM_LOCKED", "", "the player does not meet this item's condition")
	}
	paid := []coinPrice{}
	replayed := false
	if it.Access == "item" {
		if x.owned(it) {
			return nil, conflict("AVATAR_ALREADY_OWNED", "the player already owns this item")
		}
		res, err := s.buyItem(ctx, userID, it.ItemKey, in.RequestID)
		if err != nil {
			return nil, err
		}
		paid, replayed = res.Paid, res.Replayed
	} else {
		price := *it.Price
		fresh := false
		err = s.inTx(ctx, func(tx *sql.Tx) error {
			if err := avatarLock(ctx, tx, userID); err != nil {
				return err
			}
			var prev string
			err := tx.QueryRowContext(ctx, `SELECT request_id FROM economy_avatar_unlock WHERE user_id=$1 AND item_id=$2`, userID, it.ID).Scan(&prev)
			if err == nil {
				if !generated && prev == requestID {
					replayed = true
					paid = []coinPrice{price}
					return nil
				}
				return conflict("AVATAR_ALREADY_OWNED", "the player already owns this item")
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			grantID := requestGrantID(userID, "avatar:"+it.ID+":"+requestID)
			if price.Amount > 0 {
				if err := s.walletChange(ctx, tx, userID, grantID, map[string]int64{price.CurrencyKey: -price.Amount}, map[string]any{"reason": "avatar_buy", "sourceType": "avatar_item", "sourceId": it.Cat.Key + "/" + it.Key}); err != nil {
					return err
				}
			}
			raw, _ := json.Marshal(price)
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_avatar_unlock (user_id, item_id, source, paid, request_id) VALUES ($1,$2,'buy',$3::jsonb,$4)`, userID, it.ID, string(raw), requestID); err != nil {
				return err
			}
			paid, fresh = []coinPrice{price}, true
			return nil
		})
		if err != nil {
			return nil, err
		}
		if fresh {
			s.recordServerEvent(ctx, userID, "avatar_item_bought", float64(price.Amount), map[string]any{"category": it.Cat.Key, "item": it.Key, "currency": price.CurrencyKey})
			activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "avatar_bought", RefID: it.Cat.Key + "/" + it.Key, Actor: "player", Data: map[string]any{"category": it.Cat.Key, "item": it.Key, "price": price.Amount, "currency": price.CurrencyKey}}))
		}
	}
	x, err = s.avatarContext(ctx, userID, in.Lang, true)
	if err != nil {
		return nil, err
	}
	it = x.cat.find(in.Category, in.Key)
	products, shop, err := s.avatarItemExtras(ctx, x)
	if err != nil {
		return nil, err
	}
	view := map[string]any{}
	if it != nil {
		view, _ = x.itemView(it, nil, nil, products, shop)
	}
	wallet, err := s.userWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"lang": x.lang, "item": view, "paid": paid, "wallet": wallet, "replayed": replayed}, nil
}

func cleanUserIDs(raw []string) ([]string, error) {
	if len(raw) == 0 || len(raw) > avatarMaxIDs {
		return nil, badRequest("INVALID_USER_IDS", "userIds", "1 to 100 user ids")
	}
	ids := []string{}
	seen := map[string]bool{}
	for _, id := range raw {
		id = strings.TrimSpace(id)
		if id == "" || len(id) > 128 || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *service) activeLooks(ctx context.Context, c *avCatalog, ids []string) (map[string]*avRow, avOwnership, error) {
	valid := []string{}
	for _, id := range ids {
		if _, err := uuid.FromString(id); err == nil {
			valid = append(valid, id)
		}
	}
	rows := map[string]*avRow{}
	if len(valid) == 0 {
		return rows, avOwnership{}, nil
	}
	list, err := loadAvatars(ctx, s.db, `a.is_active AND a.user_id::text = ANY($1::text[])`, valid)
	if err != nil {
		return nil, avOwnership{}, err
	}
	owners := []string{}
	for _, a := range list {
		rows[a.UserID] = a
		owners = append(owners, a.UserID)
	}
	own, err := s.avatarOwnership(ctx, s.db, c, owners)
	return rows, own, err
}

func (s *service) rpcAvatarLooks(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserIDs []string `json:"userIds"`
		Lang    string   `json:"lang"`
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
	looks := map[string]any{}
	if !st.Enabled {
		for _, id := range ids {
			looks[id] = nil
		}
		return map[string]any{"enabled": false, "avatars": looks}, nil
	}
	c, err := loadAvatarCatalog(ctx, s.db)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, own, err := s.activeLooks(ctx, c, ids)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		a := rows[id]
		owned := func(it *avItem) bool { return own.owns(id, it) }
		v := map[string]any{"id": "", "name": "", "type": nil, "look": c.look(a, owned)}
		if a != nil {
			v["id"], v["name"] = a.ID, a.Name
			if t := c.typeByID[a.TypeID]; t != nil {
				v["type"] = map[string]any{"key": t.Key, "name": langs.Resolve(t.Name, lang), "imageUrl": t.ImageURL, "iconUrl": t.IconURL}
			}
		}
		looks[id] = v
	}
	return map[string]any{"enabled": true, "avatars": looks}, nil
}

func (s *service) avatarRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_avatar_catalog":    s.rpcAvatarCatalog,
		"economy_avatar_buy":        s.rpcAvatarBuy,
		"economy_avatar_list":       s.rpcAvatarList,
		"economy_avatar_create":     s.rpcAvatarCreate,
		"economy_avatar_rename":     s.rpcAvatarRename,
		"economy_avatar_delete":     s.rpcAvatarDelete,
		"economy_avatar_equip":      s.rpcAvatarEquip,
		"economy_avatar_unequip":    s.rpcAvatarUnequip,
		"economy_avatar_set_active": s.rpcAvatarActivate,
		"economy_avatar_looks":      s.rpcAvatarLooks,
	}
}
