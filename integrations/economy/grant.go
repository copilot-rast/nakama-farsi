package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
		for k, v := range req.Debit {
			changes[k] -= v
		}
		var writes []*runtime.StorageWrite
		var metas []prizeItem
		for _, item := range prize.Items {
			switch item.Type {
			case "currency":
				v, _ := int64FromJSON(item.Value)
				changes[item.Key] += v
			case "storage":
				writes = append(writes, &runtime.StorageWrite{Collection: "economy_items", Key: item.Key, UserID: req.UserID, Value: string(item.Value), PermissionRead: 1, PermissionWrite: 0})
			case "meta":
				metas = append(metas, item)
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
		if len(changes) > 0 && !applied {
			wallets = append(wallets, &runtime.WalletUpdate{UserID: req.UserID, Changeset: changes, Metadata: map[string]any{
				"module": "economy", "reason": req.Reason, "sourceType": req.SourceType, "sourceId": req.SourceID, "prizeId": prize.ID, "grantId": rowID,
			}})
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
		raw, _ := json.Marshal(prize.Items)
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_prize SET claimed=true, error=NULL, rewards=$2, updated_at=now() WHERE id=$1 AND claimed=false`, rowID, string(raw)); err != nil {
			return err
		}
		if req.After != nil {
			if err := req.After(ctx, tx); err != nil {
				return err
			}
		}
		rewards = prize.Items
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
