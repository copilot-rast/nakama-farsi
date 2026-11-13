package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

const (
	starGameKey       = "stargame"
	tttGameKey        = "stargame_tictactoe"
	tttLeaderboard    = "stargame_tictactoe_wins"
	defaultRoundSecs  = 30
	defaultRoundCoins = 30
)

const starSchema = `{"type":"object","additionalProperties":false,"properties":{
	"roundSeconds":{"type":"integer","minimum":20,"maximum":120,"default":30,"title":{"fa":"زمان هر دور (ثانیه)","en":"Round time (seconds)","ar":"مدة الجولة (ثانية)"},"description":{"fa":"بازیکن دور را دست‌کم ۵ ثانیه کمتر از این زمان تمام می‌کند.","en":"A round counts once it has run at least this long minus 5 seconds.","ar":"تُحتسب الجولة بعد مرور هذه المدة ناقص 5 ثوانٍ."}},
	"maxCoins":{"type":"integer","minimum":0,"maximum":100,"default":30,"title":{"fa":"بیشترین سکهٔ هر دور","en":"Most coins per round","ar":"أكثر عملات في الجولة"}}}}`

const tttSchema = `{"type":"object","additionalProperties":false,"properties":{
	"firstMove":{"type":"string","enum":["random","first_joined"],"default":"random","title":{"fa":"نوبت اول","en":"First move","ar":"الحركة الأولى"},"description":{"fa":"random: یکی به تصادف؛ first_joined: کسی که زودتر وارد شد.","en":"random: one player by chance; first_joined: whoever joined first.","ar":"random: لاعب عشوائي؛ first_joined: من دخل أولًا."}},
	"waitSeconds":{"type":"integer","minimum":10,"maximum":300,"default":60,"title":{"fa":"صبر برای حریف (ثانیه)","en":"Wait for the opponent (seconds)","ar":"انتظار الخصم (ثانية)"}},
	"forfeitOnLeave":{"type":"boolean","default":true,"title":{"fa":"ترک بازی یعنی باخت","en":"Leaving loses the match","ar":"المغادرة تعني الخسارة"}}}}`

func registerGames(initializer runtime.Initializer) error {
	if err := gamecatalog.Register(gamecatalog.Definition{Key: starGameKey, Plugin: "stargame", Name: map[string]string{"fa": "ستاره‌چین", "en": "Star Catcher"}, Schema: json.RawMessage(starSchema)}); err != nil {
		return err
	}
	if err := gamecatalog.Register(gamecatalog.Definition{Key: tttGameKey, Plugin: "stargame", Realtime: true, Name: map[string]string{"fa": "دوز", "en": "Tic-tac-toe"}, Schema: json.RawMessage(tttSchema)}); err != nil {
		return err
	}
	return initializer.RegisterMatch(tttGameKey, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &tttMatch{}, nil
	})
}

func roundSettings(ctx context.Context, logger runtime.Logger) (int, int64, error) {
	g, err := gamecatalog.Settings(ctx, starGameKey)
	if err != nil {
		if gamecatalog.Code(err) != "GAME_NOT_FOUND" && err != gamecatalog.ErrUnavailable {
			logger.Warn("Star Catcher settings not read: %v", err)
		}
		return defaultRoundSecs, defaultRoundCoins, nil
	}
	if !g.Active {
		return 0, 0, errGameDisabled
	}
	return g.Int("roundSeconds", defaultRoundSecs), int64(g.Int("maxCoins", defaultRoundCoins)), nil
}

type seedGameRoom struct {
	key      string
	name     text
	fee      int64
	prize    string
	group    string
	minLevel int
}

type seedGameDef struct {
	key, category     string
	name, desc, maint text
	min, max          int
	turn, limit       int
	board, icon       string
	settings          string
	rooms             []seedGameRoom
	sort              int
}

var seedGameCategories = []seedCategory{
	{"star_arcade", text{"fa": "آرکید", "en": "Arcade", "ar": "أركيد"}, text{"fa": "بازی‌های تند و تنها", "en": "Quick solo games", "ar": "ألعاب فردية سريعة"}, 1},
	{"star_board", text{"fa": "تخته‌ای", "en": "Board games", "ar": "ألعاب الطاولة"}, text{"fa": "بازی‌های دونفره روی تخته", "en": "Two-player board games", "ar": "ألعاب لوحية لشخصين"}, 2},
}

var seedGames = []seedGameDef{
	{key: starGameKey, category: "star_arcade", min: 1, max: 1, limit: 30, icon: "game-stargame.svg", settings: `{"roundSeconds":30,"maxCoins":30}`, sort: 1,
		name:  text{"fa": "ستاره‌چین", "en": "Star Catcher", "ar": "صائد النجوم"},
		desc:  text{"fa": "ستاره‌ها را با سبد بگیر و از بمب‌ها دور بمان.", "en": "Catch the stars with your basket and dodge the bombs.", "ar": "التقط النجوم بسلتك وتجنّب القنابل."},
		maint: text{"fa": "ستاره‌چین چند دقیقه در حال به‌روز شدن است.", "en": "Star Catcher is being updated for a few minutes.", "ar": "يجري تحديث صائد النجوم لبضع دقائق."}},
	{key: tttGameKey, category: "star_board", min: 2, max: 2, turn: 15, limit: 300, board: tttLeaderboard, icon: "game-tictactoe.svg", settings: `{"firstMove":"random","waitSeconds":60,"forfeitOnLeave":true}`, sort: 2,
		name:  text{"fa": "دوز", "en": "Tic-tac-toe", "ar": "إكس أو"},
		desc:  text{"fa": "سه خانه در یک خط بچین و حریفت را ببر. ورودی می‌دهی و برنده جایزه می‌برد.", "en": "Line up three marks to beat your opponent. Pay the entry and the winner takes the prize.", "ar": "اصنع خطًا من ثلاث علامات لتهزم خصمك. ادفع رسوم الدخول والفائز يأخذ الجائزة."},
		maint: text{"fa": "دوز فعلاً بسته است. کمی بعد سر بزن.", "en": "Tic-tac-toe is closed for now. Come back soon.", "ar": "إكس أو مغلقة الآن. عد قريبًا."},
		rooms: []seedGameRoom{
			{key: "bronze", name: text{"fa": "برنزی", "en": "Bronze", "ar": "برونزي"}, fee: 10, prize: "star_coins_20"},
			{key: "silver", name: text{"fa": "نقره‌ای", "en": "Silver", "ar": "فضي"}, fee: 25, prize: "star_coins_50", minLevel: 2},
			{key: "gold", name: text{"fa": "طلایی", "en": "Gold", "ar": "ذهبي"}, fee: 50, group: "star_box_big", minLevel: 3},
		}},
}

func seedGameData(ctx context.Context, tx *sql.Tx, env map[string]string, prizes map[string]string, report *seedReport) error {
	images := bannerImageBase(env)
	cats := map[string]string{}
	for _, c := range seedGameCategories {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO base_category (id, key_type, key, name, description, sort) VALUES ($1, 'game', $2, $3::jsonb, $4::jsonb, $5) ON CONFLICT DO NOTHING`,
			newID(), c.key, c.name.json(), c.description.json(), c.sort))
		if err != nil {
			return fmt.Errorf("game category %s: %w", c.key, err)
		}
		report.add("game category", c.key, ok)
		var id string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='game' AND key=$1`, c.key).Scan(&id); err != nil {
			return fmt.Errorf("game category %s: %w", c.key, err)
		}
		cats[c.key] = id
	}
	for _, g := range seedGames {
		rooms := []map[string]any{}
		for _, r := range g.rooms {
			room := map[string]any{"key": r.key, "name": r.name, "minLevel": r.minLevel}
			if r.fee > 0 {
				room["entryFee"] = map[string]any{"currencyKey": currencyKey, "amount": r.fee}
			}
			if r.prize != "" {
				id, ok := prizes[r.prize]
				if !ok {
					return fmt.Errorf("game %s room %s: prize %s is missing", g.key, r.key, r.prize)
				}
				room["prizeId"] = id
			}
			if r.group != "" {
				id, err := idByKey(ctx, tx, "economy_prize_group", r.group)
				if err != nil {
					return err
				}
				room["prizeGroupId"] = id
			}
			rooms = append(rooms, room)
		}
		raw, _ := json.Marshal(rooms)
		var turn, limit, board any
		if g.turn > 0 {
			turn = g.turn
		}
		if g.limit > 0 {
			limit = g.limit
		}
		if g.board != "" {
			board = g.board
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_game (id, key, is_active, name, description, maintenance, icon_url, category_id, min_players, max_players, turn_seconds, match_seconds,
			bots_allowed, rooms, leaderboard_id, platforms, settings, sort)
			VALUES ($1,$2,true,$3::jsonb,$4::jsonb,$5::jsonb,$6,$7,$8,$9,$10,$11,false,$12::jsonb,$13,'{}',$14::jsonb,$15) ON CONFLICT DO NOTHING`,
			newID(), g.key, g.name.json(), g.desc.json(), g.maint.json(), images+g.icon, cats[g.category], g.min, g.max, turn, limit, string(raw), board, g.settings, g.sort))
		if err != nil {
			return fmt.Errorf("game %s: %w", g.key, err)
		}
		report.add("game", g.key, ok)
		if _, err := tx.ExecContext(ctx, `UPDATE economy_game SET icon_url=$2 WHERE key=$1 AND icon_url=$3 AND $2<>$3`, g.key, images+g.icon, "/stargame/banners/"+g.icon); err != nil {
			return fmt.Errorf("game %s icon: %w", g.key, err)
		}
	}
	return nil
}
