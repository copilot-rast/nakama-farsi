-- Track the core ban owned by Jazebe so a later native Nakama ban is never
-- cleared by an operator's Jazebe unsuspend action.
-- +migrate Up
ALTER TABLE jazebe_user_account
    ADD COLUMN core_disable_time TIMESTAMPTZ,
    ADD COLUMN session_revocation_pending BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX idx_jazebe_account_revocation_pending
    ON jazebe_user_account(user_id) WHERE session_revocation_pending;

-- +migrate Down
-- Retain the provenance and pending work during rollback review.
SELECT 1;
