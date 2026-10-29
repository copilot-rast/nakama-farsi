package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

type testLogger struct{}

func (testLogger) Debug(string, ...any)                       {}
func (testLogger) Info(string, ...any)                        {}
func (testLogger) Warn(string, ...any)                        {}
func (testLogger) Error(string, ...any)                       {}
func (l testLogger) WithField(string, any) runtime.Logger     { return l }
func (l testLogger) WithFields(map[string]any) runtime.Logger { return l }
func (testLogger) Fields() map[string]any                     { return nil }

func testService() *service {
	cfg, err := loadSettings(map[string]string{
		"ECONOMY_SECRET_KEY":      strings.Repeat("ab", 24),
		"ECONOMY_PUBLIC_BASE_URL": "http://127.0.0.1:7450",
	}, nil)
	if err != nil {
		panic(err)
	}
	return &service{logger: testLogger{}, cfg: cfg, client: &http.Client{Timeout: 5 * time.Second}, now: time.Now}
}

func swap(t *testing.T, target *string, value string) {
	old := *target
	*target = value
	t.Cleanup(func() { *target = old })
}

type recorded struct {
	path   string
	header http.Header
	body   map[string]any
}

func fakeServer(t *testing.T, status int, reply string, seen *recorded) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			seen.path = r.URL.EscapedPath()
			seen.header = r.Header.Clone()
			raw, _ := io.ReadAll(r.Body)
			seen.body = map[string]any{}
			_ = json.Unmarshal(raw, &seen.body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestZarinpalRequestAndVerify(t *testing.T) {
	s := testService()
	ctx := context.Background()
	var seen recorded
	srv := fakeServer(t, 200, `{"data":{"code":100,"authority":"S0001","fee":10},"errors":[]}`, &seen)
	swap(t, &zarinpalSandboxAPI, srv.URL)
	swap(t, &zarinpalSandboxStart, "https://sandbox.example/StartPay")
	start := s.zarinpalRequest(ctx, methodConfig{"sandbox": true}, 50000, "http://x/v2/economy/payments/callback/abc.def", "buy")
	if !start.OK || start.Reference != "S0001" || start.Redirect != "https://sandbox.example/StartPay/S0001" {
		t.Fatalf("start = %+v", start)
	}
	if seen.path != "/request.json" || seen.body["amount"] != float64(50000) || seen.body["currency"] != "IRR" || seen.body["merchant_id"] != zarinpalSandboxMerchant {
		t.Fatalf("request body = %v path %s", seen.body, seen.path)
	}
	logged, _ := json.Marshal(start.Exchange.Request)
	if strings.Contains(string(logged), "abc.def") {
		t.Fatalf("ticket leaked into log: %s", logged)
	}

	live := fakeServer(t, 200, `{"data":{"code":100,"authority":"A1"},"errors":[]}`, &seen)
	swap(t, &zarinpalLiveAPI, live.URL)
	start = s.zarinpalRequest(ctx, methodConfig{"merchantId": "real-merchant"}, 1000, "http://x/cb", "d")
	logged, _ = json.Marshal(start.Exchange.Request)
	if !start.OK || strings.Contains(string(logged), "real-merchant") || seen.body["merchant_id"] != "real-merchant" {
		t.Fatalf("live merchant must be sent but masked in logs: %s %v", logged, seen.body)
	}

	cases := []struct {
		reply   string
		outcome outcome
	}{
		{`{"data":{"code":100,"ref_id":12345,"card_pan":"6219****1234"},"errors":[]}`, outcomePaid},
		{`{"data":{"code":101,"ref_id":12345},"errors":[]}`, outcomePaid},
		{`{"data":{},"errors":{"code":-51,"message":"not paid"}}`, outcomeFailed},
		{`{"data":{},"errors":{"code":-50,"message":"amount"}}`, outcomeFailed},
		{`{"data":{},"errors":{"code":-9,"message":"validation"}}`, outcomeUnknown},
		{`not json`, outcomeUnknown},
	}
	for _, tc := range cases {
		srv := fakeServer(t, 200, tc.reply, &seen)
		swap(t, &zarinpalSandboxAPI, srv.URL)
		check := s.zarinpalVerify(ctx, methodConfig{"sandbox": true}, 50000, "S0001")
		if check.Outcome != tc.outcome {
			t.Fatalf("reply %s outcome %v want %v", tc.reply, check.Outcome, tc.outcome)
		}
		if tc.outcome == outcomePaid && (check.Amount == nil || *check.Amount != 50000 || check.RefID != "12345") {
			t.Fatalf("paid check = %+v", check)
		}
		if seen.body["authority"] != "S0001" || seen.body["amount"] != float64(50000) {
			t.Fatalf("verify body = %v", seen.body)
		}
	}
	swap(t, &zarinpalSandboxAPI, "http://127.0.0.1:1")
	if c := s.zarinpalVerify(ctx, methodConfig{"sandbox": true}, 1, "x"); c.Outcome != outcomeUnknown || c.Error != "GATEWAY_UNAVAILABLE" {
		t.Fatalf("network error check = %+v", c)
	}
}

func TestZibalCodes(t *testing.T) {
	n := func(v int64) *int64 { return &v }
	if zibalVerifyOutcome(n(100)) != outcomePaid || zibalVerifyOutcome(n(201)) != outcomePaid || zibalVerifyOutcome(n(202)) != outcomeFailed || zibalVerifyOutcome(n(203)) != outcomeUnknown || zibalVerifyOutcome(nil) != outcomeUnknown {
		t.Fatal("verify outcomes wrong")
	}
	for status, want := range map[int64]outcome{1: outcomePaid, 2: outcomePaidNotVerified, -1: outcomePending, 3: outcomeFailed, 12: outcomeFailed, 15: outcomeFailed, 16: outcomeFailed, 18: outcomeFailed, 13: outcomeUnknown, 17: outcomeUnknown} {
		if got := zibalInquiryOutcome(n(100), n(status)); got != want {
			t.Fatalf("inquiry (100,%d) = %v want %v", status, got, want)
		}
	}
	if zibalInquiryOutcome(n(102), n(1)) != outcomeUnknown {
		t.Fatal("non-100 result must be unknown")
	}
}

func TestZibalGateway(t *testing.T) {
	s := testService()
	ctx := context.Background()
	var seen recorded
	srv := fakeServer(t, 200, `{"result":100,"trackId":4879013321,"message":"success"}`, &seen)
	swap(t, &zibalBase, srv.URL)
	start := s.zibalRequest(ctx, methodConfig{"sandbox": true}, 20000, "http://x/v2/economy/payments/callback/t.k", "d", "pay-1")
	if !start.OK || start.Reference != "4879013321" || start.Redirect != srv.URL+"/start/4879013321" {
		t.Fatalf("start = %+v", start)
	}
	if seen.path != "/v1/request" || seen.body["merchant"] != "zibal" || seen.body["orderId"] != "pay-1" || seen.body["amount"] != float64(20000) {
		t.Fatalf("request = %s %v", seen.path, seen.body)
	}

	srv = fakeServer(t, 200, `{"result":100,"amount":20000,"refNumber":777,"cardNumber":"6037****1111","status":1}`, &seen)
	swap(t, &zibalBase, srv.URL)
	check := s.zibalVerify(ctx, methodConfig{"merchant": "m-secret"}, "4879013321")
	if check.Outcome != outcomePaid || check.Amount == nil || *check.Amount != 20000 || check.RefID != "777" || check.Card != "6037****1111" {
		t.Fatalf("verify = %+v", check)
	}
	if seen.body["trackId"] != float64(4879013321) || seen.body["merchant"] != "m-secret" {
		t.Fatalf("verify body = %v", seen.body)
	}
	logged, _ := json.Marshal(check.Exchange.Request)
	if strings.Contains(string(logged), "m-secret") {
		t.Fatalf("merchant leaked: %s", logged)
	}

	srv = fakeServer(t, 200, `{"result":201,"amount":19990}`, nil)
	swap(t, &zibalBase, srv.URL)
	check = s.zibalVerify(ctx, methodConfig{"sandbox": true}, "1")
	if check.Outcome != outcomePaid || *check.Amount != 19990 {
		t.Fatalf("201 = %+v", check)
	}

	srv = fakeServer(t, 200, `{"result":202,"message":"transaction failed"}`, nil)
	swap(t, &zibalBase, srv.URL)
	if c := s.zibalVerify(ctx, methodConfig{"sandbox": true}, "1"); c.Outcome != outcomeFailed || c.Error != "ZIBAL_202" {
		t.Fatalf("202 = %+v", c)
	}

	srv = fakeServer(t, 200, `{"result":100,"status":-1,"amount":20000}`, nil)
	swap(t, &zibalBase, srv.URL)
	if c := s.zibalInquire(ctx, methodConfig{"sandbox": true}, "1"); c.Outcome != outcomePending {
		t.Fatalf("inquiry pending = %+v", c)
	}
}

func TestStoreReplyClassifier(t *testing.T) {
	cases := []struct {
		status int
		err    string
		want   storeClass
	}{
		{200, "", storeAccepted}, {204, "", storeAccepted},
		{401, "", storeConfigRejected}, {403, "", storeConfigRejected}, {404, "invalid_value", storeConfigRejected},
		{400, "", storeRejected}, {404, "", storeRejected}, {410, "", storeRejected},
		{500, "", storeUnavailable}, {502, "", storeUnavailable}, {429, "", storeUnavailable}, {302, "", storeUnavailable},
	}
	for _, tc := range cases {
		if got := classifyStoreReply(tc.status, tc.err); got != tc.want {
			t.Fatalf("classify(%d,%q) = %v want %v", tc.status, tc.err, got, tc.want)
		}
	}
}

func TestBazaarAndMyketValidation(t *testing.T) {
	s := testService()
	ctx := context.Background()
	var seen recorded
	srv := fakeServer(t, 200, `{"purchaseState":0,"developerPayload":"pay-1","purchaseTime":1}`, &seen)
	swap(t, &bazaarBase, srv.URL)
	cfg := methodConfig{"packageName": "ir.game", "apiSecret": "sec"}
	check, err := s.validateStore(ctx, "bazaar", cfg, "coin.100", "tok/../x", false)
	if err != nil || check.Class != storeAccepted || !check.Valid || check.Payload != "pay-1" {
		t.Fatalf("bazaar = %+v %v", check, err)
	}
	if seen.path != "/api/validate/ir.game/inapp/coin.100/purchases/tok%2F..%2Fx/" || seen.header.Get("CAFEBAZAAR-PISHKHAN-API-SECRET") != "sec" {
		t.Fatalf("bazaar path %s headers %v", seen.path, seen.header)
	}

	srv = fakeServer(t, 200, `{"purchaseState":2,"developerPayload":"pay-1"}`, nil)
	swap(t, &bazaarBase, srv.URL)
	check, _ = s.validateStore(ctx, "bazaar", cfg, "coin.100", "tok", false)
	if check.Valid || check.State != "REFUNDED" {
		t.Fatalf("refunded = %+v", check)
	}

	future := time.Now().Add(time.Hour).UnixMilli()
	srv = fakeServer(t, 200, `{"validUntilTimestampMsec":`+jsonInt(future)+`,"linkedSubscriptionToken":"base"}`, &seen)
	swap(t, &bazaarBase, srv.URL)
	check, _ = s.validateStore(ctx, "bazaar", cfg, "vip", "tok", true)
	if !check.Valid || check.LinkedToken != "base" || check.ValidUntil == nil || *check.ValidUntil != future || seen.path != "/api/applications/ir.game/subscriptions/vip/purchases/tok" {
		t.Fatalf("subscription = %+v path %s", check, seen.path)
	}

	for status, want := range map[int]storeClass{404: storeRejected, 401: storeConfigRejected, 503: storeUnavailable} {
		srv = fakeServer(t, status, `{"error":"not_found"}`, nil)
		swap(t, &bazaarBase, srv.URL)
		check, _ = s.validateStore(ctx, "bazaar", cfg, "coin", "tok", false)
		if check.Class != want {
			t.Fatalf("status %d class %v want %v", status, check.Class, want)
		}
	}
	srv = fakeServer(t, 404, `{"error":"invalid_value"}`, nil)
	swap(t, &bazaarBase, srv.URL)
	if check, _ = s.validateStore(ctx, "bazaar", cfg, "coin", "tok", false); check.Class != storeConfigRejected {
		t.Fatalf("invalid_value = %v", check.Class)
	}
	swap(t, &bazaarBase, "http://127.0.0.1:1")
	if check, _ = s.validateStore(ctx, "bazaar", cfg, "coin", "tok", false); check.Class != storeUnavailable {
		t.Fatalf("network = %v", check.Class)
	}
	if _, err := s.validateStore(ctx, "bazaar", cfg, "coin", "..", false); err == nil {
		t.Fatal("dots-only token must be refused")
	}

	srv = fakeServer(t, 200, `{"purchaseState":0,"developerPayload":"pay-2"}`, &seen)
	swap(t, &myketBase, srv.URL)
	mcfg := methodConfig{"packageName": "ir.game", "accessToken": "acc"}
	check, err = s.validateStore(ctx, "myket", mcfg, "coin.100", "tk", false)
	if err != nil || !check.Valid || check.Payload != "pay-2" || seen.path != "/api/partners/applications/ir.game/purchases/products/coin.100/verify" || seen.header.Get("X-Access-Token") != "acc" || seen.body["tokenId"] != "tk" {
		t.Fatalf("myket = %+v %v %s %v", check, err, seen.path, seen.body)
	}
	if _, err := s.validateStore(ctx, "myket", mcfg, "vip", "tk", true); err == nil {
		t.Fatal("myket subscriptions must be refused")
	}
}

func TestStoreSandbox(t *testing.T) {
	s := testService()
	cfg := methodConfig{"sandbox": true, "packageName": "ir.game"}
	check, err := s.validateStore(context.Background(), "bazaar", cfg, "coin", "sandbox-3f1c2a10-1111-4222-8333-444455556666-xyz", false)
	if err != nil || !check.Valid || check.Payload != "3f1c2a10-1111-4222-8333-444455556666" {
		t.Fatalf("sandbox = %+v %v", check, err)
	}
	for _, bad := range []string{"real-token", "sandbox-", "sandbox-abc", "sandbox--x", "sandbox-abc-"} {
		check, _ := s.validateStore(context.Background(), "myket", cfg, "coin", bad, false)
		if check.Valid || check.Class != storeRejected {
			t.Fatalf("token %q must be rejected: %+v", bad, check)
		}
	}
	check, _ = s.validateStore(context.Background(), "bazaar", cfg, "vip", "sandbox-p-1", true)
	if !check.Valid || check.ValidUntil == nil {
		t.Fatalf("sandbox subscription = %+v", check)
	}
}
