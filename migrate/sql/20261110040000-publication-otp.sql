-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS publication_otp_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    sandbox BOOLEAN NOT NULL DEFAULT true,
    provider_id UUID REFERENCES publication_provider (id) ON DELETE SET NULL,
    code_length INT NOT NULL DEFAULT 5 CHECK (code_length BETWEEN 4 AND 8),
    ttl_seconds INT NOT NULL DEFAULT 120 CHECK (ttl_seconds BETWEEN 30 AND 1800),
    max_attempts INT NOT NULL DEFAULT 5 CHECK (max_attempts BETWEEN 1 AND 10),
    resend_seconds INT NOT NULL DEFAULT 60 CHECK (resend_seconds BETWEEN 10 AND 600),
    phone_hourly INT NOT NULL DEFAULT 5 CHECK (phone_hourly BETWEEN 1 AND 100),
    phone_daily INT NOT NULL DEFAULT 10 CHECK (phone_daily BETWEEN 1 AND 500),
    ip_hourly INT NOT NULL DEFAULT 30 CHECK (ip_hourly BETWEEN 1 AND 10000),
    message JSONB NOT NULL DEFAULT '{"fa":"کد ورود شما: {code}\nاین کد تا {minutes} دقیقه معتبر است.","en":"Your login code: {code}\nIt is valid for {minutes} minutes."}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO publication_otp_settings (id) VALUES (1) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS publication_otp_code (
    id UUID PRIMARY KEY,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','verified','expired','failed','replaced','send_failed')),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL,
    sandbox BOOLEAN NOT NULL DEFAULT true,
    provider TEXT NOT NULL DEFAULT '',
    lang TEXT NOT NULL DEFAULT '',
    error TEXT NOT NULL DEFAULT '',
    client_ip TEXT NOT NULL DEFAULT '',
    requested_by UUID,
    user_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS publication_otp_code_phone_idx ON publication_otp_code (phone, created_at DESC);
CREATE INDEX IF NOT EXISTS publication_otp_code_created_idx ON publication_otp_code (created_at DESC);

CREATE TABLE IF NOT EXISTS publication_phone_account (
    phone TEXT PRIMARY KEY CHECK (phone ~ '^\+[1-9][0-9]{7,14}$'),
    user_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    login_key TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    verified_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % holds phone accounts and cannot be reverted automatically; restore from backup', '20261110040000-publication-otp'; END $$;
-- +migrate StatementEnd
