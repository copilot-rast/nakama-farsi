package voice

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func decodeJWT(t *testing.T, token, secret string) (map[string]any, map[string]any) {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWT: %q", token)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if base64.RawURLEncoding.EncodeToString(mac.Sum(nil)) != parts[2] {
		t.Fatal("bad HS256 signature")
	}
	var head, claims map[string]any
	for i, out := range []*map[string]any{&head, &claims} {
		raw, err := base64.RawURLEncoding.DecodeString(parts[i])
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatal(err)
		}
	}
	return head, claims
}

func newProvider(t *testing.T, kind string, values, secrets map[string]string, client *http.Client) Provider {
	t.Helper()
	spec, ok := Lookup(kind)
	if !ok {
		t.Fatalf("%s not registered", kind)
	}
	if client == nil {
		client = http.DefaultClient
	}
	p, err := spec.New(Config{Values: values, Secrets: secrets}, client)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLiveKitTokenClaims(t *testing.T) {
	p := newProvider(t, KindLiveKit, map[string]string{"url": "https://lk.example.com/", "apiKey": "APIkey1"}, map[string]string{"apiSecret": "s3cr3t-s3cr3t-s3cr3t"}, nil)
	lk := p.(*liveKit)
	fixed := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	lk.now = func() time.Time { return fixed }
	g, err := p.Token(context.Background(), Request{Room: "match.abc.node1", Identity: "user-1", Name: "سارا", GameKey: "stargame", TTL: 15 * time.Minute, CanPublish: true})
	if err != nil {
		t.Fatal(err)
	}
	if g.Provider != "livekit" || g.URL != "wss://lk.example.com" || g.Room != "match.abc.node1" || g.Identity != "user-1" || !g.ExpiresAt.Equal(fixed.Add(15*time.Minute)) {
		t.Fatalf("grant %+v", g)
	}
	head, claims := decodeJWT(t, g.Token, "s3cr3t-s3cr3t-s3cr3t")
	if head["alg"] != "HS256" || head["typ"] != "JWT" {
		t.Fatalf("head %v", head)
	}
	if claims["iss"] != "APIkey1" || claims["sub"] != "user-1" || claims["name"] != "سارا" || claims["exp"].(float64) != float64(fixed.Add(15*time.Minute).Unix()) || claims["nbf"].(float64) > float64(fixed.Unix()) || claims["jti"] == "" {
		t.Fatalf("claims %v", claims)
	}
	video := claims["video"].(map[string]any)
	if video["roomJoin"] != true || video["room"] != "match.abc.node1" || video["canPublish"] != true || video["canSubscribe"] != true || video["canPublishData"] != false {
		t.Fatalf("video grant %v", video)
	}
	if !strings.Contains(claims["metadata"].(string), "stargame") {
		t.Fatalf("metadata %v", claims["metadata"])
	}
	g, _ = p.Token(context.Background(), Request{Room: "r", Identity: "u", CanPublish: false})
	_, claims = decodeJWT(t, g.Token, "s3cr3t-s3cr3t-s3cr3t")
	if v := claims["video"].(map[string]any); v["canPublish"] != false {
		t.Fatalf("listen-only grant %v", v)
	}
	if exp := claims["exp"].(float64) - claims["iat"].(float64); exp != 600 {
		t.Fatalf("default ttl %v", exp)
	}
}

func TestLiveKitTestConnection(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/twirp/livekit.RoomService/ListRooms" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		auth = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		parts := strings.Split(auth, ".")
		mac := hmac.New(sha256.New, []byte("right-secret-right-secret"))
		mac.Write([]byte(parts[0] + "." + parts[1]))
		if len(parts) != 3 || base64.RawURLEncoding.EncodeToString(mac.Sum(nil)) != parts[2] {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"rooms":[]}`))
	}))
	defer srv.Close()
	ws := strings.Replace(srv.URL, "http://", "ws://", 1)
	ok := newProvider(t, KindLiveKit, map[string]string{"url": ws, "apiKey": "k"}, map[string]string{"apiSecret": "right-secret-right-secret"}, srv.Client())
	if err := ok.Test(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, claims := decodeJWT(t, auth, "right-secret-right-secret")
	if claims["video"].(map[string]any)["roomList"] != true {
		t.Fatalf("test token needs roomList: %v", claims)
	}
	bad := newProvider(t, KindLiveKit, map[string]string{"url": ws, "apiKey": "k"}, map[string]string{"apiSecret": "wrong"}, srv.Client())
	var pe *ProviderError
	if err := bad.Test(context.Background()); !errors.As(err, &pe) || pe.Code != "VOICE_BAD_CREDENTIALS" {
		t.Fatalf("bad secret: %v", err)
	}
	down := newProvider(t, KindLiveKit, map[string]string{"url": "http://127.0.0.1:1", "apiKey": "k"}, map[string]string{"apiSecret": "x"}, &http.Client{Timeout: time.Second})
	if err := down.Test(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unreachable: %v", err)
	}
}

type fakeComms struct {
	mu      sync.Mutex
	rooms   map[string]map[string]bool
	closed  map[string]bool
	calls   []string
	key     string
	opened  []map[string]any
	healthy bool
}

func (f *fakeComms) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.calls = append(f.calls, r.Method+" "+r.URL.Path)
		if r.URL.Path == "/healthz" {
			if !f.healthy {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+f.key {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"unauthorized","message":"bad key"}}`))
			return
		}
		fail := func(status int, code string) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":{"code":"` + code + `","message":"x"}}`))
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/server/game-rooms"), "/")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/server/game-rooms":
			var body map[string]any
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			f.opened = append(f.opened, body)
			room := body["match_id"].(string)
			f.rooms[room] = map[string]bool{}
			delete(f.closed, room)
			for _, p := range body["players"].([]any) {
				f.rooms[room][p.(map[string]any)["id"].(string)] = true
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodGet && len(parts) == 2:
			fail(http.StatusNotFound, "not_found")
		case r.Method == http.MethodPut && len(parts) == 4 && parts[2] == "players":
			if f.rooms[parts[1]] == nil {
				fail(http.StatusNotFound, "not_found")
				return
			}
			f.rooms[parts[1]][parts[3]] = true
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodPost && len(parts) == 5 && parts[4] == "token":
			room, player := parts[1], parts[3]
			switch {
			case f.closed[room]:
				fail(http.StatusConflict, "room_closed")
			case f.rooms[room] == nil:
				fail(http.StatusNotFound, "not_found")
			case !f.rooms[room][player]:
				fail(http.StatusConflict, "not_listed")
			default:
				_, _ = w.Write([]byte(`{"token":"join-` + room + `-` + player + `","expires_at":"2026-11-14T12:10:00Z"}`))
			}
		default:
			http.NotFound(w, r)
		}
	})
}

func TestCommsTokenOpensRoomsAndListsPlayers(t *testing.T) {
	f := &fakeComms{rooms: map[string]map[string]bool{}, closed: map[string]bool{}, key: "ck_live", healthy: true}
	srv := httptest.NewServer(f.handler())
	defer srv.Close()
	p := newProvider(t, KindComms, map[string]string{"url": srv.URL + "/", "publicUrl": "https://comms.example.com"}, map[string]string{"serverKey": "ck_live"}, srv.Client())

	req := Request{Room: "m1.node1", Identity: "u1", Name: "Ali", Players: []Player{{ID: "u1", Name: "Ali"}, {ID: "u2", Name: "Sara"}}, TTL: 20 * time.Minute}
	g, err := p.Token(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if g.Provider != "comms" || g.Token != "join-m1.node1-u1" || g.URL != "https://comms.example.com" || g.Room != "m1.node1" || g.Identity != "u1" || !g.ExpiresAt.Equal(time.Date(2026, 11, 14, 12, 10, 0, 0, time.UTC)) {
		t.Fatalf("grant %+v", g)
	}
	if len(f.opened) != 1 || f.opened[0]["video"] != false || f.opened[0]["ttl_seconds"].(float64) != 1200 || len(f.opened[0]["players"].([]any)) != 2 {
		t.Fatalf("open body %v", f.opened)
	}
	if g, err = p.Token(context.Background(), Request{Room: "m1.node1", Identity: "u2"}); err != nil || g.Token != "join-m1.node1-u2" || len(f.opened) != 1 {
		t.Fatalf("second player: %v %+v", err, g)
	}
	if g, err = p.Token(context.Background(), Request{Room: "m1.node1", Identity: "u3", Name: "Late"}); err != nil || g.Token != "join-m1.node1-u3" {
		t.Fatalf("not listed player: %v", err)
	}
	if f.calls[len(f.calls)-2] != "PUT /v1/server/game-rooms/m1.node1/players/u3" {
		t.Fatalf("calls %v", f.calls)
	}
	f.closed["m1.node1"] = true
	if _, err = p.Token(context.Background(), Request{Room: "m1.node1", Identity: "u1"}); err != nil || len(f.opened) != 2 {
		t.Fatalf("closed room reopened: %v %d", err, len(f.opened))
	}
	if err := p.Test(context.Background()); err != nil {
		t.Fatal(err)
	}

	bad := newProvider(t, KindComms, map[string]string{"url": srv.URL}, map[string]string{"serverKey": "ck_wrong"}, srv.Client())
	var pe *ProviderError
	if _, err := bad.Token(context.Background(), req); !errors.As(err, &pe) || pe.Code != "VOICE_BAD_CREDENTIALS" {
		t.Fatalf("bad key token: %v", err)
	}
	if err := bad.Test(context.Background()); !errors.As(err, &pe) || pe.Code != "VOICE_BAD_CREDENTIALS" {
		t.Fatalf("bad key test: %v", err)
	}
	f.healthy = false
	if err := p.Test(context.Background()); !errors.As(err, &pe) || pe.Code != "VOICE_PROVIDER_UNAVAILABLE" {
		t.Fatalf("unhealthy: %v", err)
	}
}

func TestValidateAndRegistry(t *testing.T) {
	kinds := []string{}
	for _, s := range Kinds() {
		kinds = append(kinds, s.Kind)
	}
	if strings.Join(kinds, ",") != "comms,livekit" {
		t.Fatalf("kinds %v", kinds)
	}
	var fe *FieldError
	if err := Validate("nope", Config{}); !errors.As(err, &fe) || fe.Code != "INVALID_VOICE_PROVIDER" {
		t.Fatal(err)
	}
	if err := Validate(KindLiveKit, Config{Values: map[string]string{"url": "wss://x", "apiKey": "k"}}); !errors.As(err, &fe) || fe.Field != "apiSecret" || fe.Code != "VOICE_FIELD_REQUIRED" {
		t.Fatal(err)
	}
	if err := Validate(KindLiveKit, Config{Values: map[string]string{"url": "ftp://x", "apiKey": "k"}, Secrets: map[string]string{"apiSecret": "s"}}); !errors.As(err, &fe) || fe.Field != "url" {
		t.Fatal(err)
	}
	if err := Validate(KindComms, Config{Values: map[string]string{"url": "wss://x"}, Secrets: map[string]string{"serverKey": "s"}}); !errors.As(err, &fe) || fe.Field != "url" {
		t.Fatal(err)
	}
	if err := Validate(KindComms, Config{Values: map[string]string{"url": "https://comms.example.com"}, Secrets: map[string]string{"serverKey": "s"}}); err != nil {
		t.Fatal(err)
	}
	for room, ok := range map[string]bool{"abc.node1": true, "room:stargame:lobby": true, "": false, "a b": false, strings.Repeat("x", 129): false} {
		if ValidRoom(room) != ok {
			t.Fatalf("room %q", room)
		}
	}
}
