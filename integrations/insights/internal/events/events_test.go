package events

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func raw(t *testing.T, s string) Raw {
	t.Helper()
	var r Raw
	if err := json.Unmarshal([]byte(s), &r); err != nil {
		t.Fatal(err)
	}
	return r
}

var now = time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)

func catalog() *Catalog {
	lo, hi := 0.0, 10000.0
	return &Catalog{
		Groups: map[string]Group{
			"car":     {Key: "car", Enabled: true, Closed: true, Values: map[string]bool{"345345": true}},
			"mission": {Key: "mission", Enabled: true},
		},
		Events: map[string]Def{
			"DistanceOverall": {Name: "DistanceOverall", Enabled: true, PathGroups: []string{"car", "mission"}, Value: ValueRule{Min: &lo, Max: &hi}},
			"Off":             {Name: "Off", Enabled: false},
			"Need":            {Name: "Need", Enabled: true, RequiredTags: []string{"mission"}, Value: ValueRule{Required: true}},
			"Narrow":          {Name: "Narrow", Enabled: true, AllowedTags: []string{"car"}},
		},
	}
}

func TestOwnerSketchListAndMapTags(t *testing.T) {
	p, r := Prepare(0, raw(t, `{"name":"DistanceOverall","tags":[{"tagGroup":"mission","id":"3434343"},{"tagGroup":"car","id":"345345"}],"value":150}`), catalog(), DefaultPolicy(), now)
	if r.Status != StatusStored || p.Name != "DistanceOverall" || p.Tags["mission"] != "3434343" || p.Tags["car"] != "345345" || *p.Value != 150 {
		t.Fatalf("%+v %+v", p, r)
	}
	p, r = Prepare(1, raw(t, `{"name":"DistanceOverall","tags":{"mission":3434343,"car":"345345"},"value":"12.5"}`), catalog(), DefaultPolicy(), now)
	if r.Status != StatusStored || p.Tags["mission"] != "3434343" || *p.Value != 12.5 {
		t.Fatalf("map tags %+v %+v", p, r)
	}
}

func TestLegacyPathUsesCatalogGroups(t *testing.T) {
	p, r := Prepare(0, raw(t, `{"path":"DistanceOverall:345345:3434343","value":150}`), catalog(), DefaultPolicy(), now)
	if r.Status != StatusStored || p.Tags["car"] != "345345" || p.Tags["mission"] != "3434343" {
		t.Fatalf("%+v %+v", p, r)
	}
	p, r = Prepare(0, raw(t, `{"path":"DistanceOverall::3434343","value":1}`), catalog(), DefaultPolicy(), now)
	if r.Status != StatusStored || p.Tags["mission"] != "3434343" || len(p.Tags) != 1 {
		t.Fatalf("skip slot %+v %+v", p, r)
	}
	p, r = Prepare(0, raw(t, `{"path":"Crash:Wall:Big"}`), catalog(), DefaultPolicy(), now)
	if r.Status != StatusFlagged || r.Code != "ANALYTICS_PATH_GROUPS" || p.Tags["p1"] != "Wall" || p.Tags["p2"] != "Big" {
		t.Fatalf("unknown path %+v %+v", p, r)
	}
	if _, r = Prepare(0, raw(t, `{"path":"A:1:2:3:4:5:6"}`), catalog(), DefaultPolicy(), now); r.Code != "ANALYTICS_TOO_MANY_TAGS" {
		t.Fatalf("long path %+v", r)
	}
	if _, r = Prepare(0, raw(t, `{"type":"design","id":"DistanceOverall:345345","value":3}`), catalog(), DefaultPolicy(), now); r.Status != StatusStored {
		t.Fatalf("design %+v", r)
	}
}

func TestTypedEvents(t *testing.T) {
	p, r := Prepare(0, raw(t, `{"type":"progression","status":"Complete","p1":"Chapter1","p2":"Stage3","score":1200}`), catalog(), DefaultPolicy(), now)
	if r.Status != StatusStored || p.Name != "progression" || p.Tags["status"] != "complete" || p.Tags["progression2"] != "Stage3" || *p.Value != 1200 {
		t.Fatalf("progression %+v %+v", p, r)
	}
	p, _ = Prepare(0, raw(t, `{"type":"resource","flow":"Sink","currency":"coin","amount":50,"itemType":"Shop","item":"Car1"}`), catalog(), DefaultPolicy(), now)
	if p.Name != "resource" || *p.Value != -50 || p.Tags["item_type"] != "Shop" {
		t.Fatalf("resource %+v", p)
	}
	p, _ = Prepare(0, raw(t, `{"type":"business","currency":"IRR","amount":5000,"itemType":"coins","item":"pack1","cartType":"shop","store":"bazaar"}`), catalog(), DefaultPolicy(), now)
	if p.Name != "business" || p.Tags["store"] != "bazaar" || *p.Value != 5000 {
		t.Fatalf("business %+v", p)
	}
	p, _ = Prepare(0, raw(t, `{"type":"error","severity":"Critical","message":"boom"}`), catalog(), DefaultPolicy(), now)
	if p.Name != "error" || p.Tags["severity"] != "critical" || p.Fields["message"] != "boom" || p.Value != nil {
		t.Fatalf("error %+v", p)
	}
	for _, bad := range []string{`{"type":"progression","status":"x","p1":"a"}`, `{"type":"resource","flow":"in","currency":"c","amount":1}`, `{"type":"nope"}`, `{"type":"design"}`} {
		if _, r := Prepare(0, raw(t, bad), catalog(), DefaultPolicy(), now); r.Status != StatusRejected {
			t.Fatalf("accepted %s: %+v", bad, r)
		}
	}
}

func TestValidationAndPolicies(t *testing.T) {
	pol := DefaultPolicy()
	cases := map[string]string{
		`{"name":"1bad"}`: "ANALYTICS_INVALID_EVENT_NAME",
		`{"name":"A","tags":[{"tagGroup":"Car","id":"1"}]}`:                         "ANALYTICS_INVALID_TAG",
		`{"name":"A","tags":[{"tagGroup":"car","id":"a:b"}]}`:                       "ANALYTICS_INVALID_TAG",
		`{"name":"A","tags":[{"tagGroup":"x","id":"1"},{"tagGroup":"x","id":"2"}]}`: "ANALYTICS_DUPLICATE_TAG_GROUP",
		`{"name":"A","value":"abc"}`:                                                "ANALYTICS_INVALID_VALUE",
		`{"name":"A","fields":{"k":{"x":1}}}`:                                       "ANALYTICS_INVALID_FIELDS",
		`{"name":"Need","value":1}`:                                                 "ANALYTICS_MISSING_TAG",
		`{"name":"Need","tags":{"mission":"1"}}`:                                    "ANALYTICS_VALUE_REQUIRED",
		`{"name":"Narrow","tags":{"mission":"1"}}`:                                  "ANALYTICS_UNKNOWN_TAG_GROUP",
		`{"name":"DistanceOverall","value":20000}`:                                  "ANALYTICS_VALUE_OUT_OF_RANGE",
		`{"name":"A","value":2e13}`:                                                 "ANALYTICS_VALUE_OUT_OF_RANGE",
	}
	for in, code := range cases {
		if _, r := Prepare(0, raw(t, in), catalog(), pol, now); r.Status != StatusRejected || r.Code != code {
			t.Fatalf("%s: %+v want %s", in, r, code)
		}
	}
	pol.MaxTags = 1
	if _, r := Prepare(0, raw(t, `{"name":"A","tags":{"a":"1","b":"2"}}`), catalog(), pol, now); r.Code != "ANALYTICS_TOO_MANY_TAGS" {
		t.Fatalf("max tags %+v", r)
	}
	if _, r := Prepare(0, raw(t, `{"name":"Off"}`), catalog(), DefaultPolicy(), now); r.Status != StatusDropped {
		t.Fatalf("disabled %+v", r)
	}
	if _, r := Prepare(0, raw(t, `{"name":"DistanceOverall","tags":{"car":"999"}}`), catalog(), DefaultPolicy(), now); r.Status != StatusFlagged || r.Code != "ANALYTICS_UNKNOWN_TAG_VALUE" {
		t.Fatalf("closed group %+v", r)
	}
	pol = DefaultPolicy()
	pol.UnknownEvents, pol.UnknownTagGroups, pol.UnknownTagValues, pol.ValueOutOfRange = PolicyReject, PolicyReject, PolicyReject, PolicyFlag
	if _, r := Prepare(0, raw(t, `{"name":"Other"}`), catalog(), pol, now); r.Code != "ANALYTICS_UNKNOWN_EVENT" || r.Status != StatusRejected {
		t.Fatalf("unknown event %+v", r)
	}
	if _, r := Prepare(0, raw(t, `{"name":"DistanceOverall","tags":{"weather":"rain"}}`), catalog(), pol, now); r.Code != "ANALYTICS_UNKNOWN_TAG_GROUP" {
		t.Fatalf("unknown group %+v", r)
	}
	if _, r := Prepare(0, raw(t, `{"name":"DistanceOverall","tags":{"car":"9"}}`), catalog(), pol, now); r.Code != "ANALYTICS_UNKNOWN_TAG_VALUE" || r.Status != StatusRejected {
		t.Fatalf("closed reject %+v", r)
	}
	if _, r := Prepare(0, raw(t, `{"name":"DistanceOverall","value":-1}`), catalog(), pol, now); r.Status != StatusFlagged || r.Code != "ANALYTICS_VALUE_OUT_OF_RANGE" {
		t.Fatalf("range flag %+v", r)
	}
	pol.UnknownEvents = PolicyFlag
	if _, r := Prepare(0, raw(t, `{"name":"Other"}`), catalog(), pol, now); r.Status != StatusFlagged {
		t.Fatalf("unknown flag %+v", r)
	}
}

func TestClientTimeWindow(t *testing.T) {
	p, r := Prepare(0, raw(t, `{"name":"A","time":"2026-11-21T11:00:00Z"}`), nil, DefaultPolicy(), now)
	if r.Status != StatusStored || p.ClientTime == nil || p.ClientTime.Hour() != 11 {
		t.Fatalf("%+v %+v", p, r)
	}
	ms := now.Add(-time.Hour).UnixMilli()
	p, _ = Prepare(0, raw(t, `{"name":"A","time":`+strconv.FormatInt(ms, 10)+`}`), nil, DefaultPolicy(), now)
	if p.ClientTime == nil || !p.ClientTime.Equal(time.UnixMilli(ms)) {
		t.Fatalf("ms %+v", p.ClientTime)
	}
	p, r = Prepare(0, raw(t, `{"name":"A","time":"2025-01-01T00:00:00Z"}`), nil, DefaultPolicy(), now)
	if r.Status != StatusFlagged || r.Code != "ANALYTICS_CLIENT_TIME_OUT_OF_RANGE" || p.ClientTime != nil {
		t.Fatalf("old %+v %+v", p, r)
	}
}
