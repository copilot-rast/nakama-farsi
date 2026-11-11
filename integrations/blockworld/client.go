package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

var (
	errBlockNotSold   = runtime.NewError("BLOCK_NOT_SOLD", 3)
	errPriceChanged   = runtime.NewError("PRICE_CHANGED", 9)
	errNotEnoughCoins = runtime.NewError("NOT_ENOUGH_COINS", 9)
	errUserNotFound   = runtime.NewError("USER_NOT_FOUND", 5)
)

func (s *service) rpcToken(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decode(payload, &in); err != nil {
		return nil, err
	}
	if err := restrict.RefuseSuspended(ctx, s.db, userID, in.Lang); err != nil {
		return nil, err
	}
	var username, display string
	var disabled bool
	err := s.db.QueryRowContext(ctx, `SELECT username, COALESCE(display_name, ''), disable_time > '1970-01-01 00:00:00+00' FROM users WHERE id = $1::uuid`, userID).Scan(&username, &display, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if disabled {
		return nil, runtime.NewError("ACCOUNT_BANNED", 7)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO blockworld_link (user_id) VALUES ($1::uuid)
		ON CONFLICT (user_id) DO UPDATE SET last_token_at = now(), tokens = blockworld_link.tokens + 1`, userID); err != nil {
		return nil, err
	}
	token, claims, err := s.signer.issue(userID, username, display, s.cfg.tokenTTL, s.now())
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"token":     token,
		"issuer":    claims.Issuer,
		"subject":   claims.Subject,
		"name":      claims.Name,
		"username":  claims.Username,
		"expiresAt": claims.Expires,
	}, nil
}

func (s *service) ensureStartCoins(ctx context.Context, userID string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		var value int64
		err := tx.QueryRowContext(ctx, `INSERT INTO economy_user_currency_init (user_id, currency_id)
			SELECT $1::uuid, id FROM economy_currency WHERE key = $2 AND default_value > 0
			ON CONFLICT (user_id, currency_id) DO NOTHING
			RETURNING (SELECT default_value FROM economy_currency WHERE key = $2)`, userID, s.cfg.currency).Scan(&value)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		_, _, err = s.nk.WalletUpdate(ctx, userID, map[string]int64{s.cfg.currency: value}, map[string]any{"module": "economy", "reason": "currency_default", "sourceType": "currency_default", "sourceId": userID}, true)
		return err
	})
}

func (s *service) balance(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE((wallet->>$2)::bigint, 0) FROM users WHERE id = $1::uuid`, userID, s.cfg.currency).Scan(&n)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errUserNotFound
	}
	return n, err
}

func (s *service) owned(ctx context.Context, userID string) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT block_id FROM blockworld_purchase WHERE user_id = $1::uuid ORDER BY block_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *service) rpcWallet(ctx context.Context, userID, payload string) (any, error) {
	var in struct{}
	if err := decode(payload, &in); err != nil {
		return nil, err
	}
	if err := s.ensureStartCoins(ctx, userID); err != nil {
		return nil, err
	}
	coins, err := s.balance(ctx, userID)
	if err != nil {
		return nil, err
	}
	owned, err := s.owned(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, price FROM blockworld_block WHERE is_active ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	prices := []map[string]any{}
	for rows.Next() {
		var id int
		var name string
		var price int64
		if err := rows.Scan(&id, &name, &price); err != nil {
			return nil, err
		}
		prices = append(prices, map[string]any{"id": id, "name": name, "price": price})
	}
	return map[string]any{"currency": s.cfg.currency, "coins": coins, "owned": owned, "prices": prices}, rows.Err()
}

func (s *service) rpcBuyBlock(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Blocks        []int `json:"blocks"`
		ExpectedTotal int64 `json:"expectedTotal"`
	}
	if err := decode(payload, &in); err != nil {
		return nil, err
	}
	ids := map[int]bool{}
	var blocks []int
	for _, b := range in.Blocks {
		if !ids[b] {
			ids[b] = true
			blocks = append(blocks, b)
		}
	}
	if len(blocks) == 0 || len(blocks) > 64 {
		return nil, runtime.NewError("CHOOSE_1_TO_64_BLOCKS", 3)
	}
	if err := restrict.RefuseSuspended(ctx, s.db, userID, ""); err != nil {
		return nil, err
	}
	if err := s.ensureStartCoins(ctx, userID); err != nil {
		return nil, err
	}
	type item struct {
		id    int
		name  string
		price int64
	}
	var items []item
	var total int64
	for _, b := range blocks {
		it := item{id: b}
		err := s.db.QueryRowContext(ctx, `SELECT price, name FROM blockworld_block WHERE id = $1 AND is_active`, b).Scan(&it.price, &it.name)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errBlockNotSold
		}
		if err != nil {
			return nil, err
		}
		total += it.price
		items = append(items, it)
	}
	if in.ExpectedTotal != total {
		return nil, errPriceChanged
	}
	for _, it := range items {
		var owned bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM blockworld_purchase WHERE user_id = $1::uuid AND block_id = $2)`, userID, it.id).Scan(&owned); err != nil {
			return nil, err
		}
		if owned {
			return nil, runtime.NewError("ALREADY_OWNED:"+it.name, 6)
		}
	}
	purchaseID := uuid.Must(uuid.NewV4()).String()
	names := make([]string, len(items))
	for i, it := range items {
		names[i] = it.name
	}
	updated, _, err := s.nk.WalletUpdate(ctx, userID, map[string]int64{s.cfg.currency: -total}, map[string]any{
		"module": "blockworld", "reason": "block_purchase", "sourceType": "blockworld_purchase", "sourceId": purchaseID,
		"blockworldKey": "purchase:" + purchaseID, "blocks": blocks, "blockNames": strings.Join(names, ", "),
	}, true)
	if err != nil {
		var neg *runtime.WalletNegativeError
		if errors.As(err, &neg) || strings.Contains(strings.ToLower(err.Error()), "negative") {
			return nil, errNotEnoughCoins
		}
		return nil, err
	}
	coins := updated[s.cfg.currency]
	var actions []int64
	err = s.inTx(context.WithoutCancel(ctx), func(tx *sql.Tx) error {
		actions = actions[:0]
		for i, it := range items {
			rowID := purchaseID
			ref := purchaseID
			if i > 0 {
				rowID = uuid.Must(uuid.NewV4()).String()
				ref = fmt.Sprintf("%s#%d", purchaseID, i)
			}
			res, err := tx.ExecContext(ctx, `INSERT INTO blockworld_purchase (id, user_id, block_id, price, source, source_ref) VALUES ($1::uuid, $2::uuid, $3, $4, 'shop', $5) ON CONFLICT DO NOTHING`, rowID, userID, it.id, it.price, ref)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return runtime.NewError("ALREADY_OWNED:"+it.name, 6)
			}
			var outboxID int64
			if err := tx.QueryRowContext(ctx, `INSERT INTO blockworld_outbox (user_id, kind, dedupe_key, payload) VALUES ($1::uuid, 'unlock', $2, jsonb_build_object('block', $3::int, 'source', 'shop', 'purchaseId', $4::text))
				ON CONFLICT (dedupe_key) DO UPDATE SET dedupe_key = EXCLUDED.dedupe_key RETURNING id`,
				userID, unlockKey(userID, it.id), it.id, purchaseID).Scan(&outboxID); err != nil {
				return err
			}
			actions = append(actions, outboxID)
		}
		return nil
	})
	if err != nil {
		if _, _, rerr := s.nk.WalletUpdate(context.WithoutCancel(ctx), userID, map[string]int64{s.cfg.currency: total}, map[string]any{
			"module": "blockworld", "reason": "purchase_refund", "sourceType": "blockworld_purchase", "sourceId": purchaseID, "blockworldKey": "refund:" + purchaseID,
		}, true); rerr != nil {
			s.logger.Error("BlockWorld refund of purchase %s failed, the sweep will retry: %v", purchaseID, rerr)
		}
		return nil, err
	}
	for _, it := range items {
		activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{
			UserID: userID, Kind: activity.KindPurchase, Action: "blockworld/block", RefID: purchaseID,
			Data: map[string]any{"block": it.id, "name": it.name, "price": it.price, "currency": s.cfg.currency},
		}))
	}
	return map[string]any{"purchaseId": purchaseID, "blocks": blocks, "total": total, "coins": coins, "actionIds": actions}, nil
}

func unlockKey(userID string, block int) string {
	return fmt.Sprintf("unlock:%s:%d", userID, block)
}
