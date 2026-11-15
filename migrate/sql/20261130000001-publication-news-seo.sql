-- +migrate Up notransaction
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS slug TEXT NOT NULL DEFAULT '';
UPDATE publication_news SET slug = id::text WHERE slug = '';
CREATE UNIQUE INDEX IF NOT EXISTS publication_news_slug_idx ON publication_news (slug) WHERE slug <> '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS meta_title JSONB NOT NULL DEFAULT '{}';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS meta_description JSONB NOT NULL DEFAULT '{}';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS keywords TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS canonical_url TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS og_image TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS noindex BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE publication_news_settings ADD COLUMN IF NOT EXISTS slug_style TEXT NOT NULL DEFAULT 'native';
ALTER TABLE publication_news_settings DROP CONSTRAINT IF EXISTS publication_news_settings_slug_style_check;
ALTER TABLE publication_news_settings ADD CONSTRAINT publication_news_settings_slug_style_check CHECK (slug_style IN ('native','latin','id'));
ALTER TABLE publication_news_settings ADD COLUMN IF NOT EXISTS public_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE publication_news_settings ADD COLUMN IF NOT EXISTS public_base_url TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news_settings ADD COLUMN IF NOT EXISTS public_path TEXT NOT NULL DEFAULT '/news/{slug}';
ALTER TABLE publication_news_settings ADD COLUMN IF NOT EXISTS rss_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE publication_news_settings ADD COLUMN IF NOT EXISTS public_cache_seconds INT NOT NULL DEFAULT 300;
ALTER TABLE publication_news_settings DROP CONSTRAINT IF EXISTS publication_news_settings_public_cache_check;
ALTER TABLE publication_news_settings ADD CONSTRAINT publication_news_settings_public_cache_check CHECK (public_cache_seconds BETWEEN 0 AND 86400);
ALTER TABLE publication_news_settings ADD COLUMN IF NOT EXISTS site_title JSONB NOT NULL DEFAULT '{}';

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % keeps news SEO fields and cannot be reverted automatically; restore from backup', '20261130000001-publication-news-seo'; END $$;
-- +migrate StatementEnd
