package main

import (
	"context"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOperatorReportListValidation(t *testing.T) {
	valid, err := validateOperatorReportList(operatorReportInput{SubjectType: "profile"})
	if err != nil || valid.Status != "open" || valid.Page != 1 || valid.Limit != 20 {
		t.Fatalf("default report queue: %+v %v", valid, err)
	}
	for _, input := range []operatorReportInput{
		{Status: "reviewing"},
		{SubjectType: "user"},
		{PlanetID: "bad-id"},
		{Page: -1},
		{Limit: 101},
	} {
		if _, err := validateOperatorReportList(input); status.Code(err) != codes.InvalidArgument {
			t.Errorf("invalid report filter accepted: %+v %v", input, err)
		}
	}
}

func TestOperatorReportRequiresOperator(t *testing.T) {
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", "11111111-1111-4111-8111-111111111111")
	t.Setenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS", "")
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", "")
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "22222222-2222-4222-8222-222222222222")
	if _, err := operatorReportActor(ctx); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("ordinary member allowed: %v", err)
	}
	ctx = context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "11111111-1111-4111-8111-111111111111")
	if _, err := operatorReportActor(ctx); err != nil {
		t.Fatalf("operator rejected: %v", err)
	}
}
