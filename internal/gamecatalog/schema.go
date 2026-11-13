package gamecatalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"unicode/utf8"
)

type Property struct {
	Type        string          `json:"type"`
	Title       json.RawMessage `json:"title,omitempty"`
	Description json.RawMessage `json:"description,omitempty"`
	Default     json.RawMessage `json:"default,omitempty"`
	Minimum     *float64        `json:"minimum,omitempty"`
	Maximum     *float64        `json:"maximum,omitempty"`
	MinLength   *int            `json:"minLength,omitempty"`
	MaxLength   *int            `json:"maxLength,omitempty"`
	Enum        []any           `json:"enum,omitempty"`
}

type Schema struct {
	Type                 string              `json:"type"`
	Properties           map[string]Property `json:"properties"`
	Required             []string            `json:"required,omitempty"`
	AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
}

type FieldError struct {
	Field   string
	Code    string
	Message string
}

func (e *FieldError) Error() string {
	return e.Field + ": " + e.Message
}

var allowedTypes = map[string]bool{"integer": true, "number": true, "string": true, "boolean": true}

func ParseSchema(raw json.RawMessage) (*Schema, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return &Schema{Type: "object", Properties: map[string]Property{}}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var s Schema
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("schema is not a supported JSON schema: %w", err)
	}
	if s.Type != "object" {
		return nil, errors.New(`schema type must be "object"`)
	}
	if s.Properties == nil {
		s.Properties = map[string]Property{}
	}
	if len(s.Properties) > 50 {
		return nil, errors.New("schema has more than 50 properties")
	}
	for name, p := range s.Properties {
		if !keyPattern.MatchString(name) && !camelPattern.MatchString(name) {
			return nil, fmt.Errorf("property %q needs a Latin name", name)
		}
		if !allowedTypes[p.Type] {
			return nil, fmt.Errorf("property %s: type must be integer, number, string or boolean", name)
		}
		if p.Minimum != nil && p.Maximum != nil && *p.Minimum > *p.Maximum {
			return nil, fmt.Errorf("property %s: minimum is above maximum", name)
		}
		if len(p.Default) > 0 {
			var v any
			if err := json.Unmarshal(p.Default, &v); err != nil {
				return nil, fmt.Errorf("property %s: default is not JSON", name)
			}
			if fe := p.check(name, v); fe != nil {
				return nil, fmt.Errorf("property %s: default does not fit: %s", name, fe.Message)
			}
		}
	}
	for _, name := range s.Required {
		if _, ok := s.Properties[name]; !ok {
			return nil, fmt.Errorf("required property %s is not in properties", name)
		}
	}
	return &s, nil
}

func (s *Schema) closed() bool {
	return s.AdditionalProperties != nil && !*s.AdditionalProperties
}

func (s *Schema) Apply(values map[string]any) (map[string]any, *FieldError) {
	out := map[string]any{}
	required := map[string]bool{}
	for _, r := range s.Required {
		required[r] = true
	}
	names := make([]string, 0, len(values))
	for k := range values {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		p, ok := s.Properties[k]
		if !ok {
			if s.closed() {
				return nil, &FieldError{Field: "settings." + k, Code: "UNKNOWN_SETTING", Message: "this game has no setting named " + k}
			}
			out[k] = values[k]
			continue
		}
		v := values[k]
		if v == nil {
			continue
		}
		if fe := p.check(k, v); fe != nil {
			return nil, fe
		}
		if p.Type == "integer" {
			v = math.Trunc(v.(float64))
		}
		out[k] = v
	}
	props := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		props = append(props, k)
	}
	sort.Strings(props)
	for _, k := range props {
		if _, ok := out[k]; ok {
			continue
		}
		p := s.Properties[k]
		if len(p.Default) > 0 {
			var v any
			_ = json.Unmarshal(p.Default, &v)
			out[k] = v
			continue
		}
		if required[k] {
			return nil, &FieldError{Field: "settings." + k, Code: "SETTING_REQUIRED", Message: k + " is required"}
		}
	}
	return out, nil
}

func (s *Schema) Defaults() map[string]any {
	out, _ := s.Apply(map[string]any{})
	if out == nil {
		out = map[string]any{}
	}
	return out
}

func (p Property) check(name string, v any) *FieldError {
	field := "settings." + name
	bad := func(msg string) *FieldError {
		return &FieldError{Field: field, Code: "INVALID_SETTING", Message: msg}
	}
	switch p.Type {
	case "boolean":
		if _, ok := v.(bool); !ok {
			return bad("must be true or false")
		}
	case "string":
		str, ok := v.(string)
		if !ok {
			return bad("must be text")
		}
		n := utf8.RuneCountInString(str)
		if p.MinLength != nil && n < *p.MinLength {
			return bad(fmt.Sprintf("must be at least %d characters", *p.MinLength))
		}
		if p.MaxLength != nil && n > *p.MaxLength {
			return bad(fmt.Sprintf("must be at most %d characters", *p.MaxLength))
		}
		if n > 2000 {
			return bad("must be at most 2000 characters")
		}
	case "integer", "number":
		f, ok := v.(float64)
		if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
			return bad("must be a number")
		}
		if p.Type == "integer" && f != math.Trunc(f) {
			return bad("must be a whole number")
		}
		if p.Minimum != nil && f < *p.Minimum {
			return bad(fmt.Sprintf("must be %v or more", *p.Minimum))
		}
		if p.Maximum != nil && f > *p.Maximum {
			return bad(fmt.Sprintf("must be %v or less", *p.Maximum))
		}
	}
	if len(p.Enum) > 0 {
		for _, e := range p.Enum {
			if fmt.Sprint(e) == fmt.Sprint(v) {
				return nil
			}
		}
		return &FieldError{Field: field, Code: "INVALID_SETTING", Message: "must be one of the listed values"}
	}
	return nil
}
