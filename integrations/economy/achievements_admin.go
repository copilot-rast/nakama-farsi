package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const achievementJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'hint',t.hint,'iconUrl',t.icon_url,'lockedIconUrl',t.locked_icon_url,
	'categoryId',t.category_id,'tagIds',to_jsonb(t.tag_ids),'rarity',t.rarity,'points',t.points,'isHidden',t.is_hidden,'source',t.source,'eventTypeId',t.event_type_id,'metric',t.metric,
	'valueMin',t.value_min,'valueMax',t.value_max,'taskId',t.task_id,'taskGroupId',t.task_group_id,'notify',t.notify,'sort',t.sort,'isActive',t.is_active,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'tiers',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',r.id,'tier',r.tier,'threshold',r.threshold,'name',r.name,'iconUrl',r.icon_url,'prizeId',r.prize_id,'grantMode',r.grant_mode,
		'players',(SELECT count(*) FROM economy_user_achievement u WHERE u.achievement_id=t.id AND u.tier >= r.tier)) ORDER BY r.tier),'[]'::jsonb) FROM economy_achievement_tier r WHERE r.achievement_id=t.id),
	'unlocked',(SELECT count(*) FROM economy_user_achievement u WHERE u.achievement_id=t.id AND u.tier > 0),
	'players',(SELECT count(*) FROM economy_user_achievement u WHERE u.achievement_id=t.id))`

func achievementResources() []resource {
	return []resource{{
		name:   "achievements",
		table:  "economy_achievement",
		json:   achievementJSON,
		search: []string{"t.key", translatedSearch("t.name")},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean", "source": "t.source = ?", "categoryId": "t.category_id::text = ?", "rarity": "t.rarity = ?",
			"isHidden": "t.is_hidden = ?::boolean", "eventTypeId": "t.event_type_id::text = ?", "taskId": "t.task_id::text = ?", "taskGroupId": "t.task_group_id::text = ?",
			"tiered": "((SELECT count(*) FROM economy_achievement_tier r WHERE r.achievement_id=t.id) > 1) = ?::boolean"},
		save: saveAchievement,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_user_achievement WHERE achievement_id=$1 AND tier > 0)
				OR EXISTS (SELECT 1 FROM economy_reward_claim WHERE kind='achievement' AND owner_id=$1)
				OR EXISTS (SELECT 1 FROM economy_achievement a JOIN economy_prize z ON z.items @> jsonb_build_array(jsonb_build_object('type','achievement','key',a.key)) WHERE a.id=$1)`)
		},
	}}
}

func saveAchievement(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key           string          `json:"key"`
		Name          i18n            `json:"name"`
		Description   i18n            `json:"description"`
		Hint          i18n            `json:"hint"`
		IconURL       string          `json:"iconUrl"`
		LockedIconURL string          `json:"lockedIconUrl"`
		CategoryID    string          `json:"categoryId"`
		TagIDs        []string        `json:"tagIds"`
		Rarity        string          `json:"rarity"`
		Points        int             `json:"points"`
		IsHidden      bool            `json:"isHidden"`
		Source        string          `json:"source"`
		EventTypeID   string          `json:"eventTypeId"`
		Metric        string          `json:"metric"`
		ValueMin      json.RawMessage `json:"valueMin"`
		ValueMax      json.RawMessage `json:"valueMax"`
		TaskID        string          `json:"taskId"`
		TaskGroupID   string          `json:"taskGroupId"`
		Notify        bool            `json:"notify"`
		Sort          int             `json:"sort"`
		IsActive      bool            `json:"isActive"`
		Tiers         []struct {
			Tier      int             `json:"tier"`
			Threshold json.RawMessage `json:"threshold"`
			Name      i18n            `json:"name"`
			IconURL   string          `json:"iconUrl"`
			PrizeID   string          `json:"prizeId"`
			GrantMode string          `json:"grantMode"`
		} `json:"tiers"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if err := keepKey(ctx, tx, "economy_achievement", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	hint, err := cleanI18n(langs, "hint", in.Hint, false, 300)
	if err != nil {
		return err
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	locked, err := cleanURL("lockedIconUrl", in.LockedIconURL)
	if err != nil {
		return err
	}
	category, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, category, "achievement"); err != nil {
		return err
	}
	tags, err := checkTags(ctx, tx, in.TagIDs, "achievement")
	if err != nil {
		return err
	}
	if in.Rarity == "" {
		in.Rarity = "common"
	}
	if err := oneOf("rarity", "INVALID_RARITY", in.Rarity, itemRarities...); err != nil {
		return err
	}
	if in.Points < 0 || in.Points > 100000 {
		return badRequest("INVALID_POINTS", "points", "between 0 and 100000")
	}
	if in.Source == "" {
		in.Source = "manual"
	}
	if err := oneOf("source", "INVALID_SOURCE", in.Source, "event", "task", "manual"); err != nil {
		return err
	}
	if in.Metric == "" {
		in.Metric = "count"
	}
	var typeID, taskID, groupID, vmin, vmax *string
	switch in.Source {
	case "event":
		if typeID, err = refExists(ctx, tx, "economy_event_type", "eventTypeId", "EVENT_TYPE_NOT_FOUND", in.EventTypeID, true); err != nil {
			return err
		}
		if err := oneOf("metric", "INVALID_METRIC", in.Metric, "count", "sum_value", "high_value"); err != nil {
			return err
		}
		if vmin, err = optionalNumeric("valueMin", in.ValueMin); err != nil {
			return err
		}
		if vmax, err = optionalNumeric("valueMax", in.ValueMax); err != nil {
			return err
		}
		if vmin != nil && vmax != nil && numericFloat(*vmin) > numericFloat(*vmax) {
			return badRequest("INVALID_RANGE", "valueMax", "the largest value must not be smaller than the smallest")
		}
	case "task":
		in.Metric = "count"
		if strings.TrimSpace(in.TaskID) != "" && strings.TrimSpace(in.TaskGroupID) != "" {
			return badRequest("TASK_AND_GROUP", "taskGroupId", "link a mission or a mission group, not both")
		}
		if strings.TrimSpace(in.TaskID) == "" && strings.TrimSpace(in.TaskGroupID) == "" {
			return badRequest("TASK_REQUIRED", "taskId", "pick the mission or the mission group that unlocks it")
		}
		if taskID, err = refExists(ctx, tx, "economy_task", "taskId", "TASK_NOT_FOUND", in.TaskID, false); err != nil {
			return err
		}
		if groupID, err = refExists(ctx, tx, "economy_task_group", "taskGroupId", "TASK_GROUP_NOT_FOUND", in.TaskGroupID, false); err != nil {
			return err
		}
	default:
		in.Metric = "count"
	}
	if len(in.Tiers) > maxAchievementTiers {
		return badRequest("TOO_MANY_TIERS", "tiers", "at most 10 tiers")
	}
	if in.Source != "event" && len(in.Tiers) > 1 {
		return badRequest("TIERS_NEED_EVENT", "tiers", "only an achievement counted from events can have several tiers")
	}
	if len(in.Tiers) == 0 {
		if in.Source == "event" {
			return badRequest("TIERS_REQUIRED", "tiers", "add at least one tier with its target")
		}
		in.Tiers = append(in.Tiers, struct {
			Tier      int             `json:"tier"`
			Threshold json.RawMessage `json:"threshold"`
			Name      i18n            `json:"name"`
			IconURL   string          `json:"iconUrl"`
			PrizeID   string          `json:"prizeId"`
			GrantMode string          `json:"grantMode"`
		}{Tier: 1})
	}
	type tierIn struct {
		tier            int
		threshold, mode string
		name            i18n
		icon            string
		prize           *string
	}
	tiers := make([]tierIn, 0, len(in.Tiers))
	seen := map[int]bool{}
	for i, t := range in.Tiers {
		field := "tiers[" + strconv.Itoa(i) + "]"
		if t.Tier == 0 {
			t.Tier = i + 1
		}
		if t.Tier < 1 || t.Tier > maxAchievementTiers {
			return badRequest("INVALID_TIER", field+".tier", "tiers go from 1 to 10")
		}
		if seen[t.Tier] {
			return badRequest("DUPLICATE_TIER", field+".tier", "each tier once")
		}
		seen[t.Tier] = true
		th := "1"
		if in.Source == "event" {
			v, err := optionalNumeric(field+".threshold", t.Threshold)
			if err != nil {
				return err
			}
			if v == nil || numericFloat(*v) <= 0 {
				return badRequest("INVALID_THRESHOLD", field+".threshold", "the target must be greater than zero")
			}
			th = *v
		}
		nm, err := cleanI18n(langs, field+".name", t.Name, false, 64)
		if err != nil {
			return err
		}
		ic, err := cleanURL(field+".iconUrl", t.IconURL)
		if err != nil {
			return err
		}
		prize, err := refExists(ctx, tx, "economy_prize", field+".prizeId", "PRIZE_NOT_FOUND", t.PrizeID, false)
		if err != nil {
			return err
		}
		mode := t.GrantMode
		if mode == "" {
			mode = "claim"
		}
		if err := oneOf(field+".grantMode", "INVALID_GRANT_MODE", mode, "auto", "claim"); err != nil {
			return err
		}
		tiers = append(tiers, tierIn{tier: t.Tier, threshold: th, mode: mode, name: nm, icon: ic, prize: prize})
	}
	sort.Slice(tiers, func(i, j int) bool { return tiers[i].tier < tiers[j].tier })
	for i, t := range tiers {
		if t.tier != i+1 {
			return badRequest("TIERS_NOT_CONSECUTIVE", "tiers", "tiers go 1, 2, 3 … without gaps")
		}
		if i > 0 && numericFloat(t.threshold) <= numericFloat(tiers[i-1].threshold) {
			return badRequest("THRESHOLDS_NOT_ORDERED", "tiers["+strconv.Itoa(i)+"].threshold", "a higher tier needs a larger target")
		}
	}
	if !create {
		var used int
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(max(r.tier),0) FROM economy_achievement_tier r WHERE r.achievement_id=$1 AND r.tier > $2
			AND EXISTS (SELECT 1 FROM economy_reward_claim c WHERE c.kind='achievement' AND c.ref_id=r.id)`, id, len(tiers)).Scan(&used); err != nil {
			return err
		}
		if used > 0 {
			return badRequest("TIER_IN_USE", "tiers", "tier "+strconv.Itoa(used)+" has rewards of players and cannot be removed")
		}
	}
	if err := execSave(ctx, tx, "economy_achievement", id, create, []string{"key", "name", "description", "hint", "icon_url", "locked_icon_url", "category_id", "tag_ids", "rarity", "points",
		"is_hidden", "source", "event_type_id", "metric", "value_min", "value_max", "task_id", "task_group_id", "notify", "sort", "is_active"},
		[]any{key, i18nJSON(name), i18nJSON(desc), i18nJSON(hint), icon, locked, category, tags, in.Rarity, in.Points,
			in.IsHidden, in.Source, typeID, in.Metric, vmin, vmax, taskID, groupID, in.Notify, in.Sort, in.IsActive}); err != nil {
		return err
	}
	nums := []int{}
	for _, t := range tiers {
		nums = append(nums, t.tier)
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_achievement_tier (id, achievement_id, tier, threshold, name, icon_url, prize_id, grant_mode) VALUES ($1,$2,$3,$4::numeric,$5::jsonb,$6,$7,$8)
			ON CONFLICT (achievement_id, tier) DO UPDATE SET threshold=EXCLUDED.threshold, name=EXCLUDED.name, icon_url=EXCLUDED.icon_url, prize_id=EXCLUDED.prize_id, grant_mode=EXCLUDED.grant_mode`,
			uuid.Must(uuid.NewV4()).String(), id, t.tier, t.threshold, i18nJSON(t.name), t.icon, t.prize, t.mode); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM economy_achievement_tier WHERE achievement_id=$1 AND NOT (tier = ANY($2))`, id, nums); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE economy_user_achievement SET tier=$2 WHERE achievement_id=$1 AND tier > $2`, id, len(tiers))
	return err
}

func (s *service) achievementConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/users/{userId}/achievements", http.MethodGet, s.consoleUserAchievements},
		{consolePrefix + "/users/{userId}/achievements", http.MethodPost, s.consoleUserAchievementGive},
		{consolePrefix + "/users/{userId}/achievements/{achievementId}", http.MethodDelete, s.consoleUserAchievementRevoke},
		{consolePrefix + "/users/{userId}/achievements/{achievementId}/claim", http.MethodPost, s.consoleUserAchievementClaim},
	}
}

func (s *service) consoleUserAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, r.URL.Query().Get("lang"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	view, err := s.achievementsView(ctx, userID, langs, lang, achievementViewAdmin)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *service) consoleUserAchievementGive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var in struct {
		AchievementID string `json:"achievementId"`
		Tier          int    `json:"tier"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if _, err := uuid.FromString(strings.TrimSpace(in.AchievementID)); err != nil {
		s.writeError(w, notFound("ACHIEVEMENT_NOT_FOUND"))
		return
	}
	a, err := s.findAchievement(ctx, in.AchievementID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.adminGiveAchievement(ctx, userID, a, in.Tier, consoleName(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"achievements": out.Achievements, "claims": out.Claims})
}

func (s *service) consoleAchievementTarget(r *http.Request) (string, *achievementDef, error) {
	userID, err := s.consoleUser(r)
	if err != nil {
		return "", nil, err
	}
	id := pathVar(r, "achievementId")
	if _, err := uuid.FromString(id); err != nil {
		return "", nil, notFound("ACHIEVEMENT_NOT_FOUND")
	}
	a, err := s.findAchievement(r.Context(), id)
	return userID, a, err
}

func (s *service) consoleUserAchievementRevoke(w http.ResponseWriter, r *http.Request) {
	userID, a, err := s.consoleAchievementTarget(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := s.adminRevokeAchievement(r.Context(), userID, a, consoleName(r)); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) consoleUserAchievementClaim(w http.ResponseWriter, r *http.Request) {
	userID, a, err := s.consoleAchievementTarget(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var in struct {
		Tier int `json:"tier"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	done, _, err := s.claimAchievement(r.Context(), userID, a, in.Tier)
	if err != nil {
		s.writeError(w, err)
		return
	}
	claims := []map[string]any{}
	for _, d := range done {
		claims = append(claims, map[string]any{"tier": d.Tier, "claimId": d.ClaimID, "rewards": d.Rewards})
	}
	writeJSON(w, http.StatusOK, map[string]any{"claims": claims})
}
