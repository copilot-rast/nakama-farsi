package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const tokenAudience = "blockworld"

type signer struct {
	key    *rsa.PrivateKey
	kid    string
	issuer string
}

func loadOrCreateKey(path string, create bool) (*rsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) && create {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, err
		}
		der := x509.MarshalPKCS1PrivateKey(key)
		block := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
		if err := os.WriteFile(path, block, 0o600); err != nil {
			return nil, err
		}
		return key, nil
	}
	if err != nil {
		return nil, err
	}
	return parseKey(raw)
}

func parseKey(raw []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("signing key is not PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("signing key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("signing key is not RSA")
	}
	if key.N.BitLen() < 2048 {
		return nil, errors.New("signing key must have at least 2048 bits")
	}
	return key, nil
}

func newSigner(key *rsa.PrivateKey, issuer string) *signer {
	return &signer{key: key, kid: thumbprint(&key.PublicKey), issuer: strings.TrimRight(issuer, "/")}
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func publicJWK(pub *rsa.PublicKey) map[string]string {
	e := make([]byte, 8)
	binary.BigEndian.PutUint64(e, uint64(pub.E))
	e = trimLeadingZeros(e)
	return map[string]string{"kty": "RSA", "n": b64(pub.N.Bytes()), "e": b64(e)}
}

func trimLeadingZeros(b []byte) []byte {
	for len(b) > 1 && b[0] == 0 {
		b = b[1:]
	}
	return b
}

func thumbprint(pub *rsa.PublicKey) string {
	jwk := publicJWK(pub)
	canonical := fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`, jwk["e"], jwk["n"])
	sum := sha256.Sum256([]byte(canonical))
	return b64(sum[:])
}

func (s *signer) jwks() map[string]any {
	k := publicJWK(&s.key.PublicKey)
	return map[string]any{"keys": []map[string]string{{"kty": k["kty"], "n": k["n"], "e": k["e"], "kid": s.kid, "alg": "RS256", "use": "sig"}}}
}

func (s *signer) discovery() map[string]any {
	return map[string]any{
		"issuer":                                s.issuer,
		"jwks_uri":                              s.issuer + "/jwks.json",
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"subject_types_supported":               []string{"public"},
		"response_types_supported":              []string{"id_token"},
		"claims_supported":                      []string{"sub", "iss", "aud", "iat", "exp", "name", "preferred_username"},
	}
}

type worldClaims struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Audience string `json:"aud"`
	IssuedAt int64  `json:"iat"`
	NotBef   int64  `json:"nbf"`
	Expires  int64  `json:"exp"`
	Name     string `json:"name"`
	Username string `json:"preferred_username"`
}

func (s *signer) sign(c worldClaims) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT", "kid": s.kid})
	body, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	input := b64(header) + "." + b64(body)
	sum := sha256.Sum256([]byte(input))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return input + "." + b64(sig), nil
}

func (s *signer) issue(userID, username, display string, ttl time.Duration, now time.Time) (string, worldClaims, error) {
	name := strings.TrimSpace(display)
	if name == "" {
		name = username
	}
	c := worldClaims{
		Issuer:   s.issuer,
		Subject:  userID,
		Audience: tokenAudience,
		IssuedAt: now.Unix(),
		NotBef:   now.Add(-30 * time.Second).Unix(),
		Expires:  now.Add(ttl).Unix(),
		Name:     name,
		Username: username,
	}
	tok, err := s.sign(c)
	return tok, c, err
}

func verify(token string, pub *rsa.PublicKey, issuer string, now time.Time) (worldClaims, error) {
	var c worldClaims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c, errors.New("token must have three parts")
	}
	var header struct {
		Alg string `json:"alg"`
	}
	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(hb, &header) != nil || header.Alg != "RS256" {
		return c, errors.New("bad header")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return c, errors.New("bad signature encoding")
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig); err != nil {
		return c, errors.New("bad signature")
	}
	bb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || json.Unmarshal(bb, &c) != nil {
		return c, errors.New("bad payload")
	}
	if c.Issuer != issuer {
		return c, errors.New("wrong issuer")
	}
	if c.Subject == "" || len(c.Subject) > 128 || len(c.Issuer) > 128 {
		return c, errors.New("bad subject")
	}
	if c.Expires != 0 && now.Unix() > c.Expires+60 {
		return c, errors.New("expired")
	}
	return c, nil
}

func jwkPublicKey(n, e string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, err
	}
	var ev int
	for _, b := range eb {
		ev = ev<<8 | int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: ev}, nil
}
