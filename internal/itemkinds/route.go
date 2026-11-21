package itemkinds

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

const (
	ActCreated     = "created"
	ActUpdated     = "updated"
	ActLinked      = "linked"
	ActUnchanged   = "unchanged"
	ActPending     = "pending"
	ActDeactivated = "deactivated"
	ActDependency  = "dependency"
	ActStored      = "stored"
	ActSkipped     = "skipped"
	ActError       = "error"
)

type RunInput struct {
	Changed []string
	Deleted []string
	All     bool
	Kind    string
	DryRun  bool
	Force   bool
	Actor   string
	Source  string
}

type Change struct {
	Field  string `json:"field"`
	Before any    `json:"before"`
	After  any    `json:"after"`
}

type Row struct {
	AssetID string   `json:"assetId"`
	Type    string   `json:"type"`
	Kind    string   `json:"kind,omitempty"`
	ItemKey string   `json:"itemKey,omitempty"`
	Action  string   `json:"action"`
	Reason  string   `json:"reason,omitempty"`
	Code    string   `json:"code,omitempty"`
	Parents []string `json:"parents,omitempty"`
	Changes []Change `json:"changes,omitempty"`
}

type RunOut struct {
	Enabled bool           `json:"enabled"`
	DryRun  bool           `json:"dryRun"`
	Summary map[string]int `json:"summary"`
	Rows    []Row          `json:"rows"`
	RunID   int64          `json:"runId,omitempty"`
}

type RejectError struct {
	Code   string
	Detail string
	Rows   []Row
}

func (e *RejectError) Error() string { return e.Code + ": " + e.Detail }

func LoadRoutes(ctx context.Context, q Querier, onlyEnabled bool) ([]Route, error) {
	where := ""
	if onlyEnabled {
		where = "WHERE enabled"
	}
	rows, err := q.QueryContext(ctx, `SELECT id, source_type, match::text, kind, role, priority, enabled, note, origin, updated_by, updated_at FROM itemkind_route `+where+` ORDER BY priority, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Route{}
	for rows.Next() {
		var r Route
		var match string
		var at sql.NullTime
		if err := rows.Scan(&r.ID, &r.SourceType, &match, &r.Kind, &r.Role, &r.Priority, &r.Enabled, &r.Note, &r.Origin, &r.UpdatedBy, &at); err != nil {
			return nil, err
		}
		r.Match = map[string]any{}
		_ = json.Unmarshal([]byte(match), &r.Match)
		if at.Valid {
			r.UpdatedAt = at.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type target struct {
	entry  *Entry
	kind   *Kind
	cfg    map[string]any
	mapped *Mapped
	hash   string
	used   map[string]string
	err    error
}

func Run(ctx context.Context, tx *sql.Tx, s Settings, in RunInput) (RunOut, error) {
	out := RunOut{Enabled: s.Routing.Enabled, DryRun: in.DryRun, Summary: map[string]int{}, Rows: []Row{}}
	if !s.Routing.Enabled || len(List()) == 0 {
		out.Enabled = false
		return out, nil
	}
	routes, err := LoadRoutes(ctx, tx, true)
	if err != nil {
		return out, err
	}
	m := newMapper(ctx, tx, s, routes)
	if err := m.load(); err != nil {
		return out, err
	}
	targets := map[string]*target{}
	order := []string{}
	addTarget := func(e *Entry) bool {
		r := m.routeFor(e)
		if r == nil || r.Role != RoleItem {
			return false
		}
		k := Get(r.Kind)
		if k == nil || k.Map == nil || !s.KindEnabled(k) {
			return false
		}
		if in.Kind != "" && k.Key != in.Kind {
			return false
		}
		if _, ok := targets[e.AssetID]; !ok {
			targets[e.AssetID] = &target{entry: e, kind: k, cfg: s.KindConfig(k)}
			order = append(order, e.AssetID)
		}
		return true
	}
	parentsOf := func(id string) []string {
		seen := map[string]bool{id: true}
		found := []string{}
		frontier := []string{id}
		for depth := 0; depth < s.Routing.MaxDepth && len(frontier) > 0; depth++ {
			next := []string{}
			for _, cur := range frontier {
				for _, pid := range m.referrers[cur] {
					if seen[pid] {
						continue
					}
					seen[pid] = true
					p := m.entries[pid]
					if p == nil {
						continue
					}
					if addTarget(p) {
						found = append(found, pid)
					}
					next = append(next, pid)
				}
			}
			frontier = next
		}
		sort.Strings(found)
		return found
	}
	info := []Row{}
	if in.All {
		ids := make([]string, 0, len(m.entries))
		for id := range m.entries {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			addTarget(m.entries[id])
		}
	}
	for _, id := range in.Changed {
		e := m.entries[id]
		if e == nil {
			continue
		}
		if addTarget(e) {
			continue
		}
		r := m.routeFor(e)
		if r != nil && r.Role == RoleDependent {
			info = append(info, Row{AssetID: id, Type: e.SourceType, Action: ActDependency, Parents: parentsOf(id)})
			continue
		}
		info = append(info, Row{AssetID: id, Type: e.SourceType, Action: ActStored})
	}
	deletedRows := []Row{}
	for _, id := range in.Deleted {
		parents := parentsOf(id)
		row, err := deleteProfile(ctx, tx, s, id, in.Actor)
		if err != nil {
			return out, err
		}
		if row != nil {
			deletedRows = append(deletedRows, *row)
		} else if len(parents) > 0 {
			info = append(info, Row{AssetID: id, Action: ActDependency, Parents: parents})
		}
	}
	list := make([]*target, 0, len(order))
	for _, id := range order {
		t := targets[id]
		m.reset(t.kind, t.cfg)
		t.mapped, t.err = t.kind.Map(ctx, m, t.entry)
		t.used = m.used
		t.hash = m.hash(t.entry, t.kind, t.cfg)
		list = append(list, t)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].kind.Order != list[j].kind.Order {
			return list[i].kind.Order < list[j].kind.Order
		}
		ri, rj := 0, 0
		if list[i].mapped != nil {
			ri = list[i].mapped.Rank
		}
		if list[j].mapped != nil {
			rj = list[j].mapped.Rank
		}
		if ri != rj {
			return ri < rj
		}
		return list[i].entry.AssetID < list[j].entry.AssetID
	})
	itemRows := []Row{}
	for _, t := range list {
		row, err := apply(ctx, tx, s, t, in)
		if err != nil {
			return out, err
		}
		itemRows = append(itemRows, row)
	}
	out.Rows = append(append(itemRows, deletedRows...), info...)
	for _, r := range out.Rows {
		out.Summary[r.Action]++
	}
	if s.Routing.OnMissingRef == "reject" {
		bad := []Row{}
		for _, r := range itemRows {
			if r.Action == ActPending {
				bad = append(bad, r)
			}
		}
		if len(bad) > 0 {
			return out, &RejectError{Code: "ITEMKIND_REF_MISSING", Detail: bad[0].Reason, Rows: bad}
		}
	}
	return out, nil
}

func SaveRun(ctx context.Context, db Querier, s Settings, source, actor string, out *RunOut) {
	if s.Routing.KeepRuns == 0 || (len(out.Rows) == 0 && source == "push") {
		return
	}
	sum, _ := json.Marshal(out.Summary)
	rows := out.Rows
	if len(rows) > 5000 {
		rows = rows[:5000]
	}
	raw, _ := json.Marshal(rows)
	_ = db.QueryRowContext(ctx, `INSERT INTO itemkind_run (source, dry_run, actor, summary, rows) VALUES ($1,$2,$3,$4,$5) RETURNING id`, source, out.DryRun, actor, string(sum), string(raw)).Scan(&out.RunID)
	_, _ = db.ExecContext(ctx, `DELETE FROM itemkind_run WHERE id NOT IN (SELECT id FROM itemkind_run ORDER BY id DESC LIMIT $1)`, s.Routing.KeepRuns)
}

type itemRef struct {
	id        string
	key       string
	body      map[string]any
	profile   *Profile
	unlock    []Price
	hasUnlock bool
}

func findItem(ctx context.Context, tx *sql.Tx, e *Entry, mp *Mapped) (*itemRef, error) {
	var id string
	err := tx.QueryRowContext(ctx, `SELECT item_id::text FROM itemkind_profile WHERE source_asset_id=$1`, e.AssetID).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if id == "" && mp.Key != "" {
		err = tx.QueryRowContext(ctx, `SELECT id::text FROM economy_item WHERE key=$1`, mp.Key).Scan(&id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if id == "" {
		keys := make([]string, 0, len(mp.MatchMeta))
		for k := range mp.MatchMeta {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			err = tx.QueryRowContext(ctx, `SELECT id::text FROM economy_item WHERE metadata->>$1 = $2 ORDER BY created_at LIMIT 1`, k, mp.MatchMeta[k]).Scan(&id)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if id != "" {
				break
			}
		}
	}
	if id == "" {
		return nil, nil
	}
	return loadItem(ctx, tx, id)
}

const itemSnapshot = `SELECT json_build_object('key', key, 'typeId', type_id::text, 'name', name, 'description', description, 'iconUrl', icon_url, 'imageUrl', image_url, 'soundUrl', sound_url,
	'categoryId', COALESCE(category_id::text, ''), 'tagIds', tag_ids, 'rarity', rarity, 'parentKey', COALESCE(parent_key,''), 'requiresKey', COALESCE(requires_key,''), 'isStarter', is_starter,
	'levelTableId', COALESCE(level_table_id::text, ''), 'unlockLevel', unlock_level, 'price', price, 'upgrades', upgrades, 'services', services, 'metadata', metadata, 'sort', sort, 'isActive', is_active, 'giftable', giftable)::text, key
	FROM economy_item WHERE id=$1`

func loadItem(ctx context.Context, q Querier, id string) (*itemRef, error) {
	var raw, key string
	if err := q.QueryRowContext(ctx, itemSnapshot, id).Scan(&raw, &key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ref := &itemRef{id: id, key: key, body: map[string]any{}}
	_ = json.Unmarshal([]byte(raw), &ref.body)
	p, err := loadProfile(ctx, q, id)
	if err != nil {
		return nil, err
	}
	ref.profile = p
	has, err := tableExists(ctx, q, "inventoryx_item")
	if err != nil || !has {
		return ref, err
	}
	var unlock string
	err = q.QueryRowContext(ctx, `SELECT unlock_prices::text FROM inventoryx_item WHERE item_id=$1`, id).Scan(&unlock)
	if err == nil {
		ref.hasUnlock = true
		_ = json.Unmarshal([]byte(unlock), &ref.unlock)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return ref, nil
}

const profileCols = `p.item_id::text, i.key, COALESCE(i.parent_key,''), p.kind, p.data::text, p.status, p.reason, p.schema_version, p.source_asset_id, p.source_type, p.source_hash, p.source_version, p.linked_at, p.updated_by, p.updated_at`

func scanProfile(sc interface{ Scan(...any) error }) (Profile, error) {
	var p Profile
	var data string
	var linked, at sql.NullTime
	if err := sc.Scan(&p.ItemID, &p.ItemKey, &p.ParentKey, &p.Kind, &data, &p.Status, &p.Reason, &p.SchemaVersion, &p.SourceAssetID, &p.SourceType, &p.SourceHash, &p.SourceVersion, &linked, &p.UpdatedBy, &at); err != nil {
		return p, err
	}
	p.Data = map[string]any{}
	_ = json.Unmarshal([]byte(data), &p.Data)
	if linked.Valid {
		p.LinkedAt = linked.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	if at.Valid {
		p.UpdatedAt = at.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	return p, nil
}

func loadProfile(ctx context.Context, q Querier, itemID string) (*Profile, error) {
	p, err := scanProfile(q.QueryRowContext(ctx, `SELECT `+profileCols+` FROM itemkind_profile p JOIN economy_item i ON i.id=p.item_id WHERE p.item_id=$1`, itemID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func Profiles(ctx context.Context, q Querier, itemIDs []string) (map[string]Profile, error) {
	out := map[string]Profile{}
	if len(itemIDs) == 0 {
		return out, nil
	}
	rows, err := q.QueryContext(ctx, `SELECT `+profileCols+` FROM itemkind_profile p JOIN economy_item i ON i.id=p.item_id WHERE p.item_id::text = ANY($1)`, itemIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanProfile(rows)
		if err != nil {
			return nil, err
		}
		out[p.ItemID] = p
	}
	return out, rows.Err()
}

func ProfileByKey(ctx context.Context, q Querier, key string) (*Profile, error) {
	p, err := scanProfile(q.QueryRowContext(ctx, `SELECT `+profileCols+` FROM itemkind_profile p JOIN economy_item i ON i.id=p.item_id WHERE i.key=$1`, key))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func apply(ctx context.Context, tx *sql.Tx, s Settings, t *target, in RunInput) (Row, error) {
	e := t.entry
	row := Row{AssetID: e.AssetID, Type: e.SourceType, Kind: t.kind.Key}
	if t.err != nil {
		row.Action, row.Reason = ActError, t.err.Error()
		return row, nil
	}
	mp := t.mapped
	if mp == nil || mp.Skip != "" {
		row.Action = ActSkipped
		if mp != nil {
			row.Reason = mp.Skip
		}
		return row, nil
	}
	row.ItemKey = mp.Key
	if _, err := tx.ExecContext(ctx, `SAVEPOINT itemkind_row`); err != nil {
		return row, err
	}
	fail := func(code, reason string) (Row, error) {
		if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT itemkind_row`); err != nil {
			return row, err
		}
		row.Action, row.Code, row.Reason, row.Changes = ActError, code, reason, nil
		return row, nil
	}
	cur, err := findItem(ctx, tx, e, mp)
	if err != nil {
		return fail("INTERNAL", err.Error())
	}
	if cur != nil {
		row.ItemKey = cur.key
	}
	if cur != nil && cur.profile != nil && !in.Force && cur.profile.SourceHash == t.hash && cur.profile.Status != "deleted" && cur.profile.SourceAssetID == e.AssetID {
		row.Action = ActUnchanged
		if cur.profile.Status == "pending" {
			row.Action, row.Reason = ActPending, cur.profile.Reason
		}
		_, _ = tx.ExecContext(ctx, `RELEASE SAVEPOINT itemkind_row`)
		return row, nil
	}
	if cur != nil && cur.profile != nil && cur.profile.Kind != t.kind.Key {
		return fail("ITEMKIND_KIND_MISMATCH", "item "+cur.key+" already belongs to kind "+cur.profile.Kind)
	}
	if cur != nil && cur.profile != nil && cur.profile.SourceAssetID != "" && cur.profile.SourceAssetID != e.AssetID && cur.profile.Status != "deleted" {
		return fail("ITEMKIND_SOURCE_TAKEN", "item "+cur.key+" is already linked to entry "+cur.profile.SourceAssetID)
	}
	pending := mp.Pending
	if pending == "" && t.kind.Validate != nil {
		if err := t.kind.Validate(mp.Data); err != nil {
			pending = err.Error()
		}
	}
	if pending == "" {
		if err := ValidateData(t.kind.Fields, mp.Data); err != nil {
			pending = err.Error()
		}
	}
	create := cur == nil
	firstLink := cur != nil && (cur.profile == nil || cur.profile.LinkedAt == "")
	if pending != "" && !create {
		row.Action, row.Reason = ActPending, pending
		if err := upsertProfile(ctx, tx, s, cur.id, t, mp, "pending", pending, in.Actor, cur.profile); err != nil {
			return fail("INTERNAL", err.Error())
		}
		_, _ = tx.ExecContext(ctx, `RELEASE SAVEPOINT itemkind_row`)
		return row, nil
	}
	body := map[string]any{}
	var id string
	if create {
		id = uuid.Must(uuid.NewV4()).String()
		var typeID string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='item_type' AND key=$1`, t.kind.ItemType).Scan(&typeID); err != nil {
			return fail("ITEM_TYPE_NOT_FOUND", "item type "+t.kind.ItemType+" does not exist")
		}
		body = map[string]any{"key": mp.Key, "typeId": typeID, "name": map[string]string{}, "description": map[string]string{}, "tagIds": []string{}, "upgrades": []any{}, "services": []any{}, "isActive": true, "metadata": map[string]any{}}
	} else {
		id = cur.id
		for k, v := range cur.body {
			body[k] = v
		}
	}
	before := cloneMap(body)
	writes := func(field string) bool {
		if create {
			return true
		}
		if s.Owner(t.kind, field) == "unity" {
			return true
		}
		return firstLink && s.Routing.LinkWritesUnity
	}
	langs := (&Mapper{ctx: ctx, q: tx}).Langs()
	cleanText := func(tx Text) map[string]string {
		outText := map[string]string{}
		for k, v := range tx {
			v = strings.TrimSpace(v)
			if v == "" || (langs != nil && !langs.Known(k)) {
				continue
			}
			outText[k] = v
		}
		return outText
	}
	if len(mp.Name) > 0 && writes("name") {
		if n := cleanText(mp.Name); len(n) > 0 {
			body["name"] = n
		}
	}
	if mp.Description != nil && writes("description") {
		body["description"] = cleanText(mp.Description)
	}
	if (mp.Price != nil || mp.NoPrice) && writes("price") {
		if mp.Price != nil {
			body["price"] = map[string]any{"currencyKey": mp.Price.CurrencyKey, "amount": mp.Price.Amount}
		} else {
			body["price"] = nil
		}
	}
	if mp.UnlockLevel != nil && writes("unlockLevel") {
		if *mp.UnlockLevel >= 1 {
			body["unlockLevel"] = *mp.UnlockLevel
		} else {
			body["unlockLevel"] = nil
		}
	}
	if mp.ParentKey != nil && writes("parentKey") {
		body["parentKey"] = *mp.ParentKey
	}
	if mp.RequiresKey != nil && writes("requiresKey") {
		body["requiresKey"] = *mp.RequiresKey
	}
	if mp.Sort != nil && writes("sort") {
		body["sort"] = *mp.Sort
	}
	active := true
	if mp.Active != nil {
		active = *mp.Active
	}
	if pending != "" {
		active = false
	}
	if writes("isActive") || (cur != nil && cur.profile != nil && cur.profile.Status != "ok") {
		body["isActive"] = active
	}
	if len(mp.Metadata) > 0 {
		meta, _ := body["metadata"].(map[string]any)
		if meta == nil {
			meta = map[string]any{}
		} else {
			meta = cloneMap(meta)
		}
		for k, v := range mp.Metadata {
			meta[k] = v
		}
		body["metadata"] = meta
	}
	changes := diff(before, body)
	if create || len(changes) > 0 {
		raw, _ := json.Marshal(body)
		if err := economyhooks.SaveItem(ctx, tx, id, create, raw); err != nil {
			if errors.Is(err, economyhooks.ErrNoItemSaver) {
				return fail("ECONOMY_UNAVAILABLE", "the economy plugin is not loaded")
			}
			if r, ok := economyhooks.AsRefusal(err); ok {
				field, _ := r.Meta["field"].(string)
				return fail(r.Code, strings.TrimSpace(field+" "+r.Message))
			}
			return fail("INTERNAL", err.Error())
		}
	}
	if mp.HasUnlock && writes("unlockPrices") {
		hasTable, err := tableExists(ctx, tx, "inventoryx_item")
		if err != nil {
			return fail("INTERNAL", err.Error())
		}
		if hasTable {
			var prev []Price
			if cur != nil {
				prev = cur.unlock
			}
			if !samePrices(prev, mp.UnlockPrices) {
				raw, _ := json.Marshal(nonNilPrices(mp.UnlockPrices))
				if _, err := tx.ExecContext(ctx, `INSERT INTO inventoryx_item (item_id, unlock_prices, updated_by) VALUES ($1,$2,$3)
					ON CONFLICT (item_id) DO UPDATE SET unlock_prices=EXCLUDED.unlock_prices, updated_by=EXCLUDED.updated_by, updated_at=now()`, id, string(raw), in.Actor); err != nil {
					return fail("INTERNAL", err.Error())
				}
				changes = append(changes, Change{Field: "unlockPrices", Before: prev, After: nonNilPrices(mp.UnlockPrices)})
			}
		}
	}
	var prevProfile *Profile
	if cur != nil {
		prevProfile = cur.profile
	}
	status := "ok"
	if pending != "" {
		status = "pending"
	}
	changes = append(changes, dataDiff(prevProfile, mp.Data)...)
	if err := upsertProfile(ctx, tx, s, id, t, mp, status, pending, in.Actor, prevProfile); err != nil {
		return fail("INTERNAL", err.Error())
	}
	if _, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT itemkind_row`); err != nil {
		return row, err
	}
	row.ItemKey = fmt.Sprint(body["key"])
	row.Changes = changes
	switch {
	case pending != "":
		row.Action, row.Reason = ActPending, pending
	case create:
		row.Action = ActCreated
	case firstLink:
		row.Action = ActLinked
	case len(changes) == 0:
		row.Action = ActUnchanged
	default:
		row.Action = ActUpdated
	}
	return row, nil
}

func upsertProfile(ctx context.Context, tx *sql.Tx, s Settings, itemID string, t *target, mp *Mapped, status, reason, actor string, prev *Profile) error {
	data := cloneMap(mp.Data)
	if prev != nil {
		for field, v := range prev.Data {
			if s.Owner(t.kind, "data."+field) == "console" {
				data[field] = v
			}
		}
	}
	raw, _ := json.Marshal(data)
	_, err := tx.ExecContext(ctx, `INSERT INTO itemkind_profile (item_id, kind, data, status, reason, schema_version, source_asset_id, source_type, source_hash, source_version, linked_at, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,CASE WHEN $4='ok' THEN now() END,$11)
		ON CONFLICT (item_id) DO UPDATE SET kind=EXCLUDED.kind, data=EXCLUDED.data, status=EXCLUDED.status, reason=EXCLUDED.reason, schema_version=EXCLUDED.schema_version,
			source_asset_id=EXCLUDED.source_asset_id, source_type=EXCLUDED.source_type, source_hash=EXCLUDED.source_hash, source_version=EXCLUDED.source_version,
			linked_at=COALESCE(itemkind_profile.linked_at, EXCLUDED.linked_at), updated_by=EXCLUDED.updated_by, updated_at=now()`,
		itemID, t.kind.Key, string(raw), status, reason, t.kind.SchemaVersion, t.entry.AssetID, t.entry.SourceType, t.hash, t.entry.Version, actor)
	return err
}

func deleteProfile(ctx context.Context, tx *sql.Tx, s Settings, assetID, actor string) (*Row, error) {
	var itemID, kind, key string
	err := tx.QueryRowContext(ctx, `SELECT p.item_id::text, p.kind, i.key FROM itemkind_profile p JOIN economy_item i ON i.id=p.item_id WHERE p.source_asset_id=$1`, assetID).Scan(&itemID, &kind, &key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE itemkind_profile SET status='deleted', reason='source entry deleted', source_hash='', updated_by=$2, updated_at=now() WHERE item_id=$1`, itemID, actor); err != nil {
		return nil, err
	}
	row := &Row{AssetID: assetID, Kind: kind, ItemKey: key, Action: ActSkipped, Reason: "source entry deleted; item kept"}
	if s.Routing.OnDelete == "deactivate" {
		if _, err := tx.ExecContext(ctx, `UPDATE economy_item SET is_active=false, updated_at=now() WHERE id=$1`, itemID); err != nil {
			return nil, err
		}
		row.Action, row.Reason = ActDeactivated, "source entry deleted"
	}
	return row, nil
}

func tableExists(ctx context.Context, q Querier, name string) (bool, error) {
	var ok bool
	err := q.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, name).Scan(&ok)
	return ok, err
}

func samePrices(a, b []Price) bool {
	x, _ := json.Marshal(nonNilPrices(a))
	y, _ := json.Marshal(nonNilPrices(b))
	return string(x) == string(y)
}

func nonNilPrices(p []Price) []Price {
	if p == nil {
		return []Price{}
	}
	return p
}

func cloneMap(m map[string]any) map[string]any {
	raw, _ := json.Marshal(m)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func diff(before, after map[string]any) []Change {
	a := cloneMap(after)
	keys := []string{}
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := []Change{}
	for _, k := range keys {
		x, _ := json.Marshal(before[k])
		y, _ := json.Marshal(a[k])
		if string(x) != string(y) {
			out = append(out, Change{Field: k, Before: before[k], After: a[k]})
		}
	}
	return out
}

func dataDiff(prev *Profile, data map[string]any) []Change {
	old := map[string]any{}
	if prev != nil {
		old = prev.Data
	}
	out := []Change{}
	for _, c := range diff(old, data) {
		c.Field = "profile." + c.Field
		out = append(out, c)
	}
	for k, v := range old {
		if _, ok := data[k]; !ok {
			out = append(out, Change{Field: "profile." + k, Before: v, After: nil})
		}
	}
	return out
}

func Snapshot(ctx context.Context, q Querier, itemID string) (map[string]any, *Profile, error) {
	ref, err := loadItem(ctx, q, itemID)
	if err != nil || ref == nil {
		return nil, nil, err
	}
	return ref.body, ref.profile, nil
}

func UnlockPrices(ctx context.Context, q Querier, itemID string) ([]Price, bool, error) {
	ref, err := loadItem(ctx, q, itemID)
	if err != nil || ref == nil {
		return nil, false, err
	}
	return ref.unlock, ref.hasUnlock, nil
}

func PatchItem(ctx context.Context, tx *sql.Tx, itemID string, create bool, body map[string]any) error {
	raw, _ := json.Marshal(body)
	return economyhooks.SaveItem(ctx, tx, itemID, create, raw)
}

func SaveProfile(ctx context.Context, tx *sql.Tx, itemID, kind string, schemaVersion int, data map[string]any, actor string) error {
	raw, _ := json.Marshal(data)
	_, err := tx.ExecContext(ctx, `INSERT INTO itemkind_profile (item_id, kind, data, schema_version, updated_by) VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (item_id) DO UPDATE SET data=EXCLUDED.data, kind=EXCLUDED.kind, updated_by=EXCLUDED.updated_by, updated_at=now()`, itemID, kind, string(raw), schemaVersion, actor)
	return err
}
