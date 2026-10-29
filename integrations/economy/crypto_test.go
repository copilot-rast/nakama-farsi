package main

import (
	"strings"
	"testing"
	"time"
)

func TestSealRoundTrip(t *testing.T) {
	s := testService()
	sealed, err := seal(s.cfg.sealKey, []byte(`{"merchantId":"secret"}`))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sealed, "secret") {
		t.Fatal("ciphertext contains plaintext")
	}
	plain, err := unseal(s.cfg.sealKey, sealed)
	if err != nil || string(plain) != `{"merchantId":"secret"}` {
		t.Fatalf("unseal = %s %v", plain, err)
	}
	other, _ := deriveKey([]byte(strings.Repeat("cd", 24)), "economy-config-seal")
	if _, err := unseal(other, sealed); err == nil {
		t.Fatal("another key must not open the value")
	}
	if string(s.cfg.sealKey) == string(s.cfg.ticketKey) {
		t.Fatal("seal and ticket keys must differ")
	}
}

func TestTickets(t *testing.T) {
	s := testService()
	now := time.Now()
	id := "3f1c2a10-1111-4222-8333-444455556666"
	ticket := signTicket(s.cfg.ticketKey, id, now.Add(ticketLifetime))
	if got, ok := readTicket(s.cfg.ticketKey, ticket, now); !ok || got != id {
		t.Fatalf("valid ticket rejected: %s %v", got, ok)
	}
	if _, ok := readTicket(s.cfg.ticketKey, ticket, now.Add(49*time.Hour)); ok {
		t.Fatal("expired ticket accepted")
	}
	parts := strings.Split(ticket, ".")
	forged := signTicket([]byte("another-key-another-key-another!!"), id, now.Add(time.Hour))
	for _, bad := range []string{"", "abc", parts[0], parts[0] + ".", parts[0] + "." + parts[1] + "x", strings.Split(forged, ".")[0] + "." + parts[1], forged, parts[0] + "." + parts[1] + ".x"} {
		if _, ok := readTicket(s.cfg.ticketKey, bad, now); ok {
			t.Fatalf("forged ticket %q accepted", bad)
		}
	}
	notID := signTicket(s.cfg.ticketKey, "not-a-uuid", now.Add(time.Hour))
	if _, ok := readTicket(s.cfg.ticketKey, notID, now); ok {
		t.Fatal("non-uuid payload accepted")
	}
}

func TestLoadSettings(t *testing.T) {
	if _, err := loadSettings(map[string]string{"ECONOMY_PUBLIC_BASE_URL": "http://x"}, nil); err == nil {
		t.Fatal("missing secret must fail")
	}
	if _, err := loadSettings(map[string]string{"ECONOMY_SECRET_KEY": "short", "ECONOMY_PUBLIC_BASE_URL": "http://x"}, nil); err == nil {
		t.Fatal("short secret must fail")
	}
	key := strings.Repeat("a", 32)
	if _, err := loadSettings(map[string]string{"ECONOMY_SECRET_KEY": key}, nil); err == nil {
		t.Fatal("missing base URL must fail")
	}
	cfg, err := loadSettings(map[string]string{"ECONOMY_SECRET_KEY": key, "ECONOMY_PUBLIC_BASE_URL": "http://h:7450/", "ECONOMY_PENDING_TTL_MINUTES": "30"}, []string{"nakama", "--data_dir", "/tmp/d"})
	if err != nil || cfg.publicBase != "http://h:7450" || cfg.pendingTTL != 30*time.Minute || cfg.uploadDir != "/tmp/d/economy-uploads" || cfg.deepLink != defaultDeepLink {
		t.Fatalf("cfg = %+v %v", cfg, err)
	}
	if _, err := loadSettings(map[string]string{"ECONOMY_SECRET_KEY": key, "ECONOMY_PUBLIC_BASE_URL": "http://h", "ECONOMY_PENDING_TTL_MINUTES": "1"}, nil); err == nil {
		t.Fatal("tiny TTL must fail")
	}
	s := &service{cfg: cfg}
	if got := s.redirectTarget(true, "p1"); got != "nakama://payment?success=true&paymentId=p1" {
		t.Fatalf("deep link = %s", got)
	}
	s.cfg.webRedirect = "https://web/pay?ok={success}&id={paymentId}"
	if got := s.redirectTarget(false, "p2"); got != "https://web/pay?ok=false&id=p2" {
		t.Fatalf("web redirect = %s", got)
	}
}
