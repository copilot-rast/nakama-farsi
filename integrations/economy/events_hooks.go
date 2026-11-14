package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

func (s *service) loginEvent(ctx context.Context, session *api.Session, method string) {
	if session == nil {
		return
	}
	uid := restrict.TokenUserID(session.Token)
	if uid == "" {
		return
	}
	if r, err := restrict.Active(ctx, s.db, uid, restrict.KindSuspend); err == nil && r != nil {
		return
	}
	s.noteOfferLogin(ctx, uid)
	s.recordServerEvent(ctx, uid, "login", 1, map[string]any{"method": method, "created": session.Created})
}

func (s *service) registerEventHooks(initializer runtime.Initializer) error {
	steps := []func() error{
		func() error {
			return initializer.RegisterAfterAuthenticateDevice(hookchain.After("AuthenticateDevice", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateDeviceRequest) error {
				s.loginEvent(ctx, out, "device")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateCustom(hookchain.After("AuthenticateCustom", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateCustomRequest) error {
				s.loginEvent(ctx, out, "custom")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateEmail(hookchain.After("AuthenticateEmail", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateEmailRequest) error {
				s.loginEvent(ctx, out, "email")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateGoogle(hookchain.After("AuthenticateGoogle", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateGoogleRequest) error {
				s.loginEvent(ctx, out, "google")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateApple(hookchain.After("AuthenticateApple", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateAppleRequest) error {
				s.loginEvent(ctx, out, "apple")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateFacebook(hookchain.After("AuthenticateFacebook", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateFacebookRequest) error {
				s.loginEvent(ctx, out, "facebook")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateSteam(hookchain.After("AuthenticateSteam", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateSteamRequest) error {
				s.loginEvent(ctx, out, "steam")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateGameCenter(hookchain.After("AuthenticateGameCenter", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateGameCenterRequest) error {
				s.loginEvent(ctx, out, "gamecenter")
				return nil
			}))
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
