-- +migrate Up
ALTER TABLE content_pack ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'addressables' CHECK (kind IN ('addressables', 'files'));

-- +migrate Down
ALTER TABLE content_pack DROP COLUMN IF EXISTS kind;
