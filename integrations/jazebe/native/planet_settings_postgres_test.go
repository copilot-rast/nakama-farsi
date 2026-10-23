package main

import (
	"strings"
	"testing"
)

// TestPlanetSettingsPostgres covers brand planet settings against the
// platform policy, level re-derivation and brand self-suspension (source
// planet-settings.service.ts and planet-brand.service.ts).
func TestPlanetSettingsPostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan")
	tech, mkt := f.user("tech"), f.user("mkt")
	f.ok(f.owner, "jazebe_brand_team_invite", map[string]any{"brandId": f.brandID, "userId": tech, "fullName": "مدیر فنی", "role": "technical"})
	f.ok(f.owner, "jazebe_brand_team_invite", map[string]any{"brandId": f.brandID, "userId": mkt, "fullName": "مدیر بازاریابی", "role": "marketing"})
	ref := map[string]any{"planetId": f.planetID}
	set := func(fields map[string]any) map[string]any {
		out := map[string]any{"planetId": f.planetID}
		for k, v := range fields {
			out[k] = v
		}
		return out
	}

	got := f.ok(tech, "jazebe_brand_planet_settings_get", ref)
	f.fails(f.users["fan"], "jazebe_brand_planet_settings_get", ref, 7)
	if got["groupCreationEnabled"] != true {
		t.Errorf("default settings %v", got)
	}
	f.fails(tech, "jazebe_brand_planet_settings_update", set(map[string]any{"ticketPriceMax": 10}), 7)
	updated := f.ok(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"ticketPriceMax": 10, "groupCreationEnabled": false}))
	expectJSON(t, "updated", updated, map[string]any{"ticketPriceMax": 10, "groupCreationEnabled": false})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='planet.settings.update' AND target_id=$1 AND jsonb_array_length(params->'fields')=2`, f.planetID); n != 1 {
		t.Error("settings audit missing")
	}

	policyMax := f.count(`SELECT mission_reward_point_max FROM jazebe_platform_policy WHERE key='default'`)
	_, err := f.call(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"missionRewardPointMax": policyMax + 1, "ticketPriceMax": 20}))
	if err == nil || err.Code != 3 || !strings.Contains(err.Message, "PLANET_SETTINGS_OUT_OF_POLICY") || !strings.Contains(err.Message, "missionRewardPointMax") {
		t.Fatalf("out-of-policy update: %+v", err)
	}
	if n := f.count(`SELECT ticket_price_max FROM jazebe_planet_settings WHERE planet_id=$1`, f.planetID); n != 10 {
		t.Errorf("a rejected update saved ticketPriceMax=%d", n)
	}
	f.fails(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"groupMemberCapMin": 10, "groupMemberCapMax": 5}), 3)
	f.fails(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"secretKnob": 1}), 3)
	f.fails(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"allowedGroupModes": []string{}}), 3)

	// Levels: re-derived upward only.
	fan := f.users["fan"]
	f.exec(`INSERT INTO jazebe_user_planet_stats(planet_id,user_id,total_xp,level,level_key) VALUES($1,$2,150,1,'newcomer')
		ON CONFLICT (planet_id,user_id) DO UPDATE SET total_xp=150,level=1,level_key='newcomer'`, f.planetID, fan)
	ladder := func(two, three int) []map[string]any {
		return []map[string]any{{"level": 1, "key": "newcomer", "titleFa": "تازه‌وارد", "minXp": 0},
			{"level": 2, "key": "active", "titleFa": "عضو فعال", "minXp": two}, {"level": 3, "key": "veteran", "titleFa": "کهنه‌کار", "minXp": three}}
	}
	f.fails(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"xpLevelThresholds": ladder(500, 100)}), 3)
	f.ok(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"xpLevelThresholds": ladder(100, 500)}))
	if n := f.count(`SELECT level FROM jazebe_user_planet_stats WHERE planet_id=$1 AND user_id=$2 AND level_key='active'`, f.planetID, fan); n != 2 {
		t.Errorf("level after thresholds %d, want 2", n)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='planet.settings.update' AND target_id=$1 AND params->'levelsRederived'='1'::jsonb`, f.planetID); n != 1 {
		t.Error("rederive count not audited")
	}
	f.ok(mkt, "jazebe_brand_planet_settings_update", set(map[string]any{"xpLevelThresholds": ladder(200, 500)}))
	if n := f.count(`SELECT level FROM jazebe_user_planet_stats WHERE planet_id=$1 AND user_id=$2`, f.planetID, fan); n != 2 {
		t.Errorf("raising thresholds lowered the level to %d", n)
	}
	overview := f.ok(fan, "jazebe_planet_overview", ref)
	expectJSON(t, "overview level", overview["stats"].(map[string]any), map[string]any{"level": 2, "levelName": "عضو فعال"})

	// Brand self-suspension.
	f.fails(mkt, "jazebe_brand_planet_suspend", map[string]any{"planetId": f.planetID, "reason": "بازسازی"}, 7)
	out := f.ok(tech, "jazebe_brand_planet_suspend", map[string]any{"planetId": f.planetID, "reason": "بازسازی"})
	expectJSON(t, "suspend", out, map[string]any{"status": "suspended", "from": "active"})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='planet.suspend' AND actor_type='brand_admin' AND target_id=$1 AND params->>'by'='brand' AND params->>'reason'='بازسازی'`, f.planetID); n != 1 {
		t.Error("brand suspension audit missing")
	}
	f.fails(tech, "jazebe_brand_planet_suspend", ref, 9)
	f.fails(fan, "jazebe_planet_overview", ref, 7)
}
