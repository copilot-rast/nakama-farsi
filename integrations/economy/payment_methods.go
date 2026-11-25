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
	Default  string `json:"default,omitempty"`
}

func mf(name, typ string, required bool) methodField {
	return methodField{Name: name, Type: typ, Required: required}
}

var methodFields = map[string][]methodField{
	"zarinpal":    {mf("merchantId", "secret", true), mf("sandbox", "bool", false)},
	"zibal":       {mf("merchant", "secret", true), mf("sandbox", "bool", false)},
	"stripe":      {mf("secretKey", "secret", true), mf("webhookSecret", "secret", true), mf("publishableKey", "text", false), {Name: "currency", Type: "text", Default: defaultStripeCurrency}, mf("testMode", "bool", false)},
	"bazaar":      {mf("packageName", "text", true), mf("apiSecret", "secret", true), mf("rsaPublicKey", "textarea", false), mf("sandbox", "bool", false)},
	"myket":       {mf("packageName", "text", true), mf("accessToken", "secret", true), mf("rsaPublicKey", "textarea", false), mf("sandbox", "bool", false)},
	"google_play": {mf("packageName", "text", true), mf("serviceAccountJson", "secretJson", true), mf("sandbox", "bool", false)},
	"apple":       {mf("bundleId", "text", true), mf("sharedSecret", "secret", false), mf("issuerId", "text", false), mf("keyId", "text", false), mf("privateKeyP8", "secretText", false), mf("sandbox", "bool", false)},
	"in_app":      {},
}

func secretType(typ string) bool {
	return typ == "secret" || typ == "secretJson" || typ == "secretText"
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

func sandboxField(method string) string {
	if isStripe(method) {
		return "testMode"
	}
	return "sandbox"
}

func methodSandbox(method string, cfg methodConfig) bool {
	return cfg.flag(sandboxField(method))
}

func fakeSandbox(method string) bool {
	return !isAppStore(method) && !isStripe(method)
}

func methodConfigured(method string, cfg methodConfig) bool {
	sandbox := methodSandbox(method, cfg) && fakeSandbox(method)
	for _, f := range methodFields[method] {
		if !f.Required {
			continue
		}
		if sandbox && secretType(f.Type) {
			continue
		}
		if cfg.str(f.Name) == "" {
			return false
		}
	}
	if method == "apple" {
		return cfg.str("sharedSecret") != "" || appleAPIKeys(cfg)
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
		switch {
		case f.Type == "bool":
			masked[f.Name] = pm.Config.flag(f.Name)
		case secretType(f.Type):
			if pm.Config.str(f.Name) != "" {
				masked[f.Name] = secretMask
			} else {
				masked[f.Name] = ""
			}
		default:
			masked[f.Name] = pm.Config.str(f.Name)
			if masked[f.Name] == "" && f.Default != "" {
				masked[f.Name] = f.Default
			}
		}
	}
	fields := methodFields[pm.Method]
	if fields == nil {
		fields = []methodField{}
	}
	return map[string]any{"method": pm.Method, "isActive": pm.Active, "configured": methodConfigured(pm.Method, pm.Config), "fields": fields, "config": masked,
		"sandboxField": sandboxField(pm.Method), "sandboxNeedsKeys": !fakeSandbox(pm.Method)}
}

func (s *service) consoleMethodsList(w http.ResponseWriter, r *http.Request) {
	out := []map[string]any{}
	for _, m := range allMethods {
		pm, err := s.loadMethod(r.Context(), s.db, m)
		if err != nil {
			s.writeError(w, err)
			return
		}
		view := methodView(pm)
		if isStore(m) {
			view["health"] = s.storeHealth(r.Context(), m)
		}
		out = append(out, view)
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
				if secretType(f.Type) || !given {
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
			if len(text) > 16384 {
				return nil, badRequest("TEXT_TOO_LONG", f.Name, "value is too long")
			}
			if secretType(f.Type) && text == secretMask {
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
	if err := checkMethodConfig(method, next); err != nil {
		return nil, err
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
	sandboxOff := false
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
		sandboxOff = methodSandbox(method, old.Config) && !methodSandbox(method, next)
		return nil
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	if isStore(method) {
		view["health"] = s.storeHealth(r.Context(), method)
	}
	if sandboxOff {
		if st, err := loadPaymentSettings(r.Context(), s.db); err == nil && st.CloseSandboxOnOff {
			view["closedSandboxPayments"] = s.closeSandboxPending(r.Context(), method)
		}
	}
	writeJSON(w, http.StatusOK, view)
}

func checkMethodConfig(method string, cfg methodConfig) error {
	switch method {
	case "google_play":
		if raw := cfg.str("serviceAccountJson"); raw != "" {
			if _, err := googleServiceAccount(raw); err != nil {
				return badRequest("INVALID_VALUE", "serviceAccountJson", err.Error())
			}
		}
	case "apple":
		if raw := cfg.str("privateKeyP8"); raw != "" {
			if _, err := parseAppleKey(raw); err != nil {
				return badRequest("INVALID_VALUE", "privateKeyP8", "not a PKCS#8 EC private key (.p8)")
			}
		}
		given := 0
		for _, k := range []string{"issuerId", "keyId", "privateKeyP8"} {
			if cfg.str(k) != "" {
				given++
			}
		}
		if given != 0 && given != 3 {
			return badRequest("APPLE_API_KEYS_INCOMPLETE", "issuerId", "set issuerId, keyId and privateKeyP8 together, or none of them")
		}
	case "stripe":
		return checkStripeConfig(cfg)
	}
	return nil
}
