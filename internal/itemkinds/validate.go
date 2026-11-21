package itemkinds

import (
	"fmt"
	"strings"
)

type FieldError struct {
	Path    string
	Message string
}

func (e *FieldError) Error() string { return e.Path + ": " + e.Message }

func ValidateData(fields []Field, data map[string]any) error {
	return validateFields("", fields, data)
}

func validateFields(prefix string, fields []Field, data map[string]any) error {
	for _, f := range fields {
		v, ok := data[f.Key]
		if !ok || v == nil {
			continue
		}
		if err := validateValue(prefix+f.Key, f, v); err != nil {
			return err
		}
	}
	return nil
}

func validateValue(path string, f Field, v any) error {
	bad := func(msg string) error { return &FieldError{Path: path, Message: msg} }
	switch f.Type {
	case "text", "":
		if _, ok := v.(string); !ok {
			return bad("must be text")
		}
	case "number", "int":
		n, ok := Num(v)
		if !ok {
			return bad("must be a number")
		}
		if f.Type == "int" && n != float64(int64(n)) {
			return bad("must be a whole number")
		}
		if f.Min != nil && n < *f.Min {
			return bad(fmt.Sprintf("must be at least %v", *f.Min))
		}
		if f.Max != nil && n > *f.Max {
			return bad(fmt.Sprintf("must be at most %v", *f.Max))
		}
	case "bool":
		if _, ok := v.(bool); !ok {
			return bad("must be true or false")
		}
	case "select":
		s, ok := v.(string)
		if !ok {
			return bad("must be text")
		}
		if len(f.Options) > 0 && !containsFold(f.Options, s) && s != "" {
			return bad("must be one of " + strings.Join(f.Options, ", "))
		}
	case "tags":
		list, ok := v.([]any)
		if !ok {
			return bad("must be a list")
		}
		for i, x := range list {
			if _, ok := x.(string); !ok {
				return &FieldError{Path: fmt.Sprintf("%s[%d]", path, i), Message: "must be text"}
			}
		}
		if f.MaxItems > 0 && len(list) > f.MaxItems {
			return bad(fmt.Sprintf("at most %d items", f.MaxItems))
		}
	case "i18n":
		m, ok := v.(map[string]any)
		if !ok {
			return bad("must be a text per language")
		}
		for k, x := range m {
			if _, ok := x.(string); !ok {
				return &FieldError{Path: path + "." + k, Message: "must be text"}
			}
		}
	case "group":
		m, ok := v.(map[string]any)
		if !ok {
			return bad("must be an object")
		}
		return validateFields(path+".", f.Fields, m)
	case "list":
		list, ok := v.([]any)
		if !ok {
			return bad("must be a list")
		}
		if f.MaxItems > 0 && len(list) > f.MaxItems {
			return bad(fmt.Sprintf("at most %d items", f.MaxItems))
		}
		for i, x := range list {
			m, ok := x.(map[string]any)
			if !ok {
				return &FieldError{Path: fmt.Sprintf("%s[%d]", path, i), Message: "must be an object"}
			}
			if err := validateFields(fmt.Sprintf("%s[%d].", path, i), f.Fields, m); err != nil {
				return err
			}
		}
	case "json":
	default:
		return bad("unknown field type " + f.Type)
	}
	return nil
}
