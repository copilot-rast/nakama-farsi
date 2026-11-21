-- +migrate Up
CREATE TABLE IF NOT EXISTS base_username_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_username_state (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    free_used BOOLEAN NOT NULL DEFAULT false,
    paid_count INT NOT NULL DEFAULT 0,
    last_change_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_username_change (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    old_username TEXT NOT NULL,
    new_username TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('free','paid','admin')),
    cost BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT '',
    admin TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS base_username_change_user_idx ON base_username_change (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS base_player_code_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_player_code (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION base_new_player_code() RETURNS TEXT AS $$
DECLARE
    n INT;
    alphabet TEXT;
    result TEXT;
    i INT;
BEGIN
    SELECT COALESCE((value->>'length')::INT, 6), COALESCE(NULLIF(value->>'alphabet', ''), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789')
        INTO n, alphabet FROM base_player_code_settings WHERE id = 1;
    IF n IS NULL THEN
        n := 6;
        alphabet := 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    END IF;
    IF n < 4 OR n > 16 THEN
        n := 6;
    END IF;
    FOR attempt IN 1..20 LOOP
        result := '';
        FOR i IN 1..n LOOP
            result := result || substr(alphabet, 1 + floor(random() * length(alphabet))::INT, 1);
        END LOOP;
        IF NOT EXISTS (SELECT 1 FROM base_player_code WHERE code = result) THEN
            RETURN result;
        END IF;
        IF attempt % 5 = 0 THEN
            n := n + 1;
        END IF;
    END LOOP;
    RAISE EXCEPTION 'no free player code';
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION base_player_code_assign() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.id <> '00000000-0000-0000-0000-000000000000'::UUID THEN
        FOR attempt IN 1..5 LOOP
            BEGIN
                INSERT INTO base_player_code (user_id, code) VALUES (NEW.id, base_new_player_code()) ON CONFLICT (user_id) DO NOTHING;
                EXIT;
            EXCEPTION WHEN unique_violation THEN
                IF attempt = 5 THEN
                    RAISE;
                END IF;
            END;
        END LOOP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS base_player_code_assign ON users;
CREATE TRIGGER base_player_code_assign AFTER INSERT ON users FOR EACH ROW EXECUTE FUNCTION base_player_code_assign();

-- +migrate StatementBegin
DO $$
DECLARE
    u RECORD;
BEGIN
    FOR u IN SELECT id FROM users WHERE id <> '00000000-0000-0000-0000-000000000000'::UUID AND NOT EXISTS (SELECT 1 FROM base_player_code c WHERE c.user_id = users.id) LOOP
        INSERT INTO base_player_code (user_id, code) VALUES (u.id, base_new_player_code()) ON CONFLICT DO NOTHING;
    END LOOP;
END $$;
-- +migrate StatementEnd

ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS notify_declined BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS notify_removed BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS presence_friends_only BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS dm_friends_only BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS base_friend_suggest_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_coplay (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    other_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    last_match_id TEXT NOT NULL,
    last_game_key TEXT NOT NULL DEFAULT '',
    last_at TIMESTAMPTZ NOT NULL,
    matches INT NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, other_id)
);
CREATE INDEX IF NOT EXISTS base_coplay_recent_idx ON base_coplay (user_id, last_at DESC);

CREATE TABLE IF NOT EXISTS base_coplay_cursor (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    ended_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    match_id TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS economy_game_match_finished_idx ON economy_game_match (ended_at, id) WHERE status = 'finished';
INSERT INTO base_coplay_cursor (id, ended_at) VALUES (1, now() - INTERVAL '30 days') ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS base_presence_day (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    day DATE NOT NULL,
    PRIMARY KEY (user_id, day)
);
CREATE INDEX IF NOT EXISTS base_presence_day_day_idx ON base_presence_day (day);

CREATE TABLE IF NOT EXISTS base_friend_suggest_list (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    cycle TEXT NOT NULL,
    impression_cycle TEXT NOT NULL DEFAULT '',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    generated_by TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS base_friend_suggest (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    candidate_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    slot INT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('coplay','mutual','activity','level')),
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    mutual INT NOT NULL DEFAULT 0,
    last_played_at TIMESTAMPTZ,
    shows INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, candidate_id)
);
CREATE INDEX IF NOT EXISTS base_friend_suggest_slot_idx ON base_friend_suggest (user_id, slot);

CREATE TABLE IF NOT EXISTS base_friend_suggest_skip (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    candidate_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    until_cycle TEXT NOT NULL,
    PRIMARY KEY (user_id, candidate_id)
);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds username history, player codes and friend suggestions and cannot be reverted automatically; restore from backup', '20270106000000-base-social-extras'; END $$;
-- +migrate StatementEnd
