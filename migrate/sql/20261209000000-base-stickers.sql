-- +migrate Up
CREATE TABLE IF NOT EXISTS base_sticker_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_sticker_pack (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    access TEXT NOT NULL DEFAULT 'free' CHECK (access IN ('free','item')),
    item_key TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (access = 'free' OR item_key <> '')
);

CREATE TABLE IF NOT EXISTS base_sticker (
    id UUID PRIMARY KEY,
    pack_id UUID NOT NULL REFERENCES base_sticker_pack (id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    name JSONB NOT NULL DEFAULT '{}',
    image_url TEXT NOT NULL,
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE (pack_id, key)
);

CREATE INDEX IF NOT EXISTS base_sticker_pack_sort_idx ON base_sticker_pack (sort, key);
CREATE INDEX IF NOT EXISTS base_sticker_pack_idx ON base_sticker (pack_id, sort, key);

-- +migrate Down
DROP TABLE IF EXISTS base_sticker;
DROP TABLE IF EXISTS base_sticker_pack;
DROP TABLE IF EXISTS base_sticker_settings;
