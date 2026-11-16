-- +migrate Up
CREATE TABLE IF NOT EXISTS content_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS content_blob (
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

CREATE TABLE IF NOT EXISTS content_pack (
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

CREATE INDEX IF NOT EXISTS content_pack_active_idx ON content_pack (platform, active, activated_at DESC);

CREATE TABLE IF NOT EXISTS content_pack_file (
    pack_id UUID NOT NULL REFERENCES content_pack (id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    size BIGINT NOT NULL CHECK (size >= 0),
    PRIMARY KEY (pack_id, path)
);

CREATE INDEX IF NOT EXISTS content_pack_file_sha_idx ON content_pack_file (sha256);

CREATE TABLE IF NOT EXISTS content_pack_event (
    id BIGSERIAL PRIMARY KEY,
    pack_id UUID NOT NULL,
    platform TEXT NOT NULL,
    version TEXT NOT NULL,
    action TEXT NOT NULL,
    actor TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS content_pack_event_time_idx ON content_pack_event (created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS content_pack_event;
DROP TABLE IF EXISTS content_pack_file;
DROP TABLE IF EXISTS content_pack;
DROP TABLE IF EXISTS content_blob;
DROP TABLE IF EXISTS content_settings;
