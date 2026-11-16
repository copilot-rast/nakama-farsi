-- +migrate Up
ALTER TABLE economy_item ADD COLUMN IF NOT EXISTS sound_url TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE economy_item DROP COLUMN IF EXISTS sound_url;
