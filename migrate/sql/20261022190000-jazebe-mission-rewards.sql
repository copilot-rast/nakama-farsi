-- Mission actions and per-planet point economy. All payout rows share the
-- completion transaction; Nakama's separate wallet transaction is not used.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_mission_completion (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    mission_id UUID NOT NULL REFERENCES jazebe_mission (id),
    user_id UUID NOT NULL REFERENCES users (id),
    status VARCHAR(16) NOT NULL CHECK (status IN ('recorded','pending_review','approved','rejected')),
    period_key VARCHAR(24),
    evidence JSONB,
    reviewer_admin_id UUID REFERENCES users (id),
    review_note VARCHAR(512),
    reviewed_at TIMESTAMPTZ,
    awarded_points INTEGER NOT NULL DEFAULT 0 CHECK (awarded_points >= 0),
    awarded_xp INTEGER NOT NULL DEFAULT 0 CHECK (awarded_xp >= 0),
    point_transaction_id UUID,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT jazebe_completion_reject_reason CHECK (status <> 'rejected' OR review_note IS NOT NULL)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_completion_approved ON jazebe_mission_completion
    (mission_id,user_id,COALESCE(period_key,'')) WHERE status='approved' AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_completion_pending ON jazebe_mission_completion
    (mission_id,user_id) WHERE status='pending_review' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jazebe_completion_user_history ON jazebe_mission_completion
    (planet_id,user_id,submitted_at DESC);
CREATE TABLE IF NOT EXISTS jazebe_mission_media_start (
    mission_id UUID NOT NULL REFERENCES jazebe_mission (id),
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    user_id UUID NOT NULL REFERENCES users (id),
    period_key VARCHAR(32) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (mission_id,user_id,period_key)
);
CREATE TABLE IF NOT EXISTS jazebe_point_wallet (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    user_id UUID NOT NULL REFERENCES users (id),
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    last_transaction_at TIMESTAMPTZ,
    UNIQUE (planet_id,user_id)
);
CREATE TABLE IF NOT EXISTS jazebe_point_transaction (
    id UUID PRIMARY KEY,
    seq BIGSERIAL NOT NULL,
    wallet_id UUID NOT NULL REFERENCES jazebe_point_wallet (id),
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    user_id UUID NOT NULL REFERENCES users (id),
    direction VARCHAR(8) NOT NULL CHECK (direction IN ('credit','debit')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    balance_after BIGINT NOT NULL CHECK (balance_after >= 0),
    source VARCHAR(32) NOT NULL,
    source_id UUID NOT NULL,
    description VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (wallet_id,source,source_id),
    UNIQUE (wallet_id,seq)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_point_transaction_user_seq ON jazebe_point_transaction (planet_id,user_id,seq DESC);
ALTER TABLE jazebe_mission_completion DROP CONSTRAINT IF EXISTS jazebe_completion_point_transaction_fk, ADD CONSTRAINT jazebe_completion_point_transaction_fk
    FOREIGN KEY (point_transaction_id) REFERENCES jazebe_point_transaction (id);
CREATE TABLE IF NOT EXISTS jazebe_xp_record (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    user_id UUID NOT NULL REFERENCES users (id),
    amount INTEGER NOT NULL CHECK (amount > 0),
    source VARCHAR(32) NOT NULL,
    source_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (planet_id,user_id,source,source_id)
);
CREATE TABLE IF NOT EXISTS jazebe_user_planet_stats (
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    user_id UUID NOT NULL REFERENCES users (id),
    total_xp BIGINT NOT NULL DEFAULT 0 CHECK (total_xp >= 0),
    level SMALLINT NOT NULL DEFAULT 1 CHECK (level BETWEEN 1 AND 3),
    level_key VARCHAR(32) NOT NULL DEFAULT 'newcomer',
    missions_completed INTEGER NOT NULL DEFAULT 0,
    games_played INTEGER NOT NULL DEFAULT 0,
    tickets_created INTEGER NOT NULL DEFAULT 0,
    medals_count INTEGER NOT NULL DEFAULT 0,
    last_activity_at TIMESTAMPTZ,
    PRIMARY KEY (planet_id,user_id)
);
-- +migrate Down
-- Preserve earned points and mission history for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261022190000-jazebe-mission-rewards'; END $$;
-- +migrate StatementEnd
