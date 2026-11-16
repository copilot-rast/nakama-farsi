package apierr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
)

func TestFillAndDigits(t *testing.T) {
	got := Fill("wait {retryAfter}s for {who} {missing}", map[string]any{"retryAfter": float64(30), "who": "x"}, true)
	if got != "wait ۳۰s for x {missing}" {
		t.Fatalf("got %q", got)
	}
	if Fill("scope {scope}", map[string]any{"scope": "e2e1:write"}, true) != "scope e2e1:write" {
		t.Fatal("strings keep their digits")
	}
	if Fill("level {unlockLevel}", map[string]any{"unlockLevel": 7}, false) != "level 7" {
		t.Fatal("plain digits")
	}
}

func TestBuildUsesDefaultsAndMeta(t *testing.T) {
	ctx := context.Background()
	env := Build(ctx, nil, "fa", TooManyRequests("TOO_MANY_REQUESTS", 12).With("unlockLevel", 3))
	if env.Status != http.StatusTooManyRequests || env.RetryAfterSeconds == nil || *env.RetryAfterSeconds != 12 {
		t.Fatalf("bad envelope %+v", env)
	}
	if env.Message == "" || env.Message == "TOO_MANY_REQUESTS" {
		t.Fatalf("no localized message: %q", env.Message)
	}
	if env.Meta["unlockLevel"] != 3 {
		t.Fatal("meta lost")
	}
	en := Build(ctx, nil, "en", runtime.NewError("GIFT_EXPIRED", int(codes.FailedPrecondition)))
	if en.Code != "GIFT_EXPIRED" || en.Message != "This gift code has expired." {
		t.Fatalf("runtime error not mapped: %+v", en)
	}
	in := Build(ctx, nil, "en", errors.New("boom"))
	if in.Code != "INTERNAL_ERROR" || in.Status != 500 {
		t.Fatalf("internal: %+v", in)
	}
}

func TestToRuntimeEnvelope(t *testing.T) {
	err := ToRuntime(context.Background(), nil, "", "fa", NotFound("DEVKIT_UNKNOWN_MODULE").With("module", "x"))
	var rt *runtime.Error
	if !errors.As(err, &rt) || rt.Code != int(codes.NotFound) {
		t.Fatalf("bad runtime error %v", err)
	}
	var env Envelope
	if json.Unmarshal([]byte(rt.Message), &env) != nil || env.Code != "DEVKIT_UNKNOWN_MODULE" || env.Message != "ماژولی به نام x وجود ندارد." {
		t.Fatalf("bad envelope %s", rt.Message)
	}
}

func TestRegisterOverridesDefaults(t *testing.T) {
	Register("sample", map[string]Default{"SAMPLE_LOCKED": {Status: 423, Messages: map[string]string{"en": "Unlocks at level {unlockLevel}."}}})
	d := Defaults()["SAMPLE_LOCKED"]
	if d.Status != 423 || d.Modules[0] != "sample" {
		t.Fatalf("register: %+v", d)
	}
	env := Build(context.Background(), nil, "en", New(423, "SAMPLE_LOCKED").With("unlockLevel", 5))
	if env.Message != "Unlocks at level 5." {
		t.Fatalf("got %q", env.Message)
	}
}

func TestForModuleKeepsBareErrorInCodeMode(t *testing.T) {
	bare := runtime.NewError("GIFT_EXPIRED", int(codes.FailedPrecondition))
	if got := ForModule(context.Background(), nil, "fa", bare, New(409, "GIFT_EXPIRED")); got != bare {
		t.Fatalf("code mode changed the error: %v", got)
	}
}
