-- +migrate Up
ALTER TABLE economy_game ADD COLUMN IF NOT EXISTS stealth JSONB NOT NULL DEFAULT '{}';

CREATE TABLE IF NOT EXISTS economy_bot_lease (
    user_id UUID PRIMARY KEY,
    match_id TEXT NOT NULL DEFAULT '',
    game_key TEXT NOT NULL DEFAULT '',
    busy BOOLEAN NOT NULL DEFAULT false,
    leased_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    online_until TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS economy_bot_lease_match_idx ON economy_bot_lease (match_id) WHERE busy;

CREATE TABLE IF NOT EXISTS economy_bot_guard (
    user_id UUID NOT NULL,
    game_key TEXT NOT NULL,
    day DATE NOT NULL,
    wins INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, game_key, day)
);

CREATE INDEX IF NOT EXISTS economy_game_entry_bot_idx ON economy_game_entry (user_id) WHERE bot AND status IN ('paid','free');

ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS bot_decline_min_minutes INT NOT NULL DEFAULT 0;
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS bot_decline_max_minutes INT NOT NULL DEFAULT 0;

UPDATE users u SET metadata = (u.metadata - 'bot') - CASE WHEN u.metadata->>'source' = 'legacy-pool' THEN 'source' ELSE '' END
WHERE u.id IN (SELECT user_id FROM economy_bot) AND (u.metadata ? 'bot' OR u.metadata->>'source' = 'legacy-pool');

UPDATE users u SET create_time = u.create_time - make_interval(days => 20 + (abs(hashtext(u.id::text)) % 400), secs => abs(hashtext(u.username)) % 86400)
WHERE u.id IN (SELECT user_id FROM economy_bot) AND u.create_time > now() - interval '30 days';

-- +migrate Down
UPDATE users u SET metadata = u.metadata || '{"bot":true}' WHERE u.id IN (SELECT user_id FROM economy_bot);
ALTER TABLE base_friend_rules DROP COLUMN IF EXISTS bot_decline_max_minutes;
ALTER TABLE base_friend_rules DROP COLUMN IF EXISTS bot_decline_min_minutes;
DROP INDEX IF EXISTS economy_game_entry_bot_idx;
DROP TABLE IF EXISTS economy_bot_guard;
DROP TABLE IF EXISTS economy_bot_lease;
ALTER TABLE economy_game DROP COLUMN IF EXISTS stealth;
