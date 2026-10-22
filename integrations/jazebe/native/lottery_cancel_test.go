package main

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLotteryCancellationRules(t *testing.T) {
	in := lotteryCancelInput{PlanetID: "12e30ed4-a7a6-4cf5-a98c-b445812c4d05", LotteryID: "d1693500-813c-4d23-9016-c9459f9778c7", Reason: "  not enough participants  "}
	validated, err := validateLotteryCancel(in)
	if err != nil {
		t.Fatal(err)
	}
	if validated.Reason != "not enough participants" {
		t.Fatalf("untrimmed reason: %q", validated.Reason)
	}
	in.Reason = "no"
	if _, err = validateLotteryCancel(in); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("short reason accepted: %v", err)
	}
}

func TestCancelledTicketBundlesPreserveOwnership(t *testing.T) {
	tickets := []cancelledTicket{
		{GroupID: "g", CreatorType: "member", CreatorUserID: "u", Source: "solo_mission", SourceID: "source-a"},
		{GroupID: "g", CreatorType: "member", CreatorUserID: "u", Source: "solo_mission", SourceID: "source-a"},
		{GroupID: "g", CreatorType: "group", Source: "group_mission", SourceID: "source-b"},
	}
	bundles := bundleCancelledTickets(tickets)
	if len(bundles) != 2 || bundles[0].Count != 2 || bundles[1].Count != 1 || bundles[1].Ticket.CreatorType != "group" {
		t.Fatalf("incorrect transfer bundles: %#v", bundles)
	}
}
