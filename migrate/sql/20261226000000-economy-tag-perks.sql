-- +migrate Up
ALTER TABLE base_user_tag ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS base_user_tag_expires_idx ON base_user_tag (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS economy_tag_discount (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    percent INT NOT NULL CHECK (percent BETWEEN 1 AND 100),
    all_items BOOLEAN NOT NULL DEFAULT false,
    item_keys TEXT[] NOT NULL DEFAULT '{}',
    item_type_ids TEXT[] NOT NULL DEFAULT '{}',
    all_products BOOLEAN NOT NULL DEFAULT false,
    product_ids TEXT[] NOT NULL DEFAULT '{}',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_tag_discount_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'wheel';
ALTER TABLE economy_chance_game DROP CONSTRAINT IF EXISTS economy_chance_game_kind_check;
ALTER TABLE economy_chance_game ADD CONSTRAINT economy_chance_game_kind_check CHECK (kind IN ('wheel','chest'));
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS daily_limit INT CHECK (daily_limit IS NULL OR daily_limit BETWEEN 1 AND 100000);
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS total_limit INT CHECK (total_limit IS NULL OR total_limit BETWEEN 1 AND 10000000);
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS show_odds BOOLEAN NOT NULL DEFAULT true;

-- +migrate Down
ALTER TABLE economy_chance_game DROP COLUMN IF EXISTS show_odds;
ALTER TABLE economy_chance_game DROP COLUMN IF EXISTS total_limit;
ALTER TABLE economy_chance_game DROP COLUMN IF EXISTS daily_limit;
ALTER TABLE economy_chance_game DROP CONSTRAINT IF EXISTS economy_chance_game_kind_check;
ALTER TABLE economy_chance_game DROP COLUMN IF EXISTS kind;
DROP TABLE IF EXISTS economy_tag_discount_settings;
DROP TABLE IF EXISTS economy_tag_discount;
DROP INDEX IF EXISTS base_user_tag_expires_idx;
ALTER TABLE base_user_tag DROP COLUMN IF EXISTS expires_at;
