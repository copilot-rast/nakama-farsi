-- +migrate Up
CREATE TABLE IF NOT EXISTS quiz_category (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    duel BOOLEAN NOT NULL DEFAULT true,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS quiz_duel (
    id UUID PRIMARY KEY,
    mode TEXT NOT NULL CHECK (mode IN ('random','friend')),
    status TEXT NOT NULL CHECK (status IN ('waiting','invited','active','finished','expired','declined','cancelled')),
    player_a UUID NOT NULL,
    player_b UUID,
    rules JSONB NOT NULL DEFAULT '{}',
    turn_user UUID,
    turn_started_at TIMESTAMPTZ,
    turn_deadline TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    score_a INT NOT NULL DEFAULT 0,
    score_b INT NOT NULL DEFAULT 0,
    winner UUID,
    reason TEXT NOT NULL DEFAULT '',
    fee_a TEXT NOT NULL DEFAULT '' CHECK (fee_a IN ('','paid','refunded')),
    fee_b TEXT NOT NULL DEFAULT '' CHECK (fee_b IN ('','paid','refunded')),
    paid_a BOOLEAN NOT NULL DEFAULT false,
    paid_b BOOLEAN NOT NULL DEFAULT false,
    pay_tries INT NOT NULL DEFAULT 0,
    pay_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (player_b IS NULL OR player_b <> player_a)
);
CREATE INDEX IF NOT EXISTS quiz_duel_a_idx ON quiz_duel (player_a, created_at DESC);
CREATE INDEX IF NOT EXISTS quiz_duel_b_idx ON quiz_duel (player_b, created_at DESC) WHERE player_b IS NOT NULL;
CREATE INDEX IF NOT EXISTS quiz_duel_waiting_idx ON quiz_duel (created_at) WHERE status = 'waiting';
CREATE INDEX IF NOT EXISTS quiz_duel_turn_idx ON quiz_duel (turn_deadline) WHERE status IN ('active','invited');
CREATE INDEX IF NOT EXISTS quiz_duel_expiry_idx ON quiz_duel (expires_at) WHERE status IN ('waiting','invited');
CREATE INDEX IF NOT EXISTS quiz_duel_unpaid_idx ON quiz_duel (finished_at) WHERE status = 'finished' AND (NOT paid_a OR NOT paid_b);

CREATE TABLE IF NOT EXISTS quiz_duel_round (
    duel_id UUID NOT NULL REFERENCES quiz_duel (id) ON DELETE CASCADE,
    round_no INT NOT NULL,
    picker UUID NOT NULL,
    offers TEXT[] NOT NULL DEFAULT '{}',
    category TEXT NOT NULL DEFAULT '',
    questions JSONB NOT NULL DEFAULT '[]',
    picked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (duel_id, round_no)
);

CREATE TABLE IF NOT EXISTS quiz_duel_answer (
    duel_id UUID NOT NULL REFERENCES quiz_duel (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    question_id UUID NOT NULL,
    round_no INT NOT NULL,
    idx INT NOT NULL,
    served_at TIMESTAMPTZ NOT NULL,
    answered_at TIMESTAMPTZ,
    option_id INT,
    outcome TEXT NOT NULL DEFAULT 'pending' CHECK (outcome IN ('pending','correct','wrong','late','skip')),
    seconds NUMERIC(8,2),
    PRIMARY KEY (duel_id, user_id, question_id)
);
CREATE INDEX IF NOT EXISTS quiz_duel_answer_user_idx ON quiz_duel_answer (user_id, served_at DESC);

CREATE TABLE IF NOT EXISTS quiz_duel_day (
    user_id UUID NOT NULL,
    day TEXT NOT NULL,
    starts INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, day)
);

-- +migrate Down
DROP TABLE IF EXISTS quiz_duel_day;
DROP TABLE IF EXISTS quiz_duel_answer;
DROP TABLE IF EXISTS quiz_duel_round;
DROP TABLE IF EXISTS quiz_duel;
DROP TABLE IF EXISTS quiz_category;
