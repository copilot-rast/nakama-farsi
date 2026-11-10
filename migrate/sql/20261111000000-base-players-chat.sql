-- +migrate Up notransaction
ALTER TABLE base_tag ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS base_user_tag (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES base_tag (id) ON DELETE CASCADE,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, tag_id)
);
CREATE INDEX IF NOT EXISTS base_user_tag_tag_idx ON base_user_tag (tag_id, user_id);

CREATE TABLE IF NOT EXISTS base_restriction (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('suspend','mute','warn')),
    status TEXT NOT NULL CHECK (status IN ('active','lifted','expired')),
    starts_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at TIMESTAMPTZ NOT NULL,
    reason JSONB NOT NULL DEFAULT '{}',
    note TEXT NOT NULL DEFAULT '',
    report_id UUID,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    lifted_by TEXT NOT NULL DEFAULT '',
    lifted_at TIMESTAMPTZ,
    lift_note TEXT NOT NULL DEFAULT '',
    ended_at TIMESTAMPTZ,
    CHECK (ends_at >= starts_at)
);
CREATE UNIQUE INDEX IF NOT EXISTS base_restriction_one_active_idx ON base_restriction (user_id, kind) WHERE status = 'active' AND kind <> 'warn';
CREATE INDEX IF NOT EXISTS base_restriction_user_idx ON base_restriction (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS base_restriction_due_idx ON base_restriction (ends_at) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS base_chat_word (
    id UUID PRIMARY KEY,
    word TEXT NOT NULL CHECK (char_length(word) BETWEEN 1 AND 64),
    action TEXT NOT NULL CHECK (action IN ('mask','block','flag')),
    match TEXT NOT NULL DEFAULT 'word' CHECK (match IN ('word','contains')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    note TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS base_chat_word_word_idx ON base_chat_word (lower(word));

CREATE TABLE IF NOT EXISTS base_report_title (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z0-9][a-z0-9_.-]{0,63}$'),
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_chat_report (
    id UUID PRIMARY KEY,
    source TEXT NOT NULL CHECK (source IN ('player','filter')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','dismissed','resolved')),
    message_id UUID NOT NULL,
    channel_id TEXT NOT NULL DEFAULT '',
    channel_label TEXT NOT NULL DEFAULT '',
    sender_id UUID REFERENCES users (id) ON DELETE SET NULL,
    sender_username TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    message_at TIMESTAMPTZ,
    context JSONB NOT NULL DEFAULT '[]',
    reporter_id UUID REFERENCES users (id) ON DELETE SET NULL,
    reporter_username TEXT NOT NULL DEFAULT '',
    title_id UUID REFERENCES base_report_title (id) ON DELETE SET NULL,
    title JSONB NOT NULL DEFAULT '{}',
    note TEXT NOT NULL DEFAULT '',
    matched JSONB NOT NULL DEFAULT '[]',
    action TEXT NOT NULL DEFAULT '' CHECK (action IN ('','dismiss','delete','warn','mute','suspend')),
    restriction_id UUID,
    message_deleted BOOLEAN NOT NULL DEFAULT false,
    resolved_by TEXT NOT NULL DEFAULT '',
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS base_chat_report_player_once_idx ON base_chat_report (message_id, reporter_id) WHERE source = 'player';
CREATE UNIQUE INDEX IF NOT EXISTS base_chat_report_filter_once_idx ON base_chat_report (message_id) WHERE source = 'filter';
CREATE INDEX IF NOT EXISTS base_chat_report_status_idx ON base_chat_report (status, created_at DESC);
CREATE INDEX IF NOT EXISTS base_chat_report_sender_idx ON base_chat_report (sender_id, created_at DESC);

INSERT INTO base_report_title (id, key, name, sort) VALUES
(gen_random_uuid(), 'abuse', '{"fa":"توهین یا فحاشی","en":"Insults or swearing","ar":"إهانة أو شتائم"}', 1),
(gen_random_uuid(), 'harassment', '{"fa":"آزار یا تهدید","en":"Harassment or threats","ar":"تحرش أو تهديد"}', 2),
(gen_random_uuid(), 'spam', '{"fa":"هرزنامه یا تبلیغ","en":"Spam or advertising","ar":"رسائل مزعجة أو إعلانات"}', 3),
(gen_random_uuid(), 'cheating', '{"fa":"تقلب در بازی","en":"Cheating","ar":"غش في اللعبة"}', 4),
(gen_random_uuid(), 'personal-info', '{"fa":"پخش اطلاعات شخصی","en":"Sharing personal information","ar":"نشر معلومات شخصية"}', 5),
(gen_random_uuid(), 'other', '{"fa":"دلیل دیگر","en":"Something else","ar":"سبب آخر"}', 6)
ON CONFLICT DO NOTHING;

INSERT INTO base_chat_word (id, word, action, match) VALUES
(gen_random_uuid(), 'احمق', 'mask', 'word'),
(gen_random_uuid(), 'بی‌شعور', 'mask', 'word'),
(gen_random_uuid(), 'کثافت', 'mask', 'word'),
(gen_random_uuid(), 'خفه شو', 'mask', 'word'),
(gen_random_uuid(), 'idiot', 'mask', 'word'),
(gen_random_uuid(), 'stupid', 'mask', 'word'),
(gen_random_uuid(), 'حرامزاده', 'block', 'contains'),
(gen_random_uuid(), 'bastard', 'block', 'contains'),
(gen_random_uuid(), 'تلگرام', 'flag', 'word'),
(gen_random_uuid(), 'telegram', 'flag', 'word')
ON CONFLICT DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds player tags, restrictions and chat moderation and cannot be reverted automatically; restore from backup', '20261111000000-base-players-chat'; END $$;
-- +migrate StatementEnd
