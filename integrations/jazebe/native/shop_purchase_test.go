package main

import (
	"regexp"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPurchaseIdempotencyKeyAndOrderID(t *testing.T) {
	in := ticketPurchaseInput{PlanetID: "12e30ed4-a7a6-4cf5-a98c-b445812c4d05", ProductID: "d1693500-813c-4d23-9016-c9459f9778c7", IdempotencyKey: " retry-key-123 "}
	valid, err := validateTicketPurchase(in)
	if err != nil {
		t.Fatal(err)
	}
	if valid.IdempotencyKey != "retry-key-123" {
		t.Fatalf("key not trimmed: %q", valid.IdempotencyKey)
	}
	a := purchaseOrderID("buyer", valid.ProductID, valid.IdempotencyKey)
	if a != purchaseOrderID("buyer", valid.ProductID, valid.IdempotencyKey) || a == purchaseOrderID("buyer", valid.ProductID, "another-key") || !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(a) {
		t.Fatalf("invalid deterministic order ID: %s", a)
	}
	in.IdempotencyKey = "short"
	if _, err := validateTicketPurchase(in); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("short key accepted: %v", err)
	}
}

func TestTicketProductValidation(t *testing.T) {
	in := ticketProductInput{Name: "Ticket", PricePoints: 5}
	if err := validateTicketProduct(in); err != nil {
		t.Fatal(err)
	}
	stock := 0
	in.Stock = &stock
	if err := validateTicketProduct(in); err != nil {
		t.Fatal(err)
	}
	stock = -1
	if err := validateTicketProduct(in); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("negative stock accepted: %v", err)
	}
}
