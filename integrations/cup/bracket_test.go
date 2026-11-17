package main

import (
	"reflect"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
)

func TestSeedOrderPutsTopSeedsApart(t *testing.T) {
	if got := seedOrder(8); !reflect.DeepEqual(got, []int{1, 8, 4, 5, 2, 7, 3, 6}) {
		t.Fatalf("order of 8: %v", got)
	}
	if got := seedOrder(2); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("order of 2: %v", got)
	}
	for _, size := range []int{4, 16, 32, 64, 128} {
		got := seedOrder(size)
		if len(got) != size || got[0] != 1 || got[size/2] != 2 {
			t.Fatalf("order of %d: %v", size, got)
		}
	}
}

func TestByesGoToTopSeeds(t *testing.T) {
	for _, tc := range []struct{ players, size, byes int }{{5, 8, 3}, {6, 8, 2}, {7, 8, 1}, {8, 8, 0}, {9, 16, 7}, {12, 16, 4}, {3, 4, 1}, {2, 2, 0}, {100, 128, 28}} {
		if got := bracketSize(tc.players); got != tc.size {
			t.Fatalf("%d players: size %d", tc.players, got)
		}
		pairs := firstRound(tc.size, tc.players)
		byes, seen := 0, map[int]bool{}
		for _, p := range pairs {
			if p.SeedA == 0 {
				t.Fatalf("%d players: empty match %v", tc.players, pairs)
			}
			seen[p.SeedA], seen[p.SeedB] = true, true
			if p.SeedB == 0 {
				byes++
				if p.SeedA > tc.byes {
					t.Fatalf("%d players: seed %d has a bye but a better seed plays", tc.players, p.SeedA)
				}
			}
		}
		if byes != tc.byes {
			t.Fatalf("%d players: %d byes, want %d", tc.players, byes, tc.byes)
		}
		for s := 1; s <= tc.players; s++ {
			if !seen[s] {
				t.Fatalf("%d players: seed %d missing", tc.players, s)
			}
		}
	}
}

func TestSeedingByTrophiesAndRandom(t *testing.T) {
	list := []entrant{{UserID: "a", Trophies: 10, Joined: 1}, {UserID: "b", Trophies: 30, Joined: 2}, {UserID: "c", Trophies: 30, Joined: 0}, {UserID: "d", Trophies: 5, Joined: 3}}
	got := seedEntrants(list, settings.SeedTrophies, nil)
	ids := []string{}
	for _, e := range got {
		ids = append(ids, e.UserID)
	}
	if !reflect.DeepEqual(ids, []string{"c", "b", "a", "d"}) {
		t.Fatalf("trophy seeding (ties by join time): %v", ids)
	}
	rev := seedEntrants(list, settings.SeedRandom, func(n int) int { return 0 })
	if len(rev) != 4 || rev[0].UserID == "c" && rev[3].UserID == "d" {
		t.Fatalf("random seeding did not shuffle: %v", rev)
	}
}

func TestRoundNamesPlacesAndWins(t *testing.T) {
	if roundKey(4, 4) != "final" || roundKey(3, 4) != "semi" || roundKey(2, 4) != "quarter" || roundKey(1, 4) != "r16" || roundKey(1, 6) != "r64" {
		t.Fatal("round keys")
	}
	if placeOf(4, 4) != 2 || placeOf(3, 4) != 3 || placeOf(2, 4) != 5 || placeOf(1, 4) != 9 {
		t.Fatal("places")
	}
	if winsNeeded(1) != 1 || winsNeeded(3) != 2 || winsNeeded(5) != 3 {
		t.Fatal("wins needed")
	}
	if r, s, side := nextSlot(1, 5); r != 2 || s != 2 || side != "b" {
		t.Fatal("next slot")
	}
	if poolShare(1000, 25) != 250 || poolShare(0, 50) != 0 {
		t.Fatal("pool share")
	}
}

func TestRulesValidateAndBestOfOverrides(t *testing.T) {
	r := settings.DefaultRules()
	if f, c := r.Validate(256); f != "" {
		t.Fatalf("defaults invalid: %s %s", f, c)
	}
	r.BestOf = 2
	if f, _ := r.Validate(256); f != "bestOf" {
		t.Fatalf("even best-of accepted: %s", f)
	}
	r.BestOf = 1
	r.RoundBestOf = []settings.RoundBestOf{{Round: "final", BestOf: 3}, {Round: "1", BestOf: 5}, {Round: "third", BestOf: 3}}
	if f, c := r.Validate(256); f != "" {
		t.Fatalf("overrides invalid: %s %s", f, c)
	}
	if r.BestOfFor(4, 4, false) != 3 || r.BestOfFor(1, 4, false) != 5 || r.BestOfFor(2, 4, false) != 1 || r.BestOfFor(4, 4, true) != 3 {
		t.Fatal("best-of overrides")
	}
	r.Capacity = 300
	if f, _ := r.Validate(256); f != "capacity" {
		t.Fatal("capacity over the limit accepted")
	}
	r.Capacity = 16
	r.Prizes.Pool.Shares = settings.Shares{Champion: 60, RunnerUp: 30, Semi: 10}
	if _, c := r.Validate(256); c != "CUP_SHARES_OVER_100" {
		t.Fatalf("shares over 100 accepted: %s", c)
	}
	all := settings.Defaults()
	if f, c := all.Validate(); f != "" {
		t.Fatalf("settings defaults invalid: %s %s", f, c)
	}
}
