package main

import "testing"

// TestStaffGroupModerationPostgres covers the operator and brand-panel group
// views and operator suspend/unsuspend/disband (source group-admin and
// group-brand controllers, GroupModerationService, GroupService).
func TestStaffGroupModerationPostgres(t *testing.T) {
	f := newJazebeFixture(t, "lead", "mate", "seeker")
	u := f.users
	tech, marketer := f.user("tech"), f.user("marketer")
	f.ok(f.owner, "jazebe_brand_team_invite", map[string]any{"brandId": f.brandID, "userId": tech, "fullName": "فنی", "role": "technical"})
	f.ok(f.owner, "jazebe_brand_team_invite", map[string]any{"brandId": f.brandID, "userId": marketer, "fullName": "بازاریاب", "role": "marketing"})
	group := f.ok(u["lead"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "گروه نظارت", "mode": "open", "cap": 5})["id"].(string)
	f.ok(u["mate"], "jazebe_group_join", map[string]any{"planetId": f.planetID, "groupId": group})
	f.exec(`UPDATE jazebe_group SET mode='approval' WHERE id=$1`, group)
	f.ok(u["seeker"], "jazebe_group_join", map[string]any{"planetId": f.planetID, "groupId": group})
	f.ok(u["lead"], "jazebe_group_chat_send", map[string]any{"planetId": f.planetID, "groupId": group, "text": "سلام"})
	ref := map[string]any{"planetId": f.planetID, "groupId": group}
	opRef := map[string]any{"groupId": group}

	// Brand panel: moderation role only, addressed through the planet.
	list := f.ok(marketer, "jazebe_staff_group_list", map[string]any{"planetId": f.planetID})
	if list["total"].(float64) != 1 {
		t.Fatalf("brand list %v", list)
	}
	row := list["items"].([]any)[0].(map[string]any)
	expectJSON(t, "staff row", row, map[string]any{"id": group, "status": "active", "memberCount": 2, "pendingRequests": 1, "rank": 1, "mode": "approval"})
	expectJSON(t, "creator", row["creator"].(map[string]any), map[string]any{"id": u["lead"], "name": "کاربر lead"})
	f.fails(tech, "jazebe_staff_group_list", map[string]any{"planetId": f.planetID}, 7)
	f.fails(u["lead"], "jazebe_staff_group_get", ref, 7)
	f.fails(marketer, "jazebe_staff_group_get", opRef, 3)
	f.fails(marketer, "jazebe_staff_group_get", map[string]any{"planetId": f.planetOther(), "groupId": group}, 5)
	if n := len(f.ok(marketer, "jazebe_staff_group_members", ref)["members"].([]any)); n != 2 {
		t.Errorf("staff members %d", n)
	}
	if n := len(f.ok(marketer, "jazebe_staff_group_requests", ref)["requests"].([]any)); n != 1 {
		t.Errorf("staff requests %d", n)
	}
	f.ok(marketer, "jazebe_staff_group_activity", ref)
	f.ok(marketer, "jazebe_staff_group_history", ref)
	msgs := f.ok(marketer, "jazebe_staff_group_messages", ref)["items"].([]any)
	if len(msgs) != 1 {
		t.Errorf("staff chat %d", len(msgs))
	}
	if actorType, actor, _ := f.audit("group.chat.view", group); actorType != "brand_admin" || actor != marketer {
		t.Errorf("chat view audit %s %s", actorType, actor)
	}
	if n := len(f.ok(f.operator, "jazebe_staff_group_list", map[string]any{"brandId": f.brandID, "search": "نظارت", "status": "active"})["items"].([]any)); n != 1 {
		t.Errorf("operator filtered list %d", n)
	}
	if n := len(f.ok(f.operator, "jazebe_staff_group_list", map[string]any{"brandId": f.brandID, "search": "%"})["items"].([]any)); n != 0 {
		t.Errorf("search wildcard not escaped: %d", n)
	}

	// Operator suspension.
	f.fails(tech, "jazebe_staff_group_suspend", map[string]any{"planetId": f.planetID, "groupId": group, "reason": "تخلف"}, 7)
	f.fails(marketer, "jazebe_staff_group_suspend", map[string]any{"groupId": group, "reason": "تخلف"}, 3)
	f.fails(f.operator, "jazebe_staff_group_suspend", opRef, 3)
	suspended := f.ok(f.operator, "jazebe_staff_group_suspend", map[string]any{"groupId": group, "reason": "گزارش تخلف"})
	expectJSON(t, "suspended", suspended, map[string]any{"status": "suspended", "suspendedReason": "گزارش تخلف", "rank": nil})
	if actorType, actor, params := f.audit("group.suspend", group); actorType != "operator" || actor != f.operator || params["reason"] != "گزارش تخلف" {
		t.Errorf("suspend audit %s %s %v", actorType, actor, params)
	}
	f.fails(u["mate"], "jazebe_group_chat_send", map[string]any{"planetId": f.planetID, "groupId": group, "text": "هنوز؟"}, 9)
	if n := f.count(`SELECT count(*) FROM jazebe_group_message WHERE group_id=$1 AND kind='system' AND sender_user_id IS NULL`, group); n != 1 {
		t.Errorf("suspension system line %d", n)
	}
	if n := f.count(`SELECT count(*) FROM notification WHERE code=1001 AND subject='تعلیق گروه' AND user_id = ANY($1::uuid[])`, "{"+u["lead"]+","+u["mate"]+"}"); n != 2 {
		t.Errorf("suspension notices %d", n)
	}
	f.fails(f.operator, "jazebe_staff_group_suspend", map[string]any{"groupId": group, "reason": "دوباره"}, 9)
	if n := len(f.ok(marketer, "jazebe_staff_group_list", map[string]any{"planetId": f.planetID, "status": "suspended"})["items"].([]any)); n != 1 {
		t.Errorf("suspended filter %d", n)
	}
	restored := f.ok(f.operator, "jazebe_staff_group_unsuspend", opRef)
	expectJSON(t, "restored", restored, map[string]any{"status": "active", "suspendedReason": nil, "suspendedAt": nil})
	// The brand panel suspends through the planet too.
	f.ok(marketer, "jazebe_staff_group_suspend", map[string]any{"planetId": f.planetID, "groupId": group, "reason": "بررسی برند"})
	f.ok(marketer, "jazebe_staff_group_unsuspend", map[string]any{"planetId": f.planetID, "groupId": group, "reason": "رفع"})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action IN ('group.suspend','group.unsuspend') AND target_id=$1 AND actor_type='brand_admin' AND actor_id=$2`, group, marketer); n != 2 {
		t.Errorf("brand suspension audits %d", n)
	}
	f.ok(u["mate"], "jazebe_group_chat_send", map[string]any{"planetId": f.planetID, "groupId": group, "text": "برگشتیم"})

	// Disband needs a managerial operator.
	base := f.user("base")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator+","+base)
	f.ok(base, "jazebe_staff_group_suspend", map[string]any{"groupId": group, "reason": "بررسی"})
	f.fails(base, "jazebe_operator_group_disband", map[string]any{"groupId": group, "reason": "انحلال"}, 7)
	disbanded := f.ok(f.operator, "jazebe_operator_group_disband", map[string]any{"groupId": group, "reason": "تخلف مکرر"})
	expectJSON(t, "disbanded", disbanded, map[string]any{"status": "disbanded", "memberCount": 0})
	if actorType, _, params := f.audit("group.disband", group); actorType != "operator" || params["reason"] != "تخلف مکرر" || params["members"] != 2.0 {
		t.Errorf("disband audit %s %v", actorType, params)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_group_join_request WHERE group_id=$1 AND status='pending'`, group); n != 0 {
		t.Errorf("pending requests survived disband: %d", n)
	}
	f.fails(f.operator, "jazebe_operator_group_disband", map[string]any{"groupId": group, "reason": "دوباره"}, 9)
	// Staff still read a disbanded group's retained chat.
	if n := len(f.ok(f.operator, "jazebe_staff_group_messages", opRef)["items"].([]any)); n < 3 {
		t.Errorf("retained chat %d", n)
	}

}

// TestStaffMemberModerationPostgres covers staff removal from a group and
// planet member suspension (source GroupService.removeMemberByStaff,
// BrandModerationService).
func TestStaffMemberModerationPostgres(t *testing.T) {
	f := newJazebeFixture(t, "lead", "mate", "other")
	u := f.users
	tech := f.user("tech")
	f.ok(f.owner, "jazebe_brand_team_invite", map[string]any{"brandId": f.brandID, "userId": tech, "fullName": "فنی", "role": "technical"})
	group := f.ok(u["lead"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "گروه حذف", "mode": "open", "cap": 5})["id"].(string)
	f.ok(u["mate"], "jazebe_group_join", map[string]any{"planetId": f.planetID, "groupId": group})
	remove := func(user string) map[string]any {
		return map[string]any{"planetId": f.planetID, "groupId": group, "userId": user, "reason": "رفتار نامناسب"}
	}
	f.fails(tech, "jazebe_staff_group_remove_member", remove(u["mate"]), 7)
	f.fails(f.owner, "jazebe_staff_group_remove_member", remove(u["lead"]), 7)
	f.fails(f.owner, "jazebe_staff_group_remove_member", remove(u["other"]), 5)
	f.fails(f.owner, "jazebe_staff_group_remove_member", map[string]any{"planetId": f.planetID, "groupId": group, "userId": u["mate"]}, 3)
	f.ok(f.owner, "jazebe_staff_group_remove_member", remove(u["mate"]))
	if actorType, actor, params := f.audit("group.member.remove", group); actorType != "brand_admin" || actor != f.owner || params["userId"] != u["mate"] || params["reason"] != "رفتار نامناسب" {
		t.Errorf("removal audit %s %s %v", actorType, actor, params)
	}
	if n := f.count(`SELECT member_count FROM jazebe_group WHERE id=$1`, group); n != 1 {
		t.Errorf("member count %d", n)
	}

	// Planet member suspension.
	suspend := map[string]any{"planetId": f.planetID, "userId": u["lead"], "reason": "تخلف در سیاره"}
	f.fails(tech, "jazebe_brand_member_suspend", suspend, 7)
	f.fails(f.owner, "jazebe_brand_member_suspend", map[string]any{"planetId": f.planetID, "userId": u["lead"], "reason": "x"}, 3)
	f.ok(f.owner, "jazebe_brand_member_suspend", suspend)
	f.fails(f.owner, "jazebe_brand_member_suspend", suspend, 9)
	f.fails(u["lead"], "jazebe_group_chat_send", map[string]any{"planetId": f.planetID, "groupId": group, "text": "سلام"}, 7)
	if n := f.count(`SELECT count(*) FROM notification WHERE code=1001 AND user_id=$1 AND subject='دسترسی تو به سیاره محدود شد'`, u["lead"]); n != 1 {
		t.Errorf("suspension notice %d", n)
	}
	if actorType, _, params := f.audit("planet.member.suspend", u["lead"]); actorType != "brand_admin" || params["reason"] != "تخلف در سیاره" {
		t.Errorf("suspend audit %s %v", actorType, params)
	}
	list := f.ok(f.owner, "jazebe_brand_suspended_members", map[string]any{"brandId": f.brandID})
	items := list["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("suspended list %v", list)
	}
	expectJSON(t, "suspended member", items[0].(map[string]any), map[string]any{"userId": u["lead"], "displayName": "کاربر lead"})
	if items[0].(map[string]any)["suspendedAt"] == nil {
		t.Error("suspendedAt missing")
	}
	f.fails(tech, "jazebe_brand_suspended_members", map[string]any{"brandId": f.brandID}, 7)
	f.fails(f.owner, "jazebe_brand_member_unsuspend", map[string]any{"planetId": f.planetID, "userId": u["other"], "reason": "رفع تعلیق"}, 9)
	f.ok(f.owner, "jazebe_brand_member_unsuspend", map[string]any{"planetId": f.planetID, "userId": u["lead"], "reason": "رفع تعلیق"})
	f.ok(u["lead"], "jazebe_group_chat_send", map[string]any{"planetId": f.planetID, "groupId": group, "text": "برگشتم"})
	if n := len(f.ok(f.owner, "jazebe_brand_suspended_members", map[string]any{"brandId": f.brandID})["items"].([]any)); n != 0 {
		t.Errorf("lifted suspension still listed: %d", n)
	}
}
