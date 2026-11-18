package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func newStoreHarness(t *testing.T) (*harness, *fakeStores) {
	h := newEventHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `TRUNCATE economy_payment_settings, economy_payment_webhook`); err != nil {
		t.Fatal(err)
	}
	return h, newFakeStores(t)
}

func (h *harness) appProduct(key, method string, extra map[string]any) string {
	h.t.Helper()
	prize := h.coinPrize(key+"_coins", 100, true)
	body := map[string]any{"key": key, "isActive": true, "name": map[string]string{"fa": key, "en": key + " pack"}, "prizeId": prize, "paymentMethods": []string{method}, "priceToman": 20000}
	if isStore(method) {
		body["sku"] = map[string]string{method: key + ".sku"}
	}
	for k, v := range extra {
		body[k] = v
	}
	return h.mustAPI(201, http.MethodPost, "/products", body)["id"].(string)
}

func (h *harness) startPay(user, product, method string) string {
	h.t.Helper()
	return h.mustRPC(user, "economy_purchase_start", map[string]any{"productId": product, "method": method})["paymentId"].(string)
}

func (h *harness) paySettingsPut(change map[string]any) {
	h.t.Helper()
	h.mustAPI(200, http.MethodPut, "/payment-settings", change)
}

func googlePurchase(payment, order string, extra map[string]any) map[string]any {
	out := map[string]any{"kind": "androidpublisher#productPurchase", "purchaseTimeMillis": "1700000000000", "purchaseState": 0, "consumptionState": 0,
		"acknowledgementState": 0, "orderId": order, "obfuscatedExternalProfileId": payment, "productId": "gems.sku", "quantity": 1}
	for k, v := range extra {
		if v == nil {
			delete(out, k)
			continue
		}
		out[k] = v
	}
	return out
}

func TestGooglePlayStoreFlow(t *testing.T) {
	h, f := newStoreHarness(t)
	account := googleServiceAccountJSON(t, f.srv.URL+"/token")
	view := h.mustAPI(200, http.MethodPut, "/payment-methods/google_play", map[string]any{"isActive": true, "config": map[string]any{"packageName": "com.soltan.game", "serviceAccountJson": account, "sandbox": false}})
	if view["config"].(map[string]any)["serviceAccountJson"] != secretMask || view["configured"] != true {
		t.Fatalf("google view = %v", view)
	}
	if code, out := h.api(http.MethodPut, "/payment-methods/google_play", map[string]any{"isActive": true, "config": map[string]any{"serviceAccountJson": "{\"type\":\"user\"}"}}); code != 400 || out["field"] != "serviceAccountJson" {
		t.Fatalf("bad service account = %d %v", code, out)
	}
	product := h.appProduct("gems", "google_play", nil)
	alice, bob := h.user(), h.user()

	pay := h.startPay(alice, product, "google_play")
	f.mu.Lock()
	f.google["tok-valid"] = googlePurchase(pay, "GPA.1111-2222-3333-00001", nil)
	f.mu.Unlock()
	out := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": pay, "token": "tok-valid"})
	if out["status"] != "success" || out["rewardGranted"] != true || h.wallet(alice)["coin"] != 100 {
		t.Fatalf("valid google purchase = %v wallet %v", out, h.wallet(alice))
	}
	if p := h.payment(pay); p.StoreToken.String != "tok-valid" {
		t.Fatalf("payment = %+v", p)
	}
	f.mu.Lock()
	if len(f.googleAuth) == 0 || f.googleAuth[0] != "Bearer ya29.fake" || len(f.acks) != 1 || !strings.Contains(f.acks[0], "/applications/com.soltan.game/purchases/products/gems.sku/tokens/tok-valid:acknowledge") {
		t.Fatalf("auth %v acks %v", f.googleAuth, f.acks)
	}
	f.mu.Unlock()
	again := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": pay, "token": "tok-valid"})
	if again["status"] != "success" || h.wallet(alice)["coin"] != 100 {
		t.Fatalf("resubmit = %v", again)
	}

	dup := h.startPay(bob, product, "google_play")
	f.mu.Lock()
	f.google["tok-other"] = googlePurchase(dup, "GPA.1111-2222-3333-00001", nil)
	f.mu.Unlock()
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": dup, "token": "tok-other"}); c != "TOKEN_ALREADY_USED" {
		t.Fatalf("duplicate orderId = %s", c)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": dup, "token": "tok-valid"}); c != "DEVELOPER_PAYLOAD_MISMATCH" && c != "TOKEN_ALREADY_USED" {
		t.Fatalf("someone else's token = %s", c)
	}
	if h.wallet(bob)["coin"] != 0 {
		t.Fatalf("bob got coins %v", h.wallet(bob))
	}

	cancelled := h.startPay(bob, product, "google_play")
	f.mu.Lock()
	f.google["tok-cancel"] = googlePurchase(cancelled, "GPA.c", map[string]any{"purchaseState": 1})
	f.googleCode["tok-bad"] = 400
	f.mu.Unlock()
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": cancelled, "token": "tok-cancel"}); c != "PURCHASE_INVALID" {
		t.Fatalf("cancelled = %s", c)
	}
	if p := h.payment(cancelled); p.Status != "failed" {
		t.Fatalf("cancelled payment = %+v", p)
	}
	invalid := h.startPay(bob, product, "google_play")
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": invalid, "token": "tok-bad"}); c != "PURCHASE_INVALID" {
		t.Fatalf("invalid token = %s", c)
	}

	pending := h.startPay(bob, product, "google_play")
	f.mu.Lock()
	f.google["tok-pending"] = googlePurchase(pending, "GPA.p", map[string]any{"purchaseState": 2})
	f.mu.Unlock()
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": pending, "token": "tok-pending"}); c != "PURCHASE_PENDING" {
		t.Fatalf("pending = %s", c)
	}
	if p := h.payment(pending); p.Status != "pending" {
		t.Fatalf("pending payment closed = %+v", p)
	}

	test := h.startPay(bob, product, "google_play")
	f.mu.Lock()
	f.google["tok-test"] = googlePurchase(test, "GPA.t1", map[string]any{"purchaseType": 0})
	f.mu.Unlock()
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": test, "token": "tok-test"}); c != "TEST_PURCHASE_REFUSED" {
		t.Fatalf("test purchase in live = %s", c)
	}
	h.paySettingsPut(map[string]any{"liveAcceptsTestPurchases": true})
	review := h.startPay(bob, product, "google_play")
	f.mu.Lock()
	f.google["tok-review"] = googlePurchase(review, "GPA.t2", map[string]any{"purchaseType": 0})
	f.mu.Unlock()
	if out := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": review, "token": "tok-review"}); out["status"] != "success" {
		t.Fatalf("test purchase with liveAcceptsTestPurchases = %v", out)
	}
	h.paySettingsPut(map[string]any{"liveAcceptsTestPurchases": false})

	h.mustAPI(200, http.MethodPut, "/payment-methods/google_play", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true}})
	sbTest := h.startPay(alice, product, "google_play")
	sbLive := h.startPay(bob, product, "google_play")
	f.mu.Lock()
	f.google["tok-sb-test"] = googlePurchase(sbTest, "GPA.s1", map[string]any{"purchaseType": 0})
	f.google["tok-sb-live"] = googlePurchase(sbLive, "GPA.s2", nil)
	f.mu.Unlock()
	if out := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": sbTest, "token": "tok-sb-test"}); out["status"] != "success" {
		t.Fatalf("test purchase in sandbox = %v", out)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": sbLive, "token": "tok-sb-live"}); c != "STORE_ENVIRONMENT_MISMATCH" {
		t.Fatalf("live purchase in sandbox = %s", c)
	}
	if h.payment(sbTest).Sandbox != true {
		t.Fatal("sandbox flag not recorded")
	}

	h.paySettingsPut(map[string]any{"storeConsumed": "refuse"})
	consumed := h.startPay(alice, product, "google_play")
	f.mu.Lock()
	f.google["tok-consumed"] = googlePurchase(consumed, "GPA.cons", map[string]any{"purchaseType": 0, "consumptionState": 1})
	f.mu.Unlock()
	if c := h.rpcCode(alice, "economy_purchase_submit", map[string]any{"paymentId": consumed, "token": "tok-consumed"}); c != "PURCHASE_CONSUMED" {
		t.Fatalf("consumed = %s", c)
	}
	cfg := h.mustRPC(alice, "economy_store_config", map[string]any{"method": "google_play"})
	if cfg["packageName"] != "com.soltan.game" || cfg["sandbox"] != true {
		t.Fatalf("store config = %v", cfg)
	}
}

func TestGooglePlaySubscription(t *testing.T) {
	h, f := newStoreHarness(t)
	account := googleServiceAccountJSON(t, f.srv.URL+"/token")
	h.mustAPI(200, http.MethodPut, "/payment-methods/google_play", map[string]any{"isActive": true, "config": map[string]any{"packageName": "com.soltan.game", "serviceAccountJson": account}})
	product := h.appProduct("vip", "google_play", map[string]any{"isSubscription": true})
	alice := h.user()
	pay := h.startPay(alice, product, "google_play")
	expiry := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	f.mu.Lock()
	f.google["sub-1"] = map[string]any{"subscriptionState": "SUBSCRIPTION_STATE_ACTIVE", "latestOrderId": "GPA.sub-1", "acknowledgementState": "ACKNOWLEDGEMENT_STATE_PENDING",
		"lineItems":                  []any{map[string]any{"productId": "vip.sku", "expiryTime": expiry}},
		"externalAccountIdentifiers": map[string]any{"obfuscatedExternalProfileId": pay}}
	f.mu.Unlock()
	out := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": pay, "token": "sub-1"})
	if out["status"] != "success" || !strings.HasPrefix(h.payment(pay).StoreToken.String, "sub-1#") {
		t.Fatalf("subscription = %v %+v", out, h.payment(pay))
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.acks) != 1 || !strings.Contains(f.acks[0], "/purchases/subscriptions/vip.sku/tokens/sub-1:acknowledge") {
		t.Fatalf("acks = %v", f.acks)
	}
}

func appleTxClaims(id, bundle, product, env, account string) map[string]any {
	out := map[string]any{"transactionId": id, "originalTransactionId": id, "bundleId": bundle, "productId": product, "environment": env,
		"purchaseDate": time.Now().UnixMilli(), "type": "Consumable", "quantity": 1}
	if account != "" {
		out["appAccountToken"] = account
	}
	return out
}

func checkAppleJWT(t *testing.T, token string, key *ecdsa.PrivateKey) {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("apple jwt = %q", token)
	}
	hdr, _ := base64.RawURLEncoding.DecodeString(parts[0])
	body, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var head, claims map[string]any
	_ = json.Unmarshal(hdr, &head)
	_ = json.Unmarshal(body, &claims)
	if head["alg"] != "ES256" || head["kid"] != "KEY123" || claims["iss"] != "issuer-1" || claims["aud"] != "appstoreconnect-v1" || claims["bid"] != "com.soltan.ios" {
		t.Fatalf("apple jwt header %v claims %v", head, claims)
	}
	sig, _ := base64.RawURLEncoding.DecodeString(parts[2])
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if len(sig) != 64 || !ecdsa.Verify(&key.PublicKey, sum[:], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])) {
		t.Fatal("apple jwt signature does not verify")
	}
}

func TestAppleStoreFlow(t *testing.T) {
	h, f := newStoreHarness(t)
	chain := newAppleChain(t, true)
	swap(t, &appleRootPEM, chain.rootPEM)
	p8, key := appleP8(t)
	if code, out := h.api(http.MethodPut, "/payment-methods/apple", map[string]any{"isActive": true, "config": map[string]any{"bundleId": "com.soltan.ios", "keyId": "KEY123"}}); code != 400 {
		t.Fatalf("apple without any credential = %d %v", code, out)
	}
	view := h.mustAPI(200, http.MethodPut, "/payment-methods/apple", map[string]any{"isActive": true, "config": map[string]any{"bundleId": "com.soltan.ios", "issuerId": "issuer-1", "keyId": "KEY123", "privateKeyP8": p8, "sharedSecret": "apple-shared"}})
	cfgView := view["config"].(map[string]any)
	if cfgView["privateKeyP8"] != secretMask || cfgView["sharedSecret"] != secretMask || cfgView["keyId"] != "KEY123" {
		t.Fatalf("apple view = %v", view)
	}
	product := h.appProduct("coins", "apple", nil)
	alice, bob := h.user(), h.user()

	pay := h.startPay(alice, product, "apple")
	f.mu.Lock()
	f.appleProd["2000000001"] = chain.sign(t, appleTxClaims("2000000001", "com.soltan.ios", "coins.sku", "Production", pay))
	f.mu.Unlock()
	out := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": pay, "token": "2000000001"})
	if out["status"] != "success" || out["rewardGranted"] != true || h.wallet(alice)["coin"] != 100 {
		t.Fatalf("apple valid = %v", out)
	}
	f.mu.Lock()
	checkAppleJWT(t, f.appleJWTs[0], key)
	f.mu.Unlock()
	if p := h.payment(pay); p.StoreToken.String != "2000000001" {
		t.Fatalf("apple payment = %+v", p)
	}

	clientJWS := chain.sign(t, appleTxClaims("2000000001", "com.soltan.ios", "coins.sku", "Production", pay))
	if again := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": pay, "token": clientJWS}); again["status"] != "success" || h.wallet(alice)["coin"] != 100 {
		t.Fatalf("resubmit as JWS = %v", again)
	}

	other := h.startPay(bob, product, "apple")
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": other, "token": "2000000001"}); c != "DEVELOPER_PAYLOAD_MISMATCH" {
		t.Fatalf("someone else's transaction = %s", c)
	}
	f.mu.Lock()
	f.appleProd["2000000002"] = chain.sign(t, appleTxClaims("2000000002", "com.soltan.ios", "coins.sku", "Production", ""))
	f.mu.Unlock()
	first := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": other, "token": "2000000002"})
	if first["status"] != "success" {
		t.Fatalf("no appAccountToken (payload optional) = %v", first)
	}
	dup := h.startPay(bob, product, "apple")
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": dup, "token": "2000000002"}); c != "TOKEN_ALREADY_USED" {
		t.Fatalf("duplicate transaction = %s", c)
	}
	if h.wallet(bob)["coin"] != 100 {
		t.Fatalf("bob wallet = %v", h.wallet(bob))
	}

	wrong := h.startPay(bob, product, "apple")
	f.mu.Lock()
	f.appleProd["2000000003"] = chain.sign(t, appleTxClaims("2000000003", "com.other.app", "coins.sku", "Production", wrong))
	f.appleProd["2000000004"] = chain.sign(t, appleTxClaims("2000000004", "com.soltan.ios", "coins.sku", "Sandbox", ""))
	f.appleProd["2000000005"] = chain.sign(t, appleTxClaims("2000000005", "com.soltan.ios", "other.sku", "Production", ""))
	f.mu.Unlock()
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": wrong, "token": "2000000003"}); c != "BUNDLE_MISMATCH" {
		t.Fatalf("wrong bundle = %s", c)
	}
	env := h.startPay(bob, product, "apple")
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": env, "token": "2000000004"}); c != "TEST_PURCHASE_REFUSED" {
		t.Fatalf("sandbox transaction in live = %s", c)
	}
	sku := h.startPay(bob, product, "apple")
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": sku, "token": "2000000005"}); c != "PRODUCT_MISMATCH" {
		t.Fatalf("other product = %s", c)
	}
	missing := h.startPay(bob, product, "apple")
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": missing, "token": "2999999999"}); c != "PURCHASE_INVALID" {
		t.Fatalf("unknown transaction = %s", c)
	}

	fakeChain := newAppleChain(t, true)
	forged := h.startPay(bob, product, "apple")
	f.mu.Lock()
	f.appleProd["2000000006"] = fakeChain.sign(t, appleTxClaims("2000000006", "com.soltan.ios", "coins.sku", "Production", forged))
	f.mu.Unlock()
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": forged, "token": "2000000006"}); c != "STORE_UNAVAILABLE" {
		t.Fatalf("forged signed answer = %s", c)
	}
	if p := h.payment(forged); p.Status != "pending" {
		t.Fatalf("forged answer closed payment = %+v", p)
	}

	h.mustAPI(200, http.MethodPut, "/payment-methods/apple", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true}})
	sb := h.startPay(alice, product, "apple")
	f.mu.Lock()
	f.appleSand["2000000007"] = chain.sign(t, appleTxClaims("2000000007", "com.soltan.ios", "coins.sku", "Sandbox", sb))
	f.appleSand["2000000008"] = chain.sign(t, appleTxClaims("2000000008", "com.soltan.ios", "coins.sku", "Production", ""))
	f.mu.Unlock()
	if out := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": sb, "token": "2000000007"}); out["status"] != "success" {
		t.Fatalf("sandbox transaction in sandbox = %v", out)
	}
	sbLive := h.startPay(alice, product, "apple")
	if c := h.rpcCode(alice, "economy_purchase_submit", map[string]any{"paymentId": sbLive, "token": "2000000008"}); c != "STORE_ENVIRONMENT_MISMATCH" {
		t.Fatalf("production transaction in sandbox = %s", c)
	}
	f.mu.Lock()
	hosts := strings.Join(f.appleHosts, ",")
	f.mu.Unlock()
	if !strings.HasSuffix(hosts, "sandbox,sandbox") {
		t.Fatalf("hosts = %s", hosts)
	}
}

func TestAppleReceiptAndLocalJWS(t *testing.T) {
	h, f := newStoreHarness(t)
	chain := newAppleChain(t, true)
	swap(t, &appleRootPEM, chain.rootPEM)
	h.mustAPI(200, http.MethodPut, "/payment-methods/apple", map[string]any{"isActive": true, "config": map[string]any{"bundleId": "com.soltan.ios", "sharedSecret": "apple-shared", "sandbox": true}})
	product := h.appProduct("coins", "apple", nil)
	alice := h.user()

	fake := h.startPay(alice, product, "apple")
	if c := h.rpcCode(alice, "economy_purchase_submit", map[string]any{"paymentId": fake, "token": "ZmFrZS1yZWNlaXB0"}); c != "PURCHASE_INVALID" {
		t.Fatalf("fake receipt = %s", c)
	}
	if p := h.payment(fake); p.Status != "failed" {
		t.Fatalf("fake receipt payment = %+v", p)
	}

	pay := h.startPay(alice, product, "apple")
	f.mu.Lock()
	f.receipts["cmVhbC1yZWNlaXB0"] = map[string]any{"status": 21007}
	f.sandReceipt["cmVhbC1yZWNlaXB0"] = map[string]any{"status": 0, "environment": "Sandbox", "receipt": map[string]any{"bundle_id": "com.soltan.ios",
		"in_app": []any{map[string]any{"product_id": "coins.sku", "transaction_id": "1000000001", "original_transaction_id": "1000000001", "purchase_date_ms": "1700000000000"}}}}
	f.mu.Unlock()
	out := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": pay, "token": "cmVhbC1yZWNlaXB0"})
	if out["status"] != "success" || h.payment(pay).StoreToken.String != "1000000001" {
		t.Fatalf("receipt via sandbox retry = %v %+v", out, h.payment(pay))
	}
	if again := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": pay, "token": "cmVhbC1yZWNlaXB0"}); again["status"] != "success" || h.wallet(alice)["coin"] != 100 {
		t.Fatalf("receipt resubmit = %v", again)
	}
	reuse := h.startPay(alice, product, "apple")
	if c := h.rpcCode(alice, "economy_purchase_submit", map[string]any{"paymentId": reuse, "token": "cmVhbC1yZWNlaXB0"}); c != "TOKEN_ALREADY_USED" {
		t.Fatalf("receipt reuse = %s", c)
	}

	jwsPay := h.startPay(alice, product, "apple")
	good := chain.sign(t, appleTxClaims("1000000002", "com.soltan.ios", "coins.sku", "Sandbox", jwsPay))
	if out := h.mustRPC(alice, "economy_purchase_submit", map[string]any{"paymentId": jwsPay, "token": good}); out["status"] != "success" {
		t.Fatalf("local JWS = %v", out)
	}
	bad := h.startPay(alice, product, "apple")
	forged := newAppleChain(t, true).sign(t, appleTxClaims("1000000003", "com.soltan.ios", "coins.sku", "Sandbox", bad))
	if c := h.rpcCode(alice, "economy_purchase_submit", map[string]any{"paymentId": bad, "token": forged}); c != "PURCHASE_INVALID" {
		t.Fatalf("forged JWS = %s", c)
	}
	noOID := h.startPay(alice, product, "apple")
	plain := newAppleChain(t, false)
	swap(t, &appleRootPEM, plain.rootPEM)
	if c := h.rpcCode(alice, "economy_purchase_submit", map[string]any{"paymentId": noOID, "token": plain.sign(t, appleTxClaims("1000000004", "com.soltan.ios", "coins.sku", "Sandbox", noOID))}); c != "PURCHASE_INVALID" {
		t.Fatalf("chain without App Store marker = %s", c)
	}
	if h.wallet(alice)["coin"] != 200 {
		t.Fatalf("wallet = %v", h.wallet(alice))
	}
}

func stripeSign(secret string, ts int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", ts)))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *harness) stripeHook(sigHeader string, body []byte) (int, map[string]any) {
	h.t.Helper()
	req := httptest.NewRequest(http.MethodPost, stripeWebhookPath, bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", sigHeader)
	rec := httptest.NewRecorder()
	h.init.public.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func stripeEventBody(id, typ, session, payment, status string, amount int64, currency string) []byte {
	raw, _ := json.Marshal(map[string]any{"id": id, "object": "event", "type": typ, "livemode": false, "data": map[string]any{"object": map[string]any{
		"id": session, "object": "checkout.session", "client_reference_id": payment, "payment_status": status, "amount_total": amount, "currency": currency,
		"payment_intent": "pi_" + id, "metadata": map[string]any{"paymentId": payment}}}})
	return raw
}

func TestStripeCheckoutAndWebhook(t *testing.T) {
	h, f := newStoreHarness(t)
	if code, out := h.api(http.MethodPut, "/payment-methods/stripe", map[string]any{"isActive": true, "config": map[string]any{"secretKey": "sk_test_abc", "webhookSecret": "whsec_test", "testMode": false}}); code != 400 || out["error"] != "STRIPE_KEY_MODE" {
		t.Fatalf("live mode with a test key = %d %v", code, out)
	}
	view := h.mustAPI(200, http.MethodPut, "/payment-methods/stripe", map[string]any{"isActive": true, "config": map[string]any{"secretKey": "sk_test_abc", "webhookSecret": "whsec_test", "publishableKey": "pk_test_pub", "testMode": true}})
	cv := view["config"].(map[string]any)
	if cv["secretKey"] != secretMask || cv["webhookSecret"] != secretMask || cv["publishableKey"] != "pk_test_pub" || cv["currency"] != "usd" || cv["testMode"] != true {
		t.Fatalf("stripe view = %v", view)
	}
	if code, _ := h.api(http.MethodPost, "/products", map[string]any{"key": "noprice", "isActive": true, "name": map[string]string{"fa": "x"}, "prizeId": h.coinPrize("np", 1, true), "paymentMethods": []string{"stripe"}}); code != 400 {
		t.Fatalf("stripe product without price = %d", code)
	}
	product := h.appProduct("gold", "stripe", map[string]any{"priceMinor": 199, "priceCurrency": "USD", "discountPercent": 0})
	alice := h.user()
	cat := h.mustRPC(alice, "economy_catalog", map[string]any{"lang": "en"})
	var listed map[string]any
	for _, raw := range cat["products"].([]any) {
		if p := raw.(map[string]any); p["id"] == product {
			listed = p
		}
	}
	if listed == nil || listed["priceCurrency"] != "usd" || listed["finalMinor"].(float64) != 199 {
		t.Fatalf("catalog = %v", listed)
	}

	start := h.mustRPC(alice, "economy_purchase_start", map[string]any{"productId": product, "method": "stripe", "lang": "en"})
	pay := start["paymentId"].(string)
	if start["status"] != "redirect" || start["redirectUrl"] != "https://checkout.stripe.com/c/pay/cs_test_b" || start["sessionId"] != "cs_test_b" || start["publishableKey"] != "pk_test_pub" {
		t.Fatalf("start = %v", start)
	}
	f.mu.Lock()
	form, hdr := f.stripeForms[0], f.stripeHdrs[0]
	f.mu.Unlock()
	for k, want := range map[string]string{"mode": "payment", "client_reference_id": pay, "line_items[0][quantity]": "1", "line_items[0][price_data][currency]": "usd",
		"line_items[0][price_data][unit_amount]": "199", "line_items[0][price_data][product_data][name]": "gold pack", "metadata[paymentId]": pay, "payment_intent_data[metadata][paymentId]": pay} {
		if form.Get(k) != want {
			t.Fatalf("form %s = %q want %q (%v)", k, form.Get(k), want, form)
		}
	}
	if !strings.HasPrefix(form.Get("success_url"), "http://127.0.0.1:7450/v2/economy/stripe/return/") || !strings.HasSuffix(form.Get("cancel_url"), "?result=cancel") || form.Get("expires_at") == "" {
		t.Fatalf("urls = %v", form)
	}
	if hdr.Get("Authorization") != "Bearer sk_test_abc" || hdr.Get("Idempotency-Key") != "economy-"+pay || !strings.HasPrefix(hdr.Get("Content-Type"), "application/x-www-form-urlencoded") {
		t.Fatalf("headers = %v", hdr)
	}
	if p := h.payment(pay); p.Authority.String != "cs_test_b" || p.AmountMinor.Int64 != 199 || p.Currency.String != "usd" || !p.Sandbox {
		t.Fatalf("payment = %+v", p)
	}
	if st := h.mustRPC(alice, "economy_purchase_status", map[string]any{"paymentId": pay}); st["status"] != "pending" {
		t.Fatalf("status before webhook = %v", st)
	}

	ret := httptest.NewRecorder()
	h.init.public.ServeHTTP(ret, httptest.NewRequest(http.MethodGet, strings.TrimPrefix(form.Get("success_url"), "http://127.0.0.1:7450"), nil))
	if ret.Code != http.StatusFound || !strings.Contains(ret.Header().Get("Location"), "success=true") || !strings.Contains(ret.Header().Get("Location"), pay) {
		t.Fatalf("return = %d %v", ret.Code, ret.Header())
	}
	if h.payment(pay).Status != "pending" {
		t.Fatal("return page settled the payment")
	}

	now := time.Now().Unix()
	body := stripeEventBody("evt_1", "checkout.session.completed", "cs_test_b", pay, "paid", 199, "usd")
	if code, out := h.stripeHook(fmt.Sprintf("t=%d,v1=%s", now, stripeSign("whsec_wrong", now, body)), body); code != 400 || out["error"] != "INVALID_SIGNATURE" {
		t.Fatalf("bad signature = %d %v", code, out)
	}
	old := now - 301
	if code, out := h.stripeHook(fmt.Sprintf("t=%d,v1=%s", old, stripeSign("whsec_test", old, body)), body); code != 400 || out["error"] != "SIGNATURE_TOO_OLD" {
		t.Fatalf("old timestamp = %d %v", code, out)
	}
	if code, out := h.stripeHook("v1=abc", body); code != 400 || out["error"] != "INVALID_SIGNATURE" {
		t.Fatalf("no timestamp = %d %v", code, out)
	}
	if h.payment(pay).Status != "pending" || h.wallet(alice)["coin"] != 0 {
		t.Fatal("refused webhook changed the payment")
	}
	good := fmt.Sprintf("t=%d,v1=deadbeef,v1=%s", now, stripeSign("whsec_test", now, body))
	if code, out := h.stripeHook(good, body); code != 200 || out["result"] != "paid" {
		t.Fatalf("good webhook = %d %v", code, out)
	}
	if p := h.payment(pay); p.Status != "success" || !p.RewardGranted || h.wallet(alice)["coin"] != 100 {
		t.Fatalf("after webhook = %+v %v", p, h.wallet(alice))
	}
	if code, out := h.stripeHook(good, body); code != 200 || out["duplicate"] != true {
		t.Fatalf("duplicate event = %d %v", code, out)
	}
	body2 := stripeEventBody("evt_2", "checkout.session.async_payment_succeeded", "cs_test_b", pay, "paid", 199, "usd")
	if code, out := h.stripeHook(fmt.Sprintf("t=%d,v1=%s", now, stripeSign("whsec_test", now, body2)), body2); code != 200 || out["result"] != "already_settled" {
		t.Fatalf("second event for the same payment = %d %v", code, out)
	}
	if h.wallet(alice)["coin"] != 100 {
		t.Fatalf("granted twice: %v", h.wallet(alice))
	}
	if st := h.mustRPC(alice, "economy_purchase_status", map[string]any{"paymentId": pay}); st["status"] != "success" || st["rewardGranted"] != true {
		t.Fatalf("status after webhook = %v", st)
	}

	cheat := h.startPay(alice, product, "stripe")
	body3 := stripeEventBody("evt_3", "checkout.session.completed", h.payment(cheat).Authority.String, cheat, "paid", 1, "usd")
	if code, out := h.stripeHook(fmt.Sprintf("t=%d,v1=%s", now, stripeSign("whsec_test", now, body3)), body3); code != 200 || out["result"] != "amount_mismatch" {
		t.Fatalf("amount mismatch = %d %v", code, out)
	}
	if p := h.payment(cheat); p.Status != "failed" || p.Error.String != "AMOUNT_MISMATCH" {
		t.Fatalf("cheat payment = %+v", p)
	}
	late := h.startPay(alice, product, "stripe")
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE economy_payment SET status='expired', error='TIMED_OUT', once_key=NULL WHERE id=$1`, late); err != nil {
		t.Fatal(err)
	}
	body4 := stripeEventBody("evt_4", "checkout.session.completed", h.payment(late).Authority.String, late, "paid", 199, "usd")
	if code, out := h.stripeHook(fmt.Sprintf("t=%d,v1=%s", now, stripeSign("whsec_test", now, body4)), body4); code != 200 || out["result"] != "paid" {
		t.Fatalf("late webhook = %d %v", code, out)
	}
	if h.wallet(alice)["coin"] != 200 {
		t.Fatalf("late paid session not granted: %v", h.wallet(alice))
	}
	liveBody, _ := json.Marshal(map[string]any{"id": "evt_live", "type": "checkout.session.completed", "livemode": true, "data": map[string]any{"object": map[string]any{}}})
	if code, out := h.stripeHook(fmt.Sprintf("t=%d,v1=%s", now, stripeSign("whsec_test", now, liveBody)), liveBody); code != 400 || out["error"] != "MODE_MISMATCH" {
		t.Fatalf("live event in test mode = %d %v", code, out)
	}
	h.paySettingsPut(map[string]any{"stripeToleranceSeconds": 600})
	if code, _ := h.stripeHook(fmt.Sprintf("t=%d,v1=%s", old, stripeSign("whsec_test", old, body)), body); code != 200 {
		t.Fatalf("tolerance setting ignored: %d", code)
	}
	f.mu.Lock()
	f.stripeReply = func(_ url.Values, _ int) (int, map[string]any) {
		return 400, map[string]any{"error": map[string]any{"type": "invalid_request_error", "code": "amount_too_small", "message": "too small"}}
	}
	f.mu.Unlock()
	if c := h.rpcCode(alice, "economy_purchase_start", map[string]any{"productId": product, "method": "stripe"}); c != "GATEWAY_REQUEST_FAILED" {
		t.Fatalf("stripe refused session = %s", c)
	}
}
