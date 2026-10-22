-- Generated code stock for code_entry missions.
-- +migrate Up
CREATE TABLE jazebe_mission_code (
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
CREATE INDEX idx_jazebe_code_planet_code ON jazebe_mission_code (planet_id,code);
-- +migrate Down
-- Keep issued code stock and redemption evidence for manual rollback.
SELECT 1;
