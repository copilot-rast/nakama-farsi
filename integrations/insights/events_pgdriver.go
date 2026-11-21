package main

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func init() {
	openers[store.DriverPostgres] = openExternalPG
}

func pgDSN(c pgConfig, password string) string {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(c.Host, strconv.Itoa(c.Port)), Path: "/" + c.Database}
	if c.User != "" {
		if password != "" {
			u.User = url.UserPassword(c.User, password)
		} else {
			u.User = url.User(c.User)
		}
	}
	q := url.Values{}
	q.Set("sslmode", c.SSLMode)
	q.Set("connect_timeout", "10")
	q.Set("application_name", "nakama-insights")
	u.RawQuery = q.Encode()
	return u.String()
}

func openExternalPG(ctx context.Context, cfg storageConfig, secret string) (store.Store, error) {
	db, err := sql.Open("pgx", pgDSN(cfg.Postgres, secret))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(5 * time.Minute)
	pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	st, err := store.NewExternalPG(db, cfg.Postgres.Schema)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return st, nil
}
