-- +migrate Up
ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS grant_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS grant_retry_at TIMESTAMPTZ;
ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS refund_status TEXT CHECK (refund_status IN ('needed','kept','refunded'));
ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS refund_note TEXT;
ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS refund_by TEXT;
ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS refund_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS economy_payment_reward_pending_idx ON economy_payment (grant_retry_at NULLS FIRST, verified_at) WHERE status = 'success' AND NOT reward_granted;
CREATE INDEX IF NOT EXISTS economy_payment_refund_idx ON economy_payment (refund_status, updated_at DESC) WHERE refund_status IS NOT NULL;

CREATE TABLE IF NOT EXISTS economy_payment_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
DROP TABLE IF EXISTS economy_payment_settings;
DROP INDEX IF EXISTS economy_payment_refund_idx;
DROP INDEX IF EXISTS economy_payment_reward_pending_idx;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS refund_at;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS refund_by;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS refund_note;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS refund_status;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS grant_retry_at;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS grant_attempts;
