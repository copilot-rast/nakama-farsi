package main

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

type medalRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

func medalRangeRule(raw json.RawMessage) (medalRange, bool) {
	var r medalRange
	if len(raw) == 0 || json.Unmarshal(raw, &r) != nil || r.Min < 1 || r.Max < r.Min {
		return r, false
	}
	return r, true
}

type medalParamProblem struct {
	Field string `json:"field"`
	Min   *int   `json:"min,omitempty"`
	Max   *int   `json:"max,omitempty"`
	Kind  string `json:"kind,omitempty"`
	Value any    `json:"value"`
}

// Only brand-pickable range and string fields survive. Fixed schema fields
// remain in the library and are resolved at evaluation time.
func validateMedalParams(schemaRaw, paramsRaw []byte) (map[string]any, []medalParamProblem, error) {
	var schema map[string]json.RawMessage
	var params map[string]any
	if err := json.Unmarshal(schemaRaw, &schema); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return nil, nil, err
	}
	clean := map[string]any{}
	issues := []medalParamProblem{}
	for key, rule := range schema {
		if bounds, ok := medalRangeRule(rule); ok {
			value := params[key]
			f, ok := medalNumber(value)
			if !ok || math.Trunc(f) != f || f < float64(bounds.Min) || f > float64(bounds.Max) {
				min, max := bounds.Min, bounds.Max
				issues = append(issues, medalParamProblem{Field: key, Min: &min, Max: &max, Value: value})
			} else {
				clean[key] = int(f)
			}
			continue
		}
		var fixed string
		if json.Unmarshal(rule, &fixed) == nil && fixed == "string" {
			value := params[key]
			v, ok := value.(string)
			v = strings.TrimSpace(v)
			if !ok || v == "" || len([]rune(v)) > 64 {
				issues = append(issues, medalParamProblem{Field: key, Kind: "string", Value: value})
			} else {
				clean[key] = v
			}
		}
	}
	return clean, issues, nil
}

func medalNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return n, err == nil
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

type resolvedMedalCondition struct {
	Type      string
	Activity  string
	Threshold int
	EventKey  string
}

func resolveMedalCondition(kind string, schemaRaw, paramsRaw []byte) (resolvedMedalCondition, error) {
	result := resolvedMedalCondition{Type: kind}
	var schema map[string]json.RawMessage
	var params map[string]any
	if err := json.Unmarshal(schemaRaw, &schema); err != nil {
		return result, err
	}
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return result, err
	}
	pick := func(key string) int {
		if v, ok := medalNumber(params[key]); ok && math.Trunc(v) == v {
			return int(v)
		}
		if r, ok := medalRangeRule(schema[key]); ok {
			return r.Min
		}
		return 0
	}
	switch kind {
	case "activity_count":
		_ = json.Unmarshal(schema["activity"], &result.Activity)
		result.Threshold = pick("count")
	case "duration":
		result.Threshold = pick("days")
	case "special_event":
		result.EventKey, _ = params["event"].(string)
	}
	return result, nil
}

func medalListensTo(c resolvedMedalCondition, event string) bool {
	switch c.Type {
	case "activity_count":
		if c.Activity == "group_members" {
			return event == "group_joined"
		}
		return c.Activity == event
	case "duration":
		return event != "special_event"
	case "special_event":
		return event == "special_event"
	}
	return false
}
