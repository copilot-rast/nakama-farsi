package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	disposableTestDBEnv    = "PUBLICATION_DISPOSABLE_TEST_DB_URL"
	disposableTestDBSuffix = "_disposable"
)

func validateDisposableDSN(dsn string) error {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("%s is not a valid PostgreSQL DSN", disposableTestDBEnv)
	}
	if !strings.HasSuffix(cfg.Database, disposableTestDBSuffix) {
		return fmt.Errorf("%s database %q must end in %q", disposableTestDBEnv, cfg.Database, disposableTestDBSuffix)
	}
	hosts := []string{cfg.Host}
	for _, fallback := range cfg.Fallbacks {
		hosts = append(hosts, fallback.Host)
	}
	for _, host := range hosts {
		if strings.HasPrefix(host, "/") || host == "localhost" {
			continue
		}
		if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
			return fmt.Errorf("%s host %q must be loopback or a local socket", disposableTestDBEnv, host)
		}
	}
	return nil
}

func openDisposableTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv(disposableTestDBEnv)
	if dsn == "" {
		t.Skip(disposableTestDBEnv + " not set; PostgreSQL NOT_RUN")
	}
	if err := validateDisposableDSN(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var name string
	if err = db.QueryRowContext(ctx, `SELECT current_database()`).Scan(&name); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, disposableTestDBSuffix) {
		db.Close()
		t.Fatalf("connected database %q is not disposable", name)
	}
	return db
}

func TestValidateDisposableDSN(t *testing.T) {
	for _, tc := range []struct {
		dsn string
		ok  bool
	}{
		{"postgres://u:p@127.0.0.1:5477/publication_test_disposable?sslmode=disable", true},
		{"postgres://u:p@localhost/x_disposable", true},
		{"postgres://u:p@127.0.0.1:5477/nakama", false},
		{"postgres://u:p@db.example.com:5432/x_disposable", false},
		{"::not a dsn::", false},
	} {
		if err := validateDisposableDSN(tc.dsn); (err == nil) != tc.ok {
			t.Errorf("validateDisposableDSN(%q) error=%v, want ok=%v", tc.dsn, err, tc.ok)
		}
	}
}
