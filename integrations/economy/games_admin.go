package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const gameAdminJSON = `jsonb_build_object('id',t.id,'key',t.key,'isActive',t.is_active,'name',t.name,'description',t.description,'maintenance',t.maintenance,'iconUrl',t.icon_url,
	'categoryId',t.category_id,'tagIds',to_jsonb(t.tag_ids),'minPlayers',t.min_players,'maxPlayers',t.max_players,'turnSeconds',t.turn_seconds,'matchSeconds',t.match_seconds,
	'botsAllowed',t.bots_allowed,'rooms',t.rooms,'leaderboardId',t.leaderboard_id,'platforms',to_jsonb(t.platforms),'minVersion',t.min_version,'maxVersion',t.max_version,
	'settings',t.settings,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'matches',(SELECT count(*) FROM economy_game_match m WHERE m.game_key=t.key),
	'finished',(SELECT count(*) FROM economy_game_match m WHERE m.game_key=t.key AND m.status='finished'),
	'cancelled',(SELECT count(*) FROM economy_game_match m WHERE m.game_key=t.key AND m.status='cancelled'),
	'live',(SELECT count(*) FROM economy_game_match m WHERE m.game_key=t.key AND m.status IN ('open','started')),
	'players',(SELECT count(DISTINCT e.user_id) FROM economy_game_entry e JOIN economy_game_match m ON m.id=e.match_id WHERE m.game_key=t.key),
	'lastMatchAt',(SELECT max(m.created_at) FROM economy_game_match m WHERE m.game_key=t.key))`

func gameResources() []resource {
	return []resource{{
		name:   "games",
		table:  "economy_game",
		json:   gameAdminJSON,
		search: []string{"t.key", translatedSearch("t.name")},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean", "categoryId": "t.category_id::text = ?",
			"platform": "(cardinality(t.platforms) = 0 OR ? = ANY(t.platforms))"},
		save: saveGame,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_game g JOIN economy_game_match m ON m.game_key=g.key WHERE g.id=$1 AND m.status IN ('open','started'))`)
		},
	}}
}

type gameRoomInput struct {
	Key      string `json:"key"`
	Name     i18n   `json:"name"`
	EntryFee *struct {
		CurrencyKey string          `json:"currencyKey"`
		Amount      json.RawMessage `json:"amount"`
	} `json:"entryFee"`
	PrizeID      string          `json:"prizeId"`
	PrizeGroupID string          `json:"prizeGroupId"`
	MinLevel     json.RawMessage `json:"minLevel"`
}

func optionalInt(raw json.RawMessage, field, code, msg string, min, max int) (*int, error) {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == `""` {
		return nil, nil
	}
	var v float64
	if err := json.Unmarshal(raw, &v); err != nil || v != math.Trunc(v) || v < float64(min) || v > float64(max) {
		return nil, badRequest(code, field, msg)
	}
	n := int(v)
	return &n, nil
}

func cleanGameRooms(ctx context.Context, tx *sql.Tx, langs *localized.Set, in []gameRoomInput) ([]gamecatalog.Room, error) {
	if len(in) > maxGameRooms {
		return nil, badRequest("TOO_MANY_ROOMS", "rooms", "at most 20 rooms")
	}
	out := make([]gamecatalog.Room, 0, len(in))
	seen := map[string]bool{}
	for i, r := range in {
		field := "rooms[" + strconv.Itoa(i) + "]"
		key := strings.TrimSpace(r.Key)
		if !gamecatalog.ValidKey(key) {
			return nil, badRequest("INVALID_ROOM_KEY", field, "a room key is 2 to 64 lower-case letters, digits or underscores")
		}
		if seen[key] {
			return nil, badRequest("DUPLICATE_ROOM", field, "each room key can be used once")
		}
		seen[key] = true
		name, err := langs.Clean(field, r.Name, false, 64)
		if err != nil {
			return nil, textError(err)
		}
		room := gamecatalog.Room{Key: key, Name: map[string]string(name)}
		if r.EntryFee != nil && (strings.TrimSpace(r.EntryFee.CurrencyKey) != "" || (len(r.EntryFee.Amount) > 0 && string(r.EntryFee.Amount) != "null")) {
			cur := strings.TrimSpace(r.EntryFee.CurrencyKey)
			amount, err := optionalInt(r.EntryFee.Amount, field, "INVALID_ENTRY_FEE", "the entry fee is a whole amount from 1 to 1000000000", 1, 1000000000)
			if err != nil {
				return nil, err
			}
			if cur == "" || amount == nil {
				return nil, badRequest("INVALID_ENTRY_FEE", field, "give both the currency and the amount, or neither")
			}
			var ok bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, cur).Scan(&ok); err != nil {
				return nil, err
			}
			if !ok {
				return nil, badRequest("CURRENCY_NOT_FOUND", field, "currency does not exist")
			}
			room.EntryFee = &gamecatalog.Fee{CurrencyKey: cur, Amount: int64(*amount)}
		}
		prizeID, groupID := strings.TrimSpace(r.PrizeID), strings.TrimSpace(r.PrizeGroupID)
		if prizeID != "" && groupID != "" {
			return nil, badRequest("ONE_ROOM_PRIZE", field, "a room pays a prize or a prize group, not both")
		}
		if prizeID != "" {
			ref, err := refExists(ctx, tx, "economy_prize", field, "PRIZE_NOT_FOUND", prizeID, true)
			if err != nil {
				return nil, err
			}
			var repeatable bool
			if err := tx.QueryRowContext(ctx, `SELECT is_repeatable FROM economy_prize WHERE id=$1`, *ref).Scan(&repeatable); err != nil {
				return nil, err
			}
			if !repeatable {
				return nil, badRequest("PRIZE_NOT_REPEATABLE", field, "a room prize is paid on every win, so the prize must be repeatable")
			}
			room.PrizeID = *ref
		}
		if groupID != "" {
			ref, err := refExists(ctx, tx, "economy_prize_group", field, "PRIZE_GROUP_NOT_FOUND", groupID, true)
			if err != nil {
				return nil, err
			}
			room.PrizeGroupID = *ref
		}
		lvl, err := optionalInt(r.MinLevel, field, "INVALID_MIN_LEVEL", "the level is a whole number from 0 to 1000", 0, 1000)
		if err != nil {
			return nil, err
		}
		if lvl != nil {
			room.MinLevel = *lvl
		}
		out = append(out, room)
	}
	return out, nil
}

func saveGame(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key           string          `json:"key"`
		IsActive      bool            `json:"isActive"`
		Name          i18n            `json:"name"`
		Description   i18n            `json:"description"`
		Maintenance   i18n            `json:"maintenance"`
		IconURL       string          `json:"iconUrl"`
		CategoryID    string          `json:"categoryId"`
		TagIDs        []string        `json:"tagIds"`
		MinPlayers    json.RawMessage `json:"minPlayers"`
		MaxPlayers    json.RawMessage `json:"maxPlayers"`
		TurnSeconds   json.RawMessage `json:"turnSeconds"`
		MatchSeconds  json.RawMessage `json:"matchSeconds"`
		BotsAllowed   bool            `json:"botsAllowed"`
		Rooms         []gameRoomInput `json:"rooms"`
		LeaderboardID string          `json:"leaderboardId"`
		Platforms     []string        `json:"platforms"`
		MinVersion    string          `json:"minVersion"`
		MaxVersion    string          `json:"maxVersion"`
		Settings      json.RawMessage `json:"settings"`
		Sort          int             `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key := strings.TrimSpace(in.Key)
	if !gamecatalog.ValidKey(key) {
		return badRequest("INVALID_GAME_KEY", "key", "the game key is the match handler name: 2 to 64 lower-case letters, digits or underscores, starting with a letter")
	}
	if err := keepKey(ctx, tx, "economy_game", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	maint, err := cleanI18n(langs, "maintenance", in.Maintenance, false, 300)
	if err != nil {
		return err
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	cat, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, cat, "game"); err != nil {
		return err
	}
	tags, err := checkTags(ctx, tx, in.TagIDs, "game")
	if err != nil {
		return err
	}
	minP, err := optionalInt(in.MinPlayers, "minPlayers", "INVALID_PLAYERS", "from 1 to 100 players", 1, 100)
	if err != nil {
		return err
	}
	maxP, err := optionalInt(in.MaxPlayers, "maxPlayers", "INVALID_PLAYERS", "from 1 to 100 players", 1, 100)
	if err != nil {
		return err
	}
	lo, hi := 1, 1
	if minP != nil {
		lo = *minP
	}
	if maxP != nil {
		hi = *maxP
	} else {
		hi = lo
	}
	if hi < lo {
		return badRequest("MAX_BELOW_MIN", "maxPlayers", "the most players cannot be below the fewest")
	}
	turn, err := optionalInt(in.TurnSeconds, "turnSeconds", "INVALID_SECONDS", "whole seconds from 1 to 86400", 1, 86400)
	if err != nil {
		return err
	}
	limit, err := optionalInt(in.MatchSeconds, "matchSeconds", "INVALID_SECONDS", "whole seconds from 1 to 604800", 1, 604800)
	if err != nil {
		return err
	}
	rooms, err := cleanGameRooms(ctx, tx, langs, in.Rooms)
	if err != nil {
		return err
	}
	var board any
	if b := strings.TrimSpace(in.LeaderboardID); b != "" {
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM leaderboard WHERE id=$1)`, b).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("LEADERBOARD_NOT_FOUND", "leaderboardId", "no leaderboard has this id")
		}
		board = b
	}
	platforms := []string{}
	seen := map[string]bool{}
	for _, p := range in.Platforms {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" || seen[p] {
			continue
		}
		if err := oneOf("platforms", "INVALID_PLATFORM", p, gamePlatforms...); err != nil {
			return err
		}
		seen[p] = true
		platforms = append(platforms, p)
	}
	var minV, maxV any
	if v := strings.TrimSpace(in.MinVersion); v != "" {
		if _, _, ok := versionParts(v); !ok {
			return badRequest("INVALID_VERSION", "minVersion", "use a version like 1.4.0")
		}
		minV = v
	}
	if v := strings.TrimSpace(in.MaxVersion); v != "" {
		if _, _, ok := versionParts(v); !ok {
			return badRequest("INVALID_VERSION", "maxVersion", "use a version like 1.4.0")
		}
		maxV = v
	}
	if minV != nil && maxV != nil {
		if c, _ := compareGameVersions(minV.(string), maxV.(string)); c > 0 {
			return badRequest("VERSION_RANGE", "maxVersion", "the highest version is below the lowest")
		}
	}
	settingsRaw, err := cleanObject("settings", in.Settings)
	if err != nil {
		return err
	}
	if def, ok := gamecatalog.Lookup(key); ok && len(def.Schema) > 0 {
		schema, err := gamecatalog.ParseSchema(def.Schema)
		if err != nil {
			return err
		}
		values := map[string]any{}
		_ = json.Unmarshal([]byte(settingsRaw), &values)
		out, fe := schema.Apply(values)
		if fe != nil {
			return badRequest(fe.Code, fe.Field, fe.Message)
		}
		b, _ := json.Marshal(out)
		settingsRaw = string(b)
	}
	roomsJSON, _ := json.Marshal(rooms)
	return execSave(ctx, tx, "economy_game", id, create,
		[]string{"key", "is_active", "name", "description", "maintenance", "icon_url", "category_id", "tag_ids", "min_players", "max_players", "turn_seconds", "match_seconds",
			"bots_allowed", "rooms", "leaderboard_id", "platforms", "min_version", "max_version", "settings", "sort"},
		[]any{key, in.IsActive, i18nJSON(name), i18nJSON(desc), i18nJSON(maint), icon, cat, tags, lo, hi, turn, limit,
			in.BotsAllowed, string(roomsJSON), board, platforms, minV, maxV, settingsRaw, in.Sort})
}

func (s *service) gameConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/games-registry", http.MethodGet, s.consoleGameRegistry},
		{consolePrefix + "/game-matches", http.MethodGet, s.consoleGameMatches},
		{consolePrefix + "/game-matches/{id}", http.MethodGet, s.consoleGameMatch},
		{consolePrefix + "/game-matches/{id}/cancel", http.MethodPost, s.consoleGameMatchCancel},
	}
}

func (s *service) consoleGameRegistry(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT key, id::text, name::text, is_active FROM economy_game ORDER BY sort, key`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	type configured struct {
		ID     string
		Name   json.RawMessage
		Active bool
	}
	have := map[string]configured{}
	order := []string{}
	for rows.Next() {
		var key, id, name string
		var active bool
		if err := rows.Scan(&key, &id, &name, &active); err != nil {
			s.writeError(w, err)
			return
		}
		have[key] = configured{ID: id, Name: json.RawMessage(name), Active: active}
		order = append(order, key)
	}
	items := []map[string]any{}
	registered := map[string]bool{}
	for _, d := range gamecatalog.Registered() {
		registered[d.Key] = true
		item := map[string]any{"key": d.Key, "name": d.Name, "plugin": d.Plugin, "realtime": d.Realtime, "registered": true, "configured": false}
		if len(d.Schema) > 0 {
			item["schema"] = json.RawMessage(d.Schema)
		}
		if c, ok := have[d.Key]; ok {
			item["configured"], item["id"], item["isActive"] = true, c.ID, c.Active
		}
		items = append(items, item)
	}
	missing := []map[string]any{}
	for _, key := range order {
		if registered[key] {
			continue
		}
		c := have[key]
		missing = append(missing, map[string]any{"key": key, "id": c.ID, "name": c.Name, "isActive": c.Active})
	}
	unconfigured := []map[string]any{}
	for _, it := range items {
		if it["configured"] == false {
			unconfigured = append(unconfigured, it)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "unconfigured": unconfigured, "missingHandler": missing})
}

const gameMatchAdminJSON = `jsonb_build_object('id',m.id,'gameKey',m.game_key,'gameName',g.name,'roomKey',m.room_key,'status',m.status,'players',m.players,'winners',to_jsonb(m.winners),
	'createdAt',m.created_at,'startedAt',m.started_at,'endedAt',m.ended_at,
	'fees',(SELECT COALESCE(sum((e.fee->>'amount')::bigint),0) FROM economy_game_entry e WHERE e.match_id=m.id AND e.status='paid'),
	'refunds',(SELECT count(*) FROM economy_game_entry e WHERE e.match_id=m.id AND e.status='refunded'))::text`

const gameMatchAdminFrom = ` FROM economy_game_match m LEFT JOIN economy_game g ON g.key=m.game_key`

func (s *service) consoleGameMatches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	where := &sqlWhere{}
	if v := strings.TrimSpace(q.Get("gameKey")); v != "" {
		where.add("m.game_key = ?", v)
	}
	if v := strings.TrimSpace(q.Get("status")); v != "" {
		if err := oneOf("status", "INVALID_FILTER", v, "open", "started", "finished", "cancelled"); err != nil {
			s.writeError(w, err)
			return
		}
		where.add("m.status = ?", v)
	}
	if v := strings.TrimSpace(q.Get("userId")); v != "" {
		where.add("EXISTS (SELECT 1 FROM economy_game_entry e WHERE e.match_id=m.id AND e.user_id::text = ?)", v)
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		n := where.next()
		where.args = append(where.args, likePattern(v))
		where.parts = append(where.parts, "(m.id ILIKE "+n+" OR m.game_key ILIKE "+n+")")
	}
	s.listRaw(w, r, `SELECT count(*)`+gameMatchAdminFrom, `SELECT `+gameMatchAdminJSON+gameMatchAdminFrom, where, ` ORDER BY m.created_at DESC, m.id DESC`, readPage(r))
}

func (s *service) consoleGameMatch(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var raw string
	err := s.db.QueryRowContext(r.Context(), `SELECT `+gameMatchAdminJSON+gameMatchAdminFrom+` WHERE m.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("MATCH_NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &out)
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('userId',e.user_id,'username',u.username,'status',e.status,'fee',e.fee,'prizeId',e.prize_id,'prizeStatus',e.prize_status,
		'error',e.error,'createdAt',e.created_at,'updatedAt',e.updated_at)::text FROM economy_game_entry e LEFT JOIN users u ON u.id=e.user_id WHERE e.match_id=$1 ORDER BY e.created_at, e.user_id`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	entries := []json.RawMessage{}
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			s.writeError(w, err)
			return
		}
		entries = append(entries, json.RawMessage(e))
	}
	out["entries"] = entries
	writeJSON(w, http.StatusOK, out)
}

func (s *service) consoleGameMatchCancel(w http.ResponseWriter, r *http.Request) {
	n, err := s.gameCancel(r.Context(), pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"refunded": n})
}
