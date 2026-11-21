package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func refusalCode(t *testing.T, err error) (string, map[string]any) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a refusal")
	}
	var rt *runtime.Error
	msg := err.Error()
	if e, ok := err.(*runtime.Error); ok {
		rt = e
		msg = rt.Message
	}
	var e socialError
	if json.Unmarshal([]byte(msg), &e) == nil && e.Code != "" {
		return e.Code, e.Meta
	}
	if app, ok := asApp(err); ok {
		return app.Code, nil
	}
	return msg, nil
}

func (h *harness) currency(key string) {
	h.t.Helper()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_currency (id, key, name, type) VALUES ($1,$2,'{"en":"Gem"}','hard') ON CONFLICT (key) DO NOTHING`, uuid.Must(uuid.NewV4()).String(), key); err != nil {
		h.t.Fatal(err)
	}
	h.t.Cleanup(func() { _, _ = h.s.db.Exec(`DELETE FROM economy_currency WHERE key=$1`, key) })
}

func (h *harness) wallet(user string, w map[string]int64) {
	h.t.Helper()
	raw, _ := json.Marshal(w)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE users SET wallet=$2 WHERE id=$1`, user, string(raw)); err != nil {
		h.t.Fatal(err)
	}
}

func resetSocialSettings(h *harness) {
	for _, q := range []string{`DELETE FROM base_username_settings`, `DELETE FROM base_friend_suggest_settings`, `DELETE FROM base_player_code_settings`,
		`UPDATE base_friend_rules SET notify_declined=true, notify_removed=true, presence_friends_only=false, dm_friends_only=false`} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			h.t.Fatal(err)
		}
	}
	usernameStore.Invalidate()
	suggestStore.Invalidate()
	playerCodeStore.Invalidate()
	h.t.Cleanup(func() {
		for _, q := range []string{`DELETE FROM base_username_settings`, `DELETE FROM base_friend_suggest_settings`, `DELETE FROM base_player_code_settings`,
			`UPDATE base_friend_rules SET notify_declined=true, notify_removed=true, presence_friends_only=false, dm_friends_only=false`} {
			_, _ = h.s.db.Exec(q)
		}
		usernameStore.Invalidate()
		suggestStore.Invalidate()
		playerCodeStore.Invalidate()
	})
}

func TestUsernamePolicyLadderCooldownAdminAndHook(t *testing.T) {
	h := newHarness(t)
	resetSocialSettings(h)
	h.currency("ugem")
	u := h.user("en")

	q := h.mustRPC(u, "base_username_quote", map[string]any{})
	if q["enabled"] != false || q["free"] != true || q["cost"].(float64) != 0 {
		t.Fatalf("quote off %v", q)
	}
	out := h.mustRPC(u, "base_username_change", map[string]any{"username": "plain" + u[:6]})
	if out["source"] != "free" || out["cost"].(float64) != 0 {
		t.Fatalf("change off %v", out)
	}

	h.mustAPI(http.StatusBadRequest, http.MethodPut, "/username-settings", map[string]any{"enabled": true, "currency": "nope_x", "ladder": []int{5}})
	h.mustAPI(http.StatusBadRequest, http.MethodPut, "/username-settings", map[string]any{"enabled": true, "currency": "ugem", "ladder": []int{0}})
	set := h.mustAPI(http.StatusOK, http.MethodPut, "/username-settings", map[string]any{"enabled": true, "firstFree": true, "currency": "ugem", "ladder": []int{5, 10}, "cooldownHours": 0})
	if set["enabled"] != true || len(set["ladder"].([]any)) != 2 {
		t.Fatalf("settings %v", set)
	}

	q = h.mustRPC(u, "base_username_quote", map[string]any{})
	if q["free"] != true || q["cost"].(float64) != 0 {
		t.Fatalf("first change should be free %v", q)
	}
	out = h.mustRPC(u, "base_username_change", map[string]any{"username": "first" + u[:6]})
	if out["source"] != "free" || out["oldUsername"] != "plain"+u[:6] {
		t.Fatalf("first %v", out)
	}
	quote := out["quote"].(map[string]any)
	if quote["free"] != false || quote["cost"].(float64) != 5 {
		t.Fatalf("next quote %v", quote)
	}
	_, err := h.rpc(u, "base_username_change", map[string]any{"username": "second" + u[:6]})
	code, meta := refusalCode(t, err)
	if code != CodeCurrencyNotEnough || meta["amount"].(float64) != 5 || meta["currencyKey"] != "ugem" {
		t.Fatalf("not enough %s %v", code, meta)
	}
	h.wallet(u, map[string]int64{"ugem": 20})
	_, err = h.rpc(u, "base_username_change", map[string]any{"username": "second" + u[:6], "expectedCost": 3})
	if code, meta = refusalCode(t, err); code != CodeUsernamePriceChanged || meta["cost"].(float64) != 5 {
		t.Fatalf("price changed %s %v", code, meta)
	}
	out = h.mustRPC(u, "base_username_change", map[string]any{"username": "second" + u[:6], "expectedCost": 5})
	if out["source"] != "paid" || out["cost"].(float64) != 5 {
		t.Fatalf("paid %v", out)
	}
	var balance int64
	if err := h.s.db.QueryRow(`SELECT (wallet->>'ugem')::BIGINT FROM users WHERE id=$1`, u).Scan(&balance); err != nil || balance != 15 {
		t.Fatalf("balance %d %v", balance, err)
	}
	_, err = h.rpc(u, "base_username_change", map[string]any{"username": "second" + u[:6]})
	if code, _ = refusalCode(t, err); code != CodeUsernameUnchanged {
		t.Fatalf("unchanged %s", code)
	}
	other := h.user("en")
	var otherName string
	_ = h.s.db.QueryRow(`SELECT username FROM users WHERE id=$1`, other).Scan(&otherName)
	_, err = h.rpc(u, "base_username_change", map[string]any{"username": otherName})
	if code, _ = refusalCode(t, err); code != CodeUsernameTaken {
		t.Fatalf("taken %s", code)
	}

	h.mustAPI(http.StatusOK, http.MethodPut, "/username-settings", map[string]any{"enabled": true, "firstFree": true, "currency": "ugem", "ladder": []int{5, 10}, "cooldownHours": 2})
	_, err = h.rpc(u, "base_username_change", map[string]any{"username": "third" + u[:6]})
	if code, meta = refusalCode(t, err); code != CodeUsernameCooldown || meta["cooldownEndsAt"] == "" {
		t.Fatalf("cooldown %s %v", code, meta)
	}

	admin := h.mustAPI(http.StatusOK, http.MethodPost, "/players/"+u+"/username", map[string]any{"username": "staff" + u[:6]})
	if admin["source"] != "admin" || admin["cost"].(float64) != 0 {
		t.Fatalf("admin %v", admin)
	}
	h.expectError(http.StatusBadRequest, CodeUsernameUnchanged, http.MethodPost, "/players/"+u+"/username", map[string]any{"username": "staff" + u[:6]})

	h.mustAPI(http.StatusOK, http.MethodPut, "/username-settings", map[string]any{"enabled": true, "firstFree": true, "currency": "ugem", "ladder": []int{5, 10}, "cooldownHours": 0})
	ctx := asUser(u)
	req, err := h.s.beforeUpdateAccount(ctx, nil, nil, nil, &api.UpdateAccountRequest{Username: wrapperspb.String(" hook" + u[:6] + " ")})
	if err != nil || req.Username.Value != "hook"+u[:6] {
		t.Fatalf("hook %v %v", req, err)
	}
	if err := h.s.db.QueryRow(`SELECT (wallet->>'ugem')::BIGINT FROM users WHERE id=$1`, u).Scan(&balance); err != nil || balance != 5 {
		t.Fatalf("hook should charge 10: %d %v", balance, err)
	}

	hist := h.mustAPI(http.StatusOK, http.MethodGet, "/players/"+u+"/username-changes", nil)
	list := hist["items"].([]any)
	sources := []string{}
	for _, it := range list {
		sources = append(sources, it.(map[string]any)["source"].(string))
	}
	if strings.Join(sources, ",") != "paid,admin,paid,free,free" {
		t.Fatalf("history %v", sources)
	}
	if hist["quote"].(map[string]any)["paidChanges"].(float64) != 2 {
		t.Fatalf("history quote %v", hist["quote"])
	}
	var acts int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND kind='account'`, u).Scan(&acts)
	if acts != 5 {
		t.Fatalf("activity rows %d", acts)
	}
}

func TestPlayerCodesLookupAndFriendTargets(t *testing.T) {
	h := newHarness(t)
	resetSocialSettings(h)
	a, b := h.user("en"), h.user("fa")
	my := h.mustRPC(a, "base_my_code", map[string]any{})
	code := my["code"].(string)
	if !playerCodePattern.MatchString(code) || len(code) != 6 {
		t.Fatalf("code %q", code)
	}
	codes := h.mustRPC(a, "base_player_codes", map[string]any{"ids": []string{a, b}})["codes"].(map[string]any)
	if codes[a] != code || codes[b] == "" || codes[b] == code {
		t.Fatalf("codes %v", codes)
	}
	bcode := codes[b].(string)
	found := h.mustRPC(a, "base_user_by_code", map[string]any{"code": " " + strings.ToLower(bcode) + " "})
	if found["user"].(map[string]any)["id"] != b || found["relationship"] != "none" {
		t.Fatalf("lookup %v", found)
	}
	if _, ok := found["online"]; ok {
		t.Fatalf("online must be hidden for strangers %v", found)
	}
	h.befriend(a, b)
	found = h.mustRPC(a, "base_user_by_code", map[string]any{"code": bcode})
	if found["relationship"] != "friends" || found["online"] != false {
		t.Fatalf("friend lookup %v", found)
	}
	if c := h.rpcCode(a, "base_user_by_code", map[string]any{"code": "ZZZZZZZZZZ"}); c != CodePlayerCodeNotFound {
		t.Fatalf("missing %s", c)
	}
	in, err := h.s.socialTargets(h.ctx, a, `{"codes":["`+strings.ToLower(bcode)+`"]}`)
	if err != nil || len(in.IDs) != 1 || in.IDs[0] != b {
		t.Fatalf("targets %v %v", in, err)
	}
	if _, err := h.s.socialTargets(h.ctx, a, `{"codes":["`+code+`"]}`); err == nil {
		t.Fatal("own code should be refused")
	}
	list := h.mustAPI(http.StatusOK, http.MethodGet, "/players?q="+strings.ToLower(bcode), nil)
	got := items(list)
	if len(got) != 1 || got[0]["id"] != b || got[0]["code"] != bcode {
		t.Fatalf("console search %v", got)
	}
	h.mustAPI(http.StatusBadRequest, http.MethodPut, "/player-codes/settings", map[string]any{"length": 3})
	h.mustAPI(http.StatusOK, http.MethodPut, "/player-codes/settings", map[string]any{"length": 8, "charset": "digits"})
	c := h.user("en")
	var cc string
	_ = h.s.db.QueryRow(`SELECT code FROM base_player_code WHERE user_id=$1`, c).Scan(&cc)
	if len(cc) != 8 || strings.Trim(cc, "0123456789") != "" {
		t.Fatalf("digit code %q", cc)
	}
}

func TestFriendNoticesAndFriendsOnlyRules(t *testing.T) {
	h := newHarness(t)
	resetSocialSettings(h)
	a, b, c := h.user("fa"), h.user("en"), h.user("en")
	h.edge(b, a, edgeSent, 0)
	h.edge(a, b, edgeReceived, 0)
	h.befriend(a, c)
	fake := h.s.nk.(*fakeNakama)
	fake.mu.Lock()
	fake.notes = nil
	fake.mu.Unlock()
	if _, err := h.s.beforeDeleteFriends(asUser(a), nil, nil, nil, &api.DeleteFriendsRequest{Ids: []string{b, c}}); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	notes := append([]fakeNote{}, fake.notes...)
	fake.mu.Unlock()
	got := map[int]string{}
	for _, n := range notes {
		got[n.code] = n.user
		if n.content["userId"] != a || !n.persistent {
			t.Fatalf("note %v", n)
		}
	}
	if got[notifyFriendDeclined] != b || got[notifyFriendRemoved] != c || len(notes) != 2 {
		t.Fatalf("notes %v", notes)
	}
	var skip string
	_ = h.s.db.QueryRow(`SELECT until_cycle FROM base_friend_suggest_skip WHERE user_id=$1 AND candidate_id=$2`, b, a).Scan(&skip)
	if skip == "" {
		t.Fatal("decline should hide the decliner from the requester's suggestions")
	}

	if _, err := h.s.db.Exec(`UPDATE base_friend_rules SET notify_declined=false, notify_removed=false`); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	fake.notes = nil
	fake.mu.Unlock()
	h.befriend(a, c)
	if _, err := h.s.beforeDeleteFriends(asUser(a), nil, nil, nil, &api.DeleteFriendsRequest{Ids: []string{c}}); err != nil {
		t.Fatal(err)
	}
	if len(fake.notes) != 0 {
		t.Fatalf("notices off %v", fake.notes)
	}

	dm := "4." + min(a, b) + "." + max(a, b) + "."
	if err := h.s.dmMessageGate(h.ctx, a, dm); err != nil {
		t.Fatalf("dm open by default %v", err)
	}
	rules := h.mustAPI(http.StatusOK, http.MethodGet, "/friend-rules", nil)
	rules["dmFriendsOnly"], rules["presenceFriendsOnly"] = true, true
	h.mustAPI(http.StatusOK, http.MethodPut, "/friend-rules", rules)
	code, meta := refusalCode(t, h.s.dmMessageGate(h.ctx, a, dm))
	if code != CodeFriendshipNeeded || meta["userId"] != b {
		t.Fatalf("dm gate %s %v", code, meta)
	}
	if _, err := h.s.beforeChannelJoin(asUser(a), nil, nil, nil, &rtapi.Envelope{Message: &rtapi.Envelope_ChannelJoin{ChannelJoin: &rtapi.ChannelJoin{Target: b, Type: 2}}}); err == nil {
		t.Fatal("dm join should be refused")
	}
	h.befriend(a, b)
	if err := h.s.dmMessageGate(h.ctx, a, dm); err != nil {
		t.Fatalf("friends may chat %v", err)
	}
	d := h.user("en")
	env := &rtapi.Envelope{Message: &rtapi.Envelope_StatusFollow{StatusFollow: &rtapi.StatusFollow{UserIds: []string{b, d, "bad"}}}}
	out, err := h.s.beforeStatusFollow(asUser(a), nil, nil, nil, env)
	if err != nil {
		t.Fatal(err)
	}
	if ids := out.GetStatusFollow().UserIds; len(ids) != 1 || ids[0] != b {
		t.Fatalf("follow %v", ids)
	}
	fr := h.mustRPC(a, "base_friend_rules", map[string]any{})
	if fr["dmFriendsOnly"] != true || fr["presenceFriendsOnly"] != true {
		t.Fatalf("client rules %v", fr)
	}
}

func (h *harness) finishedMatch(game string, at time.Time, users ...string) string {
	h.t.Helper()
	id := uuid.Must(uuid.NewV4()).String() + ".nf"
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_game_match (id, game_key, status, players, ended_at) VALUES ($1,$2,'finished',$3,$4)`, id, game, len(users), at); err != nil {
		h.t.Fatal(err)
	}
	for i, u := range users {
		if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_game_entry (match_id, user_id, status, created_at) VALUES ($1,$2,'free',$3)`, id, u, at.Add(time.Duration(i)*time.Millisecond)); err != nil {
			h.t.Fatal(err)
		}
	}
	h.t.Cleanup(func() { _, _ = h.s.db.Exec(`DELETE FROM economy_game_match WHERE id=$1`, id) })
	return id
}

func TestFriendSuggestionsFromCoplayMutualAndPresence(t *testing.T) {
	h := newHarness(t)
	resetSocialSettings(h)
	if _, err := h.s.db.Exec(`UPDATE base_coplay_cursor SET ended_at=now() - INTERVAL '1 minute', match_id=''`); err != nil {
		t.Fatal(err)
	}
	viewer := h.user("en")
	p1, p2, friend, mutual, active, stranger := h.user("en"), h.user("en"), h.user("en"), h.user("en"), h.user("en"), h.user("en")
	now := time.Now().UTC()
	h.finishedMatch("ludo", now.Add(-30*time.Second), viewer, p1, p2)
	h.finishedMatch("ludo", now.Add(-20*time.Second), viewer, friend)
	h.befriend(viewer, friend)
	h.befriend(friend, mutual)
	day, _ := dayBounds(now, projectZone(h.ctx, h.s.db))
	for _, u := range []string{viewer, active} {
		if _, err := h.s.db.Exec(`INSERT INTO base_presence_day (user_id, day) VALUES ($1,$2::date) ON CONFLICT DO NOTHING`, u, day); err != nil {
			t.Fatal(err)
		}
	}
	n, err := h.s.recordCoplay(h.ctx, 100)
	if err != nil || n < 2 {
		t.Fatalf("recorded %d %v", n, err)
	}
	var matches int
	_ = h.s.db.QueryRow(`SELECT matches FROM base_coplay WHERE user_id=$1 AND other_id=$2`, viewer, p1).Scan(&matches)
	if matches != 1 {
		t.Fatalf("coplay matches %d", matches)
	}
	if again, _ := h.s.recordCoplay(h.ctx, 100); again != 0 {
		t.Fatalf("cursor should skip recorded matches, got %d", again)
	}

	out := h.mustRPC(viewer, "base_friend_suggestions", map[string]any{})
	list := out["suggestions"].([]any)
	ids := []string{}
	src := map[string]string{}
	for _, it := range list {
		m := it.(map[string]any)
		ids = append(ids, m["userId"].(string))
		src[m["userId"].(string)] = m["source"].(string)
		if m["code"] == "" {
			t.Fatalf("suggestion without code %v", m)
		}
	}
	if len(ids) < 4 || src[p1] != "coplay" || src[p2] != "coplay" || src[mutual] != "mutual" || src[active] != "activity" {
		t.Fatalf("suggestions %v %v", ids, src)
	}
	if ids[0] != p1 && ids[0] != p2 {
		t.Fatalf("co-players first %v", ids)
	}
	for _, id := range ids {
		if id == friend || id == viewer || id == stranger {
			t.Fatalf("excluded user suggested %v", ids)
		}
	}
	var shows int
	_ = h.s.db.QueryRow(`SELECT shows FROM base_friend_suggest WHERE user_id=$1 AND candidate_id=$2`, viewer, p1).Scan(&shows)
	h.mustRPC(viewer, "base_friend_suggestions", map[string]any{})
	var after int
	_ = h.s.db.QueryRow(`SELECT shows FROM base_friend_suggest WHERE user_id=$1 AND candidate_id=$2`, viewer, p1).Scan(&after)
	if shows != 1 || after != 1 {
		t.Fatalf("one impression per cycle: %d %d", shows, after)
	}

	late := h.user("en")
	h.finishedMatch("ludo", time.Now().UTC(), viewer, late)
	if _, err := h.s.recordCoplay(h.ctx, 100); err != nil {
		t.Fatal(err)
	}
	out = h.mustRPC(viewer, "base_friend_suggestions", map[string]any{})
	if first := out["suggestions"].([]any)[0].(map[string]any); first["userId"] != late || first["source"] != "coplay" {
		t.Fatalf("live patch puts the new co-player first: %v", first)
	}

	h.s.dropSuggestionPairs(h.ctx, viewer, []string{p1})
	var gone bool
	_ = h.s.db.QueryRow(`SELECT NOT EXISTS (SELECT 1 FROM base_friend_suggest WHERE user_id=$1 AND candidate_id=$2)`, viewer, p1).Scan(&gone)
	var slots string
	_ = h.s.db.QueryRow(`SELECT string_agg(slot::text, ',' ORDER BY slot) FROM base_friend_suggest WHERE user_id=$1`, viewer).Scan(&slots)
	if !gone || !strings.HasPrefix(slots, "1,2,3") {
		t.Fatalf("drop and compact %v %s", gone, slots)
	}

	admin := h.mustAPI(http.StatusOK, http.MethodGet, "/players/"+viewer+"/friend-suggestions", nil)
	if len(admin["visible"].([]any)) == 0 || len(admin["coplay"].([]any)) < 3 || admin["cycle"] != day {
		t.Fatalf("admin view %v", admin)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/friend-suggestions/settings", map[string]any{"visible": 2, "reserve": 1})
	regen := h.mustAPI(http.StatusOK, http.MethodPost, "/players/"+viewer+"/friend-suggestions/regenerate", nil)
	if len(regen["visible"].([]any)) != 2 || len(regen["reserve"].([]any)) != 1 || regen["generatedBy"] == "" {
		t.Fatalf("regenerate %v", regen)
	}
	h.mustAPI(http.StatusBadRequest, http.MethodPut, "/friend-suggestions/settings", map[string]any{"visible": 0})
	h.mustAPI(http.StatusOK, http.MethodPut, "/friend-suggestions/settings", map[string]any{"enabled": false})
	if off := h.mustRPC(viewer, "base_friend_suggestions", map[string]any{}); off["enabled"] != false {
		t.Fatalf("off %v", off)
	}
}

func TestRankSuggestionsQuotasDecayAndFill(t *testing.T) {
	v := defaultSuggestSettings()
	v.Visible, v.Reserve, v.CoplayQuota, v.MutualQuota, v.ActivityQuota, v.LevelQuota, v.DecayAfter = 3, 1, 1, 1, 1, 0, 1
	pools := map[string][]suggestRow{
		"coplay":   {{Candidate: "a", Source: "coplay", Score: 10}, {Candidate: "b", Source: "coplay", Score: 9}},
		"mutual":   {{Candidate: "a", Source: "mutual", Score: 5}, {Candidate: "c", Source: "mutual", Score: 1}, {Candidate: "d", Source: "mutual", Score: 4}},
		"activity": {{Candidate: "e", Source: "activity", Score: 2}},
	}
	shows := map[string]int{"d": 3}
	got := rankSuggestions(v, pools, func(string) bool { return true }, shows, map[string]bool{}, nil)
	ids := []string{}
	for _, r := range got {
		ids = append(ids, r.Candidate)
	}
	if strings.Join(ids, ",") != "a,c,e,b" {
		t.Fatalf("rank %v", ids)
	}
	if suggestDecay(5, 4) != 0.5 || suggestDecay(4, 4) != 1 {
		t.Fatal("decay")
	}
	if activityBucket(7, 7) != 3 || activityBucket(4, 7) != 2 || activityBucket(1, 7) != 1 || activityBucket(0, 7) != 0 {
		t.Fatal("buckets")
	}
}
