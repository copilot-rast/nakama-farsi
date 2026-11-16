-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS quiz_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO quiz_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS quiz_mission (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL DEFAULT 'mission' CHECK (kind IN ('mission','game')),
    parent_id UUID REFERENCES quiz_mission (id) ON DELETE SET NULL,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    order_mode TEXT NOT NULL DEFAULT 'all_at_once' CHECK (order_mode IN ('sequential','all_at_once')),
    time_seconds INT NOT NULL DEFAULT 0 CHECK (time_seconds >= 0),
    send_answers BOOLEAN NOT NULL DEFAULT false,
    reward JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS quiz_pack (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    mission_id UUID REFERENCES quiz_mission (id) ON DELETE SET NULL,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    icon_url TEXT NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    category TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    language TEXT NOT NULL DEFAULT '',
    difficulty INT NOT NULL DEFAULT 0,
    time_seconds INT NOT NULL DEFAULT 0 CHECK (time_seconds >= 0),
    question_seconds INT NOT NULL DEFAULT 0 CHECK (question_seconds >= 0),
    pass_percent NUMERIC(5,2) CHECK (pass_percent IS NULL OR (pass_percent >= 0 AND pass_percent <= 100)),
    pass_correct INT CHECK (pass_correct IS NULL OR pass_correct >= 0),
    question_count INT NOT NULL DEFAULT 0 CHECK (question_count >= 0),
    pool_category TEXT NOT NULL DEFAULT '',
    shuffle_questions BOOLEAN,
    shuffle_options BOOLEAN,
    send_answers BOOLEAN NOT NULL DEFAULT false,
    reward JSONB NOT NULL DEFAULT '{}',
    campaign_stage_key TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS quiz_pack_mission_idx ON quiz_pack (mission_id, sort);
CREATE UNIQUE INDEX IF NOT EXISTS quiz_pack_stage_idx ON quiz_pack (campaign_stage_key) WHERE campaign_stage_key <> '';

CREATE TABLE IF NOT EXISTS quiz_question (
    id UUID PRIMARY KEY,
    key TEXT,
    pack_id UUID REFERENCES quiz_pack (id) ON DELETE SET NULL,
    type TEXT NOT NULL DEFAULT 'options' CHECK (type IN ('options','text')),
    question JSONB NOT NULL DEFAULT '{}',
    media_url TEXT NOT NULL DEFAULT '',
    options JSONB NOT NULL DEFAULT '[]',
    answers JSONB NOT NULL DEFAULT '{}',
    explanation JSONB NOT NULL DEFAULT '{}',
    difficulty INT NOT NULL DEFAULT 0,
    category TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    time_seconds INT CHECK (time_seconds IS NULL OR time_seconds >= 0),
    reward JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS quiz_question_key_idx ON quiz_question (key) WHERE key IS NOT NULL;
CREATE INDEX IF NOT EXISTS quiz_question_pack_idx ON quiz_question (pack_id, sort);
CREATE INDEX IF NOT EXISTS quiz_question_category_idx ON quiz_question (category) WHERE category <> '';

CREATE TABLE IF NOT EXISTS quiz_hint (
    id UUID PRIMARY KEY,
    question_id UUID NOT NULL REFERENCES quiz_question (id) ON DELETE CASCADE,
    name JSONB NOT NULL DEFAULT '{}',
    text JSONB NOT NULL DEFAULT '{}',
    media_url TEXT NOT NULL DEFAULT '',
    price JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS quiz_hint_question_idx ON quiz_hint (question_id, sort);

CREATE TABLE IF NOT EXISTS quiz_user_hint (
    user_id UUID NOT NULL,
    hint_id UUID NOT NULL REFERENCES quiz_hint (id) ON DELETE CASCADE,
    scope TEXT NOT NULL DEFAULT '',
    price JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'done' CHECK (status IN ('paying','done')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, hint_id, scope)
);

CREATE TABLE IF NOT EXISTS quiz_session (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    pack_id UUID NOT NULL REFERENCES quiz_pack (id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','paying','done','closed')),
    started_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    deadline_at TIMESTAMPTZ,
    questions JSONB NOT NULL DEFAULT '[]',
    current INT NOT NULL DEFAULT 0,
    served_at TIMESTAMPTZ,
    last_answer_at TIMESTAMPTZ,
    answered INT NOT NULL DEFAULT 0,
    correct INT NOT NULL DEFAULT 0,
    score NUMERIC(6,2),
    passed BOOLEAN,
    campaign_attempt TEXT NOT NULL DEFAULT '',
    result JSONB,
    plan JSONB,
    pay_tries INT NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'
);
CREATE UNIQUE INDEX IF NOT EXISTS quiz_session_open_idx ON quiz_session (user_id, pack_id) WHERE status = 'open';
CREATE INDEX IF NOT EXISTS quiz_session_user_idx ON quiz_session (user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS quiz_session_pack_idx ON quiz_session (pack_id, started_at DESC);
CREATE INDEX IF NOT EXISTS quiz_session_paying_idx ON quiz_session (completed_at) WHERE status = 'paying';
CREATE INDEX IF NOT EXISTS quiz_session_open_exp_idx ON quiz_session (expires_at) WHERE status = 'open';

CREATE TABLE IF NOT EXISTS quiz_answer (
    session_id UUID NOT NULL REFERENCES quiz_session (id) ON DELETE CASCADE,
    question_id UUID NOT NULL,
    user_id UUID NOT NULL,
    option_ids INT[] NOT NULL DEFAULT '{}',
    text TEXT NOT NULL DEFAULT '',
    outcome TEXT NOT NULL CHECK (outcome IN ('correct','wrong','skip','late')),
    seconds NUMERIC(10,2),
    answered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, question_id)
);
CREATE INDEX IF NOT EXISTS quiz_answer_question_idx ON quiz_answer (question_id);

CREATE TABLE IF NOT EXISTS quiz_user_pack (
    user_id UUID NOT NULL,
    pack_id UUID NOT NULL REFERENCES quiz_pack (id) ON DELETE CASCADE,
    plays INT NOT NULL DEFAULT 0,
    wins INT NOT NULL DEFAULT 0,
    best_score NUMERIC(6,2) NOT NULL DEFAULT 0,
    last_score NUMERIC(6,2) NOT NULL DEFAULT 0,
    last_correct INT NOT NULL DEFAULT 0,
    last_total INT NOT NULL DEFAULT 0,
    rewards INT NOT NULL DEFAULT 0,
    first_pass_at TIMESTAMPTZ,
    last_played_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, pack_id)
);
CREATE INDEX IF NOT EXISTS quiz_user_pack_pack_idx ON quiz_user_pack (pack_id);

CREATE TABLE IF NOT EXISTS quiz_user_question (
    user_id UUID NOT NULL,
    question_id UUID NOT NULL REFERENCES quiz_question (id) ON DELETE CASCADE,
    rewarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, question_id)
);

CREATE TABLE IF NOT EXISTS quiz_user_mission (
    user_id UUID NOT NULL,
    mission_id UUID NOT NULL REFERENCES quiz_mission (id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    rewarded BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (user_id, mission_id)
);

CREATE TABLE IF NOT EXISTS quiz_user_day (
    user_id UUID NOT NULL,
    day TEXT NOT NULL,
    plays INT NOT NULL DEFAULT 0,
    rewarded_passes INT NOT NULL DEFAULT 0,
    reports INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, day)
);

CREATE TABLE IF NOT EXISTS quiz_report_reason (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO quiz_report_reason (id, key, name, sort) VALUES
    ('7b2f1e40-3c2a-4d7e-9a51-0f6c1a000001', 'wrong_answer', '{"fa":"پاسخ درست اشتباه است","en":"The correct answer is wrong"}', 10),
    ('7b2f1e40-3c2a-4d7e-9a51-0f6c1a000002', 'typo', '{"fa":"غلط املایی یا نگارشی","en":"Typo or wording"}', 20),
    ('7b2f1e40-3c2a-4d7e-9a51-0f6c1a000003', 'unclear', '{"fa":"سوال مبهم است","en":"The question is unclear"}', 30),
    ('7b2f1e40-3c2a-4d7e-9a51-0f6c1a000004', 'media', '{"fa":"تصویر یا رسانه مشکل دارد","en":"Broken image or media"}', 40),
    ('7b2f1e40-3c2a-4d7e-9a51-0f6c1a000005', 'offensive', '{"fa":"محتوای نامناسب","en":"Offensive content"}', 50),
    ('7b2f1e40-3c2a-4d7e-9a51-0f6c1a000006', 'other', '{"fa":"موارد دیگر","en":"Other"}', 90)
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS quiz_question_report (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    question_id UUID NOT NULL REFERENCES quiz_question (id) ON DELETE CASCADE,
    reason_key TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    session_id UUID,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','resolved','dismissed')),
    resolution TEXT NOT NULL DEFAULT '',
    reviewed_by TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS quiz_question_report_open_idx ON quiz_question_report (user_id, question_id) WHERE status = 'open';
CREATE INDEX IF NOT EXISTS quiz_question_report_status_idx ON quiz_question_report (status, created_at DESC);
CREATE INDEX IF NOT EXISTS quiz_question_report_question_idx ON quiz_question_report (question_id);

CREATE TABLE IF NOT EXISTS quiz_result_template (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    mission_id UUID REFERENCES quiz_mission (id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    won_text TEXT NOT NULL DEFAULT 'won',
    lost_text TEXT NOT NULL DEFAULT 'lost',
    webhook_url TEXT NOT NULL DEFAULT '',
    webhook_secret TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS quiz_result_delivery (
    id UUID PRIMARY KEY,
    template_id UUID NOT NULL REFERENCES quiz_result_template (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    url TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','failed')),
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS quiz_result_delivery_due_idx ON quiz_result_delivery (next_attempt_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS quiz_result_delivery_template_idx ON quiz_result_delivery (template_id, created_at DESC);

-- +migrate StatementBegin
DO $$
DECLARE
    def TEXT;
    vals TEXT[];
BEGIN
    SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
        WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
    IF def IS NULL OR def LIKE '%''quiz''%' THEN
        RETURN;
    END IF;
    SELECT array_agg(DISTINCT m[1]) INTO vals FROM regexp_matches(def, '''([a-z_]+)''', 'g') AS m;
    vals := array_append(vals, 'quiz');
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type = ANY (' || quote_literal(vals::text) || '::text[]))';
END $$;
-- +migrate StatementEnd

-- +migrate Down
DROP TABLE IF EXISTS quiz_result_delivery;
DROP TABLE IF EXISTS quiz_result_template;
DROP TABLE IF EXISTS quiz_question_report;
DROP TABLE IF EXISTS quiz_report_reason;
DROP TABLE IF EXISTS quiz_user_day;
DROP TABLE IF EXISTS quiz_user_mission;
DROP TABLE IF EXISTS quiz_user_question;
DROP TABLE IF EXISTS quiz_user_pack;
DROP TABLE IF EXISTS quiz_answer;
DROP TABLE IF EXISTS quiz_session;
DROP TABLE IF EXISTS quiz_user_hint;
DROP TABLE IF EXISTS quiz_hint;
DROP TABLE IF EXISTS quiz_question;
DROP TABLE IF EXISTS quiz_pack;
DROP TABLE IF EXISTS quiz_mission;
DROP TABLE IF EXISTS quiz_settings;
