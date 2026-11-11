package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

func growthPrizes() []seedPrize {
	inviter := coinPrize("referral_inviter", 100, "۱۰۰")
	inviter.message = text{"fa": "دوستت با کد تو آمد و بازی کرد؛ ۱۰۰ سکه جایزه گرفتی", "en": "Your friend joined with your code and played; you got 100 coins", "ar": "انضم صديقك برمزك ولعب؛ حصلت على 100 عملة"}
	invited := coinPrize("referral_invited", 50, "۵۰")
	invited.message = text{"fa": "به خاطر کد معرف ۵۰ سکه هدیه گرفتی", "en": "You got 50 coins for using a referral code", "ar": "حصلت على 50 عملة لاستخدام رمز الإحالة"}
	return []seedPrize{inviter, invited}
}

type seedConfig struct {
	key, typ, description string
	value                 any
	rules                 []map[string]any
}

func seedConfigKeys(vipTag string) []seedConfig {
	return []seedConfig{
		{"stargame.banner", "text", "Banner text on the Star Catcher home screen",
			text{"fa": "به ستاره‌گیر خوش آمدی! هر روز سر بزن و جایزه بگیر.", "en": "Welcome to Star Catcher! Come back every day for a reward.", "ar": "مرحبًا بك في صائد النجوم! عُد كل يوم لتحصل على جائزة."},
			[]map[string]any{{"name": "vip", "value": text{"fa": "عضو ویژه، امروز سکه‌های بیشتری منتظرت است!", "en": "VIP, more coins are waiting for you today!", "ar": "أيها العضو المميز، المزيد من العملات بانتظارك اليوم!"}, "tagIds": []string{vipTag}, "platforms": []string{}, "countries": []string{}}}},
		{"stargame.golden_rush", "bool", "Golden rush: golden stars fall twice as often", true, []map[string]any{}},
	}
}

func seedGrowth(ctx context.Context, tx *sql.Tx, prizes map[string]string, vipTag string, report *seedReport) error {
	res, err := tx.ExecContext(ctx, `UPDATE economy_referral_settings SET referrer_prize_id=$1, referee_prize_id=$2,
		goal_event_type_id=(SELECT id FROM economy_event_type WHERE key='round_finished'), goal_measure='count', goal_value=1, updated_at=now()
		WHERE id=1 AND referrer_prize_id IS NULL AND referee_prize_id IS NULL AND goal_event_type_id IS NULL`, prizes["referral_inviter"], prizes["referral_invited"])
	if err != nil {
		return fmt.Errorf("referral settings: %w", err)
	}
	n, _ := res.RowsAffected()
	report.add("referral rule", "default", n > 0)
	for _, c := range seedConfigKeys(vipTag) {
		value, _ := json.Marshal(c.value)
		rules, _ := json.Marshal(c.rules)
		id := newID()
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO base_config_key (id, key, type, description, value, rules, updated_by) VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,'seed') ON CONFLICT (key) DO NOTHING`,
			id, c.key, c.typ, c.description, string(value), string(rules)))
		if err != nil {
			return fmt.Errorf("config %s: %w", c.key, err)
		}
		report.add("config key", c.key, ok)
		if ok {
			snap, _ := json.Marshal(map[string]any{"key": c.key, "type": c.typ, "description": c.description, "value": json.RawMessage(value), "rules": json.RawMessage(rules), "isActive": true})
			if _, err := tx.ExecContext(ctx, `INSERT INTO base_config_history (id, key_id, key, version, action, snapshot, changed_by) VALUES ($1,$2,$3,1,'create',$4::jsonb,'seed')`, newID(), id, c.key, string(snap)); err != nil {
				return fmt.Errorf("config %s history: %w", c.key, err)
			}
		}
	}
	return nil
}
