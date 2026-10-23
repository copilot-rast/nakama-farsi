-- Platform-wide bounds used by medal, mission, group and planet settings.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_platform_policy (
    id UUID PRIMARY KEY,
    key VARCHAR(16) NOT NULL UNIQUE CHECK (key='default'),
    mission_reward_point_min INTEGER NOT NULL DEFAULT 0 CHECK (mission_reward_point_min>=0),
    mission_reward_point_max INTEGER NOT NULL DEFAULT 1000 CHECK (mission_reward_point_max>=0),
    mission_reward_ticket_max INTEGER NOT NULL DEFAULT 10 CHECK (mission_reward_ticket_max>=0),
    mission_reward_xp_max INTEGER NOT NULL DEFAULT 1000 CHECK (mission_reward_xp_max>=0),
    ticket_price_min INTEGER NOT NULL DEFAULT 1 CHECK (ticket_price_min>=0),
    ticket_price_max INTEGER NOT NULL DEFAULT 100000 CHECK (ticket_price_max>=0),
    lottery_fee_percent_min NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (lottery_fee_percent_min BETWEEN 0 AND 100),
    lottery_fee_percent_max NUMERIC(5,2) NOT NULL DEFAULT 100 CHECK (lottery_fee_percent_max BETWEEN 0 AND 100),
    group_member_cap_min INTEGER NOT NULL DEFAULT 2 CHECK (group_member_cap_min>=1),
    group_member_cap_max INTEGER NOT NULL DEFAULT 100 CHECK (group_member_cap_max>=1),
    group_admin_cap_max INTEGER NOT NULL DEFAULT 5 CHECK (group_admin_cap_max>=1),
    group_max_count_max INTEGER CHECK (group_max_count_max>=1),
    otp_length SMALLINT NOT NULL DEFAULT 6 CHECK (otp_length BETWEEN 4 AND 8),
    otp_ttl_seconds INTEGER NOT NULL DEFAULT 120 CHECK (otp_ttl_seconds>=30),
    otp_max_attempts SMALLINT NOT NULL DEFAULT 5 CHECK (otp_max_attempts>=1),
    otp_resend_cooldown_seconds INTEGER NOT NULL DEFAULT 60 CHECK (otp_resend_cooldown_seconds>=0),
    rate_limit_per_minute INTEGER NOT NULL DEFAULT 120 CHECK (rate_limit_per_minute>=1),
    display_name_change_limit_per_month INTEGER NOT NULL DEFAULT 3 CHECK (display_name_change_limit_per_month>=0),
    terms_version VARCHAR(32) NOT NULL DEFAULT '1',
    extra JSONB,
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (mission_reward_point_min<=mission_reward_point_max),
    CHECK (ticket_price_min<=ticket_price_max),
    CHECK (lottery_fee_percent_min<=lottery_fee_percent_max),
    CHECK (group_member_cap_min<=group_member_cap_max)
);
INSERT INTO jazebe_platform_policy(id,key) VALUES('6a7a11eb-0000-4000-8000-000000000001','default') ON CONFLICT(key) DO NOTHING;

-- +migrate Down
-- Retain operator policy and audit history for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023120000-jazebe-platform-policy'; END $$;
-- +migrate StatementEnd
