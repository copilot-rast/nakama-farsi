package itemkinds

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type Entry struct {
	AssetID    string
	SourceType string
	Folder     string
	Name       string
	Hash       string
	Version    int
	Data       map[string]any
	refs       []string
}

type currencyRow struct {
	Key    string
	CardID string
}

type Mapper struct {
	ctx        context.Context
	q          Querier
	settings   Settings
	loaded     bool
	entries    map[string]*Entry
	referrers  map[string][]string
	used       map[string]string
	currencies []currencyRow
	langs      *localized.Set
	routes     []Route
	kind       *Kind
	config     map[string]any
}

func newMapper(ctx context.Context, q Querier, s Settings, routes []Route) *Mapper {
	return &Mapper{ctx: ctx, q: q, settings: s, routes: routes, used: map[string]string{}}
}

func (m *Mapper) load() error {
	if m.loaded {
		return nil
	}
	m.entries = map[string]*Entry{}
	m.referrers = map[string][]string{}
	rows, err := m.q.QueryContext(m.ctx, `SELECT asset_id, type, folder, name, hash, version, data::text FROM assets_entry`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Entry{}
		var raw string
		if err := rows.Scan(&e.AssetID, &e.SourceType, &e.Folder, &e.Name, &e.Hash, &e.Version, &raw); err != nil {
			return err
		}
		e.Data = map[string]any{}
		_ = json.Unmarshal([]byte(raw), &e.Data)
		e.refs = collectRefs(e.Data, nil)
		m.entries[e.AssetID] = e
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for id, e := range m.entries {
		seen := map[string]bool{}
		for _, r := range e.refs {
			if !seen[r] {
				seen[r] = true
				m.referrers[r] = append(m.referrers[r], id)
			}
		}
	}
	for k := range m.referrers {
		sort.Strings(m.referrers[k])
	}
	m.loaded = true
	return nil
}

func collectRefs(v any, out []string) []string {
	switch t := v.(type) {
	case map[string]any:
		if id := RefID(t); id != "" {
			return append(out, id)
		}
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out = collectRefs(t[k], out)
		}
	case []any:
		for _, x := range t {
			out = collectRefs(x, out)
		}
	}
	return out
}

func RefID(v any) string {
	obj, ok := v.(map[string]any)
	if !ok || len(obj) != 1 {
		return ""
	}
	id, _ := obj["$asset"].(string)
	return id
}

func (m *Mapper) Entry(id string) *Entry {
	if id == "" || m.load() != nil {
		return nil
	}
	e := m.entries[id]
	if e != nil {
		m.used[e.AssetID] = e.Hash
	} else {
		m.used[id] = "missing"
	}
	return e
}

func (m *Mapper) Ref(v any) (*Entry, string) {
	id := RefID(v)
	if id == "" {
		return nil, ""
	}
	e := m.Entry(id)
	if e == nil {
		return nil, id
	}
	return e, id
}

func (m *Mapper) Referrers(id string, types ...string) []*Entry {
	if m.load() != nil {
		return nil
	}
	out := []*Entry{}
	for _, rid := range m.referrers[id] {
		e := m.entries[rid]
		if e == nil {
			continue
		}
		if len(types) > 0 && !containsFold(types, e.SourceType) {
			continue
		}
		m.used[e.AssetID] = e.Hash
		out = append(out, e)
	}
	return out
}

func (m *Mapper) KeyOf(e *Entry) string {
	if e == nil {
		return ""
	}
	r := m.routeFor(e)
	if r == nil || r.Role != RoleItem {
		return ""
	}
	k := Get(r.Kind)
	if k == nil || k.KeyOf == nil {
		return ""
	}
	return k.KeyOf(e)
}

func (m *Mapper) Config() map[string]any { return m.config }

func (m *Mapper) Langs() *localized.Set {
	if m.langs == nil {
		if l, err := localized.Load(m.ctx, m.q); err == nil {
			m.langs = l
		}
	}
	return m.langs
}

func (m *Mapper) CurrencyKey(e *Entry) string {
	if e == nil {
		return ""
	}
	if m.currencies == nil {
		m.currencies = []currencyRow{}
		rows, err := m.q.QueryContext(m.ctx, `SELECT key, COALESCE(metadata->>'cardId','') FROM economy_currency`)
		if err == nil {
			for rows.Next() {
				var c currencyRow
				if rows.Scan(&c.Key, &c.CardID) == nil {
					m.currencies = append(m.currencies, c)
				}
			}
			rows.Close()
		}
	}
	card := Str(e.Data["id"])
	for _, c := range m.currencies {
		if c.CardID != "" && (c.CardID == card || c.CardID == e.AssetID) {
			return c.Key
		}
	}
	name := strings.ToLower(e.Name)
	name = strings.TrimPrefix(strings.TrimPrefix(name, "currency_"), "currency-")
	for _, c := range m.currencies {
		if c.Key == name || c.Key == strings.ToLower(card) {
			return c.Key
		}
	}
	return ""
}

func (m *Mapper) routeFor(e *Entry) *Route {
	for i := range m.routes {
		r := &m.routes[i]
		if !r.Enabled || !strings.EqualFold(r.SourceType, e.SourceType) {
			continue
		}
		if m.matches(e, r.Match) {
			return r
		}
	}
	return nil
}

func (m *Mapper) matches(e *Entry, match map[string]any) bool {
	for path, want := range match {
		if path == "$folderPrefix" {
			if !strings.HasPrefix(strings.ToLower(e.Folder), strings.ToLower(Str(want))) {
				return false
			}
			continue
		}
		got := m.path(e, path)
		if !equalAny(got, want) {
			return false
		}
	}
	return true
}

func (m *Mapper) path(e *Entry, path string) any {
	var cur any = e
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			continue
		}
		switch c := cur.(type) {
		case *Entry:
			switch seg {
			case "$name":
				cur = c.Name
			case "$type":
				cur = c.SourceType
			case "$folder":
				cur = c.Folder
			case "$id":
				cur = c.AssetID
			default:
				cur = c.Data[seg]
			}
		case map[string]any:
			if seg == "$" {
				ref, _ := m.Ref(c)
				if ref == nil {
					return nil
				}
				cur = ref
				continue
			}
			cur = c[seg]
		default:
			return nil
		}
	}
	if en, ok := cur.(*Entry); ok {
		return en.Name
	}
	return cur
}

func equalAny(got, want any) bool {
	if list, ok := want.([]any); ok {
		for _, w := range list {
			if equalAny(got, w) {
				return true
			}
		}
		return false
	}
	if want == nil {
		return got == nil
	}
	return strings.EqualFold(Str(got), Str(want))
}

func (m *Mapper) reset(k *Kind, cfg map[string]any) {
	m.used = map[string]string{}
	m.kind = k
	m.config = cfg
}

func (m *Mapper) hash(e *Entry, k *Kind, cfg map[string]any) string {
	parts := []string{"self:" + e.AssetID + ":" + e.Hash, "schema:" + strconv.Itoa(k.SchemaVersion)}
	for id, h := range m.used {
		parts = append(parts, id+":"+h)
	}
	sort.Strings(parts)
	c, _ := json.Marshal(cfg)
	sum := sha256.Sum256([]byte(strings.Join(parts, "|") + "|" + string(c)))
	return hex.EncodeToString(sum[:])
}

func containsFold(list []string, v string) bool {
	for _, x := range list {
		if strings.EqualFold(x, v) {
			return true
		}
	}
	return false
}

func Str(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case float64:
		if t == math.Trunc(t) && math.Abs(t) < 1e15 {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case json.Number:
		return t.String()
	}
	return fmt.Sprint(v)
}

func Num(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	}
	return 0, false
}

func Int(v any) (int64, bool) {
	f, ok := Num(v)
	if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return int64(math.Round(f)), true
}
