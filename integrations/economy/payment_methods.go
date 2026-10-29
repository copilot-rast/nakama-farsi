package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const secretMask = "********"

type methodField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

var methodFields = map[string][]methodField{
	"zarinpal": {{"merchantId", "secret", true}, {"sandbox", "bool", false}},
	"zibal":    {{"merchant", "secret", true}, {"sandbox", "bool", false}},
	"bazaar":   {{"packageName", "text", true}, {"apiSecret", "secret", true}, {"rsaPublicKey", "textarea", false}, {"sandbox", "bool", false}},
	"myket":    {{"packageName", "text", true}, {"accessToken", "secret", true}, {"rsaPublicKey", "textarea", false}, {"sandbox", "bool", false}},
	"in_app":   {},
}

type methodConfig map[string]any

func (c methodConfig) str(name string) string {
	v, _ := c[name].(string)
	return strings.TrimSpace(v)
}

func (c methodConfig) flag(name string) bool {
	v, _ := c[name].(bool)
	return v
}

func methodConfigured(method string, cfg methodConfig) bool {
	sandbox := cfg.flag("sandbox")
	for _, f := range methodFields[method] {
		if !f.Required {
			continue
		}
		if sandbox && f.Type == "secret" {
			continue
		}
		if cfg.str(f.Name) == "" {
			return false
		}
	}
	return true
}

type paymentMethod struct {
	Method string
	Active bool
	Config methodConfig
}

func (m paymentMethod) usable() bool {
	return m.Active && methodConfigured(m.Method, m.Config)
}

func (s *service) loadMethod(ctx context.Context, q queryer, method string) (paymentMethod, error) {
	pm := paymentMethod{Method: method, Config: methodConfig{}}
	var sealed string
	err := q.QueryRowContext(ctx, `SELECT is_active, config FROM economy_payment_method WHERE method=$1`, method).Scan(&pm.Active, &sealed)
	if errors.Is(err, sql.ErrNoRows) {
		return pm, nil
	}
	if err != nil {
		return pm, err
	}
	plain, err := unseal(s.cfg.sealKey, sealed)
	if err != nil {
		return pm, errors.New("payment method config cannot be decrypted; was ECONOMY_SECRET_KEY changed?")
	}
	if len(plain) > 0 {
		if err := json.Unmarshal(plain, &pm.Config); err != nil {
			return pm, err
		}
	}
	return pm, nil
}

func (s *service) activeMethods(ctx context.Context) (map[string]paymentMethod, error) {
	out := map[string]paymentMethod{}
	for _, m := range allMethods {
		pm, err := s.loadMethod(ctx, s.db, m)
		if err != nil {
			return nil, err
		}
		if pm.usable() {
			out[m] = pm
		}
	}
	return out, nil
}

func methodView(pm paymentMethod) map[string]any {
	masked := map[string]any{}
	for _, f := range methodFields[pm.Method] {
		switch f.Type {
		case "bool":
			masked[f.Name] = pm.Config.flag(f.Name)
		case "secret":
			if pm.Config.str(f.Name) != "" {
				masked[f.Name] = secretMask
			} else {
				masked[f.Name] = ""
			}
		default:
			masked[f.Name] = pm.Config.str(f.Name)
		}
	}
	fields := methodFields[pm.Method]
	if fields == nil {
		fields = []methodField{}
	}
	return map[string]any{"method": pm.Method, "isActive": pm.Active, "configured": methodConfigured(pm.Method, pm.Config), "fields": fields, "config": masked}
}

func (s *service) consoleMethodsList(w http.ResponseWriter, r *http.Request) {
	out := []map[string]any{}
	for _, m := range allMethods {
		pm, err := s.loadMethod(r.Context(), s.db, m)
		if err != nil {
			s.writeError(w, err)
			return
		}
		out = append(out, methodView(pm))
	}
	writeJSON(w, http.StatusOK, out)
}

func mergeMethodConfig(method string, old methodConfig, in map[string]any) (methodConfig, error) {
	next := methodConfig{}
	for _, f := range methodFields[method] {
		v, given := in[f.Name]
		switch f.Type {
		case "bool":
			if !given {
				next[f.Name] = old.flag(f.Name)
				continue
			}
			b, ok := v.(bool)
			if !ok {
				return nil, badRequest("INVALID_VALUE", f.Name, "must be true or false")
			}
			next[f.Name] = b
		default:
			if !given || v == nil {
				if f.Type == "secret" || !given {
					if old.str(f.Name) != "" {
						next[f.Name] = old.str(f.Name)
					}
				}
				continue
			}
			text, ok := v.(string)
			if !ok {
				return nil, badRequest("INVALID_VALUE", f.Name, "must be a string")
			}
			text = strings.TrimSpace(text)
			if len(text) > 8192 {
				return nil, badRequest("TEXT_TOO_LONG", f.Name, "value is too long")
			}
			if f.Type == "secret" && text == secretMask {
				if old.str(f.Name) != "" {
					next[f.Name] = old.str(f.Name)
				}
				continue
			}
			if text != "" {
				next[f.Name] = text
			}
		}
	}
	return next, nil
}

func (s *service) consoleMethodUpdate(w http.ResponseWriter, r *http.Request) {
	method := pathVar(r, "method")
	if !validMethod(method) {
		s.writeError(w, notFound("PAYMENT_METHOD_NOT_FOUND"))
		return
	}
	var in struct {
		IsActive bool           `json:"isActive"`
		Config   map[string]any `json:"config"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	var view map[string]any
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO economy_payment_method (method) VALUES ($1) ON CONFLICT (method) DO NOTHING`, method); err != nil {
			return err
		}
		var lock string
		if err := tx.QueryRowContext(r.Context(), `SELECT method FROM economy_payment_method WHERE method=$1 FOR UPDATE`, method).Scan(&lock); err != nil {
			return err
		}
		old, err := s.loadMethod(r.Context(), tx, method)
		if err != nil {
			return err
		}
		if in.Config == nil {
			in.Config = map[string]any{}
		}
		next, err := mergeMethodConfig(method, old.Config, in.Config)
		if err != nil {
			return err
		}
		if in.IsActive && !methodConfigured(method, next) {
			return badRequest("PAYMENT_METHOD_NOT_CONFIGURED", "config", "fill the required fields before activating")
		}
		plain, _ := json.Marshal(next)
		sealed, err := seal(s.cfg.sealKey, plain)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE economy_payment_method SET is_active=$2, config=$3, updated_at=now() WHERE method=$1`, method, in.IsActive, sealed); err != nil {
			return err
		}
		view = methodView(paymentMethod{Method: method, Active: in.IsActive, Config: next})
		return nil
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}
