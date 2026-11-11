package main

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/gofrs/uuid/v5"
)

func pct(v float64) *float64 { return &v }

func TestBucketIsStableSpreadAndMonotonic(t *testing.T) {
	users := make([]string, 20000)
	for i := range users {
		users[i] = uuid.NewV5(uuid.NamespaceOID, strconv.Itoa(i)).String()
	}
	in10, in25 := 0, 0
	for _, u := range users {
		b := bucket("home.banner", u)
		if b != bucket("home.banner", u) || b < 0 || b >= 10000 {
			t.Fatalf("bucket %d for %s is not stable", b, u)
		}
		r10 := configRule{Percent: pct(10)}
		r25 := configRule{Percent: pct(25)}
		a := r10.matches("home.banner", configTarget{UserID: u})
		c := r25.matches("home.banner", configTarget{UserID: u})
		if a && !c {
			t.Fatalf("%s left the rollout when it grew", u)
		}
		if a {
			in10++
		}
		if c {
			in25++
		}
	}
	if in10 < 1800 || in10 > 2200 || in25 < 4700 || in25 > 5300 {
		t.Fatalf("rollout spread 10%%=%d 25%%=%d of 20000", in10, in25)
	}
	same := 0
	for _, u := range users[:2000] {
		if (bucket("a", u) < 5000) == (bucket("b", u) < 5000) {
			same++
		}
	}
	if same < 850 || same > 1150 {
		t.Fatalf("keys are not bucketed independently: %d of 2000 agree", same)
	}
}

func TestConfigRuleTargeting(t *testing.T) {
	k := configKey{Key: "shop.sale", Value: json.RawMessage(`false`), Rules: []configRule{
		{Name: "vip", Value: json.RawMessage(`"vip"`), TagIDs: []string{"tag-vip"}},
		{Name: "new-android", Value: json.RawMessage(`"new"`), Platforms: []string{"android"}, MinVersion: "1.2", MaxVersion: "1.9.9"},
		{Name: "iran", Value: json.RawMessage(`"ir"`), Countries: []string{"IR"}},
	}}
	cases := []struct {
		t    configTarget
		want string
	}{
		{configTarget{Tags: map[string]bool{"tag-vip": true}, Platform: "android", Version: "1.5"}, "vip"},
		{configTarget{Platform: "android", Version: "1.5.0"}, "new-android"},
		{configTarget{Platform: "android", Version: "2.0"}, "default"},
		{configTarget{Platform: "android", Version: "1.2-beta"}, "default"},
		{configTarget{Platform: "ios", Version: "1.5", Country: "IR"}, "iran"},
		{configTarget{Platform: "android"}, "default"},
	}
	for i, c := range cases {
		if _, got := k.resolve(c.t); got != c.want {
			t.Fatalf("case %d: got %s want %s", i, got, c.want)
		}
	}
}
