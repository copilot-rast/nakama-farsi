package handoffkit

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

const (
	AudienceAll    = "all"
	AudienceAdmins = "admins"
	AdminHeader    = "X-Nakama-Console-Admin"
)

type Settings struct {
	Enabled       bool   `json:"enabled"`
	Audience      string `json:"audience"`
	ShowServerKey bool   `json:"showServerKey"`
}

func DefaultSettings() Settings {
	return Settings{Enabled: true, Audience: AudienceAll, ShowServerKey: true}
}

func (s Settings) Clean() (Settings, bool) {
	switch s.Audience {
	case AudienceAll, AudienceAdmins:
		return s, true
	case "":
		s.Audience = AudienceAll
		return s, true
	}
	return s, false
}

func (s Settings) Allows(admin bool) bool {
	if !s.Enabled {
		return false
	}
	return admin || s.Audience != AudienceAdmins
}

type rulesRow struct {
	Handoff Settings `json:"handoff"`
}

var store = settingsrow.New("base_server_rules", func() rulesRow { return rulesRow{Handoff: DefaultSettings()} })

func LoadSettings(ctx context.Context, db *sql.DB) Settings {
	v, err := store.Load(ctx, db)
	if err != nil {
		return DefaultSettings()
	}
	out, ok := v.Handoff.Clean()
	if !ok {
		return DefaultSettings()
	}
	return out
}

func Invalidate() {
	store.Invalidate()
}
