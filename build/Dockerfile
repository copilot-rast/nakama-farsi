## Copyright 2018 The Nakama Authors
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
## http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.

ARG NEXUS_DOCKER_HOST=docker.arvancloud.ir
ARG GO_IMAGE=${NEXUS_DOCKER_HOST}/library/golang:1.27.3-bookworm@sha256:dc0963f32053e4bc451d24d87e3892d42d01021b61f093e98eb95c0352bb84b0
ARG DEBIAN_IMAGE=${NEXUS_DOCKER_HOST}/library/debian:bookworm-slim@sha256:734a72b3d6142a71993b4bc580ca01d8ca687f77a2c422d96a961d3376b179bc

FROM ${GO_IMAGE} AS builder

ARG COMMIT
ARG VERSION
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV CGO_ENABLED=1
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /go/build/nakama
COPY . .
RUN go build -o /go/build-out/nakama -trimpath -mod=vendor -gcflags "-trimpath $PWD" -asmflags "-trimpath $PWD" -ldflags "-s -w -X main.version=$VERSION -X main.commitID=$COMMIT"

FROM ${DEBIAN_IMAGE}

ARG VERSION

LABEL maintainer="Heroic Labs"
LABEL contact="support@heroiclabs.com"
LABEL version=$VERSION
LABEL variant=nakama
LABEL description="Distributed server for social and realtime games and apps."

ENV DEBIAN_FRONTEND=noninteractive

ARG APT_MIRROR=https://nexus.rastar.ir/repository/apt-proxy-debian
ARG APT_SECURITY_MIRROR=https://nexus.rastar.ir/repository/apt-proxy-debian-security

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN --mount=type=secret,id=nexus_cred bash -ec '\
    test -s /run/secrets/nexus_cred || { echo "build secret nexus_cred (NEXUS_CRED) is missing" >&2; exit 1; }; \
    cred=$(cat /run/secrets/nexus_cred); user=${cred%%:*}; pass=${cred#*:}; \
    host=${APT_MIRROR#https://}; host=${host%%/*}; \
    printf "machine %s\nlogin %b\npassword %b\n" "$host" "${user//%/\\x}" "${pass//%/\\x}" > /etc/apt/auth.conf.d/nexus.conf; \
    chmod 600 /etc/apt/auth.conf.d/nexus.conf; \
    printf "Types: deb\nURIs: %s\nSuites: bookworm bookworm-updates\nComponents: main\nSigned-By: /usr/share/keyrings/debian-archive-keyring.gpg\n\nTypes: deb\nURIs: %s\nSuites: bookworm-security\nComponents: main\nSigned-By: /usr/share/keyrings/debian-archive-keyring.gpg\n" "$APT_MIRROR" "$APT_SECURITY_MIRROR" > /etc/apt/sources.list.d/debian.sources; \
    mkdir -p /nakama/data/modules; \
    apt-get update; \
    apt-get -y upgrade; \
    apt-get install -y --no-install-recommends ca-certificates tzdata iproute2 tini; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/* /etc/apt/auth.conf.d/nexus.conf

WORKDIR /nakama/
COPY --from=builder "/go/build-out/nakama" /nakama/
EXPOSE 7349 7350 7351

ENTRYPOINT ["tini", "--", "/nakama/nakama"]

HEALTHCHECK --interval=30s --timeout=10s \
  CMD /nakama/nakama healthcheck || exit 1
