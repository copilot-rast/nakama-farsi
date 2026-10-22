-- Preserve the actor of a disband operation for audit and source response parity.
-- +migrate Up
ALTER TABLE jazebe_group ADD COLUMN disbanded_by_user_id UUID REFERENCES users(id);

-- +migrate Down
-- Disband history is intentionally retained.
SELECT 1;
