package chatfilter

import (
	"strings"
	"testing"
)

var testRules = []Rule{
	{ID: "1", Word: "احمق", Action: ActionMask, Match: MatchWord},
	{ID: "2", Word: "بی‌شعور", Action: ActionMask, Match: MatchWord},
	{ID: "3", Word: "خفه شو", Action: ActionMask, Match: MatchWord},
	{ID: "4", Word: "idiot", Action: ActionMask, Match: MatchWord},
	{ID: "5", Word: "ass", Action: ActionMask, Match: MatchWord},
	{ID: "6", Word: "حرامزاده", Action: ActionBlock, Match: MatchAny},
	{ID: "7", Word: "تلگرام", Action: ActionFlag, Match: MatchWord},
	{ID: "8", Word: "كلك", Action: ActionMask, Match: MatchWord},
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"علي":       "علی",
		"كتاب":      "کتاب",
		"مي‌خواهم":  "میخواهم",
		"احـــــمق": "احمق",
		"اَحْمَق":   "احمق",
		"۱۲۳ و ١٢٣": "و",
		"IDIOT":     "idiot",
		"i.d.i.o.t": "idiot",
		"ا ح م ق":   "احمق",
		"آب":        "اب",
		"سلااااام":  "سلام",
		"ﺍﺣﻤﻖ":      "احمق",
		"ｉｄｉｏｔ":     "idiot",
		"ідіот":     "idiot",
		"1d10t":     "idiot",
		"خونه‌ای":   "خونهای",
		"‏احمق‌‌":   "احمق",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMaskCatchesEvasion(t *testing.T) {
	f := Compile(testRules)
	cases := []string{
		"تو احمق هستی",
		"تو احمقی",
		"احمق‌ها رفتند",
		"تو احـــمق هستی",
		"تو اَحمَق هستی",
		"تو احمققققق هستی",
		"تو ا ح م ق هستی",
		"تو ا.ح.م.ق هستی",
		"تو ا_ح_م_ق هستی",
		"تو اح‌مق هستی",
		"تو ﺍﺣﻤﻖ هستی",
		"بی‌شعور",
		"بیشعور",
		"بي شعور",
		"خفه شو دیگه",
		"خفهههه شووو",
		"you IDIOT",
		"you i d i o t",
		"you 1d10t",
		"you idiots",
		"you ідіот",
		"you ｉｄｉｏｔ",
		"آسس asss",
		"این كلك است",
		"این کلک است",
		"تو احمق!!!",
	}
	for _, in := range cases {
		res := f.Apply(in)
		if !res.Changed || res.Blocked || !strings.Contains(res.Text, "*") {
			t.Errorf("Apply(%q) = %+v, want masked", in, res)
		}
	}
}

func TestMaskKeepsTheRest(t *testing.T) {
	f := Compile(testRules)
	cases := map[string]string{
		"تو احمق هستی":   "تو **** هستی",
		"you idiot!":     "you *****!",
		"i.d.i.o.t":      "*.*.*.*.*",
		"تو اح‌مق هستی":  "تو **** هستی",
		"you are an ass": "you are an ***",
	}
	for in, want := range cases {
		if got := f.Apply(in).Text; got != want {
			t.Errorf("Apply(%q).Text = %q, want %q", in, got, want)
		}
	}
}

func TestNoFalsePositives(t *testing.T) {
	f := Compile(testRules)
	cases := []string{
		"سلام دوستان",
		"احمد آمد",
		"as soon as possible",
		"classic class pass",
		"شعور داری",
		"خفه نشو",
		"idiom",
		"خانه‌ی ما",
		"",
		"۱۲۳۴",
	}
	for _, in := range cases {
		if res := f.Apply(in); res.Changed || res.Blocked || res.Flagged || len(res.Hits) > 0 {
			t.Errorf("Apply(%q) = %+v, want untouched", in, res)
		}
	}
}

func TestBlockAndFlag(t *testing.T) {
	f := Compile(testRules)
	for _, in := range []string{"حرامزاده", "تو حروم حرامزادهای", "ح ر ا م ز ا د ه", "حرام‌زاده", "حـرامـزاده", "حراااامزاده", "یه‌حرامزاده"} {
		if res := f.Apply(in); !res.Blocked {
			t.Errorf("Apply(%q) not blocked: %+v", in, res)
		}
	}
	res := f.Apply("بیا تلگرام")
	if !res.Flagged || res.Changed || res.Blocked || len(res.Hits) != 1 || res.Hits[0].Action != ActionFlag {
		t.Fatalf("flag: %+v", res)
	}
	res = f.Apply("احمق بیا تلگرام")
	if !res.Flagged || !res.Changed || len(res.Hits) != 2 {
		t.Fatalf("mask + flag: %+v", res)
	}
}

func TestEmptyFilter(t *testing.T) {
	var f *Filter
	if res := f.Apply("hello"); res.Text != "hello" || res.Changed {
		t.Fatal(res)
	}
	if Compile([]Rule{{Word: "‌‌", Action: ActionMask}}).Len() != 0 {
		t.Fatal("a word with no letters must be skipped")
	}
}
