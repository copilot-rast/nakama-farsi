-- Mission medals are processed after the approved completion commits.
-- +migrate Up
ALTER TABLE jazebe_progression_event DROP CONSTRAINT jazebe_progression_event_kind_check;
ALTER TABLE jazebe_progression_event ADD CONSTRAINT jazebe_progression_event_kind_check
    CHECK (kind IN ('purchase','group_join','mission'));
-- +migrate Down
-- Retain pending mission events; restoring the narrower check would discard work.
SELECT 1;
