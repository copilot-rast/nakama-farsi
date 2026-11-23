-- +migrate Up
CREATE TABLE IF NOT EXISTS base_config_transfer_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_config_transfer_run (
    id UUID PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('export', 'import')),
    run_by TEXT NOT NULL DEFAULT '',
    source JSONB NOT NULL DEFAULT '{}',
    areas TEXT[] NOT NULL DEFAULT '{}',
    totals JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS base_config_transfer_run_created_idx ON base_config_transfer_run (created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS base_config_transfer_run;
DROP TABLE IF EXISTS base_config_transfer_settings;
