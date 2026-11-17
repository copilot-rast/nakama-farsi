-- +migrate Up
CREATE TABLE IF NOT EXISTS trophy_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO trophy_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS trophy_league (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '',
    min_trophies INT NOT NULL DEFAULT 0 CHECK (min_trophies >= 0),
    floor BOOLEAN NOT NULL DEFAULT false,
    reset_to INT CHECK (reset_to IS NULL OR reset_to >= 0),
    prize_key TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS trophy_league_min_idx ON trophy_league (min_trophies) WHERE is_active;

CREATE TABLE IF NOT EXISTS trophy_player (
    user_id UUID PRIMARY KEY,
    trophies INT NOT NULL DEFAULT 0 CHECK (trophies >= 0),
    best INT NOT NULL DEFAULT 0,
    season_best INT NOT NULL DEFAULT 0,
    league_key TEXT NOT NULL DEFAULT '',
    wins INT NOT NULL DEFAULT 0,
    losses INT NOT NULL DEFAULT 0,
    draws INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS trophy_player_rank_idx ON trophy_player (trophies DESC, user_id);

CREATE TABLE IF NOT EXISTS trophy_log (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    source_id TEXT NOT NULL,
    reason TEXT NOT NULL CHECK (reason IN ('win','loss','draw','admin','season')),
    delta INT NOT NULL,
    before INT NOT NULL,
    after INT NOT NULL,
    match_id TEXT NOT NULL DEFAULT '',
    opponent_id TEXT NOT NULL DEFAULT '',
    opponent_trophies INT,
    league_before TEXT NOT NULL DEFAULT '',
    league_after TEXT NOT NULL DEFAULT '',
    actor TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, source_id)
);
CREATE INDEX IF NOT EXISTS trophy_log_user_idx ON trophy_log (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS trophy_log_created_idx ON trophy_log (created_at);

CREATE TABLE IF NOT EXISTS trophy_season (
    number INT PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    players_reset INT NOT NULL DEFAULT 0,
    mode TEXT NOT NULL DEFAULT '',
    actor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS trophy_prize (
    user_id UUID NOT NULL,
    claim_key TEXT NOT NULL,
    league_key TEXT NOT NULL,
    season INT NOT NULL DEFAULT 0,
    prize_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','failed')),
    tries INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    paid JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, claim_key)
);
CREATE INDEX IF NOT EXISTS trophy_prize_pending_idx ON trophy_prize (created_at) WHERE status = 'pending';

-- +migrate StatementBegin
DO $$
DECLARE
    def TEXT;
    vals TEXT[];
BEGIN
    SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
        WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
    IF def IS NULL OR def ~ ('[{,'']trophy[},'']') THEN
        RETURN;
    END IF;
    SELECT array_agg(DISTINCT m[1]) INTO vals FROM regexp_matches(def, '''([a-z_]+)''', 'g') AS m;
    IF vals IS NULL THEN
        vals := string_to_array(substring(def from '''\{([^}]*)\}'''), ',');
    END IF;
    vals := array_append(vals, 'trophy');
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type = ANY (' || quote_literal(vals::text) || '::text[]))';
END $$;
-- +migrate StatementEnd

-- +migrate Down
DROP TABLE IF EXISTS trophy_prize;
DROP TABLE IF EXISTS trophy_season;
DROP TABLE IF EXISTS trophy_log;
DROP TABLE IF EXISTS trophy_player;
DROP TABLE IF EXISTS trophy_league;
DROP TABLE IF EXISTS trophy_settings;
