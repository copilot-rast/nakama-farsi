package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const languageJSON = `jsonb_build_object('id',t.code,'code',t.code,'name',t.name,'nativeName',t.native_name,'direction',t.direction,'iconUrl',t.icon_url,
	'isActive',t.is_active,'isDefault',t.is_default,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at)`

type languageInput struct {
	Code       string         `json:"code"`
	Name       localized.Text `json:"name"`
	NativeName string         `json:"nativeName"`
	Direction  string         `json:"direction"`
	IconURL    string         `json:"iconUrl"`
	IsActive   *bool          `json:"isActive"`
	IsDefault  bool           `json:"isDefault"`
	Sort       int            `json:"sort"`
}

func (s *service) languageList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(t.code ILIKE "+n+" OR t.native_name ILIKE "+n+" OR "+translatedSearch("t.name", n)+")")
	}
	if v := strings.TrimSpace(r.URL.Query().Get("isActive")); v != "" {
		if v != "true" && v != "false" {
			s.writeError(w, badRequest("INVALID_FILTER", "isActive", "must be true or false"))
			return
		}
		where.add("t.is_active = ?::boolean", v)
	}
	out, err := s.listRaw(r, ` FROM base_language t`, languageJSON, "t.is_default DESC, t.sort, t.code", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchLanguage(ctx context.Context, q queryer, code string) (json.RawMessage, error) {
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+languageJSON+`::text FROM base_language t WHERE t.code=$1`, localized.Normalize(code)).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("LANGUAGE_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) languageGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchLanguage(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func (s *service) languageWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in languageInput
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		code := localized.Normalize(in.Code)
		if !create {
			code = localized.Normalize(pathVar(r, "id"))
		}
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			if err := s.saveLanguage(r.Context(), tx, code, create, in); err != nil {
				return err
			}
			var err error
			out, err = fetchLanguage(r.Context(), tx, code)
			return err
		})
		if err != nil {
			if _, ok := isUniqueViolation(err); ok {
				err = fail(http.StatusConflict, "ALREADY_EXISTS", "code", "this language already exists")
			}
			s.writeError(w, err)
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeRaw(w, status, out)
	}
}

func (s *service) saveLanguage(ctx context.Context, tx *sql.Tx, code string, create bool, in languageInput) error {
	if !localized.ValidCode(code) {
		if !create {
			return notFound("LANGUAGE_NOT_FOUND")
		}
		return badRequest("INVALID_LANGUAGE_CODE", "code", "use a lower-case language code like fa, en, ar or pt-br")
	}
	if _, err := tx.ExecContext(ctx, `SELECT code FROM base_language FOR UPDATE`); err != nil {
		return err
	}
	var current struct {
		exists, active, isDefault bool
	}
	err := tx.QueryRowContext(ctx, `SELECT true, is_active, is_default FROM base_language WHERE code=$1`, code).Scan(&current.exists, &current.active, &current.isDefault)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if create && current.exists {
		return fail(http.StatusConflict, "ALREADY_EXISTS", "code", "this language already exists")
	}
	if !create && !current.exists {
		return notFound("LANGUAGE_NOT_FOUND")
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	} else if !create {
		active = current.active
	}
	if current.isDefault && !active {
		return conflict("DEFAULT_LANGUAGE_LOCKED", "the default language cannot be deactivated; choose another default first")
	}
	if current.isDefault && !in.IsDefault {
		return badRequest("DEFAULT_LANGUAGE_REQUIRED", "isDefault", "one language must stay the default; make another language the default instead")
	}
	if in.IsDefault && !active {
		return badRequest("DEFAULT_MUST_BE_ACTIVE", "isActive", "the default language must be active")
	}
	native := strings.TrimSpace(in.NativeName)
	if native == "" || utf8.RuneCountInString(native) > 64 {
		return badRequest("INVALID_NATIVE_NAME", "nativeName", "native name is required (max 64)")
	}
	direction := strings.TrimSpace(in.Direction)
	if direction == "" {
		direction = "ltr"
	}
	if direction != "rtl" && direction != "ltr" {
		return badRequest("INVALID_DIRECTION", "direction", "direction must be rtl or ltr")
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	langs, err := s.languages(ctx, tx)
	if err != nil {
		return err
	}
	langs.Add(localized.Language{Code: code, Direction: direction, Active: active})
	name, err := cleanText(langs, "name", in.Name, true, 64)
	if err != nil {
		return err
	}
	if in.IsDefault && !current.isDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE base_language SET is_default=false, updated_at=now() WHERE is_default`); err != nil {
			return err
		}
	}
	if create {
		_, err = tx.ExecContext(ctx, `INSERT INTO base_language (code, name, native_name, direction, icon_url, is_active, is_default, sort) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			code, name.JSON(), native, direction, icon, active, in.IsDefault, in.Sort)
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE base_language SET name=$2, native_name=$3, direction=$4, icon_url=$5, is_active=$6, is_default=$7, sort=$8, updated_at=now() WHERE code=$1`,
		code, name.JSON(), native, direction, icon, active, in.IsDefault, in.Sort)
	return err
}

func (s *service) languageDelete(w http.ResponseWriter, r *http.Request) {
	code := localized.Normalize(pathVar(r, "id"))
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var active, isDefault bool
		err := tx.QueryRowContext(r.Context(), `SELECT is_active, is_default FROM base_language WHERE code=$1 FOR UPDATE`, code).Scan(&active, &isDefault)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("LANGUAGE_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		if isDefault {
			return conflict("DEFAULT_LANGUAGE_LOCKED", "the default language cannot be deleted")
		}
		if active {
			return conflict("LANGUAGE_ACTIVE", "deactivate the language before deleting it")
		}
		_, err = tx.ExecContext(r.Context(), `DELETE FROM base_language WHERE code=$1`, code)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

type coverageTarget struct {
	Module   string
	Resource string
	Table    string
	Fields   []string
}

var coverageTargets = []coverageTarget{
	{"base", "languages", "base_language", []string{"name"}},
	{"base", "categories", "base_category", []string{"name", "description"}},
	{"base", "tags", "base_tag", []string{"name", "description"}},
	{"economy", "currencies", "economy_currency", []string{"name"}},
	{"economy", "metas", "economy_meta", []string{"name"}},
	{"economy", "prizes", "economy_prize", []string{"name", "message"}},
	{"economy", "products", "economy_product", []string{"name", "description"}},
	{"economy", "gifts", "economy_gift", []string{"name"}},
	{"publication", "faqs", "publication_faq", []string{"question", "answer"}},
	{"publication", "news", "publication_news", []string{"title", "summary", "body"}},
}

func (s *service) languageCoverage(w http.ResponseWriter, r *http.Request) {
	out, err := s.coverage(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) coverage(ctx context.Context) (map[string]any, error) {
	langs, err := s.languages(ctx, s.db)
	if err != nil {
		return nil, err
	}
	totals := map[string]int64{}
	for _, code := range langs.Active {
		totals[code] = 0
	}
	items := []map[string]any{}
	for _, target := range coverageTargets {
		var exists bool
		if err := s.db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, target.Table).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		cols := []string{"count(*)"}
		args := []any{}
		for i, code := range langs.Active {
			args = append(args, code)
			conds := make([]string, 0, len(target.Fields))
			for _, f := range target.Fields {
				conds = append(conds, "("+f+" <> '{}'::jsonb AND COALESCE("+f+"->>$"+itoa(i+1)+",'') = '')")
			}
			cols = append(cols, "count(*) FILTER (WHERE "+strings.Join(conds, " OR ")+")")
		}
		dest := make([]any, len(cols))
		counts := make([]int64, len(cols))
		for i := range counts {
			dest[i] = &counts[i]
		}
		if err := s.db.QueryRowContext(ctx, `SELECT `+strings.Join(cols, ", ")+` FROM `+target.Table, args...).Scan(dest...); err != nil {
			return nil, err
		}
		missing := map[string]int64{}
		for i, code := range langs.Active {
			missing[code] = counts[i+1]
			totals[code] += counts[i+1]
		}
		items = append(items, map[string]any{"module": target.Module, "resource": target.Resource, "fields": target.Fields, "total": counts[0], "missing": missing})
	}
	return map[string]any{"default": langs.Default, "languages": langs.Active, "missing": totals, "items": items}, nil
}
