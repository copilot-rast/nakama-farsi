-- +migrate Up
-- Unity package registry (npm protocol at /upm): one row per published version; the archive is
-- in the object store under upm/<name>/<name>-<version>.tgz.
CREATE TABLE IF NOT EXISTS devkit_package_version (
    name TEXT NOT NULL CHECK (name ~ '^[a-z0-9][a-z0-9._-]{0,213}$'),
    version TEXT NOT NULL CHECK (char_length(version) BETWEEN 5 AND 64),
    manifest JSONB NOT NULL DEFAULT '{}',
    shasum TEXT NOT NULL,
    integrity TEXT NOT NULL,
    size BIGINT NOT NULL,
    storage_key TEXT NOT NULL,
    published_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (name, version)
);

-- +migrate Down
DROP TABLE IF EXISTS devkit_package_version;
