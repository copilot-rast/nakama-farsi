-- +migrate Up
ALTER TABLE campaign_attempt ADD COLUMN IF NOT EXISTS item_key TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE campaign_attempt DROP COLUMN IF EXISTS item_key;
