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
