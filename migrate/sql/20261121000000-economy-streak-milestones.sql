-- +migrate Up notransaction
ALTER TABLE economy_streak ADD COLUMN IF NOT EXISTS repay_after_fall BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS economy_streak_milestone (
    id UUID PRIMARY KEY,
    streak_id UUID NOT NULL REFERENCES economy_streak (id) ON DELETE CASCADE,
    day INT NOT NULL CHECK (day BETWEEN 1 AND 3650),
    prize_id UUID REFERENCES economy_prize (id),
    grant_mode TEXT NOT NULL DEFAULT 'claim' CHECK (grant_mode IN ('auto','claim')),
    UNIQUE (streak_id, day)
);
CREATE INDEX IF NOT EXISTS economy_streak_milestone_prize_idx ON economy_streak_milestone (prize_id) WHERE prize_id IS NOT NULL;

ALTER TABLE economy_user_streak ADD COLUMN IF NOT EXISTS peak INT NOT NULL DEFAULT 0;
ALTER TABLE economy_user_streak ADD COLUMN IF NOT EXISTS falls INT NOT NULL DEFAULT 0;
UPDATE economy_user_streak SET peak = count WHERE peak < count;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % holds streak milestones, the re-pay option and each player''s milestone floor and cannot be reverted automatically; restore from backup', '20261121000000-economy-streak-milestones'; END $$;
-- +migrate StatementEnd
