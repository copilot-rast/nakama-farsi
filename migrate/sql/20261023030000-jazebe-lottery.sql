-- A planet lottery and its append-only group-owned ticket pool.
-- +migrate Up
CREATE TABLE jazebe_lottery (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    title VARCHAR(128) NOT NULL,
    short_description VARCHAR(512),
    long_description TEXT,
    banner_asset VARCHAR(512),
    starts_at TIMESTAMPTZ NOT NULL,
    draw_at TIMESTAMPTZ NOT NULL,
    ticket_price_points INTEGER NOT NULL CHECK (ticket_price_points > 0),
    per_user_ticket_cap INTEGER CHECK (per_user_ticket_cap > 0),
    per_group_ticket_cap INTEGER CHECK (per_group_ticket_cap > 0),
    status VARCHAR(12) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','drawing','finished','cancelled')),
    total_ticket_count INTEGER NOT NULL DEFAULT 0 CHECK (total_ticket_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT jazebe_lottery_time CHECK (draw_at > starts_at),
    CONSTRAINT jazebe_lottery_caps CHECK (per_user_ticket_cap IS NULL OR per_group_ticket_cap IS NULL OR per_user_ticket_cap <= per_group_ticket_cap),
    CONSTRAINT jazebe_lottery_planet_id_unique UNIQUE (planet_id,id)
);
CREATE INDEX idx_jazebe_lottery_planet_status ON jazebe_lottery(planet_id,status);
CREATE INDEX idx_jazebe_lottery_draw ON jazebe_lottery(status,draw_at);
CREATE TABLE jazebe_lottery_prize (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL,
    lottery_id UUID NOT NULL,
    title VARCHAR(128) NOT NULL,
    description TEXT,
    type VARCHAR(16) NOT NULL CHECK (type IN ('cash','brand_credit','planet_points','discount_coupon')),
    value_rial BIGINT NOT NULL CHECK (value_rial >= 0),
    priority SMALLINT NOT NULL CHECK (priority BETWEEN 1 AND 50),
    image_asset VARCHAR(512),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT jazebe_lottery_prize_lottery_fk FOREIGN KEY (planet_id,lottery_id) REFERENCES jazebe_lottery(planet_id,id),
    CONSTRAINT jazebe_lottery_prize_slot UNIQUE (lottery_id,priority)
);
CREATE TABLE jazebe_lottery_group_stat (
    lottery_id UUID NOT NULL REFERENCES jazebe_lottery(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    ticket_count INTEGER NOT NULL DEFAULT 0 CHECK (ticket_count >= 0),
    individual_ticket_count INTEGER NOT NULL DEFAULT 0 CHECK (individual_ticket_count >= 0),
    collective_ticket_count INTEGER NOT NULL DEFAULT 0 CHECK (collective_ticket_count >= 0),
    rank INTEGER,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (lottery_id,group_id),
    CONSTRAINT jazebe_lottery_group_stat_sum CHECK (ticket_count = individual_ticket_count + collective_ticket_count)
);
CREATE TABLE jazebe_lottery_ticket (
    id UUID PRIMARY KEY,
    lottery_id UUID NOT NULL REFERENCES jazebe_lottery(id),
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    creator_type VARCHAR(8) NOT NULL CHECK (creator_type IN ('member','group')),
    creator_user_id UUID REFERENCES users(id),
    source VARCHAR(24) NOT NULL CHECK (source IN ('purchase','solo_mission','group_mission','game','previous_lottery_prize','medal')),
    source_id UUID,
    serial VARCHAR(24) NOT NULL UNIQUE,
    status VARCHAR(8) NOT NULL DEFAULT 'active' CHECK (status IN ('active','void')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT jazebe_ticket_creator CHECK ((creator_type='member') = (creator_user_id IS NOT NULL))
);
CREATE INDEX idx_jazebe_ticket_lottery_group ON jazebe_lottery_ticket(lottery_id,group_id);
CREATE INDEX idx_jazebe_ticket_lottery_creator ON jazebe_lottery_ticket(lottery_id,creator_user_id) WHERE status='active';
CREATE INDEX idx_jazebe_ticket_source ON jazebe_lottery_ticket(source,source_id);

-- +migrate Down
-- Tickets represent earned value and must be retained for manual rollback review.
SELECT 1;
