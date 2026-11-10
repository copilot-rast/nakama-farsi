-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_brand (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    name JSONB NOT NULL DEFAULT '{}',
    short_name TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    logo_dark_url TEXT NOT NULL DEFAULT '',
    mark_url TEXT NOT NULL DEFAULT '',
    accent TEXT NOT NULL DEFAULT '' CHECK (accent = '' OR accent ~ '^#[0-9a-f]{6}$'),
    welcome JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO base_brand (id) VALUES (1) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS base_setup (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','completed','skipped')),
    steps JSONB NOT NULL DEFAULT '{}',
    first_opened_by TEXT NOT NULL DEFAULT '',
    first_opened_at TIMESTAMPTZ,
    completed_by TEXT NOT NULL DEFAULT '',
    completed_at TIMESTAMPTZ,
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO base_setup (id) VALUES (1) ON CONFLICT DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the console brand and setup state and cannot be reverted automatically; restore from backup', '20261110000000-base-brand-setup'; END $$;
-- +migrate StatementEnd
