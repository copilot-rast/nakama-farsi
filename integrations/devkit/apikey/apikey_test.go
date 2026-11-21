package apikey

import (
	"net/http"
	"testing"
)

func TestGenerateParseHash(t *testing.T) {
	full, lookup, hash, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	got, ok := Parse(full)
	if !ok || got != lookup || len(hash) != 32 {
		t.Fatalf("parse %q -> %q %v", full, got, ok)
	}
	if string(Hash(full)) != string(hash) {
		t.Fatal("hash mismatch")
	}
	if _, ok := Parse("nfk_short_x"); ok {
		t.Fatal("bad key parsed")
	}
}

func TestAllows(t *testing.T) {
	cases := []struct {
		granted []string
		need    string
		wild    bool
		want    bool
	}{
		{[]string{"assets:write"}, "assets:write", true, true},
		{[]string{"assets:read"}, "assets:write", true, false},
		{[]string{"assets:*"}, "assets:write", true, true},
		{[]string{"assets:*"}, "assets:write", false, false},
		{[]string{"*"}, "cards:write", true, true},
		{[]string{}, "", true, true},
	}
	for _, c := range cases {
		if Allows(c.granted, c.need, c.wild) != c.want {
			t.Errorf("Allows(%v,%q,%v) != %v", c.granted, c.need, c.wild, c.want)
		}
	}
	if !ValidGrant("assets:*", true) || ValidGrant("assets:*", false) || !ValidGrant("a-b:c-d", false) || ValidGrant("Bad:Scope", true) {
		t.Fatal("grant validation")
	}
}

func TestFromRequest(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer nfk_abc")
	if FromRequest(r) != "nfk_abc" {
		t.Fatal("bearer")
	}
	r.Header.Set("Authorization", "Bearer eyJsession")
	if FromRequest(r) != "" {
		t.Fatal("session bearer must be ignored")
	}
	r.Header.Set(HeaderName, "nfk_x")
	if FromRequest(r) != "nfk_x" {
		t.Fatal("header")
	}
	if !ipAllowed([]string{"10.0.0.0/8"}, "10.1.2.3") || ipAllowed([]string{"10.0.0.0/8"}, "192.168.1.1") || !ipAllowed(nil, "1.1.1.1") {
		t.Fatal("ip allow list")
	}
}
