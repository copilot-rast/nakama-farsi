package migrate

import (
	"regexp"
	"strings"
	"testing"

	sqlmigrate "github.com/heroiclabs/sql-migrate"
)

var moduleMigration = regexp.MustCompile(`^2026\d{10}-(economy|publication|base)-`)

var nonRerunnable = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bCREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?`),
	regexp.MustCompile(`(?i)\bCREATE\s+SEQUENCE\s+(?:IF\s+NOT\s+EXISTS\s+)?`),
	regexp.MustCompile(`(?i)\bCREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?`),
	regexp.MustCompile(`(?i)\bADD\s+COLUMN\s+(?:IF\s+NOT\s+EXISTS\s+)?`),
	regexp.MustCompile(`(?i)\bDROP\s+CONSTRAINT\s+(?:IF\s+EXISTS\s+)?`),
}

func TestModuleMigrationsAreEmbeddedRerunnableAndRefuseDown(t *testing.T) {
	source := &sqlmigrate.EmbedFileSystemMigrationSource{FileSystem: sqlMigrateFS, Root: "sql"}
	migrations, err := source.FindMigrations()
	if err != nil {
		t.Fatal(err)
	}
	dropConstraint := regexp.MustCompile(`(?i)DROP\s+CONSTRAINT\s+IF\s+EXISTS\s+(\w+)`)
	addConstraint := regexp.MustCompile(`(?i)ADD\s+CONSTRAINT\s+(\w+)`)
	found := map[string]bool{}
	for _, m := range migrations {
		if !moduleMigration.MatchString(m.Id) {
			continue
		}
		found[m.Id] = true
		if len(m.Up) == 0 || len(m.Down) == 0 {
			t.Fatalf("%s has no up or down section", m.Id)
		}
		if !strings.Contains(strings.Join(m.Down, "\n"), "RAISE EXCEPTION") {
			t.Errorf("%s: Down must refuse with RAISE EXCEPTION", m.Id)
		}
		dropped := map[string]bool{}
		for _, stmt := range m.Up {
			for _, re := range nonRerunnable {
				for _, match := range re.FindAllString(stmt, -1) {
					if !strings.Contains(strings.ToUpper(match), "IF") {
						t.Errorf("%s: %q is not re-runnable", m.Id, strings.TrimSpace(match))
					}
				}
			}
			drops := dropConstraint.FindAllStringSubmatchIndex(stmt, -1)
			for _, at := range addConstraint.FindAllStringSubmatchIndex(stmt, -1) {
				name := stmt[at[2]:at[3]]
				ok := dropped[name]
				for _, d := range drops {
					if d[0] < at[0] && stmt[d[2]:d[3]] == name {
						ok = true
					}
				}
				if !ok {
					t.Errorf("%s: ADD CONSTRAINT %s must be preceded by DROP CONSTRAINT IF EXISTS of the same name", m.Id, name)
				}
			}
			for _, d := range drops {
				dropped[stmt[d[2]:d[3]]] = true
			}
		}
	}
	for _, name := range []string{
		"20261030000000-economy-schema.sql",
		"20261031000000-publication-category-kinds.sql",
		"20261031000001-publication-schema.sql",
		"20261101000000-base-schema.sql",
		"20261101000001-base-move-categories.sql",
		"20261101000002-economy-gift-name-localized.sql",
		"20261101000003-publication-recipient-lang.sql",
	} {
		if !found[name] {
			t.Errorf("module migration %s is missing from the embedded migrations", name)
		}
	}
}
