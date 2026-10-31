package secretbox

import (
	"strings"
	"testing"
)

const fixtureSecret = "fixture-secret-0123456789abcdef0123456789"

func TestOpensValuesSealedBeforeTheMove(t *testing.T) {
	for _, tc := range []struct {
		box    Box
		info   string
		sealed string
		plain  string
	}{
		{Economy, EconomyConfigInfo, "nFfhlBksAY9cri24CEbVo2CZlCwnq2939tVgn7rretyzT6qUTN7BlOg1sq4K6zrZXqWVXTbSs2SoXAHbtNzxRUzb", `{"merchantId":"m-123","sandbox":false}`},
		{Publication, PublicationKeyInfo, "s+kjOf2dNTQuP9zjiBbJITIHMocUc4FKNZJ6CDXp+x1lOKPnL6kPj+rT/sj0hg==", `{"apiKey":"k-456"}`},
	} {
		key, err := tc.box.DeriveKey([]byte(fixtureSecret), tc.info)
		if err != nil {
			t.Fatal(err)
		}
		plain, err := tc.box.Open(key, tc.sealed)
		if err != nil || string(plain) != tc.plain {
			t.Fatalf("%s: open = %q %v", tc.info, plain, err)
		}
	}
}

func TestRoundTripAndIsolation(t *testing.T) {
	key, _ := Economy.DeriveKey([]byte(fixtureSecret), EconomyConfigInfo)
	sealed, err := Economy.Seal(key, []byte(`{"merchant":"secret"}`))
	if err != nil || strings.Contains(sealed, "secret") {
		t.Fatalf("seal = %q %v", sealed, err)
	}
	if plain, err := Economy.Open(key, sealed); err != nil || string(plain) != `{"merchant":"secret"}` {
		t.Fatalf("open = %q %v", plain, err)
	}
	if _, err := Publication.Open(key, sealed); err == nil {
		t.Fatal("another AAD must not open the value")
	}
	ticket, _ := Economy.DeriveKey([]byte(fixtureSecret), EconomyTicketInfo)
	if _, err := Economy.Open(ticket, sealed); err == nil {
		t.Fatal("another info must not open the value")
	}
	if plain, err := Economy.Open(key, ""); err != nil || plain != nil {
		t.Fatalf("empty = %q %v", plain, err)
	}
	if _, err := Economy.Open(key, "AAAA"); err == nil {
		t.Fatal("short value must fail")
	}
}

func TestSecret(t *testing.T) {
	if _, err := Secret(map[string]string{}); err == nil {
		t.Fatal("missing secret must fail")
	}
	if _, err := Secret(map[string]string{SecretEnv: "  short  "}); err == nil {
		t.Fatal("short secret must fail")
	}
	got, err := Secret(map[string]string{SecretEnv: " " + fixtureSecret + " "})
	if err != nil || string(got) != fixtureSecret {
		t.Fatalf("secret = %q %v", got, err)
	}
}
