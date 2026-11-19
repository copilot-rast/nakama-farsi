package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

func hookError(err error) error {
	if r, ok := economyhooks.AsRefusal(err); ok {
		return &appError{Code: r.Code, Message: r.Message, Status: r.Status, Meta: r.Meta}
	}
	return err
}

func (s *service) saveItemHook(ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	err := saveItem(s, ctx, tx, id, create, raw)
	var ae *appError
	if errors.As(err, &ae) {
		return &economyhooks.Refusal{Status: ae.Status, Code: ae.Code, Message: ae.Message, Meta: map[string]any{"field": ae.Field}}
	}
	return err
}

func (s *service) levelIn(ctx context.Context, userID, tableID string) (int, error) {
	return newLevelCache(s.db, userID).level(ctx, tableID)
}

func (s *service) methodSandboxes(ctx context.Context) (map[string]bool, error) {
	active, err := s.activeMethods(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for m, pm := range active {
		out[m] = pm.Config.flag("sandbox")
	}
	return out, nil
}

func itemIDs(defs []itemDef) []string {
	ids := make([]string, 0, len(defs))
	for _, d := range defs {
		ids = append(ids, d.ID)
	}
	return ids
}

func (s *service) grantPrizeHook(ctx context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error) {
	out := economyhooks.PrizeResult{Prizes: []map[string]any{}}
	g.PrizeKey, g.GroupKey = strings.TrimSpace(g.PrizeKey), strings.TrimSpace(g.GroupKey)
	if strings.TrimSpace(g.UserID) == "" || g.SourceType == "" || g.SourceID == "" || (g.PrizeKey == "") == (g.GroupKey == "") {
		return out, errors.New("a prize grant needs a user, a source and either a prize key or a group key")
	}
	type ref struct{ id, key string }
	var list []ref
	if g.PrizeKey != "" {
		var r ref
		err := s.db.QueryRowContext(ctx, `SELECT id::text, key FROM economy_prize WHERE key=$1`, g.PrizeKey).Scan(&r.id, &r.key)
		if errors.Is(err, sql.ErrNoRows) {
			return out, notFound("PRIZE_NOT_FOUND")
		}
		if err != nil {
			return out, err
		}
		list = append(list, r)
	} else {
		rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.key, m.weight FROM economy_prize_group g JOIN economy_prize_group_member m ON m.group_id=g.id
			JOIN economy_prize p ON p.id=m.prize_id WHERE g.key=$1 ORDER BY m.sort, m.prize_id`, g.GroupKey)
		if err != nil {
			return out, err
		}
		var members []ref
		var weights []int
		total := 0
		for rows.Next() {
			var r ref
			var w int
			if err := rows.Scan(&r.id, &r.key, &w); err != nil {
				rows.Close()
				return out, err
			}
			members, weights, total = append(members, r), append(weights, w), total+w
		}
		rows.Close()
		if len(members) == 0 || total <= 0 {
			return out, badRequest("PRIZE_GROUP_EMPTY", "groupKey", "the prize group is missing or has no prizes")
		}
		if g.GroupMode == economyhooks.GroupAll {
			list = members
		} else {
			ids := make([]string, len(members))
			for i, m := range members {
				ids[i] = m.id
			}
			var done string
			err := s.db.QueryRowContext(ctx, `SELECT prize_id::text FROM economy_user_prize WHERE source_type=$1 AND source_id=$2 AND prize_id::text = ANY($3) ORDER BY created_at LIMIT 1`, g.SourceType, g.SourceID, ids).Scan(&done)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return out, err
			}
			sum := sha256.Sum256([]byte(g.SourceType + ":" + g.SourceID + ":" + g.GroupKey))
			roll := int(binary.BigEndian.Uint64(sum[:8]) % uint64(total))
			for i, m := range members {
				if done != "" && m.id == done || done == "" && roll < weights[i] {
					list = append(list, m)
					break
				}
				roll -= weights[i]
			}
		}
	}
	for _, r := range list {
		items, err := s.grant(ctx, grantRequest{UserID: g.UserID, PrizeID: r.id, SourceType: g.SourceType, SourceID: g.SourceID, Reason: g.Reason})
		if err != nil {
			return out, err
		}
		out.Prizes = append(out.Prizes, map[string]any{"id": r.id, "key": r.key, "items": items})
	}
	return out, nil
}

func (s *service) settleMetasHook(ctx context.Context, userID string, metaKeys []string) (map[string]any, error) {
	ids, err := s.metaIDsByKey(ctx, metaKeys)
	if err != nil {
		return nil, err
	}
	out := newOutcome()
	out.metaIDs = ids
	s.settleOutcome(ctx, userID, out, 0)
	return map[string]any{"levels": out.Levels, "claims": out.Claims}, nil
}

func (s *service) levelProgressHook(ctx context.Context, userID, tableKey string) (economyhooks.LevelProgress, error) {
	out := economyhooks.LevelProgress{TableKey: tableKey}
	var tableID string
	err := s.db.QueryRowContext(ctx, `SELECT t.id::text, COALESCE((SELECT value FROM economy_user_meta WHERE user_id=$2 AND meta_id=m.id), m.default_value)::float8,
			COALESCE((SELECT level FROM economy_user_level WHERE user_id=$2 AND table_id=t.id), 0)
		FROM economy_level_table t JOIN economy_meta m ON m.id=t.meta_id WHERE t.key=$1 AND t.is_active`, tableKey, userID).Scan(&tableID, &out.Value, &out.Level)
	if errors.Is(err, sql.ErrNoRows) {
		return out, economyhooks.ErrNoLevelTable
	}
	if err != nil {
		return out, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT level, threshold::float8 FROM economy_level_step WHERE table_id=$1 ORDER BY level`, tableID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	type step struct {
		level     int
		threshold float64
	}
	var steps []step
	for rows.Next() {
		var st step
		if err := rows.Scan(&st.level, &st.threshold); err != nil {
			return out, err
		}
		steps = append(steps, st)
		if st.level > out.MaxLevel {
			out.MaxLevel = st.level
		}
		if out.Level == 0 && st.threshold <= 0 && st.level > out.Level {
			out.Level = st.level
		}
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	for _, st := range steps {
		if st.level == out.Level {
			out.LevelStart = st.threshold
		}
		if st.level > out.Level && (out.Next == nil || st.threshold < *out.Next) {
			v := st.threshold
			out.Next = &v
		}
	}
	return out, nil
}
