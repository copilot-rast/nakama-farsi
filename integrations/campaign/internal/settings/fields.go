package settings

import (
	"math"
	"strings"
	"unicode/utf8"
)

func Applies(d FieldDef, chapterKey string) bool {
	if len(d.Chapters) == 0 {
		return true
	}
	for _, c := range d.Chapters {
		if c == chapterKey {
			return true
		}
	}
	return false
}

func number(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, !math.IsNaN(x) && !math.IsInf(x, 0)
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	}
	return 0, false
}

func CleanValue(d FieldDef, v any, depth int) (any, string) {
	switch d.Type {
	case FieldNumber, FieldInteger:
		n, ok := number(v)
		if !ok {
			return nil, "INVALID_FIELD_VALUE"
		}
		if d.Type == FieldInteger && n != math.Trunc(n) {
			return nil, "INVALID_FIELD_VALUE"
		}
		if d.Min != nil && n < *d.Min || d.Max != nil && n > *d.Max || math.Abs(n) > 1e12 {
			return nil, "VALUE_OUT_OF_RANGE"
		}
		return n, ""
	case FieldBool:
		b, ok := v.(bool)
		if !ok {
			return nil, "INVALID_FIELD_VALUE"
		}
		return b, ""
	case FieldText:
		s, ok := v.(string)
		if !ok {
			return nil, "INVALID_FIELD_VALUE"
		}
		s = strings.TrimSpace(s)
		if utf8.RuneCountInString(s) > 4000 {
			return nil, "TEXT_TOO_LONG"
		}
		return s, ""
	case FieldSelect:
		s, ok := v.(string)
		if !ok {
			return nil, "INVALID_FIELD_VALUE"
		}
		for _, o := range d.Options {
			if o == s {
				return s, ""
			}
		}
		return nil, "INVALID_OPTION"
	case FieldI18n:
		m, ok := v.(map[string]any)
		if !ok {
			if sm, ok2 := v.(map[string]string); ok2 {
				m = map[string]any{}
				for k, x := range sm {
					m[k] = x
				}
			} else {
				return nil, "INVALID_FIELD_VALUE"
			}
		}
		out := map[string]any{}
		for k, x := range m {
			s, ok := x.(string)
			if !ok || len(k) > 16 || k == "" {
				return nil, "INVALID_FIELD_VALUE"
			}
			s = strings.TrimSpace(s)
			if utf8.RuneCountInString(s) > 4000 {
				return nil, "TEXT_TOO_LONG"
			}
			if s != "" {
				out[k] = s
			}
		}
		return out, ""
	case FieldTags:
		list, ok := v.([]any)
		if !ok {
			if sl, ok2 := v.([]string); ok2 {
				for _, x := range sl {
					list = append(list, x)
				}
			} else {
				return nil, "INVALID_FIELD_VALUE"
			}
		}
		if len(list) > 200 {
			return nil, "TOO_MANY_ITEMS"
		}
		out := []any{}
		for _, x := range list {
			s, ok := x.(string)
			if !ok {
				return nil, "INVALID_FIELD_VALUE"
			}
			s = strings.TrimSpace(s)
			if utf8.RuneCountInString(s) > 200 {
				return nil, "TEXT_TOO_LONG"
			}
			if s != "" {
				out = append(out, s)
			}
		}
		return out, ""
	case FieldList:
		list, ok := v.([]any)
		if !ok {
			if ml, ok2 := v.([]map[string]any); ok2 {
				for _, x := range ml {
					list = append(list, x)
				}
			} else {
				return nil, "INVALID_FIELD_VALUE"
			}
		}
		max := d.MaxItems
		if max == 0 {
			max = 20
		}
		if len(list) > max {
			return nil, "TOO_MANY_ITEMS"
		}
		if depth >= MaxFieldDepth {
			return nil, "FIELDS_TOO_DEEP"
		}
		out := []any{}
		for _, x := range list {
			m, ok := x.(map[string]any)
			if !ok {
				return nil, "INVALID_FIELD_VALUE"
			}
			clean, _, code := cleanValues(d.Fields, "", m, depth+1, false)
			if code != "" {
				return nil, code
			}
			out = append(out, clean)
		}
		return out, ""
	}
	return nil, "INVALID_FIELD_TYPE"
}

func CleanValues(defs []FieldDef, chapterKey string, in map[string]any) (map[string]any, string, string) {
	return cleanValues(defs, chapterKey, in, 1, true)
}

func cleanValues(defs []FieldDef, chapterKey string, in map[string]any, depth int, top bool) (map[string]any, string, string) {
	out := map[string]any{}
	known := map[string]FieldDef{}
	for _, d := range defs {
		if !top || Applies(d, chapterKey) {
			known[d.Key] = d
		}
	}
	for k, v := range in {
		d, ok := known[k]
		if !ok {
			return nil, k, "UNKNOWN_FIELD"
		}
		if v == nil {
			continue
		}
		clean, code := CleanValue(d, v, depth)
		if code != "" {
			return nil, k, code
		}
		out[k] = clean
	}
	return out, "", ""
}

func Resolve(defs []FieldDef, chapterKey string, stored map[string]any, text func(map[string]string) string) map[string]any {
	return resolve(defs, chapterKey, stored, text, true)
}

func resolve(defs []FieldDef, chapterKey string, stored map[string]any, text func(map[string]string) string, top bool) map[string]any {
	out := map[string]any{}
	for _, d := range defs {
		if top && !Applies(d, chapterKey) {
			continue
		}
		v, ok := stored[d.Key]
		if !ok || v == nil {
			v = d.Default
		}
		if v == nil {
			if d.Type == FieldList || d.Type == FieldTags {
				out[d.Key] = []any{}
			}
			continue
		}
		switch d.Type {
		case FieldI18n:
			m := map[string]string{}
			if raw, ok := v.(map[string]any); ok {
				for k, x := range raw {
					if s, ok := x.(string); ok {
						m[k] = s
					}
				}
			}
			out[d.Key] = text(m)
		case FieldList:
			items := []any{}
			if list, ok := v.([]any); ok {
				for _, x := range list {
					if m, ok := x.(map[string]any); ok {
						items = append(items, resolve(d.Fields, chapterKey, m, text, false))
					}
				}
			}
			out[d.Key] = items
		default:
			out[d.Key] = v
		}
	}
	return out
}
