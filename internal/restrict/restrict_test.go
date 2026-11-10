package restrict

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
)

func TestTokenUserID(t *testing.T) {
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"uid":"6d2e7a8c-1111-4222-8333-944455556666","usn":"x"}`))
	if got := TokenUserID("h." + claims + ".s"); got != "6d2e7a8c-1111-4222-8333-944455556666" {
		t.Fatal(got)
	}
	for _, bad := range []string{"", "a.b", "a.%%%.c", "a." + base64.RawURLEncoding.EncodeToString([]byte("[]")) + ".c"} {
		if got := TokenUserID(bad); got != "" {
			t.Fatalf("%q -> %q", bad, got)
		}
	}
}

func TestNoticeError(t *testing.T) {
	n := Notice{Code: CodeSuspended, Kind: KindSuspend, Until: "2026-11-12T10:00:00Z", Reason: "به خاطر فحاشی", Lang: "fa"}
	var rt *runtime.Error
	if !errors.As(n.Error(), &rt) || rt.Code != 7 {
		t.Fatalf("want a permission-denied runtime error, got %v", n.Error())
	}
	var back Notice
	if err := json.Unmarshal([]byte(rt.Message), &back); err != nil || back != n {
		t.Fatalf("round trip %v %+v", err, back)
	}
	if n.Map()["reason"] != "به خاطر فحاشی" {
		t.Fatal(n.Map())
	}
}
