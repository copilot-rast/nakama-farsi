-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS campaign_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO campaign_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS campaign_chapter (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    unlock JSONB NOT NULL DEFAULT '[]',
    milestones JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS campaign_stage (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    chapter_id UUID NOT NULL REFERENCES campaign_chapter (id),
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    mode TEXT NOT NULL DEFAULT 'ordered' CHECK (mode IN ('ordered','repeatable','once')),
    unlock JSONB,
    time_limit_seconds NUMERIC(10,2),
    min_seconds NUMERIC(10,2),
    star_rules JSONB,
    star_multipliers JSONB,
    rewards JSONB NOT NULL DEFAULT '{}',
    xp INT NOT NULL DEFAULT 0 CHECK (xp >= 0),
    tier_prizes JSONB NOT NULL DEFAULT '[]',
    reward_group TEXT NOT NULL DEFAULT '',
    cooldown_seconds INT NOT NULL DEFAULT 0 CHECK (cooldown_seconds >= 0),
    daily_limit INT NOT NULL DEFAULT 0 CHECK (daily_limit >= 0),
    xp_when_limited BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS campaign_stage_chapter_idx ON campaign_stage (chapter_id, sort);

CREATE TABLE IF NOT EXISTS campaign_freeride (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    rewards JSONB NOT NULL DEFAULT '{}',
    prize_key TEXT NOT NULL DEFAULT '',
    prize_group_key TEXT NOT NULL DEFAULT '',
    xp INT NOT NULL DEFAULT 0 CHECK (xp >= 0),
    reward_group TEXT NOT NULL DEFAULT '',
    cooldown_seconds INT NOT NULL DEFAULT 0 CHECK (cooldown_seconds >= 0),
    daily_limit INT NOT NULL DEFAULT 0 CHECK (daily_limit >= 0),
    min_level INT NOT NULL DEFAULT 0 CHECK (min_level >= 0),
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS campaign_attempt (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    stage_id UUID NOT NULL REFERENCES campaign_stage (id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    closed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','closed','paying','done')),
    success BOOLEAN,
    stars INT NOT NULL DEFAULT 0,
    result JSONB,
    plan JSONB,
    pay_tries INT NOT NULL DEFAULT 0,
    rejected INT NOT NULL DEFAULT 0,
    last_reject TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS campaign_attempt_open_idx ON campaign_attempt (user_id) WHERE status = 'open';
CREATE INDEX IF NOT EXISTS campaign_attempt_stage_idx ON campaign_attempt (stage_id, started_at);
CREATE INDEX IF NOT EXISTS campaign_attempt_paying_idx ON campaign_attempt (completed_at) WHERE status = 'paying';
CREATE INDEX IF NOT EXISTS campaign_attempt_started_idx ON campaign_attempt (started_at);

CREATE TABLE IF NOT EXISTS campaign_user_stage (
    user_id UUID NOT NULL,
    stage_id UUID NOT NULL REFERENCES campaign_stage (id) ON DELETE CASCADE,
    best_stars INT NOT NULL DEFAULT 0,
    claimed JSONB NOT NULL DEFAULT '{}',
    claimed_xp INT NOT NULL DEFAULT 0,
    tiers_paid INT NOT NULL DEFAULT 0,
    completions INT NOT NULL DEFAULT 0,
    wins INT NOT NULL DEFAULT 0,
    best_seconds NUMERIC(10,2),
    first_win_at TIMESTAMPTZ,
    last_completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, stage_id)
);

CREATE INDEX IF NOT EXISTS campaign_user_stage_stage_idx ON campaign_user_stage (stage_id);

CREATE TABLE IF NOT EXISTS campaign_user (
    user_id UUID PRIMARY KEY,
    xp BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS campaign_user_chapter (
    user_id UUID NOT NULL,
    chapter_id UUID NOT NULL REFERENCES campaign_chapter (id) ON DELETE CASCADE,
    milestones JSONB NOT NULL DEFAULT '[]',
    PRIMARY KEY (user_id, chapter_id)
);

CREATE TABLE IF NOT EXISTS campaign_reward_group (
    user_id UUID NOT NULL,
    group_key TEXT NOT NULL,
    last_at TIMESTAMPTZ,
    day TEXT NOT NULL DEFAULT '',
    count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, group_key)
);

CREATE TABLE IF NOT EXISTS campaign_claim (
    user_id UUID NOT NULL,
    request_id TEXT NOT NULL,
    freeride_id UUID NOT NULL REFERENCES campaign_freeride (id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'paying' CHECK (status IN ('paying','done')),
    result JSONB NOT NULL DEFAULT '{}',
    plan JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, request_id)
);

CREATE INDEX IF NOT EXISTS campaign_claim_expires_idx ON campaign_claim (expires_at);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Campaign migration % holds player progress and paid rewards and cannot be reverted automatically; restore from backup', '20261205000000-campaign-schema'; END $$;
-- +migrate StatementEnd
