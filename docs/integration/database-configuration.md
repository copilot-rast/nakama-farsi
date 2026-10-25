# Database: PostgreSQL configuration, deployment and the CockroachDB import

Nakama Farsi runs on **PostgreSQL 16** (`postgres:16.8-alpine`, the version CI tests).
PostgreSQL is the only database the Nakama server, the migration job and the Jazebe
module connect to. The module uses Nakama's database handle and opens no connection
of its own.

Production ran CockroachDB until this change. Its data is copied into PostgreSQL
once, by the migration job, and CockroachDB is not used after that. The background
is in [cockroachdb-migrations.md](cockroachdb-migrations.md).

## How the connection is configured

`docker-compose.yml` passes the same `DB_*` variables to `nakama-migrate` and
`nakama` through one block (`x-database-env`). Both services start through
`/nakama/nakama-db.sh` (`build/nakama-db.sh`), which:

1. builds the Nakama address `user:password@host:port/name?sslmode=…`, with the user
   and password URL-encoded, or takes `DB_URL`;
2. writes it to a temporary config file readable only by the container user and
   passes `--config <file>` to Nakama. The password is never on a command line, in the
   container's stored command or in a log. Logs show `nakama:xxxxx@postgres:5432/nakama`;
3. waits at most `DB_WAIT_TIMEOUT_SECONDS` for the host to resolve and accept TCP
   connections, and reports DNS and TCP failures apart;
4. for `migrate`: performs the CockroachDB import if needed, then
   `nakama migrate up`. For `server`: `exec nakama --config <file> <flags>`.

The bundled `postgres` service is created from the same `DB_NAME`, `DB_USER` and
`DB_PASSWORD`, so one set of values in Dokploy sets up the database and the
connection to it.

### Variables

| Variable | Default | Meaning |
|---|---|---|
| `DB_HOST` | `postgres` | PostgreSQL host as seen from the containers |
| `DB_PORT` | `5432` | Port |
| `DB_NAME` | `nakama` | Database. Letters, digits and `_` |
| `DB_USER` | `nakama` | Role. Also creates the bundled database's owner |
| `DB_PASSWORD` | **required** | Password. Any characters. `compose up` stops with `set DB_PASSWORD in the Dokploy Environment tab` if it is missing |
| `DB_SSL_MODE` | empty: Nakama's `prefer` | `disable`, `require`, `verify-ca` or `verify-full` |
| `DB_SSL_ROOT_CERT` | empty | CA certificate path inside the containers, for `verify-ca`/`verify-full` |
| `DB_URL` | empty | Complete address, which **replaces all the fields above** for the two Nakama services. The bundled `postgres` still uses `DB_NAME`/`DB_USER`/`DB_PASSWORD`. Credentials must be URL-encoded (`@` → `%40`) |
| `DB_WAIT_TIMEOUT_SECONDS` | `120` | Maximum wait for DNS and TCP readiness |
| `DB_MIGRATIONS_ENABLED` | `true` | `false` skips `migrate up`. The server then refuses to start while migrations are pending |
| `DB_IMPORT_FROM_COCKROACH` | `true` in compose, `false` in the script | Copy the legacy CockroachDB data once. Ignored after a completed import. A new installation sets `false` |
| `COCKROACH_IMPORT_URL` | `postgresql://root@cockroachdb:26257/nakama?sslmode=disable` | Import source |
| `COMPOSE_PROFILES` | empty | `cockroach-import` makes Compose manage the legacy `cockroachdb` container (see below) |

There is no migration run timeout. Interrupting a running migration can leave a
`notransaction` migration half applied, so a slow migration is allowed to finish.
`.env.example` lists every variable, with placeholders.

### Exit codes of the migration job

| Exit | Meaning | Look at |
|---|---|---|
| 2 | Missing or invalid setting | the `nakama-db: ERROR` line |
| 3 | DNS: the host does not resolve | `DB_HOST`, and whether the services share a Docker network |
| 4 | TCP: the port refuses connections | `DB_PORT`, whether the database runs |
| 5 | PostgreSQL already holds data and no import completed; the import was refused | see "Importing the CockroachDB data" |
| 6 | Import source problem (no migrations, different schema) | the `ERROR` line |
| 1 | Nakama or psql failed. The last line names the class: authentication, TLS, SQL migration, missing database, DNS or connection | the line before it |

## Dokploy: deploying

1. Open the compose application, then **Environment**. Add at least:
   ```
   DB_PASSWORD=<a long random password>
   ```
   Optionally set `DB_USER` and `DB_NAME`. The defaults are `nakama` and `nakama`.
   Keep the existing `NAKAMA_*` values.
2. Click **Deploy**. Expected order: `postgres` becomes healthy, `nakama-migrate`
   runs and exits 0, then `nakama` starts.
3. Open **Logs**, select `nakama-migrate`, and check for:
   ```
   nakama-db: CockroachDB data imported at migration level 21; the CockroachDB database is unchanged
   ... "msg":"Successfully applied migration","count":39
   nakama-db: migrations complete
   ```
   On later deployments it says `CockroachDB data was imported at …; not importing again`
   and `"count":0`.
4. Select `nakama` and check `"Database information","version":"PostgreSQL 16.8 …"`.
   That line proves the server runs on PostgreSQL.

Exited containers such as `nakama-migrate` are also listed, with their logs, on
Dokploy's **Docker** page.

### Importing the CockroachDB data

The legacy `cockroachdb` service has the profile `cockroach-import`. Compose does not
start, recreate or remove a profile-disabled service, and keeps its `data` volume. So
the production container keeps running on its data after the deployment, and the job
reads from it.

On the first deployment, with `DB_IMPORT_FROM_COCKROACH=true`, the job:

1. checks PostgreSQL for a completed import (`nakama_farsi.cockroach_import`). If one
   exists, the source is not contacted;
2. waits for CockroachDB and reads its `migration_info` (production: 21 migrations);
3. if PostgreSQL has no migrations, runs `nakama migrate up --limit 21` and checks
   that the IDs match the source. If PostgreSQL already has other migrations or data,
   it stops with exit 5 and changes nothing;
4. exports every table at that schema level with `COPY`. CockroachDB's hidden
   columns are excluded, and the column sets must match;
5. loads all tables in **one PostgreSQL transaction**. It deletes the rows the
   migrations seeded (such as the system user), copies the data with foreign-key
   triggers deferred (`session_replication_role = replica`), checks every table's
   row count, resets sequences, and records the import. Any error rolls back the
   whole load;
6. continues with the normal `migrate up`. The remaining 39 migrations run on the
   real data, including the `fans_count` backfill.

CockroachDB is only read. Its database, including `migration_info`, stays as it was.
Stop writes to it during the import. The Nakama server is already switched to
PostgreSQL, so nothing else writes to it.

The import needs a PostgreSQL superuser for `session_replication_role`. The bundled
`postgres` role (`DB_USER`) is one. A managed PostgreSQL whose role is not a superuser
cannot take the import; import into the bundled one first, then move the data with
`pg_dump`.

**If the import reports a DNS error for `cockroachdb`** (exit 3), the legacy
container is not on the deployment's network. Add `COMPOSE_PROFILES=cockroach-import`
in Environment and deploy once. Compose then recreates `cockroachdb` on its existing
`data` volume, on the current network. Remove the variable afterwards. Do not delete
the volume.

**Check the result** with read-only queries. Use Dokploy's terminal for the
container, or `docker exec`:

```sh
# PostgreSQL (postgres container)
psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT * FROM nakama_farsi.cockroach_import"
psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT count(*) FROM users"
# CockroachDB (cockroachdb container)
cockroach sql --insecure -d nakama -e "SELECT count(*) FROM users"
```

`users` on PostgreSQL equals CockroachDB's count, because the system user is
replaced, not added.

**After the PostgreSQL deployment has been verified** and a backup taken, CockroachDB
can be retired: stop the `cockroachdb` container in Dokploy's Docker page. Delete its
container and the `<project>_data` volume only when the data is no longer needed.
A later commit can then remove the `cockroachdb` service and `data` volume from
`docker-compose.yml`. Until then they stay.

### A new installation

Set `DB_IMPORT_FROM_COCKROACH=false` and `DB_PASSWORD`. The job creates the schema
(all migrations) in the bundled PostgreSQL.

## Using a PostgreSQL outside this compose file

Set `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER` and `DB_PASSWORD` (or `DB_URL`) to the
external database. The hostname must resolve, and the port must be reachable, **from
the `nakama` and `nakama-migrate` containers**. The job's DNS and TCP checks say which
of the two fails. The bundled `postgres` service still starts but is not used.

- **Another Dokploy application or database service.** Applications in one Dokploy
  project do not share a Docker network automatically. Use the database's internal
  host only after connecting the two services to a common network. Dokploy's own
  services use the external `dokploy-network`, so add `networks: [default,
  dokploy-network]` to `nakama` and `nakama-migrate`, and declare
  `networks: { dokploy-network: { external: true } }`. Otherwise use the database's
  external address with TLS. This file does not join any external network by itself.
- **A remote server.** Use `DB_SSL_MODE=verify-full`. Mount the CA certificate into
  both containers (Dokploy compose **Advanced → Volumes/Mounts**) and set
  `DB_SSL_ROOT_CERT` to its path. Do not publish the database port without TLS and a
  firewall.
- The role needs to own the database, or have `CREATE` in it. Nakama creates the
  database only if it is missing and the role may do so.

## Troubleshooting

| Symptom (migration job log) | Cause and fix |
|---|---|
| `set DB_PASSWORD in the Dokploy Environment tab` (deployment log) | Add `DB_PASSWORD` |
| `DNS: database hostname '…' does not resolve` (exit 3) | Wrong `DB_HOST`, or no shared network |
| `DNS: legacy CockroachDB hostname 'cockroachdb' does not resolve` | See "If the import reports a DNS error" above; `false` for a new installation |
| `TCP: … refuses connections` (exit 4) | Wrong `DB_PORT`, or the database is down |
| `authentication: the database rejected DB_USER/DB_PASSWORD` | Wrong credentials. A bundled volume keeps the password it was created with: `DB_PASSWORD` must match it |
| `TLS: check DB_SSL_MODE and DB_SSL_ROOT_CERT` | The certificate or mode does not fit the server |
| `SQL migration: …` | The `Failed to apply migrations` line names the file and statement. Fix it in a new commit; never edit `migration_info` |
| `refusing to import over its data` (exit 5) | PostgreSQL was already used. To keep it as it is, set `DB_IMPORT_FROM_COCKROACH=false` |

Read-only migration status:

```sh
docker compose run --rm --no-deps --entrypoint /nakama/nakama-db.sh nakama-migrate status
```

In Dokploy, query `migration_info` from the `postgres` terminal instead:
`psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT count(*) FROM migration_info"`.

## Rollback and backups

- Back up PostgreSQL before risky changes, from the `postgres` terminal:
  `pg_dump -U "$POSTGRES_USER" -Fc "$POSTGRES_DB" > /var/lib/postgresql/data/backup-$(date +%F).dump`
- Redeploying an older application version: a version before PostgreSQL connects to
  CockroachDB again, which holds the data **as of the import**. Writes made on
  PostgreSQL since then are not in it. Never delete either volume to roll back.
- A failed deployment leaves the previous containers and both volumes in place. Fix
  the cause and deploy again; the job and the import are re-runnable.

## Tests

```sh
# Launcher and compose wiring (no database; needs bash, python3, docker compose)
scripts/nakama-db-test.sh

# CockroachDB v24.1.33 production state imported into PostgreSQL 16.8, end to end
docker build -f build/Dockerfile.local -t nakama-farsi:local .
scripts/cockroach-import-test.sh nakama-farsi:local

# Jazebe workflows on PostgreSQL: see the profile_postgres job in
# .github/workflows/jazebe-migrations.yaml. The tests only accept a loopback database.
```

`.github/workflows/jazebe-migrations.yaml` runs all three (`profile_postgres` and
`deployment_image`).
