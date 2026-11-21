package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

type grantRequest struct {
	UserID     string
	PrizeID    string
	SourceType string
	SourceID   string
	Reason     string
	Debit      map[string]int64
	CheckItems bool
	After      func(ctx context.Context, tx *sql.Tx) error
}

type prizeRow struct {
	ID         string
	Key        string
	Repeatable bool
	Items      []prizeItem
}

func loadPrize(ctx context.Context, q queryer, id string) (prizeRow, error) {
	var p prizeRow
	var items string
	err := q.QueryRowContext(ctx, `SELECT id::text, key, is_repeatable, items::text FROM economy_prize WHERE id=$1`, id).Scan(&p.ID, &p.Key, &p.Repeatable, &items)
	if errors.Is(err, sql.ErrNoRows) {
		return p, notFound("PRIZE_NOT_FOUND")
	}
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal([]byte(items), &p.Items); err != nil {
		return p, err
	}
	return p, nil
}

func prizeClaimedByUser(ctx context.Context, q queryer, userID, prizeID string) (bool, error) {
	var claimed bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_prize WHERE once_key=$1)`, userID+":"+prizeID).Scan(&claimed)
	return claimed, err
}

func (s *service) grant(ctx context.Context, req grantRequest) ([]prizeItem, error) {
	prize, err := loadPrize(ctx, s.db, req.PrizeID)
	if err != nil {
		return nil, err
	}
	var rowID string
	var claimed bool
	var stored string
	err = s.db.QueryRowContext(ctx, `SELECT id::text, claimed, rewards::text FROM economy_user_prize WHERE source_type=$1 AND source_id=$2 AND prize_id=$3`,
		req.SourceType, req.SourceID, prize.ID).Scan(&rowID, &claimed, &stored)
	if errors.Is(err, sql.ErrNoRows) {
		rowID = uuid.Must(uuid.NewV4()).String()
		var once any
		if !prize.Repeatable {
			once = req.UserID + ":" + prize.ID
		}
		_, err = s.db.ExecContext(ctx, `INSERT INTO economy_user_prize (id, user_id, prize_id, source_type, source_id, once_key) VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (source_type, source_id, prize_id) DO NOTHING`, rowID, req.UserID, prize.ID, req.SourceType, req.SourceID, once)
		if _, unique := isUniqueViolation(err); unique {
			return nil, conflict("PRIZE_ALREADY_CLAIMED", "this prize can be received only once")
		}
		if err != nil {
			return nil, err
		}
		err = s.db.QueryRowContext(ctx, `SELECT id::text, claimed, rewards::text FROM economy_user_prize WHERE source_type=$1 AND source_id=$2 AND prize_id=$3`,
			req.SourceType, req.SourceID, prize.ID).Scan(&rowID, &claimed, &stored)
	}
	if err != nil {
		return nil, err
	}
	if claimed {
		var items []prizeItem
		_ = json.Unmarshal([]byte(stored), &items)
		if req.After != nil {
			if err := s.inTx(ctx, func(tx *sql.Tx) error { return req.After(ctx, tx) }); err != nil {
				return nil, err
			}
		}
		return items, nil
	}

	var rewards []prizeItem
	var unlocked []achievementChange
	granted := false
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var owner string
		if err := tx.QueryRowContext(ctx, `SELECT user_id::text, claimed, rewards::text FROM economy_user_prize WHERE id=$1 FOR UPDATE`, rowID).Scan(&owner, &claimed, &stored); err != nil {
			return err
		}
		if claimed {
			_ = json.Unmarshal([]byte(stored), &rewards)
			return nil
		}
		if owner != req.UserID {
			return conflict("PRIZE_OWNER_MISMATCH", "grant row belongs to another user")
		}
		changes := map[string]int64{}
		debit := map[string]int64{}
		for k, v := range req.Debit {
			if v != 0 {
				debit[k] = -v
			}
		}
		var writes []*runtime.StorageWrite
		var metas []prizeItem
		var itemKeys, achKeys []string
		var tags, boosters []prizeItem
		resolved, err := s.resolvePrizeItems(ctx, tx, req.UserID, prize.Items, 0)
		if err != nil {
			return err
		}
		for _, item := range flattenPrizeItems(resolved) {
			switch item.Type {
			case "currency":
				v, _ := int64FromJSON(item.Value)
				changes[item.Key] += v
			case "storage":
				writes = append(writes, &runtime.StorageWrite{Collection: "economy_items", Key: item.Key, UserID: req.UserID, Value: string(item.Value), PermissionRead: 1, PermissionWrite: 0})
			case "meta":
				metas = append(metas, item)
			case "item":
				itemKeys = append(itemKeys, item.Key)
			case "achievement":
				achKeys = append(achKeys, item.Key)
			case "tag":
				tags = append(tags, item)
			case "booster":
				boosters = append(boosters, item)
			}
		}
		if len(itemKeys) > 0 {
			added := 0
			for _, key := range itemKeys {
				res, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref)
					SELECT $1, id, 'prize', $3 FROM economy_item WHERE key=$2 ON CONFLICT (user_id, item_id) DO NOTHING`, req.UserID, key, rowID)
				if err != nil {
					return err
				}
				n, _ := res.RowsAffected()
				added += int(n)
			}
			if added == 0 && req.CheckItems {
				return conflict("PRIZE_ALREADY_OWNED", "the player already owns everything this prize gives")
			}
		}
		for k, v := range changes {
			if v == 0 {
				delete(changes, k)
			}
		}
		applied, err := walletAlreadyApplied(ctx, tx, req.UserID, rowID)
		if err != nil {
			return err
		}
		var wallets []*runtime.WalletUpdate
		ledger := func(part string) map[string]any {
			md := map[string]any{"module": "economy", "reason": req.Reason, "sourceType": req.SourceType, "sourceId": req.SourceID, "prizeId": prize.ID, "grantId": rowID}
			if part != "" {
				md["part"] = part
			}
			return md
		}
		if len(debit) > 0 && !applied {
			wallets = append(wallets, &runtime.WalletUpdate{UserID: req.UserID, Changeset: debit, Metadata: ledger("debit")})
		}
		if len(changes) > 0 && !applied {
			part := ""
			if len(debit) > 0 {
				part = "prize"
			}
			wallets = append(wallets, &runtime.WalletUpdate{UserID: req.UserID, Changeset: changes, Metadata: ledger(part)})
		}
		if len(wallets) > 0 || len(writes) > 0 {
			if _, _, err := s.nk.MultiUpdate(ctx, nil, writes, nil, wallets, true); err != nil {
				var neg *runtime.WalletNegativeError
				if errors.As(err, &neg) {
					return fail(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "", "not enough "+neg.Path)
				}
				return err
			}
		}
		for _, item := range metas {
			if err := applyMeta(ctx, tx, req.UserID, item); err != nil {
				return err
			}
		}
		for _, item := range tags {
			if err := applyTag(ctx, tx, req.UserID, item, req.SourceType); err != nil {
				return err
			}
		}
		for _, item := range boosters {
			if err := applyBooster(ctx, tx, req.UserID, item, req.SourceType); err != nil {
				return err
			}
		}
		if len(achKeys) > 0 {
			got, err := s.grantAchievementItems(ctx, tx, req.UserID, achKeys, rowID)
			if err != nil {
				return err
			}
			unlocked = got
		}
		raw, _ := json.Marshal(resolved)
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_prize SET claimed=true, error=NULL, rewards=$2, updated_at=now() WHERE id=$1 AND claimed=false`, rowID, string(raw)); err != nil {
			return err
		}
		if req.After != nil {
			if err := req.After(ctx, tx); err != nil {
				return err
			}
		}
		rewards = resolved
		granted = true
		return nil
	})
	if err != nil {
		msg := err.Error()
		if app, ok := asApp(err); ok {
			msg = app.Code
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_user_prize SET error=$2, updated_at=now() WHERE id=$1 AND claimed=false`, rowID, truncate(msg, 500))
		return nil, err
	}
	if rewards == nil {
		rewards = []prizeItem{}
	}
	if granted && len(unlocked) > 0 {
		s.settleAchievements(ctx, req.UserID, unlocked, eventDepth(ctx), nil)
	}
	if granted && !grantEvaluatesMetas[req.SourceType] {
		var keys []string
		for _, item := range flattenPrizeItems(rewards) {
			if item.Type == "meta" {
				keys = append(keys, item.Key)
			}
		}
		if len(keys) > 0 {
			if ids, err := s.metaIDsByKey(ctx, keys); err == nil {
				if _, err := s.evaluateLevels(ctx, req.UserID, ids, eventDepth(ctx)+1); err != nil {
					s.logger.Warn("Economy meta checks after %s grant for %s failed: %v", req.SourceType, req.UserID, err)
				}
			}
		}
	}
	if granted && req.SourceType != "payment" {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: req.UserID, Kind: activity.KindPrize, Action: "granted", RefID: rowID, Actor: "system", Data: map[string]any{"prize": prize.Key, "source": req.SourceType, "reason": req.Reason}})
	}
	return rewards, nil
}

func walletAlreadyApplied(ctx context.Context, tx *sql.Tx, userID, grantID string) (bool, error) {
	var found bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM wallet_ledger WHERE user_id=$1 AND metadata->>'grantId'=$2)`, userID, grantID).Scan(&found)
	return found, err
}

func applyMeta(ctx context.Context, tx *sql.Tx, userID string, item prizeItem) error {
	value := string(item.Value)
	var err error
	if item.Action == "set" {
		_, err = tx.ExecContext(ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value)
			SELECT $1, id, $3::numeric FROM economy_meta WHERE key=$2
			ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value, updated_at=now()`, userID, item.Key, value)
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value)
			SELECT $1, id, default_value + $3::numeric FROM economy_meta WHERE key=$2
			ON CONFLICT (user_id, meta_id) DO UPDATE SET value=economy_user_meta.value + $3::numeric, updated_at=now()`, userID, item.Key, value)
	}
	if err != nil {
		return err
	}
	return nil
}

func applyTag(ctx context.Context, tx *sql.Tx, userID string, item prizeItem, source string) error {
	days := string(item.Value)
	if days == "" {
		days = "0"
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO base_user_tag (user_id, tag_id, created_by, expires_at)
		SELECT $1::uuid, t.id, $4, CASE WHEN $3::numeric > 0 THEN now() + $3::numeric * INTERVAL '1 day' END FROM base_tag t WHERE t.key=$2 AND t.key_type='user'
		ON CONFLICT (user_id, tag_id) DO UPDATE SET expires_at = CASE
			WHEN $3::numeric <= 0 THEN NULL
			WHEN base_user_tag.expires_at IS NULL THEN NULL
			WHEN $5 = 'set' OR base_user_tag.expires_at <= now() THEN now() + $3::numeric * INTERVAL '1 day'
			ELSE base_user_tag.expires_at + $3::numeric * INTERVAL '1 day' END`, userID, item.Key, days, "prize:"+source, item.Action)
	return err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
