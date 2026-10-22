package main

import "testing"

func TestShopAdminReason(t *testing.T) {
	if _, err := shopAdminReason("  Too short?  ", true); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"", "ab"} {
		if _, err := shopAdminReason(raw, true); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	if got, err := shopAdminReason("  optional  ", false); err != nil || got != "optional" {
		t.Fatalf("optional %q %v", got, err)
	}
}
