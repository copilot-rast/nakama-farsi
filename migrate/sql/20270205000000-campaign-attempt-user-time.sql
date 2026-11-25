-- +migrate Up
CREATE INDEX IF NOT EXISTS campaign_attempt_user_started_idx ON campaign_attempt (user_id, started_at);

-- +migrate Down
DROP INDEX IF EXISTS campaign_attempt_user_started_idx;
