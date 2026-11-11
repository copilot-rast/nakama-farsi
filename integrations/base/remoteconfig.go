package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/fnv"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var (
	configKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{1,63}$`)
	configTypes      = []string{"string", "number", "bool", "json", "text"}
	ruleNamePattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,39}$`)
)

type configRule struct {
	Name       string          `json:"name"`
	Value      json.RawMessage `json:"value"`
	Platforms  []string        `json:"platforms"`
	MinVersion string          `json:"minVersion"`
	MaxVersion string          `json:"maxVersion"`
	TagIDs     []string        `json:"tagIds"`
	Countries  []string        `json:"countries"`
	Percent    *float64        `json:"percent"`
}

type configKey struct {
	ID          string          `json:"id"`
	Key         string          `json:"key"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Value       json.RawMessage `json:"value"`
	Rules       []configRule    `json:"rules"`
	IsActive    bool            `json:"isActive"`
	Version     int             `json:"version"`
	UpdatedBy   string          `json:"updatedBy"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	UpdatedAt   string          `json:"updatedAt,omitempty"`
}

type configTarget struct {
	UserID   string
	Platform string
	Version  string
	Country  string
	Tags     map[string]bool
}

func bucket(key, userID string) int {
	h := fnv.New64a()
	h.Write([]byte(key + ":" + userID))
	return int(h.Sum64() % 10000)
}

func (r configRule) matches(key string, t configTarget) bool {
	if len(r.Platforms) > 0 && !oneOfList(t.Platform, r.Platforms) {
		return false
	}
	if r.MinVersion != "" || r.MaxVersion != "" {
		if t.Version == "" {
			return false
		}
		if r.MinVersion != "" {
			if c, ok := compareVersions(t.Version, r.MinVersion); !ok || c < 0 {
				return false
			}
		}
		if r.MaxVersion != "" {
			if c, ok := compareVersions(t.Version, r.MaxVersion); !ok || c > 0 {
				return false
			}
		}
	}
	if len(r.Countries) > 0 && !oneOfList(t.Country, r.Countries) {
		return false
	}
	if len(r.TagIDs) > 0 {
		hit := false
		for _, id := range r.TagIDs {
			if t.Tags[id] {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	if r.Percent != nil && *r.Percent < 100 {
		if t.UserID == "" {
			return false
		}
		return float64(bucket(key, t.UserID)) < math.Round(*r.Percent*100)
	}
	return true
}

func (k configKey) resolve(t configTarget) (json.RawMessage, string) {
	for _, r := range k.Rules {
		if r.matches(k.Key, t) {
			return r.Value, r.Name
		}
	}
	return k.Value, "default"
}

func cleanConfigValue(langs *localized.Set, typ, field string, raw json.RawMessage) (json.RawMessage, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, badRequest("VALUE_REQUIRED", field, "a value is required")
	}
	bad := func() error { return badRequest("INVALID_VALUE", field, "value does not match the type "+typ) }
	switch typ {
	case "string":
		var v string
		if json.Unmarshal(raw, &v) != nil {
			return nil, bad()
		}
		if utf8.RuneCountInString(v) > 2000 {
			return nil, badRequest("VALUE_TOO_LONG", field, "text is at most 2000 characters")
		}
		b, _ := json.Marshal(v)
		return b, nil
	case "number":
		var v float64
		if json.Unmarshal(raw, &v) != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, bad()
		}
		b, _ := json.Marshal(v)
		return b, nil
	case "bool":
		var v bool
		if json.Unmarshal(raw, &v) != nil {
			return nil, bad()
		}
		b, _ := json.Marshal(v)
		return b, nil
	case "json":
		var v any
		if json.Unmarshal(raw, &v) != nil {
			return nil, bad()
		}
		if len(raw) > 16384 {
			return nil, badRequest("VALUE_TOO_LONG", field, "JSON is at most 16 KB")
		}
		var buf bytes.Buffer
		if err := json.Compact(&buf, raw); err != nil {
			return nil, bad()
		}
		return buf.Bytes(), nil
	case "text":
		var v localized.Text
		if json.Unmarshal(raw, &v) != nil {
			return nil, bad()
		}
		clean, err := cleanText(langs, field, v, true, 2000)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(clean.JSON()), nil
	}
	return nil, badRequest("INVALID_TYPE", "type", "unknown type")
}

func (s *service) cleanConfigKey(ctx context.Context, tx *sql.Tx, in configKey) (configKey, error) {
	in.Key = strings.TrimSpace(in.Key)
	if !configKeyPattern.MatchString(in.Key) {
		return in, badRequest("INVALID_KEY", "key", "key is 2 to 64 lower-case letters, digits, dots or underscores and starts with a letter")
	}
	if !oneOfList(in.Type, configTypes) {
		return in, badRequest("INVALID_TYPE", "type", "type must be string, number, bool, json or text")
	}
	in.Description = strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(in.Description) > 500 {
		return in, badRequest("DESCRIPTION_TOO_LONG", "description", "description is at most 500 characters")
	}
	langs, err := s.languages(ctx, tx)
	if err != nil {
		return in, err
	}
	if in.Value, err = cleanConfigValue(langs, in.Type, "value", in.Value); err != nil {
		return in, err
	}
	if len(in.Rules) > 20 {
		return in, badRequest("TOO_MANY_RULES", "rules", "at most 20 rules")
	}
	names := map[string]bool{}
	for i := range in.Rules {
		r := &in.Rules[i]
		field := "rules." + strconv.Itoa(i)
		r.Name = strings.TrimSpace(r.Name)
		if r.Name == "" {
			r.Name = "rule-" + strconv.Itoa(i+1)
		}
		if !ruleNamePattern.MatchString(r.Name) || r.Name == "default" {
			return in, badRequest("INVALID_RULE_NAME", field+".name", "rule names are short lower-case keys and not default")
		}
		if names[r.Name] {
			return in, badRequest("DUPLICATE_RULE_NAME", field+".name", "rule names must be different")
		}
		names[r.Name] = true
		if r.Value, err = cleanConfigValue(langs, in.Type, field+".value", r.Value); err != nil {
			return in, err
		}
		for _, p := range r.Platforms {
			if !oneOfList(p, appPlatforms) {
				return in, badRequest("INVALID_PLATFORM", field+".platforms", "unknown platform")
			}
		}
		if r.MinVersion, err = cleanVersion(field+".minVersion", r.MinVersion, false); err != nil {
			return in, err
		}
		if r.MaxVersion, err = cleanVersion(field+".maxVersion", r.MaxVersion, false); err != nil {
			return in, err
		}
		if r.MinVersion != "" && r.MaxVersion != "" {
			if c, _ := compareVersions(r.MinVersion, r.MaxVersion); c > 0 {
				return in, badRequest("MIN_ABOVE_MAX", field+".maxVersion", "lowest version is above the highest")
			}
		}
		for j, c := range r.Countries {
			c = strings.ToUpper(strings.TrimSpace(c))
			if len(c) != 2 {
				return in, badRequest("INVALID_COUNTRY_CODE", field+".countries", "country codes are two letters")
			}
			r.Countries[j] = c
		}
		if len(r.TagIDs) > 0 {
			var found int
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM base_tag WHERE key_type='user' AND id::text = ANY($1)`, r.TagIDs).Scan(&found); err != nil {
				return in, err
			}
			if found != len(r.TagIDs) {
				return in, badRequest("TAG_NOT_FOUND", field+".tagIds", "player tag does not exist")
			}
		}
		if r.Percent != nil && (*r.Percent < 0 || *r.Percent > 100 || math.IsNaN(*r.Percent)) {
			return in, badRequest("INVALID_PERCENT", field+".percent", "percent is 0 to 100")
		}
		if r.Percent != nil {
			p := math.Round(*r.Percent*100) / 100
			r.Percent = &p
		}
		if r.Platforms == nil {
			r.Platforms = []string{}
		}
		if r.Countries == nil {
			r.Countries = []string{}
		}
		if r.TagIDs == nil {
			r.TagIDs = []string{}
		}
	}
	if in.Rules == nil {
		in.Rules = []configRule{}
	}
	return in, nil
}

const configCols = `id::text, key, type, description, value::text, rules::text, is_active, version, updated_by, created_at, updated_at`

func scanConfigKey(sc interface{ Scan(...any) error }) (configKey, error) {
	var k configKey
	var value, rules string
	var created, updated time.Time
	if err := sc.Scan(&k.ID, &k.Key, &k.Type, &k.Description, &value, &rules, &k.IsActive, &k.Version, &k.UpdatedBy, &created, &updated); err != nil {
		return k, err
	}
	k.Value = json.RawMessage(value)
	_ = json.Unmarshal([]byte(rules), &k.Rules)
	if k.Rules == nil {
		k.Rules = []configRule{}
	}
	k.CreatedAt = created.UTC().Format(time.RFC3339)
	k.UpdatedAt = updated.UTC().Format(time.RFC3339)
	return k, nil
}

func loadConfigKey(ctx context.Context, q queryer, id string, lock bool) (configKey, error) {
	if _, err := uuid.FromString(id); err != nil {
		return configKey{}, notFound("NOT_FOUND")
	}
	suffix := ""
	if lock {
		suffix = " FOR UPDATE"
	}
	k, err := scanConfigKey(q.QueryRowContext(ctx, `SELECT `+configCols+` FROM base_config_key WHERE id=$1`+suffix, id))
	if errors.Is(err, sql.ErrNoRows) {
		return k, notFound("NOT_FOUND")
	}
	return k, err
}

func addConfigHistory(ctx context.Context, tx *sql.Tx, k configKey, action, by string) error {
	snap, _ := json.Marshal(map[string]any{"key": k.Key, "type": k.Type, "description": k.Description, "value": k.Value, "rules": k.Rules, "isActive": k.IsActive})
	_, err := tx.ExecContext(ctx, `INSERT INTO base_config_history (id, key_id, key, version, action, snapshot, changed_by) VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7)`,
		uuid.Must(uuid.NewV4()).String(), k.ID, k.Key, k.Version, action, string(snap), by)
	return err
}

func (s *service) configList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(c.key ILIKE "+n+" OR c.description ILIKE "+n+")")
	}
	activeFilter(r, where, "c.is_active")
	if v := r.URL.Query().Get("type"); oneOfList(v, configTypes) {
		where.add("c.type = ?", v)
	}
	out, err := s.listRaw(r, ` FROM base_config_key c`, `jsonb_build_object('id',c.id,'key',c.key,'type',c.type,'description',c.description,'value',c.value,'rules',c.rules,'ruleCount',jsonb_array_length(c.rules),'isActive',c.is_active,'version',c.version,'updatedBy',c.updated_by,'createdAt',c.created_at,'updatedAt',c.updated_at)`, "c.key", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) configGet(w http.ResponseWriter, r *http.Request) {
	k, err := loadConfigKey(r.Context(), s.db, pathVar(r, "id"), false)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, k)
}

func (s *service) configWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in configKey
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		me := agentName(r)
		var out configKey
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			if create {
				in.ID = uuid.Must(uuid.NewV4()).String()
			} else {
				old, err := loadConfigKey(r.Context(), tx, pathVar(r, "id"), true)
				if err != nil {
					return err
				}
				in.ID = old.ID
				if old.Type != in.Type && in.Type != "" {
					return conflict("TYPE_LOCKED", "the type cannot change; make a new key")
				}
				in.Type = old.Type
				in.Version = old.Version + 1
			}
			clean, err := s.cleanConfigKey(r.Context(), tx, in)
			if err != nil {
				return err
			}
			rules, _ := json.Marshal(clean.Rules)
			if create {
				clean.Version = 1
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_config_key (id, key, type, description, value, rules, is_active, version, updated_by) VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7,1,$8)`,
					clean.ID, clean.Key, clean.Type, clean.Description, string(clean.Value), string(rules), clean.IsActive, me)
			} else {
				_, err = tx.ExecContext(r.Context(), `UPDATE base_config_key SET key=$2, description=$3, value=$4::jsonb, rules=$5::jsonb, is_active=$6, version=$7, updated_by=$8, updated_at=now() WHERE id=$1`,
					clean.ID, clean.Key, clean.Description, string(clean.Value), string(rules), clean.IsActive, clean.Version, me)
			}
			if err != nil {
				if _, unique := isUniqueViolation(err); unique {
					return fail(http.StatusConflict, "ALREADY_EXISTS", "key", "a key with this name already exists")
				}
				return err
			}
			action := "update"
			if create {
				action = "create"
			}
			if err := addConfigHistory(r.Context(), tx, clean, action, me); err != nil {
				return err
			}
			out, err = loadConfigKey(r.Context(), tx, clean.ID, false)
			return err
		})
		if err != nil {
			s.writeError(w, err)
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeJSON(w, status, out)
	}
}

func (s *service) configDelete(w http.ResponseWriter, r *http.Request) {
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		k, err := loadConfigKey(r.Context(), tx, pathVar(r, "id"), true)
		if err != nil {
			return err
		}
		k.Version++
		if err := addConfigHistory(r.Context(), tx, k, "delete", agentName(r)); err != nil {
			return err
		}
		_, err = tx.ExecContext(r.Context(), `DELETE FROM base_config_key WHERE id=$1`, k.ID)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) configHistory(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	p := readPage(r)
	where := &sqlWhere{}
	where.add("h.key_id = ?::uuid", id)
	out, err := s.listRaw(r, ` FROM base_config_history h`, `jsonb_build_object('id',h.id,'keyId',h.key_id,'key',h.key,'version',h.version,'action',h.action,'snapshot',h.snapshot,'changedBy',h.changed_by,'createdAt',h.created_at)`, "h.version DESC, h.created_at DESC", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) configRecentHistory(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		where.add("h.key ILIKE ?", likePattern(p.Q))
	}
	out, err := s.listRaw(r, ` FROM base_config_history h`, `jsonb_build_object('id',h.id,'keyId',h.key_id,'key',h.key,'version',h.version,'action',h.action,'snapshot',h.snapshot,'changedBy',h.changed_by,'createdAt',h.created_at,'exists',EXISTS (SELECT 1 FROM base_config_key k WHERE k.id=h.key_id))`, "h.created_at DESC", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) configRollback(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Version int `json:"version"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	me := agentName(r)
	var out configKey
	status := http.StatusOK
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var snap string
		err := tx.QueryRowContext(r.Context(), `SELECT snapshot::text FROM base_config_history WHERE key_id=$1 AND version=$2 AND action <> 'delete' ORDER BY created_at DESC LIMIT 1`, id, in.Version).Scan(&snap)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("VERSION_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		var k configKey
		if err := json.Unmarshal([]byte(snap), &k); err != nil {
			return err
		}
		k.ID = id
		old, err := loadConfigKey(r.Context(), tx, id, true)
		exists := err == nil
		if err != nil {
			if app, ok := asApp(err); !ok || app.Status != http.StatusNotFound {
				return err
			}
			var last int
			if err := tx.QueryRowContext(r.Context(), `SELECT COALESCE(max(version),0) FROM base_config_history WHERE key_id=$1`, id).Scan(&last); err != nil {
				return err
			}
			k.Version = last + 1
			status = http.StatusCreated
		} else {
			k.Version = old.Version + 1
		}
		clean, err := s.cleanConfigKey(r.Context(), tx, k)
		if err != nil {
			return err
		}
		rules, _ := json.Marshal(clean.Rules)
		if exists {
			_, err = tx.ExecContext(r.Context(), `UPDATE base_config_key SET key=$2, description=$3, value=$4::jsonb, rules=$5::jsonb, is_active=$6, version=$7, updated_by=$8, updated_at=now() WHERE id=$1`,
				clean.ID, clean.Key, clean.Description, string(clean.Value), string(rules), clean.IsActive, clean.Version, me)
		} else {
			_, err = tx.ExecContext(r.Context(), `INSERT INTO base_config_key (id, key, type, description, value, rules, is_active, version, updated_by) VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7,$8,$9)`,
				clean.ID, clean.Key, clean.Type, clean.Description, string(clean.Value), string(rules), clean.IsActive, clean.Version, me)
		}
		if err != nil {
			if _, unique := isUniqueViolation(err); unique {
				return fail(http.StatusConflict, "ALREADY_EXISTS", "key", "another key now uses this name")
			}
			return err
		}
		if err := addConfigHistory(r.Context(), tx, clean, "rollback", me); err != nil {
			return err
		}
		out, err = loadConfigKey(r.Context(), tx, id, false)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, status, out)
}

func (s *service) activeConfigKeys(ctx context.Context, keys []string) ([]configKey, error) {
	q := `SELECT ` + configCols + ` FROM base_config_key WHERE is_active`
	args := []any{}
	if len(keys) > 0 {
		q += ` AND key = ANY($1)`
		args = append(args, keys)
	}
	rows, err := s.db.QueryContext(ctx, q+` ORDER BY key`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []configKey{}
	for rows.Next() {
		k, err := scanConfigKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *service) userTagSet(ctx context.Context, userID string) (map[string]bool, error) {
	tags := map[string]bool{}
	if userID == "" {
		return tags, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT tag_id::text FROM base_user_tag WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			tags[id] = true
		}
	}
	return tags, rows.Err()
}

func (s *service) userCountry(ctx context.Context, userID string) string {
	var c sql.NullString
	_ = s.db.QueryRowContext(ctx, `SELECT metadata->>'country' FROM users WHERE id=$1`, userID).Scan(&c)
	return strings.ToUpper(c.String)
}

func (s *service) resolveConfig(ctx context.Context, keys []configKey, t configTarget, langs *localized.Set, lang string) (map[string]any, map[string]string) {
	values := map[string]any{}
	variants := map[string]string{}
	for _, k := range keys {
		raw, variant := k.resolve(t)
		var v any
		if k.Type == "text" {
			var txt localized.Text
			_ = json.Unmarshal(raw, &txt)
			v = langs.Resolve(txt, lang)
		} else {
			_ = json.Unmarshal(raw, &v)
		}
		values[k.Key] = v
		variants[k.Key] = variant
	}
	return values, variants
}

func configHash(values map[string]any) string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		b, _ := json.Marshal(values[k])
		h.Write([]byte(k + "=" + string(b) + ";"))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (s *service) rpcRemoteConfig(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Platform string   `json:"platform"`
		Version  string   `json:"version"`
		Keys     []string `json:"keys"`
		Lang     string   `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	in.Platform = strings.ToLower(strings.TrimSpace(in.Platform))
	if in.Platform != "" && !oneOfList(in.Platform, appPlatforms) {
		return nil, badRequest("INVALID_PLATFORM", "platform", "unknown platform")
	}
	in.Version = strings.TrimPrefix(strings.TrimSpace(in.Version), "v")
	if in.Version != "" {
		if _, ok := parseVersion(in.Version); !ok {
			return nil, badRequest("INVALID_VERSION", "version", "version is not valid")
		}
	}
	if len(in.Keys) > 100 {
		return nil, badRequest("TOO_MANY_KEYS", "keys", "ask for at most 100 keys")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	tags, err := s.userTagSet(ctx, userID)
	if err != nil {
		return nil, err
	}
	country := s.userCountry(ctx, userID)
	if country == "" {
		country = activity.FromContext(ctx, activity.Entry{}).Country
	}
	keys, err := s.activeConfigKeys(ctx, in.Keys)
	if err != nil {
		return nil, err
	}
	values, variants := s.resolveConfig(ctx, keys, configTarget{UserID: userID, Platform: in.Platform, Version: in.Version, Country: country, Tags: tags}, langs, lang)
	hash := configHash(values)
	if len(in.Keys) == 0 {
		seen := map[string]string{}
		for k, v := range variants {
			seen[k] = v + "|" + configHash(map[string]any{k: values[k]})[:8]
		}
		s.noteConfigChange(ctx, userID, hash, seen)
	}
	return map[string]any{"lang": lang, "values": values, "variants": variants, "hash": hash}, nil
}

func (s *service) noteConfigChange(ctx context.Context, userID, hash string, variants map[string]string) {
	var oldHash string
	var oldRaw string
	err := s.db.QueryRowContext(ctx, `SELECT hash, variants::text FROM base_config_seen WHERE user_id=$1`, userID).Scan(&oldHash, &oldRaw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err == nil && oldHash == hash {
		return
	}
	raw, _ := json.Marshal(variants)
	res, uerr := s.db.ExecContext(ctx, `INSERT INTO base_config_seen (user_id, hash, variants, updated_at) VALUES ($1,$2,$3::jsonb,now())
		ON CONFLICT (user_id) DO UPDATE SET hash=EXCLUDED.hash, variants=EXCLUDED.variants, updated_at=now() WHERE base_config_seen.hash <> EXCLUDED.hash`, userID, hash, string(raw))
	if uerr != nil {
		return
	}
	if n, _ := res.RowsAffected(); n == 0 || errors.Is(err, sql.ErrNoRows) {
		return
	}
	old := map[string]string{}
	_ = json.Unmarshal([]byte(oldRaw), &old)
	changed := []string{}
	for k, v := range variants {
		if old[k] != v {
			name, _, _ := strings.Cut(v, "|")
			changed = append(changed, k+":"+name)
		}
	}
	for k := range old {
		if _, ok := variants[k]; !ok {
			changed = append(changed, k+":removed")
		}
	}
	sort.Strings(changed)
	if len(changed) > 20 {
		changed = changed[:20]
	}
	data := map[string]any{"hash": hash}
	if len(changed) > 0 {
		data["variants"] = changed
	}
	s.record(ctx, activity.Entry{UserID: userID, Kind: activity.KindConfig, Action: "changed", RefID: hash, Actor: "system", Data: data})
}

func (s *service) configPreview(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID   string   `json:"userId"`
		Platform string   `json:"platform"`
		Version  string   `json:"version"`
		Country  string   `json:"country"`
		TagIDs   []string `json:"tagIds"`
		Lang     string   `json:"lang"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	t := configTarget{Platform: strings.ToLower(in.Platform), Version: strings.TrimSpace(in.Version), Country: strings.ToUpper(strings.TrimSpace(in.Country)), Tags: map[string]bool{}}
	if in.UserID != "" {
		if _, err := uuid.FromString(in.UserID); err != nil {
			s.writeError(w, badRequest("USER_NOT_FOUND", "userId", "player does not exist"))
			return
		}
		var exists bool
		if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, in.UserID).Scan(&exists); err != nil || !exists {
			s.writeError(w, badRequest("USER_NOT_FOUND", "userId", "player does not exist"))
			return
		}
		t.UserID = in.UserID
		tags, err := s.userTagSet(r.Context(), in.UserID)
		if err != nil {
			s.writeError(w, err)
			return
		}
		t.Tags = tags
		if t.Country == "" {
			t.Country = s.userCountry(r.Context(), in.UserID)
		}
	}
	for _, id := range in.TagIDs {
		t.Tags[id] = true
	}
	langs, err := s.languages(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	lang := langs.Choose(in.Lang)
	keys, err := s.activeConfigKeys(r.Context(), nil)
	if err != nil {
		s.writeError(w, err)
		return
	}
	values, variants := s.resolveConfig(r.Context(), keys, t, langs, lang)
	buckets := map[string]int{}
	if t.UserID != "" {
		for _, k := range keys {
			buckets[k.Key] = bucket(k.Key, t.UserID)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"lang": lang, "values": values, "variants": variants, "buckets": buckets, "hash": configHash(values)})
}
