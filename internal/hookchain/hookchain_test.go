package hookchain

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

func TestBeforeRunsEveryRegisteredHookInOrder(t *testing.T) {
	name := t.Name()
	var calls []string
	first := Before(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateDeviceRequest) (*api.AuthenticateDeviceRequest, error) {
		calls = append(calls, "a")
		in.Username += "a"
		return in, nil
	})
	second := Before(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateDeviceRequest) (*api.AuthenticateDeviceRequest, error) {
		calls = append(calls, "b")
		in.Username += "b"
		return in, nil
	})
	out, err := second(context.Background(), nil, nil, nil, &api.AuthenticateDeviceRequest{})
	if err != nil || out.Username != "ab" || len(calls) != 2 {
		t.Fatalf("out=%v err=%v calls=%v", out, err, calls)
	}
	calls = nil
	out, _ = first(context.Background(), nil, nil, nil, &api.AuthenticateDeviceRequest{})
	if out.Username != "a" || len(calls) != 1 {
		t.Fatalf("first dispatcher must keep its own list: %v %v", out, calls)
	}
}

func TestBeforeStopsOnErrorAndNil(t *testing.T) {
	name := t.Name()
	boom := errors.New("refused")
	ran := false
	Before(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
		return nil, boom
	})
	chain := Before(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
		ran = true
		return in, nil
	})
	if _, err := chain(context.Background(), nil, nil, nil, &rtapi.Envelope{}); !errors.Is(err, boom) || ran {
		t.Fatalf("err=%v ran=%v", err, ran)
	}
	other := t.Name() + "-nil"
	Before(other, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
		return nil, nil
	})
	chain = Before(other, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
		ran = true
		return in, nil
	})
	if out, err := chain(context.Background(), nil, nil, nil, &rtapi.Envelope{}); err != nil || out != nil || ran {
		t.Fatalf("a disabled message must stay disabled: %v %v %v", out, err, ran)
	}
}

func TestAfterRtRunsAllAndReturnsFirstError(t *testing.T) {
	name := t.Name()
	n := 0
	AfterRt(name, func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, *rtapi.Envelope, *rtapi.Envelope) error {
		n++
		return errors.New("one")
	})
	chain := AfterRt(name, func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, *rtapi.Envelope, *rtapi.Envelope) error {
		n++
		return nil
	})
	if err := chain(context.Background(), nil, nil, nil, nil, nil); err == nil || err.Error() != "one" || n != 2 {
		t.Fatalf("err=%v n=%d", err, n)
	}
}

func TestAfterRunsEveryPluginHookAndFitsTheNakamaSignature(t *testing.T) {
	name := t.Name()
	var calls []string
	After(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, in *api.AuthenticateGoogleRequest) error {
		calls = append(calls, "base:"+out.Token)
		return nil
	})
	chain := After(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, in *api.AuthenticateGoogleRequest) error {
		calls = append(calls, "economy:"+in.Account.Token)
		return errors.New("late")
	})
	var register func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, *api.Session, *api.AuthenticateGoogleRequest) error = chain
	err := register(context.Background(), nil, nil, nil, &api.Session{Token: "t"}, &api.AuthenticateGoogleRequest{Account: &api.AccountGoogle{Token: "g"}})
	if err == nil || err.Error() != "late" || len(calls) != 2 || calls[0] != "base:t" || calls[1] != "economy:g" {
		t.Fatalf("err=%v calls=%v", err, calls)
	}
}

func TestAfterInRunsEveryHookForInputOnlyMessages(t *testing.T) {
	name := t.Name()
	var calls []string
	AfterIn(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AddFriendsRequest) error {
		calls = append(calls, "base:"+in.Ids[0])
		return errors.New("first")
	})
	chain := AfterIn(name, func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AddFriendsRequest) error {
		calls = append(calls, "game:"+in.Ids[0])
		return nil
	})
	var register func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, *api.AddFriendsRequest) error = chain
	err := register(context.Background(), nil, nil, nil, &api.AddFriendsRequest{Ids: []string{"x"}})
	if err == nil || err.Error() != "first" || len(calls) != 2 || calls[0] != "base:x" || calls[1] != "game:x" {
		t.Fatalf("err=%v calls=%v", err, calls)
	}
}
