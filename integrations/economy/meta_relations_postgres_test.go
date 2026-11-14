package main

import (
	"net/http"
	"testing"
)

func newRelationHarness(t *testing.T) *harness {
	h := newEventHarness(t)
	for _, q := range []string{
		`TRUNCATE economy_prize_group_member, economy_prize_group CASCADE`,
		`INSERT INTO economy_event_type (id, key, name, is_system, value_min, rate_limit) VALUES (gen_random_uuid(), 'MetaRelation:Completed', '{"fa":"رابطه"}', true, 0, 1000)`,
		`UPDATE economy_event_settings SET relation_chain_depth=8, relation_prize_missing='complete'`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for name, fn := range h.s.metaRelationRPCs() {
		h.rpcs[name] = fn
	}
	return h
}

func (h *harness) meta(key string, def float64) string {
	h.t.Helper()
	return h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": key, "name": map[string]string{"fa": key, "en": key}, "defaultValue": def})["id"].(string)
}

func (h *harness) relation(key, metaID, op string, value float64, targets []map[string]any, extra map[string]any) map[string]any {
	h.t.Helper()
	body := map[string]any{"key": key, "name": map[string]string{"fa": key, "en": key}, "metaId": metaID, "operator": op, "value": value, "targets": targets, "isActive": true}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/meta-relations", body)
}

func (h *harness) metaValue(user, key string) float64 {
	h.t.Helper()
	var v float64
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT COALESCE(u.value, m.default_value)::float8 FROM economy_meta m LEFT JOIN economy_user_meta u ON u.meta_id=m.id AND u.user_id=$1 WHERE m.key=$2`, user, key).Scan(&v); err != nil {
		h.t.Fatal(err)
	}
	return v
}

func (h *harness) setMeta(user, metaID, action string, value float64) map[string]any {
	h.t.Helper()
	return h.mustAPI(200, http.MethodPost, "/users/"+user+"/metas", map[string]any{"metaId": metaID, "action": action, "value": value})
}

func TestMetaRelationValidation(t *testing.T) {
	h := newRelationHarness(t)
	xp := h.meta("xp", 0)
	gold := h.meta("gold", 0)
	bad := func(body map[string]any, code, field string) {
		t.Helper()
		base := map[string]any{"key": "r", "name": map[string]string{"fa": "ر"}, "metaId": xp, "operator": "gte", "value": 10, "targets": []map[string]any{{"metaId": gold, "action": "add", "amount": 1}}}
		for k, v := range body {
			base[k] = v
		}
		c, out := h.api(http.MethodPost, "/meta-relations", base)
		if c != 400 || out["error"] != code || (field != "" && out["field"] != field) {
			t.Fatalf("%s: got %d %v", code, c, out)
		}
	}
	bad(map[string]any{"operator": "more"}, "INVALID_OPERATOR", "operator")
	bad(map[string]any{"value": nil}, "INVALID_VALUE", "value")
	bad(map[string]any{"targets": []any{}, "prizeType": "none"}, "RELATION_NEEDS_RESULT", "targets")
	bad(map[string]any{"targets": []map[string]any{{"metaId": gold, "action": "mul", "amount": 1}}}, "INVALID_META_ACTION", "targets[0].action")
	bad(map[string]any{"targets": []map[string]any{{"metaId": gold, "action": "add", "amount": 0}}}, "ZERO_AMOUNT", "targets[0].amount")
	bad(map[string]any{"targets": []map[string]any{{"metaId": gold, "amount": 1}, {"metaId": gold, "action": "add", "amount": 2}}}, "DUPLICATE_TARGET", "targets[1].metaId")
	bad(map[string]any{"targets": []map[string]any{{"metaId": "00000000-0000-0000-0000-000000000009", "amount": 1}}}, "META_NOT_FOUND", "targets[0].metaId")
	bad(map[string]any{"prizeType": "prize"}, "REQUIRED", "prizeId")
	bad(map[string]any{"repeat": "always"}, "INVALID_REPEAT", "repeat")
	bad(map[string]any{"repeat": "every", "maxFires": 0}, "INVALID_MAX_FIRES", "maxFires")
	bad(map[string]any{"startsAt": "2026-12-02T00:00:00Z", "endsAt": "2026-12-01T00:00:00Z"}, "INVALID_WINDOW", "endsAt")
	r := h.relation("ok", xp, "gte", 10, []map[string]any{{"metaId": gold, "action": "add", "amount": 5}}, nil)
	if r["repeat"] != "once" || r["grantMode"] != "auto" || r["prizeType"] != "none" || r["clientVisible"] != true {
		t.Fatalf("defaults = %v", r)
	}
	list := h.mustAPI(200, http.MethodGet, "/meta-relations?targetMetaId="+gold, nil)
	if list["total"].(float64) != 1 {
		t.Fatalf("target filter = %v", list)
	}
	if c, out := h.api(http.MethodDelete, "/metas/"+gold, nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("a target meta cannot be deleted = %d %v", c, out)
	}
	if c, out := h.api(http.MethodPut, "/meta-relations/"+r["id"].(string), map[string]any{"key": "renamed", "name": map[string]string{"fa": "ر"}, "metaId": xp, "operator": "gte", "value": 1, "targets": r["targets"]}); c != 400 || out["error"] != "KEY_IMMUTABLE" {
		t.Fatalf("key change = %d %v", c, out)
	}
}

func TestMetaRelationMultiTargetOnceAndComparators(t *testing.T) {
	h := newRelationHarness(t)
	typeID := h.eventType("round_finished", true, nil)
	xp := h.meta("xp", 0)
	gold := h.meta("gold", 10)
	rank := h.meta("rank", 0)
	exact := h.meta("exact", 0)
	h.mustAPI(201, http.MethodPost, "/meta-rules", map[string]any{"key": "xp_rule", "name": map[string]string{"fa": "x"}, "eventTypeId": typeID, "metaId": xp, "action": "add", "valueSource": "event", "amount": 1, "isActive": true})
	h.relation("xp100", xp, "gte", 100, []map[string]any{{"metaId": gold, "action": "add", "amount": 50}, {"metaId": rank, "action": "set", "amount": 2}}, nil)
	h.relation("xp_eq_42", xp, "eq", 42, []map[string]any{{"metaId": exact, "action": "add", "amount": 1}}, nil)
	h.relation("xp_lt_0", xp, "lt", 0, []map[string]any{{"metaId": exact, "action": "set", "amount": -1}}, nil)
	h.relation("inactive", xp, "gte", 0, []map[string]any{{"metaId": exact, "action": "add", "amount": 1000}}, map[string]any{"isActive": false})
	user := h.user()
	h.record(user, "round_finished", 42)
	if h.metaValue(user, "exact") != 1 || h.metaValue(user, "gold") != 10 {
		t.Fatalf("eq 42 fires, gte 100 waits: exact %v gold %v", h.metaValue(user, "exact"), h.metaValue(user, "gold"))
	}
	out := h.record(user, "round_finished", 60)
	if h.metaValue(user, "gold") != 60 || h.metaValue(user, "rank") != 2 {
		t.Fatalf("102 xp: gold default 10 + 50, rank set 2: gold %v rank %v", h.metaValue(user, "gold"), h.metaValue(user, "rank"))
	}
	rels := out["relations"].([]any)
	if len(rels) != 1 || rels[0].(map[string]any)["key"] != "xp100" || out["metas"].(map[string]any)["gold"].(float64) != 60 {
		t.Fatalf("event outcome lists the relation and target metas: %v", out)
	}
	h.record(user, "round_finished", 50)
	if h.metaValue(user, "gold") != 60 || h.metaValue(user, "exact") != 1 {
		t.Fatalf("once relations never fire again: gold %v exact %v", h.metaValue(user, "gold"), h.metaValue(user, "exact"))
	}
	if n := h.count(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='MetaRelation:Completed' AND e.user_id=$1 AND e.metadata->>'key' IN ('xp100','xp_eq_42')`, user); n != 2 {
		t.Fatalf("completed events = %d", n)
	}
	h.setMeta(user, xp, "set", -5)
	if h.metaValue(user, "exact") != -1 {
		t.Fatalf("lt 0 fires on an admin set: %v", h.metaValue(user, "exact"))
	}
}

func TestMetaRelationRearmEveryAndMaxFires(t *testing.T) {
	h := newRelationHarness(t)
	xp := h.meta("xp", 0)
	level := h.meta("level", 1)
	coins := h.meta("coins", 0)
	h.relation("level_up", xp, "gte", 100, []map[string]any{{"metaId": xp, "action": "add", "amount": -100}, {"metaId": level, "action": "add", "amount": 1}}, map[string]any{"repeat": "rearm"})
	h.relation("bonus", level, "gte", 2, []map[string]any{{"metaId": coins, "action": "add", "amount": 10}}, map[string]any{"repeat": "every", "maxFires": 2})
	user := h.user()
	h.setMeta(user, xp, "add", 130)
	if h.metaValue(user, "xp") != 30 || h.metaValue(user, "level") != 2 || h.metaValue(user, "coins") != 10 {
		t.Fatalf("130 xp -> level 2, 30 left, bonus once: xp %v level %v coins %v", h.metaValue(user, "xp"), h.metaValue(user, "level"), h.metaValue(user, "coins"))
	}
	h.setMeta(user, xp, "add", 80)
	if h.metaValue(user, "level") != 3 || h.metaValue(user, "xp") != 10 || h.metaValue(user, "coins") != 20 {
		t.Fatalf("re-armed after dropping under 100: level %v xp %v coins %v", h.metaValue(user, "level"), h.metaValue(user, "xp"), h.metaValue(user, "coins"))
	}
	h.setMeta(user, xp, "add", 95)
	if h.metaValue(user, "level") != 4 || h.metaValue(user, "coins") != 20 {
		t.Fatalf("every relation stops at max fires 2: level %v coins %v", h.metaValue(user, "level"), h.metaValue(user, "coins"))
	}
	if n := h.count(`SELECT fires FROM economy_user_meta_relation s JOIN economy_meta_relation r ON r.id=s.relation_id WHERE r.key='level_up' AND s.user_id=$1`, user); n != 3 {
		t.Fatalf("level_up fires = %d", n)
	}
}

func TestMetaRelationChainDepthAndLoops(t *testing.T) {
	h := newRelationHarness(t)
	a := h.meta("a", 0)
	b := h.meta("b", 0)
	c := h.meta("c", 0)
	d := h.meta("d", 0)
	h.relation("a_to_b", a, "gte", 1, []map[string]any{{"metaId": b, "action": "add", "amount": 1}}, map[string]any{"repeat": "every"})
	h.relation("b_to_c", b, "gte", 1, []map[string]any{{"metaId": c, "action": "add", "amount": 1}}, map[string]any{"repeat": "every"})
	h.relation("c_to_d", c, "gte", 1, []map[string]any{{"metaId": d, "action": "add", "amount": 1}}, map[string]any{"repeat": "every"})
	h.relation("d_to_a", d, "gte", 1, []map[string]any{{"metaId": a, "action": "add", "amount": 1}}, map[string]any{"repeat": "every"})
	user := h.user()
	h.setMeta(user, a, "add", 1)
	if h.metaValue(user, "a") != 2 || h.metaValue(user, "b") != 1 || h.metaValue(user, "c") != 1 || h.metaValue(user, "d") != 1 {
		t.Fatalf("a loop runs each relation once per change: a %v b %v c %v d %v", h.metaValue(user, "a"), h.metaValue(user, "b"), h.metaValue(user, "c"), h.metaValue(user, "d"))
	}
	h.mustAPI(200, http.MethodPut, "/event-settings", map[string]any{"relationChainDepth": 1})
	other := h.user()
	h.setMeta(other, a, "add", 1)
	if h.metaValue(other, "b") != 1 || h.metaValue(other, "c") != 1 || h.metaValue(other, "d") != 0 {
		t.Fatalf("chain depth 1 stops after two relations: b %v c %v d %v", h.metaValue(other, "b"), h.metaValue(other, "c"), h.metaValue(other, "d"))
	}
	h.mustAPI(200, http.MethodPut, "/event-settings", map[string]any{"relationChainDepth": 0})
	third := h.user()
	h.setMeta(third, a, "add", 1)
	if h.metaValue(third, "b") != 1 || h.metaValue(third, "c") != 0 {
		t.Fatalf("chain depth 0 fires only the direct relation: b %v c %v", h.metaValue(third, "b"), h.metaValue(third, "c"))
	}
	if c, out := h.api(http.MethodPut, "/event-settings", map[string]any{"relationChainDepth": 30}); c != 400 || out["error"] != "INVALID_CHAIN_DEPTH" {
		t.Fatalf("depth bound = %d %v", c, out)
	}
	st := h.mustAPI(200, http.MethodGet, "/event-settings", nil)
	if st["relationChainDepth"].(float64) != 0 || st["relationPrizeMissing"] != "complete" || st["timezone"] != "Asia/Tehran" {
		t.Fatalf("partial settings update keeps other fields: %v", st)
	}
}

func TestMetaRelationPrizesAutoClaimAndGroups(t *testing.T) {
	h := newRelationHarness(t)
	xp := h.meta("xp", 0)
	auto := h.coinPrize("rel_auto", 20, true)
	claim := h.coinPrize("rel_claim", 30, true)
	h.relation("auto_prize", xp, "gte", 10, nil, map[string]any{"prizeType": "prize", "prizeId": auto, "grantMode": "auto"})
	h.relation("claim_prize", xp, "gte", 20, nil, map[string]any{"prizeType": "prize", "prizeId": claim, "grantMode": "claim", "description": map[string]string{"fa": "جایزه"}})
	user := h.user()
	res := h.setMeta(user, xp, "add", 15)
	if h.wallet(user)["coin"] != 20 || len(res["relations"].([]any)) != 1 {
		t.Fatalf("auto prize granted at 15 xp: %v %v", h.wallet(user), res)
	}
	h.setMeta(user, xp, "add", 10)
	if h.wallet(user)["coin"] != 20 {
		t.Fatalf("claim prize waits: %v", h.wallet(user))
	}
	view := h.mustRPC(user, "economy_meta_relations", map[string]any{"lang": "fa"})
	ready := view["ready"].([]any)
	if len(ready) != 1 || ready[0].(map[string]any)["kind"] != "relation" || len(view["items"].([]any)) != 2 {
		t.Fatalf("relations view = %v", view)
	}
	item := view["items"].([]any)[1].(map[string]any)
	if item["key"] != "claim_prize" || item["done"] != true || item["claims"].([]any)[0].(map[string]any)["status"] != "claimable" || item["condition"].(map[string]any)["holds"] != true {
		t.Fatalf("claim item = %v", item)
	}
	mine := h.mustRPC(user, "economy_rewards_mine", map[string]any{"status": "ready"})["items"].([]any)
	if len(mine) != 1 || mine[0].(map[string]any)["source"] != "claim_prize" {
		t.Fatalf("rewards_mine lists the relation claim: %v", mine)
	}
	cid := ready[0].(map[string]any)["claimId"]
	h.mustRPC(user, "economy_reward_claim", map[string]any{"claimId": cid})
	if h.wallet(user)["coin"] != 50 {
		t.Fatalf("claimed: %v", h.wallet(user))
	}
	if c := h.rpcCode(user, "economy_reward_claim", map[string]any{"claimId": cid}); c != "ALREADY_CLAIMED" {
		t.Fatalf("second claim = %s", c)
	}
	if c := h.rpcCode(h.user(), "economy_reward_claim", map[string]any{"claimId": cid}); c != "CLAIM_NOT_FOUND" {
		t.Fatalf("other player = %s", c)
	}
	done := h.mustRPC(user, "economy_meta_relations", map[string]any{"completedOnly": true})["items"].([]any)
	if len(done) != 2 {
		t.Fatalf("completed only = %v", done)
	}

	g10 := h.coinPrize("g10", 10, true)
	group := h.mustAPI(201, http.MethodPost, "/prize-groups", map[string]any{"key": "grp", "name": map[string]string{"fa": "گروه"}, "members": []any{map[string]any{"prizeId": g10, "weight": 1}}})
	gold := h.meta("gold", 0)
	h.relation("group_prize", gold, "gte", 1, nil, map[string]any{"prizeType": "prizeGroup", "prizeGroupId": group["id"], "grantMode": "auto"})
	h.setMeta(user, gold, "add", 1)
	if h.wallet(user)["coin"] != 60 {
		t.Fatalf("group prize drawn: %v", h.wallet(user))
	}
	if c, out := h.api(http.MethodDelete, "/prize-groups/"+group["id"].(string), nil); c != 409 || out["error"] != "IN_USE" {
		t.Fatalf("used group delete = %d %v", c, out)
	}
	rels := h.mustAPI(200, http.MethodGet, "/meta-relations?metaId="+xp, nil)["items"].([]any)
	for _, r := range rels {
		row := r.(map[string]any)
		if row["players"].(float64) != 1 || row["claims"].(float64) != 1 {
			t.Fatalf("admin counts = %v", row)
		}
		if c, out := h.api(http.MethodDelete, "/meta-relations/"+row["id"].(string), nil); c != 409 || out["error"] != "IN_USE" {
			t.Fatalf("relation with claims delete = %d %v", c, out)
		}
	}
	pr := h.mustAPI(200, http.MethodGet, "/users/"+user+"/meta-relations", nil)["items"].([]any)
	if len(pr) != 3 {
		t.Fatalf("player relations = %v", pr)
	}
	h.mustAPI(200, http.MethodDelete, "/users/"+user+"/meta-relations/"+rels[0].(map[string]any)["id"].(string), nil)
	if n := len(h.mustAPI(200, http.MethodGet, "/users/"+user+"/meta-relations", nil)["items"].([]any)); n != 2 {
		t.Fatalf("after reset = %d", n)
	}
}

func TestMetaRelationFromPurchaseGrantAndEventHook(t *testing.T) {
	h := newRelationHarness(t)
	xp := h.meta("xp", 0)
	badge := h.meta("badge", 0)
	xpPrize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "xp_pack", "name": map[string]string{"fa": "تجربه"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "meta", "key": "xp", "action": "add", "value": 500}}})["id"].(string)
	rel := h.relation("vip", xp, "gte", 500, []map[string]any{{"metaId": badge, "action": "set", "amount": 1}}, nil)
	h.mustAPI(201, http.MethodPost, "/event-rules", map[string]any{"key": "on_relation", "name": map[string]string{"fa": "x"}, "eventTypeId": h.typeIDByKey("MetaRelation:Completed"),
		"measure": "count", "period": "total", "threshold": 1, "repeat": "once", "prizeId": h.coinPrize("rel_event", 3, true), "grantMode": "auto", "isActive": true})
	user := h.user()
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/grant", map[string]any{"prizeId": xpPrize})
	if h.metaValue(user, "badge") != 1 {
		t.Fatalf("an admin prize grant runs relations: %v", h.metaValue(user, "badge"))
	}
	if h.wallet(user)["coin"] != 3 {
		t.Fatalf("MetaRelation:Completed feeds event rules: %v", h.wallet(user))
	}
	if n := h.count(`SELECT count(*) FROM economy_player_event e WHERE e.user_id=$1 AND e.metadata->>'entityId'=$2`, user, rel["id"]); n != 1 {
		t.Fatalf("event entity id = %d", n)
	}
	if c, out := h.api(http.MethodPost, "/users/"+user+"/metas", map[string]any{"metaId": xp, "action": "mul", "value": 1}); c != 400 || out["error"] != "INVALID_META_ACTION" {
		t.Fatalf("bad admin action = %d %v", c, out)
	}
}

func TestMetaRelationWindowAudienceAndHidden(t *testing.T) {
	h := newRelationHarness(t)
	xp := h.meta("xp", 0)
	out := h.meta("out", 0)
	h.relation("past", xp, "gte", 1, []map[string]any{{"metaId": out, "action": "add", "amount": 1}}, map[string]any{"startsAt": "2020-01-01T00:00:00Z", "endsAt": "2020-02-01T00:00:00Z"})
	h.relation("hidden", xp, "gte", 1, []map[string]any{{"metaId": out, "action": "add", "amount": 10}}, map[string]any{"clientVisible": false})
	user := h.user()
	h.setMeta(user, xp, "add", 1)
	if h.metaValue(user, "out") != 10 {
		t.Fatalf("ended relations do not fire, hidden ones do: %v", h.metaValue(user, "out"))
	}
	items := h.mustRPC(user, "economy_meta_relations", map[string]any{})["items"].([]any)
	if len(items) != 0 {
		t.Fatalf("hidden and ended relations are not listed: %v", items)
	}
}
