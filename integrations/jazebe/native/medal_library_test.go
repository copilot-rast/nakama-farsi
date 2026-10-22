package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMedalLibraryConditionSchema(t *testing.T) {
	valid := []struct{ kind, schema string }{
		{"activity_count", `{"activity":"purchase","count":{"min":1,"max":100}}`},
		{"duration", `{"unit":"day","days":{"min":30,"max":730}}`},
		{"special_event", `{"event":"string"}`},
	}
	for _, v := range valid {
		if problems := medalSchemaProblems(v.kind, json.RawMessage(v.schema)); len(problems) != 0 {
			t.Fatalf("valid %s schema rejected: %v", v.kind, problems)
		}
	}
	invalid := []struct{ kind, schema string }{
		{"activity_count", `{"activity":"wallet","count":{"min":1,"max":10}}`},
		{"activity_count", `{"activity":"purchase","count":{"min":2,"max":1}}`},
		{"duration", `{"unit":"week","days":{"min":1,"max":2}}`},
		{"duration", `{"days":{"min":1.5,"max":2}}`},
		{"special_event", `{"event":"fixed"}`},
		{"special_event", `[]`},
	}
	for _, v := range invalid {
		if problems := medalSchemaProblems(v.kind, json.RawMessage(v.schema)); len(problems) == 0 {
			t.Fatalf("invalid %s schema accepted: %s", v.kind, v.schema)
		}
	}
}

func TestMedalLibraryValidationAndPrivileges(t *testing.T) {
	base := medalLibraryInput{Key: "first_purchase", Name: "First purchase", ThemeCategory: "purchase", ConditionType: "activity_count"}
	if _, err := validateMedalLibraryInput(base, true); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []medalLibraryInput{
		{Key: "Bad Key", Name: base.Name, ThemeCategory: base.ThemeCategory, ConditionType: base.ConditionType},
		{Key: base.Key, Name: strings.Repeat("a", 129), ThemeCategory: base.ThemeCategory, ConditionType: base.ConditionType},
		{Key: base.Key, Name: base.Name, ThemeCategory: "unknown", ConditionType: base.ConditionType},
	} {
		if _, err := validateMedalLibraryInput(bad, true); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid medal accepted: %+v, %v", bad, err)
		}
	}
	actor := uuid.Must(uuid.NewV4()).String()
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, actor)
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", actor)
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", "")
	if _, err := medalOperator(ctx, false); err != nil {
		t.Fatal(err)
	}
	if _, err := medalOperator(ctx, true); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("regular operator gained super privileges: %v", err)
	}
}

func TestMedalLibraryRejectsKeyChangesAndUnknownFields(t *testing.T) {
	actor := uuid.Must(uuid.NewV4()).String()
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, actor)
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", actor)
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", actor)
	for _, payload := range []string{
		`{"id":"` + actor + `","key":"renamed"}`,
		`{"id":"` + actor + `","unsupported":true}`,
	} {
		if _, err := medalLibraryUpdate(ctx, nil, nil, nil, payload); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("unsafe update accepted: %s, %v", payload, err)
		}
	}
}
