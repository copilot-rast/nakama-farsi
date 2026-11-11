#!/bin/bash
set -e
HERE=$(cd "$(dirname "$0")" && pwd)
R=$(cd "$HERE/../.." && pwd)
RUN=${RUN:-$HERE/.run}
BIN=${BIN:-$R/.local/cluster/nakama}
MODS=${MODS:-$R/.local/cluster/modules}
REDIS_PORT=${REDIS_PORT:-6495}
LEVEL=${LEVEL:-warn}
CLDB=${CL_DB:-nakama_cluster}
ENVF=${E2E_ENV_FILE:-$HOME/.config/nakama-farsi/mac-dev.env}
cmd=$1; shift
mkdir -p "$RUN"
start() {
  local i=$1
  . "$ENVF"
  local A=${NAKAMA_DATABASE_ADDRESS#postgres://}; A=${A#postgresql://}
  local UP=${A%%@*}
  local DB="$UP@127.0.0.1:5481/$CLDB?sslmode=disable"
  local B=$((7600+i*10))
  mkdir -p "$RUN/n$i"
  nohup "$BIN" --name "node$i" --database.address "$DB" --database.max_open_conns 30 \
    --runtime.path "$MODS" --data_dir "$RUN/n$i" \
    --socket.port $((B+1)) --console.port $((B+2)) --metrics.prometheus_port $((B+4)) \
    --socket.server_key "$SERVER_KEY" --console.username "$CONSOLE_USER" --console.password "$CONSOLE_PASSWORD" \
    --session.token_expiry_sec 7200 --matchmaker.interval_sec 1 --matchmaker.max_intervals 5 --matchmaker.max_tickets 3 \
    --runtime.env "ECONOMY_SECRET_KEY=$ECONOMY_SECRET_KEY" --runtime.env "STARGAME_SERVER_KEY=$SERVER_KEY" \
    --runtime.env "ECONOMY_PUBLIC_BASE_URL=http://127.0.0.1:$((B+1))" \
    --cluster.enabled=true --cluster.redis_address 127.0.0.1:$REDIS_PORT --cluster.prefix "${CL_PREFIX:-nk}" \
    --cluster.port $((B+3)) --cluster.advertise_address 127.0.0.1 --cluster.secret "$ECONOMY_SECRET_KEY" \
    --cluster.matchmaker_shards ${MM_SHARDS:-8} --cluster.matchmaker_shard_property "${MM_PROP:-}" \
    --logger.level "$LEVEL" > "$RUN/n$i/out.log" 2>&1 &
  echo $! > "$RUN/n$i/pid"
}
for i in "$@"; do
  case $cmd in
    start) start "$i" ;;
    stop) P=$(cat "$RUN/n$i/pid" 2>/dev/null || true); [ -n "$P" ] && kill "$P" 2>/dev/null || true ;;
    kill) P=$(cat "$RUN/n$i/pid" 2>/dev/null || true); [ -n "$P" ] && kill -9 "$P" 2>/dev/null || true ;;
    status) P=$(cat "$RUN/n$i/pid" 2>/dev/null || true); [ -n "$P" ] && kill -0 "$P" 2>/dev/null && echo "node$i up $P" || echo "node$i down" ;;
  esac
done
if [ "$cmd" = start ]; then
  for i in "$@"; do
    n=0; until curl -s -o /dev/null "http://127.0.0.1:$((7600+i*10+1))/healthcheck" || [ $n -gt 60 ]; do sleep 0.5; n=$((n+1)); done
  done
fi
