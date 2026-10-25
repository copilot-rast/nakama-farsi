package schema

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestDiff(t *testing.T) {
	want := []string{"a", "b", "c"}
	missing, unexpected := Diff(want, []string{"a", "c", "d"})
	if !reflect.DeepEqual(missing, []string{"b"}) || !reflect.DeepEqual(unexpected, []string{"d"}) {
		t.Fatalf("missing=%v unexpected=%v", missing, unexpected)
	}
	if m, u := Diff(want, want); len(m)+len(u) != 0 {
		t.Fatalf("identical schemas differ: %v %v", m, u)
	}
}

// TestSnapshotMatchesMigrationsPostgres requires a disposable database with
// all migrations applied and checks it against the committed snapshot.
func TestSnapshotMatchesMigrationsPostgres(t *testing.T) {
	dsn := os.Getenv("JAZEBE_DISPOSABLE_TEST_DB_URL")
	if dsn == "" {
		t.Skip("JAZEBE_DISPOSABLE_TEST_DB_URL not set; PostgreSQL NOT_RUN")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	var name string
	if err := conn.QueryRow(ctx, `SELECT current_database()`).Scan(&name); err != nil || !strings.HasSuffix(name, "_disposable") {
		t.Fatalf("refusing non-disposable database %q (%v)", name, err)
	}
	actual, err := Snapshot(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	want, recorded, err := ReadFile("../../../migrate/jazebe_schema.snapshot")
	if err != nil {
		t.Fatal(err)
	}
	major, err := ServerMajor(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	if major != recorded {
		t.Skipf("snapshot recorded on PostgreSQL %d, server is %d: catalog text is only compared between equal majors; regenerate with -write to check this version", recorded, major)
	}
	missing, unexpected := Diff(want, actual)
	for _, l := range missing {
		t.Errorf("missing: %s", l)
	}
	for _, l := range unexpected {
		t.Errorf("unexpected: %s", l)
	}
}

// TestSnapshotDetectsDriftPostgres alters the schema inside a transaction
// that is always rolled back and checks that the drift is reported.
func TestSnapshotDetectsDriftPostgres(t *testing.T) {
	dsn := os.Getenv("JAZEBE_DISPOSABLE_TEST_DB_URL")
	if dsn == "" {
		t.Skip("JAZEBE_DISPOSABLE_TEST_DB_URL not set; PostgreSQL NOT_RUN")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	var name string
	if err := conn.QueryRow(ctx, `SELECT current_database()`).Scan(&name); err != nil || !strings.HasSuffix(name, "_disposable") {
		t.Fatalf("refusing non-disposable database %q (%v)", name, err)
	}
	var version string
	if err := conn.QueryRow(ctx, `SELECT version()`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(version, "CockroachDB") {
		// The probe relies on PostgreSQL's transactional DDL rollback and catalog text.
		t.Skip("schema drift snapshot is PostgreSQL-only; CockroachDB NOT_RUN")
	}
	before, err := Snapshot(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	for _, stmt := range []string{
		`ALTER TABLE jazebe_planet_category ALTER COLUMN title_fa TYPE varchar(64)`,
		`CREATE INDEX drift_probe_idx ON jazebe_planet_category (sort_order)`,
	} {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}
	after, err := Snapshot(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	missing, unexpected := Diff(before, after)
	joined := strings.Join(append(missing, unexpected...), "\n")
	for _, want := range []string{"jazebe_planet_category.title_fa character varying(128)", "jazebe_planet_category.title_fa character varying(64)", "drift_probe_idx"} {
		if !strings.Contains(joined, want) {
			t.Errorf("drift not reported: %s\n%s", want, joined)
		}
	}
}
