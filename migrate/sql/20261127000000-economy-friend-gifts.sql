-- +migrate Up notransaction
-- Per item and per product gift switch; NULL follows the item-type and project defaults.
ALTER TABLE economy_item ADD COLUMN IF NOT EXISTS giftable BOOLEAN;
ALTER TABLE economy_product ADD COLUMN IF NOT EXISTS giftable BOOLEAN;

ALTER TABLE economy_user_item DROP CONSTRAINT IF EXISTS economy_user_item_source_check;
ALTER TABLE economy_user_item ADD CONSTRAINT economy_user_item_source_check CHECK (source IN ('starter','buy','unlock','prize','admin','gift'));

ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak','referral','task','achievement','chance','game','relation','friend_gift'));

CREATE TABLE IF NOT EXISTS economy_friend_gift_settings (
    id INT PRIMARY KEY CHECK (id = 1),
    settings JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Gift limits per player tag; NULL keeps the project value, 0 means no limit.
CREATE TABLE IF NOT EXISTS economy_friend_gift_rule (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    priority INT NOT NULL DEFAULT 0 CHECK (priority BETWEEN -1000 AND 1000),
    per_sender_daily INT CHECK (per_sender_daily BETWEEN 0 AND 10000),
    per_recipient_daily INT CHECK (per_recipient_daily BETWEEN 0 AND 10000),
    per_pair_daily INT CHECK (per_pair_daily BETWEEN 0 AND 10000),
    can_send BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS economy_friend_gift_rule_active_idx ON economy_friend_gift_rule (priority DESC, created_at) WHERE is_active;

-- One gift from a player to another: an inventory item or currency, or a shop product the sender paid for.
-- It is also the audit trail: every state change stays on the row.
CREATE TABLE IF NOT EXISTS economy_friend_gift (
    id UUID PRIMARY KEY,
    sender_id UUID NOT NULL,
    recipient_id UUID NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('inventory','shop')),
    kind TEXT NOT NULL CHECK (kind IN ('item','currency','product')),
    item_id UUID REFERENCES economy_item (id),
    item_key TEXT NOT NULL DEFAULT '',
    currency_key TEXT NOT NULL DEFAULT '',
    amount BIGINT NOT NULL DEFAULT 1 CHECK (amount >= 1),
    product_id UUID REFERENCES economy_product (id),
    prize_id UUID REFERENCES economy_prize (id),
    payment_id UUID,
    method TEXT NOT NULL DEFAULT '',
    item_state JSONB NOT NULL DEFAULT '{}',
    message TEXT NOT NULL DEFAULT '',
    message_flagged BOOLEAN NOT NULL DEFAULT false,
    delivery TEXT NOT NULL CHECK (delivery IN ('instant','accept')),
    status TEXT NOT NULL CHECK (status IN ('sending','paying','pending','delivered','declined','expired','cancelled','reversed','failed')),
    outcome TEXT NOT NULL DEFAULT '',
    error TEXT,
    request_id TEXT NOT NULL DEFAULT '',
    day DATE NOT NULL,
    expires_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    responded_at TIMESTAMPTZ,
    reversed_at TIMESTAMPTZ,
    reversed_by TEXT NOT NULL DEFAULT '',
    reverse_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (sender_id <> recipient_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS economy_friend_gift_request_idx ON economy_friend_gift (sender_id, request_id) WHERE request_id <> '';
CREATE INDEX IF NOT EXISTS economy_friend_gift_sender_idx ON economy_friend_gift (sender_id, day);
CREATE INDEX IF NOT EXISTS economy_friend_gift_recipient_idx ON economy_friend_gift (recipient_id, day);
CREATE INDEX IF NOT EXISTS economy_friend_gift_created_idx ON economy_friend_gift (created_at DESC);
CREATE INDEX IF NOT EXISTS economy_friend_gift_pending_idx ON economy_friend_gift (expires_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS economy_friend_gift_payment_idx ON economy_friend_gift (payment_id) WHERE payment_id IS NOT NULL;

ALTER TABLE economy_payment ADD COLUMN IF NOT EXISTS friend_gift_id UUID REFERENCES economy_friend_gift (id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS economy_payment_friend_gift_idx ON economy_payment (friend_gift_id) WHERE friend_gift_id IS NOT NULL;

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'Gift:Sent', '{"fa":"فرستادن هدیه به دوست","en":"Gift sent","ar":"إرسال هدية"}', '{"fa":"وقتی بازیکن به دوستش هدیه می‌دهد (از کوله یا فروشگاه). نوع هدیه در kind و کلید آن در key می‌آید.","en":"Recorded when a player sends a friend a gift from the inventory or the shop. The gift kind is in kind and its key in key.","ar":"يُسجَّل عندما يرسل اللاعب هدية إلى صديقه."}', false, 0, 600, true, 14),
(gen_random_uuid(), 'Gift:Received', '{"fa":"گرفتن هدیه از دوست","en":"Gift received","ar":"استلام هدية"}', '{"fa":"وقتی هدیهٔ یک دوست به بازیکن می‌رسد. نوع هدیه در kind و کلید آن در key می‌آید.","en":"Recorded when a friend''s gift reaches the player. The gift kind is in kind and its key in key.","ar":"يُسجَّل عندما تصل هدية صديق إلى اللاعب."}', false, 0, 600, true, 15)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds friend gifts and their audit trail and cannot be reverted automatically; restore from backup', '20261127000000-economy-friend-gifts'; END $$;
-- +migrate StatementEnd
