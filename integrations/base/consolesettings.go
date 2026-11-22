package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

const defaultConsoleTimezone = "Asia/Tehran"

type consoleSettings struct {
	Timezone string `json:"timezone"`
}

func defaultConsoleSettings() consoleSettings {
	return consoleSettings{Timezone: defaultConsoleTimezone}
}

var consoleSettingsStore = settingsrow.New("base_console_settings", defaultConsoleSettings)

func (c consoleSettings) clean() (consoleSettings, error) {
	c.Timezone = strings.TrimSpace(c.Timezone)
	if c.Timezone == "" {
		c.Timezone = defaultConsoleTimezone
	}
	if strings.EqualFold(c.Timezone, "local") {
		return c, badRequest("INVALID_TIMEZONE", "timezone", "use an IANA time zone like Asia/Tehran")
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return c, badRequest("INVALID_TIMEZONE", "timezone", "use an IANA time zone like Asia/Tehran")
	}
	return c, nil
}

func (s *service) consoleSettingsGet(w http.ResponseWriter, r *http.Request) {
	v, meta, err := consoleSettingsStore.Read(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if v.Timezone == "" {
		v.Timezone = defaultConsoleTimezone
	}
	writeJSON(w, http.StatusOK, map[string]any{"timezone": v.Timezone, "defaultTimezone": defaultConsoleTimezone, "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt})
}

func (s *service) consoleSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultConsoleSettings()
	var raw map[string]json.RawMessage
	if err := readBody(r, &raw); err != nil {
		s.writeError(w, err)
		return
	}
	for _, k := range []string{"updatedBy", "updatedAt", "defaultTimezone"} {
		delete(raw, k)
	}
	b, _ := json.Marshal(raw)
	if err := json.Unmarshal(b, &in); err != nil {
		s.writeError(w, badRequest("INVALID_JSON", "", "settings are not valid"))
		return
	}
	v, err := in.clean()
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := consoleSettingsStore.Save(r.Context(), s.db, v, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleSettingsGet(w, r)
}
