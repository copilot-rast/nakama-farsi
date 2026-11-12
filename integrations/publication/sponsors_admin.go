package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const sponsorStateSQL = `(CASE WHEN NOT t.is_active THEN 'off' WHEN t.ends_at IS NOT NULL AND t.ends_at <= now() THEN 'ended' WHEN t.starts_at IS NOT NULL AND t.starts_at > now() THEN 'scheduled' ELSE 'live' END)`

const sponsorJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'logoUrl',t.logo_url,'color',t.color,'websiteUrl',t.website_url,
	'links',t.links,'tagIds',to_jsonb(t.tag_ids),'contactName',t.contact_name,'contactEmail',t.contact_email,'contactPhone',t.contact_phone,'notes',t.notes,
	'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,'sort',t.sort,'state',` + sponsorStateSQL + `,
	'banners',(SELECT count(*) FROM publication_banner b WHERE b.sponsor_id=t.id),
	'placements',(SELECT count(*) FROM publication_banner_list l WHERE l.sponsor_id=t.id),
	'views7',COALESCE((SELECT sum(d.views) FROM publication_sponsor_day d WHERE d.sponsor_id=t.id AND d.day > current_date - 7),0),
	'clicks7',COALESCE((SELECT sum(d.clicks) FROM publication_sponsor_day d WHERE d.sponsor_id=t.id AND d.day > current_date - 7),0),
	'views',COALESCE((SELECT sum(d.views) FROM publication_sponsor_day d WHERE d.sponsor_id=t.id),0),
	'clicks',COALESCE((SELECT sum(d.clicks) FROM publication_sponsor_day d WHERE d.sponsor_id=t.id),0),
	'createdAt',t.created_at,'updatedAt',t.updated_at)`

func sponsorResource() resource {
	return resource{
		name:    "sponsors",
		table:   "publication_sponsor",
		json:    sponsorJSON,
		from:    ` FROM publication_sponsor t`,
		search:  []string{"t.key", translatedSearch("t.name"), "t.contact_name", "t.contact_email", "t.website_url"},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean", "state": sponsorStateSQL + " = ?", "tagId": "? = ANY(t.tag_ids)", "key": "t.key = ?"},
		order:   "t.sort, t.key, t.id",
		save:    saveSponsor,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			var used bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_banner WHERE sponsor_id=$1) OR EXISTS (SELECT 1 FROM publication_banner_list WHERE sponsor_id=$1)`, id).Scan(&used); err != nil {
				return err
			}
			if used {
				return conflict("IN_USE", "banners or placements still name this sponsor; change them first")
			}
			return nil
		},
	}
}

func saveSponsor(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key          string        `json:"key"`
		Name         i18n          `json:"name"`
		Description  i18n          `json:"description"`
		LogoURL      string        `json:"logoUrl"`
		Color        string        `json:"color"`
		WebsiteURL   string        `json:"websiteUrl"`
		Links        []sponsorLink `json:"links"`
		TagIDs       []string      `json:"tagIds"`
		ContactName  string        `json:"contactName"`
		ContactEmail string        `json:"contactEmail"`
		ContactPhone string        `json:"contactPhone"`
		Notes        string        `json:"notes"`
		StartsAt     string        `json:"startsAt"`
		EndsAt       string        `json:"endsAt"`
		IsActive     *bool         `json:"isActive"`
		Sort         int           `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanBannerKey(ctx, tx, "publication_sponsor", id, create, in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 100)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	logo, err := cleanURL("logoUrl", in.LogoURL)
	if err != nil {
		return err
	}
	color := strings.ToLower(strings.TrimSpace(in.Color))
	if color != "" && !bannerColorPattern.MatchString(color) {
		return badRequest("INVALID_COLOR", "color", "use a #rrggbb colour")
	}
	website, err := cleanWebURL("websiteUrl", in.WebsiteURL)
	if err != nil {
		return err
	}
	if len(in.Links) > 8 {
		return badRequest("TOO_MANY_LINKS", "links", "at most 8 links")
	}
	links := make([]sponsorLink, 0, len(in.Links))
	for i, l := range in.Links {
		field := "links[" + strconv.Itoa(i) + "]"
		l.Type = strings.TrimSpace(l.Type)
		if !containsStr(sponsorLinkTypes, l.Type) {
			return badRequest("INVALID_LINK_TYPE", field+".type", "unknown link type")
		}
		u, err := cleanWebURL(field+".url", l.URL)
		if err != nil {
			return err
		}
		if u == "" {
			return badRequest("REQUIRED", field+".url", "link URL is required")
		}
		label, err := cleanI18n(langs, field+".label", l.Label, false, 40)
		if err != nil {
			return err
		}
		links = append(links, sponsorLink{Type: l.Type, URL: u, Label: label})
	}
	linksJSON, _ := json.Marshal(links)
	tags, err := cleanTagIDs(ctx, tx, "tagIds", in.TagIDs, "sponsor", "general")
	if err != nil {
		return err
	}
	contactName := strings.TrimSpace(in.ContactName)
	if len([]rune(contactName)) > 120 {
		return badRequest("TEXT_TOO_LONG", "contactName", "name is too long (max 120)")
	}
	email := strings.TrimSpace(in.ContactEmail)
	if email != "" {
		e, ok := normalizeEmail(email)
		if !ok {
			return badRequest("INVALID_EMAIL", "contactEmail", "must be an email address")
		}
		email = strings.ToLower(e)
	}
	phone := strings.TrimSpace(latinDigits(in.ContactPhone))
	if phone != "" {
		p, ok := normalizePhone(phone)
		if !ok {
			return badRequest("INVALID_PHONE", "contactPhone", "use 09xxxxxxxxx, +989… or an E.164 number")
		}
		phone = p
	}
	notes := strings.TrimSpace(in.Notes)
	if len([]rune(notes)) > 2000 {
		return badRequest("TEXT_TOO_LONG", "notes", "notes are too long (max 2000)")
	}
	starts, err := cleanTime("startsAt", in.StartsAt)
	if err != nil {
		return err
	}
	ends, err := cleanTime("endsAt", in.EndsAt)
	if err != nil {
		return err
	}
	if starts != nil && ends != nil && !starts.Before(*ends) {
		return badRequest("INVALID_WINDOW", "endsAt", "end must be after start")
	}
	return execSave(ctx, tx, "publication_sponsor", id, create,
		[]string{"key", "name", "description", "logo_url", "color", "website_url", "links", "tag_ids", "contact_name", "contact_email", "contact_phone", "notes", "starts_at", "ends_at", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), logo, color, website, string(linksJSON), tags, contactName, email, phone, notes, starts, ends, boolOr(in.IsActive, true), in.Sort})
}

func cleanWebURL(field, raw string) (string, error) {
	u, err := cleanURL(field, raw)
	if err != nil {
		return "", err
	}
	lower := strings.ToLower(u)
	if u != "" && !strings.HasPrefix(lower, "https://") && !strings.HasPrefix(lower, "http://") {
		return "", badRequest("INVALID_URL", field, "URL must be absolute http(s)")
	}
	return u, nil
}

func cleanSponsorID(ctx context.Context, tx *sql.Tx, raw string) (*string, error) {
	id, err := cleanUUID("sponsorId", raw, false)
	if err != nil || id == nil {
		return id, err
	}
	var ok bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_sponsor WHERE id=$1)`, *id).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return nil, badRequest("SPONSOR_NOT_FOUND", "sponsorId", "sponsor does not exist")
	}
	return id, nil
}

type sponsorReport struct {
	Sponsor    map[string]any   `json:"sponsor"`
	From       string           `json:"from"`
	To         string           `json:"to"`
	Timezone   string           `json:"timezone"`
	Days       []map[string]any `json:"days"`
	Views      int64            `json:"views"`
	Clicks     int64            `json:"clicks"`
	Players    int64            `json:"players"`
	AllViews   int64            `json:"allViews"`
	AllClicks  int64            `json:"allClicks"`
	Banners    []map[string]any `json:"banners"`
	Placements []map[string]any `json:"placements"`
}

func reportRange(r *http.Request, zone *time.Location) (time.Time, time.Time, error) {
	q := r.URL.Query()
	today := time.Now().In(zone)
	end := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if v := strings.TrimSpace(latinDigits(q.Get("to"))); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return end, end, badRequest("INVALID_DATE", "to", "use YYYY-MM-DD")
		}
		end = t
	}
	days, _ := strconv.Atoi(q.Get("days"))
	if days < 1 || days > 366 {
		days = 30
	}
	start := end.AddDate(0, 0, -days+1)
	if v := strings.TrimSpace(latinDigits(q.Get("from"))); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return end, end, badRequest("INVALID_DATE", "from", "use YYYY-MM-DD")
		}
		start = t
	}
	if start.After(end) {
		return start, end, badRequest("INVALID_DATE_RANGE", "from", "start must be on or before the end")
	}
	if end.Sub(start) > 366*24*time.Hour {
		return start, end, badRequest("INVALID_DATE_RANGE", "from", "at most 366 days")
	}
	return start, end, nil
}

func (s *service) buildSponsorReport(r *http.Request) (*sponsorReport, error) {
	ctx := r.Context()
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	res := sponsorResource()
	raw, err := s.fetchResource(ctx, s.db, res, id)
	if err != nil {
		return nil, err
	}
	zone := projectZone(ctx, s.db)
	start, end, err := reportRange(r, zone)
	if err != nil {
		return nil, err
	}
	from, to := start.Format("2006-01-02"), end.Format("2006-01-02")
	out := &sponsorReport{From: from, To: to, Timezone: zone.String(), Days: []map[string]any{}, Banners: []map[string]any{}, Placements: []map[string]any{}}
	_ = json.Unmarshal(raw, &out.Sponsor)
	rows, err := s.db.QueryContext(ctx, `SELECT day::text, views, clicks FROM publication_sponsor_day WHERE sponsor_id=$1 AND day BETWEEN $2::date AND $3::date`, id, from, to)
	if err != nil {
		return nil, err
	}
	got := map[string][2]int64{}
	for rows.Next() {
		var d string
		var v, c int64
		if err := rows.Scan(&d, &v, &c); err != nil {
			rows.Close()
			return nil, err
		}
		got[d] = [2]int64{v, c}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	players := map[string]int64{}
	prow, err := s.db.QueryContext(ctx, `SELECT day::text, count(*) FILTER (WHERE viewed_at IS NOT NULL) FROM publication_sponsor_seen WHERE sponsor_id=$1 AND day BETWEEN $2::date AND $3::date GROUP BY day`, id, from, to)
	if err != nil {
		return nil, err
	}
	for prow.Next() {
		var d string
		var n int64
		if err := prow.Scan(&d, &n); err != nil {
			prow.Close()
			return nil, err
		}
		players[d] = n
	}
	prow.Close()
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		k := d.Format("2006-01-02")
		x := got[k]
		out.Views += x[0]
		out.Clicks += x[1]
		out.Days = append(out.Days, map[string]any{"day": k, "views": x[0], "clicks": x[1], "players": players[k]})
	}
	if err := s.db.QueryRowContext(ctx, `SELECT count(DISTINCT user_id) FROM publication_sponsor_seen WHERE sponsor_id=$1 AND day BETWEEN $2::date AND $3::date AND viewed_at IS NOT NULL`, id, from, to).Scan(&out.Players); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(sum(views),0), COALESCE(sum(clicks),0) FROM publication_sponsor_day WHERE sponsor_id=$1`, id).Scan(&out.AllViews, &out.AllClicks); err != nil {
		return nil, err
	}
	brows, err := s.db.QueryContext(ctx, `SELECT jsonb_build_object('id',b.id,'key',b.key,'title',b.title,'listKey',l.key,'isActive',b.is_active,
		'views',COALESCE((SELECT sum(d.views) FROM publication_banner_day d WHERE d.banner_id=b.id AND d.day BETWEEN $2::date AND $3::date),0),
		'clicks',COALESCE((SELECT sum(d.clicks) FROM publication_banner_day d WHERE d.banner_id=b.id AND d.day BETWEEN $2::date AND $3::date),0))::text
		FROM publication_banner b JOIN publication_banner_list l ON l.id=b.list_id WHERE b.sponsor_id=$1 ORDER BY b.key`, id, from, to)
	if err != nil {
		return nil, err
	}
	items, err := rawRows(brows)
	if err != nil {
		return nil, err
	}
	for _, it := range items {
		var m map[string]any
		_ = json.Unmarshal(it, &m)
		out.Banners = append(out.Banners, m)
	}
	lrows, err := s.db.QueryContext(ctx, `SELECT jsonb_build_object('id',l.id,'key',l.key,'name',l.name,'isActive',l.is_active)::text FROM publication_banner_list l WHERE l.sponsor_id=$1 ORDER BY l.key`, id)
	if err != nil {
		return nil, err
	}
	lists, err := rawRows(lrows)
	if err != nil {
		return nil, err
	}
	for _, it := range lists {
		var m map[string]any
		_ = json.Unmarshal(it, &m)
		out.Placements = append(out.Placements, m)
	}
	return out, nil
}

func (s *service) consoleSponsorReport(w http.ResponseWriter, r *http.Request) {
	rep, err := s.buildSponsorReport(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

func ctrPercent(views, clicks int64) string {
	if views == 0 {
		return "0"
	}
	return strconv.FormatFloat(float64(clicks)*100/float64(views), 'f', 1, 64)
}

func csvSafe(v string) string {
	if v != "" && strings.ContainsRune("=+-@\t\r", rune(v[0])) {
		return "'" + v
	}
	return v
}

func (s *service) consoleSponsorReportCSV(w http.ResponseWriter, r *http.Request) {
	rep, err := s.buildSponsorReport(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	key, _ := rep.Sponsor["key"].(string)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", `attachment; filename="sponsor-`+key+`-`+rep.From+`-`+rep.To+`.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("\xef\xbb\xbf"))
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"day", "views", "clicks", "players", "ctr_percent"})
	for _, d := range rep.Days {
		v, c, p := d["views"].(int64), d["clicks"].(int64), d["players"].(int64)
		_ = cw.Write([]string{d["day"].(string), strconv.FormatInt(v, 10), strconv.FormatInt(c, 10), strconv.FormatInt(p, 10), ctrPercent(v, c)})
	}
	_ = cw.Write([]string{"total", strconv.FormatInt(rep.Views, 10), strconv.FormatInt(rep.Clicks, 10), strconv.FormatInt(rep.Players, 10), ctrPercent(rep.Views, rep.Clicks)})
	if len(rep.Banners) > 0 {
		_ = cw.Write([]string{})
		_ = cw.Write([]string{"banner", "views", "clicks", "placement", "ctr_percent"})
		for _, b := range rep.Banners {
			v, c := toInt64(b["views"]), toInt64(b["clicks"])
			bk, _ := b["key"].(string)
			lk, _ := b["listKey"].(string)
			_ = cw.Write([]string{csvSafe(bk), strconv.FormatInt(v, 10), strconv.FormatInt(c, 10), csvSafe(lk), ctrPercent(v, c)})
		}
	}
	cw.Flush()
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		x, _ := n.Int64()
		return x
	}
	return 0
}
