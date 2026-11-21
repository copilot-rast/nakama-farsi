package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
)

const (
	maxChestOpens = 20
	maxChestDepth = 3
)

type randomAmount struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

func parseRandomAmount(raw json.RawMessage) (randomAmount, bool) {
	var in struct {
		Min *float64 `json:"min"`
		Max *float64 `json:"max"`
	}
	if err := json.Unmarshal(raw, &in); err != nil || in.Min == nil || in.Max == nil {
		return randomAmount{}, false
	}
	lo, hi := *in.Min, *in.Max
	if lo != math.Trunc(lo) || hi != math.Trunc(hi) || lo < 0 || hi < lo || hi <= 0 || hi > 1e12 {
		return randomAmount{}, false
	}
	return randomAmount{Min: int64(lo), Max: int64(hi)}, true
}

func flattenPrizeItems(items []prizeItem) []prizeItem {
	out := make([]prizeItem, 0, len(items))
	for _, item := range items {
		if item.Type == "chest" {
			out = append(out, flattenPrizeItems(item.Rewards)...)
			continue
		}
		out = append(out, item)
	}
	return out
}

func needsResolve(items []prizeItem) bool {
	for _, item := range items {
		if item.Type == "chest" || (item.Type == "currency" && item.Action == "random") {
			return true
		}
	}
	return false
}

func (s *service) resolvePrizeItems(ctx context.Context, tx *sql.Tx, userID string, items []prizeItem, depth int) ([]prizeItem, error) {
	if !needsResolve(items) {
		return items, nil
	}
	out := make([]prizeItem, 0, len(items))
	for _, item := range items {
		switch {
		case item.Type == "currency" && item.Action == "random":
			r, ok := parseRandomAmount(item.Value)
			if !ok {
				return nil, badRequest("INVALID_VALUE", "items", "invalid random amount")
			}
			n, err := secureRange(r.Min, r.Max)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				continue
			}
			out = append(out, prizeItem{Type: "currency", Key: item.Key, Action: "add", Value: json.RawMessage(jsonInt(n))})
		case item.Type == "chest":
			opened, err := s.openChestItem(ctx, tx, userID, item, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, opened)
		default:
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *service) openChestItem(ctx context.Context, tx *sql.Tx, userID string, item prizeItem, depth int) (prizeItem, error) {
	if depth >= maxChestDepth {
		return prizeItem{}, conflict("CHEST_TOO_DEEP", "a chest cannot hold chests more than three levels deep")
	}
	g, err := scanChanceGame(tx.QueryRowContext(ctx, `SELECT `+chanceGameColumns+` FROM economy_chance_game WHERE key=$1`, item.Key))
	if errors.Is(err, sql.ErrNoRows) {
		return prizeItem{}, notFound("CHEST_NOT_FOUND")
	}
	if err != nil {
		return prizeItem{}, err
	}
	if g.PrizeType != "slots" {
		return prizeItem{}, notFound("CHEST_NOT_FOUND")
	}
	times, _ := int64FromJSON(item.Value)
	if times < 1 {
		times = 1
	}
	if times > maxChestOpens {
		times = maxChestOpens
	}
	rewards := []prizeItem{}
	for i := int64(0); i < times; i++ {
		outcome, _, err := s.rollSlots(ctx, tx, g, userID)
		if err != nil {
			return prizeItem{}, err
		}
		for _, o := range outcome {
			if o.Kind == "currency" {
				if o.Amount > 0 {
					rewards = append(rewards, prizeItem{Type: "currency", Key: o.CurrencyKey, Action: "add", Value: json.RawMessage(jsonInt(o.Amount))})
				}
				continue
			}
			if o.PrizeID == "" {
				continue
			}
			p, err := loadPrize(ctx, tx, o.PrizeID)
			if err != nil {
				if app, ok := asApp(err); ok && app.Code == "PRIZE_NOT_FOUND" {
					continue
				}
				return prizeItem{}, err
			}
			inner, err := s.resolvePrizeItems(ctx, tx, userID, p.Items, depth+1)
			if err != nil {
				return prizeItem{}, err
			}
			rewards = append(rewards, inner...)
		}
	}
	return prizeItem{Type: "chest", Key: item.Key, Action: "open", Value: json.RawMessage(jsonInt(times)), Rewards: rewards}, nil
}
