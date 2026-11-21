package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"google.golang.org/grpc/codes"
)

func (s *service) chanceRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_chance_games":            s.rpcChanceGames,
		"economy_chance_game":             s.rpcChanceGame,
		"economy_chance_can_play":         s.rpcChanceCanPlay,
		"economy_chance_next_play_time":   s.rpcChanceNextPlayTime,
		"economy_chance_play":             s.rpcChancePlay,
		"economy_chance_play_with_ticket": s.rpcChancePlayWithTicket,
		"economy_chance_history":          s.rpcChanceHistory,
		"economy_chance_ack":              s.rpcChanceAck,
	}
}

func chanceErrorMessage(app *appError) string {
	body := map[string]any{"code": app.Code, "status": app.Status}
	if app.Meta != nil {
		body["meta"] = app.Meta
	}
	raw, _ := json.Marshal(body)
	return string(raw)
}

func (s *service) chanceRPC(fn clientRPC) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if _, err := uuid.FromString(userID); err != nil {
			return "", runtime.NewError(chanceErrorMessage(fail(401, "UNAUTHENTICATED", "", "")), int(codes.Unauthenticated))
		}
		if len(payload) > 16384 {
			return "", runtime.NewError(chanceErrorMessage(badRequest("REQUEST_TOO_LARGE", "", "")), int(codes.InvalidArgument))
		}
		out, err := fn(ctx, userID, payload)
		if err != nil {
			if app, ok := asApp(err); ok {
				return "", runtime.NewError(chanceErrorMessage(app), int(grpcCode(app.Status)))
			}
			return "", s.runtimeError(err)
		}
		body, err := json.Marshal(out)
		if err != nil {
			return "", runtime.NewError("INTERNAL_ERROR", int(codes.Internal))
		}
		return string(body), nil
	}
}

type chanceTarget struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Lang         string `json:"lang"`
	ExpectedCost *int64 `json:"expectedCost"`
}

func (s *service) chanceTargetGame(ctx context.Context, payload string) (*chanceGame, chanceTarget, error) {
	var in chanceTarget
	if err := decodeStrict(payload, &in); err != nil {
		return nil, in, err
	}
	id := strings.TrimSpace(in.ID)
	if id == "" && strings.TrimSpace(in.Key) != "" {
		if err := s.db.QueryRowContext(ctx, `SELECT id::text FROM economy_chance_game WHERE key=$1`, strings.TrimSpace(in.Key)).Scan(&id); err != nil {
			return nil, in, notFound("CHANCE_GAME_NOT_FOUND")
		}
	}
	g, err := loadChanceGame(ctx, s.db, id, "")
	if err != nil {
		return nil, in, err
	}
	if !g.Active {
		return nil, in, notFound("CHANCE_GAME_NOT_FOUND")
	}
	return g, in, nil
}

type chanceLookups struct {
	prizes     map[string]prizeName
	groups     map[string]chanceGroupView
	currencies map[string][2]string
}

type chanceGroupPrize struct {
	ID     string   `json:"id"`
	Key    string   `json:"key"`
	Name   string   `json:"name"`
	Weight *int     `json:"weight,omitempty"`
	Share  *float64 `json:"share,omitempty"`
}

type chanceGroupView struct {
	ID     string             `json:"id"`
	Key    string             `json:"key"`
	Name   string             `json:"name"`
	Prizes []chanceGroupPrize `json:"prizes"`
}

func (s *service) chanceLookupsFor(ctx context.Context, langs *localized.Set, lang string, games []*chanceGame) (*chanceLookups, error) {
	lk := &chanceLookups{groups: map[string]chanceGroupView{}, currencies: map[string][2]string{}}
	prizeIDs := []string{}
	groupIDs := []string{}
	currencyIDs := []string{}
	for _, g := range games {
		for _, r := range g.Rows {
			if r.PrizeID != nil {
				prizeIDs = append(prizeIDs, *r.PrizeID)
			}
			if r.PrizeGroupID != nil {
				groupIDs = append(groupIDs, *r.PrizeGroupID)
			}
		}
		if g.Price != nil {
			currencyIDs = append(currencyIDs, g.Price.CurrencyID)
		}
		for _, r := range g.Rows {
			if r.CurrencyID != "" {
				currencyIDs = append(currencyIDs, r.CurrencyID)
			}
		}
	}
	members := map[string][]groupMember{}
	if len(groupIDs) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT g.id::text, g.key, g.name::text FROM economy_prize_group g WHERE g.id::text = ANY($1)`, groupIDs)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id, key, name string
			if err := rows.Scan(&id, &key, &name); err != nil {
				rows.Close()
				return nil, err
			}
			lk.groups[id] = chanceGroupView{ID: id, Key: key, Name: langs.Resolve(localized.Parse(name), lang), Prizes: []chanceGroupPrize{}}
		}
		rows.Close()
		for id := range lk.groups {
			m, err := loadGroupMembers(ctx, s.db, id)
			if err != nil {
				return nil, err
			}
			members[id] = m
			for _, x := range m {
				prizeIDs = append(prizeIDs, x.PrizeID)
			}
		}
	}
	names, err := s.prizeNames(ctx, langs, lang, prizeIDs)
	if err != nil {
		return nil, err
	}
	lk.prizes = names
	for id, g := range lk.groups {
		list := []chanceGroupPrize{}
		total := 0
		for _, m := range members[id] {
			total += m.Weight
		}
		for _, m := range members[id] {
			share := 0.0
			if total > 0 {
				share = float64(m.Weight) / float64(total)
			}
			weight := m.Weight
			list = append(list, chanceGroupPrize{ID: m.PrizeID, Key: names[m.PrizeID].Key, Name: names[m.PrizeID].Name, Weight: &weight, Share: &share})
		}
		g.Prizes = list
		lk.groups[id] = g
	}
	if len(currencyIDs) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text FROM economy_currency WHERE id::text = ANY($1)`, currencyIDs)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id, key, name string
			if err := rows.Scan(&id, &key, &name); err != nil {
				return nil, err
			}
			lk.currencies[id] = [2]string{key, langs.Resolve(localized.Parse(name), lang)}
		}
	}
	return lk, nil
}

func limitReached(out map[string]any) bool {
	v, _ := out["limitReached"].(bool)
	return v
}

func hideGroupOdds(grp chanceGroupView) chanceGroupView {
	hidden := make([]chanceGroupPrize, 0, len(grp.Prizes))
	for _, p := range grp.Prizes {
		hidden = append(hidden, chanceGroupPrize{ID: p.ID, Key: p.Key, Name: p.Name})
	}
	grp.Prizes = hidden
	return grp
}

func (s *service) chanceViewFor(ctx context.Context, g *chanceGame, lk *chanceLookups, langs *localized.Set, lang string, st *chancePlayer, userID string, now time.Time) (map[string]any, error) {
	today, tickets := st.today[g.ID], st.tickets[g.ID]
	start, end := s.chanceDay(ctx, s.db, g, now)
	if g.DayZone != "" || g.DayStart != 0 {
		if err := s.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE mode='ticket') FROM economy_chance_play WHERE game_id=$1 AND user_id=$2 AND created_at >= $3`, g.ID, userID, start).Scan(&today, &tickets); err != nil {
			return nil, err
		}
	}
	left := freeTurnLeft(g, st.last[g.ID], start, end, now)
	view := s.chanceGameView(g, lk, langs, lang, st.last[g.ID], st.plays[g.ID], today, tickets, now, left)
	view["freeReset"], view["freeFirst"] = g.FreeReset, g.FreeFirst
	if g.FreeReset == "daily" || g.DailyLimit != nil {
		view["dayEndsAt"] = end.UTC().Format(time.RFC3339)
	}
	if g.MinLevel > 0 {
		level, err := s.chanceLevel(ctx, s.db, g, userID)
		if err != nil {
			return nil, err
		}
		locked := level < g.MinLevel
		view["minLevel"], view["level"], view["locked"] = g.MinLevel, level, locked
		if locked {
			if _, ok := view["canPlay"]; ok {
				view["canPlay"] = false
			}
		}
	}
	return view, nil
}

func (s *service) chanceGameView(g *chanceGame, lk *chanceLookups, langs *localized.Set, lang string, last sql.NullTime, plays int64, today, tickets int, now time.Time, left time.Duration) map[string]any {
	rows := []map[string]any{}
	for _, r := range g.Rows {
		var chance any = r.Chance
		if !g.ShowOdds {
			chance = nil
		}
		row := map[string]any{"chance": chance, "prizeId": r.PrizeID, "prizeGroupId": r.PrizeGroupID, "prize": nil, "group": nil}
		if r.PrizeID != nil {
			row["prize"] = prizeJSON(lk.prizes[*r.PrizeID], *r.PrizeID)
		}
		if r.PrizeGroupID != nil {
			if grp, ok := lk.groups[*r.PrizeGroupID]; ok {
				if !g.ShowOdds {
					grp = hideGroupOdds(grp)
				}
				row["group"] = grp
			}
		}
		if g.PrizeType == "slots" {
			row["kind"] = slotKind(r)
			if r.CurrencyID != "" {
				cur := lk.currencies[r.CurrencyID]
				row["currency"] = map[string]any{"currencyId": r.CurrencyID, "currencyKey": cur[0], "currencyName": cur[1], "min": r.Min, "max": r.Max}
			}
		} else if r.Jackpot {
			row["jackpot"] = true
		}
		rows = append(rows, row)
	}
	var meta any = map[string]any{}
	_ = json.Unmarshal(g.Metadata, &meta)
	out := map[string]any{"id": g.ID, "key": g.Key, "name": langs.Resolve(g.Name, lang), "description": langs.Resolve(g.Description, lang), "metadata": meta,
		"prizeType": g.PrizeType, "prize": rows, "intervalTime": g.Interval, "price": nil, "plays": plays, "sort": g.Sort,
		"kind": g.Kind, "showOdds": g.ShowOdds, "dailyLimit": g.DailyLimit, "totalLimit": g.TotalLimit, "playsToday": today, "playsLeftToday": nil, "playsLeft": nil, "limitReached": false}
	if g.DailyLimit != nil {
		left := *g.DailyLimit - today
		if left < 0 {
			left = 0
		}
		out["playsLeftToday"] = left
		if left == 0 {
			out["limitReached"] = true
		}
	}
	if g.TotalLimit != nil {
		left := int64(*g.TotalLimit) - plays
		if left < 0 {
			left = 0
		}
		out["playsLeft"] = left
		if left == 0 {
			out["limitReached"] = true
		}
	}
	out["revealAck"] = g.RevealAck
	if g.Price != nil {
		cur := lk.currencies[g.Price.CurrencyID]
		amount := g.Price.Amount
		if len(g.Ladder) > 0 {
			amount = ladderAmount(g.Ladder, tickets)
			out["priceLadder"] = g.Ladder
			out["ticketsToday"] = tickets
		}
		out["price"] = map[string]any{"currencyId": g.Price.CurrencyID, "currencyKey": cur[0], "currencyName": cur[1], "amount": amount}
	}
	if g.hasFreeTurn() {
		out["canPlay"] = left <= 0 && !limitReached(out)
		out["nextPlayTime"] = waitSeconds(left)
		if left > 0 {
			out["nextPlayAt"] = now.Add(left).UTC().Format(time.RFC3339)
		}
		if last.Valid {
			out["lastPlayAt"] = last.Time.UTC().Format(time.RFC3339)
		}
	}
	return out
}

type chancePlayer struct {
	last    map[string]sql.NullTime
	plays   map[string]int64
	today   map[string]int
	tickets map[string]int
}

func (s *service) chancePlayerState(ctx context.Context, userID string) (map[string]sql.NullTime, map[string]int64, map[string]int, error) {
	st, err := s.chancePlayerFull(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	return st.last, st.plays, st.today, nil
}

func (s *service) chancePlayerFull(ctx context.Context, userID string) (*chancePlayer, error) {
	st := &chancePlayer{last: map[string]sql.NullTime{}, plays: map[string]int64{}, today: map[string]int{}, tickets: map[string]int{}}
	start := s.chanceDayStart(ctx, s.db, s.now().UTC())
	rows, err := s.db.QueryContext(ctx, `SELECT c.game_id::text, c.plays, (SELECT max(p.created_at) FROM economy_chance_play p WHERE p.game_id=c.game_id AND p.user_id=c.user_id AND p.mode='free'),
		(SELECT count(*) FROM economy_chance_play p WHERE p.game_id=c.game_id AND p.user_id=c.user_id AND p.created_at >= $2),
		(SELECT count(*) FROM economy_chance_play p WHERE p.game_id=c.game_id AND p.user_id=c.user_id AND p.created_at >= $2 AND p.mode='ticket')
		FROM economy_chance_state c WHERE c.user_id=$1`, userID, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var n int64
		var t sql.NullTime
		var d, k int
		if err := rows.Scan(&id, &n, &t, &d, &k); err != nil {
			return nil, err
		}
		st.last[id] = t
		st.plays[id] = n
		st.today[id] = d
		st.tickets[id] = k
	}
	return st, rows.Err()
}

func (s *service) rpcChanceGames(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	games, err := loadChanceGames(ctx, s.db, ` WHERE is_active`)
	if err != nil {
		return nil, err
	}
	lk, err := s.chanceLookupsFor(ctx, langs, lang, games)
	if err != nil {
		return nil, err
	}
	st, err := s.chancePlayerFull(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	items := []map[string]any{}
	for _, g := range games {
		view, err := s.chanceViewFor(ctx, g, lk, langs, lang, st, userID, now)
		if err != nil {
			return nil, err
		}
		if err := s.addPendingReveal(ctx, view, g, userID, langs, lang); err != nil {
			return nil, err
		}
		items = append(items, view)
	}
	return map[string]any{"code": "SUCCESS", "lang": lang, "items": items}, nil
}

func (s *service) rpcChanceGame(ctx context.Context, userID, payload string) (any, error) {
	g, in, err := s.chanceTargetGame(ctx, payload)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	lk, err := s.chanceLookupsFor(ctx, langs, lang, []*chanceGame{g})
	if err != nil {
		return nil, err
	}
	st, err := s.chancePlayerFull(ctx, userID)
	if err != nil {
		return nil, err
	}
	view, err := s.chanceViewFor(ctx, g, lk, langs, lang, st, userID, s.now().UTC())
	if err != nil {
		return nil, err
	}
	if err := s.addPendingReveal(ctx, view, g, userID, langs, lang); err != nil {
		return nil, err
	}
	return map[string]any{"code": "SUCCESS", "lang": lang, "game": view}, nil
}

func (s *service) chanceTurn(ctx context.Context, userID, payload string) (time.Duration, error) {
	g, _, err := s.chanceTargetGame(ctx, payload)
	if err != nil {
		return 0, err
	}
	if !g.hasFreeTurn() {
		return 0, badRequest("INTERVAL_TIME_IS_NULL", "", "this game has no free turn")
	}
	return s.chanceFreeLeft(ctx, s.db, g, userID, s.now().UTC())
}

func (s *service) rpcChanceCanPlay(ctx context.Context, userID, payload string) (any, error) {
	left, err := s.chanceTurn(ctx, userID, payload)
	if err != nil {
		return nil, err
	}
	return map[string]any{"code": "SUCCESS", "canPlay": left <= 0, "nextPlayTime": waitSeconds(left)}, nil
}

func (s *service) rpcChanceNextPlayTime(ctx context.Context, userID, payload string) (any, error) {
	left, err := s.chanceTurn(ctx, userID, payload)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"code": "SUCCESS", "nextPlayTime": waitSeconds(left)}
	if left > 0 {
		out["nextPlayAt"] = s.now().UTC().Add(left).Format(time.RFC3339)
	}
	return out, nil
}

func (s *service) rpcChancePlay(ctx context.Context, userID, payload string) (any, error) {
	return s.chancePlayRPC(ctx, userID, payload, "free")
}

func (s *service) rpcChancePlayWithTicket(ctx context.Context, userID, payload string) (any, error) {
	return s.chancePlayRPC(ctx, userID, payload, "ticket")
}

func (s *service) chancePlayRPC(ctx context.Context, userID, payload, mode string) (any, error) {
	g, in, err := s.chanceTargetGame(ctx, payload)
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	res, err := s.playChance(ctx, userID, g.ID, mode, in.ExpectedCost)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"code": "SUCCESS", "lang": lang, "prizeId": nil, "prize": nil, "prizeGroupId": nil, "group": nil, "playId": res.Play.ID, "gameId": res.Game.ID,
		"gameKey": res.Game.Key, "mode": mode, "slot": nil, "roll": float64(res.Play.Roll) / chanceScale, "rewards": res.Rewards, "spent": nil, "status": res.Play.Status}
	if res.Play.Slot.Valid {
		out["slot"] = res.Play.Slot.Int64
	}
	if res.Play.PrizeID.Valid {
		names, err := s.prizeNames(ctx, langs, lang, []string{res.Play.PrizeID.String})
		if err != nil {
			return nil, err
		}
		out["prizeId"] = res.Play.PrizeID.String
		out["prize"] = prizeJSON(names[res.Play.PrizeID.String], res.Play.PrizeID.String)
	}
	if res.Play.GroupID.Valid {
		out["prizeGroupId"] = res.Play.GroupID.String
		var key, name string
		if err := s.db.QueryRowContext(ctx, `SELECT key, name::text FROM economy_prize_group WHERE id=$1`, res.Play.GroupID.String).Scan(&key, &name); err == nil {
			out["group"] = map[string]any{"id": res.Play.GroupID.String, "key": key, "name": langs.Resolve(localized.Parse(name), lang)}
		}
	}
	if res.Play.Price != nil {
		out["spent"] = map[string]any{"currencyKey": res.Play.Price.CurrencyKey, "amount": res.Play.Price.Amount}
	}
	if res.Play.Outcome != nil {
		out["outcome"] = s.outcomeView(ctx, res.Play.Outcome, langs, lang)
	}
	if res.Game.RevealAck {
		out["needsAck"] = res.Play.Status == "done"
	}
	if len(res.Game.Ladder) > 0 && res.Game.Price != nil {
		start, _ := s.chanceDay(ctx, s.db, res.Game, s.now().UTC())
		if n, err := s.chanceTicketsSince(ctx, s.db, res.Game.ID, userID, start); err == nil {
			out["nextTicketCost"] = ladderAmount(res.Game.Ladder, n)
		}
	}
	if res.Game.hasFreeTurn() {
		out["nextPlayTime"] = waitSeconds(res.Wait)
		if res.Wait > 0 {
			out["nextPlayAt"] = s.now().UTC().Add(res.Wait).Format(time.RFC3339)
		}
	}
	out["kind"] = res.Game.Kind
	if res.Game.DailyLimit != nil || res.Game.TotalLimit != nil {
		_, plays, today, err := s.chancePlayerState(ctx, userID)
		if err == nil {
			if res.Game.DailyLimit != nil {
				n := today[res.Game.ID]
				if res.Game.DayZone != "" || res.Game.DayStart != 0 {
					start, _ := s.chanceDay(ctx, s.db, res.Game, s.now().UTC())
					if c, err := s.chancePlaysSince(ctx, s.db, res.Game.ID, userID, start); err == nil {
						n = c
					}
				}
				out["playsLeftToday"] = max(*res.Game.DailyLimit-n, 0)
			}
			if res.Game.TotalLimit != nil {
				out["playsLeft"] = max(int64(*res.Game.TotalLimit)-plays[res.Game.ID], 0)
			}
		}
	}
	if wallet, err := s.userWallet(ctx, userID); err == nil {
		out["wallet"] = wallet
	}
	return out, nil
}

func (s *service) rpcChanceHistory(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID     string `json:"id"`
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor"`
		Lang   string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Limit < 1 || in.Limit > 100 {
		in.Limit = 20
	}
	offset, _ := strconv.Atoi(in.Cursor)
	if offset < 0 {
		offset = 0
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	where := &sqlWhere{}
	where.add("p.user_id = ?::uuid", userID)
	if id := strings.TrimSpace(in.ID); id != "" {
		if _, err := uuid.FromString(id); err != nil {
			return nil, notFound("CHANCE_GAME_NOT_FOUND")
		}
		where.add("p.game_id = ?::uuid", id)
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.game_id::text, g.key, g.name::text, p.mode, p.status, p.roll, p.slot, p.prize_id::text, p.prize_group_id::text, p.price::text, p.created_at,
			COALESCE((SELECT up.rewards::text FROM economy_user_prize up WHERE up.source_type='chance' AND up.source_id=p.claim_id::text AND up.claimed), '[]'), p.outcome::text
		FROM economy_chance_play p JOIN economy_chance_game g ON g.id=p.game_id`+where.sql()+` ORDER BY p.created_at DESC, p.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
		append(where.args, in.Limit+1, offset)...)
	if err != nil {
		return nil, err
	}
	type histRow struct {
		id, gameID, gameKey, gameName, mode, status, rewards string
		roll                                                 int
		slot                                                 sql.NullInt64
		prize, group, price, outcome                         sql.NullString
		created                                              time.Time
	}
	var list []histRow
	prizeIDs := []string{}
	for rows.Next() {
		var r histRow
		if err := rows.Scan(&r.id, &r.gameID, &r.gameKey, &r.gameName, &r.mode, &r.status, &r.roll, &r.slot, &r.prize, &r.group, &r.price, &r.created, &r.rewards, &r.outcome); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, r)
		if r.prize.Valid {
			prizeIDs = append(prizeIDs, r.prize.String)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	names, err := s.prizeNames(ctx, langs, lang, prizeIDs)
	if err != nil {
		return nil, err
	}
	cursor := ""
	if len(list) > in.Limit {
		list = list[:in.Limit]
		cursor = strconv.Itoa(offset + in.Limit)
	}
	items := []map[string]any{}
	for _, r := range list {
		item := map[string]any{"playId": r.id, "gameId": r.gameID, "gameKey": r.gameKey, "gameName": langs.Resolve(localized.Parse(r.gameName), lang), "mode": r.mode, "status": r.status,
			"roll": float64(r.roll) / chanceScale, "slot": nil, "prizeId": nil, "prize": nil, "prizeGroupId": nil, "spent": nil, "createdAt": r.created.UTC().Format(time.RFC3339)}
		if r.slot.Valid {
			item["slot"] = r.slot.Int64
		}
		if r.prize.Valid {
			item["prizeId"] = r.prize.String
			item["prize"] = prizeJSON(names[r.prize.String], r.prize.String)
		}
		if r.group.Valid {
			item["prizeGroupId"] = r.group.String
		}
		if r.price.Valid && r.price.String != "" && r.price.String != "null" {
			var cp coinPrice
			if json.Unmarshal([]byte(r.price.String), &cp) == nil {
				item["spent"] = map[string]any{"currencyKey": cp.CurrencyKey, "amount": cp.Amount}
			}
		}
		var rewards []prizeItem
		_ = json.Unmarshal([]byte(r.rewards), &rewards)
		if r.outcome.Valid && r.outcome.String != "" && r.outcome.String != "null" {
			var oc []chanceOutcome
			if json.Unmarshal([]byte(r.outcome.String), &oc) == nil {
				rewards = slotRewards(oc)
				item["outcome"] = s.outcomeView(ctx, oc, langs, lang)
			}
		}
		if rewards == nil {
			rewards = []prizeItem{}
		}
		item["rewards"] = rewards
		items = append(items, item)
	}
	return map[string]any{"code": "SUCCESS", "lang": lang, "items": items, "cursor": cursor}, nil
}

func (s *service) outcomeView(ctx context.Context, outcome []chanceOutcome, langs *localized.Set, lang string) []map[string]any {
	ids := []string{}
	for _, o := range outcome {
		if o.PrizeID != "" {
			ids = append(ids, o.PrizeID)
		}
	}
	names, _ := s.prizeNames(ctx, langs, lang, ids)
	out := []map[string]any{}
	for _, o := range outcome {
		if o.Skipped {
			continue
		}
		row := map[string]any{"slot": o.Slot, "kind": o.Kind}
		if o.Kind == "currency" {
			row["currencyKey"], row["amount"] = o.CurrencyKey, o.Amount
		} else {
			row["prizeId"] = o.PrizeID
			row["prize"] = prizeJSON(names[o.PrizeID], o.PrizeID)
			if o.PrizeGroupID != "" {
				row["prizeGroupId"] = o.PrizeGroupID
			}
			rewards := o.Rewards
			if rewards == nil {
				rewards = []prizeItem{}
			}
			row["rewards"] = rewards
		}
		out = append(out, row)
	}
	return out
}

func (s *service) addPendingReveal(ctx context.Context, view map[string]any, g *chanceGame, userID string, langs *localized.Set, lang string) error {
	if !g.RevealAck {
		return nil
	}
	view["pendingReveal"] = nil
	var id, mode string
	var slot sql.NullInt64
	var prize, outcome, claim sql.NullString
	var at time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id::text, mode, slot, prize_id::text, outcome::text, claim_id::text, created_at FROM economy_chance_play
		WHERE game_id=$1 AND user_id=$2 AND status='done' AND acked_at IS NULL ORDER BY created_at DESC, id DESC LIMIT 1`, g.ID, userID).Scan(&id, &mode, &slot, &prize, &outcome, &claim, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	rev := map[string]any{"playId": id, "mode": mode, "slot": nil, "prizeId": nil, "prize": nil, "rewards": []prizeItem{}, "createdAt": at.UTC().Format(time.RFC3339)}
	if slot.Valid {
		rev["slot"] = slot.Int64
	}
	if prize.Valid {
		names, err := s.prizeNames(ctx, langs, lang, []string{prize.String})
		if err != nil {
			return err
		}
		rev["prizeId"] = prize.String
		rev["prize"] = prizeJSON(names[prize.String], prize.String)
	}
	if claim.Valid {
		if rewards, err := s.claimRewards(ctx, claim.String); err == nil {
			rev["rewards"] = rewards
		}
	}
	if outcome.Valid && outcome.String != "" && outcome.String != "null" {
		var oc []chanceOutcome
		if json.Unmarshal([]byte(outcome.String), &oc) == nil {
			rev["outcome"] = s.outcomeView(ctx, oc, langs, lang)
			rev["rewards"] = slotRewards(oc)
		}
	}
	view["pendingReveal"] = rev
	return nil
}

func (s *service) rpcChanceAck(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		PlayID string `json:"playId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(strings.TrimSpace(in.PlayID)); err != nil {
		return nil, notFound("PLAY_NOT_FOUND")
	}
	var acked time.Time
	err := s.db.QueryRowContext(ctx, `UPDATE economy_chance_play SET acked_at=COALESCE(acked_at, now()) WHERE id=$1 AND user_id=$2 RETURNING acked_at`, strings.TrimSpace(in.PlayID), userID).Scan(&acked)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("PLAY_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"code": "SUCCESS", "acknowledged": true, "playId": strings.TrimSpace(in.PlayID), "ackedAt": acked.UTC().Format(time.RFC3339)}, nil
}
