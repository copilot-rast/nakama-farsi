-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS blockworld_link (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    first_token_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_token_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    tokens BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS blockworld_block (
    id INT PRIMARY KEY CHECK (id BETWEEN 0 AND 255),
    name TEXT NOT NULL,
    price BIGINT NOT NULL CHECK (price > 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS blockworld_purchase (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    block_id INT NOT NULL,
    price BIGINT NOT NULL CHECK (price >= 0),
    source TEXT NOT NULL CHECK (source IN ('shop','migration')),
    source_ref TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS blockworld_purchase_once_idx ON blockworld_purchase (user_id, block_id);

CREATE TABLE IF NOT EXISTS blockworld_outbox (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('unlock','restrict')),
    dedupe_key TEXT NOT NULL UNIQUE,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at TIMESTAMPTZ,
    attempts INT NOT NULL DEFAULT 0,
    acked_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS blockworld_outbox_pending_idx ON blockworld_outbox (id) WHERE acked_at IS NULL;

CREATE TABLE IF NOT EXISTS blockworld_restriction_sync (
    source_key TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    state TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS blockworld_world_event (
    event_key TEXT PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('game_won','wallet_migration')),
    user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    status TEXT NOT NULL CHECK (status IN ('recording','done','skipped')),
    payload JSONB NOT NULL DEFAULT '{}',
    result JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS blockworld_migration (
    legacy_identity TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    coins BIGINT NOT NULL CHECK (coins >= 0),
    blocks INT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
DO $$ BEGIN RAISE EXCEPTION 'BlockWorld migration % holds purchases, migrated balances and the bridge outbox and cannot be reverted automatically; restore from backup', '20261113000000-blockworld'; END $$;
