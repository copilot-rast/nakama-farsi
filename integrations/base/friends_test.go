package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFriendRulesClean(t *testing.T) {
	ok := []friendRules{defaultFriendRules(), {DailyRequests: 0, MaxFriends: 10000, ExpiryDays: 365, RejectCooldownHours: 8760}}
	for _, r := range ok {
		if _, err := r.clean(); err != nil {
			t.Fatalf("%+v: %v", r, err)
		}
	}
	bad := map[string]friendRules{
		"dailyRequests":       {DailyRequests: 1001},
		"maxFriends":          {MaxFriends: -1},
		"expiryDays":          {ExpiryDays: 366},
		"rejectCooldownHours": {RejectCooldownHours: -2},
	}
	for field, r := range bad {
		_, err := r.clean()
		app, isApp := asApp(err)
		if !isApp || app.Code != "INVALID_RULE" || app.Field != field {
			t.Fatalf("%s: %v", field, err)
		}
	}
	if d := defaultFriendRules(); d.DailyRequests != 20 || d.MaxFriends != 0 || d.ExpiryDays != 0 || d.RejectCooldownHours != 0 {
		t.Fatalf("defaults %+v", d)
	}
}

func TestDayBoundsUseTheProjectZone(t *testing.T) {
	tehran, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Skip("no tzdata")
	}
	now := time.Date(2026, 11, 14, 21, 0, 0, 0, time.UTC)
	day, reset := dayBounds(now, tehran)
	if day != "2026-11-15" || !reset.Equal(time.Date(2026, 11, 16, 0, 0, 0, 0, tehran)) {
		t.Fatalf("tehran %s %s", day, reset)
	}
	day, reset = dayBounds(now, time.UTC)
	if day != "2026-11-14" || !reset.Equal(time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("utc %s %s", day, reset)
	}
}

func TestFriendErrorIsJSONWithCode(t *testing.T) {
	left := 0
	e := &friendError{Code: CodeFriendDailyLimit, Message: "x", Lang: "fa", Limit: 20, Remaining: &left, ResetsAt: "2026-11-15T20:30:00Z"}
	var m map[string]any
	if err := json.Unmarshal([]byte(e.Error()), &m); err != nil {
		t.Fatal(err)
	}
	if m["code"] != CodeFriendDailyLimit || m["remaining"].(float64) != 0 || m["limit"].(float64) != 20 || m["retryAt"] != nil {
		t.Fatalf("%v", m)
	}
	if localDigits("fa", "20 / 3") != "۲۰ / ۳" || localDigits("en", "20") != "20" {
		t.Fatal("digits")
	}
	for code, text := range friendMessages {
		for _, lang := range []string{"fa", "en", "ar"} {
			if text[lang] == "" {
				t.Fatalf("%s has no %s text", code, lang)
			}
		}
	}
}
