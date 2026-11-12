package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
)

type ruleDef struct {
	ID        string
	Key       string
	Measure   string
	Period    string
	Threshold float64
	Repeat    string
	MaxClaims int
	PrizeID   string
	GrantMode string
	Audience  []string
}

func ruleMeasure(measure string, count int64, sum, max float64) float64 {
	switch measure {
	case "sum":
		return sum
	case "max":
		return max
	}
	return float64(count)
}

func ruleDue(r ruleDef, value float64, issued int, everClaimed bool, totalClaims int) int {
	n := 0
	switch r.Repeat {
	case "once":
		if !everClaimed && issued == 0 && value >= r.Threshold {
			n = 1
		}
	case "period":
		if issued == 0 && value >= r.Threshold {
			n = 1
		}
	case "every":
		n = int(math.Floor(value/r.Threshold+1e-9)) - issued
	}
	if n < 0 {
		n = 0
	}
	if r.MaxClaims > 0 {
		left := r.MaxClaims - totalClaims
		if left < 0 {
			left = 0
		}
		if n > left {
			n = left
		}
	}
	return n
}

func (s *service) activeRules(ctx context.Context, q interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}, typeID string, now time.Time) ([]ruleDef, error) {
	rows, err := q.QueryContext(ctx, `SELECT id::text, key, measure, period, threshold::float8, repeat, COALESCE(max_claims,0), prize_id::text, grant_mode, audience_tag_ids::text
		FROM economy_event_rule WHERE event_type_id=$1 AND is_active AND (starts_at IS NULL OR starts_at <= $2) AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY sort, created_at, id`, typeID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ruleDef
	for rows.Next() {
		var r ruleDef
		var audience string
		if err := rows.Scan(&r.ID, &r.Key, &r.Measure, &r.Period, &r.Threshold, &r.Repeat, &r.MaxClaims, &r.PrizeID, &r.GrantMode, &audience); err != nil {
			return nil, err
		}
		r.Audience = parseTextArray(audience)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *service) applyEventRules(ctx context.Context, tx *sql.Tx, userID, typeID string, value float64, now time.Time, clk eventClock, tags func() (map[string]bool, error), out *eventOutcome) error {
	rules, err := s.activeRules(ctx, tx, typeID, now)
	if err != nil {
		return err
	}
	for _, r := range rules {
		ok, err := audienceAllows(r.Audience, tags)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		start := clk.periodStart(r.Period, now)
		var count int64
		var sum, max float64
		var steps int
		if err := tx.QueryRowContext(ctx, `INSERT INTO economy_event_rule_progress AS p (rule_id, user_id, period_start, count, sum, max, steps, updated_at) VALUES ($1,$2,$3::date,1,$4::numeric,$4::numeric,0,now())
			ON CONFLICT (rule_id, user_id, period_start) DO UPDATE SET count=p.count+1, sum=p.sum+EXCLUDED.sum, max=GREATEST(p.max, EXCLUDED.max), updated_at=now()
			RETURNING count, sum::float8, COALESCE(max,0)::float8, steps`, r.ID, userID, dateKey(start), formatAmount(value)).Scan(&count, &sum, &max, &steps); err != nil {
			return err
		}
		measured := ruleMeasure(r.Measure, count, sum, max)
		var total int
		var ever bool
		if r.Repeat == "once" || r.MaxClaims > 0 {
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM economy_reward_claim WHERE kind='event' AND ref_id=$1 AND user_id=$2`, r.ID, userID).Scan(&total); err != nil {
				return err
			}
			ever = total > 0
		}
		due := ruleDue(r, measured, steps, ever, total)
		prog := ruleProgress{RuleID: r.ID, Key: r.Key, Period: r.Period, PeriodStart: dateKey(start), PeriodEnd: clk.periodEnd(r.Period, now), Measure: r.Measure, Value: measured, Threshold: r.Threshold, Reached: measured >= r.Threshold}
		for i := 1; i <= due; i++ {
			scope := dateKey(start) + "#" + strconv.Itoa(steps+i)
			ref, err := insertClaim(ctx, tx, userID, "event", r.ID, r.ID, scope, r.PrizeID, r.GrantMode, measured)
			if err != nil {
				return err
			}
			if ref != nil {
				out.Claims = append(out.Claims, *ref)
				prog.NewClaims++
			}
		}
		if due > 0 {
			if _, err := tx.ExecContext(ctx, `UPDATE economy_event_rule_progress SET steps=steps+$4 WHERE rule_id=$1 AND user_id=$2 AND period_start=$3::date`, r.ID, userID, dateKey(start), due); err != nil {
				return err
			}
		}
		out.Rules = append(out.Rules, prog)
	}
	return nil
}

func insertClaim(ctx context.Context, tx *sql.Tx, userID, kind, ownerID, refID, scope, prizeID, mode string, value float64) (*claimRef, error) {
	id := uuid.Must(uuid.NewV4()).String()
	var got string
	err := tx.QueryRowContext(ctx, `INSERT INTO economy_reward_claim (id, user_id, kind, owner_id, ref_id, scope, prize_id, grant_mode, value) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::numeric)
		ON CONFLICT (kind, ref_id, user_id, scope) DO NOTHING RETURNING id::text`, id, userID, kind, ownerID, refID, scope, prizeID, mode, formatAmount(value)).Scan(&got)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &claimRef{ClaimID: got, Kind: kind, PrizeID: prizeID, Mode: mode, Status: "ready"}, nil
}

func (s *service) grantClaim(ctx context.Context, claimID string, depth int) (*eventOutcome, error) {
	var userID, kind, prizeID, status string
	err := s.db.QueryRowContext(ctx, `SELECT user_id::text, kind, prize_id::text, status FROM economy_reward_claim WHERE id=$1`, claimID).Scan(&userID, &kind, &prizeID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("CLAIM_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	rewards, err := s.grant(ctx, grantRequest{UserID: userID, PrizeID: prizeID, SourceType: kind, SourceID: claimID, Reason: kind + "_reward"})
	if err != nil {
		msg := err.Error()
		if app, ok := asApp(err); ok {
			msg = app.Code
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_reward_claim SET status='failed', error=$2, attempted_at=now() WHERE id=$1 AND status<>'granted'`, claimID, truncate(msg, 500))
		return nil, err
	}
	res, err := s.db.ExecContext(ctx, `UPDATE economy_reward_claim SET status='granted', error=NULL, granted_at=now() WHERE id=$1 AND status<>'granted'`, claimID)
	if err != nil {
		return nil, err
	}
	out := newOutcome()
	if n, _ := res.RowsAffected(); n == 0 {
		return out, nil
	}
	if kind == "task" {
		s.afterTaskGranted(ctx, claimID)
	}
	var metaKeys []string
	for _, item := range rewards {
		if item.Type == "meta" {
			metaKeys = append(metaKeys, item.Key)
		}
	}
	if len(metaKeys) > 0 && depth <= maxEventDepth {
		ids, err := s.metaIDsByKey(ctx, metaKeys)
		if err != nil {
			return out, nil
		}
		extra, err := s.evaluateLevels(ctx, userID, ids, depth+1)
		if err != nil {
			s.logger.Warn("Economy level check after reward %s failed: %v", claimID, err)
		}
		out.merge(extra)
	}
	return out, nil
}

func (s *service) metaIDsByKey(ctx context.Context, keys []string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM economy_meta WHERE key = ANY($1)`, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *service) claimRewards(ctx context.Context, claimID string) ([]prizeItem, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT up.rewards::text FROM economy_reward_claim c JOIN economy_user_prize up ON up.source_type=c.kind AND up.source_id=c.id::text AND up.prize_id=c.prize_id WHERE c.id=$1 AND up.claimed`, claimID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return []prizeItem{}, nil
	}
	if err != nil {
		return nil, err
	}
	items := []prizeItem{}
	_ = json.Unmarshal([]byte(raw), &items)
	return items, nil
}

func (s *service) claimNextRetry(ctx context.Context) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `UPDATE economy_reward_claim SET attempted_at=now(), attempts=attempts+1 WHERE id = (
			SELECT id FROM economy_reward_claim WHERE status<>'granted' AND grant_mode='auto' AND attempts < 8 AND created_at < now() - INTERVAL '1 minute'
				AND (attempted_at IS NULL OR attempted_at < now() - INTERVAL '2 minutes')
			ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED)
		RETURNING id::text`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *service) retryAutoClaims(ctx context.Context, limit int) int {
	done := 0
	for i := 0; i < limit; i++ {
		id, err := s.claimNextRetry(ctx)
		if err != nil {
			s.logger.Warn("Economy reward retry claim failed: %v", err)
			return done
		}
		if id == "" {
			return done
		}
		if _, err := s.grantClaim(ctx, id, 1); err == nil {
			done++
		}
	}
	return done
}

const eventSweepLease = "economy-event-sweep"

func (s *service) sweepEvents(ctx context.Context) (int64, error) {
	holder := clusterkit.Holder(ctx)
	ok, err := clusterkit.AcquireLease(ctx, s.db, eventSweepLease, holder, 5*time.Minute)
	if err != nil || !ok {
		return 0, err
	}
	defer func() { _ = clusterkit.ReleaseLease(context.Background(), s.db, eventSweepLease, holder) }()
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		return 0, err
	}
	var removed int64
	for i := 0; i < 50; i++ {
		res, err := s.db.ExecContext(ctx, `DELETE FROM economy_player_event WHERE id IN (SELECT id FROM economy_player_event WHERE created_at < now() - $1::int * INTERVAL '1 day' ORDER BY created_at LIMIT 5000)`, settings.RetentionDays)
		if err != nil {
			return removed, err
		}
		n, _ := res.RowsAffected()
		removed += n
		if n < 5000 {
			break
		}
	}
	keep := settings.RetentionDays
	if keep < 35 {
		keep = 35
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_player_event_day WHERE day < current_date - $1::int`, keep); err != nil {
		return removed, err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_event_rule_progress WHERE period_start > DATE '1970-01-01' AND period_start < current_date - $1::int`, keep); err != nil {
		return removed, err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_user_task WHERE window_start IS NOT NULL AND window_start < current_date - $1::int AND status='active'`, keep); err != nil {
		return removed, err
	}
	return removed, nil
}

func (s *service) eventLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	tick := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.retryAutoClaims(ctx, 20)
			if tick%20 == 0 {
				if n, err := s.sweepEvents(ctx); err != nil {
					s.logger.Warn("Economy event sweep failed: %v", err)
				} else if n > 0 {
					s.logger.Info("Economy event sweep removed %d old events", n)
				}
			}
			tick++
		}
	}
}
