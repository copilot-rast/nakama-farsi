package main

import (
	"context"
	"net/http"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/store"
)

func init() {
	openers[store.DriverClickHouse] = func(ctx context.Context, cfg storageConfig, secret string) (store.Store, error) {
		c := cfg.ClickHouse
		return store.NewClickHouse(c.URL, c.Database, c.Table, c.User, secret, &http.Client{Timeout: 60 * time.Second})
	}
}
