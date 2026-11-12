package main

import (
	"context"
	"database/sql"
	"fmt"
)

type seedTierDef struct {
	threshold float64
	name      text
	prize     string
}

type seedAchievement struct {
	key, category, source, event, metric, group, rarity string
	name, description, hint                             text
	points, sort                                        int
	hidden, notify                                      bool
	tiers                                               []seedTierDef
}

var seedAchievementCategories = []struct {
	key  string
	name text
	sort int
}{
	{"star_ach_play", text{"fa": "بازی", "en": "Play", "ar": "اللعب"}, 1},
	{"star_ach_special", text{"fa": "ویژه", "en": "Special", "ar": "مميز"}, 2},
}

var seedAchievements = []seedAchievement{
	{key: "star_ach_first", category: "star_ach_play", source: "event", event: "round_finished", metric: "count", rarity: "common", points: 5, sort: 1,
		name:        text{"fa": "اولین دور", "en": "First round", "ar": "الجولة الأولى"},
		description: text{"fa": "یک دور کامل بازی کن.", "en": "Play one full round.", "ar": "العب جولة كاملة."},
		tiers:       []seedTierDef{{threshold: 1}}},
	{key: "star_ach_rounds", category: "star_ach_play", source: "event", event: "round_finished", metric: "count", rarity: "rare", points: 20, sort: 2,
		name:        text{"fa": "ستاره‌چین پرکار", "en": "Busy star catcher", "ar": "صائد النجوم المجتهد"},
		description: text{"fa": "دور پشت دور بازی کن و از برنزی به طلایی برس.", "en": "Keep playing rounds to go from bronze to gold.", "ar": "واصل اللعب لتنتقل من البرونز إلى الذهب."},
		tiers: []seedTierDef{
			{threshold: 3, name: text{"fa": "برنزی", "en": "Bronze", "ar": "برونزي"}, prize: "star_coins_20"},
			{threshold: 10, name: text{"fa": "نقره‌ای", "en": "Silver", "ar": "فضي"}, prize: "star_coins_50"},
			{threshold: 30, name: text{"fa": "طلایی", "en": "Gold", "ar": "ذهبي"}, prize: "star_coins_100"},
		}},
	{key: "star_ach_hunter", category: "star_ach_special", source: "task", group: "star_hunter_chain", rarity: "epic", points: 30, sort: 3,
		name:        text{"fa": "شکارچی کامل", "en": "Complete hunter", "ar": "الصياد الكامل"},
		description: text{"fa": "هر سه مأموریت «زنجیرهٔ شکارچی» را تمام کن.", "en": "Finish all three missions of the hunter chain.", "ar": "أنهِ مهام سلسلة الصياد الثلاث."},
		tiers:       []seedTierDef{{threshold: 1, prize: "star_coins_50"}}},
	{key: "star_ach_supporter", category: "star_ach_special", source: "manual", rarity: "uncommon", points: 15, sort: 4, notify: true,
		name:        text{"fa": "هوادار ستاره‌چین", "en": "Star Catcher fan", "ar": "مشجع صائد النجوم"},
		description: text{"fa": "با یک کد هدیهٔ ویژه به دست می‌آید.", "en": "Comes with a special gift code.", "ar": "يأتي مع رمز هدية خاص."},
		tiers:       []seedTierDef{{threshold: 1}}},
	{key: "star_ach_legend", category: "star_ach_play", source: "event", event: "round_finished", metric: "high_value", rarity: "legendary", points: 50, sort: 5, hidden: true,
		name:        text{"fa": "شکار افسانه‌ای", "en": "Legendary catch", "ar": "صيد أسطوري"},
		description: text{"fa": "در یک دور ۳۰۰ امتیاز یا بیشتر گرفتی.", "en": "You scored 300 or more in one round.", "ar": "سجّلت 300 نقطة أو أكثر في جولة واحدة."},
		hint:        text{"fa": "یک دور واقعاً بزرگ بازی کن.", "en": "Play a really big round.", "ar": "العب جولة كبيرة حقًا."},
		tiers:       []seedTierDef{{threshold: 300, prize: "star_coins_30"}}},
}

var seedAchievementPrize = seedPrize{key: "star_fan_badge", name: text{"fa": "نشان هوادار", "en": "Fan badge", "ar": "شارة المشجع"},
	message: text{"fa": "نشان «هوادار ستاره‌چین» و ۲۰ سکه گرفتی", "en": "You got the Star Catcher fan badge and 20 coins", "ar": "حصلت على شارة مشجع صائد النجوم و20 عملة"},
	items:   `[{"type":"achievement","key":"star_ach_supporter","action":"give","value":1},{"type":"currency","key":"coin","action":"add","value":20}]`}

const fanGiftCode = "STARFAN"

func seedAchievementData(ctx context.Context, tx *sql.Tx, prizes map[string]string, report *seedReport) error {
	cats := map[string]string{}
	for _, c := range seedAchievementCategories {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO base_category (id, key_type, key, name, sort) VALUES ($1, 'achievement', $2, $3::jsonb, $4) ON CONFLICT DO NOTHING`, newID(), c.key, c.name.json(), c.sort))
		if err != nil {
			return fmt.Errorf("achievement category %s: %w", c.key, err)
		}
		report.add("achievement category", c.key, ok)
		var id string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='achievement' AND key=$1`, c.key).Scan(&id); err != nil {
			return fmt.Errorf("achievement category %s: %w", c.key, err)
		}
		cats[c.key] = id
	}
	for _, a := range seedAchievements {
		var typeID, groupID any
		if a.event != "" {
			id, err := idByKey(ctx, tx, "economy_event_type", a.event)
			if err != nil {
				return err
			}
			typeID = id
		}
		if a.group != "" {
			id, err := idByKey(ctx, tx, "economy_task_group", a.group)
			if err != nil {
				return err
			}
			groupID = id
		}
		metric := a.metric
		if metric == "" {
			metric = "count"
		}
		hint := a.hint
		if hint == nil {
			hint = text{}
		}
		id := newID()
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_achievement (id, key, name, description, hint, category_id, rarity, points, is_hidden, source, event_type_id, metric, task_group_id, notify, sort)
			VALUES ($1,$2,$3::jsonb,$4::jsonb,$5::jsonb,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT DO NOTHING`,
			id, a.key, a.name.json(), a.description.json(), hint.json(), cats[a.category], a.rarity, a.points, a.hidden, a.source, typeID, metric, groupID, a.notify, a.sort))
		if err != nil {
			return fmt.Errorf("achievement %s: %w", a.key, err)
		}
		report.add("achievement", a.key, ok)
		if !ok {
			continue
		}
		for i, t := range a.tiers {
			var prize any
			if t.prize != "" {
				prize = prizes[t.prize]
			}
			name := t.name
			if name == nil {
				name = text{}
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_achievement_tier (id, achievement_id, tier, threshold, name, prize_id, grant_mode) VALUES ($1,$2,$3,$4,$5::jsonb,$6,'claim')`,
				newID(), id, i+1, t.threshold, name.json(), prize); err != nil {
				return fmt.Errorf("achievement %s tier %d: %w", a.key, i+1, err)
			}
		}
	}
	p := seedAchievementPrize
	ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_prize (id, key, name, message, is_repeatable, items) VALUES ($1, $2, $3::jsonb, $4::jsonb, false, $5::jsonb) ON CONFLICT DO NOTHING`,
		newID(), p.key, p.name.json(), p.message.json(), p.items))
	if err != nil {
		return fmt.Errorf("prize %s: %w", p.key, err)
	}
	report.add("prize", p.key, ok)
	prizeID, err := idByKey(ctx, tx, "economy_prize", p.key)
	if err != nil {
		return err
	}
	ok, err = inserted(tx.ExecContext(ctx, `INSERT INTO economy_gift (id, name, type, code, prize_id, is_promo, per_user_limit, is_active) VALUES ($1, $2::jsonb, 'reward', $3, $4, true, 1, true) ON CONFLICT DO NOTHING`,
		newID(), text{"fa": "ستاره‌چین: نشان هوادار", "en": "Star Catcher: fan badge", "ar": "صائد النجوم: شارة المشجع"}.json(), fanGiftCode, prizeID))
	if err != nil {
		return fmt.Errorf("gift %s: %w", fanGiftCode, err)
	}
	report.add("gift", fanGiftCode, ok)
	return nil
}
