package main

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/cuphooks"
)

const (
	stPending  = "pending"
	stWaiting  = "waiting"
	stOpen     = "open"
	stPlaying  = "playing"
	stFinished = "finished"
	stBye      = "bye"
	stVoid     = "void"

	kindMain  = "main"
	kindThird = "third"
)

type note struct {
	users      []string
	subject    string
	content    map[string]any
	persistent bool
	inapp      bool
}

type launchJob struct {
	gameID  string
	matchID string
	launch  cuphooks.Launch
}

type effects struct {
	notes    []note
	launches []launchJob
	events   []cupEvent
	finished bool
}

type cupEvent struct {
	userID string
	name   string
	fields map[string]any
}

type bracket struct {
	s       *service
	cup     *cupRow
	players map[string]*playerRow
	order   []*playerRow
	matches []*matchRow
	byKey   map[string]*matchRow
	games   map[string][]*gameRow
	now     time.Time
	fx      *effects
	touched map[string]bool
	newGame []*gameRow
}

func matchKey(kind string, round, slot int) string {
	return kind + ":" + strconv.Itoa(round) + ":" + strconv.Itoa(slot)
}

func (s *service) loadBracket(ctx context.Context, tx queryer, c *cupRow) (*bracket, error) {
	b := &bracket{s: s, cup: c, players: map[string]*playerRow{}, byKey: map[string]*matchRow{}, now: s.now().UTC(), fx: &effects{}, touched: map[string]bool{}}
	list, err := loadPlayers(ctx, tx, c.ID)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		b.players[p.UserID] = p
	}
	b.order = list
	if b.matches, err = loadMatches(ctx, tx, c.ID); err != nil {
		return nil, err
	}
	for _, m := range b.matches {
		b.byKey[matchKey(m.Kind, m.Round, m.Slot)] = m
	}
	if b.games, err = loadGames(ctx, tx, c.ID); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *bracket) match(kind string, round, slot int) *matchRow { return b.byKey[matchKey(kind, round, slot)] }

func (b *bracket) matchByID(id string) *matchRow {
	for _, m := range b.matches {
		if m.ID == id {
			return m
		}
	}
	return nil
}

func (b *bracket) third() *matchRow { return b.match(kindThird, b.cup.Rounds, 0) }

func (b *bracket) touch(m *matchRow) { m.dirty = true }

func (b *bracket) touchPlayer(id string) { b.touched[id] = true }

func done(status string) bool { return status == stFinished || status == stBye || status == stVoid }

func (b *bracket) roundDone(round int) bool {
	found := false
	for _, m := range b.matches {
		if m.Kind == kindMain && m.Round == round {
			found = true
			if !done(m.Status) {
				return false
			}
		}
	}
	return found
}

func (b *bracket) notify(users []string, subject string, content map[string]any, persistent, inapp bool) {
	list := []string{}
	for _, u := range users {
		if u != "" {
			list = append(list, u)
		}
	}
	if len(list) == 0 {
		return
	}
	c := map[string]any{"cupId": b.cup.ID}
	for k, v := range content {
		c[k] = v
	}
	b.fx.notes = append(b.fx.notes, note{users: list, subject: subject, content: c, persistent: persistent, inapp: inapp})
}

func (b *bracket) event(userID, name string, fields map[string]any) {
	f := map[string]any{"cupId": b.cup.ID}
	for k, v := range fields {
		f[k] = v
	}
	b.fx.events = append(b.fx.events, cupEvent{userID: userID, name: name, fields: f})
}

func (b *bracket) seedOf(m *matchRow, user string) int {
	if user == m.A {
		return m.SeedA
	}
	return m.SeedB
}

func (b *bracket) other(m *matchRow, user string) string {
	if user == m.A {
		return m.B
	}
	return m.A
}

func (b *bracket) roundStart(round int) time.Time {
	t := b.cup.Rules.Timing
	return b.cup.StartsAt.Add(time.Duration(t.CheckInMinutes+(round-1)*t.RoundMinutes) * time.Minute)
}

func (b *bracket) build(rnd func(int) int) error {
	c := b.cup
	list := []entrant{}
	for _, p := range b.order {
		if p.Status == "registered" {
			list = append(list, entrant{UserID: p.UserID, Trophies: p.Trophies, Joined: p.JoinedAt.UnixNano()})
		}
	}
	seeded := seedEntrants(list, c.Rules.Seeding, rnd)
	c.Size = bracketSize(len(seeded))
	c.Rounds = roundCount(c.Size)
	c.Players = len(seeded)
	bySeed := map[int]string{}
	for i, e := range seeded {
		p := b.players[e.UserID]
		p.Seed, p.Status, p.Reached = i+1, "playing", 1
		bySeed[i+1] = e.UserID
		b.touchPlayer(e.UserID)
	}
	add := func(kind string, round, slot int) *matchRow {
		m := &matchRow{ID: uuid.Must(uuid.NewV4()).String(), Kind: kind, Round: round, Slot: slot, Status: stPending,
			BestOf: c.Rules.BestOfFor(round, c.Rounds, kind == kindThird), dirty: true}
		b.matches = append(b.matches, m)
		b.byKey[matchKey(kind, round, slot)] = m
		b.newMatch(m)
		return m
	}
	for r := 1; r <= c.Rounds; r++ {
		for i := 0; i < c.Size>>r; i++ {
			add(kindMain, r, i)
		}
	}
	if c.Rules.ThirdPlace && c.Rounds >= 2 {
		add(kindThird, c.Rounds, 0)
	}
	for _, p := range firstRound(c.Size, c.Players) {
		m := b.match(kindMain, 1, p.Slot)
		m.A, m.SeedA = bySeed[p.SeedA], p.SeedA
		if p.SeedB > 0 {
			m.B, m.SeedB = bySeed[p.SeedB], p.SeedB
		}
	}
	for _, m := range b.matches {
		if m.Kind == kindMain && m.Round == 1 {
			if m.B == "" {
				m.Status, m.Reason = stBye, "bye"
				b.place(m, m.A)
				m.Winner, m.FinishedAt = m.A, ptr(b.now)
				b.notify([]string{m.A}, "cup_advanced", map[string]any{"matchId": m.ID, "round": 1, "reason": "bye", "nextRound": 2}, true, false)
				continue
			}
			b.open(m)
		}
	}
	c.CurrentRound = 1
	return nil
}

func (b *bracket) newMatch(m *matchRow) {
	m.dirty = true
	m.Actor = "new"
}

func (b *bracket) place(m *matchRow, winner string) {
	c := b.cup
	if m.Kind != kindMain || m.Round >= c.Rounds {
		return
	}
	r, slot, side := nextSlot(m.Round, m.Slot)
	next := b.match(kindMain, r, slot)
	if next == nil {
		return
	}
	seed := b.seedOf(m, winner)
	if side == "a" {
		next.A, next.SeedA = winner, seed
	} else {
		next.B, next.SeedB = winner, seed
	}
	b.touch(next)
	if p := b.players[winner]; p != nil && p.Reached < r {
		p.Reached = r
		b.touchPlayer(winner)
	}
}

func (b *bracket) open(m *matchRow) {
	t := b.cup.Rules.Timing
	now := b.now
	m.ReadyA, m.ReadyB, m.ShowedA, m.ShowedB = nil, nil, false, false
	content := map[string]any{"matchId": m.ID, "round": m.Round, "kind": m.Kind, "roundKey": roundKey(m.Round, b.cup.Rounds), "bestOf": m.BestOf, "mode": t.Mode}
	if t.Mode == settings.ModeLive {
		start := b.roundStart(m.Round)
		if min := now.Add(time.Duration(t.CheckInMinutes) * time.Minute); start.Before(min) {
			start = min
		}
		opens := start.Add(-time.Duration(t.CheckInMinutes) * time.Minute)
		m.OpensAt, m.StartsAt, m.ClosesAt = ptr(opens), ptr(start), ptr(start.Add(time.Duration(t.RoundMinutes)*time.Minute))
		m.Status = stWaiting
		if !now.Before(opens) {
			m.Status = stOpen
		}
	} else {
		m.OpensAt, m.StartsAt, m.ClosesAt = ptr(now), ptr(now), ptr(now.Add(time.Duration(t.WindowMinutes)*time.Minute))
		m.Status = stOpen
	}
	b.touch(m)
	content["opensAt"], content["startsAt"], content["closesAt"] = iso(m.OpensAt), iso(m.StartsAt), iso(m.ClosesAt)
	content["checkIn"] = t.Mode == settings.ModeLive && m.Status == stOpen
	for _, u := range []string{m.A, m.B} {
		c := map[string]any{"opponentId": b.other(m, u)}
		for k, v := range content {
			c[k] = v
		}
		b.notify([]string{u}, "cup_round_started", c, true, t.Mode == settings.ModeLive && m.Status == stOpen)
	}
	if m.Round > b.cup.CurrentRound {
		b.cup.CurrentRound = m.Round
	}
}

func (b *bracket) finish(m *matchRow, winner, reason string) {
	c := b.cup
	loser := b.other(m, winner)
	m.Winner, m.Loser, m.Reason, m.Status, m.FinishedAt = winner, loser, reason, stFinished, ptr(b.now)
	m.ReadyA, m.ReadyB = nil, nil
	b.touch(m)
	third := b.third()
	isFinal := m.Kind == kindMain && m.Round == c.Rounds
	thirdFeeds := third != nil && m.Kind == kindMain && m.Round == c.Rounds-1
	info := map[string]any{"matchId": m.ID, "round": m.Round, "kind": m.Kind, "reason": reason, "winsA": m.WinsA, "winsB": m.WinsB, "winnerId": winner}
	switch {
	case m.Kind == kindThird:
		b.setPlace(winner, 3, "finished")
		b.setPlace(loser, 4, "eliminated")
		b.notify([]string{winner}, "cup_advanced", merge(info, map[string]any{"place": 3}), true, false)
		b.notify([]string{loser}, "cup_eliminated", merge(info, map[string]any{"place": 4}), true, false)
	case isFinal:
		b.setPlace(winner, 1, "finished")
		b.setPlace(loser, 2, "finished")
		b.notify([]string{loser}, "cup_eliminated", merge(info, map[string]any{"place": 2}), true, false)
	default:
		b.place(m, winner)
		b.notify([]string{winner}, "cup_advanced", merge(info, map[string]any{"nextRound": m.Round + 1}), true, false)
		if thirdFeeds {
			if m.Slot == 0 {
				third.A, third.SeedA = loser, b.seedOf(m, loser)
			} else {
				third.B, third.SeedB = loser, b.seedOf(m, loser)
			}
			b.touch(third)
			b.notify([]string{loser}, "cup_eliminated", merge(info, map[string]any{"thirdPlace": true}), true, false)
		} else {
			place := placeOf(m.Round, c.Rounds)
			b.setPlace(loser, place, "eliminated")
			b.notify([]string{loser}, "cup_eliminated", merge(info, map[string]any{"place": place}), true, false)
		}
	}
	for _, u := range []string{winner, loser} {
		if u != "" {
			b.event(u, "cup_match", map[string]any{"matchId": m.ID, "round": m.Round, "won": u == winner, "reason": reason, "final": isFinal})
		}
	}
}

func (b *bracket) voidMatch(m *matchRow, reason string) {
	m.Status, m.Reason, m.FinishedAt, m.Winner, m.Loser = stVoid, reason, ptr(b.now), "", ""
	m.ReadyA, m.ReadyB = nil, nil
	b.touch(m)
	for _, u := range []string{m.A, m.B} {
		if u == "" {
			continue
		}
		if p := b.players[u]; p != nil && p.Status == "disqualified" {
			continue
		}
		b.setPlace(u, 0, "eliminated")
		b.notify([]string{u}, "cup_eliminated", map[string]any{"matchId": m.ID, "round": m.Round, "kind": m.Kind, "reason": reason, "place": 0}, true, false)
	}
}

func (b *bracket) setPlace(user string, place int, status string) {
	p := b.players[user]
	if p == nil {
		return
	}
	if p.Status != "disqualified" {
		p.Status = status
	}
	p.Place = place
	b.touchPlayer(user)
}

func merge(a, c map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range c {
		out[k] = v
	}
	return out
}

func (b *bracket) disqualified(user string) bool {
	p := b.players[user]
	return p != nil && p.Status == "disqualified"
}

func (b *bracket) resolveOpen(m *matchRow) {
	dqA, dqB := m.A != "" && b.disqualified(m.A), m.B != "" && b.disqualified(m.B)
	switch {
	case m.A == "" && m.B == "":
		b.voidMatch(m, "empty")
	case m.A == "" || m.B == "":
		w := m.A
		if w == "" {
			w = m.B
		}
		if b.disqualified(w) {
			b.voidMatch(m, "disqualified")
			return
		}
		b.finish(m, w, "walkover")
	case dqA && dqB:
		b.voidMatch(m, "disqualified")
	case dqA:
		b.finish(m, m.B, "disqualified")
	case dqB:
		b.finish(m, m.A, "disqualified")
	default:
		b.open(m)
	}
}

func (b *bracket) advance() {
	c := b.cup
	for changed := true; changed; {
		changed = false
		for r := 2; r <= c.Rounds; r++ {
			if !b.roundDone(r - 1) {
				break
			}
			for _, m := range b.matches {
				if m.Kind == kindMain && m.Round == r && m.Status == stPending {
					b.resolveOpen(m)
					changed = true
				}
			}
		}
		if t := b.third(); t != nil && t.Status == stPending && b.roundDone(c.Rounds-1) {
			b.resolveOpen(t)
			changed = true
		}
	}
	final := b.match(kindMain, c.Rounds, 0)
	if c.Status == "running" && final != nil && done(final.Status) {
		if t := b.third(); t == nil || done(t.Status) {
			b.finishCup(final)
		}
	}
}

func (b *bracket) pick(m *matchRow, rule string, rnd func(int) int) string {
	switch rule {
	case settings.NoShowHigherSeed:
		if m.SeedB > 0 && (m.SeedA == 0 || m.SeedB < m.SeedA) {
			return m.B
		}
		return m.A
	case settings.NoShowRandom:
		if rnd(2) == 1 {
			return m.B
		}
		return m.A
	}
	return ""
}

func (b *bracket) decide(m *matchRow, rnd func(int) int) {
	t := b.cup.Rules.Timing
	switch {
	case m.WinsA > m.WinsB:
		b.finish(m, m.A, "deadline")
	case m.WinsB > m.WinsA:
		b.finish(m, m.B, "deadline")
	case m.ShowedA && !m.ShowedB:
		b.finish(m, m.A, "no_show")
	case m.ShowedB && !m.ShowedA:
		b.finish(m, m.B, "no_show")
	default:
		rule, reason := t.NoShow, "double_no_show"
		if m.ShowedA && m.ShowedB {
			rule, reason = t.Unplayed, "unplayed"
		}
		if w := b.pick(m, rule, rnd); w != "" {
			b.finish(m, w, reason+":"+rule)
			return
		}
		b.voidMatch(m, reason)
	}
}

func (b *bracket) currentGame(m *matchRow) *gameRow {
	list := b.games[m.ID]
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].GameNo == m.Games {
			return list[i]
		}
	}
	return nil
}

func (b *bracket) launch(m *matchRow) {
	c := b.cup
	m.Games++
	m.Status, m.ReadyA, m.ReadyB, m.GameMatchID = stPlaying, nil, nil, ""
	b.touch(m)
	g := &gameRow{ID: uuid.Must(uuid.NewV4()).String(), MatchID: m.ID, CupID: c.ID, GameNo: m.Games, Status: "launching", Score: map[string]int{}, Present: []string{}, CreatedAt: b.now}
	b.games[m.ID] = append(b.games[m.ID], g)
	b.newGame = append(b.newGame, g)
	draw := c.Rules.Draw.Rule
	if c.Rules.Draw.GameDecides {
		draw = "game"
	}
	b.fx.launches = append(b.fx.launches, launchJob{gameID: g.ID, matchID: m.ID, launch: cuphooks.Launch{CupID: c.ID, MatchID: m.ID, GameNo: m.Games, Game: c.Rules.Game,
		Players: []string{m.A, m.B}, Options: c.Rules.GameOptions, Rated: c.Rules.Rated, Draw: draw, Final: m.Kind == kindMain && m.Round == c.Rounds, JoinSeconds: c.Rules.Timing.JoinSeconds}})
}

func (b *bracket) fresh(t *time.Time) bool {
	return t != nil && b.now.Sub(*t) <= time.Duration(b.cup.Rules.Timing.ReadySeconds)*time.Second
}

func (b *bracket) tickMatch(m *matchRow, rnd func(int) int) {
	t := b.cup.Rules.Timing
	now := b.now
	switch m.Status {
	case stWaiting:
		if m.OpensAt != nil && !now.Before(*m.OpensAt) {
			m.Status = stOpen
			b.touch(m)
			b.notify([]string{m.A, m.B}, "cup_check_in", map[string]any{"matchId": m.ID, "round": m.Round, "startsAt": iso(m.StartsAt), "closesAt": iso(m.ClosesAt)}, true, true)
		}
	case stOpen:
		if t.Mode == settings.ModeLive && m.Games == 0 && m.StartsAt != nil && !now.Before(*m.StartsAt) {
			if m.ShowedA && m.ShowedB {
				b.launch(m)
				return
			}
			b.decide(m, rnd)
			return
		}
		if m.ClosesAt != nil && !now.Before(*m.ClosesAt) {
			b.decide(m, rnd)
			return
		}
		if (t.Mode == settings.ModeWindow || m.Games > 0) && b.fresh(m.ReadyA) && b.fresh(m.ReadyB) {
			b.launch(m)
		}
	case stPlaying:
		g := b.currentGame(m)
		if g == nil || (g.Status != "launching" && g.Status != "playing") {
			m.Status = stOpen
			b.touch(m)
			return
		}
		if now.Sub(g.CreatedAt) >= time.Duration(t.GameMinutes)*time.Minute {
			b.voidGame(m, g, "timeout")
		}
	}
}

func (b *bracket) voidGame(m *matchRow, g *gameRow, reason string) {
	g.Status, g.Reason, g.FinishedAt = "void", reason, ptr(b.now)
	b.newGame = append(b.newGame, g)
	if m.Status == stPlaying {
		m.Status, m.GameMatchID, m.ReadyA, m.ReadyB = stOpen, "", nil, nil
		b.touch(m)
		b.notify([]string{m.A, m.B}, "cup_game_void", map[string]any{"matchId": m.ID, "gameNo": g.GameNo, "reason": reason, "closesAt": iso(m.ClosesAt)}, false, false)
	}
}

func (b *bracket) applyResult(m *matchRow, g *gameRow, r cuphooks.Result, rnd func(int) int) {
	rules := b.cup.Rules
	g.FinishedAt = ptr(b.now)
	b.newGame = append(b.newGame, g)
	present := []string{}
	for _, u := range r.Present {
		if u == m.A || u == m.B {
			present = append(present, u)
		}
	}
	g.Present, g.Score, g.Reason = present, r.Score, truncate(r.Reason, 64)
	if g.Score == nil {
		g.Score = map[string]int{}
	}
	winner := ""
	switch {
	case r.Void:
		if len(present) == 1 {
			winner, g.Reason = present[0], "no_show"
		} else {
			g.Status = "void"
			if m.Status == stPlaying && m.Games == g.GameNo {
				m.Status, m.GameMatchID, m.ReadyA, m.ReadyB = stOpen, "", nil, nil
				b.touch(m)
				b.notify([]string{m.A, m.B}, "cup_game_void", map[string]any{"matchId": m.ID, "gameNo": g.GameNo, "reason": "nobody_joined", "closesAt": iso(m.ClosesAt)}, false, false)
			}
			return
		}
	case r.Draw || r.Winner == "":
		g.Draw = true
	case r.Winner == m.A || r.Winner == m.B:
		winner = r.Winner
	default:
		g.Status = "void"
		g.Reason = "unknown_winner"
		return
	}
	g.Status = "finished"
	g.Winner = winner
	if m.Status != stPlaying || m.Games != g.GameNo {
		return
	}
	b.touch(m)
	if g.Draw {
		m.Draws++
		if rules.Draw.Rule == settings.DrawReplay && m.Draws <= rules.Draw.MaxReplays {
			b.notify([]string{m.A, m.B}, "cup_game_result", map[string]any{"matchId": m.ID, "gameNo": g.GameNo, "draw": true, "replay": true, "winsA": m.WinsA, "winsB": m.WinsB}, false, false)
			b.launch(m)
			return
		}
		rule := rules.Draw.Rule
		if rule == settings.DrawReplay {
			rule = rules.Draw.Fallback
		}
		winner = b.pick(m, rule, rnd)
		g.Winner, g.Reason = winner, "draw:"+rule
	}
	if winner == m.A {
		m.WinsA++
	} else {
		m.WinsB++
	}
	need := winsNeeded(m.BestOf)
	b.notify([]string{m.A, m.B}, "cup_game_result", map[string]any{"matchId": m.ID, "gameNo": g.GameNo, "winnerId": winner, "draw": g.Draw, "winsA": m.WinsA, "winsB": m.WinsB, "bestOf": m.BestOf}, false, false)
	if m.WinsA >= need || m.WinsB >= need {
		reason := "played"
		if g.Reason == "no_show" && m.BestOf == 1 {
			reason = "forfeit"
		}
		b.finish(m, winner, reason)
		return
	}
	b.launch(m)
}

func (b *bracket) finishCup(final *matchRow) {
	c := b.cup
	c.Status, c.FinishedAt, c.Champion = "finished", ptr(b.now), final.Winner
	b.fx.finished = true
	for _, p := range b.order {
		if p.Status == "playing" {
			p.Status = "eliminated"
			b.touchPlayer(p.UserID)
		}
	}
	for _, p := range b.order {
		if p.Seed > 0 {
			b.event(p.UserID, "cup_finished", map[string]any{"place": p.Place, "champion": p.UserID == c.Champion, "reached": p.Reached})
		}
	}
	for _, p := range b.order {
		if p.Seed > 0 {
			b.notify([]string{p.UserID}, "cup_finished", map[string]any{"championId": c.Champion, "place": p.Place}, true, p.Place > 0 && p.Place <= 4)
		}
	}
}

func (b *bracket) save(ctx context.Context, tx *sql.Tx) error {
	c := b.cup
	for _, m := range b.matches {
		if !m.dirty {
			continue
		}
		if m.Actor == "new" {
			m.Actor = ""
			if _, err := tx.ExecContext(ctx, `INSERT INTO cup_match (id, cup_id, round, slot, kind, status, best_of) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				m.ID, c.ID, m.Round, m.Slot, m.Kind, m.Status, m.BestOf); err != nil {
				return err
			}
		}
		if err := saveMatch(ctx, tx, m); err != nil {
			return err
		}
		m.dirty = false
	}
	seen := map[string]bool{}
	for _, g := range b.newGame {
		if seen[g.ID] {
			continue
		}
		seen[g.ID] = true
		if _, err := tx.ExecContext(ctx, `INSERT INTO cup_game (id, match_id, cup_id, game_no, game_match_id, status, winner, draw, score, present, reason, created_at, finished_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7::uuid,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, winner=EXCLUDED.winner, draw=EXCLUDED.draw, score=EXCLUDED.score, present=EXCLUDED.present,
			reason=EXCLUDED.reason, finished_at=EXCLUDED.finished_at, game_match_id=CASE WHEN cup_game.game_match_id='' THEN EXCLUDED.game_match_id ELSE cup_game.game_match_id END`,
			g.ID, g.MatchID, g.CupID, g.GameNo, g.GameMatchID, g.Status, nullUUID(g.Winner), g.Draw, jsonOf(g.Score), jsonOf(g.Present), g.Reason, g.CreatedAt, g.FinishedAt); err != nil {
			return err
		}
	}
	b.newGame = nil
	for id := range b.touched {
		p := b.players[id]
		if _, err := tx.ExecContext(ctx, `UPDATE cup_player SET status=$3, seed=$4, place=$5, reached=$6, trophies=$7, updated_at=now() WHERE cup_id=$1::uuid AND user_id=$2::uuid`,
			c.ID, id, p.Status, p.Seed, p.Place, p.Reached, p.Trophies); err != nil {
			return err
		}
	}
	b.touched = map[string]bool{}
	cur := 0
	for _, m := range b.matches {
		if (m.Status == stOpen || m.Status == stPlaying || m.Status == stWaiting) && m.Round > cur {
			cur = m.Round
		}
	}
	if cur > 0 {
		c.CurrentRound = cur
	}
	_, err := tx.ExecContext(ctx, `UPDATE cup SET status=$2, size=$3, rounds=$4, current_round=$5, players=$6, pool=$7, champion=$8::uuid, started_at=$9, finished_at=$10, cancel_reason=$11, updated_at=now() WHERE id=$1::uuid`,
		c.ID, c.Status, c.Size, c.Rounds, c.CurrentRound, c.Players, jsonOf(c.Pool), nullUUID(c.Champion), c.StartedAt, c.FinishedAt, c.CancelReason)
	if err != nil {
		return err
	}
	if b.fx.finished {
		return b.s.queuePayouts(ctx, tx, b)
	}
	return nil
}
