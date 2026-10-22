package main

import "testing"

func TestGameRoundOutcome(t *testing.T) {
	boxes := [3]int{1, 3, 5}
	score, tickets, err := gameRoundOutcome("box", boxes, 1, 0, 10, gameMinRoundMs, 1, 5)
	if err != nil || score != 0.5 || tickets != 3 {
		t.Fatalf("box: %v %v %v", score, tickets, err)
	}
	if _, _, err = gameRoundOutcome("box", boxes, 3, 0, 10, gameMinRoundMs, 1, 5); err == nil {
		t.Fatal("invalid box accepted")
	}
	if _, _, err = gameRoundOutcome("tap", boxes, 0, 12, 10, 9000, 0, 10); err == nil {
		t.Fatal("early tap accepted")
	}
	score, tickets, err = gameRoundOutcome("tap", boxes, 0, 12, 10, 10000, 0, 10)
	if err != nil || score != 1 || tickets != 10 {
		t.Fatalf("tap: %v %v %v", score, tickets, err)
	}
	if _, _, err = gameRoundOutcome("tap", boxes, 0, 41, 10, 10000, 0, 10); err == nil {
		t.Fatal("impossible hits accepted")
	}
}

func TestGameRollBoxes(t *testing.T) {
	boxes, err := gameRollBoxes(2, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range boxes {
		if n < 2 || n > 4 {
			t.Fatalf("out of range: %d", n)
		}
	}
	if _, err = gameRollBoxes(5, 4); err == nil {
		t.Fatal("inverted range accepted")
	}
}

func TestGameConfigRequiresRealEntryAndPrizeDestination(t *testing.T) {
	base := playableGame{EntryModel: "free", FreeIndividualCount: 3, PrizeMin: 0, PrizeMax: 0, DurationSec: 10}
	if err := gameConfigValid(base); err != nil {
		t.Fatalf("valid free game rejected: %v", err)
	}
	for name, change := range map[string]func(*playableGame){
		"no quota":              func(g *playableGame) { g.FreeIndividualCount = 0 },
		"paid without cost":     func(g *playableGame) { g.EntryModel = "cost" },
		"prize without lottery": func(g *playableGame) { g.PrizeMax = 1 },
		"short tap":             func(g *playableGame) { g.DurationSec = 2 },
		"negative prize":        func(g *playableGame) { g.PrizeMin = -1 },
	} {
		g := base
		change(&g)
		if err := gameConfigValid(g); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
