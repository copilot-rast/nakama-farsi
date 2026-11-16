-- +migrate Up
CREATE TABLE IF NOT EXISTS base_server_rules (
    id INT PRIMARY KEY CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_account_deletion (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS base_account_deletion_due_idx ON base_account_deletion (due_at);

-- +migrate Down
DROP TABLE IF EXISTS base_account_deletion;
DROP TABLE IF EXISTS base_server_rules;
