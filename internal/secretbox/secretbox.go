package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

const (
	SecretEnv          = "ECONOMY_SECRET_KEY"
	MinSecretLength    = 32
	EconomyConfigInfo  = "economy-config-seal"
	EconomyTicketInfo  = "economy-callback-ticket"
	PublicationKeyInfo = "publication-provider-config"
)

type Box struct {
	Salt string
	AAD  string
}

var (
	Economy     = Box{Salt: "nakama-economy", AAD: "economy-payment-method"}
	Publication = Box{Salt: "nakama-publication", AAD: "publication-provider"}
)

func Secret(env map[string]string) ([]byte, error) {
	secret := strings.TrimSpace(env[SecretEnv])
	if len(secret) < MinSecretLength {
		return nil, errors.New(SecretEnv + " must be at least 32 characters")
	}
	return []byte(secret), nil
}

func (b Box) DeriveKey(secret []byte, info string) ([]byte, error) {
	return hkdf.Key(sha256.New, secret, []byte(b.Salt), info, 32)
}

func (b Box) Seal(key []byte, plain []byte) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, plain, []byte(b.AAD))), nil
}

func (b Box) Open(key []byte, text string) ([]byte, error) {
	if text == "" {
		return nil, nil
	}
	raw, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	if len(raw) < gcm.NonceSize() {
		return nil, errors.New("sealed value too short")
	}
	return gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], []byte(b.AAD))
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
