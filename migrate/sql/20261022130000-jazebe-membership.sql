-- Extend the initial native planet schema to preserve Jazebe membership semantics.
-- +migrate Up
ALTER TABLE jazebe_planet ADD COLUMN IF NOT EXISTS fans_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE jazebe_planet_membership ADD COLUMN IF NOT EXISTS join_path VARCHAR(24) NOT NULL DEFAULT 'join_button';
ALTER TABLE jazebe_planet_membership ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE jazebe_planet_membership ADD COLUMN IF NOT EXISTS last_activity_at TIMESTAMPTZ;
ALTER TABLE jazebe_planet_membership ADD COLUMN IF NOT EXISTS sort_order INTEGER;
ALTER TABLE jazebe_planet_membership DROP CONSTRAINT IF EXISTS jazebe_membership_join_path_check, ADD CONSTRAINT jazebe_membership_join_path_check CHECK (join_path IN ('brand_domain_signup', 'join_button', 'first_activity', 'group_invite_link'));
UPDATE jazebe_planet p SET fans_count = (SELECT count(*) FROM jazebe_planet_membership m WHERE m.planet_id = p.id AND m.is_active);
CREATE INDEX IF NOT EXISTS idx_jazebe_membership_user_activity ON jazebe_planet_membership (user_id, last_activity_at DESC NULLS LAST);

-- +migrate Down
-- Retain membership data and provenance. Manual rollback requires a data review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022130000-jazebe-membership'; END $$;
-- +migrate StatementEnd
