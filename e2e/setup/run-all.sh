#!/bin/zsh
set -e
HERE="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${E2E_ENV_FILE:-$HOME/.config/nakama-farsi/mac-dev.env}"
NODE="${NODE:-$HOME/.nvm/versions/node/v22.18.0/bin/node}"
PSQL="${PSQL:-/opt/homebrew/opt/postgresql@16/bin/psql}"
set -a
. "$ENV_FILE"
set +a
export PGPASSWORD="$DB_PASSWORD"
DB_HOST="${E2E_DB_HOST:-127.0.0.1}"
DB_PORT="${E2E_DB_PORT:-5481}"
DB_NAME="${E2E_DB_NAME:-nakama}"
OUT="${TMPDIR:-/tmp}/nakama-e2e-setup"
mkdir -p "$OUT"
q() { "$PSQL" -h "$DB_HOST" -p "$DB_PORT" -U nakama -d "$DB_NAME" -At -v ON_ERROR_STOP=1 "$@"; }

q -f "$HERE/sql/snapshot.sql" > "$OUT/before.txt"
q -f "$HERE/sql/backup.sql" > /dev/null 2>&1
rc=0
for s in walk modes welcome final; do
  echo "== $s"
  "$NODE" "$HERE/$s.cjs" | grep -E '^(PASS|FAIL|shot)' || rc=1
  [ "${pipestatus[1]}" = "0" ] || rc=1
done
q -f "$HERE/sql/restore.sql" > /dev/null
q -c "DROP SCHEMA zz_setup_backup CASCADE" > /dev/null 2>&1
q -f "$HERE/sql/snapshot.sql" > "$OUT/after.txt"
if diff -q "$OUT/before.txt" "$OUT/after.txt" > /dev/null; then echo "PASS gateway, provider, language and currency rows restored byte for byte"; else echo "FAIL config rows differ after restore (see $OUT)"; rc=1; fi
exit $rc
