package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var (
	bannerKeyPattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
	bannerScreenPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,39}$`)
	bannerColorPattern  = regexp.MustCompile(`^#[0-9a-f]{6}$`)
	bannerVersion       = regexp.MustCompile(`^\d+(\.\d+){0,3}([-+][0-9A-Za-z.-]+)?$`)
	bannerLayouts       = []string{"carousel", "single", "grid", "strip"}
	bannerActions       = []string{"none", "screen", "product", "achievement", "url", "sponsor"}
	bannerPlatforms     = []string{"android", "ios", "web", "windows", "macos", "linux"}
	bannerAges          = []string{"all", "new", "returning"}
)

type bannerRow struct {
	ID          string
	Key         string
	ListID      string
	Title       i18n
	Subtitle    i18n
	Description i18n
	Image       i18n
	BgColor     string
	ActionType  string
	ActionValue string
	ActionLabel i18n
	ActionURL   i18n
	Priority    int
	StartsAt    *time.Time
	EndsAt      *time.Time
	Days        []int
	DayStart    *int
	DayEnd      *int
	audienceRule
	IsActive    bool
	TargetOK    bool
	CreatedAt   time.Time
	Sponsor     *sponsorRef
	ListSponsor *sponsorRef
}

type viewer struct {
	UserID   string
	Platform string
	Version  string
	Country  string
	Lang     string
	Tags     map[string]bool
	Created  time.Time
	Now      time.Time
	Zone     *time.Location
	levels   map[string]int
	q        queryer
}

var bannerColumns = `b.id::text, b.key, b.list_id::text, b.title::text, b.subtitle::text, b.description::text, b.image_url::text, b.bg_color, b.action_type, b.action_value,
	b.action_label::text, b.action_url::text, b.priority, b.starts_at, b.ends_at, to_jsonb(b.days)::text, b.day_start, b.day_end, to_jsonb(b.audience_tag_ids)::text, to_jsonb(b.exclude_tag_ids)::text,
	COALESCE(b.level_table_id::text,''), b.min_level, b.max_level, to_jsonb(b.platforms)::text, b.min_version, b.max_version, to_jsonb(b.countries)::text, to_jsonb(b.languages)::text, b.player_age, b.new_days, b.is_active,
	CASE b.action_type
		WHEN 'product' THEN EXISTS (SELECT 1 FROM economy_product p WHERE p.key=b.action_value AND p.is_active AND (cardinality(p.audience_tag_ids)=0 OR p.audience_tag_ids && $1::text[]))
		WHEN 'achievement' THEN EXISTS (SELECT 1 FROM economy_achievement a WHERE a.key=b.action_value AND a.is_active)
		ELSE true END,
	b.created_at, ` + sponsorRefSQL("b.sponsor_id") + `, ` + sponsorRefSQL("(SELECT l2.sponsor_id FROM publication_banner_list l2 WHERE l2.id=b.list_id)")

type rowsQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func loadBanners(ctx context.Context, q rowsQueryer, where string, tags []string, args ...any) ([]bannerRow, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+bannerColumns+` FROM publication_banner b `+where+` ORDER BY b.priority DESC, b.created_at, b.id`, append([]any{tags}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bannerRow{}
	for rows.Next() {
		var b bannerRow
		var title, sub, desc, img, label, url string
		var days, aud, exc, plat, ctry, lng string
		var dayStart, dayEnd, minLv, maxLv sql.NullInt64
		var starts, ends sql.NullTime
		var sponsor, listSponsor sql.NullString
		if err := rows.Scan(&b.ID, &b.Key, &b.ListID, &title, &sub, &desc, &img, &b.BgColor, &b.ActionType, &b.ActionValue, &label, &url, &b.Priority, &starts, &ends,
			&days, &dayStart, &dayEnd, &aud, &exc, &b.LevelTable, &minLv, &maxLv, &plat, &b.MinVersion, &b.MaxVersion, &ctry, &lng,
			&b.PlayerAge, &b.NewDays, &b.IsActive, &b.TargetOK, &b.CreatedAt, &sponsor, &listSponsor); err != nil {
			return nil, err
		}
		b.Title, b.Subtitle, b.Description, b.Image, b.ActionLabel, b.ActionURL = parseI18n(title), parseI18n(sub), parseI18n(desc), parseI18n(img), parseI18n(label), parseI18n(url)
		_ = json.Unmarshal([]byte(days), &b.Days)
		_ = json.Unmarshal([]byte(aud), &b.Audience)
		_ = json.Unmarshal([]byte(exc), &b.Exclude)
		_ = json.Unmarshal([]byte(plat), &b.Platforms)
		_ = json.Unmarshal([]byte(ctry), &b.Countries)
		_ = json.Unmarshal([]byte(lng), &b.Languages)
		if starts.Valid {
			t := starts.Time
			b.StartsAt = &t
		}
		if ends.Valid {
			t := ends.Time
			b.EndsAt = &t
		}
		b.DayStart, b.DayEnd, b.MinLevel, b.MaxLevel = nullInt(dayStart), nullInt(dayEnd), nullInt(minLv), nullInt(maxLv)
		b.Sponsor, b.ListSponsor = parseSponsorRef(sponsor), parseSponsorRef(listSponsor)
		out = append(out, b)
	}
	return out, rows.Err()
}

func nullInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}

func projectZone(ctx context.Context, q queryer) *time.Location {
	var name string
	if err := q.QueryRowContext(ctx, `SELECT timezone FROM economy_event_settings WHERE id=1`).Scan(&name); err != nil {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (v *viewer) level(ctx context.Context, table string) (int, error) {
	if v.levels == nil {
		v.levels = map[string]int{}
	}
	if lv, ok := v.levels[table]; ok {
		return lv, nil
	}
	id := table
	if id == "" {
		err := v.q.QueryRowContext(ctx, `SELECT id::text FROM economy_level_table WHERE is_active ORDER BY sort, created_at LIMIT 1`).Scan(&id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}
	var lv int
	if id == "" {
		var f sql.NullFloat64
		err := v.q.QueryRowContext(ctx, `SELECT COALESCE(um.value, m.default_value)::float8 FROM economy_meta m
			LEFT JOIN economy_user_meta um ON um.meta_id=m.id AND um.user_id=$1 WHERE m.key='level'`, v.UserID).Scan(&f)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
		lv = int(f.Float64)
	} else if err := v.q.QueryRowContext(ctx, `SELECT GREATEST(COALESCE((SELECT level FROM economy_user_level WHERE table_id=$1::uuid AND user_id=$2::uuid),0),
		COALESCE((SELECT max(s.level) FROM economy_level_step s JOIN economy_level_table lt ON lt.id=s.table_id JOIN economy_meta m ON m.id=lt.meta_id
		LEFT JOIN economy_user_meta um ON um.meta_id=lt.meta_id AND um.user_id=$2::uuid WHERE s.table_id=$1::uuid AND s.threshold <= COALESCE(um.value, m.default_value)),0))`, id, v.UserID).Scan(&lv); err != nil {
		return 0, err
	}
	v.levels[table] = lv
	return lv, nil
}

func (v *viewer) tagList() []string {
	out := make([]string, 0, len(v.Tags))
	for k := range v.Tags {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func containsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func (b bannerRow) eligible(ctx context.Context, v *viewer) (string, error) {
	if !b.IsActive {
		return "inactive", nil
	}
	if !b.TargetOK {
		return "target_missing", nil
	}
	if b.Sponsor != nil && !b.Sponsor.live(v.Now) {
		return "sponsor_off", nil
	}
	if b.ActionType == "sponsor" && (b.Sponsor == nil || b.Sponsor.Website == "") {
		return "target_missing", nil
	}
	if b.StartsAt != nil && v.Now.Before(*b.StartsAt) {
		return "not_started", nil
	}
	if b.EndsAt != nil && !v.Now.Before(*b.EndsAt) {
		return "ended", nil
	}
	local := v.Now.In(v.Zone)
	if len(b.Days) > 0 {
		hit := false
		for _, d := range b.Days {
			if d == int(local.Weekday()) {
				hit = true
			}
		}
		if !hit {
			return "day", nil
		}
	}
	if b.DayStart != nil && b.DayEnd != nil {
		m := local.Hour()*60 + local.Minute()
		if m < *b.DayStart || m >= *b.DayEnd {
			return "hours", nil
		}
	}
	return b.audienceRule.match(ctx, v)
}

func compareBannerVersions(a, b string) (int, bool) {
	pa, ok1 := bannerVersionParts(a)
	pb, ok2 := bannerVersionParts(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	for i := 0; i < 4; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1, true
			}
			return 1, true
		}
	}
	return 0, true
}

func bannerVersionParts(v string) ([4]int64, bool) {
	var out [4]int64
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if !bannerVersion.MatchString(v) {
		return out, false
	}
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	for i, p := range strings.Split(v, ".") {
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil || i > 3 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

func (b bannerRow) view(langs *localized.Set, lang string) map[string]any {
	action := map[string]any{"type": b.ActionType}
	if b.ActionType != "none" {
		if b.ActionType != "url" && b.ActionType != "sponsor" {
			action["value"] = b.ActionValue
		}
		if l := langs.Resolve(b.ActionLabel, lang); l != "" {
			action["label"] = l
		}
		if b.ActionType == "url" {
			action["url"] = langs.Resolve(b.ActionURL, lang)
		}
		if b.ActionType == "sponsor" && b.Sponsor != nil {
			action["url"] = b.Sponsor.Website
			action["sponsorId"] = b.Sponsor.ID
		}
	}
	var sponsor any
	if b.Sponsor != nil {
		sponsor = b.Sponsor.view(langs, lang)
	}
	return map[string]any{"sponsor": sponsor,
		"id": b.ID, "key": b.Key, "title": langs.Resolve(b.Title, lang), "subtitle": langs.Resolve(b.Subtitle, lang),
		"description": langs.Resolve(b.Description, lang), "imageUrl": langs.Resolve(b.Image, lang), "bgColor": b.BgColor,
		"priority": b.Priority, "action": action, "endsAt": b.EndsAt,
	}
}

func minutesText(m *int) any {
	if m == nil {
		return nil
	}
	return strconv.Itoa(*m/60/10) + strconv.Itoa(*m/60%10) + ":" + strconv.Itoa(*m%60/10) + strconv.Itoa(*m%10)
}

func parseMinutes(field, s string, end bool) (*int, error) {
	s = strings.TrimSpace(latinDigits(s))
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 || len(parts[0]) < 1 || len(parts[0]) > 2 || len(parts[1]) != 2 {
		return nil, badRequest("INVALID_TIME_OF_DAY", field, "use HH:MM")
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || m < 0 || m > 59 || h < 0 || h > 24 || (h == 24 && (m != 0 || !end)) {
		return nil, badRequest("INVALID_TIME_OF_DAY", field, "use HH:MM between 00:00 and 24:00")
	}
	n := h*60 + m
	if !end && n > 1439 {
		return nil, badRequest("INVALID_TIME_OF_DAY", field, "start must be before 24:00")
	}
	if end && n == 0 {
		return nil, badRequest("INVALID_TIME_OF_DAY", field, "end must be after 00:00")
	}
	return &n, nil
}
