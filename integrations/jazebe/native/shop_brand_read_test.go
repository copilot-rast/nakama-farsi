package main

import "testing"

func TestValidateShopBrandList(t *testing.T) {
	in := shopBrandInput{}
	if err := validateShopBrandList(&in); err != nil || in.Page != 1 || in.Limit != 20 {
		t.Fatalf("defaults %+v: %v", in, err)
	}
	for _, bad := range []shopBrandInput{{Page: -1}, {Limit: 101}, {Status: "deleted"}, {Type: "service"}} {
		if err := validateShopBrandList(&bad); err == nil {
			t.Fatalf("accepted %+v", bad)
		}
	}
	if optionalText("") != nil || optionalSearch("") != nil || optionalSearch("a") != "%a%" {
		t.Fatal("optional filters")
	}
}
