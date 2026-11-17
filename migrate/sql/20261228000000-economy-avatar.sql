-- +migrate Up
CREATE TABLE IF NOT EXISTS economy_avatar_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_avatar_type (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    image_url TEXT NOT NULL DEFAULT '',
    icon_url TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_avatar_category (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    layer INT NOT NULL DEFAULT 0,
    required BOOLEAN NOT NULL DEFAULT false,
    slots INT NOT NULL DEFAULT 1 CHECK (slots BETWEEN 1 AND 20),
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_avatar_item (
    id UUID PRIMARY KEY,
    category_id UUID NOT NULL REFERENCES economy_avatar_category (id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    name JSONB NOT NULL DEFAULT '{}',
    image_url TEXT NOT NULL DEFAULT '',
    icon_url TEXT NOT NULL DEFAULT '',
    value TEXT NOT NULL DEFAULT '',
    access TEXT NOT NULL DEFAULT 'free' CHECK (access IN ('free','currency','product','item')),
    price JSONB,
    product_id UUID REFERENCES economy_product (id) ON DELETE RESTRICT,
    item_key TEXT NOT NULL DEFAULT '',
    requirements JSONB,
    type_ids TEXT[] NOT NULL DEFAULT '{}',
    is_default BOOLEAN NOT NULL DEFAULT false,
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (category_id, key),
    CHECK (access <> 'currency' OR price IS NOT NULL),
    CHECK (access <> 'product' OR product_id IS NOT NULL),
    CHECK (access <> 'item' OR item_key <> '')
);

CREATE TABLE IF NOT EXISTS economy_avatar_unlock (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES economy_avatar_item (id) ON DELETE CASCADE,
    source TEXT NOT NULL DEFAULT 'buy',
    paid JSONB,
    request_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE IF NOT EXISTS economy_avatar (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type_id UUID REFERENCES economy_avatar_type (id) ON DELETE RESTRICT,
    name TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT false,
    cleared TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_avatar_equip (
    avatar_id UUID NOT NULL REFERENCES economy_avatar (id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES economy_avatar_item (id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES economy_avatar_category (id) ON DELETE CASCADE,
    equipped_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (avatar_id, item_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS economy_avatar_active_idx ON economy_avatar (user_id) WHERE is_active;
CREATE INDEX IF NOT EXISTS economy_avatar_user_idx ON economy_avatar (user_id, created_at);
CREATE INDEX IF NOT EXISTS economy_avatar_type_idx ON economy_avatar (type_id);
CREATE INDEX IF NOT EXISTS economy_avatar_item_category_idx ON economy_avatar_item (category_id, sort, key);
CREATE INDEX IF NOT EXISTS economy_avatar_equip_category_idx ON economy_avatar_equip (category_id);
CREATE INDEX IF NOT EXISTS economy_avatar_equip_item_idx ON economy_avatar_equip (item_id);
CREATE INDEX IF NOT EXISTS economy_avatar_unlock_item_idx ON economy_avatar_unlock (item_id);

INSERT INTO economy_avatar_settings (id, settings, updated_by, updated_at)
SELECT 1, jsonb_build_object('enabled', COALESCE((s.settings->>'enabled')::boolean, true)), s.updated_by, s.updated_at
FROM base_avatar_settings s WHERE s.id = 1
ON CONFLICT (id) DO NOTHING;

INSERT INTO economy_avatar_category (id, key, name, layer, required, slots, sort, is_active, created_by, created_at, updated_at)
SELECT p.id, p.key, p.name, p.layer, p.required, 1, p.sort, p.is_active, p.created_by, p.created_at, p.updated_at
FROM base_avatar_part p
ON CONFLICT DO NOTHING;

INSERT INTO economy_avatar_item (id, category_id, key, name, image_url, icon_url, value, access, item_key, is_default, sort, is_active)
SELECT o.id, o.part_id, o.key, o.name, o.image_url, o.icon_url, o.value, o.access, o.item_key, o.is_default, o.sort, o.is_active
FROM base_avatar_option o
ON CONFLICT DO NOTHING;

INSERT INTO economy_avatar (id, user_id, name, is_active, cleared, created_at, updated_at)
SELECT md5('avatar:' || l.user_id::text)::uuid, l.user_id, '', true,
    ARRAY(SELECT p.id::text FROM base_avatar_part p WHERE l.look ? p.key AND l.look->>p.key = '' AND NOT p.required ORDER BY p.key),
    l.updated_at, l.updated_at
FROM base_avatar_look l
WHERE NOT EXISTS (SELECT 1 FROM economy_avatar a WHERE a.user_id = l.user_id)
ON CONFLICT DO NOTHING;

INSERT INTO economy_avatar_equip (avatar_id, item_id, category_id, equipped_at)
SELECT md5('avatar:' || l.user_id::text)::uuid, o.id, p.id, l.updated_at
FROM base_avatar_look l
JOIN base_avatar_part p ON l.look ? p.key
JOIN base_avatar_option o ON o.part_id = p.id AND o.key = l.look->>p.key
WHERE EXISTS (SELECT 1 FROM economy_avatar a WHERE a.id = md5('avatar:' || l.user_id::text)::uuid)
ON CONFLICT DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS economy_avatar_equip;
DROP TABLE IF EXISTS economy_avatar;
DROP TABLE IF EXISTS economy_avatar_unlock;
DROP TABLE IF EXISTS economy_avatar_item;
DROP TABLE IF EXISTS economy_avatar_category;
DROP TABLE IF EXISTS economy_avatar_type;
DROP TABLE IF EXISTS economy_avatar_settings;
