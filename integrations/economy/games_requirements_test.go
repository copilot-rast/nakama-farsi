package main

import (
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

func missTypes(m []reqMiss) []string {
	out := []string{}
	for _, x := range m {
		out = append(out, x.Type)
	}
	return out
}

func TestReqMissingModes(t *testing.T) {
	f := reqFacts{Level: 4, Items: map[string]bool{"i1": true}, Tags: map[string]bool{"vip": true, "cheat": false}, Metas: map[string]float64{"tk": 2, "xp": 49.99995}}
	cases := []struct {
		name string
		r    *gamecatalog.Requirements
		want []string
	}{
		{"nil", nil, nil},
		{"any item owned", &gamecatalog.Requirements{ItemIDs: []string{"i1", "i2"}}, nil},
		{"all items lacking one", &gamecatalog.Requirements{ItemIDs: []string{"i1", "i2"}, ItemsMode: "all"}, []string{"item"}},
		{"any item none owned", &gamecatalog.Requirements{ItemIDs: []string{"i3"}}, []string{"item"}},
		{"tag any", &gamecatalog.Requirements{TagIDs: []string{"vip", "gold"}}, nil},
		{"tag all", &gamecatalog.Requirements{TagIDs: []string{"vip", "gold"}, TagsMode: "all"}, []string{"tag"}},
		{"excluded", &gamecatalog.Requirements{ExcludeTagIDs: []string{"vip"}}, []string{"excludedTag"}},
		{"excluded other", &gamecatalog.Requirements{ExcludeTagIDs: []string{"cheat"}}, nil},
		{"level", &gamecatalog.Requirements{MinLevel: 5}, []string{"level"}},
		{"level ok", &gamecatalog.Requirements{MinLevel: 4}, nil},
		{"meta rounded", &gamecatalog.Requirements{Metas: []gamecatalog.MetaCondition{{MetaID: "xp", Operator: "gte", Value: 50}}}, nil},
		{"meta lt", &gamecatalog.Requirements{Metas: []gamecatalog.MetaCondition{{MetaID: "xp", Operator: "lt", Value: 10}}}, []string{"meta"}},
		{"ticket enough", &gamecatalog.Requirements{Metas: []gamecatalog.MetaCondition{{MetaID: "tk", Operator: "gte", Value: 1, Consume: 2}}}, nil},
		{"ticket short", &gamecatalog.Requirements{Metas: []gamecatalog.MetaCondition{{MetaID: "tk", Operator: "gte", Value: 1, Consume: 3}}}, []string{"meta"}},
		{"several", &gamecatalog.Requirements{MinLevel: 9, ItemIDs: []string{"i9"}, TagIDs: []string{"gold"}, ExcludeTagIDs: []string{"vip"}}, []string{"level", "item", "tag", "excludedTag"}},
	}
	for _, c := range cases {
		got := missTypes(reqMissing(c.r, f))
		if len(got) != len(c.want) {
			t.Fatalf("%s: got %v want %v", c.name, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("%s: got %v want %v", c.name, got, c.want)
			}
		}
	}
	m := reqMissing(&gamecatalog.Requirements{ItemIDs: []string{"i1", "i2", "i3"}, ItemsMode: "all"}, f)
	if len(m) != 1 || len(m[0].IDs) != 2 || m[0].IDs[0] != "i2" || m[0].Mode != "all" {
		t.Fatalf("all mode lists only the lacking items: %+v", m)
	}
	m = reqMissing(&gamecatalog.Requirements{Metas: []gamecatalog.MetaCondition{{MetaID: "tk", Operator: "gte", Value: 5}}}, f)
	if m[0].Have != 2 || m[0].Value != 5 || m[0].Operator != "gte" {
		t.Fatalf("meta miss %+v", m[0])
	}
}

func TestReqSetsInheritAndPaths(t *testing.T) {
	game := &gamecatalog.Requirements{MinLevel: 2, Paths: []string{gamecatalog.PathMatchmaking, gamecatalog.PathBot}}
	off := false
	roomOwn := &gamecatalog.Room{Key: "r", Requirements: &gamecatalog.Requirements{TagIDs: []string{"vip"}, Paths: []string{gamecatalog.PathFriend}}}
	if n := len(reqSets(game, roomOwn, "")); n != 1 {
		t.Fatalf("matchmaking gets the game set only: %d", n)
	}
	if n := len(reqSets(game, roomOwn, gamecatalog.PathFriend)); n != 1 || reqSets(game, roomOwn, gamecatalog.PathFriend)[0] != roomOwn.Requirements {
		t.Fatalf("friend gets the room set only")
	}
	roomOwn.Requirements.Inherit = &off
	if n := len(reqSets(game, roomOwn, gamecatalog.PathBot)); n != 0 {
		t.Fatalf("a room that does not inherit drops the game set: %d", n)
	}
	if n := len(reqSets(game, &gamecatalog.Room{}, gamecatalog.PathBot)); n != 1 {
		t.Fatalf("a plain room inherits: %d", n)
	}
	if n := len(reqSets(nil, nil, "")); n != 0 {
		t.Fatal("nothing to check")
	}
	if !reqHidden([]*gamecatalog.Requirements{{Lock: "show"}, {Lock: "hide"}}) || reqHidden([]*gamecatalog.Requirements{{}}) {
		t.Fatal("hide wins")
	}
}

func TestGameTicketsPublicAndPaths(t *testing.T) {
	tk := &gameTickets{Items: []gameTicketItem{{Key: "pass"}}, Metas: []gameTicketMeta{{Key: "ticket", Amount: 1}, {Key: "ticket", Amount: 2}}, Refund: "return"}
	p := tk.public()
	if p["items"].([]string)[0] != "pass" || p["metas"].(map[string]float64)["ticket"] != 3 {
		t.Fatalf("public %v", p)
	}
	var none *gameTickets
	if !none.empty() || !(&gameTickets{}).empty() || tk.empty() {
		t.Fatal("empty")
	}
	if p, err := cleanPath(""); err != nil || p != "matchmaking" {
		t.Fatal(p, err)
	}
	if _, err := cleanPath("arena"); err == nil {
		t.Fatal("bad path")
	}
	if !metaCondHolds(gamecatalog.MetaCondition{Operator: "eq", Value: 1}, 1.00001) || metaCondHolds(gamecatalog.MetaCondition{Operator: "gte", Value: 0, Consume: 1}, 0.5) {
		t.Fatal("meta holds")
	}
	if q := gameReqUsed("metas"); q == "" || gameReqUsed("itemIds") == q {
		t.Fatal("used sql")
	}
}
