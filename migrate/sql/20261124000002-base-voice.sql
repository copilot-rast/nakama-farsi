-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_voice_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT false,
    provider TEXT NOT NULL DEFAULT '' CHECK (provider ~ '^[a-z0-9_]{0,32}$'),
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    secrets_sealed TEXT NOT NULL DEFAULT '',
    ttl_seconds INT NOT NULL DEFAULT 600 CHECK (ttl_seconds BETWEEN 60 AND 86400),
    open_rooms BOOLEAN NOT NULL DEFAULT false,
    can_publish BOOLEAN NOT NULL DEFAULT true,
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO base_voice_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

ALTER TABLE economy_game ADD COLUMN IF NOT EXISTS voice_enabled BOOLEAN NOT NULL DEFAULT true;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the voice server setting with its sealed secrets and cannot be reverted automatically; restore from backup', '20261124000002-base-voice'; END $$;
-- +migrate StatementEnd
