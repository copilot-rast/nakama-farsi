package main

import (
	"strings"
	"testing"
	"time"
)

// TestGroupReadModelsPostgres checks the main app's group screens: ranked
// list with sort and search, the member's own group, member versus profile
// detail views and the creator-only history (source group-read.service.ts).
func TestGroupReadModelsPostgres(t *testing.T) {
	f := newJazebeFixture(t, "lead", "mate", "rival", "visitor")
	u := f.users
	now := time.Now().UTC()
	lottery := f.ok(f.owner, "jazebe_lottery_draft_create", map[string]any{"planetId": f.planetID, "title": "قرعه هفته",
		"startsAt": now.Add(-time.Hour).Format(time.RFC3339), "drawAt": now.Add(48 * time.Hour).Format(time.RFC3339), "ticketPricePoints": 1,
		"prizes": []map[string]any{{"title": "جایزه", "type": "cash", "valueRial": "1000", "priority": 1}}})["id"].(string)
	f.ok(f.owner, "jazebe_lottery_activate", map[string]any{"planetId": f.planetID, "lotteryId": lottery})

	star := f.ok(u["lead"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "گروه ستاره", "desc": "بهترین‌ها", "mode": "open", "cap": 5})["id"].(string)
	f.ok(u["mate"], "jazebe_group_join", map[string]any{"planetId": f.planetID, "groupId": star})
	moon := f.ok(u["rival"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "گروه ماه", "mode": "approval", "cap": 3})["id"].(string)

	// Tickets move the smaller group to rank 1.
	start, end := missionWindow()
	mission := f.ok(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "بلیت", "detectionMethod": "platform_auto",
		"rewardPoints": 1, "rewardTickets": 3, "targetLotteryId": lottery, "startsAt": start, "endsAt": end,
		"code": map[string]any{"codeMode": "unique", "maxUses": 1}})["id"].(string)
	f.ok(f.owner, "jazebe_code_mission_activate", map[string]any{"planetId": f.planetID, "missionId": mission})
	code := f.ok(f.owner, "jazebe_code_mission_generate", map[string]any{"planetId": f.planetID, "missionId": mission, "count": 1})["codes"].([]any)[0]
	if m, ok := code.(map[string]any); ok {
		code = m["code"]
	}
	f.ok(u["rival"], "jazebe_code_mission_redeem", map[string]any{"planetId": f.planetID, "missionId": mission, "code": code})

	names := func(items []any) []string {
		var out []string
		for _, it := range items {
			out = append(out, it.(map[string]any)["name"].(string))
		}
		return out
	}
	byRank := f.ok(u["visitor"], "jazebe_group_list", map[string]any{"planetId": f.planetID})["items"].([]any)
	if got := strings.Join(names(byRank), ","); got != "گروه ماه,گروه ستاره" {
		t.Errorf("rank order %s", got)
	}
	first := byRank[0].(map[string]any)
	expectJSON(t, "ranked card", first, map[string]any{"rank": 1, "members": 1, "cap": 3, "mode": "approval", "color": "#2F6EBC"})
	byMembers := f.ok(u["visitor"], "jazebe_group_list", map[string]any{"planetId": f.planetID, "sort": "members"})["items"].([]any)
	if got := strings.Join(names(byMembers), ","); got != "گروه ستاره,گروه ماه" {
		t.Errorf("members order %s", got)
	}
	expectJSON(t, "card with description", byMembers[0].(map[string]any), map[string]any{"sub": "بهترین‌ها", "rank": 2})
	if _, ok := byMembers[0].(map[string]any)["desc"]; ok {
		t.Error("list cards must not carry desc")
	}
	searched := f.ok(u["visitor"], "jazebe_group_list", map[string]any{"planetId": f.planetID, "search": "ستاره"})["items"].([]any)
	if len(searched) != 1 || searched[0].(map[string]any)["id"] != star {
		t.Errorf("search returned %v", names(searched))
	}
	if n := len(f.ok(u["visitor"], "jazebe_group_list", map[string]any{"planetId": f.planetID, "search": "%"})["items"].([]any)); n != 0 {
		t.Errorf("LIKE wildcard was not escaped: %d results", n)
	}
	f.fails(u["visitor"], "jazebe_group_list", map[string]any{"planetId": f.planetID, "sort": "random"}, 3)

	// Mine: role, pinned raffle, unread count.
	f.ok(u["mate"], "jazebe_group_chat_send", map[string]any{"planetId": f.planetID, "groupId": star, "text": "سلام"})
	mine := f.ok(u["lead"], "jazebe_group_mine", map[string]any{"planetId": f.planetID})
	expectJSON(t, "mine", mine, map[string]any{"viewerRole": "creator"})
	expectJSON(t, "mine group", mine["group"].(map[string]any), map[string]any{"id": star, "desc": "بهترین‌ها", "members": 2})
	expectJSON(t, "mine stats", mine["stats"].(map[string]any), map[string]any{"unreadMessages": 1, "tickets": 0})
	expectJSON(t, "pinned raffle", mine["pinnedRaffle"].(map[string]any), map[string]any{"id": lottery, "title": "قرعه هفته"})
	if mine["leadershipOffer"] != nil {
		t.Errorf("unexpected leadership offer %v", mine["leadershipOffer"])
	}
	if out := f.ok(u["visitor"], "jazebe_group_mine", map[string]any{"planetId": f.planetID}); out["group"] != nil {
		t.Errorf("visitor without group got %v", out)
	}
	rivalMine := f.ok(u["rival"], "jazebe_group_mine", map[string]any{"planetId": f.planetID})
	expectJSON(t, "rival stats", rivalMine["stats"].(map[string]any), map[string]any{"tickets": 3})

	// Detail: member home versus public profile.
	member := f.ok(u["mate"], "jazebe_group_detail", map[string]any{"planetId": f.planetID, "groupId": star})
	expectJSON(t, "member detail", member, map[string]any{"view": "member", "viewerRole": "member", "tickets": 0})
	profile := f.ok(u["visitor"], "jazebe_group_detail", map[string]any{"planetId": f.planetID, "groupId": moon})
	expectJSON(t, "profile detail", profile, map[string]any{"view": "profile", "tickets": 3, "requested": false})
	if _, ok := profile["viewerRole"]; ok {
		t.Error("profile view leaked member fields")
	}
	f.ok(u["visitor"], "jazebe_group_join", map[string]any{"planetId": f.planetID, "groupId": moon})
	requested := f.ok(u["visitor"], "jazebe_group_detail", map[string]any{"planetId": f.planetID, "groupId": moon})
	expectJSON(t, "requested", requested, map[string]any{"requested": true})
	f.fails(u["visitor"], "jazebe_group_detail", map[string]any{"planetId": f.planetID, "groupId": f.owner}, 5)

	// History: creator only, mapped to the source actions and Persian lines.
	f.ok(u["lead"], "jazebe_group_set_role", map[string]any{"planetId": f.planetID, "groupId": star, "memberId": u["mate"], "role": "admin"})
	f.ok(u["lead"], "jazebe_group_update", map[string]any{"planetId": f.planetID, "groupId": star, "desc": "توضیح تازه"})
	f.fails(u["mate"], "jazebe_group_history", map[string]any{"planetId": f.planetID, "groupId": star}, 7)
	history := f.ok(u["lead"], "jazebe_group_history", map[string]any{"planetId": f.planetID, "groupId": star})["items"].([]any)
	if len(history) != 2 {
		t.Fatalf("history rows %d: %v", len(history), history)
	}
	settings, promote := history[0].(map[string]any), history[1].(map[string]any)
	expectJSON(t, "settings history", settings, map[string]any{"action": "edit_settings", "kind": "approve", "text": "کاربر lead تنظیمات گروه را تغییر داد (۱ مورد)"})
	expectJSON(t, "promote history", promote, map[string]any{"action": "promote", "kind": "promote", "text": "کاربر lead «کاربر mate» را به مدیر ارتقا داد"})
	expectJSON(t, "promote target", promote["target"].(map[string]any), map[string]any{"id": u["mate"], "name": "کاربر mate"})
	rivalHistory := f.ok(u["rival"], "jazebe_group_history", map[string]any{"planetId": f.planetID, "groupId": moon})["items"].([]any)
	if len(rivalHistory) != 0 {
		t.Errorf("history leaked across groups: %v", rivalHistory)
	}
}
