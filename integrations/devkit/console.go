package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/internal/settings"
)

const (
	consoleUserHeader = "X-Nakama-Console-Username"
	maxMessage        = 500
)

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	routes := []route{
		{consolePrefix + "/settings", http.MethodGet, s.settingsGet},
		{consolePrefix + "/settings", http.MethodPut, s.settingsPut},
		{consolePrefix + "/errors", http.MethodGet, s.errorList},
		{consolePrefix + "/errors", http.MethodPost, s.errorCreate},
		{consolePrefix + "/errors/{code}", http.MethodGet, s.errorGet},
		{consolePrefix + "/errors/{code}", http.MethodPut, s.errorPut},
		{consolePrefix + "/errors/{code}", http.MethodDelete, s.errorReset},
		{consolePrefix + "/api-keys", http.MethodGet, s.keyList},
		{consolePrefix + "/api-keys", http.MethodPost, s.keyCreate},
		{consolePrefix + "/api-keys/{id}", http.MethodGet, s.keyGet},
		{consolePrefix + "/api-keys/{id}", http.MethodPut, s.keyUpdate},
		{consolePrefix + "/api-keys/{id}", http.MethodDelete, s.keyDelete},
		{consolePrefix + "/api-keys/{id}/revoke", http.MethodPost, s.keyRevoke},
		{consolePrefix + "/api-scopes", http.MethodGet, s.scopeList},
		{consolePrefix + "/api-scopes", http.MethodPost, s.scopeCreate},
		{consolePrefix + "/api-scopes/{name}", http.MethodDelete, s.scopeDelete},
		{consolePrefix + "/rpc-schema", http.MethodGet, s.schemaConsole},
		{consolePrefix + "/packages", http.MethodGet, s.consolePackages},
		{consolePrefix + "/packages/{name}/{version}", http.MethodDelete, s.consolePackageDelete},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

type consoleErr struct {
	status int
	code   string
	field  string
	msg    string
}

func (e *consoleErr) Error() string { return e.code }

func bad(code, field, msg string) error {
	return &consoleErr{status: http.StatusBadRequest, code: code, field: field, msg: msg}
}

func missing(code string) error {
	return &consoleErr{status: http.StatusNotFound, code: code}
}

func clash(code, msg string) error {
	return &consoleErr{status: http.StatusConflict, code: code, msg: msg}
}

func (s *service) fail(w http.ResponseWriter, err error) {
	var ce *consoleErr
	if errors.As(err, &ce) {
		body := map[string]string{"error": ce.code, "message": ce.msg}
		if ce.msg == "" {
			body["message"] = ce.code
		}
		if ce.field != "" {
			body["field"] = ce.field
		}
		writeJSON(w, ce.status, body)
		return
	}
	if e, ok := apierr.As(err); ok {
		writeJSON(w, e.Status, map[string]string{"error": e.Code, "message": e.Code, "field": e.Field})
		return
	}
	s.logger.Error("Devkit console request failed: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}

func readBody(r *http.Request, out any) error {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return bad("INVALID_JSON", "", "cannot read body")
	}
	if strings.TrimSpace(string(raw)) == "" {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return bad("INVALID_JSON", "", "request body is not valid JSON: "+err.Error())
	}
	return nil
}

func pathVar(r *http.Request, name string) string {
	if v := r.PathValue(name); v != "" {
		return v
	}
	return mux.Vars(r)[name]
}

func consoleUser(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(consoleUserHeader))
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults()})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := readBody(r, &in); err != nil {
		s.fail(w, err)
		return
	}
	in.Errors.FallbackLang = strings.ToLower(strings.TrimSpace(in.Errors.FallbackLang))
	if field, code := in.Validate(); code != "" {
		s.fail(w, bad(code, field, ""))
		return
	}
	if err := settings.Save(r.Context(), s.db, in, consoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	apierr.Invalidate()
	s.settingsGet(w, r)
}

type language struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Direction string `json:"direction"`
	IsDefault bool   `json:"isDefault"`
}

func (s *service) languages(ctx context.Context) ([]language, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT code, native_name, direction, is_default FROM base_language WHERE is_active ORDER BY is_default DESC, sort, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []language{}
	for rows.Next() {
		var l language
		if err := rows.Scan(&l.Code, &l.Name, &l.Direction, &l.IsDefault); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *service) errorList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	entries, _, err := apierr.ReadCatalog(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	langs, err := s.languages(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	term := strings.ToLower(strings.TrimSpace(q.Get("q")))
	module := strings.TrimSpace(q.Get("module"))
	missingLang := strings.TrimSpace(q.Get("missing"))
	state := strings.TrimSpace(q.Get("state"))
	modules := map[string]bool{}
	items := []apierr.Entry{}
	for _, code := range apierr.SortedCodes(entries) {
		e := entries[code]
		for _, m := range e.Modules {
			modules[m] = true
		}
		if module != "" && !contains(e.Modules, module) && !(module == "custom" && e.Custom) {
			continue
		}
		switch state {
		case "edited":
			if !e.Edited {
				continue
			}
		case "custom":
			if !e.Custom {
				continue
			}
		case "default":
			if e.Edited || e.Custom {
				continue
			}
		}
		if missingLang != "" && strings.TrimSpace(e.Messages[missingLang]) != "" {
			continue
		}
		if term != "" && !matches(e, term) {
			continue
		}
		items = append(items, e)
	}
	page, limit := pageArgs(q.Get("page"), q.Get("limit"))
	total := len(items)
	from := (page - 1) * limit
	if from > total {
		from = total
	}
	to := from + limit
	if to > total {
		to = total
	}
	mods := []string{}
	for m := range modules {
		mods = append(mods, m)
	}
	sort.Strings(mods)
	writeJSON(w, http.StatusOK, map[string]any{"items": items[from:to], "total": total, "page": page, "limit": limit, "languages": langs, "modules": mods})
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func matches(e apierr.Entry, term string) bool {
	if strings.Contains(strings.ToLower(e.Code), term) {
		return true
	}
	for _, m := range e.Messages {
		if strings.Contains(strings.ToLower(m), term) {
			return true
		}
	}
	return false
}

func pageArgs(p, l string) (int, int) {
	page, _ := strconv.Atoi(p)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(l)
	if limit < 1 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	return page, limit
}

func (s *service) errorGet(w http.ResponseWriter, r *http.Request) {
	entries, _, err := apierr.ReadCatalog(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	e, ok := entries[pathVar(r, "code")]
	if !ok {
		s.fail(w, missing("ERROR_CODE_NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, e)
}

type errorInput struct {
	Code              string            `json:"code"`
	Messages          map[string]string `json:"messages"`
	Status            *int              `json:"status"`
	RetryAfterSeconds *int              `json:"retryAfterSeconds"`
}

func (s *service) cleanMessages(ctx context.Context, in map[string]string, defaults map[string]string) (map[string]string, error) {
	langs, err := s.languages(ctx)
	if err != nil {
		return nil, err
	}
	known := map[string]bool{}
	for _, l := range langs {
		known[l.Code] = true
	}
	out := map[string]string{}
	for k, v := range in {
		k = strings.ToLower(strings.TrimSpace(k))
		if !known[k] {
			return nil, bad("INVALID_LANGUAGE", "messages."+k, "unknown or inactive language "+k)
		}
		v = strings.TrimSpace(v)
		if utf8.RuneCountInString(v) > maxMessage {
			return nil, bad("TEXT_TOO_LONG", "messages."+k, "a message can have at most 500 characters")
		}
		if v == "" || v == strings.TrimSpace(defaults[k]) {
			continue
		}
		out[k] = v
	}
	return out, nil
}

func checkStatus(st *int, retry *int) error {
	if st != nil && (*st < 400 || *st > 599) {
		return bad("INVALID_STATUS", "status", "status must be between 400 and 599")
	}
	if retry != nil && (*retry < 0 || *retry > 86400) {
		return bad("INVALID_RETRY", "retryAfterSeconds", "retry must be between 0 and 86400 seconds")
	}
	return nil
}

func (s *service) errorPut(w http.ResponseWriter, r *http.Request) {
	code := pathVar(r, "code")
	var in errorInput
	if err := readBody(r, &in); err != nil {
		s.fail(w, err)
		return
	}
	entries, _, err := apierr.ReadCatalog(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	cur, ok := entries[code]
	if !ok {
		s.fail(w, missing("ERROR_CODE_NOT_FOUND"))
		return
	}
	if err := checkStatus(in.Status, in.RetryAfterSeconds); err != nil {
		s.fail(w, err)
		return
	}
	msgs, err := s.cleanMessages(r.Context(), in.Messages, cur.DefaultMessages)
	if err != nil {
		s.fail(w, err)
		return
	}
	status := in.Status
	if status != nil && !cur.Custom && *status == cur.DefaultStatus {
		status = nil
	}
	if cur.Custom && status == nil {
		v := cur.Status
		status = &v
	}
	if !cur.Custom && len(msgs) == 0 && status == nil && in.RetryAfterSeconds == nil {
		if _, err := s.db.ExecContext(r.Context(), `DELETE FROM devkit_error_message WHERE code=$1`, code); err != nil {
			s.fail(w, err)
			return
		}
	} else {
		raw, _ := json.Marshal(msgs)
		if _, err := s.db.ExecContext(r.Context(), `INSERT INTO devkit_error_message (code, messages, status, retry_after_seconds, is_custom, updated_by, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,now())
			ON CONFLICT (code) DO UPDATE SET messages=EXCLUDED.messages, status=EXCLUDED.status, retry_after_seconds=EXCLUDED.retry_after_seconds, updated_by=EXCLUDED.updated_by, updated_at=now()`,
			code, raw, nullInt(status), nullInt(in.RetryAfterSeconds), cur.Custom, consoleUser(r)); err != nil {
			s.fail(w, err)
			return
		}
	}
	apierr.Invalidate()
	s.errorGet(w, r)
}

func nullInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func (s *service) errorCreate(w http.ResponseWriter, r *http.Request) {
	var in errorInput
	if err := readBody(r, &in); err != nil {
		s.fail(w, err)
		return
	}
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	if !apierr.CodePattern.MatchString(in.Code) {
		s.fail(w, bad("INVALID_CODE", "code", "use upper case letters, digits and _"))
		return
	}
	if in.Status == nil {
		v := http.StatusBadRequest
		in.Status = &v
	}
	if err := checkStatus(in.Status, in.RetryAfterSeconds); err != nil {
		s.fail(w, err)
		return
	}
	entries, _, err := apierr.ReadCatalog(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	if _, ok := entries[in.Code]; ok {
		s.fail(w, clash("CODE_EXISTS", "this code already exists"))
		return
	}
	msgs, err := s.cleanMessages(r.Context(), in.Messages, nil)
	if err != nil {
		s.fail(w, err)
		return
	}
	if len(msgs) == 0 {
		s.fail(w, bad("MESSAGE_REQUIRED", "messages", "write the message in at least one language"))
		return
	}
	raw, _ := json.Marshal(msgs)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO devkit_error_message (code, messages, status, retry_after_seconds, is_custom, updated_by) VALUES ($1,$2,$3,$4,true,$5)`,
		in.Code, raw, *in.Status, nullInt(in.RetryAfterSeconds), consoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	apierr.Invalidate()
	entries, _, err = apierr.ReadCatalog(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, entries[in.Code])
}

func (s *service) errorReset(w http.ResponseWriter, r *http.Request) {
	code := pathVar(r, "code")
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM devkit_error_message WHERE code=$1`, code)
	if err != nil {
		s.fail(w, err)
		return
	}
	apierr.Invalidate()
	if n, _ := res.RowsAffected(); n == 0 {
		if _, ok := apierr.Defaults()[code]; !ok {
			s.fail(w, missing("ERROR_CODE_NOT_FOUND"))
			return
		}
	}
	if _, ok := apierr.Defaults()[code]; !ok {
		writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "code": code})
		return
	}
	s.errorGet(w, r)
}

type keyView struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Prefix      string     `json:"prefix"`
	Scopes      []string   `json:"scopes"`
	AllowedIPs  []string   `json:"allowedIps"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	RevokedAt   *time.Time `json:"revokedAt"`
	RevokedBy   string     `json:"revokedBy"`
	LastUsedAt  *time.Time `json:"lastUsedAt"`
	LastUsedIP  string     `json:"lastUsedIp"`
	UseCount    int64      `json:"useCount"`
	CreatedBy   string     `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	Status      string     `json:"status"`
}

const keyColumns = `id::text, name, description, lookup, array_to_json(scopes)::text, array_to_json(allowed_ips)::text, expires_at, revoked_at, revoked_by, last_used_at, last_used_ip, use_count, created_by, created_at`

func scanKey(row interface{ Scan(...any) error }, now time.Time) (keyView, error) {
	var k keyView
	var lookup, scopes, ips string
	var exp, rev, used sql.NullTime
	if err := row.Scan(&k.ID, &k.Name, &k.Description, &lookup, &scopes, &ips, &exp, &rev, &k.RevokedBy, &used, &k.LastUsedIP, &k.UseCount, &k.CreatedBy, &k.CreatedAt); err != nil {
		return k, err
	}
	k.Prefix = apikey.Prefix + lookup
	k.Scopes, k.AllowedIPs = []string{}, []string{}
	_ = json.Unmarshal([]byte(scopes), &k.Scopes)
	_ = json.Unmarshal([]byte(ips), &k.AllowedIPs)
	if k.Scopes == nil {
		k.Scopes = []string{}
	}
	if k.AllowedIPs == nil {
		k.AllowedIPs = []string{}
	}
	if exp.Valid {
		k.ExpiresAt = &exp.Time
	}
	if rev.Valid {
		k.RevokedAt = &rev.Time
	}
	if used.Valid {
		k.LastUsedAt = &used.Time
	}
	switch {
	case rev.Valid:
		k.Status = "revoked"
	case exp.Valid && !exp.Time.After(now):
		k.Status = "expired"
	default:
		k.Status = "active"
	}
	return k, nil
}

func (s *service) keyList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	where := []string{}
	args := []any{}
	switch q.Get("status") {
	case "active":
		where = append(where, "revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now())")
	case "revoked":
		where = append(where, "revoked_at IS NOT NULL")
	case "expired":
		where = append(where, "revoked_at IS NULL AND expires_at IS NOT NULL AND expires_at <= now()")
	}
	if term := strings.TrimSpace(q.Get("q")); term != "" {
		args = append(args, "%"+strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.ToLower(term))+"%")
		where = append(where, "(lower(name) LIKE $1 OR lookup LIKE $1 OR EXISTS (SELECT 1 FROM unnest(scopes) sc WHERE sc LIKE $1))")
	}
	sqlWhere := ""
	if len(where) > 0 {
		sqlWhere = " WHERE " + strings.Join(where, " AND ")
	}
	page, limit := pageArgs(q.Get("page"), q.Get("limit"))
	var total int
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM devkit_api_key`+sqlWhere, args...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+keyColumns+` FROM devkit_api_key`+sqlWhere+` ORDER BY created_at DESC, id LIMIT `+strconv.Itoa(limit)+` OFFSET `+strconv.Itoa((page-1)*limit), args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []keyView{}
	now := s.now()
	for rows.Next() {
		k, err := scanKey(rows, now)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, k)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) loadKey(ctx context.Context, id string) (keyView, error) {
	if _, err := uuid.FromString(id); err != nil {
		return keyView{}, missing("KEY_NOT_FOUND")
	}
	k, err := scanKey(s.db.QueryRowContext(ctx, `SELECT `+keyColumns+` FROM devkit_api_key WHERE id=$1`, id), s.now())
	if errors.Is(err, sql.ErrNoRows) {
		return k, missing("KEY_NOT_FOUND")
	}
	return k, err
}

func (s *service) keyGet(w http.ResponseWriter, r *http.Request) {
	k, err := s.loadKey(r.Context(), pathVar(r, "id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, k)
}

type keyInput struct {
	Name         *string   `json:"name"`
	Description  *string   `json:"description"`
	Scopes       *[]string `json:"scopes"`
	AllowedIPs   *[]string `json:"allowedIps"`
	ExpiresAt    *string   `json:"expiresAt"`
	ExpiresIn    *int      `json:"expiresInDays"`
	NeverExpires bool      `json:"neverExpires"`
}

func (s *service) cleanScopes(ctx context.Context, cfg settings.All, list []string) ([]string, error) {
	seen := map[string]bool{}
	out := []string{}
	catalog := map[string]bool{}
	if cfg.APIKeys.CatalogScopesOnly {
		rows, err := s.db.QueryContext(ctx, `SELECT name FROM devkit_api_scope`)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				rows.Close()
				return nil, err
			}
			catalog[n] = true
		}
		rows.Close()
	}
	for _, sc := range list {
		sc = strings.ToLower(strings.TrimSpace(sc))
		if sc == "" || seen[sc] {
			continue
		}
		if !apikey.ValidGrant(sc, cfg.APIKeys.AllowWildcards) {
			return nil, bad("INVALID_SCOPE", "scopes", "scope "+sc+" is not valid")
		}
		if cfg.APIKeys.CatalogScopesOnly && !strings.Contains(sc, "*") && !catalog[sc] {
			return nil, bad("SCOPE_NOT_IN_CATALOG", "scopes", "scope "+sc+" is not in the scope list")
		}
		seen[sc] = true
		out = append(out, sc)
	}
	if len(out) == 0 {
		return nil, bad("SCOPES_REQUIRED", "scopes", "pick at least one scope")
	}
	sort.Strings(out)
	return out, nil
}

func cleanIPs(list []string) ([]string, error) {
	out := []string{}
	for _, ip := range list {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if strings.Contains(ip, "/") {
			if _, _, err := net.ParseCIDR(ip); err != nil {
				return nil, bad("INVALID_IP", "allowedIps", ip+" is not an IP or CIDR range")
			}
		} else if net.ParseIP(ip) == nil {
			return nil, bad("INVALID_IP", "allowedIps", ip+" is not an IP or CIDR range")
		}
		out = append(out, ip)
	}
	if len(out) > 50 {
		return nil, bad("TOO_MANY_IPS", "allowedIps", "at most 50 entries")
	}
	return out, nil
}

func (s *service) expiry(cfg settings.All, in keyInput, creating bool) (*time.Time, bool, error) {
	now := s.now()
	var at *time.Time
	set := false
	switch {
	case in.ExpiresAt != nil && strings.TrimSpace(*in.ExpiresAt) != "":
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*in.ExpiresAt))
		if err != nil {
			return nil, false, bad("INVALID_DATE", "expiresAt", "use an RFC 3339 date")
		}
		at, set = &t, true
	case in.ExpiresIn != nil && *in.ExpiresIn > 0:
		t := now.Add(time.Duration(*in.ExpiresIn) * 24 * time.Hour)
		at, set = &t, true
	case in.NeverExpires || (in.ExpiresAt != nil && strings.TrimSpace(*in.ExpiresAt) == ""):
		set = true
	case creating && cfg.APIKeys.DefaultExpiryDays > 0:
		t := now.Add(time.Duration(cfg.APIKeys.DefaultExpiryDays) * 24 * time.Hour)
		at, set = &t, true
	case creating:
		set = true
	}
	if !set {
		return nil, false, nil
	}
	if at == nil && cfg.APIKeys.RequireExpiry {
		return nil, false, bad("EXPIRY_REQUIRED", "expiresAt", "an expiry date is required")
	}
	if at != nil && !at.After(now) {
		return nil, false, bad("EXPIRY_IN_PAST", "expiresAt", "the expiry date has passed")
	}
	if cfg.APIKeys.MaxExpiryDays > 0 {
		limit := now.Add(time.Duration(cfg.APIKeys.MaxExpiryDays)*24*time.Hour + time.Minute)
		if at == nil {
			return nil, false, bad("EXPIRY_REQUIRED", "expiresAt", "keys must expire within the maximum lifetime")
		}
		if at.After(limit) {
			return nil, false, bad("EXPIRY_TOO_FAR", "expiresAt", "the expiry is beyond the maximum lifetime")
		}
	}
	return at, true, nil
}

func cleanName(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", bad("NAME_REQUIRED", "name", "name is required")
	}
	if utf8.RuneCountInString(v) > 80 {
		return "", bad("NAME_TOO_LONG", "name", "at most 80 characters")
	}
	return v, nil
}

func (s *service) keyCreate(w http.ResponseWriter, r *http.Request) {
	var in keyInput
	if err := readBody(r, &in); err != nil {
		s.fail(w, err)
		return
	}
	cfg, _ := settings.Read(r.Context(), s.db)
	name := ""
	if in.Name != nil {
		name = *in.Name
	}
	name, err := cleanName(name)
	if err != nil {
		s.fail(w, err)
		return
	}
	desc := ""
	if in.Description != nil {
		desc = strings.TrimSpace(*in.Description)
	}
	scopes := []string{}
	if in.Scopes != nil {
		scopes = *in.Scopes
	}
	if scopes, err = s.cleanScopes(r.Context(), cfg, scopes); err != nil {
		s.fail(w, err)
		return
	}
	ips := []string{}
	if in.AllowedIPs != nil {
		if ips, err = cleanIPs(*in.AllowedIPs); err != nil {
			s.fail(w, err)
			return
		}
	}
	exp, _, err := s.expiry(cfg, in, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	full, lookup, hash, err := apikey.Generate()
	if err != nil {
		s.fail(w, err)
		return
	}
	id := uuid.Must(uuid.NewV4()).String()
	sc, _ := json.Marshal(scopes)
	ip, _ := json.Marshal(ips)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO devkit_api_key (id, name, description, lookup, secret_hash, scopes, allowed_ips, expires_at, created_by)
		VALUES ($1,$2,$3,$4,$5,ARRAY(SELECT jsonb_array_elements_text($6::jsonb)),ARRAY(SELECT jsonb_array_elements_text($7::jsonb)),$8,$9)`,
		id, name, desc, lookup, hash, string(sc), string(ip), exp, consoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	k, err := s.loadKey(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": k, "key": full})
}

func (s *service) keyUpdate(w http.ResponseWriter, r *http.Request) {
	cur, err := s.loadKey(r.Context(), pathVar(r, "id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if cur.Status == "revoked" {
		s.fail(w, clash("KEY_REVOKED", "a revoked key cannot be changed"))
		return
	}
	var in keyInput
	if err := readBody(r, &in); err != nil {
		s.fail(w, err)
		return
	}
	cfg, _ := settings.Read(r.Context(), s.db)
	name, desc, scopes, ips := cur.Name, cur.Description, cur.Scopes, cur.AllowedIPs
	if in.Name != nil {
		if name, err = cleanName(*in.Name); err != nil {
			s.fail(w, err)
			return
		}
	}
	if in.Description != nil {
		desc = strings.TrimSpace(*in.Description)
	}
	if in.Scopes != nil {
		if scopes, err = s.cleanScopes(r.Context(), cfg, *in.Scopes); err != nil {
			s.fail(w, err)
			return
		}
	}
	if in.AllowedIPs != nil {
		if ips, err = cleanIPs(*in.AllowedIPs); err != nil {
			s.fail(w, err)
			return
		}
	}
	exp, changed, err := s.expiry(cfg, in, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	if !changed {
		exp = cur.ExpiresAt
	}
	sc, _ := json.Marshal(scopes)
	ip, _ := json.Marshal(ips)
	if _, err := s.db.ExecContext(r.Context(), `UPDATE devkit_api_key SET name=$2, description=$3, scopes=ARRAY(SELECT jsonb_array_elements_text($4::jsonb)), allowed_ips=ARRAY(SELECT jsonb_array_elements_text($5::jsonb)), expires_at=$6 WHERE id=$1`,
		cur.ID, name, desc, string(sc), string(ip), exp); err != nil {
		s.fail(w, err)
		return
	}
	s.keyGet(w, r)
}

func (s *service) keyRevoke(w http.ResponseWriter, r *http.Request) {
	cur, err := s.loadKey(r.Context(), pathVar(r, "id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if cur.Status != "revoked" {
		if _, err := s.db.ExecContext(r.Context(), `UPDATE devkit_api_key SET revoked_at=now(), revoked_by=$2 WHERE id=$1 AND revoked_at IS NULL`, cur.ID, consoleUser(r)); err != nil {
			s.fail(w, err)
			return
		}
	}
	s.keyGet(w, r)
}

func (s *service) keyDelete(w http.ResponseWriter, r *http.Request) {
	cur, err := s.loadKey(r.Context(), pathVar(r, "id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if cur.Status == "active" {
		s.fail(w, clash("KEY_ACTIVE", "revoke the key before deleting it"))
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `DELETE FROM devkit_api_key WHERE id=$1`, cur.ID); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": cur.ID})
}

func (s *service) scopeList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT s.name, s.description, s.source, s.created_at,
		(SELECT count(*) FROM devkit_api_key k WHERE s.name = ANY(k.scopes) AND k.revoked_at IS NULL) FROM devkit_api_scope s ORDER BY s.name`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	type scopeView struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Source      string    `json:"source"`
		CreatedAt   time.Time `json:"createdAt"`
		Keys        int       `json:"keys"`
	}
	items := []scopeView{}
	for rows.Next() {
		var v scopeView
		if err := rows.Scan(&v.Name, &v.Description, &v.Source, &v.CreatedAt, &v.Keys); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, v)
	}
	cfg, _ := settings.Load(r.Context(), s.db)
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "allowWildcards": cfg.APIKeys.AllowWildcards, "catalogOnly": cfg.APIKeys.CatalogScopesOnly})
}

func (s *service) scopeCreate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := readBody(r, &in); err != nil {
		s.fail(w, err)
		return
	}
	in.Name = strings.ToLower(strings.TrimSpace(in.Name))
	if !apikey.ScopePattern.MatchString(in.Name) {
		s.fail(w, bad("INVALID_SCOPE", "name", "use resource:action, e.g. assets:write"))
		return
	}
	if utf8.RuneCountInString(in.Description) > 200 {
		s.fail(w, bad("TEXT_TOO_LONG", "description", "at most 200 characters"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `INSERT INTO devkit_api_scope (name, description, source) VALUES ($1,$2,'console') ON CONFLICT (name) DO NOTHING`, in.Name, strings.TrimSpace(in.Description))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, clash("SCOPE_EXISTS", "this scope already exists"))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"name": in.Name, "description": strings.TrimSpace(in.Description), "source": "console"})
}

func (s *service) scopeDelete(w http.ResponseWriter, r *http.Request) {
	name := pathVar(r, "name")
	var used int
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM devkit_api_key WHERE $1 = ANY(scopes) AND revoked_at IS NULL`, name).Scan(&used); err != nil {
		s.fail(w, err)
		return
	}
	if used > 0 {
		s.fail(w, clash("SCOPE_IN_USE", "active keys still use this scope"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM devkit_api_scope WHERE name=$1`, name)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, missing("SCOPE_NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "name": name})
}

func (s *service) schemaConsole(w http.ResponseWriter, r *http.Request) {
	view, err := s.schemaView(strings.TrimSpace(r.URL.Query().Get("module")))
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}
