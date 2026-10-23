package main

import (
	"sync"
	"testing"
)

// TestGroupWorkflowPostgres drives one planet's groups through creation,
// open and approval joins, invites, roles, leadership transfer, removal,
// leaving, chat moderation, report dismissal and disband, asserting the
// persisted state and every audit and activity row each step writes.
func TestGroupWorkflowPostgres(t *testing.T) {
	f := newJazebeFixture(t, "creator", "member", "approver", "applicant", "invitee", "outsider", "late")
	u := f.users
	planet := f.planetID
	ref := func(group string) map[string]any { return map[string]any{"planetId": planet, "groupId": group} }

	// Create: validation, persistence, audit and activity.
	f.fails(u["creator"], "jazebe_group_create", map[string]any{"planetId": planet, "name": "", "mode": "open", "cap": 3}, 3)
	f.fails(u["creator"], "jazebe_group_create", map[string]any{"planetId": planet, "name": "گروه", "mode": "open", "cap": 500}, 3)
	created := f.ok(u["creator"], "jazebe_group_create", map[string]any{"planetId": planet, "name": "گروه ستاره", "mode": "open", "cap": 3, "icon": "star"})
	g1 := created["id"].(string)
	actorType, actor, params := f.audit("group.create", g1)
	if actorType != "user" || actor != u["creator"] {
		t.Errorf("group.create attributed to %s/%s", actorType, actor)
	}
	expectJSON(t, "group.create", params, map[string]any{"name": "گروه ستاره"})
	actor, payload := f.activity(g1, "group.created")
	expectJSON(t, "group.created", payload, map[string]any{"name": "گروه ستاره", "text": "کاربر creator گروه را ساخت"})
	if actor != u["creator"] {
		t.Errorf("group.created actor %s", actor)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND role='creator' AND status='active'`, g1, u["creator"]); n != 1 {
		t.Errorf("creator membership rows %d", n)
	}
	f.fails(u["creator"], "jazebe_group_create", map[string]any{"planetId": planet, "name": "دوم", "mode": "open", "cap": 3}, 6)
	f.fails(u["approver"], "jazebe_group_create", map[string]any{"planetId": planet, "name": "گروه ستاره", "mode": "open", "cap": 3}, 6)
	// Creating a group is a first planet activity: it adds the planet membership.
	if n := f.count(`SELECT count(*) FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2`, planet, f.owner); n != 0 {
		t.Fatalf("owner unexpectedly already a planet member")
	}
	ownerGroup := f.ok(f.owner, "jazebe_group_create", map[string]any{"planetId": planet, "name": "گروه مالک", "mode": "open", "cap": 2})["id"].(string)
	if n := f.count(`SELECT count(*) FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2 AND join_path='first_activity'`, planet, f.owner); n != 1 {
		t.Errorf("group creation did not record first-activity planet membership")
	}
	f.ok(f.owner, "jazebe_group_disband", ref(ownerGroup))

	// Open join and duplicate join.
	joined := f.ok(u["member"], "jazebe_group_join", ref(g1))
	expectJSON(t, "join", joined, map[string]any{"joined": true})
	actor, payload = f.activity(g1, "member.joined")
	expectJSON(t, "member.joined", payload, map[string]any{"userId": u["member"], "name": "کاربر member", "source": "list_join"})
	if actor != u["member"] {
		t.Errorf("member.joined actor %s", actor)
	}
	// Source group-request.service.ts:71: rejoining the same group is idempotent.
	expectJSON(t, "rejoin", f.ok(u["member"], "jazebe_group_join", ref(g1)), map[string]any{"joined": true, "requested": false})
	if n := f.count(`SELECT count(*) FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2`, g1, u["member"]); n != 1 {
		t.Errorf("rejoin created %d membership rows", n)
	}
	if n := f.count(`SELECT member_count FROM jazebe_group WHERE id=$1`, g1); n != 2 {
		t.Errorf("member_count=%d after open join, want 2", n)
	}

	// Approval group: request, wrong-manager decision, approval.
	g2 := f.ok(u["approver"], "jazebe_group_create", map[string]any{"planetId": planet, "name": "گروه تأیید", "mode": "approval", "cap": 5})["id"].(string)
	requested := f.ok(u["applicant"], "jazebe_group_join", ref(g2))
	expectJSON(t, "approval join", requested, map[string]any{"joined": false, "requested": true})
	var requestID string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_group_join_request WHERE group_id=$1 AND user_id=$2 AND status='pending'`, g2, u["applicant"]).Scan(&requestID); err != nil {
		t.Fatal(err)
	}
	decision := map[string]any{"planetId": planet, "groupId": g2, "requestId": requestID, "decision": "approve"}
	f.fails(u["member"], "jazebe_group_decide_request", decision, 7)
	f.ok(u["approver"], "jazebe_group_decide_request", decision)
	_, actor, params = f.audit("group.request.decision", requestID)
	expectJSON(t, "group.request.decision", params, map[string]any{"decision": "approve", "groupId": g2})
	if actor != u["approver"] {
		t.Errorf("decision actor %s", actor)
	}
	_, payload = f.activity(g2, "member.joined")
	expectJSON(t, "approved join", payload, map[string]any{"userId": u["applicant"], "source": "approved_request"})
	f.fails(u["approver"], "jazebe_group_decide_request", decision, 6)

	// Invite link join, then a full group.
	invite := f.ok(u["member"], "jazebe_group_invite_create", ref(g1))
	token := invite["token"].(string)
	f.fails(u["outsider"], "jazebe_group_invite_create", ref(g1), 7)
	f.ok(u["invitee"], "jazebe_group_invite_join", map[string]any{"token": token})
	_, payload = f.activity(g1, "member.joined")
	expectJSON(t, "invite join", payload, map[string]any{"userId": u["invitee"], "source": "invite_link"})
	f.fails(u["late"], "jazebe_group_join", ref(g1), 8)

	// Roles: promote, then offer and accept leadership.
	f.fails(u["member"], "jazebe_group_set_role", map[string]any{"planetId": planet, "groupId": g1, "memberId": u["invitee"], "role": "admin"}, 7)
	f.ok(u["creator"], "jazebe_group_set_role", map[string]any{"planetId": planet, "groupId": g1, "memberId": u["member"], "role": "admin"})
	_, payload = f.activity(g1, "member.promote")
	expectJSON(t, "promotion", payload, map[string]any{"userId": u["member"], "role": "admin"})
	_, actor, params = f.audit("group.member.promote", g1)
	expectJSON(t, "group.member.promote", params, map[string]any{"userId": u["member"]})
	if actor != u["creator"] {
		t.Errorf("promotion actor %s", actor)
	}
	f.ok(u["creator"], "jazebe_group_set_role", map[string]any{"planetId": planet, "groupId": g1, "memberId": u["member"], "role": "creator"})
	var offerID string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_group_leadership_offer WHERE group_id=$1 AND status='pending'`, g1).Scan(&offerID); err != nil {
		t.Fatal(err)
	}
	_, _, params = f.audit("group.transfer.offer", g1)
	expectJSON(t, "transfer offer", params, map[string]any{"to": u["member"], "offerId": offerID})
	f.ok(u["member"], "jazebe_group_leadership_respond", map[string]any{"planetId": planet, "groupId": g1, "accept": true})
	_, actor, params = f.audit("group.transfer", g1)
	expectJSON(t, "transfer", params, map[string]any{"from": u["creator"], "to": u["member"], "offerId": offerID})
	if actor != u["member"] {
		t.Errorf("group.transfer actor %s", actor)
	}
	_, payload = f.activity(g1, "member.transfer")
	expectJSON(t, "member.transfer", payload, map[string]any{"userId": u["member"], "fromUserId": u["creator"], "groupName": "گروه ستاره"})
	if n := f.count(`SELECT count(*) FROM jazebe_group WHERE id=$1 AND creator_user_id=$2`, g1, u["member"]); n != 1 {
		t.Error("creator_user_id not transferred")
	}

	// Decline path on the approval group.
	f.ok(u["approver"], "jazebe_group_set_role", map[string]any{"planetId": planet, "groupId": g2, "memberId": u["applicant"], "role": "admin"})
	f.ok(u["approver"], "jazebe_group_set_role", map[string]any{"planetId": planet, "groupId": g2, "memberId": u["applicant"], "role": "creator"})
	f.ok(u["applicant"], "jazebe_group_leadership_respond", map[string]any{"planetId": planet, "groupId": g2, "accept": false})
	_, _, params = f.audit("group.transfer.decline", g2)
	if params["offerId"] == nil || params["offerId"] == "" {
		t.Errorf("decline audit missing offerId: %v", params)
	}

	// Removal and leaving.
	f.fails(u["invitee"], "jazebe_group_remove_member", map[string]any{"planetId": planet, "groupId": g1, "memberId": u["creator"]}, 7)
	f.ok(u["member"], "jazebe_group_remove_member", map[string]any{"planetId": planet, "groupId": g1, "memberId": u["invitee"]})
	_, _, params = f.audit("group.member.remove", g1)
	expectJSON(t, "member.remove", params, map[string]any{"userId": u["invitee"]})
	_, payload = f.activity(g1, "member.removed")
	expectJSON(t, "member.removed", payload, map[string]any{"userId": u["invitee"]})
	f.ok(u["creator"], "jazebe_group_leave", ref(g1))
	_, payload = f.activity(g1, "member.left")
	expectJSON(t, "member.left", payload, map[string]any{"userId": u["creator"], "name": "کاربر creator"})
	f.fails(u["member"], "jazebe_group_leave", ref(g1), 7)

	// Chat: send, member cannot moderate, admin deletes as moderator, report and dismissal.
	f.nk.present = []string{u["applicant"], u["approver"], u["outsider"]}
	sent := f.ok(u["applicant"], "jazebe_group_chat_send", map[string]any{"planetId": planet, "groupId": g2, "text": "سلام همه"})
	seq := int64(sent["seq"].(float64))
	f.fails(u["outsider"], "jazebe_group_chat_send", map[string]any{"planetId": planet, "groupId": g2, "text": "نفوذ"}, 7)
	report := f.ok(u["approver"], "jazebe_group_chat_report", map[string]any{"planetId": planet, "groupId": g2, "seq": seq, "reason": "spam"})
	reportID := report["id"].(string)
	f.fails(u["applicant"], "jazebe_group_chat_dismiss_report", map[string]any{"reportId": reportID, "reason": "بررسی شد"}, 7)
	f.ok(f.owner, "jazebe_group_chat_dismiss_report", map[string]any{"reportId": reportID, "reason": "بررسی شد"})
	actorType, actor, params = f.audit("user_report.dismiss", reportID)
	if actorType != "brand_admin" || actor != f.owner {
		t.Errorf("dismiss attributed to %s/%s", actorType, actor)
	}
	expectJSON(t, "dismiss", params, map[string]any{"reason": "بررسی شد"})
	f.fails(f.owner, "jazebe_group_chat_dismiss_report", map[string]any{"reportId": reportID, "reason": "دوباره"}, 6)

	second := f.ok(u["approver"], "jazebe_group_chat_send", map[string]any{"planetId": planet, "groupId": g2, "text": "پیام دوم"})
	seq2 := int64(second["seq"].(float64))
	f.ok(u["applicant"], "jazebe_group_chat_delete", map[string]any{"planetId": planet, "groupId": g2, "seq": seq2, "asAdmin": true})
	var messageID string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_group_message WHERE group_id=$1 AND seq=$2`, g2, seq2).Scan(&messageID); err != nil {
		t.Fatal(err)
	}
	_, actor, params = f.audit("group.message.delete", messageID)
	expectJSON(t, "admin delete", params, map[string]any{"groupId": g2, "seq": seq2, "senderUserId": u["approver"]})
	if actor != u["applicant"] {
		t.Errorf("admin delete actor %s", actor)
	}
	f.fails(u["approver"], "jazebe_group_chat_delete", map[string]any{"planetId": planet, "groupId": g2, "seq": seq2}, 6)

	f.fails(u["applicant"], "jazebe_group_chat_moderate", map[string]any{"planetId": planet, "groupId": g2, "seq": seq, "reason": "تخلف"}, 7)
	f.ok(f.operator, "jazebe_group_chat_moderate", map[string]any{"planetId": planet, "groupId": g2, "seq": seq, "reason": "تخلف"})
	var firstID string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_group_message WHERE group_id=$1 AND seq=$2`, g2, seq).Scan(&firstID); err != nil {
		t.Fatal(err)
	}
	actorType, actor, params = f.audit("group.message.delete", firstID)
	if actorType != "operator" || actor != f.operator {
		t.Errorf("moderation attributed to %s/%s", actorType, actor)
	}
	expectJSON(t, "moderate", params, map[string]any{"groupId": g2, "seq": seq, "reason": "تخلف", "senderUserId": u["applicant"]})

	// Disband: only the creator, audit carries the member count.
	f.fails(u["applicant"], "jazebe_group_disband", ref(g2), 7)
	f.ok(u["approver"], "jazebe_group_disband", ref(g2))
	_, _, params = f.audit("group.disband", g2)
	expectJSON(t, "disband", params, map[string]any{"members": 2})
	if n := f.count(`SELECT count(*) FROM jazebe_group_membership WHERE group_id=$1 AND status='active'`, g2); n != 0 {
		t.Errorf("%d active memberships remain after disband", n)
	}
	if len(f.nk.sends) == 0 || f.nk.recipients[u["approver"]] == 0 {
		t.Errorf("chat events not delivered to members: %d sends", len(f.nk.sends))
	}
	if f.nk.recipients[u["outsider"]] != 0 {
		t.Errorf("chat events delivered to a non-member presence %d times", f.nk.recipients[u["outsider"]])
	}
}

// TestGroupJoinConcurrencyPostgres races joins that pre-checks alone cannot
// make safe: many users for the last seat of a group, and one user for two
// groups of the same planet (a user may be in one group per planet).
func TestGroupJoinConcurrencyPostgres(t *testing.T) {
	racers := []string{"r1", "r2", "r3", "r4", "r5", "r6"}
	f := newJazebeFixture(t, append([]string{"lead", "lead2", "solo"}, racers...)...)
	ref := func(group string) map[string]any { return map[string]any{"planetId": f.planetID, "groupId": group} }
	seat := f.ok(f.users["lead"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "یک جا", "mode": "open", "cap": 2})["id"].(string)

	var wg sync.WaitGroup
	results := make([]*rpcErr, len(racers))
	start := make(chan struct{})
	for i, name := range racers {
		wg.Add(1)
		go func(i int, user string) {
			defer wg.Done()
			<-start
			_, results[i] = f.call(user, "jazebe_group_join", ref(seat))
		}(i, f.users[name])
	}
	close(start)
	wg.Wait()
	won := 0
	for i, err := range results {
		switch {
		case err == nil:
			won++
		case err.Code != 8:
			t.Errorf("%s: unexpected %+v", racers[i], err)
		}
	}
	if won != 1 {
		t.Errorf("%d racers joined a group with one free seat", won)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_group_membership WHERE group_id=$1 AND status='active'`, seat); n != 2 {
		t.Errorf("active memberships=%d, want 2", n)
	}
	if n := f.count(`SELECT member_count FROM jazebe_group WHERE id=$1`, seat); n != 2 {
		t.Errorf("member_count=%d, want 2", n)
	}

	a := f.ok(f.users["lead2"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "الف", "mode": "open", "cap": 10})["id"].(string)
	other := f.user("other-lead")
	f.ok(other, "jazebe_planet_join", map[string]any{"planetId": f.planetID})
	b := f.ok(other, "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "ب", "mode": "open", "cap": 10})["id"].(string)
	for round := 0; round < 5; round++ {
		user := f.user("double-" + string(rune('a'+round)))
		f.ok(user, "jazebe_planet_join", map[string]any{"planetId": f.planetID})
		var dwg sync.WaitGroup
		go1 := make(chan struct{})
		for _, group := range []string{a, b} {
			dwg.Add(1)
			go func(group string) {
				defer dwg.Done()
				<-go1
				f.call(user, "jazebe_group_join", ref(group))
			}(group)
		}
		close(go1)
		dwg.Wait()
		if n := f.count(`SELECT count(*) FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`, f.planetID, user); n != 1 {
			t.Errorf("round %d: user holds %d active groups in one planet", round, n)
		}
	}
	for _, group := range []string{seat, a, b} {
		var stored, actual int
		if err := f.db.QueryRowContext(f.ctx, `SELECT g.member_count,(SELECT count(*) FROM jazebe_group_membership m WHERE m.group_id=g.id AND m.status='active')
			FROM jazebe_group g WHERE g.id=$1`, group).Scan(&stored, &actual); err != nil {
			t.Fatal(err)
		}
		if stored != actual {
			t.Errorf("group %s member_count %d != active memberships %d", group, stored, actual)
		}
	}
}
