package main

import "testing"

func TestShopSalesFilters(t *testing.T) {
	from, to, kind, err := shopSalesFilters(shopSalesInput{From: "2026-10-01", To: "2026-11-01T00:00:00Z", Type: "coupon"})
	if err != nil || !from.Valid || !to.Valid || !kind.Valid || kind.String != "coupon" {
		t.Fatalf("filters %+v %+v %+v %v", from, to, kind, err)
	}
	for _, in := range []shopSalesInput{{From: "yesterday"}, {From: "2026-11-01", To: "2026-10-01"}, {Type: "unknown"}} {
		if _, _, _, err = shopSalesFilters(in); err == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
}
