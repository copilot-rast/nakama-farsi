-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS economy_event_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    timezone TEXT NOT NULL DEFAULT 'Asia/Tehran',
    week_start INT NOT NULL DEFAULT 6 CHECK (week_start BETWEEN 0 AND 6),
    retention_days INT NOT NULL DEFAULT 30 CHECK (retention_days BETWEEN 1 AND 3650),
    player_rate_limit INT NOT NULL DEFAULT 120 CHECK (player_rate_limit BETWEEN 1 AND 100000),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO economy_event_settings (id) VALUES (1) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS economy_event_type (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    client_allowed BOOLEAN NOT NULL DEFAULT false,
    value_min NUMERIC(20,4),
    value_max NUMERIC(20,4),
    rate_limit INT NOT NULL DEFAULT 30 CHECK (rate_limit BETWEEN 1 AND 100000),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_system BOOLEAN NOT NULL DEFAULT false,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (value_min IS NULL OR value_max IS NULL OR value_min <= value_max)
);

CREATE TABLE IF NOT EXISTS economy_player_event (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type_id UUID NOT NULL REFERENCES economy_event_type (id) ON DELETE CASCADE,
    value NUMERIC(20,4) NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}',
    source TEXT NOT NULL CHECK (source IN ('client','server','admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS economy_player_event_user_idx ON economy_player_event (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_player_event_type_idx ON economy_player_event (type_id, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_player_event_created_idx ON economy_player_event (created_at);

CREATE TABLE IF NOT EXISTS economy_player_event_total (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type_id UUID NOT NULL REFERENCES economy_event_type (id) ON DELETE CASCADE,
    count BIGINT NOT NULL DEFAULT 0,
    sum NUMERIC(24,4) NOT NULL DEFAULT 0,
    max NUMERIC(20,4),
    first_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, type_id)
);
CREATE INDEX IF NOT EXISTS economy_player_event_total_type_idx ON economy_player_event_total (type_id);

CREATE TABLE IF NOT EXISTS economy_player_event_day (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type_id UUID NOT NULL REFERENCES economy_event_type (id) ON DELETE CASCADE,
    day DATE NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    sum NUMERIC(24,4) NOT NULL DEFAULT 0,
    max NUMERIC(20,4),
    PRIMARY KEY (user_id, type_id, day)
);
CREATE INDEX IF NOT EXISTS economy_player_event_day_day_idx ON economy_player_event_day (day);

CREATE TABLE IF NOT EXISTS economy_event_rule (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    event_type_id UUID NOT NULL REFERENCES economy_event_type (id),
    measure TEXT NOT NULL CHECK (measure IN ('count','sum','max')),
    period TEXT NOT NULL CHECK (period IN ('total','day','week')),
    threshold NUMERIC(20,4) NOT NULL CHECK (threshold > 0),
    repeat TEXT NOT NULL CHECK (repeat IN ('once','period','every')),
    max_claims INT CHECK (max_claims IS NULL OR max_claims >= 1),
    prize_id UUID NOT NULL REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL CHECK (grant_mode IN ('auto','claim')),
    audience_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE INDEX IF NOT EXISTS economy_event_rule_type_idx ON economy_event_rule (event_type_id) WHERE is_active;

CREATE TABLE IF NOT EXISTS economy_event_rule_progress (
    rule_id UUID NOT NULL REFERENCES economy_event_rule (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    sum NUMERIC(24,4) NOT NULL DEFAULT 0,
    max NUMERIC(20,4),
    steps INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (rule_id, user_id, period_start)
);
CREATE INDEX IF NOT EXISTS economy_event_rule_progress_user_idx ON economy_event_rule_progress (user_id, rule_id);
CREATE INDEX IF NOT EXISTS economy_event_rule_progress_period_idx ON economy_event_rule_progress (period_start);

CREATE TABLE IF NOT EXISTS economy_meta_rule (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    event_type_id UUID NOT NULL REFERENCES economy_event_type (id),
    meta_id UUID NOT NULL REFERENCES economy_meta (id),
    action TEXT NOT NULL CHECK (action IN ('add','set','max')),
    value_source TEXT NOT NULL CHECK (value_source IN ('fixed','event')),
    amount NUMERIC(20,4) NOT NULL DEFAULT 1,
    audience_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE INDEX IF NOT EXISTS economy_meta_rule_type_idx ON economy_meta_rule (event_type_id) WHERE is_active;

CREATE TABLE IF NOT EXISTS economy_level_table (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    meta_id UUID NOT NULL REFERENCES economy_meta (id),
    level_meta_id UUID REFERENCES economy_meta (id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (level_meta_id IS NULL OR level_meta_id <> meta_id)
);

CREATE TABLE IF NOT EXISTS economy_level_step (
    id UUID PRIMARY KEY,
    table_id UUID NOT NULL REFERENCES economy_level_table (id) ON DELETE CASCADE,
    level INT NOT NULL CHECK (level >= 1),
    threshold NUMERIC(20,4) NOT NULL CHECK (threshold >= 0),
    name JSONB NOT NULL DEFAULT '{}',
    prize_id UUID REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL DEFAULT 'claim' CHECK (grant_mode IN ('auto','claim')),
    UNIQUE (table_id, level)
);

CREATE TABLE IF NOT EXISTS economy_user_level (
    table_id UUID NOT NULL REFERENCES economy_level_table (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    level INT NOT NULL DEFAULT 0,
    reached_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (table_id, user_id)
);

CREATE TABLE IF NOT EXISTS economy_streak (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    event_type_id UUID NOT NULL REFERENCES economy_event_type (id),
    period_days INT NOT NULL DEFAULT 1 CHECK (period_days BETWEEN 1 AND 30),
    grace_periods INT NOT NULL DEFAULT 0 CHECK (grace_periods BETWEEN 0 AND 30),
    miss_action TEXT NOT NULL DEFAULT 'reset' CHECK (miss_action IN ('reset','step_back')),
    cycle_length INT CHECK (cycle_length IS NULL OR cycle_length BETWEEN 1 AND 365),
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_streak_reward (
    id UUID PRIMARY KEY,
    streak_id UUID NOT NULL REFERENCES economy_streak (id) ON DELETE CASCADE,
    day INT NOT NULL CHECK (day >= 1),
    prize_id UUID NOT NULL REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL DEFAULT 'claim' CHECK (grant_mode IN ('auto','claim')),
    UNIQUE (streak_id, day)
);

CREATE TABLE IF NOT EXISTS economy_user_streak (
    streak_id UUID NOT NULL REFERENCES economy_streak (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    count INT NOT NULL DEFAULT 0,
    best INT NOT NULL DEFAULT 0,
    run INT NOT NULL DEFAULT 1,
    grace_used INT NOT NULL DEFAULT 0,
    last_period BIGINT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (streak_id, user_id)
);

CREATE TABLE IF NOT EXISTS economy_reward_claim (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('event','level','streak')),
    owner_id UUID NOT NULL,
    ref_id UUID NOT NULL,
    scope TEXT NOT NULL,
    prize_id UUID NOT NULL REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL CHECK (grant_mode IN ('auto','claim')),
    status TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('ready','granted','failed')),
    value NUMERIC(24,4),
    error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    attempted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    granted_at TIMESTAMPTZ,
    UNIQUE (kind, ref_id, user_id, scope)
);
CREATE INDEX IF NOT EXISTS economy_reward_claim_user_idx ON economy_reward_claim (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_reward_claim_owner_idx ON economy_reward_claim (owner_id, user_id);
CREATE INDEX IF NOT EXISTS economy_reward_claim_retry_idx ON economy_reward_claim (created_at) WHERE status <> 'granted' AND grant_mode = 'auto';

ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak'));

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'login', '{"fa":"ورود به بازی","en":"Sign-in","ar":"تسجيل الدخول"}', '{"fa":"هر بار که بازیکن وارد بازی می‌شود، سرور این رویداد را ثبت می‌کند.","en":"Recorded by the server each time a player signs in.","ar":"يسجّله الخادم في كل مرة يسجّل فيها اللاعب الدخول."}', false, NULL, 30, true, 1),
(gen_random_uuid(), 'purchase_completed', '{"fa":"خرید موفق","en":"Purchase completed","ar":"عملية شراء ناجحة"}', '{"fa":"بعد از هر خرید موفق ثبت می‌شود. مقدار آن مبلغ نهایی به تومان است و در خرید با سکه، تعداد سکه.","en":"Recorded after every successful purchase. Its value is the final price in toman, or the coin amount for coin purchases.","ar":"يُسجَّل بعد كل عملية شراء ناجحة. قيمته السعر النهائي بالتومان أو عدد العملات."}', false, 0, 30, true, 2),
(gen_random_uuid(), 'gift_redeemed', '{"fa":"استفاده از کد هدیه","en":"Gift code redeemed","ar":"استخدام رمز الهدية"}', '{"fa":"وقتی بازیکن یک کد هدیه را با موفقیت وارد می‌کند.","en":"Recorded when a player redeems a gift code.","ar":"يُسجَّل عندما يستخدم اللاعب رمز هدية بنجاح."}', false, NULL, 30, true, 3),
(gen_random_uuid(), 'level_up', '{"fa":"بالا رفتن سطح","en":"Level up","ar":"رفع المستوى"}', '{"fa":"وقتی بازیکن در یک جدول سطح به سطح بالاتر می‌رسد. مقدار آن سطح تازه است.","en":"Recorded when a player reaches a higher level in a level table. Its value is the new level.","ar":"يُسجَّل عندما يصل اللاعب إلى مستوى أعلى. قيمته المستوى الجديد."}', false, 0, 30, true, 4)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds player events, rules, levels and streaks and cannot be reverted automatically; restore from backup', '20261111020000-economy-events'; END $$;
-- +migrate StatementEnd
