-- +migrate Up notransaction
ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','gift','currency','news','faq','general','user','task','item','achievement','banner','sponsor'));

CREATE TABLE IF NOT EXISTS publication_sponsor (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9_]{1,63}$'),
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    logo_url TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '' CHECK (color = '' OR color ~ '^#[0-9a-f]{6}$'),
    website_url TEXT NOT NULL DEFAULT '',
    links JSONB NOT NULL DEFAULT '[]',
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    contact_name TEXT NOT NULL DEFAULT '',
    contact_email TEXT NOT NULL DEFAULT '',
    contact_phone TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);

CREATE TABLE IF NOT EXISTS publication_sponsor_day (
    sponsor_id UUID NOT NULL REFERENCES publication_sponsor (id) ON DELETE CASCADE,
    day DATE NOT NULL,
    views BIGINT NOT NULL DEFAULT 0,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (sponsor_id, day)
);

CREATE TABLE IF NOT EXISTS publication_sponsor_seen (
    sponsor_id UUID NOT NULL REFERENCES publication_sponsor (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    day DATE NOT NULL,
    viewed_at TIMESTAMPTZ,
    clicked_at TIMESTAMPTZ,
    PRIMARY KEY (sponsor_id, user_id, day)
);
CREATE INDEX IF NOT EXISTS publication_sponsor_seen_day_idx ON publication_sponsor_seen (day);
CREATE INDEX IF NOT EXISTS publication_sponsor_seen_user_idx ON publication_sponsor_seen (user_id);

ALTER TABLE publication_banner ADD COLUMN IF NOT EXISTS sponsor_id UUID REFERENCES publication_sponsor (id);
ALTER TABLE publication_banner_list ADD COLUMN IF NOT EXISTS sponsor_id UUID REFERENCES publication_sponsor (id);
CREATE INDEX IF NOT EXISTS publication_banner_sponsor_idx ON publication_banner (sponsor_id) WHERE sponsor_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS publication_banner_list_sponsor_idx ON publication_banner_list (sponsor_id) WHERE sponsor_id IS NOT NULL;

ALTER TABLE publication_banner DROP CONSTRAINT IF EXISTS publication_banner_action_type_check;
ALTER TABLE publication_banner ADD CONSTRAINT publication_banner_action_type_check CHECK (action_type IN ('none','screen','product','achievement','url','sponsor'));
ALTER TABLE publication_banner DROP CONSTRAINT IF EXISTS publication_banner_check4;
ALTER TABLE publication_banner DROP CONSTRAINT IF EXISTS publication_banner_check5;
ALTER TABLE publication_banner DROP CONSTRAINT IF EXISTS publication_banner_action_value_check;
ALTER TABLE publication_banner ADD CONSTRAINT publication_banner_action_value_check CHECK ((action_type IN ('none','url','sponsor') AND action_value = '') OR (action_type NOT IN ('none','url','sponsor') AND action_value <> ''));
ALTER TABLE publication_banner DROP CONSTRAINT IF EXISTS publication_banner_sponsor_action_check;
ALTER TABLE publication_banner ADD CONSTRAINT publication_banner_sponsor_action_check CHECK (action_type <> 'sponsor' OR sponsor_id IS NOT NULL);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % holds sponsors and their counters and cannot be reverted automatically; restore from backup', '20261119000000-publication-sponsors'; END $$;
-- +migrate StatementEnd
