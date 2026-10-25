-- Staff suspension record of a group (source jz_group suspended_at,
-- suspended_by_admin_id, suspended_reason, PRD US-38). The status check
-- already allows 'suspended'.
-- +migrate Up notransaction
ALTER TABLE jazebe_group
    ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS suspended_by_user_id UUID,
    ADD COLUMN IF NOT EXISTS suspended_reason VARCHAR(512),
    ADD COLUMN IF NOT EXISTS disband_reason VARCHAR(512);
ALTER TABLE jazebe_group DROP CONSTRAINT IF EXISTS jazebe_group_suspended_by_user_id_fkey;
ALTER TABLE jazebe_group ADD CONSTRAINT jazebe_group_suspended_by_user_id_fkey FOREIGN KEY (suspended_by_user_id) REFERENCES users(id);

-- +migrate Down
-- Suspension records are moderation history.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261024050000-jazebe-group-suspension'; END $$;
-- +migrate StatementEnd
