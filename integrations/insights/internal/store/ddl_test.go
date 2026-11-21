package store

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationHasTheSameSchema(t *testing.T) {
	raw, err := os.ReadFile("../../../../migrate/sql/20270115000000-insights-events.sql")
	if err != nil {
		t.Fatal(err)
	}
	norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	text := norm(string(raw))
	for _, stmt := range PGSchema("") {
		if !strings.Contains(text, norm(stmt)+";") {
			t.Fatalf("migration is missing: %s", stmt)
		}
	}
	for _, stmt := range PGSchema("ext") {
		if !strings.Contains(stmt, `"ext".`) {
			t.Fatalf("schema not applied: %s", stmt)
		}
	}
}
