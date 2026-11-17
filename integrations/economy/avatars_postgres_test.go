package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

func newAvatarHarness(t *testing.T) *harness {
	h, _ := newItemHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_avatar_equip, economy_avatar, economy_avatar_unlock, economy_avatar_item, economy_avatar_category, economy_avatar_type, economy_avatar_settings,
		base_avatar_look, base_avatar_option, base_avatar_part, base_avatar_settings CASCADE`); err != nil {
		t.Fatal(err)
	}
	for _, rpcs := range []map[string]clientRPC{h.s.avatarRPCs(), h.s.legacyAvatarRPCs()} {
		for name, fn := range rpcs {
			h.rpcs[name] = fn
		}
	}
	return h
}

func (h *harness) avCategory(key string, extra map[string]any) string {
	h.t.Helper()
	body := map[string]any{"key": key, "name": map[string]string{"fa": key, "en": key}}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/avatar-categories", body)["id"].(string)
}

func (h *harness) avItem(cat, key string, extra map[string]any) map[string]any {
	h.t.Helper()
	body := map[string]any{"categoryId": cat, "key": key, "name": map[string]string{"fa": key, "en": key}, "imageUrl": "https://cdn.example.com/" + key + ".png"}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/avatar-items", body)
}

func (h *harness) setMetaValue(user, metaID string, v float64) {
	h.t.Helper()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ($1,$2,$3) ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value`, user, metaID, v); err != nil {
		h.t.Fatal(err)
	}
}

func catalogItems(out map[string]any) map[string]map[string]any {
	items := map[string]map[string]any{}
	for _, c := range out["categories"].([]any) {
		for _, raw := range c.(map[string]any)["items"].([]any) {
			it := raw.(map[string]any)
			items[it["category"].(string)+"/"+it["key"].(string)] = it
		}
	}
	return items
}

func lookOf(v any) map[string]any {
	return v.(map[string]any)["look"].(map[string]any)["items"].(map[string]any)
}

func activeAvatar(out map[string]any) map[string]any {
	for _, raw := range out["avatars"].([]any) {
		a := raw.(map[string]any)
		if a["isActive"] == true {
			return a
		}
	}
	return nil
}

func TestAvatarAdminValidation(t *testing.T) {
	h := newAvatarHarness(t)
	hair := h.avCategory("hair", map[string]any{"layer": 3})
	if code, out := h.api(http.MethodPost, "/avatar-categories", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "slots": 0}); code != 400 || out["error"] != "INVALID_SLOTS" {
		t.Fatalf("slots = %d %v", code, out)
	}
	cases := []struct {
		body  map[string]any
		code  string
		field string
	}{
		{map[string]any{"categoryId": hair, "key": "nolook", "name": map[string]string{"fa": "x"}}, "OPTION_LOOK_REQUIRED", "imageUrl"},
		{map[string]any{"categoryId": hair, "key": "p1", "value": "#000", "access": "currency"}, "AVATAR_PRICE_REQUIRED", "price.currencyKey"},
		{map[string]any{"categoryId": hair, "key": "p2", "value": "#000", "access": "currency", "price": map[string]any{"currencyKey": "nope", "amount": 5}}, "CURRENCY_NOT_FOUND", "price.currencyKey"},
		{map[string]any{"categoryId": hair, "key": "p3", "value": "#000", "access": "product"}, "REQUIRED", "productId"},
		{map[string]any{"categoryId": hair, "key": "p4", "value": "#000", "access": "item", "itemKey": "ghost"}, "ITEM_NOT_FOUND", "itemKey"},
		{map[string]any{"categoryId": hair, "key": "p5", "value": "#000", "access": "currency", "price": map[string]any{"currencyKey": "coin", "amount": 5}, "isDefault": true}, "DEFAULT_MUST_BE_FREE", "isDefault"},
		{map[string]any{"categoryId": hair, "key": "p6", "value": "#000", "access": "magic"}, "INVALID_ACCESS", "access"},
	}
	for _, c := range cases {
		if code, out := h.api(http.MethodPost, "/avatar-items", c.body); code != 400 || out["error"] != c.code || out["field"] != c.field {
			t.Fatalf("%v = %d %v, want %s on %s", c.body["key"], code, out, c.code, c.field)
		}
	}
	short := h.avItem(hair, "short", map[string]any{"isDefault": true})
	long := h.avItem(hair, "long", map[string]any{"isDefault": true})
	if got := h.mustAPI(200, http.MethodGet, "/avatar-items/"+short["id"].(string), nil); got["isDefault"] != false {
		t.Fatalf("a second default should clear the first: %v", got)
	}
	if code, out := h.api(http.MethodPost, "/avatar-items", map[string]any{"categoryId": hair, "key": "short", "value": "#111"}); code != 409 || out["error"] != "ALREADY_EXISTS" {
		t.Fatalf("duplicate = %d %v", code, out)
	}
	shoes := h.avCategory("shoes", nil)
	body := map[string]any{"categoryId": shoes, "key": "long", "name": map[string]string{"fa": "x"}, "imageUrl": "https://cdn.example.com/x.png"}
	if code, out := h.api(http.MethodPut, "/avatar-items/"+long["id"].(string), body); code != 400 || out["error"] != "CATEGORY_IMMUTABLE" {
		t.Fatalf("move = %d %v", code, out)
	}
	h.avItem(shoes, "gold", map[string]any{"access": "currency", "price": map[string]any{"currencyKey": "coin", "amount": 10}})
	if code, out := h.api(http.MethodPut, "/avatar-categories/"+shoes, map[string]any{"key": "shoes", "name": map[string]string{"fa": "کفش"}, "required": true}); code != 400 || out["error"] != "REQUIRED_NEEDS_FREE" {
		t.Fatalf("required without free = %d %v", code, out)
	}
	xp := h.meta("xp", 0)
	locked := h.avItem(hair, "mohawk", map[string]any{"requirements": map[string]any{"metas": []map[string]any{{"metaId": xp, "operator": "gte", "value": 300, "consume": 50}}, "paths": []string{"bot"}}})
	req := locked["requirements"].(map[string]any)
	if req["paths"] != nil || req["metas"].([]any)[0].(map[string]any)["consume"] != nil {
		t.Fatalf("requirements should drop paths and consume: %v", req)
	}
	if code, out := h.api(http.MethodPut, "/avatar-settings", map[string]any{"maxAvatars": 0}); code != 400 || out["field"] != "maxAvatars" {
		t.Fatalf("settings = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPut, "/avatar-settings", map[string]any{"maxAvatars": 1, "nameMaxLength": 24, "slotFull": "maybe"}); code != 400 || out["field"] != "slotFull" {
		t.Fatalf("settings = %d %v", code, out)
	}
	st := h.mustAPI(200, http.MethodGet, "/avatar-settings", nil)
	if st["enabled"] != true || st["maxAvatars"].(float64) != 1 || st["autoCreate"] != true || st["keepLast"] != true || st["slotFull"] != "replace" || st["keepBought"] != true {
		t.Fatalf("default settings = %v", st)
	}
}

func TestAvatarFreePaidAndMetaLocked(t *testing.T) {
	h := newAvatarHarness(t)
	xp := h.meta("xp", 0)
	hair := h.avCategory("hair", map[string]any{"layer": 3, "required": false})
	shirt := h.avCategory("shirt", map[string]any{"layer": 2})
	h.avItem(hair, "short", map[string]any{"isDefault": true})
	h.avItem(hair, "long", nil)
	h.avItem(shirt, "red", map[string]any{"access": "currency", "price": map[string]any{"currencyKey": "coin", "amount": 40}})
	h.avItem(shirt, "star", map[string]any{"requirements": map[string]any{"metas": []map[string]any{{"metaId": xp, "operator": "gte", "value": 300}}}})
	h.avItem(shirt, "secret", map[string]any{"requirements": map[string]any{"metas": []map[string]any{{"metaId": xp, "operator": "gte", "value": 1000}}, "lock": "hide"}})
	u := h.user()

	cat := h.mustRPC(u, "economy_avatar_catalog", map[string]any{"lang": "en"})
	items := catalogItems(cat)
	if cat["avatarId"] == "" || len(items) != 4 {
		t.Fatalf("catalog should auto-create an avatar and hide the hidden locked item: %v", items)
	}
	if items["hair/short"]["equipped"] != true || items["hair/short"]["usable"] != true || items["hair/long"]["equipped"] != false {
		t.Fatalf("default = %v", items["hair/short"])
	}
	red := items["shirt/red"]
	if red["owned"] != false || red["usable"] != false || red["price"].(map[string]any)["amount"].(float64) != 40 {
		t.Fatalf("paid = %v", red)
	}
	star := items["shirt/star"]
	miss := star["missing"].([]any)
	if star["locked"] != true || star["owned"] != true || star["usable"] != false || len(miss) != 1 || miss[0].(map[string]any)["key"] != "xp" || miss[0].(map[string]any)["value"].(float64) != 300 {
		t.Fatalf("locked = %v", star)
	}

	h.mustRPC(u, "economy_avatar_equip", map[string]any{"category": "hair", "key": "long"})
	if c := h.rpcCode(u, "economy_avatar_equip", map[string]any{"category": "shirt", "key": "red"}); c != "AVATAR_ITEM_NOT_OWNED" {
		t.Fatalf("paid equip = %s", c)
	}
	if c := h.rpcCode(u, "economy_avatar_equip", map[string]any{"category": "shirt", "key": "star"}); c != "AVATAR_ITEM_LOCKED" {
		t.Fatalf("locked equip = %s", c)
	}
	if c := h.rpcCode(u, "economy_avatar_buy", map[string]any{"category": "shirt", "key": "red"}); c != "INSUFFICIENT_BALANCE" {
		t.Fatalf("poor buy = %s", c)
	}
	if c := h.rpcCode(u, "economy_avatar_buy", map[string]any{"category": "hair", "key": "short"}); c != "AVATAR_ITEM_FREE" {
		t.Fatalf("free buy = %s", c)
	}
	h.give(u, 100)
	buy := h.mustRPC(u, "economy_avatar_buy", map[string]any{"category": "shirt", "key": "red", "requestId": "r1"})
	if buy["item"].(map[string]any)["owned"] != true || buy["wallet"].(map[string]any)["coin"].(float64) != 60 || buy["replayed"] != false {
		t.Fatalf("buy = %v", buy)
	}
	again := h.mustRPC(u, "economy_avatar_buy", map[string]any{"category": "shirt", "key": "red", "requestId": "r1"})
	if again["replayed"] != true || again["wallet"].(map[string]any)["coin"].(float64) != 60 {
		t.Fatalf("replay = %v", again)
	}
	if c := h.rpcCode(u, "economy_avatar_buy", map[string]any{"category": "shirt", "key": "red", "requestId": "r2"}); c != "AVATAR_ALREADY_OWNED" {
		t.Fatalf("second buy = %s", c)
	}
	list := h.mustRPC(u, "economy_avatar_equip", map[string]any{"category": "shirt", "key": "red"})
	look := lookOf(list["avatar"])
	if look["shirt"].([]any)[0] != "red" || look["hair"].([]any)[0] != "long" {
		t.Fatalf("look = %v", look)
	}

	h.setMetaValue(u, xp, 300)
	star = catalogItems(h.mustRPC(u, "economy_avatar_catalog", nil))["shirt/star"]
	if star["locked"] != false || star["usable"] != true {
		t.Fatalf("unlocked = %v", star)
	}
	out := h.mustRPC(u, "economy_avatar_equip", map[string]any{"category": "shirt", "key": "star"})
	if got := lookOf(out["avatar"])["shirt"].([]any); len(got) != 1 || got[0] != "star" {
		t.Fatalf("one slot should swap red for star: %v", got)
	}
	if n := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'reason'='avatar_buy'`, u); n != 1 {
		t.Fatalf("ledger rows = %d", n)
	}
	out = h.mustRPC(u, "economy_avatar_unequip", map[string]any{"category": "hair", "key": "long"})
	if _, has := lookOf(out["avatar"])["hair"]; has {
		t.Fatalf("hair should be empty after unequip: %v", lookOf(out["avatar"]))
	}
	if c := h.rpcCode(u, "economy_avatar_unequip", map[string]any{"category": "hair", "key": "long"}); c != "AVATAR_NOT_EQUIPPED" {
		t.Fatalf("second unequip = %s", c)
	}
}

func TestAvatarTypesSlotsCountAndLooks(t *testing.T) {
	h := newAvatarHarness(t)
	male := h.mustAPI(201, http.MethodPost, "/avatar-types", map[string]any{"key": "male", "name": map[string]string{"fa": "پسر"}, "sort": 1})["id"].(string)
	female := h.mustAPI(201, http.MethodPost, "/avatar-types", map[string]any{"key": "female", "name": map[string]string{"fa": "دختر"}, "sort": 2})["id"].(string)
	skin := h.avCategory("skin", map[string]any{"layer": 0, "required": true})
	h.avItem(skin, "light", map[string]any{"value": "#F1C27D", "imageUrl": "", "isDefault": true})
	h.avItem(skin, "dark", map[string]any{"value": "#8D5524", "imageUrl": ""})
	hat := h.avCategory("hat", map[string]any{"layer": 5})
	h.avItem(hat, "cap", nil)
	h.avItem(hat, "tiara", map[string]any{"typeIds": []string{female}})
	badge := h.avCategory("badge", map[string]any{"layer": 6, "slots": 2})
	for _, k := range []string{"b1", "b2", "b3"} {
		h.avItem(badge, k, nil)
	}
	vip := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})["id"].(string)
	u := h.user()

	first := h.mustRPC(u, "economy_avatar_list", map[string]any{"lang": "en"})
	a := activeAvatar(first)
	if a == nil || a["type"].(map[string]any)["key"] != "male" || first["limit"].(map[string]any)["max"].(float64) != 1 {
		t.Fatalf("first avatar = %v", first)
	}
	aID := a["id"].(string)
	if c := h.rpcCode(u, "economy_avatar_equip", map[string]any{"id": aID, "category": "hat", "key": "tiara"}); c != "AVATAR_TYPE_NOT_ALLOWED" {
		t.Fatalf("wrong type = %s", c)
	}
	if c := h.rpcCode(u, "economy_avatar_create", map[string]any{"name": "دوم", "type": "female"}); c != "AVATAR_LIMIT" {
		t.Fatalf("limit = %s", c)
	}
	h.mustAPI(200, http.MethodPut, "/avatar-settings", map[string]any{"enabled": true, "maxAvatars": 1, "nameMaxLength": 12, "autoCreate": true, "keepLast": true, "slotFull": "replace", "keepBought": true,
		"tagLimits": []map[string]any{{"tagId": vip, "maxAvatars": 2}}})
	if _, err := h.s.db.Exec(`INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, u, vip); err != nil {
		t.Fatal(err)
	}
	if c := h.rpcCode(u, "economy_avatar_create", map[string]any{"name": "a name that is too long", "type": "female"}); c != "INVALID_NAME" {
		t.Fatalf("name = %s", c)
	}
	if c := h.rpcCode(u, "economy_avatar_create", map[string]any{"name": "x", "type": "robot"}); c != "AVATAR_TYPE_NOT_FOUND" {
		t.Fatalf("type = %s", c)
	}
	second := h.mustRPC(u, "economy_avatar_create", map[string]any{"name": "سارا", "type": "female"})
	b := second["avatar"].(map[string]any)
	bID := b["id"].(string)
	if b["isActive"] != false || second["activeId"] != aID || second["limit"].(map[string]any)["canCreate"] != false {
		t.Fatalf("second = %v", second)
	}
	if c := h.rpcCode(u, "economy_avatar_create", map[string]any{"name": "سوم"}); c != "AVATAR_LIMIT" {
		t.Fatalf("vip limit = %s", c)
	}
	h.mustRPC(u, "economy_avatar_equip", map[string]any{"id": bID, "category": "hat", "key": "tiara"})
	h.mustRPC(u, "economy_avatar_equip", map[string]any{"id": bID, "category": "skin", "key": "dark"})
	h.mustRPC(u, "economy_avatar_equip", map[string]any{"id": aID, "category": "hat", "key": "cap"})
	for _, k := range []string{"b1", "b2", "b3"} {
		h.mustRPC(u, "economy_avatar_equip", map[string]any{"id": aID, "category": "badge", "key": k})
	}
	list := h.mustRPC(u, "economy_avatar_list", nil)
	var la, lb map[string]any
	for _, raw := range list["avatars"].([]any) {
		v := raw.(map[string]any)
		if v["id"] == aID {
			la = lookOf(v)
		} else {
			lb = lookOf(v)
		}
	}
	if la["hat"].([]any)[0] != "cap" || la["skin"].([]any)[0] != "light" || len(la["badge"].([]any)) != 2 || la["badge"].([]any)[0] != "b2" {
		t.Fatalf("avatar a = %v", la)
	}
	if lb["hat"].([]any)[0] != "tiara" || lb["skin"].([]any)[0] != "dark" {
		t.Fatalf("avatar b = %v", lb)
	}
	h.mustAPI(200, http.MethodPut, "/avatar-settings", map[string]any{"enabled": true, "maxAvatars": 1, "nameMaxLength": 12, "autoCreate": true, "keepLast": true, "slotFull": "refuse", "keepBought": true,
		"tagLimits": []map[string]any{{"tagId": vip, "maxAvatars": 2}}})
	if c := h.rpcCode(u, "economy_avatar_equip", map[string]any{"id": aID, "category": "badge", "key": "b1"}); c != "AVATAR_SLOT_FULL" {
		t.Fatalf("slot full = %s", c)
	}
	if c := h.rpcCode(u, "economy_avatar_unequip", map[string]any{"id": aID, "category": "skin", "key": "light"}); c != "AVATAR_CATEGORY_REQUIRED" {
		t.Fatalf("required = %s", c)
	}
	h.mustRPC(u, "economy_avatar_rename", map[string]any{"id": aID, "name": "  علی   "})
	act := h.mustRPC(u, "economy_avatar_set_active", map[string]any{"id": bID})
	if act["activeId"] != bID || act["avatar"].(map[string]any)["isActive"] != true {
		t.Fatalf("activate = %v", act)
	}
	other := h.user()
	looks := h.mustRPC(other, "economy_avatar_looks", map[string]any{"userIds": []string{u, other, "bot-1"}})["avatars"].(map[string]any)
	if lookOf(looks[u])["hat"].([]any)[0] != "tiara" || looks[u].(map[string]any)["type"].(map[string]any)["key"] != "female" {
		t.Fatalf("looks = %v", looks[u])
	}
	if lookOf(looks["bot-1"])["skin"].([]any)[0] != "light" {
		t.Fatalf("bot look = %v", looks["bot-1"])
	}
	if code := h.rpcCode(other, "economy_avatar_rename", map[string]any{"id": aID, "name": "x"}); code != "AVATAR_NOT_FOUND" {
		t.Fatalf("someone else's avatar = %s", code)
	}
	del := h.mustRPC(u, "economy_avatar_delete", map[string]any{"id": bID})
	if del["activeId"] != aID || len(del["avatars"].([]any)) != 1 || activeAvatar(del)["name"] != "علی" {
		t.Fatalf("delete active = %v", del)
	}
	if c := h.rpcCode(u, "economy_avatar_delete", map[string]any{"id": aID}); c != "AVATAR_LAST" {
		t.Fatalf("delete last = %s", c)
	}
	if code, out := h.api(http.MethodDelete, "/avatar-types/"+female, nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("type in use = %d %v", code, out)
	}
	view := h.mustAPI(200, http.MethodGet, "/users/"+u+"/avatars", nil)
	if len(view["avatars"].([]any)) != 1 || view["limit"].(map[string]any)["max"].(float64) != 2 {
		t.Fatalf("console view = %v", view)
	}
	_ = male
}

func TestAvatarProductAndShopItemAccess(t *testing.T) {
	h := newAvatarHarness(t)
	prize := h.coinPrize("pack_prize", 1, true)
	product := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "crown_pack", "isActive": true, "name": map[string]string{"fa": "بسته تاج"}, "prizeId": prize, "paymentMethods": []string{"zibal"}, "priceToman": 50000})["id"].(string)
	skinType := h.taxonomy("base_category", "item_type", "avatar_skin", map[string]string{"fa": "پوسته"})["id"].(string)
	h.item(map[string]any{"key": "mask_item", "typeId": skinType, "price": map[string]any{"currencyKey": "coin", "amount": 30}})
	hat := h.avCategory("hat", nil)
	h.avItem(hat, "crown", map[string]any{"access": "product", "productId": product})
	h.avItem(hat, "mask", map[string]any{"access": "item", "itemKey": "mask_item"})
	u := h.user()
	items := catalogItems(h.mustRPC(u, "economy_avatar_catalog", map[string]any{"lang": "fa"}))
	if items["hat/crown"]["product"].(map[string]any)["key"] != "crown_pack" || items["hat/crown"]["owned"] != false {
		t.Fatalf("product item = %v", items["hat/crown"])
	}
	if c := h.rpcCode(u, "economy_avatar_buy", map[string]any{"category": "hat", "key": "crown"}); c != "AVATAR_PRODUCT_REQUIRED" {
		t.Fatalf("product buy = %s", c)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, status, once_key) VALUES (gen_random_uuid(), $1, $2, 'zibal', 'success', $3)`, u, product, u+":crown"); err != nil {
		t.Fatal(err)
	}
	h.mustRPC(u, "economy_avatar_equip", map[string]any{"category": "hat", "key": "crown"})
	h.give(u, 50)
	buy := h.mustRPC(u, "economy_avatar_buy", map[string]any{"category": "hat", "key": "mask"})
	if buy["item"].(map[string]any)["owned"] != true || buy["wallet"].(map[string]any)["coin"].(float64) != 20 {
		t.Fatalf("shop item buy = %v", buy)
	}
	out := h.mustRPC(u, "economy_avatar_equip", map[string]any{"category": "hat", "key": "mask"})
	if got := lookOf(out["avatar"])["hat"].([]any); got[0] != "mask" {
		t.Fatalf("look = %v", got)
	}
}

func TestAvatarLegacyMigrationAndRPCs(t *testing.T) {
	h := newAvatarHarness(t)
	skinType := h.taxonomy("base_category", "item_type", "avatar_hat", map[string]string{"fa": "کلاه"})["id"].(string)
	h.item(map[string]any{"key": "crown_item", "typeId": skinType, "price": map[string]any{"currencyKey": "coin", "amount": 5}})
	u1, u2, u3 := h.user(), h.user(), h.user()
	for _, q := range []string{
		`INSERT INTO base_avatar_settings (id, settings) VALUES (1, '{"enabled":true}')`,
		`INSERT INTO base_avatar_part (id, key, name, layer, required, sort) VALUES ('11111111-1111-1111-1111-111111111111','skin','{"fa":"پوست"}',0,true,0), ('22222222-2222-2222-2222-222222222222','hat','{"fa":"کلاه"}',5,false,1)`,
		`INSERT INTO base_avatar_option (id, part_id, key, name, value, access, item_key, is_default, sort) VALUES
			('aaaaaaaa-0000-0000-0000-000000000001','11111111-1111-1111-1111-111111111111','light','{"fa":"روشن"}','#F1C27D','free','',true,0),
			('aaaaaaaa-0000-0000-0000-000000000002','11111111-1111-1111-1111-111111111111','dark','{"fa":"تیره"}','#8D5524','free','',false,1)`,
		`INSERT INTO base_avatar_option (id, part_id, key, name, image_url, access, item_key, is_default, sort) VALUES
			('aaaaaaaa-0000-0000-0000-000000000003','22222222-2222-2222-2222-222222222222','cap','{"fa":"کپ"}','https://cdn.example.com/cap.png','free','',true,0),
			('aaaaaaaa-0000-0000-0000-000000000004','22222222-2222-2222-2222-222222222222','crown','{"fa":"تاج"}','https://cdn.example.com/crown.png','item','crown_item',false,1)`,
		`INSERT INTO base_avatar_look (user_id, look) VALUES ($1, '{"skin":"dark","hat":"crown"}'), ($2, '{"skin":"light","hat":""}'), ($3, '{"skin":"dark"}')`,
		`INSERT INTO economy_user_item (user_id, item_id, source) SELECT $1, id, 'buy' FROM economy_item WHERE key='crown_item'`,
	} {
		args := []any{}
		if strings.Contains(q, "$3") {
			args = []any{u1, u2, u3}
		} else if strings.Contains(q, "$1") {
			args = []any{u1}
		}
		if _, err := h.s.db.ExecContext(h.ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile("../../migrate/sql/20261228000000-economy-avatar.sql")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	data := src[strings.Index(src, "INSERT INTO economy_avatar_settings"):strings.Index(src, "-- +migrate Down")]
	for i := 0; i < 2; i++ {
		for _, stmt := range strings.Split(data, ";\n") {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, err := h.s.db.ExecContext(h.ctx, stmt); err != nil {
				t.Fatalf("migration: %v\n%s", err, stmt)
			}
		}
	}
	if n := h.count(`SELECT count(*) FROM economy_avatar`); n != 3 {
		t.Fatalf("avatars = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM economy_avatar_item`); n != 4 {
		t.Fatalf("items = %d", n)
	}
	l1 := h.mustRPC(u1, "base_avatar", map[string]any{"lang": "fa"})
	look := l1["look"].(map[string]any)["parts"].(map[string]any)
	if look["skin"] != "dark" || look["hat"] != "crown" || len(l1["parts"].([]any)) != 2 {
		t.Fatalf("legacy u1 = %v", l1)
	}
	looks := h.mustRPC(u1, "base_avatar_get", map[string]any{"userIds": []string{u1, u2, u3}})["looks"].(map[string]any)
	p2 := looks[u2].(map[string]any)["parts"].(map[string]any)
	p3 := looks[u3].(map[string]any)["parts"].(map[string]any)
	if _, has := p2["hat"]; has || p2["skin"] != "light" {
		t.Fatalf("u2 cleared hat = %v", p2)
	}
	if p3["hat"] != "cap" || p3["skin"] != "dark" {
		t.Fatalf("u3 default hat = %v", p3)
	}
	list := h.mustRPC(u1, "economy_avatar_list", nil)
	if got := lookOf(activeAvatar(list)); got["hat"].([]any)[0] != "crown" {
		t.Fatalf("new api sees the migrated look: %v", got)
	}
	_, err = h.rpc(u2, "base_avatar_save", map[string]any{"look": map[string]string{"hat": "crown"}})
	var e struct {
		Code string         `json:"code"`
		Meta map[string]any `json:"meta"`
	}
	if err == nil || json.Unmarshal([]byte(err.Error()), &e) != nil || e.Code != "AVATAR_OPTION_LOCKED" || e.Meta["itemKey"] != "crown_item" {
		t.Fatalf("legacy locked = %v", err)
	}
	saved := h.mustRPC(u2, "base_avatar_save", map[string]any{"look": map[string]string{"skin": "dark", "hat": "cap"}})
	if p := saved["look"].(map[string]any)["parts"].(map[string]any); p["skin"] != "dark" || p["hat"] != "cap" {
		t.Fatalf("legacy save = %v", saved)
	}
}
