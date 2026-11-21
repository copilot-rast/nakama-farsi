package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/events"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const catalogTTL = 5 * time.Second

type catalogCache struct {
	mu   sync.Mutex
	at   map[string]time.Time
	data map[string]*events.Catalog
}

func (c *catalogCache) get(game string) *events.Catalog {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at == nil || time.Since(c.at[game]) > catalogTTL {
		return nil
	}
	return c.data[game]
}

func (c *catalogCache) put(game string, cat *events.Catalog) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at == nil {
		c.at, c.data = map[string]time.Time{}, map[string]*events.Catalog{}
	}
	c.at[game], c.data[game] = time.Now(), cat
}

func (c *catalogCache) reset() {
	c.mu.Lock()
	c.at, c.data = nil, nil
	c.mu.Unlock()
}

func (s *service) catalog(ctx context.Context, game string) (*events.Catalog, error) {
	if c := s.cat.get(game); c != nil {
		return c, nil
	}
	cat := &events.Catalog{Groups: map[string]events.Group{}, Events: map[string]events.Def{}}
	rows, err := s.db.QueryContext(ctx, `SELECT key, closed, enabled FROM insights_tag_group WHERE game=$1`, game)
	if err != nil {
		return nil, err
	}
	closed := []string{}
	for rows.Next() {
		var g events.Group
		if err := rows.Scan(&g.Key, &g.Closed, &g.Enabled); err != nil {
			rows.Close()
			return nil, err
		}
		g.Values = map[string]bool{}
		cat.Groups[g.Key] = g
		if g.Closed {
			closed = append(closed, g.Key)
		}
	}
	rows.Close()
	if len(closed) > 0 {
		rows, err = s.db.QueryContext(ctx, `SELECT group_key, id FROM insights_tag_value WHERE game=$1 AND group_key = ANY($2::text[])`, game, closed)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var g, id string
			if err := rows.Scan(&g, &id); err != nil {
				rows.Close()
				return nil, err
			}
			cat.Groups[g].Values[id] = true
		}
		rows.Close()
	}
	rows, err = s.db.QueryContext(ctx, `SELECT name, enabled, path_groups, required_tags, allowed_tags, value_rule FROM insights_event_def WHERE game=$1`, game)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var d events.Def
		var pg, rq, al, vr []byte
		if err := rows.Scan(&d.Name, &d.Enabled, &pg, &rq, &al, &vr); err != nil {
			rows.Close()
			return nil, err
		}
		_ = json.Unmarshal(pg, &d.PathGroups)
		_ = json.Unmarshal(rq, &d.RequiredTags)
		_ = json.Unmarshal(al, &d.AllowedTags)
		_ = json.Unmarshal(vr, &d.Value)
		cat.Events[d.Name] = d
	}
	rows.Close()
	s.cat.put(game, cat)
	return cat, nil
}

type tagValueView struct {
	ID      string `json:"id"`
	LabelFa string `json:"labelFa"`
	LabelEn string `json:"labelEn"`
}

type tagGroupView struct {
	Key         string         `json:"key"`
	NameFa      string         `json:"nameFa"`
	NameEn      string         `json:"nameEn"`
	Description string         `json:"description"`
	Closed      bool           `json:"closed"`
	Enabled     bool           `json:"enabled"`
	Values      []tagValueView `json:"values"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type eventDefView struct {
	Name         string           `json:"name"`
	LabelFa      string           `json:"labelFa"`
	LabelEn      string           `json:"labelEn"`
	Description  string           `json:"description"`
	Enabled      bool             `json:"enabled"`
	PathGroups   []string         `json:"pathGroups"`
	RequiredTags []string         `json:"requiredTags"`
	AllowedTags  []string         `json:"allowedTags"`
	Value        events.ValueRule `json:"value"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

func gameParam(r *http.Request) (string, error) {
	g := strings.TrimSpace(r.URL.Query().Get("game"))
	if !events.GamePattern.MatchString(g) {
		return "", pluginkit.Bad("INVALID_GAME", "game", "")
	}
	return g, nil
}

func (s *service) catalogGet(w http.ResponseWriter, r *http.Request) {
	game, err := gameParam(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	ctx := r.Context()
	groups, err := s.groupViews(ctx, game, "")
	if err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT name, label_fa, label_en, description, enabled, path_groups, required_tags, allowed_tags, value_rule, updated_at FROM insights_event_def WHERE game=$1 ORDER BY name`, game)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	defs := []eventDefView{}
	for rows.Next() {
		var d eventDefView
		var pg, rq, al, vr []byte
		if err := rows.Scan(&d.Name, &d.LabelFa, &d.LabelEn, &d.Description, &d.Enabled, &pg, &rq, &al, &vr, &d.UpdatedAt); err != nil {
			s.fail(w, err)
			return
		}
		_ = json.Unmarshal(pg, &d.PathGroups)
		_ = json.Unmarshal(rq, &d.RequiredTags)
		_ = json.Unmarshal(al, &d.AllowedTags)
		_ = json.Unmarshal(vr, &d.Value)
		defs = append(defs, nonNilDef(d))
	}
	games := []string{}
	gr, err := s.db.QueryContext(ctx, `SELECT game FROM insights_tag_group UNION SELECT game FROM insights_event_def UNION SELECT game FROM insights_funnel ORDER BY 1`)
	if err == nil {
		for gr.Next() {
			var g string
			if gr.Scan(&g) == nil {
				games = append(games, g)
			}
		}
		gr.Close()
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"game": game, "groups": groups, "events": defs, "games": games})
}

func nonNilDef(d eventDefView) eventDefView {
	if d.PathGroups == nil {
		d.PathGroups = []string{}
	}
	if d.RequiredTags == nil {
		d.RequiredTags = []string{}
	}
	if d.AllowedTags == nil {
		d.AllowedTags = []string{}
	}
	return d
}

func (s *service) groupViews(ctx context.Context, game, key string) ([]tagGroupView, error) {
	q := `SELECT key, name_fa, name_en, description, closed, enabled, updated_at FROM insights_tag_group WHERE game=$1`
	args := []any{game}
	if key != "" {
		q += ` AND key=$2`
		args = append(args, key)
	}
	rows, err := s.db.QueryContext(ctx, q+` ORDER BY key`, args...)
	if err != nil {
		return nil, err
	}
	groups := []tagGroupView{}
	idx := map[string]int{}
	for rows.Next() {
		var g tagGroupView
		if err := rows.Scan(&g.Key, &g.NameFa, &g.NameEn, &g.Description, &g.Closed, &g.Enabled, &g.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		g.Values = []tagValueView{}
		idx[g.Key] = len(groups)
		groups = append(groups, g)
	}
	rows.Close()
	vq := `SELECT group_key, id, label_fa, label_en FROM insights_tag_value WHERE game=$1`
	if key != "" {
		vq += ` AND group_key=$2`
	}
	rows, err = s.db.QueryContext(ctx, vq+` ORDER BY group_key, id LIMIT 20000`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var g string
		var v tagValueView
		if err := rows.Scan(&g, &v.ID, &v.LabelFa, &v.LabelEn); err != nil {
			return nil, err
		}
		if i, ok := idx[g]; ok {
			groups[i].Values = append(groups[i].Values, v)
		}
	}
	return groups, rows.Err()
}

func label(v string, max int) string { return cleanLabel(v, max) }

func (s *service) groupPut(w http.ResponseWriter, r *http.Request) {
	game, err := gameParam(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	key := pluginkit.PathVar(r, "key")
	if !events.GroupPattern.MatchString(key) {
		s.fail(w, pluginkit.Bad("INVALID_KEY", "key", ""))
		return
	}
	var in tagGroupView
	in.Enabled = true
	if err := pluginkit.ReadBody(r, &in, 4<<20); err != nil {
		s.fail(w, err)
		return
	}
	if len(in.Values) > 20000 {
		s.fail(w, pluginkit.Bad("VALUE_OUT_OF_RANGE", "values", ""))
		return
	}
	seen := map[string]bool{}
	for i, v := range in.Values {
		if !events.ValidTagID(v.ID) || seen[v.ID] {
			s.fail(w, pluginkit.Bad("INVALID_TAG_VALUE", "values."+itoa(i)+".id", v.ID))
			return
		}
		seen[v.ID] = true
	}
	ctx := r.Context()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO insights_tag_group (game, key, name_fa, name_en, description, closed, enabled) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (game, key) DO UPDATE SET name_fa=EXCLUDED.name_fa, name_en=EXCLUDED.name_en, description=EXCLUDED.description, closed=EXCLUDED.closed, enabled=EXCLUDED.enabled, updated_at=now()`,
		game, key, label(in.NameFa, 80), label(in.NameEn, 80), label(in.Description, 500), in.Closed, in.Enabled); err != nil {
		s.fail(w, err)
		return
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM insights_tag_value WHERE game=$1 AND group_key=$2`, game, key); err != nil {
		s.fail(w, err)
		return
	}
	if len(in.Values) > 0 {
		vals := make([]map[string]string, 0, len(in.Values))
		for _, v := range in.Values {
			vals = append(vals, map[string]string{"id": v.ID, "fa": label(v.LabelFa, 120), "en": label(v.LabelEn, 120)})
		}
		raw, _ := json.Marshal(vals)
		if _, err := tx.ExecContext(ctx, `INSERT INTO insights_tag_value (game, group_key, id, label_fa, label_en) SELECT $1, $2, x.id, x.fa, x.en FROM jsonb_to_recordset($3::jsonb) AS x(id text, fa text, en text)`, game, key, raw); err != nil {
			s.fail(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		s.fail(w, err)
		return
	}
	s.cat.reset()
	out, err := s.groupViews(ctx, game, key)
	if err != nil || len(out) == 0 {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out[0])
}

func (s *service) groupDelete(w http.ResponseWriter, r *http.Request) {
	game, err := gameParam(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM insights_tag_group WHERE game=$1 AND key=$2`, game, pluginkit.PathVar(r, "key"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	s.cat.reset()
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func cleanGroups(list []string, field string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, g := range list {
		g = strings.TrimSpace(g)
		if !events.GroupPattern.MatchString(g) || seen[g] {
			return nil, pluginkit.Bad("INVALID_KEY", field, g)
		}
		seen[g] = true
		out = append(out, g)
	}
	if len(out) > 16 {
		return nil, pluginkit.Bad("VALUE_OUT_OF_RANGE", field, "")
	}
	return out, nil
}

func (s *service) eventDefPut(w http.ResponseWriter, r *http.Request) {
	game, err := gameParam(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	name := pluginkit.PathVar(r, "name")
	cfg := s.settings(r.Context())
	if !events.ValidName(name, cfg.Events.MaxNameLength) {
		s.fail(w, pluginkit.Bad("INVALID_EVENT_NAME", "name", ""))
		return
	}
	var in eventDefView
	in.Enabled = true
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	if in.PathGroups, err = cleanGroups(in.PathGroups, "pathGroups"); err != nil {
		s.fail(w, err)
		return
	}
	if len(in.PathGroups) > events.MaxPathParts-1 {
		s.fail(w, pluginkit.Bad("VALUE_OUT_OF_RANGE", "pathGroups", ""))
		return
	}
	if in.RequiredTags, err = cleanGroups(in.RequiredTags, "requiredTags"); err != nil {
		s.fail(w, err)
		return
	}
	if in.AllowedTags, err = cleanGroups(in.AllowedTags, "allowedTags"); err != nil {
		s.fail(w, err)
		return
	}
	v := in.Value
	if v.Min != nil && v.Max != nil && *v.Min > *v.Max {
		s.fail(w, pluginkit.Bad("VALUE_OUT_OF_RANGE", "value.max", ""))
		return
	}
	v.Unit = label(v.Unit, 24)
	pg, _ := json.Marshal(in.PathGroups)
	rq, _ := json.Marshal(in.RequiredTags)
	al, _ := json.Marshal(in.AllowedTags)
	vr, _ := json.Marshal(v)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO insights_event_def (game, name, label_fa, label_en, description, enabled, path_groups, required_tags, allowed_tags, value_rule)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT (game, name) DO UPDATE SET label_fa=EXCLUDED.label_fa, label_en=EXCLUDED.label_en, description=EXCLUDED.description,
		enabled=EXCLUDED.enabled, path_groups=EXCLUDED.path_groups, required_tags=EXCLUDED.required_tags, allowed_tags=EXCLUDED.allowed_tags, value_rule=EXCLUDED.value_rule, updated_at=now()`,
		game, name, label(in.LabelFa, 120), label(in.LabelEn, 120), label(in.Description, 500), in.Enabled, pg, rq, al, vr); err != nil {
		s.fail(w, err)
		return
	}
	s.cat.reset()
	in.Name, in.Value, in.UpdatedAt = name, v, time.Now().UTC()
	pluginkit.WriteJSON(w, http.StatusOK, nonNilDef(in))
}

func (s *service) eventDefDelete(w http.ResponseWriter, r *http.Request) {
	game, err := gameParam(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM insights_event_def WHERE game=$1 AND name=$2`, game, pluginkit.PathVar(r, "name"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	s.cat.reset()
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) clientCatalog(ctx context.Context, game string) map[string]any {
	out := map[string]any{"events": []any{}, "groups": []any{}}
	rows, err := s.db.QueryContext(ctx, `SELECT name, path_groups, required_tags FROM insights_event_def WHERE game=$1 AND enabled ORDER BY name LIMIT 1000`, game)
	if err != nil {
		return out
	}
	evs := []map[string]any{}
	for rows.Next() {
		var name string
		var pg, rq []byte
		if rows.Scan(&name, &pg, &rq) == nil {
			var p, q []string
			_ = json.Unmarshal(pg, &p)
			_ = json.Unmarshal(rq, &q)
			evs = append(evs, map[string]any{"name": name, "pathGroups": nonNil(p), "requiredTags": nonNil(q)})
		}
	}
	rows.Close()
	groups := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT key, closed FROM insights_tag_group WHERE game=$1 AND enabled ORDER BY key LIMIT 1000`, game)
	if err == nil {
		for rows.Next() {
			var k string
			var c bool
			if rows.Scan(&k, &c) == nil {
				groups = append(groups, map[string]any{"key": k, "closed": c})
			}
		}
		rows.Close()
	}
	out["events"], out["groups"] = evs, groups
	return out
}

func nonNil(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}

func itoa(i int) string { return strconv.Itoa(i) }
