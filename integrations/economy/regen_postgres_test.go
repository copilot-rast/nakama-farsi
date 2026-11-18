package main

import (
	"database/sql"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

func newRegenHarness(t *testing.T, regen map[string]any) (*harness, string, *time.Time) {
	h := newHarness(t)
	for name, fn := range h.s.regenRPCs() {
		h.rpcs[name] = fn
	}
	clock := time.Now().UTC().Truncate(time.Second)
	h.s.now = func() time.Time { return clock }
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس"}, "type": "hard"})
	body := map[string]any{"key": "heart", "name": map[string]string{"fa": "جان", "en": "Heart"}, "type": "soft", "defaultValue": 5, "regen": regen}
	id := h.mustAPI(201, http.MethodPost, "/currencies", body)["id"].(string)
	return h, id, &clock
}

func defaultRegen() map[string]any {
	return map[string]any{"enabled": true, "max": 5, "intervalSeconds": 600, "refillPrice": map[string]any{"currencyKey": "gem", "amount": 10}}
}

func (h *harness) regenState(user string) map[string]any {
	h.t.Helper()
	out := h.mustRPC(user, "economy_regen_state", map[string]any{"currencies": []string{"heart"}})
	list := out["currencies"].([]any)
	if len(list) != 1 {
		h.t.Fatalf("state = %v", out)
	}
	return list[0].(map[string]any)
}

func num(v any) int64 {
	f, _ := v.(float64)
	return int64(f)
}

func TestRegenMathRefillAndRefusal(t *testing.T) {
	h, _, clock := newRegenHarness(t, defaultRegen())
	t0 := *clock
	user := h.user()
	st := h.regenState(user)
	if num(st["value"]) != 5 || st["nextAt"] != nil || st["fullAt"] != nil || num(st["max"]) != 5 || num(st["intervalSeconds"]) != 600 || st["regen"] != true {
		t.Fatalf("full state = %v", st)
	}
	sp, err := h.s.spendRegen(h.ctx, user, "heart", 3, "run-1")
	if err != nil || sp.Spent != 3 || sp.View.Value != 2 || sp.Replayed {
		t.Fatalf("spend = %+v %v", sp, err)
	}
	if !sp.View.NextAt.Equal(t0.Add(600*time.Second)) || !sp.View.FullAt.Equal(t0.Add(1800*time.Second)) {
		t.Fatalf("timers = %v %v", sp.View.NextAt, sp.View.FullAt)
	}
	again, err := h.s.spendRegen(h.ctx, user, "heart", 3, "run-1")
	if err != nil || !again.Replayed || again.View.Value != 2 || h.wallet(user)["heart"] != 2 {
		t.Fatalf("replay = %+v %v %v", again, err, h.wallet(user))
	}
	*clock = t0.Add(1199 * time.Second)
	st = h.regenState(user)
	if num(st["value"]) != 3 || st["nextAt"] != t0.Add(1200*time.Second).Format(time.RFC3339) {
		t.Fatalf("after 1199s = %v", st)
	}
	*clock = t0.Add(1200 * time.Second)
	if st = h.regenState(user); num(st["value"]) != 4 {
		t.Fatalf("after 1200s = %v", st)
	}
	*clock = t0.Add(5 * time.Hour)
	if st = h.regenState(user); num(st["value"]) != 5 || st["nextAt"] != nil {
		t.Fatalf("capped = %v", st)
	}
	if n := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'reason'='regen'`, user); n != 3 {
		t.Fatalf("regen ledger rows = %d", n)
	}
	if _, err := h.s.spendRegen(h.ctx, user, "heart", 5, "run-2"); err != nil {
		t.Fatal(err)
	}
	_, err = h.s.spendRegen(h.ctx, user, "heart", 1, "run-3")
	r, ok := economyhooks.AsRefusal(err)
	if !ok || r.Code != "NOT_ENOUGH_CURRENCY" || r.Meta["nextAt"] == nil || r.Status != http.StatusConflict {
		t.Fatalf("refusal = %v", err)
	}
	if h.count(`SELECT count(*) FROM economy_regen_spend WHERE user_id=$1 AND source_id='run-3'`, user) != 0 {
		t.Fatal("refused spend recorded")
	}
	if code := h.rpcCode(user, "economy_regen_refill", map[string]any{"currencyKey": "heart", "requestId": "r1"}); code != "INSUFFICIENT_BALANCE" {
		t.Fatalf("refill without gems = %s", code)
	}
	if code := h.rpcCode(user, "economy_regen_refill", map[string]any{"currencyKey": "heart"}); code != "REQUEST_ID_REQUIRED" {
		t.Fatalf("refill without request = %s", code)
	}
	if _, _, err := h.nk.WalletUpdate(h.ctx, user, map[string]int64{"gem": 15}, map[string]any{"reason": "test"}, true); err != nil {
		t.Fatal(err)
	}
	out := h.mustRPC(user, "economy_regen_refill", map[string]any{"currencyKey": "heart", "requestId": "r1"})
	if num(out["filled"]) != 5 || out["replayed"] != false || h.wallet(user)["heart"] != 5 || h.wallet(user)["gem"] != 5 {
		t.Fatalf("refill = %v %v", out, h.wallet(user))
	}
	out = h.mustRPC(user, "economy_regen_refill", map[string]any{"currencyKey": "heart", "requestId": "r1"})
	if out["replayed"] != true || h.wallet(user)["gem"] != 5 {
		t.Fatalf("refill replay = %v %v", out, h.wallet(user))
	}
	if code := h.rpcCode(user, "economy_regen_refill", map[string]any{"currencyKey": "heart", "requestId": "r2"}); code != "ALREADY_FULL" {
		t.Fatalf("refill when full = %s", code)
	}
	if code := h.rpcCode(user, "economy_regen_refill", map[string]any{"currencyKey": "gem", "requestId": "r3"}); code != "REGEN_OFF" {
		t.Fatalf("refill of a plain currency = %s", code)
	}
	bal := h.mustRPC(user, "economy_balances", map[string]any{})
	if list, _ := bal["regen"].([]any); len(list) != 1 || list[0].(map[string]any)["currencyKey"] != "heart" {
		t.Fatalf("balances regen = %v", bal["regen"])
	}
	cat := h.mustRPC(user, "economy_catalog", map[string]any{})
	for _, raw := range cat["currencies"].([]any) {
		c := raw.(map[string]any)
		if c["key"] == "heart" && (c["regen"] == nil || num(c["regen"].(map[string]any)["max"]) != 5) {
			t.Fatalf("catalog regen = %v", c)
		}
		if c["key"] == "gem" && c["regen"] != nil {
			t.Fatalf("gem regen = %v", c)
		}
	}
}

func TestRegenRefundIsIdempotentAndCapped(t *testing.T) {
	h, _, clock := newRegenHarness(t, defaultRegen())
	t0 := *clock
	user := h.user()
	h.regenState(user)
	if _, err := h.s.spendRegen(h.ctx, user, "heart", 2, "a"); err != nil {
		t.Fatal(err)
	}
	rf, err := h.s.refundRegen(h.ctx, user, "heart", "a")
	if err != nil || !rf.Found || rf.Refunded != 2 || rf.View.Value != 5 || rf.Replayed {
		t.Fatalf("refund = %+v %v", rf, err)
	}
	rf, err = h.s.refundRegen(h.ctx, user, "heart", "a")
	if err != nil || !rf.Replayed || rf.View.Value != 5 || h.wallet(user)["heart"] != 5 {
		t.Fatalf("refund replay = %+v %v %v", rf, err, h.wallet(user))
	}
	if _, err := h.s.spendRegen(h.ctx, user, "heart", 1, "b"); err != nil {
		t.Fatal(err)
	}
	*clock = t0.Add(601 * time.Second)
	rf, err = h.s.refundRegen(h.ctx, user, "heart", "b")
	if err != nil || rf.Refunded != 0 || rf.View.Value != 5 {
		t.Fatalf("capped refund = %+v %v", rf, err)
	}
	rf, err = h.s.refundRegen(h.ctx, user, "heart", "never")
	if err != nil || rf.Found {
		t.Fatalf("unknown source = %+v %v", rf, err)
	}
	if _, err := h.s.refundRegen(h.ctx, user, "gem", "b"); err == nil {
		t.Fatal("refund with another currency accepted")
	}
	if n := h.count(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'reason'='regen_refund'`, user); n != 1 {
		t.Fatalf("refund ledger rows = %d", n)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.s.spendRegen(h.ctx, user, "heart", 1, "parallel")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("parallel spend: %v", err)
		}
	}
	if h.wallet(user)["heart"] != 4 {
		t.Fatalf("parallel spends with one source took %v", h.wallet(user))
	}
}

func TestRegenUnlimitedTagAndPlainCurrency(t *testing.T) {
	cfg := defaultRegen()
	h, id, clock := newRegenHarness(t, cfg)
	tag := h.taxonomy("base_tag", "user", "unlimited_hearts", map[string]string{"fa": "جان بی‌نهایت"})["id"].(string)
	cfg["unlimitedTag"] = "unlimited_hearts"
	h.mustAPI(200, http.MethodPut, "/currencies/"+id, map[string]any{"key": "heart", "name": map[string]string{"fa": "جان"}, "type": "soft", "defaultValue": 5, "regen": cfg})
	user := h.user()
	h.regenState(user)
	*clock = time.Now().UTC()
	if err := h.s.inTx(h.ctx, func(tx *sql.Tx) error {
		return applyTag(h.ctx, tx, user, prizeItem{Type: "tag", Key: "unlimited_hearts", Action: "give", Value: []byte("0.0208")}, "test")
	}); err != nil {
		t.Fatal(err)
	}
	st := h.regenState(user)
	if st["unlimited"] != true || st["unlimitedUntil"] == nil {
		t.Fatalf("unlimited state = %v", st)
	}
	sp, err := h.s.spendRegen(h.ctx, user, "heart", 1, "free-run")
	if err != nil || !sp.Free || sp.Spent != 0 || h.wallet(user)["heart"] != 5 {
		t.Fatalf("free spend = %+v %v %v", sp, err, h.wallet(user))
	}
	if rf, err := h.s.refundRegen(h.ctx, user, "heart", "free-run"); err != nil || rf.Refunded != 0 || h.wallet(user)["heart"] != 5 {
		t.Fatalf("free refund = %+v %v", rf, err)
	}
	*clock = clock.Add(31 * time.Minute)
	sp, err = h.s.spendRegen(h.ctx, user, "heart", 1, "paid-run")
	if err != nil || sp.Free || sp.Spent != 1 || h.wallet(user)["heart"] != 4 {
		t.Fatalf("after expiry = %+v %v %v", sp, err, h.wallet(user))
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE base_user_tag SET expires_at=NULL WHERE user_id=$1 AND tag_id=$2`, user, tag); err != nil {
		t.Fatal(err)
	}
	if st = h.regenState(user); st["unlimited"] != true || st["unlimitedUntil"] != nil {
		t.Fatalf("permanent tag = %v", st)
	}
	if _, _, err := h.nk.WalletUpdate(h.ctx, user, map[string]int64{"gem": 3}, map[string]any{"reason": "test"}, true); err != nil {
		t.Fatal(err)
	}
	sp, err = h.s.spendRegen(h.ctx, user, "gem", 2, "plain")
	if err != nil || sp.Spent != 2 || sp.View.Regen || h.wallet(user)["gem"] != 1 {
		t.Fatalf("plain spend = %+v %v", sp, err)
	}
	if _, err := h.s.spendRegen(h.ctx, user, "gem", 2, "plain-2"); err == nil {
		t.Fatal("plain overspend accepted")
	}
	if rf, err := h.s.refundRegen(h.ctx, user, "gem", "plain"); err != nil || rf.Refunded != 2 || h.wallet(user)["gem"] != 3 {
		t.Fatalf("plain refund = %+v %v", rf, err)
	}
}

func TestRegenSettingsClampAndOutsideDebit(t *testing.T) {
	h, id, clock := newRegenHarness(t, defaultRegen())
	t0 := *clock
	base := map[string]any{"key": "heart", "name": map[string]string{"fa": "جان"}, "type": "soft", "defaultValue": 5}
	put := func(regen any) (int, map[string]any) {
		b := map[string]any{}
		for k, v := range base {
			b[k] = v
		}
		b["regen"] = regen
		return h.api(http.MethodPut, "/currencies/"+id, b)
	}
	for _, tc := range []struct {
		regen map[string]any
		code  string
	}{
		{map[string]any{"enabled": true, "max": 0, "intervalSeconds": 60}, "INVALID_REGEN_MAX"},
		{map[string]any{"enabled": true, "max": 5, "intervalSeconds": 0}, "INVALID_REGEN_INTERVAL"},
		{map[string]any{"enabled": true, "max": 5, "intervalSeconds": 60, "refillPrice": map[string]any{"currencyKey": "heart", "amount": 1}}, "REFILL_SAME_CURRENCY"},
		{map[string]any{"enabled": true, "max": 5, "intervalSeconds": 60, "refillPrice": map[string]any{"currencyKey": "nope", "amount": 1}}, "CURRENCY_NOT_FOUND"},
		{map[string]any{"enabled": true, "max": 5, "intervalSeconds": 60, "unlimitedTag": "missing"}, "TAG_NOT_FOUND"},
	} {
		if code, out := put(tc.regen); code != 400 || out["error"] != tc.code {
			t.Fatalf("%v = %d %v", tc.regen, code, out)
		}
	}
	h.mustAPI(200, http.MethodPut, "/currencies/"+id, base)
	got := h.mustAPI(200, http.MethodGet, "/currencies/"+id, nil)
	if r, _ := got["regen"].(map[string]any); r == nil || num(r["max"]) != 5 {
		t.Fatalf("save without regen dropped it: %v", got["regen"])
	}
	user := h.user()
	h.regenState(user)
	if _, _, err := h.nk.WalletUpdate(h.ctx, user, map[string]int64{"heart": -2}, map[string]any{"reason": "test"}, true); err != nil {
		t.Fatal(err)
	}
	*clock = t0.Add(2 * time.Hour)
	if st := h.regenState(user); num(st["value"]) != 3 {
		t.Fatalf("outside debit refilled at once: %v", st)
	}
	*clock = t0.Add(2*time.Hour + 600*time.Second)
	if st := h.regenState(user); num(st["value"]) != 4 {
		t.Fatalf("one step after outside debit: %v", st)
	}
	if _, _, err := h.nk.WalletUpdate(h.ctx, user, map[string]int64{"heart": 6}, map[string]any{"reason": "test"}, true); err != nil {
		t.Fatal(err)
	}
	if st := h.regenState(user); num(st["value"]) != 10 {
		t.Fatalf("grant above max without clamp: %v", st)
	}
	clamp := defaultRegen()
	clamp["clamp"] = true
	if code, out := put(clamp); code != 200 {
		t.Fatalf("clamp = %d %v", code, out)
	}
	if st := h.regenState(user); num(st["value"]) != 5 || h.wallet(user)["heart"] != 5 {
		t.Fatalf("clamped = %v", st)
	}
	if code, _ := put(nil); code != 200 {
		t.Fatal("clear regen")
	}
	if st := h.regenState(user); st["regen"] != false {
		t.Fatalf("cleared = %v", st)
	}
	list := h.mustAPI(200, http.MethodGet, "/currencies?regen=false", nil)
	if items, _ := list["items"].([]any); len(items) != 2 {
		t.Fatalf("regen filter = %v", list)
	}
}
