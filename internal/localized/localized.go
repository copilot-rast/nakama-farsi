package localized

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Text map[string]string

type Language struct {
	Code      string
	Direction string
	Active    bool
	Default   bool
}

type Set struct {
	Default string
	Active  []string
	byCode  map[string]Language
}

type Error struct {
	Code    string
	Field   string
	Message string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type RowQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var codePattern = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)

func ValidCode(code string) bool { return codePattern.MatchString(code) }

func Normalize(code string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(code)), "_", "-")
}

func NewSet(langs []Language) *Set {
	s := &Set{byCode: map[string]Language{}}
	for _, l := range langs {
		s.byCode[l.Code] = l
		if l.Active {
			s.Active = append(s.Active, l.Code)
		}
		if l.Default {
			s.Default = l.Code
		}
	}
	if s.Default == "" && len(s.Active) > 0 {
		s.Default = s.Active[0]
	}
	return s
}

func Load(ctx context.Context, q Queryer) (*Set, error) {
	rows, err := q.QueryContext(ctx, `SELECT code, direction, is_active, is_default FROM base_language ORDER BY is_default DESC, sort, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var langs []Language
	for rows.Next() {
		var l Language
		if err := rows.Scan(&l.Code, &l.Direction, &l.Active, &l.Default); err != nil {
			return nil, err
		}
		langs = append(langs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return NewSet(langs), nil
}

func (s *Set) Known(code string) bool {
	_, ok := s.byCode[code]
	return ok
}

func (s *Set) IsActive(code string) bool {
	l, ok := s.byCode[code]
	return ok && l.Active
}

func (s *Set) Direction(code string) string {
	if l, ok := s.byCode[code]; ok && l.Direction != "" {
		return l.Direction
	}
	return "ltr"
}

func (s *Set) Add(l Language) {
	if _, ok := s.byCode[l.Code]; !ok && l.Active {
		s.Active = append(s.Active, l.Code)
	}
	s.byCode[l.Code] = l
}

func (s *Set) Match(code string) string {
	code = Normalize(code)
	if code == "" {
		return ""
	}
	if s.IsActive(code) {
		return code
	}
	if i := strings.IndexByte(code, '-'); i > 0 && s.IsActive(code[:i]) {
		return code[:i]
	}
	return ""
}

func (s *Set) Choose(candidates ...string) string {
	for _, c := range candidates {
		if m := s.Match(c); m != "" {
			return m
		}
	}
	return s.Default
}

func (s *Set) Resolve(t Text, lang string) string {
	if v := t[lang]; v != "" {
		return v
	}
	if v := t[s.Default]; v != "" {
		return v
	}
	for _, code := range s.Active {
		if v := t[code]; v != "" {
			return v
		}
	}
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if t[k] != "" {
			return t[k]
		}
	}
	return ""
}

func (s *Set) Clean(field string, in Text, required bool, max int) (Text, error) {
	out := Text{}
	for k, v := range in {
		code := Normalize(k)
		if !s.Known(code) {
			return nil, &Error{Code: "INVALID_LANGUAGE", Field: field, Message: "unknown language " + k}
		}
		v = strings.TrimSpace(v)
		if !utf8.ValidString(v) || utf8.RuneCountInString(v) > max {
			return nil, &Error{Code: "TEXT_TOO_LONG", Field: field, Message: "text is too long (max " + strconv.Itoa(max) + ")"}
		}
		if v != "" {
			out[code] = v
		}
	}
	if (required || len(out) > 0) && out[s.Default] == "" {
		return nil, &Error{Code: "DEFAULT_LANGUAGE_REQUIRED", Field: field, Message: "a value in the default language (" + s.Default + ") is required"}
	}
	return out, nil
}

func (s *Set) Missing(t Text) []string {
	out := []string{}
	if len(t) == 0 {
		return out
	}
	for _, code := range s.Active {
		if t[code] == "" {
			out = append(out, code)
		}
	}
	return out
}

func Parse(raw string) Text {
	out := Text{}
	if raw == "" {
		return out
	}
	var values map[string]any
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return out
	}
	for k, v := range values {
		if str, ok := v.(string); ok {
			out[k] = str
		}
	}
	return out
}

func (t Text) JSON() string {
	if t == nil {
		t = Text{}
	}
	b, _ := json.Marshal(map[string]string(t))
	return string(b)
}

func UserLangTag(ctx context.Context, q RowQueryer, userID string) string {
	var tag sql.NullString
	if err := q.QueryRowContext(ctx, `SELECT lang_tag FROM users WHERE id=$1`, userID).Scan(&tag); err != nil {
		return ""
	}
	return tag.String
}

type DB interface {
	Queryer
	RowQueryer
}

func ForUser(ctx context.Context, db DB, userID, requested string) (*Set, string, error) {
	set, err := Load(ctx, db)
	if err != nil {
		return nil, "", err
	}
	if m := set.Match(requested); m != "" {
		return set, m, nil
	}
	return set, set.Choose(UserLangTag(ctx, db, userID)), nil
}
