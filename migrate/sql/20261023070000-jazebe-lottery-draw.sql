-- Immutable winner and share snapshots. Verification is default-deny until
-- the source KYC workflow marks a Nakama user verified.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_user_verification (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    status VARCHAR(16) NOT NULL DEFAULT 'none' CHECK (status IN ('none','review','verified','rejected')),
    verified_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS jazebe_lottery_winner (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    lottery_id UUID NOT NULL REFERENCES jazebe_lottery(id),
    prize_id UUID NOT NULL REFERENCES jazebe_lottery_prize(id),
    ticket_id UUID NOT NULL REFERENCES jazebe_lottery_ticket(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    delivery_status VARCHAR(12) NOT NULL DEFAULT 'pending' CHECK (delivery_status IN ('pending','contacted','delivered','rejected')),
    delivery_note VARCHAR(512),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(prize_id)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_lottery_winner_lottery ON jazebe_lottery_winner(lottery_id);
CREATE INDEX IF NOT EXISTS idx_jazebe_lottery_winner_group ON jazebe_lottery_winner(group_id);
CREATE TABLE IF NOT EXISTS jazebe_lottery_winner_share (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    winner_id UUID NOT NULL REFERENCES jazebe_lottery_winner(id),
    user_id UUID NOT NULL REFERENCES users(id),
    individual_ticket_count INTEGER NOT NULL DEFAULT 0 CHECK (individual_ticket_count>=0),
    share_percent NUMERIC(7,4) NOT NULL CHECK (share_percent>=0 AND share_percent<=100),
    share_value_rial BIGINT NOT NULL CHECK (share_value_rial>=0),
    payout_status VARCHAR(24) NOT NULL CHECK (payout_status IN ('pending_verification','payable','paid','rejected')),
    blocked_by_kyc BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(winner_id,user_id)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_lottery_share_user ON jazebe_lottery_winner_share(user_id,payout_status);

-- +migrate Down
-- Draw results, identity verification and payout evidence require manual review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023070000-jazebe-lottery-draw'; END $$;
-- +migrate StatementEnd
