package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func TestNormalizePhone(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		ok       bool
	}{
		{"09121234567", "+989121234567", true},
		{"۰۹۱۲۱۲۳۴۵۶۷", "+989121234567", true},
		{"0912 123 4567", "+989121234567", true},
		{"+989121234567", "+989121234567", true},
		{"989121234567", "+989121234567", true},
		{"00989121234567", "+989121234567", true},
		{"9121234567", "+989121234567", true},
		{"+14155550100", "+14155550100", true},
		{"0912123456", "", false},
		{"98912", "", false},
		{"+98", "", false},
		{"abc", "", false},
		{"", "", false},
		{"14155550100", "", false},
	} {
		got, ok := normalizePhone(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("normalizePhone(%q) = %q,%v want %q,%v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
	if local, ok := iranLocal("+989121234567"); !ok || local != "09121234567" {
		t.Errorf("iranLocal = %q,%v", local, ok)
	}
	if _, ok := iranLocal("+14155550100"); ok {
		t.Error("iranLocal must refuse foreign numbers")
	}
}

func TestNormalizeEmail(t *testing.T) {
	for in, ok := range map[string]bool{
		"a@example.com":         true,
		"A.B+c@sub.example.ir":  true,
		"Name <a@example.com>":  false,
		"a@b":                   false,
		"a@example.com\r\nBcc:": false,
		"":                      false,
		"a,b@example.com":       false,
	} {
		if _, got := normalizeEmail(in); got != ok {
			t.Errorf("normalizeEmail(%q) ok=%v want %v", in, got, ok)
		}
	}
}

func testLangs() *localized.Set {
	return localized.NewSet([]localized.Language{{Code: "fa", Direction: "rtl", Active: true, Default: true}, {Code: "en", Active: true}, {Code: "ar", Direction: "rtl", Active: true}})
}

func TestCleanAction(t *testing.T) {
	if _, err := cleanAction(testLangs(), "a", action{Type: "url", Value: "javascript:alert(1)"}, false); err == nil {
		t.Error("javascript url must be refused")
	}
	if _, err := cleanAction(testLangs(), "a", action{Type: "url", Value: "https://e.com"}, true); err == nil {
		t.Error("label is required for news actions")
	}
	if a, err := cleanAction(testLangs(), "a", action{Type: "goto", Value: "shop", Label: i18n{"fa": "فروشگاه", "ar": "متجر"}}, true); err != nil || a.Label["fa"] == "" {
		t.Errorf("goto action refused: %v", err)
	}
	if _, err := cleanAction(testLangs(), "a", action{Type: "open", Value: "x"}, false); err == nil {
		t.Error("unknown type must be refused")
	}
	if a, err := cleanAction(testLangs(), "a", action{Type: "goto", Value: "shop"}, false); err != nil || a.Label != nil {
		t.Errorf("optional label = %v %v", a, err)
	}
	if _, err := cleanAction(testLangs(), "a", action{Type: "goto", Value: "shop", Label: i18n{"de": "x"}}, false); err == nil {
		t.Error("unknown label language must be refused")
	}
}

func TestCleanNotification(t *testing.T) {
	long := string(bytes.Repeat([]byte("a"), 601))
	for name, tc := range map[string]struct {
		in   notificationInput
		code string
	}{
		"bad channel":      {notificationInput{Channel: "push", Body: i18n{"fa": "x"}}, "INVALID_CHANNEL"},
		"sms too long":     {notificationInput{Channel: "sms", Body: i18n{"fa": long}}, "TEXT_TOO_LONG"},
		"email no title":   {notificationInput{Channel: "email", Body: i18n{"fa": "x"}}, "DEFAULT_LANGUAGE_REQUIRED"},
		"in_app no title":  {notificationInput{Channel: "in_app", Body: i18n{"fa": "x"}}, "DEFAULT_LANGUAGE_REQUIRED"},
		"no body":          {notificationInput{Channel: "sms"}, "DEFAULT_LANGUAGE_REQUIRED"},
		"provider on push": {notificationInput{Channel: "in_app", Title: i18n{"fa": "t"}, Body: i18n{"fa": "x"}, ProviderID: "x"}, "PROVIDER_NOT_ALLOWED"},
		"bad action":       {notificationInput{Channel: "in_app", Title: i18n{"fa": "t"}, Body: i18n{"fa": "x"}, Action: &action{Type: "url", Value: "ftp://x"}}, "INVALID_ACTION"},
		"en only":          {notificationInput{Channel: "in_app", Title: i18n{"en": "t"}, Body: i18n{"fa": "x"}}, "DEFAULT_LANGUAGE_REQUIRED"},
		"unknown language": {notificationInput{Channel: "in_app", Title: i18n{"fa": "t", "de": "t"}, Body: i18n{"fa": "x"}}, "INVALID_LANGUAGE"},
	} {
		_, err := cleanNotification(testLangs(), tc.in)
		app, ok := asApp(err)
		if !ok || app.Code != tc.code {
			t.Errorf("%s: err=%v want %s", name, err, tc.code)
		}
	}
	if _, err := cleanNotification(testLangs(), notificationInput{Channel: "sms", Body: i18n{"fa": string(bytes.Repeat([]byte("a"), 600))}}); err != nil {
		t.Errorf("600 char sms without title must pass: %v", err)
	}
}

func TestCleanTargetShape(t *testing.T) {
	s := &service{}
	ctx := context.Background()
	for name, tc := range map[string]struct {
		in   target
		code string
	}{
		"bad mode":    {target{Mode: "everyone"}, "INVALID_TARGET"},
		"no users":    {target{Mode: "users"}, "TARGET_USERS_REQUIRED"},
		"bad user id": {target{Mode: "users", UserIDs: []string{"x"}}, "INVALID_ID"},
		"system user": {target{Mode: "users", UserIDs: []string{systemUserID}}, "USERS_NOT_FOUND"},
		"no group":    {target{Mode: "group"}, "REQUIRED"},
		"bad group":   {target{Mode: "group", GroupID: "nope"}, "INVALID_ID"},
	} {
		_, err := s.cleanTarget(ctx, nil, tc.in)
		app, ok := asApp(err)
		if !ok || app.Code != tc.code {
			t.Errorf("%s: err=%v want %s", name, err, tc.code)
		}
	}
	ids := make([]string, 1001)
	for i := range ids {
		ids[i] = "00000000-0000-4000-8000-" + padHex(i+1)
	}
	if _, err := s.cleanTarget(ctx, nil, target{Mode: "users", UserIDs: ids}); err == nil {
		t.Error("more than 1000 users must be refused")
	} else if app, _ := asApp(err); app.Code != "TOO_MANY_USERS" {
		t.Errorf("got %v", err)
	}
}

func padHex(n int) string {
	const digits = "0123456789abcdef"
	out := []byte("000000000000")
	for i := len(out) - 1; n > 0 && i >= 0; i-- {
		out[i] = digits[n%16]
		n /= 16
	}
	return string(out)
}

func TestSealRoundTrip(t *testing.T) {
	key, err := deriveKey([]byte("0123456789abcdef0123456789abcdef"), "publication-provider-config")
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := seal(key, []byte(`{"apiKey":"k"}`))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := unseal(key, sealed)
	if err != nil || string(plain) != `{"apiKey":"k"}` {
		t.Fatalf("unseal = %q %v", plain, err)
	}
	other, _ := deriveKey([]byte("0123456789abcdef0123456789abcdef"), "economy-config-seal")
	if _, err := unseal(other, sealed); err == nil {
		t.Fatal("a key with another HKDF info must not open the config")
	}
	if _, err := loadSettings(map[string]string{"ECONOMY_SECRET_KEY": "short"}); err == nil {
		t.Fatal("short secret must be refused")
	}
}

func TestProviderConfigMerge(t *testing.T) {
	def, _ := catalogFor("sms", "twilio")
	cfg, err := mergeProviderConfig(def, providerConfig{}, map[string]any{"accountSid": "AC1", "authToken": "tok", "from": "+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	view := providerView(provider{Channel: "sms", Kind: "twilio", Config: cfg})
	masked := view["config"].(map[string]any)
	if masked["authToken"] != secretMask || masked["accountSid"] != "AC1" {
		t.Fatalf("view config = %v", masked)
	}
	again, err := mergeProviderConfig(def, cfg, map[string]any{"authToken": secretMask, "accountSid": "AC2"})
	if err != nil || again["authToken"] != "tok" || again["accountSid"] != "AC2" || again["from"] != "+14155550100" {
		t.Fatalf("masked secret must keep the old value: %v %v", again, err)
	}
	if _, err := mergeProviderConfig(def, cfg, map[string]any{"from": "not a number"}); err == nil {
		t.Fatal("bad twilio from must be refused")
	}
	smtpDef, _ := catalogFor("email", "smtp")
	sc, err := mergeProviderConfig(smtpDef, providerConfig{}, map[string]any{"host": "smtp.example.com", "from": "a@example.com"})
	if err != nil || sc["port"] != "587" || sc["security"] != "starttls" {
		t.Fatalf("defaults not applied: %v %v", sc, err)
	}
	for field, bad := range map[string]any{"port": "70000", "security": "ssl", "host": "smtp://x", "from": "nope"} {
		if _, err := mergeProviderConfig(smtpDef, sc, map[string]any{field: bad}); err == nil {
			t.Errorf("%s=%v must be refused", field, bad)
		}
	}
	if !providerConfigured(provider{Channel: "sms", Kind: "smsir", Sandbox: true, Config: providerConfig{"lineNumber": "3000"}}) {
		t.Error("sandbox provider does not need secrets")
	}
	if providerConfigured(provider{Channel: "sms", Kind: "smsir", Config: providerConfig{"lineNumber": "3000"}}) {
		t.Error("live provider needs its api key")
	}
}
