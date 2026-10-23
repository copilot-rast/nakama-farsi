-- Operator purge (source admin-support.service.ts purge, PRD 12.2): personal
-- data is erased and the identity anonymized while business records stay.
-- +migrate Up
ALTER TABLE jazebe_user_account
    ADD COLUMN IF NOT EXISTS purged_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS purged_by_operator_id UUID REFERENCES users (id);

-- +migrate Down
-- The purge marker records an irreversible erasure.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261024020000-jazebe-account-purge'; END $$;
-- +migrate StatementEnd
