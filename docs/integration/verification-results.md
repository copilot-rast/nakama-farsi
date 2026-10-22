# Verification results

## Account restriction and route audit checkpoint (2026-10-23)

PASS: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate ./scripts/jazebe_route_audit -count=1`; `go vet` for the same packages; `go build .`; `go run ./scripts/jazebe_route_audit` (389 source routes, full source handler check, 28 behaviorally classified); `go run ./scripts/jazebe_route_audit -offline` (mapping structure only); nested `integrations/mcp` `go test ./... -count=1` and `go vet ./...`. Pure tests exercise reason bounds, local revocation failure reporting, super-operator protection and worker configuration. PostgreSQL tests for atomic account state, audit, ban provenance, pending retry, concurrent restriction/idempotent side effects, authorization and constellation projection are added to the CI job but **NOT_RUN** here because no disposable PostgreSQL was supplied. CI job itself, live Nakama and cross-node token behavior are **NOT_RUN**. Full root `go test ./... -count=1` was run and **FAIL**: the existing `internal/gopher-lua` nonconstant-format vet diagnostics, stale console `.gz` assets, and server tests requiring PostgreSQL at `127.0.0.1:5432` recur. The Jazebe packages pass within that run. Both Jazebe Git trees were clean; no Nakama core Go file was changed.

## Operator account read checkpoint (2026-10-23)

PASS: focused native/reconcile/migrate `go test -count=1`, matching `go vet`, root `go build`, MCP module tests/vet and `git diff --check`. New pure tests cover operator account list filter bounds. PostgreSQL account composition, profile/report workflow tests, Nakama RPC runtime, full root regression and CI migration workflow are **NOT_RUN**.

## Profile report and public scope checkpoint (2026-10-22)

After operator report queue/review/resolve were registered, PASS again: focused native/reconcile/migrate `go test -count=1`, matching `go vet`, root `go build -o $env:TEMP\nakama-jazebe-build.exe .` and `git diff --check`. Operator permission and list-filter tests pass. `TestProfileReportPostgresConstraints`, `TestProfileReportPostgresWorkflow` and the disposable PostgreSQL CI job were added but **NOT_RUN**; they will check report uniqueness, reviewing-state repeat reports, cross-planet scope, handler replay, audited operator review and resolution. Live Nakama remains **NOT_RUN**.

PASS: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`; `go vet ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; `go build -o $env:TEMP\nakama-jazebe-build.exe .`; nested MCP `go test ./... -count=1` and `go vet ./...`. Pure profile-report validation passes. PostgreSQL migration 31, report uniqueness/concurrency, profile persistence, live Nakama and full root regression are **NOT_RUN**. No SQL-level success is inferred from migration parser tests.

## Game and referral checkpoint (2026-10-22)

PASS: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; `go vet ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; root `go build .`; `git diff --check`. PASS in nested `integrations/mcp`: `go test ./...`; `go vet ./...`. The migration embed/parser test recognizes 29 Jazebe SQL files. Pure tests cover box/tap scoring, server prize bounds, game config guards, referral qualification/window/token rules, invite draft validation and worker flag validation. MCP tests discover 108 tools and invoke game/referral reads through an authenticated mock Nakama upstream. NOT_RUN: PostgreSQL migrations or service queries, concurrent game/referral transactions, live Nakama RPC loading, live HTTP landing, live MCP-to-Nakama, and full root `go test ./...`. No PostgreSQL/container service became available during this batch. The prior full-root gopher-lua vet, console asset and PostgreSQL failures remain baseline blockers.


## Medal acquisition and backfill checkpoint (2026-10-22)

PASS: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`; `go vet ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; nested MCP `go test ./... -count=1` and `go vet ./...`; root `go build .`; `git diff --check`. The migration parser recognizes migrations 26 and 27. Unit tests cover special-event key matching and explicit rejection of game/referral eligibility without source ledgers. NOT_RUN: PostgreSQL migration and reward concurrency tests, live Nakama RPCs, live MCP-to-Nakama calls and full root regression suite. No local PostgreSQL service or `psql`/`pg_ctl`/`postgres`/`docker` executable was found; TCP 127.0.0.1:5432 is closed. The previous full root suite failed on gopher-lua vet, stale console gzip assets and absent PostgreSQL; this batch did not change those core areas. Both Jazebe sources remain Git-clean.

## Platform policy checkpoint (2026-10-22)

`go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`, corresponding `go vet`, MCP module `go test ./... -count=1` and `go vet ./...`, root `go build .`, and `git diff --check` pass with 141 RPCs, 25 migrations and 98 MCP tools. Pure policy tests reject unknown fields, null or out-of-range values, excess decimal precision and ordinary-operator writes; valid source numeric-string coercion is covered. The migration parser sees the policy SQL. PostgreSQL migration, policy query/update, Nakama runtime and live MCP calls are NOT_RUN. Both Jazebe source trees remain Git-clean.

## Operator medal library checkpoint (2026-10-22)

`go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`, targeted `go vet`, root `go build .`, MCP module `go test ./... -count=1` and `go vet ./...`, and `git diff --check` passed after adding medal library code and schema. Pure tests cover source condition schemas, brand parameter cleaning, event mapping, key/name/theme validation and operator privilege separation. The migration parser test passes with 24 Jazebe files. PostgreSQL schema execution, SQL queries, live Nakama plugin/RPC behavior, medal awards, backfill and cross-node concurrency are **NOT_RUN**. Both Jazebe source Git trees were clean. No Nakama core Go file was changed in this batch.

Root `go test ./...` was also run and **failed** outside the new integration packages: `internal/gopher-lua` reports nonconstant format strings under Go vet, `console/TestLocalizedIndex` reports stale gzip assets, and server tests cannot connect to PostgreSQL at `127.0.0.1:5432`. Focused Jazebe and migration packages passed within that run. These root failures match the pre-existing baseline; no core files were changed in this batch.

## Progression outbox checkpoint (2026-10-22)

The migration parser test, focused native/reconcile/migrate tests and vet, MCP module tests and vet, root build and diff check passed after adding `jazebe_progression_event` and the opt-in retry worker. Pure worker configuration rejects invalid values. The event insert, row lock, XP ledger/stat transaction, worker retries and notifications have **not** run against PostgreSQL; Nakama plugin loading is also unverified. Both Jazebe source Git trees remained clean. The counts are 130 RPCs, 23 migrations and 97 MCP tools; no controller group has verified parity.

## Purchase and group-join XP checkpoint (2026-10-22)

After connecting source purchase and group-join XP activity to the native XP ledger, `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`, targeted `go vet`, MCP module `go test ./... -count=1` and `go vet ./...`, root `go build .`, and `git diff --check` passed. Source backend and frontend Git trees remained clean. This verifies compilation and existing pure tests only; XP SQL, transaction replay/concurrency, notification persistence, live Nakama hooks, medals and process-failure recovery have not been exercised. No controller group has verified full parity.

## Purchase throttle checkpoint (2026-10-22)

Pure `shopRateConfig` default/override/invalid values and context single-count tests pass in targeted native `go test`; migration embed/parser test passes with 22 Jazebe files. The throttle SQL has not been executed on PostgreSQL; cross-node concurrency, exact source HTTP rate-limit envelope and Retry-After remain unverified. Do not infer live rate enforcement from Go build or pure tests.

## Shop brand/operator controller batch (2026-10-22)

With 130 registered native RPCs, 21 migrations and 97 MCP tools: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`, targeted `go vet`, MCP module `go test ./... -count=1` and `go vet ./...`, root `go build .`, and `git diff --check` pass. Pure tests cover brand pagination, product status, date filters, nullable product edits and operator reason bounds. MCP SDK discovery passes with mocked Nakama upstream. The generated root binary was removed. Both Jazebe source repositories are Git-clean. The new SQL, concurrent purchases, coupon claims, operator notices, live RPCs, response parity, purchase rate limiting and progression events remain unverified or incomplete. No full controller group is marked verified.

## Instant-delivery store batch (2026-10-22)

After coupon inventory, code generation, coupon purchase, digital create/purchase, three-type catalog and generic purchase routing: targeted native/reconcile/migrate `go test -count=1` and `go vet`, MCP module `go test ./... -count=1` and `go vet ./...`, root `go build .`, and `git diff --check` passed. Pure tests cover code format/random alphabet, coupon expiration omitted/null distinction, promoted JSON input fields, digital payload validation, purchase archive record shapes and catalog filters. Migration embedding/parsing passes for 21 Jazebe SQL files. The generated binary was removed. PostgreSQL migrations and transactions, concurrent coupon claims, rate limits, progression events, live Nakama plugin loading and remote MCP OAuth have not been verified.

## Shop client read batch (2026-10-22)

After purchase archive, ticket catalog/detail, source-format purchase records and post-commit response fix: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`, targeted `go vet`, root `go build .`, MCP module `go test ./... -count=1` and `go vet ./...`, and `git diff --check` all passed. MCP SDK discovery now finds 89 tools. Pure tests cover archive filters, three delivery-record shapes, Persian shop countdown/relative time, and catalog input validation. The generated root binary was removed. SQL queries, live PostgreSQL order/list behavior, concurrent purchase replay, live Nakama RPC loading and remote OAuth remain unverified.

## Latest lottery draw and client list batch (2026-10-22)

After five operator lottery routes and the internal KYC share-release helper: targeted native/reconcile/migrate `go test` and `go vet`, root `go build .`, and MCP tests/vet passed. Operator routes use the existing allowlist. The new SQL, transaction ordering, KYC release, audit report and operator response contracts have not been exercised with PostgreSQL or live Nakama. One operator controller group is mapped in code but has **no verified full parity**.

After finished raffle cards, client detail and Persian date helpers: native/reconcile/migrate tests/vet, MCP SDK tests/vet (86 tools and detail invocation), root build and diff check passed. Persian date tests compare selected instants to source Node Intl output. The new SQL and resulting response models remain unverified against PostgreSQL and a live Nakama process. Both source Git trees were clean.

After brand winner reads/delivery, source-shaped brand list/detail, metadata and opt-in draw worker: targeted native/reconcile/migrate tests/vet, MCP tests/vet, root `go build .` and `git diff --check` passed. The worker is configuration-gated; its setting validation and source delivery transition rules have pure tests. Twenty migrations parse but have not executed against PostgreSQL; neither scheduler behavior nor winner delivery was live-tested. MCP has 84 tools; discovery passed against the SDK mock upstream.

After manual draw/replay, winner/share/KYC migration and recipient-scoped notifications: targeted native/reconcile/migrate `go test` and `go vet`, MCP module `go test ./...` and `go vet ./...`, root `go build .`, and `git diff --check` passed. Pure tests include a Node HMAC fixture and payout-rate validation. Twenty migrations parse through the embed test; none has executed on PostgreSQL. Manual draw, KYC gating, prize payout, replay and scheduled processing are not live verified. Generated `nakama.exe` was removed.

Native/migrate Go tests and vet passed after the source-compatible deterministic draw engine and active raffle list were added. Pure tests cover winner determinism, no replacement, prize priority, unawarded prizes, exact rial split, hour countdown and Persian prize labels. MCP Go tests and vet passed with 82 tools; an official SDK client invoked the new array-returning raffle tool against a mocked Nakama upstream. Root build and source integrity checks must be repeated after the latest code changes. No PostgreSQL draw, cancellation/refund, or live Nakama test ran.

## Lottery dependency batch (2026-10-22)

Shop ticket follow-up: targeted native/reconcile/migrate Go tests and vet plus MCP tests and vet, root `go build .` and `git diff --check` pass after the eighteenth migration, two shop RPCs and two MCP tools. Pure tests cover ticket product bounds and deterministic purchase-key behavior. Both Jazebe Git trees remained clean; generated `nakama.exe` was removed. No PostgreSQL order/debit/stock/ticket transaction or live Nakama call has been tested.

Group mission follow-up: targeted native/reconcile/migrate `go test` and `go vet` plus MCP module `go test ./...` and `go vet ./...`, root `go build .` and `git diff --check` passed after the seventeenth migration and media/code/content group payout changes. Generated `nakama.exe` was removed; both source Git trees stayed clean. Pure tests check group rewards, required lottery, target count and prohibition of individual points. PostgreSQL progress locking, ticket caps, notification writes, contributor XP and real Nakama RPCs remain unverified.

Targeted native/reconcile/migrate `go test` and vet plus MCP module `go test ./...` and vet passed after six lottery RPCs, one migration, and six MCP tools. Tests cover draft time/prize/cap validation, update clear fields, ticket serial and source allowlist, MCP discovery (78 tools) and mocked lottery read invocation. Root `go build .` and `git diff --check` passed before the update RPC; final repeat is pending. Both Jazebe source Git trees were clean. These checks do not execute SQL or prove concurrent ticket issuance.

## Latest group management batch (2026-10-22)

Subsequent group chat batch: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`, targeted vet, root `go build .`, MCP `go test ./...` and `go vet ./...` passed after five client chat RPCs, staff moderation, two stream RPCs and the fifteenth migration were added. Pure tests cover chat text length, self-delete and manager-delete permissions, and required realtime session. No live PostgreSQL, Nakama websocket or standard external client was available, so stream delivery and database behavior are not verified.

Further chat batch: absent-member Nakama notification refresh, brand/operator report queue and dismissal, source page/quote bounds were added. Targeted native/migrate `go test` and `go vet`, root `go build .` passed. No live database or websocket verification was performed.

- `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate` and targeted `go vet`: passed after group settings, roles, leadership, removal, disband and group reads. Pure tests cover nullable group fields, remove permission rules, member ordering and feed kind mapping.
- `go build .` at Nakama-Farsi root: passed. This is a compile check, not a live plugin or database check.
- In `integrations/mcp`, `go test ./...` and `go vet ./...`: passed with 67 tool registrations. Official SDK interoperability tests still use a mock Nakama upstream.
- Fourteen Jazebe migration files parse through the repository's migration test. The two latest migrations add leadership offer and disband actor persistence. No migration executed against PostgreSQL.
- Live group transactions, concurrent creator transfer, Nakama plugin loading, realtime delivery, mission/group payout and OAuth remote MCP authorization: not verified or not implemented as specified in the parity matrix.

Latest group/upload/unlimited/catalog batch (2026-10-22): native/reconcile/migrate `go test` and `go vet`, MCP module `go test ./...` and `go vet ./...`, root `go build .`, and `git diff --check` passed. New tests cover group field/session/reference/single and bulk decision guards, invite token/session/invalid preview guards and Persian formatting, media limits and sniffing, malformed chunks, anonymous withdrawal, image/video evidence draft validation, unlimited key behavior, and catalog fallback/filter guards. Twelve migrations parse, but none has executed against PostgreSQL in this environment. Group creation/join/request/manager-decision/invite/leave transactions, upload HTTP transfer, range responses, expiry sweep, evidence ownership SQL, and catalog SQL await live integration testing. Both Jazebe source Git trees were clean; generated root binary was removed. The full root `go test ./...` still fails on pre-existing gopher-lua vet format errors, mismatched console compressed assets, and server tests requiring PostgreSQL at `127.0.0.1:5432`.

Latest mission visibility/moderation/periodic mission batch (2026-10-22): `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`, MCP module `go test ./...`, targeted `go vet`, root `go build .`, and `git diff --check` passed after periodic media/code/content changes. New pure tests cover Tehran display windows, level/prerequisite denial, Tehran midnight period keys, operator reason validation and deny-by-default access. These do not establish live SQL behavior. Nine migrations remain unexecuted on PostgreSQL; Docker Desktop daemon failed to start because WSL2 reported virtualization disabled. Both Jazebe source Git trees were clean at the last check. The generated root binary was removed.

| Check | Result | Scope and limitation |
|---|---|---|
| `go test ./integrations/jazebe/native ./migrate` | Passed, 2026-10-22 | Compiles native module and parses embedded migrations; no live database |
| `go build .` from target root | Passed, 2026-10-22 after membership batch | Nakama core builds; generated local binary removed |
| `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate` | Passed, 2026-10-22 | Reconcile package compiles; no live DB test |
| `go build ./integrations/jazebe/reconcile` | Passed, 2026-10-22 | Generated local executable removed |
| Planet draft and operator targeted tests/vet | Passed, 2026-10-22 | Validates source name/slug rules and deny-by-default operator check; no database |
| Seven migration parser check | Passed, 2026-10-22 | Lifecycle, mission, audit and brand team SQL parsed and embedded; PostgreSQL execution pending |
| Planet publish requirements unit test | Passed, 2026-10-22 | Checks missing requirements and ready case; no live mission/action integration |
| Media mission draft validation and operator denial | Passed, 2026-10-22 | Pure validation and deny-by-default behavior; no DB execution |
| MCP 15-tool discovery | Passed, 2026-10-22 | Official SDK client against mock Nakama upstream; no live RPC calls |
| Brand team role mapping | Passed, 2026-10-22 | Pure role action test; no live database |
| Mission/reward native tests, migration parser, vet and server build | Passed, 2026-10-22 | 27 RPCs compile; media guards, review validation and level threshold tests pass; SQL transactions not executed |
| MCP 23-tool SDK discovery and media start invocation | Passed, 2026-10-22 | In-memory and Streamable HTTP against mock Nakama upstream; no live mission RPC |
| Mission payout, duplicate prevention, XP and planet approval against PostgreSQL | Not run | PostgreSQL and Docker tools unavailable in this environment |
| Code-entry unit rules, native build/vet, ninth migration parse and MCP 27-tool discovery | Passed, 2026-10-22 | Crypto code format and code draft guards checked; no live redemption or code stock concurrency test |
| Wallet read and adjustment build/vet, shared ledger validation, MCP 32-tool discovery | Passed, 2026-10-22 | Adjustment deny-by-default and scoped idempotency key tests pass; no live database or notification delivery |
| Text/code evidence validation, transactional notice compile, reconciliation build, MCP 38-tool discovery | Passed, 2026-10-22 | No live notification or database execution; image/video evidence unavailable |
| `go test ./...` from `integrations/mcp` | Passed, 2026-10-22 | Official SDK client verifies discovery/invocation over in-memory and Streamable HTTP transports; mock Nakama upstream |
| MCP dependency setup | Resolved | Fetched blocked `segmentio/encoding` directly, then normal `go mod tidy` succeeded |
| PostgreSQL migration execution and membership concurrency | Not run | No database service available |
| Docker plugin load | Not run | Docker unavailable |
| MCP client interoperability against live Nakama | Not run | Live Nakama and database unavailable; SDK mock-upstream tests pass |

The native test checks identity and JSON validation. It does not prove SQL behavior. Membership join, fan counter and ordering require database tests before a parity claim.
# Latest operator game batch (2026-10-22)

PASS after mission invite join changes: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`; `go vet ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; root `go build -o $env:TEMP\\nakama-jazebe-build.exe .`; nested MCP `go test ./... -count=1` and `go vet ./...`; `git diff --check`. PostgreSQL mission join/attribution transaction and concurrent retry behavior NOT_RUN. No live Nakama instance was launched. This validates compilation and focused pure behavior only.

PASS: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; `go vet ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; `go build -o $env:TEMP\\nakama-jazebe-build.exe .`. New pure operator tests cover catalog validation and pagination. Source Git trees are clean. Registrations: 176 native RPCs, 29 Jazebe migrations, 110 MCP tools. NOT_RUN: live PostgreSQL migrations and moderation transactions, live Nakama RPCs, full root regression suite. These results do not verify operator game route parity.

# Latest profile batch (2026-10-22)

PASS: `go test ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate -count=1`; `go vet ./integrations/jazebe/native ./integrations/jazebe/reconcile ./migrate`; root `go build -o $env:TEMP\\nakama-jazebe-build.exe .`; nested MCP `go test ./... -count=1` and `go vet ./...`; `git diff --check`. The embed/parser test includes the new 30th Jazebe migration. Pure tests cover profile name/avatar bounds, Persian month rollover, XP level projection and mission signup-context guard. MCP SDK tests discover 112 tools and invoke private/public profile reads against a mock Nakama upstream. `TestProfilePostgresConstraints` is written and **SKIPPED** locally because `JAZEBE_DISPOSABLE_TEST_DB_URL` is unset. `.github/workflows/jazebe-migrations.yaml` defines a disposable PostgreSQL 16.8 migration/constraint job; the CI job has **NOT_RUN**. Live Nakama module/RPC loading, real MCP-to-Nakama and the full root regression suite also remain NOT_RUN. No database-backed integrity result is claimed.

