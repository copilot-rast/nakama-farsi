package main

import (
	"encoding/json"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
)

type i18n map[string]string

func cleanI18n(field string, in i18n, required bool, max int) (i18n, error) {
	out := i18n{}
	for k, v := range in {
		if k != "fa" && k != "en" {
			return nil, badRequest("INVALID_LANGUAGE", field, "only fa and en are allowed")
		}
		v = strings.TrimSpace(v)
		if !utf8.ValidString(v) || utf8.RuneCountInString(v) > max {
			return nil, badRequest("TEXT_TOO_LONG", field, "text is too long (max "+strconv.Itoa(max)+")")
		}
		if v != "" {
			out[k] = v
		}
	}
	if (required || len(out) > 0) && out["fa"] == "" {
		return nil, badRequest("FA_REQUIRED", field, "a Persian (fa) value is required")
	}
	return out, nil
}

func (t i18n) pick(lang string) string {
	if v := t[lang]; v != "" {
		return v
	}
	return t["fa"]
}

func cleanLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if strings.HasPrefix(lang, "en") {
		return "en"
	}
	return "fa"
}

func cleanURL(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) > 1024 {
		return "", badRequest("URL_TOO_LONG", field, "URL is too long")
	}
	if value != "" && !safeURL(value) {
		return "", badRequest("INVALID_URL", field, "URL must be absolute http(s) or start with /")
	}
	return value, nil
}

func safeURL(value string) bool {
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return !strings.ContainsAny(value, " \t\r\n\"'<>`")
	}
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !strings.HasPrefix(value, `/\`) {
		return !strings.ContainsAny(value, " \t\r\n\"'<>`")
	}
	return false
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

func cleanUUID(field, value string, required bool) (*string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return nil, badRequest("REQUIRED", field, "this field is required")
		}
		return nil, nil
	}
	id, err := uuid.FromString(value)
	if err != nil {
		return nil, badRequest("INVALID_ID", field, "must be a UUID")
	}
	s := id.String()
	return &s, nil
}

func cleanTime(field, value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, badRequest("INVALID_TIME", field, "use an RFC 3339 time")
	}
	return &t, nil
}

type action struct {
	Type  string `json:"type"`
	Label i18n   `json:"label,omitempty"`
	Value string `json:"value"`
}

func cleanAction(field string, in action, withLabel bool) (action, error) {
	in.Type = strings.TrimSpace(in.Type)
	in.Value = strings.TrimSpace(in.Value)
	switch in.Type {
	case "url":
		if !strings.HasPrefix(strings.ToLower(in.Value), "http://") && !strings.HasPrefix(strings.ToLower(in.Value), "https://") || !safeURL(in.Value) || len(in.Value) > 1024 {
			return in, badRequest("INVALID_ACTION", field, "url actions need an http(s) URL")
		}
	case "goto":
		if in.Value == "" || utf8.RuneCountInString(in.Value) > 256 || strings.ContainsAny(in.Value, "<>\"'`") {
			return in, badRequest("INVALID_ACTION", field, "goto actions need a route name (max 256)")
		}
	default:
		return in, badRequest("INVALID_ACTION", field, "type must be url or goto")
	}
	if withLabel {
		label, err := cleanI18n(field+".label", in.Label, true, 64)
		if err != nil {
			return in, err
		}
		in.Label = label
	} else {
		in.Label = nil
	}
	return in, nil
}

func latinDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '۰' && r <= '۹':
			b.WriteRune('0' + (r - '۰'))
		case r >= '٠' && r <= '٩':
			b.WriteRune('0' + (r - '٠'))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizePhone(raw string) (string, bool) {
	s := latinDigits(strings.TrimSpace(raw))
	s = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "").Replace(s)
	plus := strings.HasPrefix(s, "+")
	s = strings.TrimPrefix(s, "+")
	if strings.HasPrefix(s, "00") {
		s = s[2:]
		plus = true
	}
	if s == "" {
		return "", false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	switch {
	case len(s) == 11 && strings.HasPrefix(s, "09") && !plus:
		return "+98" + s[1:], true
	case len(s) == 10 && strings.HasPrefix(s, "9") && !plus:
		return "+98" + s, true
	case len(s) == 12 && strings.HasPrefix(s, "989"):
		return "+" + s, true
	case strings.HasPrefix(s, "98"):
		return "", false
	case plus && len(s) >= 8 && len(s) <= 15 && s[0] != '0':
		return "+" + s, true
	}
	return "", false
}

func iranLocal(e164 string) (string, bool) {
	if strings.HasPrefix(e164, "+989") && len(e164) == 13 {
		return "0" + e164[3:], true
	}
	return "", false
}

func normalizeEmail(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 254 || strings.ContainsAny(raw, "\r\n<> ,;") {
		return "", false
	}
	addr, err := mail.ParseAddress(raw)
	if err != nil || addr.Address != raw || !strings.Contains(raw[strings.LastIndex(raw, "@")+1:], ".") {
		return "", false
	}
	return raw, true
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func i18nJSON(t i18n) string {
	if t == nil {
		t = i18n{}
	}
	b, _ := json.Marshal(t)
	return string(b)
}

func decodeInto(raw []byte, out any) error {
	if err := json.Unmarshal(raw, out); err != nil {
		return badRequest("INVALID_JSON", "", "invalid body: "+err.Error())
	}
	return nil
}
