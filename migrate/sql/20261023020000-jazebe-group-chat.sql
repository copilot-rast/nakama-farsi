-- Group chat is planet scoped. Global sequence supports descending cursor pages.
-- +migrate Up
-- BIGSERIAL spelled out: CockroachDB would otherwise fill seq from
-- unique_rowid(), whose values exceed the 2^53 a JSON client reads exactly.
CREATE SEQUENCE IF NOT EXISTS jazebe_group_message_seq_seq;
CREATE TABLE IF NOT EXISTS jazebe_group_message (
    id UUID PRIMARY KEY,
    seq BIGINT NOT NULL DEFAULT nextval('jazebe_group_message_seq_seq') UNIQUE,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    sender_user_id UUID REFERENCES users(id),
    kind VARCHAR(8) NOT NULL DEFAULT 'text' CHECK (kind IN ('text','event','system')),
    body TEXT,
    reply_to_message_id UUID REFERENCES jazebe_group_message(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users(id),
    deleted_as_admin BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT jazebe_group_message_body CHECK (kind <> 'text' OR (body IS NOT NULL AND char_length(body) BETWEEN 1 AND 2000))
);
ALTER SEQUENCE jazebe_group_message_seq_seq OWNED BY jazebe_group_message.seq;
CREATE INDEX IF NOT EXISTS idx_jazebe_group_message_page ON jazebe_group_message(group_id,seq DESC);
CREATE INDEX IF NOT EXISTS idx_jazebe_group_message_rate ON jazebe_group_message(group_id,sender_user_id,created_at DESC) WHERE kind='text';
CREATE TABLE IF NOT EXISTS jazebe_group_message_read (
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    user_id UUID NOT NULL REFERENCES users(id),
    last_read_seq BIGINT NOT NULL DEFAULT 0 CHECK (last_read_seq >= 0),
    PRIMARY KEY(group_id,user_id)
);
CREATE TABLE IF NOT EXISTS jazebe_user_report (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    reporter_user_id UUID NOT NULL REFERENCES users(id),
    subject_type VARCHAR(16) NOT NULL CHECK (subject_type IN ('message','group','user')),
    subject_id UUID NOT NULL,
    reason VARCHAR(64) NOT NULL,
    note VARCHAR(512),
    status VARCHAR(12) NOT NULL DEFAULT 'open' CHECK (status IN ('open','reviewing','actioned','dismissed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    handled_by_admin_id UUID REFERENCES users(id),
    handled_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_jazebe_user_report_open ON jazebe_user_report(reporter_user_id,subject_type,subject_id)
    WHERE status IN ('open','reviewing');
CREATE INDEX IF NOT EXISTS idx_jazebe_user_report_subject ON jazebe_user_report(subject_type,subject_id,status);

-- +migrate Down
-- Retain chat and report evidence for manual rollback review.
-- Refused rather than a no-op, which would drop this migration from
-- migration_info while its schema stays. See docs/ANDRONIX-POSTGRESQL.md
-- "Rolling back Jazebe migrations" for the backed-up manual procedure.
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Jazebe migration % retains data and cannot be reverted automatically; restore from backup or use the documented rollback procedure', '20261023020000-jazebe-group-chat'; END $$;
-- +migrate StatementEnd
