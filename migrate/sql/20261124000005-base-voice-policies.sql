-- +migrate Up notransaction
ALTER TABLE base_voice_settings ADD COLUMN IF NOT EXISTS allowed_tag_ids TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE base_voice_settings ADD COLUMN IF NOT EXISTS listen_only_tag_ids TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE base_voice_settings ADD COLUMN IF NOT EXISTS tokens_per_minute INT NOT NULL DEFAULT 30 CHECK (tokens_per_minute BETWEEN 1 AND 600);
ALTER TABLE base_voice_settings ADD COLUMN IF NOT EXISTS max_room_players INT NOT NULL DEFAULT 0 CHECK (max_room_players BETWEEN 0 AND 1000);
ALTER TABLE economy_game ADD COLUMN IF NOT EXISTS voice_max_players INT CHECK (voice_max_players BETWEEN 0 AND 1000);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the voice policies and cannot be reverted automatically; restore from backup', '20261124000005-base-voice-policies'; END $$;
-- +migrate StatementEnd
