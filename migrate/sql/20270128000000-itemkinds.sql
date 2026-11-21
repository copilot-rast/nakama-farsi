-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS itemkind_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO itemkind_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS itemkind_profile (
    item_id UUID PRIMARY KEY REFERENCES economy_item (id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'ok' CHECK (status IN ('ok', 'pending', 'deleted')),
    reason TEXT NOT NULL DEFAULT '',
    schema_version INT NOT NULL DEFAULT 1,
    source_asset_id TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_hash TEXT NOT NULL DEFAULT '',
    source_version INT NOT NULL DEFAULT 0,
    linked_at TIMESTAMPTZ,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS itemkind_profile_source_idx ON itemkind_profile (source_asset_id) WHERE source_asset_id <> '';
CREATE INDEX IF NOT EXISTS itemkind_profile_kind_idx ON itemkind_profile (kind, status);

CREATE TABLE IF NOT EXISTS itemkind_route (
    id BIGSERIAL PRIMARY KEY,
    source_type TEXT NOT NULL,
    match JSONB NOT NULL DEFAULT '{}',
    kind TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL CHECK (role IN ('item', 'dependent', 'ignore')),
    priority INT NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT true,
    note TEXT NOT NULL DEFAULT '',
    origin TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS itemkind_route_origin_idx ON itemkind_route (origin) WHERE origin <> '';
CREATE INDEX IF NOT EXISTS itemkind_route_type_idx ON itemkind_route (source_type, priority, id);

CREATE TABLE IF NOT EXISTS itemkind_run (
    id BIGSERIAL PRIMARY KEY,
    source TEXT NOT NULL,
    dry_run BOOLEAN NOT NULL DEFAULT false,
    actor TEXT NOT NULL DEFAULT '',
    summary JSONB NOT NULL DEFAULT '{}',
    rows JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS itemkind_run_time_idx ON itemkind_run (created_at DESC);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Item kinds migration % holds item profiles linked to player-owned items and cannot be reverted automatically; restore from backup', '20270128000000-itemkinds'; END $$;
-- +migrate StatementEnd
