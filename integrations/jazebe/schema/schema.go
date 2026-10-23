// Package schema reads the Jazebe part of a Nakama database schema as
// normalized, sorted facts so that a database can be compared with the
// snapshot of a freshly migrated one. It only reads system catalogs.
package schema

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Each row is one schema fact; table and object names make lines unique.
const schemaQuery = `
SELECT 'column ' || c.table_name || '.' || c.column_name || ' ' || format_type(a.atttypid, a.atttypmod) ||
       CASE WHEN c.is_nullable = 'NO' THEN ' not null' ELSE '' END ||
       COALESCE(' default ' || c.column_default, '')
FROM information_schema.columns c
JOIN pg_attribute a ON a.attrelid = (quote_ident(c.table_schema) || '.' || quote_ident(c.table_name))::regclass
     AND a.attname = c.column_name
WHERE c.table_schema = 'public' AND c.table_name LIKE 'jazebe\_%'
UNION ALL
SELECT 'constraint ' || conrelid::regclass::text || ' ' || conname || ' ' || pg_get_constraintdef(oid)
FROM pg_constraint
WHERE connamespace = 'public'::regnamespace AND conrelid::regclass::text LIKE 'jazebe\_%'
  AND contype <> 'n' -- PostgreSQL 18+ catalogs NOT NULL; column facts already carry it
UNION ALL
SELECT 'index ' || tablename || ' ' || indexname || ' ' || indexdef
FROM pg_indexes
WHERE schemaname = 'public' AND tablename LIKE 'jazebe\_%'`

// Querier is satisfied by *pgx.Conn and pgx.Tx.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Snapshot reads the Jazebe schema as sorted, normalized lines.
func Snapshot(ctx context.Context, conn Querier) ([]string, error) {
	rows, err := conn.Query(ctx, schemaQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, err
		}
		lines = append(lines, strings.Join(strings.Fields(line), " "))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(lines)
	return lines, nil
}

// ServerMajor returns the PostgreSQL major version; catalog text such as
// pg_get_constraintdef output is only compared between equal majors.
func ServerMajor(ctx context.Context, conn Querier) (int, error) {
	rows, err := conn.Query(ctx, `SELECT current_setting('server_version_num')::int / 10000`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var major int
	if rows.Next() {
		err = rows.Scan(&major)
	}
	if err == nil {
		err = rows.Err()
	}
	return major, err
}

// ReadFile parses a snapshot file and the major version it was recorded on.
func ReadFile(path string) (lines []string, major int, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "# postgres-major:"):
			major, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(l, "# postgres-major:")))
		case l == "" || strings.HasPrefix(l, "#"):
		default:
			lines = append(lines, l)
		}
	}
	sort.Strings(lines)
	return lines, major, nil
}

// Format renders a snapshot file body.
func Format(lines []string, major int) string {
	return fmt.Sprintf("# Jazebe schema after all migrations; regenerate with scripts/jazebe_schema_check -write from a fresh database.\n# postgres-major: %d\n", major) +
		strings.Join(lines, "\n") + "\n"
}

// Diff returns facts missing from actual and facts actual has in addition.
func Diff(want, actual []string) (missing, unexpected []string) {
	seen := make(map[string]int, len(actual))
	for _, l := range actual {
		seen[l]++
	}
	for _, l := range want {
		if seen[l] > 0 {
			seen[l]--
			continue
		}
		missing = append(missing, l)
	}
	for _, l := range actual {
		if seen[l] > 0 {
			seen[l]--
			unexpected = append(unexpected, l)
		}
	}
	return missing, unexpected
}
