package main

import (
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func newReferralHarness(t *testing.T) (*harness, string, string) {
	h := newEventHarness(t)
	for _, q := range []string{`DELETE FROM economy_referral`, `DELETE FROM economy_referral_mark`, `DELETE FROM economy_referral_code`, `DELETE FROM cluster_rate_limit WHERE key LIKE 'ref:%'`} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for name, fn := range h.s.referralRPCs() {
		h.rpcs[name] = fn
	}
	ref := h.coinPrize("ref_inviter", 50, true)
	ree := h.coinPrize("ref_invited", 30, true)
	h.settings(map[string]any{"referrerPrizeId": ref, "refereePrizeId": ree})
	return h, ref, ree
}

func (h *harness) settings(extra map[string]any) map[string]any {
	h.t.Helper()
	body := map[string]any{"isEnabled": true, "newAccountDays": 7, "goalMeasure": "count", "goalValue": 1, "maxPerReferrer": 100, "dailyPerReferrer": 10, "onePerDevice": true, "onePerPhone": true, "holdFlagged": true}
	cur := h.mustAPI(http.StatusOK, http.MethodGet, "/referral-settings", nil)
	for _, k := range []string{"referrerPrizeId", "refereePrizeId", "goalEventTypeId"} {
		body[k] = cur[k]
	}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(http.StatusOK, http.MethodPut, "/referral-settings", body)
}

func (h *harness) codeOf(user string) string {
	h.t.Helper()
	return h.mustRPC(user, "economy_referral_me", map[string]any{})["code"].(string)
}

func faDigitsOf(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '۰' + (r - '0')
		}
		return r
	}, s)
}

func TestReferralRulesAndAbuse(t *testing.T) {
	h, _, _ := newReferralHarness(t)
	a, b := h.user(), h.user()
	code := h.codeOf(a)
	if len(code) != 8 || h.codeOf(a) != code {
		t.Fatalf("code %q is not stable", code)
	}
	if c := h.rpcCode(b, "economy_referral_enter", map[string]any{"code": h.codeOf(b)}); c != "SELF_REFERRAL" {
		t.Fatal(c)
	}
	if c := h.rpcCode(b, "economy_referral_enter", map[string]any{"code": "NOPE2345"}); c != "CODE_NOT_FOUND" {
		t.Fatal(c)
	}
	out := h.mustRPC(b, "economy_referral_enter", map[string]any{"code": " " + faDigitsOf(strings.ToLower(code)) + " "})
	if out["entered"].(map[string]any)["status"] != "completed" || out["referredBy"] == nil || out["canEnter"] != false {
		t.Fatalf("enter %v", out)
	}
	if h.wallet(a)["coin"] != 50 || h.wallet(b)["coin"] != 30 {
		t.Fatalf("wallets %v %v", h.wallet(a), h.wallet(b))
	}
	if c := h.rpcCode(b, "economy_referral_enter", map[string]any{"code": code}); c != "ALREADY_REFERRED" {
		t.Fatal(c)
	}
	if c := h.rpcCode(a, "economy_referral_enter", map[string]any{"code": h.codeOf(b)}); c != "REFERRAL_CYCLE" {
		t.Fatal(c)
	}
	old := h.user()
	h.s.db.Exec(`UPDATE users SET create_time=now() - INTERVAL '30 days' WHERE id=$1`, old)
	if c := h.rpcCode(old, "economy_referral_enter", map[string]any{"code": code}); c != "ACCOUNT_TOO_OLD" {
		t.Fatal(c)
	}

	h.s.db.Exec(`UPDATE users SET metadata='{"phone":"+989991110000"}' WHERE id=$1`, a)
	twin := h.user()
	h.s.db.Exec(`UPDATE users SET metadata='{"phone":"+989991110000"}' WHERE id=$1`, twin)
	if c := h.rpcCode(twin, "economy_referral_enter", map[string]any{"code": code}); c != "SELF_REFERRAL" {
		t.Fatal(c)
	}
	p1 := h.user()
	h.s.db.Exec(`UPDATE users SET metadata='{"phone":"+989991112222"}' WHERE id=$1`, p1)
	h.mustRPC(p1, "economy_referral_enter", map[string]any{"code": code})
	p2 := h.user()
	h.s.db.Exec(`UPDATE users SET metadata='{"phone":"+989991112222"}' WHERE id=$1`, p2)
	if c := h.rpcCode(p2, "economy_referral_enter", map[string]any{"code": code}); c != "PHONE_ALREADY_USED" {
		t.Fatal(c)
	}
	d1 := h.user()
	h.s.db.Exec(`INSERT INTO user_device (id, user_id) VALUES ('e2e-ref-device-1', $1)`, d1)
	h.mustRPC(d1, "economy_referral_enter", map[string]any{"code": code})
	h.s.db.Exec(`DELETE FROM users WHERE id=$1`, d1)
	d2 := h.user()
	h.s.db.Exec(`INSERT INTO user_device (id, user_id) VALUES ('e2e-ref-device-1', $1)`, d2)
	if c := h.rpcCode(d2, "economy_referral_enter", map[string]any{"code": code}); c != "DEVICE_ALREADY_USED" {
		t.Fatal(c)
	}
	h.s.db.Exec(`DELETE FROM user_device WHERE id='e2e-ref-device-1'`)

	h.settings(map[string]any{"dailyPerReferrer": 2})
	held := h.user()
	res := h.mustRPC(held, "economy_referral_enter", map[string]any{"code": code})["entered"].(map[string]any)
	if res["status"] != "held" || h.wallet(held)["coin"] != 0 {
		t.Fatalf("burst referral %v", res)
	}
	list := h.mustAPI(http.StatusOK, http.MethodGet, "/referrals?status=held&flagged=true", nil)
	items := list["items"].([]any)
	if len(items) != 1 || !strings.Contains(strings.Join(anyStrings(items[0].(map[string]any)["flags"]), ","), "burst") {
		t.Fatalf("held list %v", list)
	}
	rid := items[0].(map[string]any)["id"].(string)
	approved := h.mustAPI(http.StatusOK, http.MethodPost, "/referrals/"+rid+"/approve", map[string]any{"note": "checked"})
	if approved["status"] != "completed" || h.wallet(held)["coin"] != 30 {
		t.Fatalf("approved %v %v", approved, h.wallet(held))
	}
	if code, out := h.api(http.MethodPost, "/referrals/"+rid+"/reject", nil); code != 409 || out["error"] != "REFERRAL_DECIDED" {
		t.Fatalf("reject after approve %d %v", code, out)
	}
	rej := h.user()
	r2 := h.mustRPC(rej, "economy_referral_enter", map[string]any{"code": code})["entered"].(map[string]any)
	h.mustAPI(http.StatusOK, http.MethodPost, "/referrals/"+r2["id"].(string)+"/reject", map[string]any{"note": "same person"})
	if h.wallet(rej)["coin"] != 0 {
		t.Fatal("rejected referral paid")
	}

	h.settings(map[string]any{"dailyPerReferrer": 100, "maxPerReferrer": 3})
	limit := h.user()
	if c := h.rpcCode(limit, "economy_referral_enter", map[string]any{"code": code}); c != "REFERRER_LIMIT" {
		t.Fatalf("referrer limit: %s", c)
	}
	tree := h.mustAPI(http.StatusOK, http.MethodGet, "/referrals/tree/"+a, nil)
	if len(tree["children"].([]any)) != 4 || tree["user"].(map[string]any)["code"] != code {
		t.Fatalf("tree %v", tree)
	}
	sub := h.mustAPI(http.StatusOK, http.MethodGet, "/referrals/tree/"+b, nil)
	if up := sub["referredBy"].([]any); len(up) != 1 || up[0].(map[string]any)["userId"] != a {
		t.Fatalf("upline %v", sub)
	}
	sum := h.mustAPI(http.StatusOK, http.MethodGet, "/referrals/summary", nil)
	if sum["total"].(float64) < 4 || sum["rejected"] != float64(1) || len(sum["topReferrers"].([]any)) == 0 {
		t.Fatalf("summary %v", sum)
	}
	h.settings(map[string]any{"isEnabled": false})
	if c := h.rpcCode(limit, "economy_referral_enter", map[string]any{"code": code}); c != "REFERRAL_DISABLED" {
		t.Fatal(c)
	}
}

func anyStrings(v any) []string {
	out := []string{}
	for _, x := range v.([]any) {
		out = append(out, x.(string))
	}
	return out
}

func TestReferralGoalAndRewardsAreIdempotentUnderConcurrency(t *testing.T) {
	h, _, _ := newReferralHarness(t)
	goal := h.eventType("round_done", true, nil)
	h.settings(map[string]any{"goalEventTypeId": goal, "goalMeasure": "count", "goalValue": 3})
	a := h.user()
	code := h.codeOf(a)
	b := h.user()
	nodes := make([]*service, 4)
	for i := range nodes {
		n := testService()
		n.db, n.nk, n.store, n.now = h.s.db, h.nk, h.s.store, time.Now
		nodes[i] = n
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, dup := 0, 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := nodes[i%4].enterReferral(h.ctx, b, code, "")
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				ok++
			} else if app, isApp := asApp(err); isApp && app.Code == "ALREADY_REFERRED" {
				dup++
			} else {
				t.Errorf("enter: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if ok != 1 || dup != 7 {
		t.Fatalf("concurrent enter ok=%d dup=%d", ok, dup)
	}
	if h.wallet(a)["coin"] != 0 || h.wallet(b)["coin"] != 0 {
		t.Fatal("paid before the goal")
	}
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := 1.0
			_, err := nodes[i%4].recordEvent(h.ctx, eventInput{UserID: b, Key: "round_done", Value: &v, Source: "server"})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if h.wallet(a)["coin"] != 50 || h.wallet(b)["coin"] != 30 {
		t.Fatalf("goal rewards paid once: %v %v", h.wallet(a), h.wallet(b))
	}
	if n := h.count(`SELECT count(*) FROM economy_reward_claim WHERE kind='referral' AND status='granted'`); n != 2 {
		t.Fatalf("referral claims %d", n)
	}
	rows, _ := h.s.db.Query(`SELECT id::text FROM economy_reward_claim WHERE kind='referral'`)
	ids := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(n *service, id string) {
				defer wg.Done()
				_, _ = n.grantClaim(h.ctx, id, 1)
			}(nodes[i], id)
		}
	}
	wg.Wait()
	if h.wallet(a)["coin"] != 50 || h.wallet(b)["coin"] != 30 {
		t.Fatalf("retried grants paid twice: %v %v", h.wallet(a), h.wallet(b))
	}
	me := h.mustRPC(a, "economy_referral_me", map[string]any{})
	if me["counts"].(map[string]any)["completed"] != float64(1) || me["rewards"].(map[string]any)["referrer"] == nil {
		t.Fatalf("referrer view %v", me)
	}
	if n := h.count(`SELECT count(*) FROM base_activity WHERE kind='referral' AND action='completed' AND user_id IN ($1,$2)`, a, b); n != 2 {
		t.Fatalf("completion activity %d", n)
	}
}
