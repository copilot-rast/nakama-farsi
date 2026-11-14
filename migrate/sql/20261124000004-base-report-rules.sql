-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_report_rules (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    per_hour INT NOT NULL DEFAULT 10 CHECK (per_hour BETWEEN 0 AND 1000),
    duplicate_hours INT NOT NULL DEFAULT 24 CHECK (duplicate_hours BETWEEN 0 AND 720),
    min_account_hours INT NOT NULL DEFAULT 0 CHECK (min_account_hours BETWEEN 0 AND 720),
    note_required BOOLEAN NOT NULL DEFAULT false,
    trusted_tag_ids TEXT[] NOT NULL DEFAULT '{}',
    trusted_per_hour INT NOT NULL DEFAULT 30 CHECK (trusted_per_hour BETWEEN 0 AND 1000),
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO base_report_rules (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the player report rules and cannot be reverted automatically; restore from backup', '20261124000004-base-report-rules'; END $$;
-- +migrate StatementEnd
