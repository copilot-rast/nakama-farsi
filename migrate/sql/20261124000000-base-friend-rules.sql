-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_friend_rules (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    daily_requests INT NOT NULL DEFAULT 20 CHECK (daily_requests BETWEEN 0 AND 1000),
    max_friends INT NOT NULL DEFAULT 0 CHECK (max_friends BETWEEN 0 AND 10000),
    expiry_days INT NOT NULL DEFAULT 0 CHECK (expiry_days BETWEEN 0 AND 365),
    reject_cooldown_hours INT NOT NULL DEFAULT 0 CHECK (reject_cooldown_hours BETWEEN 0 AND 8760),
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO base_friend_rules (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS base_friend_request_day (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    day DATE NOT NULL,
    count INT NOT NULL DEFAULT 0 CHECK (count >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, day)
);

CREATE INDEX IF NOT EXISTS base_friend_request_day_day_idx ON base_friend_request_day (day);

CREATE TABLE IF NOT EXISTS base_friend_rejection (
    requester_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rejecter_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rejected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (requester_id, rejecter_id)
);

CREATE INDEX IF NOT EXISTS base_friend_rejection_at_idx ON base_friend_rejection (rejected_at);

CREATE INDEX IF NOT EXISTS base_user_edge_invite_time_idx ON user_edge (update_time) WHERE state = 1;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the friend request rules, daily counters and rejections and cannot be reverted automatically; restore from backup', '20261124000000-base-friend-rules'; END $$;
-- +migrate StatementEnd
