package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

type seedBlock struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

func newID() string { return uuid.Must(uuid.NewV4()).String() }

func jsonText(m map[string]string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func faDigits(n int64) string {
	s := []rune(fmt.Sprint(n))
	for i, r := range s {
		if r >= '0' && r <= '9' {
			s[i] = '۰' + (r - '0')
		}
	}
	return string(s)
}

func (s *service) seed(ctx context.Context) error {
	var blocks []seedBlock
	if err := json.Unmarshal(blocksJSON, &blocks); err != nil {
		return fmt.Errorf("blocks.json: %w", err)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, b := range blocks {
			if _, err := tx.ExecContext(ctx, `INSERT INTO blockworld_block (id, name, price) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`, b.ID, b.Name, b.Price); err != nil {
				return fmt.Errorf("block %d: %w", b.ID, err)
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_currency (id, key, name, type, default_value, sort) VALUES ($1, $2, $3::jsonb, 'soft', $4, 5) ON CONFLICT DO NOTHING`,
			newID(), s.cfg.currency, jsonText(map[string]string{"fa": "سکهٔ بلوک‌ورلد", "en": "BlockWorld coins", "ar": "عملات بلوك ورلد"}), s.cfg.startCoins); err != nil {
			return fmt.Errorf("currency: %w", err)
		}
		prizeKey := fmt.Sprintf("blockworld_win_%d", s.cfg.winPrize)
		items := fmt.Sprintf(`[{"type":"currency","key":%q,"action":"add","value":%d}]`, s.cfg.currency, s.cfg.winPrize)
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_prize (id, key, name, message, is_repeatable, items) VALUES ($1, $2, $3::jsonb, $4::jsonb, true, $5::jsonb) ON CONFLICT DO NOTHING`,
			newID(), prizeKey,
			jsonText(map[string]string{"fa": faDigits(s.cfg.winPrize) + " سکهٔ برد", "en": fmt.Sprintf("%d win coins", s.cfg.winPrize), "ar": fmt.Sprintf("%d عملة فوز", s.cfg.winPrize)}),
			jsonText(map[string]string{"fa": "برای برد در بازی سر میز " + faDigits(s.cfg.winPrize) + " سکه گرفتی", "en": fmt.Sprintf("You won a table game: %d coins", s.cfg.winPrize), "ar": fmt.Sprintf("فزت في لعبة طاولة: %d عملة", s.cfg.winPrize)}),
			items); err != nil {
			return fmt.Errorf("prize: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, value_max, rate_limit, sort) VALUES ($1, $2, $3::jsonb, $4::jsonb, false, 0, 1, 60, 30) ON CONFLICT DO NOTHING`,
			newID(), eventGameWon,
			jsonText(map[string]string{"fa": "برد در بلوک‌ورلد", "en": "BlockWorld game won", "ar": "فوز في بلوك ورلد"}),
			jsonText(map[string]string{"fa": "پل بلوک‌ورلد بعد از هر برد در بازی‌های سر میز (سنگ‌کاغذقیچی، دوز و بقیه) ثبت می‌کند.", "en": "Recorded by the BlockWorld bridge after every table game win (Rock-Paper-Scissors, Tic-Tac-Toe and the others).", "ar": "يسجّله جسر بلوك ورلد بعد كل فوز في ألعاب الطاولة."})); err != nil {
			return fmt.Errorf("event type: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_event_rule (id, key, name, description, event_type_id, measure, period, threshold, repeat, max_claims, prize_id, grant_mode, sort)
			SELECT $1, 'blockworld_every_win', $2::jsonb, $3::jsonb, t.id, 'count', 'total', 1, 'every', NULL, p.id, 'auto', 30
			FROM economy_event_type t, economy_prize p WHERE t.key = $4 AND p.key = $5 ON CONFLICT DO NOTHING`,
			newID(),
			jsonText(map[string]string{"fa": "جایزهٔ هر برد", "en": "Prize for every win", "ar": "جائزة كل فوز"}),
			jsonText(map[string]string{"fa": "هر بار که در بلوک‌ورلد یک بازی سر میز را ببری، سکه می‌گیری.", "en": "Every table game you win in BlockWorld pays coins.", "ar": "كل لعبة طاولة تفوز بها في بلوك ورلد تمنحك عملات."}),
			eventGameWon, prizeKey); err != nil {
			return fmt.Errorf("event rule: %w", err)
		}
		return seedFlag(ctx, tx, flagPrefix+"rps", true, "BlockWorld: Rock-Paper-Scissors tables are open (the bridge copies this flag to the world)")
	})
}

func seedFlag(ctx context.Context, tx *sql.Tx, key string, value bool, description string) error {
	id := newID()
	raw, _ := json.Marshal(value)
	res, err := tx.ExecContext(ctx, `INSERT INTO base_config_key (id, key, type, description, value, rules, is_active, version, updated_by) VALUES ($1, $2, 'bool', $3, $4::jsonb, '[]', true, 1, 'blockworld') ON CONFLICT (key) DO NOTHING`,
		id, key, description, string(raw))
	if err != nil {
		return fmt.Errorf("config %s: %w", key, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}
	snap, _ := json.Marshal(map[string]any{"key": key, "type": "bool", "description": description, "value": value, "rules": []any{}, "isActive": true})
	_, err = tx.ExecContext(ctx, `INSERT INTO base_config_history (id, key_id, key, version, action, snapshot, changed_by) VALUES ($1, $2, $3, 1, 'create', $4::jsonb, 'blockworld')`, newID(), id, key, string(snap))
	return err
}
