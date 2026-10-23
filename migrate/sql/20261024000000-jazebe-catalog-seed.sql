-- Port the Jazebe planet category catalogue (jz_planet_category: key, title_en,
-- sort_order, soft delete) and the source seed data in
-- seed/jazebeh-seed.data.ts (PLANET_CATEGORY_SEED, MEDAL_LIBRARY_SEED).
-- Without categories no planet can pass the publish check. Safe to re-run;
-- existing rows and operator edits are never overwritten.

-- +migrate Up
ALTER TABLE jazebe_planet_category
    ADD COLUMN IF NOT EXISTS key VARCHAR(64),
    ADD COLUMN IF NOT EXISTS title_en VARCHAR(128),
    ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users (id),
    ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users (id),
    ADD COLUMN IF NOT EXISTS deleted_by UUID REFERENCES users (id),
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Categories created before keys existed keep a stable, unique derived key.
UPDATE jazebe_planet_category SET key = 'legacy_' || replace(id::text, '-', '') WHERE key IS NULL;
ALTER TABLE jazebe_planet_category ALTER COLUMN key SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_planet_category_key ON jazebe_planet_category (key);
ALTER TABLE jazebe_planet_category
    DROP CONSTRAINT IF EXISTS jazebe_planet_category_key_format,
    ADD CONSTRAINT jazebe_planet_category_key_format CHECK (key ~ '^[a-z0-9_]{2,64}$');

INSERT INTO jazebe_planet_category (id, key, title_fa, title_en, sort_order) VALUES
    (gen_random_uuid(), 'food', 'خوراکی', 'Food', 10),
    (gen_random_uuid(), 'food_delivery', 'سفارش غذا', 'Food delivery', 20),
    (gen_random_uuid(), 'fashion', 'پوشاک', 'Fashion', 30),
    (gen_random_uuid(), 'retail', 'فروشگاهی', 'Retail', 40),
    (gen_random_uuid(), 'transport', 'حمل و نقل', 'Transport', 50),
    (gen_random_uuid(), 'telecom', 'اپراتور', 'Telecom operator', 60),
    (gen_random_uuid(), 'banking', 'بانکداری', 'Banking', 70),
    (gen_random_uuid(), 'digital_services', 'خدمات دیجیتال', 'Digital services', 80),
    (gen_random_uuid(), 'apps', 'اپلیکیشن', 'Apps', 90),
    (gen_random_uuid(), 'video', 'ویدئو', 'Video', 100),
    (gen_random_uuid(), 'classifieds', 'آگهی', 'Classifieds', 110),
    (gen_random_uuid(), 'other', 'سایر', 'Other', 999)
ON CONFLICT (key) DO NOTHING;

-- Library medals are owned by Nakama's system user (initial schema).
INSERT INTO jazebe_medal_library (id, key, name, description, icon, theme_category, condition_type, condition_param_schema, created_by) VALUES
    (gen_random_uuid(), 'first_purchase', 'اولین خرید', 'اولین خرید از فروشگاه سیاره', 'target', 'purchase', 'activity_count',
        '{"activity":"purchase","count":{"min":1,"max":1}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'discount_hunter', 'شکارچی تخفیف', 'چند خرید از فروشگاه سیاره', 'target', 'purchase', 'activity_count',
        '{"activity":"purchase","count":{"min":2,"max":100}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'loyal', 'وفادار', 'مدت عضویت در سیاره', 'heart', 'participation', 'duration',
        '{"unit":"day","days":{"min":30,"max":730}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'one_year_member', 'یک سال عضویت', 'یک سال عضویت پیوسته در سیاره', 'heart', 'participation', 'duration',
        '{"unit":"day","days":{"min":365,"max":365}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'invite_friends', 'دعوت دوستان', 'دعوت دوستان به سیاره', 'compass', 'social', 'activity_count',
        '{"activity":"referral","count":{"min":1,"max":50}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'group_captain', 'کاپیتان گروه', 'ساخت گروه و رساندن آن به حد نصاب اعضا', 'star', 'social', 'activity_count',
        '{"activity":"group_members","count":{"min":2,"max":100}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'explorer', 'کاوشگر', 'انجام ماموریت‌های سیاره', 'compass', 'participation', 'activity_count',
        '{"activity":"mission","count":{"min":1,"max":100}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'speedster', 'سریع‌السیر', 'انجام بازی‌های سیاره', 'bolt', 'progression', 'activity_count',
        '{"activity":"game","count":{"min":1,"max":100}}', '00000000-0000-0000-0000-000000000000'),
    (gen_random_uuid(), 'special_event', 'رویداد ویژه', 'حضور در رویداد ویژهٔ برند', 'star', 'participation', 'special_event',
        '{"event":"string"}', '00000000-0000-0000-0000-000000000000')
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- Retain categories and medals: planets and medals reference them by ID.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261024000000-jazebe-catalog-seed'; END $$;
-- +migrate StatementEnd
