package main

import (
	"testing"
	"time"
)

// TestBrandTeamAndAdminPostgres covers the brand's own team page (source
// brand-team.service.ts), brand/me, and operator brand management (source
// planet-admin.service.ts brands section) including the contract switch.
func TestBrandTeamAndAdminPostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan")
	tech, mkt, outsider := f.user("tech"), f.user("mkt"), f.user("outsider")
	base := f.user("base-operator")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator+","+base)
	brand := map[string]any{"brandId": f.brandID}
	with := func(extra map[string]any) map[string]any {
		out := map[string]any{"brandId": f.brandID}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	var mktUsername string
	if err := f.db.QueryRowContext(f.ctx, `SELECT username FROM users WHERE id=$1`, mkt).Scan(&mktUsername); err != nil {
		t.Fatal(err)
	}

	// Lookup and invite: main administrator only, exact match.
	f.fails(tech, "jazebe_brand_team_lookup", with(map[string]any{"userId": tech}), 7)
	expectJSON(t, "lookup", f.ok(f.owner, "jazebe_brand_team_lookup", with(map[string]any{"username": mktUsername})), map[string]any{"found": true, "userId": mkt})
	expectJSON(t, "lookup miss", f.ok(f.owner, "jazebe_brand_team_lookup", with(map[string]any{"username": "no-such-user"})), map[string]any{"found": false})
	f.fails(f.owner, "jazebe_brand_team_invite", with(map[string]any{"userId": tech, "fullName": "ف", "role": "technical"}), 3)
	f.fails(f.owner, "jazebe_brand_team_invite", with(map[string]any{"userId": tech, "fullName": "فنی", "role": "owner"}), 3)
	team := f.ok(f.owner, "jazebe_brand_team_invite", with(map[string]any{"userId": tech, "fullName": "مدیر فنی", "role": "technical"}))["members"].([]any)
	if len(team) != 2 {
		t.Fatalf("team after invite: %d", len(team))
	}
	f.fails(f.owner, "jazebe_brand_team_invite", with(map[string]any{"userId": tech, "fullName": "مدیر فنی", "role": "technical"}), 6)
	f.ok(f.owner, "jazebe_brand_team_invite", with(map[string]any{"username": mktUsername, "fullName": "مدیر بازاریابی", "role": "marketing"}))
	_, actor, params := f.audit("brand.admin_add", tech)
	expectJSON(t, "admin_add", params, map[string]any{"userId": tech, "role": "technical"})
	if actor != f.owner {
		t.Errorf("admin_add actor %s", actor)
	}
	f.fails(tech, "jazebe_brand_team_list", brand, 7)

	// Roles: the brand always keeps an active main administrator.
	f.fails(f.owner, "jazebe_brand_team_role", with(map[string]any{"memberId": f.owner, "role": "technical"}), 9)
	f.ok(f.owner, "jazebe_brand_team_role", with(map[string]any{"memberId": tech, "role": "main"}))
	_, _, params = f.audit("brand.admin_role", tech)
	expectJSON(t, "admin_role", params, map[string]any{"from": "technical", "to": "main"})
	f.ok(tech, "jazebe_brand_team_list", brand)
	f.ok(tech, "jazebe_brand_team_role", with(map[string]any{"memberId": f.owner, "role": "technical"}))
	f.fails(tech, "jazebe_brand_team_deactivate", with(map[string]any{"memberId": tech}), 9)
	f.fails(tech, "jazebe_brand_team_role", with(map[string]any{"memberId": outsider, "role": "marketing"}), 5)

	// Deactivation removes brand powers; reactivation restores them.
	f.ok(tech, "jazebe_brand_team_deactivate", with(map[string]any{"memberId": mkt}))
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='brand.admin_deactivate' AND target_id=$1 AND params->>'role'='marketing'`, mkt); n != 1 {
		t.Error("deactivate audit missing")
	}
	now := time.Now().UTC()
	draft := map[string]any{"planetId": f.planetID, "title": "قرعه", "startsAt": now.Add(-time.Hour).Format(time.RFC3339),
		"drawAt": now.Add(time.Hour).Format(time.RFC3339), "ticketPricePoints": 1,
		"prizes": []map[string]any{{"title": "جایزه", "type": "cash", "valueRial": "100", "priority": 1}}}
	f.fails(mkt, "jazebe_lottery_draft_create", draft, 7)
	f.ok(tech, "jazebe_brand_team_reactivate", with(map[string]any{"memberId": mkt}))
	f.ok(mkt, "jazebe_lottery_draft_create", draft)

	// brand/me
	me := f.ok(tech, "jazebe_brand_me", map[string]any{})["items"].([]any)
	if len(me) != 1 {
		t.Fatalf("brand/me %d", len(me))
	}
	expectJSON(t, "brand/me", me[0].(map[string]any), map[string]any{"brandId": f.brandID, "role": "main", "fullName": "مدیر فنی", "contractStatus": "active", "isActive": true})
	if n := len(me[0].(map[string]any)["actions"].([]any)); n != 8 {
		t.Errorf("main role actions %d", n)
	}
	if n := len(f.ok(outsider, "jazebe_brand_me", map[string]any{})["items"].([]any)); n != 0 {
		t.Errorf("outsider administers %d brands", n)
	}

	// Operator reads, update and the managerial requirement.
	f.fails(f.users["fan"], "jazebe_operator_brand_list", map[string]any{}, 7)
	list := f.ok(base, "jazebe_operator_brand_list", map[string]any{"search": "برند " + f.owner[:8]})
	if list["total"].(float64) != 1 {
		t.Fatalf("operator brand search %v", list["total"])
	}
	row := list["items"].([]any)[0].(map[string]any)
	expectJSON(t, "brand row", row, map[string]any{"id": f.brandID, "contractStatus": "active", "isActive": true, "planetCount": 1})
	if n := len(row["admins"].([]any)); n != 3 {
		t.Errorf("brand row admins %d", n)
	}
	f.fails(base, "jazebe_operator_brand_update", map[string]any{"brandId": f.brandID, "legalName": "شرکت"}, 7)
	f.fails(f.operator, "jazebe_operator_brand_update", map[string]any{"brandId": f.brandID, "contactEmail": "not-an-email"}, 3)
	f.fails(f.operator, "jazebe_operator_brand_update", map[string]any{"brandId": f.brandID, "owner": "x"}, 3)
	updated := f.ok(f.operator, "jazebe_operator_brand_update", map[string]any{"brandId": f.brandID, "legalName": "شرکت آزمایشی",
		"contactEmail": "brand@example.com", "nationalCompanyId": "NC-" + f.owner[:8], "contractSignedAt": "2026-01-02T03:04:05Z"})
	expectJSON(t, "updated brand", updated, map[string]any{"legalName": "شرکت آزمایشی", "contactEmail": "brand@example.com"})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='brand.update' AND target_id=$1 AND jsonb_array_length(params->'fields')=4`, f.brandID); n != 1 {
		t.Error("brand.update audit missing")
	}
	second := f.ok(f.operator, "jazebe_brand_create", map[string]any{"name": "برند دوم " + f.owner[:8], "ownerUserId": outsider})["id"].(string)
	f.fails(f.operator, "jazebe_operator_brand_update", map[string]any{"brandId": second, "nationalCompanyId": "NC-" + f.owner[:8]}, 6)
	f.fails(base, "jazebe_brand_create", map[string]any{"name": "غیرمجاز", "ownerUserId": outsider}, 7)

	// Operator team changes are attributed to the operator.
	f.ok(f.operator, "jazebe_operator_brand_admin_add", map[string]any{"brandId": second, "userId": mkt, "fullName": "بازاریاب"})
	got := f.ok(f.operator, "jazebe_operator_brand_admin_role", map[string]any{"brandId": second, "memberId": mkt, "role": "technical"})
	if n := len(got["admins"].([]any)); n != 2 {
		t.Errorf("second brand admins %d", n)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='brand.admin_role' AND actor_type='operator' AND brand_id=$1`, second); n != 1 {
		t.Error("operator role change not attributed to the operator")
	}
	f.fails(f.operator, "jazebe_operator_brand_admin_deactivate", map[string]any{"brandId": second, "memberId": outsider}, 9)

	// Contract: suspension takes the brand down, termination is final.
	f.fails(f.operator, "jazebe_operator_brand_contract", map[string]any{"brandId": f.brandID, "status": "suspended", "reason": ""}, 3)
	suspended := f.ok(f.operator, "jazebe_operator_brand_contract", map[string]any{"brandId": f.brandID, "status": "suspended", "reason": "عدم پرداخت"})
	expectJSON(t, "suspended", suspended, map[string]any{"contractStatus": "suspended", "isActive": false})
	f.fails(f.users["fan"], "jazebe_planet_overview", map[string]any{"planetId": f.planetID}, 7)
	f.fails(mkt, "jazebe_lottery_draft_create", draft, 7)
	f.ok(f.operator, "jazebe_operator_brand_contract", map[string]any{"brandId": f.brandID, "status": "active", "reason": "تسویه شد"})
	f.ok(f.users["fan"], "jazebe_planet_overview", map[string]any{"planetId": f.planetID})
	f.ok(f.operator, "jazebe_operator_brand_contract", map[string]any{"brandId": second, "status": "terminated", "reason": "پایان قرارداد"})
	f.fails(f.operator, "jazebe_operator_brand_contract", map[string]any{"brandId": second, "status": "active", "reason": "بازگشت"}, 9)
	f.fails(f.operator, "jazebe_operator_brand_admin_add", map[string]any{"brandId": second, "userId": tech, "fullName": "فنی"}, 9)
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='brand.contract' AND target_id=$1 AND params->>'from'='active' AND params->>'to'='suspended'`, f.brandID); n != 1 {
		t.Error("contract audit missing")
	}
	// The second brand belongs to this fixture's cleanup through its owner.
	f.exec(`DELETE FROM jazebe_audit_log WHERE brand_id=$1`, second)
	f.exec(`DELETE FROM jazebe_brand_admin WHERE brand_id=$1`, second)
	f.exec(`DELETE FROM jazebe_brand_member WHERE brand_id=$1`, second)
	f.exec(`DELETE FROM jazebe_brand WHERE id=$1`, second)
}
