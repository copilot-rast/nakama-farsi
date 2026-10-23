# Native PostgreSQL on Andronix (Android, no Docker)

This guide runs nakama-farsi and its Jazebe module against the PostgreSQL
server installed by the Andronix `db` pack. It is for local development on an
Android phone running Ubuntu 26.04 ARM64 under Andronix/PRoot. It does not
change the Docker Compose files or production deployment.

## Environment and limitations

- Ubuntu 26.04 aarch64 under PRoot: there is no systemd, Docker Engine or Docker Compose.
- PostgreSQL 18.6 (`/usr/lib/postgresql/18`). Data directory `/var/lib/andronix-postgres`, socket `/run/postgresql`, port 5432. It listens only on `127.0.0.1` and `::1`.
- Authentication: local socket = `peer`, TCP = `scram-sha-256`. Your OS user is the cluster superuser, so plain `psql` is the admin connection.
- `sudo` needs an interactive terminal. Run the start/stop commands in an Andronix terminal, not through tools that lack a TTY.
- PostgreSQL stops when Android kills Andronix. After every restart, start it again and verify it (below). Data persists in `/var/lib/andronix-postgres`, which is in the distro's private storage, not shared Android storage.
- The CI workflow uses PostgreSQL 16.8. Locally verified on 18.6: all 52 migrations apply as a non-superuser, and `pg_trgm` installs as a trusted extension.
- The Nakama Jazebe module is a Go plugin and needs cgo. Install a C compiler once with `sudo apt-get install -y gcc libc6-dev` (verified with gcc 15.2.0).
- Go 1.27.1 (required by `go.mod`) is installed from the official arm64 tarball into `~/.local/go`; add `~/.local/go/bin` to `PATH`.

## Databases, roles and credentials

| Database | Owner role | Purpose | Credentials |
|---|---|---|---|
| `nakama_dev` | `nakama_dev` | Running Nakama locally | `~/.config/nakama-farsi/dev.env` |
| `nakama_jazebe_test_disposable` | `nakama_test` | PostgreSQL integration tests; may be reset at any time | `~/.config/nakama-farsi/test.env` |

- Both roles are `NOSUPERUSER NOCREATEDB NOCREATEROLE`, and each role owns only its own database and its `public` schema.
- `CONNECT` is revoked from `PUBLIC` on both databases, so neither role can connect to the other's database.
- The Jazebe tables are part of Nakama's schema: the 32 `*-jazebe-*.sql` migrations are embedded in Nakama's own migrator. They are not a separate Jazebe database, and the Jazebe NestJS database is never used.
- The credential files are outside Git, mode `0600`, in a `0700` directory. Each file defines `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD` and `NAKAMA_DATABASE_ADDRESS`. `test.env` also defines `JAZEBE_DISPOSABLE_TEST_DB_URL`.

One-time setup, as your admin user. The steps refuse to run if the role already exists:

```sh
umask 077; mkdir -p ~/.config/nakama-farsi
role=nakama_dev db=nakama_dev pw=$(openssl rand -hex 24)
psql -v ON_ERROR_STOP=1 -v role="$role" -v pw="$pw" <<'SQL'
CREATE ROLE :"role" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION PASSWORD :'pw';
SQL
createdb -O "$role" -E UTF8 -T template0 "$db"
psql -c "REVOKE ALL ON DATABASE \"$db\" FROM PUBLIC; GRANT CONNECT, TEMPORARY ON DATABASE \"$db\" TO \"$role\";"
psql -d "$db" -c "REVOKE CREATE ON SCHEMA public FROM PUBLIC; ALTER SCHEMA public OWNER TO \"$role\";"
printf 'PGHOST=127.0.0.1\nPGPORT=5432\nPGDATABASE=%s\nPGUSER=%s\nPGPASSWORD=%s\nNAKAMA_DATABASE_ADDRESS=%s:%s@127.0.0.1:5432/%s\n' \
  "$db" "$role" "$pw" "$role" "$pw" "$db" > ~/.config/nakama-farsi/dev.env
```

Repeat with `role=nakama_test db=nakama_jazebe_test_disposable` and `test.env`, then append:
`JAZEBE_DISPOSABLE_TEST_DB_URL=postgres://nakama_test:<pw>@127.0.0.1:5432/nakama_jazebe_test_disposable?sslmode=disable`.

## Start, verify, connect, stop

```sh
sudo andronix-postgres start          # in an Andronix terminal
pg_isready -h /run/postgresql         # expect: accepting connections
psql -tAc 'select version()'          # admin, over the local socket
(set -a; . ~/.config/nakama-farsi/dev.env; psql -c 'select count(*) from migration_info')
sudo andronix-postgres stop
```

`pg_ctl: another server might be running; trying to start server anyway` after an unclean Android shutdown means a stale `postmaster.pid`, and is harmless. Confirm that exactly one `postgres -D /var/lib/andronix-postgres` process is running and that `/var/log/andronix-postgres.log` ends with `database system is ready to accept connections`.

## Migrations

Nakama embeds all 52 migrations: 20 upstream, then 32 Jazebe (`20261022120000` and later). They are applied in timestamp order and recorded in `migration_info`.

```sh
export PATH=$HOME/.local/go/bin:$PATH
go build -mod=vendor -o ~/.local/share/nakama-farsi/nakama .
(set -a; . ~/.config/nakama-farsi/dev.env; ~/.local/share/nakama-farsi/nakama migrate up --database.address "$NAKAMA_DATABASE_ADDRESS")
```

- On an empty database, a first run applies 52 migrations and a second run applies 0.
- The Jazebe migrations only add `jazebe_*` objects. Their only link to Nakama core tables is foreign keys to `users`.
- **Do not run `migrate down` across Jazebe migrations.** 31 of their 32 Down sections intentionally keep user data (`SELECT 1;`), so a down run removes the `migration_info` rows but leaves the tables. The next `migrate up` then fails with errors such as `constraint ... already exists`. This was reproduced on the disposable database. Roll back from a backup (`pg_dump`) instead.

## Tests

```sh
cd nakama-farsi
export PATH=$HOME/.local/go/bin:$PATH GOFLAGS=-mod=vendor
go test ./integrations/jazebe/native ./migrate ./scripts/jazebe_route_audit -count=1
(set -a; . ~/.config/nakama-farsi/test.env; go test ./integrations/jazebe/native -run Postgres -count=1 -v)
```

- The PostgreSQL tests use `openDisposableTestDB` (`integrations/jazebe/native/disposable_db_test.go`). It refuses any target whose database name does not end in `_disposable` or whose host is not loopback or a local socket. After connecting, it also checks `current_database()`.
- `TestJazebeSQLPreparesPostgres` prepares every complete SQL statement in the module against the migrated schema, without executing any of them.

To reset the disposable database, as its owner only:

```sh
(set -a; . ~/.config/nakama-farsi/test.env; [ "$PGDATABASE" = nakama_jazebe_test_disposable ] && \
  psql -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;' && \
  ~/.local/share/nakama-farsi/nakama migrate up --database.address "$NAKAMA_DATABASE_ADDRESS")
```

## Run Nakama with the Jazebe module

The plugin and the server must be built with the same toolchain and flags as `build/Dockerfile.local`:

```sh
R=~/.local/share/nakama-farsi; mkdir -p $R/data/modules
export PATH=$HOME/.local/go/bin:$PATH CGO_ENABLED=1
go build -o $R/nakama -trimpath -mod=vendor -gcflags "-trimpath $PWD" -asmflags "-trimpath $PWD" -ldflags "-s -w -X main.version=local" .
go build -buildmode=plugin -o $R/data/modules/jazebe.so -trimpath -mod=vendor -gcflags "-trimpath $PWD" -asmflags "-trimpath $PWD" ./integrations/jazebe/native
cd $R && ./nakama --config ~/.config/nakama-farsi/nakama-local.yml
```

Keep secrets out of the command line, because any process can read another process's arguments. `~/.config/nakama-farsi/nakama-local.yml` (mode `0600`, outside Git) holds the database address. It also sets `socket.address` and `console.address` to `127.0.0.1`, and gives random values to `socket.server_key`, `console.password`, `console.signing_key`, `session.encryption_key`, `session.refresh_encryption_key` and `runtime.http_key`. Without the address settings, Nakama listens on every interface, including the phone's Wi-Fi.

```yaml
name: nakama1                # 1-16 characters
data_dir: /home/<you>/.local/share/nakama-farsi/data
logger: {level: INFO, file: /home/<you>/.local/share/nakama-farsi/nakama.log}
socket: {address: 127.0.0.1, server_key: <openssl rand -hex 24>}
console: {address: 127.0.0.1, username: admin, password: <random>, signing_key: <random>}
session: {encryption_key: <random>, refresh_encryption_key: <random>}
runtime: {http_key: <random>}
database: {address: ["nakama_dev:<password>@127.0.0.1:5432/nakama_dev"]}
```

The ports are 7350 (API), 7351 (console) and 7349 (gRPC). The worker flags (`JAZEBE_*_WORKER_ENABLED`, `JAZEBE_MEDAL_BACKFILL_ENABLED`) default to off. Operators are configured with `JAZEBE_OPERATOR_USER_IDS=<nakama user id>`.

Smoke test (verified 2026-10-23). Startup logs `Found runtime modules ... jazebe.so` and `Jazebe native integration RPCs registered`, with 196 RPCs:

```sh
curl -s http://127.0.0.1:7350/healthcheck                                   # 200
KEY=<server_key>; curl -s -u "$KEY:" -H 'Content-Type: application/json' \
  'http://127.0.0.1:7350/v2/account/authenticate/device?create=true&username=jz_op' -d '{"id":"<10-128 char device id>"}'
curl -s -H "Authorization: Bearer <token>" -H 'Content-Type: application/json' \
  http://127.0.0.1:7350/v2/rpc/jazebe_brand_create -d '"{\"name\":\"برند\",\"ownerUserId\":\"<user id>\"}"'
```

Expected results: 403 for a non-operator, 400 for an empty name, and 200 for an operator. A 200 writes `jazebe_brand`, `jazebe_brand_member` (owner), `jazebe_brand_admin` (main) and a `brand.create` audit row. `jazebe_planet_create` returns 409 for a duplicate name within a brand and 403 for callers without a main/technical brand role.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `pg_isready: no response` | The server is stopped, for example after Android closed Andronix. Run `sudo andronix-postgres start`. |
| `sudo: a terminal is required` | The command was run without a TTY. Use an Andronix terminal session. |
| `password authentication failed` | Source the right `.env` file, and use `127.0.0.1` rather than the socket for the app roles. |
| `permission denied for database` / `no CONNECT privilege` | The role is connecting to the other database. This is intended isolation. |
| `cgo: C compiler "gcc" not found` | Run `sudo apt-get install -y gcc libc6-dev`. |
| `plugin was built with a different version of package` | Rebuild both `nakama` and `jazebe.so` with the identical flags shown above. |
| `Auth token invalid` (code 16) after a restart | Tokens from before the restart are rejected. Authenticate again; with `create=false`, the existing account is reused. |
| `Name must be 1-16 characters` | The `name:` in the YAML config is too long. |
| `could not determine data type of parameter $N` | A bare placeholder was passed to a variadic `"any"` function such as `jsonb_build_object`. Add an explicit cast. `TestJazebeSQLPreparesPostgres` catches this. |

## Unavailable compared with Docker Compose

`docker-compose.yml` runs only PostgreSQL and Nakama, and both run natively here. The Jazebe NestJS stack's external services (SMS, KYC provider, object storage, webhooks) are not part of this local setup.
