package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/trophy/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/cronexpr"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
)

type change struct {
	UserID        string `json:"userId"`
	Reason        string `json:"reason"`
	Delta         int    `json:"delta"`
	Before        int    `json:"before"`
	Trophies      int    `json:"trophies"`
	Best          int    `json:"best"`
	League        string `json:"league"`
	LeagueBefore  string `json:"leagueBefore"`
	LeagueChanged bool   `json:"leagueChanged"`
	MatchID       string `json:"matchId"`
	Duplicate     bool   `json:"duplicate"`
	PrizeKey      string `json:"prizeKey,omitempty"`
	logID         int64
}

func (s *service) onEconomyEvent(ctx context.Context, ev playerevents.Recorded) {
	cfg := s.settings(ctx)
	if !cfg.Enabled || cfg.Match.Source != settings.SourceEconomy || cfg.Match.Event == "" || ev.Key != cfg.Match.Event {
		return
	}
	s.fromFields(ctx, cfg, ev.UserID, "ev:"+ev.ID, ev.Metadata)
}

func (s *service) onGameEvent(ctx context.Context, ev gameevents.Event) (any, error) {
	if strings.HasPrefix(ev.ID, owner+":") {
		return nil, nil
	}
	cfg := s.settings(ctx)
	if !cfg.Enabled || cfg.Match.Source != settings.SourceGame || cfg.Match.Event == "" || ev.Name != cfg.Match.Event {
		return nil, nil
	}
	ch := s.fromFields(ctx, cfg, ev.UserID, "ge:"+ev.ID, ev.Fields)
	if ch == nil {
		return nil, nil
	}
	return ch, nil
}

func (s *service) fromFields(ctx context.Context, cfg settings.All, userID, fallback string, fields map[string]any) *change {
	f, ok := readMatch(cfg.Match, fields)
	if !ok {
		return nil
	}
	if f.Opponent == userID {
		f.Opponent = ""
	}
	source := fallback
	if f.MatchID != "" {
		source = "m:" + f.MatchID
	}
	ch, err := s.applyMatch(ctx, cfg, userID, source, f)
	if err != nil {
		s.logger.Warn("Trophies for %s (%s) not applied: %v", userID, source, err)
		return nil
	}
	return ch
}

func (s *service) applyMatch(ctx context.Context, cfg settings.All, userID, source string, f matchFacts) (*change, error) {
	l := s.ladder(ctx)
	return s.write(ctx, cfg, l, userID, source, f.Outcome, func(before int, q *sql.Tx) (int, *int, error) {
		var opp *int
		if f.Opponent != "" {
			var v int
			err := q.QueryRowContext(ctx, `SELECT before FROM trophy_log WHERE user_id::text=$1 AND source_id=$2`, f.Opponent, source).Scan(&v)
			if errors.Is(err, sql.ErrNoRows) {
				err = q.QueryRowContext(ctx, `SELECT trophies FROM trophy_player WHERE user_id::text=$1`, f.Opponent).Scan(&v)
				if errors.Is(err, sql.ErrNoRows) {
					v, err = cfg.Amounts.Start, nil
				}
			}
			if err != nil {
				return 0, nil, err
			}
			opp = &v
		}
		return matchDelta(cfg.Amounts, f.Outcome, before, opp), opp, nil
	}, f.MatchID, f.Opponent, "")
}

func (s *service) write(ctx context.Context, cfg settings.All, l ladder, userID, source, reason string, deltaOf func(before int, tx *sql.Tx) (int, *int, error), matchID, opponent, actor string) (*change, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO trophy_player (user_id, trophies, best, season_best, league_key) VALUES ($1,$2,$2,$2,'') ON CONFLICT (user_id) DO NOTHING`, userID, cfg.Amounts.Start); err != nil {
		return nil, err
	}
	var before, best int
	var leagueBefore string
	if err := tx.QueryRowContext(ctx, `SELECT trophies, best, league_key FROM trophy_player WHERE user_id=$1 FOR UPDATE`, userID).Scan(&before, &best, &leagueBefore); err != nil {
		return nil, err
	}
	var dup bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM trophy_log WHERE user_id=$1 AND source_id=$2)`, userID, source).Scan(&dup); err != nil {
		return nil, err
	}
	if dup {
		return &change{UserID: userID, Duplicate: true, Trophies: before, Best: best, League: leagueBefore, MatchID: matchID}, nil
	}
	delta, opp, err := deltaOf(before, tx)
	if err != nil {
		return nil, err
	}
	after := before + delta
	if reason != "admin" {
		after = applyDelta(l, cfg.Amounts.Floors, before, delta)
	}
	if after < 0 {
		after = 0
	}
	leagueAfter := ""
	if cur := l.at(after); cur != nil {
		leagueAfter = cur.Key
	}
	if leagueBefore == "" {
		if cur := l.at(before); cur != nil {
			leagueBefore = cur.Key
		}
	}
	wins, losses, draws := 0, 0, 0
	switch reason {
	case outcomeWin:
		wins = 1
	case outcomeLoss:
		losses = 1
	case outcomeDraw:
		draws = 1
	}
	if _, err := tx.ExecContext(ctx, `UPDATE trophy_player SET trophies=$2, best=GREATEST(best,$2), season_best=GREATEST(season_best,$2), league_key=$3,
		wins=wins+$4, losses=losses+$5, draws=draws+$6, updated_at=now() WHERE user_id=$1`, userID, after, leagueAfter, wins, losses, draws); err != nil {
		return nil, err
	}
	var logID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO trophy_log (user_id, source_id, reason, delta, before, after, match_id, opponent_id, opponent_trophies, league_before, league_after, actor)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`, userID, source, reason, after-before, before, after, matchID, opponent, opp, leagueBefore, leagueAfter, actor).Scan(&logID); err != nil {
		return nil, err
	}
	ch := &change{UserID: userID, Reason: reason, Delta: after - before, Before: before, Trophies: after, Best: max(best, after), League: leagueAfter, LeagueBefore: leagueBefore,
		LeagueChanged: leagueAfter != leagueBefore, MatchID: matchID, logID: logID}
	if ch.LeagueChanged && leagueAfter != "" {
		risen := l.index(after) > l.index(before)
		if cur := l.at(after); cur != nil && cur.PrizeKey != "" && (risen || !cfg.Prizes.OnlyRising) {
			season := 0
			_ = tx.QueryRowContext(ctx, `SELECT COALESCE(max(number),0) FROM trophy_season WHERE ended_at IS NULL`).Scan(&season)
			res, err := tx.ExecContext(ctx, `INSERT INTO trophy_prize (user_id, claim_key, league_key, season, prize_key) VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`,
				userID, claimKey(cfg.Prizes, cur.Key, season), cur.Key, season, cur.PrizeKey)
			if err != nil {
				return nil, err
			}
			if n, _ := res.RowsAffected(); n > 0 {
				ch.PrizeKey = cur.PrizeKey
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.after(ctx, cfg, ch)
	return ch, nil
}

func (s *service) after(ctx context.Context, cfg settings.All, ch *change) {
	if ch.PrizeKey != "" {
		s.payPending(ctx, ch.UserID, 10)
	}
	if cfg.Events.Publish {
		gameevents.Publish(ctx, gameevents.Event{ID: fmt.Sprintf("%s:%d", owner, ch.logID), UserID: ch.UserID, Name: "trophy_changed", Context: ch.League, Source: owner,
			Fields: map[string]any{"trophies": float64(ch.Trophies), "delta": float64(ch.Delta), "best": float64(ch.Best), "reason": ch.Reason, "matchId": ch.MatchID}})
	}
	if !cfg.Events.Notify {
		return
	}
	s.notify(ctx, ch.UserID, "trophy_changed", map[string]any{"delta": ch.Delta, "trophies": ch.Trophies, "best": ch.Best, "league": ch.League, "leagueBefore": ch.LeagueBefore,
		"leagueChanged": ch.LeagueChanged, "reason": ch.Reason, "matchId": ch.MatchID}, cfg.Events.NotifyCode, false)
	if ch.LeagueChanged && ch.League != "" {
		s.notify(ctx, ch.UserID, "trophy_league", map[string]any{"league": ch.League, "leagueBefore": ch.LeagueBefore, "trophies": ch.Trophies, "prizeKey": ch.PrizeKey,
			"up": ch.Delta > 0}, cfg.Events.LeagueCode, ch.Delta > 0)
	}
}

func (s *service) adminWrite(ctx context.Context, userID string, value int, mode, actor string) (*change, error) {
	cfg := s.settings(ctx)
	l := s.ladder(ctx)
	return s.write(ctx, cfg, l, userID, "admin:"+uuid.Must(uuid.NewV4()).String(), "admin", func(before int, _ *sql.Tx) (int, *int, error) {
		if mode == "add" {
			return value, nil, nil
		}
		return value - before, nil, nil
	}, "", "", actor)
}

func (s *service) payPending(ctx context.Context, userID string, limit int) {
	cfg := s.settings(ctx)
	for i := 0; i < limit; i++ {
		if !s.payOne(ctx, cfg, userID) {
			return
		}
	}
}

func (s *service) payOne(ctx context.Context, cfg settings.All, userID string) bool {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false
	}
	defer func() { _ = tx.Rollback() }()
	q := `SELECT user_id::text, claim_key, league_key, prize_key, tries FROM trophy_prize WHERE status='pending'`
	args := []any{}
	if userID != "" {
		q += ` AND user_id=$1`
		args = append(args, userID)
	}
	var uid, key, lk, prize string
	var tries int
	err = tx.QueryRowContext(ctx, q+` ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED`, args...).Scan(&uid, &key, &lk, &prize, &tries)
	if err != nil {
		return false
	}
	res, gerr := s.prize(ctx, economyhooks.PrizeGrant{UserID: uid, PrizeKey: prize, SourceType: sourceType, SourceID: "league:" + uid + ":" + key, Reason: "trophy league " + lk})
	if gerr != nil {
		status := "pending"
		if tries+1 >= cfg.Prizes.MaxTries {
			status = "failed"
		}
		_, _ = tx.ExecContext(ctx, `UPDATE trophy_prize SET tries=tries+1, last_error=$3, status=$4 WHERE user_id=$1 AND claim_key=$2`, uid, key, truncate(gerr.Error(), 500), status)
		_ = tx.Commit()
		return false
	}
	paid, _ := json.Marshal(res.Prizes)
	if _, err := tx.ExecContext(ctx, `UPDATE trophy_prize SET status='paid', tries=tries+1, paid=$3, paid_at=now(), last_error='' WHERE user_id=$1 AND claim_key=$2`, uid, key, paid); err != nil {
		return false
	}
	return tx.Commit() == nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

type seasonRow struct {
	Number    int        `json:"number"`
	StartedAt time.Time  `json:"startedAt"`
	EndsAt    *time.Time `json:"endsAt"`
	EndedAt   *time.Time `json:"endedAt"`
	Reset     int        `json:"playersReset"`
	Mode      string     `json:"mode"`
	Actor     string     `json:"actor"`
}

func nextRun(schedule string, now time.Time) *time.Time {
	if schedule == "" {
		return nil
	}
	e, err := cronexpr.Parse(schedule)
	if err != nil {
		return nil
	}
	t := e.Next(now.UTC())
	if t.IsZero() {
		return nil
	}
	return &t
}

func (s *service) currentSeason(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) (*seasonRow, error) {
	var r seasonRow
	err := q.QueryRowContext(ctx, `SELECT number, started_at, ends_at, ended_at, players_reset, mode, actor FROM trophy_season WHERE ended_at IS NULL ORDER BY number DESC LIMIT 1`).
		Scan(&r.Number, &r.StartedAt, &r.EndsAt, &r.EndedAt, &r.Reset, &r.Mode, &r.Actor)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &r, err
}

func (s *service) seasonTick(ctx context.Context, actor string) error {
	cfg := s.settings(ctx)
	if !cfg.Season.Enabled && actor == "" {
		return nil
	}
	now := s.now().UTC()
	cur, err := s.currentSeason(ctx, s.db)
	if err != nil {
		return err
	}
	var ends *time.Time
	if cfg.Season.Enabled {
		ends = nextRun(cfg.Season.Schedule, now)
	}
	if cur == nil {
		_, err := s.db.ExecContext(ctx, `INSERT INTO trophy_season (number, started_at, ends_at) VALUES (1, $1, $2) ON CONFLICT DO NOTHING`, now, ends)
		if err != nil || actor == "" {
			return err
		}
		cur, err = s.currentSeason(ctx, s.db)
		if err != nil || cur == nil {
			return err
		}
	}
	if actor == "" && (cur.EndsAt == nil || now.Before(*cur.EndsAt)) {
		if cur.EndsAt == nil && ends != nil {
			_, _ = s.db.ExecContext(ctx, `UPDATE trophy_season SET ends_at=$2 WHERE number=$1 AND ended_at IS NULL AND ends_at IS NULL`, cur.Number, ends)
		}
		return nil
	}
	_, err = s.resetSeason(ctx, cfg, cur.Number, actor, ends)
	return err
}

func (s *service) resetSeason(ctx context.Context, cfg settings.All, number int, actor string, nextEnds *time.Time) (int, error) {
	l := s.ladder(ctx)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var open bool
	if err := tx.QueryRowContext(ctx, `SELECT ended_at IS NULL FROM trophy_season WHERE number=$1 FOR UPDATE`, number).Scan(&open); err != nil || !open {
		return 0, err
	}
	type step struct {
		cond string
		expr string
		args []any
	}
	steps := []step{}
	switch cfg.Season.Mode {
	case settings.ResetZero:
		steps = append(steps, step{cond: `trophies > 0`, expr: `0`})
	case settings.ResetLeague:
		for i, x := range l {
			if x.ResetTo == nil {
				continue
			}
			upper := 1 << 30
			if i+1 < len(l) {
				upper = l[i+1].MinTrophies
			}
			steps = append(steps, step{cond: `trophies >= $2 AND trophies < $3 AND trophies > $4`, expr: `$4::int`, args: []any{x.MinTrophies, upper, *x.ResetTo}})
		}
	default:
		steps = append(steps, step{cond: `trophies > $2`, expr: `$2::int + (trophies - $2::int) * $3::int / 100`, args: []any{cfg.Season.Above, cfg.Season.KeepPercent}})
	}
	total := 0
	source := fmt.Sprintf("season:%d", number)
	for _, st := range steps {
		args := append([]any{source}, st.args...)
		args = append(args, actor)
		res, err := tx.ExecContext(ctx, `WITH c AS (SELECT user_id, trophies AS before, (`+st.expr+`) AS after, league_key FROM trophy_player WHERE `+st.cond+` FOR UPDATE),
			u AS (UPDATE trophy_player p SET trophies=c.after, updated_at=now() FROM c WHERE p.user_id=c.user_id RETURNING p.user_id)
			INSERT INTO trophy_log (user_id, source_id, reason, delta, before, after, league_before, actor)
			SELECT c.user_id, $1, 'season', c.after - c.before, c.before, c.after, c.league_key, $`+fmt.Sprint(len(args))+`::text FROM c ON CONFLICT DO NOTHING`, args...)
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		total += int(n)
	}
	for i, x := range l {
		upper := 1 << 30
		if i+1 < len(l) {
			upper = l[i+1].MinTrophies
		}
		if _, err := tx.ExecContext(ctx, `UPDATE trophy_player SET league_key=$1 WHERE trophies >= $2 AND trophies < $3 AND league_key <> $1`, x.Key, x.MinTrophies, upper); err != nil {
			return 0, err
		}
	}
	if len(l) > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE trophy_player SET league_key='' WHERE trophies < $1 AND league_key <> ''`, l[0].MinTrophies); err != nil {
			return 0, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE trophy_log t SET league_after=p.league_key FROM trophy_player p WHERE t.source_id=$1 AND t.user_id=p.user_id`, source); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE trophy_player SET season_best=trophies`); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE trophy_season SET ended_at=now(), players_reset=$2, mode=$3, actor=$4 WHERE number=$1`, number, total, cfg.Season.Mode, actor); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO trophy_season (number, started_at, ends_at) VALUES ($1, now(), $2)`, number+1, nextEnds); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	s.logger.Info("Trophy season %d closed: %d players reset (%s)", number, total, cfg.Season.Mode)
	return total, nil
}

func (s *service) beforeMatchmakerAdd(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
	m := in.GetMatchmakerAdd()
	if m == nil {
		return in, nil
	}
	cfg := s.settings(ctx)
	if !cfg.Enabled || !cfg.Matchmaking.Enabled {
		return in, nil
	}
	game := strings.TrimSpace(m.StringProperties["game"])
	if len(cfg.Matchmaking.Games) > 0 {
		ok := false
		for _, g := range cfg.Matchmaking.Games {
			if g == game {
				ok = true
			}
		}
		if !ok {
			return in, nil
		}
	}
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if userID == "" {
		return in, nil
	}
	t := cfg.Amounts.Start
	_ = s.db.QueryRowContext(ctx, `SELECT trophies FROM trophy_player WHERE user_id::text=$1`, userID).Scan(&t)
	if m.NumericProperties == nil {
		m.NumericProperties = map[string]float64{}
	}
	p := cfg.Matchmaking.Property
	m.NumericProperties[p] = float64(t)
	lo := t - cfg.Matchmaking.Range
	if lo < 0 {
		lo = 0
	}
	clause := fmt.Sprintf("+properties.%s:>=%d +properties.%s:<=%d", p, lo, p, t+cfg.Matchmaking.Range)
	if q := strings.TrimSpace(m.Query); q == "" || q == "*" {
		m.Query = clause
	} else {
		m.Query = q + " " + clause
	}
	return in, nil
}
