package apierr

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed defaults/codes.json defaults/texts.json
var defaultsFS embed.FS

type Default struct {
	Status     int               `json:"status"`
	Modules    []string          `json:"modules,omitempty"`
	Messages   map[string]string `json:"messages,omitempty"`
	RetryAfter int               `json:"retryAfterSeconds,omitempty"`
}

type Entry struct {
	Code            string            `json:"code"`
	Status          int               `json:"status"`
	DefaultStatus   int               `json:"defaultStatus"`
	Messages        map[string]string `json:"messages"`
	DefaultMessages map[string]string `json:"defaultMessages"`
	RetryAfter      *int              `json:"retryAfterSeconds"`
	Modules         []string          `json:"modules"`
	Custom          bool              `json:"custom"`
	Edited          bool              `json:"edited"`
	UpdatedBy       string            `json:"updatedBy,omitempty"`
	UpdatedAt       string            `json:"updatedAt,omitempty"`
}

var (
	defaultsOnce sync.Once
	defaults     map[string]Default
	registered   = map[string]Default{}
	regMu        sync.RWMutex
)

func loadDefaults() {
	defaults = map[string]Default{}
	var codes map[string]struct {
		Status  int      `json:"status"`
		Modules []string `json:"modules"`
	}
	if raw, err := defaultsFS.ReadFile("defaults/codes.json"); err == nil {
		_ = json.Unmarshal(raw, &codes)
	}
	var texts map[string]map[string]string
	if raw, err := defaultsFS.ReadFile("defaults/texts.json"); err == nil {
		_ = json.Unmarshal(raw, &texts)
	}
	for code, c := range codes {
		defaults[code] = Default{Status: c.Status, Modules: c.Modules, Messages: texts[code]}
	}
	for code, msgs := range texts {
		if _, ok := defaults[code]; !ok {
			defaults[code] = Default{Status: 400, Messages: msgs}
		}
	}
}

func Register(module string, entries map[string]Default) {
	regMu.Lock()
	defer regMu.Unlock()
	for code, d := range entries {
		if !CodePattern.MatchString(code) {
			continue
		}
		if module != "" {
			d.Modules = appendUnique(d.Modules, module)
		}
		registered[code] = d
	}
	Invalidate()
}

func Defaults() map[string]Default {
	defaultsOnce.Do(loadDefaults)
	out := make(map[string]Default, len(defaults)+len(registered))
	for k, v := range defaults {
		out[k] = v
	}
	regMu.RLock()
	for k, v := range registered {
		if base, ok := out[k]; ok {
			if v.Status == 0 {
				v.Status = base.Status
			}
			merged := map[string]string{}
			for l, m := range base.Messages {
				merged[l] = m
			}
			for l, m := range v.Messages {
				merged[l] = m
			}
			v.Messages = merged
			for _, m := range base.Modules {
				v.Modules = appendUnique(v.Modules, m)
			}
		}
		out[k] = v
	}
	regMu.RUnlock()
	return out
}

func appendUnique(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

var (
	catMu      sync.Mutex
	catCache   map[string]Entry
	catVersion string
	catAt      time.Time
	CacheTTL   = 15 * time.Second
)

func Invalidate() {
	catMu.Lock()
	catAt = time.Time{}
	catMu.Unlock()
}

func Catalog(ctx context.Context, db *sql.DB) (map[string]Entry, string, error) {
	catMu.Lock()
	if catCache != nil && !catAt.IsZero() && time.Since(catAt) < CacheTTL {
		c, v := catCache, catVersion
		catMu.Unlock()
		return c, v, nil
	}
	catMu.Unlock()
	entries, version, err := ReadCatalog(ctx, db)
	if err != nil {
		return nil, "", err
	}
	catMu.Lock()
	catCache, catVersion, catAt = entries, version, time.Now()
	catMu.Unlock()
	return entries, version, nil
}

func ReadCatalog(ctx context.Context, db *sql.DB) (map[string]Entry, string, error) {
	out := map[string]Entry{}
	for code, d := range Defaults() {
		msgs := map[string]string{}
		for l, m := range d.Messages {
			msgs[l] = m
		}
		var retry *int
		if d.RetryAfter > 0 {
			r := d.RetryAfter
			retry = &r
		}
		out[code] = Entry{Code: code, Status: d.Status, DefaultStatus: d.Status, Messages: msgs, DefaultMessages: d.Messages, RetryAfter: retry, Modules: d.Modules}
	}
	latest := time.Time{}
	if db != nil {
		rows, err := db.QueryContext(ctx, `SELECT code, messages, status, retry_after_seconds, is_custom, updated_by, updated_at FROM devkit_error_message`)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return out, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		defer rows.Close()
		for rows.Next() {
			var code, by string
			var raw []byte
			var status, retry sql.NullInt64
			var custom bool
			var at time.Time
			if err := rows.Scan(&code, &raw, &status, &retry, &custom, &by, &at); err != nil {
				return nil, "", err
			}
			e, ok := out[code]
			if !ok {
				if !custom {
					continue
				}
				e = Entry{Code: code, Status: 400, DefaultStatus: 0, Messages: map[string]string{}, DefaultMessages: map[string]string{}, Custom: true}
			}
			var msgs map[string]string
			_ = json.Unmarshal(raw, &msgs)
			for l, m := range msgs {
				if strings.TrimSpace(m) != "" {
					e.Messages[l] = m
					e.Edited = true
				}
			}
			if status.Valid {
				e.Status = int(status.Int64)
				e.Edited = true
			}
			if retry.Valid {
				r := int(retry.Int64)
				e.RetryAfter = &r
				e.Edited = true
			}
			e.UpdatedBy = by
			e.UpdatedAt = at.UTC().Format(time.RFC3339)
			if at.After(latest) {
				latest = at
			}
			out[code] = e
		}
		if err := rows.Err(); err != nil {
			return nil, "", err
		}
	}
	stamp := int64(0)
	if !latest.IsZero() {
		stamp = latest.UnixMilli()
	}
	return out, fmt.Sprintf("%d-%d", stamp, len(out)), nil
}

func SortedCodes(entries map[string]Entry) []string {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (e Entry) MessageFor(lang, fallback string) string {
	for _, l := range []string{lang, baseLang(lang), fallback, baseLang(fallback), "en"} {
		if l == "" {
			continue
		}
		if m := strings.TrimSpace(e.Messages[l]); m != "" {
			return m
		}
	}
	return ""
}

func baseLang(l string) string {
	if i := strings.IndexByte(l, '-'); i > 0 {
		return l[:i]
	}
	return l
}

var placeholder = regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]*)\}`)

func Fill(text string, meta map[string]any, localDigits bool) string {
	if text == "" || len(meta) == 0 {
		return text
	}
	return placeholder.ReplaceAllStringFunc(text, func(m string) string {
		key := m[1 : len(m)-1]
		v, ok := meta[key]
		if !ok || v == nil {
			return m
		}
		s := fmt.Sprint(v)
		if f, ok := v.(float64); ok && f == float64(int64(f)) {
			s = fmt.Sprint(int64(f))
		}
		if localDigits && isNumber(v) {
			s = PersianDigits(s)
		}
		return s
	})
}

func isNumber(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		return true
	}
	return false
}

func PersianDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune('۰' + (r - '0'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func UsesPersianDigits(lang string) bool {
	l := baseLang(strings.ToLower(lang))
	return l == "fa" || l == "ckb" || l == "ps"
}
