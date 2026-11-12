package main

import (
	"context"
	"database/sql"
	"fmt"
)

type seedTask struct {
	key, category, event, metric, repeat, prize, parent, group string
	name, description                                          text
	target                                                     float64
	valueMin                                                   any
	dependsOn                                                  []string
	minLevel                                                   any
	notify                                                     bool
	sort                                                       int
}

var seedTaskCategories = []struct {
	key  string
	name text
	sort int
}{
	{"star_missions_daily", text{"fa": "مأموریت روزانه", "en": "Daily missions", "ar": "مهام يومية"}, 1},
	{"star_missions_story", text{"fa": "مأموریت اصلی", "en": "Main missions", "ar": "المهام الرئيسية"}, 2},
}

var seedTaskGroup = struct {
	key               string
	name, description text
}{
	"star_hunter_chain",
	text{"fa": "زنجیرهٔ شکارچی", "en": "Hunter chain", "ar": "سلسلة الصياد"},
	text{"fa": "سه مأموریت پشت‌سرهم. هر کدام را تمام کنی و جایزه‌اش را بگیری، بعدی باز می‌شود.", "en": "Three missions in a row. Finish one and collect its prize to open the next.", "ar": "ثلاث مهام متتالية. أنهِ واحدة واجمع جائزتها لتفتح التالية."},
}

var seedTasks = []seedTask{
	{key: "star_mission_two_rounds", category: "star_missions_daily", event: "round_finished", metric: "count", repeat: "daily", prize: "star_coins_20", target: 2, sort: 1,
		name:        text{"fa": "امروز دو دور بازی کن", "en": "Play two rounds today", "ar": "العب جولتين اليوم"},
		description: text{"fa": "هر روز از نو شروع می‌شود.", "en": "Starts again every day.", "ar": "تبدأ من جديد كل يوم."}},
	{key: "star_mission_golden_five", category: "star_missions_daily", event: "golden_star", metric: "count", repeat: "daily", prize: "star_coins_15", target: 5, sort: 2,
		name:        text{"fa": "پنج ستارهٔ طلایی بگیر", "en": "Catch five golden stars", "ar": "التقط خمس نجوم ذهبية"},
		description: text{"fa": "ستاره‌های طلایی امروز شمرده می‌شوند.", "en": "Golden stars caught today count.", "ar": "تُحسب النجوم الذهبية التي تلتقطها اليوم."}},
	{key: "star_mission_week_points", category: "star_missions_daily", event: "round_finished", metric: "sum_value", repeat: "weekly", prize: "star_coins_100", target: 1000, sort: 3,
		name:        text{"fa": "این هفته ۱۰۰۰ امتیاز جمع کن", "en": "Score 1000 points this week", "ar": "اجمع 1000 نقطة هذا الأسبوع"},
		description: text{"fa": "امتیاز همهٔ دورهای این هفته روی هم حساب می‌شود.", "en": "Points from every round this week add up.", "ar": "تُجمع نقاط كل جولات هذا الأسبوع."}},
	{key: "star_mission_starter", category: "star_missions_story", prize: "star_coins_200", notify: true, sort: 4,
		name:        text{"fa": "شروع ستاره‌چین", "en": "Star catcher starter", "ar": "بداية صائد النجوم"},
		description: text{"fa": "سه کار ساده؛ هر سه را انجام بده تا جایزهٔ بزرگ را بگیری.", "en": "Three simple steps. Do all three for the big prize.", "ar": "ثلاث خطوات بسيطة. أنجزها كلها لتحصل على الجائزة الكبرى."}},
	{key: "star_mission_starter_round", parent: "star_mission_starter", event: "round_finished", metric: "count", target: 1, sort: 1,
		name: text{"fa": "یک دور کامل بازی کن", "en": "Play one full round", "ar": "العب جولة كاملة"}},
	{key: "star_mission_starter_golden", parent: "star_mission_starter", event: "golden_star", metric: "count", target: 3, sort: 2,
		name: text{"fa": "سه ستارهٔ طلایی بگیر", "en": "Catch three golden stars", "ar": "التقط ثلاث نجوم ذهبية"}},
	{key: "star_mission_starter_big", parent: "star_mission_starter", event: "round_finished", metric: "count", target: 1, valueMin: 100, sort: 3,
		name: text{"fa": "یک دور با ۱۰۰ امتیاز یا بیشتر", "en": "A round of 100 points or more", "ar": "جولة بـ 100 نقطة أو أكثر"}},
	{key: "star_mission_veteran", category: "star_missions_story", event: "round_finished", metric: "count", target: 10, prize: "star_coins_100", dependsOn: []string{"star_mission_starter"}, minLevel: 2, sort: 5,
		name:        text{"fa": "کهنه‌کار: ده دور بازی کن", "en": "Veteran: play ten rounds", "ar": "المخضرم: العب عشر جولات"},
		description: text{"fa": "بعد از «شروع ستاره‌چین» و از سطح ۲ باز می‌شود.", "en": "Opens after “Star catcher starter”, from level 2.", "ar": "تُفتح بعد «بداية صائد النجوم» ومن المستوى 2."}},
	{key: "star_chain_one", group: "star_hunter_chain", event: "round_finished", metric: "count", target: 1, prize: "star_coins_10", sort: 1,
		name: text{"fa": "شکار اول: یک دور بازی کن", "en": "First hunt: play a round", "ar": "الصيد الأول: العب جولة"}},
	{key: "star_chain_two", group: "star_hunter_chain", event: "golden_star", metric: "count", target: 2, prize: "star_coins_15", sort: 2,
		name: text{"fa": "شکار دوم: دو ستارهٔ طلایی بگیر", "en": "Second hunt: catch two golden stars", "ar": "الصيد الثاني: التقط نجمتين ذهبيتين"}},
	{key: "star_chain_three", group: "star_hunter_chain", event: "round_finished", metric: "high_value", target: 150, prize: "star_coins_30", sort: 3,
		name: text{"fa": "شکار سوم: یک دور ۱۵۰ امتیازی", "en": "Third hunt: a 150-point round", "ar": "الصيد الثالث: جولة بـ 150 نقطة"}},
}

func seedTaskData(ctx context.Context, tx *sql.Tx, prizes map[string]string, report *seedReport) error {
	cats := map[string]string{}
	for _, c := range seedTaskCategories {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO base_category (id, key_type, key, name, sort) VALUES ($1, 'task', $2, $3::jsonb, $4) ON CONFLICT DO NOTHING`, newID(), c.key, c.name.json(), c.sort))
		if err != nil {
			return fmt.Errorf("task category %s: %w", c.key, err)
		}
		report.add("task category", c.key, ok)
		var id string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='task' AND key=$1`, c.key).Scan(&id); err != nil {
			return fmt.Errorf("task category %s: %w", c.key, err)
		}
		cats[c.key] = id
	}
	ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_task_group (id, key, name, description, category_id, active_count, order_mode, handout, is_repeatable, sort) VALUES ($1,$2,$3::jsonb,$4::jsonb,$5,1,'ordered','auto',false,1) ON CONFLICT DO NOTHING`,
		newID(), seedTaskGroup.key, seedTaskGroup.name.json(), seedTaskGroup.description.json(), cats["star_missions_story"]))
	if err != nil {
		return fmt.Errorf("task group: %w", err)
	}
	report.add("task group", seedTaskGroup.key, ok)
	groupID, err := idByKey(ctx, tx, "economy_task_group", seedTaskGroup.key)
	if err != nil {
		return err
	}
	tableID, err := idByKey(ctx, tx, "economy_level_table", "star_levels")
	if err != nil {
		return err
	}
	for _, s := range seedTasks {
		var typeID, parentID, group, prize, category, levelTable any
		condition := "children"
		if s.event != "" {
			id, err := idByKey(ctx, tx, "economy_event_type", s.event)
			if err != nil {
				return err
			}
			typeID, condition = id, "event"
		}
		if s.parent != "" {
			id, err := idByKey(ctx, tx, "economy_task", s.parent)
			if err != nil {
				return err
			}
			parentID = id
		}
		if s.group != "" {
			group = groupID
		}
		if s.prize != "" {
			prize = prizes[s.prize]
		}
		if s.category != "" {
			category = cats[s.category]
		}
		if s.minLevel != nil {
			levelTable = tableID
		}
		deps := []string{}
		for _, d := range s.dependsOn {
			id, err := idByKey(ctx, tx, "economy_task", d)
			if err != nil {
				return err
			}
			deps = append(deps, id)
		}
		metric := s.metric
		if metric == "" {
			metric = "count"
		}
		repeat := s.repeat
		if repeat == "" {
			repeat = "once"
		}
		target := s.target
		if target == 0 {
			target = 1
		}
		desc := s.description
		if desc == nil {
			desc = text{}
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_task (id, key, name, description, category_id, condition, event_type_id, metric, target, value_min, repeat, parent_id, depends_on, group_id, prize_id, grant_mode, notify, level_table_id, min_level, sort)
			VALUES ($1,$2,$3::jsonb,$4::jsonb,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,'claim',$16,$17,$18,$19) ON CONFLICT DO NOTHING`,
			newID(), s.key, s.name.json(), desc.json(), category, condition, typeID, metric, target, s.valueMin, repeat, parentID, deps, group, prize, s.notify, levelTable, s.minLevel, s.sort))
		if err != nil {
			return fmt.Errorf("task %s: %w", s.key, err)
		}
		report.add("task", s.key, ok)
	}
	return nil
}
