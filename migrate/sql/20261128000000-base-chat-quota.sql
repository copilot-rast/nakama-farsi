-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_chat_quota_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO base_chat_quota_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS base_chat_quota_rule (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    priority INT NOT NULL DEFAULT 0 CHECK (priority BETWEEN -1000 AND 1000),
    unlimited BOOLEAN NOT NULL DEFAULT false,
    daily_messages INT CHECK (daily_messages BETWEEN 0 AND 100000),
    daily_conversations INT CHECK (daily_conversations BETWEEN 0 AND 100000),
    friend_daily_messages INT CHECK (friend_daily_messages BETWEEN 0 AND 100000),
    friend_daily_conversations INT CHECK (friend_daily_conversations BETWEEN 0 AND 100000),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS base_chat_quota_rule_active_idx ON base_chat_quota_rule (priority DESC, created_at) WHERE is_active;

CREATE TABLE IF NOT EXISTS base_chat_quota_day (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    day DATE NOT NULL,
    messages INT NOT NULL DEFAULT 0 CHECK (messages >= 0),
    conversations INT NOT NULL DEFAULT 0 CHECK (conversations >= 0),
    friend_messages INT NOT NULL DEFAULT 0 CHECK (friend_messages >= 0),
    friend_conversations INT NOT NULL DEFAULT 0 CHECK (friend_conversations >= 0),
    pack_used INT NOT NULL DEFAULT 0 CHECK (pack_used >= 0),
    free_sent INT NOT NULL DEFAULT 0 CHECK (free_sent >= 0),
    unlimited_sent INT NOT NULL DEFAULT 0 CHECK (unlimited_sent >= 0),
    total INT NOT NULL DEFAULT 0 CHECK (total >= 0),
    blocked INT NOT NULL DEFAULT 0 CHECK (blocked >= 0),
    notified BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, day)
);

CREATE INDEX IF NOT EXISTS base_chat_quota_day_day_idx ON base_chat_quota_day (day);

CREATE TABLE IF NOT EXISTS base_chat_quota_target (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    target TEXT NOT NULL,
    first_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, target)
);

CREATE TABLE IF NOT EXISTS base_chat_quota_lot (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount INT NOT NULL CHECK (amount > 0),
    remaining INT NOT NULL CHECK (remaining >= 0 AND remaining <= amount),
    source TEXT NOT NULL CHECK (source IN ('meta', 'admin')),
    note TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS base_chat_quota_lot_user_idx ON base_chat_quota_lot (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS base_chat_quota_lot_open_idx ON base_chat_quota_lot (user_id, expires_at) WHERE remaining > 0;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the private chat quota settings, rules, daily counters and bought message packs and cannot be reverted automatically; restore from backup', '20261128000000-base-chat-quota'; END $$;
-- +migrate StatementEnd
