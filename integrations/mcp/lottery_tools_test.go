package main

import "testing"

func TestLotteryUpdateClearFields(t *testing.T) {
	name := "new"
	cap := 5
	out := lotteryUpdatePayload(lotteryUpdateToolInput{PlanetID: "p", LotteryID: "l", Title: &name, PerUserTicketCap: &cap, ClearPerUserTicketCap: true, ClearShortDescription: true})
	if out["title"] != "new" || out["perUserTicketCap"] != nil || out["shortDescription"] != nil {
		t.Fatalf("unexpected update payload %#v", out)
	}
	if _, ok := out["drawAt"]; ok {
		t.Fatal("omitted field changed")
	}
}
