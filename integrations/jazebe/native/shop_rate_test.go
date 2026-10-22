package main

import (
	"context"
	"testing"
)

func TestShopRateConfig(t *testing.T) {
	t.Setenv("RATE_LIMIT_PURCHASE", "")
	t.Setenv("RATE_LIMIT_WINDOW_SECONDS", "")
	limit, window, err := shopRateConfig()
	if err != nil || limit != 20 || window != 60 {
		t.Fatalf("defaults %d %d %v", limit, window, err)
	}
	t.Setenv("RATE_LIMIT_PURCHASE", "7")
	t.Setenv("RATE_LIMIT_WINDOW_SECONDS", "30")
	limit, window, err = shopRateConfig()
	if err != nil || limit != 7 || window != 30 {
		t.Fatalf("override %d %d %v", limit, window, err)
	}
	t.Setenv("RATE_LIMIT_PURCHASE", "0")
	if _, _, err = shopRateConfig(); err == nil {
		t.Fatal("accepted zero limit")
	}
}

func TestShopRateContext(t *testing.T) {
	ctx := context.Background()
	if shopRateAlreadyChecked(ctx) {
		t.Fatal("unchecked context")
	}
	if !shopRateAlreadyChecked(context.WithValue(ctx, shopRateContextKey{}, true)) {
		t.Fatal("checked context")
	}
}
