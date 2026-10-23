package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

// TestAccountPurgePostgres covers the operator purge (source
// admin-support.service.ts) and the guard on Nakama's client account deletion.
func TestAccountPurgePostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan", "bystander")
	fan := f.users["fan"]
	super := f.user("super")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator+","+super)
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", super)
	t.Setenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS", "")

	group := f.ok(fan, "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "گروه پاک‌سازی", "mode": "open", "cap": 3})["id"].(string)
	f.ok(f.operator, "jazebe_wallet_adjust", map[string]any{"planetId": f.planetID, "userId": fan, "direction": "credit", "amount": 70,
		"reason": "جایزه آزمایشی", "idempotencyKey": "purge-credit-1"})
	f.ok(fan, "jazebe_settings_preferences_update", map[string]any{"categories": map[string]any{"news": false}})
	f.ok(fan, "jazebe_profile_setup", map[string]any{"displayName": "فن" + fan[:4], "avatarIndex": 3})
	f.exec(`INSERT INTO jazebe_user_verification(user_id,status,verified_at) VALUES($1,'verified',now())`, fan)
	f.exec(`INSERT INTO user_device(id,user_id) VALUES($1,$2)`, "device-"+fan, fan)
	f.exec(`UPDATE users SET email=$2,custom_id=$3 WHERE id=$1`, fan, "fan-"+fan[:8]+"@example.com", "custom-"+fan)
	f.notice(fan, "purge-n1", "اعلان", map[string]any{"category": "critical"}, time.Now().UTC())

	// Client deletion of a user with Jazebe data is refused cleanly and changes nothing.
	hookCtx := context.WithValue(f.ctx, runtime.RUNTIME_CTX_USER_ID, fan)
	err := refuseJazebeAccountDeletion(hookCtx, f.logger, f.db, f.nk)
	var re *runtime.Error
	if !errors.As(err, &re) || re.Code != 9 {
		t.Fatalf("deletion hook returned %v, want code 9", err)
	}
	if n := f.count(`SELECT count(*) FROM users WHERE id=$1`, fan); n != 1 {
		t.Fatal("deletion probe was not rolled back")
	}
	plain := f.user("no-jazebe-data")
	if err := refuseJazebeAccountDeletion(context.WithValue(f.ctx, runtime.RUNTIME_CTX_USER_ID, plain), f.logger, f.db, f.nk); err != nil {
		t.Fatalf("user without Jazebe data must stay deletable: %v", err)
	}
	if n := f.count(`SELECT count(*) FROM users WHERE id=$1`, plain); n != 1 {
		t.Fatal("deletion probe deleted the user")
	}

	purge := map[string]any{"userId": fan, "reason": "درخواست قانونی حذف"}
	f.fails(f.operator, "jazebe_operator_user_purge", purge, 7)
	f.fails(super, "jazebe_operator_user_purge", map[string]any{"userId": fan, "reason": "x"}, 3)
	f.fails(super, "jazebe_operator_user_purge", purge, 9)
	f.fails(super, "jazebe_operator_user_purge", map[string]any{"userId": super, "reason": "خودم"}, 7)
	f.fails(super, "jazebe_operator_user_purge", map[string]any{"userId": f.owner[:8] + "-0000-4000-8000-000000000000", "reason": "ناشناس"}, 5)
	f.ok(f.operator, "jazebe_operator_user_suspend", map[string]any{"userId": fan, "reason": "در انتظار حذف"})

	result := f.ok(super, "jazebe_operator_user_purge", purge)
	got := map[string]bool{}
	for _, p := range result["purged"].([]any) {
		got[p.(string)] = true
	}
	for _, want := range []string{"kyc", "notifications", "notification_preferences", "devices", "identity", "sessions"} {
		if !got[want] {
			t.Errorf("purged list missing %s: %v", want, result["purged"])
		}
	}
	var username, display string
	var email, custom any
	var disabled time.Time
	if err := f.db.QueryRowContext(f.ctx, `SELECT username,display_name,email,custom_id,disable_time FROM users WHERE id=$1`, fan).
		Scan(&username, &display, &email, &custom, &disabled); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(username, "purged-") || !strings.HasPrefix(display, "حذف‌شده-") || email != nil || custom != nil || disabled.Year() < 2000 {
		t.Errorf("identity not anonymized: %s %s %v %v %v", username, display, email, custom, disabled)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_user_profile WHERE user_id=$1 AND display_name=$2 AND avatar_preset_index=0`, fan, display); n != 1 {
		t.Error("Jazebe profile not anonymized")
	}
	for table, query := range map[string]string{
		"user_device":                    `SELECT count(*) FROM user_device WHERE user_id=$1`,
		"jazebe_user_verification":       `SELECT count(*) FROM jazebe_user_verification WHERE user_id=$1`,
		"jazebe notifications":           `SELECT count(*) FROM notification WHERE user_id=$1 AND code=1001`,
		"jazebe_notification_preference": `SELECT count(*) FROM jazebe_notification_preference WHERE user_id=$1`,
	} {
		if n := f.count(query, fan); n != 0 {
			t.Errorf("%s retained %d rows", table, n)
		}
	}
	// Business records are retained under the anonymized user, as in the source.
	if n := f.count(`SELECT balance FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, f.planetID, fan); n != 70 {
		t.Errorf("wallet balance %d after purge, want 70", n)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2`, group, fan); n != 1 {
		t.Error("group membership history was removed")
	}
	_, actor, params := f.audit("account.purge", fan)
	if actor != super || params["reason"] != "درخواست قانونی حذف" {
		t.Errorf("purge audit %s %v", actor, params)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_user_account WHERE user_id=$1 AND purged_at IS NOT NULL AND purged_by_operator_id=$2 AND NOT session_revocation_pending`, fan, super); n != 1 {
		t.Error("purge marker or revocation state wrong")
	}
	f.fails(super, "jazebe_operator_user_purge", purge, 6)
	f.fails(fan, "jazebe_profile_overview", map[string]any{"planetId": f.planetID}, 7)
	if n := f.count(`SELECT count(*) FROM users WHERE id=$1`, f.users["bystander"]); n != 1 {
		t.Error("purge touched another user")
	}
}
