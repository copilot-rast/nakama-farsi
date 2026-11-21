-- +migrate Up
ALTER TABLE economy_chance_game DROP CONSTRAINT IF EXISTS economy_chance_game_prize_type_check;
ALTER TABLE economy_chance_game ADD CONSTRAINT economy_chance_game_prize_type_check CHECK (prize_type IN ('prize','prizeGroup','slots'));
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS price_ladder JSONB;
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS pity_threshold INT CHECK (pity_threshold IS NULL OR pity_threshold BETWEEN 1 AND 100000);
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS reveal_ack BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS reroll_empty BOOLEAN NOT NULL DEFAULT true;

ALTER TABLE economy_chance_state ADD COLUMN IF NOT EXISTS since_jackpot INT NOT NULL DEFAULT 0;

ALTER TABLE economy_chance_play ADD COLUMN IF NOT EXISTS outcome JSONB;
ALTER TABLE economy_chance_play ADD COLUMN IF NOT EXISTS forced BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE economy_chance_play ADD COLUMN IF NOT EXISTS acked_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS economy_chance_play_reveal_idx ON economy_chance_play (game_id, user_id, created_at DESC) WHERE acked_at IS NULL AND status = 'done';

-- +migrate StatementBegin
DO $$
DECLARE
    def TEXT;
    vals TEXT[];
BEGIN
    SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
        WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
    IF def IS NULL OR def ~ ('[{,'']chance_slot[},'']') THEN
        RETURN;
    END IF;
    SELECT array_agg(DISTINCT m[1]) INTO vals FROM regexp_matches(def, '''([a-z_]+)''', 'g') AS m;
    IF vals IS NULL THEN
        vals := string_to_array(substring(def from '''\{([^}]*)\}'''), ',');
    END IF;
    vals := array_append(vals, 'chance_slot');
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type = ANY (' || quote_literal(vals::text) || '::text[]))';
END $$;
-- +migrate StatementEnd

CREATE TABLE IF NOT EXISTS economy_booster (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    meta_id UUID NOT NULL REFERENCES economy_meta (id) ON DELETE CASCADE,
    multiplier NUMERIC(8,2) NOT NULL CHECK (multiplier > 0),
    expires_at TIMESTAMPTZ NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, meta_id)
);
CREATE INDEX IF NOT EXISTS economy_booster_expires_idx ON economy_booster (expires_at);

ALTER TABLE economy_streak DROP CONSTRAINT IF EXISTS economy_streak_miss_action_check;
ALTER TABLE economy_streak ADD CONSTRAINT economy_streak_miss_action_check CHECK (miss_action IN ('reset','step_back','week_start'));
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS week_length INT NOT NULL DEFAULT 7 CHECK (week_length BETWEEN 1 AND 365);
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS infinity BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS infinity_currency TEXT NOT NULL DEFAULT '';
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS infinity_min BIGINT NOT NULL DEFAULT 0;
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS infinity_max BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS economy_streak_extra (
    id UUID PRIMARY KEY,
    streak_id UUID NOT NULL REFERENCES economy_streak (id) ON DELETE CASCADE,
    day INT NOT NULL CHECK (day BETWEEN 1 AND 3650),
    prize_id UUID NOT NULL REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL DEFAULT 'claim' CHECK (grant_mode IN ('auto','claim')),
    UNIQUE (streak_id, day)
);
CREATE INDEX IF NOT EXISTS economy_streak_extra_prize_idx ON economy_streak_extra (prize_id);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds chest outcomes, wheel pity, boosters and streak extras and cannot be reverted automatically; restore from backup', '20270108000000-economy-chance-boosters-streaks'; END $$;
-- +migrate StatementEnd
