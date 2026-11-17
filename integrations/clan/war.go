package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/clanhooks"
	"github.com/heroiclabs/nakama/v3/internal/cronexpr"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
)

type warConfig struct {
	ID              string            `json:"id"`
	Key             string            `json:"key"`
	Name            map[string]string `json:"name"`
	Description     map[string]string `json:"description"`
	Mode            string            `json:"mode"`
	Active          bool              `json:"isActive"`
	Schedule        string            `json:"schedule"`
	DurationMinutes int               `json:"durationMinutes"`
	Signup          string            `json:"signup"`
	SignupMinutes   int               `json:"signupMinutes"`
	Pairing         string            `json:"pairing"`
	Rules           WarRules          `json:"rules"`
	UpdatedBy       string            `json:"updatedBy"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

const warConfigCols = `id::text, key, name::text, description::text, mode, is_active, schedule, duration_minutes, signup, signup_minutes, pairing, rules::text, updated_by, created_at, updated_at`

func scanWarConfig(row interface{ Scan(...any) error }) (*warConfig, error) {
	c := &warConfig{}
	var name, desc, rules string
	if err := row.Scan(&c.ID, &c.Key, &name, &desc, &c.Mode, &c.Active, &c.Schedule, &c.DurationMinutes, &c.Signup, &c.SignupMinutes, &c.Pairing, &rules, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(name), &c.Name)
	_ = json.Unmarshal([]byte(desc), &c.Description)
	c.Rules = defaultRules(c.Mode)
	_ = json.Unmarshal([]byte(rules), &c.Rules)
	c.Rules.normalize(c.Mode)
	if c.Name == nil {
		c.Name = map[string]string{}
	}
	if c.Description == nil {
		c.Description = map[string]string{}
	}
	return c, nil
}

func (s *service) warConfigs(ctx context.Context, activeOnly bool) ([]*warConfig, error) {
	q := `SELECT ` + warConfigCols + ` FROM clan_war_config`
	if activeOnly {
		q += ` WHERE is_active`
	}
	rows, err := s.db.QueryContext(ctx, q+` ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*warConfig{}
	for rows.Next() {
		c, err := scanWarConfig(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func cronNext(schedule string, t time.Time) *time.Time {
	if schedule == "" {
		return nil
	}
	e, err := cronexpr.Parse(schedule)
	if err != nil {
		return nil
	}
	n := e.Next(t.UTC())
	if n.IsZero() {
		return nil
	}
	return &n
}

func (s *service) createRound(ctx context.Context, c *warConfig, startsAt time.Time, duration int, actor string) (string, error) {
	if duration <= 0 {
		duration = c.DurationMinutes
	}
	rules, _ := json.Marshal(c.Rules)
	name, _ := json.Marshal(c.Name)
	var signupAt any
	if c.Signup == "leader" && c.SignupMinutes > 0 {
		signupAt = startsAt.Add(-time.Duration(c.SignupMinutes) * time.Minute)
	}
	id := uuid.Must(uuid.NewV4()).String()
	var got string
	err := s.db.QueryRowContext(ctx, `INSERT INTO clan_war_round (id, config_id, config_key, mode, status, signup_at, starts_at, ends_at, rules, name, pairing, signup, actor)
		VALUES ($1,$2,$3,$4,'signup',$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT (config_key, starts_at) DO UPDATE SET config_key=EXCLUDED.config_key RETURNING id::text`,
		id, c.ID, c.Key, c.Mode, signupAt, startsAt.UTC(), startsAt.UTC().Add(time.Duration(duration)*time.Minute), rules, name, c.Pairing, c.Signup, actor).Scan(&got)
	return got, err
}

func (s *service) scheduleRounds(ctx context.Context) error {
	list, err := s.warConfigs(ctx, true)
	if err != nil {
		return err
	}
	now := s.now()
	for _, c := range list {
		next := cronNext(c.Schedule, now)
		if next == nil {
			continue
		}
		var open bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_war_round WHERE config_key=$1 AND status IN ('signup','active') AND starts_at <> $2)`, c.Key, next.UTC()).Scan(&open)
		if open {
			continue
		}
		if _, err := s.createRound(ctx, c, *next, 0, "schedule"); err != nil {
			s.logger.Warn("Clan war %s round not scheduled: %v", c.Key, err)
		}
	}
	return nil
}

func (s *service) advanceRounds(ctx context.Context) {
	now := s.now().UTC()
	for _, step := range []struct {
		q  string
		fn func(context.Context, string) error
	}{
		{`SELECT id::text FROM clan_war_round WHERE status='signup' AND starts_at <= $1 ORDER BY starts_at LIMIT 20`, s.startRound},
		{`SELECT id::text FROM clan_war_round WHERE status='active' AND ends_at <= $1 ORDER BY ends_at LIMIT 20`, s.settleRound},
	} {
		rows, err := s.db.QueryContext(ctx, step.q, now)
		if err != nil {
			continue
		}
		ids := []string{}
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil {
				ids = append(ids, id)
			}
		}
		rows.Close()
		for _, id := range ids {
			if err := step.fn(ctx, id); err != nil {
				s.logger.Warn("Clan war round %s not advanced: %v", id, err)
			}
		}
	}
}

type roundRow struct {
	ID        string
	ConfigKey string
	Mode      string
	Status    string
	SignupAt  *time.Time
	StartsAt  time.Time
	EndsAt    time.Time
	Rules     WarRules
	Name      map[string]string
	Pairing   string
	Signup    string
	Wars      int
	Actor     string
	CreatedAt time.Time
	SettledAt *time.Time
}

const roundCols = `id::text, config_key, mode, status, signup_at, starts_at, ends_at, rules::text, name::text, pairing, signup, wars, actor, created_at, settled_at`

func scanRound(row interface{ Scan(...any) error }) (*roundRow, error) {
	r := &roundRow{}
	var rules, name string
	if err := row.Scan(&r.ID, &r.ConfigKey, &r.Mode, &r.Status, &r.SignupAt, &r.StartsAt, &r.EndsAt, &rules, &name, &r.Pairing, &r.Signup, &r.Wars, &r.Actor, &r.CreatedAt, &r.SettledAt); err != nil {
		return nil, err
	}
	r.Rules = defaultRules(r.Mode)
	_ = json.Unmarshal([]byte(rules), &r.Rules)
	r.Rules.normalize(r.Mode)
	_ = json.Unmarshal([]byte(name), &r.Name)
	if r.Name == nil {
		r.Name = map[string]string{}
	}
	return r, nil
}

type started struct {
	warID, a, b string
	membersA    []string
	membersB    []string
}

func (s *service) startRound(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	r, err := scanRound(tx.QueryRowContext(ctx, `SELECT `+roundCols+` FROM clan_war_round WHERE id::text=$1 AND status='signup' FOR UPDATE SKIP LOCKED`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	e := r.Rules.Eligibility
	q := `SELECT c.group_id::text, c.level, c.trophies, c.xp FROM clan c WHERE c.level >= $1`
	args := []any{e.MinClanLevel}
	if r.Signup == "leader" {
		q += ` AND EXISTS (SELECT 1 FROM clan_war_signup x WHERE x.round_id::text=$2 AND x.group_id=c.group_id)`
		args = append(args, r.ID)
	}
	rows, err := tx.QueryContext(ctx, q+` ORDER BY c.group_id`, args...)
	if err != nil {
		return err
	}
	cands := []entrant{}
	for rows.Next() {
		var x entrant
		if rows.Scan(&x.ID, &x.Level, &x.Trophies, &x.XP) == nil {
			cands = append(cands, x)
		}
	}
	rows.Close()
	cutoff := r.StartsAt.Add(-time.Duration(e.JoinedHours) * time.Hour)
	if e.JoinedHours == 0 {
		cutoff = s.now().Add(time.Hour)
	}
	roster := map[string][]string{}
	entrants := []entrant{}
	for _, c := range cands {
		limit := "ALL"
		if e.MaxMembers > 0 {
			limit = fmt.Sprint(e.MaxMembers)
		}
		mrows, err := tx.QueryContext(ctx, `SELECT m.user_id::text FROM clan_member m LEFT JOIN trophy_player t ON t.user_id=m.user_id
			WHERE m.group_id::text=$1 AND m.joined_at <= $2 AND COALESCE(t.trophies,0) >= $3 ORDER BY COALESCE(t.trophies,0) DESC, m.joined_at LIMIT `+limit, c.ID, cutoff, e.MinTrophies)
		if err != nil {
			return err
		}
		ids := []string{}
		for mrows.Next() {
			var uid string
			if mrows.Scan(&uid) == nil {
				ids = append(ids, uid)
			}
		}
		mrows.Close()
		if len(ids) >= e.MinMembers {
			roster[c.ID] = ids
			entrants = append(entrants, c)
		}
	}
	pairs := pairClans(entrants, r.Pairing, rand.New(rand.NewSource(time.Now().UnixNano())))
	out := []started{}
	for _, p := range pairs {
		wid := uuid.Must(uuid.NewV4()).String()
		var b any
		if p.B != "" {
			b = p.B
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO clan_war (id, round_id, clan_a, clan_b) VALUES ($1,$2,$3,$4)`, wid, r.ID, p.A, b); err != nil {
			return err
		}
		for _, side := range []string{p.A, p.B} {
			if side == "" {
				continue
			}
			for _, uid := range roster[side] {
				if _, err := tx.ExecContext(ctx, `INSERT INTO clan_war_member (war_id, group_id, user_id) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, wid, side, uid); err != nil {
					return err
				}
			}
		}
		out = append(out, started{warID: wid, a: p.A, b: p.B, membersA: roster[p.A], membersB: roster[p.B]})
	}
	if _, err := tx.ExecContext(ctx, `UPDATE clan_war_round SET status='active', wars=$2 WHERE id::text=$1`, r.ID, len(pairs)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	cfg := s.settings(ctx)
	for _, w := range out {
		for _, side := range []struct {
			clan, rival string
			members     []string
		}{{w.a, w.b, w.membersA}, {w.b, w.a, w.membersB}} {
			if side.clan == "" {
				continue
			}
			s.notify(ctx, s.memberIDs(ctx, side.clan), "clan_war_started", map[string]any{"warId": w.warID, "roundId": r.ID, "mode": r.Mode, "clanId": side.clan, "rivalClanId": side.rival,
				"endsAt": r.EndsAt.UTC().Format(time.RFC3339), "members": len(side.members), "bye": side.rival == ""}, true)
			s.say(ctx, cfg, side.clan, "war_started", "", w.warID)
		}
	}
	s.logger.Info("Clan war round %s (%s) started with %d wars", r.ID, r.ConfigKey, len(out))
	return nil
}

type settledSide struct {
	warID, clan, outcome, prize string
	score                       int64
}

func (s *service) settleRound(ctx context.Context, id string) error {
	cfg := s.settings(ctx)
	rd := s.road(ctx)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	r, err := scanRound(tx.QueryRowContext(ctx, `SELECT `+roundCols+` FROM clan_war_round WHERE id::text=$1 AND status='active' FOR UPDATE SKIP LOCKED`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	type war struct {
		id, a       string
		b           sql.NullString
		scoreA      int64
		scoreB      int64
		winsA, winB int
	}
	rows, err := tx.QueryContext(ctx, `SELECT id::text, clan_a::text, clan_b::text, score_a, score_b, wins_a, wins_b FROM clan_war WHERE round_id::text=$1 AND status='active' FOR UPDATE`, r.ID)
	if err != nil {
		return err
	}
	wars := []war{}
	for rows.Next() {
		var w war
		if err := rows.Scan(&w.id, &w.a, &w.b, &w.scoreA, &w.scoreB, &w.winsA, &w.winB); err != nil {
			rows.Close()
			return err
		}
		wars = append(wars, w)
	}
	rows.Close()
	sides := []settledSide{}
	xps := []*xpResult{}
	xpClans := []string{}
	for _, w := range wars {
		result := warResult(w.scoreA, w.scoreB, !w.b.Valid)
		if _, err := tx.ExecContext(ctx, `UPDATE clan_war SET status='settled', result=$2, settled_at=now() WHERE id::text=$1`, w.id, result); err != nil {
			return err
		}
		for _, side := range []struct {
			clan  string
			a     bool
			score int64
		}{{w.a, true, w.scoreA}, {w.b.String, false, w.scoreB}} {
			if side.clan == "" {
				continue
			}
			outcome := outcomeFor(result, side.a, r.Rules.ByeResult)
			col := map[string]string{"win": "war_wins", "lose": "war_losses", "draw": "war_draws"}[outcome]
			if col != "" {
				if _, err := tx.ExecContext(ctx, `UPDATE clan SET `+col+`=`+col+`+1, war_points=war_points+$2, updated_at=now() WHERE group_id::text=$1`, side.clan, side.score); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, `UPDATE clan_member m SET war_points=m.war_points+w.points, wars=m.wars+1 FROM clan_war_member w
				WHERE w.war_id::text=$1 AND w.group_id::text=$2 AND m.user_id=w.user_id AND m.group_id=w.group_id`, w.id, side.clan); err != nil {
				return err
			}
			prize := r.Rules.Prizes.keyFor(outcome)
			if prize != "" {
				cond := ` AND attacks > 0`
				if r.Rules.Prizes.To == "members" {
					cond = ``
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO clan_payout (id, source_id, kind, group_id, user_id, ref, prize_key)
					SELECT gen_random_uuid(), 'war:'||war_id::text||':'||user_id::text, 'war', group_id, user_id, $3, $4 FROM clan_war_member WHERE war_id::text=$1 AND group_id::text=$2`+cond+`
					ON CONFLICT DO NOTHING`, w.id, side.clan, outcome, prize); err != nil {
					return err
				}
			}
			amount := map[string]int64{"win": cfg.XP.WarWin, "draw": cfg.XP.WarDraw, "lose": cfg.XP.WarLoss}[outcome]
			if amount > 0 {
				x, err := s.addXPTx(ctx, tx, cfg, rd, side.clan, "", "war:"+w.id, "war_"+outcome, amount)
				if err != nil {
					return err
				}
				xps = append(xps, x)
				xpClans = append(xpClans, side.clan)
			}
			sides = append(sides, settledSide{warID: w.id, clan: side.clan, outcome: outcome, prize: prize, score: side.score})
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE clan_war_round SET status='settled', settled_at=now() WHERE id::text=$1`, r.ID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for i, x := range xps {
		s.afterXP(ctx, cfg, rd, xpClans[i], "", x)
	}
	for _, sd := range sides {
		s.notify(ctx, s.memberIDs(ctx, sd.clan), "clan_war_ended", map[string]any{"warId": sd.warID, "roundId": r.ID, "clanId": sd.clan, "outcome": sd.outcome, "score": sd.score, "prizeKey": sd.prize}, true)
		s.say(ctx, cfg, sd.clan, "war_ended", "", sd.outcome)
		if sd.outcome == "win" && cfg.Events.Publish {
			gameevents.Publish(ctx, gameevents.Event{ID: owner + ":w:" + sd.warID + ":" + sd.clan, UserID: sd.clan, Name: "clan_war_won", Context: r.ConfigKey, Source: owner,
				Fields: map[string]any{"count": float64(1), "points": float64(sd.score)}})
		}
		s.payPending(ctx, sd.clan, 500)
	}
	s.logger.Info("Clan war round %s (%s) settled: %d wars", r.ID, r.ConfigKey, len(wars))
	return nil
}

type activeWar struct {
	WarID   string
	RoundID string
	Mode    string
	Rules   WarRules
	ClanID  string
	ClanA   string
	ClanB   string
	EndsAt  time.Time
}

func (s *service) activeWars(ctx context.Context, userID string) ([]activeWar, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT w.id::text, r.id::text, r.mode, r.rules::text, wm.group_id::text, w.clan_a::text, COALESCE(w.clan_b::text,''), r.ends_at
		FROM clan_war_member wm JOIN clan_war w ON w.id=wm.war_id JOIN clan_war_round r ON r.id=w.round_id
		WHERE wm.user_id::text=$1 AND w.status='active' AND r.status='active' AND r.starts_at <= now() AND r.ends_at > now()`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []activeWar{}
	for rows.Next() {
		var a activeWar
		var rules string
		if err := rows.Scan(&a.WarID, &a.RoundID, &a.Mode, &rules, &a.ClanID, &a.ClanA, &a.ClanB, &a.EndsAt); err != nil {
			return nil, err
		}
		a.Rules = defaultRules(a.Mode)
		_ = json.Unmarshal([]byte(rules), &a.Rules)
		a.Rules.normalize(a.Mode)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *service) scoreWars(ctx context.Context, in incoming) {
	wars, err := s.activeWars(ctx, in.UserID)
	if err != nil || len(wars) == 0 {
		return
	}
	for _, w := range wars {
		if w.ClanB == "" {
			continue
		}
		var err error
		switch w.Mode {
		case modePoints:
			err = s.scorePoints(ctx, w, in)
		case modeDuel:
			err = s.scoreDuel(ctx, w, in)
		}
		if err != nil {
			s.logger.Warn("Clan war %s: event %s of %s not scored: %v", w.WarID, in.ID, in.UserID, err)
		}
	}
}

func matchKey(field string, in incoming) string {
	if field != "" {
		if v := str(in.Fields[field]); v != "" {
			return v
		}
	}
	return in.ID
}

func sideCols(w activeWar, clan string) (string, string) {
	if clan == w.ClanA {
		return "score_a", "wins_a"
	}
	return "score_b", "wins_b"
}

func (s *service) attackState(ctx context.Context, tx *sql.Tx, w activeWar, userID, match string) (bool, bool, error) {
	var seen bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_war_event WHERE war_id::text=$1 AND user_id::text=$2 AND match_id=$3 AND counted)`, w.WarID, userID, match).Scan(&seen); err != nil {
		return false, false, err
	}
	if seen || w.Rules.AttacksPerMember <= 0 {
		return !seen, false, nil
	}
	var attacks int
	if err := tx.QueryRowContext(ctx, `SELECT attacks FROM clan_war_member WHERE war_id::text=$1 AND user_id::text=$2`, w.WarID, userID).Scan(&attacks); err != nil {
		return false, false, err
	}
	return true, attacks >= w.Rules.AttacksPerMember, nil
}

func (s *service) scorePoints(ctx context.Context, w activeWar, in incoming) error {
	match := matchKey(w.Rules.MatchField, in)
	total := int64(0)
	hits := []int{}
	for i, p := range w.Rules.Points {
		if p.Type != in.Type || p.Event != in.Name {
			continue
		}
		if pts := pointsFor(p, in.Value, in.Fields); pts != 0 {
			total += pts
			hits = append(hits, i)
		}
	}
	if len(hits) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT 1 FROM clan_war WHERE id::text=$1 FOR UPDATE`, w.WarID); err != nil {
		return err
	}
	fresh, over, err := s.attackState(ctx, tx, w, in.UserID, match)
	if err != nil {
		return err
	}
	source := fmt.Sprintf("p:%s:%s", in.Name, match)
	res, err := tx.ExecContext(ctx, `INSERT INTO clan_war_event (war_id, user_id, group_id, source_id, match_id, kind, points, counted) VALUES ($1,$2,$3,$4,$5,'points',$6,$7) ON CONFLICT DO NOTHING`,
		w.WarID, in.UserID, w.ClanID, source, match, total, !(fresh && over))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 || (fresh && over) {
		return tx.Commit()
	}
	add := 0
	if fresh {
		add = 1
	}
	if _, err := tx.ExecContext(ctx, `UPDATE clan_war_member SET points=points+$3, attacks=attacks+$4 WHERE war_id::text=$1 AND user_id::text=$2`, w.WarID, in.UserID, total, add); err != nil {
		return err
	}
	score, _ := sideCols(w, w.ClanID)
	if _, err := tx.ExecContext(ctx, `UPDATE clan_war SET `+score+`=`+score+`+$2 WHERE id::text=$1`, w.WarID, total); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.afterScore(ctx, w, in.UserID, total)
	return nil
}

func (s *service) scoreDuel(ctx context.Context, w activeWar, in incoming) error {
	d := w.Rules.Duel
	if d.Type != in.Type || d.Event != in.Name {
		return nil
	}
	opp := str(in.Fields[d.OpponentField])
	if opp == "" || opp == in.UserID {
		return nil
	}
	match := matchKey(w.Rules.MatchField, in)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT 1 FROM clan_war WHERE id::text=$1 FOR UPDATE`, w.WarID); err != nil {
		return err
	}
	var oppClan string
	if err := tx.QueryRowContext(ctx, `SELECT group_id::text FROM clan_war_member WHERE war_id::text=$1 AND user_id::text=$2`, w.WarID, opp).Scan(&oppClan); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if oppClan == w.ClanID {
		return nil
	}
	won := truthy(in.Fields[d.WonField])
	draw := d.DrawField != "" && truthy(in.Fields[d.DrawField])
	fresh, over, err := s.attackState(ctx, tx, w, in.UserID, match)
	if err != nil {
		return err
	}
	kind := "duel_loss"
	switch {
	case draw:
		kind = "duel_draw"
	case won:
		kind = "duel_win"
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO clan_war_event (war_id, user_id, group_id, source_id, match_id, opponent_id, kind, points, counted) VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8) ON CONFLICT DO NOTHING`,
		w.WarID, in.UserID, w.ClanID, "d:"+match, match, opp, kind, !(fresh && over))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 || (fresh && over) {
		return tx.Commit()
	}
	col := map[string]string{"duel_win": "wins", "duel_loss": "losses", "duel_draw": "draws"}[kind]
	if _, err := tx.ExecContext(ctx, `UPDATE clan_war_member SET attacks=attacks+1, `+col+`=`+col+`+1 WHERE war_id::text=$1 AND user_id::text=$2`, w.WarID, in.UserID); err != nil {
		return err
	}
	var points int64
	if won || draw {
		var counted bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_war_event WHERE war_id::text=$1 AND source_id=$2)`, w.WarID, "r:"+match).Scan(&counted); err != nil {
			return err
		}
		if !counted {
			points = d.WinPoints
			if draw {
				points = d.DrawPoints
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO clan_war_event (war_id, user_id, group_id, source_id, match_id, opponent_id, kind, points) VALUES ($1,$2,$3,$4,$5,$6,'result',$7)`,
				w.WarID, in.UserID, w.ClanID, "r:"+match, match, opp, points); err != nil {
				return err
			}
			score, wins := sideCols(w, w.ClanID)
			if draw {
				if _, err := tx.ExecContext(ctx, `UPDATE clan_war SET score_a=score_a+$2, score_b=score_b+$2 WHERE id::text=$1`, w.WarID, points); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, `UPDATE clan_war_member SET points=points+$3 WHERE war_id::text=$1 AND user_id::text = ANY($2)`, w.WarID, []string{in.UserID, opp}, points); err != nil {
					return err
				}
			} else {
				if _, err := tx.ExecContext(ctx, `UPDATE clan_war SET `+score+`=`+score+`+$2, `+wins+`=`+wins+`+1 WHERE id::text=$1`, w.WarID, points); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, `UPDATE clan_war_member SET points=points+$3 WHERE war_id::text=$1 AND user_id::text=$2`, w.WarID, in.UserID, points); err != nil {
					return err
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.afterScore(ctx, w, in.UserID, points)
	return nil
}

func (s *service) afterScore(ctx context.Context, w activeWar, userID string, points int64) {
	var a, b int64
	var wa, wb int
	_ = s.db.QueryRowContext(ctx, `SELECT score_a, score_b, wins_a, wins_b FROM clan_war WHERE id::text=$1`, w.WarID).Scan(&a, &b, &wa, &wb)
	cfg := s.settings(ctx)
	if points != 0 && cfg.Events.Publish {
		gameevents.Publish(ctx, gameevents.Event{ID: owner + ":wp:" + uuid.Must(uuid.NewV4()).String(), UserID: userID, Name: "clan_war_points", Context: w.ClanID, Source: owner,
			Fields: map[string]any{"points": float64(points), "warId": w.WarID}})
	}
	ids := append(s.memberIDs(ctx, w.ClanA), s.memberIDs(ctx, w.ClanB)...)
	s.notify(ctx, ids, "clan_war_score", map[string]any{"warId": w.WarID, "clanA": w.ClanA, "clanB": w.ClanB, "scoreA": a, "scoreB": b, "winsA": wa, "winsB": wb, "userId": userID, "points": points}, false)
}

func (s *service) rivalOf(ctx context.Context, userID, otherID string) (clanhooks.Rival, bool) {
	var r clanhooks.Rival
	err := s.db.QueryRowContext(ctx, `SELECT w.id::text, r.id::text, r.mode, ma.group_id::text, mb.group_id::text
		FROM clan_war_member ma JOIN clan_war_member mb ON mb.war_id=ma.war_id AND mb.group_id<>ma.group_id
		JOIN clan_war w ON w.id=ma.war_id JOIN clan_war_round r ON r.id=w.round_id
		WHERE ma.user_id::text=$1 AND mb.user_id::text=$2 AND w.status='active' AND r.status='active' AND r.starts_at <= now() AND r.ends_at > now()
		ORDER BY (r.mode='duel') DESC LIMIT 1`, userID, otherID).Scan(&r.WarID, &r.RoundID, &r.Mode, &r.ClanID, &r.RivalID)
	if err != nil {
		return clanhooks.Rival{}, false
	}
	return r, true
}

func (s *service) warSide(ctx context.Context, warID, clanID string, tx texts) (map[string]any, error) {
	var name, avatar, tag, badge string
	var level int
	_ = s.db.QueryRowContext(ctx, `SELECT g.name, g.avatar_url, c.tag, c.badge::text, c.level FROM groups g JOIN clan c ON c.group_id=g.id WHERE g.id::text=$1`, clanID).Scan(&name, &avatar, &tag, &badge, &level)
	var b map[string]any
	_ = json.Unmarshal([]byte(badge), &b)
	rows, err := s.db.QueryContext(ctx, `SELECT w.user_id::text, COALESCE(u.username,''), COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), w.points, w.attacks, w.wins, w.losses, w.draws, COALESCE(t.trophies,0)
		FROM clan_war_member w LEFT JOIN users u ON u.id=w.user_id LEFT JOIN trophy_player t ON t.user_id=w.user_id WHERE w.war_id::text=$1 AND w.group_id::text=$2 ORDER BY w.points DESC, w.attacks DESC`, warID, clanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := []map[string]any{}
	for rows.Next() {
		var id, uname, dname, av string
		var points int64
		var attacks, wins, losses, draws, trophies int
		if err := rows.Scan(&id, &uname, &dname, &av, &points, &attacks, &wins, &losses, &draws, &trophies); err != nil {
			return nil, err
		}
		members = append(members, map[string]any{"userId": id, "username": uname, "displayName": dname, "avatarUrl": av, "points": points, "attacks": attacks, "wins": wins,
			"losses": losses, "draws": draws, "trophies": trophies})
	}
	return map[string]any{"clanId": clanID, "name": name, "avatarUrl": avatar, "tag": tag, "badge": b, "level": level, "members": members}, nil
}

func (s *service) warView(ctx context.Context, warID, myClan, userID string, tx texts) (map[string]any, error) {
	var roundID, mode, status, result, a, rulesRaw, nameRaw string
	var b sql.NullString
	var scoreA, scoreB int64
	var winsA, winsB int
	var starts, ends time.Time
	var settled *time.Time
	err := s.db.QueryRowContext(ctx, `SELECT w.round_id::text, r.mode, w.status, w.result, w.clan_a::text, w.clan_b::text, w.score_a, w.score_b, w.wins_a, w.wins_b, r.starts_at, r.ends_at, w.settled_at, r.rules::text, r.name::text
		FROM clan_war w JOIN clan_war_round r ON r.id=w.round_id WHERE w.id::text=$1`, warID).Scan(&roundID, &mode, &status, &result, &a, &b, &scoreA, &scoreB, &winsA, &winsB, &starts, &ends, &settled, &rulesRaw, &nameRaw)
	if err != nil {
		return nil, err
	}
	rules := defaultRules(mode)
	_ = json.Unmarshal([]byte(rulesRaw), &rules)
	rules.normalize(mode)
	var name map[string]string
	_ = json.Unmarshal([]byte(nameRaw), &name)
	usA := myClan == a
	us, them := a, b.String
	usScore, themScore, usWins, themWins := scoreA, scoreB, winsA, winsB
	if !usA {
		us, them = b.String, a
		usScore, themScore, usWins, themWins = scoreB, scoreA, winsB, winsA
	}
	usSide, err := s.warSide(ctx, warID, us, tx)
	if err != nil {
		return nil, err
	}
	usSide["score"], usSide["wins"] = usScore, usWins
	var themSide map[string]any
	if them != "" {
		if themSide, err = s.warSide(ctx, warID, them, tx); err != nil {
			return nil, err
		}
		themSide["score"], themSide["wins"] = themScore, themWins
	}
	var attacks int
	inWar := false
	if err := s.db.QueryRowContext(ctx, `SELECT attacks FROM clan_war_member WHERE war_id::text=$1 AND user_id::text=$2`, warID, userID).Scan(&attacks); err == nil {
		inWar = true
	}
	var left any
	if rules.AttacksPerMember > 0 && inWar {
		left = max(0, rules.AttacksPerMember-attacks)
	}
	outcome := ""
	if status == "settled" {
		outcome = outcomeFor(result, usA, rules.ByeResult)
	}
	points := make([]map[string]any, 0, len(rules.Points))
	for _, p := range rules.Points {
		points = append(points, map[string]any{"event": p.Event, "points": p.Points, "perValue": p.PerValue})
	}
	return map[string]any{"warId": warID, "roundId": roundID, "mode": mode, "name": tx.of(name, ""), "status": status, "outcome": outcome, "startsAt": starts.UTC().Format(time.RFC3339),
		"endsAt": ends.UTC().Format(time.RFC3339), "settledAt": iso(settled), "us": usSide, "them": themSide, "inWar": inWar, "myAttacks": attacks, "myAttacksLeft": left,
		"attacksPerMember": rules.AttacksPerMember, "points": points, "duel": map[string]any{"winPoints": rules.Duel.WinPoints, "drawPoints": rules.Duel.DrawPoints},
		"prizes": map[string]any{"win": rules.Prizes.Win.PrizeKey, "lose": rules.Prizes.Lose.PrizeKey, "draw": rules.Prizes.Draw.PrizeKey, "to": rules.Prizes.To}}, nil
}

func (s *service) rpcWarCurrent(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	m, err := s.requireMember(ctx, userID)
	if err != nil {
		return nil, err
	}
	tx := s.texts(ctx, userID, in.Lang)
	rows, err := s.db.QueryContext(ctx, `SELECT w.id::text FROM clan_war w JOIN clan_war_round r ON r.id=w.round_id WHERE (w.clan_a::text=$1 OR w.clan_b::text=$1) AND w.status='active' ORDER BY r.ends_at`, m.ClanID)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	wars := []map[string]any{}
	for _, id := range ids {
		v, err := s.warView(ctx, id, m.ClanID, userID, tx)
		if err != nil {
			return nil, err
		}
		wars = append(wars, v)
	}
	srows, err := s.db.QueryContext(ctx, `SELECT r.id::text, r.config_key, r.mode, r.name::text, r.signup, r.signup_at, r.starts_at, r.ends_at,
		EXISTS (SELECT 1 FROM clan_war_signup x WHERE x.round_id=r.id AND x.group_id::text=$1) FROM clan_war_round r WHERE r.status='signup' ORDER BY r.starts_at LIMIT 20`, m.ClanID)
	if err != nil {
		return nil, err
	}
	defer srows.Close()
	upcoming := []map[string]any{}
	cfg := s.settings(ctx)
	for srows.Next() {
		var id, key, mode, nameRaw, signup string
		var signupAt *time.Time
		var starts, ends time.Time
		var signed bool
		if err := srows.Scan(&id, &key, &mode, &nameRaw, &signup, &signupAt, &starts, &ends, &signed); err != nil {
			return nil, err
		}
		var name map[string]string
		_ = json.Unmarshal([]byte(nameRaw), &name)
		open := signup == "leader" && (signupAt == nil || !s.now().Before(*signupAt))
		upcoming = append(upcoming, map[string]any{"roundId": id, "configKey": key, "mode": mode, "name": tx.of(name, key), "signup": signup, "signupAt": iso(signupAt),
			"startsAt": starts.UTC().Format(time.RFC3339), "endsAt": ends.UTC().Format(time.RFC3339), "signedUp": signed || signup == "auto", "signupOpen": open,
			"canSignUp": open && !signed && allowed(cfg.Roles.War, m.Role)})
	}
	return map[string]any{"clanId": m.ClanID, "wars": wars, "upcoming": upcoming}, nil
}

func (s *service) rpcWarHistory(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit int    `json:"limit"`
		Lang  string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	m, err := s.requireMember(ctx, userID)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	limit := in.Limit
	if limit <= 0 || limit > cfg.Display.HistoryLimit {
		limit = cfg.Display.HistoryLimit
	}
	tx := s.texts(ctx, userID, in.Lang)
	rows, err := s.db.QueryContext(ctx, `SELECT w.id::text, r.mode, r.name::text, r.rules::text, w.result, w.clan_a::text, COALESCE(w.clan_b::text,''), w.score_a, w.score_b, w.wins_a, w.wins_b, w.settled_at, COALESCE(g.name,'')
		FROM clan_war w JOIN clan_war_round r ON r.id=w.round_id LEFT JOIN groups g ON g.id = CASE WHEN w.clan_a::text=$1 THEN w.clan_b ELSE w.clan_a END
		WHERE (w.clan_a::text=$1 OR w.clan_b::text=$1) AND w.status='settled' ORDER BY w.settled_at DESC LIMIT $2`, m.ClanID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, mode, nameRaw, rulesRaw, result, a, b, rival string
		var sa, sb int64
		var wa, wb int
		var at *time.Time
		if err := rows.Scan(&id, &mode, &nameRaw, &rulesRaw, &result, &a, &b, &sa, &sb, &wa, &wb, &at, &rival); err != nil {
			return nil, err
		}
		var name map[string]string
		_ = json.Unmarshal([]byte(nameRaw), &name)
		rules := defaultRules(mode)
		_ = json.Unmarshal([]byte(rulesRaw), &rules)
		usA := a == m.ClanID
		rivalID, us, them, uw, tw := b, sa, sb, wa, wb
		if !usA {
			rivalID, us, them, uw, tw = a, sb, sa, wb, wa
		}
		items = append(items, map[string]any{"warId": id, "mode": mode, "name": tx.of(name, ""), "outcome": outcomeFor(result, usA, rules.ByeResult), "rivalClanId": rivalID, "rivalName": rival,
			"score": us, "rivalScore": them, "wins": uw, "rivalWins": tw, "settledAt": iso(at)})
	}
	return map[string]any{"clanId": m.ClanID, "items": items}, nil
}

func (s *service) rpcWarSignup(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		RoundID string `json:"roundId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.War)
	if err != nil {
		return nil, err
	}
	if !uuidPattern.MatchString(in.RoundID) {
		return nil, apierr.BadRequest("CLAN_WAR_ROUND_REQUIRED").WithField("roundId")
	}
	r, err := scanRound(s.db.QueryRowContext(ctx, `SELECT `+roundCols+` FROM clan_war_round WHERE id::text=$1`, in.RoundID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("CLAN_WAR_ROUND_NOT_FOUND").WithField("roundId")
	}
	if err != nil {
		return nil, err
	}
	if r.Status != "signup" || r.Signup != "leader" || (r.SignupAt != nil && s.now().Before(*r.SignupAt)) {
		return nil, apierr.Conflict("CLAN_WAR_SIGNUP_CLOSED").With("status", r.Status)
	}
	var lv int
	_ = s.db.QueryRowContext(ctx, `SELECT level FROM clan WHERE group_id::text=$1`, m.ClanID).Scan(&lv)
	if lv < r.Rules.Eligibility.MinClanLevel {
		return nil, apierr.Forbidden("CLAN_WAR_LEVEL_TOO_LOW").With("need", r.Rules.Eligibility.MinClanLevel)
	}
	var n int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM clan_member WHERE group_id::text=$1`, m.ClanID).Scan(&n)
	if n < r.Rules.Eligibility.MinMembers {
		return nil, apierr.Forbidden("CLAN_WAR_NOT_ENOUGH_MEMBERS").With("need", r.Rules.Eligibility.MinMembers)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO clan_war_signup (round_id, group_id, by_user) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, r.ID, m.ClanID, userID); err != nil {
		return nil, err
	}
	s.say(ctx, cfg, m.ClanID, "war_signup", userID, r.ID)
	return map[string]any{"roundId": r.ID, "clanId": m.ClanID, "signedUp": true, "startsAt": r.StartsAt.UTC().Format(time.RFC3339)}, nil
}

func (s *service) rpcWarOpponents(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		WarID string `json:"warId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	wars, err := s.activeWars(ctx, userID)
	if err != nil {
		return nil, err
	}
	var w *activeWar
	for i := range wars {
		if wars[i].ClanB != "" && (in.WarID == "" || wars[i].WarID == in.WarID) && (w == nil || wars[i].Mode == modeDuel) {
			w = &wars[i]
		}
	}
	if w == nil {
		return nil, apierr.NotFound("CLAN_WAR_NOT_ACTIVE")
	}
	rival := w.ClanA
	if rival == w.ClanID {
		rival = w.ClanB
	}
	rows, err := s.db.QueryContext(ctx, `SELECT w.user_id::text, COALESCE(u.username,''), COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), w.attacks, w.points, COALESCE(t.trophies,0),
		EXISTS (SELECT 1 FROM clan_war_event e WHERE e.war_id=w.war_id AND e.user_id::text=$3 AND e.opponent_id=w.user_id::text AND e.counted)
		FROM clan_war_member w LEFT JOIN users u ON u.id=w.user_id LEFT JOIN trophy_player t ON t.user_id=w.user_id WHERE w.war_id::text=$1 AND w.group_id::text=$2 ORDER BY COALESCE(t.trophies,0) DESC`, w.WarID, rival, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, uname, dname, av string
		var attacks, trophies int
		var points int64
		var played bool
		if err := rows.Scan(&id, &uname, &dname, &av, &attacks, &points, &trophies, &played); err != nil {
			return nil, err
		}
		var left any
		if w.Rules.AttacksPerMember > 0 {
			left = max(0, w.Rules.AttacksPerMember-attacks)
		}
		items = append(items, map[string]any{"userId": id, "username": uname, "displayName": dname, "avatarUrl": av, "attacks": attacks, "attacksLeft": left, "points": points,
			"trophies": trophies, "playedByMe": played})
	}
	return map[string]any{"warId": w.WarID, "mode": w.Mode, "rivalClanId": rival, "endsAt": w.EndsAt.UTC().Format(time.RFC3339), "items": items}, nil
}

func (s *service) reconcileCaps(ctx context.Context) {
	r := s.road(ctx)
	rows, err := s.db.QueryContext(ctx, `SELECT c.group_id::text, c.level, c.join_type, g.max_count FROM clan c JOIN groups g ON g.id=c.group_id LIMIT 5000`)
	if err != nil {
		return
	}
	type fix struct {
		id   string
		open bool
		cap  int
	}
	list := []fix{}
	for rows.Next() {
		var id, join string
		var lv, cur int
		if rows.Scan(&id, &lv, &join, &cur) == nil {
			if want := r.memberCap(lv, 0); want > 0 && want != cur {
				list = append(list, fix{id: id, open: join == settings.JoinOpen, cap: want})
			}
		}
	}
	rows.Close()
	for _, f := range list {
		if err := s.groups.Update(ctx, f.id, "", "", "", "", f.open, f.cap); err != nil {
			s.logger.Debug("Clan %s cap not reconciled: %v", f.id, err)
		}
	}
}
