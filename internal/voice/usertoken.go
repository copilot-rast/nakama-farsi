package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	UserTokenMaxLifetime = 10 * time.Minute
	UserTokenSecretMin   = 32
	UserTokenSecretMax   = 512
	UserTokenTestSubject = "nakama-connection-test"
)

var (
	appIDPattern   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	subjectPattern = regexp.MustCompile(`^[A-Za-z0-9._~@+-]{1,128}$`)
)

var UserTokenFields = []string{"appId", "tokenIssuer", "tokenKeyId", "tokenSecret"}

type UserTokenSigner struct {
	AppID  string
	Issuer string
	KeyID  string
	Secret string
}

type userTokenClaims struct {
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
	Subject  string `json:"sub"`
	Name     string `json:"name,omitempty"`
	IssuedAt int64  `json:"iat"`
	Expires  int64  `json:"exp"`
	ID       string `json:"jti"`
}

func printable(s string, max int) bool {
	if s == "" || len(s) > max {
		return false
	}
	return !strings.ContainsFunc(s, func(r rune) bool { return r < 0x21 || r > 0x7e })
}

func checkUserTokenFields(cfg Config) error {
	if v := cfg.Get("appId"); v != "" && !appIDPattern.MatchString(v) {
		return &FieldError{Code: "COMMS_APP_ID_INVALID", Field: "appId", Message: "the comms app id is a UUID"}
	}
	if v := cfg.Get("tokenIssuer"); v != "" && len(v) > 255 {
		return &FieldError{Code: "COMMS_ISSUER_INVALID", Field: "tokenIssuer", Message: "1 to 255 characters"}
	}
	if v := cfg.Get("tokenKeyId"); v != "" && !printable(v, 128) {
		return &FieldError{Code: "COMMS_KEY_ID_INVALID", Field: "tokenKeyId", Message: "1 to 128 printable characters without spaces"}
	}
	if v := cfg.Get("tokenSecret"); v != "" {
		if len(v) < UserTokenSecretMin {
			return &FieldError{Code: "COMMS_SECRET_TOO_SHORT", Field: "tokenSecret", Message: "at least 32 bytes"}
		}
		if len(v) > UserTokenSecretMax {
			return &FieldError{Code: "COMMS_SECRET_TOO_LONG", Field: "tokenSecret", Message: "at most 512 bytes"}
		}
	}
	return nil
}

func NewUserTokenSigner(cfg Config) (UserTokenSigner, error) {
	for _, name := range UserTokenFields {
		if cfg.Get(name) == "" {
			return UserTokenSigner{}, &FieldError{Code: "VOICE_FIELD_REQUIRED", Field: name, Message: "needed for comms user tokens"}
		}
	}
	if err := checkUserTokenFields(cfg); err != nil {
		return UserTokenSigner{}, err
	}
	return UserTokenSigner{AppID: strings.ToLower(cfg.Get("appId")), Issuer: cfg.Get("tokenIssuer"), KeyID: cfg.Get("tokenKeyId"), Secret: cfg.Get("tokenSecret")}, nil
}

func UserTokenReady(cfg Config) bool {
	_, err := NewUserTokenSigner(cfg)
	return err == nil
}

func CleanDisplayName(name string) string {
	if !utf8.ValidString(name) {
		name = strings.ToValidUTF8(name, "")
	}
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > 100 {
		name = string(r[:100])
	}
	return name
}

func (s UserTokenSigner) Sign(subject, name string, ttl time.Duration, now time.Time) (string, time.Time, error) {
	if !subjectPattern.MatchString(subject) {
		return "", time.Time{}, &ProviderError{Code: "COMMS_SUBJECT_INVALID"}
	}
	if ttl <= 0 || ttl > UserTokenMaxLifetime {
		ttl = UserTokenMaxLifetime
	}
	iat := now.Unix()
	exp := now.Add(ttl).Unix()
	token, err := signHS256Kid(s.KeyID, userTokenClaims{Issuer: s.Issuer, Audience: s.AppID, Subject: subject, Name: CleanDisplayName(name), IssuedAt: iat, Expires: exp, ID: randomID() + randomID()}, s.Secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, time.Unix(exp, 0).UTC(), nil
}

func ExchangeUserToken(ctx context.Context, client *http.Client, base, token string) error {
	base = strings.TrimRight(base, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/sessions", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	switch {
	case res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK:
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return &ProviderError{Code: "COMMS_USER_TOKEN_REFUSED", Status: res.StatusCode}
	case res.StatusCode == http.StatusTooManyRequests:
		return &ProviderError{Code: "VOICE_PROVIDER_BUSY", Status: res.StatusCode}
	case res.StatusCode == http.StatusServiceUnavailable:
		return &ProviderError{Code: "VOICE_PROVIDER_UNAVAILABLE", Status: res.StatusCode}
	default:
		return &ProviderError{Code: "VOICE_PROVIDER_ERROR", Status: res.StatusCode}
	}
	session := sessionToken(raw)
	if session == "" {
		return &ProviderError{Code: "VOICE_PROVIDER_ERROR", Detail: "no session token"}
	}
	out, err := http.NewRequestWithContext(ctx, http.MethodDelete, base+"/v1/sessions/current", bytes.NewReader(nil))
	if err != nil {
		return nil
	}
	out.Header.Set("Authorization", "Bearer "+session)
	if done, err := client.Do(out); err == nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(done.Body, 1<<16))
		done.Body.Close()
	}
	return nil
}

func sessionToken(raw []byte) string {
	var out struct {
		SessionToken string `json:"session_token"`
	}
	if json.Unmarshal(raw, &out) != nil {
		return ""
	}
	return out.SessionToken
}
