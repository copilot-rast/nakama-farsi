package migrate

import (
	"regexp"
	"strings"
	"testing"

	sqlmigrate "github.com/heroiclabs/sql-migrate"
)

func TestJazebeMigrationIsEmbedded(t *testing.T) {
	source := &sqlmigrate.EmbedFileSystemMigrationSource{FileSystem: sqlMigrateFS, Root: "sql"}
	migrations, err := source.FindMigrations()
	if err != nil {
		t.Fatal(err)
	}
	found := make(map[string]bool)
	for _, migration := range migrations {
		if strings.Contains(migration.Id, "jazebe-") {
			if len(migration.Up) == 0 || len(migration.Down) == 0 {
				t.Fatalf("Jazebe migration %s has no up or down section", migration.Id)
			}
			found[migration.Id] = true
		}
	}
	for _, name := range []string{"20261022120000-jazebe-brand-planet.sql", "20261022130000-jazebe-membership.sql", "20261022140000-jazebe-planet-lifecycle.sql", "20261022150000-jazebe-planet-identity.sql", "20261022160000-jazebe-mission.sql", "20261022170000-jazebe-audit.sql", "20261022180000-jazebe-brand-team.sql", "20261022190000-jazebe-mission-rewards.sql", "20261022200000-jazebe-mission-codes.sql", "20261022210000-jazebe-media-upload.sql", "20261022220000-jazebe-groups.sql", "20261022230000-jazebe-group-invites.sql", "20261023000000-jazebe-group-leadership.sql", "20261023010000-jazebe-group-disband.sql", "20261023020000-jazebe-group-chat.sql", "20261023030000-jazebe-lottery.sql", "20261023040000-jazebe-group-mission.sql", "20261023050000-jazebe-shop-orders.sql", "20261023060000-jazebe-lottery-cancel.sql", "20261023070000-jazebe-lottery-draw.sql", "20261023080000-jazebe-coupon-inventory.sql", "20261023090000-jazebe-shop-rate.sql", "20261023100000-jazebe-progression-events.sql", "20261023110000-jazebe-medals.sql", "20261023120000-jazebe-platform-policy.sql", "20261023130000-jazebe-medal-backfill.sql", "20261023140000-jazebe-mission-medal-events.sql", "20261023150000-jazebe-games.sql"} {
		if !found[name] {
			t.Fatalf("Jazebe schema migration %s is missing from the embedded Nakama migrations", name)
		}
	}
	if !found["20261023160000-jazebe-referrals.sql"] {
		t.Fatal("Jazebe referral migration is missing")
	}
	if !found["20261023170000-jazebe-account-profile.sql"] {
		t.Fatal("Jazebe account/profile migration is missing")
	}
	if !found["20261023180000-jazebe-report-scope.sql"] {
		t.Fatal("Jazebe planet-scoped report migration is missing")
	}
	if !found["20261023190000-jazebe-account-restriction.sql"] {
		t.Fatal("Jazebe account restriction migration is missing")
	}
	if !found["20261024040000-jazebe-active-brand-admin.sql"] {
		t.Fatal("Jazebe active brand administrator view migration is missing")
	}
	if !found["20261024030000-jazebe-brand-commercial.sql"] {
		t.Fatal("Jazebe brand commercial record migration is missing")
	}
	if !found["20261024020000-jazebe-account-purge.sql"] {
		t.Fatal("Jazebe account purge migration is missing")
	}
	if !found["20261024010000-jazebe-notification-centre.sql"] {
		t.Fatal("Jazebe notification centre migration is missing")
	}
	if !found["20261024000000-jazebe-catalog-seed.sql"] {
		t.Fatal("Jazebe planet category and medal library seed migration is missing")
	}
}

// Jazebe migrations must survive replay over an existing schema, and their
// Down sections must refuse instead of silently forgetting a migration whose
// tables remain (see docs/ANDRONIX-POSTGRESQL.md, "Rolling back Jazebe
// migrations"). The PostgreSQL replay test proves this against real data.
var jazebeReversibleDown = map[string]bool{
	// Drops only a rate-limit counter table that holds no user data.
	"20261023090000-jazebe-shop-rate.sql": true,
	// Drops only a view.
	"20261024040000-jazebe-active-brand-admin.sql": true,
}

var nonRerunnable = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bCREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?`),
	regexp.MustCompile(`(?i)\bCREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?`),
	regexp.MustCompile(`(?i)\bADD\s+COLUMN\s+(?:IF\s+NOT\s+EXISTS\s+)?`),
	regexp.MustCompile(`(?i)\bDROP\s+CONSTRAINT\s+(?:IF\s+EXISTS\s+)?`),
}

func TestJazebeMigrationsAreRerunnableAndRefuseDown(t *testing.T) {
	source := &sqlmigrate.EmbedFileSystemMigrationSource{FileSystem: sqlMigrateFS, Root: "sql"}
	migrations, err := source.FindMigrations()
	if err != nil {
		t.Fatal(err)
	}
	addConstraint := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(\w+)\s+((?:DROP\s+CONSTRAINT\s+IF\s+EXISTS\s+(\w+)\s*,\s*)?)ADD\s+CONSTRAINT\s+(\w+)`)
	checked := 0
	for _, m := range migrations {
		if !strings.Contains(m.Id, "jazebe-") {
			continue
		}
		checked++
		for _, stmt := range m.Up {
			for _, re := range nonRerunnable {
				for _, match := range re.FindAllString(stmt, -1) {
					if !strings.Contains(strings.ToUpper(match), "IF") {
						t.Errorf("%s: %q is not re-runnable", m.Id, strings.TrimSpace(match))
					}
				}
			}
			for _, match := range addConstraint.FindAllStringSubmatch(stmt, -1) {
				if match[3] == "" || match[3] != match[4] {
					t.Errorf("%s: ADD CONSTRAINT %s must be preceded by DROP CONSTRAINT IF EXISTS of the same name", m.Id, match[4])
				}
			}
		}
		down := strings.Join(m.Down, "\n")
		if !jazebeReversibleDown[m.Id] && !strings.Contains(down, "RAISE EXCEPTION") {
			t.Errorf("%s: Down must refuse (RAISE EXCEPTION) or be listed as reversible", m.Id)
		}
	}
	if checked < 33 {
		t.Fatalf("checked %d Jazebe migrations, want at least 33", checked)
	}
}
