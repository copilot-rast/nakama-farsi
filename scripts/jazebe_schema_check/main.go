// Command jazebe_schema_check compares a database's Jazebe schema with the
// committed snapshot produced from a fresh migration. It only reads catalogs.
// Because Jazebe migrations are re-runnable (CREATE ... IF NOT EXISTS), an
// object that already existed with a different definition would be accepted
// silently; this check reports such drift instead.
//
//	go run ./scripts/jazebe_schema_check -db "$DSN"          # verify
//	go run ./scripts/jazebe_schema_check -db "$DSN" -write   # refresh snapshot
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/jazebe/schema"
	"github.com/jackc/pgx/v5"
)

const defaultSnapshot = "migrate/jazebe_schema.snapshot"

func run() error {
	dsn := flag.String("db", os.Getenv("JAZEBE_SCHEMA_CHECK_DB_URL"), "PostgreSQL DSN (read-only use)")
	path := flag.String("snapshot", defaultSnapshot, "schema snapshot file")
	write := flag.Bool("write", false, "rewrite the snapshot from this database (use a freshly migrated one)")
	flag.Parse()
	if *dsn == "" {
		return errors.New("-db or JAZEBE_SCHEMA_CHECK_DB_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := pgx.Connect(ctx, *dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)
	actual, err := schema.Snapshot(ctx, conn)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	major, err := schema.ServerMajor(ctx, conn)
	if err != nil {
		return fmt.Errorf("server version: %w", err)
	}
	if *write {
		return os.WriteFile(*path, []byte(schema.Format(actual, major)), 0o644)
	}
	want, recorded, err := schema.ReadFile(*path)
	if err != nil {
		return err
	}
	if recorded != major {
		fmt.Fprintf(os.Stderr, "warning: snapshot recorded on PostgreSQL %d, server is %d; catalog text may differ\n", recorded, major)
	}
	missing, unexpected := schema.Diff(want, actual)
	for _, l := range missing {
		fmt.Println("MISSING    ", l)
	}
	for _, l := range unexpected {
		fmt.Println("UNEXPECTED ", l)
	}
	if len(missing)+len(unexpected) > 0 {
		return fmt.Errorf("schema drift: %d missing, %d unexpected", len(missing), len(unexpected))
	}
	fmt.Printf("Jazebe schema matches snapshot (%d facts)\n", len(actual))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
