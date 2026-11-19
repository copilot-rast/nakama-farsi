package voice

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type Request struct {
	Room       string
	Identity   string
	Name       string
	GameKey    string
	Players    []Player
	TTL        time.Duration
	CanPublish bool
}

type Grant struct {
	Provider  string    `json:"provider"`
	URL       string    `json:"url"`
	Token     string    `json:"token"`
	Room      string    `json:"room"`
	Identity  string    `json:"identity"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Provider interface {
	Token(ctx context.Context, req Request) (Grant, error)
	Test(ctx context.Context) error
}

type Field struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Secret   bool   `json:"secret"`
	Required bool   `json:"required"`
}

type Config struct {
	Values  map[string]string
	Secrets map[string]string
}

func (c Config) Get(name string) string {
	if v, ok := c.Secrets[name]; ok {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(c.Values[name])
}

type Spec struct {
	Kind   string
	Fields []Field
	Check  func(cfg Config) error
	New    func(cfg Config, client *http.Client) (Provider, error)
}

var (
	mu       sync.RWMutex
	registry = map[string]Spec{}
)

func Register(spec Spec) {
	mu.Lock()
	defer mu.Unlock()
	registry[spec.Kind] = spec
}

func Lookup(kind string) (Spec, bool) {
	mu.RLock()
	defer mu.RUnlock()
	s, ok := registry[kind]
	return s, ok
}

func Kinds() []Spec {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Spec, 0, len(registry))
	for _, s := range registry {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}

type FieldError struct {
	Code    string
	Field   string
	Message string
}

func (e *FieldError) Error() string { return e.Code + ": " + e.Field + ": " + e.Message }

func Validate(kind string, cfg Config) error {
	spec, ok := Lookup(kind)
	if !ok {
		return &FieldError{Code: "INVALID_VOICE_PROVIDER", Field: "provider", Message: "unknown voice provider"}
	}
	for _, f := range spec.Fields {
		v := cfg.Get(f.Name)
		if v == "" {
			if f.Required {
				return &FieldError{Code: "VOICE_FIELD_REQUIRED", Field: f.Name, Message: "required"}
			}
			continue
		}
		if len(v) > 4096 {
			return &FieldError{Code: "VOICE_FIELD_TOO_LONG", Field: f.Name, Message: "at most 4096 characters"}
		}
		switch f.Kind {
		case "url":
			if !validURL(v, "http", "https") {
				return &FieldError{Code: "INVALID_URL", Field: f.Name, Message: "an http(s) address"}
			}
		case "wsurl":
			if !validURL(v, "ws", "wss", "http", "https") {
				return &FieldError{Code: "INVALID_URL", Field: f.Name, Message: "a ws(s) or http(s) address"}
			}
		}
	}
	if spec.Check != nil {
		return spec.Check(cfg)
	}
	return nil
}

func validURL(v string, schemes ...string) bool {
	u, err := url.Parse(v)
	if err != nil || u.Host == "" || u.User != nil {
		return false
	}
	for _, s := range schemes {
		if u.Scheme == s {
			return true
		}
	}
	return false
}

var roomPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

func ValidRoom(room string) bool { return roomPattern.MatchString(room) }

type ProviderError struct {
	Code   string
	Status int
	Detail string
}

func (e *ProviderError) Error() string {
	if e.Detail != "" {
		return e.Code + ": " + e.Detail
	}
	return e.Code
}

var ErrUnavailable = errors.New("VOICE_PROVIDER_UNAVAILABLE")
