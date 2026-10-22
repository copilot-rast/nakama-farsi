package main

import "testing"

func TestMedalConditionSelectionAndEventMapping(t *testing.T) {
	schema := []byte(`{"activity":"group_members","count":{"min":2,"max":20}}`)
	params := []byte(`{"count":5,"activity":"purchase","ignored":true}`)
	clean, issues, err := validateMedalParams(schema, params)
	if err != nil || len(issues) != 0 || len(clean) != 1 || clean["count"] != 5 {
		t.Fatalf("unexpected clean params: %#v %#v %v", clean, issues, err)
	}
	condition, err := resolveMedalCondition("activity_count", schema, params)
	if err != nil || condition.Activity != "group_members" || condition.Threshold != 5 || !medalListensTo(condition, "group_joined") || medalListensTo(condition, "purchase") {
		t.Fatalf("wrong group medal resolution: %+v %v", condition, err)
	}
	_, issues, err = validateMedalParams(schema, []byte(`{"count":21}`))
	if err != nil || len(issues) != 1 || issues[0].Field != "count" {
		t.Fatalf("out-of-range count accepted: %#v %v", issues, err)
	}
	clean, issues, err = validateMedalParams(schema, []byte(`{"count":"5"}`))
	if err != nil || len(issues) != 0 || clean["count"] != 5 {
		t.Fatalf("source numeric-string coercion failed: %#v %#v %v", clean, issues, err)
	}
	eventSchema := []byte(`{"event":"string"}`)
	_, issues, err = validateMedalParams(eventSchema, []byte(`{"event":"  festival  "}`))
	if err != nil || len(issues) != 0 {
		t.Fatalf("valid event rejected: %#v %v", issues, err)
	}
}
