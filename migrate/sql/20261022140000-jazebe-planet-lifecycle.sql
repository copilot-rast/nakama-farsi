-- Add Jazebe planet draft lifecycle and its default settings without changing old active rows.
-- +migrate Up
ALTER TABLE jazebe_planet DROP CONSTRAINT IF EXISTS jazebe_planet_status_check;
UPDATE jazebe_planet SET status = 'suspended' WHERE status = 'disabled';
ALTER TABLE jazebe_planet ADD CONSTRAINT jazebe_planet_status_check CHECK (status IN ('setup', 'pending_review', 'active', 'suspended', 'archived'));
ALTER TABLE jazebe_planet ALTER COLUMN status SET DEFAULT 'setup';
ALTER TABLE jazebe_planet ALTER COLUMN name TYPE VARCHAR(128);
ALTER TABLE jazebe_planet ADD COLUMN slug VARCHAR(64);
ALTER TABLE jazebe_planet ADD COLUMN latin_name VARCHAR(64);
ALTER TABLE jazebe_planet ADD COLUMN short_description VARCHAR(512);
ALTER TABLE jazebe_planet ADD COLUMN category_id UUID;
ALTER TABLE jazebe_planet ADD COLUMN point_name VARCHAR(32) NOT NULL DEFAULT 'امتیاز';
ALTER TABLE jazebe_planet ADD COLUMN brand_color VARCHAR(9);
ALTER TABLE jazebe_planet ADD COLUMN icon_asset VARCHAR(512);
ALTER TABLE jazebe_planet ADD COLUMN glyph VARCHAR(32);
ALTER TABLE jazebe_planet ADD COLUMN activated_at TIMESTAMPTZ;
ALTER TABLE jazebe_planet ADD COLUMN suspended_at TIMESTAMPTZ;
ALTER TABLE jazebe_planet ADD COLUMN archived_at TIMESTAMPTZ;
ALTER TABLE jazebe_planet ADD COLUMN created_by UUID REFERENCES users (id);
CREATE UNIQUE INDEX uq_jazebe_planet_slug ON jazebe_planet (slug) WHERE slug IS NOT NULL;
CREATE TABLE jazebe_planet_category (id UUID PRIMARY KEY, title_fa VARCHAR(128) NOT NULL, is_active BOOLEAN NOT NULL DEFAULT true);
ALTER TABLE jazebe_planet ADD CONSTRAINT jazebe_planet_category_fk FOREIGN KEY (category_id) REFERENCES jazebe_planet_category (id);
CREATE TABLE jazebe_planet_settings (
    planet_id UUID PRIMARY KEY REFERENCES jazebe_planet (id),
    mission_reward_point_min INTEGER NOT NULL DEFAULT 0,
    mission_reward_point_max INTEGER NOT NULL DEFAULT 100,
    mission_reward_ticket_max INTEGER NOT NULL DEFAULT 5,
    ticket_price_min INTEGER NOT NULL DEFAULT 1,
    ticket_price_max INTEGER NOT NULL DEFAULT 1000,
    lottery_fee_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
    group_member_cap_min INTEGER NOT NULL DEFAULT 2,
    group_member_cap_max INTEGER NOT NULL DEFAULT 20,
    group_max_count INTEGER,
    group_admin_cap INTEGER NOT NULL DEFAULT 2,
    group_creation_enabled BOOLEAN NOT NULL DEFAULT true,
    allowed_group_modes JSONB NOT NULL DEFAULT '["open","approval"]',
    xp_level_thresholds JSONB NOT NULL DEFAULT '[]',
    free_game_daily_cap INTEGER,
    xp_per_purchase INTEGER NOT NULL DEFAULT 0,
    xp_per_group_join INTEGER NOT NULL DEFAULT 0,
    display_name_change_limit_per_month INTEGER NOT NULL DEFAULT 3,
    point_to_toman_rate INTEGER NOT NULL DEFAULT 0
);
INSERT INTO jazebe_planet_settings (planet_id) SELECT id FROM jazebe_planet ON CONFLICT DO NOTHING;

-- +migrate Down
-- Retain draft and settings data for manual rollback review.
SELECT 1;
