package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type fakePresence struct{ user, name string }

func (p fakePresence) GetHidden() bool                   { return false }
func (p fakePresence) GetPersistence() bool              { return false }
func (p fakePresence) GetUsername() string               { return p.name }
func (p fakePresence) GetStatus() string                 { return "" }
func (p fakePresence) GetReason() runtime.PresenceReason { return 0 }
func (p fakePresence) GetUserId() string                 { return p.user }
func (p fakePresence) GetSessionId() string              { return "s-" + p.user }
func (p fakePresence) GetNodeId() string                 { return "node1" }

type voiceNakama struct {
	*fakeNakama
	matches map[string]*api.Match
	streams map[string][]runtime.Presence
}

func (n *voiceNakama) MatchGet(_ context.Context, id string) (*api.Match, error) {
	return n.matches[id], nil
}

func (n *voiceNakama) StreamUserList(mode uint8, subject, _, label string, _, _ bool) ([]runtime.Presence, error) {
	return n.streams[string(rune('0'+mode))+subject+"."+label], nil
}

func streamKey(mode uint8, id string) string { return string(rune('0'+mode)) + id }

func jwtClaims(t *testing.T, token, secret string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if len(parts) != 3 || base64.RawURLEncoding.EncodeToString(mac.Sum(nil)) != parts[2] {
		t.Fatalf("bad token %q", token)
	}
	raw, _ := base64.RawURLEncoding.DecodeString(parts[1])
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

const lkSecret = "lk-secret-lk-secret-lk-secret-01"

func voiceHarness(t *testing.T) (*harness, *voiceNakama) {
	h := newHarness(t)
	vn := &voiceNakama{fakeNakama: h.s.nk.(*fakeNakama), matches: map[string]*api.Match{}, streams: map[string][]runtime.Presence{}}
	h.s.nk = vn
	h.s.voiceKey = []byte("0123456789abcdef0123456789abcdef")
	h.s.voiceHTTP = &http.Client{Timeout: 2 * time.Second}
	for _, q := range []string{`UPDATE base_voice_settings SET enabled=false, provider='', config='{}', secrets_sealed='', ttl_seconds=600, open_rooms=false, can_publish=true, user_tokens=false, user_token_ttl_seconds=300, user_tokens_per_minute=10 WHERE id=1`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%'`, `DELETE FROM base_restriction`} {
		if _, err := h.s.db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		h.s.db.Exec(`UPDATE base_voice_settings SET enabled=false, provider='', config='{}', secrets_sealed='', user_tokens=false, user_token_ttl_seconds=300, user_tokens_per_minute=10, updated_by='' WHERE id=1`)
		h.s.db.Exec(`DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%'`)
	})
	return h, vn
}

func (h *harness) game(key string, voiceOn bool) {
	h.t.Helper()
	if _, err := h.s.db.Exec(`INSERT INTO economy_game (id, key, voice_enabled) VALUES ($1,$2,$3) ON CONFLICT (key) DO UPDATE SET voice_enabled=EXCLUDED.voice_enabled`, uuid.Must(uuid.NewV4()).String(), key, voiceOn); err != nil {
		h.t.Fatal(err)
	}
	h.t.Cleanup(func() { h.s.db.Exec(`DELETE FROM economy_game WHERE key=$1`, key) })
}

func TestVoiceSettingsSealedAndMasked(t *testing.T) {
	h, _ := voiceHarness(t)
	got := h.mustAPI(200, "GET", "/voice/settings", nil)
	if got["enabled"] != false || got["provider"] != "" || got["canSeal"] != true || len(got["providers"].([]any)) != 2 {
		t.Fatalf("defaults %v", got)
	}
	h.expectError(400, "VOICE_PROVIDER_REQUIRED", "PUT", "/voice/settings", map[string]any{"enabled": true, "ttlSeconds": 600})
	h.expectError(400, "INVALID_VOICE_PROVIDER", "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "jitsi", "ttlSeconds": 600})
	h.expectError(400, "INVALID_URL", "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 600, "config": map[string]string{"url": "ftp://x", "apiKey": "k"}, "secrets": map[string]string{"apiSecret": lkSecret}})
	h.expectError(400, "VOICE_FIELD_REQUIRED", "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 600, "config": map[string]string{"url": "wss://lk.example.com", "apiKey": "k"}})
	h.expectError(400, "INVALID_RULE", "PUT", "/voice/settings", map[string]any{"enabled": false, "ttlSeconds": 10})
	saved := h.mustAPI(200, "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 900, "canPublish": true,
		"config": map[string]string{"url": "wss://lk.example.com", "apiKey": "APIk", "junk": "x"}, "secrets": map[string]string{"apiSecret": lkSecret}})
	if saved["secrets"].(map[string]any)["apiSecret"] != secretMask || saved["config"].(map[string]any)["junk"] != nil {
		t.Fatalf("saved %v", saved)
	}
	var sealed string
	h.s.db.QueryRow(`SELECT secrets_sealed FROM base_voice_settings WHERE id=1`).Scan(&sealed)
	if sealed == "" || strings.Contains(sealed, lkSecret) {
		t.Fatalf("secret not sealed: %q", sealed)
	}
	h.mustAPI(200, "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 900, "config": map[string]string{"url": "wss://lk.example.com", "apiKey": "APIk"}, "secrets": map[string]string{"apiSecret": secretMask}})
	v, err := h.s.loadVoiceSettings(h.ctx, h.s.db)
	if err != nil || v.Secrets["apiSecret"] != lkSecret || v.TTLSeconds != 900 {
		t.Fatalf("mask must keep the secret: %v %+v", err, v)
	}
	h.expectError(400, "VOICE_FIELD_REQUIRED", "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "comms", "ttlSeconds": 900, "config": map[string]string{"url": "https://comms.example.com"}, "secrets": map[string]string{"serverKey": secretMask}})
	h.s.voiceKey = nil
	h.expectError(400, "VOICE_SEAL_UNAVAILABLE", "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 900, "config": map[string]string{"url": "wss://lk.example.com", "apiKey": "APIk"}, "secrets": map[string]string{"apiSecret": "new-secret"}})
}

func TestVoiceTokenPermissions(t *testing.T) {
	h, vn := voiceHarness(t)
	a, b, c := h.user("fa"), h.user("fa"), h.user("fa")
	if out := h.mustRPC(a, "voice_token", map[string]any{"matchId": uuid.Must(uuid.NewV4()).String() + ".node1"}); out["enabled"] != false || out["reason"] != "VOICE_DISABLED" {
		t.Fatalf("off %v", out)
	}
	h.mustAPI(200, "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 600, "canPublish": true,
		"config": map[string]string{"url": "https://lk.example.com", "apiKey": "APIk"}, "secrets": map[string]string{"apiSecret": lkSecret}})
	if code := h.rpcCode(a, "voice_token", map[string]any{}); code != "VOICE_ROOM_REQUIRED" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "voice_token", map[string]any{"matchId": "x.y", "room": "lobby"}); code != "VOICE_ROOM_REQUIRED" {
		t.Fatal(code)
	}

	key := "vg" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	h.game(key, true)
	econ := uuid.Must(uuid.NewV4()).String() + ".node1"
	h.s.db.Exec(`INSERT INTO economy_game_match (id, game_key, status) VALUES ($1,$2,'started')`, econ, key)
	t.Cleanup(func() { h.s.db.Exec(`DELETE FROM economy_game_match WHERE id=$1`, econ) })
	h.s.db.Exec(`INSERT INTO economy_game_entry (match_id, user_id, status) VALUES ($1,$2,'paid'), ($1,$3,'free')`, econ, a, b)
	out := h.mustRPC(a, "voice_token", map[string]any{"matchId": econ, "gameKey": "someother"})
	if out["enabled"] != true || out["provider"] != "livekit" || out["url"] != "wss://lk.example.com" || out["room"] != econ || out["identity"] != a || out["gameKey"] != key || out["canPublish"] != true {
		t.Fatalf("token %v", out)
	}
	claims := jwtClaims(t, out["token"].(string), lkSecret)
	if claims["sub"] != a || claims["iss"] != "APIk" || claims["video"].(map[string]any)["room"] != econ {
		t.Fatalf("claims %v", claims)
	}
	exp, _ := time.Parse(time.RFC3339, out["expiresAt"].(string))
	if d := time.Until(exp); d < 9*time.Minute || d > 11*time.Minute {
		t.Fatalf("expires in %v", d)
	}
	if code := h.rpcCode(c, "voice_token", map[string]any{"matchId": econ}); code != "NOT_IN_MATCH" {
		t.Fatalf("outsider: %s", code)
	}
	h.s.db.Exec(`UPDATE economy_game_entry SET status='left' WHERE match_id=$1 AND user_id=$2`, econ, b)
	if code := h.rpcCode(b, "voice_token", map[string]any{"matchId": econ}); code != "NOT_IN_MATCH" {
		t.Fatalf("a player who left: %s", code)
	}
	h.s.db.Exec(`UPDATE economy_game SET voice_enabled=false WHERE key=$1`, key)
	if out := h.mustRPC(a, "voice_token", map[string]any{"matchId": econ}); out["enabled"] != false || out["reason"] != "GAME_VOICE_DISABLED" {
		t.Fatalf("game off %v", out)
	}
	h.s.db.Exec(`UPDATE economy_game SET voice_enabled=true WHERE key=$1`, key)

	live := uuid.Must(uuid.NewV4()).String()
	vn.matches[live+".node1"] = &api.Match{MatchId: live + ".node1", Authoritative: true, HandlerName: key}
	vn.streams[streamKey(streamModeMatchAuthoritative, live+".node1")] = []runtime.Presence{fakePresence{c, "cc"}, fakePresence{b, "bb"}}
	if out := h.mustRPC(c, "voice_token", map[string]any{"matchId": live + ".node1"}); out["gameKey"] != key || out["room"] != live+".node1" {
		t.Fatalf("stream match %v", out)
	}
	if code := h.rpcCode(a, "voice_token", map[string]any{"matchId": live + ".node1"}); code != "NOT_IN_MATCH" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "voice_token", map[string]any{"matchId": uuid.Must(uuid.NewV4()).String() + ".node1"}); code != "MATCH_NOT_FOUND" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "voice_token", map[string]any{"matchId": "not-a-match"}); code != "MATCH_NOT_FOUND" {
		t.Fatal(code)
	}

	party := uuid.Must(uuid.NewV4()).String() + ".node1"
	vn.streams[streamKey(streamModeParty, party)] = []runtime.Presence{fakePresence{a, "aa"}}
	if out := h.mustRPC(a, "voice_token", map[string]any{"partyId": party}); out["room"] != "party."+party {
		t.Fatalf("party %v", out)
	}
	if code := h.rpcCode(b, "voice_token", map[string]any{"partyId": party}); code != "NOT_IN_PARTY" {
		t.Fatal(code)
	}

	if code := h.rpcCode(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key}); code != "ROOM_NOT_ALLOWED" {
		t.Fatal(code)
	}
	h.s.db.Exec(`UPDATE base_voice_settings SET open_rooms=true WHERE id=1`)
	if code := h.rpcCode(a, "voice_token", map[string]any{"room": "lobby"}); code != "GAME_KEY_REQUIRED" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "voice_token", map[string]any{"room": "Lobby Room", "gameKey": key}); code != "INVALID_ROOM" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "voice_token", map[string]any{"room": "lobby", "gameKey": "nosuchgame"}); code != "GAME_NOT_FOUND" {
		t.Fatal(code)
	}
	if out := h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key}); out["room"] != "room."+key+".lobby" {
		t.Fatalf("open room %v", out)
	}

	h.mustAPI(201, "POST", "/players/"+a+"/restrictions", map[string]any{"kind": "mute", "minutes": 30, "reason": map[string]string{"fa": "سکوت"}})
	out = h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key})
	if out["canPublish"] != false || jwtClaims(t, out["token"].(string), lkSecret)["video"].(map[string]any)["canPublish"] != false {
		t.Fatalf("a muted player listens only: %v", out)
	}

	h.s.db.Exec(`DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%'`)
	limited := ""
	for i := 0; i < voiceTokenLimit+1; i++ {
		if _, err := h.rpc(c, "voice_token", map[string]any{"room": "lobby", "gameKey": key}); err != nil {
			if app, ok := asApp(err); ok {
				limited = app.Code
			}
		}
	}
	if limited != "VOICE_RATE_LIMITED" {
		t.Fatalf("rate limit %q", limited)
	}
}

func TestVoiceTestAndCommsToken(t *testing.T) {
	h, _ := voiceHarness(t)
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.URL.Path == "/healthz" {
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer ck_good" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"unauthorized"}}`))
			return
		}
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/token"):
			_, _ = w.Write([]byte(`{"token":"comms-join","expires_at":"2030-01-01T00:00:00Z"}`))
		default:
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer srv.Close()
	h.expectError(400, "VOICE_PROVIDER_REQUIRED", "POST", "/voice/test", map[string]any{})
	res := h.mustAPI(200, "POST", "/voice/test", map[string]any{"provider": "comms", "ttlSeconds": 600, "config": map[string]string{"url": srv.URL}, "secrets": map[string]string{"serverKey": "ck_bad"}})
	if res["ok"] != false || res["code"] != "VOICE_BAD_CREDENTIALS" {
		t.Fatalf("bad key test %v", res)
	}
	h.mustAPI(200, "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "comms", "ttlSeconds": 600, "openRooms": true, "canPublish": true,
		"config": map[string]string{"url": srv.URL, "publicUrl": "https://comms.example.com"}, "secrets": map[string]string{"serverKey": "ck_good"}})
	res = h.mustAPI(200, "POST", "/voice/test", map[string]any{})
	if res["ok"] != true || res["provider"] != "comms" {
		t.Fatalf("saved test %v", res)
	}
	res = h.mustAPI(200, "POST", "/voice/test", map[string]any{"provider": "comms", "ttlSeconds": 600, "config": map[string]string{"url": srv.URL}, "secrets": map[string]string{"serverKey": secretMask}})
	if res["ok"] != true {
		t.Fatalf("masked secret test %v", res)
	}
	key := "vc" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	h.game(key, true)
	a := h.user("fa")
	out := h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key})
	if out["provider"] != "comms" || out["token"] != "comms-join" || out["url"] != "https://comms.example.com" || out["expiresAt"] != "2030-01-01T00:00:00Z" {
		t.Fatalf("comms token %v", out)
	}
	for i := 0; i < voiceTestLimit; i++ {
		h.api("POST", "/voice/test", map[string]any{})
	}
	h.expectError(429, "TOO_MANY_TESTS", "POST", "/voice/test", map[string]any{})
}

func TestVoicePolicies(t *testing.T) {
	h, vn := voiceHarness(t)
	t.Cleanup(func() {
		h.s.db.Exec(`UPDATE base_voice_settings SET allowed_tag_ids='{}', listen_only_tag_ids='{}', tokens_per_minute=30, max_room_players=0 WHERE id=1`)
	})
	suffix := strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	a, b, c := h.user("fa"), h.user("fa"), h.user("fa")
	vipTag := h.userTag("vvip"+suffix, a, b)
	quiet := h.userTag("vquiet"+suffix, b)
	base := map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 600, "canPublish": true, "openRooms": true,
		"config": map[string]string{"url": "wss://lk.example.com", "apiKey": "APIk"}, "secrets": map[string]string{"apiSecret": lkSecret}}
	with := func(extra map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	h.expectError(400, "INVALID_RULE", "PUT", "/voice/settings", with(map[string]any{"tokensPerMinute": 0}))
	h.expectError(400, "INVALID_RULE", "PUT", "/voice/settings", with(map[string]any{"maxRoomPlayers": 5000}))
	saved := h.mustAPI(200, "PUT", "/voice/settings", with(map[string]any{"allowedTagIds": []string{vipTag}, "listenOnlyTagIds": []string{quiet}, "tokensPerMinute": 2, "maxRoomPlayers": 2}))
	if len(saved["allowedTagIds"].([]any)) != 1 || saved["tokensPerMinute"].(float64) != 2 || saved["maxRoomPlayers"].(float64) != 2 {
		t.Fatalf("saved %v", saved)
	}
	key := "vp" + suffix
	h.game(key, true)
	if out := h.mustRPC(c, "voice_token", map[string]any{"room": "lobby", "gameKey": key}); out["enabled"] != false || out["reason"] != "VOICE_NOT_FOR_PLAYER" {
		t.Fatalf("not allowed %v", out)
	}
	if out := h.mustRPC(b, "voice_token", map[string]any{"room": "lobby", "gameKey": key}); out["canPublish"] != false {
		t.Fatalf("listen only %v", out)
	}
	h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key})
	h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key})
	if code := h.rpcCode(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key}); code != "VOICE_RATE_LIMITED" {
		t.Fatalf("per minute from settings: %s", code)
	}
	h.s.db.Exec(`DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%'`)
	h.s.db.Exec(`UPDATE base_voice_settings SET allowed_tag_ids='{}', tokens_per_minute=30 WHERE id=1`)

	match := uuid.Must(uuid.NewV4()).String() + ".node1"
	h.s.db.Exec(`INSERT INTO economy_game_match (id, game_key, status) VALUES ($1,$2,'started')`, match, key)
	t.Cleanup(func() { h.s.db.Exec(`DELETE FROM economy_game_match WHERE id=$1`, match) })
	h.s.db.Exec(`INSERT INTO economy_game_entry (match_id, user_id, status, created_at) VALUES ($1,$2,'paid',now()-interval '3 minutes'), ($1,$3,'paid',now()-interval '2 minutes'), ($1,$4,'paid',now()-interval '1 minute')`, match, a, b, c)
	h.mustRPC(a, "voice_token", map[string]any{"matchId": match})
	if code := h.rpcCode(c, "voice_token", map[string]any{"matchId": match}); code != "VOICE_ROOM_FULL" {
		t.Fatalf("third player: %s", code)
	}
	h.s.db.Exec(`UPDATE economy_game SET voice_max_players=0 WHERE key=$1`, key)
	h.mustRPC(c, "voice_token", map[string]any{"matchId": match})
	h.s.db.Exec(`UPDATE economy_game SET voice_max_players=1 WHERE key=$1`, key)
	if code := h.rpcCode(b, "voice_token", map[string]any{"matchId": match}); code != "VOICE_ROOM_FULL" {
		t.Fatalf("per-game size: %s", code)
	}
	_ = vn
}
