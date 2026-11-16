-- +migrate Up
CREATE TABLE IF NOT EXISTS insights_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS insights_session (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    client_id TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'client',
    day DATE NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    seconds INT,
    pings INT NOT NULL DEFAULT 0,
    app_version TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    device TEXT NOT NULL DEFAULT '',
    os TEXT NOT NULL DEFAULT '',
    UNIQUE (user_id, client_id)
);

CREATE INDEX IF NOT EXISTS insights_session_day_idx ON insights_session (day);
CREATE INDEX IF NOT EXISTS insights_session_user_idx ON insights_session (user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS insights_session_open_idx ON insights_session (last_seen_at) WHERE ended_at IS NULL;

CREATE TABLE IF NOT EXISTS insights_player (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    first_day DATE NOT NULL,
    first_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_day DATE NOT NULL,
    app_version TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS insights_player_first_idx ON insights_player (first_day);

CREATE TABLE IF NOT EXISTS insights_player_day (
    day DATE NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    sessions INT NOT NULL DEFAULT 0,
    seconds BIGINT NOT NULL DEFAULT 0,
    app_version TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (day, user_id)
);

CREATE INDEX IF NOT EXISTS insights_player_day_user_idx ON insights_player_day (user_id, day);

CREATE TABLE IF NOT EXISTS insights_heatmap_map (
    key TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    cell_size DOUBLE PRECISION NOT NULL DEFAULT 0,
    background_url TEXT NOT NULL DEFAULT '',
    image_bounds JSONB NOT NULL DEFAULT '{}',
    min_x DOUBLE PRECISION,
    max_x DOUBLE PRECISION,
    min_z DOUBLE PRECISION,
    max_z DOUBLE PRECISION,
    enabled BOOLEAN NOT NULL DEFAULT true,
    uploads BIGINT NOT NULL DEFAULT 0,
    points BIGINT NOT NULL DEFAULT 0,
    last_upload_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS insights_heatmap_cell (
    map_key TEXT NOT NULL REFERENCES insights_heatmap_map (key) ON DELETE CASCADE,
    size DOUBLE PRECISION NOT NULL,
    day DATE NOT NULL,
    cx INT NOT NULL,
    cz INT NOT NULL,
    points BIGINT NOT NULL DEFAULT 0,
    speed_sum DOUBLE PRECISION NOT NULL DEFAULT 0,
    speed_max DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (map_key, size, day, cx, cz)
);

CREATE TABLE IF NOT EXISTS insights_heatmap_upload (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    map_key TEXT NOT NULL REFERENCES insights_heatmap_map (key) ON DELETE CASCADE,
    upload_key TEXT NOT NULL,
    part INT NOT NULL DEFAULT 0,
    points INT NOT NULL DEFAULT 0,
    dropped INT NOT NULL DEFAULT 0,
    players INT NOT NULL DEFAULT 0,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    raw_key TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, upload_key, part)
);

CREATE INDEX IF NOT EXISTS insights_heatmap_upload_map_idx ON insights_heatmap_upload (map_key, created_at DESC);
CREATE INDEX IF NOT EXISTS insights_heatmap_upload_user_idx ON insights_heatmap_upload (user_id, created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS insights_heatmap_upload;
DROP TABLE IF EXISTS insights_heatmap_cell;
DROP TABLE IF EXISTS insights_heatmap_map;
DROP TABLE IF EXISTS insights_player_day;
DROP TABLE IF EXISTS insights_player;
DROP TABLE IF EXISTS insights_session;
DROP TABLE IF EXISTS insights_settings;
