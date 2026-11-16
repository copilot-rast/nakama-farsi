package main

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) duelRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"quiz_duel_categories":    s.rpcDuelCategories,
		"quiz_duel_start":         s.rpcDuelStart,
		"quiz_duel_list":          s.rpcDuelList,
		"quiz_duel_get":           s.rpcDuelGet,
		"quiz_duel_accept":        s.rpcDuelAccept,
		"quiz_duel_decline":       s.rpcDuelDecline,
		"quiz_duel_pick_category": s.rpcDuelPick,
		"quiz_duel_question":      s.rpcDuelQuestion,
		"quiz_duel_answer":        s.rpcDuelAnswer,
		"quiz_duel_forfeit":       s.rpcDuelForfeit,
		"quiz_duel_stats":         s.rpcDuelStats,
	}
}

type categoryInfo struct {
	Key, IconURL, Color string
	Name, Description   map[string]string
}

func (s *service) categoryMap(ctx context.Context, keys []string) map[string]categoryInfo {
	out := map[string]categoryInfo{}
	if len(keys) == 0 {
		return out
	}
	rows, err := s.db.QueryContext(ctx, `SELECT key, name, description, icon_url, color FROM quiz_category WHERE key = ANY($1::text[])`, keys)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var c categoryInfo
		var name, desc []byte
		if rows.Scan(&c.Key, &name, &desc, &c.IconURL, &c.Color) == nil {
			c.Name, c.Description = map[string]string{}, map[string]string{}
			unmarshal(name, &c.Name)
			unmarshal(desc, &c.Description)
			out[c.Key] = c
		}
	}
	return out
}

func categoryView(langs *localized.Set, lang string, cats map[string]categoryInfo, key string) map[string]any {
	c, ok := cats[key]
	if !ok {
		return map[string]any{"key": key, "name": key, "description": "", "iconUrl": "", "color": ""}
	}
	name := text(langs, lang, c.Name)
	if name == "" {
		name = key
	}
	return map[string]any{"key": key, "name": name, "description": text(langs, lang, c.Description), "iconUrl": c.IconURL, "color": c.Color}
}

type playerInfo struct {
	ID, Username, DisplayName, AvatarURL string
}

func (s *service) players(ctx context.Context, ids ...string) map[string]playerInfo {
	out := map[string]playerInfo{}
	clean := []string{}
	for _, id := range ids {
		if id != "" {
			clean = append(clean, id)
		}
	}
	if len(clean) == 0 {
		return out
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, username, COALESCE(display_name,''), COALESCE(avatar_url,'') FROM users WHERE id::text = ANY($1::text[])`, clean)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var p playerInfo
		if rows.Scan(&p.ID, &p.Username, &p.DisplayName, &p.AvatarURL) == nil {
			out[p.ID] = p
		}
	}
	return out
}

func playerView(p map[string]playerInfo, id string, score any) any {
	if id == "" {
		return nil
	}
	x := p[id]
	return map[string]any{"userId": id, "username": x.Username, "displayName": x.DisplayName, "avatarUrl": x.AvatarURL, "score": score}
}

func who(d *duelRow, me, id string) string {
	switch {
	case id == "":
		return ""
	case id == me:
		return "you"
	default:
		return "opponent"
	}
}

func (s *service) duelView(ctx context.Context, st *duelState, me string, langs *localized.Set, lang string, full bool) map[string]any {
	d := st.d
	opp := d.other(me)
	over := !d.open()
	keys := []string{}
	for _, rr := range st.rounds {
		keys = append(keys, rr.Category)
		keys = append(keys, rr.Offers...)
	}
	cats := s.categoryMap(ctx, keys)
	seen := func(r int) bool { return over || st.roundDone(me, r) }
	oppScore := st.correct(opp, seen)
	ps := s.players(ctx, me, opp)
	sp := st.step(me)
	step := map[string]any{"kind": sp.Kind, "round": sp.Round}
	switch {
	case over:
		step = map[string]any{"kind": "over", "round": 0}
	case d.Status == "invited" && me == d.B:
		step = map[string]any{"kind": "accept", "round": 0}
	case d.Turn != me && (sp.Kind == "pick" || sp.Kind == "answer"):
		step["kind"] = "wait"
	}
	if step["kind"] == "pick" {
		offers := []any{}
		if rr := st.rounds[sp.Round]; rr != nil {
			for _, k := range rr.Offers {
				offers = append(offers, categoryView(langs, lang, cats, k))
			}
		}
		step["offers"] = offers
	}
	if step["kind"] == "answer" {
		if rr := st.rounds[sp.Round]; rr != nil {
			step["category"] = categoryView(langs, lang, cats, rr.Category)
			n := 0
			for _, p := range rr.Questions {
				if a := st.answer(me, p.ID); a != nil && a.Outcome != "pending" {
					n++
				}
			}
			step["answered"], step["total"] = n, len(rr.Questions)
		}
	}
	turn := who(d, me, d.Turn)
	if turn == "" {
		turn = "none"
	}
	result := any(nil)
	if d.Status == "finished" {
		r := "draw"
		if d.Winner == me {
			r = "win"
		} else if d.Winner != "" {
			r = "lose"
		}
		result = r
	}
	out := map[string]any{
		"duelId": d.ID, "mode": d.Mode, "status": d.Status, "you": playerView(ps, me, st.correct(me, nil)), "opponent": playerView(ps, opp, oppScore),
		"turn": turn, "step": step, "rounds": d.Rules.Rounds, "questionsPerRound": d.Rules.PerRound, "questionSeconds": d.Rules.QuestionSeconds,
		"turnDeadline": iso(d.TurnDeadline), "expiresAt": iso(d.Expires), "result": result, "reason": d.Reason, "winner": who(d, me, d.Winner),
		"createdAt": d.Created.UTC().Format(time.RFC3339), "finishedAt": iso(d.Finished), "updatedAt": d.Updated.UTC().Format(time.RFC3339),
		"serverTime": s.now().UTC().Format(time.RFC3339), "entryFee": d.Rules.EntryFee, "invitedByYou": d.Mode == "friend" && d.A == me,
	}
	if !full {
		return out
	}
	list := []any{}
	for r := 1; r <= d.Rules.Rounds; r++ {
		rr := st.rounds[r]
		row := map[string]any{"round": r, "picker": who(d, me, d.picker(r)), "category": nil, "you": []string{}, "opponent": nil, "youCorrect": 0, "opponentCorrect": nil}
		if rr != nil && rr.Category != "" {
			row["category"] = categoryView(langs, lang, cats, rr.Category)
			mine, theirs := []string{}, []string{}
			mc, tc := 0, 0
			for _, p := range rr.Questions {
				o := "none"
				if a := st.answer(me, p.ID); a != nil {
					o = a.Outcome
					if o == "correct" {
						mc++
					}
				}
				mine = append(mine, o)
				t := "none"
				if a := st.answer(opp, p.ID); a != nil && a.Outcome != "pending" {
					t = a.Outcome
					if t == "correct" {
						tc++
					}
				}
				theirs = append(theirs, t)
			}
			row["you"], row["youCorrect"] = mine, mc
			if seen(r) {
				row["opponent"], row["opponentCorrect"] = theirs, tc
			}
		}
		list = append(list, row)
	}
	out["roundList"] = list
	return out
}

func (s *service) duelLang(ctx context.Context, userID, lang string) (*localized.Set, string) {
	langs, l, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, lang
	}
	return langs, l
}

func (s *service) rpcDuelCategories(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	langs, lang := s.duelLang(ctx, userID, in.Lang)
	r := rulesOf(cfg.Duel)
	rows, err := s.db.QueryContext(ctx, `SELECT c.key, (SELECT count(*) FROM quiz_question q WHERE q.category=c.key AND `+s.eligibleSQL(r)+`)
		FROM quiz_category c WHERE c.is_active AND c.duel ORDER BY c.sort, c.key`)
	if err != nil {
		return nil, err
	}
	type row struct {
		key string
		n   int
	}
	var list []row
	keys := []string{}
	for rows.Next() {
		var x row
		if rows.Scan(&x.key, &x.n) == nil {
			list = append(list, x)
			keys = append(keys, x.key)
		}
	}
	rows.Close()
	cats := s.categoryMap(ctx, keys)
	out := []any{}
	for _, x := range list {
		v := categoryView(langs, lang, cats, x.key)
		v["playable"] = x.n >= r.PerRound
		out = append(out, v)
	}
	return map[string]any{"lang": lang, "enabled": cfg.Duel.Enabled, "categories": out, "rules": map[string]any{
		"rounds": r.Rounds, "questionsPerRound": r.PerRound, "categoryChoices": r.Choices, "questionSeconds": r.QuestionSeconds, "turnHours": r.TurnHours,
		"waitingHours": cfg.Duel.WaitingHours, "inviteHours": cfg.Duel.InviteHours, "friendsOnly": cfg.Duel.FriendsOnly, "maxActive": cfg.Duel.MaxActive,
		"dailyStarts": cfg.Duel.DailyStarts, "entryFee": r.EntryFee, "perCorrect": r.PerCorrect, "winPrize": r.WinPrize, "losePrize": r.LosePrize, "drawPrize": r.DrawPrize,
	}}, nil
}

func (s *service) activeCount(ctx context.Context, q queryer, userID string) (int, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM quiz_duel WHERE (player_a=$1 OR player_b=$1) AND status IN ('waiting','invited','active')`, userID).Scan(&n)
	return n, err
}

func blockedSQL(me, other string) string {
	return `EXISTS (SELECT 1 FROM user_edge e WHERE e.state=3 AND ((e.source_id=` + me + ` AND e.destination_id=` + other + `) OR (e.source_id=` + other + ` AND e.destination_id=` + me + `)))`
}

func (s *service) rpcDuelStart(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Mode   string `json:"mode"`
		UserID string `json:"userId"`
		Lang   string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if !cfg.Duel.Enabled {
		return nil, apierr.Unavailable("QUIZ_DUEL_DISABLED")
	}
	in.Mode = strings.TrimSpace(in.Mode)
	if in.Mode == "" {
		in.Mode = "random"
	}
	if in.Mode != "random" && in.Mode != "friend" {
		return nil, apierr.BadRequest("QUIZ_DUEL_INVALID_MODE").WithField("mode")
	}
	in.UserID = strings.TrimSpace(in.UserID)
	if in.Mode == "friend" {
		if _, err := uuid.FromString(in.UserID); err != nil {
			return nil, apierr.NotFound("QUIZ_DUEL_PLAYER_NOT_FOUND").WithField("userId")
		}
		if in.UserID == userID {
			return nil, apierr.BadRequest("QUIZ_DUEL_SELF").WithField("userId")
		}
	}
	rules := rulesOf(cfg.Duel)
	now := s.now()
	day := dayKey(now, s.zone(cfg))
	var joined *duelRow
	var created *duelRow
	x := &duelTx{}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		x, joined, created = &duelTx{}, nil, nil
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('quiz_duel:' || $1::text))`, userID); err != nil {
			return err
		}
		n, err := s.activeCount(ctx, tx, userID)
		if err != nil {
			return err
		}
		if n >= cfg.Duel.MaxActive {
			return apierr.Conflict("QUIZ_DUEL_ACTIVE_LIMIT").With("limit", cfg.Duel.MaxActive)
		}
		if cfg.Duel.DailyStarts > 0 {
			var starts int
			if err := tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT starts FROM quiz_duel_day WHERE user_id=$1 AND day=$2),0)`, userID, day).Scan(&starts); err != nil {
				return err
			}
			if starts >= cfg.Duel.DailyStarts {
				return apierr.TooManyRequests("QUIZ_DUEL_DAILY_LIMIT", secondsToNextDay(now, s.zone(cfg))).With("limit", cfg.Duel.DailyStarts)
			}
		}
		if in.Mode == "random" {
			d, err := scanDuel(tx.QueryRowContext(ctx, `SELECT `+duelCols+` FROM quiz_duel d WHERE d.status='waiting' AND d.player_b IS NULL AND d.player_a<>$1
				AND NOT `+blockedSQL("$1::uuid", "d.player_a")+` AND (d.expires_at IS NULL OR d.expires_at > $2) ORDER BY d.created_at LIMIT 1 FOR UPDATE SKIP LOCKED`, userID, now))
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if err == nil {
				st, err := loadState(ctx, tx, d)
				if err != nil {
					return err
				}
				t := now
				d.B, d.Status, d.Accepted, d.Expires = userID, "active", &t, nil
				if err := s.chargeFee(ctx, d, userID); err != nil {
					return err
				}
				if d.Rules.EntryFee.Amount > 0 {
					d.FeeB = "paid"
				}
				d.Turn = d.A
				if err := s.advance(ctx, tx, st, x, now); err != nil {
					return err
				}
				kept := x.notes[:0]
				for _, n := range x.notes {
					if n.to != userID {
						kept = append(kept, n)
					}
				}
				x.notes = kept
				x.note(d.A, "joined", userID, false)
				if err := s.saveDuel(ctx, tx, d); err != nil {
					return err
				}
				joined = d
			}
		} else {
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1 AND disable_time = '1970-01-01 00:00:00+00')`, in.UserID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return apierr.NotFound("QUIZ_DUEL_PLAYER_NOT_FOUND").WithField("userId")
			}
			var blocked, friends bool
			if err := tx.QueryRowContext(ctx, `SELECT `+blockedSQL("$1::uuid", "$2::uuid")+`, EXISTS (SELECT 1 FROM user_edge WHERE source_id=$1 AND destination_id=$2 AND state=0)`, userID, in.UserID).Scan(&blocked, &friends); err != nil {
				return err
			}
			if blocked {
				return apierr.Forbidden("QUIZ_DUEL_BLOCKED").WithField("userId")
			}
			if cfg.Duel.FriendsOnly && !friends {
				return apierr.Forbidden("QUIZ_DUEL_NOT_FRIENDS").WithField("userId")
			}
		}
		if joined == nil {
			id := uuid.Must(uuid.NewV4()).String()
			d := &duelRow{ID: id, Mode: in.Mode, Status: "active", A: userID, Rules: rules, Created: now, Updated: now}
			if in.Mode == "friend" {
				d.Status, d.B = "invited", in.UserID
				exp := now.Add(time.Duration(cfg.Duel.InviteHours) * time.Hour)
				d.Expires = &exp
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_duel (id, mode, status, player_a, player_b, rules, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`,
				d.ID, d.Mode, d.Status, d.A, nullable(d.B), jsonOf(d.Rules), now); err != nil {
				return err
			}
			st := &duelState{d: d, rounds: map[int]*duelRound{}, answers: map[string]map[string]*duelAnswer{}}
			s.startTurn(d, userID, now)
			if err := s.ensureOffers(ctx, tx, st, 1); err != nil {
				return err
			}
			if err := s.saveDuel(ctx, tx, d); err != nil {
				return err
			}
			if d.Mode == "friend" {
				x.note(d.B, "invite", userID, true)
			}
			created = d
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO quiz_duel_day (user_id, day, starts) VALUES ($1,$2,1) ON CONFLICT (user_id, day) DO UPDATE SET starts=quiz_duel_day.starts+1`, userID, day)
		return err
	})
	if err != nil {
		return nil, err
	}
	d := joined
	if d == nil {
		d = created
		if err := s.chargeFee(ctx, d, userID); err != nil {
			_, _ = s.db.ExecContext(ctx, `DELETE FROM quiz_duel WHERE id=$1`, d.ID)
			_, _ = s.db.ExecContext(ctx, `UPDATE quiz_duel_day SET starts=GREATEST(starts-1,0) WHERE user_id=$1 AND day=$2`, userID, day)
			return nil, err
		}
		if d.Rules.EntryFee.Amount > 0 {
			_, _ = s.db.ExecContext(ctx, `UPDATE quiz_duel SET fee_a='paid' WHERE id=$1`, d.ID)
		}
	}
	s.flush(ctx, d, x)
	return s.duelAnswerView(ctx, userID, d.ID, in.Lang, map[string]any{"joined": joined != nil})
}

func (s *service) duelAnswerView(ctx context.Context, userID, id, lang string, extra map[string]any) (any, error) {
	d, err := loadDuel(ctx, s.db, id, false)
	if err != nil {
		return nil, err
	}
	if !d.has(userID) {
		return nil, apierr.NotFound("QUIZ_DUEL_NOT_FOUND").WithField("duelId")
	}
	st, err := loadState(ctx, s.db, d)
	if err != nil {
		return nil, err
	}
	langs, l := s.duelLang(ctx, userID, lang)
	v := s.duelView(ctx, st, userID, langs, l, true)
	for k, val := range extra {
		v[k] = val
	}
	return v, nil
}

type duelIn struct {
	DuelID      string `json:"duelId"`
	CategoryKey string `json:"categoryKey"`
	QuestionID  string `json:"questionId"`
	OptionID    *int   `json:"optionId"`
	Lang        string `json:"lang"`
}

func (s *service) decodeDuel(payload string) (duelIn, error) {
	var in duelIn
	if err := apierr.Decode(payload, &in); err != nil {
		return in, err
	}
	in.DuelID = strings.TrimSpace(in.DuelID)
	if in.DuelID == "" {
		return in, apierr.NotFound("QUIZ_DUEL_NOT_FOUND").WithField("duelId")
	}
	return in, nil
}

func mine(st *duelState, userID string) error {
	if !st.d.has(userID) {
		return apierr.NotFound("QUIZ_DUEL_NOT_FOUND").WithField("duelId")
	}
	return nil
}

func (s *service) rpcDuelGet(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.decodeDuel(payload)
	if err != nil {
		return nil, err
	}
	return s.duelAnswerView(ctx, userID, in.DuelID, in.Lang, nil)
}

func (s *service) rpcDuelAccept(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.decodeDuel(payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	_, err = s.mutate(ctx, in.DuelID, func(tx *sql.Tx, st *duelState, x *duelTx) error {
		d := st.d
		if err := mine(st, userID); err != nil {
			return err
		}
		if d.Status != "invited" || d.B != userID {
			return apierr.Conflict("QUIZ_DUEL_NOT_INVITED").With("status", d.Status)
		}
		n, err := s.activeCount(ctx, tx, userID)
		if err != nil {
			return err
		}
		if n > cfg.Duel.MaxActive {
			return apierr.Conflict("QUIZ_DUEL_ACTIVE_LIMIT").With("limit", cfg.Duel.MaxActive)
		}
		if err := s.chargeFee(ctx, d, userID); err != nil {
			return err
		}
		if d.Rules.EntryFee.Amount > 0 {
			d.FeeB = "paid"
		}
		now := s.now()
		d.Status, d.Accepted, d.Expires = "active", &now, nil
		if d.Turn == d.B {
			s.startTurn(d, d.B, now)
		}
		x.note(d.A, "accepted", userID, true)
		return s.advance(ctx, tx, st, x, now)
	})
	if err != nil {
		return nil, err
	}
	return s.duelAnswerView(ctx, userID, in.DuelID, in.Lang, nil)
}

func (s *service) rpcDuelDecline(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.decodeDuel(payload)
	if err != nil {
		return nil, err
	}
	_, err = s.mutate(ctx, in.DuelID, func(tx *sql.Tx, st *duelState, x *duelTx) error {
		d := st.d
		if err := mine(st, userID); err != nil {
			return err
		}
		if d.Status != "invited" || d.B != userID {
			return apierr.Conflict("QUIZ_DUEL_NOT_INVITED").With("status", d.Status)
		}
		s.endWithout(d, "declined", "declined", x, s.now())
		x.note(d.A, "declined", userID, true)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.duelAnswerView(ctx, userID, in.DuelID, in.Lang, nil)
}

func (s *service) rpcDuelForfeit(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.decodeDuel(payload)
	if err != nil {
		return nil, err
	}
	_, err = s.mutate(ctx, in.DuelID, func(tx *sql.Tx, st *duelState, x *duelTx) error {
		d := st.d
		if err := mine(st, userID); err != nil {
			return err
		}
		if !d.open() {
			return apierr.Conflict("QUIZ_DUEL_OVER").With("status", d.Status)
		}
		now := s.now()
		if d.B == "" || d.Accepted == nil {
			if d.Status == "invited" && userID == d.B {
				s.endWithout(d, "declined", "declined", x, now)
				x.note(d.A, "declined", userID, true)
				return nil
			}
			s.endWithout(d, "cancelled", "forfeit", x, now)
			x.note(d.B, "cancelled", userID, false)
			return nil
		}
		d.ScoreA, d.ScoreB = st.correct(d.A, nil), st.correct(d.B, nil)
		w := d.other(userID)
		s.finishDuel(d, "forfeit", &w, x, now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.duelAnswerView(ctx, userID, in.DuelID, in.Lang, nil)
}

func myTurn(st *duelState, userID string) error {
	d := st.d
	if !d.open() {
		return apierr.Conflict("QUIZ_DUEL_OVER").With("status", d.Status)
	}
	if d.Status == "invited" && userID == d.B {
		return apierr.Conflict("QUIZ_DUEL_NOT_ACCEPTED")
	}
	if d.Turn != userID {
		return apierr.Conflict("QUIZ_DUEL_NOT_YOUR_TURN")
	}
	return nil
}

func (s *service) rpcDuelPick(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.decodeDuel(payload)
	if err != nil {
		return nil, err
	}
	in.CategoryKey = strings.TrimSpace(in.CategoryKey)
	_, err = s.mutate(ctx, in.DuelID, func(tx *sql.Tx, st *duelState, x *duelTx) error {
		if err := mine(st, userID); err != nil {
			return err
		}
		if err := myTurn(st, userID); err != nil {
			return err
		}
		sp := st.step(userID)
		if sp.Kind != "pick" {
			return apierr.Conflict("QUIZ_DUEL_NOT_PICKING").With("step", sp.Kind)
		}
		if err := s.ensureOffers(ctx, tx, st, sp.Round); err != nil {
			return err
		}
		rr := st.rounds[sp.Round]
		if !contains(rr.Offers, in.CategoryKey) {
			return apierr.BadRequest("QUIZ_DUEL_CATEGORY_NOT_OFFERED").WithField("categoryKey")
		}
		qs, err := s.drawRound(ctx, tx, st, in.CategoryKey)
		if err != nil {
			return err
		}
		now := s.now()
		if _, err := tx.ExecContext(ctx, `UPDATE quiz_duel_round SET category=$3, questions=$4, picked_at=$5 WHERE duel_id=$1 AND round_no=$2 AND category=''`,
			st.d.ID, sp.Round, in.CategoryKey, jsonOf(qs), now); err != nil {
			return err
		}
		rr.Category, rr.Questions, rr.PickedAt = in.CategoryKey, qs, &now
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.duelAnswerView(ctx, userID, in.DuelID, in.Lang, nil)
}

func duelDeadline(r duelRules, served time.Time) time.Time {
	return served.Add(time.Duration(r.QuestionSeconds) * time.Second)
}

func duelLate(r duelRules, served, now time.Time) bool {
	return now.After(duelDeadline(r, served).Add(time.Duration(r.GraceSeconds) * time.Second))
}

func (s *service) closeLate(ctx context.Context, tx *sql.Tx, st *duelState, userID string, now time.Time) (bool, error) {
	changed := false
	for _, a := range st.answers[userID] {
		if a.Outcome != "pending" || !duelLate(st.d.Rules, a.Served, now) {
			continue
		}
		dl := duelDeadline(st.d.Rules, a.Served)
		if _, err := tx.ExecContext(ctx, `UPDATE quiz_duel_answer SET outcome='late', answered_at=$4, seconds=$5 WHERE duel_id=$1 AND user_id=$2 AND question_id=$3 AND outcome='pending'`,
			st.d.ID, userID, a.QuestionID, now, float64(st.d.Rules.QuestionSeconds)); err != nil {
			return false, err
		}
		a.Outcome, a.Answered, a.Seconds = "late", &dl, float64(st.d.Rules.QuestionSeconds)
		changed = true
	}
	return changed, nil
}

func (s *service) serve(ctx context.Context, tx *sql.Tx, st *duelState, userID string, now time.Time) (*duelAnswer, int, error) {
	sp := st.step(userID)
	if sp.Kind != "answer" {
		return nil, 0, nil
	}
	rr := st.rounds[sp.Round]
	for i, p := range rr.Questions {
		a := st.answer(userID, p.ID)
		if a != nil && a.Outcome != "pending" {
			continue
		}
		if a != nil {
			return a, sp.Round, nil
		}
		a = &duelAnswer{UserID: userID, QuestionID: p.ID, Round: sp.Round, Idx: i, Served: now, Outcome: "pending"}
		if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_duel_answer (duel_id, user_id, question_id, round_no, idx, served_at) VALUES ($1,$2,$3,$4,$5,$6)`,
			st.d.ID, userID, p.ID, sp.Round, i, now); err != nil {
			return nil, 0, err
		}
		if st.answers[userID] == nil {
			st.answers[userID] = map[string]*duelAnswer{}
		}
		st.answers[userID][p.ID] = a
		return a, sp.Round, nil
	}
	return nil, 0, nil
}

func (s *service) questionOut(ctx context.Context, st *duelState, a *duelAnswer, userID string, langs *localized.Set, lang string) (map[string]any, error) {
	if a == nil {
		return nil, nil
	}
	qs, err := listQuestions(ctx, s.db, `q.id=$1`, a.QuestionID)
	if err != nil {
		return nil, err
	}
	if len(qs) == 0 {
		return nil, apierr.NotFound("QUIZ_QUESTION_NOT_FOUND")
	}
	q := qs[0]
	rr := st.rounds[a.Round]
	order := []int{}
	for _, p := range rr.Questions {
		if p.ID == q.ID {
			order = p.Order
		}
	}
	byID := map[int]Option{}
	for _, o := range q.Options {
		byID[o.ID] = o
	}
	opts := []any{}
	for _, id := range order {
		if o, ok := byID[id]; ok {
			opts = append(opts, map[string]any{"id": o.ID, "text": text(langs, lang, o.Text), "mediaUrl": o.MediaURL})
		}
	}
	dl := duelDeadline(st.d.Rules, a.Served)
	return map[string]any{"duelId": st.d.ID, "round": a.Round, "index": a.Idx, "total": len(rr.Questions), "questionId": q.ID, "text": text(langs, lang, q.Question),
		"mediaUrl": q.MediaURL, "options": opts, "servedAt": a.Served.UTC().Format(time.RFC3339Nano), "deadlineAt": dl.UTC().Format(time.RFC3339Nano),
		"seconds": st.d.Rules.QuestionSeconds, "serverTime": s.now().UTC().Format(time.RFC3339Nano)}, nil
}

func (s *service) rpcDuelQuestion(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.decodeDuel(payload)
	if err != nil {
		return nil, err
	}
	var served *duelAnswer
	st, err := s.mutate(ctx, in.DuelID, func(tx *sql.Tx, st *duelState, x *duelTx) error {
		served = nil
		if err := mine(st, userID); err != nil {
			return err
		}
		if err := myTurn(st, userID); err != nil {
			return err
		}
		now := s.now()
		if _, err := s.closeLate(ctx, tx, st, userID, now); err != nil {
			return err
		}
		sp := st.step(userID)
		if sp.Kind == "pick" {
			if err := s.ensureOffers(ctx, tx, st, sp.Round); err != nil {
				return err
			}
			return nil
		}
		a, _, err := s.serve(ctx, tx, st, userID, now)
		if err != nil {
			return err
		}
		served = a
		if a == nil {
			return s.advance(ctx, tx, st, x, now)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	langs, lang := s.duelLang(ctx, userID, in.Lang)
	q, err := s.questionOut(ctx, st, served, userID, langs, lang)
	if err != nil {
		return nil, err
	}
	if q == nil && st.d.Turn == userID {
		if sp := st.step(userID); sp.Kind == "pick" {
			return nil, apierr.Conflict("QUIZ_DUEL_PICK_FIRST").With("round", sp.Round)
		}
	}
	return map[string]any{"question": q, "duel": s.duelView(ctx, st, userID, langs, lang, true)}, nil
}

func (s *service) rpcDuelAnswer(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.decodeDuel(payload)
	if err != nil {
		return nil, err
	}
	in.QuestionID = strings.TrimSpace(in.QuestionID)
	var res map[string]any
	var next *duelAnswer
	var qq *question
	st, err := s.mutate(ctx, in.DuelID, func(tx *sql.Tx, st *duelState, x *duelTx) error {
		res, next, qq = nil, nil, nil
		if err := mine(st, userID); err != nil {
			return err
		}
		a := st.answer(userID, in.QuestionID)
		if a == nil {
			if !st.d.open() {
				return apierr.Conflict("QUIZ_DUEL_OVER").With("status", st.d.Status)
			}
			return apierr.Conflict("QUIZ_DUEL_QUESTION_NOT_SERVED").WithField("questionId")
		}
		qs, err := listQuestions(ctx, tx, `q.id=$1`, a.QuestionID)
		if err != nil {
			return err
		}
		if len(qs) == 0 {
			return apierr.NotFound("QUIZ_QUESTION_NOT_FOUND")
		}
		qq = qs[0]
		replayed := a.Outcome != "pending"
		now := s.now()
		if !replayed {
			if err := myTurn(st, userID); err != nil {
				return err
			}
			outcome := "wrong"
			if in.OptionID == nil {
				outcome = "skip"
			} else {
				found := false
				for _, o := range qq.Options {
					if o.ID == *in.OptionID {
						found = true
						if o.Correct {
							outcome = "correct"
						}
					}
				}
				if !found {
					return apierr.BadRequest("QUIZ_OPTION_NOT_FOUND").WithField("optionId")
				}
			}
			if duelLate(st.d.Rules, a.Served, now) {
				outcome = "late"
			}
			secs := math.Round(now.Sub(a.Served).Seconds()*100) / 100
			if _, err := tx.ExecContext(ctx, `UPDATE quiz_duel_answer SET outcome=$4, option_id=$5, answered_at=$6, seconds=$7 WHERE duel_id=$1 AND user_id=$2 AND question_id=$3 AND outcome='pending'`,
				st.d.ID, userID, a.QuestionID, outcome, in.OptionID, now, secs); err != nil {
				return err
			}
			a.Outcome, a.Option, a.Answered, a.Seconds = outcome, in.OptionID, &now, secs
			if err := s.advance(ctx, tx, st, x, now); err != nil {
				return err
			}
			if st.d.Rules.ServeNext && st.d.Turn == userID {
				if next, _, err = s.serve(ctx, tx, st, userID, now); err != nil {
					return err
				}
			}
		}
		res = map[string]any{"duelId": st.d.ID, "questionId": a.QuestionID, "round": a.Round, "index": a.Idx, "outcome": a.Outcome, "correct": a.Outcome == "correct",
			"replayed": replayed, "optionId": a.Option, "correctOptionIds": correctIDs(qq), "seconds": a.Seconds}
		if o := st.answer(st.d.other(userID), a.QuestionID); o != nil && o.Outcome != "pending" {
			res["opponentOptionId"], res["opponentOutcome"] = o.Option, o.Outcome
		}
		rr := st.rounds[a.Round]
		res["roundDone"] = st.roundDone(userID, a.Round)
		res["roundTotal"] = len(rr.Questions)
		return nil
	})
	if err != nil {
		return nil, err
	}
	langs, lang := s.duelLang(ctx, userID, in.Lang)
	if s.settings(ctx).Reveal.Explanation {
		res["explanation"] = text(langs, lang, qq.Explanation)
	}
	if next != nil {
		nq, err := s.questionOut(ctx, st, next, userID, langs, lang)
		if err != nil {
			return nil, err
		}
		res["next"] = nq
	}
	res["duel"] = s.duelView(ctx, st, userID, langs, lang, true)
	return res, nil
}

func (s *service) rpcDuelList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Status string `json:"status"`
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor"`
		Lang   string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if in.Limit <= 0 || in.Limit > 50 {
		in.Limit = 20
	}
	offset, _ := strconv.Atoi(in.Cursor)
	if offset < 0 {
		offset = 0
	}
	where := `(player_a=$1 OR player_b=$1)`
	switch in.Status {
	case "", "all":
	case "open":
		where += ` AND status IN ('waiting','invited','active')`
	case "finished":
		where += ` AND status NOT IN ('waiting','invited','active')`
	case "waiting", "invited", "active":
		where += ` AND status='` + in.Status + `'`
	default:
		return nil, apierr.BadRequest("QUIZ_DUEL_INVALID_STATUS").WithField("status")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+duelCols+` FROM quiz_duel WHERE `+where+`
		ORDER BY (status IN ('waiting','invited','active')) DESC, (turn_user=$1) DESC NULLS LAST, updated_at DESC, id LIMIT $2 OFFSET $3`, userID, in.Limit+1, offset)
	if err != nil {
		return nil, err
	}
	var list []*duelRow
	for rows.Next() {
		d, err := scanDuel(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, d)
	}
	rows.Close()
	next := ""
	if len(list) > in.Limit {
		list = list[:in.Limit]
		next = strconv.Itoa(offset + in.Limit)
	}
	langs, lang := s.duelLang(ctx, userID, in.Lang)
	out := []any{}
	for _, d := range list {
		st, err := loadState(ctx, s.db, d)
		if err != nil {
			return nil, err
		}
		out = append(out, s.duelView(ctx, st, userID, langs, lang, false))
	}
	return map[string]any{"duels": out, "nextCursor": next}, nil
}

func (s *service) rpcDuelStats(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserID string `json:"userId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	target := userID
	if in.UserID != "" {
		if _, err := uuid.FromString(in.UserID); err != nil {
			return nil, apierr.NotFound("QUIZ_DUEL_PLAYER_NOT_FOUND").WithField("userId")
		}
		target = in.UserID
	}
	var played, wins, draws, open int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE status='finished'), count(*) FILTER (WHERE status='finished' AND winner=$1),
		count(*) FILTER (WHERE status='finished' AND winner IS NULL), count(*) FILTER (WHERE status IN ('waiting','invited','active'))
		FROM quiz_duel WHERE player_a=$1 OR player_b=$1`, target).Scan(&played, &wins, &draws, &open); err != nil {
		return nil, err
	}
	var answered, right int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE outcome<>'pending'), count(*) FILTER (WHERE outcome='correct') FROM quiz_duel_answer WHERE user_id=$1`, target).Scan(&answered, &right); err != nil {
		return nil, err
	}
	rate := 0.0
	if answered > 0 {
		rate = math.Round(float64(right)*10000/float64(answered)) / 100
	}
	return map[string]any{"userId": target, "played": played, "wins": wins, "losses": played - wins - draws, "draws": draws, "open": open,
		"answered": answered, "correct": right, "correctRate": rate}, nil
}
