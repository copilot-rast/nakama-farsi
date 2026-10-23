-- Pre-generated tenant-scoped coupon stock. Issuance is an atomic locked claim.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_coupon_code (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    shop_product_id UUID NOT NULL REFERENCES jazebe_shop_product(id),
    code VARCHAR(64) NOT NULL,
    status VARCHAR(12) NOT NULL DEFAULT 'available' CHECK (status IN ('available','issued','expired')),
    issued_to_user_id UUID REFERENCES users(id),
    issued_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_jazebe_coupon_code UNIQUE(shop_product_id,code),
    CONSTRAINT ck_jazebe_coupon_issued CHECK (status <> 'issued' OR (issued_to_user_id IS NOT NULL AND issued_at IS NOT NULL))
);
CREATE INDEX IF NOT EXISTS idx_jazebe_coupon_product_status ON jazebe_coupon_code(shop_product_id,status,created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jazebe_coupon_user ON jazebe_coupon_code(issued_to_user_id,issued_at DESC) WHERE issued_to_user_id IS NOT NULL;

-- +migrate Down
-- Issued codes are customer property; rollback requires a reviewed data migration.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023080000-jazebe-coupon-inventory'; END $$;
-- +migrate StatementEnd
