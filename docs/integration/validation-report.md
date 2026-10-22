# Validation report (2026-10-22)

## Executed

- Original Jazebe repositories were clean at `dfdc7f8133c1335f5cf3807cdbe1bba17aa71d89` and `d6b54b50b0a73a7137fb69494f07cd12d513f12f`; destination baseline was `f9a48a22`.
- `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate` passes. It compiles 18 native RPCs and the fan reconciliation worker, checks identity, planet/mission validation, brand roles, publish requirements and operator denial, and confirms seven migrations are embedded and parsed.
- Source inventory generation succeeded and counted 389 HTTP route declarations.
- `go build .` passed after the native module and migration changes. The generated binary was removed afterward.
- `go build ./integrations/jazebe/reconcile` and targeted `go vet` passed. The generated executable was removed.
- MCP `go test ./...` passes with an official SDK client over in-memory and Streamable HTTP transports, including 15-tool discovery, account invocation, selected brand/planet RPC adapters, resource retrieval, prompt discovery, and HTTP denial tests. It uses a mock Nakama upstream.
- `git diff --check` passed; both Jazebe source Git working trees remain clean.

## Unverified or pre-existing failures

- Docker is unavailable, so the Linux Go plugin build, Nakama startup, SQL migration, and live RPC calls could not run.
- Baseline `go test ./...` failed from Go 1.27 vet errors in bundled Lua code, Console gzip mismatches, and server tests requiring unavailable PostgreSQL on `127.0.0.1:5432`.
- `go build ./...` includes a pre-existing sample main package without a `main` function; use `go build .` for the server.
- MCP transitive dependencies required a direct download of `segmentio/encoding`, then normal `go mod tidy`; live Nakama interoperability and compliant remote OAuth remain unverified.
- Jazebe source tests were not run; source dependencies are absent. No complete Jazebe controller group has parity. Planet creation now begins in `setup` with a settings row, but brand product/member roles, mission action/rewards, audit coverage, and most source APIs remain incomplete.
