package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	codeRequirementsNotMet = "REQUIREMENTS_NOT_MET"
	maxReqItems            = 20
	maxReqTags             = 20
	maxReqMetas            = 10
	maxReqProducts         = 10
	maxReqConsume          = 1000000000
)

var reqDefaultMessage = i18n{
	"fa": "برای ورود به این اتاق شرط‌هایش را نداری.",
	"en": "You do not meet the requirements to enter this room.",
	"ar": "لا تستوفي شروط الدخول إلى هذه الغرفة.",
}

// ---- admin input ----

type gameReqInput struct {
	ItemIDs       []string        `json:"itemIds"`
	ItemsMode     string          `json:"itemsMode"`
	ConsumeItems  bool            `json:"consumeItems"`
	TagIDs        []string        `json:"tagIds"`
	TagsMode      string          `json:"tagsMode"`
	ExcludeTagIDs []string        `json:"excludeTagIds"`
	MinLevel      json.RawMessage `json:"minLevel"`
	Metas         []struct {
		MetaID   string          `json:"metaId"`
		Operator string          `json:"operator"`
		Value    json.RawMessage `json:"value"`
		Consume  json.RawMessage `json:"consume"`
	} `json:"metas"`
	ProductIDs []string `json:"productIds"`
	Paths      []string `json:"paths"`
	Refund     string   `json:"refund"`
	Lock       string   `json:"lock"`
	Inherit    *bool    `json:"inherit"`
	Message    i18n     `json:"message"`
}

func reqIDs(ctx context.Context, tx *sql.Tx, field, table, code string, ids []string, max int) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, raw := range ids {
		ref, err := refExists(ctx, tx, table, field, code, raw, true)
		if err != nil {
			return nil, err
		}
		if seen[*ref] {
			continue
		}
		seen[*ref] = true
		out = append(out, *ref)
	}
	if len(out) > max {
		return nil, badRequest("TOO_MANY_REQUIREMENTS", field, "at most "+strconv.Itoa(max))
	}
	return out, nil
}

func reqUserTags(ctx context.Context, tx *sql.Tx, field string, ids []string) ([]string, error) {
	out, err := checkUserTags(ctx, tx, ids)
	if app, ok := asApp(err); ok {
		app.Field = field
	}
	return out, err
}

// cleanGameRequirements validates an admin requirement set; nil input or an empty set gives nil.
func cleanGameRequirements(ctx context.Context, tx *sql.Tx, langs *localized.Set, field string, in *gameReqInput, room bool) (*gamecatalog.Requirements, error) {
	if in == nil {
		return nil, nil
	}
	pre := field + "."
	r := &gamecatalog.Requirements{ItemsMode: strings.TrimSpace(in.ItemsMode), TagsMode: strings.TrimSpace(in.TagsMode), Refund: strings.TrimSpace(in.Refund), Lock: strings.TrimSpace(in.Lock)}
	var err error
	if r.ItemIDs, err = reqIDs(ctx, tx, pre+"itemIds", "economy_item", "ITEM_NOT_FOUND", in.ItemIDs, maxReqItems); err != nil {
		return nil, err
	}
	if r.TagIDs, err = reqUserTags(ctx, tx, pre+"tagIds", in.TagIDs); err != nil {
		return nil, err
	}
	if r.ExcludeTagIDs, err = reqUserTags(ctx, tx, pre+"excludeTagIds", in.ExcludeTagIDs); err != nil {
		return nil, err
	}
	for _, id := range r.ExcludeTagIDs {
		for _, need := range r.TagIDs {
			if id == need {
				return nil, badRequest("TAG_BOTH_WAYS", pre+"excludeTagIds", "a tag cannot be both required and excluded")
			}
		}
	}
	if r.ProductIDs, err = reqIDs(ctx, tx, pre+"productIds", "economy_product", "PRODUCT_NOT_FOUND", in.ProductIDs, maxReqProducts); err != nil {
		return nil, err
	}
	for _, f := range []struct {
		name string
		v    *string
		opts []string
	}{{"itemsMode", &r.ItemsMode, []string{"any", "all"}}, {"tagsMode", &r.TagsMode, []string{"any", "all"}}, {"refund", &r.Refund, []string{"return", "keep"}}, {"lock", &r.Lock, []string{"show", "hide"}}} {
		if *f.v == "" {
			*f.v = f.opts[0]
		}
		if err := oneOf(pre+f.name, "INVALID_REQUIREMENT", *f.v, f.opts...); err != nil {
			return nil, err
		}
	}
	lvl, err := optionalInt(in.MinLevel, pre+"minLevel", "INVALID_MIN_LEVEL", "the level is a whole number from 0 to 1000", 0, 1000)
	if err != nil {
		return nil, err
	}
	if lvl != nil {
		r.MinLevel = *lvl
	}
	if len(in.Metas) > maxReqMetas {
		return nil, badRequest("TOO_MANY_REQUIREMENTS", pre+"metas", "at most "+strconv.Itoa(maxReqMetas))
	}
	for i, m := range in.Metas {
		mf := pre + "metas[" + strconv.Itoa(i) + "]"
		ref, err := refExists(ctx, tx, "economy_meta", mf+".metaId", "META_NOT_FOUND", m.MetaID, true)
		if err != nil {
			return nil, err
		}
		op := strings.TrimSpace(m.Operator)
		if op == "" {
			op = "gte"
		}
		if err := oneOf(mf+".operator", "INVALID_OPERATOR", op, "gt", "gte", "lt", "lte", "eq"); err != nil {
			return nil, err
		}
		v, err := optionalNumeric(mf+".value", m.Value)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, badRequest("INVALID_VALUE", mf+".value", "a value is required")
		}
		cond := gamecatalog.MetaCondition{MetaID: *ref, Operator: op, Value: numericFloat(*v)}
		c, err := optionalNumeric(mf+".consume", m.Consume)
		if err != nil {
			return nil, err
		}
		if c != nil {
			cond.Consume = numericFloat(*c)
			if cond.Consume < 0 || cond.Consume > maxReqConsume {
				return nil, badRequest("INVALID_CONSUME", mf+".consume", "between 0 and 1000000000")
			}
		}
		r.Metas = append(r.Metas, cond)
	}
	r.ConsumeItems = in.ConsumeItems && len(r.ItemIDs) > 0
	if r.ConsumeItems {
		var starter string
		err := tx.QueryRowContext(ctx, `SELECT key FROM economy_item WHERE id::text = ANY($1::text[]) AND is_starter ORDER BY key LIMIT 1`, r.ItemIDs).Scan(&starter)
		if err == nil {
			return nil, badRequest("STARTER_TICKET", pre+"consumeItems", "starter item "+starter+" belongs to every player and cannot be used up as a ticket")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	seen := map[string]bool{}
	for _, p := range in.Paths {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		if err := oneOf(pre+"paths", "INVALID_PATH", p, gamecatalog.Paths...); err != nil {
			return nil, err
		}
		seen[p] = true
	}
	if len(seen) < len(gamecatalog.Paths) {
		for _, p := range gamecatalog.Paths {
			if seen[p] {
				r.Paths = append(r.Paths, p)
			}
		}
		if len(r.Paths) == 0 {
			r.Paths = nil
		}
	}
	if room && in.Inherit != nil && !*in.Inherit {
		off := false
		r.Inherit = &off
	}
	msg, err := cleanI18n(langs, pre+"message", in.Message, false, 300)
	if err != nil {
		return nil, err
	}
	if len(msg) > 0 {
		r.Message = map[string]string(msg)
	}
	if r.Empty() && r.Inherit == nil {
		return nil, nil
	}
	return r, nil
}

// ---- evaluation ----

type reqFacts struct {
	Level int
	Items map[string]bool
	Tags  map[string]bool
	Metas map[string]float64
}

type reqMiss struct {
	Type     string   `json:"type"`
	Mode     string   `json:"mode,omitempty"`
	IDs      []string `json:"ids,omitempty"`
	MetaID   string   `json:"metaId,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Value    float64  `json:"value,omitempty"`
	Have     float64  `json:"have,omitempty"`
	Consume  float64  `json:"consume,omitempty"`
	MinLevel int      `json:"minLevel,omitempty"`
	Level    int      `json:"level,omitempty"`
}

func metaCondHolds(c gamecatalog.MetaCondition, have float64) bool {
	if !relationHolds(c.Operator, have, c.Value) {
		return false
	}
	return c.Consume <= 0 || round4(have) >= round4(c.Consume)
}

// reqMissing is the pure check of one set against a player's facts.
func reqMissing(r *gamecatalog.Requirements, f reqFacts) []reqMiss {
	out := []reqMiss{}
	if r.Empty() {
		return out
	}
	if r.MinLevel > 0 && f.Level < r.MinLevel {
		out = append(out, reqMiss{Type: "level", MinLevel: r.MinLevel, Level: f.Level})
	}
	if len(r.ItemIDs) > 0 {
		lacking := []string{}
		for _, id := range r.ItemIDs {
			if !f.Items[id] {
				lacking = append(lacking, id)
			}
		}
		if (r.ItemsMode == "all" && len(lacking) > 0) || (r.ItemsMode != "all" && len(lacking) == len(r.ItemIDs)) {
			out = append(out, reqMiss{Type: "item", Mode: modeOr(r.ItemsMode), IDs: lacking})
		}
	}
	if len(r.TagIDs) > 0 {
		lacking := []string{}
		for _, id := range r.TagIDs {
			if !f.Tags[id] {
				lacking = append(lacking, id)
			}
		}
		if (r.TagsMode == "all" && len(lacking) > 0) || (r.TagsMode != "all" && len(lacking) == len(r.TagIDs)) {
			out = append(out, reqMiss{Type: "tag", Mode: modeOr(r.TagsMode), IDs: lacking})
		}
	}
	hit := []string{}
	for _, id := range r.ExcludeTagIDs {
		if f.Tags[id] {
			hit = append(hit, id)
		}
	}
	if len(hit) > 0 {
		out = append(out, reqMiss{Type: "excludedTag", IDs: hit})
	}
	for _, c := range r.Metas {
		have := f.Metas[c.MetaID]
		if !metaCondHolds(c, have) {
			out = append(out, reqMiss{Type: "meta", MetaID: c.MetaID, Operator: c.Operator, Value: c.Value, Have: round4(have), Consume: c.Consume})
		}
	}
	return out
}

func modeOr(m string) string {
	if m == "" {
		return "any"
	}
	return m
}

// reqSets are the sets that apply to a room on a path: the game's (unless the room opts out) and the room's.
func reqSets(game *gamecatalog.Requirements, room *gamecatalog.Room, path string) []*gamecatalog.Requirements {
	out := []*gamecatalog.Requirements{}
	var own *gamecatalog.Requirements
	if room != nil {
		own = room.Requirements
	}
	if own.Inherits() && game.AppliesTo(path) {
		out = append(out, game)
	}
	if own.AppliesTo(path) {
		out = append(out, own)
	}
	return out
}

func reqHidden(sets []*gamecatalog.Requirements) bool {
	for _, r := range sets {
		if r.Lock == "hide" {
			return true
		}
	}
	return false
}

// ---- facts ----

func reqNeeds(games []*gameRecord) (found, items bool, metaIDs []string) {
	seen := map[string]bool{}
	add := func(r *gamecatalog.Requirements) {
		if r.Empty() {
			return
		}
		found = true
		items = items || len(r.ItemIDs) > 0
		for _, m := range r.Metas {
			if !seen[m.MetaID] {
				seen[m.MetaID] = true
				metaIDs = append(metaIDs, m.MetaID)
			}
		}
	}
	for _, g := range games {
		add(g.Requirements)
		for i := range g.Rooms {
			add(g.Rooms[i].Requirements)
		}
	}
	return found, items, metaIDs
}

func (s *service) loadReqFacts(ctx context.Context, q rowsQueryer, userID string, level int, items bool, metaIDs []string) (reqFacts, error) {
	f := reqFacts{Level: level, Items: map[string]bool{}, Tags: map[string]bool{}, Metas: map[string]float64{}}
	if items {
		if err := s.ensureStarters(ctx, userID); err != nil {
			return f, err
		}
		rows, err := q.QueryContext(ctx, `SELECT item_id::text FROM economy_user_item WHERE user_id=$1`, userID)
		if err != nil {
			return f, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return f, err
			}
			f.Items[id] = true
		}
		rows.Close()
	}
	rows, err := q.QueryContext(ctx, `SELECT tag_id::text FROM base_user_tag WHERE user_id=$1`, userID)
	if err != nil {
		return f, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return f, err
		}
		f.Tags[id] = true
	}
	rows.Close()
	if len(metaIDs) > 0 {
		rows, err := q.QueryContext(ctx, `SELECT m.id::text, COALESCE(um.value, m.default_value)::float8 FROM economy_meta m
			LEFT JOIN economy_user_meta um ON um.meta_id=m.id AND um.user_id=$1 WHERE m.id::text = ANY($2::text[])`, userID, metaIDs)
		if err != nil {
			return f, err
		}
		for rows.Next() {
			var id string
			var v float64
			if err := rows.Scan(&id, &v); err != nil {
				rows.Close()
				return f, err
			}
			f.Metas[id] = v
		}
		rows.Close()
	}
	return f, nil
}

// ---- names, views and products ----

type reqNames struct {
	s      *service
	langs  *localized.Set
	lang   string
	items  map[string]map[string]any
	tags   map[string]map[string]any
	metas  map[string]map[string]any
	loaded bool
}

func newReqNames(s *service, langs *localized.Set, lang string) *reqNames {
	return &reqNames{s: s, langs: langs, lang: lang, items: map[string]map[string]any{}, tags: map[string]map[string]any{}, metas: map[string]map[string]any{}}
}

func (n *reqNames) resolve(t i18n, key string) string {
	if v := n.langs.Resolve(t, n.lang); v != "" {
		return v
	}
	return key
}

func (n *reqNames) load(ctx context.Context) error {
	if n.loaded {
		return nil
	}
	n.loaded = true
	rows, err := n.s.db.QueryContext(ctx, `SELECT id::text, key, name::text, icon_url FROM economy_item`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, key, name, icon string
		if err := rows.Scan(&id, &key, &name, &icon); err != nil {
			rows.Close()
			return err
		}
		n.items[id] = map[string]any{"id": id, "key": key, "name": n.resolve(parseI18n(name), key), "iconUrl": icon}
	}
	rows.Close()
	rows, err = n.s.db.QueryContext(ctx, `SELECT id::text, key, name::text, is_public FROM base_tag WHERE key_type='user'`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, key, name string
		var public bool
		if err := rows.Scan(&id, &key, &name, &public); err != nil {
			rows.Close()
			return err
		}
		n.tags[id] = map[string]any{"key": key, "name": n.resolve(parseI18n(name), key), "public": public}
	}
	rows.Close()
	rows, err = n.s.db.QueryContext(ctx, `SELECT id::text, key, name::text FROM economy_meta`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, key, name string
		if err := rows.Scan(&id, &key, &name); err != nil {
			rows.Close()
			return err
		}
		n.metas[id] = map[string]any{"key": key, "name": n.resolve(parseI18n(name), key)}
	}
	rows.Close()
	return nil
}

func (n *reqNames) item(id string, f *reqFacts) map[string]any {
	v, ok := n.items[id]
	if !ok {
		v = map[string]any{"id": id, "key": "", "name": "", "iconUrl": ""}
	}
	out := map[string]any{"key": v["key"], "name": v["name"], "iconUrl": v["iconUrl"]}
	if f != nil {
		out["owned"] = f.Items[id]
	}
	return out
}

func (n *reqNames) tag(id string, f *reqFacts) map[string]any {
	v, ok := n.tags[id]
	if !ok {
		v = map[string]any{"key": "", "name": ""}
	}
	out := map[string]any{"key": v["key"], "name": v["name"]}
	if f != nil {
		out["has"] = f.Tags[id]
	}
	return out
}

func (n *reqNames) metaKey(id string) string {
	if v, ok := n.metas[id]; ok {
		return v["key"].(string)
	}
	return ""
}

// view is the public description of one set (what the client shows on a room card).
func (n *reqNames) view(r *gamecatalog.Requirements, f *reqFacts) map[string]any {
	out := map[string]any{"minLevel": r.MinLevel, "items": nil, "tags": nil, "metas": []map[string]any{}, "excludesTags": len(r.ExcludeTagIDs) > 0,
		"paths": r.Paths, "consumes": r.Consumes(), "refund": modeDefault(r.Refund, "return")}
	if r.Paths == nil {
		out["paths"] = gamecatalog.Paths
	}
	if len(r.ItemIDs) > 0 {
		list := []map[string]any{}
		for _, id := range r.ItemIDs {
			list = append(list, n.item(id, f))
		}
		out["items"] = map[string]any{"mode": modeOr(r.ItemsMode), "consume": r.ConsumeItems, "list": list}
	}
	if len(r.TagIDs) > 0 {
		list := []map[string]any{}
		for _, id := range r.TagIDs {
			list = append(list, n.tag(id, f))
		}
		out["tags"] = map[string]any{"mode": modeOr(r.TagsMode), "list": list}
	}
	metas := []map[string]any{}
	for _, c := range r.Metas {
		v := n.metas[c.MetaID]
		m := map[string]any{"key": "", "name": "", "operator": c.Operator, "value": c.Value, "consume": c.Consume}
		if v != nil {
			m["key"], m["name"] = v["key"], v["name"]
		}
		if f != nil {
			m["have"] = round4(f.Metas[c.MetaID])
		}
		metas = append(metas, m)
	}
	out["metas"] = metas
	return out
}

func modeDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func (n *reqNames) missView(m reqMiss) map[string]any {
	out := map[string]any{"type": m.Type}
	switch m.Type {
	case "level":
		out["minLevel"], out["level"] = m.MinLevel, m.Level
	case "item":
		list := []map[string]any{}
		for _, id := range m.IDs {
			list = append(list, n.item(id, nil))
		}
		out["mode"], out["items"] = m.Mode, list
	case "tag":
		list := []map[string]any{}
		for _, id := range m.IDs {
			list = append(list, n.tag(id, nil))
		}
		out["mode"], out["tags"] = m.Mode, list
	case "excludedTag":
		list := []map[string]any{}
		for _, id := range m.IDs {
			if v, ok := n.tags[id]; ok && v["public"] == true {
				list = append(list, n.tag(id, nil))
			}
		}
		out["tags"] = list
	case "meta":
		v := n.metas[m.MetaID]
		out["key"], out["name"] = "", ""
		if v != nil {
			out["key"], out["name"] = v["key"], v["name"]
		}
		out["operator"], out["value"], out["have"], out["consume"] = m.Operator, m.Value, m.Have, m.Consume
	}
	return out
}

// products lists active products that satisfy the missing items or metas: the sets' own product
// list when the admin gave one, otherwise products whose prize gives a missing item or adds a missing meta.
func (n *reqNames) products(ctx context.Context, sets []*gamecatalog.Requirements, miss []reqMiss) ([]map[string]any, error) {
	fixed := []string{}
	for _, r := range sets {
		fixed = append(fixed, r.ProductIDs...)
	}
	itemKeys, metaKeys := []string{}, []string{}
	for _, m := range miss {
		switch m.Type {
		case "item":
			for _, id := range m.IDs {
				if k, _ := n.item(id, nil)["key"].(string); k != "" {
					itemKeys = append(itemKeys, k)
				}
			}
		case "meta":
			if k := n.metaKey(m.MetaID); k != "" && (m.Operator == "gt" || m.Operator == "gte" || m.Operator == "eq" || m.Consume > 0) {
				metaKeys = append(metaKeys, k)
			}
		}
	}
	out := []map[string]any{}
	if len(fixed) == 0 && len(itemKeys) == 0 && len(metaKeys) == 0 {
		return out, nil
	}
	rows, err := n.s.db.QueryContext(ctx, `SELECT p.id::text, p.key, p.name::text, p.icon_url, p.price_toman, COALESCE(p.coin_price::text,''),
			COALESCE((SELECT jsonb_agg(DISTINCT (it->>'type') || ':' || (it->>'key')) FROM jsonb_array_elements(z.items) it
				WHERE (it->>'type'='item' AND it->>'key' = ANY($2::text[])) OR (it->>'type'='meta' AND it->>'key' = ANY($3::text[]) AND COALESCE(it->>'action','add')='add')),'[]'::jsonb)::text
		FROM economy_product p JOIN economy_prize z ON z.id=p.prize_id
		WHERE p.is_active AND (CASE WHEN cardinality($1::text[]) > 0 THEN p.id::text = ANY($1::text[]) ELSE EXISTS (SELECT 1 FROM jsonb_array_elements(z.items) it
				WHERE (it->>'type'='item' AND it->>'key' = ANY($2::text[])) OR (it->>'type'='meta' AND it->>'key' = ANY($3::text[]) AND COALESCE(it->>'action','add')='add')) END)
		ORDER BY p.sort, p.key LIMIT 10`, fixed, itemKeys, metaKeys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, key, name, icon, coin, sat string
		var toman sql.NullInt64
		if err := rows.Scan(&id, &key, &name, &icon, &toman, &coin, &sat); err != nil {
			return nil, err
		}
		p := map[string]any{"id": id, "key": key, "name": n.resolve(parseI18n(name), key), "iconUrl": icon, "priceToman": nil, "coinPrice": nil, "satisfies": json.RawMessage(sat)}
		if toman.Valid {
			p["priceToman"] = toman.Int64
		}
		if coin != "" && coin != "null" {
			p["coinPrice"] = json.RawMessage(coin)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ---- checks ----

type reqCheck struct {
	Sets    []*gamecatalog.Requirements
	Missing []reqMiss
	Facts   reqFacts
}

func (s *service) checkRoomRequirements(ctx context.Context, userID string, g *gameRecord, room *gamecatalog.Room, path string, level int) (reqCheck, error) {
	out := reqCheck{Sets: reqSets(g.Requirements, room, path)}
	if len(out.Sets) == 0 {
		return out, nil
	}
	items := false
	metaIDs := []string{}
	for _, r := range out.Sets {
		items = items || len(r.ItemIDs) > 0
		for _, m := range r.Metas {
			metaIDs = append(metaIDs, m.MetaID)
		}
	}
	f, err := s.loadReqFacts(ctx, s.db, userID, level, items, metaIDs)
	if err != nil {
		return out, err
	}
	out.Facts = f
	for _, r := range out.Sets {
		out.Missing = append(out.Missing, reqMissing(r, f)...)
	}
	return out, nil
}

func (s *service) requirementsError(ctx context.Context, userID, lang string, g *gameRecord, room *gamecatalog.Room, path string, rc reqCheck) error {
	langs, l, err := s.gameLangs(ctx, userID, lang)
	if err != nil {
		return err
	}
	names := newReqNames(s, langs, l)
	if err := names.load(ctx); err != nil {
		return err
	}
	missing := []map[string]any{}
	for _, m := range rc.Missing {
		missing = append(missing, names.missView(m))
	}
	products, err := names.products(ctx, rc.Sets, rc.Missing)
	if err != nil {
		return err
	}
	msg := ""
	for i := len(rc.Sets) - 1; i >= 0 && msg == ""; i-- {
		if len(rc.Sets[i].Message) > 0 {
			msg = langs.Resolve(i18n(rc.Sets[i].Message), l)
		}
	}
	if msg == "" {
		msg = langs.Resolve(reqDefaultMessage, l)
	}
	if msg == "" {
		msg = reqDefaultMessage["en"]
	}
	e := fail(http.StatusForbidden, codeRequirementsNotMet, "room", "the player does not meet this room's entry requirements")
	roomKey := ""
	if room != nil {
		roomKey = room.Key
	}
	e.Meta = map[string]any{"game": g.Key, "room": roomKey, "path": pathOr(path), "missing": missing, "products": products, "message": msg}
	return e
}

func pathOr(p string) string {
	if p == "" {
		return gamecatalog.PathMatchmaking
	}
	return p
}

func cleanPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if !gamecatalog.ValidPath(p) {
		return "", badRequest("INVALID_PATH", "path", "matchmaking, bot or friend")
	}
	return pathOr(p), nil
}

// ---- tickets ----

type gameTicketItem struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Source       string `json:"source"`
	SourceRef    string `json:"sourceRef"`
	UpgradeLevel int    `json:"upgradeLevel"`
}

type gameTicketMeta struct {
	ID     string  `json:"id"`
	Key    string  `json:"key"`
	Amount float64 `json:"amount"`
}

type gameTickets struct {
	Items    []gameTicketItem `json:"items"`
	Metas    []gameTicketMeta `json:"metas"`
	Refund   string           `json:"refund"`
	Returned bool             `json:"returned,omitempty"`
}

func (t *gameTickets) empty() bool {
	return t == nil || (len(t.Items) == 0 && len(t.Metas) == 0)
}

func (t *gameTickets) public() map[string]any {
	items := []string{}
	for _, it := range t.Items {
		items = append(items, it.Key)
	}
	metas := map[string]float64{}
	for _, m := range t.Metas {
		metas[m.Key] += m.Amount
	}
	return map[string]any{"items": items, "metas": metas, "refund": t.Refund}
}

// consumeTickets takes the ticket items and metas of the sets inside the entry transaction; it
// re-checks ownership under row locks, so a parallel entry cannot spend the same ticket twice.
func consumeTickets(ctx context.Context, tx *sql.Tx, userID string, sets []*gamecatalog.Requirements) (*gameTickets, error) {
	t := &gameTickets{Items: []gameTicketItem{}, Metas: []gameTicketMeta{}, Refund: "return"}
	for _, r := range sets {
		if !r.Consumes() {
			continue
		}
		if r.Refund == "keep" {
			t.Refund = "keep"
		}
		if r.ConsumeItems && len(r.ItemIDs) > 0 {
			taken := 0
			for _, id := range r.ItemIDs {
				var it gameTicketItem
				err := tx.QueryRowContext(ctx, `DELETE FROM economy_user_item u USING economy_item i WHERE u.user_id=$1 AND u.item_id=$2 AND i.id=u.item_id
					RETURNING u.item_id::text, i.key, u.source, u.source_ref, u.upgrade_level`, userID, id).Scan(&it.ID, &it.Key, &it.Source, &it.SourceRef, &it.UpgradeLevel)
				if errors.Is(err, sql.ErrNoRows) {
					if r.ItemsMode == "all" {
						return nil, fail(http.StatusForbidden, codeRequirementsNotMet, "room", "a ticket item was used up meanwhile")
					}
					continue
				}
				if err != nil {
					return nil, err
				}
				t.Items = append(t.Items, it)
				taken++
				if r.ItemsMode != "all" {
					break
				}
			}
			if r.ItemsMode != "all" && taken == 0 {
				return nil, fail(http.StatusForbidden, codeRequirementsNotMet, "room", "a ticket item was used up meanwhile")
			}
		}
		for _, c := range r.Metas {
			if c.Consume <= 0 {
				continue
			}
			var key string
			var have float64
			err := tx.QueryRowContext(ctx, `SELECT key FROM economy_meta WHERE id=$1`, c.MetaID).Scan(&key)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fail(http.StatusForbidden, codeRequirementsNotMet, "room", "a ticket meta no longer exists")
			}
			if err != nil {
				return nil, err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) SELECT $1, id, default_value FROM economy_meta WHERE id=$2 ON CONFLICT DO NOTHING`, userID, c.MetaID); err != nil {
				return nil, err
			}
			if err := tx.QueryRowContext(ctx, `SELECT value::float8 FROM economy_user_meta WHERE user_id=$1 AND meta_id=$2 FOR UPDATE`, userID, c.MetaID).Scan(&have); err != nil {
				return nil, err
			}
			if !metaCondHolds(c, have) {
				return nil, fail(http.StatusForbidden, codeRequirementsNotMet, "room", "a ticket meta was used up meanwhile")
			}
			if _, err := tx.ExecContext(ctx, `UPDATE economy_user_meta SET value = value - $3::numeric, updated_at=now() WHERE user_id=$1 AND meta_id=$2`, userID, c.MetaID, strconv.FormatFloat(c.Consume, 'f', -1, 64)); err != nil {
				return nil, err
			}
			t.Metas = append(t.Metas, gameTicketMeta{ID: c.MetaID, Key: key, Amount: c.Consume})
		}
	}
	return t, nil
}

func decodeTickets(raw sql.NullString) *gameTickets {
	if !raw.Valid || raw.String == "" || raw.String == "null" {
		return nil
	}
	t := &gameTickets{}
	if json.Unmarshal([]byte(raw.String), t) != nil {
		return nil
	}
	return t
}

// returnTickets gives back an entry's tickets once (refund "return" only) and marks them returned.
func returnTickets(ctx context.Context, tx *sql.Tx, matchID, userID string, raw sql.NullString) (*gameTickets, error) {
	t := decodeTickets(raw)
	if t.empty() || t.Returned || t.Refund == "keep" {
		return nil, nil
	}
	for _, it := range t.Items {
		src := it.Source
		if src == "" || src == "starter" {
			src = "game"
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref, upgrade_level) SELECT $1, id, $3, $4, $5 FROM economy_item WHERE id=$2
			ON CONFLICT (user_id, item_id) DO NOTHING`, userID, it.ID, src, it.SourceRef, it.UpgradeLevel); err != nil {
			return nil, err
		}
	}
	for _, m := range t.Metas {
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) SELECT $1, id, default_value + $3::numeric FROM economy_meta WHERE id=$2
			ON CONFLICT (user_id, meta_id) DO UPDATE SET value = economy_user_meta.value + $3::numeric, updated_at=now()`, userID, m.ID, strconv.FormatFloat(m.Amount, 'f', -1, 64)); err != nil {
			return nil, err
		}
	}
	t.Returned = true
	b, _ := json.Marshal(t)
	if _, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET tickets=$3::jsonb, updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, userID, string(b)); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) logTickets(ctx context.Context, userID, matchID, gameKey, roomKey, action string, t *gameTickets) {
	if t.empty() {
		return
	}
	data := t.public()
	data["game"], data["room"] = gameKey, roomKey
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: action, RefID: matchID, Actor: "system", Data: data})
}

// gameCheck answers whether a player may enter a room on a path without entering (gamecatalog.Check).
func (s *service) gameCheck(ctx context.Context, e gamecatalog.Entry) error {
	path, err := cleanPath(e.Path)
	if err != nil {
		return err
	}
	g, err := s.gameByKey(ctx, e.Game)
	if err != nil {
		return err
	}
	if !g.Active {
		return s.gameDisabled(ctx, g, e.UserID, "")
	}
	room, err := g.room(e.Room)
	if err != nil {
		return err
	}
	level, err := s.playerLevel(ctx, e.UserID)
	if err != nil {
		return err
	}
	if room.MinLevel > 0 && level < room.MinLevel {
		app := fail(http.StatusForbidden, "LEVEL_TOO_LOW", "room", "this room needs a higher level")
		app.Meta = map[string]any{"minLevel": room.MinLevel, "level": level}
		return app
	}
	rc, err := s.checkRoomRequirements(ctx, e.UserID, g, room, path, level)
	if err != nil {
		return err
	}
	if len(rc.Missing) > 0 {
		return s.requirementsError(ctx, e.UserID, "", g, room, path, rc)
	}
	return nil
}

func (p gameProvider) Check(ctx context.Context, e gamecatalog.Entry) error {
	return providerError(p.s.gameCheck(ctx, e))
}

// gameReqUsed is an IN_USE check for an item, product (list field) or meta id ($1) named in a game or room requirement set.
func gameReqUsed(field string) string {
	match := `(%s->'` + field + `') ? $1::text`
	if field == "metas" {
		match = `(%s->'metas') @> jsonb_build_array(jsonb_build_object('metaId',$1::text))`
	}
	return `EXISTS (SELECT 1 FROM economy_game g WHERE ` + strings.ReplaceAll(match, "%s", "g.requirements") +
		` OR EXISTS (SELECT 1 FROM jsonb_array_elements(g.rooms) r WHERE ` + strings.ReplaceAll(match, "%s", "(r->'requirements')") + `))`
}
