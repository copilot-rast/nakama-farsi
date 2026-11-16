-- +migrate Up
CREATE TABLE IF NOT EXISTS base_avatar_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_avatar_part (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    layer INT NOT NULL DEFAULT 0,
    required BOOLEAN NOT NULL DEFAULT false,
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_avatar_option (
    id UUID PRIMARY KEY,
    part_id UUID NOT NULL REFERENCES base_avatar_part (id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    name JSONB NOT NULL DEFAULT '{}',
    image_url TEXT NOT NULL DEFAULT '',
    icon_url TEXT NOT NULL DEFAULT '',
    value TEXT NOT NULL DEFAULT '',
    access TEXT NOT NULL DEFAULT 'free' CHECK (access IN ('free','item')),
    item_key TEXT NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT false,
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE (part_id, key),
    CHECK (access = 'free' OR item_key <> '')
);

CREATE TABLE IF NOT EXISTS base_avatar_look (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    look JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS base_avatar_part_sort_idx ON base_avatar_part (sort, key);
CREATE INDEX IF NOT EXISTS base_avatar_option_part_idx ON base_avatar_option (part_id, sort, key);

-- +migrate Down
DROP TABLE IF EXISTS base_avatar_look;
DROP TABLE IF EXISTS base_avatar_option;
DROP TABLE IF EXISTS base_avatar_part;
DROP TABLE IF EXISTS base_avatar_settings;
