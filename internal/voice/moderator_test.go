package voice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLiveKitModeratorUpdatesPermissionWithRoomAdminToken(t *testing.T) {
	var seen []map[string]any
	var claims map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims = decodeJWT(t, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), "lk-secret")
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		seen = append(seen, body)
		switch r.URL.Path {
		case "/twirp/livekit.RoomService/UpdateParticipant":
			if body["identity"] == "gone" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{}`))
		case "/twirp/livekit.RoomService/ListParticipants":
			_, _ = w.Write([]byte(`{"participants":[{"identity":"u1","name":"Ada","permission":{"canPublish":true}},{"identity":"u2","name":"Ben","permission":{}}]}`))
		default:
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	defer srv.Close()
	spec, _ := Lookup(KindLiveKit)
	p, err := spec.New(Config{Values: map[string]string{"url": srv.URL, "apiKey": "APIk"}, Secrets: map[string]string{"apiSecret": "lk-secret"}}, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	m, ok := p.(Moderator)
	if !ok {
		t.Fatal("livekit must moderate")
	}
	if err := m.SetPublish(context.Background(), "room.g.lobby", "u1", false); err != nil {
		t.Fatal(err)
	}
	video, _ := claims["video"].(map[string]any)
	if video["roomAdmin"] != true || video["room"] != "room.g.lobby" {
		t.Fatalf("claims %v", claims)
	}
	perm, _ := seen[0]["permission"].(map[string]any)
	if perm["canPublish"] != false || perm["canSubscribe"] != true || len(perm["canPublishSources"].([]any)) != 0 {
		t.Fatalf("permission %v", perm)
	}
	if err := m.SetPublish(context.Background(), "room.g.lobby", "gone", true); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("missing participant: %v", err)
	}
	list, err := m.Participants(context.Background(), "room.g.lobby")
	if err != nil || len(list) != 2 || !list[0].CanPublish || list[1].CanPublish || !list[1].Connected {
		t.Fatalf("participants %v %v", list, err)
	}
}

func TestCommsModeratorMutesThroughTheServerAPI(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		calls = append(calls, r.Method+" "+r.URL.Path+" "+string(raw))
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/players/u1/mute"):
			_, _ = w.Write([]byte(`{"id":"u1","muted":true}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/players/u9/mute"):
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":{"code":"not_listed"}}`))
		case r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"players":[{"id":"u1","name":"Ada","muted":true,"connected":true}]}`))
		}
	}))
	defer srv.Close()
	spec, _ := Lookup(KindComms)
	p, err := spec.New(Config{Values: map[string]string{"url": srv.URL}, Secrets: map[string]string{"serverKey": "k"}}, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	m := p.(Moderator)
	if err := m.SetPublish(context.Background(), "abc.node", "u1", false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(calls[0], `/v1/server/game-rooms/abc.node/players/u1/mute {"muted":true}`) {
		t.Fatalf("call %q", calls[0])
	}
	if err := m.SetPublish(context.Background(), "abc.node", "u9", true); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("not listed: %v", err)
	}
	list, err := m.Participants(context.Background(), "abc.node")
	if err != nil || len(list) != 1 || list[0].CanPublish || !list[0].Connected {
		t.Fatalf("participants %v %v", list, err)
	}
}
