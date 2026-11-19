package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	utAppID  = "5f1c9a0e-2b3d-4c5e-8f70-112233445566"
	utSecret = "comms-user-token-secret-0123456789abcdef"
)

func utSettings(extra map[string]any) map[string]any {
	out := map[string]any{"enabled": false, "provider": "comms", "ttlSeconds": 600, "userTokens": true, "userTokenTtlSeconds": 300, "userTokensPerMinute": 10,
		"config":  map[string]string{"url": "http://127.0.0.1:1", "publicUrl": "https://comms.example.com", "appId": utAppID, "tokenIssuer": "nakama-farsi", "tokenKeyId": "game-k1"},
		"secrets": map[string]string{"serverKey": "ck_good", "tokenSecret": utSecret}}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func TestCommsUserTokenSettingsSealed(t *testing.T) {
	h, _ := voiceHarness(t)
	got := h.mustAPI(200, "GET", "/voice/settings", nil)
	if got["userTokens"] != false || got["userTokenTtlSeconds"].(float64) != 300 || got["userTokensPerMinute"].(float64) != 10 || got["userTokenReady"] != false {
		t.Fatalf("defaults %v", got)
	}
	h.expectError(400, "COMMS_PROVIDER_REQUIRED", "PUT", "/voice/settings", map[string]any{"enabled": false, "ttlSeconds": 600, "userTokens": true})
	h.expectError(400, "COMMS_PROVIDER_REQUIRED", "PUT", "/voice/settings", map[string]any{"enabled": false, "provider": "livekit", "ttlSeconds": 600, "userTokens": true,
		"config": map[string]string{"url": "wss://lk.example.com", "apiKey": "k"}, "secrets": map[string]string{"apiSecret": lkSecret}})
	h.expectError(400, "INVALID_RULE", "PUT", "/voice/settings", utSettings(map[string]any{"userTokenTtlSeconds": 601}))
	h.expectError(400, "INVALID_RULE", "PUT", "/voice/settings", utSettings(map[string]any{"userTokenTtlSeconds": 30}))
	h.expectError(400, "INVALID_RULE", "PUT", "/voice/settings", utSettings(map[string]any{"userTokensPerMinute": 0}))
	h.expectError(400, "COMMS_SECRET_TOO_SHORT", "PUT", "/voice/settings", utSettings(map[string]any{"secrets": map[string]string{"serverKey": "ck_good", "tokenSecret": "short"}}))
	h.expectError(400, "COMMS_APP_ID_INVALID", "PUT", "/voice/settings", utSettings(map[string]any{"config": map[string]string{"url": "http://127.0.0.1:1", "appId": "app-1", "tokenIssuer": "i", "tokenKeyId": "k"}}))
	code, body := h.api("PUT", "/voice/settings", utSettings(map[string]any{"config": map[string]string{"url": "http://127.0.0.1:1", "appId": utAppID, "tokenIssuer": "nakama-farsi"}}))
	if code != 400 || !strings.Contains(jsonText(body), "VOICE_FIELD_REQUIRED") || !strings.Contains(jsonText(body), "tokenKeyId") {
		t.Fatalf("missing kid %d %v", code, body)
	}
	h.mustAPI(200, "PUT", "/voice/settings", utSettings(map[string]any{"userTokens": false, "config": map[string]string{"url": "http://127.0.0.1:1", "appId": utAppID}}))
	saved := h.mustAPI(200, "PUT", "/voice/settings", utSettings(nil))
	if saved["secrets"].(map[string]any)["tokenSecret"] != secretMask || saved["secrets"].(map[string]any)["serverKey"] != secretMask || saved["userTokens"] != true || saved["userTokenReady"] != true {
		t.Fatalf("saved %v", saved)
	}
	if saved["config"].(map[string]any)["tokenKeyId"] != "game-k1" || saved["config"].(map[string]any)["appId"] != utAppID {
		t.Fatalf("config %v", saved["config"])
	}
	var sealed, cfg string
	h.s.db.QueryRow(`SELECT secrets_sealed, config::text FROM base_voice_settings WHERE id=1`).Scan(&sealed, &cfg)
	if sealed == "" || strings.Contains(sealed, utSecret) || strings.Contains(cfg, utSecret) {
		t.Fatalf("secret not sealed: %q %q", sealed, cfg)
	}
	h.mustAPI(200, "PUT", "/voice/settings", utSettings(map[string]any{"userTokenTtlSeconds": 120, "secrets": map[string]string{"serverKey": secretMask, "tokenSecret": secretMask}}))
	v, err := h.s.loadVoiceSettings(h.ctx, h.s.db)
	if err != nil || v.Secrets["tokenSecret"] != utSecret || v.UserTTL != 120 || !v.UserTokens {
		t.Fatalf("mask must keep the secret: %v %+v", err, v)
	}
	key := h.s.voiceKey
	h.s.voiceKey = nil
	h.expectError(400, "VOICE_SEAL_UNAVAILABLE", "PUT", "/voice/settings", utSettings(nil))
	h.s.voiceKey = key
}

func TestCommsUserTokenRPC(t *testing.T) {
	h, _ := voiceHarness(t)
	a := h.user("fa")
	h.s.db.Exec(`UPDATE users SET display_name=$2 WHERE id=$1`, a, "بازیکن یک")
	if out := h.mustRPC(a, "comms_user_token", nil); out["enabled"] != false || out["reason"] != "COMMS_TOKENS_DISABLED" {
		t.Fatalf("off %v", out)
	}
	h.mustAPI(200, "PUT", "/voice/settings", utSettings(map[string]any{"userTokensPerMinute": 3}))
	start := time.Now().Unix()
	out := h.mustRPC(a, "comms_user_token", nil)
	if out["enabled"] != true || out["url"] != "https://comms.example.com" || out["appId"] != utAppID {
		t.Fatalf("token %v", out)
	}
	token := out["token"].(string)
	claims := jwtClaims(t, token, utSecret)
	rawHead, _ := base64.RawURLEncoding.DecodeString(strings.Split(token, ".")[0])
	head := map[string]any{}
	_ = json.Unmarshal(rawHead, &head)
	if head["kid"] != "game-k1" || head["alg"] != "HS256" {
		t.Fatalf("head %v", head)
	}
	if claims["sub"] != a || claims["aud"] != utAppID || claims["iss"] != "nakama-farsi" || claims["name"] != "بازیکن یک" || claims["jti"] == "" {
		t.Fatalf("claims %v", claims)
	}
	iat, exp := int64(claims["iat"].(float64)), int64(claims["exp"].(float64))
	if exp-iat != 300 || iat < start-1 || iat > time.Now().Unix()+1 {
		t.Fatalf("times %v", claims)
	}
	if at, _ := time.Parse(time.RFC3339, out["expiresAt"].(string)); at.Unix() != exp {
		t.Fatalf("expiresAt %v vs %d", out["expiresAt"], exp)
	}
	b := h.user("fa")
	if c := jwtClaims(t, h.mustRPC(b, "comms_user_token", map[string]any{})["token"].(string), utSecret); c["sub"] != b || !strings.HasPrefix(c["name"].(string), "b") {
		t.Fatalf("username as name %v", c)
	}
	h.mustRPC(a, "comms_user_token", nil)
	h.mustRPC(a, "comms_user_token", nil)
	if code := h.rpcCode(a, "comms_user_token", nil); code != "COMMS_TOKEN_RATE_LIMITED" {
		t.Fatalf("rate limit %q", code)
	}
	h.s.db.Exec(`UPDATE base_voice_settings SET provider='livekit' WHERE id=1`)
	if out := h.mustRPC(b, "comms_user_token", nil); out["enabled"] != false || out["reason"] != "COMMS_NOT_CONFIGURED" {
		t.Fatalf("not comms %v", out)
	}
	h.s.db.Exec(`UPDATE base_voice_settings SET provider='comms', config=config - 'tokenKeyId' WHERE id=1`)
	if out := h.mustRPC(b, "comms_user_token", nil); out["enabled"] != false || out["reason"] != "COMMS_NOT_CONFIGURED" {
		t.Fatalf("incomplete %v", out)
	}
	if code := h.rpcCode(b, "comms_user_token", "{not json"); code != "INVALID_JSON" {
		t.Fatalf("bad json %q", code)
	}
}

func TestCommsUserTokenConnectionTest(t *testing.T) {
	h, _ := voiceHarness(t)
	var sessions int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/healthz":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case r.Header.Get("Authorization") == "Bearer ck_good" && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/sessions":
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			parts := strings.Split(token, ".")
			if len(parts) != 3 || !strings.Contains(string(mustDecode(parts[1])), `"aud":"`+utAppID+`"`) || !validHS(parts, utSecret) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			sessions++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"session_token":"s1"}`))
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()
	cfg := map[string]string{"url": srv.URL, "appId": utAppID, "tokenIssuer": "nakama-farsi", "tokenKeyId": "game-k1"}
	res := h.mustAPI(200, "POST", "/voice/test", utSettings(map[string]any{"config": cfg}))
	if res["ok"] != true || res["userToken"] != true || sessions != 1 {
		t.Fatalf("good %v %d", res, sessions)
	}
	res = h.mustAPI(200, "POST", "/voice/test", utSettings(map[string]any{"config": cfg, "secrets": map[string]string{"serverKey": "ck_good", "tokenSecret": "a-wrong-secret-a-wrong-secret-0000"}}))
	if res["ok"] != false || res["code"] != "COMMS_USER_TOKEN_REFUSED" {
		t.Fatalf("refused %v", res)
	}
}

func mustDecode(part string) []byte {
	raw, _ := base64.RawURLEncoding.DecodeString(part)
	return raw
}

func validHS(parts []string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)) == parts[2]
}

func jsonText(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}
