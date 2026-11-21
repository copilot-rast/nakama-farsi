package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const maxPayTries = 8

func (s *service) payPlan(ctx context.Context, userID string, pl *plan, res map[string]any, first bool) bool {
	cfg := s.settings(ctx)
	ok := true
	if len(pl.Currencies) > 0 {
		changes := map[string]int64{}
		for k, v := range pl.Currencies {
			if v > 0 {
				changes[k] = v
			}
		}
		if len(changes) > 0 {
			applied, err := pluginkit.WalletApplied(ctx, s.db, userID, pl.GrantID)
			if err == nil && !applied {
				md := map[string]any{"grantId": pl.GrantID, "module": "campaign", "sourceType": sourceType, "sourceId": pl.SourceID, "reason": pl.Reason}
				_, _, err = s.nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: changes, Metadata: md}}, true)
			}
			if err != nil {
				s.logger.Warn("Campaign wallet payment %s failed: %v", pl.GrantID, err)
				return false
			}
		}
	}
	if pl.Refund != nil && !pl.Refund.Done {
		if err := s.giveBack(ctx, userID, strings.TrimPrefix(pl.GrantID, "campaign:"), pl.Refund); err != nil {
			s.logger.Warn("Campaign play cost refund %s failed: %v", pl.GrantID, err)
			ok = false
		}
	}
	if pl.Refund != nil && pl.Refund.Done {
		res["playCostRefund"] = map[string]any{"currencyKey": pl.Refund.CurrencyKey, "amount": pl.Refund.Refunded}
	}
	paid := []map[string]any{}
	for i := range pl.Prizes {
		pr := &pl.Prizes[i]
		if pr.Paid {
			continue
		}
		out, err := s.prize(ctx, economyhooks.PrizeGrant{UserID: userID, PrizeKey: pr.PrizeKey, GroupKey: pr.PrizeGroupKey, GroupMode: cfg.Rewards.PrizeGroupMode,
			SourceType: sourceType, SourceID: pr.SourceID, Reason: pl.Reason})
		if err != nil {
			pr.Error = err.Error()
			ok = false
			continue
		}
		pr.Paid, pr.Error = true, ""
		for _, p := range out.Prizes {
			p["for"] = pr.For
			paid = append(paid, p)
		}
	}
	if len(paid) > 0 {
		prev, _ := res["prizes"].([]any)
		list := append([]any{}, prev...)
		for _, p := range paid {
			list = append(list, p)
		}
		res["prizes"] = list
	}
	if _, set := res["prizes"]; !set {
		res["prizes"] = []any{}
	}
	pending := []map[string]any{}
	for _, pr := range pl.Prizes {
		if !pr.Paid {
			pending = append(pending, map[string]any{"prizeKey": pr.PrizeKey, "prizeGroupKey": pr.PrizeGroupKey, "for": pr.For})
		}
	}
	res["pendingPrizes"] = pending
	if pl.MetaKey != "" && first {
		if out, err := economyhooks.SettleMetas(ctx, userID, []string{pl.MetaKey}); err == nil && out != nil {
			res["levels"] = out["levels"]
		}
	}
	if pl.Event != nil && !pl.Event.Sent {
		pl.Event.Sent = true
		if pl.Event.Analytics != "" || len(pl.Event.Typed) > 0 {
			answers, errs := gameevents.Publish(ctx, gameevents.Event{ID: strings.TrimPrefix(pl.GrantID, "campaign:"), UserID: userID, Name: pl.Event.Analytics, Context: pl.Event.Context, Fields: pl.Event.Fields,
				Source: "campaign", Typed: pl.Event.Typed})
			if a, found := answers["analytics"]; found {
				res["trophies"] = a
			}
			for owner, err := range errs {
				s.logger.Debug("Campaign event to %s not recorded: %v", owner, err)
			}
		}
		if pl.Event.Economy != "" && playerevents.Available() {
			if _, err := playerevents.Record(ctx, playerevents.Event{UserID: userID, Key: pl.Event.Economy, Value: pl.Event.Value, Metadata: pl.Event.Fields}); err != nil {
				s.logger.Debug("Campaign economy event %s not recorded: %v", pl.Event.Economy, err)
			}
		}
	}
	return ok
}

func (s *service) finishAttempt(ctx context.Context, attemptID string) (any, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var userID, status string
	var rawRes, rawPlan []byte
	var tries int
	if err := tx.QueryRowContext(ctx, `SELECT user_id::text, status, result, plan, pay_tries FROM campaign_attempt WHERE id=$1 FOR UPDATE`, attemptID).Scan(&userID, &status, &rawRes, &rawPlan, &tries); err != nil {
		return nil, err
	}
	res := map[string]any{}
	unmarshal(rawRes, &res)
	if status != "paying" {
		return res, nil
	}
	var pl plan
	unmarshal(rawPlan, &pl)
	ok := s.payPlan(ctx, userID, &pl, res, tries == 0)
	if tries == 0 {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindEvent, Action: "stage_completed", RefID: pl.SourceID, Actor: "player",
			Data: map[string]any{"stageKey": pl.SourceID, "stars": res["stars"], "success": res["success"], "rewards": res["rewards"], "xp": res["xpGained"]}})
	}
	if xp, lv, next := s.levelAfter(ctx, userID); lv >= 0 {
		res["xp"], res["level"], res["nextLevelXp"] = xp, lv, next
	}
	next := "done"
	if !ok && tries+1 < maxPayTries {
		next = "paying"
	}
	rawRes, _ = json.Marshal(res)
	rawPlan, _ = json.Marshal(pl)
	if _, err := tx.ExecContext(ctx, `UPDATE campaign_attempt SET status=$2, result=$3, plan=$4, pay_tries=pay_tries+1 WHERE id=$1`, attemptID, next, rawRes, rawPlan); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *service) levelAfter(ctx context.Context, userID string) (int64, int, any) {
	cfg := s.settings(ctx)
	p := &player{s: s, q: s.db, userID: userID, cfg: cfg, levels: map[string]int{}}
	var xp int64
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE((SELECT xp FROM campaign_user WHERE user_id=$1),0)`, userID).Scan(&xp)
	lv := p.level(ctx, cfg.XP.LevelTableKey)
	return xp, lv, s.nextLevelXP(ctx, p, lv)
}

func (s *service) retryPayments(ctx context.Context, limit int) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM campaign_attempt WHERE status='paying' AND completed_at < now() - interval '30 seconds' ORDER BY completed_at LIMIT $1`, limit)
	if err != nil {
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		if _, err := s.finishAttempt(ctx, id); err != nil {
			s.logger.Warn("Campaign could not finish attempt %s: %v", id, err)
		}
	}
	rows, err = s.db.QueryContext(ctx, `SELECT user_id::text, request_id FROM campaign_claim WHERE status='paying' AND created_at < now() - interval '30 seconds' ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return
	}
	var claims [][2]string
	for rows.Next() {
		var u, r string
		if rows.Scan(&u, &r) == nil {
			claims = append(claims, [2]string{u, r})
		}
	}
	rows.Close()
	for _, c := range claims {
		if _, err := s.finishClaim(ctx, c[0], c[1]); err != nil {
			s.logger.Warn("Campaign could not finish claim %s: %v", c[1], err)
		}
	}
}

func (s *service) rpcFreerideClaim(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key       string `json:"key"`
		RequestID string `json:"requestId"`
		Lang      string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	return s.claimFreeride(ctx, userID, strings.TrimSpace(in.Key), strings.TrimSpace(in.RequestID))
}

func (s *service) claimFreeride(ctx context.Context, userID, key, requestID string) (any, error) {
	cfg := s.settings(ctx)
	if key == "" {
		return nil, apierr.BadRequest("FREERIDE_KEY_REQUIRED").WithField("key")
	}
	if !cfg.Freeride.TableEnabled {
		return nil, apierr.NotFound("FREERIDE_NOT_FOUND").WithField("key")
	}
	if requestID == "" {
		if cfg.Freeride.RequireRequestID {
			return nil, apierr.BadRequest("REQUEST_ID_REQUIRED").WithField("requestId")
		}
		requestID = uuid.Must(uuid.NewV4()).String()
	}
	if len(requestID) > 64 {
		return nil, apierr.BadRequest("INVALID_REQUEST_ID").WithField("requestId")
	}
	var status string
	err := s.db.QueryRowContext(ctx, `SELECT status FROM campaign_claim WHERE user_id=$1 AND request_id=$2`, userID, requestID).Scan(&status)
	if err == nil {
		out, err := s.finishClaim(ctx, userID, requestID)
		if m, ok := out.(map[string]any); ok && err == nil {
			m["replayed"] = true
		}
		return out, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	f, err := scanFreeride(s.db.QueryRowContext(ctx, `SELECT `+freerideCols+` FROM campaign_freeride WHERE key=$1 AND is_active`, key))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("FREERIDE_NOT_FOUND").WithField("key")
	}
	if err != nil {
		return nil, err
	}
	p := &player{s: s, q: s.db, userID: userID, cfg: cfg, levels: map[string]int{}, groups: map[string]groupState{}}
	if f.MinLevel > 0 {
		if lv := p.level(ctx, cfg.XP.LevelTableKey); lv < f.MinLevel {
			return nil, apierr.Forbidden("LEVEL_TOO_LOW").With("level", lv).With("minLevel", f.MinLevel)
		}
	}
	now := s.now().UTC()
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		why, err := s.takeGroup(ctx, tx, p, f.group(), f.Cooldown, f.DailyLimit, now)
		if err != nil {
			return err
		}
		g := p.groups[f.group()]
		if why != "" {
			e := apierr.Conflict("REWARD_LIMITED").With("reason", why)
			if at := nextRewardAt(g, f.Cooldown, now); at != nil {
				e = e.With("nextRewardAt", at.Format(time.RFC3339))
			}
			if left := rewardsLeft(g, f.DailyLimit, now, s.zone(cfg)); left != nil {
				e = e.With("rewardsLeftToday", *left)
			}
			return e
		}
		var xpTotal int64
		if err := s.addXP(ctx, tx, userID, cfg.XP.MetaKey, f.XP, &xpTotal); err != nil {
			return err
		}
		pl := plan{GrantID: "campaign:claim:" + userID + ":" + requestID, Currencies: f.Rewards.Currencies, MetaKey: cfg.XP.MetaKey, Reason: "freeride", SourceID: f.Key}
		if f.Rewards.PrizeKey != "" || f.Rewards.PrizeGroupKey != "" {
			pl.Prizes = []prizeRef{{PrizeKey: f.Rewards.PrizeKey, PrizeGroupKey: f.Rewards.PrizeGroupKey, SourceID: "claim:" + userID + ":" + requestID, For: "freeride"}}
		}
		if f.XP == 0 {
			pl.MetaKey = ""
		}
		pl.Event = &eventRef{Context: f.Key, Typed: stageTyped("freeride", f.Key, true, 0, f.Rewards.Currencies)}
		pl.Event.Typed[0].Score = nil
		res := map[string]any{"key": f.Key, "requestId": requestID, "rewards": rewardList(f.Rewards.Currencies), "xpGained": f.XP, "xp": xpTotal,
			"nextRewardAt": iso(nextRewardAt(g, f.Cooldown, now)), "rewardsLeftToday": rewardsLeft(g, f.DailyLimit, now, s.zone(cfg)), "replayed": false}
		rawRes, _ := json.Marshal(res)
		rawPlan, _ := json.Marshal(pl)
		_, err = tx.ExecContext(ctx, `INSERT INTO campaign_claim (user_id, request_id, freeride_id, result, plan, expires_at) VALUES ($1,$2,$3,$4,$5,$6)`,
			userID, requestID, f.ID, rawRes, rawPlan, now.Add(time.Duration(cfg.Freeride.RequestHours)*time.Hour))
		return err
	})
	if err != nil {
		var pg interface{ SQLState() string }
		if errors.As(err, &pg) && pg.SQLState() == "23505" {
			return s.finishClaim(ctx, userID, requestID)
		}
		return nil, err
	}
	return s.finishClaim(ctx, userID, requestID)
}

func (s *service) finishClaim(ctx context.Context, userID, requestID string) (any, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var status string
	var rawRes, rawPlan []byte
	var created time.Time
	if err := tx.QueryRowContext(ctx, `SELECT status, result, plan, created_at FROM campaign_claim WHERE user_id=$1 AND request_id=$2 FOR UPDATE`, userID, requestID).Scan(&status, &rawRes, &rawPlan, &created); err != nil {
		return nil, err
	}
	res := map[string]any{}
	unmarshal(rawRes, &res)
	if status == "done" {
		return res, nil
	}
	var pl plan
	unmarshal(rawPlan, &pl)
	ok := s.payPlan(ctx, userID, &pl, res, true)
	if xp, lv, next := s.levelAfter(ctx, userID); lv >= 0 {
		res["xp"], res["level"], res["nextLevelXp"] = xp, lv, next
	}
	next := "done"
	if !ok && time.Since(created) < time.Hour {
		next = "paying"
	}
	rawRes, _ = json.Marshal(res)
	rawPlan, _ = json.Marshal(pl)
	if _, err := tx.ExecContext(ctx, `UPDATE campaign_claim SET status=$3, result=$4, plan=$5 WHERE user_id=$1 AND request_id=$2`, userID, requestID, next, rawRes, rawPlan); err != nil {
		return nil, err
	}
	return res, tx.Commit()
}
