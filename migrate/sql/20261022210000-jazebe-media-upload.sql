-- Chunked Jazebe media uploads fit Nakama's existing HTTP request limit.
-- Upload bytes live in PostgreSQL so every Nakama instance sees the same asset.
-- +migrate Up
CREATE TABLE jazebe_media_upload (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet (id),
    owner_user_id UUID NOT NULL REFERENCES users (id),
    purpose VARCHAR(16) NOT NULL CHECK (purpose IN ('evidence','avatar','mission_media')),
    token_hash BYTEA NOT NULL,
    file_name VARCHAR(200) NOT NULL,
    expected_bytes INTEGER NOT NULL CHECK (expected_bytes BETWEEN 1 AND 52428800),
    received_bytes INTEGER NOT NULL DEFAULT 0 CHECK (received_bytes >= 0),
    received_chunks INTEGER NOT NULL DEFAULT 0 CHECK (received_chunks >= 0),
    status VARCHAR(12) NOT NULL DEFAULT 'open' CHECK (status IN ('open','complete','deleted')),
    mime_type VARCHAR(32),
    sha256 CHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT jazebe_media_received_bound CHECK (received_bytes <= expected_bytes)
);
CREATE INDEX idx_jazebe_media_owner ON jazebe_media_upload (planet_id,owner_user_id,purpose,status);
CREATE INDEX idx_jazebe_media_expiry ON jazebe_media_upload (expires_at) WHERE status='open';
CREATE TABLE jazebe_media_chunk (
    upload_id UUID NOT NULL REFERENCES jazebe_media_upload (id) ON DELETE CASCADE,
    seq INTEGER NOT NULL CHECK (seq >= 0),
    data BYTEA NOT NULL,
    size_bytes INTEGER NOT NULL CHECK (size_bytes BETWEEN 1 AND 131072),
    PRIMARY KEY (upload_id,seq),
    CONSTRAINT jazebe_media_chunk_size CHECK (octet_length(data) = size_bytes)
);

-- +migrate Down
-- Preserve uploaded member media for manual rollback review.
SELECT 1;
