-- Platform medal catalogue, planet selections and append-only acquisitions.
-- +migrate Up
CREATE TABLE jazebe_medal_library (
    id UUID PRIMARY KEY,
    key VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    icon VARCHAR(64),
    theme_category VARCHAR(16) NOT NULL CHECK (theme_category IN ('participation','purchase','social','progression')),
    condition_type VARCHAR(16) NOT NULL CHECK (condition_type IN ('activity_count','duration','special_event')),
    condition_param_schema JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE jazebe_planet_medal (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    medal_library_id UUID NOT NULL REFERENCES jazebe_medal_library(id),
    condition_params JSONB NOT NULL DEFAULT '{}',
    reward_points INTEGER NOT NULL DEFAULT 0 CHECK (reward_points >= 0),
    reward_tickets INTEGER NOT NULL DEFAULT 0 CHECK (reward_tickets >= 0),
    status VARCHAR(8) NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (planet_id,medal_library_id),
    UNIQUE (id,planet_id)
);
CREATE INDEX idx_jazebe_planet_medal_active ON jazebe_planet_medal(planet_id) WHERE status='active';
CREATE INDEX idx_jazebe_planet_medal_library ON jazebe_planet_medal(medal_library_id);
CREATE TABLE jazebe_medal_acquisition (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    planet_medal_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reward_point_transaction_id UUID REFERENCES jazebe_point_transaction(id),
    reward_ticket_ids UUID[],
    UNIQUE (planet_medal_id,user_id),
    FOREIGN KEY (planet_medal_id,planet_id) REFERENCES jazebe_planet_medal(id,planet_id)
);
CREATE INDEX idx_jazebe_medal_acquisition_user ON jazebe_medal_acquisition(planet_id,user_id,acquired_at DESC);

-- +migrate Down
-- Preserve earned medals and their rewards for manual rollback review.
SELECT 1;
