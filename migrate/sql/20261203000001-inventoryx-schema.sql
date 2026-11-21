-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS inventoryx_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO inventoryx_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- Extra rules of an economy item: the price to open its level lock early and its upgrade tracks.
CREATE TABLE IF NOT EXISTS inventoryx_item (
    item_id UUID PRIMARY KEY REFERENCES economy_item (id) ON DELETE CASCADE,
    unlock_prices JSONB NOT NULL DEFAULT '[]',
    tracks JSONB NOT NULL DEFAULT '[]',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Items a player opened before reaching their level. No foreign key to users: the
-- wallet update locks the user row from another connection (see economy_user_item).
CREATE TABLE IF NOT EXISTS inventoryx_user_unlock (
    user_id UUID NOT NULL,
    item_id UUID NOT NULL REFERENCES economy_item (id) ON DELETE CASCADE,
    paid JSONB NOT NULL DEFAULT '[]',
    grant_id TEXT NOT NULL DEFAULT '',
    actor TEXT NOT NULL DEFAULT 'player',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, item_id)
);

-- Level of each upgrade track (Power, MagazineCapacity …) of an item a player owns.
CREATE TABLE IF NOT EXISTS inventoryx_user_track (
    user_id UUID NOT NULL,
    item_id UUID NOT NULL REFERENCES economy_item (id) ON DELETE CASCADE,
    track TEXT NOT NULL,
    level INT NOT NULL DEFAULT 0 CHECK (level >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, item_id, track)
);

-- Answers kept per requestId so a retried unlock or upgrade never pays twice.
CREATE TABLE IF NOT EXISTS inventoryx_request (
    user_id UUID NOT NULL,
    request_id TEXT NOT NULL,
    action TEXT NOT NULL,
    item_key TEXT NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    result JSONB NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, request_id)
);

CREATE INDEX IF NOT EXISTS inventoryx_request_expires_idx ON inventoryx_request (expires_at);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Inventoryx migration % holds paid unlocks and upgrade levels and cannot be reverted automatically; restore from backup', '20261203000001-inventoryx-schema'; END $$;
-- +migrate StatementEnd
