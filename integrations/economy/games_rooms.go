package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/friendroom"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

const (
	defaultFriendWaitSeconds = 120
	defaultInviteNotifyCode  = 2310
	defaultBotAfterSeconds   = 20
	gameInviteSubject        = "game_invite"
)

func (s *service) gameRoomRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_game_friend_create":  s.rpcGameFriendCreate,
		"economy_game_friend_join":    s.rpcGameFriendJoin,
		"economy_game_friend_cancel":  s.rpcGameFriendCancel,
		"economy_game_friend_decline": s.rpcGameFriendDecline,
		"economy_game_bot":            s.rpcGameBot,
	}
}

type gameRoomRequest struct {
	Key      string `json:"key"`
	Room     string `json:"room"`
	FriendID string `json:"friendId"`
	Code     string `json:"code"`
	MatchID  string `json:"matchId"`
	Lang     string `json:"lang"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

func (s *service) realtimeGame(ctx context.Context, userID string, in gameRoomRequest, path string) (*gameRecord, *gameTicket, error) {
	key := strings.TrimSpace(in.Key)
	if key == "" {
		return nil, nil, badRequest("GAME_KEY_REQUIRED", "key", "a game key is required")
	}
	def, ok := gamecatalog.Lookup(key)
	if !ok || !def.Realtime {
		return nil, nil, badRequest("GAME_NOT_REALTIME", "key", "this game has no live matches on this server")
	}
	t, err := s.gameJoinCheck(ctx, userID, key, in.Room, in.Platform, in.Version, in.Lang, path)
	if err != nil {
		return nil, nil, err
	}
	g, err := s.gameByKey(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	return g, t, nil
}

func gameSettingInt(g *gameRecord, name string, fallback int) int {
	return g.public().Int(name, fallback)
}

func gameSettingBool(g *gameRecord, name string, fallback bool) bool {
	return g.public().Bool(name, fallback)
}

func (s *service) rpcGameFriendCreate(ctx context.Context, userID, payload string) (any, error) {
	var in gameRoomRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	g, t, err := s.realtimeGame(ctx, userID, in, gamecatalog.PathFriend)
	if err != nil {
		return nil, err
	}
	if !gameSettingBool(g, "friendRooms", true) {
		return nil, fail(http.StatusForbidden, "FRIEND_ROOMS_DISABLED", "key", "private matches are off for this game")
	}
	friend := strings.TrimSpace(in.FriendID)
	if friend == userID {
		return nil, badRequest("CANNOT_INVITE_SELF", "friendId", "you cannot invite yourself")
	}
	if friend != "" && !friendroom.IsFriend(ctx, s.nk, userID, friend) {
		return nil, fail(http.StatusForbidden, "NOT_FRIENDS", "friendId", "you can only invite a friend")
	}
	wait := gameSettingInt(g, "friendWaitSeconds", defaultFriendWaitSeconds)
	if id, label := friendroom.Open(ctx, s.nk, g.Key, "host", userID); id != "" {
		code, _ := label["code"].(string)
		room, _ := label["room"].(string)
		invited, _ := label["invited"].(string)
		if friend == "" || friend == invited {
			return map[string]any{"matchId": id, "code": code, "key": g.Key, "room": room, "friendId": invited, "existing": true, "expiresInSeconds": friendroom.SecondsLeft(label, time.Now(), wait)}, nil
		}
		if _, err := s.nk.MatchSignal(ctx, id, friendroom.SignalData(friendroom.ActionCancel, userID)); err != nil {
			s.logger.Warn("Economy friend room %s cancel before a new invite failed: %v", id, err)
		}
	}
	p := friendroom.Params{Game: g.Key, Room: t.Room, Mode: friendroom.ModeFriend, Host: userID, Code: friendroom.NewCode(), Invited: friend, ExpiresAt: time.Now().Add(time.Duration(wait) * time.Second).UTC()}
	id, err := s.nk.MatchCreate(ctx, g.Key, p.Map())
	if err != nil {
		s.logger.Error("Economy friend room for %s failed: %v", g.Key, err)
		return nil, fail(http.StatusServiceUnavailable, "MATCH_CREATE_FAILED", "", "the match could not be made, try again")
	}
	friendroom.Remember(g.Key, p.Code, userID, id)
	if friend != "" {
		var name string
		_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(display_name,''), username) FROM users WHERE id=$1`, userID).Scan(&name)
		content := map[string]any{"game": g.Key, "matchId": id, "code": p.Code, "room": t.Room, "hostId": userID, "hostName": name, "expiresAt": p.ExpiresAt.Format(time.RFC3339)}
		if err := s.nk.NotificationSend(ctx, friend, gameInviteSubject, content, gameSettingInt(g, "inviteNotifyCode", defaultInviteNotifyCode), userID, true); err != nil {
			s.logger.Warn("Economy game invite to %s failed: %v", friend, err)
		}
	}
	return map[string]any{"matchId": id, "code": p.Code, "key": g.Key, "room": t.Room, "friendId": friend, "existing": false, "expiresInSeconds": wait}, nil
}

func (s *service) rpcGameFriendJoin(ctx context.Context, userID, payload string) (any, error) {
	var in gameRoomRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	matchID := strings.TrimSpace(in.MatchID)
	var id string
	var label map[string]any
	switch {
	case matchID != "":
		m, err := s.nk.MatchGet(ctx, matchID)
		if err != nil || m == nil {
			return nil, notFound("ROOM_NOT_FOUND")
		}
		label = map[string]any{}
		if err := jsonUnmarshalString(m.GetLabel().GetValue(), &label); err != nil || label["mode"] != friendroom.ModeFriend || label["status"] != friendroom.StatusWaiting {
			return nil, notFound("ROOM_NOT_FOUND")
		}
		id = matchID
	case friendroom.ValidCode(code):
		id, label = friendroom.Open(ctx, s.nk, "", "code", code)
		if id == "" {
			return nil, notFound("ROOM_NOT_FOUND")
		}
	default:
		return nil, badRequest("INVALID_CODE", "code", "send the 6-character room code or the match id")
	}
	host, _ := label["host"].(string)
	invited, _ := label["invited"].(string)
	key, _ := label["game"].(string)
	room, _ := label["room"].(string)
	if host == userID {
		return nil, badRequest("OWN_ROOM", "code", "this is your own room")
	}
	if invited != "" && invited != userID {
		return nil, fail(http.StatusForbidden, "NOT_INVITED", "code", "this room is for another player")
	}
	g, _, err := s.realtimeGame(ctx, userID, gameRoomRequest{Key: key, Room: room, Lang: in.Lang, Platform: in.Platform, Version: in.Version}, gamecatalog.PathFriend)
	if err != nil {
		return nil, err
	}
	if invited == "" && !gameSettingBool(g, "friendCodeJoin", true) {
		return nil, fail(http.StatusForbidden, "CODE_JOIN_DISABLED", "code", "joining with a code is off for this game")
	}
	return map[string]any{"matchId": id, "key": key, "room": room, "hostId": host, "expiresInSeconds": friendroom.SecondsLeft(label, time.Now(), 0)}, nil
}

func (s *service) friendSignal(ctx context.Context, userID, payload, action, refused string) (any, error) {
	var in gameRoomRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	id, err := cleanMatchID(in.MatchID)
	if err != nil {
		return nil, err
	}
	out, err := s.nk.MatchSignal(ctx, id, friendroom.SignalData(action, userID))
	if err != nil {
		return nil, notFound("ROOM_NOT_FOUND")
	}
	ok, reason := friendroom.ParseAnswer(out)
	if !ok {
		e := fail(http.StatusConflict, refused, "matchId", "")
		if reason != "" {
			e.Meta = map[string]any{"reason": reason}
		}
		return nil, e
	}
	return map[string]any{"matchId": id, "done": true}, nil
}

func (s *service) rpcGameFriendCancel(ctx context.Context, userID, payload string) (any, error) {
	return s.friendSignal(ctx, userID, payload, friendroom.ActionCancel, "NOT_CANCELLABLE")
}

func (s *service) rpcGameFriendDecline(ctx context.Context, userID, payload string) (any, error) {
	return s.friendSignal(ctx, userID, payload, friendroom.ActionDecline, "NOT_DECLINABLE")
}

func (s *service) rpcGameBot(ctx context.Context, userID, payload string) (any, error) {
	var in gameRoomRequest
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	g, t, err := s.realtimeGame(ctx, userID, in, gamecatalog.PathBot)
	if err != nil {
		return nil, err
	}
	if !g.BotsAllowed {
		return nil, fail(http.StatusForbidden, "BOTS_DISABLED", "key", "bot matches are off for this game")
	}
	stealth := g.stealth().Enabled
	p := friendroom.Params{Game: g.Key, Room: t.Room, Mode: friendroom.ModeBot, Users: []string{userID}, Stealth: stealth}
	id, err := s.nk.MatchCreate(ctx, g.Key, p.Map())
	if err != nil {
		s.logger.Error("Economy bot match for %s failed: %v", g.Key, err)
		return nil, fail(http.StatusServiceUnavailable, "MATCH_CREATE_FAILED", "", "the match could not be made, try again")
	}
	if stealth {
		return map[string]any{"matchId": id, "key": g.Key, "room": t.Room}, nil
	}
	return map[string]any{"matchId": id, "key": g.Key, "room": t.Room, "botAfterSeconds": gameSettingInt(g, "botAfterSeconds", defaultBotAfterSeconds)}, nil
}
