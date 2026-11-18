package main

import (
	"net/http"
	"testing"
)

func newPaymentHarness(t *testing.T) *harness {
	h := newEventHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_payment_settings`); err != nil {
		t.Fatal(err)
	}
	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true, "packageName": "ir.game"}})
	return h
}

func (h *harness) storeProduct(key string, oneTime bool) string {
	h.t.Helper()
	prize := h.coinPrize(key+"_coins", 100, true)
	p := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": key, "isActive": true, "name": map[string]string{"fa": key}, "prizeId": prize,
		"paymentMethods": []string{"bazaar"}, "priceToman": 20000, "sku": map[string]string{"bazaar": key + ".sku"}, "isOneTime": oneTime})
	return p["id"].(string)
}

func (h *harness) storeBuy(user, product string) (string, map[string]any) {
	h.t.Helper()
	id := h.mustRPC(user, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"})["paymentId"].(string)
	return id, h.mustRPC(user, "economy_purchase_submit", map[string]any{"paymentId": id, "token": "sandbox-" + id + "-t"})
}

func (h *harness) ageVerified(id string) {
	h.t.Helper()
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_payment SET verified_at=now()-interval '5 minutes', grant_retry_at=NULL WHERE id=$1`, id); err != nil {
		h.t.Fatal(err)
	}
}

func TestPaymentRewardRetryAfterFailedGrant(t *testing.T) {
	h := newPaymentHarness(t)
	product := h.storeProduct("pack", false)
	bob := h.user()
	start := h.wallet(bob)["coin"]

	h.nk.failWallet = true
	id, sub := h.storeBuy(bob, product)
	if sub["status"] != "success" || sub["rewardGranted"] != false || sub["error"] != "REWARD_PENDING" {
		t.Fatalf("failed grant answer = %v", sub)
	}
	again := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": id, "token": "sandbox-" + id + "-t"})
	if again["rewardGranted"] != false || again["error"] != "REWARD_PENDING" {
		t.Fatalf("resubmit while pending = %v", again)
	}
	list := h.mustAPI(200, http.MethodGet, "/payments?reward=pending", nil)
	if list["total"].(float64) != 1 {
		t.Fatalf("console pending list = %v", list)
	}
	if n, err := h.s.rewardRetryBatch(h.ctx); err != nil || n != 0 {
		t.Fatalf("retry before the grace time = %d %v", n, err)
	}
	h.ageVerified(id)
	if n, err := h.s.rewardRetryBatch(h.ctx); err != nil || n != 1 {
		t.Fatalf("retry while still failing = %d %v", n, err)
	}
	if p := h.payment(id); p.RewardGranted {
		t.Fatal("granted while the wallet was down")
	}
	if n, _ := h.s.rewardRetryBatch(h.ctx); n != 0 {
		t.Fatal("retry ignored the backoff")
	}

	h.nk.failWallet = false
	h.ageVerified(id)
	if n, err := h.s.rewardRetryBatch(h.ctx); err != nil || n != 1 {
		t.Fatalf("retry = %d %v", n, err)
	}
	if p := h.payment(id); !p.RewardGranted || h.wallet(bob)["coin"] != start+100 {
		t.Fatalf("after retry = %+v wallet %v", p, h.wallet(bob))
	}
	h.ageVerified(id)
	if n, _ := h.s.rewardRetryBatch(h.ctx); n != 0 || h.wallet(bob)["coin"] != start+100 {
		t.Fatal("granted twice")
	}
	if code, out := h.api(http.MethodPost, "/payments/"+id+"/retry-grant", nil); code != 409 || out["error"] != "REWARD_ALREADY_GRANTED" {
		t.Fatalf("console retry after grant = %d %v", code, out)
	}
	st := h.mustRPC(bob, "economy_purchase_status", map[string]any{"paymentId": id})
	if st["rewardGranted"] != true {
		t.Fatalf("status = %v", st)
	}

	h.nk.failWallet = true
	id2, _ := h.storeBuy(bob, product)
	h.nk.failWallet = false
	h.mustRPC(bob, "economy_balances", map[string]any{})
	if p := h.payment(id2); !p.RewardGranted || h.wallet(bob)["coin"] != start+200 {
		t.Fatalf("grant on next call = %+v wallet %v", p, h.wallet(bob))
	}

	h.mustAPI(200, http.MethodPut, "/payment-settings", map[string]any{"rewardRetry": false, "grantOnNextCall": false})
	h.nk.failWallet = true
	id3, _ := h.storeBuy(bob, product)
	h.nk.failWallet = false
	h.ageVerified(id3)
	h.mustRPC(bob, "economy_balances", map[string]any{})
	if n, _ := h.s.rewardRetryBatch(h.ctx); n != 0 || h.payment(id3).RewardGranted {
		t.Fatal("retry ran while turned off")
	}
	h.mustAPI(200, http.MethodPost, "/payments/"+id3+"/retry-grant", nil)
	if !h.payment(id3).RewardGranted || h.wallet(bob)["coin"] != start+300 {
		t.Fatal("console retry did not grant")
	}
	if code, out := h.api(http.MethodPut, "/payment-settings", map[string]any{"paidConflict": "nope"}); code != 400 || out["field"] != "paidConflict" {
		t.Fatalf("bad setting = %d %v", code, out)
	}
}

func TestPaymentPaidConflictRecordedForRefund(t *testing.T) {
	h := newPaymentHarness(t)
	product := h.storeProduct("starter", true)
	bob := h.user()
	a, _ := h.storeBuy(bob, product)
	if !h.payment(a).RewardGranted {
		t.Fatal("first one-time purchase not granted")
	}
	dup := "00000000-0000-4000-8000-00000000d0b1"
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, sandbox, status, original_toman, final_toman, amount_rial)
		SELECT $2, user_id, product_id, method, sandbox, 'pending', original_toman, final_toman, amount_rial FROM economy_payment WHERE id=$1`, a, dup); err != nil {
		t.Fatal(err)
	}
	sub := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": dup, "token": "sandbox-" + dup + "-t"})
	if sub["rewardGranted"] != false || sub["error"] != "PRODUCT_ALREADY_BOUGHT" {
		t.Fatalf("duplicate answer = %v", sub)
	}
	var refund string
	_ = h.s.db.QueryRow(`SELECT COALESCE(refund_status,'') FROM economy_payment WHERE id=$1`, dup).Scan(&refund)
	if refund != "needed" {
		t.Fatalf("refund status = %q", refund)
	}
	if list := h.mustAPI(200, http.MethodGet, "/payments?reward=refund", nil); list["total"].(float64) != 1 {
		t.Fatalf("refund list = %v", list)
	}
	if list := h.mustAPI(200, http.MethodGet, "/payments?reward=pending", nil); list["total"].(float64) != 0 {
		t.Fatalf("refund payment still in the pending list = %v", list)
	}
	if code, out := h.api(http.MethodPost, "/payments/"+dup+"/refund", map[string]any{"status": "refunded"}); code != 400 || out["error"] != "NOTE_REQUIRED" {
		t.Fatalf("refund without note = %d %v", code, out)
	}
	out := h.mustAPI(200, http.MethodPost, "/payments/"+dup+"/refund", map[string]any{"status": "refunded", "note": "returned in Bazaar panel"})
	if out["refundStatus"] != "refunded" || out["refundNote"] != "returned in Bazaar panel" {
		t.Fatalf("refund detail = %v", out)
	}
	if code, _ := h.api(http.MethodPost, "/payments/"+a+"/refund", map[string]any{"status": "refunded", "note": "x"}); code != 409 {
		t.Fatal("a delivered payment was marked refunded")
	}

	h.mustAPI(200, http.MethodPut, "/payment-settings", map[string]any{"paidConflict": "keep"})
	dup2 := "00000000-0000-4000-8000-00000000d0b2"
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, sandbox, status, original_toman, final_toman, amount_rial)
		SELECT $2, user_id, product_id, method, sandbox, 'pending', original_toman, final_toman, amount_rial FROM economy_payment WHERE id=$1`, a, dup2); err != nil {
		t.Fatal(err)
	}
	h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": dup2, "token": "sandbox-" + dup2 + "-t"})
	_ = h.s.db.QueryRow(`SELECT COALESCE(refund_status,'') FROM economy_payment WHERE id=$1`, dup2).Scan(&refund)
	if refund != "kept" {
		t.Fatalf("keep policy = %q", refund)
	}
}

func TestStoreOneTimeOnceKeyAtStart(t *testing.T) {
	h := newPaymentHarness(t)
	product := h.storeProduct("vip", true)
	bob := h.user()
	first := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"})["paymentId"].(string)
	if c := h.rpcCode(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"}); c != "PAYMENT_IN_PROGRESS" {
		t.Fatalf("second start while the first is open = %s", c)
	}
	h.mustAPI(200, http.MethodPut, "/payment-settings", map[string]any{"oncePending": "replace"})
	second := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"})["paymentId"].(string)
	if p := h.payment(first); p.Status != "expired" || p.Error.String != "REPLACED" {
		t.Fatalf("replaced payment = %+v", p)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": first, "token": "sandbox-" + first + "-t"}); c != "PAYMENT_EXPIRED" {
		t.Fatalf("late sandbox receipt for the replaced payment = %s", c)
	}
	if sub := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": second, "token": "sandbox-" + second + "-t"}); sub["rewardGranted"] != true {
		t.Fatalf("second = %v", sub)
	}
	if c := h.rpcCode(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"}); c != "PRODUCT_ALREADY_BOUGHT" {
		t.Fatalf("start after buying = %s", c)
	}
	h.mustAPI(200, http.MethodPut, "/payment-settings", map[string]any{"oncePending": "hold", "onceHoldMinutes": 15})
	alice := h.user()
	old := h.mustRPC(alice, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"})["paymentId"].(string)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_payment SET created_at=now()-interval '20 minutes' WHERE id=$1`, old); err != nil {
		t.Fatal(err)
	}
	if _, err := h.rpc(alice, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"}); err != nil {
		t.Fatalf("start after the hold time = %v", err)
	}
	if p := h.payment(old); p.Status != "expired" {
		t.Fatalf("stale open payment = %+v", p)
	}
}

func TestStoreSandboxFollowsCurrentSetting(t *testing.T) {
	h := newPaymentHarness(t)
	product := h.storeProduct("gems", false)
	bob := h.user()
	open := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"})["paymentId"].(string)
	stale := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"})["paymentId"].(string)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_payment SET status='expired', error='TIMED_OUT' WHERE id=$1`, stale); err != nil {
		t.Fatal(err)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": stale, "token": "sandbox-" + stale + "-t"}); c != "PAYMENT_EXPIRED" {
		t.Fatalf("expired sandbox payment = %s", c)
	}
	view := h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": false, "config": map[string]any{"sandbox": false}})
	if view["closedSandboxPayments"].(float64) != 1 {
		t.Fatalf("closed on sandbox off = %v", view)
	}
	if p := h.payment(open); p.Status != "failed" || p.Error.String != "SANDBOX_OFF" {
		t.Fatalf("open sandbox payment after turning sandbox off = %+v", p)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": open, "token": "sandbox-" + open + "-t"}); c != "PAYMENT_CLOSED" {
		t.Fatalf("fake receipt after sandbox off = %s", c)
	}
	if h.wallet(bob)["coin"] != 0 {
		t.Fatalf("coins = %v", h.wallet(bob))
	}
	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true, "packageName": "ir.game"}})
	keep := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"})["paymentId"].(string)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_payment SET sandbox=false WHERE id=$1`, keep); err != nil {
		t.Fatal(err)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": keep, "token": "sandbox-" + keep + "-t"}); c != "STORE_MODE_CHANGED" {
		t.Fatalf("real payment checked in sandbox = %s", c)
	}
	if p := h.payment(keep); p.Status != "pending" {
		t.Fatalf("real payment closed = %+v", p)
	}
}
