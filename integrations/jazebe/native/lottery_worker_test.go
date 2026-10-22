package main

import (
	"context"
	"testing"
)

func TestLotteryDrawWorkerOptIn(t *testing.T) {
	t.Setenv("JAZEBE_LOTTERY_DRAW_WORKER_ENABLED", "")
	if err := startLotteryDrawWorker(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_LOTTERY_DRAW_WORKER_ENABLED", "false")
	if err := startLotteryDrawWorker(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_LOTTERY_DRAW_WORKER_ENABLED", "yes")
	if err := startLotteryDrawWorker(context.Background(), nil, nil); err == nil {
		t.Fatal("accepted ambiguous worker setting")
	}
}
