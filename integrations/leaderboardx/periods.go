package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

var epoch = time.Unix(0, 0).UTC()

func (s *service) closeDue(ctx context.Context) error {
	boards, err := s.boards(ctx, "")
	if err != nil {
		return err
	}
	now := s.now().UTC()
	var firstErr error
	for _, b := range boards {
		if b.ResetSchedule != "" {
			rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT r.expiry_time FROM leaderboard_record r
				WHERE r.leaderboard_id=$1 AND r.expiry_time > $2 AND r.expiry_time <= $3 AND r.expiry_time > $4
				AND NOT EXISTS (SELECT 1 FROM leaderboardx_period p WHERE p.board_key=$1 AND p.expiry_time=r.expiry_time AND p.reason IN ('reset','end'))
				ORDER BY 1 LIMIT 20`, b.Key, epoch, now, b.CreatedAt)
			if err != nil {
				return err
			}
			var due []time.Time
			for rows.Next() {
				var t time.Time
				if rows.Scan(&t) == nil {
					due = append(due, t.UTC())
				}
			}
			rows.Close()
			for _, exp := range due {
				start, _ := periodBounds(b.ResetSchedule, exp.Add(-time.Second))
				if _, err := s.closePeriod(ctx, b, exp, exp, start, "reset", "", true); err != nil && firstErr == nil {
					firstErr = fmt.Errorf("%s at %s: %w", b.Key, exp.Format(time.RFC3339), err)
				}
				if b, err = s.boardByID(ctx, b.ID); err != nil || b.FinishedAt != nil {
					break
				}
			}
			if b == nil {
				continue
			}
		}
		if b.FinishedAt == nil && b.EndsAt != nil && !now.Before(*b.EndsAt) {
			exp := expiry(b.ResetSchedule, b.EndsAt.Add(-time.Second))
			start, _ := periodBounds(b.ResetSchedule, b.EndsAt.Add(-time.Second))
			if start == nil && b.StartsAt != nil {
				start = b.StartsAt
			}
			if _, err := s.closePeriod(ctx, b, exp, b.EndsAt.UTC(), start, "end", "", true); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("%s end: %w", b.Key, err)
			}
			_, _ = s.db.ExecContext(ctx, `UPDATE leaderboardx_board SET finished_at=COALESCE(finished_at, now()) WHERE id=$1`, b.ID)
			invalidateBoards()
		}
	}
	return firstErr
}

func (s *service) closePeriod(ctx context.Context, b *board, exp, periodEnd time.Time, periodStart *time.Time, reason, by string, pay bool) (string, error) {
	cfg := s.settings(ctx)
	now := s.now().UTC()
	if cfg.Rewards.CatchUpDays > 0 && periodEnd.Before(now.Add(-time.Duration(cfg.Rewards.CatchUpDays)*24*time.Hour)) {
		pay = false
	}
	mode := b.payMode(cfg)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	id := uuid.Must(uuid.NewV4()).String()
	conf := map[string]any{"sortOrder": b.SortOrder, "operator": b.Operator, "tierBasis": b.TierBasis, "tiers": b.Tiers, "rewards": b.Rewards, "rewardMinScore": b.RewardMinScore, "paid": pay}
	res, err := tx.ExecContext(ctx, `INSERT INTO leaderboardx_period (id, board_id, board_key, board_name, period_start, period_end, expiry_time, reason, config, pay_mode, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT (board_key, period_end) DO NOTHING`,
		id, b.ID, b.Key, mustJSON(b.Name), periodStart, periodEnd, exp, reason, mustJSON(conf), mode, by)
	if err != nil {
		return "", err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var existing string
		err := s.db.QueryRowContext(ctx, `SELECT id::text FROM leaderboardx_period WHERE board_key=$1 AND period_end=$2`, b.Key, periodEnd).Scan(&existing)
		return existing, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT owner_id::text, COALESCE(username,''), score, subscore FROM leaderboard_record WHERE leaderboard_id=$1 AND expiry_time=$2 ORDER BY `+coreOrder(b.SortOrder), b.Key, exp)
	if err != nil {
		return "", err
	}
	list := []placed{}
	for rows.Next() {
		var p placed
		if err := rows.Scan(&p.UserID, &p.Username, &p.Score, &p.Subscore); err != nil {
			rows.Close()
			return "", err
		}
		p.Rank = len(list) + 1
		list = append(list, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return "", err
	}
	assignTiers(b, list)
	rewarded := 0
	claimUntil := now.Add(time.Duration(cfg.Rewards.ClaimDays) * 24 * time.Hour)
	const batch = 500
	for i := 0; i < len(list); i += batch {
		end := min(i+batch, len(list))
		var sb strings.Builder
		args := make([]any, 0, (end-i)*10)
		for j, p := range list[i:end] {
			var pl *plan
			if pay {
				pl = planFor(b, p, cfg.Rewards.Stack, id+":"+p.UserID)
			}
			status, rawPlan := "none", any(nil)
			var claim any
			if pl != nil {
				rewarded++
				rawPlan = mustJSON(pl)
				status = "pending"
				if mode == settings.PayClaim {
					status, claim = "claimable", claimUntil
				}
			}
			if j > 0 {
				sb.WriteString(",")
			}
			n := len(args)
			fmt.Fprintf(&sb, "($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d::jsonb,$%d,$%d)", n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8, n+9, n+10, n+11)
			args = append(args, id, p.UserID, p.Rank, p.Tier, p.TierRank, p.Score, p.Subscore, p.Username, rawPlan, status, claim)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO leaderboardx_result (period_id, user_id, rank, tier, tier_rank, score, subscore, username, plan, status, claim_expires_at) VALUES `+sb.String(), args...); err != nil {
			return "", err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE leaderboardx_period SET players=$2, rewarded=$3 WHERE id=$1`, id, len(list), rewarded); err != nil {
		return "", err
	}
	if reason == "reset" || reason == "end" {
		if _, err := tx.ExecContext(ctx, `UPDATE leaderboardx_board SET periods_closed=periods_closed+1,
			finished_at = CASE WHEN max_periods > 0 AND periods_closed+1 >= max_periods THEN COALESCE(finished_at, now()) ELSE finished_at END WHERE id=$1`, b.ID); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	invalidateBoards()
	s.logger.Info("Leaderboard %s: period ending %s closed (%s): %d players, %d rewarded", b.Key, periodEnd.Format(time.RFC3339), reason, len(list), rewarded)
	if rewarded > 0 {
		s.afterClose(ctx, b, id, mode, cfg)
	}
	return id, nil
}

func (s *service) afterClose(ctx context.Context, b *board, periodID, mode string, cfg settings.All) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text, rank, tier, score, status FROM leaderboardx_result WHERE period_id=$1 AND plan IS NOT NULL ORDER BY rank`, periodID)
	if err != nil {
		return
	}
	type win struct {
		user, tier, status string
		rank               int
		score              int64
	}
	var wins []win
	for rows.Next() {
		var w win
		if rows.Scan(&w.user, &w.rank, &w.tier, &w.score, &w.status) == nil {
			wins = append(wins, w)
		}
	}
	rows.Close()
	for _, w := range wins {
		fields := map[string]any{"boardKey": b.Key, "rank": w.rank, "tier": w.tier, "score": w.score, "periodId": periodID}
		if cfg.Events.AnalyticsEvent != "" {
			_, errs := gameevents.Publish(ctx, gameevents.Event{ID: owner + ":" + periodID + ":" + w.user, UserID: w.user, Name: cfg.Events.AnalyticsEvent, Context: b.Key, Fields: fields})
			for o, err := range errs {
				s.logger.Debug("Leaderboard event to %s not recorded: %v", o, err)
			}
		}
		if cfg.Events.EconomyEvent != "" && playerevents.Available() {
			if _, err := playerevents.Record(ctx, playerevents.Event{UserID: w.user, Key: cfg.Events.EconomyEvent, Value: float64(w.rank), Metadata: fields}); err != nil {
				s.logger.Debug("Leaderboard economy event not recorded: %v", err)
			}
		}
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: w.user, Kind: activity.KindEvent, Action: "leaderboard_ranked", RefID: b.Key, Actor: "system",
			Data: map[string]any{"boardKey": b.Key, "rank": w.rank, "tier": w.tier, "score": w.score, "periodId": periodID, "status": w.status}})
		if w.status == "claimable" && cfg.Events.Notify {
			s.notify(ctx, w.user, "leaderboard_reward", map[string]any{"boardKey": b.Key, "periodId": periodID, "rank": w.rank, "tier": w.tier, "claim": true}, cfg.Events.Code())
		}
	}
	if mode == settings.PayAuto {
		for _, w := range wins {
			if _, err := s.payResult(ctx, periodID, w.user, false); err != nil {
				s.logger.Warn("Leaderboard reward %s/%s not paid yet: %v", periodID, w.user, err)
			}
		}
	}
}

func (s *service) payResult(ctx context.Context, periodID, userID string, claim bool) (map[string]any, error) {
	cfg := s.settings(ctx)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var status, boardKey string
	var rawPlan, rawPaid []byte
	var tries, rank int
	var tier string
	var claimUntil *time.Time
	err = tx.QueryRowContext(ctx, `SELECT r.status, r.plan, r.paid, r.pay_tries, r.claim_expires_at, p.board_key, r.rank, r.tier FROM leaderboardx_result r JOIN leaderboardx_period p ON p.id=r.period_id
		WHERE r.period_id=$1 AND r.user_id=$2 FOR UPDATE OF r`, periodID, userID).Scan(&status, &rawPlan, &rawPaid, &tries, &claimUntil, &boardKey, &rank, &tier)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("LEADERBOARD_REWARD_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{"periodId": periodID, "boardKey": boardKey, "rank": rank, "tier": tier, "status": status}
	paid := map[string]any{}
	unmarshal(rawPaid, &paid)
	switch status {
	case "paid":
		out["paid"], out["replayed"] = paid, true
		return out, nil
	case "claimable":
		if !claim {
			return out, nil
		}
		if claimUntil != nil && claimUntil.Before(s.now()) {
			_, _ = tx.ExecContext(ctx, `UPDATE leaderboardx_result SET status='expired' WHERE period_id=$1 AND user_id=$2`, periodID, userID)
			_ = tx.Commit()
			return nil, apierr.Conflict("LEADERBOARD_REWARD_EXPIRED").With("expiredAt", claimUntil.UTC().Format(time.RFC3339))
		}
	case "pending":
	default:
		if claim {
			return nil, apierr.Conflict("LEADERBOARD_REWARD_NOT_CLAIMABLE").With("status", status)
		}
		return out, nil
	}
	var pl plan
	unmarshal(rawPlan, &pl)
	ok, prizes, lastErr := s.payPlan(ctx, userID, periodID, boardKey, &pl, cfg)
	next := "paid"
	if !ok {
		next = "pending"
		if !claim && tries+1 >= cfg.Rewards.MaxTries {
			next = "failed"
		}
	}
	if prev, ok := paid["prizes"].([]any); ok {
		for _, p := range prev {
			if m, ok := p.(map[string]any); ok {
				prizes = append([]map[string]any{m}, prizes...)
			}
		}
	}
	paid = map[string]any{"currencies": currencyList(pl.Currencies), "prizes": prizes}
	if !pl.WalletDone {
		paid["currencies"] = []map[string]any{}
	}
	pending := []map[string]any{}
	for _, p := range pl.Prizes {
		if !p.Paid {
			pending = append(pending, map[string]any{"prizeKey": p.PrizeKey, "prizeGroupKey": p.PrizeGroupKey})
		}
	}
	var paidAt any
	if next == "paid" {
		paidAt = s.now().UTC()
	}
	if _, err := tx.ExecContext(ctx, `UPDATE leaderboardx_result SET status=$3, plan=$4, paid=$5, pay_tries=pay_tries+1, last_error=$6, paid_at=$7 WHERE period_id=$1 AND user_id=$2`,
		periodID, userID, next, mustJSON(pl), mustJSON(paid), lastErr, paidAt); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	out["status"], out["paid"], out["pendingPrizes"], out["replayed"] = next, paid, pending, false
	if next == "paid" && !claim && cfg.Events.Notify {
		s.notify(ctx, userID, "leaderboard_reward", map[string]any{"boardKey": boardKey, "periodId": periodID, "rank": rank, "tier": tier, "claim": false, "rewards": paid}, cfg.Events.Code())
	}
	return out, nil
}

func (s *service) payPlan(ctx context.Context, userID, periodID, boardKey string, pl *plan, cfg settings.All) (bool, []map[string]any, string) {
	grant := owner + ":" + periodID + ":" + userID
	reason := "leaderboard:" + boardKey
	if !pl.WalletDone && len(pl.Currencies) > 0 {
		applied, err := pluginkit.WalletApplied(ctx, s.db, userID, grant)
		if err == nil && !applied {
			md := map[string]any{"grantId": grant, "module": owner, "sourceType": sourceType, "sourceId": periodID, "reason": reason, "boardKey": boardKey}
			_, _, err = s.nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: pl.Currencies, Metadata: md}}, true)
		}
		if err != nil {
			return false, nil, "wallet: " + err.Error()
		}
	}
	pl.WalletDone = true
	ok, lastErr := true, ""
	prizes := []map[string]any{}
	for i := range pl.Prizes {
		pr := &pl.Prizes[i]
		if pr.Paid {
			continue
		}
		res, err := s.prize(ctx, economyhooks.PrizeGrant{UserID: userID, PrizeKey: pr.PrizeKey, GroupKey: pr.PrizeGroupKey, GroupMode: cfg.Rewards.PrizeGroupMode,
			SourceType: sourceType, SourceID: pr.SourceID, Reason: reason})
		if err != nil {
			pr.Error, ok, lastErr = err.Error(), false, "prize: "+err.Error()
			continue
		}
		pr.Paid, pr.Error = true, ""
		prizes = append(prizes, res.Prizes...)
	}
	return ok, prizes, lastErr
}

func (s *service) retryPayments(ctx context.Context, limit int) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.period_id::text, r.user_id::text FROM leaderboardx_result r WHERE r.status='pending' ORDER BY r.period_id, r.rank LIMIT $1`, limit)
	if err != nil {
		return
	}
	var list [][2]string
	for rows.Next() {
		var p, u string
		if rows.Scan(&p, &u) == nil {
			list = append(list, [2]string{p, u})
		}
	}
	rows.Close()
	for _, x := range list {
		if _, err := s.payResult(ctx, x[0], x[1], false); err != nil {
			s.logger.Warn("Leaderboard reward %s/%s not paid: %v", x[0], x[1], err)
		}
	}
}

func resultRow(periodID, userID string, rank, tierRank int, tier string, score, sub int64, username, status string, rawPlan, rawPaid []byte, claimUntil, paidAt *time.Time, lastErr string, tries int) map[string]any {
	var pl plan
	unmarshal(rawPlan, &pl)
	paid := map[string]any{}
	unmarshal(rawPaid, &paid)
	prizes := []map[string]any{}
	for _, p := range pl.Prizes {
		prizes = append(prizes, map[string]any{"prizeKey": p.PrizeKey, "prizeGroupKey": p.PrizeGroupKey, "paid": p.Paid})
	}
	out := map[string]any{"periodId": periodID, "userId": userID, "rank": rank, "tier": tier, "tierRank": tierRank, "score": score, "subscore": sub, "username": username,
		"status": status, "rewards": map[string]any{"currencies": currencyList(pl.Currencies), "prizes": prizes}, "paid": paid, "lastError": lastErr, "payTries": tries}
	if claimUntil != nil {
		out["claimExpiresAt"] = claimUntil.UTC().Format(time.RFC3339)
	} else {
		out["claimExpiresAt"] = nil
	}
	if paidAt != nil {
		out["paidAt"] = paidAt.UTC().Format(time.RFC3339)
	} else {
		out["paidAt"] = nil
	}
	return out
}

func rawJSON(v any) json.RawMessage { return json.RawMessage(mustJSON(v)) }
