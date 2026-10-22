package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOperatorConstellationPostgres(t *testing.T) {
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
	operator, user, brand, planet, wallet := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String(),
		uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		for _, q := range []struct {
			query string
			args  []any
		}{
			{`DELETE FROM jazebe_user_planet_stats WHERE planet_id=$1`, []any{planet}},
			{`DELETE FROM jazebe_point_wallet WHERE id=$1`, []any{wallet}},
			{`DELETE FROM jazebe_planet_membership WHERE planet_id=$1`, []any{planet}},
			{`DELETE FROM jazebe_planet WHERE id=$1`, []any{planet}},
			{`DELETE FROM jazebe_brand WHERE id=$1`, []any{brand}},
			{`DELETE FROM users WHERE id=$1 OR id=$2`, []any{user, operator}},
		} {
			if _, e := db.ExecContext(cleanup, q.query, q.args...); e != nil {
				t.Errorf("fixture cleanup: %v", e)
			}
		}
	}()
	for _, id := range []string{operator, user} {
		if _, err = db.ExecContext(ctx, `INSERT INTO users(id,username) VALUES($1,$2)`, id, "constellation-"+id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO jazebe_brand(id,owner_id,name) VALUES($1,$2,'Support brand')`, brand, operator); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO jazebe_planet(id,brand_id,name,status) VALUES($1,$2,'Support planet','active')`, planet, brand); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO jazebe_planet_membership(planet_id,user_id,join_path) VALUES($1,$2,'join_button')`, planet, user); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO jazebe_point_wallet(id,planet_id,user_id,balance) VALUES($1,$2,$3,25)`, wallet, planet, user); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id,total_xp,level,medals_count)
		VALUES($1,$2,100,2,1)`, planet, user); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", operator)
	request := `{"userId":"` + user + `"}`
	if _, err = operatorUserConstellation(context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, user), nil, db, nil, request); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("ordinary user accessed support view: %v", err)
	}
	out, err := operatorUserConstellation(context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, operator), nil, db, nil, request)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		PlanetID string `json:"planetId"`
		Balance  int64  `json:"balance"`
		TotalXP  int64  `json:"totalXp"`
		Level    int    `json:"level"`
		Medals   int    `json:"medals"`
	}
	if err = json.Unmarshal([]byte(out), &rows); err != nil || len(rows) != 1 || rows[0].PlanetID != planet || rows[0].Balance != 25 || rows[0].TotalXP != 100 || rows[0].Level != 2 || rows[0].Medals != 1 {
		t.Fatalf("support projection: %s %v", out, err)
	}
}
