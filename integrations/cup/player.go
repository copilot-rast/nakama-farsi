package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

type profile struct {
	Username    string
	DisplayName string
	AvatarURL   string
}

func (s *service) profiles(ctx context.Context, ids []string) map[string]profile {
	out := map[string]profile{}
	list := []string{}
	for _, id := range ids {
		if id != "" {
			if _, ok := out[id]; !ok {
				out[id] = profile{}
				list = append(list, id)
			}
		}
	}
	if len(list) == 0 {
		return out
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, username, COALESCE(display_name,''), COALESCE(avatar_url,'') FROM users WHERE id::text = ANY($1)`, list)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var p profile
		if rows.Scan(&id, &p.Username, &p.DisplayName, &p.AvatarURL) == nil {
			out[id] = p
		}
	}
	return out
}

func feeJSON(r settings.Rules) map[string]any {
	return map[string]any{"currencyKey": r.Entry.Fee.CurrencyKey, "amount": r.Entry.Fee.Amount}
}

func prizesJSON(c *cupRow) map[string]any {
	p := c.Rules.Prizes
	reached := []map[string]any{}
	for _, x := range p.Reached {
		reached = append(reached, map[string]any{"round": x.Round, "prizeKey": x.PrizeKey})
	}
	pool := c.Pool
	if c.Status != "running" && c.Status != "finished" {
		pool = poolValue{CurrencyKey: c.Rules.Entry.Fee.CurrencyKey, Fees: c.Rules.Entry.Fee.Amount * int64(c.Players)}
		pool.Amount = poolShare(pool.Fees, p.Pool.Percent) + p.Pool.Base
		if pool.CurrencyKey == "" {
			pool.Amount = 0
		}
	}
	sh := p.Pool.Shares
	return map[string]any{"champion": p.Champion, "runnerUp": p.RunnerUp, "semi": p.Semi, "third": p.Third, "fourth": p.Fourth, "reached": reached,
		"pool": map[string]any{"currencyKey": pool.CurrencyKey, "amount": pool.Amount, "percent": p.Pool.Percent,
			"shares": map[string]any{"champion": sh.Champion, "runnerUp": sh.RunnerUp, "semi": sh.Semi, "third": sh.Third, "fourth": sh.Fourth}}}
}

func requirementsJSON(r settings.Requirements) map[string]any {
	return map[string]any{"minTrophies": r.MinTrophies, "maxTrophies": r.MaxTrophies, "leagues": r.Leagues, "tagIds": r.TagIDs, "tagsMode": r.TagsMode, "excludeTagIds": r.ExcludeTagIDs}
}

func (s *service) summary(c *cupRow, tx texts, me *playerRow, now time.Time) map[string]any {
	r := c.Rules
	t := r.Timing
	mine := map[string]any{"registered": false, "status": "", "seed": 0, "place": 0, "reached": 0, "canSignUp": false, "canWithdraw": false}
	if me != nil {
		active := me.Status == "registered" || me.Status == "playing"
		mine = map[string]any{"registered": active || (me.Seed > 0), "status": me.Status, "seed": me.Seed, "place": me.Place, "reached": me.Reached,
			"canSignUp": c.Status == "signup" && !active && c.Players < r.Capacity, "canWithdraw": me.Status == "registered" && (c.Status == "signup" || c.Status == "upcoming")}
	} else {
		mine["canSignUp"] = c.Status == "signup" && c.Players < r.Capacity
	}
	return map[string]any{
		"id": c.ID, "templateKey": c.TemplateKey, "name": tx.of(c.Name, c.TemplateKey), "description": tx.of(c.Description, ""), "iconUrl": c.IconURL, "status": c.Status,
		"game": r.Game, "mode": t.Mode, "capacity": r.Capacity, "minPlayers": r.MinPlayers, "players": c.Players, "size": c.Size, "rounds": c.Rounds, "currentRound": c.CurrentRound,
		"signupAt": isoT(c.SignupAt), "startsAt": isoT(c.StartsAt), "startedAt": iso(c.StartedAt), "finishedAt": iso(c.FinishedAt), "championId": c.Champion,
		"cancelReason": c.CancelReason, "bestOf": r.BestOf, "thirdPlace": r.ThirdPlace, "seeding": r.Seeding, "rated": r.Rated,
		"timing": map[string]any{"mode": t.Mode, "windowMinutes": t.WindowMinutes, "roundMinutes": t.RoundMinutes, "checkInMinutes": t.CheckInMinutes, "readySeconds": t.ReadySeconds},
		"fee": feeJSON(r), "withdrawRefundPercent": r.Entry.WithdrawRefundPercent, "requirements": requirementsJSON(r.Entry.Requirements), "prizes": prizesJSON(c), "me": mine,
		"serverTime": isoT(now),
	}
}

func sideJSON(b *bracket, m *matchRow, user string, seed, wins int, ready *time.Time, showed bool, profiles map[string]profile) map[string]any {
	if user == "" {
		return nil
	}
	p := profiles[user]
	status := ""
	if pl := b.players[user]; pl != nil {
		status = pl.Status
	}
	return map[string]any{"userId": user, "username": p.Username, "displayName": p.DisplayName, "avatarUrl": p.AvatarURL, "seed": seed, "wins": wins,
		"ready": b.fresh(ready), "showed": showed, "status": status}
}

func (b *bracket) matchJSON(m *matchRow, profiles map[string]profile) map[string]any {
	games := []map[string]any{}
	for _, g := range b.games[m.ID] {
		games = append(games, map[string]any{"no": g.GameNo, "status": g.Status, "winnerId": g.Winner, "draw": g.Draw, "score": g.Score, "reason": g.Reason,
			"createdAt": isoT(g.CreatedAt), "finishedAt": iso(g.FinishedAt)})
	}
	var next map[string]any
	if m.Kind == kindMain && m.Round < b.cup.Rounds {
		r, slot, side := nextSlot(m.Round, m.Slot)
		if n := b.match(kindMain, r, slot); n != nil {
			next = map[string]any{"matchId": n.ID, "side": side}
		}
	}
	return map[string]any{"id": m.ID, "round": m.Round, "slot": m.Slot, "kind": m.Kind, "roundKey": roundKey(m.Round, b.cup.Rounds), "status": m.Status, "bestOf": m.BestOf,
		"a": sideJSON(b, m, m.A, m.SeedA, m.WinsA, m.ReadyA, m.ShowedA, profiles), "b": sideJSON(b, m, m.B, m.SeedB, m.WinsB, m.ReadyB, m.ShowedB, profiles),
		"winnerId": m.Winner, "loserId": m.Loser, "reason": m.Reason, "draws": m.Draws, "gameNo": m.Games, "gameMatchId": m.GameMatchID, "games": games,
		"opensAt": iso(m.OpensAt), "startsAt": iso(m.StartsAt), "closesAt": iso(m.ClosesAt), "finishedAt": iso(m.FinishedAt), "next": next}
}

func (b *bracket) userIDs() []string {
	ids := []string{}
	for _, p := range b.order {
		ids = append(ids, p.UserID)
	}
	return ids
}

func (b *bracket) view(profiles map[string]profile) map[string]any {
	c := b.cup
	rounds := []map[string]any{}
	for r := 1; r <= c.Rounds; r++ {
		list := []map[string]any{}
		status := "pending"
		open, finished, total := 0, 0, 0
		for _, m := range b.matches {
			if m.Kind != kindMain || m.Round != r {
				continue
			}
			list = append(list, b.matchJSON(m, profiles))
			total++
			if done(m.Status) {
				finished++
			} else if m.Status != stPending {
				open++
			}
		}
		switch {
		case total > 0 && finished == total:
			status = "finished"
		case open > 0 || finished > 0:
			status = "running"
		}
		rounds = append(rounds, map[string]any{"round": r, "key": roundKey(r, c.Rounds), "bestOf": c.Rules.BestOfFor(r, c.Rounds, false), "status": status,
			"players": c.Size >> (r - 1), "matches": list})
	}
	var third map[string]any
	if t := b.third(); t != nil {
		third = b.matchJSON(t, profiles)
	}
	return map[string]any{"size": c.Size, "rounds": rounds, "third": third}
}

func (s *service) eligible(ctx context.Context, r settings.Requirements, userID string) error {
	trophies, league := 0, ""
	var exists bool
	_ = s.db.QueryRowContext(ctx, `SELECT to_regclass('trophy_player') IS NOT NULL`).Scan(&exists)
	if exists {
		_ = s.db.QueryRowContext(ctx, `SELECT trophies, league_key FROM trophy_player WHERE user_id::text=$1`, userID).Scan(&trophies, &league)
	}
	if r.MinTrophies > 0 && trophies < r.MinTrophies {
		return apierr.Forbidden("CUP_TROPHIES_TOO_LOW").With("min", r.MinTrophies).With("trophies", trophies)
	}
	if r.MaxTrophies > 0 && trophies > r.MaxTrophies {
		return apierr.Forbidden("CUP_TROPHIES_TOO_HIGH").With("max", r.MaxTrophies).With("trophies", trophies)
	}
	if len(r.Leagues) > 0 {
		hit := false
		for _, l := range r.Leagues {
			hit = hit || l == league
		}
		if !hit {
			return apierr.Forbidden("CUP_LEAGUE_REQUIRED").With("leagues", r.Leagues)
		}
	}
	if len(r.TagIDs) == 0 && len(r.ExcludeTagIDs) == 0 {
		return nil
	}
	own := map[string]bool{}
	rows, err := s.db.QueryContext(ctx, `SELECT tag_id::text FROM base_user_tag WHERE user_id::text=$1 AND (expires_at IS NULL OR expires_at > now())`, userID)
	if err != nil {
		rows, err = s.db.QueryContext(ctx, `SELECT tag_id::text FROM base_user_tag WHERE user_id::text=$1`, userID)
	}
	if err == nil {
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil {
				own[id] = true
			}
		}
		rows.Close()
	}
	for _, id := range r.ExcludeTagIDs {
		if own[id] {
			return apierr.Forbidden("CUP_TAG_EXCLUDED")
		}
	}
	if len(r.TagIDs) > 0 {
		n := 0
		for _, id := range r.TagIDs {
			if own[id] {
				n++
			}
		}
		if (r.TagsMode == settings.TagsAll && n < len(r.TagIDs)) || n == 0 {
			return apierr.Forbidden("CUP_TAG_REQUIRED").With("tagIds", r.TagIDs)
		}
	}
	return nil
}

func cupIDOf(payload string) (string, error) {
	var in struct {
		CupID string `json:"cupId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return "", err
	}
	if !uuidPattern.MatchString(in.CupID) {
		return "", apierr.BadRequest("CUP_ID_REQUIRED").WithField("cupId")
	}
	return in.CupID, nil
}

func (s *service) rpcSignup(ctx context.Context, userID, payload string) (any, error) {
	cupID, err := cupIDOf(payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if !cfg.Enabled {
		return nil, apierr.Unavailable("CUP_DISABLED")
	}
	var charged *feeValue
	b, err := s.withCup(ctx, cupID, false, func(tx *sql.Tx, b *bracket) error {
		c := b.cup
		switch c.Status {
		case "signup":
		case "upcoming":
			return apierr.Conflict("CUP_SIGNUP_NOT_OPEN").With("signupAt", isoT(c.SignupAt))
		case "draft":
			return apierr.NotFound("CUP_NOT_FOUND")
		default:
			return apierr.Conflict("CUP_SIGNUP_CLOSED")
		}
		if p := b.players[userID]; p != nil && p.Status == "registered" {
			return apierr.Conflict("CUP_ALREADY_SIGNED_UP")
		} else if p != nil && (p.Status == "removed" || p.Status == "disqualified") {
			return apierr.Forbidden("CUP_REMOVED")
		}
		n := 0
		for _, p := range b.order {
			if p.Status == "registered" {
				n++
			}
		}
		if n >= c.Rules.Capacity {
			return apierr.Conflict("CUP_FULL").With("capacity", c.Rules.Capacity)
		}
		if lim := cfg.Limits.ActivePerPlayer; lim > 0 {
			var active int
			_ = tx.QueryRowContext(ctx, `SELECT count(*) FROM cup_player p JOIN cup c ON c.id=p.cup_id WHERE p.user_id::text=$1 AND p.status IN ('registered','playing')
				AND c.status IN ('upcoming','signup','running') AND c.id::text<>$2`, userID, c.ID).Scan(&active)
			if active >= lim {
				return apierr.Conflict("CUP_TOO_MANY_ACTIVE").With("max", lim)
			}
		}
		if err := s.eligible(ctx, c.Rules.Entry.Requirements, userID); err != nil {
			return err
		}
		fee := feeValue{CurrencyKey: c.Rules.Entry.Fee.CurrencyKey, Amount: c.Rules.Entry.Fee.Amount}
		state := "none"
		if fee.Amount > 0 && fee.CurrencyKey != "" {
			fee.Grant = "cup:fee:" + c.ID + ":" + userID + ":" + strconv.FormatInt(b.now.UnixNano(), 36)
			err := s.debit(ctx, userID, fee.Grant, map[string]int64{fee.CurrencyKey: fee.Amount}, map[string]any{"module": owner, "sourceType": sourceType, "sourceId": c.ID, "reason": "cup_entry"})
			var short *pluginkit.ErrInsufficient
			if errors.As(err, &short) {
				return apierr.Conflict("CUP_NOT_ENOUGH_CURRENCY").With("currencyKey", fee.CurrencyKey).With("amount", fee.Amount)
			}
			if err != nil {
				return err
			}
			state = "paid"
			charged = &fee
		} else {
			fee = feeValue{}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO cup_player (cup_id, user_id, status, fee, fee_state, joined_at) VALUES ($1,$2,'registered',$3,$4,$5)
			ON CONFLICT (cup_id, user_id) DO UPDATE SET status='registered', fee=EXCLUDED.fee, fee_state=EXCLUDED.fee_state, joined_at=EXCLUDED.joined_at, seed=0, place=0, reached=0, updated_at=now()`,
			c.ID, userID, jsonOf(fee), state, b.now); err != nil {
			return err
		}
		c.Players = n + 1
		p := &playerRow{UserID: userID, Status: "registered", Fee: fee, FeeState: state, JoinedAt: b.now}
		b.players[userID] = p
		b.order = append(b.order, p)
		return nil
	})
	if err != nil {
		if charged != nil {
			_ = s.refundPlayer(ctx, cupID, &playerRow{UserID: userID, Fee: *charged, FeeState: "paid"}, 100)
		}
		return nil, err
	}
	tx := s.texts(ctx, userID, apierr.PayloadLang(payload))
	return map[string]any{"cup": s.summary(b.cup, tx, b.players[userID], b.now)}, nil
}

func (s *service) rpcWithdraw(ctx context.Context, userID, payload string) (any, error) {
	cupID, err := cupIDOf(payload)
	if err != nil {
		return nil, err
	}
	var me *playerRow
	b, err := s.withCup(ctx, cupID, false, func(tx *sql.Tx, b *bracket) error {
		c := b.cup
		p := b.players[userID]
		if p == nil || p.Status != "registered" {
			return apierr.Conflict("CUP_NOT_SIGNED_UP")
		}
		if c.Status != "signup" && c.Status != "upcoming" {
			return apierr.Conflict("CUP_WITHDRAW_CLOSED")
		}
		p.Status = "withdrawn"
		b.touchPlayer(userID)
		if c.Players > 0 {
			c.Players--
		}
		me = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	percent := b.cup.Rules.Entry.WithdrawRefundPercent
	refunded := int64(0)
	if me.FeeState == "paid" {
		if err := s.refundPlayer(ctx, cupID, me, percent); err != nil {
			s.logger.Warn("Cup %s withdraw refund for %s failed, retried later: %v", cupID, userID, err)
		} else {
			refunded = me.Fee.Refunded
		}
	}
	tx := s.texts(ctx, userID, apierr.PayloadLang(payload))
	return map[string]any{"cup": s.summary(b.cup, tx, me, b.now), "refunded": map[string]any{"currencyKey": me.Fee.CurrencyKey, "amount": refunded, "percent": percent}}, nil
}

func (s *service) readBracket(ctx context.Context, cupID string) (*bracket, error) {
	c, err := loadCup(ctx, s.db, cupID, "")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("CUP_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return s.loadBracket(ctx, s.db, c)
}

func (s *service) rpcGet(ctx context.Context, userID, payload string) (any, error) {
	cupID, err := cupIDOf(payload)
	if err != nil {
		return nil, err
	}
	b, err := s.readBracket(ctx, cupID)
	if err != nil {
		return nil, err
	}
	if b.cup.Status == "draft" {
		return nil, apierr.NotFound("CUP_NOT_FOUND")
	}
	tx := s.texts(ctx, userID, apierr.PayloadLang(payload))
	profiles := s.profiles(ctx, b.userIDs())
	var current map[string]any
	if m := b.activeMatchOf(userID); m != nil {
		current = b.matchJSON(m, profiles)
	}
	var champion map[string]any
	if b.cup.Champion != "" {
		p := profiles[b.cup.Champion]
		champion = map[string]any{"userId": b.cup.Champion, "username": p.Username, "displayName": p.DisplayName, "avatarUrl": p.AvatarURL}
	}
	standings := []map[string]any{}
	if b.cup.Status == "finished" {
		for _, p := range b.order {
			if p.Seed > 0 && p.Place > 0 && p.Place <= 4 {
				pr := profiles[p.UserID]
				standings = append(standings, map[string]any{"userId": p.UserID, "username": pr.Username, "displayName": pr.DisplayName, "place": p.Place, "seed": p.Seed})
			}
		}
	}
	return map[string]any{"cup": s.summary(b.cup, tx, b.players[userID], b.now), "bracket": b.view(profiles), "myMatch": current, "champion": champion, "standings": standings}, nil
}

func (s *service) rpcList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Status string `json:"status"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
		Lang   string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	limit := in.Limit
	if limit <= 0 || limit > cfg.Display.ListLimit {
		limit = cfg.Display.ListLimit
	}
	if in.Offset < 0 || in.Offset > 10000 {
		in.Offset = 0
	}
	q := `SELECT ` + cupCols + ` FROM cup WHERE (announce_at IS NULL OR announce_at <= now())`
	args := []any{}
	switch in.Status {
	case "upcoming", "signup", "running":
		q += ` AND status='` + in.Status + `' ORDER BY starts_at`
	case "finished":
		q += ` AND status IN ('finished','cancelled') AND COALESCE(finished_at, starts_at) > now() - make_interval(days => $1) ORDER BY COALESCE(finished_at, starts_at) DESC`
		args = append(args, cfg.Display.FinishedDays)
	case "mine":
		q += ` AND status <> 'draft' AND EXISTS (SELECT 1 FROM cup_player p WHERE p.cup_id=cup.id AND p.user_id::text=$1 AND p.status NOT IN ('withdrawn','removed')) ORDER BY starts_at DESC`
		args = append(args, userID)
	case "":
		q += ` AND (status IN ('upcoming','signup','running') OR (status IN ('finished','cancelled') AND COALESCE(finished_at, starts_at) > now() - make_interval(days => $1)))
			ORDER BY CASE status WHEN 'running' THEN 0 WHEN 'signup' THEN 1 WHEN 'upcoming' THEN 2 ELSE 3 END, CASE WHEN status IN ('finished','cancelled') THEN NULL ELSE starts_at END, starts_at DESC`
		args = append(args, cfg.Display.FinishedDays)
	default:
		return nil, apierr.BadRequest("CUP_BAD_STATUS").WithField("status")
	}
	args = append(args, limit+1, in.Offset)
	q += ` LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	cups := []*cupRow{}
	for rows.Next() {
		c, err := scanCup(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		cups = append(cups, c)
	}
	rows.Close()
	more := len(cups) > limit
	if more {
		cups = cups[:limit]
	}
	mine := map[string]*playerRow{}
	if len(cups) > 0 {
		ids := []string{}
		for _, c := range cups {
			ids = append(ids, c.ID)
		}
		prow, err := s.db.QueryContext(ctx, `SELECT cup_id::text, `+playerCols+` FROM cup_player WHERE user_id::text=$1 AND cup_id::text = ANY($2)`, userID, ids)
		if err == nil {
			for prow.Next() {
				var cid, fee string
				p := &playerRow{}
				if prow.Scan(&cid, &p.UserID, &p.Status, &p.Seed, &p.Trophies, &p.Place, &p.Reached, &fee, &p.FeeState, &p.AddedBy, &p.JoinedAt) == nil {
					p.Fee = feeFrom(fee)
					mine[cid] = p
				}
			}
			prow.Close()
		}
	}
	tx := s.texts(ctx, userID, in.Lang)
	now := s.now()
	items := []map[string]any{}
	for _, c := range cups {
		items = append(items, s.summary(c, tx, mine[c.ID], now))
	}
	return map[string]any{"items": items, "more": more, "offset": in.Offset}, nil
}

func (s *service) rpcMyMatch(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		CupID string `json:"cupId"`
		Lang  string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	ids := []string{}
	if in.CupID != "" {
		if !uuidPattern.MatchString(in.CupID) {
			return nil, apierr.BadRequest("CUP_ID_REQUIRED").WithField("cupId")
		}
		ids = append(ids, in.CupID)
	} else {
		ids = s.ids(ctx, `SELECT c.id::text FROM cup c JOIN cup_player p ON p.cup_id=c.id WHERE p.user_id::text=$1 AND c.status='running' AND p.status='playing' ORDER BY c.starts_at`, userID)
	}
	tx := s.texts(ctx, userID, in.Lang)
	items := []map[string]any{}
	for _, id := range ids {
		b, err := s.readBracket(ctx, id)
		if err != nil {
			if in.CupID != "" {
				return nil, err
			}
			continue
		}
		p := b.players[userID]
		var match map[string]any
		action := "none"
		if m := b.activeMatchOf(userID); m != nil {
			match = b.matchJSON(m, s.profiles(ctx, []string{m.A, m.B}))
			mine := (userID == m.A && b.fresh(m.ReadyA)) || (userID == m.B && b.fresh(m.ReadyB))
			switch {
			case m.Status == stWaiting:
				action = "wait_check_in"
			case m.Status == stPlaying:
				action = "play"
			case b.cup.Rules.Timing.Mode == settings.ModeLive && m.Games == 0 && ((userID == m.A && m.ShowedA) || (userID == m.B && m.ShowedB)):
				action = "wait_start"
			case b.cup.Rules.Timing.Mode == settings.ModeLive && m.Games == 0:
				action = "check_in"
			case mine:
				action = "wait_opponent"
			default:
				action = "ready"
			}
		} else if p != nil && p.Status == "playing" && b.cup.Status == "running" {
			action = "wait_round"
		}
		status := ""
		if p != nil {
			status = p.Status
		}
		items = append(items, map[string]any{"cupId": b.cup.ID, "cupName": tx.of(b.cup.Name, b.cup.TemplateKey), "cupStatus": b.cup.Status, "playerStatus": status,
			"action": action, "match": match, "serverTime": isoT(b.now)})
	}
	return map[string]any{"items": items}, nil
}

func (s *service) rpcHistory(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit int    `json:"limit"`
		Lang  string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if in.Limit <= 0 || in.Limit > cfg.Display.HistoryLimit {
		in.Limit = cfg.Display.HistoryLimit
	}
	rows, err := s.db.QueryContext(ctx, `SELECT c.id::text, c.name::text, c.template_key, c.status, c.players, c.starts_at, c.finished_at, p.status, p.place, p.reached, p.seed,
		COALESCE((SELECT json_agg(json_build_object('kind', x.kind, 'prizeKey', x.prize_key, 'amounts', x.amounts, 'status', x.status) ORDER BY x.created_at)::text FROM cup_payout x WHERE x.cup_id=c.id AND x.user_id=p.user_id), '[]')
		FROM cup_player p JOIN cup c ON c.id=p.cup_id WHERE p.user_id::text=$1 AND p.seed > 0 ORDER BY c.starts_at DESC LIMIT $2`, userID, in.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tx := s.texts(ctx, userID, in.Lang)
	items := []map[string]any{}
	for rows.Next() {
		var id, name, key, status, pstatus, prizes string
		var players, place, reached, seed int
		var starts time.Time
		var finished *time.Time
		if err := rows.Scan(&id, &name, &key, &status, &players, &starts, &finished, &pstatus, &place, &reached, &seed, &prizes); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"cupId": id, "name": tx.of(textMap(name), key), "status": status, "players": players, "startsAt": isoT(starts), "finishedAt": iso(finished),
			"playerStatus": pstatus, "place": place, "reached": reached, "seed": seed, "payouts": payoutList(prizes)})
	}
	return map[string]any{"items": items}, rows.Err()
}
