-- +migrate Up
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS reroll_cap INT CHECK (reroll_cap IS NULL OR reroll_cap BETWEEN 1 AND 100000);
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS free_reset TEXT NOT NULL DEFAULT 'interval' CHECK (free_reset IN ('interval','daily'));
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS free_first BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS day_zone TEXT NOT NULL DEFAULT '';
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS day_start_minute INT NOT NULL DEFAULT 0 CHECK (day_start_minute BETWEEN 0 AND 1439);
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS min_level INT NOT NULL DEFAULT 0 CHECK (min_level BETWEEN 0 AND 100000);
ALTER TABLE economy_chance_game ADD COLUMN IF NOT EXISTS level_table_id UUID;

ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS day_zone TEXT NOT NULL DEFAULT '';
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS day_start_minute INT NOT NULL DEFAULT 0 CHECK (day_start_minute BETWEEN 0 AND 1439);
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS signup_delay_days INT NOT NULL DEFAULT 0 CHECK (signup_delay_days BETWEEN 0 AND 365);
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS claim_window TEXT NOT NULL DEFAULT 'open' CHECK (claim_window IN ('open','period'));
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS infinity_mode TEXT NOT NULL DEFAULT 'pay' CHECK (infinity_mode IN ('pay','claim'));
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS infinity_prize_id UUID REFERENCES economy_prize (id);

ALTER TABLE economy_reward_claim ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

ALTER TABLE economy_offer DROP CONSTRAINT IF EXISTS economy_offer_trigger_type_check;
ALTER TABLE economy_offer ADD CONSTRAINT economy_offer_trigger_type_check CHECK (trigger_type IN ('first_entry','level','meta','tag','account_days','lapsed','game'));

INSERT INTO economy_event_type (id, key, name, description, client_allowed, value_min, rate_limit, is_system, sort) VALUES
(gen_random_uuid(), 'online', '{"fa":"آنلاین بودن","en":"Online","ar":"متصل"}', '{"fa":"وقتی بازیکن به سرور وصل می‌شود و هر بار که روز تازه‌ای شروع می‌شود و بازیکن هنوز وصل است، سرور این رویداد را ثبت می‌کند.","en":"Recorded by the server when a player connects, and again when a new day starts while the player is still connected.","ar":"يسجّله الخادم عند اتصال اللاعب، ومرة أخرى عند بدء يوم جديد واللاعب ما زال متصلاً."}', false, NULL, 30, true, 5)
ON CONFLICT (key) DO NOTHING;

ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS day_zone TEXT NOT NULL DEFAULT '';
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS day_start_minute INT NOT NULL DEFAULT 0 CHECK (day_start_minute BETWEEN 0 AND 1439);
ALTER TABLE base_friend_rules ADD COLUMN IF NOT EXISTS bot_decline_hours INT NOT NULL DEFAULT 0 CHECK (bot_decline_hours BETWEEN 0 AND 8760);

ALTER TABLE publication_otp_settings ADD COLUMN IF NOT EXISTS user_hourly INT NOT NULL DEFAULT 0 CHECK (user_hourly BETWEEN 0 AND 10000);

-- +migrate Down
ALTER TABLE publication_otp_settings DROP COLUMN IF EXISTS user_hourly;
ALTER TABLE base_friend_rules DROP COLUMN IF EXISTS bot_decline_hours, DROP COLUMN IF EXISTS day_start_minute, DROP COLUMN IF EXISTS day_zone;
ALTER TABLE economy_reward_claim DROP COLUMN IF EXISTS expires_at;
UPDATE economy_offer SET trigger_type = 'first_entry', trigger_ref = '', trigger_value = NULL, is_active = false WHERE trigger_type = 'game';
ALTER TABLE economy_offer DROP CONSTRAINT IF EXISTS economy_offer_trigger_type_check;
ALTER TABLE economy_offer ADD CONSTRAINT economy_offer_trigger_type_check CHECK (trigger_type IN ('first_entry','level','meta','tag','account_days','lapsed'));
ALTER TABLE economy_streak DROP COLUMN IF EXISTS infinity_prize_id, DROP COLUMN IF EXISTS infinity_mode, DROP COLUMN IF EXISTS claim_window, DROP COLUMN IF EXISTS signup_delay_days, DROP COLUMN IF EXISTS day_start_minute, DROP COLUMN IF EXISTS day_zone;
ALTER TABLE economy_chance_game DROP COLUMN IF EXISTS level_table_id, DROP COLUMN IF EXISTS min_level, DROP COLUMN IF EXISTS day_start_minute, DROP COLUMN IF EXISTS day_zone, DROP COLUMN IF EXISTS free_first, DROP COLUMN IF EXISTS free_reset, DROP COLUMN IF EXISTS reroll_cap;
