# Cup: knockout tournaments

A generic plugin (`integrations/cup`, `cup.so`). Players sign up for a cup, the server builds a single-elimination bracket, plays the matches through a game that registers in `internal/cuphooks`, advances winners, handles byes, no-shows and draws, and pays the prizes. Nakama's own tournaments (time-boxed leaderboards) are not used. Nothing in it names a game.

## Decisions

- **Server-authoritative.** A game reports each game result with `cuphooks.Report` (or the server-to-server RPC `cup_report`, http key only) using its own match id. Clients never report winners. A result is counted once: a `cup_game` row is finished or void only once, so a resent result is a no-op.
- **Games.** A game plugin calls `cuphooks.Register(Game{Key, Name, Launch})`. When both players are ready the cup calls `Launch` with the two players, the game number, the per-cup game options, `Rated` (count trophies or rating, default off), the draw hint and the join timeout, and stores the returned match id. A cup with no game (`rules.game = ""`) is external: the match id is `cup:<gameId>` and another server reports with `cup_report`.
- **Bracket.** Size is the next power of two above the players; byes go to the top seeds (standard seed order, seed 1 and 2 meet only in the final). Seeding is `random` or `trophies` (trophy plugin, snapshot at the start). An optional third-place match takes the semi-final losers.
- **Games per match.** `bestOf` (odd, 1..9, default 1), overridden per round with `roundBestOf` (`final`, `semi`, `third` or a round number). After a game the next one starts at once until one side has `bestOf/2+1` wins.
- **Timing, chosen per cup.** `window`: each round opens when the previous round is done and stays open `windowMinutes`; a player calls `cup_ready`, and when both are ready within `readySeconds` the game launches. `live`: round N starts at `startsAt + checkInMinutes + (N-1) * roundMinutes` (or later if the previous round runs late); check-in opens `checkInMinutes` before; at the start both checked in = launch, one = he wins (`no_show`). A round closes at its deadline: the series leader wins, else the one who showed wins, else `noShow` (`both_out`, `higher_seed`, `random`) when nobody showed, or `unplayed` when both showed but did not finish. Both out leaves the next match with one player, who walks over.
- **Draws.** `draw.rule`: `replay` (up to `maxReplays`, then `fallback`: `higher_seed` or `random`), `higher_seed` or `random`. With `draw.gameDecides` the game is told to break ties itself (FlickKick plays sudden death).
- **No-shows inside a game.** A game that reports `void` with one player present gives that player the game; nobody present voids it and the match waits for ready again. A game that never reports is void after `gameMinutes`.
- **Entry.** Capacity 2..`limits.maxCapacity` (any number, byes fill the bracket), minimum players, fee in any currency or free, requirements (min/max trophies, leagues, tags any/all, excluded tags), signup window, `limits.activePerPlayer`. Withdraw before the start refunds `withdrawRefundPercent`. At the start, fewer than `minPlayers` cancels and refunds everything; fewer than capacity follows `underfill` (`shrink` with byes, or `cancel`).
- **Prizes.** Economy prize keys for champion, runner-up, semi-finalists (or third and fourth with a third-place match) and "reached round N". Optional prize pool: `pool.percent` of the fees plus `pool.base`, split by `pool.shares`, paid in the fee currency. Payouts are `cup_payout` rows unique per (cup, player, kind), paid with `SKIP LOCKED` and retried; refunds and pool credits use wallet grant ids, so they are paid once.
- **Recurring cups.** Templates (`cup_template`) with a CRON schedule create the next cup (one open cup per template, unique on template and start).
- **Admin.** Create, edit, publish, start now, cancel with refunds, add or remove players before the start, disqualify, swap a player who has not played yet, force a result or void a match, retry payouts.
- **Cluster.** One worker under the `cup-worker` lease every `limits.tickSeconds`; every bracket change locks the cup row (`FOR UPDATE`, the worker uses `SKIP LOCKED`), so reports from any node, ready calls and the worker never race.
- **Events out.** Socket notifications (code `events.code`, default 2420), in-app messages for check-in, finish, prize and cancel (`events.inApp`), game events `cup_match` and `cup_finished` (`events.publish`).

## Client RPCs

`cup_list {status?: upcoming|signup|running|finished|mine, limit, offset}`, `cup_get {cupId}` (cup, bracket, my match, champion, top four), `cup_signup`, `cup_withdraw`, `cup_my_match {cupId?}` (match and next action), `cup_ready {cupId}` (ready or check-in), `cup_start {cupId}` (ready and launch now if the opponent is ready; returns the game to join), `cup_history {limit}`. Server only: `cup_report`. Types are in `/docs/openapi.json`.

Notifications: `cup_round_started`, `cup_check_in`, `cup_opponent_ready`, `cup_match_ready` (with `gameMatchId`), `cup_game_result`, `cup_game_void`, `cup_advanced`, `cup_eliminated`, `cup_finished`, `cup_prize`, `cup_cancelled`.

## Console (`/v2/console/base/cups`, ACL `LEADERBOARD`)

Pages under «تورنمنت‌ها» (bundle `console/ui/cup`): `#/cups/list` (filters, search, create and edit with every rule), `#/cups/list/<id>` (live bracket refreshed every 5 seconds, match panel with games, force a result or void, players with remove, swap and disqualify, add a player by username or id, publish, start now, cancel with a reason and refunds, delete, payouts with retry), `#/cups/templates` (recurring templates, run now) and `#/cups/settings`. e2e `78-cups`.

## Data

`cup_settings`, `cup_template`, `cup`, `cup_player`, `cup_match`, `cup_game`, `cup_payout` (migration `20270101000000-cup-schema.sql`, also adds `cup` to the economy prize source types).
