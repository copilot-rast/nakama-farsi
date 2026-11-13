-- +migrate Up notransaction
ALTER TABLE publication_provider DROP CONSTRAINT IF EXISTS publication_provider_channel_check;
ALTER TABLE publication_provider ADD CONSTRAINT publication_provider_channel_check CHECK (channel IN ('sms','email','push'));
INSERT INTO publication_provider (id, channel, kind, is_active, is_default, sandbox, config)
SELECT gen_random_uuid(), 'push', 'log', true, NOT EXISTS (SELECT 1 FROM publication_provider WHERE channel='push' AND is_default), true, ''
WHERE NOT EXISTS (SELECT 1 FROM publication_provider WHERE channel='push' AND kind='log');

CREATE TABLE IF NOT EXISTS publication_notify_category (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9_]{1,63}$'),
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    can_opt_out BOOLEAN NOT NULL DEFAULT true,
    default_on BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (can_opt_out OR default_on)
);
INSERT INTO publication_notify_category (id, key, name, description, can_opt_out, default_on, sort) VALUES
(gen_random_uuid(), 'system', '{"fa":"پیام‌های مهم","en":"Important messages","ar":"رسائل مهمة"}', '{"fa":"دربارهٔ حساب، پرداخت و پشتیبانی. خاموش نمی‌شود.","en":"About your account, payments and support. Cannot be turned off.","ar":"عن حسابك والدفع والدعم. لا يمكن إيقافها."}', false, true, 1),
(gen_random_uuid(), 'offers', '{"fa":"پیشنهادها و تخفیف‌ها","en":"Offers and discounts","ar":"العروض والخصومات"}', '{"fa":"بسته‌های ویژه و تخفیف‌های فروشگاه","en":"Special packs and shop discounts","ar":"حزم خاصة وخصومات المتجر"}', true, true, 2),
(gen_random_uuid(), 'rewards', '{"fa":"جایزه‌ها و دستاوردها","en":"Rewards and achievements","ar":"الجوائز والإنجازات"}', '{"fa":"وقتی جایزه یا نشانی می‌گیری","en":"When you win a prize or a badge","ar":"عندما تربح جائزة أو شارة"}', true, true, 3),
(gen_random_uuid(), 'reminders', '{"fa":"یادآوری‌ها","en":"Reminders","ar":"التذكيرات"}', '{"fa":"نوبت رایگان، زنجیرهٔ روزانه و برگشتن به بازی","en":"Free turns, daily streaks and coming back to the game","ar":"الأدوار المجانية والسلاسل اليومية والعودة إلى اللعبة"}', true, true, 4),
(gen_random_uuid(), 'news', '{"fa":"خبرها و رویدادها","en":"News and events","ar":"الأخبار والفعاليات"}', '{"fa":"رویدادهای تازه و خبرهای بازی","en":"New events and game news","ar":"فعاليات جديدة وأخبار اللعبة"}', true, true, 5)
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS publication_notify_pref (
    user_id UUID NOT NULL,
    category_id UUID NOT NULL REFERENCES publication_notify_category (id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, category_id)
);
CREATE INDEX IF NOT EXISTS publication_notify_pref_category_idx ON publication_notify_pref (category_id) WHERE NOT enabled;

CREATE TABLE IF NOT EXISTS publication_push_device (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    token TEXT NOT NULL UNIQUE CHECK (length(token) BETWEEN 16 AND 4096),
    platform TEXT NOT NULL CHECK (platform IN ('android','ios','web')),
    app_version TEXT NOT NULL DEFAULT '',
    device_key TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    fail_count INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    disabled_at TIMESTAMPTZ,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS publication_push_device_user_idx ON publication_push_device (user_id, last_seen_at DESC) WHERE is_active;

CREATE TABLE IF NOT EXISTS publication_auto_rule (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9_]{1,63}$'),
    name TEXT NOT NULL DEFAULT '',
    trigger TEXT NOT NULL CHECK (trigger IN ('event','streak_at_risk','chance_ready','inactive')),
    event_key TEXT NOT NULL DEFAULT '',
    match_field TEXT NOT NULL DEFAULT '',
    match_value TEXT NOT NULL DEFAULT '',
    params JSONB NOT NULL DEFAULT '{}',
    channel TEXT NOT NULL DEFAULT 'push_in_app' CHECK (channel IN ('push','in_app','push_in_app')),
    category_id UUID NOT NULL REFERENCES publication_notify_category (id),
    title JSONB NOT NULL DEFAULT '{}',
    body JSONB NOT NULL DEFAULT '{}',
    image_url TEXT NOT NULL DEFAULT '',
    action JSONB,
    cooldown_minutes INT NOT NULL DEFAULT 1440 CHECK (cooldown_minutes BETWEEN 0 AND 525600),
    quiet_start INT CHECK (quiet_start BETWEEN 0 AND 1439),
    quiet_end INT CHECK (quiet_end BETWEEN 0 AND 1439),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((quiet_start IS NULL) = (quiet_end IS NULL)),
    CHECK (quiet_start IS NULL OR quiet_start <> quiet_end),
    CHECK ((trigger = 'event') = (event_key <> ''))
);

CREATE TABLE IF NOT EXISTS publication_auto_fire (
    rule_id UUID NOT NULL REFERENCES publication_auto_rule (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    occurrence TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','skipped')),
    reason TEXT NOT NULL DEFAULT '',
    vars JSONB NOT NULL DEFAULT '{}',
    due_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    notification_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    done_at TIMESTAMPTZ,
    PRIMARY KEY (rule_id, user_id, occurrence)
);
CREATE INDEX IF NOT EXISTS publication_auto_fire_due_idx ON publication_auto_fire (due_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS publication_auto_fire_sent_idx ON publication_auto_fire (rule_id, user_id, done_at DESC) WHERE status = 'sent';
CREATE INDEX IF NOT EXISTS publication_auto_fire_created_idx ON publication_auto_fire (created_at);

CREATE TABLE IF NOT EXISTS publication_auto_state (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    event_cursor TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO publication_auto_state (id) VALUES (1) ON CONFLICT DO NOTHING;

ALTER TABLE publication_notification DROP CONSTRAINT IF EXISTS publication_notification_channel_check;
ALTER TABLE publication_notification ADD CONSTRAINT publication_notification_channel_check CHECK (channel IN ('in_app','sms','email','push','push_in_app'));
ALTER TABLE publication_notification DROP CONSTRAINT IF EXISTS publication_notification_status_check;
ALTER TABLE publication_notification ADD CONSTRAINT publication_notification_status_check CHECK (status IN ('scheduled','queued','sending','done','failed','cancelled'));
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'message';
ALTER TABLE publication_notification DROP CONSTRAINT IF EXISTS publication_notification_kind_check;
ALTER TABLE publication_notification ADD CONSTRAINT publication_notification_kind_check CHECK (kind IN ('message','campaign','auto','test'));
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES publication_notify_category (id);
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS audience JSONB NOT NULL DEFAULT '{}';
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ;
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS rule_id UUID REFERENCES publication_auto_rule (id) ON DELETE SET NULL;
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS cancelled_by TEXT;
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS opened_count INT NOT NULL DEFAULT 0;
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS push_sent INT NOT NULL DEFAULT 0;
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS push_failed INT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS publication_notification_scheduled_idx ON publication_notification (scheduled_at) WHERE status = 'scheduled';
CREATE INDEX IF NOT EXISTS publication_notification_rule_idx ON publication_notification (rule_id, created_at DESC) WHERE rule_id IS NOT NULL;

ALTER TABLE publication_notification_recipient ADD COLUMN IF NOT EXISTS vars JSONB NOT NULL DEFAULT '{}';
ALTER TABLE publication_notification_recipient ADD COLUMN IF NOT EXISTS opened_at TIMESTAMPTZ;
ALTER TABLE publication_notification_recipient ADD COLUMN IF NOT EXISTS push_sent INT NOT NULL DEFAULT 0;
ALTER TABLE publication_notification_recipient ADD COLUMN IF NOT EXISTS push_failed INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS publication_push_delivery (
    notification_id UUID NOT NULL REFERENCES publication_notification (id) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('sending','sent','failed','dead')),
    error TEXT NOT NULL DEFAULT '',
    provider_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (notification_id, device_id)
);

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % holds push devices, notification categories, player choices and automatic notification history and cannot be reverted automatically; restore from backup', '20261120000000-publication-push'; END $$;
-- +migrate StatementEnd
