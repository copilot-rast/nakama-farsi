-- +migrate Up
ALTER TABLE campaign_stage ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}';
ALTER TABLE campaign_stage ADD COLUMN IF NOT EXISTS star_template TEXT NOT NULL DEFAULT '';
ALTER TABLE campaign_stage ADD COLUMN IF NOT EXISTS no_replay_after_win BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE campaign_chapter ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}';

UPDATE campaign_chapter c SET milestones = (
    SELECT jsonb_agg(x.m || jsonb_build_object('id', x.o::text) ORDER BY x.o)
    FROM jsonb_array_elements(c.milestones) WITH ORDINALITY AS x(m, o))
WHERE jsonb_typeof(c.milestones) = 'array' AND jsonb_array_length(c.milestones) > 0
    AND NOT EXISTS (SELECT 1 FROM jsonb_array_elements(c.milestones) e WHERE COALESCE(e->>'id', '') <> '');

UPDATE campaign_user_chapter u SET milestones = (
    SELECT COALESCE(jsonb_agg(COALESCE(
        (SELECT to_jsonb(m->>'id') FROM jsonb_array_elements(c.milestones) m
            WHERE jsonb_typeof(p.v) = 'number' AND COALESCE(m->>'id', '') <> '' AND abs((m->>'percent')::float8 - (p.v #>> '{}')::float8) < 0.000001 LIMIT 1),
        p.v) ORDER BY p.o), '[]'::jsonb)
    FROM jsonb_array_elements(u.milestones) WITH ORDINALITY AS p(v, o))
FROM campaign_chapter c
WHERE c.id = u.chapter_id AND jsonb_typeof(u.milestones) = 'array'
    AND EXISTS (SELECT 1 FROM jsonb_array_elements(u.milestones) e WHERE jsonb_typeof(e) = 'number');

-- +migrate Down
ALTER TABLE campaign_chapter DROP COLUMN IF EXISTS settings;
ALTER TABLE campaign_stage DROP COLUMN IF EXISTS no_replay_after_win;
ALTER TABLE campaign_stage DROP COLUMN IF EXISTS star_template;
ALTER TABLE campaign_stage DROP COLUMN IF EXISTS settings;
