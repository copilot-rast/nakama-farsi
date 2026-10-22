package main

import "testing"

func TestLotteryPrizeConversionConfiguration(t *testing.T) {
	t.Setenv("LOTTERY_PRIZE_RIAL_PER_POINT", "")
	if n, err := lotteryPrizeRate(); err != nil || n != 1 {
		t.Fatalf("default rate: %d %v", n, err)
	}
	t.Setenv("LOTTERY_PRIZE_RIAL_PER_POINT", "25")
	if n, err := lotteryPrizeRate(); err != nil || n != 25 {
		t.Fatalf("configured rate: %d %v", n, err)
	}
	for _, bad := range []string{"0", "-2", "1.5", "secret"} {
		t.Setenv("LOTTERY_PRIZE_RIAL_PER_POINT", bad)
		if _, err := lotteryPrizeRate(); err == nil {
			t.Fatalf("accepted invalid rate %q", bad)
		}
	}
}
