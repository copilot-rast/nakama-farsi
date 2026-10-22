package main

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOperatorUserFilter(t *testing.T) {
	in, err := validateOperatorUserFilter(operatorUserInput{Search: "  Leila  ", State: "suspended", Verification: "review"})
	if err != nil || in.Search != "Leila" || in.Page != 1 || in.Limit != 20 {
		t.Fatalf("valid operator user filter: %+v %v", in, err)
	}
	for _, invalid := range []operatorUserInput{
		{Search: strings.Repeat("x", 65)},
		{State: "deleted"},
		{Verification: "pending"},
		{Page: -1},
		{Limit: 101},
	} {
		if _, err = validateOperatorUserFilter(invalid); status.Code(err) != codes.InvalidArgument {
			t.Errorf("invalid filter accepted: %+v %v", invalid, err)
		}
	}
}
