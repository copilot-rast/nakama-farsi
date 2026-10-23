package main

import (
	"strings"
	"testing"
	"time"
)

func missionWindow() (string, string) {
	now := time.Now().UTC()
	return now.Add(-time.Hour).Format(time.RFC3339), now.Add(72 * time.Hour).Format(time.RFC3339)
}

// TestMissionWorkflowPostgres covers code and content missions from draft to
// payout: brand-admin authorization, activation/status/admin audits, code
// generation, one-use codes, repeat protection, the planet's last-active solo
// mission rule, manual review and the points ledger.
func TestMissionWorkflowPostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan", "fan2")
	u := f.users
	start, end := missionWindow()
	codeDraft := func(title string) map[string]any {
		return map[string]any{"planetId": f.planetID, "title": title, "detectionMethod": "platform_auto",
			"rewardPoints": 50, "rewardXp": 10, "startsAt": start, "endsAt": end,
			"code": map[string]any{"codeMode": "unique", "maxUses": 1}}
	}
	ref := func(id string) map[string]any { return map[string]any{"planetId": f.planetID, "missionId": id} }

	f.fails(u["fan"], "jazebe_code_mission_draft_create", codeDraft("ماموریت کد"), 7)
	f.fails(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "x", "detectionMethod": "guess",
		"startsAt": start, "endsAt": end, "code": map[string]any{"codeMode": "unique"}}, 3)
	m1 := f.ok(f.owner, "jazebe_code_mission_draft_create", codeDraft("ماموریت کد"))["id"].(string)
	f.fails(u["fan"], "jazebe_code_mission_activate", ref(m1), 7)
	f.ok(f.owner, "jazebe_code_mission_activate", ref(m1))
	actorType, actor, params := f.audit("mission.status", m1)
	if actorType != "brand_admin" || actor != f.owner {
		t.Errorf("activation attributed to %s/%s", actorType, actor)
	}
	expectJSON(t, "activate", params, map[string]any{"from": "draft", "to": "active"})

	f.fails(u["fan"], "jazebe_code_mission_generate", map[string]any{"planetId": f.planetID, "missionId": m1, "count": 2}, 7)
	gen := f.ok(f.owner, "jazebe_code_mission_generate", map[string]any{"planetId": f.planetID, "missionId": m1, "count": 3})
	_, _, params = f.audit("mission.codes.generate", m1)
	expectJSON(t, "generate", params, map[string]any{"created": 3})
	codesOut := gen["codes"].([]any)
	if len(codesOut) != 3 {
		t.Fatalf("generated %d codes", len(codesOut))
	}
	code := func(i int) string {
		switch c := codesOut[i].(type) {
		case string:
			return c
		case map[string]any:
			return c["code"].(string)
		}
		t.Fatalf("unexpected code shape %T", codesOut[i])
		return ""
	}

	redeem := func(user string, c string) *rpcErr {
		_, err := f.call(user, "jazebe_code_mission_redeem", map[string]any{"planetId": f.planetID, "missionId": m1, "code": c})
		return err
	}
	if err := redeem(u["fan"], code(0)); err != nil {
		t.Fatalf("redeem: %+v", err)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_mission_completion WHERE mission_id=$1 AND user_id=$2 AND status='approved'`, m1, u["fan"]); n != 1 {
		t.Errorf("approved completions %d", n)
	}
	if n := f.count(`SELECT COALESCE(sum(balance),0) FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, f.planetID, u["fan"]); n != 50 {
		t.Errorf("wallet balance %d after redeem, want 50", n)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_point_transaction t JOIN jazebe_point_wallet w ON w.id=t.wallet_id
		WHERE w.planet_id=$1 AND w.user_id=$2`, f.planetID, u["fan"]); n != 1 {
		t.Errorf("point transactions %d, want 1", n)
	}
	if n := f.count(`SELECT COALESCE(sum(amount),0) FROM jazebe_xp_record WHERE planet_id=$1 AND user_id=$2`, f.planetID, u["fan"]); n != 10 {
		t.Errorf("xp %d after redeem, want 10", n)
	}
	if err := redeem(u["fan2"], code(0)); err == nil || err.Code != 9 {
		t.Errorf("reused unique code: %+v, want 9", err)
	}
	if err := redeem(u["fan"], code(1)); err == nil || err.Code != 6 {
		t.Errorf("second completion of a once mission: %+v, want 6", err)
	}
	if err := redeem(u["fan2"], "not-a-real-code"); err == nil || err.Code != 3 {
		t.Errorf("unknown code: %+v, want 3", err)
	}
	if n := f.count(`SELECT COALESCE(sum(balance),0) FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, f.planetID, u["fan"]); n != 50 {
		t.Errorf("failed redemptions changed the balance to %d", n)
	}

	// The planet must keep one active solo point mission.
	f.fails(f.owner, "jazebe_mission_status", map[string]any{"planetId": f.planetID, "missionId": m1, "status": "paused"}, 9)
	m2 := f.ok(f.owner, "jazebe_code_mission_draft_create", codeDraft("ماموریت دوم"))["id"].(string)
	f.ok(f.owner, "jazebe_code_mission_activate", ref(m2))
	f.fails(u["fan"], "jazebe_mission_status", map[string]any{"planetId": f.planetID, "missionId": m1, "status": "paused"}, 7)
	f.ok(f.owner, "jazebe_mission_status", map[string]any{"planetId": f.planetID, "missionId": m1, "status": "paused"})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='mission.status' AND target_id=$1 AND params->>'from'='active' AND params->>'to'='paused'`, m1); n != 1 {
		t.Errorf("pause audit rows %d", n)
	}

	// Operator suspension: operator only, reason required and audited.
	f.fails(f.owner, "jazebe_mission_admin_suspend", map[string]any{"missionId": m2, "reason": "بررسی"}, 7)
	f.fails(f.operator, "jazebe_mission_admin_suspend", map[string]any{"missionId": m2, "reason": ""}, 3)
	f.ok(f.operator, "jazebe_mission_admin_suspend", map[string]any{"missionId": m2, "reason": "محتوای نامناسب"})
	actorType, actor, params = f.audit("mission.suspend", m2)
	if actorType != "operator" || actor != f.operator {
		t.Errorf("suspend attributed to %s/%s", actorType, actor)
	}
	expectJSON(t, "suspend", params, map[string]any{"reason": "محتوای نامناسب", "previousStatus": "active", "title": "ماموریت دوم"})
	f.fails(f.owner, "jazebe_mission_status", map[string]any{"planetId": f.planetID, "missionId": m2, "status": "active"}, 9)
	f.ok(f.operator, "jazebe_mission_admin_unsuspend", map[string]any{"missionId": m2, "reason": "رفع شد"})

	// Content mission with manual review.
	content := f.ok(f.owner, "jazebe_content_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "ماموریت محتوا",
		"rewardPoints": 30, "startsAt": start, "endsAt": end,
		"content": map[string]any{"evidenceTypes": []string{"text"}, "instructions": "یک متن بنویسید"}})["id"].(string)
	f.ok(f.owner, "jazebe_content_mission_activate", ref(content))
	_, _, params = f.audit("mission.status", content)
	expectJSON(t, "content activate", params, map[string]any{"from": "draft", "to": "active"})
	f.ok(u["fan2"], "jazebe_content_mission_submit", map[string]any{"planetId": f.planetID, "missionId": content, "types": []string{"text"}, "note": "متن من"})
	f.fails(u["fan2"], "jazebe_content_mission_submit", map[string]any{"planetId": f.planetID, "missionId": content, "types": []string{"text"}, "note": "دوباره"}, 6)
	var completion string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_mission_completion WHERE mission_id=$1 AND user_id=$2 AND status='pending_review'`, content, u["fan2"]).Scan(&completion); err != nil {
		t.Fatal(err)
	}
	review := map[string]any{"planetId": f.planetID, "completionId": completion, "decision": "approve"}
	f.fails(u["fan"], "jazebe_mission_review", review, 7)
	f.ok(f.owner, "jazebe_mission_review", review)
	_, actor, params = f.audit("mission.review", completion)
	expectJSON(t, "review", params, map[string]any{"decision": "approve"})
	if actor != f.owner {
		t.Errorf("review actor %s", actor)
	}
	if n := f.count(`SELECT COALESCE(sum(balance),0) FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, f.planetID, u["fan2"]); n != 30 {
		t.Errorf("reviewed completion paid %d, want 30", n)
	}
	f.fails(f.owner, "jazebe_mission_review", review, 6)
	if n := f.count(`SELECT COALESCE(sum(balance),0) FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, f.planetID, u["fan2"]); n != 30 {
		t.Errorf("repeated review changed the balance to %d", n)
	}
}

// TestMissionActivationAndGroupTargetPostgres activates media and invite
// missions and completes a group-target code mission, which pays tickets to
// the group once the target count is reached.
func TestMissionActivationAndGroupTargetPostgres(t *testing.T) {
	f := newJazebeFixture(t, "m1", "m2")
	u := f.users
	start, end := missionWindow()
	ref := func(id string) map[string]any { return map[string]any{"planetId": f.planetID, "missionId": id} }

	media := f.ok(f.owner, "jazebe_media_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "ویدئو را ببین",
		"detectionMethod": "platform_auto", "rewardPoints": 15, "startsAt": start, "endsAt": end,
		"media": map[string]any{"assetUrl": "https://cdn.example.com/v.mp4", "totalSec": 60, "minWatchSec": 30}})["id"].(string)
	f.fails(u["m1"], "jazebe_media_mission_activate", ref(media), 7)
	f.ok(f.owner, "jazebe_media_mission_activate", ref(media))
	_, actor, params := f.audit("mission.status", media)
	expectJSON(t, "media activate", params, map[string]any{"from": "draft", "to": "active"})
	if actor != f.owner {
		t.Errorf("media activation actor %s", actor)
	}

	invite := f.ok(f.owner, "jazebe_invite_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "دوستانت را دعوت کن",
		"rewardPoints": 20, "startsAt": start, "endsAt": end,
		"invite": map[string]any{"scope": "planet", "successCondition": "signup", "targetCount": 3, "windowDays": 7}})["id"].(string)
	f.ok(f.owner, "jazebe_invite_mission_activate", ref(invite))
	_, _, params = f.audit("mission.status", invite)
	expectJSON(t, "invite activate", params, map[string]any{"from": "draft", "to": "active"})

	now := time.Now().UTC()
	lottery := f.ok(f.owner, "jazebe_lottery_draft_create", map[string]any{"planetId": f.planetID, "title": "قرعه گروهی",
		"startsAt": now.Add(-time.Hour).Format(time.RFC3339), "drawAt": now.Add(24 * time.Hour).Format(time.RFC3339), "ticketPricePoints": 1,
		"prizes": []map[string]any{{"title": "جایزه", "type": "cash", "valueRial": "1000", "priority": 1}}})["id"].(string)
	f.ok(f.owner, "jazebe_lottery_activate", map[string]any{"planetId": f.planetID, "lotteryId": lottery})
	group := f.ok(u["m1"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "تیم هدف", "mode": "open", "cap": 4})["id"].(string)
	f.ok(u["m2"], "jazebe_group_join", map[string]any{"planetId": f.planetID, "groupId": group})
	target := f.ok(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "هدف گروهی",
		"detectionMethod": "platform_auto", "targetType": "group", "groupTargetValue": 2, "rewardTickets": 3, "targetLotteryId": lottery,
		"startsAt": start, "endsAt": end, "code": map[string]any{"codeMode": "unique", "maxUses": 1}})["id"].(string)
	f.ok(f.owner, "jazebe_code_mission_activate", ref(target))
	codes := f.ok(f.owner, "jazebe_code_mission_generate", map[string]any{"planetId": f.planetID, "missionId": target, "count": 2})["codes"].([]any)
	codeAt := func(i int) string {
		if m, ok := codes[i].(map[string]any); ok {
			return m["code"].(string)
		}
		return codes[i].(string)
	}
	f.ok(u["m1"], "jazebe_code_mission_redeem", map[string]any{"planetId": f.planetID, "missionId": target, "code": codeAt(0)})
	if n := f.count(`SELECT count(*) FROM jazebe_group_activity WHERE group_id=$1 AND message_key='mission.completed'`, group); n != 0 {
		t.Fatalf("group mission completed before reaching its target")
	}
	f.ok(u["m2"], "jazebe_code_mission_redeem", map[string]any{"planetId": f.planetID, "missionId": target, "code": codeAt(1)})
	actor, payload := f.activity(group, "mission.completed")
	expectJSON(t, "group target completed", payload, map[string]any{"missionId": target, "tickets": 3})
	if actor != "" {
		t.Errorf("group completion should have no single actor, got %s", actor)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND group_id=$2`, lottery, group); n != 3 {
		t.Errorf("group received %d tickets, want 3", n)
	}
}

// TestGroupInviteMissionJoinPostgres joins a newcomer to the inviter's group
// through a group-scope invite mission link, which writes the invite-link
// member.joined activity and qualifies the invitation for the inviter.
func TestGroupInviteMissionJoinPostgres(t *testing.T) {
	f := newJazebeFixture(t, "inviter")
	u := f.users
	start, end := missionWindow()
	group := f.ok(u["inviter"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "دعوت‌کنندگان", "mode": "open", "cap": 4})["id"].(string)
	mission := f.ok(f.owner, "jazebe_invite_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "دوستت را به گروه بیاور",
		"rewardPoints": 20, "startsAt": start, "endsAt": end,
		"invite": map[string]any{"scope": "group", "successCondition": "signup", "targetCount": 1, "windowDays": 7}})["id"].(string)
	f.ok(f.owner, "jazebe_invite_mission_activate", map[string]any{"planetId": f.planetID, "missionId": mission})
	link := f.ok(u["inviter"], "jazebe_referral_mission_link", map[string]any{"planetId": f.planetID, "missionId": mission})
	url, _ := link["url"].(string)
	token := url[strings.LastIndex(url, "/")+1:]
	if len(token) < 20 {
		t.Fatalf("invite link has no token: %v", link)
	}

	newcomer := f.user("newcomer")
	f.fails(u["inviter"], "jazebe_group_invite_join", map[string]any{"token": token}, 7)
	joined := f.ok(newcomer, "jazebe_group_invite_join", map[string]any{"token": token})
	expectJSON(t, "mission invite join", joined, map[string]any{"groupId": group, "joined": true})
	actor, payload := f.activity(group, "member.joined")
	expectJSON(t, "invite member.joined", payload, map[string]any{"userId": newcomer, "source": "invite_link", "name": "کاربر newcomer"})
	if actor != newcomer {
		t.Errorf("member.joined actor %s", actor)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2 AND join_path='group_invite_link'`, f.planetID, newcomer); n != 1 {
		t.Error("newcomer planet membership not attributed to the group invite link")
	}
	if n := f.count(`SELECT count(*) FROM jazebe_mission_invite WHERE mission_id=$1 AND inviter_user_id=$2 AND invitee_user_id=$3 AND status='qualified'`,
		mission, u["inviter"], newcomer); n != 1 {
		t.Error("group-scope invitation not qualified")
	}
	again := f.ok(newcomer, "jazebe_group_invite_join", map[string]any{"token": token})
	expectJSON(t, "repeat join", again, map[string]any{"joined": false})
	if n := f.count(`SELECT count(*) FROM jazebe_mission_invite WHERE mission_id=$1 AND invitee_user_id=$2`, mission, newcomer); n != 1 {
		t.Errorf("repeat join created %d invitation rows", n)
	}
}
