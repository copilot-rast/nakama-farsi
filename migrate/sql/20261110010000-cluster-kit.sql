-- +migrate Up
CREATE TABLE IF NOT EXISTS cluster_lease (
    name TEXT PRIMARY KEY,
    holder TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cluster_rate_limit (
    key TEXT PRIMARY KEY,
    window_start TIMESTAMPTZ NOT NULL,
    count BIGINT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS cluster_rate_limit_expires_idx ON cluster_rate_limit (expires_at);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Migration % is required by running plugins and cannot be reverted automatically', '20261110010000-cluster-kit'; END $$;
-- +migrate StatementEnd
