package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

const maxConsoleModules = 100

var consoleModuleID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`)

type consoleModules struct {
	Disabled []string `json:"disabled"`
}

func defaultConsoleModules() consoleModules {
	return consoleModules{Disabled: []string{}}
}

var consoleModulesStore = settingsrow.New("base_console_modules", defaultConsoleModules)

func (c consoleModules) clean() (consoleModules, error) {
	seen := make(map[string]bool, len(c.Disabled))
	out := make([]string, 0, len(c.Disabled))
	for _, raw := range c.Disabled {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] {
			continue
		}
		if !consoleModuleID.MatchString(id) {
			return c, badRequest("INVALID_MODULE", "disabled", "module ids use lowercase letters, digits and dashes")
		}
		seen[id] = true
		out = append(out, id)
	}
	if len(out) > maxConsoleModules {
		return c, badRequest("TOO_MANY_MODULES", "disabled", "too many modules")
	}
	sort.Strings(out)
	return consoleModules{Disabled: out}, nil
}

func (s *service) consoleModulesGet(w http.ResponseWriter, r *http.Request) {
	v, meta, err := consoleModulesStore.Read(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if v.Disabled == nil {
		v.Disabled = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"disabled": v.Disabled, "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt})
}

func (s *service) consoleModulesPut(w http.ResponseWriter, r *http.Request) {
	var raw map[string]json.RawMessage
	if err := readBody(r, &raw); err != nil {
		s.writeError(w, err)
		return
	}
	in := defaultConsoleModules()
	if b, ok := raw["disabled"]; ok {
		if err := json.Unmarshal(b, &in.Disabled); err != nil {
			s.writeError(w, badRequest("INVALID_JSON", "disabled", "disabled must be a list of module ids"))
			return
		}
	}
	v, err := in.clean()
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := consoleModulesStore.Save(r.Context(), s.db, v, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleModulesGet(w, r)
}
