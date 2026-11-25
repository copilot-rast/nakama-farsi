package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
)

type fakeStoreReply struct {
	mu     sync.Mutex
	status int
	body   string
	paths  []string
	keys   []string
}

func (f *fakeStoreReply) set(status int, body string) {
	f.mu.Lock()
	f.status, f.body = status, body
	f.mu.Unlock()
}

func newFakeStoreReply(t *testing.T) (*fakeStoreReply, *httptest.Server) {
	f := &fakeStoreReply{status: 200, body: `{}`}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.paths = append(f.paths, r.URL.EscapedPath())
		f.keys = append(f.keys, r.Header.Get("CAFEBAZAAR-PISHKHAN-API-SECRET")+r.Header.Get("X-Access-Token"))
		status, body := f.status, f.body
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return f, srv
}

func (h *harness) rpcFailure(user, name string, payload any) (*appError, *runtime.Error) {
	h.t.Helper()
	_, err := h.rpc(user, name, payload)
	if err == nil {
		h.t.Fatalf("%s should fail", name)
	}
	app, ok := asApp(err)
	if !ok {
		h.t.Fatalf("%s: not an app error: %v", name, err)
	}
	var re *runtime.Error
	if !errors.As(h.s.runtimeError(err), &re) {
		h.t.Fatalf("%s: no runtime error", name)
	}
	return app, re
}

func TestStoreConfigErrorsAreNotRetryable(t *testing.T) {
	h, _ := newStoreHarness(t)
	fake, srv := newFakeStoreReply(t)
	swap(t, &bazaarBase, srv.URL)
	swap(t, &myketBase, srv.URL)
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_payment_method SET store_last_failure=NULL, store_last_ok_at=NULL, store_last_test=NULL`); err != nil {
		t.Fatal(err)
	}
	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": true, "config": map[string]any{"sandbox": false, "packageName": "ir.road.storm", "apiSecret": "WRONG-SECRET"}})
	h.mustAPI(200, http.MethodPut, "/payment-methods/myket", map[string]any{"isActive": true, "config": map[string]any{"sandbox": false, "packageName": "ir.road.storm", "accessToken": "MYKET-TOKEN"}})
	product := h.appProduct("fir_oil_100", "bazaar", map[string]any{"paymentMethods": []string{"bazaar", "myket"}, "sku": map[string]string{"bazaar": "fir_oil_100", "myket": "fir_oil_100"}})
	bob, alice := h.user(), h.user()
	coins := h.wallet(bob)["coin"]

	stuck := h.startPay(bob, product, "bazaar")
	realToken := "real-bazaar-token-1"
	fake.set(403, `{"error":"access_forbidden","error_description":"Invalid API secret"}`)
	app, re := h.rpcFailure(bob, "economy_purchase_submit", map[string]any{"paymentId": stuck, "token": realToken})
	if app.Code != "STORE_CONFIGURATION_REJECTED" || app.Status != http.StatusPreconditionFailed || re.Code != int(codes.FailedPrecondition) {
		t.Fatalf("config error = %+v grpc %d", app, re.Code)
	}
	if p := h.payment(stuck); p.Status != "pending" || p.Error.String != "STORE_CONFIGURATION_REJECTED" {
		t.Fatalf("stuck payment = %+v", p)
	}
	if fake.keys[len(fake.keys)-1] != "WRONG-SECRET" {
		t.Fatal("secret header not sent")
	}
	app, re = h.rpcFailure(bob, "economy_purchase_review", map[string]any{"method": "bazaar", "sku": "fir_oil_100", "token": "probe-not-a-real-token-000"})
	if app.Code != "STORE_CONFIGURATION_REJECTED" || re.Code != int(codes.FailedPrecondition) {
		t.Fatalf("review config error = %+v", app)
	}

	list := h.mustAPI(200, http.MethodGet, "/payment-methods", nil)
	var health map[string]any
	raw, _ := json.Marshal(list)
	items, _ := list["list"].([]any)
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["method"] == "bazaar" {
			health, _ = m["health"].(map[string]any)
		}
		if m["method"] == "zarinpal" && m["health"] != nil {
			t.Fatal("gateways have no store health")
		}
	}
	failure, _ := health["lastFailure"].(map[string]any)
	if failure == nil || failure["httpStatus"] != float64(403) || failure["reason"] != reasonCredentials || failure["storeError"] != "access_forbidden" || failure["storeDescription"] != "Invalid API secret" || health["stuckPayments"] != float64(1) {
		t.Fatalf("health = %v", health)
	}
	if strings.Contains(string(raw), "WRONG-SECRET") || strings.Contains(string(raw), realToken) {
		t.Fatal("methods list leaks the secret or a token")
	}

	cases := []struct {
		status int
		body   string
		want   string
	}{
		{403, `{"error":"access_forbidden"}`, "credentials_rejected"},
		{401, `{"error":"invalid_credentials"}`, "credentials_rejected"},
		{404, `{"error":"not_found","error_description":"The requested purchase is not found!"}`, "credentials_ok"},
		{404, `{"error":"invalid_value","error_description":"Package name is invalid"}`, "package_not_found"},
		{404, `{"error":"invalid_value","error_description":"Product is not found."}`, "product_not_found"},
		{502, `{}`, "unreachable"},
	}
	for _, tc := range cases {
		fake.set(tc.status, tc.body)
		out := h.mustAPI(200, http.MethodPost, "/payment-methods/bazaar/test", map[string]any{})
		if out["result"] != tc.want || out["sku"] != "fir_oil_100" || out["skuSource"] != "product" {
			t.Fatalf("test %d %s = %v", tc.status, tc.body, out)
		}
	}
	if p := fake.paths[len(fake.paths)-1]; !strings.HasPrefix(p, "/api/validate/ir.road.storm/inapp/fir_oil_100/purchases/nakama-credentials-test-") {
		t.Fatalf("test path = %s", p)
	}
	out := h.mustAPI(200, http.MethodPost, "/payment-methods/bazaar/test", map[string]any{"sku": "other_sku"})
	if out["sku"] != "other_sku" || out["skuSource"] != "given" {
		t.Fatalf("given sku = %v", out)
	}
	swap(t, &bazaarBase, "http://127.0.0.1:1")
	if out := h.mustAPI(200, http.MethodPost, "/payment-methods/bazaar/test", nil); out["result"] != "unreachable" {
		t.Fatalf("unreachable test = %v", out)
	}
	app, re = h.rpcFailure(bob, "economy_purchase_submit", map[string]any{"paymentId": stuck, "token": realToken})
	if app.Code != "STORE_UNAVAILABLE" || app.Status != http.StatusServiceUnavailable || re.Code != int(codes.Unavailable) {
		t.Fatalf("unreachable = %+v grpc %d", app, re.Code)
	}
	if p := h.payment(stuck); p.Status != "pending" || p.Error.String != "STORE_UNAVAILABLE" {
		t.Fatalf("unreachable payment = %+v", p)
	}
	swap(t, &bazaarBase, srv.URL)
	if code, _ := h.api(http.MethodPost, "/payment-methods/zarinpal/test", nil); code != 400 {
		t.Fatalf("gateway test = %d", code)
	}

	unknown := h.startPay(bob, product, "bazaar")
	fake.set(404, `{"error":"not_found","error_description":"The requested purchase is not found!"}`)
	if app, re = h.rpcFailure(bob, "economy_purchase_submit", map[string]any{"paymentId": unknown, "token": "made-up-token"}); app.Code != "PURCHASE_INVALID" || re.Code != int(codes.FailedPrecondition) {
		t.Fatalf("unknown token = %+v", app)
	}
	if p := h.payment(unknown); p.Status != "failed" {
		t.Fatalf("unknown token payment = %+v", p)
	}
	fake.set(404, `{"error":"invalid_value","error_description":"Invalid token!"}`)
	if app, _ = h.rpcFailure(bob, "economy_purchase_review", map[string]any{"method": "bazaar", "sku": "fir_oil_100", "token": "probe-not-a-real-token-000"}); app.Code != "PURCHASE_INVALID" {
		t.Fatalf("invalid token review = %+v", app)
	}

	second := h.startPay(bob, product, "bazaar")
	fake.set(401, `{"error":"invalid_credentials"}`)
	if app, _ = h.rpcFailure(bob, "economy_purchase_submit", map[string]any{"paymentId": second, "token": "real-bazaar-token-2"}); app.Code != "STORE_CONFIGURATION_REJECTED" {
		t.Fatalf("second = %+v", app)
	}

	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": true, "config": map[string]any{"apiSecret": "RIGHT-SECRET"}})
	fake.set(200, `{"purchaseState":0,"consumptionState":1,"developerPayload":"`+stuck+`","purchaseTime":1}`)
	detail := h.mustAPI(200, http.MethodPost, "/payments/"+stuck+"/recheck", nil)
	rc, _ := detail["recheck"].(map[string]any)
	if detail["status"] != "success" || rc["status"] != "success" || rc["rewardGranted"] != true || h.wallet(bob)["coin"] != coins+100 {
		t.Fatalf("recheck = %v", detail)
	}
	if fake.keys[len(fake.keys)-1] != "RIGHT-SECRET" || !strings.HasSuffix(fake.paths[len(fake.paths)-1], "/purchases/"+realToken+"/") {
		t.Fatalf("recheck call = %v %v", fake.paths[len(fake.paths)-1], fake.keys[len(fake.keys)-1])
	}
	logs, _ := detail["logs"].([]any)
	validates := 0
	for _, l := range logs {
		if m, _ := l.(map[string]any); m["action"] == "store_validate" {
			validates++
		}
	}
	if validates < 3 {
		t.Fatalf("store_validate logs = %d", validates)
	}
	if code, out := h.api(http.MethodPost, "/payments/"+stuck+"/recheck", nil); code != 409 || out["error"] != "PAYMENT_NOT_PENDING" {
		t.Fatalf("recheck settled = %d %v", code, out)
	}
	fresh := h.startPay(bob, product, "bazaar")
	if code, out := h.api(http.MethodPost, "/payments/"+fresh+"/recheck", nil); code != 409 || out["error"] != "STORE_TOKEN_UNKNOWN" {
		t.Fatalf("recheck without token = %d %v", code, out)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_payment SET status='expired', error='TIMED_OUT', once_key=NULL WHERE id=$1`, second); err != nil {
		t.Fatal(err)
	}
	fake.set(200, `{"purchaseState":0,"consumptionState":1,"developerPayload":"`+second+`","purchaseTime":1}`)
	bulk := h.mustAPI(200, http.MethodPost, "/payment-methods/bazaar/recheck", nil)
	if bulk["checked"] != float64(1) || bulk["delivered"] != float64(1) || h.payment(second).Status != "success" || h.wallet(bob)["coin"] != coins+200 {
		t.Fatalf("bulk = %v", bulk)
	}

	rv := h.mustRPC(bob, "economy_purchase_review", map[string]any{"method": "bazaar", "sku": "fir_oil_100", "token": realToken})
	if rv["status"] != "success" || rv["paymentId"] != stuck || rv["rewardGranted"] != true || h.wallet(bob)["coin"] != coins+200 {
		t.Fatalf("review delivered = %v", rv)
	}
	if c := h.rpcCode(alice, "economy_purchase_review", map[string]any{"method": "bazaar", "sku": "fir_oil_100", "token": realToken}); c != "TOKEN_ALREADY_USED" {
		t.Fatalf("other account review = %s", c)
	}
	sub := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": stuck, "token": realToken})
	if sub["status"] != "success" || sub["rewardGranted"] != true {
		t.Fatalf("submit delivered = %v", sub)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": stuck, "token": "another-token"}); c != "PAYMENT_ALREADY_SETTLED" {
		t.Fatalf("submit other token on settled = %s", c)
	}

	mp := h.startPay(bob, product, "myket")
	fake.set(404, `{"code":404,"messageCode":"SkuIdNotFound","translatedMessage":"x"}`)
	if app, _ = h.rpcFailure(bob, "economy_purchase_submit", map[string]any{"paymentId": mp, "token": "myket-token"}); app.Code != "STORE_CONFIGURATION_REJECTED" || app.Status != http.StatusPreconditionFailed {
		t.Fatalf("myket sku = %+v", app)
	}
	fake.set(401, `{}`)
	if out := h.mustAPI(200, http.MethodPost, "/payment-methods/myket/test", nil); out["result"] != "credentials_rejected" {
		t.Fatalf("myket test = %v", out)
	}
	fake.set(404, `{"code":404,"messageCode":"PurchaseNotFound"}`)
	if out := h.mustAPI(200, http.MethodPost, "/payment-methods/myket/test", nil); out["result"] != "credentials_ok" {
		t.Fatalf("myket ok = %v", out)
	}

	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": false, "config": map[string]any{}})
	if app, re = h.rpcFailure(bob, "economy_purchase_start", map[string]any{"productId": product, "method": "bazaar"}); app.Code != "METHOD_UNAVAILABLE" || app.Status != http.StatusPreconditionFailed || re.Code != int(codes.FailedPrecondition) {
		t.Fatalf("inactive = %+v", app)
	}
	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true, "packageName": "ir.game"}})
	h.mustAPI(200, http.MethodPut, "/payment-methods/myket", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true, "packageName": "ir.game"}})
}
