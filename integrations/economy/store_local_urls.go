package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func localURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("%q is not an http(s) URL", raw)
	}
	host := u.Hostname()
	if host != "localhost" {
		if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
			return "", fmt.Errorf("%q must point at localhost or a loopback address", raw)
		}
	}
	return strings.TrimRight(raw, "/"), nil
}

func applyLocalStoreURLs(env map[string]string) error {
	targets := []struct {
		name string
		dst  *string
		path string
	}{
		{"ECONOMY_LOCAL_STRIPE_URL", &stripeAPI, ""},
		{"ECONOMY_LOCAL_GOOGLE_PLAY_URL", &googlePlayBase, "/androidpublisher/v3"},
		{"ECONOMY_LOCAL_BAZAAR_URL", &bazaarBase, ""},
		{"ECONOMY_LOCAL_MYKET_URL", &myketBase, ""},
		{"ECONOMY_LOCAL_APPLE_URL", &appleAPIProduction, ""},
		{"ECONOMY_LOCAL_APPLE_URL", &appleAPISandbox, ""},
		{"ECONOMY_LOCAL_APPLE_URL", &appleReceiptProduction, "/verifyReceipt"},
		{"ECONOMY_LOCAL_APPLE_URL", &appleReceiptSandbox, "/verifyReceipt"},
	}
	for _, t := range targets {
		raw := strings.TrimSpace(env[t.name])
		if raw == "" {
			continue
		}
		base, err := localURL(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", t.name, err)
		}
		*t.dst = base + t.path
	}
	return nil
}
