package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	bannerListPerMinute  = 60
	bannerCountPerMinute = 60
	bannerKeepSeenDays   = 45
)

type placement struct {
	id, key, layout string
	name            i18n
	maxItems        int
	rotate          int
	active          bool
}

func (s *service) loadPlacement(ctx context.Context, by, value string) (placement, error) {
	var p placement
	var name string
	col := "key"
	if by == "id" {
		col = "id::text"
	}
	err := s.db.QueryRowContext(ctx, `SELECT id::text, key, name::text, layout, max_items, rotate_seconds, is_active FROM publication_banner_list WHERE `+col+`=$1`, value).
		Scan(&p.id, &p.key, &name, &p.layout, &p.maxItems, &p.rotate, &p.active)
	if errors.Is(err, sql.ErrNoRows) {
		return p, notFound("BANNER_LIST_NOT_FOUND")
	}
	if err != nil {
		return p, err
	}
	p.name = parseI18n(name)
	return p, nil
}

func cleanClientTarget(platform, version string) (string, string, error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform != "" && !containsStr(bannerPlatforms, platform) {
		return "", "", badRequest("INVALID_PLATFORM", "platform", "unknown platform")
	}
	version = strings.TrimPrefix(strings.TrimSpace(latinDigits(version)), "v")
	if version != "" {
		if _, ok := bannerVersionParts(version); !ok || len(version) > 40 {
			return "", "", badRequest("INVALID_VERSION", "version", "version is not valid")
		}
	}
	return platform, version, nil
}

func (s *service) bannerViewer(ctx context.Context, userID, platform, version, requested, ipCountry string) (*viewer, error) {
	platform, version, err := cleanClientTarget(platform, version)
	if err != nil {
		return nil, err
	}
	v := &viewer{UserID: userID, Platform: platform, Version: version, Tags: map[string]bool{}, Now: time.Now(), q: s.db}
	v.Zone = projectZone(ctx, s.db)
	var country sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT metadata->>'country', create_time FROM users WHERE id=$1`, userID).Scan(&country, &v.Created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("USER_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	v.Country = strings.ToUpper(strings.TrimSpace(country.String))
	if v.Country == "" {
		v.Country = strings.ToUpper(ipCountry)
	}
	_, lang, err := localized.ForUser(ctx, s.db, userID, requested)
	if err != nil {
		return nil, err
	}
	v.Lang = lang
	rows, err := s.db.QueryContext(ctx, `SELECT tag_id::text FROM base_user_tag WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		v.Tags[id] = true
	}
	return v, rows.Err()
}

func (s *service) bannerLimit(ctx context.Context, key string, limit int) error {
	d, err := clusterkit.Allow(ctx, s.db, key, int64(limit), time.Minute)
	if err != nil {
		return err
	}
	if !d.Allowed {
		return fail(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "", "try again in "+strconv.Itoa(int(d.RetryAfter.Seconds())+1)+" seconds")
	}
	return nil
}

func (s *service) rpcBanners(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Placement string `json:"placement"`
		Platform  string `json:"platform"`
		Version   string `json:"version"`
		Lang      string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(in.Placement)
	if !bannerKeyPattern.MatchString(key) {
		return nil, badRequest("INVALID_PLACEMENT", "placement", "placement key is not valid")
	}
	if err := s.bannerLimit(ctx, "bnr:l:"+userID, bannerListPerMinute); err != nil {
		return nil, err
	}
	p, err := s.loadPlacement(ctx, "key", key)
	if err != nil {
		return nil, err
	}
	if !p.active {
		return nil, notFound("BANNER_LIST_NOT_FOUND")
	}
	v, err := s.bannerViewer(ctx, userID, in.Platform, in.Version, in.Lang, activity.FromContext(ctx, activity.Entry{}).Country)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	banners, err := loadBanners(ctx, s.db, `WHERE b.list_id=$2 AND b.is_active`, v.tagList(), p.id)
	if err != nil {
		return nil, err
	}
	items := []map[string]any{}
	for _, b := range banners {
		if len(items) >= p.maxItems {
			break
		}
		reason, err := b.eligible(ctx, v)
		if err != nil {
			return nil, err
		}
		if reason == "" {
			items = append(items, b.view(langs, lang))
		}
	}
	return map[string]any{"lang": lang, "placement": map[string]any{"key": p.key, "name": langs.Resolve(p.name, lang), "layout": p.layout, "rotateSeconds": p.rotate, "maxItems": p.maxItems},
		"items": items}, nil
}

func (s *service) eligibleBanners(ctx context.Context, v *viewer, ids []string) (map[string]bannerRow, error) {
	banners, err := loadBanners(ctx, s.db, `JOIN publication_banner_list l ON l.id=b.list_id WHERE b.id = ANY($2::uuid[]) AND b.is_active AND l.is_active`, v.tagList(), ids)
	if err != nil {
		return nil, err
	}
	out := map[string]bannerRow{}
	for _, b := range banners {
		reason, err := b.eligible(ctx, v)
		if err != nil {
			return nil, err
		}
		if reason == "" {
			out[b.ID] = b
		}
	}
	return out, nil
}

func cleanBannerIDs(field string, ids []string, max int) ([]string, error) {
	if len(ids) == 0 {
		return nil, badRequest("REQUIRED", field, "give at least one banner id")
	}
	if len(ids) > max {
		return nil, badRequest("TOO_MANY_BANNERS", field, "at most "+strconv.Itoa(max)+" banners")
	}
	out := []string{}
	seen := map[string]bool{}
	for _, raw := range ids {
		id, err := uuid.FromString(strings.TrimSpace(raw))
		if err != nil {
			return nil, badRequest("INVALID_ID", field, "banner ids are UUIDs")
		}
		if !seen[id.String()] {
			seen[id.String()] = true
			out = append(out, id.String())
		}
	}
	return out, nil
}

func localDay(v *viewer) string {
	return v.Now.In(v.Zone).Format("2006-01-02")
}

func (s *service) rpcBannerView(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		BannerIDs []string `json:"bannerIds"`
		Platform  string   `json:"platform"`
		Version   string   `json:"version"`
		Lang      string   `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	ids, err := cleanBannerIDs("bannerIds", in.BannerIDs, 20)
	if err != nil {
		return nil, err
	}
	if err := s.bannerLimit(ctx, "bnr:v:"+userID, bannerCountPerMinute); err != nil {
		return nil, err
	}
	v, err := s.bannerViewer(ctx, userID, in.Platform, in.Version, in.Lang, activity.FromContext(ctx, activity.Entry{}).Country)
	if err != nil {
		return nil, err
	}
	ok, err := s.eligibleBanners(ctx, v, ids)
	if err != nil {
		return nil, err
	}
	valid := []string{}
	for _, id := range ids {
		if _, hit := ok[id]; hit {
			valid = append(valid, id)
		}
	}
	counted := []string{}
	if len(valid) > 0 {
		rows, err := s.db.QueryContext(ctx, `WITH s AS (
			INSERT INTO publication_banner_seen (banner_id, user_id, day, viewed_at) SELECT x::uuid, $2, $3::date, now() FROM unnest($1::text[]) x
			ON CONFLICT (banner_id, user_id, day) DO UPDATE SET viewed_at = now() WHERE publication_banner_seen.viewed_at IS NULL
			RETURNING banner_id)
			INSERT INTO publication_banner_day (banner_id, day, views) SELECT banner_id, $3::date, 1 FROM s
			ON CONFLICT (banner_id, day) DO UPDATE SET views = publication_banner_day.views + 1 RETURNING banner_id::text`, valid, userID, localDay(v))
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			counted = append(counted, id)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	skipped := []string{}
	for _, id := range ids {
		if _, hit := ok[id]; !hit {
			skipped = append(skipped, id)
		}
	}
	return map[string]any{"counted": counted, "skipped": skipped, "day": localDay(v)}, nil
}

func (s *service) rpcBannerClick(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		BannerID string `json:"bannerId"`
		Platform string `json:"platform"`
		Version  string `json:"version"`
		Lang     string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	ids, err := cleanBannerIDs("bannerId", []string{in.BannerID}, 1)
	if err != nil {
		return nil, err
	}
	if err := s.bannerLimit(ctx, "bnr:c:"+userID, bannerCountPerMinute); err != nil {
		return nil, err
	}
	v, err := s.bannerViewer(ctx, userID, in.Platform, in.Version, in.Lang, activity.FromContext(ctx, activity.Entry{}).Country)
	if err != nil {
		return nil, err
	}
	ok, err := s.eligibleBanners(ctx, v, ids)
	if err != nil {
		return nil, err
	}
	b, hit := ok[ids[0]]
	if !hit {
		return nil, notFound("BANNER_NOT_AVAILABLE")
	}
	var counted bool
	err = s.db.QueryRowContext(ctx, `WITH s AS (
		INSERT INTO publication_banner_seen (banner_id, user_id, day, clicked_at) VALUES ($1, $2, $3::date, now())
		ON CONFLICT (banner_id, user_id, day) DO UPDATE SET clicked_at = now() WHERE publication_banner_seen.clicked_at IS NULL
		RETURNING banner_id),
		d AS (INSERT INTO publication_banner_day (banner_id, day, clicks) SELECT banner_id, $3::date, 1 FROM s
		ON CONFLICT (banner_id, day) DO UPDATE SET clicks = publication_banner_day.clicks + 1 RETURNING 1)
		SELECT EXISTS (SELECT 1 FROM d)`, b.ID, userID, localDay(v)).Scan(&counted)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	view := b.view(langs, lang)
	return map[string]any{"counted": counted, "action": view["action"], "bannerId": b.ID}, nil
}

func (s *service) bannerSweepLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := clusterkit.AcquireLease(ctx, s.db, "publication-banner-sweep", holder, 29*time.Minute)
			if err != nil || !ok {
				continue
			}
			if _, err := s.sweepBannerSeen(ctx); err != nil && ctx.Err() == nil {
				s.logger.Error("Publication banner sweep failed: %v", err)
			}
		}
	}
}

func (s *service) sweepBannerSeen(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM publication_banner_seen WHERE day < current_date - $1::int OR NOT EXISTS (SELECT 1 FROM users u WHERE u.id = publication_banner_seen.user_id)`, bannerKeepSeenDays)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
