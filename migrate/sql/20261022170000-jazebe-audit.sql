-- Insert-only record of privileged Jazebe actions.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_audit_log (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    action VARCHAR(64) NOT NULL,
    actor_type VARCHAR(16) NOT NULL CHECK (actor_type IN ('user','brand_admin','operator','system')),
    actor_id UUID,
    target_type VARCHAR(64) NOT NULL,
    target_id UUID,
    brand_id UUID,
    planet_id UUID,
    params JSONB,
    result VARCHAR(8) NOT NULL DEFAULT 'success',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_jazebe_audit_created ON jazebe_audit_log (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jazebe_audit_target ON jazebe_audit_log (target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_jazebe_audit_actor ON jazebe_audit_log (actor_type, actor_id);

-- +migrate Down
-- Audit records must be retained for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022170000-jazebe-audit'; END $$;
-- +migrate StatementEnd
