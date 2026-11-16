# Leaderboardx: admin boards, tiers, period results and rewards on Nakama leaderboards

A generic plugin (`integrations/leaderboardx`, `leaderboardx.so`) that turns Nakama's core leaderboards into admin-run boards: a translated name and description, an icon, a reset period, a sort order and score operator, an active window, where scores come from (economy events, server events of other plugins, the game itself, or server code only) with anti-cheat caps, tiers by score, rank or top percent, and rewards at the end of each period (wallet currencies and economy prizes or prize groups) paid once per player per period, automatically or when the player claims them. Nothing in it names a game; it can move into nakama-farsi as it is. It needs `devkit` in the same tree (it imports `apierr`) and uses the economy plugin at run time for prizes, prize groups and event types (without economy: currencies are still paid, economy prizes wait as failed payments).

Reference behaviour: the legacy .NET `Gamification/Leaderboard` module (`Leaderboard` with `LifeSpanHours`, `LifeCycle`, `EventType`, `EventCalculationType count | sum_value | high_value`, `Tiers [{Name, Order, Value, PrizeIds}]`, `PrizeIds` by rank, `Version`; `UserLeaderboard`; `LeaderboardHistory` with `Winners` and `TierWinners`; `LeaderboardScoreListener` on every user event; `LeaderboardResetService` that archives, bumps the version and grants prizes; `LeaderboardRankingService` ranks / neighbours / top-rank) and the Unity SDK `LeaderboardProvider` (`List`, `Get`, `GetByKey`, `Ranks`, `TopRank`, `Neighbours`, `History(id, version)`).

## Decisions

- **Core keeps the scores.** A board's key is a Nakama core leaderboard id (created authoritative, so clients cannot write it through the core API). Records, ranks, the rank cache, `LeaderboardRecordsList`, the haystack and the CRON reset all stay in core; the plugin never stores live scores. Writes go through `nk.LeaderboardRecordWrite` with the board's operator (`incr`, `best`, `set`, `decr`) and sort order. Any other server code (an economy game linked to the leaderboard id, a sample) can keep writing the same core leaderboard.
- **Legacy mapping.** `LifeSpanHours`/`LifeCycle` → a reset period (none, daily, weekly, monthly, custom CRON in UTC) + `maxPeriods`, plus an optional start/end window. `EventType` + `EventCalculationType` → a source (`economy_event` with `calc: count | value`; `count` adds 1, `value` with `incr` is `sum_value`, with `best` is `high_value`). Legacy tiers are score bands (`Value` is the exclusive upper bound of tier `Order`, players are ranked inside their band and `PrizeIds[i]` goes to in-band rank i+1); here the tier basis is an admin choice — `score` (lowest score of each tier, best tier first = legacy), `rank` (last rank) or `percent` (top percent of the period's players) — and reward rows say "overall ranks a..b" or "ranks a..b inside tier X" (`toRank 0` = the whole tier), so legacy `PrizeIds` by rank and `Tier.PrizeIds` are both rows. `Version` → a closed period row (`leaderboardx_period`); `LeaderboardHistory.Winners/TierWinners` → `leaderboardx_result` (every ranked player with rank, tier, rank in tier, score and reward state). `Leaderboard:Ranked` / `Leaderboard:TierRanked` user events → an economy player event and an analytics server event per rewarded player (setting).
- **Where scores come from (legacy ScoreListener).** Legacy counted every user event of the board's event type. Here the economy plugin (which owns player events in this fork) tells observers about every stored event through a new `internal/playerevents.Observe/Notify` slot (one line in economy's `recordEvent`); a board with source `economy_event` counts it once per event id. Source `game_event` follows `internal/gameevents` (a game plugin publishes its events with a context and number fields): an optional context filter and a number field (`stars`, `elapsedSeconds` …). Source `client` takes `leaderboardx_submit` only. Source `server` takes nothing automatically. Any board may also allow client submits (`limits.clientSubmit`).
- **Anti-cheat per board.** `maxPerWrite` (largest score of a `best`/`set` write, largest step of an `incr`/`decr` write) with `reject` or `clamp`, `perMinute` per player (cluster-wide counter `clusterkit.Allow`), plus a player cap over all boards (setting) for client submits; refused writes are logged and shown in the report and on the player card. Every write is logged once per source id (`ev:<economy event id>`, `ge:<server event id>`, `c:<user>:<requestId>`), so a resent event or submit never counts twice.
- **Period end without trusting one callback.** Core resets a leaderboard by moving new records to the next expiry time; the old records stay. The plugin registers `RegisterLeaderboardReset` (in this fork the scheduler runs it once per cluster under a `clusterkit` lease), but the callback only wakes the worker. The worker (every minute, under the `leaderboardx-worker` lease) archives every expiry time that has records, has passed and has no period row yet — so periods that ended while the server was down are caught up (older than `rewards.catchUpDays` are archived without rewards). A board whose window ended gets a final `end` period; `maxPeriods` stops a board after that many periods. Archiving is one transaction keyed on `(board, period end)`, so two nodes or a retry cannot archive twice. Note: Nakama keeps one leaderboard reset function for all Go modules; if another plugin registers one later, leaderboardx still works through its worker (at most a minute later).
- **Two-step pay.** The archive transaction writes the results with a payment plan (`pending` in auto mode, `claimable` with an expiry in claim mode). Payments run under the result's row lock: the wallet once per grant id (`leaderboardx:<period>:<user>` in `wallet_ledger` metadata, `sourceType: "leaderboard"`), each economy prize once per source id (`<period>:<user>:r<row>`, prize source type `leaderboard`). A failed payment stays `pending` and the worker retries it up to `rewards.maxTries`, then marks it `failed`; the winners page retries failed rows. A claim pays at once and leaves anything unpaid to the worker.
- **Board changes.** Core cannot change a leaderboard's sort order, operator or reset schedule. Saving such a change recreates the core leaderboard; when the running period has records the console answers `LEADERBOARD_RESTART_REQUIRED` with the player count and the admin confirms: the running period is archived first (`restart`, paid or not as the admin chooses). An admin can also close the running period at once (`manual`, paid or not); the core leaderboard then starts empty.

## Client RPCs (devkit envelope)

| RPC | In | Out |
|---|---|---|
| `leaderboardx_list` | `{lang?}` | `{serverTime, boards:[{key, name, description, iconUrl, sort, status (active, upcoming, ended), sortOrder, operator, period, resetSchedule, periodStart, periodEnd, startsAt, endsAt, clientSubmit, tierBasis, tiers:[{key, name, iconUrl, threshold}], rewards:[{tier, fromRank, toRank, currencies:[{currencyKey, amount}], prizeKey, prizeGroupKey}], rewardMinScore, payMode, metadata, me:{rank, score, subscore, tier, nextTier}\|null, claimable}]}` |
| `leaderboardx_board` | `{key, lang?, limit?, around?, cursor?}` | board fields + `{players, top:[{rank, userId, username, displayName, avatarUrl, score, subscore, tier, me}], around:[…], me:{rank, score, subscore, tier, nextTier:{key, index, need, score\|rank}}, nextCursor}`; an ended board answers its last archived period (`periodId`); `LEADERBOARD_NOT_FOUND`, `LEADERBOARD_KEY_REQUIRED`, `LEADERBOARD_INVALID_CURSOR` |
| `leaderboardx_submit` | `{key, score, subscore?, requestId}` | `{key, requestId, accepted, clamped, replayed, written, score, subscore, rank, tier}`; `LEADERBOARD_SUBMIT_NOT_ALLOWED` (403), `LEADERBOARD_NOT_ACTIVE` (409, `meta.status`), `LEADERBOARD_INVALID_SCORE`, `LEADERBOARD_SCORE_TOO_HIGH` (`meta.max`), `LEADERBOARD_RATE_LIMITED` (429), `REQUEST_ID_REQUIRED`, `INVALID_REQUEST_ID` |
| `leaderboardx_history` | `{key?, status?, limit?, offset?, lang?}` | `{items:[{periodId, boardKey, boardName, periodStart, periodEnd, players, rank, tier, tierRank, score, subscore, status (none, pending, claimable, paid, failed, expired), rewards:{currencies, prizes}, paid, claimExpiresAt, paidAt}], more, claimable}` |
| `leaderboardx_claim` | `{periodId}` or `{all: true}` | `{claimed:[{periodId, boardKey, rank, tier, status, paid:{currencies, prizes}, pendingPrizes, replayed}]}`; `LEADERBOARD_REWARD_NOT_FOUND`, `LEADERBOARD_REWARD_NOT_CLAIMABLE` (`meta.status`), `LEADERBOARD_REWARD_EXPIRED`, `LEADERBOARD_PERIOD_REQUIRED` |
| `leaderboardx_winners` | `{key, periodId?, limit?}` | `{key, period:{periodId, periodStart, periodEnd, players, reason}, winners:[{rank, userId, username, displayName, avatarUrl, score, tier, me, rewards}], periods:[last 20]}`; `LEADERBOARD_PERIOD_NOT_FOUND` |

Legacy Unity calls map as: `List`/`Get`/`GetByKey` → `leaderboardx_list` (+ `leaderboardx_board` for one), `Ranks`/`TopRank`/`Neighbours` → `leaderboardx_board` (`limit`, `cursor`, `around`), `History(id, version)` → `leaderboardx_winners` (`periodId`); per-tier pages are the `tier` of each row (a tier filter is proposed, not made). A player without a chosen username keeps Nakama's generated one.

## Admin settings (`leaderboardx_settings.value`)

| Key | Default | Choices |
|---|---|---|
| `display.topLimit` / `display.maxLimit` | 50 / 200 | rows of a board view when the game does not ask / most rows per call |
| `display.aroundLimit` | 5 | rows around the player (0 = none) |
| `display.showDisplayName` | true | rows carry display name and avatar url |
| `display.showUpcoming` | true | boards before their window are listed |
| `display.showEndedDays` | 3 | days an ended board stays listed (0 = hidden at once) |
| `rewards.payMode` | `auto` | `auto` (paid at period end), `claim` (the player claims); a board may override |
| `rewards.claimDays` | 7 | days a claimable reward waits, then `expired` |
| `rewards.stack` | `first` | `first` matching reward row only, or `all` rows added up |
| `rewards.prizeGroupMode` | `draw` | `draw` (one prize by weight, fixed per source), `all` |
| `rewards.maxTries` | 8 | payment attempts before `failed` |
| `rewards.catchUpDays` | 7 | a period that ended longer ago is archived without rewards (0 = always pay) |
| `scores.requireRequestId` | true | client submits need a `requestId` |
| `scores.playerPerMinute` | 60 | client submits per player per minute over all boards (0 = no cap) |
| `scores.logDays` | 30 | days the score log is kept |
| `history.keepDays` | 365 | days periods and results are kept (0 = forever; unsettled rewards stay) |
| `history.playerLimit` / `history.winnersLimit` | 20 / 100 | rows of `leaderboardx_history` / `leaderboardx_winners` |
| `events.economyEvent` / `events.analyticsEvent` | `leaderboard_ranked` | economy player event (value = rank) / analytics server event (context = board key) per rewarded player at period end; empty = none |
| `events.notify` / `events.notifyCode` | true / 0 | a persistent notification `leaderboard_reward` when a reward is paid or waits for a claim |

Per board: sort order, operator, period + CRON, window, `maxPeriods`, source (`type`, `event`, `context`, `field`, `calc`, `multiplier`), limits (`clientSubmit`, `maxPerWrite`, `perMinute`, `onExceed`), tier basis and tiers, reward rows, `rewardMinScore` (players below it are not rewarded), `payMode` ('' = settings).

## Data model

| Table | What |
|---|---|
| `leaderboardx_settings` | one JSON row (`internal/settingsrow`) |
| `leaderboardx_board` | key (= core leaderboard id), texts, icon, sort, active, sort order, operator, period, reset CRON, window, max periods, source, limits, tier basis, tiers, reward rows, min score, pay mode, periods closed, finished at, metadata |
| `leaderboardx_period` | one closed period: board key and name, start, end, core expiry time it read, reason (`reset`, `manual`, `restart`, `end`), players, rewarded, the board's tiers and rewards at that time, pay mode; unique on (board key, period end) |
| `leaderboardx_result` | every ranked player of a period: rank, tier, rank in tier, score, username, reward plan, status `none → pending → paid` (or `claimable → paid / expired`, `failed`), tries, last error, paid prizes |
| `leaderboardx_score_log` | writes through the plugin: board, source id (unique per board), player, source, value, written delta, outcome (accepted, clamped, too_high, rate_limited, admin set/add) |

Migration `20261207000000-leaderboardx-schema.sql` also adds `leaderboard` to `economy_user_prize_source_type_check`.

## Console («جدول امتیاز», `#/leaderboardx/…`)

A sidebar group with: **جدول‌ها** (list with status, period end, source, tiers, reward sentence and the live player count; a docked editor with key, translated name and description, icon, order, period with a CRON field, window and period count, sort order and operator, source with an event key list from economy, context and field for server events, calculation and multiplier, anti-cheat caps, tier basis and tier rows with a live sentence, reward rows (overall ranks or inside a tier) with currencies, prize and prize group pickers and a live sentence, minimum score and pay mode; a restart notice with «جایزهٔ دورهٔ جاری داده شود» when the change recreates the core leaderboard), **رتبه‌های زنده** (pick a board, period bounds and player count, rows with rank, player link, score and tier, find by username or id, change or remove a score, «بستن دوره همین حالا» with or without paying), **برنده‌های هر دوره** (periods with reason, players, rewarded and paid / queued / claimable / failed / expired counts, «تلاش دوباره» for failed payments, a panel with every result and its error), **گزارش جدول‌ها** (7 / 30 / 90 days: KPI tiles, scores per day, per board players, writes, refused writes, active players, closed periods and reward states) and **تنظیمات جدول‌ها** (every setting above). The native player page gets a «جدول امتیاز» card: the rank, score and tier on each board now, the last 30 results with rewards and payment state, refused writes in 30 days. fa/en, RTL, 390 px.

Console API under `/v2/console/base/leaderboardx`: `settings` (GET/PUT), `refs` (currencies, prizes, prize groups, economy event types, core leaderboards not yet adopted), `boards` (GET/POST; `{id}` GET/PUT/DELETE — `IN_USE` while rewards are unsettled, `?core=keep` keeps the core leaderboard; `{id}/standings` GET `limit`, `cursor`, `q`; `{id}/records/{userId}` PUT `{score, subscore, mode: set|add}` / DELETE; `{id}/close` POST `{pay}`), `periods` (GET `boardKey`, paging; `{id}` GET `status` (`rewarded` or a status), `q`, paging; `{id}/retry` POST), `players/{id}` (GET), `report` (`days`).

## Shared helpers

`internal/settingsrow`, `internal/pluginkit`, `internal/plugintest`, `internal/clusterkit` (worker lease, per-minute caps), `internal/economyhooks` (`GrantPrize`), `internal/gameevents` (subscribe to server events; publish `leaderboard_ranked`), `internal/playerevents` (record `leaderboard_ranked`; new `Observe`/`Notify`), `internal/activity` (`leaderboard_ranked`, `leaderboard_score_set`, `leaderboard_score_removed` rows on the player timeline), `internal/cronexpr` (period bounds).

## nakama-farsi core changes

Made: `internal/playerevents` gains `Recorded`, `Observe` and `Notify` (observers of stored economy player events); economy calls `playerevents.Notify` after each stored event (one line in `integrations/economy/events.go`); migration `20261207000000` widens `economy_user_prize_source_type_check` with `leaderboard`. No other existing file changed.

Proposed (not made): a tier filter for `leaderboardx_board` (legacy `Ranks(tierIndex)`); analytics stats as a score source (analytics would need to publish stat changes, e.g. through `internal/gameevents`); a client-side `LeaderboardProvider` in the generic SDK core mapping these RPCs; letting several Go modules register a leaderboard reset function (today the last one wins).
