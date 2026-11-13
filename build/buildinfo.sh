#!/bin/sh
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)
PKG=github.com/heroiclabs/nakama/v3/internal/buildinfo
FILE="$ROOT/build/upstream.json"

commit=${GIT_COMMIT:-}
branch=${GIT_BRANCH:-}
dirty=${GIT_DIRTY:-}
ahead=${GIT_OWN_AHEAD:-}
date=${BUILD_DATE:-}

g() { git -C "$ROOT" -c safe.directory="$ROOT" "$@"; }

if command -v git >/dev/null 2>&1 && g rev-parse --git-dir >/dev/null 2>&1; then
    [ -n "$commit" ] || commit=$(g rev-parse HEAD 2>/dev/null || true)
    if [ -z "$branch" ]; then
        branch=$(g rev-parse --abbrev-ref HEAD 2>/dev/null || true)
        [ "$branch" != HEAD ] || branch=""
    fi
    if [ -z "$dirty" ]; then
        if g status --porcelain --untracked-files=no 2>/dev/null | grep -v '^ D ' | grep -q .; then dirty=true; else dirty=false; fi
    fi
    if [ -z "$ahead" ] && [ -f "$FILE" ]; then
        base=$(sed -n 's/.*"commit": *"\([0-9a-f]*\)".*/\1/p' "$FILE" | head -n 1)
        if [ -n "$base" ]; then ahead=$(g rev-list --count "$base..HEAD" 2>/dev/null || true); fi
    fi
fi

[ -n "$date" ] || date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
commit=$(printf '%s' "$commit" | tr -cd '0-9a-fA-F')
branch=$(printf '%s' "$branch" | tr -cd 'A-Za-z0-9._/-')
date=$(printf '%s' "$date" | tr -cd '0-9TZ:+.-')
ahead=$(printf '%s' "$ahead" | tr -cd '0-9')
case $dirty in true | 1 | yes) dirty=true ;; *) dirty=false ;; esac
upstream=""
[ ! -f "$FILE" ] || upstream=$(base64 <"$FILE" | tr -d '\n')
short=$(printf '%s' "$commit" | cut -c1-7)

out="-X $PKG.Commit=$commit -X $PKG.Branch=$branch -X $PKG.Date=$date -X $PKG.Dirty=$dirty -X $PKG.OwnAhead=$ahead -X $PKG.Upstream=$upstream"
[ -z "$short" ] || out="$out -X main.commitID=$short"
printf '%s\n' "$out"
