-- +migrate Up
ALTER TABLE leaderboardx_board ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL DEFAULT 'player';
ALTER TABLE leaderboardx_board DROP CONSTRAINT IF EXISTS leaderboardx_board_scope_check;
ALTER TABLE leaderboardx_board ADD CONSTRAINT leaderboardx_board_scope_check CHECK (scope IN ('player','clan'));

-- +migrate Down
ALTER TABLE leaderboardx_board DROP CONSTRAINT IF EXISTS leaderboardx_board_scope_check;
ALTER TABLE leaderboardx_board DROP COLUMN IF EXISTS scope;
