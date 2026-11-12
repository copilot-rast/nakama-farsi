package main

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func newTaskHarness(t *testing.T) *harness {
	h := newEventHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_user_task_group, economy_user_task, economy_task, economy_task_group CASCADE`); err != nil {
		t.Fatal(err)
	}
	for name, fn := range h.s.taskRPCs() {
		h.rpcs[name] = fn
	}
	return h
}

func (h *harness) task(body map[string]any) string {
	h.t.Helper()
	base := map[string]any{"name": map[string]string{"fa": "مأموریت " + body["key"].(string), "en": body["key"].(string)}, "isActive": true, "grantMode": "claim"}
	for k, v := range body {
		base[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/tasks", base)["id"].(string)
}

func (h *harness) tasks(user string) map[string]map[string]any {
	h.t.Helper()
	out := map[string]map[string]any{}
	res := h.mustRPC(user, "economy_tasks_list", map[string]any{"lang": "en"})
	var walk func(list []any)
	walk = func(list []any) {
		for _, raw := range list {
			it := raw.(map[string]any)
			out[it["key"].(string)] = it
			if kids, ok := it["children"].([]any); ok {
				walk(kids)
			}
		}
	}
	walk(res["items"].([]any))
	for _, g := range res["groups"].([]any) {
		walk(g.(map[string]any)["tasks"].([]any))
	}
	return out
}

func (h *harness) adminEvent(user, key string, value float64) {
	h.t.Helper()
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/events", map[string]any{"key": key, "value": value})
}

func TestTaskValidation(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	prize := h.coinPrize("coins_10", 10, true)
	bad := func(body map[string]any, code string) {
		t.Helper()
		base := map[string]any{"key": "bad", "name": map[string]string{"fa": "بد"}, "eventTypeId": round, "target": 1}
		for k, v := range body {
			base[k] = v
		}
		if c, out := h.api(http.MethodPost, "/tasks", base); c != 400 || out["error"] != code {
			t.Fatalf("%v = %d %v, want %s", body, c, out, code)
		}
	}
	bad(map[string]any{"target": 0}, "INVALID_TARGET")
	bad(map[string]any{"metric": "avg"}, "INVALID_METRIC")
	bad(map[string]any{"repeat": "interval"}, "INVALID_REPEAT_MINUTES")
	bad(map[string]any{"valueMin": 10, "valueMax": 5}, "INVALID_RANGE")
	bad(map[string]any{"eventTypeId": ""}, "REQUIRED")
	bad(map[string]any{"minLevel": 2}, "LEVEL_TABLE_REQUIRED")
	wrongCat := h.taxonomy("base_category", "prize", "chests", map[string]string{"fa": "صندوق"})
	bad(map[string]any{"categoryId": wrongCat["id"]}, "CATEGORY_KIND_MISMATCH")
	group := h.mustAPI(201, http.MethodPost, "/task-groups", map[string]any{"key": "g1", "name": map[string]string{"fa": "گروه"}, "isActive": true})["id"].(string)
	bad(map[string]any{"groupId": group, "repeat": "daily"}, "GROUP_TASK_REPEAT")
	if c, out := h.api(http.MethodPost, "/task-groups", map[string]any{"key": "g2", "name": map[string]string{"fa": "گروه"}, "handout": "timer"}); c != 400 || out["error"] != "INVALID_HANDOUT_MINUTES" {
		t.Fatalf("timer without minutes = %d %v", c, out)
	}
	leaf := h.task(map[string]any{"key": "leaf", "eventTypeId": round, "target": 1, "prizeId": prize})
	bad(map[string]any{"parentId": leaf}, "PARENT_NOT_GROUPING")
	parent := h.task(map[string]any{"key": "parent", "condition": "children", "prizeId": prize})
	child := h.task(map[string]any{"key": "child", "eventTypeId": round, "target": 1, "parentId": parent, "repeat": "daily", "minLevel": nil})
	got := h.mustAPI(200, http.MethodGet, "/tasks/"+child, nil)
	if got["repeat"] != "once" || got["parentId"] != parent {
		t.Fatalf("a sub-task follows its parent: %v", got)
	}
	bad(map[string]any{"parentId": child}, "NESTED_SUBTASK")
	if c, out := h.api(http.MethodPut, "/tasks/"+parent, map[string]any{"key": "parent", "name": map[string]string{"fa": "x"}, "condition": "event", "eventTypeId": round, "target": 1}); c != 400 || out["error"] != "HAS_SUBTASKS" {
		t.Fatalf("parent to event = %d %v", c, out)
	}
	a := h.task(map[string]any{"key": "a", "eventTypeId": round, "target": 1})
	b := h.task(map[string]any{"key": "b", "eventTypeId": round, "target": 1, "dependsOn": []string{a}})
	if c, out := h.api(http.MethodPut, "/tasks/"+a, map[string]any{"key": "a", "name": map[string]string{"fa": "x"}, "eventTypeId": round, "target": 1, "dependsOn": []string{b}}); c != 400 || out["error"] != "DEPENDENCY_CYCLE" {
		t.Fatalf("cycle = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPut, "/tasks/"+a, map[string]any{"key": "a2", "name": map[string]string{"fa": "x"}, "eventTypeId": round, "target": 1}); c != 400 || out["error"] != "KEY_IMMUTABLE" {
		t.Fatalf("key change = %d %v", c, out)
	}
	if c, out := h.api(http.MethodDelete, "/tasks/"+parent, nil); c != 409 || out["error"] != "HAS_SUBTASKS" {
		t.Fatalf("delete parent = %d %v", c, out)
	}
	h.mustAPI(200, http.MethodDelete, "/tasks/"+a, nil)
	if deps := h.mustAPI(200, http.MethodGet, "/tasks/"+b, nil)["dependsOn"].([]any); len(deps) != 0 {
		t.Fatalf("deleting a task removes it from dependencies: %v", deps)
	}
	list := h.mustAPI(200, http.MethodGet, "/tasks?level=sub", nil)
	if list["total"].(float64) != 1 {
		t.Fatalf("sub-task filter = %v", list)
	}
}

func TestTaskMetricsFilterClaimAndExploits(t *testing.T) {
	h := newTaskHarness(t)
	h.eventType("round_finished", false, map[string]any{"valueMin": 0, "valueMax": 400})
	star := h.eventType("golden_star", true, map[string]any{"valueMin": 0, "valueMax": 1, "rateLimit": 5})
	round := h.typeIDByKey("round_finished")
	p50 := h.coinPrize("coins_50", 50, true)
	h.task(map[string]any{"key": "big_rounds", "eventTypeId": round, "metric": "count", "target": 2, "valueMin": 100, "prizeId": p50})
	h.task(map[string]any{"key": "best_round", "eventTypeId": round, "metric": "high_value", "target": 300, "prizeId": p50, "grantMode": "auto"})
	h.task(map[string]any{"key": "points", "eventTypeId": round, "metric": "sum_value", "target": 500, "prizeId": p50})
	h.task(map[string]any{"key": "stars", "eventTypeId": star, "target": 3, "prizeId": p50})
	user := h.user()
	h.adminEvent(user, "round_finished", 50)
	h.adminEvent(user, "round_finished", 150)
	v := h.tasks(user)
	if v["big_rounds"]["value"].(float64) != 1 || v["points"]["value"].(float64) != 200 || v["best_round"]["value"].(float64) != 150 {
		t.Fatalf("a round of 50 does not count for big rounds; sum and high value move: %v %v %v", v["big_rounds"], v["points"], v["best_round"])
	}
	if code := h.rpcCode(user, "economy_task_claim", map[string]any{"key": "big_rounds"}); code != "TASK_NOT_COMPLETED" {
		t.Fatalf("claim early = %s", code)
	}
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "round_finished", "value": 400}); code != "EVENT_NOT_ALLOWED" {
		t.Fatalf("a client cannot send a server-only event = %s", code)
	}
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "golden_star", "value": 50}); code != "VALUE_OUT_OF_RANGE" {
		t.Fatalf("value range = %s", code)
	}
	for i := 0; i < 5; i++ {
		h.mustRPC(user, "economy_event_record", map[string]any{"key": "golden_star", "value": 1})
	}
	if code := h.rpcCode(user, "economy_event_record", map[string]any{"key": "golden_star", "value": 1}); code != "TOO_MANY_EVENTS" {
		t.Fatalf("rate limit = %s", code)
	}
	v = h.tasks(user)
	if v["stars"]["state"] != "completed" || v["stars"]["value"].(float64) != 3 || v["stars"]["claimable"] != true {
		t.Fatalf("stars stop at the target and wait for a claim: %v", v["stars"])
	}
	if n := h.count(`SELECT count FROM economy_user_task ut JOIN economy_task t ON t.id=ut.task_id WHERE t.key='stars' AND ut.user_id=$1`, user); n != 3 {
		t.Fatalf("no counting after completion: %d", n)
	}
	h.adminEvent(user, "round_finished", 320)
	v = h.tasks(user)
	if v["big_rounds"]["state"] != "completed" || v["best_round"]["state"] != "claimed" || v["points"]["state"] != "completed" {
		t.Fatalf("three tasks done, the automatic one paid: %v %v %v", v["big_rounds"]["state"], v["best_round"]["state"], v["points"]["state"])
	}
	if h.wallet(user)["coin"] != 50 {
		t.Fatalf("auto prize = %v", h.wallet(user))
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND kind='task' AND status<>'granted'`, user); n != 0 {
		t.Fatalf("claim-mode tasks do not appear as waiting rewards until claimed: %d", n)
	}
	res := h.mustRPC(user, "economy_task_claim", map[string]any{"key": "big_rounds", "lang": "fa"})
	if res["task"].(map[string]any)["state"] != "claimed" || res["prize"].(map[string]any)["key"] != "coins_50" || h.wallet(user)["coin"] != 100 {
		t.Fatalf("claim = %v %v", res, h.wallet(user))
	}
	if code := h.rpcCode(user, "economy_task_claim", map[string]any{"key": "big_rounds"}); code != "ALREADY_CLAIMED" {
		t.Fatalf("second claim = %s", code)
	}
	claimID := v["points"]["claimId"].(string)
	if code := h.rpcCode(h.user(), "economy_reward_claim", map[string]any{"claimId": claimID}); code != "CLAIM_NOT_FOUND" {
		t.Fatalf("someone else's claim = %s", code)
	}
	h.mustRPC(user, "economy_reward_claim", map[string]any{"claimId": claimID})
	if h.wallet(user)["coin"] != 150 {
		t.Fatalf("the generic claim RPC pays a task claim too: %v", h.wallet(user))
	}
	if n := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'sourceType'='task'`, user); n != 3 {
		t.Fatalf("one ledger row per task prize: %d", n)
	}
	mine := h.mustRPC(user, "economy_rewards_mine", map[string]any{"lang": "en"})["items"].([]any)
	if len(mine) != 3 || mine[0].(map[string]any)["kind"] != "task" || mine[0].(map[string]any)["sourceName"] == "" {
		t.Fatalf("rewards list shows the task name: %v", mine)
	}
}

func TestTaskDailyWeeklyWindows(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	prize := h.coinPrize("coins_10", 10, true)
	h.task(map[string]any{"key": "daily2", "eventTypeId": round, "target": 2, "repeat": "daily", "maxCompletions": 2, "prizeId": prize, "grantMode": "auto"})
	h.task(map[string]any{"key": "weekly3", "eventTypeId": round, "target": 3, "repeat": "weekly", "prizeId": prize, "grantMode": "auto"})
	user := h.user()
	h.at(time.Date(2026, 11, 12, 20, 0, 0, 0, time.UTC))
	h.adminEvent(user, "round_finished", 1)
	h.at(time.Date(2026, 11, 12, 20, 40, 0, 0, time.UTC))
	h.adminEvent(user, "round_finished", 1)
	v := h.tasks(user)
	if v["daily2"]["value"].(float64) != 1 || v["daily2"]["state"] != "active" {
		t.Fatalf("20:40 UTC is already the next day in Tehran, so the daily task starts again: %v", v["daily2"])
	}
	if v["weekly3"]["value"].(float64) != 2 {
		t.Fatalf("the week did not change (Thursday to Friday, weeks start on Saturday): %v", v["weekly3"])
	}
	h.adminEvent(user, "round_finished", 1)
	if h.wallet(user)["coin"] != 20 {
		t.Fatalf("daily done on day two and weekly done on its third round = %v", h.wallet(user))
	}
	if v := h.tasks(user); v["daily2"]["state"] != "claimed" || v["daily2"]["periodEndsAt"] == nil {
		t.Fatalf("daily claimed with an end time: %v", v["daily2"])
	}
	h.at(time.Date(2026, 11, 13, 21, 0, 0, 0, time.UTC))
	h.adminEvent(user, "round_finished", 1)
	h.adminEvent(user, "round_finished", 1)
	v = h.tasks(user)
	if v["daily2"]["state"] != "claimed" || v["daily2"]["completions"].(float64) != 2 || h.wallet(user)["coin"] != 30 {
		t.Fatalf("Saturday in Tehran: the second completion of daily2 pays: %v wallet %v", v["daily2"], h.wallet(user))
	}
	if v["weekly3"]["value"].(float64) != 2 || v["weekly3"]["state"] != "active" {
		t.Fatalf("Saturday starts a new week: %v", v["weekly3"])
	}
	h.at(time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC))
	h.adminEvent(user, "round_finished", 1)
	h.adminEvent(user, "round_finished", 1)
	v = h.tasks(user)
	if v["daily2"]["state"] != "done" || v["daily2"]["value"].(float64) != 0 || h.wallet(user)["coin"] != 40 {
		t.Fatalf("Sunday: daily2 used its two completions; weekly3 pays on its third round: %v wallet %v", v["daily2"], h.wallet(user))
	}
}

func TestTaskIntervalRepeat(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	prize := h.coinPrize("coins_10", 10, true)
	h.task(map[string]any{"key": "hourly", "eventTypeId": round, "target": 1, "repeat": "interval", "repeatMinutes": 60, "maxCompletions": 2, "prizeId": prize})
	user := h.user()
	t0 := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	h.at(t0)
	h.adminEvent(user, "round_finished", 1)
	h.mustRPC(user, "economy_task_claim", map[string]any{"key": "hourly"})
	h.at(t0.Add(30 * time.Minute))
	h.adminEvent(user, "round_finished", 1)
	v := h.tasks(user)["hourly"]
	if v["state"] != "cooldown" || v["nextAt"] != t0.Add(time.Hour).Format(time.RFC3339) {
		t.Fatalf("cooldown for an hour: %v", v)
	}
	h.at(t0.Add(61 * time.Minute))
	if v := h.tasks(user)["hourly"]; v["state"] != "active" || v["value"].(float64) != 0 {
		t.Fatalf("a new round of the task after an hour: %v", v)
	}
	h.adminEvent(user, "round_finished", 1)
	h.mustRPC(user, "economy_task_claim", map[string]any{"key": "hourly"})
	h.at(t0.Add(5 * time.Hour))
	if v := h.tasks(user)["hourly"]; v["state"] != "done" || v["completions"].(float64) != 2 || h.wallet(user)["coin"] != 20 {
		t.Fatalf("two completions and done: %v %v", v, h.wallet(user))
	}
}

func TestTaskDependenciesAndSubtasks(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	star := h.eventType("golden_star", false, nil)
	prize := h.coinPrize("coins_10", 10, true)
	big := h.coinPrize("coins_100", 100, true)
	first := h.task(map[string]any{"key": "first", "eventTypeId": round, "target": 1})
	h.task(map[string]any{"key": "second", "eventTypeId": round, "target": 1, "dependsOn": []string{first}, "prizeId": prize})
	parent := h.task(map[string]any{"key": "combo", "condition": "children", "prizeId": big, "grantMode": "auto"})
	h.task(map[string]any{"key": "combo_round", "eventTypeId": round, "target": 3, "parentId": parent})
	h.task(map[string]any{"key": "combo_star", "eventTypeId": star, "target": 2, "parentId": parent})
	user := h.user()
	v := h.tasks(user)
	if v["second"]["state"] != "locked" || v["second"]["dependsOn"].([]any)[0].(map[string]any)["done"] != false {
		t.Fatalf("second waits for first: %v", v["second"])
	}
	h.adminEvent(user, "round_finished", 1)
	v = h.tasks(user)
	if v["first"]["state"] != "claimed" || v["second"]["state"] != "active" || v["second"]["value"].(float64) != 0 {
		t.Fatalf("the event that finished first does not count for second: first %v second %v", v["first"], v["second"])
	}
	h.adminEvent(user, "round_finished", 1)
	h.adminEvent(user, "round_finished", 1)
	v = h.tasks(user)
	if v["combo"]["value"].(float64) != 1 || v["combo"]["target"].(float64) != 2 || v["combo_round"]["state"] != "claimed" {
		t.Fatalf("one of two sub-tasks done: %v", v["combo"])
	}
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.s.recordServerEvent(h.ctx, user, "golden_star", 1, nil)
		}()
	}
	wg.Wait()
	v = h.tasks(user)
	if v["combo"]["state"] != "claimed" || h.wallet(user)["coin"] != 100 {
		t.Fatalf("both sub-tasks done: the parent pays once: %v %v", v["combo"], h.wallet(user))
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND owner_id=$2`, user, parent); n != 1 {
		t.Fatalf("one parent claim: %d", n)
	}
}

func TestTaskTargetingByTagAndLevel(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	prize := h.coinPrize("coins_10", 10, true)
	xp := h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "xp", "name": map[string]string{"fa": "تجربه"}})["id"].(string)
	h.mustAPI(201, http.MethodPost, "/meta-rules", map[string]any{"key": "xp_round", "name": map[string]string{"fa": "تجربه"}, "eventTypeId": round, "metaId": xp, "action": "add", "valueSource": "event", "amount": 1, "isActive": true})
	table := h.mustAPI(201, http.MethodPost, "/level-tables", map[string]any{"key": "lv", "name": map[string]string{"fa": "سطح"}, "metaId": xp, "isActive": true,
		"steps": []map[string]any{{"level": 1, "threshold": 0}, {"level": 2, "threshold": 100}, {"level": 3, "threshold": 300}}})["id"].(string)
	vip := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})["id"].(string)
	h.task(map[string]any{"key": "vip_only", "eventTypeId": round, "target": 1, "audienceTagIds": []string{vip}, "prizeId": prize})
	h.task(map[string]any{"key": "level2", "eventTypeId": round, "target": 1, "levelTableId": table, "minLevel": 2, "maxLevel": 2, "prizeId": prize})
	user := h.user()
	h.adminEvent(user, "round_finished", 50)
	v := h.tasks(user)
	if _, ok := v["vip_only"]; ok {
		t.Fatalf("hidden without the tag: %v", v["vip_only"])
	}
	if _, ok := v["level2"]; ok {
		t.Fatalf("hidden below level 2")
	}
	if n := h.count(`SELECT count(*) FROM economy_user_task WHERE user_id=$1`, user); n != 0 {
		t.Fatalf("no progress for tasks the player cannot see: %d", n)
	}
	if _, err := h.s.db.Exec(`INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, user, vip); err != nil {
		t.Fatal(err)
	}
	h.adminEvent(user, "round_finished", 60)
	v = h.tasks(user)
	if v["vip_only"]["state"] != "completed" || v["level2"]["state"] != "active" {
		t.Fatalf("tag added; level 2 reached by 110 xp after this event: %v %v", v["vip_only"], v["level2"])
	}
	h.adminEvent(user, "round_finished", 10)
	if v := h.tasks(user); v["level2"]["state"] != "completed" {
		t.Fatalf("level 2 task progresses: %v", v["level2"])
	}
	h.adminEvent(user, "round_finished", 300)
	if _, ok := h.tasks(user)["level2"]; ok {
		t.Fatalf("above the highest level the task is hidden again")
	}
}

func TestTaskGroupsHandout(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	prize := h.coinPrize("coins_10", 10, true)
	group := h.mustAPI(201, http.MethodPost, "/task-groups", map[string]any{"key": "chain", "name": map[string]string{"fa": "زنجیره", "en": "Chain"}, "activeCount": 1, "orderMode": "ordered", "handout": "auto", "isRepeatable": true, "isActive": true})["id"].(string)
	ids := []string{}
	for _, k := range []string{"c1", "c2", "c3"} {
		ids = append(ids, h.task(map[string]any{"key": k, "eventTypeId": round, "target": 1, "prizeId": prize}))
	}
	got := h.mustAPI(200, http.MethodPut, "/task-groups/"+group, map[string]any{"key": "chain", "name": map[string]string{"fa": "زنجیره", "en": "Chain"}, "activeCount": 1, "orderMode": "ordered", "handout": "auto", "isRepeatable": true, "isActive": true,
		"taskIds": []string{ids[2], ids[0], ids[1]}})
	if order := got["taskIds"].([]any); order[0] != ids[2] || order[2] != ids[1] {
		t.Fatalf("group order follows the editor: %v", order)
	}
	user := h.user()
	cur := func() map[string]any {
		return h.mustRPC(user, "economy_task_group_current", map[string]any{"key": "chain", "lang": "en"})["group"].(map[string]any)
	}
	g := cur()
	if tasks := g["tasks"].([]any); len(tasks) != 1 || tasks[0].(map[string]any)["key"] != "c3" || g["name"] != "Chain" {
		t.Fatalf("first task by group order: %v", g)
	}
	if _, ok := h.tasks(user)["c1"]; ok {
		t.Fatalf("c1 is not handed out yet")
	}
	h.adminEvent(user, "round_finished", 1)
	if code := h.rpcCode(user, "economy_task_group_next", map[string]any{"key": "chain"}); code != "HANDOUT_AUTOMATIC" {
		t.Fatalf("next on an automatic group = %s", code)
	}
	h.mustRPC(user, "economy_task_claim", map[string]any{"key": "c3"})
	g = cur()
	if tasks := g["tasks"].([]any); len(tasks) != 2 || tasks[1].(map[string]any)["key"] != "c1" || g["done"].(float64) != 1 {
		t.Fatalf("claim hands out the next one: %v", g)
	}
	for _, k := range []string{"c1", "c2"} {
		h.adminEvent(user, "round_finished", 1)
		h.mustRPC(user, "economy_task_claim", map[string]any{"key": k})
	}
	g = cur()
	if g["cycle"].(float64) != 2 || len(g["tasks"].([]any)) != 1 || g["tasks"].([]any)[0].(map[string]any)["state"] != "active" {
		t.Fatalf("a repeatable group starts round two: %v", g)
	}
	if h.wallet(user)["coin"] != 30 {
		t.Fatalf("three prizes = %v", h.wallet(user))
	}

	h.mustAPI(201, http.MethodPost, "/task-groups", map[string]any{"key": "ask", "name": map[string]string{"fa": "درخواستی"}, "activeCount": 1, "handout": "request", "isActive": true})
	askID := h.mustAPI(200, http.MethodGet, "/task-groups?q=ask", nil)["items"].([]any)[0].(map[string]any)["id"].(string)
	h.task(map[string]any{"key": "r1", "eventTypeId": round, "target": 1, "groupId": askID})
	h.task(map[string]any{"key": "r2", "eventTypeId": round, "target": 1, "groupId": askID})
	ask := h.mustRPC(user, "economy_task_group_current", map[string]any{"key": "ask"})["group"].(map[string]any)
	if len(ask["tasks"].([]any)) != 0 || ask["canRequest"] != true {
		t.Fatalf("a request group waits for the player: %v", ask)
	}
	next := h.mustRPC(user, "economy_task_group_next", map[string]any{"key": "ask"})
	if keys := next["handed"].([]any); len(keys) != 1 || keys[0] != "r1" {
		t.Fatalf("next = %v", next)
	}
	if code := h.rpcCode(user, "economy_task_group_next", map[string]any{"key": "ask"}); code != "CURRENT_TASK_MUST_BE_DONE" {
		t.Fatalf("second request = %s", code)
	}
	h.adminEvent(user, "round_finished", 1)
	if g := h.mustRPC(user, "economy_task_group_current", map[string]any{"key": "ask"})["group"].(map[string]any); g["canRequest"] != true {
		t.Fatalf("a task without a prize frees its place at once: %v", g)
	}
	h.mustRPC(user, "economy_task_group_next", map[string]any{"key": "ask"})
	h.adminEvent(user, "round_finished", 1)
	if code := h.rpcCode(user, "economy_task_group_next", map[string]any{"key": "ask"}); code != "TASK_GROUP_FINISHED" {
		t.Fatalf("a one-time group ends = %s", code)
	}

	h.mustAPI(201, http.MethodPost, "/task-groups", map[string]any{"key": "slow", "name": map[string]string{"fa": "آهسته"}, "activeCount": 1, "handout": "timer", "handoutMinutes": 30, "isActive": true})
	slowID := h.mustAPI(200, http.MethodGet, "/task-groups?q=slow", nil)["items"].([]any)[0].(map[string]any)["id"].(string)
	h.task(map[string]any{"key": "s1", "eventTypeId": round, "target": 1, "groupId": slowID})
	h.task(map[string]any{"key": "s2", "eventTypeId": round, "target": 1, "groupId": slowID})
	t0 := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	h.at(t0)
	if g := h.mustRPC(user, "economy_task_group_current", map[string]any{"key": "slow"})["group"].(map[string]any); len(g["tasks"].([]any)) != 1 {
		t.Fatalf("the first timer task comes at once: %v", g)
	}
	h.adminEvent(user, "round_finished", 1)
	g = h.mustRPC(user, "economy_task_group_current", map[string]any{"key": "slow"})["group"].(map[string]any)
	if g["nextAt"] != t0.Add(30*time.Minute).Format(time.RFC3339) || g["canRequest"] != false {
		t.Fatalf("the next timer task waits 30 minutes: %v", g)
	}
	if code := h.rpcCode(user, "economy_task_group_next", map[string]any{"key": "slow"}); code != "DURATION_NOT_ELAPSED" {
		t.Fatalf("early = %s", code)
	}
	h.at(t0.Add(31 * time.Minute))
	if g := h.mustRPC(user, "economy_task_group_current", map[string]any{"key": "slow"})["group"].(map[string]any); len(g["tasks"].([]any)) != 2 {
		t.Fatalf("after 30 minutes the next task is there: %v", g)
	}
}

func TestTaskClaimRollbackAndConcurrency(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	prize := h.coinPrize("coins_25", 25, false)
	h.task(map[string]any{"key": "once", "eventTypeId": round, "target": 1, "prizeId": prize})
	user := h.user()
	h.adminEvent(user, "round_finished", 1)
	h.nk.mu.Lock()
	h.nk.failWallet = true
	h.nk.mu.Unlock()
	if _, err := h.rpc(user, "economy_task_claim", map[string]any{"key": "once"}); err == nil {
		t.Fatal("claim must fail while the wallet is down")
	}
	v := h.tasks(user)["once"]
	if v["state"] != "completed" || v["claimable"] != true || h.wallet(user)["coin"] != 0 {
		t.Fatalf("a failed grant rolls the claim back: %v %v", v, h.wallet(user))
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1`, user); n != 0 {
		t.Fatalf("no claim row is kept after the rollback: %d", n)
	}
	h.nk.mu.Lock()
	h.nk.failWallet = false
	h.nk.mu.Unlock()
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, already := 0, 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.rpc(user, "economy_task_claim", map[string]any{"key": "once"})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				ok++
			} else if app, isApp := asApp(err); isApp && app.Code == "ALREADY_CLAIMED" {
				already++
			} else {
				t.Errorf("claim error %v", err)
			}
		}()
	}
	wg.Wait()
	if ok != 1 || already != 9 || h.wallet(user)["coin"] != 25 {
		t.Fatalf("ten parallel claims pay once: ok %d already %d wallet %v", ok, already, h.wallet(user))
	}
	if n := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1`, user); n != 1 {
		t.Fatalf("one ledger row: %d", n)
	}

	auto := h.coinPrize("coins_5", 5, true)
	h.task(map[string]any{"key": "ten", "eventTypeId": round, "target": 10, "prizeId": auto, "grantMode": "auto"})
	fresh := h.user()
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.s.recordServerEvent(h.ctx, fresh, "round_finished", 1, nil)
		}()
	}
	wg.Wait()
	if n := h.count(`SELECT count FROM economy_user_task ut JOIN economy_task t ON t.id=ut.task_id WHERE t.key='ten' AND ut.user_id=$1`, fresh); n != 10 {
		t.Fatalf("parallel events stop counting at the target: %d", n)
	}
	if h.wallet(fresh)["coin"] != 5 || h.count(`SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND kind='task'`, fresh) != 1 {
		t.Fatalf("25 parallel events pay the automatic task once (the claim-mode one waits): %v", h.wallet(fresh))
	}
}

func TestTaskConsolePlayerTools(t *testing.T) {
	h := newTaskHarness(t)
	round := h.eventType("round_finished", false, nil)
	prize := h.coinPrize("coins_10", 10, true)
	id := h.task(map[string]any{"key": "five", "eventTypeId": round, "target": 5, "prizeId": prize, "notify": true})
	user := h.user()
	h.adminEvent(user, "round_finished", 1)
	h.adminEvent(user, "round_finished", 1)
	list := h.mustAPI(200, http.MethodGet, "/users/"+user+"/tasks?lang=en", nil)
	item := list["items"].([]any)[0].(map[string]any)
	if item["value"].(float64) != 2 || item["status"] != "active" {
		t.Fatalf("admin view = %v", item)
	}
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/tasks/"+id+"/reset", nil)
	if v := h.tasks(user)["five"]; v["value"].(float64) != 0 {
		t.Fatalf("reset = %v", v)
	}
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/tasks/"+id+"/complete", nil)
	if c, out := h.api(http.MethodPost, "/users/"+user+"/tasks/"+id+"/complete", nil); c != 409 || out["error"] != "TASK_ALREADY_COMPLETED" {
		t.Fatalf("complete twice = %d %v", c, out)
	}
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/tasks/"+id+"/claim", nil)
	if h.wallet(user)["coin"] != 10 {
		t.Fatalf("support claim = %v", h.wallet(user))
	}
	if c, out := h.api(http.MethodPost, "/users/"+user+"/tasks/"+id+"/reset", nil); c != 409 || out["error"] != "TASK_ALREADY_CLAIMED" {
		t.Fatalf("reset after claim = %d %v", c, out)
	}
	claims := h.mustAPI(200, http.MethodGet, "/event-claims?userId="+user+"&kind=task", nil)
	if claims["total"].(float64) != 1 || claims["items"].([]any)[0].(map[string]any)["sourceKey"] != "five" {
		t.Fatalf("claims list = %v", claims)
	}
	if c, out := h.api(http.MethodDelete, "/tasks/"+id, nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete used task = %d %v", c, out)
	}
	sum := h.mustAPI(200, http.MethodGet, "/summary", nil)
	if sum["counts"].(map[string]any)["tasks"].(float64) != 1 {
		t.Fatalf("summary = %v", sum)
	}
}
