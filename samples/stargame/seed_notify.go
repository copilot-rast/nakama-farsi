package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type seedAutoRule struct {
	key, name, trigger, eventKey, category, channel string
	params                                          map[string]any
	title, body                                     text
	action                                          map[string]any
	cooldown                                        int
	quietStart, quietEnd                            *int
}

func minutesOf(h, m int) *int {
	v := h*60 + m
	return &v
}

var seedAutoRules = []seedAutoRule{
	{key: "star_auto_badge", name: "Star Catcher: a new badge", trigger: "event", eventKey: "achievement_unlocked", category: "rewards", channel: "push", cooldown: 0,
		title:  text{"fa": "نشان تازه گرفتی!", "en": "You got a new badge!", "ar": "حصلت على شارة جديدة!"},
		body:   text{"fa": "{player}، دستاورد تازه‌ات مبارک. سر بزن و جایزه‌اش را بگیر.", "en": "Well done, {player}! Come and collect its prize.", "ar": "أحسنت يا {player}! تعال واجمع جائزتها."},
		action: map[string]any{"type": "screen", "value": "achievements", "label": text{"fa": "دستاوردها", "en": "Achievements", "ar": "الإنجازات"}}},
	{key: "star_auto_wheel", name: "Star Catcher: the daily wheel is ready", trigger: "chance_ready", category: "reminders", channel: "push_in_app", cooldown: 720,
		params:     map[string]any{"gameKey": "star_wheel_daily"},
		title:      text{"fa": "گردونهٔ رایگان آماده است", "en": "Your free spin is ready", "ar": "دورتك المجانية جاهزة"},
		body:       text{"fa": "«{game}» دوباره رایگان شد. بچرخان و سکه ببر!", "en": "\"{game}\" is free again. Spin and win coins!", "ar": "«{game}» مجانية من جديد. أدر واربح العملات!"},
		action:     map[string]any{"type": "screen", "value": "chance"},
		quietStart: minutesOf(23, 0), quietEnd: minutesOf(8, 0)},
	{key: "star_auto_streak", name: "Star Catcher: the sign-in streak is about to break", trigger: "streak_at_risk", category: "reminders", channel: "push", cooldown: 1200,
		params:     map[string]any{"streakKey": "daily_login", "hoursBefore": 4},
		title:      text{"fa": "زنجیره‌ات را نگه دار", "en": "Keep your streak", "ar": "حافظ على سلسلتك"},
		body:       text{"fa": "{count} روز پشت هم آمده‌ای. تا {hours} ساعت دیگر سر بزن تا زنجیره نشکند.", "en": "You came {count} days in a row. Drop by within {hours} hours so it does not break.", "ar": "جئت {count} أيام متتالية. عد خلال {hours} ساعات حتى لا تنقطع."},
		quietStart: minutesOf(23, 0), quietEnd: minutesOf(8, 0)},
	{key: "star_auto_comeback", name: "Star Catcher: come back after three days", trigger: "inactive", category: "reminders", channel: "push", cooldown: 4320,
		params:     map[string]any{"days": 3},
		title:      text{"fa": "دلمان برایت تنگ شده", "en": "We miss you", "ar": "اشتقنا إليك"},
		body:       text{"fa": "{days} روز است که ستاره نگرفته‌ای. یک دور بازی کن!", "en": "No stars for {days} days. Play a round!", "ar": "لا نجوم منذ {days} أيام. العب جولة!"},
		quietStart: minutesOf(22, 0), quietEnd: minutesOf(9, 0)},
}

func seedNotifyData(ctx context.Context, tx *sql.Tx, report *seedReport) error {
	for _, r := range seedAutoRules {
		var catID string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM publication_notify_category WHERE key=$1`, r.category).Scan(&catID); err != nil {
			return fmt.Errorf("notification category %s: %w", r.category, err)
		}
		params, _ := json.Marshal(r.params)
		if r.params == nil {
			params = []byte("{}")
		}
		var act any
		if r.action != nil {
			b, _ := json.Marshal(r.action)
			act = string(b)
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO publication_auto_rule (id, key, name, trigger, event_key, params, channel, category_id, title, body, action, cooldown_minutes, quiet_start, quiet_end)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9::jsonb, $10::jsonb, $11::jsonb, $12, $13, $14) ON CONFLICT DO NOTHING`,
			newID(), r.key, r.name, r.trigger, r.eventKey, string(params), r.channel, catID, r.title.json(), r.body.json(), act, r.cooldown, r.quietStart, r.quietEnd))
		if err != nil {
			return fmt.Errorf("automatic notification %s: %w", r.key, err)
		}
		report.add("automatic notification", r.key, ok)
	}
	return nil
}
