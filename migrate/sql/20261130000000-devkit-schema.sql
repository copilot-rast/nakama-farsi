-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS devkit_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO devkit_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS devkit_error_message (
    code TEXT PRIMARY KEY CHECK (code ~ '^[A-Z][A-Z0-9_]{1,79}$'),
    messages JSONB NOT NULL DEFAULT '{}',
    status INT CHECK (status IS NULL OR status BETWEEN 100 AND 599),
    retry_after_seconds INT CHECK (retry_after_seconds IS NULL OR retry_after_seconds BETWEEN 0 AND 86400),
    is_custom BOOLEAN NOT NULL DEFAULT false,
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS devkit_api_scope (
    name TEXT PRIMARY KEY CHECK (name ~ '^[a-z][a-z0-9-]{0,39}:[a-z][a-z0-9-]{0,39}$'),
    description TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'console',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS devkit_api_key (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    description TEXT NOT NULL DEFAULT '',
    lookup TEXT NOT NULL UNIQUE,
    secret_hash BYTEA NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    allowed_ips TEXT[] NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_by TEXT NOT NULL DEFAULT '',
    last_used_at TIMESTAMPTZ,
    last_used_ip TEXT NOT NULL DEFAULT '',
    use_count BIGINT NOT NULL DEFAULT 0,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS devkit_api_key_created_idx ON devkit_api_key (created_at DESC);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Devkit migration % holds API keys, error messages and devkit settings and cannot be reverted automatically; restore from backup', '20261130000000-devkit-schema'; END $$;
-- +migrate StatementEnd
