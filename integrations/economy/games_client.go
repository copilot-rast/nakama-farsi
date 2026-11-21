package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) gameRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_games":     s.rpcGames,
		"economy_game":      s.rpcGame,
		"economy_game_join": s.rpcGameJoin,
	}
}

type gameRequest struct {
	Key      string `json:"key"`
	Room     string `json:"room"`
	Lang     string `json:"lang"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
	Path     string `json:"path"`
}

type gameLookups struct {
	langs      *localized.Set
	lang       string
	prizes     map[string]map[string]any
	groups     map[string]map[string]any
	currencies map[string]i18n
	level      int
	userID     string
	path       string
	facts      *reqFacts
	names      *reqNames
}

// reqFor loads the player's facts and the names once for every game in the answer.
func (s *service) reqFor(ctx context.Context, lk *gameLookups, games []*gameRecord) error {
	if lk.facts != nil {
		return nil
	}
	found, items, metaIDs := reqNeeds(games)
	lk.names = newReqNames(s, lk.langs, lk.lang)
	if !found {
		lk.facts = &reqFacts{Level: lk.level, Items: map[string]bool{}, Tags: map[string]bool{}, Metas: map[string]float64{}}
		lk.names.loaded = true
		return nil
	}
	f, err := s.loadReqFacts(ctx, s.db, lk.userID, lk.level, items, metaIDs)
	if err != nil {
		return err
	}
	lk.facts = &f
	return lk.names.load(ctx)
}

type roomEligibility struct {
	sets     []*gamecatalog.Requirements
	missing  []reqMiss
	eligible bool
	by       map[string]bool
	hidden   bool
}

func eligibilityFor(lk *gameLookups, game *gamecatalog.Requirements, room *gamecatalog.Room) roomEligibility {
	out := roomEligibility{by: map[string]bool{}}
	for _, p := range gamecatalog.Paths {
		sets := reqSets(game, room, p)
		miss := []reqMiss{}
		for _, r := range sets {
			miss = append(miss, reqMissing(r, *lk.facts)...)
		}
		out.by[p] = len(miss) == 0
		if p == lk.path {
			out.sets, out.missing, out.eligible = sets, miss, len(miss) == 0
			out.hidden = !out.eligible && reqHidden(sets)
		}
	}
	return out
}

func (s *service) eligibilityView(ctx context.Context, lk *gameLookups, el roomEligibility, view map[string]any) error {
	missing := []map[string]any{}
	for _, m := range el.missing {
		missing = append(missing, lk.names.missView(m))
	}
	view["eligible"], view["eligibleBy"], view["missing"], view["products"] = el.eligible, el.by, missing, []map[string]any{}
	if !el.eligible {
		products, err := lk.names.products(ctx, el.sets, el.missing)
		if err != nil {
			return err
		}
		view["products"] = products
	}
	consumes := false
	for _, r := range el.sets {
		consumes = consumes || r.Consumes()
	}
	view["consumes"] = consumes
	return nil
}

func (s *service) gameLookupsFor(ctx context.Context, userID, lang string) (*gameLookups, error) {
	langs, l, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	lk := &gameLookups{langs: langs, lang: l, userID: userID, path: gamecatalog.PathMatchmaking, prizes: map[string]map[string]any{}, groups: map[string]map[string]any{}, currencies: map[string]i18n{}}
	rows, err := s.db.QueryContext(ctx, `SELECT key, name::text FROM economy_currency`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var key, name string
		if err := rows.Scan(&key, &name); err != nil {
			rows.Close()
			return nil, err
		}
		lk.currencies[key] = parseI18n(name)
	}
	rows.Close()
	if lk.level, err = s.playerLevel(ctx, userID); err != nil {
		return nil, err
	}
	return lk, nil
}

func jsonUnmarshalString(raw string, v any) error {
	return json.Unmarshal([]byte(raw), v)
}

func parseI18n(raw string) i18n {
	var t i18n
	_ = jsonUnmarshalString(raw, &t)
	return t
}

func (s *service) gamePrizeView(ctx context.Context, lk *gameLookups, room gamecatalog.Room) (map[string]any, error) {
	if room.PrizeID != "" {
		if v, ok := lk.prizes[room.PrizeID]; ok {
			return v, nil
		}
		var key, name, items string
		err := s.db.QueryRowContext(ctx, `SELECT key, name::text, items::text FROM economy_prize WHERE id=$1`, room.PrizeID).Scan(&key, &name, &items)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		var list []any
		_ = jsonUnmarshalString(items, &list)
		v := map[string]any{"type": "prize", "id": room.PrizeID, "key": key, "name": lk.langs.Resolve(parseI18n(name), lk.lang), "items": list}
		lk.prizes[room.PrizeID] = v
		return v, nil
	}
	if room.PrizeGroupID != "" {
		if v, ok := lk.groups[room.PrizeGroupID]; ok {
			return v, nil
		}
		var key, name string
		err := s.db.QueryRowContext(ctx, `SELECT key, name::text FROM economy_prize_group WHERE id=$1`, room.PrizeGroupID).Scan(&key, &name)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		v := map[string]any{"type": "prizeGroup", "id": room.PrizeGroupID, "key": key, "name": lk.langs.Resolve(parseI18n(name), lk.lang)}
		lk.groups[room.PrizeGroupID] = v
		return v, nil
	}
	return nil, nil
}

func (s *service) gameView(ctx context.Context, g *gameRecord, lk *gameLookups) (map[string]any, error) {
	if err := s.reqFor(ctx, lk, []*gameRecord{g}); err != nil {
		return nil, err
	}
	rooms := []map[string]any{}
	order, isNew, err := s.roomsFor(ctx, lk.userID, g, lk.path)
	if err != nil {
		return nil, err
	}
	npOffered := len(s.newPlayerRooms(g)) > 0 && pathOr(lk.path) == gamecatalog.PathMatchmaking
	for _, r := range order {
		el := eligibilityFor(lk, g.Requirements, &r)
		if el.hidden {
			continue
		}
		name := lk.langs.Resolve(i18n(r.Name), lk.lang)
		if name == "" {
			name = r.Key
		}
		levelLocked := r.MinLevel > 0 && lk.level < r.MinLevel
		room := map[string]any{"key": r.Key, "name": name, "minLevel": r.MinLevel, "locked": levelLocked || !el.eligible, "levelLocked": levelLocked, "entryFee": nil, "prize": nil, "requirements": nil}
		if !r.Requirements.Empty() {
			room["requirements"] = lk.names.view(r.Requirements, lk.facts)
		}
		room["inheritsGame"] = r.Requirements.Inherits()
		if err := s.eligibilityView(ctx, lk, el, room); err != nil {
			return nil, err
		}
		if r.EntryFee != nil {
			room["entryFee"] = map[string]any{"currencyKey": r.EntryFee.CurrencyKey, "amount": r.EntryFee.Amount, "currencyName": lk.langs.Resolve(lk.currencies[r.EntryFee.CurrencyKey], lk.lang)}
		}
		prize, err := s.gamePrizeView(ctx, lk, r)
		if err != nil {
			return nil, err
		}
		if prize != nil {
			room["prize"] = prize
		}
		s.roomExtras(ctx, lk, g, r, room)
		room["joinable"] = !(isNew && npOffered) || r.NewPlayersOnly
		rooms = append(rooms, room)
	}
	def, registered := gamecatalog.Lookup(g.Key)
	view := map[string]any{
		"key": g.Key, "isActive": g.Active, "name": lk.langs.Resolve(g.Name, lk.lang), "description": lk.langs.Resolve(g.Description, lk.lang),
		"iconUrl": g.IconURL, "category": nil, "tags": g.Tags, "minPlayers": g.MinPlayers, "maxPlayers": g.MaxPlayers,
		"turnSeconds": g.TurnSeconds, "matchSeconds": g.MatchSeconds, "botsAllowed": g.BotsAllowed, "leaderboardId": g.LeaderboardID,
		"rooms": rooms, "settings": g.Settings, "sort": g.Sort, "voice": g.Voice, "realtime": registered && def.Realtime, "multiplayer": g.MaxPlayers > 1, "level": lk.level,
		"oneMatchPerPlayer": g.OneMatch, "newPlayer": isNew,
	}
	if !g.Active {
		view["maintenance"] = lk.langs.Resolve(g.Maintenance, lk.lang)
	}
	view["requirements"] = nil
	if g.Requirements != nil {
		view["requirements"] = lk.names.view(g.Requirements, lk.facts)
	}
	gel := eligibilityFor(lk, g.Requirements, nil)
	if err := s.eligibilityView(ctx, lk, gel, view); err != nil {
		return nil, err
	}
	view["path"] = lk.path
	view["hidden"] = gel.hidden || (len(g.Rooms) > 0 && len(rooms) == 0)
	if g.CategoryKey != "" {
		view["category"] = map[string]any{"key": g.CategoryKey, "name": lk.langs.Resolve(g.CategoryName, lk.lang)}
	}
	if g.Tags == nil {
		view["tags"] = []string{}
	}
	return view, nil
}

func (s *service) rpcGames(ctx context.Context, userID, payload string) (any, error) {
	var in gameRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	games, _, err := s.gameRecords(ctx)
	if err != nil {
		return nil, err
	}
	path, err := cleanPath(in.Path)
	if err != nil {
		return nil, err
	}
	lk, err := s.gameLookupsFor(ctx, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	lk.path = path
	if err := s.reqFor(ctx, lk, games); err != nil {
		return nil, err
	}
	items := []map[string]any{}
	for _, g := range games {
		if g.availability(in.Platform, in.Version) != "" {
			continue
		}
		if !g.Active && lk.langs.Resolve(g.Maintenance, lk.lang) == "" {
			continue
		}
		v, err := s.gameView(ctx, g, lk)
		if err != nil {
			return nil, err
		}
		if v["hidden"] == true {
			continue
		}
		delete(v, "hidden")
		items = append(items, v)
	}
	return map[string]any{"items": items, "level": lk.level, "lang": lk.lang, "path": lk.path}, nil
}

func (s *service) rpcGame(ctx context.Context, userID, payload string) (any, error) {
	var in gameRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	g, err := s.gameByKey(ctx, in.Key)
	if err != nil {
		return nil, err
	}
	if code := g.availability(in.Platform, in.Version); code != "" {
		return nil, gameUnavailable(code, g)
	}
	path, err := cleanPath(in.Path)
	if err != nil {
		return nil, err
	}
	lk, err := s.gameLookupsFor(ctx, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	lk.path = path
	v, err := s.gameView(ctx, g, lk)
	if err != nil {
		return nil, err
	}
	if v["hidden"] == true {
		return nil, notFound("GAME_NOT_FOUND")
	}
	delete(v, "hidden")
	return v, nil
}

func (s *service) rpcGameJoin(ctx context.Context, userID, payload string) (any, error) {
	var in gameRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.gameJoinCheck(ctx, userID, in.Key, in.Room, in.Platform, in.Version, in.Lang, in.Path)
}

func (s *service) beforeMatchmakerAdd(ctx context.Context, logger runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
	m := in.GetMatchmakerAdd()
	if m == nil {
		return in, nil
	}
	key := strings.TrimSpace(m.StringProperties["game"])
	if key == "" {
		return in, nil
	}
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	t, err := s.gameJoinCheck(ctx, userID, key, m.StringProperties["room"], m.StringProperties["platform"], m.StringProperties["version"], m.StringProperties["lang"], gamecatalog.PathMatchmaking)
	if err != nil {
		if app, ok := asApp(err); ok {
			return nil, runtime.NewError(chanceErrorMessage(app), int(grpcCode(app.Status)))
		}
		return nil, s.runtimeError(err)
	}
	if m.StringProperties == nil {
		m.StringProperties = map[string]string{}
	}
	if m.NumericProperties == nil {
		m.NumericProperties = map[string]float64{}
	}
	for k, v := range t.StringProperties {
		m.StringProperties[k] = v
	}
	if t.Room == "" {
		delete(m.StringProperties, "room")
	}
	m.NumericProperties["level"] = float64(t.Level)
	m.Query = t.Query
	m.MinCount = int32(t.MinCount)
	m.MaxCount = int32(t.MaxCount)
	m.CountMultiple = nil
	return in, nil
}

func (s *service) registerGameHooks(initializer runtime.Initializer) error {
	if err := initializer.RegisterBeforeRt("MatchmakerAdd", hookchain.Before("rt:MatchmakerAdd", s.beforeMatchmakerAdd)); err != nil {
		return err
	}
	return initializer.RegisterMatchmakerMatched(s.matchmakerMatched)
}
