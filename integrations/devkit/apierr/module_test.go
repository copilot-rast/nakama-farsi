package apierr_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr/apierrtest"
	"google.golang.org/grpc/codes"
)

func TestForModuleEnvelopeKeepsGRPCCode(t *testing.T) {
	db := apierrtest.DB(map[string]any{"mode": "code", "moduleMode": "envelope"})
	defer db.Close()
	ctx := context.Background()
	bare := runtime.NewError("GIFT_EXPIRED", int(codes.FailedPrecondition))
	err := apierr.ForModule(ctx, db, "en", bare, apierr.New(409, "GIFT_EXPIRED").WithField("code").With("n", 1))
	var rt *runtime.Error
	if !errors.As(err, &rt) || rt.Code != int(codes.FailedPrecondition) {
		t.Fatalf("grpc code changed: %v", err)
	}
	var env apierr.Envelope
	if json.Unmarshal([]byte(rt.Message), &env) != nil || env.Code != "GIFT_EXPIRED" || env.Status != 409 || env.Field != "code" || env.Meta["n"] != float64(1) || env.Message != "This gift code has expired." || env.Lang != "en" {
		t.Fatalf("bad envelope %s", rt.Message)
	}

	err = apierr.ForModule(ctx, db, "fa", runtime.NewError("UNAUTHENTICATED", int(codes.Unauthenticated)), nil)
	if !errors.As(err, &rt) || rt.Code != int(codes.Unauthenticated) || json.Unmarshal([]byte(rt.Message), &env) != nil || env.Status != 401 || env.Message != "لطفاً دوباره وارد شوید." {
		t.Fatalf("bare runtime error: %v", err)
	}

	free := runtime.NewError("something odd", int(codes.Internal))
	if apierr.ForModule(ctx, db, "fa", free, nil) != free {
		t.Fatal("free text was wrapped")
	}
}
