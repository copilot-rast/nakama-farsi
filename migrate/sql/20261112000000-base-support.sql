-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_support_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    max_open_tickets INT NOT NULL DEFAULT 10 CHECK (max_open_tickets BETWEEN 1 AND 100),
    tickets_per_hour INT NOT NULL DEFAULT 6 CHECK (tickets_per_hour BETWEEN 1 AND 1000),
    messages_per_hour INT NOT NULL DEFAULT 40 CHECK (messages_per_hour BETWEEN 1 AND 10000),
    max_attachments INT NOT NULL DEFAULT 5 CHECK (max_attachments BETWEEN 0 AND 10),
    max_attachment_kb INT NOT NULL DEFAULT 2048 CHECK (max_attachment_kb BETWEEN 16 AND 10240),
    auto_close_days INT NOT NULL DEFAULT 7 CHECK (auto_close_days BETWEEN 1 AND 365),
    chat_enabled BOOLEAN NOT NULL DEFAULT true,
    chat_greeting JSONB NOT NULL DEFAULT '{}',
    chat_offline JSONB NOT NULL DEFAULT '{}',
    chat_idle_minutes INT NOT NULL DEFAULT 30 CHECK (chat_idle_minutes BETWEEN 5 AND 1440),
    chat_messages_per_minute INT NOT NULL DEFAULT 12 CHECK (chat_messages_per_minute BETWEEN 1 AND 600),
    activity_retention_days INT NOT NULL DEFAULT 180 CHECK (activity_retention_days BETWEEN 7 AND 3650),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO base_support_settings (id, chat_greeting, chat_offline) VALUES (1,
    '{"fa":"سلام! پیامتان را بنویسید؛ یکی از همکاران پشتیبانی به‌زودی جواب می‌دهد.","en":"Hi! Write your message and someone from support will answer soon.","ar":"مرحبًا! اكتب رسالتك وسيرد عليك أحد أفراد الدعم قريبًا."}',
    '{"fa":"الان کسی از پشتیبانی آنلاین نیست. پیامتان می‌ماند و اگر بخواهید، آن را به تیکت تبدیل می‌کنیم.","en":"Nobody from support is online right now. Your message stays here, and you can turn it into a ticket.","ar":"لا يوجد أحد من الدعم متصل الآن. تبقى رسالتك هنا ويمكنك تحويلها إلى تذكرة."}')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS base_support_department (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    first_response_hours INT NOT NULL DEFAULT 24 CHECK (first_response_hours BETWEEN 1 AND 720),
    agents TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_support_canned (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    title JSONB NOT NULL DEFAULT '{}',
    body JSONB NOT NULL DEFAULT '{}',
    department_id UUID REFERENCES base_support_department (id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_support_agent (
    username TEXT PRIMARY KEY,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    away BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS base_support_ticket (
    id UUID PRIMARY KEY,
    number TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    subject TEXT NOT NULL CHECK (char_length(subject) BETWEEN 1 AND 150),
    department_id UUID REFERENCES base_support_department (id) ON DELETE SET NULL,
    priority TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low','normal','high','urgent')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','awaiting_staff','awaiting_player','resolved','closed')),
    assignee TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'player' CHECK (source IN ('player','chat','staff')),
    chat_id UUID,
    lang TEXT NOT NULL DEFAULT '',
    message_count INT NOT NULL DEFAULT 0,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_message_by TEXT NOT NULL DEFAULT 'player',
    last_preview TEXT NOT NULL DEFAULT '',
    last_player_message_at TIMESTAMPTZ,
    last_staff_message_at TIMESTAMPTZ,
    player_seen_at TIMESTAMPTZ,
    due_at TIMESTAMPTZ,
    first_response_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    rating SMALLINT CHECK (rating BETWEEN 1 AND 5),
    rating_comment TEXT NOT NULL DEFAULT '',
    rated_at TIMESTAMPTZ,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS base_support_ticket_user_idx ON base_support_ticket (user_id, last_message_at DESC);
CREATE INDEX IF NOT EXISTS base_support_ticket_status_idx ON base_support_ticket (status, last_message_at DESC);
CREATE INDEX IF NOT EXISTS base_support_ticket_assignee_idx ON base_support_ticket (assignee, status);
CREATE INDEX IF NOT EXISTS base_support_ticket_due_idx ON base_support_ticket (due_at) WHERE status IN ('open','awaiting_staff');

CREATE TABLE IF NOT EXISTS base_support_message (
    id UUID PRIMARY KEY,
    ticket_id UUID NOT NULL REFERENCES base_support_ticket (id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('message','note','event')),
    author_type TEXT NOT NULL CHECK (author_type IN ('player','staff','system')),
    author TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    attachments JSONB NOT NULL DEFAULT '[]',
    event TEXT NOT NULL DEFAULT '',
    event_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX IF NOT EXISTS base_support_message_ticket_idx ON base_support_message (ticket_id, created_at);

CREATE TABLE IF NOT EXISTS base_support_upload (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    uploaded_by TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    size INT NOT NULL CHECK (size > 0),
    parts INT NOT NULL CHECK (parts BETWEEN 1 AND 200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS base_support_upload_part (
    upload_id UUID NOT NULL REFERENCES base_support_upload (id) ON DELETE CASCADE,
    idx INT NOT NULL,
    data BYTEA NOT NULL,
    PRIMARY KEY (upload_id, idx)
);

CREATE TABLE IF NOT EXISTS base_support_file (
    key TEXT PRIMARY KEY,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    uploaded_by TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size INT NOT NULL,
    attached BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS base_support_file_loose_idx ON base_support_file (created_at) WHERE NOT attached;

CREATE TABLE IF NOT EXISTS base_live_chat (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting','active','closed')),
    topic TEXT NOT NULL DEFAULT '',
    department_id UUID REFERENCES base_support_department (id) ON DELETE SET NULL,
    agent TEXT NOT NULL DEFAULT '',
    lang TEXT NOT NULL DEFAULT '',
    last_seq BIGINT NOT NULL DEFAULT 0,
    player_read_seq BIGINT NOT NULL DEFAULT 0,
    agent_read_seq BIGINT NOT NULL DEFAULT 0,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_preview TEXT NOT NULL DEFAULT '',
    player_typing_at TIMESTAMPTZ,
    agent_typing_at TIMESTAMPTZ,
    player_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    waiting_since TIMESTAMPTZ NOT NULL DEFAULT now(),
    taken_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    closed_by TEXT NOT NULL DEFAULT '',
    close_reason TEXT NOT NULL DEFAULT '',
    ticket_id UUID REFERENCES base_support_ticket (id) ON DELETE SET NULL,
    rating SMALLINT CHECK (rating BETWEEN 1 AND 5),
    rating_comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS base_live_chat_one_open_idx ON base_live_chat (user_id) WHERE status <> 'closed';
CREATE INDEX IF NOT EXISTS base_live_chat_status_idx ON base_live_chat (status, waiting_since);
CREATE INDEX IF NOT EXISTS base_live_chat_agent_idx ON base_live_chat (agent, status);
CREATE INDEX IF NOT EXISTS base_live_chat_user_idx ON base_live_chat (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS base_live_chat_message (
    chat_id UUID NOT NULL REFERENCES base_live_chat (id) ON DELETE CASCADE,
    seq BIGINT NOT NULL,
    sender TEXT NOT NULL CHECK (sender IN ('player','agent','system')),
    author TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    event TEXT NOT NULL DEFAULT '',
    event_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (chat_id, seq)
);

INSERT INTO base_support_department (id, key, name, description, first_response_hours, sort) VALUES
(gen_random_uuid(), 'general', '{"fa":"پشتیبانی عمومی","en":"General support","ar":"الدعم العام"}', '{"fa":"هر سؤال یا مشکلی دربارهٔ بازی و حساب کاربری.","en":"Any question or problem about the game or your account.","ar":"أي سؤال أو مشكلة حول اللعبة أو حسابك."}', 24, 1),
(gen_random_uuid(), 'payments', '{"fa":"خرید و پرداخت","en":"Purchases and payments","ar":"المشتريات والمدفوعات"}', '{"fa":"خریدی که سکه‌اش نرسیده، پرداخت ناموفق یا بازگشت پول.","en":"A purchase that did not arrive, a failed payment or a refund.","ar":"عملية شراء لم تصل أو دفع فاشل أو استرداد."}', 12, 2)
ON CONFLICT (key) DO NOTHING;

INSERT INTO base_support_canned (id, key, title, body, sort) VALUES
(gen_random_uuid(), 'greeting', '{"fa":"سلام و خوش‌آمد","en":"Hello and welcome","ar":"مرحبًا وأهلًا"}', '{"fa":"سلام، وقتتان بخیر. ممنون که با پشتیبانی تماس گرفتید. همین الان پیامتان را بررسی می‌کنم.","en":"Hello, thanks for contacting support. I am looking at your message right now.","ar":"مرحبًا، شكرًا لتواصلك مع الدعم. أراجع رسالتك الآن."}', 1),
(gen_random_uuid(), 'need-details', '{"fa":"درخواست جزئیات بیشتر","en":"Ask for more details","ar":"طلب مزيد من التفاصيل"}', '{"fa":"برای اینکه بهتر کمکتان کنم، لطفاً بگویید مشکل دقیقاً کِی پیش آمد و اگر می‌توانید یک تصویر از صفحه بفرستید.","en":"To help you better, please tell me exactly when this happened and, if you can, send a screenshot.","ar":"لمساعدتك بشكل أفضل، أخبرني متى حدث ذلك بالضبط وأرسل لقطة شاشة إن أمكن."}', 2),
(gen_random_uuid(), 'resolved', '{"fa":"مشکل حل شد","en":"Problem solved","ar":"تم حل المشكلة"}', '{"fa":"مشکل برطرف شد. اگر باز هم سؤالی داشتید، همین‌جا بنویسید. از صبر و همراهی‌تان ممنونیم.","en":"The problem is fixed. If you have any other question, just write here. Thank you for your patience.","ar":"تم حل المشكلة. إن كان لديك أي سؤال آخر فاكتب هنا. شكرًا لصبرك."}', 3)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds support tickets and live chats and cannot be reverted automatically; restore from backup', '20261112000000-base-support'; END $$;
-- +migrate StatementEnd
