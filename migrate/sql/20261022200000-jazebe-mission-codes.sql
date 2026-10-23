-- Generated code stock for code_entry missions.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_mission_code (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    mission_id UUID NOT NULL REFERENCES jazebe_mission (id),
    code VARCHAR(64) NOT NULL,
    mode VARCHAR(8) NOT NULL CHECK (mode IN ('unique','shared')),
    max_uses INTEGER NOT NULL CHECK (max_uses > 0),
    used_count INTEGER NOT NULL DEFAULT 0 CHECK (used_count >= 0 AND used_count <= max_uses),
    assigned_user_id UUID REFERENCES users (id),
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (mission_id,code),
    CONSTRAINT jazebe_code_unique_uses CHECK (mode <> 'unique' OR max_uses = 1)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_code_planet_code ON jazebe_mission_code (planet_id,code);
-- +migrate Down
-- Keep issued code stock and redemption evidence for manual rollback.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022200000-jazebe-mission-codes'; END $$;
-- +migrate StatementEnd
