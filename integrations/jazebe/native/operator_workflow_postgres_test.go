package main

import (
	"testing"
)

// TestPlanetCategoryAndLifecyclePostgres covers the operator category
// catalogue (seeded by migration) and a planet's full lifecycle from setup
// through publish request, operator rejection, approval, suspension,
// unsuspension and archive, asserting each transition's audit row.
func TestPlanetCategoryAndLifecyclePostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan")
	manager := f.user("manager")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator+","+manager)
	t.Setenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS", manager)

	// Seeded, ordered catalogue is visible to brand admins only.
	listed := f.ok(f.owner, "jazebe_brand_planet_categories", map[string]any{})["items"].([]any)
	if len(listed) < 12 || listed[0].(map[string]any)["key"] != "food" {
		t.Fatalf("seeded categories missing or misordered: %d", len(listed))
	}
	f.fails(f.users["fan"], "jazebe_brand_planet_categories", map[string]any{}, 7)

	// Operator CRUD: list for any operator, writes for managerial operators.
	f.fails(f.users["fan"], "jazebe_operator_planet_category_list", map[string]any{}, 7)
	page := f.ok(f.operator, "jazebe_operator_planet_category_list", map[string]any{"limit": 5})
	if page["total"].(float64) < 12 || len(page["items"].([]any)) != 5 {
		t.Errorf("operator category page: %v", page)
	}
	key := "fx_" + f.owner[:8]
	create := map[string]any{"key": key, "titleFa": "دسته آزمایشی", "titleEn": "Fixture", "sortOrder": 500}
	f.fails(f.operator, "jazebe_operator_planet_category_create", create, 7)
	f.fails(manager, "jazebe_operator_planet_category_create", map[string]any{"key": "Bad Key", "titleFa": "x"}, 3)
	created := f.ok(manager, "jazebe_operator_planet_category_create", create)
	categoryID := created["id"].(string)
	expectJSON(t, "category", created, map[string]any{"key": key, "titleFa": "دسته آزمایشی", "isActive": true})
	f.fails(manager, "jazebe_operator_planet_category_create", create, 6)
	_, actor, params := f.audit("planet_category.create", categoryID)
	expectJSON(t, "category.create", params, map[string]any{"key": key})
	if actor != manager {
		t.Errorf("category.create actor %s", actor)
	}
	updated := f.ok(manager, "jazebe_operator_planet_category_update", map[string]any{"id": categoryID, "titleFa": "دسته ویرایش‌شده", "sortOrder": 5})
	expectJSON(t, "updated", updated, map[string]any{"titleFa": "دسته ویرایش‌شده", "sortOrder": 5})
	f.fails(manager, "jazebe_operator_planet_category_update", map[string]any{"id": categoryID, "key": "renamed"}, 3)
	f.fails(manager, "jazebe_operator_planet_category_update", map[string]any{"id": f.owner, "titleFa": "x"}, 5)

	// Planet setup: the publish check reports what is missing.
	planet := f.ok(f.owner, "jazebe_planet_create", map[string]any{"brandId": f.brandID, "name": "سیاره چرخه", "slug": "cycle-" + f.owner[:8],
		"latinName": "Cycle", "pointName": "سکه"})["id"].(string)
	pref := map[string]any{"planetId": planet}
	check := f.ok(f.owner, "jazebe_planet_publish_check", pref)
	if missing, _ := check["missing"].([]any); len(missing) == 0 {
		t.Errorf("incomplete planet reported publishable: %v", check)
	}
	f.fails(f.owner, "jazebe_planet_publish_request", pref, 9)
	f.ok(f.owner, "jazebe_brand_planet_update", map[string]any{"planetId": planet, "shortDescription": "سیاره آزمایشی چرخه عمر",
		"categoryId": categoryID, "brandColor": "#2F6EBC", "glyph": "⭐"})
	start, end := missionWindow()
	mission := f.ok(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": planet, "title": "اولین ماموریت",
		"detectionMethod": "platform_auto", "rewardPoints": 10, "startsAt": start, "endsAt": end,
		"code": map[string]any{"codeMode": "unique", "maxUses": 1}})["id"].(string)
	f.ok(f.owner, "jazebe_code_mission_activate", map[string]any{"planetId": planet, "missionId": mission})
	check = f.ok(f.owner, "jazebe_planet_publish_check", pref)
	if missing, _ := check["missing"].([]any); len(missing) != 0 {
		t.Fatalf("complete planet still missing %v", missing)
	}

	// Soft-deleted categories stay referenced but cannot be newly chosen.
	f.fails(manager, "jazebe_operator_planet_category_delete", map[string]any{"id": categoryID, "reason": "x"}, 3)

	transition := func(action string, payload map[string]any, from, to string) {
		t.Helper()
		f.ok(f.operator, "jazebe_planet_"+action, payload)
		var n int
		n = f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action=$1 AND target_id=$2 AND params->>'from'=$3 AND params->>'to'=$4`,
			"planet."+action, planet, from, to)
		if n < 1 {
			t.Errorf("planet.%s audit %s->%s missing", action, from, to)
		}
		if got := f.count(`SELECT count(*) FROM jazebe_planet WHERE id=$1 AND status=$2`, planet, to); got != 1 {
			t.Errorf("after %s planet is not %s", action, to)
		}
	}
	f.ok(f.owner, "jazebe_planet_publish_request", pref)
	f.fails(f.owner, "jazebe_planet_approve", pref, 7)
	f.fails(f.operator, "jazebe_planet_reject", pref, 3)
	transition("reject", map[string]any{"planetId": planet, "reason": "توضیحات ناقص است"}, "pending_review", "setup")
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='planet.reject' AND target_id=$1 AND params->>'reason'='توضیحات ناقص است'`, planet); n != 1 {
		t.Error("rejection reason not audited")
	}
	f.ok(f.owner, "jazebe_planet_publish_request", pref)
	transition("approve", pref, "pending_review", "active")
	f.fails(f.operator, "jazebe_planet_approve", pref, 9)
	f.fails(f.operator, "jazebe_planet_suspend", pref, 3)
	transition("suspend", map[string]any{"planetId": planet, "reason": "تخلف"}, "active", "suspended")
	transition("unsuspend", pref, "suspended", "active")
	transition("archive", map[string]any{"planetId": planet, "reason": "پایان همکاری"}, "active", "archived")
	f.fails(f.operator, "jazebe_planet_archive", map[string]any{"planetId": planet, "reason": "دوباره"}, 9)

	deleted := f.ok(manager, "jazebe_operator_planet_category_delete", map[string]any{"id": categoryID, "reason": "دسته تکراری"})
	expectJSON(t, "deleted", deleted, map[string]any{"isActive": false})
	_, _, params = f.audit("planet_category.delete", categoryID)
	expectJSON(t, "category.delete", params, map[string]any{"key": key, "reason": "دسته تکراری"})
	f.fails(manager, "jazebe_operator_planet_category_delete", map[string]any{"id": categoryID, "reason": "دوباره"}, 5)
	for _, item := range f.ok(f.owner, "jazebe_brand_planet_categories", map[string]any{})["items"].([]any) {
		if item.(map[string]any)["id"] == categoryID {
			t.Error("deleted category still offered to brands")
		}
	}
	if n := f.count(`SELECT count(*) FROM jazebe_planet WHERE id=$1 AND category_id=$2`, planet, categoryID); n != 1 {
		t.Error("soft delete broke the planet's category reference")
	}
}

// TestWalletAdjustAndProfilePostgres covers operator point adjustments
// (authorization, super-operator threshold, idempotency, overdraft, audit and
// notification) and profile re-setup with an avatar preset.
func TestWalletAdjustAndProfilePostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan")
	fan := f.users["fan"]
	super := f.user("super")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator+","+super)
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", super)
	adjust := func(direction string, amount int64, key string) map[string]any {
		return map[string]any{"planetId": f.planetID, "userId": fan, "direction": direction, "amount": amount,
			"reason": "جبران خطای سیستم", "reference": "TICKET-1", "idempotencyKey": key}
	}
	f.fails(fan, "jazebe_wallet_adjust", adjust("credit", 100, "adj-key-0001"), 7)
	f.fails(f.operator, "jazebe_wallet_adjust", adjust("credit", 100, "short"), 3)
	first := f.ok(f.operator, "jazebe_wallet_adjust", adjust("credit", 100, "adj-key-0001"))
	expectJSON(t, "credit", first, map[string]any{"balance": 100, "applied": true, "large": false})
	replay := f.ok(f.operator, "jazebe_wallet_adjust", adjust("credit", 100, "adj-key-0001"))
	expectJSON(t, "replay", replay, map[string]any{"balance": 100, "applied": false})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='wallet.adjust' AND params->>'userId'=$1`, fan); n != 1 {
		t.Errorf("idempotent replay wrote %d audit rows", n)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='wallet.adjust' AND params->>'userId'=$1
		AND params->>'direction'='credit' AND params->'amount'='100'::jsonb AND params->'balanceAfter'='100'::jsonb
		AND params->'large'='false'::jsonb AND params->>'reference'='TICKET-1'`, fan); n != 1 {
		t.Error("wallet.adjust audit params wrong")
	}
	f.fails(f.operator, "jazebe_wallet_adjust", adjust("debit", 500, "adj-key-0002"), 9)
	f.fails(f.operator, "jazebe_wallet_adjust", adjust("credit", 20000, "adj-key-0003"), 7)
	big := f.ok(super, "jazebe_wallet_adjust", adjust("credit", 20000, "adj-key-0003"))
	expectJSON(t, "large", big, map[string]any{"balance": 20100, "large": true})
	if n := f.count(`SELECT count(*) FROM notification WHERE user_id=$1`, fan); n < 2 {
		t.Errorf("adjustment notices %d", n)
	}

	// Profile setup, then re-setup switching avatar preset.
	f.ok(fan, "jazebe_profile_setup", map[string]any{"displayName": "طرفدار" + fan[:4], "avatarIndex": 2})
	f.ok(fan, "jazebe_profile_setup", map[string]any{"displayName": "طرفدار" + fan[:4], "avatarIndex": 4})
	if n := f.count(`SELECT count(*) FROM jazebe_user_profile WHERE user_id=$1 AND avatar_kind='preset' AND avatar_preset_index=4`, fan); n != 1 {
		t.Error("avatar re-setup not persisted")
	}
	f.fails(fan, "jazebe_profile_setup", map[string]any{"displayName": "طرفدار" + fan[:4], "avatarIndex": 9}, 3)
}
