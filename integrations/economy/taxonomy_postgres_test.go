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
