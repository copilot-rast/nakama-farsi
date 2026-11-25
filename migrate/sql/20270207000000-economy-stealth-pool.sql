-- +migrate Up
CREATE TABLE IF NOT EXISTS economy_bot_ghost (
    user_id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    avatar_url TEXT NOT NULL DEFAULT '',
    level INT NOT NULL DEFAULT 1,
    look JSONB NOT NULL DEFAULT '{}',
    game_key TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS economy_bot_ghost_username_idx ON economy_bot_ghost (username);
CREATE INDEX IF NOT EXISTS economy_bot_ghost_created_idx ON economy_bot_ghost (created_at);

CREATE TABLE IF NOT EXISTS economy_game_claim (
    user_id UUID NOT NULL,
    game_key TEXT NOT NULL,
    match_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, game_key)
);

ALTER TABLE economy_game_match ADD COLUMN IF NOT EXISTS stats_done BOOLEAN NOT NULL DEFAULT false;

-- +migrate Down
ALTER TABLE economy_game_match DROP COLUMN IF EXISTS stats_done;
DROP TABLE IF EXISTS economy_game_claim;
DROP TABLE IF EXISTS economy_bot_ghost;
