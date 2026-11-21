# Game analytics (insights events)

Status: phase 1 in progress on branch `analytics` (not merged, not pushed). This file is the design; the "Progress" section at the end says what is built and tested.

## Goals

- One event standard for every game on nakama-farsi and its forks: a named event, a set of tags (tag group + id) and an optional numeric value, sent by the game client or by server code.
- Every event is stored once with user, session, time, platform, app version, game and source, and can be aggregated by any combination of tags, by user, by platform/version, by day/hour.
- The admin picks where events live: Nakama's own Postgres (default), an external PostgreSQL, MongoDB or ClickHouse. A durable outbox in Nakama's Postgres means a store that is down loses nothing.
- Import from Google Analytics 4 and GameAnalytics into the same store.
- A console section «آنالیتیکس بازی» / "Game analytics" with overview, active users, retention, cohorts, funnels, event explorer, session log, event catalog, storage, imports and settings.
- Every policy is an admin setting with a default. Cluster-safe by `docs/cluster-readiness.md`.

## Where it lives (decision)

The generic system extends the existing `integrations/insights` plugin (sessions, retention, heatmaps) instead of moving PP's `integrations/analytics` here:

- PP's plugin is mostly PP gameplay: stat cards from the editor, trophies, rules and rewards, all on `analytics_*` tables, the `analytics_*` RPCs, the `/v2/console/base/analytics` console prefix and the `nakama-analytics` console bundle. Moving it here would collide with PP's copy on every merge and bring a second trophy system next to `integrations/trophy`.
- Everything new uses the `insights` names, which PP does not use: tables `insights_*`, RPCs `insights_*`, console routes `/v2/console/base/insights/*`, console bundle `console/ui/insights` (index marker `nakama-insights`, hash routes `#/insights/*`, nav attribute `data-insights-group`). PP can run both plugins side by side.
- PP's SDK (`com.policepatrol.sdk` 0.5.21: `analytics_track` with GameAnalytics typed events) keeps working unchanged because PP keeps its own plugin. The new standard accepts the same GameAnalytics typed events (`design`, `progression`, `resource`, `business`, `error`) and the legacy `A:B:C` path, so PP can forward its events into the new store with one call to the server helper (`insightsapi.Send`) from its ingest, a PP-side change done when PP merges this work. PP's `gameevents` are picked up automatically (below).
- The generic parts of PP's ingest are re-implemented on the new standard, not copied: batch ids with stored receipts, per-event results, stored/flagged/rejected policies, client time window, database rate limits, typed event parsing.

## Event standard

### Client form

```
Analytics.Send(name: "DistanceOverall",
               tags: [{tagGroup: "mission", id: "3434343"}, {tagGroup: "car", id: "345345"}],
               value: 150)
```

On the wire one event is:

```json
{"name": "DistanceOverall", "tags": [{"tagGroup": "mission", "id": "3434343"}, {"tagGroup": "car", "id": "345345"}], "value": 150, "time": 1763700000000, "fields": {"weather": "rain"}}
```

| Field | Rule |
|---|---|
| `name` | `^[A-Za-z][A-Za-z0-9_.-]{0,63}$` (max length is a setting). Case-sensitive; PascalCase recommended. |
| `tags` | list of `{tagGroup, id}` (also accepted: an object `{"mission": "3434343"}`). Group key `^[a-z][a-z0-9_]{0,31}$`; id 1..64 printable chars, no `:`. One id per group; at most `events.maxTags` (default 8). |
| `value` | optional finite number inside the event's range (or the global range). Missing value: the event counts, sums treat it as 0. |
| `time` | optional client time (unix ms, unix s or RFC 3339). Outside the window (default 72 h back, 5 min ahead) the server time is used and the event is flagged `client_time`. |
| `fields` | optional free map, max `events.maxFields` (16) keys, strings up to `events.maxStringLength` (256). Stored, shown in timelines, not aggregated. |
| `path` | legacy form, instead of `name` + `tags`: `"DistanceOverall:345345:3434343"`. |
| `type` | GameAnalytics typed event instead of `name` (see below). |

### Legacy path

`DistanceOverall:CarId:MissionId` maps to the event named by the first part and the tag groups listed, in order, on that event's catalog entry (`pathGroups`, for example `["car", "mission"]`). Empty parts are skipped. Without a catalog entry (or with more parts than groups) the extra parts get the groups `p1`, `p2`, … and the event is flagged `path_groups`. Max 6 parts.

### GameAnalytics typed events

Accepted so the same client code (and imports from GameAnalytics) map to one model:

| `type` | Event name | Tags | Value |
|---|---|---|---|
| `design` (`id` = `a:b:c`, `value`) | treated as the legacy path `id` | from the path | `value` |
| `progression` (`status`, `p1`, `p2`, `p3`, `score`) | `progression` | `status`, `progression1..3` | `score` |
| `resource` (`flow`, `currency`, `amount`, `itemType`, `item`) | `resource` | `flow`, `currency`, `item_type`, `item` | `amount` (Sink is stored negative) |
| `business` (`currency`, `amount`, `itemType`, `item`, `cartType`, `store`) | `business` | `currency`, `item_type`, `item`, `cart_type`, `store` | `amount` |
| `error` (`severity`, `message`) | `error` | `severity` | — (`message` in `fields`) |

### Batch (envelope)

RPC `insights_send` (socket first: `socket.rpc("insights_send", json)`, HTTP `/v2/rpc/insights_send` as fallback). One batch:

```json
{"batchId": "b-01HX…", "sessionId": "s-…", "platform": "android", "appVersion": "1.4.0", "game": "", "events": [ … ]}
```

- `batchId` `^[A-Za-z0-9_.:-]{8,64}$` is required. Sending the same batch again is safe: the first answer is stored (`insights_batch`, kept `events.receiptDays`, default 7) and returned with `duplicate: true`.
- `sessionId` is the client session (the one the client also sends to `insights_session`); `platform` and `appVersion` are cleaned like the session RPC (32 chars, no control chars). `game` is a game key (`^[a-z0-9_-]{0,32}$`), empty = the app itself; a server hosting several mini games (gamecatalog) sends its key.
- At most `events.maxBatch` (default 100) events; a payload over 256 KB is refused by the RPC wrapper.

Answer:

```json
{"batchId": "…", "duplicate": false, "received": 3, "accepted": 2, "flagged": 1, "rejected": 0,
 "results": [{"index": 0, "status": "stored"}, {"index": 1, "status": "flagged", "code": "ANALYTICS_UNKNOWN_EVENT"}, …]}
```

`status` is `stored`, `flagged` (stored with a reason, kept out of dashboards unless "include flagged" is on), `rejected` (not stored, `code` says why) or `dropped` (event disabled in the catalog). Whole-batch errors: `ANALYTICS_BATCH_ID_REQUIRED`, `ANALYTICS_INVALID_BATCH_ID`, `ANALYTICS_BATCH_EMPTY`, `ANALYTICS_BATCH_TOO_LARGE`, `ANALYTICS_RATE_LIMITED` (429 with retry seconds), `ANALYTICS_BACKLOG_FULL` (503, the outbox is over its limit), `ANALYTICS_DISABLED`. Per-event codes: `ANALYTICS_INVALID_EVENT_NAME`, `ANALYTICS_INVALID_TAG`, `ANALYTICS_TOO_MANY_TAGS`, `ANALYTICS_DUPLICATE_TAG_GROUP`, `ANALYTICS_UNKNOWN_EVENT`, `ANALYTICS_UNKNOWN_TAG_GROUP`, `ANALYTICS_UNKNOWN_TAG_VALUE`, `ANALYTICS_MISSING_TAG`, `ANALYTICS_INVALID_VALUE`, `ANALYTICS_VALUE_REQUIRED`, `ANALYTICS_VALUE_OUT_OF_RANGE`, `ANALYTICS_INVALID_FIELDS`, `ANALYTICS_CLIENT_TIME_OUT_OF_RANGE`, `ANALYTICS_INVALID_TYPED_EVENT`.

`insights_config` (existing RPC) gains an `events` block: enabled, maxBatch, maxTags, flushSeconds (client hint, default 30), maxQueue (client hint, default 1000), and the enabled catalog (event names with their `pathGroups`, tag groups) so a client can validate before sending.

### Server events

- Go helper for other plugins: package `internal/insightsapi`:

  ```go
  insightsapi.Send(ctx, insightsapi.Event{UserID: uid, Name: "MatchWon",
      Tags: []insightsapi.Tag{{Group: "mode", ID: "ranked"}}, Value: insightsapi.Num(1), Key: "match:" + matchID})
  ```

  `Key` makes the event idempotent (the stored id is a UUIDv5 of it); without it every call is a new event. The insights plugin registers the sink at start; without the plugin `Send` returns `insightsapi.ErrNoSink` and nothing else happens. Server events skip rate limits but go through the same catalog checks, then into the outbox in the caller's context (one insert). `insightsapi.Path("Distance:SUV:SantaFe", 120)` builds a legacy-path event.
- `gameevents`: insights subscribes as `insights`. Every published game event (economy tasks, cup, clan, trophy, leaderboardx, quiz, campaign, and PP's) becomes a server event: name = event name (or the typed mapping), tag `context` = its context, numeric fields → `value` when the event has one numeric field named `value`/`amount`/`score`, the rest in `fields`, `Key` = the game event id. Setting `events.gameEvents` (default on).
- Session RPC `insights_session` and socket sessions also write `session_start` / `session_end` events (value = seconds on end) so session activity is in the store too.

## Storage

### Pipeline

1. Ingest validates the batch, writes the receipt and one outbox row per batch (`insights_outbox`: events as JSON, count, created_at, next_at, attempts, claimed_until, last_error) in one transaction in Nakama's Postgres.
2. A worker on every node claims due outbox rows with `UPDATE … WHERE id IN (SELECT … FOR UPDATE SKIP LOCKED LIMIT n) RETURNING` and a claim lease (`claimed_until`, default 60 s), writes the events to the selected store in one batch, then deletes the rows it still owns (`claim_token` match). On error the row gets `attempts+1`, `next_at = now() + backoff` (2 s doubling to 5 min) and `last_error`; after `storage.deadAfter` attempts (default 50) it is marked dead and shown in the console with "retry" and "discard".
3. Store writes are idempotent on the event id (UUIDv5 of user + batch + index, or of the server key), so a row delivered twice (a node died after the write and before the delete) is stored once.
4. `storage.maxPendingEvents` (default 2,000,000): above it new client batches get `ANALYTICS_BACKLOG_FULL` (clients keep their own queue); server events are still accepted up to twice the limit.

### Store interface (`integrations/insights/internal/store`)

```go
type Store interface {
    Kind() string
    Ensure(ctx) error                                    // create tables / collections / indexes
    Write(ctx, []Event) error                            // idempotent on Event.ID
    Aggregate(ctx, AggQuery) ([]AggRow, error)           // explorer
    Activity(ctx, ActivityQuery) (Activity, error)       // DAU / WAU / MAU / new users series
    Retention(ctx, RetentionQuery) (Retention, error)    // D1/D7/D30 and cohort table
    Funnel(ctx, FunnelQuery) (FunnelResult, error)
    Events(ctx, EventQuery) ([]Event, error)             // timelines, newest or oldest first
    Values(ctx, ValuesQuery) ([]ValueCount, error)       // distinct names / tag ids for pickers
    Ping(ctx) error
    Close() error
}
```

Drivers:

| Driver | Writes | Queries | Notes |
|---|---|---|---|
| `nakama` (default) | `insights_event` + rollups in Nakama's Postgres | SQL | tables from the migration |
| `postgres` | same schema in an external PostgreSQL (DSN, schema) | same SQL | `Ensure` creates the tables; pgx pool, max 4 conns per node |
| `clickhouse` | HTTP interface (`/?query=INSERT … FORMAT JSONEachRow`), `insert_deduplication_token` per batch | ClickHouse SQL (`uniqExact`, `windowFunnel`, `tags['car']`) | `ReplacingMergeTree` ordered by `(game, name, day, id)`, `Map(String,String)` tags; standard library only |
| `mongodb` | `insertMany` (unordered, duplicate `_id` ignored) | aggregation pipelines; funnels in Go over sorted events (capped) | own small OP_MSG + BSON + SCRAM-SHA-256 client (no new module); TLS optional |

Postgres schema (both Postgres drivers):

| Table | Columns |
|---|---|
| `insights_event` | `id UUID PK`, `game`, `name`, `value DOUBLE`, `tags JSONB` (`{"car":"345345"}`), `fields JSONB`, `user_id TEXT`, `session_id`, `at TIMESTAMPTZ` (client time or server time), `day DATE` (in the settings timezone), `server_at`, `platform`, `app_version`, `source` (`client`, `server`, `game_event`, `import:ga4`, `import:gameanalytics`), `flagged`, `flag_reason`. Indexes `(name, day)`, `(user_id, at)`, `(day)`, GIN `tags jsonb_path_ops`. |
| `insights_event_user` | `(game, user_id)` PK, `first_day`, `first_at`, `last_day`, `platform`, `app_version` — first seen, for new users and cohorts |
| `insights_event_user_day` | `(game, day, user_id)` PK, `events`, `platform`, `app_version` — DAU/WAU/MAU and retention |

User ids are text so imported players (`ga4:<pseudo id>`, `ga:<user id>`) live next to Nakama users. Rollups are written by the driver in the same statement batch as the events, only for events that were new (`INSERT … ON CONFLICT DO NOTHING RETURNING`), so a redelivered batch does not count twice.

Retention: `storage.keepDays` (default 400, 0 = forever) for raw events, swept in chunks under the `insights-event-sweep` lease (Postgres drivers; ClickHouse uses a table TTL; MongoDB a TTL index on `at`). Rollups follow `sessions.keepDays`.

### Switching stores

The selected driver and its config live in `insights_storage` (one row): driver, non-secret config (host, port, database, schema/collection, TLS, user), sealed secret (password / DSN password, `internal/secretbox`, key from `ECONOMY_SECRET_KEY`, info `insights-storage`). The console never gets the secret back, only `hasSecret`. "Test connection" pings with the form values before saving. Switching does not move old events: dashboards read only the current store; the outbox drains to whatever is selected when a row is delivered. Copying history between stores is a later item.

Each node opens the store lazily and reopens it when `insights_storage.updated_at` changes (checked at most every 10 s), so all nodes follow the console without shared memory.

## Query model

All console queries take: `from`, `to` (dates, inclusive, max 366 days), `game`, `platform`, `appVersion`, `includeFlagged`, and tag filters `tag.<group>=<id>[,<id>…]`.

- Explorer (`GET …/insights/explore`): `name` (one or more), `metric` = `count` | `sum` | `avg` | `min` | `max` | `users`, `groupBy` = up to 4 of `tag:<group>`, `name`, `platform`, `appVersion`, `user`, `source`, `game`; `bucket` = `none` | `hour` | `day` | `week`; `limit` (default 100, max 1000). Answer rows `{bucket, keys[], value, count, users}`. This answers "DistanceOverall total" (no group), "by user" (`groupBy=user`), "by car" (`tag:car`), "by mission", "by car + mission + user".
- Activity (`…/activity`): per day `dau`, `newUsers`, `events`; plus `dau`, `wau`, `mau` at `to`, and `stickiness` (dau/mau).
- Retention (`…/retention`): `days` (default from settings, 1/7/30), `mode` classic (came back exactly on day N) or rolling (on or after day N); answer summary per N and a cohort table by first day (`cohortRows`).
- Funnels (`…/funnels/{id}/run`): an admin-defined funnel = ordered steps (event name + optional tag filters), a window (default 24 h, 1 min..30 days). A user enters at the first step inside the range; each next step counts when it happens after the previous step and within the window from the first. Answer per step: users, conversion from step 1, conversion from previous, drop-off, median seconds from the previous step.
- Event log (`…/events`): newest first, by user and/or session, cursor paging.
- Sessions: from `insights_session` (always in Nakama's Postgres) with the existing filters; `…/sessions/{id}/timeline` lists that session's events from the store (by session id, falling back to the session's time window for the user).
- Values (`…/values?kind=name|tag:<group>`): distinct values with counts in the range for pickers.

Every list answer is bounded; CSV export is done in the browser from the shown rows (all rows of the query, up to the limit).

## Catalog

Per game (`game` key, empty = the app), edited in the console:

- Tag groups (`insights_tag_group`): key, labels fa/en, description, `closed` (only listed ids are valid), enabled. Values (`insights_tag_value`): group, id, labels fa/en. Import values from CSV text (`id,label_fa,label_en`) in the console.
- Events (`insights_event_def`): name, labels fa/en, description, enabled, `pathGroups` (ordered tag groups for the legacy path), `requiredTags`, `allowedTags` (empty = any group), value rule (`required`, `min`, `max`, `unit`).

Policies (settings `events`): `unknownEvents` accept / flag / reject (default accept), `unknownTagGroups` accept / flag / reject (default accept), `unknownTagValues` (for closed groups) flag / reject (default flag), `valueOutOfRange` flag / reject (default reject).

## Settings (`insights_settings`, new `events` and `storage` blocks)

| Setting | Default | Range |
|---|---|---|
| `events.enabled` | true | |
| `events.maxBatch` | 100 | 1..500 |
| `events.maxTags` | 8 | 1..16 |
| `events.maxNameLength` | 64 | 8..128 |
| `events.maxFields` / `maxStringLength` | 16 / 256 | 0..64 / 16..2048 |
| `events.valueMin` / `valueMax` | -1e12 / 1e12 | |
| `events.batchesPerMinute` / `eventsPerMinute` (per player) | 60 / 1200 | 1..10000 / 1..100000 |
| `events.pastHours` / `futureMinutes` (client time window) | 72 / 5 | 1..720 / 0..1440 |
| `events.receiptDays` | 7 | 1..90 |
| `events.unknownEvents` / `unknownTagGroups` / `unknownTagValues` / `valueOutOfRange` | accept / accept / flag / reject | |
| `events.gameEvents` | true | |
| `events.sessionEvents` | true | |
| `events.flushSeconds` / `maxQueue` (client hints) | 30 / 1000 | 5..3600 / 10..100000 |
| `storage.keepDays` | 400 | 0..3650 |
| `storage.batchSize` (outbox rows per claim) | 50 | 1..500 |
| `storage.maxPendingEvents` | 2000000 | 1000..100000000 |
| `storage.deadAfter` | 50 | 1..1000 |

Rate limits use `cluster_rate_limit` (database), keys `ins:b:<user>` and `ins:e:<user>`.

## Importers

Connectors (`insights_connector`): id, kind, name, enabled, config (non-secret JSON), sealed secret, schedule (minutes between pulls, default 60), cursor (JSON, per kind), next_run_at, lease (`claimed_until`, `claimed_by`), last status / error / counts. Runs (`insights_import_run`): connector, started, finished, status, events read / written / skipped, error. A worker on every node claims a due connector with `FOR UPDATE SKIP LOCKED` and a lease (renewed per page), so one node pulls each connector; imported events go through the same outbox. "Run now" sets `next_run_at = now()`.

| Kind | Source | Auth | Mapping |
|---|---|---|---|
| `ga4_bigquery` | GA4 BigQuery export: `jobs.query` on `` `project.dataset.events_*` `` with `_TABLE_SUFFIX` per day, paged with `pageToken` | Google service account JSON (sealed): JWT RS256 → `oauth2.googleapis.com/token`, scope `bigquery.readonly` | `event_name` → name (rename map in config), `event_params` → tags for the param keys listed in config (`paramTags`), one numeric param → value (`valueParam`), `user_id` or `ga4:` + `user_pseudo_id` → user, `ga_session_id` → session, `platform`, `app_info.version`, `event_timestamp` (µs) → time. Cursor = last finished day; reads up to `daysPerRun` days. Id = UUIDv5 of `event_timestamp`, user pseudo id, event name, bundle sequence |
| `ga4_data` | GA4 Data API `properties/{id}:runReport` by date + eventName (+ dimensions listed as tags) | same service account, scope `analytics.readonly` | aggregated: one event per row with `value` = `eventCount`, user `ga4:aggregate`, source `import:ga4_data`; for event totals only, not users |
| `gameanalytics` | GameAnalytics raw data export: gzip JSON-lines files listed from a URL prefix (S3-compatible bucket you own, or an HTTPS index) | access key/secret (S3 SigV4 via `internal/objectstore`) or bearer token | each line `{data: {category, event_id, amount, …, user_id, session_id, platform, build}, …}` → the typed mapping above (`design` → path, `progression`, `resource`, `business`, `error`; `user`/`session_end` → `session_start`/`session_end`). Cursor = last object key |

Real credentials do not exist yet: the importers are tested against fake HTTP servers (Google token + BigQuery + Data API, S3 listing + gzip objects); live runs are untested.

## Console (`console/ui/insights`, section «آنالیتیکس بازی»)

Routes `#/insights/<page>`: `overview`, `active`, `retention`, `funnels`, `explore`, `sessions`, `catalog`, `storage`, `imports`, `settings`. Every page has the date range (presets 7/30/90 days + custom), game, platform and version filters in a shared bar, kept in the hash query so a link reopens the same view; tables export to CSV; Persian digits in `fa`, Latin in `en`; Plane look (`console/ui/plane`), light/dark.

- Overview: KPI tiles (DAU today, WAU, MAU, new users, events, avg session), DAU + new users line chart, top events, platform split.
- Active users: DAU/WAU/MAU series, new vs returning bars.
- Retention: D1/D7/D30 summary tiles, cohort table (rows = first day, cells = % returned, heat colours).
- Funnels: list, editor (steps with event picker and tag filters, window), result bars with conversion and drop-off.
- Explorer: event picker, metric, group by (tags of the catalog + builtins), tag filters, bucket; chart (line by bucket or bars by key) and table.
- Sessions: list (player, platform, version, start, length, status), side panel with the session's event timeline; player filter.
- Catalog: events and tag groups (with values) per game, CSV value import.
- Storage: driver form (+ test connection), outbox status (pending, dead, oldest, last error), retry/discard dead rows.
- Imports: connectors (kind form, secret upload, schedule, run now, enable), runs list.
- Settings: the `events` and `storage` settings above.

The old «Insights» pages in the economy bundle (overview, sessions, heatmaps, settings) stay as they are in phase 1; overview and sessions there overlap the new section and can be removed later (owner decision).

## Console API (`/v2/console/base/insights`, console login)

| Method | Path | |
|---|---|---|
| GET/PUT | `/events/settings` | `events` + `storage` settings blocks |
| GET | `/events/catalog?game=` | tag groups, values, events |
| PUT/DELETE | `/events/catalog/groups/{key}?game=` | tag group (+ values) |
| PUT/DELETE | `/events/catalog/events/{name}?game=` | event definition |
| GET | `/explore`, `/activity`, `/retention`, `/events`, `/values` | queries |
| GET/POST | `/funnels`, PUT/DELETE `/funnels/{id}`, GET `/funnels/{id}/run` | funnels |
| GET | `/sessions/{id}/timeline` | session events |
| GET/PUT | `/storage`, POST `/storage/test` | driver config |
| GET | `/outbox`, POST `/outbox/retry`, POST `/outbox/discard` | queue |
| GET/POST | `/connectors`, PUT/DELETE `/connectors/{id}`, POST `/connectors/{id}/run`, GET `/connectors/{id}/runs` | importers |

## Cluster safety

- No shared state in memory: settings through `settingsrow` (TTL), store handle per node reopened on `updated_at`.
- Outbox and connector work claimed with `FOR UPDATE SKIP LOCKED` + leases; deletes guarded by the claim token.
- Rate limits in `cluster_rate_limit`. Sweeps under `clusterkit` leases.
- Idempotent store writes on event id; receipts per batch id.

## Client handoff

`docs/handoff/general/14-analytics-events.md`: the standard, `insights_send` over the socket with the batch format, queueing and retry rules, `insights_config` hints, examples (nakama-js), error codes. Unity: list of what `nakama-unity` needs (not changed in this work).

## Progress (2026-11-21, branch `analytics`)

Built:

- Event standard: `integrations/insights/internal/events` (name + tags + value, legacy path with catalog `pathGroups`, GameAnalytics typed events, policies, client time window). Unit tests.
- Server helper `internal/insightsapi` (`Send`, `Path`, `Tags`, `Num`); insights registers the sink and subscribes to `gameevents`; session RPC and socket sessions write `session_start` / `session_end` (sessions now store `game`).
- Client RPC `insights_send` with receipts, database rate limits, backlog guard; `insights_config` carries the event hints and catalog.
- Outbox `insights_outbox` + worker on every node (`FOR UPDATE SKIP LOCKED`, claim token, backoff, dead rows, retry/discard in the console), event sweep under a lease.
- Store drivers: Nakama Postgres (default) and external PostgreSQL (shared SQL, `Ensure` creates the schema), ClickHouse (HTTP interface, bound parameters, `ReplacingMergeTree`, `windowFunnel`), MongoDB (own OP_MSG/BSON/SCRAM-SHA-256 client in `internal/mongowire`, aggregation pipelines, funnels and retention computed in Go). Secrets sealed with `internal/secretbox`.
- Queries: explorer (metric × up to 4 group-bys × bucket × tag filters), activity (DAU/WAU/MAU/new/returning), retention summary + cohorts (classic/rolling), funnels, event log with cursor, values, session timelines.
- Importers: `ga4_bigquery`, `ga4_data`, `gameanalytics` (S3 raw export), connector table with leases and runs.
- Console bundle `console/ui/insights` (section «آنالیتیکس بازی»), every page listed above, CSV export, fa/en with Persian digits.
- Client handoff `docs/handoff/general/14-analytics-events.md`.

Tested:

- Go: events, store (Postgres contract on a disposable DB for both Postgres drivers; the same contract against an in-memory MongoDB wire server; ClickHouse against a fake HTTP server checking SQL and bound parameters), mongowire (BSON, RFC 7677 SCRAM vector, auth/insert/aggregate against a fake server), importers (fake Google token + BigQuery + Data API, fake S3), plugin (ingest, duplicates, rate limit, backlog, catalog policies, server helper, game events, outbox while the store is down then switch to an external Postgres schema, dashboards API, connector run).
- e2e `96-game-analytics` (catalog from the console, 4 device players over HTTP RPC, overview, active users, CSV, retention, explorer, funnel editor and result, session timeline, storage test, settings validation, imports form, mobile width, English).
- Real run: nakama-js client over the socket (`insights_session`, `insights_send` incl. duplicate batches, `insights_config`), then console API checks (explorer by tag combinations and by user, activity, retention, funnel, session timeline, empty outbox).

Not tested live: ClickHouse, MongoDB, GA4 (BigQuery and Data API), GameAnalytics S3 export: no servers or credentials on this Mac.

Left:

- PP: forward PP's `analytics_track` events into the new store from PP's ingest (`insightsapi.Send` with the typed events), done in the PP fork on merge.
- Copying history between stores when the driver changes.
- The old economy-bundle «Insights» overview/sessions tabs overlap the new section (owner decision to remove them).
- Unity wrappers (list in the handoff part).
