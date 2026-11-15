-- +migrate Up notransaction
-- Entry requirements for a whole game (rooms keep theirs inside economy_game.rooms[].requirements).
ALTER TABLE economy_game ADD COLUMN IF NOT EXISTS requirements JSONB NOT NULL DEFAULT '{}'::jsonb;

-- How the player entered (matchmaking, bot, friend) and the tickets taken on entry, so a leave before
-- the start or a cancelled match can give them back.
ALTER TABLE economy_game_entry ADD COLUMN IF NOT EXISTS path TEXT NOT NULL DEFAULT 'matchmaking';
ALTER TABLE economy_game_entry ADD COLUMN IF NOT EXISTS tickets JSONB;
ALTER TABLE economy_game_entry DROP CONSTRAINT IF EXISTS economy_game_entry_path_check;
ALTER TABLE economy_game_entry ADD CONSTRAINT economy_game_entry_path_check CHECK (path IN ('matchmaking','bot','friend'));

-- A ticket item that comes back after a leave keeps the source it had; 'game' is for rows a game hands back without one.
ALTER TABLE economy_user_item DROP CONSTRAINT IF EXISTS economy_user_item_source_check;
ALTER TABLE economy_user_item ADD CONSTRAINT economy_user_item_source_check CHECK (source IN ('starter','buy','unlock','prize','admin','gift','game'));

-- +migrate Down
ALTER TABLE economy_game_entry DROP CONSTRAINT IF EXISTS economy_game_entry_path_check;
ALTER TABLE economy_game_entry DROP COLUMN IF EXISTS tickets;
ALTER TABLE economy_game_entry DROP COLUMN IF EXISTS path;
ALTER TABLE economy_game DROP COLUMN IF EXISTS requirements;
