-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS publication_faq (
    id UUID PRIMARY KEY,
    category_id UUID REFERENCES economy_category (id),
    question JSONB NOT NULL,
    answer JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS publication_faq_active_sort_idx ON publication_faq (is_active, sort, created_at);
CREATE INDEX IF NOT EXISTS publication_faq_category_idx ON publication_faq (category_id);

CREATE TABLE IF NOT EXISTS publication_news (
    id UUID PRIMARY KEY,
    title JSONB NOT NULL,
    summary JSONB NOT NULL DEFAULT '{}',
    body JSONB NOT NULL DEFAULT '{}',
    cover_url TEXT NOT NULL DEFAULT '',
    category_id UUID REFERENCES economy_category (id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_pinned BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    allow_comment BOOLEAN NOT NULL DEFAULT true,
    free_comment BOOLEAN NOT NULL DEFAULT false,
    actions JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS publication_news_feed_idx ON publication_news (is_active, is_pinned DESC, published_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS publication_news_category_idx ON publication_news (category_id);

CREATE TABLE IF NOT EXISTS publication_news_seen (
    news_id UUID NOT NULL REFERENCES publication_news (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (news_id, user_id)
);
CREATE INDEX IF NOT EXISTS publication_news_seen_user_idx ON publication_news_seen (user_id);

CREATE TABLE IF NOT EXISTS publication_news_comment (
    id UUID PRIMARY KEY,
    news_id UUID NOT NULL REFERENCES publication_news (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    parent_id UUID REFERENCES publication_news_comment (id) ON DELETE CASCADE,
    message TEXT NOT NULL CHECK (char_length(message) BETWEEN 1 AND 1000),
    is_accepted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS publication_news_comment_thread_idx ON publication_news_comment (news_id, parent_id, created_at);
CREATE INDEX IF NOT EXISTS publication_news_comment_user_idx ON publication_news_comment (user_id, created_at);
CREATE INDEX IF NOT EXISTS publication_news_comment_pending_idx ON publication_news_comment (created_at) WHERE NOT is_accepted;

CREATE TABLE IF NOT EXISTS publication_provider (
    id UUID PRIMARY KEY,
    channel TEXT NOT NULL CHECK (channel IN ('sms','email')),
    kind TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    is_default BOOLEAN NOT NULL DEFAULT false,
    sandbox BOOLEAN NOT NULL DEFAULT true,
    config TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (channel, kind)
);
CREATE UNIQUE INDEX IF NOT EXISTS publication_provider_default_idx ON publication_provider (channel) WHERE is_default;

CREATE TABLE IF NOT EXISTS publication_notification (
    id UUID PRIMARY KEY,
    channel TEXT NOT NULL CHECK (channel IN ('in_app','sms','email')),
    title JSONB NOT NULL DEFAULT '{}',
    body JSONB NOT NULL,
    action JSONB,
    target JSONB NOT NULL,
    provider TEXT,
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','sending','done','failed')),
    recipient_count INT NOT NULL DEFAULT 0,
    sent_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    skipped_count INT NOT NULL DEFAULT 0,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS publication_notification_list_idx ON publication_notification (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS publication_notification_open_idx ON publication_notification (status) WHERE status IN ('queued','sending');

CREATE TABLE IF NOT EXISTS publication_notification_recipient (
    notification_id UUID NOT NULL REFERENCES publication_notification (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','failed','skipped')),
    address TEXT,
    error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    seen_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    PRIMARY KEY (notification_id, user_id)
);
CREATE INDEX IF NOT EXISTS publication_recipient_due_idx ON publication_notification_recipient (next_attempt_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS publication_recipient_user_idx ON publication_notification_recipient (user_id, notification_id);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % retains notification and comment records and cannot be reverted automatically; restore from backup', '20261031000001-publication-schema'; END $$;
-- +migrate StatementEnd
