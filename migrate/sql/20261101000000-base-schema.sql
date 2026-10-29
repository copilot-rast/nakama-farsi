-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_language (
    code TEXT PRIMARY KEY CHECK (code ~ '^[a-z]{2,3}(-[a-z0-9]{2,8})?$'),
    name JSONB NOT NULL DEFAULT '{}',
    native_name TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT 'ltr' CHECK (direction IN ('rtl','ltr')),
    icon_url TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (is_active OR NOT is_default)
);
CREATE UNIQUE INDEX IF NOT EXISTS base_language_default_idx ON base_language (is_default) WHERE is_default;

INSERT INTO base_language (code, name, native_name, direction, is_active, is_default, sort)
VALUES ('fa', '{"fa":"فارسی","en":"Persian"}', 'فارسی', 'rtl', true, true, 1)
ON CONFLICT DO NOTHING;
INSERT INTO base_language (code, name, native_name, direction, is_active, is_default, sort)
VALUES ('en', '{"fa":"انگلیسی","en":"English"}', 'English', 'ltr', true, false, 2)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS base_category (
    id UUID PRIMARY KEY,
    key_type TEXT NOT NULL CHECK (key_type IN ('product','prize','gift','currency','faq','news')),
    key TEXT NOT NULL,
    name JSONB NOT NULL,
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (key_type, key)
);
CREATE INDEX IF NOT EXISTS base_category_list_idx ON base_category (key_type, sort, key);

CREATE TABLE IF NOT EXISTS base_tag (
    id UUID PRIMARY KEY,
    key_type TEXT NOT NULL CHECK (key_type IN ('product','prize','news','faq','general')),
    key TEXT NOT NULL,
    name JSONB NOT NULL,
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (key_type, key)
);
CREATE INDEX IF NOT EXISTS base_tag_list_idx ON base_tag (key_type, sort, key);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds languages, categories and tags used by other modules and cannot be reverted automatically; restore from backup', '20261101000000-base-schema'; END $$;
-- +migrate StatementEnd
