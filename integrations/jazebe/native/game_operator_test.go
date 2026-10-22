package main

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGameTemplateValidation(t *testing.T) {
	name, engine := "Tap", "tap"
	valid := gameOperatorInput{Key: "tap-rush", Name: &name, Engine: &engine}
	if err := validateGameTemplate(valid, true); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []gameOperatorInput{
		{Key: "BAD KEY", Name: &name, Engine: &engine},
		{Key: "tap-rush", Name: &name},
		{Key: "tap-rush", Name: &name, Engine: &engine, DurationSec: intPtr(601)},
	} {
		if err := validateGameTemplate(tc, true); status.Code(err) != codes.InvalidArgument {
			t.Errorf("input %+v: got %v", tc, err)
		}
	}
}

func TestGameOperatorPage(t *testing.T) {
	in := gameOperatorInput{}
	if err := gameOperatorPage(&in); err != nil || in.Page != 1 || in.Limit != 20 {
		t.Fatalf("defaults: %+v %v", in, err)
	}
	in.Limit = 101
	if status.Code(gameOperatorPage(&in)) != codes.InvalidArgument {
		t.Fatal("oversized page accepted")
	}
}

func intPtr(v int) *int { return &v }
