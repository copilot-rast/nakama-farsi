package main

import (
	"testing"
	"time"
)

func TestShopPurchasePersianRecord(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	record := shopTicketRecord("order", "product", "ticket", 4, now, "lottery", "raffle", now)
	if record["when"] != "امروز · ۱۵:۳۰" || record["mo"] != 0 || record["raffle"] != "raffle" {
		t.Fatalf("record %+v", record)
	}
	if got := shopWhenFa(now.Add(-24*time.Hour), now); got != "دیروز · ۱۵:۳۰" {
		t.Fatalf("yesterday: %q", got)
	}
}

func TestShopUntilFa(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		after time.Duration
		want  string
	}{
		{0, "به پایان رسیده"},
		{30 * time.Minute, "۳۰ دقیقه دیگر"},
		{5 * time.Hour, "۵ ساعت دیگر"},
		{51 * time.Hour, "۲ روز و ۳ ساعت دیگر"},
	} {
		if got := shopUntilFa(now.Add(tc.after), now); got != tc.want {
			t.Fatalf("%v: %q", tc.after, got)
		}
	}
}
