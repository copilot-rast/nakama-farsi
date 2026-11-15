package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const publicPrefix = "/v2/publication/public/"

var slugPattern = regexp.MustCompile(`^[\p{L}\p{N}]+(-[\p{L}\p{N}]+)*$`)

var persianLatin = map[rune]string{
	'ا': "a", 'آ': "a", 'أ': "a", 'إ': "e", 'ب': "b", 'پ': "p", 'ت': "t", 'ث': "s", 'ج': "j", 'چ': "ch", 'ح': "h", 'خ': "kh", 'د': "d", 'ذ': "z", 'ر': "r", 'ز': "z",
	'ژ': "zh", 'س': "s", 'ش': "sh", 'ص': "s", 'ض': "z", 'ط': "t", 'ظ': "z", 'ع': "a", 'غ': "gh", 'ف': "f", 'ق': "gh", 'ک': "k", 'ك': "k", 'گ': "g", 'ل': "l",
	'م': "m", 'ن': "n", 'و': "v", 'ه': "h", 'ة': "h", 'ی': "i", 'ي': "i", 'ى': "i", 'ئ': "i", 'ؤ': "o", 'ء': "",
}

func normalizeSlugRune(r rune) rune {
	switch {
	case r >= '۰' && r <= '۹':
		return '0' + (r - '۰')
	case r >= '٠' && r <= '٩':
		return '0' + (r - '٠')
	case r == 'ي' || r == 'ى':
		return 'ی'
	case r == 'ك':
		return 'ک'
	}
	return unicode.ToLower(r)
}

func slugify(title, style string) string {
	if style == "id" {
		return ""
	}
	var b strings.Builder
	dash := false
	put := func(s string) {
		if s == "" {
			return
		}
		b.WriteString(s)
		dash = false
	}
	sep := func() {
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	for _, r := range title {
		r = normalizeSlugRune(r)
		if style == "latin" {
			if r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
				put(string(r))
			} else if lat, ok := persianLatin[r]; ok {
				put(lat)
			} else if unicode.Is(unicode.Mn, r) {
				continue
			} else {
				sep()
			}
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			put(string(r))
		} else if unicode.Is(unicode.Mn, r) {
			continue
		} else {
			sep()
		}
	}
	out := strings.Trim(b.String(), "-")
	if utf8.RuneCountInString(out) > 80 {
		rs := []rune(out)[:80]
		out = string(rs)
		if i := strings.LastIndex(out, "-"); i > 20 {
			out = out[:i]
		}
		out = strings.Trim(out, "-")
	}
	return out
}

func cleanSlug(raw string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		b.WriteRune(normalizeSlugRune(r))
	}
	s := b.String()
	if s == "" {
		return "", nil
	}
	if utf8.RuneCountInString(s) > 120 || !slugPattern.MatchString(s) {
		return "", badRequest("INVALID_SLUG", "slug", "use letters and digits joined by - (max 120)")
	}
	return s, nil
}

func uniqueSlug(ctx context.Context, q queryer, base, id string) (string, error) {
	if base == "" {
		base = strings.SplitN(id, "-", 2)[0]
	}
	for i := 1; i <= 100; i++ {
		cand := base
		if i > 1 {
			cand = base + "-" + strconv.Itoa(i)
		}
		var taken bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_news WHERE slug=$1 AND id<>$2)`, cand, id).Scan(&taken); err != nil {
			return "", err
		}
		if !taken {
			return cand, nil
		}
	}
	return base + "-" + strings.SplitN(id, "-", 2)[0], nil
}

func slugTaken(ctx context.Context, q queryer, slug, id string) (bool, error) {
	var taken bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_news WHERE slug=$1 AND id<>$2)`, slug, id).Scan(&taken)
	return taken, err
}

func cleanKeywords(in []string) ([]string, error) {
	out := []string{}
	for _, k := range in {
		k = strings.Join(strings.Fields(k), " ")
		if k == "" || containsStr(out, k) {
			continue
		}
		if utf8.RuneCountInString(k) > 50 || strings.ContainsAny(k, "<>\"") {
			return nil, badRequest("INVALID_KEYWORD", "keywords", "keywords are at most 50 characters")
		}
		out = append(out, k)
	}
	if len(out) > 20 {
		return nil, badRequest("TOO_MANY_KEYWORDS", "keywords", "at most 20 keywords")
	}
	return out, nil
}

func cleanAbsURL(field, v string) (string, error) {
	v, err := cleanURL(field, v)
	if err != nil {
		return "", err
	}
	if v != "" && !strings.HasPrefix(strings.ToLower(v), "http://") && !strings.HasPrefix(strings.ToLower(v), "https://") {
		return "", badRequest("INVALID_URL", field, "use an absolute http(s) URL")
	}
	return v, nil
}

type seoRow struct {
	Slug, Canonical, OgImage string
	MetaTitle, MetaDesc      i18n
	Keywords                 []string
	Noindex                  bool
}

func seoView(langs *localized.Set, lang string, r seoRow, title, summary, bodyText, cover string) map[string]any {
	mt := langs.Resolve(r.MetaTitle, lang)
	if mt == "" {
		mt = title
	}
	md := langs.Resolve(r.MetaDesc, lang)
	if md == "" {
		md = summary
	}
	if md == "" {
		md = truncRunes(strings.Join(strings.Fields(bodyText), " "), 160)
	}
	og := r.OgImage
	if og == "" {
		og = cover
	}
	kw := r.Keywords
	if kw == nil {
		kw = []string{}
	}
	return map[string]any{"slug": r.Slug, "metaTitle": mt, "metaDescription": md, "keywords": kw, "canonicalUrl": nullable(r.Canonical), "ogImage": nullable(og), "noindex": r.Noindex}
}

func (st newsSettings) publicURL(slug string) string {
	path := st.PublicPath
	if path == "" {
		path = "/news/{slug}"
	}
	return strings.TrimRight(st.PublicBaseURL, "/") + strings.ReplaceAll(path, "{slug}", slug)
}

const publicVisibleSQL = newsLiveSQL + ` AND ` + newsNotExpiredSQL + ` AND NOT n.noindex`

func (s *service) publicGate(w http.ResponseWriter, r *http.Request) (newsSettings, bool) {
	st, err := loadNewsSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return st, false
	}
	if !st.PublicEnabled {
		http.NotFound(w, r)
		return st, false
	}
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(st.PublicCacheSeconds))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	return st, true
}

type publicItem struct {
	row     newsRow
	seo     seoRow
	updated time.Time
	cat     string
}

const publicCols = `n.id::text, n.title::text, n.summary::text, n.body::text, n.body_format, n.cover_url, COALESCE(c.key,''), n.published_at, n.expires_at, n.updated_at,
	n.slug, n.meta_title::text, n.meta_description::text, to_jsonb(n.keywords)::text, n.canonical_url, n.og_image, n.noindex, to_jsonb(n.tag_ids)::text, n.source_url`

func scanPublic(row interface{ Scan(...any) error }) (publicItem, error) {
	var p publicItem
	var exp sql.NullTime
	var mt, md, kw, tags string
	err := row.Scan(&p.row.ID, &p.row.Title, &p.row.Summary, &p.row.Body, &p.row.Format, &p.row.Cover, &p.row.CategoryKey, &p.row.Published, &exp, &p.updated,
		&p.seo.Slug, &mt, &md, &kw, &p.seo.Canonical, &p.seo.OgImage, &p.seo.Noindex, &tags, &p.row.SourceURL)
	if err != nil {
		return p, err
	}
	if exp.Valid {
		t := exp.Time
		p.row.Expires = &t
	}
	p.seo.MetaTitle, p.seo.MetaDesc = parseI18n(mt), parseI18n(md)
	_ = jsonUnmarshalStrings(kw, &p.seo.Keywords)
	_ = jsonUnmarshalStrings(tags, &p.row.Tags)
	return p, nil
}

func (s *service) publicView(langs *localized.Set, lang string, st newsSettings, p publicItem, full bool) map[string]any {
	title := langs.Resolve(parseI18n(p.row.Title), lang)
	summary := langs.Resolve(parseI18n(p.row.Summary), lang)
	body := langs.Resolve(parseI18n(p.row.Body), lang)
	out := map[string]any{"id": p.row.ID, "slug": p.seo.Slug, "title": title, "summary": summary, "coverUrl": nullable(p.row.Cover), "categoryKey": nullable(p.row.CategoryKey),
		"publishedAt": p.row.Published.UTC(), "updatedAt": p.updated.UTC(), "expiresAt": p.row.Expires, "seo": seoView(langs, lang, p.seo, title, summary, newsBodyText(p.row.Format, body), p.row.Cover)}
	if st.PublicBaseURL != "" {
		out["url"] = st.publicURL(p.seo.Slug)
	}
	if full {
		out["bodyHtml"] = newsBodyHTML(p.row.Format, body)
		out["sourceUrl"] = nullable(p.row.SourceURL)
	}
	return out
}

func (s *service) publicNewsList(w http.ResponseWriter, r *http.Request) {
	st, ok := s.publicGate(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	q := r.URL.Query()
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	lang := langs.Choose(q.Get("lang"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	limit = clampLimit(limit, 20, 50)
	offset, err := readCursor(q.Get("cursor"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	where := &sqlWhere{}
	where.add(publicVisibleSQL)
	f := newsFilter{CategoryKey: q.Get("categoryKey"), TagKey: q.Get("tagKey")}
	if err := f.apply(where); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(ctx, `SELECT `+publicCols+` FROM publication_news n LEFT JOIN base_category c ON c.id=n.category_id`+where.sql()+`
		ORDER BY n.is_pinned DESC, n.published_at DESC, n.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2), append(where.args, limit, offset)...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		p, err := scanPublic(rows)
		if err != nil {
			s.writeError(w, err)
			return
		}
		items = append(items, s.publicView(langs, lang, st, p, false))
	}
	writePublic(w, map[string]any{"lang": lang, "dir": langs.Direction(lang), "items": items, "cursor": nextCursor(offset, limit, len(items))})
}

func (s *service) publicNewsGet(w http.ResponseWriter, r *http.Request) {
	st, ok := s.publicGate(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	lang := langs.Choose(r.URL.Query().Get("lang"))
	raw := pathVar(r, "slug")
	if u, err := url.PathUnescape(raw); err == nil {
		raw = u
	}
	slug, _ := cleanSlug(raw)
	if slug == "" {
		s.writeError(w, notFound("NEWS_NOT_FOUND"))
		return
	}
	p, err := scanPublic(s.db.QueryRowContext(ctx, `SELECT `+publicCols+` FROM publication_news n LEFT JOIN base_category c ON c.id=n.category_id WHERE n.slug=$1 AND `+publicVisibleSQL, slug))
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("NEWS_NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := s.publicView(langs, lang, st, p, true)
	out["lang"], out["dir"] = lang, langs.Direction(lang)
	writePublic(w, out)
}

func (s *service) publicFeedRows(ctx context.Context, limit int) ([]publicItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+publicCols+` FROM publication_news n LEFT JOIN base_category c ON c.id=n.category_id WHERE `+publicVisibleSQL+`
		ORDER BY n.published_at DESC, n.id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []publicItem
	for rows.Next() {
		p, err := scanPublic(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *service) publicSitemap(w http.ResponseWriter, r *http.Request) {
	st, ok := s.publicGate(w, r)
	if !ok {
		return
	}
	if st.PublicBaseURL == "" {
		s.writeError(w, fail(http.StatusConflict, "PUBLIC_BASE_URL_REQUIRED", "", "set the website address in the news settings"))
		return
	}
	items, err := s.publicFeedRows(r.Context(), 50000)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var b strings.Builder
	b.WriteString(xml.Header + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, p := range items {
		loc := st.publicURL(p.seo.Slug)
		if p.seo.Canonical != "" {
			loc = p.seo.Canonical
		}
		b.WriteString("<url><loc>" + html.EscapeString(loc) + "</loc><lastmod>" + p.updated.UTC().Format(time.RFC3339) + "</lastmod></url>\n")
	}
	b.WriteString("</urlset>\n")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

func (s *service) publicRSS(w http.ResponseWriter, r *http.Request) {
	st, ok := s.publicGate(w, r)
	if !ok {
		return
	}
	if !st.RSSEnabled {
		http.NotFound(w, r)
		return
	}
	if st.PublicBaseURL == "" {
		s.writeError(w, fail(http.StatusConflict, "PUBLIC_BASE_URL_REQUIRED", "", "set the website address in the news settings"))
		return
	}
	ctx := r.Context()
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	lang := langs.Choose(r.URL.Query().Get("lang"))
	items, err := s.publicFeedRows(ctx, 50)
	if err != nil {
		s.writeError(w, err)
		return
	}
	title := langs.Resolve(st.SiteTitle, lang)
	if title == "" {
		title = "News"
	}
	esc := html.EscapeString
	var b strings.Builder
	b.WriteString(xml.Header + `<rss version="2.0"><channel><title>` + esc(title) + `</title><link>` + esc(strings.TrimRight(st.PublicBaseURL, "/")) + `</link><description>` + esc(title) + `</description><language>` + esc(lang) + "</language>\n")
	for _, p := range items {
		t := langs.Resolve(parseI18n(p.row.Title), lang)
		sum := langs.Resolve(parseI18n(p.row.Summary), lang)
		if sum == "" {
			sum = truncRunes(newsBodyText(p.row.Format, langs.Resolve(parseI18n(p.row.Body), lang)), 300)
		}
		link := st.publicURL(p.seo.Slug)
		b.WriteString("<item><title>" + esc(t) + "</title><link>" + esc(link) + "</link><guid isPermaLink=\"false\">" + esc(p.row.ID) + "</guid><pubDate>" + p.row.Published.UTC().Format(time.RFC1123Z) + "</pubDate><description>" + esc(sum) + "</description></item>\n")
	}
	b.WriteString("</channel></rss>\n")
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

func jsonUnmarshalStrings(raw string, out *[]string) error {
	return json.Unmarshal([]byte(raw), out)
}

func writePublic(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}
