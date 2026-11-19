package voice

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testAppID  = "0b5c2a4e-6f0d-4e5b-9a43-1d2c3b4a5f60"
	testSecret = "comms-hmac-secret-0123456789abcdef"
)

func userCfg() Config {
	return Config{Values: map[string]string{"url": "http://comms.local", "appId": testAppID, "tokenIssuer": "nakama-farsi", "tokenKeyId": "game-k1"}, Secrets: map[string]string{"serverKey": "ck_x", "tokenSecret": testSecret}}
}

func TestUserTokenHeaderClaimsAndSignature(t *testing.T) {
	s, err := NewUserTokenSigner(userCfg())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_800_000_000, 0)
	token, exp, err := s.Sign("8a7b6c5d-0000-4000-8000-000000000001", "  Ali\x07 علی  ", 300*time.Second, now)
	if err != nil {
		t.Fatal(err)
	}
	head, claims := decodeJWT(t, token, testSecret)
	if head["alg"] != "HS256" || head["typ"] != "JWT" || head["kid"] != "game-k1" || len(head) != 3 {
		t.Fatalf("header %v", head)
	}
	if claims["aud"] != testAppID || claims["iss"] != "nakama-farsi" || claims["sub"] != "8a7b6c5d-0000-4000-8000-000000000001" || claims["name"] != "Ali علی" {
		t.Fatalf("claims %v", claims)
	}
	if claims["iat"].(float64) != float64(now.Unix()) || claims["exp"].(float64) != float64(now.Unix()+300) || !exp.Equal(time.Unix(now.Unix()+300, 0)) {
		t.Fatalf("times %v %v", claims, exp)
	}
	if jti, _ := claims["jti"].(string); len(jti) != 32 {
		t.Fatalf("jti %v", claims["jti"])
	}
	parts := strings.Split(token, ".")
	mac := hmac.New(sha256.New, []byte("another-secret-another-secret-0000"))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if base64.RawURLEncoding.EncodeToString(mac.Sum(nil)) == parts[2] {
		t.Fatal("a different secret verified the token")
	}
	token2, _, _ := s.Sign("8a7b6c5d-0000-4000-8000-000000000001", "", time.Hour, now)
	_, c2 := decodeJWT(t, token2, testSecret)
	if c2["exp"].(float64)-c2["iat"].(float64) != 600 {
		t.Fatalf("lifetime must be capped at 10 minutes: %v", c2)
	}
	if _, ok := c2["name"]; ok {
		t.Fatal("empty name must be left out")
	}
	if _, _, err := s.Sign("bad subject", "", time.Minute, now); err == nil {
		t.Fatal("bad subject accepted")
	}
	long := strings.Repeat("ن", 150)
	if got := CleanDisplayName(long); len([]rune(got)) != 100 {
		t.Fatalf("name not clipped: %d", len([]rune(got)))
	}
}

func TestUserTokenFieldChecks(t *testing.T) {
	cases := []struct {
		field, value, code string
	}{
		{"appId", "not-a-uuid", "COMMS_APP_ID_INVALID"},
		{"tokenKeyId", "has space", "COMMS_KEY_ID_INVALID"},
		{"tokenSecret", "short", "COMMS_SECRET_TOO_SHORT"},
		{"tokenSecret", strings.Repeat("x", 513), "COMMS_SECRET_TOO_LONG"},
		{"tokenIssuer", strings.Repeat("i", 256), "COMMS_ISSUER_INVALID"},
	}
	for _, c := range cases {
		cfg := userCfg()
		if c.field == "tokenSecret" {
			cfg.Secrets[c.field] = c.value
		} else {
			cfg.Values[c.field] = c.value
		}
		var fe *FieldError
		if err := Validate(KindComms, cfg); !errors.As(err, &fe) || fe.Code != c.code || fe.Field != c.field {
			t.Fatalf("%s=%q: %v", c.field, c.value, err)
		}
	}
	if err := Validate(KindComms, userCfg()); err != nil {
		t.Fatal(err)
	}
	cfg := userCfg()
	delete(cfg.Values, "tokenKeyId")
	if err := Validate(KindComms, cfg); err != nil {
		t.Fatalf("user token fields are optional for voice: %v", err)
	}
	var fe *FieldError
	if _, err := NewUserTokenSigner(cfg); !errors.As(err, &fe) || fe.Code != "VOICE_FIELD_REQUIRED" || fe.Field != "tokenKeyId" {
		t.Fatalf("missing kid: %v", err)
	}
	if UserTokenReady(cfg) || !UserTokenReady(userCfg()) {
		t.Fatal("ready")
	}
}

func TestCommsTestExchangesUserToken(t *testing.T) {
	var deleted bool
	secret := testSecret
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/healthz":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/sessions":
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			parts := strings.Split(token, ".")
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(parts[0] + "." + parts[1]))
			if len(parts) != 3 || base64.RawURLEncoding.EncodeToString(mac.Sum(nil)) != parts[2] {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":{"code":"unauthorized"}}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"session_token":"sess-1"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/sessions/current":
			deleted = r.Header.Get("Authorization") == "Bearer sess-1"
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()
	cfg := userCfg()
	cfg.Values["url"] = srv.URL
	p := newProvider(t, KindComms, cfg.Values, cfg.Secrets, srv.Client())
	if err := p.Test(context.Background()); err != nil || !deleted {
		t.Fatalf("test %v deleted=%v", err, deleted)
	}
	secret = "the-comms-side-has-another-secret-000"
	var pe *ProviderError
	if err := p.Test(context.Background()); !errors.As(err, &pe) || pe.Code != "COMMS_USER_TOKEN_REFUSED" {
		t.Fatalf("refused %v", err)
	}
	delete(cfg.Secrets, "tokenSecret")
	if err := newProvider(t, KindComms, cfg.Values, cfg.Secrets, srv.Client()).Test(context.Background()); err != nil {
		t.Fatalf("without user token fields only the room check runs: %v", err)
	}
}
