package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type seedGroupMember struct {
	prize  string
	weight int
}

type seedPrizeGroup struct {
	key        string
	name, desc text
	members    []seedGroupMember
	sort       int
}

type seedChanceRow struct {
	chance float64
	prize  string
	group  string
}

type seedChanceGame struct {
	key        string
	name, desc text
	prizeType  string
	interval   int
	price      int64
	rows       []seedChanceRow
	metadata   string
	sort       int
}

var seedPrizeGroups = []seedPrizeGroup{
	{key: "star_box_coins", sort: 1, name: text{"fa": "کیسهٔ سکه", "en": "Coin pouch", "ar": "كيس العملات"},
		desc:    text{"fa": "۱۵، ۲۵ یا ۴۰ سکه؛ هرچه بیشتر، کمیاب‌تر.", "en": "15, 25 or 40 coins; the bigger, the rarer.", "ar": "15 أو 25 أو 40 عملة؛ كلما زادت كانت أندر."},
		members: []seedGroupMember{{"star_coins_15", 3}, {"star_coins_25", 2}, {"star_coins_40", 1}}},
	{key: "star_box_big", sort: 2, name: text{"fa": "گنج بزرگ", "en": "Big treasure", "ar": "الكنز الكبير"},
		desc:    text{"fa": "۱۰۰ یا ۲۰۰ سکه.", "en": "100 or 200 coins.", "ar": "100 أو 200 عملة."},
		members: []seedGroupMember{{"star_coins_100", 2}, {"star_coins_200", 1}}},
}

var seedChanceGames = []seedChanceGame{
	{key: "star_wheel_daily", sort: 1, prizeType: "prize", interval: 1440, metadata: `{"style":"wheel"}`,
		name: text{"fa": "گردونهٔ روزانه", "en": "Daily wheel", "ar": "العجلة اليومية"},
		desc: text{"fa": "روزی یک بار رایگان بچرخان و سکه ببر.", "en": "Spin free once a day and win coins.", "ar": "أدر العجلة مجانًا مرة في اليوم واربح عملات."},
		rows: []seedChanceRow{{chance: 35, prize: "star_coins_10"}, {chance: 25, prize: "star_coins_20"}, {chance: 10, prize: "star_coins_50"}, {chance: 30}}},
	{key: "star_lucky_box", sort: 2, prizeType: "prizeGroup", price: 15, metadata: `{"style":"box"}`,
		name: text{"fa": "جعبهٔ شانس", "en": "Lucky box", "ar": "صندوق الحظ"},
		desc: text{"fa": "با ۱۵ سکه یک جعبه باز کن؛ شاید گنج بزرگ نصیبت شود.", "en": "Open a box for 15 coins; the big treasure might be yours.", "ar": "افتح صندوقًا مقابل 15 عملة؛ قد يكون الكنز الكبير من نصيبك."},
		rows: []seedChanceRow{{chance: 55, group: "star_box_coins"}, {chance: 10, group: "star_box_big"}, {chance: 35}}},
}

func seedChanceData(ctx context.Context, tx *sql.Tx, prizes map[string]string, report *seedReport) error {
	groups := map[string]string{}
	for _, g := range seedPrizeGroups {
		id := newID()
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_prize_group (id, key, name, description, sort) VALUES ($1,$2,$3::jsonb,$4::jsonb,$5) ON CONFLICT DO NOTHING`,
			id, g.key, g.name.json(), g.desc.json(), g.sort))
		if err != nil {
			return fmt.Errorf("prize group %s: %w", g.key, err)
		}
		report.add("prize group", g.key, ok)
		if ok {
			for i, m := range g.members {
				prize, found := prizes[m.prize]
				if !found {
					return fmt.Errorf("prize group %s: prize %s is missing", g.key, m.prize)
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO economy_prize_group_member (group_id, prize_id, weight, sort) VALUES ($1,$2,$3,$4)`, id, prize, m.weight, i); err != nil {
					return fmt.Errorf("prize group %s member %s: %w", g.key, m.prize, err)
				}
			}
		}
		gid, err := idByKey(ctx, tx, "economy_prize_group", g.key)
		if err != nil {
			return err
		}
		groups[g.key] = gid
	}
	coin, err := idByKey(ctx, tx, "economy_currency", "coin")
	if err != nil {
		return err
	}
	for _, g := range seedChanceGames {
		rows := make([]map[string]any, 0, len(g.rows))
		for _, r := range g.rows {
			row := map[string]any{"chance": r.chance, "prizeId": nil, "prizeGroupId": nil}
			if r.prize != "" {
				id, found := prizes[r.prize]
				if !found {
					return fmt.Errorf("chance game %s: prize %s is missing", g.key, r.prize)
				}
				row["prizeId"] = id
			}
			if r.group != "" {
				row["prizeGroupId"] = groups[r.group]
			}
			rows = append(rows, row)
		}
		rawRows, _ := json.Marshal(rows)
		var interval, price any
		if g.interval > 0 {
			interval = g.interval
		}
		if g.price > 0 {
			raw, _ := json.Marshal(map[string]any{"currencyId": coin, "amount": g.price})
			price = string(raw)
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_chance_game (id, key, name, description, metadata, prize_type, prize, interval_time, price, is_active, sort)
			VALUES ($1,$2,$3::jsonb,$4::jsonb,$5::jsonb,$6,$7::jsonb,$8,$9::jsonb,true,$10) ON CONFLICT DO NOTHING`,
			newID(), g.key, g.name.json(), g.desc.json(), g.metadata, g.prizeType, string(rawRows), interval, price, g.sort))
		if err != nil {
			return fmt.Errorf("chance game %s: %w", g.key, err)
		}
		report.add("chance game", g.key, ok)
	}
	return nil
}
