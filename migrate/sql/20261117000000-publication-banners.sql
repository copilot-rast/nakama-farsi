-- +migrate Up notransaction
ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','gift','currency','news','faq','general','user','task','item','achievement','banner'));

CREATE TABLE IF NOT EXISTS publication_banner_list (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9_]{1,63}$'),
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    layout TEXT NOT NULL DEFAULT 'carousel' CHECK (layout IN ('carousel','single','grid','strip')),
    max_items INT NOT NULL DEFAULT 5 CHECK (max_items BETWEEN 1 AND 20),
    rotate_seconds INT NOT NULL DEFAULT 5 CHECK (rotate_seconds BETWEEN 0 AND 60),
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS publication_banner (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9_]{1,63}$'),
    list_id UUID NOT NULL REFERENCES publication_banner_list (id),
    name TEXT NOT NULL DEFAULT '',
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    title JSONB NOT NULL DEFAULT '{}',
    subtitle JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    image_url JSONB NOT NULL DEFAULT '{}',
    bg_color TEXT NOT NULL DEFAULT '' CHECK (bg_color = '' OR bg_color ~ '^#[0-9a-f]{6}$'),
    action_type TEXT NOT NULL DEFAULT 'none' CHECK (action_type IN ('none','screen','product','achievement','url')),
    action_value TEXT NOT NULL DEFAULT '',
    action_label JSONB NOT NULL DEFAULT '{}',
    action_url JSONB NOT NULL DEFAULT '{}',
    priority INT NOT NULL DEFAULT 0 CHECK (priority BETWEEN -1000 AND 1000),
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    days INT[] NOT NULL DEFAULT '{}',
    day_start INT CHECK (day_start BETWEEN 0 AND 1439),
    day_end INT CHECK (day_end BETWEEN 1 AND 1440),
    audience_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    exclude_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    level_table_id UUID REFERENCES economy_level_table (id),
    min_level INT CHECK (min_level BETWEEN 0 AND 100000),
    max_level INT CHECK (max_level BETWEEN 0 AND 100000),
    platforms TEXT[] NOT NULL DEFAULT '{}',
    min_version TEXT NOT NULL DEFAULT '',
    max_version TEXT NOT NULL DEFAULT '',
    countries TEXT[] NOT NULL DEFAULT '{}',
    languages TEXT[] NOT NULL DEFAULT '{}',
    player_age TEXT NOT NULL DEFAULT 'all' CHECK (player_age IN ('all','new','returning')),
    new_days INT NOT NULL DEFAULT 7 CHECK (new_days BETWEEN 1 AND 365),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
    CHECK (min_level IS NULL OR max_level IS NULL OR min_level <= max_level),
    CHECK ((day_start IS NULL) = (day_end IS NULL)),
    CHECK (day_start IS NULL OR day_start < day_end),
    CHECK (action_type <> 'none' OR action_value = ''),
    CHECK (action_type IN ('none','url') OR action_value <> '')
);
CREATE INDEX IF NOT EXISTS publication_banner_list_idx ON publication_banner (list_id, priority DESC, created_at) WHERE is_active;

CREATE TABLE IF NOT EXISTS publication_banner_day (
    banner_id UUID NOT NULL REFERENCES publication_banner (id) ON DELETE CASCADE,
    day DATE NOT NULL,
    views BIGINT NOT NULL DEFAULT 0,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (banner_id, day)
);

CREATE TABLE IF NOT EXISTS publication_banner_seen (
    banner_id UUID NOT NULL REFERENCES publication_banner (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    day DATE NOT NULL,
    viewed_at TIMESTAMPTZ,
    clicked_at TIMESTAMPTZ,
    PRIMARY KEY (banner_id, user_id, day)
);
CREATE INDEX IF NOT EXISTS publication_banner_seen_day_idx ON publication_banner_seen (day);
CREATE INDEX IF NOT EXISTS publication_banner_seen_user_idx ON publication_banner_seen (user_id);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % holds banners, placements and their counters and cannot be reverted automatically; restore from backup', '20261117000000-publication-banners'; END $$;
-- +migrate StatementEnd
