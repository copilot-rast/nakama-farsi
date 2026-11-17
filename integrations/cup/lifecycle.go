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
	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/cronexpr"
	"github.com/heroiclabs/nakama/v3/internal/cuphooks"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

var errSkip = errors.New("cup busy")

func (s *service) withCup(ctx context.Context, id string, skipLocked bool, fn func(tx *sql.Tx, b *bracket) error) (*bracket, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	lock := "FOR UPDATE"
	if skipLocked {
		lock = "FOR UPDATE SKIP LOCKED"
	}
	c, err := loadCup(ctx, tx, id, lock)
	if errors.Is(err, sql.ErrNoRows) {
		if skipLocked {
			return nil, errSkip
		}
		return nil, apierr.NotFound("CUP_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	b, err := s.loadBracket(ctx, tx, c)
	if err != nil {
		return nil, err
	}
	if err := fn(tx, b); err != nil {
		return nil, err
	}
	if err := b.save(ctx, tx); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.after(ctx, b)
	return b, nil
}

func (s *service) after(ctx context.Context, b *bracket) {
	for _, n := range b.fx.notes {
		s.notify(ctx, n)
	}
	s.publish(ctx, b.fx.events)
	for _, job := range b.fx.launches {
		s.runLaunch(ctx, job)
	}
	if b.fx.finished {
		s.payPending(ctx, b.cup.ID, 200)
	}
	if b.cup.Status == "cancelled" {
		s.refundCup(ctx, b.cup.ID)
	}
	b.fx = &effects{}
}

func (s *service) runLaunch(ctx context.Context, job launchJob) {
	var out cuphooks.Launched
	var err error
	if job.launch.Game == "" {
		out = cuphooks.Launched{GameMatchID: "cup:" + job.gameID, Info: map[string]any{"external": true}}
	} else {
		out, err = s.launch(ctx, job.launch)
		if err == nil && out.GameMatchID == "" {
			err = errors.New("the game returned no match id")
		}
	}
	if err != nil {
		s.logger.Warn("Cup game %s for match %s not launched: %v", job.gameID, job.matchID, err)
		_, _ = s.withCup(ctx, job.launch.CupID, false, func(tx *sql.Tx, b *bracket) error {
			m := b.matchByID(job.matchID)
			for _, g := range b.games[job.matchID] {
				if g.ID == job.gameID && g.Status == "launching" && m != nil {
					b.voidGame(m, g, "launch_failed")
				}
			}
			return nil
		})
		return
	}
	res, err := s.db.ExecContext(ctx, `UPDATE cup_game SET game_match_id=$2, status='playing' WHERE id=$1::uuid AND status='launching'`, job.gameID, out.GameMatchID)
	if err != nil {
		s.logger.Warn("Cup game %s not stored: %v", job.gameID, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE cup_match SET game_match_id=$2, updated_at=now() WHERE id=$1::uuid AND games=$3 AND status='playing'`, job.matchID, out.GameMatchID, job.launch.GameNo)
	info := out.Info
	if info == nil {
		info = map[string]any{}
	}
	s.notify(ctx, note{users: job.launch.Players, subject: "cup_match_ready", persistent: false, content: map[string]any{"cupId": job.launch.CupID, "matchId": job.matchID,
		"gameNo": job.launch.GameNo, "game": job.launch.Game, "gameMatchId": out.GameMatchID, "players": job.launch.Players, "info": info}})
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

func (s *service) createFromTemplate(ctx context.Context, t *templateRow, startsAt time.Time, actor string) (string, error) {
	signup := startsAt.Add(-time.Duration(t.SignupMinutes) * time.Minute)
	if now := s.now().UTC(); signup.Before(now) && actor != "schedule" {
		signup = now
	}
	var announce any
	if t.AnnounceMinutes > 0 {
		announce = signup.Add(-time.Duration(t.AnnounceMinutes) * time.Minute)
	}
	status := "upcoming"
	if !s.now().Before(signup) {
		status = "signup"
	}
	id := uuid.Must(uuid.NewV4()).String()
	var got string
	err := s.db.QueryRowContext(ctx, `INSERT INTO cup (id, template_id, template_key, name, description, icon_url, status, announce_at, signup_at, starts_at, rules, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT (template_id, starts_at) WHERE template_id IS NOT NULL DO UPDATE SET template_key=EXCLUDED.template_key RETURNING id::text`,
		id, t.ID, t.Key, jsonOf(t.Name), jsonOf(t.Description), t.IconURL, status, announce, signup.UTC(), startsAt.UTC(), jsonOf(t.Rules), actor).Scan(&got)
	if err == nil {
		_, _ = s.db.ExecContext(ctx, `UPDATE cup_template SET last_starts_at=GREATEST(COALESCE(last_starts_at, $2), $2) WHERE id=$1::uuid`, t.ID, startsAt.UTC())
	}
	return got, err
}

func (s *service) scheduleTemplates(ctx context.Context) error {
	list, err := s.templates(ctx, true)
	if err != nil {
		return err
	}
	now := s.now()
	for _, t := range list {
		next := cronNext(t.Schedule, now)
		if next == nil {
			continue
		}
		var open bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM cup WHERE template_id=$1::uuid AND status IN ('draft','upcoming','signup') AND starts_at <> $2)`, t.ID, next.UTC()).Scan(&open)
		if open {
			continue
		}
		if _, err := s.createFromTemplate(ctx, t, *next, "schedule"); err != nil {
			s.logger.Warn("Cup template %s not scheduled: %v", t.Key, err)
		}
	}
	return nil
}

func (s *service) ids(ctx context.Context, q string, args ...any) []string {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out
}

func (s *service) advanceCups(ctx context.Context) {
	now := s.now().UTC()
	_, _ = s.db.ExecContext(ctx, `UPDATE cup SET status='signup', updated_at=now() WHERE status='upcoming' AND signup_at <= $1`, now)
	for _, id := range s.ids(ctx, `SELECT id::text FROM cup WHERE status='signup' AND starts_at <= $1 ORDER BY starts_at LIMIT 20`, now) {
		if _, err := s.withCup(ctx, id, true, func(tx *sql.Tx, b *bracket) error {
			if b.cup.Status != "signup" || b.now.Before(b.cup.StartsAt) {
				return nil
			}
			return s.startLocked(ctx, tx, b)
		}); err != nil && !errors.Is(err, errSkip) {
			s.logger.Warn("Cup %s not started: %v", id, err)
		}
	}
	for _, id := range s.ids(ctx, `SELECT id::text FROM cup WHERE status='running' ORDER BY starts_at LIMIT 200`) {
		if _, err := s.withCup(ctx, id, true, func(tx *sql.Tx, b *bracket) error {
			if b.cup.Status != "running" {
				return nil
			}
			for _, m := range b.matches {
				b.tickMatch(m, s.rnd)
			}
			b.advance()
			return nil
		}); err != nil && !errors.Is(err, errSkip) {
			s.logger.Warn("Cup %s not advanced: %v", id, err)
		}
	}
}

func (s *service) startLocked(ctx context.Context, tx *sql.Tx, b *bracket) error {
	c := b.cup
	n := 0
	for _, p := range b.order {
		if p.Status == "registered" {
			n++
		}
	}
	min := c.Rules.MinPlayers
	if min < 2 {
		min = 2
	}
	if n < min || (c.Rules.Underfill == settings.UnderfillCancel && n < c.Rules.Capacity) {
		s.cancelLocked(b, "underfill")
		return nil
	}
	if c.Rules.Seeding == settings.SeedTrophies {
		s.readTrophies(ctx, tx, b)
	}
	var fees int64
	for _, p := range b.order {
		if p.Status == "registered" && p.FeeState == "paid" && p.Fee.CurrencyKey == c.Rules.Entry.Fee.CurrencyKey {
			fees += p.Fee.Amount
		}
	}
	c.Pool = poolValue{CurrencyKey: c.Rules.Entry.Fee.CurrencyKey, Fees: fees, Amount: poolShare(fees, c.Rules.Prizes.Pool.Percent) + c.Rules.Prizes.Pool.Base}
	if c.Pool.CurrencyKey == "" {
		c.Pool.Amount = 0
	}
	c.Status, c.StartedAt = "running", ptr(b.now)
	if err := b.build(s.rnd); err != nil {
		return err
	}
	b.advance()
	return nil
}

func (s *service) readTrophies(ctx context.Context, tx *sql.Tx, b *bracket) {
	var exists bool
	if tx.QueryRowContext(ctx, `SELECT to_regclass('trophy_player') IS NOT NULL`).Scan(&exists) != nil || !exists {
		return
	}
	ids := []string{}
	for _, p := range b.order {
		if p.Status == "registered" {
			ids = append(ids, p.UserID)
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id::text, trophies FROM trophy_player WHERE user_id::text = ANY($1)`, ids)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var n int
		if rows.Scan(&id, &n) == nil {
			if p := b.players[id]; p != nil {
				p.Trophies = n
				b.touchPlayer(id)
			}
		}
	}
}

func (s *service) cancelLocked(b *bracket, reason string) {
	c := b.cup
	c.Status, c.CancelReason, c.FinishedAt = "cancelled", truncate(reason, 200), ptr(b.now)
	for _, m := range b.matches {
		if !done(m.Status) {
			m.Status, m.Reason, m.FinishedAt = stVoid, "cancelled", ptr(b.now)
			b.touch(m)
		}
	}
	users := []string{}
	for _, p := range b.order {
		if p.Status == "registered" || p.Status == "playing" {
			users = append(users, p.UserID)
		}
	}
	b.notify(users, "cup_cancelled", map[string]any{"reason": c.CancelReason}, true, true)
}

func (s *service) refundPlayer(ctx context.Context, cupID string, p *playerRow, percent int) error {
	if p.FeeState != "paid" {
		return nil
	}
	amount := p.Fee.Amount * int64(percent) / 100
	if amount > 0 && p.Fee.CurrencyKey != "" {
		meta := map[string]any{"module": owner, "sourceType": sourceType, "sourceId": cupID, "reason": "cup_refund"}
		if err := s.credit(ctx, p.UserID, "cup:refund:"+p.Fee.Grant, map[string]int64{p.Fee.CurrencyKey: amount}, meta); err != nil {
			return err
		}
	}
	p.Fee.Refunded = amount
	_, err := s.db.ExecContext(ctx, `UPDATE cup_player SET fee_state='refunded', fee=$3, updated_at=now() WHERE cup_id=$1::uuid AND user_id=$2::uuid AND fee_state='paid'`,
		cupID, p.UserID, jsonOf(p.Fee))
	return err
}

func (s *service) refundCup(ctx context.Context, cupID string) {
	list, err := loadPlayers(ctx, s.db, cupID)
	if err != nil {
		return
	}
	for _, p := range list {
		if p.FeeState == "paid" && p.Status != "withdrawn" {
			if err := s.refundPlayer(ctx, cupID, p, 100); err != nil {
				s.logger.Warn("Cup %s refund for %s failed, retried later: %v", cupID, p.UserID, err)
			}
		}
	}
}

func (s *service) retryRefunds(ctx context.Context, limit int) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.cup_id::text, c.rules::text, `+prefixed("p.", playerCols)+` FROM cup_player p JOIN cup c ON c.id=p.cup_id
		WHERE p.fee_state='paid' AND (p.status IN ('withdrawn','removed') OR c.status='cancelled') ORDER BY p.updated_at LIMIT $1`, limit)
	if err != nil {
		return
	}
	type job struct {
		cup     string
		percent int
		p       *playerRow
	}
	jobs := []job{}
	for rows.Next() {
		var cupID, rawRules string
		p := &playerRow{}
		var fee string
		if rows.Scan(&cupID, &rawRules, &p.UserID, &p.Status, &p.Seed, &p.Trophies, &p.Place, &p.Reached, &fee, &p.FeeState, &p.AddedBy, &p.JoinedAt) != nil {
			continue
		}
		_ = json.Unmarshal([]byte(fee), &p.Fee)
		percent := 100
		if p.Status == "withdrawn" {
			r := settings.DefaultRules()
			_ = json.Unmarshal([]byte(rawRules), &r)
			percent = r.Entry.WithdrawRefundPercent
		}
		jobs = append(jobs, job{cupID, percent, p})
	}
	rows.Close()
	for _, j := range jobs {
		if err := s.refundPlayer(ctx, j.cup, j.p, j.percent); err != nil {
			s.logger.Debug("Cup %s refund for %s still failing: %v", j.cup, j.p.UserID, err)
		}
	}
}

func prefixed(prefix, cols string) string {
	parts := strings.Split(cols, ",")
	for i, part := range parts {
		parts[i] = prefix + strings.TrimSpace(part)
	}
	return strings.Join(parts, ", ")
}

type payout struct {
	user, kind, prize string
	amount            int64
}

func (s *service) planPayouts(b *bracket) []payout {
	c := b.cup
	r := c.Rules.Prizes
	third := b.third() != nil
	out := []payout{}
	add := func(user, kind, prize string, share int) {
		var amount int64
		if share > 0 && c.Pool.CurrencyKey != "" {
			amount = poolShare(c.Pool.Amount, share)
		}
		if prize != "" || amount > 0 {
			out = append(out, payout{user: user, kind: kind, prize: prize, amount: amount})
		}
	}
	for _, p := range b.order {
		if p.Seed == 0 || p.Status == "disqualified" {
			continue
		}
		switch {
		case p.Place == 1:
			add(p.UserID, "champion", r.Champion, r.Pool.Shares.Champion)
		case p.Place == 2:
			add(p.UserID, "runner_up", r.RunnerUp, r.Pool.Shares.RunnerUp)
		case p.Place == 3 && third:
			add(p.UserID, "third", r.Third, r.Pool.Shares.Third)
		case p.Place == 4 && third:
			add(p.UserID, "fourth", r.Fourth, r.Pool.Shares.Fourth)
		case p.Place == 3:
			add(p.UserID, "semi", r.Semi, r.Pool.Shares.Semi)
		}
		for _, x := range r.Reached {
			if p.Reached >= x.Round {
				add(p.UserID, "reached_"+strconv.Itoa(x.Round), x.PrizeKey, 0)
			}
		}
	}
	return out
}

func (s *service) queuePayouts(ctx context.Context, tx *sql.Tx, b *bracket) error {
	for _, p := range s.planPayouts(b) {
		amounts := map[string]int64{}
		if p.amount > 0 {
			amounts[b.cup.Pool.CurrencyKey] = p.amount
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO cup_payout (id, cup_id, user_id, kind, prize_key, amounts) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (cup_id, user_id, kind) DO NOTHING`,
			uuid.Must(uuid.NewV4()).String(), b.cup.ID, p.user, p.kind, p.prize, jsonOf(amounts)); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) payPending(ctx context.Context, cupID string, limit int) {
	for i := 0; i < limit; i++ {
		if !s.payOne(ctx, cupID) {
			return
		}
	}
}

func (s *service) payOne(ctx context.Context, cupID string) bool {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false
	}
	defer func() { _ = tx.Rollback() }()
	q := `SELECT id::text, cup_id::text, user_id::text, kind, prize_key, amounts::text, tries FROM cup_payout WHERE status='pending'`
	args := []any{}
	if cupID != "" {
		q += ` AND cup_id::text=$1`
		args = append(args, cupID)
	}
	var id, cup, uid, kind, prize, rawAmounts string
	var tries int
	if err := tx.QueryRowContext(ctx, q+` ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED`, args...).Scan(&id, &cup, &uid, &kind, &prize, &rawAmounts, &tries); err != nil {
		return false
	}
	amounts := map[string]int64{}
	_ = json.Unmarshal([]byte(rawAmounts), &amounts)
	var paid []map[string]any
	var gerr error
	if prize != "" {
		var res economyhooks.PrizeResult
		res, gerr = s.prize(ctx, economyhooks.PrizeGrant{UserID: uid, PrizeKey: prize, SourceType: sourceType, SourceID: cup + ":" + kind, Reason: "cup " + kind})
		paid = res.Prizes
	}
	if gerr == nil && len(amounts) > 0 {
		gerr = s.credit(ctx, uid, "cup:pool:"+id, amounts, map[string]any{"module": owner, "sourceType": sourceType, "sourceId": cup, "reason": "cup_pool_" + kind})
	}
	if gerr != nil {
		status := "pending"
		if tries+1 >= 8 {
			status = "failed"
		}
		_, _ = tx.ExecContext(ctx, `UPDATE cup_payout SET tries=tries+1, last_error=$2, status=$3 WHERE id::text=$1`, id, truncate(gerr.Error(), 500), status)
		_ = tx.Commit()
		return false
	}
	if paid == nil {
		paid = []map[string]any{}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE cup_payout SET status='paid', tries=tries+1, paid=$2, paid_at=now(), last_error='' WHERE id::text=$1`, id, jsonOf(paid)); err != nil {
		return false
	}
	if tx.Commit() != nil {
		return false
	}
	s.notify(ctx, note{users: []string{uid}, subject: "cup_prize", persistent: true, inapp: true, content: map[string]any{"cupId": cup, "kind": kind, "prizeKey": prize, "amounts": amounts, "prizes": paid}})
	return true
}
