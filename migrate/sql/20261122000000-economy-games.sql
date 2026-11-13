-- +migrate Up notransaction
ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak','referral','task','achievement','chance','game'));
ALTER TABLE base_category DROP CONSTRAINT IF EXISTS base_category_key_type_check;
ALTER TABLE base_category ADD CONSTRAINT base_category_key_type_check CHECK (key_type IN ('product','prize','gift','currency','faq','news','task','item','item_type','achievement','game'));
ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','gift','currency','news','faq','general','user','task','item','achievement','banner','sponsor','game'));

CREATE TABLE IF NOT EXISTS economy_game (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    maintenance JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    category_id UUID REFERENCES base_category (id),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    min_players INT NOT NULL DEFAULT 1 CHECK (min_players BETWEEN 1 AND 100),
    max_players INT NOT NULL DEFAULT 1 CHECK (max_players BETWEEN 1 AND 100),
    turn_seconds INT CHECK (turn_seconds IS NULL OR turn_seconds BETWEEN 1 AND 86400),
    match_seconds INT CHECK (match_seconds IS NULL OR match_seconds BETWEEN 1 AND 604800),
    bots_allowed BOOLEAN NOT NULL DEFAULT false,
    rooms JSONB NOT NULL DEFAULT '[]',
    leaderboard_id TEXT,
    platforms TEXT[] NOT NULL DEFAULT '{}',
    min_version TEXT,
    max_version TEXT,
    settings JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (max_players >= min_players)
);
CREATE INDEX IF NOT EXISTS economy_game_category_idx ON economy_game (category_id) WHERE category_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS economy_game_version (
    id INT PRIMARY KEY CHECK (id = 1),
    version BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO economy_game_version (id, version) VALUES (1, 1) ON CONFLICT DO NOTHING;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION economy_game_bump() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    UPDATE economy_game_version SET version = version + 1, updated_at = now() WHERE id = 1;
    RETURN NULL;
END $$;
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS economy_game_bump_trg ON economy_game;
CREATE TRIGGER economy_game_bump_trg AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE ON economy_game FOR EACH STATEMENT EXECUTE FUNCTION economy_game_bump();

CREATE TABLE IF NOT EXISTS economy_game_match (
    id TEXT PRIMARY KEY,
    game_key TEXT NOT NULL,
    room_key TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','started','finished','cancelled')),
    players INT NOT NULL DEFAULT 0,
    winners TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS economy_game_match_game_idx ON economy_game_match (game_key, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_game_match_open_idx ON economy_game_match (created_at) WHERE status IN ('open','started');

CREATE TABLE IF NOT EXISTS economy_game_entry (
    match_id TEXT NOT NULL REFERENCES economy_game_match (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','free','refunded','left')),
    fee JSONB,
    attempt INT NOT NULL DEFAULT 0,
    prize_id UUID,
    prize_status TEXT NOT NULL DEFAULT 'none' CHECK (prize_status IN ('none','pending','paid')),
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, user_id)
);
CREATE INDEX IF NOT EXISTS economy_game_entry_user_idx ON economy_game_entry (user_id, created_at DESC);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds game settings, matches and entry fees and cannot be reverted automatically; restore from backup', '20261122000000-economy-games'; END $$;
-- +migrate StatementEnd
