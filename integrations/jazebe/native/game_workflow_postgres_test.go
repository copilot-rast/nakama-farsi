package main

import (
	"testing"
	"time"
)

// TestGameWorkflowPostgres enables a box game from the seeded template, plays
// one free round as a group member and verifies quota, timing guards, ticket
// issuance into the linked lottery and the group game.played activity.
func TestGameWorkflowPostgres(t *testing.T) {
	f := newJazebeFixture(t, "player", "loner")
	u := f.users
	now := time.Now().UTC()
	group := f.ok(u["player"], "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "بازیکنان", "mode": "open", "cap": 4})["id"].(string)
	lottery := f.ok(f.owner, "jazebe_lottery_draft_create", map[string]any{"planetId": f.planetID, "title": "قرعه بازی",
		"startsAt": now.Add(-time.Hour).Format(time.RFC3339), "drawAt": now.Add(24 * time.Hour).Format(time.RFC3339), "ticketPricePoints": 1,
		"prizes": []map[string]any{{"title": "جایزه", "type": "cash", "valueRial": "1000", "priority": 1}}})["id"].(string)
	f.ok(f.owner, "jazebe_lottery_activate", map[string]any{"planetId": f.planetID, "lotteryId": lottery})

	var templateID string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_game_template WHERE key='box'`).Scan(&templateID); err != nil {
		t.Fatalf("seeded box template: %v", err)
	}
	// Source EnableGameDto: enable takes the template and name and starts paused;
	// the entry model, quotas, prize and lottery are set with PATCH (update).
	enable := map[string]any{"planetId": f.planetID, "templateId": templateID, "name": "جعبه شانس"}
	f.fails(u["player"], "jazebe_brand_game_enable", enable, 7)
	enabled := f.ok(f.owner, "jazebe_brand_game_enable", enable)
	gameID, _ := enabled["id"].(string)
	if gameID == "" {
		if g, ok := enabled["game"].(map[string]any); ok {
			gameID, _ = g["id"].(string)
		}
	}
	if gameID == "" {
		t.Fatalf("enable returned no game id: %v", enabled)
	}
	f.fails(f.owner, "jazebe_brand_game_status", map[string]any{"planetId": f.planetID, "gameId": gameID, "status": "active"}, 3)
	update := map[string]any{"planetId": f.planetID, "gameId": gameID, "entryModel": "free", "freeIndividualCount": 1,
		"freeIndividualPeriod": "daily", "prizeTicketMin": 1, "prizeTicketMax": 3, "targetLotteryId": lottery}
	f.fails(u["player"], "jazebe_brand_game_update", update, 7)
	f.ok(f.owner, "jazebe_brand_game_update", update)
	f.ok(f.owner, "jazebe_brand_game_status", map[string]any{"planetId": f.planetID, "gameId": gameID, "status": "active"})
	play := map[string]any{"planetId": f.planetID, "gameId": gameID}

	f.fails(u["loner"], "jazebe_game_start", play, 7)
	f.ok(u["player"], "jazebe_game_start", play)
	if n := f.count(`SELECT used FROM jazebe_game_quota WHERE game_id=$1 AND scope='user' AND user_id=$2`, gameID, u["player"]); n != 1 {
		t.Errorf("user quota used=%d, want 1", n)
	}
	claim := map[string]any{"planetId": f.planetID, "gameId": gameID, "box": 0}
	f.fails(u["player"], "jazebe_game_claim", claim, 3)
	time.Sleep(time.Duration(gameMinRoundMs+200) * time.Millisecond)
	result := f.ok(u["player"], "jazebe_game_claim", claim)
	tickets := int(result["tickets"].(float64))
	if tickets < 1 || tickets > 3 {
		t.Fatalf("box round paid %d tickets outside 1..3", tickets)
	}
	if n := f.count(`SELECT count(*) FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND creator_user_id=$2 AND group_id=$3`, lottery, u["player"], group); n != tickets {
		t.Errorf("issued %d lottery tickets, claim reported %d", n, tickets)
	}
	var session string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_game_session WHERE game_id=$1 AND user_id=$2`, gameID, u["player"]).Scan(&session); err != nil {
		t.Fatal(err)
	}
	actor, payload := f.activity(group, "game.played")
	expectJSON(t, "game.played", payload, map[string]any{"gameId": gameID, "sessionId": session, "tickets": tickets})
	if actor != u["player"] {
		t.Errorf("game.played actor %s", actor)
	}
	f.fails(u["player"], "jazebe_game_claim", claim, 5)
	f.fails(u["player"], "jazebe_game_start", play, 9)
}
