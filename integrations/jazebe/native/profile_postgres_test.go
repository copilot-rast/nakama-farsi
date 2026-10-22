package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This test only runs when a caller explicitly provides a disposable database
// with the Nakama and Jazebe migrations already applied. All fixture writes
// stay inside a transaction which rolls back, including failed assertions.
func TestProfilePostgresConstraints(t *testing.T) {
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
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	var table sql.NullString
	if err = db.QueryRowContext(ctx, `SELECT to_regclass('public.jazebe_user_profile')::text`).Scan(&table); err != nil || !table.Valid {
		t.Fatalf("apply Jazebe profile migration to disposable database: %v", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	first, second := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	for _, id := range []string{first, second} {
		if _, err = tx.ExecContext(ctx, `INSERT INTO users(id,username) VALUES($1,$2)`, id, "profile-test-"+id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_profile(user_id,display_name,avatar_preset_index)
		VALUES($1,'Mehrsa',0)`, first); err != nil {
		t.Fatal(err)
	}
	// A savepoint keeps the transaction usable after the expected constraint error.
	if _, err = tx.ExecContext(ctx, `SAVEPOINT before_duplicate`); err != nil {
		t.Fatal(err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_profile(user_id,display_name,avatar_preset_index)
		VALUES($1,'mehrsa',0)`, second)
	var pg *pgconn.PgError
	if !errors.As(err, &pg) || pg.Code != "23505" {
		t.Fatalf("case-insensitive duplicate was not rejected: %v", err)
	}
	if _, err = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT before_duplicate`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(ctx, `SAVEPOINT before_avatar`); err != nil {
		t.Fatal(err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_profile(user_id,display_name,avatar_kind,avatar_preset_index)
		VALUES($1,'Other','upload',0)`, second)
	if !errors.As(err, &pg) || pg.Code != "23514" {
		t.Fatalf("incompatible avatar state was not rejected: %v", err)
	}
	if _, err = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT before_avatar`); err != nil {
		t.Fatal(err)
	}
}

// Uses only a database explicitly identified as disposable by the caller.
// Unlike constraint tests, this exercises the registered RPC handlers against
// committed rows, including the operator's review and audited resolution.
func TestProfileReportPostgresWorkflow(t *testing.T) {
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
	owner, reporter, subject, operator := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	brand, planet := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		for _, q := range []struct {
			sql  string
			args []any
		}{
			{`DELETE FROM jazebe_audit_log WHERE target_type='user_report' AND target_id IN (SELECT id FROM jazebe_user_report WHERE planet_id=$1)`, []any{planet}},
			{`DELETE FROM jazebe_user_report WHERE planet_id=$1`, []any{planet}},
			{`DELETE FROM jazebe_planet WHERE id=$1`, []any{planet}},
			{`DELETE FROM jazebe_brand WHERE id=$1`, []any{brand}},
			{`DELETE FROM users WHERE id IN ($1,$2,$3,$4)`, []any{owner, reporter, subject, operator}},
		} {
			if _, e := db.ExecContext(cleanup, q.sql, q.args...); e != nil {
				t.Errorf("fixture cleanup: %v", e)
			}
		}
	}()
	for _, id := range []string{owner, reporter, subject, operator} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,username) VALUES($1,$2)`, id, "report-workflow-"+id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO jazebe_brand(id,owner_id,name) VALUES($1,$2,'Report workflow')`, brand, owner); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO jazebe_planet(id,brand_id,name,status) VALUES($1,$2,'Report workflow','active')`, planet, brand); err != nil {
		t.Fatal(err)
	}
	caller := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, reporter)
	request, _ := json.Marshal(profileReportInput{PlanetID: planet, UserID: subject, Reason: "inappropriate"})
	first, err := profileReport(caller, nil, db, nil, string(request))
	if err != nil {
		t.Fatal(err)
	}
	second, err := profileReport(caller, nil, db, nil, string(request))
	if err != nil || first != second {
		t.Fatalf("open report replay changed identity: %s %s %v", first, second, err)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err = json.Unmarshal([]byte(first), &created); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", operator)
	admin := context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, operator)
	review, _ := json.Marshal(operatorReportInput{ReportID: created.ID})
	if _, err = operatorReportReview(admin, nil, db, nil, string(review)); err != nil {
		t.Fatal(err)
	}
	if _, err = operatorReportReview(admin, nil, db, nil, string(review)); err != nil {
		t.Fatalf("source permits reviewing an already reviewing report: %v", err)
	}
	resolve, _ := json.Marshal(operatorReportInput{ReportID: created.ID, Outcome: "dismissed", Reason: "not substantiated"})
	if _, err = operatorReportResolve(admin, nil, db, nil, string(resolve)); err != nil {
		t.Fatal(err)
	}
	if _, err = operatorReportResolve(admin, nil, db, nil, string(resolve)); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("second resolution: %v", err)
	}
	var state string
	var audits int
	if err = db.QueryRowContext(ctx, `SELECT status FROM jazebe_user_report WHERE id=$1`, created.ID).Scan(&state); err != nil || state != "dismissed" {
		t.Fatalf("report state %s: %v", state, err)
	}
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_audit_log WHERE target_type='user_report' AND target_id=$1`, created.ID).Scan(&audits); err != nil || audits != 3 {
		t.Fatalf("report audit count %d: %v", audits, err)
	}
}

// The profile report key is scoped to a planet and only to the open state.
// A report under review may coexist with a new open report, as in Jazebe.
func TestProfileReportPostgresConstraints(t *testing.T) {
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
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	owner, reporter, subject := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	for _, id := range []string{owner, reporter, subject} {
		if _, err = tx.ExecContext(ctx, `INSERT INTO users(id,username) VALUES($1,$2)`, id, "report-test-"+id); err != nil {
			t.Fatal(err)
		}
	}
	brand, planet1, planet2 := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_brand(id,owner_id,name) VALUES($1,$2,'Report test')`, brand, owner); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct{ id, name string }{{planet1, "First"}, {planet2, "Second"}} {
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_planet(id,brand_id,name) VALUES($1,$2,$3)`, row.id, brand, row.name); err != nil {
			t.Fatal(err)
		}
	}
	insert := func(planet string) error {
		_, e := tx.ExecContext(ctx, `INSERT INTO jazebe_user_report(id,planet_id,reporter_user_id,subject_type,subject_id,reason)
			VALUES($1,$2,$3,'profile',$4,'inappropriate')`, uuid.Must(uuid.NewV4()).String(), planet, reporter, subject)
		return e
	}
	if err = insert(planet1); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(ctx, `SAVEPOINT duplicate_open`); err != nil {
		t.Fatal(err)
	}
	err = insert(planet1)
	var pg *pgconn.PgError
	if !errors.As(err, &pg) || pg.Code != "23505" {
		t.Fatalf("duplicate open report accepted: %v", err)
	}
	if _, err = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT duplicate_open`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_report SET status='reviewing' WHERE planet_id=$1 AND reporter_user_id=$2`, planet1, reporter); err != nil {
		t.Fatal(err)
	}
	if err = insert(planet1); err != nil {
		t.Fatalf("new report while earlier is reviewing: %v", err)
	}
	if err = insert(planet2); err != nil {
		t.Fatalf("same subject on second planet: %v", err)
	}
}
