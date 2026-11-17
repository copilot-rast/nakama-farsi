# Cluster readiness (phase 0b)

Goal: every plugin of this fork (`integrations/economy`, `integrations/publication`, `integrations/base`, `samples/stargame`) and the shared packages (`internal/localized`, `internal/secretbox`) behave correctly when several Nakama nodes run against the same database, so the cluster phase (phase 6, porting the cluster layer of doublemo/nakama-plus) needs no rework in plugin code. The database engine and its settings are unchanged.

Audit date: 2026-11-10, branch `dev`.

## Rules from now on

Every phase follows these rules; review new code against them.

1. No state that must be shared lives in process memory. Package-level variables hold only constants, compiled regexps and immutable tables. No caches today; if one is added it gets a short TTL, and admin-visible config (languages, brand, payment methods, providers) is read from the database or checked against `updated_at`.
2. Background jobs run on every node and must be safe there: claim rows with `FOR UPDATE SKIP LOCKED` in one statement (with a lease column such as `inquired_at` / `next_attempt_at`), re-check ownership before any external side effect, and keep every money/grant/delivery transition guarded by its current status. For a job that must run on one node at a time, use the lease in `internal/clusterkit` (`AcquireLease` / `ReleaseLease`, table `cluster_lease`).
3. Rate limits live in the database. New limits use `clusterkit.Allow` (fixed-window counter, one atomic upsert on `cluster_rate_limit`); phase 2 OTP limits use it. Existing limits are already database-backed (see the table).
4. Uploaded files go through `internal/objectstore` (`STORAGE_DRIVER=local|s3`). Never write user content to the node's disk directly.
5. Config comes only from env (`--runtime.env` or the process environment) and is identical on every node.
6. Persist first, push second. Since phase 6 (`docs/cluster-design.md`, clustering on) presence/stream lists, `StreamSend`, `StreamUserJoin/Kick`, `SessionDisconnect`, `SessionLogout`, bans, the live part of `NotificationsSend`, match listing, `MatchSignal`, parties and the matchmaker work across nodes; the push is still a push (a player who is offline gets it from the database).
7. A plugin that needs to wake waiters on other nodes uses `internal/cluster` `Signal(topic)` / `WaitSignal(ctx, topic, max)` (Redis pub/sub with clustering on, a plain sleep of `max` with it off), never its own in-memory channel.

## Findings

Status: **fixed** = changed in phase 0b; **safe** = already correct on N nodes, no change; **phase 6** = node-local behaviour inside Nakama itself, fixed by the cluster layer, plugin code needs no change.

| Plugin | File:line | Single-node assumption | Risk | Fix | Status |
|---|---|---|---|---|---|
| economy | `integrations/economy/uploads.go:46`, `:86` (was `os.WriteFile`/`os.Open` on `ECONOMY_UPLOAD_DIR`) | Console uploads were written to the node's disk and served from it | An image uploaded through node A is 404 on node B; lost when a node is replaced | `internal/objectstore`: `local` (unchanged default) or `s3` (S3/MinIO, SigV4 with the standard library, no new deps). URLs stay `/v2/economy/files/{name}`; with `s3` Nakama proxies the object with the same safety headers (CSP sandbox, nosniff, Range via `ServeContent`), or answers `302` to `STORAGE_S3_PUBLIC_BASE_URL` when that is set | fixed |
| economy | `integrations/economy/module.go:115` (`dataDirFromArgs`) | Default upload dir comes from the node's `--data_dir` | Different per node | Only used by the `local` driver; `s3` ignores it | fixed (by driver) |
| economy | `integrations/economy/module.go:71`, `payments.go:388` | Reconciler on every node selected the same 20 pending payments, then bumped them one by one | Two nodes inquire/verify the same payment in the same minute: duplicate gateway calls, `inquiry_count` counted twice. Settlement itself was already guarded (`status='pending'` transitions, idempotent grant) | `claimForReconcile` (`payments.go:406`): one `UPDATE … WHERE id = (SELECT … FOR UPDATE SKIP LOCKED) RETURNING`, one payment per claim, up to 20 per pass; the console «inquire» keeps its own bump | fixed |
| economy | `integrations/economy/client.go` gift redeem / purchase | Gift limits, wallet, prize claims | — | Row locks (`FOR UPDATE`), unique `once_key`, `source_type/source_id` dedupe | safe |
| economy | `integrations/economy/gateways.go:17` | Package vars for gateway base URLs | — | Constants overridden only by tests | safe |
| publication | `integrations/publication/module.go:58`, `worker.go:107` | Delivery worker on every node | — | Claims already used `FOR UPDATE SKIP LOCKED` with a 2-minute lease in `next_attempt_at`; marks are guarded by `attempts` | safe |
| publication | `integrations/publication/worker.go` (send loop) | A batch of 50 live SMS/email sends with a 45 s timeout each can run past the 2-minute lease | Another node re-claims the rest of the batch after the lease and both nodes send the same SMS/email | `renewClaim` (`worker.go:165`) before every external send: extends the lease only while `attempts` is unchanged; if another node took the row, the send is skipped | fixed |
| publication | `integrations/publication/worker.go` (`sendInApp`) → `NotificationsSend` | Live push goes only to sockets on the sending node | Users connected to another node see the notification only on the next inbox load | Inbox row is persistent (stored for every user); live routing is server-side | phase 6 |
| publication | `integrations/publication/client.go:357` | Comment rate limit (5/min/user) | — | Already database-backed: count of the user's comments in the last minute under a per-user transaction lock | safe |
| base | `integrations/base/setup.go:95` | Setup wizard claim-once | — | Single conditional `UPDATE … WHERE status='pending' RETURNING` | safe |
| base | `integrations/base/brand.go:22`, languages, categories, tags | Brand/languages config | — | Read from the database on every request; the public brand endpoint only sets HTTP `Cache-Control: max-age=60` (client/CDN side, same on all nodes) | safe |
| internal/localized | `internal/localized/localized.go:70` | Language set | — | `Load` reads `base_language` per call; no cache | safe |
| internal/secretbox | `internal/secretbox/secretbox.go:27` | Key derivation | — | Pure functions; `ECONOMY_SECRET_KEY` must be identical on every node (config rule 5) | safe |
| publication | `integrations/publication/otp.go` (phase 2) | OTP rate limits, code tries, sweeping | — | Limits through `clusterkit.Allow`; verify under `pg_advisory_xact_lock` on the number with the try count committed in the same transaction; old codes swept under the `publication-otp-sweep` lease | safe |
| base | `integrations/base/rules.go`, `versions.go`, `countries.go` (phase 2) | Account rules, versions, countries | — | Read from the database on every hook/RPC; no cache | safe |
| base | `integrations/base/players.go` (phase 3) restrictions | Suspension/mute end times, auto-lift, live kick | — | Enforced by `ends_at > now()` on every check (no timer); `expired` status set by one `UPDATE` under the `base-restriction-expire` lease; every node's `restrictionLoop` logs out and disconnects its own sockets of players suspended in the last two minutes, so the kick reaches every node without shared state | safe |
| base | `integrations/base/chathooks.go` (phase 3) chat filter and reports | Word list, mute check, report rate limit | — | Word list and mute read from the database for every message (no cache); one report per player and message by a partial unique index; report rate limit through `clusterkit.Allow`; report resolution locks the report row | safe |
| economy | `integrations/economy/events*.go` (phase 4) events, rewards, streaks | Rate limits, grants, workers | — | Client limits through `clusterkit.Allow`; progress, streak and level rows locked per player in the recording transaction; claims unique per scope and granted through the prize path with `source_id` = claim id (`TestEventGrantsAreIdempotentOnManyNodes`: 4 services, 40 events, 8 claims paid once); left-over auto claims picked with `FOR UPDATE SKIP LOCKED` and an `attempted_at` lease; log sweep under the `economy-event-sweep` lease; settings read from the database per event | safe |
| internal/playerevents | `internal/playerevents/playerevents.go` (phase 4) | Package-level recorder set by economy | — | Written once while plugins initialise, read-only afterwards, same on every node | safe |
| base | `integrations/base/support*.go`, `livechat.go` (phase 5) support tickets and live chat | Ticket numbers, take/transfer races, uploads, typing, online agents, auto-close | — | Ticket numbers unique with a retry; take is one conditional `UPDATE … WHERE status='waiting'`, transfer and messages run under the chat row lock, sequence numbers come from the chat row; one open chat per player by a partial unique index; uploads are chunks in the database, then the object store (never the node disk); attachment URLs are HMAC-signed with a key from `ECONOMY_SECRET_KEY` (the same on every node); typing, online and agent heartbeats are timestamps in the database; console agents long-poll the database (no in-memory subscribers); rate limits through `clusterkit.Allow`; auto-close, idle close, file and activity sweeps under the `base-support-autoclose` lease (`TestLiveChatTakeAndTransferRaces`: 6 agents race for one chat, 2 transfers race, 8 concurrent opens make one chat) | safe |
| base | `integrations/base/remoteconfig.go`, `activity.go` (phase 5) | Config resolution, rollout, activity writes | — | Keys read from the database on every call (no cache); buckets are a hash of key and player id, identical on every node; history and rollback in one transaction with a row lock; activity rows are plain inserts | safe |
| economy | `integrations/economy/referral*.go` (phase 5) | Double entry, double reward | — | Entry under an advisory lock on the player plus a unique referee and unique device/phone marks; completion is a locked conditional update; rewards are claims (`kind='referral'`) paid through the prize path (`TestReferralGoalAndRewardsAreIdempotentUnderConcurrency`: 8 entries on 4 services → 1, 20 goal events on 4 services → each side paid once, re-granting each claim from 3 services pays nothing more) | safe |
| internal/inapp, internal/activity | phase 5 | Package-level sender set by publication; activity writer | — | Sender written once while plugins initialise; activity is a stateless insert | safe |
| internal/hookchain | `internal/hookchain/hookchain.go` (phase 3) | Package-level registry of hook chains | — | Written only while plugins initialise, read-only afterwards; identical on every node because every node loads the same plugins | safe |
| stargame | `samples/stargame/seed.go:173` | Seed on every node start | Concurrent starts | One transaction, every insert `ON CONFLICT DO NOTHING` in a fixed order; covered by the existing 4-way concurrent `TestSeedPostgres` | safe |
| stargame | `samples/stargame/main.go:79`, `:96` | Rounds | — | Stored in Nakama storage with version checks (`ErrStorageRejectedVersion`) | safe |
| stargame | `samples/stargame/web.go` | Embedded static files | — | Read-only, built into the plugin | safe |

No plugin uses match listing, `MatchSignal`, the matchmaker, parties or `sync.Once`; there are no timers besides the worker tickers listed above and no in-memory caches, counters or maps that hold shared state.

## Phase 6 result (2026-11-11)

Every row marked **phase 6** above is now handled by the cluster layer when `cluster.enabled` is on; the plugin code needed no change except the live support chat signal:

| Item | Delivered by |
|---|---|
| Tracker/presence and streams across nodes (`StreamUserList/Send/Kick/Join`) | Redis stream hashes + per-node views (`server/cluster_tracker.go`), remote stream joins through the session's node (`ClusterStreamManager`) |
| Live `NotificationsSend` (publication in-app, support 1210/1211) and `SessionDisconnect` | `ClusterRouter` routes by the node in each presence ID; disconnects go to the node found from the session ID |
| `SessionLogout`, bans and phase 3 suspension | invalidation log `nk:inv` replayed and tailed by every node; ban disconnects routed (`server/cluster_session.go`) |
| Matches, matchmaker, parties, status | host routing (`cluster_match.go`, `cluster_party.go`), sharded exactly-once matchmaker (`cluster_matchmaker.go`), status followers in Redis (`cluster_status.go`) |
| Live support chat "chat changed" | `integrations/base/livechat.go` publishes `chat:<id>` after every message, typing, take, transfer, close and player view; the console long-poll waits up to 5 s on it (600 ms sleep when clustering is off, as before) |
| Leaderboard/tournament schedulers | each callback claimed once with `clusterkit.AcquireLease` (`cluster_scheduler.go`) |

Multi-node checks: `e2e/cluster/` (four nodes) and `go test ./internal/cluster ./server -run Cluster` with `CLUSTER_TEST_REDIS`. Agent presence for support could still move from the `base_support_agent` heartbeat row to a Redis TTL key; it is correct as it is and was left in the database.

## What phase 6 had to deliver at the server layer

These are inside Nakama OSS, not in our plugins. When porting the nakama-plus cluster layer, check each one with a multi-node e2e:

- Tracker/presence and streams across nodes: `StreamUserList`, `StreamSend`, `StreamUserKick`, `StreamUserJoin`.
- Message routing to sessions on other nodes: live part of `NotificationsSend` (publication in-app, any future notification), `SessionDisconnect`.
- Session cache invalidation broadcast: `SessionLogout`, bans, suspension (OSS `LocalSessionCache` keeps invalidations per node, so a token logged out on node A stays valid on node B). Phase 3 temporary suspensions already repeat the logout on every node for two minutes (`restrictionLoop`), and sign-in, refresh and chat are refused from the database, so a cluster broadcast only shortens that window.
- Matches, matchmaker, parties, status, leaderboard rank cache (not used by our plugins today; Star Catcher rounds are RPC + storage).
- Live support chat (phase 5): correctness never depends on the node, but two things are only fast on one node today. (1) The live notices to the player (codes 1210/1211, `NotificationsSend`) reach only sockets on the node that handled the agent's request; on another node the game still sees the change through its 3-second poll. Phase 6 must route these notifications to the node holding the player's socket. (2) Console long-polls re-read the chat row every 600 ms on whichever node holds the request; with many agents this is many small reads. Phase 6 should add a "chat changed" signal on the shared Redis (publish on every message, typing, take, transfer, close; the long-poll waits on it with the database read as fallback) and could give agent presence a Redis TTL key instead of the `base_support_agent` heartbeat row. Nothing in the chat state needs to move: it stays in Postgres.

## Shared helpers added

| Package | What | Tables |
|---|---|---|
| `internal/objectstore` | `Store` interface (`Put/Get/Delete/PublicURL`), `local` and `s3` drivers, `ConfigFromEnv`, `SignV4` | — |
| `internal/clusterkit` | `AcquireLease/ReleaseLease` (one holder per name until expiry; holder = node name + process id from `Holder(ctx)`), `Allow` (fixed-window rate limit, one atomic upsert), `SweepRateLimits` | `cluster_lease`, `cluster_rate_limit` (migration `20261110010000-cluster-kit.sql`) |

## Storage env

| Key | Meaning | Default |
|---|---|---|
| `STORAGE_DRIVER` | `local` or `s3` | `local` |
| `STORAGE_S3_ENDPOINT` | `http(s)://host[:port]`, no path | required for `s3` |
| `STORAGE_S3_BUCKET` | Existing bucket | required for `s3` |
| `STORAGE_S3_ACCESS_KEY`, `STORAGE_S3_SECRET_KEY` | Credentials | required for `s3` |
| `STORAGE_S3_REGION` | Signing region | `us-east-1` |
| `STORAGE_S3_PATH_STYLE` | `true` = `endpoint/bucket/key` (MinIO), `false` = virtual host | `true` |
| `STORAGE_S3_PREFIX` | Optional folder; economy keys are `<prefix>/economy/<name>` | empty |
| `STORAGE_S3_PUBLIC_BASE_URL` | Empty = Nakama proxies; set = `302` to `<base>/<key>` | empty |

Serving choice: proxy by default, because the bucket would not send the `Content-Security-Policy: sandbox` and `nosniff` headers that make uploaded SVGs safe, and because the file URL stays on the API host. A CDN/public bucket can be used by setting `STORAGE_S3_PUBLIC_BASE_URL`, after configuring those headers there.

## Verification

- Unit: SigV4 matches the three AWS S3 documentation examples (GET object with Range, PUT object, GET bucket with query); config parsing; key validation; local driver round trip.
- Real S3: `TestS3StoreAgainstRealServer` and `TestEconomyUploadsOnS3` against a throwaway MinIO (`OBJECTSTORE_TEST_S3_*`): put/get/delete, wrong secret → `SignatureDoesNotMatch`, upload through one service and read through another (proxy, Range 206, 404, 302 with a public base).
- Many nodes against a disposable database: `TestEconomyReconcilerOnManyNodes` (4 services × 2 passes over 12 pending payments: each payment claimed, inquired and verified exactly once, granted once; the same test fails on the old reconciler with 44 checks for 12 payments), `TestExpiredClaimIsNotSentTwice` (live SMS slower than the lease: every phone gets exactly one message; fails without `renewClaim`), the existing `TestConcurrentWorkersDoNotDoubleSend` (6 workers, 130 recipients), `TestLeaseOneHolderAtATime` (16 contenders), `TestRateLimitIsSharedAndAtomic` (40 concurrent hits, limit 10 → exactly 10), `TestSeedPostgres` (4 concurrent seeds).
- Disposable test databases: `ECONOMY_DISPOSABLE_TEST_DB_URL`, `PUBLICATION_DISPOSABLE_TEST_DB_URL`, `BASE_DISPOSABLE_TEST_DB_URL`, `STARGAME_DISPOSABLE_TEST_DB_URL`, `CLUSTERKIT_DISPOSABLE_TEST_DB_URL` (names must end in `_disposable`; run `nakama migrate up` on them after adding a migration).
- Two real nodes on this Mac (2026-11-10): `nakama-mac` (7450/7451) and `nakama-mac-2` (7460/7461), same binary, modules, database and runtime env, both on `STORAGE_DRIVER=s3` against a throwaway MinIO. 25 pending payments (10 created for the check, aged 2 h) were split by the two reconcilers in the same second, 11 and 14, every one claimed once (`inquiry_count` 1, all 10 test payments `expired`). A 20-recipient sandbox SMS notification created on node 2: 20 sends logged in total, every recipient `sent` with `attempts` 1. An image uploaded through each node's console was served byte-identical, with `image/png` and the sandbox CSP, by the other node's API port; nothing was written to the node's disk. Test rows were removed afterwards.
- Full Playwright suite (`cd e2e && npm test`): 58 passed, twice, on the local driver.
