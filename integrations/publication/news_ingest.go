package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	ingestPrefix     = "/v2/publication/news/ingest/"
	ingestMaxBody    = 256 << 10 // the same as the server's default socket.max_request_size_bytes, which also caps this route
	ingestMaxItems   = 50
	newsSourcePrefix = "webhook:"
)

var (
	newsSourceKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,39}$`)
	headerNamePattern    = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)
	mappingFields        = []string{"title", "summary", "body", "cover", "category", "tags", "lang", "publishedAt", "expiresAt", "externalId", "url",
		"slug", "metaTitle", "metaDescription", "keywords", "canonicalUrl", "ogImage", "noindex"}
	defaultMapping = map[string]string{"title": "title", "summary": "summary", "body": "body", "cover": "image", "category": "category", "tags": "tags", "lang": "lang",
		"publishedAt": "publishedAt", "expiresAt": "expiresAt", "externalId": "id", "url": "url",
		"slug": "slug", "metaTitle": "seo.title", "metaDescription": "seo.description", "keywords": "keywords", "canonicalUrl": "canonicalUrl", "ogImage": "seo.image", "noindex": "noindex"}
)

type newsSource struct {
	ID, Key, Name, AuthMode, TokenHeader, SignatureHeader, TimestampHeader string
	SkewSeconds                                                            int
	Secret                                                                 string
	ItemsPath, BodyFormat, DefaultLang, PublishMode, UpdateMode            string
	PublishedAtMode, NotifyMode                                            string
	Mapping                                                                map[string]string
	DefaultCategoryID                                                      *string
	DefaultTagIDs                                                          []string
	RateLimit                                                              int
	CopyImages, AllowComment, AllowReaction, Active                        bool
}

func (src newsSource) sourceName() string { return newsSourcePrefix + src.Key }

const newsSourceJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'isActive',t.is_active,'authMode',t.auth_mode,'tokenHeader',t.token_header,
	'signatureHeader',t.signature_header,'timestampHeader',t.timestamp_header,'skewSeconds',t.skew_seconds,'secretSet',t.secret <> '','secretHint',t.secret_hint,
	'secretRotatedAt',t.secret_rotated_at,'itemsPath',t.items_path,'mapping',t.mapping,'bodyFormat',t.body_format,'defaultCategoryId',t.default_category_id,
	'defaultLang',t.default_lang,'defaultTagIds',to_jsonb(t.default_tag_ids),'publishMode',t.publish_mode,'updateMode',t.update_mode,'publishedAtMode',t.published_at_mode,
	'rateLimitPerMinute',t.rate_limit_per_minute,'copyImages',t.copy_images,'notifyMode',t.notify_mode,'allowComment',t.allow_comment,'allowReaction',t.allow_reaction,
	'lastReceivedAt',t.last_received_at,'endpoint','` + ingestPrefix + `' || t.key,
	'newsCount',(SELECT count(*) FROM publication_news n WHERE n.source_id=t.id),
	'pendingCount',(SELECT count(*) FROM publication_news n WHERE n.source_id=t.id AND n.status='pending_review'),
	'received24h',(SELECT count(*) FROM publication_news_ingest_log l WHERE l.source_id=t.id AND NOT l.is_test AND l.created_at > now() - interval '1 day'),
	'failed24h',(SELECT count(*) FROM publication_news_ingest_log l WHERE l.source_id=t.id AND NOT l.is_test AND l.status IN ('rejected','error') AND l.created_at > now() - interval '1 day'),
	'createdBy',t.created_by,'updatedBy',t.updated_by,'createdAt',t.created_at,'updatedAt',t.updated_at)`

func newsSourceResource() resource {
	return resource{
		name:    "news-sources",
		table:   "publication_news_source",
		json:    newsSourceJSON,
		from:    ` FROM publication_news_source t`,
		search:  []string{"t.key", "t.name"},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean", "publishMode": "t.publish_mode = ?", "authMode": "t.auth_mode = ?"},
		order:   "t.key, t.id",
		save:    saveNewsSource,
	}
}

func randomSecret() string {
	raw := make([]byte, 32)
	_, _ = rand.Read(raw)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func secretHint(secret string) string {
	if len(secret) < 8 {
		return ""
	}
	return "…" + secret[len(secret)-4:]
}

func pickOne(field, v, def string, allowed ...string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return def, nil
	}
	if !containsStr(allowed, v) {
		return "", badRequest("INVALID_VALUE", field, field+" must be one of "+strings.Join(allowed, ", "))
	}
	return v, nil
}

func saveNewsSource(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key                string            `json:"key"`
		Name               string            `json:"name"`
		IsActive           *bool             `json:"isActive"`
		AuthMode           string            `json:"authMode"`
		TokenHeader        string            `json:"tokenHeader"`
		SignatureHeader    string            `json:"signatureHeader"`
		TimestampHeader    string            `json:"timestampHeader"`
		SkewSeconds        *int              `json:"skewSeconds"`
		Secret             string            `json:"secret"`
		ItemsPath          string            `json:"itemsPath"`
		Mapping            map[string]string `json:"mapping"`
		BodyFormat         string            `json:"bodyFormat"`
		DefaultCategoryID  string            `json:"defaultCategoryId"`
		DefaultLang        string            `json:"defaultLang"`
		DefaultTagIDs      []string          `json:"defaultTagIds"`
		PublishMode        string            `json:"publishMode"`
		UpdateMode         string            `json:"updateMode"`
		PublishedAtMode    string            `json:"publishedAtMode"`
		RateLimitPerMinute *int              `json:"rateLimitPerMinute"`
		CopyImages         bool              `json:"copyImages"`
		NotifyMode         string            `json:"notifyMode"`
		AllowComment       *bool             `json:"allowComment"`
		AllowReaction      *bool             `json:"allowReaction"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	key := strings.TrimSpace(in.Key)
	if !newsSourceKeyPattern.MatchString(key) {
		return badRequest("INVALID_KEY", "key", "key must start with a letter and use a-z, 0-9, _ and - (2-40)")
	}
	if !create {
		var old string
		if err := tx.QueryRowContext(ctx, `SELECT key FROM publication_news_source WHERE id=$1 FOR UPDATE`, id).Scan(&old); errors.Is(err, sql.ErrNoRows) {
			return notFound("NOT_FOUND")
		} else if err != nil {
			return err
		}
		if old != key {
			return badRequest("KEY_LOCKED", "key", "the key cannot change after creation")
		}
	} else {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_news_source WHERE key=$1)`, key).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return conflict("KEY_EXISTS", "this key is already used")
		}
	}
	name := strings.TrimSpace(in.Name)
	if utf8.RuneCountInString(name) > 120 {
		return badRequest("TEXT_TOO_LONG", "name", "name is too long (max 120)")
	}
	var err error
	auth, err := pickOne("authMode", in.AuthMode, "token", "token", "hmac")
	if err != nil {
		return err
	}
	headers := map[string]*string{"tokenHeader": &in.TokenHeader, "signatureHeader": &in.SignatureHeader, "timestampHeader": &in.TimestampHeader}
	defaults := map[string]string{"tokenHeader": "X-News-Token", "signatureHeader": "X-News-Signature", "timestampHeader": "X-News-Timestamp"}
	for f, p := range headers {
		*p = strings.TrimSpace(*p)
		if *p == "" {
			*p = defaults[f]
		}
		if !headerNamePattern.MatchString(*p) {
			return badRequest("INVALID_HEADER", f, "header names use A-Z, a-z, 0-9 and - (max 64)")
		}
	}
	skew := 300
	if in.SkewSeconds != nil {
		skew = *in.SkewSeconds
	}
	if skew < 30 || skew > 3600 {
		return badRequest("OUT_OF_RANGE", "skewSeconds", "between 30 and 3600 seconds")
	}
	rate := 60
	if in.RateLimitPerMinute != nil {
		rate = *in.RateLimitPerMinute
	}
	if rate < 0 || rate > 10000 {
		return badRequest("OUT_OF_RANGE", "rateLimitPerMinute", "between 0 and 10000")
	}
	if !validJSONPath(in.ItemsPath) {
		return badRequest("INVALID_PATH", "itemsPath", "use a JSON path like data.items")
	}
	mapping := map[string]string{}
	for k := range in.Mapping {
		if !containsStr(mappingFields, k) {
			return badRequest("INVALID_MAPPING", "mapping."+k, "unknown field "+k)
		}
	}
	for _, f := range mappingFields {
		v, ok := in.Mapping[f]
		if !ok {
			v = defaultMapping[f]
		}
		v = strings.TrimSpace(v)
		if v != "" && !validJSONPath(v) {
			return badRequest("INVALID_PATH", "mapping."+f, "use a JSON path like data.title or items[0].name")
		}
		mapping[f] = v
	}
	if mapping["title"] == "" {
		return badRequest("REQUIRED", "mapping.title", "the title path is required")
	}
	mappingJSON, _ := json.Marshal(mapping)
	format, err := pickOne("bodyFormat", in.BodyFormat, "auto", "auto", "html", "markdown", "text")
	if err != nil {
		return err
	}
	cat, err := cleanUUID("defaultCategoryId", in.DefaultCategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, cat, "news"); err != nil {
		return err
	}
	lang := localized.Normalize(in.DefaultLang)
	if lang != "" {
		langs, err := localized.Load(ctx, tx)
		if err != nil {
			return err
		}
		if !langs.Known(lang) {
			return badRequest("INVALID_LANGUAGE", "defaultLang", "unknown language "+lang)
		}
	}
	tags, err := cleanTagIDs(ctx, tx, "defaultTagIds", in.DefaultTagIDs, "news", "general")
	if err != nil {
		return err
	}
	publish, err := pickOne("publishMode", in.PublishMode, "review", "review", "auto")
	if err != nil {
		return err
	}
	update, err := pickOne("updateMode", in.UpdateMode, "update", "update", "ignore")
	if err != nil {
		return err
	}
	pubAt, err := pickOne("publishedAtMode", in.PublishedAtMode, "source", "source", "received")
	if err != nil {
		return err
	}
	notify, err := pickOne("notifyMode", in.NotifyMode, "none", newsNotifyModes...)
	if err != nil {
		return err
	}
	who := ctxConsoleUser(ctx)
	cols := []string{"key", "name", "is_active", "auth_mode", "token_header", "signature_header", "timestamp_header", "skew_seconds", "items_path", "mapping", "body_format",
		"default_category_id", "default_lang", "default_tag_ids", "publish_mode", "update_mode", "published_at_mode", "rate_limit_per_minute", "copy_images", "notify_mode",
		"allow_comment", "allow_reaction", "updated_by"}
	vals := []any{key, name, boolOr(in.IsActive, true), auth, in.TokenHeader, in.SignatureHeader, in.TimestampHeader, skew, strings.TrimSpace(in.ItemsPath), string(mappingJSON), format,
		cat, lang, "{" + strings.Join(tags, ",") + "}", publish, update, pubAt, rate, in.CopyImages, notify, boolOr(in.AllowComment, true), boolOr(in.AllowReaction, true), who}
	secret := strings.TrimSpace(in.Secret)
	if secret != "" && (len(secret) < 16 || len(secret) > 256 || strings.ContainsAny(secret, " \t\r\n")) {
		return badRequest("INVALID_SECRET", "secret", "the secret needs 16 to 256 characters without spaces")
	}
	if create && secret == "" {
		secret = randomSecret()
	}
	if secret != "" {
		sealed, err := seal(s.cfg.sealKey, []byte(secret))
		if err != nil {
			return err
		}
		cols = append(cols, "secret", "secret_hint", "secret_rotated_at")
		vals = append(vals, sealed, secretHint(secret), time.Now().UTC())
	}
	if create {
		cols = append(cols, "created_by")
		vals = append(vals, who)
	}
	return execSave(ctx, tx, "publication_news_source", id, create, cols, vals)
}

func scanNewsSource(row interface{ Scan(...any) error }, key []byte) (newsSource, error) {
	var src newsSource
	var sealed, mapping, tags string
	var cat sql.NullString
	err := row.Scan(&src.ID, &src.Key, &src.Name, &src.Active, &src.AuthMode, &src.TokenHeader, &src.SignatureHeader, &src.TimestampHeader, &src.SkewSeconds, &sealed,
		&src.ItemsPath, &mapping, &src.BodyFormat, &cat, &src.DefaultLang, &tags, &src.PublishMode, &src.UpdateMode, &src.PublishedAtMode, &src.RateLimit, &src.CopyImages,
		&src.NotifyMode, &src.AllowComment, &src.AllowReaction)
	if err != nil {
		return src, err
	}
	if cat.Valid {
		src.DefaultCategoryID = &cat.String
	}
	_ = json.Unmarshal([]byte(mapping), &src.Mapping)
	if src.Mapping == nil {
		src.Mapping = map[string]string{}
	}
	_ = json.Unmarshal([]byte(tags), &src.DefaultTagIDs)
	if sealed != "" {
		plain, err := unseal(key, sealed)
		if err != nil {
			return src, err
		}
		src.Secret = string(plain)
	}
	return src, nil
}

const newsSourceCols = `id::text, key, name, is_active, auth_mode, token_header, signature_header, timestamp_header, skew_seconds, secret, items_path, mapping::text, body_format,
	default_category_id::text, default_lang, to_jsonb(default_tag_ids)::text, publish_mode, update_mode, published_at_mode, rate_limit_per_minute, copy_images, notify_mode, allow_comment, allow_reaction`

func (s *service) loadNewsSource(ctx context.Context, where string, arg any) (newsSource, error) {
	src, err := scanNewsSource(s.db.QueryRowContext(ctx, `SELECT `+newsSourceCols+` FROM publication_news_source WHERE `+where, arg), s.cfg.sealKey)
	if errors.Is(err, sql.ErrNoRows) {
		return src, notFound("SOURCE_NOT_FOUND")
	}
	return src, err
}

type ingestResult struct {
	Index      int            `json:"index"`
	ExternalID string         `json:"externalId,omitempty"`
	Status     string         `json:"status"`
	NewsID     string         `json:"newsId,omitempty"`
	Code       string         `json:"code,omitempty"`
	Message    string         `json:"message,omitempty"`
	Notes      []string       `json:"notes,omitempty"`
	Preview    map[string]any `json:"preview,omitempty"`
}

type ingestSummary struct {
	Results   []ingestResult `json:"results"`
	Created   int            `json:"created"`
	Updated   int            `json:"updated"`
	Unchanged int            `json:"unchanged"`
	Ignored   int            `json:"ignored"`
	Failed    int            `json:"failed"`
}

func (s *service) ingestHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := pathVar(r, "source")
	if !newsSourceKeyPattern.MatchString(key) {
		s.writeError(w, notFound("SOURCE_NOT_FOUND"))
		return
	}
	src, err := s.loadNewsSource(ctx, `key=$1 AND is_active`, key)
	if err != nil {
		s.writeError(w, err)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, ingestMaxBody+1))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) || len(body) > ingestMaxBody {
		s.writeError(w, fail(http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "", "the body is larger than 256 KB"))
		return
	}
	if err != nil {
		s.writeError(w, badRequest("INVALID_BODY", "", "cannot read body"))
		return
	}
	if src.RateLimit > 0 {
		d, err := clusterkit.Allow(ctx, s.db, "news-ingest:"+src.ID, int64(src.RateLimit), time.Minute)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if !d.Allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(d.RetryAfter.Seconds()))))
			s.writeError(w, fail(http.StatusTooManyRequests, "RATE_LIMITED", "", "too many requests for this source; wait a minute"))
			return
		}
	}
	st, err := loadNewsSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	remote := r.RemoteAddr
	if err := s.checkIngestAuth(ctx, src, r, body); err != nil {
		app, _ := asApp(err)
		if app == nil {
			s.writeError(w, err)
			return
		}
		s.writeIngestLog(ctx, src, st, "rejected", app.Status, nil, app.Code, app.Message, "", remote, false, "")
		s.writeError(w, err)
		return
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE publication_news_source SET last_received_at=now() WHERE id=$1`, src.ID)
	sum, err := s.runIngest(ctx, src, st, body, false, false)
	if err != nil {
		app, ok := asApp(err)
		if !ok {
			s.writeIngestLog(ctx, src, st, "error", 500, nil, "INTERNAL_ERROR", err.Error(), string(body), remote, false, "")
			s.writeError(w, err)
			return
		}
		s.writeIngestLog(ctx, src, st, "rejected", app.Status, nil, app.Code, app.Message, string(body), remote, false, "")
		s.writeError(w, err)
		return
	}
	s.writeIngestLog(ctx, src, st, sum.logStatus(), http.StatusOK, sum, "", "", string(body), remote, false, "")
	writeJSON(w, http.StatusOK, sum)
}

func (sum *ingestSummary) logStatus() string {
	switch {
	case sum.Created > 0:
		return "created"
	case sum.Updated > 0:
		return "updated"
	case sum.Failed > 0 && sum.Unchanged == 0 && sum.Ignored == 0:
		return "error"
	case sum.Ignored > 0 && sum.Unchanged == 0:
		return "ignored"
	}
	return "unchanged"
}

func (s *service) checkIngestAuth(ctx context.Context, src newsSource, r *http.Request, body []byte) error {
	if src.Secret == "" {
		return fail(http.StatusUnauthorized, "SOURCE_SECRET_MISSING", "", "this source has no secret")
	}
	if src.AuthMode == "token" {
		got := strings.TrimSpace(r.Header.Get(src.TokenHeader))
		if strings.EqualFold(src.TokenHeader, "Authorization") {
			got = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(got, "Bearer "), "bearer "))
		}
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(src.Secret)) != 1 {
			return fail(http.StatusUnauthorized, "INVALID_TOKEN", "", "the token is missing or wrong")
		}
		return nil
	}
	tsRaw := strings.TrimSpace(r.Header.Get(src.TimestampHeader))
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		return fail(http.StatusUnauthorized, "INVALID_TIMESTAMP", "", "the timestamp header is missing or not a unix time")
	}
	if ts > 1e12 {
		ts /= 1000
	}
	now := s.now().Unix()
	if ts < now-int64(src.SkewSeconds) || ts > now+int64(src.SkewSeconds) {
		return fail(http.StatusUnauthorized, "TIMESTAMP_OUT_OF_WINDOW", "", "the timestamp is too far from the server time")
	}
	sig := strings.ToLower(strings.TrimSpace(r.Header.Get(src.SignatureHeader)))
	sig = strings.TrimPrefix(sig, "sha256=")
	want := signIngest(src.Secret, tsRaw, body)
	if sig == "" || subtle.ConstantTimeCompare([]byte(sig), []byte(want)) != 1 {
		return fail(http.StatusUnauthorized, "INVALID_SIGNATURE", "", "the signature does not match")
	}
	res, err := s.db.ExecContext(ctx, `INSERT INTO publication_news_replay (source_id, signature, expires_at) VALUES ($1,$2, now() + $3::int * interval '2 second') ON CONFLICT DO NOTHING`, src.ID, sig, src.SkewSeconds)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fail(http.StatusConflict, "REPLAYED", "", "this signed request was already received")
	}
	return nil
}

func signIngest(secret, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func extractItems(src newsSource, payload any) ([]any, error) {
	root := payload
	if strings.TrimSpace(src.ItemsPath) != "" {
		v, ok := lookupJSONPath(payload, src.ItemsPath)
		if !ok {
			return nil, badRequest("ITEMS_NOT_FOUND", "itemsPath", "nothing at "+src.ItemsPath)
		}
		root = v
	}
	var items []any
	switch v := root.(type) {
	case []any:
		items = v
	case map[string]any:
		items = []any{v}
	default:
		return nil, badRequest("ITEMS_NOT_FOUND", "", "send a JSON object or an array of objects")
	}
	if len(items) == 0 {
		return nil, badRequest("NO_ITEMS", "", "the payload has no items")
	}
	if len(items) > ingestMaxItems {
		return nil, badRequest("TOO_MANY_ITEMS", "", "at most 50 items per request")
	}
	return items, nil
}

func (s *service) runIngest(ctx context.Context, src newsSource, st newsSettings, body []byte, dry, test bool) (*ingestSummary, error) {
	var payload any
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return nil, badRequest("INVALID_JSON", "", "the body is not valid JSON")
	}
	items, err := extractItems(src, payload)
	if err != nil {
		return nil, err
	}
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	sum := &ingestSummary{Results: []ingestResult{}}
	for i, item := range items {
		res := s.ingestItem(ctx, src, st, langs, item, dry)
		res.Index = i
		switch res.Status {
		case "created":
			sum.Created++
		case "updated":
			sum.Updated++
		case "unchanged":
			sum.Unchanged++
		case "ignored":
			sum.Ignored++
		case "failed":
			sum.Failed++
		}
		if !test && !dry {
			res.Preview = nil
		}
		sum.Results = append(sum.Results, res)
	}
	return sum, nil
}

func scalarString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case json.Number:
		return x.String()
	case bool:
		return strconv.FormatBool(x)
	}
	return ""
}

// imageURL reads an image address given as a string, an object such as {"url": …}, {"src": …}, {"href": …} or {"@url": …}
// (also one level deeper, as in {"image": {"url": …}}), or a list of those, where the first usable entry wins.
func imageURL(v any) string {
	return imageURLDepth(v, 0)
}

func imageURLDepth(v any, depth int) string {
	if depth > 2 {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case map[string]any:
		for _, k := range []string{"url", "src", "href", "@url", "link", "uri"} {
			if s := scalarString(x[k]); s != "" {
				return s
			}
		}
		for _, k := range []string{"image", "original", "large", "full"} {
			if s := imageURLDepth(x[k], depth+1); s != "" {
				return s
			}
		}
	case []any:
		for _, e := range x {
			if s := imageURLDepth(e, depth+1); s != "" {
				return s
			}
		}
	}
	return ""
}

func parseIngestTime(v any) (*time.Time, bool) {
	switch x := v.(type) {
	case nil:
		return nil, true
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			f, err := x.Float64()
			if err != nil {
				return nil, false
			}
			n = int64(f)
		}
		if n > 1e12 {
			n /= 1000
		}
		t := time.Unix(n, 0).UTC()
		return &t, true
	case string:
		x = strings.TrimSpace(x)
		if x == "" {
			return nil, true
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, time.RFC1123Z, time.RFC1123, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, x); err == nil {
				t = t.UTC()
				return &t, true
			}
		}
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return parseIngestTime(json.Number(strconv.FormatInt(n, 10)))
		}
	}
	return nil, false
}

func truncRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:max-1])) + "…"
}

func plainLine(s string) string {
	if looksLikeHTML(s) {
		s = richText(s)
	}
	return strings.Join(strings.Fields(s), " ")
}

type mappedItem struct {
	Title, Summary, Body, Cover, CategoryKey, Lang, ExternalID, URL string
	Tags                                                            []string
	PublishedAt, ExpiresAt                                          *time.Time
	Slug, MetaTitle, MetaDescription, Canonical, OgImage            string
	Keywords                                                        []string
	Noindex                                                         bool
}

func stringList(v any) []string {
	var out []string
	switch x := v.(type) {
	case []any:
		for _, e := range x {
			if k := scalarString(e); k != "" {
				out = append(out, k)
			} else if obj, ok := e.(map[string]any); ok {
				if k := scalarString(obj["key"]); k != "" {
					out = append(out, k)
				}
			}
		}
	case string:
		for _, k := range strings.Split(x, ",") {
			if k = strings.TrimSpace(k); k != "" {
				out = append(out, k)
			}
		}
	}
	return out
}

func mapIngestItem(src newsSource, item any) (mappedItem, []string, error) {
	var m mappedItem
	var notes []string
	get := func(field string) any {
		path := src.Mapping[field]
		if path == "" {
			return nil
		}
		v, _ := lookupJSONPath(item, path)
		return v
	}
	m.Title = truncRunes(plainLine(scalarString(get("title"))), 200)
	if m.Title == "" {
		return m, notes, badRequest("ITEM_TITLE_REQUIRED", "mapping.title", "no title at "+src.Mapping["title"])
	}
	m.Summary = truncRunes(plainLine(scalarString(get("summary"))), 1000)
	rawBody := scalarString(get("body"))
	format := src.BodyFormat
	if format == "auto" {
		format = "markdown"
		if looksLikeHTML(rawBody) {
			format = "html"
		}
	}
	switch format {
	case "html":
		m.Body = sanitizeRichHTML(rawBody)
	case "markdown":
		m.Body = sanitizeRichHTML(renderMarkdown(rawBody))
	default:
		m.Body = textToRich(rawBody)
	}
	if utf8.RuneCountInString(m.Body) > newsRichMax {
		return m, notes, badRequest("ITEM_BODY_TOO_LONG", "mapping.body", "the body is too long")
	}
	if strings.TrimSpace(richText(m.Body)) == "" {
		fallback := m.Summary
		if fallback == "" {
			fallback = m.Title
		}
		m.Body = textToRich(fallback)
		notes = append(notes, "BODY_EMPTY_USED_SUMMARY")
	}
	if c := imageURL(get("cover")); c != "" {
		if safeURL(c) && len(c) <= 1024 {
			m.Cover = c
		} else {
			notes = append(notes, "COVER_URL_IGNORED")
		}
	}
	if u := scalarString(get("url")); u != "" {
		if safeURL(u) && len(u) <= 1024 {
			m.URL = u
		} else {
			notes = append(notes, "SOURCE_URL_IGNORED")
		}
	}
	m.CategoryKey = scalarString(get("category"))
	if obj, ok := get("category").(map[string]any); ok {
		m.CategoryKey = scalarString(obj["key"])
	}
	m.Tags = stringList(get("tags"))
	if len(m.Tags) > 50 {
		m.Tags = m.Tags[:50]
	}
	m.Lang = scalarString(get("lang"))
	var ok bool
	if m.PublishedAt, ok = parseIngestTime(get("publishedAt")); !ok {
		notes = append(notes, "PUBLISHED_AT_IGNORED")
	}
	if m.ExpiresAt, ok = parseIngestTime(get("expiresAt")); !ok {
		notes = append(notes, "EXPIRES_AT_IGNORED")
	}
	m.ExternalID = truncRunes(scalarString(get("externalId")), 200)
	if raw := scalarString(get("slug")); raw != "" {
		if slug, err := cleanSlug(raw); err == nil {
			m.Slug = slug
		} else {
			notes = append(notes, "SLUG_IGNORED")
		}
	}
	m.MetaTitle = truncRunes(plainLine(scalarString(get("metaTitle"))), 200)
	m.MetaDescription = truncRunes(plainLine(scalarString(get("metaDescription"))), 400)
	if kw, err := cleanKeywords(stringList(get("keywords"))); err == nil {
		m.Keywords = kw
	} else {
		notes = append(notes, "KEYWORDS_IGNORED")
	}
	if c := scalarString(get("canonicalUrl")); c != "" {
		if v, err := cleanAbsURL("canonicalUrl", c); err == nil {
			m.Canonical = v
		} else {
			notes = append(notes, "CANONICAL_URL_IGNORED")
		}
	}
	if o := imageURL(get("ogImage")); o != "" {
		if safeURL(o) && len(o) <= 1024 {
			m.OgImage = o
		} else {
			notes = append(notes, "OG_IMAGE_IGNORED")
		}
	}
	if b, ok := get("noindex").(bool); ok {
		m.Noindex = b
	}
	return m, notes, nil
}

func (m mappedItem) hash() string {
	b, _ := json.Marshal(m)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (s *service) ingestItem(ctx context.Context, src newsSource, st newsSettings, langs *localized.Set, item any, dry bool) ingestResult {
	res := ingestResult{Status: "failed"}
	m, notes, err := mapIngestItem(src, item)
	res.Notes = notes
	if err != nil {
		app, _ := asApp(err)
		res.Code, res.Message = app.Code, app.Message
		return res
	}
	hash := m.hash()
	ext := m.ExternalID
	if ext == "" {
		ext = "sha256:" + hash
	}
	res.ExternalID = ext
	lang := langs.Choose(m.Lang, src.DefaultLang)
	now := s.now().UTC()
	published := now
	if src.PublishedAtMode == "source" && m.PublishedAt != nil {
		published = *m.PublishedAt
	}
	expires := m.ExpiresAt
	if expires != nil && !expires.After(published) {
		expires = nil
		res.Notes = append(res.Notes, "EXPIRES_AT_IGNORED")
	}
	if expires != nil && !expires.After(now) {
		res.Code, res.Message = "ITEM_EXPIRED", "the item has already expired"
		return res
	}
	text := func(v string) i18n {
		out := i18n{}
		if v == "" {
			return out
		}
		out[lang] = v
		if st.FillDefaultLang && lang != langs.Default && langs.Default != "" {
			out[langs.Default] = v
		}
		return out
	}
	status := "pending_review"
	if src.PublishMode == "auto" {
		status = "published"
	}
	res.Preview = map[string]any{"title": m.Title, "summary": m.Summary, "bodyHtml": m.Body, "coverUrl": m.Cover, "categoryKey": m.CategoryKey, "tags": m.Tags, "lang": lang,
		"publishedAt": published, "expiresAt": expires, "externalId": ext, "sourceUrl": m.URL, "status": status}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var cat *string = src.DefaultCategoryID
		if m.CategoryKey != "" {
			var id string
			err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='news' AND lower(key)=lower($1)`, m.CategoryKey).Scan(&id)
			if err == nil {
				cat = &id
			} else if errors.Is(err, sql.ErrNoRows) {
				res.Notes = append(res.Notes, "CATEGORY_UNKNOWN")
			} else {
				return err
			}
		}
		tags := append([]string{}, src.DefaultTagIDs...)
		if len(m.Tags) > 0 {
			lower := make([]string, len(m.Tags))
			for i, t := range m.Tags {
				lower[i] = strings.ToLower(t)
			}
			rows, err := tx.QueryContext(ctx, `SELECT id::text, lower(key) FROM base_tag WHERE key_type IN ('news','general') AND lower(key) = ANY($1::text[])`, "{"+quoteArray(lower)+"}")
			if err != nil {
				return err
			}
			found := map[string]bool{}
			for rows.Next() {
				var id, k string
				if err := rows.Scan(&id, &k); err != nil {
					rows.Close()
					return err
				}
				found[k] = true
				if !containsStr(tags, id) {
					tags = append(tags, id)
				}
			}
			rows.Close()
			if len(found) < len(lower) {
				res.Notes = append(res.Notes, "TAGS_UNKNOWN")
			}
		}
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('publication-news-ingest:' || $1 || ':' || $2))`, src.sourceName(), ext); err != nil {
			return err
		}
		var id, oldHash string
		err := tx.QueryRowContext(ctx, `SELECT id::text, content_hash FROM publication_news WHERE source=$1 AND external_id=$2 FOR UPDATE`, src.sourceName(), ext).Scan(&id, &oldHash)
		exists := err == nil
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		switch {
		case exists && oldHash == hash:
			res.Status, res.NewsID = "unchanged", id
			return nil
		case exists && src.UpdateMode == "ignore":
			res.Status, res.NewsID = "ignored", id
			return nil
		}
		if dry {
			res.Status = "created"
			if exists {
				res.Status, res.NewsID = "updated", id
			}
			return nil
		}
		cover := m.Cover
		if src.CopyImages && cover != "" {
			if copied, err := s.copyRemoteImage(ctx, cover); err != nil {
				res.Notes = append(res.Notes, "COVER_COPY_FAILED")
				s.logger.Warn("Publication news cover copy failed for %s: %v", src.Key, err)
			} else {
				cover = copied
			}
		}
		tagArr := "{" + strings.Join(tags, ",") + "}"
		kwArr := "{" + quoteArray(m.Keywords) + "}"
		slug := m.Slug
		if slug != "" {
			self := id
			if self == "" {
				self = systemUserID
			}
			taken, err := slugTaken(ctx, tx, slug, self)
			if err != nil {
				return err
			}
			if taken {
				res.Notes = append(res.Notes, "SLUG_TAKEN")
				slug = ""
			}
		}
		if exists {
			if slug != "" {
				if _, err := tx.ExecContext(ctx, `UPDATE publication_news SET slug=$2 WHERE id=$1`, id, slug); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, `UPDATE publication_news SET meta_title=$2, meta_description=$3, keywords=$4, canonical_url=$5, og_image=$6, noindex=$7 WHERE id=$1`,
				id, i18nJSON(text(m.MetaTitle)), i18nJSON(text(m.MetaDescription)), kwArr, m.Canonical, m.OgImage, m.Noindex); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, `UPDATE publication_news SET title=$2, summary=$3, body=$4, body_format='html', cover_url=$5, category_id=$6, tag_ids=$7, published_at=$8, expires_at=$9,
				source_url=$10, lang=$11, content_hash=$12, source_id=$13, updated_by=$14, updated_at=now() WHERE id=$1`,
				id, i18nJSON(text(m.Title)), i18nJSON(text(m.Summary)), i18nJSON(text(m.Body)), cover, cat, tagArr, published, expires, m.URL, lang, hash, src.ID, src.sourceName())
			if err != nil {
				return err
			}
			res.Status, res.NewsID = "updated", id
			return nil
		}
		id = uuid.Must(uuid.NewV4()).String()
		if slug == "" {
			if slug, err = uniqueSlug(ctx, tx, slugify(m.Title, st.SlugStyle), id); err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO publication_news (slug, meta_title, meta_description, keywords, canonical_url, og_image, noindex, id, title, summary, body, body_format, cover_url, category_id, tag_ids, status, is_active, published_at, expires_at,
			allow_comment, allow_reaction, notify_mode, source, source_id, external_id, source_url, content_hash, lang, created_by, updated_by, approved_by, approved_at)
			VALUES ($23,$24,$25,$26,$27,$28,$29,$1,$2,$3,$4,'html',$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$15,$15,$21,$22)`,
			id, i18nJSON(text(m.Title)), i18nJSON(text(m.Summary)), i18nJSON(text(m.Body)), cover, cat, tagArr, status, status == "published", published, expires,
			src.AllowComment, src.AllowReaction, src.NotifyMode, src.sourceName(), src.ID, ext, m.URL, hash, lang, nullIf(status == "published", src.sourceName()), nullTime(status == "published", now),
			slug, i18nJSON(text(m.MetaTitle)), i18nJSON(text(m.MetaDescription)), kwArr, m.Canonical, m.OgImage, m.Noindex)
		if err != nil {
			return err
		}
		res.Status, res.NewsID = "created", id
		return nil
	})
	if err != nil {
		res.Status = "failed"
		if app, ok := asApp(err); ok {
			res.Code, res.Message = app.Code, app.Message
		} else {
			s.logger.Error("Publication news ingest item failed for %s: %v", src.Key, err)
			res.Code, res.Message = "INTERNAL_ERROR", "the item could not be saved"
		}
	}
	return res
}

func nullIf(ok bool, v string) any {
	if !ok {
		return nil
	}
	return v
}

func nullTime(ok bool, t time.Time) any {
	if !ok {
		return nil
	}
	return t
}

func quoteArray(items []string) string {
	out := make([]string, len(items))
	for i, v := range items {
		out[i] = `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + `"`
	}
	return strings.Join(out, ",")
}

func (s *service) writeIngestLog(ctx context.Context, src newsSource, st newsSettings, status string, httpStatus int, sum *ingestSummary, code, msg, payload, remote string, test bool, who string) string {
	if len(payload) > st.LogPayloadBytes {
		cut := st.LogPayloadBytes
		for cut > 0 && !utf8.RuneStart(payload[cut]) {
			cut--
		}
		payload = payload[:cut]
	}
	results := "[]"
	var items, created, updated, unchanged, failed int
	if sum != nil {
		b, _ := json.Marshal(sum.Results)
		results = string(b)
		items, created, updated, unchanged, failed = len(sum.Results), sum.Created, sum.Updated, sum.Unchanged+sum.Ignored, sum.Failed
	}
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO publication_news_ingest_log (id, source_id, source_key, status, http_status, items, created_count, updated_count, unchanged_count, failed_count,
		error_code, message, results, payload, remote_addr, is_test, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::jsonb,$14,$15,$16,$17)`,
		id, src.ID, src.Key, status, httpStatus, items, created, updated, unchanged, failed, code, truncRunes(msg, 500), results, payload, truncRunes(remote, 100), test, who); err != nil {
		s.logger.Error("Publication news ingest log failed: %v", err)
	}
	return id
}

func (s *service) consoleNewsSourceTest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Payload json.RawMessage `json:"payload"`
		Save    bool            `json:"save"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("SOURCE_NOT_FOUND"))
		return
	}
	src, err := s.loadNewsSource(r.Context(), `id=$1`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	body := []byte(in.Payload)
	var str string
	if json.Unmarshal(in.Payload, &str) == nil {
		body = []byte(str)
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		s.writeError(w, badRequest("PAYLOAD_REQUIRED", "payload", "paste a sample payload"))
		return
	}
	if len(body) > ingestMaxBody {
		s.writeError(w, badRequest("PAYLOAD_TOO_LARGE", "payload", "the body is larger than 256 KB"))
		return
	}
	st, err := loadNewsSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	sum, err := s.runIngest(r.Context(), src, st, body, !in.Save, true)
	if err != nil {
		if app, ok := asApp(err); ok {
			logID := s.writeIngestLog(r.Context(), src, st, "rejected", app.Status, nil, app.Code, app.Message, string(body), "console", true, consoleUser(r))
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "code": app.Code, "message": app.Message, "field": app.Field, "logId": logID})
			return
		}
		s.writeError(w, err)
		return
	}
	status := sum.logStatus()
	if !in.Save {
		status = "preview"
	}
	logID := s.writeIngestLog(r.Context(), src, st, status, http.StatusOK, sum, "", "", string(body), "console", true, consoleUser(r))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "saved": in.Save, "logId": logID, "summary": sum})
}

func (s *service) consoleNewsSourceSecret(reveal bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathVar(r, "id")
		if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("SOURCE_NOT_FOUND"))
			return
		}
		src, err := s.loadNewsSource(r.Context(), `id=$1`, id)
		if err != nil {
			s.writeError(w, err)
			return
		}
		secret := src.Secret
		if !reveal {
			secret = randomSecret()
			sealed, err := seal(s.cfg.sealKey, []byte(secret))
			if err != nil {
				s.writeError(w, err)
				return
			}
			if _, err := s.db.ExecContext(r.Context(), `UPDATE publication_news_source SET secret=$2, secret_hint=$3, secret_rotated_at=now(), updated_by=$4, updated_at=now() WHERE id=$1`,
				id, sealed, secretHint(secret), consoleUser(r)); err != nil {
				s.writeError(w, err)
				return
			}
			s.logger.Info("Publication news source %s secret rotated by %s", src.Key, consoleUser(r))
		} else {
			s.logger.Info("Publication news source %s secret revealed to %s", src.Key, consoleUser(r))
		}
		writeJSON(w, http.StatusOK, map[string]any{"secret": secret, "secretHint": secretHint(secret)})
	}
}

const ingestLogJSON = `jsonb_build_object('id',l.id,'sourceId',l.source_id,'sourceKey',l.source_key,'status',l.status,'httpStatus',l.http_status,'items',l.items,
	'createdCount',l.created_count,'updatedCount',l.updated_count,'unchangedCount',l.unchanged_count,'failedCount',l.failed_count,'errorCode',l.error_code,'message',l.message,
	'remoteAddr',l.remote_addr,'isTest',l.is_test,'createdBy',l.created_by,'createdAt',l.created_at`

func (s *service) consoleIngestLogList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if v := q.Get("sourceId"); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("INVALID_FILTER", "sourceId", "must be a UUID"))
			return
		}
		where.add("l.source_id = ?::uuid", v)
	}
	if v := q.Get("status"); v != "" {
		where.add("l.status = ?", v)
	}
	if v := q.Get("isTest"); v == "true" || v == "false" {
		where.add("l.is_test = ?::boolean", v)
	}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(l.source_key ILIKE "+n+" OR l.error_code ILIKE "+n+" OR l.results::text ILIKE "+n+")")
	}
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM publication_news_ingest_log l`+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT (`+ingestLogJSON+`))::text FROM publication_news_ingest_log l`+where.sql()+` ORDER BY l.created_at DESC, l.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
		append(append([]any{}, where.args...), p.Limit, p.Offset)...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	items, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit})
}

func (s *service) consoleIngestLogGet(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	var raw string
	err := s.db.QueryRowContext(r.Context(), `SELECT (`+ingestLogJSON+`,'results',l.results,'payload',l.payload))::text FROM publication_news_ingest_log l WHERE l.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, []byte(raw))
}

func (s *service) sweepNewsIngest(ctx context.Context) error {
	st, err := loadNewsSettings(ctx, s.db)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM publication_news_ingest_log WHERE created_at < now() - $1::int * interval '1 day'`, st.LogRetentionDays); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM publication_news_replay WHERE expires_at < now()`)
	return err
}
