-- +migrate Up notransaction
ALTER TABLE base_category DROP CONSTRAINT IF EXISTS base_category_key_type_check;
ALTER TABLE base_category ADD CONSTRAINT base_category_key_type_check CHECK (key_type IN ('product','prize','gift','currency','faq','news','task','item','item_type'));
ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','gift','currency','news','faq','general','user','task','item'));

INSERT INTO base_category (id, key_type, key, name, sort) VALUES
(gen_random_uuid(), 'item_type', 'avatar', '{"fa":"آواتار","en":"Avatar","ar":"صورة رمزية"}', 1),
(gen_random_uuid(), 'item_type', 'frame', '{"fa":"قاب","en":"Frame","ar":"إطار"}', 2),
(gen_random_uuid(), 'item_type', 'skin', '{"fa":"پوسته","en":"Skin","ar":"مظهر"}', 3),
(gen_random_uuid(), 'item_type', 'board', '{"fa":"صفحهٔ بازی","en":"Board","ar":"لوحة"}', 4),
(gen_random_uuid(), 'item_type', 'piece', '{"fa":"مهره","en":"Piece","ar":"قطعة"}', 5),
(gen_random_uuid(), 'item_type', 'ball', '{"fa":"توپ","en":"Ball","ar":"كرة"}', 6),
(gen_random_uuid(), 'item_type', 'trail', '{"fa":"دنباله","en":"Trail","ar":"أثر"}', 7),
(gen_random_uuid(), 'item_type', 'emote', '{"fa":"شکلک","en":"Emote","ar":"تعبير"}', 8),
(gen_random_uuid(), 'item_type', 'other', '{"fa":"سایر","en":"Other","ar":"أخرى"}', 9)
ON CONFLICT (key_type, key) DO NOTHING;

CREATE TABLE IF NOT EXISTS economy_item (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    type_id UUID NOT NULL REFERENCES base_category (id),
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    category_id UUID REFERENCES base_category (id),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    rarity TEXT NOT NULL DEFAULT 'common' CHECK (rarity IN ('common','uncommon','rare','epic','legendary')),
    parent_key TEXT,
    requires_key TEXT,
    is_starter BOOLEAN NOT NULL DEFAULT false,
    level_table_id UUID REFERENCES economy_level_table (id),
    unlock_level INT CHECK (unlock_level IS NULL OR unlock_level >= 1),
    price JSONB,
    upgrades JSONB NOT NULL DEFAULT '[]',
    services JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS economy_item_list_idx ON economy_item (is_active, sort, key);

CREATE TABLE IF NOT EXISTS economy_user_item (
    user_id UUID NOT NULL,
    item_id UUID NOT NULL REFERENCES economy_item (id),
    source TEXT NOT NULL CHECK (source IN ('starter','buy','unlock','prize','admin')),
    source_ref TEXT NOT NULL DEFAULT '',
    owned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    upgrade_level INT NOT NULL DEFAULT 0 CHECK (upgrade_level >= 0),
    equipped BOOLEAN NOT NULL DEFAULT false,
    slot TEXT NOT NULL DEFAULT '',
    version INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, item_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS economy_user_item_slot_idx ON economy_user_item (user_id, slot) WHERE equipped;
CREATE INDEX IF NOT EXISTS economy_user_item_item_idx ON economy_user_item (item_id);

CREATE TABLE IF NOT EXISTS economy_item_request (
    user_id UUID NOT NULL,
    request_id TEXT NOT NULL,
    action TEXT NOT NULL,
    item_key TEXT NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    result JSONB NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, request_id)
);
CREATE INDEX IF NOT EXISTS economy_item_request_expiry_idx ON economy_item_request (expires_at);

ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS max_level INT CHECK (max_level IS NULL OR max_level >= 0);
ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS level_table_id UUID REFERENCES economy_level_table (id);
ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS is_one_time BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS check_contents BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS purchase_message JSONB NOT NULL DEFAULT '{}';

ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS once_key TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS economy_payment_once_idx ON economy_payment (once_key) WHERE once_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS economy_payment_user_product_idx ON economy_payment (user_id, product_id) WHERE status = 'success';

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'item_bought', '{"fa":"خرید آیتم","en":"Item bought","ar":"شراء عنصر"}', '{"fa":"وقتی بازیکن آیتمی را با سکه می‌خرد. مقدار آن قیمت آیتم است.","en":"Recorded when a player buys an item with currency. Its value is the price.","ar":"يُسجَّل عندما يشتري اللاعب عنصرًا. قيمته السعر."}', false, 0, 60, true, 5),
(gen_random_uuid(), 'item_unlocked', '{"fa":"باز شدن آیتم","en":"Item unlocked","ar":"فتح عنصر"}', '{"fa":"وقتی بازیکن با رسیدن به سطح، آیتمی را باز می‌کند. مقدار آن سطح لازم است.","en":"Recorded when a player unlocks an item by level. Its value is the required level.","ar":"يُسجَّل عندما يفتح اللاعب عنصرًا بالمستوى."}', false, 0, 60, true, 6),
(gen_random_uuid(), 'item_upgraded', '{"fa":"ارتقای آیتم","en":"Item upgraded","ar":"ترقية عنصر"}', '{"fa":"وقتی بازیکن آیتمی را ارتقا می‌دهد. مقدار آن سطح تازهٔ ارتقاست.","en":"Recorded when a player upgrades an item. Its value is the new upgrade level.","ar":"يُسجَّل عند ترقية عنصر. قيمته مستوى الترقية الجديد."}', false, 0, 60, true, 7),
(gen_random_uuid(), 'item_equipped', '{"fa":"انتخاب آیتم","en":"Item equipped","ar":"تجهيز عنصر"}', '{"fa":"وقتی بازیکن آیتمی را برای استفاده انتخاب می‌کند.","en":"Recorded when a player equips an item.","ar":"يُسجَّل عندما يجهّز اللاعب عنصرًا."}', false, NULL, 60, true, 8),
(gen_random_uuid(), 'item_used', '{"fa":"استفاده از خدمت آیتم","en":"Item service used","ar":"استخدام خدمة عنصر"}', '{"fa":"وقتی بازیکن از یکی از خدمت‌های آیتم استفاده می‌کند. مقدار آن تعداد واحدهاست.","en":"Recorded when a player uses an item service. Its value is the number of units.","ar":"يُسجَّل عند استخدام خدمة عنصر. قيمته عدد الوحدات."}', false, 0, 60, true, 9)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds the item catalog and player inventories and cannot be reverted automatically; restore from backup', '20261115000000-economy-items'; END $$;
-- +migrate StatementEnd
