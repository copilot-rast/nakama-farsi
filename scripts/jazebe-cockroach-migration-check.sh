#!/bin/sh
# Apply the embedded Nakama and Jazebe migrations to disposable CockroachDB
# databases, the production engine, and check the results:
#   A. a fresh database migrates completely, and a second run applies nothing;
#   B. the production state of 2026-10-25 (20 upstream migrations plus
#      20261022120000-jazebe-brand-planet, holding planets and members)
#      upgrades, keeps its rows and backfills fans_count from active members.
#
#   scripts/jazebe-cockroach-migration-check.sh /path/to/nakama
#
# It starts its own throwaway container and never connects to another server.
set -eu

nakama=${1:?usage: $0 /path/to/nakama}
image=${COCKROACH_IMAGE:-cockroachdb/cockroach:v24.1.33}
name=jazebe-crdb-check-$$
port=${COCKROACH_CHECK_PORT:-26290}

cleanup() { docker rm -f "$name" >/dev/null 2>&1 || true; }
trap cleanup EXIT INT TERM

fail() { echo "FAIL: $*" >&2; exit 1; }
sql() { docker exec "$name" cockroach sql --insecure --format=tsv -d "$1" -e "$2" | tail -n +2; }
migrate() { db=$1; shift; "$nakama" migrate up --database.address "root@127.0.0.1:$port/$db" "$@" 2>&1 | tail -1; }
applied() { sed -n 's/.*"count":\([0-9]*\).*/\1/p'; }

docker run -d --name "$name" -p "127.0.0.1:$port:26257" "$image" start-single-node --insecure >/dev/null
i=0
until docker exec "$name" cockroach sql --insecure -e 'SELECT 1' >/dev/null 2>&1; do
    i=$((i + 1)); [ "$i" -lt 60 ] || fail "CockroachDB did not become ready"; sleep 1
done
echo "engine: $(sql defaultdb 'SELECT version()' | cut -d' ' -f1-3)"
embedded=$(ls "$(dirname "$0")/../migrate/sql/"*.sql | wc -l)

# A. Fresh database.
sql defaultdb 'CREATE DATABASE fresh' >/dev/null
out=$(migrate fresh); echo "A first run:  $out"
echo "$out" | grep -q 'Successfully applied migration' || fail "fresh migration failed"
total=$(echo "$out" | applied)
out=$(migrate fresh); echo "A second run: $out"
[ "$(echo "$out" | applied)" = 0 ] || fail "second run re-applied migrations"
[ "$(sql fresh 'SELECT count(*) FROM migration_info')" = "$total" ] || fail "migration_info does not record $total migrations"
[ "$embedded" = "$total" ] || fail "applied $total of $embedded migration files; is the binary built from this tree?"
col=$(sql fresh "SELECT data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name='jazebe_planet' AND column_name='fans_count'")
echo "A fans_count: $col"
echo "$col" | grep -q 'bigint	NO	0' || fail "unexpected fans_count definition"
[ "$(sql fresh "SELECT count(*) FROM information_schema.table_constraints WHERE table_name='jazebe_user_report' AND constraint_name='check_subject_type'")" = 0 ] ||
    fail "the superseded report subject check survived"

# B. Production state: Nakama core plus the first Jazebe migration only.
sql defaultdb 'CREATE DATABASE prodstate' >/dev/null
out=$(migrate prodstate --limit 21); echo "B prepare:    $out"
[ "$(echo "$out" | applied)" = 21 ] || fail "could not reproduce the production state"
sql prodstate "
INSERT INTO users (id, username) VALUES
  ('00000000-0000-4000-8000-000000000001','crdb-owner'),
  ('00000000-0000-4000-8000-000000000002','crdb-fan-a'),
  ('00000000-0000-4000-8000-000000000003','crdb-fan-b'),
  ('00000000-0000-4000-8000-000000000004','crdb-fan-c');
INSERT INTO jazebe_brand (id, owner_id, name) VALUES
  ('10000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000001','brand');
INSERT INTO jazebe_planet (id, brand_id, name, status) VALUES
  ('20000000-0000-4000-8000-000000000000','10000000-0000-4000-8000-000000000001','empty','active'),
  ('20000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','one','active'),
  ('20000000-0000-4000-8000-000000000003','10000000-0000-4000-8000-000000000001','three','disabled');
INSERT INTO jazebe_planet_membership (planet_id, user_id) VALUES
  ('20000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000002'),
  ('20000000-0000-4000-8000-000000000003','00000000-0000-4000-8000-000000000002'),
  ('20000000-0000-4000-8000-000000000003','00000000-0000-4000-8000-000000000003'),
  ('20000000-0000-4000-8000-000000000003','00000000-0000-4000-8000-000000000004');
INSERT INTO jazebe_planet_membership (planet_id, user_id) VALUES
  ('20000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000002')
  ON CONFLICT DO NOTHING;" >/dev/null
out=$(migrate prodstate); echo "B upgrade:    $out"
echo "$out" | grep -q 'Successfully applied migration' || fail "upgrade failed"
[ "$(sql prodstate 'SELECT count(*) FROM migration_info')" = "$total" ] || fail "upgrade did not reach $total migrations"
fans=$(sql prodstate "SELECT name || '=' || fans_count || ':' || status FROM jazebe_planet ORDER BY name" | tr '\n' ' ')
echo "B fans_count: $fans"
[ "$fans" = "empty=0:active one=1:active three=3:suspended " ] || fail "unexpected backfill: $fans"
[ "$(sql prodstate 'SELECT count(*) FROM jazebe_planet_membership WHERE is_active')" = 4 ] || fail "memberships changed"
out=$(migrate prodstate); echo "B rerun:      $out"
[ "$(echo "$out" | applied)" = 0 ] || fail "rerun re-applied migrations"
[ "$(sql prodstate "SELECT string_agg(name || '=' || fans_count, ' ' ORDER BY name) FROM jazebe_planet")" = "empty=0 one=1 three=3" ] ||
    fail "rerun changed fans_count"

echo "PASS: $total migrations on $(sql defaultdb 'SELECT version()' | cut -d' ' -f1-3)"
