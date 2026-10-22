package main

import (
	"context"
	"testing"
)

func TestProgressionWorkerConfiguration(t *testing.T) {
	t.Setenv("JAZEBE_PROGRESSION_WORKER_ENABLED", "")
	if err := startProgressionWorker(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_PROGRESSION_WORKER_ENABLED", "false")
	if err := startProgressionWorker(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_PROGRESSION_WORKER_ENABLED", "yes")
	if err := startProgressionWorker(context.Background(), nil, nil); err == nil {
		t.Fatal("invalid worker setting accepted")
	}
}
