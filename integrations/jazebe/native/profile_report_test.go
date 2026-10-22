package main

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProfileReportValidation(t *testing.T) {
	reporter := "11111111-1111-4111-8111-111111111111"
	target := "22222222-2222-4222-8222-222222222222"
	planet := "33333333-3333-4333-8333-333333333333"
	note := "  inappropriate name  "
	valid, err := validateProfileReport(profileReportInput{PlanetID: planet, UserID: target, Reason: "  inappropriate profile  ", Note: &note}, reporter)
	if err != nil || valid.Reason != "inappropriate profile" || *valid.Note != "inappropriate name" {
		t.Fatalf("valid report: %+v %v", valid, err)
	}
	for _, in := range []profileReportInput{
		{PlanetID: planet, UserID: reporter, Reason: "self"},
		{PlanetID: planet, UserID: target, Reason: " "},
		{PlanetID: planet, UserID: target, Reason: strings.Repeat("x", 65)},
		{PlanetID: planet, UserID: target, Reason: "reason", Note: stringPtr(strings.Repeat("x", 513))},
	} {
		if _, err = validateProfileReport(in, reporter); status.Code(err) != codes.InvalidArgument {
			t.Errorf("invalid report accepted: %+v %v", in, err)
		}
	}
}

func stringPtr(v string) *string { return &v }
