package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

const ticketLifetime = 48 * time.Hour

func deriveKey(secret []byte, info string) ([]byte, error) {
	return hkdf.Key(sha256.New, secret, []byte("nakama-economy"), info, 32)
}

func seal(key []byte, plain []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, plain, []byte("economy-payment-method"))
	return base64.StdEncoding.EncodeToString(out), nil
}

func unseal(key []byte, text string) ([]byte, error) {
	if text == "" {
		return nil, nil
	}
	raw, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(raw) < gcm.NonceSize() {
		return nil, errors.New("sealed value too short")
	}
	return gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], []byte("economy-payment-method"))
}

func signTicket(key []byte, paymentID string, expires time.Time) string {
	payload := []byte(paymentID + "|" + strconv.FormatInt(expires.Unix(), 10))
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func readTicket(key []byte, ticket string, now time.Time) (string, bool) {
	parts := strings.Split(ticket, ".")
	if len(parts) != 2 {
		return "", false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	given, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	if subtle.ConstantTimeCompare(mac.Sum(nil), given) != 1 {
		return "", false
	}
	fields := strings.Split(string(payload), "|")
	if len(fields) != 2 {
		return "", false
	}
	seconds, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || !time.Unix(seconds, 0).After(now) {
		return "", false
	}
	if _, err := uuid.FromString(fields[0]); err != nil {
		return "", false
	}
	return fields[0], true
}
