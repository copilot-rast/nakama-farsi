-- +migrate Up
ALTER TABLE base_sticker_pack ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'sticker';
ALTER TABLE base_sticker_pack ADD COLUMN IF NOT EXISTS in_chat BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE base_sticker_pack ADD COLUMN IF NOT EXISTS in_match BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE base_sticker_pack DROP CONSTRAINT IF EXISTS base_sticker_pack_kind_check;
ALTER TABLE base_sticker_pack ADD CONSTRAINT base_sticker_pack_kind_check CHECK (kind IN ('sticker','text','gif','emoji'));
ALTER TABLE base_sticker ADD COLUMN IF NOT EXISTS text JSONB NOT NULL DEFAULT '{}';
ALTER TABLE base_sticker ADD COLUMN IF NOT EXISTS asset_id TEXT NOT NULL DEFAULT '';
ALTER TABLE base_sticker ALTER COLUMN image_url SET DEFAULT '';

CREATE TABLE IF NOT EXISTS base_comm_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_comm_favorite (
    user_id UUID NOT NULL,
    pack_id UUID NOT NULL REFERENCES base_sticker_pack (id) ON DELETE CASCADE,
    sticker_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, pack_id, sticker_key)
);
CREATE INDEX IF NOT EXISTS base_comm_favorite_user_idx ON base_comm_favorite (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS base_comm_channel (
    match_id TEXT PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','closed')),
    close_reason TEXT NOT NULL DEFAULT '',
    closed_by TEXT NOT NULL DEFAULT '',
    closed_at TIMESTAMPTZ,
    messages INT NOT NULL DEFAULT 0,
    last_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS base_comm_channel_status_idx ON base_comm_channel (status, created_at DESC);

CREATE TABLE IF NOT EXISTS base_comm_member (
    match_id TEXT NOT NULL REFERENCES base_comm_channel (match_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    can_send BOOLEAN NOT NULL DEFAULT true,
    sent INT NOT NULL DEFAULT 0,
    changed_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, user_id)
);

-- +migrate Down
DROP TABLE IF EXISTS base_comm_member;
DROP TABLE IF EXISTS base_comm_channel;
DROP TABLE IF EXISTS base_comm_favorite;
DROP TABLE IF EXISTS base_comm_settings;
ALTER TABLE base_sticker DROP COLUMN IF EXISTS asset_id;
ALTER TABLE base_sticker DROP COLUMN IF EXISTS text;
ALTER TABLE base_sticker_pack DROP CONSTRAINT IF EXISTS base_sticker_pack_kind_check;
ALTER TABLE base_sticker_pack DROP COLUMN IF EXISTS in_match;
ALTER TABLE base_sticker_pack DROP COLUMN IF EXISTS in_chat;
ALTER TABLE base_sticker_pack DROP COLUMN IF EXISTS kind;
