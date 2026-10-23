# Remaining Jazebe migration work

## Route parity progress (2026-10-24)

`docs/integration/jazebe-route-parity.json` (validated by `go run ./scripts/jazebe_route_audit`): 131 IMPLEMENTED, 123 PARTIAL, 83 NOT_STARTED, 10 BLOCKED, 8 EQUIVALENT_NATIVE, 34 SOURCE_ONLY_NOT_APPLICABLE of 389 routes. IMPLEMENTED means a PostgreSQL workflow test exercises the RPC.

Added in this batch: brand progression XP panel and level ladder; operator/me, operator directory and audit-log explorer; operator and brand-panel group views; group suspend/unsuspend/disband; staff member removal; planet member suspension. Migrations `20261024050000-jazebe-group-suspension` and `20261024060000-jazebe-planet-member-suspension` add nullable columns only.

Deliberate adaptations:
- Operator levels come from the `JAZEBE_*_OPERATOR_USER_IDS` deployment allowlists. `operators/:id/level` (route 286) is NOT_STARTED by design, because a database-backed registry would replace every operator check and needs an explicit decision.
- Brand-panel staff RPCs take `planetId` (or `brandId` for lists) in place of the source `X-Product-ID` header. A group outside the addressed planet reads as not found.
- Staff group and member audit rows carry `planet_id`. Like the existing group audit rows, they do not set `brand_id`.

Next: wallet admin (91, 92), analytics (56–62), broadcasts (248–262), integrations/webhooks/public API (183–200), operations (275–283), audit exports (288–291), operator planet list/get (309, 310), planet domains (292, 316–319) and localization (1–20). After that, workflow tests for the PARTIAL routes.

## Native PostgreSQL follow-ups (2026-10-23)

PostgreSQL integration is now runnable locally ([../ANDRONIX-POSTGRESQL.md](../ANDRONIX-POSTGRESQL.md)), and the 7 gated tests plus `TestJazebeSQLPreparesPostgres` pass. Next:
1. DONE 2026-10-23: native Nakama + plugin run on `nakama_dev`; brand and planet create verified live; RPC error codes fixed. Next: live workflow tests for the other controller groups, starting with the 43 fixed audit/activity paths.
2. Decide the Jazebe `migrate down` policy: make Up sections re-runnable, write real Downs, or document it as unsupported (current state).
3. The prepare test proves type-correctness only. The 43 fixed audit/activity paths (groups, missions, lottery, wallet, planet, brand) still need behavioral PostgreSQL workflow tests.

## Current medal dependency gap (2026-10-22)

Medal purchase, group join, approved solo mission, point/ticket acquisition and opt-in durable backfill now have native code paths, but the domain is still **partial**. Next: migrate actual game-session and qualified-referral histories and events, trace source special-event ingress, integrate the new medal wall into the complete account-profile contract with translations, complete source brand progression XP/levels operations, then run PostgreSQL duplicate-award/concurrency and live Nakama tests. Remote MCP OAuth, source frontend workflows, and the other incomplete controller groups remain. `JAZEBE_MEDAL_BACKFILL_ENABLED` defaults off; enable only after applying migration `20261023130000` and reviewing queued jobs. Earlier sections below record historical gaps.

## Policy now present; medal awards next (2026-10-22)

Jazebe's singleton platform policy and its three GET/PATCH entry points now have native code and the 25th migration. Medals still lack planet selection, event evaluation, unique acquisitions, points/tickets payout, notices and backfill. Wire these before enabling brand medal writes. Counts are 141 RPCs, 25 migrations and 98 MCP tools; zero controller groups have runtime-verified full parity. Live PostgreSQL/Nakama, KYC, webhooks, other backend groups, frontend workflows and remote MCP OAuth remain.

## Medal dependency next (2026-10-22)

Operator medal-library CRUD/read/history and condition helpers now have native code and a 24th schema migration, but planet selection, acquisition, event evaluation, payouts, backfill and client reads do not. Implement those before enabling brand medal creation. Run the new SQL against disposable PostgreSQL and exercise operator authorization, concurrent version updates and unique acquisitions when available. Counts: 138 RPCs, 24 migrations, 97 MCP tools; zero of 51 groups verified. KYC, webhooks, other backend groups, frontend workflows and remote MCP OAuth remain.

## Durable event checkpoint (2026-10-22)

The 23rd migration and native progression worker now persist and retry purchase/group-join XP events. Enable `JAZEBE_PROGRESSION_WORKER_ENABLED=true` only after applying all Jazebe migrations; otherwise immediate processing can log failures and pending rows will wait. Source medals, creator group-size evaluations, backfill, point/ticket awards and live SQL tests remain. Counts: 130 RPCs, 23 migrations, 97 MCP tools, zero verified controller groups. Next implement the medal domain, then account/KYC and other remaining groups.

## Progression follow-through (2026-10-22)

Configured purchase and group-join XP now have native after-commit paths keyed by order or membership UUID, with a transactional XP ledger, stats and level-up notification. The source also evaluates active medals on these events, including group-size medals for the creator; medal definitions, backfill, acquisition, point/ticket payout and event delivery are missing. After-commit XP is not yet durable across process failure. Source shop, group and progression controller groups remain partial. Counts stay 130 native RPCs, 22 migrations, 97 MCP tools and zero verified complete groups. Implement durable event delivery and the medal domain next, then account/KYC, webhooks and the other source groups. PostgreSQL/live Nakama and remote MCP OAuth remain unavailable or incomplete.

## Purchase throttle follow-through (2026-10-22)

The shop purchase rate guard is implemented in code with the 22nd migration; PostgreSQL and exact source 429/Retry-After semantics remain unverified. Progression purchase events and medal evaluation are still absent. The store, brand shop and operator shop groups remain partial, despite route-name coverage. Current counts: 130 native RPCs, 22 migrations, 97 MCP tools; zero verified complete controller groups. Continue the source progression dependency, then other controller groups; run live shop concurrency tests when a safe database is available.

## Current checkpoint — shop controller follow-through (2026-10-22)

130 native RPCs, 21 migrations and 97 MCP tools are registered. The client store, 11-route brand shop and four-route operator shop have code mappings, but remain **partial** until source business rules and response shape are fully reconciled and PostgreSQL/live Nakama tests run. Specific shop gaps: purchase rate limiting, progression events, exact source notification and admin report behavior, coupon/digital concurrency, and SQL execution. No controller group has verified full parity. Next audit the new shop files and run a safe database suite when available; then continue account/KYC/reward release and webhook outbox dependencies, all other backend/controller groups, frontend workflows and remote MCP OAuth. Both Jazebe repositories remain read only.

## Current shop continuation (2026-10-22, after coupon and digital work)

Registered: 116 native RPCs, 21 migrations, 92 MCP tools. The client store has code paths for all three source instant-delivery types, but remains **partial**: source purchase rate limiting and progression events need reconciliation, and SQL/live runtime parity is unverified. Complete all 11 brand shop routes and 4 operator shop routes next: product lifecycle, coupon inventory/list, lottery options, sales and order reports, and moderation. Keep coupon/digital creation and raw code generation native-only until privilege-scoped remote MCP authorization exists. PostgreSQL concurrency and migration testing remain unavailable here. After the shop controller group, continue the next dependency-ready Jazebe domain; global KYC, webhooks, jobs, frontend and OAuth gaps remain.

## Current shop continuation (2026-10-22)

110 native RPCs, 20 Jazebe migrations and 89 MCP tools are registered. Zero of 51 controller groups have verified full parity. Client store now maps ticket list/detail, buyer purchase archive and ticket purchase, but coupon/digital product creation, coupon inventory and generation, both purchase deliveries, brand shop management, operator product moderation, sales/order reports and live SQL tests remain. Implement coupon schema/generation/purchase first, then digital product lifecycle and purchase, then brand/operator shop routes. The archive mapper for coupon/digital rows is present but those orders cannot yet be created. Concurrent stock, wallet and replay behavior requires PostgreSQL. Other global gaps from the previous checkpoint still apply, including KYC, webhooks, frontends and remote MCP OAuth.

## Latest checkpoint (2026-10-22)

Current verified registration counts are 107 native RPCs, 20 additive migrations and 86 MCP tools. No controller group has verified full parity. The five operator lottery routes are implemented in code, awaiting live integration verification. Lottery cancellation/adoption, active/finished raffle cards and detail, deterministic draw/share engine, brand manual draw, opt-in scheduled worker, operator replay/audit/report, KYC-gated winner/share persistence, internal share release, brand winner reads/delivery and full brand record projection have been added. Actual KYC verification still lacks a safe Nakama registered-phone mapping and provider flow; webhooks and other domains remain. PostgreSQL and live Nakama are unavailable. Historical counts in older notes below are superseded by this checkpoint.

Current mission dependency work: finish lottery cancellation/adoption, draw and payout; verify group progress, one-time ticket/XP payout and idempotent shop purchase against PostgreSQL. Lottery draft create/update/activate/list/get/pool and transaction-scoped ticket issuance are implemented. Shop ticket products and one-ticket purchase now call the shared issuer and point ledger, but coupon/digital products and most shop routes remain. Media, code and content mission drafts/actions support group targets and ticket rewards; invite/external/social types and progression events remain. Also complete invite-mission attribution, profile avatar/brand assets, external verification, webhook outbox, jobs, frontend, and live SQL/Nakama tests. Docker Desktop cannot start here because WSL2 reports virtualization disabled. Current counts: 94 native RPCs, eighteen migrations, 80 MCP tools, zero verified complete controller groups.

The complete source entry-point list is [jazebe-feature-inventory.md](jazebe-feature-inventory.md) and the 389 HTTP routes are [jazebe-http-routes.md](jazebe-http-routes.md). Each route still needs service, data, security, and test tracing. The groups below are a dependency order, not a reduced scope.

1. Complete brand and planet service rules: ownership/team, lifecycle, resolver, domains, settings, categories, discovery, membership paths, outbox, fan reconciliation deployment and source response contracts.
2. Map source account, profile, identity, roles and tenant permissions to Nakama users and sessions; then implement source-only behavior.
3. Migrate groups, chat, moderation and activity transactions.
4. Migrate wallet/economy, products, inventory and payment transaction semantics using Nakama wallet where safe.
5. Migrate progression, missions and rewards with atomic state transitions.
6. Migrate games, predictions, lottery and their scheduled jobs.
7. Migrate notifications, analytics, operations, media, public integrations, referrals and brand webhooks.
8. Adapt 12 Next.js modules and 5 admin UI modules to the Nakama-backed services.
9. Extend MCP with authorized capabilities as services become real; implement compliant remote OAuth before enabling non-loopback exposure.
10. Run DB, Docker, source parity, MCP client and Nakama regression suites; repair any failures.

Immediate next action: implement lottery cancellation with purchased-ticket refunds, earned-ticket transfer and later adoption, then lottery draw/payout. Complete remaining shop product types, reads and administrative routes. Do not retry Docker without a change in virtualization availability. Group mission and shop ticket purchase transactions require live PostgreSQL concurrency tests. The content evidence path checks uploaded image/video ownership, MIME and size, but needs live HTTP/PostgreSQL testing. Durable Nakama notices are written transactionally, but immediate realtime delivery and brand webhook side effects remain. Invite and external verification still need account/integration dependencies. Finish these before claiming controller parity. Brand team mutations, other planet endpoints, frontend workflows and OAuth for remote MCP remain.
# Current account and audit gaps (2026-10-23)

Operator account restriction is PARTIAL: transactional state/core-ban/audit/critical in-app notification and local session revocation are implemented, with conditional restoration and durable pending retry. The Jazebe RPC dispatch now checks authoritative account state on every call. Cross-node socket/token revocation, native Nakama endpoint denial, critical SMS, trusted self-deactivation OTP and source HTTP contract remain. Operator constellation projection exists but its SQL and group behavior are unverified. Run `TestAccountTransitionPostgres`, `TestAccountRevocationRetryPostgres` and `TestOperatorConstellationPostgres` on an explicitly disposable migrated PostgreSQL database. The source-linked route audit validates 389 route identities/handlers but only 28 have behavioral classifications (19 PARTIAL, 5 NOT_STARTED, 4 BLOCKED); 361 need source contract review. No group is fully verified. Next account code task: privacy-safe KYC review queue and provider-backed KYC/account lifecycle closure.

