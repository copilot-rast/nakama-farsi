package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

type gameStatsView struct {
	Game            string           `json:"game"`
	Games           int              `json:"totalGames"`
	Wins            int              `json:"wins"`
	Losses          int              `json:"losses"`
	Forfeits        int              `json:"forfeits"`
	Timeouts        int              `json:"timeouts"`
	WinRate         float64          `json:"winRate"`
	LossRate        float64          `json:"lossRate"`
	PlaytimeSeconds int64            `json:"playtimeSeconds"`
	Winnings        map[string]int64 `json:"winnings"`
	LastPlayedAt    *string          `json:"lastPlayedAt"`
}

func rate(n, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(n) / float64(total)
}

func (s *service) statsOf(ctx context.Context, userID, game string) ([]gameStatsView, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT game_key, games, wins, losses, forfeits, timeouts, playtime_seconds, winnings::text, to_char(last_played_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM economy_game_stats WHERE user_id::text=$1 AND ($2='' OR game_key=$2) ORDER BY game_key`, userID, game)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []gameStatsView{}
	for rows.Next() {
		var v gameStatsView
		var win string
		var last sql.NullString
		if err := rows.Scan(&v.Game, &v.Games, &v.Wins, &v.Losses, &v.Forfeits, &v.Timeouts, &v.PlaytimeSeconds, &win, &last); err != nil {
			return nil, err
		}
		v.Winnings = map[string]int64{}
		_ = json.Unmarshal([]byte(win), &v.Winnings)
		v.WinRate, v.LossRate = rate(v.Wins, v.Games), rate(v.Losses, v.Games)
		if last.Valid {
			v.LastPlayedAt = &last.String
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if game != "" && len(out) == 0 {
		out = append(out, gameStatsView{Game: game, Winnings: map[string]int64{}})
	}
	return out, nil
}

func (s *service) rpcGameStats(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserID string `json:"userId"`
		Game   string `json:"game"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	target := strings.TrimSpace(in.UserID)
	if target == "" {
		target = userID
	}
	if _, err := uuid.FromString(target); err != nil {
		return nil, badRequest("INVALID_ID", "userId", "must be a UUID")
	}
	items, err := s.statsOf(ctx, target, strings.TrimSpace(in.Game))
	if err != nil {
		return nil, err
	}
	return map[string]any{"userId": target, "items": items}, nil
}

func (s *service) mainLevelTable(ctx context.Context) string {
	var key string
	_ = s.db.QueryRowContext(ctx, `SELECT key FROM economy_level_table WHERE is_active ORDER BY sort, created_at LIMIT 1`).Scan(&key)
	return key
}

func (s *service) levelView(ctx context.Context, userID, table string, botLevel int) map[string]any {
	if table == "" {
		table = s.mainLevelTable(ctx)
	}
	out := map[string]any{"table": table, "level": botLevel, "xp": 0, "levelStart": 0, "next": nil, "maxLevel": 0}
	if table == "" {
		return out
	}
	if botLevel > 0 {
		var start sql.NullFloat64
		var next sql.NullFloat64
		var max int
		_ = s.db.QueryRowContext(ctx, `SELECT (SELECT threshold::float8 FROM economy_level_step s WHERE s.table_id=t.id AND s.level=$2), (SELECT min(threshold)::float8 FROM economy_level_step s WHERE s.table_id=t.id AND s.level>$2),
			COALESCE((SELECT max(level) FROM economy_level_step s WHERE s.table_id=t.id),0) FROM economy_level_table t WHERE t.key=$1`, table, botLevel).Scan(&start, &next, &max)
		out["xp"], out["levelStart"], out["maxLevel"] = start.Float64, start.Float64, max
		if next.Valid {
			out["next"] = next.Float64
			if s.stealthPool(ctx) && next.Float64 > start.Float64 {
				out["xp"] = math.Floor(start.Float64 + (next.Float64-start.Float64)*botXPFraction(userID))
			}
		}
		return out
	}
	p, err := s.levelProgressHook(ctx, userID, table)
	if err != nil {
		return out
	}
	out["level"], out["xp"], out["levelStart"], out["maxLevel"] = p.Level, p.Value, p.LevelStart, p.MaxLevel
	if p.Next != nil {
		out["next"] = *p.Next
	}
	return out
}

func (s *service) rpcPlayerCard(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserID     string `json:"userId"`
		Game       string `json:"game"`
		LevelTable string `json:"levelTable"`
		Lang       string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	target := strings.TrimSpace(in.UserID)
	if target == "" {
		target = userID
	}
	if _, err := uuid.FromString(target); err != nil {
		return nil, badRequest("INVALID_ID", "userId", "must be a UUID")
	}
	return s.playerCard(ctx, userID, target, strings.TrimSpace(in.Game), strings.TrimSpace(in.LevelTable), in.Lang)
}

func (s *service) playerCard(ctx context.Context, viewer, target, game, table, lang string) (map[string]any, error) {
	var username, display, avatar, metadata string
	var botLevel sql.NullInt64
	var look sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT u.username, COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), COALESCE(u.metadata::text,'{}'), b.level, b.look::text
		FROM users u LEFT JOIN economy_bot b ON b.user_id=u.id WHERE u.id=$1`, target).Scan(&username, &display, &avatar, &metadata, &botLevel, &look)
	if errors.Is(err, sql.ErrNoRows) {
		return s.ghostCard(ctx, viewer, target, game, table, lang)
	}
	if err != nil {
		return nil, err
	}
	langs, l, err := s.gameLangs(ctx, viewer, lang)
	if err != nil {
		return nil, err
	}
	slots := map[string]string{}
	if look.Valid {
		_ = json.Unmarshal([]byte(look.String), &slots)
	}
	if len(slots) == 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT u.slot, i.key FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1 AND u.equipped AND u.slot IS NOT NULL AND u.slot<>''`, target)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var slot, key string
			if rows.Scan(&slot, &key) == nil {
				slots[slot] = key
			}
		}
		rows.Close()
	}
	equipped := s.equippedView(ctx, slots, langs, l)
	stats, err := s.statsOf(ctx, target, game)
	if err != nil {
		return nil, err
	}
	if display == "" {
		display = username
	}
	card := map[string]any{"userId": target, "username": username, "displayName": display, "avatarUrl": avatar, "equipped": equipped,
		"level": s.levelView(ctx, target, table, int(botLevel.Int64)), "stats": stats, "lang": l}
	md := map[string]any{}
	if json.Unmarshal([]byte(metadata), &md) == nil {
		if v, ok := md["avatar"]; ok {
			card["avatar"] = v
		}
	}
	if game != "" && len(stats) == 1 {
		card["gameStats"] = stats[0]
	}
	return card, nil
}

func (s *service) equippedView(ctx context.Context, slots map[string]string, langs *localized.Set, l string) map[string]any {
	equipped := map[string]any{}
	for slot, key := range slots {
		v := map[string]any{"key": key, "name": key, "iconUrl": ""}
		var name, icon string
		if s.db.QueryRowContext(ctx, `SELECT name::text, COALESCE(icon_url,'') FROM economy_item WHERE key=$1`, key).Scan(&name, &icon) == nil {
			v["name"], v["iconUrl"] = langs.Resolve(parseI18n(name), l), icon
		}
		equipped[slot] = v
	}
	return equipped
}

func (s *service) consoleGameStats(w http.ResponseWriter, r *http.Request) {
	user := strings.TrimSpace(r.URL.Query().Get("userId"))
	if _, err := uuid.FromString(user); err != nil {
		s.writeError(w, badRequest("INVALID_ID", "userId", "must be a UUID"))
		return
	}
	items, err := s.statsOf(r.Context(), user, strings.TrimSpace(r.URL.Query().Get("game")))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"userId": user, "items": items})
}

func (s *service) consoleGameStatsSet(w http.ResponseWriter, r *http.Request) {
	user, game := pathVar(r, "userId"), pathVar(r, "game")
	if _, err := uuid.FromString(user); err != nil {
		s.writeError(w, badRequest("INVALID_ID", "userId", "must be a UUID"))
		return
	}
	var in struct {
		Wins            int              `json:"wins"`
		Losses          int              `json:"losses"`
		Forfeits        int              `json:"forfeits"`
		Timeouts        int              `json:"timeouts"`
		PlaytimeSeconds int64            `json:"playtimeSeconds"`
		Winnings        map[string]int64 `json:"winnings"`
	}
	if err := pluginkit.ReadBody(r, &in, 8192); err != nil {
		s.writeError(w, err)
		return
	}
	for name, v := range map[string]int64{"wins": int64(in.Wins), "losses": int64(in.Losses), "forfeits": int64(in.Forfeits), "timeouts": int64(in.Timeouts), "playtimeSeconds": in.PlaytimeSeconds} {
		if v < 0 || v > 2147483647 {
			s.writeError(w, badRequest("INVALID_STAT", name, "a whole number of 0 or more"))
			return
		}
	}
	if in.Winnings == nil {
		in.Winnings = map[string]int64{}
	}
	for k, v := range in.Winnings {
		if v < 0 || strings.TrimSpace(k) == "" {
			s.writeError(w, badRequest("INVALID_STAT", "winnings", "a currency key and a whole amount of 0 or more"))
			return
		}
	}
	win, _ := json.Marshal(in.Winnings)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_game_stats AS t (user_id, game_key, games, wins, losses, forfeits, timeouts, playtime_seconds, winnings, updated_at, updated_by)
		VALUES ($1,$2,$3+$4+$6,$3,$4,$5,$6,$7,$8::jsonb, now(), $9)
		ON CONFLICT (user_id, game_key) DO UPDATE SET games=EXCLUDED.games, wins=EXCLUDED.wins, losses=EXCLUDED.losses, forfeits=EXCLUDED.forfeits, timeouts=EXCLUDED.timeouts,
		playtime_seconds=EXCLUDED.playtime_seconds, winnings=EXCLUDED.winnings, updated_at=now(), updated_by=EXCLUDED.updated_by`,
		user, game, in.Wins, in.Losses, in.Forfeits, in.Timeouts, in.PlaytimeSeconds, string(win), pluginkit.ConsoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	items, _ := s.statsOf(r.Context(), user, game)
	writeJSON(w, http.StatusOK, map[string]any{"userId": user, "items": items})
}
