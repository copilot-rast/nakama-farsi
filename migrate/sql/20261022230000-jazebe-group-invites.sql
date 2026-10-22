-- Product availability and planet group links. Tokens are stored as hashes.
-- +migrate Up
ALTER TABLE jazebe_brand ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
CREATE TABLE jazebe_group_invite (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    use_count INTEGER NOT NULL DEFAULT 0 CHECK (use_count >= 0),
    max_uses INTEGER CHECK (max_uses IS NULL OR max_uses > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_jazebe_group_invite_group ON jazebe_group_invite(group_id);

-- +migrate Down
-- Keep issued links and brand state for manual review.
SELECT 1;
