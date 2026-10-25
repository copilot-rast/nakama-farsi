#!/bin/bash
# Tests for build/nakama-db.sh, the launcher that turns DB_* environment
# variables into the Nakama database address. A stub replaces the Nakama
# binary, so no database is needed; a local TCP listener stands in for one.
# With Docker available, also checks the rendered docker-compose.yml.
#
#   scripts/nakama-db-test.sh
set -u

root=$(cd "$(dirname "$0")/.." && pwd)
launcher=$root/build/nakama-db.sh
tmp=$(mktemp -d)
listener_pid=''
cleanup() { [ -n "$listener_pid" ] && kill "$listener_pid" 2>/dev/null; rm -rf "$tmp"; }
trap cleanup EXIT

failures=0
pass() { echo "ok   $1"; }
fail() { echo "FAIL $1"; failures=$((failures + 1)); }

# The stub prints its arguments and the config file it was given, then
# prints STUB_OUTPUT and exits with STUB_EXIT.
cat >"$tmp/nakama" <<'EOF'
#!/bin/bash
echo "ARGS: $*"
while [ $# -gt 0 ]; do
    if [ "$1" = --config ]; then echo "CONFIG:"; cat "$2"; fi
    shift
done
[ -n "${STUB_OUTPUT:-}" ] && echo "$STUB_OUTPUT"
exit "${STUB_EXIT:-0}"
EOF
chmod +x "$tmp/nakama"

port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1])')
python3 -c "
import socket
s = socket.socket(); s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(('127.0.0.1', $port)); s.listen(64)
while True: s.accept()[0].close()
" &
listener_pid=$!
sleep 0.5

# run <env...> -- <launcher args...>: output in $tmp/out, status in $status.
run() {
    local envs=()
    while [ "$1" != -- ]; do envs+=("$1"); shift; done
    shift
    env -i PATH="$PATH" NAKAMA_BIN="$tmp/nakama" DB_WAIT_INTERVAL_SECONDS=1 "${envs[@]}" \
        bash "$launcher" "$@" >"$tmp/out" 2>&1
    status=$?
}
has() { grep -qF -- "$1" "$tmp/out"; }

# Individual fields, with credentials that need URL encoding.
run DB_HOST=127.0.0.1 DB_PORT=$port DB_NAME=game_db DB_USER='app user' DB_PASSWORD='p@ss:w/rd%#?' DB_SSL_MODE=verify-full DB_SSL_ROOT_CERT=/certs/ca.crt -- migrate
[ $status = 0 ] && has '    - "app%20user:p%40ss%3Aw%2Frd%25%23%3F@127.0.0.1:'"$port"'/game_db?sslmode=verify-full&sslrootcert=%2Fcerts%2Fca.crt"' &&
    has 'ARGS: migrate up --config' && pass "fields: encoded address in the config file" || { fail "fields: encoded address in the config file"; cat "$tmp/out"; }
grep -v '^    - ' "$tmp/out" | grep -qE 'p@ss|p%40ss' &&
    fail "fields: password leaked outside the config file" || pass "fields: password only in the private config file"
has "database (DB_HOST): app%20user:xxxxx@127.0.0.1:$port/game_db" && pass "fields: redacted log line" || fail "fields: redacted log line"

# Defaults: database nakama, no password, no sslmode (Nakama's prefer).
run DB_HOST=127.0.0.1 DB_PORT=$port DB_USER=nakama -- migrate
has "    - \"nakama@127.0.0.1:$port/nakama\"" && has 'DB_PASSWORD is empty' && pass "defaults" || { fail "defaults"; cat "$tmp/out"; }
# The default port is PostgreSQL's.
run DB_HOST=127.0.0.1 DB_USER=nakama DB_WAIT_TIMEOUT_SECONDS=0 -- migrate
has '127.0.0.1:5432/nakama' && pass "default port 5432" || { fail "default port 5432"; cat "$tmp/out"; }
run DB_HOST=127.0.0.1 -- migrate
[ $status = 2 ] && has 'DB_USER is not set' && pass "missing DB_USER" || fail "missing DB_USER"
# The import is off unless asked for, and a bad value is rejected.
run DB_HOST=127.0.0.1 DB_PORT=$port DB_USER=nakama -- migrate
! has 'legacy CockroachDB' && pass "no import by default" || fail "no import by default"
run DB_HOST=127.0.0.1 DB_PORT=$port DB_USER=nakama DB_IMPORT_FROM_COCKROACH=maybe -- migrate
[ $status = 2 ] && has 'DB_IMPORT_FROM_COCKROACH must be true or false' && pass "invalid DB_IMPORT_FROM_COCKROACH" || { fail "invalid DB_IMPORT_FROM_COCKROACH"; cat "$tmp/out"; }

# DB_URL takes precedence over the individual fields.
run DB_URL="postgresql://u:s%40cret@127.0.0.1:$port/other?sslmode=require" DB_HOST=ignored.invalid DB_NAME=x -- migrate
[ $status = 0 ] && has "    - \"u:s%40cret@127.0.0.1:$port/other?sslmode=require\"" && has 'DB_HOST and the other DB_* fields are ignored' &&
    has "database (DB_URL): u:xxxxx@127.0.0.1:$port/other?sslmode=require" && pass "DB_URL precedence and redaction" || { fail "DB_URL precedence and redaction"; cat "$tmp/out"; }

# A raw @ inside a DB_URL password is still redacted up to the last @.
run DB_URL="u:a@b@127.0.0.1:$port/nakama" -- migrate
has "database (DB_URL): u:xxxxx@127.0.0.1:$port/nakama" && ! grep 'database (DB_URL)' "$tmp/out" | grep -q 'a@b' &&
    pass "DB_URL raw @ redacted" || { fail "DB_URL raw @ redacted"; cat "$tmp/out"; }

# Invalid or missing configuration fails early with exit 2.
run -- migrate
[ $status = 2 ] && has 'DB_HOST is not set' && pass "missing DB_HOST" || fail "missing DB_HOST"
run DB_HOST=h DB_USER=u DB_PORT=abc -- migrate
[ $status = 2 ] && has 'DB_PORT must be a number' && pass "invalid DB_PORT" || fail "invalid DB_PORT"
run DB_HOST=h DB_USER=u DB_NAME='bad-name;' -- migrate
[ $status = 2 ] && has 'DB_NAME may only contain' && pass "invalid DB_NAME" || fail "invalid DB_NAME"
run DB_HOST=h DB_USER=u DB_MIGRATIONS_ENABLED=maybe -- migrate
[ $status = 2 ] && pass "invalid DB_MIGRATIONS_ENABLED" || fail "invalid DB_MIGRATIONS_ENABLED"
run DB_HOST=h DB_USER=u -- bogus
[ $status = 2 ] && has 'usage:' && pass "unknown mode" || fail "unknown mode"

# Unreachable databases: DNS (3) and TCP (4) are reported apart, after the timeout.
start=$SECONDS
run DB_HOST=no-such-db.invalid DB_USER=u DB_WAIT_TIMEOUT_SECONDS=2 -- migrate
[ $status = 3 ] && has "DNS: database hostname 'no-such-db.invalid' does not resolve" && has 'waiting for database' && [ $((SECONDS - start)) -ge 2 ] &&
    pass "DNS failure retried, then exit 3" || { fail "DNS failure retried, then exit 3"; cat "$tmp/out"; }
run DB_HOST=127.0.0.1 DB_USER=u DB_PORT=1 DB_WAIT_TIMEOUT_SECONDS=1 -- migrate
[ $status = 4 ] && has 'TCP: database 127.0.0.1 resolves but port 1 refuses' && ! has 'database. 4' && pass "TCP failure, exit 4" || { fail "TCP failure, exit 4"; cat "$tmp/out"; }
! has 'ARGS:' && pass "nakama not started while unreachable" || fail "nakama not started while unreachable"

# Migration failures keep Nakama's exit status and are classified.
run DB_HOST=127.0.0.1 DB_USER=u DB_PORT=$port STUB_EXIT=1 STUB_OUTPUT='{"msg":"Failed to apply migrations","error":"ERROR: column \"x\" does not exist (SQLSTATE 42703)"}' -- migrate
[ $status = 1 ] && has 'migration failed (exit 1), SQL migration' && pass "SQL migration failure" || { fail "SQL migration failure"; cat "$tmp/out"; }
run DB_HOST=127.0.0.1 DB_USER=u DB_PORT=$port STUB_EXIT=1 STUB_OUTPUT='Error pinging database: password authentication failed for user root (SQLSTATE 28P01)' -- migrate
[ $status = 1 ] && has 'authentication: the database rejected' && pass "authentication failure" || fail "authentication failure"
run DB_HOST=127.0.0.1 DB_USER=u DB_PORT=$port STUB_EXIT=1 STUB_OUTPUT='Error pinging database: tls: failed to verify certificate: x509: unknown authority' -- migrate
[ $status = 1 ] && has 'TLS: check DB_SSL_MODE' && pass "TLS failure" || fail "TLS failure"
run DB_HOST=127.0.0.1 DB_USER=u DB_PORT=$port STUB_EXIT=1 STUB_OUTPUT='hostname resolving error: lookup cockroachdb on 127.0.0.11:53: server misbehaving' -- migrate
[ $status = 1 ] && has 'DNS: hostname' && pass "DNS failure inside Nakama" || fail "DNS failure inside Nakama"

# Migrations can be disabled; Nakama is not run.
run DB_HOST=no-such-db.invalid DB_USER=u DB_MIGRATIONS_ENABLED=false -- migrate
[ $status = 0 ] && ! has 'ARGS:' && pass "DB_MIGRATIONS_ENABLED=false" || fail "DB_MIGRATIONS_ENABLED=false"

# Server mode forwards flags after the private config file.
run DB_HOST=127.0.0.1 DB_USER=u DB_PORT=$port -- server --name nakama1 --logger.level DEBUG
[ $status = 0 ] && has 'ARGS: --config ' && has ' --name nakama1 --logger.level DEBUG' && ! has 'database.address' &&
    pass "server mode" || { fail "server mode"; cat "$tmp/out"; }

# Status mode is read-only `migrate status`, and does not import.
run DB_HOST=127.0.0.1 DB_USER=u DB_PORT=$port DB_IMPORT_FROM_COCKROACH=true -- status
[ $status = 0 ] && has 'ARGS: migrate status --config ' && ! has 'migrate up' && ! has 'legacy CockroachDB' &&
    pass "status mode" || { fail "status mode"; cat "$tmp/out"; }

# docker-compose.yml: PostgreSQL is the database of both Nakama services,
# which get the same DB_* settings, none in a command line; the bundled
# postgres is initialised from the same values; CockroachDB is only the
# profile-gated import source.
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    compose() { docker compose -f "$root/docker-compose.yml" "$@"; }
    printf 'DB_HOST=db.example.internal\nDB_PORT=6543\nDB_NAME=prod\nDB_USER=app\nDB_PASSWORD=compose-secret\n' >"$tmp/env"
    if compose --env-file "$tmp/env" config --format json >"$tmp/compose.json" 2>"$tmp/compose.err"; then
        python3 - "$tmp/compose.json" >"$tmp/out" 2>&1 <<'EOF'
import json, sys
cfg = json.load(open(sys.argv[1]))
c = cfg["services"]
keys = ["DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE", "DB_SSL_ROOT_CERT", "DB_URL"]
m, n, p = c["nakama-migrate"]["environment"], c["nakama"]["environment"], c["postgres"]
assert m["DB_HOST"] == "db.example.internal" and m["DB_PORT"] == "6543", m
assert all(m.get(k) == n.get(k) for k in keys), "services differ"
for s in ("nakama-migrate", "nakama"):
    cmd = json.dumps(c[s].get("entrypoint", [])) + json.dumps(c[s].get("command", []))
    assert "compose-secret" not in cmd and "cockroach" not in cmd and "database.address" not in cmd, s
    assert set(c[s]["depends_on"]) <= {"postgres", "nakama-migrate"}, c[s]["depends_on"]
    assert c[s]["depends_on"]["postgres"]["condition"] == "service_healthy", s
assert p["image"].startswith("postgres:16."), p["image"]
assert p["environment"] == {"POSTGRES_DB": "prod", "POSTGRES_USER": "app", "POSTGRES_PASSWORD": "compose-secret"}, p["environment"]
assert any(v["target"] == "/var/lib/postgresql/data" and v["source"] == "postgres-data" for v in p["volumes"]), p["volumes"]
assert "compose-secret" not in json.dumps(p["healthcheck"]), "password in healthcheck"
assert "cockroachdb" not in c, "cockroachdb is active without its profile"
assert m["DB_IMPORT_FROM_COCKROACH"] == "true", m
print("compose ok")
EOF
        has 'compose ok' && pass "compose: PostgreSQL for both services, same settings, none in commands" || { fail "compose"; cat "$tmp/out"; }
    else
        fail "compose config"; cat "$tmp/compose.err"
    fi
    # The legacy service is still defined behind its profile, on its volume.
    if COMPOSE_PROFILES=cockroach-import compose --env-file "$tmp/env" config --format json >"$tmp/compose.json" 2>/dev/null &&
        python3 -c 'import json,sys; c=json.load(open(sys.argv[1]))["services"]["cockroachdb"]; assert any(v["source"]=="data" for v in c["volumes"])' "$tmp/compose.json"; then
        pass "compose: legacy cockroachdb behind the cockroach-import profile, on volume data"
    else
        fail "compose: legacy cockroachdb profile"
    fi
    # Without DB_PASSWORD the deployment is refused with a clear message.
    printf 'DB_HOST=postgres\n' >"$tmp/env"
    if compose --env-file "$tmp/env" config -q >/dev/null 2>"$tmp/out"; then
        fail "compose: DB_PASSWORD required"
    else
        has 'set DB_PASSWORD in the Dokploy Environment tab' && pass "compose: DB_PASSWORD required" || { fail "compose: DB_PASSWORD required"; cat "$tmp/out"; }
    fi
else
    echo "skip compose checks (docker compose not available)"
fi

[ $failures = 0 ] && echo "PASS" || { echo "$failures failed"; exit 1; }
