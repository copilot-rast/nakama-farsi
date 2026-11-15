package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Minimal user_edge behaviour of Nakama's friend calls, enough for the RPCs.
func (n *fakeNakama) FriendsAdd(ctx context.Context, userID, _ string, ids, _ []string, _ map[string]any) error {
	for _, id := range ids {
		mine, theirs := -1, -1
		n.db.QueryRowContext(ctx, `SELECT state FROM user_edge WHERE source_id=$1 AND destination_id=$2`, userID, id).Scan(&mine)
		n.db.QueryRowContext(ctx, `SELECT state FROM user_edge WHERE source_id=$1 AND destination_id=$2`, id, userID).Scan(&theirs)
		var q string
		switch {
		case mine == edgeReceived:
			q = `UPDATE user_edge SET state=0, update_time=now() WHERE (source_id=$1 AND destination_id=$2) OR (source_id=$2 AND destination_id=$1)`
		case mine == -1 && theirs != edgeBlocked:
			q = `INSERT INTO user_edge (source_id, destination_id, state, position, update_time) VALUES ($1,$2,1,1,now()),($2,$1,2,1,now())`
		default:
			continue
		}
		if _, err := n.db.ExecContext(ctx, q, userID, id); err != nil {
			return err
		}
	}
	return nil
}

func (n *fakeNakama) FriendsBlock(ctx context.Context, userID, _ string, ids, _ []string) error {
	for _, id := range ids {
		if _, err := n.db.ExecContext(ctx, `DELETE FROM user_edge WHERE (source_id=$1 AND destination_id=$2) OR (source_id=$2 AND destination_id=$1 AND state <> 3)`, userID, id); err != nil {
			return err
		}
		if _, err := n.db.ExecContext(ctx, `INSERT INTO user_edge (source_id, destination_id, state, position, update_time) VALUES ($1,$2,3,1,now())`, userID, id); err != nil {
			return err
		}
	}
	return nil
}

func (n *fakeNakama) edges(ctx context.Context, q string, args ...any) ([]*api.Friend, error) {
	rows, err := n.db.QueryContext(ctx, `SELECT u.id::text, u.username, u.display_name, e.state FROM user_edge e JOIN users u ON u.id=e.destination_id WHERE `+q+` ORDER BY u.username`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*api.Friend{}
	for rows.Next() {
		u := &api.User{}
		var display sql.NullString
		var state int32
		if err := rows.Scan(&u.Id, &u.Username, &display, &state); err != nil {
			return nil, err
		}
		u.DisplayName = display.String
		out = append(out, &api.Friend{User: u, State: wrapperspb.Int32(state), UpdateTime: timestamppb.Now()})
	}
	return out, rows.Err()
}

func (n *fakeNakama) FriendsList(ctx context.Context, userID string, limit int, state *int, _ string) ([]*api.Friend, string, error) {
	s := -1
	if state != nil {
		s = *state
	}
	f, err := n.edges(ctx, `e.source_id=$1 AND ($2 < 0 OR e.state=$2)`, userID, s)
	if len(f) > limit {
		f = f[:limit]
	}
	return f, "", err
}

func (n *fakeNakama) UsersGetFriendStatus(ctx context.Context, userID string, ids []string) ([]*api.Friend, error) {
	return n.edges(ctx, `e.source_id=$1 AND e.destination_id = ANY($2::uuid[])`, userID, ids)
}

func (n *fakeNakama) users(ctx context.Context, q string, arg any) ([]*api.User, error) {
	rows, err := n.db.QueryContext(ctx, `SELECT id::text, username, display_name FROM users WHERE `+q+` ORDER BY username`, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*api.User{}
	for rows.Next() {
		u := &api.User{}
		var display sql.NullString
		if err := rows.Scan(&u.Id, &u.Username, &display); err != nil {
			return nil, err
		}
		u.DisplayName = display.String
		out = append(out, u)
	}
	return out, rows.Err()
}

func (n *fakeNakama) UsersGetId(ctx context.Context, ids, _ []string) ([]*api.User, error) {
	return n.users(ctx, `id = ANY($1::uuid[])`, ids)
}

func (n *fakeNakama) UsersGetUsername(ctx context.Context, names []string) ([]*api.User, error) {
	return n.users(ctx, `username = ANY($1::text[])`, names)
}

func (n *fakeNakama) AccountGetId(ctx context.Context, userID string) (*api.Account, error) {
	users, err := n.users(ctx, `id = $1::uuid`, userID)
	if err != nil || len(users) == 0 {
		return nil, errors.New("account not found")
	}
	return &api.Account{User: users[0], DisableTime: timestamppb.Now()}, nil
}

// The hook chain is process wide; the hooks forward to the service of the
// running test so a closed database from an earlier test is never used.
var (
	socialHookOnce sync.Once
	socialHookSvc  *service
)

func registerSocialHooks(h *harness) {
	socialHookSvc = h.s
	socialHookOnce.Do(func() {
		hookchain.Before("AddFriends", func(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AddFriendsRequest) (*api.AddFriendsRequest, error) {
			return socialHookSvc.beforeAddFriends(ctx, l, db, nk, in)
		})
		hookchain.AfterIn("AddFriends", func(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AddFriendsRequest) error {
			return socialHookSvc.afterAddFriends(ctx, l, db, nk, in)
		})
		hookchain.Before("DeleteFriends", func(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.DeleteFriendsRequest) (*api.DeleteFriendsRequest, error) {
			return socialHookSvc.beforeDeleteFriends(ctx, l, db, nk, in)
		})
		hookchain.Before("ListFriends", func(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.ListFriendsRequest) (*api.ListFriendsRequest, error) {
			return socialHookSvc.beforeListFriends(ctx, l, db, nk, in)
		})
		hookchain.Before("UpdateAccount", func(ctx context.Context, l runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.UpdateAccountRequest) (*api.UpdateAccountRequest, error) {
			return socialHookSvc.beforeUpdateAccount(ctx, l, db, nk, in)
		})
	})
}

func (h *harness) social(user, name string, payload any) (map[string]any, error) {
	h.t.Helper()
	raw, _ := json.Marshal(payload)
	out, err := h.rpcs[name](asUser(user), user, string(raw))
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(out)
	m := map[string]any{}
	if err := json.Unmarshal(b, &m); err != nil {
		h.t.Fatalf("%s answer is not an object: %s", name, b)
	}
	return m, nil
}

func (h *harness) mustSocial(user, name string, payload any) map[string]any {
	h.t.Helper()
	out, err := h.social(user, name, payload)
	if err != nil {
		h.t.Fatalf("%s: %v", name, err)
	}
	return out
}

// socialCode returns the error code: the code of a JSON message, the message of
// a runtime error, or the code of an app error.
func (h *harness) socialCode(user, name string, payload any) (string, int) {
	h.t.Helper()
	_, err := h.social(user, name, payload)
	if err == nil {
		h.t.Fatalf("%s %v should fail", name, payload)
	}
	err = h.s.runtimeError(err)
	var rt *runtime.Error
	if !errors.As(err, &rt) {
		h.t.Fatalf("%s: not a runtime error: %v", name, err)
	}
	var body struct {
		Code string `json:"code"`
	}
	if json.Unmarshal([]byte(rt.Message), &body) == nil && body.Code != "" {
		return body.Code, rt.Code
	}
	return rt.Message, rt.Code
}

func friendRows(out map[string]any) map[string]float64 {
	res := map[string]float64{}
	list, _ := out["friends"].([]any)
	for _, it := range list {
		f := it.(map[string]any)
		user := f["user"].(map[string]any)
		state, _ := f["state"].(float64)
		res[user["id"].(string)] = state
	}
	return res
}

func TestSocialFriendRPCsApplyFriendRules(t *testing.T) {
	h := newHarness(t)
	resetFriendTables(h)
	registerSocialHooks(h)
	h.setFriendRules(1, 0, 0, 24)
	a, b, c := h.user("fa"), h.user("fa"), h.user("en")

	out := h.mustSocial(a, "base_friends_add", map[string]any{"ids": []string{b}})
	if friendRows(out)[b] != edgeSent {
		t.Fatalf("add answer = %v", out)
	}
	if got := friendRows(h.mustSocial(b, "base_friends_list", map[string]any{"state": 2})); got[a] != edgeReceived || len(got) != 1 {
		t.Fatalf("received list = %v", got)
	}
	if code, grpc := h.socialCode(a, "base_friends_add", map[string]any{"ids": []string{c}}); code != CodeFriendDailyLimit || grpc != int(codes.ResourceExhausted) {
		t.Fatalf("daily limit over RPC = %s %d", code, grpc)
	}
	_, err := h.social(a, "base_friends_add", map[string]any{"ids": []string{c}, "lang": "en"})
	var rt *runtime.Error
	fe := friendError{}
	if !errors.As(err, &rt) || json.Unmarshal([]byte(rt.Message), &fe) != nil || fe.Lang != "en" || fe.Remaining == nil || *fe.Remaining != 0 {
		t.Fatalf("lang in the body picks the message language: %v", err)
	}
	// Accepting is not a new request and is not counted.
	var bName string
	h.s.db.QueryRow(`SELECT username FROM users WHERE id=$1`, b).Scan(&bName)
	out = h.mustSocial(b, "base_friends_add", map[string]any{"usernames": []string{mustName(h, a)}})
	if friendRows(out)[a] != edgeFriend {
		t.Fatalf("accept by username = %v", out)
	}
	list := h.mustSocial(a, "base_friends_list", map[string]any{})
	if friendRows(list)[b] != edgeFriend {
		t.Fatalf("friend list = %v", list)
	}
	first := list["friends"].([]any)[0].(map[string]any)
	if u := first["user"].(map[string]any); u["username"] != bName || first["update_time"] == nil {
		t.Fatalf("Nakama JSON shape (snake_case) = %v", first)
	}

	// Declining an invite over RPC starts the cooldown for the sender.
	h.setFriendRules(0, 0, 0, 24)
	h.mustSocial(c, "base_friends_add", map[string]any{"ids": []string{a}})
	h.mustSocial(a, "base_friends_remove", map[string]any{"ids": []string{c}})
	if code, _ := h.socialCode(c, "base_friends_add", map[string]any{"ids": []string{a}}); code != CodeFriendCooldown {
		t.Fatalf("cooldown after decline = %s", code)
	}

	// Max friends.
	h.setFriendRules(0, 1, 0, 0)
	d := h.user("fa")
	if code, _ := h.socialCode(a, "base_friends_add", map[string]any{"ids": []string{d}}); code != CodeFriendLimit {
		t.Fatalf("max friends = %s", code)
	}

	// Block, then unblock with remove.
	out = h.mustSocial(a, "base_friends_block", map[string]any{"ids": []string{d}})
	if friendRows(out)[d] != edgeBlocked {
		t.Fatalf("block = %v", out)
	}
	if got := friendRows(h.mustSocial(a, "base_friends_list", map[string]any{"state": 3})); got[d] != edgeBlocked {
		t.Fatalf("blocked list = %v", got)
	}
	if got := friendRows(h.mustSocial(d, "base_friends_add", map[string]any{"ids": []string{a}})); len(got) != 0 {
		t.Fatalf("an invite to someone who blocked you is silent: %v", got)
	}
	if got := friendRows(h.mustSocial(a, "base_friends_remove", map[string]any{"ids": []string{d}})); len(got) != 0 {
		t.Fatalf("unblock = %v", got)
	}

	for _, tc := range []struct {
		name    string
		payload any
		code    string
	}{
		{"base_friends_add", map[string]any{}, "NO_USERS"},
		{"base_friends_add", map[string]any{"ids": []string{a}}, "CANNOT_TARGET_SELF"},
		{"base_friends_add", map[string]any{"usernames": []string{mustName(h, a)}}, "CANNOT_TARGET_SELF"},
		{"base_friends_block", map[string]any{"ids": []string{"nope"}}, "INVALID_USER_ID"},
		{"base_friends_remove", map[string]any{"usernames": []string{"no_such_user_zz"}}, "USER_NOT_FOUND"},
		{"base_friends_add", map[string]any{"ids": make([]string, 101)}, CodeFriendTooManyNames},
		{"base_friends_list", map[string]any{"limit": 101}, "INVALID_LIMIT"},
		{"base_friends_list", map[string]any{"state": 4}, "INVALID_STATE"},
		{"base_friends_list", "[]", "INVALID_JSON"},
	} {
		var payload any = tc.payload
		if s, ok := tc.payload.(string); ok {
			payload = json.RawMessage(s)
		}
		if code, _ := h.socialCode(a, tc.name, payload); code != tc.code {
			t.Errorf("%s %v = %s, want %s", tc.name, tc.payload, code, tc.code)
		}
	}
}

func mustName(h *harness, id string) string {
	var name string
	if err := h.s.db.QueryRow(`SELECT username FROM users WHERE id=$1`, id).Scan(&name); err != nil {
		h.t.Fatal(err)
	}
	return name
}

func TestSocialAccountAndUsersRPCs(t *testing.T) {
	h := newHarness(t)
	registerSocialHooks(h)
	t.Cleanup(func() {
		_, _ = h.s.db.Exec(`UPDATE base_account_rules SET username_min=1, username_max=128, username_charset='any', password_min=8, password_letter=false, password_digit=false, display_name_max=255`)
	})
	h.mustAPI(http.StatusOK, http.MethodPut, "/account-rules", map[string]any{"usernameMin": 4, "usernameMax": 20, "usernameCharset": "latin", "passwordMin": 8, "displayNameMax": 10})
	a, b := h.user("fa"), h.user("fa")

	acc := h.mustSocial(a, "base_account_get", map[string]any{})
	if acc["user"].(map[string]any)["id"] != a || acc["disable_time"] != nil {
		t.Fatalf("account = %v", acc)
	}
	for payload, want := range map[string]string{
		`{"username":"ab"}`:                   "USERNAME_TOO_SHORT",
		`{"username":"سارا_جان"}`:             "USERNAME_INVALID_CHARACTERS",
		`{"display_name":"یک نام بلند خیلی"}`: "DISPLAY_NAME_TOO_LONG",
	} {
		if code, _ := h.socialCode(a, "base_account_update", json.RawMessage(payload)); code != want {
			t.Errorf("%s = %s, want %s", payload, code, want)
		}
	}
	if code, grpc := h.socialCode(a, "base_account_update", map[string]any{"username": mustName(h, b)}); code != "USERNAME_TAKEN" || grpc != int(codes.AlreadyExists) {
		t.Fatalf("taken username = %s %d", code, grpc)
	}
	acc = h.mustSocial(a, "base_account_update", map[string]any{"username": "saman92x", "display_name": "سامان"})
	if u := acc["user"].(map[string]any); u["username"] != "saman92x" || u["display_name"] != "سامان" {
		t.Fatalf("updated account = %v", acc)
	}
	acc = h.mustSocial(a, "base_account_update", map[string]any{"username": "saman92x"})
	if acc["user"].(map[string]any)["display_name"] != "سامان" {
		t.Fatalf("fields left out stay as they were: %v", acc)
	}

	users := h.mustSocial(b, "base_users_get", map[string]any{"usernames": []string{"saman92x"}, "ids": []string{a, b}})
	list, _ := users["users"].([]any)
	if len(list) != 2 {
		t.Fatalf("users by id and username, no duplicates = %v", users)
	}
	if got := h.mustSocial(b, "base_users_get", map[string]any{"usernames": []string{"nobody_here_x"}}); len(got) != 0 {
		t.Fatalf("no match is an empty object like the HTTP call: %v", got)
	}
	if code, _ := h.socialCode(b, "base_users_get", map[string]any{"ids": []string{"x"}}); code != "INVALID_USER_ID" {
		t.Fatalf("bad id = %s", code)
	}
}
