package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const bannerStateSQL = `(CASE WHEN NOT t.is_active THEN 'off' WHEN t.ends_at IS NOT NULL AND t.ends_at <= now() THEN 'ended' WHEN t.starts_at IS NOT NULL AND t.starts_at > now() THEN 'scheduled' ELSE 'live' END)`

const bannerJSON = `jsonb_build_object('id',t.id,'key',t.key,'sponsorId',t.sponsor_id,'sponsorKey',(SELECT sp.key FROM publication_sponsor sp WHERE sp.id=t.sponsor_id),'sponsorName',(SELECT sp.name FROM publication_sponsor sp WHERE sp.id=t.sponsor_id),'listId',t.list_id,'listKey',l.key,'listName',l.name,'listActive',l.is_active,'name',t.name,'tagIds',to_jsonb(t.tag_ids),
	'title',t.title,'subtitle',t.subtitle,'description',t.description,'imageUrl',t.image_url,'bgColor',t.bg_color,
	'action',jsonb_build_object('type',t.action_type,'value',t.action_value,'label',t.action_label,'url',t.action_url),
	'priority',t.priority,'startsAt',t.starts_at,'endsAt',t.ends_at,'days',to_jsonb(t.days),
	'dayStart',CASE WHEN t.day_start IS NULL THEN NULL ELSE lpad((t.day_start/60)::text,2,'0')||':'||lpad((t.day_start%60)::text,2,'0') END,
	'dayEnd',CASE WHEN t.day_end IS NULL THEN NULL ELSE lpad((t.day_end/60)::text,2,'0')||':'||lpad((t.day_end%60)::text,2,'0') END,
	'audienceTagIds',to_jsonb(t.audience_tag_ids),'excludeTagIds',to_jsonb(t.exclude_tag_ids),'levelTableId',t.level_table_id,'minLevel',t.min_level,'maxLevel',t.max_level,
	'platforms',to_jsonb(t.platforms),'minVersion',t.min_version,'maxVersion',t.max_version,'countries',to_jsonb(t.countries),'languages',to_jsonb(t.languages),
	'playerAge',t.player_age,'newDays',t.new_days,'isActive',t.is_active,'state',` + bannerStateSQL + `,
	'views',COALESCE((SELECT sum(d.views) FROM publication_banner_day d WHERE d.banner_id=t.id),0),'clicks',COALESCE((SELECT sum(d.clicks) FROM publication_banner_day d WHERE d.banner_id=t.id),0),
	'views7',COALESCE((SELECT sum(d.views) FROM publication_banner_day d WHERE d.banner_id=t.id AND d.day > current_date - 7),0),
	'clicks7',COALESCE((SELECT sum(d.clicks) FROM publication_banner_day d WHERE d.banner_id=t.id AND d.day > current_date - 7),0),
	'createdAt',t.created_at,'updatedAt',t.updated_at)`

const bannerListJSON = `jsonb_build_object('id',t.id,'key',t.key,'sponsorId',t.sponsor_id,'sponsorName',(SELECT sp.name FROM publication_sponsor sp WHERE sp.id=t.sponsor_id),'name',t.name,'description',t.description,'layout',t.layout,'maxItems',t.max_items,'rotateSeconds',t.rotate_seconds,
	'isActive',t.is_active,'sort',t.sort,'banners',(SELECT count(*) FROM publication_banner b WHERE b.list_id=t.id),
	'liveBanners',(SELECT count(*) FROM publication_banner b WHERE b.list_id=t.id AND b.is_active AND (b.starts_at IS NULL OR b.starts_at <= now()) AND (b.ends_at IS NULL OR b.ends_at > now())),
	'views7',COALESCE((SELECT sum(d.views) FROM publication_banner_day d JOIN publication_banner b ON b.id=d.banner_id WHERE b.list_id=t.id AND d.day > current_date - 7),0),
	'clicks7',COALESCE((SELECT sum(d.clicks) FROM publication_banner_day d JOIN publication_banner b ON b.id=d.banner_id WHERE b.list_id=t.id AND d.day > current_date - 7),0),
	'items',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',b.id,'key',b.key,'title',b.title,'imageUrl',b.image_url,'priority',b.priority,'isActive',b.is_active,
		'state',(CASE WHEN NOT b.is_active THEN 'off' WHEN b.ends_at IS NOT NULL AND b.ends_at <= now() THEN 'ended' WHEN b.starts_at IS NOT NULL AND b.starts_at > now() THEN 'scheduled' ELSE 'live' END))
		ORDER BY b.priority DESC, b.created_at, b.id),'[]'::jsonb) FROM publication_banner b WHERE b.list_id=t.id),
	'createdAt',t.created_at,'updatedAt',t.updated_at)`

func bannerResources() []resource {
	return []resource{
		{
			name:    "banner-lists",
			table:   "publication_banner_list",
			json:    bannerListJSON,
			from:    ` FROM publication_banner_list t`,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "layout": "t.layout = ?", "key": "t.key = ?", "sponsorId": "t.sponsor_id = ?::uuid"},
			order:   "t.sort, t.key, t.id",
			save:    saveBannerList,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				var used bool
				if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_banner WHERE list_id=$1)`, id).Scan(&used); err != nil {
					return err
				}
				if used {
					return conflict("IN_USE", "this placement still has banners; move or delete them first")
				}
				return nil
			},
		},
		{
			name:   "banners",
			table:  "publication_banner",
			json:   bannerJSON,
			from:   ` FROM publication_banner t JOIN publication_banner_list l ON l.id=t.list_id`,
			search: []string{"t.key", "t.name", translatedSearch("t.title"), "l.key", translatedSearch("l.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "listId": "t.list_id = ?::uuid", "state": bannerStateSQL + " = ?",
				"actionType": "t.action_type = ?", "tagId": "? = ANY(t.tag_ids)", "sponsorId": "t.sponsor_id = ?::uuid"},
			order: "l.sort, l.key, t.priority DESC, t.created_at, t.id",
			save:  saveBanner,
		},
	}
}

func saveBannerList(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key           string `json:"key"`
		Name          i18n   `json:"name"`
		Description   i18n   `json:"description"`
		Layout        string `json:"layout"`
		MaxItems      *int   `json:"maxItems"`
		RotateSeconds *int   `json:"rotateSeconds"`
		IsActive      *bool  `json:"isActive"`
		Sort          int    `json:"sort"`
		SponsorID     string `json:"sponsorId"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanBannerKey(ctx, tx, "publication_banner_list", id, create, in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 100)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 500)
	if err != nil {
		return err
	}
	if in.Layout == "" {
		in.Layout = "carousel"
	}
	if !containsStr(bannerLayouts, in.Layout) {
		return badRequest("INVALID_LAYOUT", "layout", "layout must be carousel, single, grid or strip")
	}
	maxItems, rotate := 5, 5
	if in.MaxItems != nil {
		maxItems = *in.MaxItems
	}
	if in.RotateSeconds != nil {
		rotate = *in.RotateSeconds
	}
	if maxItems < 1 || maxItems > 20 {
		return badRequest("INVALID_MAX_ITEMS", "maxItems", "between 1 and 20")
	}
	if rotate < 0 || rotate > 60 {
		return badRequest("INVALID_ROTATE", "rotateSeconds", "between 0 and 60 seconds")
	}
	sponsor, err := cleanSponsorID(ctx, tx, in.SponsorID)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "publication_banner_list", id, create, []string{"key", "name", "description", "layout", "max_items", "rotate_seconds", "is_active", "sort", "sponsor_id"},
		[]any{key, i18nJSON(name), i18nJSON(desc), in.Layout, maxItems, rotate, boolOr(in.IsActive, true), in.Sort, sponsor})
}

func cleanBannerKey(ctx context.Context, tx *sql.Tx, table, id string, create bool, raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if !bannerKeyPattern.MatchString(key) {
		return "", badRequest("INVALID_KEY", "key", "key must start with a letter and use a-z, 0-9 and _ (2-64)")
	}
	if !create {
		var old string
		err := tx.QueryRowContext(ctx, `SELECT key FROM `+table+` WHERE id=$1`, id).Scan(&old)
		if errors.Is(err, sql.ErrNoRows) {
			return "", notFound("NOT_FOUND")
		}
		if err != nil {
			return "", err
		}
		if old != key {
			return "", badRequest("KEY_LOCKED", "key", "the key cannot change after creation")
		}
		return key, nil
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE key=$1)`, key).Scan(&exists); err != nil {
		return "", err
	}
	if exists {
		return "", conflict("KEY_EXISTS", "this key is already used")
	}
	return key, nil
}

func cleanURLs(langs *localized.Set, field string, in i18n, required bool) (i18n, error) {
	out, err := cleanI18n(langs, field, in, required, 1024)
	if err != nil {
		return nil, err
	}
	for code, v := range out {
		if !safeURL(v) {
			return nil, badRequest("INVALID_URL", field, "URL for "+code+" must be absolute http(s) or start with /")
		}
	}
	return out, nil
}

func cleanTagIDs(ctx context.Context, tx *sql.Tx, field string, ids []string, kinds ...string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	if len(ids) > 50 {
		return nil, badRequest("TOO_MANY_TAGS", field, "at most 50 tags")
	}
	for _, raw := range ids {
		id, err := cleanUUID(field, raw, true)
		if err != nil {
			return nil, err
		}
		if seen[*id] {
			continue
		}
		seen[*id] = true
		var kt string
		err = tx.QueryRowContext(ctx, `SELECT key_type FROM base_tag WHERE id=$1`, *id).Scan(&kt)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, badRequest("TAG_NOT_FOUND", field, "tag does not exist")
		}
		if err != nil {
			return nil, err
		}
		if !containsStr(kinds, kt) {
			return nil, badRequest("TAG_KIND_MISMATCH", field, "tag key type must be "+strings.Join(kinds, " or "))
		}
		out = append(out, *id)
	}
	return out, nil
}

func cleanList(field string, in []string, max int, norm func(string) string, ok func(string) bool, code string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	if len(in) > max {
		return nil, badRequest(code, field, "too many values")
	}
	for _, v := range in {
		v = norm(strings.TrimSpace(v))
		if v == "" || seen[v] {
			continue
		}
		if !ok(v) {
			return nil, badRequest(code, field, "unknown value "+v)
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Strings(out)
	return out, nil
}

func saveBanner(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string   `json:"key"`
		ListID      string   `json:"listId"`
		Name        string   `json:"name"`
		TagIDs      []string `json:"tagIds"`
		Title       i18n     `json:"title"`
		Subtitle    i18n     `json:"subtitle"`
		Description i18n     `json:"description"`
		ImageURL    i18n     `json:"imageUrl"`
		BgColor     string   `json:"bgColor"`
		Action      struct {
			Type  string `json:"type"`
			Value string `json:"value"`
			Label i18n   `json:"label"`
			URL   i18n   `json:"url"`
		} `json:"action"`
		Priority       int      `json:"priority"`
		StartsAt       string   `json:"startsAt"`
		EndsAt         string   `json:"endsAt"`
		Days           []int    `json:"days"`
		DayStart       string   `json:"dayStart"`
		DayEnd         string   `json:"dayEnd"`
		AudienceTagIDs []string `json:"audienceTagIds"`
		ExcludeTagIDs  []string `json:"excludeTagIds"`
		LevelTableID   string   `json:"levelTableId"`
		MinLevel       *int     `json:"minLevel"`
		MaxLevel       *int     `json:"maxLevel"`
		Platforms      []string `json:"platforms"`
		MinVersion     string   `json:"minVersion"`
		MaxVersion     string   `json:"maxVersion"`
		Countries      []string `json:"countries"`
		Languages      []string `json:"languages"`
		PlayerAge      string   `json:"playerAge"`
		NewDays        *int     `json:"newDays"`
		IsActive       *bool    `json:"isActive"`
		SponsorID      string   `json:"sponsorId"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanBannerKey(ctx, tx, "publication_banner", id, create, in.Key)
	if err != nil {
		return err
	}
	listID, err := cleanUUID("listId", in.ListID, true)
	if err != nil {
		return err
	}
	var listOK bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_banner_list WHERE id=$1)`, *listID).Scan(&listOK); err != nil {
		return err
	}
	if !listOK {
		return badRequest("BANNER_LIST_NOT_FOUND", "listId", "placement does not exist")
	}
	name := strings.TrimSpace(in.Name)
	if len([]rune(name)) > 120 {
		return badRequest("TEXT_TOO_LONG", "name", "name is too long (max 120)")
	}
	tags, err := cleanTagIDs(ctx, tx, "tagIds", in.TagIDs, "banner", "general")
	if err != nil {
		return err
	}
	title, err := cleanI18n(langs, "title", in.Title, true, 120)
	if err != nil {
		return err
	}
	subtitle, err := cleanI18n(langs, "subtitle", in.Subtitle, false, 200)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	image, err := cleanURLs(langs, "imageUrl", in.ImageURL, false)
	if err != nil {
		return err
	}
	color := strings.ToLower(strings.TrimSpace(in.BgColor))
	if color != "" && !bannerColorPattern.MatchString(color) {
		return badRequest("INVALID_COLOR", "bgColor", "use a #rrggbb colour")
	}
	act := in.Action
	act.Type = strings.TrimSpace(act.Type)
	act.Value = strings.TrimSpace(act.Value)
	if act.Type == "" {
		act.Type = "none"
	}
	if !containsStr(bannerActions, act.Type) {
		return badRequest("INVALID_ACTION", "action.type", "type must be none, screen, product, achievement, url or sponsor")
	}
	sponsor, err := cleanSponsorID(ctx, tx, in.SponsorID)
	if err != nil {
		return err
	}
	label, err := cleanI18n(langs, "action.label", act.Label, false, 40)
	if err != nil {
		return err
	}
	urls := i18n{}
	switch act.Type {
	case "none":
		act.Value, label = "", i18n{}
	case "screen":
		if !bannerScreenPattern.MatchString(act.Value) {
			return badRequest("INVALID_ACTION", "action.value", "screen names use a-z, 0-9, _ and - (max 40)")
		}
	case "product", "achievement":
		table := "economy_product"
		if act.Type == "achievement" {
			table = "economy_achievement"
		}
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE key=$1)`, act.Value).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("ACTION_TARGET_NOT_FOUND", "action.value", act.Type+" does not exist")
		}
	case "sponsor":
		act.Value = ""
		if sponsor == nil {
			return badRequest("SPONSOR_REQUIRED", "sponsorId", "pick the sponsor whose website the button opens")
		}
		var site string
		if err := tx.QueryRowContext(ctx, `SELECT website_url FROM publication_sponsor WHERE id=$1`, *sponsor).Scan(&site); err != nil {
			return err
		}
		if site == "" {
			return badRequest("SPONSOR_NO_WEBSITE", "action.type", "the sponsor has no website")
		}
	case "url":
		act.Value = ""
		urls, err = cleanURLs(langs, "action.url", act.URL, true)
		if err != nil {
			return err
		}
		for code, v := range urls {
			lower := strings.ToLower(v)
			if !strings.HasPrefix(lower, "https://") && !strings.HasPrefix(lower, "http://") {
				return badRequest("INVALID_URL", "action.url", "URL for "+code+" must be absolute http(s)")
			}
		}
	}
	if in.Priority < -1000 || in.Priority > 1000 {
		return badRequest("INVALID_PRIORITY", "priority", "between -1000 and 1000")
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
	days := []int{}
	seenDay := map[int]bool{}
	for _, d := range in.Days {
		if d < 0 || d > 6 {
			return badRequest("INVALID_WEEKDAY", "days", "days are 0 (Sunday) to 6 (Saturday)")
		}
		if !seenDay[d] {
			seenDay[d] = true
			days = append(days, d)
		}
	}
	sort.Ints(days)
	if len(days) == 7 {
		days = []int{}
	}
	dayStart, err := parseMinutes("dayStart", in.DayStart, false)
	if err != nil {
		return err
	}
	dayEnd, err := parseMinutes("dayEnd", in.DayEnd, true)
	if err != nil {
		return err
	}
	if (dayStart == nil) != (dayEnd == nil) {
		field := "dayEnd"
		if dayStart == nil {
			field = "dayStart"
		}
		return badRequest("INVALID_HOURS", field, "give both a start and an end time")
	}
	if dayStart != nil && *dayStart >= *dayEnd {
		return badRequest("INVALID_HOURS", "dayEnd", "end must be after start")
	}
	audience, err := cleanTagIDs(ctx, tx, "audienceTagIds", in.AudienceTagIDs, "user")
	if err != nil {
		return err
	}
	exclude, err := cleanTagIDs(ctx, tx, "excludeTagIds", in.ExcludeTagIDs, "user")
	if err != nil {
		return err
	}
	for _, x := range exclude {
		if containsStr(audience, x) {
			return badRequest("TAG_BOTH_WAYS", "excludeTagIds", "a tag cannot be both shown to and hidden from")
		}
	}
	levelTable, err := cleanUUID("levelTableId", in.LevelTableID, false)
	if err != nil {
		return err
	}
	if levelTable != nil {
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_level_table WHERE id=$1)`, *levelTable).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("LEVEL_TABLE_NOT_FOUND", "levelTableId", "level table does not exist")
		}
	}
	for _, p := range []struct {
		f string
		v *int
	}{{"minLevel", in.MinLevel}, {"maxLevel", in.MaxLevel}} {
		if p.v != nil && (*p.v < 0 || *p.v > 100000) {
			return badRequest("INVALID_LEVEL", p.f, "between 0 and 100000")
		}
	}
	if in.MinLevel != nil && in.MaxLevel != nil && *in.MinLevel > *in.MaxLevel {
		return badRequest("INVALID_RANGE", "maxLevel", "max level is below min level")
	}
	if in.MinLevel == nil && in.MaxLevel == nil {
		levelTable = nil
	}
	platforms, err := cleanList("platforms", in.Platforms, 10, strings.ToLower, func(v string) bool { return containsStr(bannerPlatforms, v) }, "INVALID_PLATFORM")
	if err != nil {
		return err
	}
	minV := strings.TrimPrefix(strings.TrimSpace(latinDigits(in.MinVersion)), "v")
	maxV := strings.TrimPrefix(strings.TrimSpace(latinDigits(in.MaxVersion)), "v")
	for _, p := range []struct{ f, v string }{{"minVersion", minV}, {"maxVersion", maxV}} {
		if p.v != "" {
			if _, ok := bannerVersionParts(p.v); !ok || len(p.v) > 40 {
				return badRequest("INVALID_VERSION", p.f, "use a version like 1.4.0")
			}
		}
	}
	if minV != "" && maxV != "" {
		if c, _ := compareBannerVersions(minV, maxV); c > 0 {
			return badRequest("INVALID_RANGE", "maxVersion", "max version is below min version")
		}
	}
	countries, err := cleanList("countries", in.Countries, 250, strings.ToUpper, func(v string) bool {
		return len(v) == 2 && v[0] >= 'A' && v[0] <= 'Z' && v[1] >= 'A' && v[1] <= 'Z'
	}, "INVALID_COUNTRY_CODE")
	if err != nil {
		return err
	}
	languages, err := cleanList("languages", in.Languages, 50, localized.Normalize, langs.Known, "INVALID_LANGUAGE")
	if err != nil {
		return err
	}
	if in.PlayerAge == "" {
		in.PlayerAge = "all"
	}
	if !containsStr(bannerAges, in.PlayerAge) {
		return badRequest("INVALID_PLAYER_AGE", "playerAge", "playerAge must be all, new or returning")
	}
	newDays := 7
	if in.NewDays != nil {
		newDays = *in.NewDays
	}
	if newDays < 1 || newDays > 365 {
		return badRequest("INVALID_NEW_DAYS", "newDays", "between 1 and 365 days")
	}
	return execSave(ctx, tx, "publication_banner", id, create,
		[]string{"key", "list_id", "name", "tag_ids", "title", "subtitle", "description", "image_url", "bg_color", "action_type", "action_value", "action_label", "action_url",
			"priority", "starts_at", "ends_at", "days", "day_start", "day_end", "audience_tag_ids", "exclude_tag_ids", "level_table_id", "min_level", "max_level",
			"platforms", "min_version", "max_version", "countries", "languages", "player_age", "new_days", "is_active", "sponsor_id"},
		[]any{key, *listID, name, tags, i18nJSON(title), i18nJSON(subtitle), i18nJSON(desc), i18nJSON(image), color, act.Type, act.Value, i18nJSON(label), i18nJSON(urls),
			in.Priority, starts, ends, days, dayStart, dayEnd, audience, exclude, levelTable, in.MinLevel, in.MaxLevel,
			platforms, minV, maxV, countries, languages, in.PlayerAge, newDays, boolOr(in.IsActive, true), sponsor})
}

func (s *service) consoleBannerStats(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 || days > 90 {
		days = 14
	}
	var exists bool
	if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM publication_banner WHERE id=$1)`, id).Scan(&exists); err != nil {
		s.writeError(w, err)
		return
	}
	if !exists {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	zone := projectZone(r.Context(), s.db)
	today := time.Now().In(zone)
	start := time.Date(today.Year(), today.Month(), today.Day()-days+1, 0, 0, 0, 0, time.UTC)
	rows, err := s.db.QueryContext(r.Context(), `SELECT day::text, views, clicks FROM publication_banner_day WHERE banner_id=$1 AND day >= $2::date ORDER BY day`, id, start.Format("2006-01-02"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	got := map[string][2]int64{}
	for rows.Next() {
		var d string
		var v, c int64
		if err := rows.Scan(&d, &v, &c); err != nil {
			rows.Close()
			s.writeError(w, err)
			return
		}
		got[d] = [2]int64{v, c}
	}
	rows.Close()
	series := make([]map[string]any, 0, days)
	var tv, tc int64
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i).Format("2006-01-02")
		x := got[d]
		tv += x[0]
		tc += x[1]
		series = append(series, map[string]any{"day": d, "views": x[0], "clicks": x[1]})
	}
	var allV, allC int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT COALESCE(sum(views),0), COALESCE(sum(clicks),0) FROM publication_banner_day WHERE banner_id=$1`, id).Scan(&allV, &allC); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"days": series, "views": tv, "clicks": tc, "allViews": allV, "allClicks": allC, "timezone": zone.String()})
}

func (s *service) consoleBannerSettings(w http.ResponseWriter, r *http.Request) {
	zone := projectZone(r.Context(), s.db)
	writeJSON(w, http.StatusOK, map[string]any{"timezone": zone.String(), "now": time.Now().UTC(), "platforms": bannerPlatforms, "layouts": bannerLayouts, "actions": bannerActions, "linkTypes": sponsorLinkTypes})
}

func (s *service) consoleBannerPreview(w http.ResponseWriter, r *http.Request) {
	listID := pathVar(r, "id")
	if _, err := uuid.FromString(listID); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	q := r.URL.Query()
	in := struct{ UserID, Platform, Version, Lang, At string }{strings.TrimSpace(q.Get("userId")), q.Get("platform"), q.Get("version"), q.Get("lang"), q.Get("at")}
	at, err := cleanTime("at", in.At)
	if err != nil {
		s.writeError(w, err)
		return
	}
	now := time.Now()
	if at != nil {
		now = *at
	}
	if _, err := uuid.FromString(in.UserID); err != nil {
		s.writeError(w, badRequest("INVALID_ID", "userId", "pick a player"))
		return
	}
	var exists bool
	if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, in.UserID).Scan(&exists); err != nil {
		s.writeError(w, err)
		return
	}
	if !exists {
		s.writeError(w, badRequest("USER_NOT_FOUND", "userId", "player does not exist"))
		return
	}
	v, err := s.bannerViewer(r.Context(), in.UserID, in.Platform, in.Version, in.Lang, "")
	if err != nil {
		s.writeError(w, err)
		return
	}
	v.Now = now
	list, err := s.loadPlacement(r.Context(), "id", listID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	banners, err := loadBanners(r.Context(), s.db, `WHERE b.list_id=$2`, v.tagList(), listID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	langs, lang, err := localized.ForUser(r.Context(), s.db, in.UserID, in.Lang)
	if err != nil {
		s.writeError(w, err)
		return
	}
	shown := []map[string]any{}
	hidden := []map[string]any{}
	for _, b := range banners {
		reason, err := b.eligible(r.Context(), v)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if reason == "" && len(shown) >= list.maxItems {
			reason = "max_items"
		}
		if reason == "" {
			shown = append(shown, b.view(langs, lang))
		} else {
			hidden = append(hidden, map[string]any{"id": b.ID, "key": b.Key, "title": langs.Resolve(b.Title, lang), "reason": reason})
		}
	}
	if !list.active {
		for _, x := range shown {
			hidden = append(hidden, map[string]any{"id": x["id"], "key": x["key"], "title": x["title"], "reason": "placement_off"})
		}
		shown = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"lang": lang, "timezone": v.Zone.String(), "player": map[string]any{"country": v.Country, "lang": v.Lang, "tags": v.tagList(), "createdAt": v.Created},
		"items": shown, "hidden": hidden})
}
