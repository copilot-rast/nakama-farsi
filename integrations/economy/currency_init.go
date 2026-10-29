package main

import (
	"context"
	"database/sql"
)

func (s *service) ensureDefaults(ctx context.Context, userID string) error {
	rows, err := s.db.QueryContext(ctx, `SELECT c.id::text FROM economy_currency c
		WHERE c.default_value > 0 AND NOT EXISTS (SELECT 1 FROM economy_user_currency_init i WHERE i.user_id=$1 AND i.currency_id=c.id)`, userID)
	if err != nil {
		return err
	}
	var missing []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		missing = append(missing, id)
	}
	rows.Close()
	if len(missing) == 0 {
		return nil
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		changes := map[string]int64{}
		for _, id := range missing {
			var key string
			var value int64
			err := tx.QueryRowContext(ctx, `INSERT INTO economy_user_currency_init (user_id, currency_id)
				SELECT $1, id FROM economy_currency WHERE id=$2 AND default_value > 0
				ON CONFLICT (user_id, currency_id) DO NOTHING
				RETURNING (SELECT key FROM economy_currency WHERE id=$2), (SELECT default_value FROM economy_currency WHERE id=$2)`, userID, id).Scan(&key, &value)
			if err == sql.ErrNoRows {
				continue
			}
			if err != nil {
				return err
			}
			changes[key] += value
		}
		if len(changes) == 0 {
			return nil
		}
		_, _, err := s.nk.WalletUpdate(ctx, userID, changes, map[string]any{"module": "economy", "reason": "currency_default", "sourceType": "currency_default", "sourceId": userID}, true)
		return err
	})
}
