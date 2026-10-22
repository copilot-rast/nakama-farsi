-- Source-compatible shop and delivered-order ledgers; purchase support begins with tickets.
-- +migrate Up
CREATE TABLE jazebe_shop_product (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    name VARCHAR(128) NOT NULL,
    short_description VARCHAR(512),
    long_description TEXT,
    image_asset VARCHAR(512),
    type VARCHAR(8) NOT NULL CHECK (type IN ('ticket','coupon','digital','physical')),
    price_points INTEGER NOT NULL CHECK (price_points > 0),
    stock INTEGER CHECK (stock >= 0),
    sold_count INTEGER NOT NULL DEFAULT 0 CHECK (sold_count >= 0),
    terms TEXT,
    delivery_note VARCHAR(512),
    type_config JSONB,
    status VARCHAR(12) NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused','archived','suspended')),
    freshness_rank INTEGER NOT NULL DEFAULT 0,
    price_updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_jazebe_shop_product_planet_status ON jazebe_shop_product(planet_id,status);
CREATE TABLE jazebe_order (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    user_id UUID NOT NULL REFERENCES users(id),
    shop_product_id UUID NOT NULL REFERENCES jazebe_shop_product(id),
    product_name_snapshot VARCHAR(128) NOT NULL,
    product_type_snapshot VARCHAR(8) NOT NULL,
    paid_points INTEGER NOT NULL CHECK (paid_points > 0),
    status VARCHAR(12) NOT NULL CHECK (status IN ('delivered','failed')),
    delivery JSONB,
    point_transaction_id UUID NOT NULL REFERENCES jazebe_point_transaction(id),
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_jazebe_order_user_history ON jazebe_order(planet_id,user_id,purchased_at DESC);
CREATE INDEX idx_jazebe_order_product ON jazebe_order(shop_product_id);

-- +migrate Down
-- Orders and purchase ledgers represent earned value and require manual rollback.
SELECT 1;
