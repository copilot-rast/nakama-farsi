package main

import (
	"context"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/mongowire"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/store"
)

func init() {
	openers[store.DriverMongo] = func(ctx context.Context, cfg storageConfig, secret string) (store.Store, error) {
		c := cfg.Mongo
		hosts := []string{}
		for _, h := range strings.Split(c.Hosts, ",") {
			if h = strings.TrimSpace(h); h != "" {
				if !strings.Contains(h, ":") {
					h += ":27017"
				}
				hosts = append(hosts, h)
			}
		}
		cl, err := mongowire.New(mongowire.Options{Hosts: hosts, User: c.User, Password: secret, AuthSource: c.AuthSource, TLS: c.TLS, Timeout: 30 * time.Second, PoolSize: 4})
		if err != nil {
			return nil, err
		}
		st, err := store.NewMongo(cl, c.Database, c.Collection)
		if err != nil {
			_ = cl.Close()
			return nil, err
		}
		return st, nil
	}
}
