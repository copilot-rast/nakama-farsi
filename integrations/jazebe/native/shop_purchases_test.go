package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestShopPurchaseFilters(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		kind, period string
		days         int
	}{
		{"ticket", "month", 30},
		{"coupon", "quarter", 120},
		{"digital", "all", 0},
	} {
		since, err := shopPurchaseFilters(shopPurchasesInput{Type: tc.kind, Range: tc.period}, now)
		if err != nil || since.Valid != (tc.days != 0) || since.Valid && !since.Time.Equal(now.Add(-time.Duration(tc.days)*24*time.Hour)) {
			t.Fatalf("%+v: %+v %v", tc, since, err)
		}
	}
	for _, in := range []shopPurchasesInput{{Type: "physical"}, {Range: "year"}} {
		if _, err := shopPurchaseFilters(in, now); err == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
}

func TestShopPurchaseRecordDelivery(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		kind, payload string
		want          any
	}{
		{"ticket", `{"lotteryId":"lottery"}`, nil},
		{"coupon", `{"couponCode":"SAVE-1","couponExpiresAt":"2026-10-23T12:00:00Z"}`, "SAVE-1"},
		{"digital", `{"digitalPayload":{"value":"https://example.test/item"}}`, "https://example.test/item"},
	} {
		record, err := shopPurchaseRecord("order", "product", "name", tc.kind, 10, now, json.RawMessage(tc.payload), "raffle", now)
		if err != nil || record["code"] != tc.want {
			t.Fatalf("%s: %+v %v", tc.kind, record, err)
		}
		if tc.kind == "coupon" && (record["expired"] != false || record["expiry"] == "") {
			t.Fatalf("coupon expiry: %+v", record)
		}
	}
	if _, err := shopPurchaseRecord("", "", "", "digital", 0, now, []byte(`{"digitalPayload":`), "", now); err == nil {
		t.Fatal("accepted malformed delivery")
	}
}
