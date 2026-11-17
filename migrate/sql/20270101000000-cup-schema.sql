-- +migrate Up
CREATE TABLE IF NOT EXISTS cup_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO cup_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS cup_template (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT false,
    schedule TEXT NOT NULL DEFAULT '',
    signup_minutes INT NOT NULL DEFAULT 1440 CHECK (signup_minutes >= 0),
    announce_minutes INT NOT NULL DEFAULT 0 CHECK (announce_minutes >= 0),
    rules JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    last_starts_at TIMESTAMPTZ,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cup (
    id UUID PRIMARY KEY,
    template_id UUID REFERENCES cup_template (id) ON DELETE SET NULL,
    template_key TEXT NOT NULL DEFAULT '',
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','upcoming','signup','running','finished','cancelled')),
    announce_at TIMESTAMPTZ,
    signup_at TIMESTAMPTZ NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    rules JSONB NOT NULL DEFAULT '{}',
    size INT NOT NULL DEFAULT 0,
    rounds INT NOT NULL DEFAULT 0,
    current_round INT NOT NULL DEFAULT 0,
    players INT NOT NULL DEFAULT 0,
    pool JSONB NOT NULL DEFAULT '{}',
    champion UUID,
    cancel_reason TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS cup_template_start_idx ON cup (template_id, starts_at) WHERE template_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS cup_status_idx ON cup (status, starts_at);

CREATE TABLE IF NOT EXISTS cup_player (
    cup_id UUID NOT NULL REFERENCES cup (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'registered' CHECK (status IN ('registered','withdrawn','removed','playing','eliminated','disqualified','finished')),
    seed INT NOT NULL DEFAULT 0,
    trophies INT NOT NULL DEFAULT 0,
    place INT NOT NULL DEFAULT 0,
    reached INT NOT NULL DEFAULT 0,
    fee JSONB NOT NULL DEFAULT '{}',
    fee_state TEXT NOT NULL DEFAULT 'none' CHECK (fee_state IN ('none','paid','refunded')),
    added_by TEXT NOT NULL DEFAULT '',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (cup_id, user_id)
);
CREATE INDEX IF NOT EXISTS cup_player_user_idx ON cup_player (user_id, joined_at DESC);

CREATE TABLE IF NOT EXISTS cup_match (
    id UUID PRIMARY KEY,
    cup_id UUID NOT NULL REFERENCES cup (id) ON DELETE CASCADE,
    round INT NOT NULL,
    slot INT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'main' CHECK (kind IN ('main','third')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','waiting','open','playing','finished','bye','void')),
    best_of INT NOT NULL DEFAULT 1 CHECK (best_of >= 1),
    user_a UUID,
    user_b UUID,
    seed_a INT NOT NULL DEFAULT 0,
    seed_b INT NOT NULL DEFAULT 0,
    wins_a INT NOT NULL DEFAULT 0,
    wins_b INT NOT NULL DEFAULT 0,
    draws INT NOT NULL DEFAULT 0,
    games INT NOT NULL DEFAULT 0,
    ready_a TIMESTAMPTZ,
    ready_b TIMESTAMPTZ,
    showed_a BOOLEAN NOT NULL DEFAULT false,
    showed_b BOOLEAN NOT NULL DEFAULT false,
    opens_at TIMESTAMPTZ,
    starts_at TIMESTAMPTZ,
    closes_at TIMESTAMPTZ,
    winner UUID,
    loser UUID,
    reason TEXT NOT NULL DEFAULT '',
    game_match_id TEXT NOT NULL DEFAULT '',
    actor TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    UNIQUE (cup_id, kind, round, slot)
);
CREATE INDEX IF NOT EXISTS cup_match_status_idx ON cup_match (status, closes_at);
CREATE INDEX IF NOT EXISTS cup_match_user_a_idx ON cup_match (user_a) WHERE user_a IS NOT NULL;
CREATE INDEX IF NOT EXISTS cup_match_user_b_idx ON cup_match (user_b) WHERE user_b IS NOT NULL;

CREATE TABLE IF NOT EXISTS cup_game (
    id UUID PRIMARY KEY,
    match_id UUID NOT NULL REFERENCES cup_match (id) ON DELETE CASCADE,
    cup_id UUID NOT NULL,
    game_no INT NOT NULL,
    game_match_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'launching' CHECK (status IN ('launching','playing','finished','void')),
    winner UUID,
    draw BOOLEAN NOT NULL DEFAULT false,
    score JSONB NOT NULL DEFAULT '{}',
    present JSONB NOT NULL DEFAULT '[]',
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    UNIQUE (match_id, game_no)
);
CREATE UNIQUE INDEX IF NOT EXISTS cup_game_external_idx ON cup_game (game_match_id) WHERE game_match_id <> '';

CREATE TABLE IF NOT EXISTS cup_payout (
    id UUID PRIMARY KEY,
    cup_id UUID NOT NULL REFERENCES cup (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    kind TEXT NOT NULL,
    prize_key TEXT NOT NULL DEFAULT '',
    amounts JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','failed')),
    tries INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    paid JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at TIMESTAMPTZ,
    UNIQUE (cup_id, user_id, kind)
);
CREATE INDEX IF NOT EXISTS cup_payout_pending_idx ON cup_payout (created_at) WHERE status = 'pending';

-- +migrate StatementBegin
DO $$
DECLARE
    def TEXT;
    vals TEXT[];
BEGIN
    SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
        WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
    IF def IS NULL OR def ~ ('[{,'']cup[},'']') THEN
        RETURN;
    END IF;
    SELECT array_agg(DISTINCT m[1]) INTO vals FROM regexp_matches(def, '''([a-z_]+)''', 'g') AS m;
    IF vals IS NULL THEN
        vals := string_to_array(substring(def from '''\{([^}]*)\}'''), ',');
    END IF;
    vals := array_append(vals, 'cup');
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type = ANY (' || quote_literal(vals::text) || '::text[]))';
END $$;
-- +migrate StatementEnd

-- +migrate Down
DROP TABLE IF EXISTS cup_payout;
DROP TABLE IF EXISTS cup_game;
DROP TABLE IF EXISTS cup_match;
DROP TABLE IF EXISTS cup_player;
DROP TABLE IF EXISTS cup;
DROP TABLE IF EXISTS cup_template;
DROP TABLE IF EXISTS cup_settings;
