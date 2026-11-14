-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_player_report (
    id UUID PRIMARY KEY,
    reporter_id UUID REFERENCES users (id) ON DELETE SET NULL,
    reporter_username TEXT NOT NULL DEFAULT '',
    target_id UUID REFERENCES users (id) ON DELETE SET NULL,
    target_username TEXT NOT NULL DEFAULT '',
    title_id UUID REFERENCES base_report_title (id) ON DELETE SET NULL,
    title_key TEXT NOT NULL DEFAULT '',
    title JSONB NOT NULL DEFAULT '{}'::jsonb,
    note TEXT NOT NULL DEFAULT '' CHECK (length(note) <= 1000),
    match_id TEXT NOT NULL DEFAULT '' CHECK (length(match_id) <= 128),
    content_id TEXT NOT NULL DEFAULT '' CHECK (length(content_id) <= 128),
    game_key TEXT NOT NULL DEFAULT '' CHECK (length(game_key) <= 64),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'reviewing', 'resolved', 'dismissed')),
    reviewer TEXT NOT NULL DEFAULT '',
    review_started_at TIMESTAMPTZ,
    notes JSONB NOT NULL DEFAULT '[]'::jsonb,
    action TEXT NOT NULL DEFAULT '' CHECK (action IN ('', 'dismiss', 'warn', 'mute', 'suspend')),
    restriction_id UUID,
    resolved_by TEXT NOT NULL DEFAULT '',
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS base_player_report_status_idx ON base_player_report (status, created_at DESC);
CREATE INDEX IF NOT EXISTS base_player_report_target_idx ON base_player_report (target_id, created_at DESC);
CREATE INDEX IF NOT EXISTS base_player_report_reporter_idx ON base_player_report (reporter_id, created_at DESC);
CREATE INDEX IF NOT EXISTS base_player_report_pair_idx ON base_player_report (reporter_id, target_id, title_id, created_at DESC) WHERE status IN ('open', 'reviewing');

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds player reports and their decisions and cannot be reverted automatically; restore from backup', '20261124000001-base-player-reports'; END $$;
-- +migrate StatementEnd
