package configtransfer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const missingPrefix = "missing:"

type tableInfo struct {
	t       Table
	cols    []string
	types   map[string]string
	export  []string
	exportS map[string]bool
	uuidID  bool
	paths   map[int][]seg
}

type refIndex struct {
	byID  map[string]string
	byKey map[string]string
}

type session struct {
	q       Querier
	infos   map[string]*tableInfo
	indexes map[string]*refIndex
}

func newSession(q Querier) *session {
	return &session{q: q, infos: map[string]*tableInfo{}, indexes: map[string]*refIndex{}}
}

func quote(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

var skipAlways = map[string]bool{"created_at": true, "updated_at": true, "updated_by": true}

func (s *session) info(ctx context.Context, t Table) (*tableInfo, error) {
	if ti, ok := s.infos[t.Name]; ok {
		ti.t = t
		return ti, nil
	}
	rows, err := s.q.QueryContext(ctx, `SELECT column_name, udt_name FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 ORDER BY ordinal_position`, t.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ti := &tableInfo{t: t, types: map[string]string{}, exportS: map[string]bool{}, paths: map[int][]seg{}}
	for rows.Next() {
		var c, typ string
		if err := rows.Scan(&c, &typ); err != nil {
			return nil, err
		}
		ti.cols = append(ti.cols, c)
		ti.types[c] = typ
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ti.cols) == 0 {
		return nil, nil
	}
	inKey := map[string]bool{}
	for _, k := range t.Key {
		inKey[k] = true
		if _, ok := ti.types[k]; !ok {
			return nil, fmt.Errorf("configtransfer: table %s has no key column %s", t.Name, k)
		}
	}
	drop := map[string]bool{}
	for _, c := range t.Skip {
		drop[c] = true
	}
	for _, c := range t.Secret {
		drop[c] = true
	}
	ti.uuidID = ti.types["id"] == "uuid" && !inKey["id"]
	for _, c := range ti.cols {
		if inKey[c] {
			ti.export = append(ti.export, c)
			continue
		}
		if drop[c] || skipAlways[c] || (c == "id" && !inKey[c]) || SecretName(c) {
			continue
		}
		ti.export = append(ti.export, c)
	}
	for _, c := range ti.export {
		ti.exportS[c] = true
	}
	for i, r := range t.Refs {
		p, err := parsePath(r.Path)
		if err != nil {
			return nil, err
		}
		ti.paths[i] = p
	}
	s.infos[t.Name] = ti
	return ti, nil
}

func (ti *tableInfo) isJSON(col string) bool {
	typ := ti.types[col]
	return typ == "jsonb" || typ == "json"
}

func indexName(r Ref) string {
	return r.Table + "|" + r.byCol() + "|" + strings.Join(r.keyCols(), ",")
}

func (s *session) index(ctx context.Context, r Ref) (*refIndex, error) {
	name := indexName(r)
	if ix, ok := s.indexes[name]; ok {
		return ix, nil
	}
	ix := &refIndex{byID: map[string]string{}, byKey: map[string]string{}}
	var exists bool
	if err := s.q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1)`, r.Table).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		parts := make([]string, 0, len(r.keyCols()))
		for _, k := range r.keyCols() {
			parts = append(parts, "t."+quote(k))
		}
		rows, err := s.q.QueryContext(ctx, `SELECT (t.`+quote(r.byCol())+`)::text, jsonb_build_array(`+strings.Join(parts, ",")+`)::text FROM `+quote(r.Table)+` t`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id sql.NullString
			var raw string
			if err := rows.Scan(&id, &raw); err != nil {
				return nil, err
			}
			if !id.Valid {
				continue
			}
			v, err := decodeJSON([]byte(raw))
			if err != nil {
				return nil, err
			}
			parts, _ := v.([]any)
			label := keyLabel(parts)
			ix.byID[id.String] = label
			ix.byKey[label] = id.String
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	s.indexes[name] = ix
	return ix, nil
}

func (s *session) touched(table string) {
	for name := range s.indexes {
		if strings.HasPrefix(name, table+"|") {
			delete(s.indexes, name)
		}
	}
}

type tableRow struct {
	key    string
	ident  string
	values map[string]any
	raw    map[string]any
}

func (s *session) readTable(ctx context.Context, ti *tableInfo, export bool) ([]tableRow, error) {
	order := make([]string, 0, len(ti.t.Key))
	for _, k := range ti.t.Key {
		order = append(order, "t."+quote(k))
	}
	q := `SELECT to_jsonb(t)::text FROM ` + quote(ti.t.Name) + ` t`
	if export && strings.TrimSpace(ti.t.Where) != "" {
		q += ` WHERE ` + ti.t.Where
	}
	q += ` ORDER BY ` + strings.Join(order, ",")
	rows, err := s.q.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	var raws []map[string]any
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			rows.Close()
			return nil, err
		}
		v, err := decodeJSON([]byte(raw))
		if err != nil {
			rows.Close()
			return nil, err
		}
		m, _ := v.(map[string]any)
		raws = append(raws, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]tableRow, 0, len(raws))
	for _, raw := range raws {
		vals := map[string]any{}
		for _, c := range ti.export {
			v := deepCopy(raw[c])
			if ti.isJSON(c) {
				v = Scrub(v)
			}
			vals[c] = v
		}
		for i, r := range ti.t.Refs {
			if _, ok := vals[r.Column]; !ok {
				continue
			}
			ix, err := s.index(ctx, r)
			if err != nil {
				return nil, err
			}
			vals[r.Column] = mapPath(vals[r.Column], ti.paths[i], func(v any, inArray bool) (any, bool) {
				id, _ := v.(string)
				if key, ok := ix.byID[id]; ok {
					return key, true
				}
				if inArray {
					return nil, false
				}
				return missingPrefix + id, true
			})
		}
		key, ident := ti.identity(vals)
		out = append(out, tableRow{key: key, ident: ident, values: vals, raw: raw})
	}
	return out, nil
}

func (ti *tableInfo) identity(vals map[string]any) (string, string) {
	parts := make([]any, 0, len(ti.t.Key))
	for _, k := range ti.t.Key {
		parts = append(parts, vals[k])
	}
	return keyLabel(parts), canon(parts)
}
