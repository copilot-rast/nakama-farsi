-- Leadership transfers are offered to a current administrator and accepted separately.
-- +migrate Up
ALTER TABLE jazebe_group ADD COLUMN creator_transferred_at TIMESTAMPTZ;
CREATE TABLE jazebe_group_leadership_offer (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    from_user_id UUID NOT NULL REFERENCES users(id),
    to_user_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(12) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','declined','cancelled','expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_jazebe_group_leadership_pending ON jazebe_group_leadership_offer(group_id) WHERE status='pending';
CREATE INDEX idx_jazebe_group_leadership_recipient ON jazebe_group_leadership_offer(to_user_id,status);

-- +migrate Down
-- Historic ownership transfers must survive rollbacks.
SELECT 1;
