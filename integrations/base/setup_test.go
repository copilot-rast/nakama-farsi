package main

import "testing"

func TestCleanAccent(t *testing.T) {
	for in, want := range map[string]string{"": "", "#6B59EE": "#6b59ee", "6b59ee": "#6b59ee", "#0ea": "#00eeaa", " #abcdef ": "#abcdef"} {
		got, err := cleanAccent(in)
		if err != nil || got != want {
			t.Errorf("cleanAccent(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"red", "#12345", "#ggg", "rgb(1,2,3)", "#1234567"} {
		if _, err := cleanAccent(in); err == nil {
			t.Errorf("cleanAccent(%q) must fail", in)
		}
	}
}

func TestApplySetupTransitions(t *testing.T) {
	str := func(s string) *string { return &s }
	cur := setupState{Status: "in_progress", Steps: map[string]string{"brand": "done"}}
	next, err := applySetup(cur, setupInput{Steps: map[string]string{"languages": "skipped", "brand": "pending"}})
	if err != nil || next.Steps["brand"] != "" || next.Steps["languages"] != "skipped" || next.NextStep != "brand" || cur.Steps["brand"] != "done" {
		t.Fatalf("steps merge = %+v %v", next, err)
	}
	done := setupState{Status: "completed", Steps: map[string]string{}}
	for _, s := range []string{"skipped", "in_progress"} {
		next, _ = applySetup(done, setupInput{Status: str(s)})
		if next.Status != "completed" {
			t.Fatalf("%s must not downgrade a completed setup", s)
		}
	}
	all := map[string]string{}
	for _, k := range setupSteps {
		all[k] = "done"
	}
	if nextSetupStep(all) != "review" {
		t.Fatal("all steps done must point at review")
	}
	if _, err := applySetup(cur, setupInput{Status: str("finished")}); err == nil {
		t.Fatal("unknown status must fail")
	}
}

func TestKeyTypes(t *testing.T) {
	for _, list := range [][]string{categoryKeyTypes, tagKeyTypes} {
		for _, kt := range keyTypeList(list) {
			if kt.Name["fa"] == "" || kt.Name["en"] == "" {
				t.Errorf("key type %q has no fa/en name", kt.Key)
			}
		}
	}
	if !validKeyType(tagKeyTypes, "user") || validKeyType(categoryKeyTypes, "user") {
		t.Fatal("user is a tag key type only")
	}
	if validKeyType(clientTagKeyTypes, "user") {
		t.Fatal("players must not list player tags")
	}
	for _, kt := range clientTagKeyTypes {
		if !validKeyType(tagKeyTypes, kt) {
			t.Errorf("client tag key type %q is not a tag key type", kt)
		}
	}
	if tagKeyTypes[len(tagKeyTypes)-1] != "user" || tagKeyTypes[len(tagKeyTypes)-2] != "general" {
		t.Fatalf("tag key types order = %v", tagKeyTypes)
	}
}
