package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountRevocationNakama struct {
	runtime.NakamaModule
	loggedOut bool
	listErr   error
}

type accountTestLogger struct{ runtime.Logger }

func (accountTestLogger) Error(string, ...interface{}) {}

func (n *accountRevocationNakama) SessionLogout(userID, token, refreshToken string) error {
	n.loggedOut = true
	return nil
}

func (n *accountRevocationNakama) StreamUserList(mode uint8, subject, subcontext, label string, hidden, visible bool) ([]runtime.Presence, error) {
	return nil, n.listErr
}

func TestOperatorAccountActionValidation(t *testing.T) {
	id := "11111111-1111-4111-8111-111111111111"
	in, err := validateOperatorAccountAction(operatorAccountAction{UserID: id, Reason: "  policy violation  "})
	if err != nil || in.Reason != "policy violation" {
		t.Fatalf("valid reason: %+v %v", in, err)
	}
	for _, input := range []operatorAccountAction{
		{UserID: "invalid", Reason: "valid reason"},
		{UserID: id, Reason: " x "},
		{UserID: id, Reason: strings.Repeat("a", 513)},
	} {
		if _, err = validateOperatorAccountAction(input); status.Code(err) != codes.InvalidArgument {
			t.Errorf("accepted invalid action %+v: %v", input, err)
		}
	}
}

func TestAccountRevocationFailureIsReported(t *testing.T) {
	nk := &accountRevocationNakama{listErr: errors.New("tracker unavailable")}
	if err := revokeAccountSessions(context.Background(), nk, "11111111-1111-4111-8111-111111111111"); err == nil || !nk.loggedOut {
		t.Fatalf("revocation must surface tracker failure after cache logout: loggedOut=%v err=%v", nk.loggedOut, err)
	}
}

func TestAccountRevocationSuccess(t *testing.T) {
	nk := &accountRevocationNakama{}
	if err := revokeAccountSessions(context.Background(), nk, "11111111-1111-4111-8111-111111111111"); err != nil || !nk.loggedOut {
		t.Fatalf("revocation failed: loggedOut=%v err=%v", nk.loggedOut, err)
	}
}

func TestAccountRevocationWorkerConfig(t *testing.T) {
	t.Setenv("JAZEBE_ACCOUNT_REVOCATION_WORKER_ENABLED", "false")
	if err := startAccountRevocationWorker(context.Background(), nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JAZEBE_ACCOUNT_REVOCATION_WORKER_ENABLED", "yes")
	if err := startAccountRevocationWorker(context.Background(), nil, nil, nil); err == nil {
		t.Fatal("invalid worker configuration accepted")
	}
}

func TestAccountRestrictionProtectsSuperOperator(t *testing.T) {
	actor := "11111111-1111-4111-8111-111111111111"
	target := "22222222-2222-4222-8222-222222222222"
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", actor)
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", target)
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, actor)
	_, err := operatorAccountSuspend(ctx, nil, nil, nil, `{"userId":"`+target+`","reason":"verified policy breach"}`)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("ordinary operator changed super operator: %v", err)
	}
}
