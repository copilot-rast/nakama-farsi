package main

import (
	"context"
	"sort"
	"time"
)

type purchaseCount struct {
	ProductID       string    `json:"productId"`
	Key             string    `json:"key"`
	Count           int64     `json:"count"`
	LastPurchasedAt time.Time `json:"lastPurchasedAt"`
}

func sortPurchases(items []purchaseCount) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].LastPurchasedAt.Equal(items[j].LastPurchasedAt) {
			return items[i].Key < items[j].Key
		}
		return items[i].LastPurchasedAt.After(items[j].LastPurchasedAt)
	})
}

func prizeItemKeys(items []prizeItem) []string {
	out := []string{}
	for _, it := range items {
		if it.Type == "item" {
			out = append(out, it.Key)
		}
	}
	return out
}

func (s *service) ownedItemKeys(ctx context.Context, userID string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT i.key FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1`, userID)
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

type productState struct {
	Level       int
	LevelLocked bool
	OverMax     bool
	Count       int64
	Owned       bool
	Missing     string
}

func (s *service) productStateFor(ctx context.Context, p productView, prize prizeRow, lc *levelCache, counts map[string]purchaseCount, ownedKeys map[string]bool) (productState, error) {
	var st productState
	if p.MinLevel > 0 || p.MaxLevel != nil {
		table := ""
		if p.LevelTableID != nil {
			table = *p.LevelTableID
		}
		lv, err := lc.level(ctx, table)
		if err != nil {
			return st, err
		}
		st.Level = lv
		st.LevelLocked = lv < p.MinLevel
		st.OverMax = p.MaxLevel != nil && lv > *p.MaxLevel
	}
	st.Count = counts[p.ID].Count
	keys := prizeItemKeys(prize.Items)
	if p.CheckContents && len(keys) > 0 {
		all := true
		given := map[string]bool{}
		for _, k := range keys {
			given[k] = true
			if !ownedKeys[k] {
				all = false
			}
		}
		st.Owned = all
		if !all {
			defs, err := loadItems(ctx, s.db, `WHERE t.key = ANY($1)`, keys)
			if err != nil {
				return st, err
			}
			for _, d := range defs {
				if need := d.needs(); need != "" && !ownedKeys[need] && !given[need] {
					st.Missing = need
					break
				}
			}
		}
	}
	return st, nil
}

func (s *service) checkProductRules(ctx context.Context, userID string, p productView, prize prizeRow) error {
	counts, err := s.purchaseCounts(ctx, userID)
	if err != nil {
		return err
	}
	owned, err := s.ownedItemKeys(ctx, userID)
	if err != nil {
		return err
	}
	st, err := s.productStateFor(ctx, p, prize, newLevelCache(s.db, userID), counts, owned)
	if err != nil {
		return err
	}
	if st.LevelLocked {
		return conflict("LEVEL_TOO_LOW", "your level is too low for this product")
	}
	if st.OverMax {
		return conflict("LEVEL_TOO_HIGH", "this product is only for players up to a lower level")
	}
	if p.IsOneTime && st.Count > 0 {
		return conflict("PRODUCT_ALREADY_BOUGHT", "this product can be bought only once")
	}
	if st.Owned {
		return conflict("PRIZE_ALREADY_OWNED", "you already own everything this product gives")
	}
	if st.Missing != "" {
		return conflict("REQUIRES_ITEM", "you must own "+st.Missing+" first")
	}
	return nil
}

func productOnceKey(p productView, userID string) any {
	if !p.IsOneTime {
		return nil
	}
	return userID + ":" + p.ID
}
