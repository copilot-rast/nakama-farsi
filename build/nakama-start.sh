#!/bin/bash
set -euo pipefail

DIR=${NAKAMA_DIR:-/nakama}
DB_SCRIPT=${NAKAMA_DB_SCRIPT:-$DIR/nakama-db.sh}

log() { printf 'nakama-start: %s\n' "$*" >&2; }

case ${NAKAMA_MIGRATE_ON_START:-true} in
    true | 1 | yes) "$DB_SCRIPT" migrate ;;
    false | 0 | no) log "migrations skipped (NAKAMA_MIGRATE_ON_START=false)" ;;
    *) log "ERROR: NAKAMA_MIGRATE_ON_START must be true or false"; exit 2 ;;
esac

node_name() {
    if [ "${NAKAMA_CLUSTER_ENABLED:-false}" != "true" ]; then
        printf '%s' "${NAKAMA_NODE_NAME:-nakama1}"
        return
    fi
    if [ -n "${NAKAMA_NODE_NAME:-}" ]; then
        printf '%s' "$NAKAMA_NODE_NAME"
        return
    fi
    local host
    host=$(hostname)
    if [ ${#host} -le 16 ]; then
        printf '%s' "$host"
    else
        printf 'n%s' "$(printf '%s' "$host" | sha1sum | cut -c1-15)"
    fi
}

advertise() {
    if [ -n "${NAKAMA_CLUSTER_ADVERTISE_ADDRESS:-}" ]; then
        printf '%s' "$NAKAMA_CLUSTER_ADVERTISE_ADDRESS"
    else
        hostname -i | cut -d' ' -f1
    fi
}

args=(
    --name "$(node_name)"
    --logger.level "${NAKAMA_LOG_LEVEL:-INFO}"
    --session.token_expiry_sec "${NAKAMA_SESSION_TOKEN_EXPIRY_SEC:-7200}"
    --metrics.prometheus_port "${NAKAMA_METRICS_PORT:-9100}"
    --socket.server_key "${NAKAMA_SERVER_KEY:?NAKAMA_SERVER_KEY is required}"
    --session.encryption_key "${NAKAMA_SESSION_ENCRYPTION_KEY:?NAKAMA_SESSION_ENCRYPTION_KEY is required}"
    --runtime.http_key "${NAKAMA_RUNTIME_HTTP_KEY:?NAKAMA_RUNTIME_HTTP_KEY is required}"
    --console.username "${NAKAMA_CONSOLE_USERNAME:?NAKAMA_CONSOLE_USERNAME is required}"
    --console.password "${NAKAMA_CONSOLE_PASSWORD:?NAKAMA_CONSOLE_PASSWORD is required}"
    --console.signing_key "${NAKAMA_CONSOLE_SIGNING_KEY:?NAKAMA_CONSOLE_SIGNING_KEY is required}"
)

if [ -n "${NAKAMA_REFRESH_ENCRYPTION_KEY:-}" ]; then
    args+=(--session.refresh_encryption_key "$NAKAMA_REFRESH_ENCRYPTION_KEY")
fi

if [ "${NAKAMA_CLUSTER_ENABLED:-false}" = "true" ]; then
    args+=(
        --cluster.enabled=true
        --cluster.redis_address "${NAKAMA_CLUSTER_REDIS_ADDRESS:?NAKAMA_CLUSTER_REDIS_ADDRESS is required}"
        --cluster.redis_password "${NAKAMA_CLUSTER_REDIS_PASSWORD:-}"
        --cluster.port "${NAKAMA_CLUSTER_PORT:-7353}"
        --cluster.advertise_address "$(advertise)"
        --cluster.secret "${NAKAMA_CLUSTER_SECRET:-$NAKAMA_SESSION_ENCRYPTION_KEY}"
        --cluster.matchmaker_shards "${NAKAMA_CLUSTER_MATCHMAKER_SHARDS:-16}"
        --cluster.matchmaker_shard_property "${NAKAMA_CLUSTER_MATCHMAKER_SHARD_PROPERTY:-}"
    )
fi

runtime_env=(
    ECONOMY_SECRET_KEY ECONOMY_PUBLIC_BASE_URL ECONOMY_WEB_REDIRECT_URL ECONOMY_DEEP_LINK
    STORAGE_DRIVER STORAGE_S3_ENDPOINT STORAGE_S3_BUCKET STORAGE_S3_REGION
    STORAGE_S3_ACCESS_KEY STORAGE_S3_SECRET_KEY STORAGE_S3_PATH_STYLE
    STORAGE_S3_PUBLIC_BASE_URL STORAGE_S3_PREFIX STARGAME_SEED BASE_UPDATES_API_URL
)
args+=(--runtime.env "ECONOMY_UPLOAD_DIR=${ECONOMY_UPLOAD_DIR:-$DIR/uploads}")
args+=(--runtime.env "STARGAME_SERVER_KEY=$NAKAMA_SERVER_KEY")
: "${ECONOMY_SECRET_KEY:=$NAKAMA_SESSION_ENCRYPTION_KEY}" "${STORAGE_DRIVER:=local}" "${STORAGE_S3_REGION:=us-east-1}" "${STORAGE_S3_PATH_STYLE:=true}" "${STARGAME_SEED:=true}"
for name in "${runtime_env[@]}"; do
    args+=(--runtime.env "$name=${!name:-}")
done
while IFS='=' read -r name _; do
    case $name in
        GAME_* | BLOCKWORLD_*) args+=(--runtime.env "$name=${!name}") ;;
    esac
done < <(env)

log "starting node $(node_name)"
exec "$DB_SCRIPT" server "${args[@]}" "$@"
