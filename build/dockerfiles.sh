#!/bin/sh
set -eu

cd "$(dirname "$0")"

GO_REF='golang:1.27.3-bookworm@sha256:dc0963f32053e4bc451d24d87e3892d42d01021b61f093e98eb95c0352bb84b0'
DEBIAN_REF='debian:bookworm-slim@sha256:734a72b3d6142a71993b4bc580ca01d8ca687f77a2c422d96a961d3376b179bc'

FOREIGN_GO_IMAGE="docker.io/library/$GO_REF"
FOREIGN_DEBIAN_IMAGE="docker.io/library/$DEBIAN_REF"
IRAN_GO_IMAGE="docker.arvancloud.ir/library/$GO_REF"
IRAN_DEBIAN_IMAGE="docker.arvancloud.ir/library/$DEBIAN_REF"

if [ -f iran-images.env ]; then
  . ./iran-images.env
fi

render() {
  awk -v go="$1" -v deb="$2" '
    /^ARG NEXUS_DOCKER_HOST=/ { next }
    /^ARG GO_IMAGE=/ { print "ARG GO_IMAGE=" go; next }
    /^ARG DEBIAN_IMAGE=/ { print "ARG DEBIAN_IMAGE=" deb; next }
    { print }
  ' Dockerfile.local
}

mode="${1:-write}"
status=0
for target in foreign iran; do
  if [ "$target" = foreign ]; then
    out=$(render "$FOREIGN_GO_IMAGE" "$FOREIGN_DEBIAN_IMAGE")
  else
    out=$(render "$IRAN_GO_IMAGE" "$IRAN_DEBIAN_IMAGE")
  fi
  file="Dockerfile.$target"
  if [ "$mode" = --check ]; then
    if [ ! -f "$file" ] || [ "$(cat "$file")" != "$out" ]; then
      echo "dockerfiles: build/$file is out of date; run sh build/dockerfiles.sh" >&2
      status=1
    fi
  else
    printf '%s\n' "$out" > "$file"
    echo "dockerfiles: wrote build/$file"
  fi
done
exit $status
