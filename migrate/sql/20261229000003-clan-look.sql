-- +migrate Up
ALTER TABLE clan ADD COLUMN IF NOT EXISTS motto TEXT NOT NULL DEFAULT '';
ALTER TABLE clan ADD COLUMN IF NOT EXISTS color TEXT NOT NULL DEFAULT '';
ALTER TABLE clan ADD COLUMN IF NOT EXISTS country TEXT NOT NULL DEFAULT '';
ALTER TABLE clan ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE clan ADD COLUMN IF NOT EXISTS emblem_key TEXT NOT NULL DEFAULT '';
ALTER TABLE clan ADD COLUMN IF NOT EXISTS cover_key TEXT NOT NULL DEFAULT '';
ALTER TABLE clan ADD COLUMN IF NOT EXISTS cover_url TEXT NOT NULL DEFAULT '';
ALTER TABLE clan ADD COLUMN IF NOT EXISTS look_changed_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS clan_art (
    id UUID PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('emblem','cover')),
    key TEXT NOT NULL,
    name JSONB NOT NULL DEFAULT '{}',
    image_url TEXT NOT NULL,
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    min_level INT NOT NULL DEFAULT 1 CHECK (min_level >= 1),
    price_currency TEXT NOT NULL DEFAULT '',
    price_amount BIGINT NOT NULL DEFAULT 0 CHECK (price_amount >= 0),
    is_default BOOLEAN NOT NULL DEFAULT false,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, key)
);

CREATE TABLE IF NOT EXISTS clan_art_owned (
    group_id UUID NOT NULL REFERENCES clan (group_id) ON DELETE CASCADE,
    art_id UUID NOT NULL REFERENCES clan_art (id) ON DELETE CASCADE,
    by_user UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, art_id)
);

CREATE TABLE IF NOT EXISTS clan_upload (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES clan (group_id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('emblem','cover')),
    file_name TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    size INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','replaced')),
    uploaded_by UUID,
    reviewed_by TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS clan_upload_status_idx ON clan_upload (status, created_at);
CREATE INDEX IF NOT EXISTS clan_upload_group_idx ON clan_upload (group_id, created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS clan_upload;
DROP TABLE IF EXISTS clan_art_owned;
DROP TABLE IF EXISTS clan_art;
ALTER TABLE clan DROP COLUMN IF EXISTS look_changed_at;
ALTER TABLE clan DROP COLUMN IF EXISTS cover_url;
ALTER TABLE clan DROP COLUMN IF EXISTS cover_key;
ALTER TABLE clan DROP COLUMN IF EXISTS emblem_key;
ALTER TABLE clan DROP COLUMN IF EXISTS tags;
ALTER TABLE clan DROP COLUMN IF EXISTS country;
ALTER TABLE clan DROP COLUMN IF EXISTS color;
ALTER TABLE clan DROP COLUMN IF EXISTS motto;
