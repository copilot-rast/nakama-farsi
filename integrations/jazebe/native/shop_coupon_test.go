package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCouponFormatAndRandomCode(t *testing.T) {
	for _, raw := range []string{"abc", "long-" + strings.Repeat("#", 50), "without-marker"} {
		if _, err := couponFormat(raw); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	format, err := couponFormat("")
	if err != nil || format != couponDefaultFormat {
		t.Fatalf("default %q: %v", format, err)
	}
	for i := 0; i < 100; i++ {
		code, err := renderCouponCode("OFF-####-####")
		if err != nil || len(code) != 13 || !strings.HasPrefix(code, "OFF-") {
			t.Fatalf("code %q: %v", code, err)
		}
		for _, r := range strings.ReplaceAll(strings.TrimPrefix(code, "OFF-"), "-", "") {
			if !strings.ContainsRune(couponAlphabet, r) {
				t.Fatalf("invalid character %q", r)
			}
		}
	}
}

func TestCouponProductInputPromotedFields(t *testing.T) {
	var in couponProductInput
	if err := json.Unmarshal([]byte(`{"planetId":"p","name":"Coupon","pricePoints":10,"coupon":{"codeFormat":"OFF-####"}}`), &in); err != nil {
		t.Fatal(err)
	}
	if in.PlanetID != "p" || in.Name != "Coupon" || in.PricePoints != 10 || in.Coupon.CodeFormat != "OFF-####" {
		t.Fatalf("decode %+v", in)
	}
}

func TestCouponGenerateExpirationClear(t *testing.T) {
	var absent, cleared couponGenerateInput
	if err := json.Unmarshal([]byte(`{"count":1}`), &absent); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"count":1,"expiresAt":null}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if absent.ExpiresAt != nil || string(cleared.ExpiresAt) != "null" {
		t.Fatalf("absent=%s cleared=%s", absent.ExpiresAt, cleared.ExpiresAt)
	}
}
