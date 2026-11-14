package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"github.com/heroiclabs/nakama/v3/internal/activity"
)

const relationEventKey = "MetaRelation:Completed"

type relationTarget struct {
	MetaID string  `json:"metaId"`
	Action string  `json:"action"`
	Amount float64 `json:"amount"`
}

type relationDef struct {
	ID           string
	Key          string
	MetaID       string
	Operator     string
	Value        float64
	Targets      []relationTarget
	PrizeType    string
	PrizeID      string
	PrizeGroupID string
	GrantMode    string
	Repeat       string
	MaxFires     int
	Audience     []string
}

type relationChange struct {
	RelationID string `json:"relationId"`
	Key        string `json:"key"`
	Fires      int    `json:"fires"`
	ClaimID    string `json:"claimId,omitempty"`
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

// relationHolds compares the meta value with the rule threshold at the
// 4-decimal precision the database stores, so 0.1+0.2 equals 0.3.
func relationHolds(op string, current, threshold float64) bool {
	c, t := round4(current), round4(threshold)
	switch op {
	case "gt":
		return c > t
	case "gte":
		return c >= t
	case "lt":
		return c < t
	case "lte":
		return c <= t
	case "eq":
		return c == t
	}
	return false
}

// relationDue decides what one evaluation does: fire the relation, re-arm it
// (rearm mode, after the condition stopped holding), or nothing.
func relationDue(r relationDef, holds bool, fires int, armed bool) (fire, rearm bool) {
	if !holds {
		return false, r.Repeat == "rearm" && !armed
	}
	if r.MaxFires > 0 && fires >= r.MaxFires {
		return false, false
	}
	switch r.Repeat {
	case "once":
		return fires == 0, false
	case "rearm":
		return armed, false
	case "every":
		return true, false
	}
	return false, false
}

type relationCascade struct {
	visited map[string]bool
}

type relationCascadeKey struct{}
type relationChainKey struct{}

func relationChain(ctx context.Context) int {
	if n, ok := ctx.Value(relationChainKey{}).(int); ok {
		return n
	}
	return 0
}

func (s *service) activeRelations(ctx context.Context, metaIDs []string) ([]relationDef, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, meta_id::text, operator, value::float8, targets::text, prize_type, COALESCE(prize_id::text,''), COALESCE(prize_group_id::text,''),
			grant_mode, repeat, COALESCE(max_fires,0), audience_tag_ids::text
		FROM economy_meta_relation WHERE is_active AND meta_id::text = ANY($1) AND (starts_at IS NULL OR starts_at <= $2) AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY sort, created_at, id`, metaIDs, s.now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []relationDef
	for rows.Next() {
		var r relationDef
		var targets, audience string
		if err := rows.Scan(&r.ID, &r.Key, &r.MetaID, &r.Operator, &r.Value, &targets, &r.PrizeType, &r.PrizeID, &r.PrizeGroupID, &r.GrantMode, &r.Repeat, &r.MaxFires, &audience); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(targets), &r.Targets)
		r.Audience = parseTextArray(audience)
		out = append(out, r)
	}
	return out, rows.Err()
}

// applyMetaRelations runs every active relation whose source meta is in
// metaIDs. A fired relation changes its target metas, issues its prize claim
// and records MetaRelation:Completed; the changed targets are evaluated in
// turn, one chain step deeper, until the admin's chain depth is reached.
func (s *service) applyMetaRelations(ctx context.Context, userID string, metaIDs []string, depth int) *eventOutcome {
	out := newOutcome()
	if len(metaIDs) == 0 {
		return out
	}
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		s.logger.Warn("Economy meta relations for %s skipped: %v", userID, err)
		return out
	}
	chain := relationChain(ctx)
	if chain > settings.RelationChainDepth {
		s.logger.Warn("Economy meta relation chain for %s stopped at depth %d", userID, chain)
		return out
	}
	rels, err := s.activeRelations(ctx, metaIDs)
	if err != nil {
		s.logger.Warn("Economy meta relations for %s not loaded: %v", userID, err)
		return out
	}
	if len(rels) == 0 {
		return out
	}
	cascade, ok := ctx.Value(relationCascadeKey{}).(*relationCascade)
	if !ok {
		cascade = &relationCascade{visited: map[string]bool{}}
		ctx = context.WithValue(ctx, relationCascadeKey{}, cascade)
	}
	var tags map[string]bool
	userTags := func() (map[string]bool, error) {
		if tags != nil {
			return tags, nil
		}
		t, err := s.userTagIDs(ctx, userID)
		tags = t
		return t, err
	}
	for _, r := range rels {
		if ok, err := audienceAllows(r.Audience, userTags); err != nil || !ok {
			continue
		}
		// A relation fires at most once per cascade; later changes in the same
		// cascade can still re-arm it.
		fired, err := s.fireRelation(ctx, userID, r, settings, !cascade.visited[r.ID], out)
		if err != nil {
			s.logger.Warn("Economy meta relation %s for %s failed: %v", r.Key, userID, err)
			continue
		}
		if fired == nil {
			continue
		}
		cascade.visited[r.ID] = true
		next := context.WithValue(ctx, relationChainKey{}, chain+1)
		activity.Safe(next, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindEvent, Action: "meta_relation", RefID: r.Key, Actor: "system", Data: map[string]any{"relation": r.Key, "fires": fired.change.Fires}})
		for i := range fired.claims {
			c := &fired.claims[i]
			if c.Mode == "auto" {
				extra, err := s.grantClaim(next, c.ClaimID, depth+1)
				if err != nil {
					c.Status = "failed"
					s.logger.Warn("Economy meta relation reward %s not granted yet: %v", c.ClaimID, err)
				} else {
					c.Status = "granted"
					out.merge(extra)
				}
			}
			out.Claims = append(out.Claims, *c)
		}
		if len(fired.changed) > 0 {
			extra, err := s.evaluateLevels(next, userID, fired.changed, depth)
			if err != nil {
				s.logger.Warn("Economy level check after meta relation %s failed: %v", r.Key, err)
			}
			out.merge(extra)
		}
		v := float64(fired.change.Fires)
		ev, err := s.recordEvent(next, eventInput{UserID: userID, Key: relationEventKey, Value: &v, Metadata: map[string]any{"entityId": r.ID, "key": r.Key}, Source: "server", depth: depth + 1})
		if err == nil {
			out.merge(ev)
		} else if app, ok := asApp(err); !ok || app.Code != "EVENT_TYPE_NOT_FOUND" {
			s.logger.Warn("Economy %s for %s not recorded: %v", relationEventKey, userID, err)
		}
		out.Relations = append(out.Relations, fired.change)
	}
	return out
}

type firedRelation struct {
	change  relationChange
	changed []string
	claims  []claimRef
}

func (s *service) fireRelation(ctx context.Context, userID string, r relationDef, settings eventSettings, canFire bool, out *eventOutcome) (*firedRelation, error) {
	var res *firedRelation
	metas := map[string]float64{}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		res, metas = nil, map[string]float64{}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_meta_relation (relation_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, r.ID, userID); err != nil {
			if isForeignKeyViolation(err) {
				return nil
			}
			return err
		}
		var fires int
		var armed bool
		if err := tx.QueryRowContext(ctx, `SELECT fires, armed FROM economy_user_meta_relation WHERE relation_id=$1 AND user_id=$2 FOR UPDATE`, r.ID, userID).Scan(&fires, &armed); err != nil {
			return err
		}
		var value float64
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT value FROM economy_user_meta WHERE user_id=$1 AND meta_id=$2), (SELECT default_value FROM economy_meta WHERE id=$2), 0)::float8`, userID, r.MetaID).Scan(&value); err != nil {
			return err
		}
		fire, rearm := relationDue(r, relationHolds(r.Operator, value, r.Value), fires, armed)
		if rearm {
			_, err := tx.ExecContext(ctx, `UPDATE economy_user_meta_relation SET armed=true, updated_at=now() WHERE relation_id=$1 AND user_id=$2`, r.ID, userID)
			return err
		}
		if !fire || !canFire {
			return nil
		}
		prizeID := ""
		switch r.PrizeType {
		case "prize":
			prizeID = r.PrizeID
		case "prizeGroup":
			members, err := loadGroupMembers(ctx, tx, r.PrizeGroupID)
			if err != nil {
				return err
			}
			if picked, err := drawGroupMember(members); err == nil {
				prizeID = picked
			} else if settings.RelationPrizeMissing == "wait" {
				s.logger.Warn("Economy meta relation %s waits: its prize group has no prizes", r.Key)
				return nil
			} else {
				s.logger.Warn("Economy meta relation %s completes without a prize: its prize group has no prizes", r.Key)
			}
		}
		f := &firedRelation{change: relationChange{RelationID: r.ID, Key: r.Key, Fires: fires + 1}}
		seen := map[string]bool{}
		for _, t := range r.Targets {
			var q string
			if t.Action == "set" {
				q = `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) SELECT $1, id, $3::numeric FROM economy_meta WHERE id=$2
					ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value, updated_at=now() RETURNING u.value::float8, (SELECT key FROM economy_meta WHERE id=$2)`
			} else {
				q = `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) SELECT $1, id, default_value + $3::numeric FROM economy_meta WHERE id=$2
					ON CONFLICT (user_id, meta_id) DO UPDATE SET value=u.value + $3::numeric, updated_at=now() RETURNING u.value::float8, (SELECT key FROM economy_meta WHERE id=$2)`
			}
			var now float64
			var key string
			err := tx.QueryRowContext(ctx, q, userID, t.MetaID, formatAmount(t.Amount)).Scan(&now, &key)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			metas[key] = now
			if !seen[t.MetaID] {
				seen[t.MetaID] = true
				f.changed = append(f.changed, t.MetaID)
			}
		}
		if prizeID != "" {
			ref, err := insertClaim(ctx, tx, userID, "relation", r.ID, r.ID, "F"+strconv.Itoa(fires+1)+"@"+strconv.FormatInt(s.now().UnixNano(), 36), prizeID, r.GrantMode, value)
			if err != nil {
				return err
			}
			if ref != nil {
				f.claims = append(f.claims, *ref)
				f.change.ClaimID = ref.ClaimID
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_meta_relation SET fires=fires+1, armed=false, first_done_at=COALESCE(first_done_at, now()), last_done_at=now(), updated_at=now()
			WHERE relation_id=$1 AND user_id=$2`, r.ID, userID); err != nil {
			return err
		}
		res = f
		return nil
	})
	if err != nil {
		return nil, err
	}
	for k, v := range metas {
		out.Metas[k] = v
	}
	return res, nil
}

// grantEvaluatesMetas lists the grant sources whose callers already run the
// meta-change checks (claims, tasks, achievements); every other grant
// (purchase, gift, admin, game) runs them inside grant itself.
var grantEvaluatesMetas = map[string]bool{"event": true, "level": true, "streak": true, "referral": true, "task": true, "achievement": true, "chance": true, "relation": true}
