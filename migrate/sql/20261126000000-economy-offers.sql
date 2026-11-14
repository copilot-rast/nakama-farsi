-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS economy_offer (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL DEFAULT 'daily' CHECK (kind IN ('daily','entry')),
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    badge JSONB NOT NULL DEFAULT '{}',
    product_id UUID NOT NULL REFERENCES economy_product (id),
    price_mode TEXT NOT NULL DEFAULT 'discount' CHECK (price_mode IN ('discount','price')),
    discount_percent INT NOT NULL DEFAULT 0 CHECK (discount_percent BETWEEN 0 AND 100),
    price_toman BIGINT CHECK (price_toman IS NULL OR price_toman >= 0),
    coin_price JSONB,
    sku JSONB NOT NULL DEFAULT '{}',
    purchase_limit INT NOT NULL DEFAULT 1 CHECK (purchase_limit BETWEEN 1 AND 1000),
    weight INT NOT NULL DEFAULT 10 CHECK (weight BETWEEN 0 AND 1000000),
    audience_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    exclude_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    min_level INT NOT NULL DEFAULT 0 CHECK (min_level >= 0),
    max_level INT CHECK (max_level IS NULL OR max_level >= 0),
    level_table_id UUID REFERENCES economy_level_table (id),
    meta_conditions JSONB NOT NULL DEFAULT '[]',
    min_account_days INT CHECK (min_account_days IS NULL OR min_account_days >= 0),
    max_account_days INT CHECK (max_account_days IS NULL OR max_account_days >= 0),
    trigger_type TEXT NOT NULL DEFAULT 'first_entry' CHECK (trigger_type IN ('first_entry','level','meta','tag','account_days','lapsed')),
    trigger_value NUMERIC(20,4),
    trigger_ref TEXT NOT NULL DEFAULT '',
    duration_minutes INT NOT NULL DEFAULT 1440 CHECK (duration_minutes BETWEEN 1 AND 525600),
    repeat TEXT NOT NULL DEFAULT 'once' CHECK (repeat IN ('once','cooldown')),
    cooldown_hours INT NOT NULL DEFAULT 168 CHECK (cooldown_hours BETWEEN 0 AND 87600),
    max_triggers INT CHECK (max_triggers IS NULL OR max_triggers >= 1),
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE INDEX IF NOT EXISTS economy_offer_kind_idx ON economy_offer (kind, is_active, sort);

CREATE TABLE IF NOT EXISTS economy_offer_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per player and daily period; holds the reroll count and serialises generation.
CREATE TABLE IF NOT EXISTS economy_user_offer_period (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    period_start TIMESTAMPTZ NOT NULL,
    rerolls INT NOT NULL DEFAULT 0,
    generated BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, period_start)
);

-- A personal copy of an offer: a daily slot or a triggered entry offer, with the price frozen when it was made.
CREATE TABLE IF NOT EXISTS economy_user_offer (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    offer_id UUID NOT NULL REFERENCES economy_offer (id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('daily','entry')),
    period_start TIMESTAMPTZ,
    slot INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','replaced','closed')),
    original_toman BIGINT,
    final_toman BIGINT,
    coin_currency TEXT,
    coin_amount BIGINT,
    coin_original BIGINT,
    discount_percent INT NOT NULL DEFAULT 0,
    sku JSONB NOT NULL DEFAULT '{}',
    purchase_limit INT NOT NULL DEFAULT 1,
    bought INT NOT NULL DEFAULT 0,
    reroll INT NOT NULL DEFAULT 0,
    forgotten BOOLEAN NOT NULL DEFAULT false,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    shown_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    last_bought_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS economy_user_offer_user_idx ON economy_user_offer (user_id, kind, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_user_offer_offer_idx ON economy_user_offer (offer_id, created_at);

-- Login times for the lapsed-return trigger.
CREATE TABLE IF NOT EXISTS economy_user_offer_state (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    last_login_at TIMESTAMPTZ,
    prev_login_at TIMESTAMPTZ,
    logins BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS user_offer_id UUID REFERENCES economy_user_offer (id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS economy_payment_user_offer_idx ON economy_payment (user_offer_id) WHERE user_offer_id IS NOT NULL;

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'Offer:Bought', '{"fa":"خرید پیشنهاد ویژه","en":"Offer bought","ar":"شراء عرض"}', '{"fa":"وقتی بازیکن یک پیشنهاد روزانه یا ورودی را می‌خرد. کلید پیشنهاد در key و نوع آن در kind می‌آید.","en":"Recorded when a player buys a daily or entry offer. The offer key is in key and its kind in kind.","ar":"يُسجَّل عندما يشتري اللاعب عرضًا يوميًا أو عرض دخول."}', false, 0, 600, true, 13)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds offers and player offer history and cannot be reverted automatically; restore from backup', '20261126000000-economy-offers'; END $$;
-- +migrate StatementEnd
