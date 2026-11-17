# Trophy: match trophies, leagues and seasons

A generic plugin (`integrations/trophy`, `trophy.so`). A match result gains or loses trophies, the trophy count sets the player's league, reaching a league pays a one-time economy prize, and an optional season reset trims trophies. Nothing in it names a game: the match event, its fields and the counted modes are settings.

## Decisions

- **Server-authoritative.** Trophies move only from a server event: an economy player event (`match.source = economy_event`, recorded by a game plugin with `client_allowed=false`) or an `internal/gameevents` event (`game_event`). There is no client RPC that writes trophies.
- **Once per match.** Each change is a `trophy_log` row unique on `(user_id, source_id)`; the source is `m:<matchId>` when the event carries `match.matchField`, else the event id. A resent result never counts twice.
- **Amounts.** Win `+amounts.win`, loss `-amounts.loss`, draw `amounts.draw`. With `amounts.scale` the amount shifts by `scaleStep` per `scalePer` trophies of difference against the opponent's count before that match (read from the opponent's own log row of the same match), clamped to `minWin..maxWin` / `minLoss..maxLoss`. Never below 0. With `amounts.floors`, a loss cannot drop a player below the minimum of a league marked `floor`.
- **Leagues.** Admin rows (`trophy_league`) with a minimum. The league is the highest active league whose minimum is reached. Entering a league with a `prizeKey` queues a `trophy_prize` row once per player (`prizes.oncePer = ever`) or once per season (`season`); `prizes.onlyRising` pays only when rising. Payment goes through `economyhooks.GrantPrize` (source type `trophy`), retried by the worker.
- **Seasons.** Optional (`season.enabled`, CRON `season.schedule`). Modes: `cut` (above `season.above` keep `keepPercent`), `league` (each league's `resetTo`), `zero`. One transaction per season under a row lock; the console can reset now.
- **Matchmaking.** With `matchmaking.enabled`, a MatchmakerAdd before-hook (after economy's, through `hookchain`) adds a numeric property (`matchmaking.property`) and a `±range` query clause for the listed games.
- **Cluster.** One worker per cluster under the `trophy-worker` lease (season tick, prize retries); prize rows are taken with `FOR UPDATE SKIP LOCKED`.
- **Events out.** `trophy_changed` game event (fields `trophies`, `delta`, `best`; context = league) for leaderboardx and clans; socket notifications `trophy_changed` (code 2401, not persistent) and `trophy_league` (2402, persistent when rising).

## Client RPCs

| RPC | In | Out |
|---|---|---|
| `trophy_me` | `{lang?}` | `{enabled, trophies, best, seasonBest, wins, losses, draws, league, nextLeague:{…, need}, season, pendingPrizes, amounts}` |
| `trophy_leagues` | `{lang?}` | `{leagues:[{key, name, description, iconUrl, color, minTrophies, maxTrophies, floor, prize:{prizeKey, status}, reached, current}], trophies, best, season}` |
| `trophy_history` | `{limit?}` | `{items:[{reason, delta, before, after, matchId, leagueBefore, leagueAfter, opponent, createdAt}]}` |
| `trophy_players` | `{userIds:[…≤100]}` | `{players:[{userId, trophies, best, league}]}` |

## Console (`/v2/console/base/trophies`, ACL `LEADERBOARD`)

Leagues, players (set/add trophies, log, prizes), seasons (reset now), settings. Pages under «جام و لیگ» (`#/trophies/…`).

## Data

`trophy_settings`, `trophy_league`, `trophy_player`, `trophy_log`, `trophy_season`, `trophy_prize` (migration `20261229000000-trophy-schema.sql`, also adds `trophy` to the economy prize source types).
