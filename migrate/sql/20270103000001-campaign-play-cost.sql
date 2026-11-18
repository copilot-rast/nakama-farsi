-- +migrate Up
ALTER TABLE campaign_attempt ADD COLUMN IF NOT EXISTS play_cost JSONB;

-- +migrate Down
ALTER TABLE campaign_attempt DROP COLUMN IF EXISTS play_cost;
