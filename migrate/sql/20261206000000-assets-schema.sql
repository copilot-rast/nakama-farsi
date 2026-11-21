-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS assets_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO assets_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- Versioned data entries (ScriptableObject or any typed JSON) pushed from an editor tool.
-- data is json (not jsonb) so the sent key order is kept; hash is SHA-256 of the canonical form.
CREATE TABLE IF NOT EXISTS assets_entry (
    asset_id TEXT PRIMARY KEY,
    source_guid TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    folder TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    data JSON NOT NULL,
    hash TEXT NOT NULL,
    size_bytes INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1 CHECK (version >= 1),
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS assets_entry_folder_idx ON assets_entry (folder, asset_id);
CREATE INDEX IF NOT EXISTS assets_entry_type_idx ON assets_entry (type);
CREATE INDEX IF NOT EXISTS assets_entry_source_idx ON assets_entry (source_guid) WHERE source_guid <> '';

-- Every state an entry had: created, updated, restored (rollback) and deleted (the last state).
CREATE TABLE IF NOT EXISTS assets_entry_history (
    id BIGSERIAL PRIMARY KEY,
    asset_id TEXT NOT NULL,
    version INT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('created', 'updated', 'restored', 'deleted')),
    source_guid TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    folder TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    data JSON NOT NULL,
    hash TEXT NOT NULL,
    size_bytes INT NOT NULL DEFAULT 0,
    restored_from BIGINT,
    actor TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS assets_entry_history_asset_idx ON assets_entry_history (asset_id, id DESC);
CREATE INDEX IF NOT EXISTS assets_entry_history_time_idx ON assets_entry_history (created_at);

-- Content-addressed files of content packs (Addressables bundles, catalogs). A file arrives
-- in parts (each below the server's request size limit) and is joined once all parts are in.
CREATE TABLE IF NOT EXISTS assets_blob (
    sha256 TEXT PRIMARY KEY,
    size BIGINT NOT NULL CHECK (size >= 0),
    part_bytes INT NOT NULL CHECK (part_bytes > 0),
    parts INT NOT NULL CHECK (parts >= 1),
    received INT[] NOT NULL DEFAULT '{}',
    stored BOOLEAN NOT NULL DEFAULT false,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    stored_at TIMESTAMPTZ
);

-- One upload of a remote content build for one platform; the game reads the latest active pack
-- whose app version range holds its own version.
CREATE TABLE IF NOT EXISTS assets_pack (
    id UUID PRIMARY KEY,
    platform TEXT NOT NULL,
    version TEXT NOT NULL,
    min_app_version TEXT NOT NULL DEFAULT '',
    max_app_version TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'uploading' CHECK (status IN ('uploading', 'ready')),
    active BOOLEAN NOT NULL DEFAULT false,
    catalog_path TEXT NOT NULL DEFAULT '',
    hash_path TEXT NOT NULL DEFAULT '',
    catalog_hash TEXT NOT NULL DEFAULT '',
    file_count INT NOT NULL DEFAULT 0,
    total_bytes BIGINT NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    activated_at TIMESTAMPTZ,
    activated_by TEXT NOT NULL DEFAULT '',
    UNIQUE (platform, version)
);

CREATE INDEX IF NOT EXISTS assets_pack_active_idx ON assets_pack (platform, active, activated_at DESC);

CREATE TABLE IF NOT EXISTS assets_pack_file (
    pack_id UUID NOT NULL REFERENCES assets_pack (id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    size BIGINT NOT NULL CHECK (size >= 0),
    PRIMARY KEY (pack_id, path)
);

CREATE INDEX IF NOT EXISTS assets_pack_file_sha_idx ON assets_pack_file (sha256);

-- Who uploaded, activated, deactivated or removed a pack.
CREATE TABLE IF NOT EXISTS assets_pack_event (
    id BIGSERIAL PRIMARY KEY,
    pack_id UUID NOT NULL,
    platform TEXT NOT NULL,
    version TEXT NOT NULL,
    action TEXT NOT NULL,
    actor TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS assets_pack_event_time_idx ON assets_pack_event (created_at DESC);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Assets migration % holds game data and content packs and cannot be reverted automatically; restore from backup', '20261206000000-assets-schema'; END $$;
-- +migrate StatementEnd
