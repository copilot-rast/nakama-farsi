-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS economy_category (
    id UUID PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('product','prize','gift','currency')),
    key TEXT NOT NULL,
    name JSONB NOT NULL,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, key)
);

CREATE TABLE IF NOT EXISTS economy_tag (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_currency (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z0-9_]{2,32}$'),
    name JSONB NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('soft','hard')),
    default_value BIGINT NOT NULL DEFAULT 0 CHECK (default_value >= 0),
    icon_url TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_user_currency_init (
    user_id UUID NOT NULL,
    currency_id UUID NOT NULL REFERENCES economy_currency (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, currency_id)
);

CREATE TABLE IF NOT EXISTS economy_meta (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL,
    default_value NUMERIC(20,4) NOT NULL DEFAULT 0,
    server_only BOOLEAN NOT NULL DEFAULT false,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_user_meta (
    user_id UUID NOT NULL,
    meta_id UUID NOT NULL REFERENCES economy_meta (id) ON DELETE CASCADE,
    value NUMERIC(20,4) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, meta_id)
);

CREATE TABLE IF NOT EXISTS economy_prize (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL,
    message JSONB NOT NULL DEFAULT '{}',
    category_id UUID REFERENCES economy_category (id),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    is_repeatable BOOLEAN NOT NULL DEFAULT true,
    items JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_user_prize (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    prize_id UUID NOT NULL REFERENCES economy_prize (id),
    source_type TEXT NOT NULL CHECK (source_type IN ('payment','gift','admin')),
    source_id TEXT NOT NULL,
    claimed BOOLEAN NOT NULL DEFAULT false,
    error TEXT,
    once_key TEXT UNIQUE,
    rewards JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id, prize_id)
);
CREATE INDEX IF NOT EXISTS economy_user_prize_user_idx ON economy_user_prize (user_id, created_at);

CREATE TABLE IF NOT EXISTS economy_product (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL,
    description JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT false,
    category_id UUID REFERENCES economy_category (id),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    price_toman BIGINT CHECK (price_toman IS NULL OR price_toman >= 0),
    coin_price JSONB,
    discount_percent INT NOT NULL DEFAULT 0 CHECK (discount_percent BETWEEN 0 AND 100),
    payment_methods TEXT[] NOT NULL DEFAULT '{}',
    sku JSONB NOT NULL DEFAULT '{}',
    is_subscription BOOLEAN NOT NULL DEFAULT false,
    prize_id UUID NOT NULL REFERENCES economy_prize (id),
    icon_url TEXT NOT NULL DEFAULT '',
    cover_url TEXT NOT NULL DEFAULT '',
    min_level INT NOT NULL DEFAULT 0 CHECK (min_level >= 0),
    metadata JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_gift (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('reward','discount')),
    code TEXT UNIQUE,
    prize_id UUID REFERENCES economy_prize (id),
    discount_percent INT CHECK (discount_percent IS NULL OR discount_percent BETWEEN 1 AND 100),
    is_promo BOOLEAN NOT NULL DEFAULT false,
    per_user_limit INT NOT NULL DEFAULT 1 CHECK (per_user_limit >= 1),
    max_total_uses INT CHECK (max_total_uses IS NULL OR max_total_uses >= 1),
    total_used INT NOT NULL DEFAULT 0,
    valid_from TIMESTAMPTZ,
    expire_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_gift_use (
    id UUID PRIMARY KEY,
    gift_id UUID NOT NULL REFERENCES economy_gift (id),
    user_id UUID NOT NULL,
    payment_id UUID,
    status TEXT NOT NULL CHECK (status IN ('reserved','used','released')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS economy_gift_use_gift_user_idx ON economy_gift_use (gift_id, user_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS economy_gift_use_payment_idx ON economy_gift_use (payment_id) WHERE payment_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS economy_payment_method (
    method TEXT PRIMARY KEY CHECK (method IN ('zarinpal','zibal','bazaar','myket','in_app')),
    is_active BOOLEAN NOT NULL DEFAULT false,
    config TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_payment (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES economy_product (id),
    method TEXT NOT NULL,
    sandbox BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL CHECK (status IN ('pending','success','failed','expired')),
    original_toman BIGINT NOT NULL DEFAULT 0,
    final_toman BIGINT NOT NULL DEFAULT 0,
    amount_rial BIGINT NOT NULL DEFAULT 0,
    coin_currency TEXT,
    coin_amount BIGINT,
    gift_id UUID REFERENCES economy_gift (id),
    discount_percent INT NOT NULL DEFAULT 0,
    authority TEXT,
    track_id TEXT,
    ref_id TEXT,
    store_token TEXT,
    card_mask TEXT,
    outcome TEXT,
    error TEXT,
    reward_granted BOOLEAN NOT NULL DEFAULT false,
    reward_error TEXT,
    inquiry_count INT NOT NULL DEFAULT 0,
    inquired_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS economy_payment_authority_idx ON economy_payment (method, authority) WHERE authority IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS economy_payment_track_idx ON economy_payment (method, track_id) WHERE track_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS economy_payment_store_token_idx ON economy_payment (method, store_token) WHERE store_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS economy_payment_status_idx ON economy_payment (status, created_at);
CREATE INDEX IF NOT EXISTS economy_payment_user_idx ON economy_payment (user_id, created_at);

CREATE TABLE IF NOT EXISTS economy_payment_log (
    id UUID PRIMARY KEY,
    payment_id UUID NOT NULL REFERENCES economy_payment (id),
    action TEXT NOT NULL CHECK (action IN ('request','callback','verify','inquiry','store_validate')),
    url TEXT NOT NULL DEFAULT '',
    request JSONB NOT NULL DEFAULT '{}',
    response JSONB NOT NULL DEFAULT '{}',
    http_status INT,
    duration_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS economy_payment_log_payment_idx ON economy_payment_log (payment_id, created_at);

INSERT INTO economy_payment_method (method) VALUES ('zarinpal'), ('zibal'), ('bazaar'), ('myket'), ('in_app') ON CONFLICT (method) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % retains payment and grant records and cannot be reverted automatically; restore from backup', '20261030000000-economy-schema'; END $$;
-- +migrate StatementEnd
