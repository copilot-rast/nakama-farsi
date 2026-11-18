package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type appleChain struct {
	rootPEM string
	x5c     []string
	leaf    *ecdsa.PrivateKey
}

func oidFrom(s string) asn1.ObjectIdentifier {
	var out asn1.ObjectIdentifier
	for _, part := range strings.Split(s, ".") {
		n := 0
		for _, c := range part {
			n = n*10 + int(c-'0')
		}
		out = append(out, n)
	}
	return out
}

func newAppleChain(t *testing.T, leafOID bool) appleChain {
	t.Helper()
	now := time.Now()
	mk := func() *ecdsa.PrivateKey {
		k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	rootKey, interKey, leafKey := mk(), mk(), mk()
	rootTpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Test Apple Root"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTpl, rootTpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := x509.ParseCertificate(rootDER)
	interTpl := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "Test Apple WWDR"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, ExtraExtensions: []pkix.Extension{{Id: oidFrom(appleIntermediateOID), Value: []byte{5, 0}}}}
	interDER, err := x509.CreateCertificate(rand.Reader, interTpl, root, &interKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	inter, _ := x509.ParseCertificate(interDER)
	leafTpl := &x509.Certificate{SerialNumber: big.NewInt(3), Subject: pkix.Name{CommonName: "Test StoreKit"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(24 * time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature}
	if leafOID {
		leafTpl.ExtraExtensions = []pkix.Extension{{Id: oidFrom(appleLeafOID), Value: []byte{5, 0}}}
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTpl, inter, &leafKey.PublicKey, interKey)
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.StdEncoding.EncodeToString
	return appleChain{
		rootPEM: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootDER})),
		x5c:     []string{enc(leafDER), enc(interDER), enc(rootDER)},
		leaf:    leafKey,
	}
}

func (c appleChain) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]any{"alg": "ES256", "x5c": c.x5c})
	body, _ := json.Marshal(claims)
	input := b64url(header) + "." + b64url(body)
	sig, err := es256(c.leaf, input)
	if err != nil {
		t.Fatal(err)
	}
	return input + "." + sig
}

func appleP8(t *testing.T) (string, *ecdsa.PrivateKey) {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKCS8PrivateKey(k)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), k
}

type fakeStores struct {
	mu          sync.Mutex
	google      map[string]map[string]any
	googleCode  map[string]int
	googleAuth  []string
	acks        []string
	tokenCalls  int
	appleProd   map[string]string
	appleSand   map[string]string
	appleJWTs   []string
	appleHosts  []string
	receipts    map[string]map[string]any
	sandReceipt map[string]map[string]any
	stripeForms []url.Values
	stripeHdrs  []http.Header
	stripeReply func(form url.Values, n int) (int, map[string]any)
	srv         *httptest.Server
}

func newFakeStores(t *testing.T) *fakeStores {
	f := &fakeStores{google: map[string]map[string]any{}, googleCode: map[string]int{}, appleProd: map[string]string{}, appleSand: map[string]string{},
		receipts: map[string]map[string]any{}, sandReceipt: map[string]map[string]any{}}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	swap(t, &googlePlayBase, f.srv.URL+"/androidpublisher/v3")
	swap(t, &appleAPIProduction, f.srv.URL+"/apple-prod")
	swap(t, &appleAPISandbox, f.srv.URL+"/apple-sandbox")
	swap(t, &appleReceiptProduction, f.srv.URL+"/receipt-prod/verifyReceipt")
	swap(t, &appleReceiptSandbox, f.srv.URL+"/receipt-sandbox/verifyReceipt")
	swap(t, &stripeAPI, f.srv.URL+"/stripe")
	return f
}

func (f *fakeStores) reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (f *fakeStores) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path := r.URL.Path
	switch {
	case path == "/token":
		_ = r.ParseForm()
		f.tokenCalls++
		if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || r.Form.Get("assertion") == "" {
			f.reply(w, 400, map[string]any{"error": "invalid_grant"})
			return
		}
		f.reply(w, 200, map[string]any{"access_token": "ya29.fake", "token_type": "Bearer", "expires_in": 3600})
	case strings.HasPrefix(path, "/androidpublisher/v3/"):
		if strings.HasSuffix(path, ":acknowledge") {
			f.acks = append(f.acks, path)
			f.reply(w, 200, map[string]any{})
			return
		}
		f.googleAuth = append(f.googleAuth, r.Header.Get("Authorization"))
		token := path[strings.LastIndex(path, "/")+1:]
		if code, ok := f.googleCode[token]; ok {
			f.reply(w, code, map[string]any{"error": map[string]any{"code": code, "message": "fake"}})
			return
		}
		if body, ok := f.google[token]; ok {
			f.reply(w, 200, body)
			return
		}
		f.reply(w, 404, map[string]any{"error": map[string]any{"code": 404, "message": "not found"}})
	case strings.HasPrefix(path, "/apple-prod/inApps/v1/transactions/"), strings.HasPrefix(path, "/apple-sandbox/inApps/v1/transactions/"):
		f.appleJWTs = append(f.appleJWTs, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		id := path[strings.LastIndex(path, "/")+1:]
		store := f.appleProd
		host := "prod"
		if strings.HasPrefix(path, "/apple-sandbox/") {
			store, host = f.appleSand, "sandbox"
		}
		f.appleHosts = append(f.appleHosts, host)
		if jws, ok := store[id]; ok {
			f.reply(w, 200, map[string]any{"signedTransactionInfo": jws})
			return
		}
		f.reply(w, 404, map[string]any{"errorCode": 4040010, "errorMessage": "Transaction id not found."})
	case path == "/receipt-prod/verifyReceipt", path == "/receipt-sandbox/verifyReceipt":
		raw, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(raw, &in)
		data, _ := in["receipt-data"].(string)
		store := f.receipts
		if strings.HasPrefix(path, "/receipt-sandbox/") {
			store = f.sandReceipt
		}
		if body, ok := store[data]; ok {
			if in["password"] != "apple-shared" {
				f.reply(w, 200, map[string]any{"status": 21004})
				return
			}
			f.reply(w, 200, body)
			return
		}
		f.reply(w, 200, map[string]any{"status": 21002})
	case path == "/stripe/v1/checkout/sessions":
		_ = r.ParseForm()
		f.stripeForms = append(f.stripeForms, r.PostForm)
		f.stripeHdrs = append(f.stripeHdrs, r.Header.Clone())
		n := len(f.stripeForms)
		if f.stripeReply != nil {
			code, body := f.stripeReply(r.PostForm, n)
			f.reply(w, code, body)
			return
		}
		f.reply(w, 200, map[string]any{"id": "cs_test_" + string(rune('a'+n)), "object": "checkout.session", "url": "https://checkout.stripe.com/c/pay/cs_test_" + string(rune('a'+n)), "livemode": false,
			"payment_status": "unpaid", "amount_total": r.PostForm.Get("line_items[0][price_data][unit_amount]"), "currency": r.PostForm.Get("line_items[0][price_data][currency]")})
	default:
		f.reply(w, 404, map[string]any{"error": "no route " + path})
	}
}

func googleServiceAccountJSON(t *testing.T, tokenURI string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKCS8PrivateKey(key)
	sum := sha256.Sum256(der)
	raw, _ := json.Marshal(map[string]any{"type": "service_account", "project_id": "test", "private_key_id": base64.RawURLEncoding.EncodeToString(sum[:8]),
		"private_key": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), "client_email": "play@test.iam.gserviceaccount.com",
		"client_id": "1", "token_uri": tokenURI})
	return string(raw)
}
