package main

import (
	"database/sql"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/activity"
)

var adminUsernamePattern = regexp.MustCompile(`^[\p{L}\p{N}_.-]{3,128}$`)

func (s *service) consolePlayerCreate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Kind        string `json:"kind"`
		Phone       string `json:"phone"`
		Country     string `json:"country"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	ctx := r.Context()
	username := strings.TrimSpace(latinDigits(in.Username))
	if username != "" && !adminUsernamePattern.MatchString(username) {
		s.writeError(w, badRequest("INVALID_USERNAME", "username", "3 to 128 letters, digits, _ . -"))
		return
	}
	if username != "" {
		var taken bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE lower(username)=lower($1))`, username).Scan(&taken); err != nil {
			s.writeError(w, err)
			return
		}
		if taken {
			s.writeError(w, fail(http.StatusConflict, "USERNAME_TAKEN", "username", "another player has this username"))
			return
		}
	}
	display := strings.TrimSpace(in.DisplayName)
	if len([]rune(display)) > 255 {
		s.writeError(w, badRequest("INVALID_DISPLAY_NAME", "displayName", "at most 255 letters"))
		return
	}
	out := map[string]any{"kind": in.Kind}
	var userID string
	switch in.Kind {
	case "phone":
		phone, err := s.otpPhone(ctx, in.Phone, in.Country)
		if err != nil {
			s.writeError(w, err)
			return
		}
		var taken bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_phone_account WHERE phone=$1)`, phone).Scan(&taken); err != nil {
			s.writeError(w, err)
			return
		}
		if taken {
			s.writeError(w, fail(http.StatusConflict, "PHONE_TAKEN", "phone", "another player has this phone number"))
			return
		}
		key, err := newLoginKey()
		if err != nil {
			s.writeError(w, err)
			return
		}
		uid, _, _, err := s.nk.AuthenticateCustom(ctx, key, username, true)
		if err != nil {
			s.writeError(w, err)
			return
		}
		err = s.inTx(ctx, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, `INSERT INTO publication_phone_account (phone, user_id, login_key) VALUES ($1,$2,$3)`, phone, uid, key); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, `UPDATE users SET metadata = jsonb_set(coalesce(metadata,'{}'::jsonb), '{phone}', to_jsonb($2::text)), update_time=now() WHERE id=$1`, uid, phone)
			return err
		})
		if err != nil {
			_ = s.nk.AccountDeleteId(ctx, uid, false)
			if strings.Contains(err.Error(), "publication_phone_account_pkey") {
				s.writeError(w, fail(http.StatusConflict, "PHONE_TAKEN", "phone", "another player has this phone number"))
				return
			}
			s.writeError(w, err)
			return
		}
		userID = uid
		out["phone"] = phone
	case "email":
		email := strings.ToLower(strings.TrimSpace(in.Email))
		if a, err := mail.ParseAddress(email); err != nil || a.Address != email || len(email) > 255 {
			s.writeError(w, badRequest("INVALID_EMAIL", "email", "enter an email address"))
			return
		}
		if n := len(in.Password); n < 8 || n > 128 {
			s.writeError(w, badRequest("INVALID_PASSWORD", "password", "8 to 128 characters"))
			return
		}
		var taken bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`, email).Scan(&taken); err != nil {
			s.writeError(w, err)
			return
		}
		if taken {
			s.writeError(w, fail(http.StatusConflict, "EMAIL_TAKEN", "email", "another player has this email"))
			return
		}
		uid, _, _, err := s.nk.AuthenticateEmail(ctx, email, in.Password, username, true)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE users SET metadata = jsonb_set(coalesce(metadata,'{}'::jsonb), '{emailVerified}', 'true'::jsonb), update_time=now() WHERE id=$1`, uid); err != nil {
			s.writeError(w, err)
			return
		}
		userID = uid
		out["email"] = email
	default:
		s.writeError(w, badRequest("INVALID_KIND", "kind", "phone or email"))
		return
	}
	if display != "" {
		if err := s.nk.AccountUpdateId(ctx, userID, "", nil, display, "", "", "", ""); err != nil {
			s.logger.Warn("Publication console player %s: display name not set: %v", userID, err)
		}
	}
	var name string
	if err := s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, userID).Scan(&name); err != nil {
		s.writeError(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindSignIn, Action: "created_by_admin", Actor: "console:" + consoleUser(r), Data: map[string]any{"kind": in.Kind}})
	out["userId"], out["username"], out["displayName"] = userID, name, display
	writeJSON(w, http.StatusCreated, out)
}
