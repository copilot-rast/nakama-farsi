-- Track the core ban owned by Jazebe so a later native Nakama ban is never
-- cleared by an operator's Jazebe unsuspend action.
-- +migrate Up
ALTER TABLE jazebe_user_account
    ADD COLUMN IF NOT EXISTS core_disable_time TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS session_revocation_pending BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_jazebe_account_revocation_pending
    ON jazebe_user_account(user_id) WHERE session_revocation_pending;

-- +migrate Down
-- Retain the provenance and pending work during rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023190000-jazebe-account-restriction'; END $$;
-- +migrate StatementEnd
