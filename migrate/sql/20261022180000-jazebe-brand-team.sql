-- Brand tenant roles map directly to Nakama user UUIDs.
-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS jazebe_brand_member (
    brand_id UUID NOT NULL REFERENCES jazebe_brand (id),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL CHECK (role IN ('owner','admin','member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (brand_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_brand_member_user ON jazebe_brand_member (user_id, brand_id);
CREATE TABLE IF NOT EXISTS jazebe_brand_admin (
    brand_id UUID NOT NULL REFERENCES jazebe_brand (id),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL CHECK (role IN ('main','technical','marketing')),
    full_name VARCHAR(128) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (brand_id, user_id),
    FOREIGN KEY (brand_id, user_id) REFERENCES jazebe_brand_member (brand_id, user_id)
);
INSERT INTO jazebe_brand_member (brand_id,user_id,role)
    SELECT id,owner_id,'owner' FROM jazebe_brand ON CONFLICT DO NOTHING;
INSERT INTO jazebe_brand_admin (brand_id,user_id,role,full_name)
    SELECT id,owner_id,'main',name FROM jazebe_brand ON CONFLICT DO NOTHING;

-- +migrate Down
-- Retain team and role records for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022180000-jazebe-brand-team'; END $$;
-- +migrate StatementEnd
