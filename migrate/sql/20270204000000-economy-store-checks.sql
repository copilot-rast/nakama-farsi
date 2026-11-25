-- +migrate Up
ALTER TABLE economy_payment_method ADD COLUMN IF NOT EXISTS store_last_failure JSONB;
ALTER TABLE economy_payment_method ADD COLUMN IF NOT EXISTS store_last_ok_at TIMESTAMPTZ;
ALTER TABLE economy_payment_method ADD COLUMN IF NOT EXISTS store_last_test JSONB;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % keeps store check history and is not reverted automatically; restore from backup', '20270204000000-economy-store-checks'; END $$;
-- +migrate StatementEnd
