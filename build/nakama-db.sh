#!/bin/bash
# Run Nakama against the PostgreSQL database described by DB_* environment
# variables.
#
#   nakama-db.sh migrate            wait for the database, import the legacy
#                                   CockroachDB data once if present, then
#                                   `nakama migrate up`
#   nakama-db.sh server [flags...]  wait for the database, then exec the server
#   nakama-db.sh status             read-only `nakama migrate status`
#
# DB_URL, when set, takes precedence over DB_HOST/DB_PORT/DB_NAME/DB_USER/
# DB_PASSWORD/DB_SSL_MODE/DB_SSL_ROOT_CERT. The address reaches Nakama through a
# private config file, never the command line, and is only logged redacted.
# See docs/integration/database-configuration.md.
set -euo pipefail

NAKAMA_BIN=${NAKAMA_BIN:-/nakama/nakama}
PSQL_BIN=${PSQL_BIN:-psql}

log() { printf 'nakama-db: %s\n' "$*" >&2; }
die() { log "ERROR: $1"; exit "${2:-1}"; }

# Percent-encode a URL component byte by byte.
urlencode() {
    local LC_ALL=C s=$1 out='' c i
    for ((i = 0; i < ${#s}; i++)); do
        c=${s:i:1}
        case $c in
            [a-zA-Z0-9.~_-]) out+=$c ;;
            *) printf -v c '%%%02X' "'$c"; out+=$c ;;
        esac
    done
    printf '%s' "$out"
}

# parse_target <address without scheme> <default port>: sets PT_HOST, PT_PORT
# and PT_REDACTED. The userinfo ends at the last @, as in Go's URL parser.
parse_target() {
    local address=$1 hostport=$1
    case $address in *@*) hostport=${address##*@} ;; esac
    hostport=${hostport%%[/?]*}
    case $hostport in
        \[*\]:*) PT_HOST=${hostport%]:*}; PT_HOST=${PT_HOST#[}; PT_PORT=${hostport##*]:} ;;
        \[*\]) PT_HOST=${hostport#[}; PT_HOST=${PT_HOST%]}; PT_PORT=$2 ;;
        *:*) PT_HOST=${hostport%:*}; PT_PORT=${hostport##*:} ;;
        *) PT_HOST=$hostport; PT_PORT=$2 ;;
    esac
    PT_REDACTED=$address
    case $address in
        *@*)
            local userinfo=${address%@*}
            case $userinfo in *:*) PT_REDACTED="${userinfo%%:*}:xxxxx@${address##*@}" ;; esac
            ;;
    esac
}

strip_scheme() { local s=${1#postgresql://}; printf '%s' "${s#postgres://}"; }

# Sets DB_ADDRESS (without scheme), DB_SOURCE, DB_TARGET_HOST, DB_TARGET_PORT
# and DB_REDACTED from the environment.
build_address() {
    if [ -n "${DB_URL:-}" ]; then
        DB_ADDRESS=$(strip_scheme "$DB_URL")
        DB_SOURCE=DB_URL
    else
        [ -n "${DB_HOST:-}" ] || die "DB_HOST is not set (or set DB_URL)" 2
        [ -n "${DB_USER:-}" ] || die "DB_USER is not set (or set DB_URL)" 2
        local port=${DB_PORT:-5432} name=${DB_NAME:-nakama} query=''
        case $port in '' | *[!0-9]*) die "DB_PORT must be a number, got '$port'" 2 ;; esac
        case $name in '' | *[!a-zA-Z0-9_]*) die "DB_NAME may only contain letters, digits and _, got '$name'" 2 ;; esac
        local userinfo
        userinfo=$(urlencode "$DB_USER")
        if [ -n "${DB_PASSWORD:-}" ]; then
            userinfo+=:$(urlencode "$DB_PASSWORD")
        fi
        [ -n "${DB_SSL_MODE:-}" ] && query+="&sslmode=$(urlencode "$DB_SSL_MODE")"
        [ -n "${DB_SSL_ROOT_CERT:-}" ] && query+="&sslrootcert=$(urlencode "$DB_SSL_ROOT_CERT")"
        DB_ADDRESS="$userinfo@$DB_HOST:$port/$name${query:+?${query#&}}"
        DB_SOURCE=DB_HOST
    fi
    parse_target "$DB_ADDRESS" 5432
    DB_TARGET_HOST=$PT_HOST DB_TARGET_PORT=$PT_PORT DB_REDACTED=$PT_REDACTED
    [ -n "$DB_TARGET_HOST" ] || die "$DB_SOURCE has no host" 2
    case $DB_TARGET_PORT in '' | *[!0-9]*) die "$DB_SOURCE has an invalid port" 2 ;; esac
}

# wait_for <host> <port> <what>: retry until <host> resolves and accepts TCP
# connections, for at most DB_WAIT_TIMEOUT_SECONDS. DNS and TCP failures are
# reported apart.
wait_for() {
    local host=$1 port=$2 what=$3
    local timeout=${DB_WAIT_TIMEOUT_SECONDS:-120} interval=${DB_WAIT_INTERVAL_SECONDS:-2}
    case $timeout in '' | *[!0-9]*) die "DB_WAIT_TIMEOUT_SECONDS must be a number" 2 ;; esac
    local deadline=$((SECONDS + timeout)) state=''
    while :; do
        if ! getent hosts "$host" >/dev/null 2>&1; then
            state=dns
        elif ! timeout 5 bash -c 'exec 3<>"/dev/tcp/$0/$1"' "$host" "$port" 2>/dev/null; then
            state=tcp
        else
            log "$what $host:$port is reachable"
            return 0
        fi
        if [ "$SECONDS" -ge "$deadline" ]; then
            if [ "$state" = dns ]; then
                die "DNS: $what hostname '$host' does not resolve from this container after ${timeout}s. Check the host setting, and that this service shares a Docker network with the database." 3
            fi
            die "TCP: $what $host resolves but port $port refuses connections after ${timeout}s. Check the port and that the database is running." 4
        fi
        log "waiting for $what $host:$port ($state not ready yet)"
        sleep "$interval"
    done
}

# Write the address to a config file readable only by this user.
write_config() {
    DB_CONFIG=$(mktemp "${TMPDIR:-/tmp}/nakama-db.XXXXXX.yml")
    chmod 600 "$DB_CONFIG"
    local escaped=${DB_ADDRESS//\\/\\\\}
    escaped=${escaped//\"/\\\"}
    printf 'database:\n  address:\n    - "%s"\n' "$escaped" >"$DB_CONFIG"
}

# Name the failure class of a failed Nakama or psql run from its log.
classify_failure() {
    local out=$1
    if grep -qE 'SQLSTATE 28(000|P01)|password authentication failed|authentication failed|no pg_hba.conf entry' "$out"; then
        echo "authentication: the database rejected DB_USER/DB_PASSWORD"
    elif grep -qiE 'tls|x509|certificate|ssl' "$out"; then
        echo "TLS: check DB_SSL_MODE and DB_SSL_ROOT_CERT"
    elif grep -q 'Failed to apply migrations' "$out"; then
        echo "SQL migration: see the 'Failed to apply migrations' line above for the file and statement"
    elif grep -qE 'SQLSTATE 3D000|database "[^"]*" does not exist' "$out"; then
        echo "database: DB_NAME does not exist and DB_USER may not create it"
    elif grep -qE 'no such host|server misbehaving|lookup |could not translate host name' "$out"; then
        echo "DNS: hostname '$DB_TARGET_HOST' stopped resolving"
    elif grep -qE 'connection refused|i/o timeout|connect: |Connection refused' "$out"; then
        echo "connection: $DB_TARGET_HOST:$DB_TARGET_PORT is unreachable"
    else
        echo "unknown: see the log above"
    fi
}

# run_logged <description> <command...>: run, keep the output for
# classification, and exit with the command's status on failure.
run_logged() {
    local what=$1 out status=0
    shift
    out=$(mktemp "${TMPDIR:-/tmp}/nakama-db.XXXXXX.log")
    "$@" 2>&1 | tee "$out" || status=$?
    if [ "$status" -ne 0 ]; then
        log "ERROR: $what failed (exit $status), $(classify_failure "$out")"
        rm -f "$out" "$DB_CONFIG"
        exit "$status"
    fi
    rm -f "$out"
}

tpsql() { "$PSQL_BIN" -X -q -At -v ON_ERROR_STOP=1 -d "postgresql://$DB_ADDRESS" "$@"; }
spsql() { "$PSQL_BIN" -X -q -At -v ON_ERROR_STOP=1 -d "postgresql://$IMPORT_ADDRESS" "$@"; }

# One-time copy of the legacy CockroachDB database into an empty PostgreSQL
# target, at the source's migration level. See "Importing the CockroachDB
# data" in docs/integration/database-configuration.md.
import_from_cockroach() {
    case ${DB_IMPORT_FROM_COCKROACH:-false} in
        false | 0 | no) return 0 ;;
        true | 1 | yes) ;;
        *) die "DB_IMPORT_FROM_COCKROACH must be true or false" 2 ;;
    esac
    command -v "$PSQL_BIN" >/dev/null 2>&1 || die "psql is required to import from CockroachDB" 2

    # A completed import is final: the source is not contacted again, so it
    # can be removed afterwards.
    if [ "$(tpsql -c "SELECT to_regclass('nakama_farsi.cockroach_import') IS NOT NULL")" = t ]; then
        local done_at
        done_at=$(tpsql -c "SELECT max(completed_at) FROM nakama_farsi.cockroach_import")
        if [ -n "$done_at" ]; then
            log "CockroachDB data was imported at $done_at; not importing again"
            return 0
        fi
    fi

    IMPORT_ADDRESS=$(strip_scheme "${COCKROACH_IMPORT_URL:-postgresql://root@cockroachdb:26257/nakama?sslmode=disable}")
    parse_target "$IMPORT_ADDRESS" 26257
    log "legacy CockroachDB import source: $PT_REDACTED (set DB_IMPORT_FROM_COCKROACH=false for a new installation without one)"
    wait_for "$PT_HOST" "$PT_PORT" "legacy CockroachDB"

    local work
    work=$(mktemp -d "${TMPDIR:-/tmp}/nakama-import.XXXXXX")
    spsql -c "SELECT id FROM migration_info ORDER BY id" >"$work/source.ids" 2>"$work/err" ||
        { cat "$work/err" >&2; die "could not read migration_info from the legacy CockroachDB" 6; }
    local n
    n=$(wc -l <"$work/source.ids")
    [ "$n" -gt 0 ] || die "legacy CockroachDB has no applied migrations; set DB_IMPORT_FROM_COCKROACH=false if there is nothing to import" 6

    : >"$work/target.ids"
    if [ "$(tpsql -c "SELECT to_regclass('public.migration_info') IS NOT NULL")" = t ]; then
        tpsql -c "SELECT id FROM migration_info ORDER BY id" >"$work/target.ids"
    fi
    if [ -s "$work/target.ids" ] && ! cmp -s "$work/source.ids" "$work/target.ids"; then
        die "PostgreSQL already has $(wc -l <"$work/target.ids") migrations and no completed CockroachDB import; refusing to import over its data. Set DB_IMPORT_FROM_COCKROACH=false to keep it as it is." 5
    fi
    if [ ! -s "$work/target.ids" ]; then
        log "bringing PostgreSQL to the source's $n migrations"
        run_logged "migration to the source's level" "$NAKAMA_BIN" migrate up --limit "$n" --config "$DB_CONFIG"
        tpsql -c "SELECT id FROM migration_info ORDER BY id" >"$work/target.ids"
        cmp -s "$work/source.ids" "$work/target.ids" ||
            die "the source's migrations differ from this build's first $n; refusing to import" 6
    else
        log "retrying an import that did not complete (PostgreSQL is at the source's $n migrations)"
    fi

    # Tables and columns as created on PostgreSQL by the same migrations.
    local tables table cols src_cols rows script=$work/load.sql
    tables=$(tpsql -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' AND table_name <> 'migration_info' ORDER BY table_name")
    {
        echo "BEGIN;"
        echo "SET LOCAL session_replication_role = replica;"
    } >"$script"
    for table in $tables; do
        cols=$(tpsql -c "SELECT string_agg(quote_ident(column_name), ',' ORDER BY ordinal_position) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '$table' AND is_generated = 'NEVER'")
        src_cols=$(spsql -c "SELECT string_agg(quote_ident(column_name), ',' ORDER BY ordinal_position) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '$table' AND is_hidden = 'NO'")
        [ -n "$src_cols" ] || die "table $table is missing in the legacy CockroachDB" 6
        if [ "$(tr , '\n' <<<"$cols" | sort)" != "$(tr , '\n' <<<"$src_cols" | sort)" ]; then
            die "table $table has different columns in CockroachDB ($src_cols) and PostgreSQL ($cols)" 6
        fi
        spsql -c "\\copy (SELECT $cols FROM \"$table\") TO '$work/$table.copy'" ||
            die "could not export table $table from the legacy CockroachDB" 6
        rows=$(wc -l <"$work/$table.copy")
        log "exported $table: $rows rows"
        {
            echo "DELETE FROM \"$table\";"
            echo "\\copy \"$table\" ($cols) FROM '$work/$table.copy'"
            echo "DO \$\$ BEGIN IF (SELECT count(*) FROM \"$table\") <> $rows THEN RAISE EXCEPTION 'row count mismatch in $table'; END IF; END \$\$;"
        } >>"$script"
    done
    cat >>"$script" <<'EOF'
DO $$
DECLARE r record;
BEGIN
    FOR r IN
        SELECT c.table_name, c.column_name,
               pg_get_serial_sequence(quote_ident(c.table_schema) || '.' || quote_ident(c.table_name), c.column_name) AS seq
        FROM information_schema.columns c
        WHERE c.table_schema = 'public' AND (c.column_default LIKE 'nextval(%' OR c.is_identity = 'YES')
    LOOP
        IF r.seq IS NOT NULL THEN
            EXECUTE format('SELECT setval(%L, COALESCE((SELECT max(%I) FROM %I), 0) + 1, false)', r.seq, r.column_name, r.table_name);
        END IF;
    END LOOP;
END $$;
CREATE SCHEMA IF NOT EXISTS nakama_farsi;
CREATE TABLE IF NOT EXISTS nakama_farsi.cockroach_import (
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    source_migrations INTEGER NOT NULL,
    source_tables INTEGER NOT NULL
);
EOF
    echo "INSERT INTO nakama_farsi.cockroach_import (source_migrations, source_tables) VALUES ($n, $(wc -w <<<"$tables"));" >>"$script"
    echo "COMMIT;" >>"$script"
    log "loading $(wc -w <<<"$tables") tables into PostgreSQL in one transaction"
    run_logged "import into PostgreSQL" tpsql -f "$script"
    rm -rf "$work"
    log "CockroachDB data imported at migration level $n; the CockroachDB database is unchanged"
}

main() {
    local mode=${1:-}
    [ $# -gt 0 ] && shift
    case $mode in migrate | server | status) ;; *) die "usage: $0 migrate | server [nakama flags...] | status" 2 ;; esac
    build_address
    log "database ($DB_SOURCE): $DB_REDACTED"
    [ -n "${DB_URL:-}" ] && [ -n "${DB_HOST:-}" ] && log "DB_URL is set, so DB_HOST and the other DB_* fields are ignored"
    [ "$DB_SOURCE" = DB_HOST ] && [ -z "${DB_PASSWORD:-}" ] && log "DB_PASSWORD is empty: connecting without a password"

    if [ "$mode" = migrate ]; then
        case ${DB_MIGRATIONS_ENABLED:-true} in
            true | 1 | yes) ;;
            false | 0 | no) log "DB_MIGRATIONS_ENABLED=${DB_MIGRATIONS_ENABLED}: skipping migrations; the server refuses to start if any are pending"; exit 0 ;;
            *) die "DB_MIGRATIONS_ENABLED must be true or false" 2 ;;
        esac
    fi

    wait_for "$DB_TARGET_HOST" "$DB_TARGET_PORT" database
    write_config

    if [ "$mode" = server ]; then
        exec "$NAKAMA_BIN" --config "$DB_CONFIG" "$@"
    fi
    if [ "$mode" = status ]; then
        run_logged "migration status" "$NAKAMA_BIN" migrate status --config "$DB_CONFIG"
        rm -f "$DB_CONFIG"
        exit 0
    fi

    import_from_cockroach
    run_logged migration "$NAKAMA_BIN" migrate up --config "$DB_CONFIG" "$@"
    rm -f "$DB_CONFIG"
    log "migrations complete"
}

main "$@"
