package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const taskJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'iconUrl',t.icon_url,'categoryId',t.category_id,'tagIds',to_jsonb(t.tag_ids),
	'condition',t.condition,'eventTypeId',t.event_type_id,'metric',t.metric,'target',t.target,'valueMin',t.value_min,'valueMax',t.value_max,'repeat',t.repeat,'repeatMinutes',t.repeat_minutes,
	'maxCompletions',t.max_completions,'parentId',t.parent_id,'dependsOn',to_jsonb(t.depends_on),'groupId',t.group_id,'prizeId',t.prize_id,'grantMode',t.grant_mode,'notify',t.notify,
	'audienceTagIds',to_jsonb(t.audience_tag_ids),'levelTableId',t.level_table_id,'minLevel',t.min_level,'maxLevel',t.max_level,'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,
	'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,'children',(SELECT count(*) FROM economy_task c WHERE c.parent_id=t.id),
	'players',(SELECT count(DISTINCT u.user_id) FROM economy_user_task u WHERE u.task_id=t.id),'completed',(SELECT count(*) FROM economy_user_task u WHERE u.task_id=t.id AND u.status<>'active'))`

const taskGroupJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'iconUrl',t.icon_url,'categoryId',t.category_id,'activeCount',t.active_count,
	'orderMode',t.order_mode,'handout',t.handout,'handoutMinutes',t.handout_minutes,'isRepeatable',t.is_repeatable,'audienceTagIds',to_jsonb(t.audience_tag_ids),'levelTableId',t.level_table_id,
	'minLevel',t.min_level,'maxLevel',t.max_level,'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'taskIds',(SELECT COALESCE(jsonb_agg(m.id ORDER BY m.sort, m.created_at),'[]'::jsonb) FROM economy_task m WHERE m.group_id=t.id AND m.parent_id IS NULL),
	'players',(SELECT count(*) FROM economy_user_task_group u WHERE u.group_id=t.id))`

func taskResources() []resource {
	return []resource{
		{
			name:   "tasks",
			table:  "economy_task",
			json:   taskJSON,
			search: []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "groupId": "t.group_id::text = ?", "eventTypeId": "t.event_type_id::text = ?", "categoryId": "t.category_id::text = ?",
				"prizeId": "t.prize_id::text = ?", "parentId": "t.parent_id::text = ?", "repeat": "t.repeat = ?", "level": "(CASE ? WHEN 'sub' THEN t.parent_id IS NOT NULL ELSE t.parent_id IS NULL END)"},
			save: saveTask,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				var kids bool
				if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_task WHERE parent_id=$1)`, id).Scan(&kids); err != nil {
					return err
				}
				if kids {
					return conflict("HAS_SUBTASKS", "delete or move the sub-tasks first")
				}
				if err := refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_reward_claim WHERE kind='task' AND owner_id=$1) OR EXISTS (SELECT 1 FROM economy_user_task WHERE task_id=$1 AND status<>'active')`); err != nil {
					return err
				}
				_, err := tx.ExecContext(ctx, `UPDATE economy_task SET depends_on=array_remove(depends_on, $1::text), updated_at=now() WHERE $1::text = ANY(depends_on)`, id)
				return err
			},
		},
		{
			name:    "task-groups",
			table:   "economy_task_group",
			json:    taskGroupJSON,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "handout": "t.handout = ?", "categoryId": "t.category_id::text = ?"},
			save:    saveTaskGroup,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_user_task_group WHERE group_id=$1 AND cardinality(handed) > 0)`)
			},
		},
	}
}

type levelTarget struct {
	table    *string
	min, max *int
}

func cleanLevelTarget(ctx context.Context, tx *sql.Tx, tableID string, min, max *int) (levelTarget, error) {
	var out levelTarget
	table, err := refExists(ctx, tx, "economy_level_table", "levelTableId", "LEVEL_TABLE_NOT_FOUND", tableID, false)
	if err != nil {
		return out, err
	}
	if min != nil && *min < 0 {
		return out, badRequest("INVALID_LEVEL", "minLevel", "zero or more")
	}
	if max != nil && *max < 0 {
		return out, badRequest("INVALID_LEVEL", "maxLevel", "zero or more")
	}
	if min != nil && max != nil && *min > *max {
		return out, badRequest("INVALID_LEVEL_RANGE", "maxLevel", "the highest level must not be below the lowest")
	}
	if (min != nil || max != nil) && table == nil {
		return out, badRequest("LEVEL_TABLE_REQUIRED", "levelTableId", "pick the level table the levels belong to")
	}
	if min == nil && max == nil {
		table = nil
	}
	return levelTarget{table: table, min: min, max: max}, nil
}

func saveTask(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key            string          `json:"key"`
		Name           i18n            `json:"name"`
		Description    i18n            `json:"description"`
		IconURL        string          `json:"iconUrl"`
		CategoryID     string          `json:"categoryId"`
		TagIDs         []string        `json:"tagIds"`
		Condition      string          `json:"condition"`
		EventTypeID    string          `json:"eventTypeId"`
		Metric         string          `json:"metric"`
		Target         json.RawMessage `json:"target"`
		ValueMin       json.RawMessage `json:"valueMin"`
		ValueMax       json.RawMessage `json:"valueMax"`
		Repeat         string          `json:"repeat"`
		RepeatMinutes  *int            `json:"repeatMinutes"`
		MaxCompletions *int            `json:"maxCompletions"`
		ParentID       string          `json:"parentId"`
		DependsOn      []string        `json:"dependsOn"`
		GroupID        string          `json:"groupId"`
		PrizeID        string          `json:"prizeId"`
		GrantMode      string          `json:"grantMode"`
		Notify         bool            `json:"notify"`
		AudienceTagIDs []string        `json:"audienceTagIds"`
		LevelTableID   string          `json:"levelTableId"`
		MinLevel       *int            `json:"minLevel"`
		MaxLevel       *int            `json:"maxLevel"`
		StartsAt       string          `json:"startsAt"`
		EndsAt         string          `json:"endsAt"`
		IsActive       bool            `json:"isActive"`
		Sort           int             `json:"sort"`
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
	if err := keepKey(ctx, tx, "economy_task", id, create, key); err != nil {
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
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	category, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, category, "task"); err != nil {
		return err
	}
	tags, err := checkTags(ctx, tx, in.TagIDs, "task")
	if err != nil {
		return err
	}
	if in.Condition == "" {
		in.Condition = "event"
	}
	if err := oneOf("condition", "INVALID_CONDITION", in.Condition, "event", "children"); err != nil {
		return err
	}
	if in.Metric == "" {
		in.Metric = "count"
	}
	if err := oneOf("metric", "INVALID_METRIC", in.Metric, "count", "high_value", "sum_value"); err != nil {
		return err
	}
	var typeID *string
	target := "1"
	var vmin, vmax *string
	if in.Condition == "event" {
		if typeID, err = refExists(ctx, tx, "economy_event_type", "eventTypeId", "EVENT_TYPE_NOT_FOUND", in.EventTypeID, true); err != nil {
			return err
		}
		t, err := optionalNumeric("target", in.Target)
		if err != nil {
			return err
		}
		if t == nil || numericFloat(*t) <= 0 {
			return badRequest("INVALID_TARGET", "target", "the target must be greater than zero")
		}
		target = *t
		if vmin, err = optionalNumeric("valueMin", in.ValueMin); err != nil {
			return err
		}
		if vmax, err = optionalNumeric("valueMax", in.ValueMax); err != nil {
			return err
		}
		if vmin != nil && vmax != nil && numericFloat(*vmin) > numericFloat(*vmax) {
			return badRequest("INVALID_RANGE", "valueMax", "the largest value must not be smaller than the smallest")
		}
	} else {
		in.Metric = "count"
	}
	if in.Repeat == "" {
		in.Repeat = "once"
	}
	if err := oneOf("repeat", "INVALID_REPEAT", in.Repeat, "once", "daily", "weekly", "interval"); err != nil {
		return err
	}
	if in.Repeat == "interval" {
		if in.RepeatMinutes == nil || *in.RepeatMinutes < 1 || *in.RepeatMinutes > 525600 {
			return badRequest("INVALID_REPEAT_MINUTES", "repeatMinutes", "between 1 minute and one year")
		}
	} else {
		in.RepeatMinutes = nil
	}
	if in.MaxCompletions != nil && *in.MaxCompletions < 1 {
		return badRequest("INVALID_MAX_COMPLETIONS", "maxCompletions", "at least 1, or empty for no limit")
	}
	if in.Repeat == "once" {
		in.MaxCompletions = nil
	}
	parent, err := refExists(ctx, tx, "economy_task", "parentId", "TASK_NOT_FOUND", in.ParentID, false)
	if err != nil {
		return err
	}
	var hasKids bool
	if !create {
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_task WHERE parent_id=$1)`, id).Scan(&hasKids); err != nil {
			return err
		}
	}
	if hasKids && in.Condition != "children" {
		return badRequest("HAS_SUBTASKS", "condition", "this mission has sub-tasks, so it is finished by its sub-tasks")
	}
	group, err := refExists(ctx, tx, "economy_task_group", "groupId", "TASK_GROUP_NOT_FOUND", in.GroupID, false)
	if err != nil {
		return err
	}
	audience, err := checkUserTags(ctx, tx, in.AudienceTagIDs)
	if err != nil {
		return err
	}
	lv, err := cleanLevelTarget(ctx, tx, in.LevelTableID, in.MinLevel, in.MaxLevel)
	if err != nil {
		return err
	}
	starts, ends, err := cleanWindow(in.StartsAt, in.EndsAt)
	if err != nil {
		return err
	}
	if parent != nil {
		if *parent == id {
			return badRequest("SELF_PARENT", "parentId", "a mission cannot be its own sub-task")
		}
		var pCond string
		var pParent sql.NullString
		if err := tx.QueryRowContext(ctx, `SELECT condition, parent_id::text FROM economy_task WHERE id=$1`, *parent).Scan(&pCond, &pParent); err != nil {
			return err
		}
		if pParent.Valid || hasKids {
			return badRequest("NESTED_SUBTASK", "parentId", "sub-tasks go one level deep")
		}
		if pCond != "children" {
			return badRequest("PARENT_NOT_GROUPING", "parentId", "the parent must be finished by its sub-tasks")
		}
		if in.Condition != "event" {
			return badRequest("SUBTASK_NEEDS_EVENT", "condition", "a sub-task is finished by an event")
		}
		if group != nil {
			return badRequest("SUBTASK_IN_GROUP", "groupId", "sub-tasks follow their parent; put the parent in the group")
		}
		in.Repeat, in.RepeatMinutes, in.MaxCompletions = "once", nil, nil
		audience = []string{}
		lv = levelTarget{}
		starts, ends = nil, nil
	}
	if group != nil && in.Repeat != "once" {
		return badRequest("GROUP_TASK_REPEAT", "repeat", "missions in a group run once per round of the group")
	}
	deps, err := cleanDependencies(ctx, tx, id, in.DependsOn)
	if err != nil {
		return err
	}
	prize, err := refExists(ctx, tx, "economy_prize", "prizeId", "PRIZE_NOT_FOUND", in.PrizeID, false)
	if err != nil {
		return err
	}
	if in.GrantMode == "" {
		in.GrantMode = "claim"
	}
	if err := oneOf("grantMode", "INVALID_GRANT_MODE", in.GrantMode, "auto", "claim"); err != nil {
		return err
	}
	return execSave(ctx, tx, "economy_task", id, create, []string{"key", "name", "description", "icon_url", "category_id", "tag_ids", "condition", "event_type_id", "metric", "target", "value_min", "value_max",
		"repeat", "repeat_minutes", "max_completions", "parent_id", "depends_on", "group_id", "prize_id", "grant_mode", "notify", "audience_tag_ids", "level_table_id", "min_level", "max_level",
		"starts_at", "ends_at", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), icon, category, tags, in.Condition, typeID, in.Metric, target, vmin, vmax,
			in.Repeat, in.RepeatMinutes, in.MaxCompletions, parent, deps, group, prize, in.GrantMode, in.Notify, audience, lv.table, lv.min, lv.max,
			starts, ends, in.IsActive, in.Sort})
}

func cleanDependencies(ctx context.Context, tx *sql.Tx, id string, raw []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, r := range raw {
		ref, err := refExists(ctx, tx, "economy_task", "dependsOn", "TASK_NOT_FOUND", r, true)
		if err != nil {
			return nil, err
		}
		if *ref == id {
			return nil, badRequest("SELF_DEPENDENCY", "dependsOn", "a mission cannot wait for itself")
		}
		if seen[*ref] {
			continue
		}
		seen[*ref] = true
		out = append(out, *ref)
	}
	if len(out) > 10 {
		return nil, badRequest("TOO_MANY_DEPENDENCIES", "dependsOn", "at most 10")
	}
	if len(out) == 0 {
		return out, nil
	}
	graph := map[string][]string{}
	rows, err := tx.QueryContext(ctx, `SELECT id::text, depends_on::text FROM economy_task WHERE cardinality(depends_on) > 0`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tid, deps string
		if err := rows.Scan(&tid, &deps); err != nil {
			rows.Close()
			return nil, err
		}
		graph[tid] = parseTextArray(deps)
	}
	rows.Close()
	graph[id] = out
	visiting := map[string]int{}
	var walk func(n string) bool
	walk = func(n string) bool {
		switch visiting[n] {
		case 1:
			return true
		case 2:
			return false
		}
		visiting[n] = 1
		for _, m := range graph[n] {
			if walk(m) {
				return true
			}
		}
		visiting[n] = 2
		return false
	}
	if walk(id) {
		return nil, badRequest("DEPENDENCY_CYCLE", "dependsOn", "these missions would wait for each other forever")
	}
	return out, nil
}

func saveTaskGroup(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key            string   `json:"key"`
		Name           i18n     `json:"name"`
		Description    i18n     `json:"description"`
		IconURL        string   `json:"iconUrl"`
		CategoryID     string   `json:"categoryId"`
		ActiveCount    int      `json:"activeCount"`
		OrderMode      string   `json:"orderMode"`
		Handout        string   `json:"handout"`
		HandoutMinutes int      `json:"handoutMinutes"`
		IsRepeatable   bool     `json:"isRepeatable"`
		AudienceTagIDs []string `json:"audienceTagIds"`
		LevelTableID   string   `json:"levelTableId"`
		MinLevel       *int     `json:"minLevel"`
		MaxLevel       *int     `json:"maxLevel"`
		StartsAt       string   `json:"startsAt"`
		EndsAt         string   `json:"endsAt"`
		IsActive       bool     `json:"isActive"`
		Sort           int      `json:"sort"`
		TaskIDs        []string `json:"taskIds"`
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
	if err := keepKey(ctx, tx, "economy_task_group", id, create, key); err != nil {
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
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	category, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, category, "task"); err != nil {
		return err
	}
	if in.ActiveCount == 0 {
		in.ActiveCount = 1
	}
	if in.ActiveCount < 1 || in.ActiveCount > 20 {
		return badRequest("INVALID_ACTIVE_COUNT", "activeCount", "between 1 and 20")
	}
	if in.OrderMode == "" {
		in.OrderMode = "ordered"
	}
	if err := oneOf("orderMode", "INVALID_ORDER_MODE", in.OrderMode, "ordered", "random"); err != nil {
		return err
	}
	if in.Handout == "" {
		in.Handout = "auto"
	}
	if err := oneOf("handout", "INVALID_HANDOUT", in.Handout, "auto", "request", "timer"); err != nil {
		return err
	}
	if in.Handout == "timer" {
		if in.HandoutMinutes < 1 || in.HandoutMinutes > 525600 {
			return badRequest("INVALID_HANDOUT_MINUTES", "handoutMinutes", "between 1 minute and one year")
		}
	} else {
		in.HandoutMinutes = 0
	}
	audience, err := checkUserTags(ctx, tx, in.AudienceTagIDs)
	if err != nil {
		return err
	}
	lv, err := cleanLevelTarget(ctx, tx, in.LevelTableID, in.MinLevel, in.MaxLevel)
	if err != nil {
		return err
	}
	starts, ends, err := cleanWindow(in.StartsAt, in.EndsAt)
	if err != nil {
		return err
	}
	if err := execSave(ctx, tx, "economy_task_group", id, create, []string{"key", "name", "description", "icon_url", "category_id", "active_count", "order_mode", "handout", "handout_minutes",
		"is_repeatable", "audience_tag_ids", "level_table_id", "min_level", "max_level", "starts_at", "ends_at", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), icon, category, in.ActiveCount, in.OrderMode, in.Handout, in.HandoutMinutes,
			in.IsRepeatable, audience, lv.table, lv.min, lv.max, starts, ends, in.IsActive, in.Sort}); err != nil {
		return err
	}
	if in.TaskIDs == nil {
		return nil
	}
	members := []string{}
	for i, r := range in.TaskIDs {
		field := "taskIds[" + itoa(i) + "]"
		ref, err := refExists(ctx, tx, "economy_task", field, "TASK_NOT_FOUND", r, true)
		if err != nil {
			return err
		}
		if containsString(members, *ref) {
			continue
		}
		var repeat string
		var parent sql.NullString
		if err := tx.QueryRowContext(ctx, `SELECT repeat, parent_id::text FROM economy_task WHERE id=$1`, *ref).Scan(&repeat, &parent); err != nil {
			return err
		}
		if parent.Valid {
			return badRequest("SUBTASK_IN_GROUP", field, "sub-tasks follow their parent; put the parent in the group")
		}
		if repeat != "once" {
			return badRequest("GROUP_TASK_REPEAT", field, "missions in a group run once per round of the group")
		}
		members = append(members, *ref)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE economy_task SET group_id=NULL, updated_at=now() WHERE group_id=$1 AND NOT (id::text = ANY($2))`, id, members); err != nil {
		return err
	}
	for i, m := range members {
		if _, err := tx.ExecContext(ctx, `UPDATE economy_task SET group_id=$1, sort=$3, updated_at=now() WHERE id=$2`, id, m, i+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) taskConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/users/{userId}/tasks", http.MethodGet, s.consoleUserTasks},
		{consolePrefix + "/users/{userId}/tasks/{taskId}/reset", http.MethodPost, s.consoleTaskReset},
		{consolePrefix + "/users/{userId}/tasks/{taskId}/complete", http.MethodPost, s.consoleTaskComplete},
		{consolePrefix + "/users/{userId}/tasks/{taskId}/claim", http.MethodPost, s.consoleTaskClaim},
		{consolePrefix + "/users/{userId}/task-groups/{groupId}/next", http.MethodPost, s.consoleTaskGroupNext},
	}
}

func (s *service) consoleUserTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if _, err := s.refreshGroups(ctx, userID, nil, ""); err != nil {
		s.writeError(w, err)
		return
	}
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	lang := langs.Choose(r.URL.Query().Get("lang"))
	v, err := s.newTaskViewer(ctx, userID, langs, lang, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	top := []*taskDef{}
	for _, t := range v.e.tasks {
		if t != nil && t.ParentID == "" && t.GroupID == "" {
			top = append(top, t)
		}
	}
	sortTasks(top)
	items := []map[string]any{}
	for _, t := range top {
		tv, err := v.view(ctx, t)
		if err != nil {
			s.writeError(w, err)
			return
		}
		items = append(items, tv)
	}
	groups, err := loadGroups(ctx, s.db, ``)
	if err != nil {
		s.writeError(w, err)
		return
	}
	gviews := []map[string]any{}
	for _, g := range groups {
		gv, _, err := v.groupView(ctx, g)
		if err != nil {
			s.writeError(w, err)
			return
		}
		gv["isActive"] = g.Active
		gviews = append(gviews, gv)
	}
	writeJSON(w, http.StatusOK, map[string]any{"lang": lang, "items": items, "groups": gviews})
}

func (s *service) consoleTaskTarget(r *http.Request) (string, *taskDef, error) {
	userID, err := s.consoleUser(r)
	if err != nil {
		return "", nil, err
	}
	id := pathVar(r, "taskId")
	if _, err := uuid.FromString(id); err != nil {
		return "", nil, notFound("TASK_NOT_FOUND")
	}
	t, err := s.findTask(r.Context(), id)
	if err != nil {
		return "", nil, err
	}
	return userID, t, nil
}

func (s *service) consoleTaskReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, t, err := s.consoleTaskTarget(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var scope, status, claimID string
		err := tx.QueryRowContext(ctx, `SELECT scope, status, COALESCE(claim_id::text,'') FROM economy_user_task WHERE task_id=$1 AND user_id=$2 AND status<>'claimed' ORDER BY created_at DESC LIMIT 1 FOR UPDATE`, t.ID, userID).Scan(&scope, &status, &claimID)
		if errors.Is(err, sql.ErrNoRows) {
			var claimed bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_task WHERE task_id=$1 AND user_id=$2)`, t.ID, userID).Scan(&claimed); err != nil {
				return err
			}
			if claimed {
				return conflict("TASK_ALREADY_CLAIMED", "the prize of this mission was already received")
			}
			return notFound("TASK_NOT_STARTED")
		}
		if err != nil {
			return err
		}
		if claimID != "" {
			if _, err := tx.ExecContext(ctx, `DELETE FROM economy_reward_claim WHERE id=$1 AND status<>'granted'`, claimID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM economy_user_task WHERE task_id=$1 AND user_id=$2 AND scope=$3`, t.ID, userID, scope); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `DELETE FROM economy_user_task ut USING economy_task c WHERE c.id=ut.task_id AND c.parent_id=$1 AND ut.user_id=$2 AND ut.scope=$3 AND ut.status<>'claimed'`, t.ID, userID, scope)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) consoleTaskComplete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, t, err := s.consoleTaskTarget(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := newOutcome()
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		e := s.newTaskEnv(ctx, tx, userID, s.now(), settings.clock(), nil)
		sl, err := e.slot(ctx, t)
		if err != nil {
			return err
		}
		if !sl.open() {
			return fail(http.StatusConflict, "TASK_NOT_AVAILABLE", sl.Reason, "this mission is not open for the player now ("+strings.TrimSpace(sl.State+" "+sl.Reason)+")")
		}
		status, err := e.lockRow(ctx, t, sl)
		if err != nil {
			return err
		}
		if status != "active" {
			return conflict("TASK_ALREADY_COMPLETED", "this mission is already finished")
		}
		return e.complete(ctx, t, sl.Scope, "admin", t.Target, out)
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.settleTasks(ctx, userID, out)
	s.settleOutcome(ctx, userID, out, 1)
	writeJSON(w, http.StatusOK, map[string]any{"tasks": out.Tasks, "claims": out.Claims})
}

func (s *service) consoleTaskClaim(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, t, err := s.consoleTaskTarget(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	res, err := s.claimTask(ctx, userID, t.ID, "", "admin")
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"claimId": res.ClaimID, "rewards": res.Rewards})
}

func (s *service) consoleTaskGroupNext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	g, err := s.findGroup(ctx, pathVar(r, "groupId"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	res, err := s.refreshGroups(ctx, userID, []string{g.ID}, g.ID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if len(res) > 0 && res[0].Error != nil {
		s.writeError(w, res[0].Error)
		return
	}
	handed := []string{}
	if len(res) > 0 && res[0].Handed != nil {
		handed = res[0].Handed
	}
	writeJSON(w, http.StatusOK, map[string]any{"handed": handed})
}
