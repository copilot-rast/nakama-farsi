-- +migrate Up
ALTER TABLE economy_payment_method DROP CONSTRAINT IF EXISTS economy_payment_method_method_check;
ALTER TABLE economy_payment_method ADD CONSTRAINT economy_payment_method_method_check CHECK (method IN ('zarinpal','zibal','bazaar','myket','in_app','google_play','apple','stripe'));
INSERT INTO economy_payment_method (method) VALUES ('google_play'), ('apple'), ('stripe') ON CONFLICT (method) DO NOTHING;

ALTER TABLE economy_payment_log DROP CONSTRAINT IF EXISTS economy_payment_log_action_check;
ALTER TABLE economy_payment_log ADD CONSTRAINT economy_payment_log_action_check CHECK (action IN ('request','callback','verify','inquiry','store_validate','webhook'));

ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS price_minor BIGINT CHECK (price_minor IS NULL OR price_minor > 0);
ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS price_currency TEXT CHECK (price_currency IS NULL OR price_currency ~ '^[a-z]{3}$');

ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS currency TEXT;
ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS amount_minor BIGINT;
CREATE UNIQUE INDEX IF NOT EXISTS economy_payment_store_ref_idx ON economy_payment (method, ref_id) WHERE ref_id IS NOT NULL AND status = 'success' AND method IN ('google_play','apple');

CREATE TABLE IF NOT EXISTS economy_payment_webhook (
    method TEXT NOT NULL,
    event_id TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT '',
    payment_id UUID,
    result TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (method, event_id)
);
CREATE INDEX IF NOT EXISTS economy_payment_webhook_payment_idx ON economy_payment_webhook (payment_id) WHERE payment_id IS NOT NULL;

-- +migrate Down
DROP TABLE IF EXISTS economy_payment_webhook;
DROP INDEX IF EXISTS economy_payment_store_ref_idx;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS amount_minor;
ALTER TABLE economy_payment DROP COLUMN IF EXISTS currency;
ALTER TABLE economy_product DROP COLUMN IF EXISTS price_currency;
ALTER TABLE economy_product DROP COLUMN IF EXISTS price_minor;
DELETE FROM economy_payment_log WHERE action = 'webhook';
ALTER TABLE economy_payment_log DROP CONSTRAINT IF EXISTS economy_payment_log_action_check;
ALTER TABLE economy_payment_log ADD CONSTRAINT economy_payment_log_action_check CHECK (action IN ('request','callback','verify','inquiry','store_validate'));
DELETE FROM economy_payment_method WHERE method IN ('google_play','apple','stripe');
ALTER TABLE economy_payment_method DROP CONSTRAINT IF EXISTS economy_payment_method_method_check;
ALTER TABLE economy_payment_method ADD CONSTRAINT economy_payment_method_method_check CHECK (method IN ('zarinpal','zibal','bazaar','myket','in_app'));
