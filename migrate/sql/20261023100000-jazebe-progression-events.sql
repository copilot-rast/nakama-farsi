-- Keep after-commit XP activities recoverable when a Nakama process exits.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_progression_event (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    user_id UUID NOT NULL REFERENCES users(id),
    kind VARCHAR(20) NOT NULL CHECK (kind IN ('purchase','group_join')),
    source_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    UNIQUE (kind,source_id,user_id)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_progression_pending ON jazebe_progression_event(created_at,id)
    WHERE processed_at IS NULL;

-- +migrate Down
-- Preserve unprocessed activity and earned XP for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023100000-jazebe-progression-events'; END $$;
-- +migrate StatementEnd
