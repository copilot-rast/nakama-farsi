package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/itemkinds"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	p := consolePrefix
	routes := []route{
		{p + "/overview", http.MethodGet, s.overview},
		{p + "/kinds", http.MethodGet, s.kindList},
		{p + "/kinds/{kind}/items", http.MethodGet, s.itemList},
		{p + "/kinds/{kind}/items", http.MethodPost, s.itemCreate},
		{p + "/items/{id}", http.MethodGet, s.itemGet},
		{p + "/items/{id}", http.MethodPut, s.itemPut},
		{p + "/settings", http.MethodGet, s.settingsGet},
		{p + "/settings", http.MethodPut, s.settingsPut},
		{p + "/routes", http.MethodGet, s.routeList},
		{p + "/routes", http.MethodPost, s.routeCreate},
		{p + "/routes/{id}", http.MethodPut, s.routeUpdate},
		{p + "/routes/{id}", http.MethodDelete, s.routeDelete},
		{p + "/reroute", http.MethodPost, s.reroute},
		{p + "/runs", http.MethodGet, s.runList},
		{p + "/runs/{id}", http.MethodGet, s.runGet},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func actor(r *http.Request) string {
	if u := pluginkit.ConsoleUser(r); u != "" {
		return "console:" + u
	}
	return "console"
}

type kindView struct {
	Key       string             `json:"key"`
	ItemType  string             `json:"itemType"`
	Label     itemkinds.Text     `json:"label"`
	Icon      string             `json:"icon"`
	Order     int                `json:"order"`
	Schema    int                `json:"schemaVersion"`
	Enabled   bool               `json:"enabled"`
	Fields    []itemkinds.Field  `json:"fields"`
	Columns   []itemkinds.Column `json:"columns"`
	Settings  []itemkinds.Field  `json:"settings"`
	Defaults  map[string]any     `json:"defaults"`
	Config    map[string]any     `json:"config"`
	Ownership []ownershipView    `json:"ownership"`
	Counts    map[string]int     `json:"counts"`
	Routes    []itemkinds.Route  `json:"defaultRoutes"`
}

type ownershipView struct {
	Field   string         `json:"field"`
	Label   itemkinds.Text `json:"label"`
	Default string         `json:"default"`
	Value   string         `json:"value"`
}

func (s *service) kinds(ctx context.Context) ([]kindView, error) {
	cfg := s.settings(ctx)
	counts := map[string]map[string]int{}
	rows, err := s.db.QueryContext(ctx, `SELECT kind, status, count(*) FROM itemkind_profile GROUP BY kind, status`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var k, st string
		var n int
		if err := rows.Scan(&k, &st, &n); err != nil {
			rows.Close()
			return nil, err
		}
		if counts[k] == nil {
			counts[k] = map[string]int{}
		}
		counts[k][st] = n
	}
	rows.Close()
	out := []kindView{}
	for _, k := range itemkinds.List() {
		v := kindView{Key: k.Key, ItemType: k.ItemType, Label: k.Label, Icon: k.Icon, Order: k.Order, Schema: k.SchemaVersion, Enabled: cfg.KindEnabled(k),
			Fields: nonNilFields(k.Fields), Columns: k.Columns, Settings: nonNilFields(k.Settings), Defaults: k.Defaults, Config: cfg.KindConfig(k), Ownership: []ownershipView{},
			Counts: map[string]int{"ok": 0, "pending": 0, "deleted": 0}, Routes: k.Routes}
		if v.Columns == nil {
			v.Columns = []itemkinds.Column{}
		}
		if v.Routes == nil {
			v.Routes = []itemkinds.Route{}
		}
		for st, n := range counts[k.Key] {
			v.Counts[st] = n
		}
		for _, o := range k.Ownership {
			v.Ownership = append(v.Ownership, ownershipView{Field: o.Field, Label: o.Label, Default: o.Default, Value: cfg.Owner(k, o.Field)})
		}
		out = append(out, v)
	}
	return out, nil
}

func nonNilFields(f []itemkinds.Field) []itemkinds.Field {
	if f == nil {
		return []itemkinds.Field{}
	}
	return f
}

func (s *service) kindList(w http.ResponseWriter, r *http.Request) {
	list, err := s.kinds(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"kinds": list})
}

func (s *service) overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.kinds(ctx)
	if err != nil {
		s.fail(w, err)
		return
	}
	var routes, entries int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM itemkind_route WHERE enabled`).Scan(&routes)
	if ok, _ := tableExists(ctx, s.db, "assets_entry"); ok {
		_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM assets_entry`).Scan(&entries)
	}
	runs, err := s.runs(ctx, 5)
	if err != nil {
		s.fail(w, err)
		return
	}
	cfg := s.settings(ctx)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"kinds": list, "routes": routes, "entries": entries, "routing": cfg.Routing, "runs": runs})
}

func tableExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var ok bool
	err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, name).Scan(&ok)
	return ok, err
}

func (s *service) kindOf(r *http.Request) (*itemkinds.Kind, error) {
	k := itemkinds.Get(pluginkit.PathVar(r, "kind"))
	if k == nil {
		return nil, pluginkit.Missing("KIND_NOT_FOUND")
	}
	return k, nil
}

func (s *service) itemList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	k, err := s.kindOf(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	page, limit := pluginkit.Page(r, 25, 200)
	q := r.URL.Query()
	args := &pluginkit.Args{}
	where := []string{"(p.kind=" + args.Add(k.Key) + " OR (p.item_id IS NULL AND c.key=" + args.Add(k.ItemType) + "))"}
	if text := strings.TrimSpace(q.Get("q")); text != "" {
		pat := args.Add(pluginkit.LikePattern(text))
		where = append(where, "(i.key ILIKE "+pat+" OR i.name::text ILIKE "+pat+" OR COALESCE(p.source_asset_id,'') ILIKE "+pat+" OR COALESCE(i.parent_key,'') ILIKE "+pat+")")
	}
	switch st := q.Get("status"); st {
	case "ok", "pending", "deleted":
		where = append(where, "p.status="+args.Add(st))
	case "unlinked":
		where = append(where, "p.item_id IS NULL")
	}
	if parent := strings.TrimSpace(q.Get("parent")); parent != "" {
		where = append(where, "i.parent_key="+args.Add(parent))
	}
	base := ` FROM economy_item i JOIN base_category c ON c.id=i.type_id LEFT JOIN itemkind_profile p ON p.item_id=i.id WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*)`+base, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	off := args.Add((page - 1) * limit)
	lim := args.Add(limit)
	rows, err := s.db.QueryContext(ctx, `SELECT i.id::text, i.key, i.name::text, COALESCE(i.price::text,'null'), i.unlock_level, i.is_active, COALESCE(i.parent_key,''), COALESCE(i.requires_key,''),
		COALESCE(p.status,'unlinked'), COALESCE(p.reason,''), COALESCE(p.data::text,'{}'), COALESCE(p.source_asset_id,''), COALESCE(p.source_version,0), p.updated_at`+base+
		` ORDER BY i.sort, i.key LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, key, name, price, parent, requires, status, reason, data, src string
		var unlock sql.NullInt64
		var active bool
		var version int
		var at sql.NullTime
		if err := rows.Scan(&id, &key, &name, &price, &unlock, &active, &parent, &requires, &status, &reason, &data, &src, &version, &at); err != nil {
			s.fail(w, err)
			return
		}
		item := map[string]any{"id": id, "key": key, "name": rawJSON(name), "price": rawJSON(price), "isActive": active, "parentKey": parent, "requiresKey": requires,
			"status": status, "reason": reason, "data": rawJSON(data), "sourceAssetId": src, "sourceVersion": version}
		if unlock.Valid {
			item["unlockLevel"] = unlock.Int64
		}
		if at.Valid {
			item["updatedAt"] = at.Time.UTC().Format(time.RFC3339)
		}
		items = append(items, item)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func rawJSON(s string) json.RawMessage {
	if strings.TrimSpace(s) == "" {
		return json.RawMessage("null")
	}
	return json.RawMessage(s)
}

func (s *service) itemGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("ITEM_NOT_FOUND"))
		return
	}
	body, profile, err := itemkinds.Snapshot(ctx, s.db, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if body == nil {
		s.fail(w, pluginkit.Missing("ITEM_NOT_FOUND"))
		return
	}
	unlock, hasUnlock, err := itemkinds.UnlockPrices(ctx, s.db, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	out := map[string]any{"id": id, "item": body, "profile": profile, "unlockPrices": unlock, "hasUnlockPrices": hasUnlock}
	if profile != nil && profile.SourceAssetID != "" {
		if ok, _ := tableExists(ctx, s.db, "assets_entry"); ok {
			var typ, name string
			var version int
			if err := s.db.QueryRowContext(ctx, `SELECT type, name, version FROM assets_entry WHERE asset_id=$1`, profile.SourceAssetID).Scan(&typ, &name, &version); err == nil {
				out["source"] = map[string]any{"assetId": profile.SourceAssetID, "type": typ, "name": name, "version": version}
			}
		}
	}
	if profile != nil {
		if k := itemkinds.Get(profile.Kind); k != nil && k.View != nil {
			cfg := s.settings(ctx)
			if extra, err := k.View(ctx, itemkinds.ViewInput{Q: s.db, Profiles: []itemkinds.Profile{*profile}, Config: cfg.KindConfig(k)}); err == nil {
				out["view"] = extra[id]
			}
		}
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

type itemIn struct {
	Key         string            `json:"key"`
	Name        map[string]string `json:"name"`
	Price       *itemkinds.Price  `json:"price"`
	NoPrice     bool              `json:"noPrice"`
	UnlockLevel *int              `json:"unlockLevel"`
	IsActive    *bool             `json:"isActive"`
	ParentKey   *string           `json:"parentKey"`
	RequiresKey *string           `json:"requiresKey"`
	Data        map[string]any    `json:"data"`
}

func (in itemIn) patch(body map[string]any) {
	if len(in.Name) > 0 {
		body["name"] = in.Name
	}
	if in.NoPrice {
		body["price"] = nil
	} else if in.Price != nil {
		body["price"] = in.Price
	}
	if in.UnlockLevel != nil {
		if *in.UnlockLevel >= 1 {
			body["unlockLevel"] = *in.UnlockLevel
		} else {
			body["unlockLevel"] = nil
		}
	}
	if in.IsActive != nil {
		body["isActive"] = *in.IsActive
	}
	if in.ParentKey != nil {
		body["parentKey"] = strings.TrimSpace(*in.ParentKey)
	}
	if in.RequiresKey != nil {
		body["requiresKey"] = strings.TrimSpace(*in.RequiresKey)
	}
}

func checkData(k *itemkinds.Kind, data map[string]any) error {
	if data == nil {
		return nil
	}
	if err := itemkinds.ValidateData(k.Fields, data); err != nil {
		return err
	}
	if k.Validate != nil {
		if err := k.Validate(data); err != nil {
			return pluginkit.Bad("INVALID_PROFILE", "data", err.Error())
		}
	}
	return nil
}

func (s *service) itemPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("ITEM_NOT_FOUND"))
		return
	}
	var in itemIn
	if err := pluginkit.ReadBody(r, &in, 1<<20); err != nil {
		s.fail(w, err)
		return
	}
	err := pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		body, profile, err := itemkinds.Snapshot(ctx, tx, id)
		if err != nil {
			return err
		}
		if body == nil {
			return pluginkit.Missing("ITEM_NOT_FOUND")
		}
		kindKey := r.URL.Query().Get("kind")
		if profile != nil {
			kindKey = profile.Kind
		}
		k := itemkinds.Get(kindKey)
		if k == nil {
			return pluginkit.Bad("KIND_NOT_FOUND", "kind", "the item has no kind")
		}
		if err := checkData(k, in.Data); err != nil {
			return err
		}
		in.patch(body)
		if err := itemkinds.PatchItem(ctx, tx, id, false, body); err != nil {
			return err
		}
		if in.Data != nil {
			if err := itemkinds.SaveProfile(ctx, tx, id, k.Key, k.SchemaVersion, in.Data, actor(r)); err != nil {
				return err
			}
		} else if profile == nil {
			if err := itemkinds.SaveProfile(ctx, tx, id, k.Key, k.SchemaVersion, map[string]any{}, actor(r)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	s.itemGet(w, r)
}

func (s *service) itemCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	k, err := s.kindOf(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var in itemIn
	if err := pluginkit.ReadBody(r, &in, 1<<20); err != nil {
		s.fail(w, err)
		return
	}
	if err := checkData(k, in.Data); err != nil {
		s.fail(w, err)
		return
	}
	id := uuid.Must(uuid.NewV4()).String()
	err = pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		var typeID string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='item_type' AND key=$1`, k.ItemType).Scan(&typeID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return pluginkit.Bad("ITEM_TYPE_NOT_FOUND", "kind", "item type "+k.ItemType+" does not exist")
			}
			return err
		}
		name := in.Name
		if len(name) == 0 {
			name = map[string]string{}
		}
		body := map[string]any{"key": strings.TrimSpace(in.Key), "typeId": typeID, "name": name, "description": map[string]string{}, "tagIds": []string{}, "upgrades": []any{}, "services": []any{}, "isActive": true, "metadata": map[string]any{}}
		in.patch(body)
		if err := itemkinds.PatchItem(ctx, tx, id, true, body); err != nil {
			return err
		}
		data := in.Data
		if data == nil {
			data = map[string]any{}
		}
		return itemkinds.SaveProfile(ctx, tx, id, k.Key, k.SchemaVersion, data, actor(r))
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	r.SetPathValue("id", id)
	s.itemGet(w, r)
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := itemkinds.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	list, err := s.kinds(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": itemkinds.Defaults(), "kinds": list, "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := itemkinds.Defaults()
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if field, code := in.Validate(); code != "" {
		s.fail(w, pluginkit.Bad(code, field, ""))
		return
	}
	for key, cfg := range in.Kinds {
		k := itemkinds.Get(key)
		if k == nil {
			continue
		}
		data := map[string]any{}
		for kk, v := range cfg {
			if kk != "enabled" {
				data[kk] = v
			}
		}
		if err := itemkinds.ValidateData(k.Settings, data); err != nil {
			var fe *itemkinds.FieldError
			if errors.As(err, &fe) {
				s.fail(w, pluginkit.Bad("INVALID_SETTING", "kinds."+key+"."+fe.Path, fe.Message))
				return
			}
			s.fail(w, err)
			return
		}
	}
	if err := itemkinds.Save(r.Context(), s.db, in, actor(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.settingsGet(w, r)
}

func (s *service) routeList(w http.ResponseWriter, r *http.Request) {
	list, err := itemkinds.LoadRoutes(r.Context(), s.db, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	kinds := []string{}
	for _, k := range itemkinds.List() {
		kinds = append(kinds, k.Key)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"routes": list, "kinds": kinds, "roles": []string{itemkinds.RoleItem, itemkinds.RoleDependent, itemkinds.RoleIgnore}})
}

type routeIn struct {
	SourceType string         `json:"sourceType"`
	Match      map[string]any `json:"match"`
	Kind       string         `json:"kind"`
	Role       string         `json:"role"`
	Priority   *int           `json:"priority"`
	Enabled    *bool          `json:"enabled"`
	Note       string         `json:"note"`
}

func (in *routeIn) clean() error {
	in.SourceType = strings.TrimSpace(in.SourceType)
	in.Kind = strings.TrimSpace(in.Kind)
	if in.SourceType == "" || len(in.SourceType) > 255 {
		return pluginkit.Bad("INVALID_ROUTE", "sourceType", "the Unity class name is required")
	}
	switch in.Role {
	case itemkinds.RoleItem:
		if itemkinds.Get(in.Kind) == nil {
			return pluginkit.Bad("KIND_NOT_FOUND", "kind", "unknown kind")
		}
	case itemkinds.RoleDependent, itemkinds.RoleIgnore:
		in.Kind = ""
	default:
		return pluginkit.Bad("INVALID_ROUTE", "role", "role must be item, dependent or ignore")
	}
	if in.Match == nil {
		in.Match = map[string]any{}
	}
	if len(in.Match) > 10 {
		return pluginkit.Bad("INVALID_ROUTE", "match", "at most 10 conditions")
	}
	if in.Priority != nil && (*in.Priority < 0 || *in.Priority > 100000) {
		return pluginkit.Bad("INVALID_ROUTE", "priority", "priority must be 0..100000")
	}
	if len(in.Note) > 500 {
		return pluginkit.Bad("INVALID_ROUTE", "note", "at most 500 characters")
	}
	return nil
}

func (s *service) routeCreate(w http.ResponseWriter, r *http.Request) {
	var in routeIn
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if err := in.clean(); err != nil {
		s.fail(w, err)
		return
	}
	prio, enabled := 100, true
	if in.Priority != nil {
		prio = *in.Priority
	}
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	match, _ := json.Marshal(in.Match)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO itemkind_route (source_type, match, kind, role, priority, enabled, note, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		in.SourceType, string(match), in.Kind, in.Role, prio, enabled, in.Note, actor(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.routeList(w, r)
}

func routeID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(pluginkit.PathVar(r, "id"), 10, 64)
	if err != nil || id < 1 {
		return 0, pluginkit.Missing("ROUTE_NOT_FOUND")
	}
	return id, nil
}

func (s *service) routeUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := routeID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var in routeIn
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if err := in.clean(); err != nil {
		s.fail(w, err)
		return
	}
	match, _ := json.Marshal(in.Match)
	prio := 100
	if in.Priority != nil {
		prio = *in.Priority
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE itemkind_route SET source_type=$2, match=$3, kind=$4, role=$5, priority=$6, enabled=$7, note=$8, updated_by=$9, updated_at=now() WHERE id=$1`,
		id, in.SourceType, string(match), in.Kind, in.Role, prio, enabled, in.Note, actor(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("ROUTE_NOT_FOUND"))
		return
	}
	s.routeList(w, r)
}

func (s *service) routeDelete(w http.ResponseWriter, r *http.Request) {
	id, err := routeID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var origin string
	if err := s.db.QueryRowContext(r.Context(), `SELECT origin FROM itemkind_route WHERE id=$1`, id).Scan(&origin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.fail(w, pluginkit.Missing("ROUTE_NOT_FOUND"))
			return
		}
		s.fail(w, err)
		return
	}
	q := `DELETE FROM itemkind_route WHERE id=$1`
	if origin != "" {
		q = `UPDATE itemkind_route SET enabled=false, updated_by=$2, updated_at=now() WHERE id=$1`
	}
	args := []any{id}
	if origin != "" {
		args = append(args, actor(r))
	}
	if _, err := s.db.ExecContext(r.Context(), q, args...); err != nil {
		s.fail(w, err)
		return
	}
	s.routeList(w, r)
}

func (s *service) reroute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		DryRun bool   `json:"dryRun"`
		Force  bool   `json:"force"`
		Kind   string `json:"kind"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if in.Kind != "" && itemkinds.Get(in.Kind) == nil {
		s.fail(w, pluginkit.Bad("KIND_NOT_FOUND", "kind", "unknown kind"))
		return
	}
	if ok, err := tableExists(ctx, s.db, "assets_entry"); err != nil || !ok {
		s.fail(w, pluginkit.Bad("ASSETS_UNAVAILABLE", "", "the assets plugin has no stored entries"))
		return
	}
	cfg := s.settings(ctx)
	cfg.Routing.Enabled = true
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer func() { _ = tx.Rollback() }()
	who := actor(r)
	out, err := itemkinds.Run(ctx, tx, cfg, itemkinds.RunInput{All: true, Kind: in.Kind, DryRun: in.DryRun, Force: in.Force, Actor: who, Source: "reroute"})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !in.DryRun {
		if err := tx.Commit(); err != nil {
			s.fail(w, err)
			return
		}
	}
	itemkinds.SaveRun(ctx, s.db, cfg, "reroute", who, &out)
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) runs(ctx context.Context, limit int) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, source, dry_run, actor, summary::text, created_at FROM itemkind_run ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var src, who, sum string
		var dry bool
		var at time.Time
		if err := rows.Scan(&id, &src, &dry, &who, &sum, &at); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "source": src, "dryRun": dry, "actor": who, "summary": rawJSON(sum), "createdAt": at.UTC().Format(time.RFC3339)})
	}
	return out, rows.Err()
}

func (s *service) runList(w http.ResponseWriter, r *http.Request) {
	_, limit := pluginkit.Page(r, 20, 100)
	list, err := s.runs(r.Context(), limit)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"runs": list})
}

func (s *service) runGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(pluginkit.PathVar(r, "id"), 10, 64)
	if err != nil {
		s.fail(w, pluginkit.Missing("RUN_NOT_FOUND"))
		return
	}
	var src, who, sum, rows string
	var dry bool
	var at time.Time
	err = s.db.QueryRowContext(r.Context(), `SELECT source, dry_run, actor, summary::text, rows::text, created_at FROM itemkind_run WHERE id=$1`, id).Scan(&src, &dry, &who, &sum, &rows, &at)
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("RUN_NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "source": src, "dryRun": dry, "actor": who, "summary": rawJSON(sum), "rows": rawJSON(rows), "createdAt": at.UTC().Format(time.RFC3339)})
}
