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

func (s *service) taskRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_tasks_list":         s.rpcTasksList,
		"economy_task_claim":         s.rpcTaskClaim,
		"economy_task_group_current": s.rpcTaskGroupCurrent,
		"economy_task_group_next":    s.rpcTaskGroupNext,
	}
}

type taskViewer struct {
	s         *service
	e         *taskEnv
	langs     *localized.Set
	lang      string
	admin     bool
	prizes    map[string]prizeName
	events    map[string][2]string
	cats      map[string]string
	tags      map[string]string
	children  map[string][]*taskDef
	allLoaded bool
}

func (s *service) newTaskViewer(ctx context.Context, userID string, langs *localized.Set, lang string, admin bool) (*taskViewer, error) {
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	e := s.newTaskEnv(ctx, s.db, userID, s.now(), settings.clock(), nil)
	all, err := loadTasks(ctx, s.db, ``)
	if err != nil {
		return nil, err
	}
	e.addTasks(all)
	ids := make([]string, 0, len(all))
	v := &taskViewer{s: s, e: e, langs: langs, lang: lang, admin: admin, events: map[string][2]string{}, cats: map[string]string{}, tags: map[string]string{}, children: map[string][]*taskDef{}, allLoaded: true}
	prizeIDs := []string{}
	for _, t := range all {
		ids = append(ids, t.ID)
		if t.PrizeID != "" {
			prizeIDs = append(prizeIDs, t.PrizeID)
		}
		if t.ParentID != "" {
			v.children[t.ParentID] = append(v.children[t.ParentID], t)
		}
	}
	if err := e.ensureRows(ctx, ids); err != nil {
		return nil, err
	}
	if v.prizes, err = s.prizeNames(ctx, langs, lang, prizeIDs); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text FROM economy_event_type`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, key, name string
		if err := rows.Scan(&id, &key, &name); err != nil {
			rows.Close()
			return nil, err
		}
		v.events[id] = [2]string{key, langs.Resolve(localized.Parse(name), lang)}
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT 'c', id::text, key FROM base_category WHERE key_type='task' UNION ALL SELECT 't', id::text, key FROM base_tag WHERE key_type IN ('task','general')`)
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
			v.cats[id] = key
		} else {
			v.tags[id] = key
		}
	}
	rows.Close()
	return v, nil
}

func (v *taskViewer) text(raw string) string {
	return v.langs.Resolve(localized.Parse(raw), v.lang)
}

func timeOrNil(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func (v *taskViewer) view(ctx context.Context, t *taskDef) (map[string]any, error) {
	e := v.e
	sl, err := e.slot(ctx, t)
	if err != nil {
		return nil, err
	}
	var row *userTaskRow
	if sl.Scope != "" {
		if row, err = e.row(ctx, t.ID, sl.Scope); err != nil {
			return nil, err
		}
	}
	state := "active"
	switch sl.State {
	case "hidden":
		state = "hidden"
	case "locked":
		state = "locked"
	case "cooldown", "done":
		state = sl.State
		if row != nil && row.Status == "completed" {
			state = "completed"
		}
	default:
		if row != nil && row.Status != "active" {
			state = row.Status
		}
	}
	target := t.Target
	value := 0.0
	if row != nil {
		value = taskMeasure(t.Metric, row.Count, row.Sum, row.Max)
	}
	kids := []map[string]any{}
	if t.Condition == "children" {
		total := 0
		done := 0
		for _, c := range v.children[t.ID] {
			if !c.Active {
				continue
			}
			total++
			cv, err := v.view(ctx, c)
			if err != nil {
				return nil, err
			}
			if s, _ := cv["state"].(string); s == "completed" || s == "claimed" {
				done++
			}
			kids = append(kids, cv)
		}
		target = float64(total)
		value = float64(done)
		if row != nil && row.Status != "active" {
			value = target
		}
	}
	if state == "completed" || state == "claimed" {
		if value < target {
			value = target
		}
	}
	progress := 0.0
	if target > 0 {
		progress = value / target
		if progress > 1 {
			progress = 1
		}
	}
	pending := 0
	pendingClaim := ""
	for _, r := range e.rows[t.ID] {
		if r.Status == "completed" && r.Scope != sl.Scope {
			pending++
			pendingClaim = r.ClaimID
		}
	}
	claimID := ""
	if row != nil && row.Status == "completed" {
		claimID = row.ClaimID
	}
	claimable := t.PrizeID != "" && ((state == "completed" && (t.GrantMode == "claim" || (row != nil && row.Error != ""))) || pending > 0)
	if claimID == "" && pending > 0 {
		claimID = pendingClaim
	}
	out := map[string]any{
		"id": t.ID, "key": t.Key, "name": v.text(t.Name), "description": v.text(t.Description), "iconUrl": t.IconURL,
		"categoryKey": v.cats[t.CategoryID], "tags": v.tagKeys(t.TagIDs), "condition": t.Condition, "metric": t.Metric,
		"target": target, "value": value, "progress": progress, "state": state, "claimable": claimable, "pending": pending,
		"grantMode": t.GrantMode, "repeat": t.Repeat, "repeatMinutes": t.RepeatMinutes, "maxCompletions": t.MaxCompletions, "completions": sl.Completions,
		"scope": sl.Scope, "periodEndsAt": timeOrNil(sl.PeriodEnd), "expiresAt": timeOrNil(sl.ExpiresAt), "nextAt": timeOrNil(sl.NextAt),
		"prize": nil, "children": kids, "dependsOn": []map[string]any{},
	}
	if claimID != "" {
		out["claimId"] = claimID
	}
	if t.EventTypeID != "" {
		ev := v.events[t.EventTypeID]
		out["eventKey"], out["eventName"] = ev[0], ev[1]
	}
	if t.ValueMin != nil {
		out["valueMin"] = *t.ValueMin
	}
	if t.ValueMax != nil {
		out["valueMax"] = *t.ValueMax
	}
	if t.PrizeID != "" {
		out["prize"] = prizeJSON(v.prizes[t.PrizeID], t.PrizeID)
	}
	if t.GroupID != "" {
		if g, _ := e.group(ctx, t.GroupID); g != nil {
			out["groupKey"] = g.Key
		}
	}
	if t.ParentID != "" {
		if p := e.tasks[t.ParentID]; p != nil {
			out["parentKey"] = p.Key
		}
	}
	deps := []map[string]any{}
	for _, id := range t.DependsOn {
		d := e.tasks[id]
		if d == nil {
			continue
		}
		done := false
		for _, r := range e.rows[id] {
			if r.Status != "active" {
				done = true
			}
		}
		deps = append(deps, map[string]any{"id": d.ID, "key": d.Key, "name": v.text(d.Name), "done": done})
	}
	out["dependsOn"] = deps
	if v.admin {
		out["reason"] = sl.Reason
		if row != nil {
			out["status"] = row.Status
			out["count"], out["sum"], out["max"] = row.Count, row.Sum, row.Max
			if row.CompletedAt.Valid {
				out["completedAt"] = row.CompletedAt.Time
			}
			if row.ClaimedAt.Valid {
				out["claimedAt"] = row.ClaimedAt.Time
			}
			if row.Error != "" {
				out["error"] = row.Error
			}
		}
		out["isActive"] = t.Active
		out["parentId"] = t.ParentID
		out["groupId"] = t.GroupID
	}
	return out, nil
}

func (v *taskViewer) tagKeys(ids []string) []string {
	out := []string{}
	for _, id := range ids {
		if k, ok := v.tags[id]; ok {
			out = append(out, k)
		}
	}
	return out
}

func (v *taskViewer) groupView(ctx context.Context, g *groupDef) (map[string]any, bool, error) {
	e := v.e
	open, reason, err := e.groupOpen(ctx, g)
	if err != nil {
		return nil, false, err
	}
	if !open && !v.admin {
		return nil, false, nil
	}
	st, err := e.groupState(ctx, g.ID)
	if err != nil {
		return nil, false, err
	}
	members := []*taskDef{}
	for _, t := range e.tasks {
		if t != nil && t.GroupID == g.ID && t.ParentID == "" && t.Active {
			members = append(members, t)
		}
	}
	sortTasks(members)
	scope := "g" + itoa(st.Cycle)
	tasks := []map[string]any{}
	done := 0
	unfinished := 0
	for _, id := range st.Handed {
		t := e.tasks[id]
		if t == nil {
			continue
		}
		tv, err := v.view(ctx, t)
		if err != nil {
			return nil, false, err
		}
		if tv["state"] == "hidden" && !v.admin {
			continue
		}
		claimed := false
		if r, _ := e.row(ctx, id, scope); r != nil && r.Status == "claimed" {
			claimed = true
		}
		if claimed || !t.Active {
			done++
		} else {
			unfinished++
		}
		tasks = append(tasks, tv)
	}
	e.ignoreG = true
	remaining := 0
	for _, t := range members {
		if containsString(st.Handed, t.ID) {
			continue
		}
		sl, err := e.slot(ctx, t)
		if err != nil {
			e.ignoreG = false
			return nil, false, err
		}
		if sl.open() {
			remaining++
		}
	}
	e.ignoreG = false
	free := g.ActiveCount - unfinished
	canRequest := false
	if remaining > 0 && free > 0 {
		switch g.Handout {
		case "request":
			canRequest = true
		case "timer":
			canRequest = st.NextAt != nil && !e.now.Before(*st.NextAt)
		}
	}
	finished := !g.Repeatable && unfinished == 0 && remaining == 0 && len(st.Handed) > 0
	out := map[string]any{
		"id": g.ID, "key": g.Key, "name": v.text(g.Name), "description": v.text(g.Description), "iconUrl": g.IconURL, "categoryKey": v.cats[g.CategoryID],
		"cycle": st.Cycle, "activeCount": g.ActiveCount, "orderMode": g.OrderMode, "handout": g.Handout, "handoutMinutes": g.HandoutMins, "repeatable": g.Repeatable,
		"total": len(members), "done": done, "remaining": remaining, "tasks": tasks, "nextAt": timeOrNil(st.NextAt), "canRequest": canRequest, "finished": finished,
		"expiresAt": timeOrNil(g.EndsAt),
	}
	if v.admin {
		out["open"] = open
		out["reason"] = reason
	}
	return out, true, nil
}

func (s *service) rpcTasksList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang     string `json:"lang"`
		Category string `json:"category"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := s.refreshGroups(ctx, userID, nil, ""); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	v, err := s.newTaskViewer(ctx, userID, langs, lang, false)
	if err != nil {
		return nil, err
	}
	top := []*taskDef{}
	for _, t := range v.e.tasks {
		if t != nil && t.ParentID == "" && t.GroupID == "" && t.Active {
			top = append(top, t)
		}
	}
	sortTasks(top)
	items := []map[string]any{}
	ready := 0
	for _, t := range top {
		if in.Category != "" && v.cats[t.CategoryID] != in.Category {
			continue
		}
		tv, err := v.view(ctx, t)
		if err != nil {
			return nil, err
		}
		if tv["state"] == "hidden" {
			continue
		}
		if tv["claimable"] == true {
			ready++
		}
		items = append(items, tv)
	}
	groups, err := loadGroups(ctx, s.db, `WHERE g.is_active`)
	if err != nil {
		return nil, err
	}
	gviews := []map[string]any{}
	for _, g := range groups {
		if in.Category != "" && v.cats[g.CategoryID] != in.Category {
			continue
		}
		gv, ok, err := v.groupView(ctx, g)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		for _, tv := range gv["tasks"].([]map[string]any) {
			if tv["claimable"] == true {
				ready++
			}
		}
		gviews = append(gviews, gv)
	}
	settings, _ := loadEventSettings(ctx, s.db)
	return map[string]any{"lang": lang, "timezone": settings.Timezone, "items": items, "groups": gviews, "ready": ready}, nil
}

func (s *service) findGroup(ctx context.Context, key string) (*groupDef, error) {
	key = strings.TrimSpace(key)
	var list []*groupDef
	var err error
	if _, uerr := uuid.FromString(key); uerr == nil {
		list, err = loadGroups(ctx, s.db, `WHERE g.id=$1`, key)
	} else {
		list, err = loadGroups(ctx, s.db, `WHERE g.key=$1`, key)
	}
	if err != nil {
		return nil, err
	}
	if len(list) == 0 || !list[0].Active {
		return nil, notFound("TASK_GROUP_NOT_FOUND")
	}
	return list[0], nil
}

func (s *service) groupResponse(ctx context.Context, userID, lang string, g *groupDef, handed []string) (map[string]any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	v, err := s.newTaskViewer(ctx, userID, langs, lang, false)
	if err != nil {
		return nil, err
	}
	gv, ok, err := v.groupView(ctx, g)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, notFound("TASK_GROUP_NOT_FOUND")
	}
	keys := []string{}
	for _, id := range handed {
		if t := v.e.tasks[id]; t != nil {
			keys = append(keys, t.Key)
		}
	}
	return map[string]any{"lang": lang, "group": gv, "handed": keys}, nil
}

func (s *service) rpcTaskGroupCurrent(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key  string `json:"key"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	g, err := s.findGroup(ctx, in.Key)
	if err != nil {
		return nil, err
	}
	res, err := s.refreshGroups(ctx, userID, []string{g.ID}, "")
	if err != nil {
		return nil, err
	}
	var handed []string
	if len(res) > 0 {
		handed = res[0].Handed
	}
	return s.groupResponse(ctx, userID, in.Lang, g, handed)
}

func (s *service) rpcTaskGroupNext(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key  string `json:"key"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	g, err := s.findGroup(ctx, in.Key)
	if err != nil {
		return nil, err
	}
	if g.Handout == "auto" {
		return nil, conflict("HANDOUT_AUTOMATIC", "missions of this group are handed out automatically")
	}
	res, err := s.refreshGroups(ctx, userID, []string{g.ID}, g.ID)
	if err != nil {
		return nil, err
	}
	var handed []string
	if len(res) > 0 {
		if res[0].Error != nil {
			return nil, res[0].Error
		}
		handed = res[0].Handed
	}
	return s.groupResponse(ctx, userID, in.Lang, g, handed)
}

func (s *service) findTask(ctx context.Context, idOrKey string) (*taskDef, error) {
	idOrKey = strings.TrimSpace(idOrKey)
	var list []*taskDef
	var err error
	if _, uerr := uuid.FromString(idOrKey); uerr == nil {
		list, err = loadTasks(ctx, s.db, `WHERE t.id=$1`, idOrKey)
	} else if idOrKey != "" {
		list, err = loadTasks(ctx, s.db, `WHERE t.key=$1`, idOrKey)
	}
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, notFound("TASK_NOT_FOUND")
	}
	return list[0], nil
}

func (s *service) rpcTaskClaim(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		TaskID  string `json:"taskId"`
		Key     string `json:"key"`
		ClaimID string `json:"claimId"`
		Lang    string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	taskID := ""
	claimID := ""
	if in.ClaimID != "" {
		if _, err := uuid.FromString(in.ClaimID); err != nil {
			return nil, notFound("CLAIM_NOT_FOUND")
		}
		claimID = in.ClaimID
	} else {
		ref := in.TaskID
		if ref == "" {
			ref = in.Key
		}
		t, err := s.findTask(ctx, ref)
		if err != nil {
			return nil, err
		}
		taskID = t.ID
	}
	res, err := s.claimTask(ctx, userID, taskID, claimID, "player")
	if err != nil {
		return nil, err
	}
	return s.taskClaimResponse(ctx, userID, in.Lang, res)
}

func (s *service) taskClaimResponse(ctx context.Context, userID, lang string, res *taskClaimResult) (map[string]any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	names, err := s.prizeNames(ctx, langs, lang, []string{res.PrizeID})
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "claimId": res.ClaimID, "kind": "task", "rewards": res.Rewards, "prize": prizeJSON(names[res.PrizeID], res.PrizeID), "levels": []levelChange{}}
	if res.Extra != nil {
		out["levels"] = res.Extra.Levels
	}
	if res.Task != nil {
		v, err := s.newTaskViewer(ctx, userID, langs, lang, false)
		if err != nil {
			return nil, err
		}
		if t := v.e.tasks[res.Task.ID]; t != nil {
			tv, err := v.view(ctx, t)
			if err != nil {
				return nil, err
			}
			out["task"] = tv
		}
	}
	return out, nil
}

func (s *service) taskClaimOwner(ctx context.Context, claimID string) (string, bool, error) {
	var owner string
	err := s.db.QueryRowContext(ctx, `SELECT user_id::text FROM economy_user_task WHERE claim_id=$1`, claimID).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return owner, err == nil, err
}
