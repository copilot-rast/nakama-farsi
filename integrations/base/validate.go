package main

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var keyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)

type keyType struct {
	Key  string            `json:"key"`
	Name map[string]string `json:"name"`
}

var keyTypeNames = map[string]map[string]string{
	"product":  {"fa": "محصول", "en": "Product"},
	"prize":    {"fa": "جایزه", "en": "Prize"},
	"gift":     {"fa": "هدیه", "en": "Gift"},
	"currency": {"fa": "ارز", "en": "Currency"},
	"faq":      {"fa": "سؤال متداول", "en": "FAQ"},
	"news":     {"fa": "خبر", "en": "News"},
	"general":  {"fa": "عمومی", "en": "General"},
}

var categoryKeyTypes = []string{"product", "prize", "gift", "currency", "faq", "news"}

var tagKeyTypes = []string{"product", "prize", "news", "faq", "general"}

func keyTypeList(keys []string) []keyType {
	out := make([]keyType, 0, len(keys))
	for _, k := range keys {
		out = append(out, keyType{Key: k, Name: keyTypeNames[k]})
	}
	return out
}

func validKeyType(keys []string, v string) bool {
	for _, k := range keys {
		if k == v {
			return true
		}
	}
	return false
}

func textError(err error) error {
	var e *localized.Error
	if errors.As(err, &e) {
		return badRequest(e.Code, e.Field, e.Message)
	}
	return err
}

func cleanText(langs *localized.Set, field string, in localized.Text, required bool, max int) (localized.Text, error) {
	out, err := langs.Clean(field, in, required, max)
	return out, textError(err)
}

func cleanKey(field, key string) (string, error) {
	key = strings.TrimSpace(key)
	if !keyPattern.MatchString(key) {
		return "", badRequest("INVALID_KEY", field, "use lower-case letters, digits, _ . - (max 64)")
	}
	return key, nil
}

func cleanURL(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) > 1024 {
		return "", badRequest("URL_TOO_LONG", field, "URL is too long")
	}
	lower := strings.ToLower(value)
	if value != "" && !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") && (!strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//")) {
		return "", badRequest("INVALID_URL", field, "URL must be absolute http(s) or start with /")
	}
	if strings.ContainsAny(value, " \t\r\n\"'<>`") {
		return "", badRequest("INVALID_URL", field, "URL contains invalid characters")
	}
	return value, nil
}

func cleanObject(field string, raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "{}", nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", badRequest("INVALID_OBJECT", field, "must be a JSON object")
	}
	if len(raw) > 16384 {
		return "", badRequest("OBJECT_TOO_LARGE", field, "object is too large")
	}
	out, _ := json.Marshal(obj)
	return string(out), nil
}

func decodeInto(raw []byte, out any) error {
	if err := json.Unmarshal(raw, out); err != nil {
		return badRequest("INVALID_JSON", "", "invalid body: "+err.Error())
	}
	return nil
}

func (s *service) languages(ctx context.Context, q localized.Queryer) (*localized.Set, error) {
	return localized.Load(ctx, q)
}

func rawJSON(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(s)
}
