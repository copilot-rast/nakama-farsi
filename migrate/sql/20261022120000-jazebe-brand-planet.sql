-- Jazebe brand and planet foundation. Nakama user UUIDs are the only user IDs.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_brand (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_jazebe_brand_owner ON jazebe_brand (owner_id, created_at DESC);

CREATE TABLE IF NOT EXISTS jazebe_planet (
    id UUID PRIMARY KEY,
    brand_id UUID NOT NULL REFERENCES jazebe_brand (id),
    name VARCHAR(120) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT jazebe_planet_status_check CHECK (status IN ('active', 'disabled')),
    CONSTRAINT jazebe_planet_brand_name UNIQUE (brand_id, name)
);

CREATE INDEX IF NOT EXISTS idx_jazebe_planet_status_created ON jazebe_planet (status, created_at DESC);

CREATE TABLE IF NOT EXISTS jazebe_planet_membership (
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (planet_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_jazebe_planet_membership_user ON jazebe_planet_membership (user_id, joined_at DESC);

-- +migrate Down
-- Deliberately retain Jazebe user data. Manual rollback requires a data review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022120000-jazebe-brand-planet'; END $$;
-- +migrate StatementEnd
