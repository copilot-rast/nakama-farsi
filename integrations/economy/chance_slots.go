package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

const chanceRerollCap = 1000

type chanceOutcome struct {
	Slot         int         `json:"slot"`
	Kind         string      `json:"kind"`
	CurrencyKey  string      `json:"currencyKey,omitempty"`
	Amount       int64       `json:"amount,omitempty"`
	PrizeID      string      `json:"prizeId,omitempty"`
	PrizeGroupID string      `json:"prizeGroupId,omitempty"`
	Skipped      bool        `json:"skipped,omitempty"`
	Rewards      []prizeItem `json:"rewards,omitempty"`
}

func slotKind(r chanceRow) string {
	switch {
	case r.Kind != "":
		return r.Kind
	case r.CurrencyID != "":
		return "currency"
	case r.PrizeGroupID != nil && *r.PrizeGroupID != "":
		return "prizeGroup"
	}
	return "prize"
}

func jackpotRow(rows []chanceRow) int {
	for i, r := range rows {
		if r.Jackpot {
			return i
		}
	}
	return -1
}

func secureRange(min, max int64) (int64, error) {
	if max <= min {
		return min, nil
	}
	n := max - min + 1
	if n > 1<<31-1 {
		n = 1<<31 - 1
	}
	v, err := secureInt(int(n))
	if err != nil {
		return 0, err
	}
	return min + int64(v), nil
}

func priceChanged(amount, expected int64) *appError {
	e := conflict("CHANCE_PRICE_CHANGED", "the price of this play has changed")
	e.Meta = map[string]any{"amount": amount, "expected": expected}
	return e
}

func ladderAmount(ladder []int64, n int) int64 {
	if len(ladder) == 0 {
		return 0
	}
	if n < 0 {
		n = 0
	}
	if n >= len(ladder) {
		n = len(ladder) - 1
	}
	return ladder[n]
}

func (s *service) chanceTicketAmount(ctx context.Context, q queryer, g *chanceGame, userID string, now time.Time) (int64, error) {
	if g.Price == nil {
		return 0, nil
	}
	if len(g.Ladder) == 0 {
		return g.Price.Amount, nil
	}
	start, _ := s.chanceDay(ctx, q, g, now)
	n, err := s.chanceTicketsSince(ctx, q, g.ID, userID, start)
	if err != nil {
		return 0, err
	}
	return ladderAmount(g.Ladder, n), nil
}

func (s *service) chanceTicketsSince(ctx context.Context, q queryer, gameID, userID string, start time.Time) (int, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM economy_chance_play WHERE game_id=$1 AND user_id=$2 AND mode='ticket' AND created_at >= $3`, gameID, userID, start).Scan(&n)
	return n, err
}

func (s *service) slotPrizeOpen(ctx context.Context, tx *sql.Tx, userID, prizeID string) (bool, error) {
	var repeatable bool
	if err := tx.QueryRowContext(ctx, `SELECT is_repeatable FROM economy_prize WHERE id=$1`, prizeID).Scan(&repeatable); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if repeatable {
		return true, nil
	}
	owned, err := prizeClaimedByUser(ctx, tx, userID, prizeID)
	return !owned, err
}

func (s *service) rollSlots(ctx context.Context, tx *sql.Tx, g *chanceGame, userID string) ([]chanceOutcome, int, error) {
	keys := map[int]string{}
	for i, r := range g.Rows {
		if slotKind(r) == "currency" {
			key, err := s.chanceCurrencyKey(ctx, tx, r.CurrencyID)
			if err != nil {
				return nil, 0, err
			}
			keys[i] = key
		}
	}
	first := -1
	tries := 1
	if g.Reroll {
		tries = chanceRerollCap
		if g.RerollCap != nil && *g.RerollCap > 0 {
			tries = *g.RerollCap
		}
	}
	for attempt := 0; attempt < tries; attempt++ {
		out := []chanceOutcome{}
		for i, r := range g.Rows {
			roll, err := secureInt(chanceTotal)
			if err != nil {
				return nil, 0, err
			}
			if first < 0 {
				first = roll
			}
			if roll >= scaledChance(r.Chance) {
				continue
			}
			o := chanceOutcome{Slot: i, Kind: slotKind(r)}
			switch o.Kind {
			case "currency":
				amount, err := secureRange(r.Min, r.Max)
				if err != nil {
					return nil, 0, err
				}
				o.CurrencyKey, o.Amount = keys[i], amount
			case "prizeGroup":
				if r.PrizeGroupID == nil {
					continue
				}
				id, err := s.drawFromPrizeGroup(ctx, tx, *r.PrizeGroupID)
				if err != nil {
					return nil, 0, err
				}
				o.PrizeGroupID, o.PrizeID = *r.PrizeGroupID, id
			default:
				if r.PrizeID == nil {
					continue
				}
				o.PrizeID = *r.PrizeID
			}
			if o.PrizeID != "" {
				open, err := s.slotPrizeOpen(ctx, tx, userID, o.PrizeID)
				if err != nil {
					return nil, 0, err
				}
				if !open {
					continue
				}
			}
			out = append(out, o)
		}
		if len(out) > 0 {
			return out, max(first, 0), nil
		}
	}
	return []chanceOutcome{}, max(first, 0), nil
}

func (s *service) insertChancePlay(ctx context.Context, tx *sql.Tx, play *chancePlayRow, g *chanceGame, userID string, price *coinPrice, now time.Time, since int) error {
	var priceJSON, outcomeJSON any
	if price != nil {
		raw, _ := json.Marshal(price)
		priceJSON = string(raw)
	}
	if play.Outcome != nil {
		raw, _ := json.Marshal(play.Outcome)
		outcomeJSON = string(raw)
	}
	var finished any
	if play.Status == "done" {
		finished = now
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO economy_chance_play (id, game_id, user_id, mode, status, roll, slot, prize_id, prize_group_id, claim_id, price, created_at, finished_at, outcome, forced)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13,$14::jsonb,$15)`,
		play.ID, g.ID, userID, play.Mode, play.Status, play.Roll, play.Slot, play.PrizeID, play.GroupID, play.ClaimID, priceJSON, now, finished, outcomeJSON, play.Forced); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE economy_chance_state SET plays=plays+1, last_play_at=$3, since_jackpot=$4 WHERE game_id=$1 AND user_id=$2`, g.ID, userID, now, since)
	return err
}

func slotRewards(outcome []chanceOutcome) []prizeItem {
	out := []prizeItem{}
	for _, o := range outcome {
		if o.Skipped {
			continue
		}
		if o.Kind == "currency" {
			out = append(out, prizeItem{Type: "currency", Key: o.CurrencyKey, Action: "add", Value: json.RawMessage(jsonInt(o.Amount))})
			continue
		}
		out = append(out, o.Rewards...)
	}
	return out
}

func (s *service) finishSlotsPlay(ctx context.Context, play *chancePlayRow) ([]prizeItem, error) {
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM economy_chance_play WHERE id=$1`, play.ID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("PLAY_NOT_FOUND")
		}
		return nil, err
	}
	if status == "done" {
		play.Status = "done"
		return slotRewards(play.Outcome), nil
	}
	credits := map[string]int64{}
	for _, o := range play.Outcome {
		if o.Kind == "currency" && o.Amount > 0 {
			credits[o.CurrencyKey] += o.Amount
		}
	}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if play.Price != nil {
			if err := s.walletChange(ctx, tx, play.UserID, play.ID, map[string]int64{play.Price.CurrencyKey: -play.Price.Amount}, map[string]any{
				"reason": "chance_ticket", "sourceType": "chance", "sourceId": play.ID, "part": "debit",
			}); err != nil {
				return err
			}
		}
		return s.walletChange(ctx, tx, play.UserID, play.ID+":slots", credits, map[string]any{
			"reason": "chance_game", "sourceType": "chance", "sourceId": play.ID, "part": "prize",
		})
	})
	if err != nil {
		return nil, err
	}
	for i := range play.Outcome {
		o := &play.Outcome[i]
		if o.PrizeID == "" || o.Skipped {
			continue
		}
		rewards, err := s.grant(ctx, grantRequest{UserID: play.UserID, PrizeID: o.PrizeID, SourceType: "chance_slot", SourceID: play.ID + ":" + itoa(o.Slot), Reason: "chance_game"})
		if err != nil {
			if app, ok := asApp(err); ok && (app.Code == "PRIZE_ALREADY_CLAIMED" || app.Code == "PRIZE_ALREADY_OWNED" || app.Code == "PRIZE_NOT_FOUND") {
				o.Skipped = true
				continue
			}
			return nil, err
		}
		o.Rewards = rewards
	}
	raw, _ := json.Marshal(play.Outcome)
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_chance_play SET status='done', error=NULL, finished_at=now(), outcome=$2::jsonb WHERE id=$1 AND status='pending'`, play.ID, string(raw)); err != nil {
		return nil, err
	}
	play.Status = "done"
	return slotRewards(play.Outcome), nil
}
