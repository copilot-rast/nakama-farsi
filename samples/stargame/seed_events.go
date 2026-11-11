package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type seedEventType struct {
	key               string
	name, description text
	client            bool
	max               any
	rate              int
	sort              int
}

type seedRule struct {
	key, event, measure, period, repeat, prize, mode string
	name, description                                text
	threshold                                        float64
	maxClaims                                        any
	sort                                             int
}

type seedMetaRule struct {
	key, event, meta, source string
	name                     text
	amount                   float64
	sort                     int
}

type seedLevel struct {
	level     int
	threshold float64
	name      text
	prize     string
}

func faDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '۰' + (r - '0')
		}
		return r
	}, s)
}

var (
	eventCoinAmounts = []int{10, 15, 20, 25, 30, 40, 50, 200}
	seedEventTypes   = []seedEventType{
		{"round_finished", text{"fa": "پایان یک دور", "en": "Round finished", "ar": "انتهاء جولة"}, text{"fa": "سرور بعد از هر دور کامل ثبت می‌کند. مقدار آن امتیاز همان دور است.", "en": "Recorded by the server after every full round. Its value is the round score.", "ar": "يسجّله الخادم بعد كل جولة كاملة. قيمته نقاط الجولة."}, false, 400, 30, 10},
		{"golden_star", text{"fa": "گرفتن ستارهٔ طلایی", "en": "Golden star caught", "ar": "التقاط نجمة ذهبية"}, text{"fa": "بازی هر بار که بازیکن یک ستارهٔ طلایی می‌گیرد، این رویداد را می‌فرستد.", "en": "Sent by the game each time the player catches a golden star.", "ar": "ترسله اللعبة كلما التقط اللاعب نجمة ذهبية."}, true, 1, 20, 11},
	}
	seedMetas = []struct {
		key  string
		name text
		def  int
	}{
		{"xp", text{"fa": "امتیاز تجربه", "en": "Experience", "ar": "نقاط الخبرة"}, 0},
		{"level", text{"fa": "سطح", "en": "Level", "ar": "المستوى"}, 1},
	}
	seedRules = []seedRule{
		{"star_daily_rounds", "round_finished", "count", "day", "period", "star_coins_20", "auto", text{"fa": "سه دور در روز", "en": "Three rounds a day", "ar": "ثلاث جولات في اليوم"}, text{"fa": "هر روز سه دور بازی کن تا ۲۰ سکه خودکار بگیری.", "en": "Play three rounds a day to get 20 coins automatically.", "ar": "العب ثلاث جولات يوميًا لتحصل على 20 عملة تلقائيًا."}, 3, nil, 1},
		{"star_golden_daily", "golden_star", "count", "day", "period", "star_coins_15", "claim", text{"fa": "۱۰ ستارهٔ طلایی در روز", "en": "10 golden stars a day", "ar": "10 نجوم ذهبية في اليوم"}, text{"fa": "هر روز ۱۰ ستارهٔ طلایی بگیر و ۱۵ سکه جایزه بردار.", "en": "Catch 10 golden stars a day and collect 15 coins.", "ar": "التقط 10 نجوم ذهبية يوميًا واجمع 15 عملة."}, 10, nil, 2},
		{"star_first_100", "round_finished", "max", "total", "once", "star_coins_50", "claim", text{"fa": "رکورد ۱۰۰ امتیاز", "en": "Score 100 in a round", "ar": "100 نقطة في جولة"}, text{"fa": "در یک دور به ۱۰۰ امتیاز برس تا ۵۰ سکه بگیری.", "en": "Reach 100 points in one round to earn 50 coins.", "ar": "اجمع 100 نقطة في جولة واحدة لتربح 50 عملة."}, 100, nil, 3},
		{"star_every_1000", "round_finished", "sum", "total", "every", "star_coins_100", "auto", text{"fa": "هر ۱۰۰۰ امتیاز", "en": "Every 1000 points", "ar": "كل 1000 نقطة"}, text{"fa": "هر بار که جمع امتیازهایت ۱۰۰۰ تا بیشتر شود، ۱۰۰ سکه می‌گیری.", "en": "Every time your total score grows by 1000, you get 100 coins.", "ar": "كلما زاد مجموع نقاطك 1000 نقطة تحصل على 100 عملة."}, 1000, 20, 4},
	}
	seedMetaRules = []seedMetaRule{
		{"star_xp_rounds", "round_finished", "xp", "event", text{"fa": "امتیاز تجربه از هر دور", "en": "Experience from every round", "ar": "خبرة من كل جولة"}, 1, 1},
		{"star_xp_golden", "golden_star", "xp", "fixed", text{"fa": "امتیاز تجربه از ستارهٔ طلایی", "en": "Experience from golden stars", "ar": "خبرة من النجوم الذهبية"}, 2, 2},
	}
	seedLevels = []seedLevel{
		{1, 0, text{"fa": "تازه‌کار", "en": "Rookie", "ar": "مبتدئ"}, ""},
		{2, 100, text{"fa": "ستاره‌چین", "en": "Star catcher", "ar": "صائد النجوم"}, "star_coins_20"},
		{3, 300, text{"fa": "شکارچی ستاره", "en": "Star hunter", "ar": "قنّاص النجوم"}, "star_coins_50"},
		{4, 600, text{"fa": "ستاره‌شناس", "en": "Stargazer", "ar": "عالم النجوم"}, "star_coins_100"},
		{5, 1000, text{"fa": "استاد آسمان", "en": "Sky master", "ar": "سيد السماء"}, "star_coins_200"},
		{6, 1600, text{"fa": "افسانهٔ کهکشان", "en": "Galaxy legend", "ar": "أسطورة المجرة"}, "star_coins_200"},
	}
	streakDays = []int{10, 15, 20, 25, 30, 40, 100}
)

func eventCoinPrizes() []seedPrize {
	out := make([]seedPrize, 0, len(eventCoinAmounts))
	for _, n := range eventCoinAmounts {
		out = append(out, coinPrize(fmt.Sprintf("star_coins_%d", n), n, faDigits(fmt.Sprint(n))))
	}
	return out
}

func idByKey(ctx context.Context, tx *sql.Tx, table, key string) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `SELECT id::text FROM `+table+` WHERE key=$1`, key).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", table, key, err)
	}
	return id, nil
}

func seedEvents(ctx context.Context, tx *sql.Tx, prizes map[string]string, report *seedReport) error {
	for _, m := range seedMetas {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_meta (id, key, name, default_value) VALUES ($1,$2,$3::jsonb,$4) ON CONFLICT DO NOTHING`, newID(), m.key, m.name.json(), m.def))
		if err != nil {
			return fmt.Errorf("meta %s: %w", m.key, err)
		}
		report.add("meta", m.key, ok)
	}
	for _, e := range seedEventTypes {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, value_max, rate_limit, sort) VALUES ($1,$2,$3::jsonb,$4::jsonb,$5,0,$6,$7,$8) ON CONFLICT DO NOTHING`,
			newID(), e.key, e.name.json(), e.description.json(), e.client, e.max, e.rate, e.sort))
		if err != nil {
			return fmt.Errorf("event type %s: %w", e.key, err)
		}
		report.add("event type", e.key, ok)
	}
	for _, r := range seedRules {
		typeID, err := idByKey(ctx, tx, "economy_event_type", r.event)
		if err != nil {
			return err
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_event_rule (id, key, name, description, event_type_id, measure, period, threshold, repeat, max_claims, prize_id, grant_mode, sort)
			VALUES ($1,$2,$3::jsonb,$4::jsonb,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT DO NOTHING`,
			newID(), r.key, r.name.json(), r.description.json(), typeID, r.measure, r.period, r.threshold, r.repeat, r.maxClaims, prizes[r.prize], r.mode, r.sort))
		if err != nil {
			return fmt.Errorf("event rule %s: %w", r.key, err)
		}
		report.add("event rule", r.key, ok)
	}
	for _, r := range seedMetaRules {
		typeID, err := idByKey(ctx, tx, "economy_event_type", r.event)
		if err != nil {
			return err
		}
		metaID, err := idByKey(ctx, tx, "economy_meta", r.meta)
		if err != nil {
			return err
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_meta_rule (id, key, name, event_type_id, meta_id, action, value_source, amount, sort) VALUES ($1,$2,$3::jsonb,$4,$5,'add',$6,$7,$8) ON CONFLICT DO NOTHING`,
			newID(), r.key, r.name.json(), typeID, metaID, r.source, r.amount, r.sort))
		if err != nil {
			return fmt.Errorf("meta rule %s: %w", r.key, err)
		}
		report.add("meta rule", r.key, ok)
	}
	xpID, err := idByKey(ctx, tx, "economy_meta", "xp")
	if err != nil {
		return err
	}
	levelID, err := idByKey(ctx, tx, "economy_meta", "level")
	if err != nil {
		return err
	}
	ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_level_table (id, key, name, description, meta_id, level_meta_id, sort) VALUES ($1,'star_levels',$2::jsonb,$3::jsonb,$4,$5,1) ON CONFLICT DO NOTHING`,
		newID(), text{"fa": "سطح‌های ستاره‌چین", "en": "Star Catcher levels", "ar": "مستويات صائد النجوم"}.json(),
		text{"fa": "با هر دور امتیاز تجربه جمع کن و سطحت را بالا ببر. هر سطح تازه یک جایزه دارد.", "en": "Collect experience every round to level up. Every new level has a prize.", "ar": "اجمع الخبرة في كل جولة لترفع مستواك. لكل مستوى جديد جائزة."}.json(), xpID, levelID))
	if err != nil {
		return fmt.Errorf("level table: %w", err)
	}
	report.add("level table", "star_levels", ok)
	if ok {
		tableID, err := idByKey(ctx, tx, "economy_level_table", "star_levels")
		if err != nil {
			return err
		}
		for _, l := range seedLevels {
			var prize any
			if l.prize != "" {
				prize = prizes[l.prize]
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_level_step (id, table_id, level, threshold, name, prize_id, grant_mode) VALUES ($1,$2,$3,$4,$5::jsonb,$6,'claim') ON CONFLICT DO NOTHING`,
				newID(), tableID, l.level, l.threshold, l.name.json(), prize); err != nil {
				return fmt.Errorf("level %d: %w", l.level, err)
			}
		}
	}
	loginID, err := idByKey(ctx, tx, "economy_event_type", "login")
	if err != nil {
		return err
	}
	ok, err = inserted(tx.ExecContext(ctx, `INSERT INTO economy_streak (id, key, name, description, event_type_id, period_days, grace_periods, miss_action, cycle_length, sort) VALUES ($1,'daily_login',$2::jsonb,$3::jsonb,$4,1,1,'reset',7,1) ON CONFLICT DO NOTHING`,
		newID(), text{"fa": "جایزهٔ ورود روزانه", "en": "Daily sign-in reward", "ar": "مكافأة الدخول اليومي"}.json(),
		text{"fa": "هر روز وارد بازی شو و جایزهٔ همان روز را بگیر. اگر یک روز جا بماند اشکالی ندارد، ولی با دو روز غیبت از روز اول شروع می‌شود.", "en": "Sign in every day and collect that day's prize. Missing one day is fine, but two days away start you again from day one.", "ar": "سجّل الدخول كل يوم واجمع جائزة اليوم. لا بأس بتفويت يوم واحد، لكن غياب يومين يعيدك إلى اليوم الأول."}.json(), loginID))
	if err != nil {
		return fmt.Errorf("streak: %w", err)
	}
	report.add("streak", "daily_login", ok)
	if ok {
		streakID, err := idByKey(ctx, tx, "economy_streak", "daily_login")
		if err != nil {
			return err
		}
		for i, n := range streakDays {
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_streak_reward (id, streak_id, day, prize_id, grant_mode) VALUES ($1,$2,$3,$4,'claim') ON CONFLICT DO NOTHING`,
				newID(), streakID, i+1, prizes[fmt.Sprintf("star_coins_%d", n)]); err != nil {
				return fmt.Errorf("streak day %d: %w", i+1, err)
			}
		}
	}
	return nil
}
