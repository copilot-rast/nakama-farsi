package main

import (
	"net/http"
	"testing"
	"time"
)

func newOfferHarness(t *testing.T) *harness {
	h := newEventHarness(t)
	for _, q := range []string{
		`TRUNCATE economy_user_offer_period, economy_user_offer_state, economy_user_offer, economy_offer CASCADE`,
		`DELETE FROM economy_offer_settings`,
		`INSERT INTO economy_event_type (id, key, name, is_system, value_min, rate_limit) VALUES (gen_random_uuid(), 'Offer:Bought', '{"fa":"خرید پیشنهاد"}', true, 0, 1000)`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for name, fn := range h.s.offerRPCs() {
		h.rpcs[name] = fn
	}
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس", "en": "Gem"}, "type": "hard"})
	h.mustAPI(200, http.MethodPut, "/payment-methods/in_app", map[string]any{"isActive": true})
	h.mustAPI(200, http.MethodPut, "/payment-methods/zibal", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true}})
	return h
}

// coinProduct makes an in-app (gem) and gateway product that gives `coins`.
func (h *harness) coinProduct(key string, gems, toman int64, extra map[string]any) string {
	h.t.Helper()
	prize := h.coinPrize("pz_"+key, 100, true)
	body := map[string]any{"key": key, "isActive": true, "name": map[string]string{"fa": "بستهٔ " + key, "en": "Pack " + key}, "prizeId": prize,
		"paymentMethods": []string{"in_app", "zibal"}, "coinPrice": map[string]any{"currencyKey": "gem", "amount": gems}, "priceToman": toman}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/products", body)["id"].(string)
}

func (h *harness) offer(path, key, productID string, extra map[string]any) string {
	h.t.Helper()
	body := map[string]any{"key": key, "name": map[string]string{"fa": "پیشنهاد " + key, "en": "Offer " + key}, "productId": productID, "discountPercent": 50}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/"+path, body)["id"].(string)
}

func (h *harness) offerSettings(change map[string]any) map[string]any {
	h.t.Helper()
	return h.mustAPI(200, http.MethodPut, "/offer-settings", change)
}

func (h *harness) gems(user string, n int64) {
	h.t.Helper()
	if _, _, err := h.nk.WalletUpdate(h.ctx, user, map[string]int64{"gem": n}, nil, false); err != nil {
		h.t.Fatal(err)
	}
}

func offerKeys(list []any) []string {
	out := []string{}
	for _, x := range list {
		out = append(out, x.(map[string]any)["key"].(string))
	}
	return out
}

func dailyOf(out map[string]any) []any {
	return out["daily"].(map[string]any)["items"].([]any)
}

func entryOf(out map[string]any) []any {
	return out["entry"].(map[string]any)["items"].([]any)
}

func TestOfferValidationAndSettings(t *testing.T) {
	h := newOfferHarness(t)
	prod := h.coinProduct("pv", 10, 20000, nil)
	bad := func(path string, body map[string]any, code, field string) {
		t.Helper()
		base := map[string]any{"key": "o1", "name": map[string]string{"fa": "x"}, "productId": prod}
		for k, v := range body {
			base[k] = v
		}
		c, out := h.api(http.MethodPost, path, base)
		if c != 400 || out["error"] != code || (field != "" && out["field"] != field) {
			t.Fatalf("%s: got %d %v", code, c, out)
		}
	}
	bad("/offers", map[string]any{"productId": ""}, "REQUIRED", "productId")
	bad("/offers", map[string]any{"discountPercent": 120}, "INVALID_DISCOUNT", "discountPercent")
	bad("/offers", map[string]any{"priceMode": "price"}, "PRICE_REQUIRED", "priceToman")
	bad("/offers", map[string]any{"priceMode": "price", "priceToman": 50}, "INVALID_PRICE", "priceToman")
	bad("/offers", map[string]any{"priceMode": "price", "coinPrice": map[string]any{"currencyKey": "zz", "amount": 3}}, "CURRENCY_NOT_FOUND", "coinPrice.currencyKey")
	bad("/offers", map[string]any{"sku": map[string]string{"bazaar": "bad sku"}}, "INVALID_SKU", "sku.bazaar")
	bad("/offers", map[string]any{"purchaseLimit": 0}, "INVALID_LIMIT", "purchaseLimit")
	bad("/offers", map[string]any{"weight": -1}, "INVALID_WEIGHT", "weight")
	bad("/offers", map[string]any{"minLevel": 5, "maxLevel": 2}, "INVALID_LEVEL", "maxLevel")
	bad("/offers", map[string]any{"metaConditions": []map[string]any{{"metaId": "00000000-0000-0000-0000-000000000009", "operator": "gte", "value": 1}}}, "META_NOT_FOUND", "metaConditions[0].metaId")
	bad("/entry-offers", map[string]any{"triggerType": "level"}, "INVALID_VALUE", "triggerValue")
	bad("/entry-offers", map[string]any{"triggerType": "tag"}, "TAG_NOT_FOUND", "triggerRef")
	bad("/entry-offers", map[string]any{"triggerType": "soon"}, "INVALID_TRIGGER", "triggerType")
	bad("/entry-offers", map[string]any{"durationMinutes": 0}, "INVALID_DURATION", "durationMinutes")
	bad("/entry-offers", map[string]any{"repeat": "cooldown", "maxTriggers": 0}, "INVALID_MAX_TRIGGERS", "maxTriggers")

	d := h.mustAPI(201, http.MethodPost, "/offers", map[string]any{"key": "daily1", "name": map[string]string{"fa": "روزانه"}, "productId": prod, "triggerType": "lapsed", "durationMinutes": 5})
	if d["kind"] != "daily" || d["weight"].(float64) != 10 || d["purchaseLimit"].(float64) != 1 || d["triggerType"] != "first_entry" || d["durationMinutes"].(float64) != 1440 || d["isActive"] != true {
		t.Fatalf("daily defaults %v", d)
	}
	e := h.mustAPI(201, http.MethodPost, "/entry-offers", map[string]any{"key": "entry1", "name": map[string]string{"fa": "ورود"}, "productId": prod, "triggerType": "lapsed", "triggerValue": 7, "repeat": "cooldown", "cooldownHours": 24})
	if e["kind"] != "entry" || e["triggerType"] != "lapsed" || e["repeat"] != "cooldown" {
		t.Fatalf("entry %v", e)
	}
	if c, _ := h.api(http.MethodGet, "/offers/"+e["id"].(string), nil); c != 404 {
		t.Fatalf("an entry offer is not a daily offer: %d", c)
	}
	if l := h.mustAPI(200, http.MethodGet, "/entry-offers", nil); l["total"].(float64) != 1 {
		t.Fatalf("entry list %v", l)
	}

	st := h.mustAPI(200, http.MethodGet, "/offer-settings", nil)["settings"].(map[string]any)
	if st["timezone"] != "Asia/Tehran" || st["count"].(float64) != 3 || st["resetHour"].(float64) != 0 || st["weighting"] != "weighted" || st["rerollLimit"].(float64) != 1 ||
		st["rerollCost"] != nil || st["entryTimerStart"] != "list" || st["inactiveProducts"] != "allow" || st["boughtCooldownDays"].(float64) != 7 {
		t.Fatalf("default settings %v", st)
	}
	for _, c := range []struct {
		body  map[string]any
		code  string
		field string
	}{
		{map[string]any{"timezone": "Mars/Base"}, "INVALID_TIMEZONE", "timezone"},
		{map[string]any{"resetHour": 24}, "INVALID_SETTING", "resetHour"},
		{map[string]any{"count": 21}, "INVALID_SETTING", "count"},
		{map[string]any{"weighting": "random"}, "INVALID_POLICY", "weighting"},
		{map[string]any{"entryTimerStart": "never"}, "INVALID_POLICY", "entryTimerStart"},
		{map[string]any{"rerollCost": map[string]any{"currencyKey": "gem", "amount": 0}}, "INVALID_SETTING", "rerollCost.amount"},
		{map[string]any{"countByTag": []map[string]any{{"tagId": "00000000-0000-0000-0000-000000000009", "count": 4}}}, "TAG_NOT_FOUND", "countByTag[0].tagId"},
	} {
		code, out := h.api(http.MethodPut, "/offer-settings", c.body)
		if code != 400 || out["error"] != c.code || out["field"] != c.field {
			t.Fatalf("%v: %d %v", c.body, code, out)
		}
	}
	saved := h.offerSettings(map[string]any{"count": 2, "rerollCost": map[string]any{"currencyKey": "gem", "amount": 5}})["settings"].(map[string]any)
	if saved["count"].(float64) != 2 || saved["timezone"] != "Asia/Tehran" || saved["rerollCost"].(map[string]any)["amount"].(float64) != 5 {
		t.Fatalf("partial update keeps the rest %v", saved)
	}
}

func TestDailyOffersPerPlayerPeriodAndPurchase(t *testing.T) {
	h := newOfferHarness(t)
	teh, _ := time.LoadLocation("Asia/Tehran")
	day1 := time.Date(2026, 11, 14, 10, 0, 0, 0, teh)
	h.at(day1)
	prods := map[string]string{}
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		prods[k] = h.coinProduct("p"+k, 10, 20000, nil)
		h.offer("offers", "o"+k, prods[k], nil)
	}
	h.offer("offers", "never", h.coinProduct("pz", 10, 20000, nil), map[string]any{"weight": 0})
	h.offer("offers", "off", h.coinProduct("poff", 10, 20000, nil), map[string]any{"isActive": false})
	h.offerSettings(map[string]any{"count": 2, "noRepeatDays": 1, "fillWithRepeats": true})
	alice, bob := h.user(), h.user()

	first := h.mustRPC(alice, "economy_offers", map[string]any{"lang": "fa"})
	items := dailyOf(first)
	if len(items) != 2 {
		t.Fatalf("two daily offers %v", first)
	}
	again := dailyOf(h.mustRPC(alice, "economy_offers", map[string]any{}))
	if k1, k2 := offerKeys(items), offerKeys(again); k1[0] != k2[0] || k1[1] != k2[1] {
		t.Fatalf("the same offers all period: %v vs %v", k1, k2)
	}
	for _, k := range offerKeys(items) {
		if k == "never" || k == "off" {
			t.Fatalf("weight 0 or inactive drawn: %v", k)
		}
	}
	it := items[0].(map[string]any)
	daily := first["daily"].(map[string]any)
	if it["discountPercent"].(float64) != 50 || it["coinPrice"].(map[string]any)["amount"].(float64) != 5 || it["finalToman"].(float64) != 10000 || it["remaining"].(float64) != 1 ||
		it["name"] != "پیشنهاد "+it["key"].(string) || it["originalToman"].(float64) != 20000 {
		t.Fatalf("item %v", it)
	}
	if left := daily["secondsLeft"].(float64); left != 14*3600 {
		t.Fatalf("countdown to Tehran midnight: %v", left)
	}
	if h.count(`SELECT count(*) FROM economy_user_offer WHERE user_id=$1 AND shown_at IS NOT NULL`, alice) != 2 {
		t.Fatal("listing marks offers shown")
	}
	h.mustRPC(bob, "economy_offers", map[string]any{})
	if h.count(`SELECT count(*) FROM economy_user_offer WHERE user_id=$1`, bob) != 2 {
		t.Fatal("bob gets his own set")
	}

	// Buy at the personal price with coins; the limit of one closes it.
	h.gems(alice, 100)
	offerID := it["offerId"].(string)
	if c := h.rpcCode(alice, "economy_offer_buy", map[string]any{"offerId": offerID, "method": "zarinpal"}); c != "METHOD_NOT_ALLOWED" {
		t.Fatalf("method %s", c)
	}
	if c := h.rpcCode(bob, "economy_offer_buy", map[string]any{"offerId": offerID, "method": "in_app"}); c != "OFFER_NOT_FOUND" {
		t.Fatalf("someone else's offer: %s", c)
	}
	res := h.mustRPC(alice, "economy_offer_buy", map[string]any{"offerId": offerID, "method": "in_app"})
	if res["status"] != "success" || h.wallet(alice)["gem"] != 95 {
		t.Fatalf("buy %v wallet %v", res, h.wallet(alice))
	}
	pay := h.payment(res["paymentId"].(string))
	if pay.Status != "success" {
		t.Fatalf("payment %+v", pay)
	}
	if h.count(`SELECT count(*) FROM economy_payment WHERE id=$1 AND user_offer_id=$2 AND coin_amount=5 AND discount_percent=50`, pay.ID, offerID) != 1 {
		t.Fatal("payment carries the offer and its price")
	}
	if c := h.rpcCode(alice, "economy_offer_buy", map[string]any{"offerId": offerID, "method": "in_app"}); c != "OFFER_SOLD_OUT" {
		t.Fatalf("second buy %s", c)
	}
	if c := h.rpcCode(alice, "economy_purchase_start", map[string]any{"offerId": offerID, "productId": prods["a"], "method": "in_app"}); c != "OFFER_SOLD_OUT" && c != "OFFER_PRODUCT_MISMATCH" {
		t.Fatalf("purchase_start with offerId %s", c)
	}
	after := dailyOf(h.mustRPC(alice, "economy_offers", map[string]any{}))
	if len(after) != 2 || after[0].(map[string]any)["soldOut"] != true {
		t.Fatalf("sold-out offers stay listed by default: %v", after)
	}
	h.offerSettings(map[string]any{"showSoldOut": false})
	if n := len(dailyOf(h.mustRPC(alice, "economy_offers", map[string]any{}))); n != 1 {
		t.Fatalf("hidden when sold out: %d", n)
	}
	if h.count(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='Offer:Bought' AND e.user_id=$1`, alice) != 1 {
		t.Fatal("Offer:Bought recorded")
	}

	// Next day: new period, yesterday's offers are not repeated while others remain.
	h.at(day1.Add(24 * time.Hour))
	yesterday := map[string]bool{}
	for _, k := range offerKeys(items) {
		yesterday[k] = true
	}
	next := offerKeys(dailyOf(h.mustRPC(alice, "economy_offers", map[string]any{})))
	if len(next) != 2 {
		t.Fatalf("next day %v", next)
	}
	for _, k := range next {
		if yesterday[k] {
			t.Fatalf("no repeat within a day: %v after %v", next, yesterday)
		}
	}
	if c := h.rpcCode(alice, "economy_offer_buy", map[string]any{"offerId": items[1].(map[string]any)["offerId"], "method": "in_app"}); c != "OFFER_EXPIRED" {
		t.Fatalf("yesterday's offer %s", c)
	}

	// Preview equals what a new player would get this period.
	carol := h.user()
	pv := h.mustAPI(200, http.MethodGet, "/offer-preview?userId="+carol, nil)
	picked := []string{}
	for _, r := range pv["draw"].([]any) {
		row := r.(map[string]any)
		if row["picked"] == true {
			picked = append(picked, row["key"].(string))
		}
		if row["key"] == "never" && row["reason"] != "weight_zero" || row["key"] == "off" && row["reason"] != "inactive" {
			t.Fatalf("reason %v", row)
		}
	}
	got := offerKeys(dailyOf(h.mustRPC(carol, "economy_offers", map[string]any{})))
	if len(picked) != 2 || picked[0] != got[0] || picked[1] != got[1] {
		t.Fatalf("preview %v vs real %v", picked, got)
	}

	rep := h.mustAPI(200, http.MethodGet, "/offer-report?days=30", nil)
	tot := rep["totals"].(map[string]any)
	if tot["purchases"].(float64) != 1 || tot["buyers"].(float64) != 1 || tot["shown"].(float64) < 6 || tot["coins"].(map[string]any)["gem"].(float64) != 5 {
		t.Fatalf("report %v", tot)
	}
	user := h.mustAPI(200, http.MethodGet, "/users/"+alice+"/offers", nil)
	if len(user["daily"].([]any)) != 2 || user["bought"].(float64) != 1 {
		t.Fatalf("player card %v", user)
	}
	if c, _ := h.api(http.MethodDelete, "/offers/"+pay.ProductID, nil); c != 404 {
		t.Fatalf("delete by product id %d", c)
	}
}

func TestDailyOfferRulesAndReroll(t *testing.T) {
	h := newOfferHarness(t)
	h.at(time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC))
	vip := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})["id"].(string)
	xp := h.meta("xp", 0)
	for _, k := range []string{"a", "b", "c", "d"} {
		h.offer("offers", "o"+k, h.coinProduct("p"+k, 10, 20000, nil), nil)
	}
	h.offer("offers", "vip_only", h.coinProduct("pv", 10, 20000, nil), map[string]any{"audienceTagIds": []string{vip}, "weight": 1000000})
	h.offer("offers", "veteran", h.coinProduct("pw", 10, 20000, nil), map[string]any{"metaConditions": []map[string]any{{"metaId": xp, "operator": "gte", "value": 50}}, "weight": 1000000})
	once := h.coinProduct("ponce", 10, 20000, map[string]any{"isOneTime": true})
	h.offer("offers", "once", once, map[string]any{"weight": 1000000})
	h.offerSettings(map[string]any{"count": 2, "countByTag": []map[string]any{{"tagId": vip, "count": 3}}, "rerollLimit": 2, "rerollCost": map[string]any{"currencyKey": "gem", "amount": 4}, "rerollCostStep": 2})

	u := h.user()
	h.gems(u, 30)
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, status, once_key) VALUES (gen_random_uuid(), $1, $2, 'in_app', 'success', $3)`, u, once, u+":"+once); err != nil {
		t.Fatal(err)
	}
	pv := h.mustAPI(200, http.MethodGet, "/offer-preview?userId="+u, nil)
	reasons := map[string]string{}
	for _, r := range pv["draw"].([]any) {
		reasons[r.(map[string]any)["key"].(string)] = r.(map[string]any)["reason"].(string)
	}
	if reasons["vip_only"] != "audience" || reasons["veteran"] != "meta" || reasons["once"] != "product_bought" {
		t.Fatalf("reasons %v", reasons)
	}
	h.setMeta(u, xp, "set", 60)
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, u, vip); err != nil {
		t.Fatal(err)
	}
	out := h.mustRPC(u, "economy_offers", map[string]any{})
	keys := offerKeys(dailyOf(out))
	if len(keys) != 3 || !contains(keys, "vip_only") || !contains(keys, "veteran") || contains(keys, "once") {
		t.Fatalf("vip count 3 with the heavy targeted offers: %v", keys)
	}
	rr := out["daily"].(map[string]any)["reroll"].(map[string]any)
	if rr["left"].(float64) != 2 || rr["cost"].(map[string]any)["amount"].(float64) != 4 {
		t.Fatalf("reroll info %v", rr)
	}

	// Buy one, then reroll: the bought one stays, the rest are replaced and cost 4 gems.
	bought := dailyOf(out)[0].(map[string]any)
	h.mustRPC(u, "economy_offer_buy", map[string]any{"offerId": bought["offerId"], "method": "in_app"})
	gemsBefore := h.wallet(u)["gem"]
	rolled := h.mustRPC(u, "economy_offers_reroll", map[string]any{})
	newKeys := offerKeys(dailyOf(rolled))
	if !contains(newKeys, bought["key"].(string)) || len(newKeys) > 3 {
		t.Fatalf("kept the bought offer: %v", newKeys)
	}
	for _, k := range keys {
		if k != bought["key"] && contains(newKeys, k) {
			t.Fatalf("replaced offers avoided: %v -> %v", keys, newKeys)
		}
	}
	if h.wallet(u)["gem"] != gemsBefore-4 {
		t.Fatalf("reroll cost: %d -> %d", gemsBefore, h.wallet(u)["gem"])
	}
	if c := h.count(`SELECT count(*) FROM economy_user_offer WHERE user_id=$1 AND status='replaced'`, u); c != 2 {
		t.Fatalf("replaced rows %d", c)
	}
	if rolled["daily"].(map[string]any)["reroll"].(map[string]any)["cost"].(map[string]any)["amount"].(float64) != 6 {
		t.Fatal("cost steps up")
	}
	// Not enough gems: nothing changes.
	h.gems(u, -h.wallet(u)["gem"])
	if c := h.rpcCode(u, "economy_offers_reroll", map[string]any{}); c != "INSUFFICIENT_BALANCE" && c != "NO_OFFERS_LEFT" {
		t.Fatalf("broke reroll %s", c)
	}
	if c := h.count(`SELECT rerolls FROM economy_user_offer_period WHERE user_id=$1`, u); c != 1 {
		t.Fatalf("failed reroll rolled back: %d", c)
	}
	h.offerSettings(map[string]any{"rerollCost": nil, "rerollLimit": 1})
	if c := h.rpcCode(u, "economy_offers_reroll", map[string]any{}); c != "REROLL_LIMIT" {
		t.Fatalf("limit %s", c)
	}
	h.offerSettings(map[string]any{"rerollEnabled": false})
	if c := h.rpcCode(u, "economy_offers_reroll", map[string]any{}); c != "REROLL_DISABLED" {
		t.Fatalf("disabled %s", c)
	}

	// Admin reset closes this period's set; the next list draws again.
	h.mustAPI(200, http.MethodPost, "/users/"+u+"/offers/reset-daily", map[string]any{})
	if n := len(dailyOf(h.mustRPC(u, "economy_offers", map[string]any{}))); n == 0 {
		t.Fatal("reset draws again")
	}
	// The bought cooldown keeps the bought offer out of the new draw.
	pv = h.mustAPI(200, http.MethodGet, "/offer-preview?userId="+u, nil)
	for _, r := range pv["draw"].([]any) {
		if row := r.(map[string]any); row["key"] == bought["key"] && row["reason"] != "bought_recently" {
			t.Fatalf("bought cooldown %v", row)
		}
	}
	h.offerSettings(map[string]any{"dailyEnabled": false})
	if out := h.mustRPC(u, "economy_offers", map[string]any{}); len(dailyOf(out)) != 0 || out["daily"].(map[string]any)["enabled"] != false {
		t.Fatalf("daily off %v", out["daily"])
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func TestEntryOfferTimersAndTriggers(t *testing.T) {
	h := newOfferHarness(t)
	now := time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
	h.at(now)
	h.offerSettings(map[string]any{"dailyEnabled": false})
	welcome := h.offer("entry-offers", "welcome", h.coinProduct("pw", 10, 20000, nil), map[string]any{"durationMinutes": 60, "sort": 1})
	h.offer("entry-offers", "comeback", h.coinProduct("pc", 10, 20000, nil), map[string]any{"triggerType": "lapsed", "triggerValue": 3, "durationMinutes": 120, "repeat": "cooldown", "cooldownHours": 1, "sort": 0})
	xp := h.meta("xp", 0)
	h.offer("entry-offers", "xp_reached", h.coinProduct("px", 10, 20000, nil), map[string]any{"triggerType": "meta", "triggerRef": xp, "triggerValue": 100, "durationMinutes": 30, "sort": 2})
	u := h.user()

	out := h.mustRPC(u, "economy_offers", map[string]any{})
	entry := entryOf(out)
	if len(entry) != 1 || entry[0].(map[string]any)["key"] != "welcome" || entry[0].(map[string]any)["secondsLeft"].(float64) != 3600 {
		t.Fatalf("welcome with a personal hour %v", entry)
	}
	// The timer is the player's own: later lists count down from the first sight.
	h.at(now.Add(20 * time.Minute))
	entry = entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))
	if entry[0].(map[string]any)["secondsLeft"].(float64) != 2400 {
		t.Fatalf("countdown %v", entry)
	}
	// One open at a time (default): xp_reached waits even though it holds.
	h.setMeta(u, xp, "set", 150)
	if len(entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))) != 1 {
		t.Fatal("max one active entry offer")
	}
	// After the hour the welcome offer is gone for good and the next one opens.
	h.at(now.Add(61 * time.Minute))
	entry = entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))
	if len(entry) != 1 || entry[0].(map[string]any)["key"] != "xp_reached" || entry[0].(map[string]any)["secondsLeft"].(float64) != 1800 {
		t.Fatalf("next entry %v", entry)
	}
	welcomeRow := ""
	_ = h.s.db.QueryRowContext(h.ctx, `SELECT id::text FROM economy_user_offer WHERE user_id=$1 AND offer_id=$2`, u, welcome).Scan(&welcomeRow)
	if c := h.rpcCode(u, "economy_offer_buy", map[string]any{"offerId": welcomeRow, "method": "in_app"}); c != "OFFER_EXPIRED" {
		t.Fatalf("expired welcome %s", c)
	}
	h.gems(u, 50)
	h.mustRPC(u, "economy_offer_buy", map[string]any{"offerId": entry[0].(map[string]any)["offerId"], "method": "in_app"})
	h.at(now.Add(62 * time.Minute))
	if n := len(entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))); n != 1 {
		t.Fatalf("a bought entry offer stays as sold out until its timer ends: %d", n)
	}
	h.at(now.Add(100 * time.Minute))
	if n := len(entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))); n != 0 {
		t.Fatalf("all done: %d", n)
	}

	// Lapsed return: logins 5 days apart fire the comeback offer once per return.
	h.at(now.Add(2 * time.Hour))
	h.s.noteOfferLogin(h.ctx, u)
	h.at(now.Add(5 * 24 * time.Hour))
	h.s.noteOfferLogin(h.ctx, u)
	entry = entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))
	if len(entry) != 1 || entry[0].(map[string]any)["key"] != "comeback" {
		t.Fatalf("comeback %v", entry)
	}
	h.at(now.Add(5*24*time.Hour + 3*time.Hour))
	if n := len(entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))); n != 0 {
		t.Fatalf("same return does not fire twice: %d", n)
	}
	h.s.noteOfferLogin(h.ctx, u)
	h.at(now.Add(9 * 24 * time.Hour))
	h.s.noteOfferLogin(h.ctx, u)
	if n := len(entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))); n != 1 {
		t.Fatalf("a new return after the cooldown fires again: %d", n)
	}
	// The admin forgets the welcome history: it fires again.
	h.mustAPI(200, http.MethodPost, "/users/"+u+"/entry-offers/"+welcome+"/reset", map[string]any{})
	card := h.mustAPI(200, http.MethodGet, "/users/"+u+"/offers", nil)
	if len(card["entry"].([]any)) < 3 {
		t.Fatalf("player card entry history %v", card["entry"])
	}
	h.offerSettings(map[string]any{"entryMaxActive": 3})
	keys := offerKeys(entryOf(h.mustRPC(u, "economy_offers", map[string]any{})))
	if !contains(keys, "welcome") {
		t.Fatalf("welcome again after reset: %v", keys)
	}
}

func TestEntryOfferTimerModes(t *testing.T) {
	h := newOfferHarness(t)
	now := time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
	h.at(now)
	h.offerSettings(map[string]any{"dailyEnabled": false, "entryTimerStart": "seen"})
	h.offer("entry-offers", "welcome", h.coinProduct("pw", 10, 20000, nil), map[string]any{"durationMinutes": 10})
	u := h.user()
	e := entryOf(h.mustRPC(u, "economy_offers", map[string]any{}))[0].(map[string]any)
	if e["started"] != false || e["secondsLeft"].(float64) != 0 {
		t.Fatalf("waits for seen %v", e)
	}
	h.at(now.Add(time.Hour))
	seen := h.mustRPC(u, "economy_offer_seen", map[string]any{"offerId": e["offerId"]})
	if seen["secondsLeft"].(float64) != 600 {
		t.Fatalf("seen starts the timer %v", seen)
	}
	h.offerSettings(map[string]any{"entryTimerStart": "trigger"})
	v := h.user()
	h.mustRPC(v, "economy_offers", map[string]any{})
	h.at(now.Add(time.Hour + 5*time.Minute))
	e = entryOf(h.mustRPC(v, "economy_offers", map[string]any{}))[0].(map[string]any)
	if e["secondsLeft"].(float64) != 300 {
		t.Fatalf("trigger mode counts from the trigger %v", e)
	}
}

func TestOfferGatewayFreeAndPerOfferLimit(t *testing.T) {
	h := newOfferHarness(t)
	h.at(time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC))
	h.offerSettings(map[string]any{"count": 1})
	prod := h.coinProduct("pf", 10, 20000, nil)
	h.offer("offers", "free", prod, map[string]any{"discountPercent": 100, "purchaseLimit": 2})
	u := h.user()
	it := dailyOf(h.mustRPC(u, "economy_offers", map[string]any{}))[0].(map[string]any)
	if it["finalToman"].(float64) != 0 || it["purchaseLimit"].(float64) != 2 {
		t.Fatalf("free offer %v", it)
	}
	if c := h.rpcCode(u, "economy_offer_buy", map[string]any{"offerId": it["offerId"], "method": "zibal", "giftCode": "X"}); c != "GIFT_NOT_APPLICABLE" {
		t.Fatalf("codes on offers are off by default: %s", c)
	}
	for i := 0; i < 2; i++ {
		res := h.mustRPC(u, "economy_offer_buy", map[string]any{"offerId": it["offerId"], "method": "zibal"})
		if res["status"] != "success" {
			t.Fatalf("free gateway buy %v", res)
		}
		if h.count(`SELECT count(*) FROM economy_payment WHERE id=$1 AND final_toman=0 AND original_toman=20000 AND discount_percent=100 AND user_offer_id IS NOT NULL`, res["paymentId"]) != 1 {
			t.Fatal("payment row carries the offer price")
		}
	}
	if h.count(`SELECT bought FROM economy_user_offer WHERE id=$1`, it["offerId"]) != 2 {
		t.Fatal("gateway grants count toward the offer limit")
	}
	if c := h.rpcCode(u, "economy_offer_buy", map[string]any{"offerId": it["offerId"], "method": "zibal"}); c != "OFFER_SOLD_OUT" {
		t.Fatalf("limit 2: %s", c)
	}
	// The plain shop is unchanged: the product still sells at its own price.
	cat := h.mustRPC(u, "economy_catalog", map[string]any{})
	for _, p := range cat["products"].([]any) {
		if p.(map[string]any)["key"] == "pf" && p.(map[string]any)["finalToman"].(float64) != 20000 {
			t.Fatalf("catalog price changed %v", p)
		}
	}
	// Deleting a template that sold is refused; deactivating it hides it.
	if c, out := h.api(http.MethodDelete, "/offers/"+it["offerId"].(string), nil); c != 404 {
		t.Fatalf("a personal copy id is not a template id: %d %v", c, out)
	}
}
