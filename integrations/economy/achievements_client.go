package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	achievementViewPlayer = iota
	achievementViewAdmin
)

func (s *service) achievementRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_achievements":         s.rpcAchievements,
		"economy_achievement_claim":    s.rpcAchievementClaim,
		"economy_achievements_profile": s.rpcAchievementsProfile,
	}
}

type achievementLookups struct {
	events map[string][2]string
	tasks  map[string][2]string
	groups map[string][2]string
	cats   map[string]string
	tags   map[string]string
	prizes map[string]prizeName
}

func (s *service) achievementLookups(ctx context.Context, langs *localized.Set, lang string, defs []*achievementDef) (*achievementLookups, error) {
	lk := &achievementLookups{events: map[string][2]string{}, tasks: map[string][2]string{}, groups: map[string][2]string{}, cats: map[string]string{}, tags: map[string]string{}}
	load := func(q string, into map[string][2]string) error {
		rows, err := s.db.QueryContext(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, key, name string
			if err := rows.Scan(&id, &key, &name); err != nil {
				return err
			}
			into[id] = [2]string{key, langs.Resolve(localized.Parse(name), lang)}
		}
		return rows.Err()
	}
	need := map[string]bool{}
	prizeIDs := []string{}
	for _, a := range defs {
		need[a.Source] = true
		for _, t := range a.Tiers {
			if t.PrizeID != "" {
				prizeIDs = append(prizeIDs, t.PrizeID)
			}
		}
	}
	if need["event"] {
		if err := load(`SELECT id::text, key, name::text FROM economy_event_type`, lk.events); err != nil {
			return nil, err
		}
	}
	if need["task"] {
		if err := load(`SELECT id::text, key, name::text FROM economy_task WHERE id IN (SELECT task_id FROM economy_achievement WHERE task_id IS NOT NULL)`, lk.tasks); err != nil {
			return nil, err
		}
		if err := load(`SELECT id::text, key, name::text FROM economy_task_group WHERE id IN (SELECT task_group_id FROM economy_achievement WHERE task_group_id IS NOT NULL)`, lk.groups); err != nil {
			return nil, err
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT 'c', id::text, key FROM base_category WHERE key_type='achievement' UNION ALL SELECT 't', id::text, key FROM base_tag WHERE key_type IN ('achievement','general')`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var kind, id, key string
		if err := rows.Scan(&kind, &id, &key); err != nil {
			rows.Close()
			return nil, err
		}
		if kind == "c" {
			lk.cats[id] = key
		} else {
			lk.tags[id] = key
		}
	}
	rows.Close()
	if lk.prizes, err = s.prizeNames(ctx, langs, lang, prizeIDs); err != nil {
		return nil, err
	}
	return lk, nil
}

func nullTime(t sql.NullTime) any {
	if !t.Valid {
		return nil
	}
	return t.Time.UTC().Format(time.RFC3339)
}

func (s *service) achievementItem(a *achievementDef, row userAchievementRow, has bool, claims map[string]achievementClaimRow, lk *achievementLookups, langs *localized.Set, lang string, mode int) map[string]any {
	text := func(raw string) string { return langs.Resolve(localized.Parse(raw), lang) }
	tier := 0
	if has {
		tier = row.Tier
	}
	unlocked := tier > 0
	if a.Hidden && !unlocked && mode == achievementViewPlayer {
		return map[string]any{"id": a.ID, "hidden": true, "state": "hidden", "unlocked": false, "rarity": a.Rarity, "points": a.Points, "hint": text(a.Hint),
			"lockedIconUrl": a.LockedIconURL, "categoryKey": lk.cats[a.CategoryID], "maxTier": a.maxTier(), "tier": 0, "claimable": []int{}, "sort": a.Sort}
	}
	value := 0.0
	if a.Source == "event" {
		if has {
			value = taskMeasure(a.Metric, row.Count, row.Sum, row.Max)
		}
	} else if unlocked {
		value = 1
	}
	tiers := []map[string]any{}
	claimable := []int{}
	var next map[string]any
	tierName := ""
	for _, t := range a.Tiers {
		status := "locked"
		tv := map[string]any{"tier": t.Tier, "name": text(t.Name), "threshold": t.Threshold, "iconUrl": t.IconURL, "grantMode": t.GrantMode}
		if t.PrizeID != "" {
			tv["prize"] = prizeJSON(lk.prizes[t.PrizeID], t.PrizeID)
		}
		if t.Tier <= tier {
			status = "reached"
			if t.PrizeID != "" {
				c, ok := claims[t.ID]
				switch {
				case ok && c.Status == "granted":
					status = "claimed"
				case ok && c.Mode == "auto":
					status = "pending"
				case !ok && t.GrantMode == "auto":
					status = "pending"
				default:
					status = "claimable"
					claimable = append(claimable, t.Tier)
				}
				if ok {
					tv["claimId"] = c.ID
				}
			}
			if t.Tier == tier {
				tierName = text(t.Name)
			}
		} else if next == nil {
			next = map[string]any{"tier": t.Tier, "threshold": t.Threshold, "name": text(t.Name)}
		}
		tv["status"] = status
		tiers = append(tiers, tv)
	}
	progress := 0.0
	if next == nil {
		progress = 1
	} else if a.Source == "event" {
		goal := next["threshold"].(float64)
		if goal > 0 {
			progress = value / goal
		}
		if progress > 1 {
			progress = 1
		}
		if progress < 0 {
			progress = 0
		}
	}
	state := "locked"
	if unlocked {
		state = "unlocked"
		if next == nil {
			state = "completed"
		}
	}
	icon := a.IconURL
	if !unlocked && a.LockedIconURL != "" {
		icon = a.LockedIconURL
	}
	tags := []string{}
	for _, id := range a.TagIDs {
		if k := lk.tags[id]; k != "" {
			tags = append(tags, k)
		}
	}
	out := map[string]any{
		"id": a.ID, "key": a.Key, "name": text(a.Name), "description": text(a.Description), "hint": text(a.Hint), "iconUrl": a.IconURL, "lockedIconUrl": a.LockedIconURL, "displayIconUrl": icon,
		"categoryKey": lk.cats[a.CategoryID], "tags": tags, "rarity": a.Rarity, "points": a.Points, "hidden": a.Hidden, "source": a.Source, "metric": a.Metric,
		"state": state, "unlocked": unlocked, "tier": tier, "maxTier": a.maxTier(), "tierName": tierName, "tiered": len(a.Tiers) > 1, "value": value, "progress": progress, "next": next,
		"tiers": tiers, "claimable": claimable, "sort": a.Sort,
	}
	if a.ValueMin != nil {
		out["valueMin"] = *a.ValueMin
	}
	if a.ValueMax != nil {
		out["valueMax"] = *a.ValueMax
	}
	if ev, ok := lk.events[a.EventTypeID]; ok && a.Source == "event" {
		out["eventKey"], out["eventName"] = ev[0], ev[1]
	}
	if tk, ok := lk.tasks[a.TaskID]; ok {
		out["taskKey"], out["taskName"] = tk[0], tk[1]
	}
	if g, ok := lk.groups[a.GroupID]; ok {
		out["groupKey"], out["groupName"] = g[0], g[1]
	}
	if has {
		out["unlockedAt"] = nullTime(row.UnlockedAt)
		out["tierAt"] = nullTime(row.TierAt)
	}
	if mode == achievementViewAdmin {
		out["isActive"] = a.Active
		if has {
			out["unlockedBy"] = row.Source
			out["unlockedRef"] = row.SourceRef
		}
	}
	return out
}

func (s *service) achievementsView(ctx context.Context, userID string, langs *localized.Set, lang string, mode int) (map[string]any, error) {
	where := `WHERE a.is_active`
	if mode == achievementViewAdmin {
		where = ``
	}
	defs, err := loadAchievements(ctx, s.db, where)
	if err != nil {
		return nil, err
	}
	rows, err := s.userAchievementRows(ctx, userID)
	if err != nil {
		return nil, err
	}
	claims, err := s.achievementClaims(ctx, userID)
	if err != nil {
		return nil, err
	}
	tiers := map[string]int{}
	for id, r := range rows {
		tiers[id] = r.Tier
	}
	if s.healAchievementClaims(ctx, userID, defs, tiers, claims) {
		if claims, err = s.achievementClaims(ctx, userID); err != nil {
			return nil, err
		}
	}
	lk, err := s.achievementLookups(ctx, langs, lang, defs)
	if err != nil {
		return nil, err
	}
	items := []map[string]any{}
	points, unlocked, ready, total := 0, 0, 0, 0
	for _, a := range defs {
		row, has := rows[a.ID]
		it := s.achievementItem(a, row, has, claims, lk, langs, lang, mode)
		items = append(items, it)
		if a.Active {
			total++
			if it["unlocked"] == true {
				unlocked++
				points += a.Points
			}
			if c, ok := it["claimable"].([]int); ok {
				ready += len(c)
			}
		}
	}
	return map[string]any{"lang": lang, "points": points, "unlocked": unlocked, "total": total, "ready": ready, "items": items}, nil
}

func (s *service) rpcAchievements(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang     string `json:"lang"`
		Category string `json:"category"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	view, err := s.achievementsView(ctx, userID, langs, lang, achievementViewPlayer)
	if err != nil {
		return nil, err
	}
	if in.Category != "" {
		kept := []map[string]any{}
		for _, it := range view["items"].([]map[string]any) {
			if it["categoryKey"] == in.Category {
				kept = append(kept, it)
			}
		}
		view["items"] = kept
	}
	return view, nil
}

func (s *service) rpcAchievementClaim(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key  string `json:"key"`
		ID   string `json:"id"`
		Tier int    `json:"tier"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	ref := in.ID
	if ref == "" {
		ref = in.Key
	}
	a, err := s.findAchievement(ctx, ref)
	if err != nil {
		return nil, err
	}
	if !a.Active {
		return nil, notFound("ACHIEVEMENT_NOT_FOUND")
	}
	done, extra, err := s.claimAchievement(ctx, userID, a, in.Tier)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, d := range done {
		ids = append(ids, d.PrizeID)
	}
	names, err := s.prizeNames(ctx, langs, lang, ids)
	if err != nil {
		return nil, err
	}
	claims := []map[string]any{}
	for _, d := range done {
		claims = append(claims, map[string]any{"tier": d.Tier, "claimId": d.ClaimID, "rewards": d.Rewards, "prize": prizeJSON(names[d.PrizeID], d.PrizeID)})
	}
	rows, err := s.userAchievementRows(ctx, userID)
	if err != nil {
		return nil, err
	}
	cl, err := s.achievementClaims(ctx, userID)
	if err != nil {
		return nil, err
	}
	fresh, err := s.findAchievement(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	lk, err := s.achievementLookups(ctx, langs, lang, []*achievementDef{fresh})
	if err != nil {
		return nil, err
	}
	row, has := rows[a.ID]
	out := map[string]any{"lang": lang, "kind": "achievement", "claims": claims, "achievement": s.achievementItem(fresh, row, has, cl, lk, langs, lang, achievementViewPlayer), "levels": []levelChange{}}
	if extra != nil {
		out["levels"] = extra.Levels
	}
	return out, nil
}

func (s *service) rpcAchievementsProfile(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserID   string `json:"userId"`
		Username string `json:"username"`
		Lang     string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	var target, username, display, avatar string
	var err error
	switch {
	case strings.TrimSpace(in.UserID) != "":
		if _, uerr := uuid.FromString(strings.TrimSpace(in.UserID)); uerr != nil {
			return nil, notFound("USER_NOT_FOUND")
		}
		err = s.db.QueryRowContext(ctx, `SELECT id::text, username, COALESCE(display_name,''), COALESCE(avatar_url,'') FROM users WHERE id=$1 AND disable_time = '1970-01-01 00:00:00 UTC'`, strings.TrimSpace(in.UserID)).Scan(&target, &username, &display, &avatar)
	case strings.TrimSpace(in.Username) != "":
		err = s.db.QueryRowContext(ctx, `SELECT id::text, username, COALESCE(display_name,''), COALESCE(avatar_url,'') FROM users WHERE username=$1 AND disable_time = '1970-01-01 00:00:00 UTC'`, strings.TrimSpace(in.Username)).Scan(&target, &username, &display, &avatar)
	default:
		target = userID
		err = s.db.QueryRowContext(ctx, `SELECT username, COALESCE(display_name,''), COALESCE(avatar_url,'') FROM users WHERE id=$1`, userID).Scan(&username, &display, &avatar)
	}
	if errors.Is(err, sql.ErrNoRows) || target == "00000000-0000-0000-0000-000000000000" {
		return nil, notFound("USER_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	defs, err := loadAchievements(ctx, s.db, `WHERE a.is_active`)
	if err != nil {
		return nil, err
	}
	rows, err := s.userAchievementRows(ctx, target)
	if err != nil {
		return nil, err
	}
	lk, err := s.achievementLookups(ctx, langs, lang, nil)
	if err != nil {
		return nil, err
	}
	text := func(raw string) string { return langs.Resolve(localized.Parse(raw), lang) }
	items := []map[string]any{}
	points, total := 0, 0
	for _, a := range defs {
		row, has := rows[a.ID]
		open := has && row.Tier > 0
		if !a.Hidden || open {
			total++
		}
		if !open {
			continue
		}
		points += a.Points
		tierName := ""
		if t := a.tier(row.Tier); t != nil {
			tierName = text(t.Name)
		}
		items = append(items, map[string]any{"key": a.Key, "name": text(a.Name), "description": text(a.Description), "iconUrl": a.IconURL, "rarity": a.Rarity, "points": a.Points,
			"hidden": a.Hidden, "tier": row.Tier, "maxTier": a.maxTier(), "tierName": tierName, "categoryKey": lk.cats[a.CategoryID], "unlockedAt": nullTime(row.UnlockedAt)})
	}
	return map[string]any{"lang": lang, "user": map[string]any{"userId": target, "username": username, "displayName": display, "avatarUrl": avatar, "self": target == userID},
		"points": points, "unlocked": len(items), "total": total, "items": items}, nil
}
