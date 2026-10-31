package main

import "github.com/heroiclabs/nakama/v3/internal/secretbox"

func deriveKey(secret []byte, info string) ([]byte, error) {
	return secretbox.Publication.DeriveKey(secret, info)
}

func seal(key []byte, plain []byte) (string, error) {
	return secretbox.Publication.Seal(key, plain)
}

func unseal(key []byte, text string) ([]byte, error) {
	return secretbox.Publication.Open(key, text)
}
