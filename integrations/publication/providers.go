package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const secretMask = "********"

type providerField struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
	Default  string   `json:"default,omitempty"`
}

type providerKind struct {
	Channel string          `json:"channel"`
	Kind    string          `json:"kind"`
	Fields  []providerField `json:"fields"`
}

var providerCatalog = []providerKind{
	{"sms", "kavenegar", []providerField{{Name: "apiKey", Type: "secret", Required: true}, {Name: "sender", Type: "text"}}},
	{"sms", "smsir", []providerField{{Name: "apiKey", Type: "secret", Required: true}, {Name: "lineNumber", Type: "number", Required: true}}},
	{"sms", "twilio", []providerField{{Name: "accountSid", Type: "text", Required: true}, {Name: "authToken", Type: "secret", Required: true}, {Name: "from", Type: "text", Required: true}}},
	{"email", "smtp", []providerField{
		{Name: "host", Type: "text", Required: true},
		{Name: "port", Type: "number", Required: true, Default: "587"},
		{Name: "security", Type: "select", Required: true, Options: []string{"starttls", "tls", "none"}, Default: "starttls"},
		{Name: "username", Type: "text"},
		{Name: "password", Type: "secret"},
		{Name: "from", Type: "email", Required: true},
		{Name: "fromName", Type: "text"},
	}},
	{"email", "resend", []providerField{{Name: "apiKey", Type: "secret", Required: true}, {Name: "from", Type: "email", Required: true}, {Name: "fromName", Type: "text"}}},
}

func catalogFor(channel, kind string) (providerKind, bool) {
	for _, k := range providerCatalog {
		if k.Channel == channel && k.Kind == kind {
			return k, true
		}
	}
	return providerKind{}, false
}

type providerConfig map[string]string

func (c providerConfig) str(name string) string { return strings.TrimSpace(c[name]) }

type provider struct {
	ID        string
	Channel   string
	Kind      string
	Active    bool
	Default   bool
	Sandbox   bool
	Config    providerConfig
	UpdatedAt time.Time
}

func providerConfigured(p provider) bool {
	def, ok := catalogFor(p.Channel, p.Kind)
	if !ok {
		return false
	}
	for _, f := range def.Fields {
		if !f.Required {
			continue
		}
		if p.Sandbox && f.Type == "secret" {
			continue
		}
		if p.Config.str(f.Name) == "" {
			return false
		}
	}
	return true
}

const providerColumns = `id::text, channel, kind, is_active, is_default, sandbox, config, updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *service) scanProvider(row rowScanner) (provider, error) {
	var p provider
	var sealed string
	if err := row.Scan(&p.ID, &p.Channel, &p.Kind, &p.Active, &p.Default, &p.Sandbox, &sealed, &p.UpdatedAt); err != nil {
		return p, err
	}
	p.Config = providerConfig{}
	plain, err := unseal(s.cfg.sealKey, sealed)
	if err != nil {
		return p, errors.New("provider config cannot be decrypted; was ECONOMY_SECRET_KEY changed?")
	}
	if len(plain) > 0 {
		if err := json.Unmarshal(plain, &p.Config); err != nil {
			return p, err
		}
	}
	return p, nil
}

func (s *service) loadProvider(ctx context.Context, q queryer, id string, lock bool) (provider, error) {
	if _, err := uuid.FromString(id); err != nil {
		return provider{}, notFound("PROVIDER_NOT_FOUND")
	}
	query := `SELECT ` + providerColumns + ` FROM publication_provider WHERE id=$1`
	if lock {
		query += ` FOR UPDATE`
	}
	p, err := s.scanProvider(q.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return p, notFound("PROVIDER_NOT_FOUND")
	}
	return p, err
}

func (s *service) defaultProvider(ctx context.Context, q queryer, channel string) (provider, error) {
	p, err := s.scanProvider(q.QueryRowContext(ctx, `SELECT `+providerColumns+` FROM publication_provider WHERE channel=$1 AND is_active
		ORDER BY is_default DESC, updated_at DESC LIMIT 1`, channel))
	if errors.Is(err, sql.ErrNoRows) {
		return p, badRequest("PROVIDER_NOT_CONFIGURED", "providerId", "no active "+channel+" provider; add one under providers")
	}
	return p, err
}

func providerView(p provider) map[string]any {
	def, _ := catalogFor(p.Channel, p.Kind)
	masked := map[string]any{}
	for _, f := range def.Fields {
		if f.Type == "secret" {
			if p.Config.str(f.Name) != "" {
				masked[f.Name] = secretMask
			} else {
				masked[f.Name] = ""
			}
			continue
		}
		masked[f.Name] = p.Config.str(f.Name)
	}
	fields := def.Fields
	if fields == nil {
		fields = []providerField{}
	}
	return map[string]any{"id": p.ID, "channel": p.Channel, "kind": p.Kind, "isActive": p.Active, "isDefault": p.Default, "sandbox": p.Sandbox,
		"configured": providerConfigured(p), "fields": fields, "config": masked, "updatedAt": p.UpdatedAt}
}

func mergeProviderConfig(def providerKind, old providerConfig, in map[string]any) (providerConfig, error) {
	next := providerConfig{}
	for _, f := range def.Fields {
		v, given := in[f.Name]
		if !given || v == nil {
			if old.str(f.Name) != "" {
				next[f.Name] = old.str(f.Name)
			} else if f.Default != "" {
				next[f.Name] = f.Default
			}
			continue
		}
		var text string
		switch t := v.(type) {
		case string:
			text = strings.TrimSpace(t)
		case float64:
			if f.Type != "number" {
				return nil, badRequest("INVALID_VALUE", f.Name, "must be a string")
			}
			text = strconv.FormatFloat(t, 'f', -1, 64)
		default:
			return nil, badRequest("INVALID_VALUE", f.Name, "must be a string")
		}
		if len(text) > 4096 || strings.ContainsAny(text, "\r\n") {
			return nil, badRequest("INVALID_VALUE", f.Name, "value is too long or has line breaks")
		}
		if f.Type == "secret" && text == secretMask {
			if old.str(f.Name) != "" {
				next[f.Name] = old.str(f.Name)
			}
			continue
		}
		if text == "" {
			continue
		}
		switch f.Type {
		case "number":
			n, err := strconv.ParseInt(text, 10, 64)
			if err != nil || n <= 0 || (f.Name == "port" && n > 65535) {
				return nil, badRequest("INVALID_VALUE", f.Name, "must be a positive whole number")
			}
		case "email":
			if _, ok := normalizeEmail(text); !ok {
				return nil, badRequest("INVALID_VALUE", f.Name, "must be an email address")
			}
		case "select":
			found := false
			for _, o := range f.Options {
				found = found || o == text
			}
			if !found {
				return nil, badRequest("INVALID_VALUE", f.Name, "must be one of "+strings.Join(f.Options, ", "))
			}
		}
		if f.Name == "host" && strings.ContainsAny(text, " /:") {
			return nil, badRequest("INVALID_VALUE", f.Name, "must be a host name without scheme or port")
		}
		if f.Name == "from" && def.Kind == "twilio" {
			if _, ok := normalizePhone(text); !ok && !strings.HasPrefix(text, "MG") {
				return nil, badRequest("INVALID_VALUE", f.Name, "must be an E.164 number")
			}
		}
		next[f.Name] = text
	}
	return next, nil
}

func (s *service) consoleProvidersList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+providerColumns+` FROM publication_provider ORDER BY channel DESC, is_default DESC, kind`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		p, err := s.scanProvider(rows)
		if err != nil {
			s.writeError(w, err)
			return
		}
		items = append(items, providerView(p))
	}
	if err := rows.Err(); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "catalog": providerCatalog})
}

type providerInput struct {
	Channel   string         `json:"channel"`
	Kind      string         `json:"kind"`
	IsActive  *bool          `json:"isActive"`
	IsDefault *bool          `json:"isDefault"`
	Sandbox   *bool          `json:"sandbox"`
	Config    map[string]any `json:"config"`
}

func (s *service) consoleProviderCreate(w http.ResponseWriter, r *http.Request) {
	s.providerSave(w, r, true)
}

func (s *service) consoleProviderUpdate(w http.ResponseWriter, r *http.Request) {
	s.providerSave(w, r, false)
}

func (s *service) providerSave(w http.ResponseWriter, r *http.Request, create bool) {
	var in providerInput
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	var view map[string]any
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var p provider
		if create {
			if _, ok := catalogFor(in.Channel, in.Kind); !ok {
				return badRequest("INVALID_PROVIDER_KIND", "kind", "unknown channel/kind pair")
			}
			p = provider{ID: uuid.Must(uuid.NewV4()).String(), Channel: in.Channel, Kind: in.Kind, Sandbox: true, Config: providerConfig{}}
		} else {
			var err error
			if p, err = s.loadProvider(r.Context(), tx, pathVar(r, "id"), true); err != nil {
				return err
			}
			if (in.Channel != "" && in.Channel != p.Channel) || (in.Kind != "" && in.Kind != p.Kind) {
				return badRequest("PROVIDER_KIND_FIXED", "kind", "channel and kind cannot change; create another provider")
			}
		}
		def, _ := catalogFor(p.Channel, p.Kind)
		if in.Config == nil {
			in.Config = map[string]any{}
		}
		next, err := mergeProviderConfig(def, p.Config, in.Config)
		if err != nil {
			return err
		}
		p.Config = next
		if in.Sandbox != nil {
			p.Sandbox = *in.Sandbox
		}
		if in.IsActive != nil {
			p.Active = *in.IsActive
		}
		if in.IsDefault != nil {
			p.Default = *in.IsDefault
		}
		if p.Active && !providerConfigured(p) {
			return badRequest("PROVIDER_NOT_CONFIGURED", "config", "fill the required fields before activating")
		}
		if p.Default && !p.Active {
			return badRequest("DEFAULT_MUST_BE_ACTIVE", "isDefault", "only an active provider can be the default")
		}
		if p.Default {
			if _, err := tx.ExecContext(r.Context(), `UPDATE publication_provider SET is_default=false, updated_at=now() WHERE channel=$1 AND id<>$2 AND is_default`, p.Channel, p.ID); err != nil {
				return err
			}
		}
		plain, _ := json.Marshal(p.Config)
		sealed, err := seal(s.cfg.sealKey, plain)
		if err != nil {
			return err
		}
		if create {
			_, err = tx.ExecContext(r.Context(), `INSERT INTO publication_provider (id, channel, kind, is_active, is_default, sandbox, config) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				p.ID, p.Channel, p.Kind, p.Active, p.Default, p.Sandbox, sealed)
		} else {
			_, err = tx.ExecContext(r.Context(), `UPDATE publication_provider SET is_active=$2, is_default=$3, sandbox=$4, config=$5, updated_at=now() WHERE id=$1`,
				p.ID, p.Active, p.Default, p.Sandbox, sealed)
		}
		if err != nil {
			if _, ok := isUniqueViolation(err); ok {
				return fail(http.StatusConflict, "ALREADY_EXISTS", "kind", "a provider of this kind already exists for the channel")
			}
			return err
		}
		saved, err := s.loadProvider(r.Context(), tx, p.ID, false)
		if err != nil {
			return err
		}
		view = providerView(saved)
		return nil
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	status := http.StatusOK
	if create {
		status = http.StatusCreated
	}
	writeJSON(w, status, view)
}

func (s *service) consoleProviderDelete(w http.ResponseWriter, r *http.Request) {
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		p, err := s.loadProvider(r.Context(), tx, pathVar(r, "id"), true)
		if err != nil {
			return err
		}
		var busy bool
		if err := tx.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM publication_notification WHERE provider=$1 AND status IN ('queued','sending'))`, p.ID).Scan(&busy); err != nil {
			return err
		}
		if busy {
			return conflict("IN_USE", "notifications are still being delivered with this provider")
		}
		_, err = tx.ExecContext(r.Context(), `DELETE FROM publication_provider WHERE id=$1`, p.ID)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) consoleProviderTest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		To      string `json:"to"`
		Message string `json:"message"`
		Subject string `json:"subject"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	p, err := s.loadProvider(r.Context(), s.db, pathVar(r, "id"), false)
	if err != nil {
		s.writeError(w, err)
		return
	}
	in.Message = strings.TrimSpace(in.Message)
	if in.Message == "" || len([]rune(in.Message)) > 600 {
		s.writeError(w, badRequest("INVALID_MESSAGE", "message", "message is required (max 600 characters)"))
		return
	}
	langs, err := localized.Load(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	msg := outMessage{Text: in.Message, Lang: langs.Default}
	switch p.Channel {
	case "sms":
		addr, ok := normalizePhone(in.To)
		if !ok {
			s.writeError(w, badRequest("INVALID_PHONE", "to", "use 09xxxxxxxxx, +989… or an E.164 number"))
			return
		}
		msg.To = addr
	default:
		addr, ok := normalizeEmail(in.To)
		if !ok {
			s.writeError(w, badRequest("INVALID_EMAIL", "to", "must be an email address"))
			return
		}
		msg.To = addr
		msg.Subject = strings.TrimSpace(in.Subject)
		if msg.Subject == "" {
			msg.Subject = "Test message"
		}
		msg.HTML = emailHTML(msg.Subject, renderMarkdown(in.Message), langs.Default, langs.Direction(langs.Default))
		msg.Text = markdownText(in.Message)
	}
	if !providerConfigured(p) {
		s.writeError(w, badRequest("PROVIDER_NOT_CONFIGURED", "config", "fill the required fields first"))
		return
	}
	if p.Sandbox {
		s.logger.Info("Publication sandbox test via %s/%s to %s: %s", p.Channel, p.Kind, maskAddress(msg.To), truncate(in.Message, 200))
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sandbox": true, "reply": "sandbox: nothing was sent"})
		return
	}
	reply, err := s.deliver(r.Context(), p, msg)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "sandbox": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sandbox": false, "reply": reply})
}

func maskAddress(a string) string {
	if at := strings.IndexByte(a, '@'); at > 0 {
		if at <= 2 {
			return a[:1] + "***" + a[at:]
		}
		return a[:2] + "***" + a[at:]
	}
	if len(a) > 6 {
		return a[:4] + strings.Repeat("*", len(a)-6) + a[len(a)-2:]
	}
	return a
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
