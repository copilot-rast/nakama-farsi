-- +migrate Up
CREATE TABLE IF NOT EXISTS clan_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO clan_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS clan (
    group_id UUID PRIMARY KEY,
    tag TEXT NOT NULL UNIQUE,
    badge JSONB NOT NULL DEFAULT '{}',
    join_type TEXT NOT NULL DEFAULT 'open' CHECK (join_type IN ('open','closed','invite')),
    min_trophies INT NOT NULL DEFAULT 0 CHECK (min_trophies >= 0),
    level INT NOT NULL DEFAULT 1,
    xp BIGINT NOT NULL DEFAULT 0 CHECK (xp >= 0),
    trophies BIGINT NOT NULL DEFAULT 0,
    war_wins INT NOT NULL DEFAULT 0,
    war_losses INT NOT NULL DEFAULT 0,
    war_draws INT NOT NULL DEFAULT 0,
    war_points BIGINT NOT NULL DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS clan_trophies_idx ON clan (trophies DESC);
CREATE INDEX IF NOT EXISTS clan_xp_idx ON clan (xp DESC);

CREATE TABLE IF NOT EXISTS clan_member (
    group_id UUID NOT NULL REFERENCES clan (group_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('leader','co_leader','elder','member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    xp BIGINT NOT NULL DEFAULT 0,
    xp_day DATE,
    xp_today BIGINT NOT NULL DEFAULT 0,
    war_points BIGINT NOT NULL DEFAULT 0,
    wars INT NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, user_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS clan_member_user_idx ON clan_member (user_id);

CREATE TABLE IF NOT EXISTS clan_request (
    group_id UUID NOT NULL REFERENCES clan (group_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('request','invite')),
    message TEXT NOT NULL DEFAULT '',
    by_user UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    PRIMARY KEY (group_id, user_id, kind)
);
CREATE INDEX IF NOT EXISTS clan_request_user_idx ON clan_request (user_id, kind);

CREATE TABLE IF NOT EXISTS clan_cooldown (
    user_id UUID NOT NULL,
    group_id UUID NOT NULL,
    reason TEXT NOT NULL CHECK (reason IN ('left','kicked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, group_id)
);
CREATE INDEX IF NOT EXISTS clan_cooldown_user_idx ON clan_cooldown (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS clan_level (
    level INT PRIMARY KEY CHECK (level >= 1),
    xp BIGINT NOT NULL DEFAULT 0 CHECK (xp >= 0),
    name JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    member_cap INT NOT NULL DEFAULT 30 CHECK (member_cap BETWEEN 1 AND 1000),
    prize_key TEXT NOT NULL DEFAULT '',
    perks JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clan_level_reached (
    group_id UUID NOT NULL REFERENCES clan (group_id) ON DELETE CASCADE,
    level INT NOT NULL,
    reached_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    members INT NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, level)
);

CREATE TABLE IF NOT EXISTS clan_xp_log (
    id BIGSERIAL PRIMARY KEY,
    group_id UUID NOT NULL,
    user_id UUID,
    source_id TEXT NOT NULL,
    source TEXT NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (group_id, source_id)
);
CREATE INDEX IF NOT EXISTS clan_xp_log_group_idx ON clan_xp_log (group_id, created_at DESC);

CREATE TABLE IF NOT EXISTS clan_payout (
    id UUID PRIMARY KEY,
    source_id TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL CHECK (kind IN ('level','war')),
    group_id UUID NOT NULL,
    user_id UUID NOT NULL,
    ref TEXT NOT NULL DEFAULT '',
    prize_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','failed')),
    tries INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    paid JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS clan_payout_pending_idx ON clan_payout (created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS clan_payout_user_idx ON clan_payout (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS clan_war_config (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    mode TEXT NOT NULL CHECK (mode IN ('points','duel')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    schedule TEXT NOT NULL DEFAULT '',
    duration_minutes INT NOT NULL DEFAULT 1440 CHECK (duration_minutes BETWEEN 1 AND 43200),
    signup TEXT NOT NULL DEFAULT 'auto' CHECK (signup IN ('auto','leader')),
    signup_minutes INT NOT NULL DEFAULT 0 CHECK (signup_minutes BETWEEN 0 AND 43200),
    pairing TEXT NOT NULL DEFAULT 'random' CHECK (pairing IN ('random','level','trophies')),
    rules JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clan_war_round (
    id UUID PRIMARY KEY,
    config_id UUID REFERENCES clan_war_config (id) ON DELETE SET NULL,
    config_key TEXT NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('points','duel')),
    status TEXT NOT NULL CHECK (status IN ('signup','active','settled','cancelled')),
    signup_at TIMESTAMPTZ,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    rules JSONB NOT NULL DEFAULT '{}',
    name JSONB NOT NULL DEFAULT '{}',
    pairing TEXT NOT NULL DEFAULT 'random',
    signup TEXT NOT NULL DEFAULT 'auto',
    wars INT NOT NULL DEFAULT 0,
    actor TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    settled_at TIMESTAMPTZ,
    UNIQUE (config_key, starts_at)
);
CREATE INDEX IF NOT EXISTS clan_war_round_due_idx ON clan_war_round (status, starts_at, ends_at);

CREATE TABLE IF NOT EXISTS clan_war_signup (
    round_id UUID NOT NULL REFERENCES clan_war_round (id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES clan (group_id) ON DELETE CASCADE,
    by_user UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (round_id, group_id)
);

CREATE TABLE IF NOT EXISTS clan_war (
    id UUID PRIMARY KEY,
    round_id UUID NOT NULL REFERENCES clan_war_round (id) ON DELETE CASCADE,
    clan_a UUID NOT NULL,
    clan_b UUID,
    score_a BIGINT NOT NULL DEFAULT 0,
    score_b BIGINT NOT NULL DEFAULT 0,
    wins_a INT NOT NULL DEFAULT 0,
    wins_b INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','settled')),
    result TEXT NOT NULL DEFAULT '' CHECK (result IN ('','a','b','draw','bye')),
    settled_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS clan_war_round_idx ON clan_war (round_id);
CREATE INDEX IF NOT EXISTS clan_war_a_idx ON clan_war (clan_a, status);
CREATE INDEX IF NOT EXISTS clan_war_b_idx ON clan_war (clan_b, status);

CREATE TABLE IF NOT EXISTS clan_war_member (
    war_id UUID NOT NULL REFERENCES clan_war (id) ON DELETE CASCADE,
    group_id UUID NOT NULL,
    user_id UUID NOT NULL,
    points BIGINT NOT NULL DEFAULT 0,
    attacks INT NOT NULL DEFAULT 0,
    wins INT NOT NULL DEFAULT 0,
    losses INT NOT NULL DEFAULT 0,
    draws INT NOT NULL DEFAULT 0,
    PRIMARY KEY (war_id, user_id)
);
CREATE INDEX IF NOT EXISTS clan_war_member_user_idx ON clan_war_member (user_id);

CREATE TABLE IF NOT EXISTS clan_war_event (
    id BIGSERIAL PRIMARY KEY,
    war_id UUID NOT NULL REFERENCES clan_war (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    group_id UUID NOT NULL,
    source_id TEXT NOT NULL,
    match_id TEXT NOT NULL DEFAULT '',
    opponent_id TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL,
    points BIGINT NOT NULL DEFAULT 0,
    counted BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (war_id, user_id, source_id)
);
CREATE INDEX IF NOT EXISTS clan_war_event_match_idx ON clan_war_event (war_id, match_id);

-- +migrate StatementBegin
DO $$
DECLARE
    def TEXT;
    vals TEXT[];
BEGIN
    SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
        WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
    IF def IS NULL OR def ~ ('[{,'']clan[},'']') THEN
        RETURN;
    END IF;
    SELECT array_agg(DISTINCT m[1]) INTO vals FROM regexp_matches(def, '''([a-z_]+)''', 'g') AS m;
    IF vals IS NULL THEN
        vals := string_to_array(substring(def from '''\{([^}]*)\}'''), ',');
    END IF;
    vals := array_append(vals, 'clan');
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type = ANY (' || quote_literal(vals::text) || '::text[]))';
END $$;
-- +migrate StatementEnd

INSERT INTO clan_level (level, xp, name, member_cap) VALUES (1, 0, '{"fa":"سطح ۱","en":"Level 1"}', 30) ON CONFLICT (level) DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS clan_war_event;
DROP TABLE IF EXISTS clan_war_member;
DROP TABLE IF EXISTS clan_war;
DROP TABLE IF EXISTS clan_war_signup;
DROP TABLE IF EXISTS clan_war_round;
DROP TABLE IF EXISTS clan_war_config;
DROP TABLE IF EXISTS clan_payout;
DROP TABLE IF EXISTS clan_xp_log;
DROP TABLE IF EXISTS clan_level_reached;
DROP TABLE IF EXISTS clan_level;
DROP TABLE IF EXISTS clan_cooldown;
DROP TABLE IF EXISTS clan_request;
DROP TABLE IF EXISTS clan_member;
DROP TABLE IF EXISTS clan;
DROP TABLE IF EXISTS clan_settings;
