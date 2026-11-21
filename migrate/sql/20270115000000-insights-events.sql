-- +migrate Up
CREATE TABLE IF NOT EXISTS insights_event (
    id UUID PRIMARY KEY,
    game TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    value DOUBLE PRECISION,
    tags JSONB NOT NULL DEFAULT '{}',
    fields JSONB NOT NULL DEFAULT '{}',
    user_id TEXT NOT NULL DEFAULT '',
    session_id TEXT NOT NULL DEFAULT '',
    at TIMESTAMPTZ NOT NULL,
    day DATE NOT NULL,
    server_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    platform TEXT NOT NULL DEFAULT '',
    app_version TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'client',
    flagged BOOLEAN NOT NULL DEFAULT false,
    flag_reason TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS insights_event_name_day_idx ON insights_event (name, day);

CREATE INDEX IF NOT EXISTS insights_event_user_at_idx ON insights_event (user_id, at);

CREATE INDEX IF NOT EXISTS insights_event_day_idx ON insights_event (day);

CREATE INDEX IF NOT EXISTS insights_event_session_idx ON insights_event (session_id, at) WHERE session_id <> '';

CREATE INDEX IF NOT EXISTS insights_event_tags_idx ON insights_event USING GIN (tags jsonb_path_ops);

CREATE TABLE IF NOT EXISTS insights_event_user (
    game TEXT NOT NULL DEFAULT '',
    user_id TEXT NOT NULL,
    first_day DATE NOT NULL,
    first_at TIMESTAMPTZ NOT NULL,
    last_day DATE NOT NULL,
    platform TEXT NOT NULL DEFAULT '',
    app_version TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (game, user_id)
);

CREATE INDEX IF NOT EXISTS insights_event_user_first_idx ON insights_event_user (first_day);

CREATE TABLE IF NOT EXISTS insights_event_user_day (
    game TEXT NOT NULL DEFAULT '',
    day DATE NOT NULL,
    user_id TEXT NOT NULL,
    events BIGINT NOT NULL DEFAULT 0,
    platform TEXT NOT NULL DEFAULT '',
    app_version TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (game, day, user_id)
);

CREATE INDEX IF NOT EXISTS insights_event_user_day_user_idx ON insights_event_user_day (user_id, day);

CREATE INDEX IF NOT EXISTS insights_event_user_day_day_idx ON insights_event_user_day (day);
CREATE TABLE IF NOT EXISTS insights_batch (
    user_id TEXT NOT NULL,
    batch_id TEXT NOT NULL,
    events INT NOT NULL DEFAULT 0,
    result JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, batch_id)
);

CREATE INDEX IF NOT EXISTS insights_batch_created_idx ON insights_batch (created_at);

CREATE TABLE IF NOT EXISTS insights_outbox (
    id BIGSERIAL PRIMARY KEY,
    source TEXT NOT NULL DEFAULT 'client',
    events JSONB NOT NULL,
    count INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    next_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempts INT NOT NULL DEFAULT 0,
    claimed_until TIMESTAMPTZ,
    claim_token TEXT NOT NULL DEFAULT '',
    dead BOOLEAN NOT NULL DEFAULT false,
    last_error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS insights_outbox_due_idx ON insights_outbox (next_at) WHERE NOT dead;

CREATE TABLE IF NOT EXISTS insights_storage (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    driver TEXT NOT NULL DEFAULT 'nakama',
    config JSONB NOT NULL DEFAULT '{}',
    secret TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS insights_tag_group (
    game TEXT NOT NULL DEFAULT '',
    key TEXT NOT NULL,
    name_fa TEXT NOT NULL DEFAULT '',
    name_en TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    closed BOOLEAN NOT NULL DEFAULT false,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (game, key)
);

CREATE TABLE IF NOT EXISTS insights_tag_value (
    game TEXT NOT NULL DEFAULT '',
    group_key TEXT NOT NULL,
    id TEXT NOT NULL,
    label_fa TEXT NOT NULL DEFAULT '',
    label_en TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (game, group_key, id),
    FOREIGN KEY (game, group_key) REFERENCES insights_tag_group (game, key) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS insights_event_def (
    game TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    label_fa TEXT NOT NULL DEFAULT '',
    label_en TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    path_groups JSONB NOT NULL DEFAULT '[]',
    required_tags JSONB NOT NULL DEFAULT '[]',
    allowed_tags JSONB NOT NULL DEFAULT '[]',
    value_rule JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (game, name)
);

CREATE TABLE IF NOT EXISTS insights_funnel (
    id UUID PRIMARY KEY,
    game TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    steps JSONB NOT NULL DEFAULT '[]',
    window_seconds INT NOT NULL DEFAULT 86400,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS insights_connector (
    id UUID PRIMARY KEY,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    config JSONB NOT NULL DEFAULT '{}',
    secret TEXT NOT NULL DEFAULT '',
    every_minutes INT NOT NULL DEFAULT 60,
    cursor JSONB NOT NULL DEFAULT '{}',
    next_run_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    claimed_until TIMESTAMPTZ,
    claimed_by TEXT NOT NULL DEFAULT '',
    last_run_at TIMESTAMPTZ,
    last_status TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    last_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS insights_import_run (
    id UUID PRIMARY KEY,
    connector_id UUID NOT NULL REFERENCES insights_connector (id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'running',
    read BIGINT NOT NULL DEFAULT 0,
    written BIGINT NOT NULL DEFAULT 0,
    skipped BIGINT NOT NULL DEFAULT 0,
    error TEXT NOT NULL DEFAULT '',
    cursor_from JSONB NOT NULL DEFAULT '{}',
    cursor_to JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS insights_import_run_connector_idx ON insights_import_run (connector_id, started_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS insights_import_run;
DROP TABLE IF EXISTS insights_connector;
DROP TABLE IF EXISTS insights_funnel;
DROP TABLE IF EXISTS insights_event_def;
DROP TABLE IF EXISTS insights_tag_value;
DROP TABLE IF EXISTS insights_tag_group;
DROP TABLE IF EXISTS insights_storage;
DROP TABLE IF EXISTS insights_outbox;
DROP TABLE IF EXISTS insights_batch;
DROP TABLE IF EXISTS insights_event_user_day;
DROP TABLE IF EXISTS insights_event_user;
DROP TABLE IF EXISTS insights_event;
