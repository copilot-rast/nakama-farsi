# Campaign: chapters, stages, attempts, stars and freeride rewards

A generic plugin (`integrations/campaign`, `campaign.so`) for a game's mission map: chapters of stages that unlock in order or by rules, runs timed by the server clock, stars computed by the server from the run's numbers with admin rules, rewards that grow with the stars and pay only the difference on a replay, one-time stages (quizzes), repeatable stages and free ride rewards with a cooldown and a daily cap. Nothing in it names a game; it can move into nakama-farsi as it is. It needs `devkit` in the same tree (it imports `apierr`) and uses the economy plugin at run time for levels, metas, items and prizes (without economy: wallet currencies and campaign XP still work).

Reference behaviour: the legacy .NET `Gamification/Progress` module (`ProgressChapter`, `ProgressStage` with `Mode ordered|repeatable|once`, `StarRules`, `StarMultipliers`, `RewardGroup`, `CooldownSeconds`, `DailyRewardLimit`, `XpWhenLimited`, `StageAttempt` with a 24 h window, `ProgressRules.Payout = floor(amount × multiplier) − claimed`, chapter milestones by star percent) and the Unity SDK `StageProvider.Get/Start/Complete` (`PlayerProgress`, `StageAttempt`, `StageResult` with `RewardLimited`, `NextRewardAt`, `RewardsLeftToday`).

## Decisions

- **A plugin, not an economy edit.** Chapters, stages, attempts and their policies are their own tables and pages. Economy is reached only through slots: `internal/economyhooks` (`Level`, `GrantPrize`, `SettleMetas`) and `internal/playerevents`. XP is added to an economy meta row directly (`economy_user_meta`, the same upsert economy's grant uses) inside the completion transaction, then `economyhooks.SettleMetas` lets economy's level tables and their prizes follow.
- **The server owns the clock and the stars.** `campaign_attempt_start` stores `started_at` from the server; the game sends raw numbers (`success`, `elapsedSeconds`, `healthPercent`, any number `fields`). A run cannot be longer than the time since the start (+ grace) or longer than the time limit (+ grace), nor shorter than the stage's minimum (by the client time and by the server clock). Stars come from admin rules on `timeSavedPercent` (server computed), `elapsedSeconds`, `healthPercent` and the game's own fields; the game cannot send `timeSavedPercent`, `stars` and the like.
- **Two-step payment.** The completion transaction locks the attempt row and the player's stage row, scores the run, updates best stars, claimed amounts, the reward group and chapter milestones, adds XP and saves the answer and a payment plan on the attempt (`status = paying`). Payments then run under the attempt's row lock: wallet once per grant id (`campaign:<attemptId>` in `wallet_ledger` metadata), economy prizes once per source id, events once. A crash between the two leaves `paying`; a resend or the worker (every minute, under a lease) finishes it. Wallet changes carry `sourceType: "stage"`, so the ledger's `stage` kind lists them.
- **Resend is safe.** A completed attempt answers its stored result with `replayed: true` at any time; nothing is paid twice. A newer start closes older open attempts (setting) so an old result cannot be sent after a new run began.
- **Free ride.** Repeatable stages and free ride rewards share cooldowns and daily caps through a reward group key; a free ride reward is a claim without a stage (`campaign_freeride_claim`, idempotent per `requestId`).
- **Events.** Each completion records an economy player event (key from settings, value = stars) and publishes a server event through the new `internal/gameevents` slot; analytics subscribes and records it as a one-event batch with the stage key as `context` (batch id = attempt id), so trophies with a context key follow stage results. The analytics answer (trophy steps unlocked) is returned as `trophies`.

## Rules

### Stars

`stars.rules` (or the stage's own `starRules`): `[{stars, conditions:[{field, op, value}]}]`, ops `>= > <= < == !=`. The highest rule whose conditions all hold wins; a won run that matches none gets `stars.minOnSuccess` (1); a lost run gets 0. A condition on a field the game did not send holds or fails by `stars.missingField` (`fail`). Built-in fields: `timeSavedPercent = (limit − elapsed) / limit × 100` (100 without a limit; elapsed is clamped to the limit), `elapsedSeconds`, `healthPercent`. Defaults (legacy): 3 stars at ≥ 30 % saved and ≥ 75 % health, 2 stars at ≥ 10 % and ≥ 50 %.

### Payouts

The stage reward is the full (3-star) bundle `rewards.currencies` + `xp`; `starMultipliers` (stage) or `stars.multipliers` (0.2 / 0.5 / 1) scale it per star (`stars.repeatableMultipliers` 1 / 1 / 1 for repeatable stages).

| Mode | Opens | Pays |
|---|---|---|
| `ordered` | by its rules, else the settings default (all earlier ordered stages of the chapter have `unlock.minStars`) | `rewards.replayMode`: `difference` (default) pays `floor(full × mult) − already paid` per currency and XP; `none` pays only the first win; `full` pays every win |
| `once` | always (unless it has rules) | like `ordered`; with 1 / 1 / 1 multipliers the first pass pays everything and later passes pay nothing |
| `repeatable` | always (unless it has rules) | `floor(full × mult)` every rewarded win |

`tierPrizes` `[{stars, prizeKey | prizeGroupKey}]` pay an economy prize when the best stars first reach the tier (source id `user:stage:tN`); on repeatable stages every rewarded win pays the tiers it reached (source `attempt:tN`). A stage with `cooldownSeconds` or `dailyLimit` counts a reward in its reward group (`rewardGroup`, else its own key); when limited the answer says `rewardLimited: "cooldown" | "daily_limit"` and pays nothing (XP too unless `xpWhenLimited`). The day starts in `rewards.dayZone`.

Chapter `milestones` `[{percent, currencies, prizeKey, prizeGroupKey}]` pay once when the chapter's star percent (stars of its counted stages / (stages × max stars); counted modes = `rewards.milestoneModes`, default ordered) reaches `percent`.

### Unlock rules

Chapters and stages take `[{type, key, value, table}]`, all must hold: `level` (player level ≥ value in the level table `table`, empty = economy's main table), `meta` (economy meta `key` ≥ value), `item` (owns economy item `key`, starters count), `stage` (stage `key` has ≥ value stars, 0 = `unlock.minStars`), `chapter` (chapter `key` star percent ≥ value, 0 = every counted stage passed). A chapter without rules is open, or (setting `unlock.chapterDefault = previous`) opens when the chapter before is finished. A stage's chapter rules apply to it too. Start and complete both check; `STAGE_LOCKED` (403) has `meta.locks` with the unmet rules.

## Client RPCs (devkit envelope)

| RPC | In | Out |
|---|---|---|
| `campaign_chapters` | `{lang?, chapter?}` | `{serverTime, xp, level, maxStars, chapters:[{key, name, description, iconUrl, address, sort, unlocked, locks, stars, maxStars, starPercent, milestones:[{percent, rewards, prizeKey, prizeGroupKey, claimed, reached}], stages:[{key, name, description, iconUrl, address, sort, mode, unlocked, locks, bestStars, completions, wins, bestSeconds, rewards:[{currencyKey, amount}], xp, tierPrizes, starMultipliers, claimed, claimedXp, timeLimitSeconds, rewardGroup, cooldownSeconds, dailyRewardLimit, nextRewardAt, rewardsLeftToday, metadata}]}], freeride:[{key, name, rewards, prizeKey, prizeGroupKey, xp, minLevel, rewardGroup, cooldownSeconds, dailyRewardLimit, nextRewardAt, rewardsLeftToday}]}` |
| `campaign_attempt_start` | `{stageKey}` | `{attemptId, stageKey, startedAt, expiresAt, timeLimitSeconds, playCost: {currencyKey, amount, spent, free, via} \| null, regen?}`; `STAGE_NOT_FOUND`, `STAGE_LOCKED`, `NOT_ENOUGH_PLAY_COST` (409, `meta {currencyKey, amount, value?, nextAt?, fullAt?, refillPrice?}`) |
| `campaign_attempt_complete` | `{attemptId, stageKey?, success, elapsedSeconds, healthPercent?, fields?}` | `{attemptId, stageKey, chapterKey, mode, success, playCostRefund: {currencyKey, amount} | null, stars, bestStars, previousBestStars, rewards, milestoneRewards, milestones, prizes:[{key, items, for}], pendingPrizes, xpGained, xp, level, levels?, rewardLimited, nextRewardAt, rewardsLeftToday, trophies?, serverSeconds, replayed}`; `ATTEMPT_REQUIRED`, `ATTEMPT_NOT_FOUND`, `ATTEMPT_CLOSED`, `ATTEMPT_EXPIRED`, `INVALID_RESULT`, `STAGE_FINISHED_TOO_FAST`, `STAGE_LOCKED`, `TOO_MANY_FIELDS`, `STAGE_BUSY` (409, send once more) |
| `campaign_freeride_claim` | `{key, requestId}` | `{key, requestId, rewards, xpGained, xp, level, prizes, nextRewardAt, rewardsLeftToday, replayed}`; `FREERIDE_NOT_FOUND`, `REQUEST_ID_REQUIRED`, `LEVEL_TOO_LOW`, `REWARD_LIMITED` (409, `meta.reason`, `nextRewardAt`, `rewardsLeftToday`) |

`INVALID_RESULT` and `STAGE_FINISHED_TOO_FAST` are counted on the attempt (`rejected`, `last_reject`; shown on the player card and in the report). The attempt stays open, so a correct resend still works.

## Admin settings (`campaign_settings.value`)

| Key | Default | Choices |
|---|---|---|
| `play.costCurrency` | `""` | wallet currency taken at every start (empty = no cost, PP behaviour) |
| `play.costAmount` | 1 | 1–1 000 000 per start; a stage's `metadata.playCost` (whole number ≥ 0, 0 = free) overrides it |
| `play.refundOnWin` | true | a won run gives the cost back, once |
| `attempts.hours` | 24 | 1–168: how long a started run can be completed |
| `attempts.graceSeconds` | 3 | 0–120 added to the server clock and the time limit |
| `attempts.defaultMinSeconds` | 5 | 0–3600: shortest won run of a stage without its own minimum |
| `attempts.closeOthers` | true | a start closes the player's other open attempts |
| `attempts.serverClosesOthers` | false | a stage started by the server (campaignhooks.Start, e.g. a quiz pack) also closes the other open attempts |
| `attempts.keepDays` | 90 | finished attempts kept (0 = forever) |
| `stars.rules` | legacy rules | star rules (see above) |
| `stars.multipliers` / `stars.repeatableMultipliers` | 0.2, 0.5, 1 / 1, 1, 1 | one per star, 0–10 |
| `stars.maxStars` | 3 | 1–5 |
| `stars.minOnSuccess` | 1 | stars of a win matching no rule |
| `stars.missingField` | `fail` | `pass`, `fail` |
| `unlock.stageDefault` | `all_previous` | `all_previous`, `previous`, `none` |
| `unlock.chapterDefault` | `none` | `none`, `previous` |
| `unlock.minStars` | 1 | stars that pass a stage |
| `unlock.showLocked` | true | locked chapters and stages are listed with their rules |
| `rewards.replayMode` | `difference` | `difference`, `none`, `full` |
| `rewards.dayZone` | `UTC` | any IANA zone |
| `rewards.prizeGroupMode` | `draw` | `draw` (one prize by weight, fixed per source), `all` |
| `rewards.milestoneModes` | `["ordered"]` | modes counted in a chapter's star percent |
| `xp.metaKey` | `xp` | economy meta that receives XP (empty = campaign only) |
| `xp.levelTableKey` | `""` | economy level table for the player level (empty = main) |
| `events.economyEvent` / `events.analyticsEvent` | `stage_completed` | event key / name (empty = none) |
| `events.onFailure` | true | lost runs send the events too |
| `freeride.requireRequestId` | true | claims need a `requestId` |
| `freeride.requestHours` | 24 | how long a claim answer is kept |
| `freeride.tableEnabled` | false | the separate `campaign_freeride` table path; off = `freeride[]` is empty and `campaign_freeride_claim` answers `FREERIDE_NOT_FOUND` (free ride runs as repeatable stages) |
| `stars.templates` | `base`, `race`, `stealth`, `quest` | named star rule sets a stage can pick (`starTemplate`); base = time + health + `timesSeen` (missing passes), race = no health, stealth = no time, quest = base |
| `fields.stage` / `fields.chapter` | `[]` | field definitions for the stage / chapter `settings` form (types `number`, `integer`, `bool`, `text`, `i18n`, `select`, `tags`, `list` up to 3 levels; `default`, `min`, `max`, `options`, `maxItems`, `chapters`) |

### Stage settings, templates and result details (2026-11-22)

- `campaign_stage.settings` / `campaign_chapter.settings` (JSONB) hold values checked against `fields.*` (`UNKNOWN_FIELD`, `INVALID_FIELD_VALUE`, `INVALID_OPTION`, `VALUE_OUT_OF_RANGE`, `TOO_MANY_ITEMS`). `campaign_chapters` returns them resolved: defaults filled, `i18n` text in the player's language. Stages also return `starRules`, `starTemplate`, `minSeconds` (stage or default), `noReplayAfterWin`, `replayable`.
- Star rule order: own `star_rules`, else the stage's `star_template`, else `stars.rules` (default = the base template). A condition may carry `missing: pass|fail` over `stars.missingField`.
- `no_replay_after_win` (default off): a player with a win gets `STAGE_ALREADY_WON` on start and complete.
- `campaign_attempt_complete` adds `criteria` (`[{field, value, stars}]`, the best tier each criterion reached), `starRules`, `minSeconds`, `timeSavedPercent`, `milestoneIds`, `replayable`.
- Chapter milestones have a fixed `id` (the console gives new steps the next free number). Paid steps are stored by id in `campaign_user_chapter.milestones`, so changing a percent never pays twice; old numeric entries still count by percent. A step is reached when the star percent is within 0.005 of it, so 16.67 is paid at exactly 1/6.
- Migration `20270126000000-campaign-stage-settings.sql`: the new columns, ids `1..n` (array order) for chapters without ids, paid percents turned into ids.

### Play cost

With `play.costCurrency` set, `campaign_attempt_start` spends the cost before it creates the attempt, through `economyhooks.SpendRegen` with source `campaign:<attemptId>` (regen currencies settle first; a player holding the currency's unlimited tag plays free). Without economy it is a plain wallet debit (grant id `campaign:<attemptId>:cost`). A refusal creates no attempt and answers `NOT_ENOUGH_PLAY_COST` with the next refill time. The cost is stored on the attempt (`campaign_attempt.play_cost`, migration `20270103000001-campaign-play-cost.sql`). A won completion with `play.refundOnWin` puts the refund in the payment plan, so it runs once with the other payments (and the payment worker retries it): `RefundRegen`, or a wallet credit with grant id `campaign:<attemptId>:refund`. Lost, abandoned, expired and closed attempts keep the cost. Server-owned stages started through `campaignhooks.Start` pay the cost too.

## Data model

| Table | What |
|---|---|
| `campaign_settings` | one JSON row (`internal/settingsrow`) |
| `campaign_chapter` | key, translated name/description, icon url, Addressables address, sort, active, `unlock` rules, `milestones`, metadata |
| `campaign_stage` | key (the game's stage id), chapter, texts, icon, address, sort, active, `mode`, `unlock` (null = default), time limit, min seconds, `star_rules` / `star_multipliers` (null = settings), `rewards`, `xp`, `tier_prizes`, reward group, cooldown, daily limit, `xp_when_limited`, metadata |
| `campaign_freeride` | key, texts, rewards, prize / prize group, xp, reward group, cooldown, daily limit, min level |
| `campaign_attempt` | id = attempt token, user, stage, server `started_at`, `expires_at`, `status open → paying → done` (or `closed`), stars, stored `result` and payment `plan` (incl. the play cost refund), `rejected`, `play_cost` |
| `campaign_user_stage` | best stars, `claimed` per currency, claimed XP, tiers paid, completions, wins, best seconds |
| `campaign_user` / `campaign_user_chapter` | XP total / milestones paid |
| `campaign_reward_group` | cooldown and daily cap state per group |
| `campaign_claim` | free ride claims per `requestId` (answer + plan) |

Economy: migration `20261204000000` adds `stage` (and `trophy`) to `economy_user_prize_source_type_check`, so prizes paid for stages are rows with `source_type = 'stage'`.

## Console («مرحله‌ها», `#/campaign/…`)

A sidebar group (same shell pattern as the other plugin pages) with: **فصل‌ها** (order with ↑↓, unlock rules with a live sentence, star milestones with currencies and prize pickers, icon and address), **مرحله‌ها** (filter by chapter and kind, order within a chapter, editor with kind hints, time limit and minimum, own unlock rules or the default, own star rules with a live sentence per rule including custom game fields, multipliers, full reward + XP with a per-star payout preview, star prizes from economy's prize and prize group lists, cooldown / daily cap / reward group, metadata), **جایزهٔ گشت آزاد**, **گزارش مرحله‌ها** (7 / 30 / 90 days, by chapter: players, attempts, wins, completion rate, average best stars, drop-off from the previous story stage, refused results, KPI tiles) and **تنظیمات مرحله‌ها** (every setting above with «هزینهٔ بازی» first, warnings when economy, the XP meta or the cost currency is missing, a note when the cost currency does not regenerate). The native player page gets a «پیشرفت مرحله‌ها» card: XP and level, each chapter's star bar, every stage with stars and wins, «از نو» per stage (keeps paid amounts unless «جایزه‌های گرفته‌شده هم پاک شود» is ticked) and the last attempts with refused results. fa/en, RTL, 390 px.

Console API under `/v2/console/base/campaign`: `settings` (GET/PUT), `refs` (economy and campaign keys for pickers), `chapters` (GET/POST, `order` PUT, `{id}` GET/PUT/DELETE — `IN_USE` while it has stages), `stages` (GET with `chapterId`, `mode`, `q`, paging; POST, `order`, `{id}` GET/PUT/DELETE — `IN_USE` once played), `freeride` (GET/POST, `{id}` PUT/DELETE), `players/{id}` (GET), `players/{id}/stages/{stageId}` (DELETE, `?claims=reset`), `report` (`days`, `chapterId`).

## Shared helpers

`internal/settingsrow`, `internal/pluginkit`, `internal/plugintest`, `internal/clusterkit` (worker lease), `internal/activity` (`stage_completed`, `stage_reset` rows on the player timeline), and the new `internal/gameevents` (publish a server-side player event to subscribers by owner; analytics is the first).

## nakama-farsi core changes

Made (owner-approved): economy registers `economyhooks.SetPrizeGranter` and `SetMetaSettler` (`integrations/economy/module.go`, `hooks.go`) and migration `20261204000000-economy-prize-sources.sql`. Every `internal/gameevents` subscriber (trophy, leaderboardx, clan, a fork's analytics) gets `stage_completed`. No other existing file changed.

Proposed (not made): a client-side `StageProvider` in the generic SDK core mapping these RPCs; partitioning `campaign_attempt` by month for very large games.
