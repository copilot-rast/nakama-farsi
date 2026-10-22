-- Group mission completions are distinct from each contributor's completion.
-- +migrate Up
ALTER TABLE jazebe_mission_completion ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE jazebe_mission_completion ADD COLUMN group_id UUID REFERENCES jazebe_group(id);
ALTER TABLE jazebe_mission_completion ADD COLUMN awarded_tickets INTEGER NOT NULL DEFAULT 0 CHECK (awarded_tickets >= 0);
ALTER TABLE jazebe_mission_completion ADD CONSTRAINT jazebe_completion_has_owner CHECK (user_id IS NOT NULL OR group_id IS NOT NULL);
CREATE UNIQUE INDEX uq_jazebe_group_mission_completion ON jazebe_mission_completion(mission_id,group_id) WHERE user_id IS NULL AND group_id IS NOT NULL AND status='approved' AND deleted_at IS NULL;
CREATE INDEX idx_jazebe_mission_completion_group ON jazebe_mission_completion(group_id,mission_id,status) WHERE group_id IS NOT NULL;
CREATE TABLE jazebe_group_mission_progress (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    mission_id UUID NOT NULL REFERENCES jazebe_mission(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    current_value INTEGER NOT NULL DEFAULT 0 CHECK (current_value >= 0),
    target_value INTEGER NOT NULL CHECK (target_value > 0),
    status VARCHAR(12) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress','completed','failed')),
    completed_at TIMESTAMPTZ,
    UNIQUE (mission_id,group_id)
);
CREATE INDEX idx_jazebe_group_mission_progress_group ON jazebe_group_mission_progress(group_id,status);
CREATE TABLE jazebe_group_mission_contribution (
    progress_id UUID NOT NULL REFERENCES jazebe_group_mission_progress(id),
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    user_id UUID NOT NULL REFERENCES users(id),
    value INTEGER NOT NULL DEFAULT 0 CHECK (value >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (progress_id,user_id)
);
CREATE INDEX idx_jazebe_group_mission_contribution_user ON jazebe_group_mission_contribution(user_id,progress_id);

-- +migrate Down
-- Group completion and contribution records preserve earned value and require manual rollback.
SELECT 1;
