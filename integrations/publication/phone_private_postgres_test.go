package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

func (h *harness) metaPhone(id string) bool {
	h.t.Helper()
	var has bool
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT metadata ? 'phone' FROM users WHERE id=$1`, id).Scan(&has); err != nil {
		h.t.Fatal(err)
	}
	return has
}

func (h *harness) tablePhone(id string) string {
	h.t.Helper()
	var phone string
	_ = h.s.db.QueryRowContext(h.ctx, `SELECT phone FROM publication_phone_account WHERE user_id=$1`, id).Scan(&phone)
	return phone
}

func TestConsoleCreatedPlayerKeepsPhoneOutOfMetadataPostgres(t *testing.T) {
	h, _ := otpHarness(t)
	out := h.mustAPI(http.StatusCreated, http.MethodPost, "/players", map[string]any{"kind": "phone", "phone": "09121230001"})
	id := out["userId"].(string)
	if out["phone"] != "+989121230001" || h.tablePhone(id) != "+989121230001" || h.metaPhone(id) {
		t.Fatalf("created %v table %q metadata phone %v", out, h.tablePhone(id), h.metaPhone(id))
	}
	if code, res := h.api(http.MethodPost, "/players", map[string]any{"kind": "phone", "phone": "+989121230001"}); code != http.StatusConflict || res["error"] != "PHONE_TAKEN" {
		t.Fatalf("taken = %d %v", code, res)
	}
	me, err := h.s.rpcPhone(callerCtx(id, ""), id, `{}`)
	if err != nil || me.(map[string]any)["phone"] != "+989121230001" {
		t.Fatalf("own phone = %v %v", me, err)
	}
	other := h.user(nil, "", "fa")
	if me, _ := h.s.rpcPhone(callerCtx(other, ""), other, `{}`); me.(map[string]any)["phone"] != "" {
		t.Fatalf("another player's own phone = %v", me)
	}
}

func TestSMSTargetsUseThePrivatePhoneOnlyPostgres(t *testing.T) {
	h := newHarness(t)
	private := h.user(map[string]any{"phone": "09121230002"}, "", "fa")
	stale := h.user(nil, "", "fa")
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE users SET metadata='{"phone":"09121230003"}' WHERE id=$1`, stale); err != nil {
		t.Fatal(err)
	}
	if h.metaPhone(private) {
		t.Fatal("helper must not write metadata phone")
	}
	preview := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/preview-target?mode=users&userIds="+private+","+stale, nil)
	if num(preview["count"]) != 2 || num(preview["withPhone"]) != 1 {
		t.Fatalf("preview %v", preview)
	}
}

func TestPhoneMigrationMovesAndStripsMetadataPostgres(t *testing.T) {
	h := newHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE publication_phone_account, publication_phone_backfill_issue`); err != nil {
		t.Fatal(err)
	}
	mk := func(meta string, age int) string {
		id := h.user(nil, "", "fa")
		if _, err := h.s.db.ExecContext(h.ctx, `UPDATE users SET metadata=$2::jsonb, create_time=now() - make_interval(days => $3) WHERE id=$1`, id, meta, age); err != nil {
			t.Fatal(err)
		}
		return id
	}
	own := func(id, phone string) {
		if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO publication_phone_account (phone, user_id) VALUES ($1,$2)`, phone, id); err != nil {
			t.Fatal(err)
		}
	}
	first := mk(`{"phone":"09121240001","country":"IR"}`, 10)
	second := mk(`{"phone":"+989121240001"}`, 5)
	same := mk(`{"phone":"+989121240002"}`, 5)
	own(same, "+989121240002")
	other := mk(`{"phone":"+989121240003"}`, 5)
	own(other, "+989121240004")
	bad := mk(`{"phone":"not a number"}`, 5)
	owner := h.user(nil, "", "fa")
	own(owner, "+989121240005")
	taken := mk(`{"phone":"+989121240005"}`, 5)
	number := mk(`{"phone":9121240006}`, 5)
	persian := mk(`{"phone":"۰۹۱۲ ۱۲۴ ۰۰۰۷"}`, 5)
	plain := mk(`{"country":"DE"}`, 5)

	raw, err := os.ReadFile("../../migrate/sql/20270201000000-publication-phone-private.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.SplitN(strings.SplitN(string(raw), "-- +migrate Down", 2)[0], "-- +migrate Up", 2)[1]
	up = strings.NewReplacer("-- +migrate StatementBegin", "", "-- +migrate StatementEnd", "").Replace(up)
	for i := 0; i < 2; i++ {
		if _, err := h.s.db.ExecContext(h.ctx, up); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}

	want := map[string]string{first: "+989121240001", second: "", same: "+989121240002", other: "+989121240004", bad: "", owner: "+989121240005", taken: "", number: "+989121240006", persian: "+989121240007", plain: ""}
	for id, phone := range want {
		if got := h.tablePhone(id); got != phone {
			t.Fatalf("%s phone = %q, want %q", id, got, phone)
		}
		if h.metaPhone(id) {
			t.Fatalf("%s still has metadata phone", id)
		}
	}
	var country string
	_ = h.s.db.QueryRowContext(h.ctx, `SELECT metadata->>'country' FROM users WHERE id=$1`, first).Scan(&country)
	if country != "IR" {
		t.Fatalf("other keys must stay, country = %q", country)
	}
	issues := map[string]string{}
	rows, err := h.s.db.QueryContext(h.ctx, `SELECT user_id::text, reason || ':' || coalesce(owner_id::text,'') FROM publication_phone_backfill_issue`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id, v string
		if err := rows.Scan(&id, &v); err != nil {
			t.Fatal(err)
		}
		issues[id] = v
	}
	rows.Close()
	wantIssues := map[string]string{second: "phone_taken:" + first, other: "user_has_other_phone:" + other, bad: "invalid:", taken: "phone_taken:" + owner}
	if len(issues) != len(wantIssues) {
		t.Fatalf("issues %v", issues)
	}
	for id, v := range wantIssues {
		if issues[id] != v {
			t.Fatalf("issue for %s = %q, want %q (all %v)", id, issues[id], v, issues)
		}
	}
	var leftover int
	_ = h.s.db.QueryRowContext(h.ctx, `SELECT count(*) FROM users WHERE metadata ? 'phone'`).Scan(&leftover)
	if leftover != 0 {
		t.Fatalf("%d users still have metadata phone", leftover)
	}
}
