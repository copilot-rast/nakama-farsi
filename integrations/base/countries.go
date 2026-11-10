package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var (
	countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)
	dialCodePattern    = regexp.MustCompile(`^\+[1-9][0-9]{0,4}$`)
)

func flagEmoji(code string) string {
	if !countryCodePattern.MatchString(code) {
		return ""
	}
	return string([]rune{rune(code[0]) - 'A' + 0x1F1E6, rune(code[1]) - 'A' + 0x1F1E6})
}

func cleanCountryCode(v string) string {
	return strings.ToUpper(strings.TrimSpace(v))
}

const countryJSON = `jsonb_build_object('id',c.code,'code',c.code,'name',c.name,'dialCode',c.dial_code,'flagUrl',c.flag_url,'isActive',c.is_active,'sort',c.sort,'createdAt',c.created_at,'updatedAt',c.updated_at)`

type countryInput struct {
	Code     string         `json:"code"`
	Name     localized.Text `json:"name"`
	DialCode string         `json:"dialCode"`
	FlagURL  string         `json:"flagUrl"`
	IsActive *bool          `json:"isActive"`
	Sort     *int           `json:"sort"`
}

func withFlag(raw string) json.RawMessage {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return json.RawMessage(raw)
	}
	code, _ := m["code"].(string)
	m["flag"] = flagEmoji(code)
	out, _ := json.Marshal(m)
	return out
}

func (s *service) countryList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(c.code ILIKE "+n+" OR c.dial_code ILIKE "+n+" OR "+translatedSearch("c.name", n)+")")
	}
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add("c.is_active = ?", v == "true")
	}
	out, err := s.listRaw(r, ` FROM base_country c`, countryJSON, "c.sort, c.name->>'en', c.code", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if items, ok := out["items"].([]json.RawMessage); ok {
		for i, it := range items {
			items[i] = withFlag(string(it))
		}
	}
	var active int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_country WHERE is_active`).Scan(&active); err != nil {
		s.writeError(w, err)
		return
	}
	out["activeTotal"] = active
	writeJSON(w, http.StatusOK, out)
}

func fetchCountry(ctx context.Context, q queryer, code string) (json.RawMessage, error) {
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+countryJSON+`::text FROM base_country c WHERE c.code=$1`, code).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("COUNTRY_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return withFlag(raw), nil
}

func (s *service) countryGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchCountry(r.Context(), s.db, cleanCountryCode(pathVar(r, "id")))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func (s *service) countryWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in countryInput
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		code := cleanCountryCode(in.Code)
		if !create {
			code = cleanCountryCode(pathVar(r, "id"))
		}
		if !countryCodePattern.MatchString(code) {
			s.writeError(w, badRequest("INVALID_COUNTRY_CODE", "code", "use the two-letter ISO code, like IR"))
			return
		}
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			langs, err := s.languages(r.Context(), tx)
			if err != nil {
				return err
			}
			known := localized.Text{}
			for k, v := range in.Name {
				if langs.Known(localized.Normalize(k)) {
					known[k] = v
				}
			}
			name, err := cleanText(langs, "name", known, true, 100)
			if err != nil {
				return err
			}
			dial := strings.ReplaceAll(strings.TrimSpace(in.DialCode), " ", "")
			if dial != "" && !strings.HasPrefix(dial, "+") {
				dial = "+" + dial
			}
			if dial != "" && !dialCodePattern.MatchString(dial) {
				return badRequest("INVALID_DIAL_CODE", "dialCode", "use + and up to five digits, like +98")
			}
			flag, err := cleanURL("flagUrl", in.FlagURL)
			if err != nil {
				return err
			}
			active := in.IsActive == nil || *in.IsActive
			sortN := 100
			if in.Sort != nil {
				sortN = *in.Sort
			}
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_country (code, name, dial_code, flag_url, is_active, sort) VALUES ($1,$2,$3,$4,$5,$6)`, code, name.JSON(), dial, flag, active, sortN)
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_country SET name=$2, dial_code=$3, flag_url=$4, is_active=$5, sort=$6, updated_at=now() WHERE code=$1`, code, name.JSON(), dial, flag, active, sortN)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("COUNTRY_NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchCountry(r.Context(), tx, code)
			return err
		})
		if err != nil {
			if _, ok := isUniqueViolation(err); ok {
				err = fail(http.StatusConflict, "ALREADY_EXISTS", "code", "this country already exists")
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

func (s *service) countryDelete(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_country WHERE code=$1`, cleanCountryCode(pathVar(r, "id")))
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("COUNTRY_NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) rpcCountries(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT code, name::text, dial_code, flag_url, sort FROM base_country WHERE is_active`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type item struct {
		Code     string `json:"code"`
		Name     string `json:"name"`
		DialCode string `json:"dialCode"`
		Flag     string `json:"flag"`
		FlagURL  string `json:"flagUrl"`
		sort     int
	}
	items := []item{}
	for rows.Next() {
		var it item
		var name string
		if err := rows.Scan(&it.Code, &name, &it.DialCode, &it.FlagURL, &it.sort); err != nil {
			return nil, err
		}
		it.Name = langs.Resolve(localized.Parse(name), lang)
		if it.Name == "" {
			it.Name = it.Code
		}
		it.Flag = flagEmoji(it.Code)
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].sort != items[j].sort {
			return items[i].sort < items[j].sort
		}
		return items[i].Name < items[j].Name
	})
	var selected string
	_ = s.db.QueryRowContext(ctx, `SELECT coalesce(metadata->>'country','') FROM users WHERE id=$1`, userID).Scan(&selected)
	return map[string]any{"lang": lang, "selected": selected, "items": items}, nil
}

func (s *service) rpcSetCountry(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Code string `json:"code"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	code := cleanCountryCode(in.Code)
	var active bool
	err := s.db.QueryRowContext(ctx, `SELECT is_active FROM base_country WHERE code=$1`, code).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !active) {
		return nil, badRequest("COUNTRY_NOT_AVAILABLE", "code", "this country is not available")
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE users SET metadata = jsonb_set(coalesce(metadata,'{}'::jsonb), '{country}', to_jsonb($2::text)), update_time=now() WHERE id=$1`, userID, code); err != nil {
		return nil, err
	}
	return map[string]any{"code": code}, nil
}
