-- Cancellation metadata and the draw fields retained by the source lifecycle.
-- +migrate Up
ALTER TABLE jazebe_lottery ADD COLUMN IF NOT EXISTS cancel_reason VARCHAR(512);
ALTER TABLE jazebe_lottery ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;
ALTER TABLE jazebe_lottery ADD COLUMN IF NOT EXISTS drawn_at TIMESTAMPTZ;
ALTER TABLE jazebe_lottery ADD COLUMN IF NOT EXISTS draw_seed VARCHAR(128);
ALTER TABLE jazebe_lottery ADD COLUMN IF NOT EXISTS draw_meta JSONB;
CREATE INDEX IF NOT EXISTS idx_jazebe_lottery_cancelled_stranded ON jazebe_lottery(planet_id,status,cancelled_at) WHERE status='cancelled';

-- +migrate Down
-- Cancellation, draw and ticket provenance must survive manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023060000-jazebe-lottery-cancel'; END $$;
-- +migrate StatementEnd
