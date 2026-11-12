package main

import (
	"net/http"
	"sync"
	"testing"
)

func newItemHarness(t *testing.T) (*harness, string) {
	h := newEventHarness(t)
	for name, fn := range h.s.itemRPCs() {
		h.rpcs[name] = fn
	}
	for _, q := range []string{
		`INSERT INTO economy_event_type (id, key, name, is_system) VALUES (gen_random_uuid(), 'item_bought', '{"fa":"خرید آیتم"}', true), (gen_random_uuid(), 'item_upgraded', '{"fa":"ارتقا"}', true), (gen_random_uuid(), 'purchase_completed', '{"fa":"خرید"}', true)`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	skin := h.taxonomy("base_category", "item_type", "skin", map[string]string{"fa": "پوسته", "en": "Skin"})["id"].(string)
	return h, skin
}

func (h *harness) item(body map[string]any) map[string]any {
	h.t.Helper()
	if body["name"] == nil {
		body["name"] = map[string]string{"fa": "آیتم " + body["key"].(string), "en": body["key"].(string)}
	}
	return h.mustAPI(201, http.MethodPost, "/items", body)
}

func (h *harness) give(user string, coins int64) {
	h.t.Helper()
	if _, _, err := h.nk.WalletUpdate(h.ctx, user, map[string]int64{"coin": coins}, map[string]any{"reason": "test"}, true); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) inventory(user string) map[string]map[string]any {
	h.t.Helper()
	out := map[string]map[string]any{}
	for _, raw := range h.mustRPC(user, "economy_inventory", map[string]any{"lang": "en"})["items"].([]any) {
		it := raw.(map[string]any)
		out[it["key"].(string)] = it
	}
	return out
}

func TestItemCatalogValidation(t *testing.T) {
	h, skin := newItemHarness(t)
	product := h.taxonomy("base_category", "product", "packs", map[string]string{"fa": "بسته"})["id"].(string)
	if code, out := h.api(http.MethodPost, "/items", map[string]any{"key": "x1", "typeId": product, "name": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "CATEGORY_KIND_MISMATCH" || out["field"] != "typeId" {
		t.Fatalf("type kind = %d %v", code, out)
	}
	a := h.item(map[string]any{"key": "ship_a", "typeId": skin, "rarity": "rare", "price": map[string]any{"currencyKey": "coin", "amount": 10},
		"upgrades": []map[string]any{{"currencyKey": "coin", "amount": 5}}, "services": []map[string]any{{"key": "repair", "currencyKey": "coin", "unitPrice": 2, "maxUnits": 3}}})
	if a["typeKey"] != "skin" || a["rarity"] != "rare" || a["isActive"] != true {
		t.Fatalf("item = %v", a)
	}
	if code, out := h.api(http.MethodPut, "/items/"+a["id"].(string), map[string]any{"key": "ship_b", "typeId": skin, "name": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "KEY_IMMUTABLE" {
		t.Fatalf("key change = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/items", map[string]any{"key": "self", "typeId": skin, "name": map[string]string{"fa": "x"}, "requiresKey": "self"}); code != 400 || out["error"] != "SELF_REFERENCE" {
		t.Fatalf("self = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/items", map[string]any{"key": "bad_r", "typeId": skin, "name": map[string]string{"fa": "x"}, "rarity": "mythic"}); code != 400 || out["error"] != "INVALID_RARITY" {
		t.Fatalf("rarity = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/items", map[string]any{"key": "bad_p", "typeId": skin, "name": map[string]string{"fa": "x"}, "price": map[string]any{"currencyKey": "gem", "amount": 1}}); code != 400 || out["error"] != "CURRENCY_NOT_FOUND" {
		t.Fatalf("price currency = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/items", map[string]any{"key": "bad_s", "typeId": skin, "name": map[string]string{"fa": "x"}, "isStarter": true, "unlockLevel": 2}); code != 400 || out["error"] != "STARTER_HAS_LEVEL" {
		t.Fatalf("starter level = %d %v", code, out)
	}
	h.item(map[string]any{"key": "ship_c", "typeId": skin, "requiresKey": "ship_a"})
	if code, out := h.api(http.MethodDelete, "/items/"+a["id"].(string), nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete required item = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "pz", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "item", "key": "nope", "action": "give"}}}); code != 400 || out["error"] != "ITEM_NOT_FOUND" {
		t.Fatalf("prize unknown item = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "pz", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "item", "key": "ship_a", "action": "add"}}}); code != 400 || out["error"] != "INVALID_ACTION" {
		t.Fatalf("prize item action = %d %v", code, out)
	}
	list := h.mustAPI(200, http.MethodGet, "/items?typeId="+skin+"&q=ship", nil)
	if list["total"].(float64) != 2 {
		t.Fatalf("list = %v", list)
	}
}

func TestItemStarterBuyAndIdempotency(t *testing.T) {
	h, skin := newItemHarness(t)
	h.item(map[string]any{"key": "ship_basic", "typeId": skin, "isStarter": true})
	h.item(map[string]any{"key": "ship_gold", "typeId": skin, "price": map[string]any{"currencyKey": "coin", "amount": 30}})
	h.item(map[string]any{"key": "ship_neon", "typeId": skin, "requiresKey": "ship_gold", "price": map[string]any{"currencyKey": "coin", "amount": 10}})
	user := h.user()
	inv := h.inventory(user)
	if inv["ship_basic"]["owned"] != true || inv["ship_basic"]["equipped"] != true || inv["ship_gold"]["owned"] != false || inv["ship_gold"]["canBuy"] != true || inv["ship_neon"]["canBuy"] != false {
		t.Fatalf("starter view = %v", inv)
	}
	h.inventory(user)
	if n := h.count(`SELECT count(*) FROM economy_user_item WHERE user_id=$1`, user); n != 1 {
		t.Fatalf("starter rows = %d", n)
	}
	if c := h.rpcCode(user, "economy_item_buy", map[string]any{"key": "ship_gold"}); c != "INSUFFICIENT_BALANCE" {
		t.Fatalf("poor buy = %s", c)
	}
	if n := h.count(`SELECT count(*) FROM economy_user_item WHERE user_id=$1`, user); n != 1 {
		t.Fatalf("failed buy kept a row: %d", n)
	}
	if c := h.rpcCode(user, "economy_item_buy", map[string]any{"key": "ship_neon"}); c != "REQUIRES_ITEM" {
		t.Fatalf("requires = %s", c)
	}
	h.give(user, 100)
	var wg sync.WaitGroup
	results := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.rpc(user, "economy_item_buy", map[string]any{"key": "ship_gold"})
			if err == nil {
				results <- "ok"
				return
			}
			app, _ := asApp(err)
			if app != nil {
				results <- app.Code
			} else {
				results <- err.Error()
			}
		}()
	}
	wg.Wait()
	close(results)
	oks := 0
	for r := range results {
		if r == "ok" {
			oks++
		} else if r != "ALREADY_OWNED" {
			t.Fatalf("parallel buy error %s", r)
		}
	}
	if oks != 1 || h.wallet(user)["coin"] != 70 {
		t.Fatalf("parallel buys: %d ok, wallet %v", oks, h.wallet(user))
	}
	first := h.mustRPC(user, "economy_item_buy", map[string]any{"key": "ship_neon", "requestId": "r-1"})
	again := h.mustRPC(user, "economy_item_buy", map[string]any{"key": "ship_neon", "requestId": "r-1"})
	if first["item"].(map[string]any)["owned"] != true || again["replayed"] != true || h.wallet(user)["coin"] != 60 {
		t.Fatalf("replay = %v %v wallet %v", first, again, h.wallet(user))
	}
	if c := h.rpcCode(user, "economy_item_buy", map[string]any{"key": "ship_gold", "requestId": "r-1"}); c != "REQUEST_ID_REUSED" {
		t.Fatalf("reuse = %s", c)
	}
	if n := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'reason'='item_buy'`, user); n != 2 {
		t.Fatalf("buy ledger rows = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE e.user_id=$1 AND t.key='item_bought'`, user); n != 2 {
		t.Fatalf("item_bought events = %d", n)
	}
}

func TestItemUnlockUpgradeEquipUse(t *testing.T) {
	h, skin := newItemHarness(t)
	xp := h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "xp", "name": map[string]string{"fa": "تجربه"}})
	h.mustAPI(201, http.MethodPost, "/level-tables", map[string]any{"key": "lv", "name": map[string]string{"fa": "سطح"}, "metaId": xp["id"], "isActive": true,
		"steps": []map[string]any{{"level": 1, "threshold": 0}, {"level": 2, "threshold": 100}}})
	h.item(map[string]any{"key": "ship_lv2", "typeId": skin, "unlockLevel": 2})
	h.item(map[string]any{"key": "ship_up", "typeId": skin, "price": map[string]any{"currencyKey": "coin", "amount": 1},
		"upgrades": []map[string]any{{"currencyKey": "coin", "amount": 5}, {"currencyKey": "coin", "amount": 7}},
		"services": []map[string]any{{"key": "paint", "currencyKey": "coin", "unitPrice": 3, "maxUnits": 2}}})
	user := h.user()
	h.give(user, 100)
	if c := h.rpcCode(user, "economy_item_unlock", map[string]any{"key": "ship_lv2"}); c != "LEVEL_TOO_LOW" {
		t.Fatalf("unlock low = %s", c)
	}
	if c := h.rpcCode(user, "economy_item_unlock", map[string]any{"key": "ship_up"}); c != "NOT_UNLOCKABLE" {
		t.Fatalf("unlock priced = %s", c)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ($1,$2,150)`, user, xp["id"]); err != nil {
		t.Fatal(err)
	}
	if inv := h.inventory(user); inv["ship_lv2"]["canUnlock"] != true || inv["ship_lv2"]["locked"] != false {
		t.Fatalf("unlock view = %v", inv["ship_lv2"])
	}
	h.mustRPC(user, "economy_item_unlock", map[string]any{"key": "ship_lv2"})
	if c := h.rpcCode(user, "economy_item_unlock", map[string]any{"key": "ship_lv2"}); c != "ALREADY_OWNED" {
		t.Fatalf("unlock twice = %s", c)
	}
	if c := h.rpcCode(user, "economy_item_upgrade", map[string]any{"key": "ship_up", "fromLevel": 0}); c != "NOT_OWNED" {
		t.Fatalf("upgrade not owned = %s", c)
	}
	h.mustRPC(user, "economy_item_buy", map[string]any{"key": "ship_up"})
	if c := h.rpcCode(user, "economy_item_upgrade", map[string]any{"key": "ship_up", "fromLevel": 1}); c != "UPGRADE_CHANGED" {
		t.Fatalf("upgrade changed = %s", c)
	}
	up := h.mustRPC(user, "economy_item_upgrade", map[string]any{"key": "ship_up", "fromLevel": 0, "requestId": "u1"})
	if up["level"].(float64) != 1 || h.wallet(user)["coin"] != 94 {
		t.Fatalf("upgrade = %v wallet %v", up, h.wallet(user))
	}
	h.mustRPC(user, "economy_item_upgrade", map[string]any{"key": "ship_up", "fromLevel": 0, "requestId": "u1"})
	if h.wallet(user)["coin"] != 94 {
		t.Fatalf("upgrade replay charged: %v", h.wallet(user))
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	oks := 0
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.rpc(user, "economy_item_upgrade", map[string]any{"key": "ship_up", "fromLevel": 1}); err == nil {
				mu.Lock()
				oks++
				mu.Unlock()
			} else if app, _ := asApp(err); app == nil || app.Code != "UPGRADE_CHANGED" {
				t.Errorf("parallel upgrade: %v", err)
			}
		}()
	}
	wg.Wait()
	if oks != 1 || h.wallet(user)["coin"] != 87 {
		t.Fatalf("parallel upgrades %d wallet %v", oks, h.wallet(user))
	}
	if c := h.rpcCode(user, "economy_item_upgrade", map[string]any{"key": "ship_up", "fromLevel": 2}); c != "UPGRADE_MAXED" {
		t.Fatalf("maxed = %s", c)
	}
	h.mustRPC(user, "economy_item_equip", map[string]any{"key": "ship_lv2"})
	eq := h.mustRPC(user, "economy_item_equip", map[string]any{"key": "ship_up"})
	if off := eq["unequipped"].([]any); len(off) != 1 || off[0] != "ship_lv2" || eq["slot"] != "skin" {
		t.Fatalf("equip = %v", eq)
	}
	wg = sync.WaitGroup{}
	for i := 0; i < 6; i++ {
		key := "ship_up"
		if i%2 == 0 {
			key = "ship_lv2"
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.rpc(user, "economy_item_equip", map[string]any{"key": key}); err != nil {
				t.Errorf("parallel equip: %v", err)
			}
		}()
	}
	wg.Wait()
	if n := h.count(`SELECT count(*) FROM economy_user_item WHERE user_id=$1 AND equipped`, user); n != 1 {
		t.Fatalf("equipped after parallel = %d", n)
	}
	h.mustRPC(user, "economy_item_unequip", map[string]any{"key": "ship_up"})
	h.mustRPC(user, "economy_item_unequip", map[string]any{"key": "ship_lv2"})
	if n := h.count(`SELECT count(*) FROM economy_user_item WHERE user_id=$1 AND equipped`, user); n != 0 {
		t.Fatalf("equipped after unequip = %d", n)
	}
	if c := h.rpcCode(user, "economy_item_use", map[string]any{"key": "ship_up", "service": "paint"}); c != "REQUEST_ID_REQUIRED" {
		t.Fatalf("use without id = %s", c)
	}
	if c := h.rpcCode(user, "economy_item_use", map[string]any{"key": "ship_up", "service": "paint", "units": 3, "requestId": "s0"}); c != "INVALID_UNITS" {
		t.Fatalf("units = %s", c)
	}
	h.mustRPC(user, "economy_item_use", map[string]any{"key": "ship_up", "service": "paint", "units": 2, "requestId": "s1"})
	h.mustRPC(user, "economy_item_use", map[string]any{"key": "ship_up", "service": "paint", "units": 2, "requestId": "s1"})
	if h.wallet(user)["coin"] != 81 {
		t.Fatalf("use charged twice: %v", h.wallet(user))
	}
	if c := h.rpcCode(user, "economy_item_use", map[string]any{"key": "ship_up", "service": "paint", "units": 1, "requestId": "s1"}); c != "REQUEST_ID_REUSED" {
		t.Fatalf("use reuse = %s", c)
	}
}

func TestItemPrizesAndProductRules(t *testing.T) {
	h, skin := newItemHarness(t)
	h.mustAPI(200, http.MethodPut, "/payment-methods/in_app", map[string]any{"isActive": true})
	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true, "packageName": "ir.game"}})
	xp := h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "xp", "name": map[string]string{"fa": "تجربه"}})
	h.mustAPI(201, http.MethodPost, "/level-tables", map[string]any{"key": "lv", "name": map[string]string{"fa": "سطح"}, "metaId": xp["id"], "isActive": true,
		"steps": []map[string]any{{"level": 1, "threshold": 0}, {"level": 2, "threshold": 100}}})
	h.item(map[string]any{"key": "ship_red", "typeId": skin})
	h.item(map[string]any{"key": "ship_blue", "typeId": skin})
	skinPrize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "red", "name": map[string]string{"fa": "قرمز"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "item", "key": "ship_red", "action": "give"}}})
	user := h.user()
	h.give(user, 100)
	for i := 0; i < 2; i++ {
		if _, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: skinPrize["id"].(string), SourceType: "admin", SourceID: "same", Reason: "t"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: skinPrize["id"].(string), SourceType: "admin", SourceID: "other", Reason: "t"}); err != nil {
		t.Fatal(err)
	}
	if n := h.count(`SELECT count(*) FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1 AND i.key='ship_red' AND u.source='prize'`, user); n != 1 {
		t.Fatalf("prize item rows = %d", n)
	}
	if _, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: skinPrize["id"].(string), SourceType: "admin", SourceID: "third", Reason: "t", CheckItems: true}); err == nil {
		t.Fatal("check items should refuse an owned item")
	}
	sellRed := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "sell-red", "isActive": true, "name": map[string]string{"fa": "قرمز"}, "prizeId": skinPrize["id"],
		"paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "coin", "amount": 5}})
	if sellRed["checkContents"] != true {
		t.Fatalf("check contents default = %v", sellRed)
	}
	if c := h.rpcCode(user, "economy_purchase_start", map[string]any{"productId": sellRed["id"], "method": "in_app"}); c != "PRIZE_ALREADY_OWNED" {
		t.Fatalf("owned product = %s", c)
	}
	bluePrize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "blue", "name": map[string]string{"fa": "آبی"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "item", "key": "ship_blue", "action": "give"}}})
	sellBlue := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "sell-blue", "isActive": true, "name": map[string]string{"fa": "آبی"}, "prizeId": bluePrize["id"],
		"paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "coin", "amount": 5}, "purchaseMessage": map[string]string{"fa": "مبارکه", "en": "Enjoy"}})
	catalog := func(u string) map[string]map[string]any {
		out := map[string]map[string]any{}
		for _, raw := range h.mustRPC(u, "economy_catalog", map[string]any{"lang": "en"})["products"].([]any) {
			p := raw.(map[string]any)
			out[p["key"].(string)] = p
		}
		return out
	}
	if c := catalog(user); c["sell-red"]["owned"] != true || c["sell-blue"]["owned"] != false || c["sell-blue"]["purchaseMessage"] != "Enjoy" {
		t.Fatalf("catalog owned = %v", c)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	oks := 0
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.rpc(user, "economy_purchase_start", map[string]any{"productId": sellBlue["id"], "method": "in_app"}); err == nil {
				mu.Lock()
				oks++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if oks != 1 || h.wallet(user)["coin"] != 95 {
		t.Fatalf("parallel item product: %d ok wallet %v", oks, h.wallet(user))
	}
	coins := h.coinPrize("starter_coins", 50, true)
	once := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "starter-pack", "isActive": true, "name": map[string]string{"fa": "شروع"}, "prizeId": coins,
		"paymentMethods": []string{"in_app", "bazaar"}, "coinPrice": map[string]any{"currencyKey": "coin", "amount": 1}, "priceToman": 1000, "sku": map[string]string{"bazaar": "starter"},
		"isOneTime": true, "maxLevel": 1})
	bob := h.user()
	h.give(bob, 10)
	a := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": once["id"], "method": "bazaar"})["paymentId"].(string)
	b := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": once["id"], "method": "bazaar"})["paymentId"].(string)
	wg = sync.WaitGroup{}
	for _, id := range []string{a, b} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if _, err := h.rpc(bob, "economy_purchase_submit", map[string]any{"paymentId": id, "token": "sandbox-" + id + "-t"}); err != nil {
				t.Errorf("submit %s: %v", id, err)
			}
		}(id)
	}
	wg.Wait()
	pa, pb := h.payment(a), h.payment(b)
	if pa.Status != "success" || pb.Status != "success" || pa.RewardGranted == pb.RewardGranted || (pa.Outcome != "duplicate" && pb.Outcome != "duplicate") {
		t.Fatalf("parallel one-time settlement: %+v %+v", pa, pb)
	}
	if h.wallet(bob)["coin"] != 60 {
		t.Fatalf("one-time delivered twice: %v", h.wallet(bob))
	}
	if c := h.rpcCode(bob, "economy_purchase_start", map[string]any{"productId": once["id"], "method": "in_app"}); c != "PRODUCT_ALREADY_BOUGHT" {
		t.Fatalf("one-time again = %s", c)
	}
	if _, ok := catalog(bob)["starter-pack"]; ok {
		t.Fatal("bought one-time product still listed")
	}
	dup := a
	if pb.Outcome == "duplicate" {
		dup = b
	}
	if code, out := h.api(http.MethodPost, "/payments/"+dup+"/retry-grant", nil); code != 409 || out["error"] != "PRODUCT_ALREADY_BOUGHT" {
		t.Fatalf("retry duplicate = %d %v", code, out)
	}
	mine := h.mustRPC(bob, "economy_purchases_mine", map[string]any{})["items"].([]any)
	if len(mine) != 1 || mine[0].(map[string]any)["count"].(float64) != 1 || mine[0].(map[string]any)["key"] != "starter-pack" {
		t.Fatalf("purchases mine = %v", mine)
	}
	carol := h.user()
	h.give(carol, 10)
	if _, ok := catalog(carol)["starter-pack"]; !ok {
		t.Fatal("one-time product hidden from a new player")
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ($1,$2,150)`, carol, xp["id"]); err != nil {
		t.Fatal(err)
	}
	if _, ok := catalog(carol)["starter-pack"]; ok {
		t.Fatal("over max level still listed")
	}
	if c := h.rpcCode(carol, "economy_purchase_start", map[string]any{"productId": once["id"], "method": "in_app"}); c != "LEVEL_TOO_HIGH" {
		t.Fatalf("max level = %s", c)
	}
	pro := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "pro-pack", "isActive": true, "name": map[string]string{"fa": "حرفه‌ای"}, "prizeId": coins,
		"paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "coin", "amount": 1}, "minLevel": 2})
	if c := catalog(bob); c["pro-pack"]["levelLocked"] != true {
		t.Fatalf("level locked = %v", c["pro-pack"])
	}
	if c := h.rpcCode(bob, "economy_purchase_start", map[string]any{"productId": pro["id"], "method": "in_app"}); c != "LEVEL_TOO_LOW" {
		t.Fatalf("min level = %s", c)
	}
	h.mustRPC(carol, "economy_purchase_start", map[string]any{"productId": pro["id"], "method": "in_app"})
	if code, out := h.api(http.MethodPost, "/products", map[string]any{"key": "bad-lv", "isActive": true, "name": map[string]string{"fa": "x"}, "prizeId": coins,
		"paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "coin", "amount": 1}, "minLevel": 3, "maxLevel": 2}); code != 400 || out["error"] != "INVALID_LEVEL_RANGE" {
		t.Fatalf("level range = %d %v", code, out)
	}
}

func TestItemConsolePlayerTools(t *testing.T) {
	h, skin := newItemHarness(t)
	starter := h.item(map[string]any{"key": "ship_start", "typeId": skin, "isStarter": true})
	extra := h.item(map[string]any{"key": "ship_extra", "typeId": skin})
	user := h.user()
	view := h.mustAPI(200, http.MethodGet, "/users/"+user+"/items?lang=en", nil)
	if view["counts"].(map[string]any)["owned"].(float64) != 1 {
		t.Fatalf("console view = %v", view)
	}
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/items", map[string]any{"itemId": extra["id"]})
	if code, out := h.api(http.MethodPost, "/users/"+user+"/items", map[string]any{"itemId": extra["id"]}); code != 409 || out["error"] != "ALREADY_OWNED" {
		t.Fatalf("grant twice = %d %v", code, out)
	}
	if code, out := h.api(http.MethodDelete, "/users/"+user+"/items/"+starter["id"].(string), nil); code != 409 || out["error"] != "STARTER_ITEM" {
		t.Fatalf("revoke starter = %d %v", code, out)
	}
	if code, out := h.api(http.MethodDelete, "/items/"+extra["id"].(string), nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete owned = %d %v", code, out)
	}
	h.mustAPI(200, http.MethodDelete, "/users/"+user+"/items/"+extra["id"].(string), nil)
	if n := h.count(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND kind='item'`, user); n != 3 {
		t.Fatalf("activity rows = %d", n)
	}
	h.mustAPI(200, http.MethodDelete, "/items/"+extra["id"].(string), nil)
}
