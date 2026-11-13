-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS base_updates_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT true,
    provider TEXT NOT NULL DEFAULT 'github' CHECK (provider IN ('github')),
    repo TEXT NOT NULL CHECK (length(repo) BETWEEN 3 AND 201),
    branch TEXT NOT NULL CHECK (length(branch) BETWEEN 1 AND 200),
    interval_hours INT NOT NULL DEFAULT 6 CHECK (interval_hours BETWEEN 1 AND 168),
    token_sealed TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_updates_state (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    repo TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT '',
    base TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'never' CHECK (status IN ('never','ok','error','rate_limited','base_missing','base_mismatch','no_base')),
    error TEXT NOT NULL DEFAULT '',
    result JSONB,
    checks BIGINT NOT NULL DEFAULT 0,
    checked_at TIMESTAMPTZ,
    last_ok_at TIMESTAMPTZ,
    next_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    requested_at TIMESTAMPTZ,
    requested_by TEXT NOT NULL DEFAULT '',
    checked_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO base_updates_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % holds the parent repo setting with its sealed token and the last update check and cannot be reverted automatically; restore from backup', '20261123000000-base-updates'; END $$;
-- +migrate StatementEnd
