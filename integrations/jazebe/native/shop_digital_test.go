package main

import "testing"

func TestValidDigitalPayload(t *testing.T) {
	for _, kind := range []string{"code", "link", "file"} {
		if !validDigitalPayload(kind, "https://example.test/asset") {
			t.Fatal(kind)
		}
	}
	for _, tc := range []struct{ kind, value string }{{"physical", "x"}, {"code", " "}, {"link", string(make([]byte, 2049))}} {
		if validDigitalPayload(tc.kind, tc.value) {
			t.Fatalf("accepted %+v", tc)
		}
	}
}
