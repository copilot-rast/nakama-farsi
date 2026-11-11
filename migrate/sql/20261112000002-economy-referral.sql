-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS economy_referral_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    new_account_days INT NOT NULL DEFAULT 7 CHECK (new_account_days BETWEEN 0 AND 3650),
    referrer_prize_id UUID REFERENCES economy_prize (id) ON DELETE SET NULL,
    referee_prize_id UUID REFERENCES economy_prize (id) ON DELETE SET NULL,
    goal_event_type_id UUID REFERENCES economy_event_type (id) ON DELETE SET NULL,
    goal_measure TEXT NOT NULL DEFAULT 'count' CHECK (goal_measure IN ('count','sum','max')),
    goal_value NUMERIC(20,4) NOT NULL DEFAULT 1 CHECK (goal_value > 0),
    max_per_referrer INT NOT NULL DEFAULT 100 CHECK (max_per_referrer BETWEEN 1 AND 100000),
    daily_per_referrer INT NOT NULL DEFAULT 10 CHECK (daily_per_referrer BETWEEN 1 AND 100000),
    one_per_device BOOLEAN NOT NULL DEFAULT true,
    one_per_phone BOOLEAN NOT NULL DEFAULT true,
    hold_flagged BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO economy_referral_settings (id) VALUES (1) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS economy_referral_code (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS economy_referral (
    id UUID PRIMARY KEY,
    referrer_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    referee_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','held','completed','rejected')),
    flags TEXT[] NOT NULL DEFAULT '{}',
    referrer_prize_id UUID REFERENCES economy_prize (id) ON DELETE SET NULL,
    referee_prize_id UUID REFERENCES economy_prize (id) ON DELETE SET NULL,
    goal_event_type_id UUID REFERENCES economy_event_type (id) ON DELETE SET NULL,
    goal_measure TEXT NOT NULL DEFAULT 'count',
    goal_value NUMERIC(20,4),
    ip TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    decided_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    decided_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS economy_referral_referrer_idx ON economy_referral (referrer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS economy_referral_status_idx ON economy_referral (status, created_at DESC);

CREATE TABLE IF NOT EXISTS economy_referral_mark (
    kind TEXT NOT NULL CHECK (kind IN ('device','phone')),
    hash TEXT NOT NULL,
    referral_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (kind, hash)
);

ALTER TABLE economy_reward_claim DROP CONSTRAINT IF EXISTS economy_reward_claim_kind_check;
ALTER TABLE economy_reward_claim ADD CONSTRAINT economy_reward_claim_kind_check CHECK (kind IN ('event','level','streak','referral'));
ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak','referral'));

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds referrals and their rewards and cannot be reverted automatically; restore from backup', '20261112000002-economy-referral'; END $$;
-- +migrate StatementEnd
