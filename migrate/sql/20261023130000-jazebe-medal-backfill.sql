-- Bounded, resumable medal catch-up. No rows are processed until the worker is enabled.
-- +migrate Up
CREATE TABLE jazebe_medal_backfill (
    planet_medal_id UUID PRIMARY KEY REFERENCES jazebe_planet_medal(id),
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    last_user_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    evaluated BIGINT NOT NULL DEFAULT 0 CHECK (evaluated >= 0),
    awarded BIGINT NOT NULL DEFAULT 0 CHECK (awarded >= 0),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    last_error TEXT,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_jazebe_medal_backfill_pending ON jazebe_medal_backfill(next_attempt_at,planet_medal_id)
    WHERE completed_at IS NULL;
-- +migrate Down
-- Preserve backfill progress and awarded medals for manual rollback review.
SELECT 1;
