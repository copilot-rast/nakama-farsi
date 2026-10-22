-- Cancellation metadata and the draw fields retained by the source lifecycle.
-- +migrate Up
ALTER TABLE jazebe_lottery ADD COLUMN cancel_reason VARCHAR(512);
ALTER TABLE jazebe_lottery ADD COLUMN cancelled_at TIMESTAMPTZ;
ALTER TABLE jazebe_lottery ADD COLUMN drawn_at TIMESTAMPTZ;
ALTER TABLE jazebe_lottery ADD COLUMN draw_seed VARCHAR(128);
ALTER TABLE jazebe_lottery ADD COLUMN draw_meta JSONB;
CREATE INDEX idx_jazebe_lottery_cancelled_stranded ON jazebe_lottery(planet_id,status,cancelled_at) WHERE status='cancelled';

-- +migrate Down
-- Cancellation, draw and ticket provenance must survive manual rollback review.
SELECT 1;
