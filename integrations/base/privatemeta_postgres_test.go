package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func noPhoneIn(t *testing.T, what string, out any, digits string) {
	t.Helper()
	b, _ := json.Marshal(out)
	if strings.Contains(string(b), digits) || strings.Contains(string(b), `\"phone\"`) || strings.Contains(string(b), `"phone"`) {
		t.Fatalf("%s leaks the phone: %s", what, b)
	}
}

func TestOtherPlayersNeverSeePhoneInMetadata(t *testing.T) {
	h := newHarness(t)
	resetFriendTables(h)
	resetSocialSettings(h)
	registerSocialHooks(h)
	if _, err := h.s.db.Exec(`UPDATE base_coplay_cursor SET ended_at=now() - INTERVAL '1 minute', match_id=''`); err != nil {
		t.Fatal(err)
	}
	a, b, c := h.user("fa"), h.user("fa"), h.user("fa")
	digits := fmt.Sprintf("999%07d", rand.Intn(10000000))
	phone := "+98" + digits
	if _, err := h.s.db.Exec(`UPDATE users SET metadata = jsonb_build_object('phone', $2::text, 'mobile', $2::text, 'email', 'a@example.com', 'country', 'IR') WHERE id=$1`, a, phone); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO publication_phone_account (phone, user_id) VALUES ($1,$2)`, phone, a); err != nil {
		t.Fatal(err)
	}

	got := h.mustSocial(b, "base_users_get", map[string]any{"ids": []string{a}, "usernames": []string{mustName(h, a)}})
	noPhoneIn(t, "base_users_get", got, digits)
	users := got["users"].([]any)
	if len(users) != 1 {
		t.Fatalf("users %v", got)
	}
	md := map[string]any{}
	_ = json.Unmarshal([]byte(users[0].(map[string]any)["metadata"].(string)), &md)
	if md["country"] != "IR" || len(md) != 1 {
		t.Fatalf("other metadata keys must stay: %v", md)
	}

	noPhoneIn(t, "base_friends_add", h.mustSocial(b, "base_friends_add", map[string]any{"ids": []string{a}}), digits)
	noPhoneIn(t, "base_friends_add accept", h.mustSocial(a, "base_friends_add", map[string]any{"ids": []string{b}}), digits)
	list := h.mustSocial(b, "base_friends_list", map[string]any{})
	if friendRows(list)[a] != edgeFriend {
		t.Fatalf("friends %v", list)
	}
	noPhoneIn(t, "base_friends_list", list, digits)
	noPhoneIn(t, "base_account_get", h.mustSocial(a, "base_account_get", map[string]any{}), digits)

	code := h.mustRPC(a, "base_my_code", map[string]any{})["code"].(string)
	noPhoneIn(t, "base_user_by_code", h.mustRPC(b, "base_user_by_code", map[string]any{"code": code}), digits)
	noPhoneIn(t, "base_player_codes", h.mustRPC(b, "base_player_codes", map[string]any{"ids": []string{a}}), digits)

	h.finishedMatch("ludo", time.Now().UTC().Add(-10*time.Second), c, a)
	if _, err := h.s.recordCoplay(h.ctx, 100); err != nil {
		t.Fatal(err)
	}
	sugg := h.mustRPC(c, "base_friend_suggestions", map[string]any{})
	found := false
	for _, it := range sugg["suggestions"].([]any) {
		if it.(map[string]any)["userId"] == a {
			found = true
		}
	}
	if !found {
		t.Fatalf("player A should be suggested: %v", sugg)
	}
	noPhoneIn(t, "base_friend_suggestions", sugg, digits)

	card := h.mustAPI(http.StatusOK, http.MethodGet, "/players/"+a, nil)
	if card["phone"] != phone || card["phoneVerifiedAt"] == nil {
		t.Fatalf("staff must still see the phone: %v", card)
	}
	if other := h.mustAPI(http.StatusOK, http.MethodGet, "/players/"+b, nil); other["phone"] != "" {
		t.Fatalf("player without phone: %v", other)
	}
	for _, q := range []string{"0" + digits[:3] + " " + digits[3:], phone, "۰" + strings.Map(func(r rune) rune { return '۰' + (r - '0') }, digits)} {
		res := h.mustAPI(http.StatusOK, http.MethodGet, "/players?q="+url.QueryEscape(q), nil)
		items, _ := res["items"].([]any)
		if len(items) != 1 || items[0].(map[string]any)["id"] != a {
			t.Fatalf("search %q = %v", q, res)
		}
	}
}
