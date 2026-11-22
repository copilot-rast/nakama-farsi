package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type Errors struct {
	Mode                  string `json:"mode"`
	ModuleMode            string `json:"moduleMode"`
	FallbackLang          string `json:"fallbackLang"`
	PreferAccountLanguage bool   `json:"preferAccountLanguage"`
	LocalizeDigits        bool   `json:"localizeDigits"`
	PublicCatalog         bool   `json:"publicCatalog"`
	CacheSeconds          int    `json:"cacheSeconds"`
	IncludeDetail         bool   `json:"includeDetail"`
}

type APIKeys struct {
	DefaultExpiryDays    int  `json:"defaultExpiryDays"`
	MaxExpiryDays        int  `json:"maxExpiryDays"`
	RequireExpiry        bool `json:"requireExpiry"`
	AllowWildcards       bool `json:"allowWildcards"`
	CatalogScopesOnly    bool `json:"catalogScopesOnly"`
	TrustProxyHeaders    bool `json:"trustProxyHeaders"`
	LastUsedEverySeconds int  `json:"lastUsedEverySeconds"`
}

type Schema struct {
	Access string `json:"access"`
}

type All struct {
	Errors    Errors  `json:"errors"`
	APIKeys   APIKeys `json:"apiKeys"`
	Schema    Schema  `json:"schema"`
	UpdatedBy string  `json:"updatedBy,omitempty"`
	UpdatedAt string  `json:"updatedAt,omitempty"`
}

const (
	ModeEnvelope = "envelope"
	ModeCode     = "code"

	AccessPublic  = "public"
	AccessSession = "session"
	AccessAPIKey  = "apikey"
	AccessConsole = "console"
)

func Defaults() All {
	return All{
		Errors: Errors{
			Mode:           ModeEnvelope,
			ModuleMode:     ModeCode,
			LocalizeDigits: true,
			PublicCatalog:  true,
			CacheSeconds:   300,
		},
		APIKeys: APIKeys{
			AllowWildcards:       true,
			CatalogScopesOnly:    true,
			LastUsedEverySeconds: 60,
		},
		Schema: Schema{Access: AccessSession},
	}
}

func (a All) Validate() (string, string) {
	switch a.Errors.Mode {
	case ModeEnvelope, ModeCode:
	default:
		return "errors.mode", "INVALID_MODE"
	}
	switch a.Errors.ModuleMode {
	case ModeEnvelope, ModeCode:
	default:
		return "errors.moduleMode", "INVALID_MODE"
	}
	if a.Errors.CacheSeconds < 0 || a.Errors.CacheSeconds > 86400 {
		return "errors.cacheSeconds", "VALUE_OUT_OF_RANGE"
	}
	if len(a.Errors.FallbackLang) > 12 {
		return "errors.fallbackLang", "INVALID_LANGUAGE_CODE"
	}
	if a.APIKeys.DefaultExpiryDays < 0 || a.APIKeys.DefaultExpiryDays > 3650 {
		return "apiKeys.defaultExpiryDays", "VALUE_OUT_OF_RANGE"
	}
	if a.APIKeys.MaxExpiryDays < 0 || a.APIKeys.MaxExpiryDays > 3650 {
		return "apiKeys.maxExpiryDays", "VALUE_OUT_OF_RANGE"
	}
	if a.APIKeys.MaxExpiryDays > 0 && a.APIKeys.DefaultExpiryDays > a.APIKeys.MaxExpiryDays {
		return "apiKeys.defaultExpiryDays", "DEFAULT_ABOVE_MAX"
	}
	if a.APIKeys.LastUsedEverySeconds < 0 || a.APIKeys.LastUsedEverySeconds > 3600 {
		return "apiKeys.lastUsedEverySeconds", "VALUE_OUT_OF_RANGE"
	}
	switch a.Schema.Access {
	case AccessPublic, AccessSession, AccessAPIKey, AccessConsole:
	default:
		return "schema.access", "INVALID_ACCESS"
	}
	return "", ""
}

var (
	mu      sync.Mutex
	cached  All
	fetched time.Time
	TTL     = 10 * time.Second
)

func Invalidate() {
	mu.Lock()
	fetched = time.Time{}
	mu.Unlock()
}

func Load(ctx context.Context, db *sql.DB) (All, error) {
	if db == nil {
		return Defaults(), nil
	}
	mu.Lock()
	if !fetched.IsZero() && time.Since(fetched) < TTL {
		out := cached
		mu.Unlock()
		return out, nil
	}
	mu.Unlock()
	out, err := Read(ctx, db)
	if err != nil {
		return Defaults(), err
	}
	mu.Lock()
	cached, fetched = out, time.Now()
	mu.Unlock()
	return out, nil
}

func Read(ctx context.Context, db *sql.DB) (All, error) {
	out := Defaults()
	var raw []byte
	var by string
	var at time.Time
	err := db.QueryRowContext(ctx, `SELECT value, updated_by, updated_at FROM devkit_settings WHERE id=1`).Scan(&raw, &by, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			return Defaults(), err
		}
	}
	out.UpdatedBy = by
	if by != "" {
		out.UpdatedAt = at.UTC().Format(time.RFC3339)
	}
	return out, nil
}

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	in.UpdatedBy, in.UpdatedAt = "", ""
	raw, err := json.Marshal(in)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO devkit_settings (id, value, updated_by, updated_at) VALUES (1, $1, $2, now())
		ON CONFLICT (id) DO UPDATE SET value=EXCLUDED.value, updated_by=EXCLUDED.updated_by, updated_at=now()`, raw, by)
	Invalidate()
	return err
}
