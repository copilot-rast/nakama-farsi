package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestLotteryDraftRules(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	userCap, groupCap := 2, 3
	in := lotteryDraftInput{Title: "Raffle", StartsAt: now.Add(time.Hour).Format(time.RFC3339), DrawAt: now.Add(2 * time.Hour).Format(time.RFC3339), TicketPricePoints: 5, PerUserTicketCap: &userCap, PerGroupTicketCap: &groupCap, Prizes: []lotteryPrizeInput{{Title: "Prize", Type: "cash", ValueRial: "0", Priority: 1}}}
	if _, _, err := validateLotteryDraft(in, now, 1, 10); err != nil {
		t.Fatal(err)
	}
	in.DrawAt = now.Format(time.RFC3339)
	if _, _, err := validateLotteryDraft(in, now, 1, 10); err == nil {
		t.Fatal("past draw accepted")
	}
	in.DrawAt = now.Add(2 * time.Hour).Format(time.RFC3339)
	in.Prizes = append(in.Prizes, in.Prizes[0])
	if _, _, err := validateLotteryDraft(in, now, 1, 10); err == nil {
		t.Fatal("duplicate priority accepted")
	}
	in.Prizes = in.Prizes[:1]
	userCap = 4
	if _, _, err := validateLotteryDraft(in, now, 1, 10); err == nil {
		t.Fatal("user cap above group cap accepted")
	}
}

func TestLotteryUpdateOptionalFields(t *testing.T) {
	n := 2
	cap := &n
	if err := decodeLotteryCap(json.RawMessage("null"), &cap); err != nil || cap != nil {
		t.Fatalf("cap clear: %v, %#v", err, cap)
	}
	if err := decodeLotteryCap(json.RawMessage("0"), &cap); err != nil || cap == nil || *cap != 0 {
		t.Fatalf("cap parse: %v, %#v", err, cap)
	}
	if err := decodeLotteryCap(json.RawMessage(`"bogus"`), &cap); err == nil {
		t.Fatal("invalid cap accepted")
	}
	if got, err := (smallintArray{1, 3, 5}).Value(); err != nil || got != "{1,3,5}" {
		t.Fatalf("priority array: %v, %v", got, err)
	}
}

func TestTicketSerialAndSource(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	a, err := newTicketSerial(now)
	if err != nil {
		t.Fatal(err)
	}
	b, err := newTicketSerial(now)
	if err != nil {
		t.Fatal(err)
	}
	if a == b || len(a) > 24 || !strings.HasPrefix(a, "T-") {
		t.Fatalf("invalid serials %q %q", a, b)
	}
	for _, source := range []string{"purchase", "solo_mission", "group_mission", "game", "previous_lottery_prize", "medal"} {
		if !validTicketSource(source) {
			t.Fatalf("rejected %s", source)
		}
	}
	if validTicketSource("arbitrary") {
		t.Fatal("accepted arbitrary ticket source")
	}
}
