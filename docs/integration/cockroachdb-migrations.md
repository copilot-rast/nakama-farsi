# Jazebe migrations on CockroachDB

Production runs CockroachDB (CCL v24.1.33). The Jazebe migrations and module were
previously tested only on PostgreSQL. This page records the incident of 2026-10-25,
the fix and how to operate migrations safely.

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

## Automatic migrations in Docker Compose (Dokploy)

`docker-compose.yml` has a one-shot `nakama-migrate` service:

1. `cockroachdb` must be healthy (`/health?ready=1`, with `start_period: 120s` for slow
   starts).
2. `nakama-migrate` runs `nakama migrate up --database.address root@cockroachdb:26257`
   once, with `restart: "no"`.
3. `nakama` starts only after `nakama-migrate` exits 0
   (`service_completed_successfully`). The server also refuses to start while
   migrations are pending ("DB schema outdated").
4. A failed migration makes `docker compose up` exit non-zero with
   `service "nakama-migrate" didn't complete successfully: exit 1`. The server is not
   started or recreated, and nothing restarts in a loop.
5. Restarting `nakama`, or a host reboot, does not re-run migrations. Every deployment
   re-runs the job, which applies 0 when nothing is pending.
6. One job per deployment means replicas never migrate concurrently. If two migrators
   still race, the loser fails on `migration_info_pkey` and the winner completes. This
   was verified twice on CockroachDB.

The server entrypoint now uses `sh -ec` instead of `sh -ecx`. `-x` printed the
expanded server key, console password and signing key into container logs.

Migrations on CockroachDB are slow. Upgrading from the production state took about
3 minutes with the Compose limit of 1 CPU and 2 GB.

## Inspecting migration status (read-only)

```sh
docker compose logs nakama-migrate
docker compose run --rm --no-deps --entrypoint /nakama/nakama nakama-migrate migrate status --database.address root@cockroachdb:26257
docker compose exec cockroachdb cockroach sql --insecure -d nakama -e "SELECT count(*) FROM migration_info"
```

## Recovering from a failed deployment

1. Read `docker compose logs nakama-migrate`. The last line names the migration and
   SQL error.
2. Do not drop, truncate or reset the database. Do not delete `migration_info` rows.
3. Fix the migration in a new commit and deploy again. The job re-runs, and a
   re-runnable migration completes whatever an earlier attempt left.
4. Take a backup before any manual repair:
   `cockroach sql --insecure -e "BACKUP DATABASE nakama INTO 'nodelocal://1/<name>'"`.
   `scripts/jazebe-rollback.sql` is PostgreSQL-only (psql, `DO`).
5. On CockroachDB 24.1 a refused Jazebe `migrate down` fails with a `DO` syntax error
   instead of the explanatory message (no `DO` support). It is still refused and
   `migration_info` is unchanged.

## Reproducing the tests

Build the binary (Go 1.27.1, or the `golang:1.27.1-bookworm` image):

```sh
go build -mod=vendor -trimpath -o /tmp/nakama .
scripts/jazebe-cockroach-migration-check.sh /tmp/nakama
```

The script starts a throwaway `cockroachdb/cockroach:v24.1.33` container and checks two
scenarios:
- **Fresh install:** 60 migrations, then 0, and `fans_count` is `bigint NOT NULL DEFAULT 0`.
- **Production state:** `migrate up --limit 21` plus planets with 0, 1 and 3 members,
  then an upgrade (39 migrations, `fans_count` 0/1/3, rows intact), then a no-op rerun.

The pre-fix binary fails this script at `20261022130000`.

Run the module's database tests against a migrated CockroachDB database:

```sh
docker run -d --name crdb -p 127.0.0.1:26257:26257 cockroachdb/cockroach:v24.1.33 start-single-node --insecure
docker exec crdb cockroach sql --insecure -e 'CREATE DATABASE jazebe_test_disposable'
/tmp/nakama migrate up --database.address root@127.0.0.1:26257/jazebe_test_disposable
JAZEBE_DISPOSABLE_TEST_DB_URL='postgresql://root@127.0.0.1:26257/jazebe_test_disposable?sslmode=disable' \
  go test -mod=vendor ./integrations/jazebe/... -run Postgres -count=1
```

The `profile_cockroachdb` job in `.github/workflows/jazebe-migrations.yaml` runs both.
The PostgreSQL schema-snapshot tests skip on CockroachDB because they read PostgreSQL
catalog text.

## Known limits

- CockroachDB runs `SERIALIZABLE` and can return `40001` restart errors under
  contention where PostgreSQL would wait. The module has no automatic transaction
  retry, so a contended RPC may fail with `Internal`. The transaction rolls back
  completely, and the client can retry. `TestAccountTransitionConcurrentPostgres`
  showed this once in about 30 runs.
- `jazebe_point_transaction.seq` still uses `unique_rowid()` on CockroachDB. It is
  returned to clients as a string, so precision is preserved.
