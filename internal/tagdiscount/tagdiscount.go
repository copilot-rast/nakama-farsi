package tagdiscount

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

const (
	CombineBest = "best"
	CombineSum  = "sum"
	OwnStack    = "stack"
	OwnBest     = "best"
)

type Settings struct {
	Enabled         bool   `json:"enabled"`
	Combine         string `json:"combine"`
	MaxPercent      int    `json:"maxPercent"`
	WithOwnDiscount string `json:"withOwnDiscount"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
}

func Defaults() Settings {
	return Settings{Enabled: true, Combine: CombineBest, MaxPercent: 90, WithOwnDiscount: OwnStack}
}

func (s *Settings) Normalize() {
	if s.Combine != CombineSum {
		s.Combine = CombineBest
	}
	if s.WithOwnDiscount != OwnBest {
		s.WithOwnDiscount = OwnStack
	}
	if s.MaxPercent < 1 || s.MaxPercent > 100 {
		s.MaxPercent = 90
	}
}

func LoadSettings(ctx context.Context, q Queryer) (Settings, error) {
	st := Defaults()
	var raw string
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_at FROM economy_tag_discount_settings WHERE id=1`).Scan(&raw, &updated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return st, err
	}
	if err == nil {
		_ = json.Unmarshal([]byte(raw), &st)
		st.UpdatedAt = updated.UTC().Format(time.RFC3339)
	}
	st.Normalize()
	return st, nil
}

type Rule struct {
	ID          string
	Key         string
	Name        json.RawMessage
	Percent     int
	AllItems    bool
	ItemKeys    map[string]bool
	ItemTypes   map[string]bool
	AllProducts bool
	Products    map[string]bool
}

type Match struct {
	Percent int      `json:"percent"`
	Rules   []string `json:"rules"`
}

type Set struct {
	Settings Settings
	Rules    []Rule
}

func ForUser(ctx context.Context, q Queryer, userID string) (*Set, error) {
	st, err := LoadSettings(ctx, q)
	if err != nil {
		return nil, err
	}
	set := &Set{Settings: st}
	if !st.Enabled || userID == "" {
		return set, nil
	}
	rows, err := q.QueryContext(ctx, `SELECT d.id::text, d.key, d.name::text, d.percent, d.all_items, to_json(d.item_keys)::text, to_json(d.item_type_ids)::text, d.all_products, to_json(d.product_ids)::text
		FROM economy_tag_discount d
		WHERE d.is_active AND (d.starts_at IS NULL OR d.starts_at <= now()) AND (d.ends_at IS NULL OR d.ends_at > now())
		AND EXISTS (SELECT 1 FROM base_user_tag ut WHERE ut.user_id=$1::uuid AND ut.tag_id::text = ANY(d.tag_ids) AND (ut.expires_at IS NULL OR ut.expires_at > now()))
		ORDER BY d.sort, d.key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var r Rule
		var name, items, types, products string
		if err := rows.Scan(&r.ID, &r.Key, &name, &r.Percent, &r.AllItems, &items, &types, &r.AllProducts, &products); err != nil {
			return nil, err
		}
		r.Name = json.RawMessage(name)
		r.ItemKeys, r.ItemTypes, r.Products = set2(items), set2(types), set2(products)
		set.Rules = append(set.Rules, r)
	}
	return set, rows.Err()
}

func set2(raw string) map[string]bool {
	var list []string
	_ = json.Unmarshal([]byte(raw), &list)
	out := make(map[string]bool, len(list))
	for _, v := range list {
		out[v] = true
	}
	return out
}

func (s *Set) combine(pick func(Rule) bool) Match {
	m := Match{Rules: []string{}}
	if s == nil || !s.Settings.Enabled {
		return m
	}
	for _, r := range s.Rules {
		if !pick(r) {
			continue
		}
		m.Rules = append(m.Rules, r.Key)
		if s.Settings.Combine == CombineSum {
			m.Percent += r.Percent
		} else if r.Percent > m.Percent {
			m.Percent = r.Percent
		}
	}
	if m.Percent > s.Settings.MaxPercent {
		m.Percent = s.Settings.MaxPercent
	}
	return m
}

func (s *Set) Item(key, typeID string) Match {
	return s.combine(func(r Rule) bool { return r.AllItems || r.ItemKeys[key] || (typeID != "" && r.ItemTypes[typeID]) })
}

func (s *Set) Product(id string) Match {
	return s.combine(func(r Rule) bool { return r.AllProducts || r.Products[id] })
}

func (s *Set) WithOwn(own, tag int) int {
	if tag <= 0 {
		return own
	}
	if own <= 0 {
		return tag
	}
	if s != nil && s.Settings.WithOwnDiscount == OwnBest {
		if tag > own {
			return tag
		}
		return own
	}
	out := 100 - ((100-own)*(100-tag)+50)/100
	if out > 100 {
		out = 100
	}
	return out
}

func Apply(amount int64, percent int) int64 {
	if percent <= 0 {
		return amount
	}
	if percent >= 100 {
		return 0
	}
	return amount * int64(100-percent) / 100
}
