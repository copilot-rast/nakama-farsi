# Jazebe migrations on CockroachDB (historical)

> **Superseded.** The deployment now runs on PostgreSQL. CockroachDB is kept only
> as the source of a one-time data import. Configuration, deployment and the import
> are described in [database-configuration.md](database-configuration.md). This page
> remains as the record of the CockroachDB incident and of why the Jazebe
> migrations contain `notransaction` and CockroachDB constraint names. Released
> migrations stay as they are, and they run unchanged on PostgreSQL.

Production ran CockroachDB (CCL v24.1.33) until the move to PostgreSQL. The Jazebe migrations and module were
previously tested only on PostgreSQL. This page records the incident of 2026-10-25,
and the fix.

## Incident

Every start of the `nakama` container ran `nakama migrate up` and then failed:

```text
failed to exec migration statement "UPDATE jazebe_planet p SET fans_count = (...)":
ERROR: column "fans_count" does not exist (SQLSTATE 42703)
handling 20261022130000-jazebe-membership.sql
```

`restart: unless-stopped` restarted the container, and the same migration failed again.

## Root cause

`fans_count` was not missing from the migration. `20261022130000-jazebe-membership.sql`
adds it on its first line and backfills it a few lines later. sql-migrate runs each
migration in one transaction. In CockroachDB, a column added in a transaction cannot
be read or written by later statements in that transaction until it commits:

```sql
BEGIN;
ALTER TABLE p ADD COLUMN IF NOT EXISTS c INT NOT NULL DEFAULT 0;
UPDATE p SET c = 5;   -- ERROR: column "c" does not exist (SQLSTATE 42703)
```

PostgreSQL allows this, so CI and local runs passed. On CockroachDB the transaction
rolled back cleanly every time, which left the database at 21 applied migrations
(20 upstream plus `20261022120000-jazebe-brand-planet`). No partial objects remained.

Running the whole chain on CockroachDB v24.1.33 found the same class of problem in
other migrations and in the module:

| Problem on CockroachDB | Where | Fix |
|---|---|---|
| A new column is used by DML or a partial index in the same transaction (`42703`, `0A000 ... not public`, `55000 table is being added`) | membership, planet-lifecycle, brand-team, catalog-seed, brand-commercial, account-restriction and others | `-- +migrate Up notransaction`, so each statement commits before the next one runs |
| `DROP CONSTRAINT IF EXISTS x, ADD CONSTRAINT x` in one statement or transaction (`0A000 constraint in the middle of being dropped`, `42710 duplicate constraint name`) | 10 statements | Split into two statements in a `notransaction` migration |
| An inline `CHECK` is named `check_<col>` by CockroachDB but `<table>_<col>_check` by PostgreSQL, so `DROP CONSTRAINT IF EXISTS <table>_<col>_check` dropped nothing. The old check survived and rejected `profile` reports and new progression event kinds | report-scope, mission-medal-events | Also drop the CockroachDB name (a no-op on PostgreSQL) |
| `ADD COLUMN IF NOT EXISTS c ... REFERENCES/CHECK` adds the constraint again when the column exists (`..._fkey_1`) | 10 columns | Add the column without the constraint, then add a constraint explicitly named with PostgreSQL's generated name |
| `BIGSERIAL` becomes `unique_rowid()` (about 10^18), above the 2^53 a JSON client reads exactly. Chat `seq` values were corrupted in clients and moderation by `seq` failed | group-chat | Explicit `jazebe_group_message_seq_seq` sequence, identical to PostgreSQL's `BIGSERIAL` |
| `pg_advisory_xact_lock` / `pg_try_advisory_xact_lock` do not exist (`42883`). This included planet creation | 19 RPC call sites, reconcile worker | New `jazebe_xact_lock` table (`20261025000000-jazebe-xact-lock.sql`). `lockXact` upserts a row to hold a transaction-scoped lock; the worker uses `FOR UPDATE NOWAIT` |
| `make_interval(secs=>$3)` named-argument syntax | shop purchase rate limit | `$3::INT8 * INTERVAL '1 second'` |

## Compatibility strategy

- Every edited migration produces the same PostgreSQL schema as before. A
  PostgreSQL 18.6 database migrated with the old files, then upgraded, matches the
  regenerated `migrate/jazebe_schema.snapshot` exactly. The snapshot diff is only the
  new lock table. sql-migrate does not checksum files, so applied migrations are
  never re-run and `migration_info` is untouched.
- CockroachDB production never applied any Jazebe migration after
  `20261022120000`, so it receives the corrected files.
- `notransaction` migrations are not atomic. Each is re-runnable, which
  `TestJazebeMigrationsAreRerunnableAndRefuseDown` enforces, so after a failure you
  fix the cause and run `migrate up` again. CockroachDB does not make multi-DDL
  transactions atomic anyway: it can report
  `XXA00 transaction committed but schema change aborted`.
- `fans_count` is a stored counter of **active** memberships: the migration
  backfill, the reconcile worker and now member suspension use this definition. A
  first join increments it. Suspension decrements it, floored at 0, and lifting a
  suspension increments it, both in the moderation transaction. Rows removed by
  cascading user deletion are corrected by the daily reconcile worker.

## Operations

Deployment, migration status, recovery and the tests are described in
[database-configuration.md](database-configuration.md). The CockroachDB-only check
script and CI job were removed with the move to PostgreSQL; they remain in the Git
history at commit `d9b54e0c`.
