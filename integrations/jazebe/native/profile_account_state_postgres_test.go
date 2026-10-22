package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Requires a database explicitly designated disposable with all Nakama and
// Jazebe migrations applied. No existing user is modified by this fixture.
func TestAccountTransitionPostgres(t *testing.T) {
	dsn := os.Getenv("JAZEBE_DISPOSABLE_TEST_DB_URL")
	if dsn == "" {
		t.Skip("JAZEBE_DISPOSABLE_TEST_DB_URL not set; PostgreSQL NOT_RUN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	actor, target := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		for _, query := range []string{
			`DELETE FROM notification WHERE user_id=$1 AND content->>'subjectType'='account'`,
			`DELETE FROM jazebe_audit_log WHERE target_type='user' AND target_id=$1`,
			`DELETE FROM jazebe_user_account WHERE user_id=$1`,
			`DELETE FROM users WHERE id=$1`,
		} {
			if _, e := db.ExecContext(cleanup, query, target); e != nil {
				t.Errorf("fixture cleanup: %v", e)
			}
		}
		if _, e := db.ExecContext(cleanup, `DELETE FROM users WHERE id=$1`, actor); e != nil {
			t.Errorf("operator cleanup: %v", e)
		}
	}()
	for _, id := range []string{actor, target} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,username) VALUES($1,$2)`, id, "account-state-"+id); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", actor)
	caller := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, actor)
	request, _ := json.Marshal(operatorAccountAction{UserID: target, Reason: "verified moderation finding"})
	nk := &accountRevocationNakama{}
	if _, err = operatorAccountSuspend(caller, nil, db, nk, string(request)); err != nil || !nk.loggedOut {
		t.Fatalf("suspend/revoke: %v loggedOut=%v", err, nk.loggedOut)
	}
	protected := guardJazebeRPC(func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
		return `{"allowed":true}`, nil
	})
	targetCaller := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, target)
	if _, err = protected(targetCaller, nil, db, nil, "{}"); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("restricted account reached Jazebe RPC: %v", err)
	}
	var state string
	var pending bool
	var disabled time.Time
	if err = db.QueryRowContext(ctx, `SELECT a.state,a.session_revocation_pending,u.disable_time
		FROM jazebe_user_account a JOIN users u ON u.id=a.user_id WHERE a.user_id=$1`, target).Scan(&state, &pending, &disabled); err != nil || state != "suspended" || pending || disabled.Unix() == 0 {
		t.Fatalf("suspension persisted incorrectly: %s %v %v %v", state, pending, disabled, err)
	}
	if _, err = operatorAccountSuspend(caller, nil, db, nk, string(request)); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("duplicate suspend: %v", err)
	}
	outsider := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, target)
	if _, err = operatorAccountUnsuspend(outsider, nil, db, nk, string(request)); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unauthorized unsuspend: %v", err)
	}
	if _, err = operatorAccountUnsuspend(caller, nil, db, nk, string(request)); err != nil {
		t.Fatalf("unsuspend: %v", err)
	}
	if err = db.QueryRowContext(ctx, `SELECT a.state,u.disable_time FROM jazebe_user_account a
		JOIN users u ON u.id=a.user_id WHERE a.user_id=$1`, target).Scan(&state, &disabled); err != nil || state != "active" || disabled.Unix() != 0 {
		t.Fatalf("restoration persisted incorrectly: %s %v %v", state, disabled, err)
	}
	if _, err = protected(targetCaller, nil, db, nil, "{}"); err != nil {
		t.Fatalf("reactivated account denied Jazebe RPC: %v", err)
	}
	if _, err = operatorAccountSuspend(caller, nil, db, nk, string(request)); err != nil {
		t.Fatal(err)
	}
	// Simulate a separate native Nakama ban after the Jazebe suspension.
	if _, err = db.ExecContext(ctx, `UPDATE users SET disable_time=disable_time+interval '1 second' WHERE id=$1`, target); err != nil {
		t.Fatal(err)
	}
	if _, err = operatorAccountUnsuspend(caller, nil, db, nk, string(request)); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("later native ban must be preserved: %v", err)
	}
	var audits int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_audit_log WHERE target_type='user' AND target_id=$1`, target).Scan(&audits); err != nil || audits != 3 {
		t.Fatalf("audit count %d: %v", audits, err)
	}
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM notification WHERE user_id=$1 AND content->>'subjectType'='account'`, target).Scan(&audits); err != nil || audits != 3 {
		t.Fatalf("critical notification count %d: %v", audits, err)
	}
}

func TestAccountRevocationRetryPostgres(t *testing.T) {
	dsn := os.Getenv("JAZEBE_DISPOSABLE_TEST_DB_URL")
	if dsn == "" {
		t.Skip("JAZEBE_DISPOSABLE_TEST_DB_URL not set; PostgreSQL NOT_RUN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	actor, target := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		for _, query := range []string{
			`DELETE FROM notification WHERE user_id=$1 AND content->>'subjectType'='account'`,
			`DELETE FROM jazebe_audit_log WHERE target_type='user' AND target_id=$1`,
			`DELETE FROM jazebe_user_account WHERE user_id=$1`,
			`DELETE FROM users WHERE id=$1`,
		} {
			if _, e := db.ExecContext(cleanup, query, target); e != nil {
				t.Errorf("fixture cleanup: %v", e)
			}
		}
		if _, e := db.ExecContext(cleanup, `DELETE FROM users WHERE id=$1`, actor); e != nil {
			t.Errorf("operator cleanup: %v", e)
		}
	}()
	for _, id := range []string{actor, target} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,username) VALUES($1,$2)`, id, "account-retry-"+id); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", actor)
	caller := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, actor)
	request, _ := json.Marshal(operatorAccountAction{UserID: target, Reason: "verified moderation finding"})
	nk := &accountRevocationNakama{listErr: context.DeadlineExceeded}
	if _, err = operatorAccountSuspend(caller, accountTestLogger{}, db, nk, string(request)); status.Code(err) != codes.Unavailable {
		t.Fatalf("revocation failure must remain visible: %v", err)
	}
	var pending bool
	if err = db.QueryRowContext(ctx, `SELECT session_revocation_pending FROM jazebe_user_account WHERE user_id=$1`, target).Scan(&pending); err != nil || !pending {
		t.Fatalf("pending revocation lost: pending=%v err=%v", pending, err)
	}
	if _, err = operatorAccountUnsuspend(caller, nil, db, nk, string(request)); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("pending revocation must block restoration: %v", err)
	}
	nk.listErr = nil
	if _, err = operatorAccountRevocationRetry(caller, nil, db, nk, `{"userId":"`+target+`"}`); err != nil {
		t.Fatalf("revocation retry: %v", err)
	}
	if err = db.QueryRowContext(ctx, `SELECT session_revocation_pending FROM jazebe_user_account WHERE user_id=$1`, target).Scan(&pending); err != nil || pending {
		t.Fatalf("pending flag after retry: pending=%v err=%v", pending, err)
	}
}

func TestAccountTransitionConcurrentPostgres(t *testing.T) {
	dsn := os.Getenv("JAZEBE_DISPOSABLE_TEST_DB_URL")
	if dsn == "" {
		t.Skip("JAZEBE_DISPOSABLE_TEST_DB_URL not set; PostgreSQL NOT_RUN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	actor, target := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		for _, query := range []string{
			`DELETE FROM notification WHERE user_id=$1 AND content->>'subjectType'='account'`,
			`DELETE FROM jazebe_audit_log WHERE target_type='user' AND target_id=$1`,
			`DELETE FROM jazebe_user_account WHERE user_id=$1`,
			`DELETE FROM users WHERE id=$1`,
		} {
			if _, e := db.ExecContext(cleanup, query, target); e != nil {
				t.Errorf("fixture cleanup: %v", e)
			}
		}
		if _, e := db.ExecContext(cleanup, `DELETE FROM users WHERE id=$1`, actor); e != nil {
			t.Errorf("operator cleanup: %v", e)
		}
	}()
	for _, id := range []string{actor, target} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,username) VALUES($1,$2)`, id, "concurrent-account-"+id); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", actor)
	caller := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, actor)
	request, _ := json.Marshal(operatorAccountAction{UserID: target, Reason: "verified moderation finding"})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for i := 0; i < 2; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, e := operatorAccountSuspend(caller, accountTestLogger{}, db, &accountRevocationNakama{}, string(request))
			results <- e
		}()
	}
	workers.Wait()
	close(results)
	var successes, conflicts int
	for e := range results {
		switch status.Code(e) {
		case codes.OK:
			successes++
		case codes.FailedPrecondition:
			conflicts++
		default:
			t.Fatalf("unexpected concurrent suspension error: %v", e)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent results: successes=%d conflicts=%d", successes, conflicts)
	}
	var audits, notices int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_audit_log WHERE target_type='user' AND target_id=$1`, target).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM notification WHERE user_id=$1 AND content->>'subjectType'='account'`, target).Scan(&notices); err != nil {
		t.Fatal(err)
	}
	if audits != 1 || notices != 1 {
		t.Fatalf("duplicate side effects: audits=%d notices=%d", audits, notices)
	}
}
