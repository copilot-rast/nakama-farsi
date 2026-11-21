-- +migrate Up
CREATE TABLE IF NOT EXISTS base_voice_room (
    room TEXT PRIMARY KEY,
    provider TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL DEFAULT 'room',
    game_key TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_token_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS base_voice_room_expires_idx ON base_voice_room (expires_at DESC);

CREATE TABLE IF NOT EXISTS base_voice_member (
    room TEXT NOT NULL REFERENCES base_voice_room (room) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    can_publish BOOLEAN NOT NULL DEFAULT true,
    first_token_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_token_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room, user_id)
);

CREATE TABLE IF NOT EXISTS base_voice_mute (
    room TEXT NOT NULL,
    user_id UUID NOT NULL,
    muted_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room, user_id)
);

-- +migrate Down
DROP TABLE IF EXISTS base_voice_mute;
DROP TABLE IF EXISTS base_voice_member;
DROP TABLE IF EXISTS base_voice_room;
