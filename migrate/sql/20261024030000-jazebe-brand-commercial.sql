-- The brand's commercial record (source jz_brand, PRD 12.3.4). The existing
-- name column is the trade name and is_active stays the tenant switch every
-- brand-scoped query already checks; contract_status drives it.
-- +migrate Up notransaction
ALTER TABLE jazebe_brand
    ADD COLUMN IF NOT EXISTS legal_name VARCHAR(256),
    ADD COLUMN IF NOT EXISTS national_company_id VARCHAR(32),
    ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(32),
    ADD COLUMN IF NOT EXISTS contact_email VARCHAR(320),
    ADD COLUMN IF NOT EXISTS contact_address TEXT,
    ADD COLUMN IF NOT EXISTS contract_status VARCHAR(16) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS contract_signed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
-- Brands deactivated before contracts existed read as suspended.
UPDATE jazebe_brand SET contract_status='suspended' WHERE NOT is_active AND contract_status='active';
ALTER TABLE jazebe_brand
    DROP CONSTRAINT IF EXISTS jazebe_brand_contract_status_check;
ALTER TABLE jazebe_brand
    ADD CONSTRAINT jazebe_brand_contract_status_check CHECK (contract_status IN ('active', 'suspended', 'terminated'));
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_brand_national_company_id ON jazebe_brand (national_company_id) WHERE national_company_id IS NOT NULL;

-- +migrate Down
-- Commercial records are business data.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261024030000-jazebe-brand-commercial'; END $$;
-- +migrate StatementEnd
