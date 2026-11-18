package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

const (
	regenMaxCap      = 1_000_000
	regenMaxInterval = 30 * 24 * 3600
	regenSourceMax   = 128
)

var regenNS = uuid.Must(uuid.FromString("8a2c6f0e-3b7d-4e51-9c2a-1d4e7f9b0a63"))

type regenQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type regenConfig struct {
	Enabled         bool                     `json:"enabled"`
	Max             int64                    `json:"max"`
	IntervalSeconds int64                    `json:"intervalSeconds"`
	Clamp           bool                     `json:"clamp"`
	RefillPrice     *economyhooks.RegenPrice `json:"refillPrice"`
	UnlimitedTag    string                   `json:"unlimitedTag"`
}

func (c *regenConfig) on() bool {
	return c != nil && c.Enabled && c.Max > 0 && c.IntervalSeconds > 0
}

func cleanRegen(ctx context.Context, tx *sql.Tx, key string, raw json.RawMessage) (any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var in regenConfig
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, badRequest("INVALID_REGEN", "regen", "regen must be an object")
	}
	in.UnlimitedTag = strings.TrimSpace(in.UnlimitedTag)
	if in.RefillPrice != nil {
		in.RefillPrice.CurrencyKey = strings.TrimSpace(in.RefillPrice.CurrencyKey)
		if in.RefillPrice.CurrencyKey == "" && in.RefillPrice.Amount == 0 {
			in.RefillPrice = nil
		}
	}
	if !in.Enabled {
		out, _ := json.Marshal(in)
		return string(out), nil
	}
	if in.Max < 1 || in.Max > regenMaxCap {
		return nil, badRequest("INVALID_REGEN_MAX", "regen.max", "max must be from 1 to 1000000")
	}
	if in.IntervalSeconds < 1 || in.IntervalSeconds > regenMaxInterval {
		return nil, badRequest("INVALID_REGEN_INTERVAL", "regen.intervalSeconds", "interval must be from 1 second to 30 days")
	}
	if p := in.RefillPrice; p != nil {
		if !currencyPattern.MatchString(p.CurrencyKey) {
			return nil, badRequest("INVALID_KEY", "regen.refillPrice.currencyKey", "invalid currency key")
		}
		if p.CurrencyKey == key {
			return nil, badRequest("REFILL_SAME_CURRENCY", "regen.refillPrice.currencyKey", "a refill cannot be paid with the same currency")
		}
		if p.Amount < 1 {
			return nil, badRequest("INVALID_PRICE", "regen.refillPrice.amount", "price must be at least 1")
		}
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, p.CurrencyKey).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, badRequest("CURRENCY_NOT_FOUND", "regen.refillPrice.currencyKey", "currency does not exist")
		}
	}
	if in.UnlimitedTag != "" {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_tag WHERE key=$1 AND key_type='user')`, in.UnlimitedTag).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, badRequest("TAG_NOT_FOUND", "regen.unlimitedTag", "no player tag with this key")
		}
	}
	out, _ := json.Marshal(in)
	return string(out), nil
}

func (s *service) regenConfigs(ctx context.Context, q regenQueryer, keys []string) (map[string]*regenConfig, []string, error) {
	query := `SELECT key, regen::text FROM economy_currency WHERE regen IS NOT NULL AND regen->>'enabled' = 'true' ORDER BY sort, key`
	var args []any
	if len(keys) > 0 {
		query = `SELECT key, COALESCE(regen::text,'') FROM economy_currency WHERE key = ANY($1) ORDER BY sort, key`
		args = append(args, keys)
	}
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	out := map[string]*regenConfig{}
	var order []string
	for rows.Next() {
		var key, raw string
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, nil, err
		}
		var cfg *regenConfig
		if raw != "" {
			cfg = &regenConfig{}
			if err := json.Unmarshal([]byte(raw), cfg); err != nil {
				cfg = nil
			}
		}
		out[key] = cfg
		order = append(order, key)
	}
	return out, order, rows.Err()
}

func (s *service) regenConfigOf(ctx context.Context, q regenQueryer, key string) (*regenConfig, error) {
	m, _, err := s.regenConfigs(ctx, q, []string{key})
	if err != nil {
		return nil, err
	}
	return m[key], nil
}

type regenState struct {
	key    string
	cfg    *regenConfig
	value  int64
	anchor time.Time
	until  *time.Time
	free   bool
	now    time.Time
}

func (st *regenState) view() economyhooks.Regen {
	v := economyhooks.Regen{CurrencyKey: st.key, Value: st.value, Unlimited: st.free, UnlimitedUntil: st.until}
	if !st.cfg.on() {
		return v
	}
	v.Regen, v.Max, v.IntervalSeconds = true, st.cfg.Max, st.cfg.IntervalSeconds
	if st.cfg.RefillPrice != nil {
		p := *st.cfg.RefillPrice
		v.RefillPrice = &p
	}
	if st.value < st.cfg.Max {
		step := time.Duration(st.cfg.IntervalSeconds) * time.Second
		next := st.anchor.Add(step).UTC()
		full := st.anchor.Add(time.Duration(st.cfg.Max-st.value) * step).UTC()
		v.NextAt, v.FullAt = &next, &full
	}
	return v
}

func lockWallet(ctx context.Context, tx *sql.Tx, userID string) (map[string]int64, error) {
	var raw []byte
	err := tx.QueryRowContext(ctx, `SELECT wallet FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("USER_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	wallet := map[string]int64{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &wallet); err != nil {
			return nil, err
		}
	}
	return wallet, nil
}

func (s *service) regenSettle(ctx context.Context, tx *sql.Tx, userID, key string, cfg *regenConfig) (*regenState, error) {
	wallet, err := lockWallet(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	st := &regenState{key: key, cfg: cfg, value: wallet[key], anchor: now, now: now}
	if cfg != nil && cfg.UnlimitedTag != "" {
		var until sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT ut.expires_at FROM base_user_tag ut JOIN base_tag t ON t.id=ut.tag_id
			WHERE ut.user_id=$1 AND t.key=$2 AND t.key_type='user' AND (ut.expires_at IS NULL OR ut.expires_at > $3)`, userID, cfg.UnlimitedTag, now).Scan(&until)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			st.free = true
			if until.Valid {
				u := until.Time.UTC()
				st.until = &u
			}
		}
	}
	if !cfg.on() {
		return st, nil
	}
	var anchor time.Time
	var last int64
	err = tx.QueryRowContext(ctx, `SELECT anchor_at, last_value FROM economy_user_regen WHERE user_id=$1 AND currency_key=$2 FOR UPDATE`, userID, key).Scan(&anchor, &last)
	if errors.Is(err, sql.ErrNoRows) {
		anchor, last = now, st.value
	} else if err != nil {
		return nil, err
	}
	if last >= cfg.Max && st.value < cfg.Max {
		anchor = now
	}
	if anchor.After(now) {
		anchor = now
	}
	var change int64
	reason := "regen"
	if st.value >= cfg.Max {
		if cfg.Clamp && st.value > cfg.Max {
			change, reason = cfg.Max-st.value, "regen_clamp"
		}
		anchor = now
	} else {
		step := time.Duration(cfg.IntervalSeconds) * time.Second
		n := int64(now.Sub(anchor) / step)
		gain := n
		if gain > cfg.Max-st.value {
			gain = cfg.Max - st.value
		}
		change = gain
		if st.value+gain >= cfg.Max {
			anchor = now
		} else {
			anchor = anchor.Add(time.Duration(n) * step)
		}
	}
	if change != 0 {
		meta := map[string]any{"reason": reason, "sourceType": "regen", "sourceId": key}
		if err := s.walletChange(ctx, tx, userID, uuid.Must(uuid.NewV4()).String(), map[string]int64{key: change}, meta); err != nil {
			return nil, err
		}
		st.value += change
	}
	st.anchor = anchor
	return st, s.regenSave(ctx, tx, userID, st)
}

func (s *service) regenSave(ctx context.Context, tx *sql.Tx, userID string, st *regenState) error {
	if !st.cfg.on() {
		return nil
	}
	if st.value >= st.cfg.Max {
		st.anchor = st.now
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO economy_user_regen (user_id, currency_key, anchor_at, last_value, updated_at) VALUES ($1,$2,$3,$4,now())
		ON CONFLICT (user_id, currency_key) DO UPDATE SET anchor_at=EXCLUDED.anchor_at, last_value=EXCLUDED.last_value, updated_at=now()`, userID, st.key, st.anchor, st.value)
	return err
}

func (s *service) regenViews(ctx context.Context, userID string, keys []string) ([]economyhooks.Regen, error) {
	cfgs, order, err := s.regenConfigs(ctx, s.db, keys)
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		if _, ok := cfgs[k]; !ok {
			return nil, badRequest("CURRENCY_NOT_FOUND", "currencies", "currency "+k+" does not exist")
		}
	}
	out := []economyhooks.Regen{}
	if len(order) == 0 {
		return out, nil
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		out = out[:0]
		for _, k := range order {
			st, err := s.regenSettle(ctx, tx, userID, k, cfgs[k])
			if err != nil {
				return err
			}
			out = append(out, st.view())
		}
		return nil
	})
	return out, err
}

func (s *service) regenView(ctx context.Context, userID, key string) (economyhooks.Regen, error) {
	key = strings.TrimSpace(key)
	if !currencyPattern.MatchString(key) {
		return economyhooks.Regen{}, badRequest("INVALID_KEY", "currencyKey", "invalid currency key")
	}
	cfg, err := s.regenConfigOf(ctx, s.db, key)
	if err != nil {
		return economyhooks.Regen{}, err
	}
	var view economyhooks.Regen
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		st, err := s.regenSettle(ctx, tx, userID, key, cfg)
		if err != nil {
			return err
		}
		view = st.view()
		return nil
	})
	return view, err
}

func regenRefusal(view economyhooks.Regen, amount int64) *economyhooks.Refusal {
	meta := map[string]any{"currencyKey": view.CurrencyKey, "value": view.Value, "amount": amount, "nextAt": view.NextAt, "fullAt": view.FullAt}
	if view.RefillPrice != nil {
		meta["refillPrice"] = view.RefillPrice
	}
	return &economyhooks.Refusal{Status: http.StatusConflict, Code: "NOT_ENOUGH_CURRENCY", Message: "not enough " + view.CurrencyKey, Meta: meta}
}

func cleanSource(sourceID string) (string, error) {
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" || len(sourceID) > regenSourceMax {
		return "", badRequest("INVALID_SOURCE", "sourceId", "send a source id of 1-128 characters")
	}
	return sourceID, nil
}

func (s *service) spendRegen(ctx context.Context, userID, key string, amount int64, sourceID string) (economyhooks.RegenSpend, error) {
	key = strings.TrimSpace(key)
	if !currencyPattern.MatchString(key) {
		return economyhooks.RegenSpend{}, badRequest("INVALID_KEY", "currencyKey", "invalid currency key")
	}
	if amount < 1 {
		return economyhooks.RegenSpend{}, badRequest("INVALID_AMOUNT", "amount", "amount must be at least 1")
	}
	sourceID, err := cleanSource(sourceID)
	if err != nil {
		return economyhooks.RegenSpend{}, err
	}
	cfg, err := s.regenConfigOf(ctx, s.db, key)
	if err != nil {
		return economyhooks.RegenSpend{}, err
	}
	var out economyhooks.RegenSpend
	var refused *economyhooks.Refusal
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		out, refused = economyhooks.RegenSpend{}, nil
		st, err := s.regenSettle(ctx, tx, userID, key, cfg)
		if err != nil {
			return err
		}
		var prevKey string
		var prevAmount int64
		var prevFree bool
		err = tx.QueryRowContext(ctx, `SELECT currency_key, amount, free FROM economy_regen_spend WHERE user_id=$1 AND source_id=$2`, userID, sourceID).Scan(&prevKey, &prevAmount, &prevFree)
		if err == nil {
			if prevKey != key {
				return conflict("SOURCE_REUSED", "this source id already spent another currency")
			}
			out = economyhooks.RegenSpend{View: st.view(), Spent: prevAmount, Free: prevFree, Replayed: true}
			if prevFree {
				out.Spent = 0
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if st.free {
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_regen_spend (user_id, source_id, currency_key, amount, free, created_at) VALUES ($1,$2,$3,$4,true,$5)`, userID, sourceID, key, amount, st.now); err != nil {
				return err
			}
			out = economyhooks.RegenSpend{View: st.view(), Free: true}
			return nil
		}
		if st.value < amount {
			refused = regenRefusal(st.view(), amount)
			out = economyhooks.RegenSpend{View: st.view()}
			return nil
		}
		meta := map[string]any{"reason": "regen_spend", "sourceType": "regen", "sourceId": sourceID}
		if err := s.walletChange(ctx, tx, userID, uuid.NewV5(regenNS, "spend:"+userID+":"+sourceID).String(), map[string]int64{key: -amount}, meta); err != nil {
			return err
		}
		st.value -= amount
		if err := s.regenSave(ctx, tx, userID, st); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_regen_spend (user_id, source_id, currency_key, amount, created_at) VALUES ($1,$2,$3,$4,$5)`, userID, sourceID, key, amount, st.now); err != nil {
			return err
		}
		out = economyhooks.RegenSpend{View: st.view(), Spent: amount}
		return nil
	})
	if err != nil {
		return economyhooks.RegenSpend{}, err
	}
	if refused != nil {
		return out, refused
	}
	return out, nil
}

func regenHookError(err error) error {
	if app, ok := asApp(err); ok {
		return &economyhooks.Refusal{Status: app.Status, Code: app.Code, Message: app.Message, Meta: app.Meta}
	}
	return err
}

func (s *service) regenViewHook(ctx context.Context, userID, key string) (economyhooks.Regen, error) {
	v, err := s.regenView(ctx, userID, key)
	return v, regenHookError(err)
}

func (s *service) spendRegenHook(ctx context.Context, userID, key string, amount int64, sourceID string) (economyhooks.RegenSpend, error) {
	v, err := s.spendRegen(ctx, userID, key, amount, sourceID)
	return v, regenHookError(err)
}

func (s *service) refundRegenHook(ctx context.Context, userID, key, sourceID string) (economyhooks.RegenRefund, error) {
	v, err := s.refundRegen(ctx, userID, key, sourceID)
	return v, regenHookError(err)
}

func (s *service) refundRegen(ctx context.Context, userID, key, sourceID string) (economyhooks.RegenRefund, error) {
	key = strings.TrimSpace(key)
	sourceID, err := cleanSource(sourceID)
	if err != nil {
		return economyhooks.RegenRefund{}, err
	}
	var out economyhooks.RegenRefund
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		out = economyhooks.RegenRefund{}
		var rowKey string
		var amount int64
		var free bool
		var refunded sql.NullInt64
		err := tx.QueryRowContext(ctx, `SELECT currency_key, amount, free, refunded FROM economy_regen_spend WHERE user_id=$1 AND source_id=$2 FOR UPDATE`, userID, sourceID).Scan(&rowKey, &amount, &free, &refunded)
		if errors.Is(err, sql.ErrNoRows) {
			if key == "" {
				return nil
			}
			cfg, err := s.regenConfigOf(ctx, tx, key)
			if err != nil {
				return err
			}
			st, err := s.regenSettle(ctx, tx, userID, key, cfg)
			if err != nil {
				return err
			}
			out.View = st.view()
			return nil
		}
		if err != nil {
			return err
		}
		if key != "" && key != rowKey {
			return conflict("SOURCE_CURRENCY_MISMATCH", "this source id spent another currency")
		}
		cfg, err := s.regenConfigOf(ctx, tx, rowKey)
		if err != nil {
			return err
		}
		st, err := s.regenSettle(ctx, tx, userID, rowKey, cfg)
		if err != nil {
			return err
		}
		out.Found = true
		if refunded.Valid {
			out.View, out.Refunded, out.Replayed = st.view(), refunded.Int64, true
			return nil
		}
		credit := amount
		if free {
			credit = 0
		}
		if cfg.on() && st.value+credit > cfg.Max {
			credit = cfg.Max - st.value
			if credit < 0 {
				credit = 0
			}
		}
		if credit > 0 {
			meta := map[string]any{"reason": "regen_refund", "sourceType": "regen", "sourceId": sourceID}
			if err := s.walletChange(ctx, tx, userID, uuid.NewV5(regenNS, "refund:"+userID+":"+sourceID).String(), map[string]int64{rowKey: credit}, meta); err != nil {
				return err
			}
			st.value += credit
			if err := s.regenSave(ctx, tx, userID, st); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_regen_spend SET refunded=$3, refunded_at=$4 WHERE user_id=$1 AND source_id=$2`, userID, sourceID, credit, st.now); err != nil {
			return err
		}
		out.View, out.Refunded = st.view(), credit
		return nil
	})
	return out, err
}

func (s *service) regenRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_regen_state":  s.rpcRegenState,
		"economy_regen_refill": s.rpcRegenRefill,
	}
}

func (s *service) rpcRegenState(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Currencies []string `json:"currencies"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if len(in.Currencies) > 20 {
		return nil, badRequest("TOO_MANY_CURRENCIES", "currencies", "at most 20 currencies")
	}
	keys := []string{}
	for _, k := range in.Currencies {
		k = strings.TrimSpace(k)
		if !currencyPattern.MatchString(k) {
			return nil, badRequest("INVALID_KEY", "currencies", "invalid currency key")
		}
		keys = append(keys, k)
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	views, err := s.regenViews(ctx, userID, keys)
	if err != nil {
		return nil, err
	}
	return map[string]any{"serverTime": s.now().UTC(), "currencies": views}, nil
}

func (s *service) rpcRegenRefill(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		CurrencyKey string `json:"currencyKey"`
		RequestID   string `json:"requestId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(in.CurrencyKey)
	if !currencyPattern.MatchString(key) {
		return nil, badRequest("INVALID_KEY", "currencyKey", "invalid currency key")
	}
	requestID, _, err := cleanRequestID(in.RequestID, true)
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	cfg, err := s.regenConfigOf(ctx, s.db, key)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		out = nil
		var prevKey, prev string
		err := tx.QueryRowContext(ctx, `SELECT currency_key, answer::text FROM economy_regen_refill WHERE user_id=$1 AND request_id=$2`, userID, requestID).Scan(&prevKey, &prev)
		if err == nil {
			if prevKey != key {
				return conflict("REQUEST_ID_REUSED", "this requestId was used for another currency")
			}
			out = map[string]any{}
			_ = json.Unmarshal([]byte(prev), &out)
			out["replayed"] = true
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if !cfg.on() {
			return conflict("REGEN_OFF", "this currency does not regenerate")
		}
		if cfg.RefillPrice == nil {
			return conflict("REFILL_OFF", "this currency has no refill price")
		}
		st, err := s.regenSettle(ctx, tx, userID, key, cfg)
		if err != nil {
			return err
		}
		if st.value >= cfg.Max {
			return conflict("ALREADY_FULL", "the currency is already full")
		}
		fill := cfg.Max - st.value
		price := *cfg.RefillPrice
		meta := map[string]any{"reason": "regen_refill", "sourceType": "regen", "sourceId": key, "requestId": requestID}
		if err := s.walletChange(ctx, tx, userID, uuid.NewV5(regenNS, "refill:"+userID+":"+requestID).String(), map[string]int64{price.CurrencyKey: -price.Amount, key: fill}, meta); err != nil {
			return err
		}
		st.value += fill
		if err := s.regenSave(ctx, tx, userID, st); err != nil {
			return err
		}
		wallet, err := lockWallet(ctx, tx, userID)
		if err != nil {
			return err
		}
		out = map[string]any{"currencyKey": key, "filled": fill, "paid": price, "state": st.view(), "wallet": wallet, "replayed": false}
		raw, _ := json.Marshal(out)
		_, err = tx.ExecContext(ctx, `INSERT INTO economy_regen_refill (user_id, request_id, currency_key, answer, created_at) VALUES ($1,$2,$3,$4,$5)`, userID, requestID, key, string(raw), st.now)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *service) catalogHook(ctx context.Context, userID, lang string) (map[string]any, error) {
	raw, _ := json.Marshal(map[string]string{"lang": lang})
	out, err := s.rpcCatalog(ctx, userID, string(raw))
	if err != nil {
		return nil, err
	}
	m, _ := out.(map[string]any)
	return m, nil
}

func (s *service) balancesHook(ctx context.Context, userID string) (map[string]any, error) {
	out, err := s.rpcBalances(ctx, userID, "{}")
	if err != nil {
		return nil, err
	}
	m, _ := out.(map[string]any)
	return m, nil
}
