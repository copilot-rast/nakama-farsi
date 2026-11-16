package main

import (
	"testing"

	"github.com/gofrs/uuid/v5"
)

func TestTimedTagsStickerPriceAndExpiry(t *testing.T) {
	h := newHarness(t)
	resetStickers(h)
	if _, err := h.s.db.Exec(`DELETE FROM economy_tag_discount_settings`); err != nil {
		t.Fatal(err)
	}
	itemID := h.stickerItem("e2e_vip_pack")
	if _, err := h.s.db.Exec(`UPDATE economy_item SET price='{"currencyKey":"coin","amount":200}' WHERE id=$1`, itemID); err != nil {
		t.Fatal(err)
	}
	prizeID, productID := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.Exec(`INSERT INTO economy_prize (id, key, name, items) VALUES ($1,'e2e_pack_prize','{"fa":"بسته"}','[{"type":"item","key":"e2e_vip_pack","action":"give","value":1}]')`, prizeID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO economy_product (id, key, name, prize_id, payment_methods, price_toman, is_active) VALUES ($1,'e2e_pack_product','{"fa":"بسته"}',$2,'{zibal}',30000,true)`, productID, prizeID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		h.s.db.Exec(`DELETE FROM economy_tag_discount WHERE key='e2e_vip5'`)
		h.s.db.Exec(`DELETE FROM economy_product WHERE id=$1`, productID)
		h.s.db.Exec(`DELETE FROM economy_prize WHERE id=$1`, prizeID)
	})
	h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "vip", "name": map[string]string{"fa": "ویژه"}, "access": "item", "itemKey": "e2e_vip_pack",
		"stickers": []map[string]any{{"key": "crown", "imageUrl": "/u/crown.png"}}})
	u := h.user("fa")
	vip := h.userTag("e2e_vip")
	if _, err := h.s.db.Exec(`UPDATE base_tag SET is_public=true WHERE id=$1`, vip); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO economy_tag_discount (id, key, name, tag_ids, percent, item_keys) VALUES (gen_random_uuid(),'e2e_vip5','{"fa":"ویژه"}',ARRAY[$1],5,'{e2e_vip_pack}')`, vip); err != nil {
		t.Fatal(err)
	}
	item := func() map[string]any {
		t.Helper()
		packs := h.mustRPC(u, "base_stickers", map[string]any{})["packs"].([]any)
		return packs[0].(map[string]any)["item"].(map[string]any)
	}
	it := item()
	if it["discountPercent"].(float64) != 0 || it["finalPrice"].(map[string]any)["amount"].(float64) != 200 || len(it["productIds"].([]any)) != 1 || it["productIds"].([]any)[0] != productID {
		t.Fatalf("pack item without tag: %v", it)
	}
	if _, err := h.s.db.Exec(`INSERT INTO base_user_tag (user_id, tag_id, expires_at) VALUES ($1,$2,now()+interval '2 days')`, u, vip); err != nil {
		t.Fatal(err)
	}
	if it = item(); it["discountPercent"].(float64) != 5 || it["finalPrice"].(map[string]any)["amount"].(float64) != 190 {
		t.Fatalf("pack item with vip: %v", it)
	}
	mine := h.mustRPC(u, "base_my_tags", map[string]any{})["items"].([]any)
	if len(mine) != 1 || mine[0].(map[string]any)["expiresAt"] == nil {
		t.Fatalf("my tags: %v", mine)
	}
	if n, err := h.s.expireTags(h.ctx); err != nil || n != 0 {
		t.Fatalf("live tag swept: %d %v", n, err)
	}
	if _, err := h.s.db.Exec(`UPDATE base_user_tag SET expires_at=now()-interval '1 second' WHERE user_id=$1`, u); err != nil {
		t.Fatal(err)
	}
	if mine = h.mustRPC(u, "base_my_tags", map[string]any{})["items"].([]any); len(mine) != 0 {
		t.Fatalf("expired tag listed: %v", mine)
	}
	if it = item(); it["discountPercent"].(float64) != 0 {
		t.Fatalf("expired tag still discounts: %v", it)
	}
	if n, err := h.s.expireTags(h.ctx); err != nil || n != 1 {
		t.Fatalf("sweep = %d %v", n, err)
	}
	var left int
	h.s.db.QueryRow(`SELECT count(*) FROM base_user_tag WHERE user_id=$1`, u).Scan(&left)
	if left != 0 {
		t.Fatalf("rows left after sweep: %d", left)
	}
}
