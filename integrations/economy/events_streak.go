package main

import (
	"context"
	"database/sql"
	"strconv"
	"time"
)

type streakState struct {
	Count     int
	Best      int
	Run       int
	GraceUsed int
	Last      *int64
}

type streakRule struct {
	Grace      int
	MissAction string
}

func advanceStreak(st streakState, idx int64, r streakRule) (streakState, bool, bool) {
	if st.Run < 1 {
		st.Run = 1
	}
	if st.Last != nil && idx <= *st.Last {
		return st, false, false
	}
	reset := false
	if st.Last == nil || st.Count == 0 {
		st.Count = 1
	} else {
		missed := int(idx - *st.Last - 1)
		left := r.Grace - st.GraceUsed
		if left < 0 {
			left = 0
		}
		switch {
		case missed == 0:
			st.Count++
		case missed <= left:
			st.Count++
			st.GraceUsed += missed
		case r.MissAction == "step_back":
			excess := missed - left
			st.GraceUsed = r.Grace
			st.Count -= excess
			if st.Count < 0 {
				st.Count = 0
			}
			st.Count++
		default:
			st.Run++
			st.Count = 1
			st.GraceUsed = 0
			reset = true
		}
	}
	last := idx
	st.Last = &last
	if st.Count > st.Best {
		st.Best = st.Count
	}
	return st, true, reset
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

func streakBroken(st streakState, idx int64, r streakRule) bool {
	if st.Last == nil || st.Count == 0 {
		return false
	}
	missed := int(idx - *st.Last - 1)
	left := r.Grace - st.GraceUsed
	if left < 0 {
		left = 0
	}
	return missed > left
}

func (s *service) applyStreaks(ctx context.Context, tx *sql.Tx, userID, typeID string, now time.Time, clk eventClock, out *eventOutcome) error {
	rows, err := tx.QueryContext(ctx, `SELECT id::text, key, period_days, grace_periods, miss_action, COALESCE(cycle_length,0) FROM economy_streak WHERE event_type_id=$1 AND is_active ORDER BY sort, created_at, id`, typeID)
	if err != nil {
		return err
	}
	type streak struct {
		id, key, miss      string
		days, grace, cycle int
	}
	var list []streak
	for rows.Next() {
		var st streak
		if err := rows.Scan(&st.id, &st.key, &st.days, &st.grace, &st.miss, &st.cycle); err != nil {
			rows.Close()
			return err
		}
		list = append(list, st)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, sk := range list {
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_streak (streak_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, sk.id, userID); err != nil {
			return err
		}
		var st streakState
		var last sql.NullInt64
		if err := tx.QueryRowContext(ctx, `SELECT count, best, run, grace_used, last_period FROM economy_user_streak WHERE streak_id=$1 AND user_id=$2 FOR UPDATE`, sk.id, userID).
			Scan(&st.Count, &st.Best, &st.Run, &st.GraceUsed, &last); err != nil {
			return err
		}
		if last.Valid {
			v := last.Int64
			st.Last = &v
		}
		idx := clk.streakIndex(now, sk.days)
		next, advanced, reset := advanceStreak(st, idx, streakRule{Grace: sk.grace, MissAction: sk.miss})
		if !advanced {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_streak SET count=$3, best=$4, run=$5, grace_used=$6, last_period=$7,
			started_at=CASE WHEN $8::boolean OR last_period IS NULL THEN now() ELSE started_at END, updated_at=now() WHERE streak_id=$1 AND user_id=$2`,
			sk.id, userID, next.Count, next.Best, next.Run, next.GraceUsed, *next.Last, reset); err != nil {
			return err
		}
		out.Streaks = append(out.Streaks, streakChange{StreakID: sk.id, Key: sk.key, Count: next.Count, Best: next.Best, Advanced: true, Reset: reset})
		day := streakRewardDay(next.Count, sk.cycle)
		var rewardID, prizeID, mode string
		err := tx.QueryRowContext(ctx, `SELECT id::text, prize_id::text, grant_mode FROM economy_streak_reward WHERE streak_id=$1 AND day=$2`, sk.id, day).Scan(&rewardID, &prizeID, &mode)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		ref, err := insertClaim(ctx, tx, userID, "streak", sk.id, rewardID, "r"+strconv.Itoa(next.Run)+":"+strconv.Itoa(next.Count), prizeID, mode, float64(next.Count))
		if err != nil {
			return err
		}
		if ref != nil {
			out.Claims = append(out.Claims, *ref)
		}
	}
	return nil
}
