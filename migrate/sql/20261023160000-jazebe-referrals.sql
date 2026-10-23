-- Account referral identity and mission invitation qualification use Nakama user IDs.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_referral_identity (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    invite_code VARCHAR(16) NOT NULL UNIQUE,
    referred_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK(referred_by_user_id IS NULL OR referred_by_user_id<>user_id)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_referral_identity_inviter ON jazebe_referral_identity(referred_by_user_id)
    WHERE referred_by_user_id IS NOT NULL;
CREATE TABLE IF NOT EXISTS jazebe_mission_invite (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    mission_id UUID REFERENCES jazebe_mission(id),
    inviter_user_id UUID NOT NULL REFERENCES users(id),
    invitee_user_id UUID REFERENCES users(id),
    token VARCHAR(64) NOT NULL UNIQUE,
    group_id UUID REFERENCES jazebe_group(id),
    status VARCHAR(12) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','registered','qualified')),
    registered_at TIMESTAMPTZ,
    qualified_at TIMESTAMPTZ,
    completion_id UUID REFERENCES jazebe_mission_completion(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK(invitee_user_id IS NULL OR invitee_user_id<>inviter_user_id),
    CHECK(status='pending' OR invitee_user_id IS NOT NULL),
    CHECK(status<>'qualified' OR qualified_at IS NOT NULL)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_mission_invite_attribution ON jazebe_mission_invite
    (inviter_user_id,invitee_user_id,COALESCE(mission_id,'00000000-0000-0000-0000-000000000000'::uuid))
    WHERE invitee_user_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_mission_invite_unclaimed ON jazebe_mission_invite
    (mission_id,inviter_user_id) WHERE invitee_user_id IS NULL AND mission_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_jazebe_mission_invite_qualification ON jazebe_mission_invite(planet_id,invitee_user_id,status)
    WHERE status='registered';
CREATE INDEX IF NOT EXISTS idx_jazebe_mission_invite_history ON jazebe_mission_invite(planet_id,inviter_user_id,status,qualified_at);
CREATE TABLE IF NOT EXISTS jazebe_referral_completion_job (
    mission_id UUID NOT NULL REFERENCES jazebe_mission(id),
    inviter_user_id UUID NOT NULL REFERENCES users(id),
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    pending BOOLEAN NOT NULL DEFAULT true,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts>=0),
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(mission_id,inviter_user_id)
);
CREATE INDEX IF NOT EXISTS idx_jazebe_referral_completion_pending ON jazebe_referral_completion_job(next_attempt_at)
    WHERE pending;
-- +migrate Down
-- Keep referral relationships and qualification evidence for manual rollback.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023160000-jazebe-referrals'; END $$;
-- +migrate StatementEnd
