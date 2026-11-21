package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	gameCacheTTL     = time.Duration(0)
	gameOpenTTL      = 15 * time.Minute
	maxGameRooms     = 20
	maxGameMatchID   = 160
	gameSourceType   = "game"
	gameEntryReason  = "game_entry"
	gameRefundReason = "game_refund"
	gamePrizeReason  = "game_prize"
)

var (
	gameGrantNS     = uuid.Must(uuid.FromString("4f2d8f0e-5b8b-4a8e-9a51-6c0f3b1d2e77"))
	gamePlatforms   = []string{"android", "ios", "web", "windows", "macos", "linux"}
	gameVersionExpr = regexp.MustCompile(`^v?[0-9]{1,6}(\.[0-9]{1,6}){0,3}([-+][0-9A-Za-z.-]{1,32})?$`)
)

type gameRecord struct {
	ID            string                    `json:"id"`
	Key           string                    `json:"key"`
	Active        bool                      `json:"isActive"`
	Name          i18n                      `json:"name"`
	Description   i18n                      `json:"description"`
	Maintenance   i18n                      `json:"maintenance"`
	IconURL       string                    `json:"iconUrl"`
	CategoryID    string                    `json:"categoryId"`
	CategoryKey   string                    `json:"categoryKey"`
	CategoryName  i18n                      `json:"categoryName"`
	Tags          []string                  `json:"tags"`
	MinPlayers    int                       `json:"minPlayers"`
	MaxPlayers    int                       `json:"maxPlayers"`
	TurnSeconds   int                       `json:"turnSeconds"`
	MatchSeconds  int                       `json:"matchSeconds"`
	BotsAllowed   bool                      `json:"botsAllowed"`
	Rooms         []gamecatalog.Room        `json:"rooms"`
	LeaderboardID string                    `json:"leaderboardId"`
	Platforms     []string                  `json:"platforms"`
	MinVersion    string                    `json:"minVersion"`
	MaxVersion    string                    `json:"maxVersion"`
	Settings      map[string]any            `json:"settings"`
	Sort          int                       `json:"sort"`
	Voice         bool                      `json:"voiceEnabled"`
	Requirements  *gamecatalog.Requirements `json:"requirements"`
	OneMatch      bool                      `json:"oneMatchPerPlayer"`
}

const gameRecordJSON = `jsonb_build_object('id',t.id,'key',t.key,'isActive',t.is_active,'name',t.name,'description',t.description,'maintenance',t.maintenance,'iconUrl',t.icon_url,
	'categoryId',COALESCE(t.category_id::text,''),'categoryKey',COALESCE((SELECT c.key FROM base_category c WHERE c.id=t.category_id),''),
	'categoryName',COALESCE((SELECT c.name FROM base_category c WHERE c.id=t.category_id),'{}'::jsonb),
	'tags',(SELECT COALESCE(jsonb_agg(g.key ORDER BY g.key),'[]'::jsonb) FROM base_tag g WHERE g.id::text = ANY(t.tag_ids)),
	'minPlayers',t.min_players,'maxPlayers',t.max_players,'turnSeconds',COALESCE(t.turn_seconds,0),'matchSeconds',COALESCE(t.match_seconds,0),'botsAllowed',t.bots_allowed,
	'rooms',t.rooms,'leaderboardId',COALESCE(t.leaderboard_id,''),'platforms',to_jsonb(t.platforms),'minVersion',COALESCE(t.min_version,''),'maxVersion',COALESCE(t.max_version,''),
	'settings',t.settings,'sort',t.sort,'voiceEnabled',t.voice_enabled,'requirements',t.requirements,'oneMatchPerPlayer',t.one_match)`

type gameCache struct {
	mu      sync.Mutex
	version int64
	checked time.Time
	byKey   map[string]*gameRecord
	order   []*gameRecord
	ttl     time.Duration
}

func (s *service) gameRecords(ctx context.Context) ([]*gameRecord, map[string]*gameRecord, error) {
	c := &s.games
	c.mu.Lock()
	defer c.mu.Unlock()
	ttl := c.ttl
	if ttl == 0 {
		ttl = gameCacheTTL
	}
	now := s.now()
	if c.byKey != nil && now.Sub(c.checked) < ttl && !now.Before(c.checked) {
		return c.order, c.byKey, nil
	}
	var version int64
	if err := s.db.QueryRowContext(ctx, `SELECT version FROM economy_game_version WHERE id=1`).Scan(&version); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}
	if c.byKey != nil && version == c.version {
		c.checked = now
		return c.order, c.byKey, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+gameRecordJSON+`::text FROM economy_game t ORDER BY t.sort, t.key`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	byKey := map[string]*gameRecord{}
	order := []*gameRecord{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, nil, err
		}
		g := &gameRecord{}
		if err := json.Unmarshal([]byte(raw), g); err != nil {
			return nil, nil, err
		}
		if g.Settings == nil {
			g.Settings = map[string]any{}
		}
		if g.Rooms == nil {
			g.Rooms = []gamecatalog.Room{}
		}
		if g.Requirements.Empty() {
			g.Requirements = nil
		}
		byKey[g.Key] = g
		order = append(order, g)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	c.byKey, c.order, c.version, c.checked = byKey, order, version, now
	return order, byKey, nil
}

func (s *service) gameByKey(ctx context.Context, key string) (*gameRecord, error) {
	_, byKey, err := s.gameRecords(ctx)
	if err != nil {
		return nil, err
	}
	g, ok := byKey[strings.TrimSpace(key)]
	if !ok {
		return nil, notFound("GAME_NOT_FOUND")
	}
	return g, nil
}

func (g *gameRecord) public() *gamecatalog.Game {
	settings := make(map[string]any, len(g.Settings))
	for k, v := range g.Settings {
		settings[k] = v
	}
	rooms := make([]gamecatalog.Room, len(g.Rooms))
	copy(rooms, g.Rooms)
	return &gamecatalog.Game{ID: g.ID, Key: g.Key, Active: g.Active, Name: map[string]string(g.Name), MinPlayers: g.MinPlayers, MaxPlayers: g.MaxPlayers,
		TurnSeconds: g.TurnSeconds, MatchSeconds: g.MatchSeconds, BotsAllowed: g.BotsAllowed, LeaderboardID: g.LeaderboardID, OneMatch: g.OneMatch, Rooms: rooms, Settings: settings, Requirements: g.Requirements}
}

func (g *gameRecord) room(key string) (*gamecatalog.Room, error) {
	key = strings.TrimSpace(key)
	if len(g.Rooms) == 0 {
		if key == "" {
			return &gamecatalog.Room{}, nil
		}
		return nil, notFound("ROOM_NOT_FOUND")
	}
	if key == "" {
		if len(g.Rooms) == 1 {
			return &g.Rooms[0], nil
		}
		return nil, badRequest("ROOM_REQUIRED", "room", "pick one of this game's rooms")
	}
	for i := range g.Rooms {
		if g.Rooms[i].Key == key {
			return &g.Rooms[i], nil
		}
	}
	return nil, notFound("ROOM_NOT_FOUND")
}

func (g *gameRecord) liveRoom(key string) (*gamecatalog.Room, error) {
	r, err := g.room(key)
	if err != nil {
		return nil, err
	}
	if r.Disabled {
		return nil, conflict("ROOM_DISABLED", "this room is switched off for now")
	}
	return r, nil
}

func versionParts(v string) ([]int64, string, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" || !gameVersionExpr.MatchString(v) {
		return nil, "", false
	}
	pre := ""
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		if v[i] == '-' {
			pre = v[i+1:]
		}
		v = v[:i]
	}
	var nums []int64
	for _, p := range strings.Split(v, ".") {
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, "", false
		}
		nums = append(nums, n)
	}
	return nums, pre, true
}

func compareGameVersions(a, b string) (int, bool) {
	na, pa, ok1 := versionParts(a)
	nb, pb, ok2 := versionParts(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	for i := 0; i < 4; i++ {
		var x, y int64
		if i < len(na) {
			x = na[i]
		}
		if i < len(nb) {
			y = nb[i]
		}
		if x != y {
			if x < y {
				return -1, true
			}
			return 1, true
		}
	}
	switch {
	case pa == pb:
		return 0, true
	case pa == "":
		return 1, true
	case pb == "":
		return -1, true
	case pa < pb:
		return -1, true
	}
	return 1, true
}

func (g *gameRecord) availability(platform, version string) string {
	platform = strings.ToLower(strings.TrimSpace(platform))
	if len(g.Platforms) > 0 && platform != "" {
		found := false
		for _, p := range g.Platforms {
			found = found || p == platform
		}
		if !found {
			return "PLATFORM_NOT_SUPPORTED"
		}
	}
	if strings.TrimSpace(version) == "" {
		return ""
	}
	if g.MinVersion != "" {
		if c, ok := compareGameVersions(version, g.MinVersion); ok && c < 0 {
			return "UPDATE_REQUIRED"
		}
	}
	if g.MaxVersion != "" {
		if c, ok := compareGameVersions(version, g.MaxVersion); ok && c > 0 {
			return "VERSION_NOT_SUPPORTED"
		}
	}
	return ""
}

func gameUnavailable(code string, g *gameRecord) *appError {
	e := fail(http.StatusForbidden, code, "", "this game is not available on this platform or version")
	e.Meta = map[string]any{"minVersion": g.MinVersion, "maxVersion": g.MaxVersion, "platforms": g.Platforms}
	return e
}

func (s *service) gameLangs(ctx context.Context, userID, lang string) (*localized.Set, string, error) {
	return localized.ForUser(ctx, s.db, userID, lang)
}

func (s *service) gameDisabled(ctx context.Context, g *gameRecord, userID, lang string) *appError {
	e := conflict("GAME_DISABLED", "this game is switched off for now")
	msg := ""
	if langs, l, err := s.gameLangs(ctx, userID, lang); err == nil {
		msg = langs.Resolve(g.Maintenance, l)
	}
	e.Meta = map[string]any{"message": msg}
	return e
}

type gameTicket struct {
	Game              string             `json:"game"`
	Room              string             `json:"room"`
	Query             string             `json:"query"`
	MinCount          int                `json:"minCount"`
	MaxCount          int                `json:"maxCount"`
	StringProperties  map[string]string  `json:"stringProperties"`
	NumericProperties map[string]float64 `json:"numericProperties"`
	EntryFee          *gamecatalog.Fee   `json:"entryFee"`
	Realtime          bool               `json:"realtime"`
	Level             int                `json:"level"`
	Path              string             `json:"path"`
	Consumes          bool               `json:"consumes"`
}

func ticketQuery(game, room string) string {
	q := "+properties.game:" + game
	if room != "" {
		q += " +properties.room:" + room
	}
	return q
}

func (s *service) playerLevel(ctx context.Context, userID string) (int, error) {
	return newLevelCache(s.db, userID).level(ctx, "")
}

func (s *service) gameJoinCheck(ctx context.Context, userID, key, roomKey, platform, version, lang, path string) (*gameTicket, error) {
	path, err := cleanPath(path)
	if err != nil {
		return nil, err
	}
	g, err := s.gameByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if code := g.availability(platform, version); code != "" {
		return nil, gameUnavailable(code, g)
	}
	if !g.Active {
		return nil, s.gameDisabled(ctx, g, userID, lang)
	}
	if g.MaxPlayers < 2 {
		return nil, badRequest("SINGLE_PLAYER_GAME", "game", "this game is played alone; it has no matchmaking")
	}
	room, err := g.liveRoom(roomKey)
	if err != nil {
		return nil, err
	}
	if err := s.newPlayerGate(ctx, userID, g, room, path); err != nil {
		return nil, err
	}
	if g.OneMatch {
		if err := s.oneMatchGate(ctx, userID, ""); err != nil {
			return nil, err
		}
	}
	level, err := s.playerLevel(ctx, userID)
	if err != nil {
		return nil, err
	}
	if room.MinLevel > 0 && level < room.MinLevel {
		e := fail(http.StatusForbidden, "LEVEL_TOO_LOW", "room", "this room needs a higher level")
		e.Meta = map[string]any{"minLevel": room.MinLevel, "level": level}
		return nil, e
	}
	rc, err := s.checkRoomRequirements(ctx, userID, g, room, path, level)
	if err != nil {
		return nil, err
	}
	if len(rc.Missing) > 0 {
		return nil, s.requirementsError(ctx, userID, lang, g, room, path, rc)
	}
	if room.EntryFee != nil && room.EntryFee.Amount > 0 && (!room.Overdraft || room.ChargesAtStart()) {
		if err := s.ensureDefaults(ctx, userID); err != nil {
			return nil, err
		}
		wallet, err := s.userWallet(ctx, userID)
		if err != nil {
			return nil, err
		}
		if wallet[room.EntryFee.CurrencyKey] < room.EntryFee.Amount {
			e := fail(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "room", "not enough "+room.EntryFee.CurrencyKey+" for the entry fee")
			e.Meta = map[string]any{"currencyKey": room.EntryFee.CurrencyKey, "amount": room.EntryFee.Amount, "balance": wallet[room.EntryFee.CurrencyKey]}
			return nil, e
		}
	}
	def, _ := gamecatalog.Lookup(g.Key)
	t := &gameTicket{Game: g.Key, Room: room.Key, Query: ticketQuery(g.Key, room.Key), MinCount: g.MinPlayers, MaxCount: room.Seats(g.public()),
		StringProperties: map[string]string{"game": g.Key}, NumericProperties: map[string]float64{"level": float64(level)}, EntryFee: room.EntryFee, Realtime: def.Realtime, Level: level, Path: path}
	for _, r := range rc.Sets {
		t.Consumes = t.Consumes || r.Consumes()
	}
	if t.MinCount < 2 {
		t.MinCount = 2
	}
	if room.Key != "" {
		t.StringProperties["room"] = room.Key
	}
	return t, nil
}

func gameGrantID(parts ...string) string {
	return uuid.NewV5(gameGrantNS, strings.Join(parts, ":")).String()
}

func cleanMatchID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > maxGameMatchID {
		return "", badRequest("INVALID_MATCH_ID", "matchId", "a match id is required")
	}
	return id, nil
}

type gameFee struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

func (s *service) refreshPlayers(ctx context.Context, tx *sql.Tx, matchID string) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx, `UPDATE economy_game_match SET players=(SELECT count(*) FROM economy_game_entry WHERE match_id=$1 AND status IN ('paid','free')) WHERE id=$1 RETURNING players`, matchID).Scan(&n)
	return n, err
}

func (s *service) gameEnter(ctx context.Context, e gamecatalog.Entry) (gamecatalog.EntryResult, error) {
	var res gamecatalog.EntryResult
	matchID, err := cleanMatchID(e.MatchID)
	if err != nil {
		return res, err
	}
	if _, err := uuid.FromString(e.UserID); err != nil {
		return res, badRequest("INVALID_ID", "userId", "must be a UUID")
	}
	path, err := cleanPath(e.Path)
	if err != nil {
		return res, err
	}
	g, err := s.gameByKey(ctx, e.Game)
	if err != nil {
		return res, err
	}
	room, err := g.room(e.Room)
	if err != nil {
		return res, err
	}
	bot := e.Bot && s.isBot(ctx, e.UserID)
	if e.Bot && !bot {
		return res, badRequest("NOT_A_BOT", "userId", "this account is not a bot")
	}
	sets := reqSets(g.Requirements, room, path)
	if bot {
		sets = nil
	}
	if !bot && g.OneMatch {
		if err := s.oneMatchGate(ctx, e.UserID, matchID); err != nil {
			return res, err
		}
	}
	if !bot && (room.MinLevel > 0 || len(sets) > 0) {
		level, err := s.playerLevel(ctx, e.UserID)
		if err != nil {
			return res, err
		}
		if room.MinLevel > 0 && level < room.MinLevel {
			app := fail(http.StatusForbidden, "LEVEL_TOO_LOW", "room", "this room needs a higher level")
			app.Meta = map[string]any{"minLevel": room.MinLevel, "level": level}
			return res, app
		}
		if len(sets) > 0 {
			var rejoin bool
			_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_game_entry WHERE match_id=$1 AND user_id=$2 AND status IN ('paid','free'))`, matchID, e.UserID).Scan(&rejoin)
			if !rejoin {
				rc, err := s.checkRoomRequirements(ctx, e.UserID, g, room, path, level)
				if err != nil {
					return res, err
				}
				if len(rc.Missing) > 0 {
					return res, s.requirementsError(ctx, e.UserID, "", g, room, path, rc)
				}
			}
		}
	}
	var tickets *gameTickets
	charge := room.EntryFee != nil && room.EntryFee.Amount > 0 && (!bot || room.BotsPay)
	if charge && !bot {
		if err := s.ensureDefaults(ctx, e.UserID); err != nil {
			return res, err
		}
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_game_match (id, game_key, room_key) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, matchID, g.Key, room.Key); err != nil {
			return err
		}
		var gameKey, roomKey, status string
		if err := tx.QueryRowContext(ctx, `SELECT game_key, room_key, status FROM economy_game_match WHERE id=$1 FOR UPDATE`, matchID).Scan(&gameKey, &roomKey, &status); err != nil {
			return err
		}
		if gameKey != g.Key || roomKey != room.Key {
			return conflict("MATCH_MISMATCH", "this match belongs to another game or room")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_game_entry (match_id, user_id, bot) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, matchID, e.UserID, bot); err != nil {
			return err
		}
		var entryStatus string
		var attempt int
		if err := tx.QueryRowContext(ctx, `SELECT status, attempt FROM economy_game_entry WHERE match_id=$1 AND user_id=$2 FOR UPDATE`, matchID, e.UserID).Scan(&entryStatus, &attempt); err != nil {
			return err
		}
		if entryStatus == "paid" || entryStatus == "free" {
			res.Status, res.Rejoined = entryStatus, true
			n, err := s.refreshPlayers(ctx, tx, matchID)
			res.Players = n
			return err
		}
		switch status {
		case "started", "finished":
			return conflict("MATCH_STARTED", "this match has already started")
		case "cancelled":
			return conflict("MATCH_CANCELLED", "this match was cancelled")
		}
		if !g.Active {
			return s.gameDisabled(ctx, g, e.UserID, "")
		}
		if room.Disabled {
			return conflict("ROOM_DISABLED", "this room is switched off for now")
		}
		if entryStatus != "pending" {
			attempt++
		}
		next := "free"
		var feeJSON any
		if charge && room.ChargesAtStart() {
			if !bot {
				var have int64
				if err := tx.QueryRowContext(ctx, `SELECT COALESCE((wallet->>$2)::bigint, 0) FROM users WHERE id=$1`, e.UserID, room.EntryFee.CurrencyKey).Scan(&have); err != nil {
					return err
				}
				if have < room.EntryFee.Amount {
					return fail(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "room", "not enough "+room.EntryFee.CurrencyKey+" for the entry fee")
				}
			}
		} else if charge {
			fee := gameFee{CurrencyKey: room.EntryFee.CurrencyKey, Amount: room.EntryFee.Amount}
			change := s.walletChange
			if room.Overdraft && !bot {
				change = s.walletChangeOverdraft
			}
			if err := change(ctx, tx, e.UserID, gameGrantID(matchID, e.UserID, "fee", strconv.Itoa(attempt)), map[string]int64{fee.CurrencyKey: -fee.Amount}, map[string]any{
				"reason": gameEntryReason, "sourceType": gameSourceType, "sourceId": matchID, "game": g.Key, "room": room.Key, "part": "debit", "bot": bot,
			}); err != nil {
				return err
			}
			raw, _ := json.Marshal(fee)
			feeJSON = string(raw)
			next = "paid"
			res.Charged = true
			res.Fee = &gamecatalog.Fee{CurrencyKey: fee.CurrencyKey, Amount: fee.Amount}
		}
		var ticketsJSON any
		if bot {
			if _, err := tx.ExecContext(ctx, `UPDATE economy_bot SET games=games+1, last_used_at=now() WHERE user_id=$1`, e.UserID); err != nil {
				return err
			}
		}
		for _, r := range sets {
			if r.Consumes() {
				t, err := consumeTickets(ctx, tx, e.UserID, sets)
				if err != nil {
					return err
				}
				if !t.empty() {
					tickets = t
					raw, _ := json.Marshal(t)
					ticketsJSON = string(raw)
					res.Tickets = t.public()
				}
				break
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET status=$3, fee=$4::jsonb, attempt=$5, path=$6, tickets=$7::jsonb, error=NULL, updated_at=now() WHERE match_id=$1 AND user_id=$2`,
			matchID, e.UserID, next, feeJSON, attempt, path, ticketsJSON); err != nil {
			return err
		}
		res.Status = next
		n, err := s.refreshPlayers(ctx, tx, matchID)
		res.Players = n
		return err
	})
	if err != nil {
		if app, ok := asApp(err); ok && app.Code == "INSUFFICIENT_BALANCE" && room.EntryFee != nil {
			wallet, _ := s.userWallet(ctx, e.UserID)
			app.Meta = map[string]any{"currencyKey": room.EntryFee.CurrencyKey, "amount": room.EntryFee.Amount, "balance": wallet[room.EntryFee.CurrencyKey]}
		}
		if app, ok := asApp(err); ok && app.Code == codeRequirementsNotMet && app.Meta == nil {
			if level, lerr := s.playerLevel(ctx, e.UserID); lerr == nil {
				if rc, cerr := s.checkRoomRequirements(ctx, e.UserID, g, room, path, level); cerr == nil && len(rc.Missing) > 0 {
					return res, s.requirementsError(ctx, e.UserID, "", g, room, path, rc)
				}
			}
		}
		return res, err
	}
	s.logTickets(ctx, e.UserID, matchID, g.Key, room.Key, "game_ticket", tickets)
	if res.Charged && !bot {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: e.UserID, Kind: activity.KindPrize, Action: "game_entry", RefID: matchID, Actor: "system",
			Data: map[string]any{"game": g.Key, "room": room.Key, "currency": res.Fee.CurrencyKey, "amount": res.Fee.Amount}})
	}
	return res, nil
}

func (s *service) refundEntry(ctx context.Context, tx *sql.Tx, matchID, userID string, status string, attempt int, feeRaw sql.NullString, gameKey string) (bool, error) {
	if status != "paid" {
		return false, nil
	}
	var fee gameFee
	if feeRaw.Valid {
		_ = json.Unmarshal([]byte(feeRaw.String), &fee)
	}
	if fee.CurrencyKey == "" || fee.Amount <= 0 {
		return false, nil
	}
	if err := s.walletChange(ctx, tx, userID, gameGrantID(matchID, userID, "refund", strconv.Itoa(attempt)), map[string]int64{fee.CurrencyKey: fee.Amount}, map[string]any{
		"reason": gameRefundReason, "sourceType": gameSourceType, "sourceId": matchID, "game": gameKey, "part": "refund",
	}); err != nil {
		return false, err
	}
	_, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET status='refunded', updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, userID)
	return err == nil, err
}

func (s *service) gameLeave(ctx context.Context, matchID, userID string) (bool, error) {
	matchID, err := cleanMatchID(matchID)
	if err != nil {
		return false, err
	}
	refunded := false
	var returned *gameTickets
	var leftGame string
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var status, gameKey string
		err := tx.QueryRowContext(ctx, `SELECT status, game_key FROM economy_game_match WHERE id=$1 FOR UPDATE`, matchID).Scan(&status, &gameKey)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("MATCH_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		if status != "open" {
			return nil
		}
		leftGame = gameKey
		var entryStatus string
		var attempt int
		var fee, ticketsRaw sql.NullString
		err = tx.QueryRowContext(ctx, `SELECT status, attempt, fee::text, tickets::text FROM economy_game_entry WHERE match_id=$1 AND user_id=$2 FOR UPDATE`, matchID, userID).Scan(&entryStatus, &attempt, &fee, &ticketsRaw)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if entryStatus == "paid" || entryStatus == "free" {
			t, err := returnTickets(ctx, tx, matchID, userID, ticketsRaw)
			if err != nil {
				return err
			}
			if t != nil {
				returned = t
			}
		}
		switch entryStatus {
		case "paid":
			ok, err := s.refundEntry(ctx, tx, matchID, userID, entryStatus, attempt, fee, gameKey)
			if err != nil {
				return err
			}
			refunded = ok
		case "free", "pending":
			if _, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET status='left', updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, userID); err != nil {
				return err
			}
		}
		_, err = s.refreshPlayers(ctx, tx, matchID)
		return err
	})
	if err == nil && returned != nil {
		s.logTickets(ctx, userID, matchID, leftGame, "", "game_ticket_returned", returned)
	}
	return refunded, err
}

func (s *service) gameStart(ctx context.Context, matchID string) error {
	matchID, err := cleanMatchID(matchID)
	if err != nil {
		return err
	}
	var gameKey, roomKey, current string
	err = s.db.QueryRowContext(ctx, `SELECT game_key, room_key, status FROM economy_game_match WHERE id=$1`, matchID).Scan(&gameKey, &roomKey, &current)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound("MATCH_NOT_FOUND")
	}
	if err != nil {
		return err
	}
	if current == "open" {
		if g, gerr := s.gameByKey(ctx, gameKey); gerr == nil {
			if room, rerr := g.room(roomKey); rerr == nil && room.ChargesAtStart() && room.EntryFee != nil && room.EntryFee.Amount > 0 {
				return s.startCharging(ctx, matchID, g, room)
			}
		}
	}
	var status string
	err = s.db.QueryRowContext(ctx, `UPDATE economy_game_match SET status=CASE WHEN status='open' THEN 'started' ELSE status END,
		started_at=CASE WHEN status='open' THEN now() ELSE started_at END WHERE id=$1 RETURNING status`, matchID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound("MATCH_NOT_FOUND")
	}
	if err != nil {
		return err
	}
	if status == "cancelled" {
		return conflict("MATCH_CANCELLED", "this match was cancelled")
	}
	return nil
}

func (s *service) startCharging(ctx context.Context, matchID string, g *gameRecord, room *gamecatalog.Room) error {
	type payer struct {
		user    string
		bot     bool
		attempt int
	}
	var humans []string
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text FROM economy_game_entry WHERE match_id=$1 AND status='free' AND NOT bot`, matchID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return err
		}
		humans = append(humans, u)
	}
	rows.Close()
	for _, u := range humans {
		if err := s.ensureDefaults(ctx, u); err != nil {
			return err
		}
	}
	fee := gameFee{CurrencyKey: room.EntryFee.CurrencyKey, Amount: room.EntryFee.Amount}
	raw, _ := json.Marshal(fee)
	var charged []string
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		charged = nil
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM economy_game_match WHERE id=$1 FOR UPDATE`, matchID).Scan(&status); err != nil {
			return err
		}
		switch status {
		case "cancelled":
			return conflict("MATCH_CANCELLED", "this match was cancelled")
		case "started", "finished":
			return nil
		}
		rows, err := tx.QueryContext(ctx, `SELECT user_id::text, bot, attempt FROM economy_game_entry WHERE match_id=$1 AND status='free' ORDER BY user_id FOR UPDATE`, matchID)
		if err != nil {
			return err
		}
		var payers []payer
		for rows.Next() {
			var p payer
			if err := rows.Scan(&p.user, &p.bot, &p.attempt); err != nil {
				rows.Close()
				return err
			}
			payers = append(payers, p)
		}
		rows.Close()
		for _, p := range payers {
			if p.bot && !room.BotsPay {
				continue
			}
			change := s.walletChange
			if room.Overdraft {
				change = s.walletChangeOverdraft
			}
			if err := change(ctx, tx, p.user, gameGrantID(matchID, p.user, "fee", strconv.Itoa(p.attempt)), map[string]int64{fee.CurrencyKey: -fee.Amount}, map[string]any{
				"reason": gameEntryReason, "sourceType": gameSourceType, "sourceId": matchID, "game": g.Key, "room": room.Key, "part": "debit", "bot": p.bot, "at": gamecatalog.FeeAtStart,
			}); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET status='paid', fee=$3::jsonb, updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, p.user, string(raw)); err != nil {
				return err
			}
			if !p.bot {
				charged = append(charged, p.user)
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_game_match SET status='started', started_at=now() WHERE id=$1`, matchID)
		return err
	})
	if err != nil {
		return err
	}
	for _, u := range charged {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: u, Kind: activity.KindPrize, Action: "game_entry", RefID: matchID, Actor: "system",
			Data: map[string]any{"game": g.Key, "room": room.Key, "currency": fee.CurrencyKey, "amount": fee.Amount}})
	}
	return nil
}

func (s *service) gameCancel(ctx context.Context, matchID string) (int, error) {
	return s.cancelMatch(ctx, matchID, false, "", "")
}

func (s *service) cancelMatch(ctx context.Context, matchID string, force bool, reason, by string) (int, error) {
	matchID, err := cleanMatchID(matchID)
	if err != nil {
		return 0, err
	}
	refunded := 0
	returnedTo := map[string]*gameTickets{}
	cancelledGame := ""
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var status, gameKey string
		err := tx.QueryRowContext(ctx, `SELECT status, game_key FROM economy_game_match WHERE id=$1 FOR UPDATE`, matchID).Scan(&status, &gameKey)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("MATCH_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		cancelledGame = gameKey
		switch {
		case status == "cancelled":
			return nil
		case status == "finished" && force:
			return conflict("MATCH_FINISHED", "this match has already finished")
		case (status == "started" && !force) || status == "finished":
			return conflict("MATCH_STARTED", "a match that has started cannot be cancelled")
		}
		type row struct {
			user, status string
			attempt      int
			fee, tickets sql.NullString
		}
		rows, err := tx.QueryContext(ctx, `SELECT user_id::text, status, attempt, fee::text, tickets::text FROM economy_game_entry WHERE match_id=$1 ORDER BY user_id FOR UPDATE`, matchID)
		if err != nil {
			return err
		}
		var entries []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.user, &r.status, &r.attempt, &r.fee, &r.tickets); err != nil {
				rows.Close()
				return err
			}
			entries = append(entries, r)
		}
		rows.Close()
		for _, r := range entries {
			if r.status == "paid" || r.status == "free" {
				t, err := returnTickets(ctx, tx, matchID, r.user, r.tickets)
				if err != nil {
					return err
				}
				if t != nil {
					returnedTo[r.user] = t
				}
			}
			ok, err := s.refundEntry(ctx, tx, matchID, r.user, r.status, r.attempt, r.fee, gameKey)
			if err != nil {
				return err
			}
			if ok {
				refunded++
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_game_match SET status='cancelled', ended_at=now(), end_reason=COALESCE(NULLIF($2,''), end_reason, 'cancelled'), forced_by=NULLIF($3,'') WHERE id=$1`, matchID, reason, by); err != nil {
			return err
		}
		_, err = s.refreshPlayers(ctx, tx, matchID)
		return err
	})
	if err == nil {
		for user, t := range returnedTo {
			s.logTickets(ctx, user, matchID, cancelledGame, "", "game_ticket_returned", t)
		}
	}
	return refunded, err
}

func (s *service) gameFinish(ctx context.Context, matchID string, winners []string) ([]gamecatalog.Payout, error) {
	matchID, err := cleanMatchID(matchID)
	if err != nil {
		return nil, err
	}
	clean := []string{}
	seen := map[string]bool{}
	for _, w := range winners {
		if _, err := uuid.FromString(w); err == nil && !seen[w] {
			seen[w] = true
			clean = append(clean, w)
		}
	}
	var gameKey, roomKey string
	first := false
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var status, storedRaw string
		err := tx.QueryRowContext(ctx, `SELECT status, game_key, room_key, to_jsonb(winners)::text FROM economy_game_match WHERE id=$1 FOR UPDATE`, matchID).Scan(&status, &gameKey, &roomKey, &storedRaw)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("MATCH_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		if status == "cancelled" {
			return conflict("MATCH_CANCELLED", "this match was cancelled")
		}
		if status == "finished" {
			stored := []string{}
			_ = json.Unmarshal([]byte(storedRaw), &stored)
			clean = stored
			return nil
		}
		first = true
		if _, err := tx.ExecContext(ctx, `UPDATE economy_game_match SET status='finished', winners=ARRAY(SELECT jsonb_array_elements_text($2::jsonb)), started_at=COALESCE(started_at, now()), ended_at=now() WHERE id=$1`, matchID, stringsJSON(clean)); err != nil {
			return err
		}
		g, gerr := s.gameByKey(ctx, gameKey)
		var room *gamecatalog.Room
		if gerr == nil {
			room, _ = g.room(roomKey)
		}
		for _, w := range clean {
			var entryStatus, prizeStatus string
			err := tx.QueryRowContext(ctx, `SELECT status, prize_status FROM economy_game_entry WHERE match_id=$1 AND user_id=$2 FOR UPDATE`, matchID, w).Scan(&entryStatus, &prizeStatus)
			if errors.Is(err, sql.ErrNoRows) || (err == nil && entryStatus != "paid" && entryStatus != "free") {
				continue
			}
			if err != nil {
				return err
			}
			if prizeStatus != "none" || room == nil {
				continue
			}
			prizeID := room.PrizeID
			if prizeID == "" && room.PrizeGroupID != "" {
				prizeID, err = s.drawFromPrizeGroup(ctx, tx, room.PrizeGroupID)
				if err != nil {
					return err
				}
			}
			if prizeID == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET prize_id=$3, prize_status='pending', updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, w, prizeID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	payouts := make([]gamecatalog.Payout, 0, len(clean))
	for _, w := range clean {
		p := gamecatalog.Payout{UserID: w, Rewards: []map[string]any{}}
		var prizeID sql.NullString
		var prizeStatus string
		err := s.db.QueryRowContext(ctx, `SELECT prize_id::text, prize_status FROM economy_game_entry WHERE match_id=$1 AND user_id=$2`, matchID, w).Scan(&prizeID, &prizeStatus)
		if errors.Is(err, sql.ErrNoRows) {
			p.Error = "NOT_IN_MATCH"
			payouts = append(payouts, p)
			continue
		}
		if err != nil {
			return payouts, err
		}
		if prizeID.Valid {
			p.PrizeID = prizeID.String
			mid, uid := matchID, w
			rewards, err := s.grant(ctx, grantRequest{UserID: w, PrizeID: prizeID.String, SourceType: gameSourceType, SourceID: gameGrantID(matchID, w, "prize"), Reason: gamePrizeReason,
				After: func(ctx context.Context, tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET prize_status='paid', error=NULL, updated_at=now() WHERE match_id=$1 AND user_id=$2`, mid, uid)
					return err
				}})
			if err != nil {
				code := err.Error()
				if app, ok := asApp(err); ok {
					code = app.Code
				}
				p.Error = code
				_, _ = s.db.ExecContext(ctx, `UPDATE economy_game_entry SET error=$3, updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, w, truncate(code, 500))
			}
			for _, r := range rewards {
				raw, _ := json.Marshal(r)
				m := map[string]any{}
				_ = json.Unmarshal(raw, &m)
				p.Rewards = append(p.Rewards, m)
			}
		}
		payouts = append(payouts, p)
	}
	if first {
		s.writeGameLeaderboard(ctx, gameKey, clean)
	}
	return payouts, nil
}

func (s *service) writeGameLeaderboard(ctx context.Context, gameKey string, winners []string) {
	g, err := s.gameByKey(ctx, gameKey)
	if err != nil || g.LeaderboardID == "" {
		return
	}
	for _, w := range winners {
		var username string
		_ = s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, w).Scan(&username)
		if _, err := s.nk.LeaderboardRecordWrite(ctx, g.LeaderboardID, w, username, 1, 0, map[string]any{"game": gameKey}, nil); err != nil {
			s.logger.Warn("Economy game %s leaderboard %s write failed: %v", gameKey, g.LeaderboardID, err)
		}
	}
}

func (s *service) expireGameMatches(ctx context.Context, limit int) int {
	done := 0
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM economy_game_match WHERE status='open' AND created_at < $1 ORDER BY created_at LIMIT $2`, s.now().UTC().Add(-gameOpenTTL), limit)
	if err != nil {
		s.logger.Warn("Economy game match sweep failed: %v", err)
		return 0
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		if _, err := s.gameCancel(ctx, id); err != nil {
			if app, ok := asApp(err); !ok || app.Code != "MATCH_STARTED" {
				s.logger.Warn("Economy game match %s not cancelled: %v", id, err)
			}
			continue
		}
		done++
	}
	return done
}

func providerError(err error) error {
	if err == nil {
		return nil
	}
	if app, ok := asApp(err); ok {
		return &gamecatalog.Error{Code: app.Code, Message: app.Message, Meta: app.Meta}
	}
	return err
}

type gameProvider struct{ s *service }

func (p gameProvider) Game(ctx context.Context, key string) (*gamecatalog.Game, error) {
	g, err := p.s.gameByKey(ctx, key)
	if err != nil {
		if app, ok := asApp(err); ok && app.Code == "GAME_NOT_FOUND" {
			return nil, gamecatalog.ErrNotFound
		}
		return nil, providerError(err)
	}
	return g.public(), nil
}

func (p gameProvider) Enter(ctx context.Context, e gamecatalog.Entry) (gamecatalog.EntryResult, error) {
	r, err := p.s.gameEnter(ctx, e)
	return r, providerError(err)
}

func (p gameProvider) Leave(ctx context.Context, matchID, userID string) (bool, error) {
	r, err := p.s.gameLeave(ctx, matchID, userID)
	return r, providerError(err)
}

func (p gameProvider) Start(ctx context.Context, matchID string) error {
	return providerError(p.s.gameStart(ctx, matchID))
}

func (p gameProvider) Cancel(ctx context.Context, matchID string) (int, error) {
	n, err := p.s.gameCancel(ctx, matchID)
	return n, providerError(err)
}

func (p gameProvider) Finish(ctx context.Context, matchID string, winners []string) ([]gamecatalog.Payout, error) {
	r, err := p.s.gameFinish(ctx, matchID, winners)
	return r, providerError(err)
}

func (s *service) matchmakerMatched(ctx context.Context, logger runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
	if len(entries) == 0 {
		return "", nil
	}
	props := entries[0].GetProperties()
	key, _ := props["game"].(string)
	room, _ := props["room"].(string)
	if key == "" {
		return "", nil
	}
	def, ok := gamecatalog.Lookup(key)
	if !ok || !def.Realtime {
		return "", nil
	}
	users := make([]string, 0, len(entries))
	for _, e := range entries {
		users = append(users, e.GetPresence().GetUserId())
	}
	id, err := nk.MatchCreate(ctx, key, map[string]any{"game": key, "room": room, "users": strings.Join(users, ",")})
	if err != nil {
		logger.Error("Economy game %s match create failed: %v", key, err)
		return "", err
	}
	return id, nil
}

func (s *service) sweepGames(ctx context.Context) {
	s.sweepGameHistory(ctx)
	if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_game_entry WHERE (match_id, user_id) IN (SELECT e.match_id, e.user_id FROM economy_game_entry e
		JOIN economy_game_match m ON m.id=e.match_id WHERE m.status IN ('finished','cancelled') AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = e.user_id) LIMIT 500)`); err != nil {
		s.logger.Warn("Economy game entry sweep failed: %v", err)
	}
}
