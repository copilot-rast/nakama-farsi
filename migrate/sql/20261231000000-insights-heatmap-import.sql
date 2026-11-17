-- +migrate Up
ALTER TABLE insights_heatmap_upload ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'game';
ALTER TABLE insights_heatmap_upload ADD COLUMN IF NOT EXISTS label TEXT NOT NULL DEFAULT '';
ALTER TABLE insights_heatmap_upload ADD COLUMN IF NOT EXISTS file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE insights_heatmap_upload ADD COLUMN IF NOT EXISTS content_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE insights_heatmap_upload ADD COLUMN IF NOT EXISTS uploaded_by TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS insights_heatmap_upload_hash_idx ON insights_heatmap_upload (map_key, content_hash) WHERE content_hash <> '';

-- +migrate Down
DROP INDEX IF EXISTS insights_heatmap_upload_hash_idx;
ALTER TABLE insights_heatmap_upload DROP COLUMN IF EXISTS uploaded_by;
ALTER TABLE insights_heatmap_upload DROP COLUMN IF EXISTS content_hash;
ALTER TABLE insights_heatmap_upload DROP COLUMN IF EXISTS file_name;
ALTER TABLE insights_heatmap_upload DROP COLUMN IF EXISTS label;
ALTER TABLE insights_heatmap_upload DROP COLUMN IF EXISTS source;
