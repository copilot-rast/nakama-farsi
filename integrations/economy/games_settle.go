package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	gamePotReason         = "game_pot"
	gameActionDays        = 60
	gameActionDaysSetting = "actionLogDays"
	gameStaleLiveAfter    = 2 * time.Minute
	maxSettleResults      = 100
	maxLogActions         = 200
)

func (s *service) walletChangeOverdraft(ctx context.Context, tx *sql.Tx, userID, grantID string, changes map[string]int64, meta map[string]any) error {
	md := map[string]any{"module": "economy", "overdraft": true}
	for k, v := range meta {
		md[k] = v
	}
	return pluginkit.WalletTxOverdraft(ctx, tx, userID, grantID, changes, md)
}

func (s *service) isBot(ctx context.Context, userID string) bool {
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_bot WHERE user_id::text=$1)`, userID).Scan(&ok)
	return ok
}

func (s *service) oneMatchGate(ctx context.Context, userID, matchID string) error {
	a, err := s.gameActive(ctx, userID)
	if err != nil {
		return err
	}
	if a == nil || a.MatchID == matchID {
		return nil
	}
	e := conflict("ACTIVE_MATCH_EXISTS", "finish or leave your other match first")
	e.Meta = map[string]any{"matchId": a.MatchID, "game": a.Game, "room": a.Room, "status": a.Status}
	return e
}

func (s *service) gameActive(ctx context.Context, userID string) (*gamecatalog.ActiveMatch, error) {
	if _, err := uuid.FromString(userID); err != nil {
		return nil, badRequest("INVALID_ID", "userId", "must be a UUID")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT m.id, m.game_key, m.room_key, m.status, m.created_at FROM economy_game_entry e JOIN economy_game_match m ON m.id=e.match_id
		WHERE e.user_id=$1 AND e.status IN ('paid','free') AND NOT e.released AND NOT e.bot AND m.status IN ('open','started') ORDER BY m.created_at DESC LIMIT 5`, userID)
	if err != nil {
		return nil, err
	}
	var found []gamecatalog.ActiveMatch
	for rows.Next() {
		var a gamecatalog.ActiveMatch
		if err := rows.Scan(&a.MatchID, &a.Game, &a.Room, &a.Status, &a.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		found = append(found, a)
	}
	rows.Close()
	for i := range found {
		a := found[i]
		if s.nk != nil && strings.Contains(a.MatchID, ".") && s.now().Sub(a.CreatedAt) > gameStaleLiveAfter {
			if m, err := s.nk.MatchGet(ctx, a.MatchID); err != nil || m == nil {
				_, _ = s.db.ExecContext(ctx, `UPDATE economy_game_entry SET released=true, updated_at=now() WHERE match_id=$1 AND user_id=$2`, a.MatchID, userID)
				continue
			}
		}
		return &a, nil
	}
	return nil, nil
}

func (s *service) gameRelease(ctx context.Context, matchID, userID string) error {
	matchID, err := cleanMatchID(matchID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE economy_game_entry SET released=true, updated_at=now() WHERE match_id=$1 AND user_id::text=$2`, matchID, userID)
	return err
}

func (s *service) newPlayerRooms(g *gameRecord) map[string]gamecatalog.Room {
	live := map[string]bool{}
	for _, r := range g.Rooms {
		if !r.NewPlayersOnly && !r.Disabled {
			live[r.Key] = true
		}
	}
	out := map[string]gamecatalog.Room{}
	for _, r := range g.Rooms {
		if r.NewPlayersOnly && !r.Disabled && r.Replaces != "" && live[r.Replaces] {
			if _, taken := out[r.Replaces]; !taken {
				out[r.Replaces] = r
			}
		}
	}
	return out
}

func (s *service) isNewPlayer(ctx context.Context, userID, gameKey string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE((SELECT wins+losses FROM economy_game_stats WHERE user_id::text=$1 AND game_key=$2),0)`, userID, gameKey).Scan(&n)
	return n == 0, err
}

func (s *service) roomsFor(ctx context.Context, userID string, g *gameRecord, path string) ([]gamecatalog.Room, bool, error) {
	np := s.newPlayerRooms(g)
	isNew := false
	if len(np) > 0 && userID != "" {
		var err error
		if isNew, err = s.isNewPlayer(ctx, userID, g.Key); err != nil {
			return nil, false, err
		}
	}
	out := []gamecatalog.Room{}
	for _, r := range g.Rooms {
		if r.Disabled || r.NewPlayersOnly {
			continue
		}
		if swap, ok := np[r.Key]; ok && isNew && pathOr(path) == gamecatalog.PathMatchmaking {
			out = append(out, swap)
			continue
		}
		out = append(out, r)
	}
	return out, isNew, nil
}

func (s *service) newPlayerGate(ctx context.Context, userID string, g *gameRecord, room *gamecatalog.Room, path string) error {
	if pathOr(path) != gamecatalog.PathMatchmaking {
		if room.NewPlayersOnly {
			return notFound("ROOM_NOT_FOUND")
		}
		return nil
	}
	np := s.newPlayerRooms(g)
	offered := false
	for _, r := range np {
		offered = offered || r.Key == room.Key
	}
	if room.NewPlayersOnly && !offered {
		return notFound("ROOM_NOT_FOUND")
	}
	if len(np) == 0 {
		return nil
	}
	isNew, err := s.isNewPlayer(ctx, userID, g.Key)
	if err != nil {
		return err
	}
	if room.NewPlayersOnly && !isNew {
		return fail(http.StatusForbidden, "VETERAN_ROOM_FORBIDDEN", "room", "this room is only for new players")
	}
	if !room.NewPlayersOnly && isNew {
		e := fail(http.StatusForbidden, "NEW_PLAYER_ROOM_REQUIRED", "room", "new players start in the new player rooms")
		keys := []string{}
		for _, r := range np {
			keys = append(keys, r.Key)
		}
		sort.Strings(keys)
		e.Meta = map[string]any{"rooms": keys}
		if r, ok := np[room.Key]; ok {
			e.Meta["room"] = r.Key
		}
		return e
	}
	return nil
}

func (s *service) roomExtras(ctx context.Context, lk *gameLookups, g *gameRecord, r gamecatalog.Room, room map[string]any) {
	pub := g.public()
	seats := r.Seats(pub)
	room["players"] = seats
	room["humanSeats"] = r.Humans(pub)
	room["botFillSeconds"] = r.BotFillSeconds
	room["newPlayersOnly"] = r.NewPlayersOnly
	room["winnerXp"], room["loserXp"] = r.WinnerXP, r.LoserXP
	room["payout"] = map[string]any{"mode": gamecatalog.PayoutPrize}
	if r.PotMode() {
		pay := map[string]any{"mode": gamecatalog.PayoutPot, "percent": r.PayoutPercent}
		if r.EntryFee != nil {
			pay["currencyKey"] = r.EntryFee.CurrencyKey
			pay["amount"] = gamecatalog.PotPayout(r.EntryFee.Amount*int64(seats), r.PayoutPercent)
		}
		room["payout"] = pay
	}
	room["loserPrize"] = nil
	if r.LoserPrizeID != "" {
		if v, err := s.gamePrizeView(ctx, lk, gamecatalog.Room{PrizeID: r.LoserPrizeID}); err == nil && v != nil {
			room["loserPrize"] = v
		}
	}
}

type settleEntry struct {
	user, status string
	attempt      int
	fee, tickets sql.NullString
	bot          bool
}

func (s *service) gameSettle(ctx context.Context, st gamecatalog.Settlement) ([]gamecatalog.Payout, error) {
	matchID, err := cleanMatchID(st.MatchID)
	if err != nil {
		return nil, err
	}
	if len(st.Results) > maxSettleResults {
		return nil, badRequest("TOO_MANY_RESULTS", "results", "at most 100 results")
	}
	results := map[string]gamecatalog.Result{}
	for i, r := range st.Results {
		if _, err := uuid.FromString(r.UserID); err != nil {
			return nil, badRequest("INVALID_ID", "results["+strconv.Itoa(i)+"].userId", "must be a UUID")
		}
		if !gamecatalog.ValidResult(r.Status) {
			return nil, badRequest("INVALID_RESULT", "results["+strconv.Itoa(i)+"].status", "won, lost, forfeit, timeout or refunded")
		}
		results[r.UserID] = r
	}
	var gameKey, roomKey string
	first := false
	winners := []string{}
	type prizeDue struct{ user, prize, part string }
	dues := []prizeDue{}
	refundedTickets := map[string]*gameTickets{}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var status string
		var startedAt, createdAt sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT status, game_key, room_key, started_at, created_at FROM economy_game_match WHERE id=$1 FOR UPDATE`, matchID).Scan(&status, &gameKey, &roomKey, &startedAt, &createdAt)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("MATCH_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		switch status {
		case "cancelled":
			return conflict("MATCH_CANCELLED", "this match was cancelled")
		case "finished":
			return nil
		}
		first = true
		var room *gamecatalog.Room
		if g, gerr := s.gameByKey(ctx, gameKey); gerr == nil {
			room, _ = g.room(roomKey)
		}
		if room == nil {
			room = &gamecatalog.Room{}
		}
		rows, err := tx.QueryContext(ctx, `SELECT user_id::text, status, attempt, fee::text, tickets::text, bot FROM economy_game_entry WHERE match_id=$1 ORDER BY user_id FOR UPDATE`, matchID)
		if err != nil {
			return err
		}
		var entries []settleEntry
		for rows.Next() {
			var e settleEntry
			if err := rows.Scan(&e.user, &e.status, &e.attempt, &e.fee, &e.tickets, &e.bot); err != nil {
				rows.Close()
				return err
			}
			entries = append(entries, e)
		}
		rows.Close()
		potCur := ""
		if room.EntryFee != nil {
			potCur = room.EntryFee.CurrencyKey
		}
		var pot int64
		playing := []settleEntry{}
		for _, e := range entries {
			if e.status != "paid" && e.status != "free" {
				continue
			}
			r, ok := results[e.user]
			if !ok {
				r = gamecatalog.Result{UserID: e.user, Status: gamecatalog.ResultLost}
				results[e.user] = r
			}
			if r.Status == gamecatalog.ResultRefunded {
				t, err := returnTickets(ctx, tx, matchID, e.user, e.tickets)
				if err != nil {
					return err
				}
				if t != nil {
					refundedTickets[e.user] = t
				}
				if _, err := s.refundEntry(ctx, tx, matchID, e.user, e.status, e.attempt, e.fee, gameKey); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET result='refunded', rank=$3, updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, e.user, nullInt(r.Rank)); err != nil {
					return err
				}
				continue
			}
			if e.status == "paid" && e.fee.Valid {
				var fee gameFee
				_ = json.Unmarshal([]byte(e.fee.String), &fee)
				if fee.CurrencyKey == potCur {
					pot += fee.Amount
				}
			}
			playing = append(playing, e)
			if r.Status == gamecatalog.ResultWon {
				winners = append(winners, e.user)
			}
		}
		var share int64
		if room.PotMode() && potCur != "" && len(winners) > 0 {
			share = gamecatalog.PotPayout(pot, room.PayoutPercent) / int64(len(winners))
		}
		played := int64(0)
		if startedAt.Valid {
			played = int64(s.now().Sub(startedAt.Time).Seconds())
		}
		if played < 0 {
			played = 0
		}
		for _, e := range playing {
			r := results[e.user]
			var amountJSON any
			won := int64(0)
			if r.Status == gamecatalog.ResultWon && share > 0 {
				if err := s.walletChange(ctx, tx, e.user, gameGrantID(matchID, e.user, "pot"), map[string]int64{potCur: share}, map[string]any{
					"reason": gamePotReason, "sourceType": gameSourceType, "sourceId": matchID, "game": gameKey, "room": roomKey, "part": "pot", "pot": pot, "percent": room.PayoutPercent,
				}); err != nil {
					return err
				}
				raw, _ := json.Marshal(gameFee{CurrencyKey: potCur, Amount: share})
				amountJSON = string(raw)
				won = share
			}
			prizeID := ""
			part := "prize"
			switch r.Status {
			case gamecatalog.ResultWon:
				prizeID = room.PrizeID
				if prizeID == "" && room.PrizeGroupID != "" {
					if prizeID, err = s.drawFromPrizeGroup(ctx, tx, room.PrizeGroupID); err != nil {
						return err
					}
				}
			case gamecatalog.ResultLost:
				prizeID, part = room.LoserPrizeID, "loser"
			}
			if prizeID != "" {
				dues = append(dues, prizeDue{e.user, prizeID, part})
			}
			if _, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET result=$3, rank=$4, amount=$5::jsonb, prize_id=COALESCE(NULLIF($6,'')::uuid, prize_id),
				prize_status=CASE WHEN NULLIF($6,'') IS NOT NULL AND prize_status='none' THEN 'pending' ELSE prize_status END, updated_at=now() WHERE match_id=$1 AND user_id=$2`,
				matchID, e.user, r.Status, nullInt(r.Rank), amountJSON, prizeID); err != nil {
				return err
			}
			if !e.bot {
				if err := s.addGameStats(ctx, tx, e.user, gameKey, r.Status, played, potCur, won); err != nil {
					return err
				}
			}
		}
		list := make([]gamecatalog.Result, 0, len(results))
		for _, r := range results {
			list = append(list, r)
		}
		sort.Slice(list, func(i, j int) bool {
			if list[i].Rank != list[j].Rank {
				return list[i].Rank < list[j].Rank
			}
			return list[i].UserID < list[j].UserID
		})
		resultsJSON, _ := json.Marshal(list)
		var summary any
		if len(st.Summary) > 0 && json.Valid(st.Summary) {
			summary = string(st.Summary)
		}
		reason := strings.TrimSpace(st.Reason)
		if reason == "" {
			reason = "finished"
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_game_match SET status='finished', winners=ARRAY(SELECT jsonb_array_elements_text($2::jsonb)), results=$3::jsonb, end_reason=$4,
			forced_by=NULLIF($5,''), note=NULLIF($6,''), summary=COALESCE($7::jsonb, summary), started_at=COALESCE(started_at, now()), ended_at=now() WHERE id=$1`,
			matchID, stringsJSON(winners), string(resultsJSON), truncate(reason, 64), truncate(st.ForcedBy, 200), truncate(st.Note, 500), summary)
		return err
	})
	if err != nil {
		return nil, err
	}
	for user, t := range refundedTickets {
		s.logTickets(ctx, user, matchID, gameKey, "", "game_ticket_returned", t)
	}
	for _, d := range dues {
		mid, uid := matchID, d.user
		if _, err := s.grant(ctx, grantRequest{UserID: d.user, PrizeID: d.prize, SourceType: gameSourceType, SourceID: gameGrantID(matchID, d.user, d.part), Reason: gamePrizeReason,
			After: func(ctx context.Context, tx *sql.Tx) error {
				_, err := tx.ExecContext(ctx, `UPDATE economy_game_entry SET prize_status='paid', error=NULL, updated_at=now() WHERE match_id=$1 AND user_id=$2`, mid, uid)
				return err
			}}); err != nil {
			code := err.Error()
			if app, ok := asApp(err); ok {
				code = app.Code
			}
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_game_entry SET error=$3, updated_at=now() WHERE match_id=$1 AND user_id=$2`, matchID, d.user, truncate(code, 500))
		}
	}
	if first {
		humans := []string{}
		for _, w := range winners {
			if !s.isBot(ctx, w) {
				humans = append(humans, w)
			}
		}
		s.writeGameLeaderboard(ctx, gameKey, humans)
	}
	return s.settledPayouts(ctx, matchID)
}

func nullInt(n int) any {
	if n <= 0 {
		return nil
	}
	return n
}

func (s *service) settledPayouts(ctx context.Context, matchID string) ([]gamecatalog.Payout, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text, COALESCE(result,''), COALESCE(rank,0), COALESCE(prize_id::text,''), prize_status, COALESCE(amount::text,''), status, COALESCE(fee::text,''), COALESCE(error,'')
		FROM economy_game_entry WHERE match_id=$1 AND status IN ('paid','free','refunded') ORDER BY COALESCE(rank, 1000), user_id`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []gamecatalog.Payout{}
	for rows.Next() {
		var p gamecatalog.Payout
		var prizeStatus, amount, status, fee string
		if err := rows.Scan(&p.UserID, &p.Status, &p.Rank, &p.PrizeID, &prizeStatus, &amount, &status, &fee, &p.Error); err != nil {
			return nil, err
		}
		p.Rewards = []map[string]any{}
		if amount != "" {
			var f gameFee
			if json.Unmarshal([]byte(amount), &f) == nil && f.Amount > 0 {
				p.Amount = &gamecatalog.Fee{CurrencyKey: f.CurrencyKey, Amount: f.Amount}
			}
		}
		if status == "refunded" && fee != "" {
			var f gameFee
			if json.Unmarshal([]byte(fee), &f) == nil && f.Amount > 0 {
				p.Refunded = &gamecatalog.Fee{CurrencyKey: f.CurrencyKey, Amount: f.Amount}
			}
		}
		if p.PrizeID != "" && prizeStatus == "paid" {
			if prize, err := loadPrize(ctx, s.db, p.PrizeID); err == nil {
				for _, it := range prize.Items {
					raw, _ := json.Marshal(it)
					m := map[string]any{}
					_ = json.Unmarshal(raw, &m)
					p.Rewards = append(p.Rewards, m)
				}
			}
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *service) addGameStats(ctx context.Context, tx *sql.Tx, userID, gameKey, status string, played int64, currency string, won int64) error {
	var games, wins, losses, forfeits, timeouts int
	switch status {
	case gamecatalog.ResultWon:
		games, wins = 1, 1
	case gamecatalog.ResultLost:
		games, losses = 1, 1
	case gamecatalog.ResultForfeit:
		forfeits = 1
	case gamecatalog.ResultTimeout:
		games, timeouts = 1, 1
	default:
		return nil
	}
	winnings := "{}"
	if won > 0 && currency != "" {
		raw, _ := json.Marshal(map[string]int64{currency: won})
		winnings = string(raw)
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO economy_game_stats AS t (user_id, game_key, games, wins, losses, forfeits, timeouts, playtime_seconds, winnings, last_played_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb, now(), now())
		ON CONFLICT (user_id, game_key) DO UPDATE SET games=t.games+EXCLUDED.games, wins=t.wins+EXCLUDED.wins, losses=t.losses+EXCLUDED.losses, forfeits=t.forfeits+EXCLUDED.forfeits,
		timeouts=t.timeouts+EXCLUDED.timeouts, playtime_seconds=t.playtime_seconds+EXCLUDED.playtime_seconds,
		winnings=(SELECT COALESCE(jsonb_object_agg(k, v), '{}'::jsonb) FROM (SELECT k, sum(v)::bigint AS v FROM (SELECT key AS k, value::bigint AS v FROM jsonb_each_text(t.winnings) UNION ALL SELECT key, value::bigint FROM jsonb_each_text(EXCLUDED.winnings)) x GROUP BY k) y),
		last_played_at=now(), updated_at=now()`, userID, gameKey, games, wins, losses, forfeits, timeouts, played, winnings)
	return err
}

func (s *service) pickBots(ctx context.Context, q gamecatalog.BotQuery) ([]gamecatalog.Bot, error) {
	if q.Count <= 0 {
		return []gamecatalog.Bot{}, nil
	}
	if q.Count > 50 {
		q.Count = 50
	}
	exclude := []string{}
	for _, id := range q.Exclude {
		if _, err := uuid.FromString(id); err == nil {
			exclude = append(exclude, id)
		}
	}
	cur, amount := "", int64(0)
	if q.Fee != nil && q.Fee.Amount > 0 {
		cur, amount = q.Fee.CurrencyKey, q.Fee.Amount
	}
	rows, err := s.db.QueryContext(ctx, `SELECT b.user_id::text, u.username, COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), b.level, b.template, b.look::text
		FROM economy_bot b JOIN users u ON u.id=b.user_id
		WHERE b.enabled AND u.disable_time <= '1970-01-01 00:00:01+00' AND b.level >= $1 AND NOT (b.user_id::text = ANY($2::text[]))
		AND (b.rooms = '[]'::jsonb OR b.rooms @> jsonb_build_array(jsonb_build_object('game',$3::text,'room',$4::text)) OR b.rooms @> jsonb_build_array(jsonb_build_object('game',$3::text,'room','')))
		AND ($5 = '' OR COALESCE((u.wallet->>$5)::bigint, 0) >= $6)
		ORDER BY random() LIMIT $7`, q.MinLevel, exclude, q.Game, q.Room, cur, amount, q.Count+5)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []gamecatalog.Bot{}
	for rows.Next() {
		var b gamecatalog.Bot
		var look string
		if err := rows.Scan(&b.UserID, &b.Username, &b.Name, &b.Avatar, &b.Level, &b.Template, &look); err != nil {
			return nil, err
		}
		b.Look = map[string]string{}
		_ = json.Unmarshal([]byte(look), &b.Look)
		if b.Name == "" {
			b.Name = b.Username
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *service) logActions(ctx context.Context, matchID string, actions []gamecatalog.Action) error {
	matchID, err := cleanMatchID(matchID)
	if err != nil {
		return err
	}
	if len(actions) > maxLogActions {
		actions = actions[:maxLogActions]
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		var last int
		err := tx.QueryRowContext(ctx, `UPDATE economy_game_match SET actions=actions+$2 WHERE id=$1 RETURNING actions`, matchID, len(actions)).Scan(&last)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		seq := last - len(actions)
		for _, a := range actions {
			seq++
			var user any
			if _, err := uuid.FromString(a.UserID); err == nil {
				user = a.UserID
			}
			data := "{}"
			if len(a.Data) > 0 {
				raw, _ := json.Marshal(a.Data)
				data = string(raw)
			}
			at := a.At
			if at.IsZero() {
				at = s.now()
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_game_action (match_id, seq, user_id, kind, data, at) VALUES ($1,$2,$3,$4,$5::jsonb,$6) ON CONFLICT DO NOTHING`,
				matchID, seq, user, truncate(a.Kind, 64), data, at.UTC()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *service) sweepGameHistory(ctx context.Context) {
	games, _, err := s.gameRecords(ctx)
	if err != nil {
		s.logger.Warn("Economy game action sweep failed: %v", err)
		return
	}
	exclude, days := []string{}, map[int][]string{}
	for _, g := range games {
		d := gameSettingInt(g, gameActionDaysSetting, gameActionDays)
		if d != gameActionDays {
			exclude = append(exclude, g.Key)
		}
		if d > 0 && d != gameActionDays {
			days[d] = append(days[d], g.Key)
		}
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_game_action WHERE (match_id, seq) IN (SELECT a.match_id, a.seq FROM economy_game_action a
		LEFT JOIN economy_game_match m ON m.id=a.match_id WHERE a.at < $1 AND (m.game_key IS NULL OR NOT (m.game_key = ANY($2::text[]))) LIMIT 5000)`,
		s.now().UTC().AddDate(0, 0, -gameActionDays), exclude); err != nil {
		s.logger.Warn("Economy game action sweep failed: %v", err)
	}
	for d, keys := range days {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_game_action WHERE (match_id, seq) IN (SELECT a.match_id, a.seq FROM economy_game_action a
			JOIN economy_game_match m ON m.id=a.match_id WHERE a.at < $1 AND m.game_key = ANY($2::text[]) LIMIT 5000)`, s.now().UTC().AddDate(0, 0, -d), keys); err != nil {
			s.logger.Warn("Economy game action sweep failed: %v", err)
		}
	}
}

func (p gameProvider) Settle(ctx context.Context, st gamecatalog.Settlement) ([]gamecatalog.Payout, error) {
	r, err := p.s.gameSettle(ctx, st)
	return r, providerError(err)
}

func (p gameProvider) Void(ctx context.Context, matchID, reason, by string) (int, error) {
	n, err := p.s.cancelMatch(ctx, matchID, true, reason, by)
	return n, providerError(err)
}

func (p gameProvider) Release(ctx context.Context, matchID, userID string) error {
	return providerError(p.s.gameRelease(ctx, matchID, userID))
}

func (p gameProvider) Active(ctx context.Context, userID string) (*gamecatalog.ActiveMatch, error) {
	a, err := p.s.gameActive(ctx, userID)
	return a, providerError(err)
}

func (p gameProvider) Bots(ctx context.Context, q gamecatalog.BotQuery) ([]gamecatalog.Bot, error) {
	b, err := p.s.pickBots(ctx, q)
	return b, providerError(err)
}

func (p gameProvider) Log(ctx context.Context, matchID string, actions []gamecatalog.Action) error {
	return providerError(p.s.logActions(ctx, matchID, actions))
}
