-- Mission definitions are scoped to a Jazebe planet; completion and rewards follow separately.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_mission (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    title VARCHAR(128) NOT NULL,
    short_description VARCHAR(512),
    long_description TEXT,
    icon VARCHAR(64),
    functional_type VARCHAR(24) NOT NULL CHECK (functional_type IN ('media_view','code_entry','content_creation','invite','external_verification','social_action')),
    target_type VARCHAR(8) NOT NULL DEFAULT 'solo' CHECK (target_type IN ('solo','group')),
    repeat_pattern VARCHAR(12) NOT NULL DEFAULT 'once' CHECK (repeat_pattern IN ('once','periodic','unlimited')),
    repeat_period VARCHAR(8) CHECK (repeat_period IN ('daily','weekly','monthly')),
    detection_method VARCHAR(16) NOT NULL CHECK (detection_method IN ('platform_auto','brand_api','manual_review')),
    reward_points INTEGER NOT NULL DEFAULT 0 CHECK (reward_points >= 0),
    reward_tickets INTEGER NOT NULL DEFAULT 0 CHECK (reward_tickets >= 0),
    reward_xp INTEGER NOT NULL DEFAULT 0 CHECK (reward_xp >= 0),
    target_lottery_id UUID,
    group_target_value INTEGER,
    display_conditions JSONB,
    type_config JSONB NOT NULL DEFAULT '{}',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(12) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','paused','expired','suspended')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_by UUID REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT jazebe_mission_group_no_points CHECK (target_type <> 'group' OR reward_points = 0),
    CONSTRAINT jazebe_mission_group_target CHECK (target_type <> 'group' OR group_target_value IS NOT NULL),
    CONSTRAINT jazebe_mission_periodic_period CHECK (repeat_pattern <> 'periodic' OR repeat_period IS NOT NULL),
    CONSTRAINT jazebe_mission_time CHECK (ends_at > starts_at)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_mission_planet_status ON jazebe_mission (planet_id, status);
CREATE INDEX IF NOT EXISTS idx_jazebe_mission_planet_sort ON jazebe_mission (planet_id, sort_order);

-- +migrate Down
-- Retain mission definitions and linked activity for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022160000-jazebe-mission'; END $$;
-- +migrate StatementEnd
