package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/privatemeta"
	"github.com/heroiclabs/nakama/v3/internal/stealthbots"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	socialListMax = 100
	socialUserMax = 100
)

var nakamaJSON = protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true}

func protoOut(m proto.Message) (json.RawMessage, error) {
	b, err := nakamaJSON.Marshal(m)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func (s *service) socialRPCs() map[string]clientRPC {
	rpcs := map[string]clientRPC{
		"base_friends_list":   s.rpcFriendsList,
		"base_friends_add":    s.rpcFriendsAdd,
		"base_friends_remove": s.rpcFriendsRemove,
		"base_friends_block":  s.rpcFriendsBlock,
		"base_account_get":    s.rpcAccountGet,
		"base_account_update": s.rpcAccountUpdate,
		"base_users_get":      s.rpcUsersGet,
	}
	for name, fn := range rpcs {
		rpcs[name] = withBodyLang(fn)
	}
	return rpcs
}

func withBodyLang(fn clientRPC) clientRPC {
	return func(ctx context.Context, userID, payload string) (any, error) {
		var in struct {
			Lang string `json:"lang"`
		}
		if json.Unmarshal([]byte(payload), &in) == nil && strings.TrimSpace(in.Lang) != "" {
			ctx = context.WithValue(ctx, runtime.RUNTIME_CTX_LANG, strings.TrimSpace(in.Lang))
		}
		return fn(ctx, userID, payload)
	}
}

func hookDisabled() error {
	return runtime.NewError("NOT_FOUND", int(codes.NotFound))
}

func (s *service) rpcFriendsList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		State  *int   `json:"state"`
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Limit == 0 {
		in.Limit = socialListMax
	}
	if in.Limit < 1 || in.Limit > socialListMax {
		return nil, badRequest("INVALID_LIMIT", "limit", "between 1 and 100")
	}
	req := &api.ListFriendsRequest{Limit: wrapperspb.Int32(int32(in.Limit)), Cursor: in.Cursor}
	if in.State != nil {
		if *in.State < 0 || *in.State > 3 {
			return nil, badRequest("INVALID_STATE", "state", "0, 1, 2 or 3")
		}
		req.State = wrapperspb.Int32(int32(*in.State))
	}
	req, err := hookchain.RunBefore(ctx, "ListFriends", s.logger, s.db, s.nk, req)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, hookDisabled()
	}
	var state *int
	if req.State != nil {
		v := int(req.State.Value)
		state = &v
	}
	limit := in.Limit
	if req.Limit != nil && req.Limit.Value >= 1 && req.Limit.Value <= socialListMax {
		limit = int(req.Limit.Value)
	}
	friends, cursor, err := s.nk.FriendsList(ctx, userID, limit, state, req.Cursor)
	if err != nil {
		if strings.Contains(err.Error(), "cursor") {
			return nil, badRequest("INVALID_CURSOR", "cursor", "cursor is not valid")
		}
		return nil, err
	}
	out := &api.FriendList{Friends: friends, Cursor: cursor}
	_ = hookchain.RunAfterIn(ctx, "ListFriends", s.logger, s.db, s.nk, out)
	privatemeta.Friends(out.Friends)
	stealthbots.MarkFriends(ctx, s.db, out.Friends)
	return protoOut(out)
}

type friendTargetsIn struct {
	IDs       []string `json:"ids"`
	Usernames []string `json:"usernames"`
	Codes     []string `json:"codes"`
}

func (s *service) socialTargets(ctx context.Context, userID, payload string) (friendTargetsIn, error) {
	var in friendTargetsIn
	if err := decodeStrict(payload, &in); err != nil {
		return in, err
	}
	if len(in.IDs)+len(in.Usernames)+len(in.Codes) == 0 {
		return in, badRequest("NO_USERS", "ids", "give ids, usernames or codes")
	}
	if len(in.IDs)+len(in.Usernames)+len(in.Codes) > friendLookupMax {
		return in, badRequest(CodeFriendTooManyNames, "ids", "at most 100 users")
	}
	for _, id := range in.IDs {
		u, err := uuid.FromString(id)
		if err != nil || u == uuid.Nil {
			return in, badRequest("INVALID_USER_ID", "ids", "not a user ID")
		}
		if id == userID {
			return in, badRequest("CANNOT_TARGET_SELF", "ids", "this is your own account")
		}
	}
	for _, name := range in.Usernames {
		if name == "" {
			return in, badRequest("INVALID_USERNAME", "usernames", "empty username")
		}
	}
	if len(in.Codes) > 0 {
		ids, err := s.idsForCodes(ctx, userID, in.Codes)
		if err != nil {
			return in, err
		}
		in.IDs = append(in.IDs, ids...)
		in.Codes = nil
	}
	return in, nil
}

func (s *service) resolveUsernames(ctx context.Context, userID string, ids, usernames []string) ([]string, error) {
	out := append([]string{}, ids...)
	if len(usernames) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM users WHERE username = ANY($1::text[])`, usernames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if id == userID {
			return nil, badRequest("CANNOT_TARGET_SELF", "usernames", "this is your own account")
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *service) friendStatus(ctx context.Context, userID string, ids []string) (any, error) {
	friends, err := s.nk.UsersGetFriendStatus(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	privatemeta.Friends(friends)
	stealthbots.MarkFriends(ctx, s.db, friends)
	return protoOut(&api.FriendList{Friends: friends})
}

func (s *service) rpcFriendsAdd(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.socialTargets(ctx, userID, payload)
	if err != nil {
		return nil, err
	}
	req, err := hookchain.RunBefore(ctx, "AddFriends", s.logger, s.db, s.nk, &api.AddFriendsRequest{Ids: in.IDs, Usernames: in.Usernames})
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, hookDisabled()
	}
	ids, err := s.resolveUsernames(ctx, userID, req.GetIds(), req.GetUsernames())
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, notFound("USER_NOT_FOUND")
	}
	username, _ := ctx.Value(runtime.RUNTIME_CTX_USERNAME).(string)
	if err := s.nk.FriendsAdd(ctx, userID, username, ids, nil, nil); err != nil {
		return nil, err
	}
	_ = hookchain.RunAfterIn(ctx, "AddFriends", s.logger, s.db, s.nk, req)
	return s.friendStatus(ctx, userID, ids)
}

func (s *service) rpcFriendsRemove(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.socialTargets(ctx, userID, payload)
	if err != nil {
		return nil, err
	}
	req, err := hookchain.RunBefore(ctx, "DeleteFriends", s.logger, s.db, s.nk, &api.DeleteFriendsRequest{Ids: in.IDs, Usernames: in.Usernames})
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, hookDisabled()
	}
	ids, err := s.resolveUsernames(ctx, userID, req.GetIds(), req.GetUsernames())
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, notFound("USER_NOT_FOUND")
	}
	username, _ := ctx.Value(runtime.RUNTIME_CTX_USERNAME).(string)
	if err := s.nk.FriendsDelete(ctx, userID, username, ids, nil); err != nil {
		return nil, err
	}
	_ = hookchain.RunAfterIn(ctx, "DeleteFriends", s.logger, s.db, s.nk, req)
	return s.friendStatus(ctx, userID, ids)
}

func (s *service) rpcFriendsBlock(ctx context.Context, userID, payload string) (any, error) {
	in, err := s.socialTargets(ctx, userID, payload)
	if err != nil {
		return nil, err
	}
	req, err := hookchain.RunBefore(ctx, "BlockFriends", s.logger, s.db, s.nk, &api.BlockFriendsRequest{Ids: in.IDs, Usernames: in.Usernames})
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, hookDisabled()
	}
	ids, err := s.resolveUsernames(ctx, userID, req.GetIds(), req.GetUsernames())
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, notFound("USER_NOT_FOUND")
	}
	username, _ := ctx.Value(runtime.RUNTIME_CTX_USERNAME).(string)
	if err := s.nk.FriendsBlock(ctx, userID, username, ids, nil); err != nil {
		return nil, err
	}
	_ = hookchain.RunAfterIn(ctx, "BlockFriends", s.logger, s.db, s.nk, req)
	return s.friendStatus(ctx, userID, ids)
}

func (s *service) rpcAccountGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct{}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	account, err := s.nk.AccountGetId(ctx, userID)
	if err != nil {
		return nil, err
	}
	account.DisableTime = nil
	privatemeta.Account(account)
	return protoOut(account)
}

func (s *service) rpcAccountUpdate(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Username    *string `json:"username"`
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
		LangTag     *string `json:"lang_tag"`
		Location    *string `json:"location"`
		Timezone    *string `json:"timezone"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	wrap := func(v *string) *wrapperspb.StringValue {
		if v == nil || *v == "" {
			return nil
		}
		return wrapperspb.String(*v)
	}
	req := &api.UpdateAccountRequest{Username: wrap(in.Username), DisplayName: wrap(in.DisplayName), AvatarUrl: wrap(in.AvatarURL),
		LangTag: wrap(in.LangTag), Location: wrap(in.Location), Timezone: wrap(in.Timezone)}
	req, err := hookchain.RunBefore(ctx, "UpdateAccount", s.logger, s.db, s.nk, req)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, hookDisabled()
	}
	username := req.GetUsername().GetValue()
	if req.GetUsername() != nil {
		if len(username) > 128 {
			return nil, runtime.NewError("USERNAME_TOO_LONG", int(codes.InvalidArgument))
		}
		var taken bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username=$1 AND id<>$2::uuid)`, username, userID).Scan(&taken); err != nil {
			return nil, err
		}
		if taken {
			return nil, runtime.NewError("USERNAME_TAKEN", int(codes.AlreadyExists))
		}
	}
	err = s.nk.AccountUpdateId(ctx, userID, username, nil, req.GetDisplayName().GetValue(), req.GetTimezone().GetValue(),
		req.GetLocation().GetValue(), req.GetLangTag().GetValue(), req.GetAvatarUrl().GetValue())
	if err != nil {
		if strings.Contains(err.Error(), "already in use") {
			return nil, runtime.NewError("USERNAME_TAKEN", int(codes.AlreadyExists))
		}
		return nil, err
	}
	_ = hookchain.RunAfterIn(ctx, "UpdateAccount", s.logger, s.db, s.nk, req)
	return s.rpcAccountGet(ctx, userID, "")
}

func (s *service) rpcUsersGet(ctx context.Context, _ string, payload string) (any, error) {
	var in friendTargetsIn
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if len(in.IDs)+len(in.Usernames) > socialUserMax {
		return nil, badRequest(CodeFriendTooManyNames, "ids", "at most 100 users")
	}
	for _, id := range in.IDs {
		if _, err := uuid.FromString(id); err != nil {
			return nil, badRequest("INVALID_USER_ID", "ids", "not a user ID")
		}
	}
	req, err := hookchain.RunBefore(ctx, "GetUsers", s.logger, s.db, s.nk, &api.GetUsersRequest{Ids: in.IDs, Usernames: in.Usernames})
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, hookDisabled()
	}
	out := &api.Users{Users: []*api.User{}}
	seen := map[string]bool{}
	add := func(users []*api.User) {
		for _, u := range users {
			if !seen[u.Id] {
				seen[u.Id] = true
				out.Users = append(out.Users, u)
			}
		}
	}
	if len(req.GetIds()) > 0 {
		users, err := s.nk.UsersGetId(ctx, req.GetIds(), nil)
		if err != nil {
			return nil, err
		}
		add(users)
	}
	if len(req.GetUsernames()) > 0 {
		users, err := s.nk.UsersGetUsername(ctx, req.GetUsernames())
		if err != nil {
			return nil, err
		}
		add(users)
	}
	_ = hookchain.RunAfter(ctx, "GetUsers", s.logger, s.db, s.nk, out, req)
	privatemeta.Users(out.Users)
	stealthbots.MarkUsers(ctx, s.db, out.Users)
	return protoOut(out)
}
