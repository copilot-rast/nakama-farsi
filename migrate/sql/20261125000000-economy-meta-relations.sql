-- +migrate Up notransaction
ALTER TABLE economy_reward_claim DROP CONSTRAINT IF EXISTS economy_reward_claim_kind_check;
ALTER TABLE economy_reward_claim ADD CONSTRAINT economy_reward_claim_kind_check CHECK (kind IN ('event','level','streak','referral','task','achievement','chance','relation'));
ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check;
ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN ('payment','gift','admin','event','level','streak','referral','task','achievement','chance','game','relation'));

CREATE TABLE IF NOT EXISTS economy_meta_relation (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',
    description JSONB NOT NULL DEFAULT '{}',
    meta_id UUID NOT NULL REFERENCES economy_meta (id),
    operator TEXT NOT NULL CHECK (operator IN ('gt','gte','lt','lte','eq')),
    value NUMERIC(20,4) NOT NULL,
    targets JSONB NOT NULL DEFAULT '[]',
    prize_type TEXT NOT NULL DEFAULT 'none' CHECK (prize_type IN ('none','prize','prizeGroup')),
    prize_id UUID REFERENCES economy_prize (id),
    prize_group_id UUID REFERENCES economy_prize_group (id),
    grant_mode TEXT NOT NULL DEFAULT 'auto' CHECK (grant_mode IN ('auto','claim')),
    repeat TEXT NOT NULL DEFAULT 'once' CHECK (repeat IN ('once','rearm','every')),
    max_fires INT CHECK (max_fires IS NULL OR max_fires BETWEEN 1 AND 1000000),
    audience_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    client_visible BOOLEAN NOT NULL DEFAULT true,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
    CHECK ((prize_type = 'none' AND prize_id IS NULL AND prize_group_id IS NULL)
        OR (prize_type = 'prize' AND prize_id IS NOT NULL AND prize_group_id IS NULL)
        OR (prize_type = 'prizeGroup' AND prize_group_id IS NOT NULL AND prize_id IS NULL))
);
CREATE INDEX IF NOT EXISTS economy_meta_relation_meta_idx ON economy_meta_relation (meta_id) WHERE is_active;

CREATE TABLE IF NOT EXISTS economy_user_meta_relation (
    relation_id UUID NOT NULL REFERENCES economy_meta_relation (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    fires INT NOT NULL DEFAULT 0,
    armed BOOLEAN NOT NULL DEFAULT true,
    first_done_at TIMESTAMPTZ,
    last_done_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (relation_id, user_id)
);
CREATE INDEX IF NOT EXISTS economy_user_meta_relation_user_idx ON economy_user_meta_relation (user_id, last_done_at DESC) WHERE fires > 0;

ALTER TABLE economy_event_settings ADD COLUMN IF NOT EXISTS relation_chain_depth INT NOT NULL DEFAULT 8 CHECK (relation_chain_depth BETWEEN 0 AND 25);
ALTER TABLE economy_event_settings ADD COLUMN IF NOT EXISTS relation_prize_missing TEXT NOT NULL DEFAULT 'complete' CHECK (relation_prize_missing IN ('complete','wait'));

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'MetaRelation:Completed', '{"fa":"رابطهٔ متا انجام شد","en":"Meta relation completed","ar":"اكتملت علاقة المقياس"}', '{"fa":"هر بار که یک رابطهٔ متا برای بازیکن انجام می‌شود. شناسهٔ رابطه در entityId و کلید آن در key می‌آید.","en":"Recorded each time a meta relation fires for a player. The relation id is in entityId and its key in key.","ar":"يُسجَّل في كل مرة تُنفَّذ فيها علاقة مقياس للاعب. معرّف العلاقة في entityId ومفتاحها في key."}', false, 0, 600, true, 12)
ON CONFLICT (key) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds meta relations and player completions and cannot be reverted automatically; restore from backup', '20261125000000-economy-meta-relations'; END $$;
-- +migrate StatementEnd
