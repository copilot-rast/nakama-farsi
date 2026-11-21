package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type boosterValue struct {
	Multiplier float64 `json:"multiplier"`
	Minutes    int     `json:"minutes"`
}

func parseBooster(raw json.RawMessage) (boosterValue, bool) {
	var b boosterValue
	if err := json.Unmarshal(raw, &b); err != nil {
		return b, false
	}
	b.Multiplier = math.Round(b.Multiplier*100) / 100
	if math.IsNaN(b.Multiplier) || b.Multiplier < 1.01 || b.Multiplier > 100 || b.Minutes < 1 || b.Minutes > 525600 {
		return b, false
	}
	return b, true
}

func applyBooster(ctx context.Context, tx *sql.Tx, userID string, item prizeItem, source string) error {
	b, ok := parseBooster(item.Value)
	if !ok {
		return badRequest("INVALID_VALUE", "items", "booster value is not valid")
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO economy_booster (user_id, meta_id, multiplier, expires_at, source, updated_at)
		SELECT $1::uuid, m.id, $3::numeric, now() + $4::int * INTERVAL '1 minute', $6, now() FROM economy_meta m WHERE m.key=$2
		ON CONFLICT (user_id, meta_id) DO UPDATE SET multiplier=EXCLUDED.multiplier, source=EXCLUDED.source, updated_at=now(), expires_at = CASE
			WHEN $5 = 'give' AND economy_booster.expires_at > now() THEN economy_booster.expires_at + $4::int * INTERVAL '1 minute'
			ELSE now() + $4::int * INTERVAL '1 minute' END`, userID, item.Key, b.Multiplier, b.Minutes, item.Action, "prize:"+source)
	return err
}

func activeBooster(ctx context.Context, tx *sql.Tx, userID, metaID string, now time.Time) (float64, error) {
	var m float64
	err := tx.QueryRowContext(ctx, `SELECT multiplier::float8 FROM economy_booster WHERE user_id=$1 AND meta_id=$2 AND expires_at > $3`, userID, metaID, now).Scan(&m)
	if errors.Is(err, sql.ErrNoRows) {
		return 1, nil
	}
	return m, err
}

func boostedAmount(amount, multiplier float64) float64 {
	if multiplier <= 1 || amount <= 0 {
		return amount
	}
	return math.Round(amount*multiplier*10000) / 10000
}

func (s *service) rpcBoosters(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	rows, err := s.db.QueryContext(ctx, `SELECT m.key, m.name::text, b.multiplier::float8, b.expires_at FROM economy_booster b JOIN economy_meta m ON m.id=b.meta_id
		WHERE b.user_id=$1 AND b.expires_at > $2 ORDER BY b.expires_at, m.key`, userID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var key, name string
		var mult float64
		var exp time.Time
		if err := rows.Scan(&key, &name, &mult, &exp); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"metaKey": key, "metaName": langs.Resolve(localized.Parse(name), lang), "multiplier": mult,
			"expiresAt": exp.UTC().Format(time.RFC3339), "secondsLeft": int64(math.Ceil(exp.Sub(now).Seconds()))})
	}
	return map[string]any{"code": "SUCCESS", "lang": lang, "items": items}, rows.Err()
}
