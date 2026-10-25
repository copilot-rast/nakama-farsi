-- Transaction-scoped lock rows. pg_advisory_xact_lock does not exist in
-- CockroachDB, so RPCs serialize by upserting a row keyed by a hash of the
-- lock name. One row per distinct key; rows hold no user data.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_xact_lock (
    lock_key BIGINT PRIMARY KEY,
    locked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
-- Refused: the Jazebe RPCs that serialize on this table would fail without it.
-- See docs/ANDRONIX-POSTGRESQL.md "Rolling back Jazebe migrations".
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % is required by running Jazebe RPCs and cannot be reverted automatically; use the documented rollback procedure', '20261025000000-jazebe-xact-lock'; END $$;
-- +migrate StatementEnd
