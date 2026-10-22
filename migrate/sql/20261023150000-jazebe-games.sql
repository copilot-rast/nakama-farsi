-- Server-authoritative game rounds, distinct from Nakama login sessions.
-- +migrate Up
CREATE TABLE jazebe_game_template (
    id UUID PRIMARY KEY,
    key VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    engine VARCHAR(32) NOT NULL,
    description TEXT,
    icon VARCHAR(64),
    default_config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    suspended_at TIMESTAMPTZ,
    suspend_reason VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO jazebe_game_template(id,key,name,engine,icon,default_config)
VALUES ('aa63ef7d-07af-5200-81b6-d0c34a9760c1','tap','بازی ضربه','tap','gmico-bolt','{"durationSec":10}'),
       ('0b00b262-320a-5ca2-8315-238273f59c4e','box','جعبه شانس','box','gmico-box','{"boxes":3}')
ON CONFLICT(key) DO NOTHING;
CREATE TABLE jazebe_game (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    template_id UUID REFERENCES jazebe_game_template(id),
    kind VARCHAR(8) NOT NULL DEFAULT 'catalog' CHECK(kind IN ('catalog','custom')),
    name VARCHAR(128) NOT NULL,
    description TEXT,
    icon VARCHAR(64),
    accent VARCHAR(16) NOT NULL DEFAULT 'brand' CHECK(accent IN ('gold','brand','success')),
    entry_model VARCHAR(8) NOT NULL DEFAULT 'free' CHECK(entry_model IN ('free','cost','hybrid')),
    point_cost INTEGER NOT NULL DEFAULT 0 CHECK(point_cost BETWEEN 0 AND 1000000),
    free_individual_count INTEGER CHECK(free_individual_count BETWEEN 1 AND 1000),
    free_individual_period VARCHAR(8) CHECK(free_individual_period IN ('daily','weekly','monthly')),
    free_group_count INTEGER CHECK(free_group_count BETWEEN 1 AND 10000),
    free_group_period VARCHAR(8) CHECK(free_group_period IN ('daily','weekly','monthly')),
    prize_ticket_min INTEGER NOT NULL DEFAULT 0 CHECK(prize_ticket_min BETWEEN 0 AND 10000),
    prize_ticket_max INTEGER NOT NULL DEFAULT 0 CHECK(prize_ticket_max BETWEEN prize_ticket_min AND 10000),
    reward_xp INTEGER NOT NULL DEFAULT 0 CHECK(reward_xp BETWEEN 0 AND 1000000),
    target_lottery_id UUID REFERENCES jazebe_lottery(id),
    config JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(12) NOT NULL DEFAULT 'paused' CHECK(status IN ('active','paused','suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(planet_id,template_id)
);
CREATE INDEX idx_jazebe_game_planet_status ON jazebe_game(planet_id,status);
CREATE TABLE jazebe_game_quota (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    game_id UUID NOT NULL REFERENCES jazebe_game(id),
    scope VARCHAR(8) NOT NULL CHECK(scope IN ('user','group')),
    user_id UUID REFERENCES users(id),
    group_id UUID REFERENCES jazebe_group(id),
    period_key VARCHAR(24) NOT NULL,
    used INTEGER NOT NULL DEFAULT 0 CHECK(used>=0),
    CHECK((scope='user' AND user_id IS NOT NULL AND group_id IS NULL) OR (scope='group' AND group_id IS NOT NULL AND user_id IS NULL))
);
CREATE UNIQUE INDEX uq_jazebe_game_quota_user ON jazebe_game_quota(game_id,user_id,period_key) WHERE scope='user';
CREATE UNIQUE INDEX uq_jazebe_game_quota_group ON jazebe_game_quota(game_id,group_id,period_key) WHERE scope='group';
CREATE TABLE jazebe_game_session (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    game_id UUID NOT NULL REFERENCES jazebe_game(id),
    user_id UUID NOT NULL REFERENCES users(id),
    group_id UUID REFERENCES jazebe_group(id),
    entry_type VARCHAR(16) NOT NULL CHECK(entry_type IN ('free_individual','free_group','paid')),
    point_cost INTEGER NOT NULL DEFAULT 0 CHECK(point_cost>=0),
    point_transaction_id UUID REFERENCES jazebe_point_transaction(id),
    score NUMERIC(5,4) CHECK(score BETWEEN 0 AND 1),
    result JSONB,
    awarded_tickets INTEGER NOT NULL DEFAULT 0 CHECK(awarded_tickets>=0),
    awarded_xp INTEGER NOT NULL DEFAULT 0 CHECK(awarded_xp>=0),
    status VARCHAR(12) NOT NULL DEFAULT 'open' CHECK(status IN ('open','finished','abandoned')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    CHECK((status='open' AND finished_at IS NULL) OR (status<>'open' AND finished_at IS NOT NULL))
);
CREATE UNIQUE INDEX uq_jazebe_game_session_open ON jazebe_game_session(game_id,user_id) WHERE status='open';
CREATE INDEX idx_jazebe_game_session_game_user ON jazebe_game_session(game_id,user_id,started_at DESC);
CREATE INDEX idx_jazebe_game_session_planet_user ON jazebe_game_session(planet_id,user_id,status);
ALTER TABLE jazebe_progression_event DROP CONSTRAINT jazebe_progression_event_kind_check;
ALTER TABLE jazebe_progression_event ADD CONSTRAINT jazebe_progression_event_kind_check
    CHECK(kind IN ('purchase','group_join','mission','game'));
-- +migrate Down
-- Preserve paid rounds, rewards, and qualification history for manual rollback.
SELECT 1;
