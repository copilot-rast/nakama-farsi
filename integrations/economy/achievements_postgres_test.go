package main

import (
	"net/http"
	"sync"
	"testing"
)

func newAchievementHarness(t *testing.T) *harness {
	h := newTaskHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_user_achievement, economy_achievement_tier, economy_achievement CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_event_type (id, key, name, is_system, value_min) VALUES (gen_random_uuid(), 'achievement_unlocked', '{"fa":"دستاورد"}', true, 0)`); err != nil {
		t.Fatal(err)
	}
	for name, fn := range h.s.achievementRPCs() {
		h.rpcs[name] = fn
	}
	return h
}

func (h *harness) achievement(body map[string]any) map[string]any {
	h.t.Helper()
	base := map[string]any{"name": map[string]string{"fa": "دستاورد " + body["key"].(string), "en": body["key"].(string)}, "isActive": true}
	for k, v := range body {
		base[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/achievements", base)
}

func (h *harness) achievements(user string) (map[string]any, map[string]map[string]any) {
	h.t.Helper()
	res := h.mustRPC(user, "economy_achievements", map[string]any{"lang": "en"})
	out := map[string]map[string]any{}
	for _, raw := range res["items"].([]any) {
		it := raw.(map[string]any)
		if k, ok := it["key"].(string); ok {
			out[k] = it
		} else {
			out["hidden:"+it["id"].(string)] = it
		}
	}
	return res, out
}

func tiersOf(rows ...[2]any) []map[string]any {
	out := []map[string]any{}
	for i, r := range rows {
		t := map[string]any{"tier": i + 1, "threshold": r[0], "grantMode": "claim"}
		if r[1] != nil {
			t["prizeId"] = r[1]
		}
		out = append(out, t)
	}
	return out
}

func TestAchievementValidation(t *testing.T) {
	h := newAchievementHarness(t)
	round := h.eventType("round_finished", false, nil)
	task := h.task(map[string]any{"key": "t1", "eventTypeId": round, "target": 1})
	bad := func(body map[string]any, code string) {
		t.Helper()
		base := map[string]any{"key": "bad", "name": map[string]string{"fa": "بد"}, "source": "event", "eventTypeId": round, "tiers": tiersOf([2]any{1, nil})}
		for k, v := range body {
			base[k] = v
		}
		if c, out := h.api(http.MethodPost, "/achievements", base); c != 400 || out["error"] != code {
			t.Fatalf("%v = %d %v, want %s", body, c, out, code)
		}
	}
	bad(map[string]any{"tiers": []any{}}, "TIERS_REQUIRED")
	bad(map[string]any{"tiers": tiersOf([2]any{10, nil}, [2]any{5, nil})}, "THRESHOLDS_NOT_ORDERED")
	bad(map[string]any{"tiers": tiersOf([2]any{0, nil})}, "INVALID_THRESHOLD")
	bad(map[string]any{"rarity": "mythic"}, "INVALID_RARITY")
	bad(map[string]any{"points": -1}, "INVALID_POINTS")
	bad(map[string]any{"source": "magic"}, "INVALID_SOURCE")
	bad(map[string]any{"metric": "avg"}, "INVALID_METRIC")
	bad(map[string]any{"valueMin": 10, "valueMax": 1}, "INVALID_RANGE")
	bad(map[string]any{"source": "task"}, "TASK_REQUIRED")
	bad(map[string]any{"source": "task", "taskId": task, "tiers": tiersOf([2]any{1, nil}, [2]any{2, nil})}, "TIERS_NEED_EVENT")
	wrong := h.taxonomy("base_category", "task", "c1", map[string]string{"fa": "ت"})
	bad(map[string]any{"categoryId": wrong["id"]}, "CATEGORY_KIND_MISMATCH")
	many := []map[string]any{}
	for i := 1; i <= 11; i++ {
		many = append(many, map[string]any{"tier": i, "threshold": i})
	}
	bad(map[string]any{"tiers": many}, "TOO_MANY_TIERS")
	manual := h.achievement(map[string]any{"key": "manual_one", "source": "manual"})
	if tiers := manual["tiers"].([]any); len(tiers) != 1 || tiers[0].(map[string]any)["tier"].(float64) != 1 {
		t.Fatalf("a manual achievement gets one tier: %v", manual)
	}
	if c, out := h.api(http.MethodPut, "/achievements/"+manual["id"].(string), map[string]any{"key": "manual_two", "name": map[string]string{"fa": "x"}}); c != 400 || out["error"] != "KEY_IMMUTABLE" {
		t.Fatalf("key change = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "pa", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "achievement", "key": "nope", "action": "give"}}}); c != 400 || out["error"] != "ACHIEVEMENT_NOT_FOUND" {
		t.Fatalf("prize unknown achievement = %d %v", c, out)
	}
	h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "pa", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "achievement", "key": "manual_one", "action": "give"}}})
	if c, out := h.api(http.MethodDelete, "/achievements/"+manual["id"].(string), nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete used in a prize = %d %v", c, out)
	}
	h.achievement(map[string]any{"key": "linked", "source": "task", "taskId": task})
	if c, out := h.api(http.MethodDelete, "/tasks/"+task, nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete a task an achievement waits for = %d %v", c, out)
	}
	list := h.mustAPI(200, http.MethodGet, "/achievements?source=manual", nil)
	if list["total"].(float64) != 1 {
		t.Fatalf("source filter = %v", list)
	}
}

func TestAchievementTiersProgressAndClaims(t *testing.T) {
	h := newAchievementHarness(t)
	round := h.eventType("round_finished", false, map[string]any{"valueMin": 0, "valueMax": 400})
	p10 := h.coinPrize("coins_10", 10, true)
	p20 := h.coinPrize("coins_20", 20, true)
	p50 := h.coinPrize("coins_50", 50, true)
	tiers := tiersOf([2]any{2, p10}, [2]any{4, p20}, [2]any{6, p50})
	tiers[1]["grantMode"] = "auto"
	h.achievement(map[string]any{"key": "rounds", "source": "event", "eventTypeId": round, "points": 10, "tiers": tiers})
	h.achievement(map[string]any{"key": "big_round", "source": "event", "eventTypeId": round, "metric": "high_value", "points": 5, "tiers": tiersOf([2]any{300, nil})})
	h.achievement(map[string]any{"key": "only_big", "source": "event", "eventTypeId": round, "valueMin": 100, "tiers": tiersOf([2]any{2, nil})})
	h.eventType("unlocks", false, nil)
	h.achievement(map[string]any{"key": "collector", "source": "event", "eventTypeId": h.typeIDByKey("achievement_unlocked"), "metric": "count", "tiers": tiersOf([2]any{2, nil})})
	user := h.user()
	h.adminEvent(user, "round_finished", 50)
	res, v := h.achievements(user)
	if v["rounds"]["state"] != "locked" || v["rounds"]["value"].(float64) != 1 || v["rounds"]["progress"].(float64) != 0.5 || res["unlocked"].(float64) != 0 {
		t.Fatalf("one round: %v", v["rounds"])
	}
	if code := h.rpcCode(user, "economy_achievement_claim", map[string]any{"key": "rounds"}); code != "ACHIEVEMENT_LOCKED" {
		t.Fatalf("claim locked = %s", code)
	}
	h.adminEvent(user, "round_finished", 120)
	res, v = h.achievements(user)
	if v["rounds"]["tier"].(float64) != 1 || v["rounds"]["state"] != "unlocked" || v["only_big"]["value"].(float64) != 1 {
		t.Fatalf("tier 1 after two rounds, the 50 round does not count for only_big: %v %v", v["rounds"], v["only_big"])
	}
	if c := v["rounds"]["claimable"].([]any); len(c) != 1 || c[0].(float64) != 1 || res["ready"].(float64) != 1 {
		t.Fatalf("tier 1 waits for a claim: %v", v["rounds"])
	}
	if code := h.rpcCode(user, "economy_achievement_claim", map[string]any{"key": "rounds", "tier": 2}); code != "TIER_NOT_REACHED" {
		t.Fatalf("claim a later tier = %s", code)
	}
	var wg sync.WaitGroup
	results := make(chan string, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.rpc(user, "economy_achievement_claim", map[string]any{"key": "rounds", "tier": 1})
			if err == nil {
				results <- "ok"
				return
			}
			if app, ok := asApp(err); ok {
				results <- app.Code
			} else {
				results <- err.Error()
			}
		}()
	}
	wg.Wait()
	close(results)
	oks := 0
	for r := range results {
		if r == "ok" {
			oks++
		} else if r != "ALREADY_CLAIMED" {
			t.Fatalf("parallel claim error %s", r)
		}
	}
	if oks != 1 || h.wallet(user)["coin"] != 10 {
		t.Fatalf("parallel claims pay once: %d %v", oks, h.wallet(user))
	}
	h.adminEvent(user, "round_finished", 10)
	h.adminEvent(user, "round_finished", 310)
	_, v = h.achievements(user)
	if v["rounds"]["tier"].(float64) != 2 || h.wallet(user)["coin"] != 30 {
		t.Fatalf("tier 2 is paid automatically: %v %v", v["rounds"], h.wallet(user))
	}
	if v["big_round"]["state"] != "completed" || v["collector"]["tier"].(float64) != 1 {
		t.Fatalf("a 310 round unlocks big_round; two unlocks (rounds, big_round) open collector: %v %v", v["big_round"], v["collector"])
	}
	if n := h.count(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='achievement_unlocked' AND e.user_id=$1`, user); n != 4 {
		t.Fatalf("achievement_unlocked recorded once per unlock (rounds, only_big, big_round, collector): %d", n)
	}
	h.adminEvent(user, "round_finished", 10)
	h.adminEvent(user, "round_finished", 10)
	h.adminEvent(user, "round_finished", 10)
	res, v = h.achievements(user)
	if v["rounds"]["tier"].(float64) != 3 || v["rounds"]["state"] != "completed" || v["rounds"]["next"] != nil || v["rounds"]["progress"].(float64) != 1 {
		t.Fatalf("max tier: %v", v["rounds"])
	}
	out := h.mustRPC(user, "economy_achievement_claim", map[string]any{"key": "rounds", "lang": "fa"})
	if cl := out["claims"].([]any); len(cl) != 1 || cl[0].(map[string]any)["tier"].(float64) != 3 || h.wallet(user)["coin"] != 80 {
		t.Fatalf("claim without a tier collects the waiting one: %v %v", out, h.wallet(user))
	}
	if code := h.rpcCode(user, "economy_achievement_claim", map[string]any{"key": "rounds"}); code != "ALREADY_CLAIMED" {
		t.Fatalf("nothing left = %s", code)
	}
	if n := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'sourceType'='achievement'`, user); n != 3 {
		t.Fatalf("one ledger row per tier: %d", n)
	}
	if res["points"].(float64) != 15 {
		t.Fatalf("points = %v", res["points"])
	}
	mine := h.mustRPC(user, "economy_rewards_mine", map[string]any{"lang": "en"})["items"].([]any)
	if len(mine) == 0 || mine[0].(map[string]any)["kind"] != "achievement" || mine[0].(map[string]any)["source"] != "rounds" {
		t.Fatalf("rewards list names the achievement: %v", mine)
	}
}

func TestAchievementParallelEvents(t *testing.T) {
	h := newAchievementHarness(t)
	round := h.eventType("round_finished", false, nil)
	p5 := h.coinPrize("coins_5", 5, true)
	tiers := tiersOf([2]any{5, p5}, [2]any{10, p5})
	tiers[0]["grantMode"] = "auto"
	tiers[1]["grantMode"] = "auto"
	h.achievement(map[string]any{"key": "grinder", "source": "event", "eventTypeId": round, "tiers": tiers})
	user := h.user()
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := 1.0
			if _, err := h.s.recordEvent(h.ctx, eventInput{UserID: user, Key: "round_finished", Value: &v, Source: "admin"}); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	_, v := h.achievements(user)
	if v["grinder"]["value"].(float64) != 12 || v["grinder"]["tier"].(float64) != 2 {
		t.Fatalf("every event counted once: %v", v["grinder"])
	}
	if h.wallet(user)["coin"] != 10 {
		t.Fatalf("two automatic tiers paid once each: %v", h.wallet(user))
	}
	if n := h.count(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='achievement_unlocked' AND e.user_id=$1`, user); n != 1 {
		t.Fatalf("one unlock event: %d", n)
	}
}

func TestAchievementUnlockByTaskGroupPrizeAndAdmin(t *testing.T) {
	h := newAchievementHarness(t)
	round := h.eventType("round_finished", false, nil)
	coins := h.coinPrize("coins_30", 30, true)
	solo := h.task(map[string]any{"key": "solo", "eventTypeId": round, "target": 2})
	group := h.mustAPI(201, http.MethodPost, "/task-groups", map[string]any{"key": "chain", "name": map[string]string{"fa": "زنجیره"}, "isActive": true, "activeCount": 1})["id"].(string)
	g1 := h.task(map[string]any{"key": "g1", "eventTypeId": round, "target": 1})
	g2 := h.task(map[string]any{"key": "g2", "eventTypeId": round, "target": 1})
	h.mustAPI(200, http.MethodPut, "/task-groups/"+group, map[string]any{"key": "chain", "name": map[string]string{"fa": "زنجیره"}, "isActive": true, "activeCount": 2, "taskIds": []string{g1, g2}})
	tier := tiersOf([2]any{1, coins})
	tier[0]["grantMode"] = "auto"
	h.achievement(map[string]any{"key": "by_task", "source": "task", "taskId": solo, "tiers": tier})
	h.achievement(map[string]any{"key": "by_group", "source": "task", "taskGroupId": group, "points": 7})
	secret := h.achievement(map[string]any{"key": "secret", "source": "manual", "isHidden": true, "hint": map[string]string{"fa": "ماه را ببین", "en": "Look at the moon"}})
	h.achievement(map[string]any{"key": "gifted", "source": "manual", "notify": true})
	user := h.user()
	h.mustRPC(user, "economy_tasks_list", map[string]any{})
	h.adminEvent(user, "round_finished", 1)
	_, v := h.achievements(user)
	if v["by_task"]["unlocked"] != false || v["by_group"]["unlocked"] != true {
		t.Fatalf("one round: the group's two tasks done, solo not yet: %v %v", v["by_task"], v["by_group"])
	}
	h.adminEvent(user, "round_finished", 1)
	_, v = h.achievements(user)
	if v["by_task"]["unlocked"] != true || h.wallet(user)["coin"] != 30 {
		t.Fatalf("finishing the linked task unlocks and pays the automatic prize: %v %v", v["by_task"], h.wallet(user))
	}
	hidden := v["hidden:"+secret["id"].(string)]
	if hidden == nil || hidden["name"] != nil || hidden["hint"] != "Look at the moon" || hidden["state"] != "hidden" {
		t.Fatalf("a locked secret shows only its hint: %v", hidden)
	}
	friend := h.user()
	prof := h.mustRPC(friend, "economy_achievements_profile", map[string]any{"userId": user, "lang": "en"})
	if prof["unlocked"].(float64) != 2 || prof["total"].(float64) != 3 || prof["points"].(float64) != 7 {
		t.Fatalf("profile counts only visible achievements: %v", prof)
	}
	prize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "secret_box", "name": map[string]string{"fa": "جعبه"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "achievement", "key": "secret", "action": "give"}, {"type": "currency", "key": "coin", "action": "add", "value": 5}}})["id"].(string)
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/grant", map[string]any{"prizeId": prize})
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/grant", map[string]any{"prizeId": prize})
	_, v = h.achievements(user)
	if v["secret"] == nil || v["secret"]["unlocked"] != true || v["secret"]["name"] != "secret" {
		t.Fatalf("the prize item unlocks the secret and shows it: %v", v)
	}
	if n := h.count(`SELECT count(*) FROM economy_user_achievement ua JOIN economy_achievement a ON a.id=ua.achievement_id WHERE a.key='secret' AND ua.user_id=$1 AND ua.source='prize'`, user); n != 1 {
		t.Fatalf("one row, source prize: %d", n)
	}
	prof = h.mustRPC(friend, "economy_achievements_profile", map[string]any{"userId": user, "lang": "en"})
	if prof["unlocked"].(float64) != 3 || prof["total"].(float64) != 4 {
		t.Fatalf("an unlocked secret shows on the profile: %v", prof)
	}
	gifted := h.mustAPI(200, http.MethodGet, "/achievements?q=gifted", nil)["items"].([]any)[0].(map[string]any)["id"].(string)
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/achievements", map[string]any{"achievementId": gifted})
	if c, out := h.api(http.MethodPost, "/users/"+user+"/achievements", map[string]any{"achievementId": gifted}); c != 409 || out["error"] != "ALREADY_UNLOCKED" {
		t.Fatalf("give twice = %d %v", c, out)
	}
	card := h.mustAPI(200, http.MethodGet, "/users/"+user+"/achievements?lang=en", nil)
	found := false
	for _, raw := range card["items"].([]any) {
		it := raw.(map[string]any)
		if it["key"] == "gifted" && it["unlockedBy"] == "admin" {
			found = true
		}
	}
	if !found {
		t.Fatalf("console card shows the admin unlock: %v", card)
	}
	h.mustAPI(200, http.MethodDelete, "/users/"+user+"/achievements/"+gifted, nil)
	if c, out := h.api(http.MethodDelete, "/users/"+user+"/achievements/"+gifted, nil); c != 409 || out["error"] != "NOT_UNLOCKED" {
		t.Fatalf("revoke twice = %d %v", c, out)
	}
	if n := h.count(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND kind='achievement'`, user); n < 5 {
		t.Fatalf("activity rows for unlocks and the revoke: %d", n)
	}
}

func TestAchievementRevokeKeepsPaidTiers(t *testing.T) {
	h := newAchievementHarness(t)
	round := h.eventType("round_finished", false, nil)
	p10 := h.coinPrize("coins_10", 10, true)
	a := h.achievement(map[string]any{"key": "two", "source": "event", "eventTypeId": round, "tiers": tiersOf([2]any{1, p10}, [2]any{2, p10})})
	user := h.user()
	h.adminEvent(user, "round_finished", 1)
	h.mustRPC(user, "economy_achievement_claim", map[string]any{"key": "two"})
	h.adminEvent(user, "round_finished", 1)
	h.mustAPI(200, http.MethodDelete, "/users/"+user+"/achievements/"+a["id"].(string), nil)
	_, v := h.achievements(user)
	if v["two"]["unlocked"] != false || v["two"]["value"].(float64) != 0 {
		t.Fatalf("revoke resets progress: %v", v["two"])
	}
	h.adminEvent(user, "round_finished", 1)
	h.adminEvent(user, "round_finished", 1)
	_, v = h.achievements(user)
	st := v["two"]["tiers"].([]any)
	if st[0].(map[string]any)["status"] != "claimed" || st[1].(map[string]any)["status"] != "claimable" {
		t.Fatalf("tier 1 was paid before the revoke and is never paid twice: %v", st)
	}
	h.mustRPC(user, "economy_achievement_claim", map[string]any{"key": "two"})
	if h.wallet(user)["coin"] != 20 {
		t.Fatalf("wallet = %v", h.wallet(user))
	}
	if c, out := h.api(http.MethodPut, "/achievements/"+a["id"].(string), map[string]any{"key": "two", "name": map[string]string{"fa": "x"}, "source": "event", "eventTypeId": round, "tiers": tiersOf([2]any{1, p10})}); c != 400 || out["error"] != "TIER_IN_USE" {
		t.Fatalf("removing a paid tier = %d %v", c, out)
	}
}
