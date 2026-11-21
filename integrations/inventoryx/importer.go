package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

// Bulk item import: rows create or update economy items by key (through economy's own item checks,
// via economyhooks.SaveItem) and set this module's unlock prices and upgrade tracks. Legacy rows
// ({key, type, name, prices:[{currencyId, amount}], order, upgrades:[{key, currencyId, prices}],
// unlockPrices, services}) are accepted as they are.

const ScopeItems = "items:write"

type importService struct {
	Key         string          `json:"key"`
	Name        json.RawMessage `json:"name"`
	CurrencyKey string          `json:"currencyKey"`
	CurrencyID  string          `json:"currencyId"`
	UnitPrice   float64         `json:"unitPrice"`
	MaxUnits    int             `json:"maxUnits"`
}

// fullRow is one import row; fields left out are not sent (nil / absent in sent).
type fullRow struct {
	importRow
	Type        string          `json:"type"`
	TypeID      string          `json:"typeId"`
	Name        json.RawMessage `json:"name"`
	Description json.RawMessage `json:"description"`
	Rarity      string          `json:"rarity"`
	ParentKey   *string         `json:"parentKey"`
	RequiresKey *string         `json:"requiresKey"`
	IsActive    *bool           `json:"isActive"`
	IsStarter   *bool           `json:"isStarter"`
	UnlockLevel *int            `json:"unlockLevel"`
	LevelTable  *string         `json:"levelTable"`
	Category    *string         `json:"category"`
	Prices      []importPrice   `json:"prices"`
	Price       *importPrice    `json:"price"`
	Order       *int            `json:"order"`
	Sort        *int            `json:"sort"`
	Services    []importService `json:"services"`
	IconURL     *string         `json:"iconUrl"`
	ImageURL    *string         `json:"imageUrl"`
	Metadata    json.RawMessage `json:"metadata"`

	sent map[string]json.RawMessage
}

type importIn struct {
	Items  []fullRow `json:"items"`
	DryRun bool      `json:"dryRun"`
}

type importReport struct {
	Key     string   `json:"key"`
	Action  string   `json:"action"`
	ItemID  string   `json:"itemId,omitempty"`
	Code    string   `json:"code,omitempty"`
	Field   string   `json:"field,omitempty"`
	Message string   `json:"message,omitempty"`
	Changes []string `json:"changes"`
}

func parseImport(raw []byte) (importIn, error) {
	in := importIn{}
	var list []json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		var wrapped struct {
			Items  []json.RawMessage `json:"items"`
			DryRun bool              `json:"dryRun"`
		}
		if err := json.Unmarshal(raw, &wrapped); err != nil {
			return in, apierr.BadRequest("INVALID_JSON").WithDetail("send an array of items or {items:[…], dryRun}")
		}
		list, in.DryRun = wrapped.Items, wrapped.DryRun
	}
	for i, r := range list {
		var row fullRow
		if err := json.Unmarshal(r, &row); err != nil {
			return in, apierr.BadRequest("INVALID_JSON").WithField(fmt.Sprintf("items[%d]", i)).WithDetail(err.Error())
		}
		if err := json.Unmarshal(r, &row.sent); err != nil {
			return in, apierr.BadRequest("INVALID_JSON").WithField(fmt.Sprintf("items[%d]", i))
		}
		in.Items = append(in.Items, row)
	}
	return in, nil
}

func (r fullRow) has(names ...string) bool {
	for _, n := range names {
		if _, ok := r.sent[n]; ok {
			return true
		}
	}
	return false
}

var itemFieldNames = []string{"type", "typeId", "name", "description", "rarity", "parentKey", "requiresKey", "isActive", "isStarter", "unlockLevel", "levelTable", "category", "prices", "price", "order", "sort", "services", "iconUrl", "imageUrl", "metadata"}

func textValue(raw json.RawMessage, def string, cur map[string]string) (map[string]string, bool) {
	out := map[string]string{}
	for k, v := range cur {
		out[k] = v
	}
	if len(raw) == 0 || string(raw) == "null" {
		return out, false
	}
	var plain string
	if json.Unmarshal(raw, &plain) == nil {
		out[def] = strings.TrimSpace(plain)
		if out[def] == "" {
			delete(out, def)
		}
		return out, true
	}
	var m map[string]string
	if json.Unmarshal(raw, &m) == nil {
		for k, v := range m {
			if v = strings.TrimSpace(v); v == "" {
				delete(out, k)
			} else {
				out[k] = v
			}
		}
		return out, true
	}
	return out, false
}

type refusal struct{ code, field, message string }

func (e *refusal) Error() string { return e.code }

func refuse(code, field, msg string) error { return &refusal{code, field, msg} }

// itemBody builds economy's item editor body from the stored item (or nothing) and a row.
func (s *service) itemBody(ctx context.Context, tx *sql.Tx, row fullRow, base map[string]any, def string, cfg settings.Import) (map[string]any, error) {
	body := map[string]any{"key": row.Key, "name": map[string]string{}, "description": map[string]string{}, "tagIds": []string{}, "upgrades": []any{}, "services": []any{}, "isActive": true}
	if base != nil && cfg.MissingFields == "keep" {
		for k, v := range base {
			body[k] = v
		}
	}
	lookup := func(q, key, code, field string) (any, error) {
		if strings.TrimSpace(key) == "" {
			return nil, nil
		}
		var id string
		if err := tx.QueryRowContext(ctx, q, strings.TrimSpace(key)).Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, refuse(code, field, key)
			}
			return nil, err
		}
		return id, nil
	}
	typeKey := strings.TrimSpace(row.Type)
	switch {
	case strings.TrimSpace(row.TypeID) != "":
		body["typeId"] = strings.TrimSpace(row.TypeID)
	case typeKey != "":
		id, err := lookup(`SELECT id::text FROM base_category WHERE key_type='item_type' AND key=$1`, typeKey, "ITEM_TYPE_NOT_FOUND", "type")
		if err != nil {
			return nil, err
		}
		body["typeId"] = id
	case body["typeId"] == nil && cfg.DefaultType != "":
		id, err := lookup(`SELECT id::text FROM base_category WHERE key_type='item_type' AND key=$1`, cfg.DefaultType, "ITEM_TYPE_NOT_FOUND", "type")
		if err != nil {
			return nil, err
		}
		body["typeId"] = id
	case body["typeId"] == nil:
		return nil, refuse("ITEM_TYPE_REQUIRED", "type", "a new item needs a type")
	}
	curName, _ := body["name"].(map[string]string)
	name, _ := textValue(row.Name, def, curName)
	if len(name) == 0 {
		name = map[string]string{def: row.Key}
	}
	body["name"] = name
	curDesc, _ := body["description"].(map[string]string)
	if desc, ok := textValue(row.Description, def, curDesc); ok {
		body["description"] = desc
	}
	if row.has("rarity") {
		body["rarity"] = row.Rarity
	}
	if row.ParentKey != nil {
		body["parentKey"] = *row.ParentKey
	}
	if row.RequiresKey != nil {
		body["requiresKey"] = *row.RequiresKey
	}
	if row.IsActive != nil {
		body["isActive"] = *row.IsActive
	}
	if row.IsStarter != nil {
		body["isStarter"] = *row.IsStarter
	}
	if row.has("unlockLevel") {
		if row.UnlockLevel == nil || *row.UnlockLevel <= 0 {
			body["unlockLevel"] = nil
		} else {
			body["unlockLevel"] = *row.UnlockLevel
		}
	}
	if row.LevelTable != nil {
		id, err := lookup(`SELECT id::text FROM economy_level_table WHERE key=$1`, *row.LevelTable, "LEVEL_TABLE_NOT_FOUND", "levelTable")
		if err != nil {
			return nil, err
		}
		body["levelTableId"] = id
	}
	if row.Category != nil {
		id, err := lookup(`SELECT id::text FROM base_category WHERE key_type='item' AND key=$1`, *row.Category, "CATEGORY_NOT_FOUND", "category")
		if err != nil {
			return nil, err
		}
		body["categoryId"] = id
	}
	switch {
	case row.Price != nil:
		body["price"] = map[string]any{"currencyKey": firstNonEmpty(row.Price.CurrencyKey, row.Price.CurrencyID), "amount": int64(math.Round(row.Price.Amount))}
	case row.has("price") || (row.has("prices") && len(row.Prices) == 0):
		body["price"] = nil
	case len(row.Prices) == 1:
		body["price"] = map[string]any{"currencyKey": firstNonEmpty(row.Prices[0].CurrencyKey, row.Prices[0].CurrencyID), "amount": int64(math.Round(row.Prices[0].Amount))}
	case len(row.Prices) > 1:
		return nil, refuse("TOO_MANY_PRICES", "prices", "economy items have one buy price")
	}
	if row.Order != nil {
		body["sort"] = *row.Order
	}
	if row.Sort != nil {
		body["sort"] = *row.Sort
	}
	if row.has("services") {
		list := []map[string]any{}
		for _, sv := range row.Services {
			n, _ := textValue(sv.Name, def, nil)
			list = append(list, map[string]any{"key": sv.Key, "name": n, "currencyKey": firstNonEmpty(sv.CurrencyKey, sv.CurrencyID), "unitPrice": int64(math.Round(sv.UnitPrice)), "maxUnits": sv.MaxUnits})
		}
		body["services"] = list
	}
	if row.IconURL != nil {
		body["iconUrl"] = *row.IconURL
	}
	if row.ImageURL != nil {
		body["imageUrl"] = *row.ImageURL
	}
	if len(row.Metadata) > 0 {
		body["metadata"] = row.Metadata
	}
	return body, nil
}

const itemSnapshot = `SELECT json_build_object('key', key, 'typeId', type_id::text, 'name', name, 'description', description, 'iconUrl', icon_url, 'imageUrl', image_url,
	'categoryId', category_id::text, 'tagIds', tag_ids, 'rarity', rarity, 'parentKey', coalesce(parent_key,''), 'requiresKey', coalesce(requires_key,''), 'isStarter', is_starter,
	'levelTableId', level_table_id::text, 'unlockLevel', unlock_level, 'price', price, 'upgrades', upgrades, 'services', services, 'metadata', metadata, 'sort', sort, 'isActive', is_active)::text, id::text
	FROM economy_item WHERE key=$1`

func snapshot(ctx context.Context, tx *sql.Tx, key string) (map[string]any, string, error) {
	var raw, id string
	err := tx.QueryRowContext(ctx, itemSnapshot, key).Scan(&raw, &id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	out := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &out)
	for _, k := range []string{"name", "description"} {
		m := map[string]string{}
		if v, ok := out[k].(map[string]any); ok {
			for kk, vv := range v {
				if s, ok := vv.(string); ok {
					m[kk] = s
				}
			}
		}
		out[k] = m
	}
	for _, k := range []string{"typeId", "categoryId", "levelTableId"} {
		if out[k] == nil {
			out[k] = ""
		}
	}
	return out, id, nil
}

func diffKeys(before, after map[string]any) []string {
	out := []string{}
	for k, v := range after {
		a, _ := json.Marshal(v)
		b, _ := json.Marshal(before[k])
		if string(a) != string(b) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// orderRows puts rows whose parent or required item is another row of the batch after that row.
func orderRows(rows []fullRow) []int {
	idx := map[string]int{}
	for i, r := range rows {
		idx[strings.TrimSpace(r.Key)] = i
	}
	done := map[int]bool{}
	visiting := map[int]bool{}
	out := []int{}
	var visit func(i int)
	visit = func(i int) {
		if done[i] || visiting[i] {
			return
		}
		visiting[i] = true
		for _, ref := range []*string{rows[i].ParentKey, rows[i].RequiresKey} {
			if ref != nil {
				if j, ok := idx[strings.TrimSpace(*ref)]; ok && j != i {
					visit(j)
				}
			}
		}
		visiting[i] = false
		done[i] = true
		out = append(out, i)
	}
	for i := range rows {
		visit(i)
	}
	return out
}

// importItems applies the rows; in atomic mode a failed row leaves everything as it was.
func (s *service) runImport(ctx context.Context, in importIn, actor string) (map[string]any, error) {
	all := s.settings(ctx)
	cfg := all.Import
	if len(in.Items) == 0 {
		return nil, apierr.BadRequest("IMPORT_EMPTY")
	}
	if len(in.Items) > cfg.MaxItems {
		return nil, apierr.BadRequest("IMPORT_TOO_MANY").With("max", cfg.MaxItems)
	}
	seen := map[string]bool{}
	for i := range in.Items {
		k := strings.TrimSpace(in.Items[i].Key)
		in.Items[i].Key = k
		if k == "" || seen[k] {
			return nil, apierr.BadRequest("IMPORT_DUPLICATE_KEY").WithField(fmt.Sprintf("items[%d].key", i)).With("key", k)
		}
		seen[k] = true
	}
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	def := langs.Default
	if def == "" {
		def = "fa"
	}
	currencies, err := currencyKeys(ctx, s.db)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	reports := make([]importReport, len(in.Items))
	failed := 0
	for _, i := range orderRows(in.Items) {
		row := in.Items[i]
		rep, err := s.importRow(ctx, tx, row, def, all, langs, currencies, actor, cfg.Mode == "partial")
		if err != nil {
			if errors.Is(err, economyhooks.ErrNoItemSaver) {
				return nil, apierr.Unavailable("ECONOMY_UNAVAILABLE")
			}
			return nil, err
		}
		if rep.Action == "failed" {
			failed++
		}
		reports[i] = rep
	}
	summary := map[string]int{}
	for _, r := range reports {
		summary[r.Action]++
	}
	applied := !in.DryRun && (failed == 0 || cfg.Mode == "partial")
	if applied {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else {
		_ = tx.Rollback()
	}
	s.saveImportReport(ctx, in.DryRun, actor, summary, reports, cfg.KeepReports)
	return map[string]any{"dryRun": in.DryRun, "applied": applied, "mode": cfg.Mode, "rows": reports, "summary": summary,
		"updated": summary["created"] + summary["updated"]}, nil
}

func (s *service) importRow(ctx context.Context, tx *sql.Tx, row fullRow, def string, all settings.All, langs *localized.Set, currencies map[string]bool, actor string, savepoint bool) (importReport, error) {
	rep := importReport{Key: row.Key, Changes: []string{}}
	if savepoint {
		if _, err := tx.ExecContext(ctx, `SAVEPOINT import_row`); err != nil {
			return rep, err
		}
	}
	fail := func(code, field, msg string) (importReport, error) {
		if savepoint {
			if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT import_row`); err != nil {
				return rep, err
			}
		}
		rep.Action, rep.Code, rep.Field, rep.Message, rep.Changes = "failed", code, field, msg, []string{}
		return rep, nil
	}
	before, id, err := snapshot(ctx, tx, row.Key)
	if err != nil {
		return rep, err
	}
	sendsItem := row.has(itemFieldNames...)
	if before == nil && !sendsItem {
		return fail("ITEM_NOT_FOUND", "key", "the item does not exist and the row has no item fields")
	}
	if sendsItem || before == nil {
		body, err := s.itemBody(ctx, tx, row, before, def, all.Import)
		var rf *refusal
		if errors.As(err, &rf) {
			return fail(rf.code, rf.field, rf.message)
		}
		if err != nil {
			return rep, err
		}
		raw, _ := json.Marshal(body)
		create := before == nil
		if create {
			id = uuid.Must(uuid.NewV4()).String()
		}
		if err := economyhooks.SaveItem(ctx, tx, id, create, raw); err != nil {
			if r, ok := economyhooks.AsRefusal(err); ok {
				field, _ := r.Meta["field"].(string)
				return fail(r.Code, field, r.Message)
			}
			return rep, err
		}
		after, _, err := snapshot(ctx, tx, row.Key)
		if err != nil {
			return rep, err
		}
		if create {
			rep.Action = "created"
		} else {
			rep.Changes = diffKeys(before, after)
		}
	}
	rep.ItemID = id
	if row.has("unlockPrices", "tracks", "upgrades") {
		c, err := cleanConfig(all, langs, currencies, row.config(def))
		if err != nil {
			var ce *pluginkit.ConsoleError
			if errors.As(err, &ce) {
				return fail(ce.Code, ce.Field, ce.Message)
			}
			return rep, err
		}
		old, err := loadConfigs(ctx, tx, []string{id})
		if err != nil {
			return rep, err
		}
		prev := old[id]
		if !reflect.DeepEqual(normConfig(prev), normConfig(c)) {
			if !reflect.DeepEqual(normConfig(prev).UnlockPrices, normConfig(c).UnlockPrices) {
				rep.Changes = append(rep.Changes, "unlockPrices")
			}
			if !reflect.DeepEqual(normConfig(prev).Tracks, normConfig(c).Tracks) {
				rep.Changes = append(rep.Changes, "tracks")
			}
			if err := saveConfig(ctx, tx, id, c, actor); err != nil {
				return rep, err
			}
		}
	}
	if rep.Action == "" {
		rep.Action = "unchanged"
		if len(rep.Changes) > 0 {
			rep.Action = "updated"
		}
	}
	if savepoint {
		if _, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT import_row`); err != nil {
			return rep, err
		}
	}
	return rep, nil
}

func normConfig(c itemConfig) itemConfig {
	b, _ := json.Marshal(c)
	var out itemConfig
	_ = json.Unmarshal(b, &out)
	if out.UnlockPrices == nil {
		out.UnlockPrices = []price{}
	}
	if out.Tracks == nil {
		out.Tracks = []track{}
	}
	return out
}

func (s *service) saveImportReport(ctx context.Context, dryRun bool, actor string, summary map[string]int, rows []importReport, keep int) {
	sj, _ := json.Marshal(summary)
	rj, _ := json.Marshal(rows)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO inventoryx_import (dry_run, actor, summary, rows) VALUES ($1,$2,$3,$4)`, dryRun, actor, string(sj), string(rj)); err != nil && s.logger != nil {
		s.logger.Warn("Item import report not saved: %v", err)
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM inventoryx_import WHERE id NOT IN (SELECT id FROM inventoryx_import ORDER BY id DESC LIMIT $1)`, keep)
}

// ---- HTTP ----

func (s *service) importItems(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	if err := pluginkit.ReadBody(r, &raw, 16<<20); err != nil {
		s.fail(w, err)
		return
	}
	in, err := parseImport(raw)
	if err != nil {
		s.fail(w, err)
		return
	}
	if r.URL.Query().Get("dryRun") == "true" {
		in.DryRun = true
	}
	actor := "console"
	if u := pluginkit.ConsoleUser(r); u != "" {
		actor = "console:" + u
	}
	out, err := s.runImport(r.Context(), in, actor)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) httpImport(w http.ResponseWriter, r *http.Request, key *apikey.Key) {
	var raw json.RawMessage
	if err := pluginkit.ReadBody(r, &raw, 16<<20); err != nil {
		var ce *pluginkit.ConsoleError
		if errors.As(err, &ce) {
			err = apierr.New(ce.Status, ce.Code)
		}
		apierr.WriteHTTP(w, r, s.db, err)
		return
	}
	in, err := parseImport(raw)
	if err == nil {
		var out map[string]any
		if out, err = s.runImport(r.Context(), in, "editor:"+key.Name); err == nil {
			pluginkit.WriteJSON(w, http.StatusOK, out)
			return
		}
	}
	if e, ok := apierr.As(err); !ok || e.Status >= 500 {
		s.logger.Error("Item import failed: %v", err)
	}
	apierr.WriteHTTP(w, r, s.db, err)
}

func (s *service) importReports(w http.ResponseWriter, r *http.Request) {
	if id, err := strconv.ParseInt(pluginkit.PathVar(r, "rid"), 10, 64); err == nil && id > 0 {
		var dry bool
		var actor, summary, rows string
		var at any
		err := s.db.QueryRowContext(r.Context(), `SELECT dry_run, actor, summary::text, rows::text, created_at FROM inventoryx_import WHERE id=$1`, id).Scan(&dry, &actor, &summary, &rows, &at)
		if errors.Is(err, sql.ErrNoRows) {
			s.fail(w, pluginkit.Missing("REPORT_NOT_FOUND"))
			return
		}
		if err != nil {
			s.fail(w, err)
			return
		}
		pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "dryRun": dry, "actor": actor, "summary": json.RawMessage(summary), "rows": json.RawMessage(rows), "createdAt": at})
		return
	}
	q, err := s.db.QueryContext(r.Context(), `SELECT id, dry_run, actor, summary::text, created_at FROM inventoryx_import ORDER BY id DESC LIMIT 100`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer q.Close()
	list := []map[string]any{}
	for q.Next() {
		var id int64
		var dry bool
		var actor, summary string
		var at any
		if err := q.Scan(&id, &dry, &actor, &summary, &at); err != nil {
			s.fail(w, err)
			return
		}
		list = append(list, map[string]any{"id": id, "dryRun": dry, "actor": actor, "summary": json.RawMessage(summary), "createdAt": at})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": list})
}
