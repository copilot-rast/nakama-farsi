package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) metaRelationRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_meta_relations": s.rpcMetaRelations,
	}
}

// rpcMetaRelations lists the relations a player can see: the ones they
// completed (with their prize claims) and, unless completedOnly is set, the
// active ones still waiting. Conditions on server-only metas stay hidden.
func (s *service) rpcMetaRelations(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		CompletedOnly bool   `json:"completedOnly"`
		Lang          string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	tags, err := s.userTagIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	rows, err := s.db.QueryContext(ctx, `SELECT r.id::text, r.key, r.name::text, r.description::text, r.operator, r.value::float8, r.repeat, COALESCE(r.max_fires,0), r.grant_mode, r.prize_type,
			COALESCE(r.prize_id::text,''), r.audience_tag_ids::text, r.ends_at, r.is_active AND (r.starts_at IS NULL OR r.starts_at <= $2) AND (r.ends_at IS NULL OR r.ends_at > $2),
			m.key, m.name::text, m.server_only, COALESCE(u.value, m.default_value)::float8, COALESCE(s.fires,0), s.last_done_at
		FROM economy_meta_relation r JOIN economy_meta m ON m.id=r.meta_id
		LEFT JOIN economy_user_meta u ON u.meta_id=m.id AND u.user_id=$1
		LEFT JOIN economy_user_meta_relation s ON s.relation_id=r.id AND s.user_id=$1
		WHERE r.client_visible AND (COALESCE(s.fires,0) > 0 OR (r.is_active AND (r.starts_at IS NULL OR r.starts_at <= $2) AND (r.ends_at IS NULL OR r.ends_at > $2)))
		ORDER BY r.sort, r.created_at, r.id`, userID, now)
	if err != nil {
		return nil, err
	}
	type rel struct {
		id, key, name, desc, op, repeat, mode, prizeType, prizeID, audience, metaKey, metaName string
		value, current                                                                         float64
		maxFires, fires                                                                        int
		endsAt, lastDone                                                                       sql.NullTime
		live, serverOnly                                                                       bool
	}
	var list []rel
	for rows.Next() {
		var r rel
		if err := rows.Scan(&r.id, &r.key, &r.name, &r.desc, &r.op, &r.value, &r.repeat, &r.maxFires, &r.mode, &r.prizeType, &r.prizeID, &r.audience, &r.endsAt, &r.live,
			&r.metaKey, &r.metaName, &r.serverOnly, &r.current, &r.fires, &r.lastDone); err != nil {
			rows.Close()
			return nil, err
		}
		if r.fires == 0 {
			if in.CompletedOnly {
				continue
			}
			if ok, _ := audienceAllows(parseTextArray(r.audience), func() (map[string]bool, error) { return tags, nil }); !ok {
				continue
			}
		}
		list = append(list, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ids := []string{}
	prizeIDs := []string{}
	for _, r := range list {
		ids = append(ids, r.id)
		if r.prizeID != "" {
			prizeIDs = append(prizeIDs, r.prizeID)
		}
	}
	claims, err := s.userClaims(ctx, userID, "relation", ids)
	if err != nil {
		return nil, err
	}
	for _, c := range claims {
		prizeIDs = append(prizeIDs, c.PrizeID)
	}
	names, err := s.prizeNames(ctx, langs, lang, prizeIDs)
	if err != nil {
		return nil, err
	}
	items := []map[string]any{}
	ready := []map[string]any{}
	for _, r := range list {
		name := langs.Resolve(localized.Parse(r.name), lang)
		mine := []map[string]any{}
		for i := range claims {
			c := &claims[i]
			if c.OwnerID != r.id {
				continue
			}
			row := map[string]any{"claimId": c.ID, "status": claimStatusWord(c), "grantMode": c.Mode, "prize": prizeJSON(names[c.PrizeID], c.PrizeID), "createdAt": c.CreatedAt}
			mine = append(mine, row)
			if c.Status != "granted" && c.Mode == "claim" {
				ready = append(ready, map[string]any{"claimId": c.ID, "kind": "relation", "source": r.key, "sourceName": name, "prize": prizeJSON(names[c.PrizeID], c.PrizeID), "createdAt": c.CreatedAt})
			}
		}
		done := r.fires > 0 && (r.repeat == "once" || (r.maxFires > 0 && r.fires >= r.maxFires))
		item := map[string]any{"key": r.key, "name": name, "description": langs.Resolve(localized.Parse(r.desc), lang), "repeat": r.repeat, "fires": r.fires,
			"completed": r.fires > 0, "done": done, "active": r.live, "grantMode": r.mode, "claims": mine}
		if r.maxFires > 0 {
			item["maxFires"] = r.maxFires
		}
		if r.prizeType == "prize" && r.prizeID != "" {
			item["prize"] = prizeJSON(names[r.prizeID], r.prizeID)
		} else if r.prizeType == "prizeGroup" {
			item["prizeGroup"] = true
		}
		if !r.serverOnly {
			item["condition"] = map[string]any{"meta": r.metaKey, "metaName": langs.Resolve(localized.Parse(r.metaName), lang), "operator": r.op, "value": r.value, "current": r.current, "holds": relationHolds(r.op, r.current, r.value)}
		}
		if r.lastDone.Valid {
			item["lastDoneAt"] = r.lastDone.Time.UTC().Format(time.RFC3339)
		}
		if r.endsAt.Valid {
			item["endsAt"] = r.endsAt.Time
		}
		items = append(items, item)
	}
	return map[string]any{"lang": lang, "items": items, "ready": ready}, nil
}
