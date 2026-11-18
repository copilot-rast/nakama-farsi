package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/friendroom"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type roomNakama struct {
	*fakeNakama
	mu       sync.Mutex
	n        int
	matches  map[string]*friendroom.Params
	status   map[string]string
	friends  map[string][]string
	notified []string
}

func (n *roomNakama) MatchCreate(_ context.Context, module string, params map[string]any) (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.n++
	id := "00000000-0000-4000-8000-00000000000" + string(rune('0'+n.n)) + ".node"
	p := friendroom.FromParams(params)
	n.matches[id], n.status[id] = &p, friendroom.StatusWaiting
	return id, nil
}

func (n *roomNakama) MatchGet(_ context.Context, id string) (*api.Match, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	p, ok := n.matches[id]
	if !ok {
		return nil, runtime.NewError("not found", 5)
	}
	return &api.Match{MatchId: id, Authoritative: true, Label: wrapperspb.String(p.Label(n.status[id], nil))}, nil
}

func (n *roomNakama) MatchList(_ context.Context, _ int, _ bool, _ string, _, _ *int, query string) ([]*api.Match, error) {
	n.mu.Lock()
	ids := []string{}
	for id := range n.matches {
		ids = append(ids, id)
	}
	n.mu.Unlock()
	out := []*api.Match{}
	for _, id := range ids {
		m, _ := n.MatchGet(context.Background(), id)
		out = append(out, m)
	}
	return out, nil
}

func (n *roomNakama) MatchSignal(_ context.Context, id, data string) (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	p, ok := n.matches[id]
	if !ok {
		return "", runtime.NewError("not found", 5)
	}
	sig, _ := friendroom.ParseSignal(data)
	if n.status[id] != friendroom.StatusWaiting {
		return friendroom.Answer(false, "STARTED"), nil
	}
	switch {
	case sig.Action == friendroom.ActionCancel && sig.UserID == p.Host, sig.Action == friendroom.ActionDecline && sig.UserID == p.Invited:
		n.status[id] = friendroom.StatusFinished
		return friendroom.Answer(true, ""), nil
	}
	return friendroom.Answer(false, "NOT_ALLOWED"), nil
}

func (n *roomNakama) FriendsList(_ context.Context, userID string, _ int, _ *int, _ string) ([]*api.Friend, string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := []*api.Friend{}
	for _, f := range n.friends[userID] {
		out = append(out, &api.Friend{User: &api.User{Id: f}, State: wrapperspb.Int32(0)})
	}
	return out, "", nil
}

func (n *roomNakama) NotificationSend(_ context.Context, userID, subject string, content map[string]any, code int, _ string, _ bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	raw, _ := json.Marshal(content)
	n.notified = append(n.notified, userID+"|"+subject+"|"+string(raw))
	return nil
}

func TestGameFriendRoomsAndBot(t *testing.T) {
	h := newGameHarness(t)
	for name, fn := range h.s.gameRoomRPCs() {
		h.rpcs[name] = fn
	}
	nk := &roomNakama{fakeNakama: h.nk, matches: map[string]*friendroom.Params{}, status: map[string]string{}, friends: map[string][]string{}}
	h.s.nk = nk
	h.boardGame(map[string]any{"botsAllowed": false})
	host, friend, other := h.user(), h.user(), h.user()
	nk.friends[host] = []string{friend}
	for _, u := range []string{host, friend, other} {
		h.give(u, 100)
	}

	if code := h.rpcCode(host, "economy_game_friend_create", map[string]any{"key": "tst_board", "room": "bronze", "friendId": other}); code != "NOT_FRIENDS" {
		t.Fatal(code)
	}
	if code := h.rpcCode(host, "economy_game_friend_create", map[string]any{"key": "tst_board", "room": "bronze", "friendId": host}); code != "CANNOT_INVITE_SELF" {
		t.Fatal(code)
	}
	if code := h.rpcCode(host, "economy_game_friend_create", map[string]any{"key": "tst_solo"}); code != "GAME_NOT_REALTIME" {
		t.Fatal(code)
	}
	room := h.mustRPC(host, "economy_game_friend_create", map[string]any{"key": "tst_board", "room": "bronze", "friendId": friend})
	code, _ := room["code"].(string)
	if !friendroom.ValidCode(code) || room["existing"] != false || room["expiresInSeconds"] != float64(defaultFriendWaitSeconds) {
		t.Fatalf("room %v", room)
	}
	if len(nk.notified) != 1 || !strings.HasPrefix(nk.notified[0], friend+"|game_invite|") || !strings.Contains(nk.notified[0], room["matchId"].(string)) {
		t.Fatalf("invite %v", nk.notified)
	}
	again := h.mustRPC(host, "economy_game_friend_create", map[string]any{"key": "tst_board", "room": "bronze", "friendId": friend})
	if again["matchId"] != room["matchId"] || again["existing"] != true {
		t.Fatalf("second create must answer the open room: %v", again)
	}
	if c := h.rpcCode(other, "economy_game_friend_join", map[string]any{"code": code}); c != "NOT_INVITED" {
		t.Fatal(c)
	}
	if c := h.rpcCode(host, "economy_game_friend_join", map[string]any{"code": code}); c != "OWN_ROOM" {
		t.Fatal(c)
	}
	if c := h.rpcCode(friend, "economy_game_friend_join", map[string]any{"code": "abc"}); c != "INVALID_CODE" {
		t.Fatal(c)
	}
	joined := h.mustRPC(friend, "economy_game_friend_join", map[string]any{"code": strings.ToLower(code)})
	if joined["matchId"] != room["matchId"] || joined["room"] != "bronze" || joined["hostId"] != host {
		t.Fatalf("join %v", joined)
	}
	byID := h.mustRPC(friend, "economy_game_friend_join", map[string]any{"matchId": room["matchId"]})
	if byID["matchId"] != room["matchId"] {
		t.Fatalf("join by id %v", byID)
	}
	if c := h.rpcCode(other, "economy_game_friend_cancel", map[string]any{"matchId": room["matchId"]}); c != "NOT_CANCELLABLE" {
		t.Fatal(c)
	}
	h.mustRPC(friend, "economy_game_friend_decline", map[string]any{"matchId": room["matchId"]})
	if c := h.rpcCode(friend, "economy_game_friend_join", map[string]any{"code": code}); c != "ROOM_NOT_FOUND" {
		t.Fatal(c)
	}

	open := h.mustRPC(host, "economy_game_friend_create", map[string]any{"key": "tst_board", "room": "bronze"})
	if open["matchId"] == room["matchId"] || open["friendId"] != "" {
		t.Fatalf("open room %v", open)
	}
	if got := h.mustRPC(other, "economy_game_friend_join", map[string]any{"code": open["code"]}); got["matchId"] != open["matchId"] {
		t.Fatalf("anyone with the code joins an open room: %v", got)
	}
	h.mustRPC(host, "economy_game_friend_cancel", map[string]any{"matchId": open["matchId"]})

	if c := h.rpcCode(host, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"}); c != "BOTS_DISABLED" {
		t.Fatal(c)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_game SET bots_allowed=true WHERE key='tst_board'`); err != nil {
		t.Fatal(err)
	}
	h.s.games.version = -1
	bot := h.mustRPC(host, "economy_game_bot", map[string]any{"key": "tst_board", "room": "bronze"})
	p := nk.matches[bot["matchId"].(string)]
	if p == nil || p.Mode != friendroom.ModeBot || len(p.Users) != 1 || p.Users[0] != host || bot["botAfterSeconds"] != float64(defaultBotAfterSeconds) {
		t.Fatalf("bot %v %+v", bot, p)
	}
	if c := h.rpcCode(host, "economy_game_bot", map[string]any{"key": "tst_board", "room": "gold"}); c != "LEVEL_TOO_LOW" {
		t.Fatal(c)
	}
}
