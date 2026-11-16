package main

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	duelNotifyCode    = 1410
	duelNotifySubject = "quiz_duel"
)

type duelRules struct {
	Rounds          int                 `json:"rounds"`
	PerRound        int                 `json:"questionsPerRound"`
	Choices         int                 `json:"categoryChoices"`
	QuestionSeconds int                 `json:"questionSeconds"`
	GraceSeconds    int                 `json:"graceSeconds"`
	Source          string              `json:"questionSource"`
	AvoidDays       int                 `json:"avoidRepeatDays"`
	Shuffle         bool                `json:"shuffleOptions"`
	ServeNext       bool                `json:"serveNextOnAnswer"`
	TurnHours       int                 `json:"turnHours"`
	TimeoutResult   string              `json:"timeoutResult"`
	WaitingHours    int                 `json:"waitingHours"`
	EntryFee        settings.Price      `json:"entryFee"`
	WinPrize        string              `json:"winPrize"`
	LosePrize       string              `json:"losePrize"`
	DrawPrize       string              `json:"drawPrize"`
	PerCorrect      settings.Price      `json:"perCorrect"`
	Events          settings.DuelEvents `json:"events"`
	Notify          bool                `json:"notify"`
}

func rulesOf(d settings.Duel) duelRules {
	return duelRules{Rounds: d.Rounds, PerRound: d.QuestionsPerRound, Choices: d.CategoryChoices, QuestionSeconds: d.QuestionSeconds, GraceSeconds: d.GraceSeconds,
		Source: d.QuestionSource, AvoidDays: d.AvoidRepeatDays, Shuffle: d.ShuffleOptions, ServeNext: d.ServeNextOnAnswer, TurnHours: d.TurnHours,
		TimeoutResult: d.TimeoutResult, WaitingHours: d.WaitingHours, EntryFee: d.EntryFee, WinPrize: d.WinPrize, LosePrize: d.LosePrize, DrawPrize: d.DrawPrize,
		PerCorrect: d.PerCorrect, Events: d.Events, Notify: d.Notify}
}

type duelRow struct {
	ID, Mode, Status, A, B string
	Rules                  duelRules
	Turn                   string
	TurnStarted            *time.Time
	TurnDeadline           *time.Time
	Expires                *time.Time
	ScoreA, ScoreB         int
	Winner, Reason         string
	FeeA, FeeB             string
	PaidA, PaidB           bool
	PayError               string
	Created                time.Time
	Accepted, Finished     *time.Time
	Updated                time.Time
}

const duelCols = `id::text, mode, status, player_a::text, COALESCE(player_b::text,''), rules, COALESCE(turn_user::text,''), turn_started_at, turn_deadline, expires_at,
	score_a, score_b, COALESCE(winner::text,''), reason, fee_a, fee_b, paid_a, paid_b, pay_error, created_at, accepted_at, finished_at, updated_at`

func scanDuel(sc scanner) (*duelRow, error) {
	d := &duelRow{}
	var rules []byte
	var ts, td, ex, ac, fi sql.NullTime
	if err := sc.Scan(&d.ID, &d.Mode, &d.Status, &d.A, &d.B, &rules, &d.Turn, &ts, &td, &ex, &d.ScoreA, &d.ScoreB, &d.Winner, &d.Reason,
		&d.FeeA, &d.FeeB, &d.PaidA, &d.PaidB, &d.PayError, &d.Created, &ac, &fi, &d.Updated); err != nil {
		return nil, err
	}
	unmarshal(rules, &d.Rules)
	for _, p := range []struct {
		n   sql.NullTime
		dst **time.Time
	}{{ts, &d.TurnStarted}, {td, &d.TurnDeadline}, {ex, &d.Expires}, {ac, &d.Accepted}, {fi, &d.Finished}} {
		if p.n.Valid {
			v := p.n.Time
			*p.dst = &v
		}
	}
	return d, nil
}

func (d *duelRow) other(user string) string {
	if user == d.A {
		return d.B
	}
	return d.A
}

func (d *duelRow) has(user string) bool { return user != "" && (user == d.A || user == d.B) }

func (d *duelRow) picker(round int) string {
	if round%2 == 1 {
		return d.A
	}
	return d.B
}

func (d *duelRow) open() bool {
	return d.Status == "waiting" || d.Status == "invited" || d.Status == "active"
}

type duelRound struct {
	No        int
	Picker    string
	Offers    []string
	Category  string
	Questions []picked
	PickedAt  *time.Time
}

type duelAnswer struct {
	UserID, QuestionID string
	Round, Idx         int
	Served             time.Time
	Answered           *time.Time
	Option             *int
	Outcome            string
	Seconds            float64
}

type duelState struct {
	d       *duelRow
	rounds  map[int]*duelRound
	answers map[string]map[string]*duelAnswer
}

func loadDuel(ctx context.Context, q queryer, id string, lock bool) (*duelRow, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, apierr.NotFound("QUIZ_DUEL_NOT_FOUND").WithField("duelId")
	}
	sfx := ""
	if lock {
		sfx = " FOR UPDATE"
	}
	d, err := scanDuel(q.QueryRowContext(ctx, `SELECT `+duelCols+` FROM quiz_duel WHERE id=$1`+sfx, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("QUIZ_DUEL_NOT_FOUND").WithField("duelId")
	}
	return d, err
}

func loadState(ctx context.Context, q queryer, d *duelRow) (*duelState, error) {
	st := &duelState{d: d, rounds: map[int]*duelRound{}, answers: map[string]map[string]*duelAnswer{}}
	rows, err := q.QueryContext(ctx, `SELECT round_no, picker::text, array_to_json(offers)::text, category, questions, picked_at FROM quiz_duel_round WHERE duel_id=$1`, d.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		r := &duelRound{}
		var offers, qs []byte
		var at sql.NullTime
		if err := rows.Scan(&r.No, &r.Picker, &offers, &r.Category, &qs, &at); err != nil {
			rows.Close()
			return nil, err
		}
		unmarshal(offers, &r.Offers)
		unmarshal(qs, &r.Questions)
		if at.Valid {
			v := at.Time
			r.PickedAt = &v
		}
		st.rounds[r.No] = r
	}
	rows.Close()
	rows, err = q.QueryContext(ctx, `SELECT user_id::text, question_id::text, round_no, idx, served_at, answered_at, option_id, outcome, COALESCE(seconds,0)::float8 FROM quiz_duel_answer WHERE duel_id=$1`, d.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		a := &duelAnswer{}
		var at sql.NullTime
		var opt sql.NullInt64
		if err := rows.Scan(&a.UserID, &a.QuestionID, &a.Round, &a.Idx, &a.Served, &at, &opt, &a.Outcome, &a.Seconds); err != nil {
			return nil, err
		}
		if at.Valid {
			v := at.Time
			a.Answered = &v
		}
		if opt.Valid {
			v := int(opt.Int64)
			a.Option = &v
		}
		if st.answers[a.UserID] == nil {
			st.answers[a.UserID] = map[string]*duelAnswer{}
		}
		st.answers[a.UserID][a.QuestionID] = a
	}
	return st, rows.Err()
}

func (st *duelState) answer(user, qid string) *duelAnswer {
	if m := st.answers[user]; m != nil {
		return m[qid]
	}
	return nil
}

func (st *duelState) roundDone(user string, r int) bool {
	rr := st.rounds[r]
	if rr == nil || rr.Category == "" || len(rr.Questions) == 0 {
		return false
	}
	for _, p := range rr.Questions {
		a := st.answer(user, p.ID)
		if a == nil || a.Outcome == "pending" {
			return false
		}
	}
	return true
}

func (st *duelState) correct(user string, upTo func(int) bool) int {
	n := 0
	for _, a := range st.answers[user] {
		if a.Outcome == "correct" && (upTo == nil || upTo(a.Round)) {
			n++
		}
	}
	return n
}

type duelStep struct {
	Kind  string
	Round int
}

func (st *duelState) step(user string) duelStep {
	if user == "" {
		return duelStep{}
	}
	for r := 1; r <= st.d.Rules.Rounds; r++ {
		if st.roundDone(user, r) {
			continue
		}
		rr := st.rounds[r]
		if rr != nil && rr.Category != "" {
			return duelStep{Kind: "answer", Round: r}
		}
		if st.d.picker(r) == user {
			return duelStep{Kind: "pick", Round: r}
		}
		return duelStep{Kind: "wait", Round: r}
	}
	return duelStep{Kind: "done", Round: st.d.Rules.Rounds}
}

func (st *duelState) playable(user string) bool {
	k := st.step(user).Kind
	return k == "pick" || k == "answer"
}

type duelNote struct {
	to    string
	event string
	from  string
	keep  bool
}

type duelTx struct {
	notes    []duelNote
	finished bool
	refund   bool
}

func (x *duelTx) note(to, event, from string, keep bool) {
	if to != "" {
		x.notes = append(x.notes, duelNote{to: to, event: event, from: from, keep: keep})
	}
}

func (s *service) eligibleSQL(r duelRules) string {
	w := `q.is_active AND q.type='options' AND jsonb_array_length(q.options) >= 2`
	if r.Source == settings.SourceBank {
		w += ` AND q.pack_id IS NULL`
	}
	return w
}

func (s *service) ensureOffers(ctx context.Context, tx *sql.Tx, st *duelState, round int) error {
	if st.rounds[round] != nil {
		return nil
	}
	d := st.d
	used := map[string]bool{}
	for _, rr := range st.rounds {
		if rr.Category != "" {
			used[rr.Category] = true
		}
	}
	usedIDs := []string{}
	for _, rr := range st.rounds {
		for _, p := range rr.Questions {
			usedIDs = append(usedIDs, p.ID)
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT c.key FROM quiz_category c WHERE c.is_active AND c.duel
		AND (SELECT count(*) FROM quiz_question q WHERE q.category=c.key AND `+s.eligibleSQL(d.Rules)+` AND NOT (q.id::text = ANY($2::text[]))) >= $1 ORDER BY c.sort, c.key`, d.Rules.PerRound, usedIDs)
	if err != nil {
		return err
	}
	var fresh, old []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			rows.Close()
			return err
		}
		if used[k] {
			old = append(old, k)
		} else {
			fresh = append(fresh, k)
		}
	}
	rows.Close()
	if len(fresh)+len(old) == 0 {
		return apierr.Unavailable("QUIZ_DUEL_NO_CATEGORIES")
	}
	rng := rand.New(rand.NewSource(s.now().UnixNano()))
	rng.Shuffle(len(fresh), func(i, j int) { fresh[i], fresh[j] = fresh[j], fresh[i] })
	rng.Shuffle(len(old), func(i, j int) { old[i], old[j] = old[j], old[i] })
	offers := append(fresh, old...)
	if len(offers) > d.Rules.Choices {
		offers = offers[:d.Rules.Choices]
	}
	picker := d.picker(round)
	if picker == "" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_duel_round (duel_id, round_no, picker, offers) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, d.ID, round, picker, offers); err != nil {
		return err
	}
	st.rounds[round] = &duelRound{No: round, Picker: picker, Offers: offers}
	return nil
}

func (s *service) drawRound(ctx context.Context, tx *sql.Tx, st *duelState, category string) ([]picked, error) {
	d := st.d
	usedIDs := []string{}
	for _, rr := range st.rounds {
		for _, p := range rr.Questions {
			usedIDs = append(usedIDs, p.ID)
		}
	}
	players := []string{d.A}
	if d.B != "" {
		players = append(players, d.B)
	}
	rows, err := tx.QueryContext(ctx, `SELECT q.id::text, q.options, EXISTS (SELECT 1 FROM quiz_duel_answer a WHERE a.question_id=q.id AND a.user_id::text = ANY($3::text[]) AND a.served_at > now() - make_interval(days => $4))
		FROM quiz_question q WHERE q.category=$1 AND `+s.eligibleSQL(d.Rules)+` AND NOT (q.id::text = ANY($2::text[]))`, category, usedIDs, players, d.Rules.AvoidDays)
	if err != nil {
		return nil, err
	}
	type cand struct {
		id   string
		opts []Option
	}
	var fresh, seen []cand
	for rows.Next() {
		var c cand
		var raw []byte
		var recent bool
		if err := rows.Scan(&c.id, &raw, &recent); err != nil {
			rows.Close()
			return nil, err
		}
		unmarshal(raw, &c.opts)
		if recent && d.Rules.AvoidDays > 0 {
			seen = append(seen, c)
		} else {
			fresh = append(fresh, c)
		}
	}
	rows.Close()
	rng := rand.New(rand.NewSource(s.now().UnixNano()))
	rng.Shuffle(len(fresh), func(i, j int) { fresh[i], fresh[j] = fresh[j], fresh[i] })
	rng.Shuffle(len(seen), func(i, j int) { seen[i], seen[j] = seen[j], seen[i] })
	all := append(fresh, seen...)
	if len(all) < d.Rules.PerRound {
		return nil, apierr.Conflict("QUIZ_DUEL_CATEGORY_EMPTY").WithField("categoryKey").With("categoryKey", category)
	}
	out := make([]picked, 0, d.Rules.PerRound)
	for _, c := range all[:d.Rules.PerRound] {
		p := picked{ID: c.id}
		for _, o := range c.opts {
			p.Order = append(p.Order, o.ID)
		}
		if d.Rules.Shuffle {
			rng.Shuffle(len(p.Order), func(i, j int) { p.Order[i], p.Order[j] = p.Order[j], p.Order[i] })
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *service) saveDuel(ctx context.Context, tx *sql.Tx, d *duelRow) error {
	_, err := tx.ExecContext(ctx, `UPDATE quiz_duel SET status=$2, player_b=$3, turn_user=$4, turn_started_at=$5, turn_deadline=$6, expires_at=$7, score_a=$8, score_b=$9,
		winner=$10, reason=$11, fee_a=$12, fee_b=$13, accepted_at=$14, finished_at=$15, updated_at=now() WHERE id=$1`,
		d.ID, d.Status, nullable(d.B), nullable(d.Turn), d.TurnStarted, d.TurnDeadline, d.Expires, d.ScoreA, d.ScoreB, nullable(d.Winner), d.Reason, d.FeeA, d.FeeB, d.Accepted, d.Finished)
	return err
}

func (s *service) startTurn(d *duelRow, user string, now time.Time) {
	d.Turn = user
	t := now
	d.TurnStarted = &t
	if d.Status == "invited" && user == d.B {
		d.TurnDeadline = nil
		return
	}
	dl := now.Add(time.Duration(d.Rules.TurnHours) * time.Hour)
	d.TurnDeadline = &dl
}

func (s *service) advance(ctx context.Context, tx *sql.Tx, st *duelState, x *duelTx, now time.Time) error {
	d := st.d
	d.ScoreA, d.ScoreB = st.correct(d.A, nil), st.correct(d.B, nil)
	if !d.open() {
		return nil
	}
	if d.Turn != "" && st.playable(d.Turn) {
		if sp := st.step(d.Turn); sp.Kind == "pick" {
			return s.ensureOffers(ctx, tx, st, sp.Round)
		}
		return nil
	}
	prev := d.Turn
	other := d.other(prev)
	if prev == "" {
		other = d.A
	}
	if other == "" {
		d.Status, d.Turn, d.TurnStarted, d.TurnDeadline = "waiting", "", nil, nil
		exp := now.Add(time.Duration(d.Rules.WaitingHours) * time.Hour)
		d.Expires = &exp
		return nil
	}
	if st.playable(other) {
		s.startTurn(d, other, now)
		if sp := st.step(other); sp.Kind == "pick" {
			if err := s.ensureOffers(ctx, tx, st, sp.Round); err != nil {
				return err
			}
		}
		if d.Status == "active" {
			x.note(other, "turn", prev, false)
		}
		return nil
	}
	if st.step(d.A).Kind == "done" && st.step(d.B).Kind == "done" {
		s.finishDuel(d, "", nil, x, now)
	}
	return nil
}

func (s *service) finishDuel(d *duelRow, reason string, winner *string, x *duelTx, now time.Time) {
	t := now
	d.Status, d.Finished, d.Turn, d.TurnDeadline, d.Expires = "finished", &t, "", nil, nil
	if winner != nil {
		d.Winner, d.Reason = *winner, reason
	} else {
		switch {
		case d.ScoreA > d.ScoreB:
			d.Winner, d.Reason = d.A, "score"
		case d.ScoreB > d.ScoreA:
			d.Winner, d.Reason = d.B, "score"
		default:
			d.Winner, d.Reason = "", "draw"
		}
		if reason != "" {
			d.Reason = reason
		}
	}
	x.finished = true
	x.note(d.A, "finished", d.B, true)
	x.note(d.B, "finished", d.A, true)
}

func (s *service) endWithout(d *duelRow, status, reason string, x *duelTx, now time.Time) {
	t := now
	d.Status, d.Reason, d.Finished, d.Turn, d.TurnDeadline = status, reason, &t, "", nil
	x.refund = true
}

func (s *service) timeoutDuel(ctx context.Context, tx *sql.Tx, st *duelState, x *duelTx, now time.Time) {
	d := st.d
	if d.Accepted == nil || d.B == "" {
		s.endWithout(d, "expired", "timeout", x, now)
		x.note(d.A, "expired", "", true)
		if d.B != "" {
			x.note(d.B, "expired", d.A, false)
		}
		return
	}
	d.ScoreA, d.ScoreB = st.correct(d.A, nil), st.correct(d.B, nil)
	if d.Rules.TimeoutResult == settings.TimeoutScore {
		s.finishDuel(d, "timeout", nil, x, now)
		return
	}
	w := d.other(d.Turn)
	s.finishDuel(d, "timeout", &w, x, now)
}

func (s *service) flush(ctx context.Context, d *duelRow, x *duelTx) {
	if x.finished {
		s.payDuel(ctx, d.ID)
	}
	if x.refund {
		s.refundDuel(ctx, d.ID)
	}
	if !d.Rules.Notify {
		return
	}
	for _, n := range x.notes {
		content := map[string]any{"event": n.event, "duelId": d.ID, "from": n.from, "status": d.Status}
		if err := s.nk.NotificationSend(ctx, n.to, duelNotifySubject, content, duelNotifyCode, "", n.keep); err != nil {
			s.logger.Debug("Quiz duel notification to %s not sent: %v", n.to, err)
		}
	}
}

func (s *service) mutate(ctx context.Context, id string, fn func(tx *sql.Tx, st *duelState, x *duelTx) error) (*duelState, error) {
	var st *duelState
	var x *duelTx
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		x = &duelTx{}
		d, err := loadDuel(ctx, tx, id, true)
		if err != nil {
			return err
		}
		if st, err = loadState(ctx, tx, d); err != nil {
			return err
		}
		if err := fn(tx, st, x); err != nil {
			return err
		}
		return s.saveDuel(ctx, tx, st.d)
	})
	if err != nil {
		return nil, err
	}
	s.flush(ctx, st.d, x)
	return st, nil
}

func feeGrant(d *duelRow, user string) string { return "quiz:duelfee:" + d.ID + ":" + user }

func (s *service) chargeFee(ctx context.Context, d *duelRow, user string) error {
	fee := d.Rules.EntryFee
	if fee.Amount <= 0 || fee.CurrencyKey == "" {
		return nil
	}
	err := pluginkit.Debit(ctx, s.nk, s.db, user, feeGrant(d, user), map[string]int64{fee.CurrencyKey: fee.Amount},
		map[string]any{"module": "quiz", "sourceType": sourceType, "sourceId": d.ID, "reason": "quiz_duel_fee"})
	var short *pluginkit.ErrInsufficient
	if errors.As(err, &short) {
		return apierr.Conflict("QUIZ_DUEL_NOT_ENOUGH_CURRENCY").With("currencyKey", fee.CurrencyKey).With("amount", fee.Amount)
	}
	return err
}

func (s *service) refundDuel(ctx context.Context, id string) {
	d, err := loadDuel(ctx, s.db, id, false)
	if err != nil || d.open() || d.Status == "finished" {
		return
	}
	fee := d.Rules.EntryFee
	for _, p := range []struct {
		user, state, col string
	}{{d.A, d.FeeA, "fee_a"}, {d.B, d.FeeB, "fee_b"}} {
		if p.user == "" || p.state != "paid" {
			continue
		}
		if fee.Amount > 0 && fee.CurrencyKey != "" {
			grant := "quiz:duelrefund:" + d.ID + ":" + p.user
			applied, err := pluginkit.WalletApplied(ctx, s.db, p.user, grant)
			if err != nil {
				continue
			}
			if !applied {
				md := map[string]any{"grantId": grant, "module": "quiz", "sourceType": sourceType, "sourceId": d.ID, "reason": "quiz_duel_refund"}
				if _, _, err := s.nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: p.user, Changeset: map[string]int64{fee.CurrencyKey: fee.Amount}, Metadata: md}}, true); err != nil {
					s.logger.Warn("Quiz duel refund %s failed: %v", grant, err)
					continue
				}
			}
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE quiz_duel SET `+p.col+`='refunded', updated_at=now() WHERE id=$1 AND `+p.col+`='paid'`, d.ID)
	}
}

func (s *service) payDuel(ctx context.Context, id string) {
	d, err := loadDuel(ctx, s.db, id, false)
	if err != nil || d.Status != "finished" || (d.PaidA && d.PaidB) {
		return
	}
	st, err := loadState(ctx, s.db, d)
	if err != nil {
		return
	}
	failed := ""
	for _, p := range []struct {
		user string
		paid bool
		col  string
	}{{d.A, d.PaidA, "paid_a"}, {d.B, d.PaidB, "paid_b"}} {
		if p.paid {
			continue
		}
		if p.user == "" {
			_, _ = s.db.ExecContext(ctx, `UPDATE quiz_duel SET `+p.col+`=true WHERE id=$1`, d.ID)
			continue
		}
		right := st.correct(p.user, nil)
		outcome := "lose"
		prize := d.Rules.LosePrize
		switch {
		case d.Winner == "":
			outcome, prize = "draw", d.Rules.DrawPrize
		case d.Winner == p.user:
			outcome, prize = "win", d.Rules.WinPrize
		}
		src := "duel:" + d.ID + ":" + p.user
		if pc := d.Rules.PerCorrect; pc.Amount > 0 && pc.CurrencyKey != "" && right > 0 {
			grant := "quiz:duel:" + d.ID + ":" + p.user
			applied, err := pluginkit.WalletApplied(ctx, s.db, p.user, grant)
			if err == nil && !applied {
				md := map[string]any{"grantId": grant, "module": "quiz", "sourceType": sourceType, "sourceId": src, "reason": "quiz_duel_reward"}
				_, _, err = s.nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: p.user, Changeset: map[string]int64{pc.CurrencyKey: pc.Amount * int64(right)}, Metadata: md}}, true)
			}
			if err != nil {
				failed = err.Error()
				continue
			}
		}
		if prize != "" {
			if _, err := s.prize(ctx, economyhooks.PrizeGrant{UserID: p.user, PrizeKey: prize, SourceType: sourceType, SourceID: src, Reason: "quiz_duel_" + outcome}); err != nil {
				failed = err.Error()
				continue
			}
		}
		res, err := s.db.ExecContext(ctx, `UPDATE quiz_duel SET `+p.col+`=true, updated_at=now() WHERE id=$1 AND NOT `+p.col, d.ID)
		if err != nil {
			failed = err.Error()
			continue
		}
		if n, _ := res.RowsAffected(); n == 1 && playerevents.Available() {
			fields := map[string]any{"duelId": d.ID, "outcome": outcome, "score": right, "reason": d.Reason}
			rec := func(key string, v float64) {
				if key == "" {
					return
				}
				if _, err := playerevents.Record(ctx, playerevents.Event{UserID: p.user, Key: key, Value: v, Metadata: fields}); err != nil {
					s.logger.Debug("Quiz duel event %s not recorded: %v", key, err)
				}
			}
			rec(d.Rules.Events.Played, 1)
			if outcome == "win" {
				rec(d.Rules.Events.Won, 1)
			}
			if right > 0 {
				rec(d.Rules.Events.Correct, float64(right))
			}
		}
	}
	if failed != "" {
		_, _ = s.db.ExecContext(ctx, `UPDATE quiz_duel SET pay_tries=pay_tries+1, pay_error=$2 WHERE id=$1`, d.ID, truncateText(failed, 500))
		s.logger.Warn("Quiz duel %s payment failed: %s", d.ID, failed)
	} else {
		_, _ = s.db.ExecContext(ctx, `UPDATE quiz_duel SET pay_error='' WHERE id=$1 AND pay_error<>''`, d.ID)
	}
}

func truncateText(v string, n int) string {
	if len(v) > n {
		return v[:n]
	}
	return v
}

func (s *service) duelWorker(ctx context.Context, limit int) {
	now := s.now()
	ids := func(q string, args ...any) []string {
		out := []string{}
		rows, err := s.db.QueryContext(ctx, q, args...)
		if err != nil {
			return out
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil {
				out = append(out, id)
			}
		}
		return out
	}
	for _, id := range ids(`SELECT id::text FROM quiz_duel WHERE status IN ('waiting','invited') AND expires_at < $1 ORDER BY expires_at LIMIT $2`, now, limit) {
		_, _ = s.mutate(ctx, id, func(tx *sql.Tx, st *duelState, x *duelTx) error {
			d := st.d
			if (d.Status != "waiting" && d.Status != "invited") || d.Expires == nil || d.Expires.After(s.now()) {
				return nil
			}
			s.endWithout(d, "expired", "expired", x, s.now())
			x.note(d.A, "expired", d.B, true)
			x.note(d.B, "expired", d.A, false)
			return nil
		})
	}
	for _, id := range ids(`SELECT id::text FROM quiz_duel WHERE status IN ('active','invited') AND turn_deadline < $1 ORDER BY turn_deadline LIMIT $2`, now, limit) {
		_, _ = s.mutate(ctx, id, func(tx *sql.Tx, st *duelState, x *duelTx) error {
			d := st.d
			if !d.open() || d.TurnDeadline == nil || d.TurnDeadline.After(s.now()) {
				return nil
			}
			s.timeoutDuel(ctx, tx, st, x, s.now())
			return nil
		})
	}
	for _, id := range ids(`SELECT id::text FROM quiz_duel WHERE status='finished' AND (NOT paid_a OR NOT paid_b) AND pay_tries < 20 ORDER BY finished_at LIMIT $1`, limit) {
		s.payDuel(ctx, id)
	}
	for _, id := range ids(`SELECT id::text FROM quiz_duel WHERE status IN ('expired','declined','cancelled') AND (fee_a='paid' OR fee_b='paid') LIMIT $1`, limit) {
		s.refundDuel(ctx, id)
	}
}

func (s *service) sweepDuels(ctx context.Context, cfg settings.All) {
	if cfg.Duel.HistoryDays > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM quiz_duel WHERE id IN (SELECT id FROM quiz_duel WHERE status NOT IN ('waiting','invited','active') AND fee_a<>'paid' AND fee_b<>'paid'
			AND (status<>'finished' OR (paid_a AND paid_b)) AND updated_at < now() - make_interval(days => $1) LIMIT 5000)`, cfg.Duel.HistoryDays)
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM quiz_duel_day WHERE day < to_char(now() - interval '7 days', 'YYYY-MM-DD')`)
}
