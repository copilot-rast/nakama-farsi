package main

import (
	"math/big"
	"reflect"
	"testing"
)

func TestDrawLotteryDeterministicAndNoReplacement(t *testing.T) {
	tickets := make([]drawTicket, 500)
	for i := range tickets {
		tickets[i] = drawTicket{ID: string(rune(i + 1)), Serial: serialForTest(i), GroupID: "group"}
	}
	prizes := []drawPrize{{"p3", 3}, {"p1", 1}, {"p2", 2}}
	a, err := drawLottery("fixed seed", tickets, prizes)
	if err != nil {
		t.Fatal(err)
	}
	for i, j := 0, len(tickets)-1; i < j; i, j = i+1, j-1 {
		tickets[i], tickets[j] = tickets[j], tickets[i]
	}
	b, err := drawLottery("fixed seed", tickets, []drawPrize{prizes[2], prizes[1], prizes[0]})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("draw depends on input order")
	}
	if len(a.Picks) != 3 || a.Picks[0].Priority != 1 || a.Picks[1].Priority != 2 || a.Picks[2].Priority != 3 {
		t.Fatalf("wrong picks: %+v", a.Picks)
	}
	seen := map[string]bool{}
	for _, p := range a.Picks {
		if seen[p.TicketID] {
			t.Fatal("ticket won twice")
		}
		seen[p.TicketID] = true
	}
	if a.TicketCount != 500 || len(a.TicketDigest) != 64 {
		t.Fatalf("wrong pool metadata: %+v", a)
	}
}

func serialForTest(i int) string {
	return "T-" + leftPad4(string([]byte{'0' + byte(i/1000%10), '0' + byte(i/100%10), '0' + byte(i/10%10), '0' + byte(i%10)}))
}

func TestDrawLotteryUnawarded(t *testing.T) {
	a, err := drawLottery("seed", []drawTicket{{"t", "T-0001", "g"}}, []drawPrize{{"p3", 3}, {"p1", 1}, {"p2", 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Picks) != 1 || a.Picks[0].PrizeID != "p1" || !reflect.DeepEqual(a.UnawardedPrizeIDs, []string{"p2", "p3"}) {
		t.Fatalf("wrong unawarded prizes: %+v", a)
	}
}

func TestDrawLotteryCrossLanguageFixture(t *testing.T) {
	// These indices come from Node's createHmac('sha256', 'seed') over
	// decimal counters 0, 1 and 2, then its first six bytes.
	pool := []drawTicket{{"t2", "T-2", "g"}, {"t0", "T-0", "g"}, {"t1", "T-1", "g"}}
	out, err := drawLottery("seed", pool, []drawPrize{{"p3", 3}, {"p1", 1}, {"p2", 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Picks) != 3 || out.Picks[0].TicketID != "t0" || out.Picks[1].TicketID != "t2" || out.Picks[2].TicketID != "t1" {
		t.Fatalf("Go draw diverges from source HMAC fixture: %+v", out.Picks)
	}
}

func TestSplitLotteryPrizeSourceRules(t *testing.T) {
	shares, err := splitLotteryPrize(big.NewInt(10_000_000), []shareInput{{"u-b", 2}, {"u-c", 0}, {"u-a", 1}})
	if err != nil {
		t.Fatal(err)
	}
	if shares[0].SharePercent != "33.3333" || shares[0].ShareValueRial != "3333333" || shares[1].SharePercent != "66.6667" || shares[1].ShareValueRial != "6666667" || shares[2].ShareValueRial != "0" {
		t.Fatalf("wrong weighted split: %+v", shares)
	}
	equal, err := splitLotteryPrize(big.NewInt(100), []shareInput{{"u-3", 0}, {"u-2", 0}, {"u-1", 0}})
	if err != nil {
		t.Fatal(err)
	}
	if equal[0].ShareValueRial != "34" || equal[1].ShareValueRial != "33" || equal[2].ShareValueRial != "33" {
		t.Fatalf("wrong equal split: %+v", equal)
	}
}
