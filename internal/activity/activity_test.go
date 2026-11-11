package activity

import (
	"context"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
)

func TestHeadersGiveCountryAndPlatform(t *testing.T) {
	h := map[string][]string{"Cf-Ipcountry": {"ir"}, "User-Agent": {"Mozilla/5.0 (Linux; Android 14)"}}
	if c := CountryFromHeaders(h); c != "IR" {
		t.Fatalf("country %q", c)
	}
	if p := PlatformFromHeaders(h); p != "android" {
		t.Fatalf("platform %q", p)
	}
	if c := CountryFromHeaders(map[string][]string{"CF-IPCountry": {"XX"}}); c != "" {
		t.Fatalf("unknown country %q", c)
	}
	if p := PlatformFromHeaders(map[string][]string{"X-Client-Platform": {"iOS"}, "User-Agent": {"Mozilla"}}); p != "ios" {
		t.Fatalf("explicit platform %q", p)
	}
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_CLIENT_IP, "10.1.2.3:5555")
	ctx = context.WithValue(ctx, runtime.RUNTIME_CTX_HEADERS, h)
	e := FromContext(ctx, Entry{UserID: "u"})
	if e.IP != "10.1.2.3" || e.Country != "IR" || e.Platform != "android" {
		t.Fatalf("entry %+v", e)
	}
}

func TestKindsAndClip(t *testing.T) {
	if !ValidKind(KindTicket) || ValidKind("other") {
		t.Fatal("kinds")
	}
	if got := clip("سلامسلام", 6); got != "سلا" {
		t.Fatalf("clip %q", got)
	}
}
