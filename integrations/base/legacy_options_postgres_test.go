package main

import (
	"context"
	"testing"
	"time"
)

func TestSocialDayStartAndBotRequestAutoDecline(t *testing.T) {
	h := newHarness(t)
	resetFriendTables(h)
	h.t.Cleanup(func() {
		h.s.db.Exec(`UPDATE base_friend_rules SET day_zone='', day_start_minute=0, bot_decline_hours=0 WHERE id=1`)
	})
	if c, out := h.api("PUT", "/friend-rules", map[string]any{"dayZone": "Nowhere/Land"}); c != 400 || out["field"] != "dayZone" {
		t.Fatalf("bad zone = %d %v", c, out)
	}
	if c, out := h.api("PUT", "/friend-rules", map[string]any{"botDeclineHours": 9000}); c != 400 || out["field"] != "botDeclineHours" {
		t.Fatalf("bad hours = %d %v", c, out)
	}
	rules := h.mustAPI(200, "PUT", "/friend-rules", map[string]any{"dailyRequests": 0, "dayZone": "UTC", "dayStartMinute": 90, "botDeclineHours": 5})
	if rules["dayZone"] != "UTC" || rules["dayStartMinute"] != float64(90) || rules["botDeclineHours"] != float64(5) {
		t.Fatalf("rules %v", rules)
	}
	h.s.now = func() time.Time { return time.Date(2026, 11, 21, 1, 0, 0, 0, time.UTC) }
	if day, next := dayBounds(h.s.now(), h.s.socialClock(context.Background())); day != "2026-11-20" || !next.Equal(time.Date(2026, 11, 21, 1, 30, 0, 0, time.UTC)) {
		t.Fatalf("social day turns at 01:30 UTC: %s %s", day, next)
	}
	player, bot, person := h.user("fa"), h.user("fa"), h.user("fa")
	if _, err := h.s.db.Exec(`CREATE TABLE IF NOT EXISTS economy_bot (user_id UUID PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO economy_bot (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, bot); err != nil {
		t.Fatal(err)
	}
	h.t.Cleanup(func() { h.s.db.Exec(`DELETE FROM economy_bot WHERE user_id=$1`, bot) })
	h.invite(player, bot, 4*time.Hour)
	h.invite(player, person, 6*time.Hour)
	edges := func(a, b string) int {
		var n int
		if err := h.s.db.QueryRow(`SELECT count(*) FROM user_edge WHERE source_id=$1 AND destination_id=$2`, a, b).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	h.s.db.Exec(`DELETE FROM cluster_lease WHERE name=$1`, friendsLease)
	if _, err := h.s.friendsSweep(context.Background(), "t1"); err != nil {
		t.Fatal(err)
	}
	if edges(player, bot) != 1 {
		t.Fatal("4 h old request to a bot stays")
	}
	h.invite(player, bot, 5*time.Hour+time.Minute)
	h.s.db.Exec(`DELETE FROM cluster_lease WHERE name=$1`, friendsLease)
	if _, err := h.s.friendsSweep(context.Background(), "t1"); err != nil {
		t.Fatal(err)
	}
	if edges(player, bot) != 0 || edges(bot, player) != 0 {
		t.Fatal("a request to a bot is declined after 5 h")
	}
	if edges(player, person) != 1 {
		t.Fatal("requests to people are not touched")
	}
	h.mustAllow(player, bot)
}

func TestCategoryKeysKeepCamelCase(t *testing.T) {
	h := newHarness(t)
	h.s.db.Exec(`DELETE FROM base_category WHERE key IN ('diceSkin','tokenSkin') AND key_type='item_type'`)
	c := h.mustAPI(201, "POST", "/categories", map[string]any{"key": "diceSkin", "keyType": "item_type", "name": map[string]string{"fa": "تاس", "en": "Dice"}})
	if c["key"] != "diceSkin" {
		t.Fatalf("created %v", c)
	}
	up := h.mustAPI(200, "PUT", "/categories/"+c["id"].(string), map[string]any{"key": "diceSkin", "keyType": "item_type", "name": map[string]string{"fa": "پوستهٔ تاس", "en": "Dice skin"}, "sort": 3})
	if up["key"] != "diceSkin" || up["sort"] != float64(3) {
		t.Fatalf("re-saved %v", up)
	}
	for _, bad := range []string{"DiceSkin", "dice skin", ""} {
		if code, out := h.api("POST", "/categories", map[string]any{"key": bad, "keyType": "item_type", "name": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "INVALID_KEY" {
			t.Fatalf("%q = %d %v", bad, code, out)
		}
	}
	h.s.db.Exec(`DELETE FROM base_category WHERE id=$1`, c["id"])
}
