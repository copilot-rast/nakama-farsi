#!/bin/sh
# End-to-end test of the one-time CockroachDB -> PostgreSQL import done by
# build/nakama-db.sh in the nakama-migrate service, in throwaway containers:
#   1. a CockroachDB at production's state of 2026-10-25 (21 migrations)
#      with representative rows;
#   2. the migrate job imports it into an empty PostgreSQL and finishes the
#      remaining migrations; rows, bytes and the fans_count backfill match;
#   3. a rerun imports nothing, also with the source gone; the source is
#      unchanged;
#   4. a PostgreSQL that already holds data is refused (exit 5);
#   5. the server starts on the imported database and passes its healthcheck.
#
#   docker build -f build/Dockerfile.local -t nakama-farsi:local .
#   scripts/cockroach-import-test.sh [image]
#
# It never connects to anything but the containers it starts.
set -eu

image=${1:-nakama-farsi:local}
crdb_image=${COCKROACH_IMAGE:-cockroachdb/cockroach:v24.1.33}
pg_image=${POSTGRES_IMAGE:-postgres:16.8-alpine}
id=$$
net=import-test-$id
crdb=import-crdb-$id
pg=import-pg-$id
pg2=import-pg2-$id
srv=import-srv-$id
# Credentials that need URL encoding.
pass='p@ss:w/rd%#?'

cleanup() {
    docker rm -f "$crdb" "$pg" "$pg2" "$srv" >/dev/null 2>&1 || true
    docker network rm "$net" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

fail() { echo "FAIL: $*" >&2; exit 1; }
csql() { docker exec "$crdb" cockroach sql --insecure --format=tsv -d nakama -e "$1" | tail -n +2; }
psql_on() { docker exec -e PGPASSWORD="$pass" "$1" psql -X -At -U nakama -d nakama -c "$2"; }
# job <postgres container> <import true|false> [extra docker args]: run the migrate job.
job() {
    target=$1 import=$2
    shift 2
    docker run --rm --network "$net" "$@" \
        -e DB_HOST="$target" -e DB_USER=nakama -e DB_PASSWORD="$pass" -e DB_NAME=nakama \
        -e DB_IMPORT_FROM_COCKROACH="$import" \
        -e COCKROACH_IMPORT_URL="postgresql://root@$crdb:26257/nakama?sslmode=disable" \
        -e DB_WAIT_TIMEOUT_SECONDS=20 \
        --entrypoint /nakama/nakama-db.sh "$image" migrate 2>&1
}
start_pg() {
    docker run -d --name "$1" --network "$net" -e POSTGRES_USER=nakama -e POSTGRES_PASSWORD="$pass" -e POSTGRES_DB=nakama "$pg_image" >/dev/null
    i=0
    until docker exec "$1" pg_isready -q -h 127.0.0.1 -U nakama -d nakama; do
        i=$((i + 1)); [ "$i" -lt 60 ] || fail "PostgreSQL did not become ready"; sleep 1
    done
}

docker network create "$net" >/dev/null
docker run -d --name "$crdb" --network "$net" "$crdb_image" start-single-node --insecure >/dev/null
i=0
until docker exec "$crdb" cockroach sql --insecure -e 'SELECT 1' >/dev/null 2>&1; do
    i=$((i + 1)); [ "$i" -lt 60 ] || fail "CockroachDB did not become ready"; sleep 1
done
start_pg "$pg"

# 1. Production state: Nakama core plus the first Jazebe migration.
docker run --rm --network "$net" --entrypoint /nakama/nakama "$image" \
    migrate up --database.address "root@$crdb:26257" --limit 21 2>&1 | tail -1
[ "$(csql 'SELECT count(*) FROM migration_info')" = 21 ] || fail "could not reproduce the production state"
csql "
INSERT INTO users (id, username, display_name, password, metadata, email) VALUES
  ('00000000-0000-4000-8000-000000000001','owner','مالک برند', '\x00ff10'::BYTES, '{\"note\":\"line\\nnext\",\"q\":\"\\\"x\\\"\"}', 'owner@example.com'),
  ('00000000-0000-4000-8000-000000000002','fan-a','هوادار', NULL, '{}', NULL),
  ('00000000-0000-4000-8000-000000000003','fan-b',NULL, NULL, '{}', NULL),
  ('00000000-0000-4000-8000-000000000004','fan-c',NULL, NULL, '{}', NULL);
INSERT INTO user_device (id, user_id) VALUES ('device-a','00000000-0000-4000-8000-000000000002');
INSERT INTO storage (collection, key, user_id, value, version) VALUES
  ('saves','slot1','00000000-0000-4000-8000-000000000002','{\"level\": 3, \"tab\": \"a\\tb\"}','v1');
INSERT INTO jazebe_brand (id, owner_id, name) VALUES
  ('10000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000001','برند');
INSERT INTO jazebe_planet (id, brand_id, name, status) VALUES
  ('20000000-0000-4000-8000-000000000000','10000000-0000-4000-8000-000000000001','empty','active'),
  ('20000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','one','active'),
  ('20000000-0000-4000-8000-000000000003','10000000-0000-4000-8000-000000000001','three','disabled');
INSERT INTO jazebe_planet_membership (planet_id, user_id) VALUES
  ('20000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000002'),
  ('20000000-0000-4000-8000-000000000003','00000000-0000-4000-8000-000000000002'),
  ('20000000-0000-4000-8000-000000000003','00000000-0000-4000-8000-000000000003'),
  ('20000000-0000-4000-8000-000000000003','00000000-0000-4000-8000-000000000004');" >/dev/null
tables="users user_device storage jazebe_brand jazebe_planet jazebe_planet_membership"
src_counts=''
for t in $tables; do src_counts="$src_counts $t=$(csql "SELECT count(*) FROM $t")"; done
echo "source:$src_counts"

# 2. Import and finish the migrations.
out=$(job "$pg" true) || { echo "$out"; fail "import job failed"; }
echo "$out" | grep -E 'nakama-db:|Successfully applied' | sed 's/^/  /'
echo "$out" | grep -q 'CockroachDB data imported at migration level 21' || fail "import did not report completion"
echo "$out" | grep -qF "$pass" && fail "the password appeared in the job log"
echo "$out" | grep -qF 'p%40ss' && fail "the encoded password appeared in the job log"
total=$(ls "$(dirname "$0")/../migrate/sql/"*.sql | wc -l)
[ "$(psql_on "$pg" 'SELECT count(*) FROM migration_info')" = "$total" ] || fail "PostgreSQL is not at all $total migrations"
dst_counts=''
for t in $tables; do dst_counts="$dst_counts $t=$(psql_on "$pg" "SELECT count(*) FROM $t")"; done
echo "target:$dst_counts"
[ "$src_counts" = "$dst_counts" ] || fail "row counts differ"
[ "$(psql_on "$pg" "SELECT encode(password,'hex') || '|' || display_name || '|' || (metadata->>'note') FROM users WHERE username = 'owner'")" = "$(printf '00ff10|مالک برند|line\nnext')" ] ||
    fail "user values changed: $(psql_on "$pg" "SELECT encode(password,'hex'), display_name, metadata FROM users WHERE username = 'owner'")"
[ "$(psql_on "$pg" "SELECT value->>'tab' FROM storage WHERE key = 'slot1'")" = "$(printf 'a\tb')" ] || fail "storage value changed"
fans=$(psql_on "$pg" "SELECT string_agg(name || '=' || fans_count || ':' || status, ' ' ORDER BY name) FROM jazebe_planet")
echo "fans_count: $fans"
[ "$fans" = "empty=0:active one=1:active three=3:suspended" ] || fail "unexpected fans_count backfill"
[ "$(psql_on "$pg" "SELECT count(*) FROM users WHERE id = '00000000-0000-0000-0000-000000000000'")" = 1 ] || fail "system user missing or duplicated"
[ "$(psql_on "$pg" 'SELECT source_migrations FROM nakama_farsi.cockroach_import')" = 21 ] || fail "import marker missing"

# 3. Reruns import nothing, also without the source; the source is unchanged.
out=$(job "$pg" true) || { echo "$out"; fail "rerun failed"; }
echo "$out" | grep -q 'not importing again' || fail "rerun did not skip the import"
echo "$out" | grep -q '"count":0' || fail "rerun applied migrations"
[ "$(csql 'SELECT count(*) FROM migration_info')" = 21 ] || fail "the source's migrations changed"
[ "$(csql 'SELECT status FROM jazebe_planet WHERE name = '"'three'")" = disabled ] || fail "the source's rows changed"
docker stop "$crdb" >/dev/null
out=$(job "$pg" true) || { echo "$out"; fail "rerun without the source failed"; }
echo "$out" | grep -q 'not importing again' || fail "rerun without the source contacted it"
docker start "$crdb" >/dev/null
echo "rerun: skipped, with and without the source"

# 4. A PostgreSQL that already holds data is never imported over.
start_pg "$pg2"
job "$pg2" false >/dev/null || fail "fresh migration without import failed"
psql_on "$pg2" "INSERT INTO users (id, username) VALUES ('30000000-0000-4000-8000-000000000001','new-user')" >/dev/null
i=0
until docker exec "$crdb" cockroach sql --insecure -e 'SELECT 1' >/dev/null 2>&1; do
    i=$((i + 1)); [ "$i" -lt 60 ] || fail "CockroachDB did not restart"; sleep 1
done
set +e
out=$(job "$pg2" true); status=$?
set -e
[ "$status" = 5 ] || { echo "$out"; fail "import over existing data exited $status, not 5"; }
echo "$out" | grep -q 'refusing to import over its data' || fail "no refusal message"
[ "$(psql_on "$pg2" "SELECT count(*) FROM users WHERE username = 'new-user'")" = 1 ] || fail "existing PostgreSQL data changed"
echo "existing PostgreSQL data: refused (exit 5), unchanged"

# 5. The server starts on the imported database.
docker run -d --name "$srv" --network "$net" --entrypoint /nakama/nakama-db.sh \
    -e DB_HOST="$pg" -e DB_USER=nakama -e DB_PASSWORD="$pass" -e DB_NAME=nakama \
    "$image" server --name importtest >/dev/null
i=0
until docker exec "$srv" /nakama/nakama healthcheck >/dev/null 2>&1; do
    i=$((i + 1))
    [ "$i" -lt 60 ] || { docker logs "$srv" 2>&1 | tail -20; fail "server did not become healthy"; }
    sleep 1
done
docker logs "$srv" 2>&1 | grep -q '"Database information","version":"PostgreSQL' || fail "server is not on PostgreSQL"
docker logs "$srv" 2>&1 | grep -qF "$pass" && fail "the password appeared in the server log"
echo "server: healthy on $(docker logs "$srv" 2>&1 | sed -n 's/.*"Database information","version":"\(PostgreSQL [0-9.]*\).*/\1/p' | head -1)"

echo "PASS: CockroachDB import into PostgreSQL"
