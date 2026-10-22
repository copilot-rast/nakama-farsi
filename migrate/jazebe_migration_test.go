package migrate

import (
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
}
