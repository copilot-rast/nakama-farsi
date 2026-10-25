-- Mission medals are processed after the approved completion commits.
-- +migrate Up notransaction
-- An inline CHECK is named jazebe_progression_event_kind_check by PostgreSQL and check_kind by CockroachDB.
ALTER TABLE jazebe_progression_event DROP CONSTRAINT IF EXISTS check_kind;
ALTER TABLE jazebe_progression_event DROP CONSTRAINT IF EXISTS jazebe_progression_event_kind_check;
ALTER TABLE jazebe_progression_event ADD CONSTRAINT jazebe_progression_event_kind_check
    CHECK (kind IN ('purchase','group_join','mission')) NOT VALID;
-- +migrate Down
-- Retain pending mission events; restoring the narrower check would discard work.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023140000-jazebe-mission-medal-events'; END $$;
-- +migrate StatementEnd
