package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

// sweepLogger records module errors so internal failures can be traced to a
// cause without exposing it to the client.
type sweepLogger struct {
	runtime.Logger
	mu     sync.Mutex
	errors []string
}

func (l *sweepLogger) Debug(string, ...interface{}) {}
func (l *sweepLogger) Info(string, ...interface{})  {}
func (l *sweepLogger) Warn(string, ...interface{})  {}
func (l *sweepLogger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	l.errors = append(l.errors, fmt.Sprintf(format, v...))
	l.mu.Unlock()
}
func (l *sweepLogger) WithField(string, interface{}) runtime.Logger     { return l }
func (l *sweepLogger) WithFields(map[string]interface{}) runtime.Logger { return l }
func (l *sweepLogger) Fields() map[string]interface{}                   { return nil }
func (l *sweepLogger) take() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := l.errors
	l.errors = nil
	return out
}

// sweepNakama fails loudly on any NakamaModule call the sweep does not stub,
// which the sweep reports as a dependency instead of crashing.
type sweepNakama struct{ runtime.NakamaModule }

type sweepResult struct {
	RPC     string   `json:"rpc"`
	Case    string   `json:"case"`
	Code    int      `json:"code"`
	Message string   `json:"message,omitempty"`
	Panic   string   `json:"panic,omitempty"`
	Logged  []string `json:"logged,omitempty"`
	Writes  []string `json:"writes,omitempty"`
}

var sqlLeak = regexp.MustCompile(`(?i)SQLSTATE|\bERROR:|syntax error|relation "|column "|violates|pgx|pq:|sql: `)

// TestJazebeRPCSweepPostgres drives every registered RPC through the same
// wrappers Nakama uses, as an anonymous caller and as an ordinary user with
// malformed or well-formed but nonexistent input, against a disposable
// migrated database. It asserts on behavior every RPC must share and writes a
// per-RPC report when JAZEBE_RPC_SWEEP_REPORT names a file.
func TestJazebeRPCSweepPostgres(t *testing.T) {
	// The fixture's verified cleanup also removes rows RPCs create lazily for
	// the caller (account, referral identity).
	f := newJazebeFixture(t)
	db, ctx := f.db, f.ctx
	plain := f.user("plain")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", "")
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", "")
	t.Setenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS", "")

	missing := map[string]any{}
	for _, key := range []string{"id", "planetId", "brandId", "groupId", "missionId", "lotteryId", "productId",
		"userId", "memberId", "reportId", "gameId", "medalId", "completionId", "requestId", "inviteId",
		"offerId", "orderId", "sessionId", "mediaId", "categoryId", "templateId", "winnerId", "shareId",
		"messageId", "targetId", "subjectId"} {
		missing[key] = uuid.Must(uuid.NewV4()).String()
	}
	missingJSON, _ := json.Marshal(missing)
	cases := []struct {
		name, payload string
		anonymous     bool
	}{
		{"anonymous", `{}`, true},
		{"empty", ``, false},
		{"malformed", `{"planetId":`, false},
		{"object", `{}`, false},
		{"missing-ids", string(missingJSON), false},
	}

	tables := jazebeTables(ctx, t, db)
	logger := &sweepLogger{}
	var results []sweepResult
	for _, entry := range jazebeRPCEntries() {
		rpc := nakamaRPCError(guardJazebeRPC(entry.fn))
		for _, tc := range cases {
			callCtx := ctx
			if !tc.anonymous {
				callCtx = context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, plain)
			}
			before := tableCounts(ctx, t, db, tables)
			res := sweepResult{RPC: entry.name, Case: tc.name}
			func() {
				defer func() {
					if r := recover(); r != nil {
						res.Panic = fmt.Sprint(r)
					}
				}()
				_, err := rpc(callCtx, logger, db, &sweepNakama{}, tc.payload)
				var re *runtime.Error
				switch {
				case err == nil:
					res.Code = 0
				case errors.As(err, &re):
					res.Code, res.Message = re.Code, re.Message
				default:
					res.Code, res.Message = -1, err.Error()
				}
			}()
			res.Logged = logger.take()
			after := tableCounts(ctx, t, db, tables)
			for _, name := range tables {
				if before[name] != after[name] {
					res.Writes = append(res.Writes, fmt.Sprintf("%s%+d", name, after[name]-before[name]))
				}
			}
			results = append(results, res)

			if res.Panic != "" && !strings.Contains(res.Panic, "nil pointer") {
				t.Errorf("%s/%s panicked: %s", entry.name, tc.name, res.Panic)
			}
			if sqlLeak.MatchString(res.Message) {
				t.Errorf("%s/%s leaks database detail to client: %q", entry.name, tc.name, res.Message)
			}
			if tc.anonymous && res.Code != 16 {
				t.Errorf("%s anonymous call returned code %d, want 16", entry.name, res.Code)
			}
		}
	}
	if path := os.Getenv("JAZEBE_RPC_SWEEP_REPORT"); path != "" {
		b, _ := json.MarshalIndent(results, "", " ")
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("swept %d RPCs x %d cases", len(jazebeRPCEntries()), len(cases))
}

func jazebeTables(ctx context.Context, t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public'
		AND (tablename LIKE 'jazebe\_%' OR tablename IN ('users','notification','wallet_ledger','storage','groups','group_edge'))`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func tableCounts(ctx context.Context, t *testing.T, db *sql.DB, tables []string) map[string]int64 {
	t.Helper()
	parts := make([]string, len(tables))
	for i, name := range tables {
		parts[i] = fmt.Sprintf("(SELECT count(*) FROM %q)", name)
	}
	row := db.QueryRowContext(ctx, "SELECT "+strings.Join(parts, ","))
	vals := make([]int64, len(tables))
	ptrs := make([]any, len(tables))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := row.Scan(ptrs...); err != nil {
		t.Fatal(err)
	}
	out := make(map[string]int64, len(tables))
	for i, name := range tables {
		out[name] = vals[i]
	}
	return out
}
