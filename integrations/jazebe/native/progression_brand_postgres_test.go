package main

import "testing"

// TestBrandProgressionPostgres covers the brand XP panel (source
// progression-brand.service.ts xpOverview, setLevels, setActivityXp).
func TestBrandProgressionPostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan", "pro")
	tech := f.user("tech")
	f.ok(f.owner, "jazebe_brand_team_invite", map[string]any{"brandId": f.brandID, "userId": tech, "fullName": "مدیر فنی", "role": "technical"})
	start, end := missionWindow()
	mission := f.ok(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "ماموریت XP",
		"detectionMethod": "platform_auto", "rewardPoints": 10, "rewardXp": 5, "startsAt": start, "endsAt": end,
		"code": map[string]any{"codeMode": "unique", "maxUses": 1}})["id"].(string)
	var template string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_game_template WHERE key='tap'`).Scan(&template); err != nil {
		t.Fatal(err)
	}
	game := f.ok(f.owner, "jazebe_brand_game_enable", map[string]any{"planetId": f.planetID, "templateId": template, "name": "ضربه"})["id"].(string)
	for user, xp := range map[string]int{f.users["fan"]: 50, f.users["pro"]: 400} {
		f.exec(`INSERT INTO jazebe_user_planet_stats(planet_id,user_id,total_xp,level,level_key) VALUES($1,$2,$3,1,'newcomer')
			ON CONFLICT (planet_id,user_id) DO UPDATE SET total_xp=$3,level=1,level_key='newcomer'`, f.planetID, user, xp)
	}
	ref := map[string]any{"planetId": f.planetID}

	overview := f.ok(tech, "jazebe_brand_progression_xp", ref)
	expectJSON(t, "overview", overview, map[string]any{"configured": false})
	if n := len(overview["activities"].([]any)); n != 2 {
		t.Errorf("activities %d, want mission and game", n)
	}
	f.fails(f.users["fan"], "jazebe_brand_progression_xp", ref, 7)
	levels := map[string]any{"planetId": f.planetID, "level2MinXp": 100, "level3MinXp": 300}
	f.fails(tech, "jazebe_brand_progression_set_levels", levels, 7)
	f.fails(f.owner, "jazebe_brand_progression_set_levels", map[string]any{"planetId": f.planetID, "level2MinXp": 300, "level3MinXp": 100}, 3)
	set := f.ok(f.owner, "jazebe_brand_progression_set_levels", levels)
	expectJSON(t, "configured", set, map[string]any{"configured": true})
	ladder := set["levels"].([]any)
	expectJSON(t, "level 1", ladder[0].(map[string]any), map[string]any{"level": 1, "members": 1, "titleFa": "تازه‌وارد"})
	expectJSON(t, "level 3", ladder[2].(map[string]any), map[string]any{"level": 3, "members": 1, "minXp": 300})
	if n := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action='planet.levels' AND target_id=$1 AND params->'rederived'='1'::jsonb`, f.planetID); n != 1 {
		t.Error("planet.levels audit missing")
	}
	f.ok(f.owner, "jazebe_brand_progression_set_levels", map[string]any{"planetId": f.planetID, "level2MinXp": 500, "level3MinXp": 900})
	if n := f.count(`SELECT level FROM jazebe_user_planet_stats WHERE planet_id=$1 AND user_id=$2`, f.planetID, f.users["pro"]); n != 3 {
		t.Errorf("raised thresholds lowered a reached level to %d", n)
	}

	xpMax := f.count(`SELECT mission_reward_xp_max FROM jazebe_platform_policy WHERE key='default'`)
	f.fails(f.owner, "jazebe_brand_progression_set_xp", map[string]any{"planetId": f.planetID, "items": []map[string]any{{"kind": "mission", "id": mission, "rewardXp": xpMax + 1}}}, 3)
	f.fails(f.owner, "jazebe_brand_progression_set_xp", map[string]any{"planetId": f.planetID, "items": []map[string]any{{"kind": "mission", "id": f.owner, "rewardXp": 1}}}, 5)
	f.fails(f.owner, "jazebe_brand_progression_set_xp", map[string]any{"planetId": f.planetID, "items": []map[string]any{{"kind": "badge", "id": mission, "rewardXp": 1}}}, 3)
	f.ok(f.owner, "jazebe_brand_progression_set_xp", map[string]any{"planetId": f.planetID, "items": []map[string]any{
		{"kind": "mission", "id": mission, "rewardXp": 12}, {"kind": "game", "id": game, "rewardXp": 7}}})
	if f.count(`SELECT reward_xp FROM jazebe_mission WHERE id=$1`, mission) != 12 || f.count(`SELECT reward_xp FROM jazebe_game WHERE id=$1`, game) != 7 {
		t.Error("activity XP not saved")
	}
}
