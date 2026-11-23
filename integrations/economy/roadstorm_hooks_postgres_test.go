package main

import (
	"net/http"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

func TestReferralDisplayNameAndRebind(t *testing.T) {
	h, _, _ := newReferralHarness(t)
	oldGoal := h.eventType("round_finished", true, nil)
	h.settings(map[string]any{"goalEventTypeId": oldGoal, "goalMeasure": "count", "goalValue": 1})
	a := h.user()
	code := h.codeOf(a)
	b, c, d := h.user(), h.user(), h.user()
	for i, u := range []string{b, c, d} {
		h.s.db.Exec(`UPDATE users SET display_name=$2 WHERE id=$1`, u, []string{"راننده یک", "Driver Two", ""}[i])
		h.s.db.Exec(`INSERT INTO user_device (id, user_id) VALUES ($1, $2)`, "rb-dev-"+u, u)
		h.mustRPC(u, "economy_referral_enter", map[string]any{"code": code})
	}
	t.Cleanup(func() { h.s.db.Exec(`DELETE FROM user_device WHERE id LIKE 'rb-dev-%'`) })
	me := h.mustRPC(a, "economy_referral_me", map[string]any{})
	names := map[string]bool{}
	for _, raw := range me["recent"].([]any) {
		r := raw.(map[string]any)
		if _, ok := r["displayName"]; !ok {
			t.Fatalf("recent row without displayName: %v", r)
		}
		names[r["displayName"].(string)] = true
	}
	if !names["راننده یک"] || !names["Driver Two"] || !names[""] {
		t.Fatalf("display names %v", names)
	}
	if rb := h.mustRPC(b, "economy_referral_me", map[string]any{})["referredBy"].(map[string]any); rb["displayName"] == nil {
		t.Fatalf("referredBy %v", rb)
	}
	list := h.mustAPI(http.StatusOK, http.MethodGet, "/referrals?q="+"Driver", nil)
	if items := list["items"].([]any); len(items) != 1 || items[0].(map[string]any)["refereeDisplayName"] != "Driver Two" {
		t.Fatalf("search by display name %v", list)
	}

	newGoal := h.eventType("rs_match_finished", true, map[string]any{"valueMin": 1, "valueMax": 1})
	newRef := h.coinPrize("rs_oil_100", 100, true)
	newRee := h.coinPrize("rs_oil_50", 50, true)
	h.settings(map[string]any{"goalEventTypeId": newGoal, "referrerPrizeId": newRef, "refereePrizeId": newRee})
	h.record(b, "rs_match_finished", 1)
	if n := h.count(`SELECT count(*) FROM economy_referral WHERE status='pending'`); n != 3 {
		t.Fatalf("old invites must still wait on the old goal, pending=%d", n)
	}
	pre := h.mustAPI(http.StatusOK, http.MethodGet, "/referral-rebind", nil)
	if pre["total"].(float64) != 3 || pre["willComplete"].(float64) != 1 {
		t.Fatalf("preview %v", pre)
	}
	if code, out := h.api(http.MethodPost, "/referral-rebind", map[string]any{}); code != 400 || out["error"] != "CONFIRM_REQUIRED" {
		t.Fatalf("no confirm %d %v", code, out)
	}
	coinsA := h.wallet(a)["coin"]
	res := h.mustAPI(http.StatusOK, http.MethodPost, "/referral-rebind", map[string]any{"confirm": "rebind"})
	if res["rebound"].(float64) != 3 || res["completed"].(float64) != 1 {
		t.Fatalf("rebind %v", res)
	}
	if h.wallet(a)["coin"] != coinsA+100 || h.wallet(b)["coin"] != 50 {
		t.Fatalf("new prizes not paid: a=%v b=%v", h.wallet(a), h.wallet(b))
	}
	h.record(c, "rs_match_finished", 1)
	if n := h.count(`SELECT count(*) FROM economy_referral WHERE status='completed'`); n != 2 {
		t.Fatalf("rebound invite does not complete on the new goal, completed=%d", n)
	}
	again := h.mustAPI(http.StatusOK, http.MethodGet, "/referral-rebind", nil)
	if again["total"].(float64) != 0 {
		t.Fatalf("second preview %v", again)
	}
	h.s.db.Exec(`UPDATE economy_referral SET goal_event_type_id=NULL WHERE referee_id=$1`, d)
	if res := h.mustAPI(http.StatusOK, http.MethodPost, "/referral-rebind", map[string]any{"confirm": "rebind", "prizes": false}); res["rebound"].(float64) != 1 {
		t.Fatalf("an invite that lost its goal must be rebound: %v", res)
	}
}

func TestTaskHooks(t *testing.T) {
	h := newTaskHarness(t)
	economyhooks.SetTasks(h.s.listTasksHook, h.s.claimTaskHook)
	ev := h.eventType("rs_zombie_kills", false, nil)
	prize := h.coinPrize("rs_pits_100", 100, true)
	cat := h.taxonomy("base_category", "task", "rs_missions", map[string]string{"fa": "مأموریت"})
	other := h.taxonomy("base_category", "task", "other", map[string]string{"fa": "دیگر"})
	h.task(map[string]any{"key": "rs_kill_10", "eventTypeId": ev, "metric": "sum_value", "target": 10, "repeat": "daily", "prizeId": prize, "categoryId": cat["id"]})
	h.task(map[string]any{"key": "other_task", "eventTypeId": ev, "target": 1, "prizeId": prize, "categoryId": other["id"]})
	u := h.user()
	list, err := economyhooks.ListTasks(h.ctx, u, economyhooks.TaskFilter{Lang: "en", Category: "rs_missions"})
	if err != nil || len(list) != 1 || list[0]["key"] != "rs_kill_10" || list[0]["prizeId"] != prize {
		t.Fatalf("list %v %v", list, err)
	}
	if _, err := economyhooks.ClaimTask(h.ctx, u, "rs_kill_10", "en"); err == nil {
		t.Fatal("claim before done must fail")
	} else if r, ok := economyhooks.AsRefusal(err); !ok || r.Code == "" {
		t.Fatalf("claim error is not a refusal: %v", err)
	}
	h.adminEvent(u, "rs_zombie_kills", 6)
	h.adminEvent(u, "rs_zombie_kills", 6)
	list, _ = economyhooks.ListTasks(h.ctx, u, economyhooks.TaskFilter{Lang: "en", Keys: []string{"rs_kill_10"}})
	if len(list) != 1 || list[0]["claimable"] != true {
		t.Fatalf("after progress %v", list)
	}
	before := h.wallet(u)["coin"]
	out, err := economyhooks.ClaimTask(h.ctx, u, "rs_kill_10", "en")
	if err != nil || out["claimId"] == nil || h.wallet(u)["coin"] != before+100 {
		t.Fatalf("claim %v %v wallet %v", out, err, h.wallet(u))
	}
}
