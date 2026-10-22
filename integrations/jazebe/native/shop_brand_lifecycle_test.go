package main

import "testing"

func TestBrandProductStatusAllowed(t *testing.T) {
	for _, tc := range []struct {
		current, next string
		ok            bool
	}{{"active", "paused", true}, {"paused", "active", true}, {"archived", "archived", true}, {"archived", "active", false}, {"suspended", "paused", false}, {"active", "suspended", false}} {
		if (brandProductStatusAllowed(tc.current, tc.next) == nil) != tc.ok {
			t.Fatalf("%+v", tc)
		}
	}
}
