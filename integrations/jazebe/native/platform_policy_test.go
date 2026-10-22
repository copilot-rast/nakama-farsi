package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPlatformPolicyPatchValidation(t *testing.T) {
	valid := map[string]json.RawMessage{
		"missionRewardPointMax": []byte(`"1000"`),
		"lotteryFeePercentMin":  []byte(`0.29`),
		"groupMaxCountMax":      []byte(`null`),
		"termsVersion":          []byte(`"2026.10"`),
	}
	parsed, err := validatePolicyPatch(valid)
	if err != nil || parsed["missionRewardPointMax"] != float64(1000) || parsed["groupMaxCountMax"] != nil {
		t.Fatalf("valid policy patch rejected: %#v %v", parsed, err)
	}
	for _, bad := range []map[string]json.RawMessage{
		{"unknown": []byte(`1`)},
		{"missionRewardPointMax": []byte(`null`)},
		{"missionRewardPointMax": []byte(`-1`)},
		{"otpLength": []byte(`9`)},
		{"lotteryFeePercentMax": []byte(`10.999`)},
		{"termsVersion": []byte(`"bad version"`)},
	} {
		if _, err := validatePolicyPatch(bad); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid policy patch accepted: %#v %v", bad, err)
		}
	}
}

func TestPlatformPolicyWriteRequiresManagerialOperator(t *testing.T) {
	actor := uuid.Must(uuid.NewV4()).String()
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, actor)
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", actor)
	t.Setenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS", "")
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", "")
	if _, err := platformPolicyUpdate(ctx, nil, nil, nil, `{}`); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("ordinary operator could edit policy: %v", err)
	}
}
