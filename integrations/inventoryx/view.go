package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

// itemRow is what this module reads of an economy item.
type itemRow struct {
	ID          string
	Key         string
	Active      bool
	Starter     bool
	UnlockLevel int
	TableID     string
	ForSale     bool
	Needs       string
}

const itemSelect = `SELECT i.id::text, i.key, i.is_active, i.is_starter, COALESCE(i.unlock_level,0), COALESCE(i.level_table_id::text,''), i.price IS NOT NULL,
	COALESCE(NULLIF(i.requires_key,''), NULLIF(i.parent_key,''), '') FROM economy_item i`

func scanItems(rows *sql.Rows) ([]itemRow, error) {
	defer rows.Close()
	out := []itemRow{}
	for rows.Next() {
		var it itemRow
		if err := rows.Scan(&it.ID, &it.Key, &it.Active, &it.Starter, &it.UnlockLevel, &it.TableID, &it.ForSale, &it.Needs); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// playerState holds what a player has of the given items: ownership, early unlocks and track levels.
type playerState struct {
	owned    map[string]bool
	unlocked map[string]bool
	levels   map[string]map[string]int
	keys     map[string]bool
}

func (s *service) loadPlayer(ctx context.Context, userID string, ids []string) (playerState, error) {
	st := playerState{owned: map[string]bool{}, unlocked: map[string]bool{}, levels: map[string]map[string]int{}, keys: map[string]bool{}}
	rows, err := s.db.QueryContext(ctx, `SELECT u.item_id::text, i.key FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1`, userID)
	if err != nil {
		return st, err
	}
	for rows.Next() {
		var id, key string
		if err := rows.Scan(&id, &key); err != nil {
			rows.Close()
			return st, err
		}
		st.owned[id] = true
		st.keys[key] = true
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT item_id::text FROM inventoryx_user_unlock WHERE user_id=$1 AND item_id::text = ANY($2)`, userID, ids)
	if err != nil {
		return st, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return st, err
		}
		st.unlocked[id] = true
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT item_id::text, track, level FROM inventoryx_user_track WHERE user_id=$1 AND item_id::text = ANY($2)`, userID, ids)
	if err != nil {
		return st, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, tr string
		var lv int
		if err := rows.Scan(&id, &tr, &lv); err != nil {
			return st, err
		}
		if st.levels[id] == nil {
			st.levels[id] = map[string]int{}
		}
		st.levels[id][tr] = lv
	}
	return st, rows.Err()
}

type levelFn func(tableID string) int

func (s *service) levelReader(ctx context.Context, userID string) levelFn {
	cache := map[string]int{}
	return func(tableID string) int {
		if v, ok := cache[tableID]; ok {
			return v
		}
		v, err := s.level(ctx, userID, tableID)
		if err != nil {
			v = 0
		}
		cache[tableID] = v
		return v
	}
}

func pricesView(list []price) []price {
	if list == nil {
		return []price{}
	}
	return list
}

// trackView is one upgrade track as the client sees it.
func trackView(t track, levels map[string]int, owned bool, playerLevel int, langs *localized.Set, lang string) map[string]any {
	level := levels[t.Key]
	if level > len(t.Levels) {
		level = len(t.Levels)
	}
	v := map[string]any{"key": t.Key, "name": t.Key, "currencyKey": t.CurrencyKey, "unit": t.Unit, "level": level, "maxLevel": len(t.Levels),
		"value": t.valueAt(level), "base": t.Base, "nextValue": nil, "nextPrice": nil, "nextMinPlayerLevel": 0, "nextRequires": []map[string]any{}, "canUpgrade": false}
	if langs != nil {
		if n := langs.Resolve(t.Name, lang); n != "" {
			v["name"] = n
		}
	}
	if level < len(t.Levels) {
		s := t.Levels[level]
		v["nextValue"] = t.valueAt(level + 1)
		v["nextPrice"] = price{CurrencyKey: t.CurrencyKey, Amount: s.Price}
		v["nextMinPlayerLevel"] = s.MinPlayerLevel
		met := s.MinPlayerLevel <= playerLevel
		reqs := []map[string]any{}
		for _, r := range s.Requires {
			ok := levels[r.Track] >= r.Level
			met = met && ok
			reqs = append(reqs, map[string]any{"track": r.Track, "level": r.Level, "met": ok})
		}
		v["nextRequires"] = reqs
		v["canUpgrade"] = owned && met
	}
	return v
}

// itemView is what this module adds to one item for one player.
func itemView(it itemRow, c itemConfig, st playerState, level levelFn, langs *localized.Set, lang string) map[string]any {
	owned := st.owned[it.ID] || it.Starter
	early := st.unlocked[it.ID]
	pl := level(it.TableID)
	reached := it.UnlockLevel == 0 || pl >= it.UnlockLevel
	tracks := []map[string]any{}
	attrs := map[string]any{}
	levels := st.levels[it.ID]
	if levels == nil {
		levels = map[string]int{}
	}
	for _, t := range c.Tracks {
		tv := trackView(t, levels, owned, pl, langs, lang)
		tracks = append(tracks, tv)
		attrs[t.Key] = tv["value"]
	}
	return map[string]any{
		"unlockPrices":   pricesView(c.UnlockPrices),
		"unlockedEarly":  early,
		"unlocked":       owned || early || reached,
		"canUnlockEarly": !owned && !early && !reached && it.Active && len(c.UnlockPrices) > 0 && !it.Starter,
		"tracks":         tracks,
		"attributes":     attrs,
	}
}

// itemStates fills economy_inventory through economyhooks.
func (s *service) itemStates(ctx context.Context, userID, lang string, ids []string) (map[string]economyhooks.ItemState, error) {
	rows, err := s.db.QueryContext(ctx, itemSelect+` WHERE i.id::text = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	items, err := scanItems(rows)
	if err != nil {
		return nil, err
	}
	configs, err := loadConfigs(ctx, s.db, ids)
	if err != nil {
		return nil, err
	}
	st, err := s.loadPlayer(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	var langs *localized.Set
	if lang != "" {
		langs, lang, _ = localized.ForUser(ctx, s.db, userID, lang)
	}
	level := s.levelReader(ctx, userID)
	out := map[string]economyhooks.ItemState{}
	for _, it := range items {
		c := configs[it.ID]
		state := economyhooks.ItemState{Unlocked: st.unlocked[it.ID]}
		if len(c.UnlockPrices) > 0 {
			state.LockMeta = map[string]any{"unlockPrices": c.UnlockPrices, "unlockRpc": "inventoryx_unlock"}
		}
		if lang != "" {
			state.Fields = itemView(it, c, st, level, langs, lang)
		}
		out[it.ID] = state
	}
	return out, nil
}

func cleanKeys(keys []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" && !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}
