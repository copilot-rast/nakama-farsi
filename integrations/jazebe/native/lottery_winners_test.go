package main

import "testing"

func TestLotteryDeliveryTransitions(t *testing.T) {
	for _, tc := range []struct {
		from, to string
		want     bool
	}{
		{"pending", "contacted", true},
		{"pending", "delivered", false},
		{"contacted", "delivered", true},
		{"contacted", "rejected", true},
		{"delivered", "contacted", false},
		{"rejected", "delivered", false},
	} {
		if got := lotteryDeliveryTransition(tc.from, tc.to); got != tc.want {
			t.Fatalf("%s -> %s: %v", tc.from, tc.to, got)
		}
	}
}
