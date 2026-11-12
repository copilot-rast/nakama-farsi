package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const bannerPlacementKey = "menu_carousel"

type seedBanner struct {
	key, image, imageEn, actionType, actionValue string
	title, subtitle, label                       text
	priority                                     int
	vipOnly                                      bool
	playerAge                                    string
}

var seedBanners = []seedBanner{
	{key: "star_bn_welcome", image: "welcome-fa.svg", imageEn: "welcome-en.svg", actionType: "screen", actionValue: "gift", priority: 100, playerAge: "new",
		title:    text{"fa": "به ستاره‌چین خوش آمدی", "en": "Welcome to Star Catcher", "ar": "مرحبًا بك في صائد النجوم"},
		subtitle: text{"fa": "با کد STAR100 صد سکهٔ هدیه بگیر", "en": "Get 100 free coins with the code STAR100", "ar": "احصل على 100 عملة مجانية بالرمز STAR100"},
		label:    text{"fa": "کد هدیه", "en": "Gift code", "ar": "رمز الهدية"}},
	{key: "star_bn_vip", image: "vip.svg", actionType: "product", actionValue: "star_pack_vip", priority: 90, vipOnly: true,
		title:    text{"fa": "پیشنهاد ویژهٔ اعضا", "en": "Members-only offer", "ar": "عرض خاص للأعضاء"},
		subtitle: text{"fa": "۱۲۰۰ سکه با نصف قیمت، فقط برای تو", "en": "1200 coins at half price, just for you", "ar": "1200 عملة بنصف السعر، لك وحدك"},
		label:    text{"fa": "بخر", "en": "Buy", "ar": "اشترِ"}},
	{key: "star_bn_galaxy", image: "skins.svg", actionType: "product", actionValue: "star_skin_galaxy", priority: 80,
		title:    text{"fa": "سبد کهکشانی رسید", "en": "The galaxy basket is here", "ar": "وصلت سلة المجرة"},
		subtitle: text{"fa": "با سکه‌هایت بخرش و بدرخش", "en": "Buy it with your coins and shine", "ar": "اشترها بعملاتك وتألّق"},
		label:    text{"fa": "در فروشگاه", "en": "In the shop", "ar": "في المتجر"}},
	{key: "star_bn_missions", image: "missions.svg", actionType: "screen", actionValue: "tasks", priority: 60,
		title:    text{"fa": "مأموریت‌های امروز", "en": "Today's missions", "ar": "مهام اليوم"},
		subtitle: text{"fa": "مأموریت‌ها را تمام کن و سکه بگیر", "en": "Finish missions and earn coins", "ar": "أنهِ المهام واربح العملات"},
		label:    text{"fa": "مأموریت‌ها", "en": "Missions", "ar": "المهام"}},
	{key: "star_bn_fan", image: "fan.svg", actionType: "achievement", actionValue: "star_ach_supporter", priority: 40,
		title:    text{"fa": "هوادار ستاره‌چین شو", "en": "Become a Star Catcher fan", "ar": "كن مشجعًا لصائد النجوم"},
		subtitle: text{"fa": "با کد STARFAN نشان هوادار بگیر", "en": "Get the fan badge with the code STARFAN", "ar": "احصل على شارة المشجع بالرمز STARFAN"},
		label:    text{"fa": "نشان را ببین", "en": "See the badge", "ar": "شاهد الشارة"}},
}

func bannerImageBase(env map[string]string) string {
	base := strings.TrimRight(strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"]), "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = ""
	}
	return base + "/stargame/banners/"
}

func seedBannerData(ctx context.Context, tx *sql.Tx, env map[string]string, vipTag string, report *seedReport) error {
	ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO publication_banner_list (id, key, name, description, layout, max_items, rotate_seconds, sort)
		VALUES ($1, $2, $3::jsonb, $4::jsonb, 'carousel', 5, 5, 1) ON CONFLICT DO NOTHING`,
		newID(), bannerPlacementKey, text{"fa": "چرخ‌فلک منوی اصلی", "en": "Home menu carousel", "ar": "دوّار القائمة الرئيسية"}.json(),
		text{"fa": "بنرهای بالای منوی صفحهٔ اصلی ستاره‌چین", "en": "Banners above the Star Catcher home menu", "ar": "اللافتات فوق قائمة صائد النجوم الرئيسية"}.json()))
	if err != nil {
		return fmt.Errorf("banner placement: %w", err)
	}
	report.add("banner placement", bannerPlacementKey, ok)
	listID, err := idByKey(ctx, tx, "publication_banner_list", bannerPlacementKey)
	if err != nil {
		return err
	}
	base := bannerImageBase(env)
	for _, b := range seedBanners {
		image := text{"fa": base + b.image}
		if b.imageEn != "" {
			image["en"] = base + b.imageEn
		}
		audience := []string{}
		if b.vipOnly {
			audience = append(audience, vipTag)
		}
		age := b.playerAge
		if age == "" {
			age = "all"
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO publication_banner (id, key, list_id, name, title, subtitle, image_url, action_type, action_value, action_label, priority, audience_tag_ids, player_age, new_days)
			VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8, $9, $10::jsonb, $11, $12, $13, 7) ON CONFLICT DO NOTHING`,
			newID(), b.key, listID, "Star Catcher: "+b.key, b.title.json(), b.subtitle.json(), image.json(), b.actionType, b.actionValue, b.label.json(), b.priority, audience, age))
		if err != nil {
			return fmt.Errorf("banner %s: %w", b.key, err)
		}
		report.add("banner", b.key, ok)
	}
	return nil
}
