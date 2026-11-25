package stealthbots

import (
	"math/rand/v2"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestIdentityLooksLikeAPlayer(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		u, d := Identity(r)
		if u == "" || utf8.RuneCountInString(u) > 40 || strings.ContainsAny(u, " \t") || strings.Contains(strings.ToLower(u), "bot") {
			t.Fatal(u)
		}
		seen[u] = true
		if d != "" && strings.ContainsAny(d, "abcdefghijklmnopqrstuvwxyz") {
			t.Fatal(d)
		}
		if m := Markers([]byte(`{"username":"` + u + `"}`)); len(m) != 0 {
			t.Fatal(m)
		}
	}
	if len(seen) < 300 {
		t.Fatalf("names repeat too much: %d", len(seen))
	}
}

func TestGrowLevelStaysNearAndInRange(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for i := 0; i < 1000; i++ {
		if lv := GrowLevel(r, 12, 2, 0); lv < 10 || lv > 14 {
			t.Fatal(lv)
		}
		if lv := GrowLevel(r, 1, 3, 4); lv < 4 {
			t.Fatal(lv)
		}
		if lv := GrowLevel(r, 0, 2, 0); lv < 1 || lv > 30 {
			t.Fatal(lv)
		}
	}
	s := SeedStats(r, 10)
	if s.Games < 33 || s.Wins+s.Losses != s.Games || s.Wins == 0 || s.Trophies < 0 {
		t.Fatalf("%+v", s)
	}
}

func TestPickLookFollowsLevel(t *testing.T) {
	items := []LookItem{
		{Key: "legend", Rarity: "legendary", Currency: "gem", Price: 300},
		{Key: "starter", Starter: true},
		{Key: "cheap", Currency: "coin", Price: 100},
		{Key: "mid", Currency: "coin", Price: 600},
		{Key: "rare", Rarity: "rare", Currency: "coin", Price: 900},
		{Key: "epic", Rarity: "epic", Currency: "gem", Price: 120, Unlock: 25},
	}
	ranked := RankLook(items)
	order := []string{}
	for _, it := range ranked {
		order = append(order, it.Key)
	}
	if strings.Join(order, ",") != "starter,cheap,mid,rare,epic,legend" {
		t.Fatal(order)
	}
	r := rand.New(rand.NewPCG(5, 6))
	if got := PickLook(items, 1, 30, 0, r); got != "starter" {
		t.Fatal(got)
	}
	if got := PickLook(items, 30, 30, 0, r); got != "legend" {
		t.Fatal(got)
	}
	if got := PickLook(items, 22, 30, 0, r); got != "rare" {
		t.Fatalf("epic unlocks at 25: %s", got)
	}
	low, high := 0, 0
	for i := 0; i < 400; i++ {
		switch PickLook(items, 3, 30, 15, r) {
		case "starter", "cheap":
			low++
		case "epic", "legend":
			high++
		}
	}
	if low < 300 || high != 0 {
		t.Fatalf("a low level bot stays basic: low=%d high=%d", low, high)
	}
	if PickLook(nil, 5, 30, 10, r) != "" {
		t.Fatal("empty")
	}
}
