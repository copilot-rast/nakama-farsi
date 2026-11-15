#!/bin/bash
# Tests for build/nakama-start.sh, the image entrypoint that turns env into
# Nakama flags. A stub replaces nakama-db.sh and prints the arguments.
#
#   scripts/nakama-start-test.sh
set -u

root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

failures=0
pass() { echo "ok   $1"; }
fail() { echo "FAIL $1"; failures=$((failures + 1)); }

cat >"$tmp/nakama-db.sh" <<'EOF'
#!/bin/bash
[ "$1" = server ] && { shift; echo "ARGS: $*"; }
exit 0
EOF
chmod +x "$tmp/nakama-db.sh"

run() {
    env -i PATH="$PATH" NAKAMA_DIR="$tmp" NAKAMA_DB_SCRIPT="$tmp/nakama-db.sh" \
        NAKAMA_SERVER_KEY=k NAKAMA_SESSION_ENCRYPTION_KEY=e NAKAMA_RUNTIME_HTTP_KEY=h \
        NAKAMA_CONSOLE_USERNAME=u NAKAMA_CONSOLE_PASSWORD=p NAKAMA_CONSOLE_SIGNING_KEY=s \
        "$@" bash "$root/build/nakama-start.sh" 2>/dev/null
}

# Socket messages carry RPCs whose bodies may be up to 16 KB; Nakama's default
# 4096-byte socket limit would close the connection on them.
out=$(run)
case $out in
    *"--socket.max_message_size_bytes 32768"*) pass "socket message limit defaults to 32768" ;;
    *) fail "socket message limit default: $out" ;;
esac

out=$(run NAKAMA_SOCKET_MAX_MESSAGE_SIZE_BYTES=65536)
case $out in
    *"--socket.max_message_size_bytes 65536"*) pass "socket message limit follows env" ;;
    *) fail "socket message limit env: $out" ;;
esac

[ "$failures" -eq 0 ] || { echo "$failures failure(s)"; exit 1; }
