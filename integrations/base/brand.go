package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var (
	accentPattern      = regexp.MustCompile(`^#?([0-9a-f]{3}|[0-9a-f]{6})$`)
	consoleUserHeader  = "X-Nakama-Console-Username"
	brandPublicMaxAge  = "public, max-age=60, stale-while-revalidate=300"
	brandShortNameMax  = 24
	brandNameMax       = 64
	brandWelcomeMax    = 280
	brandSelectColumns = `name::text, short_name, logo_url, logo_dark_url, mark_url, accent, welcome::text, updated_by, updated_at`
)

type brand struct {
	Name        localized.Text `json:"name"`
	ShortName   string         `json:"shortName"`
	LogoURL     string         `json:"logoUrl"`
	LogoDarkURL string         `json:"logoDarkUrl"`
	MarkURL     string         `json:"markUrl"`
	Accent      string         `json:"accent"`
	Welcome     localized.Text `json:"welcome"`
	UpdatedBy   string         `json:"updatedBy,omitempty"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

func (b brand) custom() bool {
	return len(b.Name) > 0 || b.ShortName != "" || b.LogoURL != "" || b.LogoDarkURL != "" || b.MarkURL != "" || b.Accent != "" || len(b.Welcome) > 0
}

func (b brand) public() map[string]any {
	return map[string]any{
		"name": nonNil(b.Name), "shortName": b.ShortName, "logoUrl": b.LogoURL, "logoDarkUrl": b.LogoDarkURL, "markUrl": b.MarkURL,
		"accent": b.Accent, "welcome": nonNil(b.Welcome), "custom": b.custom(), "updatedAt": b.UpdatedAt,
	}
}

func (b brand) admin() map[string]any {
	out := b.public()
	out["updatedBy"] = b.UpdatedBy
	return out
}

func nonNil(t localized.Text) localized.Text {
	if t == nil {
		return localized.Text{}
	}
	return t
}

func loadBrand(ctx context.Context, q queryer) (brand, error) {
	var b brand
	var name, welcome string
	err := q.QueryRowContext(ctx, `SELECT `+brandSelectColumns+` FROM base_brand WHERE id=1`).Scan(&name, &b.ShortName, &b.LogoURL, &b.LogoDarkURL, &b.MarkURL, &b.Accent, &welcome, &b.UpdatedBy, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return brand{Name: localized.Text{}, Welcome: localized.Text{}}, nil
	}
	if err != nil {
		return b, err
	}
	b.Name = localized.Parse(name)
	b.Welcome = localized.Parse(welcome)
	return b, nil
}

func cleanAccent(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "", nil
	}
	m := accentPattern.FindStringSubmatch(v)
	if m == nil {
		return "", badRequest("INVALID_COLOR", "accent", "use a hex color like #6b59ee")
	}
	hexv := m[1]
	if len(hexv) == 3 {
		hexv = string([]byte{hexv[0], hexv[0], hexv[1], hexv[1], hexv[2], hexv[2]})
	}
	return "#" + hexv, nil
}

type brandInput struct {
	Name        localized.Text `json:"name"`
	ShortName   string         `json:"shortName"`
	LogoURL     string         `json:"logoUrl"`
	LogoDarkURL string         `json:"logoDarkUrl"`
	MarkURL     string         `json:"markUrl"`
	Accent      string         `json:"accent"`
	Welcome     localized.Text `json:"welcome"`
}

func (s *service) cleanBrand(ctx context.Context, in brandInput) (brand, error) {
	langs, err := s.languages(ctx, s.db)
	if err != nil {
		return brand{}, err
	}
	var out brand
	if out.Name, err = cleanText(langs, "name", in.Name, false, brandNameMax); err != nil {
		return out, err
	}
	if out.Welcome, err = cleanText(langs, "welcome", in.Welcome, false, brandWelcomeMax); err != nil {
		return out, err
	}
	out.ShortName = strings.TrimSpace(in.ShortName)
	if utf8.RuneCountInString(out.ShortName) > brandShortNameMax {
		return out, badRequest("TEXT_TOO_LONG", "shortName", "short name is at most 24 characters")
	}
	if out.LogoURL, err = cleanURL("logoUrl", in.LogoURL); err != nil {
		return out, err
	}
	if out.LogoDarkURL, err = cleanURL("logoDarkUrl", in.LogoDarkURL); err != nil {
		return out, err
	}
	if out.MarkURL, err = cleanURL("markUrl", in.MarkURL); err != nil {
		return out, err
	}
	if out.Accent, err = cleanAccent(in.Accent); err != nil {
		return out, err
	}
	return out, nil
}

func (s *service) brandGet(w http.ResponseWriter, r *http.Request) {
	b, err := loadBrand(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b.admin())
}

func (s *service) brandPut(w http.ResponseWriter, r *http.Request) {
	var in brandInput
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	b, err := s.cleanBrand(r.Context(), in)
	if err != nil {
		s.writeError(w, err)
		return
	}
	by := consoleUser(r)
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO base_brand (id, name, short_name, logo_url, logo_dark_url, mark_url, accent, welcome, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,now())
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, short_name=EXCLUDED.short_name, logo_url=EXCLUDED.logo_url, logo_dark_url=EXCLUDED.logo_dark_url,
		mark_url=EXCLUDED.mark_url, accent=EXCLUDED.accent, welcome=EXCLUDED.welcome, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		b.Name.JSON(), b.ShortName, b.LogoURL, b.LogoDarkURL, b.MarkURL, b.Accent, b.Welcome.JSON(), by)
	if err != nil {
		s.writeError(w, err)
		return
	}
	saved, err := loadBrand(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, saved.admin())
}

func (s *service) brandPublic(w http.ResponseWriter, r *http.Request) {
	b, err := loadBrand(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	body, err := json.Marshal(b.public())
	if err != nil {
		s.writeError(w, err)
		return
	}
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:8]) + `"`
	w.Header().Set("Cache-Control", brandPublicMaxAge)
	w.Header().Set("ETag", etag)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if match := r.Header.Get("If-None-Match"); match != "" && strings.Contains(match, etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

func consoleUser(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(consoleUserHeader))
}

func (s *service) rpcBrand(ctx context.Context, userID, payload string) (any, error) {
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
	b, err := loadBrand(ctx, s.db)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"lang": lang, "custom": b.custom(), "name": langs.Resolve(b.Name, lang), "shortName": b.ShortName,
		"logoUrl": b.LogoURL, "logoDarkUrl": b.LogoDarkURL, "markUrl": b.MarkURL, "accent": b.Accent,
		"welcome": langs.Resolve(b.Welcome, lang), "updatedAt": b.UpdatedAt,
	}, nil
}
