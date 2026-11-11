package main

import (
	"context"
	"database/sql"
	"strconv"
	"time"
)

func metaRuleAmount(source string, amount, value float64) float64 {
	if source == "event" {
		return amount * value
	}
	return amount
}

func (s *service) applyMetaRules(ctx context.Context, tx *sql.Tx, userID, typeID string, value float64, now time.Time, tags func() (map[string]bool, error), out *eventOutcome) error {
	rows, err := tx.QueryContext(ctx, `SELECT r.action, r.value_source, r.amount::float8, r.audience_tag_ids::text, m.id::text, m.key
		FROM economy_meta_rule r JOIN economy_meta m ON m.id=r.meta_id
		WHERE r.event_type_id=$1 AND r.is_active AND (r.starts_at IS NULL OR r.starts_at <= $2) AND (r.ends_at IS NULL OR r.ends_at > $2)
		ORDER BY r.sort, r.created_at, r.id`, typeID, now)
	if err != nil {
		return err
	}
	type metaRule struct {
		action, source, audience, metaID, metaKey string
		amount                                    float64
	}
	var list []metaRule
	for rows.Next() {
		var r metaRule
		if err := rows.Scan(&r.action, &r.source, &r.amount, &r.audience, &r.metaID, &r.metaKey); err != nil {
			rows.Close()
			return err
		}
		list = append(list, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, r := range list {
		ok, err := audienceAllows(parseTextArray(r.audience), tags)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		amount := formatAmount(metaRuleAmount(r.source, r.amount, value))
		var q string
		switch r.action {
		case "set":
			q = `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) SELECT $1, id, $3::numeric FROM economy_meta WHERE id=$2
				ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value, updated_at=now() RETURNING u.value::float8`
		case "max":
			q = `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) SELECT $1, id, GREATEST(default_value, $3::numeric) FROM economy_meta WHERE id=$2
				ON CONFLICT (user_id, meta_id) DO UPDATE SET value=GREATEST(u.value, $3::numeric), updated_at=now() RETURNING u.value::float8`
		default:
			q = `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) SELECT $1, id, default_value + $3::numeric FROM economy_meta WHERE id=$2
				ON CONFLICT (user_id, meta_id) DO UPDATE SET value=u.value + $3::numeric, updated_at=now() RETURNING u.value::float8`
		}
		var now float64
		if err := tx.QueryRowContext(ctx, q, userID, r.metaID, amount).Scan(&now); err != nil {
			return err
		}
		out.Metas[r.metaKey] = now
		if !seen[r.metaID] {
			seen[r.metaID] = true
			out.metaIDs = append(out.metaIDs, r.metaID)
		}
	}
	return nil
}

type levelStep struct {
	ID        string
	Level     int
	Threshold float64
	PrizeID   sql.NullString
	GrantMode string
}

func levelFor(steps []levelStep, value float64) int {
	level := 0
	for _, st := range steps {
		if value >= st.Threshold && st.Level > level {
			level = st.Level
		}
	}
	return level
}

func (s *service) evaluateLevels(ctx context.Context, userID string, metaIDs []string, depth int) (*eventOutcome, error) {
	out := newOutcome()
	if len(metaIDs) == 0 || depth > maxEventDepth {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, meta_id::text, COALESCE(level_meta_id::text,''), (SELECT key FROM economy_meta WHERE id=t.level_meta_id)
		FROM economy_level_table t WHERE is_active AND meta_id::text = ANY($1) ORDER BY sort, created_at, id`, metaIDs)
	if err != nil {
		return out, err
	}
	type table struct {
		id, key, metaID, levelMetaID string
		levelMetaKey                 sql.NullString
	}
	var tables []table
	for rows.Next() {
		var t table
		if err := rows.Scan(&t.id, &t.key, &t.metaID, &t.levelMetaID, &t.levelMetaKey); err != nil {
			rows.Close()
			return out, err
		}
		tables = append(tables, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, err
	}
	for _, t := range tables {
		var change *levelChange
		var claims []claimRef
		var levelValue float64
		var threshold float64
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_level (table_id, user_id, level) VALUES ($1,$2,0) ON CONFLICT DO NOTHING`, t.id, userID); err != nil {
				return err
			}
			var cur int
			if err := tx.QueryRowContext(ctx, `SELECT level FROM economy_user_level WHERE table_id=$1 AND user_id=$2 FOR UPDATE`, t.id, userID).Scan(&cur); err != nil {
				return err
			}
			var value float64
			if err := tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT value FROM economy_user_meta WHERE user_id=$1 AND meta_id=$2), (SELECT default_value FROM economy_meta WHERE id=$2), 0)::float8`, userID, t.metaID).Scan(&value); err != nil {
				return err
			}
			srows, err := tx.QueryContext(ctx, `SELECT id::text, level, threshold::float8, prize_id::text, grant_mode FROM economy_level_step WHERE table_id=$1 ORDER BY level`, t.id)
			if err != nil {
				return err
			}
			var steps []levelStep
			for srows.Next() {
				var st levelStep
				if err := srows.Scan(&st.ID, &st.Level, &st.Threshold, &st.PrizeID, &st.GrantMode); err != nil {
					srows.Close()
					return err
				}
				steps = append(steps, st)
			}
			srows.Close()
			if err := srows.Err(); err != nil {
				return err
			}
			next := levelFor(steps, value)
			if next <= cur {
				return nil
			}
			if _, err := tx.ExecContext(ctx, `UPDATE economy_user_level SET level=$3, reached_at=now() WHERE table_id=$1 AND user_id=$2`, t.id, userID, next); err != nil {
				return err
			}
			if t.levelMetaID != "" {
				if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) VALUES ($1,$2,$3)
					ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value, updated_at=now()`, userID, t.levelMetaID, next); err != nil {
					return err
				}
			}
			for _, st := range steps {
				if st.Level <= cur || st.Level > next {
					continue
				}
				if st.Level == next {
					threshold = st.Threshold
				}
				if !st.PrizeID.Valid {
					continue
				}
				ref, err := insertClaim(ctx, tx, userID, "level", t.id, st.ID, "L"+strconv.Itoa(st.Level), st.PrizeID.String, st.GrantMode, float64(st.Level))
				if err != nil {
					return err
				}
				if ref != nil {
					claims = append(claims, *ref)
				}
			}
			change = &levelChange{TableID: t.id, Key: t.key, From: cur, To: next}
			levelValue = float64(next)
			return nil
		})
		if err != nil {
			return out, err
		}
		if change == nil {
			continue
		}
		out.Levels = append(out.Levels, *change)
		out.Claims = append(out.Claims, claims...)
		if t.levelMetaKey.Valid {
			out.Metas[t.levelMetaKey.String] = levelValue
			out.metaIDs = append(out.metaIDs, t.levelMetaID)
		}
		if change.From > 0 || threshold > 0 {
			v := levelValue
			extra, err := s.recordEvent(ctx, eventInput{UserID: userID, Key: "level_up", Value: &v, Metadata: map[string]any{"table": t.key, "from": change.From}, Source: "server", depth: depth + 1})
			if err == nil {
				out.merge(extra)
			} else if app, ok := asApp(err); !ok || app.Code != "EVENT_TYPE_NOT_FOUND" {
				s.logger.Warn("Economy level_up event for %s not recorded: %v", userID, err)
			}
		}
	}
	s.settleOutcome(ctx, userID, out, depth)
	return out, nil
}
