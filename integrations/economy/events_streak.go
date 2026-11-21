package main

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"time"
)

type streakState struct {
	Count     int
	Best      int
	Run       int
	GraceUsed int
	Peak      int
	Falls     int
	Last      *int64
}

type streakRule struct {
	Grace      int
	MissAction string
	Milestones []int
	WeekLength int
	Infinity   bool
	Track      int
}

type streakStep struct {
	Advanced    bool
	Reset       bool
	Floored     bool
	Fell        bool
	Consecutive bool
	Infinity    bool
}

func (r streakRule) inInfinity(st streakState) bool {
	return r.Infinity && r.Track > 0 && st.Peak >= r.Track
}

func streakFloor(peak int, milestones []int) int {
	floor := 0
	for _, d := range milestones {
		if d <= peak && d > floor {
			floor = d
		}
	}
	return floor
}

func streakNextMilestone(peak int, milestones []int) int {
	next := 0
	for _, d := range milestones {
		if d > peak && (next == 0 || d < next) {
			next = d
		}
	}
	return next
}

func streakMissed(st streakState, idx int64, r streakRule) (int, int) {
	missed := int(idx - *st.Last - 1)
	left := r.Grace - st.GraceUsed
	if left < 0 {
		left = 0
	}
	return missed, left
}

func streakFallCount(st streakState, excess int, r streakRule) (int, bool) {
	floor := streakFloor(st.Peak, r.Milestones)
	if r.MissAction == "week_start" {
		wl := r.WeekLength
		if wl < 1 {
			wl = 7
		}
		c := 0
		if st.Count > 0 {
			c = ((st.Count - 1) / wl) * wl
		}
		if c < floor {
			return floor, true
		}
		return c, false
	}
	if r.MissAction == "step_back" {
		c := st.Count - excess
		floored := false
		if c < floor {
			c = floor
			floored = floor > 0
		}
		if c < 0 {
			c = 0
		}
		return c, floored
	}
	if floor > 0 {
		return floor, true
	}
	return 0, false
}

func advanceStreak(st streakState, idx int64, r streakRule) (streakState, streakStep) {
	step := streakStep{}
	if st.Run < 1 {
		st.Run = 1
	}
	if st.Last != nil && idx <= *st.Last {
		return st, step
	}
	if st.Last == nil || st.Count == 0 {
		st.Count = 1
	} else if r.inInfinity(st) {
		st.Count++
		step.Infinity = true
		step.Consecutive = idx == *st.Last+1
	} else {
		missed, left := streakMissed(st, idx, r)
		switch {
		case missed == 0:
			st.Count++
			step.Consecutive = true
		case missed <= left:
			st.Count++
			st.GraceUsed += missed
		case r.MissAction == "week_start":
			c, floored := streakFallCount(st, missed-left, r)
			st.GraceUsed = 0
			st.Count = c + 1
			st.Falls++
			step.Fell, step.Floored = true, floored
		case r.MissAction == "step_back":
			c, floored := streakFallCount(st, missed-left, r)
			st.GraceUsed = r.Grace
			st.Count = c + 1
			st.Falls++
			step.Fell, step.Floored = true, floored
		default:
			c, floored := streakFallCount(st, missed-left, r)
			st.GraceUsed = 0
			if floored {
				st.Count = c + 1
				st.Falls++
				step.Fell, step.Floored = true, true
			} else {
				st.Run++
				st.Count = 1
				st.Peak = 0
				st.Falls = 0
				step.Reset = true
			}
		}
	}
	last := idx
	st.Last = &last
	if st.Count > st.Peak {
		st.Peak = st.Count
	}
	if st.Count > st.Best {
		st.Best = st.Count
	}
	step.Advanced = true
	return st, step
}

func streakRewardDay(count int, cycle int) int {
	if count < 1 {
		return 0
	}
	if cycle > 0 {
		return (count-1)%cycle + 1
	}
	return count
}

func streakDayScope(run, falls, count int, repay bool) string {
	if repay && falls > 0 {
		return "r" + strconv.Itoa(run) + "f" + strconv.Itoa(falls) + ":" + strconv.Itoa(count)
	}
	return "r" + strconv.Itoa(run) + ":" + strconv.Itoa(count)
}

func streakDayScopes(run, falls, count int, repay bool) []string {
	if !repay {
		return []string{streakDayScope(run, falls, count, false)}
	}
	out := []string{}
	for f := falls; f >= 0; f-- {
		out = append(out, streakDayScope(run, f, count, true))
	}
	return out
}

func streakMilestoneScope(day int) string {
	return "m" + strconv.Itoa(day)
}

func streakBroken(st streakState, idx int64, r streakRule) bool {
	if st.Last == nil || st.Count == 0 || r.inInfinity(st) {
		return false
	}
	missed, left := streakMissed(st, idx, r)
	return missed > left
}

type streakMilestone struct {
	id, prize, mode string
	day             int
}

func loadStreakMilestones(ctx context.Context, q dbRunner, streakIDs []string) (map[string][]streakMilestone, error) {
	out := map[string][]streakMilestone{}
	if len(streakIDs) == 0 {
		return out, nil
	}
	rows, err := q.QueryContext(ctx, `SELECT streak_id::text, id::text, day, COALESCE(prize_id::text,''), grant_mode FROM economy_streak_milestone WHERE streak_id::text = ANY($1) ORDER BY day`, streakIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sid string
		var m streakMilestone
		if err := rows.Scan(&sid, &m.id, &m.day, &m.prize, &m.mode); err != nil {
			return nil, err
		}
		out[sid] = append(out[sid], m)
	}
	return out, rows.Err()
}

func milestoneDays(list []streakMilestone) []int {
	days := make([]int, 0, len(list))
	for _, m := range list {
		days = append(days, m.day)
	}
	sort.Ints(days)
	return days
}

func (s *service) applyStreaks(ctx context.Context, tx *sql.Tx, userID, typeID string, now time.Time, clk eventClock, out *eventOutcome) error {
	rows, err := tx.QueryContext(ctx, `SELECT id::text, key, period_days, grace_periods, miss_action, COALESCE(cycle_length,0), repay_after_fall, week_length, infinity, infinity_currency, infinity_min, infinity_max,
		day_zone, day_start_minute, signup_delay_days, claim_window, infinity_mode, COALESCE(infinity_prize_id::text,'') FROM economy_streak WHERE event_type_id=$1 AND is_active ORDER BY sort, created_at, id`, typeID)
	if err != nil {
		return err
	}
	type streak struct {
		id, key, miss, infCur string
		days, grace, cycle    int
		repay, inf            bool
		week                  int
		infMin, infMax        int64
		zone, window, infMode string
		infPrize              string
		dayStart, delay       int
	}
	var list []streak
	ids := []string{}
	for rows.Next() {
		var st streak
		if err := rows.Scan(&st.id, &st.key, &st.days, &st.grace, &st.miss, &st.cycle, &st.repay, &st.week, &st.inf, &st.infCur, &st.infMin, &st.infMax,
			&st.zone, &st.dayStart, &st.delay, &st.window, &st.infMode, &st.infPrize); err != nil {
			rows.Close()
			return err
		}
		list = append(list, st)
		ids = append(ids, st.id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	milestones, err := loadStreakMilestones(ctx, tx, ids)
	if err != nil {
		return err
	}
	extras, err := loadStreakExtras(ctx, tx, ids)
	if err != nil {
		return err
	}
	var created *time.Time
	for _, sk := range list {
		if sk.delay > 0 && created == nil {
			var at time.Time
			if err := tx.QueryRowContext(ctx, `SELECT create_time FROM users WHERE id=$1`, userID).Scan(&at); err != nil {
				return err
			}
			created = &at
		}
	}
	for _, sk := range list {
		sclk := clk.withDay(sk.zone, sk.dayStart)
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_streak (streak_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, sk.id, userID); err != nil {
			return err
		}
		var st streakState
		var last sql.NullInt64
		if err := tx.QueryRowContext(ctx, `SELECT count, best, run, grace_used, peak, falls, last_period FROM economy_user_streak WHERE streak_id=$1 AND user_id=$2 FOR UPDATE`, sk.id, userID).
			Scan(&st.Count, &st.Best, &st.Run, &st.GraceUsed, &st.Peak, &st.Falls, &last); err != nil {
			return err
		}
		if last.Valid {
			v := last.Int64
			st.Last = &v
		}
		ms := milestones[sk.id]
		rule := streakRule{Grace: sk.grace, MissAction: sk.miss, Milestones: milestoneDays(ms), WeekLength: sk.week, Infinity: streakInfinityOn(sk.inf, sk.infMode, sk.infCur, sk.infPrize), Track: sk.cycle}
		idx := sclk.streakIndex(now, sk.days)
		next, step := advanceStreak(st, idx, rule)
		if !step.Advanced {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_streak SET count=$3, best=$4, run=$5, grace_used=$6, last_period=$7, peak=$9, falls=$10,
			started_at=CASE WHEN $8::boolean OR last_period IS NULL THEN now() ELSE started_at END, updated_at=now() WHERE streak_id=$1 AND user_id=$2`,
			sk.id, userID, next.Count, next.Best, next.Run, next.GraceUsed, *next.Last, step.Reset, next.Peak, next.Falls); err != nil {
			return err
		}
		floor := streakFloor(next.Peak, rule.Milestones)
		out.Streaks = append(out.Streaks, streakChange{StreakID: sk.id, Key: sk.key, Count: next.Count, Best: next.Best, Advanced: true, Reset: step.Reset,
			Floored: step.Floored, Floor: floor, NextMilestone: streakNextMilestone(next.Peak, rule.Milestones)})
		claimsOpen := sk.delay <= 0 || created == nil || sclk.dayIndex(now) >= sclk.dayIndex(*created)+int64(sk.delay)
		var expires *time.Time
		if sk.window == "period" {
			end := sclk.streakPeriodStart(idx+1, sk.days)
			expires = &end
		}
		addClaim := func(refID, scope, prizeID, mode string, value float64, windowed bool) error {
			if !claimsOpen {
				return nil
			}
			ref, err := insertClaim(ctx, tx, userID, "streak", sk.id, refID, scope, prizeID, mode, value)
			if err != nil || ref == nil {
				return err
			}
			if windowed && expires != nil && mode == "claim" {
				if _, err := tx.ExecContext(ctx, `UPDATE economy_reward_claim SET expires_at=$2 WHERE id=$1`, ref.ClaimID, *expires); err != nil {
					return err
				}
				ref.ExpiresAt = expires
			}
			out.Claims = append(out.Claims, *ref)
			return nil
		}
		if step.Infinity && sk.infMode == "claim" {
			ch := &out.Streaks[len(out.Streaks)-1]
			ch.Phase = "infinity"
			if err := addClaim(sk.id, "inf:r"+strconv.Itoa(next.Run)+":"+strconv.Itoa(next.Count), sk.infPrize, "claim", float64(next.Count), true); err != nil {
				return err
			}
			continue
		}
		if step.Infinity {
			amount, err := secureRange(sk.infMin, sk.infMax)
			if err != nil {
				return err
			}
			if amount > 0 {
				if err := s.walletChange(ctx, tx, userID, "streak:"+sk.id+":r"+strconv.Itoa(next.Run)+":"+strconv.Itoa(next.Count), map[string]int64{sk.infCur: amount}, map[string]any{
					"reason": "streak_infinity", "sourceType": "streak", "sourceId": sk.id}); err != nil {
					return err
				}
			}
			ch := &out.Streaks[len(out.Streaks)-1]
			ch.Phase, ch.InfinityCurrency, ch.InfinityAmount = "infinity", sk.infCur, amount
			continue
		}
		if step.Consecutive {
			for _, x := range extras[sk.id] {
				if x.day != next.Count {
					continue
				}
				if err := addClaim(x.id, "x"+streakDayScope(next.Run, next.Falls, next.Count, sk.repay), x.prize, x.mode, float64(next.Count), true); err != nil {
					return err
				}
			}
		}
		day := streakRewardDay(next.Count, sk.cycle)
		var rewardID, prizeID, mode string
		err := tx.QueryRowContext(ctx, `SELECT id::text, prize_id::text, grant_mode FROM economy_streak_reward WHERE streak_id=$1 AND day=$2`, sk.id, day).Scan(&rewardID, &prizeID, &mode)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err == nil {
			if err := addClaim(rewardID, streakDayScope(next.Run, next.Falls, next.Count, sk.repay), prizeID, mode, float64(next.Count), true); err != nil {
				return err
			}
		}
		for _, m := range ms {
			if m.prize == "" || m.day > next.Count {
				continue
			}
			if err := addClaim(sk.id, streakMilestoneScope(m.day), m.prize, m.mode, float64(m.day), false); err != nil {
				return err
			}
		}
	}
	return nil
}

func streakInfinityOn(on bool, mode, currency, prize string) bool {
	if !on {
		return false
	}
	if mode == "claim" {
		return prize != ""
	}
	return currency != ""
}

func loadStreakExtras(ctx context.Context, q dbRunner, streakIDs []string) (map[string][]streakMilestone, error) {
	out := map[string][]streakMilestone{}
	if len(streakIDs) == 0 {
		return out, nil
	}
	rows, err := q.QueryContext(ctx, `SELECT streak_id::text, id::text, day, prize_id::text, grant_mode FROM economy_streak_extra WHERE streak_id::text = ANY($1) ORDER BY day`, streakIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sid string
		var m streakMilestone
		if err := rows.Scan(&sid, &m.id, &m.day, &m.prize, &m.mode); err != nil {
			return nil, err
		}
		out[sid] = append(out[sid], m)
	}
	return out, rows.Err()
}
