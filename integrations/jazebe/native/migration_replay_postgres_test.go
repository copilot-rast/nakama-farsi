package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/jazebe/schema"
	sqlmigrate "github.com/heroiclabs/sql-migrate"
	"github.com/jackc/pgx/v5"
)

// TestJazebeMigrationReplayPostgres re-executes every Jazebe migration's Up
// section, one transaction per migration as sql-migrate does, against a
// migrated database that already holds Jazebe data, including rows only
// later migrations allow (a 'game' progression event). The replay must
// succeed, keep the schema identical to the committed snapshot and leave
// the data untouched.
func TestJazebeMigrationReplayPostgres(t *testing.T) {
	f := newJazebeFixture(t, "member")
	group := f.ok(f.users["member"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "گروه بازپخش", "mode": "open", "cap": 3})["id"].(string)
	f.exec(`INSERT INTO jazebe_progression_event(id,planet_id,user_id,kind,source_id) VALUES($1,$2,$3,'game',$4)`,
		uuid.Must(uuid.NewV4()).String(), f.planetID, f.users["member"], uuid.Must(uuid.NewV4()).String())
	counts := func() map[string]int {
		out := map[string]int{}
		for _, table := range []string{"jazebe_brand", "jazebe_planet", "jazebe_planet_membership", "jazebe_group", "jazebe_group_membership",
			"jazebe_audit_log", "jazebe_progression_event", "jazebe_planet_category", "jazebe_medal_library", "jazebe_game_template", "jazebe_platform_policy"} {
			out[table] = f.count(`SELECT count(*) FROM ` + table)
		}
		out["fans"] = f.count(`SELECT fans_count FROM jazebe_planet WHERE id=$1`, f.planetID)
		out["members"] = f.count(`SELECT member_count FROM jazebe_group WHERE id=$1`, group)
		return out
	}
	before := counts()

	files, err := filepath.Glob("../../../migrate/sql/*-jazebe-*.sql")
	if err != nil || len(files) < 30 {
		t.Fatalf("Jazebe migrations not found: %d %v", len(files), err)
	}
	sort.Strings(files)
	conn, err := pgx.Connect(f.ctx, os.Getenv(disposableTestDBEnv))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(f.ctx)
	baseline, err := schema.Snapshot(f.ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		r, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		m, err := sqlmigrate.ParseMigration(filepath.Base(file), r)
		r.Close()
		if err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		// Replay the way sql-migrate applies it: notransaction sections run
		// statement by statement, as CockroachDB requires for them.
		if m.DisableTransactionUp {
			for _, stmt := range m.Up {
				if _, err := conn.Exec(f.ctx, stmt); err != nil {
					t.Fatalf("replaying %s failed: %v\n%s", filepath.Base(file), err, strings.TrimSpace(stmt))
				}
			}
			continue
		}
		tx, err := conn.Begin(f.ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, stmt := range m.Up {
			if _, err := tx.Exec(f.ctx, stmt); err != nil {
				tx.Rollback(f.ctx)
				t.Fatalf("replaying %s failed: %v\n%s", filepath.Base(file), err, strings.TrimSpace(stmt))
			}
		}
		if err := tx.Commit(f.ctx); err != nil {
			t.Fatal(err)
		}
	}

	actual, err := schema.Snapshot(f.ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	missing, unexpected := schema.Diff(baseline, actual)
	for _, l := range missing {
		t.Errorf("replay lost: %s", l)
	}
	for _, l := range unexpected {
		t.Errorf("replay added: %s", l)
	}
	after := counts()
	for k, v := range before {
		if after[k] != v {
			t.Errorf("replay changed %s: %d -> %d", k, v, after[k])
		}
	}
}
