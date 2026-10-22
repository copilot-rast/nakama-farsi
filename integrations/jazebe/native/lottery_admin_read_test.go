package main

import "testing"

func TestLotteryAdminPageBounds(t *testing.T) {
	page, limit, err := lotteryAdminPage(0, 0)
	if err != nil || page != 1 || limit != 20 {
		t.Fatalf("defaults %d %d %v", page, limit, err)
	}
	for _, tc := range [][2]int{{-1, 20}, {1, -1}, {1, 101}} {
		if _, _, err := lotteryAdminPage(tc[0], tc[1]); err == nil {
			t.Fatalf("accepted page bounds %+v", tc)
		}
	}
}
