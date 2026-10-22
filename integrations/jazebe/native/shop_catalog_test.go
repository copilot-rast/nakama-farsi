package main

import "testing"

func TestValidateShopCatalog(t *testing.T) {
	for _, in := range []shopCatalogInput{{}, {Type: "ticket", Sort: "cheap"}, {Type: "coupon", Sort: "expensive"}, {Type: "digital", Sort: "fresh"}} {
		if err := validateShopCatalog(in); err != nil {
			t.Fatalf("%+v: %v", in, err)
		}
	}
	for _, in := range []shopCatalogInput{{Type: "physical"}, {Sort: "random"}} {
		if err := validateShopCatalog(in); err == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
}
