-- +migrate Up notransaction
ALTER TABLE economy_reward_claim DROP CONSTRAINT IF EXISTS economy_reward_claim_kind_check;
ALTER TABLE economy_reward_claim ADD CONSTRAINT economy_reward_claim_kind_check CHECK (kind IN ('event','level','streak','referral','task','achievement','chance'));
ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak','referral','task','achievement','chance'));

CREATE TABLE IF NOT EXISTS economy_prize_group (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_prize_group_member (
    group_id UUID NOT NULL REFERENCES economy_prize_group (id) ON DELETE CASCADE,
    prize_id UUID NOT NULL REFERENCES economy_prize (id),
    weight INT NOT NULL DEFAULT 1 CHECK (weight BETWEEN 1 AND 1000000),
    sort INT NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, prize_id)
);
CREATE INDEX IF NOT EXISTS economy_prize_group_member_prize_idx ON economy_prize_group_member (prize_id);

CREATE TABLE IF NOT EXISTS economy_chance_game (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}',
    prize_type TEXT NOT NULL DEFAULT 'prize' CHECK (prize_type IN ('prize','prizeGroup')),
    prize JSONB NOT NULL DEFAULT '[]',
    interval_time INT CHECK (interval_time IS NULL OR interval_time BETWEEN 1 AND 525600),
    price JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_chance_state (
    game_id UUID NOT NULL REFERENCES economy_chance_game (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    plays BIGINT NOT NULL DEFAULT 0,
    last_play_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (game_id, user_id)
);
CREATE INDEX IF NOT EXISTS economy_chance_state_user_idx ON economy_chance_state (user_id);

CREATE TABLE IF NOT EXISTS economy_chance_play (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES economy_chance_game (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('free','ticket')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','done')),
    roll INT NOT NULL CHECK (roll BETWEEN 0 AND 999999),
    slot INT,
    prize_id UUID REFERENCES economy_prize (id),
    prize_group_id UUID REFERENCES economy_prize_group (id),
    claim_id UUID,
    price JSONB,
    error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    attempted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    event_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS economy_chance_play_turn_idx ON economy_chance_play (game_id, user_id, mode, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_chance_play_user_idx ON economy_chance_play (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_chance_play_game_idx ON economy_chance_play (game_id, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_chance_play_pending_idx ON economy_chance_play (created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS economy_chance_play_prize_idx ON economy_chance_play (prize_id) WHERE prize_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS economy_chance_play_group_idx ON economy_chance_play (prize_group_id) WHERE prize_group_id IS NOT NULL;

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'ChanceGame:Played', '{"fa":"بازی شانس","en":"Chance game played","ar":"لعب لعبة الحظ"}', '{"fa":"هر بار که بازیکن یک بازی شانس را بازی می‌کند، رایگان یا با بلیت. شناسهٔ بازی در entityId می‌آید.","en":"Recorded each time a player plays a chance game, free or with a ticket. The game id is in entityId.","ar":"يُسجَّل في كل مرة يلعب فيها اللاعب لعبة حظ، مجانًا أو بتذكرة. معرّف اللعبة في entityId."}', false, 0, 120, true, 11)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds chance games, prize groups and the play history and cannot be reverted automatically; restore from backup', '20261118000000-economy-chance-games'; END $$;
-- +migrate StatementEnd
