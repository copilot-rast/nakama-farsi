package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	sponsorListPerMinute  = 60
	sponsorCountPerMinute = 60
)

var sponsorLinkTypes = []string{"website", "instagram", "telegram", "x", "youtube", "aparat", "whatsapp", "app", "other"}

type sponsorLink struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Label i18n   `json:"label,omitempty"`
}

type sponsorRef struct {
	ID       string     `json:"id"`
	Key      string     `json:"key"`
	Name     i18n       `json:"name"`
	Logo     string     `json:"logoUrl"`
	Color    string     `json:"color"`
	Website  string     `json:"websiteUrl"`
	IsActive bool       `json:"isActive"`
	StartsAt *time.Time `json:"startsAt"`
	EndsAt   *time.Time `json:"endsAt"`
}

func sponsorRefSQL(col string) string {
	return `(SELECT jsonb_build_object('id',sp.id,'key',sp.key,'name',sp.name,'logoUrl',sp.logo_url,'color',sp.color,'websiteUrl',sp.website_url,
		'isActive',sp.is_active,'startsAt',sp.starts_at,'endsAt',sp.ends_at)::text FROM publication_sponsor sp WHERE sp.id=` + col + `)`
}

func parseSponsorRef(raw sql.NullString) *sponsorRef {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	var r sponsorRef
	if err := json.Unmarshal([]byte(raw.String), &r); err != nil || r.ID == "" {
		return nil
	}
	return &r
}

func (r *sponsorRef) live(now time.Time) bool {
	if r == nil || !r.IsActive {
		return false
	}
	if r.StartsAt != nil && now.Before(*r.StartsAt) {
		return false
	}
	if r.EndsAt != nil && !now.Before(*r.EndsAt) {
		return false
	}
	return true
}

func (r *sponsorRef) view(langs *localized.Set, lang string) map[string]any {
	return map[string]any{"id": r.ID, "key": r.Key, "name": langs.Resolve(r.Name, lang), "logoUrl": r.Logo, "color": r.Color, "websiteUrl": r.Website}
}

func sponsorState(active bool, starts, ends *time.Time, now time.Time) string {
	switch {
	case !active:
		return "off"
	case ends != nil && !now.Before(*ends):
		return "ended"
	case starts != nil && now.Before(*starts):
		return "scheduled"
	}
	return "live"
}

type sponsorFull struct {
	sponsorRef
	Description i18n
	Links       []sponsorLink
	Sort        int
}

func (s *service) liveSponsors(ctx context.Context, where string, args ...any) ([]sponsorFull, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text, description::text, logo_url, color, website_url, links::text, is_active, starts_at, ends_at, sort
		FROM publication_sponsor WHERE is_active AND (starts_at IS NULL OR starts_at <= now()) AND (ends_at IS NULL OR ends_at > now())`+where+` ORDER BY sort, key, id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []sponsorFull{}
	for rows.Next() {
		var sp sponsorFull
		var name, desc, links string
		var starts, ends sql.NullTime
		if err := rows.Scan(&sp.ID, &sp.Key, &name, &desc, &sp.Logo, &sp.Color, &sp.Website, &links, &sp.IsActive, &starts, &ends, &sp.Sort); err != nil {
			return nil, err
		}
		sp.Name, sp.Description = parseI18n(name), parseI18n(desc)
		_ = json.Unmarshal([]byte(links), &sp.Links)
		if starts.Valid {
			t := starts.Time
			sp.StartsAt = &t
		}
		if ends.Valid {
			t := ends.Time
			sp.EndsAt = &t
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (sp sponsorFull) clientView(langs *localized.Set, lang string) map[string]any {
	v := sp.view(langs, lang)
	v["description"] = langs.Resolve(sp.Description, lang)
	links := []map[string]any{}
	for _, l := range sp.Links {
		item := map[string]any{"type": l.Type, "url": l.URL}
		if label := langs.Resolve(l.Label, lang); label != "" {
			item["label"] = label
		}
		links = append(links, item)
	}
	v["links"] = links
	v["endsAt"] = sp.EndsAt
	return v
}

func (s *service) countSponsors(ctx context.Context, q rowsQueryer, ids []string, userID, day string, click bool) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	col, counter := "viewed_at", "views"
	if click {
		col, counter = "clicked_at", "clicks"
	}
	rows, err := q.QueryContext(ctx, `WITH s AS (
		INSERT INTO publication_sponsor_seen (sponsor_id, user_id, day, `+col+`) SELECT x::uuid, $2, $3::date, now() FROM unnest($1::text[]) x
		ON CONFLICT (sponsor_id, user_id, day) DO UPDATE SET `+col+` = now() WHERE publication_sponsor_seen.`+col+` IS NULL
		RETURNING sponsor_id)
		INSERT INTO publication_sponsor_day (sponsor_id, day, `+counter+`) SELECT sponsor_id, $3::date, 1 FROM s
		ON CONFLICT (sponsor_id, day) DO UPDATE SET `+counter+` = publication_sponsor_day.`+counter+` + 1 RETURNING sponsor_id::text`, ids, userID, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *service) clientDay(ctx context.Context) (string, time.Time) {
	now := time.Now()
	return now.In(projectZone(ctx, s.db)).Format("2006-01-02"), now
}

func (s *service) rpcSponsors(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang   string `json:"lang"`
		TagKey string `json:"tagKey"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if err := s.bannerLimit(ctx, "spn:l:"+userID, sponsorListPerMinute); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	where, args := "", []any{}
	if key := strings.TrimSpace(in.TagKey); key != "" {
		if !bannerKeyPattern.MatchString(key) && len(key) > 64 {
			return nil, badRequest("INVALID_TAG", "tagKey", "tag key is not valid")
		}
		where = ` AND EXISTS (SELECT 1 FROM base_tag bt WHERE bt.key=$1 AND bt.id::text = ANY(publication_sponsor.tag_ids))`
		args = append(args, key)
	}
	list, err := s.liveSponsors(ctx, where, args...)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(list))
	for _, sp := range list {
		items = append(items, sp.clientView(langs, lang))
	}
	return map[string]any{"lang": lang, "items": items}, nil
}

func (s *service) liveSponsorIDs(ctx context.Context, ids []string) (map[string]sponsorFull, error) {
	list, err := s.liveSponsors(ctx, ` AND id = ANY($1::uuid[])`, ids)
	if err != nil {
		return nil, err
	}
	out := map[string]sponsorFull{}
	for _, sp := range list {
		out[sp.ID] = sp
	}
	return out, nil
}

func cleanSponsorIDs(field string, ids []string, max int) ([]string, error) {
	out, err := cleanBannerIDs(field, ids, max)
	if app, ok := asApp(err); ok && app.Code == "TOO_MANY_BANNERS" {
		return nil, badRequest("TOO_MANY_SPONSORS", field, app.Message)
	}
	return out, err
}

func (s *service) rpcSponsorView(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		SponsorIDs []string `json:"sponsorIds"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	ids, err := cleanSponsorIDs("sponsorIds", in.SponsorIDs, 20)
	if err != nil {
		return nil, err
	}
	if err := s.bannerLimit(ctx, "spn:v:"+userID, sponsorCountPerMinute); err != nil {
		return nil, err
	}
	live, err := s.liveSponsorIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	valid, skipped := []string{}, []string{}
	for _, id := range ids {
		if _, ok := live[id]; ok {
			valid = append(valid, id)
		} else {
			skipped = append(skipped, id)
		}
	}
	day, _ := s.clientDay(ctx)
	counted, err := s.countSponsors(ctx, s.db, valid, userID, day, false)
	if err != nil {
		return nil, err
	}
	return map[string]any{"counted": counted, "skipped": skipped, "day": day}, nil
}

func (s *service) rpcSponsorClick(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		SponsorID string `json:"sponsorId"`
		Link      string `json:"link"`
		Lang      string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	ids, err := cleanSponsorIDs("sponsorId", []string{in.SponsorID}, 1)
	if err != nil {
		return nil, err
	}
	if err := s.bannerLimit(ctx, "spn:c:"+userID, sponsorCountPerMinute); err != nil {
		return nil, err
	}
	live, err := s.liveSponsorIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	sp, ok := live[ids[0]]
	if !ok {
		return nil, notFound("SPONSOR_NOT_AVAILABLE")
	}
	url := sp.Website
	if t := strings.TrimSpace(in.Link); t != "" && t != "website" {
		url = ""
		for _, l := range sp.Links {
			if l.Type == t {
				url = l.URL
				break
			}
		}
		if url == "" {
			return nil, notFound("SPONSOR_LINK_NOT_FOUND")
		}
	}
	day, _ := s.clientDay(ctx)
	counted, err := s.countSponsors(ctx, s.db, ids, userID, day, true)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	return map[string]any{"counted": len(counted) == 1, "sponsorId": sp.ID, "url": url, "sponsor": sp.clientView(langs, lang)}, nil
}

func (s *service) sweepSponsorSeen(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM publication_sponsor_seen WHERE day < current_date - $1::int OR NOT EXISTS (SELECT 1 FROM users u WHERE u.id = publication_sponsor_seen.user_id)`, bannerKeepSeenDays)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
