package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var versionPattern = regexp.MustCompile(`^(\d{1,9})(\.\d{1,9}){0,3}([-+][0-9A-Za-z.-]{1,40})?$`)

var appPlatforms = []string{"android", "ios", "web", "windows", "macos", "linux"}

var appStores = []string{"google_play", "app_store", "bazaar", "myket", "microsoft_store", "steam", "web", "direct"}

var platformNames = map[string]map[string]string{
	"android": {"fa": "اندروید", "en": "Android"},
	"ios":     {"fa": "آی‌اواس", "en": "iOS"},
	"web":     {"fa": "وب", "en": "Web"},
	"windows": {"fa": "ویندوز", "en": "Windows"},
	"macos":   {"fa": "مک", "en": "macOS"},
	"linux":   {"fa": "لینوکس", "en": "Linux"},
}

var storeNames = map[string]map[string]string{
	"google_play":     {"fa": "گوگل‌پلی", "en": "Google Play"},
	"app_store":       {"fa": "اپ‌استور", "en": "App Store"},
	"bazaar":          {"fa": "کافه‌بازار", "en": "Cafe Bazaar"},
	"myket":           {"fa": "مایکت", "en": "Myket"},
	"microsoft_store": {"fa": "مایکروسافت استور", "en": "Microsoft Store"},
	"steam":           {"fa": "استیم", "en": "Steam"},
	"web":             {"fa": "وب‌سایت", "en": "Website"},
	"direct":          {"fa": "دانلود مستقیم", "en": "Direct download"},
}

const (
	versionOK        = "ok"
	versionAvailable = "update_available"
	versionRequired  = "update_required"
)

type versionParts struct {
	nums []int64
	pre  string
}

func parseVersion(v string) (versionParts, bool) {
	v = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "v"))
	if !versionPattern.MatchString(v) {
		return versionParts{}, false
	}
	var out versionParts
	core := v
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		core = v[:i]
		if v[i] == '-' {
			out.pre = v[i+1:]
			if j := strings.IndexByte(out.pre, '+'); j >= 0 {
				out.pre = out.pre[:j]
			}
		}
	}
	for _, p := range strings.Split(core, ".") {
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return versionParts{}, false
		}
		out.nums = append(out.nums, n)
	}
	return out, true
}

func compareVersions(a, b string) (int, bool) {
	pa, ok1 := parseVersion(a)
	pb, ok2 := parseVersion(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	for i := 0; i < 4; i++ {
		var x, y int64
		if i < len(pa.nums) {
			x = pa.nums[i]
		}
		if i < len(pb.nums) {
			y = pb.nums[i]
		}
		if x != y {
			if x < y {
				return -1, true
			}
			return 1, true
		}
	}
	switch {
	case pa.pre == pb.pre:
		return 0, true
	case pa.pre == "":
		return 1, true
	case pb.pre == "":
		return -1, true
	case pa.pre < pb.pre:
		return -1, true
	}
	return 1, true
}

func versionStatus(current, latest, minimum string) string {
	if minimum != "" {
		if c, ok := compareVersions(current, minimum); ok && c < 0 {
			return versionRequired
		}
	}
	if latest != "" {
		if c, ok := compareVersions(current, latest); ok && c < 0 {
			return versionAvailable
		}
	}
	return versionOK
}

type storeLink struct {
	Store string         `json:"store"`
	URL   string         `json:"url"`
	Label localized.Text `json:"label"`
}

type appVersionInput struct {
	Platform      string         `json:"platform"`
	LatestVersion string         `json:"latestVersion"`
	MinVersion    string         `json:"minVersion"`
	StoreLinks    []storeLink    `json:"storeLinks"`
	UpdateMessage localized.Text `json:"updateMessage"`
	ForceMessage  localized.Text `json:"forceMessage"`
	IsActive      *bool          `json:"isActive"`
}

const appVersionJSON = `jsonb_build_object('id',v.id,'platform',v.platform,'latestVersion',v.latest_version,'minVersion',v.min_version,'storeLinks',v.store_links,
	'updateMessage',v.update_message,'forceMessage',v.force_message,'isActive',v.is_active,'createdAt',v.created_at,'updatedAt',v.updated_at)`

func cleanVersion(field, v string, required bool) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		if required {
			return "", badRequest("VERSION_REQUIRED", field, "a version is required")
		}
		return "", nil
	}
	if _, ok := parseVersion(v); !ok {
		return "", badRequest("INVALID_VERSION", field, "use numbers like 1.4.2")
	}
	return strings.TrimPrefix(v, "v"), nil
}

func (s *service) cleanAppVersion(ctx context.Context, q localized.Queryer, in appVersionInput) (appVersionInput, error) {
	in.Platform = strings.TrimSpace(in.Platform)
	if !validKeyType(appPlatforms, in.Platform) {
		return in, badRequest("INVALID_PLATFORM", "platform", "platform must be one of "+strings.Join(appPlatforms, ", "))
	}
	var err error
	if in.LatestVersion, err = cleanVersion("latestVersion", in.LatestVersion, true); err != nil {
		return in, err
	}
	if in.MinVersion, err = cleanVersion("minVersion", in.MinVersion, false); err != nil {
		return in, err
	}
	if in.MinVersion != "" {
		if c, _ := compareVersions(in.MinVersion, in.LatestVersion); c > 0 {
			return in, badRequest("MIN_ABOVE_LATEST", "minVersion", "the minimum version cannot be newer than the latest version")
		}
	}
	langs, err := s.languages(ctx, q)
	if err != nil {
		return in, err
	}
	if len(in.StoreLinks) > 8 {
		return in, badRequest("TOO_MANY_LINKS", "storeLinks", "at most 8 store links")
	}
	seen := map[string]bool{}
	links := make([]storeLink, 0, len(in.StoreLinks))
	for i, l := range in.StoreLinks {
		field := "storeLinks[" + itoa(i) + "]"
		l.Store = strings.TrimSpace(l.Store)
		if !validKeyType(appStores, l.Store) {
			return in, badRequest("INVALID_STORE", field+".store", "store must be one of "+strings.Join(appStores, ", "))
		}
		if seen[l.Store] {
			return in, badRequest("DUPLICATE_STORE", field+".store", "each store can be listed once")
		}
		seen[l.Store] = true
		u := strings.TrimSpace(l.URL)
		lower := strings.ToLower(u)
		if u == "" || len(u) > 1024 || strings.ContainsAny(u, " \t\r\n\"'<>`") || !(strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "market://") || strings.HasPrefix(lower, "bazaar://") || strings.HasPrefix(lower, "myket://") || strings.HasPrefix(lower, "itms-apps://")) {
			return in, badRequest("INVALID_URL", field+".url", "use an https:// or store link")
		}
		l.URL = u
		if l.Label, err = cleanText(langs, field+".label", l.Label, false, 64); err != nil {
			return in, err
		}
		links = append(links, l)
	}
	in.StoreLinks = links
	if in.UpdateMessage, err = cleanText(langs, "updateMessage", in.UpdateMessage, false, 500); err != nil {
		return in, err
	}
	if in.ForceMessage, err = cleanText(langs, "forceMessage", in.ForceMessage, false, 500); err != nil {
		return in, err
	}
	return in, nil
}

func (s *service) appVersionList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(v.platform ILIKE "+n+" OR v.latest_version ILIKE "+n+" OR v.min_version ILIKE "+n+")")
	}
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add("v.is_active = ?", v == "true")
	}
	order := "array_position(ARRAY['android','ios','web','windows','macos','linux']::text[], v.platform)"
	out, err := s.listRaw(r, ` FROM base_app_version v`, appVersionJSON, order, where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchAppVersion(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+appVersionJSON+`::text FROM base_app_version v WHERE v.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) appVersionGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchAppVersion(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func (s *service) appVersionWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in appVersionInput
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		id := pathVar(r, "id")
		if create {
			id = uuid.Must(uuid.NewV4()).String()
		} else if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			clean, err := s.cleanAppVersion(r.Context(), tx, in)
			if err != nil {
				return err
			}
			active := clean.IsActive == nil || *clean.IsActive
			links, _ := json.Marshal(clean.StoreLinks)
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_app_version (id, platform, latest_version, min_version, store_links, update_message, force_message, is_active) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
					id, clean.Platform, clean.LatestVersion, clean.MinVersion, string(links), clean.UpdateMessage.JSON(), clean.ForceMessage.JSON(), active)
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_app_version SET platform=$2, latest_version=$3, min_version=$4, store_links=$5, update_message=$6, force_message=$7, is_active=$8, updated_at=now() WHERE id=$1`,
					id, clean.Platform, clean.LatestVersion, clean.MinVersion, string(links), clean.UpdateMessage.JSON(), clean.ForceMessage.JSON(), active)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchAppVersion(r.Context(), tx, id)
			return err
		})
		if err != nil {
			if _, ok := isUniqueViolation(err); ok {
				err = fail(http.StatusConflict, "ALREADY_EXISTS", "platform", "this platform already has a version row")
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

func (s *service) appVersionDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_app_version WHERE id=$1`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func namedList(keys []string, names map[string]map[string]string) []keyType {
	out := make([]keyType, 0, len(keys))
	for _, k := range keys {
		out = append(out, keyType{Key: k, Name: names[k]})
	}
	return out
}

func (s *service) appVersionCatalog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"platforms": namedList(appPlatforms, platformNames), "stores": namedList(appStores, storeNames)})
}

func builtinName(names map[string]string, lang, def string) string {
	if v := names[lang]; v != "" {
		return v
	}
	if v := names[def]; v != "" {
		return v
	}
	return names["en"]
}

func (s *service) rpcAppVersion(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Platform string `json:"platform"`
		Version  string `json:"version"`
		Store    string `json:"store"`
		Lang     string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	platform := strings.ToLower(strings.TrimSpace(in.Platform))
	if !validKeyType(appPlatforms, platform) {
		return nil, badRequest("INVALID_PLATFORM", "platform", "platform must be one of "+strings.Join(appPlatforms, ", "))
	}
	current := strings.TrimPrefix(strings.TrimSpace(in.Version), "v")
	if _, ok := parseVersion(current); !ok {
		return nil, badRequest("INVALID_VERSION", "version", "use numbers like 1.4.2")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "platform": platform, "current": current, "status": versionOK, "latest": "", "minimum": "", "message": "", "storeUrl": "", "stores": []any{}}
	var latest, minimum, linksRaw, upd, force string
	err = s.db.QueryRowContext(ctx, `SELECT latest_version, min_version, store_links::text, update_message::text, force_message::text FROM base_app_version WHERE platform=$1 AND is_active`, platform).
		Scan(&latest, &minimum, &linksRaw, &upd, &force)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	status := versionStatus(current, latest, minimum)
	var links []storeLink
	_ = json.Unmarshal([]byte(linksRaw), &links)
	stores := make([]map[string]any, 0, len(links))
	storeURL := ""
	want := strings.TrimSpace(in.Store)
	for _, l := range links {
		name := langs.Resolve(l.Label, lang)
		if name == "" {
			name = builtinName(storeNames[l.Store], lang, langs.Default)
		}
		stores = append(stores, map[string]any{"store": l.Store, "name": name, "url": l.URL})
		if l.Store == want {
			storeURL = l.URL
		}
	}
	if storeURL == "" && len(links) > 0 {
		storeURL = links[0].URL
	}
	message := ""
	switch status {
	case versionRequired:
		message = langs.Resolve(localized.Parse(force), lang)
		if message == "" {
			message = langs.Resolve(localized.Parse(upd), lang)
		}
	case versionAvailable:
		message = langs.Resolve(localized.Parse(upd), lang)
	}
	out["status"], out["latest"], out["minimum"], out["message"], out["storeUrl"], out["stores"] = status, latest, minimum, message, storeURL, stores
	return out, nil
}
