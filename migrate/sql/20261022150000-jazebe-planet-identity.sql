-- Source planet identity fields used by brand editing.
-- +migrate Up
ALTER TABLE jazebe_planet ADD COLUMN long_description TEXT;
ALTER TABLE jazebe_planet ADD COLUMN logo_asset VARCHAR(512);
ALTER TABLE jazebe_planet ADD COLUMN cover_asset VARCHAR(512);
ALTER TABLE jazebe_planet ADD COLUMN deployment_mode VARCHAR(16) NOT NULL DEFAULT 'cloud';
ALTER TABLE jazebe_planet ADD CONSTRAINT jazebe_planet_deployment_mode_check CHECK (deployment_mode IN ('cloud', 'on_premise'));

-- +migrate Down
-- Retain identity data for manual rollback review.
SELECT 1;
