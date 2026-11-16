package main

import (
	"net/http"
	"testing"
	"time"
)

func (h *harness) tagExpiry(user, tagID string) (bool, *time.Time) {
	h.t.Helper()
	var exp *time.Time
	err := h.s.db.QueryRowContext(h.ctx, `SELECT expires_at FROM base_user_tag WHERE user_id=$1 AND tag_id=$2`, user, tagID).Scan(&exp)
	if err != nil {
		return false, nil
	}
	return true, exp
}

func TestTagPrizeGrantsTimedTag(t *testing.T) {
	h, _ := newItemHarness(t)
	vip := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})["id"].(string)
	h.taxonomy("base_tag", "product", "hot", map[string]string{"fa": "داغ"})
	bad := func(items []map[string]any, code string) {
		t.Helper()
		c, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "items": items})
		if c != 400 || out["error"] != code {
			t.Fatalf("%v = %d %v, want %s", items, c, out, code)
		}
	}
	bad([]map[string]any{{"type": "tag", "key": "hot", "action": "give", "value": 30}}, "TAG_NOT_FOUND")
	bad([]map[string]any{{"type": "tag", "key": "vip", "action": "add", "value": 30}}, "INVALID_ACTION")
	bad([]map[string]any{{"type": "tag", "key": "vip", "action": "give", "value": -1}}, "INVALID_VALUE")
	give := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "vip30", "name": map[string]string{"fa": "ویژه"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "tag", "key": "vip", "action": "give", "value": 30}}})["id"].(string)
	set := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "vip1", "name": map[string]string{"fa": "ویژه"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "tag", "key": "vip", "action": "set", "value": 1}}})["id"].(string)
	forever := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "vipf", "name": map[string]string{"fa": "ویژه"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "tag", "key": "vip", "action": "give", "value": 0}}})["id"].(string)
	u := h.user()
	grant := func(prize, src string) {
		t.Helper()
		if _, err := h.s.grant(h.ctx, grantRequest{UserID: u, PrizeID: prize, SourceType: "admin", SourceID: src, Reason: "test"}); err != nil {
			t.Fatal(err)
		}
	}
	grant(give, "a")
	ok, exp := h.tagExpiry(u, vip)
	if !ok || exp == nil || time.Until(*exp) < 29*24*time.Hour || time.Until(*exp) > 31*24*time.Hour {
		t.Fatalf("first grant = %v %v", ok, exp)
	}
	grant(give, "b")
	_, exp = h.tagExpiry(u, vip)
	if time.Until(*exp) < 59*24*time.Hour || time.Until(*exp) > 61*24*time.Hour {
		t.Fatalf("give must add time: %v", exp)
	}
	grant(give, "b")
	if _, again := h.tagExpiry(u, vip); !again.Equal(*exp) {
		t.Fatal("a replayed grant added time twice")
	}
	grant(set, "c")
	_, exp = h.tagExpiry(u, vip)
	if time.Until(*exp) > 25*time.Hour {
		t.Fatalf("set must start from now: %v", exp)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE base_user_tag SET expires_at=now()-interval '1 hour' WHERE user_id=$1`, u); err != nil {
		t.Fatal(err)
	}
	tags, err := h.s.userTagIDs(h.ctx, u)
	if err != nil || len(tags) != 0 {
		t.Fatalf("an expired tag still counts: %v %v", tags, err)
	}
	grant(give, "d")
	_, exp = h.tagExpiry(u, vip)
	if time.Until(*exp) < 29*24*time.Hour || time.Until(*exp) > 31*24*time.Hour {
		t.Fatalf("give after expiry must start from now: %v", exp)
	}
	grant(forever, "e")
	if _, exp = h.tagExpiry(u, vip); exp != nil {
		t.Fatalf("a forever grant must clear the end: %v", exp)
	}
	grant(give, "f")
	if _, exp = h.tagExpiry(u, vip); exp != nil {
		t.Fatalf("a forever tag must stay forever: %v", exp)
	}
}

func TestTagDiscountPrices(t *testing.T) {
	h, skin := newItemHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_tag_discount, economy_tag_discount_settings`); err != nil {
		t.Fatal(err)
	}
	sound := h.taxonomy("base_category", "item_type", "sound", map[string]string{"fa": "صدا"})["id"].(string)
	vip := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})["id"].(string)
	gold := h.taxonomy("base_tag", "user", "gold", map[string]string{"fa": "طلایی"})["id"].(string)
	pack := h.item(map[string]any{"key": "stickers_vip", "typeId": skin, "price": map[string]any{"currencyKey": "coin", "amount": 200}})
	h.item(map[string]any{"key": "goal_boom", "typeId": sound, "price": map[string]any{"currencyKey": "coin", "amount": 100}})
	h.item(map[string]any{"key": "plain", "typeId": skin, "price": map[string]any{"currencyKey": "coin", "amount": 100}})
	prize := h.coinPrize("p50", 50, true)
	h.mustAPI(200, http.MethodPut, "/payment-methods/in_app", map[string]any{"isActive": true})
	h.mustAPI(200, http.MethodPut, "/payment-methods/zibal", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true}})
	product := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "coins", "isActive": true, "name": map[string]string{"fa": "سکه"}, "prizeId": prize,
		"paymentMethods": []string{"in_app", "zibal"}, "coinPrice": map[string]any{"currencyKey": "coin", "amount": 100}, "priceToman": 20000, "discountPercent": 10})

	bad := func(body map[string]any, code string) {
		t.Helper()
		base := map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "percent": 5, "tagIds": []string{vip}, "allItems": true}
		for k, v := range body {
			base[k] = v
		}
		c, out := h.api(http.MethodPost, "/tag-discounts", base)
		if c != 400 || out["error"] != code {
			t.Fatalf("%v = %d %v, want %s", body, c, out, code)
		}
	}
	bad(map[string]any{"percent": 0}, "INVALID_PERCENT")
	bad(map[string]any{"tagIds": []string{}}, "TAGS_REQUIRED")
	bad(map[string]any{"allItems": false}, "TARGETS_REQUIRED")
	bad(map[string]any{"allItems": false, "itemKeys": []string{"nope"}}, "ITEM_NOT_FOUND")

	rule := h.mustAPI(201, http.MethodPost, "/tag-discounts", map[string]any{"key": "vip5", "name": map[string]string{"fa": "۵٪ ویژه"}, "percent": 5, "tagIds": []string{vip},
		"itemIds": []string{pack["id"].(string)}, "itemTypeIds": []string{sound}, "productIds": []string{product["id"].(string)}})
	if rule["itemKeys"].([]any)[0] != "stickers_vip" || rule["players"].(float64) != 0 {
		t.Fatalf("rule = %v", rule)
	}
	u := h.user()
	h.give(u, 1000)
	inv := h.inventory(u)
	if inv["stickers_vip"]["discountPercent"].(float64) != 0 || inv["stickers_vip"]["finalPrice"].(map[string]any)["amount"].(float64) != 200 {
		t.Fatalf("no tag, no discount: %v", inv["stickers_vip"])
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_user_tag (user_id, tag_id, expires_at) VALUES ($1,$2,now()+interval '1 day')`, u, vip); err != nil {
		t.Fatal(err)
	}
	inv = h.inventory(u)
	if inv["stickers_vip"]["discountPercent"].(float64) != 5 || inv["stickers_vip"]["finalPrice"].(map[string]any)["amount"].(float64) != 190 {
		t.Fatalf("item discount = %v", inv["stickers_vip"])
	}
	if inv["goal_boom"]["finalPrice"].(map[string]any)["amount"].(float64) != 95 || inv["plain"]["discountPercent"].(float64) != 0 {
		t.Fatalf("type discount = %v / %v", inv["goal_boom"], inv["plain"])
	}
	res := h.mustRPC(u, "economy_item_buy", map[string]any{"key": "stickers_vip", "requestId": "r1"})
	if h.wallet(u)["coin"] != 810 {
		t.Fatalf("paid = %v wallet %v", res, h.wallet(u))
	}
	var cat map[string]any
	for _, raw := range h.mustRPC(u, "economy_catalog", map[string]any{})["products"].([]any) {
		if p := raw.(map[string]any); p["key"] == "coins" {
			cat = p
		}
	}
	if cat["tagDiscountPercent"].(float64) != 5 || cat["discountPercent"].(float64) != 14 || cat["finalToman"].(float64) != 17100 || cat["coinPrice"].(map[string]any)["amount"].(float64) != 95 || cat["storeFinalToman"].(float64) != 18000 {
		t.Fatalf("catalog = %v", cat)
	}
	h.mustRPC(u, "economy_purchase_start", map[string]any{"productId": product["id"], "method": "in_app"})
	if w := h.wallet(u)["coin"]; w != 810-95+50 {
		t.Fatalf("in-app paid wrong: %d", w)
	}
	h.mustAPI(200, http.MethodPut, "/tag-discount-settings", map[string]any{"enabled": true, "combine": "best", "maxPercent": 90, "withOwnDiscount": "best"})
	for _, raw := range h.mustRPC(u, "economy_catalog", map[string]any{})["products"].([]any) {
		if p := raw.(map[string]any); p["key"] == "coins" && (p["discountPercent"].(float64) != 10 || p["finalToman"].(float64) != 18000) {
			t.Fatalf("best of own and tag = %v", p)
		}
	}
	h.mustAPI(201, http.MethodPost, "/tag-discounts", map[string]any{"key": "gold20", "name": map[string]string{"fa": "طلایی"}, "percent": 20, "tagIds": []string{gold, vip}, "allItems": true})
	h.mustAPI(200, http.MethodPut, "/tag-discount-settings", map[string]any{"enabled": true, "combine": "sum", "maxPercent": 22, "withOwnDiscount": "stack"})
	if inv = h.inventory(u); inv["goal_boom"]["discountPercent"].(float64) != 22 {
		t.Fatalf("sum with cap = %v", inv["goal_boom"])
	}
	pv := h.mustAPI(200, http.MethodGet, "/tag-discount-preview?userId="+u, nil)
	if len(pv["rules"].([]any)) != 2 || len(pv["products"].([]any)) != 1 {
		t.Fatalf("preview = %v", pv)
	}
	if c, out := h.api(http.MethodPut, "/tag-discount-settings", map[string]any{"enabled": true, "combine": "max", "maxPercent": 90, "withOwnDiscount": "stack"}); c != 400 || out["field"] != "combine" {
		t.Fatalf("bad setting = %d %v", c, out)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE base_user_tag SET expires_at=now()-interval '1 minute' WHERE user_id=$1`, u); err != nil {
		t.Fatal(err)
	}
	if inv = h.inventory(u); inv["goal_boom"]["discountPercent"].(float64) != 0 {
		t.Fatalf("an expired tag still discounts: %v", inv["goal_boom"])
	}
	h.mustAPI(200, http.MethodPut, "/tag-discount-settings", map[string]any{"enabled": false, "combine": "best", "maxPercent": 90, "withOwnDiscount": "stack"})
	if inv = h.inventory(u); inv["stickers_vip"]["discountPercent"].(float64) != 0 {
		t.Fatalf("switched off still discounts: %v", inv["stickers_vip"])
	}
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_tag_discount, economy_tag_discount_settings`); err != nil {
		t.Fatal(err)
	}
}

func TestChanceLimitsAndKind(t *testing.T) {
	h := newChanceHarness(t)
	p := h.coinPrize("c1", 1, true)
	bad := func(body map[string]any, code string) {
		t.Helper()
		base := map[string]any{"key": "bad", "name": map[string]string{"fa": "بد"}, "price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 1}, "prize": []any{row(100, p)}}
		for k, v := range body {
			base[k] = v
		}
		c, out := h.api(http.MethodPost, "/chance-games", base)
		if c != 400 || out["error"] != code {
			t.Fatalf("%v = %d %v, want %s", body, c, out, code)
		}
	}
	bad(map[string]any{"kind": "box"}, "INVALID_KIND")
	bad(map[string]any{"dailyLimit": 0}, "INVALID_LIMIT")
	bad(map[string]any{"totalLimit": -2}, "INVALID_LIMIT")
	g := h.chanceGame(map[string]any{"key": "chest", "kind": "chest", "dailyLimit": 3, "totalLimit": 4, "showOdds": false,
		"price": map[string]any{"currencyId": h.currencyID("coin"), "amount": 5}, "prize": []any{row(60, p), row(40, "")}})
	if g["kind"] != "chest" || g["dailyLimit"].(float64) != 3 || g["showOdds"] != false {
		t.Fatalf("chest = %v", g)
	}
	u := h.user()
	if _, _, err := h.nk.WalletUpdate(h.ctx, u, map[string]int64{"coin": 100}, nil, true); err != nil {
		t.Fatal(err)
	}
	view := h.mustRPC(u, "economy_chance_game", map[string]any{"id": g["id"]})["game"].(map[string]any)
	if view["kind"] != "chest" || view["playsLeftToday"].(float64) != 3 || view["prize"].([]any)[0].(map[string]any)["chance"] != nil {
		t.Fatalf("view = %v", view)
	}
	for i := 0; i < 3; i++ {
		out := h.mustRPC(u, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]})
		if out["playsLeftToday"].(float64) != float64(2-i) {
			t.Fatalf("play %d = %v", i, out)
		}
	}
	if c := h.rpcCode(u, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]}); c != "CHANCE_DAILY_LIMIT" {
		t.Fatalf("daily = %s", c)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_chance_play SET created_at=created_at-interval '2 days' WHERE user_id=$1`, u); err != nil {
		t.Fatal(err)
	}
	h.mustRPC(u, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]})
	if c := h.rpcCode(u, "economy_chance_play_with_ticket", map[string]any{"id": g["id"]}); c != "CHANCE_TOTAL_LIMIT" {
		t.Fatalf("total = %s", c)
	}
	if w := h.wallet(u)["coin"]; w > 100-20+4 || w < 80 {
		t.Fatalf("wallet after 4 opens = %d", w)
	}
	view = h.mustRPC(u, "economy_chance_game", map[string]any{"id": g["id"]})["game"].(map[string]any)
	if view["limitReached"] != true || view["playsLeft"].(float64) != 0 {
		t.Fatalf("view after limit = %v", view)
	}
}
