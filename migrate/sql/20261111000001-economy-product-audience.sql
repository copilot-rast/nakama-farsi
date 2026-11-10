-- +migrate Up
ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS audience_tag_ids TEXT[] NOT NULL DEFAULT '{}';

-- +migrate Down
ALTER TABLE economy_product DROP COLUMN IF EXISTS audience_tag_ids;
