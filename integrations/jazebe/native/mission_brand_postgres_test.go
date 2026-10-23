package main

import (
	"testing"
	"time"
)

// TestBrandMissionManagePostgres covers the brand mission builder's meta,
// detail, update rules and bulk review (source mission-brand.service.ts).
func TestBrandMissionManagePostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan", "fan2")
	u := f.users
	start, end := missionWindow()
	ref := func(id string) map[string]any { return map[string]any{"planetId": f.planetID, "missionId": id} }
	solo := f.ok(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "تنها ماموریت امتیازی",
		"detectionMethod": "platform_auto", "rewardPoints": 20, "startsAt": start, "endsAt": end, "code": map[string]any{"codeMode": "unique", "maxUses": 1}})["id"].(string)
	f.ok(f.owner, "jazebe_code_mission_activate", ref(solo))
	content := f.ok(f.owner, "jazebe_content_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "محتوا",
		"rewardPoints": 30, "startsAt": start, "endsAt": end, "content": map[string]any{"evidenceTypes": []string{"text"}}})["id"].(string)
	f.ok(f.owner, "jazebe_content_mission_activate", ref(content))
	now := time.Now().UTC()
	lottery := f.ok(f.owner, "jazebe_lottery_draft_create", map[string]any{"planetId": f.planetID, "title": "قرعه", "startsAt": now.Add(-time.Hour).Format(time.RFC3339),
		"drawAt": now.Add(time.Hour).Format(time.RFC3339), "ticketPricePoints": 1, "prizes": []map[string]any{{"title": "ج", "type": "cash", "valueRial": "1", "priority": 1}}})["id"].(string)

	meta := f.ok(f.owner, "jazebe_brand_mission_meta", map[string]any{"planetId": f.planetID})
	expectJSON(t, "meta", meta, map[string]any{"pointName": "امتیاز"})
	if n := len(meta["lotteries"].([]any)); n != 1 {
		t.Errorf("meta lotteries %d", n)
	}
	f.fails(u["fan"], "jazebe_brand_mission_meta", map[string]any{"planetId": f.planetID}, 7)
	row := f.ok(f.owner, "jazebe_brand_mission_get", ref(content))
	expectJSON(t, "row", row, map[string]any{"id": content, "functionalType": "content_creation", "status": "active", "rewardPoints": 30})

	upd := func(fields map[string]any) map[string]any {
		out := ref(content)
		for k, v := range fields {
			out[k] = v
		}
		return out
	}
	updated := f.ok(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"title": "محتوای ویرایش‌شده", "rewardPoints": 40, "sortOrder": 3}))
	expectJSON(t, "updated", updated, map[string]any{"title": "محتوای ویرایش‌شده", "rewardPoints": 40, "sortOrder": 3})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='mission.update' AND target_id=$1 AND params->'rewardPoints'='40'::jsonb`, content); n != 1 {
		t.Error("mission.update audit missing")
	}
	f.fails(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"rewardPoints": 1_000_000}), 3)
	f.fails(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"targetType": "group"}), 3)
	f.fails(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"rewardTickets": 2}), 3)
	f.ok(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"rewardTickets": 2, "targetLotteryId": lottery}))
	f.fails(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"displayConditions": map[string]any{"prerequisiteMissionIds": []string{content}}}), 3)
	f.ok(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"displayConditions": map[string]any{"prerequisiteMissionIds": []string{solo}}}))
	f.fails(f.owner, "jazebe_brand_mission_update", map[string]any{"planetId": f.planetID, "missionId": solo, "rewardPoints": 0}, 9)
	f.fails(u["fan"], "jazebe_brand_mission_update", upd(map[string]any{"title": "x"}), 7)

	// Bulk review: each completion decided and reported on its own.
	// Ticket rewards require a group; take them off so solo members can submit.
	f.ok(f.owner, "jazebe_brand_mission_update", upd(map[string]any{"rewardTickets": 0, "targetLotteryId": nil, "displayConditions": nil}))
	for _, name := range []string{"fan", "fan2"} {
		f.ok(u[name], "jazebe_content_mission_submit", map[string]any{"planetId": f.planetID, "missionId": content, "types": []string{"text"}, "note": "متن"})
	}
	var ids []string
	rows, err := f.db.QueryContext(f.ctx, `SELECT id FROM jazebe_mission_completion WHERE mission_id=$1 AND status='pending_review' ORDER BY id`, content)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	if len(ids) != 2 {
		t.Fatalf("pending completions %d", len(ids))
	}
	bogus := f.owner[:8] + "-0000-4000-8000-000000000001"
	bulk := map[string]any{"planetId": f.planetID, "completionIds": append(ids, bogus), "decision": "approve"}
	f.fails(u["fan"], "jazebe_brand_mission_review_bulk", bulk, 7)
	f.fails(f.owner, "jazebe_brand_mission_review_bulk", map[string]any{"planetId": f.planetID, "completionIds": []string{}, "decision": "approve"}, 3)
	results := f.ok(f.owner, "jazebe_brand_mission_review_bulk", bulk)["results"].([]any)
	if len(results) != 3 || results[0].(map[string]any)["ok"] != true || results[1].(map[string]any)["ok"] != true {
		t.Fatalf("bulk results %v", results)
	}
	expectJSON(t, "bogus", results[2].(map[string]any), map[string]any{"ok": false, "code": "NOT_FOUND"})
	for _, name := range []string{"fan", "fan2"} {
		if n := f.count(`SELECT COALESCE(sum(balance),0) FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, f.planetID, u[name]); n != 40 {
			t.Errorf("%s balance %d after bulk approve, want 40", name, n)
		}
	}
	again := f.ok(f.owner, "jazebe_brand_mission_review_bulk", map[string]any{"planetId": f.planetID, "completionIds": ids, "decision": "approve"})["results"].([]any)
	expectJSON(t, "repeat", again[0].(map[string]any), map[string]any{"ok": false, "code": "ALREADY_EXISTS"})
}
