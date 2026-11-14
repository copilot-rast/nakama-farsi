-- +migrate Up notransaction
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS requests_bonus_meta TEXT NOT NULL DEFAULT '';
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS friends_bonus_meta TEXT NOT NULL DEFAULT '';
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS match_policy TEXT NOT NULL DEFAULT 'priority' CHECK (match_policy IN ('priority', 'generous'));
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS over_limit TEXT NOT NULL DEFAULT 'refuse' CHECK (over_limit IN ('refuse', 'trim'));

CREATE TABLE IF NOT EXISTS base_friend_rule (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
    tag_ids TEXT[] NOT NULL DEFAULT '{}',
    priority INT NOT NULL DEFAULT 0 CHECK (priority BETWEEN -1000 AND 1000),
    daily_requests INT CHECK (daily_requests BETWEEN 0 AND 1000),
    max_friends INT CHECK (max_friends BETWEEN 0 AND 10000),
    reject_cooldown_hours INT CHECK (reject_cooldown_hours BETWEEN 0 AND 8760),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS base_friend_rule_active_idx ON base_friend_rule (priority DESC, created_at) WHERE is_active;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the friend rules per player tag and cannot be reverted automatically; restore from backup', '20261124000003-base-friend-tiers'; END $$;
-- +migrate StatementEnd
