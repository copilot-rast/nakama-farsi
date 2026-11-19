-- +migrate Up notransaction
ALTER TABLE base_voice_settings ADD COLUMN IF NOT EXISTS user_tokens BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE base_voice_settings ADD COLUMN IF NOT EXISTS user_token_ttl_seconds INT NOT NULL DEFAULT 300 CHECK (user_token_ttl_seconds BETWEEN 60 AND 600);
ALTER TABLE base_voice_settings ADD COLUMN IF NOT EXISTS user_tokens_per_minute INT NOT NULL DEFAULT 10 CHECK (user_tokens_per_minute BETWEEN 1 AND 600);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the comms user token settings and cannot be reverted automatically; restore from backup', '20270104000000-base-comms-user-tokens'; END $$;
-- +migrate StatementEnd
