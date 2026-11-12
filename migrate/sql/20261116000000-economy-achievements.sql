-- +migrate Up notransaction
ALTER TABLE base_category DROP CONSTRAINT IF EXISTS base_category_key_type_check;
ALTER TABLE base_category ADD CONSTRAINT base_category_key_type_check CHECK (key_type IN ('product','prize','gift','currency','faq','news','task','item','item_type','achievement'));
ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','gift','currency','news','faq','general','user','task','item','achievement'));

ALTER TABLE economy_reward_claim DROP CONSTRAINT IF EXISTS economy_reward_claim_kind_check;
ALTER TABLE economy_reward_claim ADD CONSTRAINT economy_reward_claim_kind_check CHECK (kind IN ('event','level','streak','referral','task','achievement'));
ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak','referral','task','achievement'));

CREATE TABLE IF NOT EXISTS economy_achievement (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    hint JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    locked_icon_url TEXT NOT NULL DEFAULT '',
    category_id UUID REFERENCES base_category (id),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    rarity TEXT NOT NULL DEFAULT 'common' CHECK (rarity IN ('common','uncommon','rare','epic','legendary')),
    points INT NOT NULL DEFAULT 0 CHECK (points BETWEEN 0 AND 100000),
    is_hidden BOOLEAN NOT NULL DEFAULT false,
    source TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('event','task','manual')),
    event_type_id UUID REFERENCES economy_event_type (id),
    metric TEXT NOT NULL DEFAULT 'count' CHECK (metric IN ('count','sum_value','high_value')),
    value_min NUMERIC(20,4),
    value_max NUMERIC(20,4),
    task_id UUID REFERENCES economy_task (id),
    task_group_id UUID REFERENCES economy_task_group (id),
    notify BOOLEAN NOT NULL DEFAULT false,
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (source <> 'event' OR event_type_id IS NOT NULL),
    CHECK (source <> 'task' OR ((task_id IS NULL) <> (task_group_id IS NULL))),
    CHECK (value_min IS NULL OR value_max IS NULL OR value_min <= value_max)
);
CREATE INDEX IF NOT EXISTS economy_achievement_event_idx ON economy_achievement (event_type_id) WHERE is_active AND source = 'event';
CREATE INDEX IF NOT EXISTS economy_achievement_task_idx ON economy_achievement (task_id) WHERE task_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS economy_achievement_group_idx ON economy_achievement (task_group_id) WHERE task_group_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS economy_achievement_tier (
    id UUID PRIMARY KEY,
    achievement_id UUID NOT NULL REFERENCES economy_achievement (id) ON DELETE CASCADE,
    tier INT NOT NULL CHECK (tier BETWEEN 1 AND 10),
    threshold NUMERIC(20,4) NOT NULL DEFAULT 1 CHECK (threshold > 0),
    name JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    prize_id UUID REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL DEFAULT 'claim' CHECK (grant_mode IN ('auto','claim')),
    UNIQUE (achievement_id, tier)
);

CREATE TABLE IF NOT EXISTS economy_user_achievement (
    achievement_id UUID NOT NULL REFERENCES economy_achievement (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    sum NUMERIC(24,4) NOT NULL DEFAULT 0,
    max NUMERIC(20,4),
    tier INT NOT NULL DEFAULT 0 CHECK (tier >= 0),
    unlocked_at TIMESTAMPTZ,
    tier_at TIMESTAMPTZ,
    source TEXT NOT NULL DEFAULT 'progress' CHECK (source IN ('progress','event','task','prize','admin')),
    source_ref TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (achievement_id, user_id)
);
CREATE INDEX IF NOT EXISTS economy_user_achievement_user_idx ON economy_user_achievement (user_id, achievement_id);
CREATE INDEX IF NOT EXISTS economy_user_achievement_unlocked_idx ON economy_user_achievement (achievement_id) WHERE tier > 0;

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'achievement_unlocked', '{"fa":"باز شدن دستاورد","en":"Achievement unlocked","ar":"فتح إنجاز"}', '{"fa":"وقتی بازیکن دستاوردی را باز می‌کند، یک بار برای هر دستاورد. مقدار آن امتیاز دستاورد است.","en":"Recorded once when a player unlocks an achievement. Its value is the achievement points.","ar":"يُسجَّل مرة واحدة عندما يفتح اللاعب إنجازًا. قيمته نقاط الإنجاز."}', false, 0, 60, true, 10)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds achievements, their tiers and player achievements and cannot be reverted automatically; restore from backup', '20261116000000-economy-achievements'; END $$;
-- +migrate StatementEnd
