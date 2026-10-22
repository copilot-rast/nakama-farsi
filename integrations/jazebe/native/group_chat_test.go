package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestChatTextAndDeleteRules(t *testing.T) {
	if text, err := validateChatText("  سلام  "); err != nil || text != "سلام" {
		t.Fatalf("trim: %q %v", text, err)
	}
	for _, value := range []string{" ", strings.Repeat("x", 2001), string([]byte{0xff})} {
		if _, err := validateChatText(value); err == nil {
			t.Errorf("accepted invalid message %q", value)
		}
	}
	now := time.Now()
	for _, tc := range []struct {
		role               string
		own, asAdmin       bool
		age                time.Duration
		allowed, moderated bool
	}{
		{"member", true, false, time.Minute, true, false},
		{"member", true, false, 16 * time.Minute, false, false},
		{"member", false, false, time.Minute, false, false},
		{"admin", false, false, 16 * time.Minute, true, true},
		{"creator", true, true, 16 * time.Minute, true, true},
		{"member", true, true, time.Minute, true, false},
	} {
		moderated, err := chatDeleteAllowed(tc.role, tc.own, tc.asAdmin, now.Add(-tc.age), now)
		if (err == nil) != tc.allowed || (tc.allowed && moderated != tc.moderated) {
			t.Errorf("%+v: moderated=%v err=%v", tc, moderated, err)
		}
	}
}

func TestGroupChatSubscribeRequiresRealtimeSession(t *testing.T) {
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "11111111-1111-4111-8111-111111111111")
	_, err := groupChatSubscribe(ctx, nil, nil, nil, `{"planetId":"22222222-2222-4222-8222-222222222222","groupId":"33333333-3333-4333-8333-333333333333"}`)
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("got %v, want missing realtime session", err)
	}
}
