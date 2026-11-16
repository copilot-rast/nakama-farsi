# Quiz: packs, questions, hints, missions, reports and the campaign stage link

A generic plugin (`integrations/quiz`, `quiz.so`) for in-game quizzes: packs of options or text questions (their own, or drawn from a category bank), timed by the server clock, scored on the server (the right answers never reach the client before the player answers unless the admin allows it), idempotent answers, hints paid once from the wallet, rewards paid once through the wallet and economy prizes, missions and games that group packs, player reports on questions reviewed in the console, result reports with placeholders and webhooks, and a link that lets a pack finish a `once` campaign stage on the server. Nothing in it names a game; it can move into nakama-farsi as it is. It needs `devkit` in the same tree (it imports `apierr`) and uses at run time the economy plugin (prizes, player events), the campaign plugin (stage link) and analytics (game events); without them wallet rewards and everything else still work.

Reference behaviour: the legacy .NET `Gamification/Quiz` module (`QuizPack` with `WinThreshold`, `Time`, `SendAnswers`, `PrizeId`; `Quiz` + `QuizOptionsAnswer` / `QuizTextAnswer`; `QuizHint` with a one-currency price; `QuizMission` and `QuizGame` with `Order sequential|all_at_once`, `Time`, `PrizeId`; `UserQuizPack` sessions with `QuizDeadlines`; `UserQuizPackStatistics`; `QuizReportTemplate` with placeholders and `QuizReportDelivery` webhooks; the Excel import) and the Unity SDK providers `QuizProvider`, `QuizPackProvider`, `QuizGameProvider`, `QuizHintProvider`, `QuizMissionProvider`, `QuizReportProvider`. Legacy linked quizzes to Progress only by convention: four `once` stages (`quiz-…`) whose result the client sent after checking the answers on the device.

## Decisions

- **A plugin, not an economy edit.** Packs, questions, sessions and their policies are their own tables and pages. Economy is reached only through slots: `internal/economyhooks.GrantPrize` (source type `quiz`), `internal/playerevents` (task / achievement events) and `internal/gameevents` (analytics).
- **The server owns the answers.** Options are stored with `correct`; a session view, `quiz_packs` and a hint never carry it. The answer result reveals the right options (and the explanation) by `reveal.mode`; packs and missions marked `sendAnswers` send them with the questions only when `reveal.allowBeforeAnswer` is on (legacy `SendAnswers`).
- **The server owns the clock.** A session stores `started_at`, the pack deadline and, per question, when its clock started: the previous answer or the start (`play.delivery = all`, legacy) or the moment the question was sent (`one_by_one`, where the client only ever sees the current question). A late answer is stored as `late` (wrong) or refused (`timing.lateAnswer`). After the pack time the session is scored with what it has (legacy closes with award) — by the next answer, `quiz_play` or the worker.
- **Resend is safe.** One answer row per (session, question): a resend answers the stored outcome with `replayed: true`. A finished session answers its stored result. `quiz_play` on a pack with an open session resumes it (`play.onRestart = resume`) or closes it without a reward (`discard`, legacy).
- **Two-step payment** (like campaign). The finish transaction locks the session, counts the answers, updates the player's pack row, question rewards, missions and day counters and stores the result and a payment plan (`status = paying`). The plan then pays under the session row lock: wallet once per grant id (`quiz:<sessionId>` in `wallet_ledger` metadata, `sourceType: "quiz"`), economy prizes once per source id (`pack:<user>:<pack>`, `question:<user>:<question>`, `mission:<user>:<mission>`, or per session for `every_*` modes), the campaign completion once (the campaign attempt is idempotent), events once. A failure leaves `paying`; the worker retries.
- **Hints** are bought once per (player, hint, scope): the `quiz_user_hint` row is locked while `pluginkit.Debit` takes the price with grant id `quiz:hint:<user>:<hint>:<scope>`, so parallel or repeated buys pay once.
- **Limits live in the database.** Daily plays, rewarded passes and reports are counted in a locked `quiz_user_day` row; plays per pack and cooldowns come from `quiz_user_pack`; one open session per (player, pack) is a unique index. All of it is cluster-safe.
- **Missions vs economy tasks.** Quiz missions keep legacy semantics (ordered packs, a time from the first start, a one-time reward; games group missions). Anything task-like (answer 50 questions, pass 3 quizzes this week) is left to economy tasks / achievements: every finished session records the economy player events `quiz_completed` (value = score), `quiz_passed` (1) and `quiz_correct` (right answers), with `packKey`, `missionKey`, `category` in the metadata, and publishes the game event `quiz_completed` (context = pack key) to analytics for trophies.
- **Campaign link.** A pack may name a campaign stage (`campaignStageKey`, normally a `once` stage). `quiz_play` starts the stage attempt through the new slot `internal/campaignhooks` (the stage's unlock rules apply), the finish completes it on the server with `success = passed` and the server's elapsed time, and the campaign pays the stage reward as usual (first pass only for `once`). The campaign refuses `campaign_attempt_complete` from the client for a stage a quiz pack owns (`STAGE_SERVER_OWNED`, 403) and lists the owner on the stage (`completedBy: {owner: "quiz", ref: packKey}`) in `campaign_chapters`.

## Client RPCs (devkit envelope)

| RPC | In | Out |
|---|---|---|
| `quiz_packs` | `{lang?, missionKey?, category?, tag?}` | `{lang, serverTime, packs:[{key, name, description, iconUrl, sort, category, tags, language, difficulty, missionKey, timeSeconds, questionSeconds, questionCount, passCorrect, reward:{currencies:[{currencyKey, amount}], prizeKey, prizeGroupKey}, campaignStageKey, metadata, plays, wins, bestScore, lastScore, passed, rewarded, locked:{code, meta}\|null, openSessionId}]}` |
| `quiz_pack_random` | `{lang?, categories?, tags?, forceNew?, preferNew?}` | one pack as above; `QUIZ_NO_PACK_FOUND` |
| `quiz_play` | `{packKey, lang?, metadata?}` | session view `{sessionId, packKey, status, startedAt, deadlineAt, expiresAt, serverTime, delivery, total, answered, correct, current, passCorrect, campaignAttemptId, resumed, questions:[{id, index, type, text, mediaUrl, difficulty, timeSeconds, deadlineAt?, options?:[{id, text, mediaUrl}], answered:{outcome, optionIds, text, correctOptionIds?, acceptedAnswers?, explanation?}\|null, hints:[{id, name, price, owned, text?, mediaUrl?}]}]}`; `QUIZ_PACK_NOT_FOUND`, `QUIZ_MISSION_PACK_LOCKED`, `QUIZ_GAME_MISSION_LOCKED`, `QUIZ_MISSION_TIME_OVER`, `QUIZ_GAME_TIME_OVER`, `QUIZ_PACK_ALREADY_PASSED`, `QUIZ_PACK_PLAY_LIMIT`, `QUIZ_PACK_COOLDOWN` (429), `QUIZ_DAILY_PLAY_LIMIT` (429), `QUIZ_PACK_EMPTY`, campaign `STAGE_LOCKED` |
| `quiz_session` | `{sessionId? \| packKey?, lang?}` | the session view (+ `result` when finished); `QUIZ_SESSION_NOT_FOUND` |
| `quiz_answer` | `{sessionId, questionId, optionId? \| optionIds? \| text?, lang?}` | `{sessionId, questionId, outcome: correct\|wrong\|skip\|late, correct, replayed, answered, correctCount, total, correctOptionIds?, acceptedAnswers?, explanation?, next?, result?}`; `QUIZ_QUESTION_NOT_IN_SESSION`, `QUIZ_OPTION_NOT_FOUND`, `QUIZ_QUESTION_NOT_CURRENT` (409, one by one), `QUIZ_QUESTION_TIME_OVER` (409, reject mode), `QUIZ_TIME_OVER` (409, the session was scored), `QUIZ_SESSION_CLOSED` |
| `quiz_answer_all` | `{sessionId, answers:[{questionId, optionId?, optionIds?, text?}]}` | `{sessionId, answers:[…], result}` (legacy answer-all) |
| `quiz_finish` | `{sessionId}` | result `{sessionId, packKey, passed, score, correct, total, passCorrect, answered, expired, rewards, questionRewards, missionRewards, missionsCompleted, rewardLimited, prizes, pendingPrizes, review?:[{questionId, outcome, optionIds, text, correctOptionIds|acceptedAnswers, explanation (answered questions only)}], campaign?, trophies?, serverSeconds, replayed}` |
| `quiz_hint_buy` | `{hintId, sessionId?, lang?}` | `{hintId, questionId, name, text, mediaUrl, price, owned, replayed}`; `QUIZ_HINT_NOT_FOUND`, `QUIZ_INSUFFICIENT_CURRENCY_FOR_HINT`, `QUIZ_HINT_LIMIT`, `QUIZ_SESSION_REQUIRED` (keep = session) |
| `quiz_missions` | `{lang?, kind?: mission\|game, key?}` | `{missions:[{key, kind, parentKey, name, description, iconUrl, sort, orderMode, timeSeconds, reward, metadata, startedAt, expiresAt, timeOver, completedAt, rewarded, items:[{key, name, sort, passed\|completed, playable}]}]}` or one mission with `key` |
| `quiz_mission_start` | `{key}` | the mission (its clock starts once; legacy start) |
| `quiz_stats` | `{packKey?}` | `{plays, wins, packsPassed, answers, correctAnswers}` or one pack's `{plays, wins, bestScore, lastScore}` |
| `quiz_report_reasons` | `{lang?}` | `{enabled, reasons:[{key, name}]}` |
| `quiz_report_question` | `{questionId, reasonKey, note?, sessionId?}` | `{reportId, duplicate}`; `QUIZ_REPORTS_CLOSED`, `QUIZ_REPORT_NOT_SEEN`, `QUIZ_REPORT_REASON_NOT_FOUND`, `QUIZ_REPORT_NOTE_REQUIRED`, `QUIZ_REPORT_LIMIT` (429) |
| `quiz_result_report` | `{templateKey}` | `{templateKey, name, missionKey, body, values, createdAt}` (and a webhook delivery is queued when the template has one) |

Text answers match an accepted answer in any language; `normalized` matching folds case, Arabic/Persian letter forms (ي/ی, ك/ک, ة/ه, أ/ا…), Persian and Arabic digits, diacritics, half spaces and repeated spaces. A question with several right options is right with any one of them (`any`, legacy) or only with exactly all of them (`all`).

## Admin settings (`quiz_settings.value`)

| Key | Default | Choices |
|---|---|---|
| `play.delivery` | `all` | `all` (legacy), `one_by_one` |
| `play.shuffleQuestions` / `play.shuffleOptions` | true / true | packs may override (null = settings) |
| `play.onRestart` | `resume` | `resume`, `discard` |
| `play.sessionHours` | 24 | 1–720: open session lifetime without a pack time |
| `play.autoFinish` | true | score after the last answer |
| `timing.questionSeconds` | 0 | 0–3600 (question → pack → settings) |
| `timing.graceSeconds` | 2 | 0–60 |
| `timing.lateAnswer` | `wrong` | `wrong` (stored as `late`), `reject` |
| `scoring.passPercent` | 100 | 0–100 (legacy `WinThreshold`); packs: `passCorrect` or `passPercent` |
| `scoring.multiCorrect` | `any` | `any`, `all` |
| `scoring.textMatch` | `normalized` | `normalized`, `exact` |
| `reveal.mode` | `after_answer` | `after_answer`, `after_finish`, `never` |
| `reveal.allowBeforeAnswer` | false | honour pack / mission `sendAnswers` |
| `reveal.explanation` | true | |
| `rewards.packMode` | `first_pass` | `first_pass`, `every_pass` (legacy), `none` |
| `rewards.questionMode` | `first_correct` | `first_correct`, `every_correct` (legacy), `none` |
| `rewards.prizeGroupMode` | `draw` | `draw`, `all` |
| `rewards.dayZone` | `UTC` | IANA zone for the daily limits |
| `limits.dailyPlays` / `limits.dailyRewardedPasses` / `limits.maxPlaysPerPack` | 0 / 0 / 0 | 0 = none |
| `limits.replayAfterPass` | true | |
| `limits.retryCooldownSeconds` | 0 | |
| `hints.pricing` | `per_hint` | `per_hint`, `fixed` (`hints.fixedPrice`), `free` |
| `hints.keep` | `forever` | `forever` (legacy), `session` |
| `hints.maxPerQuestion` | 0 | 0 = none |
| `reports.enabled` / `reports.requireSeen` / `reports.noteRequired` | true / true / false | |
| `reports.dailyLimit` | 20 | |
| `reports.autoDisableAt` | 0 | open reports from different players that switch a question off |
| `events.completed` / `events.passed` / `events.correct` | `quiz_completed` / `quiz_passed` / `quiz_correct` | economy player event keys (empty = none) |
| `events.analytics` | `quiz_completed` | game event name (context = pack key) |
| `events.onFailure` | true | |
| `campaign.enabled` / `campaign.sendFailures` | true / true | the stage link; failed quizzes complete the attempt with `success=false` |
| `retention.sessionDays` / `retention.deliveryDays` / `retention.webhookAttempts` | 180 / 30 / 5 | |

## Data model (migration `20261208000000-quiz-schema.sql`)

| Table | What |
|---|---|
| `quiz_settings` | one JSON row (`internal/settingsrow`) |
| `quiz_mission` | key, `kind mission\|game`, `parent_id` (mission → game), translated name/description, icon, sort, `order_mode`, `time_seconds`, `send_answers`, `reward`, active, metadata |
| `quiz_pack` | key, mission, texts, icon, sort, category, tags, language, difficulty, `time_seconds` (whole pack), `question_seconds`, `pass_percent` / `pass_correct`, `question_count` (0 = all), `pool_category` (draw from the category bank), shuffle overrides, `send_answers`, `reward`, `campaign_stage_key` (unique when set), active, metadata |
| `quiz_question` | optional unique key (import upsert), pack (null = bank), `type options\|text`, translated question, media url, `options [{id, text, mediaUrl, correct}]`, `answers {lang:[…]}`, explanation, difficulty, category, tags, own time, `reward`, sort, active, metadata |
| `quiz_hint` / `quiz_user_hint` | hints per question (translated name and text, media, `{currencyKey, amount}`) / purchases per (user, hint, scope) |
| `quiz_session` | id, user, pack, `status open → paying → done` (or `closed`), server `started_at`, `deadline_at`, `expires_at`, picked questions with option order, current index, `served_at`, `last_answer_at`, counts, score, `campaign_attempt`, stored `result` and `plan` |
| `quiz_answer` | one row per (session, question): options or text, `outcome correct\|wrong\|skip\|late`, seconds |
| `quiz_user_pack` | plays, wins, best/last score, last correct/total, rewards paid, first pass, last play |
| `quiz_user_question` / `quiz_user_mission` / `quiz_user_day` | question rewards paid / mission start, completion and reward / daily counters |
| `quiz_report_reason` / `quiz_question_report` | report reasons (generic defaults: wrong_answer, typo, unclear, media, offensive, other) / player reports (one open per player and question) |
| `quiz_result_template` / `quiz_result_delivery` | result report text, won/lost words, mission or game, webhook url and sealed secret / webhook queue with retries |

The migration adds `quiz` to `economy_user_prize_source_type_check`, keeping every source type already in the constraint (it reads the current list), so it composes with other plugins that extend the list the same way.

## Console («آزمون‌ها», `#/quiz/…`)

A sidebar group with: **پک‌های آزمون** (filter by mission, editor with pass rule, question count, bank category, timing, shuffle overrides, send answers, pass reward with currencies and economy prize pickers, campaign stage picker, pack stats), **سوال‌ها** (filter by pack / bank, type, category, active, reported; editor with translated question, option rows per language with the right option, accepted text answers per language, explanation, media, difficulty, category, tags, own time and reward; hints with price and purchase counts), **ماموریت‌ها و بازی‌ها**, **گزارش بازیکنان** (open / resolved / dismissed, review panel: resolve or dismiss one or every open report of the question, switch the question off, open it to fix; report reasons editor), **گزارش نتیجه** (templates with placeholder chips, webhook url and signing key, delivery log), **ورود و خروج** (JSON import with check-only and all-or-nothing errors listed per row, legacy spreadsheet columns as `rows`, export of a pack or everything), **آمار آزمون‌ها** (7/30/90 days: players, plays, finished, passes, pass rate, average score and time per pack, hardest questions, hints bought, open reports) and **تنظیمات آزمون**. The native player page gets an «آزمون‌ها» card (packs with plays, passes and best score, «از نو» per pack with an option to forget the paid reward, missions, last sessions). fa/en, RTL, 390 px.

Console API under `/v2/console/base/quiz`: `settings` (GET/PUT), `refs`, `missions` (GET `?kind`, POST, `{id}` GET/PUT/DELETE — `IN_USE` while it has packs or missions), `packs` (GET `missionId`, `category`, `q`, paging; POST, `{id}` GET/PUT/DELETE — `IN_USE` once played unless `?force=true`), `questions` (GET `packId` (`none` = bank), `type`, `category`, `tag`, `active`, `reported`, `q`; POST, `{id}` GET/PUT/DELETE — `IN_USE` once answered unless `?force=true`), `questions/{id}/hints` (GET/POST), `hints/{hid}` (PUT/DELETE), `import` (POST), `export` (GET `packId` / `missionId`), `reports` (GET `status`, `questionId`, `reasonKey`), `reports/{id}` (PUT `{status, resolution, sameQuestion, disableQuestion}`), `reasons` (GET/POST, `{id}` PUT/DELETE), `templates` (GET/POST, `{id}` PUT/DELETE, `{id}/deliveries`), `stats` (`days`), `players/{id}` (GET), `players/{id}/packs/{packId}` (DELETE, `?rewards=reset`).

Import format (export writes the same): `{"missions":[{key, kind, parentKey, name, …}], "packs":[{key, missionKey, name, questionCount, passCorrect, poolCategory, …}], "questions":[{key?, packKey, type, question, options:[{id, text, correct}], answers, explanation, hints:[{name, text, price}], …}], "rows":[{"Pack Key", "Question", "Option 1"…"Option 6", "Correct Options", "Accepted Answers", "Answer Language", "Difficulty", "Time", "Category", "Explanation", "Key"}], "packKey"?, "lang"?, "dryRun"?}`. One error anywhere and nothing is written (`QUIZ_IMPORT_HAS_ERRORS` with `meta.errors[{section, index, field, code}]`).

## Worker

Every 30 s under the `clusterkit` lease `quiz-worker`: score open sessions past their time, finish sessions left in `paying`, post due result webhooks (10 s × 3ⁿ back-off, failed after `retention.webhookAttempts`; the resolved address must not be private or loopback), and every 30 min sweep old sessions, deliveries, day counters and rows of deleted players.

## Changed vs legacy

- Kept: packs / missions / games with sequential or all-at-once order and times, pass threshold, options and text questions (several right options), hints with a one-currency price owned for good, per-question and per-pack time, random pack (`forceNew`, `preferNew`), answer-all, statistics, result report templates with the same placeholders (`user.*` now `id`, `username`, `displayName`; `game.*` became `mission.*`/`missions.*`), webhook deliveries with an HMAC signature, the Excel column set (as JSON rows).
- Fixed: right answers are not sent before answering unless the admin allows it (legacy sent them whenever a pack or mission had `SendAnswers`, and the Progress quiz stages trusted the success the client sent), server clock and per-question deadlines, resend-safe answers and finish, one open session per pack in the database, rewards and hint prices paid once with ledger grant ids, limits counted in the database, prizes through economy with source ids.
- Added: question banks by category with N questions per play, one-by-one delivery, player question reports and review, the campaign stage link, economy and analytics events, translated texts everywhere, console stats and the player card.
- Dropped: the partner portal (`PlatformPartnerId`, partner endpoints and permissions), content / content-pack ids on questions, options and hints (one media url instead), the per-language text answer storage through `LanguageItem` (answers are JSON per language on the question), the download-as-file endpoint of reports (the client can save the body).

## Duels (Quiz of Kings style)

A turn-based duel between two players on top of the question bank. Tables (migration `20261225000001-quiz-duel.sql`): `quiz_category` (key, translated name/description, icon, color, sort, active, `duel`), `quiz_duel` (players, mode `random|friend`, status `waiting|invited|active|finished|expired|declined|cancelled`, the duel rules copied from settings at start, turn user and deadline, expiry, scores, winner, reason `score|draw|timeout|forfeit|expired|declined|admin`, entry fee and payment flags), `quiz_duel_round` (picker, offered categories fixed when the round opens, picked category, questions with option order), `quiz_duel_answer` (one row per duel, player and question: served time, answer time, option, outcome `pending|correct|wrong|late|skip`, seconds), `quiz_duel_day` (daily starts).

Flow: round 1 is picked and answered by the starter, then the opponent answers it and picks round 2, then the starter answers round 2 and picks round 3, and so on. Questions come from active option questions whose `category` equals the category key (`duel.questionSource` any or bank), never twice in one duel, questions either player saw in `duel.avoidRepeatDays` last. A question's clock starts when it is served; an answer after `questionSeconds + graceSeconds` is `late`. An expired served question is closed as `late` on the next `quiz_duel_question`. Answers are resend-safe (`replayed: true`). The opponent's answers of a round show only after the player finished that round.

Random duels join the oldest waiting duel of someone else (not blocked) or start a new one that waits `duel.waitingHours`. Friend duels need a friendship when `duel.friendsOnly`, and an accept within `duel.inviteHours`. A turn not finished in `duel.turnHours` ends the duel (`timeoutResult`: the opponent wins, or the score so far). Entry fees (`duel.entryFee`) are debited once per player with a ledger grant id and refunded on expiry, decline and cancel. At the end each player gets `winPrize`/`losePrize`/`drawPrize` (economy prize keys, source `duel:<duel>:<user>`), `perCorrect` currency per right answer, and the economy events `duel.events.played|won|correct`, once. Live notifications: code 1410, subject `quiz_duel`, content `{event, duelId, from, status}` with events `invite`, `joined`, `accepted`, `declined`, `turn`, `finished`, `expired`, `cancelled`.

Client RPCs: `quiz_duel_categories`, `quiz_duel_start`, `quiz_duel_list`, `quiz_duel_get`, `quiz_duel_accept`, `quiz_duel_decline`, `quiz_duel_pick_category`, `quiz_duel_question`, `quiz_duel_answer`, `quiz_duel_forfeit`, `quiz_duel_stats`. Console: «دوئل‌ها» (stats tiles, filters, detail with every answer, cancel with refund), «دسته‌های دوئل», and the duel section in «تنظیمات آزمون». Console API: `duel-categories` (GET/POST, `{id}` PUT/DELETE with `IN_USE` unless `?force=true`), `duels` (GET `status`, `mode`, `player`), `duels/{id}` (GET), `duels/{id}/cancel` (POST), `duel-stats` (GET `days`).

## Port notes

Moved from the PP fork with `internal/campaignhooks` and `internal/contentimport`. nakama-farsi has no campaign plugin, so the campaign stage link is unavailable (the settings page says so) and packs never own a stage.
