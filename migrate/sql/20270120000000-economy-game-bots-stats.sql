-- +migrate Up
ALTER TABLE economy_game ADD COLUMN IF NOT EXISTS one_match BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE economy_game_match ADD COLUMN IF NOT EXISTS end_reason TEXT;
ALTER TABLE economy_game_match ADD COLUMN IF NOT EXISTS forced_by TEXT;
ALTER TABLE economy_game_match ADD COLUMN IF NOT EXISTS note TEXT;
ALTER TABLE economy_game_match ADD COLUMN IF NOT EXISTS results JSONB;
ALTER TABLE economy_game_match ADD COLUMN IF NOT EXISTS summary JSONB;
ALTER TABLE economy_game_match ADD COLUMN IF NOT EXISTS actions INT NOT NULL DEFAULT 0;

ALTER TABLE economy_game_entry ADD COLUMN IF NOT EXISTS bot BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE economy_game_entry ADD COLUMN IF NOT EXISTS released BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE economy_game_entry ADD COLUMN IF NOT EXISTS result TEXT;
ALTER TABLE economy_game_entry ADD COLUMN IF NOT EXISTS rank INT;
ALTER TABLE economy_game_entry ADD COLUMN IF NOT EXISTS amount JSONB;
CREATE INDEX IF NOT EXISTS economy_game_entry_active_idx ON economy_game_entry (user_id) WHERE status IN ('paid','free') AND NOT released AND NOT bot;

CREATE TABLE IF NOT EXISTS economy_game_action (
    match_id TEXT NOT NULL,
    seq INT NOT NULL,
    user_id UUID,
    kind TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, seq)
);
CREATE INDEX IF NOT EXISTS economy_game_action_at_idx ON economy_game_action (at);

CREATE TABLE IF NOT EXISTS economy_game_stats (
    user_id UUID NOT NULL,
    game_key TEXT NOT NULL,
    games INT NOT NULL DEFAULT 0,
    wins INT NOT NULL DEFAULT 0,
    losses INT NOT NULL DEFAULT 0,
    forfeits INT NOT NULL DEFAULT 0,
    timeouts INT NOT NULL DEFAULT 0,
    playtime_seconds BIGINT NOT NULL DEFAULT 0,
    winnings JSONB NOT NULL DEFAULT '{}',
    last_played_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by TEXT,
    PRIMARY KEY (user_id, game_key)
);
CREATE INDEX IF NOT EXISTS economy_game_stats_game_idx ON economy_game_stats (game_key, wins DESC);

CREATE TABLE IF NOT EXISTS economy_bot (
    user_id UUID PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT true,
    level INT NOT NULL DEFAULT 1 CHECK (level BETWEEN 1 AND 1000),
    template TEXT NOT NULL DEFAULT '',
    rooms JSONB NOT NULL DEFAULT '[]',
    look JSONB NOT NULL DEFAULT '{}',
    note TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'console',
    games INT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by TEXT
);
CREATE INDEX IF NOT EXISTS economy_bot_pick_idx ON economy_bot (enabled, level);

-- +migrate Down
DROP TABLE IF EXISTS economy_bot;
DROP TABLE IF EXISTS economy_game_stats;
DROP TABLE IF EXISTS economy_game_action;
DROP INDEX IF EXISTS economy_game_entry_active_idx;
ALTER TABLE economy_game_entry DROP COLUMN IF EXISTS amount, DROP COLUMN IF EXISTS rank, DROP COLUMN IF EXISTS result, DROP COLUMN IF EXISTS released, DROP COLUMN IF EXISTS bot;
ALTER TABLE economy_game_match DROP COLUMN IF EXISTS actions, DROP COLUMN IF EXISTS summary, DROP COLUMN IF EXISTS results, DROP COLUMN IF EXISTS note, DROP COLUMN IF EXISTS forced_by, DROP COLUMN IF EXISTS end_reason;
ALTER TABLE economy_game DROP COLUMN IF EXISTS one_match;
