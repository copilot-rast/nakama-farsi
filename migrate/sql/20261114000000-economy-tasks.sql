-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS economy_task_group (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    category_id UUID REFERENCES base_category (id),
    active_count INT NOT NULL DEFAULT 1 CHECK (active_count BETWEEN 1 AND 20),
    order_mode TEXT NOT NULL DEFAULT 'ordered' CHECK (order_mode IN ('ordered','random')),
    handout TEXT NOT NULL DEFAULT 'auto' CHECK (handout IN ('auto','request','timer')),
    handout_minutes INT NOT NULL DEFAULT 0 CHECK (handout_minutes BETWEEN 0 AND 525600),
    is_repeatable BOOLEAN NOT NULL DEFAULT false,
    audience_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    level_table_id UUID REFERENCES economy_level_table (id),
    min_level INT CHECK (min_level IS NULL OR min_level >= 0),
    max_level INT CHECK (max_level IS NULL OR max_level >= 0),
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
    CHECK (min_level IS NULL OR max_level IS NULL OR min_level <= max_level)
);

CREATE TABLE IF NOT EXISTS economy_task (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    category_id UUID REFERENCES base_category (id),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    condition TEXT NOT NULL DEFAULT 'event' CHECK (condition IN ('event','children')),
    event_type_id UUID REFERENCES economy_event_type (id),
    metric TEXT NOT NULL DEFAULT 'count' CHECK (metric IN ('count','high_value','sum_value')),
    target NUMERIC(20,4) NOT NULL DEFAULT 1 CHECK (target > 0),
    value_min NUMERIC(20,4),
    value_max NUMERIC(20,4),
    repeat TEXT NOT NULL DEFAULT 'once' CHECK (repeat IN ('once','daily','weekly','interval')),
    repeat_minutes INT CHECK (repeat_minutes IS NULL OR repeat_minutes BETWEEN 1 AND 525600),
    max_completions INT CHECK (max_completions IS NULL OR max_completions >= 1),
    parent_id UUID REFERENCES economy_task (id),
    depends_on TEXT[] NOT NULL DEFAULT '{}',
    group_id UUID REFERENCES economy_task_group (id) ON DELETE SET NULL,
    prize_id UUID REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL DEFAULT 'claim' CHECK (grant_mode IN ('auto','claim')),
    notify BOOLEAN NOT NULL DEFAULT false,
    audience_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    level_table_id UUID REFERENCES economy_level_table (id),
    min_level INT CHECK (min_level IS NULL OR min_level >= 0),
    max_level INT CHECK (max_level IS NULL OR max_level >= 0),
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (condition = 'children' OR event_type_id IS NOT NULL),
    CHECK (value_min IS NULL OR value_max IS NULL OR value_min <= value_max),
    CHECK (repeat <> 'interval' OR repeat_minutes IS NOT NULL),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
    CHECK (min_level IS NULL OR max_level IS NULL OR min_level <= max_level),
    CHECK (parent_id IS NULL OR parent_id <> id)
);
CREATE INDEX IF NOT EXISTS economy_task_event_idx ON economy_task (event_type_id) WHERE is_active;
CREATE INDEX IF NOT EXISTS economy_task_parent_idx ON economy_task (parent_id);
CREATE INDEX IF NOT EXISTS economy_task_group_idx ON economy_task (group_id);

CREATE TABLE IF NOT EXISTS economy_user_task (
    task_id UUID NOT NULL REFERENCES economy_task (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    window_start DATE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed','claimed')),
    count BIGINT NOT NULL DEFAULT 0,
    sum NUMERIC(24,4) NOT NULL DEFAULT 0,
    max NUMERIC(20,4),
    completed_at TIMESTAMPTZ,
    completed_by TEXT,
    claim_id UUID,
    claimed_at TIMESTAMPTZ,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (task_id, user_id, scope)
);
CREATE INDEX IF NOT EXISTS economy_user_task_user_idx ON economy_user_task (user_id, task_id);
CREATE UNIQUE INDEX IF NOT EXISTS economy_user_task_claim_idx ON economy_user_task (claim_id) WHERE claim_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS economy_user_task_window_idx ON economy_user_task (window_start) WHERE window_start IS NOT NULL;

CREATE TABLE IF NOT EXISTS economy_user_task_group (
    group_id UUID NOT NULL REFERENCES economy_task_group (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    cycle INT NOT NULL DEFAULT 1,
    handed TEXT[] NOT NULL DEFAULT '{}',
    next_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

ALTER TABLE economy_reward_claim DROP CONSTRAINT IF EXISTS economy_reward_claim_kind_check;
ALTER TABLE economy_reward_claim ADD CONSTRAINT economy_reward_claim_kind_check CHECK (kind IN ('event','level','streak','referral','task'));
ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak','referral','task'));

ALTER TABLE base_category DROP CONSTRAINT IF EXISTS base_category_key_type_check;
ALTER TABLE base_category ADD CONSTRAINT base_category_key_type_check CHECK (key_type IN ('product','prize','gift','currency','faq','news','task'));
ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','gift','currency','news','faq','general','user','task'));

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds tasks, task groups and player task progress and cannot be reverted automatically; restore from backup', '20261114000000-economy-tasks'; END $$;
-- +migrate StatementEnd
