package main

import (
	"testing"
	"time"
)

// TestOperatorAccessAndAuditPostgres covers operator/me, the operator
// directory and the audit-log explorer (source operator-access and
// audit-log-explorer services).
func TestOperatorAccessAndAuditPostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan")
	manager, super := f.user("manager"), f.user("super")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator+","+manager+","+super)
	t.Setenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS", manager+","+f.operator)
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", super)
	base := f.user("base")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator+","+manager+","+super+","+base)

	me := f.ok(base, "jazebe_operator_me", map[string]any{})
	expectJSON(t, "normal", me, map[string]any{"level": "normal", "levelSource": "deployment"})
	expectJSON(t, "normal can", me["can"].(map[string]any), map[string]any{"moderation": true, "brands": false, "auditLog": false})
	expectJSON(t, "managerial can", f.ok(manager, "jazebe_operator_me", map[string]any{})["can"].(map[string]any),
		map[string]any{"brands": true, "planetLifecycle": true, "auditLog": false})
	expectJSON(t, "super can", f.ok(super, "jazebe_operator_me", map[string]any{})["can"].(map[string]any),
		map[string]any{"auditLog": true, "operatorLevels": true, "brands": true})
	f.fails(f.users["fan"], "jazebe_operator_me", map[string]any{}, 7)

	f.fails(manager, "jazebe_operator_directory", map[string]any{}, 7)
	dir := f.ok(super, "jazebe_operator_directory", map[string]any{})["items"].([]any)
	if len(dir) != 4 {
		t.Fatalf("directory %d", len(dir))
	}
	levels := map[string]string{}
	for _, it := range dir {
		m := it.(map[string]any)
		levels[m["userId"].(string)] = m["level"].(string)
	}
	if levels[super] != "super" || levels[manager] != "managerial" || levels[base] != "normal" {
		t.Errorf("directory levels %v", levels)
	}

	// Audit explorer over the fixture's brand.create / planet.create rows.
	f.fails(manager, "jazebe_operator_audit_search", map[string]any{}, 7)
	search := func(filters map[string]any) map[string]any {
		return f.ok(super, "jazebe_operator_audit_search", filters)
	}
	byBrand := search(map[string]any{"brandId": f.brandID})
	if byBrand["total"].(float64) < 2 {
		t.Fatalf("brand-scoped audit rows %v", byBrand["total"])
	}
	prefix := search(map[string]any{"brandId": f.brandID, "action": "brand."})["items"].([]any)
	if len(prefix) != 1 || prefix[0].(map[string]any)["action"] != "brand.create" {
		t.Errorf("prefix search %v", prefix)
	}
	row := prefix[0].(map[string]any)
	expectJSON(t, "row", row, map[string]any{"actorType": "operator", "actorId": f.operator, "brandId": f.brandID, "result": "success", "actorName": "کاربر operator"})
	if row["brand"] == nil {
		t.Error("brand name not resolved")
	}
	if n := len(search(map[string]any{"brandId": f.brandID, "action": "planet"})["items"].([]any)); n != 1 {
		t.Errorf("dotted-family search found %d", n)
	}
	if n := len(search(map[string]any{"brandId": f.brandID, "action": "plan*"})["items"].([]any)); n != 1 {
		t.Errorf("star prefix found %d", n)
	}
	if n := len(search(map[string]any{"brandId": f.brandID, "action": "brand_"})["items"].([]any)); n != 0 {
		t.Errorf("LIKE wildcard not escaped: %d", n)
	}
	if n := len(search(map[string]any{"actorId": f.operator, "actorType": "user"})["items"].([]any)); n != 0 {
		t.Errorf("actor type filter ignored: %d", n)
	}
	future := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	if n := search(map[string]any{"brandId": f.brandID, "from": future})["total"].(float64); n != 0 {
		t.Errorf("time range ignored: %v", n)
	}
	f.fails(super, "jazebe_operator_audit_search", map[string]any{"from": "yesterday"}, 3)
	page := search(map[string]any{"brandId": f.brandID, "limit": 1, "page": 2})
	if len(page["items"].([]any)) != 1 {
		t.Errorf("paging %v", page)
	}
}
