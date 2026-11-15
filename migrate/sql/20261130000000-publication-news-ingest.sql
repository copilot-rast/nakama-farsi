-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS publication_news_source (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9_-]{1,39}$'),
    name TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    auth_mode TEXT NOT NULL DEFAULT 'token' CHECK (auth_mode IN ('token','hmac')),
    token_header TEXT NOT NULL DEFAULT 'X-News-Token',
    signature_header TEXT NOT NULL DEFAULT 'X-News-Signature',
    timestamp_header TEXT NOT NULL DEFAULT 'X-News-Timestamp',
    skew_seconds INT NOT NULL DEFAULT 300 CHECK (skew_seconds BETWEEN 30 AND 3600),
    secret TEXT NOT NULL DEFAULT '',
    secret_hint TEXT NOT NULL DEFAULT '',
    secret_rotated_at TIMESTAMPTZ,
    items_path TEXT NOT NULL DEFAULT '',
    mapping JSONB NOT NULL DEFAULT '{}',
    body_format TEXT NOT NULL DEFAULT 'auto' CHECK (body_format IN ('auto','html','markdown','text')),
    default_category_id UUID REFERENCES base_category (id) ON DELETE SET NULL,
    default_lang TEXT NOT NULL DEFAULT '',
    default_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    publish_mode TEXT NOT NULL DEFAULT 'review' CHECK (publish_mode IN ('review','auto')),
    update_mode TEXT NOT NULL DEFAULT 'update' CHECK (update_mode IN ('update','ignore')),
    published_at_mode TEXT NOT NULL DEFAULT 'source' CHECK (published_at_mode IN ('source','received')),
    rate_limit_per_minute INT NOT NULL DEFAULT 60 CHECK (rate_limit_per_minute BETWEEN 0 AND 10000),
    copy_images BOOLEAN NOT NULL DEFAULT false,
    notify_mode TEXT NOT NULL DEFAULT 'none' CHECK (notify_mode IN ('none','in_app','push','both')),
    allow_comment BOOLEAN NOT NULL DEFAULT true,
    allow_reaction BOOLEAN NOT NULL DEFAULT true,
    last_received_at TIMESTAMPTZ,
    created_by TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'published';
UPDATE publication_news SET status='draft' WHERE NOT is_active AND status='published';
ALTER TABLE publication_news DROP CONSTRAINT IF EXISTS publication_news_status_check;
ALTER TABLE publication_news ADD CONSTRAINT publication_news_status_check CHECK (status IN ('draft','pending_review','published','archived','rejected'));
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'console';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS source_id UUID REFERENCES publication_news_source (id) ON DELETE SET NULL;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS external_id TEXT;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS source_url TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS content_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS lang TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS body_format TEXT NOT NULL DEFAULT 'markdown';
ALTER TABLE publication_news DROP CONSTRAINT IF EXISTS publication_news_body_format_check;
ALTER TABLE publication_news ADD CONSTRAINT publication_news_body_format_check CHECK (body_format IN ('markdown','html'));
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS tag_ids TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS allow_reaction BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS notify_mode TEXT NOT NULL DEFAULT 'none';
ALTER TABLE publication_news DROP CONSTRAINT IF EXISTS publication_news_notify_mode_check;
ALTER TABLE publication_news ADD CONSTRAINT publication_news_notify_mode_check CHECK (notify_mode IN ('none','in_app','push','both'));
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS notify_audience JSONB NOT NULL DEFAULT '{}';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS notified_at TIMESTAMPTZ DEFAULT now();
ALTER TABLE publication_news ALTER COLUMN notified_at DROP DEFAULT;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS notify_claimed_until TIMESTAMPTZ;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS notification_id UUID;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS approved_by TEXT;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS reject_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_news ADD COLUMN IF NOT EXISTS updated_by TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS publication_news_external_idx ON publication_news (source, external_id) WHERE external_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS publication_news_status_idx ON publication_news (status, published_at DESC);
CREATE INDEX IF NOT EXISTS publication_news_notify_due_idx ON publication_news (published_at) WHERE notified_at IS NULL AND status = 'published';
CREATE INDEX IF NOT EXISTS publication_news_tags_idx ON publication_news USING GIN (tag_ids);

CREATE TABLE IF NOT EXISTS publication_news_ingest_log (
    id UUID PRIMARY KEY,
    source_id UUID REFERENCES publication_news_source (id) ON DELETE CASCADE,
    source_key TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('created','updated','unchanged','ignored','rejected','error','preview')),
    http_status INT NOT NULL DEFAULT 200,
    items INT NOT NULL DEFAULT 0,
    created_count INT NOT NULL DEFAULT 0,
    updated_count INT NOT NULL DEFAULT 0,
    unchanged_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    error_code TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    results JSONB NOT NULL DEFAULT '[]',
    payload TEXT NOT NULL DEFAULT '',
    remote_addr TEXT NOT NULL DEFAULT '',
    is_test BOOLEAN NOT NULL DEFAULT false,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS publication_news_ingest_log_list_idx ON publication_news_ingest_log (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS publication_news_ingest_log_source_idx ON publication_news_ingest_log (source_id, created_at DESC);

CREATE TABLE IF NOT EXISTS publication_news_replay (
    source_id UUID NOT NULL REFERENCES publication_news_source (id) ON DELETE CASCADE,
    signature TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (source_id, signature)
);
CREATE INDEX IF NOT EXISTS publication_news_replay_expiry_idx ON publication_news_replay (expires_at);

CREATE TABLE IF NOT EXISTS publication_news_reaction_type (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9_]{1,63}$'),
    icon TEXT NOT NULL DEFAULT '',
    name JSONB NOT NULL DEFAULT '{}',
    sort INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO publication_news_reaction_type (id, key, icon, name, sort) VALUES
    ('7a0e5f3e-0b1c-4f3e-9d61-000000000001', 'like', '👍', '{"fa":"پسندیدم","en":"Like"}', 1),
    ('7a0e5f3e-0b1c-4f3e-9d61-000000000002', 'love', '❤️', '{"fa":"عالی","en":"Love"}', 2),
    ('7a0e5f3e-0b1c-4f3e-9d61-000000000003', 'haha', '😂', '{"fa":"خنده‌دار","en":"Haha"}', 3),
    ('7a0e5f3e-0b1c-4f3e-9d61-000000000004', 'wow', '😮', '{"fa":"شگفت‌انگیز","en":"Wow"}', 4),
    ('7a0e5f3e-0b1c-4f3e-9d61-000000000005', 'sad', '😢', '{"fa":"ناراحت","en":"Sad"}', 5)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS publication_news_reaction (
    news_id UUID NOT NULL REFERENCES publication_news (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reaction TEXT NOT NULL REFERENCES publication_news_reaction_type (key) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (news_id, user_id)
);
CREATE INDEX IF NOT EXISTS publication_news_reaction_count_idx ON publication_news_reaction (news_id, reaction);
CREATE INDEX IF NOT EXISTS publication_news_reaction_user_idx ON publication_news_reaction (user_id);

CREATE TABLE IF NOT EXISTS publication_news_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    comments_per_minute INT NOT NULL DEFAULT 5 CHECK (comments_per_minute BETWEEN 1 AND 600),
    reactions_per_minute INT NOT NULL DEFAULT 30 CHECK (reactions_per_minute BETWEEN 1 AND 600),
    unread_scope TEXT NOT NULL DEFAULT 'signup' CHECK (unread_scope IN ('signup','all')),
    show_expired BOOLEAN NOT NULL DEFAULT false,
    fill_default_lang BOOLEAN NOT NULL DEFAULT true,
    realtime_on_publish BOOLEAN NOT NULL DEFAULT true,
    notify_category_key TEXT NOT NULL DEFAULT 'news',
    log_retention_days INT NOT NULL DEFAULT 30 CHECK (log_retention_days BETWEEN 1 AND 365),
    log_payload_bytes INT NOT NULL DEFAULT 16384 CHECK (log_payload_bytes BETWEEN 0 AND 262144),
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO publication_news_settings (id) VALUES (1) ON CONFLICT DO NOTHING;

ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS news_id UUID REFERENCES publication_news (id) ON DELETE SET NULL;
ALTER TABLE publication_notification DROP CONSTRAINT IF EXISTS publication_notification_kind_check;
ALTER TABLE publication_notification ADD CONSTRAINT publication_notification_kind_check CHECK (kind IN ('message','campaign','auto','test','news'));

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % keeps news sources, reactions and ingest logs and cannot be reverted automatically; restore from backup', '20261130000000-publication-news-ingest'; END $$;
-- +migrate StatementEnd
