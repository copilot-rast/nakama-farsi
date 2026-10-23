-- Platform-wide Jazebe profile/state extension keyed by the existing Nakama user.
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_user_profile (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    display_name VARCHAR(20) NOT NULL CHECK(char_length(display_name) BETWEEN 2 AND 20),
    display_name_changed_at TIMESTAMPTZ,
    display_name_change_count_month SMALLINT NOT NULL DEFAULT 0 CHECK(display_name_change_count_month>=0),
    avatar_kind VARCHAR(8) NOT NULL DEFAULT 'preset' CHECK(avatar_kind IN ('preset','upload')),
    avatar_preset_index SMALLINT CHECK(avatar_preset_index BETWEEN 0 AND 5),
    avatar_media_id UUID REFERENCES jazebe_media_upload(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK((avatar_kind='preset' AND avatar_preset_index IS NOT NULL AND avatar_media_id IS NULL)
       OR (avatar_kind='upload' AND avatar_preset_index IS NULL))
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_user_profile_name ON jazebe_user_profile(lower(display_name));
CREATE TABLE IF NOT EXISTS jazebe_user_account (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    state VARCHAR(16) NOT NULL DEFAULT 'active' CHECK(state IN ('active','deactivated','suspended')),
    deactivated_at TIMESTAMPTZ,
    suspended_by_operator_id UUID REFERENCES users(id),
    suspended_reason VARCHAR(512),
    suspended_at TIMESTAMPTZ,
    reactivated_at TIMESTAMPTZ,
    terms_version VARCHAR(32),
    terms_accepted_at TIMESTAMPTZ,
    signup_source_planet_id UUID REFERENCES jazebe_planet(id),
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_jazebe_user_account_state ON jazebe_user_account(state) WHERE state<>'active';

-- +migrate Down
-- Profile and account history are retained for explicit, reviewed rollback.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023170000-jazebe-account-profile'; END $$;
-- +migrate StatementEnd
