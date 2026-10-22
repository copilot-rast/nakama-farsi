package main

import (
	"encoding/json"
	"testing"
)

func TestGroupOptionalString(t *testing.T) {
	for _, tc := range []struct {
		name     string
		raw      json.RawMessage
		limit    int
		nullable bool
		want     string
		present  bool
		bad      bool
	}{
		{"absent", nil, 64, false, "", false, false},
		{"trim", json.RawMessage(`"  نمونه  "`), 64, false, "نمونه", true, false},
		{"clear", json.RawMessage(`null`), 280, true, "", true, false},
		{"null name", json.RawMessage(`null`), 64, false, "", false, true},
		{"empty name", json.RawMessage(`" "`), 64, false, "", false, true},
		{"too long", json.RawMessage(`"12345"`), 4, true, "", false, true},
		{"wrong type", json.RawMessage(`42`), 64, true, "", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, present, err := groupOptionalString(tc.raw, tc.limit, tc.nullable)
			if (err != nil) != tc.bad || (!tc.bad && (got != tc.want || present != tc.present)) {
				t.Fatalf("got (%q,%v,%v), want (%q,%v,bad=%v)", got, present, err, tc.want, tc.present, tc.bad)
			}
		})
	}
}

func TestGroupRemovePermissions(t *testing.T) {
	for _, tc := range []struct {
		actor, target string
		self, allowed bool
	}{
		{"creator", "admin", false, true},
		{"creator", "member", false, true},
		{"admin", "member", false, true},
		{"admin", "admin", false, false},
		{"member", "member", false, false},
		{"creator", "creator", false, false},
		{"creator", "member", true, false},
	} {
		if got := groupCanRemove(tc.actor, tc.target, tc.self); (got == nil) != tc.allowed {
			t.Errorf("actor=%s target=%s self=%v: %v", tc.actor, tc.target, tc.self, got)
		}
	}
}

func TestGroupReadProjectionRules(t *testing.T) {
	if memberRolePriority("creator") >= memberRolePriority("admin") || memberRolePriority("admin") >= memberRolePriority("member") {
		t.Fatal("member role order changed")
	}
	for kind, want := range map[string]string{"join": "join", "leave": "join", "role": "join", "ticket": "ticket", "game": "game", "prize": "prize", "mission": "game"} {
		if got := groupFeedKind(kind); got != want {
			t.Errorf("%s: %s, want %s", kind, got, want)
		}
	}
}
