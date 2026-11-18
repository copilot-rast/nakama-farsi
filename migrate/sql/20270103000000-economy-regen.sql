-- +migrate Up
ALTER TABLE economy_currency ADD COLUMN IF NOT EXISTS regen JSONB;

CREATE TABLE IF NOT EXISTS economy_user_regen (
    user_id UUID NOT NULL,
    currency_key TEXT NOT NULL,
    anchor_at TIMESTAMPTZ NOT NULL,
    last_value BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, currency_key)
);

CREATE TABLE IF NOT EXISTS economy_regen_spend (
    user_id UUID NOT NULL,
    source_id TEXT NOT NULL,
    currency_key TEXT NOT NULL,
    amount BIGINT NOT NULL,
    free BOOLEAN NOT NULL DEFAULT false,
    refunded BIGINT,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, source_id)
);

CREATE INDEX IF NOT EXISTS economy_regen_spend_created_idx ON economy_regen_spend (created_at);

CREATE TABLE IF NOT EXISTS economy_regen_refill (
    user_id UUID NOT NULL,
    request_id TEXT NOT NULL,
    currency_key TEXT NOT NULL,
    answer JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, request_id)
);

-- +migrate Down
DROP TABLE IF EXISTS economy_regen_refill;
DROP TABLE IF EXISTS economy_regen_spend;
DROP TABLE IF EXISTS economy_user_regen;
ALTER TABLE economy_currency DROP COLUMN IF EXISTS regen;
