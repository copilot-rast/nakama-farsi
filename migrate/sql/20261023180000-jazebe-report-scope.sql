-- Scope open reports to the planet. The older uniqueness key prevented the
-- same reporter from reporting one person in a different planet.
-- +migrate Up
ALTER TABLE jazebe_user_report DROP CONSTRAINT IF EXISTS jazebe_user_report_subject_type_check;
UPDATE jazebe_user_report SET subject_type='profile' WHERE subject_type='user';
ALTER TABLE jazebe_user_report DROP CONSTRAINT IF EXISTS jazebe_user_report_subject_type_check, ADD CONSTRAINT jazebe_user_report_subject_type_check
    CHECK (subject_type IN ('message','group','profile'));
DROP INDEX IF EXISTS uq_jazebe_user_report_open;
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_user_report_open ON jazebe_user_report
    (planet_id,reporter_user_id,subject_type,subject_id)
    WHERE status IN ('open','reviewing') AND subject_type IN ('message','group');
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_profile_report_open ON jazebe_user_report
    (planet_id,reporter_user_id,subject_type,subject_id)
    WHERE status='open' AND subject_type='profile';
CREATE INDEX IF NOT EXISTS idx_jazebe_user_report_queue ON jazebe_user_report
    (status,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_jazebe_user_report_planet_queue ON jazebe_user_report
    (planet_id,status,created_at DESC,id DESC);
-- +migrate Down
-- Retain planet-scoped report history and the corrected uniqueness rule.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023180000-jazebe-report-scope'; END $$;
-- +migrate StatementEnd
