#!/bin/sh
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)
FILE="$ROOT/build/upstream.json"

usage() {
    cat >&2 <<'EOF'
usage: build/upstream-sync.sh record [remote] [branch]
       build/upstream-sync.sh show

record  writes build/upstream.json with the newest commit of <remote>/<branch>
        that HEAD already contains (git merge-base HEAD <remote>/<branch>).
        Run it after merging the parent, then commit the file.
        Defaults: the remote and branch already in build/upstream.json,
        else "upstream" and "master".
show    prints build/upstream.json and how many parent commits are not
        merged yet, as of the last git fetch.
EOF
    exit 2
}

field() {
    [ -f "$FILE" ] || return 0
    sed -n "s/.*\"$1\": *\"\\([^\"]*\\)\".*/\\1/p" "$FILE" | head -n 1
}

repo_of() {
    printf '%s' "$1" | sed -e 's#\.git$##' -e 's#^.*github\.com[:/]##' -e 's#^.*gitlab\.com[:/]##'
}

provider_of() {
    case $1 in
        *gitlab*) printf gitlab ;;
        *) printf github ;;
    esac
}

json_escape() {
    printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g'
}

cmd=${1:-}
case $cmd in
    record)
        remote=${2:-$(field remote)}
        remote=${remote:-upstream}
        branch=${3:-$(field branch)}
        branch=${branch:-master}
        ref="$remote/$branch"
        cd "$ROOT"
        git rev-parse --verify --quiet "$ref^{commit}" >/dev/null || {
            echo "upstream-sync: $ref is not known here; run: git fetch $remote $branch" >&2
            exit 1
        }
        base=$(git merge-base HEAD "$ref") || {
            echo "upstream-sync: HEAD and $ref share no commit; merge the parent first" >&2
            exit 1
        }
        url=$(git remote get-url "$remote")
        date=$(git log -1 --format=%cI "$base")
        subject=$(git log -1 --format=%s "$base")
        now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
        cat >"$FILE" <<EOF
{
  "provider": "$(provider_of "$url")",
  "repo": "$(json_escape "$(repo_of "$url")")",
  "remote": "$(json_escape "$remote")",
  "branch": "$(json_escape "$branch")",
  "commit": "$base",
  "date": "$date",
  "subject": "$(json_escape "$subject")",
  "recordedAt": "$now"
}
EOF
        echo "upstream-sync: recorded $(repo_of "$url") $branch at $base ($date)" >&2
        behind=$(git rev-list --count "$base..$ref")
        echo "upstream-sync: $behind commit(s) of $ref are not merged yet" >&2
        ;;
    show)
        [ -f "$FILE" ] || { echo "upstream-sync: $FILE is missing; run record" >&2; exit 1; }
        cat "$FILE"
        cd "$ROOT"
        remote=$(field remote)
        branch=$(field branch)
        base=$(field commit)
        if [ -n "$remote" ] && git rev-parse --verify --quiet "$remote/$branch^{commit}" >/dev/null; then
            echo "upstream-sync: $(git rev-list --count "$base..$remote/$branch") commit(s) of $remote/$branch are not merged yet (as of the last fetch)" >&2
        fi
        ;;
    *) usage ;;
esac
