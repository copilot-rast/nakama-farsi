-- Source planet identity fields used by brand editing.
-- +migrate Up notransaction
ALTER TABLE jazebe_planet ADD COLUMN IF NOT EXISTS long_description TEXT;
ALTER TABLE jazebe_planet ADD COLUMN IF NOT EXISTS logo_asset VARCHAR(512);
ALTER TABLE jazebe_planet ADD COLUMN IF NOT EXISTS cover_asset VARCHAR(512);
ALTER TABLE jazebe_planet ADD COLUMN IF NOT EXISTS deployment_mode VARCHAR(16) NOT NULL DEFAULT 'cloud';
ALTER TABLE jazebe_planet DROP CONSTRAINT IF EXISTS jazebe_planet_deployment_mode_check;
ALTER TABLE jazebe_planet ADD CONSTRAINT jazebe_planet_deployment_mode_check CHECK (deployment_mode IN ('cloud', 'on_premise'));

-- +migrate Down
-- Retain identity data for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022150000-jazebe-planet-identity'; END $$;
-- +migrate StatementEnd
