package main

import (
	"math/rand"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
)

func TestRolesPromoteDemoteAndRank(t *testing.T) {
	cases := []struct {
		actor, target, want string
		ok                  bool
	}{
		{settings.RoleLeader, settings.RoleMember, settings.RoleElder, true},
		{settings.RoleLeader, settings.RoleElder, settings.RoleCoLeader, true},
		{settings.RoleCoLeader, settings.RoleMember, settings.RoleElder, true},
		{settings.RoleCoLeader, settings.RoleElder, "", false},
		{settings.RoleCoLeader, settings.RoleCoLeader, "", false},
		{settings.RoleElder, settings.RoleMember, "", false},
		{settings.RoleLeader, settings.RoleCoLeader, "", false},
	}
	for _, c := range cases {
		got, ok := canPromote(c.actor, c.target)
		if ok != c.ok || got != c.want {
			t.Fatalf("%s promotes %s: %q %v, want %q %v", c.actor, c.target, got, ok, c.want, c.ok)
		}
	}
	if demoteTo(settings.RoleCoLeader) != settings.RoleElder || demoteTo(settings.RoleElder) != settings.RoleMember || demoteTo(settings.RoleMember) != "" {
		t.Fatal("demote ladder")
	}
	if !outranks(settings.RoleLeader, settings.RoleCoLeader) || outranks(settings.RoleElder, settings.RoleElder) || outranks(settings.RoleMember, settings.RoleLeader) {
		t.Fatal("outranks")
	}
	if stateOf(settings.RoleLeader) != 0 || stateOf(settings.RoleCoLeader) != 1 || stateOf(settings.RoleElder) != 2 || stateOf(settings.RoleMember) != 2 {
		t.Fatal("native group states")
	}
}

func TestRoadLevelsAndCaps(t *testing.T) {
	r := newRoad([]level{{Level: 3, XP: 300, MemberCap: 50}, {Level: 1, XP: 0, MemberCap: 30}, {Level: 2, XP: 100}})
	for xp, want := range map[int64]int{0: 1, 99: 1, 100: 2, 299: 2, 300: 3, 99999: 3} {
		if got := r.levelFor(xp); got != want {
			t.Fatalf("xp %d: level %d want %d", xp, got, want)
		}
	}
	if r.memberCap(2, 10) != 30 || r.memberCap(3, 10) != 50 || newRoad(nil).memberCap(4, 10) != 10 {
		t.Fatal("member cap is the last cap set at or below the level")
	}
	if n := r.next(1); n == nil || n.Level != 2 {
		t.Fatal("next level")
	}
	if r.next(3) != nil {
		t.Fatal("no level after the last")
	}
}

func TestPairingModes(t *testing.T) {
	list := []entrant{{ID: "a", Level: 1, Trophies: 10}, {ID: "b", Level: 5, Trophies: 900}, {ID: "c", Level: 3, Trophies: 50}, {ID: "d", Level: 4, Trophies: 800}, {ID: "e", Level: 2, Trophies: 20}}
	byT := pairClans(list, "trophies", rand.New(rand.NewSource(1)))
	if len(byT) != 3 || byT[0] != (pair{A: "b", B: "d"}) || byT[1] != (pair{A: "c", B: "e"}) || byT[2] != (pair{A: "a"}) {
		t.Fatalf("by trophies %+v", byT)
	}
	byL := pairClans(list, "level", rand.New(rand.NewSource(1)))
	if byL[0] != (pair{A: "b", B: "d"}) || byL[1] != (pair{A: "c", B: "e"}) {
		t.Fatalf("by level %+v", byL)
	}
	seen := map[string]bool{}
	for _, p := range pairClans(list, "random", rand.New(rand.NewSource(7))) {
		seen[p.A], seen[p.B] = true, p.B != "" || seen[p.B]
	}
	if len(seen) < 5 {
		t.Fatalf("random pairing keeps every clan once: %v", seen)
	}
}

func TestWarResultsAndOutcomes(t *testing.T) {
	if warResult(3, 1, false) != "a" || warResult(1, 3, false) != "b" || warResult(2, 2, false) != "draw" || warResult(0, 0, true) != "bye" {
		t.Fatal("results")
	}
	if outcomeFor("a", true, "none") != "win" || outcomeFor("a", false, "none") != "lose" || outcomeFor("draw", false, "none") != "draw" {
		t.Fatal("outcomes")
	}
	if outcomeFor("bye", true, "win") != "win" || outcomeFor("bye", true, "none") != "bye" {
		t.Fatal("bye outcome follows the rule")
	}
	p := Prizes{Win: Prize{PrizeKey: "w"}, Lose: Prize{PrizeKey: "l"}, Draw: Prize{PrizeKey: "d"}}
	if p.keyFor("win") != "w" || p.keyFor("lose") != "l" || p.keyFor("draw") != "d" || p.keyFor("bye") != "" {
		t.Fatal("prize keys")
	}
}

func TestRulesValidateBothModes(t *testing.T) {
	r := defaultRules(modePoints)
	if f, _ := r.validate(modePoints); f != "rules.points" {
		t.Fatalf("points mode needs point rules: %s", f)
	}
	r.Points = []PointRule{{Type: "economy_event", Event: "match.won", Points: 3}}
	if f, c := r.validate(modePoints); c != "" {
		t.Fatalf("valid points rules: %s %s", f, c)
	}
	d := defaultRules(modeDuel)
	if f, _ := d.validate(modeDuel); f != "rules.duel.event" {
		t.Fatalf("duel mode needs an event: %s", f)
	}
	d.Duel.Event = "match.played"
	if f, c := d.validate(modeDuel); c != "" {
		t.Fatalf("valid duel rules: %s %s", f, c)
	}
	d.Prizes.To = "everyone"
	if f, _ := d.validate(modeDuel); f != "rules.prizes.to" {
		t.Fatalf("prizes.to: %s", f)
	}
}

func TestXPAndPointsFromEvents(t *testing.T) {
	src := settings.XPSource{Type: "economy_event", Event: "goal", Amount: 2, PerValue: true}
	if xpFor(src, 3, nil) != 6 {
		t.Fatal("per value")
	}
	src = settings.XPSource{Type: "economy_event", Event: "played", Field: "won", Equals: "true", Amount: 10}
	if xpFor(src, 1, map[string]any{"won": true}) != 10 || xpFor(src, 1, map[string]any{"won": false}) != 0 {
		t.Fatal("field filter")
	}
	if pointsFor(PointRule{Points: 3}, 1, nil) != 3 || pointsFor(PointRule{Points: 1, PerValue: true, Field: "goals"}, 1, map[string]any{"goals": float64(4)}) != 4 {
		t.Fatal("points")
	}
	if clipDaily(100, 90, 30) != 10 || clipDaily(100, 100, 5) != 0 || clipDaily(0, 1e6, 5) != 5 {
		t.Fatal("daily cap")
	}
	d := settings.Defaults()
	if f, c := d.Validate(); c != "" {
		t.Fatalf("defaults %s %s", f, c)
	}
	d.Roles.Kick = []string{"boss"}
	if f, _ := d.Validate(); f != "roles.kick" {
		t.Fatalf("unknown role %s", f)
	}
}
