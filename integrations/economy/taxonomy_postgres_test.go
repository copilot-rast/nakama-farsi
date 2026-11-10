package main

import (
	"net/http"
	"testing"
)

func TestGiftAndCurrencyTaxonomy(t *testing.T) {
	h := newHarness(t)
	curCat := h.taxonomy("base_category", "currency", "premium", map[string]string{"fa": "ویژه", "en": "Premium"})
	giftCat := h.taxonomy("base_category", "gift", "campaign", map[string]string{"fa": "کمپین"})
	prizeCat := h.taxonomy("base_category", "prize", "rewards", map[string]string{"fa": "جایزه"})
	curTag := h.taxonomy("base_tag", "currency", "shiny", map[string]string{"fa": "براق"})
	giftTag := h.taxonomy("base_tag", "gift", "spring", map[string]string{"fa": "بهار"})
	general := h.taxonomy("base_tag", "general", "hot", map[string]string{"fa": "داغ"})
	productTag := h.taxonomy("base_tag", "product", "pack", map[string]string{"fa": "بسته"})

	base := map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس", "en": "Gem"}, "type": "hard"}
	with := func(m map[string]any, extra map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range m {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	if code, out := h.api(http.MethodPost, "/currencies", with(base, map[string]any{"categoryId": prizeCat["id"]})); code != 400 || out["error"] != "CATEGORY_KIND_MISMATCH" {
		t.Fatalf("currency with a prize category = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/currencies", with(base, map[string]any{"tagIds": []any{productTag["id"]}})); code != 400 || out["error"] != "TAG_KIND_MISMATCH" {
		t.Fatalf("currency with a product tag = %d %v", code, out)
	}
	gem := h.mustAPI(201, http.MethodPost, "/currencies", with(base, map[string]any{"categoryId": curCat["id"], "tagIds": []any{curTag["id"], general["id"], curTag["id"]}}))
	if gem["categoryId"] != curCat["id"] || len(gem["tagIds"].([]any)) != 2 {
		t.Fatalf("currency = %v", gem)
	}
	if list := h.mustAPI(200, http.MethodGet, "/currencies?categoryId="+curCat["id"].(string), nil); len(list["items"].([]any)) != 1 {
		t.Fatalf("currency category filter = %v", list)
	}
	prize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "gems", "name": map[string]string{"fa": "الماس"}, "items": []map[string]any{{"type": "currency", "key": "gem", "action": "add", "value": 5}}})

	gift := map[string]any{"name": map[string]string{"fa": "هدیه‌ی بهار", "en": "Spring gift"}, "type": "reward", "code": "SPRING-1", "prizeId": prize["id"], "perUserLimit": 1}
	if code, out := h.api(http.MethodPost, "/gifts", with(gift, map[string]any{"categoryId": curCat["id"]})); code != 400 || out["error"] != "CATEGORY_KIND_MISMATCH" {
		t.Fatalf("gift with a currency category = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/gifts", with(gift, map[string]any{"tagIds": []any{curTag["id"]}})); code != 400 || out["error"] != "TAG_KIND_MISMATCH" {
		t.Fatalf("gift with a currency tag = %d %v", code, out)
	}
	g := h.mustAPI(201, http.MethodPost, "/gifts", with(gift, map[string]any{"categoryId": giftCat["id"], "tagIds": []any{giftTag["id"], general["id"]}}))
	if g["categoryId"] != giftCat["id"] || len(g["tagIds"].([]any)) != 2 {
		t.Fatalf("gift = %v", g)
	}

	user := h.user()
	catalog := h.mustRPC(user, "economy_catalog", map[string]any{"lang": "en"})
	var found map[string]any
	for _, c := range catalog["currencies"].([]any) {
		if c.(map[string]any)["key"] == "gem" {
			found = c.(map[string]any)
		}
	}
	if found == nil || found["categoryKey"] != "premium" || len(found["tags"].([]any)) != 2 || found["tags"].([]any)[0] != "hot" || found["tags"].([]any)[1] != "shiny" {
		t.Fatalf("catalog currency = %v", found)
	}
	red := h.mustRPC(user, "economy_gift_redeem", map[string]any{"code": "spring-1"})
	info, _ := red["gift"].(map[string]any)
	if info == nil || info["categoryKey"] != "campaign" || info["name"] != "Spring gift" || len(info["tags"].([]any)) != 2 {
		t.Fatalf("redeem = %v", red)
	}

	h.mustAPI(200, http.MethodPut, "/gifts/"+g["id"].(string), with(gift, map[string]any{"code": "SPRING-1"}))
	again := h.mustAPI(200, http.MethodGet, "/gifts/"+g["id"].(string), nil)
	if again["categoryId"] != nil || len(again["tagIds"].([]any)) != 0 {
		t.Fatalf("clearing category and tags = %v", again)
	}
}

func TestProductAudienceByPlayerTag(t *testing.T) {
	h := newHarness(t)
	vip := h.taxonomy("base_tag", "user", "vip", map[string]string{"fa": "ویژه"})
	general := h.taxonomy("base_tag", "general", "hot", map[string]string{"fa": "داغ"})
	h.mustAPI(200, http.MethodPut, "/payment-methods/in_app", map[string]any{"isActive": true})
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "audcoin", "name": map[string]string{"fa": "سکه"}, "type": "soft"})
	prize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "aud-prize", "isRepeatable": true, "name": map[string]string{"fa": "جایزه"}, "items": []map[string]any{{"type": "currency", "key": "audcoin", "action": "add", "value": 1}}})
	body := map[string]any{"key": "vip-pack", "isActive": true, "name": map[string]string{"fa": "ویژه"}, "prizeId": prize["id"], "paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "audcoin", "amount": 1}}
	body["audienceTagIds"] = []any{general["id"]}
	if code, out := h.api(http.MethodPost, "/products", body); code != 400 || out["error"] != "TAG_KIND_MISMATCH" || out["field"] != "audienceTagIds" {
		t.Fatalf("audience with a general tag = %d %v", code, out)
	}
	body["audienceTagIds"] = []any{vip["id"], vip["id"]}
	product := h.mustAPI(201, http.MethodPost, "/products", body)
	if got := product["audienceTagIds"].([]any); len(got) != 1 || got[0] != vip["id"] {
		t.Fatalf("audience = %v", got)
	}
	tagged, plain := h.user(), h.user()
	if _, err := h.s.db.Exec(`INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1, $2)`, tagged, vip["id"]); err != nil {
		t.Fatal(err)
	}
	find := func(user string) map[string]any {
		for _, p := range h.mustRPC(user, "economy_catalog", map[string]any{})["products"].([]any) {
			if p.(map[string]any)["key"] == "vip-pack" {
				return p.(map[string]any)
			}
		}
		return nil
	}
	if p := find(tagged); p == nil || p["exclusive"] != true {
		t.Fatalf("tagged player sees the exclusive product: %v", p)
	}
	if p := find(plain); p != nil {
		t.Fatalf("other players do not: %v", p)
	}
	if c := h.rpcCode(plain, "economy_purchase_start", map[string]any{"productId": product["id"], "method": "in_app"}); c != "PRODUCT_NOT_FOUND" {
		t.Fatalf("purchase by an untagged player = %s", c)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_product SET audience_tag_ids = array_remove(audience_tag_ids, $1::text) WHERE id=$2`, vip["id"], product["id"]); err != nil {
		t.Fatal(err)
	}
	if p := find(plain); p == nil || p["exclusive"] != false {
		t.Fatalf("an empty audience is everyone: %v", p)
	}
}
