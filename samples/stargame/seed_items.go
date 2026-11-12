package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type seedItem struct {
	key, kind, rarity, art string
	name, description      text
	starter                bool
	price                  int64
	unlockLevel            any
	sort                   int
}

var seedItems = []seedItem{
	{key: "star_basket_classic", kind: "skin", rarity: "common", art: "default", starter: true, sort: 1,
		name:        text{"fa": "سبد چوبی", "en": "Wooden basket", "ar": "السلة الخشبية"},
		description: text{"fa": "سبدی که همه از اول دارند.", "en": "The basket everyone starts with.", "ar": "السلة التي يبدأ بها الجميع."}},
	{key: "star_basket_ruby", kind: "skin", rarity: "rare", art: "skin_ruby", sort: 2,
		name:        text{"fa": "سبد یاقوتی", "en": "Ruby basket", "ar": "سلة الياقوت"},
		description: text{"fa": "سبدی سرخ و براق؛ از فروشگاه بخر.", "en": "A bright red basket, sold in the shop.", "ar": "سلة حمراء لامعة تُباع في المتجر."}},
	{key: "star_basket_ice", kind: "skin", rarity: "epic", art: "skin_ice", price: 120, sort: 3,
		name:        text{"fa": "سبد یخی", "en": "Ice basket", "ar": "السلة الجليدية"},
		description: text{"fa": "سبدی سرد و روشن؛ همین‌جا با سکه بخر.", "en": "A cold, bright basket you buy here with coins.", "ar": "سلة باردة ومضيئة تشتريها هنا بالعملات."}},
	{key: "star_frame_gold", kind: "frame", rarity: "uncommon", unlockLevel: 2, sort: 4,
		name:        text{"fa": "قاب طلایی", "en": "Golden frame", "ar": "الإطار الذهبي"},
		description: text{"fa": "قابی دور نام تو؛ در سطح ۲ رایگان باز می‌شود.", "en": "A frame around your name, free at level 2.", "ar": "إطار حول اسمك، مجاني في المستوى 2."}},
}

var seedItemPrizes = []seedPrize{
	{key: "star_item_ruby", name: text{"fa": "سبد یاقوتی", "en": "Ruby basket", "ar": "سلة الياقوت"},
		message: text{"fa": "سبد یاقوتی به انبارت اضافه شد", "en": "The ruby basket is in your inventory", "ar": "أُضيفت سلة الياقوت إلى مخزونك"},
		repeatable: true, items: `[{"type":"item","key":"star_basket_ruby","action":"give","value":1}]`},
	{key: "star_starter_pack", name: text{"fa": "بستهٔ شروع", "en": "Starter pack", "ar": "حزمة البداية"},
		message: text{"fa": "۱۵۰ سکه به کیف پولت اضافه شد", "en": "150 coins were added to your wallet", "ar": "تمت إضافة 150 عملة إلى محفظتك"},
		repeatable: true, items: `[{"type":"currency","key":"coin","action":"add","value":150}]`},
}

func seedItemData(ctx context.Context, tx *sql.Tx, categories map[string]string, report *seedReport) error {
	var levels any
	var tableID string
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM economy_level_table WHERE key='star_levels'`).Scan(&tableID); err == nil {
		levels = tableID
	}
	for _, it := range seedItems {
		var typeID string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_category WHERE key_type='item_type' AND key=$1`, it.kind).Scan(&typeID); err != nil {
			return fmt.Errorf("item type %s: %w", it.kind, err)
		}
		var price any
		if it.price > 0 {
			b, _ := json.Marshal(map[string]any{"currencyKey": currencyKey, "amount": it.price})
			price = string(b)
		}
		meta := "{}"
		if it.art != "" {
			b, _ := json.Marshal(map[string]string{"art": it.art})
			meta = string(b)
		}
		var table any
		if it.unlockLevel != nil {
			table = levels
		}
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_item (id, key, type_id, name, description, rarity, is_starter, price, unlock_level, level_table_id, metadata, sort)
			VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7, $8::jsonb, $9, $10, $11::jsonb, $12) ON CONFLICT DO NOTHING`,
			newID(), it.key, typeID, it.name.json(), it.description.json(), it.rarity, it.starter, price, it.unlockLevel, table, meta, it.sort))
		if err != nil {
			return fmt.Errorf("item %s: %w", it.key, err)
		}
		report.add("item", it.key, ok)
	}
	prizes := map[string]string{}
	for _, p := range seedItemPrizes {
		ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_prize (id, key, name, message, is_repeatable, items) VALUES ($1, $2, $3::jsonb, $4::jsonb, $5, $6::jsonb) ON CONFLICT DO NOTHING`,
			newID(), p.key, p.name.json(), p.message.json(), p.repeatable, p.items))
		if err != nil {
			return fmt.Errorf("prize %s: %w", p.key, err)
		}
		report.add("prize", p.key, ok)
		var id string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM economy_prize WHERE key=$1`, p.key).Scan(&id); err != nil {
			return fmt.Errorf("prize %s: %w", p.key, err)
		}
		prizes[p.key] = id
	}
	ruby, _ := json.Marshal(map[string]any{"currencyKey": currencyKey, "amount": 60})
	ok, err := inserted(tx.ExecContext(ctx, `INSERT INTO economy_product (id, key, name, description, is_active, category_id, coin_price, payment_methods, sku, prize_id, sort, check_contents, purchase_message)
		VALUES ($1, 'star_item_ruby', $2::jsonb, $3::jsonb, true, $4, $5::jsonb, '{in_app}', '{}'::jsonb, $6, 14, true, $7::jsonb) ON CONFLICT DO NOTHING`,
		newID(), text{"fa": "سبد یاقوتی", "en": "Ruby basket", "ar": "سلة الياقوت"}.json(),
		text{"fa": "سبدی سرخ و براق که به انبارت می‌رود", "en": "A bright red basket for your inventory", "ar": "سلة حمراء لامعة لمخزونك"}.json(),
		categories["star_skins"], string(ruby), prizes["star_item_ruby"],
		text{"fa": "سبد یاقوتی مال تو شد؛ در «انبار» انتخابش کن.", "en": "The ruby basket is yours. Pick it in the inventory.", "ar": "سلة الياقوت لك الآن. اخترها في المخزون."}.json()))
	if err != nil {
		return fmt.Errorf("product star_item_ruby: %w", err)
	}
	report.add("product", "star_item_ruby", ok)
	pack, _ := json.Marshal(map[string]any{"currencyKey": currencyKey, "amount": 30})
	ok, err = inserted(tx.ExecContext(ctx, `INSERT INTO economy_product (id, key, name, description, is_active, category_id, price_toman, coin_price, payment_methods, sku, prize_id, sort, is_one_time, max_level, level_table_id, purchase_message)
		VALUES ($1, 'star_starter_pack', $2::jsonb, $3::jsonb, true, $4, 5000, $5::jsonb, '{in_app,zarinpal,zibal}', '{}'::jsonb, $6, 0, true, 2, $7, $8::jsonb) ON CONFLICT DO NOTHING`,
		newID(), text{"fa": "بستهٔ شروع", "en": "Starter pack", "ar": "حزمة البداية"}.json(),
		text{"fa": "۱۵۰ سکه، فقط یک بار و فقط تا سطح ۲", "en": "150 coins, once, up to level 2", "ar": "150 عملة، مرة واحدة حتى المستوى 2"}.json(),
		categories["star_coin_packs"], string(pack), prizes["star_starter_pack"], levels,
		text{"fa": "خوش آمدی! ۱۵۰ سکه به کیف پولت اضافه شد.", "en": "Welcome! 150 coins were added to your wallet.", "ar": "أهلًا بك! أُضيفت 150 عملة إلى محفظتك."}.json()))
	if err != nil {
		return fmt.Errorf("product star_starter_pack: %w", err)
	}
	report.add("product", "star_starter_pack", ok)
	return nil
}
