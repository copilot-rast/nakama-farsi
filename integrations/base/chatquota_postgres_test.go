package main

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

func resetChatQuota(h *harness) {
	h.t.Helper()
	for _, q := range []string{
		`DELETE FROM base_chat_quota_rule`, `DELETE FROM base_chat_quota_lot`, `DELETE FROM base_chat_quota_day`, `DELETE FROM base_chat_quota_target`,
		`UPDATE base_chat_quota_settings SET settings='{}'::jsonb WHERE id=1`,
	} {
		if _, err := h.s.db.Exec(q); err != nil {
			h.t.Fatal(err)
		}
	}
	h.t.Cleanup(func() { _, _ = h.s.db.Exec(`UPDATE base_chat_quota_settings SET settings='{}'::jsonb WHERE id=1`) })
}

func (h *harness) chatSettings(patch map[string]any) map[string]any {
	h.t.Helper()
	body := map[string]any{"enabled": true}
	for k, v := range patch {
		body[k] = v
	}
	return h.mustAPI(200, "PUT", "/chat-quota/settings", body)
}

func dmChannel(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return "4." + a + "." + b + "."
}

func (h *harness) sendTo(user, channel string) (*chatQuotaError, error) {
	h.t.Helper()
	err := h.s.chatQuotaGate(asUser(user), user, channel, `{"text":"hi"}`)
	if err == nil {
		return nil, nil
	}
	var rt *runtime.Error
	if !errors.As(err, &rt) {
		return nil, err
	}
	e := &chatQuotaError{}
	if json.Unmarshal([]byte(rt.Message), e) != nil {
		return nil, err
	}
	return e, nil
}

func (h *harness) dmOK(user, peer string, times int) {
	h.t.Helper()
	for i := 0; i < times; i++ {
		e, err := h.sendTo(user, dmChannel(user, peer))
		if err != nil || e != nil {
			h.t.Fatalf("send %d from %s refused: %+v %v", i+1, user, e, err)
		}
	}
}

func (h *harness) dmRefused(code, user, peer string) *chatQuotaError {
	h.t.Helper()
	e, err := h.sendTo(user, dmChannel(user, peer))
	if err != nil || e == nil || e.Code != code {
		h.t.Fatalf("want %s, got %+v %v", code, e, err)
	}
	if e.Message == "" || e.Lang == "" || e.ResetsAt == "" {
		h.t.Fatalf("refusal without message or reset: %+v", e)
	}
	return e
}

func (h *harness) packProduct(metaKey string, amount int) (string, string) {
	h.t.Helper()
	k := "cq" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	prize, product := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	items, _ := json.Marshal([]map[string]any{{"type": "meta", "key": metaKey, "value": amount, "action": "add"}})
	if _, err := h.s.db.Exec(`INSERT INTO economy_prize (id, key, name, items) VALUES ($1,$2,'{"fa":"بسته"}',$3)`, prize, k, string(items)); err != nil {
		h.t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO economy_product (id, key, name, prize_id, is_active, price_toman) VALUES ($1,$2,'{"fa":"بستهٔ پیام","en":"Message pack"}',$3,true,50000)`, product, k, prize); err != nil {
		h.t.Fatal(err)
	}
	h.t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DELETE FROM economy_payment WHERE product_id=$1`, product)
		_, _ = h.s.db.Exec(`DELETE FROM economy_product WHERE id=$1`, product)
		_, _ = h.s.db.Exec(`DELETE FROM economy_prize WHERE id=$1`, prize)
	})
	return product, k
}

func (h *harness) setMeta(user, metaID string, v float64) {
	h.t.Helper()
	if _, err := h.s.db.Exec(`INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ($1,$2,$3) ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value, updated_at=now()`, user, metaID, v); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) chatDayRow(user string) chatDay {
	h.t.Helper()
	st, _ := loadChatQuotaSettings(h.ctx, h.s.db)
	day, _, _ := h.s.chatPeriod(st)
	var d chatDay
	if err := d.scan(h.s.db.QueryRow(`SELECT `+chatDayCols+` FROM base_chat_quota_day WHERE user_id=$1 AND day=$2::date`, user, day)); err != nil {
		h.t.Fatal(err)
	}
	return d
}

func num(v any) int {
	f, _ := v.(float64)
	return int(f)
}

func TestChatQuotaCountingAndReset(t *testing.T) {
	h := newHarness(t)
	resetChatQuota(h)
	a, b, c := h.user("fa"), h.user("en"), h.user("fa")

	h.dmOK(a, b, 30)
	if out := h.mustRPC(a, "base_chat_quota", map[string]any{}); out["enabled"] != false {
		t.Fatalf("off by default: %v", out)
	}
	var n int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM base_chat_quota_day WHERE user_id=$1`, a).Scan(&n)
	if n != 0 {
		t.Fatal("nothing is tracked while off")
	}

	h.chatSettings(map[string]any{"dailyMessages": 3})
	h.dmOK(a, b, 2)
	h.dmOK(a, c, 1)
	e := h.dmRefused(CodeChatQuota, a, b)
	if e.Unit != "messages" || e.Limit != 3 || e.Used != 3 || e.Remaining != 0 || e.PacksLeft != 0 || e.Lang != "fa" || !strings.Contains(e.Message, "پیام") || len(e.ProductIDs) != 0 {
		t.Fatalf("refusal: %+v", e)
	}
	if _, err := time.Parse(time.RFC3339, e.ResetsAt); err != nil {
		t.Fatal(err)
	}
	e = h.dmRefused(CodeChatQuota, a, c)
	h.dmOK(b, a, 3)
	if e := h.dmRefused(CodeChatQuota, b, a); e.Lang != "en" || !strings.Contains(e.Message, "private messages") {
		t.Fatalf("english refusal: %+v", e)
	}
	d := h.chatDayRow(a)
	if d.Messages != 3 || d.Total != 3 || d.Blocked != 2 {
		t.Fatalf("day row: %+v", d)
	}

	room := "2...lobby" + a[:4]
	group := "3." + uuid.Must(uuid.NewV4()).String() + ".."
	for _, ch := range []string{room, group} {
		if e, err := h.sendTo(a, ch); e != nil || err != nil {
			t.Fatalf("%s is not counted by default: %+v %v", ch, e, err)
		}
	}
	h.chatSettings(map[string]any{"dailyMessages": 3, "countRooms": true})
	if e, _ := h.sendTo(a, room); e == nil || e.Code != CodeChatQuota {
		t.Fatalf("rooms counted when on: %+v", e)
	}
	if e, _ := h.sendTo(a, group); e != nil {
		t.Fatal("groups still free")
	}
	if e, _ := h.sendTo(a, "4.x.y."); e != nil {
		t.Fatal("a bad channel id is left to Nakama")
	}
	if err := h.s.chatQuotaGate(asUser(a), a, dmChannel(a, b), `not json`); err != nil {
		t.Fatal("bad content is left to Nakama")
	}

	out := h.mustRPC(a, "base_chat_quota", map[string]any{"lang": "en"})
	msgs := out["messages"].(map[string]any)
	if out["enabled"] != true || out["unlimited"] != false || num(out["freeLeft"]) != 0 || num(msgs["used"]) != 3 || num(msgs["limit"]) != 3 || out["timeZone"] != "Asia/Tehran" || out["rule"] != nil {
		t.Fatalf("status: %v", out)
	}
	resetAt, _ := time.Parse(time.RFC3339, out["resetsAt"].(string))
	if resetAt.Sub(time.Now()) <= 0 || resetAt.Sub(time.Now()) > 24*time.Hour {
		t.Fatalf("reset at %v", resetAt)
	}

	h.s.now = func() time.Time { return resetAt.Add(time.Minute) }
	h.dmOK(a, b, 3)
	h.dmRefused(CodeChatQuota, a, b)
	h.s.now = time.Now

	h.chatSettings(map[string]any{"dailyMessages": 3, "resetTime": "07:30", "timeZone": "UTC"})
	out = h.mustRPC(a, "base_chat_quota", map[string]any{})
	r2, _ := time.Parse(time.RFC3339, out["resetsAt"].(string))
	if r2.UTC().Hour() != 7 || r2.UTC().Minute() != 30 || out["timeZone"] != "UTC" {
		t.Fatalf("admin reset time: %v", out["resetsAt"])
	}

	h.expectError(400, "INVALID_TIMEZONE", "PUT", "/chat-quota/settings", map[string]any{"timeZone": "Nowhere/City"})
	h.expectError(400, "INVALID_SETTING", "PUT", "/chat-quota/settings", map[string]any{"resetTime": "25:00"})
	h.expectError(400, "INVALID_SETTING", "PUT", "/chat-quota/settings", map[string]any{"countMode": "letters"})
	h.expectError(400, "META_NOT_FOUND", "PUT", "/chat-quota/settings", map[string]any{"packMeta": "nope_" + a[:6]})
	h.expectError(400, "ITEM_NOT_FOUND", "PUT", "/chat-quota/settings", map[string]any{"unlimitedItems": []string{"nope_" + a[:6]}})
	h.expectError(400, "PRODUCT_NOT_FOUND", "PUT", "/chat-quota/settings", map[string]any{"packProducts": []string{"nope_" + a[:6]}})
	got := h.mustAPI(200, "GET", "/chat-quota/settings", nil)
	if got["resetTime"] != "07:30" || got["enabled"] != true || got["countMode"] != "messages" {
		t.Fatalf("settings round trip: %v", got)
	}
}

func TestChatQuotaPacksOrderExpiryAndAdmin(t *testing.T) {
	h := newHarness(t)
	resetChatQuota(h)
	a, b := h.user("fa"), h.user("fa")
	key := "cqpack" + strings.ReplaceAll(a[:8], "-", "")
	metaID := h.meta(key)
	h.expectError(400, "SAME_META", "PUT", "/chat-quota/settings", map[string]any{"packMeta": key, "bonusMeta": key})
	product, productKey := h.packProduct(key, 10)
	set := h.chatSettings(map[string]any{"dailyMessages": 2, "packMeta": key})
	if det, _ := set["detectedProducts"].([]any); len(det) != 1 || num(det[0].(map[string]any)["amount"]) != 10 {
		t.Fatalf("pack product found from its prize: %v", set["detectedProducts"])
	}

	h.setMeta(a, metaID, 3.5)
	h.dmOK(a, b, 2)
	d := h.chatDayRow(a)
	if d.Messages != 2 || d.PackUsed != 0 {
		t.Fatalf("free first: %+v", d)
	}
	h.dmOK(a, b, 3)
	e := h.dmRefused(CodeChatQuota, a, b)
	if e.PacksLeft != 0 || len(e.ProductIDs) != 1 || e.ProductIDs[0] != product || e.Products[0].Amount != 10 || e.Products[0].Name != "بستهٔ پیام" {
		t.Fatalf("refusal lists the pack products: %+v", e)
	}
	var left float64
	_ = h.s.db.QueryRow(`SELECT value::float8 FROM economy_user_meta WHERE user_id=$1 AND meta_id=$2`, a, metaID).Scan(&left)
	if left != 0.5 {
		t.Fatalf("whole units moved into a lot, %v left on the meta", left)
	}
	if d := h.chatDayRow(a); d.PackUsed != 3 || d.Total != 5 {
		t.Fatalf("packs used: %+v", d)
	}

	h.chatSettings(map[string]any{"dailyMessages": 2, "packMeta": key, "order": "packs_first", "packExpiryDays": 7})
	h.setMeta(b, metaID, 2)
	h.dmOK(b, a, 2)
	if d := h.chatDayRow(b); d.PackUsed != 2 || d.Messages != 0 {
		t.Fatalf("packs first: %+v", d)
	}
	var exp time.Time
	_ = h.s.db.QueryRow(`SELECT expires_at FROM base_chat_quota_lot WHERE user_id=$1`, b).Scan(&exp)
	if d := time.Until(exp); d < 6*24*time.Hour || d > 8*24*time.Hour {
		t.Fatalf("pack expires in 7 days from the purchase: %v", exp)
	}
	h.dmOK(b, a, 2)
	h.dmRefused(CodeChatQuota, b, a)

	h.mustAPI(200, "POST", "/players/"+a+"/chat-quota/grant", map[string]any{"amount": 2, "expiresInDays": 0, "note": "e2e جبران"})
	card := h.mustAPI(200, "POST", "/players/"+a+"/chat-quota/grant", map[string]any{"amount": 1, "expiresInDays": 1})
	if num(card["packsLeft"]) != 3 {
		t.Fatalf("granted packs: %v", card["packsLeft"])
	}
	if _, err := h.s.db.Exec(`UPDATE base_chat_quota_lot SET expires_at = now() - INTERVAL '1 minute' WHERE user_id=$1 AND expires_at IS NOT NULL AND remaining > 0`, a); err != nil {
		t.Fatal(err)
	}
	h.dmOK(a, b, 2)
	h.dmRefused(CodeChatQuota, a, b)
	card = h.mustAPI(200, "GET", "/players/"+a+"/chat-quota", nil)
	lots := card["lots"].([]any)
	if num(card["packsLeft"]) != 0 || len(lots) != 3 || len(card["days"].([]any)) != 1 {
		t.Fatalf("player card: %v", card)
	}
	h.mustAPI(200, "POST", "/players/"+a+"/chat-quota/grant", map[string]any{"amount": 5})
	card = h.mustAPI(200, "GET", "/players/"+a+"/chat-quota", nil)
	var lotID string
	for _, l := range card["lots"].([]any) {
		m := l.(map[string]any)
		if m["active"] == true {
			lotID = m["id"].(string)
		}
	}
	card = h.mustAPI(200, "POST", "/players/"+a+"/chat-quota/lots/"+lotID+"/revoke", nil)
	if num(card["packsLeft"]) != 0 {
		t.Fatal("revoked lot")
	}
	h.expectError(404, "LOT_NOT_FOUND", "POST", "/players/"+b+"/chat-quota/lots/"+lotID+"/revoke", nil)
	h.expectError(400, "INVALID_AMOUNT", "POST", "/players/"+a+"/chat-quota/grant", map[string]any{"amount": 0})
	h.expectError(404, "PLAYER_NOT_FOUND", "GET", "/players/"+uuid.Must(uuid.NewV4()).String()+"/chat-quota", nil)
	card = h.mustAPI(200, "POST", "/players/"+a+"/chat-quota/reset", nil)
	if num(card["freeLeft"]) != 2 {
		t.Fatalf("reset today: %v", card["freeLeft"])
	}
	h.dmOK(a, b, 2)

	if _, err := h.s.db.Exec(`INSERT INTO economy_payment (id, user_id, product_id, method, status, final_toman) VALUES ($1,$2,$3,'zarinpal','success',50000)`, uuid.Must(uuid.NewV4()).String(), a, product); err != nil {
		t.Fatal(err)
	}
	rep := h.mustAPI(200, "GET", "/chat-quota/report?days=7", nil)
	tot := rep["totals"].(map[string]any)
	packs := rep["packs"].(map[string]any)
	if num(tot["blocked"]) < 3 || num(tot["fromPacks"]) != 7 || num(tot["sent"]) < 13 || num(packs["sold"]) != 1 || num(packs["toman"]) != 50000 || num(packs["granted"].(map[string]any)["admin"]) != 8 || num(packs["granted"].(map[string]any)["meta"]) != 5 {
		t.Fatalf("report: %v", rep)
	}
	if top := rep["topBlocked"].([]any); len(top) == 0 {
		t.Fatal("top blocked players")
	}
	if prods := packs["products"].([]any); len(prods) != 1 || prods[0].(map[string]any)["key"] != productKey {
		t.Fatalf("pack sales per product: %v", prods)
	}

	h.chatSettings(map[string]any{"dailyMessages": 2, "packMeta": key, "packProducts": []string{productKey}})
	if err := h.s.chatQuotaSweep(h.ctx, "test"); err != nil {
		t.Fatal(err)
	}
}

func TestChatQuotaUnlimitedRulesItemsAndMetas(t *testing.T) {
	h := newHarness(t)
	resetChatQuota(h)
	suffix := strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	vip, pro, both, plain, peer := h.user("fa"), h.user("fa"), h.user("fa"), h.user("fa"), h.user("fa")
	vipTag := h.userTag("cqvip"+suffix, vip, both)
	proTag := h.userTag("cqpro"+suffix, pro, both)
	h.chatSettings(map[string]any{"dailyMessages": 1})

	h.expectError(400, "LIMIT_REQUIRED", "POST", "/chat-quota/rules", map[string]any{"name": "x", "tagIds": []string{vipTag}})
	h.expectError(400, "TAGS_REQUIRED", "POST", "/chat-quota/rules", map[string]any{"name": "x", "unlimited": true})
	h.expectError(400, "INVALID_RULE", "POST", "/chat-quota/rules", map[string]any{"name": "x", "tagIds": []string{vipTag}, "dailyMessages": -1})
	vipRule := h.mustAPI(201, "POST", "/chat-quota/rules", map[string]any{"name": "VIP", "tagIds": []string{vipTag}, "unlimited": true, "priority": 1})
	proRule := h.mustAPI(201, "POST", "/chat-quota/rules", map[string]any{"name": "Pro", "tagIds": []string{proTag}, "dailyMessages": 3, "priority": 5})
	if vipRule["unlimited"] != true || num(proRule["players"]) != 2 || num(proRule["dailyMessages"]) != 3 {
		t.Fatalf("rules: %v %v", vipRule, proRule)
	}
	list := h.mustAPI(200, "GET", "/chat-quota/rules", nil)
	if items(list)[0]["name"] != "Pro" {
		t.Fatal("rules by priority")
	}

	h.dmOK(vip, peer, 50)
	if d := h.chatDayRow(vip); d.UnlimitedSent != 50 || d.Messages != 0 {
		t.Fatalf("unlimited is not counted: %+v", d)
	}
	h.dmOK(pro, peer, 3)
	h.dmRefused(CodeChatQuota, pro, peer)
	h.dmOK(both, peer, 3)
	e := h.dmRefused(CodeChatQuota, both, peer)
	if e.Rule["name"] != "Pro" || e.Limit != 3 {
		t.Fatalf("priority picks Pro: %+v", e)
	}
	h.chatSettings(map[string]any{"dailyMessages": 1, "matchPolicy": "generous"})
	h.dmOK(both, peer, 5)
	out := h.mustRPC(both, "base_chat_quota", map[string]any{})
	if out["unlimited"] != true || out["unlimitedBy"] != "rule" || out["rule"].(map[string]any)["name"] != "VIP" {
		t.Fatalf("generous gives unlimited: %v", out)
	}
	h.mustAPI(200, "PUT", "/chat-quota/rules/"+vipRule["id"].(string), map[string]any{"name": "VIP", "tagIds": []string{vipTag}, "unlimited": true, "isActive": false})
	h.dmOK(vip, peer, 1)
	h.dmRefused(CodeChatQuota, vip, peer)
	h.mustAPI(200, "DELETE", "/chat-quota/rules/"+proRule["id"].(string), nil)
	h.expectError(404, "NOT_FOUND", "GET", "/chat-quota/rules/"+proRule["id"].(string), nil)

	h.dmOK(plain, peer, 1)
	h.dmRefused(CodeChatQuota, plain, peer)
	typeID := h.mustAPI(201, "POST", "/categories", map[string]any{"key": "cqtype" + suffix, "keyType": "item_type", "name": map[string]string{"fa": "نوع"}})["id"].(string)
	itemID := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.Exec(`INSERT INTO economy_item (id, key, type_id, name) VALUES ($1,$2,$3,'{"fa":"تلفن طلایی"}')`, itemID, "cqitem"+suffix, typeID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DELETE FROM economy_user_item WHERE item_id=$1`, itemID)
		_, _ = h.s.db.Exec(`DELETE FROM economy_item WHERE id=$1`, itemID)
	})
	h.chatSettings(map[string]any{"dailyMessages": 1, "unlimitedItems": []string{"cqitem" + suffix}})
	h.dmRefused(CodeChatQuota, plain, peer)
	if _, err := h.s.db.Exec(`INSERT INTO economy_user_item (user_id, item_id, source) VALUES ($1,$2,'admin')`, plain, itemID); err != nil {
		t.Fatal(err)
	}
	h.dmOK(plain, peer, 3)
	if out := h.mustRPC(plain, "base_chat_quota", map[string]any{}); out["unlimitedBy"] != "item" {
		t.Fatalf("item: %v", out["unlimitedBy"])
	}

	subKey := "cqsub" + suffix
	subID := h.meta(subKey)
	h.chatSettings(map[string]any{"dailyMessages": 1, "unlimitedMeta": subKey, "unlimitedMetaMode": "until"})
	h.dmRefused(CodeChatQuota, pro, peer)
	h.setMeta(pro, subID, float64(time.Now().Add(time.Hour).Unix()))
	h.dmOK(pro, peer, 2)
	h.setMeta(pro, subID, float64(time.Now().Add(-time.Hour).Unix()))
	h.dmRefused(CodeChatQuota, pro, peer)
	h.chatSettings(map[string]any{"dailyMessages": 1, "unlimitedMeta": subKey, "unlimitedMetaMode": "positive"})
	h.dmOK(pro, peer, 2)

	bonusKey := "cqbonus" + suffix
	bonusID := h.meta(bonusKey)
	h.chatSettings(map[string]any{"dailyMessages": 1, "bonusMeta": bonusKey})
	h.setMeta(peer, bonusID, 2)
	h.dmOK(peer, plain, 3)
	h.dmRefused(CodeChatQuota, peer, plain)
	if out := h.mustRPC(peer, "base_chat_quota", map[string]any{}); num(out["bonus"]) != 2 {
		t.Fatalf("bonus meta: %v", out["bonus"])
	}
}

func TestChatQuotaConversationsFriendsAndReplies(t *testing.T) {
	h := newHarness(t)
	resetChatQuota(h)
	a, b, c, d, f := h.user("fa"), h.user("fa"), h.user("fa"), h.user("fa"), h.user("fa")

	h.chatSettings(map[string]any{"countMode": "conversations", "dailyConversations": 2})
	h.dmOK(a, b, 5)
	h.dmOK(a, c, 5)
	e := h.dmRefused(CodeChatQuota, a, d)
	if e.Unit != "conversations" || e.Limit != 2 {
		t.Fatalf("conversation refusal: %+v", e)
	}
	h.dmOK(a, b, 2)
	h.s.now = func() time.Time { return time.Now().Add(25 * time.Hour) }
	h.dmOK(a, d, 1)
	h.dmOK(a, b, 1)
	h.dmRefused(CodeChatQuota, a, c)
	h.s.now = time.Now

	h.chatSettings(map[string]any{"countMode": "conversations", "dailyConversations": 1, "conversationScope": "ever"})
	h.dmOK(b, c, 1)
	h.dmRefused(CodeChatQuota, b, d)
	h.s.now = func() time.Time { return time.Now().Add(25 * time.Hour) }
	h.dmOK(b, c, 3)
	h.dmOK(b, d, 1)
	h.s.now = time.Now

	resetChatQuota(h)
	h.chatSettings(map[string]any{"countMode": "both", "dailyMessages": 3, "dailyConversations": 1})
	h.dmOK(c, d, 2)
	e = h.dmRefused(CodeChatConversationLimit, c, a)
	if e.Unit != "conversations" || !strings.Contains(e.Message, "۱") {
		t.Fatalf("conversation cap in both mode: %+v", e)
	}
	h.dmOK(c, d, 1)
	h.dmRefused(CodeChatQuota, c, d)

	resetChatQuota(h)
	h.befriend(a, f)
	h.chatSettings(map[string]any{"dailyMessages": 1, "friendPolicy": "separate", "friendDailyMessages": 3})
	h.dmOK(a, f, 3)
	h.dmRefused(CodeChatQuota, a, f)
	h.dmOK(a, b, 1)
	e = h.dmRefused(CodeChatQuota, a, b)
	if e.Friend || e.Limit != 1 {
		t.Fatalf("stranger pool: %+v", e)
	}
	out := h.mustRPC(a, "base_chat_quota", map[string]any{})
	fr := out["friends"].(map[string]any)["messages"].(map[string]any)
	if num(fr["used"]) != 3 || num(fr["limit"]) != 3 {
		t.Fatalf("friend pool in status: %v", out["friends"])
	}
	h.chatSettings(map[string]any{"dailyMessages": 1, "friendPolicy": "free"})
	h.dmOK(a, f, 10)
	if d := h.chatDayRow(a); d.FreeSent != 10 {
		t.Fatalf("free to friends: %+v", d)
	}

	resetChatQuota(h)
	h.chatSettings(map[string]any{"dailyMessages": 1, "replyFree": "started"})
	h.dmOK(b, c, 1)
	h.dmOK(c, b, 4)
	h.dmRefused(CodeChatQuota, b, c)
	h.chatSettings(map[string]any{"dailyMessages": 1, "replyFree": "reply"})
	h.dmOK(d, a, 1)
	h.dmOK(a, d, 1)
	if d := h.chatDayRow(a); d.FreeSent != 1 || d.Messages != 0 {
		t.Fatalf("a reply is free: %+v", d)
	}
	h.dmOK(a, d, 1)
	h.dmRefused(CodeChatQuota, a, d)
	h.dmOK(d, a, 1)
	h.dmRefused(CodeChatQuota, d, a)
	h.chatSettings(map[string]any{"dailyMessages": 1, "replyFree": "off"})
	h.dmOK(c, b, 1)
	h.dmRefused(CodeChatQuota, c, b)
}

func TestChatQuotaParallelSendsAndNotice(t *testing.T) {
	h := newHarness(t)
	resetChatQuota(h)
	a, b := h.user("fa"), h.user("fa")
	h.chatSettings(map[string]any{"dailyMessages": 5, "notice": "inbox", "message": map[string]string{"fa": "سهمیه تمام شد؛ تا {reset} صبر کن.", "en": ""}})
	h.mustAPI(200, "POST", "/players/"+a+"/chat-quota/grant", map[string]any{"amount": 3})
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, refused := 0, 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e, err := h.sendTo(a, dmChannel(a, b))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				t.Error(err)
				return
			}
			if e == nil {
				ok++
			} else {
				refused++
			}
		}()
	}
	wg.Wait()
	if ok != 8 || refused != 12 {
		t.Fatalf("parallel: %d sent, %d refused", ok, refused)
	}
	d := h.chatDayRow(a)
	if d.Total != 8 || d.Messages != 5 || d.PackUsed != 3 || d.Blocked != 12 || !d.Notified {
		t.Fatalf("day row after parallel sends: %+v", d)
	}
	var rem int
	_ = h.s.db.QueryRow(`SELECT sum(remaining) FROM base_chat_quota_lot WHERE user_id=$1`, a).Scan(&rem)
	if rem != 0 {
		t.Fatalf("lots left %d", rem)
	}
	e := h.dmRefused(CodeChatQuota, a, b)
	if !strings.HasPrefix(e.Message, "سهمیه تمام شد؛ تا ") || strings.Contains(e.Message, "{reset}") {
		t.Fatalf("admin message: %q", e.Message)
	}
	fake := h.s.nk.(*fakeNakama)
	fake.mu.Lock()
	notes := 0
	for _, n := range fake.notes {
		if n.user == a && n.code == notifyChatQuota {
			notes++
			if !n.persistent || n.content["code"] != CodeChatQuota {
				t.Errorf("notice: %+v", n)
			}
		}
	}
	fake.mu.Unlock()
	if notes != 1 {
		t.Fatalf("one notice per day, got %d", notes)
	}
}
