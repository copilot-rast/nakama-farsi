package main

import (
	"testing"
	"time"
)

func (f *jazebeFixture) notice(user, key, subject string, content map[string]any, at time.Time) {
	f.t.Helper()
	tx, err := f.db.BeginTx(f.ctx, nil)
	if err != nil {
		f.t.Fatal(err)
	}
	if err := persistJazebeNotification(f.ctx, tx, user, key, subject, content, at); err != nil {
		tx.Rollback()
		f.t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		f.t.Fatal(err)
	}
}

// TestNotificationCentrePostgres covers the main app's notification centre
// and /settings (source notification-centre/-settings services).
func TestNotificationCentrePostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan", "other")
	fan := f.users["fan"]
	other := f.planetOther()
	base := time.Now().UTC().Add(-time.Hour)
	f.notice(fan, "g1", "عضویت تأیید شد", map[string]any{"category": "group", "planetId": f.planetID, "body": "خوش آمدی", "actionLink": "/groups"}, base)
	f.notice(fan, "m1", "مدال گرفتی", map[string]any{"category": "mission_game", "subjectType": "medal", "planetId": f.planetID}, base.Add(time.Minute))
	f.notice(fan, "s1", "خبر سیاره دیگر", map[string]any{"category": "system", "planetId": other}, base.Add(2*time.Minute))
	f.notice(fan, "p1", "خبر پلتفرم", map[string]any{"category": "system"}, base.Add(3*time.Minute))
	f.notice(fan, "c1", "هشدار امنیتی", map[string]any{"category": "critical", "priority": "urgent"}, base.Add(4*time.Minute))
	f.notice(f.users["other"], "x1", "مال دیگری", map[string]any{"category": "system"}, base)

	list := f.ok(fan, "jazebe_notification_list", map[string]any{})
	items := list["items"].([]any)
	if len(items) != 5 || list["unread"].(float64) != 5 {
		t.Fatalf("central inbox %d items unread %v", len(items), list["unread"])
	}
	cats := ""
	for _, it := range items {
		cats += it.(map[string]any)["cat"].(string) + ","
	}
	if cats != "critical,system,system,reward,group," {
		t.Errorf("newest-first categories %s", cats)
	}
	expectJSON(t, "critical", items[0].(map[string]any), map[string]any{"priority": "urgent", "read": false, "action": false})
	expectJSON(t, "group notice", items[4].(map[string]any), map[string]any{"body": "خوش آمدی", "link": "/groups", "action": true, "priority": "normal"})
	if label := items[4].(map[string]any)["planet"]; label == nil || label == "" {
		t.Error("central mode lost the planet label")
	}
	if items[1].(map[string]any)["planet"] != nil {
		t.Error("platform notice got a planet label")
	}

	domain := f.ok(fan, "jazebe_notification_list", map[string]any{"planetId": f.planetID})
	if n := len(domain["items"].([]any)); n != 4 {
		t.Errorf("brand-domain inbox shows %d, want 4 (this planet + platform)", n)
	}
	for _, it := range domain["items"].([]any) {
		if it.(map[string]any)["planet"] != nil {
			t.Error("brand-domain mode must not label planets")
		}
	}
	page1 := f.ok(fan, "jazebe_notification_list", map[string]any{"limit": 2})
	page2 := f.ok(fan, "jazebe_notification_list", map[string]any{"limit": 2, "cursor": page1["nextCursor"]})
	page3 := f.ok(fan, "jazebe_notification_list", map[string]any{"limit": 2, "cursor": page2["nextCursor"]})
	seen := map[string]bool{}
	for _, p := range []map[string]any{page1, page2, page3} {
		for _, it := range p["items"].([]any) {
			id := it.(map[string]any)["id"].(string)
			if seen[id] {
				t.Errorf("cursor paging repeated %s", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != 5 || page3["nextCursor"] != nil {
		t.Errorf("paging covered %d items, last cursor %v", len(seen), page3["nextCursor"])
	}
	f.fails(fan, "jazebe_notification_list", map[string]any{"cursor": "not-a-cursor"}, 3)
	f.fails(fan, "jazebe_notification_list", map[string]any{"category": "spam"}, 3)
	if n := len(f.ok(fan, "jazebe_notification_list", map[string]any{"category": "missions"})["items"].([]any)); n != 1 {
		t.Errorf("missions filter %d", n)
	}

	firstID := items[0].(map[string]any)["id"].(string)
	foreign := f.ok(f.users["other"], "jazebe_notification_list", map[string]any{})["items"].([]any)[0].(map[string]any)["id"].(string)
	f.fails(fan, "jazebe_notification_mark_read", map[string]any{"id": foreign}, 5)
	f.fails(fan, "jazebe_notification_remove", map[string]any{"id": foreign}, 5)
	f.ok(fan, "jazebe_notification_mark_read", map[string]any{"id": firstID})
	f.ok(fan, "jazebe_notification_mark_read", map[string]any{"id": firstID})
	expectJSON(t, "unread", f.ok(fan, "jazebe_notification_unread_count", map[string]any{}), map[string]any{"unread": 4})
	expectJSON(t, "domain unread", f.ok(fan, "jazebe_notification_unread_count", map[string]any{"planetId": f.planetID}), map[string]any{"unread": 3})
	expectJSON(t, "read all", f.ok(fan, "jazebe_notification_read_all", map[string]any{}), map[string]any{"updated": 4})
	expectJSON(t, "after read all", f.ok(fan, "jazebe_notification_unread_count", map[string]any{}), map[string]any{"unread": 0})
	expectJSON(t, "other untouched", f.ok(f.users["other"], "jazebe_notification_unread_count", map[string]any{}), map[string]any{"unread": 1})
	f.ok(fan, "jazebe_notification_remove", map[string]any{"id": firstID})
	f.fails(fan, "jazebe_notification_mark_read", map[string]any{"id": firstID}, 5)
	if n := f.count(`SELECT count(*) FROM notification WHERE id=$1`, firstID); n != 1 {
		t.Error("remove must soft-delete, not drop the Nakama notification")
	}
	expectJSON(t, "clear", f.ok(fan, "jazebe_notification_clear", map[string]any{}), map[string]any{"cleared": 4})
	if n := len(f.ok(fan, "jazebe_notification_list", map[string]any{})["items"].([]any)); n != 0 {
		t.Errorf("cleared inbox still lists %d", n)
	}

	// Preferences and their effect on delivery.
	prefs := f.ok(fan, "jazebe_settings_preferences_get", map[string]any{})
	expectJSON(t, "default categories", prefs["categories"].(map[string]any), map[string]any{"missions": true, "friends": true, "marketing": false})
	updated := f.ok(fan, "jazebe_settings_preferences_update", map[string]any{"categories": map[string]any{"friends": false},
		"channels": map[string]any{"system": map[string]any{"app": false}}})
	expectJSON(t, "updated categories", updated["categories"].(map[string]any), map[string]any{"friends": false, "missions": true})
	expectJSON(t, "updated channel", updated["channels"].(map[string]any)["system"].(map[string]any), map[string]any{"app": false, "sms": false})
	f.fails(fan, "jazebe_settings_preferences_update", map[string]any{"categories": map[string]any{"critical": false}}, 3)
	f.notice(fan, "g2", "گروه", map[string]any{"category": "group", "planetId": f.planetID}, time.Now().UTC())
	f.notice(fan, "s2", "خبر", map[string]any{"category": "system"}, time.Now().UTC())
	f.notice(fan, "m2", "ماموریت", map[string]any{"category": "mission_game"}, time.Now().UTC())
	f.notice(fan, "c2", "حیاتی", map[string]any{"category": "critical"}, time.Now().UTC())
	delivered := f.ok(fan, "jazebe_notification_list", map[string]any{})["items"].([]any)
	if len(delivered) != 2 {
		t.Fatalf("muted categories delivered: %d items", len(delivered))
	}
	if delivered[0].(map[string]any)["cat"] != "critical" || delivered[1].(map[string]any)["cat"] != "game" {
		t.Errorf("delivered %v", delivered)
	}

	// Planet order.
	f.ok(fan, "jazebe_planet_join", map[string]any{"planetId": other})
	order := f.ok(fan, "jazebe_settings_planet_order_get", map[string]any{})
	if order["auto"] != true || len(order["planets"].([]any)) != 2 {
		t.Fatalf("planet order %v", order)
	}
	manual := f.ok(fan, "jazebe_settings_planet_order_update", map[string]any{"planets": []map[string]any{{"id": f.planetID}, {"id": other}}})
	first := manual["planets"].([]any)[0].(map[string]any)
	if manual["auto"] != false || first["id"] != f.planetID {
		t.Errorf("manual order %v", manual)
	}
	byName := f.ok(fan, "jazebe_settings_planet_order_update", map[string]any{"planets": []map[string]any{{"fa": "سیاره دوم"}}})
	if byName["planets"].([]any)[0].(map[string]any)["id"] != other {
		t.Errorf("order by name %v", byName)
	}
	f.fails(fan, "jazebe_settings_planet_order_update", map[string]any{"planets": []map[string]any{{"id": other}, {"id": other}}}, 3)
	f.fails(fan, "jazebe_settings_planet_order_update", map[string]any{"planets": []map[string]any{{"id": f.owner}}}, 3)
	reset := f.ok(fan, "jazebe_settings_planet_order_update", map[string]any{"auto": true})
	if reset["auto"] != true {
		t.Errorf("reset %v", reset)
	}
}

// planetOther creates and activates a second planet in the fixture brand.
func (f *jazebeFixture) planetOther() string {
	f.t.Helper()
	id := f.ok(f.owner, "jazebe_planet_create", map[string]any{"brandId": f.brandID, "name": "سیاره دوم", "slug": "second-" + f.owner[:8],
		"latinName": "Second " + f.owner[:8], "pointName": "سکه"})["id"].(string)
	f.exec(`UPDATE jazebe_planet SET status='active',activated_at=now() WHERE id=$1`, id)
	return id
}

// TestPlanetHomePostgres covers GET /planet, the constellation, the three
// Home-hub previews and the Sky (source planet-overview.service.ts).
func TestPlanetHomePostgres(t *testing.T) {
	f := newJazebeFixture(t, "fan", "newcomer")
	fan := f.users["fan"]
	other := f.planetOther()
	var food string
	if err := f.db.QueryRowContext(f.ctx, `SELECT id FROM jazebe_planet_category WHERE key='food'`).Scan(&food); err != nil {
		t.Fatal(err)
	}
	f.ok(f.owner, "jazebe_brand_planet_update", map[string]any{"planetId": f.planetID, "shortDescription": "سیاره خوراکی‌ها",
		"categoryId": food, "brandColor": "#AA3300", "glyph": "🍕"})
	now := time.Now().UTC()
	lottery := f.ok(f.owner, "jazebe_lottery_draft_create", map[string]any{"planetId": f.planetID, "title": "قرعه بزرگ",
		"startsAt": now.Add(-time.Hour).Format(time.RFC3339), "drawAt": now.Add(26 * time.Hour).Format(time.RFC3339), "ticketPricePoints": 3,
		"prizes": []map[string]any{{"title": "موبایل", "type": "cash", "valueRial": "5000000", "priority": 1}, {"title": "هدفون", "type": "cash", "valueRial": "1000000", "priority": 2}}})["id"].(string)
	f.ok(f.owner, "jazebe_lottery_activate", map[string]any{"planetId": f.planetID, "lotteryId": lottery})
	start, end := missionWindow()
	mission := f.ok(f.owner, "jazebe_code_mission_draft_create", map[string]any{"planetId": f.planetID, "title": "ماموریت روز", "shortDescription": "کد را وارد کن",
		"detectionMethod": "platform_auto", "rewardPoints": 40, "startsAt": start, "endsAt": end, "code": map[string]any{"codeMode": "unique", "maxUses": 1}})["id"].(string)
	f.ok(f.owner, "jazebe_code_mission_activate", map[string]any{"planetId": f.planetID, "missionId": mission})
	group := f.ok(fan, "jazebe_group_create", map[string]any{"planetId": f.planetID, "name": "گروه خانه", "mode": "open", "cap": 4})["id"].(string)
	f.notice(fan, "home-1", "تازه", map[string]any{"category": "system", "planetId": f.planetID}, now)
	f.notice(fan, "home-2", "دیگر", map[string]any{"category": "system", "planetId": other}, now)

	overview := f.ok(fan, "jazebe_planet_overview", map[string]any{"planetId": f.planetID})
	brand := overview["brand"].(map[string]any)
	expectJSON(t, "brand", brand, map[string]any{"id": f.planetID, "fa": "سیاره " + f.owner[:8], "name": "Fixture Planet", "category": "خوراکی",
		"pointName": "امتیاز", "color": "#AA3300", "glyph": "🍕", "shortDescription": "سیاره خوراکی‌ها", "fans": 2})
	expectJSON(t, "stats", overview["stats"].(map[string]any), map[string]any{"points": 0, "level": 1, "levelMax": 3, "levelName": "تازه‌وارد"})
	expectJSON(t, "teaser", overview["raffleTeaser"].(map[string]any), map[string]any{"id": lottery, "title": "قرعه بزرگ", "prize": "موبایل"})
	sections := overview["sections"].(map[string]any)
	expectJSON(t, "missions section", sections["missions"].(map[string]any), map[string]any{"sub": "۱ ماموریت تازه", "hot": true})
	expectJSON(t, "groups section", sections["groups"].(map[string]any), map[string]any{"sub": "۱ گروه فعال"})
	expectJSON(t, "games section", sections["games"].(map[string]any), map[string]any{"sub": "هر روز امتیاز بگیر"})
	if overview["unreadNotifications"].(float64) != 1 {
		t.Errorf("planet-scoped unread %v, want 1 (other planet's notice excluded)", overview["unreadNotifications"])
	}
	expectJSON(t, "viewer", overview["viewer"].(map[string]any), map[string]any{"member": true, "joinPath": "join_button", "hasActivity": false})
	latest := f.ok(fan, "jazebe_planet_overview", map[string]any{})
	if latest["brand"].(map[string]any)["id"] != f.planetID {
		t.Errorf("latest-membership fallback chose %v", latest["brand"])
	}
	stranger := f.user("stranger")
	expectJSON(t, "newcomer", f.ok(stranger, "jazebe_planet_overview", map[string]any{"planetId": f.planetID})["viewer"].(map[string]any),
		map[string]any{"member": false, "joinPath": nil})
	f.fails(stranger, "jazebe_planet_overview", map[string]any{}, 5)

	f.ok(fan, "jazebe_planet_join", map[string]any{"planetId": other})
	constellation := f.ok(fan, "jazebe_planet_constellation", map[string]any{})["items"].([]any)
	if len(constellation) != 2 {
		t.Fatalf("constellation %d", len(constellation))
	}
	// Recent activity first; a plain join records no activity (NULLS LAST).
	expectJSON(t, "constellation card", constellation[0].(map[string]any), map[string]any{"id": f.planetID, "category": "خوراکی", "glyph": "🍕"})
	expectJSON(t, "second card", constellation[1].(map[string]any), map[string]any{"id": other, "glyph": "S"})

	missions := f.ok(fan, "jazebe_planet_missions_preview", map[string]any{"planet": "fx-" + f.owner[:8]})["items"].([]any)
	if len(missions) != 1 {
		t.Fatalf("missions preview %d", len(missions))
	}
	expectJSON(t, "mission preview", missions[0].(map[string]any), map[string]any{"id": mission, "title": "ماموریت روز", "desc": "کد را وارد کن", "points": 40, "progress": 0, "total": 1, "period": "۳ روز"})
	groups := f.ok(fan, "jazebe_planet_groups_preview", map[string]any{"planet": "fixture planet"})["items"].([]any)
	expectJSON(t, "group preview", groups[0].(map[string]any), map[string]any{"id": group, "meta": "کاپیتان", "cta": "باز کردن", "joined": true, "tokens": "۰"})
	raffle := f.ok(fan, "jazebe_planet_raffle_preview", map[string]any{"planet": f.planetID})["raffle"].(map[string]any)
	expectJSON(t, "raffle headline", raffle["headline"].(map[string]any), map[string]any{"label": "جایزه بزرگ", "prize": "موبایل"})
	if raffle["ticketCost"].(float64) != 3 {
		t.Errorf("ticket cost %v", raffle["ticketCost"])
	}
	if r := f.ok(fan, "jazebe_planet_raffle_preview", map[string]any{"planet": other}); r["raffle"] != nil {
		t.Errorf("planet without lottery returned %v", r)
	}
	f.fails(fan, "jazebe_planet_missions_preview", map[string]any{"planet": "no-such-planet"}, 5)

	sky := f.ok(f.users["newcomer"], "jazebe_planet_discover", map[string]any{"category": "food"})["items"].([]any)
	if len(sky) != 1 {
		t.Fatalf("category filter %d", len(sky))
	}
	expectJSON(t, "sky card", sky[0].(map[string]any), map[string]any{"id": f.planetID, "totalActivePrizeValue": 6000000, "activeGroupCount": 1, "isMember": true})
	expectJSON(t, "sky category", sky[0].(map[string]any)["category"].(map[string]any), map[string]any{"key": "food", "titleFa": "خوراکی"})
	searched := f.ok(f.users["newcomer"], "jazebe_planet_discover", map[string]any{"search": "Second"})["items"].([]any)
	if len(searched) != 1 || searched[0].(map[string]any)["id"] != other {
		t.Errorf("search %v", searched)
	}

	f.ok(f.operator, "jazebe_planet_suspend", map[string]any{"planetId": f.planetID, "reason": "بررسی"})
	f.fails(fan, "jazebe_planet_overview", map[string]any{"planetId": f.planetID}, 7)
	f.fails(fan, "jazebe_planet_raffle_preview", map[string]any{"planet": f.planetID}, 7)
	for _, it := range f.ok(fan, "jazebe_planet_constellation", map[string]any{})["items"].([]any) {
		if it.(map[string]any)["id"] == f.planetID {
			t.Error("suspended planet still in the constellation")
		}
	}
}
