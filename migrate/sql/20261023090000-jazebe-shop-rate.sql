-- Per-user purchase throttle shared across Nakama instances. One row per buyer.
-- +migrate Up
CREATE TABLE jazebe_shop_purchase_rate (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    window_started_at TIMESTAMPTZ NOT NULL,
    count INTEGER NOT NULL CHECK (count >= 1)
);

-- +migrate Down
DROP TABLE IF EXISTS jazebe_shop_purchase_rate;
