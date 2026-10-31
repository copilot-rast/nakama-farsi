package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

const ticketLifetime = 48 * time.Hour

func deriveKey(secret []byte, info string) ([]byte, error) {
	return secretbox.Economy.DeriveKey(secret, info)
}

func seal(key []byte, plain []byte) (string, error) {
	return secretbox.Economy.Seal(key, plain)
}

func unseal(key []byte, text string) ([]byte, error) {
	return secretbox.Economy.Open(key, text)
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
