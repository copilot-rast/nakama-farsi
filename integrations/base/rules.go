package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"google.golang.org/grpc/codes"
)

var latinUsername = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

var usernameCharsets = []string{"any", "latin"}

type accountRules struct {
	UsernameMin     int       `json:"usernameMin"`
	UsernameMax     int       `json:"usernameMax"`
	UsernameCharset string    `json:"usernameCharset"`
	PasswordMin     int       `json:"passwordMin"`
	PasswordLetter  bool      `json:"passwordLetter"`
	PasswordDigit   bool      `json:"passwordDigit"`
	DisplayNameMax  int       `json:"displayNameMax"`
	UpdatedBy       string    `json:"updatedBy,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func defaultRules() accountRules {
	return accountRules{UsernameMin: 1, UsernameMax: 128, UsernameCharset: "any", PasswordMin: 8, DisplayNameMax: 255}
}

func loadRules(ctx context.Context, q queryer) (accountRules, error) {
	r := defaultRules()
	err := q.QueryRowContext(ctx, `SELECT username_min, username_max, username_charset, password_min, password_letter, password_digit, display_name_max, updated_by, updated_at FROM base_account_rules WHERE id=1`).
		Scan(&r.UsernameMin, &r.UsernameMax, &r.UsernameCharset, &r.PasswordMin, &r.PasswordLetter, &r.PasswordDigit, &r.DisplayNameMax, &r.UpdatedBy, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultRules(), nil
	}
	return r, err
}

func (r accountRules) clean() (accountRules, error) {
	r.UsernameCharset = strings.TrimSpace(r.UsernameCharset)
	if r.UsernameCharset == "" {
		r.UsernameCharset = "any"
	}
	switch {
	case r.UsernameMin < 1 || r.UsernameMin > 128:
		return r, badRequest("INVALID_RULE", "usernameMin", "between 1 and 128")
	case r.UsernameMax < 1 || r.UsernameMax > 128:
		return r, badRequest("INVALID_RULE", "usernameMax", "between 1 and 128")
	case r.UsernameMin > r.UsernameMax:
		return r, badRequest("INVALID_RULE", "usernameMax", "must not be less than the minimum")
	case !validKeyType(usernameCharsets, r.UsernameCharset):
		return r, badRequest("INVALID_RULE", "usernameCharset", "any or latin")
	case r.PasswordMin < 8 || r.PasswordMin > 128:
		return r, badRequest("INVALID_RULE", "passwordMin", "between 8 and 128")
	case r.DisplayNameMax < 1 || r.DisplayNameMax > 255:
		return r, badRequest("INVALID_RULE", "displayNameMax", "between 1 and 255")
	}
	return r, nil
}

func (r accountRules) usernameError(name string) string {
	n := utf8.RuneCountInString(name)
	switch {
	case n < r.UsernameMin:
		return "USERNAME_TOO_SHORT"
	case n > r.UsernameMax:
		return "USERNAME_TOO_LONG"
	case r.UsernameCharset == "latin" && !latinUsername.MatchString(name):
		return "USERNAME_INVALID_CHARACTERS"
	}
	return ""
}

func (r accountRules) passwordError(pw string) string {
	if utf8.RuneCountInString(pw) < r.PasswordMin {
		return "PASSWORD_TOO_SHORT"
	}
	var letter, digit bool
	for _, c := range pw {
		letter = letter || unicode.IsLetter(c)
		digit = digit || unicode.IsDigit(c)
	}
	if r.PasswordLetter && !letter {
		return "PASSWORD_NEEDS_LETTER"
	}
	if r.PasswordDigit && !digit {
		return "PASSWORD_NEEDS_DIGIT"
	}
	return ""
}

func (r accountRules) displayNameError(name string) string {
	if utf8.RuneCountInString(name) > r.DisplayNameMax {
		return "DISPLAY_NAME_TOO_LONG"
	}
	return ""
}

func (r accountRules) client() map[string]any {
	return map[string]any{"usernameMin": r.UsernameMin, "usernameMax": r.UsernameMax, "usernameCharset": r.UsernameCharset,
		"passwordMin": r.PasswordMin, "passwordLetter": r.PasswordLetter, "passwordDigit": r.PasswordDigit, "displayNameMax": r.DisplayNameMax}
}

func ruleError(code string) error {
	return runtime.NewError(code, int(codes.InvalidArgument))
}

func (s *service) rulesGet(w http.ResponseWriter, r *http.Request) {
	rules, err := loadRules(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (s *service) rulesPut(w http.ResponseWriter, r *http.Request) {
	in := defaultRules()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	rules, err := in.clean()
	if err != nil {
		s.writeError(w, err)
		return
	}
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO base_account_rules (id, username_min, username_max, username_charset, password_min, password_letter, password_digit, display_name_max, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,now())
		ON CONFLICT (id) DO UPDATE SET username_min=EXCLUDED.username_min, username_max=EXCLUDED.username_max, username_charset=EXCLUDED.username_charset,
		password_min=EXCLUDED.password_min, password_letter=EXCLUDED.password_letter, password_digit=EXCLUDED.password_digit,
		display_name_max=EXCLUDED.display_name_max, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		rules.UsernameMin, rules.UsernameMax, rules.UsernameCharset, rules.PasswordMin, rules.PasswordLetter, rules.PasswordDigit, rules.DisplayNameMax, consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.rulesGet(w, r)
}

func (s *service) rpcAccountRules(ctx context.Context, userID, payload string) (any, error) {
	rules, err := loadRules(ctx, s.db)
	if err != nil {
		return nil, err
	}
	return rules.client(), nil
}

func (s *service) checkNewUsername(ctx context.Context, username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil
	}
	var taken bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username=$1)`, username).Scan(&taken); err != nil {
		return s.hookFailure(err)
	}
	if taken {
		return nil
	}
	rules, err := loadRules(ctx, s.db)
	if err != nil {
		return s.hookFailure(err)
	}
	if code := rules.usernameError(username); code != "" {
		return ruleError(code)
	}
	return nil
}

func (s *service) hookFailure(err error) error {
	s.logger.Error("Base account rules check failed: %v", err)
	return runtime.NewError("INTERNAL_ERROR", int(codes.Internal))
}

func (s *service) beforeAuthEmail(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateEmailRequest) (*api.AuthenticateEmailRequest, error) {
	email := strings.ToLower(strings.TrimSpace(in.GetAccount().GetEmail()))
	if email == "" {
		return in, nil
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists); err != nil {
		return nil, s.hookFailure(err)
	}
	if exists || (in.Create != nil && !in.Create.Value) {
		return in, nil
	}
	rules, err := loadRules(ctx, s.db)
	if err != nil {
		return nil, s.hookFailure(err)
	}
	if code := rules.passwordError(in.GetAccount().GetPassword()); code != "" {
		return nil, ruleError(code)
	}
	if err := s.checkNewUsername(ctx, in.Username); err != nil {
		return nil, err
	}
	return in, nil
}

func (s *service) beforeLinkEmail(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AccountEmail) (*api.AccountEmail, error) {
	rules, err := loadRules(ctx, s.db)
	if err != nil {
		return nil, s.hookFailure(err)
	}
	if code := rules.passwordError(in.GetPassword()); code != "" {
		return nil, ruleError(code)
	}
	return in, nil
}

func (s *service) beforeUpdateAccount(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.UpdateAccountRequest) (*api.UpdateAccountRequest, error) {
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if in.Username != nil {
		name := strings.TrimSpace(in.Username.Value)
		var current string
		if userID != "" {
			if err := s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, userID).Scan(&current); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, s.hookFailure(err)
			}
		}
		if name != "" && name != current {
			if err := s.checkNewUsername(ctx, name); err != nil {
				return nil, err
			}
			done, err := s.usernameThroughPolicy(ctx, userID, name)
			if err != nil {
				return nil, err
			}
			if done {
				in.Username.Value = name
			}
		}
	}
	if in.DisplayName != nil {
		rules, err := loadRules(ctx, s.db)
		if err != nil {
			return nil, s.hookFailure(err)
		}
		if code := rules.displayNameError(strings.TrimSpace(in.DisplayName.Value)); code != "" {
			return nil, ruleError(code)
		}
	}
	return in, nil
}

func (s *service) emailUser(ctx context.Context, in *api.AuthenticateEmailRequest) string {
	if email := strings.ToLower(strings.TrimSpace(in.GetAccount().GetEmail())); email != "" {
		return s.userBy(ctx, `SELECT id::text FROM users WHERE email=$1`, email)
	}
	return s.userBy(ctx, `SELECT id::text FROM users WHERE username=$1`, in.Username)
}

func (s *service) registerRuleHooks(initializer runtime.Initializer) error {
	checks := []func() error{
		func() error {
			return initializer.RegisterBeforeAuthenticateEmail(hookchain.Before("AuthenticateEmail", func(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AuthenticateEmailRequest) (*api.AuthenticateEmailRequest, error) {
				if err := s.refuseSuspended(ctx, s.emailUser(ctx, in)); err != nil {
					return nil, err
				}
				return s.beforeAuthEmail(ctx, l, db, nk, in)
			}))
		},
		func() error {
			return initializer.RegisterBeforeLinkEmail(hookchain.Before("LinkEmail", s.beforeLinkEmail))
		},
		func() error {
			return initializer.RegisterBeforeUpdateAccount(hookchain.Before("UpdateAccount", s.beforeUpdateAccount))
		},
		func() error {
			return initializer.RegisterBeforeAuthenticateDevice(hookchain.Before("AuthenticateDevice", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateDeviceRequest) (*api.AuthenticateDeviceRequest, error) {
				if err := s.refuseSuspended(ctx, s.userBy(ctx, `SELECT user_id::text FROM user_device WHERE id=$1`, in.GetAccount().GetId())); err != nil {
					return nil, err
				}
				return in, s.checkNewUsername(ctx, in.Username)
			}))
		},
		func() error {
			return initializer.RegisterBeforeAuthenticateCustom(hookchain.Before("AuthenticateCustom", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateCustomRequest) (*api.AuthenticateCustomRequest, error) {
				if err := s.refuseSuspended(ctx, s.userBy(ctx, `SELECT id::text FROM users WHERE custom_id=$1`, in.GetAccount().GetId())); err != nil {
					return nil, err
				}
				return in, s.checkNewUsername(ctx, in.Username)
			}))
		},
		func() error {
			return initializer.RegisterBeforeAuthenticateGoogle(hookchain.Before("AuthenticateGoogle", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateGoogleRequest) (*api.AuthenticateGoogleRequest, error) {
				return in, s.checkNewUsername(ctx, in.Username)
			}))
		},
		func() error {
			return initializer.RegisterBeforeAuthenticateApple(hookchain.Before("AuthenticateApple", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateAppleRequest) (*api.AuthenticateAppleRequest, error) {
				return in, s.checkNewUsername(ctx, in.Username)
			}))
		},
		func() error {
			return initializer.RegisterBeforeAuthenticateFacebook(hookchain.Before("AuthenticateFacebook", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateFacebookRequest) (*api.AuthenticateFacebookRequest, error) {
				return in, s.checkNewUsername(ctx, in.Username)
			}))
		},
		func() error {
			return initializer.RegisterBeforeAuthenticateSteam(hookchain.Before("AuthenticateSteam", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateSteamRequest) (*api.AuthenticateSteamRequest, error) {
				return in, s.checkNewUsername(ctx, in.Username)
			}))
		},
		func() error {
			return initializer.RegisterBeforeAuthenticateGameCenter(hookchain.Before("AuthenticateGameCenter", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateGameCenterRequest) (*api.AuthenticateGameCenterRequest, error) {
				return in, s.checkNewUsername(ctx, in.Username)
			}))
		},
	}
	for _, register := range checks {
		if err := register(); err != nil {
			return err
		}
	}
	return nil
}
