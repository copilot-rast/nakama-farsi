-- Jazebe's notification centre (source notification-centre.service.ts and
-- jz_notification_preference). Jazebe notices are Nakama notification rows
-- (code 1001); Nakama has no read or soft-delete state, so it is kept here.
-- A row disappears with its notification (ON DELETE CASCADE).
-- +migrate Up
CREATE TABLE IF NOT EXISTS jazebe_notification_state (
    notification_id UUID PRIMARY KEY REFERENCES notification (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_jazebe_notification_state_user ON jazebe_notification_state (user_id);

-- One row per user; critical notices have no toggle (always delivered).
CREATE TABLE IF NOT EXISTS jazebe_notification_preference (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    channel_prefs JSONB NOT NULL DEFAULT '{"group":{"in_app":true,"sms":false},"mission_game":{"in_app":true,"sms":false},"system":{"in_app":true,"sms":false}}',
    category_prefs JSONB NOT NULL DEFAULT '{"missions":true,"rewards":true,"friends":true,"news":true,"marketing":false}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
-- Read state and preferences are user data.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261024010000-jazebe-notification-centre'; END $$;
-- +migrate StatementEnd
