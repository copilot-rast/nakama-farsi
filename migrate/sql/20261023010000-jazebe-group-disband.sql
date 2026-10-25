-- Preserve the actor of a disband operation for audit and source response parity.
-- +migrate Up notransaction
ALTER TABLE jazebe_group ADD COLUMN IF NOT EXISTS disbanded_by_user_id UUID;
ALTER TABLE jazebe_group DROP CONSTRAINT IF EXISTS jazebe_group_disbanded_by_user_id_fkey;
ALTER TABLE jazebe_group ADD CONSTRAINT jazebe_group_disbanded_by_user_id_fkey FOREIGN KEY (disbanded_by_user_id) REFERENCES users(id);

-- +migrate Down
-- Disband history is intentionally retained.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023010000-jazebe-group-disband'; END $$;
-- +migrate StatementEnd
