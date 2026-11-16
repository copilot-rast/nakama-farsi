-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS leaderboardx_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO leaderboardx_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS leaderboardx_board (
    id UUID PRIMARY KEY,
    -- the Nakama core leaderboard id
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order TEXT NOT NULL DEFAULT 'desc' CHECK (sort_order IN ('desc','asc')),
    operator TEXT NOT NULL DEFAULT 'incr' CHECK (operator IN ('best','set','incr','decr')),
    -- none | daily | weekly | monthly | custom; reset_schedule is the core CRON (UTC), '' = never resets
    period TEXT NOT NULL DEFAULT 'none',
    reset_schedule TEXT NOT NULL DEFAULT '',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    -- 0 = unlimited; after this many closed periods the board stops (legacy LifeCycle)
    max_periods INT NOT NULL DEFAULT 0 CHECK (max_periods >= 0),
    -- {type: economy_event|game_event|client|server, event, context, field, calc: value|count, multiplier}
    source JSONB NOT NULL DEFAULT '{}',
    -- {clientSubmit, maxPerWrite, perMinute, onExceed: reject|clamp}
    limits JSONB NOT NULL DEFAULT '{}',
    -- none | score | rank | percent
    tier_basis TEXT NOT NULL DEFAULT 'none' CHECK (tier_basis IN ('none','score','rank','percent')),
    -- [{key, name:{lang:text}, iconUrl, threshold}] best tier first
    tiers JSONB NOT NULL DEFAULT '[]',
    -- [{tier, fromRank, toRank, currencies:{key:n}, prizeKey, prizeGroupKey}]
    rewards JSONB NOT NULL DEFAULT '[]',
    reward_min_score BIGINT NOT NULL DEFAULT 0,
    -- '' = the settings default
    pay_mode TEXT NOT NULL DEFAULT '' CHECK (pay_mode IN ('','auto','claim')),
    periods_closed INT NOT NULL DEFAULT 0,
    finished_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS leaderboardx_period (
    id UUID PRIMARY KEY,
    board_id UUID REFERENCES leaderboardx_board (id) ON DELETE SET NULL,
    board_key TEXT NOT NULL,
    board_name JSONB NOT NULL DEFAULT '{}',
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ NOT NULL,
    expiry_time TIMESTAMPTZ NOT NULL,
    -- reset | manual | restart | end
    reason TEXT NOT NULL DEFAULT 'reset',
    players INT NOT NULL DEFAULT 0,
    rewarded INT NOT NULL DEFAULT 0,
    -- board tiers, rewards, tier basis, sort order as they were at the end
    config JSONB NOT NULL DEFAULT '{}',
    pay_mode TEXT NOT NULL DEFAULT 'auto',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (board_key, period_end)
);
CREATE INDEX IF NOT EXISTS leaderboardx_period_board_idx ON leaderboardx_period (board_id, period_end DESC);

CREATE TABLE IF NOT EXISTS leaderboardx_result (
    period_id UUID NOT NULL REFERENCES leaderboardx_period (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    rank INT NOT NULL,
    tier TEXT NOT NULL DEFAULT '',
    tier_rank INT NOT NULL DEFAULT 0,
    score BIGINT NOT NULL,
    subscore BIGINT NOT NULL DEFAULT 0,
    username TEXT NOT NULL DEFAULT '',
    -- {currencies:{key:n}, prizes:[{prizeKey, prizeGroupKey, sourceId, paid, error}], rows:[i]}
    plan JSONB,
    -- none | pending (auto, waits for the worker) | claimable | paid | failed | expired
    status TEXT NOT NULL DEFAULT 'none' CHECK (status IN ('none','pending','claimable','paid','failed','expired')),
    pay_tries INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    paid JSONB,
    claim_expires_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    PRIMARY KEY (period_id, user_id)
);
CREATE INDEX IF NOT EXISTS leaderboardx_result_user_idx ON leaderboardx_result (user_id, period_id);
CREATE INDEX IF NOT EXISTS leaderboardx_result_open_idx ON leaderboardx_result (status) WHERE status IN ('pending','claimable');
CREATE INDEX IF NOT EXISTS leaderboardx_result_rank_idx ON leaderboardx_result (period_id, rank);

CREATE TABLE IF NOT EXISTS leaderboardx_score_log (
    board_id UUID NOT NULL,
    source_id TEXT NOT NULL,
    user_id UUID NOT NULL,
    -- economy_event | game_event | client | admin
    source TEXT NOT NULL,
    value NUMERIC NOT NULL DEFAULT 0,
    delta BIGINT NOT NULL DEFAULT 0,
    -- '' accepted, clamped, refused reason
    outcome TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (board_id, source_id)
);
CREATE INDEX IF NOT EXISTS leaderboardx_score_log_time_idx ON leaderboardx_score_log (created_at);
CREATE INDEX IF NOT EXISTS leaderboardx_score_log_user_idx ON leaderboardx_score_log (user_id, board_id, created_at);

-- +migrate StatementBegin
DO $$
DECLARE def text;
BEGIN
  SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
  WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
  IF def IS NOT NULL AND position('''leaderboard''' IN def) = 0 THEN
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check ' || replace(def, 'ARRAY[', 'ARRAY[''leaderboard''::text, ');
  END IF;
END $$;
-- +migrate StatementEnd

-- +migrate Down
DROP TABLE IF EXISTS leaderboardx_score_log;
DROP TABLE IF EXISTS leaderboardx_result;
DROP TABLE IF EXISTS leaderboardx_period;
DROP TABLE IF EXISTS leaderboardx_board;
DROP TABLE IF EXISTS leaderboardx_settings;
