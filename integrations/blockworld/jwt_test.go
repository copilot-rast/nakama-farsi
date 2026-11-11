package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testSigner(t *testing.T) *signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return newSigner(key, "http://127.0.0.1:7450/blockworld/")
}

func TestIssuedTokenVerifiesAgainstPublishedJWKS(t *testing.T) {
	s := testSigner(t)
	now := time.Unix(1_800_000_000, 0)
	tok, claims, err := s.issue("6d2e7a8c-1111-4222-8333-944455556666", "mina", "Mina", time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Issuer != "http://127.0.0.1:7450/blockworld" || claims.Audience != tokenAudience || claims.Name != "Mina" {
		t.Fatalf("claims %+v", claims)
	}
	raw, _ := json.Marshal(s.jwks())
	var set struct {
		Keys []struct{ Kty, N, E, Kid, Alg, Use string }
	}
	if err := json.Unmarshal(raw, &set); err != nil || len(set.Keys) != 1 {
		t.Fatalf("jwks %s", raw)
	}
	k := set.Keys[0]
	if k.Kty != "RSA" || k.Alg != "RS256" || k.Use != "sig" || k.Kid != s.kid {
		t.Fatalf("jwk %+v", k)
	}
	pub, err := jwkPublicKey(k.N, k.E)
	if err != nil {
		t.Fatal(err)
	}
	got, err := verify(tok, pub, claims.Issuer, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if got.Subject != claims.Subject || got.IssuedAt != now.Unix() || got.Expires != now.Add(time.Hour).Unix() {
		t.Fatalf("got %+v", got)
	}
	header, _ := base64.RawURLEncoding.DecodeString(strings.Split(tok, ".")[0])
	if !strings.Contains(string(header), `"kid":"`+s.kid+`"`) || !strings.Contains(string(header), `"alg":"RS256"`) {
		t.Fatalf("header %s", header)
	}
}

func TestVerifyRefusesTamperedExpiredAndForeignTokens(t *testing.T) {
	s := testSigner(t)
	other := testSigner(t)
	now := time.Unix(1_800_000_000, 0)
	tok, claims, _ := s.issue("u1", "a", "", time.Hour, now)
	parts := strings.Split(tok, ".")
	forged, _ := json.Marshal(worldClaims{Issuer: claims.Issuer, Subject: "someone-else", IssuedAt: now.Unix(), Expires: now.Add(time.Hour).Unix()})
	cases := map[string]struct {
		token string
		at    time.Time
		iss   string
		pub   *rsa.PublicKey
	}{
		"payload swapped": {parts[0] + "." + base64.RawURLEncoding.EncodeToString(forged) + "." + parts[2], now, claims.Issuer, &s.key.PublicKey},
		"expired":         {tok, now.Add(2 * time.Hour), claims.Issuer, &s.key.PublicKey},
		"other key":       {tok, now, claims.Issuer, &other.key.PublicKey},
		"other issuer":    {tok, now, "https://evil.example/blockworld", &s.key.PublicKey},
		"not a jwt":       {"abc", now, claims.Issuer, &s.key.PublicKey},
	}
	for name, c := range cases {
		if _, err := verify(c.token, c.pub, c.iss, c.at); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := verify(tok, &s.key.PublicKey, claims.Issuer, now.Add(time.Hour+30*time.Second)); err != nil {
		t.Fatalf("leeway: %v", err)
	}
}

func TestDiscoveryPointsAtJWKSUnderTheIssuer(t *testing.T) {
	s := testSigner(t)
	d := s.discovery()
	if d["issuer"] != "http://127.0.0.1:7450/blockworld" || d["jwks_uri"] != "http://127.0.0.1:7450/blockworld/jwks.json" {
		t.Fatalf("%v", d)
	}
	if issuerPath("http://127.0.0.1:7450/blockworld") != "/blockworld" || issuerPath("https://h.example") != "" {
		t.Fatal("issuerPath")
	}
}

func TestKeyIsCreatedOnceAndReloaded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "k", "jwt.pem")
	a, err := loadOrCreateKey(path, true)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v %v", info.Mode(), err)
	}
	b, err := loadOrCreateKey(path, true)
	if err != nil || thumbprint(&a.PublicKey) != thumbprint(&b.PublicKey) {
		t.Fatalf("reloaded a different key: %v", err)
	}
	if _, err := loadOrCreateKey(filepath.Join(t.TempDir(), "missing.pem"), false); err == nil {
		t.Fatal("a missing configured key must fail")
	}
}

func TestConfigNeedsAnIssuer(t *testing.T) {
	if _, err := loadConfig(map[string]string{}, "/tmp"); err == nil {
		t.Fatal("no issuer accepted")
	}
	c, err := loadConfig(map[string]string{"ECONOMY_PUBLIC_BASE_URL": "http://h:7350/"}, "/data")
	if err != nil || c.issuer != "http://h:7350/blockworld" || c.keyFile != "/data/blockworld/jwt-rs256.pem" || !c.createKey || c.currency != "bw_coin" {
		t.Fatalf("%+v %v", c, err)
	}
	if _, err := loadConfig(map[string]string{"BLOCKWORLD_ISSUER": "ftp://x/y"}, "/d"); err == nil {
		t.Fatal("ftp issuer accepted")
	}
	if _, err := loadConfig(map[string]string{"BLOCKWORLD_ISSUER": "https://x/" + strings.Repeat("a", 130)}, "/d"); err == nil {
		t.Fatal("long issuer accepted")
	}
}
