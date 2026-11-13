package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

type text map[string]string

func (t text) json() string {
	b, _ := json.Marshal(t)
	return string(b)
}

type seedLanguage struct {
	code, native, direction string
	name                    text
	sort                    int
}

type seedCategory struct {
	key               string
	name, description text
	sort              int
}

type seedPrize struct {
	key           string
	name, message text
	repeatable    bool
	items         string
}

type seedProduct struct {
	key, category, prize string
	name, description    text
	priceToman           any
	coinPrice            any
	methods              []string
	sku                  string
	sort                 int
	audience             string
}

type seedGift struct {
	code, kind, prize string
	name              text
	discount          any
}

type seedMethod struct {
	method string
	config map[string]any
}

const storePackage = "ir.nakama.stargame"

var (
	seedLanguages = []seedLanguage{
		{"fa", "فارسی", "rtl", text{"fa": "فارسی", "en": "Persian"}, 1},
		{"en", "English", "ltr", text{"fa": "انگلیسی", "en": "English"}, 2},
		{"ar", "العربية", "rtl", text{"fa": "عربی", "en": "Arabic", "ar": "العربية"}, 3},
	}
	seedCategories = []seedCategory{
		{"star_coin_packs", text{"fa": "بسته‌های سکه", "en": "Coin packs", "ar": "حزم العملات"}, text{"fa": "سکه بخر و زودتر سبد تازه بگیر", "en": "Buy coins and unlock new baskets sooner", "ar": "اشترِ العملات واحصل على سلال جديدة أسرع"}, 1},
		{"star_skins", text{"fa": "اسکین‌ها", "en": "Skins", "ar": "المظاهر"}, text{"fa": "سبدهای تازه برای بازی", "en": "New baskets for the game", "ar": "سلال جديدة للعبة"}, 2},
	}
	gold       = text{"fa": "سبد طلایی", "en": "Golden basket", "ar": "السلة الذهبية"}
	neon       = text{"fa": "سبد نئونی", "en": "Neon basket", "ar": "سلة النيون"}
	galaxy     = text{"fa": "سبد کهکشانی", "en": "Galaxy basket", "ar": "سلة المجرة"}
	seedPrizes = []seedPrize{
		coinPrize("star_coins_100", 100, "۱۰۰"),
		coinPrize("star_coins_500", 500, "۵۰۰"),
		coinPrize("star_coins_1200", 1200, "۱۲۰۰"),
		skinPrize("star_skin_gold", "skin_gold", gold),
		skinPrize("star_skin_neon", "skin_neon", neon),
		skinPrize("star_skin_galaxy", "skin_galaxy", galaxy),
	}
	seedProducts = []seedProduct{
		pack("star_pack_small", text{"fa": "بسته کوچک", "en": "Small pack", "ar": "حزمة صغيرة"}, 100, "۱۰۰", 10000, "star_coins_100", 1),
		pack("star_pack_medium", text{"fa": "بسته متوسط", "en": "Medium pack", "ar": "حزمة متوسطة"}, 500, "۵۰۰", 40000, "star_coins_500", 2),
		pack("star_pack_large", text{"fa": "بسته بزرگ", "en": "Large pack", "ar": "حزمة كبيرة"}, 1200, "۱۲۰۰", 90000, "star_coins_1200", 3),
		skin("star_skin_gold", gold, text{"fa": "سبدی براق از طلا", "en": "A shiny basket made of gold", "ar": "سلة لامعة من الذهب"}, 80, 11),
		skin("star_skin_neon", neon, text{"fa": "سبدی که در تاریکی می‌درخشد", "en": "A basket that glows in the dark", "ar": "سلة تتوهج في الظلام"}, 200, 12),
		skin("star_skin_galaxy", galaxy, text{"fa": "سبدی پر از ستاره و رنگ", "en": "A basket full of stars and color", "ar": "سلة مليئة بالنجوم والألوان"}, 400, 13),
		vipPack(),
	}
	seedGifts = []seedGift{
		{"STAR100", "reward", "star_coins_100", text{"fa": "ستاره‌چین: ۱۰۰ سکه هدیه", "en": "Star Catcher: 100 free coins", "ar": "صائد النجوم: 100 عملة هدية"}, nil},
		{"HALF50", "discount", "", text{"fa": "ستاره‌چین: ۵۰٪ تخفیف", "en": "Star Catcher: 50% off", "ar": "صائد النجوم: خصم 50٪"}, 50},
	}
	seedMethods = []seedMethod{
		{"in_app", map[string]any{}},
		{"zarinpal", map[string]any{"sandbox": true}},
		{"zibal", map[string]any{"sandbox": true}},
		{"bazaar", map[string]any{"packageName": storePackage, "sandbox": true}},
		{"myket", map[string]any{"packageName": storePackage, "sandbox": true}},
	}
)

func coinPrize(key string, amount int, faAmount string) seedPrize {
	return seedPrize{
		key:        key,
		name:       text{"fa": faAmount + " سکه", "en": fmt.Sprintf("%d coins", amount), "ar": fmt.Sprintf("%d عملة", amount)},
		message:    text{"fa": faAmount + " سکه به کیف پولت اضافه شد", "en": fmt.Sprintf("%d coins were added to your wallet", amount), "ar": fmt.Sprintf("تمت إضافة %d عملة إلى محفظتك", amount)},
		repeatable: true,
		items:      fmt.Sprintf(`[{"type":"currency","key":"coin","action":"add","value":%d}]`, amount),
	}
}

func skinPrize(key, skin string, name text) seedPrize {
	value, _ := json.Marshal(map[string]string{"skin": skin, "name": name["fa"]})
	items, _ := json.Marshal([]map[string]any{{"type": "storage", "key": skin, "action": "give", "value": json.RawMessage(value)}})
	return seedPrize{
		key:     key,
		name:    name,
		message: text{"fa": name["fa"] + " مال تو شد", "en": "The " + name["en"] + " is yours", "ar": name["ar"] + " أصبحت لك"},
		items:   string(items),
	}
}

func pack(key string, name text, amount int, faAmount string, price int64, prize string, sort int) seedProduct {
	desc := text{"fa": faAmount + " سکه", "en": fmt.Sprintf("%d coins", amount), "ar": fmt.Sprintf("%d عملة", amount)}
	sku, _ := json.Marshal(map[string]string{"bazaar": key, "myket": key})
	return seedProduct{key: key, category: "star_coin_packs", prize: prize, name: name, description: desc, priceToman: price, methods: []string{"zarinpal", "zibal", "bazaar", "myket"}, sku: string(sku), sort: sort}
}

const vipTagKey = "star-vip"

func vipPack() seedProduct {
	p := pack("star_pack_vip", text{"fa": "بستهٔ اعضای ویژه", "en": "VIP pack", "ar": "حزمة الأعضاء المميزين"}, 1200, "۱۲۰۰", 45000, "star_coins_1200", 4)
	p.description = text{"fa": "۱۲۰۰ سکه به نصف قیمت، فقط برای اعضای ویژه", "en": "1200 coins at half price, only for VIP members", "ar": "1200 عملة بنصف السعر للأعضاء المميزين فقط"}
	p.audience = vipTagKey
	return p
}

func skin(key string, name, desc text, coins int64, sort int) seedProduct {
	price, _ := json.Marshal(map[string]any{"currencyKey": currencyKey, "amount": coins})
	return seedProduct{key: key, category: "star_skins", prize: key, name: name, description: desc, coinPrice: string(price), methods: []string{"in_app"}, sku: "{}", sort: sort}
}

type seedReport struct {
	created []string
	skipped []string
}

func (r *seedReport) add(kind, key string, inserted bool) {
	if inserted {
		r.created = append(r.created, kind+":"+key)
	}
}

func (r *seedReport) String() string {
	out := "nothing new"
	if len(r.created) > 0 {
		out = "created " + strings.Join(r.created, ", ")
	}
	if len(r.skipped) > 0 {
		out += "; skipped " + strings.Join(r.skipped, ", ")
	}
	return out
}

func seedEnabled(env map[string]string) bool {
	return !strings.EqualFold(strings.TrimSpace(env["STARGAME_SEED"]), "false")
}

func newID() string { return uuid.Must(uuid.NewV4()).String() }

func inserted(res sql.Result, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func seed(ctx context.Context, db *sql.DB, env map[string]string) (*seedReport, error) {
	report := &seedReport{}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, l := range seedLanguages {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO base_language (code, name, native_name, direction, is_active, is_default, sort)
			SELECT $1, $2::jsonb, $3, $4, true, NOT EXISTS (SELECT 1 FROM base_language WHERE is_default), $5
			ON CONFLICT DO NOTHING`, l.code, l.name.json(), l.native, l.direction, l.sort))
		if err != nil {
			return nil, fmt.Errorf("language %s: %w", l.code, err)
		}
		report.add("language", l.code, ok)
	}

	ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_currency (id, key, name, type, default_value, sort) VALUES ($1, $2, $3::jsonb, 'soft', 50, 1) ON CONFLICT DO NOTHING`,
		newID(), currencyKey, text{"fa": "سکه", "en": "Coin", "ar": "عملة"}.json()))
	if err != nil {
		return nil, fmt.Errorf("currency: %w", err)
	}
	report.add("currency", currencyKey, ok)

	categories := map[string]string{}
	for _, c := range seedCategories {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO base_category (id, key_type, key, name, description, sort) VALUES ($1, 'product', $2, $3::jsonb, $4::jsonb, $5) ON CONFLICT DO NOTHING`,
			newID(), c.key, c.name.json(), c.description.json(), c.sort))
		if err != nil {
			return nil, fmt.Errorf("category %s: %w", c.key, err)
		}
		report.add("category", c.key, ok)
		var id string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='product' AND key=$1`, c.key).Scan(&id); err != nil {
			return nil, fmt.Errorf("category %s: %w", c.key, err)
		}
		categories[c.key] = id
	}

	prizes := map[string]string{}
	for _, p := range append(append(append([]seedPrize{}, seedPrizes...), eventCoinPrizes()...), growthPrizes()...) {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_prize (id, key, name, message, is_repeatable, items) VALUES ($1, $2, $3::jsonb, $4::jsonb, $5, $6::jsonb) ON CONFLICT DO NOTHING`,
			newID(), p.key, p.name.json(), p.message.json(), p.repeatable, p.items))
		if err != nil {
			return nil, fmt.Errorf("prize %s: %w", p.key, err)
		}
		report.add("prize", p.key, ok)
		if _, err := tx.ExecContext(ctx, `UPDATE economy_prize SET message = jsonb_set(message, '{fa}', to_jsonb($2::text)), updated_at = now() WHERE key = $1 AND message->>'fa' <> $2 AND translate(message->>'fa', '0123456789', '۰۱۲۳۴۵۶۷۸۹') = $2`, p.key, p.message["fa"]); err != nil {
			return nil, fmt.Errorf("prize %s: %w", p.key, err)
		}
		var id string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM economy_prize WHERE key=$1`, p.key).Scan(&id); err != nil {
			return nil, fmt.Errorf("prize %s: %w", p.key, err)
		}
		prizes[p.key] = id
	}

	ok, err = inserted(tx.ExecContext(ctx, `INSERT INTO base_tag (id, key_type, key, name, description, is_public, sort) VALUES ($1, 'user', $2, $3::jsonb, $4::jsonb, true, 1) ON CONFLICT DO NOTHING`,
		newID(), vipTagKey, text{"fa": "عضو ویژه", "en": "VIP member", "ar": "عضو مميز"}.json(), text{"fa": "بازیکنانی که پیشنهادهای ویژه می‌گیرند", "en": "Players who get special offers", "ar": "لاعبون يحصلون على عروض خاصة"}.json()))
	if err != nil {
		return nil, fmt.Errorf("player tag: %w", err)
	}
	report.add("player tag", vipTagKey, ok)
	var vipTag string
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_tag WHERE key_type='user' AND key=$1`, vipTagKey).Scan(&vipTag); err != nil {
		return nil, fmt.Errorf("player tag: %w", err)
	}

	for _, p := range seedProducts {
		audience := []string{}
		if p.audience == vipTagKey {
			audience = append(audience, vipTag)
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_product (id, key, name, description, is_active, category_id, price_toman, coin_price, payment_methods, sku, prize_id, sort, audience_tag_ids)
			VALUES ($1, $2, $3::jsonb, $4::jsonb, true, $5, $6, $7::jsonb, $8, $9::jsonb, $10, $11, $12) ON CONFLICT DO NOTHING`,
			newID(), p.key, p.name.json(), p.description.json(), categories[p.category], p.priceToman, p.coinPrice, p.methods, p.sku, prizes[p.prize], p.sort, audience))
		if err != nil {
			return nil, fmt.Errorf("product %s: %w", p.key, err)
		}
		report.add("product", p.key, ok)
	}

	for _, g := range seedGifts {
		var prize any
		if g.prize != "" {
			prize = prizes[g.prize]
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_gift (id, name, type, code, prize_id, discount_percent, is_promo, per_user_limit, is_active)
			VALUES ($1, $2::jsonb, $3, $4, $5, $6, true, 1, true) ON CONFLICT DO NOTHING`,
			newID(), g.name.json(), g.kind, g.code, prize, g.discount))
		if err != nil {
			return nil, fmt.Errorf("gift %s: %w", g.code, err)
		}
		report.add("gift", g.code, ok)
	}

	if err := seedEvents(ctx, tx, prizes, report); err != nil {
		return nil, err
	}
	if err := seedGrowth(ctx, tx, prizes, vipTag, report); err != nil {
		return nil, err
	}
	if err := seedTaskData(ctx, tx, prizes, report); err != nil {
		return nil, err
	}
	if err := seedItemData(ctx, tx, categories, report); err != nil {
		return nil, err
	}
	if err := seedAchievementData(ctx, tx, prizes, report); err != nil {
		return nil, err
	}
	if err := seedChanceData(ctx, tx, prizes, report); err != nil {
		return nil, err
	}
	if err := seedGameData(ctx, tx, env, prizes, report); err != nil {
		return nil, err
	}
	if err := seedBannerData(ctx, tx, env, vipTag, report); err != nil {
		return nil, err
	}
	if err := seedSponsorData(ctx, tx, env, report); err != nil {
		return nil, err
	}
	if err := seedNotifyData(ctx, tx, report); err != nil {
		return nil, err
	}
	if err := seedPaymentMethods(ctx, tx, env, report); err != nil {
		return nil, err
	}
	if err := seedAppVersions(ctx, tx, env, report); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return report, nil
}

func seedPaymentMethods(ctx context.Context, tx *sql.Tx, env map[string]string, report *seedReport) error {
	secret, err := secretbox.Secret(env)
	if err != nil {
		report.skipped = append(report.skipped, "payment methods ("+err.Error()+")")
		return nil
	}
	key, err := secretbox.Economy.DeriveKey(secret, secretbox.EconomyConfigInfo)
	if err != nil {
		return err
	}
	for _, m := range seedMethods {
		plain, _ := json.Marshal(m.config)
		sealed, err := secretbox.Economy.Seal(key, plain)
		if err != nil {
			return err
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_payment_method (method, is_active, config) VALUES ($1, true, $2)
			ON CONFLICT (method) DO UPDATE SET is_active = true, config = EXCLUDED.config, updated_at = now()
			WHERE economy_payment_method.config = '' AND NOT economy_payment_method.is_active`, m.method, sealed))
		if err != nil {
			return fmt.Errorf("payment method %s: %w", m.method, err)
		}
		report.add("method", m.method, ok)
	}
	return nil
}

const GameVersion = "1.4.0"

func seedAppVersions(ctx context.Context, tx *sql.Tx, env map[string]string, report *seedReport) error {
	base := strings.TrimRight(strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"]), "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://nakama.dastar.dev"
	}
	update := text{"fa": "نسخهٔ تازهٔ ستاره‌چین آمده. با به‌روزرسانی، ستاره‌های بیشتری می‌گیری!", "en": "A new Star Catcher is out. Update to catch more stars!", "ar": "إصدار جديد من صائد النجوم متاح. حدّث اللعبة!"}
	force := text{"fa": "این نسخه دیگر کار نمی‌کند. برای ادامه، بازی را به‌روز کن.", "en": "This version no longer works. Please update to keep playing.", "ar": "هذا الإصدار لم يعد يعمل. يرجى التحديث للمتابعة."}
	rows := []struct {
		platform string
		links    string
	}{
		{"web", `[{"store":"web","url":"` + base + `/stargame/","label":{}}]`},
		{"android", `[{"store":"bazaar","url":"https://cafebazaar.ir/app/ir.nakama.stargame","label":{}},{"store":"myket","url":"https://myket.ir/app/ir.nakama.stargame","label":{}}]`},
	}
	for _, r := range rows {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO base_app_version (id, platform, latest_version, min_version, store_links, update_message, force_message)
			VALUES ($1, $2, $3, '1.0.0', $4::jsonb, $5::jsonb, $6::jsonb) ON CONFLICT DO NOTHING`, newID(), r.platform, GameVersion, r.links, update.json(), force.json()))
		if err != nil {
			return fmt.Errorf("app version %s: %w", r.platform, err)
		}
		report.add("app version", r.platform, ok)
	}
	return nil
}
