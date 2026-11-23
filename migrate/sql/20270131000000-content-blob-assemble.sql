-- +migrate Up
ALTER TABLE content_blob
    ADD COLUMN IF NOT EXISTS assemble_holder TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS assemble_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS assemble_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS assemble_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS assemble_error TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS content_blob_assemble_idx ON content_blob (assemble_until) WHERE NOT stored;
ALTER TABLE content_pack_event ADD COLUMN IF NOT EXISTS detail JSONB NOT NULL DEFAULT '{}';

-- +migrate Down
ALTER TABLE content_pack_event DROP COLUMN IF EXISTS detail;
DROP INDEX IF EXISTS content_blob_assemble_idx;
ALTER TABLE content_blob
    DROP COLUMN IF EXISTS assemble_error,
    DROP COLUMN IF EXISTS assemble_attempts,
    DROP COLUMN IF EXISTS assemble_started_at,
    DROP COLUMN IF EXISTS assemble_until,
    DROP COLUMN IF EXISTS assemble_holder;
