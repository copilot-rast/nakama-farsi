package main

import (
	"encoding/json"
	"testing"
)

func TestPrizeItems(t *testing.T) {
	good := []prizeItem{
		{Type: "currency", Key: "coin", Action: "add", Value: json.RawMessage(`100`)},
		{Type: "currency", Key: "coin", Action: "add", Value: json.RawMessage(`-5`)},
		{Type: "meta", Key: "xp", Action: "add", Value: json.RawMessage(`1.25`)},
		{Type: "meta", Key: "level", Action: "set", Value: json.RawMessage(`"3"`)},
		{Type: "storage", Key: "skin.red", Action: "give", Value: json.RawMessage(`{"color":"red"}`)},
	}
	out, err := normalizePrizeItems(good)
	if err != nil {
		t.Fatal(err)
	}
	if string(out[3].Value) != "3" || string(out[2].Value) != "1.25" {
		t.Fatalf("normalized = %s %s", out[3].Value, out[2].Value)
	}
	bad := [][]prizeItem{
		nil,
		{{Type: "currency", Key: "coin", Action: "set", Value: json.RawMessage(`1`)}},
		{{Type: "currency", Key: "coin", Action: "add", Value: json.RawMessage(`0`)}},
		{{Type: "currency", Key: "coin", Action: "add", Value: json.RawMessage(`1.5`)}},
		{{Type: "currency", Key: "Coin!", Action: "add", Value: json.RawMessage(`1`)}},
		{{Type: "meta", Key: "xp", Action: "give", Value: json.RawMessage(`1`)}},
		{{Type: "meta", Key: "xp", Action: "add", Value: json.RawMessage(`"abc"`)}},
		{{Type: "meta", Key: "xp", Action: "add", Value: json.RawMessage(`12345678901234567`)}},
		{{Type: "storage", Key: "a", Action: "give", Value: json.RawMessage(`[1]`)}},
		{{Type: "avatar", Key: "a", Action: "give", Value: json.RawMessage(`{}`)}},
	}
	for i, items := range bad {
		if _, err := normalizePrizeItems(items); err == nil {
			t.Fatalf("case %d must fail", i)
		}
	}
}

func TestProductShape(t *testing.T) {
	price := func(v int64) *int64 { return &v }
	ok := []productInput{
		{PaymentMethods: []string{"zarinpal", "zibal"}, PriceToman: price(100)},
		{PaymentMethods: []string{"in_app"}, CoinPrice: &coinPrice{CurrencyKey: "coin", Amount: 5}},
		{PaymentMethods: []string{"bazaar", "myket"}, SKU: map[string]string{"bazaar": "a.b", "myket": "c"}},
		{PaymentMethods: []string{"zibal"}, PriceToman: price(1000), DiscountPercent: 100},
		{PaymentMethods: []string{"bazaar"}, SKU: map[string]string{"bazaar": "vip"}, IsSubscription: true},
	}
	for i := range ok {
		if err := validateProductShape(&ok[i]); err != nil {
			t.Fatalf("ok case %d: %v", i, err)
		}
	}
	bad := map[string]productInput{
		"PRICE_REQUIRED":             {PaymentMethods: []string{"zarinpal"}, PriceToman: price(99)},
		"COIN_PRICE_REQUIRED":        {PaymentMethods: []string{"in_app"}},
		"SKU_REQUIRED":               {PaymentMethods: []string{"myket"}, SKU: map[string]string{"bazaar": "x"}},
		"INVALID_PAYMENT_METHOD":     {PaymentMethods: []string{"paypal"}},
		"INVALID_DISCOUNT":           {DiscountPercent: 101},
		"SUBSCRIPTION_NOT_SUPPORTED": {PaymentMethods: []string{"myket"}, SKU: map[string]string{"myket": "v"}, IsSubscription: true},
		"PRICE_TOO_LOW":              {PaymentMethods: []string{"zibal"}, PriceToman: price(150), DiscountPercent: 50},
		"INVALID_PRICE":              {CoinPrice: &coinPrice{CurrencyKey: "coin", Amount: 0}},
	}
	for code, in := range bad {
		err := validateProductShape(&in)
		app, isApp := asApp(err)
		if !isApp || app.Code != code {
			t.Fatalf("want %s got %v", code, err)
		}
	}
}

func TestI18nAndMisc(t *testing.T) {
	if _, err := cleanI18n("name", i18n{"en": "x"}, true, 10); err == nil {
		t.Fatal("fa is required")
	}
	if _, err := cleanI18n("name", i18n{"fa": "x", "de": "y"}, true, 10); err == nil {
		t.Fatal("unknown language accepted")
	}
	if out, err := cleanI18n("message", i18n{}, false, 10); err != nil || len(out) != 0 {
		t.Fatal("optional empty must pass")
	}
	if discounted(1000, 25) != 750 || discounted(1000, 100) != 0 || discounted(999, 0) != 999 {
		t.Fatal("discount math")
	}
	if sandboxPayload("sandbox-a-b-c") != "a-b" || sandboxPayload("sandbox-x") != "" {
		t.Fatal("sandbox payload")
	}
	if maskTicketURL("http://h/v2/economy/payments/callback/abc.def") != "http://h/v2/economy/payments/callback/********" {
		t.Fatal("ticket mask")
	}
	merged, err := mergeMethodConfig("zarinpal", methodConfig{"merchantId": "old"}, map[string]any{"merchantId": secretMask, "sandbox": true})
	if err != nil || merged.str("merchantId") != "old" || !merged.flag("sandbox") {
		t.Fatalf("mask keeps value: %v %v", merged, err)
	}
	merged, _ = mergeMethodConfig("zarinpal", methodConfig{"merchantId": "old"}, map[string]any{"merchantId": ""})
	if merged.str("merchantId") != "" || methodConfigured("zarinpal", merged) {
		t.Fatal("empty clears value")
	}
	if !methodConfigured("zarinpal", methodConfig{"sandbox": true}) || !methodConfigured("zibal", methodConfig{"sandbox": true}) {
		t.Fatal("sandbox gateways need no merchant")
	}
	if methodConfigured("bazaar", methodConfig{"sandbox": true}) || !methodConfigured("bazaar", methodConfig{"sandbox": true, "packageName": "p"}) {
		t.Fatal("stores always need a package name")
	}
	view := methodView(paymentMethod{Method: "myket", Config: methodConfig{"accessToken": "tok", "packageName": "p"}})
	if cfg := view["config"].(map[string]any); cfg["accessToken"] != secretMask || cfg["packageName"] != "p" {
		t.Fatalf("view = %v", view)
	}
	if ext, ok := sniffImage([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`)); !ok || ext != "svg" {
		t.Fatal("plain svg refused")
	}
	if _, ok := sniffImage([]byte(`<svg onload="alert(1)"></svg>`)); ok {
		t.Fatal("svg with handler accepted")
	}
	if _, ok := sniffImage([]byte("GIF89a....")); ok {
		t.Fatal("gif accepted")
	}
	if ext, ok := sniffImage(append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 20)...)); !ok || ext != "png" {
		t.Fatal("png refused")
	}
}
