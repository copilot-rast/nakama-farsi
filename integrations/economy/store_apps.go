package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/heroiclabs/nakama/v3/iap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

var (
	googlePlayBase         = "https://androidpublisher.googleapis.com/androidpublisher/v3"
	appleAPIProduction     = "https://api.storekit.itunes.apple.com"
	appleAPISandbox        = "https://api.storekit-sandbox.itunes.apple.com"
	appleReceiptProduction = "https://buy.itunes.apple.com/verifyReceipt"
	appleReceiptSandbox    = "https://sandbox.itunes.apple.com/verifyReceipt"
	appleRootPEM           = iap.AppleRootPEM
)

const (
	googlePlayScope      = "https://www.googleapis.com/auth/androidpublisher"
	appleLeafOID         = "1.2.840.113635.100.6.11.1"
	appleIntermediateOID = "1.2.840.113635.100.6.2.1"
)

var appleTransactionID = regexp.MustCompile(`^[0-9]{1,32}$`)

func storeTokenLimit(method string) int {
	if method == "apple" {
		return 65536
	}
	return 1024
}

func storeKeyMatches(stored, key string) bool {
	return stored == key || strings.HasPrefix(stored, key+"#")
}

func (s *service) paySettings(ctx context.Context) paymentSettings {
	if s.db == nil {
		return defaultPaymentSettings()
	}
	st, err := loadPaymentSettings(ctx, s.db)
	if err != nil {
		return defaultPaymentSettings()
	}
	return st
}

func (s *service) payloadOptional(ctx context.Context, method string) bool {
	return isAppStore(method) && s.paySettings(ctx).AppStorePayload == "optional"
}

func (s *service) storeEnvironment(ctx context.Context, check *storeCheck, sandbox, test bool) {
	if !check.Valid {
		return
	}
	switch {
	case sandbox && !test:
		check.Valid, check.Code, check.State = false, "STORE_ENVIRONMENT_MISMATCH", "LIVE_PURCHASE_IN_SANDBOX"
	case !sandbox && test:
		if !s.paySettings(ctx).LiveAcceptsTestPurchases {
			check.Valid, check.Code, check.State = false, "TEST_PURCHASE_REFUSED", "TEST_PURCHASE"
		} else {
			check.State = "VALID_TEST"
		}
	}
}

type googleTokenCache struct {
	mu      sync.Mutex
	sources map[string]oauth2.TokenSource
}

var googleTokens = &googleTokenCache{sources: map[string]oauth2.TokenSource{}}

func googleServiceAccount(raw string) (*jwt.Config, error) {
	conf, err := google.JWTConfigFromJSON([]byte(raw), googlePlayScope)
	if err != nil {
		return nil, errors.New("not a Google service account JSON key")
	}
	if conf.Email == "" || len(conf.PrivateKey) == 0 {
		return nil, errors.New("the key has no client_email or private_key")
	}
	block, _ := pem.Decode(conf.PrivateKey)
	if block == nil {
		return nil, errors.New("private_key is not PEM")
	}
	if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
			return nil, errors.New("private_key cannot be parsed")
		}
	}
	return conf, nil
}

func (s *service) googleAccessToken(raw string) (string, storeClass) {
	sum := sha256.Sum256([]byte(raw))
	key := hex.EncodeToString(sum[:])
	googleTokens.mu.Lock()
	ts := googleTokens.sources[key]
	if ts == nil {
		conf, err := googleServiceAccount(raw)
		if err != nil {
			googleTokens.mu.Unlock()
			return "", storeConfigRejected
		}
		ts = oauth2.ReuseTokenSource(nil, conf.TokenSource(context.WithValue(context.Background(), oauth2.HTTPClient, s.client)))
		googleTokens.sources[key] = ts
	}
	googleTokens.mu.Unlock()
	tok, err := ts.Token()
	if err != nil {
		var re *oauth2.RetrieveError
		if errors.As(err, &re) && re.Response != nil && re.Response.StatusCode >= 400 && re.Response.StatusCode < 500 {
			return "", storeConfigRejected
		}
		return "", storeUnavailable
	}
	return tok.AccessToken, storeAccepted
}

func (s *service) validateGoogle(ctx context.Context, cfg methodConfig, sku, token string, subscription bool) (storeCheck, error) {
	pkg, err := storeSegment(cfg.str("packageName"))
	if err != nil {
		return storeCheck{}, err
	}
	skuSeg, err := storeSegment(sku)
	if err != nil {
		return storeCheck{}, err
	}
	tokenSeg, err := storeSegment(token)
	if err != nil {
		return storeCheck{}, err
	}
	base := strings.TrimRight(googlePlayBase, "/") + "/applications/" + pkg + "/purchases/"
	target := base + "products/" + skuSeg + "/tokens/" + tokenSeg
	if subscription {
		target = base + "subscriptionsv2/tokens/" + tokenSeg
	}
	logged := map[string]any{"sku": sku, "token": token}
	access, class := s.googleAccessToken(cfg.str("serviceAccountJson"))
	if class != storeAccepted {
		reason := reasonCredentials
		if class == storeUnavailable {
			reason = reasonNetwork
		}
		return storeCheck{Class: class, Reason: reason, StoreError: "OAUTH_TOKEN_FAILED", Exchange: exchange{URL: target, Request: logged, Response: map[string]any{"error": "OAUTH_TOKEN_FAILED"}}}, nil
	}
	header := map[string]string{"Authorization": "Bearer " + access}
	reply, ex, err := s.doJSON(ctx, http.MethodGet, target, header, nil, logged)
	check := storeCheck{Exchange: ex, StoreKey: token}
	if err != nil || ex.HTTPStatus == nil {
		check.Class, check.Reason = storeUnavailable, reasonNetwork
		return check, nil
	}
	check.noteReply("google_play", *ex.HTTPStatus, reply)
	if check.Class == storeAccepted && reply == nil {
		check.Class, check.Reason = storeUnavailable, reasonStoreError
	}
	if check.Class != storeAccepted {
		return check, nil
	}
	st := s.paySettings(ctx)
	sandbox := cfg.flag("sandbox")
	if subscription {
		return s.googleSubscription(ctx, check, reply, header, base, skuSeg, tokenSeg, sku, sandbox, st), nil
	}
	state := numberField(reply, "purchaseState")
	consumption := numberField(reply, "consumptionState")
	ack := numberField(reply, "acknowledgementState")
	purchaseType := numberField(reply, "purchaseType")
	check.RefID = textField(reply, "orderId")
	check.Payload = firstNonEmpty(textField(reply, "obfuscatedExternalProfileId"), textField(reply, "obfuscatedExternalAccountId"), textField(reply, "developerPayload"))
	check.Valid = state != nil && *state == 0
	switch {
	case check.Valid:
		check.State = "VALID"
	case state != nil && *state == 1:
		check.State = "CANCELED"
	case state != nil && *state == 2:
		check.State, check.Code = "PENDING", "PURCHASE_PENDING"
	default:
		check.State = "INVALID"
	}
	if pid := textField(reply, "productId"); check.Valid && pid != "" && pid != sku {
		check.Valid, check.Code, check.State = false, "PRODUCT_MISMATCH", "PRODUCT_MISMATCH"
	}
	if check.Valid && consumption != nil && *consumption == 1 && st.StoreConsumed == "refuse" {
		check.Valid, check.Code, check.State = false, "PURCHASE_CONSUMED", "CONSUMED"
	}
	s.storeEnvironment(ctx, &check, sandbox, purchaseType != nil && *purchaseType == 0)
	if check.Valid && st.GoogleAcknowledge && ack != nil && *ack == 0 && (consumption == nil || *consumption == 0) {
		check.Acknowledge = s.googleAcknowledger(base+"products/"+skuSeg+"/tokens/"+tokenSeg+":acknowledge", header)
	}
	return check, nil
}

func (s *service) googleSubscription(ctx context.Context, check storeCheck, reply map[string]any, header map[string]string, base, skuSeg, tokenSeg, sku string, sandbox bool, st paymentSettings) storeCheck {
	state := textField(reply, "subscriptionState")
	check.RefID = textField(reply, "latestOrderId")
	check.LinkedToken = textField(reply, "linkedPurchaseToken")
	ids := section(reply, "externalAccountIdentifiers")
	check.Payload = firstNonEmpty(textField(ids, "obfuscatedExternalProfileId"), textField(ids, "obfuscatedExternalAccountId"))
	var until *int64
	matched := false
	items, _ := reply["lineItems"].([]any)
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if textField(item, "productId") != sku {
			continue
		}
		matched = true
		if t, err := time.Parse(time.RFC3339Nano, textField(item, "expiryTime")); err == nil {
			ms := t.UnixMilli()
			if until == nil || ms > *until {
				until = &ms
			}
		}
	}
	if !matched {
		check.State, check.Code = "PRODUCT_MISMATCH", "PRODUCT_MISMATCH"
		return check
	}
	check.ValidUntil = until
	live := state == "SUBSCRIPTION_STATE_ACTIVE" || state == "SUBSCRIPTION_STATE_IN_GRACE_PERIOD" || state == "SUBSCRIPTION_STATE_CANCELED"
	check.Valid = live && until != nil && *until >= s.now().UnixMilli()
	switch {
	case check.Valid:
		check.State = "ACTIVE"
	case state == "SUBSCRIPTION_STATE_PENDING":
		check.State, check.Code = "PENDING", "PURCHASE_PENDING"
	default:
		check.State = "EXPIRED"
	}
	_, test := reply["testPurchase"]
	s.storeEnvironment(ctx, &check, sandbox, test)
	if check.Valid && st.GoogleAcknowledge && state != "SUBSCRIPTION_STATE_PENDING" && textField(reply, "acknowledgementState") == "ACKNOWLEDGEMENT_STATE_PENDING" {
		check.Acknowledge = s.googleAcknowledger(base+"subscriptions/"+skuSeg+"/tokens/"+tokenSeg+":acknowledge", header)
	}
	return check
}

func (s *service) googleAcknowledger(target string, header map[string]string) func(context.Context, string) {
	return func(ctx context.Context, paymentID string) {
		_, ex, err := s.doJSON(ctx, http.MethodPost, target, header, map[string]any{}, map[string]any{"acknowledge": true})
		s.logExchange(ctx, paymentID, "store_validate", ex)
		if err != nil || ex.HTTPStatus == nil || *ex.HTTPStatus >= 300 {
			s.logger.Warn("Economy Google Play acknowledge failed for payment %s", paymentID)
		}
	}
}

func appleAPIKeys(cfg methodConfig) bool {
	return cfg.str("issuerId") != "" && cfg.str("keyId") != "" && cfg.str("privateKeyP8") != ""
}

func parseAppleKey(raw string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(raw)))
	if block == nil {
		return nil, errors.New("not PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := key.(*ecdsa.PrivateKey)
	if !ok || ec.Curve != elliptic.P256() {
		return nil, errors.New("not a P-256 key")
	}
	return ec, nil
}

func b64url(raw []byte) string { return base64.RawURLEncoding.EncodeToString(raw) }

func es256(key *ecdsa.PrivateKey, input string) (string, error) {
	sum := sha256.Sum256([]byte(input))
	r, sv, err := ecdsa.Sign(rand.Reader, key, sum[:])
	if err != nil {
		return "", err
	}
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	sv.FillBytes(sig[32:])
	return b64url(sig), nil
}

func (s *service) appleJWT(cfg methodConfig) (string, error) {
	key, err := parseAppleKey(cfg.str("privateKeyP8"))
	if err != nil {
		return "", err
	}
	now := s.now()
	header, _ := json.Marshal(map[string]any{"alg": "ES256", "kid": cfg.str("keyId"), "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{"iss": cfg.str("issuerId"), "iat": now.Unix(), "exp": now.Add(20 * time.Minute).Unix(), "aud": "appstoreconnect-v1", "bid": cfg.str("bundleId")})
	input := b64url(header) + "." + b64url(claims)
	sig, err := es256(key, input)
	if err != nil {
		return "", err
	}
	return input + "." + sig, nil
}

func splitJWS(token string) ([]string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, false
	}
	return parts, true
}

func jwsPayloadUnverified(token string) (map[string]any, bool) {
	parts, ok := splitJWS(token)
	if !ok {
		return nil, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}
	out := map[string]any{}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if dec.Decode(&out) != nil {
		return nil, false
	}
	return out, true
}

func hasExtension(cert *x509.Certificate, oid string) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.String() == oid {
			return true
		}
	}
	return false
}

func verifyAppleJWS(token string, now time.Time) (map[string]any, error) {
	parts, ok := splitJWS(token)
	if !ok {
		return nil, errors.New("not a JWS")
	}
	rawHeader, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header struct {
		Alg string   `json:"alg"`
		X5c []string `json:"x5c"`
	}
	if err := json.Unmarshal(rawHeader, &header); err != nil {
		return nil, err
	}
	if header.Alg != "ES256" || len(header.X5c) < 2 {
		return nil, errors.New("unexpected JWS header")
	}
	certs := make([]*x509.Certificate, 0, len(header.X5c))
	for _, enc := range header.X5c {
		der, err := base64.StdEncoding.DecodeString(enc)
		if err != nil {
			return nil, err
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(appleRootPEM)) {
		return nil, errors.New("apple root certificate missing")
	}
	inter := x509.NewCertPool()
	for _, c := range certs[1:] {
		inter.AddCert(c)
	}
	if !hasExtension(certs[0], appleLeafOID) || !hasExtension(certs[1], appleIntermediateOID) {
		return nil, errors.New("certificate chain is not an App Store chain")
	}
	if _, err := certs[0].Verify(x509.VerifyOptions{Roots: roots, Intermediates: inter, CurrentTime: now, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}}); err != nil {
		return nil, err
	}
	pub, ok := certs[0].PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("leaf key is not ECDSA")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sig) != 64 {
		return nil, errors.New("bad signature encoding")
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if !ecdsa.Verify(pub, sum[:], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])) {
		return nil, errors.New("signature does not match")
	}
	payload, ok := jwsPayloadUnverified(token)
	if !ok {
		return nil, errors.New("payload is not JSON")
	}
	return payload, nil
}

type appleTx struct {
	TransactionID         string
	OriginalTransactionID string
	BundleID              string
	ProductID             string
	Environment           string
	AppAccountToken       string
	ExpiresMs             *int64
	RevokedMs             *int64
	PurchaseMs            *int64
}

func appleTxFromJWS(m map[string]any) appleTx {
	return appleTx{
		TransactionID: textField(m, "transactionId"), OriginalTransactionID: textField(m, "originalTransactionId"),
		BundleID: textField(m, "bundleId"), ProductID: textField(m, "productId"), Environment: textField(m, "environment"),
		AppAccountToken: textField(m, "appAccountToken"), ExpiresMs: numberField(m, "expiresDate"), RevokedMs: numberField(m, "revocationDate"),
		PurchaseMs: numberField(m, "purchaseDate"),
	}
}

func appleTxFromReceipt(m map[string]any, bundle, env string) appleTx {
	return appleTx{
		TransactionID: textField(m, "transaction_id"), OriginalTransactionID: textField(m, "original_transaction_id"),
		BundleID: bundle, ProductID: textField(m, "product_id"), Environment: env, AppAccountToken: textField(m, "app_account_token"),
		ExpiresMs: numberField(m, "expires_date_ms"), RevokedMs: numberField(m, "cancellation_date_ms"), PurchaseMs: numberField(m, "purchase_date_ms"),
	}
}

func (s *service) appleFill(ctx context.Context, check *storeCheck, cfg methodConfig, tx appleTx, sku string, subscription bool) {
	check.Class = storeAccepted
	check.Payload = strings.ToLower(tx.AppAccountToken)
	check.RefID = tx.TransactionID
	check.StoreKey = tx.TransactionID
	switch {
	case tx.TransactionID == "":
		check.State = "INVALID"
		return
	case tx.BundleID != cfg.str("bundleId"):
		check.State, check.Code = "BUNDLE_MISMATCH", "BUNDLE_MISMATCH"
		return
	case tx.ProductID != sku:
		check.State, check.Code = "PRODUCT_MISMATCH", "PRODUCT_MISMATCH"
		return
	case tx.RevokedMs != nil && *tx.RevokedMs > 0:
		check.State = "REFUNDED"
		return
	}
	if subscription {
		check.ValidUntil = tx.ExpiresMs
		check.LinkedToken = tx.OriginalTransactionID
		check.Valid = tx.ExpiresMs != nil && *tx.ExpiresMs >= s.now().UnixMilli()
		check.State = "EXPIRED"
		if check.Valid {
			check.State = "ACTIVE"
		}
	} else {
		check.Valid, check.State = true, "VALID"
	}
	s.storeEnvironment(ctx, check, cfg.flag("sandbox"), tx.Environment != "Production")
}

func (s *service) validateApple(ctx context.Context, cfg methodConfig, sku, token string, subscription bool) (storeCheck, error) {
	if _, ok := splitJWS(token); ok {
		if appleAPIKeys(cfg) {
			claims, _ := jwsPayloadUnverified(token)
			id := textField(claims, "transactionId")
			if !appleTransactionID.MatchString(id) {
				return storeCheck{Class: storeRejected, State: "SIGNATURE_INVALID", Exchange: exchange{URL: "jws://apple", Request: map[string]any{"sku": sku}, Response: map[string]any{"error": "NO_TRANSACTION_ID"}}}, nil
			}
			return s.appleLookup(ctx, cfg, id, sku, subscription)
		}
		claims, err := verifyAppleJWS(token, s.now())
		ex := exchange{URL: "jws://apple", Request: map[string]any{"sku": sku}}
		if err != nil {
			ex.Response = map[string]any{"error": "SIGNATURE_INVALID", "detail": err.Error()}
			return storeCheck{Class: storeRejected, State: "SIGNATURE_INVALID", Exchange: ex}, nil
		}
		ex.Response = claims
		check := storeCheck{Exchange: ex}
		s.appleFill(ctx, &check, cfg, appleTxFromJWS(claims), sku, subscription)
		return check, nil
	}
	if appleTransactionID.MatchString(token) {
		if !appleAPIKeys(cfg) {
			return storeCheck{}, badRequest("APPLE_API_KEYS_REQUIRED", "token", "transaction ids need the App Store Server API keys; send the receipt instead")
		}
		return s.appleLookup(ctx, cfg, token, sku, subscription)
	}
	if cfg.str("sharedSecret") == "" {
		return storeCheck{}, badRequest("APPLE_SHARED_SECRET_REQUIRED", "token", "receipts need the shared secret; send the transaction id instead")
	}
	return s.appleReceipt(ctx, cfg, sku, token, subscription)
}

func (s *service) appleLookup(ctx context.Context, cfg methodConfig, id, sku string, subscription bool) (storeCheck, error) {
	bearer, err := s.appleJWT(cfg)
	if err != nil {
		return storeCheck{Class: storeConfigRejected, Reason: reasonCredentials, StoreError: "KEY_INVALID", Exchange: exchange{URL: "apple://jwt", Response: map[string]any{"error": "KEY_INVALID"}}}, nil
	}
	hosts := []string{appleAPIProduction}
	if cfg.flag("sandbox") {
		hosts = []string{appleAPISandbox}
	} else if s.paySettings(ctx).LiveAcceptsTestPurchases {
		hosts = append(hosts, appleAPISandbox)
	}
	var check storeCheck
	for _, host := range hosts {
		target := strings.TrimRight(host, "/") + "/inApps/v1/transactions/" + url.PathEscape(id)
		reply, ex, err := s.doJSON(ctx, http.MethodGet, target, map[string]string{"Authorization": "Bearer " + bearer}, nil, map[string]any{"transactionId": id, "sku": sku})
		check = storeCheck{Exchange: ex}
		if err != nil || ex.HTTPStatus == nil {
			check.Class, check.Reason = storeUnavailable, reasonNetwork
			return check, nil
		}
		check.noteReply("apple", *ex.HTTPStatus, reply)
		if check.Class == storeRejected {
			continue
		}
		if check.Class != storeAccepted {
			return check, nil
		}
		claims, err := verifyAppleJWS(textField(reply, "signedTransactionInfo"), s.now())
		if err != nil {
			check.Class, check.State = storeUnavailable, "SIGNATURE_INVALID"
			return check, nil
		}
		s.appleFill(ctx, &check, cfg, appleTxFromJWS(claims), sku, subscription)
		if check.RefID != id {
			check.Valid, check.State, check.Code = false, "INVALID", ""
		}
		return check, nil
	}
	return check, nil
}

func appleReceiptClass(status int64) storeClass {
	switch {
	case status == 0:
		return storeAccepted
	case status == 21004:
		return storeConfigRejected
	case status == 21005 || status == 21009 || (status >= 21100 && status <= 21199):
		return storeUnavailable
	}
	return storeRejected
}

func (s *service) appleReceipt(ctx context.Context, cfg methodConfig, sku, receipt string, subscription bool) (storeCheck, error) {
	body := map[string]any{"receipt-data": receipt, "password": cfg.str("sharedSecret"), "exclude-old-transactions": true}
	short := receipt
	if len(short) > 24 {
		short = short[:24] + "..."
	}
	logged := map[string]any{"receipt-data": short, "password": secretMask, "sku": sku}
	reply, ex, err := s.doJSON(ctx, http.MethodPost, appleReceiptProduction, nil, body, logged)
	if err == nil && ex.HTTPStatus != nil && *ex.HTTPStatus == 200 {
		if st := numberField(reply, "status"); st != nil && *st == 21007 {
			reply, ex, err = s.doJSON(ctx, http.MethodPost, appleReceiptSandbox, nil, body, logged)
		}
	}
	check := storeCheck{Exchange: ex}
	if err != nil || ex.HTTPStatus == nil || *ex.HTTPStatus != 200 || reply == nil {
		check.Class = storeUnavailable
		return check, nil
	}
	status := numberField(reply, "status")
	if status == nil {
		check.Class = storeUnavailable
		return check, nil
	}
	check.Class = appleReceiptClass(*status)
	if check.Class != storeAccepted {
		check.State = "STATUS_" + strconv.FormatInt(*status, 10)
		check.StoreError = strconv.FormatInt(*status, 10)
		switch check.Class {
		case storeConfigRejected:
			check.Reason = reasonCredentials
		case storeUnavailable:
			check.Reason = reasonStoreError
		default:
			check.Reason = reasonTokenInvalid
		}
		return check, nil
	}
	rec := section(reply, "receipt")
	bundle := textField(rec, "bundle_id")
	env := textField(reply, "environment")
	if env == "" {
		env = "Production"
	}
	var entries []map[string]any
	collect := func(list any) {
		items, _ := list.([]any)
		for _, raw := range items {
			if m, ok := raw.(map[string]any); ok && textField(m, "product_id") == sku {
				entries = append(entries, m)
			}
		}
	}
	collect(rec["in_app"])
	if subscription {
		collect(reply["latest_receipt_info"])
	}
	if bundle != cfg.str("bundleId") {
		s.appleFill(ctx, &check, cfg, appleTx{TransactionID: "-", BundleID: bundle, ProductID: sku, Environment: env}, sku, subscription)
		return check, nil
	}
	if len(entries) == 0 {
		s.appleFill(ctx, &check, cfg, appleTx{TransactionID: "-", BundleID: bundle, Environment: env}, sku, subscription)
		return check, nil
	}
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := numberField(entries[i], "purchase_date_ms"), numberField(entries[j], "purchase_date_ms")
		if subscription {
			a, b = numberField(entries[i], "expires_date_ms"), numberField(entries[j], "expires_date_ms")
		}
		return a != nil && (b == nil || *a > *b)
	})
	pick := entries[0]
	if !subscription {
		for _, e := range entries {
			if !s.appleTransactionUsed(ctx, textField(e, "transaction_id")) {
				pick = e
				break
			}
		}
	}
	s.appleFill(ctx, &check, cfg, appleTxFromReceipt(pick, bundle, env), sku, subscription)
	return check, nil
}

func (s *service) appleTransactionUsed(ctx context.Context, id string) bool {
	if s.db == nil || id == "" {
		return false
	}
	var used bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_payment WHERE method='apple' AND status='success' AND (ref_id=$1 OR store_token=$1))`, id).Scan(&used)
	return used
}
