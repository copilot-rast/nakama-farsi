package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

var trackKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,31}$`)

const (
	maxAmount      = 1_000_000_000_000
	maxUnlockPrice = 5
	maxPlayerLevel = 10000
)

type price struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

type need struct {
	Track string `json:"track"`
	Level int    `json:"level"`
}

// step is one level of an upgrade track: what it costs, the attribute value it
// gives and what the player needs before buying it.
type step struct {
	Price          int64    `json:"price"`
	Value          *float64 `json:"value,omitempty"`
	MinPlayerLevel int      `json:"minPlayerLevel,omitempty"`
	Requires       []need   `json:"requires,omitempty"`
}

type track struct {
	Key         string         `json:"key"`
	Name        localized.Text `json:"name"`
	CurrencyKey string         `json:"currencyKey"`
	Base        float64        `json:"base"`
	Unit        string         `json:"unit,omitempty"`
	Levels      []step         `json:"levels"`
}

type itemConfig struct {
	UnlockPrices []price `json:"unlockPrices"`
	Tracks       []track `json:"tracks"`
}

func (c itemConfig) empty() bool { return len(c.UnlockPrices) == 0 && len(c.Tracks) == 0 }

func (c itemConfig) track(key string) (track, bool) {
	for _, t := range c.Tracks {
		if t.Key == key {
			return t, true
		}
	}
	return track{}, false
}

// hasValues reports whether any level names an attribute value; without values
// the attribute is the level number itself.
func (t track) hasValues() bool {
	for _, s := range t.Levels {
		if s.Value != nil {
			return true
		}
	}
	return false
}

// valueAt is the attribute value after level upgrades (0 = base).
func (t track) valueAt(level int) float64 {
	if !t.hasValues() {
		return float64(level)
	}
	v := t.Base
	for i := 0; i < level && i < len(t.Levels); i++ {
		if t.Levels[i].Value != nil {
			v = *t.Levels[i].Value
		}
	}
	return v
}

func currencyKeys(ctx context.Context, q interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}) (map[string]bool, error) {
	rows, err := q.QueryContext(ctx, `SELECT key FROM economy_currency`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		out[k] = true
	}
	return out, rows.Err()
}

func finite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) && math.Abs(f) < 1e15 }

// cleanConfig validates an item's unlock prices and tracks.
func cleanConfig(cfg settings.All, langs *localized.Set, currencies map[string]bool, in itemConfig) (itemConfig, error) {
	out := itemConfig{UnlockPrices: []price{}, Tracks: []track{}}
	if len(in.UnlockPrices) > maxUnlockPrice {
		return out, pluginkit.Bad("TOO_MANY_PRICES", "unlockPrices", "at most 5 unlock prices")
	}
	seenCur := map[string]bool{}
	for i, p := range in.UnlockPrices {
		field := "unlockPrices[" + strconv.Itoa(i) + "]"
		p.CurrencyKey = strings.TrimSpace(p.CurrencyKey)
		if !currencies[p.CurrencyKey] {
			return out, pluginkit.Bad("CURRENCY_NOT_FOUND", field, "currency does not exist")
		}
		if p.Amount <= 0 || p.Amount > maxAmount {
			return out, pluginkit.Bad("INVALID_PRICE", field, "price must be a positive whole number")
		}
		if seenCur[p.CurrencyKey] {
			return out, pluginkit.Bad("DUPLICATE_CURRENCY", field, "each currency once")
		}
		seenCur[p.CurrencyKey] = true
		out.UnlockPrices = append(out.UnlockPrices, p)
	}
	if len(in.Tracks) > cfg.Upgrade.MaxTracks {
		return out, pluginkit.Bad("TOO_MANY_TRACKS", "tracks", "at most "+strconv.Itoa(cfg.Upgrade.MaxTracks)+" upgrade tracks")
	}
	lengths := map[string]int{}
	for i, t := range in.Tracks {
		field := "tracks[" + strconv.Itoa(i) + "]"
		t.Key = strings.TrimSpace(t.Key)
		if !trackKeyPattern.MatchString(t.Key) {
			return out, pluginkit.Bad("INVALID_KEY", field+".key", "start with a letter; letters, digits, _ . - (max 32)")
		}
		if _, dup := lengths[t.Key]; dup {
			return out, pluginkit.Bad("DUPLICATE_TRACK", field+".key", "track keys must be unique")
		}
		lengths[t.Key] = len(t.Levels)
	}
	for i, t := range in.Tracks {
		field := "tracks[" + strconv.Itoa(i) + "]"
		t.Key = strings.TrimSpace(t.Key)
		if langs != nil {
			name, err := langs.Clean(field+".name", t.Name, false, 64)
			if err != nil {
				var le *localized.Error
				if errors.As(err, &le) {
					return out, pluginkit.Bad(le.Code, le.Field, le.Message)
				}
				return out, err
			}
			t.Name = name
		}
		if t.Name == nil {
			t.Name = localized.Text{}
		}
		t.CurrencyKey = strings.TrimSpace(t.CurrencyKey)
		if !currencies[t.CurrencyKey] {
			return out, pluginkit.Bad("CURRENCY_NOT_FOUND", field+".currencyKey", "currency does not exist")
		}
		t.Unit = strings.TrimSpace(t.Unit)
		if utf8.RuneCountInString(t.Unit) > 16 {
			return out, pluginkit.Bad("VALUE_OUT_OF_RANGE", field+".unit", "at most 16 characters")
		}
		if !finite(t.Base) {
			return out, pluginkit.Bad("INVALID_VALUE", field+".base", "base must be a number")
		}
		if len(t.Levels) == 0 {
			return out, pluginkit.Bad("LEVELS_REQUIRED", field+".levels", "a track needs at least one level")
		}
		if len(t.Levels) > cfg.Upgrade.MaxLevels {
			return out, pluginkit.Bad("TOO_MANY_LEVELS", field+".levels", "at most "+strconv.Itoa(cfg.Upgrade.MaxLevels)+" levels")
		}
		levels := []step{}
		for j, s := range t.Levels {
			sf := field + ".levels[" + strconv.Itoa(j) + "]"
			if s.Price < 0 || s.Price > maxAmount {
				return out, pluginkit.Bad("INVALID_PRICE", sf+".price", "price must be 0 or a positive whole number")
			}
			if s.Value != nil && !finite(*s.Value) {
				return out, pluginkit.Bad("INVALID_VALUE", sf+".value", "value must be a number")
			}
			if s.MinPlayerLevel < 0 || s.MinPlayerLevel > maxPlayerLevel {
				return out, pluginkit.Bad("INVALID_LEVEL", sf+".minPlayerLevel", "player level must be 0 to 10000")
			}
			reqs := []need{}
			for k, r := range s.Requires {
				rf := sf + ".requires[" + strconv.Itoa(k) + "]"
				r.Track = strings.TrimSpace(r.Track)
				n, ok := lengths[r.Track]
				if !ok || r.Track == t.Key {
					return out, pluginkit.Bad("INVALID_REQUIREMENT", rf+".track", "name another track of this item")
				}
				if r.Level < 1 || r.Level > n {
					return out, pluginkit.Bad("INVALID_REQUIREMENT", rf+".level", "level must be 1 to "+strconv.Itoa(n))
				}
				reqs = append(reqs, r)
			}
			if len(reqs) > 0 {
				s.Requires = reqs
			} else {
				s.Requires = nil
			}
			levels = append(levels, s)
		}
		t.Levels = levels
		out.Tracks = append(out.Tracks, t)
	}
	return out, nil
}

// importRow is one row of a bulk import. It accepts the tracks of this module
// and the legacy shape: upgrades [{key, name, currencyId, prices[]}] and
// unlockPrices [{currencyId, amount}].
type importRow struct {
	Key          string          `json:"key"`
	UnlockPrices []importPrice   `json:"unlockPrices"`
	Tracks       []track         `json:"tracks"`
	Upgrades     []importUpgrade `json:"upgrades"`
}

type importPrice struct {
	CurrencyKey string  `json:"currencyKey"`
	CurrencyID  string  `json:"currencyId"`
	Amount      float64 `json:"amount"`
}

type importUpgrade struct {
	Key         string          `json:"key"`
	Name        json.RawMessage `json:"name"`
	CurrencyKey string          `json:"currencyKey"`
	CurrencyID  string          `json:"currencyId"`
	Prices      []float64       `json:"prices"`
	Values      []float64       `json:"values"`
	Base        float64         `json:"base"`
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s = strings.TrimSpace(s); s != "" {
			return s
		}
	}
	return ""
}

// config turns an import row into an item config (before validation).
func (r importRow) config(defaultLang string) itemConfig {
	out := itemConfig{}
	for _, p := range r.UnlockPrices {
		out.UnlockPrices = append(out.UnlockPrices, price{CurrencyKey: firstNonEmpty(p.CurrencyKey, p.CurrencyID), Amount: int64(math.Round(p.Amount))})
	}
	out.Tracks = append(out.Tracks, r.Tracks...)
	for _, u := range r.Upgrades {
		t := track{Key: u.Key, CurrencyKey: firstNonEmpty(u.CurrencyKey, u.CurrencyID), Base: u.Base, Name: localized.Text{}}
		if len(u.Name) > 0 {
			var text map[string]string
			var plain string
			if json.Unmarshal(u.Name, &text) == nil {
				t.Name = text
			} else if json.Unmarshal(u.Name, &plain) == nil && strings.TrimSpace(plain) != "" && defaultLang != "" {
				t.Name = localized.Text{defaultLang: plain}
			}
		}
		for i, p := range u.Prices {
			s := step{Price: int64(math.Round(p))}
			if i < len(u.Values) {
				v := u.Values[i]
				s.Value = &v
			}
			t.Levels = append(t.Levels, s)
		}
		out.Tracks = append(out.Tracks, t)
	}
	return out
}

func loadConfigs(ctx context.Context, q interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}, ids []string) (map[string]itemConfig, error) {
	out := map[string]itemConfig{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := q.QueryContext(ctx, `SELECT item_id::text, unlock_prices::text, tracks::text FROM inventoryx_item WHERE item_id::text = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, up, tr string
		if err := rows.Scan(&id, &up, &tr); err != nil {
			return nil, err
		}
		var c itemConfig
		_ = json.Unmarshal([]byte(up), &c.UnlockPrices)
		_ = json.Unmarshal([]byte(tr), &c.Tracks)
		out[id] = c
	}
	return out, rows.Err()
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func saveConfig(ctx context.Context, db execer, itemID string, c itemConfig, by string) error {
	if c.empty() {
		_, err := db.ExecContext(ctx, `DELETE FROM inventoryx_item WHERE item_id=$1`, itemID)
		return err
	}
	up, _ := json.Marshal(c.UnlockPrices)
	tr, _ := json.Marshal(c.Tracks)
	_, err := db.ExecContext(ctx, `INSERT INTO inventoryx_item (item_id, unlock_prices, tracks, updated_by, updated_at) VALUES ($1,$2,$3,$4,now())
		ON CONFLICT (item_id) DO UPDATE SET unlock_prices=EXCLUDED.unlock_prices, tracks=EXCLUDED.tracks, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		itemID, string(up), string(tr), by)
	return err
}
