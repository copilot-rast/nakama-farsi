package main

import (
	"encoding/json"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var (
	keyPattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)
	currencyPattern = regexp.MustCompile(`^[a-z0-9_]{2,32}$`)
	giftCodePattern = regexp.MustCompile(`^[A-Z0-9_-]{3,32}$`)
	skuPattern      = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,128}$`)
	numericPattern  = regexp.MustCompile(`^-?[0-9]{1,16}(\.[0-9]{1,4})?$`)
)

var allMethods = []string{"zarinpal", "zibal", "bazaar", "myket", "in_app"}

func isGateway(method string) bool { return method == "zarinpal" || method == "zibal" }

func isStore(method string) bool { return method == "bazaar" || method == "myket" }

func validMethod(method string) bool {
	for _, m := range allMethods {
		if m == method {
			return true
		}
	}
	return false
}

type i18n = localized.Text

func cleanI18n(langs *localized.Set, field string, in i18n, required bool, max int) (i18n, error) {
	out, err := langs.Clean(field, in, required, max)
	return out, textError(err)
}

func textError(err error) error {
	var e *localized.Error
	if errors.As(err, &e) {
		return badRequest(e.Code, e.Field, e.Message)
	}
	return err
}

func translatedSearch(col string) string {
	return "(SELECT string_agg(e.value, ' ') FROM jsonb_each_text(" + col + ") e)"
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
	if value != "" && !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") && !strings.HasPrefix(value, "/") {
		return "", badRequest("INVALID_URL", field, "URL must be absolute http(s) or start with /")
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

func validNumeric(value string) bool {
	if !numericPattern.MatchString(value) {
		return false
	}
	_, ok := new(big.Rat).SetString(value)
	return ok
}

func numericFromJSON(raw json.RawMessage) (string, bool) {
	var n json.Number
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return "", false
	}
	switch t := v.(type) {
	case json.Number:
		n = t
	case string:
		n = json.Number(strings.TrimSpace(t))
	default:
		return "", false
	}
	r, ok := new(big.Rat).SetString(string(n))
	if !ok {
		return "", false
	}
	text := r.FloatString(4)
	text = strings.TrimRight(strings.TrimRight(text, "0"), ".")
	if text == "" || text == "-" {
		text = "0"
	}
	return text, validNumeric(text)
}

func int64FromJSON(raw json.RawMessage) (int64, bool) {
	var n json.Number
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&n); err != nil {
		return 0, false
	}
	v, err := n.Int64()
	return v, err == nil
}

func cleanTime(field, value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, badRequest("INVALID_TIME", field, "use an RFC 3339 timestamp")
	}
	t = t.UTC()
	return &t, nil
}

type prizeItem struct {
	Type   string          `json:"type"`
	Key    string          `json:"key"`
	Action string          `json:"action"`
	Value  json.RawMessage `json:"value"`
}

func normalizePrizeItems(items []prizeItem) ([]prizeItem, error) {
	if len(items) == 0 {
		return nil, badRequest("ITEMS_REQUIRED", "items", "at least one prize item is required")
	}
	if len(items) > 50 {
		return nil, badRequest("TOO_MANY_ITEMS", "items", "at most 50 items")
	}
	out := make([]prizeItem, 0, len(items))
	for _, item := range items {
		item.Key = strings.TrimSpace(item.Key)
		switch item.Type {
		case "currency":
			if item.Action != "add" {
				return nil, badRequest("INVALID_ACTION", "items", "currency items only support add")
			}
			if !currencyPattern.MatchString(item.Key) {
				return nil, badRequest("INVALID_KEY", "items", "invalid currency key")
			}
			v, ok := int64FromJSON(item.Value)
			if !ok || v == 0 || v > 1_000_000_000_000 || v < -1_000_000_000_000 {
				return nil, badRequest("INVALID_VALUE", "items", "currency value must be a non-zero integer")
			}
			item.Value = json.RawMessage(jsonInt(v))
		case "meta":
			if item.Action != "add" && item.Action != "set" {
				return nil, badRequest("INVALID_ACTION", "items", "meta items support add or set")
			}
			if !keyPattern.MatchString(item.Key) {
				return nil, badRequest("INVALID_KEY", "items", "invalid meta key")
			}
			v, ok := numericFromJSON(item.Value)
			if !ok {
				return nil, badRequest("INVALID_VALUE", "items", "meta value must be a number with at most 4 decimals")
			}
			item.Value = json.RawMessage(v)
		case "storage":
			if item.Action != "give" {
				return nil, badRequest("INVALID_ACTION", "items", "storage items only support give")
			}
			if !keyPattern.MatchString(item.Key) {
				return nil, badRequest("INVALID_KEY", "items", "invalid storage key")
			}
			var obj map[string]any
			if err := json.Unmarshal(item.Value, &obj); err != nil || obj == nil {
				return nil, badRequest("INVALID_VALUE", "items", "storage value must be a JSON object")
			}
			b, _ := json.Marshal(obj)
			item.Value = b
		default:
			return nil, badRequest("INVALID_ITEM_TYPE", "items", "type must be currency, meta or storage")
		}
		out = append(out, item)
	}
	return out, nil
}

func jsonInt(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func discounted(price int64, percent int) int64 {
	if percent <= 0 {
		return price
	}
	if percent >= 100 {
		return 0
	}
	return price * int64(100-percent) / 100
}
