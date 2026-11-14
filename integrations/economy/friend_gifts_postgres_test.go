package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

// The harness has no socket; friend gift notices go nowhere.
func (n *fakeNakama) NotificationSend(ctx context.Context, userID, subject string, content map[string]any, code int, sender string, persistent bool) error {
	return nil
}

func newGiftHarness(t *testing.T) (*harness, string) {
	h := newEventHarness(t)
	for _, q := range []string{
		`TRUNCATE economy_friend_gift, economy_friend_gift_rule CASCADE`,
		`DELETE FROM economy_friend_gift_settings`,
		`INSERT INTO economy_event_type (id, key, name, is_system, value_min, rate_limit) VALUES (gen_random_uuid(), 'Gift:Sent', '{"fa":"هدیه"}', true, 0, 1000), (gen_random_uuid(), 'Gift:Received', '{"fa":"هدیه"}', true, 0, 1000), (gen_random_uuid(), 'purchase_completed', '{"fa":"خرید"}', true, 0, 1000)`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for _, rpcs := range []map[string]clientRPC{h.s.friendGiftRPCs(), h.s.itemRPCs()} {
		for name, fn := range rpcs {
			h.rpcs[name] = fn
		}
	}
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس", "en": "Gem"}, "type": "hard"})
	h.mustAPI(200, http.MethodPut, "/payment-methods/in_app", map[string]any{"isActive": true})
	h.mustAPI(200, http.MethodPut, "/payment-methods/zibal", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true}})
	skin := h.taxonomy("base_category", "item_type", "skin", map[string]string{"fa": "پوسته", "en": "Skin"})["id"].(string)
	return h, skin
}

func (h *harness) friends(a, b string) {
	h.t.Helper()
	for _, pair := range [][2]string{{a, b}, {b, a}} {
		if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO user_edge (source_id, position, destination_id, state, update_time) VALUES ($1, $3, $2, 0, now() - interval '10 days')
			ON CONFLICT (source_id, destination_id) DO UPDATE SET state=0`, pair[0], pair[1], time.Now().UnixNano()); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *harness) giftSettings(change map[string]any) map[string]any {
	h.t.Helper()
	cur := h.mustAPI(200, http.MethodGet, "/friend-gift-settings", nil)["settings"].(map[string]any)
	for k, v := range change {
		cur[k] = v
	}
	return h.mustAPI(200, http.MethodPut, "/friend-gift-settings", cur)["settings"].(map[string]any)
}

func (h *harness) grantItem(user, itemID string) {
	h.t.Helper()
	h.mustAPI(200, http.MethodPost, "/users/"+user+"/items", map[string]any{"itemId": itemID})
}

func (h *harness) ownsItem(user, key string) bool {
	h.t.Helper()
	ok, err := owns(h.ctx, h.s.db, user, key)
	if err != nil {
		h.t.Fatal(err)
	}
	return ok
}

func (h *harness) gift(id string) map[string]any {
	h.t.Helper()
	return h.mustAPI(200, http.MethodGet, "/friend-gifts/"+id, nil)
}

func (h *harness) send(user string, body map[string]any) map[string]any {
	h.t.Helper()
	return h.mustRPC(user, "economy_friend_gift_send", body)["gift"].(map[string]any)
}

func (h *harness) events(user, key string) int {
	h.t.Helper()
	var n int
	if err := h.s.db.QueryRow(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE e.user_id=$1 AND t.key=$2`, user, key).Scan(&n); err != nil {
		h.t.Fatal(err)
	}
	return n
}

func TestFriendGiftInventoryItems(t *testing.T) {
	h, skin := newGiftHarness(t)
	car := h.taxonomy("base_category", "item_type", "car", map[string]string{"fa": "ماشین"})["id"].(string)
	a, b, c, d := h.user(), h.user(), h.user(), h.user()
	h.friends(a, b)
	h.friends(a, d)
	h.friends(b, d)
	sword := h.item(map[string]any{"key": "sword", "typeId": skin, "upgrades": []map[string]any{{"currencyKey": "coin", "amount": 5}}})
	shield := h.item(map[string]any{"key": "shield", "typeId": skin, "giftable": false})
	starter := h.item(map[string]any{"key": "starter", "typeId": skin, "isStarter": true})
	truck := h.item(map[string]any{"key": "truck", "typeId": car})
	hat := h.item(map[string]any{"key": "hat", "typeId": skin})
	if shield["giftable"] != false || sword["giftable"] != nil {
		t.Fatalf("giftable flag %v %v", shield["giftable"], sword["giftable"])
	}
	for _, it := range []map[string]any{sword, shield, truck, hat} {
		h.grantItem(a, it["id"].(string))
	}
	_ = starter
	if _, err := h.s.db.Exec(`UPDATE economy_user_item SET upgrade_level=1 WHERE user_id=$1 AND item_id=$2`, a, sword["id"]); err != nil {
		t.Fatal(err)
	}

	opts := h.mustRPC(a, "economy_friend_gift_options", map[string]any{"friendId": b, "lang": "en"})
	if opts["canSend"] != true || opts["delivery"] != "instant" {
		t.Fatalf("options %v", opts)
	}
	reasons := map[string]string{}
	for _, raw := range opts["items"].([]any) {
		it := raw.(map[string]any)
		reasons[it["key"].(string)] = it["reason"].(string)
	}
	if reasons["sword"] != "" || reasons["shield"] != "ITEM_NOT_GIFTABLE" || reasons["truck"] != "" || len(reasons) != 4 {
		t.Fatalf("item reasons %v", reasons)
	}

	g := h.send(a, map[string]any{"friendId": b, "itemKey": "sword", "message": "  for you  "})
	if g["status"] != "delivered" || g["box"] != "sent" || g["message"] != "for you" || g["kind"] != "item" {
		t.Fatalf("sent %v", g)
	}
	if h.ownsItem(a, "sword") || !h.ownsItem(b, "sword") {
		t.Fatal("the item moved from the sender to the recipient")
	}
	var src string
	var lvl int
	_ = h.s.db.QueryRow(`SELECT source, upgrade_level FROM economy_user_item WHERE user_id=$1 AND item_id=$2`, b, sword["id"]).Scan(&src, &lvl)
	if src != "gift" || lvl != 1 {
		t.Fatalf("recipient copy %s %d", src, lvl)
	}
	if h.events(a, "Gift:Sent") != 1 || h.events(b, "Gift:Received") != 1 {
		t.Fatal("gift events")
	}
	if c := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "itemKey": "sword"}); c != "ITEM_NOT_OWNED" {
		t.Fatalf("second send %s", c)
	}
	if c := h.rpcCode(b, "economy_friend_gift_send", map[string]any{"friendId": d, "itemKey": "sword"}); c != "REGIFT_BLOCKED" {
		t.Fatalf("regift %s", c)
	}
	h.giftSettings(map[string]any{"regift": "allow"})
	h.send(b, map[string]any{"friendId": d, "itemKey": "sword"})
	if !h.ownsItem(d, "sword") {
		t.Fatal("regift allowed")
	}

	for _, tc := range []struct {
		to, key, code string
	}{
		{c, "truck", "NOT_FRIENDS"},
		{a, "truck", "GIFT_SELF"},
		{b, "shield", "ITEM_NOT_GIFTABLE"},
		{b, "starter", "STARTER_ITEM"},
		{b, "nope", "ITEM_NOT_FOUND"},
		{uuid.Must(uuid.NewV4()).String(), "truck", "RECIPIENT_NOT_FOUND"},
	} {
		if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": tc.to, "itemKey": tc.key}); got != tc.code {
			t.Fatalf("%s to %s = %s want %s", tc.key, tc.to, got, tc.code)
		}
	}
	h.grantItem(b, hat["id"].(string))
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "itemKey": "hat"}); got != "RECIPIENT_OWNS_ITEM" {
		t.Fatalf("recipient owns = %s", got)
	}
	opts = h.mustRPC(a, "economy_friend_gift_options", map[string]any{"friendId": b})
	for _, raw := range opts["items"].([]any) {
		if it := raw.(map[string]any); it["key"] == "hat" && (it["recipientOwns"] != true || it["giftable"] != false) {
			t.Fatalf("options mark what the friend owns: %v", it)
		}
	}
	// item-type default and per-item override
	h.giftSettings(map[string]any{"itemTypes": []map[string]any{{"typeId": car, "giftable": false}}})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "itemKey": "truck"}); got != "ITEM_NOT_GIFTABLE" {
		t.Fatalf("type default = %s", got)
	}
	truck["giftable"] = true
	h.mustAPI(200, http.MethodPut, "/items/"+truck["id"].(string), truck)
	// equipped
	if _, err := h.s.db.Exec(`UPDATE economy_user_item SET equipped=true, slot='car' WHERE user_id=$1 AND item_id=$2`, a, truck["id"]); err != nil {
		t.Fatal(err)
	}
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "itemKey": "truck"}); got != "ITEM_EQUIPPED" {
		t.Fatalf("equipped = %s", got)
	}
	h.giftSettings(map[string]any{"equipped": "unequip", "upgradeLevel": "reset"})
	h.send(a, map[string]any{"friendId": b, "itemKey": "truck"})
	if !h.ownsItem(b, "truck") {
		t.Fatal("unequip allows the gift")
	}
	// blocked either way
	if _, err := h.s.db.Exec(`UPDATE user_edge SET state=3 WHERE source_id=$1 AND destination_id=$2`, d, a); err != nil {
		t.Fatal(err)
	}
	_, _ = h.s.db.Exec(`DELETE FROM economy_user_item WHERE user_id=$1 AND item_id=$2`, b, hat["id"])
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": d, "itemKey": "hat"}); got != "GIFT_BLOCKED" {
		t.Fatalf("blocked = %s", got)
	}
	// friendship age and "any player"
	h.giftSettings(map[string]any{"minFriendDays": 30})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "itemKey": "hat"}); got != "FRIENDSHIP_TOO_NEW" {
		t.Fatalf("friend days = %s", got)
	}
	h.giftSettings(map[string]any{"minFriendDays": 0, "recipients": "any"})
	h.send(a, map[string]any{"friendId": c, "itemKey": "hat"})
	if !h.ownsItem(c, "hat") {
		t.Fatal("any player")
	}
	list := h.mustRPC(c, "economy_friend_gifts", map[string]any{"box": "received", "lang": "en"})
	items := list["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["name"] != "hat" || items[0].(map[string]any)["from"].(map[string]any)["id"] != a || items[0].(map[string]any)["method"] != nil {
		t.Fatalf("received list %v", list)
	}
	sent := h.mustRPC(a, "economy_friend_gifts", map[string]any{"box": "sent", "limit": 2})
	if len(sent["items"].([]any)) != 2 || sent["cursor"] == "" {
		t.Fatalf("sent page %v", sent)
	}
	more := h.mustRPC(a, "economy_friend_gifts", map[string]any{"box": "sent", "limit": 2, "cursor": sent["cursor"]})
	if len(more["items"].([]any)) != 1 {
		t.Fatalf("second page %v", more)
	}
}

func TestFriendGiftLimitsRulesAndCurrency(t *testing.T) {
	h, _ := newGiftHarness(t)
	a, b, c, d := h.user(), h.user(), h.user(), h.user()
	h.friends(a, b)
	h.friends(a, c)
	h.friends(d, b)
	h.give(a, 1000)
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 5}); got != "CURRENCY_NOT_GIFTABLE" {
		t.Fatalf("currency off by default = %s", got)
	}
	h.giftSettings(map[string]any{"currencyEnabled": true, "currencies": []map[string]any{{"currencyKey": "coin", "min": 5, "max": 100}}, "perSenderDaily": 3, "perRecipientDaily": 3, "perPairDaily": 2})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 4}); got != "INVALID_AMOUNT" {
		t.Fatalf("min = %s", got)
	}
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 101}); got != "INVALID_AMOUNT" {
		t.Fatalf("max = %s", got)
	}
	h.send(a, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 10})
	if h.wallet(a)["coin"] != 990 || h.wallet(b)["coin"] != 10 {
		t.Fatalf("wallets %v %v", h.wallet(a), h.wallet(b))
	}
	h.send(a, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 10})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 10}); got != "GIFT_PAIR_DAILY_LIMIT" {
		t.Fatalf("pair = %s", got)
	}
	h.send(a, map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10}); got != "GIFT_SENDER_DAILY_LIMIT" {
		t.Fatalf("sender = %s", got)
	}
	h.give(d, 100)
	h.send(d, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 10})
	if got := h.rpcCode(d, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 10}); got != "GIFT_RECIPIENT_DAILY_LIMIT" {
		t.Fatalf("recipient = %s", got)
	}
	opts := h.mustRPC(a, "economy_friend_gift_options", map[string]any{"friendId": c})
	if opts["canSend"] != false || opts["reason"] != "GIFT_SENDER_DAILY_LIMIT" {
		t.Fatalf("options reason %v", opts)
	}
	lim := opts["limits"].(map[string]any)["sender"].(map[string]any)
	if lim["used"].(float64) != 3 || lim["limit"].(float64) != 3 || lim["remaining"].(float64) != 0 {
		t.Fatalf("sender limits %v", lim)
	}
	// a tag rule raises the sender limit, a bonus meta adds purchasable quota
	vip := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})["id"].(string)
	if _, err := h.s.db.Exec(`INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, a, vip); err != nil {
		t.Fatal(err)
	}
	if code, out := h.api(http.MethodPost, "/friend-gift-rules", map[string]any{"name": "VIP", "tagIds": []string{}}); code != 400 || out["error"] != "TAGS_REQUIRED" {
		t.Fatalf("rule without tags %d %v", code, out)
	}
	rule := h.mustAPI(201, http.MethodPost, "/friend-gift-rules", map[string]any{"name": "VIP", "tagIds": []string{vip}, "perSenderDaily": 4, "priority": 5})
	if rule["players"].(float64) != 1 || rule["canSend"] != true {
		t.Fatalf("rule %v", rule)
	}
	h.send(a, map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10}); got != "GIFT_SENDER_DAILY_LIMIT" {
		t.Fatalf("limit after rule = %s", got)
	}
	h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "gift_bonus", "name": map[string]string{"fa": "سهمیهٔ هدیه"}})
	h.giftSettings(map[string]any{"sendBonusMeta": "gift_bonus", "perPairDaily": 0, "perRecipientDaily": 0})
	if _, err := h.s.db.Exec(`INSERT INTO economy_user_meta (user_id, meta_id, value) SELECT $1, id, 2 FROM economy_meta WHERE key='gift_bonus'`, a); err != nil {
		t.Fatal(err)
	}
	h.send(a, map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10})
	h.send(a, map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10}); got != "GIFT_SENDER_DAILY_LIMIT" {
		t.Fatalf("limit with bonus = %s", got)
	}
	h.mustAPI(200, http.MethodPut, "/friend-gift-rules/"+rule["id"].(string), map[string]any{"name": "VIP", "tagIds": []string{vip}, "canSend": false})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": c, "currencyKey": "coin", "amount": 10}); got != "GIFT_SENDING_BLOCKED" {
		t.Fatalf("rule blocks = %s", got)
	}
	card := h.mustAPI(200, http.MethodGet, "/users/"+a+"/friend-gifts", nil)
	if card["sentToday"].(float64) != 6 || card["limits"].(map[string]any)["canSend"] != false {
		t.Fatalf("player card %v", card)
	}
	// insufficient coins: nothing moves, the gift is failed and does not count
	h.mustAPI(200, http.MethodDelete, "/friend-gift-rules/"+rule["id"].(string), nil)
	h.giftSettings(map[string]any{"perSenderDaily": 0, "perRecipientDaily": 0, "currencies": []map[string]any{{"currencyKey": "coin", "min": 1, "max": 0}}})
	poor := h.user()
	h.friends(poor, c)
	if got := h.rpcCode(poor, "economy_friend_gift_send", map[string]any{"friendId": c, "currencyKey": "coin", "amount": 5}); got != "INSUFFICIENT_BALANCE" {
		t.Fatalf("poor = %s", got)
	}
	var st string
	_ = h.s.db.QueryRow(`SELECT status FROM economy_friend_gift WHERE sender_id=$1`, poor).Scan(&st)
	if st != "failed" {
		t.Fatalf("failed gift status %s", st)
	}
}

func TestFriendGiftAcceptDeclineExpire(t *testing.T) {
	h, skin := newGiftHarness(t)
	a, b := h.user(), h.user()
	h.friends(a, b)
	h.give(a, 500)
	h.giftSettings(map[string]any{"delivery": "accept", "expiryHours": 2, "currencyEnabled": true, "currencies": []map[string]any{{"currencyKey": "coin", "min": 1, "max": 0}}, "perPairDaily": 0})
	cape := h.item(map[string]any{"key": "cape", "typeId": skin})
	h.grantItem(a, cape["id"].(string))
	if _, err := h.s.db.Exec(`UPDATE economy_user_item SET upgrade_level=2 WHERE user_id=$1`, a); err != nil {
		t.Fatal(err)
	}

	// item: held while pending, declined goes back as it was
	g := h.send(a, map[string]any{"friendId": b, "itemKey": "cape"})
	if g["status"] != "pending" || g["secondsLeft"].(float64) < 7000 || h.ownsItem(a, "cape") || h.ownsItem(b, "cape") {
		t.Fatalf("pending item %v", g)
	}
	if got := h.rpcCode(b, "economy_friend_gift_send", map[string]any{"friendId": a, "itemKey": "cape"}); got != "ITEM_NOT_OWNED" {
		t.Fatalf("the recipient does not own a pending gift: %s", got)
	}
	if got := h.rpcCode(a, "economy_friend_gift_accept", map[string]any{"giftId": g["id"]}); got != "GIFT_NOT_FOUND" {
		t.Fatalf("the sender cannot accept: %s", got)
	}
	list := h.mustRPC(b, "economy_friend_gifts", map[string]any{"status": "pending"})
	if list["pendingCount"].(float64) != 1 || len(list["items"].([]any)) != 1 {
		t.Fatalf("pending list %v", list)
	}
	dec := h.mustRPC(b, "economy_friend_gift_decline", map[string]any{"giftId": g["id"]})["gift"].(map[string]any)
	if dec["status"] != "declined" || dec["outcome"] != "returned" || !h.ownsItem(a, "cape") {
		t.Fatalf("declined %v", dec)
	}
	var lvl int
	_ = h.s.db.QueryRow(`SELECT upgrade_level FROM economy_user_item WHERE user_id=$1 AND item_id=$2`, a, cape["id"]).Scan(&lvl)
	if lvl != 2 {
		t.Fatal("the returned item keeps its upgrade")
	}
	if got := h.rpcCode(b, "economy_friend_gift_accept", map[string]any{"giftId": g["id"]}); got != "GIFT_NOT_PENDING" {
		t.Fatalf("accept after decline %s", got)
	}

	// item accepted
	g = h.send(a, map[string]any{"friendId": b, "itemKey": "cape"})
	acc := h.mustRPC(b, "economy_friend_gift_accept", map[string]any{"giftId": g["id"]})["gift"].(map[string]any)
	if acc["status"] != "delivered" || !h.ownsItem(b, "cape") || h.events(b, "Gift:Received") != 1 {
		t.Fatalf("accepted %v", acc)
	}

	// currency: escrow, accept, cancel by sender, expiry
	g = h.send(a, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 100})
	if h.wallet(a)["coin"] != 400 || h.wallet(b)["coin"] != 0 {
		t.Fatalf("escrow %v %v", h.wallet(a), h.wallet(b))
	}
	h.mustRPC(b, "economy_friend_gift_accept", map[string]any{"giftId": g["id"]})
	if h.wallet(b)["coin"] != 100 {
		t.Fatal("accepted coins")
	}
	g = h.send(a, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 50})
	h.mustRPC(a, "economy_friend_gift_cancel", map[string]any{"giftId": g["id"]})
	if h.wallet(a)["coin"] != 400 || h.gift(g["id"].(string))["outcome"] != "returned" {
		t.Fatalf("cancelled %v", h.gift(g["id"].(string)))
	}
	h.giftSettings(map[string]any{"senderCancel": false})
	g = h.send(a, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 50})
	if got := h.rpcCode(a, "economy_friend_gift_cancel", map[string]any{"giftId": g["id"]}); got != "GIFT_CANCEL_OFF" {
		t.Fatalf("cancel off %s", got)
	}
	later := time.Now().Add(3 * time.Hour)
	h.s.now = func() time.Time { return later }
	if got := h.rpcCode(b, "economy_friend_gift_accept", map[string]any{"giftId": g["id"]}); got != "GIFT_EXPIRED" {
		t.Fatalf("accept after expiry %s", got)
	}
	h.s.expireFriendGifts(h.ctx, 10)
	if row := h.gift(g["id"].(string)); row["status"] != "expired" || row["outcome"] != "returned" || h.wallet(a)["coin"] != 400 {
		t.Fatalf("expired %v %v", row, h.wallet(a))
	}
	h.s.now = time.Now
	h.giftSettings(map[string]any{"onExpiry": "accept"})
	g = h.send(a, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 25})
	h.s.now = func() time.Time { return later }
	h.s.expireFriendGifts(h.ctx, 10)
	h.s.now = time.Now
	if row := h.gift(g["id"].(string)); row["status"] != "delivered" || h.wallet(b)["coin"] != 125 {
		t.Fatalf("auto-accept %v %v", row, h.wallet(b))
	}
}

func TestFriendGiftShopPurchase(t *testing.T) {
	h, skin := newGiftHarness(t)
	a, b, c := h.user(), h.user(), h.user()
	h.friends(a, b)
	h.friends(a, c)
	h.gems(a, 1000)
	h.giftSettings(map[string]any{"perPairDaily": 0})
	h.item(map[string]any{"key": "wings", "typeId": skin})
	wingsPrize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "pz_wings", "name": map[string]string{"fa": "بال"}, "isRepeatable": true,
		"items": []map[string]any{{"type": "item", "key": "wings", "action": "give"}}})["id"].(string)
	wings := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "wings_pack", "isActive": true, "name": map[string]string{"fa": "بستهٔ بال", "en": "Wings pack"}, "prizeId": wingsPrize,
		"paymentMethods": []string{"in_app", "zibal"}, "coinPrice": map[string]any{"currencyKey": "gem", "amount": 100}, "priceToman": 50000})["id"].(string)
	coins := h.coinProduct("coins_pack", 30, 1000, map[string]any{"isOneTime": true})
	h.coinProduct("no_gift", 10, 1000, map[string]any{"giftable": false})

	opts := h.mustRPC(a, "economy_friend_gift_options", map[string]any{"friendId": b, "lang": "en"})
	keys := map[string]map[string]any{}
	for _, raw := range opts["products"].([]any) {
		p := raw.(map[string]any)
		keys[p["key"].(string)] = p
	}
	if keys["no_gift"] != nil || keys["wings_pack"] == nil || keys["wings_pack"]["giftable"] != true || keys["coins_pack"] == nil {
		t.Fatalf("product options %v", keys)
	}
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": wings, "method": "bazaar"}); got != "METHOD_NOT_ALLOWED" {
		t.Fatalf("method = %s", got)
	}
	res := h.mustRPC(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": wings, "method": "in_app", "message": "happy birthday"})
	g := res["gift"].(map[string]any)
	if res["status"] != "success" || res["paymentId"] == nil || h.wallet(a)["gem"] != 900 || !h.ownsItem(b, "wings") || h.ownsItem(a, "wings") {
		t.Fatalf("in-app gift %v gems %v", res, h.wallet(a))
	}
	row := h.gift(g["id"].(string))
	if row["status"] != "delivered" || row["outcome"] != "delivered" || row["paymentStatus"] != "success" || row["coinAmount"].(float64) != 100 {
		t.Fatalf("gift row %v", row)
	}
	p := h.payment(res["paymentId"].(string))
	if p.UserID != a || !p.RewardGranted {
		t.Fatalf("payment %+v", p)
	}
	st := h.mustRPC(a, "economy_purchase_status", map[string]any{"paymentId": p.ID})
	if st["friendGiftId"] != g["id"] || st["friendGiftStatus"] != "delivered" {
		t.Fatalf("purchase status %v", st)
	}
	if h.events(a, "Gift:Sent") != 1 || h.events(b, "Gift:Received") != 1 {
		t.Fatal("events for a shop gift")
	}
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": wings, "method": "in_app"}); got != "RECIPIENT_OWNS_PRODUCT" {
		t.Fatalf("owned = %s", got)
	}
	// a one-time product: the gift does not use up the sender's own purchase
	h.mustRPC(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": coins, "method": "in_app"})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": coins, "method": "in_app"}); got != "RECIPIENT_OWNS_PRODUCT" {
		t.Fatalf("one-time had = %s", got)
	}
	if counts, _ := h.s.purchaseCounts(h.ctx, a); counts[coins].Count != 0 {
		t.Fatal("a gift purchase is not the sender's own purchase")
	}
	if h.wallet(b)["coin"] != 100 {
		t.Fatalf("coins prize to the recipient %v", h.wallet(b))
	}
	// free gateway (price 0) for another friend
	free := h.coinProduct("free_pack", 5, 1000, map[string]any{"discountPercent": 100})
	res = h.mustRPC(a, "economy_friend_gift_send", map[string]any{"friendId": c, "productId": free, "method": "zibal"})
	if res["status"] != "success" || h.wallet(c)["coin"] != 100 {
		t.Fatalf("free gateway gift %v %v", res, h.wallet(c))
	}

	// paid gateway with a callback, then decline with money -> the sender gets it
	h.giftSettings(map[string]any{"delivery": "accept"})
	zibal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/request":
			_, _ = w.Write([]byte(`{"result":100,"trackId":` + fmt.Sprint(time.Now().UnixNano()%1e12) + `}`))
		case "/v1/verify":
			_, _ = w.Write([]byte(`{"result":100,"amount":500000,"refNumber":77,"cardNumber":"6037****"}`))
		case "/v1/inquiry":
			_, _ = w.Write([]byte(`{"result":100,"status":1,"amount":500000}`))
		}
	}))
	defer zibal.Close()
	swap(t, &zibalBase, zibal.URL)
	res = h.mustRPC(a, "economy_friend_gift_send", map[string]any{"friendId": c, "productId": wings, "method": "zibal"})
	g = res["gift"].(map[string]any)
	if res["status"] != "redirect" || g["status"] != "paying" {
		t.Fatalf("gateway start %v", res)
	}
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": c, "productId": wings, "method": "in_app"}); got != "GIFT_ALREADY_PENDING" {
		t.Fatalf("second buy while paying = %s", got)
	}
	pay := h.payment(res["paymentId"].(string))
	ticket := signTicket(h.s.cfg.ticketKey, pay.ID, time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/v2/economy/payments/callback/"+ticket+"?"+url.Values{"trackId": {pay.TrackID.String}, "success": {"1"}}.Encode(), nil)
	rec := httptest.NewRecorder()
	h.init.public.ServeHTTP(rec, req)
	if row := h.gift(g["id"].(string)); row["status"] != "pending" || h.ownsItem(c, "wings") || !h.payment(pay.ID).RewardGranted {
		t.Fatalf("paid then pending %d %v", rec.Code, row)
	}
	h.mustRPC(c, "economy_friend_gift_decline", map[string]any{"giftId": g["id"]})
	if row := h.gift(g["id"].(string)); row["status"] != "declined" || row["outcome"] != "sender_got" || !h.ownsItem(a, "wings") {
		t.Fatalf("money gift declined %v", row)
	}
	// coin gift declined -> coins back
	before := h.wallet(a)["gem"]
	res = h.mustRPC(a, "economy_friend_gift_send", map[string]any{"friendId": c, "productId": coins, "method": "in_app"})
	g = res["gift"].(map[string]any)
	if h.wallet(a)["gem"] != before-30 {
		t.Fatal("coins paid")
	}
	h.mustRPC(c, "economy_friend_gift_decline", map[string]any{"giftId": g["id"]})
	if row := h.gift(g["id"].(string)); row["outcome"] != "refunded" || h.wallet(a)["gem"] != before {
		t.Fatalf("coin gift refunded %v %v", row, h.wallet(a))
	}
	// a failed gateway start fails the gift
	zibal.Close()
	h.giftSettings(map[string]any{"delivery": "instant", "perSenderDaily": 0})
	paid := h.coinProduct("paid_pack", 5, 2000, nil)
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": paid, "method": "zibal"}); got != "GATEWAY_REQUEST_FAILED" {
		t.Fatalf("gateway down = %s", got)
	}
	var st2, perr string
	_ = h.s.db.QueryRow(`SELECT g.status, p.status FROM economy_friend_gift g JOIN economy_payment p ON p.friend_gift_id=g.id WHERE g.product_id=$1`, paid).Scan(&st2, &perr)
	if st2 != "failed" || perr != "failed" {
		t.Fatalf("a closed payment fails the gift: %s %s", st2, perr)
	}
	h.mustRPC(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": paid, "method": "in_app"})
}

func TestFriendGiftConcurrencyAndIdempotency(t *testing.T) {
	h, skin := newGiftHarness(t)
	a := h.user()
	ring := h.item(map[string]any{"key": "ring", "typeId": skin})
	h.grantItem(a, ring["id"].(string))
	var friends []string
	for i := 0; i < 8; i++ {
		f := h.user()
		h.friends(a, f)
		friends = append(friends, f)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, codes := 0, map[string]int{}
	for _, f := range friends {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			_, err := h.rpc(a, "economy_friend_gift_send", map[string]any{"friendId": f, "itemKey": "ring"})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				ok++
			} else if app, isApp := asApp(err); isApp {
				codes[app.Code]++
			} else {
				codes[err.Error()]++
			}
		}(f)
	}
	wg.Wait()
	var owners int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM economy_user_item WHERE item_id=$1`, ring["id"]).Scan(&owners)
	if ok != 1 || codes["ITEM_NOT_OWNED"] != 7 || owners != 1 {
		t.Fatalf("one item, one gift: ok=%d codes=%v owners=%d", ok, codes, owners)
	}

	// parallel currency sends against a daily limit of 3
	h.give(a, 1000)
	h.giftSettings(map[string]any{"currencyEnabled": true, "currencies": []map[string]any{{"currencyKey": "coin", "min": 1, "max": 0}}, "perSenderDaily": 4, "perPairDaily": 0})
	ok = 0
	codes = map[string]int{}
	for _, f := range friends {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			_, err := h.rpc(a, "economy_friend_gift_send", map[string]any{"friendId": f, "currencyKey": "coin", "amount": 10})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				ok++
			} else if app, isApp := asApp(err); isApp {
				codes[app.Code]++
			}
		}(f)
	}
	wg.Wait()
	if ok != 3 || codes["GIFT_SENDER_DAILY_LIMIT"] != 5 || h.wallet(a)["coin"] != 970 {
		t.Fatalf("limit under load: ok=%d codes=%v wallet=%v", ok, codes, h.wallet(a))
	}

	// the same requestId twice sends one gift
	h.giftSettings(map[string]any{"perSenderDaily": 0})
	b := friends[0]
	body := map[string]any{"friendId": b, "currencyKey": "coin", "amount": 7, "requestId": "req-1"}
	var replays int
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := h.rpc(a, "economy_friend_gift_send", body)
			if err == nil && out["replayed"] == true {
				mu.Lock()
				replays++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	var n int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM economy_friend_gift WHERE sender_id=$1 AND request_id='req-1'`, a).Scan(&n)
	if n != 1 || replays != 3 || h.wallet(a)["coin"] != 963 {
		t.Fatalf("idempotent send: rows=%d replays=%d wallet=%v", n, replays, h.wallet(a))
	}
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 8, "requestId": "req-1"}); got != "REQUEST_ID_REUSED" {
		t.Fatalf("reused id = %s", got)
	}
}

func TestFriendGiftReverseAndConsole(t *testing.T) {
	h, skin := newGiftHarness(t)
	a, b := h.user(), h.user()
	h.friends(a, b)
	h.give(a, 300)
	h.gems(a, 500)
	h.giftSettings(map[string]any{"currencyEnabled": true, "currencies": []map[string]any{{"currencyKey": "coin", "min": 1, "max": 0}}, "perPairDaily": 0})
	orb := h.item(map[string]any{"key": "orb", "typeId": skin})
	h.grantItem(a, orb["id"].(string))
	gi := h.send(a, map[string]any{"friendId": b, "itemKey": "orb"})
	gc := h.send(a, map[string]any{"friendId": b, "currencyKey": "coin", "amount": 100})
	pack := h.coinProduct("rev_pack", 50, 1000, nil)
	gp := h.mustRPC(a, "economy_friend_gift_send", map[string]any{"friendId": b, "productId": pack, "method": "in_app"})["gift"].(map[string]any)
	if h.wallet(b)["coin"] != 200 || h.wallet(a)["gem"] != 450 {
		t.Fatalf("before reverse %v %v", h.wallet(a), h.wallet(b))
	}

	list := h.mustAPI(200, http.MethodGet, "/friend-gifts?userId="+a, nil)
	if list["total"].(float64) != 3 {
		t.Fatalf("log %v", list)
	}
	if l := h.mustAPI(200, http.MethodGet, "/friend-gifts?kind=currency&recipientId="+b, nil); l["total"].(float64) != 1 {
		t.Fatalf("filtered log %v", l)
	}
	if code, _ := h.api(http.MethodGet, "/friend-gifts?status=bogus", nil); code != 400 {
		t.Fatal("bad filter")
	}
	if code, out := h.api(http.MethodPost, "/friend-gifts/"+gi["id"].(string)+"/reverse", map[string]any{"returnTo": "refund"}); code != 400 || out["error"] != "INVALID_POLICY" {
		t.Fatalf("refund an item %d %v", code, out)
	}
	r := h.mustAPI(200, http.MethodPost, "/friend-gifts/"+gi["id"].(string)+"/reverse", map[string]any{"returnTo": "sender", "note": "chargeback"})
	if r["status"] != "reversed" || r["outcome"] != "returned" || r["reverseNote"] != "chargeback" || !h.ownsItem(a, "orb") || h.ownsItem(b, "orb") {
		t.Fatalf("reverse item %v", r)
	}
	if code, out := h.api(http.MethodPost, "/friend-gifts/"+gi["id"].(string)+"/reverse", map[string]any{}); code != 409 || out["error"] != "GIFT_NOT_REVERSIBLE" {
		t.Fatalf("reverse twice %d %v", code, out)
	}
	h.mustAPI(200, http.MethodPost, "/friend-gifts/"+gc["id"].(string)+"/reverse", map[string]any{"returnTo": "none"})
	if h.wallet(b)["coin"] != 100 || h.wallet(a)["coin"] != 200 {
		t.Fatalf("currency reversed to nobody %v %v", h.wallet(a), h.wallet(b))
	}
	r = h.mustAPI(200, http.MethodPost, "/friend-gifts/"+gp["id"].(string)+"/reverse", map[string]any{"returnTo": "refund"})
	if r["outcome"] != "refunded" || h.wallet(b)["coin"] != 0 || h.wallet(a)["gem"] != 500 {
		t.Fatalf("product reversed with a refund %v %v %v", r, h.wallet(a), h.wallet(b))
	}
	if len(r["trail"].([]any)) == 0 {
		t.Fatal("audit trail")
	}
	// pending gift reversed
	h.giftSettings(map[string]any{"delivery": "accept"})
	gpend := h.send(a, map[string]any{"friendId": b, "itemKey": "orb"})
	h.mustAPI(200, http.MethodPost, "/friend-gifts/"+gpend["id"].(string)+"/reverse", map[string]any{"returnTo": "sender"})
	if !h.ownsItem(a, "orb") {
		t.Fatal("pending reverse returns the item")
	}
	rep := h.mustAPI(200, http.MethodGet, "/friend-gift-report?days=7", nil)
	tot := rep["totals"].(map[string]any)
	if tot["sent"].(float64) != 4 || rep["statuses"].(map[string]any)["reversed"].(float64) != 4 || len(rep["topSenders"].([]any)) != 1 {
		t.Fatalf("report %v", rep)
	}

	// settings validation
	for _, c := range []struct {
		change map[string]any
		field  string
	}{
		{map[string]any{"delivery": "soon"}, "delivery"},
		{map[string]any{"expiryHours": 0}, "expiryHours"},
		{map[string]any{"methods": []string{"cash"}}, "methods"},
		{map[string]any{"currencies": []map[string]any{{"currencyKey": "nope"}}}, "currencies[0].currencyKey"},
		{map[string]any{"currencies": []map[string]any{{"currencyKey": "coin", "min": 5, "max": 2}}}, "currencies[0].max"},
		{map[string]any{"itemTypes": []map[string]any{{"typeId": uuid.Must(uuid.NewV4()).String()}}}, "itemTypes[0].typeId"},
		{map[string]any{"sendBonusMeta": "nope"}, "sendBonusMeta"},
		{map[string]any{"notifyChannel": "sms"}, "notifyChannel"},
	} {
		cur := h.mustAPI(200, http.MethodGet, "/friend-gift-settings", nil)["settings"].(map[string]any)
		for k, v := range c.change {
			cur[k] = v
		}
		if code, out := h.api(http.MethodPut, "/friend-gift-settings", cur); code != 400 || out["field"] != c.field {
			t.Fatalf("%v = %d %v", c.change, code, out)
		}
	}
	// message filter reuses the chat word list
	if _, err := h.s.db.Exec(`INSERT INTO base_chat_word (id, word, action, match) VALUES (gen_random_uuid(), 'zzforbidden', 'block', 'word') ON CONFLICT DO NOTHING`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = h.s.db.Exec(`DELETE FROM base_chat_word WHERE word='zzforbidden'`) })
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 1, "message": "zzforbidden"}); got != "MESSAGE_BLOCKED" {
		t.Fatalf("blocked word = %s", got)
	}
	h.giftSettings(map[string]any{"messageMaxLength": 5})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 1, "message": "too long"}); got != "MESSAGE_TOO_LONG" {
		t.Fatalf("long = %s", got)
	}
	h.giftSettings(map[string]any{"enabled": false})
	if got := h.rpcCode(a, "economy_friend_gift_send", map[string]any{"friendId": b, "currencyKey": "coin", "amount": 1}); got != "FRIEND_GIFTS_OFF" {
		t.Fatalf("off = %s", got)
	}
}
