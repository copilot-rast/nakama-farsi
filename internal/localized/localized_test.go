package localized

import (
	"errors"
	"strings"
	"testing"
)

func testSet() *Set {
	return NewSet([]Language{
		{Code: "fa", Direction: "rtl", Active: true, Default: true},
		{Code: "en", Direction: "ltr", Active: true},
		{Code: "ar", Direction: "rtl", Active: true},
		{Code: "tr", Direction: "ltr"},
	})
}

func TestCleanRules(t *testing.T) {
	s := testSet()
	out, err := s.Clean("name", Text{"fa": "  سلام ", "EN": "Hi", "ar": " ", "tr": "Merhaba"}, true, 10)
	if err != nil || out["fa"] != "سلام" || out["en"] != "Hi" || out["tr"] != "Merhaba" || len(out) != 3 {
		t.Fatalf("clean = %v %v", out, err)
	}
	cases := map[string]struct {
		in       Text
		required bool
		code     string
	}{
		"unknown":          {Text{"fa": "x", "de": "y"}, true, "INVALID_LANGUAGE"},
		"default missing":  {Text{"en": "x"}, true, "DEFAULT_LANGUAGE_REQUIRED"},
		"optional partial": {Text{"en": "x"}, false, "DEFAULT_LANGUAGE_REQUIRED"},
		"too long":         {Text{"fa": strings.Repeat("ب", 11)}, true, "TEXT_TOO_LONG"},
		"required empty":   {Text{}, true, "DEFAULT_LANGUAGE_REQUIRED"},
	}
	for name, tc := range cases {
		_, err := s.Clean("f", tc.in, tc.required, 10)
		var e *Error
		if !errors.As(err, &e) || e.Code != tc.code || e.Field != "f" {
			t.Errorf("%s: err = %v, want %s", name, err, tc.code)
		}
	}
	if out, err := s.Clean("m", nil, false, 10); err != nil || len(out) != 0 {
		t.Fatalf("optional empty = %v %v", out, err)
	}
}

func TestResolveOrder(t *testing.T) {
	s := testSet()
	text := Text{"fa": "فارسی", "en": "English"}
	if got := s.Resolve(text, "en"); got != "English" {
		t.Fatal(got)
	}
	if got := s.Resolve(text, "ar"); got != "فارسی" {
		t.Fatal(got)
	}
	if got := s.Resolve(Text{"ar": "عربي", "tr": "x"}, "en"); got != "عربي" {
		t.Fatal(got)
	}
	if got := s.Resolve(Text{"tr": "x"}, "en"); got != "x" {
		t.Fatal(got)
	}
	if got := s.Resolve(nil, "en"); got != "" {
		t.Fatal(got)
	}
}

func TestChooseLanguage(t *testing.T) {
	s := testSet()
	for in, want := range map[string]string{"AR": "ar", "en-US": "en", "en_gb": "en", "tr": "fa", "xx": "fa", "": "fa"} {
		if got := s.Choose(in); got != want {
			t.Errorf("Choose(%q) = %q, want %q", in, got, want)
		}
	}
	if got := s.Choose("zz", "en"); got != "en" {
		t.Fatalf("fallback candidate = %q", got)
	}
	if s.Direction("ar") != "rtl" || s.Direction("en") != "ltr" || s.Direction("nope") != "ltr" {
		t.Fatal("direction")
	}
}

func TestCodesAndJSON(t *testing.T) {
	for code, ok := range map[string]bool{"fa": true, "pt-br": true, "fil": true, "EN": false, "e": false, "en-": false, "en-us-x": false} {
		if ValidCode(code) != ok {
			t.Errorf("ValidCode(%q) != %v", code, ok)
		}
	}
	if Parse(`{"fa":"x","n":1}`)["fa"] != "x" || len(Parse(`bad`)) != 0 {
		t.Fatal("parse")
	}
	if Text(nil).JSON() != "{}" {
		t.Fatal("json")
	}
	if m := testSet().Missing(Text{"fa": "x"}); len(m) != 2 || m[0] != "en" || m[1] != "ar" {
		t.Fatalf("missing = %v", m)
	}
}
