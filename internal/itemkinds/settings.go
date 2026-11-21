package itemkinds

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Routing struct {
	Enabled         bool   `json:"enabled"`
	OnMissingRef    string `json:"onMissingRef"`
	OnDelete        string `json:"onDelete"`
	LinkWritesUnity bool   `json:"linkWritesUnity"`
	MaxDepth        int    `json:"maxDepth"`
	KeepRuns        int    `json:"keepRuns"`
}

type Inventory struct {
	Profile string `json:"profile"`
}

type Settings struct {
	Routing   Routing                      `json:"routing"`
	Inventory Inventory                    `json:"inventory"`
	Ownership map[string]map[string]string `json:"ownership"`
	Kinds     map[string]map[string]any    `json:"kinds"`
}

func Defaults() Settings {
	return Settings{
		Routing:   Routing{Enabled: true, OnMissingRef: "pending", OnDelete: "deactivate", LinkWritesUnity: true, MaxDepth: 4, KeepRuns: 50},
		Inventory: Inventory{Profile: "ask"},
		Ownership: map[string]map[string]string{},
		Kinds:     map[string]map[string]any{},
	}
}

func (s Settings) Validate() (string, string) {
	switch s.Routing.OnMissingRef {
	case "pending", "reject":
	default:
		return "routing.onMissingRef", "INVALID_SETTING"
	}
	switch s.Routing.OnDelete {
	case "deactivate", "keep":
	default:
		return "routing.onDelete", "INVALID_SETTING"
	}
	if s.Routing.MaxDepth < 1 || s.Routing.MaxDepth > 8 {
		return "routing.maxDepth", "INVALID_SETTING"
	}
	if s.Routing.KeepRuns < 0 || s.Routing.KeepRuns > 1000 {
		return "routing.keepRuns", "INVALID_SETTING"
	}
	switch s.Inventory.Profile {
	case "ask", "always", "never":
	default:
		return "inventory.profile", "INVALID_SETTING"
	}
	for kind, m := range s.Ownership {
		for field, v := range m {
			if v != "unity" && v != "console" {
				return "ownership." + kind + "." + field, "INVALID_SETTING"
			}
		}
	}
	return "", ""
}

var store = settingsrow.New("itemkind_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) Settings {
	s, _ := store.Load(ctx, db)
	return s
}

func Read(ctx context.Context, db *sql.DB) (Settings, settingsrow.Meta, error) {
	return store.Read(ctx, db)
}

func Save(ctx context.Context, db *sql.DB, s Settings, by string) error {
	return store.Save(ctx, db, s, by)
}

func (s Settings) KindConfig(k *Kind) map[string]any {
	out := map[string]any{"enabled": true}
	for key, v := range k.Defaults {
		out[key] = v
	}
	for key, v := range s.Kinds[k.Key] {
		out[key] = v
	}
	return out
}

func (s Settings) KindEnabled(k *Kind) bool {
	v, ok := s.KindConfig(k)["enabled"].(bool)
	return !ok || v
}

func (s Settings) Owner(k *Kind, field string) string {
	if v := s.Ownership[k.Key][field]; v != "" {
		return v
	}
	for _, o := range k.Ownership {
		if o.Field == field && o.Default != "" {
			return o.Default
		}
	}
	return "unity"
}

func Bool(cfg map[string]any, key string, def bool) bool {
	if v, ok := cfg[key].(bool); ok {
		return v
	}
	return def
}

func String(cfg map[string]any, key, def string) string {
	if v, ok := cfg[key].(string); ok && v != "" {
		return v
	}
	return def
}

func Number(cfg map[string]any, key string, def float64) float64 {
	switch v := cfg[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return def
}
