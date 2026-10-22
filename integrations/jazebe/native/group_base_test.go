package main

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGroupCreationFields(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	valid := groupCreateInput{PlanetID: id, Name: "  ستاره  ", Mode: "open", Cap: 10, Color: "#2F6EBC", Icon: "star"}
	got, err := validateGroupCreate(valid)
	if err != nil || got.Name != "ستاره" {
		t.Fatalf("valid Persian group rejected: %#v, %v", got, err)
	}
	for _, change := range []func(*groupCreateInput){
		func(i *groupCreateInput) { i.PlanetID = "bad" },
		func(i *groupCreateInput) { i.Name = "   " },
		func(i *groupCreateInput) { i.Name = strings.Repeat("x", 65) },
		func(i *groupCreateInput) { i.Description = strings.Repeat("x", 281) },
		func(i *groupCreateInput) { i.Mode = "closed" },
		func(i *groupCreateInput) { i.Cap = 0 },
		func(i *groupCreateInput) { i.Color = "#fff" },
		func(i *groupCreateInput) { i.Icon = "unknown" },
	} {
		in := valid
		change(&in)
		if _, err := validateGroupCreate(in); err == nil {
			t.Fatalf("invalid group accepted: %#v", in)
		}
	}
}

func TestGroupCreateRequiresSession(t *testing.T) {
	if _, err := groupCreate(context.Background(), nil, nil, nil, `{}`); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("anonymous group creation accepted: %v", err)
	}
}

func TestGroupJoinRequiresSessionAndReferences(t *testing.T) {
	if _, err := groupJoin(context.Background(), nil, nil, nil, `{}`); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("anonymous group join accepted: %v", err)
	}
	for _, handler := range []func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error){
		groupMembershipState, groupCancelRequest, groupLeave,
	} {
		if _, err := handler(context.Background(), nil, nil, nil, `{}`); status.Code(err) != codes.Unauthenticated {
			t.Fatalf("anonymous group operation accepted: %v", err)
		}
	}
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := readGroupRef(`{"planetId":"` + id + `","groupId":"bad"}`); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid group reference accepted: %v", err)
	}
	if _, err := readGroupRef(`{"planetId":"` + id + `","groupId":"` + id + `"}`); err != nil {
		t.Fatalf("valid group reference rejected: %v", err)
	}
}

func TestGroupDecisionValidationAndSession(t *testing.T) {
	if _, err := groupDecideRequest(context.Background(), nil, nil, nil, `{}`); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("anonymous manager decision accepted: %v", err)
	}
	id := uuid.Must(uuid.NewV4()).String()
	valid := groupDecisionInput{PlanetID: id, GroupID: id, RequestID: id, Decision: "approve"}
	if _, err := validateGroupDecision(valid); err != nil {
		t.Fatalf("valid decision rejected: %v", err)
	}
	valid.Decision = "maybe"
	if _, err := validateGroupDecision(valid); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid decision accepted: %v", err)
	}
}

func TestGroupQueueAndPersianFormatting(t *testing.T) {
	if _, err := groupRequestQueue(context.Background(), nil, nil, nil, `{}`); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("anonymous request queue accepted: %v", err)
	}
	if faNumber(123) != "۱۲۳" {
		t.Fatalf("Persian digits broken: %s", faNumber(123))
	}
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	if groupRelativeFa(now.Add(-2*time.Hour), now) != "۲ ساعت پیش" {
		t.Fatal("relative Persian time broken")
	}
	if groupInitial(" ستاره ") != "س" {
		t.Fatal("Persian initial broken")
	}
	if groupMemberColor("user-1") != groupMemberColor("user-1") {
		t.Fatal("unstable member color")
	}
}

func TestGroupBulkDecisionValidatesBeforeWrites(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, id)
	for _, payload := range []string{
		`{"planetId":"` + id + `","groupId":"` + id + `","decision":"approve","requestIds":[]}`,
		`{"planetId":"` + id + `","groupId":"` + id + `","decision":"approve","requestIds":["bad"]}`,
		`{"planetId":"` + id + `","groupId":"` + id + `","decision":"maybe","requestIds":["` + id + `"]}`,
	} {
		if _, err := groupDecideMany(ctx, nil, nil, nil, payload); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid bulk decision accepted: %s: %v", payload, err)
		}
	}
}

func TestGroupInviteGuards(t *testing.T) {
	for _, handler := range []func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error){
		groupInviteMint, groupInvitePreview, groupInviteJoin, groupInviteRevoke,
	} {
		if _, err := handler(context.Background(), nil, nil, nil, `{}`); status.Code(err) != codes.Unauthenticated {
			t.Fatalf("anonymous group invite operation accepted: %v", err)
		}
	}
	if _, err := groupInviteToken(`{"token":"short"}`); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("short invite token accepted: %v", err)
	}
	id := uuid.Must(uuid.NewV4()).String()
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, id)
	result, err := groupInvitePreview(ctx, nil, nil, nil, `{"token":"short"}`)
	if err != nil || !strings.Contains(result, `"status":"invalid"`) {
		t.Fatalf("invalid invite preview: %s %v", result, err)
	}
}
