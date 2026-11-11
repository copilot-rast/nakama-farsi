-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_config_key (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK (type IN ('string','number','bool','json','text')),
    description TEXT NOT NULL DEFAULT '',
    value JSONB NOT NULL,
    rules JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    version INT NOT NULL DEFAULT 1,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_config_history (
    id UUID PRIMARY KEY,
    key_id UUID NOT NULL,
    key TEXT NOT NULL,
    version INT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create','update','delete','rollback')),
    snapshot JSONB NOT NULL,
    changed_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS base_config_history_key_idx ON base_config_history (key_id, version DESC);
CREATE INDEX IF NOT EXISTS base_config_history_created_idx ON base_config_history (created_at DESC);

CREATE TABLE IF NOT EXISTS base_config_seen (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    hash TEXT NOT NULL,
    variants JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_activity (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    action TEXT NOT NULL,
    ref_id TEXT NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}',
    actor TEXT NOT NULL DEFAULT '',
    ip TEXT NOT NULL DEFAULT '',
    country TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS base_activity_user_idx ON base_activity (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS base_activity_kind_idx ON base_activity (kind, created_at DESC);
CREATE INDEX IF NOT EXISTS base_activity_created_idx ON base_activity (created_at);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds remote config history and player activity and cannot be reverted automatically; restore from backup', '20261112000001-base-remote-config-activity'; END $$;
-- +migrate StatementEnd
