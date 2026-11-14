package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

func (n *fakeNakama) FriendsDelete(ctx context.Context, userID, _ string, ids, _ []string) error {
	for _, id := range ids {
		if _, err := n.db.ExecContext(ctx, `DELETE FROM user_edge WHERE (source_id=$1 AND destination_id=$2) OR (source_id=$2 AND destination_id=$1 AND state <> 3)`, userID, id); err != nil {
			return err
		}
	}
	return nil
}

func (h *harness) edge(from, to string, state int, age time.Duration) {
	h.t.Helper()
	if _, err := h.s.db.Exec(`INSERT INTO user_edge (source_id, destination_id, state, position, update_time) VALUES ($1,$2,$3,$4,now() - $5::INT8 * INTERVAL '1 second')
		ON CONFLICT (source_id, destination_id) DO UPDATE SET state=EXCLUDED.state, update_time=EXCLUDED.update_time`, from, to, state, time.Now().UnixNano(), int64(age.Seconds())); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) invite(from, to string, age time.Duration) {
	h.edge(from, to, edgeSent, age)
	h.edge(to, from, edgeReceived, age)
}

func (h *harness) befriend(a, b string) {
	h.edge(a, b, edgeFriend, 0)
	h.edge(b, a, edgeFriend, 0)
}

func (h *harness) setFriendRules(daily, maxFriends, expiry, cooldown int) {
	h.t.Helper()
	h.mustAPI(200, "PUT", "/friend-rules", map[string]any{"dailyRequests": daily, "maxFriends": maxFriends, "expiryDays": expiry, "rejectCooldownHours": cooldown})
}

func asUser(user string) context.Context {
	return context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, user)
}

func (h *harness) addFriends(user string, ids ...string) (*friendError, error) {
	h.t.Helper()
	_, err := h.s.beforeAddFriends(asUser(user), nil, nil, nil, &api.AddFriendsRequest{Ids: ids})
	if err == nil {
		return nil, nil
	}
	var rt *runtime.Error
	if !errors.As(err, &rt) {
		return nil, err
	}
	fe := &friendError{}
	if json.Unmarshal([]byte(rt.Message), fe) != nil {
		h.t.Fatalf("friend error is not JSON: %q", rt.Message)
	}
	fe.grpc = 0
	return fe, nil
}

func (h *harness) mustRefuse(code, user string, ids ...string) *friendError {
	h.t.Helper()
	fe, err := h.addFriends(user, ids...)
	if err != nil || fe == nil || fe.Code != code {
		h.t.Fatalf("want %s, got %+v %v", code, fe, err)
	}
	if fe.Message == "" || fe.Lang == "" {
		h.t.Fatalf("no translated message: %+v", fe)
	}
	return fe
}

func (h *harness) mustAllow(user string, ids ...string) {
	h.t.Helper()
	if fe, err := h.addFriends(user, ids...); err != nil || fe != nil {
		h.t.Fatalf("want allowed, got %+v %v", fe, err)
	}
}

func resetFriendTables(h *harness) {
	for _, q := range []string{`DELETE FROM base_friend_request_day`, `DELETE FROM base_friend_rejection`, `DELETE FROM cluster_lease WHERE name='` + friendsLease + `'`} {
		if _, err := h.s.db.Exec(q); err != nil {
			h.t.Fatal(err)
		}
	}
	h.t.Cleanup(func() {
		h.s.db.Exec(`UPDATE base_friend_rules SET daily_requests=20, max_friends=0, expiry_days=0, reject_cooldown_hours=0, updated_by='' WHERE id=1`)
	})
}

func TestFriendRulesSettingsAndRPC(t *testing.T) {
	h := newHarness(t)
	resetFriendTables(h)
	h.s.db.Exec(`UPDATE base_friend_rules SET daily_requests=20, max_friends=0, expiry_days=0, reject_cooldown_hours=0 WHERE id=1`)
	got := h.mustAPI(200, "GET", "/friend-rules", nil)
	if got["dailyRequests"].(float64) != 20 || got["maxFriends"].(float64) != 0 || got["expiryDays"].(float64) != 0 || got["rejectCooldownHours"].(float64) != 0 {
		t.Fatalf("defaults %v", got)
	}
	h.expectError(400, "INVALID_RULE", "PUT", "/friend-rules", map[string]any{"dailyRequests": -1})
	h.expectError(400, "INVALID_RULE", "PUT", "/friend-rules", map[string]any{"dailyRequests": 5, "maxFriends": 20000})
	h.expectError(400, "INVALID_RULE", "PUT", "/friend-rules", map[string]any{"expiryDays": 400})
	h.expectError(400, "INVALID_RULE", "PUT", "/friend-rules", map[string]any{"rejectCooldownHours": 9000})
	h.setFriendRules(3, 50, 7, 24)

	a, b := h.user("fa"), h.user("en")
	h.mustAllow(a, b)
	out := h.mustRPC(a, "base_friend_rules", map[string]any{})
	if out["dailyRequests"].(float64) != 3 || out["sentToday"].(float64) != 1 || out["remainingToday"].(float64) != 2 || out["maxFriends"].(float64) != 50 || out["expiryDays"].(float64) != 7 || out["rejectCooldownHours"].(float64) != 24 {
		t.Fatalf("rpc %v", out)
	}
	if _, err := time.Parse(time.RFC3339, out["resetsAt"].(string)); err != nil {
		t.Fatal(err)
	}
	h.setFriendRules(0, 0, 0, 0)
	out = h.mustRPC(a, "base_friend_rules", map[string]any{})
	if out["remainingToday"] != nil || out["sentToday"].(float64) != 1 {
		t.Fatalf("unlimited %v", out)
	}
}

func TestFriendDailyLimitAcrossParallelCalls(t *testing.T) {
	h := newHarness(t)
	resetFriendTables(h)
	h.setFriendRules(5, 0, 0, 0)
	a := h.user("fa")
	targets := make([]string, 12)
	for i := range targets {
		targets[i] = h.user("fa")
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed, refused := 0, 0
	for _, to := range targets {
		wg.Add(1)
		go func(to string) {
			defer wg.Done()
			_, err := h.s.beforeAddFriends(asUser(a), nil, nil, nil, &api.AddFriendsRequest{Ids: []string{to}})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				allowed++
			} else {
				refused++
			}
		}(to)
	}
	wg.Wait()
	if allowed != 5 || refused != 7 {
		t.Fatalf("allowed=%d refused=%d", allowed, refused)
	}
	fe := h.mustRefuse(CodeFriendDailyLimit, a, h.user("fa"))
	if fe.Limit != 5 || fe.Remaining == nil || *fe.Remaining != 0 || fe.ResetsAt == "" || fe.Lang != "fa" {
		t.Fatalf("error %+v", fe)
	}
	if want := "امروز بیشتر از ۵ درخواست"; len(fe.Message) < len(want) || fe.Message[:len(want)] != want {
		t.Fatalf("message %q", fe.Message)
	}

	b := h.user("en")
	many := []string{h.user("en"), h.user("en"), h.user("en"), h.user("en"), h.user("en"), h.user("en")}
	fe = h.mustRefuse(CodeFriendDailyLimit, b, many...)
	if fe.Lang != "en" || *fe.Remaining != 5 {
		t.Fatalf("one request above the limit: %+v", fe)
	}
	h.mustAllow(b, many[:5]...)

	c := h.user("fa")
	h.invite(many[5], c, 0)
	h.mustAllow(c, many[5])
	if n, _ := h.s.sentToday(context.Background(), c, mustDay(h)); n != 0 {
		t.Fatalf("accepting an invite counted as a request: %d", n)
	}
	h.befriend(c, b)
	h.mustAllow(c, b)
	if n, _ := h.s.sentToday(context.Background(), c, mustDay(h)); n != 0 {
		t.Fatalf("an existing friend counted: %d", n)
	}

	h.s.db.Exec(`UPDATE base_friend_request_day SET day = day - 1 WHERE user_id=$1`, a)
	h.mustAllow(a, h.user("fa"))
}

func mustDay(h *harness) string {
	day, _ := dayBounds(h.s.now(), projectZone(context.Background(), h.s.db))
	return day
}

func TestFriendMaxAndCooldown(t *testing.T) {
	h := newHarness(t)
	resetFriendTables(h)
	h.setFriendRules(0, 2, 0, 48)
	a, b, c, d := h.user("fa"), h.user("fa"), h.user("fa"), h.user("fa")
	h.befriend(a, b)
	h.befriend(a, c)
	h.mustRefuse(CodeFriendLimit, a, d)
	h.invite(d, a, 0)
	h.mustRefuse(CodeFriendLimit, a, d)

	e := h.user("fa")
	h.invite(a, e, 0)
	fe := h.mustRefuse(CodeFriendTargetLimit, e, a)
	if fe.UserID != a {
		t.Fatalf("target %+v", fe)
	}

	f, g := h.user("fa"), h.user("en")
	h.invite(f, g, 0)
	if _, err := h.s.beforeDeleteFriends(asUser(g), nil, nil, nil, &api.DeleteFriendsRequest{Ids: []string{f}}); err != nil {
		t.Fatal(err)
	}
	h.s.nk.FriendsDelete(context.Background(), g, "", []string{f}, nil)
	fe = h.mustRefuse(CodeFriendCooldown, f, g)
	retry, _ := time.Parse(time.RFC3339, fe.RetryAt)
	if d := time.Until(retry); d < 47*time.Hour || d > 49*time.Hour || fe.UserID != g {
		t.Fatalf("cooldown %+v", fe)
	}
	if _, err := h.s.beforeDeleteFriends(asUser(f), nil, nil, nil, &api.DeleteFriendsRequest{Ids: []string{g}}); err != nil {
		t.Fatal(err)
	}
	var n int
	h.s.db.QueryRow(`SELECT count(*) FROM base_friend_rejection WHERE requester_id=$1`, g).Scan(&n)
	if n != 0 {
		t.Fatal("removing a friend that never invited is not a rejection")
	}
	if err := h.s.afterAddFriends(asUser(g), nil, nil, nil, &api.AddFriendsRequest{Ids: []string{f}}); err != nil {
		t.Fatal(err)
	}
	h.mustAllow(f, g)

	k, m := h.user("fa"), h.user("fa")
	h.invite(k, m, 0)
	h.s.beforeDeleteFriends(asUser(m), nil, nil, nil, &api.DeleteFriendsRequest{Ids: []string{k}})
	h.s.db.Exec(`UPDATE base_friend_rejection SET rejected_at = now() - INTERVAL '49 hours' WHERE requester_id=$1`, k)
	h.mustAllow(k, m)
}

func TestFriendInviteExpiry(t *testing.T) {
	h := newHarness(t)
	resetFriendTables(h)
	h.setFriendRules(0, 0, 0, 0)
	a, b, c, d := h.user("fa"), h.user("fa"), h.user("fa"), h.user("fa")
	h.invite(a, b, 10*24*time.Hour)
	h.invite(c, a, 10*24*time.Hour)
	h.invite(a, d, time.Hour)
	if _, err := h.s.beforeListFriends(asUser(a), nil, nil, nil, &api.ListFriendsRequest{}); err != nil {
		t.Fatal(err)
	}
	count := func(user string) int {
		var n int
		h.s.db.QueryRow(`SELECT count(*) FROM user_edge WHERE source_id=$1 OR destination_id=$1`, user).Scan(&n)
		return n
	}
	if count(a) != 6 {
		t.Fatalf("expiry off must keep invites, have %d", count(a))
	}
	h.setFriendRules(0, 0, 7, 0)
	h.s.beforeListFriends(asUser(a), nil, nil, nil, &api.ListFriendsRequest{})
	if count(b) != 0 || count(c) != 0 || count(d) != 2 {
		t.Fatalf("lazy expiry b=%d c=%d d=%d", count(b), count(c), count(d))
	}

	x, y := h.user("fa"), h.user("fa")
	h.invite(x, y, 8*24*time.Hour)
	if n, err := h.s.friendsSweep(context.Background(), "node-a/1"); err != nil || n < 1 {
		t.Fatalf("sweep n=%d err=%v", n, err)
	}
	if count(x) != 0 {
		t.Fatal("sweeper left an expired invite")
	}
	h.invite(x, y, 8*24*time.Hour)
	if n, err := h.s.friendsSweep(context.Background(), "node-b/2"); err != nil || n != 0 {
		t.Fatalf("a second node must not sweep while the lease is held: n=%d err=%v", n, err)
	}
	if count(x) != 2 {
		t.Fatal("second node swept")
	}
	h.s.db.Exec(`INSERT INTO base_friend_request_day (user_id, day, count) VALUES ($1, current_date - 10, 4)`, x)
	h.s.friendsSweep(context.Background(), "node-a/1")
	var old int
	h.s.db.QueryRow(`SELECT count(*) FROM base_friend_request_day WHERE user_id=$1`, x).Scan(&old)
	if count(x) != 0 || old != 0 {
		t.Fatalf("holder sweep edges=%d counters=%d", count(x), old)
	}
}
