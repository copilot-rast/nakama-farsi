package main

import (
	"testing"
	"time"
)

// TestLotteryWorkflowPostgres runs a planet-points lottery end to end: draft
// and edit rules, activation, tickets earned from a mission (with the group
// mission.completed activity), a brand draw split across group members, KYC
// gating, later release for a newly verified member, and operator replay.
func TestLotteryWorkflowPostgres(t *testing.T) {
	f := newJazebeFixture(t, "alpha", "beta", "stranger")
	u := f.users
	t.Setenv("LOTTERY_PRIZE_RIAL_PER_POINT", "1000")
	now := time.Now().UTC()
	group := f.ok(u["alpha"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "برندگان", "mode": "open", "cap": 5})["id"].(string)
	f.ok(u["beta"], "jazebe_group_join", map[string]any{"planetId": f.planetID, "groupId": group})
	f.exec(`INSERT INTO jazebe_user_verification(user_id,status,verified_at) VALUES($1,'verified',now())`, u["alpha"])

	draft := map[string]any{"planetId": f.planetID, "title": "قرعه‌کشی بزرگ", "startsAt": now.Add(-time.Hour).Format(time.RFC3339),
		"drawAt": now.Add(time.Hour).Format(time.RFC3339), "ticketPricePoints": 1,
		"prizes": []map[string]any{{"title": "جایزه امتیازی", "type": "planet_points", "valueRial": "100000", "priority": 1}}}
	f.fails(u["alpha"], "jazebe_lottery_draft_create", draft, 7)
	bad := map[string]any{}
	for k, v := range draft {
		bad[k] = v
	}
	bad["drawAt"] = now.Add(-time.Minute).Format(time.RFC3339)
	f.fails(f.owner, "jazebe_lottery_draft_create", bad, 3)
	lottery := f.ok(f.owner, "jazebe_lottery_draft_create", draft)["id"].(string)
	lref := map[string]any{"planetId": f.planetID, "lotteryId": lottery}

	f.ok(f.owner, "jazebe_lottery_brand_update", map[string]any{"planetId": f.planetID, "lotteryId": lottery, "drawAt": now.Add(2 * time.Hour).Format(time.RFC3339)})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='lottery.update' AND target_id=$1 AND params->'structural'='true'::jsonb`, lottery); n != 1 {
		t.Errorf("structural update audit rows %d", n)
	}
	f.fails(u["alpha"], "jazebe_lottery_activate", lref, 7)
	f.ok(f.owner, "jazebe_lottery_activate", lref)
	_, _, params := f.audit("lottery.activate", lottery)
	expectJSON(t, "activate", params, map[string]any{"adoptedTickets": 0})
	f.fails(f.owner, "jazebe_lottery_brand_update", map[string]any{"planetId": f.planetID, "lotteryId": lottery, "ticketPricePoints": 2}, 9)
	f.ok(f.owner, "jazebe_lottery_brand_update", map[string]any{"planetId": f.planetID, "lotteryId": lottery, "title": "قرعه‌کشی ویژه"})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='lottery.update' AND target_id=$1 AND params->'structural'='false'::jsonb`, lottery); n != 1 {
		t.Errorf("non-structural update audit rows %d", n)
	}

	// Tickets from a mission completion.
	start, end := missionWindow()
	mission := f.ok(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "بلیت رایگان",
		"detectionMethod": "platform_auto", "rewardPoints": 5, "rewardTickets": 2, "targetLotteryId": lottery,
		"startsAt": start, "endsAt": end, "code": map[string]any{"codeMode": "unique", "maxUses": 1}})["id"].(string)
	f.ok(f.owner, "jazebe_code_mission_activate", map[string]any{"planetId": f.planetID, "missionId": mission})
	codes := f.ok(f.owner, "jazebe_code_mission_generate", map[string]any{"planetId": f.planetID, "missionId": mission, "count": 3})["codes"].([]any)
	codeAt := func(i int) string {
		if m, ok := codes[i].(map[string]any); ok {
			return m["code"].(string)
		}
		return codes[i].(string)
	}
	f.ok(u["alpha"], "jazebe_code_mission_redeem", map[string]any{"planetId": f.planetID, "missionId": mission, "code": codeAt(0)})
	f.ok(u["beta"], "jazebe_code_mission_redeem", map[string]any{"planetId": f.planetID, "missionId": mission, "code": codeAt(1)})
	f.fails(u["stranger"], "jazebe_code_mission_redeem", map[string]any{"planetId": f.planetID, "missionId": mission, "code": codeAt(2)}, 7)
	if n := f.count(`SELECT count(*) FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND group_id=$2`, lottery, group); n != 4 {
		t.Errorf("group holds %d tickets, want 4", n)
	}
	actor, payload := f.activity(group, "mission.completed")
	expectJSON(t, "mission.completed", payload, map[string]any{"missionId": mission, "tickets": 2})
	if actor == "" {
		t.Error("mission.completed activity has no actor")
	}

	// Draw as the brand: KYC-verified alpha is paid, beta is held.
	f.fails(f.owner, "jazebe_lottery_brand_draw", lref, 9)
	f.exec(`UPDATE jazebe_lottery SET draw_at=now()-interval '1 second' WHERE id=$1`, lottery)
	f.fails(u["alpha"], "jazebe_lottery_brand_draw", lref, 7)
	f.ok(f.owner, "jazebe_lottery_brand_draw", lref)
	if n := f.count(`SELECT count(*) FROM jazebe_lottery WHERE id=$1 AND status='finished'`, lottery); n != 1 {
		t.Fatal("lottery not finished after draw")
	}
	var alphaShare, betaShare string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_lottery_winner_share WHERE user_id=$1 AND payout_status='paid' AND NOT blocked_by_kyc`, u["alpha"]).Scan(&alphaShare); err != nil {
		t.Fatalf("alpha share: %v", err)
	}
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_lottery_winner_share WHERE user_id=$1 AND payout_status='pending_verification' AND blocked_by_kyc`, u["beta"]).Scan(&betaShare); err != nil {
		t.Fatalf("beta share: %v", err)
	}
	balance := func(user string) int {
		return f.count(`SELECT COALESCE(sum(balance),0) FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2`, f.planetID, user)
	}
	if b := balance(u["alpha"]); b != 5+50 {
		t.Errorf("alpha balance %d, want 55 (mission 5 + prize 50)", b)
	}
	if b := balance(u["beta"]); b != 5 {
		t.Errorf("beta balance %d before KYC, want 5", b)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='lottery.payout' AND target_id=$1
		AND params->>'userId'=$2 AND params->>'points'='50' AND params->'rialPerPoint'='1000'::jsonb AND params->>'shareId'=$3`, lottery, u["alpha"], alphaShare); n != 1 {
		t.Errorf("alpha payout audit rows %d", n)
	}

	// KYC release (no RPC caller yet: exercised directly in a transaction).
	f.exec(`INSERT INTO jazebe_user_verification(user_id,status,verified_at) VALUES($1,'verified',now())`, u["beta"])
	tx, err := f.db.BeginTx(f.ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	released, err := releaseVerifiedLotteryShares(f.ctx, tx, u["beta"], time.Now().UTC())
	if err != nil {
		tx.Rollback()
		t.Fatalf("release: %v", err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if released != 1 {
		t.Errorf("released %d shares, want 1", released)
	}
	if b := balance(u["beta"]); b != 5+50 {
		t.Errorf("beta balance %d after release, want 55", b)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='lottery.payout' AND params->>'shareId'=$1 AND params->'rialPerPoint'='1000'::jsonb`, betaShare); n != 1 {
		t.Errorf("beta payout audit rows %d", n)
	}
	tx, _ = f.db.BeginTx(f.ctx, nil)
	again, err := releaseVerifiedLotteryShares(f.ctx, tx, u["beta"], time.Now().UTC())
	tx.Commit()
	if err != nil || again != 0 || balance(u["beta"]) != 55 {
		t.Errorf("second release: released=%d err=%v balance=%d", again, err, balance(u["beta"]))
	}

	// Operator replay reproduces the recorded draw.
	f.fails(f.owner, "jazebe_lottery_admin_replay", map[string]any{"lotteryId": lottery}, 7)
	replay := f.ok(f.operator, "jazebe_lottery_admin_replay", map[string]any{"lotteryId": lottery})
	if replay["verdict"] != "match" {
		t.Errorf("replay verdict %v", replay["verdict"])
	}
	actorType, actorID, params := f.audit("lottery.replay", lottery)
	if actorType != "operator" || actorID != f.operator {
		t.Errorf("replay attributed to %s/%s", actorType, actorID)
	}
	expectJSON(t, "replay", params, map[string]any{"verdict": "match", "shareMismatches": 0})
}
