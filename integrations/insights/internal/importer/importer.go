package importer

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/events"
)

type Item struct {
	Key        string
	Raw        events.Raw
	UserID     string
	SessionID  string
	At         time.Time
	Platform   string
	AppVersion string
}

type Env struct {
	HTTP   *http.Client
	Now    time.Time
	Config json.RawMessage
	Secret string
	Cursor json.RawMessage
	Max    int
	Emit   func(ctx context.Context, items []Item) error
	Renew  func(ctx context.Context) error
}

type Outcome struct {
	Cursor  json.RawMessage
	Read    int64
	Skipped int64
	More    bool
}

type Importer interface {
	Kind() string
	Check(config json.RawMessage, secret string) (json.RawMessage, error)
	Pull(ctx context.Context, env Env) (Outcome, error)
}

type ConfigError struct {
	Field string
	Code  string
}

func (e *ConfigError) Error() string { return e.Code + " " + e.Field }

func bad(field, code string) error { return &ConfigError{Field: field, Code: code} }

var All = map[string]Importer{}

func register(i Importer) { All[i.Kind()] = i }

var (
	paramKey  = regexp.MustCompile(`[^a-z0-9_]+`)
	nameClean = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)
)

func GroupKey(s string) string {
	g := strings.Trim(paramKey.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "_"), "_")
	if g == "" {
		return ""
	}
	if g[0] < 'a' || g[0] > 'z' {
		g = "p_" + g
	}
	if len(g) > 32 {
		g = g[:32]
	}
	return g
}

func EventName(s string) string {
	n := nameClean.ReplaceAllString(strings.TrimSpace(s), "_")
	if n == "" {
		return ""
	}
	if !(n[0] >= 'a' && n[0] <= 'z' || n[0] >= 'A' && n[0] <= 'Z') {
		n = "e_" + n
	}
	if len(n) > 64 {
		n = n[:64]
	}
	return n
}

func TagID(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), ":", "_")
	if r := []rune(s); len(r) > 64 {
		s = string(r[:64])
	}
	return s
}

func CheckEndpoint(raw, def string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return def, nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", errors.New("invalid url")
	}
	if u.Scheme == "https" {
		return raw, nil
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); u.Scheme == "http" && (host == "localhost" || ip != nil && ip.IsLoopback()) {
		return raw, nil
	}
	return "", errors.New("only https or a loopback http url")
}

func decodeCursor[T any](raw json.RawMessage) T {
	var out T
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out
}

func encode(v any) json.RawMessage {
	raw, _ := json.Marshal(v)
	return raw
}
