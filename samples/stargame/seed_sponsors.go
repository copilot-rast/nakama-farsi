package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	sponsorKey          = "star_sp_nova"
	sponsorPlacementKey = "luck_sponsor"
	sponsorBannerKey    = "star_bn_nova"
)

func sponsorSiteBase(env map[string]string) string {
	return strings.TrimSuffix(bannerImageBase(env), "banners/")
}

func seedSponsorData(ctx context.Context, tx *sql.Tx, env map[string]string, report *seedReport) error {
	site := sponsorSiteBase(env)
	if !strings.HasPrefix(site, "http") {
		site = "http://127.0.0.1:7450" + site
	}
	links, _ := json.Marshal([]map[string]any{
		{"type": "instagram", "url": site + "sponsor.html?via=instagram", "label": text{"fa": "نوا در اینستاگرام", "en": "Nova on Instagram", "ar": "نوفا على إنستغرام"}},
		{"type": "telegram", "url": site + "sponsor.html?via=telegram", "label": text{"fa": "کانال تلگرام نوا", "en": "Nova on Telegram", "ar": "قناة نوفا على تيليغرام"}},
	})
	ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO publication_sponsor (id, key, name, description, logo_url, color, website_url, links, contact_name, contact_email, notes, sort)
		VALUES ($1, $2, $3::jsonb, $4::jsonb, $5, '#0f766e', $6, $7::jsonb, 'Star Catcher sample', 'sponsor@example.com', $8, 1) ON CONFLICT DO NOTHING`,
		newID(), sponsorKey, text{"fa": "آبمیوهٔ نوا", "en": "Nova Juice", "ar": "عصير نوفا"}.json(),
		text{"fa": "برند خیالی نمونه که چرخ روزانهٔ ستاره‌چین را حمایت می‌کند.", "en": "A made-up sample brand that backs the Star Catcher daily wheel.", "ar": "علامة تجارية خيالية تدعم عجلة صائد النجوم اليومية."}.json(),
		bannerImageBase(env)+"nova-logo.svg", site+"sponsor.html", string(links), "Seeded by the Star Catcher sample; not a real company."))
	if err != nil {
		return fmt.Errorf("sponsor: %w", err)
	}
	report.add("sponsor", sponsorKey, ok)
	spID, err := idByKey(ctx, tx, "publication_sponsor", sponsorKey)
	if err != nil {
		return err
	}
	ok, err = inserted(tx.ExecContext(ctx, `INSERT INTO publication_banner_list (id, key, name, description, layout, max_items, rotate_seconds, sort, sponsor_id)
		VALUES ($1, $2, $3::jsonb, $4::jsonb, 'single', 1, 0, 2, $5) ON CONFLICT DO NOTHING`,
		newID(), sponsorPlacementKey, text{"fa": "حامی صفحهٔ شانس", "en": "Luck screen sponsor", "ar": "راعي شاشة الحظ"}.json(),
		text{"fa": "یک بنر بالای صفحهٔ شانس با نام حامی", "en": "One banner above the luck screen with the sponsor's name", "ar": "لافتة واحدة أعلى شاشة الحظ باسم الراعي"}.json(), spID))
	if err != nil {
		return fmt.Errorf("sponsor placement: %w", err)
	}
	report.add("banner placement", sponsorPlacementKey, ok)
	listID, err := idByKey(ctx, tx, "publication_banner_list", sponsorPlacementKey)
	if err != nil {
		return err
	}
	ok, err = inserted(tx.ExecContext(ctx, `INSERT INTO publication_banner (id, key, list_id, name, title, subtitle, image_url, bg_color, action_type, action_value, action_label, priority, sponsor_id)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, '#0f766e', 'sponsor', '', $8::jsonb, 10, $9) ON CONFLICT DO NOTHING`,
		newID(), sponsorBannerKey, listID, "Star Catcher: sponsored daily wheel",
		text{"fa": "چرخ روزانه با حمایت آبمیوهٔ نوا", "en": "The daily wheel, sponsored by Nova Juice", "ar": "العجلة اليومية برعاية عصير نوفا"}.json(),
		text{"fa": "هر روز یک چرخش رایگان، هدیهٔ نوا", "en": "One free spin every day, a gift from Nova", "ar": "دورة مجانية كل يوم، هدية من نوفا"}.json(),
		text{"fa": bannerImageBase(env) + "nova.svg"}.json(),
		text{"fa": "دربارهٔ نوا", "en": "About Nova", "ar": "عن نوفا"}.json(), spID))
	if err != nil {
		return fmt.Errorf("sponsor banner: %w", err)
	}
	report.add("banner", sponsorBannerKey, ok)
	return nil
}
